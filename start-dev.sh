#!/bin/bash

# Start development servers for shipshipship
# This script starts both the backend and admin panel in development mode

set -e

# Colors for output (define before any usage)
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Backend environment defaults (can be overridden via user env)
: "${ADMIN_USERNAME:=admin}"
: "${ADMIN_PASSWORD:=admin}"
: "${JWT_SECRET:=dev-secret}"
: "${PORT:=8080}"
: "${GIN_MODE:=debug}"
: "${DB_PATH:=./data/changelog.db}"

echo -e "${YELLOW}🧪 Using backend env -> ADMIN_USERNAME=$ADMIN_USERNAME PORT=$PORT GIN_MODE=$GIN_MODE DB_PATH=$DB_PATH${NC}"

# Parse command line arguments
REBUILD=false
if [ "$1" = "--rebuild" ]; then
    REBUILD=true
    echo -e "${GREEN}🔄 Rebuild mode enabled${NC}"
fi

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
    # Ensure dependencies are installed
    if [ ! -d node_modules ]; then
        echo -e "${YELLOW}📦 Installing admin dependencies...${NC}"
        npm install
    fi
    npm run build
    cd ..
    echo -e "${GREEN}✅ Admin built successfully${NC}"
fi

# Kill any existing process using backend port (8080)
if lsof -ti:8080 > /dev/null 2>&1; then
    echo -e "${YELLOW}⚠️  Killing existing process on port 8080...${NC}"
    kill -9 $(lsof -ti:8080) 2>/dev/null || true
fi

# Start backend server
echo -e "${GREEN}🔧 Starting backend server on port $PORT...${NC}"
PORT="$PORT" ADMIN_USERNAME="$ADMIN_USERNAME" ADMIN_PASSWORD="$ADMIN_PASSWORD" JWT_SECRET="$JWT_SECRET" GIN_MODE="$GIN_MODE" DB_PATH="$DB_PATH" ./backend/main > backend.log 2>&1 &
BACKEND_PID=$!
export BACKEND_PID

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
# Ensure dependencies are installed before dev
if [ ! -d node_modules ]; then
    echo -e "${YELLOW}📦 Installing admin dependencies for dev...${NC}"
    npm install
fi
npm run dev > ../admin.log 2>&1 &
ADMIN_PID=$!
export ADMIN_PID
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
