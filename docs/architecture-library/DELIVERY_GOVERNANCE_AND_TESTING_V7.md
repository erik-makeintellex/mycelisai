# Delivery Governance And Testing V7
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Status: Canonical
> Last Updated: 2026-04-17
> Scope: Delivery slices, evidence requirements, documentation discipline, and product-aligned testing.

Related:
- [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)
- [Target Deliverable V7](TARGET_DELIVERABLE_V7.md)
- [System Architecture V7](SYSTEM_ARCHITECTURE_V7.md)
- [Execution And Manifest Library V7](EXECUTION_AND_MANIFEST_LIBRARY_V7.md)
- [Intent To Manifestation And Team Interaction V7](INTENT_TO_MANIFESTATION_AND_TEAM_INTERACTION_V7.md)
- [UI And Operator Experience V7](UI_AND_OPERATOR_EXPERIENCE_V7.md)
- [V8 Trusted Memory Arbitration And Team Vector Contract](V8_TRUSTED_MEMORY_ARBITRATION_AND_TEAM_VECTOR_CONTRACT.md)

Supporting specialized docs:
- [Testing](../TESTING.md)
- [Operations](../architecture/OPERATIONS.md)
- [Next Target Gated Delivery Program](../architecture/NEXT_TARGET_GATED_DELIVERY_PROGRAM.md)

## V8 Migration Alignment

- Delivery governance now references `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md` as the bootstrap and V7->V8 migration contract, even when slices still operate on legacy V7 artifacts.
- `Template ≠ instantiated organization`; all templates/YAML/runtime/DB/operator inputs must be translated through the V8 template -> instantiation -> inheritance -> precedence pipeline before they change runtime behavior.
- `.state/V8_DEV_STATE.md` is the authoritative queue for slice/NEXT markers; `.state/V7_DEV_STATE.md` remains a historical reference only.
- Governance review should confirm that every slice uses V7 docs strictly as migration inputs and reports evidence back into the V8 state file.

## 1. Delivery Principle

A slice is not accepted because code changed.

A slice is accepted when:
- the intended product behavior is explicit
- tests prove that behavior
- docs describe the same behavior
- failure and rollback behavior are also proven

## 2. Required Slice Structure

Each delivery slice should define:
- objective
- targeted files
- operator-visible outcome
- backend/runtime effect
- tests
- acceptance criteria
- rollback plan
- evidence block

## 3. Documentation Discipline

When behavior changes, update the canonical docs in the same slice.

Minimum candidates:
- `README.md`
- `.state/V8_DEV_STATE.md` (state scoreboard; reference `.state/V7_DEV_STATE.md` only for historical migration evidence)
- `docs/TESTING.md`
- `docs/architecture/OPERATIONS.md`
- `docs/README.md`
- `interface/lib/docsManifest.ts` when the documentation should be visible in the in-app docs page
- this architecture library when the change alters target, architecture, execution, UI, or delivery rules
- `docs/API_REFERENCE.md` when API behavior or payload meaning changes

Documentation review rule:
- every implementation slice must include a docs review for the touched surface, even when the result is "reviewed, no content change required"

## 3.1 Feature Status Markers

Use the canonical status markers below when tracking delivery in planning and state docs:
- `REQUIRED`
- `NEXT`
- `ACTIVE`
- `IN_REVIEW`
- `COMPLETE`
- `BLOCKED`

These markers should be used consistently across:
- `README.md`
- `.state/V7_DEV_STATE.md`
- target-delivery/action-card docs
- phase/gate documentation

Do not replace them with ad hoc synonyms when a canonical marker fits.

## 4. Testing Must Prove Delivery

### 4.1 Contract tests

Prove:
- routes
- subject families
- schemas
- docs/task contracts

### 4.2 Behavioral unit tests

Prove:
- local logic for the changed outcome

### 4.3 Integration tests

Prove:
- the boundary between layers still behaves correctly

### 4.4 Product-flow tests

Prove:
- the user story reaches a real end state

### 4.5 Failure and rollback tests

Prove:
- blocked/degraded/retry/rollback behavior is safe and understandable

### 4.6 Trusted memory tests

When a slice changes memory posture, promotion rules, or recall precedence, prove:
- candidate-first promotion behavior
- precedence/arbitration behavior when memories conflict
- operator-visible trust or governance outcomes for the changed path

## 5. UI-Specific Rule

A UI test is incomplete if it proves only rendering.

For execution-facing UI, tests must prove:
- initiating interaction
- terminal UI state
- backend effect
- failure behavior
- recovery affordance

## 5.1 Backend/API -> UI Target Plan Rule

If a slice changes backend/API behavior, it must include a UI-targeted review/test plan in the same delivery window.

Minimum required plan fields:
- changed backend/API surface:
  - routes, payload shape, response/error envelope, or runtime channel behavior
- expected UI surfaces:
  - exact pages/components/store paths affected by the backend/API change
- terminal UI states:
  - which of `answer` / `proposal` / `execution_result` / `blocker` are expected after the change
- backend effect evidence:
  - how the frontend-triggered backend effect will be proven (route call, DB/NATS/run/event outcome)
- failure-path evidence:
  - expected degraded/rejection/timeout behavior and how it will be tested
- command evidence plan:
  - exact commands that will be run before review (`uv run inv ...` plus focused suites)

Review gate:
- if backend/API changed and no UI target plan is attached, the slice is not review-ready.

## 6. Runtime-Specific Rule

Runtime changes are incomplete if they do not prove:
- startup behavior
- shutdown behavior
- health/recovery behavior
- degraded behavior where relevant

## 7. Recurring-Plan Testing Rule

For `scheduled`, `persistent_active`, or `event_driven` plans, acceptance requires proof of persistence behavior, not only first activation.

Minimum proof:
- survives restart or rehydrates correctly
- pause/resume or activation control behaves correctly
- UI reflects recurring state correctly

## 8. Evidence Blocks

Each accepted slice should record:
- commands run
- whether they passed or failed
- date
- relevant focused suites
- any skipped verification and why

## 9. Release And Phase Gates

Phase advancement requires:
- acceptance criteria met
- tests passing at the required level
- docs updated
- no unresolved critical regressions in the scoped area

Release promotion additionally requires:
- clean preflight gates
- current operator docs
- no stale task contracts

## 10. Immediate Implication

The system should bias toward small, high-proof slices:
- one operator journey or runtime contract at a time
- docs and tests updated in the same change
- avoid broad speculative rewrites without acceptance proof

## 11. Target-Action Test Lockstep

Target actions and test actions must move together.

For the current intent-to-manifestation target, acceptance requires the following alignment:

| Target action | Minimum required test actions |
| --- | --- |
| Soma-first Team Expression + module binding | component/store tests for expression rendering/editing, integration tests for adapter payload normalization (`internal`/`mcp`/`openapi`/`host`), product-flow test proving `proposal` -> confirmation -> run-linked outcome |
| Created-team workspace + channel inspector | UI tests for scoped communications filters, integration tests for created-team command -> `signal.status`/`signal.result` round-trip, product-flow test for interjection/pause-resume/reroute control effects |
| Scheduler recurring execution (`scheduled_missions`) | backend tests for schedule CRUD and tick behavior, persistence tests across restart/rehydration, UI tests proving recurring status and control states |
| Causal chain operator surface | UI tests for `/runs/[id]/chain` rendering and navigation, integration tests verifying chain query mapping and error states |

A slice that changes one of these target actions is not `COMPLETE` until the matching test actions are attached as evidence.

## 12. Canonical Command Gate

Testing evidence for target actions should be captured with:
- `uv run inv core.test`
- `uv run inv interface.test`
- `uv run inv interface.build`
- `uv run inv ci.baseline`

When a slice adds E2E-critical operator behavior, also include focused Playwright proof with `uv run inv interface.e2e ...`.

Next:
- return to [Target Deliverable V7](TARGET_DELIVERABLE_V7.md)
