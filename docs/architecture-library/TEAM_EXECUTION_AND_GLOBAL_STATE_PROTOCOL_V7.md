# Team Execution And Global State Protocol V7

> Status: Canonical working protocol
> Last Updated: 2026-03-09
> Purpose: Define the execution architecture for delivery teams, the next-step plan, global state maintenance rules, and deep-testing obligations.

Related:
- [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)
- [Target Deliverable V7](TARGET_DELIVERABLE_V7.md)
- [Next Execution Slices V7](NEXT_EXECUTION_SLICES_V7.md)
- [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md)
- [Intent To Manifestation And Team Interaction V7](INTENT_TO_MANIFESTATION_AND_TEAM_INTERACTION_V7.md)
- [V7 Development State](../../V7_DEV_STATE.md)

## 1. Team Execution Architecture

Delivery executes as four coordinated lanes:

| Lane | Primary scope | Owns status transitions |
| --- | --- | --- |
| `Architecture/Governance` | target contract decisions, status marker enforcement, scope boundaries, acceptance gating | `REQUIRED` -> `NEXT` -> `ACTIVE` -> `IN_REVIEW` -> `COMPLETE` |
| `Runtime/Core` | Go runtime behavior, APIs, NATS contracts, persistence and migration behavior | runtime and API-oriented slices |
| `Interface/Operator` | execution-facing UI, manifestation flow, team workspace and channel inspector surfaces | UI and transaction-contract slices |
| `QA/Verification` | test design, evidence capture, failure-path validation, release gate verification | slice evidence and gate outcomes |

Cross-lane execution rule:
- no slice can be marked `COMPLETE` unless QA evidence and global-state updates are attached in the same delivery window.

## 2. Next-Step Execution Plan

### 2.1 Step A: Slice 6 activation (`ACTIVE`)

Goal:
- deliver Soma-first Team Expression and module-binding execution path as the default operator flow.

Primary deliverables:
- Team Expression rendering/editing in execution path
- module binding visibility before confirmation
- adapter normalization contract for `internal`/`mcp`/`openapi`/`host`

Exit criteria:
- proof of `proposal` -> confirmation -> run-linked outcome
- no adapter-specific leakage in primary operator terminal state language

### 2.2 Step B: Scheduler + chain readiness (`NEXT` + `REQUIRED`)

Goal:
- close prerequisites for created-team operations by shipping recurring execution and causal chain operator surface.

Primary deliverables:
- scheduler runtime and persistence (`scheduled_missions`)
- chain UI behavior aligned to existing chain API

Exit criteria:
- recurring behavior survives restart/rehydration
- chain navigation and error-state handling are operator-usable

### 2.3 Step C: Slice 7 activation (`BLOCKED` -> `ACTIVE`)

Goal:
- ship created-team workspace and channel inspector once Step B gates are satisfied.

Primary deliverables:
- created-team workspace tabs: `Overview`, `Communications`, `Members`, `Manifest`, `Controls`
- channel inspector filters and control actions

Exit criteria:
- command -> status/result round-trip evidence
- interjection/reroute/pause-resume/cancel controls validated

## 3. Global State File Maintenance Contract

Canonical state file:
- `V7_DEV_STATE.md`

Update triggers (must update state file when any occur):
1. slice start (`NEXT` -> `ACTIVE`)
2. status transition (`ACTIVE` -> `IN_REVIEW` or `BLOCKED`)
3. gate outcome (`pass` or `fail`)
4. blocker discovered or resolved
5. release-candidate cut or rollback

Required state entry fields:
- date stamp
- active slice and owner lane
- status marker change
- evidence commands and result (`pass`/`fail`)
- blockers and dependency notes
- next 24-48h actions

Quality rule:
- if a PR changes execution behavior and does not update `V7_DEV_STATE.md`, it is not review-ready.

## 4. Deep Testing Architecture

Every active slice must carry deep testing at five levels:

1. contract tests:
- API shapes, channel contracts, metadata requirements, docs/task contract checks

2. behavioral unit tests:
- changed local logic, guardrails, validation and failure classification

3. integration tests:
- UI/backend transaction mapping, run/event/correlation behavior, module adapter normalization

4. product-flow tests:
- real user journey to terminal state (`answer`, `proposal`, `execution_result`, `blocker`)

5. resilience tests:
- degraded dependency behavior, restart/rehydration, retry/rollback safety

Mandatory command gate baseline:
- `uv run inv core.test`
- `uv run inv interface.test`
- `uv run inv interface.build`
- `uv run inv ci.baseline`

When execution-facing UI behavior changes:
- include focused Playwright evidence via `uv run inv interface.e2e ...`

## 5. Weekly Operating Rhythm

1. Monday plan lock:
- select active slice, confirm dependencies, set status markers and test scope.

2. Daily execution:
- implement narrow behavior slices, attach evidence, update global state file.

3. Wednesday risk review:
- verify blockers, test failures, and dependency drift.

4. Friday gate review:
- evaluate `IN_REVIEW` slices against acceptance criteria and deep-testing evidence.

## 6. Immediate Assignment Baseline

Current baseline assignment:
- `Architecture/Governance`: enforce status marker and scope discipline for Slice 6 and Step B prerequisites
- `Runtime/Core`: module normalization + scheduler/chain prerequisite runtime work
- `Interface/Operator`: Team Expression/module-binding UI and chain/created-team surfaces
- `QA/Verification`: deep-test matrix execution, evidence capture, and gate decision logging in `V7_DEV_STATE.md`

### 6.1 Engaged execution snapshot (2026-03-09)

Current engaged slice:
- Slice 6 (`Soma-First Team Expression And Module Binding`) is `ACTIVE`

First execution checkpoint:
1. lock Team Expression + module-binding payload contract in runtime and UI docs
2. implement runtime normalization and UI read/write path for bindings
3. execute deep-test matrix and attach evidence in `V7_DEV_STATE.md`
4. hold Slice 7 as `BLOCKED` until scheduler + chain prerequisites are accepted

Current checkpoint evidence (2026-03-09):
- runtime payload contract wired:
  - mutation proposals now emit `team_expressions[]` with `module_bindings[]` in chat and direct council routes.
- operator surface wired:
  - proposal cards and orchestration inspector expose expression/binding counts and adapter-tagged binding chips.
- deep-test checkpoint:
  - `uv run inv core.test` -> pass
  - `uv run inv interface.test` -> pass
  - `uv run inv interface.build` -> pass
  - `uv run inv ci.baseline` -> fail (`quality.max-lines` gate: `core/internal/swarm/agent.go`, `core/internal/swarm/internal_tools.go`, `interface/store/useCortexStore.ts`)
- immediate next-step:
  - keep Slice 6 `ACTIVE` while attaching product-flow proof that links Team Expression proposal payload -> confirmation -> run-linked outcome.
