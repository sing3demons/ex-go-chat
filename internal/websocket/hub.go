package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"realtime-chat-system/pkg/logAction"
	"realtime-chat-system/pkg/logger"
	"realtime-chat-system/pkg/mlog"
)

// IncomingMessage represents an incoming WebSocket message
type IncomingMessage struct {
	Connection *Connection
	Message    *WSMessage
}

// RoomBroadcast represents a broadcast to a room
type RoomBroadcast struct {
	RoomID  string
	Message *WSMessage
	Exclude string // Exclude this userID from broadcast
}

// UserBroadcast represents a broadcast to a specific user
type UserBroadcast struct {
	UserID  string
	Message *WSMessage
}

// MessageHandler handles incoming WebSocket messages
type MessageHandler interface {
	HandleMessage(conn *Connection, msg *WSMessage)
}

// Hub maintains the set of active connections and broadcasts messages
type Hub struct {
	// Registered connections (userID -> Connection)
	connections map[string]*Connection
	mu          sync.RWMutex

	// Register requests from connections
	Register chan *Connection

	// Unregister requests from connections
	Unregister chan *Connection

	// Handle incoming messages
	HandleMessage chan *IncomingMessage

	// Broadcast message to specific room
	BroadcastToRoom chan *RoomBroadcast

	// Broadcast message to specific user
	BroadcastToUser chan *UserBroadcast

	// Message handler
	messageHandler MessageHandler

	// Presence service
	presenceService PresenceService

	// Redis broadcaster for multi-server distribution
	broadcaster *RedisBroadcaster

	log *logger.Logger
}

// PresenceService interface for presence tracking
type PresenceService interface {
	SetOnline(ctx context.Context, userID string)
	SetOffline(ctx context.Context, userID string)
}

// NewHub creates a new Hub
func NewHub(presenceService PresenceService, log *logger.Logger) *Hub {
	return &Hub{
		connections:     make(map[string]*Connection),
		Register:        make(chan *Connection),
		Unregister:      make(chan *Connection),
		HandleMessage:   make(chan *IncomingMessage, 256),
		BroadcastToRoom: make(chan *RoomBroadcast, 256),
		BroadcastToUser: make(chan *UserBroadcast, 256),
		messageHandler:  nil,
		presenceService: presenceService,
		broadcaster:     nil,
		log:             log,
	}
}

// SetRedisBroadcaster sets the Redis broadcaster for multi-server distribution
func (h *Hub) SetRedisBroadcaster(broadcaster *RedisBroadcaster) {
	h.broadcaster = broadcaster
	if broadcaster != nil {
		h.log.Infof("Redis Pub/Sub enabled for multi-server broadcasting (ServerID: %s)", broadcaster.GetServerID())
	} else {
		h.log.Info("Redis Pub/Sub disabled - running in single-server mode")
	}
}

// SetMessageHandler sets the message handler
func (h *Hub) SetMessageHandler(handler MessageHandler) {
	h.messageHandler = handler
}

// SetPresenceService sets the presence service
func (h *Hub) SetPresenceService(service PresenceService) {
	h.presenceService = service
}

// Run starts the hub
func (h *Hub) Run() {
	// Set up broadcaster channel (will be nil if broadcaster not initialized, blocking forever)
	var broadcasterChan <-chan *DistributedBroadcast
	if h.broadcaster != nil {
		broadcasterChan = h.broadcaster.IncomingBroadcast
	}

	for {
		select {
		case conn := <-h.Register:
			h.registerConnection(conn)

		case conn := <-h.Unregister:
			h.unregisterConnection(conn)

		case incoming := <-h.HandleMessage:
			h.messageHandler.HandleMessage(incoming.Connection, incoming.Message)

		case broadcast := <-h.BroadcastToRoom:
			h.broadcastToRoom(broadcast)

		case broadcast := <-h.BroadcastToUser:
			h.broadcastToUser(broadcast)

		case distributed := <-broadcasterChan:
			if distributed != nil {
				h.handleDistributedBroadcast(distributed)
			}
		}
	}
}

// registerConnection registers a new connection
func (h *Hub) registerConnection(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	// Close existing connection if any
	if existingConn, exists := h.connections[conn.UserID]; exists {
		// Gracefully cancel and close the previous connection to avoid leaks
		existingConn.Close()
	}

	h.connections[conn.UserID] = conn
	h.log.Infof("User %s connected (total: %d)", conn.UserID, len(h.connections))
	mlog.L(conn.ctx).Debug(logAction.APP_LOGIC(fmt.Sprintf("User %s connected (total: %d)", conn.UserID, len(h.connections))), "User "+conn.UserID+" connected")

	// Subscribe to user's personal Redis channel (if broadcaster enabled)
	if h.broadcaster != nil {
		ctx := context.Background()
		if err := h.broadcaster.SubscribeToUserBroadcasts(ctx, conn.UserID); err != nil {
			h.log.Errorf("Failed to subscribe to Redis user channel %s: %v", conn.UserID, err)
		} else {
			h.log.Debugf("Subscribed to Redis user channel: %s", conn.UserID)
		}
	}

	// Broadcast presence update to user's rooms
	h.broadcastPresenceUpdate(conn.UserID, true)
}

// unregisterConnection unregisters a connection
func (h *Hub) unregisterConnection(conn *Connection) {
	h.mu.Lock()
	defer h.mu.Unlock()

	if existingConn, exists := h.connections[conn.UserID]; exists && existingConn == conn {
		log := mlog.L(conn.ctx)
		delete(h.connections, conn.UserID)

		// Ensure the websocket and context are closed
		conn.Close()

		// Safely close the channel
		select {
		case <-conn.Send:
			// Channel already closed
		default:
			close(conn.Send)
		}

		h.log.Infof("User %s disconnected (total: %d)", conn.UserID, len(h.connections))
		log.Debug(logAction.APP_LOGIC(fmt.Sprintf("User %s disconnected (total: %d)", conn.UserID, len(h.connections))), "User "+conn.UserID+" disconnected")

		// Unsubscribe from Redis user channel (if broadcaster enabled)
		if h.broadcaster != nil {
			if err := h.broadcaster.UnsubscribeFromUser(conn.UserID); err != nil {
				h.log.Errorf("Failed to unsubscribe from Redis user channel %s: %v", conn.UserID, err)
			} else {
				h.log.Debugf("Unsubscribed from Redis user channel: %s", conn.UserID)
			}
		}

		// Mark user as offline
		if h.presenceService != nil {
			h.presenceService.SetOffline(conn.ctx, conn.UserID)
			// flush log with offline status
		}

		// Broadcast presence update to user's rooms
		h.broadcastPresenceUpdate(conn.UserID, false)
	}
}

// broadcastToRoom broadcasts a message to all connections in a room
func (h *Hub) broadcastToRoom(broadcast *RoomBroadcast) {
	ctx := context.Background()

	// Publish to Redis for multi-server distribution
	if h.broadcaster != nil {
		if err := h.broadcaster.PublishToRoom(ctx, broadcast.RoomID, broadcast.Message, broadcast.Exclude); err != nil {
			h.log.Errorf("Failed to publish to Redis room: %v", err)
		}
	}

	// Also deliver locally to this server's connections
	h.mu.RLock()
	defer h.mu.RUnlock()

	h.log.Infof("Broadcasting to room %s, total connections: %d, exclude: %s", broadcast.RoomID, len(h.connections), broadcast.Exclude)

	deliveredCount := 0
	for userID, conn := range h.connections {
		// Skip excluded user
		if userID == broadcast.Exclude {
			h.log.Infof("Skipping excluded user: %s", userID)
			continue
		}

		// Check if connection is subscribed to the room
		isSubscribed := conn.IsSubscribedToRoom(broadcast.RoomID)
		h.log.Infof("User %s subscribed to room %s: %v, rooms: %v", userID, broadcast.RoomID, isSubscribed, conn.GetRooms())

		if isSubscribed {
			conn.SendMessage(broadcast.Message)
			deliveredCount++
			h.log.Infof("Message sent to user %s in room %s", userID, broadcast.RoomID)
		}
	}

	h.log.Infof("Broadcast complete: delivered to %d users in room %s", deliveredCount, broadcast.RoomID)
}

// SubscribeConnectionToRoom subscribes a connection to a room and coordinates Redis subscription
func (h *Hub) SubscribeConnectionToRoom(conn *Connection, roomID string) {
	// Subscribe connection locally
	conn.SubscribeToRoom(roomID)

	// Subscribe to Redis channel for distributed broadcasts (if broadcaster enabled)
	if h.broadcaster != nil {
		ctx := context.Background()
		if err := h.broadcaster.SubscribeToRoomBroadcasts(ctx, roomID); err != nil {
			h.log.Errorf("Failed to subscribe to Redis room channel %s: %v", roomID, err)
		} else {
			h.log.Debugf("Subscribed to Redis room channel: %s", roomID)
		}
	}
}

// UnsubscribeConnectionFromRoom unsubscribes a connection from a room
func (h *Hub) UnsubscribeConnectionFromRoom(conn *Connection, roomID string) {
	conn.UnsubscribeFromRoom(roomID)

	// Note: We don't unsubscribe from Redis channel here because other connections
	// on this server might still be subscribed to the same room
}

// broadcastToUser broadcasts a message to a specific user
func (h *Hub) broadcastToUser(broadcast *UserBroadcast) {
	ctx := context.Background()

	// Publish to Redis for multi-server distribution
	if h.broadcaster != nil {
		if err := h.broadcaster.PublishToUser(ctx, broadcast.UserID, broadcast.Message); err != nil {
			h.log.Errorf("Failed to publish to Redis user: %v", err)
		}
	}

	// Also try to deliver locally to this server's connection
	h.mu.RLock()
	defer h.mu.RUnlock()

	if conn, exists := h.connections[broadcast.UserID]; exists {
		conn.SendMessage(broadcast.Message)
	}
}

// GetConnection returns a connection by user ID
func (h *Hub) GetConnection(userID string) (*Connection, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()
	conn, exists := h.connections[userID]
	return conn, exists
}

// IsUserOnline checks if a user is online
func (h *Hub) IsUserOnline(userID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, exists := h.connections[userID]
	return exists
}

// GetOnlineUsers returns all online user IDs
func (h *Hub) GetOnlineUsers() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()

	users := make([]string, 0, len(h.connections))
	for userID := range h.connections {
		users = append(users, userID)
	}
	return users
}

// GetConnectionCount returns the number of active connections
func (h *Hub) GetConnectionCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.connections)
}

// broadcastPresenceUpdate broadcasts presence update to user's rooms
func (h *Hub) broadcastPresenceUpdate(userID string, online bool) {
	// Note: This is called with lock held, so we need to be careful
	// We'll send to all connections and let them filter by room
	payload := PresencePayload{
		UserID: userID,
		Online: online,
	}

	if !online {
		payload.LastSeen = time.Now().Format(time.RFC3339)
	}

	payloadBytes, _ := json.Marshal(payload)

	msg := &WSMessage{
		Type:    MessageTypePresence,
		Payload: payloadBytes,
	}

	// Broadcast to all connections (they'll filter by shared rooms)
	for _, conn := range h.connections {
		if conn.UserID != userID {
			conn.SendMessage(msg)
		}
	}
}

// NotifyRoomCreated notifies users about a new room and subscribes them to it
func (h *Hub) NotifyRoomCreated(roomID, roomType, name string, members []string) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	payload := RoomCreatedPayload{
		RoomID:   roomID,
		RoomType: roomType,
		Name:     name,
		Members:  members,
	}

	payloadBytes, _ := json.Marshal(payload)

	msg := &WSMessage{
		Type:    MessageTypeRoomCreated,
		RoomID:  roomID,
		Payload: payloadBytes,
	}

	// Notify all members and subscribe them to the room
	for _, memberID := range members {
		if conn, exists := h.connections[memberID]; exists {
			// Subscribe connection to the new room (including Redis)
			h.SubscribeConnectionToRoom(conn, roomID)
			// Send notification
			conn.SendMessage(msg)
			h.log.Infof("User %s subscribed to new room %s", memberID, roomID)
		}
	}
}

// handleDistributedBroadcast handles incoming broadcasts from Redis (other servers)
func (h *Hub) handleDistributedBroadcast(broadcast *DistributedBroadcast) {
	// Avoid re-processing messages from this server
	if broadcast.ServerID == h.broadcaster.GetServerID() {
		return
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	if broadcast.Type == "room" {
		h.log.Infof("Received distributed broadcast for room %s from server %s", broadcast.RoomID, broadcast.ServerID)

		deliveredCount := 0
		for userID, conn := range h.connections {
			// Skip excluded user
			if userID == broadcast.Exclude {
				continue
			}

			// Check if connection is subscribed to the room
			if conn.IsSubscribedToRoom(broadcast.RoomID) {
				conn.SendMessage(broadcast.Message)
				deliveredCount++
			}
		}

		h.log.Infof("Distributed broadcast to room %s delivered to %d local connections", broadcast.RoomID, deliveredCount)
	} else if broadcast.Type == "user" {
		h.log.Infof("Received distributed broadcast for user %s from server %s", broadcast.UserID, broadcast.ServerID)

		if conn, exists := h.connections[broadcast.UserID]; exists {
			conn.SendMessage(broadcast.Message)
			h.log.Infof("Distributed broadcast to user %s delivered", broadcast.UserID)
		}
	}
}
