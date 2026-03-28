# MVP Release Strike Team Plan V7

> Status: Canonical
> Last Updated: 2026-03-10
> Purpose: Define the active team architecture, communication protocol, and state-discipline required to deliver MVP release gates without coordination drift.

Related:
- [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)
- [Target Deliverable V7](TARGET_DELIVERABLE_V7.md)
- [Next Execution Slices V7](NEXT_EXECUTION_SLICES_V7.md)
- [Team Execution And Global State Protocol V7](TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md)
- [V7 Development State](../../V7_DEV_STATE.md)

## 1. Most Vital Delivery For MVP

Primary delivery target for MVP release:
1. close Slice 2 execution-feedback and Workspace clarity acceptance
2. close Slice 3 prime-development reply reliability
3. reduce Slice 4 hot-path pressure enough to pass release gate baseline

Release-critical outcome:
- all changed operator flows terminate in `answer`, `proposal`, `execution_result`, or `blocker`
- release gate commands are green
- docs, tests, and runtime behavior match

## 2. Strike Team Architecture

The delivery team runs as five lanes with one accountable owner each.

| Lane | Team owner | Primary responsibility | Status owner |
| --- | --- | --- | --- |
| `Architecture/Governance` | `prime-architect` | scope control, acceptance criteria, marker discipline, blocker arbitration | status transitions + release verdict |
| `Runtime/Core` | `prime-development` | API/runtime/NATS behavior, reliability, hot-path extraction | backend slice evidence |
| `Interface/Operator` | `agui-design-architect` | Workspace simplification, direct-first Soma UX, blocker/recovery consistency | UI slice evidence |
| `QA/Verification` | `council-sentry` | deep test matrix, failure-path verification, gate evidence | pass/fail verdicts |
| `Release/Ops` | `admin-core` | lifecycle health, startup/shutdown reliability, release candidate packaging | go/no-go handoff |

## 3. Communication Contract (Stay In Touch)

### 3.1 Canonical lane channels

Use these canonical subjects:
- `swarm.team.{team_id}.internal.command`
- `swarm.team.{team_id}.signal.status`
- `swarm.team.{team_id}.signal.result`
- `swarm.mission.events.{run_id}`

Team-to-team rule:
- lanes are expected to communicate directly over NATS when a dependency crosses team boundaries.
- example: `prime-architect` can issue scoped work to `prime-development` on `swarm.team.prime-development.internal.command`.
- dependent teams should publish machine-readable progress on `signal.status` and bounded completion payloads on `signal.result`.
- no lane should depend on out-of-band chat as the only coordination record.
- teams may also self-coordinate on their own lane subject (for staged execution), for example:
  - command: `swarm.team.prime-development.internal.command`
  - status: `swarm.team.prime-development.signal.status`
  - result: `swarm.team.prime-development.signal.result`

### 3.2 Required coordination metadata

Every lane status/result signal must include:
- `run_id` (when execution-linked)
- `team_id`
- `agent_id` (when agent-scoped)
- `source_kind`
- `source_channel`
- `payload_kind`
- `timestamp`

### 3.3 Coordination cadence

Required touchpoints:
1. kickoff sync (start of day): confirm active slice, acceptance criteria, and expected evidence commands
2. checkpoint sync (every 2 hours): publish concise `signal.status` updates with status marker and blocker state
3. gate sync (end of day): QA publishes `signal.result` pass/fail summary and required carry-forward tasks

Escalation rule:
- blockers older than one checkpoint cycle must be raised to `Architecture/Governance` and reflected in `V7_DEV_STATE.md` immediately.

## 4. Global State File Discipline

Canonical state file:
- `V7_DEV_STATE.md`

Mandatory update triggers:
1. status marker transition (`NEXT -> ACTIVE`, `ACTIVE -> IN_REVIEW`, etc.)
2. gate run completion (`pass` or `fail`)
3. blocker found or cleared
4. release candidate cut decision

Each update entry must include:
1. date/time
2. lane and slice
3. marker transition
4. commands executed + outcomes
5. blocker/dependency notes
6. next 24-48h actions

No slice is review-ready if this file is stale relative to actual work.

## 5. MVP Delivery Board (Current)

### 5.1 Release stream A: Slice 2 closure

Status: `ACTIVE`

Objectives:
1. align Workspace blocker/recovery behavior across chat, banner, and drawer
2. enforce Soma direct-first for routine intent
3. keep diagnostics progressive-disclosure and reduce default dashboard density

Acceptance proof:
- focused Vitest suites for chat/failure/reroute
- Playwright proof for degraded/recovery flow
- live-backend proof when API proxy behavior changes

### 5.2 Release stream B: Slice 3 reliability

Status: `NEXT`

Objectives:
1. make `prime-development` reply reliably over canonical team channels
2. verify team sync path under realistic lifecycle conditions

Acceptance proof:
- `uv run inv team.architecture-sync`
- supporting task/runtime tests proving canonical request/reply behavior

### 5.3 Release stream C: Slice 4 gate pressure reduction

Status: `REQUIRED`

Objectives:
1. reduce max-lines pressure in known hot paths
2. preserve behavior while extracting bounded helpers

Acceptance proof:
- `uv run inv quality.max-lines --limit 350`
- `uv run inv ci.baseline`

## 6. Deep Testing Pack (Per Cycle)

Baseline command pack:
1. `uv run inv core.test`
2. `uv run inv interface.test`
3. `uv run inv interface.build`
4. `uv run inv ci.baseline`

Focused UI/recovery pack:
1. `uv run inv interface.e2e --project=chromium --spec=e2e/specs/v7-operational-ux.spec.ts`
2. `uv run inv interface.e2e --live-backend --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts`
3. `uv run inv interface.e2e --live-backend --project=chromium --spec=e2e/specs/workspace-live-backend.spec.ts`

Slice-specific focused tests should be attached in the same state-file update that reports the lane transition.

## 7. Definition Of Done For MVP Candidate

MVP candidate is `IN_REVIEW` only when:
1. Slice 2 acceptance criteria are met with evidence
2. Slice 3 is no longer blocking canonical team sync
3. Slice 4 has measurable gate relief and no regressions
4. release gate command pack is green
5. `V7_DEV_STATE.md` reflects all current markers, blockers, and evidence

MVP candidate is `COMPLETE` only when QA and Architecture/Governance both sign off with matching evidence in the state file.
