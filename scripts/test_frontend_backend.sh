#!/bin/bash

echo "=== Testing Frontend & Backend Integration ==="
echo ""

# Test 1: Backend Health
echo "1. Testing Backend (Port 8080)..."
BACKEND_RESPONSE=$(curl -s -o /dev/null -w "%{http_code}" http://localhost:8080/api/auth/register -X POST -H "Content-Type: application/json" -d '{"username":"test","email":"test@test.com","password":"Test123"}' 2>&1)
if [[ $BACKEND_RESPONSE == *"200"* ]] || [[ $BACKEND_RESPONSE == *"400"* ]]; then
    echo "✅ Backend is responding"
else
    echo "❌ Backend is NOT responding (HTTP: $BACKEND_RESPONSE)"
fi
echo ""

# Test 2: Frontend Health
echo "2. Testing Frontend (Port 5173)..."
FRONTEND_RESPONSE=$(curl -s http://localhost:5173/ 2>&1)
if [[ $FRONTEND_RESPONSE == *"root"* ]] && [[ $FRONTEND_RESPONSE == *"main.tsx"* ]]; then
    echo "✅ Frontend is serving HTML"
else
    echo "❌ Frontend is NOT serving properly"
fi
echo ""

# Test 3: Frontend API Configuration
echo "3. Testing Frontend API Configuration..."
API_CONFIG=$(curl -s http://localhost:5173/src/services/api.ts 2>&1)
if [[ $API_CONFIG == *"localhost:8080"* ]]; then
    echo "✅ Frontend is configured to use localhost:8080"
else
    echo "❌ Frontend API configuration is incorrect"
fi
echo ""

# Test 4: Frontend React Loading
echo "4. Testing Frontend React Components..."
APP_COMPONENT=$(curl -s http://localhost:5173/src/App.tsx 2>&1)
if [[ $APP_COMPONENT == *"BrowserRouter"* ]] && [[ $APP_COMPONENT == *"LoginPage"* ]]; then
    echo "✅ Frontend React components are loading"
else
    echo "❌ Frontend React components are NOT loading"
fi
echo ""

echo "=== Summary ==="
echo "Backend: http://localhost:8080"
echo "Frontend: http://localhost:5173"
echo ""
echo "To test in browser:"
echo "1. Open http://localhost:5173 in your browser"
echo "2. You should see the login/register page"
echo "3. Try registering a new user"
echo "4. Check browser console (F12) for any errors"
echo ""
