#!/bin/bash

# Real-time Chat System - Complete System Test
# This script tests all major features of the application

set -e

API_URL="http://localhost:8080"
FRONTEND_URL="http://localhost:5173"

echo "🧪 Real-time Chat System - System Test"
echo "======================================"
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Test counter
TESTS_PASSED=0
TESTS_FAILED=0

# Function to print test result
print_result() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓ PASS${NC}: $2"
        ((TESTS_PASSED++))
    else
        echo -e "${RED}✗ FAIL${NC}: $2"
        ((TESTS_FAILED++))
    fi
}

# Function to test endpoint
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local token=$4
    local expected_status=$5
    
    if [ -n "$token" ]; then
        response=$(curl -s -w "\n%{http_code}" -X $method "$API_URL$endpoint" \
            -H "Content-Type: application/json" \
            -H "Authorization: Bearer $token" \
            -d "$data" 2>/dev/null || echo "000")
    else
        response=$(curl -s -w "\n%{http_code}" -X $method "$API_URL$endpoint" \
            -H "Content-Type: application/json" \
            -d "$data" 2>/dev/null || echo "000")
    fi
    
    status_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')
    
    if [ "$status_code" = "$expected_status" ]; then
        echo "$body"
        return 0
    else
        echo "Expected $expected_status, got $status_code"
        return 1
    fi
}

echo "📋 Test Plan:"
echo "1. Backend Health Check"
echo "2. User Registration"
echo "3. User Login"
echo "4. Room Management"
echo "5. Message Operations"
echo "6. Frontend Availability"
echo ""

# Test 1: Backend Health Check
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test 1: Backend Health Check"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check if backend is running
if curl -s "$API_URL/health" > /dev/null 2>&1; then
    print_result 0 "Backend server is running"
else
    print_result 1 "Backend server is NOT running"
    echo ""
    echo -e "${YELLOW}⚠️  Backend is not running. Please start it with:${NC}"
    echo "   make run"
    echo "   OR"
    echo "   go run cmd/server/main.go"
    exit 1
fi

# Test 2: User Registration
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test 2: User Registration"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Generate unique usernames
TIMESTAMP=$(date +%s)
USER1="testuser1_$TIMESTAMP"
USER2="testuser2_$TIMESTAMP"

# Register User 1
echo "Registering user: $USER1"
REGISTER1=$(test_endpoint POST "/api/auth/register" \
    "{\"username\":\"$USER1\",\"email\":\"$USER1@test.com\",\"password\":\"Test123456\"}" \
    "" "201")

if [ $? -eq 0 ]; then
    print_result 0 "User 1 registration successful"
    USER1_ID=$(echo "$REGISTER1" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
else
    print_result 1 "User 1 registration failed"
fi

# Register User 2
echo "Registering user: $USER2"
REGISTER2=$(test_endpoint POST "/api/auth/register" \
    "{\"username\":\"$USER2\",\"email\":\"$USER2@test.com\",\"password\":\"Test123456\"}" \
    "" "201")

if [ $? -eq 0 ]; then
    print_result 0 "User 2 registration successful"
    USER2_ID=$(echo "$REGISTER2" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
else
    print_result 1 "User 2 registration failed"
fi

# Test duplicate registration
echo "Testing duplicate registration (should fail)"
DUPLICATE=$(test_endpoint POST "/api/auth/register" \
    "{\"username\":\"$USER1\",\"email\":\"$USER1@test.com\",\"password\":\"Test123456\"}" \
    "" "400")

if [ $? -eq 0 ]; then
    print_result 0 "Duplicate registration correctly rejected"
else
    print_result 1 "Duplicate registration not handled properly"
fi

# Test 3: User Login
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test 3: User Login"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Login User 1
echo "Logging in user: $USER1"
LOGIN1=$(test_endpoint POST "/api/auth/login" \
    "{\"username\":\"$USER1\",\"password\":\"Test123456\"}" \
    "" "200")

if [ $? -eq 0 ]; then
    print_result 0 "User 1 login successful"
    TOKEN1=$(echo "$LOGIN1" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
else
    print_result 1 "User 1 login failed"
    exit 1
fi

# Login User 2
echo "Logging in user: $USER2"
LOGIN2=$(test_endpoint POST "/api/auth/login" \
    "{\"username\":\"$USER2\",\"password\":\"Test123456\"}" \
    "" "200")

if [ $? -eq 0 ]; then
    print_result 0 "User 2 login successful"
    TOKEN2=$(echo "$LOGIN2" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
else
    print_result 1 "User 2 login failed"
    exit 1
fi

# Test invalid login
echo "Testing invalid login (should fail)"
INVALID_LOGIN=$(test_endpoint POST "/api/auth/login" \
    "{\"username\":\"$USER1\",\"password\":\"WrongPassword\"}" \
    "" "401")

if [ $? -eq 0 ]; then
    print_result 0 "Invalid login correctly rejected"
else
    print_result 1 "Invalid login not handled properly"
fi

# Test 4: Room Management
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test 4: Room Management"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Get rooms (should be empty initially)
echo "Getting user rooms (should be empty)"
ROOMS=$(test_endpoint GET "/api/rooms" "" "$TOKEN1" "200")

if [ $? -eq 0 ]; then
    print_result 0 "Get rooms successful"
else
    print_result 1 "Get rooms failed"
fi

# Create group room
echo "Creating group room"
GROUP_ROOM=$(test_endpoint POST "/api/rooms" \
    "{\"name\":\"Test Group\",\"type\":\"group\",\"members\":[\"$USER2_ID\"]}" \
    "$TOKEN1" "201")

if [ $? -eq 0 ]; then
    print_result 0 "Group room creation successful"
    ROOM_ID=$(echo "$GROUP_ROOM" | grep -o '"id":"[^"]*"' | cut -d'"' -f4)
else
    print_result 1 "Group room creation failed"
fi

# Get rooms again (should have 1 room)
echo "Getting user rooms (should have 1 room)"
ROOMS_AFTER=$(test_endpoint GET "/api/rooms" "" "$TOKEN1" "200")

if [ $? -eq 0 ]; then
    print_result 0 "Get rooms after creation successful"
else
    print_result 1 "Get rooms after creation failed"
fi

# Test 5: Message Operations
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test 5: Message Operations"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if [ -n "$ROOM_ID" ]; then
    # Get messages (should be empty)
    echo "Getting messages from room (should be empty)"
    MESSAGES=$(test_endpoint GET "/api/rooms/$ROOM_ID/messages?limit=50" "" "$TOKEN1" "200")
    
    if [ $? -eq 0 ]; then
        print_result 0 "Get messages successful"
    else
        print_result 1 "Get messages failed"
    fi
    
    # Note: Sending messages requires WebSocket connection
    echo ""
    echo -e "${YELLOW}ℹ️  Message sending requires WebSocket connection${NC}"
    echo "   This is tested through the frontend UI"
else
    print_result 1 "Cannot test messages - no room created"
fi

# Test 6: Frontend Availability
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test 6: Frontend Availability"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Check if frontend is running
if curl -s "$FRONTEND_URL" > /dev/null 2>&1; then
    print_result 0 "Frontend server is running"
else
    print_result 1 "Frontend server is NOT running"
    echo ""
    echo -e "${YELLOW}⚠️  Frontend is not running. Please start it with:${NC}"
    echo "   cd frontend && npm run dev"
fi

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "Test Summary"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo -e "${GREEN}Passed: $TESTS_PASSED${NC}"
echo -e "${RED}Failed: $TESTS_FAILED${NC}"
echo ""

if [ $TESTS_FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed!${NC}"
    echo ""
    echo "🎉 System is working correctly!"
    echo ""
    echo "Next steps:"
    echo "1. Open browser to: $FRONTEND_URL"
    echo "2. Register a new user"
    echo "3. Create a chat room"
    echo "4. Send messages"
    echo "5. Test real-time features"
    exit 0
else
    echo -e "${RED}✗ Some tests failed${NC}"
    echo ""
    echo "Please check the errors above and fix them."
    exit 1
fi
