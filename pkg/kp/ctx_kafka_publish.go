package kp

import (
	"context"
	"encoding/json"
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
