# Soma Symbiote Growth and Host Actuation Architecture V7

Version: `1.0`
Status: `Authoritative`
Last Updated: `2026-02-28`
Scope: Soma thought configuration, backend compiled contracts, long-term learning growth, and localhost host actuation

This document defines how Soma should think, decide, learn, and actuate over time while staying governed and local-first.

---

## Table of Contents

1. Purpose
2. Soma Thought Configuration Model
3. Backend Compiled Contracts (Soma Form)
4. Symbiote Growth Pattern (Learning Over Time)
5. Localhost Host Actuation Model (OpenClaw-Like)
6. Governance and Safety Requirements
7. Observability and Feedback Loops
8. API Surface (Planned)
9. Testing Matrix
10. Timeline Fit
11. Local Ollama Communication Contract
12. Parallel Delivery Alignment

---

## 1. Purpose

Enable Soma to:
1. run direct actions when safe
2. manifest teams when planning depth is required
3. improve through memory/recipe/feedback loops
4. actuate host capabilities locally under strict controls

---

## 2. Soma Thought Configuration Model

### 2.1 Thought Profile

Soma behavior is configured via explicit profiles:
- `thought_depth`: `shallow|standard|deep`
- `planning_mode`: `direct_first|team_first|adaptive`
- `consultation_policy`: `none|targeted|full_council`
- `memory_policy`: `minimal|contextual|aggressive_recall`
- `risk_posture`: `conservative|balanced|aggressive` (policy-bounded)

### 2.2 Decision Policy

For each request, Soma chooses:
- direct action path
- manifested team path
- defer/propose path

Selection inputs:
- risk level
- complexity score
- repeat intent
- hardware/IoT channel requirements
- governance mode

### 2.3 Operator Co-Thought Contract

Soma should always emit a concise structured reasoning summary for execution decisions:
- chosen path (direct vs team)
- key blockers/assumptions
- required approvals (if any)
- expected outputs and run trace

---

## 3. Backend Compiled Contracts (Soma Form)

Soma form must be represented by typed backend contracts.

### 3.1 Core structs (planned)

- `SomaThoughtProfile`
- `SomaDecisionFrame`
- `SomaExecutionPlan`
- `SomaLearningSignal`
- `SomaReflectionRecord`

### 3.2 Required decision frame fields

- `request_id`
- `run_id`
- `path_selected` (`direct|manifest_team|propose`)
- `team_lifetime` (`ephemeral|persistent|auto`)
- `risk_level`
- `confidence`
- `approval_required`

### 3.3 Event binding

All decisions and plan transitions emit run-linked mission events for replay/audit.

---

## 4. Symbiote Growth Pattern (Learning Over Time)

Growth ladder:

### Stage 0 - Baseline Identity
- stable Soma profile
- deterministic default decision policies

### Stage 1 - Episodic Capture
- capture structured conversation + execution outcomes
- tag success/failure/fallback patterns

### Stage 2 - Recipe Distillation
- synthesize reusable inception recipes from successful runs
- attach quality score and usage metrics

### Stage 3 - Adaptive Selection
- use recipe quality + recent outcomes to bias planning path selection
- still policy-bounded; no autonomous policy mutation

### Stage 4 - Symbiote Extension Loop
- continuous suggestion engine for improved routing/tool/team choices
- operator confirmation required for high-impact behavior shifts

Learning constraints:
- no silent policy escalation
- all adaptation changes produce reviewable artifacts

---

## 5. Localhost Host Actuation Model (OpenClaw-Like)

### 5.1 Default mode

- host actuation is localhost-first
- bound to local control gateway
- no remote host actuation by default

### 5.2 Host Actuation Adapter (planned)

Adapter supports controlled host operations through allowlisted actions only.

Action classes:
- process/service control (bounded)
- local file/workspace operations
- local diagnostics/health probes
- optional local device bridge calls

### 5.3 OpenClaw-style controls to adopt

- typed gateway protocol
- strict handshake with role/scope
- idempotency keys for side effects
- device/session identity with pairing
- replay-window protections

### 5.4 Non-goals by default

- arbitrary shell command execution
- unrestricted OS mutation
- remote host control without security-gated enablement

---

## 6. Governance and Safety Requirements

1. High-risk actuation requires proposal + explicit approval.
2. Host actuation actions are allowlisted and schema-validated.
3. Thought profile updates are governed mutations.
4. Symbiote adaptation proposals are auditable and reversible.
5. Emergency halt path must preempt active actuation loops.

---

## 7. Observability and Feedback Loops

Required visibility:
- thought profile in effect
- path decision per request
- fallback occurrences
- learning deltas (what changed, why)
- host actuation attempt/success/failure metrics

All with `run_id` linkage and timeline visibility.

---

## 8. API Surface (Planned)

### 8.1 Soma thought/profile APIs

- `GET /api/v1/soma/profile`
- `PUT /api/v1/soma/profile`
- `POST /api/v1/soma/decide`
- `GET /api/v1/soma/learning/signals`
- `POST /api/v1/soma/learning/review/{id}`

### 8.2 Host actuation APIs (localhost-first)

- `GET /api/v1/host/actions`
- `POST /api/v1/host/actions/{id}/invoke`
- `GET /api/v1/host/status`
- `POST /api/v1/host/bridge/pair`

These APIs map into universal action contracts and security profile enforcement.

---

## 9. Testing Matrix

### Unit
- thought decision policy selection tests
- profile validation tests
- learning signal classification tests

### Integration
- direct vs team path routing tests
- recipe-driven adaptation tests
- host adapter allowlist + rejection tests

### API
- profile CRUD + decision endpoints
- actuation invoke authorization and approval-gate tests

### E2E
- user request -> Soma path decision -> execution trace
- user request -> team manifestation -> execution trace
- scheduled repeat promotion and persistent team behavior
- localhost host actuation with governance confirmation

---

## 10. Timeline Fit

This architecture lands in existing rollout waves:
- Wave 2: Soma thought profile contracts + decision APIs scaffolded
- Wave 3: localhost host actuation adapter under secure gateway controls
- Wave 4: symbiote growth loop activation with review/rollback controls

Remote host actuation remains disabled until secure remote actuation gates pass.

---

## 11. Local Ollama Communication Contract

Soma growth depends on high-confidence local cognition. Ollama must therefore be treated as an explicit subsystem, not an implicit provider detail.

### 11.1 Required health dimensions

1. endpoint reachability
2. configured model availability
3. recent inference p95 latency
4. recent token throughput
5. failure ratio over rolling window

### 11.2 Required runtime policy

- Healthy local provider:
  - local-first routing remains active for eligible roles.
- Degraded local provider:
  - fallback profile may engage, and fallback state is surfaced in UI and logs.
- Unavailable local provider:
  - high-risk autonomous pathways are blocked unless policy explicitly permits remote fallback.

### 11.3 Planned API alignment

- `GET /api/v1/cognitive/status` must include local readiness fields.
- `GET /api/v1/system/services-status` must expose local provider state in global health map.
- `GET /api/v1/soma/profile` should expose local-first/fallback behavior toggles.

### 11.4 Operator visibility requirements

- Status drawer shows current local model and health state.
- Degraded banner includes local-provider failure as a first-class reason.
- System quick checks can run an on-demand local provider probe.

---

## 12. Parallel Delivery Alignment

This document maps to the parallel program as follows:

- Lane A (decision/memory): Sections 2, 3, 4
- Lane B (action runtime): Sections 5, 8
- Lane C (lifecycle/schedule): Sections 2.2, 4, 10
- Lane D (operator UX): Sections 7, 11.4
- Lane Q (QA/reliability): Section 9 + degraded/fallback drills

No lane is complete until its associated tests and operator-facing diagnostics are in place.

---

End of document.
