# Superviz.io [![Reference](https://pkg.go.dev/badge/github.com/kodflow/superviz.io.svg)](https://pkg.go.dev/github.com/kodflow/superviz.io) [![Latest Stable Version](https://img.shields.io/github/v/tag/kodflow/superviz.io?label=version)](https://github.com/kodflow/superviz.io/releases/latest) [![CI](https://img.shields.io/github/actions/workflow/status/kodflow/superviz.io/ci.yml?label=CI)](https://github.com/kodflow/superviz.io/actions/workflows/ci.yml)

[![Quality Gate Status](https://sonarcloud.io/api/project_badges/measure?project=kodflow_superviz.io&metric=alert_status)](https://sonarcloud.io/summary/new_code?id=kodflow_superviz.io)
[![Bugs](https://sonarcloud.io/api/project_badges/measure?project=kodflow_superviz.io&metric=bugs)](https://sonarcloud.io/summary/new_code?id=kodflow_superviz.io)
[![Code Smells](https://sonarcloud.io/api/project_badges/measure?project=kodflow_superviz.io&metric=code_smells)](https://sonarcloud.io/summary/new_code?id=kodflow_superviz.io)
[![Coverage](https://sonarcloud.io/api/project_badges/measure?project=kodflow_superviz.io&metric=coverage)](https://sonarcloud.io/summary/new_code?id=kodflow_superviz.io)
[![Duplicated Lines (%)](https://sonarcloud.io/api/project_badges/measure?project=kodflow_superviz.io&metric=duplicated_lines_density)](https://sonarcloud.io/summary/new_code?id=kodflow_superviz.io)
[![Reliability Rating](https://sonarcloud.io/api/project_badges/measure?project=kodflow_superviz.io&metric=reliability_rating)](https://sonarcloud.io/summary/new_code?id=kodflow_superviz.io)
[![Security Rating](https://sonarcloud.io/api/project_badges/measure?project=kodflow_superviz.io&metric=security_rating)](https://sonarcloud.io/summary/new_code?id=kodflow_superviz.io)
[![Technical Debt](https://sonarcloud.io/api/project_badges/measure?project=kodflow_superviz.io&metric=sqale_index)](https://sonarcloud.io/summary/new_code?id=kodflow_superviz.io)
[![Maintainability Rating](https://sonarcloud.io/api/project_badges/measure?project=kodflow_superviz.io&metric=sqale_rating)](https://sonarcloud.io/summary/new_code?id=kodflow_superviz.io)
[![Vulnerabilities](https://sonarcloud.io/api/project_badges/measure?project=kodflow_superviz.io&metric=vulnerabilities)](https://sonarcloud.io/summary/new_code?id=kodflow_superviz.io)

## 🧠 Project Goal

**Superviz.io** is a universal service manager designed to operate as `PID 1` inside containers or as a complement to `systemd` on traditional hosts. It centralizes provisioning, service launching, observability and logging - whether in Docker, VMs, bare-metal or cloud-init environments.

### Key Features

- Acts as an **init system** (`PID 1`) inside containers or as a complement to **systemd** on traditional systems
- Handles **post-provisioning**: auto-installs required binaries (Vault, Consul, etc.)
- Supports **pre-runtime build** for Docker layer caching
- Launches services with configuration, environment and dependencies
- Monitors services and performs automatic restarts if needed
- Manages **environment variables** with schema validation and conditional injection
- Verifies integrity of critical files via **hashes** (SHA-256, etc.)
- Exposes **OpenTelemetry metrics** (service state, errors, resource usage)
- Handles **log management** with rotation, max size and auto-cleanup
- Planned: **hot-reload** through configuration diffing and targeted service restarts
- Advanced health model with **granular service status tracking**
- Optional agent installation via **SSH** when credentials/key are provided
- Follows a **desired state model**: continuously evaluates system state and executes actions to converge to the target configuration

### Desired State Provisioning

Superviz.io follows a **desired state model**: you declare the **target state** of services, files, and runtime conditions. At startup and during runtime, Superviz.io continuously evaluates the current state and **executes actions** to reach or restore the desired one.

Examples:

- If a binary is missing, it will be downloaded and verified.
- If a service is not running but should be, it will be launched.
- If configuration has drifted, a restart or reload is triggered.
- If health-checks fail, restart logic or alerts apply depending on thresholds.

This model ensures **idempotency**, **self-healing**, and better **predictability** across environments.

## 🚠 Typical Use Cases

- Docker container with a smart, self-sufficient **entrypoint**
- Infrastructure provisioning using **cloud-init**, Terraform or Ansible
- Lightweight distributions or systems without `systemd`
- Optimised Docker image builds using pre-compiled binaries
- Replacement for legacy init systems like `supervisord`, `runit` or `s6`

## 📦 Getting Started

### Installation

```bash
curl -Lo /usr/local/bin/superviz https://superviz.io/download/latest/superviz-linux-amd64
chmod +x /usr/local/bin/superviz
```

## 💼 Enterprise Edition & Trial

Enterprise features (HA clustering, RBAC, audit UI, etc.) are available in an **EE build**.
A 30-day evaluation binary is published on the [releases page](https://github.com/kodflow/superviz.io/releases).
Activate it with an evaluation key or run without `LICENSE_KEY` for a limited trial.

> Production use of EE requires a commercial licence issued by Making Codes.

## 🥉 Community vs Commercial Edition

| Feature                                      | Community Edition (CE) ✅ | Enterprise Edition (EE) 💼 |
| -------------------------------------------- | ------------------------- | -------------------------- |
| `PID 1` init system (bare-metal / container) | ✅ Yes                    | ✅ Yes                     |
| Service orchestration with restart/reload    | ✅ Yes                    | ✅ Yes                     |
| Configuration via `superviz.yaml`            | ✅ Yes                    | ✅ Yes                     |
| Environment injection / validation           | ✅ Yes                    | ✅ Yes                     |
| Logs rotation and metrics (OpenTelemetry)    | ✅ Yes                    | ✅ Yes                     |
| Multi-check health model                     | ✅ Yes                    | ✅ Yes                     |
| SSH agent provisioning                       | ✅ Yes                    | ✅ Yes                     |
| 🔒 Centralized web dashboard                 | ❌                        | ✅ Included                |
| 🔒 Role-based access control (RBAC)          | ❌                        | ✅ Included                |
| 🔒 Multi-node clustering / HA failover       | ❌                        | ✅ Included                |
| 🔒 Audit logs & compliance exports           | ❌                        | ✅ Included                |
| 🔒 License verification & key activation     | ❌                        | ✅ Required                |
| Commercial SLA (3-day response)              | ❌                        | ✅ Included with support   |

📝 **Want to try Enterprise features?** Create a free account at [https://superviz.io](https://superviz.io) to request your 30-day trial licence key and download the EE binary.

## 📈 Observability

Superviz.io exposes metrics via **OpenTelemetry**:

- **Pull mode** - Prometheus-compatible `/metrics`
- **Push mode** - OTLP exporter (HTTP/gRPC)

Configurable in `superviz.yaml`.

Example metrics: CPU/RAM per service • uptime • crash loops.

### Health-check Status Model

| Status       | Meaning                                            |
| ------------ | -------------------------------------------------- |
| `starting`   | Process launched, not yet verified (`gracePeriod`) |
| `running`    | Process active at OS level                         |
| `ready`      | App live and able to serve traffic                 |
| `degraded`   | Running but partially failing (latency, 5xx, …)    |
| `unhealthy`  | Health-checks failed, process still alive          |
| `crashed`    | Process exited unexpectedly                        |
| `restarting` | Auto-restart in progress                           |
| `stopped`    | Stopped voluntarily                                |
| `dead`       | Cannot be started / permanently failed             |

Combine multiple checks (TCP, HTTP, command) in `superviz.yaml`.

## �️ Development & Testing

### Building from Source

The project uses **Bazel** as the primary build system for fast, reproducible builds with advanced caching.

```bash
# Clone the repository
git clone https://github.com/kodflow/superviz.io.git
cd superviz.io

# Build cross-platform binaries for all supported targets
make build

# Available binaries in .dist/:
# - svz_linux_amd64
# - svz_linux_arm64
# - svz_darwin_amd64
# - svz_darwin_arm64
# - svz_windows_amd64.exe
# - svz_windows_arm64.exe

# Run tests (all unit tests with Bazel)
make test

# Format code (Go, Terraform, YAML)
make fmt
```

#### Advanced Bazel Commands

For more control over the build process:

```bash
# Build specific target
bazel build //cmd/svz:svz_linux_amd64

# Build all targets
bazel build //cmd/svz:all

# Run tests with detailed output
bazel test //... --test_output=all

# Query available targets
bazel query //cmd/svz:all

# Clean build cache
bazel clean
```

#### Build Features

- **Cross-compilation**: Native support for Linux, macOS, Windows (amd64/arm64)
- **Static binaries**: CGO disabled for maximum portability
- **Size optimization**: Stripped symbols (`-s -w` equivalent)
- **Hermetic builds**: Reproducible with Bazel's sandboxing
- **Build metadata**: Version, commit, timestamp injected automatically via Bazel stamping
- **Dynamic injection**: Build metadata values are calculated at build time, not hardcoded
- **Environment support**: Custom version/metadata via environment variables (e.g., `VERSION=1.2.3 make build`)

### Install Command Testing

The `svz install` command supports SSH-based repository installation on remote systems:

```bash
# Install with SSH key authentication
svz install user@hostname -i ~/.ssh/id_rsa

# Install with password authentication (for automation)
svz install user@hostname --password mypassword

# Install with custom SSH port
svz install user@hostname -p 2222 --skip-host-key-check
```

### E2E Testing with Docker

Full end-to-end testing across multiple Linux distributions:

```bash
# Setup test containers (requires Docker)
make e2e-setup

# Run tests on all distributions (Ubuntu, Debian, Alpine, CentOS, Fedora, Arch)
make e2e-test

# Test single distribution
make e2e-test-single DISTRO=ubuntu

# Cleanup test environment
make e2e-clean
```

For more details on E2E testing, see [test/e2e/README.md](./test/e2e/README.md).

## �📝 Logs

Structured logs with:

- Rotation (size/date)
- Max files retention
- Optional compression

OTel-Logs export planned.

## ❌ What Superviz.io **is not**

- A container runtime (`containerd`, Docker)
- An orchestrator (no scheduler/clustering)
- A remote log collector (use Promtail, Fluent-bit, …)
- A sandbox manager (`chroot`, `jail` unsupported)

## 📜 License

This project is released under a **non-commercial license**. You may use, modify and redistribute it **only for personal, educational or non-profit purposes**.

> Commercial use requires a separate agreement and licence key.

If you modify the code, you must publish your changes in a public repository under the same licence.

### 🧾 Additional Legal Information

- 🔒 [Non-Commercial License](./LICENSE.md)
- 💼 [Commercial License Terms](./COMMERCIAL-LICENSE.md)
- 🛡️ [Security Policy](./SECURITY.md)
- 📚 [Documentation complète](https://superviz.io/docs)

## 🌐 Contact

- Website : [https://superviz.io](https://superviz.io)
