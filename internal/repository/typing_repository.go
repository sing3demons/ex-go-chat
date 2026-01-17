package repository

import (
	"context"
	"fmt"
	"time"

	redisClient "realtime-chat-system/pkg/redis"
)

// TypingRepository handles typing indicator operations
type TypingRepository interface {
	SetUserTyping(ctx context.Context, roomID, userID string, isTyping bool) error
	GetTypingUsers(ctx context.Context, roomID string) ([]string, error)
}

type redisTypingRepository struct {
	redis *redisClient.Client
}

// NewRedisTypingRepository creates a new Redis-based typing repository
func NewRedisTypingRepository(redisClient *redisClient.Client) TypingRepository {
	return &redisTypingRepository{
		redis: redisClient,
	}
}

const (
	typingUsersPrefix = "typing:"
	typingTTL         = 10 * time.Second
)

// SetUserTyping sets or removes a user's typing status in a room
func (r *redisTypingRepository) SetUserTyping(ctx context.Context, roomID, userID string, isTyping bool) error {
	typingKey := fmt.Sprintf("%s%s", typingUsersPrefix, roomID)
	
	if isTyping {
		// Add user to typing set with TTL
		if err := r.redis.SAdd(typingKey, userID); err != nil {
			return fmt.Errorf("failed to add user to typing set: %w", err)
		}
		if err := r.redis.Expire(typingKey, typingTTL); err != nil {
			return fmt.Errorf("failed to set TTL for typing indicator: %w", err)
		}
	} else {
		// Remove user from typing set
		if err := r.redis.SRem(typingKey, userID); err != nil {
			return fmt.Errorf("failed to remove user from typing set: %w", err)
		}
	}

	return nil
}

// GetTypingUsers gets users currently typing in a room
func (r *redisTypingRepository) GetTypingUsers(ctx context.Context, roomID string) ([]string, error) {
	typingKey := fmt.Sprintf("%s%s", typingUsersPrefix, roomID)
	return r.redis.SMembers(typingKey)
}