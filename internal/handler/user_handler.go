package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"realtime-chat-system/internal/middleware"
	"realtime-chat-system/internal/repository"
	"realtime-chat-system/internal/service"
	"realtime-chat-system/pkg/response"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userRepo    repository.UserRepository
	roomService service.RoomService
	authMw      *middleware.AuthMiddleware
}

// NewUserHandler creates a new user handler
func NewUserHandler(userRepo repository.UserRepository, roomService service.RoomService, authMw *middleware.AuthMiddleware) *UserHandler {
	return &UserHandler{
		userRepo:    userRepo,
		roomService: roomService,
		authMw:      authMw,
	}
}

// UserSearchResponse represents a user in search results
type UserSearchResponse struct {
	ID       string `json:"id"`
	Username string `json:"username"`
}

// CreateDirectChatRequest represents a request to create direct chat
type CreateDirectChatRequest struct {
	Username string `json:"username"`
}

// SearchUsers handles GET /api/users/search?q=query - search users by username
func (h *UserHandler) SearchUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.BadRequest(w, "Method not allowed")
		return
	}

	// Get current user ID
	currentUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	// Get search query
	query := strings.TrimSpace(r.URL.Query().Get("q"))
	if query == "" {
		response.Success(w, []UserSearchResponse{}, "No query provided")
		return
	}

	// Search users
	users, err := h.userRepo.SearchByUsername(r.Context(), query, 10)
	if err != nil {
		response.Error(w, err)
		return
	}

	// Convert to response format, excluding current user
	var results []UserSearchResponse
	for _, user := range users {
		if user.ID != currentUserID {
			results = append(results, UserSearchResponse{
				ID:       user.ID,
				Username: user.Username,
			})
		}
	}

	if results == nil {
		results = []UserSearchResponse{}
	}

	response.Success(w, results, "Users found")
}

// CreateDirectChat handles POST /api/users/chat - create or get direct chat with user
func (h *UserHandler) CreateDirectChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.BadRequest(w, "Method not allowed")
		return
	}

	// Get current user ID
	currentUserID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	// Parse request
	var req CreateDirectChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	username := strings.TrimSpace(req.Username)
	if username == "" {
		response.BadRequest(w, "Username is required")
		return
	}

	// Find target user by username
	targetUser, err := h.userRepo.FindByUsername(r.Context(), username)
	if err != nil {
		response.NotFound(w, "User not found")
		return
	}

	// Cannot chat with yourself
	if targetUser.ID == currentUserID {
		response.BadRequest(w, "Cannot create chat with yourself")
		return
	}

	// Create or get existing direct room
	room, err := h.roomService.CreateDirectRoom(r.Context(), currentUserID, targetUser.ID)
	if err != nil {
		response.Error(w, err)
		return
	}

	// Return room data
	roomResp := RoomResponse{
		ID:        room.ID,
		Type:      room.Type,
		Name:      targetUser.Username, // Use target user's name as room name for direct chat
		Members:   room.Members,
		CreatedBy: room.CreatedBy,
		CreatedAt: room.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Success(w, roomResp, "Direct chat created")
}

// RegisterRoutes registers user routes
func (h *UserHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/users/search", h.authMw.Authenticate(h.SearchUsers))
	mux.HandleFunc("/api/users/chat", h.authMw.Authenticate(h.CreateDirectChat))
}
