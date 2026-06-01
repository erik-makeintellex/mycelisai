# Mycelis V8 - Active Development State
> Navigation: [Project README](../README.md) | [Docs Home](../docs/README.md)

> Updated: 2026-06-01
> Canonical active scoreboard for V8 delivery. Historical V7/V8 migration evidence remains in `.state/V7_DEV_STATE.md`, git history, and canonical architecture docs.

## Current Checkpoint

| Field | State |
| --- | --- |
| Current committed checkpoint | `11ca823e Add schedule handoff approval controls` |
| Working tree at checkpoint | Clean after the committed scheduler handoff approval-state/testing slice |
| Release proof | `uv run inv ci.release-preflight --lane=baseline --no-e2e` passed on checkpoint `11ca823e` |
| Focused browser proof | Schedule Rules Chromium E2E passed for the committed handoff controls slice |
| Product posture | Governed schedule proposals are durable and reviewable; scheduler ticks do not execute work autonomously |
| Main release risk | Full live GUI/media proof is still blocked by local provider/gateway availability |
| Environment blocker | Live media proof is blocked until NATS/Core/Frontend plus a local media provider and `cognitive.media-gateway` are online |

## Delivery Status

| Lane | Status | Where We Are | Next Target | Proof Gate |
| --- | --- | --- | --- | --- |
| Scheduler/Cadence | `COMPLETE` | Schedule due ticks persist idempotent handoff records with `handoff_key`, intent proof, execution contract, proposal status, bounded payload, and explicit terminal handoff approval transitions (`approved`, `rejected`, `cancelled`). | Keep the no-autonomous-execution boundary intact while later slices connect handoff approval into live run creation. | Go scheduler/history/triggers tests, API history proof, autonomous-execution guard review. |
| UI Workflow/Expression | `COMPLETE` | Schedule Rules and Approvals tolerate and display explicit schedule handoff approval states from rule, execution, audit, or payload-shaped state fields. Schedule Rules exposes approve/reject/cancel controls only for persisted `awaiting_approval` schedule handoff executions. | Apply the same proof discipline to the next visible workflow slice: active media output, focused team context, and output folder affordances. | Focused Vitest, TypeScript, max-lines, and headed/browser proof where the workflow is visible. |
| API/Runtime Contract | `COMPLETE` | Trigger execution history includes schedule handoff proof fields, and `POST /api/v1/triggers/{id}/history/{executionId}/approval` transitions persisted schedule handoff state only. | Keep API docs synchronized for the next runtime contract change. | Go API tests, docs/API reference review, history reload proof. |
| Deployment/Proof | `ACTIVE` | Baseline source release-preflight is green on committed checkpoint `11ca823e`; local Go reports `go1.26.3`. Native status check shows PostgreSQL and Ollama up, with NATS, Core API, and Frontend down. | Start native NATS/Core/Frontend intentionally, then run focused live GUI/media proof before Compose/WSL/K8s promotion. | `ci.release-preflight`, focused live GUI, `compose.health`, `wsl.validate --lane=release`, K8s proof. |
| Docs/State | `ACTIVE` | API reference, automation user docs, and this scoreboard reflect the committed scheduler handoff slice and next-phase kickoff. | Keep docs/state synchronized in the same slice as behavior, API, runtime, workflow, testing, or terminology changes. | docs/workflow tests, `git diff --check`, max-lines. |
| Media Lane | `NEXT` | Media gateway unit coverage is green, but live provider/gateway checks remain offline until services are intentionally started. | Start Forge/AUTOMATIC1111 or ComfyUI and the media gateway, verify `/health`, then run live retained-media output proof through Soma/UI. | `tests/test_media_gateway.py` and headed live media retained-output proof. |

## Proof Evidence

Latest green proof for checkpoint `11ca823e`:

1. `go test ./...`
2. `go test ./internal/triggers -run "Test(LogExecution_WithHandoffRefs|GetExecutionByHandoffKey|TransitionScheduleHandoffApproval)"`
3. `go test ./internal/server -run "TestHandle(TriggerHistory|ScheduleHandoffApproval)"`
4. `npm test -- ScheduleRulesTab.test.tsx ScheduleRulesHandoffActions.test.tsx ScheduleHandoffStore.test.ts ApprovalsTab.test.tsx`
5. `npx tsc --noEmit`
6. `npm run build`
7. `uv run pytest tests/test_docs_links.py tests/test_workflow_contracts.py tests/test_runtime_deploy_contract_text.py -q`
8. `uv run inv quality.max-lines --limit 300`
9. `uv run inv interface.e2e --server-mode=start --project=chromium --workers=1 --spec=e2e/specs/schedule-rules.spec.ts`
10. `uv run pytest tests/test_media_gateway.py -q`
11. `uv run inv ci.release-preflight --lane=baseline --no-e2e`

Known non-blocking warnings:

- Interface unit tests still emit existing React `act(...)` and localstorage-file warnings, but the suite passes.

## Next Run Targets

| Priority | Target | Action | Owner Lane | Exit Criteria |
| --- | --- | --- | --- | --- |
| P0 | Native service kickoff | Start NATS, Core API, and Frontend through repo tasks in native mode and verify authenticated UI access. | Deployment/Proof + QA/Embodiment | `lifecycle.status` shows services up and browser reaches the Soma dashboard without auth-state loops. |
| P0 | Full workflow proof standard | For each product slice, prove both UI workflow and API/runtime contract across create/propose/approve/execute/output/proof/recovery/reload where applicable. | QA/Embodiment | Test plan names every user-visible and API-visible step touched by the slice. |
| P0 | Media unblock | Start Forge/AUTOMATIC1111 or ComfyUI and the media gateway; validate retained output from browser UI. | Media Lane | Live media retained-output proof passes with open/viewable artifact evidence and obvious open-folder affordance. |
| P1 | Toolchain alignment | Keep baseline preflight green with local Go `1.26.3`, Node `v25.2.1`, and npm `11.6.2`. | Deployment/Proof | Toolchain check remains green and baseline remains green. |
| P1 | Visual expression review | Run focused review on Dashboard/Soma, Automations, Approvals, Groups, Resources, and System after each touched slice. | Visual/UI | No overlap, no confusing density, no stale wording, and target Soma-governed expression is preserved. |
| P2 | Promotion proof | After local source proof is green and services are intentionally up, run Compose, WSL, Rancher/K8s, and hosted/manual workflows as corroboration. | Deployment/Proof | Promotion proof records exact commit, environment, commands, and pass/fail result. |
| P2 | Documentation hygiene | Keep canonical docs compact and linked; avoid turning state back into a historical transcript. | Docs/State | State file stays as active scoreboard; deep evidence lives in commit history and owning docs. |

## Team Engagement Plan

| Team | Immediate Assignment | Coordination Rule |
| --- | --- | --- |
| Scheduler/Cadence | Hold the approved handoff contract steady while live run creation remains a later governed slice. | Do not weaken the no-autonomous-execution boundary. |
| Visual/UI | Review active media output, focused team context, output folder access, and dashboard density before new feature expansion. | Browser proof required for visible workflow changes. |
| Deployment/Proof | Own native service startup, authenticated UI access, and ordered promotion proof. | Source proof first; deployment proof corroborates after local correctness. |
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

## Release-Proof Sequencing Anchor

This compact anchor preserves the release-proof contract required by the active test suite without expanding state into a historical transcript.

- Use the guarded `uv run inv wsl.validate --lane=release` path from the refreshed `mycelis-root` deployment checkout for release-style WSL proof.
- `wsl.validate` maps `--lane=service` and `--lane=release` to `ci.release-preflight --lane=runtime --no-e2e`.
- keep the focused `/runs` and guided retry/recovery browser proofs green, refresh the WSL proof checkout from the committed slice, run `wsl.validate`, and then rerun broader headed certification from committed state
- Always run `uv run inv wsl.validate` from the refreshed WSL proof checkout before accepting the new browser-gap evidence as authoritative.
- The live MCP workflow correlation is now green from the refreshed WSL proof checkout.
- Next release-proof promotion should run `wsl.validate` from the refreshed proof checkout, keep the new `/runs` and guided Soma retry/recovery browser workflow proofs green, then rerun the broader headed Chromium certification pass from committed state.

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
