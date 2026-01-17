# Technical Architecture Documentation

## 🏗️ System Architecture

### **High-Level Architecture**
```
┌─────────────────────────────────────────────────────────────────┐
│                        Client Layer                             │
├─────────────────────────────────────────────────────────────────┤
│  React App (TypeScript)                                        │
│  ├── Components (UI)                                           │
│  ├── Stores (Zustand)                                          │
│  ├── Services (API, WebSocket)                                 │
│  └── Hooks (Custom logic)                                      │
└─────────────────────────────────────────────────────────────────┘
                                │
                        HTTP/WebSocket
                                │
┌─────────────────────────────────────────────────────────────────┐
│                      API Gateway Layer                         │
├─────────────────────────────────────────────────────────────────┤
│  Go HTTP Server                                                │
│  ├── CORS Middleware                                           │
│  ├── Auth Middleware                                           │
│  └── Request/Response Handling                                 │
└─────────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────────┐
│                     Application Layer                          │
├─────────────────────────────────────────────────────────────────┤
│  Handlers (HTTP/WebSocket)                                     │
│  ├── AuthHandler                                               │
│  ├── MessageHandler                                            │
│  ├── RoomHandler                                               │
│  ├── UserHandler                                               │
│  └── WebSocket Hub                                             │
└─────────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────────┐
│                      Business Layer                            │
├─────────────────────────────────────────────────────────────────┤
│  Services (Business Logic)                                     │
│  ├── AuthService                                               │
│  ├── MessageService                                            │
│  ├── RoomService                                               │
│  ├── PresenceService                                           │
│  └── NotificationService                                       │
└─────────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────────┐
│                      Data Access Layer                         │
├─────────────────────────────────────────────────────────────────┤
│  Repositories (Data Access)                                    │
│  ├── UserRepository                                            │
│  ├── MessageRepository                                         │
│  ├── RoomRepository                                            │
│  ├── PresenceRepository                                        │
│  ├── MessageCacheRepository                                    │
│  ├── TypingRepository                                          │
│  ├── SessionRepository                                         │
│  └── RateLimitRepository                                       │
└─────────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────────┐
│                       Storage Layer                            │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────────┐              ┌─────────────────┐          │
│  │    MongoDB      │              │     Redis       │          │
│  │                 │              │                 │          │
│  │ • Users         │              │ • Presence      │          │
│  │ • Messages      │              │ • Message Cache │          │
│  │ • Rooms         │              │ • Typing        │          │
│  │ • Notifications │              │ • Sessions      │          │
│  └─────────────────┘              └─────────────────┘          │
└─────────────────────────────────────────────────────────────────┘
```

## 🔄 Data Flow Architecture

### **1. Message Flow**
```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   User A    │    │  WebSocket  │    │   Service   │    │  Database   │
│   Types     │───►│    Hub      │───►│   Layer     │───►│   Layer     │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                           │                   │                   │
                           │                   ▼                   ▼
                           │            MessageService      MongoDB (save)
                           │                   │                   │
                           │                   ▼                   ▼
                           │         MessageCacheRepo        Redis (cache)
                           │                   │
                           ▼                   ▼
                    Broadcast to Room    Return to Hub
                           │
                           ▼
                   ┌─────────────┐
                   │   User B    │
                   │  Receives   │
                   └─────────────┘
```

### **2. Authentication Flow**
```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Client    │    │   Handler   │    │   Service   │    │ Repository  │
│   Login     │───►│ AuthHandler │───►│ AuthService │───►│ UserRepo    │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
       ▲                   │                   │                   │
       │                   │                   ▼                   ▼
       │                   │            Password Check       MongoDB Query
       │                   │                   │                   │
       │                   │                   ▼                   ▼
       │                   │              JWT Generate        User Data
       │                   │                   │
       │                   ▼                   ▼
       └────────────── JWT Token ◄─────── Return Success
```

### **3. Real-time Presence Flow**
```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│ User Online │    │ WebSocket   │    │ Presence    │    │   Redis     │
│ Connection  │───►│    Hub      │───►│  Service    │───►│   Store     │
└─────────────┘    └─────────────┘    └─────────────┘    └─────────────┘
                           │                   │                   │
                           │                   ▼                   ▼
                           │            Update Status        SET online_users
                           │                   │                   │
                           │                   ▼                   ▼
                           │            Broadcast Event     HASH presence:user
                           │                   │
                           ▼                   ▼
                   All Connected Users   Heartbeat Monitor
                           │                   │
                           ▼                   ▼
                   Update UI Status    Auto-cleanup Stale
```

## 🏛️ Clean Architecture Implementation

### **Dependency Direction**
```
Handler ──► Service ──► Repository ──► Database
   │           │            │
   │           │            └── Interface (abstraction)
   │           └── Business Logic (no external deps)
   └── HTTP/WebSocket handling (framework specific)
```

### **Layer Responsibilities**

#### **1. Handler Layer**
```go
// Responsibilities:
- HTTP request/response handling
- WebSocket connection management
- Input validation
- Authentication middleware
- Error response formatting

// Example:
type MessageHandler struct {
    messageService service.MessageService
    authMiddleware middleware.AuthMiddleware
}
```

#### **2. Service Layer**
```go
// Responsibilities:
- Business logic implementation
- Data validation
- Transaction management
- Cross-service communication
- Error handling

// Example:
type MessageService interface {
    SendMessage(ctx context.Context, roomID, senderID, content string) (*models.Message, error)
    GetMessages(ctx context.Context, roomID string, limit, offset int) ([]*models.Message, error)
}
```

#### **3. Repository Layer**
```go
// Responsibilities:
- Data access abstraction
- Database operations
- Cache management
- Query optimization

// Example:
type MessageRepository interface {
    Create(ctx context.Context, message *models.Message) error
    FindByRoom(ctx context.Context, roomID string, limit, offset int) ([]*models.Message, error)
}
```

## 🔌 WebSocket Architecture

### **Hub Pattern Implementation**
```go
type Hub struct {
    // Connection management
    connections map[string]*Connection  // userID -> Connection
    
    // Channel-based communication
    Register        chan *Connection
    Unregister      chan *Connection
    HandleMessage   chan *IncomingMessage
    BroadcastToRoom chan *RoomBroadcast
    BroadcastToUser chan *UserBroadcast
    
    // Dependencies
    messageHandler  MessageHandler
    presenceService PresenceService
}
```

### **Connection Lifecycle**
```
1. WebSocket Upgrade
   ├── Authentication check
   ├── Create Connection object
   └── Register with Hub

2. Active Connection
   ├── ReadPump (incoming messages)
   ├── WritePump (outgoing messages)
   ├── Heartbeat monitoring
   └── Room subscriptions

3. Connection Cleanup
   ├── Unregister from Hub
   ├── Update presence status
   ├── Close channels
   └── Broadcast offline status
```

### **Message Broadcasting**
```go
// Room-based broadcasting
func (h *Hub) BroadcastToRoom(broadcast *RoomBroadcast) {
    for userID, conn := range h.connections {
        if userID == broadcast.Exclude {
            continue // Skip sender
        }
        
        if conn.IsSubscribedToRoom(broadcast.RoomID) {
            conn.SendMessage(broadcast.Message)
        }
    }
}
```

## 💾 Data Storage Architecture

### **MongoDB Schema Design**

#### **Users Collection**
```javascript
{
  _id: ObjectId,
  username: String (unique),
  email: String (unique),
  password: String (hashed),
  createdAt: Date,
  updatedAt: Date,
  lastSeen: Date
}
```

#### **Rooms Collection**
```javascript
{
  _id: ObjectId,
  name: String (optional for direct messages),
  type: String ("group" | "direct"),
  members: [String], // Array of user IDs
  createdBy: String, // User ID
  createdAt: Date,
  updatedAt: Date
}
```

#### **Messages Collection**
```javascript
{
  _id: ObjectId,
  roomId: String,
  senderId: String,
  content: String,
  createdAt: Date,
  updatedAt: Date,
  editedAt: Date (optional),
  deleted: Boolean,
  status: {
    [userId]: {
      delivered: Boolean,
      deliveredAt: Date,
      read: Boolean,
      readAt: Date
    }
  }
}
```

### **Redis Data Structures**

#### **Presence Management**
```redis
# Online users set
SET online_users: {user1, user2, user3}

# Individual presence data
HASH presence:user1: {
  online: true,
  lastSeen: 1642678800,
  updatedAt: 1642678800
}
```

#### **Message Caching**
```redis
# Individual message cache
STRING message:msg123: "{json_message_data}"

# Room message index (sorted by timestamp)
ZSET room_messages:room456: {
  msg123: 1642678800,
  msg124: 1642678801,
  msg125: 1642678802
}
```

#### **Typing Indicators**
```redis
# Users typing in a room (auto-expire 10s)
SET typing:room456: {user1, user2}
EXPIRE typing:room456 10
```

## 🔄 State Management (Frontend)

### **Zustand Store Architecture**
```typescript
// Store separation by domain
interface AuthStore {
  user: User | null
  token: string | null
  login: (credentials: LoginData) => Promise<void>
  logout: () => void
}

interface ChatStore {
  messages: Record<string, Message[]>  // roomId -> messages
  rooms: Room[]
  activeRoom: string | null
  sendMessage: (roomId: string, content: string) => void
  loadMessages: (roomId: string) => Promise<void>
}

interface PresenceStore {
  onlineUsers: Set<string>
  typingUsers: Record<string, Set<string>>  // roomId -> userIds
  updatePresence: (userId: string, online: boolean) => void
}
```

### **State Synchronization**
```typescript
// WebSocket event handling
websocket.on('message', (data) => {
  chatStore.getState().addMessage(data.roomId, data.message)
})

websocket.on('presence', (data) => {
  presenceStore.getState().updatePresence(data.userId, data.online)
})

websocket.on('typing', (data) => {
  presenceStore.getState().updateTyping(data.roomId, data.userId, data.isTyping)
})
```

## 🚀 Performance Optimizations

### **Backend Optimizations**

#### **1. Redis Caching Strategy**
```go
// Cache-aside pattern
func (s *messageService) GetMessages(ctx context.Context, roomID string, limit, offset int) ([]*models.Message, error) {
    // Try cache first (only for recent messages)
    if offset == 0 {
        if cached, err := s.cacheRepo.GetCachedRoomMessages(ctx, roomID, int64(limit)); err == nil {
            return cached, nil
        }
    }
    
    // Fallback to database
    messages, err := s.messageRepo.FindByRoom(ctx, roomID, limit, offset)
    if err != nil {
        return nil, err
    }
    
    // Cache for next time
    if offset == 0 {
        for _, msg := range messages {
            s.cacheRepo.CacheMessage(ctx, msg)
        }
    }
    
    return messages, nil
}
```

#### **2. Connection Pooling**
```go
// MongoDB connection pool
clientOptions := options.Client().
    ApplyURI(uri).
    SetMaxPoolSize(100).
    SetMinPoolSize(10).
    SetMaxConnIdleTime(30 * time.Second)

// Redis connection pool
redisOptions := &redis.Options{
    Addr:         addr,
    PoolSize:     100,
    MinIdleConns: 10,
    IdleTimeout:  30 * time.Second,
}
```

#### **3. Efficient Broadcasting**
```go
// Avoid unnecessary JSON marshaling
type Hub struct {
    messageCache sync.Map  // Pre-marshaled messages
}

func (h *Hub) BroadcastToRoom(roomID string, message *WSMessage) {
    // Marshal once, send to many
    data, _ := json.Marshal(message)
    
    for _, conn := range h.getRoomConnections(roomID) {
        select {
        case conn.Send <- data:
        default:
            // Non-blocking send, close slow connections
            close(conn.Send)
            delete(h.connections, conn.UserID)
        }
    }
}
```

### **Frontend Optimizations**

#### **1. Virtual Scrolling**
```typescript
// Large message lists
const MessageList = ({ messages }: { messages: Message[] }) => {
  const [visibleRange, setVisibleRange] = useState({ start: 0, end: 50 })
  
  const visibleMessages = useMemo(() => 
    messages.slice(visibleRange.start, visibleRange.end),
    [messages, visibleRange]
  )
  
  return (
    <VirtualizedList
      items={visibleMessages}
      renderItem={MessageItem}
      onRangeChange={setVisibleRange}
    />
  )
}
```

#### **2. Optimistic Updates**
```typescript
const sendMessage = async (content: string) => {
  const tempId = generateTempId()
  
  // Optimistic update
  addOptimisticMessage({
    id: tempId,
    content,
    senderId: currentUser.id,
    status: 'sending'
  })
  
  try {
    const message = await api.sendMessage(roomId, content, tempId)
    // Replace optimistic message with real one
    replaceOptimisticMessage(tempId, message)
  } catch (error) {
    // Mark as failed, allow retry
    markMessageFailed(tempId)
  }
}
```

#### **3. Debounced Typing Indicators**
```typescript
const useTypingIndicator = (roomId: string) => {
  const [isTyping, setIsTyping] = useState(false)
  
  const debouncedStopTyping = useMemo(
    () => debounce(() => {
      setIsTyping(false)
      websocket.send({
        type: 'typing',
        roomId,
        payload: { isTyping: false }
      })
    }, 1000),
    [roomId]
  )
  
  const startTyping = () => {
    if (!isTyping) {
      setIsTyping(true)
      websocket.send({
        type: 'typing',
        roomId,
        payload: { isTyping: true }
      })
    }
    debouncedStopTyping()
  }
  
  return { startTyping }
}
```

## 🔒 Security Architecture

### **Authentication Flow**
```
1. User Login
   ├── Password verification (bcrypt)
   ├── JWT token generation
   └── Secure token storage

2. Request Authentication
   ├── JWT token validation
   ├── Token expiration check
   └── User context injection

3. WebSocket Authentication
   ├── Token validation on upgrade
   ├── Connection authorization
   └── Room access control
```

### **Authorization Layers**
```go
// Middleware chain
HTTP Request → CORS → Auth → Handler → Service → Repository

// WebSocket authorization
WebSocket Upgrade → Token Validation → Room Access Check → Connection Allow
```

### **Data Validation**
```go
// Input validation at multiple layers
type CreateRoomRequest struct {
    Name    string   `json:"name" validate:"required,min=1,max=100"`
    Type    string   `json:"type" validate:"required,oneof=group direct"`
    Members []string `json:"members" validate:"required,min=1,dive,required"`
}

// Sanitization
func sanitizeMessage(content string) string {
    // Remove dangerous HTML/JS
    return html.EscapeString(strings.TrimSpace(content))
}
```

## 📊 Monitoring & Observability

### **Logging Strategy**
```go
// Structured logging
logger.Info("User connected",
    "userId", userID,
    "connectionCount", hub.GetConnectionCount(),
    "timestamp", time.Now(),
)

logger.Error("Message send failed",
    "userId", userID,
    "roomId", roomID,
    "error", err,
    "timestamp", time.Now(),
)
```

### **Health Checks**
```go
// Health check endpoint
func healthCheck(w http.ResponseWriter, r *http.Request) {
    status := map[string]string{
        "status":   "healthy",
        "database": checkMongoDB(),
        "redis":    checkRedis(),
        "version":  version,
    }
    
    json.NewEncoder(w).Encode(status)
}
```

### **Metrics Collection**
```go
// Custom metrics
var (
    activeConnections = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "websocket_active_connections",
        Help: "Number of active WebSocket connections",
    })
    
    messagesProcessed = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "messages_processed_total",
        Help: "Total number of messages processed",
    })
)
```

---

## 🎯 Architecture Benefits

### **Scalability**
- **Horizontal scaling** ready with Redis Pub/Sub
- **Database sharding** support
- **Load balancing** compatible

### **Maintainability**
- **Clean separation** of concerns
- **Testable** architecture
- **Modular** design

### **Performance**
- **Efficient caching** strategy
- **Optimized** database queries
- **Non-blocking** I/O operations

### **Reliability**
- **Error handling** at all layers
- **Graceful degradation**
- **Connection recovery**