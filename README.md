# Mycelis Cortex V7.0

**The Recursive Swarm Operating System.**

> [!IMPORTANT]
> **MASTER STATE AUTHORITY**
> This README is the **Single Source of Truth** for project state, architecture, and operational commands.
>
> **Architecture PRD:** The detailed architecture specification lives in 4 focused documents:
> | Document | Load When |
> | :--- | :--- |
> | [Architecture Overview](docs/architecture/OVERVIEW.md) | Planning phases, architectural decisions |
> | [Backend Specification](docs/architecture/BACKEND.md) | Working on Go code, APIs, DB, NATS |
> | [Frontend Specification](docs/architecture/FRONTEND.md) | Working on React/Next.js, Zustand, design |
> | [Operations Manual](docs/architecture/OPERATIONS.md) | Deploying, testing, CI/CD, config |

Mycelis is a "Neural Organism" that orchestrates AI agents to solve complex tasks. Built through 12 phases — from genesis through **Admin Orchestrator**, **Council Activation**, **Trust Economy**, **RAG Persistence**, **Agent Visualization**, **Neural Wiring Edit/Delete**, **Meta-Agent Research**, **Team Management**, and the **Standardized Council Chat API** with CTS-enveloped responses, trust scores, and any council member addressable via HTTP.

## Architecture

### Tier 1: Core (Go 1.26 + Postgres + pgvector)

- **Soma → Axon → Teams → Agents:** Mission activation pipeline with heartbeat + proof-of-work.
- **Standing Teams:** Admin (orchestrator, 17 tools, 5 ReAct iterations) + Council (architect, coder, creative, sentry) — all individually addressable via `POST /api/v1/council/{member}/chat`.
- **Council Chat API:** Standardized CTS-enveloped responses with trust scores, provenance metadata, and tools-used tracking. Dynamic member validation via Soma — add a YAML, restart, done.
- **Runtime Context Injection:** Every agent receives live system state (active teams, NATS topology, MCP servers, cognitive config, interaction protocols) via `InternalToolRegistry.BuildContext()`.
- **Internal Tool Registry:** 17 built-in tools — consult_council, delegate_task, search_memory, remember, recall, file I/O, NATS bus sensing, image generation, and more.
- **Composite Tool Executor:** Unified interface routing tool calls to InternalToolRegistry or MCP ToolExecutorAdapter.
- **MCP Ingress:** Install, manage, and invoke MCP tool servers. Curated library with one-click install.
- **Archivist:** Context engine — SitReps, auto-embed to pgvector (768-dim, nomic-embed-text), semantic search.
- **Governance:** Policy engine with YAML rules, approval queue, trust economy (0.0–1.0 threshold).
- **Cognitive Router:** 6 LLM providers (ollama, vllm, lmstudio, OpenAI, Anthropic, Gemini), profile-based routing, token telemetry.

### Tier 2: Nervous System (NATS JetStream 2.12)

- 30+ topics: heartbeat, audit trace, team internals, council request-reply, sensor ingress, mission DAG.
- All topics use constants from `pkg/protocol/topics.go` — never hardcode.

### Tier 3: The Face (Next.js 16 + React 19 + Zustand 5)

- **Mission Control (`/dashboard`):** Council Chat (member selector dropdown) + OperationsBoard (priority alerts, standing workloads, missions) + Telemetry + Sensors + Cognitive Status.
- **Neural Wiring (`/wiring`):** ArchitectChat + CircuitBoard (ReactFlow) + ToolsPalette + NatsWaterfall. Interactive edit/delete: click agent nodes to modify manifests, delete agents, discard drafts, or terminate active missions.
- **Agent Visualization:** Observable Plot charts (bar, line, area, dot, waffle, tree), Leaflet geo maps, DataTable — rendered inline via ChartRenderer from `MycelisChartSpec`.
- **Team Management (`/teams`):** Browse standing + mission teams, agent roster, delivery targets.
- **Memory Explorer (`/memory`):** Hot/Warm/Cold three-tier browser with semantic search.
- **Settings (`/settings`):** Cognitive Matrix + MCP Tools (with curated library).
- **Visual Protocol:** Midnight Cortex theme — `cortex-bg #09090b`, `cortex-primary #06b6d4` (cyan). Zero `bg-white` in new code.

> Full architecture details: [Architecture Overview](docs/architecture/OVERVIEW.md) | [Backend Spec](docs/architecture/BACKEND.md) | [Frontend Spec](docs/architecture/FRONTEND.md) | [Operations Manual](docs/architecture/OPERATIONS.md)

## Getting Started

> **Detailed guide:** See [Local Dev Workflow](docs/LOCAL_DEV_WORKFLOW.md) for configuration reference, port map, health checks, and troubleshooting.

### 1. Configure Secrets

```bash
cp .env.example .env
# Edit .env — set DB credentials, OLLAMA_HOST, NATS_URL, etc.
# See docs/LOCAL_DEV_WORKFLOW.md for full variable reference.
```

### 2. Boot the Infrastructure

```bash
uvx inv k8s.reset    # Full System Reset (Cluster + Core + DB)
```

### 3. Open the Development Bridge (Terminal 1)

Port-forwards PostgreSQL (5432), NATS (4222), and HTTP (8081) from the Kind cluster to localhost.

```bash
uvx inv k8s.bridge
```

### 4. Initialize the Database

```bash
uvx inv db.migrate    # Apply all 21 migrations (idempotent)
```

### 5. Build + Start the Core Server (Terminal 2)

```bash
uvx inv core.build   # Compile Go binary (first time + after code changes)
uvx inv core.run     # Start server on port 8081 (foreground)
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
- **Env:** `OLLAMA_HOST` in `.env` sets the default Ollama endpoint.

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
| `uvx inv interface.e2e` | Run Playwright E2E tests (requires running servers) |
| `uvx inv interface.check` | Smoke-test running server (9 pages, no light-mode leaks) |
| `uvx inv interface.stop` | Kill dev server |
| `uvx inv interface.clean` | Clear `.next` cache |
| `uvx inv interface.restart` | Full restart: stop → clean → build → dev → check |
| **Database** | |
| `uvx inv db.migrate` | Apply SQL migrations (idempotent) |
| `uvx inv db.reset` | Drop + recreate + migrate |
| `uvx inv db.status` | Show tables + row counts |
| **Infrastructure** | |
| `uvx inv k8s.reset` | Full cluster reset |
| `uvx inv k8s.status` | Cluster status |
| `uvx inv k8s.deploy` | Deploy Helm chart |
| `uvx inv k8s.bridge` | Port-forward NATS, API, Postgres |
| **Cognitive** | |
| `uvx inv cognitive.up` | Start vLLM + Diffusers (full stack) |
| `uvx inv cognitive.status` | Health check providers |
| `uvx inv cognitive.stop` | Kill cognitive processes |
| **CI Pipeline** | |
| `uvx inv ci.check` | Full CI: lint → test → build (with timers) |
| `uvx inv ci.deploy` | Full CI: lint → test → build → Docker → K8s |

### CI Workflows (GitHub Actions)

Three workflows run on push/PR to `main` and `develop`:

| Workflow | Trigger Paths | Checks |
| :--- | :--- | :--- |
| **Core CI** (`core-ci.yaml`) | `core/**`, `ops/core.py` | Go test + coverage, GolangCI-Lint v1.64.5, binary build |
| **Interface CI** (`interface-ci.yaml`) | `interface/**`, `ops/interface.py` | ESLint, `tsc --noEmit`, Vitest, production build |
| **E2E CI** (`e2e-ci.yaml`) | `interface/**`, `core/**` | Build Core + Next.js, start servers, Playwright (Chromium) |

> [!TIP]
> If you run `uv venv` and activate your virtual environment, you can use `inv` directly without `uvx`.

## Frontend Routes

| Route | Description |
| :--- | :--- |
| `/` | Product landing page (marketing) — links to `/dashboard` to launch console |
| `/dashboard` | Mission Control — Council chat (member selector), operations board, telemetry, sensors, cognitive status |
| `/wiring` | Neural Wiring — ArchitectChat + CircuitBoard (edit/delete agents) + NatsWaterfall |
| `/architect` | Redirects to `/wiring` |
| `/teams` | Team Management — browse standing + mission teams, agent roster, delivery targets |
| `/catalogue` | Agent Catalogue — CRUD for agent blueprints |
| `/memory` | Memory Explorer — Hot/Warm/Cold three-tier browser |
| `/approvals` | Governance — approval queue, policy config, team proposals (3 tabs) |
| `/missions/[id]/teams` | Team Actuation — live team drill-down |
| `/settings` | Profile, Teams, Cognitive Matrix, MCP Tools |
| `/settings/brain` | Cognitive Matrix — provider routing grid |
| `/settings/tools` | MCP Tools — server management + curated library |
| `/matrix` | Cognitive Matrix data table |
| `/marketplace` | Skills Market — connector registry |
| `/telemetry` | System Status — infrastructure monitoring |

## Stack Versions (Locked)

| Component | Version | Notes |
| :--- | :--- | :--- |
| Go | 1.26 | Module: `github.com/mycelis/core` |
| Next.js | 16.1.6 | Turbopack, App Router |
| React | 19.2.3 | `"use client"` required for hooks/state |
| Tailwind CSS | v4 | `@import "tailwindcss"`, `@theme` directive |
| ReactFlow | 11.11.4 | Package is `reactflow`, NOT `@xyflow/react` |
| Zustand | 5.0.11 | Single atomic store: `useCortexStore` |
| PostgreSQL | 16 | pgvector/pgvector:16-alpine |
| NATS | 2.12.4 | JetStream enabled |
| Python | >=3.12 | uv managed, invoke for tasks |

## Key Configurations

| Config | Location | Managed Via |
| :--- | :--- | :--- |
| Cognitive (Bootstrap) | `core/config/cognitive.yaml` | UI (`/settings` → Matrix) or YAML |
| Standing Teams | `core/config/teams/*.yaml` | YAML (auto-loaded at startup, council members auto-addressable via API) |
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
| **Local Dev Workflow** | [docs/LOCAL_DEV_WORKFLOW.md](docs/LOCAL_DEV_WORKFLOW.md) — Setup, config reference, port map, troubleshooting |
| **Architecture Overview** | [docs/architecture/OVERVIEW.md](docs/architecture/OVERVIEW.md) — Philosophy, 4-layer anatomy, phases, upcoming roadmap |
| **Backend Specification** | [docs/architecture/BACKEND.md](docs/architecture/BACKEND.md) — Go packages, APIs, DB schema, NATS, execution pipelines |
| **Frontend Specification** | [docs/architecture/FRONTEND.md](docs/architecture/FRONTEND.md) — Routes, components, Zustand, design system |
| **Operations Manual** | [docs/architecture/OPERATIONS.md](docs/architecture/OPERATIONS.md) — Deployment, config, testing, CI/CD |
| **Swarm Operations** | [docs/SWARM_OPERATIONS.md](docs/SWARM_OPERATIONS.md) — Hierarchy, blueprints, activation, teams, tools, governance |
| **API Reference** | [docs/API_REFERENCE.md](docs/API_REFERENCE.md) — Full endpoint table (80+ routes) |
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
uvx inv core.test             # Go unit tests (~112 handler tests)
uvx inv interface.test        # Vitest component tests (~114 tests)
uvx inv interface.e2e         # Playwright E2E specs (20 spec files, requires running servers)
uvx inv interface.check       # HTTP smoke test against running dev server (9 pages)
uvx inv core.smoke            # Governance smoke tests
```

> Full testing documentation: [docs/TESTING.md](docs/TESTING.md)

## Delivered Phases

| Phase | Name | Key Deliverables |
| :--- | :--- | :--- |
| 1–3 | Genesis Build | Core server, NATS, Postgres, ReactFlow, basic UI |
| 4.1–4.6 | Foundation | Zustand store, intent commit, SSE binding, Overseer DAG |
| 4.4 | Governance | Deliverables Tray, Governance Modal (Human-in-the-Loop) |
| 5.0 | Archivist Daemon | NATS buffer → LLM compress → sitreps table |
| 5.1 | SquadRoom | Fractal navigation, Mission Control layout |
| 5.2 | Sovereign UX & Trust | Trust Economy, Telemetry API, AgentNode iconography, TrustSlider |
| 5.3 | RAG & Sensory UI | pgvector, SemanticSearch, SensorLibrary, NatsWaterfall signals |
| 6.0 | Host Internalization | Activation Bridge, SensorAgent, Blueprint Converter, Symbiotic Seed |
| 7.7 | Admin & Council | NATS-only chat, Cognitive Matrix, 17 tools, ToolsPalette |
| 8.0 | Visualization | Observable Plot, Leaflet, DataTable, ChartRenderer |
| 9.0 | Neural Wiring CRUD | WiringAgentEditor, mission edit/delete API, draft/active modes |
| 10.0 | Meta-Agent Research | research_for_blueprint, admin-routed negotiate |
| 11.0 | Team Management | /teams route, TeamCard, TeamDetailDrawer |
| 17.0 | Legacy Migration | Complete cortex-* migration, Vuexy CSS vars removed, landing page rewritten |

> Full phase history with details: [Architecture Overview](docs/architecture/OVERVIEW.md#vi-delivered-phases)

## Upcoming Architecture

Planned phases with detailed specifications are documented in the Architecture Overview:

| Phase | Name | Summary |
| :--- | :--- | :--- |
| 12 | Persistent Agent Memory | Cross-mission memory, semantic recall, memory consolidation daemon |
| 13 | Multi-Agent Collaboration | Intra-team debate protocol, consensus detection, SquadRoom live chat |
| 14 | Hot-Reload Runtime | Live agent goroutine replacement, zero-downtime reconfiguration |
| 15 | Advanced Governance & RBAC | Role enforcement, API keys, audit trail, policy versioning |
| 16 | Distributed Federation | Multi-node NATS, team affinity, cross-instance delegation |
| 18 | Streaming LLM | Token-by-token streaming, SSE relay, mid-stream tool detection |
| 19 | Workflow Templates | Mission template store, one-click instantiation, built-in pipelines |
| 20 | Observability Dashboard | Historical metrics, Prometheus export, agent performance analytics |

> Full roadmap with technical details: [Architecture Overview](docs/architecture/OVERVIEW.md#vii-upcoming-architecture)

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
