# Implementation Plan: Real-time Chat System

## Overview

Implementation จะแบ่งเป็นขั้นตอนหลักๆ ดังนี้:
1. Setup project structure และ database connection
2. Implement authentication system (User, JWT)
3. Implement room management (Direct และ Group chat)
4. Implement message handling และ WebSocket
5. Implement presence tracking
6. Implement notification system
7. Integration และ testing

แต่ละ task จะมี property-based tests และ unit tests เพื่อให้มั่นใจว่าระบบทำงานถูกต้อง

## Tasks

- [x] 1. Project Setup และ Database Connection
  - สร้าง project structure ตาม layered architecture (handler, service, repository)
  - Setup MongoDB connection และ configuration
  - สร้าง base types และ error handling utilities
  - Setup testing framework (gopter สำหรับ property-based testing)
  - _Requirements: 11.5_

- [-] 2. Implement User Model และ Repository
  - [ ] 2.1 สร้าง User model และ validation functions
    - สร้าง User struct พร้อม BSON tags
    - เขียน validation functions (username, email, password format)
    - _Requirements: 1.1, 1.2_

  - [ ]* 2.2 Write property test for User validation
    - **Property 1: Valid registration creates user account**
    - **Property 2: Invalid registration is rejected**
    - **Validates: Requirements 1.1, 1.2**

  - [ ] 2.3 Implement UserRepository
    - เขียน Create, FindByID, FindByUsername, FindByEmail methods
    - Setup MongoDB indexes สำหรับ username และ email
    - _Requirements: 1.1, 1.3_

  - [ ]* 2.4 Write unit tests for UserRepository
    - Test duplicate username/email handling
    - Test not found scenarios
    - _Requirements: 1.1, 1.2_

- [ ] 3. Implement Authentication Service
  - [ ] 3.1 Implement password hashing และ comparison
    - ใช้ bcrypt สำหรับ hash passwords
    - เขียน HashPassword และ ComparePassword functions
    - _Requirements: 1.5_

  - [ ]* 3.2 Write property test for password hashing
    - **Property 5: Password hashing**
    - **Validates: Requirements 1.5**

  - [ ] 3.3 Implement JWT token generation และ validation
    - สร้าง Claims struct
    - เขียน GenerateToken และ ValidateToken functions
    - กำหนด token expiration time
    - _Requirements: 1.6, 11.4_

  - [ ]* 3.4 Write property test for JWT tokens
    - **Property 6: JWT token contains user info and expiration**
    - **Property 7: JWT tokens are stateless**
    - **Validates: Requirements 1.6, 11.4**

  - [ ] 3.5 Implement AuthService (Register และ Login)
    - เขียน Register method พร้อม validation
    - เขียน Login method พร้อม credential verification
    - _Requirements: 1.1, 1.2, 1.3, 1.4_

  - [ ]* 3.6 Write property tests for AuthService
    - **Property 3: Login round-trip with valid credentials**
    - **Property 4: Invalid credentials are rejected**
    - **Validates: Requirements 1.3, 1.4**

- [ ] 4. Implement HTTP Handlers สำหรับ Authentication
  - [ ] 4.1 สร้าง HTTP handlers สำหรับ register และ login
    - POST /api/auth/register endpoint
    - POST /api/auth/login endpoint
    - Error response formatting
    - _Requirements: 1.1, 1.3_

  - [ ] 4.2 Implement JWT middleware สำหรับ authentication
    - Extract และ validate JWT token from request header
    - Inject user claims into request context
    - _Requirements: 2.1, 2.2_

  - [ ]* 4.3 Write integration tests for auth endpoints
    - Test registration flow
    - Test login flow
    - Test invalid inputs
    - _Requirements: 1.1, 1.2, 1.3, 1.4_

- [ ] 5. Checkpoint - Authentication System Complete
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 6. Implement Room Model และ Repository
  - [ ] 6.1 สร้าง Room model และ validation
    - สร้าง Room struct (direct และ group types)
    - Validation สำหรับ member count และ room type
    - _Requirements: 3.1, 4.1_

  - [ ] 6.2 Implement RoomRepository
    - เขียน Create, FindByID, FindByMembers, FindUserRooms methods
    - เขียน UpdateMembers method
    - Setup MongoDB indexes
    - _Requirements: 3.1, 4.1, 4.5, 4.6_

  - [ ]* 6.3 Write unit tests for RoomRepository
    - Test room creation
    - Test member queries
    - Test member updates
    - _Requirements: 3.1, 4.1, 4.5, 4.6_

- [ ] 7. Implement Room Service
  - [ ] 7.1 Implement RoomService methods
    - CreateDirectRoom (check existing room first)
    - CreateGroupRoom (validate members และ name)
    - GetRoom, GetUserRooms
    - AddMembers, RemoveMembers
    - IsMember (authorization check)
    - _Requirements: 3.1, 4.1, 4.5, 4.6, 10.1, 10.2, 10.3_

  - [ ]* 7.2 Write property tests for RoomService
    - **Property 12: First message creates direct room**
    - **Property 13: Room membership authorization**
    - **Property 14: Group creation with valid inputs**
    - **Property 15: Adding members updates room and notifies**
    - **Property 16: Removing members updates room and notifies**
    - **Property 17: Room member lists are accurate**
    - **Validates: Requirements 3.1, 3.2, 4.1, 4.2, 4.5, 4.6, 8.1, 10.1, 10.2, 10.3, 10.4**

- [ ] 8. Implement HTTP Handlers สำหรับ Rooms
  - [ ] 8.1 สร้าง room management endpoints
    - GET /api/rooms (list user's rooms)
    - POST /api/rooms (create group room)
    - POST /api/rooms/:id/members (add members)
    - DELETE /api/rooms/:id/members (remove members)
    - _Requirements: 4.1, 4.5, 4.6_

  - [ ]* 8.2 Write integration tests for room endpoints
    - Test room creation
    - Test member management
    - Test authorization
    - _Requirements: 4.1, 4.5, 4.6, 10.1_

- [ ] 9. Checkpoint - Room Management Complete
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 10. Implement Message Model และ Repository
  - [ ] 10.1 สร้าง Message model พร้อม Status tracking
    - สร้าง Message struct พร้อม status map
    - สร้าง Status struct (delivered, read timestamps)
    - _Requirements: 3.3, 5.1, 5.2, 5.5_

  - [ ] 10.2 Implement MessageRepository
    - เขียน Create, FindByID, FindByRoom methods
    - เขียน Update และ Delete methods
    - Support pagination ใน FindByRoom
    - Setup MongoDB indexes
    - _Requirements: 3.3, 8.2, 8.3, 9.1, 9.2_

  - [ ]* 10.3 Write unit tests for MessageRepository
    - Test message creation
    - Test pagination
    - Test updates และ deletes
    - _Requirements: 3.3, 8.2, 8.3, 9.1, 9.2_

- [ ] 11. Implement Message Service
  - [ ] 11.1 Implement core message operations
    - SendMessage (verify membership, save to DB)
    - GetMessages (with pagination)
    - EditMessage (verify ownership)
    - DeleteMessage (verify ownership)
    - _Requirements: 3.2, 3.3, 8.1, 8.2, 8.3, 9.1, 9.2, 9.5_

  - [ ] 11.2 Implement message status tracking
    - UpdateDeliveryStatus
    - UpdateReadStatus
    - _Requirements: 5.1, 5.2, 5.5_

  - [ ]* 11.3 Write property tests for MessageService
    - **Property 18: Message persistence with metadata**
    - **Property 20: Offline message storage**
    - **Property 21: Message edit updates content and marks edited**
    - **Property 22: Message delete marks as deleted**
    - **Property 25: Only owner can edit or delete**
    - **Property 26: Delivery status update on receipt**
    - **Property 27: Read status update on view**
    - **Property 30: Status timestamps are stored**
    - **Property 37: History ordered by timestamp**
    - **Property 38: Pagination returns correct subset**
    - **Property 39: History includes complete metadata**
    - **Validates: Requirements 3.2, 3.3, 3.5, 5.1, 5.2, 5.5, 8.1, 8.2, 8.3, 8.5, 9.1, 9.2, 9.5**

- [ ] 12. Implement HTTP Handler สำหรับ Chat History
  - [ ] 12.1 สร้าง chat history endpoint
    - GET /api/rooms/:id/messages (with pagination)
    - Verify user is room member
    - _Requirements: 8.1, 8.2, 8.3, 8.5_

  - [ ]* 12.2 Write integration tests for chat history
    - Test history retrieval
    - Test pagination
    - Test authorization
    - _Requirements: 8.1, 8.2, 8.3_

- [ ] 13. Checkpoint - Message System Complete
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 14. Implement WebSocket Connection Management
  - [ ] 14.1 สร้าง WebSocket connection pool
    - สร้าง ConnectionPool struct พร้อม thread-safe operations
    - สร้าง Connection struct พร้อม send channel
    - Implement Add, Remove, Get, Broadcast methods
    - _Requirements: 2.5_

  - [ ]* 14.2 Write property test for connection pool
    - **Property 11: Connection pool maintains accurate mapping**
    - **Validates: Requirements 2.5**

  - [ ] 14.3 Implement WebSocket handler พร้อม JWT authentication
    - Upgrade HTTP connection to WebSocket
    - Validate JWT token from query parameter
    - Add connection to pool
    - Handle connection lifecycle (read/write loops)
    - _Requirements: 2.1, 2.2_

  - [ ]* 14.4 Write property tests for WebSocket authentication
    - **Property 8: Valid token accepts connection and marks online**
    - **Property 9: Invalid token rejects connection**
    - **Validates: Requirements 2.1, 2.2, 2.3**

- [ ] 15. Implement Presence Service
  - [ ] 15.1 สร้าง PresenceService พร้อม in-memory store
    - สร้าง PresenceStore struct
    - Implement SetOnline, SetOffline methods
    - Implement IsOnline, GetLastSeen methods
    - Implement Heartbeat mechanism
    - _Requirements: 2.3, 2.4, 7.1, 7.2, 7.5_

  - [ ]* 15.2 Write property tests for PresenceService
    - **Property 10: Disconnection marks user offline with timestamp**
    - **Property 36: Heartbeat detects stale connections**
    - **Validates: Requirements 2.4, 7.5**

  - [ ] 15.3 Integrate presence tracking with WebSocket
    - Mark user online on connection
    - Mark user offline on disconnection
    - Start heartbeat monitoring goroutine
    - _Requirements: 2.3, 2.4_

  - [ ]* 15.4 Write property test for presence broadcast
    - **Property 35: Presence change broadcast**
    - **Validates: Requirements 7.3**

- [ ] 16. Implement WebSocket Message Handling
  - [ ] 16.1 สร้าง WebSocket message types และ routing
    - สร้าง WSMessage struct
    - สร้าง ChatMessage, TypingIndicator, StatusUpdate structs
    - Implement message type routing
    - _Requirements: 3.4, 4.4_

  - [ ] 16.2 Implement chat message handling via WebSocket
    - Handle incoming chat messages
    - Verify room membership
    - Save message via MessageService
    - Broadcast to room members
    - _Requirements: 3.2, 3.3, 3.4, 4.2, 4.3, 4.4_

  - [ ]* 16.3 Write property tests for message broadcast
    - **Property 19: Message broadcast to online members**
    - **Property 23: Edit broadcast to all members**
    - **Property 24: Delete broadcast to all members**
    - **Validates: Requirements 3.4, 4.4, 9.3, 9.4**

  - [ ] 16.4 Implement typing indicator handling
    - Handle typing start/stop events
    - Broadcast to online room members only
    - Do not persist to database
    - _Requirements: 6.1, 6.2, 6.3, 6.4_

  - [ ]* 16.5 Write property tests for typing indicators
    - **Property 31: Typing indicator broadcast to online members**
    - **Property 32: Stop typing broadcast to online members**
    - **Property 33: Typing indicators are not persisted**
    - **Property 34: Typing indicators only to online members**
    - **Validates: Requirements 6.1, 6.2, 6.3, 6.4**

  - [ ] 16.6 Implement message status updates via WebSocket
    - Handle delivery acknowledgments
    - Handle read receipts
    - Update message status in database
    - Broadcast status to sender
    - _Requirements: 5.1, 5.2, 5.3, 5.4_

  - [ ]* 16.7 Write property tests for status updates
    - **Property 28: Delivery status broadcast to sender**
    - **Property 29: Read status broadcast to sender**
    - **Validates: Requirements 5.3, 5.4**

- [ ] 17. Checkpoint - WebSocket System Complete
  - Ensure all tests pass, ask the user if questions arise.

- [ ] 18. Implement Notification System
  - [ ] 18.1 สร้าง Notification model และ repository
    - สร้าง Notification struct
    - Implement NotificationRepository
    - Setup MongoDB indexes
    - _Requirements: 12.1, 12.5_

  - [ ] 18.2 Implement NotificationService
    - CreateNotification (for offline users)
    - GetPendingNotifications
    - MarkAsRead
    - _Requirements: 12.1, 12.2, 12.3, 12.4, 12.5_

  - [ ]* 18.3 Write property tests for NotificationService
    - **Property 40: Notification created for offline recipients**
    - **Property 41: Pending notifications retrieved on login**
    - **Property 42: Notification type is correctly set**
    - **Property 43: Notifications marked read on view**
    - **Property 44: Users can retrieve notification history**
    - **Validates: Requirements 12.1, 12.2, 12.3, 12.4, 12.5**

  - [ ] 18.3 Integrate notifications with message sending
    - Check recipient online status
    - Create notifications for offline recipients
    - _Requirements: 12.1_

- [ ] 19. Implement Authorization และ Error Handling
  - [ ] 19.1 Centralize authorization checks
    - Ensure all operations verify permissions
    - Return consistent error responses
    - _Requirements: 10.1, 10.2, 10.3, 10.5_

  - [ ]* 19.2 Write property test for authorization
    - **Property 45: Unauthorized access is rejected with error**
    - **Validates: Requirements 10.5**

  - [ ] 19.2 Implement comprehensive error handling
    - Create error types for each category
    - Implement error response formatting
    - Add error logging
    - _Requirements: 1.2, 1.4, 3.2, 10.5_

- [ ] 20. Final Integration และ End-to-End Testing
  - [ ] 20.1 Wire all components together
    - Setup dependency injection
    - Initialize all services
    - Start HTTP และ WebSocket servers
    - _Requirements: All_

  - [ ]* 20.2 Write end-to-end integration tests
    - Test complete user registration → login → chat flow
    - Test group chat creation และ messaging
    - Test presence tracking
    - Test notifications
    - _Requirements: All_

  - [ ] 20.3 Performance และ load testing
    - Test concurrent connections
    - Test message throughput
    - Test database query performance
    - _Requirements: 11.1, 11.2, 11.3_

- [ ] 21. Final Checkpoint - System Complete
  - Ensure all tests pass, ask the user if questions arise.
  - Verify all requirements are implemented
  - Review code quality และ documentation

## Frontend Implementation Tasks

- [ ] 22. Frontend Project Setup
  - [ ] 22.1 Initialize React + TypeScript + Vite project
    - Setup Vite configuration
    - Install dependencies (React, TypeScript, Zustand, Tailwind CSS, Axios)
    - Configure TypeScript (tsconfig.json)
    - Setup Tailwind CSS
    - _Requirements: Frontend setup_

  - [ ] 22.2 Create project structure
    - สร้าง folder structure (components, hooks, services, store, types, pages, utils)
    - Setup routing (React Router)
    - _Requirements: Frontend setup_

- [ ] 23. Implement TypeScript Types และ Interfaces
  - [ ] 23.1 สร้าง type definitions
    - User, Room, Message, Notification types
    - WebSocket message types
    - API response types
    - _Requirements: Type safety_

- [ ] 24. Implement State Management (Zustand)
  - [ ] 24.1 Create auth store
    - Login, logout, user state management
    - Token persistence in localStorage
    - _Requirements: 1.1, 1.3_

  - [ ] 24.2 Create chat store
    - Messages state management
    - Add, update, delete message operations
    - Current room tracking
    - _Requirements: 3.3, 9.1, 9.2_

  - [ ] 24.3 Create room store
    - Rooms list management
    - Current room selection
    - _Requirements: 3.1, 4.1_

  - [ ] 24.4 Create presence store
    - Online users tracking
    - Last seen timestamps
    - _Requirements: 2.3, 2.4, 7.1, 7.2_

- [ ] 25. Implement API Service
  - [ ] 25.1 Setup Axios instance
    - Base URL configuration
    - Request interceptor สำหรับ JWT token
    - Response interceptor สำหรับ error handling
    - _Requirements: 1.3, 2.1_

  - [ ] 25.2 Implement auth API
    - Register endpoint
    - Login endpoint
    - _Requirements: 1.1, 1.3_

  - [ ] 25.3 Implement room API
    - Get rooms endpoint
    - Create group endpoint
    - Add/remove members endpoints
    - _Requirements: 4.1, 4.5, 4.6_

  - [ ] 25.4 Implement message API
    - Get messages endpoint (with pagination)
    - _Requirements: 8.2, 8.3_

- [ ] 26. Implement WebSocket Service
  - [ ] 26.1 Create WebSocket service class
    - Connection management
    - Reconnection logic with exponential backoff
    - Message routing system
    - Heartbeat mechanism
    - _Requirements: 2.1, 2.2, 7.5_

  - [ ] 26.2 Implement message handlers
    - Handle incoming messages
    - Handle typing indicators
    - Handle presence updates
    - Handle status updates (delivered/read)
    - Handle edit/delete events
    - _Requirements: 3.4, 5.3, 5.4, 6.1, 6.2, 7.3, 9.3, 9.4_

- [ ] 27. Implement Custom Hooks
  - [ ] 27.1 Create useWebSocket hook
    - Connect/disconnect WebSocket
    - Connection status tracking
    - _Requirements: 2.1, 2.2_

  - [ ] 27.2 Create useAuth hook
    - Login, logout, register functions
    - Auth state management
    - _Requirements: 1.1, 1.3_

  - [ ] 27.3 Create useChat hook
    - Send message function
    - Typing indicator management
    - Mark as read function
    - Message state management
    - _Requirements: 3.3, 3.4, 5.2, 6.1, 6.2_

  - [ ] 27.4 Create usePresence hook
    - Subscribe to presence updates
    - Check online status
    - _Requirements: 2.3, 2.4, 7.3_

  - [ ] 27.5 Create useNotifications hook
    - Fetch notifications
    - Mark as read
    - _Requirements: 12.2, 12.4_

- [ ] 28. Implement Authentication Pages
  - [ ] 28.1 Create LoginPage component
    - Login form with validation
    - Error handling
    - Redirect after successful login
    - _Requirements: 1.3, 1.4_

  - [ ] 28.2 Create RegisterPage component
    - Registration form with validation
    - Password strength indicator
    - Error handling
    - _Requirements: 1.1, 1.2_

  - [ ]* 28.3 Add form validation
    - Email format validation
    - Password strength validation
    - Username validation
    - _Requirements: 1.1, 1.2_

- [ ] 29. Implement Chat Components
  - [ ] 29.1 Create RoomList component
    - Display user's rooms
    - Show last message preview
    - Show unread count
    - Room selection
    - _Requirements: 3.1, 4.1_

  - [ ] 29.2 Create ChatWindow component
    - Message list display
    - Message input
    - Typing indicator
    - Auto-scroll to bottom
    - _Requirements: 3.3, 3.4, 6.1, 6.2_

  - [ ] 29.3 Create MessageList component
    - Display messages in chronological order
    - Infinite scroll for history loading
    - Date separators
    - _Requirements: 8.2, 8.3_

  - [ ] 29.4 Create MessageItem component
    - Display message content
    - Show sender info และ avatar
    - Show timestamp
    - Show edited indicator
    - Show delivery/read status
    - Online status indicator
    - _Requirements: 3.3, 5.1, 5.2, 7.3, 9.1_

  - [ ] 29.5 Create MessageInput component
    - Text input with auto-resize
    - Send button
    - Typing indicator trigger
    - _Requirements: 3.3, 6.1_

  - [ ] 29.6 Create TypingIndicator component
    - Display typing users
    - Animated dots
    - _Requirements: 6.1, 6.2_

- [ ] 30. Implement Room Management Components
  - [ ] 30.1 Create CreateRoomModal component
    - Group name input
    - Member selection
    - Create group button
    - _Requirements: 4.1_

  - [ ] 30.2 Create RoomItem component
    - Room name/avatar
    - Last message preview
    - Unread badge
    - Online status for direct chats
    - _Requirements: 3.1, 4.1, 7.3_

  - [ ]* 30.3 Create RoomSettings component
    - Add/remove members
    - Leave group
    - _Requirements: 4.5, 4.6_

- [ ] 31. Implement Common Components
  - [ ] 31.1 Create Avatar component
    - User avatar display
    - Fallback to initials
    - Size variants
    - _Requirements: UI_

  - [ ] 31.2 Create OnlineStatus component
    - Online/offline indicator
    - Last seen display
    - _Requirements: 7.1, 7.2, 7.3_

  - [ ] 31.3 Create Notification component
    - Toast notifications
    - Notification list
    - Mark as read
    - _Requirements: 12.1, 12.2, 12.4_

- [ ] 32. Implement Main Chat Page
  - [ ] 32.1 Create ChatPage layout
    - Sidebar with room list
    - Main chat area
    - Responsive design (mobile/desktop)
    - _Requirements: All chat features_

  - [ ] 32.2 Integrate all components
    - Wire up room selection
    - Wire up message sending
    - Wire up presence tracking
    - _Requirements: All chat features_

- [ ] 33. Implement Real-time Features
  - [ ] 33.1 Integrate WebSocket with chat
    - Connect on login
    - Subscribe to room events
    - Handle incoming messages
    - _Requirements: 2.1, 3.4, 4.4_

  - [ ] 33.2 Implement optimistic updates
    - Show sent messages immediately
    - Update UI before server confirmation
    - Handle failures gracefully
    - _Requirements: 3.3, 3.4_

  - [ ] 33.3 Implement message status tracking
    - Show delivery status
    - Show read status
    - Update status in real-time
    - _Requirements: 5.1, 5.2, 5.3, 5.4_

- [ ] 34. Implement Error Handling และ Loading States
  - [ ] 34.1 Add loading indicators
    - Page loading
    - Message sending
    - History loading
    - _Requirements: UX_

  - [ ] 34.2 Add error handling
    - Network errors
    - Authentication errors
    - Validation errors
    - Display error messages
    - _Requirements: 1.2, 1.4, 10.5_

  - [ ] 34.3 Add retry mechanisms
    - Retry failed message sends
    - Retry WebSocket connection
    - _Requirements: 2.1, 3.3_

- [ ] 35. Styling และ UI Polish
  - [ ] 35.1 Implement responsive design
    - Mobile layout
    - Tablet layout
    - Desktop layout
    - _Requirements: UX_

  - [ ] 35.2 Add animations และ transitions
    - Message animations
    - Typing indicator animation
    - Smooth scrolling
    - _Requirements: UX_

  - [ ]* 35.3 Add dark mode support
    - Theme toggle
    - Dark color scheme
    - _Requirements: UX_

- [ ] 36. Frontend Testing
  - [ ]* 36.1 Write component tests
    - Test key components with React Testing Library
    - Test user interactions
    - _Requirements: Testing_

  - [ ]* 36.2 Write integration tests
    - Test complete user flows
    - Test WebSocket integration
    - _Requirements: Testing_

- [ ] 37. Final Frontend Checkpoint
  - Ensure all features work correctly
  - Test on different browsers
  - Test responsive design
  - Ask the user if questions arise

## Notes

- Tasks marked with `*` are optional and can be skipped for faster MVP
- Each property test should run minimum 100 iterations
- Use gopter for property-based testing in Go
- Co-locate test files with source files using `_test.go` suffix
- Property tests should reference design document properties in comments
- Unit tests focus on edge cases and error conditions
- Integration tests verify component interactions
- All tests must pass before moving to next checkpoint
