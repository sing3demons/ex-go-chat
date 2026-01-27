package kp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/segmentio/kafka-go"
)

// Publish sends a Kafka message from HTTP handler context
func (c *Ctx) Publish(topic string, key []byte, value []byte) error {
	if c.KafkaWriters == nil {
		return fmt.Errorf("kafka not configured")
	}

	timeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(c.Context(), timeout)
	defer cancel()

	return c.KafkaWriters.WriteMessages(ctx, kafka.Message{
		Topic: topic,
		Key:   key,
		Value: value,
	})
}

// PublishJSON marshals v to JSON and sends to Kafka topic
func (c *Ctx) PublishJSON(topic string, key string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return c.Publish(topic, []byte(key), data)
}

// PublishWithHeaders sends a Kafka message with custom headers
func (c *Ctx) PublishWithHeaders(topic string, key []byte, value []byte, headers map[string]string) error {
	if c.KafkaWriters == nil {
		return fmt.Errorf("kafka not configured")
	}

	kafkaHeaders := make([]kafka.Header, 0, len(headers))
	for k, v := range headers {
		kafkaHeaders = append(kafkaHeaders, kafka.Header{
			Key:   k,
			Value: []byte(v),
		})
	}

	timeout := 10 * time.Second
	ctx, cancel := context.WithTimeout(c.Context(), timeout)
	defer cancel()

	return c.KafkaWriters.WriteMessages(ctx, kafka.Message{
		Topic:   topic,
		Key:     key,
		Value:   value,
		Headers: kafkaHeaders,
	})
}

// PublishJSONWithTracing publishes JSON with trace/span IDs from request
func (c *Ctx) PublishJSONWithTracing(topic string, key string, v any) error {
	headers := map[string]string{
		"x-trace-id": c.TraceID(),
		"x-span-id":  c.SpanID(),
	}

	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	return c.PublishWithHeaders(topic, []byte(key), data, headers)
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
