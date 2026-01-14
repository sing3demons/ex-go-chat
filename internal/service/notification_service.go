package service

import (
	"context"

	"realtime-chat-system/internal/models"
	"realtime-chat-system/internal/repository"
)

// NotificationService handles notification business logic
type NotificationService interface {
	CreateNotification(ctx context.Context, userID, roomID, messageID, notificationType string) error
	GetPendingNotifications(ctx context.Context, userID string) ([]*models.Notification, error)
	GetNotifications(ctx context.Context, userID string, limit, offset int) ([]*models.Notification, error)
	MarkAsRead(ctx context.Context, notificationID string) error
	MarkAllAsRead(ctx context.Context, userID string) error
	GetUnreadCount(ctx context.Context, userID string) (int64, error)
}

type notificationService struct {
	notificationRepo repository.NotificationRepository
}

// NewNotificationService creates a new notification service
func NewNotificationService(notificationRepo repository.NotificationRepository) NotificationService {
	return &notificationService{
		notificationRepo: notificationRepo,
	}
}

// CreateNotification creates a new notification for a user
func (s *notificationService) CreateNotification(ctx context.Context, userID, roomID, messageID, notificationType string) error {
	notification := &models.Notification{
		UserID:    userID,
		RoomID:    roomID,
		MessageID: messageID,
		Type:      notificationType,
		Read:      false,
	}

	return s.notificationRepo.Create(ctx, notification)
}

// GetPendingNotifications retrieves unread notifications for a user
func (s *notificationService) GetPendingNotifications(ctx context.Context, userID string) ([]*models.Notification, error) {
	return s.notificationRepo.FindPendingByUserID(ctx, userID)
}

// GetNotifications retrieves notifications for a user with pagination
func (s *notificationService) GetNotifications(ctx context.Context, userID string, limit, offset int) ([]*models.Notification, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	return s.notificationRepo.FindByUserID(ctx, userID, limit, offset)
}

// MarkAsRead marks a notification as read
func (s *notificationService) MarkAsRead(ctx context.Context, notificationID string) error {
	return s.notificationRepo.MarkAsRead(ctx, notificationID)
}

// MarkAllAsRead marks all notifications for a user as read
func (s *notificationService) MarkAllAsRead(ctx context.Context, userID string) error {
	return s.notificationRepo.MarkAllAsRead(ctx, userID)
}

// GetUnreadCount returns the count of unread notifications for a user
func (s *notificationService) GetUnreadCount(ctx context.Context, userID string) (int64, error) {
	return s.notificationRepo.CountUnread(ctx, userID)
}
