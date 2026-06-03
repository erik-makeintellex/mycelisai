# Mycelis V8 - Active Development State
> Navigation: [Project README](../README.md) | [Docs Home](../docs/README.md)

> Updated: 2026-06-03
> Canonical active scoreboard for V8 delivery. Historical V7/V8 migration evidence remains in `.state/V7_DEV_STATE.md`, git history, and canonical architecture docs.

## Current Checkpoint

| Field | State |
| --- | --- |
| Current committed checkpoint | `Prioritize team work attention` |
| Working tree at checkpoint | Next-phase UI execution slice: closed Soma review panels no longer mount bulky Work/Output/Trust content, active work keeps Soma readable by suppressing duplicate output previews, latest generated outputs use plain `Open file` / `Open folder` guidance with verification collapsed, and approval feedback now refuses proposals missing executable proof linkage while showing a durable run-started message. |
| Release proof | `uv run inv ci.release-preflight --lane=baseline --no-e2e` passed on checkpoint `11ca823e`; next-phase state committed at `c33216f6` |
| Focused browser proof | Full Chromium UI pass on 2026-06-02 covered login/front door, dashboard workbench scroll containment, focused team output dock, live governed execution, live Ask Team, live team-generated playable output, Groups retention, retained media, and focused ComfyUI media journey. Recovery/degradation UI proof passes in mocked mode. The headed business-owner live dashboard proof now passes on visible port `3000`: clean fresh entry, proposal approval/running feedback, retained output, visible output digest file/folder actions, generated workspace file, output rail showing the current file instead of `Guided proposal`, Output-first review opening, and review-panel scroll containment without page scroll. Mocked focused-team dashboard proof verifies selected team context, newest focused output priority, open/folder controls, output review rail, and reload retention; live focused-team proof now verifies the latest-output primary action opens the generated playable file rather than the team folder, and the same live path now asserts durable team-work API readback. |
| Product posture | Governed schedule proposals are durable and reviewable; scheduler ticks do not execute work autonomously |
| Main release risk | Async team ask acceptance, correlated status/result projection, retained output-ref projection, focused output access, and live team output review are operator-visible. Remaining release-candidate risk is browser-certified interaction clarity across richer team/media workflows and post-approval team-work feedback. Focused team context, workspace-path output refs, closed-panel dashboard simplification, latest-output discovery, and proof-linked proposal execution feedback are now under unit/type proof; the next risk is live browser proof of the combined focused-team/media package journey. |
| Environment blocker | None for local ComfyUI media proof while ComfyUI and the gateway remain online |

## Delivery Status

| Lane | Status | Where We Are | Next Target | Proof Gate |
| --- | --- | --- | --- | --- |
| Scheduler/Cadence | `COMPLETE` | Schedule due ticks persist idempotent handoff records with `handoff_key`, intent proof, execution contract, proposal status, bounded payload, and explicit terminal handoff approval transitions (`approved`, `rejected`, `cancelled`). | Keep the no-autonomous-execution boundary intact while later slices connect handoff approval into live run creation. | Go scheduler/history/triggers tests, API history proof, autonomous-execution guard review. |
| UI Workflow/Expression | `IN_REVIEW` | Schedule Rules and Approvals are complete. Dashboard team focus now appears only when there is focused or active/output work, so standing teams no longer crowd the clean root Soma surface. Ask Team now appears as queued non-blocking work immediately, Active Work polls active durable rows until terminal state, correlated team signals can advance queued rows to output-ready/degraded, retained `outputs[]`/`output_refs[]` project back onto the original row, and focused team contexts now keep chat/work/output/proof scoped together. Proposal approval now shows immediate running feedback, refuses malformed proposals that cannot update proof-linked lifecycle, and appends a clear run-started message. Output summaries present the latest generated file-like output first with plain `Open file` / `Open folder` guidance, visible workspace path hints, proof-backed/openable outputs only, and verification details collapsed by default. Closed review panels no longer mount bulky Work/Output/Trust content; when queued/running/degraded/operator-needed work exists, Soma opens Work first and hides duplicate output previews so the next action stays obvious; finished-output flows remain output-first. Folder reveal feedback now uses direct user language (`Folder opened` / `Open failed`). `/dashboard?fresh=1` clears browser-persisted Soma chat/session scopes for fresh interaction testing. | Continue reducing dashboard setup text and run the same attention-first/output-first proof across richer team/media/package journeys. | Focused Vitest, TypeScript, max-lines, and headed/browser proof where the workflow is visible. |
| API/Runtime Contract | `IN_REVIEW` | Trigger execution history includes schedule handoff proof fields, and `POST /api/v1/triggers/{id}/history/{executionId}/approval` transitions persisted schedule handoff state only. Focused `/api/v1/chat` requests preserve selected `team_id` for proposal wiring and conversation turns. Durable team `output_refs[].storage_ref` now stores workspace paths/folders instead of viewer URLs for both signal-projected refs and confirm-action deliverables, including viewer URLs that arrive through execution-output `folder` fields. Confirm-action success responses now include `data.team_work_refs[]` for newly persisted create-team, delegated, and deliverable work visibility. | Keep the live chat -> confirm -> team work -> status events -> reveal proof green while extending richer media/package journeys. | Go API tests, docs/API reference review, history reload proof, focused live GUI proof. |
| Deployment/Proof | `ACTIVE` | Baseline source release-preflight is green; local Go reports `go1.26.3`. Native NATS, Core API, Frontend, PostgreSQL, Ollama, ComfyUI, and the media gateway are currently up and healthy. Fresh Soma interaction reset cleared volatile run/team context, retained memory vectors, artifacts, collaboration groups, exchange items, workspace outputs, and reports while keeping auth, providers, capability manifests, and deployment config intact. Latest focused live dashboard proof found no stale dashboard media/content and approval created a retained output in `MYCELIS_WORKSPACE`. | Run one coherent live Soma journey across team/media/output context before Compose/WSL/K8s promotion. | `ci.release-preflight`, focused live GUI, `compose.health`, `wsl.validate --lane=release`, K8s proof. |
| Docs/State | `ACTIVE` | API reference, Teams/Soma user docs, automation user docs, and this scoreboard reflect the committed scheduler handoff slice, next-phase kickoff, live GUI proof, live ComfyUI media proof, current max-line convergence proof, focused-team context propagation, workspace-path output-ref semantics, and clearer output re-entry behavior. | Keep docs/state synchronized in the same slice as behavior, API, runtime, workflow, testing, or terminology changes. | docs/workflow tests, `git diff --check`, max-lines. |
| Media Lane | `COMPLETE` | Media gateway unit coverage, direct ComfyUI gateway smoke, mocked retained-media browser proof, live retained-media UI proof, and focused ComfyUI journey proof are green. Windows-host optional Diffusers/vLLM helpers remain unsupported, so local image generation uses Pinokio ComfyUI through the gateway. Runtime media calls now fail closed on provider HTTP/parse/no-image errors so stale cached media is not saved as a fresh output. | Keep media visible in Soma/team output views and carry the same proof standard into richer team-generated media packages. | `tests/test_media_gateway.py`, direct gateway generation smoke, mocked retained-media browser proof, live retained-output proof, focused ComfyUI journey proof, and `go test ./internal/swarm -run "TestHandleGenerateImage"`. |

## Proof Evidence

Latest green proof through next-phase native-services engagement:

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
12. `uv run inv lifecycle.up --frontend --build`
13. `uv run inv lifecycle.status`
14. `uv run inv lifecycle.health`
15. `uv run inv native-infra.status`
16. `uv run inv db.clear-runtime-context --yes`
17. `uv run inv interface.e2e --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/homepage.spec.ts`
18. `uv run inv interface.e2e --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/dashboard-workbench-live-review.spec.ts`
19. `uv run inv interface.e2e --project=chromium --workers=1 --spec=e2e/specs/soma-media-retained-output-live.spec.ts`
20. `uv run inv lifecycle.up --frontend`
21. `curl.exe -fsS http://127.0.0.1:8188/system_stats`
22. `curl.exe -fsS http://127.0.0.1:8001/health`
23. `Invoke-RestMethod -Method Post -Uri http://127.0.0.1:8001/v1/images/generations ...` returned `data[0].b64_json`
24. `$env:PLAYWRIGHT_LIVE_MEDIA_RETAINED_OUTPUT='1'; uv run inv interface.e2e --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/soma-media-retained-output-live.spec.ts`
25. `uv run inv interface.e2e --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/ui-finalization-browser-package-live.spec.ts`
26. `uv run inv interface.e2e --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/soma-governance-live.spec.ts`
27. `uv run inv interface.e2e --project=chromium --workers=1 --spec=e2e/specs/soma-proposal-mode.spec.ts`
28. `uv run inv lifecycle.up --frontend`
29. `uv run pytest tests/test_media_gateway.py -q`
30. `npx tsc --noEmit`
31. `uv run inv quality.max-lines --limit 300`
32. `$env:PLAYWRIGHT_LIVE_MEDIA_RETAINED_OUTPUT='1'; uv run inv interface.e2e --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/soma-media-comfyui-journey.spec.ts`
33. `npm test -- SomaOperatingSurface.actions.test.tsx SomaOperatingSurface.test.ts CentralSomaHome.test.tsx MissionControlChat.header.test.tsx OutputWorkbench.test.tsx ActiveWorkLane.test.tsx`
34. `npx tsc --noEmit`
35. `uv run inv quality.max-lines --limit 300`
36. `uv run inv interface.e2e --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/desktop-mobile-compression.spec.ts`
37. `uv run inv interface.e2e --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/dashboard-workbench-live-review.spec.ts`
38. `go test ./internal/server -run "TestHandleTeamWorkAsk|TestTeamWorkSignal"`
39. `npm test -- TeamsPage.ask.test.tsx ActiveWorkLane.test.tsx SomaOperatingSurface.actions.test.tsx useDurableTeamWork.test.tsx`
40. `npx tsc --noEmit`
41. `uv run pytest tests/test_docs_links.py tests/test_workflow_contracts.py tests/test_runtime_deploy_contract_text.py -q`
42. `uv run inv quality.max-lines --limit 300`
43. `git diff --check`
44. `go test ./internal/swarm -run "TestTeam_ResponseDelivery|TestTeam_StatusDelivery|TestTeam_TriggerLogic_UnwrapsCommandEnvelope|TestNormalizeCommandPayload"`
45. `go test ./internal/server -run "TestTeamWorkSignalProjection|TestHandleTeamWorkAsk"`
46. `uv run pytest tests/test_docs_links.py tests/test_workflow_contracts.py tests/test_runtime_deploy_contract_text.py -q`
47. `uv run inv quality.max-lines --limit 300`
48. `git diff --check`
49. `go test ./internal/server -run "TestTeamWorkSignalProjection|TestHandleTeamWorkAsk"` proves correlated result `outputs[]` and `output_refs[]` now persist openable team output refs on the original Active Work item.
50. `go test ./internal/swarm -run "TestTeam_ResponseDelivery|TestTeam_StatusDelivery|TestTeam_TriggerLogic_UnwrapsCommandEnvelope|TestNormalizeCommandPayload"` remains green for team signal delivery.
51. `npm test -- SomaOperatingSurface.actions.test.tsx ActiveWorkLane.test.tsx` proves focused team retained outputs appear before the work panel is opened while compact Active Work evidence remains green.
52. `npx tsc --noEmit` remains green after adding the focused team output dock.
53. `npm test -- SomaOperatingSurface.actions.test.tsx ActiveWorkLane.test.tsx OutputWorkbench.test.tsx useDurableTeamWork.test.tsx TeamsPage.ask.test.tsx`
54. `npx tsc --noEmit`
55. `uv run inv interface.e2e --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/desktop-mobile-compression.spec.ts`
56. `uv run inv interface.e2e --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/homepage.spec.ts`
57. `uv run inv interface.e2e --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/dashboard-workbench-live-review.spec.ts`
58. `uv run inv interface.e2e --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/soma-governance-live.spec.ts`
59. `uv run inv interface.e2e --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/team-output-content-live.spec.ts`
60. `$env:PLAYWRIGHT_TEAM_WORK_GUI_LIVE='1'; $env:PLAYWRIGHT_TEAM_WORK_API_TEAM_ID='admin-core'; uv run inv interface.e2e --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/active-work-ask-live.spec.ts`
61. `$env:PLAYWRIGHT_LIVE_MEDIA_RETAINED_OUTPUT='1'; uv run inv interface.e2e --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/soma-media-retained-output-live.spec.ts`
62. `$env:PLAYWRIGHT_LIVE_MEDIA_RETAINED_OUTPUT='1'; uv run inv interface.e2e --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/soma-media-comfyui-journey.spec.ts`
63. `go test ./internal/server` proves confirmation summary, focused team latest-output context, and server regressions remain green.
64. `npm test -- --run __tests__/dashboard/ProposedActionBlock.test.tsx __tests__/dashboard/OutputWorkbench.test.tsx __tests__/store/useCortexStore.confirm-proposal.pending-proof.test.ts __tests__/dashboard/MissionControlChat.executionSummaryMedia.test.tsx` proves immediate approval feedback, pending-proof state, latest-output workbench, and media output actions.
65. `npm run build` proves the production Interface bundle after the approval/output UX correction.
66. `npx playwright test e2e/specs/desktop-mobile-compression.spec.ts e2e/specs/soma-media-artifacts.spec.ts --project=chromium` proves browser-level dashboard compression, focused team output access, and media artifact rendering.
67. `uv run inv db.clear-runtime-context --yes --include-memory-vectors` cleared volatile Soma/team state and retained context vectors for fresh interaction proof.
68. `DELETE FROM exchange_items; DELETE FROM artifacts; DELETE FROM collaboration_groups;` cleared old generated/test content that remained visible outside run tables.
69. `uv run inv clean.reports` removed Playwright/test report residue.
70. `uv run inv interface.restart` rebuilt and restarted the Interface; page health checks passed for `/`, `/dashboard`, `/teams`, `/approvals`, and other core routes.
71. `npm test -- --run __tests__/dashboard/SomaOperatingSurface.actions.test.tsx __tests__/dashboard/MissionControlChat.header.test.tsx __tests__/dashboard/OutputWorkbench.test.tsx __tests__/dashboard/ProposedActionBlock.test.tsx __tests__/store/useCortexStore.confirm-proposal.pending-proof.test.ts` proves root dashboard context compression, `?fresh=1` client chat reset, approval feedback, and output workbench behavior.
72. `npx tsc --noEmit` remains green after the fresh-dashboard cleanup.
73. `npx playwright test e2e/specs/desktop-mobile-compression.spec.ts --project=chromium` proves browser-level dashboard compression remains green.
74. `npx tsc --noEmit` remains green after adding the business-owner fresh dashboard/proposal e2e contract.
75. `$env:PLAYWRIGHT_LIVE_BACKEND='1'; $env:PLAYWRIGHT_SKIP_WEBSERVER='1'; $env:PLAYWRIGHT_PORT='3000'; npx playwright test e2e/specs/dashboard-workbench-live-review.spec.ts --project=chromium --grep "fresh business-owner"` passes after rebuilding Core and restarting the visible Interface: clean fresh dashboard state, proposal approval, retained output, and generated workspace file `E:\random\test\generated\business-owner-flow-1780423995673\owner-note.md`.
76. `go test ./internal/swarm -run "TestHandleGenerateImage"` proves media generation now fails closed on provider HTTP failure and missing image data instead of allowing stale cached media to be saved as fresh output.
77. `go test ./internal/server -run "TestBuildConfirmActionExecutionSummaryNamesTeamMediaDeliverable|TestExecutionOutputsFromArtifactsUsesWorkspaceViewerForSavedMedia"` proves retained media/file output summary wiring remains green after artifact-id handoff tightening.
78. `npm test -- --run __tests__/store/useCortexStore.confirm-proposal.execution.test.ts __tests__/dashboard/OutputWorkbench.test.tsx __tests__/dashboard/MissionControlChat.executionSummary.test.tsx __tests__/dashboard/MissionControlChat.executionSummaryMedia.test.tsx __tests__/dashboard/SomaOperatingSurface.actions.test.tsx __tests__/dashboard/CentralSomaHome.test.tsx` proves clicked proposal token precedence, result-card compression, larger output actions, and simplified cold dashboard tests.
79. `npx tsc --noEmit` remains green after the result-card, context-switcher, output-action, and confirmation-store changes.
80. `$env:PLAYWRIGHT_LIVE_BACKEND='1'; $env:PLAYWRIGHT_SKIP_WEBSERVER='1'; $env:PLAYWRIGHT_PORT='3000'; npx playwright test e2e/specs/dashboard-workbench-live-review.spec.ts --project=chromium --headed` passes in a visible browser: clean business-owner entry, approval/running feedback, retained output, output digest with `Open file` and `Open folder`, generated workspace file `E:\random\test\generated\business-owner-flow-1780453021517\owner-note.md`, review rail showing the current file instead of `Guided proposal`, Output-first review opening, and contained side-rail scrolling with page scroll remaining at `0`.
81. `npm test -- --run __tests__/dashboard/SomaCausalSummary.test.tsx` and `npx tsc --noEmit` pass after extracting `SomaCausalSummaryModel.ts`; `SomaCausalSummary.tsx` is no longer a max-line violation.
82. `npm test -- --run __tests__/dashboard/SomaWorkspaceFrame.test.tsx __tests__/dashboard/OutputWorkbench.test.tsx __tests__/dashboard/ProposedActionBlock.test.tsx` and `npx tsc --noEmit` pass after adding the output digest and extracting `ProposedActionDetails.tsx`/`proposedActionCopy.ts`. `ProposedActionBlock.tsx` is no longer a max-line violation.
83. `go test ./internal/server`, `uv run inv quality.max-lines --limit 300`, and `git diff --check` pass after splitting `cognitive_chat_helpers_test.go`, `cognitive_tool_plan_parse.go`, and `confirm_action_visibility_test.go` into focused helper files. The max-line gate now has no uncapped violations; legacy capped files remain recorded under `ops/quality_legacy_caps.txt`.
84. `go test ./internal/server`, `npm test -- --run __tests__/dashboard/OutputWorkbench.test.tsx __tests__/dashboard/SomaOperatingSurface.actions.test.tsx __tests__/dashboard/useDurableTeamWork.test.tsx`, `npx tsc --noEmit`, and `uv run inv quality.max-lines --limit 300` pass after focused-team context propagation, workspace-path `output_refs` normalization, and focused-output newest-first UI priority.
85. `npx playwright test e2e/specs/desktop-mobile-compression.spec.ts --project=chromium --grep "focused team output dock"` passes, proving the focused team output dock stays scannable in browser proof after the output-priority/context changes.
86. `npm test -- --run __tests__/dashboard/OutputWorkbench.test.tsx __tests__/dashboard/SomaOperatingSurface.actions.test.tsx __tests__/dashboard/SomaWorkspaceFrame.test.tsx __tests__/dashboard/CentralSomaHome.test.tsx __tests__/resources/WorkspaceExplorer.test.tsx`, `npx tsc --noEmit`, `uv run inv quality.max-lines --limit 300`, and `uv run inv interface.e2e --project=chromium --workers=1 --spec=e2e/specs/focused-team-output-dashboard.spec.ts` pass after dashboard copy compression, workspace path hints, clearer folder reveal feedback, and focused-team reload browser proof.
87. `npm test -- --run __tests__/dashboard/OutputWorkbench.test.tsx __tests__/dashboard/SomaWorkspaceFrame.test.tsx` and `uv run inv interface.e2e --live-backend --server-mode=start --project=chromium --workers=1 --spec=team-output-content-live.spec.ts` pass after latest-output selection now prefers generated file artifacts over team folders. The live proof creates a new team, confirms the governed action, opens the generated playable HTML file, verifies interaction, reveals the containing folder, verifies focused dashboard re-entry, and confirms Groups retention.
88. `npm test -- SomaOperatingSurface.actions.test.tsx`, `npm test -- --run __tests__/dashboard/SomaWorkspaceFrame.test.tsx __tests__/dashboard/SomaOperatingSurface.actions.test.tsx __tests__/dashboard/OutputWorkbench.test.tsx`, `npx tsc --noEmit`, `go test ./internal/server -run "Test(OutputRefsForTeamWork_NormalizesViewerURLFolderForDeliverable|PersistConfirmedActionTeamWork_DeliverableOutputReadyHasRefs|OutputRefFromMapNormalizesViewerURLToWorkspaceStorageRef|OutputRefFromMapDerivesFilePathFromViewerHrefForMedia)" -count=1`, `uv run inv core.compile`, and `uv run inv interface.e2e --live-backend --server-mode=start --project=chromium --workers=1 --spec=team-output-content-live.spec.ts` pass after team-focus wording cleanup, attention-first work review behavior, and live API readback coverage. The first live readback run caught confirm-action deliverables persisting viewer URLs in `output_refs[].storage_ref`; backend normalization now decodes viewer URLs from execution-output `folder` fields, and the rerun passed.
89. `npm test -- --run __tests__/dashboard/ProposedActionBlock.test.tsx __tests__/store/useCortexStore.confirm-proposal.execution.test.ts __tests__/dashboard/OutputWorkbench.test.tsx __tests__/dashboard/SomaWorkspaceFrame.test.tsx`, `npx tsc --noEmit`, and `uv run inv quality.max-lines --limit 300` pass after the closed workbench panel stopped mounting bulky review content, latest output discovery gained plain file/folder guidance with verification collapsed, and proposal approval gained proof-link guards plus explicit run-started feedback.
90. `uv run inv interface.e2e --live-backend --server-mode=start --project=chromium --workers=1 --spec=e2e/specs/dashboard-workbench-live-review.spec.ts` passes after the compact latest-output digest stopped duplicating exact path text and the browser proof opens collapsed verification details before asserting path/readback/hash proof. The proof covers fresh business-owner entry, approval, retained output, file/folder controls, output review rail, collapsed verification, Trust rail wording, and contained rail scrolling without page scroll.
91. `go test ./internal/server -run 'Test(HandleConfirmAction_CompletesVerifiedExecutionWithPlannedToolCalls|PersistConfirmedActionTeamWork_|ConfirmActionResponseDataIncludesTeamWorkRefs)'`, `go test ./internal/server -run 'Test.*(ConfirmAction|TeamWork)'`, `go test ./internal/server -count=1`, `npm test -- --run __tests__/store/useCortexStore.confirm-proposal.execution.test.ts __tests__/store/useCortexStore.confirm-proposal.team-work-refs.test.ts`, `npx tsc --noEmit`, `uv run pytest tests/test_docs_links.py tests/test_workflow_contracts.py tests/test_runtime_deploy_contract_text.py -q`, `uv run inv quality.max-lines --limit 300`, and `git diff --check` pass after confirm-action success responses gained explicit durable team-work refs and the frontend started surfacing concise Active Work cues from those refs.

Known non-blocking warnings:

- Interface unit tests still emit existing React `act(...)` and localstorage-file warnings, but the suite passes.

## Next Run Targets

| Priority | Target | Action | Owner Lane | Exit Criteria |
| --- | --- | --- | --- | --- |
| P0 | Full workflow proof standard | For each product slice, prove both UI workflow and API/runtime contract across create/propose/approve/execute/output/proof/recovery/reload where applicable. | QA/Embodiment | Test plan names every user-visible and API-visible step touched by the slice. |
| P0 | Async team execution | Ask Team now queues without blocking the operator, keeps Active Work polling while work is active, preserves `work_item_id` through team status/result signals, and projects retained `outputs[]`/`output_refs[]` from correlated result payloads onto the original work item. | Runtime Teams + QA/Embodiment | Team work starts quickly, remains visible in the focused context, and returns retained output/proof or actionable degradation. |
| P0 | Team output usability | Focused team contexts now surface retained output refs in a compact Team outputs dock while preserving the detailed output list in the Work panel/Team page. Focused output ordering now prefers newest selected-team refs over older root Soma outputs, and durable output refs use workspace paths/folders rather than viewer URLs. Live team output proof already creates a playable file, opens it, reveals the folder, returns to focused Dashboard, and verifies the Team outputs dock. | Media Lane + Visual/UI | Generated media appears in the relevant team/output context with one-click file/folder access and proof/recovery nearby, then remains correct after focused-team switching/reload. |
| P0 | Live approval clarity proof | Unit/type proof now covers proof-linked approval feedback and malformed proposal blocking. | UI + QA/Embodiment | Run visible browser proof that a business-owner user clicks approval, sees clear run-started feedback, sees active work or output without stale content, and can open the latest file/folder. |
| P0 | Fresh dashboard approval output | Ask -> proposal -> approve -> retained output is live for the business-owner file-generation path on visible port `3000`; stale retained content is rejected before the ask. | Runtime + QA/Docs | Keep this proof green while expanding the same standard to team/media follow-up work and richer output packages. |
| P1 | Toolchain alignment | Keep baseline preflight green with local Go `1.26.3`, Node `v25.2.1`, and npm `11.6.2`. | Deployment/Proof | Toolchain check remains green and baseline remains green. |
| P1 | Visual expression review | Run focused review on Dashboard/Soma, Automations, Approvals, Groups, Resources, and System after each touched slice. | Visual/UI | No overlap, no confusing density, no stale wording, and target Soma-governed expression is preserved. |
| P1 | File-size convergence | Keep the max-line gate green for new work and reduce legacy capped files opportunistically without blocking focused product delivery. | Quality | `quality.max-lines` has no uncapped violations and behavior-focused tests remain green. |
| P2 | Promotion proof | After local source proof is green and services are intentionally up, run Compose, WSL, Rancher/K8s, and hosted/manual workflows as corroboration. | Deployment/Proof | Promotion proof records exact commit, environment, commands, and pass/fail result. |
| P2 | Documentation hygiene | Keep canonical docs compact and linked; avoid turning state back into a historical transcript. | Docs/State | State file stays as active scoreboard; deep evidence lives in commit history and owning docs. |

## Team Engagement Plan

| Team | Immediate Assignment | Coordination Rule |
| --- | --- | --- |
| Scheduler/Cadence | Hold the approved handoff contract steady while live run creation remains a later governed slice. | Do not weaken the no-autonomous-execution boundary. |
| Visual/UI | Review active media output, focused team context switching, output folder access, and dashboard density before new feature expansion. | Browser proof required for visible workflow changes. |
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
