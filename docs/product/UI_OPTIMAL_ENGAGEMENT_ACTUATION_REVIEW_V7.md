# UI Engagement and Actuation Review (V7)

> Purpose: review of existing UI behavior with planning recommendations for cleaner directed engagement and reliable feature actuation.
> Baseline assessed: current `Workspace`, `Automations`, `Resources`, `System`, `Runs`, `Teams` surfaces.

---

## Table of Contents

1. Executive Summary
2. Current Strengths
3. Engagement Friction Review
4. Actuation Friction Review
5. Workflow-Level Improvement Plan
6. IA and Navigation Refinements
7. Interaction and Content Design Refinements
8. Reliability and State Consistency Plan
9. Test Planning Matrix
10. Handoff Package for Delivery Teams

---

## 1. Executive Summary

Current UI has strong operational foundations:
- central command surface (Workspace)
- global reliability primitives (degraded banner, status drawer)
- explicit workflow routes (Automations, Resources, System, Runs)

Main planning gap:
- users still encounter fragmented guidance between decision surfaces and actuation surfaces.

Primary planning objective:
- tighten directed engagement so each feature path leads to execution confidence, not interpretation work.

---

## 2. Current Strengths

- Workspace is workflow-first and keeps proposal/failure handling inline.
- Global status and degraded controls are available across routes.
- Resources now includes filesystem capability proof via workspace explorer.
- System page includes quick checks and service card actions.
- Runs page gives a clear route into execution history.

---

## 3. Engagement Friction Review

### F1. Decision context is split across multiple tabs
- Impact: users infer next actions from multiple places.
- Planning response: add a per-surface "Next Best Action" strip driven by current state.

### F2. Some surfaces still require interpretation effort
- Impact: operators spend time determining what is blocked vs available.
- Planning response: enforce state templates with explicit availability and fallback actions.

### F3. Microcopy consistency varies
- Impact: low confidence in severity and urgency.
- Planning response: standardize severity copy tokens for healthy/degraded/failure/offline.

### F4. Tiny metadata text reduces scanability
- Impact: slower diagnosis on dense operational cards.
- Planning response: enforce minimum readable caption size and line-height in standards.

---

## 4. Actuation Friction Review

### A1. Capability readiness path is not fully explicit
- Impact: user can still ask "can I run this now?" after visiting Resources/System.
- Planning response: introduce `Actuation Readiness` summary block:
  - provider ready
  - tool ready
  - governance mode
  - stream health
  - run visibility

### A2. Cross-surface follow-through links are incomplete
- Impact: users jump between routes manually.
- Planning response: add deterministic pivots:
  - Teams -> Runs
  - Runs -> Workspace follow-up
  - System failure -> specific recovery route

### A3. Scheduler and advanced automations are partially staged
- Impact: users may interpret planned features as broken features.
- Planning response: roadmap-card pattern with explicit "available now" alternatives.

---

## 5. Workflow-Level Improvement Plan

### P1. Workspace
- Add persistent next-step recommendations under proposal/failure blocks.
- Add one-click "open related run" where applicable.

### P2. Automations
- Keep hub pattern as default landing.
- Add completion checklist states for chain setup progress.

### P3. Resources
- Add readiness badge row across tabs:
  - brains
  - tools
  - workspace
  - capabilities

### P4. System
- Keep checks + services aligned to one status source.
- Add "last recovered" timestamp for degraded subsystems.

### P5. Teams and Runs
- Add quick pivots for triage:
  - "view latest failed run"
  - "open team logs"
  - "inspect related approval"

---

## 6. IA and Navigation Refinements

Planning decisions:

- Preserve 5 core nav panels as primary operational IA.
- Keep advanced surfaces gated, but maintain clear explanatory copy when hidden.
- Standardize tab ordering by user intention:
  - Observe
  - Decide
  - Act
  - Verify

Suggested intent-first ordering examples:

- System:
  - Health
  - Services
  - NATS
  - Database
  - Matrix
  - Debug
- Resources:
  - Brains
  - MCP Tools
  - Workspace Explorer
  - Capabilities

---

## 7. Interaction and Content Design Refinements

### Directed Engagement Components (Planning Additions)

- `NextActionStrip`
  - per surface
  - state-driven
  - one primary, up to two secondary actions

- `ActuationReadinessCard`
  - cross-domain readiness summary
  - clear pass/fail for execution confidence

- `RecoveryPlaybookPopover`
  - quick command and route hints for common failures

### Copy and Messaging Standards

- Replace ambiguous labels with state-specific copy.
- Use "what this means" microcopy on complex panels.
- Keep action labels verb-first:
  - Retry Check
  - Open Status
  - Route to Team
  - Inspect Run

---

## 8. Reliability and State Consistency Plan

Planning requirement:
- all stateful status indicators must use shared status semantics and a central source where possible.

Consistency checks:

1. Does this component derive health from shared state?
2. Does degraded behavior match banner/drawer messaging?
3. Does retry action have deterministic expected outcome?
4. Does resolved health clear warnings automatically?

---

## 9. Test Planning Matrix

### Unit
- State mapping for each major surface state.
- Next-action availability per state.

### Integration
- Shared status source consumed consistently across System/Workspace indicators.
- Resource readiness cards map backend payloads correctly.

### E2E
- Intent -> proposal -> confirm -> run -> inspection path.
- Degraded -> retry -> recovered path.
- Team issue -> run diagnosis path.

### Reliability
- NATS/SSE/db/provider partial failures.
- Inline recovery without navigation dead ends.

---

## 10. Handoff Package for Delivery Teams

For each feature PR:

1. Map PR to one workflow PRD section in `UI_OPTIMAL_WORKFLOW_PRDS_V7.md`.
2. State which friction items from this review are addressed.
3. Include before/after interaction map.
4. Include test evidence for degraded and success paths.
5. Include lane ownership and gate target from `../ui-delivery/PARALLEL_DELIVERY_BOARD.md`.

Execution order recommendation:

1. Reliability/state consistency fixes.
2. Directed engagement components.
3. Cross-surface actuation pivots.
4. Microcopy and typography consistency pass.
5. Final regression and acceptance evidence.
