#!/bin/bash
# test/e2e/run_e2e_tests.sh - Ultra-performance E2E testing for superviz.io install command

set -euo pipefail

# ✅ ALWAYS: Set strict error handling
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="$(cd "${SCRIPT_DIR}/../.." && pwd)"
readonly DOCKER_DIR="${SCRIPT_DIR}/docker"
readonly TIMEOUT=300

# Color codes for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly NC='\033[0m' # No Color

# Distribution configurations
declare -A DISTRO_CONFIGS=(
    ["ubuntu"]="2201:testuser:testpass"
    ["debian"]="2202:testuser:testpass"
    ["alpine"]="2203:testuser:testpass"
    ["centos"]="2204:testuser:testpass"
    ["fedora"]="2205:testuser:testpass"
    ["arch"]="2206:testuser:testpass"
)

# ✅ ALWAYS: Validate environment
validate_environment() {
    echo -e "${BLUE}🔍 Validating environment...${NC}"
    
    # Check required commands
    local required_commands=("docker" "docker-compose" "ssh" "nc")
    for cmd in "${required_commands[@]}"; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            echo -e "${RED}❌ Required command '$cmd' not found${NC}" >&2
            exit 1
        fi
    done
    
    # Check if superviz binary exists
    if [[ ! -f "${PROJECT_ROOT}/cmd/svz/main.go" ]]; then
        echo -e "${RED}❌ superviz source not found at ${PROJECT_ROOT}/cmd/svz/main.go${NC}" >&2
        exit 1
    fi
    
    echo -e "${GREEN}✅ Environment validation passed${NC}"
}

# ✅ ALWAYS: Build superviz binary for testing
build_superviz() {
    echo -e "${BLUE}🔧 Building superviz binary...${NC}"
    
    cd "${PROJECT_ROOT}"
    if ! go build -o "${SCRIPT_DIR}/superviz" ./cmd/svz/main.go; then
        echo -e "${RED}❌ Failed to build superviz binary${NC}" >&2
        exit 1
    fi
    
    echo -e "${GREEN}✅ superviz binary built successfully${NC}"
}

# ✅ ALWAYS: Start Docker containers
start_containers() {
    echo -e "${BLUE}🐳 Starting Docker containers...${NC}"
    
    cd "${DOCKER_DIR}"
    
    # Stop any existing containers
    docker-compose down --remove-orphans >/dev/null 2>&1 || true
    
    # Start all containers
    if ! docker-compose up -d --build; then
        echo -e "${RED}❌ Failed to start Docker containers${NC}" >&2
        exit 1
    fi
    
    echo -e "${GREEN}✅ Docker containers started${NC}"
}

# ✅ ALWAYS: Wait for SSH services to be ready
wait_for_ssh() {
    echo -e "${BLUE}⏳ Waiting for SSH services to be ready...${NC}"
    
    for distro in "${!DISTRO_CONFIGS[@]}"; do
        local config="${DISTRO_CONFIGS[$distro]}"
        local port="${config%%:*}"
        
        echo -e "${YELLOW}  Waiting for $distro SSH on port $port...${NC}"
        
        local attempts=0
        local max_attempts=30
        
        while ! nc -z localhost "$port" >/dev/null 2>&1; do
            attempts=$((attempts + 1))
            if [[ $attempts -gt $max_attempts ]]; then
                echo -e "${RED}❌ Timeout waiting for $distro SSH${NC}" >&2
                return 1
            fi
            sleep 2
        done
        
        echo -e "${GREEN}  ✅ $distro SSH ready${NC}"
    done
    
    echo -e "${GREEN}✅ All SSH services ready${NC}"
}

# ✅ ALWAYS: Test install command on a distribution
test_install_on_distro() {
    local distro="$1"
    local config="${DISTRO_CONFIGS[$distro]}"
    local port user password
    
    IFS=':' read -r port user password <<< "$config"
    
    echo -e "${BLUE}🧪 Testing install on $distro...${NC}"
    
    # Create temporary SSH key for testing
    local ssh_key="${SCRIPT_DIR}/test_key_${distro}"
    ssh-keygen -t rsa -b 2048 -f "$ssh_key" -N "" -q
    
    # Copy public key to container using sshpass (password authentication)
    # First, install sshpass if needed
    if ! command -v sshpass >/dev/null 2>&1; then
        echo -e "${YELLOW}  Installing sshpass for password authentication...${NC}"
        if command -v apt-get >/dev/null 2>&1; then
            sudo apt-get update >/dev/null 2>&1 && sudo apt-get install -y sshpass >/dev/null 2>&1
        elif command -v apk >/dev/null 2>&1; then
            sudo apk add --no-cache sshpass >/dev/null 2>&1
        fi
    fi
    
    # Setup SSH key authentication
    local ssh_opts="-o StrictHostKeyChecking=no -o UserKnownHostsFile=/dev/null -o LogLevel=ERROR"
    
    # Copy SSH key to remote host
    if command -v sshpass >/dev/null 2>&1; then
        sshpass -p "$password" ssh-copy-id $ssh_opts -p "$port" "${user}@localhost" -i "${ssh_key}.pub" >/dev/null 2>&1 || {
            # Fallback: manual key copy
            sshpass -p "$password" ssh $ssh_opts -p "$port" "${user}@localhost" \
                "mkdir -p ~/.ssh && chmod 700 ~/.ssh" >/dev/null 2>&1
            sshpass -p "$password" scp $ssh_opts -P "$port" "${ssh_key}.pub" \
                "${user}@localhost:~/.ssh/authorized_keys" >/dev/null 2>&1
            sshpass -p "$password" ssh $ssh_opts -p "$port" "${user}@localhost" \
                "chmod 600 ~/.ssh/authorized_keys" >/dev/null 2>&1
        }
    else
        echo -e "${YELLOW}  Using password authentication (sshpass not available)${NC}"
    fi
    
    # Test the install command
    local output_file="${SCRIPT_DIR}/test_output_${distro}.log"
    local success=false
    
    if timeout $TIMEOUT "${SCRIPT_DIR}/superviz" install \
        --ssh-key="${ssh_key}" \
        --ssh-port="$port" \
        --timeout=60s \
        --skip-host-key-check \
        "${user}@localhost" > "$output_file" 2>&1; then
        success=true
    fi
    
    # Check results
    if [[ "$success" == true ]] && grep -q "Repository setup completed successfully" "$output_file"; then
        echo -e "${GREEN}  ✅ $distro install test PASSED${NC}"
        # Show key output lines
        echo -e "${BLUE}    Key output:${NC}"
        grep -E "(Connected to|Detected distribution|Repository setup completed|You can now install)" "$output_file" | \
            sed 's/^/      /'
    else
        echo -e "${RED}  ❌ $distro install test FAILED${NC}"
        echo -e "${RED}    Error output:${NC}"
        tail -10 "$output_file" | sed 's/^/      /'
        return 1
    fi
    
    # Cleanup
    rm -f "$ssh_key" "${ssh_key}.pub" "$output_file"
    
    return 0
}

# ✅ ALWAYS: Run all distribution tests
run_all_tests() {
    echo -e "${BLUE}🚀 Running E2E tests on all distributions...${NC}"
    
    local failed_distros=()
    local total_tests=0
    local passed_tests=0
    
    for distro in "${!DISTRO_CONFIGS[@]}"; do
        total_tests=$((total_tests + 1))
        
        if test_install_on_distro "$distro"; then
            passed_tests=$((passed_tests + 1))
        else
            failed_distros+=("$distro")
        fi
    done
    
    echo
    echo -e "${BLUE}📊 Test Results Summary:${NC}"
    echo -e "  Total tests: $total_tests"
    echo -e "  Passed: ${GREEN}$passed_tests${NC}"
    echo -e "  Failed: ${RED}$((total_tests - passed_tests))${NC}"
    
    if [[ ${#failed_distros[@]} -gt 0 ]]; then
        echo -e "${RED}  Failed distributions: ${failed_distros[*]}${NC}"
        return 1
    else
        echo -e "${GREEN}  All tests passed! 🎉${NC}"
        return 0
    fi
}

# ✅ ALWAYS: Cleanup function
cleanup() {
    local exit_code=$?
    
    echo -e "${BLUE}🧹 Cleaning up...${NC}"
    
    # Stop containers
    cd "${DOCKER_DIR}" 2>/dev/null && docker-compose down --remove-orphans >/dev/null 2>&1 || true
    
    # Remove test artifacts
    rm -f "${SCRIPT_DIR}/superviz"
    rm -f "${SCRIPT_DIR}"/test_key_*
    rm -f "${SCRIPT_DIR}"/test_output_*.log
    
    echo -e "${GREEN}✅ Cleanup completed${NC}"
    exit $exit_code
}

# ✅ ALWAYS: Main function
main() {
    echo -e "${BLUE}🎯 Starting superviz.io E2E Tests${NC}"
    echo -e "${BLUE}===================================${NC}"
    
    validate_environment
    build_superviz
    start_containers
    wait_for_ssh
    
    if run_all_tests; then
        echo -e "${GREEN}🎉 All E2E tests completed successfully!${NC}"
        exit 0
    else
        echo -e "${RED}❌ Some E2E tests failed${NC}"
        exit 1
    fi
}

# ✅ ALWAYS: Set cleanup trap
trap cleanup EXIT INT TERM

# Execute main function with all arguments
main "$@"
