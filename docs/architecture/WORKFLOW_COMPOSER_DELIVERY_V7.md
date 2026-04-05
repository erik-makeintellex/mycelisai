# Workflow Composer Delivery Plan V7
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Status: Active supporting plan
> Last Updated: 2026-03-09
> Scope: Drag-and-drop workflow composition, team manifestation orchestration, and release-safe delivery path

## 1. Purpose

Define the architecture-native workflow for delivering an Airflow-style visual composer in Mycelis, while preserving:

- governance-first execution
- mission run/event lineage integrity
- deterministic delivery gates
- strict git discipline per action

This plan is implementation-facing and designed for Team A/C/Q/T/D parallel execution.

Immediate alignment:
- the active operator-facing slice is Launch Crew and workflow onboarding
- onboarding must prove execution-facing outcomes before deeper workflow-composer expansion proceeds
- use [Delivery Governance And Testing V7](../architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md) and [Team Execution And Global State Protocol V7](../architecture-library/TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md) as the current delivery-discipline authority

## 1.1 Prerequisite Gate: Logging Standardization

Before any workflow-composer runtime or UI expansion:

- logging contract in `docs/logging.md` must be treated as authoritative.
- mission event emission must remain DB-first (`core/internal/events/store.go`).
- topic usage must remain constant-driven (`core/pkg/protocol/topics.go`).
- new node/runtime features must include mission event coverage in tests.

This prerequisite is mandatory so manifested workflows remain auditable and agent-parseable from day one.

## 2. Current Architecture Status (As Of 2026-03-05)

### 2.1 Completed Foundation

- V7 Event Spine is live (`mission_runs`, `mission_events`, run timeline APIs).
- Trigger Engine is live (rules CRUD + guarded execution paths).
- Conversation Log + Interjection are live.
- Inception Recipes are live.
- Runs list and timeline UX are live.

### 2.2 Current Gaps Blocking Full GUI Workflow Orchestration

- Team C Scheduler remains critical-path for repeat/scheduled execution behavior.
- Active Automations tab still blocked behind Team C completion.
- Causal Chain UI remains pending (`/runs/[id]/chain` UI not yet shipped).
- created-team workspace and channel inspector are not yet delivered.
- module-binding UX across internal/MCP/third-party API adapters is not yet unified.
- Worktree discipline is not yet standardized across in-flight slices.

### 2.3 Architectural Opportunity

The backend model already supports lineage, events, and trigger/schedule primitives. A visual composer can now become the unifying operator workflow for:

- creating execution DAGs
- manifesting teams from node intents
- attaching approvals/policies at mutation gates
- monitoring run outcomes and chain propagation

## 3. Target Product Shape: Workflow Composer

### 3.1 UX Model

Single canvas where operators compose a DAG with drag-and-drop blocks, connect dependencies, validate policy coverage, and execute as proposal-first.

### 3.2 Canonical Node Types (V1)

| Node Type | Purpose | Maps To |
| --- | --- | --- |
| Trigger | Event ingress rule | Trigger rules engine |
| Schedule | Time-based activation | Scheduler (`scheduled_missions`) |
| Manifest Team | Create team intent | `create_team` tool + governance path |
| Delegate Task | Team task execution | `delegate_task` tool |
| Decision/Gate | Conditional branch | Guard/policy evaluation |
| Approval | Human confirmation gate | Intent proof + confirm token flow |
| Artifact Output | Delivery/output sink | artifacts service + event emission |
| MCP Action | External tool invocation | MCP tool execution adapter |

### 3.3 Execution Semantics

- Graph must be acyclic and policy-valid before activation.
- High-impact nodes require explicit Approval node coverage.
- Every node execution emits `tool.invoked/tool.completed/tool.failed` + mission events linked to `run_id`.
- Child run propagation must remain queryable in chain view.

### 3.4 Soma-first manifestation thread (default)

- default operator path is intent -> Team Expressions -> module bindings -> proposal/activation.
- graph canvas is available for advanced composition, but must not replace terminal-state clarity in the default path.
- manifestation thread and composer graph must compile to one shared runtime contract.

### 3.5 Created-team operations

- every manifested team must expose an operator workspace with communication and control surfaces.
- created-team operation is part of workflow delivery, not a separate support feature.
- communications must be filterable and correlated by run/team/agent metadata.

## 4. Detailed Delivery Workflow (Build -> Deliver)

## Phase 0: Scope Freeze And Contract Lock

Owner: Team A + Team C + Team D

Deliverables:

- Node contract schema (input/output/validation/error model).
- DAG persistence contract (`workflow_definition`, `workflow_version`).
- Governance rules for high-impact node classes.
- UX state contract for draft/validated/proposed/active/completed.

Verification:

- contract test matrix defined and approved by Team Q.
- architecture docs updated (`README`, `V7_DEV_STATE`, `V7_IMPLEMENTATION_PLAN`, docs manifest).

Git discipline:

- one PR for contracts only; no runtime code mixed in.

## Phase 1: Backend Vertical Slice

Owner: Team A

Deliverables:

- workflow definition CRUD endpoints.
- compile service: DAG -> executable mission plan.
- scheduler/trigger integration adapter.
- governance preflight evaluator for composed plans.

Verification:

- package tests for compile validation and policy coverage.
- integration tests for schedule + trigger + run lineage creation.

Suggested commands:

- `uv run inv core.test`
- `cd core && go test ./internal/server -count=1`
- `cd core && go test ./internal/runs -count=1`

## Phase 2: Frontend Composer Vertical Slice

Owner: Team C

Deliverables:

- drag-and-drop canvas (ReactFlow-based).
- Soma-first manifestation thread with Team Expression editing and module binding.
- node palette + edge linking + form-based node config.
- draft validation UI (cycle detection, missing approvals, invalid refs).
- save/load workflow definitions.

Verification:

- unit tests for node editor/store reducers.
- route tests for composer page.
- accessibility and keyboard interaction coverage.

Suggested commands:

- `uv run inv interface.test`
- `uv run inv interface.build`
- `cd interface && npx vitest run --reporter=dot`

## Phase 3: Runtime Integration And Observability

Owner: Team A + Team C

Deliverables:

- run monitor panel for node-level status.
- chain linkage from composer runs to `/runs` and `/runs/[id]/chain`.
- created-team workspace with channel inspector and operator controls.
- failure/retry visualization and restart semantics.

Verification:

- integration tests: trigger -> schedule -> delegation -> artifact path.
- e2e flow from compose -> propose -> confirm -> run monitor -> timeline.

Suggested commands:

- `uv run inv core.test`
- `uv run inv interface.e2e`
- `cd interface && npx playwright test --reporter=dot`

## Phase 4: Release Hardening And Gate Enforcement

Owner: Team Q + Team T + Team D

Deliverables:

- deterministic release checklist.
- warning budget and regression thresholds.
- docs/state evidence bundle automation.

Verification:

- full baseline pass with reproducible command outputs.
- clean-tree enforcement for release promotions.

## 5. Team Collaboration Model

### Team A (Core Runtime)

- Owns execution contracts, compile path, scheduler/trigger coupling, chain correctness.

### Team C (GUI/Workflow UX)

- Owns composer UX, node configuration ergonomics, monitor surfaces, run navigation.

### Team Q (QA/Gates)

- Owns acceptance gates, deterministic test matrices, release evidence criteria.

### Team T (Tooling/Automation)

- Owns `inv` task ergonomics, CI parity, preflight and pre-push automation.

### Team D (Documentation/State)

- Owns architecture/status docs and in-app docs registration for every delivery slice.

## 6. Git Discipline (Mandatory For Each Confirmed Action)

1. Start an action branch with one scoped objective.
2. Create/update an action card with expected files and required validation commands.
3. Commit only objective-relevant files.
4. Run required validations before push.
5. Push and open PR with evidence block (commands + outputs + date).
6. Merge only after Team Q gate pass and docs gate completion.

Rules:

- No mixed-scope commits.
- No merge without tests (except docs-only changes).
- No release promotion from dirty worktree.

## 7. Invoke Task Strategy

### Existing Tasks To Use As Baseline

- `uv run inv ci.check`
- `uv run inv core.test`
- `uv run inv interface.build`
- `uv run inv interface.test`
- `uv run inv interface.e2e`
- `uv run inv lifecycle.health`

### Recommended Task Additions

| Task | Purpose |
| --- | --- |
| `ci.baseline` | strict full baseline sweep (core tests + interface build + vitest + playwright) |
| `ci.release-preflight` | enforce clean tree + docs gate + baseline pass |
| `ci.toolchain-check` | verify go/node/npm versions against declared policy |
| `test.scheduler` | targeted scheduler + automations integration test suite |
| `interface.test-runs` | targeted runs/timeline/chain frontend tests |
| `quality.warnings` | capture and summarize warning budget from test/build runs |
| `git.slice-check` | ensure staged-file scope matches branch objective |

## 8. Ideal Next Delivery Plan

## Sprint N+1 (Critical Path)

- Complete Team C scheduler runtime behavior.
- Ship Active Automations real data wiring.
- Stabilize policy and execution status surfaces.

Exit criteria:

- schedule fires create runs correctly.
- suspend/resume on NATS transitions works.
- automations page shows accurate active state and health.

## Sprint N+2

- Ship Causal Chain UI end-to-end.
- Connect composer monitor events to run timeline and chain views.
- Ship created-team workspace and communications inspector.
- Add retry/failure overlays for node-level debugging.

Exit criteria:

- `/runs/[id]/chain` is production-usable.
- chain traversal is validated by e2e tests.

## Sprint N+3

- Deliver V1 Workflow Composer (DAG draft -> validate -> propose -> execute).
- Enforce approval coverage for high-impact graph nodes.
- Finalize release gating automation in `inv` tasks.

Exit criteria:

- operators can build and execute governed DAG workflows visually.
- governance and lineage guarantees remain intact.
- release preflight is executable and enforced.

## 9. Definition Of Done For This Plan

- Architecture doc is published and registered in docs manifest.
- Team lanes and acceptance gates are unambiguous.
- Delivery order is explicit and dependency-safe.
- Git and `inv` operational discipline is codified for execution.
