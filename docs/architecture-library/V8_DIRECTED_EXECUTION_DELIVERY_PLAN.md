# V8 Directed Execution Delivery Plan
> Navigation: [Project README](../../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)

> Status: ACTIVE
> Last Updated: 2026-05-14
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
| Runtime/Run Spine | IN_REVIEW | Run attachment, execution identity, output references, recovery and degradation metadata. | Wave 1 trust-spine contract is proven; failed approved execution and search blockers now carry bounded degradation proof. |
| Output/Exchange | ACTIVE | Durable output object mapping, retained artifacts, exchange normalization. | Run/Output hardening across chat, tools, teams, artifacts, proof links, and recovery metadata. |
| Capability/MCP | ACTIVE | Capability manifest registry, MCP-to-capability mapping, health/fallback. | Capability list that answers "what Soma can use" with risk, schema, approval, output destinations. |
| Governance/Trust | IN_REVIEW | Proposal/approval/audit language, risk policy, proof semantics, degraded-trust boundaries. | Default-safe proof copy plus advanced audit detail mapping. |
| Interface/Soma UX | IN_REVIEW | Default Soma execution summary, proof links, output cards, next-step and degradation display. | Soma response surface showing intent, understanding, execution shape, capability/team use, outputs, proof, recovery, degradation, and next step. |
| Interface/Advanced UX | ACTIVE | Connected Tools capability view, Runs proof view, System/Deployments trust view. | GUI directed-execution surfaces that reveal manifests, run traces, deployment roots, and recovery without polluting default UX. |
| Validation | ACTIVE | Browser-visible proof, unit/API contract coverage, failure/recovery checks. | Keep Wave 1 final-review evidence green while proving the next capability/output/GUI lanes. |
| Docs/In-App Docs | ACTIVE | Architecture/user/state/docs manifest alignment. | Keep canonical docs and in-app docs synced for each slice. |

## Active Embodiment Cell

This is the current cross-team operating cell. It is an execution overlay, not a new doctrine layer.

| Lane | Active Team | Architecture Guidance | Execution Target | Proof Gate |
| --- | --- | --- | --- | --- |
| Soma experience | Interface/Soma UX + Governance/Trust | Keep Soma singular; compress default UX into intent, visible result, proof/recovery. | Verify Operator trust package, proposal lifecycle, recovery/degradation, and retained outputs in the default Soma path. | Focused Vitest plus headed browser proof for Soma governance. |
| Runtime and capability | Runtime/Run Spine + Capability/MCP + Output/Exchange | Capabilities are governed runtime objects; outputs are durable product objects. | Redeploy current HEAD to Rancher K3s and prove run-linked execution, retained artifacts, search/capability posture, and degradation metadata. | `go test ./... -count=1 -p 1`, K3s deploy/wait/bridge, lifecycle health. |
| Governance and trust | Governance/Trust + Runtime/Run Spine | Proposal -> confirm -> execute is durable infrastructure, not UI decoration. | Prove successful and failed approved execution carry run/proof/audit/recovery boundaries. | API/server tests plus live proposal/confirm/run proof. |
| Deployment and proof | Interface/Advanced UX + Validation | Deployment roots and runtime health must become understandable proof, not topology noise. | Prove current image/commit in Rancher K3s, Core bridge, storage/PVC root, AI/search posture, and workspace output review. | K3s status/deploy/wait/bridge plus workspace live-backend proof. |
| QA and embodiment | Validation + Orchestration Lead | Acceptance asks whether users can trust visible execution and recovery. | Run the canonical MVP path: Soma intent -> proposal -> approval -> run -> retained output -> proof/revisitability. | Headed Chromium live specs for governance, team execution, groups, and workspace. |

Immediate orchestration rule: the next broadening slice is deferred until current HEAD is live-proven on the Rancher K3s lane.

## Work Packages

| ID | Priority | Owner | Work Package | Acceptance |
| --- | --- | --- | --- | --- |
| DE-1 | 1 | Runtime/Run Spine + Interface/Soma UX | Define and consume a reusable execution-summary contract. | Soma can display intent, understanding, execution shape, capability/team use, outputs, proof, and next step from one normalized object. |
| DE-2 | 2 | Runtime/Run Spine + Output/Exchange | Ensure meaningful actions attach to runs and meaningful outputs attach to durable output objects. | Direct retained answers, proposals, tool use, team execution, automations, and plugins all have run/output proof or explicit non-retained classification. |
| DE-3 | 3 | Interface/Soma UX + Governance/Trust | Put proof links and proposal/audit/recovery/degradation state next to meaningful outputs. | User can reach run proof, approval/audit evidence, recovery state, and degraded-trust boundaries from the post-execution workflow. |
| DE-4 | 4 | Capability/MCP + Interface/Advanced UX | Reshape Connected Tools around governed capabilities rather than server inventory. | Resources shows what Soma can use, availability, risk, approval posture, schemas, output types, and destinations; raw MCP details sit behind disclosure. |
| DE-5 | 5 | Interface/Advanced UX + Runtime/Deployment | Add deployment trust visibility. | System/Deployments shows checkout, deployment root, execution root, artifact/log/cache roots, current commit, proof status, and recovery action. |
| DE-6 | 6 | Governance/Trust + Validation | Normalize proposal/proof/recovery/degradation language and test it. | Proposal never mutates before approval; success always has proof; failures answer what failed, what remains trusted, what proof is invalid, and what can be retried. |
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
- `IN_REVIEW` direct Soma answers, guided proposals, and confirmed proposal execution share the additive `execution_summary` runtime payload.
- `IN_REVIEW` the Soma chat surface renders an Operator trust package with intent, understanding, execution shape/status, capability use, outputs, proof, audit/recovery, degradation, and next step from that payload.
- `IN_REVIEW` tool-assisted chat/search preserves read-only `tools_used`, classifies tool/artifact responses as `tool_assisted_work`, adds `execution_summary` to direct `web_search`, and marks blocked search as `blocked` with `audit_recovery.degradation` metadata.
- `IN_REVIEW` confirmed proposal execution failures return failed run/proof/audit data and `audit_recovery.degradation` instead of only a flat API error, so the UI can show what failed, what remains trusted, what proof is invalid, and what can be retried.
- `IN_REVIEW` Team Lead guidance and group broadcast accepted responses attach `execution_summary` without fabricating `run_id`; group/team proof stays audit/group scoped until a concrete run exists.
- `IN_REVIEW` focused browser-visible component proof covers direct tool-assisted Soma search summaries and Team Lead execution summaries, including the no-fabricated-run-link boundary.
- `IN_REVIEW` mocked Chromium browser proof covers Groups broadcast execution-summary/audit-proof visibility after `POST /api/v1/groups/{id}/broadcast`.
- `IN_REVIEW` Wave 1 is proven and in final review with live-backend/WSL release evidence from the dedicated `mycelis-root` proof lane recorded by state commit `f332c680cc6eec285da018dc48c9760dd15cb4e7`.

Wave 1 validation evidence:
- `go test ./pkg/protocol ./internal/server -run "TestChatResponsePayload_ExecutionSummaryIsAdditive|TestHandleChat_UnwrapsReadableJSONEnvelopeFromAgent|TestHandleChat_RoutesLatestMutationTurnToProposalAcrossThreadHistory|TestHandleConfirmAction_CompletesVerifiedExecutionWithPlannedToolCalls|TestHandleConfirmAction_NormalizesWriteFileAliasesInStoredPlan" -count=1 -v`
- `go test ./pkg/protocol ./internal/server -run 'ExecutionSummary|Search|ConfirmAction' -count=1`
- `cd interface; npx tsc --noEmit`
- `cd interface; npx vitest run __tests__/dashboard/MissionControlChat.outputs.test.tsx __tests__/dashboard/MissionControlChat.executionSummary.test.tsx __tests__/store/useCortexStore.confirm-proposal.failure.test.ts`
- `uv run pytest tests/test_docs_links.py tests/test_documentation_layout_contract.py -q`
- `uv run inv quality.max-lines --limit 300`
- `git diff --check`
- `uv run inv wsl.refresh`
- `uv run inv wsl.validate --lane=release`

Wave 1 final-review status:
- `COMPLETE` runtime/tool-assisted implementation is in review with focused Go proof.
- `COMPLETE` team/group retained-output implementation is in review with focused Go/UI proof.
- `COMPLETE` validation team added focused browser-visible component coverage for direct search/tool-assisted proof and Team Lead proof rendering.
- `COMPLETE` validation team added focused mocked Chromium coverage for group broadcast execution-summary proof visibility.
- `IN_REVIEW` dedicated `mycelis-root` WSL Compose proof is green as historical deployment-mimic evidence, while Rancher Desktop K3s is now the Windows local Kubernetes release-parity proof lane.

Next active handoff:
- `IN_REVIEW` Capability/MCP first slice is implemented locally: `/api/v1/capabilities` exposes a derived manifest snapshot from exchange seed capabilities, MCP registry/library entries, Mycelis Search, internal tools, and host command allowlists; persistence schema exists but refresh results are still in-memory.
- `IN_REVIEW` Runtime/Run Spine and Output/Exchange first slice is implemented locally: execution summaries now expose explicit run/proof classes, no-run reasons, exchange item proof, and output retention classes; runtime-state chat, Exchange/artifact normalization, and direct MCP tool-call responses now classify proof more explicitly.
- `IN_REVIEW` Interface/Soma UX and Interface/Advanced UX first slice is implemented locally: the default Soma surface renders a causal directed-execution package, and Connected Tools now prioritizes capability visibility while preserving MCP server drill-down.
- `ACTIVE` Validation should promote the local focused proof into broader managed build/browser checks, then WSL release proof after commit/push.

### Wave 2: Capability Clarity

Proceed with DE-4 now that Wave 1 has a stable proof contract.

Rationale:
- Capability manifests should feed the same execution-summary and output contracts instead of becoming a separate registry feature.
- This is the first active post-Wave-1 lane.

Exit gate:
- Connected Tools can answer: what Soma can use, why, risk, approval, schema posture, availability, outputs, and result destination
- MCP server details remain inspectable but secondary

Current implementation status:
- `IN_REVIEW` backend capability manifests are available through `GET /api/v1/capabilities`, `GET /api/v1/capabilities/{id}`, and `POST /api/v1/capabilities/refresh`.
- `IN_REVIEW` Connected Tools consumes the capability API when available and falls back to existing MCP/search data when it is not.
- `NEXT` persist refreshed manifest rows into `capability_manifests` and reconcile the runtime snapshot with long-lived deployment health/probe state.

### Wave 3: Deployment Trust

Proceed with DE-5 from the dedicated `mycelis-root` proof model while Run/Output hardening continues.

Rationale:
- Self-hosting credibility needs visible roots and proof status, but this should not distract from the primary Soma workflow.

Exit gate:
- System/Deployments exposes checkout/deployment/execution/artifact/log/cache roots, current commit, proof status, and recovery action

Current implementation status:
- `NEXT` deployment trust is still the next explicit UI/runtime lane; current work prepared proof metadata and capability/run surfaces, but did not add the System/Deployments root-visibility surface.

### Wave 4: Governance Polish And Cleanup

Proceed with DE-6 and DE-7 alongside the active GUI directed-execution cleanup once duplicated surfaces are tied back to the proof model.

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
