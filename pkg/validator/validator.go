package validator

import (
	"regexp"
	"strings"
	"unicode"

	"realtime-chat-system/pkg/errors"
)

var (
	// Email regex pattern
	emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)
	// Username regex pattern (alphanumeric and underscore only)
	usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]+$`)
)

// ValidateUsername validates username format
// Rules: 3-30 characters, alphanumeric and underscore only
func ValidateUsername(username string) error {
	username = strings.TrimSpace(username)
	
	if len(username) < 3 {
		return errors.ErrInvalidInput("Username must be at least 3 characters")
	}
	
	if len(username) > 30 {
		return errors.ErrInvalidInput("Username must not exceed 30 characters")
	}
	
	if !usernameRegex.MatchString(username) {
		return errors.ErrInvalidInput("Username can only contain letters, numbers, and underscores")
	}
	
	return nil
}

// ValidateEmail validates email format
func ValidateEmail(email string) error {
	email = strings.TrimSpace(email)
	
	if email == "" {
		return errors.ErrInvalidInput("Email is required")
	}
	
	if !emailRegex.MatchString(email) {
		return errors.ErrInvalidEmail()
	}
	
	return nil
}

// ValidatePassword validates password strength
// Rules: minimum 8 characters, at least one uppercase, one lowercase, one number
func ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.ErrWeakPassword()
	}
	
	var (
		hasUpper   bool
		hasLower   bool
		hasNumber  bool
	)
	
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		}
	}
	
	if !hasUpper || !hasLower || !hasNumber {
		return errors.ErrWeakPassword()
	}
	
	return nil
}

// ValidateMessageContent validates message content
// Rules: 1-10000 characters
func ValidateMessageContent(content string) error {
	content = strings.TrimSpace(content)
	
	if len(content) == 0 {
		return errors.ErrInvalidInput("Message content cannot be empty")
	}
	
	if len(content) > 10000 {
		return errors.ErrInvalidInput("Message content must not exceed 10000 characters")
	}
	
	return nil
}

// ValidateRoomName validates room name for group chats
// Rules: 1-100 characters
func ValidateRoomName(name string) error {
	name = strings.TrimSpace(name)
	
	if len(name) == 0 {
		return errors.ErrInvalidInput("Room name cannot be empty")
	}
	
	if len(name) > 100 {
		return errors.ErrInvalidInput("Room name must not exceed 100 characters")
	}
	
	return nil
}
