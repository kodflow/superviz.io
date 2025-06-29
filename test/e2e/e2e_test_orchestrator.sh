#!/bin/bash
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
SVZ_BINARY="$PROJECT_ROOT/.dist/bin/svz_linux_arm64"

# Cleanup function
cleanup() {
    cd "$SCRIPT_DIR"
    docker-compose --profile test down --remove-orphans >/dev/null 2>&1 || true
}

# Build all test images
build_images() {
    cd "$SCRIPT_DIR"
    docker-compose --profile test build --parallel >/dev/null 2>&1
}

# Test distribution
test_distribution() {
    local distro="$1"
    local container="svz-test-${distro}"
    
    cd "$SCRIPT_DIR"
    docker-compose --profile test up -d "$container" >/dev/null 2>&1 || return 1
    sleep 3
    
    case "$distro" in
        ubuntu|debian) docker exec "$container" service ssh start >/dev/null 2>&1 || true ;;
        alpine) docker exec "$container" /usr/sbin/sshd >/dev/null 2>&1 || true ;;
        *) docker exec "$container" /usr/sbin/sshd >/dev/null 2>&1 || true ;;
    esac
    
    sleep 2
    docker exec "$container" su - testuser -c "echo test" >/dev/null 2>&1 || return 1
    docker cp "$SVZ_BINARY" "$container:/tmp/svz" >/dev/null 2>&1
    
    output=$(docker exec "$container" bash -c 'chmod +x /tmp/svz && timeout 30s /tmp/svz install testuser@localhost --password "testpass123" --skip-host-key-check' 2>&1)
    echo "$output" | grep -q -E "(Setting up.*repository|Using sudo|sudo apt|sudo yum|sudo apk|sudo pacman)"
}

# Run tests
run_tests() {
    local distros=("ubuntu" "debian" "alpine" "centos" "fedora" "arch")
    
    # Single distribution mode
    if [ -n "${SINGLE_DISTRO:-}" ]; then
        distros=("$SINGLE_DISTRO")
    fi
    
    local pids=() failed=0 temp_dir="/tmp/svz-e2e-$$"
    
    mkdir -p "$temp_dir"
    
    for distro in "${distros[@]}"; do
        (test_distribution "$distro" && echo "PASS" || echo "FAIL") > "$temp_dir/${distro}.result" &
        pids+=($!)
    done
    
    for pid in "${pids[@]}"; do
        wait "$pid"
    done
    
    for distro in "${distros[@]}"; do
        if [ "$(cat "$temp_dir/${distro}.result" 2>/dev/null)" = "PASS" ]; then
            echo "$distro: PASS"
        else
            echo "$distro: FAIL"
            ((failed++))
        fi
    done
    
    rm -rf "$temp_dir"
    return $failed
}

# Main
main() {
    [ -f "$SVZ_BINARY" ] || { echo "Binary not found. Run 'make build' first."; exit 1; }
    build_images
    run_tests && echo "All tests passed" || { echo "Some tests failed"; exit 1; }
}

trap cleanup EXIT
main "$@"
