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
	dependency := "redis"
	log := mlog.L(ctx)
	start := time.Now()

	userJSON, err := json.Marshal(user)
	if err != nil {
		return fmt.Errorf("failed to marshal user: %w", err)
	}

	// Cache by ID
	userIDKey := fmt.Sprintf("%s%s", userByIDPrefix, user.ID)

	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_CREATE, "redis set user"), map[string]any{
		"key":        userIDKey,
		"value":      user,
		"expiration": userCacheTTL,
	})
	err = r.redis.Set(userIDKey, userJSON, userCacheTTL)
	end := time.Since(start).Milliseconds()
	if err != nil {
		log.SetDependencyMetadata(logger.DependencyMetadata{
			Dependency:   dependency,
			ResponseTime: end,
		}).Debug(logAction.DB_RESPONSE(logAction.DB_CREATE, "redis set user"), map[string]any{
			"status": "error",
			"error":  err.Error(),
		})
		return fmt.Errorf("failed to cache user by ID: %w", err)
	}
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency:   dependency,
		ResponseTime: end,
	}).Debug(logAction.DB_RESPONSE(logAction.DB_CREATE, "redis set user"), map[string]any{
		"status": "success",
	})

	// Cache by username (for login/search optimization)
	start = time.Now()
	usernameKey := fmt.Sprintf("%s%s", userByUsernamePrefix, user.Username)
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_CREATE, "redis set user by username"), map[string]any{
		"key":        usernameKey,
		"value":      user,
		"expiration": userCacheTTL,
	})
	err = r.redis.Set(usernameKey, userJSON, userCacheTTL)
	end = time.Since(start).Milliseconds()
	result := map[string]any{}
	if err != nil {
		result["status"] = "error"
		result["error"] = err.Error()
	} else {
		result["status"] = "success"
	}
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency:   dependency,
		ResponseTime: end,
	}).Debug(logAction.DB_RESPONSE(logAction.DB_CREATE, "redis set user by username"), result)

	return err
}

// GetCachedUserByID retrieves a cached user by ID
func (r *redisUserCacheRepository) GetCachedUserByID(ctx context.Context, userID string) (*models.User, error) {
	dependency := "redis"
	log := mlog.L(ctx)
	start := time.Now()

	userKey := fmt.Sprintf("%s%s", userByIDPrefix, userID)

	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_READ, "redis get user"), map[string]any{
		"key": userKey,
	})
	data, err := r.redis.Get(userKey)
	end := time.Since(start).Milliseconds()
	result := map[string]any{}
	if err != nil {
		result["status"] = "error"
		result["error"] = err.Error()
		log.SetDependencyMetadata(logger.DependencyMetadata{
			Dependency:   dependency,
			ResponseTime: end,
		}).Debug(logAction.DB_RESPONSE(logAction.DB_READ, "redis get user"), result)
		return nil, err // Cache miss
	}
	result["status"] = "success"
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency:   dependency,
		ResponseTime: end,
	}).Debug(logAction.DB_RESPONSE(logAction.DB_READ, "redis get user"), result)

	var user models.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

// GetCachedUserByUsername retrieves a cached user by username
func (r *redisUserCacheRepository) GetCachedUserByUsername(ctx context.Context, username string) (*models.User, error) {
	dependency := "redis"
	log := mlog.L(ctx)
	start := time.Now()
	usernameKey := fmt.Sprintf("%s%s", userByUsernamePrefix, username)

	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_READ, "redis get user by username"), map[string]any{
		"key": usernameKey,
	})
	data, err := r.redis.Get(usernameKey)
	end := time.Since(start).Milliseconds()
	result := map[string]any{}
	if err != nil {
		result["status"] = "error"
		result["error"] = err.Error()
		log.SetDependencyMetadata(logger.DependencyMetadata{
			Dependency:   dependency,
			ResponseTime: end,
		}).Debug(logAction.DB_RESPONSE(logAction.DB_READ, "redis get user by username"), result)
		return nil, err // Cache miss
	}
	result["status"] = "success"
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency:   dependency,
		ResponseTime: end,
	}).Debug(logAction.DB_RESPONSE(logAction.DB_READ, "redis get user by username"), result)

	var user models.User
	if err := json.Unmarshal([]byte(data), &user); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user: %w", err)
	}

	return &user, nil
}

// InvalidateUser removes a user from cache (by ID)
func (r *redisUserCacheRepository) InvalidateUser(ctx context.Context, userID string) error {
	dependency := "redis"
	log := mlog.L(ctx)
	start := time.Now()

	userKey := fmt.Sprintf("%s%s", userByIDPrefix, userID)

	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_DELETE, "redis del user"), map[string]any{
		"key": userKey,
	})
	err := r.redis.Del(userKey)
	end := time.Since(start).Milliseconds()
	result := map[string]any{}
	if err != nil {
		result["status"] = "error"
		result["error"] = err.Error()
		log.SetDependencyMetadata(logger.DependencyMetadata{
			Dependency:   dependency,
			ResponseTime: end,
		}).Debug(logAction.DB_RESPONSE(logAction.DB_DELETE, "redis del user"), result)
		return err
	}
	result["status"] = "success"
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency:   dependency,
		ResponseTime: end,
	}).Debug(logAction.DB_RESPONSE(logAction.DB_DELETE, "redis del user"), result)

	return nil
}
// InvalidateUserByUsername removes a user from cache (by username)
func (r *redisUserCacheRepository) InvalidateUserByUsername(ctx context.Context, username string) error {
	dependency := "redis"
	log := mlog.L(ctx)
	start := time.Now()

	usernameKey := fmt.Sprintf("%s%s", userByUsernamePrefix, username)

	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_DELETE, "redis del user by username"), map[string]any{
		"key": usernameKey,
	})
	err := r.redis.Del(usernameKey)
	end := time.Since(start).Milliseconds()
	result := map[string]any{}
	if err != nil {
		result["status"] = "error"
		result["error"] = err.Error()
		log.SetDependencyMetadata(logger.DependencyMetadata{
			Dependency:   dependency,
			ResponseTime: end,
		}).Debug(logAction.DB_RESPONSE(logAction.DB_DELETE, "redis del user by username"), result)
		return err
	}
	result["status"] = "success"
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency:   dependency,
		ResponseTime: end,
	}).Debug(logAction.DB_RESPONSE(logAction.DB_DELETE, "redis del user by username"), result)

	return nil
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
