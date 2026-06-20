# Mycelis V8.3 - Operational Embodiment PRD
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Index](ARCHITECTURE_LIBRARY_INDEX.md)
## Status
**Date:** 2026-05-23  
**Phase:** Release Candidate Embodiment  
**Target:** Governed operational cognition platform maturity  
**Architecture posture:** Converged and entering operational refinement  
**Primary focus:** Operator trust, async runtime execution, proof universality, recovery embodiment, and deployable runtime coherence
## Executive Summary
Mycelis has reached a critical transition point. The architecture is no longer primarily blocked by conceptual design. The remaining challenges are operational embodiment, async execution coherence, runtime trust semantics, recovery systems, deployment reliability, proof universality, operator-visible execution, and reduction of cognitive load.
The product is no longer proving that multi-agent orchestration is possible. It is now proving:
```text
Governed cognition can safely direct meaningful operational execution.
```
The next threshold is whether a fresh operator can log in, ask Soma for work, understand what is happening, approve governed execution, observe active work, inspect outputs and proof, recover from failure, and trust the result later without reading logs, understanding NATS, interpreting MCP topology, or manually debugging infrastructure.
## Product Truth
Mycelis is a Soma-centered Outcome Management Engine powered by governed execution. It is not a chatbot, orchestration playground, topology console, MCP registry, AI dashboard pile, automation toy, or multi-agent roleplay environment.
The system exists to:
```text
Transform outcome needs into owned work, deliverables, proof, recovery, and revisit paths.
```
Canonical loop:
```text
I tell Soma what outcome I need.
Soma safely directs the work.
I can see active work, deliverables, proof, and recovery.
I can revisit and trust what happened later.
```
The operator should never feel like they are operating an agent topology stack. They should feel like they are owning outcomes through Soma.
## Core Principles
Every decision must improve one or more of: observable, recoverable, governable, deployable, interruptible, inspectable, and trustworthy. If a feature does not improve those dimensions, defer it.
## Product Identity Doctrine
The product category is `Soma-centered Outcome Management Engine`, not artificial general intelligence, an autonomous agent society, a recursive autonomy framework, or a universal automation system.
The architecture must stay disciplined around bounded cognition, governed execution, explicit authority, visible runtime state, durable proof, recoverable workflows, and operator trust.
### Autonomy Control Boundary
Future autonomy is an operating mode, not an MVP bypass. The active boundary is defined in [V8.3 Autonomy Control Architecture](V8_3_AUTONOMY_CONTROL_ARCHITECTURE.md): autonomy may change the source of intent, but it must still pass through ExecutionContract, policy evaluation, governed run, capability invocation, event emission, output/proof, and review/recovery.
V8.3 should prepare for autonomy only where the work strengthens observability, governance, interruption, proof, recovery, capability boundaries, policy, or budgets. It must not introduce silent mutation, hidden memory promotion, self-granted permissions, unbounded tool use, automatic policy changes, or self-improvement paths that mutate production behavior without review and rollback.
## Current Architectural Assessment
Already implemented or substantially embodied:
- Soma-centered operational surface
- Event Spine runtime
- durable outputs and proof-linked outputs
- local-first runtime posture
- NATS runtime fabric
- PostgreSQL plus pgvector persistence
- local/private inference posture
- Google Workspace SSO foundation
- capability manifests and MCP normalization
- governed proposal -> confirm -> execute flow
- deployment proof lanes
- retained outputs and artifact roots
- runtime team persistence
- active work visibility
- deployment trust posture
The next risks are operational clarity, runtime cohesion, async execution, recovery embodiment, proof consistency, and release reliability.
## Primary Architectural Risks
### Risk 1 - Blocking Runtime Execution
The current synchronous request posture for slow local-model execution is incompatible with local inference, media generation, multimodal workflows, orchestrated execution, long-running capabilities, and recoverable operational cognition.
Resolution: Mycelis must become async-first through queued work, event-driven execution, resumable workflows, progressive output, streaming state, durable run progression, and non-blocking UI.
### Risk 2 - Recovery Is Still Informational
Recovery currently exists too often as explanatory text. The operator must not leave Soma to manually repair runtime state.
Resolution: recovery must become executable runtime objects.
### Risk 3 - Runtime Trust Is Too Diffuse
Trust semantics exist conceptually but are not fully canonicalized.
Resolution: canonical trust states, proof validity states, degraded execution semantics, recovery trust transitions, and output trust lineage must become runtime state.
### Risk 4 - Operator Mental Model Drift
As topology becomes hidden behind Soma, the operator still needs to understand what is happening, why it is happening, whether it is safe, whether it succeeded, whether it degraded, and whether it can recover.
Resolution: `ExpressionFrame` becomes the translation layer between internal topology and operator understanding.
### Risk 5 - Working Tree / Release Discipline
Dirty working trees, incomplete proof chains, and unverifiable runtime state are now operational risks.
Resolution: release-candidate proof must run only from explainable committed state.
## Canonical Runtime Doctrine
All meaningful execution must become queued, stateful, recoverable, and event-driven. The UI must never depend on blocking runtime execution. Every meaningful run should support queued state, resumable state, degraded continuation, streaming outputs, partial completion, retry behavior, recovery actions, and proof linkage.
## Canonical Runtime Objects

### Outcome

```ts
Outcome {
  outcome_id, requested_by, source_surface, goal, owner_scope, active_state,
  deliverable_refs, proof_refs, recovery_refs, trust_state, next_action,
  created_at, updated_at
}
```

Doctrine: `Outcome` is the primary product object. Execution contracts, runs, team work, capabilities, events, and output packages exist to make the outcome visible, trusted, recoverable, and revisitable.

### ExecutionContract

```ts
ExecutionContract {
  contract_id, created_by, source_surface, operator_intent, normalized_goal,
  execution_type, allowed_capabilities, forbidden_capabilities, allowed_mutations,
  approval_requirements, expected_outputs, proof_requirements, degradation_policy,
  retry_policy, recovery_policy, trust_constraints, created_at, updated_at
}
```

Doctrine: no governed execution occurs without an `ExecutionContract`.

### RunState

```text
queued, awaiting_approval, approved, executing, streaming_output, output_ready,
partially_complete, degraded, blocked, recovery_available, completed, failed,
cancelled
```

All runtime surfaces must use canonical `RunState` semantics.

### ProofArtifact

```ts
ProofArtifact {
  proof_id, run_id, contract_id, output_refs, capability_refs, evidence_refs,
  validation_results, trusted_state, invalidated_state, degradation_refs,
  recovery_options, generated_at
}
```

Doctrine: a run is complete only when outputs, proof, validation, and trust state are inspectable.

### TrustState

```text
trusted, partially_trusted, unverified, degraded, invalidated, requires_review,
recovered
```

Trust becomes canonical runtime state.

### RecoveryAction

```ts
RecoveryAction {
  action_id, run_id, recovery_type, title, explanation, required_approval,
  risk_level, capability_required, expected_effect, target_state, retry_target,
  proof_expected
}
```

Examples: reconnect filesystem capability, retry with fallback provider, repair workspace root, resume queued run, rebuild output index, restart media generation, reconnect MCP server.

### TeamWorkItem

```ts
TeamWorkItem {
  id, requested_by, source_surface, intent, status, assigned_team, output_refs,
  proof_refs, degradation_refs, created_at, updated_at
}
```

## Event Spine Doctrine

The Event Spine is the operational truth layer. It is not logging, analytics, or telemetry decoration. It must support replay, reconstruction, audit, debugging, recovery, proof lineage, and execution history. Every meaningful execution transition must emit durable events.

Current implementation state: run-linked `TeamStatusEvent` records mirror into `mission_events` as `team_work.status` payloads with normalized team/work/run/proof/source/state/blocker/next-action metadata. This makes confirmed team creation, delegated work, retained deliverables, and degraded team-work transitions reconstructable from the run event timeline without exposing raw bus topology in the default UI.

## ExpressionFrame Doctrine

`ExpressionFrame` is the canonical translation layer between internal topology and operator understanding. Every frame must show outcome need, current state, next action, approval posture, risk posture, deliverables, proof, recovery, and inspectable advanced details. For governed actions, the default operator question should remain simple: "Run this now?" Risk, cost, affected resources, capability IDs, proof intent, and team/tool wiring remain available through review details unless they require immediate operator attention.

Canonical frame types: `DirectAnswerFrame`, `ProposalFrame`, `ActiveWorkFrame`, `OutputReadyFrame`, `ProofFrame`, `RecoveryFrame`, `BlockedFrame`, `DegradedFrame`.

Focused work context is part of ExpressionFrame, not a separate topology view. Selecting a running or executed team should switch the workbench into that team's scoped chat, active-work, output, proof, and recovery lane while keeping Soma available as the cross-context coordinator. This lets one team own a story-writing context, another team own comic-page visual generation, and Soma reference either context deliberately without merging every interaction into one unbounded root chat.

The default workbench posture should favor operator focus. A compact current-work lane may summarize the selected workflow, active task posture, latest output, and next review action above Soma, while Active Work, outputs, trust, and context remain available through an expandable/minimizable Work panel overlay rather than permanently squeezing Soma into a narrow chat column. The panel should be tabbed and quick-action oriented; dense backlog review, output browsing, proof/audit review, and resource configuration belong on full pages or Inspect surfaces.

### Enterprise Workflow UI Pattern Checkpoint
Research checkpoint, 2026-06-04: Airflow separates home health, DAG lists, details, grid/graph inspection, run tabs, task logs, events, code, and asset lineage; Temporal separates saved workflow views, failure views, event history, actions, relationships, workers, pending activities, and schedules; GitHub Actions separates workflow lists, run summaries, jobs/steps/logs, rerun/cancel actions, and artifacts; Linear-style systems emphasize filters, grouping, saved/custom views, team pages, and display options. Sources: [Airflow UI Overview](https://airflow.apache.org/docs/apache-airflow/stable/ui.html), [Temporal Web UI](https://docs.temporal.io/web-ui), [GitHub Actions workflow run history](https://docs.github.com/en/actions/how-tos/monitor-workflows/view-workflow-run-history), and [Linear filters](https://linear.app/docs/filters).
Applied UI direction: default Soma chat uses trusted-result receipts, one obvious latest-output open path, compact media previews, a compact current-work lane, and plain run confirmation while keeping proof/log/resource density in the workbench panel or full pages. Work contexts switch chat, active work, outputs, proof, and recovery together through a bounded current-workflow picker, not a horizontally growing team tab strip or pre-chat output dock. Activity, Runs, Resources, Groups, Teams, System, and Settings each need a clear page job: overview, active work, output/recovery, or advanced inspect. Multi-job pages should use tabs or inner scrolling instead of extending the browser page indefinitely. Copy should speak to outcomes: ask, run, running, output ready, open file, open folder, proof, recovery. Internal vocabulary belongs behind Advanced/Inspect.

### MVP UI Runtime Delivery Overlay

The market pattern review on 2026-06-14 confirms that production agent platforms are converging around agent runtimes, runs, traces, human review, workflow visibility, capability catalogs, templates, and deployment/observability consoles. Mycelis should keep its default experience more compressed than node-builder products: Soma remains the primary operating surface, while visual run maps, capability topology, and raw event payloads stay behind Inspect.

The executable overlay lives in [V8.3 MVP UI Runtime Delivery Plan](V8_3_MVP_UI_RUNTIME_DELIVERY_PLAN.md). It locks the next delivery train around:

- one Soma work inbox for active, output-ready, failed, recovery, and archived work
- one run receipt standard for outcome, output, trust, proof, recovery, and inspect
- one output package card standard for preview, open, open folder, README/PROOF, download, and Resources re-entry
- one capability catalog that answers "what Soma can use" before exposing raw MCP server structure
- one recovery queue that states failed, still trusted, not trusted, safe next, and operator action
- one advanced run-map/timeline grammar for inspection without default topology exposure

## Understanding Attunement Doctrine

Soma must treat operator expression as intent-bearing work context, not as a bare command string. Before direct answers or governed proposals, Soma should compactly infer desired outcome, audience, output form, constraints, relevant prior/workspace context, and uncertainty.

Knowledge lookup is part of attunement, not an ungoverned side channel. Soma should prefer workspace, organization, retained-output, and deployment context first; use configured search/source capability for current, external, or research-heavy questions; disclose the source boundary in the trust package; and surface a blocker or recovery path when necessary source capability is unavailable.

The goal is better-fitting delivery, not more doctrine. Attunement should make outputs land closer to the user's intended world while preserving governance, proof, and progressive disclosure.

## Progressive Operational Disclosure

New operators should see Soma, task, result, proof, and recovery. Advanced users may inspect capabilities, providers, runs, deployment roots, MCP state, audit chains, Event Spine details, and provider routing. The product must never default to topology exposure.

## Operator Attention Doctrine

The system requires one unified operator attention model: needs approval, needs review, needs recovery, needs retry, capability degraded, proof invalidated, output ready, run stalled, auth posture incomplete. This becomes operational continuity infrastructure, trust infrastructure, and workflow glue.

## Outputs Doctrine

Outputs are durable operational artifacts, not chat residue. They may include plans, reviews, files, media, generated packages, deployment proof, audit objects, capability outputs, and workflow artifacts. Every meaningful output must persist, attach to runs, attach to proof, remain inspectable, remain attributable, and support recovery/review.

## Capability Doctrine

Every capability is a governed operational authority, not just a tool. Capabilities define purpose, risk, approval requirements, allowed roles, output schemas, recovery behavior, fallback posture, and trust implications.

## Deployment Doctrine

Preferred development posture is Windows source editing, local PostgreSQL, local NATS, local inference, and source Core plus Interface. Compose and Kubernetes remain proof/deployment lanes, not primary development ergonomics.

Local/private runtime capability remains strategic infrastructure for local inference, local media generation, sovereign outputs, private prompts, enterprise deployment, and portability.

## Media Lane Doctrine

The local/private media lane is sovereign cognition infrastructure, not merely image generation support.

Current state: the local/private media gateway and ComfyUI retained-output proof are green for the current Soma journey. The gateway exposes the OpenAI-compatible image endpoint, returns `b64_json`, blocks public upstreams by default, adapts Forge/AUTOMATIC1111 `txt2img`, and can submit a reviewed ComfyUI API-format workflow to `/prompt`, poll `/history/{prompt_id}`, and retrieve generated files through `/view`.

Required next: keep richer workflow-template UX and package media journeys in review while preserving retained artifacts, proof linkage, degradation handling, and recovery actions through Soma/Core/UI.

## Runtime Teams Doctrine

Runtime teams are bounded operational collaborators, not autonomous digital coworkers. They are not production-ready delivery collaborators until async execution works, degradation is visible, output reliability stabilizes, and bounded tasks complete within expected operational posture.

Generic team creation should remain lead-owned and compact. Explicit specialist-output creation is a first-class exception: when the operator asks for a concrete retained deliverable and names the needed roles, Soma should preserve the smallest useful specialist roster in the ExecutionContract, execute the first deliverable through governed capabilities, and return retained output/proof or a degraded RecoveryAction. A comic-page request with artist, character, dialogue, layout, proof, `generate_image`, and `save_cached_image` is the current embodiment target for this path.

## Release Candidate Threshold

The system reaches RC posture when a fresh operator can log in, ask Soma for work, approve governed execution, observe active work, open retained outputs, inspect proof, encounter degradation, execute recovery, and trust the result later without reading logs, inspecting NATS, debugging infrastructure, manually interpreting runtime state, or understanding topology.

The release journey is:

```text
Ask
-> Understand
-> Approve
-> Execute
-> Deliver
-> Trust
-> Recover
-> Revisit
```

All P0 work must strengthen this Trusted Outcome Journey. Output Packages strengthen Deliver, Run Receipts strengthen Trust, Recovery Queue strengthens Recover, Review Inbox strengthens Understand and Approve, Capability Catalog strengthens Trust, and Resources/Groups convergence strengthens Revisit. Subsystems are not complete until they help a non-technical user own the outcome and trust the result later.

## Delivery Priorities

P0:

1. Keep output packages, service health, and headed browser proof green.
2. Finish the Soma/Teams review inbox pattern without card sprawl.
3. Promote Resources into a capability catalog that answers what Soma can use, repair, or request before raw MCP topology.
4. Standardize run receipts across successful, degraded, failed, and proposed work.
5. Turn recovery into actionable queue items with still-trusted, not-trusted, safe-next, and repair/retry posture.
6. Run full fresh-state GUI proof only after the visible workflow language converges.

P1:

1. ComfyUI workflow adapter live proof and operator workflow-template UX.
2. Canonical trust package assembly and confidence provenance preparation.
3. Team steering lifecycle beyond the landed bounded ask/action APIs.
4. Deeper MCP/operator capability UX after the catalog is usable.
5. Documentation compression after each accepted MVP slice.

## Non-Goals

Do not build recursive autonomy, self-modifying agents, public capability marketplace, speculative AGI systems, enterprise IAM expansion, confidence scoring systems, massive runtime societies, or unrestricted plugin ecosystems.

## Acceptance Gates

A slice is acceptable only when code exists, runtime object exists, UI represents the outcome, events emit correctly, proof exists, failure degrades safely, recovery is actionable, docs/state are updated, browser proof passed, and the slice names which Trusted Outcome Journey step it improves.

Every slice must answer: which journey step improved, what operator problem was solved, what runtime object changed, what UI surface changed, what proof exists, what recovery exists, what remains untrusted, and what should still be deferred.

## Final Doctrine

Do not expand the organism. Embody it. Do not expose topology. Translate it through Soma. Do not call work complete because code exists. Call it complete when an operator can trust what happened.

Do not optimize for autonomy first. Optimize for observable, recoverable, governable, proof-backed execution.

The next threshold is whether a fresh operator can use Soma, approve work, observe execution, recover from failure, inspect proof, and trust the result later through a coherent governed operating experience.
