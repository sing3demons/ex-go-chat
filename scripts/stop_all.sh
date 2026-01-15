#!/bin/bash

# Real-time Chat System - Stop All Services

echo "🛑 Stopping Real-time Chat System"
echo "=================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

# Stop backend
echo "Stopping backend..."
if lsof -ti :8080 > /dev/null 2>&1; then
    kill $(lsof -ti :8080) 2>/dev/null
    echo -e "${GREEN}✓${NC} Backend stopped"
else
    echo "Backend was not running"
fi

# Stop frontend
echo "Stopping frontend..."
if lsof -ti :5173 > /dev/null 2>&1; then
    kill $(lsof -ti :5173) 2>/dev/null
    echo -e "${GREEN}✓${NC} Frontend stopped"
else
    echo "Frontend was not running"
fi

# Optional: Stop MongoDB (if running in Docker)
echo ""
echo "MongoDB is still running (not stopped automatically)"
echo "To stop MongoDB:"
echo "  docker stop mongodb"
echo "  OR"
echo "  docker compose down"
echo ""

echo "✅ Services stopped"
