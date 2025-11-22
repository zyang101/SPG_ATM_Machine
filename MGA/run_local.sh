#!/bin/bash

# Script to run both backend and frontend services
set -e

# Change to script directory (project root)
cd "$(dirname "$0")" || exit 1

# Cleanup function to kill background processes
cleanup() {
    if [ ! -z "$BACKEND_PID" ]; then
        kill $BACKEND_PID 2>/dev/null || true
    fi
    if [ ! -z "$FRONTEND_PID" ]; then
        kill $FRONTEND_PID 2>/dev/null || true
    fi
    exit 0
}

# Set trap to cleanup on script exit
trap cleanup SIGINT SIGTERM EXIT

OS="$(uname -s | tr '[:upper:]' '[:lower:]')"

if [[ "$OS" == *mingw* ]] || [[ "$OS" == *msys* ]] || [[ "$OS" == *cygwin* ]]; then
    PLATFORM="windows"
elif [[ "$OS" == "linux" ]]; then
    PLATFORM="linux"
else
    echo "Unknown OS: $OS"
    exit 1
fi

echo "Detected platform: $PLATFORM"

cd Backend || { echo "Backend folder not found!" >&2; exit 1; }

if [[ $PLATFORM == "windows" ]]; then
    BACKEND="./smart_thermostat.exe"
else
    BACKEND="./smart_thermostat"
fi

if [ ! -f "$BACKEND" ]; then
    echo "Backend executable not found: $BACKEND"
    exit 1
fi

echo "Starting backend using: $BACKEND"

echo "checking go dependencies"
cd ..
# Check for Go installation
if command -v go &> /dev/null; then
    cd Backend || { echo "Backend folder not found!" >&2; exit 1; }
    # go mod download >/dev/null 2>&1 || { echo "Backend dependency install failed." >&2; exit 1; }
    # go mod tidy >/dev/null 2>&1 || { echo "Backend dependency tidy failed." >&2; exit 1; }
    echo "Starting backend..."
    MGA_SEED=1 MGA_HTTP_ADDR=:8080 $BACKEND >/dev/null 2>&1 &
    BACKEND_PID=$!
    sleep 2
    if ! kill -0 $BACKEND_PID 2>/dev/null; then
        echo "Backend failed to start." >&2
        exit 1
    fi
    cd ..
else
    BACKEND_PID=""
fi

# Frontend setup
cd Frontend || { echo "Frontend folder not found!" >&2; exit 1; }

# Create .env.local if it doesn't exist
if [ ! -f .env.local ]; then
    echo "NEXT_PUBLIC_API_URL=http://localhost:8080" > .env.local
fi

echo "checking front end dependencies"
npm install --legacy-peer-deps >/dev/null 2>&1 || { echo "Frontend dependency install failed." >&2; exit 1; }

# Start frontend in foreground (this will block)
echo "Starting frontend..."
npm run dev >/dev/null 2>&1 &
FRONTEND_PID=$!

echo "Backend and frontend services started. Ready at http://localhost:3000"
# Wait for both processes
wait $FRONTEND_PID

