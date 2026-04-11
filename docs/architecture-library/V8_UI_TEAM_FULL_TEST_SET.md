# V8 UI Team Full Test Set
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Status: ACTIVE
> Last Updated: 2026-04-11
> Owner: UI Testing Agentry Team
> Purpose: Provide the full browser and operator-experience test set for the current bounded V8.1 release target, including Central Soma home, AI Organization context flows, settings persistence, governed execution, and deployed-runtime proof.
> Supporting Docs: `V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md`, `V8_UNIVERSAL_SOMA_AND_CONTEXT_MODEL_PRD.md`, `V8_HOME_DOCKER_COMPOSE_RUNTIME.md`, `V8_MVP_MEDIA_TEAM_OUTPUT_AND_TEMPLATE_REGISTRY.md`, `docs/TESTING.md`

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
3. `uv run inv lifecycle.up --frontend`
4. `uv run inv db.migrate`
5. `uv run inv lifecycle.status`
6. `uv run inv lifecycle.health`

## 3. Required Preflight Record

Every run must record:

- branch
- commit SHA
- local date/time
- environment: `compose`, `kind`, or `managed-mock`
- browser
- whether the run is `stable`, `live-backend`, or `manual-trust`
- whether the run used a visible headed browser window or a headless/browser-cache-only path

## 3A. Browser-Visible Certification Rule

For the operator-facing UX pass, do not stop at headless Playwright or API-only checks.

Run at least the critical Chromium matrix in headed mode against the managed Interface server:

1. `uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/v8-ui-testing-agentry.spec.ts`
2. `uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/v8-organization-entry.spec.ts`
3. `uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/teams.spec.ts`
4. `uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/team-creation.spec.ts`
5. `uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/settings.spec.ts`
6. `uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/mcp-connected-tools.spec.ts`
7. `uv run inv interface.e2e --headed --live-backend --server-mode=start --project=chromium --spec=e2e/specs/groups-live-backend.spec.ts`
8. `uv run inv interface.e2e --headed --live-backend --server-mode=start --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts`

These runs certify:
- Soma-first entry and re-entry
- deterministic runtime-state answers
- AI Organization creation and re-entry
- team hub, guided team creation, direct temporary-workflow launch, archive/closure, retained-output review, and group specialization/review surfaces
- live backend-stored group-output aggregation and retained review after temporary-lane closure
- settings and Connected Tools MCP visibility
- governed live-backend execution

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

### Workflow 5B: Organization workspace guided start

Surface:

- `/organizations/[id]`

Action:

- load a freshly created or reopened AI Organization
- inspect the first visible organization-start actions
- switch into the team-design lane from the guided start surface

Expected:

- the workspace leads with a clear `start here` layer before deeper inspect surfaces
- the operator can immediately choose between Soma conversation, team design, and setup review
- guided starter prompts in the Soma lane are actionable, not decorative labels
- deeper inspect surfaces remain reachable without replacing the guided start path

Terminal state:

- `n/a`

Evidence:

- screenshot of guided organization start surface
- screenshot of team-design lane opened from the start surface

### Workflow 5C: Deterministic Soma runtime summary

Surface:

- `/organizations/[id]` or `/dashboard`

Action:

- ask `what is your current state`
- ask `what teams currently exist`

Expected:

- both prompts end in a direct `answer`
- the answer is grounded in current runtime/team state rather than provider fallback language
- no proposal is shown
- no raw transport/runtime error strings are visible in the primary answer lane

Evidence:

- screenshot of each returned answer

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
- ask for one collaboration-heavy request where Soma may use specialists/models if policy allows
- ask for one media or richer generated-content request that should return either a readable inline result or a clearly referenced governed artifact
- repeat one artifact-bearing ask and one specialist-bearing ask from `/organizations/[id]` so the default organization workspace proves the same visible output cues as the standalone chat lane

Expected:

- short reviewable work stays inline
- inline or durable artifact-returning answers visibly identify themselves as artifact-bearing output rather than looking like plain chat residue
- artifact-bearing answers also summarize the returned output in plain operator language before the artifact cards
- durable artifact requests may become proposals
- low-risk collaboration is not blanket-blocked by approval theater
- consultation-shaped answers visibly identify specialist involvement instead of hiding it only in secondary detail
- consultation-shaped answers also summarize the specialist perspective in plain operator language before the detailed consultation trace
- media or richer generated-content requests end in visible value, not only `permission requested` or `tool invoked`
- the chosen posture matches intent

Terminal state:

- `answer` or `proposal`

Evidence:

- screenshot of each result
- note whether posture matched the request
- when run from the organization workspace, capture that the organization frame stays intact while `Artifact result` and `Specialist support` remain visible

### Workflow 7A: Native team instantiation for target output

Action:

- ask Soma to create or activate a bounded delivery team for a concrete target output
- preferred release proof: ask for an image-generation team or equivalent bounded creative team and follow the path through returned output
- current bounded release proof may begin in the guided organization team-design lane before deeper instantiation/activation wiring lands

Expected:

- the workflow makes it clear that Mycelis is instantiating a native managed team rather than only calling a direct-generation tool
- the target output stays explicit
- the team-owned result returns as visible value with artifact/result framing
- the operator can tell that the output belongs to managed Mycelis execution

Terminal state:

- `proposal`, `execution_result`, or `answer` depending on the bounded path and policy

Evidence:

- screenshot of the instantiation/proposal surface
- screenshot of the returned target output
- note whether team ownership and target output stayed legible

### Workflow 7B: External workflow-contract instantiation

Action:

- target an external workflow surface such as `n8n`, ComfyUI, or comparable supported service through the intended contract path
- current bounded release proof may begin in the guided organization team-design lane before runnable invocation is available

Expected:

- the workflow makes it clear that this is an external workflow contract, not a native Mycelis team
- governance and capability posture remain visible
- normalized output or artifact returns back into Mycelis clearly

Terminal state:

- `proposal`, `execution_result`, `answer`, or `blocker` depending on support level and policy

Evidence:

- screenshot of the external contract target surface
- screenshot of the normalized returned result when runnable proof exists
- note whether ownership separation stayed clear

### Workflow 7C: MVP media and team-managed output package

Action:

- ask Soma for a marketing or product-presentation package that includes at least two different output types
- include one media-shaped target, such as hero image, visual concept, storyboard, or launch graphic
- compare it with one direct short-answer prompt so the reviewer can see when a single Soma response is enough vs when a team-managed lane is valuable
- if a media engine is configured, follow the returned artifact path through preview or download
- if no media engine is configured, confirm the UI returns a precise missing-engine blocker and keeps the prompt pack or team plan visible as useful work

Expected:

- Soma does not judge ordinary user-desired creative/business output by taste or preference
- any governance shown is tied to capability, attribution, security, side effects, external exposure, cost, or audit needs
- direct Soma output remains inline when the request is small
- the team-managed path names the team or temporary group, target outputs, and lead/coordinator
- text/code outputs are readable inline or in an artifact preview
- binary media outputs show a preview when available and a clickable path/download reference when not previewable
- retained outputs remain visible after temporary group archive

Terminal state:

- `answer`, `proposal`, `execution_result`, or `blocker` depending on configured media capability

Evidence:

- screenshot of the direct short-answer result
- screenshot of the team-managed output contract or temporary group
- screenshot of generated artifact preview/download or the precise missing-engine blocker
- note whether team ownership, target outputs, and retained outputs stayed legible

Automated proof:

- `interface/e2e/specs/v8-ui-testing-agentry.spec.ts` now includes focused Chromium coverage for direct Soma output vs team-managed output packages and for media artifact preview/save/download behavior.
- `interface/e2e/specs/mcp-connected-tools.spec.ts` now covers the adjacent Connected Tools proof: persisted MCP activity, expanded server tool visibility, and curated library install into the current user-owned group.

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

- inspect the default guided setup path
- change theme
- change assistant name
- reload the app

Expected:

- the page leads with a guided setup path instead of reading like a raw admin panel
- profile, mission-profile, and people-access actions are obvious before advanced setup
- theme saves
- theme persists after reload
- assistant name saves
- assistant name persists after reload
- advanced controls stay intentionally framed as advanced follow-ons, not the default first step
- no hidden 404 or silent revert

Terminal state:

- `n/a`

Evidence:

- screenshot of guided setup path
- screenshot before reload
- screenshot after reload

### Workflow 12A: Layered organization permissions

Surface:

- `/settings`

Action:

- open `People & Access`
- inspect the base-release access layer
- verify collaboration groups remain visible
- run one enterprise-layer proof with owner access
- run one enterprise-layer proof with non-owner access

Expected:

- base release keeps People & Access centered on visible organization roles and collaboration groups
- base release does not expose enterprise user-directory management by default
- enterprise can add the user-directory layer without replacing the organization-access layer
- enterprise owner can reach user-directory controls
- enterprise operator/viewer does not receive user-directory mutation controls

Terminal state:

- `n/a`

Evidence:

- screenshot of base-release People & Access layer
- screenshot of enterprise owner user-directory layer
- screenshot of enterprise non-owner read-only or withheld directory layer

### Workflow 12B: AI Engine and Memory & Continuity inspectability

Surface:

- `/organizations/[id]`
- `/settings`

Action:

- inspect AI Engine Settings from the organization workspace
- inspect Memory & Continuity visibility from the organization workspace or advanced reveal path

Expected:

- the operator can understand current AI Engine posture without dropping into raw configuration jargon
- Memory & Continuity remains visible and understandable as a product surface
- the inspectability path feels intentional rather than half-wired

Terminal state:

- `n/a`

Evidence:

- screenshot of AI Engine inspectability surface
- screenshot of Memory & Continuity inspectability surface

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

- the old compose fresh-organization `500` is no longer the active risk; the current compose-sensitive issue was a backend-workspace-root assumption in the live proof helper for file-side-effect assertions

Classification:

- this is now `test` / `docs` contract drift unless a clean compose rerun reproduces a real runtime failure

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
