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

## 📝 Logs

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
