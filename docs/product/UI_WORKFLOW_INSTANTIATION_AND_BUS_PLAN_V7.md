# UI Workflow Instantiation and Bus Plan (V7)

> Purpose: define how users instantiate and manage teams with minimal manual effort, how UI handles input/output channels, and how NATS is exposed safely without overwhelming operators.
> Status: execution authority for kickoff.
> Companion docs: `docs/UI_FRAMEWORK_V7.md`, `docs/UI_ELEMENTS_PLANNING_V7.md`, `docs/ui-delivery/PARALLEL_DELIVERY_BOARD.md`

---

## Table of Contents

1. Operating Principles
2. User Outcomes
3. Workflow Instantiation Model
4. Team Lifecycle Management UX
5. Input/Output Contract Model
6. NATS Exposure Model for Users
7. Minimal-Manual Instantiation Strategy
8. User Workflow Blueprints
9. Parallel Team Execution Plan
10. Sprint 0 and Sprint 1 Work Packages
11. Acceptance Criteria and Test Plan
12. Handoff and Governance

---

## 1. Operating Principles

1. Users configure intent, not infrastructure internals.
2. Advanced controls exist, but remain progressive disclosure.
3. Team instantiation is wizard-driven and readiness-gated.
4. Every step shows blockers, impact, and next action.
5. NATS complexity is abstracted into workflow language by default.
6. Any failure state must provide inline recovery without route switching.

---

## 2. User Outcomes

The UI is considered operationally optimal when a user can:
- instantiate a team from intent in 3 steps or fewer
- understand readiness before launch
- monitor outputs and failures from one control surface
- recover from degraded state without manual topic plumbing
- trace all execution and channel outputs to a run ID

---

## 3. Workflow Instantiation Model

### 3.1 Instantiation Pipeline

1. Intent capture
2. Capability readiness check
3. Team composition proposal
4. Governance and risk review
5. Launch confirmation
6. Runtime observation

### 3.2 Required UI Objects

- `IntentBrief`
  - intent text
  - objective type
  - urgency
  - risk/approval posture

- `InstantiationPlan`
  - target team(s)
  - required capabilities
  - expected inputs/outputs
  - fallback behavior

- `ReadinessGate`
  - providers
  - MCP tools
  - governance mode
  - stream health
  - data store availability

- `LaunchDecision`
  - launch now
  - launch propose-only
  - save as template

### 3.3 Planned UI Components

- `TeamInstantiationWizard`
- `CapabilityReadinessGateCard`
- `LaunchReviewSheet`
- `RuntimeActuationPanel`

### 3.4 Data Contracts for Instantiation

```ts
type TeamProfileTemplate = {
  id: string;
  name: string;
  objectiveKinds: string[];
  defaultGovernanceMode: "passive" | "approval_required" | "halted";
  requiredCapabilities: string[];
  suggestedRoutes: string[];
};

type ReadinessSnapshot = {
  providerReady: boolean;
  mcpReady: boolean;
  governanceReady: boolean;
  natsReady: boolean;
  sseReady: boolean;
  dbReady: boolean;
  blockers: string[];
};
```

---

## 4. Team Lifecycle Management UX

### 4.1 Lifecycle States

- Draft
- Ready
- Running
- Degraded
- Halted
- Completed

### 4.2 Required Management Actions

- instantiate from template or intent
- pause/resume (where supported)
- reroute execution target
- inspect latest run
- open related approvals
- rollback routing changes (guided mode)

### 4.3 Team Card Contract

Each team card must include:
- health status
- active run count
- last heartbeat
- next best action
- quick links: runs, logs, wiring, approvals

---

## 5. Input/Output Contract Model

### 5.1 Input Channels (User-Facing)

- workspace prompt
- trigger event
- schedule event
- API invocation
- sensor message (future-facing)

### 5.2 Output Channels (User-Facing)

- chat response
- proposal block
- run timeline event
- artifact/output card
- governance decision request

### 5.3 Unified I/O Envelope

All channels normalize to:

```json
{
  "ok": true,
  "data": {},
  "error": "",
  "meta": {
    "channel": "workspace|trigger|schedule|api|sensor",
    "run_id": "",
    "team_id": "",
    "timestamp": ""
  }
}
```

### 5.4 UI Contract Rules

- normalize at shared adapter/store layer
- avoid per-component parsing branches
- render with channel-aware but uniform UX patterns
- surface `channel` and `run_id` on all mutation-relevant outputs

---

## 6. NATS Exposure Model for Users

Goal: expose bus state and routing value without requiring topic engineering by default.

### 6.1 Three-Tier Exposure

1. Basic (default for most users)
   - health
   - throughput
   - recent failures
   - no raw topic editing

2. Guided (builder)
   - route templates
   - route selectors
   - safe subscription presets

3. Expert (advanced mode)
   - raw topic visibility
   - pattern subscriptions
   - diagnostics stream filters

### 6.2 NATS UI Surfaces

- `BusHealthPanel` (status + reconnect actions)
- `RouteTemplatePicker` (predefined patterns)
- `BusActivityLens` (semantic event feed)
- `AdvancedTopicInspector` (expert only)

### 6.3 Manual Burden Reduction

- provide route templates by workflow type
- auto-populate suggested routes from selected team profile
- show impact preview before enabling route
- offer one-click rollback for routing changes
- default to semantic labels (not raw subjects) in Basic mode

---

## 7. Minimal-Manual Instantiation Strategy

### 7.1 One-Click Profiles

Provide profile-driven setup:
- Research Team
- Incident Team
- Delivery Team
- Governance-First Team

Each profile auto-fills:
- recommended team composition
- provider defaults
- required tools
- expected I/O channels
- governance mode

### 7.2 Guided Wizard Steps

1. Objective
2. Capability checks
3. Composition
4. Risk and governance
5. Launch and monitor

### 7.3 Auto-Assist Behaviors

- auto-detect missing capabilities and suggest install path
- auto-route to approvals for risky actions
- auto-open run on launch success
- auto-suggest fallback to Soma when council targets fail

---

## 8. User Workflow Blueprints

### 8.1 Blueprint A - Instantiate from Workspace

1. User submits objective in Workspace.
2. Wizard pre-fills profile from intent classification.
3. Readiness gate shows pass/fail and blockers.
4. User confirms launch (or propose-only).
5. Runtime panel opens with run link and bus health context.

### 8.2 Blueprint B - Builder Route Setup (No Raw Topics)

1. User chooses Guided mode in NATS panel.
2. Route template is selected by workflow type.
3. UI previews impact and required capabilities.
4. User applies route with one-click rollback option.
5. Status drawer reflects live routing health.

### 8.3 Blueprint C - Degraded Recovery Without Context Loss

1. NATS/SSE failure detected.
2. Degraded banner provides Retry, Open Status, Continue Degraded.
3. Retry re-checks health and preserves current draft context.
4. Recovery auto-clears banner and returns to prior flow state.

### 8.4 Blueprint D - Team Diagnosis to Intervention

1. User opens Teams view and sees degraded team card.
2. Card shows latest run and heartbeat drift.
3. User opens logs or run in one click.
4. User reroutes or pauses team from same context.

---

## 9. Parallel Team Execution Plan

### Team Atlas (Orchestration UX)
- Owns: instantiation wizard, launch review, lifecycle actions
- Outputs: workflow-start UX and launch controls
- Interfaces: Team Helios for readiness data, Team Circuit for bus state

### Team Helios (Resources + I/O)
- Owns: readiness gate, I/O contract adapters, capability matrix
- Outputs: capability confidence and channel consistency
- Interfaces: Team Atlas for launch gating, Team Argus for run metadata

### Team Circuit (Bus + Reliability UX)
- Owns: NATS exposure model, stream continuity, recovery actions
- Outputs: bus visibility with low cognitive load
- Interfaces: Team Atlas for lifecycle gating, Team Sentinel for reliability criteria

### Team Argus (Runs + Teams Observability)
- Owns: run pivots, team diagnostics, triage shortcuts
- Outputs: fast diagnosis and intervention
- Interfaces: Team Atlas for lifecycle links, Team Helios for channel metadata

### Team Sentinel (QA + Governance)
- Owns: reliability tests, gate criteria, evidence and policy checks
- Outputs: release confidence and compliance
- Interfaces: all teams for gate signoff

---

## 10. Sprint 0 and Sprint 1 Work Packages

### Sprint 0 (now) - Contract and Scaffolding

- Atlas
  - scaffold `TeamInstantiationWizard` with state machine
  - define launch summary schema and validation
- Helios
  - implement shared channel metadata adapters in store/contracts
  - produce readiness gate mapping from existing service endpoints
- Circuit
  - implement Basic/Guided/Expert mode state and visibility policy
  - scaffold `RouteTemplatePicker` and rollback action contract
- Argus
  - define team-to-run pivot model and related quick actions
  - map required run metadata fields for team cards
- Sentinel
  - add test plan matrix for wizard, readiness, and bus recovery paths
  - lock gate criteria for Gate A -> Gate B transition

### Sprint 1 - First Usable Vertical Slice

- deliver instantiate -> launch -> run -> diagnose happy path
- deliver NATS degraded -> retry -> recovered path
- deliver team card quick actions with run/log pivots
- deliver UI evidence package for Gate B readiness

---

## 11. Acceptance Criteria and Test Plan

### Core Success Criteria

1. User can instantiate a team via guided flow without manual topic edits.
2. User can read channel I/O context in each critical surface.
3. NATS is understandable in Basic mode and controllable in Guided mode.
4. Degraded bus state always provides immediate recovery actions.
5. Team diagnosis to intervention is possible in 2 clicks or fewer.

### Test Requirements

- Unit:
  - wizard state transitions
  - readiness gate status mapping
  - channel badge rendering
  - NATS exposure mode gating

- Integration:
  - input channels map into unified contract
  - output channels map into run/artifact/proposal surfaces
  - readiness gate combines services and resource states correctly

- E2E:
  - intent -> instantiate -> launch -> run -> diagnosis
  - NATS degraded -> recover path without route dead ends
  - teams degraded -> logs/run pivot in <= 2 clicks

- Reliability:
  - NATS disconnect/reconnect churn
  - SSE interruption while wizard in progress
  - partial readiness data with fallback copy

---

## 12. Handoff and Governance

Every lane PR must include:
- mapping to section(s) in this plan
- explicit user outcome being improved
- before/after workflow map
- test evidence for success and degraded paths

Release gating:
- no feature merges without reduced manual burden evidence
- no bus-facing UX merges without Basic mode validation
- no instantiation feature merges without readiness gate tests

