# V8 UI Team Full Test Set

> Status: ACTIVE
> Last Updated: 2026-03-30
> Owner: UI Testing Agentry Team
> Purpose: Provide the full browser and operator-experience test set for the current V8.2 MVP, including Central Soma home, AI Organization context flows, settings persistence, governed execution, and deployed-runtime proof.
> Supporting Docs: `V8_UI_TESTING_AGENTRY_EXECUTION_RUNBOOK.md`, `V8_UI_WORKFLOW_VERIFICATION_PLAN.md`, `V8_PARTNER_DEMO_VERIFICATION_CHECKLIST.md`, `V8_UNIVERSAL_SOMA_AND_CONTEXT_MODEL_PRD.md`, `V8_HOME_DOCKER_COMPOSE_RUNTIME.md`

## 1. Test Goal

The UI team is not only checking that pages render.

The UI team must prove that:

- Mycelis reads as a product, not a dev console
- Soma is one persistent counterpart across contexts
- AI Organizations read as governed work contexts
- content is visibly delivered inline or as clearly referenced artifacts
- governed execution remains trustworthy
- settings and continuity actually persist
- advanced depth remains reachable without polluting the default path

## 2. Supported Test Environments

### 2.1 Compose runtime

Use for:

- home-runtime proof
- partner-demo proof
- single-host product validation

Bring-up order:

1. `uv run inv compose.down --volumes`
2. `uv run inv compose.up --build`
3. `uv run inv compose.status`
4. `uv run inv compose.health`
5. `uv run inv interface.check`

### 2.2 Kind/Kubernetes runtime

Use when:

- cluster behavior matters
- bridge/recovery behavior is under review
- a chart-oriented runtime issue is suspected

Bring-up order:

1. `uv run inv lifecycle.down`
2. `uv run inv k8s.reset`
3. `uv run inv k8s.bridge`
4. `uv run inv db.migrate`
5. `uv run inv lifecycle.up --frontend`
6. `uv run inv lifecycle.status`
7. `uv run inv lifecycle.health`

## 3. Required Preflight Record

Every run must record:

- branch
- commit SHA
- local date/time
- environment: `compose`, `kind`, or `managed-mock`
- browser
- whether the run is `stable`, `live-backend`, or `manual-trust`

## 4. Execution Order

Run in this exact order:

1. environment preflight
2. Central Soma and navigation proof
3. AI Organization entry and creation proof
4. organization workspace bootstrap proof
5. direct-answer and content-delivery proof
6. governed mutation and artifact proof
7. continuity and refresh proof
8. settings persistence proof
9. advanced-surface boundary proof
10. audit/activity visibility proof
11. disruption and recovery proof
12. deployed live-backend proof

## 5. Workflow Test Set

### Workflow 1: Central Soma home

Surface:

- `/dashboard`

Action:

- load the dashboard in a clean session

Expected:

- `Central Soma` is visible
- page explains that Soma works across AI Organizations
- create/re-enter actions are visible
- default path still leads cleanly into AI Organization creation

Terminal state:

- `n/a`

Evidence:

- full-page dashboard screenshot

### Workflow 2: Primary navigation

Surface:

- Zone A rail

Action:

- inspect the default rail with Advanced mode off

Expected:

- primary nav includes `Soma`, `Docs`, and `Settings`
- advanced routes remain hidden
- if a prior organization exists, `Current Organization` or `Return to Organization` appears

Terminal state:

- `n/a`

Evidence:

- screenshot of rail

### Workflow 3: AI Organization creation from template

Surface:

- `/dashboard`

Action:

- choose template mode
- enter name and purpose
- create organization

Expected:

- fields render correctly
- starter cards use user-facing language
- successful creation routes to `/organizations/[id]`

Terminal state:

- `n/a`

Evidence:

- screenshot before submit
- screenshot after landing

### Workflow 4: Empty-start AI Organization creation

Action:

- choose `Start empty`
- enter name and purpose
- create organization

Expected:

- name and purpose inputs are visible
- create button becomes enabled when inputs are valid
- success lands in organization context

Terminal state:

- `n/a`

Evidence:

- screenshot of valid form state
- screenshot after landing

### Workflow 5: Reopen current organization

Action:

- leave the current organization
- return using dashboard or rail

Expected:

- resume path is obvious
- organization context is preserved
- Soma does not appear to become a different identity

Terminal state:

- `n/a`

Evidence:

- screenshot of return path

### Workflow 6: Direct Soma answer

Surface:

- `/organizations/[id]`

Action:

- ask: `Draft a three sentence product positioning statement for Mycelis.`

Expected:

- direct inline answer
- no proposal card
- answer is readable and useful

Terminal state:

- `answer`

Evidence:

- screenshot of answer

### Workflow 7: Inline content vs artifact posture

Action:

- ask for a short reviewable draft
- ask for a large durable output such as a JSON file or wide data export

Expected:

- short reviewable work stays inline
- durable artifact requests may become proposals
- the chosen posture matches intent

Terminal state:

- `answer` or `proposal`

Evidence:

- screenshot of each result
- note whether posture matched the request

### Workflow 8: Governed mutation proposal

Action:

- ask to create a file in the workspace

Expected:

- proposal appears
- default-visible surface answers:
  - what Soma wants to do
  - why approval is needed, or why the action is optional/auto-allowed
  - expected result
  - what will change
- approval posture and risk/cost context are visible
- low-level tool, module-binding, and expression detail is not dumped into the default view
- advanced execution detail is reachable behind `Show details`
- no side effect before confirm

Terminal state:

- `proposal`

Evidence:

- proposal screenshot

### Workflow 9: Cancel behavior

Action:

- cancel the proposal

Expected:

- UI clearly says no action executed
- no lingering ambiguity remains

Terminal state:

- `execution_result`

Evidence:

- cancelled-state screenshot

### Workflow 10: Confirm-and-execute behavior

Action:

- approve a safe file-creation proposal

Expected:

- execution completes
- artifact path or result is clearly referenced
- if Launch Crew or another bounded outcome surface returns artifacts, the success state itself names the returned output instead of only pointing to a run
- reload preserves execution status

Terminal state:

- `execution_result`

Evidence:

- screenshot of success state
- artifact reference note

### Workflow 11: Activity and audit visibility

Surface:

- approvals/activity/audit surfaces

Action:

- inspect recent activity after a governed action

Expected:

- recent action is visible
- status is readable
- review surface stays inspect-only by default

Terminal state:

- `n/a`

Evidence:

- screenshot of activity surface

### Workflow 12: Settings persistence

Surface:

- `/settings`

Action:

- change theme
- reload the app

Expected:

- theme saves
- theme persists after reload
- no hidden 404 or silent revert

Terminal state:

- `n/a`

Evidence:

- screenshot before reload
- screenshot after reload

### Workflow 13: Advanced mode boundaries

Action:

- toggle Advanced mode on and off

Expected:

- `Resources`, `Memory`, and `System` appear only when Advanced mode is on
- default path remains clean when Advanced mode is off
- advanced power is hidden, not removed

Terminal state:

- `n/a`

Evidence:

- one screenshot off
- one screenshot on

### Workflow 14: Continuity after refresh

Action:

- refresh during or after recent interaction

Expected:

- organization context remains clear
- prior result state remains readable
- operator does not need to reconstruct what happened

Terminal state:

- same as prior visible state

Evidence:

- screenshot after refresh

### Workflow 15: Disruption and degraded recovery

Action:

- trigger a transient failure path or interrupt a run mid-stream when the lane allows it

Expected:

- error is bounded and readable
- recovery path is visible
- the app does not collapse into raw backend noise

Terminal state:

- `blocker` then recovery to `answer`, `proposal`, or `execution_result`

Evidence:

- screenshot of degraded state
- screenshot of recovery state

## 6. Deployed Live-Backend Proof

### 6.1 Stable compose-hosted browser proof

Run against the deployed compose stack:

- `PLAYWRIGHT_SKIP_WEBSERVER=1`
- `PLAYWRIGHT_PORT=3000`

Required stable spec set:

- `e2e/specs/v8-organization-entry.spec.ts`
- `e2e/specs/v8-ui-testing-agentry.spec.ts`

### 6.2 Live governed browser proof

Run against the deployed compose stack:

- `PLAYWRIGHT_LIVE_BACKEND=1`
- `PLAYWRIGHT_SKIP_WEBSERVER=1`
- `PLAYWRIGHT_PORT=3000`
- `PLAYWRIGHT_BACKEND_WORKSPACE_ROOT=d:\MakeIntellex\Projects\mycelisai\scratch\workspace\docker-compose\data\workspace`

Required live spec:

- `e2e/specs/soma-governance-live.spec.ts`

Current known risk:

- fresh-organization direct-answer Scenario A may still return `500 Internal Server Error` in compose even while proposal, cancel, confirm, and persistence scenarios pass

Classification:

- this is `product` or `runtime`, not `environment`, if compose health is green

## 7. Reporting Format

Every finding must include:

- Workflow
- Surface
- User action
- Expected
- Observed
- Terminal state
- Outcome
- Classification
- Impact class
- Why it matters
- Evidence

Allowed outcomes:

- `PASS`
- `SOFT_FAIL`
- `HARD_FAIL`

Allowed classifications:

- `product`
- `environment`
- `test`
- `docs`

Allowed impact classes:

- `clarity`
- `hierarchy`
- `continuity`
- `causality`
- `navigation`
- `trust`

## 8. Final Verdict Format

End every run with:

- Overall verdict: `READY`, `READY_WITH_NOTES`, or `BLOCKED`
- Critical failures
- Soft failures
- Trust wins
- Environment notes
- Recommended next actions in strict priority order

## 9. Non-Negotiable Rule

Do not recommend changes that make Mycelis simpler by reducing the platform into a demo shell.

The correct target is:

- one persistent Soma
- scoped AI Organizations
- governed execution
- visible value delivery
- persistent settings and continuity
- advanced depth that remains reachable and coherent
