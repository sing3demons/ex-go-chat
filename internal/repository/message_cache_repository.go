package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"realtime-chat-system/internal/models"
	redisClient "realtime-chat-system/pkg/redis"

	"github.com/redis/go-redis/v9"
)

// MessageCacheRepository handles message caching operations
type MessageCacheRepository interface {
	CacheMessage(ctx context.Context, message *models.Message) error
	GetCachedMessage(ctx context.Context, messageID string) (*models.Message, error)
	GetCachedRoomMessages(ctx context.Context, roomID string, limit int64) ([]*models.Message, error)
	InvalidateMessage(ctx context.Context, messageID, roomID string) error
}

type redisMessageCacheRepository struct {
	redis *redisClient.Client
}

// NewRedisMessageCacheRepository creates a new Redis-based message cache repository
func NewRedisMessageCacheRepository(redisClient *redisClient.Client) MessageCacheRepository {
	return &redisMessageCacheRepository{
		redis: redisClient,
	}
}

const (
	messageKeyPrefix   = "message:"
	roomMessagesPrefix = "room_messages:"
	messageCacheTTL    = 1 * time.Hour
)

// CacheMessage caches a message
func (r *redisMessageCacheRepository) CacheMessage(ctx context.Context, message *models.Message) error {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Cache individual message
	messageKey := fmt.Sprintf("%s%s", messageKeyPrefix, message.ID)
	if err := r.redis.Set(messageKey, messageJSON, messageCacheTTL); err != nil {
		return fmt.Errorf("failed to cache message: %w", err)
	}

	// Add to room messages sorted set (sorted by timestamp)
	roomKey := fmt.Sprintf("%s%s", roomMessagesPrefix, message.RoomID)
	score := float64(message.CreatedAt.Unix())
	if err := r.redis.ZAdd(roomKey, redis.Z{Score: score, Member: message.ID}); err != nil {
		return fmt.Errorf("failed to add message to room cache: %w", err)
	}

	// Set TTL for room messages
	if err := r.redis.Expire(roomKey, messageCacheTTL); err != nil {
		return fmt.Errorf("failed to set TTL for room messages: %w", err)
	}

	return nil
}

// GetCachedMessage retrieves a cached message
func (r *redisMessageCacheRepository) GetCachedMessage(ctx context.Context, messageID string) (*models.Message, error) {
	messageKey := fmt.Sprintf("%s%s", messageKeyPrefix, messageID)
	messageJSON, err := r.redis.Get(messageKey)
	if err != nil {
		return nil, err
	}

	var message models.Message
	if err := json.Unmarshal([]byte(messageJSON), &message); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &message, nil
}

// GetCachedRoomMessages retrieves cached messages for a room
func (r *redisMessageCacheRepository) GetCachedRoomMessages(ctx context.Context, roomID string, limit int64) ([]*models.Message, error) {
	roomKey := fmt.Sprintf("%s%s", roomMessagesPrefix, roomID)
	
	// Get message IDs in reverse chronological order (newest first)
	messageIDs, err := r.redis.ZRevRange(roomKey, 0, limit-1)
	if err != nil {
		return nil, err
	}

	messages := make([]*models.Message, 0, len(messageIDs))
	for _, messageID := range messageIDs {
		message, err := r.GetCachedMessage(ctx, messageID)
		if err != nil {
			// If message not found in cache, skip it
			continue
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// InvalidateMessage removes a message from cache
func (r *redisMessageCacheRepository) InvalidateMessage(ctx context.Context, messageID, roomID string) error {
	// Remove individual message
	messageKey := fmt.Sprintf("%s%s", messageKeyPrefix, messageID)
	if err := r.redis.Del(messageKey); err != nil {
		return fmt.Errorf("failed to delete message from cache: %w", err)
	}

	// Remove from room messages
	roomKey := fmt.Sprintf("%s%s", roomMessagesPrefix, roomID)
	if err := r.redis.ZRem(roomKey, messageID); err != nil {
		return fmt.Errorf("failed to remove message from room cache: %w", err)
	}

	return nil
}