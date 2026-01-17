# API Reference Documentation

## 📡 REST API Endpoints

### **Base URL**
```
Development: http://localhost:8080
Production: https://your-domain.com
```

### **Authentication**
All protected endpoints require JWT token in Authorization header:
```
Authorization: Bearer <jwt_token>
```

---

## 🔐 Authentication Endpoints

### **POST /api/auth/register**
Register a new user account.

**Request Body:**
```json
{
  "username": "john_doe",
  "email": "john@example.com",
  "password": "securePassword123"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "507f1f77bcf86cd799439011",
      "username": "john_doe",
      "email": "john@example.com",
      "createdAt": "2024-01-15T10:30:00Z"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Error Responses:**
```json
// 400 Bad Request - Validation error
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Username already exists"
  }
}

// 500 Internal Server Error
{
  "success": false,
  "error": {
    "code": "INTERNAL_ERROR",
    "message": "Failed to create user"
  }
}
```

### **POST /api/auth/login**
Authenticate user and get JWT token.

**Request Body:**
```json
{
  "username": "john_doe",
  "password": "securePassword123"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "507f1f77bcf86cd799439011",
      "username": "john_doe",
      "email": "john@example.com"
    },
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

**Error Responses:**
```json
// 401 Unauthorized
{
  "success": false,
  "error": {
    "code": "INVALID_CREDENTIALS",
    "message": "Invalid username or password"
  }
}
```

---

## 👥 User Endpoints

### **GET /api/users/me**
Get current user profile.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439011",
    "username": "john_doe",
    "email": "john@example.com",
    "createdAt": "2024-01-15T10:30:00Z",
    "lastSeen": "2024-01-15T15:45:00Z"
  }
}
```

### **GET /api/users/search**
Search for users by username.

**Query Parameters:**
- `q` (string, required): Search query
- `limit` (integer, optional): Maximum results (default: 10)

**Example:**
```
GET /api/users/search?q=john&limit=5
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": [
    {
      "id": "507f1f77bcf86cd799439011",
      "username": "john_doe",
      "online": true,
      "lastSeen": "2024-01-15T15:45:00Z"
    },
    {
      "id": "507f1f77bcf86cd799439012",
      "username": "johnny_smith",
      "online": false,
      "lastSeen": "2024-01-15T14:30:00Z"
    }
  ]
}
```

---

## 🏠 Room Endpoints

### **GET /api/rooms**
Get all rooms for the current user.

**Headers:**
```
Authorization: Bearer <jwt_token>
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": [
    {
      "id": "507f1f77bcf86cd799439013",
      "name": "General Chat",
      "type": "group",
      "members": [
        "507f1f77bcf86cd799439011",
        "507f1f77bcf86cd799439012"
      ],
      "createdBy": "507f1f77bcf86cd799439011",
      "createdAt": "2024-01-15T10:00:00Z",
      "lastMessage": {
        "content": "Hello everyone!",
        "senderId": "507f1f77bcf86cd799439011",
        "createdAt": "2024-01-15T15:30:00Z"
      },
      "unreadCount": 3
    }
  ]
}
```

### **POST /api/rooms**
Create a new room.

**Request Body:**
```json
{
  "name": "Project Discussion",
  "type": "group",
  "members": [
    "507f1f77bcf86cd799439012",
    "507f1f77bcf86cd799439013"
  ]
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439014",
    "name": "Project Discussion",
    "type": "group",
    "members": [
      "507f1f77bcf86cd799439011",
      "507f1f77bcf86cd799439012",
      "507f1f77bcf86cd799439013"
    ],
    "createdBy": "507f1f77bcf86cd799439011",
    "createdAt": "2024-01-15T16:00:00Z"
  }
}
```

### **GET /api/rooms/:roomId**
Get room details.

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439013",
    "name": "General Chat",
    "type": "group",
    "members": [
      {
        "id": "507f1f77bcf86cd799439011",
        "username": "john_doe",
        "online": true
      },
      {
        "id": "507f1f77bcf86cd799439012",
        "username": "jane_smith",
        "online": false
      }
    ],
    "createdBy": "507f1f77bcf86cd799439011",
    "createdAt": "2024-01-15T10:00:00Z"
  }
}
```

### **POST /api/rooms/direct**
Create or get direct message room.

**Request Body:**
```json
{
  "userId": "507f1f77bcf86cd799439012"
}
```

**Response (200 OK or 201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439015",
    "type": "direct",
    "members": [
      "507f1f77bcf86cd799439011",
      "507f1f77bcf86cd799439012"
    ],
    "createdAt": "2024-01-15T16:15:00Z"
  }
}
```

---

## 💬 Message Endpoints

### **GET /api/rooms/:roomId/messages**
Get messages from a room with pagination.

**Query Parameters:**
- `limit` (integer, optional): Number of messages (default: 50, max: 100)
- `offset` (integer, optional): Offset for pagination (default: 0)
- `before` (string, optional): Get messages before this message ID

**Example:**
```
GET /api/rooms/507f1f77bcf86cd799439013/messages?limit=20&offset=0
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "messages": [
      {
        "id": "507f1f77bcf86cd799439016",
        "roomId": "507f1f77bcf86cd799439013",
        "senderId": "507f1f77bcf86cd799439011",
        "content": "Hello everyone!",
        "createdAt": "2024-01-15T15:30:00Z",
        "updatedAt": "2024-01-15T15:30:00Z",
        "editedAt": null,
        "deleted": false,
        "status": {
          "507f1f77bcf86cd799439012": {
            "delivered": true,
            "deliveredAt": "2024-01-15T15:30:05Z",
            "read": true,
            "readAt": "2024-01-15T15:31:00Z"
          }
        }
      }
    ],
    "hasMore": true,
    "total": 150
  }
}
```

### **POST /api/rooms/:roomId/messages**
Send a message to a room (HTTP fallback for WebSocket).

**Request Body:**
```json
{
  "content": "Hello everyone!"
}
```

**Response (201 Created):**
```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439017",
    "roomId": "507f1f77bcf86cd799439013",
    "senderId": "507f1f77bcf86cd799439011",
    "content": "Hello everyone!",
    "createdAt": "2024-01-15T15:35:00Z",
    "status": {
      "507f1f77bcf86cd799439012": {
        "delivered": false,
        "read": false
      }
    }
  }
}
```

### **PUT /api/messages/:messageId**
Edit a message (only by sender).

**Request Body:**
```json
{
  "content": "Hello everyone! (edited)"
}
```

**Response (200 OK):**
```json
{
  "success": true,
  "data": {
    "id": "507f1f77bcf86cd799439017",
    "content": "Hello everyone! (edited)",
    "editedAt": "2024-01-15T15:40:00Z"
  }
}
```

### **DELETE /api/messages/:messageId**
Delete a message (only by sender).

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Message deleted successfully"
}
```

---

## 🔔 Notification Endpoints

### **GET /api/notifications**
Get user notifications.

**Query Parameters:**
- `limit` (integer, optional): Number of notifications (default: 20)
- `unread` (boolean, optional): Filter unread notifications

**Response (200 OK):**
```json
{
  "success": true,
  "data": [
    {
      "id": "507f1f77bcf86cd799439018",
      "userId": "507f1f77bcf86cd799439011",
      "roomId": "507f1f77bcf86cd799439013",
      "messageId": "507f1f77bcf86cd799439016",
      "type": "message",
      "read": false,
      "createdAt": "2024-01-15T15:30:00Z"
    }
  ]
}
```

### **PUT /api/notifications/:notificationId/read**
Mark notification as read.

**Response (200 OK):**
```json
{
  "success": true,
  "message": "Notification marked as read"
}
```

### **PUT /api/notifications/read-all**
Mark all notifications as read.

**Response (200 OK):**
```json
{
  "success": true,
  "message": "All notifications marked as read"
}
```

---

## 🌐 WebSocket API

### **Connection**
```
ws://localhost:8080/ws?token=<jwt_token>
```

### **Message Format**
All WebSocket messages follow this format:
```json
{
  "type": "message_type",
  "roomId": "room_id_if_applicable",
  "payload": {
    // Type-specific data
  }
}
```

---

## 📨 WebSocket Message Types

### **1. message - Send/Receive Messages**

**Client → Server (Send):**
```json
{
  "type": "message",
  "roomId": "507f1f77bcf86cd799439013",
  "payload": {
    "content": "Hello everyone!",
    "tempId": "temp_123" // For optimistic updates
  }
}
```

**Server → Client (Receive):**
```json
{
  "type": "message",
  "roomId": "507f1f77bcf86cd799439013",
  "payload": {
    "messageId": "507f1f77bcf86cd799439016",
    "content": "Hello everyone!",
    "senderId": "507f1f77bcf86cd799439011",
    "timestamp": "2024-01-15T15:30:00Z",
    "tempId": "temp_123" // Confirms optimistic update
  }
}
```

### **2. typing - Typing Indicators**

**Client → Server:**
```json
{
  "type": "typing",
  "roomId": "507f1f77bcf86cd799439013",
  "payload": {
    "isTyping": true
  }
}
```

**Server → Client:**
```json
{
  "type": "typing",
  "roomId": "507f1f77bcf86cd799439013",
  "payload": {
    "userId": "507f1f77bcf86cd799439012",
    "username": "jane_smith",
    "isTyping": true
  }
}
```

### **3. read - Mark Message as Read**

**Client → Server:**
```json
{
  "type": "read",
  "roomId": "507f1f77bcf86cd799439013",
  "payload": {
    "messageId": "507f1f77bcf86cd799439016"
  }
}
```

**Server → Client (to sender):**
```json
{
  "type": "read",
  "roomId": "507f1f77bcf86cd799439013",
  "payload": {
    "messageId": "507f1f77bcf86cd799439016",
    "userId": "507f1f77bcf86cd799439012",
    "status": "read",
    "timestamp": "2024-01-15T15:31:00Z"
  }
}
```

### **4. delivered - Mark Message as Delivered**

**Client → Server:**
```json
{
  "type": "delivered",
  "roomId": "507f1f77bcf86cd799439013",
  "payload": {
    "messageId": "507f1f77bcf86cd799439016"
  }
}
```

**Server → Client (to sender):**
```json
{
  "type": "delivered",
  "roomId": "507f1f77bcf86cd799439013",
  "payload": {
    "messageId": "507f1f77bcf86cd799439016",
    "userId": "507f1f77bcf86cd799439012",
    "status": "delivered",
    "timestamp": "2024-01-15T15:30:05Z"
  }
}
```

### **5. presence - User Online/Offline Status**

**Server → Client:**
```json
{
  "type": "presence",
  "payload": {
    "userId": "507f1f77bcf86cd799439012",
    "online": true,
    "lastSeen": "2024-01-15T15:45:00Z"
  }
}
```

### **6. edit - Edit Message**

**Client → Server:**
```json
{
  "type": "edit",
  "roomId": "507f1f77bcf86cd799439013",
  "payload": {
    "messageId": "507f1f77bcf86cd799439016",
    "content": "Hello everyone! (edited)"
  }
}
```

**Server → Client:**
```json
{
  "type": "edit",
  "roomId": "507f1f77bcf86cd799439013",
  "payload": {
    "messageId": "507f1f77bcf86cd799439016",
    "content": "Hello everyone! (edited)",
    "editedAt": "2024-01-15T15:40:00Z"
  }
}
```

### **7. delete - Delete Message**

**Client → Server:**
```json
{
  "type": "delete",
  "roomId": "507f1f77bcf86cd799439013",
  "payload": {
    "messageId": "507f1f77bcf86cd799439016"
  }
}
```

**Server → Client:**
```json
{
  "type": "delete",
  "roomId": "507f1f77bcf86cd799439013",
  "payload": {
    "messageId": "507f1f77bcf86cd799439016"
  }
}
```

### **8. heartbeat - Keep Connection Alive**

**Client → Server:**
```json
{
  "type": "heartbeat"
}
```

### **9. join_room - Subscribe to Room**

**Client → Server:**
```json
{
  "type": "join_room",
  "roomId": "507f1f77bcf86cd799439013"
}
```

### **10. room_created - New Room Notification**

**Server → Client:**
```json
{
  "type": "room_created",
  "roomId": "507f1f77bcf86cd799439014",
  "payload": {
    "roomId": "507f1f77bcf86cd799439014",
    "roomType": "group",
    "name": "Project Discussion",
    "members": [
      "507f1f77bcf86cd799439011",
      "507f1f77bcf86cd799439012"
    ]
  }
}
```

### **11. error - Error Messages**

**Server → Client:**
```json
{
  "type": "error",
  "payload": {
    "code": "UNAUTHORIZED",
    "message": "You are not a member of this room"
  }
}
```

---

## 🔧 Error Codes

### **HTTP Error Codes**
- `400` - Bad Request (validation errors)
- `401` - Unauthorized (authentication required)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found (resource doesn't exist)
- `409` - Conflict (resource already exists)
- `429` - Too Many Requests (rate limiting)
- `500` - Internal Server Error

### **WebSocket Error Codes**
- `INVALID_MESSAGE` - Malformed message format
- `UNAUTHORIZED` - Authentication failed
- `FORBIDDEN` - Access denied to resource
- `ROOM_NOT_FOUND` - Room doesn't exist
- `NOT_ROOM_MEMBER` - User not a member of room
- `INVALID_PAYLOAD` - Invalid message payload
- `RATE_LIMITED` - Too many requests
- `INTERNAL_ERROR` - Server error

---

## 📝 Usage Examples

### **JavaScript/TypeScript Client**

#### **REST API Usage**
```typescript
// Authentication
const loginResponse = await fetch('/api/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({
    username: 'john_doe',
    password: 'password123'
  })
})

const { data } = await loginResponse.json()
const token = data.token

// Get rooms
const roomsResponse = await fetch('/api/rooms', {
  headers: { 'Authorization': `Bearer ${token}` }
})

const rooms = await roomsResponse.json()
```

#### **WebSocket Usage**
```typescript
// Connect to WebSocket
const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`)

// Send message
ws.send(JSON.stringify({
  type: 'message',
  roomId: 'room123',
  payload: {
    content: 'Hello!',
    tempId: 'temp_' + Date.now()
  }
}))

// Handle incoming messages
ws.onmessage = (event) => {
  const message = JSON.parse(event.data)
  
  switch (message.type) {
    case 'message':
      console.log('New message:', message.payload)
      break
    case 'typing':
      console.log('User typing:', message.payload)
      break
    case 'presence':
      console.log('Presence update:', message.payload)
      break
  }
}

// Send typing indicator
const sendTyping = (isTyping: boolean) => {
  ws.send(JSON.stringify({
    type: 'typing',
    roomId: currentRoomId,
    payload: { isTyping }
  }))
}

// Mark message as read
const markAsRead = (messageId: string) => {
  ws.send(JSON.stringify({
    type: 'read',
    roomId: currentRoomId,
    payload: { messageId }
  }))
}
```

---

## 🚀 Rate Limiting

### **REST API Limits**
- **Authentication**: 5 requests per minute per IP
- **Message sending**: 60 requests per minute per user
- **Room creation**: 10 requests per hour per user
- **General API**: 1000 requests per hour per user

### **WebSocket Limits**
- **Message sending**: 60 messages per minute per user
- **Typing indicators**: 10 per minute per room per user
- **Connection**: 1 active connection per user

---

## 🔍 Health Check

### **GET /health**
Check system health status.

**Response (200 OK):**
```json
{
  "status": "healthy",
  "database": "connected",
  "redis": "connected",
  "version": "1.0.0",
  "uptime": "2h 30m 15s"
}
```

---

## 📊 Response Format

All API responses follow this consistent format:

### **Success Response**
```json
{
  "success": true,
  "data": {
    // Response data
  }
}
```

### **Error Response**
```json
{
  "success": false,
  "error": {
    "code": "ERROR_CODE",
    "message": "Human readable error message",
    "details": {
      // Additional error details (optional)
    }
  }
}
```

### **Pagination Response**
```json
{
  "success": true,
  "data": {
    "items": [...],
    "pagination": {
      "page": 1,
      "limit": 20,
      "total": 150,
      "hasMore": true
    }
  }
}
```