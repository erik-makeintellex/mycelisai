# Mycelis Cortex V7.0

**The Recursive Swarm Operating System.**

> [!IMPORTANT]
> **MASTER STATE AUTHORITY**
> This README is the **Single Source of Truth** for project state, architecture, and operational commands.
>
> **Architecture PRD:** The detailed architecture specification lives in core documents:
> | Document | Load When |
> | :--- | :--- |
> | [Architecture Overview](docs/architecture/OVERVIEW.md) | Planning phases, architectural decisions |
> | [Backend Specification](docs/architecture/BACKEND.md) | Working on Go code, APIs, DB, NATS |
> | [Frontend Specification](docs/architecture/FRONTEND.md) | Working on React/Next.js, Zustand, design |
> | [Operations Manual](docs/architecture/OPERATIONS.md) | Deploying, testing, CI/CD, config |
> | [V7 PRD](mycelis-architecture-v7.md) | Event spine, triggers, scheduler, workflow-first IA |
> | [V7 UI Framework](docs/UI_FRAMEWORK_V7.md) | Canonical UI element/state/testing framework |
> | [V7 UI Parallel Delivery Board](docs/ui-delivery/PARALLEL_DELIVERY_BOARD.md) | Active lane plan, merge gates, evidence checklist |

Mycelis is a governed orchestration system ("Neural Organism") where users express intent, Mycelis proposes structured plans, and any state mutation requires explicit confirmation plus a complete Intent Proof bundle. Missions are not isolated — they emit structured events that trigger other missions. Observability is not optional: execution must never be a black box.

Built through 19 phases — from genesis through **Admin Orchestrator**, **Council Activation**, **Trust Economy**, **RAG Persistence**, **Agent Visualization**, **Neural Wiring Edit/Delete**, **Meta-Agent Research**, **Team Management**, **Soma Identity & Artifacts**, **Conversation Memory**, **Natural Human Interface**, **Phase 0 Security Containment**, **Agent & Provider Orchestration** — and now executing **V7: Event Spine & Workflow-First Orchestration**. V7 Team A (Event Spine), Team B (Trigger Engine), Agent Conversation Log, and Inception Recipes are complete: persistent mission runs, `MissionEventEnvelope` audit records, tool event emission, run timeline APIs, declarative trigger rules with cooldown/recursion/concurrency guards, full-fidelity agent conversation transcripts with operator interjection, and structured inception recipe patterns for RAG-based knowledge reuse. Team C (Scheduler) and Causal Chain UI follow next.

## Architecture

### Tier 1: Core (Go 1.26 + Postgres + pgvector)

- **Soma → Axon → Teams → Agents:** Mission activation pipeline with heartbeat + proof-of-work.
- **Parallel Team Activation:** `ActivateBlueprint` starts eligible teams concurrently (bounded worker pool), then inserts them race-safely into Soma's team map. Duplicate IDs are skipped idempotently even under concurrent activation calls.
- **Standing Teams:** Admin/Soma (orchestrator, 18 tools, 10 ReAct iterations, persistent identity) + Council (architect, coder, creative, sentry) — all individually addressable via `POST /api/v1/council/{member}/chat`.
- **Council Chat API:** Standardized CTS-enveloped responses with trust scores, provenance metadata, and tools-used tracking. Dynamic member validation via Soma — add a YAML, restart, done.
- **Runtime Context Injection:** Every agent receives live system state (active teams, NATS topology, MCP servers, cognitive config, interaction protocols) via `InternalToolRegistry.BuildContext()`.
- **Internal Tool Registry:** 22 built-in tools — consult_council, delegate_task, search_memory, remember, recall, broadcast, file I/O (workspace-sandboxed), NATS bus sensing, image generation, summarize_conversation, research_for_blueprint, store_inception_recipe, recall_inception_recipes, and more.
- **Composite Tool Executor:** Unified interface routing tool calls to InternalToolRegistry or MCP ToolExecutorAdapter.
- **MCP Ingress:** Install, manage, and invoke MCP tool servers. Curated library with one-click install. Raw install endpoint disabled (Phase 0 security) — library-only installs enforced.
- **MCP Test Coverage:** Service/toolset/library/executor suites plus DB-backed MCP handler tests; toolset update not-found now returns HTTP 404.
- **Archivist:** Context engine — SitReps, auto-embed to pgvector (768-dim, nomic-embed-text), semantic search.
- **Governance:** Policy engine with YAML rules, approval queue, trust economy (0.0–1.0 threshold).
- **Cognitive Router:** 6 LLM providers (ollama, vllm, lmstudio, OpenAI, Anthropic, Gemini), profile-based routing, token telemetry. Brain provenance tracks which provider/model executed each response.
- **CE-1 Templates:** Orchestration template engine with intent proofs, confirm tokens (15min TTL), and audit trail. Chat-to-Answer (read-only) and Chat-to-Proposal (mutation-gated) execution modes.
- **Brains API:** Full provider CRUD — add, edit, delete, and probe providers at runtime with zero restart (`AddProvider`/`UpdateProvider`/`RemoveProvider` with `RWMutex` hot-reload). Type presets for Ollama, vLLM, LM Studio, OpenAI, Anthropic, Gemini, Custom. Location/data_boundary/usage_policy/roles_allowed enforced. All mutations persist to `cognitive.yaml`.
- **Mission Profiles:** Named workflow configurations that map agent roles to specific providers. Activate a profile to instantly reroute `architect → vllm`, `coder → ollama`, etc. Context Switch strategies: Cache & Transfer (auto-snapshot before switch), Start Fresh, Load Snapshot. Profiles persist to `mission_profiles` table (migration 029).
- **Context Snapshots:** Point-in-time saves of conversation state (messages, run state, active role providers). Created automatically on Cache & Transfer activation or manually. Stored in `context_snapshots` table (migration 028). Last 20 shown in Load Snapshot picker.
- **Reactive Subscription Engine:** Mission profiles can subscribe to NATS topic patterns (e.g. `swarm.team.research-team.*`). On message receipt, Soma evaluates whether to engage — reactive-watch, not forced auto-chain. Subscriptions survive NATS reconnects via `ReactivateFromDB`. Engine in `core/internal/reactive/engine.go`.
- **Self-Healing Connectivity:** NATS client uses `MaxReconnects(-1)` (unlimited) with 20s ping interval for stale connection detection. Core startup retries DB and NATS for up to 90s (45×2s) before failing. Reactive engine re-subscribes all active profiles on NATS reconnect.
- **Event Spine (V7):** Dual-layer event architecture — CTS for real-time signal transport, MissionEventEnvelope for persistent audit-grade records. Every execution creates a `mission_run` with unique `run_id`. Events persisted to `mission_events` before CTS publish. CTS payloads reference `mission_event_id` for timeline reconstruction.
- **Trigger Rules Engine (V7):** Declarative IF/THEN trigger rules evaluated on event ingest. Supports cooldown, recursion guard (max depth), and concurrency guard. Default mode: propose-only. Auto-execute requires explicit policy allowance.
- **Conversation Log (V7):** Full-fidelity agent conversation persistence — every system prompt, user message, tool call, tool result, and assistant response stored in `conversation_turns` table (migration 030). Separate from lightweight `mission_events` — turns are full-text blobs (10KB+). Session-scoped with `session_id`; run-linked when available. `ConversationLogger` interface propagated Soma → Team → Agent, matching `EventEmitter` pattern.
- **Operator Interjection (V7):** Mid-run user redirection via NATS mailbox (`swarm.agent.{id}.interjection`). Agents check a mutex-protected buffer between ReAct iterations. Interjections are injected as `[OPERATOR INTERJECTION]` messages and logged as `role=interjection` turns. Only applies to in-progress runs with active agents.
- **Inception Recipes (V7):** Structured prompt pattern library for knowledge reuse. When Soma completes a complex task, it can distill an inception recipe capturing: intent pattern, key parameters, example prompt, and outcome shape. Recipes are dual-persisted — RDBMS (`inception_recipes` table, migration 031) for structured queries + pgvector (`context_vectors`) for semantic recall. Quality feedback loop (`quality_score` 0.0–1.0) + usage tracking. Automatically recalled during `research_for_blueprint` pipeline.
- **Scheduler (V7):** In-process goroutine scheduler backed by `scheduled_missions` table. Enforces max_active_runs, suspends when NATS offline. Cron expressions for recurring missions.

### Tier 2: Nervous System (NATS JetStream 2.12)

- 30+ topics: heartbeat, audit trace, team internals, council request-reply, sensor ingress, mission DAG.
- All topics use constants from `pkg/protocol/topics.go` — never hardcode.

### Tier 4: Event & Scheduling Layer (V7)

- **mission_runs (023):** Execution identity — `run_id`, `mission_id`, `intent_proof_id`, `triggered_by_rule_id`, `parent_run_id`, `depth`, `status`. Anchors all timelines and chains.
- **mission_events (024):** MissionEventEnvelope persistence — 17-field audit-grade event records with causal linking (`parent_event_id`, `parent_run_id`, `trigger_rule_id`).
- **trigger_rules (025):** Declarative event → action rules with cooldown, recursion guard, concurrency guard. Default mode: propose.
- **trigger_executions (026):** Audit log of rule evaluations — evaluated/fired/skipped with reason.
- **scheduled_missions (027):** Cron-backed recurring execution with `max_active_runs` guard.
- **conversation_turns (030):** Full-fidelity agent conversation log — session-scoped, role-typed (system/user/assistant/tool_call/tool_result/interjection), with provider/model provenance and tool call threading via `parent_turn_id`.
- **inception_recipes (031):** Structured prompt patterns for RAG recall — category-indexed, trigram-searchable titles, dual-persisted (RDBMS + pgvector), quality/usage tracking.

### Tier 3: The Face (Next.js 16 + React 19 + Zustand 5)

- **Workspace (`/dashboard`):** The admin's primary command interface — renamed from "Mission Control". See [Workspace Reference](#workspace-reference) below.
- **V7 Navigation (Workflow-First):** 5 primary panels — Workspace (chat, proposals, timelines), Automations (scheduled missions, triggers, drafts, approvals, teams, Neural Wiring), Resources (brains, tools, catalogue), Memory (semantic search, recall events), System (health, NATS, DB, services dashboard — Advanced Mode toggle). Architecture-surface routes (`/wiring`, `/architect`, `/teams`, `/catalogue`, `/approvals`, `/telemetry`, `/matrix`) redirect server-side to their workflow parent with `?tab=` deep-linking.
- **Neural Wiring (Automations → Wiring tab, Advanced Mode):** ArchitectChat + CircuitBoard (ReactFlow) + ToolsPalette + NatsWaterfall. Interactive edit/delete: click agent nodes to modify manifests, delete agents, discard drafts, or terminate active missions.
- **Agent Visualization:** Observable Plot charts (bar, line, area, dot, waffle, tree), Leaflet geo maps, DataTable — rendered inline via ChartRenderer from `MycelisChartSpec`.
- **Memory Explorer (`/memory`):** Two-column redesign — Warm (sitreps + artifacts, 40%) + Cold semantic search (60%). Hot signal stream hidden behind Advanced Mode toggle (collapsible). Human-facing labels throughout.
- **Settings (`/settings`):** Brains (full provider CRUD — Add/Edit/Delete/Probe with type presets, remote-enable confirmation, LOCAL/LEAVES_ORG boundary badge), Profiles (mission profile CRUD + activate + Context Switch Modal), Cognitive Matrix, MCP Tools (curated library), Users (stub auth). Policy/approval rules in Automations → Approvals tab.
- **Run Timeline (V7):** Vertical event timeline per mission run — policy decisions, tool invocations, trigger firings, artifacts, completion. `RunTimeline.tsx` + `EventCard.tsx` + `/runs/[id]` page. Auto-polls every 5s; stops on terminal events. Tab bar switches between Conversation and Events views.
- **Conversation Log (V7):** Full agent transcript viewer per run — `ConversationLog.tsx` + `TurnCard.tsx`. Role-based coloring (system gray, user cyan, assistant green, tool_call violet, tool_result amber, interjection red). Agent filter bar for multi-agent runs. System prompts collapsed by default. Provider/model badges on assistant turns. Tool name badges on tool_call turns. Auto-polls 5s while run is active. Operator interjection input bar visible when run status is `running`.
- **Run List (V7):** `/runs` page listing all recent runs across missions, with status dots and timestamps. Also surfaced in OpsOverview as a `Recent Runs` widget.
- **Causal Chain View (V7, backend ready):** Parent run → event → trigger → child run traversal. `GET /api/v1/runs/{id}/chain` handler complete; UI pending.
- **Mode Ribbon:** Always-visible status bar showing current execution mode, active brain (with local/remote badge), and governance state.
- **Operational Reliability UX (V7 Gate A):** Global `DegradedModeBanner`, global `StatusDrawer` (opened from Mode Ribbon or floating status action), structured `CouncilCallErrorCard` with retry/reroute/copy diagnostics actions, Workspace Focus Mode (`F` key), and `SystemQuickChecks` panel on `/system`.
- **Proposal Blocks:** Inline chat cards for mutation-gated actions — shows intent, tools, risk level, confirm/cancel buttons wired to CE-1 confirm token flow.
- **Orchestration Inspector:** Expandable audit panel showing template ID, intent proof, confirm token, and execution mode for each chat response.
- **Visual Protocol:** Midnight Cortex theme — `cortex-bg #09090b`, `cortex-primary #06b6d4` (cyan). Zero `bg-white` in new code. Base font-size 17px for rem-proportional readability across all Tailwind utility classes.

### Workspace Reference

Workspace (`/dashboard`) is the admin's primary interface — a resizable two-panel layout where 80% of work happens in conversation.

```
┌───────────────────────────────────────────────────────────────┐
│  Workspace                SIGNAL: LIVE   [Launch Crew] [⚙]   │
├─── Telemetry Row ─────────────────────────────────────────────┤
│  [Goroutines: 42] [Heap: 18MB] [System: 52MB] [LLM: 3.2t/s] │
├───────────────────────────────────────────────────────────────┤
│                                                               │
│   ADMIN / COUNCIL CHAT  (55% — resizable)                     │
│   ┌───────────────────────────────────────────────────────┐   │
│   │  ● Soma  [⚡ Direct ▾]  (Soma is always primary)      │   │
│   │  Rich messages: markdown, code blocks, tables, links  │   │
│   │  Inline artifacts: charts, images, audio, data        │   │
│   │  DelegationTrace: council members Soma consulted      │   │
│   │  Trust score badges + tool-use pills                  │   │
│   │  /all prefix or broadcast toggle for swarm-wide msgs  │   │
│   │  Soma Offline Guide when no council members reachable │   │
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
│   │ MISSIONS (full width)  — registerOpsWidget order:50      ││
│   │ mission-abc ●LIVE  2T/6A    mission-xyz ●DONE  1T/3A    ││
│   └──────────────────────────────────────────────────────────┘│
│   ┌──────────────────────────────────────────────────────────┐│
│   │ RUNS (full width)  — registerOpsWidget order:60          ││
│   │ ● abc1234  running   12s ago  ⚡                         ││
│   │ ● def5678  completed  5m ago  ⚡                         ││
│   └──────────────────────────────────────────────────────────┘│
└───────────────────────────────────────────────────────────────┘
```

#### Layout

| Zone | Component | Description |
| :--- | :--- | :--- |
| **Header** | `MissionControl` | Signal status (SSE live/offline), **Launch Crew** button (opens `LaunchCrewModal` — guided 3-step crew intent → proposal → confirm), Settings gear (→ `/settings`) |
| **Telemetry Row** | `TelemetryRow` | 4 sparkline cards — Goroutines, Heap, System Memory, LLM Tokens/s. Polls `/api/v1/telemetry/compute` every 5s. Shows offline banner after 3 failures. |
| **Chat (top)** | `MissionControlChat` | The primary interaction surface — see Chat section below. Shows `SomaOfflineGuide` with startup instructions + retry button when no council members are reachable. |
| **Resize Handle** | custom drag | Pointer-event drag handler. Split ratio persisted to localStorage (`workspace-split`). Clamped 25%–80%. |
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
| **Soma-first header** | Soma is always the primary target — locked in the header. A `⚡ Direct` popover allows advanced users to target a specific council member directly (shown in amber when active). Resets to Soma on page load and on `LaunchCrewModal` open. |
| **Delegation Trace** | When Soma calls `consult_council` during its ReAct loop, a `DelegationTrace` card appears below the response showing which council members were consulted and a 300-char summary of their contribution. Color-coded per member (Architect=info, Coder=success, Creative=warning, Sentry=danger). |
| **Live activity** | While Soma processes, a `SomaActivityIndicator` reads `streamLogs` for `tool.invoked` events and shows contextual text: "Consulting Coder...", "Generating blueprint...", "Searching memory..." instead of a static spinner. |
| **Broadcast mode** | Toggle or `/all` prefix — sends message to ALL active teams via NATS |
| **File I/O** | Admin + council agents can `read_file` and `write_file` within the workspace sandbox (`MYCELIS_WORKSPACE`, default `./workspace`). Paths must resolve inside the boundary — symlink escapes are detected. Max 1MB per write. Sentry is read-only. |
| **Tool access** | 20 internal tools: consult_council, delegate_task, search_memory, remember, recall, broadcast, publish_signal, read_signals, read_file, write_file, generate_image, research_for_blueprint, generate_blueprint, list_teams, list_missions, get_system_status, list_available_tools, list_catalogue, store_artifact, summarize_conversation |
| **MCP tools** | Any installed MCP server tools are also available (filesystem, fetch, brave-search, etc.) |
| **Trust scores** | Each response carries a CTS trust score (0.0–1.0), displayed as a colored badge |
| **Multi-turn** | Full conversation history is forwarded to the agent — maintains context across turns |
| **Chat memory** | Conversation persists to localStorage key `mycelis-workspace-chat` (survives page refresh, 200-message cap). Use `clearMissionChat` to reset. Migrates transparently from legacy `mycelis-mission-chat` key. |
| **Broadcast replies** | Broadcast collects responses from all active teams via NATS request-reply (60s timeout per team) |
| **Artifact pipeline** | Tool results with artifacts (images, charts, code) flow through CTS envelope to frontend for inline rendering |

#### Ops Overview Cards

| Card | Layout | Data Source | Deep Link | Refresh |
| :--- | :--- | :--- | :--- | :--- |
| **System Status** | grid (order 10) | `GET /api/v1/cognitive/status` + `GET /api/v1/sensors` | `/settings/brain` | 15s / 60s |
| **Priority Alerts** | grid (order 20) | SSE signal stream (governance_halt, error, task_complete, artifact) | Signal Detail Drawer (click row) | Real-time |
| **Standing Teams** | grid (order 30) | `GET /api/v1/teams/detail` (filtered: `type === "standing"`) | `/automations?tab=teams` | 10s |
| **MCP Tools** | grid (order 40) | `GET /api/v1/mcp/servers` | `/settings/tools` | On mount |
| **Missions** | fullWidth (order 50) | `GET /api/v1/missions` + `GET /api/v1/teams/detail` | `/automations?tab=active` | 15s / 10s |
| **Runs** | fullWidth (order 60) | `GET /api/v1/runs` | `/runs` | 10s |

The top 4 grid cards use `grid-cols-[repeat(auto-fit,minmax(240px,1fr))]`; full-width sections stack below. Each card has an ↗ deep-link. The MCP card shows a **Recommended** banner for `brave-search` and `github` if not installed. Mission rows link to run timelines via `/runs/{id}`.

**Widget Registry:** All sections are registered via `registerOpsWidget()` in `lib/opsWidgetRegistry.ts`. Adding a new widget: create a React component, call `registerOpsWidget({ id, order, layout, Component })` — OpsOverview renders all registered widgets automatically. Use `order` multiples of 10 to slot between existing widgets.

#### MCP Baseline Operating Profile (V7 MVOS)

V7 ships a **Minimum Viable Operational Stack** for MCP with curated defaults and library installs.
Current runtime behavior:
- Bootstrap defaults: `filesystem` + `fetch` are auto-installed/connected when available.
- Tool sets seeded at bootstrap: `workspace` (`mcp:filesystem/*`) and `research` (`mcp:fetch/*`).
- Additional curated servers (for example `memory`) are installed from library on demand.
- `artifact-renderer` remains planned in MCP baseline docs and is not currently a default auto-installed server.

| Server | Purpose | Default Config | Risk |
| :--- | :--- | :--- | :--- |
| `filesystem` | Sandboxed file I/O (read, write, list, create, append) | Auto-bootstrap default | Low (read) / Medium (write) |
| `fetch` | Controlled web research (HTTP GET, domain allowlist) | Auto-bootstrap default | Medium |
| `memory` | Semantic store + recall (pgvector-backed) | Curated library install (manual) | Low |
| `artifact-renderer` | Render structured outputs (markdown, JSON, tables, images) | Planned baseline component | Low |

**Default workspace structure** (auto-created on first run):

```text
/workspace
  /projects
  /research
  /artifacts
  /reports
  /exports
```

**Tool risk classification:**

- **Low:** read_file, semantic_search, render_artifact — no confirm required
- **Medium:** write_file, scheduling — confirm required
- **High:** remote provider usage, trigger rule creation, MCP install — always confirm

**All MCP tools emit MissionEventEnvelope events:** `tool.invoked`, `tool.completed`, `tool.failed`, `artifact.created`. Every tool action is traceable in the Run Timeline.

> Detailed specification: [V7 MCP Baseline](docs/V7_MCP_BASELINE.md)
> Full architecture details: [Architecture Overview](docs/architecture/OVERVIEW.md) | [Backend Spec](docs/architecture/BACKEND.md) | [Frontend Spec](docs/architecture/FRONTEND.md) | [Operations Manual](docs/architecture/OPERATIONS.md)

---

## Soma Workflow — End-to-End Reference

This section documents the complete interaction loop from user intent to mission execution, covering both the **GUI path** (browser) and the **API path** (direct HTTP). All paths converge on the same backend — the GUI is a thin consumer of the same endpoints available to any API client.

### Quick Reference

| Step | GUI Component | HTTP Endpoint | Go Handler | Reference |
| :--- | :--- | :--- | :--- | :--- |
| 1. Send message | `MissionControlChat` input | `POST /api/v1/council/admin/chat` | `HandleCouncilChat` | [Chat Panel](#chat-panel--rich-message-rendering) |
| 2. Live activity | `SomaActivityIndicator` | SSE `GET /api/v1/stream` | `HandleSSEStream` | [Chat Capabilities](#chat-capabilities) |
| 3. View delegation | `DelegationTrace` in bubble | ← in chat response body | `processMessageStructured` | [Chat Capabilities](#chat-capabilities) |
| 4. View proposal | `ProposedActionBlock` | ← in chat response body (`mode: proposal`) | `HandleCouncilChat` mutation detector | [Proposal Blocks](#workspace-reference) |
| 5. Confirm mutation | `ProposedActionBlock` confirm btn | `POST /api/v1/intent/confirm-action` | `HandleConfirmAction` | [Confirm API](#5-confirm-action-post-apiv1intentconfirm-action) |
| 6. Mission activated | system message pill in chat | ← in confirm response (`run_id`) | `HandleConfirmAction` | [System Message](#6-mission-activated--system-message) |
| 7. Run timeline | `RunTimeline` at `/runs/{id}` | `GET /api/v1/runs/{id}/events` | `handleGetRunEvents` | [Run Timeline API](#7-run-events-get-apiv1runsid-events) |
| 8. All runs | `RecentRunsSection` in OpsOverview | `GET /api/v1/runs` | `handleListRuns` | [Runs List API](#8-all-runs-get-apiv1runs) |
| 9. Causal chain | `/runs/{id}/chain` (UI pending) | `GET /api/v1/runs/{id}/chain` | `handleGetRunChain` | [Chain API](#9-causal-chain-get-apiv1runsidchain) |

---

### GUI Execution Path

The browser workflow from intent to run timeline — every numbered step corresponds to a distinct UI state.

#### 1. Open Workspace

Navigate to `http://localhost:3000/dashboard`.

```
┌───────────────────────────────────────────────────────────────┐
│  Workspace              SIGNAL: LIVE    [Launch Crew]  [⚙]   │
├───────────────────────────────────────────────────────────────┤
│  Chat header:  ● Soma  [⚡ Direct ▾]                         │
│  (Soma is always primary — [⚡ Direct ▾] opens council list)  │
```

- Component: `MissionControlChat` — Soma header is **locked**; no dropdown
- On mount: `setCouncilTarget('admin')` is called — always resets to Soma
- `LaunchCrewModal` also calls `setCouncilTarget('admin')` on open, clearing stale proposals

#### 2. Send a Message

Type intent in the textarea and press Enter or click Send.

```
You: "Write me a Python CSV parser that handles quoted fields"
```

- Store action: `sendMissionChat(text)` → `POST /api/v1/council/admin/chat`
- NATS routes to Soma (`swarm.council.admin.request`)

#### 3. Live Activity — SomaActivityIndicator

While Soma processes its ReAct loop, the loading state reads `streamLogs` for `tool.invoked` events:

```
⟳  Consulting Coder...       ● ● ●
```

| Tool invoked | Activity label |
| :--- | :--- |
| `consult_council` | `Consulting {member}...` |
| `generate_blueprint` | `Generating mission blueprint...` |
| `research_for_blueprint` | `Researching past missions...` |
| `write_file` | `Writing {path}...` |
| `read_file` | `Reading {path}...` |
| `search_memory` | `Searching memory...` |
| `recall` | `Recalling context...` |
| `store_artifact` | `Storing artifact...` |
| `list_teams` | `Checking active teams...` |
| `get_system_status` | `Reading system status...` |
| *(other)* | `{tool name with underscores as spaces}...` |

- Component: `SomaActivityIndicator` in `MissionControlChat.tsx`
- Data source: `useCortexStore(s => s.streamLogs)` — SSE events from `GET /api/v1/stream`

#### 4. Response Arrives — DelegationTrace

When Soma completes, the message bubble renders:

```
┌─────────────────────────────────────────────────────────────┐
│ Here's a robust CSV parser that handles quoted fields...    │
│                                                             │
│ [markdown content rendered by react-markdown + remark-gfm] │
│                                                             │
│ ─── Soma consulted ─────────────────────────────────────── │
│ ┌──────────────┐  ┌────────────────────────────────────┐   │
│ │ Coder        │  │ ...                                │   │
│ │ Here is a    │  │ (summary of each member's reply)   │   │
│ │ Python CSV   │  │                                    │   │
│ └──────────────┘  └────────────────────────────────────┘   │
│                                                             │
│  [consult_council]  [read_file]           C:0.82  answer   │
└─────────────────────────────────────────────────────────────┘
```

- **DelegationTrace:** renders `msg.consultations[]` — each entry: `{ member, summary }`
- Member color coding: Architect=`cortex-info`, Coder=`cortex-success`, Creative=`cortex-warning`, Sentry=`cortex-danger`
- Summary is first 300 chars of the council member's response
- Tool-use pills show every tool Soma invoked in its ReAct iterations
- Trust badge `C:{score}` from CTS `trust_score` field
- Mode badge: `answer` (read-only) or `proposal` (mutation pending)

#### 5. Confirm a Mutation — ProposedActionBlock

If Soma detects a mutation in its response (file write, mission activation, trigger rule, etc.), the mode is `proposal` and a block appears below the message:

```
┌─────────────────────────────────────────────────────────────┐
│ ⚠ PROPOSED ACTION                                           │
│ Intent: Write Python CSV parser to /workspace/csv_parser.py │
│ Tools:  write_file                           Risk: MEDIUM   │
│                                                             │
│    [✗ Cancel]                [✓ Confirm & Execute]          │
└─────────────────────────────────────────────────────────────┘
```

- Component: `ProposedActionBlock` in `MissionControlChat.tsx`
- Store: `pendingProposal` + `activeConfirmToken` hold the CE-1 confirm token (15min TTL)
- "Cancel" → `cancelProposal()` clears both
- "Confirm" → `confirmProposal()` → `POST /api/v1/intent/confirm-action`

#### 6. Mission Activated — System Message

After confirm, a green pill appears in chat:

```
         ⚡ Mission activated — abc1234...  ↗
```

- Role: `system` in `ChatMessage` — renders centered, not as a normal bubble
- Clicking navigates to `/runs/{run_id}`
- If `run_id` is null (lightweight chat proposal, no blueprint), pill shows plain "Mission activated" text

#### 7. Run Timeline — `/runs/{id}`

Full-page vertical event timeline. Auto-polls every 5 seconds.

```
┌─────────────────────────────────────────────────────────────┐
│  ← Workspace    Run: abc1234-aaaa-...    ● running          │
│  Started: 12s ago    [↺ Refresh]    (auto-refresh)         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ●────── mission.started        soma · admin-core   12s    │
│  │       {"mission_id": "m-abc1234"}                       │
│  │                                                          │
│  ●────── tool.invoked           coder · council    10s     │
│  │       write_file → /workspace/csv_parser.py             │
│  │                                                          │
│  ●────── tool.completed         coder · council     8s     │
│  │       write_file ✓  [▸ show payload]                    │
│  │                                                          │
│  ●        mission.completed     soma · admin-core    4s    │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

- Component: `RunTimeline.tsx` → `EventCard.tsx` per event
- Polling: `setInterval(fetch, 5000)` — clears when `mission.completed`, `mission.failed`, or `mission.cancelled` is detected
- Event dot colors: `mission.*` → green/red, `tool.invoked` → cyan, `tool.completed` → info, `tool.failed` → red, `memory.*` → amber, `artifact.created` → amber

#### 8. OpsOverview Runs Widget

The Runs card in the dashboard lower pane shows all recent runs, polling every 10s:

```
┌─────────────────────────────────────────── Runs  ↗ ─────┐
│  ● abc1234  running   12s ago  ⚡                        │
│  ● def5678  completed  5m ago  ⚡                        │
│  ● ghi9012  failed    12m ago  ⚡                        │
└──────────────────────────────────────────────────────────┘
```

Clicking any row navigates to `/runs/{id}`.

#### LaunchCrewModal Alternative Path

A guided 3-step modal — use when you want to define intent before committing to a conversation thread:

1. **Step 1:** "Launch Crew" button → modal opens → type mission intent → Send
   - `setCouncilTarget('admin')` called before `sendMissionChat` — always Soma
   - `cancelProposal()` called on open — clears any stale pending proposal
2. **Step 2:** Waiting for Soma to process (SomaActivityIndicator in modal)
3. **Step 3:** ProposedActionBlock appears inside modal → review → "Launch Crew" button

---

### API Execution Path

The same workflow executed via direct HTTP. All endpoints require `Authorization: Bearer {MYCELIS_API_KEY}`.

#### 1. Send Message — `POST /api/v1/council/admin/chat`

```http
POST /api/v1/council/admin/chat
Authorization: Bearer mycelis-dev-key-change-in-prod
Content-Type: application/json

{
  "message": "Write me a Python CSV parser that handles quoted fields",
  "history": []
}
```

Response (CTS envelope wrapping `ChatResponsePayload`):

```json
{
  "signal_type": "chat.response",
  "source": "admin-core",
  "trust_score": 0.82,
  "payload": {
    "text": "Here's a robust CSV parser...",
    "tools_used": ["consult_council", "write_file"],
    "consultations": [
      {
        "member": "council-coder",
        "summary": "Here is a Python CSV parser implementation that handles..."
      }
    ],
    "mode": "answer",
    "provider_id": "ollama",
    "model_used": "qwen2.5-coder:7b",
    "artifacts": [],
    "brain_provenance": {
      "provider_id": "ollama",
      "model_id": "qwen2.5-coder:7b",
      "profile": "admin"
    }
  }
}
```

When Soma detects a mutation (`mode: "proposal"`), the response also includes:

```json
{
  "payload": {
    "mode": "proposal",
    "proposed_action": {
      "template_id": "write-file-v1",
      "intent_proof_id": "ip-xyz",
      "confirm_token": "ct-abc123",
      "description": "Write /workspace/csv_parser.py",
      "tools": ["write_file"],
      "risk_level": "medium"
    }
  }
}
```

> **Note:** The admin chat endpoint is `POST /api/v1/council/admin/chat`. Other council members are `POST /api/v1/council/{member}/chat` where `{member}` is `architect`, `coder`, `creative`, or `sentry`. Use `GET /api/v1/council/members` to list all available members dynamically.

#### 2. Get Available Council Members — `GET /api/v1/council/members`

```http
GET /api/v1/council/members
Authorization: Bearer mycelis-dev-key-change-in-prod
```

```json
{
  "ok": true,
  "data": [
    { "id": "admin", "name": "Admin", "team": "admin-core", "online": true },
    { "id": "architect", "name": "Architect", "team": "council-core", "online": true },
    { "id": "coder", "name": "Coder", "team": "council-core", "online": true },
    { "id": "creative", "name": "Creative", "team": "council-core", "online": true },
    { "id": "sentry", "name": "Sentry", "team": "council-core", "online": true }
  ]
}
```

#### 3. Direct Council Chat — `POST /api/v1/council/{member}/chat`

Skip Soma entirely and send directly to a specialist:

```http
POST /api/v1/council/coder/chat
Authorization: Bearer mycelis-dev-key-change-in-prod
Content-Type: application/json

{
  "message": "Review this Go code for race conditions",
  "history": []
}
```

Response shape is identical to admin chat, but `consultations[]` will be empty (Coder doesn't delegate back).

#### 4. Broadcast to All Teams — `POST /api/v1/chat` with `/all` prefix

```http
POST /api/v1/chat
Authorization: Bearer mycelis-dev-key-change-in-prod
Content-Type: application/json

{
  "message": "/all What is the current system status?"
}
```

Sends to every active team via NATS; collects all responses (60s timeout per team).

#### 5. Confirm Action — `POST /api/v1/intent/confirm-action`

```http
POST /api/v1/intent/confirm-action
Authorization: Bearer mycelis-dev-key-change-in-prod
Content-Type: application/json

{
  "confirm_token": "ct-abc123"
}
```

```json
{
  "status": "confirmed",
  "run_id": "aaaa1111-aaaa-aaaa-aaaa-aaaaaaaaaaaa"
}
```

`run_id` is `null` for lightweight chat proposals that don't activate a mission blueprint. For full blueprint executions it will be a UUID — use it to poll the run timeline.

#### 6. Mission Activated — System Message

> GUI only — the system message pill is a frontend construct. In the API path, the `run_id` from the confirm response is your handle.

#### 7. Run Events — `GET /api/v1/runs/{id}/events`

```http
GET /api/v1/runs/aaaa1111-aaaa-aaaa-aaaa-aaaaaaaaaaaa/events
Authorization: Bearer mycelis-dev-key-change-in-prod
```

```json
[
  {
    "id": "ev-1",
    "run_id": "aaaa1111-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
    "tenant_id": "default",
    "event_type": "mission.started",
    "severity": "info",
    "source_agent": "soma",
    "source_team": "admin-core",
    "payload": { "mission_id": "m-abc1234" },
    "audit_event_id": "",
    "emitted_at": "2026-02-23T10:15:00Z"
  },
  {
    "id": "ev-2",
    "run_id": "aaaa1111-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
    "tenant_id": "default",
    "event_type": "tool.invoked",
    "severity": "info",
    "source_agent": "coder",
    "source_team": "council-core",
    "payload": { "tool": "write_file", "path": "/workspace/csv_parser.py" },
    "audit_event_id": "",
    "emitted_at": "2026-02-23T10:15:01Z"
  }
]
```

Poll this endpoint every 5s until you see `mission.completed`, `mission.failed`, or `mission.cancelled`.

**Event types and their meaning:**

| event_type | Emitter | Meaning |
| :--- | :--- | :--- |
| `mission.started` | Soma/activation | A mission run began |
| `mission.completed` | Soma | Run finished successfully |
| `mission.failed` | Soma | Run ended with error |
| `tool.invoked` | Any agent | ReAct loop called a tool |
| `tool.completed` | Any agent | Tool returned successfully |
| `tool.failed` | Any agent | Tool returned an error |
| `agent.started` | Team manager | An agent goroutine started |
| `memory.stored` | Archivist / MCP memory | Fact persisted to pgvector |
| `memory.recalled` | Any agent | Semantic search returned results |
| `artifact.created` | Any agent | File/chart/code artifact stored |
| `trigger.fired` | Trigger engine | A trigger rule activated |
| `schedule.fired` | Scheduler | A scheduled mission launched |

#### 8. All Runs — `GET /api/v1/runs`

```http
GET /api/v1/runs
Authorization: Bearer mycelis-dev-key-change-in-prod
```

```json
{
  "ok": true,
  "data": [
    {
      "id": "aaaa1111-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
      "mission_id": "m-abc1234",
      "tenant_id": "default",
      "status": "completed",
      "run_depth": 0,
      "parent_run_id": "",
      "started_at": "2026-02-23T10:15:00Z",
      "completed_at": "2026-02-23T10:15:08Z"
    }
  ]
}
```

Returns the 20 most recent runs for `tenant_id = "default"`, newest first.

#### 9. Causal Chain — `GET /api/v1/runs/{id}/chain`

```http
GET /api/v1/runs/aaaa1111-aaaa-aaaa-aaaa-aaaaaaaaaaaa/chain
Authorization: Bearer mycelis-dev-key-change-in-prod
```

```json
{
  "run_id": "aaaa1111-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
  "mission_id": "m-abc1234",
  "chain": [
    {
      "id": "aaaa1111-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
      "mission_id": "m-abc1234",
      "status": "running",
      "run_depth": 0,
      "started_at": "2026-02-23T10:15:00Z"
    },
    {
      "id": "bbbb2222-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
      "mission_id": "m-abc1234",
      "status": "completed",
      "run_depth": 0,
      "started_at": "2026-02-23T09:00:00Z",
      "completed_at": "2026-02-23T09:01:00Z"
    }
  ]
}
```

Returns all runs for the same `mission_id` as the target run — newest first. Use `parent_run_id` and `run_depth` fields to reconstruct trigger-chain trees. UI: `ViewChain.tsx` (pending).

---

### Trigger Rules — Automation Workflow

Trigger rules let you wire missions into reactive chains: "when event X happens, launch mission Y." This section covers creating rules, understanding how they fire, viewing execution history, and building multi-mission chains.

#### Overview

```
Mission A completes
    │
    ▼
CTS event: mission.completed  ──► Trigger Engine evaluates rules
    │                                    │
    │                         ┌──────────┼──────────────┐
    │                         ▼          ▼              ▼
    │                    Guard 1     Guard 2        Guard 3
    │                    Cooldown    Recursion      Concurrency
    │                    (min secs)  (max depth)    (max active)
    │                         │          │              │
    │                         └──────────┼──────────────┘
    │                                    │
    │                           All guards pass?
    │                            ╱          ╲
    │                         Yes            No
    │                          │              │
    │                 mode=propose?      logSkip()
    │                  ╱        ╲         (audit record)
    │               Yes          No
    │                │            │
    │         proposeTrigger  fireTrigger
    │         (await human    (auto-create
    │          approval)       child run)
    │                            │
    │                            ▼
    │                    Mission B starts
    │                    (run_depth + 1)
```

#### 10. List Trigger Rules — `GET /api/v1/triggers`

```http
GET /api/v1/triggers
Authorization: Bearer mycelis-dev-key-change-in-prod
```

```json
{
  "ok": true,
  "data": [
    {
      "id": "aaaa1111-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
      "tenant_id": "default",
      "name": "On CSV parse complete → analyze",
      "description": "When a CSV parsing mission completes, trigger the analysis mission",
      "event_pattern": "mission.completed",
      "condition": {},
      "target_mission_id": "m-analyze-csv",
      "mode": "propose",
      "cooldown_seconds": 60,
      "max_depth": 5,
      "max_active_runs": 3,
      "is_active": true,
      "last_fired_at": "2026-02-25T10:30:00Z",
      "created_at": "2026-02-25T10:00:00Z",
      "updated_at": "2026-02-25T10:00:00Z"
    }
  ]
}
```

#### 11. Create Trigger Rule — `POST /api/v1/triggers`

```http
POST /api/v1/triggers
Authorization: Bearer mycelis-dev-key-change-in-prod
Content-Type: application/json

{
  "name": "On file write → run tests",
  "description": "When any tool writes a file, propose running the test suite",
  "event_pattern": "tool.completed",
  "target_mission_id": "m-run-tests",
  "mode": "propose",
  "cooldown_seconds": 120,
  "max_depth": 3,
  "max_active_runs": 2,
  "is_active": true
}
```

```json
{
  "ok": true,
  "data": {
    "id": "bbbb2222-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
    "name": "On file write → run tests",
    "event_pattern": "tool.completed",
    "target_mission_id": "m-run-tests",
    "mode": "propose",
    "cooldown_seconds": 120,
    "max_depth": 3,
    "max_active_runs": 2,
    "is_active": true,
    "created_at": "2026-02-25T11:00:00Z",
    "updated_at": "2026-02-25T11:00:00Z"
  }
}
```

**Required fields:** `name`, `event_pattern`, `target_mission_id`.

**Defaults applied if omitted:** `mode` → `"propose"`, `cooldown_seconds` → `60`, `max_depth` → `5`, `max_active_runs` → `3`, `condition` → `{}`.

**Available event patterns:**

| event_pattern | Fires when |
| :--- | :--- |
| `mission.started` | A mission run begins |
| `mission.completed` | A mission run finishes successfully |
| `mission.failed` | A mission run ends with an error |
| `tool.invoked` | Any agent invokes a tool in its ReAct loop |
| `tool.completed` | A tool returns successfully |
| `tool.failed` | A tool returns an error |
| `agent.started` | An agent goroutine starts within a team |
| `memory.stored` | A fact is persisted to pgvector |
| `memory.recalled` | Semantic search returns results |
| `artifact.created` | A file, chart, or code artifact is stored |
| `schedule.fired` | A scheduled mission launches |

**Mode values:**

| Mode | Behavior | Use when |
| :--- | :--- | :--- |
| `propose` (default) | Logs as "proposed" — awaits human approval before creating a child run | Normal operation — human-in-the-loop |
| `auto_execute` | Immediately creates a child run and fires the target mission | Trusted chains — e.g. test suite after build |

> **Safety:** `auto_execute` requires explicit intent. Any other value (including omitted or invalid) defaults to `propose`.

#### 12. Update Trigger Rule — `PUT /api/v1/triggers/{id}`

```http
PUT /api/v1/triggers/bbbb2222-bbbb-bbbb-bbbb-bbbbbbbbbbbb
Authorization: Bearer mycelis-dev-key-change-in-prod
Content-Type: application/json

{
  "name": "On file write → run tests (v2)",
  "event_pattern": "tool.completed",
  "target_mission_id": "m-run-tests-v2",
  "mode": "auto_execute",
  "cooldown_seconds": 60,
  "max_depth": 5,
  "max_active_runs": 3,
  "is_active": true
}
```

```json
{ "ok": true, "data": { "id": "bbbb2222-...", "updated": true } }
```

#### 13. Delete Trigger Rule — `DELETE /api/v1/triggers/{id}`

```http
DELETE /api/v1/triggers/bbbb2222-bbbb-bbbb-bbbb-bbbbbbbbbbbb
Authorization: Bearer mycelis-dev-key-change-in-prod
```

```json
{ "ok": true, "data": { "id": "bbbb2222-...", "deleted": true } }
```

#### 14. Toggle Rule Active/Inactive — `POST /api/v1/triggers/{id}/toggle`

```http
POST /api/v1/triggers/bbbb2222-bbbb-bbbb-bbbb-bbbbbbbbbbbb/toggle
Authorization: Bearer mycelis-dev-key-change-in-prod
Content-Type: application/json

{ "is_active": false }
```

```json
{ "ok": true, "data": { "id": "bbbb2222-...", "is_active": false } }
```

Deactivated rules stay in the database but are removed from the in-memory evaluation cache — zero cost on event ingest.

#### 15. Trigger Execution History — `GET /api/v1/triggers/{id}/history`

Every evaluation is logged — whether it fired, proposed, or skipped.

```http
GET /api/v1/triggers/bbbb2222-bbbb-bbbb-bbbb-bbbbbbbbbbbb/history
Authorization: Bearer mycelis-dev-key-change-in-prod
```

```json
{
  "ok": true,
  "data": [
    {
      "id": "exec-1",
      "rule_id": "bbbb2222-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
      "event_id": "ev-aaa",
      "run_id": "run-child-1",
      "status": "fired",
      "skip_reason": "",
      "executed_at": "2026-02-25T11:05:00Z"
    },
    {
      "id": "exec-2",
      "rule_id": "bbbb2222-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
      "event_id": "ev-bbb",
      "run_id": "",
      "status": "skipped",
      "skip_reason": "cooldown: 30s since last fire, cooldown is 120s",
      "executed_at": "2026-02-25T11:05:30Z"
    },
    {
      "id": "exec-3",
      "rule_id": "bbbb2222-bbbb-bbbb-bbbb-bbbbbbbbbbbb",
      "event_id": "ev-ccc",
      "run_id": "",
      "status": "proposed",
      "skip_reason": "",
      "executed_at": "2026-02-25T11:07:00Z"
    }
  ]
}
```

**Status values:**

| Status | Meaning | `run_id` |
| :--- | :--- | :--- |
| `fired` | Guards passed, child run created | Set — the child run |
| `proposed` | Guards passed, awaiting human approval | Empty — no run yet |
| `skipped` | A guard blocked execution | Empty |

**Skip reasons:**

| Reason | Explanation |
| :--- | :--- |
| `cooldown: Xs since last fire, cooldown is Ys` | Rule fired too recently |
| `recursion_limit: source run depth X >= max Y` | Trigger chain too deep |
| `concurrency_limit: X active runs >= max Y` | Too many runs already in-flight |

### Conversation Log & Interjection — Agent Transcript Browsing

Full-fidelity agent conversation transcripts stored per run, with operator interjection for mid-run redirection.

#### 16. Run Conversation — `GET /api/v1/runs/{id}/conversation`

Returns all conversation turns for a run, ordered chronologically. Optional `?agent=X` filter.

```http
GET /api/v1/runs/aaaa1111-aaaa-aaaa-aaaa-aaaaaaaaaaaa/conversation?agent=admin
Authorization: Bearer mycelis-dev-key-change-in-prod
```

```json
{
  "ok": true,
  "data": [
    { "id": "...", "session_id": "...", "agent_id": "admin", "turn_index": 0,
      "role": "user", "content": "Analyze our deployment status", "created_at": "..." },
    { "id": "...", "agent_id": "admin", "turn_index": 1,
      "role": "tool_call", "tool_name": "consult_council",
      "tool_args": {"member":"council-architect","question":"..."}, "created_at": "..." },
    { "id": "...", "agent_id": "admin", "turn_index": 2,
      "role": "assistant", "content": "Based on my analysis...",
      "provider_id": "ollama", "model_used": "qwen2.5-coder:7b", "created_at": "..." }
  ]
}
```

#### 17. Session Conversation — `GET /api/v1/conversations/{session_id}`

Returns turns for a specific session (useful for standing-team chats without a run).

#### 18. Operator Interjection — `POST /api/v1/runs/{id}/interject`

Redirects active agents mid-run by publishing to their NATS interjection mailbox.

```http
POST /api/v1/runs/aaaa1111-aaaa-aaaa-aaaa-aaaaaaaaaaaa/interject
Authorization: Bearer mycelis-dev-key-change-in-prod
Content-Type: application/json

{ "message": "Focus on database schema, not deployment", "agent_id": "admin" }
```

If `agent_id` is empty, the interjection is broadcast to all agents on the run's teams. The interjection is also logged as a `role=interjection` conversation turn.

### Inception Recipes — Structured Prompt Patterns

Inception recipes capture proven "how to ask for X" patterns that agents distill after completing complex tasks.

#### 19. List Recipes — `GET /api/v1/inception/recipes`

```http
GET /api/v1/inception/recipes?category=blueprint&limit=10
Authorization: Bearer mycelis-dev-key-change-in-prod
```

Optional query params: `category`, `agent`, `limit` (default 20). Results ordered by usage count + quality score.

#### 20. Search Recipes — `GET /api/v1/inception/recipes/search`

Trigram text search on recipe titles and intent patterns.

```http
GET /api/v1/inception/recipes/search?q=microservices&limit=5
Authorization: Bearer mycelis-dev-key-change-in-prod
```

#### 21. Get Recipe — `GET /api/v1/inception/recipes/{id}`

Returns a single recipe with all fields including parameters, example prompt, and outcome shape.

#### 22. Create Recipe — `POST /api/v1/inception/recipes`

Manual recipe creation (agents also create recipes via the `store_inception_recipe` internal tool).

```http
POST /api/v1/inception/recipes
Authorization: Bearer mycelis-dev-key-change-in-prod
Content-Type: application/json

{
  "category": "blueprint",
  "title": "How to create a microservices deployment blueprint",
  "intent_pattern": "Create a blueprint for deploying N microservices with load balancing...",
  "parameters": { "service_count": "number of services", "orchestrator": "k8s or docker-compose" },
  "example_prompt": "I need a blueprint to deploy 5 microservices on Kubernetes with Istio...",
  "outcome_shape": "A structured blueprint with service definitions, networking, and scaling policies",
  "tags": ["blueprint", "deployment", "kubernetes"]
}
```

#### 23. Quality Feedback — `PATCH /api/v1/inception/recipes/{id}/quality`

Updates the quality score (0.0–1.0) for a recipe. Usage count is incremented automatically on recall.

```http
PATCH /api/v1/inception/recipes/cccc3333-.../quality
Content-Type: application/json

{ "score": 0.85 }
```

#### Guard Reference

Three mandatory guards protect every trigger evaluation:

| Guard | Config field | Default | Purpose |
| :--- | :--- | :--- | :--- |
| Cooldown | `cooldown_seconds` | 60 | Minimum seconds between firings — prevents rapid-fire loops |
| Recursion | `max_depth` | 5 (ceiling: 10) | Maximum trigger chain depth — A triggers B triggers C... stops at depth limit |
| Concurrency | `max_active_runs` | 3 | Maximum in-flight runs for the target mission — prevents resource exhaustion |

Guards are evaluated in order: cooldown → recursion → concurrency. First failure skips the rule; all must pass.

#### GUI Path — Automations → Trigger Rules Tab

Navigate to `/automations?tab=triggers`:

```
┌──────────────────────────────────────────────────────────────────┐
│  Automations    Active · Drafts · [Triggers] · Approvals · ...   │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  [+ Create Trigger Rule]                                         │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  On file write → run tests              ● active           │  │
│  │  mission.completed → m-run-tests        propose            │  │
│  │  ⏱ 120s   ↕ max 3   ⚡ max 2                              │  │
│  │                                    [Toggle] [Delete]       │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐  │
│  │  Nightly report chain                    ○ inactive        │  │
│  │  schedule.fired → m-generate-report      auto_execute      │  │
│  │  ⏱ 3600s  ↕ max 2   ⚡ max 1                              │  │
│  └────────────────────────────────────────────────────────────┘  │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

- **Guard badges:** ⏱ cooldown, ↕ max depth, ⚡ max active runs
- **Mode badge:** `propose` (cyan) or `auto_execute` (amber with warning)
- **Create form:** name, description, event pattern dropdown (11 types), target mission ID, mode, guard values
- **Auto-execute warning:** banner shown when selecting `auto_execute` mode — requires explicit acknowledgment

#### Typical Workflows

**One-shot trigger (propose mode):**

1. Create rule: `event_pattern = "mission.completed"`, `mode = "propose"`, `is_active = true`
2. Run a mission — when it completes, the trigger engine logs a proposal
3. Review in Automations → Approvals tab → confirm or reject
4. On confirm, child run is created with `run_depth + 1`

**Automated chain (auto_execute mode):**

1. Create rule: `event_pattern = "tool.completed"`, `mode = "auto_execute"`, `is_active = true`
2. When any tool completes, the engine immediately creates a child run
3. The child run starts at `run_depth = parent_depth + 1`
4. If the child mission also completes and another trigger matches, the chain continues
5. Recursion guard stops at `max_depth` — prevents infinite loops

**Monitoring trigger health:**

1. `GET /api/v1/triggers/{id}/history` — view every evaluation (fired/skipped/proposed)
2. Filter by `status = "skipped"` to identify over-aggressive cooldowns or depth limits
3. Use `GET /api/v1/runs` to see child runs with `parent_run_id` set
4. Use `GET /api/v1/runs/{child_id}/events` to view the child run's event timeline
5. Use `GET /api/v1/runs/{id}/chain` to traverse the full parent → child → grandchild tree

**Combining triggers with the scheduler (Team C — pending):**

1. Scheduler fires `schedule.fired` events on cron intervals
2. A trigger rule on `event_pattern = "schedule.fired"` chains into a mission
3. This creates recurring → reactive chains: scheduler fires → trigger evaluates → mission runs → outputs become events for more triggers

---

### Workflow State Diagram

```
User input
    │
    ▼
POST /api/v1/council/admin/chat
    │
    ├─── NATS: swarm.council.admin.request ──► Soma ReAct loop
    │                                              │
    │                                    ┌─────────┴──────────┐
    │                                    │  tool: consult_council
    │                                    │  NATS: swarm.council.{member}.request
    │                                    │  ◄── council reply (summary)
    │                                    │  ConsultationEntry appended
    │                                    └─────────┬──────────┘
    │                                              │
    │                                    Soma synthesizes
    │                                    mode = answer | proposal
    │
    ◄── ChatResponsePayload (CTS envelope)
         │
         ├── consultations[] → DelegationTrace in UI
         ├── tools_used[]    → tool-use pills in UI
         │
         └── mode == "proposal"?
                │
                ▼
         ProposedActionBlock shown
                │
         User clicks Confirm
                │
                ▼
         POST /api/v1/intent/confirm-action
                │ { confirm_token }
                │
                ◄── { status: "confirmed", run_id: "uuid" }
                │
                ├── system message pill in chat ──► click ──► /runs/{id}
                │
                └── run_id ──► GET /api/v1/runs/{id}/events (poll 5s)
                                   │
                                   └── EventCard timeline in RunTimeline.tsx
                                           │
                                   Poll until terminal event:
                                   mission.completed | mission.failed | mission.cancelled
```

---

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

> **V7 Navigation (Active):** 5 workflow-first panels — Workspace, Automations, Resources, Memory, System (advanced). Legacy architecture-surface routes (`/wiring`, `/catalogue`, `/matrix`, etc.) redirect to their workflow parent with tab deep-linking.

| Route | Description |
| :--- | :--- |
| `/` | Product landing page (marketing) — links to `/dashboard` to launch console |
| `/dashboard` | **Workspace** — Resizable chat-dominant layout (Launch Crew button, SomaOfflineGuide, council chat), OpsOverview dashboard, telemetry sparklines |
| `/automations` | **Automations** — 6 tabs: Active Automations, Draft Blueprints, Trigger Rules, Approvals + Policy, Teams, Neural Wiring (Advanced Mode only) |
| `/resources` | **Resources** — 4 tabs: Brains, MCP Tools, Workspace Explorer, Capabilities |
| `/memory` | **Memory** — 2-column: Warm sitreps/artifacts (left) + Cold semantic search (right). Hot signal stream collapsible under Advanced Mode. |
| `/system` | **System** (Advanced Mode only) — 5 tabs: Event Health, NATS Status, Database, Cognitive Matrix, Debug |
| `/docs` | **In-App Docs** — Two-column documentation browser (sidebar + rendered markdown). Serves docs via `GET /docs-api` (manifest) + `GET /docs-api/[slug]` (content). `/docs-api` prefix avoids the `/api/*` → Go proxy rewrite. Curated sections: User Guides, Getting Started, Soma Workflow, API Reference, Architecture, Governance & Testing, V7 Development. Add entries to `lib/docsManifest.ts`. Deep-link: `/docs?doc={slug}`. |
| `/settings` | Settings — Brains (provider CRUD), Profiles (mission profiles + activate), Cognitive Matrix, MCP Tools, Users |
| `/runs` | Run List — all recent runs across missions, status dots, timestamps. Navigates to timeline on click. |
| `/runs/[id]` | Run Timeline — vertical `EventCard` timeline, auto-polls every 5s, stops on terminal events (`mission.completed/failed`). |
| `/runs/[id]/chain` | Causal Chain View — parent/child run traversal (V7 — backend `GET /api/v1/runs/{id}/chain` complete; UI pending). |
| `/wiring` | Server redirect → `/automations?tab=wiring` |
| `/architect` | Server redirect → `/automations?tab=wiring` |
| `/teams` | Server redirect → `/automations?tab=teams` |
| `/catalogue` | Server redirect → `/resources?tab=catalogue` |
| `/marketplace` | Server redirect → `/resources?tab=catalogue` |
| `/approvals` | Server redirect → `/automations?tab=approvals` |
| `/telemetry` | Server redirect → `/system?tab=health` |
| `/matrix` | Server redirect → `/system?tab=matrix` |

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

> **In-app doc browser:** Navigate to `/docs` in the running UI to browse all docs below with rendered markdown, sidebar navigation, and search. Add new entries to [interface/lib/docsManifest.ts](interface/lib/docsManifest.ts).

| Topic | Document | In-App |
| :--- | :--- | :--- |
| **User Guides** | `docs/user/` — 7 plain-language guides for every workflow and concept | |
| ↳ Core Concepts | [docs/user/core-concepts.md](docs/user/core-concepts.md) — Soma, Council, Mission, Run, Brain, Event, Trust | [/docs?doc=core-concepts](/docs?doc=core-concepts) |
| ↳ Using Soma Chat | [docs/user/soma-chat.md](docs/user/soma-chat.md) — Send messages, delegation traces, confirm proposals | [/docs?doc=soma-chat](/docs?doc=soma-chat) |
| ↳ Meta-Agent & Blueprints | [docs/user/meta-agent-blueprint.md](docs/user/meta-agent-blueprint.md) — Architect as meta-agent, blueprint structure, team/agent/tool planning | [/docs?doc=meta-agent-blueprint](/docs?doc=meta-agent-blueprint) |
| ↳ Run Timeline | [docs/user/run-timeline.md](docs/user/run-timeline.md) — Reading execution timelines, event types, navigation | [/docs?doc=run-timeline](/docs?doc=run-timeline) |
| ↳ Automations | [docs/user/automations.md](docs/user/automations.md) — Triggers, schedules, approvals, teams | [/docs?doc=automations-guide](/docs?doc=automations-guide) |
| ↳ Resources | [docs/user/resources.md](docs/user/resources.md) — Brains, MCP tools, workspace, catalogue | [/docs?doc=resources-guide](/docs?doc=resources-guide) |
| ↳ Memory | [docs/user/memory.md](docs/user/memory.md) — Semantic search, SitReps, artifacts, hot/warm/cold | [/docs?doc=memory-guide](/docs?doc=memory-guide) |
| ↳ Governance & Trust | [docs/user/governance-trust.md](docs/user/governance-trust.md) — Trust scores, approvals, policy, propose vs execute | [/docs?doc=governance-trust](/docs?doc=governance-trust) |
| **Overview** | [README.md](README.md) — Architecture, stack, commands, current phase | [/docs?doc=readme](/docs?doc=readme) |
| **Local Dev Workflow** | [docs/LOCAL_DEV_WORKFLOW.md](docs/LOCAL_DEV_WORKFLOW.md) — Setup, config reference, port map, troubleshooting | [/docs?doc=local-dev](/docs?doc=local-dev) |
| **Soma Workflow** | [docs/WORKFLOWS.md](docs/WORKFLOWS.md) — End-to-end GUI + API workflow reference | [/docs?doc=workflows](/docs?doc=workflows) |
| **MVP Agentry Plan** | [docs/archive/MVP_AGENTRY_PLAN.md](docs/archive/MVP_AGENTRY_PLAN.md) — Full agentry chain map: User → Workspace → NATS → Soma | [/docs?doc=mvp-agentry](/docs?doc=mvp-agentry) |
| **Council Chat QA** | [docs/QA_COUNCIL_CHAT_API.md](docs/QA_COUNCIL_CHAT_API.md) — QA procedures and test cases for council chat | [/docs?doc=council-chat-qa](/docs?doc=council-chat-qa) |
| **API Reference** | [docs/API_REFERENCE.md](docs/API_REFERENCE.md) — Full endpoint table (80+ routes) | [/docs?doc=api-reference](/docs?doc=api-reference) |
| **Architecture Overview** | [docs/architecture/OVERVIEW.md](docs/architecture/OVERVIEW.md) — Philosophy, 4-layer anatomy, phases, upcoming roadmap | [/docs?doc=arch-overview](/docs?doc=arch-overview) |
| **Backend Specification** | [docs/architecture/BACKEND.md](docs/architecture/BACKEND.md) — Go packages, APIs, DB schema, NATS, execution pipelines | [/docs?doc=arch-backend](/docs?doc=arch-backend) |
| **Frontend Specification** | [docs/architecture/FRONTEND.md](docs/architecture/FRONTEND.md) — Routes, components, Zustand, design system | [/docs?doc=arch-frontend](/docs?doc=arch-frontend) |
| **Operations Manual** | [docs/architecture/OPERATIONS.md](docs/architecture/OPERATIONS.md) — Deployment, config, testing, CI/CD | [/docs?doc=arch-operations](/docs?doc=arch-operations) |
| **Memory Service** | [docs/architecture/DIRECTIVE_MEMORY_SERVICE.md](docs/architecture/DIRECTIVE_MEMORY_SERVICE.md) — State Engine, event projection, pgvector schema | [/docs?doc=arch-memory-service](/docs?doc=arch-memory-service) |
| **V7 Architecture PRD** | [mycelis-architecture-v7.md](mycelis-architecture-v7.md) — V7 product requirements: event spine, mission graph, observability | [/docs?doc=v7-architecture-prd](/docs?doc=v7-architecture-prd) |
| **V7 UI Framework** | [docs/UI_FRAMEWORK_V7.md](docs/UI_FRAMEWORK_V7.md) — Default UI instantiation contract (state model, failure templates, testing matrix, PR gate) | [/docs?doc=v7-ui-framework](/docs?doc=v7-ui-framework) |
| **V7 UI Parallel Delivery** | [docs/ui-delivery/PARALLEL_DELIVERY_BOARD.md](docs/ui-delivery/PARALLEL_DELIVERY_BOARD.md) — Active gate model + lane matrix (A/B/C/D/Q) with evidence tracking | [/docs?doc=v7-ui-parallel-delivery](/docs?doc=v7-ui-parallel-delivery) |
| **V7 MCP Baseline** | [docs/V7_MCP_BASELINE.md](docs/V7_MCP_BASELINE.md) — MVOS: filesystem, memory, artifact-renderer, fetch | [/docs?doc=v7-mcp-baseline](/docs?doc=v7-mcp-baseline) |
| **Swarm Operations** | [docs/SWARM_OPERATIONS.md](docs/SWARM_OPERATIONS.md) — Hierarchy, blueprints, activation, teams, tools, governance | [/docs?doc=swarm-operations](/docs?doc=swarm-operations) |
| **Cognitive Architecture** | [docs/COGNITIVE_ARCHITECTURE.md](docs/COGNITIVE_ARCHITECTURE.md) — Providers, profiles, matrix UI, embedding | [/docs?doc=cognitive-architecture](/docs?doc=cognitive-architecture) |
| **Signal Log Schema** | [docs/logging.md](docs/logging.md) — LogEntry format, NATS cortex.logs subject, field reference | [/docs?doc=logging-schema](/docs?doc=logging-schema) |
| **Governance** | [docs/governance.md](docs/governance.md) — Policy enforcement, approvals, security | [/docs?doc=governance](/docs?doc=governance) |
| **Testing** | [docs/TESTING.md](docs/TESTING.md) — Unit, integration, smoke protocols | [/docs?doc=testing](/docs?doc=testing) |
| **V7 UI Verification** | [docs/verification/v7-step-01-ui.md](docs/verification/v7-step-01-ui.md) — Manual UI checklist for V7 Step 01 navigation | [/docs?doc=v7-ui-verification](/docs?doc=v7-ui-verification) |
| **V7 Implementation Plan** | [docs/V7_IMPLEMENTATION_PLAN.md](docs/V7_IMPLEMENTATION_PLAN.md) — Teams A/B/C/D/E technical plan | [/docs?doc=v7-implementation-plan](/docs?doc=v7-implementation-plan) |
| **V7 Dev State** | [V7_DEV_STATE.md](V7_DEV_STATE.md) — Authoritative map of what's done vs pending | [/docs?doc=v7-dev-state](/docs?doc=v7-dev-state) |
| **IA Step 01** | [docs/product/ia-v7-step-01.md](docs/product/ia-v7-step-01.md) — Workflow-first navigation PRD and decisions | [/docs?doc=v7-ia-step01](/docs?doc=v7-ia-step01) |
| **Registry** | [core/internal/registry/README.md](core/internal/registry/README.md) — Connector marketplace | — |
| **Core API** | [core/README.md](core/README.md) — Go service architecture | — |
| **CLI** | [cli/README.md](cli/README.md) — `myc` command-line tool | — |
| **Interface** | [interface/README.md](interface/README.md) — Next.js frontend architecture | — |

## Verification

```bash
uvx inv core.test             # Go unit tests (188 tests across 16 packages — server, events, runs, swarm, governance, ...)
uvx inv interface.test        # Vitest component tests (~70 V7 tests, 56 pass, 2 pre-existing DashboardPage failures)
uvx inv interface.e2e         # Playwright E2E specs (requires running servers)
uvx inv interface.check       # HTTP smoke test against running dev server (9 pages)
uvx inv core.smoke            # Governance smoke tests
cd interface && npx playwright test e2e/specs/v7-operational-ux.spec.ts  # Gate A operational UX E2E (degraded banner, status drawer, council reroute, automations hub, quick checks, focus mode)
cd core && go test ./internal/mcp/ -count=1
cd core && go test ./internal/server/ -run TestHandleMCP -count=1
cd core && go test ./internal/swarm/ -run TestScoped -count=1
cd interface && npx vitest run __tests__/dashboard/MissionControlChat.test.tsx __tests__/dashboard/CouncilCallErrorCard.test.tsx __tests__/dashboard/DegradedModeBanner.test.tsx __tests__/dashboard/StatusDrawer.test.tsx __tests__/pages/AutomationsPage.test.tsx __tests__/shell/ShellLayout.test.tsx  # Gate A UI reliability baseline (38 pass on 2026-02-26)
```

> Note: `cd core && go test ./... -count=1` currently fails due to an existing root-package conflict (`probe.go` and `probe_test.go` both declare `main`).

**Go test breakdown (V7 additions):**

| Package | Tests | Coverage |
| :--- | :--- | :--- |
| `internal/server` | 157 | Handler tests: missions, governance, templates, MCP, council, runs (list/events/chain), artifacts, proposals |
| `internal/events` | 16 | Emit (happy/nil-db/empty-run-id), GetRunTimeline (multi-row/empty/nil-db), summarizePayload |
| `internal/runs` | 19 | CreateRun, CreateChildRun, UpdateRunStatus (running/completed/failed), GetRun, ListRunsForMission, ListRecentRuns (multi-row/empty/nil-db/default-limit) |
| Other packages | ~80 | artifacts, bootstrap, catalogue, cognitive, governance, memory, overseer, scip, state, swarm, protocol |

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
| Workspace UX | Workspace Rename + Crew Launch + Memory Redesign | "Mission Control" → "Workspace" across rail/header/loading. `LaunchCrewModal` (3-step intent → proposal → confirm). `SomaOfflineGuide` (startup command, retry button). `MemoryExplorer` redesigned to 2-col (Warm+Cold primary, Hot behind Advanced Mode). `OpsOverview` dead route fix (`/missions/{id}/teams` removed). Auth fix: `interface/.env.local` + `ops/interface.py _load_env()`. |
| V7 Step 01 | Workflow-First Navigation (Team D) | Nav collapsed from 12+ routes to 5 workflow-first panels. `ZoneA_Rail` (5 items + Advanced Mode toggle). `/automations` (6 tabs + deep-link + advanced gate). `/resources` (4 tabs + deep-link). `/system` (5 tabs + advanced gate). 8 legacy routes → server-side `redirect()`. `PolicyTab` CRUD migrated from `/approvals` into `ApprovalsTab`. 56 unit tests pass. |
| V7 Team A | Event Spine | `mission_runs` (023) + `mission_events` (024) migrations. `protocol.MissionEventEnvelope` + 17 `EventType` constants. `events.Store` (Emit — DB-first + async CTS publish). `runs.Manager` (CreateRun, CreateChildRun, UpdateRunStatus). `GET /api/v1/runs/{id}/events` + `GET /api/v1/runs/{id}/chain` handlers. Propagation chain: Soma → activation → team → agent. Agent emits `tool.invoked`/`tool.completed`/`tool.failed` per ReAct iteration. `CommitResponse.RunID` returned to UI. TypeScript types in `interface/types/events.ts`. |
| V7 Soma Workflow | End-to-End Working Flow | **Backend:** `ConsultationEntry` type in `protocol.ChatResponsePayload`; ReAct loop captures `consult_council` calls into `ProcessResult.Consultations`; `agentResult.Consultations` wired into `chatPayload`; `GET /api/v1/runs` global listing endpoint (`runs.Manager.ListRecentRuns`). **Store:** `MissionRun` + `MissionEvent` types; `activeRunId`, `runTimeline`, `recentRuns` state; `confirmProposal` injects `role:'system'` message with `run_id`; `fetchRunTimeline` + `fetchRecentRuns` actions. **Chat UI:** Soma-locked header (no dropdown), `DirectCouncilButton` popover, `DelegationTrace` council cards, `SomaActivityIndicator` (live `streamLogs` activity), system message bubble linking to `/runs/{id}`. **Runs UI:** `RunTimeline.tsx` (auto-poll 5s), `EventCard.tsx`, `/runs/[id]` page, `/runs` list page, `RecentRunsSection` in OpsOverview. **OpsWidget Registry:** `lib/opsWidgetRegistry.ts` — `registerOpsWidget()` / `getOpsWidgets()` / `unregisterOpsWidget()` plugin API; OpsOverview renders from registry. **LaunchCrewModal:** Always targets Soma on open; clears stale proposals. **Tests:** 7 new passing Go tests (4 `ListRecentRuns`, 3 `handleListRuns`). |
| Provider CRUD + Profiles | Provider Management + Mission Profiles + Reactive | **Backend:** `AddProvider`/`UpdateProvider`/`RemoveProvider` with `RWMutex` hot-reload on `cognitive.Router`. `POST/PUT/DELETE /api/v1/brains` + `POST /api/v1/brains/{id}/probe`. Context snapshot CRUD (`context_snapshots` migration 028). Mission profile CRUD + activate (`mission_profiles` migration 029). Reactive NATS subscription engine (`core/internal/reactive/engine.go`). `GET /api/v1/services/status` health aggregation. `MaxReconnects(-1)` + DB/NATS startup retry loops (45×2s). **Frontend:** `BrainsPage.tsx` (add/edit/delete/probe with type presets), `ContextSwitchModal.tsx` (Cache & Transfer / Start Fresh / Load Snapshot), `MissionProfilesPage.tsx` (role→provider table, NATS subscriptions, context strategy), Profiles tab in Settings, Services tab in System. |
| V7 Team B | Trigger Engine | **Migrations:** `trigger_rules` (025) + `trigger_executions` (026). **Backend:** `triggers.Store` (rule CRUD, in-memory cache, `LogExecution`, `ActiveCount`). `triggers.Engine` (CTS subscription on `swarm.mission.events.*`, 4-guard `evaluateRule` — cooldown, recursion depth, concurrency, condition — `fireTrigger` creates child run, `proposeTrigger` logs for approval). 6 HTTP handlers (`GET/POST/PUT/DELETE /api/v1/triggers`, `POST /toggle`, `GET /history`). Wired into `AdminServer` + `main.go` with graceful shutdown. **Frontend:** `TriggerRulesTab.tsx` (full CRUD UI — RuleCard, CreateRuleForm, guard badges, mode warnings). Trigger types + 5 async actions in `useCortexStore`. Automations → Triggers tab now live (was DegradedState). **Bug fixes:** `/runs/[id]/page.tsx` (`"use client"` + `use(params)` for Next.js 15+), `/docs/page.tsx` (Suspense boundary for `useSearchParams`). |
| V7 Conversation Log | Agent Transcript Browsing + Interjection | **Migration 030** (`conversation_turns`). **Backend:** `conversations.Store` (LogTurn, GetRunConversation, GetSessionTurns). `ConversationLogger` interface in `protocol/events.go` propagated Soma → Team → Agent (mirrors EventEmitter). 6 emission points in `processMessageStructured()` (system, user, tool_call, tool_result, interjection, assistant). Interjection via NATS mailbox `swarm.agent.{id}.interjection` — agent checks between ReAct iterations. 3 HTTP handlers (run conversation, session turns, interject). **Frontend:** `ConversationLog.tsx` (agent filter, 5s auto-poll, interjection input), `TurnCard.tsx` (role-based colors/icons/badges), `types/conversations.ts`. `/runs/[id]` tab bar (Conversation + Events). **Tests:** 13 Go store tests, 11 Go handler tests, 9 frontend tests. |
| V7 Inception Recipes | Structured Prompt Patterns for RAG | **Migration 031** (`inception_recipes`). **Backend:** `inception.Store` (CreateRecipe, GetRecipe, ListRecipes, SearchByTitle, IncrementUsage, UpdateQuality). `store_inception_recipe` + `recall_inception_recipes` internal tools (dual-persist: RDBMS + pgvector). Recipe recall integrated into `research_for_blueprint` pipeline (step 5). Interaction protocol updated: agents prompted to store recipes after complex tasks. 5 HTTP handlers (list, search, get, create, quality feedback). **Tests:** 16 Go store tests, 10 Go handler tests. |
| MCP Test Hardening | Service + Handler Coverage | **Backend Tests:** new suites for MCP library loading/config conversion, registry service CRUD/cache/find flows, toolset CRUD/ref resolution, and executor adapter result formatting. **Handler Tests:** DB-backed happy paths for MCP list/delete/tools/library-install plus toolset update matrix (happy/not-found/bad UUID/missing name/nil service). **Semantics:** `handleUpdateToolSet` returns `404` when the tool set does not exist. |
| V7 UI Gate A | Parallel UI Framework + Reliability Baseline | **Docs:** `docs/UI_FRAMEWORK_V7.md` + `docs/ui-delivery/*` lane playbooks and board. **UX:** `StatusDrawer`, `DegradedModeBanner`, `CouncilCallErrorCard`, `FocusModeToggle`, `SystemQuickChecks`, and `AutomationHub` baseline integrated. **Tests:** Added/updated Vitest suites for dashboard/pages/shell with targeted Gate A run passing (`38` tests). |
| In-App Docs Browser | `/docs` + Doc Registry | **Next.js Route Handlers:** `GET /docs-api` (manifest) + `GET /docs-api/[slug]` (file content, path-validated against manifest). `/docs-api` prefix avoids the `/api/*` → Go backend proxy rewrite; `params` awaited for Next.js 15+ async param requirement. **Manifest:** `lib/docsManifest.ts` — 29 entries across 7 curated sections; `DOC_BY_SLUG` flat map for O(1) slug validation; add a doc by adding one `DocEntry`. **User Guides (new):** 7 plain-language guides in `docs/user/` — Core Concepts, Using Soma Chat, Run Timeline, Automations, Resources, Memory, Governance & Trust — covering every implemented workflow and concept. **UI:** `/docs` page — two-column layout: sidebar (grouped nav, filter search, active state) + content pane (react-markdown + remark-gfm, Midnight Cortex styled). `?doc={slug}` deep-link; URL synced on every sidebar click. **Nav:** `BookOpen` Docs link in main nav directly below Memory (not in footer). |

> Full phase history with details: [Architecture Overview](docs/architecture/OVERVIEW.md#vi-delivered-phases)

## Upcoming Architecture

Planned phases with detailed specifications are documented in the Architecture Overview:

| Phase | Name | Summary |
| :--- | :--- | :--- |
| **V7** | **Event Spine & Workflow-First Orchestration** | **IN PROGRESS** — Team D (nav) ✓, Workspace UX ✓, Team A (Event Spine) ✓, Soma Workflow E2E ✓, Provider CRUD + Mission Profiles ✓, Team B (Trigger Engine) ✓, Conversation Log + Interjection ✓ (migration 030, full transcript persistence, operator redirect), Inception Recipes ✓ (migration 031, dual-persist RAG patterns, quality feedback). **Next:** Team C (Scheduler: migration 027, cron goroutine, NATS suspend/resume) → Causal Chain UI (`ViewChain.tsx`) |
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
