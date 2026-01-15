# Message Status Tracking Implementation

## Overview

This document describes the implementation of real-time message status tracking, which shows delivery and read status for messages in the chat system.

## Features Implemented

### 1. Automatic Read Receipts
- Uses Intersection Observer API to detect when messages are visible
- Automatically sends read receipts when messages are 50% visible on screen
- Only sends read receipts for messages from other users
- Prevents duplicate read receipts with tracking mechanism

### 2. Visual Status Indicators

Messages display different status states with appropriate icons and tooltips:

#### For Own Messages:
- **Sending**: Spinning loader (gray) - "Sending..."
- **Failed**: Red X icon - "Failed to send"
- **Sent**: Single gray checkmark - "Sent"
- **Delivered**: 
  - All delivered: Double gray checkmark - "Delivered to X people"
  - Some delivered: Double light gray checkmark - "Delivered to X of Y"
- **Read**:
  - All read: Double blue checkmark - "Read by X people"
  - Some read: Double light blue checkmark - "Read by X of Y"

#### For Group Chats:
- Shows partial status (e.g., "Read by 2 of 5")
- Different colors indicate different completion levels
- Tooltips show exact counts

### 3. Real-time Updates
- Status updates are received via WebSocket
- UI updates immediately when status changes
- No page refresh required

## Technical Implementation

### Intersection Observer (`frontend/src/components/MessageList.tsx`)

```typescript
// Setup observer to detect visible messages
observerRef.current = new IntersectionObserver(
  (entries) => {
    entries.forEach((entry) => {
      if (entry.isIntersecting) {
        const messageId = entry.target.getAttribute('data-message-id');
        const senderId = entry.target.getAttribute('data-sender-id');
        
        if (messageId && senderId && senderId !== user.id) {
          // Mark as read if not our own message
          if (!observedMessagesRef.current.has(messageId)) {
            observedMessagesRef.current.add(messageId);
            markAsRead(roomId, messageId);
          }
        }
      }
    });
  },
  {
    root: messagesContainerRef.current,
    threshold: 0.5, // 50% visible
  }
);
```

### Status Icon Logic (`frontend/src/components/MessageItem.tsx`)

The status icon shows different states based on message status:

1. **Check pending/failed states first**
2. **Calculate status from all recipients**:
   - `allRead`: Every recipient has read
   - `someRead`: At least one recipient has read
   - `allDelivered`: Every recipient has received
   - `someDelivered`: At least one recipient has received
3. **Show appropriate icon with tooltip**

### WebSocket Integration (`frontend/src/hooks/useWebSocket.ts`)

Status updates are handled via WebSocket:

```typescript
// Handle delivery status updates
const unsubDelivered = websocket.on('delivered', (msg) => {
  const payload = msg.payload as StatusPayload;
  updateMessageStatus(payload.messageId, payload.userId, 'delivered', payload.timestamp);
});

// Handle read status updates
const unsubRead = websocket.on('read', (msg) => {
  const payload = msg.payload as StatusPayload;
  updateMessageStatus(payload.messageId, payload.userId, 'read', payload.timestamp);
});
```

### Store Updates (`frontend/src/store/chatStore.ts`)

Status updates are stored per user:

```typescript
updateMessageStatus: (messageId, userId, statusType, timestamp) => {
  // Update status for specific user
  if (statusType === 'delivered') {
    newStatus[userId].delivered = timestamp;
  } else if (statusType === 'read') {
    newStatus[userId].read = timestamp;
  }
}
```

## User Experience Flow

### Sending a Message
1. User sends message
2. Message appears with "sending..." indicator
3. Server confirms → changes to "sent" (single checkmark)
4. Recipients receive → changes to "delivered" (double gray checkmark)
5. Recipients view → changes to "read" (double blue checkmark)

### Receiving a Message
1. Message appears in chat
2. If message is visible (50%+ on screen):
   - Automatically sends read receipt
   - Sender sees status change to "read"
3. If message is not visible:
   - Waits until user scrolls to it
   - Sends read receipt when visible

### Group Chat Status
1. Shows aggregate status across all members
2. Partial status shown when some (but not all) have read
3. Tooltip shows exact counts (e.g., "Read by 3 of 5")

## Configuration

### Visibility Threshold
Messages must be 50% visible to trigger read receipt.
Adjust in `MessageList.tsx`:

```typescript
{
  root: messagesContainerRef.current,
  threshold: 0.5, // Change this value (0.0 to 1.0)
}
```

### Status Colors
- **Blue (#3B82F6)**: All read
- **Light Blue (#60A5FA)**: Some read
- **Gray (#9CA3AF)**: Delivered
- **Light Gray (#D1D5DB)**: Sent/Partial delivery
- **Red (#EF4444)**: Failed

## Requirements Validated

This implementation validates:
- **Requirement 5.1**: Delivery status updates when recipients receive messages ✓
- **Requirement 5.2**: Read status updates when recipients view messages ✓
- **Requirement 5.3**: Delivery status broadcast to sender ✓
- **Requirement 5.4**: Read status broadcast to sender ✓

## Performance Considerations

### Optimization Techniques:
1. **Intersection Observer**: Efficient visibility detection
2. **Duplicate Prevention**: Tracks observed messages to prevent redundant receipts
3. **Batch Updates**: Status updates are batched by WebSocket
4. **Cleanup**: Observers are properly disconnected on unmount

### Memory Management:
- Observer references are cleaned up on component unmount
- Observed message set is cleared when room changes
- No memory leaks from event listeners

## Testing

To test message status tracking:

1. **Delivery Status**: 
   - Send message to offline user
   - User comes online
   - Observe status change to "delivered"

2. **Read Status**:
   - Send message to online user
   - User views message
   - Observe status change to "read"

3. **Group Chat**:
   - Send message to group with 3+ members
   - Observe partial status as members read
   - Check tooltip shows correct counts

4. **Visibility Detection**:
   - Send multiple messages
   - Scroll slowly through chat
   - Verify read receipts sent only for visible messages

## Future Enhancements

Potential improvements:
- Show individual read status per user in group chats
- Add "seen by" list on message hover
- Show typing indicators with read status
- Add notification when message is read
- Configurable visibility threshold per user
