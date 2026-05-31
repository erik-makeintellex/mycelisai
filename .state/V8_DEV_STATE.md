# Mycelis V8 - Active Development State
> Navigation: [Project README](../README.md) | [Docs Home](../docs/README.md)

> Updated: 2026-05-31
> Canonical active scoreboard for V8 delivery. Historical V7/V8 migration evidence remains in `.state/V7_DEV_STATE.md`, git history, and canonical architecture docs.

## Current Checkpoint

| Field | State |
| --- | --- |
| Current committed checkpoint | `21595000 Add scheduler handoff persistence` |
| Working tree at checkpoint | Clean after commit and proof |
| Release proof | `uv run inv ci.release-preflight --lane=baseline --no-e2e` passed |
| Focused browser proof | Schedule Rules Chromium E2E passed before the final baseline gate |
| Product posture | Governed schedule proposals are durable and reviewable; scheduler ticks do not execute work autonomously |
| Main release risk | Local Go is `go1.25.6`; docs/toolchain lock expects `go1.26` |
| Environment blocker | Live media proof is blocked until a local media provider and `cognitive.media-gateway` are online |

## Delivery Status

| Lane | Status | Where We Are | Next Target | Proof Gate |
| --- | --- | --- | --- | --- |
| Scheduler/Cadence | `COMPLETE` | Schedule due ticks persist idempotent handoff records with `handoff_key`, intent proof, execution contract, proposal status, and bounded payload. No scheduler-created runs, no confirm tokens in history, no NATS/team command publication, and no confirm-action calls from scheduler ticks. | Surface approval-state transitions from the durable handoff path. | Go scheduler/history/triggers tests, API history proof, autonomous-execution guard review. |
| UI Workflow/Expression | `COMPLETE` | Approvals now shows schedule-origin audit context and Schedule Rules narrow-layout resilience is green. UI work is judged on workflow fit and visual target expression, not only feature function. | Apply the same full workflow plus visual-expression review to the next touched UI/API slice. | Focused Vitest, TypeScript, max-lines, and headed/browser proof where the workflow is visible. |
| API/Runtime Contract | `COMPLETE` | Trigger execution history includes schedule handoff proof fields while preserving governed review boundaries. | Add approval transition fields/endpoints only when execution handoff is ready to move past proposal evidence. | Go API tests, docs/API reference review, history reload proof. |
| Deployment/Proof | `COMPLETE` | Baseline source release-preflight is green on the committed tree. | Align local Go to `1.26`, then promote through full E2E, Compose, WSL, Rancher/K8s, and hosted/manual proof as needed. | `ci.release-preflight`, focused live GUI, `compose.health`, `wsl.validate --lane=release`, K8s proof. |
| Docs/State | `ACTIVE` | API reference, automation user docs, and this scoreboard reflect the scheduler handoff slice. | Keep docs/state synchronized in the same slice as behavior, API, runtime, workflow, testing, or terminology changes. | docs/workflow tests, `git diff --check`, max-lines. |
| Media Lane | `BLOCKED` | Media provider/gateway checks are offline: no gateway on `127.0.0.1:8001`, no Forge/AUTOMATIC1111 on `:7860`, no ComfyUI on `:8188`. | Start a local provider, run `uv run inv cognitive.media-gateway`, verify `/health`, then run live retained-media output proof. | `tests/test_media_gateway.py` and headed live media retained-output proof. |

## Proof Evidence

Latest green proof for checkpoint `21595000`:

1. `go test ./internal/triggers ./internal/server -count=1`
2. `go test ./...`
3. `npm test -- ApprovalsTab ScheduleRulesTab --maxWorkers=1`
4. `npm test -- ApprovalsTab --maxWorkers=1`
5. `npx tsc --noEmit`
6. `uv run pytest tests/test_db_tasks.py tests/test_docs_links.py tests/test_workflow_contracts.py tests/test_runtime_deploy_contract_text.py -q`
7. `uv run inv quality.max-lines --limit 300`
8. `uv run inv interface.e2e --server-mode=start --project=chromium --workers=1 --spec=e2e/specs/schedule-rules.spec.ts`
9. `uv run inv ci.release-preflight --lane=baseline --no-e2e`

Known non-blocking warnings:

- Baseline reports Go version drift: expected `go1.26`, found `go1.25.6`.
- Interface unit tests still emit existing React `act(...)` and localstorage-file warnings, but the suite passes.

## Next Run Targets

| Priority | Target | Action | Owner Lane | Exit Criteria |
| --- | --- | --- | --- | --- |
| P0 | Toolchain alignment | Install/select Go `1.26` for release-proof lanes and rerun baseline preflight. | Deployment/Proof | Toolchain check no longer warns and baseline remains green. |
| P0 | Full workflow proof standard | For the next product slice, prove both UI workflow and API/runtime contract across create/propose/approve/execute/output/proof/recovery/reload where applicable. | QA/Embodiment | Test plan names every user-visible and API-visible step touched by the slice. |
| P1 | Approval-state handoff | Design and implement proposal approval-state transitions from persisted scheduler handoff records without autonomous scheduler execution. | Scheduler/Cadence + API | Durable transition state reloads through API/history and UI; scheduler guard tests remain explicit. |
| P1 | Media unblock | Start Forge/AUTOMATIC1111 or ComfyUI and the media gateway; validate retained output from browser UI. | Media Lane | Live media retained-output proof passes with open/viewable artifact evidence. |
| P1 | Visual expression review | Run focused review on Dashboard/Soma, Automations, Approvals, Groups, Resources, and System after each touched slice. | Visual/UI | No overlap, no confusing density, no stale wording, and target Soma-governed expression is preserved. |
| P2 | Promotion proof | After local source proof is green and services are intentionally up, run Compose, WSL, Rancher/K8s, and hosted/manual workflows as corroboration. | Deployment/Proof | Promotion proof records exact commit, environment, commands, and pass/fail result. |
| P2 | Documentation hygiene | Keep canonical docs compact and linked; avoid turning state back into a historical transcript. | Docs/State | State file stays as active scoreboard; deep evidence lives in commit history and owning docs. |

## Team Engagement Plan

| Team | Immediate Assignment | Coordination Rule |
| --- | --- | --- |
| Scheduler/Cadence | Own next approval-state transition slice. | Do not weaken the no-autonomous-execution boundary. |
| Visual/UI | Review every touched workflow for function, clarity, layout, and target expression. | Browser proof required for visible workflow changes. |
| Deployment/Proof | Own Go `1.26` alignment and ordered promotion proof. | Source proof first; deployment proof corroborates after local correctness. |
| Media Lane | Unblock local provider/gateway and prove retained media artifacts. | No release claim for media until live provider proof is green. |
| Docs/State | Keep state/docs synchronized with behavior changes. | Use canonical status markers only: `REQUIRED`, `NEXT`, `ACTIVE`, `IN_REVIEW`, `COMPLETE`, `BLOCKED`. |

## Documentation Map

| Topic | Canonical Location |
| --- | --- |
| Product and architecture entrypoint | `architecture/mycelis-architecture-v7.md` |
| V8.3 operational embodiment | `docs/architecture-library/V8_3_OPERATIONAL_EMBODIMENT_PRD.md` |
| Runtime contracts | `docs/architecture-library/V8_RUNTIME_CONTRACTS.md` |
| API behavior | `docs/API_REFERENCE.md` |
| Testing and task running | `docs/TESTING.md`, `docs/architecture/OPERATIONS.md`, `ops/README.md` |
| User-facing Automations behavior | `docs/user/automations.md` |
| Active delivery state | `.state/V8_DEV_STATE.md` |

## Historical Contract Anchors

These compact anchors preserve repo-level documentation invariants without keeping the full historical transcript in this active scoreboard.

### 17. Template and instantiation entry points definition

- V8.3 release-candidate embodiment planning is now canonical over the older delivery transcript, while the Soma-primary compatibility baseline remains distinct from the full V8.2/B2+ production target.
- the `Migration from V7 bootstrap assumptions` section now explains how fixed V7 startup assumptions collapse into explicit V8 configuration sources.
- Task 004  Config and bootstrap model planning                       [COMPLETE]
- Task 008  Planning-integration validation pass                      [COMPLETE]
- Task 009  Next-execution/governance guidance migration              [NEXT]
- `COMPLETE` run the planning-integration validation pass so README, the architecture-library index, docs manifests, and doc-tests all confirm the new V7-to-V8 bootstrap migration contract.
- startup now instantiates runtime organization truth directly from self-contained bundle data, retired the remaining no-bundle bootstrap fallback, and `MYCELIS_BOOTSTRAP_TEMPLATE_ID` must be set whenever more than one bundle is present.

## Parallel Execution Overlay (2026-04-27)

| Historical Lane | Anchor |
| --- | --- |
| Backend/Auth Architecture | Authentication and runtime source-of-truth architecture must remain explicit and verified. |
| Frontend/Auth UX | Default UX stays Soma-first while auth and advanced controls remain understandable and bounded. |
| Workflow Testing | Workflow proof must cover the user-visible path and the corresponding API/runtime contract. |
| Runtime/MCP/Web Capability | Capability surfaces must be governed, inspectable, and separated from raw runtime internals. |

## Deferred And Advanced Contracts

- managed exchange security foundation now exists: managed exchange remains permissioned, and normalization into managed exchange does not imply unrestricted trust.
- enterprise identity, approval workflows, and multi-user access management remain deferred beyond the free-node security foundation.
- `NEXT` advanced architecture/runtime configuration remains a separate contract and implementation lane; advanced UI may explain deployment/env influence, but it must not become a second source of runtime truth.

## Architecture Synchronization Rule

Every slice must:

- update state
- verify README alignment
- verify V8.2 alignment

a slice is not complete unless tests pass, documentation is updated where meaning changed, and architecture alignment is verified.

- `COMPLETE` records accepted delivered work
- `ACTIVE` records work in progress
- `NEXT` records the next committed follow-on slices

slice close-out should explicitly report tests run, docs changed, and docs reviewed unchanged for the touched scope.

## State Rules

1. Keep this file compact and current; do not append long historical transcripts.
2. Record only accepted checkpoint evidence, blockers, and next-run targets.
3. Update owning docs in the same slice when behavior, API, runtime, workflow, testing, or canonical terminology changes.
4. Use `.state/V7_DEV_STATE.md` and git history for migration evidence.
5. Before release claims, record commit, environment, command, and result.
