# Scalable Presence Design for Kubernetes

## Current Problem
- Online users stored in memory (per pod)
- Cannot scale beyond 1 pod effectively
- No cross-pod communication

## Solution: Redis-Centralized Architecture

### 1. Redis as Single Source of Truth
```
Redis:
- SET online_users: {userA, userB, userC, userD}
- HASH presence:userA: {online: true, lastSeen: timestamp, podId: pod1}
- HASH presence:userB: {online: true, lastSeen: timestamp, podId: pod2}
```

### 2. Pod Communication via Redis Pub/Sub
```
Pod 1 → Redis Pub/Sub → Pod 2, Pod 3, Pod N
```

### 3. Architecture Changes Needed:

#### A. Update Presence Service
```go
// Remove local memory store, use Redis only
type presenceService struct {
    redis *CacheService
    podID string  // Unique pod identifier
    log   *logger.Logger
}

func (s *presenceService) SetOnline(ctx context.Context, userID string) {
    // 1. Add to Redis online_users set
    s.redis.SAdd("online_users", userID)
    
    // 2. Set presence data with pod info
    presenceData := map[string]interface{}{
        "online": true,
        "lastSeen": time.Now().Unix(),
        "podId": s.podID,
    }
    s.redis.HSet(fmt.Sprintf("presence:%s", userID), presenceData)
    
    // 3. Publish presence update to all pods
    s.redis.Publish("presence_updates", map[string]interface{}{
        "userId": userID,
        "online": true,
        "podId": s.podID,
    })
}
```

#### B. Add Redis Pub/Sub Listener
```go
func (h *Hub) StartPresenceListener() {
    pubsub := h.redis.Subscribe("presence_updates")
    
    for msg := range pubsub.Channel() {
        var update PresenceUpdate
        json.Unmarshal([]byte(msg.Payload), &update)
        
        // Broadcast to local connections
        h.broadcastPresenceUpdate(update.UserID, update.Online)
    }
}
```

#### C. WebSocket Message Routing
```go
func (h *Hub) SendMessageToUser(userID string, message *WSMessage) {
    // 1. Check local connections first
    if conn, exists := h.connections[userID]; exists {
        conn.SendMessage(message)
        return
    }
    
    // 2. Check Redis for user's pod
    podID, err := h.redis.HGet(fmt.Sprintf("presence:%s", userID), "podId")
    if err == nil && podID != h.podID {
        // 3. Route via Redis to correct pod
        h.redis.Publish(fmt.Sprintf("messages:%s", podID), message)
    }
}
```

## 4. Kubernetes Deployment

### A. StatefulSet vs Deployment
```yaml
apiVersion: apps/v1
kind: Deployment  # Use Deployment for stateless scaling
metadata:
  name: chat-backend
spec:
  replicas: 3  # Can scale to any number
  selector:
    matchLabels:
      app: chat-backend
  template:
    metadata:
      labels:
        app: chat-backend
    spec:
      containers:
      - name: chat-backend
        image: chat-backend:latest
        env:
        - name: POD_NAME
          valueFrom:
            fieldRef:
              fieldPath: metadata.name
        - name: REDIS_ADDR
          value: "redis-service:6379"
```

### B. Redis Cluster
```yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: redis-cluster
spec:
  replicas: 3
  serviceName: redis-cluster
  selector:
    matchLabels:
      app: redis-cluster
```

## 5. Load Balancing Strategy

### A. Sticky Sessions (WebSocket)
```yaml
apiVersion: v1
kind: Service
metadata:
  name: chat-backend-ws
  annotations:
    nginx.ingress.kubernetes.io/affinity: "cookie"
    nginx.ingress.kubernetes.io/session-cookie-name: "chat-session"
spec:
  selector:
    app: chat-backend
  ports:
  - port: 8080
    targetPort: 8080
```

### B. Round Robin (HTTP API)
```yaml
apiVersion: v1
kind: Service
metadata:
  name: chat-backend-api
spec:
  selector:
    app: chat-backend
  ports:
  - port: 8080
    targetPort: 8080
```

## 6. Scaling Capabilities

### Current (Memory-based):
- ❌ Max: 1 pod
- ❌ Users per pod: ~1,000-5,000 WebSocket connections
- ❌ Total capacity: ~5,000 concurrent users

### After Redis-centralized:
- ✅ Max: Unlimited pods (horizontal scaling)
- ✅ Users per pod: ~1,000-5,000 WebSocket connections
- ✅ Total capacity: pods × 5,000 users
- ✅ Example: 10 pods = ~50,000 concurrent users

## 7. Implementation Priority

1. **Phase 1**: Move presence to Redis-only
2. **Phase 2**: Add Redis Pub/Sub for cross-pod communication
3. **Phase 3**: Implement message routing between pods
4. **Phase 4**: Add Kubernetes manifests
5. **Phase 5**: Load testing and optimization

## 8. Performance Considerations

### Redis Performance:
- **Memory**: ~1KB per online user
- **Operations**: ~100-1000 ops/sec per user
- **Network**: Pub/Sub adds ~10ms latency

### Pod Resource Requirements:
```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "100m"
  limits:
    memory: "512Mi" 
    cpu: "500m"
```

## 9. Monitoring & Observability

```yaml
# Metrics to track:
- chat_online_users_total
- chat_websocket_connections_per_pod
- chat_redis_operations_per_second
- chat_message_routing_latency
```