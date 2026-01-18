package websocket

import (
	"net/http"

	"realtime-chat-system/internal/service"
	"realtime-chat-system/pkg/kp"
	"realtime-chat-system/pkg/logger"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// TODO: In production, check origin properly
		return true
	},
}

// Handler handles WebSocket connections
type Handler struct {
	hub             *Hub
	authService     service.AuthService
	roomService     service.RoomService
	presenceService service.PresenceService
	log             *logger.Logger
}

// NewHandler creates a new WebSocket handler
func NewHandler(hub *Hub, authService service.AuthService, roomService service.RoomService, presenceService service.PresenceService, log *logger.Logger) *Handler {
	return &Handler{
		hub:             hub,
		authService:     authService,
		roomService:     roomService,
		presenceService: presenceService,
		log:             log,
	}
}

// ServeWS handles WebSocket requests
func (h *Handler) ServeWS(ctx *kp.Ctx) {
	// Get token from query parameter
	// token := ctx.Request().URL.Query().Get("token")
	token := ctx.Query("token")
	if token == "" {
		h.log.Error("WebSocket connection rejected: missing token")
		ctx.JSON(http.StatusUnauthorized, "Missing token")
		return
	}

	// Validate token
	claims, err := h.authService.ValidateToken(ctx, token)
	if err != nil {
		h.log.Errorf("WebSocket connection rejected: invalid token - %v", err)
		ctx.JSON(http.StatusUnauthorized, "Invalid token")
		return
	}

	// Upgrade connection
	conn, err := upgrader.Upgrade(ctx.Res, ctx.Req, nil)
	if err != nil {
		h.log.Errorf("Failed to upgrade WebSocket connection: %v", err)
		return
	}

	// Create connection
	wsConn := NewConnection(claims.UserID, claims.Username, conn, h.log)

	// Subscribe to user's rooms
	rooms, err := h.roomService.GetUserRooms(ctx, claims.UserID)
	if err != nil {
		h.log.Errorf("Failed to get user rooms: %v", err)
	} else {
		for _, room := range rooms {
			wsConn.SubscribeToRoom(room.ID)
		}
		h.log.Infof("User %s subscribed to %d rooms", claims.UserID, len(rooms))
	}

	// Register connection
	h.hub.Register <- wsConn

	// Mark user as online
	h.presenceService.SetOnline(ctx, claims.UserID)

	// Start write pump in goroutine
	go wsConn.WritePump()

	// Start read pump (blocking - keeps connection alive)
	wsConn.ReadPump(h.hub)

	h.log.Infof("WebSocket connection closed for user %s", claims.UserID)
}
