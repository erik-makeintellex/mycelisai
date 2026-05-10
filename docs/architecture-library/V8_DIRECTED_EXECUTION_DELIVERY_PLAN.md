# V8 Directed Execution Delivery Plan
> Navigation: [Project README](../../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)

> Status: ACTIVE
> Last Updated: 2026-05-09
> Module Boundary: advanced UI, runtime/deployment, capability/MCP, governance/trust, team/workflow
> Purpose: Orchestrate the delivery teams that turn the directed-execution directive and capability-manifest standard into product behavior.

## Delivery Objective

Make Mycelis feel like a governed AI operating environment:

```text
Intent -> Soma understanding -> directed execution -> capabilities/teams -> runs/outputs/proof -> review/recovery
```

The release goal is not another panel. The release goal is one coherent execution story that is visible in Soma, durable in runtime, and provable in browser tests.

## Team Roster

| Team | Status | Owns | First Output |
| --- | --- | --- | --- |
| Orchestration Lead | ACTIVE | Plan, sequencing, dependency resolution, state/docs sync, acceptance decisions. | Keep this plan current and coordinate gates. |
| Runtime/Run Spine | IN_REVIEW | Run attachment, execution identity, output references, recovery metadata. | Shared `ExecutionSummary`/run-link contract for direct and guided proposal paths. |
| Output/Exchange | ACTIVE | Durable output object mapping, retained artifacts, exchange normalization. | Output object shape and proof linkage across chat, tools, teams, and artifacts. |
| Capability/MCP | NEXT | Capability manifest registry, MCP-to-capability mapping, health/fallback. | Capability list that answers "what Soma can use" with risk, schema, approval, output destinations. |
| Governance/Trust | IN_REVIEW | Proposal/approval/audit language, risk policy, proof semantics. | Default-safe proof copy plus advanced audit detail mapping. |
| Interface/Soma UX | IN_REVIEW | Default Soma execution summary, proof links, output cards, next-step display. | Soma response surface showing intent, understanding, execution shape, capability/team use, outputs, proof, next step. |
| Interface/Advanced UX | NEXT | Connected Tools capability view, Runs proof view, System/Deployments trust view. | Advanced surfaces that reveal manifests, run traces, deployment roots, and recovery without polluting default UX. |
| Validation | REQUIRED | Browser-visible proof, unit/API contract coverage, failure/recovery checks. | Test matrix for directed execution, capability use, output durability, governance, and deployment trust. |
| Docs/In-App Docs | ACTIVE | Architecture/user/state/docs manifest alignment. | Keep canonical docs and in-app docs synced for each slice. |

## Work Packages

| ID | Priority | Owner | Work Package | Acceptance |
| --- | --- | --- | --- | --- |
| DE-1 | 1 | Runtime/Run Spine + Interface/Soma UX | Define and consume a reusable execution-summary contract. | Soma can display intent, understanding, execution shape, capability/team use, outputs, proof, and next step from one normalized object. |
| DE-2 | 2 | Runtime/Run Spine + Output/Exchange | Ensure meaningful actions attach to runs and meaningful outputs attach to durable output objects. | Direct retained answers, proposals, tool use, team execution, automations, and plugins all have run/output proof or explicit non-retained classification. |
| DE-3 | 3 | Interface/Soma UX + Governance/Trust | Put proof links and proposal/audit/recovery state next to meaningful outputs. | User can reach run proof, approval/audit evidence, and recovery state from the post-execution workflow. |
| DE-4 | 4 | Capability/MCP + Interface/Advanced UX | Reshape Connected Tools around governed capabilities rather than server inventory. | Resources shows what Soma can use, availability, risk, approval posture, schemas, output types, and destinations; raw MCP details sit behind disclosure. |
| DE-5 | 5 | Interface/Advanced UX + Runtime/Deployment | Add deployment trust visibility. | System/Deployments shows checkout, deployment root, execution root, artifact/log/cache roots, current commit, proof status, and recovery action. |
| DE-6 | 6 | Governance/Trust + Validation | Normalize proposal/proof/recovery language and test it. | Proposal never mutates before approval; success always has proof; failures have bounded recovery. |
| DE-7 | 7 | Orchestration Lead + Interface/Soma UX | Collapse dashboard/admin-feeling duplicate workflow surfaces. | Default path feels like Soma-directed work, with advanced runtime/admin structure behind layers. |

## Execution Waves

### Wave 1: Trust Spine

Start with DE-1, DE-2, and DE-3.

Rationale:
- The user must see directed execution before capability polish matters.
- Runs and outputs must become natural proof surfaces.
- Governance trust depends on proof being close to execution.

Exit gate:
- one direct retained answer path
- one guided proposal path
- one tool-assisted path
- one team/group retained-output path
- all visibly carry execution summary, run/output proof, and recovery/audit state where applicable

Current implementation status:
- `IN_REVIEW` direct Soma answers, guided proposals, and confirmed proposal execution now share the additive `execution_summary` runtime payload.
- `IN_REVIEW` the Soma chat surface renders intent, understanding, execution shape/status, capability use, outputs, proof, audit/recovery, and next step from that payload.
- `IN_REVIEW` tool-assisted chat/search now preserves read-only `tools_used`, classifies tool/artifact responses as `tool_assisted_work`, adds `execution_summary` to direct `web_search`, and marks blocked search as `blocked`.
- `IN_REVIEW` Team Lead guidance and group broadcast accepted responses now attach `execution_summary` without fabricating `run_id`; group/team proof stays audit/group scoped until a concrete run exists.
- `IN_REVIEW` focused browser-visible component proof now covers direct tool-assisted Soma search summaries and Team Lead execution summaries, including the no-fabricated-run-link boundary.
- `ACTIVE` remaining Wave 1 release work is live-backend browser proof for group broadcast proof visibility, then WSL release proof from the committed state.

Wave 1 validation evidence:
- `go test ./pkg/protocol ./internal/server -run "TestChatResponsePayload_ExecutionSummaryIsAdditive|TestHandleChat_UnwrapsReadableJSONEnvelopeFromAgent|TestHandleChat_RoutesLatestMutationTurnToProposalAcrossThreadHistory|TestHandleConfirmAction_CompletesVerifiedExecutionWithPlannedToolCalls|TestHandleConfirmAction_NormalizesWriteFileAliasesInStoredPlan" -count=1 -v`
- `cd interface; npx tsc --noEmit`
- `cd interface; npx vitest run __tests__/dashboard/MissionControlChat.outputs.test.tsx __tests__/dashboard/MissionControlChat.executionSummary.test.tsx`
- `uv run pytest tests/test_docs_links.py tests/test_documentation_layout_contract.py -q`
- `uv run inv quality.max-lines --limit 300`
- `git diff --check`

Next team handoff:
- Runtime/Tool-Assisted should classify non-mutating chat responses with tools or artifacts as `tool_assisted_work`, preserve read-only `tools_used`, add `execution_summary` to direct `web_search` responses, and cover blocked search as `blocked`.
- Team/Group Retained Output should add `execution_summary` to Team Lead guidance and group broadcast/retained-output responses without inventing run proof where only audit/group proof exists.
- Interface should reuse `ExecutionSummaryCard` near existing Team Lead execution contracts and avoid duplicating retained artifact lists already shown in group detail views.
- Validation should extend focused server tests for direct search, tool-backed chat, guidance contracts, group broadcast proof, and targeted organization UI tests before live browser proof.

Next team handoff status:
- `COMPLETE` runtime/tool-assisted implementation is in review with focused Go proof.
- `COMPLETE` team/group retained-output implementation is in review with focused Go/UI proof.
- `IN_REVIEW` validation team added focused browser-visible component coverage for direct search/tool-assisted proof and Team Lead proof rendering.
- `NEXT` validation team should promote group broadcast proof visibility into live-backend browser coverage.

### Wave 2: Capability Clarity

Proceed with DE-4 after Wave 1 has a stable proof contract.

Rationale:
- Capability manifests should feed the same execution-summary and output contracts instead of becoming a separate registry feature.

Exit gate:
- Connected Tools can answer: what Soma can use, why, risk, approval, schema posture, availability, outputs, and result destination
- MCP server details remain inspectable but secondary

### Wave 3: Deployment Trust

Proceed with DE-5 when Wave 1 proof links are stable.

Rationale:
- Self-hosting credibility needs visible roots and proof status, but this should not distract from the primary Soma workflow.

Exit gate:
- System/Deployments exposes checkout/deployment/execution/artifact/log/cache roots, current commit, proof status, and recovery action

### Wave 4: Governance Polish And Cleanup

Proceed with DE-6 and DE-7 after Waves 1-3 identify duplicated surfaces.

Rationale:
- Cleanup should follow the proof model so the team removes duplicate noise instead of deleting useful evidence.

Exit gate:
- default UX no longer feels dashboard/admin-heavy
- advanced runtime detail remains reachable
- browser proof validates directed execution, capability clarity, governance, output durability, and deployment trust

## Dependencies

- DE-1 blocks DE-2/DE-3 UI consistency.
- DE-2 blocks durable output and audit proof claims.
- DE-4 depends on the Capability Manifest standard and should reuse its field vocabulary.
- DE-5 depends on the WSL/deployment directory model and proof status contract.
- DE-7 depends on observing real duplication after proof links land.

## Validation Matrix

| Proof | Required Coverage |
| --- | --- |
| Unit/typecheck | execution-summary types, output object normalization, capability mapping helpers. |
| API/runtime | run attachment, output references, audit/proposal state, capability manifest availability. |
| Browser mocked | Soma directed-execution summary, proof links, durable output cards, capability view. |
| Browser live backend | proposal/confirm/run proof, tool-assisted execution, team/group retained output, recovery. |
| Headed GUI | default path feels like directed execution, not dashboard operation. |
| WSL proof | deployment trust surface matches real checkout/deployment/execution/artifact/log/cache roots. |

## Orchestration Rules

- One execution story beats multiple panels.
- Do not add a new surface if an existing Soma, Runs, Resources, or System surface can carry the contract.
- Every slice must name which work package it advances.
- Every slice must report docs reviewed, docs changed, and validation evidence.
- A slice cannot claim release-ready if it lacks run proof, output proof, governance proof, and browser-visible proof for the touched behavior.
