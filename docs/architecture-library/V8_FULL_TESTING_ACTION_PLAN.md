# V8 Full Testing Action Plan

> Status: ACTIVE
> Last Updated: 2026-04-03
> Owner: Product Management / Delivery Coordination
> Purpose: Define one canonical full-testing execution order that ties together repo gates, stable user-workflow proof, live governed-browser proof, compose-aware runtime proof, and final evidence recording.

## Why This Plan Exists

Mycelis already has detailed testing docs, but the practical release question still comes up repeatedly:

- what is the real full gate
- which order should the team run it in
- which docs are for users vs agents vs release reviewers
- how do we prove the product workflows instead of only command success

This plan answers that with one ordered runbook.

## Audience Split

### User guidance

User-facing docs should tell operators how to use Soma, organizations, approvals, artifacts, settings, and recovery workflows.

Primary user docs:

- `docs/README.md`
- `docs/user/core-concepts.md`
- `docs/user/soma-chat.md`
- `docs/user/governance-trust.md`
- `docs/user/automations.md`
- `docs/user/resources.md`
- `docs/user/memory.md`
- `docs/user/system-status-recovery.md`
- `docs/user/run-timeline.md`

### Agent and developer guidance

Contributor-facing docs should explain current authority, implementation truth, task contracts, and release validation.

Primary agent/developer docs:

- `AGENTS.md`
- `README.md`
- `V8_DEV_STATE.md`
- `docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md`
- `docs/TESTING.md`
- `docs/architecture/OPERATIONS.md`

### Release and QA guidance

Release reviewers should use the workflow and testing plans directly:

- `docs/TESTING.md`
- `docs/architecture-library/V8_UI_WORKFLOW_VERIFICATION_PLAN.md`
- `docs/architecture-library/V8_UI_TEAM_FULL_TEST_SET.md`
- `docs/architecture-library/V8_UI_TESTING_AGENTRY_EXECUTION_RUNBOOK.md`
- `docs/architecture-library/V8_USER_WORKFLOW_EXECUTION_AND_VERIFICATION_PLAN.md`
- `docs/architecture-library/V8_USER_INTENT_WORKFLOW_REVIEW_AND_INSTANTIATION_PLAN.md`

## Full Testing Goals

The full gate is complete only when it proves:

1. repo integrity
2. stable product workflows
3. live governed runtime behavior
4. compose/home-runtime viability when that runtime is part of the release story
5. documentation and state alignment after the proof

## Canonical Test Waves

### Wave 0: Preflight and cleanliness

Required:

- confirm the worktree is in the intended state
- confirm the docs and state file are updated for the slice
- confirm no stale expectations remain in README, testing docs, or in-app docs manifest

Recommended commands:

```powershell
git status --short
uv run inv -l
```

### Wave 1: Repo baseline gate

Purpose:

- prove the branch passes the standard local repo gate from committed code

Required command:

```powershell
uv run inv ci.baseline
```

This is the canonical baseline and should be treated as the default release-readiness gate for:

- docs checks
- quality gates
- backend tests
- interface tests
- stable Playwright coverage

### Wave 2: Live governed-browser gate

Purpose:

- prove the real `/api/v1/chat` and governed execution path against a live backend

Required command:

```powershell
uv run inv ci.service-check --live-backend
```

Use this when:

- proxy/runtime contracts changed
- chat/governance/approval behavior changed
- the slice affects real backend/browser interaction

### Wave 3: Compose-aware runtime proof

Purpose:

- prove the supported home-runtime stack remains usable when compose is part of the release story

Required commands:

```powershell
uv run inv compose.up
uv run inv compose.status
uv run inv compose.health
```

When compose-backed browser proof is needed, use the backend workspace root explicitly:

```powershell
$env:MYCELIS_BACKEND_WORKSPACE_ROOT='workspace/docker-compose/data/workspace'
uv run inv interface.e2e --live-backend --server-mode=start --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts
```

### Wave 4: Workflow-specific proof

Purpose:

- prove the changed user workflows themselves, not only the general gate

Required rule:

- rerun the focused unit/component/page/browser coverage for every touched workflow lane

Examples:

- organization workspace and team-design changes:
  - `cd interface; npx vitest run __tests__/organizations/TeamLeadInteractionPanel.test.tsx __tests__/pages/OrganizationPage.test.tsx --reporter=dot`
  - `uv run inv interface.e2e --project=chromium --spec=e2e/specs/v8-organization-entry.spec.ts`
- settings workflow changes:
  - `cd interface; npx vitest run __tests__/pages/SettingsPage.test.tsx --reporter=dot`
  - `uv run inv interface.e2e --project=chromium --spec=e2e/specs/settings.spec.ts`

### Wave 5: Documentation and state verification

Purpose:

- prove the docs and live state tell the same story as the code and tests

Required command:

```powershell
$env:PYTHONPATH='.'
uv run pytest tests/test_docs_links.py -q
```

Required follow-through:

- update `README.md` when navigation or top-level contract meaning changes
- update `docs/README.md` when the guidance map changes
- update `interface/lib/docsManifest.ts` when a new canonical doc should be visible in `/docs`
- update `V8_DEV_STATE.md` with the delivered slice and executed evidence

## Expected Evidence Output

Every full-testing run should leave behind:

- executed command list
- pass/fail result for each wave
- any scoped exceptions or accepted residual warnings
- the matching `V8_DEV_STATE.md` update
- a clean commit when the slice is accepted

## Failure Classification

Classify failures as one of:

- `product`
- `runtime`
- `environment`
- `test`
- `docs`

Do not treat a workflow as passing if:

- only the control rendered
- only the API call happened
- the badge appeared without a meaningful workflow result
- docs still point users or agents at the wrong authority

## Current Action Order

For active V8 release-candidate work, the default full-testing order is:

1. `uv run inv ci.baseline`
2. `uv run inv ci.service-check --live-backend`
3. `uv run inv compose.up`
4. `uv run inv compose.status`
5. `uv run inv compose.health`
6. compose-backed live browser proof when the home-runtime stack is in scope
7. focused workflow suites for the touched slice
8. `uv run pytest tests/test_docs_links.py -q`
9. update `V8_DEV_STATE.md`
10. commit from a clean tree

## Current Execution Checkpoint

This plan should be updated in the same slice whenever:

- the canonical gate changes
- stable vs live proof order changes
- compose becomes more or less central to release proof
- new user-workflow proof becomes mandatory for MVP acceptance

Current recorded execution checkpoint for 2026-04-03:

- `uv run inv ci.baseline` -> passed
- `uv run inv ci.service-check --live-backend` -> passed
- `uv run inv compose.up` -> passed
- `uv run inv compose.status` -> passed
- `uv run inv compose.health` -> passed
- `$env:MYCELIS_BACKEND_WORKSPACE_ROOT='workspace/docker-compose/data/workspace'; uv run inv interface.e2e --live-backend --server-mode=start --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts` -> passed
- `$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q` -> passed

Known accepted note:

- Vitest worker startup still emits repeated `--localstorage-file` path warnings during the broader baseline run; that warning remains an accepted non-blocker in the current RC state.
