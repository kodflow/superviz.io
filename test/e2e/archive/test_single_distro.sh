#!/bin/bash
# test/e2e/test_single_distro.sh - Test single distribution for debugging

set -euo pipefail

readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
readonly DOCKER_DIR="${SCRIPT_DIR}/docker"

# Default to Ubuntu if no distro specified
DISTRO="${1:-ubuntu}"

# Color codes
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m'

# Distribution configurations
declare -A DISTRO_CONFIGS=(
    ["ubuntu"]="2201:testuser:testpass"
    ["debian"]="2202:testuser:testpass"
    ["alpine"]="2203:testuser:testpass"
    ["centos"]="2204:testuser:testpass"
    ["fedora"]="2205:testuser:testpass"
    ["arch"]="2206:testuser:testpass"
)

cleanup() {
    echo -e "${BLUE}🧹 Cleaning up...${NC}"
    cd "${DOCKER_DIR}" 2>/dev/null && docker-compose down --remove-orphans >/dev/null 2>&1 || true
    rm -f "${SCRIPT_DIR}/superviz"
    rm -f "${SCRIPT_DIR}"/test_key_*
    rm -f "${SCRIPT_DIR}"/test_output_*.log
}

trap cleanup EXIT INT TERM

main() {
    echo -e "${BLUE}🎯 Testing superviz install on $DISTRO${NC}"
    
    # Validate distro
    if [[ -z "${DISTRO_CONFIGS[$DISTRO]:-}" ]]; then
        echo -e "${RED}❌ Unknown distro: $DISTRO${NC}"
        echo -e "${BLUE}Available distros: ${!DISTRO_CONFIGS[*]}${NC}"
        exit 1
    fi
    
    # Build superviz
    echo -e "${BLUE}🔧 Building superviz...${NC}"
    cd "${PROJECT_ROOT}"
    go build -o "${SCRIPT_DIR}/superviz" ./cmd/svz/main.go
    
    # Start specific container
    echo -e "${BLUE}🐳 Starting $DISTRO container...${NC}"
    cd "${DOCKER_DIR}"
    docker-compose up -d --build "${DISTRO}-test"
    
    # Wait for SSH
    local config="${DISTRO_CONFIGS[$DISTRO]}"
    local port="${config%%:*}"
    
    echo -e "${BLUE}⏳ Waiting for SSH on port $port...${NC}"
    local attempts=0
    while ! nc -z localhost "$port" >/dev/null 2>&1; do
        attempts=$((attempts + 1))
        if [[ $attempts -gt 30 ]]; then
            echo -e "${RED}❌ Timeout waiting for SSH${NC}"
            exit 1
        fi
        sleep 2
    done
    
    # Test install
    local user password
    IFS=':' read -r port user password <<< "$config"
    
    echo -e "${BLUE}🧪 Testing install with password authentication...${NC}"
    
    # Test install with password authentication
    if "${SCRIPT_DIR}/superviz" install \
        --password="$password" \
        --ssh-port="$port" \
        --timeout=60s \
        --skip-host-key-check \
        "${user}@localhost"; then
        echo -e "${GREEN}✅ Test completed successfully!${NC}"
    else
        echo -e "${RED}❌ Test failed${NC}"
        exit 1
    fi
}

main "$@"
