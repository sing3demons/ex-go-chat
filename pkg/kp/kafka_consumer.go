package kp

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"realtime-chat-system/pkg/logAction"
	"strings"

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
		Dialer:         dialer,
		Brokers:        brokerAddrs,
		Topic:          config.Topic,
		GroupID:        cfg.GroupID,
		StartOffset:    kafka.LastOffset,
		CommitInterval: 1000, // commits every 1000 messages
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

				// Create a new context with message data
				msgCtx := &Ctx{
					Cfg: appCtx.Cfg,
					Log: appCtx.Log,
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

// GetMessage retrieves the Kafka message from context
func (c *Ctx) GetMessage() *kafka.Message {
	if msg, ok := c.Value("kafka.message").(*kafka.Message); ok {
		return msg
	}
	return nil
}

// GetMessageValue retrieves the message value as string
func (c *Ctx) GetMessageValue() string {
	if val, ok := c.Value("kafka.value").(string); ok {
		return val
	}
	return ""
}

// GetMessageKey retrieves the message key as string
func (c *Ctx) GetMessageKey() string {
	if key, ok := c.Value("kafka.key").(string); ok {
		return key
	}
	return ""
}

// GetTopic retrieves the topic name
func (c *Ctx) GetTopic() string {
	if topic, ok := c.Value("kafka.topic").(string); ok {
		return topic
	}
	return ""
}

// GetPartition retrieves the partition number
func (c *Ctx) GetPartition() int {
	if partition, ok := c.Value("kafka.partition").(int); ok {
		return partition
	}
	return -1
}

// GetOffset retrieves the message offset
func (c *Ctx) GetOffset() int64 {
	if offset, ok := c.Value("kafka.offset").(int64); ok {
		return offset
	}
	return -1
}

// BindMessageValue binds the message value to a struct
func (c *Ctx) BindMessageValue(v any) error {
	msgValue := c.GetMessageValue()
	if msgValue == "" {
		// return NewError(400, "empty_message", "message value is empty")
		return &Error{
			StatusCode: 400,
			Message:    "empty_message",
			Err:        errors.New("message value is empty"),
		}
	}

	if err := json.Unmarshal([]byte(msgValue), v); err != nil {
		return &Error{
			StatusCode: 400,
			Message:    "invalid_json",
			Err:        errors.New(err.Error()),
		}
	}

	return nil
}
