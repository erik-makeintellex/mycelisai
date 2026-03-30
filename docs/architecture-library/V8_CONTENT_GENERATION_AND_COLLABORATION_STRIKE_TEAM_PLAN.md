# V8 Content Generation And Collaboration Strike Team Plan

> Status: ACTIVE
> Owner: Product Management / Delivery Coordination
> Last Updated: 2026-03-30
> Primary Goal: Make Soma reliably deliver visible content, governed artifact generation, and policy-configurable specialist/model collaboration without flattening Mycelis into a narrow chat tool.

---

## Mission

Mycelis must not fail a content request by stopping at approval theater.

When an operator asks Soma for content, the product must:

- deliver inline content when inline content satisfies the request
- generate durable artifacts when the operator actually wants a saved/exported asset
- allow Soma and council/specialist units to collaborate with other agent/model types when policy allows it
- surface the result back into the Soma conversation clearly enough that the operator understands what happened

This strike team exists to close the gap between:

- tool invocation
- approval posture
- actual delivered user value

---

## Non-Negotiable Rules

1. Do not reduce Soma capability.
2. Do not treat all MCP use, external model interaction, or specialist collaboration as manual-approval-only by default.
3. Approval must be configurable by policy, risk, integration trust, cost, data exposure, and mutation impact.
4. Inline content should be preferred when the operator is asking to inspect, compare, brainstorm, review, or draft.
5. Durable artifact generation should be used when the operator is asking to save, export, package, render, or hand off a real asset.
6. If a specialist/model/tool path is used, the result must still come back through Soma in a readable way.
7. The operator must always be able to answer:
   - what was created
   - where it was created
   - why approval was or was not required
   - what contributed to the result

---

## Product Problem Statement

Current drift observed by review and testing:

- Soma can choose proposal/tool paths too early for content requests that would be better served inline.
- Some generated-content flows do not surface the final result clearly enough in the conversation.
- Approval prompts can appear without making the delivered outcome obvious enough.
- Media/file/image/content workflows need clearer policy behavior so collaboration remains fluid when low-risk and governed when higher-risk.

This is a product-value problem, not only a runtime problem.

---

## Target Product Behavior

### 1. Direct content path

Use `answer` when the operator asks for:

- summaries
- drafts
- planning
- comparisons
- reviewable tables
- brainstormed concepts
- copy ideation
- strategy framing

### 2. Governed artifact path

Use `proposal` when the operator asks for:

- file creation
- saved/exported content
- images or other generated media saved as assets
- packaged deliverables
- privileged tool use
- higher-cost generation
- external/model-integrated execution that policy marks as review-worthy

### 3. Policy-configured collaboration path

Allow fluid collaboration without mandatory manual approval when policy allows low-risk:

- council to specialist planning exchange
- model-to-model content drafting
- low-risk MCP-backed analysis
- marketing/media ideation from product-supplied intent

Require approval when policy says the path is higher-risk because of:

- mutation impact
- external data exposure
- spend/cost
- trust level of the target integration
- privileged capability use
- durable artifact side effects

### 4. Result-surfacing path

Regardless of which path is chosen, Soma must return:

- readable content
- or a readable artifact summary plus explicit reference to what was created

Never allow the user experience to end at:

- “permission requested”
- “tool invoked”
- “execution started”

without a clear visible result.

---

## Product Manager Role

The Product Manager is the coordination authority for this lane.

The Product Manager must:

- keep runtime, UI, docs, and testing teams aligned on one contract
- reject blanket-governance decisions that over-restrict useful collaboration
- reject UX shortcuts that hide generated results or artifacts from the user
- require proof that any new content-generation flow preserves the product story:
  - Soma delivers value
  - governance builds trust
  - advanced capability stays available

The Product Manager must keep the team synchronized on:

- content mode selection
- approval-policy behavior
- artifact/media result visibility
- docs and UI wording
- browser proof and release readiness

---

## Team Structure

## 1. Product Delivery Team

### Purpose

Define the canonical operator-facing behavior for content requests.

### Required deliverables

- mode-selection rules for `answer` vs `proposal`
- accepted UX examples for:
  - inline draft
  - artifact proposal
  - artifact execution result
  - specialist-collaboration result
- explicit failure definitions for “content requested but not delivered”

## 2. Runtime and Governance Team

### Purpose

Align execution behavior with policy-configurable governance.

### Required deliverables

- capability-risk and policy mapping for:
  - inline content
  - durable file creation
  - media generation
  - MCP-backed generation
  - specialist/model collaboration
- proof that low-risk collaboration can be auto-allowed when configured
- proof that higher-risk mutation/external paths still require approval when configured

## 3. Specialist and Media Collaboration Team

### Purpose

Ensure Soma and council units can use other model/agent types for targeted output generation.

### Required deliverables

- supported collaboration patterns
- image/media/content generation examples
- result-return contract so generated outputs still come back through Soma
- clear separation between:
  - planning collaboration
  - draft generation
  - durable artifact generation

## 4. UI and Operator Experience Team

### Purpose

Make generated content and generated artifacts intelligible in the workspace.

### Required deliverables

- better user-facing explanation when Soma chooses proposal instead of inline answer
- explicit artifact references after execution
- readable preview/summary behavior for generated content
- operator wording that explains approval without bureaucracy

## 5. Documentation and Product Language Team

### Purpose

Keep docs and visible product language aligned with actual behavior.

### Required deliverables

- user-facing explanation of:
  - when Soma answers directly
  - when Soma proposes artifact creation
  - when collaboration with other agents/models can happen
  - when approval is policy-driven instead of universally required
- updated testing guidance for content/media/artifact requests

## 6. UI Testing Agentry Team

### Purpose

Verify the content-generation contract from the operator perspective.

### Required deliverables

- browser proof for inline content generation
- browser proof for governed artifact generation
- browser proof for visible result-surfacing after artifact generation
- manual trust pass for media/file/content workflows

## 7. Release and Ops Team

### Purpose

Keep environment and runtime drift from being mistaken for product failure.

### Required deliverables

- clean test environment discipline
- proof that generated-content failures are classified correctly as product vs environment vs test drift
- evidence capture support for live backend/content-generation flows

---

## Phase Plan

## Phase 1: Truth Mapping

Status: `NEXT`

Deliverables:

- identify current content requests that should stay inline but drift into proposal/tool behavior
- identify current artifact-generation flows that do not surface results clearly enough
- identify current approval-policy assumptions that are too broad for MCP/model collaboration
- identify where media/image/file generation is already conceptually supported but not surfaced well

Exit criteria:

- one mapped table of current request types -> current behavior -> target behavior

## Phase 2: Contract Definition

Status: `ACTIVE`

Deliverables:

- canonical content-mode decision rules
- canonical policy posture rules for collaboration and external/model paths
- canonical result-surfacing rules for artifacts and generated media
- canonical UX language for approval reasons and returned artifacts

Exit criteria:

- runtime, UI, docs, and testing teams are all working from the same contract
- first result-surfacing slice is identified and attached to concrete product surfaces

## Phase 3: Runtime and UX Delivery

Status: `ACTIVE`

Deliverables:

- implementation changes for mode selection
- implementation changes for result return/preview/reference
- implementation changes for policy-configurable collaboration posture
- first execution slice: Launch Crew execution outcomes must surface returned artifact references in the modal itself when Soma returns durable outputs, instead of forcing the operator to infer success from run activation alone
- updated browser coverage

Exit criteria:

- content requests reliably deliver visible value
- artifact requests reliably show artifact proof
- collaboration paths remain governed without becoming blanket-blocked

## Phase 4: Verification and Release Proof

Status: `NEXT`

Deliverables:

- focused stable browser proof
- focused live backend proof
- manual trust pass
- updated release/state/doc artifacts

Exit criteria:

- the team can prove that Mycelis delivers content/media/artifacts in a way that is both powerful and understandable

---

## Acceptance Criteria

This lane is successful only if all of the following become true:

1. Content requests no longer fail by ending at approval without visible value.
2. Inline requests are answered inline when that is the better operator outcome.
3. Artifact/media/file generation is still available and governed when appropriate.
4. Low-risk specialist/model collaboration can stay fluid when policy allows it.
5. Higher-risk external or mutating paths still surface approval when policy requires it.
6. Generated outputs always return through Soma with readable proof, preview, or artifact reference.
7. The product feels more worth paying for, not less powerful.

---

## Testing Expectations

The UI testing agentry must specifically verify:

- inline draft request -> `answer`
- durable file request -> `proposal`
- approved artifact request -> `execution_result` with artifact reference
- media/image/content-generation request -> correct mode for the user’s intent
- specialist/model collaboration request -> no false blanket approval assumption
- operator can tell what happened without raw logs

The testing team must treat these as failures:

- content requested but no visible content delivered
- proposal shown where inline answer should have sufficed
- artifact generated without readable user-facing reference
- specialist/model collaboration occurs but the result does not come back through Soma
- approval posture is treated as globally required instead of policy-configurable

---

## Immediate Team Engagement Order

1. Product Delivery Team:
   - define the initial request taxonomy for inline vs artifact vs collaboration paths
2. Runtime and Governance Team:
   - map current approval posture and identify blanket-approval assumptions
3. UI and Operator Experience Team:
   - map where the conversation fails to surface generated results clearly
4. Documentation Team:
   - draft the operator-language explanation for content/artifact/collaboration posture
5. UI Testing Agentry Team:
   - prepare a focused verification pass for content/media/file generation workflows
6. Release and Ops Team:
   - ensure clean environment proof for the focused lane

---

## Project Manager Reporting Format

Every subteam must report back in this shape:

- `What we reviewed`
- `What is already strong`
- `What is drifting`
- `What must change`
- `Dependencies on other teams`
- `Recommended priority`

The Product Manager will consolidate those reports into:

- one active lane status
- one ordered implementation plan
- one proof checklist
- one release recommendation
