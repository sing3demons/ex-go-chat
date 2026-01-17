package websocket

import "encoding/json"

// MessageType represents the type of WebSocket message
type MessageType string

const (
	MessageTypeMessage   MessageType = "message"
	MessageTypeTyping    MessageType = "typing"
	MessageTypeRead      MessageType = "read"
	MessageTypeDelivered MessageType = "delivered"
	MessageTypePresence  MessageType = "presence"
	MessageTypeEdit      MessageType = "edit"
	MessageTypeDelete    MessageType = "delete"
	MessageTypeHeartbeat MessageType = "heartbeat"
	MessageTypeJoinRoom  MessageType = "join_room"
	MessageTypeRoomCreated MessageType = "room_created"
	MessageTypeError     MessageType = "error"
)

// WSMessage represents a WebSocket message
type WSMessage struct {
	Type    MessageType     `json:"type"`
	RoomID  string          `json:"roomId,omitempty"`
	Payload json.RawMessage `json:"payload"`
}

// ChatMessagePayload represents a chat message payload
type ChatMessagePayload struct {
	MessageID string `json:"messageId"`
	Content   string `json:"content"`
	SenderID  string `json:"senderId"`
	Timestamp string `json:"timestamp"`
	TempID    string `json:"tempId,omitempty"` // For optimistic updates
}

// TypingPayload represents a typing indicator payload
type TypingPayload struct {
	UserID   string `json:"userId"`
	Username string `json:"username"`
	IsTyping bool   `json:"isTyping"`
}

// StatusPayload represents a message status update payload
type StatusPayload struct {
	MessageID string `json:"messageId"`
	UserID    string `json:"userId"`
	Status    string `json:"status"` // "delivered" or "read"
	Timestamp string `json:"timestamp"`
}

// PresencePayload represents a presence update payload
type PresencePayload struct {
	UserID   string `json:"userId"`
	Online   bool   `json:"online"`
	LastSeen string `json:"lastSeen,omitempty"`
}

// EditPayload represents a message edit payload
type EditPayload struct {
	MessageID string `json:"messageId"`
	Content   string `json:"content"`
	EditedAt  string `json:"editedAt"`
}

// DeletePayload represents a message delete payload
type DeletePayload struct {
	MessageID string `json:"messageId"`
}

// ErrorPayload represents an error payload
type ErrorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// RoomCreatedPayload represents a room creation notification payload
type RoomCreatedPayload struct {
	RoomID   string   `json:"roomId"`
	RoomType string   `json:"roomType"`
	Name     string   `json:"name,omitempty"`
	Members  []string `json:"members"`
}
