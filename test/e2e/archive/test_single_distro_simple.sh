#!/bin/bash
# test_single_distro.sh - Test install command on a single distribution using Docker

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

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Show usage if no arguments or help requested
if [ $# -eq 0 ] || [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    cat << EOF
Usage: $0 <distribution> [OPTIONS]

Test the superviz.io install command on a single Linux distribution using Docker.

DISTRIBUTION:
    ubuntu      Test on Ubuntu
    debian      Test on Debian
    alpine      Test on Alpine Linux
    centos      Test on CentOS
    fedora      Test on Fedora
    arch        Test on Arch Linux

OPTIONS:
    --no-cleanup    Don't remove container after test (for debugging)
    --help, -h      Show this help message

EXAMPLES:
    $0 ubuntu                    # Test Ubuntu with cleanup
    $0 alpine --no-cleanup       # Test Alpine and keep container
    
Prerequisites:
- Docker installed and running
- Binary built (run 'make build')

EOF
    exit 0
fi

DISTRO="$1"
shift

# Validate distribution
case "$DISTRO" in
    ubuntu|debian|alpine|centos|fedora|arch)
        ;;
    *)
        log_error "Unsupported distribution: $DISTRO"
        log_error "Supported: ubuntu, debian, alpine, centos, fedora, arch"
        exit 1
        ;;
esac

# Parse remaining arguments
EXTRA_ARGS=()
while [ $# -gt 0 ]; do
    case "$1" in
        --no-cleanup)
            EXTRA_ARGS+=("--no-cleanup")
            ;;
        *)
            EXTRA_ARGS+=("$1")
            ;;
    esac
    shift
done

log_info "Testing single distribution: $DISTRO"

# Delegate to the main Docker E2E script with single distro flag
exec "$SCRIPT_DIR/test_docker_e2e.sh" "--distro=$DISTRO" "${EXTRA_ARGS[@]}"
