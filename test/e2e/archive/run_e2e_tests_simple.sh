#!/bin/bash
# run_e2e_tests.sh - Run end-to-end tests across multiple distributions using Docker
# This script uses the new Docker-based E2E testing framework

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Show usage if help requested
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    cat << EOF
Usage: $0 [OPTIONS]

Run end-to-end tests for superviz.io install command across multiple Linux distributions using Docker.

This script uses Docker containers to test the install command on real Linux distributions:
- Ubuntu
- Debian  
- Alpine
- CentOS
- Fedora
- Arch Linux

OPTIONS:
    --no-cleanup        Don't remove containers after tests (for debugging)
    --help, -h          Show this help message

Prerequisites:
- Docker installed and running
- Binary built (run 'make build')
- Docker permissions for current user

For testing a single distribution:
  ./test_docker_e2e.sh --distro=<distribution>

For setting up the test environment:
  make e2e-setup

EOF
    exit 0
fi

# Delegate to the new Docker-based E2E test script
log_info "Delegating to Docker-based E2E test framework..."
exec "$SCRIPT_DIR/test_docker_e2e.sh" "$@"
