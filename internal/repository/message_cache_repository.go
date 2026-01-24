package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"realtime-chat-system/internal/models"
	"realtime-chat-system/pkg/logAction"
	"realtime-chat-system/pkg/logger"
	"realtime-chat-system/pkg/mlog"
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
	dependency := "redis"
	log := mlog.L(ctx)
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Cache individual message
	messageKey := fmt.Sprintf("%s%s", messageKeyPrefix, message.ID)
	start := time.Now()
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_CREATE, "redis set message"), map[string]any{
		"key":        messageKey,
		"value":      message,
		"expiration": messageCacheTTL,
	})

	err = r.redis.Set(messageKey, messageJSON, messageCacheTTL)
	end := time.Since(start).Milliseconds()
	if err != nil {
		log.SetDependencyMetadata(logger.DependencyMetadata{
			Dependency:   dependency,
			ResponseTime: end,
		}).Debug(logAction.DB_RESPONSE(logAction.DB_CREATE, "redis set message"), map[string]any{
			"status": "error",
			"error":  err.Error(),
		})

		return fmt.Errorf("failed to cache message: %w", err)
	}

	// Add to room messages sorted set (sorted by timestamp)
	roomKey := fmt.Sprintf("%s%s", roomMessagesPrefix, message.RoomID)
	score := float64(message.CreatedAt.Unix())

	start = time.Now()
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_CREATE, "redis zadd room message"), map[string]any{
		"key":    roomKey,
		"score":  score,
		"member": message.ID,
	})
	err = r.redis.ZAdd(roomKey, redis.Z{Score: score, Member: message.ID})
	end = time.Since(start).Milliseconds()
	if err != nil {
		log.SetDependencyMetadata(logger.DependencyMetadata{
			Dependency:   dependency,
			ResponseTime: end,
		}).Debug(logAction.DB_RESPONSE(logAction.DB_CREATE, "redis zadd room message"), map[string]any{
			"status": "error",
			"error":  err.Error(),
		})
		return fmt.Errorf("failed to add message to room cache: %w", err)
	}
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency:   dependency,
		ResponseTime: end,
	}).Debug(logAction.DB_RESPONSE(logAction.DB_CREATE, "redis zadd room message"), map[string]any{
		"status": "success",
	})

	// Set TTL for room messages
	start = time.Now()
	err = r.redis.Expire(roomKey, messageCacheTTL)
	end = time.Since(start).Milliseconds()
	if err != nil {
		log.SetDependencyMetadata(logger.DependencyMetadata{
			Dependency:   dependency,
			ResponseTime: end,
		}).Debug(logAction.DB_RESPONSE(logAction.DB_CREATE, "redis expire room messages"), map[string]any{
			"status": "error",
			"error":  err.Error(),
		})
		return fmt.Errorf("failed to set TTL for room messages: %w", err)
	}
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency:   dependency,
		ResponseTime: end,
	}).Debug(logAction.DB_RESPONSE(logAction.DB_CREATE, "redis expire room messages"), map[string]any{
		"status": "success",
	})

	return nil
}

// GetCachedMessage retrieves a cached message
func (r *redisMessageCacheRepository) GetCachedMessage(ctx context.Context, messageID string) (*models.Message, error) {
	dependency := "redis"
	log := mlog.L(ctx)
	start := time.Now()
	messageKey := fmt.Sprintf("%s%s", messageKeyPrefix, messageID)
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_READ, "redis get message"), map[string]any{
		"key": messageKey,
	})

	messageJSON, err := r.redis.Get(messageKey)
	end := time.Since(start).Milliseconds()
	result := map[string]any{}
	if err != nil {
		result["status"] = "error"
		result["error"] = err.Error()
		log.SetDependencyMetadata(logger.DependencyMetadata{
			Dependency:   dependency,
			ResponseTime: end,
		}).Debug(logAction.DB_RESPONSE(logAction.DB_READ, "redis get message"), result)
		return nil, err
	}
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency:   dependency,
		ResponseTime: end,
	}).Debug(logAction.DB_RESPONSE(logAction.DB_READ, "redis get message"), map[string]any{
		"status": "success",
	})

	var message models.Message
	if err := json.Unmarshal([]byte(messageJSON), &message); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return &message, nil
}

// GetCachedRoomMessages retrieves cached messages for a room
func (r *redisMessageCacheRepository) GetCachedRoomMessages(ctx context.Context, roomID string, limit int64) ([]*models.Message, error) {
	roomKey := fmt.Sprintf("%s%s", roomMessagesPrefix, roomID)

	// log Redis operations
	dependency := "redis"
	log := mlog.L(ctx)
	start := time.Now()
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_READ, "redis zrevrange room messages"), map[string]any{
		"key":   roomKey,
		"limit": limit,
	})
	// Get message IDs in reverse chronological order (newest first)
	messageIDs, err := r.redis.ZRevRange(roomKey, 0, limit-1)
	end := time.Since(start).Milliseconds()
	result := map[string]any{}
	if err != nil {
		result["status"] = "error"
		result["error"] = err.Error()
		log.SetDependencyMetadata(logger.DependencyMetadata{
			Dependency:   dependency,
			ResponseTime: end,
		}).Debug(logAction.DB_RESPONSE(logAction.DB_READ, "redis zrevrange room messages"), result)
		return nil, err
	}
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency:   dependency,
		ResponseTime: end,
	}).Debug(logAction.DB_RESPONSE(logAction.DB_READ, "redis zrevrange room messages"), map[string]any{
		"status":      "success",
		"message_ids": messageIDs,
	})

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
	dependency := "redis"
	log := mlog.L(ctx)
	start := time.Now()

	// Remove individual message
	messageKey := fmt.Sprintf("%s%s", messageKeyPrefix, messageID)

	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_DELETE, "redis del message"), map[string]any{
		"key": messageKey,
	})
	err := r.redis.Del(messageKey)
	end := time.Since(start).Milliseconds()
	result := map[string]any{}
	if err != nil {
		result["status"] = "error"
		result["error"] = err.Error()
		log.SetDependencyMetadata(logger.DependencyMetadata{
			Dependency:   dependency,
			ResponseTime: end,
		}).Debug(logAction.DB_RESPONSE(logAction.DB_DELETE, "redis del message"), result)
		return fmt.Errorf("failed to delete message from cache: %w", err)
	}
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency:   dependency,
		ResponseTime: end,
	}).Debug(logAction.DB_RESPONSE(logAction.DB_DELETE, "redis del message"), map[string]any{
		"status": "success",
	})

	// Remove from room messages
	roomKey := fmt.Sprintf("%s%s", roomMessagesPrefix, roomID)

	start = time.Now()
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_DELETE, "redis zrem room message"), map[string]any{
		"key":    roomKey,
		"member": messageID,
	})
	err = r.redis.ZRem(roomKey, messageID)
	end = time.Since(start).Milliseconds()
	result = map[string]any{}
	if err != nil {
		result["status"] = "error"
		result["error"] = err.Error()
	} else {
		result["status"] = "success"
	}
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency:   dependency,
		ResponseTime: end,
	}).Debug(logAction.DB_RESPONSE(logAction.DB_DELETE, "redis zrem room message"), result)

	return err
}
