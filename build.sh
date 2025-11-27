#!/bin/bash
# Build script for Security Rota App
# This builds the frontend and backend into a single Docker image

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLIENT_DIR="$SCRIPT_DIR/../securityrota-client"
API_DIR="$SCRIPT_DIR"

echo "=== Building Security Rota App ==="

# Check if client directory exists
if [ ! -d "$CLIENT_DIR" ]; then
    echo "Error: Client directory not found at $CLIENT_DIR"
    exit 1
fi

# Step 1: Build frontend
echo ""
echo "=== Step 1: Building Frontend ==="
cd "$CLIENT_DIR"

# Create/update .env for production build
echo "VITE_API_BASE_URL=/api/v1" > .env.production

npm ci
npm run build

# Step 2: Copy frontend build to API static folder
echo ""
echo "=== Step 2: Copying Frontend Build ==="
cd "$API_DIR"
rm -rf static
cp -r "$CLIENT_DIR/dist" static

echo "Frontend copied to $API_DIR/static"

# Step 3: Build Docker image
echo ""
echo "=== Step 3: Building Docker Image ==="
docker build -t securityrota:latest .

echo ""
echo "=== Build Complete ==="
echo "Run with: docker-compose -f docker-compose.prod.yml up -d"

