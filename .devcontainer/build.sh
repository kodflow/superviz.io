#!/bin/bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}🧹 Cleaning Docker environment...${NC}"
docker system prune -f
docker builder prune -f

echo -e "${GREEN}🏗️  Building devcontainer with BuildKit optimizations...${NC}"
export DOCKER_BUILDKIT=1
export BUILDKIT_PROGRESS=plain

# Build with aggressive caching and ARM64 targeting
docker build \
    --platform linux/arm64 \
    --build-arg BUILDKIT_INLINE_CACHE=1 \
    --cache-from type=local,src=/tmp/.buildx-cache \
    --cache-to type=local,dest=/tmp/.buildx-cache-new,mode=max \
    --tag superviz-devcontainer-arm64:latest \
    .devcontainer/

# Move cache to avoid growing it infinitely
if [ -d "/tmp/.buildx-cache-new" ]; then
    rm -rf /tmp/.buildx-cache
    mv /tmp/.buildx-cache-new /tmp/.buildx-cache
fi

echo -e "${GREEN}✅ Build completed! Image: superviz-devcontainer-arm64:latest${NC}"
echo -e "${YELLOW}💡 To rebuild faster next time, this script uses local cache.${NC}"
