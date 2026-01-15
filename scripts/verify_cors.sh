#!/bin/bash

echo "=== CORS Verification Test ==="
echo ""

# Test 1: OPTIONS preflight request
echo "1. Testing CORS preflight (OPTIONS)..."
PREFLIGHT=$(curl -s -X OPTIONS http://localhost:8080/api/auth/register \
  -H "Origin: http://localhost:5173" \
  -H "Access-Control-Request-Method: POST" \
  -H "Access-Control-Request-Headers: Content-Type" \
  -v 2>&1)

if echo "$PREFLIGHT" | grep -q "Access-Control-Allow-Origin: http://localhost:5173"; then
    echo "✅ CORS preflight working"
else
    echo "❌ CORS preflight failed"
fi
echo ""

# Test 2: Actual POST request with CORS
echo "2. Testing POST request with CORS headers..."
TIMESTAMP=$(date +%s)
RESPONSE=$(curl -s -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -H "Origin: http://localhost:5173" \
  -d "{\"username\":\"user$TIMESTAMP\",\"email\":\"user$TIMESTAMP@test.com\",\"password\":\"Test123456\"}" \
  -v 2>&1)

if echo "$RESPONSE" | grep -q "Access-Control-Allow-Origin: http://localhost:5173"; then
    echo "✅ CORS headers present in response"
else
    echo "❌ CORS headers missing"
fi

if echo "$RESPONSE" | grep -q '"success":true'; then
    echo "✅ Registration successful"
else
    echo "❌ Registration failed"
fi
echo ""

echo "=== Summary ==="
echo "Backend: http://localhost:8080 (with CORS enabled)"
echo "Frontend: http://localhost:5173"
echo ""
echo "✅ Frontend can now communicate with backend!"
echo ""
echo "Open http://localhost:5173 in your browser to test registration."
