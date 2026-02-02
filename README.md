# Mycelis Service Network
> **Codename**: "The Conscious Organism"
> **Phase**: 6.0 (Expansion & Hardening)
> **Identity**: [`v0.6.0`](VERSION)
> **Status**: [🟢 Online](http://localhost:3000)

## 🧠 The Vision
Mycelis is an **Autonomous Neuro-Symbolic Service Network** designed to function as a cohesive cybernetic organism. By unifying infrastructure, state, and intelligence into a single "nervous system," it eliminates the friction of traditional distributed systems.

The platform adheres to the **V6 Cortex Architecture**, enforcing:
- **Biological Unity**: Services act as "cells" with a shared purpose, not isolated microservices.
- **Immutable Identity**: Strict, cryptographic traceability for every component.
- **Adaptive Intelligence**: Embedded cognitive loops for self-healing and decision making.

---

## 📚 Documentation Hub
This README is the **Source of Truth**. Below are the authoritative references for the platform.

### Architecture & Standards
- 📘 **[The Cortex Blueprint (V6)](myceles_v6_arch.md)**: The active Product Requirement Document and Architecture spec.
- 📜 **[Release Standard 1.0](#-service-development--release-standard)**: The enforced rules for build, tag, and release.
- ⚖️ **[Governance Policy](core/config/policy.yaml)**: The "Conscience" defining high-stakes gates.
- 📝 **[Logging Standard](docs/logging.md)**: The "Memory" structure for events.
- 🏚️ *[Legacy V5 Arch](mycelist_v5_arch.md)*: (Reference only).

### System Connectivity Map
| Component | Role | Tech | Documentation |
| :--- | :--- | :--- | :--- |
| **[Core](core/)** | **The Brain** | Go | [README](core/README.md) |
| **[Interface](interface/)** | **The Eyes** | Next.js | [README](interface/README.md) |
| **[CLI](cli/)** | **The Synapse** | Python/Typer | `myc --help` |
| **[Ops](ops/)** | **The Builder** | Python (Invoke) | [README](ops/README.md) |
| **[SDK](sdk/python/)** | **The Nerves** | Python | Standard Library |
| **[Infra](charts/)** | **The Body** | Helm | [Values](charts/mycelis-core/values.yaml) |

---

## Configuration

Configuration is managed via environment variables. Copy `.env.example` to `.env` to customize your local setup.

### Authentication & Secrets
The following variables control authentication and connectivity:
- `DB_USER` / `DB_PASSWORD`: Support for PostgreSQL connection.
- `DB_NAME`: The target database (default: `cortex`).
- `NATS_URL`: Connection string for the NATS cluster.

See `.env.example` for the full list of configurable options.

## Development

### ⚡ Synaptic Injector (CLI)
Mycelis includes a headless control plane for rapid interaction.
```bash
# 1. Install Dependencies
cd cli && uv sync

# 2. Check Pulse
uv run python cli/main.py status

# 3. Scan Network
uv run python cli/main.py scan
```

### ⚡ Platform Management
We use `uv` and `invoke` to manage the organism.

#### Quick Start (Orchestration)
The `dev` namespace provides "One-Click" lifecycle management.

```bash
# Power Up: Clusters -> Deps -> Build -> Deploy -> Bridge
uv run inv dev.up

# Power Down: Destroy Cluster & Cleanup
uv run inv dev.down
```

### Manual Control
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
