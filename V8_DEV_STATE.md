# Mycelis V8 - Development State

> Updated: 2026-04-04
> Canonical state file for active V8 grading and delivery tracking
> References: `README.md`, `v8-2.md`, `docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md`, `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`, `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`, `docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md`, `V7_DEV_STATE.md` (legacy migration input)

---

## Purpose

`V8_DEV_STATE.md` is now the authoritative global state file for active V8 grading.

Use this file to track:
- current V8 delivery status
- accepted marker transitions
- gate evidence
- blockers and dependencies
- immediate next actions

Use `V7_DEV_STATE.md` only as a migration input and historical checkpoint source while V7 planning artifacts are being migrated.

## Layered Architecture Truth

Development is progressing toward the V8.2 full production target.

- V8.1 is the current bounded release target.
- V8.2 is the distributed, learning, capability-enabled, and actuation target beyond the current release.
- `README.md` is the primary inception summary that must distinguish those two layers and point to this live implementation state.

Release posture:
- `ACTIVE` current posture is `V1 MVP Release Candidate`.
- `COMPLETE` the previously release-blocking Soma action integrity recovery pass is now closed for the current RC path: governed mutation requests enter proposal mode, proposal generation is side-effect free, and confirmation produces durable execution proof on the validated live lane.
- `COMPLETE` the default Soma-primary operator flow is now release-candidate ready with guided first-run actions, resilient support panels, and user-facing vocabulary aligned around Automations, Memory & Continuity, AI Engine, Response Style, Advisors, and Departments.
- `COMPLETE` Soma is now the primary default interface and organization orchestrator for the AI Organization workspace.
- `COMPLETE` Team Leads remain the operational layer Soma works through, while Routing / Council remains advisory and planning-oriented rather than default visible UX jargon.
- `COMPLETE` the default workspace remains simple, guided, and non-technical, and the advanced/runtime surface remains separate from the default operator flow.
- `COMPLETE` the broader interface suite is now green again, and the RC gate is based on full unit coverage, typecheck, the managed browser suite, and docs-link enforcement rather than only the targeted Soma path.
- `COMPLETE` the managed exchange foundation now exists for governed channels, typed schemas, structured fields, threads, normalized MCP/tool outputs, and advanced inspect-only visibility over recent exchange activity.
- `COMPLETE` the managed exchange security foundation now exists for permissioned channel access, thread participation and review rights, artifact sensitivity classes, capability risk classes, trust-classified external/MCP outputs, and audit-ready metadata on exchange items.
- `COMPLETE` Chunk 8.4 is now closed: managed exchange and capability security foundations are implemented, validated, and committed.
- `IN_REVIEW` Soma cognitive-engine availability now validates chat-engine binding at startup/runtime, rebinds default execution profiles to a safe local fallback when available, and turns missing-engine failures into actionable setup guidance instead of generic chat breakage.
- `COMPLETE` accepted non-blockers for the RC gate are now limited to repeated `--localstorage-file` path warnings during Vitest worker startup; the earlier `SensorLibrary` and `CircuitBoard` warning noise has been cleaned out of the repo-owned test surface.
- the default Soma-primary workspace now expects guided first-run actions, intentional empty states, and partial-failure-safe support panels before release lock.
- `NEXT` advanced architecture/runtime configuration remains a separate contract and implementation lane; it must stay non-default until the dedicated advanced surface ships without polluting the MVP operator flow.
- `NEXT` extend the same cognitive-engine availability contract into broader team, automation, and council execution paths beyond the primary Soma chat route.
- `NEXT` build on the managed exchange foundation with richer persistent team runs, deeper inter-agent review loops, and broader channel-aware orchestration across MCP-backed services.
- `COMPLETE` Chunk 8.7 is now landed for the free-node release: Soma reads a user-level governance profile, mutation proposals now carry approval posture and capability-risk context, and audit records cover proposal generation, proposal confirmation/cancellation, execution runs, capability use, artifact creation, and channel writes.
- `IN_REVIEW` the dedicated UI-testing stabilization strike plan has now cleared the live governed-chat blocker and is primarily a release-hygiene and packaging lane.
- `IN_REVIEW` the new stable UI-testing agentry contract is now documented and partially automated: mocked browser proof for Soma-first entry, direct answer, continuity, cold-start recovery, governed proposal/cancel, audit visibility, and oversized-content handling is green.
- `COMPLETE` the live governed-chat browser gate is now green on a fresh cluster reset and local stack bring-up: the managed live backend proof passes the real `/api/v1/chat` and `/api/v1/intent/confirm-action` path with governed proposal/confirm behavior intact.
- `IN_REVIEW` a cross-team release platform review is now active to align security/governance, monitoring/ops, debug/live-browser proof, and matching documentation before release packaging.
- `NEXT` a dedicated memory continuity and RAG review lane should align pgvector-backed durable memory, temporary planning continuity, and trace-clean conversation handling after the current product-trust closeout slices.
- `ACTIVE` a dedicated content-generation and collaboration lane is now in flight to align inline content delivery, governed artifact creation, media/file generation behavior, and policy-configurable specialist/model collaboration for the current product-trust and demo-value closeout.
- `ACTIVE` a dedicated approval and product-trust lane is now in flight to simplify default approval/auth-style interactions, move low-level governance metadata behind inspectable details, and make content/artifact value delivery understandable to normal operators without weakening policy enforcement.
- `COMPLETE` a supported single-host Docker Compose runtime now exists for home-lab and demo use: compose task automation, env separation, host-port parity, compose-managed migrations, and compose health/status/log inspection are now part of the repo contract.
- `ACTIVE` a true-MVP finish lane is now defined to converge product value delivery, settings completion, demo readiness, and clean release proof into one prioritized close-out program.
- `ACTIVE` a dedicated demo-product strike team is now in flight to make Mycelis legible as an obvious product for technical partners/funders while preserving Soma capability and advanced power through advanced surfaces and documentation.
- `COMPLETE` the first demo-product execution brief is now in place: default-vs-advanced surfaces, feature-preservation rules, a golden-path partner demo, and the UI testing checklist are defined as immediate engaged-team deliverables.
- `ACTIVE` the MVP investor-demo lane is now explicitly centered on governed capability expansion: the active story must prove Soma’s initial value first, then show MCP-powered input/output, securable web/external research, and inspectable context-security posture without collapsing into a tooling console narrative.
- `IN_REVIEW` governed deployment-context intake now has a first-class operator path: Resources exposes a Deployment Context tab for loading customer-provided material into `customer_context` vectors and approved company-authored guidance into `company_knowledge` vectors, the backend persists both as governed document artifacts plus vector chunks with visibility/sensitivity/trust metadata, and Soma/Council/teams can now recall that context while keeping it separate from ordinary Soma memory.
- `ACTIVE` the follow-through plan for governed deployment context is now defined in `docs/architecture-library/V8_GOVERNED_DEPLOYMENT_CONTEXT_AND_RAG_PLAN.md`: next work is to verify shared runtime prompt coverage, add explicit promotion into `company_knowledge`, and make web/MCP-fed context loading policy-bound and inspectable.
- `IN_REVIEW` the first promotion path into `company_knowledge` now exists on the governed execution spine: a confirmed plan can promote an existing `customer_context` artifact into a new approved company-knowledge record with preserved lineage metadata instead of mutating the original source in place.
- `NEXT` enterprise identity, approval workflows, and multi-user access management remain deferred beyond the free-node governance foundation; the current release now has enterprise-capable governance foundations, not full enterprise identity management.

## Architecture Synchronization Rule

Every slice must:
- update state
- verify README alignment
- verify V8.2 alignment

Execution completion rule:
- a slice is not complete unless tests pass, documentation is updated where meaning changed, and architecture alignment is verified
- README, V8.1, V8.2, and this state file must remain synchronized when implementation or release meaning changes
- no silent divergence is allowed between current implementation, current release target, and full architecture target

State reporting rule:
- `COMPLETE` records accepted delivered work
- `ACTIVE` records work in progress
- `NEXT` records the next committed follow-on slices
- slice close-out should explicitly report tests run, docs changed, and docs reviewed unchanged for the touched scope

## Feature Status Legend

- `REQUIRED`: must exist for target delivery or gate pass
- `NEXT`: highest-priority upcoming slice
- `ACTIVE`: currently in development
- `IN_REVIEW`: implemented and awaiting validation/review
- `COMPLETE`: delivered and accepted
- `BLOCKED`: cannot advance until a dependency or defect is resolved

## Current Review (2026-04-04)

Review summary:
1. `COMPLETE` the top-level RC story is now sharper: the formerly release-blocking Soma action integrity lane is resolved on the validated live path, so the active closeout stack now centers on approval/product-trust simplification, true-MVP finish, and demo-product readiness rather than platform break/fix.
2. `IN_REVIEW` the state file now treats the UI-testing stabilization lane as a release-hygiene lane instead of a still-blocked runtime lane, because the live governed-chat browser gate is already green.
3. `NEXT` memory continuity/RAG and content-generation/collaboration remain important, but they are now explicitly staged after the current product-trust closeout slices instead of competing as parallel top-priority RC blockers.
4. `COMPLETE` the provider/auth inventory and local-model guidance are now synchronized across runtime defaults, auth implementation, and central docs: shipped `cognitive.yaml` local endpoints now match the AI Engines UI and local-engine helpers (`vllm -> 8000`, `lmstudio -> 1234`), Gemini auth now uses the documented `x-goog-api-key` header, repo-local `cognitive.*` helpers now fail clearly on unsupported Windows hosts instead of crashing through missing optional deps, and centralized docs now expose explicit TOCs plus cross-links for provider inventory, hosted-provider auth, and local-model switching.
5. `COMPLETE` provider auth contract tests now cover the listed hosted auth frameworks directly in runtime code: OpenAI Bearer auth, Anthropic `x-api-key` + version header, and Gemini `x-goog-api-key` are all asserted in `core/internal/cognitive/provider_auth_test.go`.
6. `COMPLETE` centralized guidance surfaces now link coherently from root and across associated docs: `README.md`, `docs/README.md`, `ops/README.md`, `docs/API_REFERENCE.md`, `docs/COGNITIVE_ARCHITECTURE.md`, `docs/LOCAL_DEV_WORKFLOW.md`, and `docs/architecture/OPERATIONS.md` now expose explicit TOCs or upgraded linkage for provider/auth/model support.
7. `ACTIVE` the approval and product-trust correction is now a first-class delivery lane: `docs/architecture-library/V8_APPROVAL_AND_PRODUCT_TRUST_STRIKE_TEAM_PLAN.md` defines the team structure, phase order, non-negotiable product rules, and acceptance gates for simplifying approval/auth-style interactions without flattening platform depth.
8. `ACTIVE` this lane is explicitly downstream of the universal-Soma correction and upstream of true-MVP closeout: approval moments now need to feel like trust interactions with one persistent Soma operating across governed contexts, not like isolated technical workflows bound to a single organization surface.
9. `ACTIVE` the default product problem is now named precisely: users are still being exposed to too much raw governance detail before they receive value, while content/media/file requests can still be perceived as “asked for permission but did not clearly deliver content or clearly reference the result.”
10. `ACTIVE` the strike-team plan now locks the non-negotiable posture that not all MCP/specialist/model collaboration should be manually approval-gated by default; the correct standard is policy-configurable governance tied to risk, mutation impact, external exposure, cost, and integration trust.
11. `COMPLETE` Phase 0 truth mapping now has canonical outputs in `docs/architecture-library/V8_APPROVAL_AND_PRODUCT_TRUST_STRIKE_TEAM_PLAN.md`: proposal fields are classified into `default-visible`, `details-only`, and `runtime-only`, and request types are classified into `answer`, `governed artifact`, `optional approval`, and `required approval`.
12. `COMPLETE` the repo-hygiene packaging checkpoint is now closed at commit `8344f33` (`Package approval trust and task-contract cleanup`): the approval/product-trust proposal-card slice, the paused push-triggered workflow posture, and the invoke task-surface cleanup are committed from a clean tree instead of being carried as lingering local drift.
13. `IN_REVIEW` the first approval-surface simplification slice is now wired end to end and packaged cleanly: runtime proposal payloads carry `operator_summary`, `expected_result`, and `affected_resources`, the dashboard proposal card now leads with user-legible action/result/change framing, and advanced mechanics move behind an explicit details control instead of leading the default surface.
14. `COMPLETE` targeted proof for the packaged checkpoint is green from committed state: `uv run inv core.test`, `uv run inv interface.typecheck`, `cd interface; npx vitest run __tests__/dashboard/ProposedActionBlock.test.tsx`, and `$env:PYTHONPATH='.'; uv run pytest tests/test_tasks_root.py tests/test_misc_tasks.py tests/test_docs_links.py -q` all pass against the committed approval/task-contract batch.
15. `COMPLETE` GitHub workflow posture is now intentionally quieter until initial release readiness: push-triggered CI/dev-build/release runs are paused, while local proof, `pull_request` validation, and manual `workflow_dispatch` remain the active release-discipline path.
16. `COMPLETE` the invoke service/task contract is now cleaner against the current architecture: `uv run inv install` targets the supported default Core + Interface stack, optional local vLLM/Diffusers helpers are explicitly opt-in, and the stale `team.sensors`, `team.output`, and `team.test` helper tasks have been removed from the exposed invoke surface so the task layer no longer presents them as active runtime services.
17. `ACTIVE` Phase 2 of the content/artifact value-delivery lane is now underway with a concrete first slice: Launch Crew execution outcomes now surface returned artifact references in the modal itself when Soma returns durable outputs, instead of only reporting that a run was activated somewhere else.
18. `COMPLETE` the current local validation blockers from committed state are cleared again and the browser task contract now matches the passing stable lane: `uv run inv interface.e2e` defaults to managed `dev` mode for mocked browser proof, explicit `--server-mode=start` remains the stricter built/live path, the task docs/tests are synchronized to that contract, and `uv run inv interface.test` plus `uv run inv interface.e2e --project=chromium --spec=e2e/specs/v8-ui-testing-agentry.spec.ts` both pass from the updated committed-state slice.
19. `COMPLETE` the previously named compose fresh-organization direct-answer `500` no longer reproduces on a clean compose stack: Scenario A in `interface/e2e/specs/soma-governance-live.spec.ts` returns `mode: answer` as expected, and the compose-sensitive issue was reduced to a spec-side backend-workspace-root assumption for file-side-effect assertions rather than a fresh answer-path runtime defect.
20. `IN_REVIEW` the first cleanup-swarm reference-authority slice is now materially in flight: active onboarding and execution docs now put `V8_DEV_STATE.md` ahead of historical V7 state, the team/global-state protocol no longer tells contributors to update the wrong state file, live-backend browser-proof docs now use the explicit `--server-mode=start` contract, the compose Scenario A `500` note is reclassified as resolved/reframed spec drift instead of an active runtime blocker, and the cleanup lane is still prioritizing high-signal reference fixes before broader readability churn or speculative dead-code removal.
21. `ACTIVE` the MVP investor-demo lane is now documented in `docs/architecture-library/V8_MVP_INVESTOR_DEMO_AND_GOVERNED_CAPABILITY_PLAN.md` and synchronized with the partner demo script/checklist: the canonical story now explicitly covers Soma’s initial value, governed mutation, MCP-powered input/output, securable web/external research, and inspectable context-security posture through Connected Tools, Exchange, and Approvals surfaces.
22. `IN_REVIEW` focused investor-lane proof is mostly green from the current local slice: `cd core; go test ./internal/mcp ./internal/exchange ./internal/server -count=1`, `cd interface; npx vitest run __tests__/settings/MCPLibraryBrowser.test.tsx __tests__/settings/MCPToolRegistry.test.tsx __tests__/resources/ExchangeInspector.test.tsx __tests__/pages/ResourcesPage.test.tsx __tests__/automations/ApprovalsTab.test.tsx __tests__/dashboard/ProposedActionBlock.test.tsx __tests__/dashboard/MissionControlChat.test.tsx __tests__/pages/SettingsPage.test.tsx --reporter=dot`, `cd interface; npx tsc --noEmit`, `$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q`, and `uv run inv interface.e2e --server-mode=start --project=chromium --spec=e2e/specs/governance.spec.ts` all pass.
23. `BLOCKED` the stricter built-server settings browser proof still has one investor-readiness mismatch to resolve before the lane can be treated as fully green: `uv run inv interface.e2e --server-mode=start --project=chromium --spec=e2e/specs/settings.spec.ts` currently fails its first guided-setup assertion (`Guided setup path`) even though the unit settings proof is green and the rest of the spec passes, so the investor checklist should treat `/settings` as the one remaining browser-tightening gap instead of claiming clean end-to-end readiness.
24. `IN_REVIEW` Soma can now manage governed deployment-context loading instead of relying on ad hoc memory facts alone: a shared ingestion path stores operator/customer-provided material under `customer_context` vectors and approved company-authored guidance under `company_knowledge` vectors, both persisted as approved artifacts with visibility/sensitivity/trust metadata, and agent runtime context now recalls those governed context hits alongside prior conversation continuity without blurring them into ordinary Soma memory.
25. `IN_REVIEW` the first explicit promotion workflow is now implemented behind governed execution rather than left as a plan-only concept: `promote_deployment_context` reuses the stored confirm-action plan, creates a new `company_knowledge` artifact/vector record from a `customer_context` source, preserves promotion lineage metadata, and requires the same proposal/approval spine as other high-risk learning mutations.
18. `IN_REVIEW` the second cleanup-swarm reference-authority slice is now locally validated: `docs/README.md` has been restructured so V8 navigation leads and V7 material is explicitly grouped as migration/historical input, while `team.worktree-triage` now points contributors at the V8-first authority set (`ARCHITECTURE_LIBRARY_INDEX`, `V8_RUNTIME_CONTRACTS`, `V8_CONFIG_AND_BOOTSTRAP_MODEL`, `V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT`, `V8_DEV_STATE.md`) before narrower review docs.
19. `IN_REVIEW` the first readability/decomposition cleanup slice is now locally validated in `ops/interface.py`: shared one-shot task execution/output handling has been pulled into small internal helpers, the Playwright `e2e` task now isolates command assembly, env construction, and managed-server endpoint reconciliation into named private helpers, and stale task docstrings around `stop` and `check` now describe the real cleanup/server contract instead of older narrower behavior.
20. `IN_REVIEW` the next frontend-contract cleanup slice is now locally in flight: the shipped MCP registry no longer exposes the dead custom-install modal or the stale `installMCPServer` store action wired to the disabled `/api/v1/mcp/install` endpoint, the resources/tools surface is now explicitly library-first for approved installs, and frontend/backend/workflow docs are being synchronized to the governed curated-install contract instead of preserving broken raw-install narratives.
21. `IN_REVIEW` the next store-contract cleanup slice is now locally in flight: `cortexStoreMissionChatSlice.ts` has been decomposed into named route/blocker/success helpers so the Soma chat hot path reads closer to the real execution model, and `docs/architecture/FRONTEND.md` no longer misstates the store as a monolithic `useCortexStore.ts` implementation or flags the wrong file as the current max-lines hotspot.
22. `IN_REVIEW` the next draft-graph cleanup slice is now locally in flight: `cortexStoreMissionDraftSlice.ts` now routes blueprint mutation, graph rebuild, solidification, editor-close state, and full draft reset through named helpers instead of repeating those transitions inline, and direct store tests now cover draft mutation, draft deletion, active-mission graph reconciliation, and mission-delete reset behavior.
23. `IN_REVIEW` the next store-readability cleanup slice is now locally in flight: `cortexStoreResourceCatalogSlice.ts` is being normalized around named helpers for repeated array/status/query patterns, the artifact-governance client contract has been corrected so `updateArtifactStatus` now targets the backend’s `PUT /api/v1/artifacts/{id}/status` route instead of the stale `POST` call shape, and `cortexStoreMissionDraftSlice.ts` now centralizes repeated draft-reset/blueprint-mutation helpers instead of re-implementing the same graph/editor state transitions inline.
24. `IN_REVIEW` the next state-boundary cleanup slice is now locally in flight: `cortexStoreState.ts` no longer presents one undifferentiated store contract block, and is instead grouped into draft/graph, resources, mission chat, governance/ops, automation/runs, and profiles/settings interfaces so future edits can target the right domain boundary without re-scanning the full state surface.
25. `IN_REVIEW` the next initialization-boundary cleanup slice is now locally in flight: `cortexStoreInitialState.ts` now mirrors the grouped store contract through domain-scoped default-state blocks plus a small `StripActions<>` helper, so initial state no longer maintains a second flat contract blob that drifts away from the grouped `CortexState` structure.
26. `IN_REVIEW` the next store-composition cleanup slice is now locally in flight: `useCortexStore.ts` now assembles slice groups through an explicit composition helper and a named mission-chat persistence predicate, so the entrypoint reads like the grouped contract it wires together instead of a bare spread list plus inline subscription condition.
27. `IN_REVIEW` the next slice-typing cleanup is now locally in flight: `cortexStoreSliceTypes.ts` now defines a shared `CortexSlice<>` helper, and the store slice modules now consume that shared return-type contract instead of each repeating their own inline `Pick<CortexState, ...>` declaration header.
28. `IN_REVIEW` the next duplicate/dead-surface cleanup slice is now locally in flight: the unreferenced `interface/components/LogStream.tsx` shim plus its unused `interface/components/hud/LogStream.tsx` implementation, the legacy `interface/components/wiring/CircuitBoard.tsx`, and the orphaned `interface/components/dashboard/TeamRoster.tsx` have been removed, while the historical hardening PRD now explicitly marks those duplicate references as transition-era artifacts instead of pointing contributors at deleted active code.
29. `IN_REVIEW` the next mission-control dead-widget cleanup slice is now locally in flight: the unused legacy `ApprovalDeck`, root `SystemStatus`, old dashboard `ActivityStream`, `AgentPanel`, `CommandDeck`, `SensoryPeriphery`, `TeamList`, and legacy `wiring/WireGraph` surfaces have been removed, and `docs/WORKFLOWS.md` now describes the current Mission Control contract in terms of `MissionControlChat` plus `OpsOverview` instead of the removed dashboard widget set.
30. `IN_REVIEW` the next frontend-reference cleanup slice is now locally in flight: `interface/README.md` has been rewritten around the current Central Soma home, AI Organization workspace, and advanced wiring split instead of the retired “ArchitectChat plus widget grid” product model, and `docs/TESTING.md` now describes the actual current frontend test inventory rather than naming removed Mission Control widgets as active coverage targets.
31. `IN_REVIEW` the next active-doc convergence slice is now locally in flight: `docs/WORKFLOWS.md`, `docs/QA_COUNCIL_CHAT_API.md`, `docs/SWARM_OPERATIONS.md`, and `docs/architecture/OVERVIEW.md` now treat Soma-first dashboard and AI Organization workspace surfaces as the default operator path, while keeping `/wiring` explicitly documented as the advanced graph-and-negotiation surface instead of the old default Mission Control grid model.
32. `IN_REVIEW` the next dead-dashboard-branch cleanup slice is now locally in flight: the unmounted legacy `MissionControl.tsx` container, its private dashboard-only helpers (`TelemetryRow`, `ModeRibbon`, `FocusModeToggle`, `OpsOverview`, `OperationsBoard`), its unused widget registry helper, and the corresponding isolated dashboard tests have been removed so the repo no longer carries a second unshipped dashboard branch beside the current Soma-first product surfaces.
33. `IN_REVIEW` the next cleanup-plan synchronization slice is now locally in flight: the active `V8_2_CROSS_REPO_CLEANUP_AND_RELEASE_STRUCTURE_PLAN.md` no longer points theme/readability work at the deleted `TelemetryRow.tsx` surface and now names `CentralSomaHome.tsx` as the current shipped dashboard polish target instead.
31. `IN_REVIEW` the next legacy signal-shell cleanup slice is now locally validated: the dead `SignalContext` / `ZoneC_Stream` / `ZoneD_Decision` branch and the orphaned `SensorLibrary` component/tests have been removed, `ShellLayout.tsx` now reflects the actual live shell surfaces, and active frontend/backend/testing docs now point to the store-driven `GovernanceModal` plus route-local `NatsWaterfall` / signal-detail inspection path instead of the retired SSE-era shell narrative.
32. `IN_REVIEW` the next active-frontend-metrics cleanup slice is now locally validated: `docs/architecture/FRONTEND.md` and `docs/architecture/OVERVIEW.md` are resynchronized to the current audited interface surface, including the real `page.tsx` route count (`22`), the live `/organizations/[id]` primary workflow route, and the current component inventory (`96` files across `27` folders) instead of older pre-cleanup counts.
33. `COMPLETE` the latest cleanup checkpoints are now packaged in local git history from clean state: `24d6c83` (`Remove dead signal shell branch`), `4174196` (`Sync active frontend architecture metrics`), and `0dc6b24` (`Clarify active overview route coverage`).
34. `COMPLETE` repo hygiene remains closed after those checkpoints: `git status --short` is clean, and the cleanup lane is now carrying its progress through small committed slices instead of rebuilding another mixed local batch.
35. `ACTIVE` the new dev-team execution layer is now formalized in `docs/architecture-library/V8_DEV_TEAM_EXECUTION_AND_UI_ENGAGEMENT_PLAN.md`: the team roster, slice order, UI proof standard, engagement-testing expectation, and clean-commit discipline are now pulled into one canonical execution surface for the true-MVP lane.
36. `IN_REVIEW` execution has started on the first slice from that plan: the default Settings profile no longer carries a fake Notifications toggle in the visible operator path, and focused settings tests are being aligned to the honest surface so default-path controls are either real, inspect-only, or absent.
37. `IN_REVIEW` the first testing/docs execution slice under the new dev-team plan is now locally in flight: `V8_UI_TESTING_AGENTRY_EXECUTION_RUNBOOK.md`, `V8_UI_TEAM_FULL_TEST_SET.md`, and `V8_PARTNER_DEMO_VERIFICATION_CHECKLIST.md` are being corrected so the active MVP lane explicitly covers compose/demo preflight, collaboration/media visible-value proof, assistant-name persistence, and AI Engine / Memory & Continuity inspectability instead of under-testing those expectations.
38. `IN_REVIEW` the next guided-settings cleanup slice is now locally in flight: `/settings` leads with a guided setup path for identity, mission defaults, and people access before any advanced controls, the advanced AI Engine / Connected Tools path stays explicitly intentional instead of reading like the default first step, and route/unit/browser tests are being tightened so that guided-workflow posture becomes part of the contract rather than just implementation copy.
39. `IN_REVIEW` the next organization-workspace guided-flow cleanup slice is now locally in flight: the top of `/organizations/[id]` now distinguishes the first meaningful organization moves from deeper inspect surfaces, the workspace explicitly guides operators into Soma conversation, team design, or setup review before they drop into richer support panels, and the organization page/browser contract is being tightened so that guided-start behavior remains part of the product proof.
40. `IN_REVIEW` the next Soma-chat guided-entry cleanup slice is now locally in flight: the simple-mode empty state no longer stops at passive hint text, starter prompts are becoming clickable guided actions that prefill the Soma lane directly, and the workspace/browser proof is being tightened so guided chat entry remains real interaction rather than decorative copy.
41. `IN_REVIEW` the MCP settings configuration lane now has an explicit owner-group governance contract: curated library installs inspect before install, root/owner current-group config from the MCP page auto-allows without a second approval loop, remote MCP entries return an approval boundary instead of silently prompting, and MCP toolset mutations now echo the same normalized governance posture in their responses.
42. `IN_REVIEW` the MCP governance hardening slice now has standard-library contract proof against the shipped `core/config/mcp-library.yaml`: owned-config inspection is covered for canonical entries like `filesystem` and `github`, local-first toolset links remain stable, and the raw `/api/v1/mcp/install` security block still holds even when a request mirrors a real library entry.
43. `ACTIVE` the ask-class, agent-type, and output-contract assertion lane is now formally opened in `docs/architecture-library/V8_ASK_CLASS_AGENT_TYPE_OUTPUT_CONTRACT_PLAN.md`: local repo truth has been reduced to one highest-value gap, the missing shared machine-readable registry for ask classification and output assertion, and the first execution slice is now defined as a low-risk runtime contract insertion for `direct_answer` vs `governed_mutation` before broader UI/runtime expansion.
44. `IN_REVIEW` the first ask-class runtime contract slice is now landed in code: `core/pkg/protocol/ask_contracts.go` defines the first shared machine-readable registry for `direct_answer` and `governed_mutation`, primary Soma chat plus direct council chat now resolve template/mode selection through that registry, and audit context now records `ask_class` on those bounded answer/proposal paths instead of relying only on duplicated branch defaults.
45. `COMPLETE` current bounded local validation is green from committed state for the non-live gate: `uv run inv core.test`, `uv run inv core.compile`, `uv run inv interface.typecheck`, `uv run inv interface.test`, `uv run inv interface.build`, `cd interface; npm run test`, `$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q`, and `uv run inv interface.e2e --project=chromium --spec=e2e/specs/v8-ui-testing-agentry.spec.ts` all pass. The live governed browser gate was not rerun in this slice because the local stack is intentionally down (`Docker`, `PostgreSQL`, `NATS`, `Core API`, and `Frontend` all offline).
46. `COMPLETE` the supported home-runtime compose path now matches the current database compatibility contract: `compose.up` and `compose.migrate` no longer force strict replay of every migration against an already-bootstrapped compose `cortex` schema, and instead skip replay once the required current-runtime tables/columns exist so live bring-up does not fail on stale-but-compatible compose volumes. Focused proof is green from the actual compose stack with `python -m py_compile ops/compose.py`, `$env:PYTHONPATH='.'; uv run pytest tests/test_compose_tasks.py -q`, `$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q`, `uv run inv compose.up`, `uv run inv compose.status`, `uv run inv compose.health`, and `uv run inv compose.migrate`.
47. `COMPLETE` the compose-hosted live governed browser lane is green again: `interface/e2e/specs/soma-governance-live.spec.ts` now resolves `MYCELIS_BACKEND_WORKSPACE_ROOT` relative to the repo root instead of the `interface/` working directory, so the documented compose workspace path works for durable file-side-effect assertions, and `uv run inv interface.e2e --live-backend --server-mode=start --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts` passes end to end against the healthy compose stack.
48. `COMPLETE` the next ask-class assertion slice is now wired across runtime and interface: the shared Go registry now covers `governed_artifact` and `specialist_consultation`, chat payloads now carry `ask_class` through the CTS/store contract, and Mission Control chat now marks artifact-bearing answers as `Artifact result` plus consulted answers as `Specialist support` so output class is no longer only implied by secondary details. Proof is green with `cd core; go test ./pkg/protocol ./internal/server -count=1`, `uv run inv core.test`, `cd interface; npx vitest run __tests__/store/useCortexStore.test.ts __tests__/dashboard/MissionControlChat.test.tsx --reporter=dot`, `cd interface; npx tsc --noEmit`, `uv run inv interface.test`, and `$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q`.
49. `COMPLETE` browser-level ask-class proof is now green: the stable mocked browser suite now asserts visible `Artifact result` and `Specialist support` cues, while the live governed compose-backed suite now asserts returned `ask_class` values for direct-answer and governed-mutation responses so browser evidence covers output-class assertion, not only route and terminal mode. Proof is green with `uv run inv interface.e2e --project=chromium --spec=e2e/specs/v8-ui-testing-agentry.spec.ts`, `uv run inv compose.up --build`, and `$env:MYCELIS_BACKEND_WORKSPACE_ROOT='workspace/docker-compose/data/workspace'; uv run inv interface.e2e --live-backend --server-mode=start --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts`.
50. `COMPLETE` ask-class browser proof now extends into the guided organization workspace itself: `interface/e2e/specs/v8-organization-entry.spec.ts` now proves that `/organizations/[id]` keeps the organization frame intact while artifact-bearing asks show `Artifact result` with the inline artifact surface and specialist-shaped asks show `Specialist support` plus visible consultation trace. Proof is green with `uv run inv interface.e2e --project=chromium --spec=e2e/specs/v8-organization-entry.spec.ts`, `cd interface; npx tsc --noEmit`, and `$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q`.
51. `COMPLETE` governed-artifact answers now lead with an operator-facing returned-output summary in the default chat surface: `MissionControlChat.tsx` derives a concise `Soma prepared ... for review` summary from returned artifact metadata, component proof asserts that summary for single- and multi-artifact answers, and both the stable workspace and guided organization browser suites now verify the summary alongside the existing `Artifact result` cue. Proof is green with `cd interface; npx vitest run __tests__/dashboard/MissionControlChat.test.tsx --reporter=dot`, `cd interface; npx tsc --noEmit`, `uv run inv interface.e2e --project=chromium --spec=e2e/specs/v8-ui-testing-agentry.spec.ts`, `uv run inv interface.e2e --project=chromium --spec=e2e/specs/v8-organization-entry.spec.ts`, and `$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q`.
52. `COMPLETE` specialist-consultation answers now lead with a softer operator-facing specialist-context summary in the default chat surface: `MissionControlChat.tsx` derives a concise `Soma checked with ... while shaping this answer` summary from consultation metadata, component proof asserts that summary for the bounded specialist path, and both the stable workspace and guided organization browser suites now verify the summary alongside the existing `Specialist support` cue and consultation trace. Proof is green with `cd interface; npx vitest run __tests__/dashboard/MissionControlChat.test.tsx --reporter=dot`, `cd interface; npx tsc --noEmit`, `uv run inv interface.e2e --project=chromium --spec=e2e/specs/v8-ui-testing-agentry.spec.ts`, `uv run inv interface.e2e --project=chromium --spec=e2e/specs/v8-organization-entry.spec.ts`, and `$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q`.
53. `ACTIVE` a canonical user-workflow execution lane is now opened in `docs/architecture-library/V8_USER_WORKFLOW_EXECUTION_AND_VERIFICATION_PLAN.md`: the team structure, execution waves, user-shaped delivery rule, MCP/API usability rule, and workflow-complete verification standard now explicitly state that Mycelis governs engine/execution boundaries rather than over-controlling ordinary content intent, team shape, or standards-compliant framework utilization.
54. `COMPLETE` the first workflow-complete hardening slice is now landed on the settings path: browser coverage for `/settings` no longer stops at section switching, and now proves that assistant identity and theme save through the real user-settings contract, persist after reload, and keep the guided setup framing intact while the workflow runs. Supporting unit coverage now also asserts assistant-name save from the guided profile path, and the workflow verification plan now names settings persistence as a first-class release workflow instead of leaving it implied. Proof is green with `cd interface; npx vitest run __tests__/pages/SettingsPage.test.tsx --reporter=dot`, `cd interface; npx tsc --noEmit`, `uv run inv interface.e2e --project=chromium --spec=e2e/specs/settings.spec.ts`, and `$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q`.
55. `COMPLETE` People & Access now has an explicit layered-permissions contract instead of always exposing raw user CRUD: `/api/v1/user/me` and `/api/v1/user/settings` now carry `access_management_tier`, the settings UI keeps organization roles and collaboration groups in the base-release layer, enterprise user-directory management is exposed only when that layer is enabled, backend `admin` identity now maps truthfully to owner access in the interface, and the workflow docs/tests now prove base-release, enterprise-owner, and enterprise-non-owner expectations. Proof is green with `cd core; go test ./internal/server -run 'TestHandle(Me|UserSettings)' -count=1`, `cd interface; npx vitest run __tests__/settings/UsersPage.test.tsx __tests__/pages/SettingsPage.test.tsx --reporter=dot`, `cd interface; npx tsc --noEmit`, `uv run inv interface.e2e --project=chromium --spec=e2e/specs/settings.spec.ts`, and `$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q`.
56. `ACTIVE` the full user-intent workflow review and instantiation lane is now opened as a first-class release-shaping contract: `docs/architecture-library/V8_USER_INTENT_WORKFLOW_REVIEW_AND_INSTANTIATION_PLAN.md` reviews the V8.1 release against the V8.2 target, defines canonical user-intent classes, and makes the initial-release boundary explicit between direct handling, native Mycelis team instantiation, and external workflow-contract instantiation such as `n8n`. The active workflow, UI test, runbook, architecture-index, docs-manifest, and dev-team execution surfaces now all require proof that native team-instantiated target output and external workflow-contract posture stay distinct and legible. Proof is green with `cd interface; npx tsc --noEmit` and `$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q`.
57. `COMPLETE` the first bounded native-team vs external-contract workflow proof is now green in the guided organization team-design lane: `TeamLeadInteractionPanel` preserves intent context in the request body, no longer misclassifies every prompt containing `team` as a setup review, and now renders an explicit `Execution path` card that keeps native Mycelis team output separate from external workflow-contract targeting such as `n8n`. The release docs now state the truthful boundary that current proof begins in `/organizations/[id]` team-design guidance before deeper runtime activation or runnable external invocation exists. Proof is green with `cd core; go test ./internal/server -run TestHandleTeamLead -count=1`, `cd interface; npx vitest run __tests__/organizations/TeamLeadInteractionPanel.test.tsx __tests__/pages/OrganizationPage.test.tsx --reporter=dot`, `cd interface; npx tsc --noEmit`, and `uv run inv interface.e2e --project=chromium --spec=e2e/specs/v8-organization-entry.spec.ts`.
58. `COMPLETE` README and the related guidance surfaces are now explicitly cleaned up by audience and tied back to one canonical full-testing runbook: `README.md` now separates user guidance from agent guidance and points directly to `docs/architecture-library/V8_FULL_TESTING_ACTION_PLAN.md`, `docs/README.md` now cleanly splits user docs, agent/developer docs, testing/release docs, and migration inputs, `docs/TESTING.md` now points to the ordered full-testing runbook, and the architecture index plus in-app docs manifest now expose the same plan as a first-class canonical testing surface. The full release-style testing pass from that plan is green: `uv run inv ci.baseline`, `uv run inv ci.service-check --live-backend`, `uv run inv compose.up`, `uv run inv compose.status`, `uv run inv compose.health`, `$env:MYCELIS_BACKEND_WORKSPACE_ROOT='workspace/docker-compose/data/workspace'; uv run inv interface.e2e --live-backend --server-mode=start --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts`, and `$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q` all pass. The only noted residual warning remains the already-accepted Vitest `--localstorage-file` worker warning during broad baseline runs.

## Current Review (2026-03-29)

Review summary:
1. `ACTIVE` the universal Soma correction is now canonical: `docs/architecture-library/V8_UNIVERSAL_SOMA_AND_CONTEXT_MODEL_PRD.md` defines Soma and Council as universal entities, while AI Organizations, deployments, teams, logs, memory, and runs remain governed working contexts rather than separate Soma identities.
2. `ACTIVE` the first actuation slice under that PRD is now underway and materially visible in product: the settings contract is normalized so current product settings can persist and reload through a real canonical API, and the bounded V8 UI contract now explicitly points to the universal-Soma PRD as the next product identity correction.
3. `ACTIVE` the dashboard/home contract has started evolving away from a pure organization-launch surface: `/dashboard` now presents a Central Soma home layer above the existing AI Organization creation/re-entry flow so the product teaches one persistent Soma while still preserving the current scoped organization workflow underneath.
4. `IN_REVIEW` the deployed compose runtime is now green for platform health and stable browser proof: `compose.up --build`, `compose.status`, `compose.health`, `interface.check`, and the compose-hosted stable browser suite pass with Central Soma home, AI Organization entry, and the broader UI agentry contract intact.
5. `BLOCKED` the deployed compose runtime still has one live-browser product defect: `soma-governance-live.spec.ts` Scenario A returns `500 Internal Server Error` for a fresh-organization direct answer even though the mutation/proposal/cancel/confirm scenarios pass in the same healthy compose stack.
6. `ACTIVE` the new content-generation and collaboration strike team is now engaged as a first-class delivery lane: the product problem is formally defined as “content requested but not visibly delivered,” and the team now has canonical rules for inline `answer` behavior, governed artifact generation, policy-configurable specialist/model collaboration, and readable result return through Soma.
7. `ACTIVE` the Product Manager lane now explicitly rejects blanket approval posture for all MCP/model collaboration: low-risk collaboration may stay fluid when policy allows it, while higher-risk mutation, external, or costly paths remain reviewable through approval posture.
8. `ACTIVE` the operator-settings closeout has started with a real MVP fix: theme selection is now wired end to end from persisted user settings through frontend state and DOM theme token application, replacing a previously dead Appearance control.
9. `NEXT` the first execution slice for the content/collaboration lane is truth mapping: classify current request types into inline content, governed artifact generation, and collaboration flows, then identify where mode selection or result surfacing currently fails the operator.
10. `NEXT` the UI testing agentry and product-delivery teams must now verify not only proposal safety but also visible value delivery: content/media/file requests must either return readable inline output or a clearly referenced artifact/result.
11. `NEXT` docs, testing, and UI wording must stop implying that all external/model collaboration is approval-gated by default; the correct standard is policy-configurable governance tied to risk, mutation impact, cost, and integration trust.
12. `ACTIVE` the demo-product strike team is now engaged beyond coordination: the canonical execution brief now defines the current platform review, default-vs-advanced surface truth, feature-preservation rules, team-by-team assignments, cross-team dependency order, and the concrete acceptance outputs required next.
13. `COMPLETE` the default-vs-advanced product split is now evidenced against the current UI surface: primary navigation keeps the product story centered on AI Organization entry, the current organization, and Docs, while `Resources`, `Memory`, and `System` remain intentionally advanced-gated rather than removed.
14. `COMPLETE` the first wording-drift audit now exists in canonical form: README, landing, organization setup, organization home, and user-doc surfaces were reviewed against the current product story, with `Memory & Continuity` vs older `Learning` terminology called out as the main cleanup target and internal-vs-user-facing rename boundaries made explicit.
15. `ACTIVE` product-language cleanup is now materially underway across both product and documentation: README, the V8 UI/operator contract, the workflow verification plan, organization-home fallback copy, and targeted verification assets now converge on `Memory & Continuity` and retained-knowledge language, while architecture-level references to `Learning Loops` remain intentionally separate as deeper system terminology.
16. `COMPLETE` the first feature-preservation artifact now exists in canonical form: the retained-home map proves where advanced power currently lives across navigation, advanced routes, governed chat, team-design mode, and runtime/operator depth so demo-product simplification does not become accidental feature regression.
17. `COMPLETE` the Demo Scenario Team and UI Testing Agentry Team now have canonical engagement outputs: a partner demo script and a demo-specific verification checklist now define the exact product story, governed-action proof, continuity proof, optional advanced reveal, and fallback expectations for high-stakes partner review.
18. `COMPLETE` the managed organization-entry browser blocker is now resolved: the issue was not product source drift but task-contract drift in `interface.e2e`, where a stale listener could satisfy the run and `next start` could serve an older build. The managed browser task now refreshes the production bundle in start mode, fails closed if it cannot own a clean managed UI server, and the focused organization-entry browser proof is green against current product language.
19. `COMPLETE` targeted product-language convergence is now green across the current AI Organization entry and home path: `Memory & Continuity`, retained-knowledge wording, current Soma quick-action phrasing, and non-technical empty-state language all match between the current UI, the focused page suite, the managed browser proof, and the active docs/test gates.
20. `COMPLETE` the local runtime contract now includes a supported home Docker Compose path: Core now honors `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, and `DB_NAME` for real runtime DB wiring, compose automation manages startup order and migrations, and the docs now distinguish the Kind/Kubernetes path from the single-host compose path instead of treating compose as legacy.
21. `NEXT` Phase 1 product framing, universal-Soma home evolution, content-value delivery, and remaining settings closeout should now proceed together under the true-MVP finish plan so Mycelis improves as a product instead of only as a technically valid release candidate.

## Current Review (2026-03-28)

Review summary:
1. `COMPLETE` a fresh-cluster release proof is now green: `k8s.reset`, `lifecycle.up --frontend`, `lifecycle.status`, `lifecycle.health`, `db.migrate`, and the live governed Soma browser spec now pass together on the current local stack.
2. `COMPLETE` the governed live browser lane is no longer blocked by the previous chat/proposal and confirm/write proof failures; informational summary prompts now stay answer-first, and the live file-side-effect proof supports explicit backend workspace roots when the spec checkout differs from the running Core checkout.
3. `IN_REVIEW` release-platform documentation is being synchronized across security/governance, monitoring/ops, and debug/live-browser proof through a dedicated shared review doc plus updates to the state file, architecture index, and in-app docs manifest.
4. `IN_REVIEW` monitoring/ops posture is materially stronger: `db.migrate` now behaves as a forward-bootstrap helper once the schema is already initialized, while fresh-cluster proof confirms that `k8s.reset` plus sequential lifecycle bring-up is the trustworthy recovery path.
5. `ACTIVE` the local worktree remains mixed and still needs commit-boundary packaging across the live-governed-chat, db/bootstrap cleanup, docs/state sync, and earlier cleanup lanes before release promotion.
6. `ACTIVE` the memory continuity review has now started: pgvector-backed semantic memory, team-scoped recall, temporary planning continuity, and trace-vs-memory boundaries are under cross-team review before the next runtime slice lands.
7. `IN_REVIEW` the first memory-continuity runtime slice is now locally validated: durable memory recall can be scoped by team/agent context, `remember` now stores scoped durable metadata, automatic conversation checkpointing now writes restart-safe temporary planning continuity instead of silently promoting exploratory chat into pgvector, and operator/docs surfaces now describe durable vs temporary vs trace memory more explicitly.
8. `ACTIVE` demo-product delivery is now being managed as its own strike-team lane: the immediate target is partner/funder legibility in minutes, the default experience must read as product rather than nerd tool, and advanced depth must be preserved through advanced surfaces/docs instead of removed.
9. `COMPLETE` Phase 0 team engagement outputs now exist in canonical form: the demo-product execution brief defines the default-vs-advanced surface inventory, the feature-preservation map, the golden-path partner demo sequence, and the UI-testing checklist the teams should now execute against.

## Current Review (2026-03-27)

Review summary:
1. `COMPLETE` the pushed V8.2 branch history still carries the accepted RC-delivery spine through Soma-first UX, execution reliability, governance/approval foundations, proposal-integrity closure, and provider-default cleanup (`247fe16`, `bfb3b80`, `212d5d2`, `94b531c`, `2a3c5f8`).
2. `IN_REVIEW` the build/task/workflow cleanup lane is now locally validated: `core.compile` and `interface.typecheck` are first-class invoke tasks, CI workflows now watch `uv.lock` and frontend config changes via root `.npmrc`, and canonical operator docs were updated to the current invoke contract.
3. `IN_REVIEW` operator-truthfulness cleanup is now locally validated in the ops layer: `lifecycle.status` no longer reports Docker healthy on failed probes, `ci.service-check --live-backend` now points at the current governed Soma live-browser proof instead of the stale workspace-only spec, `team.worktree-triage` now covers the full task-contract doc set plus stronger evidence commands, and `k8s.reset` no longer emits bare-`inv` guidance.
4. `COMPLETE` local validation for that cleanup lane is green: Python task-module compile checks, focused pytest task/doc/workflow suites, `uv run inv interface.typecheck`, `uv run inv interface.test`, `uv run inv interface.build`, `uv run inv test.all`, and `uv run inv ci.build`.
5. `IN_REVIEW` the live governed-chat browser lane is now green in fresh-cluster validation: direct informational chat returns `answer`, governed mutations stay proposal-first, confirm-action proof passes end-to-end, and the live spec supports explicit backend-workspace binding for cross-worktree validation.
6. `ACTIVE` the local worktree remains mixed after the cleanup pass: in addition to the ops/workflow/doc hardening lane, there are still uncommitted UI/testing stabilization edits, governance/monitoring release-review docs, and related cleanup that need packaging into intentional commit boundaries before release promotion.
7. `NEXT` low-confidence cleanup remains intentionally deferred until proven safe: do not delete `_matches_compiled_go_binary_path` or similar low-usage helpers until a focused usage review confirms no hidden dependency path remains.

Retained completed checkpoint:
- `COMPLETE` run the planning-integration validation pass so README, the architecture-library index, docs manifests, and doc-tests all confirm the new V7-to-V8 bootstrap migration contract.

## V8 Delivery Program Snapshot

```text
V8-0  Migration baseline and architecture alignment                 [ACTIVE]
V8-1  Canonical contract definition                                [COMPLETE]
V8-2  Config and bootstrap model planning                          [COMPLETE]
V8-3  Backend primitive refactor                                   [REQUIRED]
V8-4  Frontend/operator refactor                                   [REQUIRED]
V8-5  Documentation and naming migration                           [ACTIVE]
V8-6  Verification and release hardening                           [REQUIRED]
```

## Current V8 Grading Baseline

### 1. Architecture alignment and migration inventory

Status:
1. `ACTIVE` V8 alignment is now carried by the README directive, the V8 runtime contracts, and the V8 bootstrap-planning surface instead of a separate tracker artifact.
2. `COMPLETE` current fixed-organism assumptions vs V8 target model have been mapped at the planning level.
3. `COMPLETE` backend/frontend/docs hotspots have been identified for the first migration wave.
4. `ACTIVE` carry those alignment conclusions into explicit V7-bootstrap migration rules and de-hardcoding work now that the canonical plan exists.

Primary references:
- `README.md`
- `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`
- `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`

Key grading conclusion:
- Mycelis already has the execution substrate needed for V8, but still exposes one hardcoded Soma/Council organizational expression as if it were the platform itself.

### 2. Provider and service architecture

Status:
1. `ACTIVE` V8 explicitly requires support for local, hosted, and hybrid model-provider infrastructure.
2. `COMPLETE` provider abstraction is confirmed as an existing core subsystem, not a plugin afterthought.
3. `ACTIVE` instantiated-organization provider policy now drives runtime routing with inherited organization, kernel/council-role, team, and agent scope resolution.
4. `NEXT` extend provider-policy grading beyond the standing-team bridge so future instantiated organizations can declare richer approval/data-boundary posture without fallback-era assumptions.

Required V8 provider support:
- local/open-source: Ollama, LM Studio, vLLM, custom OpenAI-compatible endpoints
- hosted/commercial: OpenAI, Anthropic, Gemini, other OpenAI-compatible providers

Required routing posture:
- role-level routing
- council-level routing
- inception-level policy boundaries
- local-first and hybrid deployment support
- explicit data-boundary and approval posture

### 3. State and documentation migration

Status:
1. `ACTIVE` canonical grading is moving from `V7_DEV_STATE.md` to `V8_DEV_STATE.md`.
2. `ACTIVE` active non-archive docs are being updated to reference the V8 state file.
3. `ACTIVE` top documentation entrypoints now link the detailed framework-memory surfaces needed for restart/onboarding (`README.md`, `docs/README.md`, `docs/archive/README.md`).
4. `COMPLETE` clearly stale dated architecture-review addenda were removed from `docs/architecture/`.
5. `REQUIRED` V7 architecture-library documents remain authoritative migration inputs until V8 replacements exist.

State-file rules:
1. all new status transitions, blocker updates, and gate evidence belong in `V8_DEV_STATE.md`
2. `V7_DEV_STATE.md` remains readable for historical migration evidence, but should not be the primary grading target for new work
3. active docs should describe V7 documents as migration inputs and V8 state as the live delivery scoreboard

### 4. Runtime and UI migration posture

Status:
1. `COMPLETE` canonical V8 contracts for `Inception`, `Soma Kernel`, `Central Council`, provider-policy scope, and identity/continuity state now exist.
2. `COMPLETE` the config/bootstrap model now documents configuration sources, template entry points, scope inheritance, precedence, and V7-to-V8 migration rules.
3. `REQUIRED` replace standing-team bootstrap assumptions with configurable organization-resolution contracts.
4. `REQUIRED` keep Workspace simple and Soma-first while making the operator model kernel-aware instead of fixed-identity-bound.
5. `REQUIRED` continue the full UI retheme and density-reduction effort under V8 delivery targets, not legacy V7 framing.

### 5. Testing and gate posture

Status:
1. `REQUIRED` every V8 migration slice must still carry backend, UI, and docs proof in the same delivery window.
2. `ACTIVE` architecture and state-file changes must be validated by doc-surface checks and in-app docs alignment where applicable.
3. `ACTIVE` clean-run discipline is now part of the delivery contract for runtime and integration-style checks: stop prior services, verify ports/processes are clear, sweep stray compiled Go binaries from prior runs, start only the minimum required stack, run the check, and tear the stack back down unless explicitly needed.
4. `REQUIRED` runtime/API changes must include explicit UI review/test targets and focused live-flow evidence when user-visible behavior is touched.

### 6. Implementation-slice cleanup and convergence rule

Status:
1. `REQUIRED` every implementation slice that edits runtime, config, UI, or supporting logic must include a touched-file cleanup/convergence pass.
2. `REQUIRED` chunk completion for code-editing slices must include a focused review-team sweep covering contract alignment, dead-code scan, integration continuity, and test alignment.
3. `REQUIRED` deferred cleanup that cannot be completed safely in the same slice must be reported explicitly in the chunk result.

Implementation rule:
1. review each changed file for stale fixed-Soma/fixed-Council assumptions, dead branches, outdated comments, duplicated helpers, obsolete config paths, and naming drift
2. remove safe-to-remove stale code in the touched files as part of the same slice
3. keep the review scope to touched files and directly adjacent files needed for correctness
4. commit cleanup changes in the same chunk when they are part of the same logical convergence work

## Active V8 Queue

```text
Task 001  V8 alignment and migration inventory                      [IN_REVIEW]
Task 002  Inception / kernel / council contract definition          [COMPLETE]
Task 003  Provider-policy scope contract                            [COMPLETE]
Task 004  Config and bootstrap model planning                       [COMPLETE]
Task 005  Standing-team bootstrap de-hardcoding plan                [ACTIVE]
Task 006  Workspace/UI kernel-aware refactor plan                   [ACTIVE]
Task 007  V8 docs/state migration and grading discipline            [ACTIVE]
Task 008  Planning-integration validation pass                      [COMPLETE]
Task 009  Next-execution/governance guidance migration              [NEXT]
```

## Current Checkpoint (2026-03-21)

Restart handoff:
- latest local checkpoint: `12bda0d` `V8.1 architecture: add learning loops and semantic continuity model`
- restart focus: bundle-defined Loop Profile truth, semantic continuity and memory-promotion implementation planning, and the first bounded Learning Loop slice
- preserved V8.1 truth at restart: read-only Automations visibility is complete; Learning Loops, semantic continuity, and Procedure / Skill Sets are now canonical architecture contract layers
- keep unrelated dirty work out of the next slice unless explicitly pulled in (`README.md`, `docs/**`, `ops/misc.py`, `tests/test_misc_tasks.py`)

Delivery updates in this checkpoint:
1. `COMPLETE` reviewed the new root `README.md` and adopted V8 as the active development/grading target.
2. `COMPLETE` established `V8_DEV_STATE.md` as the canonical state file for new work.
3. `ACTIVE` migrated active non-archive documentation references from `V7_DEV_STATE.md` to `V8_DEV_STATE.md` where those docs define current execution discipline.
4. `COMPLETE` preserved V7 architecture-library documents as migration inputs rather than rewriting them into premature V8 specifications.
5. `COMPLETE` documented how V7 bootstrap sources, implicit behaviors, fixed Soma/Council posture, and runtime-state coupling roll forward into the explicit V8 configuration/bootstrap model.
6. `COMPLETE` validated cross-doc planning consistency so README, the architecture-library index, runtime contracts, bootstrap model, docs manifest, and tests all reference the canonical V7->V8 migration contract.
7. `ACTIVE` landed the next Task 005 implementation cut: startup now instantiates runtime organization truth directly from self-contained bundle data in `core/config/templates/*.yaml` and fails closed unless a valid bootstrap bundle is available.
8. `COMPLETE` promoted clean-run testing discipline into the active testing/operations contract so runtime and integration checks must not stack on unknown local processes.
9. `COMPLETE` extended the clean-run contract so compiled Go services from prior `go build`, `go run`, or manual binary launches are explicitly detected and terminated before the next runtime or integration check.
10. `BLOCKED` the latest lifecycle/doc hardening commits are local-only until GitHub SSH-agent/key access is restored; branch publication is currently blocked by `Permission denied (publickey)` during `git push`.
11. `COMPLETE` refreshed top documentation entrypoints so restart/onboarding flow now points directly at the detailed framework-memory surfaces (`README.md`, `docs/README.md`, `docs/archive/README.md`).
12. `COMPLETE` removed clearly stale dated review docs from `docs/architecture/` to reduce documentation clutter without disturbing active migration inputs.
13. `COMPLETE` doc-surface validation is green again after the documentation cleanup pass.
14. `COMPLETE` wired provider-policy inheritance from the instantiated runtime organization into live Soma routing so organization defaults, kernel/council role defaults, team defaults, and agent overrides now resolve through one policy-bounded path.
15. `COMPLETE` retired the `MYCELIS_TEAM_PROVIDER_MAP` / `MYCELIS_AGENT_PROVIDER_MAP` runtime compatibility path so provider routing now resolves only from the instantiated organization policy and never from startup env maps.
16. `COMPLETE` retired the remaining no-bundle bootstrap fallback: normal startup now has one truth path only, `template -> instantiation -> runtime organization`, while mirrored `config/teams/*.yaml` packaging remains compatibility input rather than startup truth.
17. `COMPLETE` operator startup requirements are now explicit: Core requires at least one valid bootstrap bundle under `config/templates/`, and `MYCELIS_BOOTSTRAP_TEMPLATE_ID` must be set whenever more than one bundle is present.
18. `COMPLETE` defined a dedicated V8 UI/API/operator PRD so first-run, AI Organization creation, template-vs-empty start, organization home, Team Lead-first workspace behavior, role visibility, advanced-mode boundaries, and screen-to-API mapping now have one canonical contract.
19. `COMPLETE` exposed the new V8 UI/API/operator contract through the architecture-library index, the in-app docs manifest, and doc-test enforcement so future UI slices do not drift back toward generic-chat UX.
20. `COMPLETE` implemented the first bounded V8 UI slice: `/dashboard` now opens with `Create AI Organization`, offers `Start from template` vs `Start empty`, uses user-facing AI Organization terms, and routes successful creation into a dedicated AI Organization home instead of generic workspace chat.
21. `COMPLETE` added the minimal backend contract for the new flow: starter template listing from bundle-backed organization templates, AI Organization creation requests, recent-organization summaries, and organization-home loading for the success path.
22. `COMPLETE` hardened the AI Organization entry UX before live GUI sign-off: starter-template and recent-organization failures are now decoupled, retry/recovery paths preserve any still-valid actions, architecture/dev wording is removed from operator-facing copy, and the organization home makes the Team Lead more concrete.
23. `COMPLETE` added bounded browser coverage for the AI Organization entry flow so the dominant create path, template path, empty-start path, partial-failure recovery, and forbidden-copy checks now run in Playwright before wider UI slices proceed.
24. `COMPLETE` turned the post-create organization page into the first Team Lead-first workspace shell: organization context stays visible, the Team Lead now has a concrete identity/presence block, primary next actions drive the active workspace focus, and advanced controls remain summarized instead of becoming the default surface.
25. `COMPLETE` introduced the first guided Team Lead interaction workflow: the organization workspace now offers structured Team Lead starting actions, sends a minimal workspace action request to the backend, and renders a shaped guidance response without expanding into generic chat or advisor orchestration.
26. `COMPLETE` hardened the guided Team Lead workflow for bounded production use: action failures now surface clear retry guidance without dropping the AI Organization frame, partial guidance payloads normalize into readable Team Lead sections, duplicate submits stay disabled while loading, and browser coverage now exercises guided-action failure recovery.
27. `COMPLETE` added inspect-only Advisor and Department visibility to the Team Lead workspace so the AI Organization structure now feels visible and understandable without displacing the Team Lead as the primary operating counterpart.
28. `COMPLETE` added inspect-only AI Engine Settings and Memory & Continuity surfaces to the Team Lead workspace so operators can understand the current organization posture without opening advanced controls or seeing architecture jargon.
29. `COMPLETE` added inspect-only Advisor and Department detail views to the Team Lead workspace and wired both Team Lead and support-column actions into those focused views without displacing the Team Lead as the primary counterpart.
30. `COMPLETE` added an inspect-only AI Engine Settings detail view with scoped model-assignment visibility so operators can inspect organization-wide defaults, team defaults, and specific-role override status without opening advanced editing.
31. `COMPLETE` introduced the first bounded AI Engine edit capability at the organization level: the Team Lead workspace now offers a guided `Change AI Engine` flow with curated operator-facing options, backend validation of allowed values only, immediate workspace refresh, and bounded retry handling without exposing raw provider/model identifiers or advanced configuration panels.
32. `COMPLETE` added controlled Department-level AI Engine override with inheritance clarity: Department details now show whether each Team is following the organization default or using an override, operators can apply a guided Department-specific engine choice or revert cleanly, and organization-level changes continue to flow only to Departments that still inherit the default.
33. `COMPLETE` added the first bounded organization-level Response Contract milestone: the workspace now exposes a safe default `Response Style` contract with curated tone/structure/detail profiles, backend validation of allowed values only, immediate summary refresh after updates, and bounded retry handling without exposing raw prompt text or advanced policy controls.
34. `COMPLETE` added inspect-only Agent Type Profiles to Department details so the Team workspace now exposes the missing inheritance layer between Team defaults and individual agent instances, including clear user-facing visibility into type-level AI Engine bindings and Response Style bindings without opening agent-instance editing.
35. `COMPLETE` added controlled Agent Type AI Engine binding in Department details so operators can pin a curated AI Engine to a role type, cleanly return that role type to the Team default, and preserve inheritance clarity between Team-level engine choices and type-specific specialist behavior.
36. `COMPLETE` added controlled Agent Type Response Style binding in Department details so operators can pin a curated Response Style to a role type, return that role type to the Organization / Team default, and preserve inheritance clarity between the organization-wide Response Style and type-specific specialist output behavior.
37. `COMPLETE` aligned the root page with the V8.1 living AI Organization model so product-facing entry messaging now centers AI Organizations, Team Lead-guided operation, continuous reviews/checks/updates, and the post-creation workspace instead of legacy swarm-console framing.
38. `COMPLETE` added a read-only `What the Organization is Learning` workspace surface so operators can now see safe, human-readable learning highlights with source, recency, and simple strength labels without exposing raw memory internals or any mutation controls.
39. `COMPLETE` added managed cache hygiene for local delivery work so Invoke-driven uv/pip/npm/go flows now centralize repo caches under `workspace/tool-cache`, project build artifacts have an explicit cleanup path, and Windows operators have a first-class user-policy command to keep heavy per-user caches off `C:`.
40. `COMPLETE` audited the full MVP route/tab surface against the V8.1 Team Lead-first workflow, removed the dead `Draft Blueprints` placeholder tab, tightened the default rail to core operator paths, moved `Resources`, `Memory`, and `System` behind explicit Advanced mode, and renamed stale route labels like `Neural Wiring`, `Brains`, and `Cognitive Matrix` into operator-facing release wording.
41. `COMPLETE` refreshed the frontend route contract and user-facing docs so surviving MVP surfaces, advanced boundaries, and compatibility redirects now align in code, tests, and internal docs.
42. `COMPLETE` declared `v8-2.md` as the canonical full production architecture and full actuation target, while keeping V8.1 explicit as the current bounded release across README, the architecture index, docs manifest, and doc-test enforcement.
43. `COMPLETE` enforced documentation synchronization as part of the execution contract so slices are now incomplete until tests pass, documentation is updated where meaning changed, and README, V8.1, V8.2, state, and doc tests stay aligned.
44. `COMPLETE` added deployment-friendly env override support for cognitive config so provider endpoints, model ids, provider enablement, profile routing, and media config can now be stamped by automation tools without reviving the retired team/agent provider env-map path.
45. `COMPLETE` hardened the env override architecture boundary in README, operations docs, local workflow docs, and doc tests so deployment-time env configuration stays separate from runtime organizational truth and routing remains `Bundle -> Instantiated Organization -> Inheritance -> Routing`.
46. `COMPLETE` added explicit deployment guidance for Windows x86_64, Linux x86_64, Linux arm64, and mixed-architecture host layouts so operators have one documented contract for how deployment-time configuration should vary without forking runtime organizational truth.
47. `COMPLETE` hardened the V8.1 Soma-primary UX for release readiness so first-run guidance is always visible, default workspace panels now have intentional empty/loading/error states with retry paths, and the operator-facing vocabulary now stays aligned around Automations, Memory & Continuity, AI Engine, Response Style, Advisors, and Departments.
48. `COMPLETE` defined the advanced architecture/runtime surface contract before further UI work so default operator behavior, advanced inheritance/config inspection, source-of-truth layering, deployment/env influence, and file/env-only boundaries now have one explicit doc-backed contract.
49. `NEXT` implement the advanced architecture/runtime UI only through a separate non-default surface that makes inheritance and config origin legible without undermining the Soma-primary MVP release posture.
50. `COMPLETE` restored Soma as the primary interface and orchestrator for the default workspace while keeping Team Leads visible as subordinate operational leaders, with UI and docs now aligned on `User -> Soma -> Routing / Council -> Team Leads -> Departments / Specialist Roles / Automations -> Reviews / Learning / Activity -> Soma`.
51. `COMPLETE` converged canonical documentation ownership so README, V8.1, V8.2, V8 UI/API, and `V8_DEV_STATE.md` now carry distinct non-overlapping truth roles without leaving loose root-level draft architecture files active.
52. `COMPLETE` repo hygiene is now release-readiness oriented for the canonical doc surface: the superseded root `v8-1.md` draft was archived under `docs/archive/drafts/`, stale root-level draft truth was removed, and doc tests now guard against canonical-surface drift and loose duplicate ownership.
53. `COMPLETE` fixed the last MVP-blocking UX gaps in the default organization flow: recent AI Organizations are now reopenable, persistent org navigation is available after an organization is opened, post-create routing lands directly in the workspace, Soma has an unmistakable primary prompt entry, action feedback always surfaces visible loading/result/error state, and the degraded-mode banner now makes it clear that core functionality remains available.
54. `COMPLETE` stabilized the MVP release-candidate test posture: the broader interface suite is green again, the stale browser navigation expectation was updated so route-highlighting validation resets cleanly between navigations, and the managed E2E gate now passes without relying on a previously reused stale frontend server.

Evidence:
1. README directive review completed against `README.md`
2. runtime-contract alignment confirmed against `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`
3. bootstrap-planning alignment confirmed against `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`
4. active documentation references updated to point at `V8_DEV_STATE.md`
5. V7-to-V8 bootstrap migration narrative committed in `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`
6. planning-integration validation sweep recorded in this chunk plus `tests/test_docs_links.py`
7. template-bundle loader, runtime-organization instantiation, fail-closed startup selection logic, and the `v8-migration-standing-team-bridge` bundle landed in `core/internal/bootstrap/template_bundle.go`, `core/internal/bootstrap/template_bundle_test.go`, `core/internal/bootstrap/startup_selection_test.go`, `core/cmd/server/bootstrap_startup.go`, and mirrored chart config packaging
8. clean-run testing discipline is now documented in `docs/TESTING.md`, `docs/architecture/OPERATIONS.md`, and `ops/README.md`
9. lifecycle cleanup/status now sweep and report stray compiled Go services in `ops/lifecycle.py` with focused regression coverage in `tests/test_lifecycle_tasks.py`
10. compiled-service inspection now fails closed: `lifecycle.status` reports unknown inspection state and `lifecycle.down` blocks runtime/integration testing when local process inspection cannot verify cleanup
11. the current local checkpoint series now includes the bounded Team Lead workspace hardening and visibility slices on this branch, with SSH publication still blocked until local key/agent state is repaired
12. documentation cleanup refreshed `README.md`, `docs/README.md`, and `docs/archive/README.md`, and removed `docs/architecture/AGI_ARCHITECTURE_REVIEW_2026-03-06.md` plus `docs/architecture/STANDARDIZATION_REVIEW_2026-03-06.md`
13. validation: `uv run pytest tests/test_docs_links.py -q` -> `23 passed`
14. instantiated-organization provider policy now resolves through `core/internal/swarm/provider_policy.go`, is carried by `core/internal/bootstrap/template_bundle.go`, is applied during startup in `core/cmd/server/main.go`, and is exercised by focused bootstrap/swarm coverage
15. the standing-team bridge bundle now declares a conservative provider-policy default in both `core/config/templates/v8-migration-standing-team-bridge.yaml` and `charts/mycelis-core/config/templates/v8-migration-standing-team-bridge.yaml` so local and charted startup follow the same instantiated-organization routing path as tests
16. startup now fails closed when the bootstrap bundle set is missing or invalid, when `MYCELIS_BOOTSTRAP_TEMPLATE_ID` requests a bundle that is absent, and when multiple bundles exist without an explicit selection; runtime provider routing ignores legacy env-map inputs and startup truth now remains bundle-only in code/tests/state (`core/internal/bootstrap/template_bundle.go`, `core/cmd/server/bootstrap_startup.go`, `core/cmd/server/main.go`, `core/internal/bootstrap/startup_selection_test.go`, `core/cmd/server/bootstrap_startup_test.go`)
17. dedicated V8 UI/API/operator contract PRD now lives in `docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md` and defines first-run flow, AI Organization creation, Team Lead-first workspace behavior, role visibility, advanced-mode boundaries, and screen-to-API mapping
18. architecture-library discovery and in-app docs exposure now include the V8 UI/API/operator contract through `docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md`, `interface/lib/docsManifest.ts`, and `tests/test_docs_links.py`
19. Windows `lifecycle.down` CIM timeout/inspection failure remains separate tooling debt; it is not part of the retired bootstrap fallback surface and still tracks under the lifecycle hardening work until resolved
20. the first V8 creation-flow slice now lives in `interface/app/(app)/dashboard/page.tsx`, `interface/components/organizations/CreateOrganizationEntry.tsx`, `interface/app/(app)/organizations/[id]/page.tsx`, and `interface/components/organizations/OrganizationContextShell.tsx`
21. minimal AI Organization starter/create/home APIs now live in `core/internal/server/templates.go`, `core/internal/server/organizations.go`, and `core/internal/server/admin.go`, with focused backend coverage in `core/internal/server/organizations_test.go`
22. route/API exposure docs now include the V8 AI Organization entry flow through `docs/API_REFERENCE.md`, and frontend regression coverage lives in `interface/__tests__/pages/DashboardPage.test.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/__tests__/shell/ZoneA_Rail.test.tsx`
23. the entry flow now keeps recent-organization resume and starter-template setup resilient under partial API failure, exposes retry/recovery actions in-place, removes operator-visible dev/architecture copy leaks, and strengthens Team Lead status in the AI Organization home (`interface/components/organizations/CreateOrganizationEntry.tsx`, `interface/components/organizations/OrganizationContextShell.tsx`, `interface/__tests__/pages/DashboardPage.test.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`)
24. bounded browser automation for the entry flow now lives in `interface/e2e/specs/v8-organization-entry.spec.ts` and covers dominant AI Organization entry framing, template selection, empty start, organization-home landing, retry/recovery under partial failure, and visible forbidden-copy enforcement
25. the organization page now functions as a Team Lead-first workspace shell rather than a static landing state: `interface/components/organizations/OrganizationContextShell.tsx` keeps the AI Organization header visible, adds a concrete Team Lead identity block plus next-action controls, and switches the active workspace focus between planning, advisor review, department review, AI Engine Settings summary, and Memory & Continuity summary with focused page and browser coverage
26. the default workspace now includes a direct Soma prompt entry, post-create autofocus, current-organization resume path, and clearer degraded-mode reassurance through `interface/components/organizations/CreateOrganizationEntry.tsx`, `interface/components/organizations/TeamLeadInteractionPanel.tsx`, `interface/components/organizations/OrganizationContextShell.tsx`, `interface/components/shell/ZoneA_Rail.tsx`, and `interface/components/dashboard/DegradedModeBanner.tsx`
27. release-candidate validation on 2026-03-22: `uv run inv interface.test` -> pass (`65` files, `404` tests), `uv run inv interface.typecheck` -> pass, `uv run inv interface.e2e` -> pass (`129` passed, `66` skipped), `uv run pytest tests/test_docs_links.py -q` -> pass (`37` passed)
28. accepted non-blockers at the RC gate are warning-only test noise rather than product regressions: repeated `--localstorage-file` worker warnings remain, while the earlier `SensorLibrary` `act(...)` warning and `CircuitBoard.test.tsx` ReactFlow mock-prop warnings have been cleaned up.
26. the first Team Lead interaction workflow now lives in `core/internal/server/organizations.go`, `interface/components/organizations/TeamLeadInteractionPanel.tsx`, `interface/components/organizations/OrganizationContextShell.tsx`, `interface/__tests__/organizations/TeamLeadInteractionPanel.test.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts`; it intentionally stops at guided Team Lead responses and does not yet implement advisor orchestration, raw agent selection, advanced configuration panels, or a full chat system
27. guided Team Lead resilience now adds readable fallback shaping in `core/internal/server/organizations.go`, failure/retry and malformed-response rendering in `interface/components/organizations/TeamLeadInteractionPanel.tsx`, focused backend/frontend tests in `core/internal/server/organizations_test.go`, `interface/__tests__/organizations/TeamLeadInteractionPanel.test.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and bounded browser retry coverage in `interface/e2e/specs/v8-organization-entry.spec.ts`; advisor orchestration, raw agent selection, advanced configuration panels, and a full chat system remain intentionally out of scope
28. inspect-only Advisor and Department visibility now lives in `interface/components/organizations/OrganizationContextShell.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts`; the Team Lead remains the primary workspace counterpart while Advisors and Departments are visible as supporting structure, and full advisor orchestration, raw agent selection, and department management remain intentionally out of scope
29. inspect-only AI Engine Settings and Memory & Continuity workspace surfaces now live in `interface/components/organizations/OrganizationContextShell.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts`; they explain current engine/profile and memory/continuity posture in operator language while keeping advanced provider, capability, memory, and personality controls intentionally out of scope
30. inspect-only Advisor and Department detail views now live in `interface/components/organizations/OrganizationContextShell.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts`; Team Lead action buttons and support-column buttons both open the same focused workspace views while advisor orchestration, department editing, and raw agent selection remain intentionally out of scope
31. inspect-only AI Engine Settings detail now lives in `interface/components/organizations/OrganizationContextShell.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts`; it exposes organization-wide AI engine posture, team-default visibility, and specific-role override visibility in operator language while keeping advanced editing, provider-policy terms, and runtime capability controls intentionally out of scope
32. the first bounded AI Engine edit path now lives in `core/internal/server/organizations.go`, `core/internal/server/admin.go`, `core/internal/server/organizations_test.go`, `interface/components/organizations/OrganizationContextShell.tsx`, `interface/lib/organizations.ts`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts`; operators can choose from curated organization-level AI Engine profiles, invalid values are rejected server-side, retry stays in-place on failure, and team/role-level editing plus advanced provider/capability controls remain intentionally out of scope
33. controlled Department-level AI Engine inheritance now lives in `core/internal/server/organizations.go`, `core/internal/server/admin.go`, `core/internal/server/organizations_test.go`, `interface/components/organizations/OrganizationContextShell.tsx`, `interface/lib/organizations.ts`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts`; Department details expose inherited-versus-overridden AI Engine state, guided Department overrides can be set and reverted, organization-level AI Engine changes continue to update inheriting Departments only, and agent-level overrides plus advanced configuration remain intentionally out of scope
34. the first bounded Response Contract path now lives in `core/internal/server/organizations.go`, `core/internal/server/admin.go`, `core/internal/server/organizations_test.go`, `interface/components/organizations/OrganizationContextShell.tsx`, `interface/lib/organizations.ts`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts`; the workspace exposes `Response Style` as a safe organization-wide default with curated options, invalid values are rejected server-side, retry stays in-place on failure, and agent-level response overrides plus raw prompt/policy editing remain intentionally out of scope
35. inspect-only Agent Type Profile runtime truth now lives in `core/internal/server/organizations.go`, `core/internal/server/organizations_test.go`, `interface/lib/organizations.ts`, `interface/components/organizations/OrganizationContextShell.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts`; Department details now show user-facing agent type names, helps-with summaries, Team-default versus type-specific AI Engine bindings, and Organization/Team-default versus type-specific Response Style bindings without exposing raw model IDs, raw prompt text, or agent-instance editing
36. the first bounded Agent Type AI Engine binding path now lives in `core/internal/server/organizations.go`, `core/internal/server/admin.go`, `core/internal/server/organizations_test.go`, `interface/lib/organizations.ts`, `interface/components/organizations/OrganizationContextShell.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts`; Department details now let operators apply a curated AI Engine to an Agent Type, retry in place on failure, return that role type to the Team default, and preserve Team-default inheritance for other role types without exposing raw model IDs, raw provider terms, or agent-instance overrides
37. the first bounded Agent Type Response Style binding path now lives in `core/internal/server/organizations.go`, `core/internal/server/admin.go`, `core/internal/server/organizations_test.go`, `interface/lib/organizations.ts`, `interface/components/organizations/OrganizationContextShell.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts`; Department details now let operators apply a curated Response Style to an Agent Type, retry in place on failure, return that role type to the Organization / Team default, and preserve inherited organization-wide Response Style behavior for other role types without exposing raw prompt text, raw policy text, or agent-instance overrides
38. `interface/app/(marketing)/page.tsx` and `interface/__tests__/pages/LandingPage.test.tsx` now present and enforce the V8.1 root-page story around AI Organizations, Team Lead-guided work, recent activity, safe guided control, and post-creation workspace expectations without leaking internal architecture terms or old console/chat framing
39. `GET /api/v1/organizations/{id}/learning-insights` now exposes safe read-only learning highlights derived from reviewed organization activity, and `interface/components/organizations/OrganizationContextShell.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts` now render and validate a failure-safe `What the Organization is Learning` support surface without exposing raw embeddings, raw memory stages, or editing controls
40. `ops/cache.py`, `ops/config.py`, `ops/core.py`, `ops/interface.py`, `ops/cognitive.py`, `ops/ci.py`, `tasks.py`, and `tests/test_cache_tasks.py` now define and validate the managed cache policy: repo task runs route uv/pip/npm/go/python-bytecode caches into `workspace/tool-cache`, cleanup/report tasks are explicit, `.pytest_cache` and `workspace/tool-cache/` stay out of git, and operator docs now point at `cache.status`, `cache.clean`, and `cache.apply-user-policy`

### 6. V8 contract shell introduction

Status:
1. `COMPLETE` `docs/architecture-library/V8_RUNTIME_CONTRACTS.md` now exists as the new V8 runtime contract shell.
2. `COMPLETE` the initial V8 runtime contract set is now fully defined.

### 7. Inception contract definition

Status:
1. `COMPLETE` the Inception contract is now defined in `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`.
2. `COMPLETE` it establishes the top-level organization contract for V8.

### 8. Soma Kernel contract definition

Status:
1. `COMPLETE` the Soma Kernel contract is now defined in `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`.
2. `COMPLETE` it establishes the configurable coordination/runtime layer for an Inception.

### 9. Central Council contract definition

Status:
1. `COMPLETE` the Central Council contract is now defined in `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`.
2. `COMPLETE` it establishes configurable advisory composition rather than a fixed built-in council pattern.

### 10. Provider Policy contract definition

Status:
1. `COMPLETE` the Provider Policy contract is now defined in `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`.
2. `COMPLETE` team-level and agent-level configuration scope are now captured explicitly as first-class contract needs.

### 11. Identity and Continuity State contract definition

Status:
1. `COMPLETE` the Identity and Continuity State contract is now defined in `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`.
2. `COMPLETE` the initial V8 runtime contract set is now complete.
3. `COMPLETE` the program has already moved into the config/bootstrap planning phase before backend refactor work begins.

### 12. V8 config/bootstrap planning shell introduction

Status:
1. `COMPLETE` `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md` now exists as the planning shell for V8 config/bootstrap behavior.
2. `COMPLETE` the shell now has initial structure for user-facing and bootstrap-model planning slices.

### 13. User Concept Layer definition

Status:
1. `COMPLETE` the `User Concept Layer` is now defined in `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`.
2. `COMPLETE` architecture-to-user concept mapping is now documented for the V8 UI mental model.
3. `COMPLETE` the beginner-friendly AI-organization mental model is now part of the bootstrap planning surface.

### 14. Bootstrap resolution flow definition

Status:
1. `COMPLETE` the `Bootstrap resolution flow` is now defined in `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`.
2. `COMPLETE` team-level defaults, agent-level overrides, and beginner-versus-advanced bootstrap entry modes are now explicit in the V8 planning model.
3. `COMPLETE` the staged organization-resolution path is now documented for later implementation work.

### 15. Scope inheritance definition

Status:
1. `COMPLETE` the `Scope inheritance` section is now defined in `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`.
2. `COMPLETE` the inheritance chain from `Inception -> Soma Kernel -> Central Council roles -> Team defaults -> Agent overrides` is now documented.
3. `COMPLETE` team and agent configuration are now explicitly treated as first-class scopes in bootstrap planning.

### 16. Precedence rules definition

Status:
1. `COMPLETE` the `Precedence rules` section is now defined in `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`.
2. `COMPLETE` source precedence, scope precedence, policy-before-override handling, and conflict resolution are now part of the V8 bootstrap planning model.
3. `COMPLETE` the planning model now has enough structure to define how organizations are created before bootstrap activation.

### 17. Template and instantiation entry points definition

Status:
1. `COMPLETE` the `Template and instantiation entry points` section is now defined in `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`.
2. `COMPLETE` V8 now distinguishes template-based creation, empty/manual creation, operator/API creation, and config-file/bootstrap creation as canonical organization-entry modes.
3. `COMPLETE` templates are now documented as reusable organization blueprints that feed bootstrap resolution with default kernel, council, team, and policy shape.
4. `COMPLETE` the `Migration from V7 bootstrap assumptions` section now explains how fixed V7 startup assumptions collapse into explicit V8 configuration sources, inheritance, and precedence rules.

### 18. Standing-team bootstrap de-hardcoding plan

Status:
1. `ACTIVE` Chunk 4.2 captured the standing-team de-hardcoding plan inside `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`.
2. `ACTIVE` runtime/config references to fixed `prime-*` teams, hard-coded council rosters, kernel defaults, and provider wiring have documented migration paths into template definitions, bootstrap resolution, scoped inheritance, and provider-policy configuration.
3. `ACTIVE` Helm/runtime concerns (env injection for `MYCELIS_API_KEY`, 8080/8081 port normalization, config-file mounts, container storage) are now described alongside the bootstrap plan so infrastructure slices can align with runtime refactors.
4. `COMPLETE` current-state scan now records that the no-bundle startup dependency is gone: `core/internal/bootstrap/template_bundle.go` now fails closed when bundles are missing, env-map provider overrides are retired, and mirrored Helm/config packaging of `config/teams/*.yaml` no longer participates in startup truth.
5. `ACTIVE` the bridge now owns runtime startup truth through native bundle content: `core/internal/bootstrap/template_bundle.go` loads `core/config/templates/*.yaml` bundles, instantiates a runtime organization directly from bundle-defined org/team/agent data, and fails closed when no valid bootstrap bundle is available.

Evidence:
1. Plan section `## Standing-team bootstrap de-hardcoding plan` (2026-03-13) now explicitly calls out the active files and assumptions discovered in Chunk 4.2 (Helm values/env/ports, ops bridge tasks, bootstrap services, template APIs, team manifests, runtime storage expectations).
2. README review confirmed no additional guidance changes were required after adding the plan (`README.md`, 2026-03-13).
3. `v8-migration-standing-team-bridge` now validates as a self-contained V8 bundle through the bootstrap loader path without requiring standing-team manifest refs (`core/config/templates/v8-migration-standing-team-bridge.yaml`, `core/internal/bootstrap/template_bundle_test.go`).
4. startup selection regression coverage now explicitly proves bundle-instantiated runtime organizations, missing-bundle failure, invalid-bundle failure, missing-requested-bundle failure, and registry continuity with bundle presence (`core/internal/bootstrap/startup_selection_test.go`, `core/cmd/server/bootstrap_startup_test.go`, `core/internal/swarm/registry_test.go`).

Next steps:
1. Promote generated per-organization bootstrap bundles so startup stays bundle-only without depending on the fixed standing-team bridge asset forever.
2. Replace bootstrap seeding logic with template-instantiation + scope-aware inheritance.
3. Promote provider-policy scopes and Helm env/port/mount/storage alignment into actionable runtime slices with tests.

### 19. Cluster/runtime bootstrap contract alignment

Status:
1. `COMPLETE` Helm deployment now provisions the configuration bundle (`cognitive.yaml`, `policy.yaml`, standing-team YAMLs) via a ConfigMap volume so Pods read deterministic bootstrap inputs instead of whatever was baked in the image.
2. `COMPLETE` `MYCELIS_API_KEY` is required and injected through a Kubernetes Secret created by the chart (or supplied via `coreAuth.existingSecret`); `ops.k8s.deploy` now refuses to proceed when the key is missing.
3. `COMPLETE` Core HTTP port contract is unified on `8080` (`core/cmd/server` default, Helm `PORT` env, Service/bridge forwarding, ops defaults).
4. `COMPLETE` Charts and ops scripts document/mount the writable storage contract (`/data` PVC for artifacts + `$MYCELIS_WORKSPACE` under `/data/workspace`), and the ConfigMap keeps read-only bootstrap files under `/app/config`.

Evidence:
1. Helm templates: `charts/mycelis-core/templates/deployment.yaml`, `configmap-config.yaml`, `_helpers.tpl`, `core-auth-secret.yaml`, updated `values.yaml`, plus new `config/` assets.
2. Ops automation: `ops/k8s.py` enforces API key injection and 8080 port-forward; `ops/config.py` defaults shift to 8080; `ops/interface.py` startup guidance updated.
3. Runtime entrypoint: `core/cmd/server/main.go` now defaults to `PORT=8080`.
4. Validation: `uv run pytest tests/test_docs_links.py -q`.

Next steps:
1. Wire provider-policy scopes and template-instantiation flow into the runtime without depending on standing-team tables.
2. Extend Helm/ops surface so template bundles are generated from the new template serialization path instead of the baked defaults once Task 005 code work lands.

### 20. Testing/QA alignment review

Status:
1. `COMPLETE` browser QA test plan for Workspace chat now references the V8 contract (inline Soma chat, terminal states, direct-first routing) instead of the outdated V7 framing.
2. `COMPLETE` manual testing expectations now list the canonical outcomes (`answer`, `proposal`, `execution_result`, `blocker`) and reinforce the V8 inline-governance posture.

Evidence:
1. `tests/ui/browser_qa_plan_workspace_chat.md` updated on 2026-03-13 to remove V7 references and describe the V8 workspace, happy paths, and edge cases.

Next steps:
1. Keep automated Playwright suites aligned with the same inline-chat expectations once Workspace refactors land.

### 21. V8.1 living organization architecture definition

Status:
1. `COMPLETE` `docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md` now defines the canonical V8.1 architecture contract for Loop Profiles, Learning Loops, Runtime Capabilities, promoted Response Contract inheritance, promoted Agent Type Profiles, Memory Promotion and Semantic Continuity, Procedure / Skill Sets, and bounded Automations visibility.
2. `COMPLETE` the architecture-library index, in-app docs manifest, and doc-test contract now expose V8.1 as canonical architecture truth rather than leaving `v8-1.md` as a loose planning draft.
3. `ACTIVE` the first shippable V8.1 state now includes bounded read-only Review Loop execution, scheduled and event-driven activity visibility, read-only Automations visibility, and read-only learning visibility in the Team Lead workspace; next work centers on bundle/config loop truth, capability contract surfaces, semantic continuity planning, and the first bounded Learning Loop implementation behind this operator surface.

Evidence:
1. `docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md` now carries the canonical V8.1 PRD for Loop Profiles, Learning Loops, Runtime Capabilities, promoted Response Contract and Agent Type Profile runtime truth, semantic continuity, procedure/skill memory, safety rules, testing requirements, and initial release definition.
2. `docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md`, `interface/lib/docsManifest.ts`, and `tests/test_docs_links.py` now index, expose, and enforce the V8.1 architecture doc as a canonical surface.
3. `docs/architecture-library/V8_RUNTIME_CONTRACTS.md` and `docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md` now cross-link the V8.1 architecture extension so runtime and operator docs do not drift into a parallel truth.

### 22. First read-only Review Loop execution

Status:
1. `COMPLETE` a minimal backend Loop Executor now supports the first bounded Review Loop path inside `core/internal/server`, including Loop Profile loading, owner resolution, structured review output, and in-memory result storage.
2. `COMPLETE` manual internal triggering now exists through `/api/v1/internal/organizations/{id}/loops/{loopId}/trigger`, and recent loop results are inspectable through `/api/v1/internal/organizations/{id}/loops/results` for debugging.
3. `COMPLETE` the first loop execution remains safely bounded: review-only, no actuation, no external calls, no filesystem/hardware access, and no UI exposure of raw loop internals.
4. `COMPLETE` interval-backed Review Loops now execute automatically through a minimal in-process scheduler that only runs profiles with `interval_seconds`, prevents overlap, and stays stoppable during shutdown.
5. `COMPLETE` read-only Recent Activity visibility now surfaces the latest review/check/update outcomes inside the Team Lead workspace without exposing loop controls, raw logs, or technical internals.
6. `COMPLETE` the first event-driven Review Loop milestone now reacts to bounded internal organization events (`organization created`, Team Lead guidance completion, organization AI Engine change, and Response Style change), executes read-only reviews, records safe activity results, and shares overlap protection with scheduled execution.
7. `COMPLETE` the Team Lead workspace now exposes a read-only `Automations` surface that shows what ongoing reviews/checks exist, whether they are scheduled or event-driven, who owns them in user-facing terms, and what they have reported recently without exposing raw loop/scheduler internals or editing controls.
8. `COMPLETE` the Team Lead workspace now exposes a read-only `What the Organization is Learning` surface that summarizes recent learning highlights in operator language with user-facing source, recency, and strength labels while remaining failure-safe and non-mutating.
9. `NEXT` promote loop definitions into bundle/config truth so scheduled and event-driven Automations come from reproducible organization configuration rather than only seeded defaults.

Evidence:
1. `core/internal/server/review_loops.go` now defines the first V8.1 Review Loop framework, default loop profiles, team/agent-type owner resolution, structured findings/suggestions/status output, and read-only result logging.
2. `core/internal/server/admin.go` now registers the internal trigger/debug routes, and `core/internal/server/organizations.go` seeds default review loops when new organizations are created.
3. `core/internal/server/review_loops_test.go` now proves successful execution, owner resolution, structured output, invalid-loop rejection, stored result visibility, and read-only preservation of organization state.
4. `core/internal/server/review_loop_scheduler.go` and `core/internal/server/review_loop_scheduler_test.go` now add the first bounded scheduled-loop runner with interval-based execution, invalid-config rejection, overlap protection, stoppable lifecycle wiring, and result/failure logging.
5. `GET /api/v1/organizations/{id}/loop-activity` now exposes safe user-facing activity summaries, and `interface/components/organizations/OrganizationContextShell.tsx` now renders them in a non-intrusive `Recent Activity` support panel with lightweight polling, empty-state handling, and failure-safe fallback.
6. `core/internal/server/review_loops.go`, `core/internal/server/organizations.go`, `core/internal/server/review_loops_test.go`, and `core/internal/server/organizations_test.go` now add bounded event-driven review execution for allowed internal organization events, safe failure logging into Recent Activity, and shared overlap protection so reactive reviews stay read-only and operator-visible without exposing raw event-bus internals.
7. `GET /api/v1/organizations/{id}/automations` now exposes safe read-only Automation definitions, and `interface/components/organizations/OrganizationContextShell.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts` now render and validate an inspect-only `Automations` support surface that stays separate from `Recent Activity` while preserving the Team Lead-first workspace.
8. `GET /api/v1/organizations/{id}/learning-insights` now exposes safe read-only learning highlights derived from reviewed organization activity, and `interface/components/organizations/OrganizationContextShell.tsx`, `interface/__tests__/pages/OrganizationPage.test.tsx`, and `interface/e2e/specs/v8-organization-entry.spec.ts` now render and validate a failure-safe `What the Organization is Learning` panel that stays operator-friendly and non-mutating.
9. `interface/components/shell/ZoneA_Rail.tsx`, `interface/app/(app)/automations/page.tsx`, `interface/components/automations/AutomationHub.tsx`, `interface/app/(app)/resources/page.tsx`, `interface/app/(app)/system/page.tsx`, `interface/app/(app)/memory/page.tsx`, `interface/app/(app)/settings/page.tsx`, and supporting redirect routes now enforce the MVP route audit: core Team Lead-first paths stay visible by default, advanced support surfaces are explicitly gated, and stale labels are replaced with operator-facing terms.

### 23. V8.1 learning continuity architecture synchronization

Status:
1. `COMPLETE` synchronized the canonical V8.1 architecture and linked runtime/index/state surfaces so Learning Loops are now first-class loop subtypes with candidate capture, review/promotion pathing, policy-bounded memory promotion, and no silent self-rewrite.
2. `COMPLETE` clarified pgvector as the semantic continuity substrate for event/action/result indexing, review memory, learning candidates, promoted organization/team/agent-type memory, procedure/skill retrieval, and continuity recall.
3. `COMPLETE` defined Procedure / Skill Sets as reviewed, type-bound specialist memory under Agent Type Profiles, and removed stale state wording that still implied read-only Automations visibility was pending.
4. `COMPLETE` added the first operator-visible learning surface as a safe read-only summary layer over reviewed organization activity without exposing raw semantic-continuity or memory-internal terms.
5. `NEXT` turn this architecture sync into implementation planning for bundle-defined loop truth, memory-promotion runtime contracts, and the first bounded Learning Loop slice behind the read-only visibility layer.

Evidence:
1. `docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md` now defines Learning Loops, raw/reviewed/promoted memory stages, semantic continuity layering, and Procedure / Skill Sets as canonical V8.1 architecture truth.
2. `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`, `docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md`, `interface/lib/docsManifest.ts`, and `tests/test_docs_links.py` now cross-link and enforce the semantic continuity and learning-layer model.
3. `V8_DEV_STATE.md` now reflects that read-only Automations visibility is complete, while memory-promotion planning and the first bounded Learning Loop remain next.

### 24. V8.1 MVP route/tab and UI/API audit before RC lock

Status:
1. `COMPLETE` audited the default operator-visible route set against the Team Lead-first V8.1 model and reduced the default shell to `AI Organization`, `Automations`, `Docs`, and `Settings`, with `Resources`, `Memory`, and `System` kept as advanced support routes.
2. `COMPLETE` revised or redirected stale MVP surfaces so legacy `Draft Blueprints`, `Matrix`, direct `Catalogue` / `Marketplace`, and other pre-V8.1 labels now either route into the surviving advanced tabs or stay out of the default workflow.
3. `COMPLETE` tightened request handling and browser validation around the MVP route gate: loading/error states remain failure-safe, Playwright now runs through a managed local server + managed browser cache path, and the default E2E suite skips legacy V7/raw-endpoint probes that no longer define the MVP surface.
4. `COMPLETE` reviewed advanced-config boundaries and kept AI engine / connected-tool setup behind advanced mode while leaving bundle/bootstrap/provider-policy truth in file/env/config surfaces instead of widening default UI controls.

Evidence:
1. `interface/components/shell/ZoneA_Rail.tsx`, `interface/app/(app)/automations/page.tsx`, `interface/app/(app)/resources/page.tsx`, `interface/app/(app)/memory/page.tsx`, `interface/app/(app)/system/page.tsx`, `interface/app/(app)/settings/page.tsx`, and their redirect pages now reflect the MVP keep/revise/remove decisions in code.
2. `docs/architecture/FRONTEND.md`, `docs/TESTING.md`, `README.md`, and `docs/architecture/OPERATIONS.md` now document the Team Lead-first route inventory, advanced-route boundary, and managed Playwright runtime/testing contract.
3. `interface/e2e/specs/navigation.spec.ts`, `interface/e2e/specs/missions.spec.ts`, `interface/e2e/specs/mobile.spec.ts`, `interface/e2e/specs/v8-organization-entry.spec.ts`, and the skipped legacy/raw coverage specs now align the default browser gate with the surviving MVP workflow.
4. Validation on 2026-03-20: `uv run pytest tests/test_docs_links.py -q` -> pass (`25` passed), `uv run inv interface.test` -> pass (`65` files, `401` tests), `uv run inv interface.typecheck` -> pass, `uv run inv interface.e2e` -> pass (`129` passed, `63` skipped).

### 25. V8.2 MVP release-candidate stabilization and suite classification

Status:
1. `COMPLETE` re-ran the broader interface suite from the current Soma-primary RC state and confirmed that the earlier failures were not live product regressions in the entry flow, workspace re-entry, Soma interaction, navigation, or core support panels.
2. `COMPLETE` classified the remaining browser failure as a stale test assumption in the general navigation spec rather than an MVP blocker: the spec was reusing one page instance across primary-route navigations and colliding with the docs page's own client-side navigation.
3. `COMPLETE` updated that stale browser expectation so route-highlighting validation now resets to a stable dashboard starting point before each primary-route navigation.
4. `COMPLETE` revalidated the full RC gate after the classification pass: full interface tests, typecheck, managed E2E, and doc-link checks are green.
5. `COMPLETE` accepted the remaining warning-only suite noise as non-blocking for the RC gate because it does not affect real operator behavior on the default Soma-first MVP path.

Accepted non-blockers:
1. Vitest worker startup still emits repeated `--localstorage-file` warnings.

Evidence:
1. `interface/e2e/specs/navigation.spec.ts` now resets to `/dashboard` before each primary-route navigation, which removes the stale page-reuse assumption without widening the MVP route surface.
2. Validation on 2026-03-22: `uv run inv interface.test` -> pass (`65` files, `404` tests), `uv run inv interface.typecheck` -> pass, `uv run inv interface.e2e` -> pass (`129` passed, `66` skipped), `uv run pytest tests/test_docs_links.py -q` -> pass (`37` passed).

## Immediate Next Actions

1. `NEXT` continue Phase 2 of the approval/product-trust lane: extend visible content/artifact value delivery beyond the first Launch Crew artifact-reference slice so drafting requests return inline value and durable outputs return explicit artifact references or previews wherever Soma surfaces the result.
2. `REQUIRED` keep future delivery slices on clean committed checkpoints instead of mixed local batches.
   - do not reopen a large cross-surface dirty tree
   - prove each slice from committed state with the exact command set recorded here
3. `NEXT` finish the live governed-chat stabilization lane.
   - reproduce the fresh-organization direct-answer failure under `interface/e2e/specs/soma-governance-live.spec.ts`
   - identify whether the fault is route normalization, backend panic/error envelope drift, or proposal/approval runtime state
   - restore selector parity with the current `Approve & Execute` / `Execute` UI contract
4. `NEXT` rerun the full live-service gate once the governed lane is repaired.
   - `uv run inv ci.service-check --live-backend`
   - any additional focused live browser proof needed for the repaired path
   - update this state file from `BLOCKED` only after those commands pass from committed state
5. `REQUIRED` keep the task/operator contract synchronized in the same slices that change it.
   - `README.md`
   - `docs/TESTING.md`
   - `docs/LOCAL_DEV_WORKFLOW.md`
   - `docs/architecture/OPERATIONS.md`
   - `ops/README.md`
   - `interface/lib/docsManifest.ts` whenever the in-app docs surface changes
6. `NEXT` continue documentation authority cleanup with intent.
   - keep canonical planning, delivery-governance, and UI-target docs under `docs/architecture-library/`
   - avoid letting temporary strike notes float untracked or mixed with product code
   - index any newly canonical doc in `interface/lib/docsManifest.ts` in the same slice
7. `REQUIRED` keep all new validation checkpoints and release-shaping blocker transitions recorded here in `V8_DEV_STATE.md` as they happen, rather than leaving them only in chat history or commit messages.
8. `NEXT` extend the same owner-scoped MCP configuration posture into any future MCP settings or toolset-management UI so current-group context stays explicit once those surfaces become operator-visible.
9. `NEXT` continue Slice 2 of the ask-class / agent-type / output-contract assertion lane.
   - extend browser proof from current guided organization chat into more artifact-heavy and specialist-heavy default-path requests
   - widen browser proof beyond the current targeted artifact and specialist examples into richer organization-default asks
   - keep the current four terminal states stable while broadening assertion coverage
10. `NEXT` execute the workflow-complete verification lane from `V8_USER_WORKFLOW_EXECUTION_AND_VERIFICATION_PLAN.md`.
   - map every primary user workflow to current automated coverage
   - identify where tests prove only immediate interaction and tighten them to real workflow outcomes
   - verify team-shaping and MCP-association paths against the user-shaped delivery and framework-usability rules
11. `NEXT` continue workflow-complete hardening on the next user-shaped lanes.
   - tighten team-design/browser proof so it verifies user-shaped delivery outcome instead of only lane entry
   - tighten MCP workflow proof so it verifies standards-compliant association and usable continuation, not just visible registry rendering
12. `NEXT` relaunch from WSL/Linux tooling for the next provider/runtime pass.
   - re-run the optional local `cognitive.*` helper lane from Linux tooling instead of Windows host assumptions
   - verify repo-local vLLM install/start/status behavior on a supported Linux environment
   - keep the current documented split explicit: Ollama remains the local Windows default, while repo-local vLLM helpers are validated from Linux-capable tooling or via remote OpenAI-compatible endpoints


