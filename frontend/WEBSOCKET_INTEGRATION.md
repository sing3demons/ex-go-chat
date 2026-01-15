# WebSocket Integration - Task 33.1

## Overview
This document describes the WebSocket integration implementation for the real-time chat system.

## Requirements Addressed
- **Requirement 2.1**: Connect via WebSocket with valid JWT token
- **Requirement 3.4**: Broadcast messages to online members via WebSocket
- **Requirement 4.4**: Broadcast group messages via WebSocket

## Implementation Details

### 1. Connect on Login ✓
**Location**: `frontend/src/store/authStore.ts`

The WebSocket connection is established automatically when a user logs in or registers:

```typescript
login: async (identifier: string, password: string) => {
  // ... authentication logic ...
  
  // Connect WebSocket after successful login
  await websocket.connect(response.token);
  
  // ... update state ...
}
```

**Features**:
- Automatic connection on login/register
- Token-based authentication
- Connection stored in authStore

### 2. Reconnect on Page Reload ✓
**Location**: `frontend/src/App.tsx`

Added logic to reconnect WebSocket when the app loads with a persisted token:

```typescript
useEffect(() => {
  const reconnectWebSocket = async () => {
    if (isAuthenticated && token && !websocket.isConnected()) {
      console.log('Reconnecting WebSocket on app load...');
      await websocket.connect(token);
    }
  };
  reconnectWebSocket();
}, [isAuthenticated, token]);
```

**Features**:
- Automatic reconnection on page reload
- Checks if already connected to avoid duplicate connections
- Uses persisted token from localStorage

### 3. Subscribe to Room Events ✓
**Location**: `frontend/src/hooks/useWebSocket.ts`

Rooms are automatically subscribed when loaded:

```typescript
const subscribeToRooms = useCallback(() => {
  if (rooms.length > 0) {
    console.log(`Subscribing to ${rooms.length} rooms...`);
    rooms.forEach(room => {
      websocket.send({
        type: 'join_room',
        roomId: room.id,
        payload: {}
      });
    });
  }
}, [rooms]);

useEffect(() => {
  if (websocket.isConnected() && rooms.length > 0) {
    subscribeToRooms();
  }
}, [subscribeToRooms, rooms]);
```

**Features**:
- Automatic subscription when rooms are loaded
- Re-subscribes when room list changes
- Only subscribes when WebSocket is connected

### 4. Handle Incoming Messages ✓
**Location**: `frontend/src/hooks/useWebSocket.ts`

All WebSocket message types are handled:

```typescript
useEffect(() => {
  // Handle incoming messages
  const unsubMessage = websocket.on('message', (msg) => {
    const payload = msg.payload as ChatMessagePayload;
    addMessage({...});
    
    // Send delivery acknowledgment automatically
    websocket.send({
      type: 'delivered',
      roomId: msg.roomId,
      payload: { messageId: payload.messageId }
    });
  });

  // Handle message edits
  const unsubEdit = websocket.on('edit', (msg) => {...});

  // Handle message deletes
  const unsubDelete = websocket.on('delete', (msg) => {...});

  // Handle delivery status updates
  const unsubDelivered = websocket.on('delivered', (msg) => {...});

  // Handle read status updates
  const unsubRead = websocket.on('read', (msg) => {...});

  // Handle typing indicators
  const unsubTyping = websocket.on('typing', (msg) => {...});

  // Handle presence updates
  const unsubPresence = websocket.on('presence', (msg) => {...});

  // Handle errors
  const unsubError = websocket.on('error', (msg) => {...});

  return () => {
    // Cleanup all subscriptions
  };
}, [dependencies]);
```

**Features**:
- Handles all message types (message, edit, delete, delivered, read, typing, presence, error)
- Automatic delivery acknowledgment
- Updates local state for real-time UI updates
- Proper cleanup on unmount

### 5. Load Rooms After Connection ✓
**Location**: `frontend/src/pages/ChatPage.tsx`

Rooms are loaded after WebSocket connection is established:

```typescript
useEffect(() => {
  const initializeChat = async () => {
    if (token) {
      // Connect WebSocket if not already connected
      if (!isConnected) {
        await connect(token);
      }
      
      // Load rooms after ensuring WebSocket connection
      await useRoomStore.getState().loadRooms();
    }
  };
  initializeChat();
}, [token, isConnected, connect]);
```

**Features**:
- Ensures WebSocket is connected before loading rooms
- Loads rooms automatically on chat page mount
- Handles errors gracefully

## Store Updates

### ChatStore
**Location**: `frontend/src/store/chatStore.ts`

Added `updateMessageStatus` method to handle delivery and read status updates:

```typescript
updateMessageStatus: (messageId, userId, statusType, timestamp) => {
  // Updates message status for specific user
}
```

### PresenceStore
**Location**: `frontend/src/store/presenceStore.ts`

Added typing indicator support:

```typescript
setTypingUsers: (roomId, userId, isTyping) => {
  // Manages typing users per room
}

getTypingUsers: (roomId) => {
  // Returns array of typing user IDs for a room
}
```

### Type Updates
**Location**: `frontend/src/types/index.ts`

Added `join_room` to WSMessageType:

```typescript
export type WSMessageType = 
  | 'message'
  | 'typing'
  | 'read'
  | 'delivered'
  | 'presence'
  | 'edit'
  | 'delete'
  | 'heartbeat'
  | 'join_room'  // NEW
  | 'error';
```

## Connection Flow

```
1. User logs in
   ↓
2. authStore.login() called
   ↓
3. API authentication
   ↓
4. websocket.connect(token) called
   ↓
5. WebSocket connection established
   ↓
6. User navigates to /chat
   ↓
7. ChatPage mounts
   ↓
8. Rooms loaded via API
   ↓
9. useWebSocket hook subscribes to all rooms
   ↓
10. Real-time messages start flowing
```

## Reconnection Flow

```
1. User reloads page
   ↓
2. App.tsx mounts
   ↓
3. Checks if user is authenticated
   ↓
4. Checks if WebSocket is connected
   ↓
5. If not connected, calls websocket.connect(token)
   ↓
6. WebSocket reconnects with persisted token
   ↓
7. User navigates to /chat (or already there)
   ↓
8. Rooms loaded and subscribed
```

## Testing Checklist

- [x] WebSocket connects on login
- [x] WebSocket connects on register
- [x] WebSocket reconnects on page reload
- [x] Rooms are loaded after connection
- [x] Rooms are subscribed after loading
- [x] Incoming messages are received and displayed
- [x] Delivery acknowledgments are sent automatically
- [x] Message edits are handled
- [x] Message deletes are handled
- [x] Typing indicators are handled
- [x] Presence updates are handled
- [x] Read/delivered status updates are handled
- [x] WebSocket disconnects on logout
- [x] No TypeScript errors

## Files Modified

1. `frontend/src/App.tsx` - Added reconnection logic on app load
2. `frontend/src/pages/ChatPage.tsx` - Improved room loading after connection
3. `frontend/src/hooks/useWebSocket.ts` - Added logging for room subscriptions
4. `frontend/src/store/authStore.ts` - Already had WebSocket connection on login/register
5. `frontend/src/store/chatStore.ts` - Added updateMessageStatus method
6. `frontend/src/store/presenceStore.ts` - Added typing indicator support
7. `frontend/src/types/index.ts` - Added join_room message type

## Conclusion

The WebSocket integration is complete and meets all requirements:
- ✅ Connect on login (Requirement 2.1)
- ✅ Subscribe to room events (Requirement 3.4, 4.4)
- ✅ Handle incoming messages (Requirement 3.4, 4.4)
- ✅ Automatic reconnection on page reload
- ✅ Proper cleanup on logout
- ✅ Type-safe implementation with no TypeScript errors
