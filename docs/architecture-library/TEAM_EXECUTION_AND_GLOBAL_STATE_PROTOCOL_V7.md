# Team Execution And Global State Protocol V7

> Status: Canonical working protocol
> Last Updated: 2026-03-10
> Purpose: Define parallel team execution architecture, NATS coordination rules, global-state maintenance, launch gates, and deep-testing obligations.

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

### 1.1 Parallel Team Coordination Over NATS

Use canonical coordination subjects for lane execution and synchronization:
- `swarm.team.{team_id}.internal.command` for directed lane work items
- `swarm.team.{team_id}.signal.status` for concise progress/status updates
- `swarm.team.{team_id}.signal.result` for bounded completion payloads
- `swarm.mission.events.{run_id}` for run-linked delivery events and audit linkage

Required coordination metadata on team signals:
- `run_id` when execution-linked
- `team_id`
- `agent_id` when agent-scoped
- `source_kind`
- `source_channel`
- `payload_kind`
- `timestamp`

Coordination cadence:
1. architecture issues lane directives on `internal.command`
2. runtime/interface lanes emit progress on `signal.status` at each discrete contract checkpoint
3. QA emits gate verdicts on `signal.result` with evidence command list and pass/fail
4. all accepted transitions are copied into `V7_DEV_STATE.md` in the same delivery window

## 2. Primary Issue Program (Current)

### 2.1 Launch-Critical Workstreams

| Workstream | Status | Owner lanes | Primary issue |
| --- | --- | --- | --- |
| Slice 2: Logging, Error Handling, Execution Feedback | `ACTIVE` | Runtime/Core + Interface/Operator + QA | operator-facing failures still diverge across paths; some live-backend flows surface `503` without complete recovery guidance |
| Slice 3: Prime-Development Reply Reliability | `NEXT` | Runtime/Core + Architecture/Governance + QA | canonical multi-team coordination is incomplete until `prime-development` reliably replies on standard lanes |
| Slice 4: P1 Hot-Path Cleanup | `REQUIRED` | Runtime/Core + Interface/Operator + QA | `quality.max-lines` gate red on hot paths; cleanup required to stabilize delivery pace |
| Scheduler + Chain prerequisites for Slice 7 | `REQUIRED` | Runtime/Core + Interface/Operator + QA | created-team workspace remains blocked until recurring execution + chain UX are reliable |

### 2.2 Immediate 7-Day Plan

1. Move Slice 2 from `ACTIVE` to `IN_REVIEW` with aligned runtime/UI failure contract proof.
2. Move Slice 3 from `NEXT` to `ACTIVE` and close `prime-development` NATS reply gap.
3. Reduce at least one legacy max-lines cap pressure file in Slice 4 without behavior regressions.
4. Keep Slice 7 `BLOCKED` until scheduler persistence and chain operator behavior pass deep tests.

### 2.3 Launch Sequencing Rule

Launch sequencing remains strict:
1. Slice 2 acceptance
2. Slice 3 acceptance
3. Slice 4 measurable cleanup progress under gate
4. scheduler + chain prerequisites accepted
5. Slice 7 activation from `BLOCKED` to `ACTIVE`

## 3. Launch Readiness Gates

Gate stack for launch candidate:
1. `L0` contract gate:
- signal subject usage, metadata contracts, and status markers are consistent across runtime, UI, and docs
2. `L1` validation gate:
- `uv run inv core.test`
- `uv run inv interface.test`
- `uv run inv interface.build`
- `uv run inv ci.baseline`
3. `L2` execution gate:
- focused product-flow tests prove `answer|proposal|execution_result|blocker` terminal states on changed paths
- live-backend Workspace proof for any proxy/API-affecting changes
4. `L3` resilience gate:
- degraded dependency and restart/rehydration behavior validated for scoped slice

## 4. Global State File Maintenance Contract

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
- backend/API -> UI target review/test plan reference (required when runtime/API contracts changed)
- NATS coordination note (`internal.command` issue key + `signal.status/result` summary)
- blockers and dependency notes
- next 24-48h actions

Quality rule:
- if a PR changes execution behavior and does not update `V7_DEV_STATE.md`, it is not review-ready.
- if a PR changes runtime/API behavior and does not include both a UI target review/test plan reference and NATS coordination note in `V7_DEV_STATE.md`, it is not review-ready.

## 5. Deep Testing Architecture

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
- when backend/API contracts affecting UI proxy paths change, include live-backend Playwright evidence via `uv run inv interface.e2e --live-backend ...`

## 6. Weekly Operating Rhythm

1. Monday plan lock:
- select active slice, confirm dependencies, set status markers and test scope, publish lane directives on canonical command channels.

2. Daily execution:
- implement narrow behavior slices in parallel lanes, emit status/result updates on canonical channels, attach evidence, update global state file.

3. Wednesday risk review:
- verify blockers, test failures, and dependency drift.

4. Friday gate review:
- evaluate `IN_REVIEW` slices against acceptance criteria and deep-testing evidence.

## 7. Immediate Assignment Baseline

Current baseline assignment:
- `Architecture/Governance`:
  - enforce Slice 2 and Slice 3 acceptance contract and launch sequencing
  - maintain lane directive board and status-marker discipline
- `Runtime/Core`:
  - close prime-development NATS reply reliability
  - ship scheduler/chain prerequisites and hot-path reductions
- `Interface/Operator`:
  - align blocker/recovery UX across Workspace and team-facing surfaces
  - reduce Workspace default density and enforce conversation-first theme hierarchy
  - keep Soma chat direct-first for routine interaction; surface consultation only when explicitly triggered
  - prepare created-team communication inspector integration points
- `QA/Verification`:
  - run deep-test matrix with explicit UI proof for backend/API-affecting changes
  - record gate evidence and blockers in `V7_DEV_STATE.md`

### 7.1 Engaged snapshot (2026-03-10)

Current active priorities:
1. Slice 2 remains `ACTIVE`
2. Slice 3 remains `NEXT` and is first candidate to move to `ACTIVE`
3. Slice 4 remains `REQUIRED` with max-lines pressure still open
4. Workspace complexity/theme simplification is tracked inside Slice 2 (no new slice introduced)
5. Soma direct-first consultation policy is tracked inside Slice 2 and the Soma-Council protocol

Latest known gate posture:
- `uv run inv core.test` -> pass
- `uv run inv interface.test` -> pass
- `uv run inv interface.build` -> pass
- `uv run inv ci.baseline` -> fail (max-lines gate on hot paths)

Next 24-48h expected actions:
1. attach live-backend product-flow evidence for Slice 2 error/recovery contracts
2. move Slice 3 to `ACTIVE` with canonical lane proof (`internal.command` -> `signal.status/result`)
3. deliver first hot-path reduction in Slice 4 and re-run quality gate
