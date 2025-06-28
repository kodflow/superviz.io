# Shell Script Optimization Rules (.sh files ONLY)

## POSIX Shell Priority (Mandatory)

### Default Shell Selection

```bash
#!/bin/sh
# ✅ ALWAYS: Use POSIX /bin/sh as default shebang
# Maximum compatibility across systems
# Minimal resource usage
# Available on all Unix-like systems

# ❌ AVOID: Unless absolutely necessary
#!/bin/bash
#!/bin/zsh
#!/bin/dash
```

### Shell Feature Detection

```bash
# ✅ ALWAYS: Check shell capabilities before using advanced features
check_shell_features() {
    # Test for bash-specific features
    if [ -n "$BASH_VERSION" ]; then
        HAS_BASH=1
    else
        HAS_BASH=0
    fi

    # Test for array support
    if command -v bash >/dev/null 2>&1; then
        HAS_ARRAYS=1
    else
        HAS_ARRAYS=0
    fi
}
```

## Memory Optimization

### Variable Management

```bash
# ✅ ALWAYS: Unset large variables when done
process_large_data() {
    large_data=$(cat large_file.txt)

    # Process data
    echo "$large_data" | process_command

    # Free memory immediately
    unset large_data
}

# ✅ ALWAYS: Use local variables in functions
process_file() {
    local file="$1"
    local temp_data
    local result

    temp_data=$(cat "$file")
    result=$(echo "$temp_data" | transform)
    echo "$result"

    # Variables automatically freed when function exits
}
```

### Efficient String Operations

```bash
# ✅ ALWAYS: Use parameter expansion instead of external commands
filename="/path/to/file.txt"

# Good: Parameter expansion (no external process)
basename="${filename##*/}"
dirname="${filename%/*}"
extension="${filename##*.}"
name="${filename%.*}"

# ❌ BAD: External commands (memory + process overhead)
basename=$(basename "$filename")
dirname=$(dirname "$filename")
extension=$(echo "$filename" | cut -d'.' -f2)
```

## Disk I/O Optimization

### Minimize File Operations

```bash
# ✅ ALWAYS: Read files once and store in memory
read_config_once() {
    if [ -z "$CONFIG_LOADED" ]; then
        CONFIG_DATA=$(cat config.txt)
        CONFIG_LOADED=1
        export CONFIG_DATA CONFIG_LOADED
    fi
}

# ✅ ALWAYS: Use here documents for multi-line output
generate_config() {
    cat > config.txt << 'EOF'
# Configuration file
setting1=value1
setting2=value2
setting3=value3
EOF
}

# ✅ ALWAYS: Batch file operations
process_multiple_files() {
    # Bad: Multiple separate operations
    # for file in *.txt; do
    #     cat "$file" >> combined.txt
    # done

    # Good: Single operation
    cat *.txt > combined.txt
}
```

### Efficient Log Writing

```bash
# ✅ ALWAYS: Buffer log writes
LOG_BUFFER=""
LOG_MAX_SIZE=1024

log_message() {
    local message="$1"
    local timestamp=$(date '+%Y-%m-%d %H:%M:%S')

    LOG_BUFFER="${LOG_BUFFER}${timestamp}: ${message}\n"

    # Flush when buffer is full
    if [ ${#LOG_BUFFER} -gt $LOG_MAX_SIZE ]; then
        printf "%s" "$LOG_BUFFER" >> logfile.log
        LOG_BUFFER=""
    fi
}

# ✅ ALWAYS: Flush buffer on exit
cleanup_logs() {
    if [ -n "$LOG_BUFFER" ]; then
        printf "%s" "$LOG_BUFFER" >> logfile.log
    fi
}
trap cleanup_logs EXIT
```

## CPU Optimization

### Command Substitution Efficiency

```bash
# ✅ ALWAYS: Use $() instead of backticks
result=$(command arg1 arg2)

# ❌ AVOID: Backticks (harder to nest, less efficient)
result=`command arg1 arg2`

# ✅ ALWAYS: Minimize subshells
# Bad: Multiple subshells
count=$(echo "$data" | wc -l)
size=$(echo "$data" | wc -c)

# Good: Single operation with multiple outputs
{
    echo "$data" | wc -l
    echo "$data" | wc -c
} | {
    read count
    read size
}
```

### Efficient Loops and Conditions

```bash
# ✅ ALWAYS: Use built-in test conditions
if [ -f "$file" ] && [ -r "$file" ]; then
    process_file "$file"
fi

# ✅ ALWAYS: Avoid unnecessary command substitutions in loops
# Bad: Command substitution in each iteration
for file in $(ls *.txt); do
    process "$file"
done

# Good: Direct globbing
for file in *.txt; do
    [ -f "$file" ] || continue
    process "$file"
done

# ✅ ALWAYS: Use case for multiple string comparisons
check_file_type() {
    local file="$1"
    case "$file" in
        *.txt) echo "text file" ;;
        *.log) echo "log file" ;;
        *.conf|*.cfg) echo "config file" ;;
        *) echo "unknown file" ;;
    esac
}
```

## Error Handling (POSIX Compatible)

### Robust Error Management

```bash
#!/bin/sh
# ✅ ALWAYS: Set strict error handling
set -e  # Exit on error
set -u  # Exit on undefined variable
set -f  # Disable globbing

# ✅ ALWAYS: Define cleanup function
cleanup() {
    local exit_code=$?

    # Remove temporary files
    rm -f "$TEMP_FILE" 2>/dev/null

    # Kill background processes
    [ -n "$BACKGROUND_PID" ] && kill "$BACKGROUND_PID" 2>/dev/null

    exit $exit_code
}
trap cleanup EXIT INT TERM

# ✅ ALWAYS: Check command success explicitly
run_command() {
    local cmd="$1"
    local error_msg="$2"

    if ! $cmd; then
        echo "Error: $error_msg" >&2
        return 1
    fi
}

# ✅ ALWAYS: Validate input parameters
validate_params() {
    if [ $# -lt 1 ]; then
        echo "Usage: $0 <required_param>" >&2
        exit 1
    fi

    if [ ! -f "$1" ]; then
        echo "Error: File '$1' does not exist" >&2
        exit 1
    fi
}
```

## POSIX Compatibility Patterns

### Portable Constructs

```bash
# ✅ ALWAYS: Use POSIX-compliant syntax
# Good: POSIX parameter expansion
remove_extension() {
    local filename="$1"
    echo "${filename%.*}"
}

# Good: POSIX string operations
contains_substring() {
    local string="$1"
    local substring="$2"
    case "$string" in
        *"$substring"*) return 0 ;;
        *) return 1 ;;
    esac
}

# ✅ ALWAYS: Use portable command options
# Good: Portable find
find . -name "*.txt" -type f

# ❌ AVOID: GNU-specific options
# find . -name "*.txt" -type f -printf "%p\n"
```

### Cross-Platform Path Handling

```bash
# ✅ ALWAYS: Handle path separators portably
normalize_path() {
    local path="$1"
    # Remove duplicate slashes
    echo "$path" | sed 's|//*|/|g'
}

# ✅ ALWAYS: Use portable temporary files
create_temp_file() {
    local prefix="$1"
    local temp_dir="${TMPDIR:-/tmp}"
    local temp_file="${temp_dir}/${prefix}.$$"

    # Ensure temp file is unique
    while [ -e "$temp_file" ]; do
        temp_file="${temp_dir}/${prefix}.$$.$(date +%s)"
    done

    touch "$temp_file"
    echo "$temp_file"
}
```

## Performance Monitoring

### Script Profiling

```bash
# ✅ ALWAYS: Add timing for performance-critical sections
time_function() {
    local start_time=$(date +%s)

    # Function logic here
    "$@"

    local end_time=$(date +%s)
    local duration=$((end_time - start_time))
    echo "Function completed in ${duration}s" >&2
}

# ✅ ALWAYS: Monitor resource usage
monitor_resources() {
    if command -v ps >/dev/null 2>&1; then
        ps -o pid,ppid,rss,vsz,pcpu,comm -p $$
    fi
}
```

## Concurrent Operations

### Safe Background Processing

```bash
# ✅ ALWAYS: Limit concurrent processes
MAX_JOBS=4
CURRENT_JOBS=0

process_file_async() {
    local file="$1"

    # Wait if too many jobs running
    while [ $CURRENT_JOBS -ge $MAX_JOBS ]; do
        wait_for_job_completion
    done

    {
        process_file "$file"
        echo "Completed: $file"
    } &

    CURRENT_JOBS=$((CURRENT_JOBS + 1))
}

wait_for_job_completion() {
    if jobs >/dev/null 2>&1; then
        # Wait for any background job
        wait
        CURRENT_JOBS=0
    fi
}
```

## Security Best Practices

### Input Sanitization

```bash
# ✅ ALWAYS: Sanitize file paths
sanitize_path() {
    local path="$1"

    # Remove dangerous characters
    path=$(echo "$path" | tr -d ';&|`$()')

    # Prevent directory traversal
    case "$path" in
        *../*|*/../*|../*)
            echo "Error: Invalid path" >&2
            return 1
            ;;
    esac

    echo "$path"
}

# ✅ ALWAYS: Quote variables to prevent injection
safe_exec() {
    local command="$1"
    local arg="$2"

    # Always quote arguments
    "$command" "$arg"
}
```

## Configuration Management

### Environment Variable Handling

```bash
# ✅ ALWAYS: Provide defaults for environment variables
CONFIG_FILE="${CONFIG_FILE:-/etc/default/myapp}"
LOG_LEVEL="${LOG_LEVEL:-info}"
MAX_RETRIES="${MAX_RETRIES:-3}"

# ✅ ALWAYS: Validate environment variables
validate_config() {
    # Check required variables
    for var in CONFIG_FILE LOG_LEVEL; do
        eval "value=\$$var"
        if [ -z "$value" ]; then
            echo "Error: $var is not set" >&2
            exit 1
        fi
    done

    # Validate numeric values
    case "$MAX_RETRIES" in
        ''|*[!0-9]*)
            echo "Error: MAX_RETRIES must be a number" >&2
            exit 1
            ;;
    esac
}
```

## Debugging and Maintenance

### Debug Mode Support

```bash
# ✅ ALWAYS: Support debug mode
DEBUG="${DEBUG:-0}"

debug_log() {
    if [ "$DEBUG" = "1" ]; then
        echo "DEBUG: $*" >&2
    fi
}

# ✅ ALWAYS: Provide verbose mode
VERBOSE="${VERBOSE:-0}"

verbose_log() {
    if [ "$VERBOSE" = "1" ]; then
        echo "INFO: $*" >&2
    fi
}

# Enable debug tracing when needed
if [ "$DEBUG" = "1" ]; then
    set -x
fi
```

## Script Template

### Standard Script Structure

```bash
#!/bin/sh
# Script: script_name.sh
# Description: Script description
# Version: 1.0.0
# Author: Your Name
# Usage: script_name.sh [options] <arguments>

# ✅ ALWAYS: Set strict mode
set -e
set -u
set -f

# ✅ ALWAYS: Define constants
readonly SCRIPT_NAME="$(basename "$0")"
readonly SCRIPT_DIR="$(dirname "$0")"
readonly VERSION="1.0.0"

# ✅ ALWAYS: Initialize variables
DEBUG="${DEBUG:-0}"
VERBOSE="${VERBOSE:-0}"
DRY_RUN="${DRY_RUN:-0}"

# ✅ ALWAYS: Define usage function
usage() {
    cat << EOF
Usage: $SCRIPT_NAME [OPTIONS] <argument>

Description of what the script does.

OPTIONS:
    -h, --help      Show this help message
    -v, --verbose   Enable verbose output
    -d, --debug     Enable debug mode
    -n, --dry-run   Show what would be done without executing

EXAMPLES:
    $SCRIPT_NAME file.txt
    $SCRIPT_NAME -v file.txt

EOF
}

# ✅ ALWAYS: Parse command line arguments
parse_args() {
    while [ $# -gt 0 ]; do
        case "$1" in
            -h|--help)
                usage
                exit 0
                ;;
            -v|--verbose)
                VERBOSE=1
                ;;
            -d|--debug)
                DEBUG=1
                set -x
                ;;
            -n|--dry-run)
                DRY_RUN=1
                ;;
            -*)
                echo "Error: Unknown option $1" >&2
                usage >&2
                exit 1
                ;;
            *)
                break
                ;;
        esac
        shift
    done
}

# ✅ ALWAYS: Define cleanup function
cleanup() {
    local exit_code=$?

    # Cleanup temporary files
    [ -n "${TEMP_FILES:-}" ] && rm -f $TEMP_FILES

    # Kill background processes
    [ -n "${BACKGROUND_PIDS:-}" ] && kill $BACKGROUND_PIDS 2>/dev/null || true

    exit $exit_code
}
trap cleanup EXIT INT TERM

# ✅ ALWAYS: Main function
main() {
    parse_args "$@"

    # Validate requirements
    validate_environment

    # Main logic here
    echo "Script execution completed successfully"
}

# ✅ ALWAYS: Validate environment
validate_environment() {
    # Check required commands
    for cmd in cat sed grep; do
        if ! command -v "$cmd" >/dev/null 2>&1; then
            echo "Error: Required command '$cmd' not found" >&2
            exit 1
        fi
    done
}

# Execute main function with all arguments
main "$@"
```

## Validation Commands

### Script Quality Checks

```bash
# ✅ ALWAYS: Validate shell syntax
shellcheck script.sh

# ✅ ALWAYS: Test POSIX compliance
checkbashisms script.sh

# ✅ ALWAYS: Performance testing
time ./script.sh test_input

# ✅ ALWAYS: Memory usage monitoring
/usr/bin/time -v ./script.sh test_input
```
