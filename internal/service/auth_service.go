package service

import (
	"context"
	"strings"
	"time"

	"realtime-chat-system/internal/models"
	"realtime-chat-system/internal/repository"
	"realtime-chat-system/pkg/auth"
	"realtime-chat-system/pkg/errors"
	"realtime-chat-system/pkg/validator"
)

// AuthService defines the interface for authentication operations
type AuthService interface {
	Register(ctx context.Context, username, email, password, ssid string) (*models.User, string, error)
	Login(ctx context.Context, identifier, password, ssid string) (string, error)
	ValidateToken(ctx context.Context, token string) (*auth.Claims, error)
}

// authService implements AuthService
type authService struct {
	userRepo   repository.UserRepository
	jwtManager *auth.JWTManager
}

// NewAuthService creates a new authentication service
func NewAuthService(userRepo repository.UserRepository, jwtManager *auth.JWTManager) AuthService {
	return &authService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Register registers a new user
func (s *authService) Register(ctx context.Context, username, email, password, ssid string) (*models.User, string, error) {
	// Trim whitespace
	username = strings.TrimSpace(username)
	email = strings.TrimSpace(email)

	// Validate username
	if err := validator.ValidateUsername(username); err != nil {
		return nil, "", err
	}

	// Validate email
	if err := validator.ValidateEmail(email); err != nil {
		return nil, "", err
	}

	// Validate password
	if err := validator.ValidatePassword(password); err != nil {
		return nil, "", err
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(password)
	if err != nil {
		return nil, "", errors.ErrInternal("Failed to hash password")
	}

	// Create user
	now := time.Now()
	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: hashedPassword,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// Save to database
	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, "", err
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(user.ID, user.Username, ssid)
	if err != nil {
		return nil, "", errors.ErrInternal("Failed to generate token")
	}

	return user, token, nil
}

// Login authenticates a user and returns a JWT token
func (s *authService) Login(ctx context.Context, identifier, password, ssid string) (string, error) {
	// Trim whitespace
	identifier = strings.TrimSpace(identifier)

	if identifier == "" {
		return "", errors.ErrInvalidInput("Username or email is required")
	}

	if password == "" {
		return "", errors.ErrInvalidInput("Password is required")
	}

	// Try to find user by username or email
	var user *models.User
	var err error

	// Check if identifier is an email
	if strings.Contains(identifier, "@") {
		user, err = s.userRepo.FindByEmail(ctx, identifier)
	} else {
		user, err = s.userRepo.FindByUsername(ctx, identifier)
	}

	if err != nil {
		// Return generic error to avoid user enumeration
		return "", errors.ErrInvalidCredentials()
	}

	// Compare password
	if err := auth.ComparePassword(user.PasswordHash, password); err != nil {
		return "", errors.ErrInvalidCredentials()
	}

	// Generate JWT token ssid
	token, err := s.jwtManager.GenerateToken(user.ID, user.Username, ssid)
	if err != nil {
		return "", errors.ErrInternal("Failed to generate token")
	}

	return token, nil
}

// ValidateToken validates a JWT token and returns the claims
func (s *authService) ValidateToken(ctx context.Context, token string) (*auth.Claims, error) {
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return nil, errors.ErrInvalidToken()
	}

	return claims, nil
}
