# Mycelis V8.3 - Multi-Agentry Steering Doctrine
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Index](ARCHITECTURE_LIBRARY_INDEX.md)

## Steering Transition

The organism now enters coordinated steering phase. This changes agentry from isolated execution, architecture speculation, and disconnected implementation into synchronized operational embodiment, shared runtime truth, and governed delivery convergence.

The purpose of steering is not to create more agents. The purpose of steering is coordinated manifestation of one coherent operational organism.

## Primary Steering Principle

No agentry operates independently anymore.

All teams operate through:

```text
Shared Runtime Truth
-> Shared Product Identity
-> Shared Execution Contracts
-> Shared Trust Semantics
-> Shared Event Spine
-> Shared Proof Doctrine
```

No subsystem may evolve outside that shared operational gravity.

## Multi-Agentry Doctrine

Multiple agentry are specialized execution lanes within one governed cognition environment, not autonomous personalities, independent products, or disconnected architectures.

Every team exists because specialization is required, execution load exists, review separation is useful, or operational continuity benefits from it. If a team does not materially improve execution, compress or remove it.

## Canonical Steering Loop

All agentry coordination flows through:

```text
Operator Intent
-> Soma Interpretation
-> Execution Contract
-> Capability / Team Assignment
-> Async Governed Execution
-> Event Spine Emission
-> Proof + Recovery
-> Retained Operational Output
```

No execution pathway may bypass this loop.

## Soma Steering Authority

Soma is the singular operational authority surface. The operator speaks to Soma. Soma directs the organism.

Teams do not compete for authority. Capabilities do not bypass governance. Runtime systems do not expose topology by default.

Soma maintains continuity, trust, state coherence, execution interpretation, and operational translation.

## Steering Goals

The steering layer optimizes for observable execution, recoverable execution, async operational continuity, bounded cognition, governed capability usage, proof-backed outputs, operator trust, and deployment realism.

It does not optimize for autonomy spectacle, recursive expansion, architectural novelty, or agent proliferation.

## Shared Runtime Semantics

All agentry must use canonical runtime semantics.

Canonical `RunState` values:

```text
queued, awaiting_approval, approved, executing, streaming_output, output_ready,
partially_complete, degraded, blocked, recovery_available, completed, failed,
cancelled
```

No team may invent alternate lifecycle language.

## Shared Trust Semantics

All teams must consistently expose:

```text
trusted, partially_trusted, unverified, degraded, invalidated, requires_review,
recovered
```

Trust is runtime state, not implication, intuition, or UI flavoring.

## Shared Proof Doctrine

Outputs are durable operational artifacts. Every meaningful execution must attach output refs, proof refs, capability lineage, trust state, recovery lineage, and Event Spine linkage.

A task is not complete because the model responded, code compiled, or UI rendered. A task is complete only when the operator can trust what happened.

Run-linked team work now participates in this doctrine through `team_work.status` mission events. Teams still own concise `TeamStatusEvent` timelines, but the run-level Event Spine carries the normalized state, blocker, proof, and next-action metadata required for replay and recovery review.

## Async Runtime Steering

All slow execution must become queued, resumable, event-driven, and non-blocking. The organism must never depend on synchronous cognition loops.

The operator must always retain continuity, visibility, and interaction capability while work is active.

## Recovery Steering

Recovery is no longer explanatory text. Recovery becomes executable operational cognition.

Every degraded workflow must attempt to produce `RecoveryAction` records, retry paths, fallback paths, or safe degradation. The operator should not leave Soma to repair ordinary runtime failures.

## ExpressionFrame Doctrine

`ExpressionFrame` is the canonical translation layer between runtime topology and operator understanding.

ExpressionFrames preserve operator mental state, operational clarity, trust continuity, and progressive disclosure. Every frame must answer what is happening, why it is happening, what is trusted, what failed, and what the operator can do next without exposing raw infrastructure by default.

## Team Steering Rules

Runtime teams are bounded operational collaborators. No team may self-expand, recursively spawn without governance, mutate capability authority, or bypass approval boundaries.

All team delegation must attach to runs, attach to proof, emit events, and remain inspectable.

## Team Deployment And Write Ownership

Parallel delivery teams must receive disjoint write scopes. A team may read anywhere, but writes are owned by one lane per slice so implementation, proof, and docs can converge without avoidable merge conflicts.

| Lane | Primary Write Scope | Shared Read-Only Context |
| --- | --- | --- |
| Runtime / Async | Core runtime packages, team APIs, migrations, runtime tests | V8.3 PRD, steering doctrine, API contracts |
| Recovery / Trust | proof, recovery, audit, trust-state runtime packages and tests | Runtime / Async contracts and Event Spine semantics |
| Soma UI | Interface dashboard, workspace, `ExpressionFrame`, Active Work, output/proof components and tests | API fixtures, canonical state objects, docs manifest |
| Media / Capability | `cognitive/`, media gateway tests, media tasks, capability docs | Proof contracts, recovery contracts, output-root config |
| Docs / State | architecture docs, user docs, ops docs, docs manifest, `.state/V8_DEV_STATE.md` | Accepted implementation changes and validation output |
| QA / GUI | e2e specs, browser fixtures, proof runbooks, test harnesses | All implementation lanes after handoff |

Collision rules:

1. No two teams edit the same file in the same slice.
2. Shared contracts change first, then implementation lanes branch from those contracts.
3. Docs / State performs the final integration pass after implementation teams finish.
4. Test fixture ownership belongs to the team that owns the touched runtime or UI surface.
5. If a second lane needs a shared file, it records an integration note instead of editing directly.
6. The orchestrator owns final merge order, quality gates, and conflict resolution.

Every team handoff must name:

```text
owning lane
files changed
files read only
runtime objects changed
events emitted
UI surfaces changed
tests run
docs/state changed
recovery behavior
remaining risks
```

## Capability Steering Rules

Capabilities are governed operational authorities. Every capability must expose purpose, risk, approval requirements, output expectations, fallback posture, recovery behavior, and trust implications.

MCP remains implementation infrastructure beneath governed capability surfaces.

## Deployment Steering Rules

Deployment posture is part of product trust. The organism must maintain deployment visibility, proof lanes, health posture, recovery posture, and artifact lineage.

Dirty runtime state is operational debt. Release-candidate proof must occur on disciplined runtime state.

## Current Steering Priorities

P0:

1. Commit and stabilize media gateway.
2. Generate break-glass posture.
3. Convert runtime execution to async.
4. Embody `ExpressionFrame`.
5. Implement `RecoveryAction`.
6. Universalize proof.
7. Run full fresh-state GUI proof.

P1:

1. ComfyUI workflow adaptation live proof and operator workflow-template UX.
2. Canonical trust package.
3. Advanced team steering.
4. Capability UX refinement.
5. Confidence provenance preparation.

## Operational Embodiment Principle

The organism no longer needs more architecture. It now needs operational coherence.

Steering exists to compress complexity, synchronize execution, stabilize runtime behavior, and embody trust through Soma.

## Final Steering Directive

Do not expand the organism. Synchronize it.

Do not optimize for apparent intelligence. Optimize for governed execution, runtime continuity, proof-backed trust, async embodiment, and deployable operational clarity.

The next threshold is not whether Mycelis can describe itself. The next threshold is whether the operator can safely trust the organism under real execution conditions.

Proceed accordingly.
