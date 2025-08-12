#!/bin/bash

# EIAM Platform Startup Script

echo "🚀 Starting EIAM Platform..."

# Function to cleanup background processes
cleanup() {
    echo "🛑 Shutting down EIAM Platform..."
    pkill -f "go run cmd/server/main.go"
    pkill -f "vite"
    exit 0
}

# Set up signal handlers
trap cleanup SIGINT SIGTERM

# Start backend server
echo "📡 Starting backend server..."
cd "$(dirname "$0")"
go run cmd/server/main.go &
BACKEND_PID=$!

# Wait a moment for backend to start
sleep 3

# Start frontend development server
echo "🎨 Starting frontend development server..."
cd frontend
npm run dev &
FRONTEND_PID=$!

echo "✅ EIAM Platform is starting up..."
echo "📱 Frontend: http://localhost:3000"
echo "🔧 Backend API: http://localhost:8080"
echo "📊 Health Check: http://localhost:8080/health"
echo ""
echo "Press Ctrl+C to stop all services"

# Wait for background processes
wait $BACKEND_PID $FRONTEND_PID
