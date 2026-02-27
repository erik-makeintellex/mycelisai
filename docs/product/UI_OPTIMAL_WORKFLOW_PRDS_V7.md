# UI Optimal Workflow PRDs (V7)

> Purpose: planning authority for optimal end-user workflows across Workspace, Automations, Resources, System, Teams, and Runs.
> Scope: frontend UX planning, interaction model, failure behavior, operator guidance.
> Out of scope: backend schema/API implementation details.

---

## Table of Contents

1. Planning Goals
2. User Archetypes
3. Global Workflow Model
4. PRD-W1: Workspace Intent to Execution
5. PRD-W2: Failure Recovery and Degraded Continuity
6. PRD-W3: Automations Chain Setup
7. PRD-W4: Resources Capability and Actuation Readiness
8. PRD-W5: System Health to Corrective Action
9. PRD-W6: Teams Diagnosis and Intervention
10. PRD-W7: Run Investigation and Causal Clarity
11. PRD-W8: Governance and Approval Throughput
12. Cross-Workflow Standards
13. Success Metrics and Release Gates
14. Parallel Delivery Workstreams

---

## 1. Planning Goals

- Remove dead-end surfaces.
- Make every critical workflow executable in 3 or fewer meaningful decisions.
- Keep failure states actionable without route-switching.
- Ensure users always understand:
  - Current system state
  - Immediate next action
  - Expected result of that action
- Optimize for operational confidence, not visual minimalism.

---

## 2. User Archetypes

### A1. Operator (Primary)
- Needs: runtime confidence, quick diagnosis, controlled actuation.
- Risk: blocked by unclear status or hidden dependencies.

### A2. Builder
- Needs: compose automations, teams, and resources with predictable outcomes.
- Risk: configuration drift and uncertain impact.

### A3. Reviewer/Governance
- Needs: fast triage of approval-required actions with context.
- Risk: delayed decisions, low trust in evidence.

### A4. Incident Responder
- Needs: identify degraded subsystems and recover service quickly.
- Risk: fragmented signals across pages.

---

## 3. Global Workflow Model

Canonical user journey:

1. Define intent (Workspace/Automations)
2. Validate capability availability (Resources)
3. Execute with governed controls (Workspace/Approvals)
4. Observe outcomes (Runs/System/Teams)
5. Recover or iterate (same surface, no dead ends)

Global UX contracts:

- Every surface answers:
  - What can I do here?
  - What is the current state?
  - What should I do next?
- Every failure state provides:
  - Reason
  - Impact
  - Retry or fallback action
  - Diagnostics copy/open path

---

## 4. PRD-W1: Workspace Intent to Execution

### User Outcome
User can go from message to governed execution without leaving Workspace.

### Entry Surface
- `Workspace` (`/dashboard`)

### Planning Requirements
- Keep Soma as default route.
- Expose direct council routing as optional, not required.
- Preserve context on retries and reroutes.
- Show proposal-to-confirm lifecycle inline.

### Optimal Interaction Flow
1. User submits prompt.
2. UI shows active mode/brain/gov.
3. If response requires mutation, render structured proposal block.
4. User confirms/cancels without leaving chat.
5. On confirm, show run link and immediate progress cues.

### Failure Design
- Council failure shown as structured action card with:
  - retry
  - switch to Soma
  - continue Soma-only
  - copy diagnostics

### Acceptance Criteria
- No raw backend error visible in message stream.
- Retry does not force message re-entry.
- Proposal flow always yields explicit next action.

---

## 5. PRD-W2: Failure Recovery and Degraded Continuity

### User Outcome
User can continue work during subsystem degradation with informed tradeoffs.

### Entry Surfaces
- Global banner and status drawer
- Workspace and System pages

### Planning Requirements
- Single source of truth for service health.
- Global degraded indicator with recovery controls.
- Clear auto-clear behavior after health restoration.

### Optimal Interaction Flow
1. Degradation detected.
2. Banner appears with reason + actions:
  - Retry
  - Open Status
  - Switch to Soma path
3. Status drawer lists per-subsystem state and likely impact.
4. Recovery action returns user to prior context.

### Acceptance Criteria
- Banner appears/disappears automatically.
- Recovery action does not require page reload.
- User reaches diagnosis in 2 interactions or fewer.

---

## 6. PRD-W3: Automations Chain Setup

### User Outcome
User creates an actionable automation chain without ambiguity.

### Entry Surface
- `Automations` (`/automations`)

### Planning Requirements
- Keep an actionable hub when scheduler is unavailable.
- Guide user from trigger to approval to execution path.
- Eliminate blank or warning-only states.

### Optimal Interaction Flow
1. Open Automations hub.
2. Select primary action: trigger, approvals, teams, wiring.
3. Follow explicit sequence:
  - create trigger
  - set mode (propose/execute)
  - route to team
  - validate approval path
  - execute/monitor

### Acceptance Criteria
- No empty-state dead zone.
- First automation chain setup starts in 1 click.
- Users can find run outcomes from the same surface.

---

## 7. PRD-W4: Resources Capability and Actuation Readiness

### User Outcome
User can quickly answer: what can this system access and execute now?

### Entry Surface
- `Resources` (`/resources`)

### Planning Requirements
- Brains, tools, workspace, capabilities each provide status + next action.
- Workspace explorer acts as practical filesystem capability proof.
- MCP/tools surfaces clarify installed/connected/error.

### Optimal Interaction Flow
1. Check brains/provider readiness.
2. Check MCP tool availability.
3. Validate filesystem operation path.
4. Confirm capability templates for execution teams.

### Acceptance Criteria
- User can verify capability readiness in 2 clicks.
- Missing capability states provide install/connect action.
- Resource state mapping is consistent with System status.

---

## 8. PRD-W5: System Health to Corrective Action

### User Outcome
User can diagnose and recover from infrastructure issues directly from System.

### Entry Surface
- `System` (`/system`)

### Planning Requirements
- Distinguish checks, health metrics, and lifecycle actions.
- Support quick copy of recovery commands.
- Align all health tabs to shared service-status semantics.

### Optimal Interaction Flow
1. Observe high-level health.
2. Run targeted quick checks.
3. Inspect service-specific cards.
4. Execute corrective command path.
5. Re-check without leaving page.

### Acceptance Criteria
- Health tabs avoid contradictory statuses.
- Every degraded/failure state includes recovery command.
- Check timestamps are visible and meaningful.

---

## 9. PRD-W6: Teams Diagnosis and Intervention

### User Outcome
User can identify team/agent health and move to intervention quickly.

### Entry Surface
- `Automations -> Teams`

### Planning Requirements
- Team cards show online count, heartbeat context, and health status.
- Detail drawer provides logs/runs/wiring navigation.
- Team filtering supports standing vs mission teams.

### Optimal Interaction Flow
1. Filter team scope.
2. Identify degraded/offline team.
3. Open detail drawer.
4. Jump to run logs or wiring path.

### Acceptance Criteria
- Unreachable team diagnosis in 2 clicks or fewer.
- Team actions are obvious and persistent.

---

## 10. PRD-W7: Run Investigation and Causal Clarity

### User Outcome
User can explain what happened, why, and what to do next.

### Entry Surfaces
- `Runs` (`/runs`)
- `Run Detail` (`/runs/{id}`)

### Planning Requirements
- Preserve conversation/events tabs with clear semantics.
- Make run status and recency obvious in list.
- Support quick pivot back to workspace actions.

### Optimal Interaction Flow
1. Select run from recent list.
2. Inspect conversation for intent/decision context.
3. Inspect events for execution trace.
4. Decide follow-up action (retry, refine prompt, policy update, team update).

### Acceptance Criteria
- Run detail remains readable under high event volumes.
- Tabs are stable and clearly differentiated.

---

## 11. PRD-W8: Governance and Approval Throughput

### User Outcome
Reviewer can approve/reject safely with sufficient context and low delay.

### Entry Surfaces
- `Automations -> Approvals`
- Workspace structured proposal/error blocks

### Planning Requirements
- Provide intent/risk/evidence summary before decision.
- Reduce context-switching for common approve/reject actions.
- Surface governance impact in run/workspace context.

### Optimal Interaction Flow
1. Open pending approval.
2. Review summarized context and diagnostics.
3. Approve/reject with clear consequence messaging.
4. Observe immediate state update.

### Acceptance Criteria
- Decision latency reduced by clear context presentation.
- Approval actions visibly propagate to execution state.

---

## 12. Cross-Workflow Standards

- Shared status vocabulary:
  - healthy
  - degraded
  - failure
  - offline
  - informational
- Shared empty-state template:
  - what is missing
  - why it is missing
  - next action (primary)
  - optional alternatives
- Shared failure template:
  - what happened
  - likely cause
  - impact
  - recovery actions
- Shared diagnostics actions:
  - copy diagnostics
  - open status
  - retry

---

## 13. Success Metrics and Release Gates

Workflow-level target KPIs:

- Task start-to-first-successful-action under 90 seconds for primary workflows.
- Failure recovery success rate above 80 percent without leaving current surface.
- Time-to-diagnosis for critical degradation under 60 seconds.
- Drop in unresolved UI dead-end sessions release-over-release.

Release gates:

1. All workflow PRD acceptance criteria mapped to tests.
2. Degraded-state recoverability verified in E2E.
3. No contradictory health states between pages.
4. Docs and in-app help updated for changed workflows.

---

## 14. Parallel Delivery Workstreams

### Workstream A: Global Reliability UX
- status, banner, drawer, SSE continuity

### Workstream B: Workspace Execution UX
- intent flow, structured errors, proposal continuity

### Workstream C: Workflow Surfaces
- automations/resources/teams directed actions

### Workstream D: Observability UX
- runs/system diagnostics clarity and performance

### Workstream Q: QA and Regression
- gate evidence, degraded-state reliability, cross-surface consistency

Planning handoff rule:
- each workstream must map implementation tasks to one or more PRD sections above before starting build.
- implementation sequencing and team ownership are controlled by `docs/product/UI_WORKFLOW_INSTANTIATION_AND_BUS_PLAN_V7.md` and `docs/ui-delivery/PARALLEL_DELIVERY_BOARD.md`.
