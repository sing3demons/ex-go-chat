package kp

import (
	"sync"

	"github.com/segmentio/kafka-go"
)

type multiConn struct {
	conns  []*kafka.Conn
	dialer *kafka.Dialer
	mu     sync.RWMutex
}

// IKafkaClient defines the interface for Kafka client operations
type IKafkaClient interface {
	Close()
	CreateTopic(topic string, partitions int) error
	GetDialer() *kafka.Dialer
	GetWriter() Writer
}

type kafkaClient struct {
	dialer *kafka.Dialer
	conn   *multiConn

	writer Writer
	reader map[string]*kafka.Reader

	mu *sync.RWMutex

	config KafkaConfig
}

func (kc *kafkaClient) GetDialer() *kafka.Dialer {
	return kc.dialer
}

func (kc *kafkaClient) GetWriter() Writer {
	kc.mu.RLock()
	defer kc.mu.RUnlock()
	return kc.writer
}

func (kc *kafkaClient) Close() {
	kc.mu.Lock()
	defer kc.mu.Unlock()

	if kc.writer != nil {
		kc.writer.Close()
	}

	for _, r := range kc.reader {
		if r != nil {
			r.Close()
		}
	}

	if kc.conn != nil {
		kc.conn.mu.Lock()
		for _, c := range kc.conn.conns {
			c.Close()
		}
		kc.conn.mu.Unlock()
	}
}

func (kc *kafkaClient) CreateTopic(topic string, partitions int) error {
	kc.mu.RLock()
	defer kc.mu.RUnlock()

	if kc.conn == nil {
		return nil
	}

	for _, c := range kc.conn.conns {
		err := c.CreateTopics(kafka.TopicConfig{
			Topic:             topic,
			NumPartitions:     partitions,
			ReplicationFactor: 1,
		})
		if err != nil {
			return err
		}
	}

	return nil

}
