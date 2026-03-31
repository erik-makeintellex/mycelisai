# V8 UI Testing Agentry Execution Runbook

> Status: ACTIVE
> Last Updated: 2026-03-30
> Owner: UI Testing Agentry Team
> Supporting Docs:
> - `V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md`
> - `V8_UI_WORKFLOW_VERIFICATION_PLAN.md`
> - `V8_PARTNER_DEMO_VERIFICATION_CHECKLIST.md`

## Purpose

This runbook tells the UI testing agentry exactly how to execute Mycelis browser verification, how to record results, and how to return evidence in a way that is easy to triage and act on.

This is an operational guide, not a replacement for the product contract.

Use it when:

- a product/UI change needs browser proof
- a release candidate needs UI verification
- a partner demo lane needs validation
- a failure needs to be classified as product, environment, test drift, or documentation drift

## Core Rule

The agentry is not here only to prove that screens render.

The agentry must prove that:

- Mycelis reads as a product, not a lab console
- Soma is clearly primary
- governed action flow remains trustworthy
- continuity survives disruption
- advanced depth remains reachable without overwhelming the default path

## Operating Principles

1. Test serially, not in parallel, when using managed `interface.e2e` runs.
   The managed frontend build/start path can interfere with itself if multiple E2E jobs try to clean and rebuild the same Next.js bundle at once.

2. Always separate `environment`, `product`, `test`, and `docs` failures.
   A stale selector, stale wording expectation, or stale route assumption is not the same as a broken product.

3. Record terminal state, not just pass/fail.
   Every key workflow must end in one of:
   - `answer`
   - `proposal`
   - `execution_result`
   - `blocker`

4. Critique the operator experience, not only the implementation.
   If something technically works but weakens clarity or trust, report it.

5. Never flatten advanced depth into a default-path requirement.
   Default-path simplicity is good. Hidden or broken advanced capability is not.

## Required Test Lanes

Run in this order unless a release lead explicitly changes the sequence.

1. Repo and environment preflight
2. Stable mocked browser proof
3. Live backend governed browser proof
4. Manual trust and disruption pass
5. Partner demo verification pass when the lane is demo-facing

## Preflight Checklist

Before browser work starts, verify:

- the repo state being tested is known and recorded
- the environment is intentional
- Mycelis services are either already up and healthy, or the runbook includes fresh bring-up proof

Record:

- branch name
- commit SHA
- local date/time
- whether this is `mocked`, `live-backend`, or `demo-lane` testing

Recommended commands:

- `uv run inv lifecycle.status`
- `uv run inv lifecycle.health`
- `uv run inv db.status`

If a fresh-environment proof is required, run and record:

- `uv run inv lifecycle.down`
- `uv run inv k8s.reset`
- `uv run inv k8s.bridge`
- `uv run inv db.migrate`
- `uv run inv lifecycle.up --frontend`

## Step 1: Stable Mocked Browser Proof

Purpose:

- prove the UI contract independently from backend variability

Command:

- `uv run inv interface.e2e --project=chromium --spec=e2e/specs/v8-ui-testing-agentry.spec.ts`

The agentry must review these workflows in order, even if the automated spec is already green.

### 1A. AI Organization entry

Surface:

- `/dashboard`

Action:

- create or open an AI Organization

Expected:

- the page reads as product entry
- creating an organization is the obvious next step
- successful entry lands in `/organizations/[id]`

Evidence:

- one screenshot of the entry surface
- one screenshot after successful organization landing

### 1B. Workspace bootstrap

Surface:

- `/organizations/[id]`

Expected:

- Soma is visually and behaviorally primary
- organization framing remains visible
- Recent Activity is readable
- Memory & Continuity is readable
- support surfaces do not overwhelm the main path

Evidence:

- one full-page screenshot after workspace load

### 1C. Direct answer flow

Action:

- submit: `Summarize the current Workspace V8 design objectives.`

Expected:

- main chat request lands in `answer`
- no proposal controls appear
- the answer is directly useful

Evidence:

- screenshot showing the answer state
- note whether the answer felt direct, delayed, or overly indirect

### 1D. Governed mutation flow

Action:

- submit: `Create a simple python file named hello_world.py in the workspace.`

Expected:

- result lands in `proposal`
- proposal card default view shows what Soma wants to do, why approval is needed or optional, the expected result, and what will change
- proposal card still shows approval posture and risk/cost context
- tool, module-binding, and expression mechanics stay behind an explicit details control instead of leading the default surface
- mutation does not silently execute

Evidence:

- screenshot of the proposal card
- note the time-to-visible-state if it felt slow or ambiguous

### 1E. Proposal cancel flow

Action:

- cancel the proposal from 1D

Expected:

- the proposal is visibly neutralized
- UI explicitly says no action executed

Evidence:

- screenshot of the cancelled state

### 1F. Cold-start recovery

Action:

- trigger the first-query transient-failure path if the lane is instrumented for it

Expected:

- failure is bounded
- retry guidance is visible
- retry returns the workspace to a valid terminal state

Evidence:

- screenshot of degraded/blocker state
- screenshot of recovered terminal state

### 1G. Continuity after refresh

Action:

- refresh or re-enter the same organization

Expected:

- recent context remains legible
- organization framing remains intact
- the operator is not forced to reconstruct what just happened

Evidence:

- screenshot after refresh showing continuity

### 1H. Oversized content resilience

Action:

- request a very large markdown table or long code block

Expected:

- content stays inside a bounded reading surface
- no layout breakage or sidebar warping

Evidence:

- screenshot proving overflow containment

## Step 2: Live Backend Governed Browser Proof

Purpose:

- prove the real `/api/v1/chat` and `/api/v1/intent/confirm-action` contract

Command:

- `uv run inv interface.e2e --live-backend --server-mode=start --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts`

If the core workspace lives in a different checkout than the browser spec, set one of:

- `PLAYWRIGHT_BACKEND_WORKSPACE_ROOT`
- `MYCELIS_BACKEND_WORKSPACE_ROOT`

The agentry must confirm these workflows:

### 2A. Fresh-organization direct answer

Action:

- create a fresh AI Organization
- submit: `Summarize the current Workspace V8 design objectives.`

Expected:

- request succeeds against the live backend
- result lands in `answer`
- answer text is non-empty and readable

Record:

- request method and path
- response status
- terminal state

### 2B. Mutation after answer-mode chat

Action:

- in the same organization, submit a mutating file request

Expected:

- the second request still lands in `proposal`
- answer-mode chat does not break mutation governance

Record:

- whether the transition from answer to proposal felt causal and understandable

### 2C. Cancel remains side-effect free

Action:

- cancel a pending proposal

Expected:

- no file/artifact appears before confirm
- cancel state persists through reload

Record:

- visible cancel proof
- whether any hidden side effect was observed

### 2D. Confirm produces durable proof

Action:

- approve a pending governed action

Expected:

- confirm endpoint succeeds
- execution result is bounded and visible
- file/artifact appears only after confirm
- success survives reload

Record:

- run ID if visible in API/body
- visible execution proof
- artifact existence proof

## Step 3: Manual Trust and Disruption Pass

Purpose:

- evaluate trust, continuity, wording, and recovery feel beyond deterministic automation

Reference:

- `tests/ui/browser_qa_plan_workspace_chat.md`

Run these passes:

### 3A. Mid-stream interruption

Action:

- interrupt a chat with refresh during or near stream activity

Expected:

- no hydration crash
- operator can re-enter and continue with confidence

### 3B. Browser navigation

Action:

- use Back and Forward across dashboard and organization routes

Expected:

- organization identity and recent context remain clear

### 3C. Failure recovery

Action:

- trigger one handled failure such as a forbidden file target

Expected:

- failure is explained cleanly
- retry/recovery path is clear
- trust can recover after the failure

### 3D. Wording audit

Check:

- landing page
- organization entry
- organization workspace
- approvals/audit surfaces
- support panels

Expected:

- no dev-facing jargon in the default path
- Mycelis reads as an AI Organization product
- Memory & Continuity wording is consistent

## Step 4: Optional Partner Demo Pass

Only run when validating a partner/funder demo lane.

Reference:

- `docs/architecture-library/V8_PARTNER_DEMO_VERIFICATION_CHECKLIST.md`

Focus:

- product clarity within minutes
- governed execution as a trust asset
- continuity as durable value
- advanced depth still reachable after the default story is understood

## Evidence Package Standard

For every workflow, capture:

- `Workflow`
- `Surface`
- `User action`
- `Expected`
- `Observed`
- `Terminal state`
- `Outcome`
- `Why it matters`

Allowed outcomes:

- `PASS`
- `SOFT_FAIL`
- `HARD_FAIL`

Terminal states:

- `answer`
- `proposal`
- `execution_result`
- `blocker`
- `n/a` for non-chat navigation checks

Every negative finding must include:

- one screenshot or visible evidence point
- one classification:
  - `product`
  - `environment`
  - `test`
  - `docs`
- one impact class:
  - `clarity`
  - `hierarchy`
  - `continuity`
  - `causality`
  - `navigation`
  - `trust`

Every positive finding must include:

- one concrete reason trust was earned

## Result Logging Format

Use this exact structure for each finding:

```text
Workflow:
Surface:
User action:
Expected:
Observed:
Terminal state:
Outcome:
Classification:
Impact class:
Why it matters:
Evidence:
```

## Summary Format For Delivery Back

The agentry should end every run with:

1. `Overall verdict`
   - `READY`
   - `READY_WITH_NOTES`
   - `BLOCKED`

2. `Critical failures`
   - only `HARD_FAIL` items

3. `Soft failures`
   - clarity/trust/navigation issues that do not block the lane

4. `Trust wins`
   - concrete reasons the product earned confidence

5. `Environment notes`
   - stale process
   - port collision
   - bridge issue
   - test harness drift

6. `Recommended next actions`
   - ordered, short, directly actionable

## Triage Rules

Classify findings carefully:

- `product`
  - UI behavior or visible product wording is wrong
- `environment`
  - ports, bridges, cluster, frontend startup, or service health caused the issue
- `test`
  - stale selector, stale wording expectation, stale route assumption, or harness race
- `docs`
  - product changed but verification docs still describe the old contract

Examples:

- If the workspace input exists but the spec is waiting for old placeholder text, classify as `test`.
- If the UI copy changed and docs still describe the old copy, classify as `docs`.
- If `/api/v1/chat` returns malformed or failing output in the live lane, classify as `product` unless proven to be environment-specific.
- If the stack fails because PostgreSQL or NATS bridges are down, classify as `environment`.

## Escalation Rules

Escalate immediately if any of the following happen:

- a mutating action bypasses proposal/approval
- cancel leaves uncertainty about whether execution happened
- refresh destroys confidence about current organization context
- the default path reads like a generic assistant or dev console
- live governed execution fails on a clean environment without a clear environment cause

## Exit Standard

The agentry is done only when:

- the tested lane is identified
- every required workflow has a recorded outcome
- every failure is classified correctly
- screenshots or visible evidence are attached for all failures
- the final verdict is explicit
- recommended next actions are prioritized for engineering follow-up
