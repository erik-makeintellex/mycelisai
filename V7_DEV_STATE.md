# Mycelis V7 — Development State

> **Updated:** 2026-02-24
> **References:** `mycelis-architecture-v7.md` (PRD), `docs/V7_IMPLEMENTATION_PLAN.md` (Blueprint)

---

## Progress Summary

```
Phase 19 (complete)
    → V7 Step 01 / Team D — Navigation (complete)
    → Workspace UX — Rename, LaunchCrew, MemoryExplorer redesign (complete)
    → V7 Team A — Event Spine (complete)
    → V7 Soma Workflow E2E — Consultations, run_id, Run Timeline UI, OpsWidget registry (complete)
    → In-App Docs Browser — /docs page, 30-entry manifest, 8 user guides (complete)
    → Provider CRUD + Mission Profiles + Reactive Subscriptions + Service Management (complete)
    → V7 Team B — Trigger Engine (NEXT)
    → V7 Team C — Scheduler (after B)
    → V7 Causal Chain UI — ViewChain.tsx (after B+C)
    → MCP Baseline — filesystem, memory, artifact-renderer, fetch (parallel)
```

---

## What Is Done

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
| docsManifest.ts (30 entries, 7 sections) | `interface/lib/docsManifest.ts` |
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

## What Is Pending

### V7 Team B — Trigger Engine (NEXT)

**Depends on:** mission_events + mission_runs tables, events.Store, runs.Manager — all live.

| File to Create | Purpose |
|---------------|---------|
| `core/migrations/025_trigger_rules.up.sql` | trigger_rules table + deferred FKs |
| `core/migrations/026_trigger_executions.up.sql` | Audit of every evaluation |
| `core/internal/triggers/store.go` | Rule CRUD + in-memory cache |
| `core/internal/triggers/engine.go` | Match → cooldown → recursion → concurrency → fire |
| `core/internal/server/triggers.go` | GET/POST/PUT/DELETE /api/v1/triggers |

Three mandatory guards: cooldown, recursion (max_depth ceiling: 10), concurrency (max_active_runs).

Default trigger mode: `propose` — requires human approval unless explicit `execute` policy.

**Done when:** Rules CRUD works, mission.completed event fires a child run, all guards pass unit tests.

---

### V7 Team C — Scheduler (after A)

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
| `filesystem` MCP | NOT STARTED |
| `memory` MCP | NOT STARTED |
| `artifact-renderer` MCP | NOT STARTED |
| `fetch` MCP | NOT STARTED |

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
| Automations | Trigger Rules | Team B (Trigger Engine) |
| System | Event Health | Team A live data wiring |
| Resources | Workspace Explorer | MCP Baseline (filesystem server) |

---

## Current Build State

```
next build:         PASSES (all routes)
vitest:             ~56 V7 tests pass (2 pre-existing DashboardPage failures, unrelated)
Go build:           go build ./... PASSES
Go tests:           188+ tests pass across 16 packages
                    Migrations 023-029 must be applied for full test coverage
Go test packages:   internal/server (157), internal/events (16), internal/runs (19), others (~80)
```

---

## Architecture Debt (Known Gaps)

| Gap | Location | Priority |
|-----|---------|---------|
| 2 pre-existing DashboardPage test failures | `__tests__/pages/DashboardPage.test.tsx` | Low (pre-V7) |
| 14 pre-existing test file transform errors | Various `__tests__/{workspace,dashboard,teams,...}` | Low (pre-V7) |
| Causal Chain UI | ViewChain.tsx + /runs/[id]/chain | After Team B+C |
| Automations → Active Automations: DegradedState | `app/(app)/automations/page.tsx` | Blocked by Team C |
| Automations → Trigger Rules: DegradedState | `app/(app)/automations/page.tsx` | Blocked by Team B |
| Resources → Workspace Explorer: DegradedState | `app/(app)/resources/page.tsx` | Blocked by MCP Baseline |

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
| Docs API prefix | `/docs-api/` not `/api/docs/` | `/api/*` → Go proxy rewrite would intercept |
| Docs params | `await params` in route handler | Next.js 15+ async params requirement |
| Nav item order | Docs below Memory in main nav | Workflow items stay together; docs is reference |
