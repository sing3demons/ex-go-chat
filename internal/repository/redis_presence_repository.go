package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	redisClient "realtime-chat-system/pkg/redis"
)

// redisPresenceRepository implements PresenceRepository using Redis
type redisPresenceRepository struct {
	redis *redisClient.Client
}

// NewRedisPresenceRepository creates a new Redis-based presence repository
func NewRedisPresenceRepository(redisClient *redisClient.Client) PresenceRepository {
	return &redisPresenceRepository{
		redis: redisClient,
	}
}

const (
	onlineUsersKey    = "online_users"
	userPresencePrefix = "presence:"
	presenceTTL       = 5 * time.Minute
)

// SetOnline marks a user as online
func (r *redisPresenceRepository) SetOnline(ctx context.Context, userID string) error {
	// Add to online users set
	if err := r.redis.SAdd(onlineUsersKey, userID); err != nil {
		return fmt.Errorf("failed to add user to online set: %w", err)
	}

	// Set user presence with TTL
	presenceKey := fmt.Sprintf("%s%s", userPresencePrefix, userID)
	presenceData := map[string]interface{}{
		"online":    true,
		"lastSeen":  time.Now().Unix(),
		"updatedAt": time.Now().Unix(),
	}

	presenceJSON, err := json.Marshal(presenceData)
	if err != nil {
		return fmt.Errorf("failed to marshal presence data: %w", err)
	}

	if err := r.redis.Set(presenceKey, string(presenceJSON), presenceTTL); err != nil {
		return fmt.Errorf("failed to set user presence: %w", err)
	}

	return nil
}

// SetOffline marks a user as offline
func (r *redisPresenceRepository) SetOffline(ctx context.Context, userID string) error {
	// Remove from online users set
	if err := r.redis.SRem(onlineUsersKey, userID); err != nil {
		return fmt.Errorf("failed to remove user from online set: %w", err)
	}

	// Update presence to offline
	presenceKey := fmt.Sprintf("%s%s", userPresencePrefix, userID)
	presenceData := map[string]interface{}{
		"online":    false,
		"lastSeen":  time.Now().Unix(),
		"updatedAt": time.Now().Unix(),
	}

	presenceJSON, err := json.Marshal(presenceData)
	if err != nil {
		return fmt.Errorf("failed to marshal presence data: %w", err)
	}

	if err := r.redis.Set(presenceKey, string(presenceJSON), presenceTTL); err != nil {
		return fmt.Errorf("failed to set user presence: %w", err)
	}

	return nil
}

// IsOnline checks if a user is online
func (r *redisPresenceRepository) IsOnline(ctx context.Context, userID string) (bool, error) {
	return r.redis.SIsMember(onlineUsersKey, userID)
}

// GetOnlineUsers returns all online user IDs
func (r *redisPresenceRepository) GetOnlineUsers(ctx context.Context) ([]string, error) {
	return r.redis.SMembers(onlineUsersKey)
}

// GetLastSeen returns the last seen time for a user
func (r *redisPresenceRepository) GetLastSeen(ctx context.Context, userID string) (time.Time, bool) {
	presenceKey := fmt.Sprintf("%s%s", userPresencePrefix, userID)
	presenceData, err := r.redis.HGetAll(presenceKey)
	if err != nil {
		return time.Time{}, false
	}

	if lastSeenStr, exists := presenceData["lastSeen"]; exists {
		if lastSeenUnix, err := strconv.ParseInt(lastSeenStr, 10, 64); err == nil {
			return time.Unix(lastSeenUnix, 0), true
		}
	}

	return time.Time{}, false
}

// UpdateHeartbeat updates the last seen time for a user
func (r *redisPresenceRepository) UpdateHeartbeat(ctx context.Context, userID string) error {
	// Refresh user's online status (extends TTL)
	return r.SetOnline(ctx, userID)
}

// GetStaleUsers returns users who haven't been seen within the threshold
func (r *redisPresenceRepository) GetStaleUsers(ctx context.Context, threshold time.Duration) ([]string, error) {
	// Get all online users
	onlineUsers, err := r.GetOnlineUsers(ctx)
	if err != nil {
		return nil, err
	}

	staleUsers := make([]string, 0)
	now := time.Now()

	for _, userID := range onlineUsers {
		lastSeen, exists := r.GetLastSeen(ctx, userID)
		if exists && now.Sub(lastSeen) > threshold {
			staleUsers = append(staleUsers, userID)
		}
	}

	return staleUsers, nil
}