package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"realtime-chat-system/internal/models"
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
	roomJSON, err := json.Marshal(room)
	if err != nil {
		return fmt.Errorf("failed to marshal room: %w", err)
	}

	roomKey := fmt.Sprintf("%s%s", roomKeyPrefix, room.ID)
	if err := r.redis.Set(roomKey, roomJSON, roomCacheTTL); err != nil {
		return fmt.Errorf("failed to cache room: %w", err)
	}

	return nil
}

// GetCachedRoom retrieves a cached room
func (r *redisRoomCacheRepository) GetCachedRoom(ctx context.Context, roomID string) (*models.Room, error) {
	roomKey := fmt.Sprintf("%s%s", roomKeyPrefix, roomID)

	data, err := r.redis.Get(roomKey)
	if err != nil {
		return nil, err // Cache miss
	}

	var room models.Room
	if err := json.Unmarshal([]byte(data), &room); err != nil {
		return nil, fmt.Errorf("failed to unmarshal room: %w", err)
	}

	return &room, nil
}

// InvalidateRoom removes a room from cache
func (r *redisRoomCacheRepository) InvalidateRoom(ctx context.Context, roomID string) error {
	roomKey := fmt.Sprintf("%s%s", roomKeyPrefix, roomID)
	return r.redis.Del(roomKey)
}

// CacheUserRooms caches all rooms for a user
func (r *redisRoomCacheRepository) CacheUserRooms(ctx context.Context, userID string, rooms []*models.Room) error {
	roomsJSON, err := json.Marshal(rooms)
	if err != nil {
		return fmt.Errorf("failed to marshal user rooms: %w", err)
	}

	key := fmt.Sprintf("%s%s", userRoomsKeyPrefix, userID)
	if err := r.redis.Set(key, roomsJSON, roomCacheTTL); err != nil {
		return fmt.Errorf("failed to cache user rooms: %w", err)
	}

	return nil
}

// GetCachedUserRooms retrieves cached rooms for a user
func (r *redisRoomCacheRepository) GetCachedUserRooms(ctx context.Context, userID string) ([]*models.Room, error) {
	key := fmt.Sprintf("%s%s", userRoomsKeyPrefix, userID)

	data, err := r.redis.Get(key)
	if err != nil {
		return nil, err // Cache miss
	}

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
