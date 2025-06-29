# E2E Testing for superviz.io Install Command

This directory contains end-to-end tests for the `install` command of superviz.io.

## Architecture

The e2e test system uses Docker to create containers for different Linux distributions with SSH enabled, then tests the `install` command on each of them.

### Supported Distributions

- **Ubuntu 22.04** - SSH enabled with testuser
- **Debian 12** - SSH enabled with testuser  
- **Alpine 3.18** - SSH enabled with testuser
- **CentOS Stream 9** - SSH enabled with testuser
- **Fedora 39** - SSH enabled with testuser
- **Arch Linux** - SSH enabled with testuser

### Default Credentials

- **Username**: `testuser`
- **Password**: `testpass123`
- **Root password**: `rootpass`

The `testuser` has sudo privileges without password for automation.

## Prerequisites

1. **Docker**: Must be installed and accessible
2. **Static Binary**: A static Go binary must be built for cross-distribution compatibility

## Usage

### Via Makefile (recommended)

```bash
# Setup e2e environment (build images)
make e2e-setup

# Tests on all distributions
make e2e-test

# Test on a specific distribution
make e2e-test-single DISTRO=ubuntu

# Cleanup
make e2e-clean
```

### Direct script execution

```bash
# Test on all distributions
./test/e2e/final_e2e_test.sh

# Test on a specific distribution
./test/e2e/final_e2e_test.sh ubuntu
```

## Docker Configuration

### Individual Containers

Each distribution has its own Dockerfile in `test/e2e/docker/`:

- `ubuntu.Dockerfile` - Ubuntu 22.04 with SSH
- `debian.Dockerfile` - Debian 12 with SSH  
- `alpine.Dockerfile` - Alpine 3.18 with SSH
- `centos.Dockerfile` - CentOS Stream 9 with SSH
- `fedora.Dockerfile` - Fedora 39 with SSH
- `arch.Dockerfile` - Arch Linux with SSH

### Container Execution Model

The tests use **docker exec** approach instead of external SSH connections:

1. **Start** each distribution container individually
2. **Wait** for SSH service initialization
3. **Copy** the static binary to the container
4. **Execute** tests inside the container via `docker exec`
5. **Test** SSH authentication within the container
6. **Cleanup** container after test completion

## Tests Performed

For each distribution, the test:

1. **Builds** the static superviz binary (if not exists)
2. **Starts** the Docker container for the distribution
3. **Waits** for the SSH service to be ready
4. **Copies** the binary to `/tmp/svz` inside the container
5. **Executes** tests via `docker exec`:
   - SSH connection establishment
   - Password authentication verification
   - Sudo command execution
   - Repository setup process (until external download)

## Implementation Details

### Binary Compatibility

- Uses **static binary** (`svz_linux_arm64_static`) for maximum compatibility
- Compatible with both glibc (Ubuntu, Debian, CentOS, Fedora) and musl (Alpine)
- No external dependencies required

### SSH Testing Strategy

- **Ubuntu/Debian/Alpine/Arch**: Uses `sshpass` for password authentication testing
- **CentOS/Fedora**: Uses `su - testuser` approach (sshpass not available)
- Tests validate:
  - SSH service availability
  - Password authentication
  - User environment setup
  - Sudo privilege escalation

### Error Handling

- Robust error detection and reporting
- Automatic container cleanup on test failure
- Detailed logging for debugging
- Graceful handling of distribution-specific differences

## Logs and Debug

Test output is displayed in real-time during execution. For debugging specific issues:

```bash
# Run with verbose output (built into the script)
./test/e2e/final_e2e_test.sh ubuntu

# Manual container debugging
docker run -d --name debug-ubuntu svz-test-ubuntu
docker exec -it debug-ubuntu /bin/bash
```

## Authentication

The system currently tests password-based authentication within the containers:

### Password Authentication (used in tests)

Within each container, the test validates:

- SSH service availability
- Password authentication with `testuser:testpass123`
- Sudo privileges for repository configuration

## Current Implementation Status

✅ **Completed Features:**

- Multi-distribution Docker testing (6 distributions)
- Static binary compilation for cross-platform compatibility
- Automated SSH service testing within containers
- Repository setup validation (until external download)
- Robust error handling and cleanup
- Makefile integration
- Comprehensive logging

🔄 **Test Coverage:**

- SSH connection establishment ✅
- Password authentication ✅  
- User environment setup ✅
- Sudo privilege escalation ✅
- Package manager detection ✅
- Repository configuration ✅
- Package installation (partial - until external download)

## Adding a New Distribution

To add a new distribution:

1. **Create** a new Dockerfile in `test/e2e/docker/`
2. **Follow** the existing pattern with SSH setup and testuser creation
3. **Add** the distribution to the `DISTRIBUTIONS` array in `final_e2e_test.sh`
4. **Test** the new distribution: `./test/e2e/final_e2e_test.sh newdistro`

## Security

⚠️ **Important**: These configurations are for testing purposes only:

- SSH with password authentication enabled
- Host key verification disabled in tests
- Sudo privileges without password
- Hardcoded credentials

**Never use in production!**

## Archive

Old development scripts have been moved to `archive/` directory for reference. The current production script is `final_e2e_test.sh`.
