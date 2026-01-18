package handler

import (
	"errors"
	"net/http"
	"strconv"

	"realtime-chat-system/internal/middleware"
	"realtime-chat-system/internal/service"
	"realtime-chat-system/pkg/kp"
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
func (h *NotificationHandler) RegisterRoutes(app kp.IMicroservice) {
	// mux.HandleFunc("/api/notifications", h.authMiddleware.Authenticate(h.GetNotifications))
	app.GET("/api/notifications", h.GetNotifications, h.authMiddleware.Authenticate)
	// mux.HandleFunc("/api/notifications/pending", h.authMiddleware.Authenticate(h.GetPendingNotifications))
	app.GET("/api/notifications/pending", h.GetPendingNotifications, h.authMiddleware.Authenticate)
	// mux.HandleFunc("/api/notifications/unread-count", h.authMiddleware.Authenticate(h.GetUnreadCount))
	app.GET("/api/notifications/unread-count", h.GetUnreadCount, h.authMiddleware.Authenticate)
	// mux.HandleFunc("/api/notifications/mark-read", h.authMiddleware.Authenticate(h.MarkAsRead))
	app.POST("/api/notifications/mark-read", h.MarkAsRead, h.authMiddleware.Authenticate)
	// mux.HandleFunc("/api/notifications/mark-all-read", h.authMiddleware.Authenticate(h.MarkAllAsRead))
	app.POST("/api/notifications/mark-all-read", h.MarkAllAsRead, h.authMiddleware.Authenticate)
}

// GetNotifications handles GET /api/notifications
func (h *NotificationHandler) GetNotifications(ctx *kp.Ctx) {
	ctx.L("get_notifications")
	var customError *kp.Error
	// Only accept GET method
	// if r.Method != http.MethodGet {

	// 	response.Error(w, &errors.AppError{
	// 		StatusCode: http.StatusMethodNotAllowed,
	// 		Message:    "Method not allowed",
	// 	})
	// 	return
	// }

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		// response.Error(w, &errors.AppError{
		// 	StatusCode: http.StatusUnauthorized,
		// 	Message:    "Unauthorized",
		// })
		customError = &kp.Error{
			StatusCode: http.StatusUnauthorized,
			Message:    "unauthorized",
			Err:        errors.New("Unauthorized"),
		}
		ctx.JSONError(customError)
		return
	}

	// Parse pagination parameters
	limitStr := ctx.Query("limit")
	offsetStr := ctx.Query("offset")

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

	notifications, err := h.notificationService.GetNotifications(ctx, userID, limit, offset)
	if err != nil {
		// response.Error(w, &errors.AppError{
		// 	StatusCode: http.StatusInternalServerError,
		// 	Message:    "Failed to get notifications",
		// })
		customError = &kp.Error{
			StatusCode: http.StatusInternalServerError,
			Message:    "internal_server",
			Err:        err,
		}
		ctx.JSONError(customError)
		return
	}

	ctx.JSON(http.StatusOK, &response.SuccessResponse{
		Data:    notifications,
		Success: true,
		Message: "Notifications retrieved successfully",
	})
	// ctx.JSON(http.StatusOK, map[string]any{
	// 	"notifications": notifications,
	// })
}

// GetPendingNotifications handles GET /api/notifications/pending
func (h *NotificationHandler) GetPendingNotifications(ctx *kp.Ctx) {
	ctx.L("get_pending_notifications")
	var customError *kp.Error
	// if r.Method != http.MethodGet {
	// 	response.Error(w, &errors.AppError{
	// 		StatusCode: http.StatusMethodNotAllowed,
	// 		Message:    "Method not allowed",
	// 	})
	// 	return
	// }

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		// response.Error(w, &errors.AppError{
		// 	StatusCode: http.StatusUnauthorized,
		// 	Message:    "Unauthorized",
		// })
		customError := &kp.Error{
			StatusCode: http.StatusUnauthorized,
			Message:    "unauthorized",
			Err:        errors.New("Unauthorized"),
		}
		ctx.JSONError(customError)
		return
	}
	notifications, err := h.notificationService.GetPendingNotifications(ctx, userID)
	if err != nil {
		// response.Error(w, &errors.AppError{
		// 	StatusCode: http.StatusInternalServerError,
		// 	Message:    "Failed to get pending notifications",
		// })
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

	// response.Success(w, &response.SuccessResponse{
	// 	Data:    notifications,
	// 	Success: true,
	// }, "Pending notifications retrieved successfully")
	ctx.JSON(http.StatusOK, &response.SuccessResponse{
		Data:    notifications,
		Success: true,
		Message: "Pending notifications retrieved successfully",
	})

}

// GetUnreadCount handles GET /api/notifications/unread-count
func (h *NotificationHandler) GetUnreadCount(ctx *kp.Ctx) {
	// if r.Method != http.MethodGet {
	// 	response.Error(w, &errors.AppError{
	// 		StatusCode: http.StatusMethodNotAllowed,
	// 		Message:    "Method not allowed",
	// 	})
	// 	return
	// }
	ctx.L("get_unread_count")
	var customError *kp.Error

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		// response.Error(w, &errors.AppError{
		// 	StatusCode: http.StatusUnauthorized,
		// 	Message:    "Unauthorized",
		// })
		customError = &kp.Error{
			StatusCode: http.StatusUnauthorized,
			Message:    "unauthorized",
			Err:        errors.New("Unauthorized"),
		}
		ctx.JSONError(customError)
		return
	}
	count, err := h.notificationService.GetUnreadCount(ctx, userID)
	if err != nil {
		// response.Error(w, &errors.AppError{
		// 	StatusCode: http.StatusInternalServerError,
		// 	Message:    "Failed to get unread count",
		// })
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

	ctx.JSON(http.StatusOK, &response.SuccessResponse{
		Data:    map[string]int64{"count": count},
		Success: true, Message: "Unread count retrieved successfully",
	})
}

// MarkAsRead handles POST /api/notifications/mark-read
func (h *NotificationHandler) MarkAsRead(ctx *kp.Ctx) {
	ctx.L("mark_as_read")
	var customError *kp.Error
	// if r.Method != http.MethodPost {
	// 	response.Error(w, &errors.AppError{
	// 		StatusCode: http.StatusMethodNotAllowed,
	// 		Message:    "Method not allowed",
	// 	})
	// 	return
	// }

	var req struct {
		NotificationID string `json:"notificationId"`
	}

	if err := ctx.Bind(&req); err != nil {
		customError = &kp.Error{
			StatusCode: http.StatusBadRequest,
			Message:    "bad_request",
			Err:        err,
		}
		ctx.JSONError(customError)
		return
	}
	// if err := response.DecodeJSON(r, &req); err != nil {
	// 	response.Error(w, &errors.AppError{
	// 		StatusCode: http.StatusBadRequest,
	// 		Message:    "Invalid request body",
	// 	})
	// 	return
	// }

	if req.NotificationID == "" {
		customError = &kp.Error{
			StatusCode: http.StatusBadRequest,
			Message:    "notification_id_required",
			Err:        errors.New("Notification ID is required"),
		}
		ctx.JSONError(customError)
		return
	}

	if err := h.notificationService.MarkAsRead(ctx, req.NotificationID); err != nil {
		customError = &kp.Error{
			StatusCode: http.StatusInternalServerError,
			Message:    "internal_server",
			Err:        err,
		}
		ctx.JSONError(customError)
		return
	}

	// response.Success(w, map[string]string{"message": "Notification marked as read"}, "Notification marked as read successfully")
	ctx.JSON(http.StatusOK, &response.SuccessResponse{
		Data:    map[string]string{"message": "Notification marked as read"},
		Success: true,
		Message: "Notification marked as read successfully",
	})
}

// MarkAllAsRead handles POST /api/notifications/mark-all-read
func (h *NotificationHandler) MarkAllAsRead(ctx *kp.Ctx) {
	ctx.L("mark_all_as_read")
	var customError *kp.Error
	// if r.Method != http.MethodPost {
	// 	response.Error(w, &errors.AppError{
	// 		StatusCode: http.StatusMethodNotAllowed,
	// 		Message:    "Method not allowed",
	// 	})
	// 	return
	// }

	userID, ok := middleware.GetUserID(ctx)
	if !ok {
		// response.Error(w, &errors.AppError{
		// 	StatusCode: http.StatusUnauthorized,
		// 	Message:    "Unauthorized",
		// })
		customError = &kp.Error{
			StatusCode: http.StatusUnauthorized,
			Message:    "unauthorized",
			Err:        errors.New("Unauthorized"),
		}
		ctx.JSONError(customError)
		return
	}
	if err := h.notificationService.MarkAllAsRead(ctx, userID); err != nil {
		customError = &kp.Error{
			StatusCode: http.StatusInternalServerError,
			Message:    "internal_server",
			Err:        err,
		}
		ctx.JSONError(customError)
		return
	}

	// ctx.JSON(http.StatusOK, map[string]string{"message": "All notifications marked as read"})
	ctx.JSON(http.StatusOK, &response.SuccessResponse{
		Data:    map[string]string{"message": "All notifications marked as read"},
		Success: true,
		Message: "All notifications marked as read successfully",
	})
}
