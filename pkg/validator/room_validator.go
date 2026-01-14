package validator

import (
	"realtime-chat-system/internal/models"
	"realtime-chat-system/pkg/errors"
)

// ValidateRoomType validates room type
func ValidateRoomType(roomType string) error {
	if roomType != models.RoomTypeDirect && roomType != models.RoomTypeGroup {
		return errors.ErrInvalidInput("Room type must be 'direct' or 'group'")
	}
	return nil
}

// ValidateDirectRoomMembers validates members for direct room
// Direct rooms must have exactly 2 members
func ValidateDirectRoomMembers(members []string) error {
	if len(members) != 2 {
		return errors.ErrInvalidInput("Direct room must have exactly 2 members")
	}
	
	// Check for duplicates
	if members[0] == members[1] {
		return errors.ErrInvalidInput("Cannot create direct room with same user")
	}
	
	return nil
}

// ValidateGroupRoomMembers validates members for group room
// Group rooms must have at least 2 members
func ValidateGroupRoomMembers(members []string) error {
	if len(members) < 2 {
		return errors.ErrInvalidInput("Group room must have at least 2 members")
	}
	
	// Check for duplicates
	seen := make(map[string]bool)
	for _, member := range members {
		if seen[member] {
			return errors.ErrInvalidInput("Duplicate members not allowed")
		}
		seen[member] = true
	}
	
	return nil
}

// ValidateGroupRoomName validates group room name
func ValidateGroupRoomName(name string) error {
	return ValidateRoomName(name)
}

// ValidateMemberIDs validates that member IDs are not empty
func ValidateMemberIDs(members []string) error {
	if len(members) == 0 {
		return errors.ErrInvalidInput("At least one member is required")
	}
	
	for _, member := range members {
		if member == "" {
			return errors.ErrInvalidInput("Member ID cannot be empty")
		}
	}
	
	return nil
}
