# Scalability Analysis: 1M Users

## ⚠️ ปัญหาปัจจุบันกับ 1 Million Users

### 1. **MongoDB Collection Scan - ปัญหาที่สุด** 🔴

**ปัญหาในพื้นที่:**

```go
// rooms collection: members field is an array
rooms.find({ members: userID })
// ← ต้อง scan ทุกๆ document ที่มี userID ในอาร์เรย์
```

**ผลกระทบกับ 1M users:**
- 1M users × 10-50 rooms/user = 10-50M membership records
- Query `members: userID` ต้อง scan entire collection
- ❌ ไม่สามารถ optimize ด้วย index array ได้มากนัก
- ❌ Database lock ขึ้น (concurrent queries ทดได้ไม่หลายรายพร้อมกัน)

**ประเมินเวลา:**
```
2 users:    5ms (cache miss)
100 users:  50ms
10K users:  5-10 seconds ⚠️
1M users:   30-60 seconds ❌ → TIMEOUT
```

---

### 2. **Message Storage - ปัญหาใหญ่เป็นที่สอง** 🟠

**ปัญหา:**
```
1M users × 10 rooms/user × 100 msgs/room 
= 1 Billion messages
```

**ผลกระทบ:**
- MongoDB ไม่ลงตัวสำหรับ timeseries data
- Sharding จะซับซ้อน
- TTL (auto-delete old messages) ทำไม่ได้ reliable

---

### 3. **Status Map ใน Message - Denormalization Problem** 🟠

```go
type Message struct {
    Status map[string]*Status // userID -> Status (100 KB!)
}
```

**ปัญหา:**
- Message ที่มี 1000 members → Status document ใหญ่ 100+ KB
- Group room messages → memory bloat
- ❌ ไม่ scale

---

### 4. **Presence/Online Status** 🟡

**ปัญหา:**
- Redis key per user: `presence:user:123`
- 1M users = 1M Redis keys
- ✅ Redis ทำได้ แต่ memory ≈ 5-10 GB
- ⚠️ Persistence issues

---

## ✅ Solutions: Scaling to 1M Users

### Phase 1: ทันที (ต้องทำ)

#### 1.1 **Separate User-Room Relationship Collection**

```go
// ✅ แทนที่ query rooms.find({members: userID})

type UserRoom struct {
    ID     string    // MongoDB ObjectID
    UserID string    // indexed
    RoomID string    // indexed
    Role   string    // "member", "admin"
    JoinedAt time.Time
}

// Create indexes:
// userRoom: {userID: 1, roomID: 1} - compound unique
// userRoom: {userID: 1} - find all rooms for user
// userRoom: {roomID: 1} - find all members of room
```

**ประโยชน์:**
- ✅ FindUserRooms: O(n) scan → O(1) indexed query
- ✅ FindByMembers: ง่ายขึ้น
- ✅ Scale to millions

**Implementation:**
```go
// Before
rooms.find({members: userID}).limit(50)
// Query time: 10-30s with 1M users

// After
userRooms.find({userID: userID}).limit(50)
// Query time: 5-10ms with 1M users
// 100-1000x faster! 🚀
```

---

#### 1.2 **Message Status Collection** (Denormalize)

```go
// ❌ Current: map[string]*Status in Message
// ✅ New: separate collection

type MessageStatus struct {
    ID        string // MongoDB ObjectID
    MessageID string // indexed
    UserID    string // indexed
    Delivered bool
    Read      bool
    // ... timestamps
}

// Indexes:
// {messageID: 1}
// {userID: 1, messageID: 1}
// {messageID: 1, userID: 1}
```

**ประโยชน์:**
- ✅ Message document เล็กลง (100 KB → 1 KB)
- ✅ Update status independent
- ✅ Aggregate status ได้ง่าย

---

#### 1.3 **Add Missing Indexes**

```go
// Add to CreateIndexes()

// User-Room relationship
userRoomsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
    {
        Keys: bson.D{
            {Key: "userId", Value: 1},
            {Key: "roomId", Value: 1},
        },
        Options: options.Index().SetUnique(true),
    },
    {Keys: bson.D{{Key: "userId", Value: 1}}},
    {Keys: bson.D{{Key: "roomId", Value: 1}}},
})

// Message with compound index for sorting
messagesCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
    {
        Keys: bson.D{
            {Key: "roomId", Value: 1},
            {Key: "createdAt", Value: -1},
        },
    },
})

// UserRoom pagination
userRoomsCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
    {
        Keys: bson.D{
            {Key: "userId", Value: 1},
            {Key: "createdAt", Value: -1},
        },
    },
})
```

---

### Phase 2: 100K-500K Users

#### 2.1 **MongoDB Sharding**

```
Shard by: userID
- Shard 1: Users A-M
- Shard 2: Users N-Z

Benefits:
✅ Parallel queries
✅ Smaller working set per shard
✅ Better indexing efficiency
```

**Shard Key:**
```go
// {userId: 1}  - for user-room lookups
// OR {roomId: 1} - for message lookups (depends on pattern)
```

---

#### 2.2 **Message TTL & Archival**

```go
// Auto-delete messages older than 90 days
messagesCollection.Indexes().CreateMany(ctx, []mongo.IndexModel{
    {
        Keys: bson.D{{Key: "createdAt", Value: 1}},
        Options: options.Index().SetExpireAfterSeconds(90 * 24 * 3600),
    },
})

// Archive to S3/Glacier for compliance
func ArchiveOldMessages(ctx context.Context) {
    // Find messages > 90 days
    // Store in S3 with room/user prefix
    // Delete from MongoDB
}
```

---

#### 2.3 **Connection Pooling**

```go
// ใน MongoDB client options
opts := options.Client().
    SetMaxPoolSize(500).        // ↑ from default 100
    SetMinPoolSize(100).
    SetMaxConnIdleTime(30 * time.Second)
```

---

### Phase 3: 500K-1M+ Users

#### 3.1 **Message Database Separation**

```
Architecture:
┌─────────────────────┐
│ Operational DB      │
│ (Users, Rooms)      │
│ MongoDB 1           │
└─────────────────────┘
          ↓
┌─────────────────────┐
│ Message Database    │
│ TimeSeries DB       │
│ InfluxDB/TimeScale  │
└─────────────────────┘
          ↓
┌─────────────────────┐
│ Message Archive     │
│ S3/Parquet          │
└─────────────────────┘
```

**ประโยชน์:**
- ✅ Optimize เพิ่มเติมได้ต่างกัน
- ✅ Message query ไม่ block user queries
- ✅ Time-series DB ดี 100x ต่อ message

---

#### 3.2 **Read Replicas**

```
MongoDB Replica Set:
- Primary (write): Users, Rooms, UserRooms
- Secondary 1 (read): UserRooms queries (lighter load)
- Secondary 2 (read): Message history (heavy queries)

Application:
router.ReadPreference("secondary")  // For reads
```

---

#### 3.3 **Caching Evolution**

```
Current (1000x speedup):
Redis L1: User, Room, UserRooms (15 min TTL)

Phase 2 (Local + Distributed):
Redis L1 (Local): User, Room (1 min)
Redis L2 (Shared): UserRooms cache (5 min)

Phase 3 (CDN + Cache):
CDN: Static user profiles
Redis: Hot room data
Local cache: Recent messages
```

---

## 📊 Expected Performance at 1M Users

### Current State (2 users):
```
FindUserRooms:     5ms  ✅
FindByMembers:     10ms ✅
SendMessage:       100ms ✅
GetMessages:       50ms ✅
```

### With Phase 1 (UserRoom collection):
```
FindUserRooms:     5-10ms   ✅ (100x improvement)
FindByMembers:     20ms     ✅
SendMessage:       150ms    ✅
GetMessages:       100ms    ✅
```

### With Phase 2 (Sharding + Read replicas):
```
FindUserRooms:     5ms      ✅
FindByMembers:     10ms     ✅
SendMessage:       100ms    ✅
GetMessages:       20ms     ✅ (parallel reads)
```

### With Phase 3 (Message DB separation):
```
FindUserRooms:     5ms      ✅
FindByMembers:     10ms     ✅
SendMessage:       100ms    ✅
GetMessages:       5-10ms   ✅✅✅
```

---

## 🚨 Red Flags ตอนนี้ต้องแก้ไม่งั้นจะ crash

### 1. ❌ `rooms.find({members: userID})` ← **ต้องแยกออก**
   - Array query scan ไม่ scale
   - Replace ด้วย UserRoom collection

### 2. ❌ `Status map[string]*Status` in Message
   - Large documents กับ 1000s members
   - Move to separate collection

### 3. ❌ No TTL on messages
   - Database size grows forever
   - Add TTL index immediately

### 4. ⚠️ Single MongoDB instance
   - Replica set required สำหรับ production
   - ไม่ scale beyond 10-50K users

---

## 🎯 Recommended Action Plan

### Immediate (ต้องทำตอนนี้):
1. ✅ Create `UserRoom` collection with proper indexes
2. ✅ Move `MessageStatus` to separate collection
3. ✅ Add TTL on messages
4. ✅ Set up MongoDB Replica Set

### Short term (1-3 months):
1. Implement sharding by userID
2. Add read replicas
3. Move to distributed caching

### Long term (6-12 months):
1. Separate message database
2. Time-series DB for messages
3. Archive system

---

## 📈 Scaling Checklist

- [ ] UserRoom collection exists
- [ ] MessageStatus separate collection
- [ ] Proper indexes on UserRoom
- [ ] TTL index on messages
- [ ] MongoDB Replica Set
- [ ] Connection pooling configured
- [ ] Cache warming strategy
- [ ] Sharding plan documented
- [ ] Read replica setup
- [ ] Message archival system

**ปัจจุบัน: ✅ 0/10 ต้องเริ่มทำทันที**
