# V8 True MVP Finish Plan

> Status: ACTIVE
> Last Updated: 2026-03-29
> Owner: Product Management / Delivery Coordination
> Purpose: Drive Mycelis from release-candidate posture to a true MVP that is product-legible, operationally reliable, and worth paying for.

## MVP Standard

Mycelis is at true MVP only when all of the following are true:

- a technical evaluator can understand the product in minutes
- Soma delivers visible value directly instead of stalling at governance mechanics
- governed mutation feels safe and understandable
- continuity and retained knowledge feel valuable and bounded
- key settings and support surfaces actually work end to end
- advanced depth remains reachable without overwhelming the default story
- clean startup, health, and live browser proof are repeatable

## Current Priority Order

### 1. Product Value Delivery

Mycelis must reliably:

- present one persistent Soma counterpart even while work stays context-scoped
- answer planning and review requests inline
- generate visible content when content is requested
- generate and reference artifacts clearly when durable output is requested
- support policy-configurable specialist/model/MCP collaboration without blanket approval overreach

### 2. MVP Reliability

The default operator path must stay solid:

- AI Organization entry
- organization workspace bootstrap
- direct Soma answer
- proposal -> cancel / confirm
- continuity after refresh/re-entry
- audit visibility
- support-panel trust behavior

### 3. Settings and Operator Controls

Operator-facing controls must be real features, not placeholders:

- theme selection
- assistant naming
- advanced-mode boundaries
- AI Engine and Memory & Continuity inspectability

### 4. Demo and Partner Readiness

The product must show:

- clear value
- governed trust
- continuity
- retained structural depth

without degrading into a narrow showcase.

### 5. Clean Release Discipline

The repo and runtime must support:

- clean environment bring-up
- clean process/port discipline
- accurate monitoring and health surfaces
- committed, validated release slices

## Active Delivery Lanes

### Lane A: Content Generation and Collaboration

Goal:

- fix the gap where Soma asks for permission or invokes tools without clearly delivering content or artifact value back to the user

Key outcomes:

- better answer-vs-proposal selection
- better artifact result-surfacing
- policy-configurable collaboration posture

### Lane B: Demo Product Hardening

Goal:

- make Mycelis read as a product, not a lab console, while preserving depth

Key outcomes:

- cleaner narrative
- better default surfaces
- preserved advanced reveal
- clearer progression toward a Central Soma home instead of a purely org-scoped front door

### Lane C: Settings and Operator Functionality

Goal:

- ensure visible operator settings actually work

Current focus:

- theme selection end-to-end
- assistant identity persistence
- no dead settings controls in the default path

### Lane D: Release and Live Proof

Goal:

- keep fresh-cluster bring-up, health, and live browser proof trustworthy

Key outcomes:

- clean startup from zero
- stable live governed execution proof
- accurate environment triage

## Immediate Next Deliverables

1. Complete truth mapping for content requests:
   - inline content
   - governed artifact creation
   - media/image generation
   - specialist/model collaboration

2. Close the content result-surfacing gap:
   - generated output must appear inline or be clearly referenced

3. Audit remaining visible settings for dead controls or partial wiring.

4. Align the UI model around universal Soma with context-scoped AI Organizations, starting from the dashboard/home contract.

5. Align the UI testing agentry around the new content/artifact/collaboration expectations.

6. Re-run full MVP validation from a clean environment after the next delivery slice.

## Release Gate For True MVP

Mycelis can be treated as true MVP only when:

- `ci.baseline` is green or any residual issue is explicitly accepted and non-user-facing
- stable browser proof is green
- live governed browser proof is green
- the partner demo lane is `READY` or `READY_WITH_NOTES`
- content/media/artifact workflows deliver visible value
- visible operator settings work end to end
- documentation and product wording match the actual platform
