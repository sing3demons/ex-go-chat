package service

import (
	"encoding/json"
	"fmt"
	"time"

	"realtime-chat-system/internal/models"
	redisClient "realtime-chat-system/pkg/redis"

	"github.com/redis/go-redis/v9"
)

// CacheService handles caching operations
type CacheService struct {
	redis *redisClient.Client
}

// NewCacheService creates a new cache service
func NewCacheService(redisClient *redisClient.Client) *CacheService {
	return &CacheService{
		redis: redisClient,
	}
}

// Message caching
const (
	messageKeyPrefix     = "message:"
	roomMessagesPrefix   = "room_messages:"
	userPresencePrefix   = "presence:"
	typingUsersPrefix    = "typing:"
	onlineUsersKey       = "online_users"
	messageCacheTTL      = 1 * time.Hour
	presenceTTL          = 5 * time.Minute
	typingTTL            = 10 * time.Second
)

// CacheMessage caches a message
func (s *CacheService) CacheMessage(message *models.Message) error {
	messageJSON, err := json.Marshal(message)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	// Cache individual message
	messageKey := fmt.Sprintf("%s%s", messageKeyPrefix, message.ID)
	if err := s.redis.Set(messageKey, messageJSON, messageCacheTTL); err != nil {
		return fmt.Errorf("failed to cache message: %w", err)
	}

	// Add to room messages sorted set (sorted by timestamp)
	roomKey := fmt.Sprintf("%s%s", roomMessagesPrefix, message.RoomID)
	score := float64(message.CreatedAt.Unix())
	if err := s.redis.ZAdd(roomKey, redis.Z{Score: score, Member: message.ID}); err != nil {
		return fmt.Errorf("failed to add message to room cache: %w", err)
	}

	// Set TTL for room messages
	if err := s.redis.Expire(roomKey, messageCacheTTL); err != nil {
		return fmt.Errorf("failed to set TTL for room messages: %w", err)
	}

	return nil
}

// GetCachedMessage retrieves a cached message
func (s *CacheService) GetCachedMessage(messageID string) (*models.Message, error) {
	messageKey := fmt.Sprintf("%s%s", messageKeyPrefix, messageID)
	messageJSON, err := s.redis.Get(messageKey)
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
func (s *CacheService) GetCachedRoomMessages(roomID string, limit int64) ([]*models.Message, error) {
	roomKey := fmt.Sprintf("%s%s", roomMessagesPrefix, roomID)
	
	// Get message IDs in reverse chronological order (newest first)
	messageIDs, err := s.redis.ZRevRange(roomKey, 0, limit-1)
	if err != nil {
		return nil, err
	}

	messages := make([]*models.Message, 0, len(messageIDs))
	for _, messageID := range messageIDs {
		message, err := s.GetCachedMessage(messageID)
		if err != nil {
			// If message not found in cache, skip it
			continue
		}
		messages = append(messages, message)
	}

	return messages, nil
}

// InvalidateMessage removes a message from cache
func (s *CacheService) InvalidateMessage(messageID, roomID string) error {
	// Remove individual message
	messageKey := fmt.Sprintf("%s%s", messageKeyPrefix, messageID)
	if err := s.redis.Del(messageKey); err != nil {
		return fmt.Errorf("failed to delete message from cache: %w", err)
	}

	// Remove from room messages
	roomKey := fmt.Sprintf("%s%s", roomMessagesPrefix, roomID)
	if err := s.redis.ZRem(roomKey, messageID); err != nil {
		return fmt.Errorf("failed to remove message from room cache: %w", err)
	}

	return nil
}

// User presence caching
func (s *CacheService) SetUserOnline(userID string) error {
	// Add to online users set
	if err := s.redis.SAdd(onlineUsersKey, userID); err != nil {
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

	if err := s.redis.Set(presenceKey, presenceJSON, presenceTTL); err != nil {
		return fmt.Errorf("failed to set user presence: %w", err)
	}

	return nil
}

// SetUserOffline sets user as offline
func (s *CacheService) SetUserOffline(userID string) error {
	// Remove from online users set
	if err := s.redis.SRem(onlineUsersKey, userID); err != nil {
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

	if err := s.redis.Set(presenceKey, presenceJSON, presenceTTL); err != nil {
		return fmt.Errorf("failed to set user presence: %w", err)
	}

	return nil
}

// GetOnlineUsers gets list of online users
func (s *CacheService) GetOnlineUsers() ([]string, error) {
	return s.redis.SMembers(onlineUsersKey)
}

// IsUserOnline checks if user is online
func (s *CacheService) IsUserOnline(userID string) (bool, error) {
	return s.redis.SIsMember(onlineUsersKey, userID)
}

// Typing indicators
func (s *CacheService) SetUserTyping(roomID, userID string, isTyping bool) error {
	typingKey := fmt.Sprintf("%s%s", typingUsersPrefix, roomID)
	
	if isTyping {
		// Add user to typing set with TTL
		if err := s.redis.SAdd(typingKey, userID); err != nil {
			return fmt.Errorf("failed to add user to typing set: %w", err)
		}
		if err := s.redis.Expire(typingKey, typingTTL); err != nil {
			return fmt.Errorf("failed to set TTL for typing indicator: %w", err)
		}
	} else {
		// Remove user from typing set
		if err := s.redis.SRem(typingKey, userID); err != nil {
			return fmt.Errorf("failed to remove user from typing set: %w", err)
		}
	}

	return nil
}

// GetTypingUsers gets users currently typing in a room
func (s *CacheService) GetTypingUsers(roomID string) ([]string, error) {
	typingKey := fmt.Sprintf("%s%s", typingUsersPrefix, roomID)
	return s.redis.SMembers(typingKey)
}

// Session management
func (s *CacheService) SetUserSession(userID, sessionID string) error {
	sessionKey := fmt.Sprintf("session:%s", userID)
	return s.redis.Set(sessionKey, sessionID, 24*time.Hour)
}

// GetUserSession gets user session
func (s *CacheService) GetUserSession(userID string) (string, error) {
	sessionKey := fmt.Sprintf("session:%s", userID)
	return s.redis.Get(sessionKey)
}

// InvalidateUserSession removes user session
func (s *CacheService) InvalidateUserSession(userID string) error {
	sessionKey := fmt.Sprintf("session:%s", userID)
	return s.redis.Del(sessionKey)
}

// Rate limiting
func (s *CacheService) CheckRateLimit(userID string, action string, limit int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("rate_limit:%s:%s", userID, action)
	
	// Get current count
	countStr, err := s.redis.Get(key)
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
	if err := s.redis.Set(key, fmt.Sprintf("%d", count), window); err != nil {
		return false, fmt.Errorf("failed to set rate limit count: %w", err)
	}
	
	return true, nil
}