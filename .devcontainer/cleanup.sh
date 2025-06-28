#!/bin/bash
set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🧹 Docker Cleanup Script for superviz.io${NC}"
echo -e "${YELLOW}This will clean up Docker resources to free space and fix cache issues.${NC}"
echo

# Function to show sizes
show_docker_usage() {
    echo -e "${BLUE}Docker disk usage:${NC}"
    docker system df
    echo
}

# Show current usage
show_docker_usage

echo -e "${YELLOW}What would you like to clean?${NC}"
echo "1. Clean build cache only (safe, preserves images)"
echo "2. Clean stopped containers and unused networks"
echo "3. Clean unused images (not tagged)"
echo "4. Nuclear option: Clean everything (containers, images, volumes, cache)"
echo "5. Show disk usage only"
echo "q. Quit"
echo

read -p "Enter your choice (1-5, q): " choice

case $choice in
    1)
        echo -e "${GREEN}🗑️  Cleaning build cache...${NC}"
        docker builder prune -f
        docker buildx prune -f
        ;;
    2)
        echo -e "${GREEN}🗑️  Cleaning containers and networks...${NC}"
        docker container prune -f
        docker network prune -f
        ;;
    3)
        echo -e "${GREEN}🗑️  Cleaning unused images...${NC}"
        docker image prune -f
        ;;
    4)
        echo -e "${RED}☢️  NUCLEAR CLEANUP - This will remove everything!${NC}"
        read -p "Are you sure? This will delete ALL Docker data (y/N): " confirm
        if [[ $confirm == [yY] || $confirm == [yY][eE][sS] ]]; then
            echo -e "${RED}Removing all Docker data...${NC}"
            docker system prune -a -f --volumes
            docker builder prune -a -f
            echo -e "${GREEN}✅ Complete cleanup done!${NC}"
        else
            echo -e "${YELLOW}❌ Cleanup cancelled.${NC}"
        fi
        ;;
    5)
        echo -e "${BLUE}Current Docker usage:${NC}"
        ;;
    q|Q)
        echo -e "${GREEN}👋 Goodbye!${NC}"
        exit 0
        ;;
    *)
        echo -e "${RED}❌ Invalid choice!${NC}"
        exit 1
        ;;
esac

echo
echo -e "${BLUE}After cleanup:${NC}"
show_docker_usage

echo -e "${GREEN}✅ Cleanup completed!${NC}"
