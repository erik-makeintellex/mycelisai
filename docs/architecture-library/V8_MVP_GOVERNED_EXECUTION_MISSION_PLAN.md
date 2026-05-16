# V8 MVP Governed Execution Mission Plan
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Library](ARCHITECTURE_LIBRARY_INDEX.md) | [V8.2 Full Production Architecture](../../architecture/v8-2.md)

> Status: Canonical
> Last Updated: 2026-05-16
> Purpose: Convert the governed-execution doctrine into executable MVP missions with scope, dependencies, governance implications, UI manifestation, emitted events, and proof requirements.

## Mission Authority

This plan is the manifestation-phase bridge between doctrine and delivery.

Planning is valid only when it produces:
- executable missions
- measurable delivery targets
- governed execution paths
- UI-visible outcomes
- validation requirements

Canonical MVP execution flow:

```text
Operator Intent
  -> Execution Contract
  -> Governance Evaluation
  -> Capability Invocation
  -> Event Spine Recording
  -> Observable Workspace State
  -> Proof Generation
  -> Mission Reconstruction
```

## MVP Objective

The initial MVP proves a stable governed execution organism, not maximum capability.

The operator must be able to issue intent, authorize execution, observe orchestration, interrupt when necessary, and receive proof-backed outcomes that can be reconstructed after the fact.

## Mission Set

### Mission 1: Minimum Deployable Governed Runtime

Status: `COMPLETE`

Execution scope:
- define the minimum runtime needed for a single-host governed execution path
- keep Windows as edit/git surface, Rancher Desktop K3s as the Windows local Kubernetes parity proof lane, and WSL/Compose as the guarded deployment-mimic proof surface
- keep `qwen3:14b` as the primary Soma/self-hosted model and specialist models behind governed routing
- produce a repeatable release proof from `origin/main`

Dependencies:
- Rancher Desktop K3s through `MYCELIS_K8S_BACKEND=rancher`
- WSL/Compose proof through `uv run inv wsl.refresh` and `uv run inv wsl.validate --lane=release` when Compose deployment-mimic evidence is required
- explicit reachable Ollama or OpenAI-compatible AI endpoint such as `http://<windows-ai-host>:11434/v1`
- WSL/Compose relay to the explicit AI endpoint when that proof lane is used
- Compose Postgres, NATS, Core, Interface, SearXNG, and storage health

Governance implications:
- local owner/break-glass posture must be explicit
- no raw secrets in Compose topology or docs
- no runtime path is accepted until text inference, API health, and operator UI proof are visible

UI manifestation:
- Windows browser reaches `http://localhost:3000`
- System/health surfaces show degraded states rather than raw backend noise
- Soma uses the configured primary model without hiding provider failure

Emitted events:
- deployment health check event
- text inference availability event
- storage health event
- release proof start/completion event

Proof requirements:
- `uv run inv wsl.validate --lane=release` passes from the refreshed WSL checkout
- Compose health reports Core, Frontend, Brains API, Telemetry, Cognitive text, Postgres, and NATS online or explicitly degraded
- Windows GUI probe returns `http://localhost:3000 [200]`
- state file records the accepted commit and proof evidence

Current state:
- Rancher Desktop K3s is the accepted Windows local Kubernetes release-parity proof lane for the MVP RC.
- WSL/Compose remains secondary deployment-mimic evidence when required.
- New runtime work should preserve this proof posture instead of reopening Mission 1 as the active gate.

### Mission 2: Workspace Operational Surface

Status: `IN_REVIEW`

Execution scope:
- make Workspace the primary live operational surface for governed execution
- keep Soma adjacent to execution state, approvals, capability use, and proof
- remove or demote disconnected workflow surfaces that make execution feel hidden

Dependencies:
- V8 UI/API and Operator Experience Contract
- Directed Execution UI/runtime alignment directive
- current Organization Workspace and Dashboard Soma surfaces
- run timeline and output/proof components

Governance implications:
- approval state must be visible before and during execution
- risky actions must show scope, reason, capability, and expected effect
- interrupt/cancel/retry posture must be legible

UI manifestation:
- active mission/workspace header shows intent, stage, involved systems, approvals, and proof state
- Soma response shows execution contract summary before governed action
- run/proof links stay reachable from the conversation and Workspace

Emitted events:
- intent captured
- execution contract proposed
- approval requested/approved/denied
- execution started/progressed/completed/blocked
- proof attached

Proof requirements:
- browser proof covers direct answer, governed proposal, approval/cancel, tool-backed execution, and proof review
- UI tests assert no hidden completion state for governed actions
- operator can reconstruct what changed from Workspace and Run surfaces

### Mission 3: Event Spine As Canonical Truth

Status: `IN_REVIEW`

Execution scope:
- promote the Event Spine as the reconstruction and audit truth for MVP execution
- define the minimum event taxonomy for intent, governance, capability, execution, output, and proof
- ensure mutating actions emit persistent mission events in addition to transient bus signals

Dependencies:
- NATS Signal Standard
- mission events persistence
- run timeline APIs and UI
- execution summary contract
- managed exchange and artifact/output records

Governance implications:
- actions without reconstructable events are untrusted
- event metadata must identify source, scope, intended consumer, and proof linkage
- high-volume telemetry must not substitute for operator status or audit proof

UI manifestation:
- Run timeline shows ordered event chain
- Event inspector summarizes state transitions in operator language
- proof links bind events to outputs, artifacts, audit records, or retained exchange items

Emitted events:
- `intent.received`
- `execution.contract.created`
- `governance.evaluation.completed`
- `capability.invocation.requested`
- `capability.invocation.completed`
- `workspace.state.updated`
- `proof.generated`
- `mission.reconstructed`

Proof requirements:
- event schema tests cover required metadata
- API tests verify mutating actions persist mission events
- browser proof reconstructs one mission from event history without relying on transient chat text

### Mission 4: Capability Governance And Approval Enforcement

Status: `IN_REVIEW`

Execution scope:
- route all meaningful tool/API/script/MCP use through governed capability records
- enforce deny-by-default posture for medium/high-risk actions
- connect capability use to execution contract, approval state, Event Spine records, and proof

Dependencies:
- Capability Manifest and Runtime Integration Standard
- `/api/v1/capabilities`
- connected tools UI
- governance decisions and approval flows
- internal tool execution handlers

Governance implications:
- capability invocation must include risk, scope, actor, target, approval posture, and audit intent
- no prompt or template grants authority by itself
- local filesystem, host commands, external APIs, and MCP tools must stay scoped and auditable

UI manifestation:
- Connected Tools shows what Soma can use, why it can use it, and what approval is required
- Workspace execution contract lists capabilities before invocation
- approval cards show capability risk and expected state changes

Emitted events:
- capability selected
- governance evaluation requested
- approval required
- approval decision recorded
- capability invocation started/completed/blocked
- audit/proof attached

Proof requirements:
- server tests prove blocked, approved, and read-only capability paths
- UI tests prove approval state is visible before execution
- no capability-backed result is accepted without execution summary and event linkage

### Mission 5: Proof-Backed Mission Completion

Status: `ACTIVE`

Execution scope:
- define the MVP completion standard for direct Soma work, governed proposals, team/group outputs, and capability-backed actions
- reject completion that lacks proof, retained output, or reconstruction path
- normalize outcomes into runs, exchange, artifacts, audit, or learning candidates as appropriate

Dependencies:
- execution summary contract
- run and output records
- managed exchange
- artifact storage
- audit log entries
- Workspace and Run proof UI
- Operator trust package UI
- `audit_recovery.degradation` for failed, blocked, or partial outcomes

Governance implications:
- completion is a state transition, not a model assertion
- proof must identify what changed, what was used, what risk was accepted, and where evidence lives
- blocked or partial outcomes must be first-class results
- failed completion must name what failed, what remains trusted, what proof is invalid, and what can be retried

UI manifestation:
- Soma shows result status, proof, retained output, next step, recovery path, and degraded-trust boundaries
- Workspace shows completion state without burying proof in logs
- Run detail can reconstruct intent, approval, capability use, output, and evidence

Emitted events:
- output retained
- artifact generated
- audit record written
- proof generated
- completion accepted
- completion blocked
- reconstruction requested/completed

Proof requirements:
- tests cover completed, blocked, partial, and failed proof states
- tests cover `execution_summary`, Operator trust package rendering, and `audit_recovery.degradation` fields
- failed approved execution returns failed run/proof/audit metadata instead of only a flat API error
- search blockers preserve provider blocker code and next action as degradation proof
- browser proof demonstrates retained-output review after refresh
- release proof includes at least one governed action with reconstructable completion

## Cross-Mission Conflict Compression

Compress or defer:
- architecture that does not name an execution pathway
- capability ideas without governance and proof
- UI panels that do not improve operator understanding of live execution
- autonomy that cannot be interrupted or reconstructed
- speculative abstractions that increase delivery surface without strengthening MVP proof

Keep:
- Soma orchestration clarity
- Event Spine truth
- capability governance
- Workspace execution visibility
- proof-backed completion
- Rancher K3s release-parity proof plus guarded WSL/Compose proof when required

## Initial Execution Sequence

1. `COMPLETE` Mission 1: keep the committed-tree Rancher K3s RC proof lane green, with WSL Compose retained as secondary deployment-mimic evidence.
2. `IN_REVIEW` Mission 3: keep event/reconstruction checks attached to governed actions, capability use, retained outputs, and failed/degraded runs.
3. `IN_REVIEW` Mission 2: keep Workspace, Soma, Runs, Groups, and retained-output surfaces aligned around one visible execution story.
4. `IN_REVIEW` Mission 4: finish capability/MCP confirmation proof so explicit MCP tool refs are visibly governed and retained as proof-backed outputs.
5. `ACTIVE` Mission 5: prove retained project-package output through the live Mycelis GUI path, including operator review, storage/open controls, and reconstruction after refresh.

## Acceptance Standard

The MVP governed execution slice is accepted only when a real operator workflow proves:
- intent became an execution contract
- governance evaluated the action
- capability use was bounded and visible
- the Event Spine recorded reconstructable truth
- Workspace showed live state and proof
- the result was retained, reviewable, and reconstructable
- the release path is deployable from a clean proof checkout
