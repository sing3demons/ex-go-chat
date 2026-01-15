package handler

import (
	"encoding/json"
	"net/http"

	"realtime-chat-system/internal/service"
	"realtime-chat-system/pkg/response"
)

// AuthHandler handles authentication HTTP requests
type AuthHandler struct {
	authService service.AuthService
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// RegisterRequest represents a registration request
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginRequest represents a login request
type LoginRequest struct {
	Identifier string `json:"identifier"` // username or email
	Password   string `json:"password"`
}

// AuthResponse represents an authentication response
type AuthResponse struct {
	Token string      `json:"token"`
}

// UserResponse represents a user in API responses
type UserResponse struct {
	ID        string `json:"id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	CreatedAt string `json:"createdAt"`
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	// Only accept POST method
	if r.Method != http.MethodPost {
		response.BadRequest(w, "Method not allowed")
		return
	}

	// Parse request body
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Register user
	user, err := h.authService.Register(r.Context(), req.Username, req.Email, req.Password)
	if err != nil {
		response.Error(w, err)
		return
	}

	// Return user data (without password hash)
	userResp := UserResponse{
		ID:        user.ID,
		Username:  user.Username,
		Email:     user.Email,
		CreatedAt: user.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Created(w, userResp, "User registered successfully")
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	// Only accept POST method
	if r.Method != http.MethodPost {
		response.BadRequest(w, "Method not allowed")
		return
	}

	// Parse request body
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Login user
	token, err := h.authService.Login(r.Context(), req.Identifier, req.Password)
	if err != nil {
		response.Error(w, err)
		return
	}

	// Return token
	authResp := AuthResponse{
		Token: token,
	}

	response.Success(w, authResp, "Login successful")
}

// RegisterRoutes registers auth routes
func (h *AuthHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/auth/register", h.Register)
	mux.HandleFunc("/api/auth/login", h.Login)
}
