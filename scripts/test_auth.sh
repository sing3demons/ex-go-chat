#!/bin/bash

# Test Authentication Endpoints
# This script tests the register and login endpoints

BASE_URL="http://localhost:8080"

echo "=== Testing Authentication System ==="
echo ""

# Test 1: Health Check
echo "1. Testing Health Check..."
HEALTH_RESPONSE=$(curl -s -w "\n%{http_code}" "$BASE_URL/health")
HEALTH_CODE=$(echo "$HEALTH_RESPONSE" | tail -n1)
if [ "$HEALTH_CODE" = "200" ]; then
    echo "✓ Health check passed"
else
    echo "✗ Health check failed (HTTP $HEALTH_CODE)"
    exit 1
fi
echo ""

# Test 2: Register User
echo "2. Testing User Registration..."
REGISTER_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d '{
        "username": "testuser",
        "email": "test@example.com",
        "password": "Password123"
    }')
REGISTER_CODE=$(echo "$REGISTER_RESPONSE" | tail -n1)
REGISTER_BODY=$(echo "$REGISTER_RESPONSE" | head -n-1)

if [ "$REGISTER_CODE" = "201" ]; then
    echo "✓ User registration successful"
    echo "Response: $REGISTER_BODY"
else
    echo "✗ User registration failed (HTTP $REGISTER_CODE)"
    echo "Response: $REGISTER_BODY"
fi
echo ""

# Test 3: Login User
echo "3. Testing User Login..."
LOGIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d '{
        "identifier": "testuser",
        "password": "Password123"
    }')
LOGIN_CODE=$(echo "$LOGIN_RESPONSE" | tail -n1)
LOGIN_BODY=$(echo "$LOGIN_RESPONSE" | head -n-1)

if [ "$LOGIN_CODE" = "200" ]; then
    echo "✓ User login successful"
    echo "Response: $LOGIN_BODY"
    
    # Extract token
    TOKEN=$(echo "$LOGIN_BODY" | grep -o '"token":"[^"]*"' | cut -d'"' -f4)
    echo "Token: $TOKEN"
else
    echo "✗ User login failed (HTTP $LOGIN_CODE)"
    echo "Response: $LOGIN_BODY"
fi
echo ""

# Test 4: Invalid Login
echo "4. Testing Invalid Login..."
INVALID_LOGIN_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/login" \
    -H "Content-Type: application/json" \
    -d '{
        "identifier": "testuser",
        "password": "WrongPassword"
    }')
INVALID_LOGIN_CODE=$(echo "$INVALID_LOGIN_RESPONSE" | tail -n1)

if [ "$INVALID_LOGIN_CODE" = "401" ]; then
    echo "✓ Invalid login correctly rejected"
else
    echo "✗ Invalid login should return 401 (got HTTP $INVALID_LOGIN_CODE)"
fi
echo ""

# Test 5: Duplicate Registration
echo "5. Testing Duplicate Registration..."
DUP_REGISTER_RESPONSE=$(curl -s -w "\n%{http_code}" -X POST "$BASE_URL/api/auth/register" \
    -H "Content-Type: application/json" \
    -d '{
        "username": "testuser",
        "email": "test2@example.com",
        "password": "Password123"
    }')
DUP_REGISTER_CODE=$(echo "$DUP_REGISTER_RESPONSE" | tail -n1)

if [ "$DUP_REGISTER_CODE" = "409" ]; then
    echo "✓ Duplicate username correctly rejected"
else
    echo "✗ Duplicate username should return 409 (got HTTP $DUP_REGISTER_CODE)"
fi
echo ""

echo "=== Authentication System Tests Complete ==="
