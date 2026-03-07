# UI Target And Transaction Contract V7

> Status: Authoritative
> Last Updated: 2026-03-06
> Scope: Required UI terminal states, frontend-triggered backend effects, and delivery-proof test expectations

## 1. Why This Exists

Mycelis has repeatedly drifted into UI behavior that only returns planning language instead of a delivered result, a governed proposal, or a concrete blocker.

This contract exists to stop that drift.

Every UI path must define:
- the intended user interaction
- the expected terminal UI state
- the backend transaction or message side effect
- the failure behavior
- the proof required in tests

If a UI path does not have all five, it is not delivery-ready.

## 2. Global Rules

### 2.1 Allowed terminal UI states

Every execution-facing UI interaction must end in exactly one of these states:

1. `answer`
   - direct user-visible content is returned
2. `proposal`
   - governed mutation is prepared and awaiting confirmation
3. `execution_result`
   - work was dispatched or completed and the user can inspect the result/run/artifact
4. `blocker`
   - execution could not continue and the UI provides a concrete recovery path

Planning-only narration is not a valid terminal state.

### 2.2 Backend effect requirement

Every execution-facing UI path must map to at least one explicit backend effect:
- HTTP API request
- persisted state mutation
- NATS publish/request-reply
- run/event emission
- artifact creation

If the UI claims something happened, tests must prove the backend effect happened or the system intentionally stopped before it.

### 2.3 Failure contract

Every UI path must define:
- what the user sees when the backend rejects the request
- whether the user can retry
- whether the user can reroute
- whether the user can continue in degraded mode
- what telemetry/logging trail should exist

Raw `500`, silent timeout, or generic "failed" states are not acceptable delivery behavior.

## 3. UI Path Matrix

## 3.1 Workspace: Soma Chat

### User interaction

- operator submits freeform intent in Workspace chat

### Frontend request

- `POST /api/v1/chat`

### Backend transaction expectations

- request is routed to admin agent through NATS request-reply
- backend returns CTS-enveloped chat response
- if tools or consultations occur, they are reflected in payload metadata
- if execution creates a proposal, proposal metadata is present
- if execution creates a run or artifact, the returned payload references it

### Valid terminal UI states

- `answer`: direct content shown in chat
- `proposal`: proposal block shown with confirm/cancel controls
- `execution_result`: user sees run link, artifact, or explicit dispatched-result state
- `blocker`: structured error/recovery card

### Invalid UI outcomes

- planning-only text with no tool/proposal/result state
- silent spinner timeout
- generic "Council agent unreachable" without recovery actions

### Required UI effects

- message enters chat history
- activity indicator reflects pending execution
- returned state is rendered using one of the four valid terminal states
- consultations are rendered only when actually present

### Required test proof

- component test: render response-state variants correctly
- integration test: successful chat request yields one terminal state, not planning-only fallback
- product-flow test: plain drafting request yields direct answer in chat
- failure test: timeout or backend rejection yields structured blocker card

## 3.2 Workspace: Direct Council Chat

### User interaction

- operator targets Architect, Coder, Creative, or Sentry directly and sends a message

### Frontend request

- `POST /api/v1/council/{member}/chat`

### Backend transaction expectations

- request reaches the selected council member via request-reply subject
- response is wrapped in CTS/API envelope
- tools used, provenance, and consultations are returned when present

### Valid terminal UI states

- `answer`
- `proposal`
- `execution_result`
- `blocker`

### Required failure behavior

- retry action
- reroute to Soma action
- "continue with Soma only" action when appropriate
- copy diagnostics action

### Required test proof

- component test: council failure card renders actions and diagnostics
- integration test: council timeout returns structured blocker payload path through UI
- product-flow test: direct council success renders specialist answer without collapsing into generic chat failure

## 3.3 Launch Crew / Guided Manifestation

### User interaction

- operator uses guided workflow to turn intent into a team/workflow proposal

### Frontend requests

- intent analysis/proposal endpoints
- proposal confirmation endpoint when mutation is governed

### Backend transaction expectations

- intent is transformed into proposal or validated manifest path
- governed mutations produce intent proof and confirmation token
- confirmed activations produce run/event lineage

### Valid terminal UI states

- `proposal`
- `execution_result`
- `blocker`

### Required UI effects

- operator can inspect what will be manifested
- operator can see why confirmation is required
- successful activation yields run link or execution summary

### Required test proof

- route/component test: wizard advances only with valid required inputs
- integration test: proposal response renders approval state and confirm control
- product-flow test: confirmation produces run-linked success state
- failure test: invalid manifest/proposal shows actionable blocker

## 3.4 Workflow Composer

### User interaction

- operator composes a workflow DAG, validates it, and executes or proposes it

### Frontend requests

- workflow draft persistence endpoints
- validation/compile endpoint
- proposal/activation endpoint

### Backend transaction expectations

- graph is validated as acyclic and policy-complete
- invalid workflows do not activate
- valid workflows produce manifest/proposal/run outputs
- run/event lineage remains queryable

### Valid terminal UI states

- `proposal`
- `execution_result`
- `blocker`

### Required UI effects

- node-level validation errors
- clear indication of direct vs manifested execution path
- activation or proposal result linked to monitor surface

### Required test proof

- component test: invalid nodes/edges are highlighted correctly
- integration test: invalid graph blocks execution and surfaces validation messages
- product-flow test: valid graph reaches proposal or activation state with returned identifiers

## 3.5 Runs / Timelines / Monitoring

### User interaction

- operator inspects current or historical execution state

### Frontend requests

- `GET /api/v1/runs`
- `GET /api/v1/runs/{id}/events`
- `GET /api/v1/runs/{id}/chain`
- `GET /api/v1/runs/{id}/conversation`

### Backend transaction expectations

- run/event/conversation data is consistent with execution state
- chain queries reflect actual parent-child lineage

### Valid terminal UI states

- `execution_result`
- `blocker`

### Required UI effects

- current status visible
- failures show cause where available
- operator can navigate from summary -> run -> event/chain/conversation

### Required test proof

- component test: run cards/status badges render consistent state
- integration test: event timeline renders returned event data and error states correctly
- product-flow test: successful workflow activation can be followed from launch surface to run/timeline

## 3.6 System / Recovery / Degraded Mode

### User interaction

- operator checks status or responds to degraded infrastructure

### Frontend requests

- health/status endpoints
- service status endpoints
- any follow-up action endpoints exposed from recovery surfaces

### Backend transaction expectations

- health data reflects actual dependency state
- degraded mode still identifies which capabilities remain available

### Valid terminal UI states

- `execution_result`
- `blocker`

### Required UI effects

- explicit degraded dependencies
- exact recovery actions
- distinction between local issue, backend issue, NATS issue, and model/provider issue

### Required test proof

- component test: degraded banner and quick checks render action text, not only status colors
- integration test: backend-offline/auth-required/partial-health responses map to distinct UI states

## 4. Transaction Proof Matrix

Each execution-facing UI test slice must prove both UI state and backend effect.

| UI Path | Minimum frontend assertion | Minimum backend assertion |
| --- | --- | --- |
| Soma chat | one valid terminal state rendered | expected API route called and response classified |
| Direct council | specialist response or structured blocker shown | targeted council route called |
| Launch Crew | proposal or activation result visible | proposal/confirm endpoint called and identifiers returned |
| Workflow composer | validation/proposal/activation state visible | validation or activation endpoint invoked with expected payload |
| Run monitor | run/timeline content visible | run/event/chain queries issued and handled |
| Degraded mode | recovery guidance visible | health/status endpoint response mapped to distinct state |

Where the backend path involves NATS, integration or backend tests must additionally prove:
- subject family is canonical
- request-reply or publish behavior occurs on the expected subject
- returned data is propagated into the UI model

## 5. Required Test Layers For UI Delivery

## 5.1 Component tests

Prove:
- rendering of each terminal state
- action affordances exist
- invalid "planning-only" fallback does not masquerade as success

## 5.2 Integration tests

Prove:
- the UI issues the right API call
- the returned backend payload produces the intended terminal state
- degraded responses map to blocker/recovery states

## 5.3 Product-flow tests

Prove end-to-end user stories such as:
- asking Soma for a short letter yields an answer, not a planning loop
- direct council chat failure yields reroute/retry controls
- launch flow yields proposal then run-linked activation
- workflow validation stops invalid graphs before execution

## 5.4 Backend transaction tests

Prove:
- frontend-triggered routes produce the expected DB/NATS/runtime side effect
- no hidden backend mutation occurs when the UI should stay read-only
- proposal/activation/logging/lineage effects exist when the UI claims they do

## 5.5 Failure and rollback tests

Prove:
- timeout
- invalid payload
- unavailable agent
- degraded dependency
- rejected governance action
- retry/reroute path

## 6. Delivery Checklist (Mandatory For UI Work)

Before merging any UI-affecting slice:

1. identify the UI path(s) touched
2. identify the valid terminal state(s) for that path
3. identify the backend transaction(s) caused by the UI
4. add or update tests for:
   - terminal UI state
   - backend effect
   - failure state
5. update docs if path behavior changed

If a PR cannot name the terminal state and backend transaction it delivers, it is not ready.
