# Chat App Frontend

React + TypeScript + Vite frontend for the real-time chat system.

## Features

- User authentication (login/register)
- Real-time messaging via WebSocket
- Room management
- Message history
- Presence tracking
- Responsive design

## Tech Stack

- React 18
- TypeScript
- Vite
- Zustand (state management)
- Axios (HTTP client)
- Tailwind CSS
- React Router

## Getting Started

### Prerequisites

- Node.js 18+ 
- Backend server running on http://localhost:8080

### Installation

```bash
npm install
```

### Configuration

Create `.env` file:

```env
VITE_API_URL=http://localhost:8080
VITE_WS_URL=ws://localhost:8080
```

### Development

```bash
npm run dev
```

The app will be available at http://localhost:5173

### Build

```bash
npm run build
```

### Preview Production Build

```bash
npm run preview
```

## Project Structure

```
src/
├── components/      # React components
│   ├── MessageItem.tsx
│   ├── MessageList.tsx
│   ├── MessageInput.tsx
│   └── RoomList.tsx
├── hooks/          # Custom hooks
│   └── useWebSocket.ts
├── pages/          # Page components
│   ├── LoginPage.tsx
│   ├── RegisterPage.tsx
│   └── ChatPage.tsx
├── services/       # API and WebSocket services
│   ├── api.ts
│   └── websocket.ts
├── store/          # Zustand stores
│   ├── authStore.ts
│   ├── chatStore.ts
│   ├── roomStore.ts
│   └── presenceStore.ts
├── types/          # TypeScript types
│   └── index.ts
├── App.tsx         # Main app component
└── main.tsx        # Entry point
```

## Usage

### Login/Register

1. Navigate to http://localhost:5173
2. Register a new account or login with existing credentials
3. You'll be redirected to the chat page

### Chat

1. Select a room from the left sidebar
2. Type a message in the input field
3. Press Enter or click Send
4. Messages appear in real-time

## State Management

The app uses Zustand for state management with the following stores:

- **authStore**: User authentication state
- **chatStore**: Messages and chat operations
- **roomStore**: Rooms list and selection
- **presenceStore**: Online/offline user tracking

## WebSocket Integration

The app connects to the WebSocket server automatically after login:

- Real-time message delivery
- Presence updates
- Message edit/delete notifications
- Auto-reconnect with exponential backoff

## API Integration

All HTTP requests go through the Axios instance in `services/api.ts`:

- Automatic JWT token injection
- Error handling
- Auto-redirect on 401 (unauthorized)

## License

MIT
