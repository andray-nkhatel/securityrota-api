#!/bin/bash
# Build script for Security Rota App
# Builds frontend and backend inside Docker (no local Node.js required)

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLIENT_DIR="$SCRIPT_DIR/../securityrota-client"

echo "=== Building Security Rota App ==="

# Check if client directory exists
if [ ! -d "$CLIENT_DIR" ]; then
    echo "Error: Client directory not found at $CLIENT_DIR"
    exit 1
fi

# Step 1: Copy client to build context
echo ""
echo "=== Step 1: Preparing Build Context ==="
rm -rf "$SCRIPT_DIR/client"
cp -r "$CLIENT_DIR" "$SCRIPT_DIR/client"

# Create production env
echo "VITE_API_BASE_URL=/api/v1" > "$SCRIPT_DIR/client/.env.production"

# Step 2: Build Docker image
echo ""
echo "=== Step 2: Building Docker Image ==="
cd "$SCRIPT_DIR"
docker build -t securityrota:latest .

# Cleanup
echo ""
echo "=== Cleanup ==="
rm -rf "$SCRIPT_DIR/client"

echo ""
echo "=== Build Complete ==="
echo ""
echo "Run with:"
echo "  docker-compose -f docker-compose.prod.yml up -d"
echo ""
echo "Access at: http://localhost:8090"
echo "Default login: admin / admin123"
