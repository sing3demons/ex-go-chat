# Real-time Chat System

> ✅ **Status**: Production Ready | 🚀 **Version**: 2.0.0 Full Stack MVP

A complete real-time chat application with Golang backend and React frontend, featuring WebSocket communication, user authentication, room management, and message persistence.

## 📚 Quick Links

- 🚀 **[Quick Start Guide](QUICKSTART.md)** - Get started in 5 minutes
- 🌐 **[Deployment Guide](DEPLOY_NOW.md)** - Deploy to production in 10 minutes
- 📖 **[Full Documentation](docs/FINAL_SUMMARY.md)** - Complete architecture overview
- 🤝 **[Contributing](CONTRIBUTING.md)** - How to contribute
- 📝 **[Project Summary](FINAL_PROJECT_SUMMARY.md)** - Complete project overview

## ✨ Features

- 🔐 User authentication (JWT)
- 💬 Real-time messaging via WebSocket
- 👥 1-on-1 and group chat
- 📝 Message history with pagination
- ✏️ Message edit and delete
- ✅ Message delivery and read status
- 👀 Typing indicators
- 🟢 Online/offline presence tracking
- 🔔 Notifications for offline users
- 📱 Responsive design

## 🚀 Quick Start

See [QUICKSTART.md](QUICKSTART.md) for a 5-minute setup guide.

## Tech Stack

### Backend
- **Language**: Golang
- **Framework**: net/http
- **WebSocket**: gorilla/websocket
- **Database**: MongoDB
- **Authentication**: JWT

### Frontend
- **Framework**: React 18
- **Language**: TypeScript
- **Build Tool**: Vite
- **State Management**: Zustand
- **Styling**: Tailwind CSS
- **HTTP Client**: Axios
- **Routing**: React Router

## Project Structure

```
.
├── backend/
│   ├── cmd/server/          # Application entry point
│   ├── config/              # Configuration management
│   ├── internal/
│   │   ├── handler/         # HTTP and WebSocket handlers
│   │   ├── middleware/      # Authentication middleware
│   │   ├── models/          # Data models
│   │   ├── repository/      # Data access layer
│   │   ├── service/         # Business logic layer
│   │   └── websocket/       # WebSocket management
│   ├── pkg/                 # Shared packages
│   ├── docs/                # Documentation
│   └── scripts/             # Test scripts
│
└── frontend/
    ├── src/
    │   ├── components/      # React components
    │   ├── hooks/           # Custom hooks
    │   ├── pages/           # Page components
    │   ├── services/        # API and WebSocket services
    │   ├── store/           # Zustand stores
    │   └── types/           # TypeScript types
    └── public/              # Static assets
```

## Prerequisites

- Go 1.21 or higher
- Node.js 18 or higher
- MongoDB 6.0 or higher
- Docker (optional, for running MongoDB)

## Getting Started

### Backend Setup

1. **Clone the repository**

```bash
git clone <repository-url>
cd realtime-chat-system
```

2. **Install Go dependencies**

```bash
go mod download
```

3. **Setup MongoDB**

Option A: Using Docker (Recommended)
```bash
make docker-up
```

Option B: Install MongoDB locally
- Follow instructions at https://www.mongodb.com/docs/manual/installation/

4. **Configure environment variables**

```bash
cp .env.example .env
# Edit .env with your configuration
```

5. **Run the backend**

```bash
make run
```

The server will start on `http://localhost:8080`

### Frontend Setup

1. **Navigate to frontend directory**

```bash
cd frontend
```

2. **Install dependencies**

```bash
npm install
```

3. **Configure environment variables**

```bash
cp .env.example .env
# Edit .env if needed (default values work for local development)
```

4. **Run the frontend**

```bash
npm run dev
```

The app will start on `http://localhost:5173`

## Development

```bash
make help           # Show all available commands
make deps           # Download dependencies
make run            # Run the application
make build          # Build the application
make test           # Run tests
make test-coverage  # Run tests with coverage
make clean          # Clean build artifacts
make docker-up      # Start MongoDB with Docker
make docker-down    # Stop MongoDB Docker container
```

### Running Tests

```bash
make test
```

### Building for Production

```bash
make build
./bin/server
```

## API Endpoints

### Authentication
- `POST /api/auth/register` - Register a new user
- `POST /api/auth/login` - Login user

### Rooms
- `GET /api/rooms` - Get user's rooms
- `POST /api/rooms` - Create a new group room
- `POST /api/rooms/:id/members` - Add members to a room
- `DELETE /api/rooms/:id/members` - Remove members from a room

### Messages
- `GET /api/rooms/:id/messages` - Get chat history (with pagination)

### Notifications
- `GET /api/notifications` - Get notifications (with pagination)
- `GET /api/notifications/pending` - Get unread notifications
- `GET /api/notifications/unread-count` - Get unread count
- `POST /api/notifications/mark-read` - Mark notification as read
- `POST /api/notifications/mark-all-read` - Mark all as read

### WebSocket
- `WS /ws?token=<jwt_token>` - WebSocket connection

## WebSocket Usage

### Connecting

```javascript
const token = "your_jwt_token";
const ws = new WebSocket(`ws://localhost:8080/ws?token=${token}`);

ws.onopen = () => {
  console.log("Connected");
};

ws.onmessage = (event) => {
  const message = JSON.parse(event.data);
  console.log("Received:", message);
};
```

### Message Format

All WebSocket messages follow this format:

```json
{
  "type": "message|typing|read|delivered|presence|edit|delete|heartbeat|error",
  "roomId": "room_id",
  "payload": { ... }
}
```

### Sending Messages

#### Chat Message
```json
{
  "type": "message",
  "roomId": "room_id",
  "payload": {
    "content": "Hello world!"
  }
}
```

#### Typing Indicator
```json
{
  "type": "typing",
  "roomId": "room_id",
  "payload": {
    "isTyping": true
  }
}
```

#### Mark as Read
```json
{
  "type": "read",
  "roomId": "room_id",
  "payload": {
    "messageId": "message_id"
  }
}
```

#### Edit Message
```json
{
  "type": "edit",
  "roomId": "room_id",
  "payload": {
    "messageId": "message_id",
    "content": "Updated content"
  }
}
```

#### Delete Message
```json
{
  "type": "delete",
  "roomId": "room_id",
  "payload": {
    "messageId": "message_id"
  }
}
```

#### Heartbeat
```json
{
  "type": "heartbeat"
}
```

### Receiving Messages

#### Chat Message
```json
{
  "type": "message",
  "roomId": "room_id",
  "payload": {
    "messageId": "msg_id",
    "content": "Hello world!",
    "senderId": "user_id",
    "timestamp": "2026-01-14T10:00:00Z"
  }
}
```

#### Typing Indicator
```json
{
  "type": "typing",
  "roomId": "room_id",
  "payload": {
    "userId": "user_id",
    "username": "john_doe",
    "isTyping": true
  }
}
```

#### Status Update
```json
{
  "type": "read",
  "roomId": "room_id",
  "payload": {
    "messageId": "msg_id",
    "userId": "user_id",
    "status": "read",
    "timestamp": "2026-01-14T10:00:00Z"
  }
}
```

#### Presence Update
```json
{
  "type": "presence",
  "payload": {
    "userId": "user_id",
    "online": true,
    "lastSeen": "2026-01-14T10:00:00Z"
  }
}
```

#### Error
```json
{
  "type": "error",
  "payload": {
    "code": "UNAUTHORIZED",
    "message": "You are not a member of this room"
  }
}
```

## Testing

### Manual Testing

1. Start the server:
```bash
make run
```

2. Test authentication:
```bash
./scripts/test_auth.sh
```

3. Test WebSocket (requires wscat: `npm install -g wscat`):
```bash
./scripts/test_websocket.sh
```

### Running Tests

```bash
make test
```

### Building for Production

```bash
make build
./bin/server
```

## License

MIT


## 📖 Documentation

- [Quick Start Guide](QUICKSTART.md) - Get started in 5 minutes
- [Deployment Guide](docs/DEPLOYMENT.md) - Production deployment
- [Contributing Guide](CONTRIBUTING.md) - How to contribute
- [Final Summary](docs/FINAL_SUMMARY.md) - Project overview
- [Checkpoint Documents](docs/) - Development milestones

## 🏗️ Architecture

### System Overview

```
┌─────────────┐         ┌─────────────┐
│   Browser   │◄───────►│   Frontend  │
│  (Client)   │  HTTPS  │   (React)   │
└─────────────┘         └─────────────┘
                              │
                              │ HTTP/WS
                              ▼
                        ┌─────────────┐
                        │   Backend   │
                        │  (Golang)   │
                        └─────────────┘
                              │
                              │ MongoDB
                              ▼
                        ┌─────────────┐
                        │   MongoDB   │
                        │  (Database) │
                        └─────────────┘
```

### Key Components

**Backend:**
- HTTP Server (REST API)
- WebSocket Server (Real-time)
- Authentication (JWT)
- Business Logic (Services)
- Data Access (Repositories)

**Frontend:**
- React Components
- State Management (Zustand)
- WebSocket Client
- HTTP Client (Axios)
- Routing (React Router)

## 🔒 Security

- Password hashing with bcrypt
- JWT token authentication
- Authorization checks on all endpoints
- Input validation
- CORS configuration
- Secure WebSocket connections (WSS in production)

## 🚀 Performance

- Connection pooling (MongoDB)
- Buffered channels (WebSocket)
- Thread-safe operations (RWMutex)
- Database indexes
- Message pagination
- Efficient state updates

## 🧪 Testing

### Manual Testing

```bash
# Backend
./scripts/test_auth.sh
./scripts/test_websocket.sh

# Frontend
cd frontend && npm run build
```

### Automated Testing

```bash
# Backend
make test

# Frontend
cd frontend && npm run lint
```

## 📝 API Documentation

See [README.md](README.md) for complete API documentation including:
- Authentication endpoints
- Room management
- Message operations
- Notifications
- WebSocket protocol

## 🤝 Contributing

We welcome contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## 📄 License

MIT License - see [LICENSE](LICENSE) file for details

## 🙏 Acknowledgments

- [gorilla/websocket](https://github.com/gorilla/websocket) - WebSocket implementation
- [MongoDB](https://www.mongodb.com/) - Database
- [React](https://react.dev/) - Frontend framework
- [Zustand](https://github.com/pmndrs/zustand) - State management
- [Tailwind CSS](https://tailwindcss.com/) - Styling

## 📞 Support

- 📧 Email: support@example.com
- 🐛 Issues: [GitHub Issues](https://github.com/your-repo/issues)
- 💬 Chat: Join our community

## 🗺️ Roadmap

### v1.1 (Next Release)
- [ ] Redis integration for scaling
- [ ] Message search
- [ ] File upload
- [ ] User profiles

### v2.0 (Future)
- [ ] Voice/Video calls
- [ ] Message encryption
- [ ] Mobile app
- [ ] Admin dashboard

## 📊 Project Status

- ✅ Backend MVP: Complete
- ✅ Frontend MVP: Complete
- ✅ Documentation: Complete
- ✅ Testing: Manual tests passing
- 🚧 Production Deployment: Ready
- 🚧 Automated Tests: In progress

## 🌟 Features Showcase

### Real-time Messaging
Messages appear instantly across all connected clients using WebSocket technology.

### Presence Tracking
See who's online in real-time with automatic status updates.

### Message Status
Track message delivery and read status for each recipient.

### Responsive Design
Works seamlessly on desktop, tablet, and mobile devices.

---

**Built with ❤️ using Golang and React**

**Last Updated**: January 14, 2026
