# V8.2 Current State And Finalization PRD
> Navigation: [Project README](../../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md) | [V8.2 Full Production Architecture](../../architecture/v8-2.md) | [V8 Development State](../../.state/V8_DEV_STATE.md)

> Status: Canonical working PRD
> Last Updated: 2026-05-16
> Module Boundary: Soma experience, runtime/capability, governance/trust, deployment/proof, QA/embodiment, docs/state
> Purpose: Give the architecture team one product-ready view of where Mycelis is now, what finalization means, and which concrete runtime objects, UI states, recovery semantics, and proof workflows must close before broader expansion.

## PRD Summary

Mycelis V8.2 is no longer primarily blocked by architecture definition. The system now has enough product, runtime, governance, capability, deployment, and testing structure to enter finalization around concretization and operational embodiment.

Finalization means:

```text
I tell Soma what I want.
Soma directs the work safely.
I can see what happened.
I can trust the result later.
```

The architecture team's job is to make that workflow coherent, repeatable, recoverable, deployable, and easy to understand by implementing the [V8.2 Finalization Concretization Contract](V8_2_FINALIZATION_CONCRETIZATION_CONTRACT.md) without widening doctrine or exposing internal topology as the default experience.

## Product Definition

Mycelis is a Soma-centered governed cognitive operating environment.

It is not a chatbot, agent console, MCP registry, dashboard pile, or orchestration demo. Runs, teams, groups, memory, exchange, capabilities, MCP, automations, proof, logs, deployment roots, and health checks are scoped operational contexts that support Soma-directed work.

Operator expression is therefore not treated as a plain user query. It is treated as a request to shape governed work: desired output, audience/use, constraints, agentry posture, acceptance proof, and continuation. Soma owns the translation from expression into an ExecutionContract, active work state, durable output, and proof. Internal topology appears only when it helps the operator understand authority, trust, recovery, or deployment reality.

## Current State

The current implementation state is recorded in `.state/V8_DEV_STATE.md`. As of 2026-05-16, the active state is:

| Area | State | What exists now |
| --- | --- | --- |
| V8.2 architecture | `ACTIVE` | Canonical full production target, operational embodiment directive, runtime contracts, UI contracts, capability manifest standard, and directed-execution delivery plan. |
| Deployment proof | `COMPLETE` / `IN_REVIEW` | Rancher Desktop K3s RC proof lane is green; Compose remains the fast local proof lane; live Compose is healthy with explicit external AI endpoint posture. |
| Soma directed execution | `IN_REVIEW` | Soma can return direct answers, governed proposals, confirmed execution summaries, retained outputs, proof links, audit/recovery metadata, and degraded-state summaries. |
| Governance | `IN_REVIEW` | Proposal -> confirm -> execute is implemented for meaningful mutations; failed approved execution can return run/proof/audit/degradation evidence. |
| Outputs | `IN_REVIEW` | Retained output cards, workspace-backed files, media previews, generated project-package shape, proof/open/folder controls, and run conversation traces exist. |
| Capabilities and MCP | `ACTIVE` / `IN_REVIEW` | Connected Tools exposes MCP/library/search posture and derived capability manifests; direct MCP tool refs can execute through confirmed plans. |
| Managed Exchange | `ACTIVE` | Channels, threads, schemas, items, learning candidates, MCP activity, and normalized output lanes are implemented as inspectable advanced surfaces. |
| Search | `IN_REVIEW` | Mycelis-owned search supports local sources, local API, SearXNG, optional Brave, disclosure metadata, and search-source provenance in Soma results. |
| Memory and context | `ACTIVE` / `IN_REVIEW` | Semantic memory, governed deployment context, user-private context, company knowledge, Soma operating context, reflection/synthesis lanes, and candidate-first learning contracts exist. |
| Teams and groups | `ACTIVE` / `IN_REVIEW` | Runtime teams, groups, team outputs, create-team proposals, group mirroring, retained team/file outputs, and an Active Work Lane projection exist; raw topology is inspectable on demand instead of default UX. |
| Automations and scheduler | `ACTIVE` / `IN_REVIEW` | Event trigger rules and review-loop scheduler health are visible; operator cadence authoring is not finalized and must ship as governed scheduler rules before schedule language returns to default UX. |
| UI compression | `IN_REVIEW` | Soma proposal/trust UI, Auth Providers, and Resources have been compressed into focused operator surfaces with details behind disclosure or menu selections. |
| Identity and auth | `ACTIVE` / `IN_REVIEW` | Local owner and deploy-owned access posture are surfaced read-only; identity/auth schema foundation exists; enterprise provider setup remains a review-only configuration contract until adapters are enabled. |
| System health | `IN_REVIEW` | `/system` distinguishes scheduler health, services status, comms optional-provider posture, group bus, NATS, Postgres, cognitive, and runtime checks. |
| QA and proof | `ACTIVE` | Focused unit, typecheck, build, docs, max-lines, Playwright, live backend, headed GUI, K3s, and Compose proof lanes exist. |

## Finalization Target

The finalization target is not "finish every V8.2 idea." It is to make the core operating loop feel production-real:

```text
Intent
-> Soma understanding
-> Execution contract
-> Governed run
-> Capability, team, or automation execution
-> Durable output
-> Proof, audit, recovery, and revisitability
```

The target user should be able to:

1. Ask Soma for meaningful work.
2. Understand whether Soma answered directly, proposed governed work, delegated to a team, used a capability, or blocked safely.
3. Approve protected work before mutation.
4. Watch or revisit the run.
5. Open durable outputs and their storage/proof.
6. See what failed when execution degrades.
7. Retry, recover, or review without reading logs first.
8. Trust the deployment and capability boundary.

## Canonical MVP Workflow

This workflow is the first finalization slice. Other work is secondary until it is smooth, replayable, and trustworthy.

```text
Operator asks Soma to create a playable browser game package with a README.
Soma creates an ExecutionContract, requests approval when mutation is needed,
executes the file/artifact capability, returns a retained project-package output,
opens proof, and shows retry/degraded trust behavior when execution fails.
```

## Concretization Objects

| Object | Status | Required concrete fields |
| --- | --- | --- |
| ExecutionContract | `REQUIRED` | `contract_id`, intent summary, execution shape, capability requirements, governance posture, approval requirements, expected outputs, expected proof, recovery posture, degradation behavior, run linkage, timestamps, version |
| ProofArtifact | `REQUIRED` | `proof_id`, `run_id`, execution status, evidence refs, output refs, validation source, proof quality, degradation state, recovery options, confidence provenance fields, audit refs, timestamps |
| CapabilityManifestState | `REQUIRED` | capability id, health, probe status, trust/risk class, approval posture, allowed roles, output schemas, failure posture, recovery posture, manifest version |
| UIResponseState | `IN_REVIEW` | direct answer, proposal, execution result, blocker, recovery state, degraded execution, awaiting approval, retry required, partial completion |
| TeamInteractionSurface | `ACTIVE` | `TeamInteraction`, `TeamWorkItem`, `TeamStatusEvent`, and `TeamOutputRef` ids, team/run/work linkage, source, verb, objective, state, expected outputs/proof, operator need, degradation, recovery, output refs, proof refs, audit refs, version |

These are runtime trust objects, not presentation conveniences. Exact schema fields, UI states, degraded lifecycle states, deployment-trust ownership, and the required slice close-out template are canonicalized in the [V8.2 Finalization Concretization Contract](V8_2_FINALIZATION_CONCRETIZATION_CONTRACT.md).

## Degraded Execution Semantics

Degraded execution is a first-class state. Every degraded result must identify what succeeded, what failed, what proof is invalid or incomplete, what remains trusted, what can continue safely, and what requires rerun, retry, or operator attention.

Minimal confidence provenance starts here through validation source, evidence strength, cross-check status, proof quality, review lineage, and disagreement markers. Do not add scoring until these fields have evidence semantics.

## Primary Personas

| Persona | Needs |
| --- | --- |
| Owner/operator | Ask Soma for work, approve actions, inspect proof, recover from failure, trust local deployment. |
| Architecture team | Keep target, runtime contracts, UI contracts, capability posture, and proof gates aligned without doctrine creep. |
| Runtime engineer | Preserve run/output/capability/audit contracts across chat, teams, tools, automations, and deployments. |
| UX engineer | Compress operator complexity while keeping inspectable proof and advanced surfaces reachable. |
| Validation/release engineer | Prove the canonical workflow in browser and deployment lanes with repeatable evidence. |
| Enterprise reviewer | See identity, auth, deployment, audit, capability, and governance readiness without false claims that planned adapters are shipped. |

## Product Requirements

### P0 - Soma Operating Loop

- One dominant Soma operating surface across dashboard and organization workspace.
- Soma must translate operator expression into an output contract before exposing raw topology. The default UI should frame outcome, output shape, audience/use, constraints, agentry posture, proof, and next action in operator language.
- Every meaningful response classifies as direct answer, proposal, execution result, blocker, recovery state, degraded execution, awaiting approval, retry required, or partial completion.
- Proposal cards must summarize the decision first, then expose lifecycle, risk, capability, team, bus, and proof details behind disclosure.
- The Operator trust package must show outcome, outputs, proof, next step, and degradation state without requiring raw runtime vocabulary.
- Broad asks should default to compact team shaping and deliberate expansion guidance, not automatic large-team sprawl.
- The primary work window should keep expression, active work, output preview, and trust summary together. Teams, Resources, Runs, and System details should use focused list/detail or menu/detail panes instead of long scrolling topology pages.

### P0 - Runs, Outputs, And Proof

- Every meaningful action creates or attaches to a run unless explicitly classified as non-retained.
- Every meaningful output becomes a durable product object: text answer, plan, review, file, media, generated project package, MCP result, audit event, deployment proof, or learning candidate.
- Output objects must be inspectable, attributable, reconstructable, reviewable, and downloadable/openable where appropriate.
- Run proof must connect intent, approval, capability/team use, output, audit, recovery, and next step.
- Failed approved execution must preserve what failed, what remains trusted, invalidated proof, safe continuation, and retry/attention requirements.

### P0 - Governance And Capability Boundary

- Proposal -> confirm -> execute remains durable runtime infrastructure, not UI decoration.
- No hidden mutation, silent capability use, raw secret exposure, or unverified completion.
- Capabilities must answer "what Soma can use" through manifests with purpose, source, risk, approval, allowed roles, schemas, outputs, health, audit, and fallback behavior.
- MCP servers are implementation detail beneath governed capability visibility.
- Direct MCP tool refs in confirmed plans must execute through registered MCP executors and return retained proof when meaningful.

### P0 - Deployment Trust

- Compose remains rapid local proof; Kubernetes/Helm remains scale-up and enterprise-aligned deployment proof.
- External AI engines must be explicit reachable endpoints, not container-local loopback assumptions.
- System surfaces must show service health, scheduler state, comms state, search posture, workspace root behavior, and recovery actions.
- `System -> Deployments` owns deployment root, execution root, workspace/artifact root, current commit, endpoint posture, runtime posture, proof lane, and recovery posture. Summary state may appear elsewhere, but detailed deployment trust and recovery belongs there.

### P1 - Teams, Groups, And Active Work

- Teams and groups must be smooth to create, inspect, steer, and review while active.
- Soma, Council, operators, and teams must use the [V8.2 Soma Team Interaction Contract](V8_2_SOMA_TEAM_INTERACTION_CONTRACT.md) for active-team verbs, work-item state, output refs, and recovery semantics.
- Team output must be visible as retained product state, not just chat transcript.
- Operators must be able to work with active teams, inspect team run state, review output, and cleanly stop temporary teams.
- Runtime-team delivery collaborator claims remain blocked until bounded role-specific team asks reliably return within product timeout and expose degradation clearly.

### P1 - Automations And Cadence

- Event trigger rules remain the current production automation surface.
- Scheduled/cadence authoring must ship as a governed scheduler rule type with propose/execute, cooldown, audit, recovery, proof, and next-run visibility.
- Schedule language must not return to default UI until this is production behavior rather than roadmap copy.

### P1 - Memory, Learning, And Context

- Team-shared execution memory stays separate from deployment context, user-private context, company knowledge, Soma operating context, and reflection/synthesis.
- Learning enters as candidates with classification, confidence, review posture, evidence refs, and promotion target before durable memory mutation.
- Confidence provenance should be prepared through validation source, evidence strength, cross-checks, proof quality, and review lineage, without overbuilding scores too early.

### P1 - Identity And Enterprise Readiness

- Local owner, break-glass, deploy-owned access posture, and principal metadata remain explicit and safe.
- Enterprise providers are configuration contracts until adapters are runtime-enabled.
- External identity may prove user identity, but Mycelis remains source of authority for roles, organization membership, Soma access, capability permissions, approvals, and audit.

## Architecture Team Workstreams

| Team | Status | Finalization Assignment | Exit Gate |
| --- | --- | --- | --- |
| Soma Experience | `ACTIVE` | Continue compressing default UX into intent, visible result, proof, and recovery. | Canonical MVP workflow passes headed browser proof and first-run review without conceptual overload. |
| Runtime And Capability | `ACTIVE` | Persist capability manifest refresh state as P0 trust infrastructure, harden MCP-backed confirmation proof, normalize project-package output. | Capabilities, tools, teams, automations, and direct retained outputs all share run/output/proof contracts. |
| Governance And Trust | `IN_REVIEW` | Normalize success, failure, blocker, and degraded-proof language across all execution paths. | Users can answer what happened, what is trusted, what failed, and what can be retried from the UI. |
| Deployment And Proof | `ACTIVE` | Promote deployment trust visibility for roots, current commit, endpoint posture, workspace storage, proof lane, and recovery. | System/Deployments trust surface proves self-hosted reality without default topology noise. |
| QA And Embodiment | `ACTIVE` | Keep focused live browser proof for Soma, teams, groups, Resources, Runs, System, and recovery. | One repeatable proof set demonstrates create/review work -> approval -> run -> output -> proof -> revisit. |
| Docs And State | `ACTIVE` | Keep canonical docs compressed and state operational. | Architecture docs, user docs, API docs, state, and in-app manifest agree after every slice. |

## Finalization Roadmap

### Phase 1 - Close The MVP Trust Loop

- Define and persist the initial ExecutionContract shape.
- Define and persist the initial ProofArtifact shape.
- Prove project-package output through live Mycelis GUI.
- Prove explicit MCP-backed confirmation path end to end.
- Ensure retained outputs open, reveal storage, and link to proof.
- Ensure failed execution returns actionable degradation metadata.

### Phase 2 - Active Team And Output Workflows

- Make active teams easier to inspect, steer, and clean up through the Active Work Lane before raw topology is opened.
- Implement `TeamInteraction`, `TeamWorkItem`, `TeamStatusEvent`, and `TeamOutputRef` persistence/API/UI projections from the Soma Team Interaction Contract.
- Add stronger output views for team deliverables and retained packages.
- Show team communication state without exposing raw NATS as the main workflow.
- Keep broad team creation deliberate and compact by default.

### Phase 3 - Scheduler/Cadence Productionization

- Add governed schedule rule authoring.
- Add next-run state, cooldown, audit, recovery, and proof.
- Add browser proof for scheduled actuation.
- Remove any remaining schedule-copy that implies unshipped behavior.

### Phase 4 - Deployment Trust And Recovery

- Add `System -> Deployments` trust visibility for root paths, commit, image/chart, workspace/artifact roots, health, endpoint posture, and proof lane.
- Clarify Compose vs Kubernetes proof claims.
- Add recovery actions for AI endpoint, search provider, filesystem MCP, scheduler, comms, NATS, Postgres, and workspace root issues.

### Phase 5 - Confidence Provenance Preparation

- Add schema room for validation source, evidence strength, proof quality, review lineage, and cross-model/tool agreement.
- Surface confidence provenance only where it helps operator trust.
- Avoid scores without evidence semantics.

## Acceptance Gates

A finalization slice is acceptable only when it answers:

1. Does this improve visible execution?
2. Does this improve operator trust?
3. Does this improve durable outputs or proof?
4. Does this improve recovery or degradation handling?
5. Does this improve deployment reality?
6. Does this reduce conceptual density?
7. Does this strengthen Soma as the singular operating surface?

Required proof per slice:

- focused unit/API tests for the touched contract
- frontend typecheck and build when UI changes
- Playwright proof for touched operator workflows
- live backend or deployment proof when runtime behavior changes
- docs-link/layout proof when docs change
- `.state/V8_DEV_STATE.md` update when state or acceptance meaning changes

## Known Blockers And Risks

| Risk | Current posture | Required action |
| --- | --- | --- |
| Doctrine expansion | Active risk | Reject architecture-only expansion unless it improves embodiment, recovery, proof, deployment trust, or confidence provenance. |
| Runtime-team usefulness | `BLOCKED` | Do not claim runtime teams are useful delivery collaborators until one bounded role-specific ask completes within timeout with visible degradation handling. |
| Scheduler/cadence authoring | `NEXT` | Ship governed scheduler rules before exposing cadence authoring as product behavior. |
| Deployment env footguns | Active operational risk | Keep Compose/Kubernetes endpoint posture explicit; avoid loopback assumptions and document container-safe overrides. |
| UI density | Active cleanup lane | Continue menu/detail and disclosure compression for lengthy advanced surfaces. |
| Expression collapse into query/chat | Active product risk | Treat the input as output/agentry evocation: capture outcome, output shape, use context, constraints, proof, and continuation before runtime topology. |
| Browser workspace scroll sprawl | Active UX risk | Replace long inner-scroll surfaces with single-window list/detail or workbench panes, and collapse mobile chrome so the input, active work, output, and trust summary remain usable. |
| Cold-start Soma trust copy | `NEXT` | Default dashboard must not show "Soma just did this" or proof/trust package language before real Soma activity. |
| Exact first demo GUI proof | `NEXT` | Extend live browser proof from playable file output to project-package output with README, validation, proof-link opening, and degraded retry. |
| Capability manifest persistence | `REQUIRED` | Persist refreshed manifests and reconcile long-lived health/probe state as P0 runtime trust infrastructure. |
| Execution/proof schemas | `REQUIRED` | Promote ExecutionContract and ProofArtifact from implied concepts into durable runtime-visible objects. |
| Active team interaction contract | `NEXT` | Promote TeamInteraction, TeamWorkItem, TeamStatusEvent, and TeamOutputRef into durable runtime-visible APIs and Soma Active Work Lane projections. |
| Enterprise auth overclaim | Active product risk | Keep provider setup review-only until adapters, policy, audit, and recovery are runtime-enabled. |
| Source size pressure | Active hygiene lane | Keep `quality.max-lines` green and ratchet legacy caps down through modularization. |

## Non-Goals For Finalization

- Do not build a marketplace before the governed capability/output loop is excellent.
- Do not add new topology language to default UX.
- Do not overbuild recursive autonomy or learning loops.
- Do not expose MCP/server inventory as the main capability story.
- Do not treat docs as a substitute for live proof.
- Do not claim enterprise auth, SCIM, distributed execution nodes, or full confidence scoring are complete before runtime proof exists.

## Open Architecture Decisions

1. What is the first minimal persistence/API implementation for the normalized TeamInteraction and TeamWorkItem surface?
2. What is the first production scheduler rule shape that covers real cadence without overbuilding an automation suite?
3. Which project-package output fields are required for reconstruction across Compose and Kubernetes proof lanes?

## Architecture Team Close-Out Contract

For every finalization slice, the architecture team must record:

- owning lane
- current status marker
- touched runtime/UI/API surfaces
- emitted events
- output object shape
- proof artifact shape
- UI proof
- recovery behavior
- validation run
- docs changed
- docs reviewed and left unchanged
- remaining blocker, if any

The next threshold is not whether Mycelis can describe the architecture. The threshold is whether an operator can trust the organism through visible execution, proof, and recovery.
