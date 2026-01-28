package kp

import (
	"context"
	"encoding/json"
	"realtime-chat-system/config"
	"realtime-chat-system/pkg/logAction"
	"realtime-chat-system/pkg/logger"

	"github.com/google/uuid"
	"github.com/segmentio/kafka-go"
)

// KafkaCtx is a context for Kafka message handlers
type KafkaCtx struct {
	Cfg       *config.Config
	Log       logger.ICustomLogger
	Message   *kafka.Message
	Topic     string
	Key       string
	Value     string
	Partition int
	Offset    int64
}

// NewKafkaCtx creates a new Kafka context
func NewKafkaCtx(cfg *config.Config, log logger.ICustomLogger, msg *kafka.Message, topic string) *KafkaCtx {
	return &KafkaCtx{
		Cfg:       cfg,
		Log:       log,
		Message:   msg,
		Topic:     topic,
		Key:       string(msg.Key),
		Value:     string(msg.Value),
		Partition: msg.Partition,
		Offset:    msg.Offset,
	}
}

// Context returns the background context
func (kc *KafkaCtx) Context() context.Context {
	return context.Background()
}

// GetMessage returns the Kafka message
func (kc *KafkaCtx) GetMessage() *kafka.Message {
	return kc.Message
}

// GetValue returns the message value as string
func (kc *KafkaCtx) GetValue() string {
	return kc.Value
}

// GetKey returns the message key as string
func (kc *KafkaCtx) GetKey() string {
	return kc.Key
}

// GetTopic returns the topic name
func (kc *KafkaCtx) GetTopic() string {
	return kc.Topic
}

// TraceID returns the trace ID from logger
func (kc *KafkaCtx) TraceID() string {
	return kc.Log.TraceID()
}

// Bind unmarshals the message value as JSON
func (kc *KafkaCtx) Bind(v any) error {
	return json.Unmarshal([]byte(kc.Value), v)
}

// BindJSON is an alias for Bind (unmarshals JSON)
func (kc *KafkaCtx) BindJSON(v any) error {
	return kc.Bind(v)
}

// UnmarshalKey unmarshals the message key as JSON
func (kc *KafkaCtx) UnmarshalKey(v any) error {
	return json.Unmarshal([]byte(kc.Key), v)
}

// UnmarshalValue unmarshals the message value as JSON
func (kc *KafkaCtx) UnmarshalValue(v any) error {
	return json.Unmarshal([]byte(kc.Value), v)
}

// UnmarshalMessage unmarshals both key and value
func (kc *KafkaCtx) UnmarshalMessage(key, value any) (keyErr, valueErr error) {
	keyErr = json.Unmarshal([]byte(kc.Key), key)
	valueErr = json.Unmarshal([]byte(kc.Value), value)
	return
}

// L initializes logger for the Kafka message (shorthand for SetUseCase)
func (kc *KafkaCtx) L(useCase string, masking ...logger.MaskRule) logger.ICustomLogger {
	if useCase == "" {
		useCase = "kafka-consumer-" + kc.Topic
	}
	// get x-session-id from headers
	var sessionID string
	for _, header := range kc.Message.Headers {
		if header.Key == "x-session-id" {
			sessionID = string(header.Value)
			break
		}
	}

	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	// sessionID := x-session-id
	kc.Log.Init(useCase, sessionID, logger.NewSpanID())

	body := make(map[string]any)
	if err := kc.Bind(&body); err != nil {
		body = map[string]any{"raw": kc.Value}
	}

	paramInbound := map[string]any{
		"topic":     kc.Topic,
		"key":       kc.Key,
		"value":     body,
		"partition": kc.Partition,
		"offset":    kc.Offset,
		"headers":   kc.Message.Headers,
	}

	kc.Log.Info(logAction.CONSUMING("consume "+kc.Topic), paramInbound, masking...)
	return kc.Log
}
