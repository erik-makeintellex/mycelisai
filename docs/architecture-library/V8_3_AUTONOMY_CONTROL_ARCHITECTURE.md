# Mycelis V8.3 - Autonomy Control Architecture
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Index](ARCHITECTURE_LIBRARY_INDEX.md)

> Status: Canonical boundary
> Last Updated: 2026-06-19
> Purpose: Define how Mycelis can prepare for future autonomy without weakening governed MVP execution.

## Scope

Autonomy is not the V8.3 MVP surface. The active MVP remains:

```text
Intent
-> Soma
-> Proposal / Approval
-> Execution
-> Output Package
-> Proof
-> Recovery
-> Re-entry
```

This document exists so future autonomous, adaptive, or self-improving behavior is shaped by control infrastructure before capability expansion.

## Control Principle

Mycelis may become more autonomous only after it becomes more controllable.

Autonomous behavior is allowed only when it is observable, governable, interruptible, reversible where possible, auditable, permissioned, scoped, rate-limited, recoverable, and explainable.

Control comes first. Autonomy comes second. Self-improvement comes last.

## No-Bypass Runtime Spine

Autonomous work changes the source of intent. It does not bypass the governed runtime.

Every autonomous action must still pass through:

```text
Intent or Trigger
-> ExecutionContract
-> Policy Evaluation
-> Governed Run
-> Capability Invocation
-> Event Emission
-> Output / Proof
-> Review / Recovery
```

Allowed intent sources may eventually include user request, schedule, event, monitoring signal, system-health signal, prior-run follow-up, external signal, or learning observation. Each source must be visible in the run receipt and audit trail.

## Prohibited Autonomy

Mycelis must not support:

- silent mutation
- uncontrolled learning
- unbounded tool use
- self-granted permissions
- hidden memory promotion
- recursive task creation without depth and budget controls
- automatic policy changes
- unaudited external action
- unlimited spend, tokens, compute, retries, or external requests
- unreviewed capability expansion
- invisible model switching
- irreversible unapproved changes
- fabricated success after degraded execution

## Self-Improvement Gate

Self-improvement must be candidate-first:

```text
Observation
-> Improvement Candidate
-> Evidence
-> Risk Classification
-> Review
-> Approval
-> Controlled Experiment
-> Measurement
-> Promotion Decision
-> Rollback Path
```

No self-improvement path may directly mutate policy, permissions, memory, prompts, procedures, models, routing, capability manifests, approval thresholds, deployment config, or external integrations without governance.

## Memory Control

Learning is scoped, inspectable, and reversible. Any promoted memory or learning candidate must include:

- source run or observation
- evidence
- confidence or provenance note
- target memory layer
- visibility scope
- expiry or retention posture
- approval requirement
- rollback path

Hidden memory promotion is not allowed.

## Capability Control

Autonomous systems may only use capabilities that are registered, scoped, health-checked, risk-classified, approval-aware, audit-wrapped, rate-limited, output-normalized, and recovery-aware.

Agents and teams must not discover or use tools outside the governed capability registry. They must not expand their own permissions.

## Policy And Budget Control

Policy is higher authority than agents, teams, schedules, triggers, and model output.

Policies control allowed and forbidden actions, approval thresholds, spend, data access, external communication, destructive operations, memory promotion, deployment changes, capability use, recursion depth, run duration, retry limits, and concurrency.

Autonomous work must support explicit budgets for time, money, tokens, compute, tool calls, retries, recursion depth, external requests, artifacts, and concurrent runs. Budget exhaustion creates a review or recovery item instead of silent continuation.

## Kill Switches

The operator and system must be able to:

- pause or cancel a run
- pause automation
- disable a capability
- freeze memory promotion
- disable external calls
- stop a team lane
- enter degraded safe mode
- require manual approval for every mutation

Kill-switch state must be visible and auditable.

## Degradation And Containment

When autonomous work degrades, Mycelis must stop unsafe escalation, preserve partial outputs, mark trust boundaries, explain uncertainty, request review, offer retry or rollback where possible, and avoid false-success messaging.

The UI and API must answer:

- what failed
- what remains trusted
- what proof is invalid
- what can continue safely
- what requires retry
- what requires operator attention
- what uncertainty now exists

## Audit And Reconstruction

Autonomous and automated runs must be reconstructable from durable records:

- trigger source
- execution contract
- policy decision
- capabilities invoked
- outputs and output packages
- memory candidates
- approvals
- proof
- recovery actions
- final state

Run receipts should summarize this for operators. Advanced Inspect may expose the technical reconstruction.

## UI Boundary

Default users should see what is running, what needs approval, what changed, what is trusted, what failed, and what can be stopped or retried.

Advanced and admin users may inspect autonomous lanes, triggers, policies, budgets, capability scope, memory promotion queues, audit trails, and kill switches.

The default experience must remain Soma-centered and outcome-first. Autonomy controls must not make the main interface feel like a topology console.

## V8.3 Release Rule

Do not build deep autonomy in V8.3.

Build foundations only when they strengthen the current MVP:

- observability
- governance
- interruption
- proof
- recovery
- capability boundaries
- budget and policy readiness

P0 remains output packages, review inbox, run receipts, recovery, capability catalog, headed proof, service health, and release readiness.
