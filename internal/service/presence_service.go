package service

import (
	"context"
	"time"

	"realtime-chat-system/internal/repository"
	"realtime-chat-system/pkg/logger"
)

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
	repo repository.PresenceRepository
	log  *logger.Logger
}

// NewPresenceService creates a new presence service
func NewPresenceService(repo repository.PresenceRepository, log *logger.Logger) PresenceService {
	return &presenceService{
		repo: repo,
		log:  log,
	}
}

// SetOnline marks a user as online
func (s *presenceService) SetOnline(ctx context.Context, userID string) {
	if err := s.repo.SetOnline(ctx, userID); err != nil {
		s.log.Errorf("Failed to set user %s online: %v", userID, err)
		return
	}
	s.log.Infof("User %s is now online", userID)
}

// SetOffline marks a user as offline
func (s *presenceService) SetOffline(ctx context.Context, userID string) {
	if err := s.repo.SetOffline(ctx, userID); err != nil {
		s.log.Errorf("Failed to set user %s offline: %v", userID, err)
		return
	}
	s.log.Infof("User %s is now offline", userID)
}

// IsOnline checks if a user is online
func (s *presenceService) IsOnline(ctx context.Context, userID string) bool {
	online, err := s.repo.IsOnline(ctx, userID)
	if err != nil {
		s.log.Errorf("Failed to check if user %s is online: %v", userID, err)
		return false
	}
	return online
}

// GetLastSeen returns the last seen time for a user
func (s *presenceService) GetLastSeen(ctx context.Context, userID string) (time.Time, bool) {
	return s.repo.GetLastSeen(ctx, userID)
}

// GetOnlineUsers returns all online user IDs
func (s *presenceService) GetOnlineUsers(ctx context.Context) []string {
	users, err := s.repo.GetOnlineUsers(ctx)
	if err != nil {
		s.log.Errorf("Failed to get online users: %v", err)
		return []string{}
	}
	return users
}

// UpdateHeartbeat updates the last seen time for a user
func (s *presenceService) UpdateHeartbeat(ctx context.Context, userID string) {
	if err := s.repo.UpdateHeartbeat(ctx, userID); err != nil {
		s.log.Errorf("Failed to update heartbeat for user %s: %v", userID, err)
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
			s.checkStaleConnections(ctx)
		}
	}
}

// checkStaleConnections marks users as offline if no heartbeat received
func (s *presenceService) checkStaleConnections(ctx context.Context) {
	staleThreshold := 60 * time.Second

	staleUsers, err := s.repo.GetStaleUsers(ctx, staleThreshold)
	if err != nil {
		s.log.Errorf("Failed to get stale users: %v", err)
		return
	}

	for _, userID := range staleUsers {
		if err := s.repo.SetOffline(ctx, userID); err != nil {
			s.log.Errorf("Failed to mark user %s offline: %v", userID, err)
		} else {
			s.log.Infof("User %s marked offline due to stale connection", userID)
		}
	}
}
