# V8 UI Testing Stabilization Strike Team Plan

> Status: Active strike plan
> Last Updated: 2026-03-28
> Purpose: Coordinate the current UI-testing stabilization effort across runtime, interface, QA, and release hygiene so mocked browser proof, live governed-chat proof, and release packaging converge on the same trustworthy product contract.

Related:
- [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)
- [V8 UI Testing Agentry Product Contract](V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md)
- [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md)
- [V8 UI/API/Operator Contract](V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md)
- [V8 Development State](../../V8_DEV_STATE.md)

## 1. Current Situation

The stabilization effort has split cleanly into three truths:

1. Stable browser proof is mostly healthy.
   - `v8-ui-testing-agentry.spec.ts` is green.
   - `v8-organization-entry.spec.ts` is green except for one still-skipped recovery scenario.
   - `governance.spec.ts` is green.

2. Live governed-chat proof is now in review instead of blocked.
   - `soma-governance-live.spec.ts` is green on the current fresh-cluster pass.
   - the live lane now proves direct `answer`, governed `proposal`, cancel safety, and confirm-to-execution under the current `Approve & Execute` / `Execute` UI contract.

3. Release hygiene is not clean yet.
   - the local worktree still mixes testing/docs work, workspace-chat retry hardening, and unrelated theme/readability changes.

Current markers:
- overall stabilization effort: `ACTIVE`
- stable browser-contract lane: `IN_REVIEW`
- live governed-chat lane: `IN_REVIEW`
- release hygiene and commit slicing: `ACTIVE`

## 2. Strike Team Architecture

| Lane | Accountable owner | Primary responsibility | Status owner |
| --- | --- | --- | --- |
| `Architecture/Governance` | `prime-architect` | scope control, marker discipline, blocker arbitration, release verdict | overall stabilization markers |
| `Runtime/Core` | `prime-development` | real `/api/v1/chat` behavior, confirm-action behavior, backend envelope truthfulness | live governed-chat evidence |
| `Interface/Operator` | `agui-design-architect` | retry/failure UX, approval-button parity, operator-visible behavior consistency | interface stabilization evidence |
| `QA/Verification` | `council-sentry` | stable browser proof, live browser proof, manual trust pass, flake control | pass/fail evidence |
| `Release/Ops` | `admin-core` | branch hygiene, commit slicing, clean-tree proof, release candidate packaging | clean release structure |

## 3. Workstreams

### 3.1 Stable Browser-Contract Lane

Status: `IN_REVIEW`

Owner pair:
- `council-sentry`
- `agui-design-architect`

What is already covered:
- Soma-first AI Organization workspace entry
- direct `answer` path
- continuity after reload
- oversized markdown/table containment
- first-query transient recovery
- governed mutation to `proposal`
- explicit cancel flow
- inspect-only audit visibility
- `/dashboard` entry dominance and recent-organization recovery
- governance route reachability smoke

Remaining gaps:
- unskip and stabilize the guided Soma retry/recovery browser scenario in `v8-organization-entry.spec.ts`
- add stable automated `proposal -> confirm -> execution_result` proof for the org workspace lane
- add automated mid-stream interruption coverage for refresh/back/forward in the org workspace
- add automated default-path isolation proof for multiple sequential Soma tasks without team mode
- deepen audit assertions beyond reachability
- deepen causal-panel assertions so support surfaces stay explainable after Soma actions

Primary files:
- `interface/e2e/specs/v8-ui-testing-agentry.spec.ts`
- `interface/e2e/specs/v8-organization-entry.spec.ts`
- `interface/e2e/specs/governance.spec.ts`
- `tests/ui/browser_qa_plan_workspace_chat.md`
- `docs/architecture-library/V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md`
- `docs/TESTING.md`

### 3.2 Live Governed-Chat Repair Lane

Status: `IN_REVIEW`

Accountable owner:
- `prime-development`

Required paired lanes:
- `council-sentry` to refresh the live proof path
- `agui-design-architect` to verify approval-button contract parity

Latest resolved issues:
1. informational live `/api/v1/chat` requests no longer drift into proposal mode for the direct-answer proof path
2. the live approval-execution selector now matches the current button label contract
3. confirm/write proof no longer depends on the browser spec sharing the same checkout path as the live backend
4. focused live-lane validation now passes on the fresh-cluster path

Primary write scope:
- `core/internal/server/cognitive.go`
- `core/internal/server/templates.go`
- `core/internal/swarm/agent.go`
- `interface/e2e/specs/soma-governance-live.spec.ts`
- verify against `interface/components/dashboard/ProposedActionBlock.tsx`

Current evidence:
- fresh direct answer returns structured JSON `answer`
- mixed answer -> mutation returns `proposal`
- approval execution is clickable under the current UI contract
- live browser proof passes end-to-end on the fresh-cluster validation path

Residual lane goal:
- keep the live browser lane green while packaging the mixed worktree into clean release commits

### 3.3 Interface Failure-Model Parity Lane

Status: `ACTIVE`

Owner:
- `agui-design-architect`

Purpose:
- keep the operator-visible failure/retry model stable while Runtime/Core repairs the live backend path
- keep the store, retry behavior, and approval copy aligned with what the browser proof expects

Primary files:
- `interface/store/useCortexStore.ts`
- `interface/__tests__/store/useCortexStore.test.ts`
- `interface/components/dashboard/ProposedActionBlock.tsx`

Rule:
- no frontend masking of broken runtime behavior may be used to claim a live-browser pass

### 3.4 Release Hygiene and Commit Structure Lane

Status: `ACTIVE`

Owner:
- `admin-core`

Dirty-tree policy:
- do not treat the current tree as one slice
- split the current work into three clean lanes:
  1. `ui-testing-stabilization`
  2. `workspace-chat-retry-hardening`
  3. `theme-readability-pass`

Branch/commit rule:
- release proof cannot be claimed from a dirty tree that mixes testing stabilization, retry hardening, and theme polish
- the theme/readability lane must not ride along with runtime/governance repairs unless explicitly promoted later

## 4. Coordination Contract

Minimum cadence:
1. kickoff sync
   - confirm which lane is `ACTIVE`
   - confirm which lane is `BLOCKED`
   - confirm exact proof commands for the day
2. checkpoint sync
   - every lane publishes concise status and blocker updates
3. gate sync
   - QA publishes the pass/fail evidence pack
   - Architecture/Governance updates the final marker transition

Escalation rule:
- any blocker that survives one full checkpoint cycle moves back to `Architecture/Governance` for scope arbitration and must be reflected in `V8_DEV_STATE.md`

## 5. Validation Gate

### 5.1 `ACTIVE`

- dirty tree allowed only inside the owning lane branch
- failing live-backend proof is acceptable only while the lane is explicitly marked `BLOCKED`

### 5.2 `IN_REVIEW`

Requires committed state for the lane under review.

Stable browser proof:
1. `uv run inv interface.typecheck`
2. `uv run pytest tests/test_docs_links.py -q`
3. `uv run inv interface.e2e --project=chromium --spec=e2e/specs/v8-ui-testing-agentry.spec.ts`
4. `uv run inv interface.e2e --project=chromium --spec=e2e/specs/v8-organization-entry.spec.ts`
5. `uv run inv interface.e2e --project=chromium --spec=e2e/specs/governance.spec.ts`

Live governed-chat proof:
1. `cd core && go test ./internal/server -run "TestHandleChat_|TestHandleConfirmAction_" -count=1`
2. `uv run inv interface.e2e --live-backend --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts`

### 5.3 `COMPLETE`

Requires:
- clean worktree on the release-candidate branch
- stable browser proof green from committed state
- live governed-chat proof green from committed state
- `V8_DEV_STATE.md` updated with commands and outcomes
- unrelated theme/readability work either committed separately or absent from the candidate branch

## 6. Immediate Sequence

1. Commit the UI-testing stabilization lane on its own.
2. Repair the stale live approval selector in `soma-governance-live.spec.ts`.
3. Reproduce and fix the live `/api/v1/chat` non-JSON 500 path in Runtime/Core.
4. Add focused confirm-action server tests.
5. Re-run stable browser proof, then live governed-chat proof.
6. Advance the stable lane to `IN_REVIEW` and keep the live lane `BLOCKED` until the real backend path is green.
