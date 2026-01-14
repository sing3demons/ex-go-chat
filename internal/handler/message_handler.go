package handler

import (
	"net/http"
	"strconv"
	"strings"

	"realtime-chat-system/internal/middleware"
	"realtime-chat-system/internal/service"
	"realtime-chat-system/pkg/response"
)

// MessageHandler handles message HTTP requests
type MessageHandler struct {
	messageService service.MessageService
	roomService    service.RoomService
	authMw         *middleware.AuthMiddleware
}

// NewMessageHandler creates a new message handler
func NewMessageHandler(messageService service.MessageService, roomService service.RoomService, authMw *middleware.AuthMiddleware) *MessageHandler {
	return &MessageHandler{
		messageService: messageService,
		roomService:    roomService,
		authMw:         authMw,
	}
}

// MessageResponse represents a message in API responses
type MessageResponse struct {
	ID        string                  `json:"id"`
	RoomID    string                  `json:"roomId"`
	SenderID  string                  `json:"senderId"`
	Content   string                  `json:"content"`
	CreatedAt string                  `json:"createdAt"`
	UpdatedAt string                  `json:"updatedAt"`
	EditedAt  *string                 `json:"editedAt,omitempty"`
	Deleted   bool                    `json:"deleted"`
	Status    map[string]StatusResponse `json:"status,omitempty"`
}

// StatusResponse represents message status in API responses
type StatusResponse struct {
	Delivered   bool    `json:"delivered"`
	DeliveredAt *string `json:"deliveredAt,omitempty"`
	Read        bool    `json:"read"`
	ReadAt      *string `json:"readAt,omitempty"`
}

// GetMessages handles GET /api/rooms/:id/messages - get chat history
func (h *MessageHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
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

	// Extract room ID from path
	roomID := extractRoomIDFromPath(r.URL.Path)
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

	// Get pagination parameters from query string
	limit := 50 // default
	offset := 0 // default

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get messages
	messages, err := h.messageService.GetMessages(r.Context(), roomID, limit, offset)
	if err != nil {
		response.Error(w, err)
		return
	}

	// Convert to response format
	messageResponses := make([]MessageResponse, len(messages))
	for i, msg := range messages {
		msgResp := MessageResponse{
			ID:        msg.ID,
			RoomID:    msg.RoomID,
			SenderID:  msg.SenderID,
			Content:   msg.Content,
			CreatedAt: msg.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
			UpdatedAt: msg.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
			Deleted:   msg.Deleted,
		}

		if msg.EditedAt != nil {
			editedAt := msg.EditedAt.Format("2006-01-02T15:04:05Z07:00")
			msgResp.EditedAt = &editedAt
		}

		// Convert status map
		if msg.Status != nil {
			statusMap := make(map[string]StatusResponse)
			for uid, status := range msg.Status {
				statusResp := StatusResponse{
					Delivered: status.Delivered,
					Read:      status.Read,
				}
				if status.DeliveredAt != nil {
					deliveredAt := status.DeliveredAt.Format("2006-01-02T15:04:05Z07:00")
					statusResp.DeliveredAt = &deliveredAt
				}
				if status.ReadAt != nil {
					readAt := status.ReadAt.Format("2006-01-02T15:04:05Z07:00")
					statusResp.ReadAt = &readAt
				}
				statusMap[uid] = statusResp
			}
			msgResp.Status = statusMap
		}

		messageResponses[i] = msgResp
	}

	response.Success(w, messageResponses, "Messages retrieved successfully")
}

// RegisterRoutes registers message routes
func (h *MessageHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/rooms/", h.authMw.Authenticate(h.handleRoomMessages))
}

// handleRoomMessages routes message-related endpoints
func (h *MessageHandler) handleRoomMessages(w http.ResponseWriter, r *http.Request) {
	// Check if path ends with /messages
	if strings.HasSuffix(r.URL.Path, "/messages") {
		h.GetMessages(w, r)
		return
	}

	// If not /messages endpoint, let room handler deal with it
	// This is handled by room handler's routes
	response.NotFound(w, "Endpoint not found")
}

// extractRoomIDFromPath extracts room ID from URL path like /api/rooms/:id/messages
func extractRoomIDFromPath(path string) string {
	// Remove prefix /api/rooms/
	path = strings.TrimPrefix(path, "/api/rooms/")
	
	// Remove suffix /messages
	path = strings.TrimSuffix(path, "/messages")
	
	return path
}
