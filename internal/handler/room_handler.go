package handler

import (
	"encoding/json"
	"net/http"
	"strings"

	"realtime-chat-system/internal/middleware"
	"realtime-chat-system/internal/service"
	"realtime-chat-system/pkg/response"
)

// RoomHandler handles room HTTP requests
type RoomHandler struct {
	roomService service.RoomService
	authMw      *middleware.AuthMiddleware
}

// NewRoomHandler creates a new room handler
func NewRoomHandler(roomService service.RoomService, authMw *middleware.AuthMiddleware) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
		authMw:      authMw,
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
func (h *RoomHandler) GetRooms(w http.ResponseWriter, r *http.Request) {
	// Only accept GET method
	if r.Method != http.MethodGet {
		response.BadRequest(w, "Method not allowed")
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	// Get user's rooms
	rooms, err := h.roomService.GetUserRooms(r.Context(), userID)
	if err != nil {
		response.Error(w, err)
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

	response.Success(w, roomResponses, "Rooms retrieved successfully")
}

// CreateGroup handles POST /api/rooms - create a new group room
func (h *RoomHandler) CreateGroup(w http.ResponseWriter, r *http.Request) {
	// Only accept POST method
	if r.Method != http.MethodPost {
		response.BadRequest(w, "Method not allowed")
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	// Parse request body
	var req CreateGroupRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Create group room
	room, err := h.roomService.CreateGroupRoom(r.Context(), req.Name, userID, req.MemberIDs)
	if err != nil {
		response.Error(w, err)
		return
	}

	// Return room data
	roomResp := RoomResponse{
		ID:        room.ID,
		Type:      room.Type,
		Name:      room.Name,
		Members:   room.Members,
		CreatedBy: room.CreatedBy,
		CreatedAt: room.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}

	response.Created(w, roomResp, "Group created successfully")
}

// AddMembers handles POST /api/rooms/:id/members - add members to a group
func (h *RoomHandler) AddMembers(w http.ResponseWriter, r *http.Request) {
	// Only accept POST method
	if r.Method != http.MethodPost {
		response.BadRequest(w, "Method not allowed")
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	// Extract room ID from path
	roomID := extractRoomID(r.URL.Path, "/api/rooms/", "/members")
	if roomID == "" {
		response.BadRequest(w, "Invalid room ID")
		return
	}

	// Check if user is a member of the room
	isMember, err := h.roomService.IsMember(r.Context(), roomID, userID)
	if err != nil {
		response.Error(w, err)
		return
	}
	if !isMember {
		response.Forbidden(w, "You are not a member of this room")
		return
	}

	// Parse request body
	var req AddMembersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Add members
	if err := h.roomService.AddMembers(r.Context(), roomID, req.MemberIDs); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, nil, "Members added successfully")
}

// RemoveMembers handles DELETE /api/rooms/:id/members - remove members from a group
func (h *RoomHandler) RemoveMembers(w http.ResponseWriter, r *http.Request) {
	// Only accept DELETE method
	if r.Method != http.MethodDelete {
		response.BadRequest(w, "Method not allowed")
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "User not authenticated")
		return
	}

	// Extract room ID from path
	roomID := extractRoomID(r.URL.Path, "/api/rooms/", "/members")
	if roomID == "" {
		response.BadRequest(w, "Invalid room ID")
		return
	}

	// Check if user is a member of the room
	isMember, err := h.roomService.IsMember(r.Context(), roomID, userID)
	if err != nil {
		response.Error(w, err)
		return
	}
	if !isMember {
		response.Forbidden(w, "You are not a member of this room")
		return
	}

	// Parse request body
	var req RemoveMembersRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Remove members
	if err := h.roomService.RemoveMembers(r.Context(), roomID, req.MemberIDs); err != nil {
		response.Error(w, err)
		return
	}

	response.Success(w, nil, "Members removed successfully")
}

// RegisterRoutes registers room routes
func (h *RoomHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/rooms", h.authMw.Authenticate(h.handleRooms))
	mux.HandleFunc("/api/rooms/", h.authMw.Authenticate(h.handleRoomMembers))
}

// handleRooms routes between GET (list rooms) and POST (create group)
func (h *RoomHandler) handleRooms(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		h.GetRooms(w, r)
	case http.MethodPost:
		h.CreateGroup(w, r)
	default:
		response.BadRequest(w, "Method not allowed")
	}
}

// handleRoomMembers routes member management endpoints
func (h *RoomHandler) handleRoomMembers(w http.ResponseWriter, r *http.Request) {
	// Check if path ends with /members
	if !strings.HasSuffix(r.URL.Path, "/members") {
		response.NotFound(w, "Endpoint not found")
		return
	}

	switch r.Method {
	case http.MethodPost:
		h.AddMembers(w, r)
	case http.MethodDelete:
		h.RemoveMembers(w, r)
	default:
		response.BadRequest(w, "Method not allowed")
	}
}

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
