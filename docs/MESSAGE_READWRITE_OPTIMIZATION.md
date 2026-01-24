# Message Read/Write Optimization Guide

## 🔴 ปัญหาปัจจุบัน (High Write Volume)

### 1. **Status Map ใน Message - ปัญหาใหญ่ที่สุด**

```go
type Message struct {
    Status map[string]*Status  // ❌ This is huge!
}

// Example: Group room 1000 members
// Message size: 1 KB content + ~100 KB status map = 101 KB per message
// 1M messages in room = 100 GB storage! 💥
```

**ปัญหา:**
- ❌ Update status → rewrite entire 100 KB document
- ❌ 1000 concurrent read/write → heavy lock contention
- ❌ Slow because field is embedded in document
- ❌ Pagination pulls huge documents

---

### 2. **Update Pattern ล้าสมัย**

```go
// Current pattern (BAD for write-heavy):
func UpdateDeliveryStatus(messageID, userID) {
    msg := FindByID(messageID)           // Read entire 100KB document
    msg.Status[userID].Delivered = true  // Update 1 field
    Update(msg)                          // Write entire 100KB document back
}
// Per update: 1 read + 1 write of 100 KB
// 1000 updates/sec × 100 KB = 100 MB/sec network traffic 🔥
```

---

### 3. **No Batch Operations**

```go
// Frontend sends 50 "mark as read" events
// Server does:
for i := 0; i < 50; i++ {
    UpdateReadStatus(msgID[i], userID)  // 50 separate DB updates 😱
}
// = 50 DB round trips instead of 1
```

---

### 4. **Pagination Problem**

```go
// Getting 50 messages
for msg in messages {
    // Each message = 100 KB
    // Total: 50 × 100 KB = 5 MB transferred
    // Rendering in UI: slow, memory heavy 📉
}
```

---

## ✅ Solutions (Priority Order)

### Phase 1: Immediate (Must Do Now)

#### 1.1 **Separate Message Status Collection**

```go
// ❌ Before: status in message
type Message struct {
    ID      string
    Content string
    Status  map[string]*Status  // ← 100 KB!
}

// ✅ After: separate collection
type Message struct {
    ID      string
    Content string
    // Status moved to separate collection
}

type MessageStatus struct {
    ID        string    // auto-generated
    MessageID string    // indexed
    UserID    string    // indexed
    Delivered bool
    DeliveredAt *time.Time
    Read      bool
    ReadAt    *time.Time
}

// Indexes:
// {messageID: 1, userID: 1}  - find status for user
// {messageID: 1}             - find all status for message
// {userID: 1}                - find all messages user has status
```

**ประโยชน์:**
- Message: 1 KB → Fast to load
- Status: Separate updates → No lock contention
- Write pattern: upsert instead of full update

**Implementation:**

```go
// Current bad pattern:
msg.Status[userID] = status
repo.Update(msg)  // ← Rewrites entire document

// New good pattern:
repo.UpsertStatus(messageID, userID, status)
// ← Only updates status document, not message
```

---

#### 1.2 **Bulk Update Operation**

```go
// Before: 50 individual updates
for each marking {
    UpdateReadStatus(msgID, userID)  // 50 DB calls
}

// After: Batch operation
BulkUpdateReadStatus(markings)  // 1 DB call with 50 records
// 50x faster! 🚀
```

**Implementation:**

```go
type MarkReadRequest struct {
    MessageID string
    Timestamp time.Time
}

// Single batch operation
func (s *messageService) BulkMarkRead(ctx context.Context, userID string, markings []MarkReadRequest) error {
    // Execute single bulk write
    statusRepo.BulkUpsert(ctx, userID, markings)
}
```

---

#### 1.3 **Projection - Don't Load Status in GetMessages**

```go
// ❌ Before: load everything including 100KB status
messages := repo.FindByRoom(roomID)  // 50 msgs × 100 KB each = 5 MB

// ✅ After: load only needed fields
messages := repo.FindByRoomProjection(roomID, []string{"id", "senderId", "content", "createdAt"})
// = 50 msgs × 1 KB = 50 KB transferred 🚀
// = 100x faster!

// Get status separately if needed:
statuses := repo.FindStatusesForMessages(msgIDs, userID)  // Only for current user
```

---

#### 1.4 **Caching Strategy for Heavy Reads**

```go
// Current: every page load = DB query
messages = repo.FindByRoom(roomID, limit=50, offset=offset)

// Better: cache message list
messages = cache.GetOrFetch("room:{roomID}:messages:0-50", () => {
    return repo.FindByRoom(roomID, 0, 50)
})
// TTL: 5 minutes
```

---

### Phase 2: Medium-term (100K+ Users)

#### 2.1 **Queue-Based Status Updates**

```
Architecture:
┌──────────────┐
│   Frontend   │
└──────┬───────┘
       │
       ▼
┌──────────────────────┐
│  Message Queue       │
│  (Redis/RabbitMQ)    │
└──────┬───────────────┘
       │
       ▼
┌──────────────────────┐
│  Batch Worker        │
│  (Update 100/sec)    │
└──────┬───────────────┘
       │
       ▼
┌──────────────────────┐
│  MongoDB             │
│  Bulk insert         │
└──────────────────────┘
```

**ประโยชน์:**
- ✅ Decouple frontend from DB
- ✅ Batch writes (1000 status/sec → 10 bulk ops/sec)
- ✅ Resilient to DB slowness
- ✅ Can replay if needed

**Implementation:**

```go
// Frontend: send to Redis queue
func MarkReadWebSocket(msgID, userID) {
    queue.Push("mark_read", {msgID, userID, time.Now()})  // Fast!
}

// Worker: batch process
func BatchStatusWorker() {
    for {
        batch := queue.PopN("mark_read", 1000)  // Get 1000 at a time
        repo.BulkInsertStatus(batch)
        time.Sleep(100 * time.Millisecond)
    }
}
```

---

#### 2.2 **Message Sharding by Time**

```
Collection Strategy:
- messages           (live, < 7 days)
- messages_archive   (7+ days)
- messages_2024_01   (older, compressed)

Benefits:
✅ Smaller indexes
✅ Faster queries
✅ Easier archival
```

---

#### 2.3 **Read Replicas for Analytics**

```
Setup:
- Primary: Write messages, status updates
- Secondary-1: Read for GetMessages (lighter queries)
- Secondary-2: Read for analytics/reporting

Application:
// Immediate consistency needed
msg = primary.FindByID(msgID)

// Eventually consistent OK
stats = secondary.GetMessageCount(roomID)
```

---

### Phase 3: Advanced (500K+ Users)

#### 3.1 **TimeSeries Database for Messages**

```
Switch from MongoDB to InfluxDB/TimescaleDB:

InfluxDB (Better for messages):
✅ Optimized for time-series
✅ Compression 100x better
✅ Range queries ultra-fast
✅ Retention policies built-in

Example:
measurements:
  - message
    tags: [roomId, senderId]
    fields: [content, edited]
    timestamp
```

---

#### 3.2 **Event Sourcing for Status**

```
Instead of updating status, append events:

✅ Current approach:
Status table:
| messageID | userID | delivered | read |
| msg1      | user1  | true      | true |

❌ Better: Event log
Events table:
| messageID | userID | event       | timestamp |
| msg1      | user1  | DELIVERED   | T1        |
| msg1      | user1  | READ        | T2        |

Benefits:
- Immutable writes (no contention)
- Full audit trail
- Can replay to compute state
- Infinitely scalable writes
```

---

#### 3.3 **Connection Pool Optimization**

```go
// Add to connection options
opts := options.Client().
    SetMaxPoolSize(1000).         // ↑ for heavy write
    SetMinPoolSize(500).
    SetMaxConnIdleTime(1 * time.Second)  // Aggressive reuse
    SetConnectionMonitoringMode(cm).     // Monitor connection health
```

---

## 📊 Performance Comparison

### Current Pattern (Bad):

```
1000 concurrent "mark as read" events:
- 1000 db.find() → 1000 MB transferred ❌
- 1000 db.update() → 1000 MB transferred ❌
- Total: 2000 MB/sec, lots of lock contention

Latency: 500ms per operation ⚠️
```

### After Phase 1 (Good):

```
1000 concurrent "mark as read" events:
- Batch into 10 operations (100 status each)
- db.bulkWrite(statuses) → 1 MB transferred ✅
- Status is separate collection → no message lock ✅

Latency: 50ms per batch (5ms per operation) 🚀
```

### After Phase 2 (Better):

```
1000 concurrent "mark as read" events:
- Queue immediately (< 1 ms) ✅
- Worker batches every 100 ms ✅
- Single bulk write to status collection ✅

Latency: 1ms frontend, 100ms eventual consistency ✅✅
```

### After Phase 3 (Best):

```
1000 concurrent "mark as read" events:
- Event log append (immutable, parallel) ✅
- Zero contention ✅
- Infinitely scalable ✅

Latency: 1-5ms to event log + eventual consistency ✅✅✅
```

---

## 🎯 Implementation Priority

### ✅ Immediate (Phase 1) - Do This First

- [ ] Create MessageStatus collection
- [ ] Add indexes on messageID, userID
- [ ] Implement BulkUpsertStatus
- [ ] Update GetMessages to use projection
- [ ] Add message list cache (Redis)

**Impact:** 100x speedup for read/write

---

### 🟡 Short-term (Phase 2) - 1-3 months

- [ ] Implement queue-based status updates
- [ ] Add batch status worker
- [ ] Message archival by date
- [ ] Read replicas setup

**Impact:** Ready for 100K+ users

---

### 🔮 Long-term (Phase 3) - 6+ months

- [ ] TimeSeries database evaluation
- [ ] Event sourcing for status
- [ ] Advanced sharding strategy
- [ ] Connection pool tuning

**Impact:** Ready for 1M+ users

---

## Code Example: Phase 1 Implementation

### MessageStatus Repository

```go
type MessageStatusRepository interface {
    UpsertStatus(ctx context.Context, messageID, userID string, status *MessageStatus) error
    BulkUpsertStatus(ctx context.Context, statuses []*MessageStatus) error
    FindStatusesForMessage(ctx context.Context, messageID string) ([]*MessageStatus, error)
    FindStatusForUser(ctx context.Context, messageID, userID string) (*MessageStatus, error)
}

// Usage in service:
func UpdateDeliveryStatus(messageID, userID string) error {
    status := &MessageStatus{
        MessageID: messageID,
        UserID:    userID,
        Delivered: true,
        DeliveredAt: time.Now(),
    }
    return statusRepo.UpsertStatus(ctx, status)
    // ← Only updates status collection, not message
}

// Batch operation:
func BulkMarkRead(markings []MarkReadRequest) error {
    statuses := make([]*MessageStatus, len(markings))
    for i, m := range markings {
        statuses[i] = &MessageStatus{
            MessageID: m.MessageID,
            UserID: userID,
            Read: true,
            ReadAt: time.Now(),
        }
    }
    return statusRepo.BulkUpsertStatus(ctx, statuses)
    // ← Single batch write instead of 50 individual updates
}
```

### GetMessages with Projection

```go
func GetMessages(roomID string, limit, offset int) ([]*MessageDTO, error) {
    // Load only needed fields
    messages := repo.FindByRoomProjection(
        ctx,
        roomID,
        []string{"id", "senderId", "content", "createdAt"},
        limit,
        offset,
    )

    // Convert to DTO (no status info)
    dtos := make([]*MessageDTO, len(messages))
    for i, msg := range messages {
        dtos[i] = &MessageDTO{
            ID:        msg.ID,
            SenderID:  msg.SenderID,
            Content:   msg.Content,
            CreatedAt: msg.CreatedAt,
            // Status loaded separately if needed
        }
    }

    // Get current user's status for these messages if needed
    if includeStatus {
        statuses := statusRepo.FindStatusForUser(
            ctx,
            ExtractMessageIDs(messages),
            currentUserID,
        )
        AttachStatuses(dtos, statuses)
    }

    return dtos, nil
}
```

---

## 📈 Scaling Roadmap

| Phase | Users | Messages/Day | Solution | Status |
|-------|-------|------------|----------|--------|
| 1 | 1K | 100K | MessageStatus collection | ⏳ TODO |
| 2 | 100K | 10M | Queue + Batching | ⏳ TODO |
| 3 | 1M | 100M | TimeSeries DB | ⏳ TODO |
| 4 | 10M | 1B | Event Sourcing | ⏳ TODO |

---

## ✅ Checklist

- [ ] MessageStatus collection created
- [ ] Indexes on MessageStatus (messageID, userID)
- [ ] BulkUpsertStatus implemented
- [ ] Projection query added
- [ ] Message list cache added
- [ ] Old message TTL configured
- [ ] Batch worker queue setup (optional)
- [ ] Read replica configured (optional)

**Next Step:** Implement Phase 1 to get 100x read/write speedup! 🚀
