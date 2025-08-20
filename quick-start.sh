#!/bin/bash

# Quick start script for shipshipship
# This starts the backend server and provides instructions for the frontend

set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Quick Start - ShipShipShip${NC}"

# Check if we're in the right directory
if [ ! -f "backend/main.go" ] || [ ! -f "frontend/package.json" ]; then
    echo -e "${RED}❌ Error: Please run this script from the shipshipship root directory${NC}"
    exit 1
fi

# Kill any existing processes on port 8080
echo -e "${YELLOW}🔍 Checking for existing processes on port 8080...${NC}"
if lsof -ti:8080 > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Killing existing process on port 8080...${NC}"
    kill -9 $(lsof -ti:8080) 2>/dev/null || true
fi

# Build backend if needed
if [ ! -f "backend/main" ]; then
    echo -e "${YELLOW}🔨 Building backend...${NC}"
    cd backend
    go build -o main .
    cd ..
fi

# Build frontend if needed
if [ ! -d "frontend/build" ]; then
    echo -e "${YELLOW}🔨 Building frontend...${NC}"
    cd frontend
    npm install
    npm run build
    cd ..
fi

# Start backend
echo -e "${GREEN}🔧 Starting backend server...${NC}"
./backend/main > backend.log 2>&1 &
BACKEND_PID=$!

# Wait for backend to start
sleep 3

# Check if backend started successfully
if curl -s http://localhost:8080/api/events > /dev/null 2>&1; then
    echo -e "${GREEN}✅ Backend started successfully on http://localhost:8080${NC}"
else
    echo -e "${RED}❌ Backend failed to start. Check backend.log for errors.${NC}"
    cat backend.log
    exit 1
fi

echo -e ""
echo -e "${GREEN}🎉 Backend is running!${NC}"
echo -e ""
echo -e "${YELLOW}📋 Next steps:${NC}"
echo -e "1. Open a new terminal window"
echo -e "2. Run: ${GREEN}cd frontend && npm run dev${NC}"
echo -e "3. Open: ${GREEN}http://localhost:5173/admin${NC}"
echo -e ""
echo -e "${YELLOW}📝 Useful commands:${NC}"
echo -e "  View backend logs: ${GREEN}tail -f backend.log${NC}"
echo -e "  Stop backend: ${GREEN}kill $BACKEND_PID${NC}"
echo -e "  Test API: ${GREEN}curl http://localhost:8080/api/events${NC}"
echo -e ""
echo -e "${GREEN}Backend PID: $BACKEND_PID${NC}"
echo -e "${GREEN}Press Ctrl+C to continue (backend will keep running)${NC}"

# Wait for user input
read -p ""

echo -e "${GREEN}✅ Backend is running in the background${NC}"
echo -e "${YELLOW}💡 Tip: Use 'kill $BACKEND_PID' to stop the backend when done${NC}"
