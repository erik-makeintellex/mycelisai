# Execution And Manifest Library V7

> Status: Canonical
> Last Updated: 2026-03-07
> Scope: Execution path selection, manifest lifecycle, run lifecycle, recurring-plan semantics, and canonical operator/runtime behavior.

Related:
- [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)
- [Target Deliverable V7](TARGET_DELIVERABLE_V7.md)
- [System Architecture V7](SYSTEM_ARCHITECTURE_V7.md)
- [UI And Operator Experience V7](UI_AND_OPERATOR_EXPERIENCE_V7.md)
- [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md)

Supporting specialized docs:
- [Soma-Council Engagement Protocol V7](../architecture/SOMA_COUNCIL_ENGAGEMENT_PROTOCOL_V7.md)
- [Soma Team Channel Architecture V7](../architecture/SOMA_TEAM_CHANNEL_ARCHITECTURE_V7.md)
- [Workflow Composer Delivery V7](../architecture/WORKFLOW_COMPOSER_DELIVERY_V7.md)

## 1. Why This Exists

The execution model must stop collapsing all intent into one-off chat behavior.

The system needs one language for:
- direct execution
- governed proposals
- team manifestation
- scheduled work
- event-driven work
- persistent-active watchers and plans

## 2. Canonical Execution Outcomes

Every actionable request must resolve into exactly one of:
- `answer`
- `proposal`
- `execution_result`
- `blocker`

Those outcomes are user-facing.

Underneath them, runtime execution may involve:
- direct answer generation
- council consultation
- internal tools
- MCP calls
- team manifestation
- delegated task runs
- scheduled or subscription-driven reactivation

## 3. Plan Modes

Every manifest, workflow definition, or standing automation must declare an operating mode.

| Mode | Runtime expectation | UI expectation |
| --- | --- | --- |
| `one_shot` | execute once and close | show terminal result and closure |
| `scheduled` | persist schedule, re-fire by time, survive restart | show schedule state, next run, pause/resume |
| `persistent_active` | maintain active watcher/service posture | show active state, health, pause/resume |
| `event_driven` | activate only when matching signals/rules fire | show subscribed triggers and recent activations |

## 4. Manifest Lifecycle

The canonical lifecycle is:
1. `draft`
2. `validated`
3. `proposed`
4. `approved`
5. `activated`
6. `paused` or `completed`
7. `rolled_back` or `cancelled` when necessary

Rules:
- validation happens before activation
- high-impact activation cannot skip approval
- activated manifests must preserve identifiers and lineage
- rollback must be explicit for durable plans

## 5. Run Lifecycle

Runs are execution instances, not definitions.

Canonical run states:
- `queued`
- `running`
- `waiting_for_approval`
- `paused`
- `completed`
- `failed`
- `cancelled`

Rules:
- every run has a stable `run_id`
- every meaningful mutation emits durable event lineage
- child runs preserve parent linkage

## 6. Team And Workflow Manifest Model

Team/workflow manifests should define:
- identity
- objective
- operating mode
- approval class
- input channels
- output channels
- artifact expectations
- runtime policy
- recovery/rollback behavior

For recurring or always-on plans, manifests must also define:
- activation trigger or schedule
- pause/resume semantics
- restart rehydration expectations
- failure escalation path

## 7. Runtime Activation Paths

Canonical activation paths:

### 7.1 Direct answer

Used when:
- no governed mutation is needed
- no durable plan needs to be created

### 7.2 Proposal-first

Used when:
- mutation or durable change is requested
- the system needs operator approval before activation

### 7.3 Manifest activation

Used when:
- work should persist as a team/workflow/automation definition
- recurring or always-on behavior is required

### 7.4 Event/schedule activation

Used when:
- a dormant definition is activated by time or signal

## 8. Canonical Transaction Surfaces

The operator may enter through UI or API, but the system must preserve the same transaction story:
- request origin
- validation
- proposal/approval if needed
- activation
- run/event emission
- result closure

Channels may include:
- HTTP
- NATS
- DB writes
- artifact writes

But the operator should experience one coherent flow.

## 9. Recurring And Always-On Plans

This is a first-class target, not a later enhancement.

The system must support plans that remain:
- scheduled repeatedly
- active continuously
- dormant until matching signals arrive

Each must survive:
- runtime restart
- deployment replacement
- temporary dependency loss

Each must expose:
- current state
- last activation
- next activation or active heartbeat
- failure condition
- recovery controls

## 10. Failure And Rollback Rules

Execution is not considered safe unless failures are classifiable.

Required failure classes:
- validation failure
- approval blocked
- dependency unavailable
- runtime execution failure
- communication timeout
- degraded recurring-plan state

Required rollback controls for durable plans:
- pause
- resume
- cancel
- deactivate
- versioned replacement when applicable

## 11. Immediate Design Implications

The workflow composer and automations UI must be built against this model.

If a plan is recurring or always-on, the UI must not present it as if it were a completed one-shot.

If a plan is one-shot, the UI must not bury the result inside persistent automation controls.

Next:
- [UI And Operator Experience V7](UI_AND_OPERATOR_EXPERIENCE_V7.md)
- [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md)
