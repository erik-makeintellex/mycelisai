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

Mycelis is a Soma-centered governed cognitive operating environment. It is not a chatbot, orchestration playground, topology console, MCP registry, AI dashboard pile, automation toy, or multi-agent roleplay environment.

The system exists to:

```text
Transform operator intent into governed, inspectable, recoverable execution.
```

Canonical loop:

```text
I tell Soma what I want.
Soma safely directs execution.
I can see active work, outputs, proof, and recovery.
I can trust what happened later.
```

The operator should never feel like they are operating an agent topology stack. They should feel like they are operating a governed cognitive operating environment.

## Core Principles

Every decision must improve one or more of:

```text
observable
recoverable
governable
deployable
interruptible
inspectable
trustworthy
```

If a feature does not improve those dimensions, defer it.

## Product Identity Doctrine

The product category is `Governed Operational Cognition Infrastructure`, not artificial general intelligence, an autonomous agent society, a recursive autonomy framework, or a universal automation system.

The architecture must stay disciplined around bounded cognition, governed execution, explicit authority, visible runtime state, durable proof, recoverable workflows, and operator trust.

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

`ExpressionFrame` is the canonical translation layer between internal topology and operator understanding. Every frame must show intent, current state, next action, approval posture, risk posture, outputs, proof, recovery, and inspectable advanced details.

Canonical frame types: `DirectAnswerFrame`, `ProposalFrame`, `ActiveWorkFrame`, `OutputReadyFrame`, `ProofFrame`, `RecoveryFrame`, `BlockedFrame`, `DegradedFrame`.

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

Current state: the local/private media gateway is landed and in review for the Pinokio Forge/AUTOMATIC1111 path. It exposes the OpenAI-compatible image endpoint, returns `b64_json`, blocks public upstreams by default, and has unit proof for adapter, privacy, and unsupported-backend behavior.

Required next: verify local generation against a running private media backend, verify retained artifacts, verify proof linkage, verify degradation handling, and verify recovery actions through Soma/Core/UI. Only after this should ComfyUI integration begin.

## Runtime Teams Doctrine

Runtime teams are bounded operational collaborators, not autonomous digital coworkers. They are not production-ready delivery collaborators until async execution works, degradation is visible, output reliability stabilizes, and bounded tasks complete within expected operational posture.

## Release Candidate Threshold

The system reaches RC posture when a fresh operator can log in, ask Soma for work, approve governed execution, observe active work, open retained outputs, inspect proof, encounter degradation, execute recovery, and trust the result later without reading logs, inspecting NATS, debugging infrastructure, manually interpreting runtime state, or understanding topology.

## Delivery Priorities

P0:

1. Push current V8.3 embodiment commits so WSL and hosted proof can refresh from git.
2. Generate and verify `MYCELIS_BREAK_GLASS_API_KEY`.
3. Rerun auth posture proof.
4. Convert team asks and confirmed execution to async/polling/event-driven runtime.
5. Canonicalize `RunState`.
6. Implement `RecoveryAction` runtime contracts.
7. Formalize `ExpressionFrame`.
8. Run local media generation through Soma.
9. Verify artifact/output/proof linkage.
10. Run full fresh-state GUI proof.

P1:

1. ComfyUI workflow adapter.
2. Canonical trust package assembly.
3. Team steering lifecycle beyond the landed bounded ask/action APIs.
4. MCP/operator capability UX refinement.
5. Documentation compression.
6. Confidence provenance preparation.

## Non-Goals

Do not build recursive autonomy, self-modifying agents, public capability marketplace, speculative AGI systems, enterprise IAM expansion, confidence scoring systems, massive runtime societies, or unrestricted plugin ecosystems.

## Acceptance Gates

A slice is acceptable only when code exists, runtime object exists, UI represents it, events emit correctly, proof exists, failure degrades safely, recovery is actionable, docs/state are updated, and browser proof passed.

Every slice must answer: what operator problem was solved, what runtime object changed, what UI surface changed, what proof exists, what recovery exists, what remains untrusted, and what should still be deferred.

## Final Doctrine

Do not expand the organism. Embody it. Do not expose topology. Translate it through Soma. Do not call work complete because code exists. Call it complete when an operator can trust what happened.

Do not optimize for autonomy first. Optimize for observable, recoverable, governable, proof-backed execution.

The next threshold is whether a fresh operator can use Soma, approve work, observe execution, recover from failure, inspect proof, and trust the result later through a coherent governed operating experience.
