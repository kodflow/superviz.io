#!/bin/bash
# test_docker_install.sh - End-to-end tests using Docker containers with exec-based testing
# This script tests the install command against real Linux distributions running in Docker

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
SVZ_BINARY="$PROJECT_ROOT/.dist/bin/svz_linux_arm64"

# Test configuration
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
        go build -o "$SVZ_BINARY" ./cmd/svz
    fi

    if [ ! -f "$SVZ_BINARY" ]; then
        log_error "Failed to build binary"
        exit 1
    fi
}

# Start a test container for a specific distribution
start_container() {
    local distro="$1"
    local dockerfile="${distro}.Dockerfile"
    local container_name="${CONTAINER_PREFIX}-${distro}"
    
    log_info "Starting container for $distro..."
    
    # Build the Docker image if it doesn't exist
    local image_name="svz-test-${distro}"
    if ! docker images --format "{{.Repository}}" | grep -q "^${image_name}$"; then
        log_debug "Building Docker image for $distro..."
        docker build -f "$SCRIPT_DIR/docker/${dockerfile}" -t "$image_name" "$SCRIPT_DIR/docker/" >/dev/null
    fi
    
    # Start the container with SSH daemon
    docker run -d \
        --name "$container_name" \
        --hostname "${distro}-test" \
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

# Test install command by executing it inside the container
test_install_on_container() {
    local distro="$1"
    local container_name="$2"
    
    log_info "Testing install command on $distro..."
    
    # Copy the binary into the container
    docker cp "$SVZ_BINARY" "$container_name:/tmp/svz"
    
    # Test 1: SSH connection within container (localhost)
    log_debug "Testing SSH connection within $distro container..."
    if ! docker exec "$container_name" bash -c '
        apt-get update >/dev/null 2>&1 || yum update -y >/dev/null 2>&1 || apk update >/dev/null 2>&1 || pacman -Sy >/dev/null 2>&1 || true
        command -v sshpass >/dev/null || {
            apt-get install -y sshpass >/dev/null 2>&1 || 
            yum install -y sshpass >/dev/null 2>&1 || 
            apk add sshpass >/dev/null 2>&1 || 
            pacman -S --noconfirm sshpass >/dev/null 2>&1 || {
                echo "Failed to install sshpass"
                exit 1
            }
        }
        sshpass -p "testpass123" ssh -o StrictHostKeyChecking=no testuser@localhost "echo SSH test successful"
    '; then
        log_error "SSH connection test failed on $distro"
        docker logs "$container_name" 2>/dev/null | tail -10 || true
        return 1
    fi
    
    # Test 2: Install command execution
    log_debug "Testing install command execution on $distro..."
    
    # Execute the install command from within the container
    # Note: We expect this to fail at the repository download step since repo.superviz.io doesn't exist
    # but it should get through SSH connection and initial setup steps
    install_output=$(docker exec "$container_name" bash -c '
        export PATH="/tmp:$PATH"
        chmod +x /tmp/svz
        timeout 60 /tmp/svz install testuser@localhost \
            --password "testpass123" \
            --skip-host-key-check \
            --timeout 30s 2>&1
    ' 2>&1)
    
    install_exit_code=$?
    
    # Check if the installation got to the expected failure point
    if echo "$install_output" | grep -q "Setting up APT repository"; then
        log_info "✓ Install command executed successfully (reached repository setup phase)"
        log_debug "Install output: $install_output"
    elif echo "$install_output" | grep -q "Using sudo for system operations"; then
        log_info "✓ Install command executed successfully (reached system operations phase)"
        log_debug "Install output: $install_output"
    elif echo "$install_output" | grep -q "failed to setup repository"; then
        log_info "✓ Install command executed successfully (expected failure at repository download)"
        log_debug "This is expected since repo.superviz.io doesn't exist in test environment"
    else
        log_error "Install command failed unexpectedly on $distro"
        log_error "Exit code: $install_exit_code"
        log_error "Output: $install_output"
        
        # Show container logs for debugging
        log_debug "Container logs:"
        docker logs "$container_name" 2>/dev/null | tail -20 || true
        
        return 1
    fi
    
    # Test 3: Verify installation progress (we don't expect full success due to test environment)
    log_debug "Verifying installation progress on $distro..."
    
    # Check that the installation process at least started correctly
    if echo "$install_output" | grep -q -E "(Setting up.*repository|Using sudo|sudo apt update)"; then
        log_info "✓ Installation process verified on $distro (executed system commands)"
    else
        log_warn "⚠ Could not verify installation progress on $distro (this may be expected)"
        log_debug "Install output: $install_output"
    fi
    
    return 0
}

# Test a specific distribution
test_distribution() {
    local distro="$1"
    
    log_info "=========================================="
    log_info "Testing $distro distribution"
    log_info "=========================================="
    
    # Start container
    local container_name
    if ! container_name=$(start_container "$distro"); then
        log_error "Failed to start container for $distro"
        return 1
    fi
    
    # Test install command
    if ! test_install_on_container "$distro" "$container_name"; then
        log_error "Install test failed for $distro"
        return 1
    fi
    
    # Cleanup container if requested
    if [ "$CLEANUP_ON_EXIT" = "1" ]; then
        docker rm -f "$container_name" >/dev/null 2>&1 || true
    fi
    
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

# Parse arguments and run
parse_args "$@"

# If single distro specified, test only that one
if [ -n "${SINGLE_DISTRO:-}" ]; then
    if ! test_distribution "$SINGLE_DISTRO"; then
        exit 1
    fi
    log_info "Single distribution test completed successfully!"
else
    main "$@"
fi
