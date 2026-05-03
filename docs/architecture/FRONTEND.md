# Mycelis Cortex - Frontend Specification
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Load this doc when: working on Next.js routes, React components, Zustand state, UI/API integration, or frontend delivery gates.
>
> Related:
> [Overview](OVERVIEW.md) |
> [Backend](BACKEND.md) |
> [Operations](OPERATIONS.md) |
> [UI Target + Transaction Contract](UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md)

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
- default operator navigation is intentionally smaller than the full route set: `Soma`, the current AI Organization when available, `Docs`, `Status`, and `Settings` stay visible by default, while `Groups`, `Activity/Runs`, `Resources/MCP`, `Memory`, and `System` are advanced/admin support routes

| Shell area | Component | Purpose |
| --- | --- | --- |
| Rail | `ZoneA_Rail.tsx` | Workflow-first navigation + advanced toggle + settings |
| Main workspace container | `ZoneB_Workspace.tsx` | Hosts route content |
| Global degraded feedback | `DegradedModeBanner.tsx` | Recovery-oriented degraded state messaging |
| Global status inspector | `StatusDrawer.tsx` | Unified service/chat failure diagnostics |

Implementation note:
- governed approval review is store-driven through `GovernanceModal.tsx` in advanced workspace contexts, not a legacy shell side rail
- stream/telemetry surfaces are route-local, with `NatsWaterfall.tsx` and `SignalDetailDrawer.tsx` carrying the active advanced inspection path

---

## 3. Full GUI Route Set

Current `page.tsx` route count: `22`.

### 3.1 Primary workflow routes

| Route | Source | Primary surface |
| --- | --- | --- |
| `/` | `app/(marketing)/page.tsx` | Product landing |
| `/dashboard` | `app/(app)/dashboard/page.tsx` | Soma intent start and AI Organization entry flow |
| `/organizations/[id]` | `app/(app)/organizations/[id]/page.tsx` | Soma-primary AI Organization workspace |
| `/automations` | `app/(app)/automations/page.tsx` | Automation hub + tabs |
| `/resources` | `app/(app)/resources/page.tsx` | Advanced resources (tools/files/AI engines/roles) |
| `/memory` | `app/(app)/memory/page.tsx` | Advanced memory explorer |
| `/docs` | `app/(app)/docs/page.tsx` | In-app markdown docs browser |
| `/system` | `app/(app)/system/page.tsx` | Advanced diagnostics and quick checks |
| `/settings` | `app/(app)/settings/page.tsx` | Profile, mission profiles, people/access, and advanced setup |

### 3.2 Execution and inspection routes

| Route | Source | Primary surface |
| --- | --- | --- |
| `/groups` | `app/(app)/groups/page.tsx` | Advanced group operations |
| `/runs` | `app/(app)/runs/page.tsx` | Run list and status summary |
| `/runs/[id]` | `app/(app)/runs/[id]/page.tsx` | Run conversation/events tabs |
| `/missions/[id]/teams` | `app/(app)/missions/[id]/teams/page.tsx` | Mission team actuation view |
| `/settings/tools` | `app/(app)/settings/tools/page.tsx` | Redirect into Settings tools tab |

### 3.3 In-app docs API routes

| Route | Source | Purpose |
| --- | --- | --- |
| `/docs-api` | `app/docs-api/route.ts` | Docs manifest for in-app browser |
| `/docs-api/[slug]` | `app/docs-api/[slug]/route.ts` | Slug-validated doc content |
| `/api/chat` | `app/(app)/api/chat/route.ts` | Secondary proxy path to `/api/v1/chat` |

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

Component files under `interface/components`: `96` (`.tsx` and `.ts`) across `27` folders.

| Folder | Count | Primary responsibility |
| --- | ---: | --- |
| `dashboard/` | 10 | Workspace mission-control chat, proposal/inspector, and shared support surfaces |
| `workspace/` | 10 | Wiring/editor and launch workflow surfaces |
| `settings/` | 9 | Providers, MCP, profiles, users/groups |
| `automations/` | 6 | Hub, approvals tab, trigger rules, instantiation wizard |
| `shell/` | 5 | Layout and navigation shell |
| `wiring/` | 3 | Graph/wiring editor node-edge surfaces |
| `runs/` | 4 | Run timeline and conversation components |
| `teams/` | 4 | Team and group management |
| `memory/` | 4 | Hot/warm/cold memory panes |
| `catalogue/` | 3 | Capability cards/editor |
| `matrix/` | 3 | Cognitive matrix views |
| Other folders | 35 | Shared, charts, missions, stream, system, command, organization, and utility surfaces |

---

## 5. Tab and Navigation Model

| Surface | Tabs / modes | Notes |
| --- | --- | --- |
| `/automations` | `active`, `triggers`, `approvals`, `teams`, `wiring` | `teams` and `wiring` are advanced-mode gated |
| `/teams` | direct route | team roster, lead-entry hub, and template specialization |
| `/teams/create` | direct route | guided team-creation workflow with Soma handoff |
| `/resources` | `tools`, `workspace`, `engines`, `roles` | Advanced support hub |
| `/system` | `health`, `nats`, `database`, `services` | Advanced diagnostics |
| `/settings` | `profile`, `profiles`, `users`, `engines`, `tools` | Preferences, access, and optional advanced setup |
| `/runs/[id]` | `conversation`, `events` | Split run investigation view |

### 5.1 MVP audit decisions (RC lock)

| Route / tab | Decision | Team Lead-first reason |
| --- | --- | --- |
| `/dashboard` | keep | first-run entry and AI Organization creation remain the primary operator start |
| `/organizations/[id]` | keep | Team Lead workspace is the compatibility-protected Soma-primary operating surface |
| `/automations` `active`, `triggers`, `approvals` | keep | directly support guided operation, recurring work, and governed decisions |
| `/automations` `teams`, `wiring` | revise | keep available, but only in advanced mode so the default workflow stays simpler |
| `/groups`, `/activity`, `/runs`, `/resources`, `/memory`, `/system` | revise | remain shipped as advanced/admin support routes, not default operator navigation |
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
| Stream events | `GET /api/v1/stream` (EventSource) | `signal.StreamHandler` |
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
2. Workspace defaults to Soma (`/api/v1/chat`) and only routes to direct council endpoint when user targeting demands it
3. stream and service health must feed shared status/failure models used by banner, drawer, and chat blockers
4. new UI flows must map API effects and failure affordances in `UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md`
5. shared store helpers that carry graph/proposal/persistence behavior live in focused modules such as `interface/store/cortexStoreUtils.ts` and require direct test coverage (`interface/__tests__/store/cortexStoreUtils.test.ts`)

---

## 8. Current Gaps And Risks

| Area | Current state | Delivery status |
| --- | --- | --- |
| MVP browser gate scope | default Playwright coverage is now trimmed to Team Lead-first MVP routes/tabs; legacy V7 operational and raw telemetry specs remain out of the default RC gate | `COMPLETE` audit alignment |
| Max-lines gate pressure | `core/internal/swarm/agent.go`, `core/internal/swarm/internal_tools.go`, and the largest remaining store slices (`cortexStoreMissionChatSlice.ts`, `cortexStoreMissionDraftSlice.ts`, `cortexStoreResourceCatalogSlice.ts`) still carry the main frontend cap pressure | `REQUIRED` Slice 4 |
| Created-team communications inspector | architecture contract exists but dedicated team workspace/communications tabs are not fully delivered | `BLOCKED` Slice 7 |
| Docs and runs route browser depth | route-level smoke coverage now exists, but failure/recovery depth (error branches + interjection/terminal transitions) still needs expansion | `NEXT` |

---

## 9. Testing Alignment For GUI Delivery

Current baseline references live in `docs/TESTING.md`.

Required proof for UI-affecting changes:
1. Vitest component/store tests for affected terminal states and failure behavior
2. integration checks proving route/API mapping correctness
3. Playwright flow proof for user-critical paths (Workspace, automations, degraded/recovery)
4. live-backend Playwright proof when proxy/back-end contract is touched

---

## 10. Development Strategy (Frontend-Facing)

This strategy aligns GUI work with active architecture slices and avoids speculative rewrites.

### 10.1 Stream A - Slice 2 UX stabilization (`ACTIVE`)

Scope:
- simplify default Workspace density while preserving diagnostics via progressive disclosure
- lock Soma direct-first chat behavior for routine prompts
- align reroute/recovery copy and interactions across `MissionControlChat`, `CouncilCallErrorCard`, and `DegradedModeBanner`

Proof targets:
- focused Vitest for Team Lead action, blocker, and degraded-state components
- Playwright MVP route/tab coverage (`navigation`, `layout`, `missions`, `automations`, `settings`, `teams`, `proposals`, `accessibility`)

### 10.2 Stream B - Contract-safe store/API cleanup (`ACTIVE`)

Scope:
- split hot-path store logic into bounded modules without changing API contracts
- keep MCP tool registry surfaces aligned to the curated-library install contract and disabled raw-admin boundary
- keep API endpoint usage aligned with backend route ownership

Proof targets:
- `uv run inv interface.test`
- `cd interface && npx vitest run --reporter=dot`
- `uv run inv interface.build`

### 10.3 Stream C - Slice 4 complexity reduction (`REQUIRED`)

Scope:
- extract high-risk logic from oversized files while preserving existing behavior and telemetry semantics
- prioritize no-regression extraction over feature addition

Proof targets:
- `uv run inv quality.max-lines --limit 300`
- `uv run inv ci.baseline`

### 10.4 Stream D - Slice 7 team workspace contract (`BLOCKED -> NEXT once prerequisites land`)

Scope once unblocked:
- deliver created-team workspace tabs and communication filters
- add explicit operator controls for interject/reroute/pause-resume where valid
- map team command + `signal.status`/`signal.result` outputs to inspectable UI state

Proof targets:
- route-level Vitest coverage for communications inspector
- integration and product-flow tests for team command/result lifecycle

---

## 11. Implementation Guardrails

1. do not add new UI flows without explicit terminal-state mapping
2. do not introduce new API calls before confirming handler ownership in backend routes
3. keep `/docs` manifest in sync for any new authoritative architecture/development document
4. maintain runner contract: `uv run inv ...` for operational commands and test gates
