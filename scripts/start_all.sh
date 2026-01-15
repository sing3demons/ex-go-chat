#!/bin/bash

# Real-time Chat System - Start All Services
# This script starts MongoDB, Backend, and Frontend

set -e

echo "🚀 Starting Real-time Chat System"
echo "=================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

# Check if MongoDB is running
echo "📦 Checking MongoDB..."
if lsof -i :27017 > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} MongoDB is already running"
else
    echo -e "${YELLOW}⚠${NC}  MongoDB is not running"
    echo ""
    echo "Please start MongoDB with one of these methods:"
    echo ""
    echo "Option 1: Using Docker"
    echo "  docker run -d -p 27017:27017 --name mongodb mongo:latest"
    echo ""
    echo "Option 2: Using Docker Compose"
    echo "  docker compose up -d mongodb"
    echo ""
    echo "Option 3: Local MongoDB"
    echo "  mongod --dbpath /path/to/data"
    echo ""
    exit 1
fi

# Check if backend is running
echo ""
echo "🔧 Checking Backend..."
if lsof -i :8080 > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Backend is already running on port 8080"
else
    echo -e "${YELLOW}⚠${NC}  Backend is not running"
    echo ""
    echo "Starting backend server..."
    echo ""
    
    # Start backend in background
    if [ -f "server" ]; then
        ./server &
        BACKEND_PID=$!
    elif command -v go &> /dev/null; then
        go run cmd/server/main.go &
        BACKEND_PID=$!
    else
        echo -e "${RED}✗${NC} Go is not installed or server binary not found"
        exit 1
    fi
    
    echo "Waiting for backend to start..."
    sleep 3
    
    if lsof -i :8080 > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Backend started successfully (PID: $BACKEND_PID)"
    else
        echo -e "${RED}✗${NC} Backend failed to start"
        exit 1
    fi
fi

# Check if frontend is running
echo ""
echo "🎨 Checking Frontend..."
if lsof -i :5173 > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} Frontend is already running on port 5173"
else
    echo -e "${YELLOW}⚠${NC}  Frontend is not running"
    echo ""
    echo "Starting frontend dev server..."
    echo ""
    
    # Check if node_modules exists
    if [ ! -d "frontend/node_modules" ]; then
        echo "Installing frontend dependencies..."
        cd frontend && npm install && cd ..
    fi
    
    # Start frontend in background
    cd frontend
    npm run dev &
    FRONTEND_PID=$!
    cd ..
    
    echo "Waiting for frontend to start..."
    sleep 5
    
    if lsof -i :5173 > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} Frontend started successfully (PID: $FRONTEND_PID)"
    else
        echo -e "${RED}✗${NC} Frontend failed to start"
        exit 1
    fi
fi

# Summary
echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo "✅ All Services Running"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
echo ""
echo "📍 Service URLs:"
echo "   Backend:  http://localhost:8080"
echo "   Frontend: http://localhost:5173"
echo "   MongoDB:  mongodb://localhost:27017"
echo ""
echo "🧪 Run tests:"
echo "   ./scripts/test_system.sh"
echo ""
echo "🌐 Open in browser:"
echo "   http://localhost:5173"
echo ""
echo "📝 View logs:"
echo "   Backend:  tail -f logs/server.log"
echo "   Frontend: Check terminal output"
echo ""
echo "🛑 Stop services:"
echo "   Press Ctrl+C or run: ./scripts/stop_all.sh"
echo ""
