#!/bin/bash

# Start development servers for shipshipship
# This script starts both the backend and admin panel in development mode

set -e

# Parse command line arguments
REBUILD=false
if [ "$1" = "--rebuild" ]; then
    REBUILD=true
    echo -e "${GREEN}🔄 Rebuild mode enabled${NC}"
fi

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🚀 Starting shipshipship admin development servers...${NC}"

# Function to cleanup processes on exit
cleanup() {
    echo -e "\n${YELLOW}🛑 Shutting down servers...${NC}"
    if [ ! -z "$BACKEND_PID" ]; then
        kill $BACKEND_PID 2>/dev/null || true
        echo -e "${GREEN}✅ Backend stopped${NC}"
    fi
    if [ ! -z "$ADMIN_PID" ]; then
        kill $ADMIN_PID 2>/dev/null || true
        echo -e "${GREEN}✅ Admin stopped${NC}"
    fi
    exit 0
}

# Set up trap to cleanup on script exit
trap cleanup INT TERM EXIT

# Check if we're in the right directory
if [ ! -f "backend/main.go" ] || [ ! -f "admin/package.json" ]; then
    echo -e "${RED}❌ Error: Please run this script from the shipshipship root directory${NC}"
    echo -e "${YELLOW}Usage: $0 [--rebuild]${NC}"
    echo -e "${YELLOW}  --rebuild: Force rebuild of backend and admin${NC}"
    exit 1
fi

# Check if backend binary exists or if rebuild is requested
if [ ! -f "backend/main" ] || [ "$REBUILD" = true ]; then
    if [ "$REBUILD" = true ]; then
        echo -e "${YELLOW}🔄 Rebuilding backend...${NC}"
    else
        echo -e "${YELLOW}⚠️  Backend binary not found. Building...${NC}"
    fi
    cd backend
    go build -o main .
    cd ..
    echo -e "${GREEN}✅ Backend built successfully${NC}"
fi



# Check if admin is built or if rebuild is requested
if [ ! -d "admin/build" ] || [ "$REBUILD" = true ]; then
    if [ "$REBUILD" = true ]; then
        echo -e "${YELLOW}🔄 Rebuilding admin...${NC}"
        cd admin
        rm -rf build node_modules/.cache 2>/dev/null || true
    else
        echo -e "${YELLOW}⚠️  Admin build not found. Building...${NC}"
        cd admin
    fi
    npm run build
    cd ..
    echo -e "${GREEN}✅ Admin built successfully${NC}"
fi

# Start backend server
echo -e "${GREEN}🔧 Starting backend server on port 8080...${NC}"
./backend/main > backend.log 2>&1 &
BACKEND_PID=$!

# Wait a moment for backend to start
sleep 2

# Check if backend is running
if ! kill -0 $BACKEND_PID 2>/dev/null; then
    echo -e "${RED}❌ Backend failed to start. Check backend.log for errors.${NC}"
    cat backend.log
    exit 1
fi

# Test backend API
if curl -s http://localhost:8080/api/events > /dev/null; then
    echo -e "${GREEN}✅ Backend is running and responding${NC}"
else
    echo -e "${YELLOW}⚠️  Backend started but API may not be responding yet${NC}"
fi

# Start admin development server
echo -e "${GREEN}🔧 Starting admin development server...${NC}"
cd admin
npm run dev > ../admin.log 2>&1 &
ADMIN_PID=$!
cd ..

# Wait a moment for servers to start
sleep 3

echo -e "${GREEN}🎉 Development servers started successfully!${NC}"
echo -e ""
echo -e "${GREEN}📊 Backend:${NC}  http://localhost:8080"
echo -e "${GREEN}🔧 Admin:${NC}    http://localhost:5173"
echo -e ""
echo -e "${YELLOW}📝 Logs:${NC}"
echo -e "   Backend: tail -f backend.log"
echo -e "   Admin:   tail -f admin.log"
echo -e ""
echo -e "${GREEN}Press Ctrl+C to stop all servers${NC}"

# Keep script running and show live logs
echo -e "\n${YELLOW}📊 Live Backend Logs:${NC}"
tail -f backend.log &
TAIL_PID=$!

# Wait for user interrupt
wait
