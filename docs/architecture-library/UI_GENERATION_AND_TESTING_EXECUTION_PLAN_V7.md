# UI Generation And Testing Execution Plan V7

> Status: Canonical execution plan
> Last Updated: 2026-03-10
> Purpose: Define a deterministic UI-generation workflow and deep test strategy so MVP UX is shippable beyond internal test-team usage.

Related:
- [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)
- [UI And Operator Experience V7](UI_AND_OPERATOR_EXPERIENCE_V7.md)
- [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md)
- [Next Execution Slices V7](NEXT_EXECUTION_SLICES_V7.md)
- [UI Target And Transaction Contract V7](../architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md)
- [Testing](../TESTING.md)

## Status Markers

Use only:
- `REQUIRED`
- `NEXT`
- `ACTIVE`
- `IN_REVIEW`
- `COMPLETE`
- `BLOCKED`

## Why This Plan Exists

UI delivery has reached the point where component-level quality is high, but route-level product proof is uneven. This plan defines one generation model and one testing model so the team can ship consistent operator outcomes instead of ad hoc UI behavior.

## UI Generation Model (Deterministic)

Every new UI feature must be generated through these stages in order:

1. `REQUIRED` Contract intake
   - identify route, user goal, terminal states (`answer|proposal|execution_result|blocker`)
   - identify backend/API effects and failure states
2. `REQUIRED` Surface specification
   - component boundaries, store responsibilities, and data contract normalization
   - explicit progressive-disclosure strategy for diagnostics vs primary user intent
3. `ACTIVE` Implementation slice
   - route/page implementation
   - component implementation
   - store action/state wiring
4. `REQUIRED` Evidence slice
   - unit/component coverage for visible states and controls
   - integration coverage for request/response mapping
   - browser flow coverage for route-level outcome
5. `IN_REVIEW` Delivery gate
   - docs updated in same slice
   - command evidence attached
   - regression risk annotated

## UI Generation Contracts

For each generated surface, include:

1. Layout contract
   - primary intent area vs diagnostics area
   - responsive behavior and mobile constraints
2. Interaction contract
   - exact state transitions for primary actions
   - retry/recovery actions for degraded paths
3. Data contract
   - endpoint ownership
   - normalized payload shape at UI boundary
4. Failure contract
   - operator-readable message
   - deterministic recovery action
5. Accessibility contract
   - keyboard path for primary action
   - role/name labels for critical controls

## Testing Architecture (Deep Coverage)

Required layers for execution-facing UI:

1. Component tests (Vitest)
   - visible terminal states
   - interactive control behavior
2. Store/integration tests (Vitest)
   - endpoint mapping
   - success/failure state propagation
3. Route browser tests (Playwright)
   - full route outcomes
   - navigation and tab transitions
4. Live-backend browser checks (Playwright)
   - required when API proxy/runtime contract is changed

## Route Priority Matrix

| Route group | Current status | Next coverage target |
| --- | --- | --- |
| `/dashboard` Workspace | `ACTIVE` | expand failure/recovery + private-channel relay UX paths |
| `/docs` | `ACTIVE` | add internal-link traversal + manifest/read error branch browser proof |
| `/runs`, `/runs/[id]` | `ACTIVE` | add interjection + terminal transition + retry flow browser proof |
| `/automations` | `ACTIVE` | add created-team channel inspector route proof when unblocked |
| `/settings`, `/system`, `/resources` | `ACTIVE` | strengthen mutation-path and degraded-state browser assertions |

## Evidence Command Set

Minimum route-slice evidence:

1. `uv run inv interface.test`
2. `cd interface && npx vitest run --reporter=dot`
3. `cd interface && npx playwright test --project=chromium <spec>`
4. `uv run inv interface.build`
5. `uv run inv ci.baseline`

When backend/API contracts changed:

1. `uv run inv core.test`
2. `uv run inv interface.e2e --live-backend --server-mode=start --project=chromium --spec=<spec>`

## Immediate Execution Queue

1. `ACTIVE` Route-level docs/runs baseline coverage (unit + browser smoke) delivered.
2. `NEXT` Expand `/docs` and `/runs` from smoke to failure/recovery depth.
3. `NEXT` Stabilize WebKit route coverage for critical UX paths.
4. `REQUIRED` Deliver created-team communications inspector route coverage once Slice 7 unblock lands.

## MVP UI Exit Criteria

UI is `IN_REVIEW` for MVP when:

1. every primary workflow route has route-level browser proof
2. every mutating UX path has proposal/confirm/effect coverage
3. every degraded path has deterministic recovery controls and tests
4. docs and state file are updated in same slice
5. `ci.baseline` and targeted UI route suites pass

