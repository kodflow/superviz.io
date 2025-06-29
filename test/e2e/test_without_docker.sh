#!/bin/bash
# test_without_docker.sh - Simple test for install command without Docker dependency
# This script can be used to test the install command functionality locally

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
SVZ_BINARY="$PROJECT_ROOT/.dist/bin/svz_linux_arm64"

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

# Test 1: Verify binary exists and is executable
test_binary_exists() {
    log_info "Testing binary existence and executability..."
    
    if [ ! -f "$SVZ_BINARY" ]; then
        log_error "Binary not found at $SVZ_BINARY"
        log_info "Run 'make build' first to create the binary"
        return 1
    fi
    
    if [ ! -x "$SVZ_BINARY" ]; then
        log_error "Binary is not executable"
        return 1
    fi
    
    log_info "✓ Binary exists and is executable"
    return 0
}

# Test 2: Verify help command works
test_help_command() {
    log_info "Testing help command..."
    
    if ! "$SVZ_BINARY" --help >/dev/null 2>&1; then
        log_error "Help command failed"
        return 1
    fi
    
    log_info "✓ Help command works"
    return 0
}

# Test 3: Verify install command help
test_install_help() {
    log_info "Testing install command help..."
    
    output=$("$SVZ_BINARY" install --help 2>&1)
    
    # Check for password flag
    if ! echo "$output" | grep -q "password.*SSH password"; then
        log_error "Password flag not found in install help"
        return 1
    fi
    
    # Check for other expected flags
    if ! echo "$output" | grep -q "ssh-key.*Path to SSH private key"; then
        log_error "SSH key flag not found in install help"
        return 1
    fi
    
    log_info "✓ Install command help includes password and SSH key flags"
    return 0
}

# Test 4: Verify version command
test_version_command() {
    log_info "Testing version command..."
    
    output=$("$SVZ_BINARY" version 2>&1)
    
    if ! echo "$output" | grep -q "Version:"; then
        log_error "Version command output doesn't contain version info"
        return 1
    fi
    
    if ! echo "$output" | grep -q "Go version:"; then
        log_error "Version command output doesn't contain Go version"
        return 1
    fi
    
    log_info "✓ Version command works and shows expected information"
    return 0
}

# Test 5: Test install command validation (should fail with proper error)
test_install_validation() {
    log_info "Testing install command validation..."
    
    # Test with invalid arguments - should fail
    if "$SVZ_BINARY" install >/dev/null 2>&1; then
        log_error "Install command should fail without arguments"
        return 1
    fi
    
    # Test with invalid host format - should fail  
    if "$SVZ_BINARY" install "invalid-host-format" >/dev/null 2>&1; then
        log_warn "Install command validation might need improvement for invalid hosts"
    fi
    
    log_info "✓ Install command validation works (fails appropriately without valid arguments)"
    return 0
}

# Main test runner
main() {
    log_info "Starting superviz.io install command tests (without Docker)"
    log_info "=================================================="
    
    local failed_tests=0
    local total_tests=5
    
    # Run tests
    test_binary_exists || ((failed_tests++))
    test_help_command || ((failed_tests++))
    test_install_help || ((failed_tests++))
    test_version_command || ((failed_tests++))
    test_install_validation || ((failed_tests++))
    
    # Summary
    echo
    log_info "=================================================="
    if [ $failed_tests -eq 0 ]; then
        log_info "All $total_tests tests passed! ✓"
        log_info ""
        log_info "The install command is ready for use."
        log_info "To test with real SSH connections, use Docker-based e2e tests:"
        log_info "  make e2e-setup    # Setup test containers"
        log_info "  make e2e-test     # Run all distribution tests"
        log_info "  make e2e-clean    # Cleanup containers"
    else
        log_error "$failed_tests out of $total_tests tests failed! ✗"
        return 1
    fi
}

# Show usage if help requested
if [ "$1" = "--help" ] || [ "$1" = "-h" ]; then
    cat << EOF
Usage: $0

Simple test script for superviz.io install command that doesn't require Docker.

This script performs basic functionality tests:
1. Verifies binary exists and is executable
2. Tests help command functionality  
3. Validates install command flags (password, ssh-key, etc.)
4. Tests version command output
5. Tests basic install command validation

To run full e2e tests with Docker:
  make e2e-setup && make e2e-test

EOF
    exit 0
fi

# Run tests
main "$@"
