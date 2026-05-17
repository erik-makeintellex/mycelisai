# V8.2 Finalization Delivery Plan
> Navigation: [Project README](../../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md) | [Current State PRD](V8_2_CURRENT_STATE_AND_FINALIZATION_PRD.md)

> Status: Canonical execution board
> Last Updated: 2026-05-17
> Purpose: Coordinate the teams that turn V8.2 architecture into visible execution, durable outputs, proof, recovery, and operator trust.

## Delivery Sequence

| Order | Slice | Primary teams | Exit proof |
| --- | --- | --- | --- |
| 1 | Refresh local live proof lane | Deployment And Proof, QA And Embodiment | Source-run Core/Interface plus infra-only Postgres/NATS are current; stale runtime defects are recorded precisely. |
| 2 | First demo package proof | Soma Experience, Runtime And Capability, Governance And Trust, QA | Playable browser game package returns `index.html`, `README.md`, validation notes, run/proof link, reload safety, and Groups output. |
| 3 | Durable trust object spine | Runtime And Capability, Governance And Trust | `ExecutionContract`, `ProofArtifact`, and `CapabilityManifestState` persist and link to runs, outputs, audit, recovery, and confidence-provenance fields. |
| 4 | Active team work truth | Runtime And Capability, Soma Experience, Governance And Trust | `TeamWorkItem`, `TeamInteraction`, `TeamStatusEvent`, and `TeamOutputRef` are durable API-backed state, not UI-only projections. |
| 5 | Single-window Soma workspace | Soma Experience, QA | Expression input, active work, output workbench, compact trust package, and scoped context fit into one coherent work surface. |
| 6 | Heavy-surface compression | Soma Experience, Deployment And Proof, QA | Teams, Resources, Runs, System, and Settings use bounded panes; raw topology sits behind Inspect/Advanced. |
| 7 | Scheduler/cadence productionization | Runtime And Capability, Governance And Trust, Soma Experience | Scheduled work is a governed rule with next-run, cooldown, approval posture, audit, proof, recovery, and last result. |
| 8 | Release-candidate embodiment pass | All teams | Replayable proof demonstrates ask -> proposal -> approval -> run -> output -> proof -> recovery/revisit in source lane, then deployment lane. |

## Team Assignments

| Team | Immediate work | First surfaces | Required evidence |
| --- | --- | --- | --- |
| Soma Experience | Reduce new-user complexity and frame expression around outcome, output shape, agentry posture, proof, and next action before topology. | `SomaOperatingSurface`, `MissionControlChat`, `ProposedActionBlock`, `ExecutionSummaryCard`, `ActiveWorkLane` | Component tests plus headed browser proof for cold start, proposal, result, degraded result, and mobile. |
| Runtime And Capability | Persist runtime truth objects while preserving additive `execution_summary` compatibility. | execution summary, confirm-action outputs, capability service/types, migrations | Go tests for contract/proof creation, failed proof, project-package output refs, capability refresh, and team work reload. |
| Governance And Trust | Normalize what happened, what is trusted, what failed, what can retry, and what needs operator attention. | proof status, audit details, team/group summaries, evidence/degraded UI | Failure tests must answer trusted remainder, invalid proof, retry scope, and operator action. |
| Deployment And Proof | Keep source proof repeatable and make `System -> Deployments` the detailed trust home. | system page, deployment trust handler, operations docs | Local source health proof before GUI; deployment/root proof when topology changes. |
| QA And Embodiment | Treat GUI as product proof, not route smoke testing. | first-demo, retry, teams, outputs, resources, settings, docs/runs specs | Desktop/mobile evidence for expression, approve, active work, output open, proof open, reload, retry, and surface compression. |
| Docs And State | Keep docs compressed and executable. | README, `.state/V8_DEV_STATE.md`, API/testing/user docs, docs manifest | Docs-link proof, diff review, docs changed/reviewed unchanged in close-out. |

## Architecture Execution Cell

The team works as one delivery cell:

- Runtime and Governance lead the next code slice because Soma, Team Work, and QA need stable `contract_id` and `proof_id` before they can stop inferring truth from cards.
- Team Work designs in parallel, but `create_team` alone means `new` or `briefed`; deliverable or delegated execution creates queued/running/output-ready/degraded work.
- Soma Experience may prototype the one-window frame against current projections, but release acceptance requires API-backed active work and durable proof links.
- Deployment And Proof proceeds on `System -> Deployments` without moving deployment context intake out of Resources or showing topology before roots, endpoint posture, proof lane, and recovery are clear.
- QA serializes browser proof. Mocked degraded/retry proof is useful, but final acceptance requires live degraded execution evidence.

## GUI Test Matrix

| Scenario | Required proof | Primary spec or action |
| --- | --- | --- |
| New-user cold start | No false "Soma just did this"; clear expression input and starter outcomes. | `soma-governance-live.spec.ts`, desktop/mobile screenshots |
| Expression framing | UI shows output/proof/agentry/next action before topology. | Soma component tests and first-demo Playwright assertions |
| First demo success | Retained game package opens and links output, storage, validation, proof, and Groups output. | `ui-finalization-browser-package-live.spec.ts` |
| First demo degraded/retry | Failure names trusted remainder, invalid proof, retry scope, operator attention, and safe next action. | `ui-finalization-browser-package-retry.spec.ts` |
| Active team work | Team created vs work queued/running/output-ready/degraded are distinct; control verbs are visible. | `teams.spec.ts`, `team-execution-live.spec.ts`, active-work API spec |
| Output workbench | Retained outputs are stronger than chat transcript and survive reload. | `team-output-content-live.spec.ts`, `resources-workspace-files.spec.ts` |
| Heavy surfaces | Teams, Resources, Runs, System, and Settings avoid long topology reading. | focused page specs plus screenshots |
| Mobile expression | Input/current state remains usable without chrome consuming the work surface. | `mobile-chromium` project |
| Deployment trust | `System -> Deployments` owns roots, commit, endpoint posture, proof lane, and recovery. | system focused Playwright |
| Re-entry | Refresh/reopen preserves retained output, proof links, and active/degraded state. | package live spec and workflow reload specs |

## Runtime Gates

1. Empty and existing migrations apply cleanly.
2. Proposal creation persists `execution_contracts` and exposes `contract_id`.
3. Confirmed success persists run, output refs, mission events, and a `proof_artifacts` row.
4. Failed or partial execution persists degraded proof with trusted remainder, invalid evidence, retryability, and operator-attention fields.
5. Capability refresh upserts long-lived `CapabilityManifestState`.
6. Team work items and interactions survive restart and drive Active Work Lane.
7. New runtime IDs are additive so existing `execution_summary` consumers keep working.

Suggested backend proof:

```powershell
cd core
go test ./internal/server ./internal/capabilities ./internal/runs ./internal/artifacts ./internal/swarm ./pkg/protocol -run "ExecutionContract|ProofArtifact|CapabilityManifest|TeamWorkItem|ConfirmAction|ExecutionSummary|ProjectPackage|Degradation" -count=1
```

## Orchestration Rules

- Do not start broad autonomy work until the first demo package proof is green.
- Do not call a team deliverable complete when only a team shell was created.
- Prefer existing Soma, Teams, Resources, Runs, System, or Settings surfaces over new UI surfaces.
- Do not expose MCP, NATS, provider, or deployment topology before output, proof, recovery, and next action are clear.
- Every slice records owning lane, touched surfaces, emitted events, output shape, proof shape, recovery behavior, validation, docs changed, docs reviewed unchanged, and remaining blocker.
- QA may fail release recommendation for practical browser usability even when unit/API gates are green.
