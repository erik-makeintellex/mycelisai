# V8.1 Living Organization Architecture
> Navigation: [Project README](../../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)

Status: V8.1 compatibility baseline.

This is the V8.1 foundation and compatibility baseline. The active V8.2/B2+ delivery frame, canonical full production architecture, and full actuation target live in `architecture/v8-2.md`.

This file protects the V8.1 Soma-primary foundation. Detail that used to make this file oversized now lives in [V8 Living Organization Baseline Details](V8_LIVING_ORGANIZATION_BASELINE_DETAILS.md).

Depends on:
- [V8 Runtime Contracts](V8_RUNTIME_CONTRACTS.md)
- [V8 Config and Bootstrap Model](V8_CONFIG_AND_BOOTSTRAP_MODEL.md)
- [V8 UI/API and Operator Experience Contract](V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md)

## 1. Purpose

V8.1 defines the compatibility baseline for AI Organizations, Soma, council/advisor support, teams/departments/specialists, continuity, capabilities, and bounded automations.

Do not add new V8.2/B2+ target scope here. Use [V8.2 Production Architecture Target](../../architecture/v8-2.md) for active expansion.

## 2. Core Product Model

The baseline product model is:

```text
AI Organization
  -> Soma
  -> advisors / council support
  -> team leads / departments / specialists
  -> governed execution, memory, activity, retained outputs
```

## 3. AI Organization

An AI Organization is the live work context. It carries identity, purpose, configuration posture, continuity, and governed execution state.

## 4. Soma

Soma is the primary counterpart. Soma answers directly when safe, proposes protected actions, routes to teams/specialists when needed, and keeps work reviewable.

## 5. Living Organization Runtime Contracts

The V8.1 baseline includes:
- loop profiles
- runtime capabilities
- response contract inheritance
- agent type profiles
- learning loops and semantic continuity
- procedure/skill sets
- managed exchange foundation

Loop Profiles as the bounded execution layer. Runtime Capabilities as the bounded action layer. Learning Loops as the bounded candidate-capture and promotion-review layer. Memory Promotion and Semantic Continuity as the pgvector-backed recall substrate. Procedure / Skill Sets as reviewed specialist memory bound to Agent Type Profiles.

### 5.1 Loop Profiles

Loop Profiles keep execution bounded and inspectable.

### 5.2 Runtime Capabilities

Runtime Capabilities define allowed action/tool scope.

### 5.3 Response Contract

Responses normalize into answer, proposal, execution result, blocker, and error states.

### 5.4 Agent Type Profiles

Agent Type Profiles bind role, provider posture, tools, and verification.

### 5.5 Memory Promotion and Semantic Continuity

Learning Loop promotion uses pgvector-backed semantic continuity with raw memory, reviewed memory, and promoted memory; no silent self-rewrite is allowed.

### 5.6 Procedure / Skill Sets

Procedure / Skill Sets are reviewed specialist memory bound to Agent Type Profiles.

### 5.7 Layering clarification

Lower layers specialize but do not bypass higher policy.

### 5.8 Managed exchange foundation

The managed exchange foundation provides named channels for work, review, learning, and normalized tool output plus structured threads for planning, work, review, escalation, and learning.

Security foundation: channels, threads, and exchange items carry explicit readers, writers, reviewers, participants, sensitivity classes, and downstream allowed-consumer metadata. capability-producing outputs carry a capability id, risk class, trust class, and audit-ready publication metadata. normalization into exchange does not imply unrestricted trust.

See [V8 Living Organization Baseline Details](V8_LIVING_ORGANIZATION_BASELINE_DETAILS.md).

## 6. Execution model

### 6.1 Interactive mode

Operator-driven interaction through Soma remains the default.

### 6.2 Loop mode

Bounded automations may run with explicit policy, visibility, and recovery posture. They must not hide persistent behavior from the operator.

## 7. System behavior rules

- Default UX starts with AI Organization and Soma, not raw agents.
- Broad work becomes compact lanes or teams.
- Governance controls mutating actions.
- Memory promotion is reviewable.
- Provider/capability policy remains visible in advanced surfaces.
- Activity and retained outputs remain reviewable.

## 8. Operator UI architecture

### 8.1 Primary operator workspace

The primary workspace is Soma-first and organization-scoped.

### 8.2 Automations surface

Automations are bounded, inspectable, and policy-aware.

Automations include Watchers and Reviews where configured.

### 8.3 Advanced UI boundaries

Advanced UI may expose inheritance, config origin, policy, provider routing, capabilities, and runtime health without making those concepts mandatory for first use.

organization defaults and inheritance visibility, deployment/env influence, and sensitive deployment secrets stay file/env/config driven.

## 9. Safety model

### 9.1 Required

- approval/proposal posture for protected actions
- source and scope metadata for signals
- normalized UI errors
- secret references instead of raw secrets
- durable events for mutating actions

### 9.2 Forbidden

- silent mutation
- raw provider/backend errors as user-facing output
- hidden recurring behavior
- unreviewed private data promotion
- committed secrets

## 10. Testing requirements

### 10.1 Backend

Prove API/runtime behavior with `uv run inv core.test` and focused handler/service tests.

### 10.2 Frontend

Prove UI state and rendering with `uv run inv interface.test` and `interface.typecheck`.

### 10.3 End-to-end

Use [V8 UI Team Full Test Set](V8_UI_TEAM_FULL_TEST_SET.md) for browser proof.

## 11. Release target

### 11.1 V8.1 initial release definition

The shippable V8.1 baseline is a Soma-primary AI Organization experience with governed proposals, compact teams/groups, retained outputs, continuity, settings, and visible activity.

### 11.2 First shippable state

The first shippable state must be useful without advanced mode and truthful when dependencies are unavailable.

For V8.1, loops exist as configuration and inspectable architecture, not broad execution; capabilities are defined but not fully exercised; learning continuity architecture is defined even when raw/reviewed/promoted memory promotion is not fully implemented yet; the system remains safe and inspectable.

## 12. Post-release V8.2 targets

V8.2 extends the baseline with distributed execution, active learning, richer capabilities, enterprise identity, and production deployment posture.

## 13. Delivery implication

Changes to baseline behavior require docs review, tests, and state updates. Do not claim compatibility if the Soma-primary path or governance path regresses.

## 14. Summary

V8.1 is the stable compatibility layer. V8.2 builds on it; it does not erase it.
