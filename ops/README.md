# Operations & Build System
> **Role**: The Builder
> **Language**: Python (Invoke)
> **Path**: `ops/`

## 🛠️ Components
This directory contains the logic for the **Service Release Standard 1.0**.

### `version.py` (Identity)
Calculates the **Immutable Tag**: `v{SEMVER}-{SHA}`.
- Source: `../VERSION` file.
- Git: `git rev-parse --short HEAD`.

### `k8s.py` (Deployment)
Handles the atomic deployment to Kubernetes (Kind).
- **Init**: `inv k8s.init` (Infra).
- **Deploy**: `inv k8s.deploy` (Core).
- **Status**: `inv k8s.status` (Health).

### `core.py` (Compilation)
Handles Go compilation and Docker image building.
- **Build**: `inv core.build` (Returns Tag).

## ⚡ Directives
- **Never tag `latest`** for production.
- **Always pinning** dependencies in `charts/`.
