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

	AutoCreateTopics bool

	BatchSize    int
	BatchBytes   int
	BatchTimeout int
}

type Microservice struct {
	config        *config.Config
	mux           *http.ServeMux
	middlewares   []Middleware
	loggerConfig  logger.LoggerConfig
	kafkaConfig   *KafkaConfig
	kafkaHandlers map[string]KafkaConsumerHandler // topic -> handler
	// kafkaConsumers []*KafkaConsumer                // running consumers
	// kafkaWriters   Writer                          // topic -> writer
	mu          sync.RWMutex
	kafkaClient *kafkaClient
	kafkaWg     sync.WaitGroup // WaitGroup for Kafka consumers
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
	Consume(topic string, handler KafkaConsumerHandler)

	// Kafka producer support
	// Publish(topic string, key []byte, value []byte) error
	// PublishJSON(topic string, key string, v any) error
}

func NewMicroservice(cfg *config.Config, loggerConfig logger.LoggerConfig) IMicroservice {
	ms := &Microservice{
		config:        cfg,
		mux:           http.NewServeMux(),
		loggerConfig:  loggerConfig,
		kafkaHandlers: make(map[string]KafkaConsumerHandler),
		// kafkaConsumers: make([]*KafkaConsumer, 0),
		kafkaClient: nil,
	}

	// Do NOT connect to Kafka here (lazy connection)
	// Store config for later use in startKafkaConsumers()
	ms.kafkaConfig = &KafkaConfig{
		Brokers:          cfg.Kafka.Brokers,
		RequestTimeout:   cfg.Kafka.RequestTimeout,
		RetryBackoff:     cfg.Kafka.RetryBackoff,
		MaxRetries:       cfg.Kafka.MaxRetries,
		GroupID:          cfg.Kafka.GroupID,
		AutoCreateTopics: cfg.Kafka.AutoCreateTopics,
		BatchSize:        cfg.Kafka.BatchSize,
		BatchBytes:       cfg.Kafka.BatchBytes,
		BatchTimeout:     cfg.Kafka.BatchTimeout,
	}

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
		return nil
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
	kc, err := m.createKafkaWriter(&cfg)
	if err != nil {
		m.mu.Unlock()
		return err
	}

	m.kafkaClient = kc
	m.mu.Unlock()

	log.Printf("Kafka configured with brokers: %s", cfg.Brokers)

	return nil
}

func (m *Microservice) isKafkaConnected() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.kafkaConfig != nil && m.kafkaClient != nil
}

func (m *Microservice) Start() {
	var handler http.Handler = m.mux
	for _, mw := range m.middlewares {
		handler = mw(handler)
	}

	srv := &http.Server{
		Addr:         ":" + m.config.Server.Port,
		Handler:      handler,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
		IdleTimeout:  120 * time.Second,
	}

	// Context for Kafka consumers
	kafkaCtx, kafkaCancel := context.WithCancel(context.Background())
	defer kafkaCancel()

	// Start Kafka consumers in background (non-blocking)
	go m.startKafkaConsumers(kafkaCtx)

	// Channel to listen for errors from HTTP server
	serverErrors := make(chan error, 1)

	// Start HTTP Server in background
	go func() {
		log.Printf("starting server on %s", srv.Addr)
		serverErrors <- srv.ListenAndServe()
	}()

	// Channel for OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Block until we receive signal or error
	select {
	case err := <-serverErrors:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("server error: %v", err)
		}
	case <-quit:
		log.Println("shutting down server...")

		// Stop Kafka consumers first
		kafkaCancel()

		// Wait for all Kafka consumers to finish
		log.Println("waiting for kafka consumers to shutdown...")
		m.kafkaWg.Wait()
		log.Println("all kafka consumers stopped")

		if m.kafkaClient != nil {
			m.kafkaClient.Close()
		}

		// Shutdown HTTP server gracefully
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("server forced to shutdown: %v", err)
		}

		log.Println("server exited")
	}
}

func (m *Microservice) Use(middleware Middleware) {
	m.middlewares = append(m.middlewares, middleware)
}

func (m *Microservice) preHandle(handler MyHandler, middlewares ...Middleware) http.HandlerFunc {
	// Wrap MyHandler into http.HandlerFunc
	final := func(w http.ResponseWriter, r *http.Request) {
		// Create a new per-request logger using the configured LoggerConfig
		requestLogger := logger.NewCustomLogger(m.config.Server.Name, m.loggerConfig)
		var writer Writer
		if m.kafkaClient != nil {
			writer = m.kafkaClient.writer
		}
		handler(newMuxContext(w, newRequest(r, writer), m.config, requestLogger).(*Ctx))
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
	// m.mux.HandleFunc(fmt.Sprintf("%s %s", http.MethodGet, path), m.preHandle(handler, middlewares...))
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
	log.Printf("Starting Kafka consumers...")
	if m.kafkaConfig == nil || strings.TrimSpace(m.kafkaConfig.Brokers) == "" {
		log.Printf("Kafka not configured; skipping consumer startup")
		return
	}

	// Lazy connect on first consumer startup
	if m.kafkaClient == nil {
		// Connect to Kafka without blocking - only retry once on failure
		go func() {
			if err := m.ConnectKafka(*m.kafkaConfig); err != nil {
				log.Printf("Failed to connect to Kafka: %v", err)
				return
			}

			// Once connected, start the consumers
			for topic, handler := range m.kafkaHandlers {
				kc, err := NewKafkaConsumer(m.kafkaClient.dialer, m.kafkaConfig, ConsumerConfig{
					Topic:   topic,
					Handler: handler,
				})
				if err != nil {
					log.Printf("failed to start kafka consumer for topic %s: %v", topic, err)
					continue
				}

				m.kafkaWg.Add(1)
				go func(consumer *KafkaConsumer, t string) {
					defer m.kafkaWg.Done()
					consumer.Run(ctx, m.createKafkaContext())
					log.Printf("Kafka consumer stopped for topic: %s", t)
				}(kc, topic)
				log.Printf("Kafka consumer started for topic: %s", topic)
			}
		}()
		return
	}

	// If already connected, start consumers directly
	for topic, handler := range m.kafkaHandlers {
		kc, err := NewKafkaConsumer(m.kafkaClient.dialer, m.kafkaConfig, ConsumerConfig{
			Topic:   topic,
			Handler: handler,
		})
		if err != nil {
			log.Printf("failed to start kafka consumer for topic %s: %v", topic, err)
			continue
		}

		m.kafkaWg.Add(1)
		go func(consumer *KafkaConsumer, t string) {
			defer m.kafkaWg.Done()
			consumer.Run(ctx, m.createKafkaContext())
			log.Printf("Kafka consumer stopped for topic: %s", t)
		}(kc, topic)
		log.Printf("Kafka consumer started for topic: %s", topic)
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

type Writer interface {
	WriteMessages(ctx context.Context, msg ...kafka.Message) error
	Close() error
	Stats() kafka.WriterStats
}

func connectToBrokers(brokers []string, dialer *kafka.Dialer, timeout time.Duration) ([]*kafka.Conn, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conns := make([]*kafka.Conn, 0, len(brokers))

	for _, broker := range brokers {

		conn, err := dialer.DialContext(ctx, "tcp", broker)
		if err != nil {
			return nil, fmt.Errorf("failed to connect to broker %s: %w", broker, err)
		}
		conns = append(conns, conn)
	}
	return conns, nil
}

// getWriter returns a cached writer for a topic or creates a new one
func (m *Microservice) createKafkaWriter(conf *KafkaConfig) (*kafkaClient, error) {
	if m.kafkaConfig == nil || strings.TrimSpace(m.kafkaConfig.Brokers) == "" {
		return nil, errors.New("kafka not configured")
	}

	brokers := strings.Split(m.kafkaConfig.Brokers, ",")
	var addrs []string
	for _, b := range brokers {
		if s := strings.TrimSpace(b); s != "" {
			addrs = append(addrs, s)
		}
	}
	if len(addrs) == 0 {
		return nil, errors.New("no valid kafka brokers")
	}

	dialer, err := setupDialer(conf)
	if err != nil {
		return nil, fmt.Errorf("failed to setup kafka dialer: %w", err)
	}

	conns, err := connectToBrokers(addrs, dialer, conf.RequestTimeout)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to kafka brokers: %w", err)
	}

	kc := &kafkaClient{
		dialer: dialer,
		conn: &multiConn{
			conns:  conns,
			dialer: dialer,
			mu:     sync.RWMutex{},
		},
		reader: make(map[string]*kafka.Reader),
		mu:     &sync.RWMutex{},
	}

	batchTimeout := time.Duration(conf.BatchTimeout) * time.Millisecond
	if batchTimeout <= 0 {
		batchTimeout = 500 * time.Millisecond
	}

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      strings.Split(conf.Brokers, ","),
		Dialer:       dialer,
		BatchSize:    conf.BatchSize,
		BatchBytes:   conf.BatchBytes,
		BatchTimeout: batchTimeout,
	})
	kc.writer = w
	return kc, nil
}

type KafkaParams struct {
	Headers       map[string]string
	HighWaterMark int64
	Key           string
	Value         string
	Partition     int
	Offset        int64
	Time          time.Time
	Topic         string
	WriterData    interface{}
}

func (m *Microservice) Consume(topic string, handler KafkaConsumerHandler) {
	if topic == "" || handler == nil {
		panic("topic and handler must be provided for Kafka consumer")
	}

	// If Kafka is not configured, skip (user may not want to use Kafka)
	if !m.isKafkaConfigured() {
		log.Printf("Kafka not configured; skipping consumer for topic %s", topic)
		return
	}

	if _, exists := m.kafkaHandlers[topic]; exists {
		return
	}

	// Register handler even if not connected yet (lazy connection in Start())
	m.kafkaHandlers[topic] = handler

	// Only create topic if already connected
	if m.kafkaClient != nil && m.kafkaConfig.AutoCreateTopics {
		_ = m.kafkaClient.CreateTopic(topic, 1)
	}

	log.Printf("Kafka consumer registered for topic: %s (lazy start)", topic)
}
