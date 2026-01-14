package models

import "time"

// Message represents a chat message
type Message struct {
	ID        string             `bson:"_id,omitempty" json:"id"`
	RoomID    string             `bson:"roomId" json:"roomId"`
	SenderID  string             `bson:"senderId" json:"senderId"`
	Content   string             `bson:"content" json:"content"`
	CreatedAt time.Time          `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time          `bson:"updatedAt" json:"updatedAt"`
	EditedAt  *time.Time         `bson:"editedAt,omitempty" json:"editedAt,omitempty"`
	Deleted   bool               `bson:"deleted" json:"deleted"`
	Status    map[string]*Status `bson:"status" json:"status"` // userID -> Status
}

// Status represents delivery and read status for a message
type Status struct {
	Delivered   bool       `bson:"delivered" json:"delivered"`
	DeliveredAt *time.Time `bson:"deliveredAt,omitempty" json:"deliveredAt,omitempty"`
	Read        bool       `bson:"read" json:"read"`
	ReadAt      *time.Time `bson:"readAt,omitempty" json:"readAt,omitempty"`
}
