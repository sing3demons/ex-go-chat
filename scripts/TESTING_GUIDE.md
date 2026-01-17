# Testing Guide - Real-time Chat System

This guide provides comprehensive instructions for testing the Real-time Chat System.

## 🚀 Quick Start Testing

### Prerequisites

Before testing, ensure you have:
- Go 1.21+ installed
- Node.js 18+ installed
- MongoDB running (locally or via Docker)

### Option 1: Using Docker (Recommended)

```bash
# Start all services
docker compose up -d

# Wait for services to be ready (about 10 seconds)
sleep 10

# Run automated tests
./scripts/test_system.sh
```

### Option 2: Manual Setup

#### Step 1: Start MongoDB

```bash
# Using Docker
docker run -d -p 27017:27017 --name mongodb mongo:latest

# OR using local MongoDB
mongod --dbpath /path/to/data
```

#### Step 2: Start Backend

```bash
# Terminal 1: Start backend server
make run

# OR
go run cmd/server/main.go
```

Wait for the message: `Server started on :8080`

#### Step 3: Start Frontend

```bash
# Terminal 2: Start frontend dev server
cd frontend
npm run dev
```

Wait for the message: `Local: http://localhost:5173/`

#### Step 4: Run Automated Tests

```bash
# Terminal 3: Run test script
./scripts/test_system.sh
```

---

## 🧪 Manual Testing Checklist

### 1. Authentication Testing

#### Test 1.1: User Registration
1. Open browser to `http://localhost:5173`
2. Click "Register" or navigate to `/register`
3. Fill in the form:
   - Username: `testuser1`
   - Email: `testuser1@test.com`
   - Password: `Test123456`
4. Click "Register"
5. **Expected**: Redirect to login page with success message

#### Test 1.2: User Login
1. Navigate to `/login`
2. Enter credentials:
   - Username: `testuser1`
   - Password: `Test123456`
3. Click "Login"
4. **Expected**: Redirect to chat page, see empty room list

#### Test 1.3: Invalid Login
1. Try logging in with wrong password
2. **Expected**: Error message displayed

#### Test 1.4: Token Persistence
1. Login successfully
2. Refresh the page
3. **Expected**: Still logged in, no redirect to login

---

### 2. Room Management Testing

#### Test 2.1: Create Group Room
1. Login as `testuser1`
2. Click "New Group" button
3. Enter room name: `Test Group`
4. **Note**: Need another user to add as member
5. Click "Create"
6. **Expected**: New room appears in room list

#### Test 2.2: Room List Display
1. Check room list on left sidebar
2. **Expected**: See created rooms with:
   - Room name/avatar
   - Last message preview (if any)
   - Unread count badge

#### Test 2.3: Select Room
1. Click on a room in the list
2. **Expected**: 
   - Room becomes highlighted
   - Chat window shows room name
   - Message history loads (if any)

---

### 3. Real-time Messaging Testing

**Note**: This requires two users. Open two browser windows (or use incognito mode).

#### Test 3.1: Send Message
**Window 1 (User 1):**
1. Login as `testuser1`
2. Select a room
3. Type a message: "Hello from User 1"
4. Press Enter or click Send
5. **Expected**: 
   - Message appears immediately (optimistic update)
   - Spinner icon shows pending state
   - Changes to checkmark when confirmed

**Window 2 (User 2):**
1. Login as `testuser2`
2. Select the same room
3. **Expected**: 
   - Message from User 1 appears in real-time
   - No page refresh needed

#### Test 3.2: Message Status Tracking
**Window 1 (User 1):**
1. Send a message
2. **Expected**: Status progression:
   - ⏳ Pending (spinner)
   - ✓ Sent (single checkmark)
   - ✓✓ Delivered (double checkmark)
   - ✓✓ Read (blue checkmarks when User 2 views)

**Window 2 (User 2):**
1. View the message
2. **Expected**: User 1 sees read status update

#### Test 3.3: Optimistic Updates
1. Send a message
2. **Expected**: Message appears instantly before server confirmation
3. If network is slow, you'll see the pending spinner

#### Test 3.4: Failed Message Handling
1. Disconnect network (turn off WiFi)
2. Try sending a message
3. **Expected**:
   - Message shows pending for 10 seconds
   - Then shows red X (failed)
   - Retry button appears
4. Reconnect network
5. Click retry button
6. **Expected**: Message sends successfully

---

### 4. Typing Indicators Testing

**Window 1 (User 1):**
1. Select a room
2. Start typing in the message input
3. **Expected**: Nothing visible to you

**Window 2 (User 2):**
1. In the same room
2. **Expected**: See "User 1 is typing..." with animated dots
3. When User 1 stops typing
4. **Expected**: Typing indicator disappears after 3 seconds

---

### 5. Presence Tracking Testing

**Window 1 (User 1):**
1. Login
2. **Expected**: Green dot appears next to your name

**Window 2 (User 2):**
1. Login
2. **Expected**: 
   - User 1 shows green dot (online)
   - In direct chat, online status visible

**Window 1 (User 1):**
1. Close browser or logout
2. **Expected**: User 2 sees User 1 go offline (gray dot)

---

### 6. Notification Testing

**Window 1 (User 1):**
1. Logout or close browser

**Window 2 (User 2):**
1. Send message to User 1
2. **Expected**: Message saved for offline delivery

**Window 1 (User 1):**
1. Login again
2. **Expected**: 
   - Notification badge shows unread count
   - Click to see notification list
   - Notifications for messages received while offline

---

### 7. Message History Testing

#### Test 7.1: Load History
1. Select a room with existing messages
2. **Expected**: Last 50 messages load automatically

#### Test 7.2: Infinite Scroll
1. In a room with many messages
2. Scroll to top
3. **Expected**: Older messages load automatically

#### Test 7.3: Pagination
1. Keep scrolling up
2. **Expected**: Messages load in batches of 50

---

### 8. Responsive Design Testing

#### Test 8.1: Mobile View (< 640px)
1. Resize browser to mobile size
2. **Expected**:
   - Hamburger menu appears
   - Room list hidden by default
   - Full-screen chat view
   - Touch-friendly buttons (44px+)

#### Test 8.2: Tablet View (640-1024px)
1. Resize to tablet size
2. **Expected**:
   - Adaptive spacing
   - Sidebar toggleable
   - Optimized layout

#### Test 8.3: Desktop View (≥ 1024px)
1. Full desktop size
2. **Expected**:
   - Persistent sidebar
   - Multi-column layout
   - All features visible

---

### 9. Error Handling Testing

#### Test 9.1: Network Error
1. Disconnect network
2. Try any action
3. **Expected**: Error message displayed

#### Test 9.2: Authentication Error
1. Manually delete token from localStorage
2. Try accessing chat page
3. **Expected**: Redirect to login

#### Test 9.3: Validation Error
1. Try registering with invalid email
2. **Expected**: Validation error message

---

### 10. WebSocket Connection Testing

#### Test 10.1: Initial Connection
1. Login
2. Open browser DevTools → Network → WS
3. **Expected**: WebSocket connection established

#### Test 10.2: Auto-reconnection
1. While logged in, disconnect network
2. **Expected**: Connection error banner appears
3. Reconnect network
4. **Expected**: 
   - Auto-reconnection happens
   - Error banner disappears
   - Messages sync

#### Test 10.3: Manual Retry
1. If auto-reconnection fails
2. Click "Retry" button in error banner
3. **Expected**: Connection re-established

---

## 🔧 Automated Testing

### Run All Tests

```bash
# Make script executable
chmod +x scripts/test_system.sh

# Run tests
./scripts/test_system.sh
```

### Test Output

The script will test:
1. ✓ Backend Health Check
2. ✓ User Registration
3. ✓ User Login
4. ✓ Room Management
5. ✓ Message Operations
6. ✓ Frontend Availability

### Expected Results

```
🧪 Real-time Chat System - System Test
======================================

✓ PASS: Backend server is running
✓ PASS: User 1 registration successful
✓ PASS: User 2 registration successful
✓ PASS: Duplicate registration correctly rejected
✓ PASS: User 1 login successful
✓ PASS: User 2 login successful
✓ PASS: Invalid login correctly rejected
✓ PASS: Get rooms successful
✓ PASS: Group room creation successful
✓ PASS: Get rooms after creation successful
✓ PASS: Get messages successful
✓ PASS: Frontend server is running

━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Test Summary
━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━
Passed: 12
Failed: 0

✓ All tests passed!

🎉 System is working correctly!
```

---

## 🐛 Troubleshooting

### Backend Not Starting

**Problem**: `make run` fails

**Solutions**:
1. Check MongoDB is running: `mongosh` or `mongo`
2. Check port 8080 is free: `lsof -i :8080`
3. Check Go version: `go version` (need 1.21+)
4. Check environment variables in `.env`

### Frontend Not Starting

**Problem**: `npm run dev` fails

**Solutions**:
1. Install dependencies: `cd frontend && npm install`
2. Check Node version: `node --version` (need 18+)
3. Check port 5173 is free: `lsof -i :5173`
4. Check `.env` file exists in `frontend/`

### MongoDB Connection Error

**Problem**: Backend can't connect to MongoDB

**Solutions**:
1. Start MongoDB: `docker run -d -p 27017:27017 mongo`
2. Check connection string in `.env`: `MONGODB_URI=mongodb://localhost:27017`
3. Test connection: `mongosh mongodb://localhost:27017`

### WebSocket Connection Failed

**Problem**: Real-time features not working

**Solutions**:
1. Check backend is running
2. Check WebSocket URL in `frontend/.env`: `VITE_WS_URL=ws://localhost:8080`
3. Open DevTools → Console for errors
4. Check browser supports WebSocket

### CORS Errors

**Problem**: Frontend can't call backend API

**Solutions**:
1. Check backend CORS configuration
2. Ensure frontend URL is in allowed origins
3. Check API URL in `frontend/.env`: `VITE_API_URL=http://localhost:8080`

---

## 📊 Performance Testing

### Load Testing

```bash
# Install Apache Bench (if not installed)
# macOS: brew install httpd
# Linux: apt-get install apache2-utils

# Test registration endpoint
ab -n 1000 -c 10 -p register.json -T application/json \
   http://localhost:8080/api/auth/register

# Test login endpoint
ab -n 1000 -c 10 -p login.json -T application/json \
   http://localhost:8080/api/auth/login
```

### WebSocket Load Testing

```bash
# Install wscat
npm install -g wscat

# Connect to WebSocket
wscat -c "ws://localhost:8080/ws?token=YOUR_JWT_TOKEN"

# Send test message
{"type":"chat_message","data":{"room_id":"ROOM_ID","content":"Test"}}
```

---

## ✅ Test Completion Checklist

- [ ] All automated tests pass
- [ ] User registration works
- [ ] User login works
- [ ] Room creation works
- [ ] Messages send in real-time
- [ ] Message status updates work
- [ ] Typing indicators work
- [ ] Presence tracking works
- [ ] Notifications work
- [ ] Message history loads
- [ ] Infinite scroll works
- [ ] Responsive design works (mobile/tablet/desktop)
- [ ] Error handling works
- [ ] WebSocket auto-reconnection works
- [ ] Optimistic updates work
- [ ] Failed message retry works

---

## 🎯 Next Steps After Testing

1. **If all tests pass**: Ready for production deployment
2. **If some tests fail**: Check troubleshooting section
3. **Performance issues**: Run load tests and optimize
4. **UI/UX issues**: Gather user feedback and iterate

---

## 📝 Test Results Template

Use this template to document your test results:

```
Test Date: _______________
Tester: _______________
Environment: [ ] Local [ ] Docker [ ] Production

Backend Tests:
- [ ] Health Check
- [ ] User Registration
- [ ] User Login
- [ ] Room Management
- [ ] Message Operations

Frontend Tests:
- [ ] Page Load
- [ ] Authentication UI
- [ ] Chat Interface
- [ ] Real-time Updates
- [ ] Responsive Design

Integration Tests:
- [ ] WebSocket Connection
- [ ] Message Delivery
- [ ] Status Updates
- [ ] Presence Tracking
- [ ] Notifications

Issues Found:
1. _______________
2. _______________
3. _______________

Overall Status: [ ] PASS [ ] FAIL
Notes: _______________
```

---

**Last Updated**: January 15, 2026
**Version**: 1.0.0
