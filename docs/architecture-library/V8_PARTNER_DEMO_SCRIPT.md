# V8 Partner Demo Script

> Status: ACTIVE
> Last Updated: 2026-03-29
> Owner: Demo Scenario Team
> Supporting Teams: Product Narrative Team, Default Experience Team, Memory, Continuity, and Trust Team, Release and Ops Team
> Source Plan: `V8_DEMO_PRODUCT_STRIKE_TEAM_PLAN.md`
> Source Brief: `V8_DEMO_PRODUCT_EXECUTION_BRIEF.md`

---

## Purpose

This is the canonical partner/funder demo story for Mycelis.

It must prove that Mycelis is:

- a real product
- worth paying for
- structurally deeper than a generic assistant
- governed and controllable
- capable of real ongoing use beyond the demo

This script is not allowed to flatten the platform into a toy. It must use the clean default product story to reveal the deeper value of the platform.

---

## Demo Promise

In under ten minutes, the audience should understand:

1. Mycelis is an AI Organization product, not a single-thread chatbot.
2. Soma is the primary working counterpart.
3. Mycelis can plan directly, then switch into governed action when risk or mutation appears.
4. The product keeps visible structure, activity, and continuity.
5. The platform has deeper retained power without making the default story overwhelming.

---

## Required Demo Principles

1. Start with product value, not architecture explanation.
2. Keep structure visible.
3. Use one primary story, not multiple branches.
4. Show governance as trust and control, not friction.
5. Show continuity as retained value, not vague memory hype.
6. Reveal advanced depth only after the default product story is already clear.

---

## Canonical Demo Scenario

## Demo organization

- Name: `Northstar Labs`
- Purpose: `Operate a governed AI product-delivery organization for planning, review, and steady execution.`
- Preferred start mode: starter template

## Demo artifact target

- Artifact name: `northstar_kickoff.md`
- Artifact purpose: prove governed file creation after planning

---

## Step-by-Step Script

## Step 1: Landing

Route:
- `/`

Operator goal:
- establish category and value quickly

Narration:
- `Mycelis is an AI Organization product. It starts from structure, keeps Soma at the center, and makes governed execution and continuity feel native.`

Show:
- `Create AI Organization`
- structure, control, and continuous operation framing
- visible language about governed execution and continuity

Do not:
- start with diagnostics
- start with docs
- start with advanced routes

Success condition:
- audience understands what the product is before any technical explanation

## Step 2: AI Organization creation

Route:
- `/dashboard`

Operator goal:
- show that the first action is product-shaped, not prompt-box-shaped

Narration:
- `The product starts by creating an AI Organization, not by dropping you into a blank assistant session.`

Action:
- create `Northstar Labs`

Show:
- starter choice
- bounded setup
- visible structure-oriented framing

Success condition:
- the audience sees a product entry workflow, not a dev tool setup flow

## Step 3: Organization home

Route:
- `/organizations/{id}`

Operator goal:
- prove that the main workspace is already structured and alive

Narration:
- `This is the real operating workspace. Soma is primary, but the wider organization stays visible.`

Show:
- Soma
- advisors / departments / specialists context
- recent activity
- continuity / retained knowledge surface

Success condition:
- the workspace reads as an operating environment, not a chat pane with extras

## Step 4: Direct planning value

Operator action:
- ask Soma: `Review this AI Organization and recommend the first operating priority.`

Expected terminal state:
- `answer`

Narration:
- `For planning and review, Soma should answer directly instead of forcing approval theater.`

Show:
- direct response in the main Soma flow
- no unnecessary proposal state

Success condition:
- audience sees immediate value and product intelligence

## Step 5: Governed mutation

Operator action:
- ask Soma: `Create a kickoff brief in the workspace called northstar_kickoff.md summarizing the first operating priority and next steps.`

Expected terminal state:
- `proposal`

Narration:
- `When the request becomes mutating, Mycelis does not hide execution. It shifts into governed action.`

Show:
- proposal card
- approval posture
- capability/risk explanation

Success condition:
- audience sees trust and control without losing momentum

## Step 6: Confirm and execute

Operator action:
- confirm the proposal

Expected terminal state:
- `execution_result`

Narration:
- `The same workflow now shows the action result and proof, instead of hiding side effects in the background.`

Show:
- successful completion
- artifact or file proof
- updated workspace context

Success condition:
- audience sees visible actuation without ambiguity

## Step 7: Continuity and retained value

Operator action:
- refresh or re-enter the same organization

Expected terminal state:
- continuity remains legible

Narration:
- `Mycelis keeps the organization oriented over time. It can retain useful patterns and working continuity without turning every conversation into permanent memory.`

Show:
- recent activity
- retained patterns / memory & continuity panel
- same organization context still intact after refresh or return

Success condition:
- audience understands this is a product with continuity, not a stateless prompt loop

## Step 8: Optional advanced reveal

Only use this step if the audience wants proof of depth.

Options:
- open `Memory`
- open `Resources`
- open `System`

Narration:
- `The platform keeps deeper operator power available, but it does not force that complexity into the default product path.`

Success condition:
- advanced depth increases confidence instead of confusing the main story

---

## Recommended Prompt Set

Primary direct-value prompt:
- `Review this AI Organization and recommend the first operating priority.`

Primary governed-action prompt:
- `Create a kickoff brief in the workspace called northstar_kickoff.md summarizing the first operating priority and next steps.`

Continuity follow-up prompt:
- `What did you retain from the kickoff decision, and what is still temporary working context?`

---

## Demo Fallback Path

## Fallback A: direct-answer lane is noisy

If direct answer is degraded:
- use a recent visible organization summary panel
- retry once
- keep the narration focused on structure and governed continuity

Narration:
- `The workspace is still intact; I’m reissuing the planning request through the same organization context.`

## Fallback B: mutation result is noisy

If the governed mutation lane is noisy:
- still show proposal generation
- cancel instead of confirm
- narrate that the contract is still valuable because action did not bypass approval

Narration:
- `Even when I stop here, the product has already proven the right behavior: risky work becomes explicit, reviewable, and cancellable.`

## Fallback C: live environment degrades

If the live environment degrades:
- show Recent Activity and continuity
- optionally use one advanced reveal to demonstrate retained depth
- do not pivot into infrastructure debugging unless the audience asks

---

## Anti-Patterns

- do not lead with architecture docs
- do not start in Advanced mode
- do not show five modes of routing
- do not spend time explaining backend internals before value is visible
- do not make approval look like bureaucracy
- do not imply the product is only this one demo

---

## What This Demo Must Prove About Platform Value

This script only passes if it makes clear that Mycelis is:

- useful immediately
- governed when acting
- structurally deeper than a one-thread assistant
- able to retain useful context over time
- backed by a broader platform that advanced users can grow into
