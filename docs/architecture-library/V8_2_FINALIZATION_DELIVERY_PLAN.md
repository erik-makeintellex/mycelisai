# V8.2 Finalization Delivery Plan
> Navigation: [Project README](../../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md) | [Current State PRD](V8_2_CURRENT_STATE_AND_FINALIZATION_PRD.md)

> Status: Canonical execution board
> Last Updated: 2026-05-20
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
| QA And Embodiment | Treat GUI as product proof, not route smoke testing. | auth/login, first-demo, retry, teams, outputs, resources, settings, docs/runs specs | Desktop/mobile evidence for login-to-Soma entry, expression, approve, active work, output open, proof open, reload, retry, and surface compression. |
| Docs And State | Keep docs compressed and executable. | README, `.state/V8_DEV_STATE.md`, API/testing/user docs, docs manifest | Docs-link proof, diff review, docs changed/reviewed unchanged in close-out. |

## Architecture Execution Cell

The team works as one delivery cell:

- Runtime and Governance lead the next code slice because Soma, Team Work, and QA need stable `contract_id` and `proof_id` before they can stop inferring truth from cards.
- Team Work designs in parallel, but `create_team` alone means `new` or `briefed`; deliverable or delegated execution creates queued/running/output-ready/degraded work.
- Soma Experience may prototype the one-window frame against current projections, but release acceptance requires API-backed active work and durable proof links.
- Deployment And Proof proceeds on `System -> Deployments` without moving deployment context intake out of Resources or showing topology before roots, endpoint posture, proof lane, and recovery are clear.
- QA serializes browser proof. Mocked degraded/retry proof is useful, but final acceptance requires live degraded execution evidence.

## Architecture-First UI Execution Gate

Do not proceed by only changing labels, attributes, clamps, or panel visibility. Tactical compression is useful only when it implements a shared structural contract.

The Soma workbench runtime spine is:

```text
ExpressionFrame
-> ExecutionContract
-> ActiveWorkItem / Run
-> OutputRef
-> ProofArtifact
-> TrustPackage
-> RecoveryAction
```

UI work must first decide which object owns truth, then project that truth into the screen. If a component has to infer runtime meaning from raw events, logs, or local string parsing, the slice should stop and add or normalize the API contract first.

### Structural Contracts Before Component Work

| Contract | Owner | Why it gates UI changes | First acceptance |
| --- | --- | --- | --- |
| `ExpressionFrame` | Runtime/Capability with Soma Experience | Gives the workbench a stable source for outcome, output shape, audience/use, constraints, agentry posture, proof, continuation, and risk. | `/api/v1/chat` and proposal responses include a normalized frame or a documented absence reason. |
| `SomaStateCard` model | Soma Experience with Governance/Trust | Prevents every surface from inventing direct/proposal/running/output/degraded/blocker/recovery rendering. | Dashboard, Organization workspace, Teams active work, and Runs use the same state vocabulary. |
| Default/Inspect taxonomy | Soma Experience with QA | Prevents inconsistent "Show details", "Inspect payload", "Advanced", and drawer behavior. | Default shows state/output/proof/next action; Inspect shows raw payloads, bus topics, prompts, full refs, and event history. |
| `TrustPackage` read model | Runtime/Capability with Governance/Trust | Stops the UI from stitching contracts, run events, team work, artifacts, and proof rows by itself. | One API projection returns output refs, proof refs, audit refs, degradation, confidence provenance fields, and recovery actions. |
| `OutputRef` standard | Runtime/Capability with Soma Experience | Makes files, project packages, media, MCP results, team outputs, and deployment proof render consistently. | Shared projection fields exist for title, kind, href/storage, entrypoint, validation, proof, run/team/work links, and retention. |
| `ActiveWorkItem` projection | Runtime/Capability with Team Work | Keeps team work, capability work, scheduled work, reviews, and deployment checks in one active-work lane. | Active Work can render team work now and add schedule/capability work without new UI grammar. |
| `RecoveryAction` standard | Governance/Trust with Runtime/Capability | Makes retry/recover/continue/request-input/pause/archive/rerun auditable and stable. | Each recovery action names source states, allowed target states, audit fields, and proof impact. |

### Team Execution Order

| Order | Team | Work | Must produce before next team depends on it |
| --- | --- | --- | --- |
| 1 | Runtime/Capability | Define and populate `ExpressionFrame`, `OutputRef`, `TrustPackage`, and broader `ActiveWorkItem` read projections. | Go/API tests plus additive payload compatibility for existing UI consumers. |
| 2 | Governance/Trust | Standardize `SomaStateCard`, degraded lifecycle labels, confidence provenance fields, and `RecoveryAction` semantics. | Component fixtures and backend tests that answer what failed, trusted remainder, invalid proof, and safe next action. |
| 3 | Soma Experience | Compose the workbench around Expression, Active Work, Output Workbench, and Trust Package using the shared contracts. | Component proof that default state is readable without raw topology. |
| 4 | Team Work/Council | Bind team steering, asking, review, and recovery into `ActiveWorkItem` without promoting rosters as the center. | Live or mocked proof for ask/respond/recover and Council review lineage. |
| 5 | QA/Embodiment | Certify workflows before visual polish: cold start, direct answer, proposal, active work, output, proof, failure/recovery, inspect. | Desktop/mobile browser proof with no horizontal overflow and no default raw dumps. |
| 6 | Docs/State | Record only the accepted contract, proof, blockers, and changed docs. | State evidence plus docs-link proof; no parallel execution-board drift. |

### Stop Conditions

Stop implementation and return to architecture when:

- a UI surface must parse raw JSON to know what happened
- a component invents a new response state outside the shared model
- proof, audit, output, or recovery refs are capped without a visible Inspect path
- a team, run, resource, or activity screen becomes a competing primary surface instead of a scoped context
- browser proof only checks layout and not whether a new operator can explain the work state
- schedule, team, or capability UI claims actionability before the backend can persist proof and recovery

### Workflow Proof Before Polish

Each architecture-first UI slice must include at least one workflow proof:

1. Cold start to first useful work.
2. Direct answer versus team-managed output.
3. Governed proposal approval or cancellation.
4. Active Work ask/respond/recover.
5. Output Workbench retained artifact review.
6. Failure/degraded recovery.
7. Inspect detail for raw payload/proof without default overload.

Only after those pass should the team tune labels, spacing, line clamps, scroll heights, or secondary visual density.

## Current UI Expression And Function Alignment

Current review shows the UI has the right vocabulary before it has the full shared functional spine. `SomaWorkspaceFrame` already expresses the target regions: Expression, Active Work, Trust, Output, and Context. The next work is to make those regions read from stable contracts instead of component-local projections.

| Surface | Current UI expression | Functional backing today | Alignment gap | Begin with |
| --- | --- | --- | --- | --- |
| Authenticated entry | `/login` routes every edition into the signed-in Soma environment and the dashboard confirms role/provider/scope before work begins. | Web session cookie, `/auth/session`, local owner login, Google Workspace OIDC, signed Interface-to-Core web identity headers. | The identity boundary, Core audit identity propagation, and normalized `actor_identity` audit proof are wired; live governed-action proof is in regression. | Keep the post-login dashboard contract in `homepage.spec.ts`; keep `soma-governance-live.spec.ts` Scenario E as the audit-identity proof gate. |
| Dashboard Soma | One dominant Soma surface with Expression, Active Work, Trust, Output, and Context slots. | Chat store, latest assistant message, durable team-work hook, execution summary helpers. | The UI says "Expression" but does not yet consume a canonical `ExpressionFrame`; Trust and Output are stitched from chat summaries plus team refs. | Add `ExpressionFrame` projection to chat/proposal responses and render it in the Expression slot before topology. |
| Mission chat | Natural-language Soma input, proposal/result bubbles, simple mode, optional advanced routing. | `/api/v1/chat`, proposal payloads, `execution_summary`, local chat state. | Chat transcript remains the dominant functional object; response state, output, and proof are not yet projected through one `SomaStateCard` model. | Build shared state-card fixtures from existing `ui_response_state` and execution summary data. |
| Active Work | Compact durable team work with Ask/Respond and output/proof refs; Soma home shows an attention-first recent slice and links the full backlog to Teams. | `TeamWorkItem`, `TeamStatusEvent`, `TeamInteraction`, `TeamOutputRef`. | Strongest aligned area, but it only represents team work. It cannot yet show capability, schedule, review, or deployment work without new local grammar. | Create `ActiveWorkItem` read projection using team work as the first source. |
| Output Workbench | Shows retained outputs, project packages, media preview, open/storage/copy controls. | `ExecutionOutput`, chat artifacts, `TeamOutputRef` conversion helpers. | Output refs are normalized in UI helpers, not one API contract; package files/proof refs need consistent caps/Inspect. | Standardize `OutputRef` on the server and map current `ExecutionOutput`/`TeamOutputRef` into it. |
| Trust Package | Shows outcome, outputs, proof, next step, and expandable details. | `execution_summary`, run id, proof links, audit recovery, local summary parsing. | Trust is visually good but functionally assembled in the component; no single `TrustPackage` read model exists. | Add server-side `TrustPackage` projection over contract, proof artifact, outputs, audit, degradation, and recovery. |
| Teams | Team page emphasizes Active Work plus roster/detail drawer. | Durable team work APIs plus team detail API. | Work lane is aligned; drawer still exposes team topology as a support surface and prompts remain default API data. | Keep roster summarized; move prompt/detail fetch behind explicit Inspect or privileged include flag. |
| Runs | Conversation/events tabs provide proof history and raw event inspection. | Run events and conversation endpoints. | Run page still behaves like transcript/log review before proof package review. | Add proof-first run summary from `TrustPackage`, then keep events/conversation as Inspect tabs. |
| Activity | Advanced operational pulse with recent runs, run inspector, live signals. | Recent runs, run timeline, SSE stream, service status. | Compression is aligned, but Activity still exists as an advanced diagnostics surface rather than a typed operations summary. | Add typed activity categories from normalized status/output/governance/error events. |
| Resources | Menu/detail advanced resource panes answer "what Soma can use." | MCP registry, workspace explorer, exchange inspector, deployment context, engines, roles. | Strong UI shape, but capability health/risk/approval/recovery needs to be centered as the common functional contract. | Promote `CapabilityManifestState` as the primary Connected Tools data model. |
| System | Deployment/services/comms health as admin support. | Service status, deployment trust endpoint, native/compose/k8s proof lanes. | System is correctly scoped, but recovery actions are not yet governed objects. | Map service degradation to `RecoveryAction` records where operator action is available. |

### Function-Matched Start Slices

1. **ExpressionFrame read projection**
   - Owns: Runtime/Capability plus Soma Experience.
   - Implements: protocol type, chat/proposal population, Interface type, Expression slot renderer.
   - Proves: asking for a direct answer and a governed output shows outcome, output shape, proof expectation, agentry posture, and continuation.

2. **SomaStateCard model**
   - Owns: Governance/Trust plus Soma Experience.
   - Implements: shared UI model derived from `ui_response_state`, execution status, degradation, output refs, proof refs, and next action.
   - Proves: direct answer, proposal, running, output-ready, degraded, blocker, and recovery fixtures render the same state grammar.

3. **OutputRef standard**
   - Owns: Runtime/Capability.
   - Implements: server projection helper for `ExecutionOutput`, `TeamOutputRef`, artifact refs, MCP results, media, project packages, and deployment proof.
   - Proves: Output Workbench renders all output kinds without bespoke component parsing.

4. **TrustPackage read model**
   - Owns: Runtime/Capability plus Governance/Trust.
   - Implements: API projection from `ExecutionContractRecord`, `ProofArtifactRecord`, run refs, output refs, audit refs, confidence provenance, degradation, and recovery actions.
   - Proves: dashboard trust slot and run page proof summary use the same payload.

5. **ActiveWorkItem generalization**
   - Owns: Runtime/Capability plus Team Work/Council.
   - Implements: read model that includes team work first, then scheduled, capability, review, and deployment work.
   - Proves: existing team Active Work continues to pass while schedule/capability placeholders require no new UI grammar.

6. **RecoveryAction standard**
   - Owns: Governance/Trust.
   - Implements: typed retry, recover, continue partial, request input, pause, archive, and rerun contract actions with source states and audit requirements.
   - Proves: degraded team ask and failed proposal execution expose stable recovery actions instead of free-text next steps.

## GUI Test Matrix

| Scenario | Required proof | Primary spec or action |
| --- | --- | --- |
| Login to Soma environment | Fresh browser requires login; successful local/SSO session lands on `/dashboard` with signed-in role/provider/scope and Soma as the first work surface. | `homepage.spec.ts`, `CentralSomaHome.test.tsx` |
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
