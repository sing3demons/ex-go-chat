package middleware

import (
	"context"
	"net/http"
	"strings"

	"realtime-chat-system/internal/service"
	"realtime-chat-system/pkg/auth"
	"realtime-chat-system/pkg/response"
)

// ContextKey is a custom type for context keys
type ContextKey string

const (
	// UserClaimsKey is the context key for user claims
	UserClaimsKey ContextKey = "userClaims"
)

// AuthMiddleware provides JWT authentication middleware
type AuthMiddleware struct {
	authService service.AuthService
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(authService service.AuthService) *AuthMiddleware {
	return &AuthMiddleware{
		authService: authService,
	}
}

// Authenticate is a middleware that validates JWT token
func (m *AuthMiddleware) Authenticate(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Extract token from Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			response.Unauthorized(w, "Authorization header is required")
			return
		}

		// Check if it's a Bearer token
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(w, "Invalid authorization header format. Use: Bearer <token>")
			return
		}

		token := parts[1]

		// Validate token
		claims, err := m.authService.ValidateToken(r.Context(), token)
		if err != nil {
			response.Unauthorized(w, "Invalid or expired token")
			return
		}

		// Add claims to context
		ctx := context.WithValue(r.Context(), UserClaimsKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

// GetUserClaims extracts user claims from context
func GetUserClaims(ctx context.Context) (*auth.Claims, bool) {
	claims, ok := ctx.Value(UserClaimsKey).(*auth.Claims)
	return claims, ok
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) (string, bool) {
	claims, ok := GetUserClaims(ctx)
	if !ok {
		return "", false
	}
	return claims.UserID, true
}
