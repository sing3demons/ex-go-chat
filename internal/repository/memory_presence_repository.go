package repository

import (
	"context"
	"sync"
	"time"
)

// PresenceStatus represents a user's presence status
type PresenceStatus struct {
	UserID   string
	Online   bool
	LastSeen time.Time
}

// memoryPresenceRepository implements PresenceRepository using in-memory storage
type memoryPresenceRepository struct {
	store map[string]*PresenceStatus
	mu    sync.RWMutex
}

// NewMemoryPresenceRepository creates a new memory-based presence repository
func NewMemoryPresenceRepository() PresenceRepository {
	return &memoryPresenceRepository{
		store: make(map[string]*PresenceStatus),
	}
}

// SetOnline marks a user as online
func (r *memoryPresenceRepository) SetOnline(ctx context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.store[userID] = &PresenceStatus{
		UserID:   userID,
		Online:   true,
		LastSeen: time.Now(),
	}
	return nil
}

// SetOffline marks a user as offline
func (r *memoryPresenceRepository) SetOffline(ctx context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if status, exists := r.store[userID]; exists {
		status.Online = false
		status.LastSeen = time.Now()
	}
	return nil
}

// IsOnline checks if a user is online
func (r *memoryPresenceRepository) IsOnline(ctx context.Context, userID string) (bool, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	status, exists := r.store[userID]
	return exists && status.Online, nil
}

// GetOnlineUsers returns all online user IDs
func (r *memoryPresenceRepository) GetOnlineUsers(ctx context.Context) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	users := make([]string, 0)
	for userID, status := range r.store {
		if status.Online {
			users = append(users, userID)
		}
	}
	return users, nil
}

// GetLastSeen returns the last seen time for a user
func (r *memoryPresenceRepository) GetLastSeen(ctx context.Context, userID string) (time.Time, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	status, exists := r.store[userID]
	if !exists {
		return time.Time{}, false
	}
	return status.LastSeen, true
}

// UpdateHeartbeat updates the last seen time for a user
func (r *memoryPresenceRepository) UpdateHeartbeat(ctx context.Context, userID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if status, exists := r.store[userID]; exists {
		status.LastSeen = time.Now()
	}
	return nil
}

// GetStaleUsers returns users who haven't been seen within the threshold
func (r *memoryPresenceRepository) GetStaleUsers(ctx context.Context, threshold time.Duration) ([]string, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now()
	staleUsers := make([]string, 0)

	for userID, status := range r.store {
		if status.Online && now.Sub(status.LastSeen) > threshold {
			staleUsers = append(staleUsers, userID)
		}
	}

	return staleUsers, nil
}