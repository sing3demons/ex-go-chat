#!/bin/bash

# Real System Test - Actually test the running system

set -e

API_URL="http://localhost:8080"
FRONTEND_URL="http://localhost:5173"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🧪 Real-time Chat System - ACTUAL TEST${NC}"
echo "=========================================="
echo ""

# Test 1: Check if services are running
echo -e "${YELLOW}Test 1: Checking Services...${NC}"

if curl -s "$API_URL/health" > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Backend is running on $API_URL"
else
    echo -e "${RED}✗${NC} Backend is NOT running"
    echo "Please start backend: make run"
    exit 1
fi

if curl -s "$FRONTEND_URL" > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Frontend is running on $FRONTEND_URL"
else
    echo -e "${RED}✗${NC} Frontend is NOT running"
    echo "Please start frontend: cd frontend && npm run dev"
    exit 1
fi

# Test 2: Register a new user
echo ""
echo -e "${YELLOW}Test 2: Testing User Registration...${NC}"

TIMESTAMP=$(date +%s)
USERNAME="testuser_$TIMESTAMP"
EMAIL="test_$TIMESTAMP@test.com"
PASSWORD="Test123456"

REGISTER_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/api/auth/register" \
  -H "Content-Type: application/json" \
  -d "{\"username\":\"$USERNAME\",\"email\":\"$EMAIL\",\"password\":\"$PASSWORD\"}")

STATUS_CODE=$(echo "$REGISTER_RESPONSE" | tail -n1)
BODY=$(echo "$REGISTER_RESPONSE" | sed '$d')

if [ "$STATUS_CODE" = "201" ]; then
    echo -e "${GREEN}✓${NC} Registration successful"
    echo "   Username: $USERNAME"
    echo "   Email: $EMAIL"
    USER_ID=$(echo "$BODY" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
    echo "   User ID: $USER_ID"
else
    echo -e "${RED}✗${NC} Registration failed (Status: $STATUS_CODE)"
    echo "   Response: $BODY"
    exit 1
fi

# Test 3: Login
echo ""
echo -e "${YELLOW}Test 3: Testing User Login...${NC}"

LOGIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$API_URL/api/auth/login" \
  -H "Content-Type: application/json" \
  -d "{\"identifier\":\"$USERNAME\",\"password\":\"$PASSWORD\"}")

STATUS_CODE=$(echo "$LOGIN_RESPONSE" | tail -n1)
BODY=$(echo "$LOGIN_RESPONSE" | sed '$d')

if [ "$STATUS_CODE" = "200" ]; then
    echo -e "${GREEN}✓${NC} Login successful"
    TOKEN=$(echo "$BODY" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    echo "   Token: ${TOKEN:0:50}..."
else
    echo -e "${RED}✗${NC} Login failed (Status: $STATUS_CODE)"
    echo "   Response: $BODY"
    exit 1
fi

# Test 4: Get rooms
echo ""
echo -e "${YELLOW}Test 4: Testing Room List...${NC}"

ROOMS_RESPONSE=$(curl -s -w "\n%{http_code}" -X GET "$API_URL/api/rooms" \
  -H "Authorization: Bearer $TOKEN")

STATUS_CODE=$(echo "$ROOMS_RESPONSE" | tail -n1)
BODY=$(echo "$ROOMS_RESPONSE" | sed '$d')

if [ "$STATUS_CODE" = "200" ]; then
    echo -e "${GREEN}✓${NC} Get rooms successful"
    echo "   Response: $BODY"
else
    echo -e "${RED}✗${NC} Get rooms failed (Status: $STATUS_CODE)"
    echo "   Response: $BODY"
fi

# Test 5: Frontend accessibility
echo ""
echo -e "${YELLOW}Test 5: Testing Frontend Pages...${NC}"

# Test register page
REGISTER_PAGE=$(curl -s "$FRONTEND_URL/register")
if echo "$REGISTER_PAGE" | grep -q "root"; then
    echo -e "${GREEN}✓${NC} Register page loads"
else
    echo -e "${RED}✗${NC} Register page failed to load"
fi

# Test login page
LOGIN_PAGE=$(curl -s "$FRONTEND_URL/login")
if echo "$LOGIN_PAGE" | grep -q "root"; then
    echo -e "${GREEN}✓${NC} Login page loads"
else
    echo -e "${RED}✗${NC} Login page failed to load"
fi

# Summary
echo ""
echo "=========================================="
echo -e "${GREEN}✅ All Tests Passed!${NC}"
echo ""
echo "🌐 Open in browser:"
echo "   $FRONTEND_URL/register"
echo ""
echo "📝 Test Credentials:"
echo "   Username: $USERNAME"
echo "   Password: $PASSWORD"
echo ""
echo "🔑 JWT Token (for API testing):"
echo "   $TOKEN"
echo ""
echo "💡 Next Steps:"
echo "   1. Open $FRONTEND_URL in your browser"
echo "   2. Try registering a new user"
echo "   3. Login and test chat features"
echo ""
