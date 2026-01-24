package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"realtime-chat-system/internal/models"
	redisClient "realtime-chat-system/pkg/redis"
)

// UserCacheRepository handles user caching operations
type UserCacheRepository interface {
	CacheUser(ctx context.Context, user *models.User) error
	GetCachedUserByID(ctx context.Context, userID string) (*models.User, error)
	GetCachedUserByUsername(ctx context.Context, username string) (*models.User, error)
	InvalidateUser(ctx context.Context, userID string) error
	InvalidateUserByUsername(ctx context.Context, username string) error
}

type redisUserCacheRepository struct {
	redis *redisClient.Client
}

// NewRedisUserCacheRepository creates a new Redis-based user cache repository
func NewRedisUserCacheRepository(redisClient *redisClient.Client) UserCacheRepository {
	return &redisUserCacheRepository{
		redis: redisClient,
	}
}

const (
	userByIDPrefix       = "user:id:"
	userByUsernamePrefix = "user:username:"
	userCacheTTL         = 15 * time.Minute // User profile rarely changes
)

// CacheUser caches a user by both ID and username
func (r *redisUserCacheRepository) CacheUser(ctx context.Context, user *models.User) error {
	userJSON, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	// Cache by ID
	userIDKey := fmt.Sprintf("%s%s", userByIDPrefix, user.ID)
	if err := r.redis.Set(userIDKey, userJSON, userCacheTTL); err != nil {
		return fmt.Errorf("failed to cache user by ID: %w", err)
	}

	// Cache by username (for login/search optimization)
	usernameKey := fmt.Sprintf("%s%s", userByUsernamePrefix, user.Username)
	if err := r.redis.Set(usernameKey, userJSON, userCacheTTL); err != nil {
		return fmt.Errorf("failed to cache user by username: %w", err)
	}

	return nil
}

// GetCachedUserByID retrieves a cached user by ID
func (r *redisUserCacheRepository) GetCachedUserByID(ctx context.Context, userID string) (*models.User, error) {
	userKey := fmt.Sprintf("%s%s", userByIDPrefix, userID)

	data, err := r.redis.Get(userKey)
	if err != nil {
		return nil, err // Cache miss
	}

	var user models.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

// GetCachedUserByUsername retrieves a cached user by username
func (r *redisUserCacheRepository) GetCachedUserByUsername(ctx context.Context, username string) (*models.User, error) {
	usernameKey := fmt.Sprintf("%s%s", userByUsernamePrefix, username)

	data, err := r.redis.Get(usernameKey)
	if err != nil {
		return nil, err // Cache miss
	}

	var user models.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

// InvalidateUser removes a user from cache (by ID)
func (r *redisUserCacheRepository) InvalidateUser(ctx context.Context, userID string) error {
	userKey := fmt.Sprintf("%s%s", userByIDPrefix, userID)
	return r.redis.Del(userKey)
}

// InvalidateUserByUsername removes a user from cache (by username)
func (r *redisUserCacheRepository) InvalidateUserByUsername(ctx context.Context, username string) error {
	usernameKey := fmt.Sprintf("%s%s", userByUsernamePrefix, username)
	return r.redis.Del(usernameKey)
}

// NoOpUserCacheRepository is a no-op implementation when Redis is not available
type noOpUserCacheRepository struct{}

func NewNoOpUserCacheRepository() UserCacheRepository {
	return &noOpUserCacheRepository{}
}

func (n *noOpUserCacheRepository) CacheUser(ctx context.Context, user *models.User) error {
	return nil
}

func (n *noOpUserCacheRepository) GetCachedUserByID(ctx context.Context, userID string) (*models.User, error) {
	return nil, fmt.Errorf("cache not available")
}

func (n *noOpUserCacheRepository) GetCachedUserByUsername(ctx context.Context, username string) (*models.User, error) {
	return nil, fmt.Errorf("cache not available")
}

func (n *noOpUserCacheRepository) InvalidateUser(ctx context.Context, userID string) error {
	return nil
}

func (n *noOpUserCacheRepository) InvalidateUserByUsername(ctx context.Context, username string) error {
	return nil
}
