# V8 Approval and Product Trust Strike Team Plan

> Status: ACTIVE
> Last Updated: 2026-03-30
> Owner: Product Management / Delivery Coordination
> Purpose: Simplify Mycelis approval/auth-style interactions so the product feels understandable and trustworthy to normal operators while preserving governed depth, advanced capability, and policy-configurable collaboration.

## Why This Lane Exists

Mycelis is now strong enough technically that the next trust failures are mostly product-shape failures:

- approval cards expose too much internal detail by default
- users are being asked to parse governance mechanics before they see value
- content requests can still drift into tool/approval posture without clearly returning visible content
- specialist/model/MCP collaboration is powerful but not yet explained in a user-legible way
- the default product path can still feel like an operator console instead of a governed AI Organization product

This lane exists to correct those issues without flattening the platform into a demo shell.

## Non-Negotiable Product Rule

The goal is not to remove structure.

The goal is:

- simple default posture
- visible delivered value
- governed mutation
- preserved advanced depth
- inspectable execution details when wanted

Mycelis must remain a powerful structured platform. This lane only simplifies how that power is presented and approved in the default path.

## Target Outcome

This lane is successful when all of the following are true:

- a normal user can understand a proposal card in seconds
- direct planning/drafting/review requests stay inline when that is the best posture
- durable artifacts are proposed or executed with clear reasoning and clear result references
- not all specialist/model/MCP collaboration is treated as manual-approval work by default
- advanced execution details remain available behind inspection controls instead of being dumped into the first-view surface
- Central Soma continues to feel like one persistent counterpart operating across governed contexts

## Core Product Principles

### 1. Value Before Mechanics

The default product path should answer:

- what Soma is doing
- why it matters
- what will happen next
- what the user must decide

It should not lead with:

- internal tool names
- execution expressions
- module binding language
- engine/runtime internals

### 2. Governance Must Feel Like Trust, Not Friction

Approvals should communicate:

- what action is proposed
- why approval is needed
- what will be changed or created
- expected risk and cost
- what happens after approval

### 3. Collaboration Is Policy-Bound, Not Blanket-Blocked

Not every MCP call, specialist-model interaction, or cross-agent collaboration should require manual approval.

Approval posture must be configurable by:

- capability risk
- mutation impact
- external exposure
- cost/spend threshold
- integration trust level
- organization and user governance profile

### 4. Content Requests Must Deliver Content

If a user asks for content, the system must provide one of:

- readable inline content
- a clearly referenced durable artifact
- a clearly described proposed artifact with understandable rationale

Using a tool or another model is not success by itself. Visible value return is required.

## Scope

### In Scope

- approval/proposal card default UX
- result-surfacing after confirm/execute
- answer-vs-proposal mode selection for content requests
- user-legible wording for governed execution
- specialist/model/MCP collaboration posture in the UI and docs
- Central Soma trust framing where approval decisions appear
- UI-testing and product-verification expectations for this behavior

### Out of Scope

- removing advanced governance metadata from the platform entirely
- weakening approval enforcement on risky mutation/external action paths
- removing advanced routes or advanced inspectability
- full enterprise IAM redesign
- replacing the broader V8 universal-Soma PRD

## Team Structure

### Product Management / Delivery Coordination

Responsibilities:

- own the lane
- keep scope aligned with the product-worth-paying-for goal
- prevent demo simplification from becoming capability regression
- maintain dependency order and acceptance gates

Deliverables:

- approved phase boundaries
- final release recommendation for this lane

### Product Narrative and Trust Team

Responsibilities:

- define the operator-facing approval story
- align wording across dashboard, workspace, docs, and test expectations
- eliminate user-visible internal jargon where it is not needed

Deliverables:

- canonical proposal-card wording contract
- canonical explanation text for answer vs proposal vs artifact posture

### Interface and Workflow Team

Responsibilities:

- implement the new default approval surface
- preserve access to advanced execution details behind inspectable controls
- improve artifact/result reference behavior after execution

Deliverables:

- updated proposal-card UI
- updated result-return UI
- updated Central Soma / organization workflow touchpoints where needed

### Runtime and Governance Team

Responsibilities:

- preserve approval integrity while simplifying UI presentation
- review how proposal metadata is mapped into operator-facing surfaces
- fix live direct-answer/runtime issues that distort approval behavior

Deliverables:

- canonical metadata map: required for default UI, required for details panel, runtime-only
- live runtime fixes needed for stable trust behavior

### Content Generation and Collaboration Team

Responsibilities:

- truth-map content/media/file requests into the right execution posture
- ensure collaboration results return through Soma clearly
- define when specialist/model/MCP collaboration should be inline, proposed, or approval-gated

Deliverables:

- request taxonomy
- result-surfacing contract
- collaboration posture matrix

### QA / UI Testing Agentry Team

Responsibilities:

- verify that simplified approval UX still preserves governance integrity
- verify content-value delivery and collaboration posture
- classify findings correctly as product, environment, test, or docs

Deliverables:

- updated approval-focused UI test checklist
- browser proof evidence for each phase

### Release and Ops Team

Responsibilities:

- keep clean-environment proofs from being confused with product issues
- validate the lane in compose and managed browser paths
- confirm health/logging/monitoring surfaces remain useful after changes

Deliverables:

- compose/runtime validation notes
- release readiness verdict for this lane

## Phase Plan

### Phase 0: Truth Mapping

Status target: ACTIVE -> COMPLETE

Goals:

- map every current approval surface and content-delivery posture
- identify where value delivery currently fails before or after approval
- identify which technical metadata should move out of the default surface

Required outputs:

- proposal-card field inventory:
  - default-visible
  - details-only
  - runtime-only
- content request taxonomy:
  - inline answer
  - governed artifact
  - optional approval
  - required approval
- collaboration posture matrix:
  - auto-allowed
  - approval optional
  - approval required

Acceptance:

- every major approval/content path is classified
- no open ambiguity about whether “approval needed” is a product issue or a policy issue

### Phase 1: Approval Surface Simplification

Status target: NEXT -> IN_REVIEW

Goals:

- simplify the proposal card to a user-first decision surface
- move technical execution details behind a collapsed details section

Default visible contract should answer:

- what Soma wants to do
- why approval is needed
- what will be created or changed
- risk
- expected cost if relevant
- approve or cancel

Technical details moved behind inspection:

- internal tool IDs
- module binding/expression detail
- engine/runtime detail
- low-level execution metadata not needed for the decision

Acceptance:

- a non-technical operator can understand the proposal without needing architecture knowledge
- governance logic and execution integrity remain unchanged

### Phase 2: Content and Artifact Value Delivery

Status target: NEXT -> IN_REVIEW

Goals:

- ensure content requests visibly produce content
- ensure artifact requests visibly produce or reference artifacts
- ensure collaboration results are surfaced back through Soma

Acceptance:

- drafting/review/planning requests prefer inline value when appropriate
- durable artifacts are clearly described and clearly referenced after execution
- the user never has to infer whether collaboration produced anything useful

### Phase 3: Central Soma Trust Integration

Status target: NEXT -> IN_REVIEW

Goals:

- make approval moments feel like a Central Soma trust interaction, not a detached technical workflow
- align dashboard/workspace/organization wording to one persistent Soma across governed contexts

Acceptance:

- default path still feels like one persistent Soma
- organizations remain governed work contexts, not separate Soma identities

### Phase 4: Runtime and Live Proof Hardening

Status target: NEXT -> IN_REVIEW

Goals:

- fix live runtime defects that break trust on fresh flows
- keep stable and live browser proof aligned

Required focus:

- compose live fresh-organization direct-answer failure
- any remaining stale test assumptions about approval/content posture

Acceptance:

- stable browser proof green
- live governed browser proof green
- no accepted blocker remains in the approval/content trust lane

### Phase 5: Release Gate

Status target: NEXT -> COMPLETE

Goals:

- prove the lane in product, docs, and runtime

Required gates:

- updated docs and in-app docs manifest
- targeted page/component tests
- targeted browser proof
- compose runtime proof
- final UI testing agentry verdict of `READY` or `READY_WITH_NOTES`

## Default Approval Card Contract

The default approval surface should read more like:

- `Soma wants to create a file`
- `Why approval is needed: this writes to your workspace`
- `Result: one new file in workspace/logs`
- `Risk: Medium`
- `Estimated cost: 0.35`
- `Approve`
- `Cancel`
- `Details`

It should not lead with raw internal categories unless the user explicitly expands details.

## Required Deliverables by Team

### Product Narrative and Trust Team

- one wording contract for proposal cards
- one wording contract for result return after execution
- one wording rule set for answer vs proposal language

### Interface and Workflow Team

- new proposal-card implementation
- details expander for advanced metadata
- user-facing artifact reference/preview behavior where applicable

### Runtime and Governance Team

- metadata mapping contract
- live direct-answer fix if still failing in compose
- compatibility review to ensure UI simplification does not weaken governance

### Content Generation and Collaboration Team

- request taxonomy
- collaboration posture matrix
- result-surfacing rules

### QA / UI Testing Agentry Team

- approval simplification checklist
- content/artifact delivery proof
- mixed-finding classification guidance

### Release and Ops Team

- clean compose proof
- monitoring/logging review for this lane
- final operator run notes

## Acceptance Criteria

This lane is accepted only if all of the following are true:

- approval cards are easier to understand without losing governance integrity
- content requests return visible value
- artifact requests return visible result references
- policy-configurable collaboration is documented and testable
- advanced details remain reachable but non-default
- Central Soma framing remains coherent
- stable and live proofs both support the new product story

## Risks

### Risk: Oversimplification

Danger:

- hiding important governance signals

Mitigation:

- move details behind explicit inspection instead of deleting them

### Risk: Demo Shell Regression

Danger:

- simplifying the default path by removing platform depth

Mitigation:

- preserve advanced routes, docs, and inspectability

### Risk: Runtime/Product Confusion

Danger:

- misclassifying live failures as UX failures or vice versa

Mitigation:

- require strict classification discipline in browser and compose runs

## Immediate Next Actions

1. complete the approval/content truth map from current UI and runtime behavior
2. design the default-visible vs details-only proposal-card contract
3. implement the first approval-card simplification slice
4. fix or confirm the compose live fresh-organization direct-answer defect
5. rerun targeted UI/browser/runtime proof
