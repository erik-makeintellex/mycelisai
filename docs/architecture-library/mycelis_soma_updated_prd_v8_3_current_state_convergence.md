# Mycelis V8.3 Current-State Convergence PRD
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Index](ARCHITECTURE_LIBRARY_INDEX.md)

> Status: `ACTIVE` candidate for `MYC-RC-000`
> Date: 2026-05-29
> Purpose: converge the current implementation toward package-first, proof-backed Soma delivery without creating a parallel greenfield architecture.

## Authority Boundary

This document is candidate convergence authority for `MYC-RC-000`. It does not supersede [V8 Development State](../../.state/V8_DEV_STATE.md), [Project README](../../README.md), [V8.2 Production Architecture Target](../../architecture/v8-2.md), [V8 Runtime Contracts](V8_RUNTIME_CONTRACTS.md), [V8 UI/API and Operator Experience Contract](V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md), [V8 MVP Media, Team Output, And Template Registry](V8_MVP_MEDIA_TEAM_OUTPUT_AND_TEMPLATE_REGISTRY.md), or [V8 Workflow Variants And Reboot Proof Set](V8_WORKFLOW_VARIANTS_AND_REBOOT_PROOF_SET.md).

`MYC-RC-001` and later slices may begin only after `MYC-RC-000` records accepted terminology, proof lanes, owners, and validation gates in state.

## Current Truth

The current implementation already has enough substrate to build the next coherent demo: Go Core, Next.js / React Interface, PostgreSQL / pgvector, NATS, Ollama and provider routing, governed proposal / confirm flow, mission runs and events, execution summaries, proof artifacts, retained outputs, team work infrastructure, MCP/capability infrastructure, and dashboard/runs/groups/output surfaces.

The convergence path should use those pieces. It should not start by introducing SQLite, an in-memory event bus, a new local prototype, a `codex_tasks` table, a broad recovery table, or autonomous team writes.

Windows remains the source-edit and local-validation surface. WSL remains the authoritative release-proof surface when release confidence matters.

## Product Thesis

The user does not receive raw tool output. The user receives a coherent, reviewable delivery package.

Default Soma UX should show:

- what Soma understood
- what Soma proposes to do
- what will be created
- what was created
- what is final, variant, editable, or exportable
- what proof exists
- what the next action is

Raw NATS subjects, MCP parameters, database identifiers, tool JSON, stack traces, provider prompts, and proof payload internals belong behind Inspect / Advanced.

## Terminology Decisions

### DeliveryPackage

`DeliveryPackage` is the user-facing grouping concept for a mission result. It must map onto existing retained-output and `project_package` behavior before becoming a new persisted object.

Minimum package fields: `package_id`, `run_id`, `title`, `summary`, `package_kind`, `status`, `primary_output_ref`, `items`, `next_actions`, `proof_ref`, and `inspect_refs`.

Initial package kinds: `codex_task_bundle`, `document_bundle`, `media_bundle`, `creative_bundle`, and `implementation_bundle`.

Do not flatten multi-artifact missions into one output card.

### OutputRef

`OutputRef` is the shared normalization boundary for retained outputs. It should absorb existing shapes such as `ExecutionOutput`, `TeamOutputRef`, `ChatArtifactRef`, `project_package` outputs, MCP results, and team results.

Minimum fields: `id`, `kind`, `label`, `storage_ref`, `open_url`, `reveal_url`, `source_run_id`, `source_contract_id`, `proof_artifact_id`, `package_id`, `role`, and `status`.

Initial roles: `final`, `variant`, `primary_task`, `editable_source`, `validation`, `context`, `proof`, `export`, `note`, `agent_summary`, and `source_asset`.

### Proof Envelope

Every meaningful write or retained output should produce a minimum proof envelope: content fingerprint, path / boundary verification, execution verification status, source run, source contract, validation source, evidence strength, and recovery options when degraded.

`write_file` proof must verify that the file exists, is inside the authorized workspace or artifact boundary, has a captured checksum, links to an output ref, and links to proof.

## First Demo Contract

The first coherent demo is a Codex task delivery package created through existing governance:

```text
User intent
-> Soma proposal
-> approval through confirm-action flow
-> mission_run
-> mission_events
-> Codex task package artifacts
-> output refs
-> proof artifact
-> package-first UI card
```

The package should contain `implementation_task.md`, `validation_checklist.md`, `affected_files_summary.md`, `next_recommended_slices.md`, and a proof artifact.

The demo is not valid if it writes local files outside proposal/confirm, bypasses mission runs/events/proof, hides the result in logs, requires raw JSON to understand success, or writes outside authorized paths.

## Run And Timeline Decisions

Preserve `mission_runs.status` as the DB status field. It may remain narrow: `pending`, `running`, `completed`, and `failed`.

Expose richer operator state as a projection from `mission_runs`, `mission_events`, execution summaries/contracts, proof artifacts, output refs, delivery package manifests, and degradation/recovery metadata.

Dashboard, `/runs`, Activity, and Soma package surfaces should consume one shared timeline/package projection instead of inventing separate run truth.

User-facing timeline sections:

- intent received
- proposal generated
- approval required
- approval granted
- mission running
- tool/action execution
- package generated
- proof generated
- recovery available
- next actions

## Recovery And Teams

Keep recovery embedded for now in proof/team-work payloads. Do not add a first-class `recovery_actions` table until multiple runtime paths prove one stable recovery object shape.

Runtime teams remain bounded collaborators. Teams may analyze, propose, and produce retained results, but write steps stay governed by user-approved confirmation until async result/proof projection is universal.

## Slice Board

### MYC-RC-000 - Convergence Authority Lock

Purpose: classify this PRD, synchronize state/docs/navigation, map terminology to existing runtime/UI objects, and lock validation gates.

Owners:

- Delivery Orchestrator
- Documentation / State
- Validation / Release

Acceptance:

- `.state/V8_DEV_STATE.md` names the convergence status.
- README and architecture index link this document with candidate scope.
- In-app docs manifest is reviewed and changed only if the doc becomes operator-facing authority.
- Validation gates use `uv run inv ...`.
- `MYC-RC-001` through `MYC-RC-003` have non-overlapping owners.

### MYC-RC-001 - Codex Task Artifact Demo Path

Purpose: generate a Codex-ready task bundle through proposal/confirm/run/event/proof flow.

Runtime owners:

- `core/internal/server/cognitive_chat_handler.go`
- `core/internal/server/cognitive_tool_plan.go`
- `core/internal/server/cognitive_tool_plan_parse.go`
- `core/internal/server/templates_execution.go`
- `core/internal/server/templates_confirm_action.go`
- `core/internal/server/confirm_action_outputs.go`
- `core/internal/swarm/internal_tools_registration_execution.go`
- `core/internal/swarm/internal_tools_handlers_io.go`
- `core/pkg/protocol/execution_summary.go`

UI owners:

- `interface/components/dashboard/MissionControlChat.tsx`
- `interface/components/dashboard/MissionControlMessageBubble.tsx`
- `interface/components/soma/ExecutionSummaryCard.tsx`
- `interface/components/soma/OutputWorkbench.tsx`
- `interface/components/runs/RunTimeline.tsx`

Acceptance: user can approve, run, view, open, and inspect proof for a generated Codex task package.

### MYC-RC-002 - OutputRef Normalization

Purpose: normalize retained outputs across file, media, MCP, team, template, and Codex-task paths.

Owners:

- `core/pkg/protocol/execution_summary.go`
- `core/internal/server/confirm_action_outputs.go`
- `core/internal/server/templates_codex_package.go`
- `core/internal/server/execution_summary_mcp.go`
- `core/internal/server/team_work_confirm_action_helpers.go`
- `core/internal/server/team_group_execution_summaries.go`
- `interface/store/cortexStoreTypesChat.ts`
- `interface/store/cortexStoreOutputRefs.ts`
- `interface/store/cortexStoreMissionChatSlice.ts`
- `interface/store/cortexStoreProposalExecutionSlice.ts`
- `interface/store/cortexStoreTypesTeamWork.ts`
- `interface/components/soma/ExecutionSummaryCardModel.ts`
- `interface/components/soma/OutputWorkbench.tsx`

Acceptance: output shape includes label, kind, storage ref, open/reveal link, source run, source contract, proof ref, status, package id, and role.

Current slice: Codex task packages now project backend `OutputRef`, `DeliveryPackageSummary`, and `ExecutionSummary` fields through confirm-action and package event payloads, while Interface store boundaries normalize package output refs into the existing Output Workbench. Broader media/team/MCP/template adoption remains follow-on work.

### MYC-RC-002B - Package Manifest And Output Card

Purpose: present multi-output results as one coherent package in OutputWorkbench.

Owners:

- `interface/components/soma/OutputWorkbench.tsx`
- `interface/components/soma/OutputAccessActions.tsx`
- `interface/components/soma/SomaOperatingSurface.tsx`
- `interface/components/teams/teamWorkProjection.ts`
- `interface/components/teams/ActiveWorkEvidence.tsx`
- `interface/components/teams/WorkTruthSummary.tsx`
- `interface/components/teams/GroupDetailPane.tsx`

Acceptance: Codex task bundle appears as one package card with primary task, checklist, affected files, next slices, proof, and Inspect details.

### MYC-RC-003 - Minimum Universal Proof Envelope

Purpose: apply a minimum proof envelope to meaningful write/output actions.

Owners:

- `docs/architecture-library/MYC_RC_003_PROOF_ENVELOPE_DELIVERY_PLAN.md`
- `core/pkg/protocol/execution_summary.go`
- `core/internal/artifacts/service.go`
- `core/internal/server/templates_codex_package.go`
- `core/internal/server/confirm_action_outputs.go`
- `core/internal/server/templates_confirm_action.go`
- `core/internal/server/templates_codex_package_test.go`
- `core/internal/swarm/internal_tools.go`
- `core/internal/swarm/internal_tools_handlers_io.go`
- `core/internal/server/templates_execution.go`
- `interface/store/cortexStoreTypesChat.ts`
- `interface/store/cortexStoreOutputRefs.ts`
- `interface/components/soma/OutputWorkbench.tsx`
- `interface/components/runs/EventCard.tsx`
- `interface/e2e/specs/codex-package-retention.spec.ts`

Acceptance: write-file proof records checksum, path boundary verification, execution status, output ref, and proof readback.

Current slice: `MYC-RC-003A` starts with Codex task packages. It adds artifact-linked package-file proof, real artifact IDs on output refs, proof/readback metadata, and Windows-managed full-browser proof while WSL remains environment-blocked.

Status: `MYC-RC-003A` backend (artifact persistence, readback, `proof_envelopes` on `artifact.created`) and UI (Output Workbench proof chip and Inspect details, degraded rendering for legacy package messages without proof) are landed. Consistent with the `MYC-RC-003` plan, this slice stays uncommitted to release until full browser proof (Windows-managed `interface.e2e` live-backend run) closes the gate.

## Validation Gates

Use runner-owned validation:

```bash
uv run inv ci.entrypoint-check
uv run inv core.test
uv run inv interface.test
uv run inv interface.typecheck
uv run inv interface.build
uv run inv quality.max-lines --limit 300
git diff --check
```

For live package proof:

```bash
uv run inv compose.up --build --wait-timeout=240
uv run inv compose.health
uv run inv compose.storage-health
uv run inv interface.e2e --project=chromium --live-backend --workers=1 --server-mode=start --spec=e2e/specs/soma-governance-live.spec.ts
```

For release authority:

```bash
uv run inv wsl.status
uv run inv wsl.refresh
uv run inv wsl.validate --lane=release --compose-wait-timeout=240
```

Do not use `uv run inv native-infra.status`; that task is not registered in this checkout.
