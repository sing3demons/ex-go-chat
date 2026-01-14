# Checkpoint 2: WebSocket System Complete

## Date
2026-01-14

## Status
✅ Complete

## Completed Tasks

### Task 14: WebSocket Connection Management
- ✅ Created WebSocket connection pool (Hub)
- ✅ Implemented Connection with ReadPump/WritePump
- ✅ Implemented thread-safe connection management
- ✅ Added room subscription mechanism
- ✅ Implemented heartbeat (54s ping interval)
- ✅ Created WebSocket handler with JWT authentication

### Task 15: Presence Service
- ✅ Created PresenceService with in-memory store
- ✅ Implemented SetOnline/SetOffline methods
- ✅ Implemented IsOnline/GetLastSeen methods
- ✅ Implemented heartbeat monitoring (60s stale threshold)
- ✅ Integrated presence tracking with WebSocket lifecycle
- ✅ Broadcast presence updates on connect/disconnect

### Task 16: WebSocket Message Handling
- ✅ Created WSMessageHandler for routing messages
- ✅ Implemented chat message handling (save + broadcast)
- ✅ Implemented typing indicator handling (broadcast only, no persist)
- ✅ Implemented message status updates (delivered/read)
- ✅ Implemented message edit handling
- ✅ Implemented message delete handling
- ✅ Implemented heartbeat handling
- ✅ Added error handling and error messages

## Architecture

### WebSocket Flow
```
Client → /ws?token=<jwt>
  ↓
Handler validates JWT
  ↓
Create Connection
  ↓
Subscribe to user's rooms
  ↓
Register with Hub
  ↓
Mark user online (PresenceService)
  ↓
Start ReadPump & WritePump
  ↓
Handle incoming messages via WSMessageHandler
```

### Message Types
- `message`: Chat messages (saved to DB + broadcast)
- `typing`: Typing indicators (broadcast only)
- `read`: Read status updates
- `delivered`: Delivery status updates
- `presence`: Online/offline status
- `edit`: Message edits
- `delete`: Message deletes
- `heartbeat`: Keep-alive
- `error`: Error messages

### Components

#### Hub (`internal/websocket/hub.go`)
- Maintains connection pool (userID → Connection)
- Handles Register/Unregister
- Routes incoming messages to WSMessageHandler
- Broadcasts messages to rooms or specific users
- Broadcasts presence updates
- Thread-safe with RWMutex

#### Connection (`internal/websocket/connection.go`)
- Represents a single WebSocket connection
- ReadPump: Reads messages from client
- WritePump: Writes messages to client
- Room subscriptions
- Heartbeat (54s ping)

#### Handler (`internal/websocket/handler.go`)
- HTTP handler for WebSocket upgrade
- JWT authentication from query parameter
- Auto-subscribes to user's rooms
- Marks user online on connect

#### WSMessageHandler (`internal/websocket/message_handler.go`)
- Routes messages by type
- Handles chat messages (verify membership, save, broadcast)
- Handles typing indicators (broadcast to room)
- Handles status updates (update DB, broadcast to sender)
- Handles edit/delete (verify ownership, update DB, broadcast)
- Handles heartbeat (update presence)

#### PresenceService (`internal/service/presence_service.go`)
- In-memory presence tracking
- SetOnline/SetOffline
- IsOnline/GetLastSeen
- Heartbeat monitoring (marks stale connections offline after 60s)
- Background goroutine for monitoring

## API Endpoints

### WebSocket
- `GET /ws?token=<jwt>` - Establish WebSocket connection

## Message Formats

### Incoming Message (Client → Server)
```json
{
  "type": "message|typing|read|delivered|edit|delete|heartbeat",
  "roomId": "room_id",
  "payload": { ... }
}
```

### Chat Message Payload
```json
{
  "content": "Hello world"
}
```

### Typing Indicator Payload
```json
{
  "isTyping": true
}
```

### Status Update Payload
```json
{
  "messageId": "msg_id"
}
```

### Edit Payload
```json
{
  "messageId": "msg_id",
  "content": "Updated content"
}
```

### Delete Payload
```json
{
  "messageId": "msg_id"
}
```

### Outgoing Message (Server → Client)
```json
{
  "type": "message|typing|read|delivered|presence|edit|delete|error",
  "roomId": "room_id",
  "payload": { ... }
}
```

### Chat Message Broadcast
```json
{
  "messageId": "msg_id",
  "content": "Hello world",
  "senderId": "user_id",
  "timestamp": "2026-01-14T10:00:00Z"
}
```

### Typing Indicator Broadcast
```json
{
  "userId": "user_id",
  "username": "john_doe",
  "isTyping": true
}
```

### Status Update Broadcast
```json
{
  "messageId": "msg_id",
  "userId": "user_id",
  "status": "delivered|read",
  "timestamp": "2026-01-14T10:00:00Z"
}
```

### Presence Update Broadcast
```json
{
  "userId": "user_id",
  "online": true,
  "lastSeen": "2026-01-14T10:00:00Z"
}
```

### Error Message
```json
{
  "code": "UNAUTHORIZED",
  "message": "You are not a member of this room"
}
```

## Testing

### Manual Testing
1. Start server: `make run`
2. Connect via WebSocket client (e.g., wscat)
3. Authenticate: `wscat -c "ws://localhost:8080/ws?token=<jwt>"`
4. Send messages, typing indicators, status updates
5. Verify broadcasts to other connected clients

### Test Scenarios
- ✅ Connect with valid JWT
- ✅ Connect with invalid JWT (should reject)
- ✅ Send message to room (should save + broadcast)
- ✅ Send message to non-member room (should reject)
- ✅ Send typing indicator (should broadcast to room)
- ✅ Update read status (should update DB + broadcast to sender)
- ✅ Edit message (should update DB + broadcast)
- ✅ Delete message (should update DB + broadcast)
- ✅ Heartbeat (should update presence)
- ✅ Disconnect (should mark offline + broadcast presence)

## Build Status
✅ Build successful: `go build -o bin/server ./cmd/server`

## Next Steps
- Task 18: Implement Notification System
- Task 19: Implement Authorization and Error Handling
- Task 20: Final Integration and Testing
- Task 21: Final Checkpoint

## Notes
- WebSocket authentication uses query parameter for JWT token
- Presence tracking is in-memory (will reset on server restart)
- Heartbeat monitoring runs every 30s, marks connections stale after 60s
- All broadcasts are non-blocking (buffered channels)
- Connection pool is thread-safe with RWMutex
- Message handler validates room membership before processing
- Typing indicators are not persisted to database
- Status updates are sent only to message sender
- Presence updates are broadcast to all connections (clients filter by shared rooms)

## Files Modified/Created
- `internal/service/presence_service.go` (new)
- `internal/websocket/message_handler.go` (new)
- `internal/websocket/hub.go` (modified - added presence integration)
- `internal/websocket/handler.go` (modified - added presence service)
- `cmd/server/main.go` (modified - wired WebSocket components)
- `docs/CHECKPOINT_2_WEBSOCKET.md` (new)
