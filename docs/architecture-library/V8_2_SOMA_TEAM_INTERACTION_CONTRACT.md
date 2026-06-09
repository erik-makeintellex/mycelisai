# V8.2 Soma Team Interaction Contract
> Navigation: [Project README](../../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md) | [V8.3 Operational Embodiment PRD](V8_3_OPERATIONAL_EMBODIMENT_PRD.md)

> Status: Canonical working contract
> Last Updated: 2026-06-09
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

Team creation alone is not a completed work outcome. When a request asks for a team to do meaningful work, the system must either start a first work item, ask for approval to do so, or clearly say that no work item has started yet. Explicit specialist-output requests, such as creating a comic team with artist, character, dialogue, layout, and proof roles plus a generated page, must preserve the bounded roster in the ExecutionContract and attach the first retained deliverable or a degraded recovery state to the governed run.

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
| `archive` | Team or work item is no longer active. | Outputs/proof retained; active review removed. |

### Production Action Semantics

The current production-safe action API is:

```text
POST /api/v1/teams/{team_id}/work/{work_item_id}/actions
```

Supported actions are `start_work`, `pause`, `resume`, `archive`, `steer`, and `recover`. Each accepted action must create a durable `TeamStatusEvent`, create a durable `TeamInteraction`, update the `TeamWorkItem` state and `last_event`, and return the updated work item. `steer` and `recover` require a summary or bounded payload so the audit trail contains real operator guidance. `create_team` shell records are intentionally rejected by this endpoint until the product creates a delegated or deliverable work item for the team.

| Action | Allowed source states | Target state | Operator meaning |
| --- | --- | --- | --- |
| `start_work` | `new`, `briefed`, `queued` on delegated/deliverable work | `running` | Operator has released the work item into active execution; output/proof still arrives through governed runtime results. |
| `pause` | `queued`, `running`, `needs_operator`, `reviewing`, `degraded` | `paused` | Operator intentionally stops active progress without losing retained context. |
| `resume` | `paused` | `queued` | Operator asks Soma/team to continue from retained state; the system does not fabricate output. |
| `archive` | any non-archived delegated/deliverable work | `archived` | Work leaves active review while retained outputs, proof, audit, and history remain inspectable. The UI labels this `Clear from review`. |
| `steer` | any non-archived delegated/deliverable work | unchanged | Operator guidance is recorded as durable status and interaction evidence without pretending execution completed. |
| `recover` | `degraded`, `needs_operator` | `queued` | Operator requests safe continuation from retained context, output, proof, and audit state. |

Active review surfaces should read `GET /api/v1/teams/{team_id}/work?include_archived=false` so cleared work does not keep crowding the operator's decision queue. Archived rows remain durable history for audit, proof, output reconstruction, and later inspection.

### Bounded Team Ask Semantics

The current runtime usefulness proof API is:

```text
POST /api/v1/teams/{team_id}/work/ask
```

The endpoint accepts a `message` or structured `TeamAsk`, creates a delegated `TeamWorkItem`, and records a queued `TeamStatusEvent` plus ask `TeamInteraction`. Product UI calls this endpoint with `async=true`: the API publishes a governed command envelope to `swarm.team.{team_id}.internal.command`, includes `work_item_id` plus expected outputs/proof in the payload, and returns `202` immediately with `accepted=true` so Active Work becomes the continuity surface while the operator keeps using Soma. NATS offline or publish failure moves the item to `degraded`, sets `needs_operator=true`, records `degradation_state`, and returns recovery options. The legacy synchronous compatibility path can still wait for a bounded team response; readable replies move the item to `output_ready`, while timeout or unreadable raw tool-call/no-response output moves it to `degraded`. `/teams` and Soma home expose this as compact Ask Team or Respond controls on durable active-work rows; broad raw broadcast tests remain reserved for explicit broadcast/degradation proof.

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

- Current Work Lane: compact Dashboard summary of selected workflow, active task posture, latest output access, and next review action.
- Active Work Lane: compact list of non-archived queued, running, blocked, degraded, and output-ready work.
- Team Control Bar: operator-labeled controls such as `Open details`, `Open run`, `Reply to team`, `Ask for changes`, `Retry recovery`, `Clear from review`, `pause`, `resume`, `start work`, and bounded Ask Team or Respond actions.
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

Explicit specialist rosters require:
- one accountable lead
- named roles requested by the operator or required by the concrete deliverable
- capability requirements, such as `generate_image` and `save_cached_image` for media work
- retained output/proof expectations
- a bounded reason the roster should exist at creation instead of being split into multiple lead-only lanes

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
3. The Dashboard current-work lane summarizes the selected workflow and next action while the Active Work Lane shows queued/running/output/degraded state.
4. The operator can inspect and steer the work while active.
5. The output becomes a retained `TeamOutputRef`.
6. Proof/audit links explain team, capability, output, and recovery lineage.
7. Temporary teams can be paused, resumed, stopped, or archived without losing outputs.
8. Browser proof covers success, active intervention, and degraded/timeout recovery.

## Current Release Posture

`IN_REVIEW`: The architectural concept now exists as a canonical contract, and the current API/UI proof paths use durable team-work state instead of treating teams as only roster projections.

`IN_REVIEW`: `TeamInteraction`, `TeamWorkItem`, `TeamStatusEvent`, and `TeamOutputRef` persistence/API projection are landed for current proof paths. The Active Work Lane now calls the durable action API for production-safe `start_work`, `pause`, `resume`, `archive`, `steer`, and `recover` controls, and `/teams` plus Soma home now post async bounded Ask Team or Respond requests to the durable ask API before refreshing active work.

`IN_REVIEW`: Local-source proof now shows the bounded ask path creating durable degraded timeout state with status events, interactions, recovery options, and Active Work visibility. This proves honest degradation, not autonomous delivery usefulness.

`IN_REVIEW`: The first live `/teams` GUI proof is green for a real local `prime-development` ask returning `output_ready` within the browser path and showing reply proof text in Active Work. This proves the interaction path, not broad team delivery usefulness. The next gate is repeatable role-specific delivery output plus a consumer/projection path that advances async queued asks from status/result events without blocking the browser request.
