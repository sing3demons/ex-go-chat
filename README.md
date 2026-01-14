# Real-time Chat System

A real-time chat system built with Golang, MongoDB, and WebSocket.

## Features

- User registration and authentication (JWT)
- 1-on-1 and group chat
- Real-time messaging via WebSocket
- Typing indicators
- Message delivery and read status
- Online/offline presence tracking
- Message edit and delete
- Notifications for offline users

## Tech Stack

- **Backend**: Golang (net/http, gorilla/websocket)
- **Database**: MongoDB
- **Authentication**: JWT
- **Frontend**: React + TypeScript (coming soon)

## Prerequisites

- Go 1.21 or higher
- MongoDB 6.0 or higher
- Docker (optional, for running MongoDB)

## Getting Started

### 1. Clone the repository

```bash
git clone <repository-url>
cd realtime-chat-system
```

### 2. Install dependencies

```bash
make deps
```

### 3. Setup MongoDB

Option A: Using Docker
```bash
make docker-up
```

Option B: Install MongoDB locally
- Follow instructions at https://www.mongodb.com/docs/manual/installation/

### 4. Configure environment variables

```bash
cp .env.example .env
# Edit .env with your configuration
```

### 5. Run the application

```bash
make run
```

The server will start on `http://localhost:8080`

## Development

### Project Structure

```
.
├── cmd/
│   └── server/          # Application entry point
├── config/              # Configuration management
├── internal/
│   ├── models/          # Data models
│   ├── repository/      # Data access layer
│   ├── service/         # Business logic layer
│   └── handler/         # HTTP and WebSocket handlers
├── pkg/
│   ├── database/        # Database utilities
│   ├── errors/          # Error handling
│   └── logger/          # Logging utilities
└── go.mod
```

### Available Commands

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
- `GET /api/rooms/:id/messages` - Get chat history

### WebSocket
- `WS /ws?token=<jwt_token>` - WebSocket connection

## WebSocket Message Types

- `message` - Send/receive chat messages
- `typing` - Typing indicator
- `delivered` - Message delivery acknowledgment
- `read` - Message read acknowledgment
- `presence` - User online/offline status
- `edit` - Message edit
- `delete` - Message delete

## License

MIT
