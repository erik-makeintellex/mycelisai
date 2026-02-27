# Mycelis V7 — Development State

> **Updated:** 2026-02-27
> **References:** `mycelis-architecture-v7.md` (PRD), `docs/V7_IMPLEMENTATION_PLAN.md` (Blueprint)

---

## Progress Summary

```
Phase 19 (complete)
    → V7 Step 01 / Team D — Navigation (complete)
    → Workspace UX — Rename, LaunchCrew, MemoryExplorer redesign (complete)
    → V7 Team A — Event Spine (complete)
    → V7 Soma Workflow E2E — Consultations, run_id, Run Timeline UI, OpsWidget registry (complete)
    → In-App Docs Browser — /docs page, 31-entry manifest, 8 user guides (complete)
    → Provider CRUD + Mission Profiles + Reactive Subscriptions + Service Management (complete)
    → V7 Team B — Trigger Engine (complete)
    → V7 Team C — Scheduler (NEXT)
    → V7 Causal Chain UI — ViewChain.tsx (after C)
    → MCP Baseline — filesystem, memory, artifact-renderer, fetch (parallel)
    → Resource API standardization + Workspace Explorer activation (in progress)
```

---

## What Is Done

### UI Parallel Delivery Engagement (Activated)

Parallelized UI delivery is now explicitly engaged with lane-level playbooks and gate-based merges.

| Deliverable | File |
|------------|------|
| Parallel board with active gate model (A/B/C/RC) | `docs/ui-delivery/PARALLEL_DELIVERY_BOARD.md` |
| Lane A playbook (Global Ops UX) | `docs/ui-delivery/LANE_A_GLOBAL_OPS.md` |
| Lane B playbook (Workspace Reliability UX) | `docs/ui-delivery/LANE_B_WORKSPACE_RELIABILITY.md` |
| Lane C playbook (Workflow Surfaces UX) | `docs/ui-delivery/LANE_C_WORKFLOW_SURFACES.md` |
| Lane D playbook (System + Observability UX) | `docs/ui-delivery/LANE_D_SYSTEM_OBSERVABILITY.md` |
| Lane Q playbook (QA reliability/regression) | `docs/ui-delivery/LANE_Q_QA_REGRESSION.md` |
| Canonical UI framework v2.0 (I/O contract, state transitions, component templates, gate checklists) | `docs/UI_FRAMEWORK_V7.md` |
| UI workflow instantiation + bus plan (execution authority) | `docs/product/UI_WORKFLOW_INSTANTIATION_AND_BUS_PLAN_V7.md` |

Execution policy:
- Lanes run in parallel.
- Merges are gate-sequenced (`P0 -> P1 -> P2 -> RC`).
- Lane Q evidence is required at every gate.
- Team ownership model: Circuit (Lane A), Atlas (Lane B), Helios + Argus (Lane C), Argus (Lane D), Sentinel (Lane Q).

Gate A baseline progress:
- Added/updated UI tests for `StatusDrawer`, `DegradedModeBanner`, `CouncilCallErrorCard`, `MissionControlChat` error UX, `ShellLayout` status action wiring, and Automations landing expectations.
- Verified targeted gate suite: `48` tests passing across dashboard/pages/shell/teams/automations subsets.
- Added Playwright Gate A operational UX suite scaffold:
  - `interface/e2e/specs/v7-operational-ux.spec.ts` (6 scenarios: degraded banner lifecycle, status drawer access, council reroute via Soma, automations actionable hub, system quick checks, focus mode toggle).
  - Run attempt currently blocked locally until Playwright browsers are installed (`npx playwright install`).
- Framework hardening completed:
  - `docs/UI_FRAMEWORK_V7.md` upgraded as canonical UI source with explicit input/output model, global operational state contract, required state transitions, component instantiation templates, telemetry contract, and release gates.
- UI element planning baseline completed:
  - `docs/UI_ELEMENTS_PLANNING_V7.md` added as research-backed planning authority for UI element management and Soma interaction patterns.
  - In-app docs manifest updated to include UI element planning and explicit Archive section labeling for non-authoritative historical docs.
- Recovery UX hardening completed:
  - `DegradedModeBanner` Retry now re-checks services, refreshes council/missions, and re-initializes SSE when disconnected.
  - Workspace council direct-target indicator now synchronizes with global `councilTarget` (including one-click fallback to Soma from banner/error card).
  - `StatusDrawer` council failure highlighting now resolves from recent failed council message source instead of only current target selection.

### UI Instantiation + Bus Planning (Kickoff Complete)

Execution-grade planning is now defined for:
- guided team instantiation workflows
- shared channel input/output envelopes
- user-safe NATS exposure (Basic/Guided/Expert)
- parallel Sprint 0 and Sprint 1 work packages across Atlas/Helios/Circuit/Argus/Sentinel

Planning source:
- `docs/product/UI_WORKFLOW_INSTANTIATION_AND_BUS_PLAN_V7.md`
- `docs/ui-delivery/PARALLEL_DELIVERY_BOARD.md`

### Sprint 0 UI Implementation (Lane C Started)

Implemented Sprint 0 scaffolds for guided team instantiation and readiness gating.

| Deliverable | File |
|------------|------|
| Team instantiation contract models (`TeamProfileTemplate`, `ReadinessSnapshot`, I/O envelope) | `interface/lib/workflowContracts.ts` |
| Capability readiness gate UI | `interface/components/automations/CapabilityReadinessGateCard.tsx` |
| Guided wizard scaffold (Objective -> Profile -> Readiness -> Launch) | `interface/components/automations/TeamInstantiationWizard.tsx` |
| Wizard integration into Automations hub | `interface/components/automations/AutomationHub.tsx` |
| Wizard tests | `interface/__tests__/automations/TeamInstantiationWizard.test.tsx` |

Verification run:
- `cd interface && npx vitest run __tests__/automations/TeamInstantiationWizard.test.tsx __tests__/pages/AutomationsPage.test.tsx __tests__/dashboard/MissionControlChat.test.tsx __tests__/dashboard/CouncilCallErrorCard.test.tsx __tests__/dashboard/DegradedModeBanner.test.tsx __tests__/dashboard/StatusDrawer.test.tsx __tests__/shell/ShellLayout.test.tsx __tests__/teams/TeamsPage.test.tsx __tests__/pages/SystemPage.test.tsx` -> pass (48 tests)
- `cd interface && npm run build` -> pass
- `cd core && go test ./internal/mcp/ -count=1` -> pass
- `cd core && go test ./internal/server/ -run TestHandleMCP -count=1` -> pass
- `cd core && go test ./internal/swarm/ -run TestScoped -count=1` -> pass

### Resource API Standardization + Workspace Explorer (Current)

| Deliverable | File |
|------------|------|
| Shared API contract helpers (`extractApiData`, `extractApiError`, MCP result formatter) | `interface/lib/apiContracts.ts` |
| Store-level standardized parsing for MCP + services status endpoints | `interface/store/useCortexStore.ts` |
| Global services-status polling moved into shell/store path | `interface/components/shell/ShellLayout.tsx` |
| Degraded banner/status drawer/quick checks now consume shared store status | `interface/components/dashboard/DegradedModeBanner.tsx`, `interface/components/dashboard/StatusDrawer.tsx`, `interface/components/system/SystemQuickChecks.tsx` |
| Services tab now uses centralized status contract | `interface/app/(app)/system/page.tsx` |
| Workspace Explorer activated in Resources tab (filesystem MCP tool calls) | `interface/components/resources/WorkspaceExplorer.tsx`, `interface/app/(app)/resources/page.tsx` |
| Focus mode NATS status now sourced from global services status state | `interface/components/dashboard/MissionControl.tsx` |

Verification run:
- `cd interface && npm run build` -> pass
- `cd interface && npx vitest run __tests__/dashboard/DegradedModeBanner.test.tsx __tests__/dashboard/StatusDrawer.test.tsx __tests__/shell/ShellLayout.test.tsx __tests__/pages/ResourcesPage.test.tsx __tests__/pages/SystemPage.test.tsx` -> pass (18 tests)
- `cd interface && npx vitest run __tests__/store/useCortexStore.test.ts` -> pass (25 tests)

### Phase 19 Foundation

| Capability | Notes |
|-----------|-------|
| Agent + Provider lifecycle | Soma, Axon, Swarm fully wired |
| Conversation memory (pgvector) | Migrations 001-022 applied |
| Intent proof pipeline | Confirm token flow + CE-1 templates |
| Mutation detection in chat | `mutationTools` set in cognitive.go |
| Brains toggle persistence | SaveConfig() persists to cognitive.yaml |
| Auth middleware | `interface/.env.local`, header injected on all proxied requests |

---

### V7 Step 01 — Navigation (Team D)

Collapsed 12+ architecture-surface routes into 5 workflow-first panels.

| Deliverable | File |
|------------|------|
| ZoneA_Rail (5 items + Docs + Advanced toggle) | `interface/components/shell/ZoneA_Rail.tsx` |
| Automations page (6 tabs + deep-link + advanced gate) | `interface/app/(app)/automations/page.tsx` |
| Resources page (4 tabs + deep-link) | `interface/app/(app)/resources/page.tsx` |
| System page (5 tabs + advanced gate) | `interface/app/(app)/system/page.tsx` |
| DegradedState shared component | `interface/components/shared/DegradedState.tsx` |
| PolicyTab CRUD → ApprovalsTab | `interface/components/automations/ApprovalsTab.tsx` |
| 8 legacy routes → server redirects | `/wiring` `/architect` `/teams` `/catalogue` `/marketplace` `/approvals` `/telemetry` `/matrix` |
| 56 unit tests passing | `interface/__tests__/pages/`, `__tests__/shared/`, `__tests__/shell/` |

---

### V7 Team A — Event Spine

Persistent audit record for every mission action.

| Deliverable | File |
|------------|------|
| `mission_runs` table (023) | `core/migrations/023_mission_runs.up.sql` |
| `mission_events` table (024) | `core/migrations/024_mission_events.up.sql` |
| MissionEventEnvelope + EventType constants | `core/pkg/protocol/events.go` |
| events.Store (Emit, GetRunTimeline) | `core/internal/events/store.go` |
| runs.Manager (CreateRun, UpdateRunStatus, ListRecentRuns) | `core/internal/runs/manager.go` |
| GET /api/v1/runs/{id}/events | `core/internal/server/runs.go` |
| GET /api/v1/runs/{id}/chain | `core/internal/server/runs.go` |
| GET /api/v1/runs (global list) | `core/internal/server/runs.go` |
| TypeScript types (MissionRun, MissionEvent) | `interface/types/events.ts` |

---

### V7 Soma Workflow E2E

Complete loop from chat → consultation trace → proposal → confirm → run_id → timeline.

**Backend changes:**

| Deliverable | File |
|------------|------|
| ConsultationEntry type in ChatResponsePayload | `core/pkg/protocol/envelopes.go` |
| ReAct loop captures consult_council into ProcessResult.Consultations | `core/internal/swarm/agent.go` |
| agentResult.Consultations wired into chatPayload | `core/internal/server/cognitive.go` |
| HandleConfirmAction returns run_id in response | `core/internal/server/templates.go` |

**Frontend — Chat UI:**

| Deliverable | File |
|------------|------|
| Soma-locked header (no dropdown) | `interface/components/dashboard/MissionControlChat.tsx` |
| DirectCouncilButton (⚡ Direct popover) | `interface/components/dashboard/MissionControlChat.tsx` |
| DelegationTrace council cards below response | `interface/components/dashboard/MissionControlChat.tsx` |
| SomaActivityIndicator (live tool.invoked labels) | `interface/components/dashboard/MissionControlChat.tsx` |
| System message bubble → /runs/{id} pill | `interface/components/dashboard/MissionControlChat.tsx` |
| LaunchCrewModal always targets Soma, clears stale proposals | `interface/components/workspace/LaunchCrewModal.tsx` |

**Frontend — Runs UI:**

| Deliverable | File |
|------------|------|
| RunTimeline.tsx (auto-poll 5s, stops on terminal events) | `interface/components/runs/RunTimeline.tsx` |
| EventCard.tsx (colored dots, expandable payload) | `interface/components/runs/EventCard.tsx` |
| /runs/[id] page | `interface/app/(app)/runs/[id]/page.tsx` |
| /runs list page | `interface/app/(app)/runs/page.tsx` |

**Frontend — Store:**

| Deliverable | Notes |
|------------|-------|
| activeRunId, runTimeline, recentRuns state | Zustand slices in useCortexStore.ts |
| confirmProposal injects system message with run_id | Replaces old stub |
| fetchRunTimeline, fetchRecentRuns actions | Poll-ready |

**Frontend — OpsOverview:**

| Deliverable | File |
|------------|------|
| OpsWidget Registry (registerOpsWidget / getOpsWidgets) | `interface/lib/opsWidgetRegistry.ts` |
| RecentRunsSection widget (order 60, fullWidth) | `interface/components/dashboard/OpsOverview.tsx` |
| OpsOverview renders all 6 widgets from registry | `interface/components/dashboard/OpsOverview.tsx` |

---

### In-App Docs Browser

Fully functional. `/docs` page with sidebar, search, and rendered markdown.

| Deliverable | File |
|------------|------|
| GET /docs-api (manifest) | `interface/app/docs-api/route.ts` |
| GET /docs-api/[slug] (content, path-validated) | `interface/app/docs-api/[slug]/route.ts` |
| docsManifest.ts (31 entries, 7 sections) | `interface/lib/docsManifest.ts` |
| /docs page (sidebar + react-markdown, deep-link) | `interface/app/(app)/docs/page.tsx` |
| Internal .md link resolution (stays in-app) | `interface/app/(app)/docs/page.tsx` |
| Docs nav item in main rail (below Memory) | `interface/components/shell/ZoneA_Rail.tsx` |

**User guide docs (docs/user/):**

| Doc | Covers |
|-----|--------|
| core-concepts.md | Soma, Council, Mission, Run, Brain, Event, Trust, NATS, MCP |
| soma-chat.md | Message → delegation trace → proposal → confirm → run link |
| meta-agent-blueprint.md | Architect as meta-agent, blueprint structure, activation pipeline |
| run-timeline.md | Event types, colors, common patterns |
| automations.md | Triggers, schedules, approvals, teams, policy |
| resources.md | Brains, Cognitive Matrix, MCP tools, workspace, catalogue |
| memory.md | Semantic search, SitReps, artifacts, hot/warm/cold tiers |
| governance-trust.md | Trust scores, halts, approval flow, policy config |

---

### Provider CRUD, Mission Profiles & Reactive Subscriptions

Full hot-reload provider management, named workflow profiles with role→provider routing, context snapshot/restore, reactive NATS subscriptions, and service health dashboard.

**Migrations:**

| Migration | Table | File |
|-----------|-------|------|
| 028 | `context_snapshots` | `core/migrations/028_context_snapshots.up.sql` |
| 029 | `mission_profiles` | `core/migrations/029_mission_profiles.up.sql` |

**Backend:**

| Deliverable | File |
|------------|------|
| `AddProvider` / `UpdateProvider` / `RemoveProvider` with `RWMutex` | `core/internal/cognitive/router.go` |
| `POST /api/v1/brains` (add), `PUT /api/v1/brains/{id}` (update), `DELETE /api/v1/brains/{id}` (delete), `POST /api/v1/brains/{id}/probe` | `core/internal/server/brains.go` |
| Context snapshot CRUD — `POST /api/v1/context/snapshot`, `GET /api/v1/context/snapshots`, `GET /api/v1/context/snapshots/{id}` | `core/internal/server/context.go` |
| Mission profile CRUD + activate — `GET/POST/PUT/DELETE /api/v1/mission-profiles`, `POST /api/v1/mission-profiles/{id}/activate` | `core/internal/server/profiles.go` |
| Reactive NATS subscription engine — `Subscribe`, `Unsubscribe`, `ReactivateFromDB`, `Connected`, `ActiveSubscriptionCount` | `core/internal/reactive/engine.go` |
| `GET /api/v1/services/status` — NATS, PostgreSQL, Cognitive, Reactive health aggregation | `core/internal/server/services.go` |
| `MaxReconnects(-1)` unlimited NATS reconnects + ping interval | `core/internal/transport/nats/client.go` |
| DB startup retry loop (45×2s), NATS startup retry (45×2s), connection pool config, reactive reactivation on boot | `core/cmd/server/main.go` |
| Longer port-forward wait (30s), Core API wait (120s), WARN instead of FATAL for infra slow-start | `ops/lifecycle.py` |

**Frontend:**

| Deliverable | File |
|------------|------|
| Provider Add/Edit/Delete/Probe UI with type presets (Ollama, vLLM, LM Studio, OpenAI, Anthropic, Google, Custom) | `interface/components/settings/BrainsPage.tsx` |
| ContextSwitchModal — Cache & Transfer / Start Fresh / Load Snapshot strategies | `interface/components/settings/ContextSwitchModal.tsx` |
| MissionProfilesPage — role→provider table, NATS subscriptions editor, context strategy, auto-start | `interface/components/settings/MissionProfilesPage.tsx` |
| Profiles tab in Settings | `interface/app/(app)/settings/page.tsx` |
| Services tab in System — live polling, service cards, restart command reference with copy | `interface/app/(app)/system/page.tsx` |
| Mission profile + context snapshot types, state, and all async actions | `interface/store/useCortexStore.ts` |

---

### V7 Team B — Trigger Engine

Declarative IF/THEN rules evaluated on CTS event ingest. Four guards: cooldown, recursion depth, concurrency, condition (reserved). Default mode `propose` — auto-execute requires explicit policy.

**Migrations:**

| Migration | Table | File |
|-----------|-------|------|
| 025 | `trigger_rules` | `core/migrations/025_trigger_rules.up.sql` |
| 026 | `trigger_executions` | `core/migrations/026_trigger_executions.up.sql` |

**Backend:**

| Deliverable | File |
|------------|------|
| TriggerRule + TriggerExecution types, in-memory cache, CRUD, LogExecution, ActiveCount | `core/internal/triggers/store.go` |
| Engine — CTS subscription, 4-guard evaluateRule, fireTrigger (child run), proposeTrigger | `core/internal/triggers/engine.go` |
| 6 HTTP handlers — List, Create, Update, Delete, Toggle, History | `core/internal/server/triggers.go` |
| AdminServer wiring — Triggers + TriggerEngine fields, 6 routes | `core/internal/server/admin.go` |
| main.go — trigger store + engine init, graceful shutdown | `core/cmd/server/main.go` |

**Frontend:**

| Deliverable | File |
|------------|------|
| TriggerRulesTab — full CRUD UI, RuleCard, CreateRuleForm, guard badges, mode warnings | `interface/components/automations/TriggerRulesTab.tsx` |
| Trigger types + state + 5 async actions (fetch, create, update, delete, toggle) | `interface/store/useCortexStore.ts` |
| Automations → Triggers tab now renders live TriggerRulesTab (was DegradedState) | `interface/app/(app)/automations/page.tsx` |

**Bug fixes (pre-existing, discovered during build verification):**

| Fix | File |
|-----|------|
| Added `"use client"` + `use(params)` for Next.js 15+ async params | `interface/app/(app)/runs/[id]/page.tsx` |
| Wrapped `useSearchParams()` in Suspense boundary | `interface/app/(app)/docs/page.tsx` |

---

### MCP Test Hardening (Service + Handlers)

Comprehensive MCP coverage added across service, adapter, and HTTP handler layers.

| Deliverable | File |
|------------|------|
| Library tests (YAML load, lookup, config conversion) | `core/internal/mcp/library_test.go` |
| Service tests (Install/List/Get/Delete/UpdateStatus/CacheTools/ListTools/ListAllTools/Find*) | `core/internal/mcp/service_test.go` |
| Tool set service tests (CRUD, FindByName, ResolveRefs, nil DB guards) | `core/internal/mcp/toolsets_test.go` |
| Executor adapter tests (`FindToolByName`, `CallTool`, text formatting edge cases) | `core/internal/mcp/executor_test.go` |
| MCP handler DB-backed happy paths (`/servers`, `/tools`, `/library/install`) + raw install forbidden route | `core/internal/server/mcp_test.go` |
| Tool set handler update-path tests (happy, bad UUID, missing name, nil service) | `core/internal/server/mcp_toolsets_test.go` |
| Update not-found HTTP semantics for tool set update | `core/internal/server/mcp_toolsets.go` (`404` when tool set not found) |

---

### Swarm Parallel Activation

Blueprint activation now fans out team startup in parallel where safe, with idempotent race-safe insertion into active runtime state.

| Deliverable | File |
|------------|------|
| Bounded parallel team startup in `ActivateBlueprint` | `core/internal/swarm/activation.go` |
| Idempotent duplicate-skip under concurrent activation | `core/internal/swarm/activation.go` |
| New tests: idempotent repeat activation + concurrent activation no-dup | `core/internal/swarm/activation_test.go` |

---

## What Is Pending

### V7 Team C — Scheduler (NEXT)

**Depends on:** runs.Manager, events.Store.

| File to Create | Purpose |
|---------------|---------|
| `core/migrations/027_scheduled_missions.up.sql` | scheduled_missions table with next_run_at index |
| `core/internal/scheduler/scheduler.go` | Goroutine ticker, Suspend/Resume, checkDue() |
| `core/internal/server/schedules.go` | GET/POST/PUT/DELETE /api/v1/schedules + pause/resume |

**Done when:** Schedule creates a run on tick, enforces max_active_runs, suspends on NATS disconnect.

---

### Causal Chain UI (after B+C)

| File to Create | Purpose |
|---------------|---------|
| `interface/components/runs/ViewChain.tsx` | Parent → event → trigger → child run traversal |
| `interface/components/runs/RunChainNode.tsx` | Recursive tree node |
| `interface/app/(app)/runs/[id]/chain/page.tsx` | /runs/{id}/chain route |

Backend handler `GET /api/v1/runs/{id}/chain` is already live.

---

### MCP Baseline (Parallel — Independent of B/C)

Per `docs/V7_MCP_BASELINE.md`.

| Server | Status |
|--------|--------|
| `filesystem` MCP | BOOTSTRAP DEFAULT (auto-install/connect path present) |
| `fetch` MCP | BOOTSTRAP DEFAULT (auto-install/connect path present) |
| `memory` MCP | CURATED LIBRARY INSTALL (available, not bootstrap default) |
| `artifact-renderer` MCP | PLANNED (not bootstrap default yet) |

Resources → Workspace Explorer tab shows DegradedState until implemented.

---

## Current Navigation Structure

```
ZoneA_Rail
├── [logo] → /
├── Workspace     (Home)      → /dashboard
├── Automations   (Workflow)  → /automations
├── Resources     (FolderCog) → /resources
├── Memory        (Brain)     → /memory
├── Docs          (BookOpen)  → /docs
├── [System       (Activity)  → /system  ← Advanced Mode only]
└── Footer
    ├── Advanced toggle (Eye/EyeOff)
    └── Settings  (Settings)  → /settings
```

---

## Current Tab Map

| Page | Tabs | Notes |
|------|------|-------|
| `/automations` | Active · Drafts · Triggers · Approvals · Teams · Wiring* | *Wiring = Advanced Mode only |
| `/resources` | Brains · Tools · Workspace · Catalogue | |
| `/system` | Health · NATS · Database · Matrix · Debug · Services | Advanced Mode gated |

**Tabs still showing DegradedState (real data pending):**

| Page | Tab | Blocked by |
|------|-----|------------|
| Automations | Active Automations | Team C (Scheduler) |
| System | Event Health | Team A live data wiring |

---

## Current Build State

```
next build:         PASSES (all routes)
vitest:             Gate A + Sprint 0 suite passes (48 tests)
Go build:           go build ./... PASSES (includes triggers package)
Go tests:           188+ tests pass across 16 packages
                    Migrations 023-029 must be applied for full test coverage
Go test packages:   internal/server (157), internal/events (16), internal/runs (19), internal/triggers (new), others (~80)
MCP verification:   `go test ./internal/mcp/ -count=1` PASSES
                    `go test ./internal/server/ -run TestHandleMCP -count=1` PASSES
                    `go test ./internal/swarm/ -run TestScoped -count=1` PASSES
                    `go test ./... -count=1` has unrelated root-package conflict:
                    `probe.go` and `probe_test.go` both declare `main`
TypeScript check:   `cd interface && npx tsc --noEmit` currently fails on pre-existing test typing gaps:
                    `TelemetryRow.test.tsx` missing `afterEach`, and typed-mock issues in
                    `MemoryPage.test.tsx`, `PrimaryRoutes.test.tsx`, `ResourcesPage.test.tsx`
```

---

## Architecture Debt (Known Gaps)

| Gap | Location | Priority |
|-----|---------|---------|
| 2 pre-existing DashboardPage test failures | `__tests__/pages/DashboardPage.test.tsx` | Low (pre-V7) |
| 14 pre-existing test file transform errors | Various `__tests__/{workspace,dashboard,teams,...}` | Low (pre-V7) |
| Causal Chain UI | ViewChain.tsx + /runs/[id]/chain | After Team C |
| Automations → Active Automations: DegradedState | `app/(app)/automations/page.tsx` | Blocked by Team C |
| Full `go test ./...` root-package conflict | `core/probe.go`, `core/probe_test.go` | Medium |

---

## Next Potential Steps

1. Resolve full-suite blocker:
   - Fix root package build conflict (`probe.go` and `probe_test.go` both declare `main`) so `go test ./...` is clean.
2. Execute UI Gate A to completion:
   - Close Lane A + B P0 acceptance tests and attach evidence in lane files.
3. Extend MCP test coverage into connection lifecycle:
   - Add seam/stubs for `ClientPool.Connect/ReconnectAll/ShutdownAll` to validate status updates and degraded behavior without real subprocesses.
4. Harden API semantics consistency:
   - Ensure not-found cases across MCP handlers/toolset handlers map to `404` consistently (currently update-path is corrected).
5. Expand Workspace Explorer capabilities:
   - Add inline save for edited file previews and add basic delete/rename flows with governed safety prompts.

## Decision Log (Locked)

| Decision | Choice | Rationale |
|----------|--------|-----------|
| Trigger default mode | `propose` | Human-in-the-loop first posture |
| run_id requirement | Mandatory on all events | A mission is a definition; a run is an execution |
| Event persist order | DB first, then CTS publish | NATS offline = degrade, not data loss |
| Scheduler resolution | 1-minute | Acceptable for V7; sub-minute is Phase 20 |
| Multi-tenant schema | `tenant_id = 'default'` everywhere | Schema-ready, operationally single-tenant |
| Advanced mode gate | System nav + Neural Wiring tab | Reduces cognitive overload for normal users |
| Legacy route strategy | Server-side `redirect()` | Runs at server level, no client state needed |
| Tab deep-linking | `?tab=` URL search params + `<Suspense>` | Next.js 16 requirement for `useSearchParams()` |
| Docs API prefix | `/docs-api/` not `/api/docs/` | `/api/*` → Go proxy rewrite would intercept |
| Docs params | `await params` in route handler | Next.js 15+ async params requirement |
| Nav item order | Docs below Memory in main nav | Workflow items stay together; docs is reference |
