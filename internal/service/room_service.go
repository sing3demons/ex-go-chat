package service

import (
	"context"
	"sort"
	"time"

	"realtime-chat-system/internal/models"
	"realtime-chat-system/internal/repository"
	"realtime-chat-system/pkg/errors"
	"realtime-chat-system/pkg/validator"
)

// RoomService defines the interface for room operations
type RoomService interface {
	CreateDirectRoom(ctx context.Context, user1ID, user2ID string) (*models.Room, error)
	CreateGroupRoom(ctx context.Context, name string, creatorID string, memberIDs []string) (*models.Room, error)
	GetRoom(ctx context.Context, roomID string) (*models.Room, error)
	GetUserRooms(ctx context.Context, userID string) ([]*models.Room, error)
	AddMembers(ctx context.Context, roomID string, memberIDs []string) error
	RemoveMembers(ctx context.Context, roomID string, memberIDs []string) error
	IsMember(ctx context.Context, roomID, userID string) (bool, error)
}

// roomService implements RoomService
type roomService struct {
	roomRepo repository.RoomRepository
}

// NewRoomService creates a new room service
func NewRoomService(roomRepo repository.RoomRepository) RoomService {
	return &roomService{
		roomRepo: roomRepo,
	}
}

// CreateDirectRoom creates a direct (1-on-1) room
func (s *roomService) CreateDirectRoom(ctx context.Context, user1ID, user2ID string) (*models.Room, error) {
	// Validate member IDs
	members := []string{user1ID, user2ID}
	if err := validator.ValidateDirectRoomMembers(members); err != nil {
		return nil, err
	}

	// Check if direct room already exists
	existingRoom, err := s.roomRepo.FindByMembers(ctx, members)
	if err != nil {
		return nil, err
	}
	if existingRoom != nil {
		return existingRoom, nil // Return existing room
	}

	// Sort members for consistent ordering
	sort.Strings(members)

	// Create new direct room
	now := time.Now()
	room := &models.Room{
		Type:      models.RoomTypeDirect,
		Members:   members,
		CreatedBy: user1ID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.roomRepo.Create(ctx, room); err != nil {
		return nil, err
	}

	return room, nil
}

// CreateGroupRoom creates a group room
func (s *roomService) CreateGroupRoom(ctx context.Context, name string, creatorID string, memberIDs []string) (*models.Room, error) {
	// Validate group name
	if err := validator.ValidateGroupRoomName(name); err != nil {
		return nil, err
	}

	// Ensure creator is in members list
	members := append([]string{creatorID}, memberIDs...)
	
	// Remove duplicates and validate
	members = removeDuplicates(members)
	if err := validator.ValidateGroupRoomMembers(members); err != nil {
		return nil, err
	}

	// Create new group room
	now := time.Now()
	room := &models.Room{
		Type:      models.RoomTypeGroup,
		Name:      name,
		Members:   members,
		CreatedBy: creatorID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := s.roomRepo.Create(ctx, room); err != nil {
		return nil, err
	}

	return room, nil
}

// GetRoom retrieves a room by ID
func (s *roomService) GetRoom(ctx context.Context, roomID string) (*models.Room, error) {
	return s.roomRepo.FindByID(ctx, roomID)
}

// GetUserRooms retrieves all rooms for a user
func (s *roomService) GetUserRooms(ctx context.Context, userID string) ([]*models.Room, error) {
	return s.roomRepo.FindUserRooms(ctx, userID)
}

// AddMembers adds members to a group room
func (s *roomService) AddMembers(ctx context.Context, roomID string, memberIDs []string) error {
	// Validate member IDs
	if err := validator.ValidateMemberIDs(memberIDs); err != nil {
		return err
	}

	// Get room
	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return err
	}

	// Only group rooms can have members added
	if room.Type != models.RoomTypeGroup {
		return errors.ErrInvalidInput("Cannot add members to direct room")
	}

	// Add new members (avoid duplicates)
	existingMembers := make(map[string]bool)
	for _, member := range room.Members {
		existingMembers[member] = true
	}

	newMembers := room.Members
	for _, memberID := range memberIDs {
		if !existingMembers[memberID] {
			newMembers = append(newMembers, memberID)
		}
	}

	// Update members
	return s.roomRepo.UpdateMembers(ctx, roomID, newMembers)
}

// RemoveMembers removes members from a group room
func (s *roomService) RemoveMembers(ctx context.Context, roomID string, memberIDs []string) error {
	// Validate member IDs
	if err := validator.ValidateMemberIDs(memberIDs); err != nil {
		return err
	}

	// Get room
	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return err
	}

	// Only group rooms can have members removed
	if room.Type != models.RoomTypeGroup {
		return errors.ErrInvalidInput("Cannot remove members from direct room")
	}

	// Create set of members to remove
	toRemove := make(map[string]bool)
	for _, memberID := range memberIDs {
		toRemove[memberID] = true
	}

	// Filter out removed members
	newMembers := []string{}
	for _, member := range room.Members {
		if !toRemove[member] {
			newMembers = append(newMembers, member)
		}
	}

	// Validate remaining members
	if err := validator.ValidateGroupRoomMembers(newMembers); err != nil {
		return errors.ErrInvalidInput("Cannot remove members: group must have at least 2 members")
	}

	// Update members
	return s.roomRepo.UpdateMembers(ctx, roomID, newMembers)
}

// IsMember checks if a user is a member of a room
func (s *roomService) IsMember(ctx context.Context, roomID, userID string) (bool, error) {
	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return false, err
	}

	for _, member := range room.Members {
		if member == userID {
			return true, nil
		}
	}

	return false, nil
}

// removeDuplicates removes duplicate strings from a slice
func removeDuplicates(slice []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	
	for _, item := range slice {
		if !seen[item] {
			seen[item] = true
			result = append(result, item)
		}
	}
	
	return result
}
