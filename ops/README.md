# Operations & Build System
> **Role**: The Builder
> **Language**: Python (Invoke)
> **Path**: `ops/`

## Runner Contract

- Real task execution uses `uv run inv <namespace>.<task>`.
- Compatibility probe only: `uvx --from invoke inv -l`.
- Do not use bare `uvx inv ...`.
- App-tied management logic stays in Python task modules; PowerShell is wrapper-only when the host needs it.

## 🛠️ Components
This directory contains the logic for the **Service Release Standard 1.0**.

### `version.py` (Identity)
Calculates the **Immutable Tag**: `v{SEMVER}-{SHA}`.
- Source: `../VERSION` file.
- Git: `git rev-parse --short HEAD`.

### `k8s.py` (Deployment)
Handles the atomic deployment to Kubernetes (Kind).
- **Init**: `uv run inv k8s.init` (Infra).
- **Deploy**: `uv run inv k8s.deploy` (Core).
- **Status**: `uv run inv k8s.status` (Health).

### `core.py` (Compilation)
Handles Go compilation and Docker image building.
- **Build**: `uv run inv core.build` (Returns Tag).

### `auth.py` (Local Operator Auth)
Keeps local API-key development access aligned.
- **Dev Key**: `uv run inv auth.dev-key`

### `logging.py` (Logging Gates)
Enforces logging contract quality checks before delivery.
- **Schema**: `uv run inv logging.check-schema` (event taxonomy + docs coverage)
- **Topics**: `uv run inv logging.check-topics` (no hardcoded `swarm.*` outside constants)

### `quality.py` (Code Hygiene Gates)
Enforces max-lines policy on hot-path files with temporary no-regression caps.
- **Max Lines**: `uv run inv quality.max-lines --limit 350`

### `lifecycle.py` (Local Stack Control)
Owns deterministic local bring-up, teardown, and deep health checks.
- **Up**: `uv run inv lifecycle.up --frontend`
- **Down**: `uv run inv lifecycle.down`
- **Health**: `uv run inv lifecycle.health`
- **Memory Restart**: `uv run inv lifecycle.memory-restart --frontend`

### `ci.py` (Delivery Gates)
Delivery-focused validation, runner checks, and release preflight.
- **Entrypoint Check**: `uv run inv ci.entrypoint-check`
- **Baseline**: `uv run inv ci.baseline`
- **Release Preflight**: `uv run inv ci.release-preflight --strict-toolchain`

### `misc.py` (Team Coordination)
Central architect sync path and utility task surfaces.
- **Architecture Sync**: `uv run inv team.architecture-sync`

## ⚡ Directives
- **Never tag `latest`** for production.
- **Always pinning** dependencies in `charts/`.
