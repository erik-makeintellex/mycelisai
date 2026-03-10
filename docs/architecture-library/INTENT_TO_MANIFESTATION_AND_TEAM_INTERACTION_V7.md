# Intent To Manifestation And Team Interaction V7

> Status: Canonical
> Last Updated: 2026-03-10
> Scope: Unified contract for Soma-first manifestation, module integration (internal/MCP/third-party APIs), created-team communication, and operator interaction from intent to durable execution.

Related:
- [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)
- [Target Deliverable V7](TARGET_DELIVERABLE_V7.md)
- [Execution And Manifest Library V7](EXECUTION_AND_MANIFEST_LIBRARY_V7.md)
- [UI And Operator Experience V7](UI_AND_OPERATOR_EXPERIENCE_V7.md)
- [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md)

Supporting specialized docs:
- [NATS Signal Standard V7](../architecture/NATS_SIGNAL_STANDARD_V7.md)
- [Soma Team Channel Architecture V7](../architecture/SOMA_TEAM_CHANNEL_ARCHITECTURE_V7.md)
- [UI Target And Transaction Contract V7](../architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md)
- [Workflow Composer Delivery Plan V7](../architecture/WORKFLOW_COMPOSER_DELIVERY_V7.md)
- [Universal Action Interface V7](../architecture/UNIVERSAL_ACTION_INTERFACE_V7.md)

## 1. Design Positioning

Mycelis should not behave like a generic node-editor automation tool.

Mycelis should behave like a governed execution platform where:
- user intent is first-class
- Soma/teams reason about execution strategy
- execution is policy-gated where required
- outcomes are observable and recoverable
- created teams remain interactable after manifestation

## 2. Core Differentiator vs Common Automation Tools

Common automation products optimize for static workflow wiring.

Mycelis optimizes for:
- intent-first orchestration (chat or API to structured execution path)
- governance-first mutation control (`proposal` and confirmation gates)
- team-native execution (standing + created agentry)
- unified execution surface across internal tools, MCP tools, and third-party APIs
- persistent run/event/conversation lineage

### 2.1 Comparison snapshot (n8n-style vs Mycelis target)

| Concern | Common automation model (n8n-style) | Mycelis target model |
| --- | --- | --- |
| Primary authoring mode | graph/node-first | intent-first with Soma thread as default; graph is advanced mode |
| Decision boundary | workflow engine executes graph directly | Soma emits terminal state (`answer`/`proposal`/`execution_result`/`blocker`) before execution |
| Governance posture | optional approvals per workflow | default proposal gating for high-impact mutation paths |
| Execution unit | workflow/job | `run_id` + team expression + module binding |
| Integration abstraction | connector catalog per platform | unified `Action Module` abstraction (`internal`/`mcp`/`openapi`/`host`) |
| Team model | single workflow context | standing teams + created teams with dedicated communication surfaces |
| Observability | run logs and node traces | mission events + run timeline + conversation turns + channel inspector |
| Operator intervention | rerun/stop | interject/reroute/pause/resume/cancel with run-linked lineage |

## 3. Unified Intent Stack

The product should be delivered as one stack with three tightly-coupled layers:

### 3.1 Brain layer

Responsibilities:
- intent interpretation
- role/provider routing
- decision framing (`answer`, `proposal`, `execution_result`, `blocker`)
- execution strategy generation

### 3.2 Nervous layer

Responsibilities:
- canonical signal transport
- run-linked event emission
- team/channel routing
- transient vs durable signal separation

### 3.3 Experience layer

Responsibilities:
- operator intent entry
- manifestation review and confirmation
- created-team interaction
- run/timeline/conversation inspection
- recovery and interjection controls

## 4. Canonical Intent -> Manifestation Flow

Every execution-capable intent should follow this sequence:

1. intake:
- user enters intent via Workspace/API
- request context is normalized with source metadata

2. decision frame:
- Soma returns one of: `answer`, `proposal`, `execution_result`, `blocker`
- planning-only text is invalid
- default conversation posture is Soma-direct (`consultation_mode=none`) for routine intents
- specialist consultation is triggered only by explicit planning/architecture/delivery asks or high-complexity/high-impact conditions

3. manifestation plan:
- if execution is needed, Soma emits structured Team Expressions
- each expression maps objective, constraints, and execution binding

4. module resolution:
- each actionable step binds to a module contract (internal/MCP/API/host)

5. governance preflight:
- high-impact operations require proposal + confirm token flow

6. activation:
- execution creates or links a `run_id`
- events are persisted and emitted with canonical metadata

7. interaction and recovery:
- operator can inspect, interject, reroute, pause, or cancel where valid

8. closure and continuity:
- result is visible in timeline/conversation/artifacts
- intent pattern can be promoted to reusable manifest/recipe/workflow

## 5. Module Integration Contract

To support internal capability plus third-party expansion, execution should use one abstraction: `Action Module`.

Each module record should define:
- `module_id`
- `adapter_kind` (`internal`, `mcp`, `openapi`, `host`)
- `display_name`
- `description`
- `input_schema`
- `output_schema`
- `auth_profile_ref`
- `risk_class`
- `idempotency_policy`
- `timeout_ms`
- `retry_policy`
- `rate_limit_policy`
- `audit_policy_ref`
- `enabled`

Execution invariants:
- UI and governance operate on module contracts, not adapter-specific details
- adapter-specific behavior is hidden behind the executor boundary
- module failure states are normalized into operator-readable blockers

## 6. Team Expression Contract

Soma should emit Team Expressions as the primary intermediate representation between chat intent and manifested agentry.

Minimum Team Expression fields:
- `expression_id`
- `team_id` (or pending identifier before manifestation)
- `objective`
- `role_plan`
- `inputs`
- `outputs`
- `module_bindings`
- `policy_scope`
- `risk_level`
- `expected_outcome`
- `success_signal`
- `fallback_strategy`

Team Expressions must be:
- human-readable
- machine-executable
- versioned
- promotable to reusable manifests/templates

## 7. Created Team Interaction Model

Created teams require first-class operator interaction, not hidden backend-only behavior.

### 7.1 Team Workspace surface

For each created team, provide tabs:
- `Overview`
- `Communications`
- `Members`
- `Manifest`
- `Controls`

### 7.2 Communications model

The Communications surface should unify:
- directed team commands
- team status/result signals
- run events
- conversation turns
- module call traces

### 7.3 Channel Inspector

Required filters:
- `run_id`
- `team_id`
- `agent_id`
- `source_kind`
- `source_channel`
- `payload_kind`
- time range

Required controls:
- send directed instruction
- interject active run
- pause/resume/cancel valid execution modes
- copy diagnostics bundle

## 8. Channel Contract For Team Interaction

Use canonical product channels only:
- inbound team command: `swarm.team.{team_id}.internal.command`
- operator status: `swarm.team.{team_id}.signal.status`
- bounded result: `swarm.team.{team_id}.signal.result`
- high-volume machine telemetry: `swarm.team.{team_id}.telemetry`
- run fanout: `swarm.mission.events.{run_id}`

Operator-facing UI must prioritize:
- `signal.status`
- `signal.result`
- run events/conversation

Telemetry channels must remain drill-down only.

## 9. Required Metadata Envelope

Every governed product signal exposed in operator flows must include:
- `run_id` (execution-linked)
- `team_id` (team-scoped)
- `agent_id` (agent-scoped where applicable)
- `source_kind`
- `source_channel`
- `payload_kind`
- `timestamp`

If these fields are absent, the signal is incomplete for operator UX and should not be shown as authoritative.

## 10. Soma-First Manifestation GUI Model

The intended GUI baseline is a Soma-first manifestation thread, not a forced graph-only editor.

### 10.1 Layout

Three-pane model:
- left: Team Expressions palette and module capabilities
- center: manifestation thread (intent -> expression blocks -> outcome)
- right: governance + run inspector

### 10.2 Interaction sequence

1. user asks Soma Architect for objective
2. Soma emits Team Expressions
3. operator edits/reorders/constraints expressions inline
4. operator binds modules for each expression
5. operator confirms governed changes
6. system manifests run/team(s) and opens live inspection

### 10.3 Relationship to DAG composer

The DAG composer remains a higher-complexity surface.

Soma-first manifestation thread is the default path for:
- speed
- clarity
- governance continuity
- lower operator cognitive load

## 11. Integration Impact (Current vs Target)

Current implementation strengths (`COMPLETE`):
- event spine, run timeline, trigger engine, conversation log, proposal gating, interjection
- standardized NATS channel posture and source metadata policy

Active checkpoint (`ACTIVE`, 2026-03-09):
- proposal payload contract now carries structured `team_expressions` + `module_bindings` across Soma and direct council chat mutation paths
- operator proposal surfaces now expose expression counts and binding-level adapter context before confirmation
- store normalization now derives teams/agents/tools from Team Expression payloads when explicit aggregate values are absent

Current gaps (`REQUIRED`/`NEXT`):
- scheduler runtime and `scheduled_missions` persistence path
- causal chain UI surfaces
- created-team first-class communications workspace
- unified module registry UX across internal/MCP/OpenAPI/host adapters

## 12. Delivery Status Map

| Area | Status | Notes |
| --- | --- | --- |
| Soma terminal-state contract | `COMPLETE` | answer/proposal/execution_result/blocker model active |
| Run/event lineage | `COMPLETE` | mission runs/events and timeline APIs are live |
| Trigger execution model | `COMPLETE` | guarded trigger engine delivered |
| Scheduler recurring execution | `NEXT` | design locked, runtime/UI pending |
| Causal chain operator UI | `REQUIRED` | backend support exists, operator surface pending |
| Created-team communications workspace | `REQUIRED` | new surface and aggregation APIs required |
| Unified Action Module registry UX | `REQUIRED` | contract exists in architecture direction, execution path needs productization |

## 13. Proof Requirements

For slices implementing this contract, proof must include:
- UI tests for manifestation-thread terminal states
- backend tests proving module execution envelope normalization
- channel tests proving canonical subject families and metadata presence
- integration tests proving created-team command round-trip and operator-visible status/result
- product-flow tests from intent to manifested run with run-linked evidence

## 14. Canonical Constraints

- never collapse execution into planning-only narration
- never expose telemetry as the default operator outcome surface
- never hardcode ad hoc channel subjects in runtime code
- never treat MCP-only capability as the total integration model
- never mark scheduler/chain/created-team surfaces as complete before behavior and tests are delivered

## 15. Immediate Design Work Order

1. complete scheduler and recurring execution surfaces
2. ship causal chain UI
3. ship created-team communications workspace and channel inspector
4. unify module registry and binding UX in Soma-first manifestation thread
5. converge composer and manifestation thread around one shared execution contract

## 16. Visual Architecture Update

### 16.1 Control-plane view

```text
Operator Intent (Workspace/API)
  -> Soma Architect (decision frame)
  -> Team Expressions (editable + governed)
  -> Action Module Binding (internal/mcp/openapi/host)
  -> Governance Preflight (proposal/confirm when required)
  -> Activation (run_id creation/link)
  -> Team Workspace (communications + controls + outcomes)
```

### 16.2 Runtime signal view

```text
Intent Input
  -> web_api/workspace_ui envelope
  -> Brain Layer (strategy + terminal state)
  -> Nervous Layer
      -> swarm.team.{team_id}.internal.command
      -> swarm.team.{team_id}.signal.status
      -> swarm.team.{team_id}.signal.result
      -> swarm.mission.events.{run_id}
  -> Experience Layer
      -> manifestation thread
      -> channel inspector
      -> timeline + conversation + artifacts
```

## 17. MVP Target State (Intent-to-Manifestation)

MVP is complete only when all items below are true:

1. `COMPLETE`: Soma terminal-state contract + proposal gating in default flow.
2. `COMPLETE`: run timeline and event lineage (`run_id` evidence from intent to result).
3. `ACTIVE` to `IN_REVIEW`: Soma-first manifestation thread with editable Team Expressions.
4. `ACTIVE` to `IN_REVIEW`: Action Module binding on each expression with normalized failure states.
5. `REQUIRED`: created-team workspace with communications and controls tabs.
6. `REQUIRED`: channel inspector with `run_id/team_id/agent_id/source_kind/source_channel/payload_kind/time` filters.
7. `NEXT`: scheduler recurring execution path (`scheduled_missions`) aligned to trigger model.
8. `REQUIRED`: causal chain operator UI linked from runs view.

Anything less than this set is pre-MVP for the full intent-to-manifestation architecture target.

Next:
- [Next Execution Slices V7](NEXT_EXECUTION_SLICES_V7.md)
- [Workflow Composer Delivery Plan V7](../architecture/WORKFLOW_COMPOSER_DELIVERY_V7.md)
- [UI Target And Transaction Contract V7](../architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md)
