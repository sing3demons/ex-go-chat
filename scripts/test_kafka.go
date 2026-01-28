package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/segmentio/kafka-go"
)

func productKafka(n int) {
	// Connect to Kafka brokers
	brokers := []string{"localhost:29092"}
	topic := "test"

	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:      brokers,
		Topic:        topic,
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
		ReadTimeout:  10 * time.Second,
	})
	defer writer.Close()

	// Use WaitGroup to wait for all goroutines
	var wg sync.WaitGroup
	successCount := 0
	errorCount := 0
	mu := sync.Mutex{}

	// Produce n messages concurrently
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()

			message := map[string]any{
				"id":        index + 1,
				"timestamp": time.Now().Unix(),
				"data":      fmt.Sprintf("Test message #%d", index+1),
			}

			value, err := json.Marshal(message)
			if err != nil {
				log.Printf("Error marshaling message %d: %v", index, err)
				mu.Lock()
				errorCount++
				mu.Unlock()
				return
			}

			err = writer.WriteMessages(context.Background(), kafka.Message{
				Key:   []byte(fmt.Sprintf("key-%d", index+1)),
				Value: value,
				Headers: []kafka.Header{
					{Key: "x-session-id", Value: []byte(fmt.Sprintf("%d-session-%d", index+1, time.Now().UnixNano()))},
				},
			})

			if err != nil {
				log.Printf("Error writing message %d: %v", index, err)
				mu.Lock()
				errorCount++
				mu.Unlock()
			} else {
				log.Printf("Produced message %d: %s", index+1, string(value))
				mu.Lock()
				successCount++
				mu.Unlock()
			}
		}(i)
	}

	// Wait for all goroutines to complete
	wg.Wait()

	log.Printf("Successfully produced %d messages, %d errors out of %d total", successCount, errorCount, n)
}

func main() {
	// Produce 10000 messages
	productKafka(1)
}
