package service

import (
	"context"

	"realtime-chat-system/internal/models"
	"realtime-chat-system/internal/repository"
)

// UserService defines the interface for user operations with caching
type UserService interface {
	GetUserByID(ctx context.Context, userID string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	SearchUsers(ctx context.Context, query string, limit int) ([]*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
}

type userService struct {
	userRepo  repository.UserRepository
	cacheRepo repository.UserCacheRepository
}

// NewUserService creates a new user service with caching
func NewUserService(userRepo repository.UserRepository, cacheRepo repository.UserCacheRepository) UserService {
	return &userService{
		userRepo:  userRepo,
		cacheRepo: cacheRepo,
	}
}

// GetUserByID retrieves a user by ID (with caching)
func (s *userService) GetUserByID(ctx context.Context, userID string) (*models.User, error) {
	// Try cache first
	if cachedUser, err := s.cacheRepo.GetCachedUserByID(ctx, userID); err == nil && cachedUser != nil {
		return cachedUser, nil
	}

	// Cache miss - fetch from DB
	user, err := s.userRepo.FindByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Cache the result
	_ = s.cacheRepo.CacheUser(ctx, user)

	return user, nil
}

// GetUserByUsername retrieves a user by username (with caching)
func (s *userService) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	// Try cache first
	if cachedUser, err := s.cacheRepo.GetCachedUserByUsername(ctx, username); err == nil && cachedUser != nil {
		return cachedUser, nil
	}

	// Cache miss - fetch from DB
	user, err := s.userRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	// Cache the result
	_ = s.cacheRepo.CacheUser(ctx, user)

	return user, nil
}

// SearchUsers searches for users by username (no caching for search results)
func (s *userService) SearchUsers(ctx context.Context, query string, limit int) ([]*models.User, error) {
	return s.userRepo.SearchByUsername(ctx, query, limit)
}

// CreateUser creates a new user and invalidates relevant caches
func (s *userService) CreateUser(ctx context.Context, user *models.User) error {
	if err := s.userRepo.Create(ctx, user); err != nil {
		return err
	}

	// Cache the newly created user
	_ = s.cacheRepo.CacheUser(ctx, user)

	return nil
}
