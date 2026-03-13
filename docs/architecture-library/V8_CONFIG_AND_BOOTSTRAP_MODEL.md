# V8 Config and Bootstrap Model

## Purpose

This document will define how V8 runtime concepts enter the system through configuration sources, templates, bootstrap resolution, scope inheritance, and precedence rules.

## Configuration sources

TBD in next bootstrap planning step.

## Bootstrap resolution flow

### Definition

Bootstrap resolution flow describes the staged transformation from configuration, template, and input sources into an active organization/runtime shape.

This section defines:
- the conceptual resolution order
- the major stages
- what becomes available downstream

It does not yet define detailed merge precedence or implementation details.

### High-level flow stages

1. Load source inputs.
2. Identify the target organization / Inception.
3. Resolve the Soma Kernel profile.
4. Resolve the Central Council structure.
5. Resolve team and agent defaults.
6. Resolve provider policy scope.
7. Resolve identity and continuity defaults.
8. Produce bootstrap-ready organization state.

### Team and agent resolution note

Bootstrap resolution must support:
- team-level defaults
- agent-level overrides
- scoped inheritance across organization layers

Resolution cannot stop at Council or team level only.

### Beginner vs advanced flow note

Beginner users may trigger this flow indirectly through a simple UI workflow.

Advanced users may shape the flow more directly through templates, configuration files, advanced panels, or APIs.

This keeps the platform approachable for new users while preserving deeper control where needed.

### Output of bootstrap resolution

Conceptually, bootstrap resolution should produce:
- a resolved organization shape
- resolved Team Lead / Soma Kernel behavior
- resolved Advisors / Council composition
- resolved Departments / Teams
- resolved Specialists / Agents
- resolved AI Engine Settings / Provider Policy scope
- resolved Memory & Personality / Identity and Continuity defaults

### Boundaries

This section does not yet define:
- detailed precedence rules
- exact merge algorithms
- storage ownership
- implementation-specific loader behavior

Those belong in later sections of the bootstrap model.

### Migration note

V7 already had partial bootstrap behavior through YAML, provider profiles, standing team configuration, runtime state, and operator flows, but it did not formalize a single V8 resolution flow for configurable AI organizations.

V8 promotes these pieces into an explicit staged bootstrap model.

## Scope inheritance

### Definition

Scope inheritance describes how configuration propagates from higher organizational scopes to more specific runtime scopes.

It defines which layers provide defaults, where narrower scopes may refine those defaults, and which forms of state should not flow downward as inherited configuration.

### Inheritance chain

The planned inheritance chain for V8 is:

```text
Inception
  -> Soma Kernel
  -> Central Council roles
  -> Team defaults
  -> Agent overrides
```

This means organization-wide defaults originate at the Inception level, become more specific at the Kernel and Council layers, and can then be narrowed further for teams and individual agents.

Team and agent configuration are first-class scopes in this model, not edge cases.

### What settings inherit

Settings that may inherit downward include:
- provider policy scope
- memory access posture
- reflection participation defaults
- tool availability and tool-policy defaults
- behavioral defaults that are intentionally configurable at lower scopes

These inherited settings provide a starting point for more specific runtime layers.

### Allowed overrides and policy blocks

Overrides are allowed only where the higher scope intentionally leaves a field configurable.

Typical allowed override cases include:
- team-specific provider choices within provider-policy constraints
- agent-specific tool access within approved tool-policy boundaries
- narrower memory or reflection posture when governance permits restriction
- specialization-specific behavioral tuning for teams or agents

Overrides are blocked when they would violate:
- higher-scope provider-policy constraints
- governance locks established at the Inception or Kernel level
- council-role restrictions that are meant to remain fixed for the organization
- organization-wide safety, compliance, or approval requirements

### Non-inheritable state

The following do not inherit as configuration:
- identity
- runtime state
- execution results
- single-run execution context
- ephemeral conversation state
- transient telemetry and task outputs

These belong to identity, execution, memory, or persistence layers and must be referenced or resolved separately rather than passed down as inherited config.

### Governance constraints

All overrides remain subordinate to governance.

A lower scope may narrow or specialize inherited settings only when the governing higher scope permits that change. Lower scopes cannot use overrides to escape policy, bypass approval boundaries, or silently widen capabilities that were intentionally constrained upstream.

### UI visibility note

Inheritance should be mostly invisible to beginners and surfaced only in advanced panels.

Beginner-facing views should present effective settings without forcing users to reason about every inheritance layer.

Advanced views may reveal where a value came from, which higher scope supplied it, and whether the current layer can override it.

### Boundaries

This section does not yet define exact precedence ordering when multiple candidate values exist at the same layer or across multiple source inputs.

That belongs in the `Precedence rules` section.

## Precedence rules

### Definition

Precedence rules describe how Mycelis chooses among competing candidate values after inheritance has established the available configuration layers.

This section defines conceptual resolution priority across sources and scopes. It does not define loader implementation details, merge algorithms, or storage-specific retrieval behavior.

### Source precedence

When multiple source types contribute candidate values, the intended conceptual source order is:

1. templates
2. static configuration
3. database or runtime state
4. operator input
5. derived context

This means later source classes may refine or replace earlier candidate values only where the active policy model allows that change.

### Scope precedence

Within the resolved organization shape, the intended conceptual scope order is:

```text
Inception
  -> Soma Kernel / Central Council role
  -> Team
  -> Agent
```

Lower scopes are more specific, but they do not automatically win. Their values apply only when the higher scope allows refinement for that field.

### Policy-before-override rule

Policy is evaluated before overrides are accepted.

A lower-scope or later-source value may refine configuration only after higher-scope governance, provider policy, and locked organization constraints have been checked. If a candidate value violates policy, it is not accepted merely because it appears later in the source or scope order.

### Conflict handling

When two candidate values conflict, Mycelis should conceptually resolve them by:
- identifying the active source class and scope for each value
- checking whether the field is inheritable and overrideable
- enforcing higher-scope governance and policy constraints first
- selecting the most specific allowed value
- falling back to the higher-governance value when the more specific value is not permitted

Conflicts should produce deterministic effective configuration, not silent ambiguity.

### Precedence vs inheritance

Inheritance determines which defaults flow downward through the organization model.

Precedence determines which candidate value becomes effective when multiple sources or scopes provide values for the same configurable field.

These are related but distinct concerns.

### Precedence vs loader behavior

This section does not define how files are read, how database state is fetched, how caches are consulted, or how runtime loaders are implemented.

Those concerns belong to later implementation-facing planning and backend work.

### Beginner vs advanced UX note

Beginner users should primarily see simple outcomes: the effective organization setup, effective AI engine settings, and working team behavior.

Advanced users may inspect the effective configuration in more detail, including which source or scope supplied a value and why a more specific override was accepted or rejected.

### Governance boundary

Lower scopes can refine allowed configuration, but they cannot bypass higher governance.

No later source or narrower scope may use precedence to escape Inception-level policy, Kernel-level governance, council-role restrictions, or locked safety/compliance requirements.

## Template and instantiation entry points

TBD in next bootstrap planning step.

## User Concept Layer

### Purpose

V8 introduces advanced internal architecture concepts that must be translated into approachable concepts for users.

The platform must support:
- simple user mental models
- powerful underlying architecture
- optional advanced configuration

Users should be able to operate the platform without needing to understand internal architecture terminology.

### User mental model

The primary mental model should feel like assembling and operating an AI team, not configuring a distributed AI runtime.

```text
AI Organization
  -> Team Lead
  -> Advisors
  -> Departments
  -> Specialists
```

### Concept translation table

| Architecture Concept | User Concept |
| --- | --- |
| Inception | AI Organization |
| Soma Kernel | Team Lead |
| Central Council | Advisors |
| Teams | Departments |
| Agents | Specialists |
| Provider Policy | AI Engine Settings |
| Identity / Continuity | Memory & Personality |

### Default user workflow

1. Create an AI Organization.
2. Choose a template.
3. Configure advisors if needed.
4. Add departments.
5. Add specialists.
6. Start interacting with the organization.

The system should allow immediate interaction without requiring deep configuration.

### Advanced configuration boundary

Advanced users may access deeper system capabilities when needed, including:
- provider routing policies
- scoped model overrides
- agent-specific configuration
- advanced templates
- infrastructure-level settings

These capabilities should be exposed through:
- advanced panels
- expert configuration views
- configuration files
- developer APIs

They should not be required for basic usage.

### Design rule

> The UI must prioritize clarity and usability over exposing internal architecture terminology.

Internal system concepts should remain implementation details unless the user explicitly enables advanced configuration.

## Migration from V7 bootstrap assumptions

TBD in next bootstrap planning step.

## Migration note

V7 runtime/config surfaces remain migration inputs.
This file is the new V8 planning surface for config/bootstrap behavior.
Implementation should not begin until the model is defined here.
