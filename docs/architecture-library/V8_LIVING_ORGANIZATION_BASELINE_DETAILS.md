# V8 Living Organization Baseline Details
> Navigation: [V8.1 Living Organization Architecture](V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md)

Status: V8.1 baseline detail.

## Loop Profiles

Loop profiles define whether work is interactive, scheduled, monitored, or bounded automation. Persistent loop behavior must be visible and policy-aware.

## Runtime Capabilities

Capabilities describe what tools, resources, services, and actions a scope may use. They are governed by organization policy and may be narrowed by team/agent scopes.

## Response Contract

Responses should normalize into:
- direct answer
- proposal
- execution result
- blocker
- error/empty/loading state

Response style may inherit, but raw model output must not replace UI contracts.

## Agent Type Profiles

Agent types describe role expectations, tools, provider posture, and verification needs. They should support specialists without turning the default UX into a flat agent list.

## Memory Promotion And Semantic Continuity

Continuity uses reviewed memory, retained outputs, and semantic context. Private or transient data does not become memory without reviewable policy.

## Procedure / Skill Sets

Procedures and skills define repeatable work patterns. They must remain inspectable and governed when they trigger tools or external systems.

## Layering Clarification

Runtime layers:

```text
Inception -> Soma -> Council/advisors -> teams/departments -> specialists -> tools/capabilities
```

Lower layers may specialize but cannot bypass higher policy.

## Managed Exchange Foundation

Managed exchange keeps team/group work reviewable through status, result, retained outputs, activity, and audit evidence.
