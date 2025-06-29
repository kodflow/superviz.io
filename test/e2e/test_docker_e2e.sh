#!/bin/bash
# test_docker_e2e.sh - End-to-end tests using Docker containers
# This script tests the install command against real Linux distributions running in Docker

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
SVZ_BINARY="$PROJECT_ROOT/.dist/bin/svz_linux_arm64"

# Test configuration
TEST_PASSWORD="testpass123"
TEST_PORT="2222"
CONTAINER_PREFIX="svz-test"
CLEANUP_ON_EXIT=1

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
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

log_debug() {
    echo -e "${BLUE}[DEBUG]${NC} $1"
}

# Cleanup function
cleanup() {
    if [ "$CLEANUP_ON_EXIT" = "1" ]; then
        log_info "Cleaning up test containers..."
        docker ps -aq --filter "name=${CONTAINER_PREFIX}" | xargs -r docker rm -f >/dev/null 2>&1 || true
    fi
}
trap cleanup EXIT

# Test if Docker is available
check_docker() {
    if ! command -v docker >/dev/null 2>&1; then
        log_error "Docker not found. Please install Docker."
        exit 1
    fi

    if ! docker ps >/dev/null 2>&1; then
        log_error "Cannot connect to Docker daemon. Please ensure Docker is running and you have permissions."
        exit 1
    fi
}

# Build binary if it doesn't exist
ensure_binary() {
    if [ ! -f "$SVZ_BINARY" ]; then
        log_info "Binary not found. Building..."
        cd "$PROJECT_ROOT"
        make build
    fi

    if [ ! -f "$SVZ_BINARY" ]; then
        log_error "Failed to build binary"
        exit 1
    fi
}

# Start a test container for a specific distribution
start_container() {
    local distro="$1"
    local dockerfile="$2"
    local container_name="${CONTAINER_PREFIX}-${distro}"
    
    log_info "Starting container for $distro..."
    
    # Build the Docker image if it doesn't exist
    local image_name="svz-test-${distro}"
    if ! docker images --format "{{.Repository}}" | grep -q "^${image_name}$"; then
        log_debug "Building Docker image for $distro..."
        docker build -f "$SCRIPT_DIR/docker/${dockerfile}" -t "$image_name" "$SCRIPT_DIR/docker/" >/dev/null
    fi
    
    # Start the container (let Docker choose available port)
    docker run -d \
        --name "$container_name" \
        --hostname "${distro}-test" \
        -P \
        "$image_name" >/dev/null
    
    # Wait for SSH to be ready
    log_debug "Waiting for SSH to be ready on $distro..."
    local max_attempts=30
    local attempt=0
    
    while [ $attempt -lt $max_attempts ]; do
        if docker exec "$container_name" pgrep sshd >/dev/null 2>&1; then
            break
        fi
        sleep 1
        attempt=$((attempt + 1))
    done
    
    if [ $attempt -eq $max_attempts ]; then
        log_error "SSH failed to start in $distro container"
        docker logs "$container_name" 2>/dev/null | tail -10 || true
        return 1
    fi
    
    # Additional wait for SSH to be fully ready
    sleep 2
    
    log_debug "Container $distro is ready"
    echo "$container_name"
}

# Test install command against a specific container
test_install_on_container() {
    local distro="$1"
    local container_name="$2"
    
    log_info "Testing install command on $distro..."
    
    # Get the mapped port for this container
    local ssh_port
    ssh_port=$(docker port "$container_name" 22/tcp 2>/dev/null | cut -d: -f2)
    
    if [ -z "$ssh_port" ]; then
        log_error "Could not find mapped SSH port for $container_name"
        log_debug "Available ports:"
        docker port "$container_name" 2>/dev/null || true
        log_debug "Container status:"
        docker ps --filter "name=$container_name" || true
        return 1
    fi
    
    local container_host="127.0.0.1"
    log_debug "Using connection: testuser@${container_host}:${ssh_port} for $distro"
    
    # Wait a bit more for SSH to be fully ready
    log_debug "Waiting for SSH service to be fully ready..."
    sleep 3
    
    # Test SSH connectivity first
    if ! timeout 10 ssh -o StrictHostKeyChecking=no -o ConnectTimeout=5 \
        -p "$ssh_port" "testuser@${container_host}" \
        -o PasswordAuthentication=yes \
        "echo SSH connection test" 2>/dev/null; then
        log_debug "Direct SSH test failed, container may still be starting..."
        
        # Check container status
        log_debug "Container status:"
        docker ps --filter "name=$container_name" || true
        log_debug "Container logs:"
        docker logs "$container_name" 2>/dev/null | tail -10 || true
        
        return 1
    fi
    
    # Test 1: Password authentication
    log_debug "Testing password authentication on $distro..."
    if ! timeout 30 "$SVZ_BINARY" install "testuser@${container_host}" \
        --password "$TEST_PASSWORD" \
        --ssh-port "$ssh_port" \
        --skip-host-key-check \
        --timeout 20s; then
        log_error "Password authentication failed on $distro"
        
        # Debug information
        log_debug "Container status:"
        docker ps --filter "name=$container_name"
        log_debug "Container logs:"
        docker logs "$container_name" | tail -10
        
        return 1
    fi
    
    # Test 2: Verify repository was set up
    log_debug "Verifying repository setup on $distro..."
    case "$distro" in
        ubuntu|debian)
            if ! docker exec "$container_name" test -f /etc/apt/sources.list.d/superviz.list; then
                log_error "Repository file not created on $distro"
                return 1
            fi
            ;;
        alpine)
            if ! docker exec "$container_name" test -f /etc/apk/repositories.d/superviz.list; then
                log_error "Repository file not created on $distro"
                return 1
            fi
            ;;
        centos|fedora)
            if ! docker exec "$container_name" test -f /etc/yum.repos.d/superviz.repo; then
                log_error "Repository file not created on $distro"
                return 1
            fi
            ;;
        arch)
            if ! docker exec "$container_name" grep -q "superviz" /etc/pacman.conf; then
                log_error "Repository not added to pacman.conf on $distro"
                return 1
            fi
            ;;
    esac
    
    log_info "✓ Install command works correctly on $distro"
    return 0
}

# Test specific distribution
test_distribution() {
    local distro="$1"
    local dockerfile="${distro}.Dockerfile"
    
    log_info "=========================================="
    log_info "Testing $distro distribution"
    log_info "=========================================="
    
    # Start container
    local container_name
    if ! container_name=$(start_container "$distro" "$dockerfile"); then
        log_error "Failed to start container for $distro"
        return 1
    fi
    
    # Test install command
    if ! test_install_on_container "$distro" "$container_name"; then
        log_error "Install test failed for $distro"
        return 1
    fi
    
    # Cleanup container
    docker rm -f "$container_name" >/dev/null
    
    log_info "✓ $distro test completed successfully"
    return 0
}

# Main test runner
main() {
    log_info "Starting superviz.io install command Docker E2E tests"
    log_info "====================================================="
    
    # Pre-flight checks
    check_docker
    ensure_binary
    
    # Available distributions
    local distributions=("ubuntu" "debian" "alpine" "centos" "fedora" "arch")
    local failed_tests=0
    local total_tests=${#distributions[@]}
    
    # Run tests for each distribution
    for distro in "${distributions[@]}"; do
        if [ -f "$SCRIPT_DIR/docker/${distro}.Dockerfile" ]; then
            if ! test_distribution "$distro"; then
                ((failed_tests++))
            fi
        else
            log_warn "Dockerfile not found for $distro, skipping..."
        fi
    done
    
    # Summary
    echo
    log_info "====================================================="
    if [ $failed_tests -eq 0 ]; then
        log_info "All $total_tests distribution tests passed! ✓"
        log_info ""
        log_info "The install command works correctly across all tested distributions:"
        for distro in "${distributions[@]}"; do
            log_info "  ✓ $distro"
        done
    else
        log_error "$failed_tests out of $total_tests distribution tests failed! ✗"
        return 1
    fi
}

# Parse command line arguments
parse_args() {
    while [ $# -gt 0 ]; do
        case "$1" in
            --no-cleanup)
                CLEANUP_ON_EXIT=0
                log_info "Cleanup disabled - containers will be left running"
                ;;
            --distro=*)
                SINGLE_DISTRO="${1#*=}"
                log_info "Testing single distribution: $SINGLE_DISTRO"
                ;;
            --help|-h)
                show_usage
                exit 0
                ;;
            *)
                log_error "Unknown argument: $1"
                show_usage
                exit 1
                ;;
        esac
        shift
    done
}

# Show usage information
show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Docker-based end-to-end tests for superviz.io install command.

OPTIONS:
    --no-cleanup        Don't remove containers after tests (for debugging)
    --distro=NAME       Test only specific distribution (ubuntu, debian, alpine, centos, fedora, arch)
    --help, -h          Show this help message

EXAMPLES:
    $0                          # Test all distributions
    $0 --distro=ubuntu          # Test only Ubuntu
    $0 --no-cleanup             # Keep containers running after tests

REQUIREMENTS:
    - Docker must be installed and running
    - User must have Docker permissions
    - Binary must be built (run 'make build' first)

EOF
}

# Parse arguments and run
parse_args "$@"

# If single distro specified, test only that one
if [ -n "${SINGLE_DISTRO:-}" ]; then
    distributions=("$SINGLE_DISTRO")
    if ! test_distribution "$SINGLE_DISTRO"; then
        exit 1
    fi
    log_info "Single distribution test completed successfully!"
else
    main "$@"
fi
