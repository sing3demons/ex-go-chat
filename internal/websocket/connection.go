package websocket

import (
	"encoding/json"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"realtime-chat-system/pkg/logger"
)

// Connection represents a WebSocket connection
type Connection struct {
	UserID   string
	Username string
	Conn     *websocket.Conn
	Send     chan []byte
	Rooms    map[string]bool // roomID -> subscribed
	mu       sync.RWMutex
	log      *logger.Logger
}

// NewConnection creates a new WebSocket connection
func NewConnection(userID, username string, conn *websocket.Conn, log *logger.Logger) *Connection {
	return &Connection{
		UserID:   userID,
		Username: username,
		Conn:     conn,
		Send:     make(chan []byte, 256),
		Rooms:    make(map[string]bool),
		log:      log,
	}
}

// ReadPump pumps messages from the WebSocket connection to the hub
func (c *Connection) ReadPump(hub *Hub) {
	defer func() {
		hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				c.log.Errorf("WebSocket error: %v", err)
			}
			break
		}

		// Parse message
		var wsMsg WSMessage
		if err := json.Unmarshal(message, &wsMsg); err != nil {
			c.log.Errorf("Failed to parse WebSocket message: %v", err)
			c.SendError("INVALID_MESSAGE", "Invalid message format")
			continue
		}

		// Handle message
		hub.HandleMessage <- &IncomingMessage{
			Connection: c,
			Message:    &wsMsg,
		}
	}
}

// WritePump pumps messages from the hub to the WebSocket connection
func (c *Connection) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				// Hub closed the channel
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued messages to the current WebSocket message
			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// SubscribeToRoom subscribes the connection to a room
func (c *Connection) SubscribeToRoom(roomID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Rooms[roomID] = true
}

// UnsubscribeFromRoom unsubscribes the connection from a room
func (c *Connection) UnsubscribeFromRoom(roomID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.Rooms, roomID)
}

// IsSubscribedToRoom checks if the connection is subscribed to a room
func (c *Connection) IsSubscribedToRoom(roomID string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.Rooms[roomID]
}

// GetRooms returns all rooms the connection is subscribed to
func (c *Connection) GetRooms() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	
	rooms := make([]string, 0, len(c.Rooms))
	for roomID := range c.Rooms {
		rooms = append(rooms, roomID)
	}
	return rooms
}

// SendMessage sends a message to the connection
func (c *Connection) SendMessage(msg *WSMessage) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}

	select {
	case c.Send <- data:
		return nil
	default:
		return websocket.ErrCloseSent
	}
}

// SendError sends an error message to the connection
func (c *Connection) SendError(code, message string) {
	payload, _ := json.Marshal(ErrorPayload{
		Code:    code,
		Message: message,
	})

	c.SendMessage(&WSMessage{
		Type:    MessageTypeError,
		Payload: payload,
	})
}
