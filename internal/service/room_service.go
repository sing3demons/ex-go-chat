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
	roomRepo  repository.RoomRepository
	cacheRepo repository.RoomCacheRepository
}

// NewRoomService creates a new room service
func NewRoomService(roomRepo repository.RoomRepository, cacheRepo repository.RoomCacheRepository) RoomService {
	return &roomService{
		roomRepo:  roomRepo,
		cacheRepo: cacheRepo,
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

	// Cache the newly created room
	_ = s.cacheRepo.CacheRoom(ctx, room)

	return room, nil
}

// GetRoom retrieves a room by ID (with caching)
func (s *roomService) GetRoom(ctx context.Context, roomID string) (*models.Room, error) {
	// Try cache first
	if cachedRoom, err := s.cacheRepo.GetCachedRoom(ctx, roomID); err == nil && cachedRoom != nil {
		return cachedRoom, nil
	}

	// Cache miss - fetch from DB
	room, err := s.roomRepo.FindByID(ctx, roomID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	_ = s.cacheRepo.CacheRoom(ctx, room)

	return room, nil
}

// GetUserRooms retrieves all rooms for a user (with caching)
func (s *roomService) GetUserRooms(ctx context.Context, userID string) ([]*models.Room, error) {
	// Try cache first
	if cachedRooms, err := s.cacheRepo.GetCachedUserRooms(ctx, userID); err == nil && cachedRooms != nil {
		return cachedRooms, nil
	}

	// Cache miss - fetch from DB
	rooms, err := s.roomRepo.FindUserRooms(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	_ = s.cacheRepo.CacheUserRooms(ctx, userID, rooms)

	return rooms, nil
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
	if err := s.roomRepo.UpdateMembers(ctx, roomID, newMembers); err != nil {
		return err
	}

	// Invalidate room cache after member change
	_ = s.cacheRepo.InvalidateRoom(ctx, roomID)

	return nil
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
	if err := s.roomRepo.UpdateMembers(ctx, roomID, newMembers); err != nil {
		return err
	}

	// Invalidate room cache after member change
	_ = s.cacheRepo.InvalidateRoom(ctx, roomID)

	return nil
}

// IsMember checks if a user is a member of a room
func (s *roomService) IsMember(ctx context.Context, roomID, userID string) (bool, error) {
	room, err := s.GetRoom(ctx, roomID) // Use cached version
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
