package handler

import (
	"errors"
	"net/http"
	"strconv"
	"strings"

	"realtime-chat-system/internal/middleware"
	"realtime-chat-system/internal/service"
	"realtime-chat-system/pkg/kp"
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
	ID        string                    `json:"id"`
	RoomID    string                    `json:"roomId"`
	SenderID  string                    `json:"senderId"`
	Content   string                    `json:"content"`
	CreatedAt string                    `json:"createdAt"`
	UpdatedAt string                    `json:"updatedAt"`
	EditedAt  *string                   `json:"editedAt,omitempty"`
	Deleted   bool                      `json:"deleted"`
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
func (h *MessageHandler) GetMessages(ctx *kp.Ctx) {
	ctx.L("get_messages")
	var customError *kp.Error
	// Only accept GET method
	// if r.Method != http.MethodGet {
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
			Err:        nil,
		}
		ctx.JSONError(customError)
		return
	}

	// Extract room ID from path
	// roomID := extractRoomIDFromPath(r.URL.Path)
	roomID := ctx.Params("roomId")
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

	// Get pagination parameters from query string
	limit := 50 // default
	offset := 0 // default

	if limitStr := ctx.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr := ctx.Query("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	// Get messages
	messages, err := h.messageService.GetMessages(ctx, roomID, limit, offset)
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

	// response.Success(w, messageResponses, "Messages retrieved successfully")
	ctx.JSON(http.StatusOK, response.SuccessResponse{
		Data:    messageResponses,
		Success: true,
		Message: "Messages retrieved successfully",
	})
}

// RegisterRoutes registers message routes
func (h *MessageHandler) RegisterRoutes(app kp.IMicroservice) {
	// Use a pattern that matches /api/rooms/{id}/messages
	// mux.HandleFunc("/api/messages/room/", h.authMw.Authenticate(h.handleRoomMessages))
	app.GET("/api/rooms/{roomId}/messages", h.GetMessages, h.authMw.Authenticate)
	app.GET("/api/messages/room/{roomId}", h.GetMessages, h.authMw.Authenticate)
}

// handleRoomMessages routes message-related endpoints
func (h *MessageHandler) handleRoomMessages(w http.ResponseWriter, r *http.Request) {
	// h.GetMessages(w, r)
}

// extractRoomIDFromPath extracts room ID from URL path like /api/messages/room/:id
func extractRoomIDFromPath(path string) string {
	// Remove prefix /api/messages/room/
	path = strings.TrimPrefix(path, "/api/messages/room/")

	return path
}
