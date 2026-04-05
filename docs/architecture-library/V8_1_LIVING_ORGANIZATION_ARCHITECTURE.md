# V8.1 Living Organization Architecture
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Status: Canonical V8.1 PRD
> Last Updated: 2026-03-21
> Purpose: Define the current bounded V8.1 release architecture for persistent execution, bounded automation, semantic continuity, learning promotion, and reproducible specialist behavior.
> Depends On: `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`, `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`, `docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md`

This is the current release architecture.

Use this document for the bounded V8.1 release target.
Use `../../v8-2.md` for the canonical full production architecture and full actuation target beyond the current release.
Not every V8.2 production target belongs in the current V8.1 release.

## 1. Why V8.1 exists

V8 established the AI Organization, Soma-primary workspace, bundle-driven startup truth, and bounded operator settings.

V8.1 extends that work from a structured interactive system into a living, persistent, policy-bounded intelligence runtime.

V8.1 introduces:
- Loop Profiles as the bounded execution layer
- Runtime Capabilities as the bounded action layer
- Response Contract as a first-class inheritance contract
- Agent Type Profiles as a first-class runtime layer between Team defaults and Agent instances
- Learning Loops as the bounded candidate-capture and promotion-review layer
- Memory Promotion and Semantic Continuity as the pgvector-backed recall substrate
- Procedure / Skill Sets as reviewed specialist memory bound to Agent Type Profiles

This enables:
- continuous but governed operation
- event-driven workflows
- safe automation
- reproducible specialist behavior
- policy-bounded learning continuity without silent self-rewrite

V8.1 operator posture:
- Soma is the primary user-facing interface and orchestrator
- Team Leads remain visible as the operational leaders Soma works through
- the default workspace stays simple and user-facing even when deeper architecture remains structured beneath it

## 2. Product goals

### 2.1 Primary goal

Enable AI Organizations to:
- operate continuously
- react to events
- supervise internal and external processes
- maintain deterministic behavior through policy and inheritance

### 2.2 Secondary goals

- maintain strict Soma-primary UX discipline
- avoid unbounded multi-agent chaos
- ensure safe extensibility for future hardware and tool integration
- preserve bundle-driven reproducibility as runtime truth

## 3. Non-goals for V8.1

- no full autonomy loops or runaway agents
- no unrestricted tool execution
- no agent-instance editing UI
- no broad advanced config panels
- no multi-organization federation

## 4. Runtime architecture update

```text
Inception
  -> Soma Kernel
  -> Central Council
  -> Specialist Teams
    -> Agent Type Profiles
      -> Procedure / Skill Sets
      -> Agent Instances
  -> Loop Profiles
    -> Review Loops
    -> Learning Loops
  -> Runs
  -> Events
  -> Memory
    -> Semantic Continuity
  -> Reflection
```

V8.1 does not replace V8.

It promotes the execution and inheritance layers that must exist before deeper automation or agent-instance override work can safely ship.

## 5. Architecture layers

### 5.1 Loop Profiles

#### Definition

A Loop Profile defines persistent execution behavior inside an AI Organization.

Loop Profiles are configuration and policy objects first. The first shippable V8.1 state may expose them before live execution is enabled.

#### Required loop types

- Scheduled Loop
  - trigger: cron or interval
  - example: daily report
- Event Loop
  - trigger: event bus such as NATS
  - example: API callback or system event
- Review Loop
  - trigger: completion of another process
  - example: QA reviewing outputs
- Learning Loop
  - trigger: review completion, event, or schedule
  - example: candidate capture, reviewed promotion, procedure-memory update
  - posture: bounded review and promotion only, never silent self-rewrite
- Actuation Loop
  - trigger: schedule or event
  - example: hardware or system action
  - posture: restricted and policy-gated

#### Initial schema contract

```yaml
id: string
name: string
type: scheduled | event | review | learning | actuation

owner:
  type: team | agent_type
  id: string

trigger:
  schedule: optional cron
  interval_seconds: optional integer
  event_subject: optional string

inputs:
  sources: [events | memory | external]

actions:
  allowed_capabilities: []

policy:
  requires_approval: boolean
  audit_required: boolean

state:
  persistence: short | long
  memory_stage: raw | reviewed | promoted
```

#### Required rules

- loops never bypass policy
- loops never imply capabilities
- actuation loops always require stronger gating than passive review loops
- loops must remain auditable even when not yet executable
- Learning Loops capture candidates, route them through review and promotion, and never silently rewrite memory or specialist behavior
- Learning Loops follow a no silent self-rewrite rule even when they surface strong promotion candidates
- memory promotion must stay policy-bounded even when the execution path is automated

### 5.2 Runtime Capabilities

#### Definition

Runtime Capabilities define what an entity is allowed to do.

Capabilities are deny-by-default and allowlist-only.

#### Required categories

- API access
- filesystem
- browser automation
- MCP and tool access
- hardware control
- messaging and event emission

#### Required capability model

```text
Organization
  -> Role (Kernel / Council)
    -> Team
      -> Agent Type
        -> Agent Instance
```

#### Required rules

- allowlist only
- deny by default
- every capability assignment must be auditable
- advanced UI may expose capability boundaries, but default UI must not collapse into a raw permission dashboard

### 5.3 Response Contract

#### Definition

Response Contract defines output behavior constraints.

V8.1 promotes it from a bounded workspace setting to a first-class runtime inheritance contract.

#### Required properties

- tone
- verbosity
- structure
- formatting rules
- safety posture

#### Required inheritance chain

```text
Organization -> Team -> Agent Type -> Agent Instance
```

#### Required rules

- Response Contract always applies
- lower scopes may only diverge through governed override paths
- raw prompt or system-policy text must not become the operator-facing contract surface

### 5.4 Agent Type Profiles

#### Definition

Agent Type Profiles are reusable definitions of specialist behavior.

They sit between Team defaults and individual Agent instances.

#### Required fields

```yaml
id: string
name: string
description: string

engine:
  default: string
  override_allowed: boolean

response_contract:
  default: string
  override_allowed: boolean

procedure_skill_sets:
  references: []
  promotion_allowed: boolean

capabilities:
  required: []
  optional: []

policy:
  mutable: boolean
```

#### Required rules

- Agent Type Profiles always define baseline specialist behavior
- Team Lead auto-instantiation must still resolve through Agent Type Profile defaults
- agent-instance mutation must not ship before Agent Type Profile truth is stable and visible

### 5.5 Memory Promotion and Semantic Continuity

#### Definition

Memory Promotion and Semantic Continuity define how an AI Organization preserves, reviews, promotes, and recalls meaning across runs without collapsing into ungoverned self-modification.

In V8.1, pgvector is not just action storage. It is the semantic continuity substrate for:
- event, action, and result semantic indexing
- review memory
- learning candidates
- promoted organization memory
- promoted team memory
- promoted agent-type memory
- procedure and skill retrieval
- continuity recall

#### Required memory stages

- raw memory
  - unreviewed records, observations, outputs, and candidate patterns
- reviewed memory
  - findings that have passed a bounded review step and are safe to compare or retrieve as governed evidence
- promoted memory
  - organization, team, agent-type, or procedure memory that has passed policy-bounded promotion and becomes reusable runtime truth

#### Required rules

- pgvector acts as the semantic persistence and recall substrate, not as a hidden policy engine
- promotion never happens through silent self-rewrite
- reviewed and promoted memory must preserve lineage back to raw evidence
- promoted memory may inform continuity and retrieval, but it does not replace Response Contract, AI Engine Settings, or Runtime Capabilities

### 5.6 Procedure / Skill Sets

#### Definition

Procedure / Skill Sets are reusable specialist procedures and reviewed execution patterns attached to Agent Type Profiles.

They represent type-bound skill memory, not ad hoc one-off run history.

#### Required properties

- reusable specialist procedures
- reviewed execution patterns
- type-bound skill memory
- retrieval through semantic continuity
- governed promotion path before becoming reusable runtime truth

#### Required rules

- Procedure / Skill Sets sit under Agent Type Profiles and must resolve before agent-instance behavior diverges
- procedure retrieval must use semantic continuity rather than hidden prompt mutation
- skill memory remains inspectable and policy-bounded even when execution becomes more continuous later

### 5.7 Layering clarification

The required V8.1 layering is:
- pgvector = semantic persistence and recall substrate
- Soma Kernel / Team Leads = orchestration and interpretation through a primary Soma interface and subordinate operational leaders
- loops = candidate generation, review, and promotion
- Response Contract, AI Engine Settings, and Runtime Capabilities = separate governed behavior layers, not interchangeable memory fields

### 5.8 Managed exchange foundation

V8.1 now requires a managed exchange foundation so outputs are not treated as unstructured one-off chat blobs.

The managed exchange foundation includes:
- named channels for work, review, learning, and normalized tool output
- a canonical field registry used across artifacts and messages
- typed schemas such as `TextResult`, `PlanResult`, `ReviewResult`, `MediaResult`, `FileResult`, `ToolResult`, `LearningCandidate`, and `Escalation`
- structured threads for planning, work, review, escalation, and learning
- structural persistence plus semantic indexing so related prior outputs can be rediscovered

Required rules:
- Soma, Team Leads, specialist roles, automations, and MCP-backed systems publish governed outputs through the exchange model rather than raw ad hoc payloads when those outputs matter operationally
- team-local channels may still exist, but operator-reviewable outputs must remain discoverable through the managed exchange layer
- the managed exchange foundation stays inspectable in V8.1 before broad editing UI ships
- managed exchange is a runtime coordination substrate, not a replacement for bundle-driven startup or inheritance truth
- channels, threads, and exchange items carry explicit readers, writers, reviewers, participants, sensitivity classes, and downstream allowed-consumer metadata
- capability-producing outputs carry a capability id, risk class, trust class, and audit-ready publication metadata
- normalization into exchange does not imply unrestricted trust; MCP and external outputs remain bounded by trust classification and review requirements

## 6. Execution model

### 6.1 Interactive mode

- user-driven
- Soma is primary

### 6.2 Loop mode

- system-driven
- persistent
- triggered by event or schedule

V8.1 introduces Loop mode as an architecture contract before broad live execution.

## 7. System behavior rules

1. Loops never bypass policy.
2. Capabilities are never implied.
3. Agent Type Profiles always define baseline behavior.
4. Response Contract always applies.
5. Bundle-driven configuration remains the single source of truth.
6. Learning Loops may produce candidates and reviewed promotion decisions, but they never silently rewrite organization truth.
7. Semantic continuity supports recall and promotion lineage; it does not replace policy, engine, response, or capability contracts.

## 8. Operator UI architecture

### 8.1 Primary operator workspace

Soma remains the primary operating counterpart.

The primary workspace includes:
- Soma
- Team Lead
- Advisors
- Departments
- Recent Activity
- Learning & Context
- AI Engine Settings
- Response Style
- Automations

### 8.2 Automations surface

User-facing translations should stay simple:
- Automations
- Watchers
- Reviews

Each automation should show:
- name
- what it watches
- what it does
- how often it runs
- who owns it

The first shippable V8.1 state may keep this surface read-only.

### 8.3 Advanced UI boundaries

Advanced UI remains hidden by default.

Advanced-only detail may include:
- organization defaults and inheritance visibility
- Department overrides
- Specialist role bindings
- Loop Profile details
- detailed automation definitions
- managed exchange inspection for channels, threads, recent artifacts, and schema types
- Capability assignments
- Response Style inheritance
- bundle/config source truth
- deployment/env influence
- inheritance chains
- policy boundaries

Default UI must not leak architecture terms or raw control surfaces unnecessarily.

Even with advanced UI later:
- sensitive deployment secrets stay file/env/config driven
- host-specific runtime wiring stays file/env/config driven
- low-level provider auth and endpoints stay file/env/config driven where appropriate
- cluster or distributed node plumbing stays out of default V8.1 scope unless intentionally promoted later

## 9. Safety model

### 9.1 Required

- audit logs for loops
- capability enforcement
- approval gates for actuation loops
- managed exchange permission checks for read, write, review, and escalation paths
- capability risk classification for browser, MCP, media, file, and API output paths
- trust classification for internal tools, MCP services, external providers, and future remote execution nodes

### 9.2 Forbidden

- implicit execution
- hidden capabilities
- unrestricted overrides
- unrestricted cross-channel consumption just because an artifact exists

## 10. Testing requirements

### 10.1 Backend

- loop schema validation
- capability inheritance correctness
- policy enforcement
- memory-promotion path validation

### 10.2 Frontend

- automation visibility
- no architecture leakage
- inheritance clarity
- semantic continuity and promotion language stays user-facing

### 10.3 End-to-end

- create AI Organization
- view Automations
- inspect loop behavior without enabling execution yet
- confirm learning continuity remains bounded and inspectable

## 11. Release target

### 11.1 V8.1 initial release definition

#### Must have

- Loop Profiles defined in architecture and schema
- Learning Loop subtype defined in architecture and schema
- Runtime Capabilities contract defined
- Agent Type Profiles fully surfaced
- Response Contract integrated at organization and Agent Type level
- Memory Promotion and Semantic Continuity model defined
- Procedure / Skill Sets defined under Agent Type Profiles
- Soma workspace stable
- Automations UI visible, even if read-only

#### Must not

- expose unsafe execution
- allow unrestricted tool use
- break bundle-driven startup

### 11.2 First shippable state

- loops exist as configuration and inspectable architecture, not broad execution
- automations are visible in the workspace
- capabilities are defined but not fully exercised
- learning continuity architecture is defined even when raw/reviewed/promoted memory promotion is not fully implemented yet
- the system remains safe and inspectable

## 12. Post-release V8.2 targets

- first bounded Learning Loop implementation
- memory-promotion and procedure-skill retrieval runtime slices
- capability-based tool execution
- hardware-safe actuation

## 13. Delivery implication

This document is the canonical V8.1 architecture contract for:
- Loop Profiles
- Runtime Capabilities
- promoted Response Contract inheritance
- promoted Agent Type Profile runtime truth
- Learning Loops
- Memory Promotion and Semantic Continuity
- Procedure / Skill Sets
- read-only Automations visibility in the Soma workspace

Implementation that touches these areas must align with:
- `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`
- `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`
- `docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md`
- `V8_DEV_STATE.md`

## 14. Summary

V8.1 transforms Mycelis from a structured AI interface into a living, persistent, policy-bounded intelligence system without sacrificing:
- safety
- clarity
- reproducibility
