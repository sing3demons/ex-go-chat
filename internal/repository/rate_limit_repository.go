package repository

import (
	"context"
	"fmt"
	"time"

	redisClient "realtime-chat-system/pkg/redis"
)

// RateLimitRepository handles rate limiting operations
type RateLimitRepository interface {
	CheckRateLimit(ctx context.Context, userID string, action string, limit int, window time.Duration) (bool, error)
}

type redisRateLimitRepository struct {
	redis *redisClient.Client
}

// NewRedisRateLimitRepository creates a new Redis-based rate limit repository
func NewRedisRateLimitRepository(redisClient *redisClient.Client) RateLimitRepository {
	return &redisRateLimitRepository{
		redis: redisClient,
	}
}

// CheckRateLimit checks if an action is within rate limits
func (r *redisRateLimitRepository) CheckRateLimit(ctx context.Context, userID string, action string, limit int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("rate_limit:%s:%s", userID, action)
	
	// Get current count
	countStr, err := r.redis.Get(key)
	if err != nil && err.Error() != "redis: nil" {
		return false, fmt.Errorf("failed to get rate limit count: %w", err)
	}
	
	count := 0
	if countStr != "" {
		if _, err := fmt.Sscanf(countStr, "%d", &count); err != nil {
			count = 0
		}
	}
	
	if count >= limit {
		return false, nil // Rate limit exceeded
	}
	
	// Increment counter
	count++
	if err := r.redis.Set(key, fmt.Sprintf("%d", count), window); err != nil {
		return false, fmt.Errorf("failed to set rate limit count: %w", err)
	}
	
	return true, nil
}