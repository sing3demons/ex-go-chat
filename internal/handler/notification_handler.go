package handler

import (
	"net/http"
	"strconv"

	"realtime-chat-system/internal/middleware"
	"realtime-chat-system/internal/service"
	"realtime-chat-system/pkg/errors"
	"realtime-chat-system/pkg/response"
)

// NotificationHandler handles notification HTTP requests
type NotificationHandler struct {
	notificationService service.NotificationService
	authMiddleware      *middleware.AuthMiddleware
}

// NewNotificationHandler creates a new notification handler
func NewNotificationHandler(notificationService service.NotificationService, authMiddleware *middleware.AuthMiddleware) *NotificationHandler {
	return &NotificationHandler{
		notificationService: notificationService,
		authMiddleware:      authMiddleware,
	}
}

// RegisterRoutes registers notification routes
func (h *NotificationHandler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/api/notifications", h.authMiddleware.Authenticate(h.GetNotifications))
	mux.HandleFunc("/api/notifications/pending", h.authMiddleware.Authenticate(h.GetPendingNotifications))
	mux.HandleFunc("/api/notifications/unread-count", h.authMiddleware.Authenticate(h.GetUnreadCount))
	mux.HandleFunc("/api/notifications/mark-read", h.authMiddleware.Authenticate(h.MarkAsRead))
	mux.HandleFunc("/api/notifications/mark-all-read", h.authMiddleware.Authenticate(h.MarkAllAsRead))
}

// GetNotifications handles GET /api/notifications
func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {

		response.Error(w, &errors.AppError{
			StatusCode: http.StatusMethodNotAllowed,
			Message:    "Method not allowed",
		})
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, &errors.AppError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Unauthorized",
		})
		return
	}

	// Parse pagination parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 20
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 {
			limit = l
		}
	}

	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil && o >= 0 {
			offset = o
		}
	}

	notifications, err := h.notificationService.GetNotifications(r.Context(), userID, limit, offset)
	if err != nil {
		response.Error(w, &errors.AppError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to get notifications",
		})
		return
	}

	response.Success(w, &response.SuccessResponse{
		Data:    notifications,
		Success: true,
	}, "Notifications retrieved successfully")
}

// GetPendingNotifications handles GET /api/notifications/pending
func (h *NotificationHandler) GetPendingNotifications(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, &errors.AppError{
			StatusCode: http.StatusMethodNotAllowed,
			Message:    "Method not allowed",
		})
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, &errors.AppError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Unauthorized",
		})
		return
	}
	notifications, err := h.notificationService.GetPendingNotifications(r.Context(), userID)
	if err != nil {
		response.Error(w, &errors.AppError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to get pending notifications",
		})
		return
	}

	response.Success(w, &response.SuccessResponse{
		Data:    notifications,
		Success: true,
	}, "Pending notifications retrieved successfully")
}

// GetUnreadCount handles GET /api/notifications/unread-count
func (h *NotificationHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		response.Error(w, &errors.AppError{
			StatusCode: http.StatusMethodNotAllowed,
			Message:    "Method not allowed",
		})
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, &errors.AppError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Unauthorized",
		})
		return
	}
	count, err := h.notificationService.GetUnreadCount(r.Context(), userID)
	if err != nil {
		response.Error(w, &errors.AppError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to get unread count",
		})
		return
	}

	response.Success(w, &response.SuccessResponse{
		Data:    map[string]int64{"count": count},
		Success: true,
	}, "Unread count retrieved successfully")
}

// MarkAsRead handles POST /api/notifications/mark-read
func (h *NotificationHandler) MarkAsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, &errors.AppError{
			StatusCode: http.StatusMethodNotAllowed,
			Message:    "Method not allowed",
		})
		return
	}

	var req struct {
		NotificationID string `json:"notificationId"`
	}

	if err := response.DecodeJSON(r, &req); err != nil {
		response.Error(w, &errors.AppError{
			StatusCode: http.StatusBadRequest,
			Message:    "Invalid request body",
		})
		return
	}

	if req.NotificationID == "" {
		response.Error(w, &errors.AppError{
			StatusCode: http.StatusBadRequest,
			Message:    "Notification ID is required",
		})
		return
	}

	if err := h.notificationService.MarkAsRead(r.Context(), req.NotificationID); err != nil {
		response.Error(w, &errors.AppError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to mark notification as read",
		})
		return
	}

	response.Success(w, map[string]string{"message": "Notification marked as read"}, "Notification marked as read successfully")
}

// MarkAllAsRead handles POST /api/notifications/mark-all-read
func (h *NotificationHandler) MarkAllAsRead(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response.Error(w, &errors.AppError{
			StatusCode: http.StatusMethodNotAllowed,
			Message:    "Method not allowed",
		})
		return
	}

	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Error(w, &errors.AppError{
			StatusCode: http.StatusUnauthorized,
			Message:    "Unauthorized",
		})
		return
	}
	if err := h.notificationService.MarkAllAsRead(r.Context(), userID); err != nil {
		response.Error(w, &errors.AppError{
			StatusCode: http.StatusInternalServerError,
			Message:    "Failed to mark all notifications as read",
		})
		return
	}

	response.Success(w, map[string]string{"message": "All notifications marked as read"}, "All notifications marked as read successfully")
}
