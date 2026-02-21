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

Mycelis is a "Neural Organism" that orchestrates AI agents to solve complex tasks. Built through 17 phases — from genesis through **Admin Orchestrator**, **Council Activation**, **Trust Economy**, **RAG Persistence**, **Agent Visualization**, **Neural Wiring Edit/Delete**, **Meta-Agent Research**, **Team Management**, the **Standardized Council Chat API** with CTS-enveloped responses, **Soma Identity** with persistent chat memory, artifact pipeline, and in-character organism personality, **Conversation Memory & Scheduled Teams**, a **Natural Human Interface** (label translation layer), **Phase 0 Security Containment** (API key auth, filesystem sandboxing, MCP lockdown), and **Agent & Provider Orchestration** (brain provenance pipeline, mutation-gated proposals, brains management, unified lifecycle ops).

## Architecture

### Tier 1: Core (Go 1.26 + Postgres + pgvector)

- **Soma → Axon → Teams → Agents:** Mission activation pipeline with heartbeat + proof-of-work.
- **Standing Teams:** Admin/Soma (orchestrator, 18 tools, 10 ReAct iterations, persistent identity) + Council (architect, coder, creative, sentry) — all individually addressable via `POST /api/v1/council/{member}/chat`.
- **Council Chat API:** Standardized CTS-enveloped responses with trust scores, provenance metadata, and tools-used tracking. Dynamic member validation via Soma — add a YAML, restart, done.
- **Runtime Context Injection:** Every agent receives live system state (active teams, NATS topology, MCP servers, cognitive config, interaction protocols) via `InternalToolRegistry.BuildContext()`.
- **Internal Tool Registry:** 20 built-in tools — consult_council, delegate_task, search_memory, remember, recall, broadcast, file I/O (workspace-sandboxed), NATS bus sensing, image generation, summarize_conversation, research_for_blueprint, and more.
- **Composite Tool Executor:** Unified interface routing tool calls to InternalToolRegistry or MCP ToolExecutorAdapter.
- **MCP Ingress:** Install, manage, and invoke MCP tool servers. Curated library with one-click install. Raw install endpoint disabled (Phase 0 security) — library-only installs enforced.
- **Archivist:** Context engine — SitReps, auto-embed to pgvector (768-dim, nomic-embed-text), semantic search.
- **Governance:** Policy engine with YAML rules, approval queue, trust economy (0.0–1.0 threshold).
- **Cognitive Router:** 6 LLM providers (ollama, vllm, lmstudio, OpenAI, Anthropic, Gemini), profile-based routing, token telemetry. Brain provenance tracks which provider/model executed each response.
- **CE-1 Templates:** Orchestration template engine with intent proofs, confirm tokens (15min TTL), and audit trail. Chat-to-Answer (read-only) and Chat-to-Proposal (mutation-gated) execution modes.
- **Brains API:** Provider management with location/data_boundary/usage_policy/roles_allowed. Enable/disable toggle and policy updates persist to `cognitive.yaml`.

### Tier 2: Nervous System (NATS JetStream 2.12)

- 30+ topics: heartbeat, audit trace, team internals, council request-reply, sensor ingress, mission DAG.
- All topics use constants from `pkg/protocol/topics.go` — never hardcode.

### Tier 3: The Face (Next.js 16 + React 19 + Zustand 5)

- **Mission Control (`/dashboard`):** The admin's primary command interface — see [Mission Control Reference](#mission-control-reference) below.
- **Neural Wiring (`/wiring`):** ArchitectChat + CircuitBoard (ReactFlow) + ToolsPalette + NatsWaterfall. Interactive edit/delete: click agent nodes to modify manifests, delete agents, discard drafts, or terminate active missions.
- **Agent Visualization:** Observable Plot charts (bar, line, area, dot, waffle, tree), Leaflet geo maps, DataTable — rendered inline via ChartRenderer from `MycelisChartSpec`.
- **Team Management (`/teams`):** Browse standing + mission teams, agent roster, delivery targets.
- **Memory Explorer (`/memory`):** Hot/Warm/Cold three-tier browser with semantic search.
- **Settings (`/settings`):** Brains (provider management with remote-enable confirmation), Cognitive Matrix, MCP Tools (curated library), Users (stub auth).
- **Mode Ribbon:** Always-visible status bar showing current execution mode, active brain (with local/remote badge), and governance state.
- **Proposal Blocks:** Inline chat cards for mutation-gated actions — shows intent, tools, risk level, confirm/cancel buttons wired to CE-1 confirm token flow.
- **Orchestration Inspector:** Expandable audit panel showing template ID, intent proof, confirm token, and execution mode for each chat response.
- **Visual Protocol:** Midnight Cortex theme — `cortex-bg #09090b`, `cortex-primary #06b6d4` (cyan). Zero `bg-white` in new code.

### Mission Control Reference

Mission Control (`/dashboard`) is the admin's primary interface — a resizable two-panel layout where 80% of work happens in conversation.

```
┌───────────────────────────────────────────────────────────────┐
│  MISSION CONTROL          SIGNAL: LIVE       [+ NEW] [⚙]     │
├─── Telemetry Row ─────────────────────────────────────────────┤
│  [Goroutines: 42] [Heap: 18MB] [System: 52MB] [LLM: 3.2t/s] │
├───────────────────────────────────────────────────────────────┤
│                                                               │
│   ADMIN / COUNCIL CHAT  (55% — resizable)                     │
│   ┌───────────────────────────────────────────────────────┐   │
│   │  Council target selector: [ARCHITECT ▾]               │   │
│   │  Rich messages: markdown, code blocks, tables, links  │   │
│   │  Inline artifacts: charts, images, audio, data        │   │
│   │  Trust score badges + tool-use pills                  │   │
│   │  /all prefix or broadcast toggle for swarm-wide msgs  │   │
│   └───────────────────────────────────────────────────────┘   │
│                                                               │
├═══════════════════ drag to resize ════════════════════════════─┤
│                                                               │
│   OPS OVERVIEW  (45% — collapsible)                           │
│   ┌─────────────┐ ┌────────────┐ ┌──────────┐ ┌────────────┐ │
│   │ SYSTEM      │ │ ALERTS     │ │ TEAMS    │ │ MCP TOOLS  │ │
│   │ Text: ●     │ │ GOV x2     │ │ admin    │ │ fs  ●      │ │
│   │ Media: ●    │ │ DONE x1    │ │ council  │ │ fetch ●    │ │
│   │ Sensors:3/5 │ │            │ │          │ │ +brave     │ │
│   └── ↗ ───────┘ └────────────┘ └── ↗ ────┘ └── ↗ ───────┘ │
│   ┌──────────────────────────────────────────────────────────┐│
│   │ ACTIVE MISSIONS (full width)                             ││
│   │ mission-abc ●LIVE  2T/6A    mission-xyz ●DONE  1T/3A    ││
│   └── ↗ ────────────────────────────────────────────────────┘│
└───────────────────────────────────────────────────────────────┘
```

#### Layout

| Zone | Component | Description |
| :--- | :--- | :--- |
| **Header** | `MissionControl` | Signal status (SSE live/offline), NEW MISSION button (→ `/wiring`), Settings gear (→ `/settings`) |
| **Telemetry Row** | `TelemetryRow` | 4 sparkline cards — Goroutines, Heap, System Memory, LLM Tokens/s. Polls `/api/v1/telemetry/compute` every 5s. Shows offline banner after 3 failures. |
| **Chat (top)** | `MissionControlChat` | The primary interaction surface — see Chat section below |
| **Resize Handle** | custom drag | Pointer-event drag handler. Split ratio persisted to localStorage (`mission-control-split`). Clamped 25%–80%. |
| **Ops Overview (bottom)** | `OpsOverview` | Responsive auto-fit grid of compact dashboard cards — see Ops section below |
| **Signal Drawer** | `SignalDetailDrawer` | Right-side slide-over for signal inspection (type badge, metadata grid, raw JSON). Opened by clicking alert rows. |

#### Chat Panel — Rich Message Rendering

The chat renders council/agent responses as **full markdown** (via `react-markdown` + `remark-gfm`):

| Content Type | Rendering |
| :--- | :--- |
| Text | Headings, bold, italic, strikethrough, paragraphs |
| Links | Styled with external-link icon, open in new tab |
| Code blocks | Mono font, cortex-themed background, scrollable |
| Inline code | Colored pill (`cortex-primary`) |
| Tables | GFM tables with header row, scrollable overflow |
| Lists | Ordered/unordered with proper indentation |
| Blockquotes | Left border accent in `cortex-primary` |
| Images | Inline `<img>` with max-height constraint |

**Inline artifacts** are rendered below the message text when the response includes `artifacts[]`:

| Artifact Type | Viewer |
| :--- | :--- |
| `chart` | Observable Plot via `ChartRenderer` (bar, line, area, dot, waffle, tree, geo, table) |
| `image` | `<img>` from URL or base64 data URI |
| `code` | Syntax block with copy-to-clipboard button |
| `audio` | HTML5 `<audio>` player |
| `data` / `document` | Expandable/collapsible JSON or text preview with copy |
| `file` | Compact reference card with external link |

**Tool-use pills** appear after messages showing which internal/MCP tools the agent invoked during its ReAct loop (e.g. `read_file`, `consult_council`, `search_memory`).

#### Chat Capabilities

| Feature | Details |
| :--- | :--- |
| **Council targeting** | Dropdown selector: Admin, Architect, Coder, Creative, Sentry. Each has its own system prompt, tools, and specialization. |
| **Broadcast mode** | Toggle or `/all` prefix — sends message to ALL active teams via NATS |
| **File I/O** | Admin + council agents can `read_file` and `write_file` within the workspace sandbox (`MYCELIS_WORKSPACE`, default `./workspace`). Paths must resolve inside the boundary — symlink escapes are detected. Max 1MB per write. Sentry is read-only. |
| **Tool access** | 20 internal tools: consult_council, delegate_task, search_memory, remember, recall, broadcast, publish_signal, read_signals, read_file, write_file, generate_image, research_for_blueprint, generate_blueprint, list_teams, list_missions, get_system_status, list_available_tools, list_catalogue, store_artifact, summarize_conversation |
| **MCP tools** | Any installed MCP server tools are also available (filesystem, fetch, brave-search, etc.) |
| **Trust scores** | Each response carries a CTS trust score (0.0–1.0), displayed as a colored badge |
| **Multi-turn** | Full conversation history is forwarded to the agent — maintains context across turns |
| **Chat memory** | Conversation persists to localStorage (survives page refresh, 200-message cap). Use `clearMissionChat` to reset. |
| **Broadcast replies** | Broadcast collects responses from all active teams via NATS request-reply (60s timeout per team) |
| **Artifact pipeline** | Tool results with artifacts (images, charts, code) flow through CTS envelope to frontend for inline rendering |

#### Ops Overview Cards

| Card | Data Source | Deep Link | Refresh |
| :--- | :--- | :--- | :--- |
| **System Status** | `GET /api/v1/cognitive/status` + `GET /api/v1/sensors` | `/settings/brain` | 15s / 60s |
| **Priority Alerts** | SSE signal stream (governance_halt, error, task_complete, artifact) | Signal Detail Drawer (click row) | Real-time |
| **Standing Teams** | `GET /api/v1/teams/detail` (filtered: `type === "standing"`) | `/teams` | 10s |
| **MCP Tools** | `GET /api/v1/mcp/servers` | `/settings/tools` | On mount |
| **Active Missions** *(full-width below grid)* | `GET /api/v1/missions` + `GET /api/v1/teams/detail` | `/missions/{id}/teams` per row | 15s / 10s |

Top 4 cards sit in an auto-fit grid; Active Missions spans full width below for better output readability. Each card header has an ↗ icon linking to its detail page. The MCP card also shows a **Recommended** banner for `brave-search` and `github` if not installed.

#### MCP Bootstrap (Zero-Config)

On first server boot, two MCP servers are automatically installed and connected:

| Server | Purpose | Config |
| :--- | :--- | :--- |
| `filesystem` | Read/write files from a mounted data directory | Path: `DATA_DIR` env (default `./workspace`) |
| `fetch` | HTTP fetch — hit any URL for APIs, web pages, data | No config needed |

These give agents immediate file and web access without manual setup. Additional servers (brave-search, github, etc.) can be installed via Settings → MCP Tools.

> Full architecture details: [Architecture Overview](docs/architecture/OVERVIEW.md) | [Backend Spec](docs/architecture/BACKEND.md) | [Frontend Spec](docs/architecture/FRONTEND.md) | [Operations Manual](docs/architecture/OPERATIONS.md)

## Getting Started

> **Detailed guide:** See [Local Dev Workflow](docs/LOCAL_DEV_WORKFLOW.md) for configuration reference, port map, health checks, and troubleshooting.

### 1. Configure Secrets

```bash
cp .env.example .env
# Edit .env — set DB credentials, OLLAMA_HOST, NATS_URL, etc.
# REQUIRED: Set MYCELIS_API_KEY — server refuses to start without it.
# See docs/LOCAL_DEV_WORKFLOW.md for full variable reference.
```

### 2. Boot the Infrastructure

```bash
uvx inv k8s.reset    # Full System Reset (Cluster + Core + DB)
```

### 3. Quick Start (Recommended)

The lifecycle system handles port-forwards, dependencies, and server startup in one command:

```bash
uvx inv core.build           # First time only (compile Go binary)
uvx inv lifecycle.up          # Bring up: bridge → wait deps → core (background)
uvx inv lifecycle.up --build  # Or combine: build + bring up
```

Check everything:

```bash
uvx inv lifecycle.status      # Dashboard: Docker, Kind, PG, NATS, Core, Frontend, Ollama
uvx inv lifecycle.health      # Deep probe: hits actual API endpoints with auth
```

### 3b. Manual Start (Alternative)

If you prefer manual control over each service:

```bash
uvx inv k8s.bridge            # Port-forward PG:5432, NATS:4222 (Terminal 1)
uvx inv db.migrate            # Apply all migrations (idempotent)
uvx inv core.build && uvx inv core.run   # Build + run (Terminal 2, foreground)
uvx inv interface.install     # First time only
uvx inv interface.dev         # Next.js dev server (Terminal 3)
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
| **Lifecycle** | |
| `uvx inv lifecycle.status` | Dashboard: Docker, Kind, PG, NATS, Core, Frontend, Ollama (with PIDs) |
| `uvx inv lifecycle.up` | Idempotent bring-up: bridge → deps → core (background). `--frontend` `--build` |
| `uvx inv lifecycle.down` | Clean teardown: core → frontend → port-forwards |
| `uvx inv lifecycle.health` | Deep health probe: hits API endpoints with auth |
| `uvx inv lifecycle.restart` | Full restart: down → settle → up. `--build` `--frontend` |
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
| `/dashboard` | Mission Control — Resizable chat-dominant layout, OpsOverview dashboard, telemetry sparklines |
| `/wiring` | Neural Wiring — ArchitectChat + CircuitBoard (edit/delete agents) + NatsWaterfall |
| `/architect` | Redirects to `/wiring` |
| `/teams` | Team Management — browse standing + mission teams, agent roster, delivery targets |
| `/catalogue` | Agent Catalogue — CRUD for agent blueprints |
| `/memory` | Memory Explorer — Hot/Warm/Cold three-tier browser |
| `/approvals` | Governance — approval queue, policy config, team proposals (3 tabs) |
| `/missions/[id]/teams` | Team Actuation — live team drill-down |
| `/settings` | Profile, Teams, Cognitive Matrix, MCP Tools, Brains, Users |
| `/settings/brain` | Cognitive Matrix — provider routing grid |
| `/settings/brains` | Brains — provider management (enable/disable, policy, remote confirmation) |
| `/settings/users` | Users — stub auth, role display |
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
| `MYCELIS_API_KEY` | *(required)* | API authentication key. Server refuses to start without this. |
| `MYCELIS_WORKSPACE` | `./workspace` | Workspace sandbox root for agent file tools (`read_file`/`write_file`). |
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
uvx inv core.test             # Go unit tests (~120 handler tests)
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
| 18.0 | Command Center | Resizable Mission Control, OpsOverview dashboard, MCP bootstrap, rich chat (markdown + inline artifacts), tools_used surfacing |
| 18.5 | Soma Identity & Artifacts | Soma in-character identity (NEVER BREAK CHARACTER), council routing table, artifact pipeline (tool → CTS → inline render), broadcast with team replies, chat memory persistence (localStorage), tool_call JSON sanitizer, self-awareness block |
| 19.0 | Conversation Memory & Scheduled Teams | `summarize_conversation` tool, chat session persistence (localStorage 200-msg cap), TeamScheduler (configurable interval triggers via NATS), `DelegationHint` struct (confidence/urgency/complexity/risk) |
| 19.5 | Natural Human Interface | Label translation layer (`lib/labels.ts`), 20 tool labels, 5 council labels, governance/trust/workspace labels, trust badge `C:{score}`, recalled-memory annotation, all components migrated to human-facing names |
| P0 | Security Containment | API key auth middleware (fail-closed, constant-time compare), filesystem sandbox (`validateToolPath` + 1MB write cap), MCP raw install disabled (library-only), schedule safety (30s min interval + atomic guard), SSE CORS wildcard removed, Next.js middleware auth injection |
| 19.A | Agent & Provider Orchestration | BrainProvenance pipeline (agent → CTS → Zustand → per-message UI header), ModeRibbon (mode/brain/governance status bar), ProposedActionBlock (inline mutation proposals with confirm/cancel), OrchestrationInspector (audit panel), BrainsPage (provider management with remote-enable confirmation), UsersPage (stub auth) |
| 19.B | Proposal Loop & Lifecycle | Mutation detection in HandleChat/HandleCouncilChat (ModeProposal), confirmProposal wired to POST /api/v1/intent/confirm-action, brains toggle persistence to cognitive.yaml, MCP pool 15s timeout (prevents boot blocking), unified lifecycle management (`lifecycle.status/up/down/health/restart`) |

> Full phase history with details: [Architecture Overview](docs/architecture/OVERVIEW.md#vi-delivered-phases)

## Upcoming Architecture

Planned phases with detailed specifications are documented in the Architecture Overview:

| Phase | Name | Summary |
| :--- | :--- | :--- |
| 12 | Persistent Agent Memory | Cross-mission memory, semantic recall, memory consolidation daemon |
| 13 | Multi-Agent Collaboration | Intra-team debate protocol, consensus detection, SquadRoom live chat |
| 14 | Hot-Reload Runtime | Live agent goroutine replacement, zero-downtime reconfiguration |
| P1 | Structural Hardening | RBAC (role-based access control), Postgres RLS, memory isolation, JWT/session auth |
| P2 | Execution Hardening | Tool execution pipeline, SCIP validator, audit logging, prompt injection defense |
| P3 | Governance Completion | HTTP-layer governance hooks, rate limiting, advanced schedule safety |
| 16 | Distributed Federation | Multi-node NATS, team affinity, cross-instance delegation |
| 18 | Streaming LLM | Token-by-token streaming via SSE, mid-stream tool detection |
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
