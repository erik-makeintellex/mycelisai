# UI And Operator Experience V7

> Status: Canonical
> Last Updated: 2026-03-10
> Scope: Operator journeys, UI targets, information hierarchy, anti-swarm rules, and intuitive interaction design.

Related:
- [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)
- [Target Deliverable V7](TARGET_DELIVERABLE_V7.md)
- [Execution And Manifest Library V7](EXECUTION_AND_MANIFEST_LIBRARY_V7.md)
- [Intent To Manifestation And Team Interaction V7](INTENT_TO_MANIFESTATION_AND_TEAM_INTERACTION_V7.md)
- [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md)

Supporting specialized docs:
- [UI Target And Transaction Contract V7](../architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md)
- [Frontend](../architecture/FRONTEND.md)
- [Next Execution Slices V7](NEXT_EXECUTION_SLICES_V7.md)

## 1. Primary Design Problem

Mycelis is at risk of becoming an information swarm:
- too many surfaces
- too much raw system detail
- too many internal concepts shown before user purpose
- planning narration presented as if it were product value

The UI must instead behave like an intentional execution environment.

## 2. Design Objective

The operator should always be able to answer:
- what can I do here
- what is happening now
- what needs my decision
- what failed
- what should I do next

If a screen cannot answer those clearly, it is not ready.

## 3. Global UX Rules

### 3.1 Terminal-state rule

Execution-facing interactions must end in:
- `answer`
- `proposal`
- `execution_result`
- `blocker`

No planning-only terminal state.

### 3.2 Progressive disclosure rule

Operational detail should be layered:
- primary: intent, outcome, next action
- secondary: run state, proposal, health, blockers
- tertiary: telemetry, raw envelopes, deep diagnostics

### 3.3 One dominant job per surface

Each primary screen should optimize for one operator job, not several mixed together.

### 3.4 Recovery visibility rule

When the system degrades, the UI must expose:
- what failed
- what still works
- what the operator can do immediately

### 3.5 Soma-First conversation economy rule

Workspace chat should default to Soma-only execution for normal interaction.

Council engagement should be triggered only when:
- operator explicitly targets a council member
- operator explicitly asks to plan, architect, or deliver
- request crosses complexity/risk thresholds that require specialist review

This preserves token efficiency and keeps routine interaction direct.

### 3.6 Theme simplification rule

Workspace must be legible and intention-first before it is information-dense.

Required layout posture:
- conversation and terminal outcome are dominant
- secondary controls are visible but compact
- diagnostics are progressive-disclosure, not always-on density
- advanced telemetry remains gated behind explicit advanced actions

## 4. Primary Operator Journeys

### 4.1 Ask or instruct

Surface:
- Workspace

Goal:
- convert operator intent into answer/proposal/execution result/blocker

### 4.2 Review and decide

Surface:
- proposal blocks and approvals

Goal:
- inspect risk and approve/reject clearly

### 4.3 Observe execution

Surface:
- runs and timeline

Goal:
- understand active or completed work without reading raw telemetry first

### 4.4 Manage recurring plans

Surface:
- automations/workflows

Goal:
- create, pause, resume, inspect, or retire scheduled/event-driven/persistent-active plans

### 4.5 Recover the system

Surface:
- status and system views

Goal:
- identify failure and take the next correct action

### 4.6 Interact with created teams

Surface:
- team workspace and communications inspector

Goal:
- let operators direct, inspect, and refine created-team execution without dropping to raw channel tooling

## 5. Canonical Screen Model

### 5.1 Workspace

Primary job:
- intent entry and immediate outcome

Must emphasize:
- conversation
- proposal/result cards
- current execution mode
- relevant next action

Must demote:
- deep telemetry
- low-signal system internals
- dense multi-widget dashboards that obscure the active conversation outcome

Workspace conversation policy:
- default route is Soma-first and direct
- council consultation is intent-triggered, not unconditional
- user should see when and why consultation happened

### 5.2 Automations

Primary job:
- manage plans that persist or recur

Must distinguish clearly:
- one-shot templates
- scheduled plans
- event-driven plans
- persistent-active plans

### 5.3 Runs

Primary job:
- inspect specific execution closure and lineage

Must emphasize:
- state
- cause
- outcome
- related child/parent links

### 5.4 Resources

Primary job:
- manage execution capability

Must emphasize:
- whether providers/tools/filesystem are usable
- not just how they are configured

### 5.5 System

Primary job:
- health and recovery

Must emphasize:
- what is degraded
- what is safe to continue
- what command or action should be taken

### 5.6 Team Workspace

Primary job:
- inspect and steer created teams after manifestation

Must emphasize:
- team objective and current run linkage
- communication stream grouped by status/result/conversation/events
- immediate controls (interject, pause/resume, reroute, cancel where valid)
- channel and module transparency without exposing operator to low-signal telemetry by default

## 6. Anti-Swarm Rules

The UI should avoid:
- showing raw NATS or telemetry streams as the default experience
- presenting configuration concepts before operator goals
- mixing chat, orchestration, diagnostics, and docs equally in one dense surface
- asking users to infer result state from side information

Instead:
- show the result state directly
- allow drill-down into technical detail when needed
- bias default layouts toward action and clarity

## 7. UI For Recurring And Persistent Plans

The UI must visibly differentiate plans that continue to matter after first activation.

Required cues:
- operating mode badge
- current lifecycle state
- last run or last heartbeat
- next run or current active-watch status
- pause/resume controls where valid

This prevents recurring work from being mistaken for completed work.

## 8. Interaction Language

Use product language that maps to operator intent:
- Ask
- Propose
- Approve
- Run
- Pause
- Resume
- Recover
- Inspect

Avoid leading with internal implementation terms unless the user drills deeper.

When execution is delegated, keep language stable:
- "Team Expression"
- "Module Binding"
- "Manifest"
- "Run"
- "Result"
- "Blocker"

## 9. UI Optimization Principle

The correct design is not the interface that shows the most information.

It is the interface that gets the operator to the correct next decision with the least ambiguity.

## 10. What To Build Next

The next UI-heavy slices should be judged against this order:
1. Workspace answer/proposal/result/blocker clarity
2. Workspace density/theme simplification under progressive-disclosure rules
3. Launch Crew and workflow onboarding clarity
4. Automations clarity for scheduled and persistent plans
5. created-team workspace and communications clarity
6. Runs and chain inspection clarity
7. System recovery clarity

Next:
- [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md)
