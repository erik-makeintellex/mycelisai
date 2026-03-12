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

TBD in next bootstrap planning step.

## Precedence rules

TBD in next bootstrap planning step.

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
