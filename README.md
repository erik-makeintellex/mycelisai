# Mycelis Service Network (Gen-2)

> **Current State**: Phase 1 - Migration to "Neural Swarm" (Go Architecture)
> **Agent Context**: This file is the Source of Truth for Project Structure and Tooling.

## 🏗️ Architecture: The Vertical Skeleton

We are migrating from a monolithic Python app to a high-performance distributed Go architecture.

| Component | Status | Path | Description |
| :--- | :--- | :--- | :--- |
| **Brain** (Core) | 🚧 In Progress | `core/` | Go-based Orchestrator (State Registry, NATS Adapter) |
| **Contracts** | ✅ Active | `proto/` | Protobuf definitions (`swarm.proto`) shared by all services |
| **Hyphae** (Nerves) | ⚠️ Legacy | `runner/` | Python-based Agents (Migrating to standalone services) |
| **Interface** (UI) | ⚠️ Legacy | `ui/` | Next.js Frontend (Planned for migration to embedded React) |

---

## 🛠️ Tooling & Standards

We enforce strict tooling to ensure deterministic environments for Humans and Agents.

-   **Dependency Manager**: **`uv`** (Python) and **`go mod`** (Go).
-   **Task Runner**: `scripts/dev.py` (Universal Runner wrapped by Makefile).
-   **Container Engine**: Docker / Podman.

### Quick Start (Universal)
You only need `uv` installed. The runner handles the rest.

```bash
# 1. Start Infrastructure (NATS + Postgres)
make dev-up

# 2. Generate Contracts (Protobuf -> Go)
make proto

# 3. Build the Core Brain
make build-core
```

### Legacy Workflow (Python API + UI)
To run the existing functionality while developing Gen-2:
```bash
make dev-api   # Runs FastAPI on localhost:8000
make dev-ui    # Runs Next.js on localhost:3000
```

---

## 📂 Directory Structure

```text
/
├── core/                  # [Gen-2] The Go Brain
│   ├── cmd/server/        # Entrypoint
│   ├── internal/state/    # In-Memory Agent Registry
│   └── go.mod             # Module: github.com/mycelis/core
├── proto/                 # [Gen-2] Shared Contracts
│   └── swarm/v1/          # swarm.proto
├── deploy/                # [Gen-2] Kubernetes/Docker Manifests
│   ├── charts/            # Helm Charts
│   └── docker/            # Distroless Dockerfiles
├── scripts/               # [Global] Dev Tooling
│   └── dev.py             # Universal Runner (PEP 723)
├── api/                   # [Legacy] FastAPI Backend
├── runner/                # [Legacy] Python Agent Runtime
└── ui/                    # [Legacy] Next.js Frontend
```

---

## 🤖 Agent Directives

**If you are an Agent working on this repo:**
1.  **Architecture**: Respect the separation between `core` (Go) and `runner` (Python). Do not mix them.
2.  **Tooling**: ALWAYS use `uv run` for Python commands. NEVER use `pip` or `python` directly.
3.  **State**: Check `task.md` in the active brain session for immediate objectives.

**Active Specialists:**
*   `spec:arch:01` - System Architect (ADR Owner)
*   `spec:golang:01` - Backend Engineer (Core Implementation)
*   `spec:devops:01` - Infrastructure Engineer (Charts/Docker)

---

## 📚 Documentation
*   [Architecture Deep Dive](architecture.md)
*   [Agent Protocol](proto/swarm/v1/swarm.proto)
