# Superviz.io

## ✨ Project Goal

**Superviz.io** is a universal service manager designed to operate as `PID 1` inside containers or as a complement to `systemd` on traditional hosts. It centralizes provisioning, service launching, observability and logging - whether in Docker, VMs, bare‑metal or cloud‑init environments.

### Key Features

* Acts as an **init system** (`PID 1`) inside containers or as a complement to **systemd** on traditional systems
* Handles **post‑provisioning**: auto‑installs required binaries (Vault, Consul, …)
* Supports **pre‑runtime build** for Docker layer caching
* Launches services with configuration, environment and dependencies
* Monitors services and performs automatic restarts if needed
* Manages **environment variables** with schema validation and conditional injection
* Verifies integrity of critical files via **hashes** (SHA‑256, etc.)
* Exposes **OpenTelemetry metrics** (service state, errors, resource usage)
* Handles **log management** with rotation, max size and auto‑cleanup
* Planned: **hot‑reload** through configuration diffing and targeted service restarts
* Advanced health model with **granular service status tracking**
* Optional agent installation via **SSH** when credentials/key are provided

## 🔢 Typical Use Cases

* Docker container with a smart, self‑sufficient **entrypoint**
* Infrastructure provisioning using **cloud‑init**, Terraform or Ansible
* Lightweight distributions or systems without `systemd`
* Optimised Docker image builds using pre‑compiled binaries
* Replacement for legacy init systems like `supervisord`, `runit` or `s6`

## 🚀 Getting Started

### Installation

```bash
curl -Lo /usr/local/bin/superviz https://superviz.io/download/latest/superviz-linux-amd64
chmod +x /usr/local/bin/superviz
```

## 🏢 Enterprise Edition & Trial

Enterprise features (HA clustering, RBAC, audit UI, etc.) are available in an **EE build**.
A 30‑day evaluation binary is published on the [releases page](https://github.com/kodflow/superviz.io/releases).
Activate it with an evaluation key or run without `LICENSE_KEY` for a limited trial.

> Production use of EE requires a commercial licence issued by Making Codes.

## 🧩 Community vs Commercial Edition

| Feature                                      | Community Edition (CE) ✅ | Enterprise Edition (EE) 💼 |
| -------------------------------------------- | ------------------------ | -------------------------- |
| `PID 1` init system (bare-metal / container) | ✅ Yes                    | ✅ Yes                      |
| Service orchestration with restart/reload    | ✅ Yes                    | ✅ Yes                      |
| Configuration via `superviz.yaml`            | ✅ Yes                    | ✅ Yes                      |
| Environment injection / validation           | ✅ Yes                    | ✅ Yes                      |
| Logs rotation and metrics (OpenTelemetry)    | ✅ Yes                    | ✅ Yes                      |
| Multi-check health model                     | ✅ Yes                    | ✅ Yes                      |
| SSH agent provisioning                       | ✅ Yes                    | ✅ Yes                      |
| 🔒 Centralized web dashboard                 | ❌                        | ✅ Included                 |
| 🔒 Role-based access control (RBAC)          | ❌                        | ✅ Included                 |
| 🔒 Multi-node clustering / HA failover       | ❌                        | ✅ Included                 |
| 🔒 Audit logs & compliance exports           | ❌                        | ✅ Included                 |
| 🔒 License verification & key activation     | ❌                        | ✅ Required                 |
| Commercial SLA (3-day response)              | ❌                        | ✅ Included with support    |

📝 **Want to try Enterprise features?** Create a free account at [https://superviz.io](https://superviz.io) to request your 30-day trial licence key and download the EE binary.

## 📊 Observability

Superviz.io exposes metrics via **OpenTelemetry**:

* **Pull mode** - Prometheus‐compatible `/metrics`
* **Push mode** - OTLP exporter (HTTP/gRPC)

Configurable in `superviz.yaml`.

Example metrics: CPU/RAM per service • uptime • crash loops.

### Health‑check Status Model

| Status       | Meaning                                            |
| ------------ | -------------------------------------------------- |
| `starting`   | Process launched, not yet verified (`gracePeriod`) |
| `running`    | Process active at OS level                         |
| `ready`      | App live and able to serve traffic                 |
| `degraded`   | Running but partially failing (latency, 5xx, …)    |
| `unhealthy`  | Health‑checks failed, process still alive          |
| `crashed`    | Process exited unexpectedly                        |
| `restarting` | Auto‑restart in progress                           |
| `stopped`    | Stopped voluntarily                                |
| `dead`       | Cannot be started / permanently failed             |

Combine multiple checks (TCP, HTTP, command) in `superviz.yaml`.

## 📃 Logs

Structured logs with:

* Rotation (size/date)
* Max files retention
* Optional compression

OTel‑Logs export planned.

## 🚫 What Superviz.io **is not**

* A container runtime (`containerd`, Docker)
* An orchestrator (no scheduler/clustering)
* A remote log collector (use Promtail, Fluent‑bit, …)
* A sandbox manager (`chroot`, `jail` unsupported)

## 🎓 License

This project is released under a **non‑commercial license**. You may use, modify and redistribute it **only for personal, educational or non‑profit purposes**.

> Commercial use requires a separate agreement and licence key.

If you modify the code, you must publish your changes in a public repository under the same licence.

### 📑 Additional Legal Information

* 🔒 [Non‑Commercial License](./LICENSE.md)
* 💼 [Commercial License Terms](./COMMERCIAL-LICENSE.md)
* 🛡️ [Security Policy](./SECURITY.md)

## 💬 Contact

* Website : [https://superviz.io](https://superviz.io)
* Email : [TODO](mailto:TODO)
