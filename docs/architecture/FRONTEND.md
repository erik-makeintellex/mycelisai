# Mycelis Frontend Implementation Contract
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Load this doc when: working on Next.js routes, React components, Zustand state, UI/API integration, or frontend delivery gates.
>
> Related:
> [Canonical PRD](../architecture-library/MYCELIS_CANONICAL_PRD.md) |
> [Backend](BACKEND.md) |
> [Operations](OPERATIONS.md) |
> [Mycelis Canonical PRD](../architecture-library/MYCELIS_CANONICAL_PRD.md)

> Baseline source of truth for this document:
> code-audited against `interface/app/**`, `interface/components/**`, `interface/store/useCortexStore.ts`,
> `core/internal/server/admin.go`, and `core/cmd/server/main.go` on 2026-03-31.

---

## 1. Current Stack

| Component | Version | Notes |
| --- | --- | --- |
| Next.js | 16.1.6 | App Router, Turbopack |
| React | 19.2.3 | Client components for stateful/interactive surfaces |
| Tailwind CSS | v4 | Cortex tokenized theme |
| Zustand | 5.0.11 | Single shared app store |
| Vitest | 4.0.18 | Unit/component/store tests |
| Playwright | 1.58.2 | Browser e2e and live-backend checks |

---

## 2. Shell Architecture

Primary app layout is implemented in `interface/app/(app)/layout.tsx` and `interface/components/shell/ShellLayout.tsx`.

Release-candidate MVP note:
- default operator navigation keeps the natural work-support paths visible: `Soma`, `Groups`, `Resources`, and `Docs`; `Status` and `Settings` remain reachable support controls
- `Activity/Runs`, `Memory`, `System`, and deeper inspect routes remain Admin tools; direct visits must show an explicit same-route release gate when Admin tools are off

| Shell area | Component | Purpose |
| --- | --- | --- |
| Rail | `ZoneA_Rail.tsx` | Workflow-first navigation + advanced toggle + settings |
| Main workspace container | `ZoneB_Workspace.tsx` | Hosts route content |
| Global degraded feedback | `DegradedModeBanner.tsx` | Recovery-oriented degraded state messaging |
| Global status inspector | `StatusDrawer.tsx` | Unified service/chat failure diagnostics |

Implementation note:
- governed approval review is store-driven through `GovernanceModal.tsx` in advanced workspace contexts, not a legacy shell side rail
- stream/telemetry surfaces are route-local, with `NatsWaterfall.tsx` and `SignalDetailDrawer.tsx` carrying the active advanced inspection path
- Soma chat renders returned artifacts inline, including images, audio, video, charts, code, and documents; governed proposals surface whether work is one-shot, scheduled, continuous, event-driven, and whether it connects to current-team or multi-team NATS lanes
- Soma chat may append compact system cards from typed `thread_event` stream payloads such as approval sent, execution started, output ready, or attention required; raw subjects and payload dumps stay in Inspect surfaces

---

## 3. Full GUI Route Set

Current `page.tsx` route count: `28`.

### 3.1 Primary workflow routes

| Route | Source | Primary surface |
| --- | --- | --- |
| `/` | `app/(marketing)/page.tsx` | Redirect to Soma workspace |
| `/login` | `app/login/page.tsx` | Authentication entry |
| `/access-denied` | `app/(app)/access-denied/page.tsx` | Access recovery guidance |
| `/dashboard` | `app/(app)/dashboard/page.tsx` | Soma threaded workspace |
| `/organizations/[id]` | `app/(app)/organizations/[id]/page.tsx` | Compatibility organization context using the Soma operating surface |
| `/automations` | `app/(app)/automations/page.tsx` | Automation hub + tabs |
| `/resources` | `app/(app)/resources/page.tsx` | Resources (output files, capabilities, exchange, context, AI engines, roles) |
| `/groups` | `app/(app)/groups/page.tsx` | Collaboration lanes, workflow logs, retained outputs |
| `/teams` | `app/(app)/teams/page.tsx` | Team workspace and review queue |
| `/teams/create` | `app/(app)/teams/create/page.tsx` | Guided Soma team creation |
| `/memory` | `app/(app)/memory/page.tsx` | Advanced memory explorer |
| `/docs` | `app/(app)/docs/page.tsx` | In-app markdown docs browser |
| `/system` | `app/(app)/system/page.tsx` | Advanced diagnostics and quick checks |
| `/settings` | `app/(app)/settings/page.tsx` | Profile, mission profiles, people/access, and advanced setup |

### 3.2 Execution and inspection routes

| Route | Source | Primary surface |
| --- | --- | --- |
| `/activity` | `app/(app)/activity/page.tsx` | Advanced activity stream |
| `/runs` | `app/(app)/runs/page.tsx` | Run list and status summary |
| `/runs/[id]` | `app/(app)/runs/[id]/page.tsx` | Run conversation/events tabs |
| `/runs/[id]/chain` | `app/(app)/runs/[id]/chain/page.tsx` | Run chain visualization |
| `/missions/[id]/teams` | `app/(app)/missions/[id]/teams/page.tsx` | Mission team actuation view |
| `/settings/tools` | `app/(app)/settings/tools/page.tsx` | Redirect into Settings tools tab |

### 3.3 In-app docs API routes

Docs and proxy support routes live outside the `page.tsx` count: `/docs-api`, `/docs-api/[slug]`, and `/api/chat`.

### 3.4 Legacy redirect routes (still shipped)

| Route | Redirect target |
| --- | --- |
| `/wiring` | `/automations?tab=wiring` |
| `/architect` | `/automations?tab=wiring` |
| `/catalogue` | `/resources?tab=roles` |
| `/marketplace` | `/resources?tab=roles` |
| `/approvals` | `/automations?tab=approvals` |
| `/telemetry` | `/system?tab=health` |
| `/matrix` | `/settings?tab=engines` |
| `/settings/brain` | `/settings?tab=engines` |

---

## 4. GUI Surface Inventory

Component ownership is organized by behavior rather than a static file-count snapshot:

| Folder family | Primary responsibility |
| --- | --- |
| `soma/`, `dashboard/`, `workspace/` | Soma conversation, compact governance, Outcome summaries, deliverables, and workspace overlays |
| `teams/`, `organizations/` | Team/group execution review and compatibility organization contexts |
| `resources/`, `catalogue/`, `settings/` | Output access, capabilities, providers, identity, scopes, and configuration |
| `memory/`, `activity/`, `runs/`, `system/` | Admin inspection, continuity, receipts, events, and deployment health |
| `automations/`, `approvals/`, `wiring/` | Scheduled/event-driven work, approval queues, and advanced workflow structure |
| `shell/`, `layout/`, `shared/`, `ui/` | Navigation, responsive composition, accessibility, and reusable primitives |

Do not duplicate live component counts in docs. Use the repository tree for inventory and keep this contract focused on ownership boundaries.

---

## 5. Tab and Navigation Model

| Surface | Tabs / modes | Notes |
| --- | --- | --- |
| `/automations` | `active`, `triggers`, `approvals`, `teams`, `wiring` | `teams` and `wiring` are advanced-mode gated |
| `/teams` | direct route | team roster, lead-entry hub, and template specialization |
| `/teams/create` | direct route | guided team-creation workflow with Soma handoff |
| `/resources` | `tools`, `workspace`, `engines`, `roles`, `exchange`, `deployment-context` | Operator support hub for files, capabilities, context, and managed configuration |
| `/groups` | side group list + `overview`, `outputs`, `message`, `settings`, `create` | Group operations use a bounded master-detail layout: group records stay in a scrollable side rail, while selected-group scope, retained outputs, broadcasts, config, and creation use compact tabs |
| `/system` | `health`, `nats`, `database`, `services`, `deployments` | Admin tools diagnostics |
| `/settings` | `profile`, `profiles`, `users`, `engines`, `tools` | Preferences, access, and optional admin setup |
| `/runs/[id]` | `conversation`, `events` | Split run investigation view |

Deep management pages should prefer a small set of related page-level tabs over stacking every panel vertically. Default to the most common review context, keep tab labels short, preserve `tablist`/`tab`/`tabpanel` semantics, support arrow-key plus Home/End tab movement, keep feedback notices in live regions, keep a clear route back to Soma on advanced pages, and bound generated prompts, logs, lists, and artifacts inside the active panel instead of increasing whole-page scroll. When a page has a dense record picker, use a master-detail layout where the picker stays in a bounded scroll rail and the selected item owns the tabbed detail surface. Keep filters collapsed by default and move the highest-value actions, such as opening Soma, outputs, folders, and lead workspaces, into the first selected-item viewport.

Generated file outputs should present the retained user outcome first. HTML/code artifacts with a saved workspace path must expose an openable rendered-file action and local folder access instead of treating raw source text as the primary result. Admin diagnostics links launched from Admin tools pages must preserve Admin tools access in the destination route.

### 5.1 MVP audit decisions (RC lock)

| Route / tab | Decision | Team Lead-first reason |
| --- | --- | --- |
| `/dashboard` | keep | first-run entry is the Soma threaded workspace for natural asking, approval, output review, and Outcome re-entry |
| `/organizations/[id]` | keep | Team Lead workspace is the compatibility-protected Soma-primary operating surface |
| `/automations` `active`, `triggers`, `approvals` | keep | directly support guided operation, recurring work, and governed decisions |
| `/automations` `teams`, `wiring` | revise | keep available, but only in advanced mode so the default workflow stays simpler |
| `/groups`, `/resources` | keep | normal operator support routes for collaboration lanes, retained outputs, generated files, and capability readiness |
| `/activity`, `/runs`, `/memory`, `/system` | revise | remain shipped as Admin tools routes, not default operator navigation |
| `/settings` `profile`, `profiles`, `users` | keep | operator-visible preferences and access are still MVP-worthy |
| `/settings` `engines`, `tools` | revise | advanced setup belongs behind explicit advanced mode |
| legacy redirects | keep as redirects | preserve bookmarks while funneling usage back into the reduced MVP route set |

---

## 6. GUI -> API Contract Map

Frontend API traffic primarily originates from:
- the composed Zustand entrypoint in `interface/store/useCortexStore.ts`
- slice modules under `interface/store/cortexStore*Slice.ts`
- route-level pages in `interface/app/(app)/**`
- targeted feature components (memory, matrix, teams, dashboard)

### 6.1 Core interaction and orchestration

| Capability | Frontend endpoint | Backend route owner |
| --- | --- | --- |
| Soma chat | `POST /api/v1/chat` | `AdminServer.HandleChat` |
| Direct council chat | `POST /api/v1/council/{member}/chat` | `AdminServer.HandleCouncilChat` |
| Council members | `GET /api/v1/council/members` | `AdminServer.HandleListCouncilMembers` |
| Intent negotiate/commit | `POST /api/v1/intent/negotiate`, `POST /api/v1/intent/commit` | `AdminServer` intent handlers |
| Confirm action | `POST /api/v1/intent/confirm-action` | `AdminServer.HandleConfirmAction` |

### 6.2 Operations and observability

| Capability | Frontend endpoint | Backend route owner |
| --- | --- | --- |
| Stream events | `GET /api/v1/stream` (EventSource) | `signal.StreamHandler`; typed `thread_event` payloads may append compact Soma-thread cards |
| Services status | `GET /api/v1/services/status` | `AdminServer.HandleServicesStatus` |
| Telemetry | `GET /api/v1/telemetry/compute` | `AdminServer.HandleTelemetry` |
| Runs list/events | `GET /api/v1/runs`, `GET /api/v1/runs/{id}/events` | `AdminServer` run handlers |
| Run conversation/interject | `GET /api/v1/runs/{id}/conversation`, `POST /api/v1/runs/{id}/interject` | conversation handlers |

### 6.3 Governance, teams, memory, and capabilities

| Capability | Frontend endpoint | Backend route owner |
| --- | --- | --- |
| Governance policy/pending/resolve | `/api/v1/governance/*` | governance handlers |
| Team/group surfaces | `/api/v1/teams`, `/api/v1/teams/detail`, `/api/v1/groups*` | identity and groups handlers |
| Memory | `/api/v1/memory/search`, `/api/v1/memory/sitreps`, `/api/v1/memory/stream` | memory handlers |
| Brains/providers | `/api/v1/brains*`, `/api/v1/cognitive/*` | brains + cognitive handlers |
| MCP | `/api/v1/mcp/servers*`, `/api/v1/mcp/library*`, `/api/v1/mcp/tools` | mcp handlers |
| Triggers/proposals/catalogue/artifacts | `/api/v1/triggers*`, `/api/v1/proposals*`, `/api/v1/catalogue/agents*`, `/api/v1/artifacts*` | corresponding domain handlers |

### 6.4 Route ownership notes

| Endpoint | Ownership note |
| --- | --- |
| `POST /api/v1/swarm/broadcast` | wired directly in `core/cmd/server/main.go` via `soma.HandleBroadcast` |
| `POST /api/v1/mcp/install` | intentionally returns `403`; the shipped UI now routes installs through the curated library flow at `/api/v1/mcp/library/install` |

---

## 7. State Orchestration Standard

All shared app state is composed through `interface/store/useCortexStore.ts`, with behavior split across bounded slice modules under `interface/store/`.

Rules:
1. execution-facing flows must classify terminal states as `answer`, `proposal`, `execution_result`, or `blocker`
2. Workspace defaults to Soma (`/api/v1/chat`); direct specialist endpoints are limited to explicit admin or compatibility contexts
3. stream and service health must feed shared status/failure models used by banner, drawer, and chat blockers
4. new UI flows must map API effects and failure affordances in `docs/architecture-library/MYCELIS_CANONICAL_PRD.md`
5. shared store helpers that carry graph/proposal/persistence behavior live in focused modules such as `interface/store/cortexStoreUtils.ts` and require direct test coverage (`interface/__tests__/store/cortexStoreUtils.test.ts`)

---

## 8. Delivery State Ownership

Current priorities, open risks, and acceptance status live only in `.state/V8_DEV_STATE.md`. This contract defines stable frontend ownership and proof expectations; it must not grow a parallel delivery scoreboard or copy transient file counts.

---

## 9. Testing Alignment For GUI Delivery

Current baseline references live in `docs/TESTING.md`.

Required proof for UI-affecting changes:
1. Vitest component/store tests for affected terminal states and failure behavior
2. integration checks proving route/API mapping correctness
3. Playwright flow proof for user-critical paths (Workspace, automations, degraded/recovery)
4. live-backend Playwright proof when proxy/back-end contract is touched

Use focused component and browser proof for the touched surface when the global Interface unit harness is under active cleanup or hangs on unrelated legacy suites. Do not use a focused pass to hide a touched failure: record the global harness issue in `.state/V8_DEV_STATE.md`, keep the focused proof named, and schedule harness cleanup as its own slice.

## 10. Development Strategy (Frontend-Facing)

The active order and status live in `.state/V8_DEV_STATE.md`; this document defines stable frontend boundaries rather than a second scoreboard. Current work should extend the Workspace/Outcome operating model, keep machine detail behind deliberate inspection, and avoid speculative surface replacement.

Frontend slices should:
- preserve Soma as the default conversational workspace and Outcomes as durable user-owned work
- use progressive disclosure for proof, recovery, capabilities, and runtime detail
- retain one primary scroll owner with bounded local scrolling only for content such as tables, code, logs, or record lists
- preserve keyboard, URL, refresh, Back, and cross-device continuity
- keep API usage aligned with backend ownership and normalized terminal states
- extract oversized logic only in bounded, behavior-preserving slices

Proof targets are focused Vitest for changed states, TypeScript and production build gates, `uv run inv quality.max-lines --limit 385` (350-line target plus 10% tolerance), and headed Playwright for every changed operator journey. Promotion still follows the feature, integrated `dev`, and release-candidate gates in `docs/TESTING.md`.

---

## 11. Implementation Guardrails

1. do not add new UI flows without explicit terminal-state mapping
2. do not introduce new API calls before confirming handler ownership in backend routes
3. keep `/docs` manifest in sync for any new authoritative architecture/development document
4. maintain runner contract: `uv run inv ...` for operational commands and test gates
