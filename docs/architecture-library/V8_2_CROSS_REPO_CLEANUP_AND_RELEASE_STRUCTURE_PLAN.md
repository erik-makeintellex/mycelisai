# V8.2 Cross-Repo Cleanup and Release Structure Plan

> Status: `ACTIVE`
> Last Updated: 2026-03-27
> Purpose: Provide one prioritized cross-team plan for cleaning the mixed local tree, repairing live blockers, packaging coherent commit lanes, and restoring a trustworthy release structure.

Related:
- [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)
- [V8 Development State](../../V8_DEV_STATE.md)
- [V8 UI Testing Stabilization Strike Team Plan](V8_UI_TESTING_STABILIZATION_STRIKE_TEAM_PLAN.md)
- [Testing](../TESTING.md)
- [Operations](../architecture/OPERATIONS.md)
- [README](../../README.md)

## 1. Current Review

### 1.1 Branch and Release Truth

Current branch truth:
- local `main` and `origin/main` are aligned at `2a3c5f8`
- the accepted V8.2 RC-delivery spine is already in pushed history
- the release problem is no longer “missing commits on remote”; it is mixed local follow-through on top of a clean pushed `main`

Pushed RC spine:
- `247fe16` `V8.2: restructure Soma-first UX and introduce light theme`
- `bfb3b80` `V8.2: enforce cognitive engine availability and Soma execution reliability`
- `1547d1a` `V8.2: clean leftover browser QA planning artifact`
- `212d5d2` `V8.2: add governance, approval workflows, and audit foundation`
- `94b531c` `V8.2: close proposal integrity gaps for governed execution`
- `2a3c5f8` `V8.2: finalize provider defaults and archive recovery docs`

Archival branch posture:
- `archive/self-hosted-pivot-stash` remains isolated and should stay out of the current release flow unless intentionally retired or mined later

### 1.2 Mixed-Tree Truth

The current local tree is mixed across at least five logical lanes:

1. Ops/workflow/task contract cleanup
   - `.github/workflows/*.yaml`
   - `ops/**`
   - `pyproject.toml`
   - `tests/test_core_tasks.py`
   - `tests/test_ci_tasks.py`
   - `tests/test_interface_tasks.py`
   - `tests/test_k8s_tasks.py`
   - `tests/test_lifecycle_tasks.py`
   - `tests/test_misc_tasks.py`
   - `tests/test_workflow_contracts.py`

2. Task/operator doc synchronization
   - `README.md`
   - `docs/LOCAL_DEV_WORKFLOW.md`
   - `docs/TESTING.md`
   - `docs/architecture/OPERATIONS.md`
   - `ops/README.md`
   - selected architecture-library follow-through docs

3. Stable UI testing agentry and Soma UX hardening
   - `interface/store/useCortexStore.ts`
   - `interface/__tests__/store/useCortexStore.test.ts`
   - `interface/e2e/specs/v8-ui-testing-agentry.spec.ts`
   - canonical testing docs and in-app docs indexing

4. Live governed-chat/browser repair
   - `interface/e2e/specs/soma-governance-live.spec.ts`
   - any paired backend/runtime fixes and focused backend tests
   - `core/internal/server/cognitive_test.go`
   - `core/internal/server/templates_test.go`

5. Theme/readability polish
   - `interface/app/globals.css`
   - `interface/app/(marketing)/page.tsx`
   - `interface/components/dashboard/CentralSomaHome.tsx`
   - `interface/components/shell/ZoneA_Rail.tsx`
   - `interface/components/workspace/DeliverablesTray.tsx`
   - `interface/components/workspace/LaunchCrewModal.tsx`

### 1.3 What Is Solid

Solid now:
- the pushed release spine is coherent
- governance/approval/audit foundations are present
- the stable mocked browser/testing contract is directionally strong
- the build/task/workflow cleanup lane has already been locally validated

Locally green evidence already observed on the mixed tree:
- Python task-module compile checks
- focused pytest task/doc/workflow suites
- `uv run inv interface.typecheck`
- `uv run inv interface.test`
- `uv run inv interface.build`
- `uv run inv test.all`
- `uv run inv ci.build`

### 1.4 What Still Needs Updating

Highest-impact updates still needed:

1. Git and release structure
   - split the mixed local tree into coherent commit lanes
   - stop treating dirty local `main` as the release-candidate proof state

2. Live governed-chat runtime proof
   - repair the real `/api/v1/chat` non-JSON failure
   - restore live approval/execute browser proof parity
   - rerun live proof from committed state

3. Canonical docs and state synchronization
   - keep state/docs/testing/ops surfaces aligned to actual contract changes
   - update the in-app Getting Started docs surface so `V8_DEV_STATE.md` is visible alongside or ahead of `V7_DEV_STATE.md`
   - add the canonical UI-testing contract docs to the architecture-library index
   - update release/testing plan docs that still reference only `workspace-live-backend.spec.ts` when the governed live lane is the real active blocker

4. Low-confidence cleanup review
   - validate usage before deleting helpers or paths that only look unused

## 2. Priority Order

### `P0` Release-Shaping Issues

1. Live governed-chat path is still the real release blocker
2. Dirty local `main` mixes unrelated slices and weakens reviewability
3. Final release proof is not yet tied to one clean committed candidate state

### `P1` Git and Lane Structure

1. separate the mixed local tree into clean lanes
2. establish explicit commit boundaries
3. use clean branch/worktree proof instead of dirty-tree confidence

### `P2` Contract and Docs Hygiene

1. synchronize task/workflow/docs surfaces
2. package canonical docs intentionally
3. keep `V8_DEV_STATE.md` current with blockers and evidence

### `P3` Deferred Cleanup

1. low-confidence dead-code removal
2. warning-noise cleanup that does not affect product truth
3. optional theme/readability refinement not needed to close live runtime trust

## 3. Strike Team Structure

| Lane | Accountable owner | Scope | Primary output |
| --- | --- | --- | --- |
| `Release/Git` | release captain | branch layout, worktrees, commit boundaries, merge order | clean candidate structure |
| `Ops/Workflow` | platform owner | invoke tasks, CI triggers, operator task docs, validation contract | clean task/workflow lane |
| `UI/Operator` | frontend owner | stable browser proof, Soma UX hardening, theme/readability lane | coherent UI lane boundaries |
| `Runtime/Governance` | backend owner | real `/api/v1/chat`, confirm-action truth, live governed browser proof | repaired live path |
| `Docs/State` | docs owner | README/state/testing/operations/index/manifest sync | synchronized canonical docs |
| `QA/Gate` | QA lead | committed-state validation only | acceptance evidence pack |

## 4. Required Lane Split

### Lane A: Ops/Workflow Contract Hardening

Include:
- `.github/workflows/*.yaml`
- `ops/**`
- `pyproject.toml`
- task/workflow tests
- task/operator docs updated because of those changes

Do not include:
- theme polish
- live governed runtime repairs
- unrelated UI state/readability work

### Lane B: Stable UI Testing Agentry + Soma UX Hardening

Include:
- `interface/store/useCortexStore.ts`
- `interface/__tests__/store/useCortexStore.test.ts`
- `interface/e2e/specs/v8-ui-testing-agentry.spec.ts`
- any paired testing docs and canonical contract docs

Do not include:
- live backend repair work
- optional theme-only polish unless it is part of operator trust or readability proof

### Lane C: Live Governed-Chat Repair

Include:
- `interface/e2e/specs/soma-governance-live.spec.ts`
- backend/runtime changes required to make live `/api/v1/chat` and confirm-action trustworthy
- focused backend tests tied directly to that repair

Do not include:
- broad workflow cleanup
- unrelated docs churn
- optional theme polish

### Lane D: Theme and Readability Polish

Include:
- visual/contrast/readability files only

Rule:
- this lane merges only after it is clear whether it is RC-critical or just polish

### Lane E: State and Planning Sync

Include:
- `V8_DEV_STATE.md`
- architecture-library planning docs
- README/testing/operations wording changes that reflect landed accepted truth rather than speculative local work

Rule:
- this lane should be finalized last, once landed code truth is known

## 5. Detailed Execution Sequence

### Phase 1: Freeze and Protect

1. stop further direct development on dirty local `main`
2. create a safety branch from the current mixed local tree
3. create one clean worktree or branch per lane from `origin/main`
4. move only lane-owned files into each branch

### Phase 2: Land the Lowest-Risk Structural Lane First

Lane A first:
- package ops/workflow/task cleanup
- rerun lane validation from committed state
- merge only if clean and independently reviewable

Why first:
- it strengthens the task/test/release contract used by all later lanes

### Phase 3: Package Stable UI Lane

Lane B second if it remains independent:
- package stable agentry/browser proof and Soma UX hardening
- keep it distinct from live governed repair
- confirm docs and manifests for canonical UI-testing docs

### Phase 4: Repair the Real Blocker

Lane C third:
- reproduce the live `/api/v1/chat` failure
- fix real backend/runtime behavior
- align governed approval/execute browser proof
- rerun live proof from committed state

### Phase 5: Decide on Theme Lane

Lane D fourth:
- either merge if judged RC-critical for readability/trust
- or hold as a separate polish branch

### Phase 6: Final State and Release Packaging

Lane E last:
- update state and planning docs to the actual landed truth
- fix visible doc/index drifts:
  - `interface/lib/docsManifest.ts` Getting Started state-doc posture
  - `docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md` UI-testing doc visibility
  - release/testing plan docs that still reference only the older workspace-live backend proof
- produce clean release-candidate evidence from committed state

## 6. Validation Requirements

### Lane A: Ops/Workflow Contract

Required:
- `python -m py_compile ops/core.py ops/interface.py ops/ci.py ops/misc.py ops/k8s.py ops/lifecycle.py ops/test.py`
- focused pytest task/workflow/doc suites
- `uv run inv -l`
- `uv run inv ci.entrypoint-check`
- `uv run inv interface.typecheck`
- `uv run inv interface.test`
- `uv run inv interface.build`
- `uv run inv test.all`
- `uv run inv ci.build`

### Lane B: Stable UI Testing Agentry

Required:
- `uv run inv interface.typecheck`
- `uv run inv interface.test`
- `uv run inv interface.e2e --project=chromium --spec=e2e/specs/v8-ui-testing-agentry.spec.ts`
- any additional focused stable browser proof for affected workspace/organization flows

### Lane C: Live Governed-Chat Repair

Required:
- focused backend tests for touched handlers
- `uv run inv interface.typecheck`
- `uv run inv interface.test`
- `uv run inv interface.e2e --live-backend --server-mode=start --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts`
- `uv run inv ci.service-check --live-backend` once the lane is believed clean

### Lane D: Theme/Readability

Required:
- `uv run inv interface.typecheck`
- `uv run inv interface.test`
- `uv run inv interface.build`
- targeted browser/manual review for the touched surfaces

### Lane E: State/Planning Sync

Required:
- `$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q`

## 7. Git Hygiene Rules

Hard rules:
- do not claim release readiness from a dirty tree
- do not mix live blocker repair with optional visual polish
- do not delete anything until current usage is validated
- do not let `V8_DEV_STATE.md` describe speculative future local work as accepted truth

Recommended merge order:
1. Lane A
2. Lane B
3. Lane C
4. Lane D if promoted
5. Lane E final sync

## 8. Exit Criteria

This cross-repo cleanup effort is only `COMPLETE` when:

1. the active candidate branch/worktree is clean
2. the mixed local tree has been split into intentional lanes
3. the live governed-chat browser proof is green against the real backend
4. task/workflow/docs contracts match actual repo behavior
5. canonical docs are intentionally packaged and indexed
6. `V8_DEV_STATE.md` reflects the accepted post-cleanup truth and evidence
