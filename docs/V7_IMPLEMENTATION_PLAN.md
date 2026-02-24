# V7 Implementation Plan — Event Spine, Mission Runs, Trigger Engine, Scheduler, Workflow-First IA

> **Status:** In Progress — Team B (Trigger Engine) is NEXT
> **Date:** 2026-02-20 | **Updated:** 2026-02-24
> **Scope:** Migrations 023-029, 6 new Go packages, 8 new API route groups, 10+ new frontend components, navigation restructure
> **Execution Order:** Team A → Team B → Team C → Team E → Team D (strict sequential, no parallel overlap on shared tables)
>
> | Team | Scope | Status |
> |------|-------|--------|
> | Team D | V7 Navigation restructure | ✅ COMPLETE |
> | Team A | Event Spine — migrations 023-024, events, runs, timeline APIs | ✅ COMPLETE |
> | Provider/Profile | Migrations 028-029, provider CRUD, mission profiles, reactive engine, services dashboard | ✅ COMPLETE |
> | Team B | Trigger Engine — migrations 025-026, triggers/store.go, engine.go, handlers | 🔲 NEXT |
> | Team C | Scheduler — migration 027, scheduler.go, schedules handlers | 🔲 PENDING |
> | Team E | Causal Chain UI — ViewChain.tsx, RunChainNode.tsx, /runs/[id]/chain page | 🔲 PENDING |
> | MCP Baseline | filesystem, memory, artifact-renderer, fetch MCP servers | 🔲 PARALLEL |

---

## 1. Architecture Alignment Summary

### Reference Documents Consulted

| Document | Role |
|----------|------|
| `mycelis-architecture-v6.2.md` | Master State Authority — current schema, CTS envelope, NATS topics, Soma lifecycle |
| `mycelis-architecture-v7.md` | V7 PRD — Event Spine, Trigger Engine, Scheduler, Workflow-First IA requirements |
| Phase 19-B Fix Report | Intent proof pipeline corrections, confirm token flow, mutation detection |

### Locked Decisions

1. **Dual-layer events (CTS + MissionEventEnvelope):**
   - CTS = real-time transport layer. Lightweight, ephemeral, NATS-native. Carries `mission_event_id` reference.
   - MissionEventEnvelope = persistent audit record. Authoritative. Written to `mission_events` table FIRST, then CTS published.
   - CTS payload includes `mission_event_id` pointer — not a full duplication of the event envelope.

2. **`run_id` is mandatory:**
   - A mission is a definition (blueprint + teams + agents). A run is an execution instance.
   - All timelines, triggers, and causal chains are anchored to `run_id`, not `mission_id`.
   - No events emitted without a valid `run_id`.

3. **Trigger default: `mode = propose`:**
   - Auto-execute requires explicit policy allowance (`mode = execute`).
   - Human-in-the-loop is the default posture. Triggers that fire with `mode = propose` create a governance proposal.

4. **Single-tenant with multi-tenant schema:**
   - Operationally single-tenant. All tables carry `tenant_id TEXT NOT NULL DEFAULT 'default'`.
   - No tenant isolation logic in V7. Schema readiness only.

5. **In-process scheduler:**
   - Goroutine with 1-minute ticker resolution. Postgres-backed (`scheduled_missions` table).
   - Suspends when NATS goes offline. Resumes on reconnect.
   - No external scheduler dependency (no cron, no Kubernetes CronJob).

### Existing Infrastructure Leveraged

| Component | Location | How V7 Uses It |
|-----------|----------|----------------|
| CTSEnvelope | `core/pkg/protocol/envelopes.go` | Extended with `MissionEventID` field (omitempty, backward-compatible) |
| Soma.ActivateBlueprint | `core/internal/swarm/soma.go` | Creates `mission_run` on activation, passes `run_id` to teams |
| Intent Proofs | Migration 022 (`intent_proofs`) | Linked to `mission_runs` via FK — every committed intent creates a run |
| NATS Topics | `core/pkg/protocol/topics.go` | New constants: `TopicMissionEvents`, `TopicMissionEventsFmt` |
| AdminServer Route Registration | `core/internal/server/admin.go` | New handler groups registered in `RegisterRoutes()` |
| APIResponse Envelope | `core/pkg/protocol/envelopes.go` | All new endpoints use `protocol.APIResponse{OK, Data, Error}` |
| Zustand Store | `interface/store/useCortexStore.ts` | New slices for run timeline, chain, triggers, schedules |
| ShellLayout / ZoneA_Rail | `interface/components/shell/` | Navigation restructure for workflow-first IA |

---

## 2. Detailed Implementation Steps Per Team

### Team A — Event Spine (Backend Core)

Execution order is strict. Each step depends on the prior step completing.

#### Step 1: Write Migration 023 (`mission_runs`) and Migration 024 (`mission_events`)

Create `core/migrations/023_mission_runs.up.sql` and `core/migrations/024_mission_events.up.sql`. Full SQL in Section 4.

Key design choices:
- `mission_runs.triggered_by_rule_id` is initially unlinked (no FK) — the `trigger_rules` table does not exist yet. Team B adds the FK constraint in Migration 025.
- `mission_runs.parent_run_id` self-references for causal chaining.
- `mission_events.event_id` uses explicit naming (not `id`) to distinguish from other UUID PKs in the schema.
- `mission_events.payload` is JSONB — flexible, schema-free, supports arbitrary event-specific data.
- `mission_events.artifact_refs` is JSONB array — links to artifacts table by ID.

Down migrations drop tables in reverse order (024 first, then 023).

#### Step 2: Define `MissionEventEnvelope` struct in `pkg/protocol/events.go`

Create `core/pkg/protocol/events.go` with:

```go
package protocol

import (
    "fmt"
    "time"
)

// EventType classifies mission events flowing through the Event Spine.
type EventType string

const (
    EventMissionStarted      EventType = "mission.started"
    EventMissionCompleted    EventType = "mission.completed"
    EventMissionFailed       EventType = "mission.failed"
    EventMissionPaused       EventType = "mission.paused"
    EventMissionResumed      EventType = "mission.resumed"
    EventTeamSpawned         EventType = "team.spawned"
    EventTeamDissolved       EventType = "team.dissolved"
    EventAgentStarted        EventType = "agent.started"
    EventAgentStopped        EventType = "agent.stopped"
    EventToolInvoked         EventType = "tool.invoked"
    EventToolCompleted       EventType = "tool.completed"
    EventToolFailed          EventType = "tool.failed"
    EventPolicyEvaluated     EventType = "policy.evaluated"
    EventProposalCreated     EventType = "proposal.created"
    EventTriggerRuleEvaluated EventType = "trigger.rule.evaluated"
    EventTriggerFired        EventType = "trigger.fired"
)

// EventSeverity classifies the urgency of a mission event.
type EventSeverity string

const (
    SeverityInfo  EventSeverity = "info"
    SeverityWarn  EventSeverity = "warn"
    SeverityError EventSeverity = "error"
)

// MissionEventEnvelope is the persistent audit record for every significant
// action within a mission run. Persisted to mission_events table FIRST,
// then a lightweight CTS reference is published to NATS.
type MissionEventEnvelope struct {
    EventID        string            `json:"event_id" db:"event_id"`
    TenantID       string            `json:"tenant_id" db:"tenant_id"`
    MissionID      string            `json:"mission_id" db:"mission_id"`
    RunID          string            `json:"run_id" db:"run_id"`
    EventType      EventType         `json:"event_type" db:"event_type"`
    Severity       EventSeverity     `json:"severity" db:"severity"`
    ProviderID     string            `json:"provider_id,omitempty" db:"provider_id"`
    ModelUsed      string            `json:"model_used,omitempty" db:"model_used"`
    Role           string            `json:"role,omitempty" db:"role"`
    Mode           string            `json:"mode,omitempty" db:"mode"`
    Payload        map[string]any    `json:"payload,omitempty" db:"payload"`
    ArtifactRefs   []string          `json:"artifact_refs,omitempty" db:"artifact_refs"`
    ParentEventID  string            `json:"parent_event_id,omitempty" db:"parent_event_id"`
    ParentRunID    string            `json:"parent_run_id,omitempty" db:"parent_run_id"`
    TriggerRuleID  string            `json:"trigger_rule_id,omitempty" db:"trigger_rule_id"`
    AuditEventID   string            `json:"audit_event_id,omitempty" db:"audit_event_id"`
    IntentProofID  string            `json:"intent_proof_id,omitempty" db:"intent_proof_id"`
    CreatedAt      time.Time         `json:"created_at" db:"created_at"`
}

// Validate enforces required fields on the event envelope.
func (e *MissionEventEnvelope) Validate() error {
    if e.MissionID == "" {
        return fmt.Errorf("event: mission_id is required")
    }
    if e.RunID == "" {
        return fmt.Errorf("event: run_id is required")
    }
    if e.EventType == "" {
        return fmt.Errorf("event: event_type is required")
    }
    if e.TenantID == "" {
        e.TenantID = "default"
    }
    if e.Severity == "" {
        e.Severity = SeverityInfo
    }
    return nil
}
```

All 17 fields from the directive are present. Validation defaults `tenant_id` and `severity`.

#### Step 3: Create Event Store (`internal/events/store.go`)

Create `core/internal/events/store.go` with the following interface:

```go
type Store struct {
    db *sql.DB
    nc *nats.Conn
}

func NewStore(db *sql.DB, nc *nats.Conn) *Store

// Emit persists a MissionEventEnvelope to the mission_events table.
// Returns the assigned event_id (UUID generated by Postgres).
func (s *Store) Emit(ctx context.Context, event *protocol.MissionEventEnvelope) (string, error)

// EmitAndPublish persists the event, then publishes a lightweight CTS
// reference to NATS. If NATS is unavailable, the event is still persisted
// and a degraded flag is logged. No data loss.
func (s *Store) EmitAndPublish(ctx context.Context, event *protocol.MissionEventEnvelope) (string, error)

// GetRunTimeline returns all events for a given run_id, ordered by created_at ASC.
func (s *Store) GetRunTimeline(ctx context.Context, runID string) ([]protocol.MissionEventEnvelope, error)

// GetChain returns the causal chain for a run: walks parent_run_id recursively
// up to the root run, then returns the full tree of runs with event counts.
func (s *Store) GetChain(ctx context.Context, runID string) (*RunChainNode, error)
```

`RunChainNode` struct:

```go
type RunChainNode struct {
    RunID            string          `json:"run_id"`
    MissionID        string          `json:"mission_id"`
    MissionName      string          `json:"mission_name"`
    Status           string          `json:"status"`
    Depth            int             `json:"depth"`
    StartedAt        *time.Time      `json:"started_at,omitempty"`
    CompletedAt      *time.Time      `json:"completed_at,omitempty"`
    TriggeredByRuleID *string        `json:"triggered_by_rule_id,omitempty"`
    EventsCount      int             `json:"events_count"`
    Children         []*RunChainNode `json:"children,omitempty"`
}
```

Implementation notes:
- `Emit()` uses a single `INSERT INTO mission_events ... RETURNING event_id` query.
- `EmitAndPublish()` calls `Emit()`, then marshals a CTS envelope with `MissionEventID` set to the returned event_id, and publishes to `TopicMissionEventsFmt`.
- `GetRunTimeline()` is a simple `SELECT * FROM mission_events WHERE run_id = $1 ORDER BY created_at ASC`.
- `GetChain()` uses a recursive CTE:
  ```sql
  WITH RECURSIVE chain AS (
      SELECT id, mission_id, parent_run_id, depth, status, ...
      FROM mission_runs WHERE id = $1
      UNION ALL
      SELECT mr.id, mr.mission_id, mr.parent_run_id, mr.depth, mr.status, ...
      FROM mission_runs mr
      JOIN chain c ON mr.parent_run_id = c.id
  )
  SELECT ... FROM chain
  ```
  Then queries event counts per run and assembles the tree.

#### Step 4: Create Mission Run Manager (`internal/runs/manager.go`)

Create `core/internal/runs/manager.go`:

```go
type Manager struct {
    db *sql.DB
}

func NewManager(db *sql.DB) *Manager

// CreateRun inserts a new mission_run record and returns the generated run_id.
// Parameters:
//   - missionID: the mission definition being executed
//   - intentProofID: optional, links to the intent proof that authorized this run
//   - triggeredByRuleID: optional, links to the trigger rule that created this run
//   - parentRunID: optional, links to the parent run in a causal chain
//   - depth: recursion depth (0 for top-level runs)
func (m *Manager) CreateRun(ctx context.Context, missionID, intentProofID, triggeredByRuleID, parentRunID string, depth int) (string, error)

// UpdateRunStatus transitions a run to a new status (running, completed, failed).
// Sets started_at on transition to "running", completed_at on "completed"/"failed".
func (m *Manager) UpdateRunStatus(ctx context.Context, runID, status string) error

// GetRun retrieves a single mission run by ID.
func (m *Manager) GetRun(ctx context.Context, runID string) (*MissionRun, error)

// ListRunsForMission returns all runs for a given mission_id, ordered by created_at DESC.
func (m *Manager) ListRunsForMission(ctx context.Context, missionID string) ([]MissionRun, error)
```

`MissionRun` struct:

```go
type MissionRun struct {
    ID               string     `json:"id" db:"id"`
    MissionID        string     `json:"mission_id" db:"mission_id"`
    TenantID         string     `json:"tenant_id" db:"tenant_id"`
    IntentProofID    *string    `json:"intent_proof_id,omitempty" db:"intent_proof_id"`
    TriggeredByRuleID *string   `json:"triggered_by_rule_id,omitempty" db:"triggered_by_rule_id"`
    ParentRunID      *string    `json:"parent_run_id,omitempty" db:"parent_run_id"`
    Depth            int        `json:"depth" db:"depth"`
    Status           string     `json:"status" db:"status"`
    StartedAt        *time.Time `json:"started_at,omitempty" db:"started_at"`
    CompletedAt      *time.Time `json:"completed_at,omitempty" db:"completed_at"`
    CreatedAt        time.Time  `json:"created_at" db:"created_at"`
}
```

#### Step 5: Wire Event Emission into Existing Execution Paths

This step modifies existing files. Each integration point:

**5a. `internal/swarm/soma.go` — `ActivateBlueprint()`**
- After successfully activating a blueprint (creating teams, spawning agents):
  - Call `runs.Manager.CreateRun(missionID, intentProofID, "", "", 0)` to get a `run_id`.
  - Call `runs.Manager.UpdateRunStatus(runID, "running")`.
  - Call `events.Store.EmitAndPublish()` with `EventMissionStarted` event.
  - Pass `run_id` into each team's context so all downstream events reference it.
- Soma needs a reference to both `runs.Manager` and `events.Store` — added as fields.

**5b. `internal/swarm/agent.go` — `processMessageStructured()`**
- Before tool invocation: emit `EventToolInvoked` with tool name, input params in payload.
- After tool completion: emit `EventToolCompleted` with tool name, result summary, duration.
- On tool failure: emit `EventToolFailed` with error details, severity `error`.
- Agent needs access to `events.Store` and `run_id` — passed through `Team` -> `Agent` context.

**5c. Mission completion**
- When all teams report completion or the Overseer marks a mission done:
  - Emit `EventMissionCompleted`.
  - Call `runs.Manager.UpdateRunStatus(runID, "completed")`.

**5d. Governance decisions**
- When the Guard evaluates a policy (approve/deny):
  - Emit `EventPolicyEvaluated` with decision, risk level, policy rule in payload.

#### Step 6: Extend CTSEnvelope with `MissionEventID` Field

Modify `core/pkg/protocol/envelopes.go`:

```go
type CTSEnvelope struct {
    Meta           CTSMeta         `json:"meta"`
    SignalType     SignalType      `json:"signal_type"`
    TrustScore     float64         `json:"trust_score,omitempty"`
    Payload        json.RawMessage `json:"payload"`
    TemplateID     TemplateID      `json:"template_id,omitempty"`
    Mode           ExecutionMode   `json:"mode,omitempty"`
    MissionEventID string          `json:"mission_event_id,omitempty"` // V7: reference to persisted event
}
```

The field is `omitempty` — existing CTS consumers are unaffected. Only events emitted through the Event Spine carry this reference.

#### Step 7: Add NATS Topics for Events

Modify `core/pkg/protocol/topics.go`:

```go
// V7: Mission Event Spine
TopicMissionEvents    = "swarm.mission.events"             // all mission events
TopicMissionEventsFmt = "swarm.mission.%s.events"          // per-mission events (mission ID)
TopicMissionRunsFmt   = "swarm.mission.%s.runs.%s.events"  // per-run events (mission ID, run ID)
```

#### Step 8: Register API Routes

Add to `core/internal/server/admin.go` in `RegisterRoutes()`:

```go
// V7: Mission Run Timeline & Causal Chain
mux.HandleFunc("GET /api/v1/runs/{run_id}/events", s.handleGetRunTimeline)
mux.HandleFunc("GET /api/v1/runs/{run_id}/chain", s.handleGetRunChain)
```

Create `core/internal/server/runs.go` with handler implementations:

```go
// handleGetRunTimeline returns all events for a given run, ordered chronologically.
// Response: { ok: true, data: MissionEventEnvelope[] }
func (s *AdminServer) handleGetRunTimeline(w http.ResponseWriter, r *http.Request)

// handleGetRunChain returns the causal chain tree for a given run.
// Response: { ok: true, data: RunChainNode }
func (s *AdminServer) handleGetRunChain(w http.ResponseWriter, r *http.Request)
```

Both use `r.PathValue("run_id")` (Go 1.22+ path params).

#### Step 9: Degraded Mode

If NATS is unavailable at the time of event emission:
- `EmitAndPublish()` persists the event to Postgres (this always succeeds if DB is up).
- The CTS publish step is skipped. A warning is logged: `"event persisted in degraded mode (NATS offline): event_id=%s"`.
- A `degraded` field is NOT stored on the event — the absence of a CTS record is the indicator.
- When NATS reconnects, events are NOT replayed (they are queryable via the timeline API).

---

### Team B — Trigger Rules Engine

**Depends on Team A:** Uses `mission_events` and `mission_runs` tables, `events.Store`, `runs.Manager`.

#### Step 1: Write Migration 025 (`trigger_rules`) and Migration 026 (`trigger_executions`)

Create `core/migrations/025_trigger_rules.up.sql` and `core/migrations/026_trigger_executions.up.sql`. Full SQL in Section 4.

Key design choices:
- `trigger_rules.source_mission_id` is nullable — a rule can match events from ANY mission (global trigger).
- `trigger_rules.predicate` is JSONB — flexible predicate expressions (e.g., `{"status": "completed", "payload.risk_level": "low"}`).
- `trigger_rules.mode` defaults to `propose` — safety-first.
- Migration 025 also adds the deferred FK from `mission_runs.triggered_by_rule_id` to `trigger_rules.trigger_rule_id`.
- `trigger_executions` tracks every evaluation result (evaluated, fired, skipped) for auditability.

#### Step 2: Create Trigger Store (`internal/triggers/store.go`)

```go
type Store struct {
    db    *sql.DB
    cache []TriggerRule  // in-memory cache for fast evaluation
    mu    sync.RWMutex
}

func NewStore(db *sql.DB) *Store

// LoadRules loads all enabled trigger rules into memory cache.
// Called on startup and after any CRUD mutation.
func (s *Store) LoadRules(ctx context.Context) error

// CreateRule persists a new trigger rule and refreshes cache.
func (s *Store) CreateRule(ctx context.Context, rule *TriggerRule) (string, error)

// UpdateRule updates an existing rule and refreshes cache.
func (s *Store) UpdateRule(ctx context.Context, rule *TriggerRule) error

// DeleteRule removes a rule and refreshes cache.
func (s *Store) DeleteRule(ctx context.Context, ruleID string) error

// GetRule retrieves a single rule by ID.
func (s *Store) GetRule(ctx context.Context, ruleID string) (*TriggerRule, error)

// ListRules returns all rules (including disabled) for API listing.
func (s *Store) ListRules(ctx context.Context) ([]TriggerRule, error)

// RecordExecution logs a trigger evaluation result.
func (s *Store) RecordExecution(ctx context.Context, exec *TriggerExecution) error
```

`TriggerRule` struct:

```go
type TriggerRule struct {
    TriggerRuleID      string         `json:"trigger_rule_id" db:"trigger_rule_id"`
    TenantID           string         `json:"tenant_id" db:"tenant_id"`
    Enabled            bool           `json:"enabled" db:"enabled"`
    SourceMissionID    *string        `json:"source_mission_id,omitempty" db:"source_mission_id"`
    EventType          string         `json:"event_type" db:"event_type"`
    Predicate          map[string]any `json:"predicate,omitempty" db:"predicate"`
    ActionTargetMissionID string      `json:"action_target_mission_id" db:"action_target_mission_id"`
    Mode               string         `json:"mode" db:"mode"`
    RequiresApproval   bool           `json:"requires_approval" db:"requires_approval"`
    CooldownMs         int            `json:"cooldown_ms" db:"cooldown_ms"`
    MaxDepth           int            `json:"max_depth" db:"max_depth"`
    MaxActiveRuns      int            `json:"max_active_runs" db:"max_active_runs"`
    CreatedAt          time.Time      `json:"created_at" db:"created_at"`
}
```

`TriggerExecution` struct:

```go
type TriggerExecution struct {
    ID             string    `json:"id" db:"id"`
    TriggerRuleID  string    `json:"trigger_rule_id" db:"trigger_rule_id"`
    SourceEventID  string    `json:"source_event_id" db:"source_event_id"`
    SourceRunID    string    `json:"source_run_id" db:"source_run_id"`
    TargetRunID    *string   `json:"target_run_id,omitempty" db:"target_run_id"`
    Result         string    `json:"result" db:"result"`
    Reason         string    `json:"reason,omitempty" db:"reason"`
    CreatedAt      time.Time `json:"created_at" db:"created_at"`
}
```

#### Step 3: Create Trigger Engine (`internal/triggers/engine.go`)

```go
type Engine struct {
    store      *Store
    runMgr     *runs.Manager
    eventStore *events.Store
}

func NewEngine(store *Store, runMgr *runs.Manager, eventStore *events.Store) *Engine

// Evaluate checks all enabled rules against a newly emitted event.
// Returns the list of rules that matched and their evaluation results.
func (e *Engine) Evaluate(ctx context.Context, event *protocol.MissionEventEnvelope) ([]EvalResult, error)

// checkCooldown returns true if the rule is within its cooldown period.
// Looks at the most recent trigger_execution for this rule.
func (e *Engine) checkCooldown(ctx context.Context, rule *TriggerRule, now time.Time) bool

// checkRecursionGuard returns true if the event's run depth exceeds the rule's max_depth.
func (e *Engine) checkRecursionGuard(rule *TriggerRule, depth int) bool

// checkConcurrencyGuard returns true if the target mission already has
// max_active_runs running (status = 'running' or 'pending').
func (e *Engine) checkConcurrencyGuard(ctx context.Context, rule *TriggerRule) bool

// Fire creates a downstream mission_run linked to the triggering event and rule.
// If mode=propose, creates a governance proposal instead of executing directly.
func (e *Engine) Fire(ctx context.Context, rule *TriggerRule, sourceEvent *protocol.MissionEventEnvelope) (*runs.MissionRun, error)
```

`EvalResult` struct:

```go
type EvalResult struct {
    Rule       TriggerRule `json:"rule"`
    Matched    bool        `json:"matched"`
    Fired      bool        `json:"fired"`
    SkipReason string      `json:"skip_reason,omitempty"` // cooldown, recursion_guard, concurrency_guard
}
```

Evaluation logic (in order):
1. Filter cached rules to those matching the event's `event_type`.
2. If rule has `source_mission_id`, check it matches the event's `mission_id`. Skip if mismatch.
3. If rule has `predicate`, evaluate JSONB predicate against event payload. Skip if mismatch.
4. Check cooldown — query most recent `trigger_execution` for this rule. Skip if within cooldown window.
5. Check recursion guard — compare event's run depth against `rule.max_depth`. Skip if exceeded.
6. Check concurrency guard — count active runs (`pending`/`running`) for target mission. Skip if at limit.
7. If all guards pass: call `Fire()`.

`Fire()` implementation:
- If `mode = propose`: create governance proposal with mission activation payload. Emit `EventProposalCreated`.
- If `mode = execute`: call `runs.Manager.CreateRun()` with `triggeredByRuleID` and `parentRunID` set. Emit `EventTriggerFired`. Call `Soma.ActivateBlueprint()` for the target mission.
- Always record a `TriggerExecution` with the result.

#### Step 4: Wire Engine to Event Ingest

After every call to `events.Store.Emit()` or `events.Store.EmitAndPublish()`:
- Call `triggers.Engine.Evaluate(ctx, event)`.
- For each matching rule that fires, the engine handles creating the downstream run.
- Emit `EventTriggerRuleEvaluated` for each evaluation (audit trail).
- Emit `EventTriggerFired` for each successful firing.

This wiring is done inside `events.Store.EmitAndPublish()` — the store holds a reference to the trigger engine (set after initialization to break circular dependency):

```go
func (s *Store) SetTriggerEngine(engine *triggers.Engine)
```

#### Step 5: Register Trigger CRUD APIs

Add to `core/internal/server/admin.go` in `RegisterRoutes()`:

```go
// V7: Trigger Rules CRUD
mux.HandleFunc("GET /api/v1/triggers", s.handleListTriggers)
mux.HandleFunc("POST /api/v1/triggers", s.handleCreateTrigger)
mux.HandleFunc("PUT /api/v1/triggers/{id}", s.handleUpdateTrigger)
mux.HandleFunc("DELETE /api/v1/triggers/{id}", s.handleDeleteTrigger)
```

Create `core/internal/server/triggers.go` with handler implementations.

---

### Team C — Scheduler

**Depends on Team A:** Uses `mission_runs` table and `runs.Manager`.

#### Step 1: Write Migration 027 (`scheduled_missions`)

Create `core/migrations/027_scheduled_missions.up.sql`. Full SQL in Section 4.

Key design choices:
- `cron_expression` stored as text — parsed at runtime using a Go cron library.
- `next_run_at` is precomputed and indexed — the scheduler tick only queries `WHERE next_run_at <= NOW() AND enabled = true`.
- `max_active_runs` defaults to 1 — prevents schedule pile-up.

#### Step 2: Create Scheduler (`internal/scheduler/scheduler.go`)

```go
type Scheduler struct {
    db        *sql.DB
    runMgr    *runs.Manager
    eventStore *events.Store
    soma      *swarm.Soma
    nc        *nats.Conn
    schedules []ScheduledMission  // in-memory cache
    mu        sync.RWMutex
    ticker    *time.Ticker
    ctx       context.Context
    cancel    context.CancelFunc
    suspended bool
}

func NewScheduler(db *sql.DB, runMgr *runs.Manager, eventStore *events.Store, soma *swarm.Soma, nc *nats.Conn) *Scheduler

// LoadSchedules reads all enabled schedules from DB into memory.
func (s *Scheduler) LoadSchedules(ctx context.Context) error

// Start begins the scheduler goroutine with a 1-minute ticker.
func (s *Scheduler) Start()

// Stop gracefully shuts down the scheduler.
func (s *Scheduler) Stop()

// Suspend pauses scheduling (called on NATS disconnect).
func (s *Scheduler) Suspend()

// Resume resumes scheduling (called on NATS reconnect).
func (s *Scheduler) Resume()

// checkDue finds all schedules where next_run_at <= now.
func (s *Scheduler) checkDue(now time.Time) []ScheduledMission

// createScheduledRun creates a mission_run for a due schedule.
func (s *Scheduler) createScheduledRun(ctx context.Context, sched *ScheduledMission) error

// enforceConcurrency checks if the mission already has max_active_runs.
func (s *Scheduler) enforceConcurrency(ctx context.Context, sched *ScheduledMission) bool
```

`ScheduledMission` struct:

```go
type ScheduledMission struct {
    ID             string     `json:"id" db:"id"`
    MissionID      string     `json:"mission_id" db:"mission_id"`
    TenantID       string     `json:"tenant_id" db:"tenant_id"`
    CronExpression string     `json:"cron_expression" db:"cron_expression"`
    Enabled        bool       `json:"enabled" db:"enabled"`
    MaxActiveRuns  int        `json:"max_active_runs" db:"max_active_runs"`
    LastRunAt      *time.Time `json:"last_run_at,omitempty" db:"last_run_at"`
    NextRunAt      *time.Time `json:"next_run_at,omitempty" db:"next_run_at"`
    CreatedAt      time.Time  `json:"created_at" db:"created_at"`
}
```

Tick logic:
1. Every 60 seconds, call `checkDue(time.Now())`.
2. For each due schedule:
   a. Check `enforceConcurrency()` — skip if at limit.
   b. Call `createScheduledRun()` — creates `mission_run`, updates `last_run_at`, computes `next_run_at`.
   c. Call `Soma.ActivateBlueprint()` for the mission.
   d. Emit `EventMissionStarted` via event store.
3. Log skipped schedules with reason.

NATS suspension:
- On NATS disconnect callback: call `scheduler.Suspend()`. Log: `"scheduler suspended: NATS offline"`.
- On NATS reconnect callback: call `scheduler.Resume()`, call `LoadSchedules()` to refresh. Log: `"scheduler resumed: NATS reconnected"`.

#### Step 3: Register Scheduler APIs

Add to `core/internal/server/admin.go` in `RegisterRoutes()`:

```go
// V7: Scheduled Missions CRUD
mux.HandleFunc("GET /api/v1/schedules", s.handleListSchedules)
mux.HandleFunc("POST /api/v1/schedules", s.handleCreateSchedule)
mux.HandleFunc("PUT /api/v1/schedules/{id}", s.handleUpdateSchedule)
mux.HandleFunc("DELETE /api/v1/schedules/{id}", s.handleDeleteSchedule)
mux.HandleFunc("POST /api/v1/schedules/{id}/pause", s.handlePauseSchedule)
mux.HandleFunc("POST /api/v1/schedules/{id}/resume", s.handleResumeSchedule)
```

Create `core/internal/server/schedules.go` with handler implementations.

#### Step 4: Wire into Server Startup

In `core/cmd/server/main.go`:
- Initialize `runs.Manager` and `events.Store` (Team A prereq).
- Initialize `scheduler.Scheduler` after Soma is created but before HTTP listen.
- Call `scheduler.LoadSchedules(ctx)` and `scheduler.Start()`.
- Register NATS disconnect/reconnect callbacks:
  ```go
  nc.SetDisconnectErrHandler(func(nc *nats.Conn, err error) {
      scheduler.Suspend()
  })
  nc.SetReconnectHandler(func(nc *nats.Conn) {
      scheduler.Resume()
  })
  ```
- In graceful shutdown: call `scheduler.Stop()` before NATS disconnect.

---

### Team E — Run Timeline + Chain UI (Frontend)

**Depends on Team A APIs** being available (`GET /api/v1/runs/{run_id}/events`, `GET /api/v1/runs/{run_id}/chain`).

#### Step 1: Create `RunTimeline` Component

Create `interface/components/runs/RunTimeline.tsx`:

- Fetches `GET /api/v1/runs/{run_id}/events` on mount.
- Renders a vertical timeline using a `<ul>` with connecting line (CSS border-left).
- Each event is an `EventCard` component.
- Color-coding by event type:
  - `mission.*` = `cortex-primary` (cyan #06b6d4)
  - `tool.*` = `cortex-success` (green #10b981)
  - `trigger.*` = amber (#f59e0b)
  - `*.failed` / severity `error` = red (#ef4444)
  - `policy.*` = purple (#8b5cf6)
- Loading state: skeleton cards with cortex-surface background.
- Empty state: informative message "No events recorded for this run."

Create `interface/components/runs/EventCard.tsx`:

- Collapsed view: timestamp (relative), event_type badge, severity indicator, provider/model if present.
- Expanded view (on click): full payload JSON (syntax highlighted), artifact links (clickable), audit event link.
- Uses cortex theme tokens throughout. `bg-cortex-surface`, `border-cortex-border`, `text-cortex-text-main`.
- Expand/collapse with smooth height transition.

#### Step 2: Create `ViewChain` Component

Create `interface/components/runs/ViewChain.tsx`:

- Fetches `GET /api/v1/runs/{run_id}/chain` on mount.
- Renders a nested tree structure. Each node is a `RunChainNode` component.
- Tree shows: parent run -> event that triggered -> trigger rule -> child run.
- Indentation increases with depth. Max visual depth of 10 (matches hard ceiling).

Create `interface/components/runs/RunChainNode.tsx`:

- Displays: mission name, status badge, depth indicator, started_at, events_count.
- Status badge colors: pending=gray, running=cyan, completed=green, failed=red.
- Click navigates to that run's timeline (`/runs/{id}`).
- Children rendered recursively.

#### Step 3: Add Zustand Slices

Modify `interface/store/useCortexStore.ts` — add run timeline slice:

```typescript
// V7: Run Timeline State
activeRunId: string | null;
runTimeline: MissionEventEnvelope[];
runChain: RunChainNode | null;
isTimelineLoading: boolean;
isChainLoading: boolean;

// V7: Run Timeline Actions
setActiveRunId: (runId: string | null) => void;
fetchRunTimeline: (runId: string) => Promise<void>;
fetchRunChain: (runId: string) => Promise<void>;
clearRunTimeline: () => void;
```

Types defined in `interface/types/events.ts`:

```typescript
interface MissionEventEnvelope {
    event_id: string;
    tenant_id: string;
    mission_id: string;
    run_id: string;
    event_type: string;
    severity: 'info' | 'warn' | 'error';
    provider_id?: string;
    model_used?: string;
    role?: string;
    mode?: string;
    payload?: Record<string, any>;
    artifact_refs?: string[];
    parent_event_id?: string;
    parent_run_id?: string;
    trigger_rule_id?: string;
    audit_event_id?: string;
    intent_proof_id?: string;
    created_at: string;
}

interface RunChainNode {
    run_id: string;
    mission_id: string;
    mission_name: string;
    status: string;
    depth: number;
    started_at?: string;
    completed_at?: string;
    triggered_by_rule_id?: string;
    events_count: number;
    children?: RunChainNode[];
}
```

#### Step 4: Wire into Mission Control

Modify `interface/components/dashboard/MissionControl.tsx`:
- After mission activation, set `activeRunId` in store.
- Add "View Timeline" button on active mission cards — opens timeline panel or navigates to `/runs/{id}`.
- Add "View Chain" button on triggered missions — navigates to `/runs/{id}/chain`.

Create route pages:
- `interface/app/(app)/runs/[id]/page.tsx` — renders `RunTimeline` for the given run ID.
- `interface/app/(app)/runs/[id]/chain/page.tsx` — renders `ViewChain` for the given run ID.

#### Step 5: Wire into Automations Page

Once Team D creates the Automations page, integrate:
- Active Automations tab shows run timelines inline.
- Trigger Rules tab links to chain views for triggered runs.

---

### Team D — Workflow-First IA Restructure (Frontend)

**Can proceed independently** once Team E wiring points are known. Last team in execution order.

#### Step 1: Create New Navigation Structure in ShellLayout

Modify `interface/components/shell/ShellLayout.tsx` and `interface/components/shell/ZoneA_Rail.tsx`:

5 primary navigation items:

| Nav Item | Route | Icon | Description |
|----------|-------|------|-------------|
| Mission Control | `/dashboard` | Command icon | Chat, proposals, confirms, execution reports, run timelines |
| Automations | `/automations` | Workflow icon | Scheduled missions, trigger rules, drafts, approvals |
| Resources | `/resources` | Database icon | Brains, tools, catalogue, service targets |
| Memory | `/memory` | Brain icon | Semantic search, sources, recall/store events |
| System | `/system` | Terminal icon | Event health, core status, NATS, DB (Advanced toggle) |

System is hidden behind an "Advanced" toggle (stored in localStorage). Default: hidden.

#### Step 2: Create Automations Page (`/automations`)

Create `interface/app/(app)/automations/page.tsx` and `interface/components/automations/AutomationsPage.tsx`:

Tabs:
- **Active Automations**: List scheduled missions with status, next_run_at, last_run_at, enabled toggle. Fetches `GET /api/v1/schedules`.
- **Trigger Rules**: List rules with enabled/disabled toggle, source mission, target mission, event type, mode badge. Fetches `GET /api/v1/triggers`.
- **Drafts**: Existing blueprint drafts (moved from current location). Fetches from Zustand store.
- **Approvals**: Governance queue (content moved from `/approvals` route). Fetches `GET /api/v1/governance/pending`.

Neural Wiring accessible as a sub-view — "Edit Wiring" button on drafts opens the CircuitBoard.

#### Step 3: Create Resources Page (`/resources`)

Create `interface/app/(app)/resources/page.tsx` and `interface/components/resources/ResourcesPage.tsx`:

Tabs:
- **Brains**: Provider management UI (moved from `/settings/brains`). Fetches `GET /api/v1/brains`.
- **Tools**: Internal + MCP tool browser (moved from `/settings/tools`). Fetches `GET /api/v1/mcp/tools`.
- **Catalogue**: Agent catalogue (moved from `/catalogue`). Fetches `GET /api/v1/catalogue/agents`.
- **Service Targets**: Capabilities list — what the system can do (read from cognitive config).

#### Step 4: Migrate `/dashboard` to Mission Control

Modify `interface/app/(app)/dashboard/page.tsx`:
- Add Run Timeline integration — inline timeline for active runs.
- Add Causal Chain View — "View Chain" button on missions with triggered child runs.
- Enhance with execution reports — summary cards showing runs/hour, success rate, avg duration.

#### Step 5: Update `/memory` Route

Modify `interface/app/(app)/memory/page.tsx`:
- Add recall/store event listing — events of type `tool.invoked` where tool is `remember`/`recall`.
- Link memory usage to run timelines — click an event to navigate to its run timeline.

#### Step 6: Create System Page (`/system`)

Create `interface/app/(app)/system/page.tsx` and `interface/components/system/SystemPage.tsx`:

Sections:
- **Event Health**: NATS connected/disconnected indicator, events/sec counter (from SSE stream), trigger queue depth.
- **Core Status**: Server uptime, API response time, DB connection pool stats.
- **NATS Status**: Connection state, reconnect count, subscription count.
- **DB Status**: Migration version, row counts for key tables, connection pool utilization.
- **Degraded Mode Notices**: Active degradation indicators with explanations.

Hidden behind "Advanced" toggle in nav. Accessible directly via `/system` URL.

#### Step 7: Degraded Mode Messaging

Replace all blank error states across the app with informative degraded messages:

Template:
```
[Icon] [Service] is currently unavailable

Why: [Specific reason — e.g., "NATS connection lost", "Core server offline"]
What's unavailable: [List of affected features]
What still works: [List of features that remain functional]
How to restore: [Action — e.g., "Restart Core server", "Check NATS connection"]
```

Apply to:
- MissionControl when core is offline
- Automations when scheduler is suspended
- Resources when brains are unreachable
- RunTimeline when no events are available
- ViewChain when chain query fails

---

## 3. API Route Definitions

### Event Spine APIs (Team A)

```
GET  /api/v1/runs/{run_id}/events    -> handleGetRunTimeline
     Auth: MYCELIS_API_KEY header
     Response: {
         "ok": true,
         "data": [
             {
                 "event_id": "uuid",
                 "tenant_id": "default",
                 "mission_id": "uuid",
                 "run_id": "uuid",
                 "event_type": "mission.started",
                 "severity": "info",
                 "provider_id": "ollama",
                 "model_used": "qwen2.5-coder:7b-instruct",
                 "role": "architect",
                 "mode": "execution",
                 "payload": { ... },
                 "artifact_refs": ["uuid1", "uuid2"],
                 "parent_event_id": null,
                 "parent_run_id": null,
                 "trigger_rule_id": null,
                 "audit_event_id": null,
                 "intent_proof_id": "uuid",
                 "created_at": "2026-02-20T12:00:00Z"
             }
         ]
     }
     Errors: 400 (invalid run_id), 404 (run not found)

GET  /api/v1/runs/{run_id}/chain     -> handleGetRunChain
     Auth: MYCELIS_API_KEY header
     Response: {
         "ok": true,
         "data": {
             "run_id": "uuid",
             "mission_id": "uuid",
             "mission_name": "Weather Monitor",
             "status": "completed",
             "depth": 0,
             "started_at": "2026-02-20T12:00:00Z",
             "completed_at": "2026-02-20T12:05:00Z",
             "triggered_by_rule_id": null,
             "events_count": 15,
             "children": [
                 {
                     "run_id": "uuid-child",
                     "mission_id": "uuid-report",
                     "mission_name": "Generate Report",
                     "status": "running",
                     "depth": 1,
                     "started_at": "2026-02-20T12:05:01Z",
                     "completed_at": null,
                     "triggered_by_rule_id": "uuid-rule",
                     "events_count": 3,
                     "children": []
                 }
             ]
         }
     }
     Errors: 400 (invalid run_id), 404 (run not found)
```

### Trigger APIs (Team B)

```
GET    /api/v1/triggers              -> handleListTriggers
       Auth: MYCELIS_API_KEY header
       Response: { "ok": true, "data": TriggerRule[] }

POST   /api/v1/triggers              -> handleCreateTrigger
       Auth: MYCELIS_API_KEY header
       Body: {
           "source_mission_id": "uuid | null",
           "event_type": "mission.completed",
           "predicate": { "payload.status": "success" },
           "action_target_mission_id": "uuid",
           "mode": "propose",
           "requires_approval": true,
           "cooldown_ms": 60000,
           "max_depth": 5,
           "max_active_runs": 1
       }
       Response: { "ok": true, "data": TriggerRule }
       Errors: 400 (validation), 404 (target mission not found)

PUT    /api/v1/triggers/{id}         -> handleUpdateTrigger
       Auth: MYCELIS_API_KEY header
       Body: { ... partial TriggerRule fields ... }
       Response: { "ok": true, "data": TriggerRule }
       Errors: 400 (validation), 404 (rule not found)

DELETE /api/v1/triggers/{id}         -> handleDeleteTrigger
       Auth: MYCELIS_API_KEY header
       Response: { "ok": true }
       Errors: 404 (rule not found)
```

### Scheduler APIs (Team C)

```
GET    /api/v1/schedules             -> handleListSchedules
       Auth: MYCELIS_API_KEY header
       Response: { "ok": true, "data": ScheduledMission[] }

POST   /api/v1/schedules             -> handleCreateSchedule
       Auth: MYCELIS_API_KEY header
       Body: {
           "mission_id": "uuid",
           "cron_expression": "0 */6 * * *",
           "max_active_runs": 1
       }
       Response: { "ok": true, "data": ScheduledMission }
       Errors: 400 (invalid cron expression), 404 (mission not found)

PUT    /api/v1/schedules/{id}        -> handleUpdateSchedule
       Auth: MYCELIS_API_KEY header
       Body: { ... partial ScheduledMission fields ... }
       Response: { "ok": true, "data": ScheduledMission }
       Errors: 400, 404

DELETE /api/v1/schedules/{id}        -> handleDeleteSchedule
       Auth: MYCELIS_API_KEY header
       Response: { "ok": true }
       Errors: 404

POST   /api/v1/schedules/{id}/pause  -> handlePauseSchedule
       Auth: MYCELIS_API_KEY header
       Response: { "ok": true, "data": ScheduledMission }
       Errors: 404

POST   /api/v1/schedules/{id}/resume -> handleResumeSchedule
       Auth: MYCELIS_API_KEY header
       Response: { "ok": true, "data": ScheduledMission }
       Errors: 404
```

---

## 4. Data Model SQL

All migrations follow the project's existing pattern:
- UUID PRIMARY KEY DEFAULT `gen_random_uuid()`
- TIMESTAMPTZ DEFAULT `NOW()`
- Foreign keys with appropriate cascade behavior
- Indexes on lookup and filter columns
- Down migrations provided for rollback

### Migration 023 — `mission_runs`

**Up (`core/migrations/023_mission_runs.up.sql`):**

```sql
-- Migration 023: Mission Runs (V7 Event Spine)
--
-- A mission is a definition (blueprint + teams + agents).
-- A run is an execution instance of that definition.
-- All events, triggers, and causal chains are anchored to run_id.

CREATE TABLE IF NOT EXISTS mission_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id UUID NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL DEFAULT 'default',
    intent_proof_id UUID REFERENCES intent_proofs(id) ON DELETE SET NULL,
    triggered_by_rule_id UUID,  -- FK added in migration 025 after trigger_rules exists
    parent_run_id UUID REFERENCES mission_runs(id) ON DELETE SET NULL,
    depth INT NOT NULL DEFAULT 0,
    status TEXT NOT NULL DEFAULT 'pending',  -- pending, running, completed, failed
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Lookup indexes
CREATE INDEX IF NOT EXISTS idx_mission_runs_mission ON mission_runs(mission_id);
CREATE INDEX IF NOT EXISTS idx_mission_runs_tenant ON mission_runs(tenant_id);
CREATE INDEX IF NOT EXISTS idx_mission_runs_status ON mission_runs(status);
CREATE INDEX IF NOT EXISTS idx_mission_runs_parent ON mission_runs(parent_run_id);
CREATE INDEX IF NOT EXISTS idx_mission_runs_created ON mission_runs(created_at DESC);

-- Concurrency guard: count active runs per mission quickly
CREATE INDEX IF NOT EXISTS idx_mission_runs_active ON mission_runs(mission_id, status)
    WHERE status IN ('pending', 'running');
```

**Down (`core/migrations/023_mission_runs.down.sql`):**

```sql
DROP TABLE IF EXISTS mission_runs CASCADE;
```

### Migration 024 — `mission_events`

**Up (`core/migrations/024_mission_events.up.sql`):**

```sql
-- Migration 024: Mission Events (V7 Event Spine)
--
-- Persistent audit record for every significant action within a mission run.
-- Persisted FIRST, then a lightweight CTS reference is published to NATS.
-- 17 fields per the V7 directive.

CREATE TABLE IF NOT EXISTS mission_events (
    event_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL DEFAULT 'default',
    mission_id UUID NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
    run_id UUID NOT NULL REFERENCES mission_runs(id) ON DELETE CASCADE,
    event_type TEXT NOT NULL,
    severity TEXT NOT NULL DEFAULT 'info',  -- info, warn, error
    provider_id TEXT,
    model_used TEXT,
    role TEXT,
    mode TEXT,  -- answer, proposal, execution
    payload JSONB DEFAULT '{}',
    artifact_refs JSONB DEFAULT '[]',
    parent_event_id UUID REFERENCES mission_events(event_id) ON DELETE SET NULL,
    parent_run_id UUID REFERENCES mission_runs(id) ON DELETE SET NULL,
    trigger_rule_id UUID,  -- FK added in migration 025 after trigger_rules exists
    audit_event_id UUID,
    intent_proof_id UUID REFERENCES intent_proofs(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Primary lookup: timeline for a run
CREATE INDEX IF NOT EXISTS idx_mission_events_run ON mission_events(run_id, created_at);

-- Secondary lookups
CREATE INDEX IF NOT EXISTS idx_mission_events_mission ON mission_events(mission_id);
CREATE INDEX IF NOT EXISTS idx_mission_events_type ON mission_events(event_type);
CREATE INDEX IF NOT EXISTS idx_mission_events_tenant ON mission_events(tenant_id);
CREATE INDEX IF NOT EXISTS idx_mission_events_created ON mission_events(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_mission_events_parent ON mission_events(parent_event_id);
```

**Down (`core/migrations/024_mission_events.down.sql`):**

```sql
DROP TABLE IF EXISTS mission_events CASCADE;
```

### Migration 025 — `trigger_rules`

**Up (`core/migrations/025_trigger_rules.up.sql`):**

```sql
-- Migration 025: Trigger Rules (V7 Trigger Engine)
--
-- Declarative rules that fire downstream mission runs based on event patterns.
-- Default mode is 'propose' — auto-execute requires explicit policy allowance.
-- Safety guards: cooldown, max_depth (recursion), max_active_runs (concurrency).

CREATE TABLE IF NOT EXISTS trigger_rules (
    trigger_rule_id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id TEXT NOT NULL DEFAULT 'default',
    enabled BOOLEAN NOT NULL DEFAULT true,
    source_mission_id UUID REFERENCES missions(id) ON DELETE SET NULL,
    event_type TEXT NOT NULL,
    predicate JSONB,  -- optional filtering predicate on event payload
    action_target_mission_id UUID NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
    mode TEXT NOT NULL DEFAULT 'propose',  -- propose, execute
    requires_approval BOOLEAN NOT NULL DEFAULT true,
    cooldown_ms INT NOT NULL DEFAULT 0,
    max_depth INT NOT NULL DEFAULT 5,
    max_active_runs INT NOT NULL DEFAULT 1,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Lookup indexes
CREATE INDEX IF NOT EXISTS idx_trigger_rules_tenant ON trigger_rules(tenant_id);
CREATE INDEX IF NOT EXISTS idx_trigger_rules_event ON trigger_rules(event_type);
CREATE INDEX IF NOT EXISTS idx_trigger_rules_enabled ON trigger_rules(enabled) WHERE enabled = true;
CREATE INDEX IF NOT EXISTS idx_trigger_rules_source ON trigger_rules(source_mission_id);
CREATE INDEX IF NOT EXISTS idx_trigger_rules_target ON trigger_rules(action_target_mission_id);

-- Deferred FK: link mission_runs.triggered_by_rule_id to trigger_rules
ALTER TABLE mission_runs ADD CONSTRAINT fk_mission_runs_trigger
    FOREIGN KEY (triggered_by_rule_id) REFERENCES trigger_rules(trigger_rule_id) ON DELETE SET NULL;

-- Deferred FK: link mission_events.trigger_rule_id to trigger_rules
ALTER TABLE mission_events ADD CONSTRAINT fk_mission_events_trigger
    FOREIGN KEY (trigger_rule_id) REFERENCES trigger_rules(trigger_rule_id) ON DELETE SET NULL;
```

**Down (`core/migrations/025_trigger_rules.down.sql`):**

```sql
-- Remove deferred FKs before dropping the table
ALTER TABLE mission_events DROP CONSTRAINT IF EXISTS fk_mission_events_trigger;
ALTER TABLE mission_runs DROP CONSTRAINT IF EXISTS fk_mission_runs_trigger;
DROP TABLE IF EXISTS trigger_rules CASCADE;
```

### Migration 026 — `trigger_executions`

**Up (`core/migrations/026_trigger_executions.up.sql`):**

```sql
-- Migration 026: Trigger Executions (V7 Trigger Engine Audit)
--
-- Records every trigger evaluation result for auditability.
-- result: 'evaluated' (checked but not fired), 'fired' (downstream run created),
--         'skipped' (guard prevented firing).
-- reason: human-readable explanation for skipped evaluations.

CREATE TABLE IF NOT EXISTS trigger_executions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trigger_rule_id UUID NOT NULL REFERENCES trigger_rules(trigger_rule_id) ON DELETE CASCADE,
    source_event_id UUID NOT NULL REFERENCES mission_events(event_id) ON DELETE CASCADE,
    source_run_id UUID NOT NULL REFERENCES mission_runs(id) ON DELETE CASCADE,
    target_run_id UUID REFERENCES mission_runs(id) ON DELETE SET NULL,
    result TEXT NOT NULL,  -- evaluated, fired, skipped
    reason TEXT,           -- explanation when skipped (cooldown, recursion_guard, concurrency_guard)
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Lookup indexes
CREATE INDEX IF NOT EXISTS idx_trigger_exec_rule ON trigger_executions(trigger_rule_id);
CREATE INDEX IF NOT EXISTS idx_trigger_exec_source ON trigger_executions(source_run_id);
CREATE INDEX IF NOT EXISTS idx_trigger_exec_target ON trigger_executions(target_run_id);
CREATE INDEX IF NOT EXISTS idx_trigger_exec_result ON trigger_executions(result);

-- Cooldown check: most recent execution per rule
CREATE INDEX IF NOT EXISTS idx_trigger_exec_cooldown ON trigger_executions(trigger_rule_id, created_at DESC);
```

**Down (`core/migrations/026_trigger_executions.down.sql`):**

```sql
DROP TABLE IF EXISTS trigger_executions CASCADE;
```

### Migration 027 — `scheduled_missions`

**Up (`core/migrations/027_scheduled_missions.up.sql`):**

```sql
-- Migration 027: Scheduled Missions (V7 Scheduler)
--
-- Cron-based mission scheduling. In-process goroutine with 1-minute resolution.
-- Suspends when NATS goes offline. Resumes on reconnect.
-- max_active_runs prevents schedule pile-up when missions run longer than intervals.

CREATE TABLE IF NOT EXISTS scheduled_missions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id UUID NOT NULL REFERENCES missions(id) ON DELETE CASCADE,
    tenant_id TEXT NOT NULL DEFAULT 'default',
    cron_expression TEXT NOT NULL,
    enabled BOOLEAN NOT NULL DEFAULT true,
    max_active_runs INT NOT NULL DEFAULT 1,
    last_run_at TIMESTAMPTZ,
    next_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Lookup indexes
CREATE INDEX IF NOT EXISTS idx_scheduled_missions_tenant ON scheduled_missions(tenant_id);
CREATE INDEX IF NOT EXISTS idx_scheduled_missions_mission ON scheduled_missions(mission_id);

-- Scheduler tick: find all due schedules efficiently
CREATE INDEX IF NOT EXISTS idx_scheduled_missions_due ON scheduled_missions(next_run_at)
    WHERE enabled = true;
```

**Down (`core/migrations/027_scheduled_missions.down.sql`):**

```sql
DROP TABLE IF EXISTS scheduled_missions CASCADE;
```

---

## 5. File Ownership Boundaries

Exact file ownership per team. **No overlap** — each file is owned by exactly one team.

### Team A — Event Spine

**Creates (new files):**

| File | Purpose |
|------|---------|
| `core/migrations/023_mission_runs.up.sql` | Mission runs table |
| `core/migrations/023_mission_runs.down.sql` | Rollback |
| `core/migrations/024_mission_events.up.sql` | Mission events table |
| `core/migrations/024_mission_events.down.sql` | Rollback |
| `core/pkg/protocol/events.go` | MissionEventEnvelope struct, EventType constants, EventSeverity |
| `core/internal/events/store.go` | Event persistence (Emit, EmitAndPublish, GetRunTimeline, GetChain) |
| `core/internal/runs/manager.go` | Mission run lifecycle (CreateRun, UpdateRunStatus, GetRun) |
| `core/internal/server/runs.go` | HTTP handlers for `/api/v1/runs/*` routes |
| `interface/types/events.ts` | TypeScript types for MissionEventEnvelope, RunChainNode |

**Modifies (existing files):**

| File | Change |
|------|--------|
| `core/pkg/protocol/envelopes.go` | Add `MissionEventID string` field to CTSEnvelope |
| `core/pkg/protocol/topics.go` | Add `TopicMissionEvents`, `TopicMissionEventsFmt`, `TopicMissionRunsFmt` |
| `core/internal/server/admin.go` | Register `GET /api/v1/runs/{run_id}/events` and `GET /api/v1/runs/{run_id}/chain` |
| `core/internal/swarm/soma.go` | Add `runs.Manager` and `events.Store` fields, wire into `ActivateBlueprint()` |
| `core/internal/swarm/agent.go` | Wire event emission into `processMessageStructured()` |
| `core/internal/server/mission.go` | Wire `mission_run` creation into `commitAndActivate()` |
| `core/cmd/server/main.go` | Initialize `events.Store` and `runs.Manager`, pass to AdminServer and Soma |

### Team B — Trigger Engine

**Creates (new files):**

| File | Purpose |
|------|---------|
| `core/migrations/025_trigger_rules.up.sql` | Trigger rules table + deferred FKs |
| `core/migrations/025_trigger_rules.down.sql` | Rollback |
| `core/migrations/026_trigger_executions.up.sql` | Trigger execution audit table |
| `core/migrations/026_trigger_executions.down.sql` | Rollback |
| `core/internal/triggers/store.go` | Rule persistence, in-memory cache, CRUD |
| `core/internal/triggers/engine.go` | Evaluation logic, guards (cooldown, recursion, concurrency), firing |
| `core/internal/server/triggers.go` | HTTP handlers for `/api/v1/triggers/*` routes |

**Modifies (existing files):**

| File | Change |
|------|--------|
| `core/internal/server/admin.go` | Register trigger CRUD routes |
| `core/internal/events/store.go` | Add `SetTriggerEngine()` method, call `Evaluate()` after `Emit()` |
| `core/cmd/server/main.go` | Initialize `triggers.Store` and `triggers.Engine`, wire into event store |

### Team C — Scheduler

**Creates (new files):**

| File | Purpose |
|------|---------|
| `core/migrations/027_scheduled_missions.up.sql` | Scheduled missions table |
| `core/migrations/027_scheduled_missions.down.sql` | Rollback |
| `core/internal/scheduler/scheduler.go` | Cron scheduler goroutine, NATS suspend/resume |
| `core/internal/server/schedules.go` | HTTP handlers for `/api/v1/schedules/*` routes |

**Modifies (existing files):**

| File | Change |
|------|--------|
| `core/internal/server/admin.go` | Register schedule CRUD + pause/resume routes |
| `core/cmd/server/main.go` | Initialize scheduler, wire NATS callbacks, graceful shutdown |

### Team E — Run Timeline + Chain UI

**Creates (new files):**

| File | Purpose |
|------|---------|
| `interface/components/runs/RunTimeline.tsx` | Vertical event timeline component |
| `interface/components/runs/ViewChain.tsx` | Causal chain tree component |
| `interface/components/runs/EventCard.tsx` | Expandable event detail card |
| `interface/components/runs/RunChainNode.tsx` | Recursive chain node component |
| `interface/app/(app)/runs/[id]/page.tsx` | Run timeline route page |
| `interface/app/(app)/runs/[id]/chain/page.tsx` | Run chain route page |

**Modifies (existing files):**

| File | Change |
|------|--------|
| `interface/store/useCortexStore.ts` | Add run timeline slices (activeRunId, runTimeline, runChain, actions) |
| `interface/components/dashboard/MissionControl.tsx` | Add "View Timeline" and "View Chain" buttons |
| `interface/components/dashboard/OpsOverview.tsx` | Add run status summary cards |

### Team D — Navigation Restructure

**Creates (new files):**

| File | Purpose |
|------|---------|
| `interface/app/(app)/automations/page.tsx` | Automations route page |
| `interface/app/(app)/resources/page.tsx` | Resources route page |
| `interface/app/(app)/system/page.tsx` | System route page |
| `interface/components/automations/AutomationsPage.tsx` | Automations tabs (active, triggers, drafts, approvals) |
| `interface/components/resources/ResourcesPage.tsx` | Resources tabs (brains, tools, catalogue, targets) |
| `interface/components/system/SystemPage.tsx` | System health dashboard |
| `interface/components/shared/DegradedMessage.tsx` | Reusable degraded mode message component |

**Modifies (existing files):**

| File | Change |
|------|--------|
| `interface/components/shell/ShellLayout.tsx` | New 5-item nav structure, Advanced toggle for System |
| `interface/components/shell/ZoneA_Rail.tsx` | Update nav items, icons, active state detection |
| `interface/app/(app)/dashboard/page.tsx` | Mission Control enhancements (run timeline integration) |
| `interface/app/(app)/memory/page.tsx` | Add recall/store event listing, run timeline links |

---

## 6. Test Plan

### Unit Tests (Team A)

**`core/internal/events/store_test.go`:**
- `TestEmit_PersistsEvent` — verify event is written to DB with all 17 fields.
- `TestEmit_GeneratesEventID` — verify returned event_id is a valid UUID.
- `TestEmit_ValidationFailure` — verify error when required fields missing (no run_id, no mission_id).
- `TestEmit_DefaultTenantAndSeverity` — verify defaults applied when omitted.
- `TestGetRunTimeline_Ordered` — insert 5 events with different timestamps, verify returned in ASC order.
- `TestGetRunTimeline_EmptyRun` — verify empty slice (not nil) for run with no events.
- `TestGetRunTimeline_FiltersByRunID` — insert events for 2 runs, verify only correct run returned.
- `TestGetChain_SingleRun` — verify chain with no children returns node with empty children array.
- `TestGetChain_NestedRuns` — create parent->child->grandchild runs, verify tree structure.
- `TestGetChain_EventCounts` — verify events_count field matches actual event count per run.
- `TestEmitAndPublish_NATSOffline` — verify event persists, no panic, degraded log emitted.

**`core/internal/runs/manager_test.go`:**
- `TestCreateRun_GeneratesUUID` — verify returned run_id is a valid UUID.
- `TestCreateRun_DefaultStatus` — verify new run has status "pending".
- `TestCreateRun_DepthIncrements` — verify child run has depth = parent depth + 1.
- `TestUpdateRunStatus_Running` — verify started_at set on transition to "running".
- `TestUpdateRunStatus_Completed` — verify completed_at set on transition to "completed".
- `TestUpdateRunStatus_InvalidTransition` — verify error on invalid status value.
- `TestGetRun_NotFound` — verify error for non-existent run.
- `TestListRunsForMission_Ordered` — verify runs returned in created_at DESC order.

**`core/internal/server/runs_test.go`:**
- `TestGetRunTimeline_200` — valid run_id returns events in APIResponse envelope.
- `TestGetRunTimeline_404` — unknown run_id returns 404 with error message.
- `TestGetRunTimeline_400` — invalid UUID format returns 400.
- `TestGetRunChain_200` — valid run_id returns chain tree.
- `TestGetRunChain_404` — unknown run_id returns 404.

### Unit Tests (Team B)

**`core/internal/triggers/engine_test.go`:**
- `TestEvaluate_MatchesByEventType` — rule with matching event_type returns matched=true.
- `TestEvaluate_NoMatchDifferentType` — rule with different event_type returns matched=false.
- `TestEvaluate_SourceMissionFilter` — rule with source_mission_id only matches correct mission.
- `TestEvaluate_GlobalRule` — rule with nil source_mission_id matches any mission.
- `TestEvaluate_PredicateMatch` — rule with predicate matches event payload fields.
- `TestEvaluate_PredicateMismatch` — rule with predicate rejects non-matching payload.
- `TestCheckCooldown_WithinWindow` — returns true when last execution within cooldown_ms.
- `TestCheckCooldown_Expired` — returns false when last execution outside cooldown_ms.
- `TestCheckCooldown_NoPriorExecution` — returns false (no cooldown to enforce).
- `TestCheckRecursionGuard_BelowMax` — returns false when depth < max_depth.
- `TestCheckRecursionGuard_AtMax` — returns true when depth >= max_depth.
- `TestCheckRecursionGuard_HardCeiling` — returns true when depth >= 10 regardless of rule max_depth.
- `TestCheckConcurrencyGuard_BelowMax` — returns false when active runs < max_active_runs.
- `TestCheckConcurrencyGuard_AtMax` — returns true when active runs >= max_active_runs.
- `TestFire_ProposeMode` — creates governance proposal, does NOT create run directly.
- `TestFire_ExecuteMode` — creates run with correct parent_run_id and triggered_by_rule_id.
- `TestFire_RecordsExecution` — trigger_execution record created with result and reason.
- `TestEvaluate_DisabledRule` — disabled rules are not evaluated.

**`core/internal/triggers/store_test.go`:**
- `TestCreateRule_Persists` — rule written to DB and retrievable.
- `TestCreateRule_DefaultMode` — mode defaults to "propose" when not specified.
- `TestUpdateRule_CacheRefresh` — modifying a rule refreshes the in-memory cache.
- `TestDeleteRule_CacheRefresh` — deleting a rule removes it from cache.
- `TestLoadRules_OnlyEnabled` — cache only contains enabled rules.
- `TestListRules_IncludesDisabled` — API listing returns all rules.
- `TestRecordExecution_Persists` — execution record written to DB.

### Unit Tests (Team C)

**`core/internal/scheduler/scheduler_test.go`:**
- `TestCheckDue_FindsDueSchedules` — schedules with next_run_at <= now are returned.
- `TestCheckDue_SkipsFuture` — schedules with next_run_at > now are not returned.
- `TestCheckDue_SkipsDisabled` — disabled schedules are not returned.
- `TestCronParsing_Valid` — valid cron expressions parse without error.
- `TestCronParsing_Invalid` — invalid cron expressions return error on schedule create.
- `TestEnforceConcurrency_BelowMax` — returns false when active runs < max.
- `TestEnforceConcurrency_AtMax` — returns true when active runs >= max.
- `TestCreateScheduledRun_UpdatesTimestamps` — last_run_at and next_run_at updated correctly.
- `TestSuspend_StopsTicking` — after Suspend(), no new runs created.
- `TestResume_ResumesTicking` — after Resume(), scheduler creates runs again.
- `TestSuspendOnNATSDisconnect` — verify suspend callback fires on NATS disconnect.
- `TestResumeOnNATSReconnect` — verify resume callback fires on NATS reconnect.

### Integration Tests

**Full Flow (Team A + B + C combined):**
- `TestFullFlow_NegotiateToTimeline` — negotiate intent -> commit -> mission_run created -> events emitted -> GET /runs/{id}/events returns timeline.
- `TestTriggerChain_CompletionFiresChild` — mission.completed event -> trigger evaluates -> fires -> child run created -> GET /runs/{id}/chain shows parent-child.
- `TestDegradedMode_NATSOffline` — disconnect NATS -> emit event -> event persists in DB -> CTS not published -> scheduler paused.
- `TestRecursionGuard_MaxDepthExceeded` — create chain of runs exceeding max_depth -> trigger skipped with reason "recursion_guard".
- `TestConcurrencyGuard_MaxActiveRuns` — create max_active_runs running runs -> new trigger skipped with reason "concurrency_guard".
- `TestScheduler_CreatesRunOnTick` — create schedule -> advance time past next_run_at -> verify mission_run created.

### Frontend Tests (Team E)

**`interface/__tests__/runs/RunTimeline.test.tsx`:**
- `renders event cards in chronological order` — mock API response with 5 events, verify DOM order.
- `color-codes events by type` — verify cyan for mission.*, green for tool.*, amber for trigger.*.
- `expands event card on click` — click card, verify payload section visible.
- `collapses event card on second click` — click again, verify payload section hidden.
- `shows loading skeleton during fetch` — verify skeleton elements before API resolves.
- `shows empty state message for run with no events` — mock empty response, verify message.

**`interface/__tests__/runs/ViewChain.test.tsx`:**
- `renders nested run nodes` — mock chain with 3 levels, verify all nodes rendered.
- `displays mission name and status badge` — verify text and badge color.
- `navigates to run timeline on click` — verify router push called with correct path.
- `shows depth indicator` — verify depth number displayed.

### Frontend Tests (Team D)

**`interface/__tests__/shell/Navigation.test.tsx`:**
- `renders 5 primary nav items` — verify Mission Control, Automations, Resources, Memory, System links.
- `hides System by default` — verify System nav item not visible initially.
- `shows System when Advanced toggled` — click toggle, verify System nav item appears.
- `highlights active nav item` — navigate to /automations, verify item highlighted.

**`interface/__tests__/automations/AutomationsPage.test.tsx`:**
- `renders 4 tabs` — verify Active Automations, Trigger Rules, Drafts, Approvals tabs.
- `switches tab content on click` — click Trigger Rules tab, verify trigger list rendered.

**`interface/__tests__/system/SystemPage.test.tsx`:**
- `shows degraded message when core offline` — mock failed health check, verify degraded component.
- `shows connected status when healthy` — mock successful health check, verify green indicator.

### E2E Tests

**`interface/e2e/specs/v7-run-timeline.spec.ts`:**
- Full user journey: ask question -> view timeline -> see events -> expand event details.
- Degraded mode: stop core -> navigate to timeline -> verify degraded message -> restart core -> verify recovery.

**`interface/e2e/specs/v7-trigger-chain.spec.ts`:**
- Create trigger rule -> execute source mission -> verify child run created -> view chain.

**`interface/e2e/specs/v7-navigation.spec.ts`:**
- Navigate to each of the 5 primary nav items -> verify page loads -> verify content.
- Toggle Advanced -> verify System page accessible.

---

## 7. Timeline Estimate

**Execution order is strict: A -> B -> C -> E -> D**

No time estimates per project convention. Steps are ordered, verifiable, and dependency-linked.

| Team | Scope | Steps | Dependencies | Verification Gate |
|------|-------|-------|-------------|-------------------|
| **A — Event Spine** | Migrations 023-024, MissionEventEnvelope struct, event store, run manager, API handlers, CTS extension, NATS topics, wiring into Soma/Agent, degraded mode | 9 major steps | None (foundational) | `inv core.test` passes, GET /runs/{id}/events returns events, GET /runs/{id}/chain returns tree |
| **B — Trigger Engine** | Migrations 025-026, trigger store with cache, evaluation engine with 3 guards, trigger CRUD APIs, wiring into event ingest | 5 major steps | Team A complete (uses mission_events, mission_runs, events.Store, runs.Manager) | `inv core.test` passes, trigger rules CRUD works, mission.completed fires child run |
| **C — Scheduler** | Migration 027, cron scheduler goroutine, schedule CRUD + pause/resume APIs, NATS disconnect/reconnect wiring | 4 major steps | Team A complete (uses runs.Manager, events.Store) | `inv core.test` passes, schedule creates run on tick, suspends on NATS disconnect |
| **E — Run Timeline UI** | RunTimeline, ViewChain, EventCard, RunChainNode components, Zustand slices, route pages, MissionControl integration | 5 major steps | Team A APIs available (GET /runs/*/events, GET /runs/*/chain) | `inv interface.build` succeeds, timeline renders events, chain renders tree |
| **D — Navigation** | 3 new pages (Automations, Resources, System), nav restructure, degraded mode messaging, content migration from old routes | 7 major steps | Team E complete (Automations page needs run timeline integration) | `inv interface.build` succeeds, all 5 nav items reachable, old routes redirect |

### Step Verification Checklist

**Team A complete when:**
- [ ] Migrations 023-024 apply cleanly (`inv core.test` includes migration test)
- [ ] MissionEventEnvelope validates correctly (unit tests pass)
- [ ] `events.Store.Emit()` persists to DB (integration test)
- [ ] `events.Store.EmitAndPublish()` publishes CTS with MissionEventID (integration test)
- [ ] `runs.Manager.CreateRun()` generates run with correct fields (unit test)
- [ ] `Soma.ActivateBlueprint()` creates mission_run and emits `mission.started` (integration test)
- [ ] GET /api/v1/runs/{id}/events returns timeline (API test)
- [ ] GET /api/v1/runs/{id}/chain returns tree (API test)
- [ ] Degraded mode: NATS offline does not panic, events persist (integration test)

**Team B complete when:**
- [ ] Migrations 025-026 apply cleanly
- [ ] Trigger rules CRUD works via API
- [ ] Rule matching by event_type + source_mission_id + predicate works
- [ ] Cooldown guard enforced
- [ ] Recursion guard enforced (hard ceiling at depth=10)
- [ ] Concurrency guard enforced
- [ ] `mode=propose` creates governance proposal (not direct execution)
- [ ] `mode=execute` creates child mission_run with correct parent linkage
- [ ] trigger_executions audit trail recorded

**Team C complete when:**
- [ ] Migration 027 applies cleanly
- [ ] Schedule CRUD works via API
- [ ] Scheduler creates run when next_run_at <= now
- [ ] Scheduler skips when max_active_runs reached
- [ ] Scheduler suspends on NATS disconnect
- [ ] Scheduler resumes on NATS reconnect
- [ ] Pause/resume endpoints work

**Team E complete when:**
- [ ] RunTimeline renders events in order with correct color coding
- [ ] EventCard expands/collapses
- [ ] ViewChain renders nested tree
- [ ] Zustand slices work (fetch, clear, active run ID)
- [ ] Route pages load (`/runs/[id]`, `/runs/[id]/chain`)
- [ ] MissionControl has "View Timeline" and "View Chain" buttons
- [ ] `inv interface.build` succeeds

**Team D complete when:**
- [ ] All 5 nav items render in ZoneA_Rail
- [ ] System hidden behind Advanced toggle
- [ ] Automations page has 4 tabs with content
- [ ] Resources page has 4 tabs with content
- [ ] System page shows health dashboard
- [ ] Degraded messages display when services offline
- [ ] Old routes redirect to new locations
- [ ] `inv interface.build` succeeds

---

## 8. Risk Analysis

### High Risk

**1. Event emission performance**

Every agent action (tool invocation, tool completion, policy evaluation) now persists a row to `mission_events`. For a mission with 5 agents each invoking 3 tools, that is 30+ events per run.

*Impact:* DB write pressure, potential latency increase on agent processing loop.

*Mitigation:*
- Batch inserts: buffer events in a channel, flush in batches of 10 or every 100ms (whichever comes first).
- Async persist goroutine: agent loop writes to channel (non-blocking), dedicated goroutine handles DB writes.
- The channel buffer size should be configurable (default 1000). If buffer fills, log warning and drop oldest event (preserve liveness over completeness).
- Index `mission_events(run_id, created_at)` as composite for efficient timeline queries.

**2. Trigger infinite loops**

Despite guards (max_depth, cooldown, max_active_runs), misconfigured rules could cascade. Example: Mission A completes -> triggers Mission B -> B completes -> triggers Mission A -> loop.

*Impact:* Runaway mission creation, resource exhaustion.

*Mitigation:*
- Hard ceiling: `max_depth = 10` is enforced regardless of rule configuration. No rule can set max_depth > 10.
- Global concurrent trigger limit: maximum 50 trigger evaluations in-flight at any time (semaphore in engine).
- Circuit breaker: if more than 20 trigger firings occur within 60 seconds for the same source mission, disable all rules for that mission and emit `EventPolicyEvaluated` with severity `error`.
- All trigger rules default to `mode = propose` — the governance loop prevents uncontrolled execution.

**3. Schema migration on existing data**

`mission_runs.mission_id` FK references `missions(id)` with `ON DELETE CASCADE`. If a mission is deleted, all its runs and events cascade-delete.

*Impact:* Accidental data loss if mission is deleted while runs are active.

*Mitigation:*
- CASCADE on delete is intentional — orphan runs are worse than cascaded deletes.
- Nullable FKs (`intent_proof_id`, `parent_run_id`, `triggered_by_rule_id`) use `ON DELETE SET NULL` — referenced entities can be removed without breaking runs.
- Down migrations drop tables in reverse dependency order (026 -> 025 -> 024 -> 023).
- All migrations are idempotent (`CREATE TABLE IF NOT EXISTS`, `CREATE INDEX IF NOT EXISTS`).

### Medium Risk

**4. CTS envelope backward compatibility**

Adding `MissionEventID` to CTSEnvelope changes the serialized JSON shape.

*Impact:* Existing consumers that strictly validate CTS schema could reject envelopes.

*Mitigation:*
- Field is `omitempty` — only present when an event is linked. Existing CTS consumers that ignore unknown fields are unaffected.
- No existing consumer uses strict schema validation (all use `json.Unmarshal` which ignores extra fields).
- SSE frontend handler forwards raw JSON — additional fields flow through transparently.

**5. Navigation restructure breaks bookmarks**

Users may have bookmarked `/catalogue`, `/settings/brains`, `/settings/tools`, etc.

*Impact:* 404 errors for bookmarked URLs.

*Mitigation:*
- Add redirect routes from old paths to new locations:
  - `/catalogue` -> `/resources?tab=catalogue`
  - `/settings/brains` -> `/resources?tab=brains`
  - `/settings/tools` -> `/resources?tab=tools`
  - `/approvals` -> `/automations?tab=approvals`
- Redirects are Next.js `redirect()` calls in the old route pages, not HTTP 301s (preserves SPA behavior).

**6. Scheduler precision**

1-minute ticker resolution means cron jobs can drift up to 59 seconds from their intended schedule.

*Impact:* Cron jobs scheduled at `0 * * * *` (every hour at :00) might run at :00 to :00:59.

*Mitigation:*
- Acceptable for V7 — sub-minute scheduling is a Phase 20 concern.
- `next_run_at` is computed precisely; the ticker just checks `<= now` so drift is at most 1 minute.
- For high-precision needs, users should use external schedulers and trigger via API.

### Low Risk

**7. Frontend state growth**

Adding run timeline slices (events array, chain tree) to the Zustand store increases memory.

*Impact:* Minimal — a timeline with 100 events is ~50KB JSON.

*Mitigation:*
- Slices are lazy-loaded (only fetched when user views a timeline).
- Cleared on run change (`clearRunTimeline()` called when `activeRunId` changes).
- No persistence to localStorage (unlike chat history) — purely ephemeral.

**8. Test coverage gap during transition**

New tables (`mission_runs`, `mission_events`, `trigger_rules`, etc.) need test fixtures and factory functions.

*Impact:* Tests may be incomplete during initial implementation.

*Mitigation:*
- Create test helper functions in `core/internal/server/testhelpers_test.go`:
  - `createTestMissionRun(t, db, missionID)` — inserts a run with sensible defaults.
  - `createTestMissionEvent(t, db, runID, eventType)` — inserts an event.
  - `createTestTriggerRule(t, db, eventType, targetMissionID)` — inserts a rule.
  - `createTestSchedule(t, db, missionID, cron)` — inserts a schedule.
- Each team writes tests alongside their implementation (not deferred).

### Mitigations Summary

| Risk | Mitigation | Default Behavior |
|------|-----------|------------------|
| Event performance | Async buffered channel, batch inserts | Buffer size 1000, flush every 100ms |
| Trigger loops | Hard max_depth=10, global semaphore=50, circuit breaker | All enforced by default |
| Schema migration | CASCADE delete, idempotent DDL, reverse-order down migrations | Safe by default |
| CTS compatibility | `omitempty` field, no strict schema validation | Backward-compatible |
| Broken bookmarks | Redirect routes from old paths | Active redirects |
| Scheduler drift | 1-minute resolution accepted | No mitigation needed for V7 |
| State growth | Lazy loading, clear on change | Ephemeral slices |
| Test coverage | Factory functions, tests alongside implementation | Enforced per team |

---

## Appendix A: Dependency Graph

```
Migration 023 (mission_runs)
    |
    v
Migration 024 (mission_events)  ──── depends on mission_runs
    |
    v
Migration 025 (trigger_rules)   ──── adds FK to mission_runs + mission_events
    |
    v
Migration 026 (trigger_executions) ── depends on trigger_rules + mission_events + mission_runs
    |
    v
Migration 027 (scheduled_missions) ── depends on missions (no V7 deps)
```

```
events.Store ──────────> runs.Manager
     |                       |
     v                       v
triggers.Engine ──> runs.Manager + events.Store
     |
     v
scheduler.Scheduler ──> runs.Manager + events.Store + Soma
```

## Appendix B: NATS Topic Map (V7 Additions)

| Topic | Publisher | Subscriber | Payload |
|-------|----------|------------|---------|
| `swarm.mission.events` | events.Store | Overseer, UI (SSE) | CTS with MissionEventID |
| `swarm.mission.{id}.events` | events.Store | Per-mission subscribers | CTS with MissionEventID |
| `swarm.mission.{id}.runs.{rid}.events` | events.Store | Per-run subscribers | CTS with MissionEventID |

## Appendix C: Event Type Catalog

| Event Type | Emitted By | Severity | Payload Contents |
|------------|-----------|----------|-----------------|
| `mission.started` | Soma.ActivateBlueprint | info | mission_name, team_count, agent_count |
| `mission.completed` | Overseer / completion handler | info | duration_ms, final_status |
| `mission.failed` | Error handler | error | error_message, failed_at |
| `mission.paused` | Governance halt | warn | reason, paused_by |
| `mission.resumed` | Governance resume | info | resumed_by |
| `team.spawned` | Team.Start | info | team_id, team_name, agent_names |
| `team.dissolved` | Team.Stop | info | team_id, reason |
| `agent.started` | Agent goroutine start | info | agent_name, role, model |
| `agent.stopped` | Agent goroutine stop | info | agent_name, reason |
| `tool.invoked` | Agent.processMessageStructured | info | tool_name, input_params |
| `tool.completed` | Agent.processMessageStructured | info | tool_name, duration_ms, output_summary |
| `tool.failed` | Agent.processMessageStructured | error | tool_name, error_message |
| `policy.evaluated` | Governance Guard | info/warn | decision, risk_level, rule_name |
| `proposal.created` | Trigger Engine (propose mode) | info | proposal_id, target_mission |
| `trigger.rule.evaluated` | Trigger Engine | info | rule_id, matched, skip_reason |
| `trigger.fired` | Trigger Engine | info | rule_id, source_run_id, target_run_id |

---

*End of V7 Implementation Plan.*
