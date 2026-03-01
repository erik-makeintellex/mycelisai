# Mycelis Cortex V7.0 — Architecture Overview

**Codename:** "The Mycelian Lattice"
**Classification:** Recursive Swarm Operating System / Multi-Agent Cybernetic OS
**Version:** 7.0 (February 2026)
**Objective:** Instantiate a self-aware, fractal operating system where all Agentry is strictly governed, universally configurable, and capable of extending the User's Will securely via localhost.

> **Related Docs:**
> - [Backend Specification](BACKEND.md) — Go packages, APIs, DB schema, NATS, execution pipelines
> - [Frontend Specification](FRONTEND.md) — Routes, components, Zustand, visual design
> - [Operations Manual](OPERATIONS.md) — Deployment, config, testing, CI/CD, invoke tasks
> - [Soma Team + Channel Architecture](SOMA_TEAM_CHANNEL_ARCHITECTURE_V7.md) — Inter-team/process channels, MCP execution I/O, and shared memory boundaries
> - [MCP Service Config (Local-First)](MCP_SERVICE_CONFIGURATION_LOCAL_FIRST_V7.md) — Standard onboarding/configuration for adding MCP services with local-default policy
> - [Universal Action Interface V7](UNIVERSAL_ACTION_INTERFACE_V7.md) — Unified action contracts across MCP/OpenAPI/Python with dynamic service onboarding APIs
> - [Actualization Beyond MCP V7](ACTUALIZATION_ARCHITECTURE_BEYOND_MCP_V7.md) — Multi-protocol architecture guidance inspired by OpenClaw/A2A/ACP patterns
> - [Secure Gateway + Remote Actuation](SECURE_GATEWAY_REMOTE_ACTUATION_PROFILE_V7.md) — Security baseline for self-hosted control planes and remote actuator services
> - [Hardware Interface API + Channels](HARDWARE_INTERFACE_API_AND_CHANNELS_V7.md) — Standard API + direct channel model for hardware interfaces and low-level IoT pathways
> - [Soma Symbiote + Host Actuation](SOMA_SYMBIOTE_GROWTH_AND_HOST_ACTUATION_V7.md) — Thought-profile backend contracts, learning growth loop, and localhost actuation model
> - [Soma Extension-of-Self PRD](../product/SOMA_EXTENSION_OF_SELF_PRD_V7.md) — Detailed action plan for local Ollama-first extension-of-self delivery with parallel lane execution
> - [README.md](../../README.md) — Operational source of truth (getting started, commands)

---

## I. Philosophy of the Organism

We do not build "software." We build **Synthetic Biology**. The system is a Lattice — a living mesh of independent nodes bound by a shared Nervous System.

### Core Tenets

1. **Universal Sovereignty:** The organism must survive in isolation. It defaults to "Localhost" air-gapped intelligence (`qwen2.5-coder:7b` via Ollama at `192.168.50.156:11434`). Cloud connectivity (OpenAI, Anthropic, Google Gemini) is a configurable privilege, never a dependency.

2. **Iron Dome Governance:** We do not trust the limb to act without the brain. Every actuation (Tool Call / File Write / External API) is intercepted, analyzed, and authorized by Tier 0 (The Human) via the Governance Valve. The Trust Economy (0.0–1.0 scoring) determines what auto-executes vs. what requires human approval.

3. **The Anti-Spaghetti Protocol:** Unidirectional data flow is absolute law. UI components do not fetch data — the Zustand store does. APIs do not panic — they degrade gracefully. State lives in isolation — one atomic store, zero distributed state.

4. **Context Propagation Mandate:** When Admin + Council research and design a blueprint, the deliberation context (research findings, council advice, design rationale) must be passed to spawned teams so they understand the problem — not just their role. Blueprint → `mission_context` field → injected into each agent's runtime context.

5. **Dependency Workflow Planning:** Before writing backend code that queries a table, the migration AND seed data for that table MUST exist. No orphan queries. No assumed schema.

---

## II. Technology Stack (Locked Versions)

| Component | Version | Module/Package |
|-----------|---------|----------------|
| **Go** | 1.26 | `github.com/mycelis/core` |
| **PostgreSQL** | 16 | pgvector/pgvector:16-alpine, pgvector extension (768-dim) |
| **NATS** | 2.12.4 (JetStream) | `github.com/nats-io/nats.go` v1.48.0 |
| **Next.js** | 16.1.6 | Turbopack, App Router |
| **React** | 19.2.3 | `"use client"` required for hooks/state |
| **Tailwind CSS** | v4 | `@import "tailwindcss"` syntax, `@theme` directive |
| **ReactFlow** | 11.11.4 | Package is `reactflow`, NOT `@xyflow/react` |
| **Zustand** | 5.0.11 | Atomic store: `useCortexStore` |
| **Python** | >=3.12 | uv managed, invoke for task orchestration |
| **Protocol Buffers** | protobuf v1.36.1 | SCIP protocol, MsgEnvelope |

---

## III. The 4-Layer Anatomy

### Layer 0: The Sovereign Base (Infrastructure & Memory)

**Purpose:** The immutable foundation — compute, storage, and identity.

- **Go 1.26** backend compiles to a single static binary
- **Kubernetes (Kind 1.33)** for container orchestration, production-ready Helm charts
- **Cognitive Registry:** Dual-source config (YAML base + DB overlay + env overrides) maps system roles (admin, architect, coder, creative, sentry, chat) to LLM provider endpoints
- **Hippocampus (Hybrid Memory):**
  - *Entity Store (Postgres):* Relational facts — users, missions, teams, manifests, artifacts (14+ tables, 21 migrations)
  - *Episodic Store (pgvector):* Semantic wisdom — SitReps auto-embedded via nomic-embed-text (768-dim), cosine distance search
  - *Working Memory:* Short-lived team context with TTL expiry
  - *Agent State:* Key-value per agent, mission-scoped, optional TTL

### Layer 1: The Nervous System (Communication)

**Purpose:** The pub/sub message bus connecting all living components.

- **NATS JetStream 2.12** — 30+ topics, 128Mi memory store, 2Gi file store
- **Global Control Plane:** heartbeat (5s), announce, broadcast, user input, audit trace
- **Team Internal:** trigger, response, command, status, telemetry (per team_id)
- **Council:** Request-reply RPC (`swarm.council.{agent_id}.request`, 30s timeout)
- **Sensor Ingress:** Hierarchical topics (`swarm.data.email.>`, `swarm.data.weather.>`)
- All topics use constants from `pkg/protocol/topics.go` — never hardcode

### Layer 2: The Fractal Fabric (Orchestration)

**Purpose:** The brain. Translates intent into executable plans, orchestrates agents, enforces governance.

| Component | File | Role |
|-----------|------|------|
| **Soma** | `swarm/soma.go` | Executive cell — user proxy, team manager, spawns/manages all teams |
| **Axon** | `swarm/axon.go` | Signal router — monitors team traffic, bridges to SSE stream |
| **Meta-Architect** | `cognitive/architect.go` | Intent → MissionBlueprint via LLM structured output |
| **Overseer** | `overseer/engine.go` | Zero-trust DAG orchestrator — validates proof before advancing |
| **Archivist** | `memory/archivist.go` | Context engine — buffers events, LLM-compresses to SitReps, auto-embeds to pgvector |
| **Agent** | `swarm/agent.go` | LLM reasoning node — ReAct loop with tool invocation (max 5 iterations) |
| **SensorAgent** | `swarm/sensor_agent.go` | Poll-based data acquisition — HTTP → CTS envelope at TrustScore 1.0 |
| **Tool Registry** | `swarm/internal_tools.go` | 17 built-in tools (consult_council, delegate_task, search_memory, file I/O, etc.) |
| **Tool Executor** | `swarm/tool_executor.go` | Composite dispatcher — internal registry first, MCP pool fallback |
| **Governance Guard** | `governance/guard.go` | Policy engine — rule evaluation, approval queue, YAML-based rules |
| **MCP Ingress** | `mcp/service.go` | Model Context Protocol — install, discover, invoke external tool servers |

**Standing Teams (auto-spawned from YAML):**
| Team ID | Name | Members | Purpose |
|---------|------|---------|---------|
| `admin-core` | Admin | 1 admin agent (17 tools, 5 ReAct iterations) | User proxy, orchestration |
| `council-core` | Council | architect, coder, creative, sentry | Specialist cognitive council |
| `genesis-core` | Genesis | architect, commander | Bootstrap operations |

### Layer 3: The Conscious Face (Interface)

**Purpose:** The visual cortex — real-time window into organism thoughts, actions, and governance.

- **4-Zone Shell:** Rail (nav) + Workspace (canvas) + Spectrum (telemetry) + Governance (overlay)
- **15 routes** across marketing and app route groups
- **93 React components** across 20 folders
- **Single Zustand store** with ~60 state fields and ~40 actions
- **Midnight Cortex theme:** `cortex-bg #09090b`, `cortex-primary #06b6d4` (cyan)
- **Agent Node Categories:** Cognitive (cyan), Sensory (blue), Actuation (green), Ledger (muted)

---

## IV. Key Execution Pipelines (Summary)

| Pipeline | Flow | Details |
|----------|------|---------|
| **Intent → Activation** | User types intent → Meta-Architect generates blueprint → Ghost-draft DAG → User commits → DB persist → Soma spawns teams/agents | [BACKEND.md §IV.1](BACKEND.md#pipeline-1-intent--blueprint--activation) |
| **Council Chat** | User selects member → POST /council/{member}/chat → NATS request-reply → Agent ReAct loop → CTS-enveloped response | [BACKEND.md §IV.2](BACKEND.md#pipeline-2-council-chat-request-reply) |
| **Agent ReAct** | Trigger → Build context → LLM infer → Parse tool_call/final → Execute tool → Loop or publish | [BACKEND.md §IV.3](BACKEND.md#pipeline-3-agent-react-loop) |
| **Memory Archival** | Events buffer → Threshold hit → LLM compress → pgvector embed → SitRep stored | [BACKEND.md §IV.4](BACKEND.md#pipeline-4-memory-archival--compression) |
| **Governance** | CTS envelope → TrustScore check → Auto-execute (≥0.7) or halt → Human approval queue | [BACKEND.md §IV.5](BACKEND.md#pipeline-5-governance--zero-trust-actuation) |
| **SSE Streaming** | Axon routes signals → StreamHandler → EventSource → Zustand streamLogs[] → NatsWaterfall + node dispatch | [BACKEND.md §IV.6](BACKEND.md#pipeline-6-sse-real-time-streaming) |

---

## V. Data Contracts (Summary)

| Contract | Purpose | Details |
|----------|---------|---------|
| **CTSEnvelope** | Universal message wrapper — ID, SourceNode, TeamID, SignalType, TrustScore, Payload, CTSMeta | [BACKEND.md §V.1](BACKEND.md#1-the-cts-envelope-cortex-telemetry-standard) |
| **MissionBlueprint** | Decomposed intent — teams, agents, constraints, resource requirements | [BACKEND.md §V.4](BACKEND.md#4-the-missionblueprint) |
| **AgentManifest** | Agent identity — role, system_prompt, model, tools, inputs, outputs, verification | [BACKEND.md §V.5](BACKEND.md#5-the-agentmanifest) |
| **SitRep Schema** | Archivist output — 3-sentence summary, key_events, strategies_applied | [BACKEND.md §V.7](BACKEND.md#7-the-sitrep-schema-archivist-output) |
| **APIResponse** | HTTP envelope — `{ ok, data, error }` | [BACKEND.md §V.3](BACKEND.md#3-the-apiresponse-envelope) |
| **ChatResponsePayload** | Council response — response text, consultations, tools_used, trust_score | [BACKEND.md §V.2](BACKEND.md#2-the-chatresponsepayload) |

---

## VI. Delivered Phases

| Phase | Name | Key Deliverables |
|-------|------|-----------------|
| 1–3 | Genesis Build | Core server, NATS, Postgres, ReactFlow, basic UI |
| 4.1–4.6 | Foundation | Zustand store, intent commit, SSE binding, Overseer DAG |
| 4.4 | Governance | Deliverables Tray, Governance Modal (Human-in-the-Loop) |
| 5.0 | Archivist Daemon | NATS buffer → LLM compress → sitreps table, periodic 5min flush |
| 5.1 | SquadRoom | Fractal navigation (double-click team → drill-down), Mission Control layout |
| 5.2 | Sovereign UX & Trust | Trust Economy (CTS TrustScore, Governance Valve, SSE governance_halt), Telemetry API, Trust API, AgentNode iconography (4 categories), TrustSlider, BlueprintDrawer, Panopticon layout |
| 5.3 | RAG & Sensory UI | EmbedProvider interface, pgvector auto-embed, SemanticSearch, SensorLibrary (grouped subscriptions), ManifestationPanel, NatsWaterfall signal classification (input/output/internal), MissionControl 3-column responsive grid |
| 6.0 | Host Internalization | Activation Bridge (commitAndActivate), SensorAgent (poll-based), Blueprint Converter (mission-scoped IDs), Symbiotic Seed (Gmail+Weather, no LLM), Team.sensorConfigs, Migration 010 |
| 7.7 | Admin & Council | HandleChat NATS-only (no raw LLM fallback), provider naming (vllm/ollama/lmstudio), dynamic cognitive status, Settings 4-tab page, Cognitive Matrix UI + ProviderConfigModal, 17 internal tools, Council YAML, ToolsPalette drawer |
| 8.0 | Visualization | Observable Plot (bar/line/area/dot/waffle/tree), Leaflet geo maps, DataTable, ChartRenderer, MycelisChartSpec type system, ArtifactViewer inline chart rendering, SquadRoom ProofCard |
| 9.0 | Neural Wiring CRUD | WiringAgentEditor drawer, CircuitBoard edit/delete, mission CRUD API (GET/PUT/DELETE), Migration 020 (FK cascade), Zustand wiring actions (7 new), draft mode vs active mode |
| 10.0 | Meta-Agent Research | research_for_blueprint tool, admin-routed negotiate via NATS |
| 11.0 | Team Management | /teams route, TeamCard, TeamDetailDrawer, GET /api/v1/teams/detail |

---

## VII. Upcoming Architecture

### V7.x Immediate Program: Soma Extension-of-Self

**Objective:** Operationalize Soma as a governed extension-of-self with explicit decision contracts, local-first cognition visibility, and universal action channels.

- **Decision Runtime:** Add typed decision frames (`direct_action`, `manifest_team`, `propose_only`, `scheduled_repeat`) and run-linked audit emission.
- **Local Ollama Contract:** Treat local provider readiness (reachability, model availability, latency, throughput, failure ratio) as a first-class operational dependency.
- **Universal Action Convergence:** Keep MCP as baseline adapter while onboarding one non-MCP adapter through the same invoke contract and governance gates.
- **Team Lifetime and Repeat:** Enforce `ephemeral|persistent|auto` with scheduler-backed repeat promotion.
- **Governed Host/Hardware Path:** Add localhost-first host/hardware action scaffolds under allowlist + approval boundaries.

Execution authority:
- `mycelis-architecture-v7.md` (Part X)
- `docs/product/SOMA_EXTENSION_OF_SELF_PRD_V7.md`

### Phase 12: Persistent Agent Memory & Long-Term Learning

**Objective:** Give agents durable, cross-mission memory so they accumulate expertise over time.

- **Agent Memory Store:** Extend `agent_state` table with mission-agnostic keys (no FK to missions). Agents can `remember` facts that persist across mission boundaries.
- **Semantic Agent Memory:** Auto-embed agent observations into pgvector with agent-scoped namespace. Agents can `recall` by semantic similarity, not just exact key.
- **Memory Consolidation Daemon:** Periodic background process that merges short-term agent_state entries into long-term sitreps using LLM compression (same Archivist pattern).
- **Cross-Mission Context Injection:** When an agent spawns, `BuildContext()` queries its past memory store and injects relevant prior observations into the system prompt.
- **Memory Decay / Pruning:** TTL-based expiry for low-value memories. High-trust observations persist longer. Governance rules can protect critical memories from pruning.
- **Frontend:** Memory Explorer (`/memory`) enhanced with per-agent memory timeline, search-by-agent, and memory health indicators.

### Phase 13: Real-Time Agent Collaboration (Multi-Agent Chat)

**Objective:** Enable agents within a team to engage in structured deliberation before producing output.

- **Intra-Team Chat Protocol:** New NATS topic pattern `swarm.team.{id}.chat.{round}` for structured multi-round debate between agents in a team.
- **Debate Orchestrator:** Team-level coordinator that poses the task to all agents, collects proposals, runs N rounds of critique/refinement, and synthesizes final output via voting or lead-agent selection.
- **SquadRoom Live View:** Real-time rendering of agent debate in the SquadRoom component. Each agent's messages appear in their column with trust scores and tool usage visible.
- **Consensus Detection:** LLM-based analysis of agent proposals to detect convergence. When agents agree, stop debating early.
- **Role-Based Debate Rules:** Architect proposes structure, Coder implements, Sentry reviews for vulnerabilities, Creative handles UX. Each role has a defined phase in the debate cycle.
- **Frontend:** SquadRoom enhanced with live chat bubbles per agent, debate progress bar, and consensus indicator.

### Phase 14: Hot-Reload Agent Runtime

**Objective:** Allow live modification of running agents without mission restart.

- **Agent Goroutine Replacement:** When `PUT /missions/{id}/agents/{name}` is called:
  1. Graceful stop signal to running agent goroutine
  2. Wait for current ReAct iteration to complete (drain)
  3. Spawn new goroutine with updated manifest
  4. Re-subscribe to NATS topics
  5. Agent resumes with updated system_prompt, model, tools
- **Zero-Downtime Team Reconfiguration:** Add/remove agents from a running team without stopping other agents.
- **Live Model Switching:** Change an agent's LLM provider mid-mission (e.g., switch from ollama to production_gpt4 for a critical decision).
- **Frontend:** WiringAgentEditor shows "Applying changes..." spinner during hot-reload. Node status briefly shows "restarting" before returning to "online".

### Phase 15: Advanced Governance & RBAC

**Objective:** Production-ready access control and audit compliance.

- **RBAC Enforcement:** Activate the `User.role` field (admin|user|observer) with middleware. Admin = full access; User = create missions, chat, view telemetry; Observer = read-only.
- **API Key Authentication:** Per-user API keys for programmatic access. Keys stored hashed in DB, scoped to role.
- **Governance Audit Trail:** Persist all approval decisions to `governance_audit` table (who, what, when, context). Queryable via API.
- **Policy Versioning:** Track policy changes over time. Rollback capability. Diff view in UI.
- **Approval Delegation:** Allow admins to delegate approval authority to specific council members for defined categories.
- **Frontend:** Settings → Security tab for RBAC management. Audit log viewer in `/approvals`.

### Phase 16: Distributed Deployment & Federation

**Objective:** Scale beyond a single node. Multiple Mycelis instances federate into a mesh.

- **Multi-Node NATS Cluster:** Clustered JetStream for HA.
- **Team Affinity:** Teams pinned to specific nodes based on hardware requirements (GPU, memory).
- **Cross-Instance Federation:** Instances discover each other via NATS and delegate tasks across the mesh. A mission on Instance A can spawn a team on Instance B if B has the required resources.
- **State Replication:** Mission state replicated across instances for fault tolerance.
- **Frontend:** Network topology view showing connected instances, team placement, cross-instance traffic.

### Phase 17: Vuexy Legacy Migration Completion

**Objective:** Eliminate all legacy Vuexy RGB CSS variables and complete cortex-* migration.

- **Pages to Migrate:** `/telemetry`, `/marketplace`, `/approvals` — three remaining routes using legacy zinc/slate classes.
- **Remove Legacy Variables:** Delete `--background`, `--surface`, `--border`, `--border-active` from globals.css.
- **Automated Audit:** Scan all components for `bg-white`, `bg-zinc-*`, `bg-slate-*`, or Vuexy color references.
- **Smoke Test Coverage:** Ensure `uvx inv interface.check` validates all 15 routes.

### Phase 18: Streaming LLM Responses

**Objective:** Replace request-response LLM calls with streaming for real-time token output.

- **Streaming Inference:** Modify `cognitive.Router.Infer()` to support streaming mode across all provider adapters.
- **SSE Token Relay:** Relay tokens from LLM to frontend via SSE in real-time.
- **ArchitectChat Streaming:** Tokens appear as Meta-Architect generates blueprints.
- **Council Chat Streaming:** Council responses stream token-by-token in MissionControlChat.
- **Tool Call Detection:** Detect `tool_call` JSON fragments mid-stream and trigger tool execution without waiting for full response.
- **Frontend:** Typing indicator and progressive markdown rendering in chat components.

### Phase 19: Workflow Templates & Mission Library

**Objective:** Pre-built mission blueprints with one-click instantiation.

- **Mission Template Store:** Versioned templates with categories, descriptions, required resources.
- **Template Editor:** Visual blueprint editor in `/wiring` for creating reusable templates.
- **One-Click Instantiation:** Select template → auto-fill parameters → commit → activate.
- **Built-In Templates:** Code Review Pipeline, Research Assistant, Data Analysis Pipeline, Content Generation Workflow.
- **Frontend:** Template browser in BlueprintDrawer with preview and parameter customization.

### Phase 20: Observability & Metrics Dashboard

**Objective:** Production-grade monitoring with historical metrics and alerting.

- **Historical Metrics:** Time-series data for goroutines, memory, token rate, active agents, mission throughput.
- **Prometheus Export:** `/metrics` endpoint in Prometheus exposition format.
- **Agent Performance:** Per-agent metrics — response time, tool call frequency, trust score distribution, ReAct loop depth.
- **Mission Analytics:** Duration, team utilization, agent contribution heatmaps.
- **Alerting:** Configurable alerts for agent offline, trust score anomalies, memory pressure.
- **Frontend:** `/telemetry` rebuilt with Observable Plot time-series charts, real-time gauges, alert configuration.

---

## VIII. Statistics

| Metric | Count |
|--------|-------|
| Go source files | 110+ |
| Go packages | 15 internal + 3 pkg |
| HTTP endpoints | 50+ |
| NATS topics | 30+ |
| Database tables | 14+ |
| SQL migrations | 21 |
| Built-in tools | 17 |
| LLM adapters | 4 |
| Frontend routes | 15 |
| React components | 93 |
| Zustand state fields | ~60 |
| Zustand actions | ~40 |
| Invoke task modules | 11 |
| @task functions | 60+ |
| Go test suites | 11 |
| Vitest tests | ~114 |
| Playwright specs | 12 |
| CI workflows | 3 |

