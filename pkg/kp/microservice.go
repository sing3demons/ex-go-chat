package kp

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"realtime-chat-system/config"
	"realtime-chat-system/pkg/logger"

	"github.com/segmentio/kafka-go"
)

type MyHandler func(ctx *Ctx)
type HandleFunc func(http.Handler) http.Handler
type Middleware HandleFunc

type KafkaConfig struct {
	Brokers        string
	RequestTimeout time.Duration
	RetryBackoff   time.Duration
	MaxRetries     int
	GroupID        string
}

type Microservice struct {
	config         *config.Config
	mux            *http.ServeMux
	middlewares    []Middleware
	loggerConfig   logger.LoggerConfig
	kafkaConfig    *KafkaConfig
	kafkaHandlers  map[string]KafkaConsumerHandler // topic -> handler
	kafkaConsumers []*KafkaConsumer                // running consumers
	mu             sync.RWMutex
}
type IMicroservice interface {
	Start()
	// GET(path string, handler http.HandlerFunc, middlewares ...Middleware)
	GET(path string, handler MyHandler, middlewares ...Middleware)
	// POST(path string, handler http.HandlerFunc, middlewares ...Middleware)
	POST(path string, handler MyHandler, middlewares ...Middleware)
	// PUT(path string, handler http.HandlerFunc, middlewares ...Middleware)
	PUT(path string, handler MyHandler, middlewares ...Middleware)
	// DELETE(path string, handler http.HandlerFunc, middlewares ...Middleware)
	DELETE(path string, handler MyHandler, middlewares ...Middleware)
	// PATCH(path string, handler http.HandlerFunc, middlewares ...Middleware)
	PATCH(path string, handler MyHandler, middlewares ...Middleware)

	Use(middleware Middleware)
	// multiple methods (GET, POST, PUT, DELETE, PATCH)
	// Any(path string, handler MyHandler, middlewares ...Middleware)
	Match(methods, path string, handler MyHandler, middlewares ...Middleware)

	// Kafka consumer support
	Consume(topic string, handler KafkaConsumerHandler) error
}

func NewMicroservice(cfg *config.Config, loggerConfig logger.LoggerConfig) IMicroservice {
	ms := &Microservice{
		config:         cfg,
		mux:            http.NewServeMux(),
		loggerConfig:   loggerConfig,
		kafkaHandlers:  make(map[string]KafkaConsumerHandler),
		kafkaConsumers: make([]*KafkaConsumer, 0),
	}

	ms.ConnectKafka(KafkaConfig{})

	return ms
}

func setupDialer(conf *KafkaConfig) (*kafka.Dialer, error) {
	dialer := &kafka.Dialer{
		Timeout:   10 * time.Second,
		DualStack: true,
	}

	// if conf.SecurityProtocol == protocolSASL || conf.SecurityProtocol == protocolSASLSSL {
	// 	mechanism, err := getSASLMechanism(conf.SASLMechanism, conf.SASLUser, conf.SASLPassword)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	dialer.SASLMechanism = mechanism
	// }

	// if conf.SecurityProtocol == "SSL" || conf.SecurityProtocol == "SASL_SSL" {
	// 	tlsConfig, err := createTLSConfig(&conf.TLS)
	// 	if err != nil {
	// 		return nil, err
	// 	}

	// 	dialer.TLS = tlsConfig
	// }

	return dialer, nil
}

// ConnectKafka sets Kafka brokers configuration (comma-separated list is supported)
func (m *Microservice) ConnectKafka(cfg KafkaConfig) error {
	if strings.TrimSpace(cfg.Brokers) == "" {
		return nil
	}
	// Validate brokers
	if cfg.Brokers == "" {
		return errors.New("brokers cannot be empty")
	}

	// Set defaults
	if cfg.RequestTimeout == 0 {
		cfg.RequestTimeout = 10 * time.Second
	}
	if cfg.RetryBackoff == 0 {
		cfg.RetryBackoff = 100 * time.Millisecond
	}
	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3
	}

	m.mu.Lock()
	m.kafkaConfig = &cfg
	m.mu.Unlock()

	log.Printf("Kafka configured with brokers: %s", cfg.Brokers)

	return nil
}

func (m *Microservice) Start() {
	var handler http.Handler = m.mux
	for _, mw := range m.middlewares {
		handler = mw(handler)
	}

	srv := http.Server{
		Addr:         ":" + m.config.Server.Port,
		Handler:      handler,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start Kafka consumers first (non-blocking)
	m.startKafkaConsumers(ctx)

	var wg sync.WaitGroup

	// Start HTTP Server
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Printf("starting server on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server listen err: %v", err)
		}
	}()

	// Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down server...")

	// stop Kafka consumers
	cancel()
	m.closeKafkaConsumers()

	// Shutdown HTTP server
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("server forced to shutdown: %v", err)
		os.Exit(1)
	}

	wg.Wait()
	log.Println("server exited")
}

//	func (m *Microservice) GET(path string, handler http.HandlerFunc, middlewares ...Middleware) {
//		m.mux.HandleFunc(fmt.Sprintf("%s %s", http.MethodGet, path), m.preHandle(handler, middlewares...))
//	}
// func (m *Microservice) Post(path string, handler http.HandlerFunc, middlewares ...Middleware) {
// 	m.mux.HandleFunc(fmt.Sprintf("%s %s", http.MethodPost, path), m.preHandle(handler, middlewares...))
// }
// func (m *Microservice) PUT(path string, handler http.HandlerFunc, middlewares ...Middleware) {
// 	m.mux.HandleFunc(fmt.Sprintf("%s %s", http.MethodPut, path), m.preHandle(handler, middlewares...))
// }
// func (m *Microservice) DELETE(path string, handler http.HandlerFunc, middlewares ...Middleware) {
// 	m.mux.HandleFunc(fmt.Sprintf("%s %s", http.MethodDelete, path), m.preHandle(handler, middlewares...))
// }
// func (m *Microservice) PATCH(path string, handler http.HandlerFunc, middlewares ...Middleware) {
// 	m.mux.HandleFunc(fmt.Sprintf("%s %s", http.MethodPatch, path), m.preHandle(handler, middlewares...))
// }

func (m *Microservice) Use(middleware Middleware) {
	m.middlewares = append(m.middlewares, middleware)
}

func (m *Microservice) preHandle(handler MyHandler, middlewares ...Middleware) http.HandlerFunc {
	// Wrap MyHandler into http.HandlerFunc
	final := func(w http.ResponseWriter, r *http.Request) {
		// Create a new per-request logger using the configured LoggerConfig
		requestLogger := logger.NewCustomLogger(m.config.Server.Name, m.loggerConfig)
		handler(newMuxContext(w, r, m.config, requestLogger).(*Ctx))
	}
	// Apply middlewares in reverse order (so the first is outermost)
	for i := len(middlewares) - 1; i >= 0; i-- {
		final = middlewares[i](http.HandlerFunc(final)).ServeHTTP
	}
	return final
}

func (m *Microservice) add(method, path string, handler MyHandler, middlewares ...Middleware) {
	m.mux.HandleFunc(fmt.Sprintf("%s %s", method, path), m.preHandle(handler, middlewares...))
}

func (m *Microservice) GET(path string, handler MyHandler, middlewares ...Middleware) {
	m.add(http.MethodGet, path, handler, middlewares...)
}

func (m *Microservice) POST(path string, handler MyHandler, middlewares ...Middleware) {
	// m.mux.HandleFunc(fmt.Sprintf("%s %s", http.MethodPost, path), m.preHandle(handler, middlewares...))
	m.add(http.MethodPost, path, handler, middlewares...)
}

func (m *Microservice) PUT(path string, handler MyHandler, middlewares ...Middleware) {
	// m.mux.HandleFunc(fmt.Sprintf("%s %s", http.MethodPut, path), m.preHandle(handler, middlewares...))
	m.add(http.MethodPut, path, handler, middlewares...)
}

func (m *Microservice) DELETE(path string, handler MyHandler, middlewares ...Middleware) {
	// m.mux.HandleFunc(fmt.Sprintf("%s %s", http.MethodDelete, path), m.preHandle(handler, middlewares...))
	m.add(http.MethodDelete, path, handler, middlewares...)
}

func (m *Microservice) PATCH(path string, handler MyHandler, middlewares ...Middleware) {
	// m.mux.HandleFunc(fmt.Sprintf("%s %s", http.MethodPatch, path), m.preHandle(handler, middlewares...))
	m.add(http.MethodPatch, path, handler, middlewares...)
}

func (m *Microservice) Match(methods, path string, handler MyHandler, middlewares ...Middleware) {
	for _, method := range []string{methods} {
		// m.mux.HandleFunc(fmt.Sprintf("%s %s", strings.ToUpper(method), path), m.preHandle(handler, middlewares...))
		m.add(strings.ToUpper(method), path, handler, middlewares...)
	}
}

func (m *Microservice) isKafkaConfigured() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.kafkaConfig != nil
}

// startKafkaConsumers spins up all registered consumers with the given context
func (m *Microservice) startKafkaConsumers(ctx context.Context) {
	m.mu.RLock()
	if m.kafkaConfig == nil || len(m.kafkaConsumers) == 0 {
		m.mu.RUnlock()
		return
	}
	consumers := append([]*KafkaConsumer(nil), m.kafkaConsumers...)
	m.mu.RUnlock()

	kafkaCtx := m.createKafkaContext()
	for _, consumer := range consumers {
		consumer.Run(ctx, kafkaCtx)
	}
}

// closeKafkaConsumers shuts down all consumers
func (m *Microservice) closeKafkaConsumers() {
	m.mu.RLock()
	consumers := append([]*KafkaConsumer(nil), m.kafkaConsumers...)
	m.mu.RUnlock()

	for i, consumer := range consumers {
		if err := consumer.Close(); err != nil {
			log.Printf("error closing kafka consumer %d: %v", i, err)
		}
	}
}

// createKafkaContext builds a base context/logger for Kafka handlers
func (m *Microservice) createKafkaContext() *Ctx {
	requestLogger := logger.NewCustomLogger(m.config.Server.Name, m.loggerConfig)
	return &Ctx{
		Cfg: m.config,
		Log: requestLogger,
	}
}

// Consume registers a Kafka consumer for the given topic
func (m *Microservice) Consume(topic string, handler KafkaConsumerHandler) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isKafkaConfigured() {
		return errors.New("kafka config not set")
	}

	// Create Kafka consumer
	kc, err := NewKafkaConsumer(m.kafkaConfig, ConsumerConfig{
		Topic:   topic,
		Handler: handler,
	})
	if err != nil {
		return err
	}

	// Subscribe to topic
	if err := kc.Subscribe([]string{topic}); err != nil {
		return err
	}

	// Store handler and consumer
	m.kafkaHandlers[topic] = handler
	m.kafkaConsumers = append(m.kafkaConsumers, kc)

	log.Printf("Kafka consumer registered for topic: %s with group: %s", topic, m.kafkaConfig.GroupID)
	return nil
}
