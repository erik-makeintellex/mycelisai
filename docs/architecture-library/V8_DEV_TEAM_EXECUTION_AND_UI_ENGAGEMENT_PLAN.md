# V8 Dev Team Execution and UI Engagement Plan

> Status: ACTIVE
> Last Updated: 2026-04-02
> Owner: Product Management / Delivery Coordination
> Purpose: Coordinate the dev team around the true-MVP finish push with explicit execution order, ownership, UI testing expectations, engagement-testing expectations, and committed-slice discipline.

## Why This Plan Exists

Mycelis is no longer blocked mainly by repo hygiene or missing foundations.

The current challenge is coordinated finish work:

- content and artifact requests must return obvious value
- approval must feel like trust instead of bureaucracy
- visible operator settings must behave like real product features
- UI/browser proof must test product value, not only rendering
- partner-demo and engagement testing must confirm that Mycelis feels worth paying for

This plan exists so the dev team can execute those lanes in one order without drifting between runtime, UI, docs, and testing.

## Target Outcome

This plan is successful when all of the following are true:

- the true-MVP finish lane advances through small committed slices
- each slice has a named owner and a named proof set
- UI and engagement testing expectations are explicit before implementation lands
- default-path product value improves without flattening advanced capability
- release readiness becomes a convergence exercise instead of a cleanup scramble

## Team Structure

### 1. Product Management / Delivery Coordination

Responsibilities:

- own lane order
- keep acceptance criteria stable
- reject slices that improve mechanics but weaken product value
- keep `V8_DEV_STATE.md` synchronized with committed checkpoints

Required outputs:

- canonical slice order
- blocker classification
- go / no-go recommendation for true-MVP and demo lanes

### 2. Product Narrative and Trust Team

Responsibilities:

- keep user-facing wording legible
- align approval, artifact, and continuity language across UI and docs
- protect the one-persistent-Soma story

Required outputs:

- approved wording for answer, proposal, execution result, and blocker states
- approval-story and artifact-result language aligned across workspace, docs, and tests

### 3. Interface and Workflow Team

Responsibilities:

- implement default-path UI changes
- preserve advanced access through explicit advanced surfaces and details controls
- keep artifact/result return visible in the Soma path

Required outputs:

- UI implementation for active slices
- matching component/store/browser coverage
- no orphaned or misleading visible controls

### 4. Runtime and Governance Team

Responsibilities:

- preserve approval integrity while improving product legibility
- keep answer-vs-proposal behavior aligned with policy and request intent
- prevent runtime drift from undermining trust behavior

Required outputs:

- runtime contract changes for content/artifact posture
- approval-policy mapping and metadata truth
- live-backend fixes when real execution behavior drifts from the product contract

### 5. Settings and Operator Controls Team

Responsibilities:

- audit visible settings for dead controls, partial wiring, and misleading defaults
- keep operator-visible controls bounded and real
- preserve advanced configuration behind intentional boundaries

Required outputs:

- dead-control audit
- end-to-end fixes for visible settings gaps
- updated settings/browser proof for each repaired control

### 6. Memory, Continuity, and Trust Team

Responsibilities:

- keep durable memory, temporary continuity, and trace semantics distinct
- ensure continuity feels valuable and bounded in product terms
- protect trust framing in continuity and approval surfaces

Required outputs:

- operator-legible continuity wording
- continuity and retention behavior validation
- no silent conflation of durable memory and temporary workspace continuity

### 7. QA / UI Testing Agentry Team

Responsibilities:

- verify product value, not only component mechanics
- classify issues as `product`, `runtime`, `environment`, `test`, or `docs`
- prove stable mocked behavior, live-backend behavior, and partner-demo readiness

Required outputs:

- updated workflow verification expectations
- browser proof after each meaningful slice
- engagement-testing verdicts for partner-demo and trust lanes

### 8. Release and Ops Team

Responsibilities:

- keep clean-environment proof trustworthy
- separate product failures from environment failures
- maintain process, port, and runtime discipline

Required outputs:

- clean bring-up notes
- service-health and live-browser proof
- release-readiness notes for each major lane

## Non-Negotiable Execution Rules

1. Do not flatten Mycelis into a demo shell.
2. Do not remove advanced capability; move complexity behind advanced surfaces or inspectable details.
3. Do not treat all specialist-model, MCP, or cross-agent collaboration as manual approval by default.
4. Do not ship slices that add visible value ambiguity.
5. Do not leave mixed local worktrees as the primary state container; package slices into commits.

## Execution Order

### Wave 1: Product Value and Trust

Focus:

- approval/product-trust simplification
- content and artifact visible-value delivery
- result-surfacing after confirm/execute

Definition of done:

- default proposal surfaces are user-legible
- content requests return inline value or explicit artifact/result references
- browser proof covers answer vs proposal vs execution result clearly

### Wave 2: Visible Settings and Operator Controls

Focus:

- remove or repair dead settings controls
- confirm assistant identity, theme, advanced-mode boundaries, AI Engine visibility, and continuity inspectability are real features

Definition of done:

- no visible default-path settings control is decorative or misleading
- stable and live proof cover the repaired settings paths

### Wave 3: Demo and Engagement Hardening

Focus:

- partner-demo sequence
- product narrative consistency
- continuity and trust visibility

Definition of done:

- the partner-demo lane reaches `READY` or `READY_WITH_NOTES`
- manual trust checks confirm the product reads as product first

### Wave 4: Clean Release Proof

Focus:

- stable mocked browser proof
- live governed browser proof
- clean startup, health, and environment validation from committed state

Definition of done:

- evidence can be rerun from clean state without hidden manual repair steps

## UI Testing Expectations

The UI team must prove:

- Central Soma is legible and primary
- AI Organizations read as governed working contexts
- direct requests land in the correct terminal state
- governed mutation remains trustworthy
- settings and continuity survive refresh, re-entry, and recovery
- advanced depth stays reachable without polluting the default path

Required terminal states:

- `answer`
- `proposal`
- `execution_result`
- `blocker`

Required UI proof layers per meaningful slice:

1. focused component/store tests for the changed state contract
2. route-level browser proof for the affected default workflow
3. live-backend proof when `/api/v1/chat`, `/api/v1/intent/confirm-action`, or real governed execution behavior changes
4. docs/state synchronization in the same slice

Canonical browser references:

- `docs/architecture-library/V8_UI_TESTING_AGENTRY_EXECUTION_RUNBOOK.md`
- `docs/architecture-library/V8_UI_TEAM_FULL_TEST_SET.md`
- `docs/architecture-library/V8_UI_WORKFLOW_VERIFICATION_PLAN.md`

## Engagement Testing Expectations

Engagement testing is required for product-worth-paying-for judgment, not only release mechanics.

The engagement lane must prove:

- a technical evaluator can understand the product in minutes
- approval feels like confidence, not friction
- continuity feels stateful and valuable
- the platform feels deeper than a single-thread assistant
- advanced reveal strengthens confidence instead of exposing drift

Required engagement evidence:

1. partner-demo verification against `V8_PARTNER_DEMO_VERIFICATION_CHECKLIST.md`
2. manual trust pass across direct answer, governed mutation, execution proof, and continuity
3. explicit `PASS`, `SOFT_FAIL`, or `HARD_FAIL` outcome per major demo step

## Slice Template

Every execution slice must record:

- `Goal`
- `Teams involved`
- `Files expected to change`
- `Expected terminal state impact`
- `Stable proof commands`
- `Live/backend proof commands`, if applicable
- `Docs/state updates required`

Every slice ends only when:

- code is landed
- tests are rerun
- docs are updated where meaning changed
- `V8_DEV_STATE.md` is updated
- the repo returns to a clean committed checkpoint

## Immediate Execution Stack

### Slice 1: UI and engagement testing contract expansion

Goal:

- align the UI-team and partner-demo testing contracts with the actual true-MVP target, including compose/demo preflight, collaboration/media result proof, and visible settings expectations

Expected owner set:

- QA / UI Testing Agentry Team
- Product Management / Delivery Coordination
- Demo Scenario Team

Minimum proof:

- `$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q`
- state/doc synchronization

### Slice 2: Visible settings audit and first repair

Goal:

- identify and repair the highest-confidence visible settings control that is still dead, partial, or misleading

Expected owner set:

- Settings and Operator Controls Team
- Interface and Workflow Team
- QA / UI Testing Agentry Team

Minimum proof:

- focused settings tests
- `uv run inv interface.test`
- `uv run inv interface.typecheck`
- docs/state sync

### Slice 3: Content/artifact visible-value delivery follow-through

Goal:

- extend the current content/artifact lane so durable outputs always return as obvious inline value or explicit artifact references

Expected owner set:

- Product Narrative and Trust Team
- Interface and Workflow Team
- Runtime and Governance Team
- QA / UI Testing Agentry Team

Minimum proof:

- focused workspace/proposal/content tests
- stable browser proof for answer vs proposal vs artifact result
- live-backend proof when governed execution behavior changes

## Acceptance

This plan remains `ACTIVE` until:

- Wave 1 through Wave 4 each have committed, validated checkpoints
- the true-MVP finish plan acceptance bullets are satisfied
- UI and engagement testing both say the product is legible, trustworthy, and structurally deep
