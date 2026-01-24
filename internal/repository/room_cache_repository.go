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

// RoomCacheRepository handles room caching operations
type RoomCacheRepository interface {
	CacheRoom(ctx context.Context, room *models.Room) error
	GetCachedRoom(ctx context.Context, roomID string) (*models.Room, error)
	InvalidateRoom(ctx context.Context, roomID string) error
	CacheUserRooms(ctx context.Context, userID string, rooms []*models.Room) error
	GetCachedUserRooms(ctx context.Context, userID string) ([]*models.Room, error)
}

type redisRoomCacheRepository struct {
	redis *redisClient.Client
}

// NewRedisRoomCacheRepository creates a new Redis-based room cache repository
func NewRedisRoomCacheRepository(redisClient *redisClient.Client) RoomCacheRepository {
	return &redisRoomCacheRepository{
		redis: redisClient,
	}
}

const (
	roomKeyPrefix      = "room:"
	userRoomsKeyPrefix = "user_rooms:"
	roomCacheTTL       = 10 * time.Minute // Room data doesn't change frequently
)

// CacheRoom caches a room
func (r *redisRoomCacheRepository) CacheRoom(ctx context.Context, room *models.Room) error {
	dependency := "redis"
	log := mlog.L(ctx)
	roomJSON, err := json.Marshal(room)
	if err != nil {
		return fmt.Errorf("failed to marshal room: %w", err)
	}
	roomKey := fmt.Sprintf("%s%s", roomKeyPrefix, room.ID)
	start := time.Now()
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_CREATE, "redis set room"), map[string]any{
		"key":        roomKey,
		"value":      room,
		"expiration": roomCacheTTL,
	})
	err = r.redis.Set(roomKey, roomJSON, roomCacheTTL)
	end := time.Since(start).Milliseconds()

	result := map[string]any{}
	if err != nil {
		// return fmt.Errorf("failed to cache room: %w", err)
		result["status"] = "error"
		result["error"] = err.Error()
	} else {
		result["status"] = "success"
	}

	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency:   dependency,
		ResponseTime: end,
	}).Debug(logAction.DB_RESPONSE(logAction.DB_CREATE, "redis set room"), result)

	return err
}

// GetCachedRoom retrieves a cached room
func (r *redisRoomCacheRepository) GetCachedRoom(ctx context.Context, roomID string) (*models.Room, error) {
	dependency := "redis"
	log := mlog.L(ctx)
	roomKey := fmt.Sprintf("%s%s", roomKeyPrefix, roomID)

	start := time.Now()
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_READ, "redis get room"), map[string]any{
		"key": roomKey,
	})
	data, err := r.redis.Get(roomKey)
	end := time.Since(start).Milliseconds()

	result := map[string]any{}
	if err != nil {
		result["status"] = "error"
		result["error"] = err.Error()
		log.SetDependencyMetadata(logger.DependencyMetadata{
			Dependency:   dependency,
			ResponseTime: end,
		}).Debug(logAction.DB_RESPONSE(logAction.DB_READ, "redis get room"), result)
		return nil, err // Cache miss
	}

	var room models.Room
	if err := json.Unmarshal([]byte(data), &room); err != nil {
		result["status"] = "error"
		result["error"] = err.Error()
		log.SetDependencyMetadata(logger.DependencyMetadata{
			Dependency:   dependency,
			ResponseTime: end,
		}).Debug(logAction.DB_RESPONSE(logAction.DB_READ, "redis get room"), result)
		return nil, fmt.Errorf("failed to unmarshal room: %w", err)
	}

	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency:   dependency,
		ResponseTime: end,
	}).Debug(logAction.DB_RESPONSE(logAction.DB_READ, "redis get room"), map[string]any{
		"status": "success",
		"room":   room,
	})

	return &room, nil
}

// InvalidateRoom removes a room from cache
func (r *redisRoomCacheRepository) InvalidateRoom(ctx context.Context, roomID string) error {
	dependency := "redis"
	log := mlog.L(ctx)

	roomKey := fmt.Sprintf("%s%s", roomKeyPrefix, roomID)
	start := time.Now()
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_DELETE, "redis del room"), map[string]any{
		"key": roomKey,
	})
	err := r.redis.Del(roomKey)
	end := time.Since(start).Milliseconds()

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
	}).Debug(logAction.DB_RESPONSE(logAction.DB_DELETE, "redis del room"), result)
	return err
}

// CacheUserRooms caches all rooms for a user
func (r *redisRoomCacheRepository) CacheUserRooms(ctx context.Context, userID string, rooms []*models.Room) error {
	dependency := "redis"
	log := mlog.L(ctx)
	start := time.Now()
	roomsJSON, err := json.Marshal(rooms)
	if err != nil {
		return fmt.Errorf("failed to marshal user rooms: %w", err)
	}

	key := fmt.Sprintf("%s%s", userRoomsKeyPrefix, userID)
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_CREATE, "redis set user rooms"), map[string]any{
		"key":        key,
		"value":      rooms,
		"expiration": roomCacheTTL,
	})
	err = r.redis.Set(key, roomsJSON, roomCacheTTL)
	end := time.Since(start).Milliseconds()

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
	}).Debug(logAction.DB_RESPONSE(logAction.DB_CREATE, "redis set user rooms"), result)

	return err
}

// GetCachedUserRooms retrieves cached rooms for a user
func (r *redisRoomCacheRepository) GetCachedUserRooms(ctx context.Context, userID string) ([]*models.Room, error) {
	dependency := "redis"
	log := mlog.L(ctx)
	key := fmt.Sprintf("%s%s", userRoomsKeyPrefix, userID)
	start := time.Now()
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency: dependency,
	}).Debug(logAction.DB_REQUEST(logAction.DB_READ, "redis get user rooms"), map[string]any{
		"key": key,
	})

	data, err := r.redis.Get(key)
	end := time.Since(start).Milliseconds()

	result := map[string]any{}
	if err != nil {
		result["status"] = "error"
		result["error"] = err.Error()
		log.SetDependencyMetadata(logger.DependencyMetadata{
			Dependency:   dependency,
			ResponseTime: end,
		}).Debug(logAction.DB_RESPONSE(logAction.DB_READ, "redis get user rooms"), result)
		return nil, err // Cache miss
	}
	result["status"] = "success"
	log.SetDependencyMetadata(logger.DependencyMetadata{
		Dependency:   dependency,
		ResponseTime: end,
	}).Debug(logAction.DB_RESPONSE(logAction.DB_READ, "redis get user rooms"), result)

	var rooms []*models.Room
	if err := json.Unmarshal([]byte(data), &rooms); err != nil {
		return nil, fmt.Errorf("failed to unmarshal user rooms: %w", err)
	}

	return rooms, nil
}

// NoOpRoomCacheRepository is a no-op implementation when Redis is not available
type noOpRoomCacheRepository struct{}

func NewNoOpRoomCacheRepository() RoomCacheRepository {
	return &noOpRoomCacheRepository{}
}

func (n *noOpRoomCacheRepository) CacheRoom(ctx context.Context, room *models.Room) error {
	return nil
}

func (n *noOpRoomCacheRepository) GetCachedRoom(ctx context.Context, roomID string) (*models.Room, error) {
	return nil, fmt.Errorf("cache not available")
}

func (n *noOpRoomCacheRepository) InvalidateRoom(ctx context.Context, roomID string) error {
	return nil
}

func (n *noOpRoomCacheRepository) CacheUserRooms(ctx context.Context, userID string, rooms []*models.Room) error {
	return nil
}

func (n *noOpRoomCacheRepository) GetCachedUserRooms(ctx context.Context, userID string) ([]*models.Room, error) {
	return nil, fmt.Errorf("cache not available")
}
