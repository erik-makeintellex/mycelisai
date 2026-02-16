# Mycelis Cortex V7.7 (The Orchestrator)

**The Recursive Swarm Operating System.**

> [!IMPORTANT]
> **MASTER STATE AUTHORITY**
> The Single Source of Truth for the project state, architectural delta, and immediate goals is:
> **[mycelis-6-2.md](mycelis-6-2.md)**
>
> **Agents & Humans:** Always consult the Master State document before making architectural decisions. This README provides general usage instructions, but the Codex defines "What is True".

Mycelis is a "Neural Organism" that orchestrates AI agents to solve complex tasks. V7.7 introduces the **Admin Orchestrator**, **Council Activation**, **Runtime Context Injection**, **Cognitive Matrix UI**, and **MCP Library**.

## Architecture

- **Tier 1: Core (Go 1.26 + Postgres + pgvector)**
  - **Soma → Axon → Teams → Agents:** Mission activation pipeline with heartbeat + proof-of-work.
  - **Standing Teams:** Admin (orchestrator, 17 tools, 5 ReAct iterations) + Council (architect, coder, creative, sentry).
  - **Runtime Context Injection:** Every agent receives live system state (active teams, NATS topology, MCP servers, cognitive config, interaction protocols) via `InternalToolRegistry.BuildContext()`.
  - **Internal Tool Registry:** 17 built-in tools — consult_council, delegate_task, search_memory, remember, recall, file I/O, NATS bus sensing, image generation, and more.
  - **Composite Tool Executor:** Unified interface routing tool calls to InternalToolRegistry or MCP ToolExecutorAdapter.
  - **MCP Ingress:** Install, manage, and invoke MCP tool servers. Curated library with one-click install.
  - **Archivist:** Context engine — SitReps, auto-embed to pgvector, semantic search.
  - **Governance:** Policy engine with YAML rules, approval queue, trust economy (0.0–1.0 threshold).

- **Tier 2: Nervous System (NATS JetStream)**
  - Heartbeat, audit trace, SCIP (Protobuf), council request-reply.

- **Tier 3: The Face (Next.js 16 + React 19 + Zustand 5)**
  - **Mission Control:** Admin Chat + Team Explorer + Telemetry + Sensors + Governance.
  - **Neural Wiring:** ArchitectChat + CircuitBoard (ReactFlow) + ToolsPalette + NatsWaterfall.
  - **Settings:** Cognitive Matrix + MCP Tools (with curated library).
  - **Visual Protocol:** Vuexy Dark palette — zero zinc/slate classes.

> Full architecture details: [Swarm Operations](docs/SWARM_OPERATIONS.md) | [Cognitive Architecture](docs/COGNITIVE_ARCHITECTURE.md) | [API Reference](docs/API_REFERENCE.md)

## Getting Started

### 1. Configure Secrets

```bash
cp .env.example .env
# Edit .env to set POSTGRES_USER, POSTGRES_PASSWORD, etc.
```

### 2. Boot the Infrastructure

```bash
uvx inv k8s.reset    # Full System Reset (Cluster + Core + DB)
```

### 3. Open the Development Bridge (Terminal 1)

Port-forwards PostgreSQL, NATS, and HTTP from the Kind cluster to localhost.

```bash
uvx inv k8s.bridge
```

### 4. Initialize the Database

```bash
uvx inv db.migrate    # Apply all migrations (idempotent)
```

### 5. Start the Core Server (Terminal 2)

```bash
uvx inv core.run
```

### 6. Launch the Cortex Console (Terminal 3)

```bash
uvx inv interface.install   # First time only
uvx inv interface.dev       # Start dev server on port 3000
```

Open [http://localhost:3000](http://localhost:3000).

### 7. Configure the Cognitive Engine

- **UI:** `/settings` → **Cognitive Matrix** tab — change provider routing, configure endpoints.
- **MCP:** `/settings` → **MCP Tools** tab — install servers from curated library or manually.
- **YAML:** Edit `core/config/cognitive.yaml` directly.

## Developer Orchestration

**Prerequisites:** [uv](https://github.com/astral-sh/uv) and [Docker](https://www.docker.com/).

Run from `scratch/` root using `uvx inv`:

| Command | Description |
| :--- | :--- |
| **Core** | |
| `uvx inv core.build` | Compile Go binary + Docker image |
| `uvx inv core.test` | Run unit tests (`go test ./...`) |
| `uvx inv core.run` | Run Core locally (foreground) |
| `uvx inv core.stop` | Kill running Core process |
| `uvx inv core.restart` | Stop + Run |
| `uvx inv core.smoke` | Governance smoke tests |
| **Interface** | |
| `uvx inv interface.dev` | Start Next.js dev server (Turbopack) |
| `uvx inv interface.build` | Production build |
| `uvx inv interface.test` | Run Vitest unit tests |
| `uvx inv interface.check` | Smoke-test running server |
| `uvx inv interface.stop` | Kill dev server |
| `uvx inv interface.clean` | Clear `.next` cache |
| `uvx inv interface.restart` | Full restart: stop → clean → build → dev → check |
| **Database** | |
| `uvx inv db.migrate` | Apply SQL migrations (idempotent) |
| `uvx inv db.reset` | Drop + recreate + migrate |
| `uvx inv db.status` | Show tables |
| **Infrastructure** | |
| `uvx inv k8s.reset` | Full cluster reset |
| `uvx inv k8s.status` | Cluster status |
| `uvx inv k8s.deploy` | Deploy Helm chart |
| `uvx inv k8s.bridge` | Port-forward NATS, API, Postgres |
| **CI Pipeline** | |
| `uvx inv ci.deploy` | Full CI: lint → test → build → check |

> [!TIP]
> If you run `uv venv` and activate your virtual environment, you can use `inv` directly without `uvx`.

## Frontend Routes

| Route | Description |
| :--- | :--- |
| `/` | Mission Control — Admin chat, broadcast, teams, telemetry, sensors, proposals |
| `/wiring` | Neural Wiring — ArchitectChat + CircuitBoard + NatsWaterfall |
| `/catalogue` | Agent Catalogue — CRUD for agent blueprints |
| `/memory` | Memory Explorer — Hot/Warm/Cold three-tier browser |
| `/approvals` | Governance — approve/reject agent actions, policy config |
| `/missions/[id]/teams` | Team Actuation — live team drill-down |
| `/settings` | Profile, Teams, Cognitive Matrix, MCP Tools |
| `/dashboard` | KPI deck, MatrixGrid, LogStream |
| `/architect` | Full Workspace (alias for `/wiring`) |
| `/matrix` | Cognitive Matrix grid |
| `/marketplace` | Skills Market — connector registry |
| `/telemetry` | System Status — infrastructure monitoring |

## Key Configurations

| Config | Location | Managed Via |
| :--- | :--- | :--- |
| Cognitive (Bootstrap) | `core/config/cognitive.yaml` | UI (`/settings` → Matrix) or YAML |
| Standing Teams | `core/config/teams/*.yaml` | YAML (auto-loaded at startup) |
| MCP Servers | Database | UI (`/settings` → MCP Tools) or API |
| Governance Policy | `core/config/policy.yaml` | UI (`/approvals` → Policy tab) or YAML |
| MCP Library | `core/config/mcp-library.yaml` | YAML (curated registry) |

### Environment Variables

| Variable | Default | Description |
| :--- | :--- | :--- |
| `MYCELIS_API_HOST` | `localhost` | Core API host |
| `MYCELIS_API_PORT` | `8081` | Core API port |
| `MYCELIS_INTERFACE_HOST` | `localhost` | Next.js dev server host |
| `MYCELIS_INTERFACE_PORT` | `3000` | Next.js dev server port |

## Documentation Hub

| Topic | Document |
| :--- | :--- |
| **Swarm Operations** | [docs/SWARM_OPERATIONS.md](docs/SWARM_OPERATIONS.md) — Hierarchy, blueprints, activation, teams, tools, governance |
| **API Reference** | [docs/API_REFERENCE.md](docs/API_REFERENCE.md) — Full endpoint table |
| **Cognitive Architecture** | [docs/COGNITIVE_ARCHITECTURE.md](docs/COGNITIVE_ARCHITECTURE.md) — Providers, profiles, matrix UI, embedding |
| **Memory Specs** | [docs/architecture/DIRECTIVE_MEMORY_SERVICE.md](docs/architecture/DIRECTIVE_MEMORY_SERVICE.md) — Event store & memory architecture |
| **Governance** | [docs/governance.md](docs/governance.md) — Policy enforcement, approvals, security |
| **Registry** | [core/internal/registry/README.md](core/internal/registry/README.md) — Connector marketplace |
| **Telemetry** | [docs/logging.md](docs/logging.md) — SCIP log structure |
| **Testing** | [docs/TESTING.md](docs/TESTING.md) — Unit, integration, smoke protocols |
| **Core API** | [core/README.md](core/README.md) — Go service architecture |
| **CLI** | [cli/README.md](cli/README.md) — `myc` command-line tool |
| **Interface** | [interface/README.md](interface/README.md) — Next.js frontend architecture |

## Verification

```bash
uvx inv ci.test           # All tests (Core + Interface)
uvx inv core.test         # Core only
uvx inv interface.test    # Interface only (Vitest)
uvx inv core.smoke        # Governance smoke tests
```

## Branching Strategy

Trunk-based development with ephemeral feature branches.

| Type | Prefix | Example |
| :--- | :--- | :--- |
| Production | `main` | `main` |
| Feature | `feat/` | `feat/neural-router` |
| Fix | `fix/` | `fix/memory-leak` |
| Chore | `chore/` | `chore/infra-reset` |
| Docs | `docs/` | `docs/api-spec` |

1. Branch off `main`.
2. Commit often (conventional commits).
3. Merge via squash.
