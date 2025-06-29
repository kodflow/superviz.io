#!/bin/bash
# test_simple_e2e.sh - Simplified E2E test for debugging

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
SVZ_BINARY="$PROJECT_ROOT/.dist/bin/svz_linux_arm64"

echo "Starting simplified E2E test..."

# Start Ubuntu container
echo "Starting Ubuntu container..."
CONTAINER_NAME="svz-test-simple"
docker run -d --name "$CONTAINER_NAME" svz-test-ubuntu

echo "Waiting for container to be ready..."
sleep 5

# Check container status
echo "Container status:"
docker ps --filter "name=$CONTAINER_NAME"

# Test docker exec
echo "Testing docker exec..."
docker exec "$CONTAINER_NAME" whoami

# Test SSH in container
echo "Installing sshpass and testing SSH..."
docker exec "$CONTAINER_NAME" bash -c 'apt-get update >/dev/null 2>&1 && apt-get install -y sshpass >/dev/null 2>&1'
docker exec "$CONTAINER_NAME" bash -c 'sshpass -p "testpass123" ssh -o StrictHostKeyChecking=no testuser@localhost "echo SSH test successful"'

# Copy and test binary
echo "Copying and testing binary..."
docker cp "$SVZ_BINARY" "$CONTAINER_NAME:/tmp/svz"
docker exec "$CONTAINER_NAME" chmod +x /tmp/svz

# Test install command
echo "Testing install command..."
docker exec "$CONTAINER_NAME" bash -c '/tmp/svz install testuser@localhost --password "testpass123" --skip-host-key-check --timeout 30s || echo "Install command completed (expected to fail at repo download)"'

echo "Test completed successfully!"

# Cleanup
docker rm -f "$CONTAINER_NAME"
