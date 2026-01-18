package handler

import (
	"errors"
	"net/http"
	"strings"

	"realtime-chat-system/internal/middleware"
	"realtime-chat-system/internal/repository"
	"realtime-chat-system/internal/service"
	"realtime-chat-system/pkg/kp"
	"realtime-chat-system/pkg/response"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userRepo    repository.UserRepository
	roomService service.RoomService
	authMw      *middleware.AuthMiddleware
	hub         HubInterface
}

// HubInterface defines the methods we need from the WebSocket hub
type HubInterface interface {
	NotifyRoomCreated(roomID, roomType, name string, members []string)
}

// NewUserHandler creates a new user handler
func NewUserHandler(userRepo repository.UserRepository, roomService service.RoomService, authMw *middleware.AuthMiddleware, hub HubInterface) *UserHandler {
	return &UserHandler{
		userRepo:    userRepo,
		roomService: roomService,
		authMw:      authMw,
		hub:         hub,
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
func (h *UserHandler) SearchUsers(ctx *kp.Ctx) {
	ctx.L("search_users")
	var customError *kp.Error
	// ctx.JSONError()

	// Get current user ID
	currentUserID, ok := middleware.GetUserID(ctx)
	if !ok {
		// response.Unauthorized(w, "User not authenticated")
		customError = &kp.Error{
			StatusCode: http.StatusUnauthorized,
			Message:    "unauthorized",
			Err:        errors.New("User not authenticated"),
		}
		ctx.JSONError(customError)
		return
	}

	// Get search query
	// query := strings.TrimSpace(r.URL.Query().Get("q"))
	query := ctx.Query("q")
	if query == "" {
		ctx.Log.AddMetadata("ErrorCause", "No query provided")
		ctx.JSON(http.StatusOK, []UserSearchResponse{})

		return
	}

	// Search users
	users, err := h.userRepo.SearchByUsername(ctx, query, 10)
	if err != nil {
		// response.Error(w, err)
		if !errors.As(err, &customError) {
			customError = &kp.Error{
				Message:    "internal_server",
				StatusCode: http.StatusInternalServerError,
				Err:        err,
			}
		}
		ctx.JSONError(customError)
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

	// ctx.JSON(http.StatusOK, results)
	ctx.JSON(http.StatusOK, response.SuccessResponse{
		Data:    results,
		Success: true,
		Message: "Users retrieved successfully",
	})
}

// CreateDirectChat handles POST /api/users/chat - create or get direct chat with user
func (h *UserHandler) CreateDirectChat(ctx *kp.Ctx) {
	ctx.L("create_direct_chat")
	var customError *kp.Error
	// ctx.JSONError()
	// if r.Method != http.MethodPost {
	// 	response.BadRequest(w, "Method not allowed")
	// 	return
	// }

	// Get current user ID
	currentUserID, ok := middleware.GetUserID(ctx)
	if !ok {
		// response.Unauthorized(w, "User not authenticated")
		customError = &kp.Error{
			StatusCode: http.StatusUnauthorized,
			Message:    "unauthorized",
			Err:        errors.New("User not authenticated"),
		}
		ctx.JSONError(customError)
		return
	}

	// Parse request
	var req CreateDirectChatRequest
	if err := ctx.Bind(&req); err != nil {
		// response.BadRequest(w, "Invalid request body")
		customError = &kp.Error{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid_request",
			Err:        err,
		}
		ctx.JSONError(customError)
		return
	}
	// if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
	// 	response.BadRequest(w, "Invalid request body")
	// 	return
	// }

	username := strings.TrimSpace(req.Username)
	if username == "" {
		customError = &kp.Error{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid_request",
			Err:        errors.New("Username is required"),
		}
		ctx.JSONError(customError)
		return
	}

	// Find target user by username
	targetUser, err := h.userRepo.FindByUsername(ctx, username)
	if err != nil {
		// response.NotFound(w, "User not found")
		if !errors.As(err, &customError) {
			customError = &kp.Error{
				StatusCode: http.StatusInternalServerError,
				Message:    "internal_server",
				Err:        err,
			}
		}
		ctx.JSONError(customError)
		return
	}

	// Cannot chat with yourself
	if targetUser.ID == currentUserID {
		// response.BadRequest(w, "Cannot create chat with yourself")
		customError = &kp.Error{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid_request",
			Err:        errors.New("Cannot create chat with yourself"),
		}
		ctx.JSONError(customError)
		return
	}

	// Create or get existing direct room
	room, err := h.roomService.CreateDirectRoom(ctx, currentUserID, targetUser.ID)
	if err != nil {
		if !errors.As(err, &customError) {
			customError = &kp.Error{
				StatusCode: http.StatusInternalServerError,
				Message:    "internal_server",
				Err:        err,
			}
		}
		ctx.JSONError(customError)
		return
	}

	// Notify WebSocket hub about the new room (only if it's actually new)
	// The room service returns existing room if it already exists
	h.hub.NotifyRoomCreated(room.ID, room.Type, targetUser.Username, room.Members)

	// Return room data
	roomResp := RoomResponse{
		ID:        room.ID,
		Type:      room.Type,
		Name:      targetUser.Username, // Use target user's name as room name for direct chat
		Members:   room.Members,
		CreatedBy: room.CreatedBy,
		CreatedAt: room.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// ctx.JSON(http.StatusOK, roomResp)
	ctx.JSON(http.StatusOK, response.SuccessResponse{
		Data:    roomResp,
		Success: true,
		Message: "Direct chat created successfully",
	})
}

// RegisterRoutes registers user routes
func (h *UserHandler) RegisterRoutes(app kp.IMicroservice) {
	app.GET("/api/users/search", h.SearchUsers, h.authMw.Authenticate)
	app.POST("/api/users/chat", h.CreateDirectChat, h.authMw.Authenticate)
}
