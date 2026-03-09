# Target Deliverable V7

> Status: Canonical
> Last Updated: 2026-03-09
> Scope: Full product target, delivery phases, recurring-plan semantics, and end-state success criteria.

Related:
- [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)
- [System Architecture V7](SYSTEM_ARCHITECTURE_V7.md)
- [Execution And Manifest Library V7](EXECUTION_AND_MANIFEST_LIBRARY_V7.md)
- [Intent To Manifestation And Team Interaction V7](INTENT_TO_MANIFESTATION_AND_TEAM_INTERACTION_V7.md)
- [UI And Operator Experience V7](UI_AND_OPERATOR_EXPERIENCE_V7.md)
- [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md)

## 1. Product End State

Mycelis is not targeting a chat app with tools attached. It is targeting a governed execution platform where operators can:
- ask for an answer
- request a governed change
- instantiate and monitor execution
- create recurring or always-on plans
- inspect what happened and recover safely

Every operator-facing action must terminate in one of four valid states:
- `answer`
- `proposal`
- `execution_result`
- `blocker`

Planning-only narration is not a valid product outcome.

## 2. Full Target Delivery

Full target delivery is achieved when all of the following are true.

### 2.1 Operational foundation

- local and cluster runtime flows are deterministic
- mounted storage supports persistent manifested material
- `uv run inv ...` is the canonical management surface
- teardown and restart behavior do not leave hanging processes behind
- signals, tasks, and docs stay synchronized

### 2.2 Runtime and architecture coherence

- execution, manifests, runs, signals, and storage use one consistent contract
- team coordination uses canonical NATS lanes
- invalid manifests fail before activation
- run lineage is queryable from intent through result

### 2.3 UI and operator clarity

- the UI leads the operator through intent, proposal, execution, observation, and recovery
- system detail is progressive disclosure
- recurring plans are visible as recurring, not disguised one-shot tasks
- degraded states explain what failed and what to do next

### 2.4 Delivery discipline

- docs, tests, runtime, and release gates all describe the same system
- acceptance is based on delivered behavior, not implementation effort
- phase transitions are gated and evidenced

### 2.5 Integration coherence

- internal tools, MCP tools, and third-party API capabilities must be exposed through one execution contract
- Soma-first intent flow must remain consistent when execution is delegated to created teams
- created teams must remain inspectable and interactable after manifestation

## 3. Plan Operating Modes

Some target plans are one-time outcomes. Others are designed to remain alive.

Every plan, workflow, or manifest must explicitly declare one of these modes:

| Mode | Meaning | Expected behavior |
| --- | --- | --- |
| `one_shot` | run once and terminate | completes, fails, or is cancelled |
| `scheduled` | run repeatedly on time rules | survives restart, fires on schedule, maintains history |
| `persistent_active` | stay continuously active | maintains subscriptions/watchers and recovers after restart |
| `event_driven` | wait for matching events/signals | dormant until trigger, then activates with lineage |

This mode must be visible in:
- manifests
- UI labels
- runtime scheduling/activation
- tests

## 4. Plan State Model

Every plan/workflow/manifest must use a clear lifecycle:
- `draft`
- `validated`
- `proposed`
- `approved`
- `scheduled`
- `active`
- `paused`
- `completed`
- `failed`
- `cancelled`

Not every mode uses every state, but the system must not collapse them into a single vague “running/planning” condition.

## 5. Delivery Phases

The canonical phase order remains:
- `P0` operational foundation and gate discipline
- `P1` logging, error handling, and hot-path cleanup
- `P2` meta-agent-owned manifest pipeline
- `P3` workflow-composer onboarding and execution-facing UI
- `P4` release hardening and promotion gates

## 6. Success Criteria By Product Surface

### 6.1 Workspace

- direct questions return answers
- governed changes return proposals
- dispatched work returns visible execution results
- failures return blockers with recovery actions

### 6.2 Automations

- scheduled, event-driven, and persistent-active plans are clearly distinguished
- operators can pause, resume, inspect, and cancel them

### 6.3 Runs

- every execution has lineage, state, events, and result closure

### 6.4 Resources

- provider, tool, and filesystem state supports actual execution, not only configuration

### 6.5 System

- health and degraded state are understandable without requiring the operator to reverse-engineer backend internals

### 6.6 Team interaction and module integration

- operators can interact directly with created teams via governed channels
- team communications are inspectable by run/team/agent scope
- module bindings are explicit and auditable across `internal`, `mcp`, `openapi`, and `host` adapter classes
- intent can be promoted from one-shot execution into durable manifestation without changing the operator mental model

## 7. Anti-Targets

The product is not considered delivered if it does any of the following:
- returns planning text as if work was done
- hides recurring intent inside ad hoc manual reruns
- treats high-volume telemetry as the main operator interface
- requires the operator to understand internal implementation details to recover from a failure
- lets documentation become a second system that disagrees with runtime

## 8. Immediate Design Implication

The system must be optimized around this sequence:
1. express intent
2. understand the chosen execution path
3. approve when needed
4. observe the result
5. recover or refine
6. persist the plan if it should recur or remain active

Next:
- [System Architecture V7](SYSTEM_ARCHITECTURE_V7.md)
- [Execution And Manifest Library V7](EXECUTION_AND_MANIFEST_LIBRARY_V7.md)
- [Intent To Manifestation And Team Interaction V7](INTENT_TO_MANIFESTATION_AND_TEAM_INTERACTION_V7.md)
- [UI And Operator Experience V7](UI_AND_OPERATOR_EXPERIENCE_V7.md)
