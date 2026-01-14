package service

import (
	"context"
	"sync"
	"time"

	"realtime-chat-system/pkg/logger"
)

// PresenceStatus represents a user's presence status
type PresenceStatus struct {
	UserID   string
	Online   bool
	LastSeen time.Time
}

// PresenceService manages user presence tracking
type PresenceService interface {
	SetOnline(ctx context.Context, userID string)
	SetOffline(ctx context.Context, userID string)
	IsOnline(ctx context.Context, userID string) bool
	GetLastSeen(ctx context.Context, userID string) (time.Time, bool)
	GetOnlineUsers(ctx context.Context) []string
	UpdateHeartbeat(ctx context.Context, userID string)
	StartHeartbeatMonitor(ctx context.Context)
}

type presenceService struct {
	store map[string]*PresenceStatus
	mu    sync.RWMutex
	log   *logger.Logger
}

// NewPresenceService creates a new presence service
func NewPresenceService(log *logger.Logger) PresenceService {
	return &presenceService{
		store: make(map[string]*PresenceStatus),
		log:   log,
	}
}

// SetOnline marks a user as online
func (s *presenceService) SetOnline(ctx context.Context, userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.store[userID] = &PresenceStatus{
		UserID:   userID,
		Online:   true,
		LastSeen: time.Now(),
	}
	s.log.Infof("User %s is now online", userID)
}

// SetOffline marks a user as offline
func (s *presenceService) SetOffline(ctx context.Context, userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if status, exists := s.store[userID]; exists {
		status.Online = false
		status.LastSeen = time.Now()
		s.log.Infof("User %s is now offline", userID)
	}
}

// IsOnline checks if a user is online
func (s *presenceService) IsOnline(ctx context.Context, userID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status, exists := s.store[userID]
	return exists && status.Online
}

// GetLastSeen returns the last seen time for a user
func (s *presenceService) GetLastSeen(ctx context.Context, userID string) (time.Time, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	status, exists := s.store[userID]
	if !exists {
		return time.Time{}, false
	}
	return status.LastSeen, true
}

// GetOnlineUsers returns all online user IDs
func (s *presenceService) GetOnlineUsers(ctx context.Context) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	users := make([]string, 0)
	for userID, status := range s.store {
		if status.Online {
			users = append(users, userID)
		}
	}
	return users
}

// UpdateHeartbeat updates the last seen time for a user
func (s *presenceService) UpdateHeartbeat(ctx context.Context, userID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if status, exists := s.store[userID]; exists {
		status.LastSeen = time.Now()
	}
}

// StartHeartbeatMonitor starts monitoring for stale connections
func (s *presenceService) StartHeartbeatMonitor(ctx context.Context) {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	s.log.Info("Heartbeat monitor started")

	for {
		select {
		case <-ctx.Done():
			s.log.Info("Heartbeat monitor stopped")
			return
		case <-ticker.C:
			s.checkStaleConnections()
		}
	}
}

// checkStaleConnections marks users as offline if no heartbeat received
func (s *presenceService) checkStaleConnections() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	staleThreshold := 60 * time.Second

	for userID, status := range s.store {
		if status.Online && now.Sub(status.LastSeen) > staleThreshold {
			status.Online = false
			s.log.Infof("User %s marked offline due to stale connection", userID)
		}
	}
}
