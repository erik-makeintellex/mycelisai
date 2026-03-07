# Mycelis Cortex V7.0

**The Recursive Swarm Operating System.**

> [!IMPORTANT]
> **MASTER STATE AUTHORITY**
> This README is the **Single Source of Truth** for project state, architecture, and operational commands.
>
> **Architecture PRD:** The detailed architecture specification lives in core documents:
> | Document | Load When |
> | :--- | :--- |
> | [Architecture Library Index](docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md) | Start here for canonical planning, target delivery, and architecture navigation |
> | [Target Deliverable V7](docs/architecture-library/TARGET_DELIVERABLE_V7.md) | Product end state, recurring-plan modes, success criteria, and phase framing |
> | [System Architecture V7](docs/architecture-library/SYSTEM_ARCHITECTURE_V7.md) | Runtime layers, persistence, NATS posture, deployment, and storage model |
> | [Execution And Manifest Library V7](docs/architecture-library/EXECUTION_AND_MANIFEST_LIBRARY_V7.md) | Runs, manifests, scheduled/event-driven/persistent-active behavior |
> | [UI And Operator Experience V7](docs/architecture-library/UI_AND_OPERATOR_EXPERIENCE_V7.md) | Canonical UI target, anti-information-swarm rules, and operator journeys |
> | [Delivery Governance And Testing V7](docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md) | Delivery proof, acceptance gates, and product-aligned testing |
> | [Operations Manual](docs/architecture/OPERATIONS.md) | Deploying, testing, CI/CD, config |
> | [V7 PRD Index](mycelis-architecture-v7.md) | Stable compatibility entrypoint that points to the modular architecture library |
> | [Architecture Overview](docs/architecture/OVERVIEW.md) | Supporting architecture summary and specialized phase context |
> | [Backend Specification](docs/architecture/BACKEND.md) | Working on Go code, APIs, DB, NATS |
> | [Frontend Specification](docs/architecture/FRONTEND.md) | Working on React/Next.js, Zustand, design |
> | [NATS Signal Standard V7](docs/architecture/NATS_SIGNAL_STANDARD_V7.md) | Canonical subject families, source normalization, and product-vs-dev channel rules |
> | [Archive Index](docs/archive/README.md) | Historical docs only (non-authoritative) |

## README TOC

- [Fresh Agent Start Here](#fresh-agent-start-here)
- [Feature Status Standard](#feature-status-standard)
- [Architecture](#architecture)
  - [Workspace Reference](#workspace-reference)
- [Soma Workflow - End-to-End Reference](#soma-workflow-end-to-end-reference)
  - [Quick Reference](#quick-reference)
  - [GUI Execution Path](#gui-execution-path)
  - [API Execution Path](#api-execution-path)
  - [Trigger Rules - Automation Workflow](#trigger-rules-automation-workflow)
  - [Conversation Log And Interjection - Agent Transcript Browsing](#conversation-log-interjection-agent-transcript-browsing)
  - [Inception Recipes - Structured Prompt Patterns](#inception-recipes-structured-prompt-patterns)
  - [Workflow State Diagram](#workflow-state-diagram)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Configure Environment](#1-configure-environment)
  - [Bring Up Cluster](#2-bring-up-cluster-canonical-order)
  - [Start Full Local Stack](#3-start-full-local-stack-recommended)
  - [Manual Start](#4-manual-start-alternative)
  - [Stop / Restart](#5-stop-restart)
  - [Fresh Deployment Reset](#5a-fresh-deployment-reset-clean-slate)
  - [Fresh Memory Restart](#5b-fresh-memory-restart-db-memory-probes)
  - [Configure Cognitive Providers](#6-configure-cognitive-providers)
  - [Rename Soma](#7-rename-soma-assistant-display-name)
- [Developer Orchestration](#developer-orchestration)
- [Frontend Routes](#frontend-routes)
- [Stack Versions](#stack-versions-locked)
- [Key Configurations](#key-configurations)
- [Documentation Hub](#documentation-hub)
- [Delivery Discipline](#delivery-discipline-required)
- [Verification](#verification)
- [Delivered Phases](#delivered-phases)
- [Upcoming Architecture](#upcoming-architecture)
- [Branching Strategy](#branching-strategy)

## Fresh Agent Start Here

If you are a fresh development agent or starting a new interaction, review the docs in this order before making changes:

1. [AGENTS.md](AGENTS.md) for repo standards, language ownership, runner contract, and canonical docs location.
2. [Architecture Library Index](docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md) for the canonical planning map.
3. [Target Deliverable V7](docs/architecture-library/TARGET_DELIVERABLE_V7.md) for the intended product end state and phase framing.
4. [Execution And Manifest Library V7](docs/architecture-library/EXECUTION_AND_MANIFEST_LIBRARY_V7.md) for run, manifest, recurring-plan, and activation semantics.
5. [UI And Operator Experience V7](docs/architecture-library/UI_AND_OPERATOR_EXPERIENCE_V7.md) and [UI Target + Transaction Contract V7](docs/architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md) before touching operator-facing UI.
6. [Operations Manual](docs/architecture/OPERATIONS.md) and [Testing Guide](docs/TESTING.md) before changing runtime tasks, lifecycle flows, or delivery gates.
7. [V7 Dev State](V7_DEV_STATE.md) for the current checkpoint and active delivery context.

Use [mycelis-architecture-v7.md](mycelis-architecture-v7.md) only as the stable PRD index and compatibility entrypoint. Do not expand it back into the primary detailed spec.

## Feature Status Standard

Use these canonical delivery markers everywhere feature status is tracked in current docs:

| Marker | Meaning |
| :--- | :--- |
| `REQUIRED` | Must exist for target delivery or gate pass, but not yet started or not yet ready |
| `NEXT` | Highest-priority upcoming slice |
| `ACTIVE` | Currently in development |
| `IN_REVIEW` | Implemented and awaiting validation, review, or gate decision |
| `COMPLETE` | Delivered and accepted |
| `BLOCKED` | Cannot advance until a dependency or defect is resolved |

Primary places to apply and review these markers:
- [V7_DEV_STATE.md](V7_DEV_STATE.md) for current development state
- [docs/architecture-library/TARGET_DELIVERABLE_V7.md](docs/architecture-library/TARGET_DELIVERABLE_V7.md) for target-phase framing
- [docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md](docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md) for gate and review expectations

Mycelis is a governed orchestration system ("Neural Organism") where users express intent, Mycelis proposes structured plans, and any state mutation requires explicit confirmation plus a complete Intent Proof bundle. Missions are not isolated — they emit structured events that trigger other missions. Observability is not optional: execution must never be a black box.

Built through 19 phases — from genesis through **Admin Orchestrator**, **Council Activation**, **Trust Economy**, **RAG Persistence**, **Agent Visualization**, **Neural Wiring Edit/Delete**, **Meta-Agent Research**, **Team Management**, **Soma Identity & Artifacts**, **Conversation Memory**, **Natural Human Interface**, **Phase 0 Security Containment**, **Agent & Provider Orchestration** — and now executing **V7: Event Spine & Workflow-First Orchestration**. V7 Team A (Event Spine), Team B (Trigger Engine), Agent Conversation Log, and Inception Recipes are complete: persistent mission runs, `MissionEventEnvelope` audit records, tool event emission, run timeline APIs, declarative trigger rules with cooldown/recursion/concurrency guards, full-fidelity agent conversation transcripts with operator interjection, and structured inception recipe patterns for RAG-based knowledge reuse. Team C (Scheduler) and Causal Chain UI follow next.

## Architecture

### Tier 1: Core (Go 1.26 + Postgres + pgvector)

- **Soma → Axon → Teams → Agents:** Mission activation pipeline with heartbeat + proof-of-work.
- **Parallel Team Activation:** `ActivateBlueprint` starts eligible teams concurrently (bounded worker pool), then inserts them race-safely into Soma's team map. Duplicate IDs are skipped idempotently even under concurrent activation calls.
- **Standing Teams:** Admin/Soma and Council remain the core standing runtime, with manifest-backed specialist teams such as `prime-architect`, `prime-development`, and `agui-design-architect` available for architecture delivery and workflow-composer coordination.
- **Team Signal Standard (Runtime):** team-facing ingress uses `swarm.team.{team_id}.internal.command`; operator-facing outputs use `swarm.team.{team_id}.signal.status` or `swarm.team.{team_id}.signal.result` with normalized signal metadata.
- **UI Signal Normalization:** frontend stream consumers normalize legacy and standardized signal envelopes through shared helpers before dashboard, drawer, and status surfaces render them.
- **Council Chat API:** Standardized CTS-enveloped responses with trust scores, provenance metadata, and tools-used tracking. Dynamic member validation via Soma — add a YAML, restart, done.
- **Runtime Context Injection:** Every agent receives live system state (active teams, NATS topology, MCP servers, cognitive config, interaction protocols) via `InternalToolRegistry.BuildContext()`.
- **Internal Tool Registry:** 23 built-in tools — consult_council, delegate_task, search_memory, remember, recall, broadcast, file I/O (workspace-sandboxed), NATS bus sensing, image generation + cache-save flow, summarize_conversation, research_for_blueprint, store_inception_recipe, recall_inception_recipes, and more.
- **Composite Tool Executor:** Unified interface routing tool calls to InternalToolRegistry or MCP ToolExecutorAdapter.
- **MCP Ingress:** Install, manage, and invoke MCP tool servers. Curated library with one-click install. Raw install endpoint disabled (Phase 0 security) — library-only installs enforced.
- **MCP Test Coverage:** Service/toolset/library/executor suites plus DB-backed MCP handler tests; toolset update not-found now returns HTTP 404.
- **Archivist:** Context engine — SitReps, auto-embed to pgvector (768-dim, nomic-embed-text), semantic search.
- **Governance:** Policy engine with YAML rules, approval queue, trust economy (0.0–1.0 threshold).
- **Cognitive Router:** 6 LLM providers (ollama, vllm, lmstudio, OpenAI, Anthropic, Gemini), profile-based routing, token telemetry. Brain provenance tracks which provider/model executed each response.
- **Startup Provider Scope:** On boot, connectivity probing is scoped to default `ollama` plus providers explicitly routed by profiles. Mycelis does not attempt to connect to every declared backend unless it is actively configured for use.
- **CE-1 Templates:** Orchestration template engine with intent proofs, confirm tokens (15min TTL), and audit trail. Chat-to-Answer (read-only) and Chat-to-Proposal (mutation-gated) execution modes.
- **Brains API:** Full provider CRUD — add, edit, delete, and probe providers at runtime with zero restart (`AddProvider`/`UpdateProvider`/`RemoveProvider` with `RWMutex` hot-reload). Type presets for Ollama, vLLM, LM Studio, OpenAI, Anthropic, Gemini, Custom. Location/data_boundary/usage_policy/roles_allowed enforced. All mutations persist to `cognitive.yaml`.
- **Mission Profiles:** Named workflow configurations that map agent roles to specific providers. Activate a profile to instantly reroute `architect → vllm`, `coder → ollama`, etc. Context Switch strategies: Cache & Transfer (auto-snapshot before switch), Start Fresh, Load Snapshot. Profiles persist to `mission_profiles` table (migration 029).
- **Root-admin Collaboration Groups (V7):** DB-backed goal-scoped groups with tenant scoping, policy refs, audit linkage, scoped permissions (`groups:read|write|broadcast`), high-impact confirm-token gating, NATS fanout, and live group-bus monitor surfaces.
- **Root-admin Full Configuration Authority (V7):** Root-admin can direct Soma to execute configuration across the whole platform (providers/profiles, governance policy, MCP/toolsets, users/groups, runtime settings), not only team instantiation, while preserving proposal/approval gates for governed mutations.
- **Local Command V0 (V7):** Root-admin host actions API (`/api/v1/host/actions`) with allowlisted no-shell command execution (`MYCELIS_LOCAL_COMMAND_ALLOWLIST`) and bounded timeout/args validation.
- **Coder-First Web Access Rule (V7):** Web search/site retrieval defaults to development-specialist ephemeral code execution (adaptive search-engine/query strategy), with MCP used when clearly easier/required.
- **MCP Intent Translation Contract (V7):** Soma/Council must translate user intent into concrete currently-installed MCP tool calls (or emit explicit missing dependency/credential requirements) instead of schema-only responses.
- **Manifest Profile Improvement (Planned V7.x):** Each active Soma profile carries a governed `manifest_improvement` policy so self-improvement is tied to user-invoked objective classes, with review gates, drift guard, and rollback history.
- **Context Snapshots:** Point-in-time saves of conversation state (messages, run state, active role providers). Created automatically on Cache & Transfer activation or manually. Stored in `context_snapshots` table (migration 028). Last 20 shown in Load Snapshot picker.
- **Reactive Subscription Engine:** Mission profiles can subscribe to NATS topic patterns (e.g. `swarm.team.research-team.*`). On message receipt, Soma evaluates whether to engage — reactive-watch, not forced auto-chain. Subscriptions survive NATS reconnects via `ReactivateFromDB`. Engine in `core/internal/reactive/engine.go`.
- **Self-Healing Connectivity:** NATS client uses `MaxReconnects(-1)` (unlimited) with 20s ping interval for stale connection detection. Core startup retries DB and NATS for up to 90s (45×2s) before failing. Reactive engine re-subscribes all active profiles on NATS reconnect.
- **Event Spine (V7):** Dual-layer event architecture — CTS for real-time signal transport, MissionEventEnvelope for persistent audit-grade records. Every execution creates a `mission_run` with unique `run_id`. Events persisted to `mission_events` before CTS publish. CTS payloads reference `mission_event_id` for timeline reconstruction.
- **Trigger Rules Engine (V7):** Declarative IF/THEN trigger rules evaluated on event ingest. Supports cooldown, recursion guard (max depth), and concurrency guard. Default mode: propose-only. Auto-execute requires explicit policy allowance.
- **Conversation Log (V7):** Full-fidelity agent conversation persistence — every system prompt, user message, tool call, tool result, and assistant response stored in `conversation_turns` table (migration 030). Separate from lightweight `mission_events` — turns are full-text blobs (10KB+). Session-scoped with `session_id`; run-linked when available. `ConversationLogger` interface propagated Soma → Team → Agent, matching `EventEmitter` pattern.
- **Soma Direct-Draft Guard:** Plain chat drafting requests (letters, emails, short written content) are answered directly in-chat. Soma should not invoke file/system/council tools for these unless the user explicitly asks to save, inspect, execute, or delegate.
- **Deterministic Local Teardown:** `uv run inv lifecycle.down` uses bounded cleanup timeouts and waits for Core/Frontend ports to close so local test runs do not hang on stale shutdown processes.
- **Operator Interjection (V7):** Mid-run user redirection via NATS mailbox (`swarm.agent.{id}.interjection`). Agents check a mutex-protected buffer between ReAct iterations. Interjections are injected as `[OPERATOR INTERJECTION]` messages and logged as `role=interjection` turns. Only applies to in-progress runs with active agents.
- **Inception Recipes (V7):** Structured prompt pattern library for knowledge reuse. When Soma completes a complex task, it can distill an inception recipe capturing: intent pattern, key parameters, example prompt, and outcome shape. Recipes are dual-persisted — RDBMS (`inception_recipes` table, migration 031) for structured queries + pgvector (`context_vectors`) for semantic recall. Quality feedback loop (`quality_score` 0.0–1.0) + usage tracking. Automatically recalled during `research_for_blueprint` pipeline.
- **Permanent Memory Growth Loop (Planned V7.x):** Postgres + pgvector loop (`capture -> distill -> vectorize -> retrieve -> promote -> rollback`) with a strict no-silence preference rule and 3-layer local continuity memory (`hot/context/archive`) synced back to system-of-record.
- **Central Agentry Memory Governance (Planned V7.x):** Correction repetition tracking with 3-strike scope prompt (`global|domain|project`), weekly maintenance digest, provenance on learned-rule application (`file` + `line`), sensitive-data exclusion, and supported forget semantics (`forget X`, `forget everything`).
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
- **Settings (`/settings`):** Profile (assistant display-name rename), Brains (full provider CRUD — Add/Edit/Delete/Probe with type presets, remote-enable confirmation, LOCAL/LEAVES_ORG boundary badge), Profiles (mission profile CRUD + activate + Context Switch Modal), Cognitive Matrix, MCP Tools (curated library), Users & Groups (operator user controls + embedded collaboration-group management panel). Policy/approval rules in Automations → Approvals tab.
- **Run Timeline (V7):** Vertical event timeline per mission run — policy decisions, tool invocations, trigger firings, artifacts, completion. `RunTimeline.tsx` + `EventCard.tsx` + `/runs/[id]` page. Auto-polls every 5s; stops on terminal events. Tab bar switches between Conversation and Events views.
- **Conversation Log (V7):** Full agent transcript viewer per run — `ConversationLog.tsx` + `TurnCard.tsx`. Role-based coloring (system gray, user cyan, assistant green, tool_call violet, tool_result amber, interjection red). Agent filter bar for multi-agent runs. System prompts collapsed by default. Provider/model badges on assistant turns. Tool name badges on tool_call turns. Auto-polls 5s while run is active. Operator interjection input bar visible when run status is `running`.
- **Run List (V7):** `/runs` page listing all recent runs across missions, with status dots and timestamps. Also surfaced in OpsOverview as a `Recent Runs` widget.
- **Causal Chain View (V7, backend ready):** Parent run → event → trigger → child run traversal. `GET /api/v1/runs/{id}/chain` handler complete; UI pending.
- **Mode Ribbon:** Always-visible status bar showing current execution mode, active brain (with local/remote badge), and governance state.
- **Operational Reliability UX (V7 Gate A):** Global `DegradedModeBanner`, global `StatusDrawer` (opened from Mode Ribbon or floating status action), structured `CouncilCallErrorCard` with retry/reroute/copy diagnostics actions, Workspace Focus Mode (`F` key), and `SystemQuickChecks` panel on `/system`.
- **Standardized Resource API Contract:** Resource surfaces normalize both raw and enveloped payloads (`{ ok, data, error }`) through shared contract helpers/store actions, so new AI resource channels can be added without per-screen parsing drift.
- **Proposal Blocks:** Inline chat cards for mutation-gated actions — shows intent, tools, risk level, confirm/cancel buttons wired to CE-1 confirm token flow.
- **Orchestration Inspector:** Expandable audit panel showing template ID, intent proof, confirm token, and execution mode for each chat response.
- **Visual Protocol:** Midnight Cortex theme — `cortex-bg #09090b`, `cortex-primary #06b6d4` (cyan). Zero `bg-white` in new code. Base font-size 17px for rem-proportional readability across all Tailwind utility classes.

### Workspace Reference

Workspace (`/dashboard`) is the admin's primary interface — a resizable two-panel layout where 80% of work happens in conversation.

```
┌───────────────────────────────────────────────────────────────┐
│  Workspace                SIGNAL: LIVE   [Launch Crew] [⚙]   │
├─── Telemetry Row (Advanced Mode) ─────────────────────────────┤
│  [Goroutines: 42] [Heap: 18MB] [System: 52MB] [LLM: 3.2t/s] │
├───────────────────────────────────────────────────────────────┤
│                                                               │
│   ADMIN / COUNCIL CHAT  (68% — resizable)                     │
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
│   OPS OVERVIEW  (32% — collapsible)                           │
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
| **Telemetry Row** | `TelemetryRow` | 4 sparkline cards — Goroutines, Heap, System Memory, LLM Tokens/s. Polls `/api/v1/telemetry/compute` every 5s. Shows offline banner after 3 failures. Rendered only in Advanced Mode to keep workspace focused. |
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
| **Chat route contract** | Workspace chat uses `POST /api/v1/chat`; direct specialist chat uses `POST /api/v1/council/{member}/chat`. Blocked Workspace requests render a Soma-specific blocker card rather than generic council-failure text. |
| **Delegation Trace** | When Soma calls `consult_council` during its ReAct loop, a `DelegationTrace` card appears below the response showing which council members were consulted and a 300-char summary of their contribution. Color-coded per member (Architect=info, Coder=success, Creative=warning, Sentry=danger). |
| **Live activity** | While Soma processes, a `SomaActivityIndicator` reads `streamLogs` for `tool.invoked` events and shows contextual text: "Consulting Coder...", "Generating blueprint...", "Searching memory..." instead of a static spinner. |
| **Broadcast mode** | Toggle or `/all` prefix — sends message to ALL active teams via NATS |
| **File I/O** | Admin + council agents can `read_file` and `write_file` within the workspace sandbox (`MYCELIS_WORKSPACE`, default `./workspace`; Kubernetes default `/data/workspace`). Paths must resolve inside the boundary — symlink escapes are detected. Max 1MB per write. Sentry is read-only. |
| **Tool access** | 21 internal tools: consult_council, delegate_task, search_memory, remember, recall, broadcast, publish_signal, read_signals, read_file, write_file, generate_image, save_cached_image, research_for_blueprint, generate_blueprint, list_teams, list_missions, get_system_status, list_available_tools, list_catalogue, store_artifact, summarize_conversation |
| **Image cache policy** | Generated images are cache-first and expire after 60 minutes unless explicitly persisted to `workspace/saved-media` via `save_cached_image` or artifact save API |
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
- Every new library/service install must pass intake inspection against Soma baseline rules, deployment defaults, and user policy overlays before enablement.

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

Kubernetes storage contract:
- the Helm chart mounts the persistent data PVC at `/data`
- manifested filesystem output uses `MYCELIS_WORKSPACE=/data/workspace`
- artifact blob/file storage uses `DATA_DIR=/data/artifacts`
- both paths are created on Core startup so requested material lands on mounted storage, not the container filesystem

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
| 1. Send message | `MissionControlChat` input | `POST /api/v1/chat` | `HandleChat` | [Chat Panel](#chat-panel--rich-message-rendering) |
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

- Store action: `sendMissionChat(text)` → `POST /api/v1/chat`
- `HandleChat` forwards the conversation to Soma over `swarm.council.admin.request`

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

#### 1. Send Message — `POST /api/v1/chat`

```http
POST /api/v1/chat
Authorization: Bearer mycelis-dev-key-change-in-prod
Content-Type: application/json

{
  "messages": [
    { "role": "user", "content": "Write me a Python CSV parser that handles quoted fields" }
  ]
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

> **Note:** Workspace / Soma chat uses `POST /api/v1/chat`. Direct specialist chat uses `POST /api/v1/council/{member}/chat`. Use `GET /api/v1/council/members` to list all available members dynamically.

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

#### Contract Baseline — `GET /api/v1/inception/contracts`

Returns the frozen Sprint-0 contract bundle used by extension-of-self runtime:
- decision frame (`direct|manifest_team|propose|scheduled_repeat`)
- heartbeat autonomy budget
- universal invoke envelope

```http
GET /api/v1/inception/contracts
Authorization: Bearer mycelis-dev-key-change-in-prod
```

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
POST /api/v1/chat
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

### Prerequisites

| Tool | Minimum | Notes |
| :--- | :--- | :--- |
| Docker Desktop | Latest | Required for Kind cluster |
| Kind | Latest | Local Kubernetes |
| kubectl | v1.35+ | Older versions have known port-forward issues |
| Helm | v3+ | Chart deployment |
| Go | 1.26 | Backend build/runtime |
| Node.js | 20+ | Frontend build/runtime |
| uv | Latest | Task runner (`uv run inv ...`) |
| Ollama | Latest | Default local model runtime |

### 1. Configure Environment

```bash
cp .env.example .env
# REQUIRED: set MYCELIS_API_KEY (server will refuse startup without it)
# Recommended helper:
#   uv run inv auth.dev-key
# Recommended defaults:
#   OLLAMA_HOST=http://127.0.0.1:11434
#   NATS_URL=nats://127.0.0.1:4222
#   DB_HOST=127.0.0.1 DB_PORT=5432
```

Optional but recommended (first-time Ollama model pull):

```bash
ollama pull qwen2.5-coder:7b
```

### 2. Bring Up Cluster (Canonical Order)

```bash
uv run inv k8s.up
```

`k8s.up` enforces dependency order in-cluster:
1. Kind cluster + namespace init
2. Helm deploy (PostgreSQL, NATS, Core API)
3. Rollout readiness gates (PostgreSQL -> NATS -> Core API)

### 3. Start Full Local Stack (Recommended)

This starts bridge + backend + frontend in one flow:

```bash
uv run inv lifecycle.up --build --frontend
```

Primary task runner contract is `uv run inv ...`.
Compatibility probe: `uvx --from invoke inv -l`.
Do not use bare `uvx inv ...`.

Language ownership is explicit:
- Go for core runtime/backend work
- TypeScript for interface/UI work
- Python for app management tasks, CI task orchestration, and test harnesses
- PowerShell only as a host wrapper when the local platform requires it

If `uv run inv ...` is unavailable in your shell, use:

```bash
.\.venv\Scripts\inv.exe lifecycle.up --build --frontend
```

Verify:

```bash
uv run inv lifecycle.status
uv run inv lifecycle.health
```

`uv run inv lifecycle.up ...` now waits for Core HTTP readiness on `/healthz`, not just an open port. If Core never becomes ready, the task fails fast instead of reporting a false-success stack state.

Open:
- `http://localhost:3000/dashboard` (frontend)
- `http://localhost:8081/healthz` (backend health)

### 4. Manual Start (Alternative)

If you prefer manual control over each service:

```bash
uv run inv k8s.up                # Cluster services ready in dependency order
uv run inv k8s.bridge            # Port-forward PG:5432, NATS:4222 (Terminal 1)
uv run inv db.migrate            # Apply all migrations (idempotent)
uv run inv core.build && uv run inv core.run   # Build + run (Terminal 2, foreground)
uv run inv interface.install     # First time only
uv run inv interface.dev         # Next.js dev server (Terminal 3)
```

### 5. Stop / Restart

```bash
uv run inv lifecycle.down
uv run inv lifecycle.restart --build --frontend
```

### 5A. Fresh Deployment Reset (Clean Slate)

Use this when you want to fully tear down local runtime state and re-bootstrap before the next architecture step.

```bash
uv run inv lifecycle.down
uv run inv k8s.reset
uv run inv lifecycle.up --build --frontend
uv run inv lifecycle.health
```

### 5B. Fresh Memory Restart (DB + Memory Probes)

Use this when memory stream/search behavior needs a deterministic reset and readiness check.

```bash
uv run inv lifecycle.memory-restart --build --frontend
```

This workflow runs:
1. `lifecycle.down`
2. `db.reset`
3. `lifecycle.up`
4. `lifecycle.health`
5. memory probes:
   - `GET /api/v1/memory/stream`
   - `GET /api/v1/memory/sitreps?limit=1`

Current guarantees:
- `db.reset` applies only the canonical forward migration path (`001_init_memory.sql` plus `*.up.sql`)
- PostgreSQL bootstrap enables both `vector` and `pg_trgm`, so trigram indexes build cleanly on fresh reset
- if background Core startup fails during `lifecycle.up` or `lifecycle.memory-restart`, inspect `workspace/logs/core-startup.log`

### 6. Configure Cognitive Providers

- **UI:** `/settings` → **Cognitive Matrix** tab — change provider routing, configure endpoints.
- **MCP:** `/settings` → **MCP Tools** tab — install servers from curated library or manually.
- **YAML:** Edit `core/config/cognitive.yaml` directly.
- **Default:** Ollama is the standard local provider unless you explicitly reroute profiles.
- **Env:** `OLLAMA_HOST` in `.env` sets the default Ollama endpoint.

### 7. Rename Soma (Assistant Display Name)

You can rename the assistant from the UI and the new name will propagate across operational surfaces.

1. Open `/settings` → `Profile`.
2. Set **Assistant Name**.
3. Click **Save**.

Updated surfaces include:
- Workspace chat header and placeholders
- Degraded banner actions
- Council error card actions
- Launch Crew copy
- Runs/Ops copy where the orchestrator name is shown

#### Multi-Host Agent Routing (Enterprise / Multi-Backend)

Mycelis supports routing different agents/teams to different AI backend hosts while preserving one shared NATS orchestration bus.

- Add providers in `cognitive.yaml` (or provider CRUD API), one provider per host.
- Route by profile (`profiles` map), team default (`team.provider`), or agent override (`agent.provider`).
- Default remains local Ollama unless overridden.

Runtime override env vars (JSON maps):

- `MYCELIS_TEAM_PROVIDER_MAP`
  - Example: `{"council-core":"ollama-local","research-team":"vllm-cluster-a"}`
- `MYCELIS_AGENT_PROVIDER_MAP`
  - Example: `{"admin":"ollama-local","council-architect":"claude-remote","council-coder":"lmstudio-local"}`

Precedence:
1. `agent.provider`
2. `team.provider`
3. role/profile routing in `cognitive.profiles`
4. fallback (`chat`/`sentry`/first available provider)

This enables Soma/Council/teams to execute across mixed backend services while still coordinating over NATS subjects.

### 7. Backend Binary Compile + Modes (No Frontend Required)

The backend binary can run standalone as the API action backend without the frontend.

Build:

```bash
cd core
go build -o bin/server ./cmd/server
```

Run API server mode (default):

```bash
MYCELIS_API_KEY=replace-me ./bin/server
# Windows PowerShell:
# $env:MYCELIS_API_KEY="replace-me"; .\bin\server.exe
```

Run action mode from the same binary (send API requests to a target Mycelis server):

```bash
./bin/server action GET /api/v1/services/status
./bin/server action POST /api/v1/council/sentry/chat '{"messages":[{"role":"user","content":"health check"}]}'
./bin/server action shell
```

Action mode defaults to `http://localhost:8081` unless overridden by config or env.

#### Action Shell Mode

`./bin/server action shell` opens an interactive REPL for direct operator communication from a shell environment.

Supported commands:

- `status` → `GET /api/v1/services/status`
- `/api/v1/...` → shorthand for `GET` on that path
- `chat <member> <message>` → `POST /api/v1/council/{member}/chat`
- `broadcast <message>` → `POST /api/v1/swarm/broadcast`
- `send <provider> <recipient> <message>` → `POST /api/v1/comms/send`
- `<METHOD> <PATH|URL> [JSON]` → raw request mode
- `use <base_url>` → switch target server without leaving shell
- `help`, `exit`

#### Action CLI Configuration (Server Binary)

Config fields:

```yaml
api_base_url: "http://localhost:8081"
api_key: "mycelis-dev-key-change-in-prod"
timeout_seconds: 20
headers:
  X-Environment: "dev"
```

Path precedence (later wins; user-home paths outrank system/project paths):
1. `/etc/mycelis/config.yaml` or `/etc/mycelis/config.yml`
2. `./mycelis.yaml`, `./mycelis.yml`, `./config/mycelis.yaml`, `./config/mycelis.yml`
3. `$XDG_CONFIG_HOME/mycelis/config.yaml` and `.yml`
4. `$HOME/.config/mycelis/config.yaml` and `.yml`
5. `$HOME/.mycelis/config.yaml` and `.yml`
6. `MYCELIS_CONFIG` explicit file path (highest precedence)

Environment overrides for action mode:
- `MYCELIS_API_URL` (overrides `api_base_url`)
- `MYCELIS_API_KEY` (overrides `api_key`)
- `MYCELIS_API_TIMEOUT_SEC` (overrides `timeout_seconds`)

#### Soma Communication Providers (Local-First, Optional)

Soma now supports outbound communication providers through `/api/v1/comms/*` and agent tooling (`send_external_message`).

Configure any subset:

- `MYCELIS_COMMS_SLACK_WEBHOOK_URL` (Slack incoming webhook)
- `MYCELIS_COMMS_TELEGRAM_BOT_TOKEN` (Telegram Bot API)
- `MYCELIS_COMMS_TWILIO_ACCOUNT_SID`
- `MYCELIS_COMMS_TWILIO_AUTH_TOKEN`
- `MYCELIS_COMMS_WHATSAPP_FROM` (e.g. `+14155238886`)
- `MYCELIS_COMMS_WEBHOOK_URL` (generic webhook sink)
- `MYCELIS_COMMS_WEBHOOK_BEARER` (optional Bearer token for generic webhook)

API endpoints:

- `GET /api/v1/comms/providers` — provider readiness/config status
- `POST /api/v1/comms/send` — direct outbound send
- `POST /api/v1/comms/inbound/{provider}` — webhook ingress to Soma input bus

## Developer Orchestration

**Prerequisites:** [uv](https://github.com/astral-sh/uv) and [Docker](https://www.docker.com/).

Run from `scratch/` root using `uv run inv`:

| Command | Description |
| :--- | :--- |
| **Core** | |
| `uv run inv core.build` | Compile Go binary + Docker image |
| `uv run inv core.test` | Run unit tests (`go test ./...`) |
| `uv run inv core.run` | Run Core locally (foreground) |
| `uv run inv auth.dev-key` | Ensure/rotate `MYCELIS_API_KEY` in `.env` and keep `.env.example` sample synced |
| `cd core && go build -o bin/server ./cmd/server` | Compile standalone backend binary (no frontend required) |
| `cd core && ./bin/server action GET /api/v1/services/status` | Use server binary as action CLI against API backend |
| `uv run inv core.stop` | Kill running Core process |
| `uv run inv core.restart` | Stop + Run |
| `uv run inv core.smoke` | Governance smoke tests |
| **Quality & Logging Gates** | |
| `uv run inv logging.check-schema` | Verify runtime event literals map to declared `EventType` constants and docs coverage |
| `uv run inv logging.check-topics` | Fail on hardcoded `swarm.*` subjects outside allowed constants file |
| `uv run inv quality.max-lines --limit 350` | Enforce max-lines rule on hot paths with temporary no-regression legacy caps |
| **Interface** | |
| `uv run inv interface.dev` | Start Next.js dev server (Turbopack) |
| `uv run inv interface.build` | Production build |
| `uv run inv interface.test` | Run Vitest unit tests |
| `uv run inv interface.e2e` | Run Playwright E2E tests (requires running servers) |
| `uv run inv interface.check` | Smoke-test running server (9 pages, no light-mode leaks) |
| `uv run inv interface.stop` | Kill dev server |
| `uv run inv interface.clean` | Clear `.next` cache |
| `uv run inv interface.restart` | Full restart: stop → clean → build → dev → check |
| **Database** | |
| `uv run inv db.migrate` | Apply canonical forward SQL migrations (`001_init_memory.sql` + `*.up.sql`) |
| `uv run inv db.reset` | Drop + recreate + migrate |
| `uv run inv db.status` | Show tables + row counts |
| **Infrastructure** | |
| `uv run inv k8s.up` | Canonical cluster bring-up: init -> deploy -> wait (PostgreSQL -> NATS -> Core API) |
| `uv run inv k8s.reset` | Full cluster reset (teardown + canonical bring-up with readiness wait) |
| `uv run inv k8s.status` | Cluster status |
| `uv run inv k8s.deploy` | Deploy Helm chart |
| `uv run inv k8s.wait` | Wait for rollout readiness gates (PostgreSQL -> NATS -> Core API) |
| `uv run inv k8s.bridge` | Port-forward NATS, API, Postgres |
| `uv run inv k8s.recover` | Restart core + infra resources (core, NATS, PostgreSQL) |
| **Cognitive** | |
| `uv run inv cognitive.up` | Start vLLM + Diffusers (full stack) |
| `uv run inv cognitive.status` | Health check providers |
| `uv run inv cognitive.stop` | Kill cognitive processes |
| **Lifecycle** | |
| `uv run inv lifecycle.status` | Dashboard: Docker, Kind, PG, NATS, Core, Frontend, Ollama (with PIDs) |
| `uv run inv lifecycle.up` | Idempotent bring-up: bridge → deps → core (background, `/healthz` gated, startup log at `workspace/logs/core-startup.log`). `--frontend` `--build` |
| `uv run inv lifecycle.down` | Clean teardown: core → frontend → port-forwards |
| `uv run inv lifecycle.health` | Deep health probe: hits API endpoints with auth |
| `uv run inv lifecycle.restart` | Full restart: down → settle → up. `--build` `--frontend` |
| `uv run inv lifecycle.memory-restart` | Fresh memory reset workflow: down -> db.reset -> up -> health -> memory probes. Verified passing on 2026-03-06. `--build` `--frontend` |
| **CI Pipeline** | |
| `uv run inv ci.check` | Full CI: lint → test → build (with timers) |
| `uv run inv ci.entrypoint-check` | Verify supported invoke runner matrix (`uv run inv` vs compatibility probe vs unsupported bare alias) |
| `uv run inv ci.baseline --e2e` | Strict delivery baseline: core tests + interface build + interface typecheck + vitest + playwright |
| `uv run inv ci.lint` | Lint gate: Go vet + Next.js lint |
| `uv run inv ci.test` | Test gate: Go unit tests + Interface tests |
| `uv run inv ci.build` | Build gate: Go binary + Next.js production build |
| `uv run inv ci.toolchain-check --strict` | Enforce locked Go toolchain policy and report node/npm versions |
| `uv run inv ci.release-preflight --e2e --strict-toolchain` | Release gate: clean tree + toolchain check + strict baseline |
| `uv run inv ci.deploy` | Full CI: lint → test → build → Docker → K8s |
| **Team Coordination** | |
| `uv run inv team.architecture-sync` | Prime-architect coordination sweep over canonical team NATS lanes |

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
| Go | 1.26 | Module: `github.com/mycelis/core` (local baseline on 2026-03-03 ran on `go1.25.6`; upgrade pending) |
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
| Standing Teams | `core/config/teams/*.yaml` | YAML (auto-loaded at startup, including council and manifest-backed specialist teams) |
| MCP Servers | Database | UI (`/settings` → MCP Tools) or API |
| Governance Policy | `core/config/policy.yaml` | UI (`/approvals` → Policy tab) or YAML |
| MCP Library | `core/config/mcp-library.yaml` | YAML (curated registry) |

### Environment Variables

| Variable | Default | Description |
| :--- | :--- | :--- |
| `MYCELIS_API_KEY` | *(required)* | API authentication key. Server refuses to start without this. |
| `MYCELIS_WORKSPACE` | `./workspace` | Workspace sandbox root for agent file tools (`read_file`/`write_file`). Kubernetes default: `/data/workspace` on the mounted PVC. |
| `DATA_DIR` | `./workspace/artifacts` (local) | Artifact storage root for file-backed outputs. Kubernetes default: `/data/artifacts` on the mounted PVC. |
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
| **Fresh Agent Review Order** | [README.md#fresh-agent-start-here](README.md#fresh-agent-start-here) — Canonical onboarding sequence for new development agents and new interaction starts | — |
| **Feature Status Standard** | [README.md#feature-status-standard](README.md#feature-status-standard) — Canonical markers for required, next, active, review, complete, and blocked work | — |
| **Local Dev Workflow** | [docs/LOCAL_DEV_WORKFLOW.md](docs/LOCAL_DEV_WORKFLOW.md) — Setup, config reference, port map, troubleshooting | [/docs?doc=local-dev](/docs?doc=local-dev) |
| **Soma Workflow** | [docs/WORKFLOWS.md](docs/WORKFLOWS.md) — End-to-end GUI + API workflow reference | [/docs?doc=workflows](/docs?doc=workflows) |
| **Archive Index** | [docs/archive/README.md](docs/archive/README.md) — Historical docs only; not implementation authority | [/docs?doc=archive-index](/docs?doc=archive-index) |
| **.build Scratch Docs (Local)** | `.build/*.md` — local planning/scratch artifacts, gitignored, non-authoritative | — |
| **Council Chat QA** | [docs/QA_COUNCIL_CHAT_API.md](docs/QA_COUNCIL_CHAT_API.md) — QA procedures and test cases for council chat | [/docs?doc=council-chat-qa](/docs?doc=council-chat-qa) |
| **API Reference** | [docs/API_REFERENCE.md](docs/API_REFERENCE.md) — Full endpoint table (80+ routes) | [/docs?doc=api-reference](/docs?doc=api-reference) |
| **Architecture Library Index** | [docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md](docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md) — Canonical modular map for target delivery, architecture, execution, UI, and testing | [/docs?doc=architecture-library-index](/docs?doc=architecture-library-index) |
| **Target Deliverable V7** | [docs/architecture-library/TARGET_DELIVERABLE_V7.md](docs/architecture-library/TARGET_DELIVERABLE_V7.md) — Product end state, recurring-plan modes, success criteria, and phase framing | [/docs?doc=target-deliverable-v7](/docs?doc=target-deliverable-v7) |
| **System Architecture V7** | [docs/architecture-library/SYSTEM_ARCHITECTURE_V7.md](docs/architecture-library/SYSTEM_ARCHITECTURE_V7.md) — Runtime layers, persistence, storage, deployment, and bus posture | [/docs?doc=system-architecture-v7](/docs?doc=system-architecture-v7) |
| **Execution And Manifest Library V7** | [docs/architecture-library/EXECUTION_AND_MANIFEST_LIBRARY_V7.md](docs/architecture-library/EXECUTION_AND_MANIFEST_LIBRARY_V7.md) — Manifest lifecycle, run lifecycle, recurring plans, and activation rules | [/docs?doc=execution-manifest-library-v7](/docs?doc=execution-manifest-library-v7) |
| **UI And Operator Experience V7** | [docs/architecture-library/UI_AND_OPERATOR_EXPERIENCE_V7.md](docs/architecture-library/UI_AND_OPERATOR_EXPERIENCE_V7.md) — Anti-information-swarm UI guidance and canonical operator journeys | [/docs?doc=ui-operator-experience-v7](/docs?doc=ui-operator-experience-v7) |
| **Delivery Governance And Testing V7** | [docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md](docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md) — Delivery proof model, evidence requirements, and product-aligned testing | [/docs?doc=delivery-governance-testing-v7](/docs?doc=delivery-governance-testing-v7) |
| **Architecture Overview** | [docs/architecture/OVERVIEW.md](docs/architecture/OVERVIEW.md) — Philosophy, 4-layer anatomy, phases, upcoming roadmap | [/docs?doc=arch-overview](/docs?doc=arch-overview) |
| **Backend Specification** | [docs/architecture/BACKEND.md](docs/architecture/BACKEND.md) — Go packages, APIs, DB schema, NATS, execution pipelines | [/docs?doc=arch-backend](/docs?doc=arch-backend) |
| **Frontend Specification** | [docs/architecture/FRONTEND.md](docs/architecture/FRONTEND.md) — Routes, components, Zustand, design system | [/docs?doc=arch-frontend](/docs?doc=arch-frontend) |
| **Operations Manual** | [docs/architecture/OPERATIONS.md](docs/architecture/OPERATIONS.md) — Deployment, config, testing, CI/CD | [/docs?doc=arch-operations](/docs?doc=arch-operations) |
| **Memory Service** | [docs/architecture/DIRECTIVE_MEMORY_SERVICE.md](docs/architecture/DIRECTIVE_MEMORY_SERVICE.md) — State Engine, event projection, pgvector schema | [/docs?doc=arch-memory-service](/docs?doc=arch-memory-service) |
| **V7 Architecture PRD Index** | [mycelis-architecture-v7.md](mycelis-architecture-v7.md) — Stable root PRD path that points to the modular architecture library | [/docs?doc=v7-architecture-prd](/docs?doc=v7-architecture-prd) |
| **V7 UI Framework** | [docs/UI_FRAMEWORK_V7.md](docs/UI_FRAMEWORK_V7.md) — Default UI instantiation contract (state model, failure templates, testing matrix, PR gate) | [/docs?doc=v7-ui-framework](/docs?doc=v7-ui-framework) |
| **V7 UI Elements Planning** | [docs/UI_ELEMENTS_PLANNING_V7.md](docs/UI_ELEMENTS_PLANNING_V7.md) — Research-backed element standards, Soma interaction patterns, and planning workflow | [/docs?doc=v7-ui-elements-planning](/docs?doc=v7-ui-elements-planning) |
| **V7 Parallel Implementation Board** | [docs/ui-delivery/PARALLEL_DELIVERY_BOARD.md](docs/ui-delivery/PARALLEL_DELIVERY_BOARD.md) — Unified gate model + lane matrix (A/B/C/D/Q) with inception-control evidence tracking | [/docs?doc=v7-ui-parallel-delivery](/docs?doc=v7-ui-parallel-delivery) |
| **Team A/B/C/Q Execution Board** | [docs/ui-delivery/TEAM_ABCQ_EXECUTION_BOARD.md](docs/ui-delivery/TEAM_ABCQ_EXECUTION_BOARD.md) — Active groups hardening sprint board with explicit parallelization and acceptance gates | [/docs?doc=team-abcq-execution-board](/docs?doc=team-abcq-execution-board) |
| **V7 UI Instantiation + Bus Plan** | [docs/product/UI_WORKFLOW_INSTANTIATION_AND_BUS_PLAN_V7.md](docs/product/UI_WORKFLOW_INSTANTIATION_AND_BUS_PLAN_V7.md) — Execution-grade plan for team lifecycle, channel I/O contracts, and Basic/Guided/Expert NATS UX | [/docs?doc=v7-ui-instantiation-bus-plan](/docs?doc=v7-ui-instantiation-bus-plan) |
| **V7 MCP Baseline** | [docs/V7_MCP_BASELINE.md](docs/V7_MCP_BASELINE.md) — MVOS: filesystem, memory, artifact-renderer, fetch | [/docs?doc=v7-mcp-baseline](/docs?doc=v7-mcp-baseline) |
| **MCP Service Config (Local-First)** | [docs/architecture/MCP_SERVICE_CONFIGURATION_LOCAL_FIRST_V7.md](docs/architecture/MCP_SERVICE_CONFIGURATION_LOCAL_FIRST_V7.md) — Service onboarding standard, local-default posture, remote exception workflow | [/docs?doc=arch-mcp-service-config-local-first](/docs?doc=arch-mcp-service-config-local-first) |
| **Universal Action Interface V7** | [docs/architecture/UNIVERSAL_ACTION_INTERFACE_V7.md](docs/architecture/UNIVERSAL_ACTION_INTERFACE_V7.md) — Unified action contracts and dynamic service APIs across MCP/OpenAPI/Python | [/docs?doc=arch-universal-action-interface-v7](/docs?doc=arch-universal-action-interface-v7) |
| **Agentry Template Marketplace + Custom Templating** | [docs/architecture/AGENTRY_TEMPLATE_MARKETPLACE_AND_CUSTOM_TEMPLATING_V7.md](docs/architecture/AGENTRY_TEMPLATE_MARKETPLACE_AND_CUSTOM_TEMPLATING_V7.md) — Marketplace/source APIs, purchase/license governance, and tenant custom template lifecycle | [/docs?doc=arch-agentry-template-marketplace-v7](/docs?doc=arch-agentry-template-marketplace-v7) |
| **Actualization Beyond MCP V7** | [docs/architecture/ACTUALIZATION_ARCHITECTURE_BEYOND_MCP_V7.md](docs/architecture/ACTUALIZATION_ARCHITECTURE_BEYOND_MCP_V7.md) — External research-informed architecture for MCP/OpenAPI/A2A/ACP and Python management interfaces | [/docs?doc=arch-actualization-beyond-mcp-v7](/docs?doc=arch-actualization-beyond-mcp-v7) |
| **Secure Gateway + Remote Actuation** | [docs/architecture/SECURE_GATEWAY_REMOTE_ACTUATION_PROFILE_V7.md](docs/architecture/SECURE_GATEWAY_REMOTE_ACTUATION_PROFILE_V7.md) — Security hardening profile for self-hosted control planes and remote actuator execution | [/docs?doc=arch-secure-gateway-remote-actuation-v7](/docs?doc=arch-secure-gateway-remote-actuation-v7) |
| **Hardware Interface API + Channels** | [docs/architecture/HARDWARE_INTERFACE_API_AND_CHANNELS_V7.md](docs/architecture/HARDWARE_INTERFACE_API_AND_CHANNELS_V7.md) — Hardware control-plane API and direct channel architecture for common protocols | [/docs?doc=arch-hardware-interface-api-v7](/docs?doc=arch-hardware-interface-api-v7) |
| **Soma Symbiote + Host Actuation** | [docs/architecture/SOMA_SYMBIOTE_GROWTH_AND_HOST_ACTUATION_V7.md](docs/architecture/SOMA_SYMBIOTE_GROWTH_AND_HOST_ACTUATION_V7.md) — Thought profile contracts, growth-loop learning, and localhost host-actuation architecture | [/docs?doc=arch-soma-symbiote-growth-host-actuation-v7](/docs?doc=arch-soma-symbiote-growth-host-actuation-v7) |
| **Agent Source Instantiation Template** | [docs/architecture/AGENT_SOURCE_INSTANTIATION_TEMPLATE_V7.md](docs/architecture/AGENT_SOURCE_INSTANTIATION_TEMPLATE_V7.md) — Canonical provider template with Ollama as default source plus ChatGPT/OpenAI, Claude, Gemini, vLLM, and LM Studio | [/docs?doc=arch-agent-source-instantiation-template-v7](/docs?doc=arch-agent-source-instantiation-template-v7) |
| **Soma Extension-of-Self PRD** | [docs/product/SOMA_EXTENSION_OF_SELF_PRD_V7.md](docs/product/SOMA_EXTENSION_OF_SELF_PRD_V7.md) — Detailed extension-of-self program plus immediate delivery manifest sequencing (P0 logging, P1 cleanup, P2 meta-agent manifests, P3 workflow composer, P4 release gates) | [/docs?doc=v7-soma-extension-self-prd](/docs?doc=v7-soma-extension-self-prd) |
| **Soma Team + Channel Architecture** | [docs/architecture/SOMA_TEAM_CHANNEL_ARCHITECTURE_V7.md](docs/architecture/SOMA_TEAM_CHANNEL_ARCHITECTURE_V7.md) — Canonical inter-team/process/MCP channels, I/O envelopes, RAG memory boundaries | [/docs?doc=arch-soma-team-channels](/docs?doc=arch-soma-team-channels) |
| **NATS Signal Standard V7** | [docs/architecture/NATS_SIGNAL_STANDARD_V7.md](docs/architecture/NATS_SIGNAL_STANDARD_V7.md) — Canonical subject families, source normalization, and product-vs-dev channel boundaries for bus traffic | [/docs?doc=arch-nats-signal-standard-v7](/docs?doc=arch-nats-signal-standard-v7) |
| **Workflow Composer Delivery Plan V7** | [docs/architecture/WORKFLOW_COMPOSER_DELIVERY_V7.md](docs/architecture/WORKFLOW_COMPOSER_DELIVERY_V7.md) — Airflow-style DAG workflow composer delivery plan, team lanes, git discipline, and invoke tooling gates | [/docs?doc=arch-workflow-composer-delivery-v7](/docs?doc=arch-workflow-composer-delivery-v7) |
| **Swarm Operations** | [docs/SWARM_OPERATIONS.md](docs/SWARM_OPERATIONS.md) — Hierarchy, blueprints, activation, teams, tools, governance | [/docs?doc=swarm-operations](/docs?doc=swarm-operations) |
| **Cognitive Architecture** | [docs/COGNITIVE_ARCHITECTURE.md](docs/COGNITIVE_ARCHITECTURE.md) — Providers, profiles, matrix UI, embedding | [/docs?doc=cognitive-architecture](/docs?doc=cognitive-architecture) |
| **Logging Standard (V7)** | [docs/logging.md](docs/logging.md) — Authoritative mission-events + memory-stream logging contract, taxonomy, onboarding checklist, and quality gates | [/docs?doc=logging-schema](/docs?doc=logging-schema) |
| **Governance** | [docs/governance.md](docs/governance.md) — Policy enforcement, approvals, security | [/docs?doc=governance](/docs?doc=governance) |
| **Testing** | [docs/TESTING.md](docs/TESTING.md) — Unit, integration, smoke protocols | [/docs?doc=testing](/docs?doc=testing) |
| **V7 UI Verification (Archive)** | [docs/archive/v7-step-01-ui.md](docs/archive/v7-step-01-ui.md) — Historical manual UI checklist for V7 Step 01 navigation | [/docs?doc=v7-ui-verification](/docs?doc=v7-ui-verification) |
| **V7 Implementation Plan** | [docs/V7_IMPLEMENTATION_PLAN.md](docs/V7_IMPLEMENTATION_PLAN.md) — Teams A/B/C/D/E technical plan | [/docs?doc=v7-implementation-plan](/docs?doc=v7-implementation-plan) |
| **V7 Dev State** | [V7_DEV_STATE.md](V7_DEV_STATE.md) — Detailed checkpoint log (evidence, risks, next actions) | [/docs?doc=v7-dev-state](/docs?doc=v7-dev-state) |
| **IA Step 01 (Archive)** | [docs/archive/ia-v7-step-01.md](docs/archive/ia-v7-step-01.md) — Historical workflow-first navigation PRD and decisions | [/docs?doc=v7-ia-step01](/docs?doc=v7-ia-step01) |
| **Registry** | [core/internal/registry/README.md](core/internal/registry/README.md) — Connector marketplace | — |
| **Core API** | [core/README.md](core/README.md) — Go service architecture | — |
| **CLI** | [cli/README.md](cli/README.md) — `myc` command-line tool | — |
| **Interface** | [interface/README.md](interface/README.md) — Next.js frontend architecture | — |

## Delivery Discipline (Required)

Every merged implementation slice must update:
1. `README.md` — operator-facing behavior/config/runtime contract changes.
2. `V7_DEV_STATE.md` — current checkpoint, verification evidence, and next-step plan.
3. `docs/V7_IMPLEMENTATION_PLAN.md` — roadmap/dependency changes.
4. `interface/lib/docsManifest.ts` — include any new authoritative doc in `/docs`.

No branch promotion without this documentation gate.

## Verification

```bash
uv run inv core.test             # Go unit tests (full core package sweep)
uv run inv interface.test        # Vitest component tests (55 files / 322 tests passing as of 2026-03-02)
uv run inv interface.e2e         # Playwright E2E specs (51 passing, 4 intentionally skipped as of 2026-03-02)
uv run inv interface.check       # HTTP smoke test against running dev server (9 pages)
uv run inv core.smoke            # Governance smoke tests
uv run inv ci.entrypoint-check   # Verify uv / uvx runner matrix
cd core && go test ./... -count=1         # Full Go validation
cd interface && npx vitest run --reporter=dot
cd interface && npx playwright test --reporter=dot
cd interface && npx playwright test e2e/specs/v7-operational-ux.spec.ts  # Gate A operational UX E2E (degraded banner, status drawer, council reroute, automations hub, quick checks, focus mode)
cd core && go test ./internal/mcp/ -count=1
cd core && go test ./internal/server/ -run TestHandleMCP -count=1
cd core && go test ./internal/swarm/ -run TestScoped -count=1
cd interface && npx vitest run __tests__/automations/TeamInstantiationWizard.test.tsx __tests__/automations/RouteTemplatePicker.test.tsx __tests__/dashboard/MissionControlChat.test.tsx __tests__/dashboard/CouncilCallErrorCard.test.tsx __tests__/dashboard/DegradedModeBanner.test.tsx __tests__/dashboard/StatusDrawer.test.tsx __tests__/pages/AutomationsPage.test.tsx __tests__/shell/ShellLayout.test.tsx __tests__/pages/SystemPage.test.tsx __tests__/teams/TeamsPage.test.tsx  # Gate A + Sprint 0 baseline (50 pass on 2026-02-27)
```

Latest full baseline sweep (2026-03-03, `feature/enterprise-multihost-soma-routing` @ `752f156`):
- `cd core && go test ./... -count=1` -> pass
- `cd interface && npm run build` -> pass
- `cd interface && npx vitest run --reporter=dot` -> pass (`55` files, `322` tests; warning-only noise)
- `cd interface && npx playwright test --reporter=dot` -> pass (`51` passed, `4` skipped)
- Worktree during sweep: dirty (`59` entries: `50` modified, `9` untracked), so this is a functional baseline, not a clean-tree release baseline.

Latest strict baseline gate (2026-03-06):
- `uv run inv ci.baseline` -> pass (logging schema + topic constants + max-lines gate + core tests + interface build/typecheck/vitest)

If Playwright reports a missing browser executable, run: `cd interface && npx playwright install chromium`.

### Remote GUI Agent Preflight (Required)

When using browser agents running outside this host, do **not** assume `http://localhost:3000` targets the same runtime. Use the host LAN URL (example: `http://192.168.50.156:3000`) and run this preflight before functional assertions:

```js
(async () => {
  const rpc = await fetch('/api/rpc/consult_council').then(r => r.status);
  const members = await fetch('/api/v1/council/members').then(r => r.status);
  const html = await fetch('/automations', { cache: 'no-store' }).then(r => r.text());
  return {
    rpcStatus: rpc,
    membersStatus: members,
    hasBaseline: html.includes('automations-hub-baseline'),
    hasWizard: html.includes('open-instantiation-wizard')
  };
})().then(console.log).catch(console.error);
```

Expected values:
- `rpcStatus: 404`
- `membersStatus: 200`
- `hasBaseline: true`
- `hasWizard: true`

If preflight fails, mark the run `INVALID_ENV` and stop. Do not file product regressions from that run.

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
| Provider CRUD + Profiles | Provider Management + Mission Profiles + Reactive | **Backend:** `AddProvider`/`UpdateProvider`/`RemoveProvider` with `RWMutex` hot-reload on `cognitive.Router`. `POST/PUT/DELETE /api/v1/brains` + `POST /api/v1/brains/{id}/probe`. Context snapshot CRUD (`context_snapshots` migration 028). Mission profile CRUD + activate (`mission_profiles` migration 029). Reactive NATS subscription engine (`core/internal/reactive/engine.go`). `GET /api/v1/services/status` health aggregation (including explicit `ollama` readiness row). `MaxReconnects(-1)` + DB/NATS startup retry loops (45×2s). **Frontend:** `BrainsPage.tsx` (add/edit/delete/probe with type presets), `ContextSwitchModal.tsx` (Cache & Transfer / Start Fresh / Load Snapshot), `MissionProfilesPage.tsx` (role→provider table, NATS subscriptions, context strategy), Profiles tab in Settings, Services tab in System. |
| V7 Team B | Trigger Engine | **Migrations:** `trigger_rules` (025) + `trigger_executions` (026). **Backend:** `triggers.Store` (rule CRUD, in-memory cache, `LogExecution`, `ActiveCount`). `triggers.Engine` (CTS subscription on `swarm.mission.events.*`, 4-guard `evaluateRule` — cooldown, recursion depth, concurrency, condition — `fireTrigger` creates child run, `proposeTrigger` logs for approval). 6 HTTP handlers (`GET/POST/PUT/DELETE /api/v1/triggers`, `POST /toggle`, `GET /history`). Wired into `AdminServer` + `main.go` with graceful shutdown. **Frontend:** `TriggerRulesTab.tsx` (full CRUD UI — RuleCard, CreateRuleForm, guard badges, mode warnings). Trigger types + 5 async actions in `useCortexStore`. Automations → Triggers tab now live (was DegradedState). **Bug fixes:** `/runs/[id]/page.tsx` (`"use client"` + `use(params)` for Next.js 15+), `/docs/page.tsx` (Suspense boundary for `useSearchParams`). |
| V7 Conversation Log | Agent Transcript Browsing + Interjection | **Migration 030** (`conversation_turns`). **Backend:** `conversations.Store` (LogTurn, GetRunConversation, GetSessionTurns). `ConversationLogger` interface in `protocol/events.go` propagated Soma → Team → Agent (mirrors EventEmitter). 6 emission points in `processMessageStructured()` (system, user, tool_call, tool_result, interjection, assistant). Interjection via NATS mailbox `swarm.agent.{id}.interjection` — agent checks between ReAct iterations. 3 HTTP handlers (run conversation, session turns, interject). **Frontend:** `ConversationLog.tsx` (agent filter, 5s auto-poll, interjection input), `TurnCard.tsx` (role-based colors/icons/badges), `types/conversations.ts`. `/runs/[id]` tab bar (Conversation + Events). **Tests:** 13 Go store tests, 11 Go handler tests, 9 frontend tests. |
| V7 Inception Recipes | Structured Prompt Patterns for RAG | **Migration 031** (`inception_recipes`). **Backend:** `inception.Store` (CreateRecipe, GetRecipe, ListRecipes, SearchByTitle, IncrementUsage, UpdateQuality). `store_inception_recipe` + `recall_inception_recipes` internal tools (dual-persist: RDBMS + pgvector). Recipe recall integrated into `research_for_blueprint` pipeline (step 5). Interaction protocol updated: agents prompted to store recipes after complex tasks. 5 HTTP handlers (list, search, get, create, quality feedback). **Tests:** 16 Go store tests, 10 Go handler tests. |
| MCP Test Hardening | Service + Handler Coverage | **Backend Tests:** new suites for MCP library loading/config conversion, registry service CRUD/cache/find flows, toolset CRUD/ref resolution, and executor adapter result formatting. **Handler Tests:** DB-backed happy paths for MCP list/delete/tools/library-install plus toolset update matrix (happy/not-found/bad UUID/missing name/nil service). **Semantics:** `handleUpdateToolSet` returns `404` when the tool set does not exist. |
| V7 UI Gate A | Parallel UI Framework + Reliability Baseline | **Docs:** `docs/UI_FRAMEWORK_V7.md` + `docs/ui-delivery/*` lane playbooks and board. **UX:** `StatusDrawer`, `DegradedModeBanner`, `CouncilCallErrorCard`, `FocusModeToggle`, `SystemQuickChecks`, and `AutomationHub` baseline integrated. **Tests:** Full verification now green in current workspace (`go test ./...`, Vitest `322` passing tests, Playwright `51` passing / `4` skipped on 2026-03-02). |
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

Branch rules:
1. Start every non-hotfix change from a fresh branch off `main`.
2. Keep branch scope single-purpose (one feature/fix lane per branch).
3. Merge via squash after tests/docs are complete.

Commit rules (required):
1. Use conventional commit subject lines: `feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `chore:`.
2. Add commit commentary in the body for non-trivial changes:
   - `Why:` reason for change
   - `What:` key files/surfaces changed
   - `Validation:` exact tests/checks run
3. Do not commit without at least one validation command unless the change is docs-only.

Commit template:

```text
feat(scope): short summary

Why:
- ...

What:
- ...

Validation:
- ...
```

