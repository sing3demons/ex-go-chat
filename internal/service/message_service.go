package service

import (
	"context"
	"time"

	"realtime-chat-system/internal/models"
	"realtime-chat-system/internal/repository"
	"realtime-chat-system/pkg/errors"
	"realtime-chat-system/pkg/validator"
)

// MessageService defines the interface for message operations
type MessageService interface {
	SendMessage(ctx context.Context, roomID, senderID, content string) (*models.Message, error)
	GetMessages(ctx context.Context, roomID string, limit, offset int) ([]*models.Message, error)
	GetMessageByID(ctx context.Context, messageID string) (*models.Message, error)
	UpdateDeliveryStatus(ctx context.Context, messageID, userID string) error
	UpdateReadStatus(ctx context.Context, messageID, userID string) error
	EditMessage(ctx context.Context, messageID, userID, newContent string) error
	DeleteMessage(ctx context.Context, messageID, userID string) error
}

// messageService implements MessageService
type messageService struct {
	messageRepo repository.MessageRepository
	roomRepo    repository.RoomRepository
	cacheRepo   repository.MessageCacheRepository
}

// NewMessageService creates a new message service
func NewMessageService(messageRepo repository.MessageRepository, roomRepo repository.RoomRepository, cacheRepo repository.MessageCacheRepository) MessageService {
	return &messageService{
		messageRepo: messageRepo,
		roomRepo:    roomRepo,
		cacheRepo:   cacheRepo,
	}
}

// SendMessage sends a new message in a room
func (s *messageService) SendMessage(ctx context.Context, roomID, senderID, content string) (*models.Message, error) {
	// Validate content
	if err := validator.ValidateMessageContent(content); err != nil {
		return nil, err
	}

	// Get room to verify membership and get members list
	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	// Verify sender is a member
	isMember := false
	for _, member := range room.Members {
		if member == senderID {
			isMember = true
			break
		}
	}
	if !isMember {
		return nil, errors.ErrNotRoomMember()
	}

	// Initialize status map for all members except sender
	statusMap := make(map[string]*models.Status)
	for _, member := range room.Members {
		if member != senderID {
			statusMap[member] = &models.Status{
				Delivered:   false,
				DeliveredAt: nil,
				Read:        false,
				ReadAt:      nil,
			}
		}
	}

	// Create message
	now := time.Now()
	message := &models.Message{
		RoomID:    roomID,
		SenderID:  senderID,
		Content:   content,
		CreatedAt: now,
		UpdatedAt: now,
		Deleted:   false,
		Status:    statusMap,
	}

	// Save to database
	if err := s.messageRepo.Create(ctx, message); err != nil {
		return nil, err
	}

	// Cache the message
	if s.cacheRepo != nil {
		if err := s.cacheRepo.CacheMessage(ctx, message); err != nil {
			// Log error but don't fail the request
			// In production, you might want to use a proper logger
		}
	}

	return message, nil
}

// GetMessages retrieves messages from a room with pagination
func (s *messageService) GetMessages(ctx context.Context, roomID string, limit, offset int) ([]*models.Message, error) {
	// Note: Authorization check should be done by the caller (handler/service)
	// to verify user is a member of the room

	// Try to get from cache first (only for recent messages, offset = 0)
	if s.cacheRepo != nil && offset == 0 {
		cachedMessages, err := s.cacheRepo.GetCachedRoomMessages(ctx, roomID, int64(limit))
		if err == nil && len(cachedMessages) > 0 {
			// Return cached messages if available
			return cachedMessages, nil
		}
	}

	// Fallback to database
	messages, err := s.messageRepo.FindByRoom(ctx, roomID, limit, offset)
	if err != nil {
		return nil, err
	}

	// Cache the messages if this is a recent fetch (offset = 0)
	if s.cacheRepo != nil && offset == 0 {
		for _, message := range messages {
			if err := s.cacheRepo.CacheMessage(ctx, message); err != nil {
				// Log error but continue
			}
		}
	}

	return messages, nil
}

// GetMessageByID retrieves a single message by ID
func (s *messageService) GetMessageByID(ctx context.Context, messageID string) (*models.Message, error) {
	return s.messageRepo.FindByID(ctx, messageID)
}

// UpdateDeliveryStatus updates the delivery status of a message for a user
func (s *messageService) UpdateDeliveryStatus(ctx context.Context, messageID, userID string) error {
	// Get message
	message, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return err
	}

	// Check if user has status entry
	status, exists := message.Status[userID]
	if !exists {
		// User is not a recipient of this message (might be the sender)
		return nil
	}

	// Update delivery status if not already delivered
	if !status.Delivered {
		now := time.Now()
		status.Delivered = true
		status.DeliveredAt = &now
		message.Status[userID] = status
		message.UpdatedAt = now

		// Save to database
		if err := s.messageRepo.Update(ctx, message); err != nil {
			return err
		}

		// Update cache
		if s.cacheRepo != nil {
			if err := s.cacheRepo.CacheMessage(ctx, message); err != nil {
				// Log error but don't fail
			}
		}
	}

	return nil
}

// UpdateReadStatus updates the read status of a message for a user
func (s *messageService) UpdateReadStatus(ctx context.Context, messageID, userID string) error {
	// Get message
	message, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return err
	}

	// Check if user has status entry
	status, exists := message.Status[userID]
	if !exists {
		// User is not a recipient of this message (might be the sender)
		return nil
	}

	// Update read status if not already read
	if !status.Read {
		now := time.Now()

		// Mark as delivered if not already
		if !status.Delivered {
			status.Delivered = true
			status.DeliveredAt = &now
		}

		status.Read = true
		status.ReadAt = &now
		message.Status[userID] = status
		message.UpdatedAt = now

		// Save to database
		if err := s.messageRepo.Update(ctx, message); err != nil {
			return err
		}

		// Update cache
		if s.cacheRepo != nil {
			if err := s.cacheRepo.CacheMessage(ctx, message); err != nil {
				// Log error but don't fail
			}
		}
	}

	return nil
}

// EditMessage edits a message (only by the sender)
func (s *messageService) EditMessage(ctx context.Context, messageID, userID, newContent string) error {
	// Validate content
	if err := validator.ValidateMessageContent(newContent); err != nil {
		return err
	}

	// Get message
	message, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return err
	}

	// Verify user is the sender
	if message.SenderID != userID {
		return errors.ErrNotMessageOwner()
	}

	// Check if message is already deleted
	if message.Deleted {
		return errors.ErrInvalidInput("Cannot edit deleted message")
	}

	// Update message
	now := time.Now()
	message.Content = newContent
	message.UpdatedAt = now
	message.EditedAt = &now

	// Save to database
	if err := s.messageRepo.Update(ctx, message); err != nil {
		return err
	}

	// Update cache
	if s.cacheRepo != nil {
		if err := s.cacheRepo.CacheMessage(ctx, message); err != nil {
			// Log error but don't fail
		}
	}

	return nil
}

// DeleteMessage deletes a message (only by the sender)
func (s *messageService) DeleteMessage(ctx context.Context, messageID, userID string) error {
	// Get message
	message, err := s.messageRepo.FindByID(ctx, messageID)
	if err != nil {
		return err
	}

	// Verify user is the sender
	if message.SenderID != userID {
		return errors.ErrNotMessageOwner()
	}

	// Check if message is already deleted
	if message.Deleted {
		return nil // Already deleted, no error
	}

	// Mark as deleted
	if err := s.messageRepo.Delete(ctx, messageID); err != nil {
		return err
	}

	// Invalidate cache
	if s.cacheRepo != nil {
		if err := s.cacheRepo.InvalidateMessage(ctx, messageID, message.RoomID); err != nil {
			// Log error but don't fail
		}
	}

	return nil
}
