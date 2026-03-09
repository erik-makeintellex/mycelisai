# Delivery Governance And Testing V7

> Status: Canonical
> Last Updated: 2026-03-09
> Scope: Delivery slices, evidence requirements, documentation discipline, and product-aligned testing.

Related:
- [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)
- [Target Deliverable V7](TARGET_DELIVERABLE_V7.md)
- [System Architecture V7](SYSTEM_ARCHITECTURE_V7.md)
- [Execution And Manifest Library V7](EXECUTION_AND_MANIFEST_LIBRARY_V7.md)
- [Intent To Manifestation And Team Interaction V7](INTENT_TO_MANIFESTATION_AND_TEAM_INTERACTION_V7.md)
- [UI And Operator Experience V7](UI_AND_OPERATOR_EXPERIENCE_V7.md)

Supporting specialized docs:
- [Testing](../TESTING.md)
- [Operations](../architecture/OPERATIONS.md)
- [Next Target Gated Delivery Program](../architecture/NEXT_TARGET_GATED_DELIVERY_PROGRAM.md)

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
- `V7_DEV_STATE.md`
- `docs/TESTING.md`
- `docs/architecture/OPERATIONS.md`
- `docs/README.md`
- `interface/lib/docsManifest.ts` when the documentation should be visible in the in-app docs page
- this architecture library when the change alters target, architecture, execution, UI, or delivery rules

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
- `V7_DEV_STATE.md`
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

## 5. UI-Specific Rule

A UI test is incomplete if it proves only rendering.

For execution-facing UI, tests must prove:
- initiating interaction
- terminal UI state
- backend effect
- failure behavior
- recovery affordance

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
