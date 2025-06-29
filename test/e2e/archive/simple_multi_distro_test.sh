#!/bin/bash
# simple_multi_distro_test.sh - Simple test for multiple distributions

# set -e  # Disabled for debugging

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
SVZ_BINARY="$PROJECT_ROOT/.dist/bin/svz_linux_arm64"

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Test function for one distribution
test_distro() {
    local distro="$1"
    local container_name="svz-simple-${distro}"
    
    log_info "Testing $distro..."
    
    # Start container
    docker run -d --name "$container_name" "svz-test-${distro}" >/dev/null
    sleep 5
    
    # Install prerequisites
    case "$distro" in
        ubuntu|debian)
            docker exec "$container_name" bash -c 'apt-get update >/dev/null 2>&1 && apt-get install -y sshpass >/dev/null 2>&1'
            ;;
    esac
    
    # Test SSH
    docker exec "$container_name" bash -c 'sshpass -p "testpass123" ssh -o StrictHostKeyChecking=no testuser@localhost "echo SSH OK" >/dev/null'
    
    # Copy and test binary
    docker cp "$SVZ_BINARY" "$container_name:/tmp/svz"
    install_output=$(docker exec "$container_name" bash -c 'chmod +x /tmp/svz && /tmp/svz install testuser@localhost --password "testpass123" --skip-host-key-check --timeout 30s 2>&1' || echo "COMPLETED")
    
    # Check result
    if echo "$install_output" | grep -q -E "(Setting up.*repository|Using sudo|sudo apt update)"; then
        log_info "✓ $distro test PASSED"
        docker rm -f "$container_name" >/dev/null
        return 0
    else
        log_error "✗ $distro test FAILED"
        log_error "Output: $install_output"
        docker rm -f "$container_name" >/dev/null
        return 1
    fi
}

# Main
log_info "Starting simple multi-distribution test"

# Test distributions
PASSED=0
FAILED=0

for distro in ubuntu debian; do
    echo "DEBUG: About to test $distro"
    set +e  # Disable exit on error for testing
    test_distro "$distro"
    exit_code=$?
    set -e  # Re-enable exit on error
    
    if [ $exit_code -eq 0 ]; then
        ((PASSED++))
        echo "DEBUG: $distro test succeeded"
    else
        ((FAILED++))
        echo "DEBUG: $distro test failed with exit code $exit_code"
    fi
    echo "DEBUG: Completed $distro, continuing to next..."
    echo
done

# Summary
log_info "============================================="
log_info "SUMMARY: $PASSED passed, $FAILED failed"
log_info "============================================="

if [ $FAILED -eq 0 ]; then
    log_info "🎉 All tests passed!"
    exit 0
else
    log_error "❌ Some tests failed!"
    exit 1
fi
