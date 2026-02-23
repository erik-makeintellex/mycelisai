# Mycelis V7 — Development State

> **Date:** 2026-02-22
> **References:** `mycelis-architecture-v7.md` (PRD), `docs/V7_IMPLEMENTATION_PLAN.md` (Blueprint)

---

## Where We Are

The V7 architecture is a 5-team sequential execution plan built on top of a completed Phase 19 foundation. One team's work is done. Four remain.

```
Phase 19 (complete)  →  V7 Step 01 (complete)  →  Workspace UX (complete)  →  V7 Teams A → B → C → E  (pending)
```

**MVP Agentry Plan:** `docs/MVP_AGENTRY_PLAN.md` — sprint-by-sprint plan from services verification through run timeline UI.

**Auth fix (2026-02-22):** `interface/.env.local` created, `ops/interface.py dev()` loads root `.env`. Next.js middleware now injects auth header on all proxied requests.

---

## Foundation: What Phase 19 Delivered (Pre-V7)

Phase 19 is the base layer. V7 builds on top of it without modifying it.

| Capability | Status | Notes |
|-----------|--------|-------|
| Agent + Provider lifecycle management | DONE | Soma, Axon, Swarm |
| Conversation memory (pgvector) | DONE | Migrations 001-022 applied |
| Scheduled team activation | DONE | In-memory scheduler (pre-V7, different from V7 cron) |
| Intent proof pipeline | DONE | Phase 19-B fixes applied |
| Confirm token flow | DONE | HandleConfirmAction wired |
| Mutation detection in chat handlers | DONE | `mutationTools` set in cognitive.go |
| Brains toggle persistence | DONE | SaveConfig() calls in brains.go |
| CE-1 orchestration templates | DONE | Templates 001-002 active |

---

## Workspace UX Polish (COMPLETE — 2026-02-22)

**Scope:** Rename Mission Control → Workspace, crew launch UX, offline guide, memory redesign, dead route fixes.

| Deliverable | Status | File |
|------------|--------|------|
| "Mission Control" → "Workspace" rename (rail, header, loading text) | DONE | `ZoneA_Rail.tsx`, `MissionControl.tsx`, `dashboard/page.tsx` |
| localStorage key migration (mission-chat → workspace-chat, split ratio) | DONE | `useCortexStore.ts`, `MissionControl.tsx` |
| SomaOfflineGuide (inv lifecycle.up command + retry button) | DONE | `MissionControlChat.tsx` |
| LaunchCrewModal (3-step guided crew launch) | DONE | `interface/components/workspace/LaunchCrewModal.tsx` |
| MemoryExplorer redesign (2-col, Hot behind Advanced Mode) | DONE | `interface/components/memory/MemoryExplorer.tsx` |
| OpsOverview dead routes fixed (`/missions/{id}/teams` removed, hrefs updated) | DONE | `interface/components/dashboard/OpsOverview.tsx` |

---

## V7 Step 01: Navigation Restructure (COMPLETE)

**Architecture team:** Team D (Navigation/IA Collapse)

**What it does:** Collapses 12+ architecture-surface routes into 5 workflow-first panels.

| Deliverable | Status | File |
|------------|--------|------|
| ZoneA_Rail (5 nav items + advanced toggle) | DONE | `interface/components/shell/ZoneA_Rail.tsx` |
| Automations page (6 tabs, deep-link, advanced gate) | DONE | `interface/app/(app)/automations/page.tsx` |
| Resources page (4 tabs, deep-link) | DONE | `interface/app/(app)/resources/page.tsx` |
| System page (5 tabs, deep-link, advanced gate) | DONE | `interface/app/(app)/system/page.tsx` |
| DegradedState component | DONE | `interface/components/shared/DegradedState.tsx` |
| PolicyTab CRUD migrated into ApprovalsTab | DONE | `interface/components/automations/ApprovalsTab.tsx` |
| 8 legacy routes → server redirects | DONE | `/wiring`, `/architect`, `/teams`, `/catalogue`, `/marketplace`, `/approvals`, `/telemetry`, `/matrix` |
| Authoritative IA doc | DONE | `docs/product/ia-v7-step-01.md` |
| Manual verification script | DONE | `docs/verification/v7-step-01-ui.md` |
| Unit tests (56 passing, 0 V7 failures) | DONE | `interface/__tests__/pages/`, `__tests__/shared/`, `__tests__/shell/` |
| E2E navigation spec | DONE | `interface/e2e/specs/navigation.spec.ts` |
| `next build` | PASSES | All 19 routes compile clean |

**Tabs currently showing DegradedState (placeholders for future teams):**

| Page | Tab | Waiting For |
|------|-----|-------------|
| Automations | Active Automations | V7 Team C (Scheduler) |
| Automations | Draft Blueprints | V7 Team E → D integration |
| Automations | Trigger Rules | V7 Team B (Trigger Engine) |
| System | Event Health | V7 Team A (Event Spine live data) |

---

## V7 Remaining Work — Strict Execution Order

The implementation plan mandates: **A → B → C → E** (Team D already done, Team A complete)

All remaining teams are backend-first. Frontend team E depends on Team A's APIs being live.

---

### Team A — Event Spine (COMPLETE — 2026-02-22)

**Scope:** Persistent event record for every mission action. The authoritative audit trail that all other V7 features build on.

| File Created | Purpose |
|-------------|---------|
| `core/migrations/023_mission_runs.up.sql` | `mission_runs` table — one row per execution instance |
| `core/migrations/024_mission_events.up.sql` | `mission_events` table — 17-field event record |
| `core/pkg/protocol/events.go` | `MissionEventEnvelope`, `EventType` constants, `EventEmitter` + `RunsManager` interfaces |
| `core/internal/events/store.go` | `Emit()`, `publishCTS()`, `GetRunTimeline()` |
| `core/internal/runs/manager.go` | `CreateRun()`, `CreateChildRun()`, `UpdateRunStatus()`, `GetRun()`, `ListRunsForMission()` |
| `core/internal/server/runs.go` | `GET /api/v1/runs/{id}/events`, `GET /api/v1/runs/{id}/chain` |
| `interface/types/events.ts` | TypeScript types: `MissionEventEnvelope`, `MissionRun`, `RunChainResponse`, `EVENT_TYPE_COLORS` |

| File Modified | Change |
|--------------|--------|
| `core/pkg/protocol/envelopes.go` | Added `MissionEventID string` (omitempty) to `CTSEnvelope` |
| `core/pkg/protocol/topics.go` | Added `TopicMissionEvents`, `TopicMissionEventsFmt`, `TopicMissionRunsFmt` |
| `core/internal/server/admin.go` | Added `Events`/`Runs` fields, registered 2 new routes |
| `core/internal/swarm/soma.go` | `SetRunsManager()` + `SetEventEmitter()` setters |
| `core/internal/swarm/activation.go` | Creates run, emits `mission.started`, propagates emitter to teams |
| `core/internal/swarm/team.go` | `SetEventEmitter()`, propagates to agents in `Start()` |
| `core/internal/swarm/agent.go` | `SetEventEmitter()`, emits `tool.invoked`/`tool.completed`/`tool.failed` |
| `core/internal/server/mission.go` | Sets `bp.MissionID`, returns `RunID` in `CommitResponse` |
| `core/cmd/server/main.go` | Initializes `events.Store` + `runs.Manager`, wires into Soma |

**Degraded mode:** NATS nil → event persists to DB, CTS publish skipped, warning logged. No panic.

**Status:** `go build ./...` passes. Migrations 023-024 must be applied before first run.

---

### Team B — Trigger Engine (after A)

**Scope:** Declarative rules: event type + predicate → fire downstream mission. Default mode = `propose` (safe).

**Depends on:** `mission_events` + `mission_runs` tables, `events.Store`, `runs.Manager`.

| File to Create | Purpose |
|---------------|---------|
| `core/migrations/025_trigger_rules.up.sql` | `trigger_rules` table + deferred FKs onto 023/024 tables |
| `core/migrations/026_trigger_executions.up.sql` | Audit table for every evaluation result |
| `core/internal/triggers/store.go` | Rule CRUD + in-memory cache (`LoadRules`, `CreateRule`, etc.) |
| `core/internal/triggers/engine.go` | Evaluation (match → cooldown guard → recursion guard → concurrency guard → fire) |
| `core/internal/server/triggers.go` | `GET/POST/PUT/DELETE /api/v1/triggers` handlers |

**Three guards (all mandatory):**
- **Cooldown:** skip if last execution within `cooldown_ms`
- **Recursion:** skip if run depth ≥ `max_depth` (hard ceiling: 10)
- **Concurrency:** skip if target mission has ≥ `max_active_runs` active runs

**Team B is done when:** Trigger rules CRUD works, `mission.completed` event fires a child run, all guards pass unit tests.

---

### Team C — Scheduler (after A)

**Scope:** Cron-based mission scheduling. In-process goroutine, 1-minute resolution, Postgres-backed. Suspends on NATS disconnect.

**Depends on:** `runs.Manager`, `events.Store`.

| File to Create | Purpose |
|---------------|---------|
| `core/migrations/027_scheduled_missions.up.sql` | `scheduled_missions` table with `next_run_at` index |
| `core/internal/scheduler/scheduler.go` | Goroutine ticker, `Suspend()`, `Resume()`, `checkDue()` |
| `core/internal/server/schedules.go` | `GET/POST/PUT/DELETE /api/v1/schedules` + pause/resume handlers |

**Team C is done when:** Schedule creates a run on tick, enforces `max_active_runs`, suspends on NATS disconnect, resumes on reconnect.

---

### Team E — Run Timeline + Chain UI (after A APIs live)

**Scope:** Frontend components that visualize the event spine data.

**Depends on:** `GET /api/v1/runs/{id}/events` and `GET /api/v1/runs/{id}/chain` being live.

| File to Create | Purpose |
|---------------|---------|
| `interface/components/runs/RunTimeline.tsx` | Vertical timeline; color-coded by event type |
| `interface/components/runs/EventCard.tsx` | Expandable event detail card |
| `interface/components/runs/ViewChain.tsx` | Causal chain tree (parent → event → trigger → child run) |
| `interface/components/runs/RunChainNode.tsx` | Recursive node component |
| `interface/app/(app)/runs/[id]/page.tsx` | Route: `/runs/{id}` |
| `interface/app/(app)/runs/[id]/chain/page.tsx` | Route: `/runs/{id}/chain` |

**Zustand additions (`useCortexStore.ts`):**
```typescript
activeRunId: string | null;
runTimeline: MissionEventEnvelope[];
runChain: RunChainNode | null;
isTimelineLoading: boolean;
fetchRunTimeline: (runId: string) => Promise<void>;
fetchRunChain: (runId: string) => Promise<void>;
clearRunTimeline: () => void;
```

**V7 Step 01 tabs that will be upgraded when Team E is done:**
- Automations → Active Automations: shows inline run timelines
- Automations → Trigger Rules: links to chain views for triggered runs

---

## MCP Baseline (Parallel Track — Independent of A/B/C/E)

Per `docs/V7_MCP_BASELINE.md`. Not yet started.

| Server | Status | Purpose |
|--------|--------|---------|
| `filesystem` MCP | NOT STARTED | Sandboxed file I/O in `/workspace` |
| `memory` MCP | NOT STARTED | pgvector semantic store/recall |
| `artifact-renderer` MCP | NOT STARTED | Inline structured output |
| `fetch` MCP | NOT STARTED | Domain-allowlisted HTTP GET |

**Resources → Workspace Explorer tab** shows DegradedState until this is built.

---

## Architecture Debt (Known Gaps)

| Gap | Location | Priority |
|-----|---------|---------|
| DashboardPage.test.tsx — 2 pre-existing assertion failures | `__tests__/pages/DashboardPage.test.tsx` | Low (pre-V7) |
| 14 pre-existing test file transform errors | Various `__tests__/{workspace,dashboard,teams,...}` | Low (pre-V7) |
| System → Event Health tab: static placeholder | `app/(app)/system/page.tsx` | Blocked by Team A |
| Automations → Active Automations: DegradedState | `app/(app)/automations/page.tsx` | Blocked by Team C |
| Automations → Trigger Rules: DegradedState | `app/(app)/automations/page.tsx` | Blocked by Team B |

---

## Current Build State

```
next build:   PASSES  (all 19 routes)
vitest:       54/56 V7 tests pass (2 pre-existing DashboardPage failures)
Go build:     go build ./... passes (Team A wired, no import cycles)
Go tests:     inv core.test passes (Phase 19 baseline; migrations 023-024 need applying for Team A tests)
```

---

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
