# Quick Start Guide - Real-time Chat System

## 🚀 Get Started in 5 Minutes

### Prerequisites
- Go 1.21+
- Node.js 18+
- MongoDB 6.0+ (or Docker)

### Step 1: Clone and Setup

```bash
# Clone the repository
git clone <your-repo-url>
cd chat-app
```

### Step 2: Start MongoDB

**Option A: Using Docker (Recommended)**
```bash
make docker-up
```

**Option B: Local MongoDB**
```bash
# Make sure MongoDB is running on localhost:27017
mongod
```

### Step 3: Start Backend

```bash
# Install Go dependencies
go mod download

# Copy environment file
cp .env.example .env

# Run the server
make run
```

Backend will start on **http://localhost:8080**

### Step 4: Start Frontend

Open a new terminal:

```bash
cd frontend

# Install dependencies
npm install

# Copy environment file
cp .env.example .env

# Start development server
npm run dev
```

Frontend will start on **http://localhost:5173**

### Step 5: Test the Application

1. Open browser: **http://localhost:5173**
2. Click "Register" to create a new account
3. Fill in:
   - Username: `alice`
   - Email: `alice@example.com`
   - Password: `password123`
4. Click "Register"
5. You'll be redirected to the chat page

### Step 6: Test Real-time Chat

**Open a second browser window (incognito mode):**

1. Go to **http://localhost:5173**
2. Register another user:
   - Username: `bob`
   - Email: `bob@example.com`
   - Password: `password123`

**Create a room and chat:**

1. In Alice's window, you should see rooms (if any exist)
2. Select a room or create one via API
3. Send messages and see them appear in real-time!

## 📝 Quick API Test

### Register a User
```bash
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com",
    "password": "password123"
  }'
```

### Login
```bash
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "identifier": "testuser",
    "password": "password123"
  }'
```

Save the token from the response!

### Create a Room
```bash
TOKEN="your-jwt-token-here"

curl -X POST http://localhost:8080/api/rooms \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{
    "type": "group",
    "name": "Test Room",
    "members": []
  }'
```

### Get Rooms
```bash
curl -X GET http://localhost:8080/api/rooms \
  -H "Authorization: Bearer $TOKEN"
```

## 🧪 Run Tests

### Backend Tests
```bash
# Test authentication
./scripts/test_auth.sh

# Test WebSocket (requires wscat: npm install -g wscat)
./scripts/test_websocket.sh
```

### Frontend Build
```bash
cd frontend
npm run build
```

## 🔧 Troubleshooting

### Backend won't start
- Check MongoDB is running: `mongosh` or `docker ps`
- Check port 8080 is available: `lsof -i :8080`
- Check `.env` file exists with correct values

### Frontend won't start
- Check Node.js version: `node --version` (should be 18+)
- Clear node_modules: `rm -rf node_modules && npm install`
- Check port 5173 is available

### WebSocket connection fails
- Check backend is running
- Check browser console for errors
- Verify token is valid
- Check CORS settings

### Messages not appearing
- Check WebSocket connection status
- Check browser console for errors
- Verify you're in the same room
- Check backend logs

## 📚 Next Steps

- Read [README.md](README.md) for detailed documentation
- Check [docs/FINAL_SUMMARY.md](docs/FINAL_SUMMARY.md) for architecture
- Review API endpoints in [README.md](README.md)
- Explore WebSocket protocol in [docs/CHECKPOINT_2_WEBSOCKET.md](docs/CHECKPOINT_2_WEBSOCKET.md)

## 🎯 Common Tasks

### Stop Services
```bash
# Stop backend: Ctrl+C in terminal

# Stop frontend: Ctrl+C in terminal

# Stop MongoDB
make docker-down
```

### Clean Build
```bash
# Backend
make clean
make build

# Frontend
cd frontend
rm -rf dist node_modules
npm install
npm run build
```

### View Logs
```bash
# Backend logs are in console

# MongoDB logs
docker logs mongodb
```

## 💡 Tips

1. **Use two browser windows** (one normal, one incognito) to test real-time features
2. **Check browser console** for WebSocket connection status
3. **Use MongoDB Compass** to view database contents
4. **Enable browser DevTools Network tab** to debug API calls
5. **Check backend logs** for detailed error messages

## 🐛 Known Issues

- Presence tracking resets on server restart (in-memory)
- No file upload support yet
- No message search yet
- Single instance only (no horizontal scaling)

## 🚀 Production Deployment

For production deployment, see:
- [docs/CHECKPOINT_3_FINAL.md](docs/CHECKPOINT_3_FINAL.md) - Deployment considerations
- [docs/FINAL_SUMMARY.md](docs/FINAL_SUMMARY.md) - Architecture overview

## 📞 Support

For issues or questions:
1. Check documentation in `docs/` folder
2. Review code comments
3. Check GitHub issues (if applicable)

---

**Happy Chatting! 💬**
