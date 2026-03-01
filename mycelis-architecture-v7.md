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
Element planning authority (research-backed standards and Soma interaction patterns) is maintained in `docs/UI_ELEMENTS_PLANNING_V7.md`.
If wording diverges, follow those docs for active delivery and review gates.

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
- `docs/product/UI_WORKFLOW_INSTANTIATION_AND_BUS_PLAN_V7.md`

Named team ownership model:
- Team Circuit (Lane A), Team Atlas (Lane B), Team Helios + Team Argus (Lane C), Team Argus (Lane D), Team Sentinel (Lane Q)

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

---

# PART VIII - STANDARDIZED AI RESOURCE API CONTRACT
## “Channel-safe expansion without UI parsing drift”

All AI resource implementations (existing and new) must follow a shared contract model so new transaction channels can be added without reworking every surface.

## 1. Canonical Payload Shape

Preferred backend envelope:

```json
{
  "ok": true,
  "data": { "...": "..." },
  "error": ""
}
```

Backward-compatible acceptance:
- UI must also tolerate raw payloads (`data` omitted) for legacy endpoints.
- Normalization is handled at shared contract/store layers, not route components.

## 2. Frontend Normalization Rule

Required flow:
1. Surface/store fetches endpoint.
2. Shared contract helper normalizes envelope/raw.
3. Store writes typed state.
4. Components render from typed state only.

No component-level ad hoc JSON-shape branching for the same endpoint class.

## 3. Channel Expansion Contract

When adding a new transaction channel (voice, hardware, RAG state, external buses):
- define request/response DTOs in shared contract module first
- map channel status into global operational state (healthy/degraded/failure/offline/info)
- expose diagnostics and retry actions through existing degraded/status primitives
- add tests at Unit + Integration + E2E for channel-specific degraded/recovery paths

## 4. Operational Health Source-of-Truth

Global health indicators must consume centralized store status state.
Direct per-component polling of health endpoints is non-compliant unless explicitly approved in a decision log.

## 5. Delivery Gate for New Channels

No channel is considered production-ready until:
1. API contract is documented
2. shared normalization path is implemented
3. degraded/error UX follows failure template
4. observability links to run/status diagnostics are present
5. regression tests prove degraded -> recovery lifecycle

---

# PART IX - ARCHITECTURE ROLLOUT FIT
## “Where new architecture concepts land in development sequence”

The following rollout sequence aligns new architecture tracks with current delivery momentum and risk posture.

## Wave 0 - Core Continuity (Current)
- Complete Scheduler critical path and ensure stable run lifecycle continuity.
- Wire team lifetime profiles (`ephemeral` / `persistent` / `auto`) into orchestration contracts.

## Wave 1 - Runs + Workflow UX Hardening
- Complete Causal Chain and run observability UX paths.
- Finalize workflow surfaces and degraded-state recovery polish.

## Wave 2 - Universal Action Runtime Scaffolding (Parallel)
- Introduce universal action registry and gateway API scaffolds.
- Keep MCP as first adapter while adding OpenAPI/Python adapter interfaces.

## Wave 3 - Security-Gated Remote Actuation (Parallel, Gated)
- Implement secure gateway controls: scoped handshake, idempotency/replay protections, private mesh posture.
- Remote actuation remains disabled until all high-risk security tests pass.

## Wave 4 - Repeat + Low-Level IoT Expansion
- Promote one-off actions to scheduled repeat when requested.
- Route low-level IoT/watch-control pathways through persistent teams with strict governance controls.

## Execution Policy
- Direct Soma path remains default for low-risk one-off actions.
- Team manifestation is selected for complexity, risk, repeat intent, and IoT pathways.
- External protocols and UIs are additive integrations; governance authority remains in Mycelis.

## Companion Specifications
- `docs/V7_IMPLEMENTATION_PLAN.md` (integrated Wave 0-4 execution map)
- `docs/architecture/UNIVERSAL_ACTION_INTERFACE_V7.md`
- `docs/architecture/ACTUALIZATION_ARCHITECTURE_BEYOND_MCP_V7.md`
- `docs/architecture/SECURE_GATEWAY_REMOTE_ACTUATION_PROFILE_V7.md`
- `docs/architecture/SOMA_TEAM_CHANNEL_ARCHITECTURE_V7.md`
- `docs/architecture/SOMA_SYMBIOTE_GROWTH_AND_HOST_ACTUATION_V7.md`

---

# PART X - SOMA EXTENSION-OF-SELF PROGRAM
## "OpenClaw-inspired depth, Mycelis-governed execution"

This part defines the next architecture layer: Soma as a governed extension-of-self that can reason, choose execution strategy, learn safely, and actuate locally with transparent controls.

## 1. Pattern Assimilation (What We Adopt, What We Do Not)

Adopt:
- explicit persona contracts (identity + behavior policy)
- layered memory (episodic, distilled, retrieval, procedural)
- scheduled maintenance/automation loops
- strict treatment of external content as untrusted input

Do not adopt:
- unconstrained shell-level autonomy
- hidden policy drift without operator review
- remote actuation defaults

## 2. Extension-of-Self Operating Modes

For every intent, Soma must choose one execution mode:
1. `direct_action`
- default for low-risk, bounded actions with available capabilities
2. `manifest_team`
- used for complex, multi-step, multi-role, or uncertain tasks
3. `propose_only`
- used when governance requires approval or confidence is low
4. `scheduled_repeat`
- promoted from direct/team flows when user requests recurrence or policy infers stable repeat value

Required decision outputs:
- selected mode
- confidence score
- risk level
- approval requirement
- fallback plan
- team lifetime (`ephemeral|persistent|auto`)

## 3. Local Ollama First-Class Contract

Local Ollama is not a hidden dependency. It is a visible runtime substrate.

Required runtime checks:
1. host reachability (`OLLAMA_HOST`)
2. model availability for configured role routes
3. inference latency and throughput
4. recent failure ratio
5. fallback route state (active/inactive)

Policy behavior:
- if healthy: prefer local routing where allowed
- if degraded: apply explicit fallback profile and expose reason
- if unavailable: block risky auto-execution unless policy explicitly allows remote fallback

UI exposure requirements:
- active provider/model and local/remote badge
- last successful local inference timestamp
- fallback-active indicator with next action
- degraded banner + status drawer alignment

## 4. Memory and Growth Architecture Requirements

Soma growth must be artifact-backed and reversible.

Required memory layers:
1. Episodic: run and conversation records
2. Distilled: recipe abstractions and reviewed heuristics
3. Retrieval: vector-indexed semantic memory
4. Procedural: approved execution playbooks

Growth controls:
- no silent mutation of policy or approval thresholds
- promotion of new behavior requires reviewable artifact
- rollback path required for every profile/recipe promotion

## 5. Multi-User + Multi-Host Routing Contract

Enterprise deployment requires both:
1. multi-user tenancy boundaries
2. multi-host model routing for Soma/Council/teams

### 5.1 Tenancy boundaries

Required controls:
- namespace all runtime records with `tenant_id`
- apply tenant guards to temp memory channels, proposals, and mission runs
- ensure user identity is carried in API and event metadata for auditability
- prevent cross-tenant toolset/channel reads by default

### 5.2 Provider-target routing hierarchy

Each agent request resolves backend target in this order:
1. `agent.provider` override
2. `team.provider` default
3. `cognitive.profiles[role_or_profile]`
4. fallback (`sentry` or first healthy provider)

Provider IDs map to explicit hosts in `cognitive.providers` (Ollama, vLLM, LM Studio, Claude, Gemini, ChatGPT-compatible endpoints).

### 5.3 Runtime override controls (deployment-level)

For host-level routing without editing manifests:
- `MYCELIS_TEAM_PROVIDER_MAP` (JSON: team ID -> provider ID)
- `MYCELIS_AGENT_PROVIDER_MAP` (JSON: agent ID -> provider ID)

Default policy remains local-first:
- unpinned agents route to local Ollama profiles where healthy
- remote routes require explicit provider config and policy allowance

### 5.4 NATS multi-host collaboration requirement

Even when agents execute against different model hosts, collaboration is unified through the shared NATS bus:
- ingress: `swarm.global.input.*`
- team trigger/response: `swarm.team.<team>.internal.*`
- direct council addressing: `swarm.council.<agent>.request`

Cross-host reliability requirements:
- deterministic envelope format for all subjects
- reconnect-safe subscriptions
- backpressure-safe retries and dead-letter path for failed actuation
- status drawer exposure of per-channel and per-provider health

## 6. Universal Action + Channel Strategy

All execution channels must converge through universal action contracts:
- MCP (default local-first adapter)
- OpenAPI service adapters
- Python runtime management adapters
- hardware interface and direct channel adapters

Channel onboarding is valid only when:
1. typed request/response schemas exist
2. idempotency/replay protections exist for side effects
3. degraded/recovery UX is implemented
4. test evidence exists for failure and recovery paths

## 7. Parallel Architecture Delivery Tracks

Track A - Soma Decision + Memory Contracts:
- decision frame structs/APIs
- growth signal capture and review endpoints

Track B - Universal Action Runtime:
- action registry APIs
- adapter lifecycle and invocation contracts

Track C - Scheduler + Team Lifetime Runtime:
- repeat promotion and persistent-team pathways
- schedule reliability under degraded conditions

Track D - UI Operationalization:
- decision transparency in Workspace
- readiness diagnostics and next-action guidance

Track Q - QA and Reliability Gates:
- cross-layer contract tests
- degraded/recovery verification
- release evidence packaging

Merge policy:
- A/B/C/D can execute in parallel.
- promotion gates are sequential (`P0 -> P1 -> P2 -> RC`).
- no promotion without Track Q evidence.

## 7. Asserted Next Steps (Detailed)

Immediate (Sprint 0):
1. freeze decision frame and universal action DTOs
2. freeze Ollama readiness API contract and system-status mapping
3. add integration tests for local readiness -> fallback transitions

Near-term (Sprint 1):
1. ship direct-vs-team decision endpoint and UI trace exposure
2. complete team instantiation + lifecycle path from wizard to run evidence
3. bind run timeline/chain to decision artifacts

Mid-term (Sprint 2):
1. complete scheduler + repeat promotion paths
2. introduce reviewed adaptation signals (recipe feedback loop)
3. run reliability drills for NATS/SSE/Ollama degradation scenarios

Expansion (Sprint 3):
1. add one non-MCP adapter in production-shape parity
2. expose governed host/hardware action scaffolds
3. complete security-gated remote actuation preconditions (still disabled by default)

## 8. Controlling PRD

Execution authority for this program:
- `docs/product/SOMA_EXTENSION_OF_SELF_PRD_V7.md`

That document defines:
- sprint-by-sprint delivery
- lane ownership
- acceptance tests by core layer
- integration handoff expectations for parallel agent teams

## 9. Inception Harsh-Truth Guardrails

The following controls are mandatory during early extension-of-self rollout.

1. Single-use-case ratchet:
- no broad autonomy at inception
- start with one stable pipeline (for example: morning brief, research digest) and prove "worth it" before expanding permissions

2. Heartbeat autonomy budget:
- long-horizon jobs run under bounded heartbeat budgets (time window, action count, spend cap, and escalation cap)
- every heartbeat cycle must emit auditable progress + next action intent
- kill-switch can pause all autonomous loops immediately

3. Self-update isolation:
- Tier 1 runtime cannot patch its own production control plane directly
- updates must go through staging, verification, and rollback checkpoints
- "overnight self-update" without guardrails is non-compliant

4. Marketplace/skill injection governance:
- external skills/adapters are treated as untrusted until verified
- require signed manifests, scope mapping, sandbox execution profile, and policy review before activation
- no direct import of skill packages into high-privilege lanes

5. Trojan-horse and credential overreach defense:
- external data and discovered local artifacts are treated as potentially malicious
- deterministic secret redaction in logs and traces is mandatory
- no brute-force probing or broad credential discovery workflows are allowed in default policy

6. Sandbox-first runtime boundary:
- autonomy loops and dynamic adapters run in sandboxed execution contexts by default
- host filesystem and network access are allowlisted, not inherited
- escape attempts trigger immediate halt + audit event

