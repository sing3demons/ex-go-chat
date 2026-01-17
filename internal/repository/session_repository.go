package repository

import (
	"context"
	"fmt"
	"time"

	redisClient "realtime-chat-system/pkg/redis"
)

// SessionRepository handles user session operations
type SessionRepository interface {
	SetUserSession(ctx context.Context, userID, sessionID string) error
	GetUserSession(ctx context.Context, userID string) (string, error)
	InvalidateUserSession(ctx context.Context, userID string) error
}

type redisSessionRepository struct {
	redis *redisClient.Client
}

// NewRedisSessionRepository creates a new Redis-based session repository
func NewRedisSessionRepository(redisClient *redisClient.Client) SessionRepository {
	return &redisSessionRepository{
		redis: redisClient,
	}
}

const (
	sessionTTL = 24 * time.Hour
)

// SetUserSession sets a user session
func (r *redisSessionRepository) SetUserSession(ctx context.Context, userID, sessionID string) error {
	sessionKey := fmt.Sprintf("session:%s", userID)
	return r.redis.Set(sessionKey, sessionID, sessionTTL)
}

// GetUserSession gets a user session
func (r *redisSessionRepository) GetUserSession(ctx context.Context, userID string) (string, error) {
	sessionKey := fmt.Sprintf("session:%s", userID)
	return r.redis.Get(sessionKey)
}

// InvalidateUserSession removes a user session
func (r *redisSessionRepository) InvalidateUserSession(ctx context.Context, userID string) error {
	sessionKey := fmt.Sprintf("session:%s", userID)
	return r.redis.Del(sessionKey)
}