# V8 Partner Demo Verification Checklist

> Status: ACTIVE
> Last Updated: 2026-03-29
> Owner: UI Testing Agentry Team
> Supporting Teams: Demo Scenario Team, Release and Ops Team, Product Management / Delivery Coordination
> Source Script: `V8_PARTNER_DEMO_SCRIPT.md`
> Source Brief: `V8_DEMO_PRODUCT_EXECUTION_BRIEF.md`

---

## Purpose

This checklist verifies that the canonical partner demo actually delivers the intended product story.

It is narrower than the full release verification plan.

The question here is not only:
- does the UI technically work?

The question is:
- does the product present value, trust, structure, and retained depth in the way a technical partner or funder needs to see it?

---

## Required Evidence Format

For every step, record:

- `Surface`
- `User action`
- `Expected`
- `Observed`
- `Outcome`
- `Why it matters`

Allowed outcomes:

- `PASS`
- `SOFT_FAIL`
- `HARD_FAIL`

Rule:
- a step can be technically functional but still be a `SOFT_FAIL` if it weakens product clarity or trust

---

## Demo Verification Sequence

## 1. Landing comprehension

Surface:
- `/`

Verify:
- Mycelis reads as an AI Organization product
- Soma is understandable as the primary counterpart
- structure, governance, and continuity are visible in plain language
- no default copy makes the product feel like a lab console

Hard fail if:
- a new technical viewer would still ask what the product actually is after the first screen

## 2. AI Organization creation

Surface:
- `/dashboard`

Verify:
- creating an organization is the obvious next step
- starter vs empty flow is clear
- the experience feels product-shaped, not setup-heavy
- hidden advanced controls feel intentionally deferred, not missing

Hard fail if:
- the page feels like advanced system configuration instead of product entry

## 3. Organization home legibility

Surface:
- `/organizations/{id}`

Verify:
- Soma is primary
- structure stays visible
- Recent Activity is visible
- retained knowledge / continuity is visible
- the page feels like an operating workspace, not generic chat

Hard fail if:
- the main workspace reads like a chat app with side widgets

## 4. Direct planning value

Surface:
- Soma workspace

User action:
- `Review this AI Organization and recommend the first operating priority.`

Verify:
- result lands in `answer`
- the answer is useful and direct
- proposal flow does not appear unnecessarily

Hard fail if:
- informational planning gets trapped in mutation or approval behavior

## 5. Governed mutation clarity

Surface:
- Soma workspace

User action:
- `Create a kickoff brief in the workspace called northstar_kickoff.md summarizing the first operating priority and next steps.`

Verify:
- result lands in `proposal`
- proposal card is readable
- approval posture is intelligible
- mutation does not silently execute

Hard fail if:
- the request mutates without proposal
- the proposal is too technical to explain quickly

## 6. Confirm and execution proof

Surface:
- Soma workspace

User action:
- confirm the proposed action

Verify:
- result lands in `execution_result` or equally bounded visible result state
- artifact/file proof is visible
- the operator can explain what happened without raw logs

Hard fail if:
- confirmation succeeds but the outcome is unclear

## 7. Continuity and retention

Surface:
- organization home after refresh or re-entry

Verify:
- organization context stays intact
- recent activity still reads clearly
- retained patterns / continuity wording is understandable
- the product distinguishes durable retained value from temporary working continuity

Hard fail if:
- the workflow feels stateless after refresh or return

## 8. Optional advanced reveal

Surface:
- `/memory`, `/resources`, or `/system`

Verify:
- advanced routes remain reachable after enabling Advanced mode
- advanced depth increases confidence
- advanced reveal does not contradict the default product story

Hard fail if:
- advanced power appears removed, broken, or disconnected from the product

---

## Product-Specific Trust Checks

## Trust Check A: governance sells confidence

Verify:
- approval feels like control, not bureaucracy
- cancel path feels safe
- proposal wording supports trust quickly

## Trust Check B: continuity sells long-term value

Verify:
- continuity wording does not sound vague or magical
- retained knowledge feels useful and bounded

## Trust Check C: structure sells platform depth

Verify:
- the audience can see that the product has teams, specialists, reviews, and ongoing work behind Soma
- the product does not collapse into one-thread assistant behavior

---

## Release/Ops Proof For Demo Lane

Before a high-stakes demo, also confirm:

- clean startup path succeeds
- health checks are green
- demo organization creation works in the intended environment
- live governed mutation lane is healthy
- one fallback path is rehearsed

---

## Acceptance Rule

The partner demo lane is ready only if the team can say all of the following are true:

1. the product is understandable in minutes
2. governed execution is visible and reassuring
3. the platform feels deeper than a simple assistant
4. continuity makes the product feel stateful and valuable
5. advanced power is preserved, not amputated for presentation
