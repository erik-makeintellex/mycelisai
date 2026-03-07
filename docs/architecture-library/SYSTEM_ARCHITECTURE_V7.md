# System Architecture V7

> Status: Canonical
> Last Updated: 2026-03-07
> Scope: Platform structure, runtime layers, storage, bus contracts, deployment posture, and supporting implementation authorities.

Related:
- [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)
- [Target Deliverable V7](TARGET_DELIVERABLE_V7.md)
- [Execution And Manifest Library V7](EXECUTION_AND_MANIFEST_LIBRARY_V7.md)
- [UI And Operator Experience V7](UI_AND_OPERATOR_EXPERIENCE_V7.md)
- [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md)

Supporting specialized docs:
- [Overview](../architecture/OVERVIEW.md)
- [Backend](../architecture/BACKEND.md)
- [Frontend](../architecture/FRONTEND.md)
- [Operations](../architecture/OPERATIONS.md)
- [NATS Signal Standard V7](../architecture/NATS_SIGNAL_STANDARD_V7.md)

## 1. Architecture Goal

The system architecture exists to support governed execution with durable lineage and a clear operator experience.

It must support:
- direct answers
- governed proposals
- workflow execution
- scheduled and event-driven plans
- persistent active plans
- runtime recovery across local and Linux/K8s deployment

## 2. Core Layers

### 2.1 Intent layer

Ingress surfaces:
- Workspace UI
- REST API
- trigger rules
- scheduler
- event-driven subscriptions

Responsibility:
- capture intent with enough metadata to route execution and preserve origin

### 2.2 Orchestration layer

Components:
- Soma/admin
- standing teams
- council specialists
- provider routing
- governance policy

Responsibility:
- decide whether the request should become an answer, proposal, execution, or blocker

### 2.3 Execution layer

Components:
- internal tools
- MCP tools
- local command allowlist
- filesystem/artifact services
- team manifestation and delegation

Responsibility:
- perform bounded, governable work tied to run identity

### 2.4 Event and signal layer

Components:
- mission events
- run lineage
- canonical NATS subjects
- signal envelopes

Responsibility:
- preserve transient communication plus durable audit history

### 2.5 Observability and operator layer

Components:
- run timeline
- conversation log
- status and degraded-state surfaces
- system checks
- docs and workflow onboarding

Responsibility:
- make the system understandable and operable

## 3. Persistence Model

Persistent system-of-record data belongs in Postgres.

Key durable objects:
- runs
- mission events
- conversation turns
- trigger rules
- scheduled missions
- manifests and definitions
- artifacts metadata
- user and policy state

High-volume or transient transport does not replace persistence.

Rule:
- mutating actions require durable event lineage in addition to bus traffic

## 4. Filesystem And Artifact Storage

The platform must support manifested material on persistent mounted storage.

Rules:
- `MYCELIS_WORKSPACE` is the authoritative workspace root
- `DATA_DIR` is the artifact/blob root
- local development may use workspace-local paths
- Linux/K8s deployments must mount persistent storage and map workspace/artifacts into that mount

This exists so created material is not trapped in ephemeral containers or transient local paths.

## 5. Bus And Signal Posture

Canonical product communication uses NATS with standardized subject families and source metadata.

Rules:
- product subjects use constants, not inline literals
- operator status is separate from high-volume telemetry
- development-only infrastructure subjects are not promoted into canonical orchestration

Core families:
- `internal.command`
- `signal.status`
- `signal.result`
- `telemetry`
- request-reply specialist channels

Detailed authority:
- [NATS Signal Standard V7](../architecture/NATS_SIGNAL_STANDARD_V7.md)

## 6. Cross-Platform Operations Posture

The system must operate on:
- local Windows development
- Linux/K8s deployment

Rules:
- app management logic is Python-owned
- PowerShell may wrap local host behavior but must not own orchestration logic
- docs and tasks must use the same runner contract
- startup and teardown must be bounded and observable

Detailed authority:
- [Operations](../architecture/OPERATIONS.md)

## 7. Supporting Runtime Contracts

### 7.1 Team coordination

- standing teams are manifest-backed
- central coordination uses canonical team NATS lanes
- runtime-created ephemeral teams must still obey canonical subject and metadata rules

### 7.2 Provider and execution routing

- provider routing is explicit
- runtime must know what model/provider executed an outcome
- unavailable external providers must fail clearly, not silently

### 7.3 Governance

- high-impact mutations require proposal/approval handling
- execution path selection is explicit
- governance is part of the architecture, not a separate overlay

## 8. Architecture Optimization Rule

The architecture should optimize for:
- clear execution lineage
- minimal duplicated contracts
- stable runtime recovery
- maintainable modular docs
- UI as an execution guide, not a raw telemetry wall

It should not optimize for:
- exposing every subsystem by default
- creating alternate undocumented execution paths
- embedding host-specific scripting assumptions into core ops

## 9. Where To Go Next

Use:
- [Execution And Manifest Library V7](EXECUTION_AND_MANIFEST_LIBRARY_V7.md) for runs, plans, and lifecycle behavior
- [UI And Operator Experience V7](UI_AND_OPERATOR_EXPERIENCE_V7.md) for operator-facing structure
- [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md) for acceptance proof
