# Kafka Code Fixes - Summary

## ปัญหาที่พบ
Kafka consumer ค้างหน้า HTTP server (HTTP ไม่ start)

## สาเหตุ
`Consume()` function สร้าง goroutine และเรียก `reader.FetchMessage(context.Background())` ตรงในฟังก์ชัน ซึ่ง block mainthread และ HTTP server ไม่สามารถ start ได้

## การแก้ไข

### 1. Lazy Registration Pattern (pkg/kp/microservice.go)
**ก่อน**: 
```go
// Consume() เรียก FetchMessage() ทันทีในหลัก thread → block HTTP
go func() {
    for {
        msg, err := reader.FetchMessage(context.Background())
        // ... handle message
    }
}()
```

**หลัง**:
```go
// Consume() เพียงแค่ลงทะเบียน handler ใน map
// ไม่สร้าง goroutine โดยตรง → HTTP สามารถ start ได้
m.kafkaHandlers[topic] = handler
```

Consumer จะถูกสร้างลาญหนึ่งเดียวครั้งใน `startKafkaConsumers()` ขณะ server start ทำให้ไม่มี blocking

### 2. Initialize kafkaClient Mutex (pkg/kp/microservice.go:381-384)
**ก่อน**: `mu: nil` (ไม่ initialize)
**หลัง**: `mu: &sync.RWMutex{}` (initialize ใช้ได้เลย)

ป้องกันการ panic ที่ `Close()` หรือ `CreateTopic()` เมื่อเรียก `kc.mu.Lock()`

### 3. BatchTimeout Duration (pkg/kp/microservice.go:387-392)
**ก่อน**: 
```go
BatchTimeout: time.Duration(conf.BatchTimeout)  // ต้องเป็น nanoseconds
// ถ้า BatchTimeout=0 → flush ทันที → hot loop
```

**หลัง**:
```go
batchTimeout := time.Duration(conf.BatchTimeout) * time.Millisecond
if batchTimeout <= 0 {
    batchTimeout = 500 * time.Millisecond
}
BatchTimeout: batchTimeout
```

คิดเป็น milliseconds และมี default 500ms

### 4. CommitInterval Duration (pkg/kp/kafka_consumer.go:43)
**ก่อน**: `CommitInterval: 1000` (เป็น nanoseconds = 1µs ⚡ เร็วมาก)
**หลัง**: `CommitInterval: time.Second` (เป็น 1 วินาที ✅)

## ผลลัพธ์
- ✅ HTTP server start ได้ปกติ
- ✅ Kafka consumers ทำงานขนานใน goroutines หลายตัว
- ✅ ไม่มี nil pointer dereference
- ✅ Proper timeout handling

## Flow ใหม่
```
1. main() เรียก app.Consume("test", handler)
   → Consume() ลงทะเบียน handler ใน map (ไม่ block)
   
2. app.Start() called
   → preHandle HTTP routes
   → startKafkaConsumers() loop and create consumers
   → HTTP ListenAndServe() ทำงาน
   
3. Consumers ทำงาน independently ใน goroutines แต่ละตัว
```

## Testing
```bash
# Build
go build ./cmd/server

# Run
./server
# HTTP server จะ start ใน <1 วินาที ✅
```
