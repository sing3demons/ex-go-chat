package websocket

import (
	"context"
	"encoding/json"
	"time"

	"realtime-chat-system/internal/service"
	"realtime-chat-system/pkg/logger"
)

// WSMessageHandler handles incoming WebSocket messages
type WSMessageHandler struct {
	hub                 *Hub
	messageService      service.MessageService
	roomService         service.RoomService
	presenceService     service.PresenceService
	notificationService service.NotificationService
	log                 *logger.Logger
}

// NewWSMessageHandler creates a new WebSocket message handler
func NewWSMessageHandler(
	hub *Hub,
	messageService service.MessageService,
	roomService service.RoomService,
	presenceService service.PresenceService,
	notificationService service.NotificationService,
	log *logger.Logger,
) *WSMessageHandler {
	return &WSMessageHandler{
		hub:                 hub,
		messageService:      messageService,
		roomService:         roomService,
		presenceService:     presenceService,
		notificationService: notificationService,
		log:                 log,
	}
}

// HandleMessage handles an incoming WebSocket message
func (h *WSMessageHandler) HandleMessage(conn *Connection, msg *WSMessage) {
	ctx := context.Background()

	switch msg.Type {
	case MessageTypeMessage:
		h.handleChatMessage(ctx, conn, msg)
	case MessageTypeTyping:
		h.handleTypingIndicator(ctx, conn, msg)
	case MessageTypeRead:
		h.handleReadStatus(ctx, conn, msg)
	case MessageTypeDelivered:
		h.handleDeliveredStatus(ctx, conn, msg)
	case MessageTypeEdit:
		h.handleEditMessage(ctx, conn, msg)
	case MessageTypeDelete:
		h.handleDeleteMessage(ctx, conn, msg)
	case MessageTypeHeartbeat:
		h.handleHeartbeat(ctx, conn)
	default:
		h.log.Warnf("Unknown message type: %s", msg.Type)
		h.sendError(conn, "UNKNOWN_TYPE", "Unknown message type")
	}
}

// handleChatMessage handles incoming chat messages
func (h *WSMessageHandler) handleChatMessage(ctx context.Context, conn *Connection, msg *WSMessage) {
	var payload ChatMessagePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.log.Errorf("Failed to unmarshal chat message: %v", err)
		h.sendError(conn, "INVALID_PAYLOAD", "Invalid message payload")
		return
	}

	// Verify room membership
	isMember, err := h.roomService.IsMember(ctx, msg.RoomID, conn.UserID)
	if err != nil {
		h.log.Errorf("Failed to check room membership: %v", err)
		h.sendError(conn, "INTERNAL_ERROR", "Failed to verify room membership")
		return
	}
	if !isMember {
		h.log.Warnf("User %s attempted to send message to room %s without membership", conn.UserID, msg.RoomID)
		h.sendError(conn, "UNAUTHORIZED", "You are not a member of this room")
		return
	}

	// Save message to database
	message, err := h.messageService.SendMessage(ctx, msg.RoomID, conn.UserID, payload.Content)
	if err != nil {
		h.log.Errorf("Failed to save message: %v", err)
		h.sendError(conn, "INTERNAL_ERROR", "Failed to send message")
		return
	}

	// Prepare broadcast payload
	broadcastPayload := ChatMessagePayload{
		MessageID: message.ID,
		Content:   message.Content,
		SenderID:  message.SenderID,
		Timestamp: message.CreatedAt.Format(time.RFC3339),
	}
	payloadBytes, _ := json.Marshal(broadcastPayload)

	// Broadcast to room
	h.hub.BroadcastToRoom <- &RoomBroadcast{
		RoomID: msg.RoomID,
		Message: &WSMessage{
			Type:    MessageTypeMessage,
			RoomID:  msg.RoomID,
			Payload: payloadBytes,
		},
	}

	// Create notifications for offline members
	go h.createNotificationsForOfflineMembers(ctx, msg.RoomID, message.ID, conn.UserID)

	h.log.Infof("Message sent to room %s by user %s", msg.RoomID, conn.UserID)
}

// handleTypingIndicator handles typing indicator events
func (h *WSMessageHandler) handleTypingIndicator(ctx context.Context, conn *Connection, msg *WSMessage) {
	var payload TypingPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.log.Errorf("Failed to unmarshal typing payload: %v", err)
		return
	}

	// Verify room membership
	isMember, err := h.roomService.IsMember(ctx, msg.RoomID, conn.UserID)
	if err != nil || !isMember {
		return
	}

	// Set sender info
	payload.UserID = conn.UserID
	payload.Username = conn.Username
	payloadBytes, _ := json.Marshal(payload)

	// Broadcast to room (exclude sender)
	h.hub.BroadcastToRoom <- &RoomBroadcast{
		RoomID: msg.RoomID,
		Message: &WSMessage{
			Type:    MessageTypeTyping,
			RoomID:  msg.RoomID,
			Payload: payloadBytes,
		},
		Exclude: conn.UserID,
	}
}

// handleReadStatus handles read status updates
func (h *WSMessageHandler) handleReadStatus(ctx context.Context, conn *Connection, msg *WSMessage) {
	var payload StatusPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.log.Errorf("Failed to unmarshal read status: %v", err)
		return
	}

	// Update read status in database
	if err := h.messageService.UpdateReadStatus(ctx, payload.MessageID, conn.UserID); err != nil {
		h.log.Errorf("Failed to update read status: %v", err)
		return
	}

	// Get message to find sender
	message, err := h.messageService.GetMessages(ctx, msg.RoomID, 1, 0)
	if err != nil || len(message) == 0 {
		return
	}

	// Prepare status payload
	payload.UserID = conn.UserID
	payload.Status = "read"
	payload.Timestamp = time.Now().Format(time.RFC3339)
	payloadBytes, _ := json.Marshal(payload)

	// Broadcast to sender
	h.hub.BroadcastToUser <- &UserBroadcast{
		UserID: message[0].SenderID,
		Message: &WSMessage{
			Type:    MessageTypeRead,
			RoomID:  msg.RoomID,
			Payload: payloadBytes,
		},
	}
}

// handleDeliveredStatus handles delivered status updates
func (h *WSMessageHandler) handleDeliveredStatus(ctx context.Context, conn *Connection, msg *WSMessage) {
	var payload StatusPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.log.Errorf("Failed to unmarshal delivered status: %v", err)
		return
	}

	// Update delivered status in database
	if err := h.messageService.UpdateDeliveryStatus(ctx, payload.MessageID, conn.UserID); err != nil {
		h.log.Errorf("Failed to update delivered status: %v", err)
		return
	}

	// Get message to find sender
	message, err := h.messageService.GetMessages(ctx, msg.RoomID, 1, 0)
	if err != nil || len(message) == 0 {
		return
	}

	// Prepare status payload
	payload.UserID = conn.UserID
	payload.Status = "delivered"
	payload.Timestamp = time.Now().Format(time.RFC3339)
	payloadBytes, _ := json.Marshal(payload)

	// Broadcast to sender
	h.hub.BroadcastToUser <- &UserBroadcast{
		UserID: message[0].SenderID,
		Message: &WSMessage{
			Type:    MessageTypeDelivered,
			RoomID:  msg.RoomID,
			Payload: payloadBytes,
		},
	}
}

// handleEditMessage handles message edit requests
func (h *WSMessageHandler) handleEditMessage(ctx context.Context, conn *Connection, msg *WSMessage) {
	var payload EditPayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.log.Errorf("Failed to unmarshal edit payload: %v", err)
		h.sendError(conn, "INVALID_PAYLOAD", "Invalid edit payload")
		return
	}

	// Edit message
	if err := h.messageService.EditMessage(ctx, payload.MessageID, conn.UserID, payload.Content); err != nil {
		h.log.Errorf("Failed to edit message: %v", err)
		h.sendError(conn, "EDIT_FAILED", "Failed to edit message")
		return
	}

	// Prepare broadcast payload
	payload.EditedAt = time.Now().Format(time.RFC3339)
	payloadBytes, _ := json.Marshal(payload)

	// Broadcast to room
	h.hub.BroadcastToRoom <- &RoomBroadcast{
		RoomID: msg.RoomID,
		Message: &WSMessage{
			Type:    MessageTypeEdit,
			RoomID:  msg.RoomID,
			Payload: payloadBytes,
		},
	}

	h.log.Infof("Message %s edited by user %s", payload.MessageID, conn.UserID)
}

// handleDeleteMessage handles message delete requests
func (h *WSMessageHandler) handleDeleteMessage(ctx context.Context, conn *Connection, msg *WSMessage) {
	var payload DeletePayload
	if err := json.Unmarshal(msg.Payload, &payload); err != nil {
		h.log.Errorf("Failed to unmarshal delete payload: %v", err)
		h.sendError(conn, "INVALID_PAYLOAD", "Invalid delete payload")
		return
	}

	// Delete message
	if err := h.messageService.DeleteMessage(ctx, payload.MessageID, conn.UserID); err != nil {
		h.log.Errorf("Failed to delete message: %v", err)
		h.sendError(conn, "DELETE_FAILED", "Failed to delete message")
		return
	}

	// Prepare broadcast payload
	payloadBytes, _ := json.Marshal(payload)

	// Broadcast to room
	h.hub.BroadcastToRoom <- &RoomBroadcast{
		RoomID: msg.RoomID,
		Message: &WSMessage{
			Type:    MessageTypeDelete,
			RoomID:  msg.RoomID,
			Payload: payloadBytes,
		},
	}

	h.log.Infof("Message %s deleted by user %s", payload.MessageID, conn.UserID)
}

// handleHeartbeat handles heartbeat messages
func (h *WSMessageHandler) handleHeartbeat(ctx context.Context, conn *Connection) {
	h.presenceService.UpdateHeartbeat(ctx, conn.UserID)
}

// sendError sends an error message to the connection
func (h *WSMessageHandler) sendError(conn *Connection, code, message string) {
	payload := ErrorPayload{
		Code:    code,
		Message: message,
	}
	payloadBytes, _ := json.Marshal(payload)

	conn.SendMessage(&WSMessage{
		Type:    MessageTypeError,
		Payload: payloadBytes,
	})
}

// createNotificationsForOfflineMembers creates notifications for offline room members
func (h *WSMessageHandler) createNotificationsForOfflineMembers(ctx context.Context, roomID, messageID, senderID string) {
	// Get room to find members
	room, err := h.roomService.GetRoom(ctx, roomID)
	if err != nil {
		h.log.Errorf("Failed to get room for notifications: %v", err)
		return
	}

	// Create notification for each offline member (except sender)
	for _, memberID := range room.Members {
		if memberID == senderID {
			continue
		}

		// Check if member is online
		if h.presenceService.IsOnline(ctx, memberID) {
			continue
		}

		// Create notification for offline member
		if err := h.notificationService.CreateNotification(ctx, memberID, roomID, messageID, "message"); err != nil {
			h.log.Errorf("Failed to create notification for user %s: %v", memberID, err)
		}
	}
}
