# V8.1 Living Organization Architecture

> Status: Canonical V8.1 PRD
> Last Updated: 2026-03-19
> Purpose: Define the first implementation-grade V8.1 architecture for persistent execution, bounded automation, runtime capabilities, and reproducible specialist behavior.
> Depends On: `docs/architecture-library/V8_RUNTIME_CONTRACTS.md`, `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`, `docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md`

## 1. Why V8.1 exists

V8 established the AI Organization, Team Lead-first workspace, bundle-driven startup truth, and bounded operator settings.

V8.1 extends that work from a structured interactive system into a living, persistent, policy-bounded intelligence runtime.

V8.1 introduces:
- Loop Profiles as the bounded execution layer
- Runtime Capabilities as the bounded action layer
- Response Contract as a first-class inheritance contract
- Agent Type Profiles as a first-class runtime layer between Team defaults and Agent instances

This enables:
- continuous but governed operation
- event-driven workflows
- safe automation
- reproducible specialist behavior

## 2. Product goals

### 2.1 Primary goal

Enable AI Organizations to:
- operate continuously
- react to events
- supervise internal and external processes
- maintain deterministic behavior through policy and inheritance

### 2.2 Secondary goals

- maintain strict Team Lead-first UX discipline
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
      -> Agent Instances
  -> Loop Profiles
  -> Runs
  -> Events
  -> Memory
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
- Actuation Loop
  - trigger: schedule or event
  - example: hardware or system action
  - posture: restricted and policy-gated

#### Initial schema contract

```yaml
id: string
name: string
type: scheduled | event | review | actuation

owner:
  type: team | agent_type
  id: string

trigger:
  schedule: optional cron
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
```

#### Required rules

- loops never bypass policy
- loops never imply capabilities
- actuation loops always require stronger gating than passive review loops
- loops must remain auditable even when not yet executable

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

## 6. Execution model

### 6.1 Interactive mode

- user-driven
- Team Lead is primary

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

## 8. Operator UI architecture

### 8.1 Primary operator workspace

The Team Lead remains the primary operating counterpart.

The primary workspace includes:
- Team Lead
- Advisors
- Departments
- AI Engine Settings
- Response Style
- Memory & Personality
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
- Loop Profile details
- Capability assignments
- inheritance chains
- policy boundaries

Default UI must not leak architecture terms or raw control surfaces unnecessarily.

## 9. Safety model

### 9.1 Required

- audit logs for loops
- capability enforcement
- approval gates for actuation loops

### 9.2 Forbidden

- implicit execution
- hidden capabilities
- unrestricted overrides

## 10. Testing requirements

### 10.1 Backend

- loop schema validation
- capability inheritance correctness
- policy enforcement

### 10.2 Frontend

- automation visibility
- no architecture leakage
- inheritance clarity

### 10.3 End-to-end

- create AI Organization
- view Automations
- inspect loop behavior without enabling execution yet

## 11. Release target

### 11.1 V8.1 initial release definition

#### Must have

- Loop Profiles defined in architecture and schema
- Runtime Capabilities contract defined
- Agent Type Profiles fully surfaced
- Response Contract integrated at organization and Agent Type level
- Team Lead workspace stable
- Automations UI visible, even if read-only

#### Must not

- expose unsafe execution
- allow unrestricted tool use
- break bundle-driven startup

### 11.2 First shippable state

- loops exist as configuration and inspectable architecture, not broad execution
- automations are visible in the workspace
- capabilities are defined but not fully exercised
- the system remains safe and inspectable

## 12. Post-release V8.2 targets

- first Review Loop execution
- capability-based tool execution
- event-driven loop triggers
- hardware-safe actuation

## 13. Delivery implication

This document is the canonical V8.1 architecture contract for:
- Loop Profiles
- Runtime Capabilities
- promoted Response Contract inheritance
- promoted Agent Type Profile runtime truth
- read-only Automations visibility in the Team Lead workspace

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
