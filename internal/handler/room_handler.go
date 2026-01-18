package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"realtime-chat-system/internal/middleware"
	"realtime-chat-system/internal/service"
	"realtime-chat-system/pkg/kp"
	"realtime-chat-system/pkg/response"
)

// RoomHandler handles room HTTP requests
type RoomHandler struct {
	roomService service.RoomService
	authMw      *middleware.AuthMiddleware
	hub         HubInterface
}

// NewRoomHandler creates a new room handler
func NewRoomHandler(roomService service.RoomService, authMw *middleware.AuthMiddleware, hub HubInterface) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
		authMw:      authMw,
		hub:         hub,
	}
}

// CreateGroupRequest represents a create group request
type CreateGroupRequest struct {
	Name      string   `json:"name"`
	MemberIDs []string `json:"memberIds"`
}

// AddMembersRequest represents an add members request
type AddMembersRequest struct {
	MemberIDs []string `json:"memberIds"`
}

// RemoveMembersRequest represents a remove members request
type RemoveMembersRequest struct {
	MemberIDs []string `json:"memberIds"`
}

// RoomResponse represents a room in API responses
type RoomResponse struct {
	ID        string   `json:"id"`
	Type      string   `json:"type"`
	Name      string   `json:"name,omitempty"`
	Members   []string `json:"members"`
	CreatedBy string   `json:"createdBy"`
	CreatedAt string   `json:"createdAt"`
}

// GetRooms handles GET /api/rooms - list user's rooms
func (h *RoomHandler) GetRooms(ctx *kp.Ctx) {
	ctx.L("get_rooms")
	var customError *kp.Error
	// Only accept GET method
	if ctx.Req.Method != http.MethodGet {
		response.BadRequest(ctx.Res, "Method not allowed")
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		// response.Unauthorized(w, "User not authenticated")
		return
	}

	// Get user's rooms
	rooms, err := h.roomService.GetUserRooms(ctx, userID)
	if err != nil {
		// response.Error(w, err)
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

	// Convert to response format
	roomResponses := make([]RoomResponse, len(rooms))
	for i, room := range rooms {
		roomResponses[i] = RoomResponse{
			ID:        room.ID,
			Type:      room.Type,
			Name:      room.Name,
			Members:   room.Members,
			CreatedBy: room.CreatedBy,
			CreatedAt: room.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}

	ctx.JSON(http.StatusOK, response.SuccessResponse{
		Data:    roomResponses,
		Success: true,
		Message: "Rooms retrieved successfully",
	})
}

// CreateGroup handles POST /api/rooms - create a new group room
func (h *RoomHandler) CreateGroup(ctx *kp.Ctx) {
	ctx.L("create_group")
	var customError *kp.Error
	// Only accept POST method
	if ctx.Req.Method != http.MethodPost {
		response.BadRequest(ctx.Res, "Method not allowed")
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(ctx)
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

	// Parse request body
	var req CreateGroupRequest
	if err := json.NewDecoder(ctx.Req.Body).Decode(&req); err != nil {
		// response.BadRequest(ctx.Res, "Invalid request body")
		customError = &kp.Error{
			StatusCode: http.StatusBadRequest,
			Message:    "invalid_request",
			Err:        err,
		}
		ctx.JSONError(customError)
		return
	}

	// Create group room
	room, err := h.roomService.CreateGroupRoom(ctx, req.Name, userID, req.MemberIDs)
	if err != nil {
		// response.Error(w, err)
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

	// Notify WebSocket hub about the new room
	h.hub.NotifyRoomCreated(room.ID, room.Type, room.Name, room.Members)

	// Return room data
	roomResp := RoomResponse{
		ID:        room.ID,
		Type:      room.Type,
		Name:      room.Name,
		Members:   room.Members,
		CreatedBy: room.CreatedBy,
		CreatedAt: room.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	// response.Created(w, roomResp, "Group created successfully")
	ctx.JSON(http.StatusCreated, response.SuccessResponse{
		Data:    roomResp,
		Success: true,
		Message: "Group created successfully",
	})
}

// AddMembers handles POST /api/rooms/:id/members - add members to a group
func (h *RoomHandler) AddMembers(ctx *kp.Ctx) {
	// Only accept POST method
	// if r.Method != http.MethodPost {
	// 	response.BadRequest(w, "Method not allowed")
	// 	return
	// }
	ctx.L("add_members")
	var customError *kp.Error

	// Get user ID from context
	userID, ok := middleware.GetUserID(ctx)
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

	// Extract room ID from path
	roomID := extractRoomID(ctx.Req.URL.Path, "/api/rooms/", "/members")
	if roomID == "" {
		// response.BadRequest(w, "Invalid room ID")
		customError = &kp.Error{
			StatusCode: http.StatusBadRequest,
			Message:    "bad_request",
			Err:        errors.New("Invalid room ID"),
		}
		ctx.JSONError(customError)
		return
	}

	// Check if user is a member of the room
	isMember, err := h.roomService.IsMember(ctx, roomID, userID)
	if err != nil {
		// response.Error(w, err)
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
	if !isMember {
		// response.Forbidden(w, "You are not a member of this room")
		customError = &kp.Error{
			StatusCode: http.StatusForbidden,
			Message:    "forbidden",
			Err:        errors.New("You are not a member of this room"),
		}
		ctx.JSONError(customError)
		return
	}

	// Parse request body
	var req AddMembersRequest
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

	// Add members
	if err := h.roomService.AddMembers(ctx, roomID, req.MemberIDs); err != nil {
		// response.Error(w, err)
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

	// response.Success(w, nil, "Members added successfully")
	ctx.JSON(http.StatusOK, response.SuccessResponse{
		Data:    nil,
		Success: true,
		Message: "Members added successfully",
	})
}

// RemoveMembers handles DELETE /api/rooms/:id/members - remove members from a group
func (h *RoomHandler) RemoveMembers(ctx *kp.Ctx) {
	ctx.L("remove_members")
	var customError *kp.Error
	// Only accept DELETE method
	// if r.Method != http.MethodDelete {
	// 	response.BadRequest(w, "Method not allowed")
	// 	return
	// }

	// Get user ID from context
	userID, ok := middleware.GetUserID(ctx)
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

	// Extract room ID from path
	roomID := extractRoomID(ctx.Req.URL.Path, "/api/rooms/", "/members")
	if roomID == "" {
		// response.BadRequest(w, "Invalid room ID")
		customError = &kp.Error{
			StatusCode: http.StatusBadRequest,
			Message:    "bad_request",
			Err:        errors.New("Invalid room ID"),
		}
		ctx.JSONError(customError)
		return
	}

	// Check if user is a member of the room
	isMember, err := h.roomService.IsMember(ctx, roomID, userID)
	if err != nil {
		// response.Error(w, err)
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
	if !isMember {
		customError = &kp.Error{
			StatusCode: http.StatusForbidden,
			Message:    "forbidden",
			Err:        errors.New("You are not a member of this room"),
		}
		ctx.JSONError(customError)
		return
	}

	// Parse request body
	var req RemoveMembersRequest
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

	// Remove members
	if err := h.roomService.RemoveMembers(ctx, roomID, req.MemberIDs); err != nil {
		// response.Error(w, err)
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

	// response.Success(w, nil, "Members removed successfully")
	ctx.JSON(http.StatusOK, response.SuccessResponse{
		Data:    nil,
		Success: true,
		Message: "Members removed successfully",
	})
}

// RegisterRoutes registers room routes
func (h *RoomHandler) RegisterRoutes(app kp.IMicroservice) {
	// app.HandleFunc("/api/rooms/members", h.authMw.Authenticate(h.handleRoomMembers))
	app.POST("/api/rooms/members", h.AddMembers, h.authMw.Authenticate)
	app.DELETE("/api/rooms/members", h.RemoveMembers, h.authMw.Authenticate)
	app.GET("/api/rooms", h.GetRooms, h.authMw.Authenticate)
	app.POST("/api/rooms", h.CreateGroup, h.authMw.Authenticate)
}

// handleRooms routes between GET (list rooms) and POST (create group)
// func (h *RoomHandler) handleRooms(w http.ResponseWriter, r *http.Request) {
// 	switch r.Method {
// 	case http.MethodGet:
// 		h.GetRooms(w, r)
// 	case http.MethodPost:
// 		h.CreateGroup(w, r)
// 	default:
// 		response.BadRequest(w, "Method not allowed")
// 	}
// }

// handleRoomMembers routes member management endpoints
// func (h *RoomHandler) handleRoomMembers(w http.ResponseWriter, r *http.Request) {
// 	// Check if path ends with /members
// 	if !strings.HasSuffix(r.URL.Path, "/members") {
// 		response.NotFound(w, "Endpoint not found")
// 		return
// 	}

// 	switch r.Method {
// 	case http.MethodPost:
// 		h.AddMembers(w, r)
// 	case http.MethodDelete:
// 		h.RemoveMembers(w, r)
// 	default:
// 		response.BadRequest(w, "Method not allowed")
// 	}
// }

// extractRoomID extracts room ID from URL path
func extractRoomID(path, prefix, suffix string) string {
	// Remove prefix
	if !strings.HasPrefix(path, prefix) {
		return ""
	}
	path = strings.TrimPrefix(path, prefix)

	// Remove suffix
	if !strings.HasSuffix(path, suffix) {
		return ""
	}
	path = strings.TrimSuffix(path, suffix)

	return path
}
