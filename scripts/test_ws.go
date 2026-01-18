package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/gorilla/websocket"
)

const (
	baseURL    = "http://localhost:8080"
	wsURL      = "ws://localhost:8080"
	identifier = "chattest1"
	password   = "Test1234567!"
)

type LoginResponse struct {
	Success bool `json:"success"`
	Data    struct {
		Token string `json:"token"`
		User  struct {
			ID       string `json:"id"`
			Username string `json:"username"`
		} `json:"user"`
	} `json:"data"`
	Message string `json:"message"`
}

type WSMessage struct {
	Type    string      `json:"type"`
	RoomID  string      `json:"roomId,omitempty"`
	Payload interface{} `json:"payload,omitempty"`
}

func main() {
	// Step 1: Login to get token
	token, userID, err := login(identifier, password)
	if err != nil {
		log.Fatalf("Login failed: %v", err)
	}
	fmt.Printf("✅ Login successful! UserID: %s\n", userID)

	// Step 2: Connect to WebSocket
	conn, err := connectWebSocket(token)
	if err != nil {
		log.Fatalf("WebSocket connection failed: %v", err)
	}
	defer conn.Close()
	fmt.Println("✅ WebSocket connected!")

	// Step 3: Handle interrupt signal for graceful shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	// Step 4: Read messages in goroutine
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("❌ Read error: %v", err)
				}
				return
			}
			fmt.Printf("📨 Received: %s\n", message)
		}
	}()

	// Step 5: Send heartbeat periodically
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	fmt.Println("🔄 Listening for messages... (Press Ctrl+C to exit)")

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			// Send heartbeat
			msg := WSMessage{Type: "heartbeat", Payload: map[string]interface{}{}}
			if err := conn.WriteJSON(msg); err != nil {
				log.Printf("❌ Heartbeat error: %v", err)
				return
			}
			fmt.Println("💓 Heartbeat sent")
		case <-interrupt:
			fmt.Println("\n🛑 Shutting down...")

			// Send close message
			err := conn.WriteMessage(websocket.CloseMessage,
				websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Printf("Close error: %v", err)
			}

			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func login(identifier, password string) (token, userID string, err error) {
	reqBody := fmt.Sprintf(`{"identifier": "%s", "password": "%s"}`, identifier, password)

	resp, err := http.Post(
		baseURL+"/api/auth/login",
		"application/json",
		strings.NewReader(reqBody),
	)
	if err != nil {
		return "", "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("login failed with status: %d", resp.StatusCode)
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", "", fmt.Errorf("decode response failed: %w", err)
	}

	if !loginResp.Success {
		return "", "", fmt.Errorf("login failed: %s", loginResp.Message)
	}

	return loginResp.Data.Token, loginResp.Data.User.ID, nil
}

func connectWebSocket(token string) (*websocket.Conn, error) {
	u, err := url.Parse(wsURL + "/ws")
	if err != nil {
		return nil, fmt.Errorf("parse URL failed: %w", err)
	}

	q := u.Query()
	q.Set("token", token)
	u.RawQuery = q.Encode()

	fmt.Printf("🔌 Connecting to %s\n", u.String())

	conn, resp, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		if resp != nil {
			return nil, fmt.Errorf("dial failed with status %d: %w", resp.StatusCode, err)
		}
		return nil, fmt.Errorf("dial failed: %w", err)
	}

	return conn, nil
}

// SendMessage sends a chat message to a room
func SendMessage(conn *websocket.Conn, roomID, content string) error {
	msg := WSMessage{
		Type:   "message",
		RoomID: roomID,
		Payload: map[string]interface{}{
			"content": content,
			"tempId":  fmt.Sprintf("temp-%d", time.Now().UnixNano()),
		},
	}
	return conn.WriteJSON(msg)
}

// SendTyping sends typing indicator
func SendTyping(conn *websocket.Conn, roomID string, isTyping bool) error {
	msg := WSMessage{
		Type:   "typing",
		RoomID: roomID,
		Payload: map[string]interface{}{
			"isTyping": isTyping,
		},
	}
	return conn.WriteJSON(msg)
}

// JoinRoom subscribes to a room
func JoinRoom(conn *websocket.Conn, roomID string) error {
	msg := WSMessage{
		Type:   "join_room",
		RoomID: roomID,
	}
	return conn.WriteJSON(msg)
}
