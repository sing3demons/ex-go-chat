# Retry Mechanisms

This document describes the retry mechanisms implemented in the chat application to handle transient failures and improve reliability.

## Overview

The application implements retry logic at multiple levels:
1. **WebSocket Connection**: Automatic reconnection with exponential backoff
2. **Message Sending**: Manual retry for failed messages
3. **API Requests**: Automatic retry for critical operations
4. **Manual Connection Retry**: User-triggered reconnection

## Retry Utilities

### Location
`frontend/src/utils/retry.ts`

### Functions

#### `retryWithExponentialBackoff<T>`
Retries an operation with exponentially increasing delays between attempts.

**Parameters:**
- `fn: () => Promise<T>` - The async function to retry
- `options`:
  - `maxRetries: number` - Maximum number of retry attempts (default: 3)
  - `initialDelay: number` - Initial delay in milliseconds (default: 1000)
  - `maxDelay: number` - Maximum delay cap (default: 30000)
  - `shouldRetry?: (error: any) => boolean` - Optional function to determine if error is retryable

**Behavior:**
- Delay doubles after each failed attempt: 1s → 2s → 4s → 8s...
- Capped at `maxDelay` to prevent excessive waiting
- Throws the last error if all retries fail

**Example:**
```typescript
const data = await retryWithExponentialBackoff(
  async () => await api.getRooms(),
  { maxRetries: 3, initialDelay: 1000 }
);
```

#### `retryWithLinearBackoff<T>`
Retries an operation with linearly increasing delays.

**Parameters:**
- Same as exponential backoff

**Behavior:**
- Delay increases linearly: 1s → 2s → 3s → 4s...
- More predictable timing than exponential
- Better for operations with consistent failure patterns

#### `isRetryableError(error: any): boolean`
Determines if an error should trigger a retry attempt.

**Retryable Errors:**
- Network errors (no response received)
- Timeout errors
- 5xx server errors (500-599)
- 429 Too Many Requests

**Non-Retryable Errors:**
- 4xx client errors (except 429)
- Authentication errors (401, 403)
- Validation errors (400)

## Implementation Details

### 1. WebSocket Auto-Reconnection

**Location:** `frontend/src/services/websocket.ts`

**Features:**
- Automatic reconnection on connection loss
- Exponential backoff: 1s → 2s → 4s → 8s → 16s → 30s (max)
- Maximum 10 reconnection attempts
- Resets retry count on successful connection
- Preserves message handlers across reconnections

**Behavior:**
```typescript
// Automatically reconnects when connection drops
websocket.connect(token); // Will auto-retry on failure
```

**Configuration:**
- Initial delay: 1 second
- Max delay: 30 seconds
- Max attempts: 10

### 2. Message Retry

**Location:** `frontend/src/components/MessageItem.tsx`

**Features:**
- Manual retry button for failed messages
- Visual feedback (red X icon)
- Removes failed message and resends
- Optimistic update on retry

**User Experience:**
1. Message fails to send (shows red X after 10s timeout)
2. User clicks retry button
3. Failed message is removed from UI
4. New optimistic message appears
5. Message is resent to server

**Implementation:**
```typescript
const handleRetry = () => {
  // Remove failed message
  chatStore.removeMessage(message.id);
  
  // Resend with new temp ID
  sendMessage(message.content);
};
```

### 3. API Request Retry

**Location:** `frontend/src/services/api.ts`

**Retried Operations:**

#### Get Rooms (Critical)
- **Max Retries:** 3
- **Initial Delay:** 1 second
- **Reason:** Essential for app initialization
- **Retry Condition:** Network errors only

```typescript
async getRooms(): Promise<Room[]> {
  return retryWithExponentialBackoff(
    async () => {
      const response = await this.client.get('/api/rooms');
      return response.data.data || [];
    },
    { maxRetries: 3, initialDelay: 1000, shouldRetry: isRetryableError }
  );
}
```

#### Get Messages (Critical)
- **Max Retries:** 3
- **Initial Delay:** 1 second
- **Reason:** Important for chat history
- **Retry Condition:** Network errors only

#### Get Pending Notifications
- **Max Retries:** 2
- **Initial Delay:** 1 second
- **Reason:** Nice to have, not critical
- **Retry Condition:** Network errors only

**Non-Retried Operations:**
- Login/Register (user should retry manually)
- Create/Update operations (avoid duplicates)
- Delete operations (avoid unintended side effects)

### 4. Manual Connection Retry

**Location:** `frontend/src/pages/ChatPage.tsx`

**Features:**
- Retry button in connection error banner
- Only shown when disconnected
- Visual loading state during retry
- Error feedback if retry fails

**User Experience:**
1. Connection fails (error banner appears)
2. User clicks "Retry Connection" button
3. Button shows loading spinner
4. On success: banner disappears, chat reconnects
5. On failure: error message updates

**Implementation:**
```typescript
const handleRetryConnection = async () => {
  if (!token || isRetrying) return;
  
  setIsRetrying(true);
  setConnectionError(null);
  
  try {
    await connect(token);
    console.log('Reconnected successfully');
  } catch (error: any) {
    console.error('Retry failed:', error);
    setConnectionError(error.message || 'Failed to reconnect');
  } finally {
    setIsRetrying(false);
  }
};
```

## Best Practices

### When to Use Retry Logic

**DO Retry:**
- Network errors (connection lost, timeout)
- Server errors (5xx)
- Rate limiting (429)
- Idempotent read operations
- WebSocket reconnections

**DON'T Retry:**
- Authentication errors (401, 403)
- Validation errors (400)
- Not found errors (404)
- Non-idempotent operations (create, update, delete)
- User-initiated actions (let user retry manually)

### Retry Strategy Selection

**Exponential Backoff:**
- Use for: Network issues, server overload
- Benefit: Reduces server load, gives time to recover
- Example: WebSocket reconnection, API retries

**Linear Backoff:**
- Use for: Predictable delays, rate limiting
- Benefit: More consistent timing
- Example: Polling operations

**No Backoff (Immediate):**
- Use for: Quick operations, user-triggered retries
- Benefit: Fast feedback
- Example: Manual message retry

### Error Handling

Always provide user feedback:
```typescript
try {
  await retryWithExponentialBackoff(operation);
} catch (error) {
  // Show user-friendly error message
  setError('Failed to load data. Please try again.');
  console.error('Operation failed after retries:', error);
}
```

## Testing Retry Logic

### Manual Testing

1. **Network Interruption:**
   - Disconnect network
   - Observe auto-reconnection
   - Check exponential backoff timing

2. **Message Failure:**
   - Stop backend server
   - Send message
   - Verify retry button appears
   - Restart server and retry

3. **API Retry:**
   - Use network throttling
   - Load rooms/messages
   - Verify automatic retries

### Monitoring

Check browser console for retry logs:
```
Retry attempt 1/3 failed: Network error
Retry attempt 2/3 failed: Network error
Retry attempt 3/3 succeeded
```

## Configuration

### Adjusting Retry Parameters

**WebSocket:**
```typescript
// In websocket.ts
private maxReconnectAttempts = 10;
private reconnectDelay = 1000;
private maxReconnectDelay = 30000;
```

**API Requests:**
```typescript
// In api.ts
retryWithExponentialBackoff(operation, {
  maxRetries: 3,        // Adjust retry count
  initialDelay: 1000,   // Adjust initial delay
  maxDelay: 30000,      // Adjust max delay
});
```

## Future Enhancements

Potential improvements:
1. **Retry with Jitter:** Add randomness to prevent thundering herd
2. **Circuit Breaker:** Stop retrying after repeated failures
3. **Retry Queue:** Queue failed operations for batch retry
4. **Adaptive Retry:** Adjust strategy based on error patterns
5. **Retry Metrics:** Track retry success/failure rates

## Related Documentation

- [Optimistic Updates](./OPTIMISTIC_UPDATES.md) - Message sending with optimistic UI
- [Error Handling](./ERROR_HANDLING.md) - Error handling strategies
- [WebSocket Integration](./WEBSOCKET_INTEGRATION.md) - WebSocket connection management
