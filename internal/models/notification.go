package models

import "time"

// Notification represents a notification for a user
type Notification struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	UserID    string    `bson:"userId" json:"userId"`
	RoomID    string    `bson:"roomId" json:"roomId"`
	MessageID string    `bson:"messageId" json:"messageId"`
	Type      string    `bson:"type" json:"type"` // "message", "mention", "group_invite"
	Read      bool      `bson:"read" json:"read"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
}

const (
	NotificationTypeMessage     = "message"
	NotificationTypeMention     = "mention"
	NotificationTypeGroupInvite = "group_invite"
)
