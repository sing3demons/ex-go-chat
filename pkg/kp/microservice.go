package kp

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"realtime-chat-system/config"
	"realtime-chat-system/pkg/logAction"
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

	ms.ConnectKafka(KafkaConfig{
		Brokers:        cfg.Kafka.Brokers,
		RequestTimeout: cfg.Kafka.RequestTimeout,
		RetryBackoff:   cfg.Kafka.RetryBackoff,
		MaxRetries:     cfg.Kafka.MaxRetries,
		GroupID:        cfg.Kafka.GroupID,
		AutoCreateTopics: cfg.Kafka.AutoCreateTopics,
		BatchSize:      cfg.Kafka.BatchSize,
		BatchBytes:     cfg.Kafka.BatchBytes,
		BatchTimeout:   cfg.Kafka.BatchTimeout,
	})

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
	// m.closeKafkaConsumers()
	// m.closeKafkaProducers()
	if m.kafkaClient != nil {
		m.kafkaClient.Close()
	}

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

func (m *Microservice) Use(middleware Middleware) {
	m.middlewares = append(m.middlewares, middleware)
}

func (m *Microservice) preHandle(handler MyHandler, middlewares ...Middleware) http.HandlerFunc {
	// Wrap MyHandler into http.HandlerFunc
	final := func(w http.ResponseWriter, r *http.Request) {
		// Create a new per-request logger using the configured LoggerConfig
		requestLogger := logger.NewCustomLogger(m.config.Server.Name, m.loggerConfig)
		handler(newMuxContext(w, newRequest(r, m.kafkaClient.writer), m.config, requestLogger).(*Ctx))
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

	if m.kafkaClient == nil {
		return
	}

	for topic, handler := range m.kafkaHandlers {
		kc, err := NewKafkaConsumer(m.kafkaClient.dialer, m.kafkaConfig, ConsumerConfig{
			Topic:   topic,
			Handler: handler,
		})
		if err != nil {
			log.Printf("failed to start kafka consumer for topic %s: %v", topic, err)
			continue
		}

		go kc.Run(ctx, m.createKafkaContext())
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
	}

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      strings.Split(conf.Brokers, ","),
		Dialer:       dialer,
		BatchSize:    conf.BatchSize,
		BatchBytes:   conf.BatchBytes,
		BatchTimeout: time.Duration(conf.BatchTimeout),
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
	m.mu.Lock()

	if !m.isKafkaConfigured() {
		m.mu.Unlock()
		return
	}

	// start the consumer immediately if microservice is running
	if !m.isKafkaConnected() {
		m.mu.Unlock()
		return
	}

	// auto create topics if enabled
	if m.kafkaConfig.AutoCreateTopics {
		m.kafkaClient.CreateTopic(topic, 1)
	}

	if _, exists := m.kafkaHandlers[topic]; exists {
		m.mu.Unlock()
		return
	}

	if m.kafkaClient.reader == nil {
		m.kafkaClient.reader = make(map[string]*kafka.Reader)
	}

	if m.kafkaClient.reader[topic] == nil {
		// Create Kafka consumer
		kc, err := NewKafkaConsumer(m.kafkaClient.dialer, m.kafkaConfig, ConsumerConfig{
			Topic:   topic,
			Handler: handler,
		})
		if err != nil {
			m.mu.Unlock()
			panic(fmt.Sprintf("failed to create kafka consumer for topic %s: %v", topic, err))
		}

		m.kafkaHandlers[topic] = handler
		m.kafkaClient.reader[topic] = kc.reader
	}

	reader := m.kafkaClient.reader[topic]
	m.mu.Unlock()

	// Start consumer loop in goroutine
	go func() {

		for {
			msg, err := reader.FetchMessage(context.Background())
			if err != nil {
				if err != context.Canceled {
					// customLog.Error(logAction.APP_LOGIC("Kafka"), "Kafka consumer read error: "+err.Error())
					log.Printf("Kafka consumer read error: %v", err)
				}
				continue
			}
			customLog := logger.NewCustomLogger(m.config.Server.Name, m.loggerConfig)
			// Create request with msg.Value as body (for ctx.Bind())
			url := &url.URL{Path: "/topic/" + topic}
			url.Host = m.kafkaConfig.Brokers
			url.Scheme = "kafka"

			// /{topic} params("topic")	
				

			req := &http.Request{
				Method: "KAFKA",
				URL:    url,
				Header: http.Header{
					"Content-Type": []string{"application/json"},
				},
				Body: io.NopCloser(bytes.NewReader(msg.Value)),
			}
			req = req.WithContext(context.Background())

			// params
			// params := &KafkaParams{
			// 	Topic:     topic,
			// 	Key:       string(msg.Key),
			// 	Value:     string(msg.Value),
			// 	Partition: msg.Partition,
			// 	Offset:    msg.Offset,
			// 	Message:   msg,
			// }
			// msg.Headers
			// msg.HighWaterMark
			// msg.Key
			// msg.Value
			// msg.Partition
			// msg.Offset
			// msg.Time
			// msg.Topic
			// msg.WriterData

			params := &KafkaParams{
				Headers:       make(map[string]string),
				HighWaterMark: msg.HighWaterMark,
				Key:           string(msg.Key),
				Value:         string(msg.Value),
				Partition:     msg.Partition,
				Offset:        msg.Offset,
				Time:          msg.Time,
				Topic:         msg.Topic,
				WriterData:    msg.WriterData,
			}
			for _, h := range msg.Headers {
				params.Headers[h.Key] = string(h.Value)
			}

			// Add Kafka message details to request context

			// Store Kafka message data in request context
			req = req.WithContext(context.WithValue(req.Context(), CtxKey("kafka.body"), params))
			req = req.WithContext(context.WithValue(req.Context(), CtxKey("kafka.message"), &msg))
			req = req.WithContext(context.WithValue(req.Context(), CtxKey("kafka.topic"), params.Topic))
			req = req.WithContext(context.WithValue(req.Context(), CtxKey("kafka.key"), params.Key))
			req = req.WithContext(context.WithValue(req.Context(), CtxKey("kafka.value"), params.Value))
			req = req.WithContext(context.WithValue(req.Context(), CtxKey("kafka.partition"), params.Partition))
			req = req.WithContext(context.WithValue(req.Context(), CtxKey("kafka.offset"), params.Offset))
			// Extract trace/span from headers if available
			for _, h := range msg.Headers {
				if h.Key == "x-trace-id" {
					req = req.WithContext(context.WithValue(req.Context(), SessionID, string(h.Value)))
				}
				if h.Key == "x-span-id" {
					req = req.WithContext(context.WithValue(req.Context(), TransactionID, string(h.Value)))
				}
			}

			// Use newMuxContext to create proper Ctx (includes validator)
			msgCtx := newMuxContext(nil, newRequest(req, m.kafkaClient.writer), m.config, customLog).(*Ctx)

			// Call handler with panic recovery
			func() {
				defer func() {
					if r := recover(); r != nil {
						customLog.Error(logAction.SYSTEM("error"), fmt.Sprintf("Kafka consumer handler panic: %v", r))
					}
				}()

				handler(msgCtx)
			}()

			// Commit message after successful handling
			if err := reader.CommitMessages(context.Background(), msg); err != nil {
				customLog.Error(logAction.APP_LOGIC("Kafka"), fmt.Sprintf("Failed to commit message: %v", err))
			}
		}
	}()

	log.Printf("Kafka consumer registered for topic: %s with group: %s", topic, m.kafkaConfig.GroupID)
}

// Consume registers a Kafka consumer for the given topic
func (m *Microservice) Consume_v1(topic string, handler KafkaConsumerHandler) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.isKafkaConfigured() {
		return
	}

	// start the consumer immediately if microservice is running
	if !m.isKafkaConnected() {
		return
	}
	// auto create topics if enabled
	if m.kafkaConfig.AutoCreateTopics {
		m.kafkaClient.CreateTopic(topic, 1)
	}

	if _, exists := m.kafkaHandlers[topic]; exists {
		return
	}

	m.mu.Lock()
	if m.kafkaClient.reader == nil {
		m.kafkaClient.reader = make(map[string]*kafka.Reader)
	}
	if m.kafkaClient.reader[topic] == nil {
		// Create Kafka consumer
		kc, err := NewKafkaConsumer(m.kafkaClient.dialer, m.kafkaConfig, ConsumerConfig{
			Topic:   topic,
			Handler: handler,
		})
		if err != nil {
			panic(fmt.Sprintf("failed to create kafka consumer for topic %s: %v", topic, err))
		}

		m.kafkaHandlers[topic] = handler
		m.kafkaClient.reader[topic] = kc.reader
	}
	reader := m.kafkaClient.reader[topic]
	m.mu.Unlock()

	msg, err := reader.FetchMessage(context.Background())
	if err != nil {
		panic(err)
	}
	customLog := logger.NewCustomLogger(m.config.Server.Name, m.loggerConfig)

	requester := &http.Request{
		Method: "",
		URL:    &url.URL{},
	}

	msgCtx := newMuxContext(nil, newRequest(requester, m.kafkaClient.writer), m.config, customLog)
	err = func(ctx *Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("kafka consumer handler panic: %v", r)
				customLog.Error(logAction.SYSTEM("error"), "Kafka consumer handler panic")
			}

		}()

		// handler(ctx)
		m.kafkaHandlers[topic](ctx)
		return nil
	}(msgCtx.(*Ctx))
	if err != nil {
		log.Printf("kafka consumer handler error: %v", err)
		return
	}

	log.Printf("Kafka consumer registered for topic: %s with group: %s", topic, m.kafkaConfig.GroupID)

	reader.CommitMessages(context.Background(), msg)
}
