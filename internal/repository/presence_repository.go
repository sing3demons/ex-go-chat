package repository

import (
	"context"
	"time"
)

// PresenceRepository defines the interface for presence data access
type PresenceRepository interface {
	SetOnline(ctx context.Context, userID string) error
	SetOffline(ctx context.Context, userID string) error
	IsOnline(ctx context.Context, userID string) (bool, error)
	GetOnlineUsers(ctx context.Context) ([]string, error)
	GetLastSeen(ctx context.Context, userID string) (time.Time, bool)
	UpdateHeartbeat(ctx context.Context, userID string) error
	GetStaleUsers(ctx context.Context, threshold time.Duration) ([]string, error)
}