package models

import "time"

// Room represents a chat room (direct or group)
type Room struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	Type      string    `bson:"type" json:"type"` // "direct" or "group"
	Name      string    `bson:"name,omitempty" json:"name,omitempty"`
	Members   []string  `bson:"members" json:"members"` // user IDs
	CreatedBy string    `bson:"createdBy" json:"createdBy"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}

const (
	RoomTypeDirect = "direct"
	RoomTypeGroup  = "group"
)
