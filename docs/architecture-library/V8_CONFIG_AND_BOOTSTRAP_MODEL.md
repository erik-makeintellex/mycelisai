# V8 Config and Bootstrap Model

## Purpose

This document will define how V8 runtime concepts enter the system through configuration sources, templates, bootstrap resolution, scope inheritance, and precedence rules.

## Configuration sources

TBD in next bootstrap planning step.

## Bootstrap resolution flow

TBD in next bootstrap planning step.

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
