# V8.2 Soma Team Interaction Contract
> Navigation: [Project README](../../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md) | [V8.2 Current State And Finalization PRD](V8_2_CURRENT_STATE_AND_FINALIZATION_PRD.md)

> Status: Canonical working contract
> Last Updated: 2026-05-16
> Module Boundary: Soma experience, team/workflow, governance/trust, runtime/capability, QA/embodiment
> Purpose: Define how Soma, Council, operators, and runtime teams talk about new or active work without exposing raw orchestration topology as the main product experience.

## Product Truth

Soma is the operator-facing continuity layer.

Council is a specialist review and governance support layer behind Soma.

Teams are scoped execution lanes. They are not separate assistant identities and must not become the default center of the product.

The operator experience must collapse into:

```text
I ask Soma for work.
Soma starts or steers the right team safely.
I can see active work, outputs, proof, and recovery.
I can intervene without reading raw bus traffic.
```

## Canonical Loop

```text
Operator intent
-> Soma frames the work
-> Council advises when useful
-> TeamInteraction is recorded
-> TeamWorkItem is queued or updated
-> Active Work Lane shows state
-> team emits normalized status/output
-> Soma summarizes result, proof, and next action
-> operator reviews, steers, retries, pauses, or archives
```

Team creation alone is not a completed work outcome. When a request asks for a team to do meaningful work, the system must either start a first work item, ask for approval to do so, or clearly say that no work item has started yet.

## Actor Responsibilities

| Actor | Owns | Must not own |
| --- | --- | --- |
| Soma | operator conversation, intent framing, work-item creation, priority, safety, summaries, next actions | raw specialist chatter as default UX |
| Council | critique, risk review, specialist advice, architecture or quality gates | independent operator identity |
| Team Lead | execution lane coordination, status events, output assembly, blockers | product-level governance decisions |
| Temporary Specialist | one bounded task with clear proof and stop condition | long-lived unowned autonomy |
| Operator | approval, steering, retry, pause/resume, archive, trust decisions | interpreting raw NATS or logs first |

## Decision Matrix

| Situation | Default path | Operator-visible explanation |
| --- | --- | --- |
| Simple answer or review | direct Soma response | answer plus evidence/proof when retained |
| Meaningful mutation | governed proposal | what will change, risk, output, proof, approval need |
| Broad deliverable | compact team or lane split | why a team is useful and what each lane produces |
| Active team needs input | operator-needed state | question, choices, impact, safe default |
| Quality or architecture concern | Council review behind Soma | review purpose, outcome, and effect on next action |
| Capability or provider failure | degraded state | what failed, what remains trusted, retry/recovery option |
| Team no longer needed | archive/stop path | retained outputs and proof remain available |

## Canonical Verbs

| Verb | Meaning | Required result |
| --- | --- | --- |
| `brief` | Soma gives a team objective, constraints, expected output, proof, and recovery posture. | `TeamWorkItem` created or updated. |
| `start` | A queued work item begins execution. | Active Work Lane moves to `running`. |
| `inspect` | Soma asks for current state without changing work. | Normalized status event. |
| `steer` | Operator or Soma changes objective, scope, priority, or acceptance criteria. | Audited interaction and updated work item. |
| `interject` | Operator adds information while work is active. | Audited input linked to the active item. |
| `review` | Council, Soma, or operator evaluates output or plan quality. | Review event and next action. |
| `escalate` | Work needs approval, specialist help, or operator attention. | Explicit blocker or approval state. |
| `recover` | Soma chooses retry, fallback, partial continuation, or stop. | Recovery state and proof impact. |
| `pause` | Work is intentionally stopped without archive. | Paused state with resume condition. |
| `resume` | Paused work continues from retained state. | Active state and lineage preserved. |
| `archive` | Team or work item is no longer active. | Outputs/proof retained; active lane removed. |

## Runtime Objects

### TeamInteraction

`TeamInteraction` is the durable record of a Soma, Council, operator, or team-lead exchange about team work.

Required fields:
- `interaction_id`
- `team_id`
- `work_item_id`
- `run_id` when execution-linked
- `source_kind`
- `source_channel`
- `actor_ref`
- `verb`
- `summary`
- `payload_kind`
- `payload_ref` or bounded payload
- `approval_ref` when governed
- `audit_refs`
- `timestamp`
- `version`

### TeamWorkItem

`TeamWorkItem` is the unit of active team execution.

Required fields:
- `work_item_id`
- `team_id`
- `run_id`
- `objective`
- `scope`
- `owner`
- `execution_shape`
- `expected_outputs`
- `expected_proof`
- `capability_requirements`
- `governance_posture`
- `state`
- `last_event`
- `needs_operator`
- `degradation_state`
- `recovery_options`
- `output_refs`
- `proof_refs`
- `created_at`
- `updated_at`
- `version`

### TeamStatusEvent

`TeamStatusEvent` is the normalized operator-readable projection of team communication.

Required fields:
- `event_id`
- `team_id`
- `work_item_id`
- `run_id`
- `state`
- `headline`
- `details`
- `confidence_posture`
- `blocked_by`
- `next_action`
- `timestamp`

### TeamOutputRef

`TeamOutputRef` links team work to retained product objects.

Required fields:
- `output_id`
- `team_id`
- `work_item_id`
- `run_id`
- `kind`
- `label`
- `storage_ref`
- `entrypoint`
- `validation_ref`
- `proof_ref`
- `created_at`

## State Model

| State | Meaning |
| --- | --- |
| `new` | Team exists but has not received a work item. |
| `briefed` | Soma has recorded an objective and expected output. |
| `queued` | Work is accepted but not running. |
| `running` | Team is actively executing. |
| `needs_operator` | Work is paused on missing input, approval, or choice. |
| `reviewing` | Output or plan is under Soma/Council/operator review. |
| `output_ready` | Retained output exists and can be opened or reviewed. |
| `degraded` | Work partially failed or proof is incomplete. |
| `paused` | Work is intentionally stopped and can resume. |
| `archived` | Team or item is inactive; outputs/proof remain retained. |

## UI Expression

The default Soma workspace should expose team work through these surfaces:

- Active Work Lane: compact list of queued, running, blocked, degraded, and output-ready work.
- Team Control Bar: `inspect`, `steer`, `pause`, `resume`, `archive`, and `start work` actions.
- Team Event Log: readable status/result/recovery events, not raw NATS subjects.
- Output Workbench: retained team deliverables with open, storage, proof, and validation controls.
- Trust Package: run, output, capability/team use, proof, audit, degradation, and next step.
- Council Review Drawer: advisory critique, risk notes, disagreement, and review lineage.

The UI must distinguish:
- team created
- first work item queued
- work running
- operator input needed
- output ready
- degraded or timed out
- archived

## Bus And Exchange Rules

Runtime code must use canonical protocol/topic constants for product subjects. Operator-facing UI consumes normalized `TeamStatusEvent`, `TeamInteraction`, `TeamOutputRef`, and run/proof projections rather than raw bus traffic.

Every governed product signal must include:
- `run_id` when execution-linked
- `team_id` when team-scoped
- `agent_id` when agent-scoped
- `source_kind`
- `source_channel`
- `payload_kind`
- `timestamp`

Mutating actions must emit persistent mission/run events in addition to transient bus signals.

## Governance And Recovery

No hidden mutation is allowed. Steering, pausing, resuming, archiving, capability use, and team expansion must be audited.

Temporary specialists require:
- why the lead cannot do the task alone
- owned task
- expected output
- expected proof
- stop/removal condition

Degraded team work must answer:
- what succeeded
- what failed
- what output remains trusted
- what proof is invalid or incomplete
- what can continue safely
- what requires retry or operator attention

## Release Acceptance

This contract is releasable only when the product proves:

1. A user asks Soma to create a team and produce a meaningful deliverable.
2. Soma creates or proposes a `TeamWorkItem`, not just a team shell.
3. The Active Work Lane shows queued/running/output/degraded state.
4. The operator can inspect and steer the work while active.
5. The output becomes a retained `TeamOutputRef`.
6. Proof/audit links explain team, capability, output, and recovery lineage.
7. Temporary teams can be paused, resumed, stopped, or archived without losing outputs.
8. Browser proof covers success, active intervention, and degraded/timeout recovery.

## Current Release Posture

`IN_REVIEW`: The architectural concept now exists as a canonical contract.

`NEXT`: Persist `TeamInteraction`, `TeamWorkItem`, `TeamStatusEvent`, and `TeamOutputRef` through backend APIs and project them into the Soma Active Work Lane.

`BLOCKED`: Do not claim runtime teams are useful delivery collaborators until a bounded role-specific team ask returns within product timeout and the UI exposes visible output or degradation.
