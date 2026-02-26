Execution Target: Codex / Web Agentry  
Priority: High (Foundational UX Stabilization)  
Architecture Context: V7 (Event Spine + Runs + Triggers + Governance + Provider Profiles Live)

---

# PART I — CENTRAL DESCRIPTIVE INDEX
## “How to Fully Utilize the Stack”

This is the canonical mental model that all agents (human or AI) must align to.

Mycelis Stack Layers:

---

## 1️⃣ Intent Layer (User / API Entry)

Surfaces:
- Workspace (chat / soma)
- REST API
- Trigger Engine
- Scheduler (pending completion)

Purpose:
Capture structured intent and generate a Mission.

Core Objects:
- mission_runs
- mission_events
- team_id
- origin (workspace / trigger / schedule / api)

UI Responsibilities:
- Make origin visible
- Make routing decision visible
- Make governance visible

---

## 2️⃣ Orchestration Layer (Cognitive Decomposition)

Components:
- Council teams
- Provider profiles
- Brain selection
- Governance rules

Purpose:
Convert intent → execution plan.

UI Responsibilities:
- Show which team handled it
- Show which model/provider handled it
- Show why governance required approval (if so)
- Link to Run Detail

---

## 3️⃣ Event Spine (NATS)

Purpose:
Pub/Sub backbone for:
- Mission events
- System health
- Hardware channels
- Trigger dispatch

UI Responsibilities:
- Surface health state
- Surface degraded mode clearly
- Allow quick diagnostics
- Never silently fail

---

## 4️⃣ Execution Layer (Tools / MCP / Hardware)

Components:
- MCP servers
- Toolsets
- Filesystem
- Sensors (future)
- Actuators (future)

Purpose:
Act on the world (digital or physical).

UI Responsibilities:
- Show what tools are connected
- Show governance gating
- Link actuation to Run ID
- Allow operator halt

---

## 5️⃣ Observability Layer (Runs & Chain)

Components:
- Run Timeline
- Event log
- Causal Chain
- Artifacts

Purpose:
Explain what happened and why.

UI Responsibilities:
- Runs are first-class
- Timeline & Chain are easily reachable
- Saved Views for operators
- Clear axis conventions

---

## Canonical System Principle

Intent → Decomposition → Governance → Execution → Event → Observability → Feedback

The UI must make that loop obvious.

---

# PART II — EXECUTION TASK PACK

This is structured sequentially.
Each task has:
- Objective
- Files/Areas
- Required Expertise
- Acceptance Criteria

---

# TASK 001 — GLOBAL STATUS DRAWER

## Objective
Implement a globally accessible Status Drawer that reflects real system state.

## Areas
- Layout shell
- Global store
- `/api/v1/services/status`

## Required Expertise
- React global state management
- SSE / polling integration
- Operational UI design

## Must Show
- NATS status
- DB status
- SSE status
- Trigger engine
- Scheduler
- Council reachability
- Active runs count
- Recent errors

## Acceptance Criteria
- Drawer opens from header click
- Updates live without reload
- Degraded subsystems highlighted
- Copy diagnostics button works

---

# TASK 002 — DEGRADED MODE BANNER

## Objective
Add actionable degraded-state UX.

## Behavior
Trigger when:
- Any critical subsystem down
- Council unreachable
- SSE broken

## Required Expertise
- UX for failure handling
- Resilient UI states

## Must Include Actions
- Retry
- Open Status Drawer
- Switch to Soma-only
- Copy diagnostics

## Acceptance Criteria
- Appears automatically
- Disappears automatically
- Does not block UI interaction

---

# TASK 003 — RUNS AS FIRST-CLASS OBJECT

## Objective
Create a full Runs page with filters and multi-lens detail.

## Areas
- `/runs`
- `/runs/[id]`

## Required Expertise
- Data-heavy UI
- Table virtualization
- Timeline rendering
- Graph visualization

## Runs List Must Include
- Status
- Origin
- Team
- Provider
- Duration
- Created time

Filters:
- status
- origin
- team
- provider
- time range

Saved Views:
- Local storage (v1)

---

## Run Detail Tabs

1. Summary
2. Timeline (time axis)
3. Full Event Log
4. Chain (dependency axis)
5. Artifacts

Axis Rules:
- Timeline = horizontal time
- Chain = dependency, not time

## Acceptance Criteria
- Any run reachable in < 2 clicks from workspace
- Chain clearly shows parent trigger
- Timeline scrolls smoothly with 1000+ events

---

# TASK 004 — AUTOMATIONS HUB REDESIGN

## Objective
Remove dead state from Automations page.

## Layout

Left:
- Triggers
- Templates
- Approvals

Right:
- Scheduler (status + explanation)

Primary CTA:
- Create Automation Chain

## Required Expertise
- Empty-state UX
- Progressive enhancement design

## Acceptance Criteria
- Page never empty
- Trigger creation reachable in 1 click
- Links to Run history exist

---

# TASK 005 — STRUCTURED ERROR CARD

## Objective
Replace generic 500 errors in Workspace.

## Required Expertise
- Failure UX
- Inline action injection

## Error Block Must Show
- What happened
- Likely cause
- Impact
- Actions:
  - Retry
  - Reroute via Soma
  - Switch team
  - Copy diagnostics

## Acceptance Criteria
- Reroute replays message
- Retry preserves context
- No raw 500 visible

---

# TASK 006 — FOCUS MODE

## Objective
Collapse Ops pane without losing health visibility.

## Behavior
Key: `F`

Collapsed view shows:
- Alert count
- Active runs
- Health badge

## Acceptance Criteria
- State persists in session
- No layout break
- Zero reload

---

# TASK 007 — SYSTEM QUICK CHECKS PANEL

## Objective
Make System page operational.

## Must Include
- NATS
- DB
- SSE
- Trigger engine
- Scheduler

Each row:
- Status badge
- Last checked time
- Run check
- Copy snippet

## Acceptance Criteria
- User can diagnose degraded system without docs

---

# TASK 008 — TEAMS ENRICHMENT

## Objective
Make Teams infrastructural.

## Card Must Show
- Agents online
- Last heartbeat
- Health badge
- Quick actions:
  - Open chat
  - View runs
  - View logs

## Acceptance Criteria
- Council unreachable traceable from Teams

---

# TASK 009 — RESOURCES PANEL (MCP VISIBILITY)

## Objective
Expose execution capabilities.

## Must Show
- MCP servers
- Toolsets
- Governance requirements
- Permission levels

## Acceptance Criteria
- User understands what system can access

---

# TASK 010 — HARDWARE CHANNELS (V1 SCAFFOLD)

## Objective
Prepare UI for sensor/motor embodiment.

## Sections

Sensors:
- Topic
- Last message
- Sample payload

Actuators:
- Topic
- Allowed action schema
- Governance requirement

Safety:
- Emergency halt
- Kill switch state

## Acceptance Criteria
- Every actuator intent links to Run ID
- Sensor events display live

---

# PART III — EXPERTISE ASSERTION MATRIX

When invoking an AI agentry to execute these tasks, explicitly assert:

### UX Architect
- Control plane IA
- Failure state design
- Observability visualization patterns

### Frontend Systems Engineer
- Global state architecture
- SSE integration
- Virtualized tables
- Graph rendering

### Observability Engineer
- Timeline UX
- Event log filtering
- Causal graph semantics

### Reliability UX Specialist
- Structured error modeling
- Retry/fallback flows
- Diagnostics bundling

### QA Automation Engineer
- Playwright E2E
- Degraded state mocking
- Event-heavy test scenarios

---

# PART IV — NON-NEGOTIABLE QUALITY BARS

- No dead pages
- No unexplained 500
- All failures actionable
- Runs linkable everywhere
- Governance visible, not hidden
- Performance under 1.5s initial load
- Timeline handles 1000+ events smoothly

---

# PART V — CORE LEVEL TEST ASSURANCE MATRIX
## “Full Delivery Requires Proof, Not Visual Completion”

Every core layer must pass all test classes below before delivery is considered complete.

Test classes:
- Unit (component/store/handler behavior)
- Integration (API + state + data shaping)
- E2E (workflow proof in browser)
- Reliability (degraded/timeout/retry/fallback)
- Performance (latency/load/large payload)

---

## Level 1 — Intent Layer (Workspace / API Entry)

Required coverage:
- Unit:
  - intent submission, origin tagging, routing label display
  - workspace controls (focus toggle, status open, retry actions)
- Integration:
  - `/api/v1/chat` and `/api/v1/council/{member}/chat` response mapping
  - origin propagation (`workspace` / `trigger` / `schedule` / `api`)
- E2E:
  - user message -> response path with visible mode/role/brain/gov
  - council failure -> inline recovery actions (retry/reroute) without retyping
- Reliability:
  - council timeout/unreachable/server error variants
  - graceful fallback to Soma-only mode
- Performance:
  - first interaction render under target budget

Delivery gate:
- No raw 500 in UI for intent paths
- All intent failures provide at least 2 next actions

---

## Level 2 — Orchestration Layer (Council / Provider / Governance)

Required coverage:
- Unit:
  - provider badges, governance badge mapping, council reachability rendering
  - proposal/approval status transitions in UI state
- Integration:
  - provider provenance displayed from backend payload
  - governance-required paths correctly gate mutation actions
- E2E:
  - proposal -> approval -> execution flow with visible rationale
  - direct council call + reroute via Soma
- Reliability:
  - provider offline/degraded handling with actionable copy
  - governance strict mode behavior under partial failures
- Performance:
  - orchestration metadata renders without blocking primary chat flow

Delivery gate:
- Governance state visible on every mutation-capable surface
- Provider/model used is visible in run/workspace details

---

## Level 3 — Event Spine (NATS / SSE / Trigger Dispatch)

Required coverage:
- Unit:
  - stream state transitions (`live`/`offline`/`degraded`)
  - banner/drawer health status mapping
- Integration:
  - `/api/v1/services/status` -> global status components
  - trigger health and event delivery indicators
- E2E:
  - degraded banner appears/disappears automatically on outage/recovery
  - status drawer updates live without page reload
- Reliability:
  - NATS disconnected, SSE broken, reconnect jitter scenarios
  - retry controls recover state without navigation
- Performance:
  - stream event bursts do not freeze workspace interaction

Delivery gate:
- Degraded mode must self-heal visually after subsystem recovery
- No silent stream failure allowed

---

## Level 4 — Execution Layer (Tools / MCP / Toolsets / Hardware Scaffolds)

Required coverage:
- Unit:
  - MCP visibility cards, permission badges, governance requirement display
  - actuator/sensor scaffold rendering contracts
- Integration:
  - MCP servers/toolsets endpoint wiring and status rendering
  - permission model displayed correctly from API payloads
- E2E:
  - operator can identify accessible capabilities in <= 2 clicks
  - execution capability state remains visible during degraded mode
- Reliability:
  - tool offline and permission-denied states with clear recovery actions
  - emergency halt/kill-switch visual state transitions (when scaffolded)
- Performance:
  - capability surfaces remain responsive with larger MCP catalogs

Delivery gate:
- User can determine “what the system can access” without docs
- All execution failures include impact + next action

---

## Level 5 — Observability Layer (Runs / Timeline / Chain / Artifacts)

Required coverage:
- Unit:
  - run list filters, tab switching, timeline/chain render logic
  - axis semantics enforced (time vs dependency)
- Integration:
  - `/api/v1/runs`, `/api/v1/runs/{id}/events`, `/api/v1/runs/{id}/chain`
  - artifact links and provenance badges
- E2E:
  - from workspace to target run in <= 2 clicks
  - trace parent trigger -> child run chain correctness
- Reliability:
  - missing/partial event payloads still render non-fatally
  - chain/timeline panels handle backend partial outage states
- Performance:
  - timeline smooth at 1000+ events
  - list/table interactions remain responsive under large run volumes

Delivery gate:
- Every run must be explainable through visible timeline + chain context
- Observability surfaces cannot degrade into blank states

---

## Cross-Layer Regression Suite (Mandatory)

Must run for every release candidate:
1. Workspace intent failure recovery suite
2. Degraded mode auto-appearance and auto-clear suite
3. Runs discoverability and chain correctness suite
4. Automations hub actionable-state suite
5. Teams diagnostics in <= 2 clicks suite

---

## Release Exit Criteria

A feature set is “fully delivered” only if:
1. All five levels pass Unit + Integration + E2E
2. Reliability cases are proven with degraded-state tests
3. Performance thresholds are met for critical surfaces
4. No non-actionable failure copy remains
5. QA signoff includes captured evidence (videos/screenshots + logs)

---

# PART VI — DEFAULT UI ELEMENT INSTANTIATION FRAMEWORK
## “Trusted Continuous UI Framework”

Implementation authority for this section is maintained in `docs/UI_FRAMEWORK_V7.md` (canonical v2.0).
If wording diverges, follow `docs/UI_FRAMEWORK_V7.md` for active delivery and review gates.

This section defines the default, reusable way to instantiate UI elements across Mycelis.
All future UI surfaces MUST conform unless explicitly exempted in a decision log.

---

## 1. UI Element Taxonomy (Required)

Every new UI element must be classified as one of:

1. Primitive
- Badge, pill, button, icon status, text block, input row.
- No direct data fetching.
- Purely presentational.

2. Composite
- Cards, list rows, drawers, tool panels, tab content blocks.
- May transform data received via props.
- No direct backend calls unless explicitly documented.

3. Surface
- Route-level views (`/dashboard`, `/automations`, `/system`, `/runs`, etc.).
- Responsible for data orchestration, not low-level rendering.

Rule:
- Fetch at Surface level first.
- Pass normalized data down to Composite/Primitive levels.

---

## 2. Standard State Model (Every Element)

Each Composite/Surface must explicitly define:
- `loading`
- `ready`
- `empty`
- `degraded`
- `error`

No element may skip state declaration.
No element may collapse degraded/error into generic empty.

Required state contract shape:
```ts
type UIState<T> = {
  status: "loading" | "ready" | "empty" | "degraded" | "error";
  data?: T;
  reason?: string;
  next_actions?: Array<{ id: string; label: string }>;
  diagnostics?: Record<string, unknown>;
};
```

---

## 3. Status Semantics (Global Canonical)

Must be consistent everywhere:
- Green = `healthy`
- Yellow = `degraded`
- Red = `failure`
- Gray = `offline`
- Blue = `informational`

Never redefine these by page.
Never use alternative color meaning in local components.

---

## 4. Required UX Blocks Per Surface

Every Surface must include:
1. Purpose line
- One sentence: what this page does.

2. System state line
- Current health summary relevant to this page.

3. Primary next action
- Single most likely operator step.

4. Secondary actions
- 1–3 fallback actions.

5. Diagnostics path
- Link/button to inspect status, runs, or logs.

If any are missing, surface is incomplete.

---

## 5. Failure and Degraded Templates (Mandatory)

All failure/degraded blocks must include:
1. What happened
2. Likely cause
3. Impact
4. Next actions (2+)
5. Copy diagnostics

Canonical action set:
- Retry
- Open Status Drawer
- Continue in degraded mode (if possible)
- Route via Soma/default path

No raw backend error string may be the only user-facing output.

---

## 6. Data Instantiation Pattern

Default pattern for route surfaces:

1. Read from Zustand store snapshot
2. Start fetch/poll in `useEffect`
3. Normalize payload to local view model
4. Render by state contract
5. Provide inline retries (no forced page switches)

Polling default:
- Critical ops status: 5–10s
- Moderate telemetry: 10–15s
- Reference data: 30–60s

SSE behavior:
- If disconnected, degrade gracefully and show fallback timestamp + retry path.

---

## 7. Component Construction Contract

Each new component must define:
- `Props` interface
- `State` source (store, props, local)
- Rendering states map
- Action handlers
- Diagnostics output
- Test IDs for key interactions

Required naming:
- `XxxCard`, `XxxDrawer`, `XxxBanner`, `XxxPanel`, `XxxTab`

Required folder alignment:
- `components/dashboard/*` for workspace operational elements
- `components/automations/*` for workflow controls
- `components/system/*` for infra health
- `components/shared/*` for reusable primitives/composites

---

## 8. Accessibility and Interaction Baseline

All interactive elements must support:
- Keyboard access
- Focus-visible state
- `aria-label` where icon-only
- Escape/close behavior for drawers/modals
- Live region updates for high-priority degraded/failure changes

Focus Mode and status interactions must not trap keyboard navigation.

---

## 9. Performance Instantiation Rules

For list/timeline/log-heavy elements:
- Virtualize when data can exceed 300 rows.
- Debounce expensive filters.
- Avoid re-render storms from stream updates.
- Keep first meaningful render under target budget.

For run timeline/chain:
- Time axis and dependency axis logic must be separated.
- Expand/collapse should be O(visible nodes), not O(total dataset).

---

## 10. Testing Framework Per UI Element

Every new Composite/Surface requires:

1. Unit tests
- Rendering by state (`loading/ready/empty/degraded/error`)
- Action handlers fire expected callbacks

2. Integration tests
- Store + API path produces expected visual state
- Diagnostics and CTA visibility correctness

3. E2E tests
- Operator can recover from degraded/failure in <= 2 interactions
- Primary flow remains available without reload

4. Reliability tests
- Simulated timeout/unreachable/500 behavior
- Auto-recovery visibility after health restore

5. Performance checks
- Interaction latency under stress scenario for that component class

---

## 11. UI PR Gate (Mandatory Checklist)

A UI PR is not merge-ready unless it includes:
1. State map declaration (`loading/ready/empty/degraded/error`)
2. Failure template compliance
3. Status semantic compliance (global palette)
4. At least one test per new state path
5. Accessibility checks for all new controls
6. Link to next action from empty/degraded/error states
7. Update to docs/state if behavior changed

---

## 12. Expansion Rule

When adding a new product area:
1. Start from this framework
2. Reuse existing primitives first
3. Add new primitive only if no existing one fits
4. Record any deviation in decision log with rationale

This ensures continuity, trust, and operational clarity at scale.

---

# PART VII - PARALLELIZED DELIVERY ENGAGEMENT
## “From Framework to Active Execution”

Parallel delivery is active as of `2026-02-26`.

Execution board and lane playbooks:
- `docs/ui-delivery/PARALLEL_DELIVERY_BOARD.md`
- `docs/ui-delivery/LANE_A_GLOBAL_OPS.md`
- `docs/ui-delivery/LANE_B_WORKSPACE_RELIABILITY.md`
- `docs/ui-delivery/LANE_C_WORKFLOW_SURFACES.md`
- `docs/ui-delivery/LANE_D_SYSTEM_OBSERVABILITY.md`
- `docs/ui-delivery/LANE_Q_QA_REGRESSION.md`

---

## Priority and Gate Sequence

1. Gate A (`P0`):
- Lane A + Lane B + Lane Q `P0` reliability suite green
- Required outcomes:
  - global status truth is coherent
  - workspace failures are actionable and recoverable inline

2. Gate B (`P1`):
- Lane C + Lane Q workflow suite green
- Required outcomes:
  - automations/teams/resources surfaces always provide a next action

3. Gate C (`P2`):
- Lane D + Lane Q observability/performance suite green
- Required outcomes:
  - system and run surfaces remain operational under load and partial failure

4. Gate RC:
- Full cross-layer regression suite green with evidence

---

## Parallel Execution Rules

1. Lanes execute in parallel; gates merge in sequence.
2. No lane may merge without:
- framework state contract compliance (`loading/ready/empty/degraded/error`)
- failure template compliance
- lane evidence block completed
3. Lane Q validates every gate before promotion.

---

# END STATE

After completing this task pack:

Mycelis will feel like:
- A governed AI control plane
- An operational substrate
- An infrastructure product
- A trustworthy soma interface
- A recoverable, observable system

Not:
- A chat wrapper
- A swarm toy
- A dark dashboard
- An experimental demo
