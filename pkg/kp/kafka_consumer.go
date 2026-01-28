package kp

import (
	"context"
	"errors"
	"log"
	"realtime-chat-system/pkg/logAction"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaConsumerHandler is a handler for Kafka messages
type KafkaConsumerHandler func(ctx *KafkaCtx)

// ConsumerConfig holds Kafka consumer configuration
type ConsumerConfig struct {
	Topic   string
	Handler KafkaConsumerHandler
}

// KafkaConsumer wraps Kafka consumer operations
type KafkaConsumer struct {
	reader *kafka.Reader
	config ConsumerConfig
	log    log.Logger
}

// NewKafkaConsumer creates a new Kafka consumer
func NewKafkaConsumer(dialer *kafka.Dialer, cfg *KafkaConfig, config ConsumerConfig) (*KafkaConsumer, error) {
	brokerAddrs := parseBrokers(cfg.Brokers)
	if len(brokerAddrs) == 0 {
		return nil, errors.New("no kafka brokers provided")
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Dialer:      dialer,
		Brokers:     brokerAddrs,
		Topic:       config.Topic,
		GroupID:     cfg.GroupID,
		StartOffset: kafka.LastOffset,
		// Commit offsets every second; adjust as needed.
		CommitInterval: time.Second,
	})

	return &KafkaConsumer{
		reader: reader,
		config: config,
	}, nil
}

func parseBrokers(brokers string) []string {
	parts := strings.Split(brokers, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if trimmed := strings.TrimSpace(p); trimmed != "" {
			out = append(out, trimmed)
		}
	}
	return out
}

// Close closes the consumer
func (kc *KafkaConsumer) Close() error {
	return kc.reader.Close()
}

// Run starts consuming messages
func (kc *KafkaConsumer) Run(ctx context.Context, appCtx *Ctx) {
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			default:
				msg, err := kc.reader.ReadMessage(ctx)
				if err != nil {
					if err != context.Canceled && err != context.DeadlineExceeded {
						appCtx.Log.Error(logAction.APP_LOGIC("Kafka"), "Kafka consumer read error: "+err.Error())
					}
					continue
				}

				// Create a Kafka context (not HTTP request)
				msgCtx := NewKafkaCtx(appCtx.Cfg, appCtx.Log, &msg, kc.config.Topic)

				// Call handler
				kc.config.Handler(msgCtx)
			}
		}
	}()
}
