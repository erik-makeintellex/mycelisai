# Mycelis Service Network
> **Codename**: "The Conscious Organism"
> **Phase**: 6.0 (Expansion & Hardening)
> **Identity**: [`v0.6.0`](VERSION)
> **Status**: [🟢 Online](http://localhost:3000)

## 🧠 The Vision
Mycelis is not a distributed system; it is a single, cohesive cybernetic organism extending the user's will. It adheres to the **V6 Cortex Architecture** and enforces strict **Immutable Identity** for all cells.

---

## 📚 Documentation Hub
This README is the **Source of Truth**. Below are the authoritative references for the platform.

### Architecture & Standards
- 📘 **[The Cortex Blueprint (V6)](myceles_v6_arch.md)**: The active Product Requirement Document and Architecture spec.
- 📜 **[Release Standard 1.0](#-service-development--release-standard)**: The enforced rules for build, tag, and release.
- ⚖️ **[Governance Policy](docs/governance.md)**: The "Conscience" defining high-stakes gates.
- 📝 **[Logging Standard](docs/logging.md)**: The "Memory" structure for events.
- 🏚️ *[Legacy V5 Arch](mycelist_v5_arch.md)*: (Reference only).

### System Connectivity Map
| Component | Role | Tech | Documentation |
| :--- | :--- | :--- | :--- |
| **[Core](core/)** | **The Brain** | Go | [README](core/README.md) |
| **[Interface](interface/)** | **The Eyes** | Next.js | [README](interface/README.md) |
| **[Ops](ops/)** | **The Builder** | Python (Invoke) | [README](ops/README.md) |
| **[SDK](sdk/python/)** | **The Nerves** | Python | Standard Library |
| **[Infra](charts/)** | **The Body** | Helm | [Values](charts/mycelis-core/values.yaml) |

---

## 📜 Service Development & Release Standard
**Version:** 1.0 | **Status:** ENFORCED

### 1. The "Immutable Identity"
Every running container must be traceable to a specific commit SHA.
* **Format**: `v{MAJOR}.{MINOR}.{PATCH}-{SHORT_SHA}`
* **Source**: [`VERSION`](VERSION) file + Git SHA.
* **Example**: `v0.6.0-42f9958`

### 2. Infrastructure Specification
We enforce a **Single Persistence** policy ("The Hippocampus").
* **PostgreSQL**: `bitnami/postgresql:16.1.0-debian-11-r12` (Pinned).
* **NATS**: `nats:2.10.7-alpine3.19` (Pinned).
* **Ollama**: `0.1.20` (External Logic).

---

## ⚡ Platform Management
We use `uv` and `invoke` to manage the organism.

### Quick Start
```bash
# 1. Initialize Body (Kind Cluster + Postgres + NATS)
uv run inv k8s.init

# 2. Build & Deploy Brain (Standard 1.0)
# (Auto-calculates tag -> Builds -> Deploys)
uv run inv k8s.deploy

# 3. Check Vital Signs
uv run inv k8s.status
```

### Development Access
```bash
# Open Neural Bridge (Port Forward NATS:4222 & API:8080)
uv run inv k8s.bridge

# Launch Visual Cortex (UI) -> http://localhost:3000
uv run inv interface.dev
```

### Release Workflow
1.  **Commit Code**.
2.  **Run**: `uv run inv k8s.deploy`.
3.  **Verify**: `uv run inv k8s.status` to confirm the new `IMAGE` tag matches your SHA.
