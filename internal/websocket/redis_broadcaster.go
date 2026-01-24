package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"realtime-chat-system/pkg/logger"

	"github.com/redis/go-redis/v9"
)

// RedisBroadcaster handles publishing and subscribing to messages via Redis Pub/Sub
type RedisBroadcaster struct {
	client        *redis.Client
	serverID      string // Unique identifier for this server instance
	subscriptions map[string]*redis.PubSub
	subMu         sync.RWMutex
	log           *logger.Logger

	// Channel to receive messages from Redis
	IncomingBroadcast chan *DistributedBroadcast

	// WaitGroup for graceful shutdown
	wg sync.WaitGroup
}

// DistributedBroadcast represents a broadcast message from Redis
type DistributedBroadcast struct {
	Type      string     `json:"type"` // "room" or "user"
	ServerID  string     `json:"serverId"`
	RoomID    string     `json:"roomId,omitempty"`
	UserID    string     `json:"userId,omitempty"`
	Message   *WSMessage `json:"message"`
	Exclude   string     `json:"exclude,omitempty"`
	Timestamp time.Time  `json:"timestamp"`
}

// NewRedisBroadcaster creates a new Redis broadcaster
func NewRedisBroadcaster(redisClient *redis.Client, serverID string, log *logger.Logger) *RedisBroadcaster {
	return &RedisBroadcaster{
		client:            redisClient,
		serverID:          serverID,
		subscriptions:     make(map[string]*redis.PubSub),
		log:               log,
		IncomingBroadcast: make(chan *DistributedBroadcast, 256),
	}
}

// PublishToRoom publishes a message to a room channel
func (rb *RedisBroadcaster) PublishToRoom(ctx context.Context, roomID string, msg *WSMessage, exclude string) error {
	broadcast := &DistributedBroadcast{
		Type:      "room",
		ServerID:  rb.serverID,
		RoomID:    roomID,
		Message:   msg,
		Exclude:   exclude,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(broadcast)
	if err != nil {
		rb.log.Errorf("Failed to marshal room broadcast: %v", err)
		return err
	}

	channelName := fmt.Sprintf("room:%s", roomID)
	if err := rb.client.Publish(ctx, channelName, data).Err(); err != nil {
		rb.log.Errorf("Failed to publish to room channel %s: %v", channelName, err)
		return err
	}

	rb.log.Debugf("Published message to room %s via Redis", roomID)
	return nil
}

// PublishToUser publishes a message to a specific user
func (rb *RedisBroadcaster) PublishToUser(ctx context.Context, userID string, msg *WSMessage) error {
	broadcast := &DistributedBroadcast{
		Type:      "user",
		ServerID:  rb.serverID,
		UserID:    userID,
		Message:   msg,
		Timestamp: time.Now(),
	}

	data, err := json.Marshal(broadcast)
	if err != nil {
		rb.log.Errorf("Failed to marshal user broadcast: %v", err)
		return err
	}

	channelName := fmt.Sprintf("user:%s", userID)
	if err := rb.client.Publish(ctx, channelName, data).Err(); err != nil {
		rb.log.Errorf("Failed to publish to user channel %s: %v", channelName, err)
		return err
	}

	rb.log.Debugf("Published message to user %s via Redis", userID)
	return nil
}

// SubscribeToRoomBroadcasts subscribes to room broadcast channel
func (rb *RedisBroadcaster) SubscribeToRoomBroadcasts(ctx context.Context, roomID string) error {
	rb.subMu.Lock()
	defer rb.subMu.Unlock()

	channelName := fmt.Sprintf("room:%s", roomID)

	// Check if already subscribed
	if _, exists := rb.subscriptions[channelName]; exists {
		return nil
	}

	pubsub := rb.client.Subscribe(ctx, channelName)
	rb.subscriptions[channelName] = pubsub

	// Start listening in a goroutine
	rb.wg.Add(1)
	go rb.listenToChannel(channelName, pubsub)

	rb.log.Debugf("Subscribed to room broadcasts: %s", roomID)
	return nil
}

// SubscribeToUserBroadcasts subscribes to user broadcast channel
func (rb *RedisBroadcaster) SubscribeToUserBroadcasts(ctx context.Context, userID string) error {
	rb.subMu.Lock()
	defer rb.subMu.Unlock()

	channelName := fmt.Sprintf("user:%s", userID)

	// Check if already subscribed
	if _, exists := rb.subscriptions[channelName]; exists {
		return nil
	}

	pubsub := rb.client.Subscribe(ctx, channelName)
	rb.subscriptions[channelName] = pubsub

	// Start listening in a goroutine
	rb.wg.Add(1)
	go rb.listenToChannel(channelName, pubsub)

	rb.log.Debugf("Subscribed to user broadcasts: %s", userID)
	return nil
}

// UnsubscribeFromRoom unsubscribes from a room channel
func (rb *RedisBroadcaster) UnsubscribeFromRoom(roomID string) error {
	rb.subMu.Lock()
	defer rb.subMu.Unlock()

	channelName := fmt.Sprintf("room:%s", roomID)
	if pubsub, exists := rb.subscriptions[channelName]; exists {
		pubsub.Close()
		delete(rb.subscriptions, channelName)
		rb.log.Debugf("Unsubscribed from room broadcasts: %s", roomID)
	}

	return nil
}

// UnsubscribeFromUser unsubscribes from a user channel
func (rb *RedisBroadcaster) UnsubscribeFromUser(userID string) error {
	rb.subMu.Lock()
	defer rb.subMu.Unlock()

	channelName := fmt.Sprintf("user:%s", userID)
	if pubsub, exists := rb.subscriptions[channelName]; exists {
		pubsub.Close()
		delete(rb.subscriptions, channelName)
		rb.log.Debugf("Unsubscribed from user broadcasts: %s", userID)
	}

	return nil
}

// listenToChannel listens for messages on a Redis channel
func (rb *RedisBroadcaster) listenToChannel(channelName string, pubsub *redis.PubSub) {
	defer rb.wg.Done()

	ch := pubsub.Channel()

	for msg := range ch {
		var broadcast DistributedBroadcast
		if err := json.Unmarshal([]byte(msg.Payload), &broadcast); err != nil {
			rb.log.Errorf("Failed to unmarshal broadcast from Redis: %v", err)
			continue
		}

		// Forward to hub's incoming channel
		select {
		case rb.IncomingBroadcast <- &broadcast:
			// Sent successfully
		default:
			rb.log.Warnf("IncomingBroadcast channel full, dropping message")
		}
	}
}

// Close closes all subscriptions and stops listening
func (rb *RedisBroadcaster) Close() error {
	rb.subMu.Lock()
	defer rb.subMu.Unlock()

	for channelName, pubsub := range rb.subscriptions {
		if err := pubsub.Close(); err != nil {
			rb.log.Errorf("Failed to close subscription %s: %v", channelName, err)
		}
	}

	rb.subscriptions = make(map[string]*redis.PubSub)
	close(rb.IncomingBroadcast)

	// Wait for all goroutines to finish
	rb.wg.Wait()

	rb.log.Infof("Redis broadcaster closed")
	return nil
}

// GetServerID returns the server instance ID
func (rb *RedisBroadcaster) GetServerID() string {
	return rb.serverID
}
