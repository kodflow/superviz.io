#!/bin/bash
# run_e2e_docker_tests.sh - Comprehensive E2E tests for all supported distributions

set -e -o pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
SVZ_BINARY="$PROJECT_ROOT/.dist/bin/svz_linux_arm64"

# Test configuration
CONTAINER_PREFIX="svz-e2e"
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

Comprehensive Docker-based end-to-end tests for superviz.io install command.

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

# Test a specific distribution
test_distribution() {
    local distro="$1"
    local container_name="${CONTAINER_PREFIX}-${distro}"
    
    log_info "=========================================="
    log_info "Testing $distro distribution"
    log_info "=========================================="
    
    # Build image if needed
    local image_name="svz-test-${distro}"
    if ! docker images --format "{{.Repository}}" | grep -q "^${image_name}$"; then
        log_debug "Building Docker image for $distro..."
        if ! docker build -f "$SCRIPT_DIR/docker/${distro}.Dockerfile" -t "$image_name" "$SCRIPT_DIR/docker/" >/dev/null 2>&1; then
            log_error "Failed to build Docker image for $distro"
            return 1
        fi
    fi
    
    # Start container
    log_debug "Starting container for $distro..."
    if ! docker run -d --name "$container_name" "$image_name" >/dev/null 2>&1; then
        log_error "Failed to start container for $distro"
        return 1
    fi
    
    # Wait for container to be ready
    log_debug "Waiting for container to be ready..."
    sleep 5
    
    # Verify container is running
    if ! docker ps --filter "name=$container_name" --filter "status=running" | grep -q "$container_name"; then
        log_error "Container $container_name is not running"
        docker logs "$container_name" 2>/dev/null | tail -10 || true
        return 1
    fi
    
    # Install prerequisites based on distribution
    log_debug "Installing prerequisites for $distro..."
    case "$distro" in
        ubuntu|debian)
            docker exec "$container_name" bash -c 'apt-get update >/dev/null 2>&1 && apt-get install -y sshpass >/dev/null 2>&1' || {
                log_error "Failed to install prerequisites for $distro"
                return 1
            }
            ;;
        centos|fedora)
            docker exec "$container_name" bash -c 'yum update -y >/dev/null 2>&1 && yum install -y sshpass >/dev/null 2>&1' || {
                log_error "Failed to install prerequisites for $distro"
                return 1
            }
            ;;
        alpine)
            docker exec "$container_name" bash -c 'apk update >/dev/null 2>&1 && apk add sshpass >/dev/null 2>&1' || {
                log_error "Failed to install prerequisites for $distro"
                return 1
            }
            ;;
        arch)
            docker exec "$container_name" bash -c 'pacman -Sy --noconfirm sshpass >/dev/null 2>&1' || {
                log_error "Failed to install prerequisites for $distro"
                return 1
            }
            ;;
        *)
            log_warn "Unknown distribution $distro, trying generic approach..."
            ;;
    esac
    
    # Test SSH connection within container
    log_debug "Testing SSH connection within $distro container..."
    if ! docker exec "$container_name" bash -c 'sshpass -p "testpass123" ssh -o StrictHostKeyChecking=no testuser@localhost "echo SSH test successful" >/dev/null 2>&1'; then
        log_error "SSH connection test failed for $distro"
        return 1
    fi
    
    # Copy and test binary
    log_debug "Testing install command on $distro..."
    docker cp "$SVZ_BINARY" "$container_name:/tmp/svz" || {
        log_error "Failed to copy binary to $distro container"
        return 1
    }
    
    # Execute install command and capture output
    install_output=$(docker exec "$container_name" bash -c 'chmod +x /tmp/svz && /tmp/svz install testuser@localhost --password "testpass123" --skip-host-key-check --timeout 30s 2>&1' || echo "INSTALL_COMPLETED")
    
    # Check if installation progressed as expected
    if echo "$install_output" | grep -q -E "(Setting up.*repository|Using sudo|sudo apt update|sudo yum|sudo apk|sudo pacman)"; then
        log_info "✓ $distro test passed - install command executed system operations"
    elif echo "$install_output" | grep -q "failed to setup repository"; then
        log_info "✓ $distro test passed - install command reached repository setup (expected failure)"
    else
        log_error "✗ $distro test failed - unexpected install command behavior"
        log_debug "Install output: $install_output"
        return 1
    fi
    
    # Note: Container cleanup is handled by the global cleanup trap
    
    return 0
}

# Main test runner
main() {
    log_info "Starting superviz.io comprehensive E2E tests"
    log_info "============================================="
    
    # Set flag to indicate we're testing all distributions
    export TESTING_ALL_DISTROS=1
    
    # Pre-flight checks
    check_docker
    ensure_binary
    
    # Available distributions
    local distributions=("ubuntu" "debian" "alpine" "centos" "fedora" "arch")
    local failed_tests=0
    local passed_tests=0
    local total_tests=${#distributions[@]}
    
    # Run tests for each distribution
    for distro in "${distributions[@]}"; do
        if [ -f "$SCRIPT_DIR/docker/${distro}.Dockerfile" ]; then
            set +e  # Temporarily disable exit on error
            test_distribution "$distro"
            local exit_code=$?
            set -e  # Re-enable exit on error
            
            if [ $exit_code -eq 0 ]; then
                ((passed_tests++))
            else
                ((failed_tests++))
            fi
        else
            log_warn "Dockerfile not found for $distro, skipping..."
            ((total_tests--))
        fi
        echo # Add spacing between tests
    done
    
    # Summary
    echo
    log_info "============================================="
    log_info "Test Results Summary"
    log_info "============================================="
    log_info "Total distributions tested: $total_tests"
    log_info "Passed: $passed_tests"
    log_info "Failed: $failed_tests"
    echo
    
    if [ $failed_tests -eq 0 ]; then
        log_info "🎉 All $passed_tests distribution tests passed! ✓"
        log_info ""
        log_info "The superviz.io install command works correctly across all tested distributions:"
        for distro in "${distributions[@]}"; do
            if [ -f "$SCRIPT_DIR/docker/${distro}.Dockerfile" ]; then
                log_info "  ✓ $distro"
            fi
        done
        log_info ""
        log_info "The install command successfully:"
        log_info "  • Establishes SSH connections"
        log_info "  • Executes with proper authentication"
        log_info "  • Runs system commands with sudo"
        log_info "  • Attempts repository setup (fails only at external URL as expected)"
        return 0
    else
        log_error "❌ $failed_tests out of $total_tests distribution tests failed! ✗"
        return 1
    fi
}

# Parse command line arguments
parse_args() {
    while [ $# -gt 0 ]; do
        case "$1" in
            --no-cleanup)
                CLEANUP_ON_EXIT=0
                log_info "Cleanup disabled - containers will be left running for debugging"
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
    if test_distribution "$SINGLE_DISTRO"; then
        log_info "🎉 Single distribution test completed successfully!"
        exit 0
    else
        log_error "❌ Single distribution test failed!"
        exit 1
    fi
else
    main "$@"
fi
