#!/bin/bash

# Test WebSocket functionality
# Requires: wscat (npm install -g wscat)

BASE_URL="http://localhost:8080"
WS_URL="ws://localhost:8080"

echo "==================================="
echo "WebSocket System Test"
echo "==================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Register two users
echo -e "${YELLOW}Step 1: Register two users${NC}"
echo "Registering user1..."
USER1_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "alice",
    "email": "alice@example.com",
    "password": "password123"
  }')

echo "$USER1_RESPONSE" | jq '.'
USER1_TOKEN=$(echo "$USER1_RESPONSE" | jq -r '.data.token')

echo ""
echo "Registering user2..."
USER2_RESPONSE=$(curl -s -X POST "$BASE_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d '{
    "username": "bob",
    "email": "bob@example.com",
    "password": "password123"
  }')

echo "$USER2_RESPONSE" | jq '.'
USER2_TOKEN=$(echo "$USER2_RESPONSE" | jq -r '.data.token')

if [ "$USER1_TOKEN" == "null" ] || [ "$USER2_TOKEN" == "null" ]; then
  echo -e "${RED}Failed to register users${NC}"
  exit 1
fi

echo -e "${GREEN}✓ Users registered successfully${NC}"
echo ""

# Step 2: Create a direct room
echo -e "${YELLOW}Step 2: Create a direct room${NC}"
USER2_ID=$(echo "$USER2_RESPONSE" | jq -r '.data.user.id')

ROOM_RESPONSE=$(curl -s -X POST "$BASE_URL/api/rooms" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $USER1_TOKEN" \
  -d "{
    \"type\": \"direct\",
    \"members\": [\"$USER2_ID\"]
  }")

echo "$ROOM_RESPONSE" | jq '.'
ROOM_ID=$(echo "$ROOM_RESPONSE" | jq -r '.data.id')

if [ "$ROOM_ID" == "null" ]; then
  echo -e "${RED}Failed to create room${NC}"
  exit 1
fi

echo -e "${GREEN}✓ Room created successfully${NC}"
echo ""

# Step 3: WebSocket connection instructions
echo -e "${YELLOW}Step 3: Test WebSocket connections${NC}"
echo ""
echo "To test WebSocket, open two terminal windows and run:"
echo ""
echo -e "${GREEN}Terminal 1 (Alice):${NC}"
echo "wscat -c \"$WS_URL/ws?token=$USER1_TOKEN\""
echo ""
echo -e "${GREEN}Terminal 2 (Bob):${NC}"
echo "wscat -c \"$WS_URL/ws?token=$USER2_TOKEN\""
echo ""
echo "Then send messages:"
echo ""
echo -e "${GREEN}Send a chat message:${NC}"
echo "{\"type\":\"message\",\"roomId\":\"$ROOM_ID\",\"payload\":{\"content\":\"Hello!\"}}"
echo ""
echo -e "${GREEN}Send typing indicator:${NC}"
echo "{\"type\":\"typing\",\"roomId\":\"$ROOM_ID\",\"payload\":{\"isTyping\":true}}"
echo ""
echo -e "${GREEN}Send heartbeat:${NC}"
echo "{\"type\":\"heartbeat\"}"
echo ""
echo -e "${GREEN}Mark message as read:${NC}"
echo "{\"type\":\"read\",\"roomId\":\"$ROOM_ID\",\"payload\":{\"messageId\":\"<message_id>\"}}"
echo ""
echo -e "${GREEN}Edit message:${NC}"
echo "{\"type\":\"edit\",\"roomId\":\"$ROOM_ID\",\"payload\":{\"messageId\":\"<message_id>\",\"content\":\"Updated!\"}}"
echo ""
echo -e "${GREEN}Delete message:${NC}"
echo "{\"type\":\"delete\",\"roomId\":\"$ROOM_ID\",\"payload\":{\"messageId\":\"<message_id>\"}}"
echo ""

# Step 4: Test REST API for messages
echo -e "${YELLOW}Step 4: Send message via REST API${NC}"
MESSAGE_RESPONSE=$(curl -s -X POST "$BASE_URL/api/rooms/$ROOM_ID/messages" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $USER1_TOKEN" \
  -d '{
    "content": "Hello from REST API!"
  }')

echo "$MESSAGE_RESPONSE" | jq '.'
MESSAGE_ID=$(echo "$MESSAGE_RESPONSE" | jq -r '.data.id')

if [ "$MESSAGE_ID" == "null" ]; then
  echo -e "${RED}Failed to send message${NC}"
  exit 1
fi

echo -e "${GREEN}✓ Message sent successfully${NC}"
echo ""

# Step 5: Get chat history
echo -e "${YELLOW}Step 5: Get chat history${NC}"
HISTORY_RESPONSE=$(curl -s -X GET "$BASE_URL/api/rooms/$ROOM_ID/messages?limit=10&offset=0" \
  -H "Authorization: Bearer $USER1_TOKEN")

echo "$HISTORY_RESPONSE" | jq '.'
echo -e "${GREEN}✓ Chat history retrieved${NC}"
echo ""

# Summary
echo "==================================="
echo -e "${GREEN}Test Summary${NC}"
echo "==================================="
echo "Room ID: $ROOM_ID"
echo "User1 (Alice) Token: $USER1_TOKEN"
echo "User2 (Bob) Token: $USER2_TOKEN"
echo "Message ID: $MESSAGE_ID"
echo ""
echo -e "${YELLOW}Next Steps:${NC}"
echo "1. Connect via WebSocket using the commands above"
echo "2. Send messages and observe real-time updates"
echo "3. Test typing indicators"
echo "4. Test presence updates (connect/disconnect)"
echo "5. Test message status updates"
echo ""
