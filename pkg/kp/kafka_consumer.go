package kp

import (
	"context"
	"errors"
	"log"
	"net/http"
	"realtime-chat-system/pkg/logAction"
	"strings"
	"time"

	"github.com/segmentio/kafka-go"
)

// KafkaConsumerHandler is a handler for Kafka messages
type KafkaConsumerHandler func(ctx *Ctx)

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

				// Create a fake HTTP request with Kafka message as body
				req, _ := http.NewRequestWithContext(
					ctx,
					"POST",
					"http://kafka/"+kc.config.Topic,
					strings.NewReader(string(msg.Value)),
				)
				req.Header.Set("Content-Type", "application/json")

				// Create a new context with message data
				msgCtx := &Ctx{
					Cfg: appCtx.Cfg,
					Log: appCtx.Log,
					Req: req,
				}

				// Store Kafka message in context
				msgCtx.Set("kafka.message", &msg)
				msgCtx.Set("kafka.topic", kc.config.Topic)
				msgCtx.Set("kafka.key", string(msg.Key))
				msgCtx.Set("kafka.value", string(msg.Value))
				msgCtx.Set("kafka.partition", msg.Partition)
				msgCtx.Set("kafka.offset", msg.Offset)

				// Call handler
				kc.config.Handler(msgCtx)
			}
		}
	}()
}
