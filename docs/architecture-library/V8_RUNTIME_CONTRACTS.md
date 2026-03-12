# V8 Runtime Contracts

## Purpose

This document will define the canonical V8 runtime contracts for:
- Inception
- Soma Kernel
- Central Council
- Provider Policy
- Identity / Continuity State

Migration note:
- V7 architecture-library docs remain migration inputs.
- This file is the new V8 contract surface being introduced incrementally.

## Inception

### 1. Definition

Inception is the top-level configured AI organization instance in V8.

It defines the organizational frame that gives downstream runtime components their identity, governance posture, provider-policy posture, memory posture, and operator relationship model.

### 2. Ownership

An Inception owns or defines the high-level organizational contract for:
- purpose
- organization type
- governance mode
- kernel profile reference
- council configuration reference
- specialist team structure reference
- provider policy reference
- memory mode
- reflection mode
- operator relationship posture

This is the layer that answers what kind of AI organization is being instantiated and what downstream contracts it is allowed to configure.

### 3. Boundaries

An Inception is not a mission.
- a mission is a bounded objective or execution path inside an already-defined organization

An Inception is not a run.
- a run is a single execution instance inside the organization configured by the Inception

An Inception is not a template.
- a template is a reusable pattern or starting point that can be used to create an Inception, but it is not itself a live configured organization instance

An Inception is not a profile.
- a profile is a narrower routing or behavior configuration within the system, while an Inception is the top-level organizational contract that determines which profiles and structures are in scope

### 4. Minimal required fields

At minimum, an Inception contract must carry:
- stable inception identity
- human-readable name
- purpose statement
- organization type
- governance mode
- kernel profile reference
- council configuration reference
- specialist team structure reference
- provider policy reference
- memory mode
- reflection mode
- operator relationship posture
- lifecycle state

These fields are structural contract requirements, not final schema definitions.

### 5. Lifecycle position

Inception sits at the top of the V8 runtime hierarchy.

It configures what comes downstream:
- Soma Kernel identity and posture
- Central Council structure
- specialist team structure
- provider-policy boundaries
- memory and reflection posture

The runtime should resolve an Inception before resolving kernel, council, team, or execution-specific behavior.

### 6. Initial migration note

This differs from the old V7 assumption where one fixed Soma and one fixed standing Council were treated as the default runtime truth.

In V8, the organization is configured first through Inception, and Soma Kernel plus Central Council are resolved as parts of that configured organization rather than as globally fixed assumptions.

## Soma Kernel

### 1. Definition

Soma Kernel is the central coordination and runtime layer for an Inception in V8.

It is the layer that holds the active coordinating posture for operator interaction, orchestrates when and how the council is engaged, and maintains the live execution frame for the configured organization.

### 2. Responsibilities

Soma Kernel is responsible for:
- operator interaction coordination
- council consultation orchestration
- goal-state handling
- task-routing authority
- continuity ownership
- reflection coordination
- execution-context ownership

This is the runtime layer that turns the Inception contract into active coordinating behavior.

### 3. Configurable aspects

Per Inception, Soma Kernel should be configurable in at least these ways:
- interaction posture
- consultation behavior
- routing posture
- continuity mode
- reflection participation
- governance constraints

These are contract-level configuration surfaces, not final implementation settings.

### 4. Boundaries

Soma Kernel is not the old fixed Soma identity.
- the old V7 framing treated Soma/admin-core as a universal built-in truth, while V8 treats the kernel as a configurable coordination layer resolved inside an Inception

Soma Kernel is not the Central Council.
- the council is an advisory or specialist structure that Soma Kernel may consult or orchestrate, but it is not the kernel itself

Soma Kernel is not a specialist team.
- specialist teams perform scoped execution roles under the organization, while Soma Kernel remains the central coordination layer above them

Soma Kernel is not a run.
- a run is a bounded execution instance governed within the context set by the kernel

Soma Kernel is not provider policy.
- provider policy defines model and routing constraints, while Soma Kernel uses those constraints during coordination and task routing

### 5. Runtime position

Soma Kernel sits directly beneath Inception in the V8 runtime hierarchy.

It governs downstream coordination behavior for:
- operator-facing interaction flow
- council engagement pathing
- task-routing posture
- execution-context handling
- reflection participation

The runtime should resolve Soma Kernel after Inception and before council, specialist-team, or run-level behavior is executed.

### 6. Initial migration note

This differs from the V7 assumption that Soma/admin-core is a universal built-in truth for every deployment.

In V8, Soma Kernel is a configured runtime layer that belongs to an Inception, not a globally fixed identity that exists independently of organizational configuration.

## Central Council

### 1. Definition

Central Council is the configurable advisory and reasoning structure that supports the Soma Kernel within an Inception.

It provides structured perspective, planning support, and domain-informed review so the kernel can govern coordination with more than one reasoning lens when the Inception requires it.

### 2. Responsibilities

Central Council is responsible for:
- advisory reasoning for Soma Kernel
- domain-specific expertise contributions
- consultation during planning or decision stages
- strategic review of proposed actions
- optional reflective or evaluative feedback

The council contributes perspective and judgment to the organization, but it is not the primary execution layer.

### 3. Configurable composition

Council composition is configurable per Inception.

Possible council roles may include:
- Architecture Lead
- Engineering Lead
- Security Lead
- Product Strategy Lead
- Research Analyst
- Ethics Advisor

Different organization types may require different council compositions. A technical delivery organization, a research organization, and a governance-sensitive organization should not be forced into one fixed advisory pattern.

### 4. Boundaries

Central Council is not the Soma Kernel.
- the kernel remains the central coordination layer and decides when council consultation is needed

Central Council is not a specialist team or execution agent.
- specialist teams and execution agents perform scoped work, while council members provide advisory reasoning and review

Central Council is not a run.
- a run is a bounded execution instance that may consult the council, but the council is an organizational advisory layer rather than a single execution instance

Central Council is not provider routing logic.
- provider routing logic defines model and policy selection constraints, while the council provides reasoning within those constraints

Council members advise but do not directly execute work as the contract baseline.

### 5. Runtime relationship

The runtime hierarchy is:
- Inception configures the organization
- Soma Kernel coordinates the organization at runtime
- Central Council provides advisory support to the kernel
- Specialist Teams carry execution work downstream

The council operates as an advisory layer consulted by the kernel, not as the primary runtime owner or execution surface.

### 6. Migration note

In V7, the council role set of architect, coder, creative, and sentry was effectively treated as a fixed system pattern.

In V8, council structures must be configurable and organization-specific so each Inception can define the advisory composition it actually needs.

## Provider Policy

### 1. Definition

Provider Policy is the V8 contract layer that governs how an Inception chooses, constrains, and routes model backends and related execution configuration.

It defines the approved provider scope, routing rules, and configuration inheritance model for an Inception, including how different organizational roles can use local or hosted model APIs.

### 2. Responsibilities

Provider Policy is responsible for:
- defining allowed providers for an Inception
- governing local vs hosted provider usage
- supporting role-based routing
- supporting council-level, team-level, and agent-level routing
- enforcing governance boundaries for model selection
- enabling hybrid deployments
- defining how scoped defaults and overrides work

This contract turns provider support from an implementation capability into an explicit organizational policy surface.

### 3. Supported provider classes

Provider Policy must support both provider classes below.

#### Local / open-source classes

Examples:
- Ollama
- LM Studio
- vLLM
- custom OpenAI-compatible local endpoints

#### Hosted / commercial classes

Examples:
- OpenAI APIs
- Anthropic / Claude APIs
- Gemini APIs
- other OpenAI-compatible hosted APIs

The contract is provider-class oriented and must not be locked to only one vendor family.

### 4. Configuration scope layers

Provider and execution-related configuration must be possible at these scopes:

#### Inception-level
Applies to the whole organization.

#### Soma Kernel-level
Applies to kernel defaults.

#### Central Council role-level
Applies to council roles or role classes.

#### Team-level
Applies to a specialist team as a whole.

#### Agent-level
Applies to an individual agent instance inside a team.

Team-level and agent-level configuration are first-class needs, not edge cases.

### 5. Inheritance and override rules

The general inheritance model is:

```text
Inception defaults
  -> kernel / council role defaults
  -> team defaults
  -> agent overrides
```

Inheritance is expected.

Lower scopes may override only allowed fields.

Overrides must remain policy-governed, and lower-scope configuration cannot violate higher-scope governance constraints.

### 6. Routing scope

Routing may be applied at these levels:
- Inception-level routing
- Soma Kernel defaults
- Central Council role routing
- specialist team routing
- individual agent routing
- task- or capability-sensitive routing where policy allows

This allows routing posture to reflect organization-wide defaults while still supporting scoped execution needs when governance permits it.

### 7. Boundaries

Provider Policy is not the Inception itself.
- the Inception defines the organization as a whole, while Provider Policy governs allowed provider and routing configuration inside that organization

Provider Policy is not Soma Kernel responsibilities.
- the kernel coordinates runtime behavior, while Provider Policy governs which model/routing configurations that coordination may use

Provider Policy is not Central Council responsibilities.
- the council provides advisory reasoning, while Provider Policy constrains model selection and routing posture

Provider Policy is not specialist team definitions.
- team definitions describe execution structure, while Provider Policy governs the allowed and intended backend configuration available to those teams

Provider Policy is not runtime execution results.
- it governs allowed and intended routing/configuration, not the outcome of a particular run or execution path

### 8. Migration note

V7 already had multi-provider routing and profile infrastructure, but it was not yet fully framed as a first-class configurable contract of an Inception.

V8 promotes provider routing from operational capability to explicit organizational policy, including scoped team-level and agent-level configuration.

## Identity and Continuity State

TBD in next contract step.




