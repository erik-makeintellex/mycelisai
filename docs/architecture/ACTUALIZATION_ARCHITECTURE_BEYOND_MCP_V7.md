# Actualization Architecture Beyond MCP V7

Version: `1.0`
Status: `Authoritative`
Last Updated: `2026-02-28`
Scope: Multi-protocol actualization strategy beyond MCP with local-first defaults

This document extends Mycelis architecture from MCP-only integration to a universal action and agent-interoperability strategy.
MCP remains foundational, but it is no longer the only integration axis.
Security companion: `docs/architecture/SECURE_GATEWAY_REMOTE_ACTUATION_PROFILE_V7.md`.
Soma cognition companion: `docs/architecture/SOMA_SYMBIOTE_GROWTH_AND_HOST_ACTUATION_V7.md`.

---

## Table of Contents

1. Why This Matters
2. Research Summary (External Systems)
3. Canonical Protocol Stack for Mycelis
4. OpenClaw-Inspired Gateway Patterns
5. Universal Action + Dynamic Service Architecture
6. Python-Bound Management Interface
7. External UI and Client Surface Strategy
8. Governance and Security Boundaries
9. Testing and Validation Matrix
10. Parallel Delivery Plan
11. Delivery Gates
12. Timeline Fit in V7 Development

---

## 1. Why This Matters

Mycelis must actualize workflows across:
- tool protocols (MCP and others)
- HTTP-native service ecosystems
- Python-native action ecosystems
- external agent ecosystems
- operator-facing client surfaces

If we stay protocol-singleton, onboarding and scaling degrade over time.

---

## 2. Research Summary (External Systems)

### 2.1 OpenClaw architecture patterns worth adopting

From OpenClaw documentation and repository:
- a single long-lived gateway process acts as control plane
- typed WebSocket API with strict handshake and role/scopes
- protocol typing generated from TypeBox + JSON Schema
- local-first operation model with remote tunnel options
- ACP bridge for IDE integration over stdio

Mycelis inference:
- adopting a typed gateway contract and strict connect handshake will improve reliability and client interoperability for our own bus/control paths.

### 2.2 Standards beyond MCP

- OpenAPI: strongest standard for HTTP capability discovery and invocation contracts
- JSON Schema: canonical contract language for action input/output validation
- A2A: complementary protocol for agent-to-agent collaboration at peer level
- ACP: useful client protocol for IDE-native agent control

Mycelis inference:
- MCP should handle tool/resource integration.
- A2A should handle external agent delegation/collaboration.
- OpenAPI + JSON Schema should handle service registration and typed action execution.

### 2.3 Platform ecosystems

- Open WebUI and LibreChat are strong self-hosted multi-provider operator clients
- both support local/self-hosted model patterns and extensibility
- these should be treated as optional client surfaces, not governance authorities

---

## 3. Canonical Protocol Stack for Mycelis

Use a layered stack:

1. **Action Contract Layer**
- JSON Schema for inputs/outputs
- OpenAPI for HTTP service capabilities

2. **Tool/Resource Layer**
- MCP adapter for MCP servers and tools

3. **Agent Collaboration Layer**
- A2A adapter for external agent peer collaboration

4. **Client Integration Layer**
- ACP bridge for IDE agent clients
- native Mycelis UI + optional external UIs (Open WebUI/LibreChat)

5. **Execution Governance Layer**
- Mycelis remains policy + approval + audit authority

---

## 4. OpenClaw-Inspired Gateway Patterns

Patterns to adopt in Mycelis control interfaces:

1. Single authoritative gateway process per deployment boundary.
2. Typed wire protocol with version negotiation.
3. Mandatory first-frame handshake for WS-like channels.
4. Explicit role + scope declaration at connection time.
5. Idempotency keys for side-effecting operations.
6. Device identity and pairing model for remote nodes/clients.

These patterns apply to Mycelis external control bridges and future node protocols.

---

## 5. Universal Action + Dynamic Service Architecture

This document is additive to:
- `docs/architecture/UNIVERSAL_ACTION_INTERFACE_V7.md`
- `docs/architecture/MCP_SERVICE_CONFIGURATION_LOCAL_FIRST_V7.md`

Required execution model:

1. Register service dynamically via governed API.
2. Derive actions into a normalized action registry.
3. Validate schemas (JSON Schema).
4. Execute through adapter (`mcp|openapi|python|a2a`).
5. Emit mission events and link to `run_id`.

Adapter contract requirements:
- deterministic error mapping
- duration + provenance metadata
- degraded-state recovery hooks

---

## 6. Python-Bound Management Interface

A dedicated Python manager is required where Python ecosystems dominate.

### 6.1 Responsibilities

- environment lifecycle (`uv` managed)
- dependency pinning and reproducible installs
- action export and schema sync
- health and capability probes

### 6.2 Service model

- each Python service has explicit runtime metadata
- all callable actions are promoted into universal action registry
- policy and risk classification happen before enablement

### 6.3 Framework posture

- PydanticAI: preferred for typed tools/schemas and production reliability
- CrewAI: supported for multi-agent workflow composition when wrapped as typed actions

---

## 7. External UI and Client Surface Strategy

1. Mycelis UI remains canonical governance/operations surface.
2. Open WebUI/LibreChat can be integrated as optional operator endpoints.
3. External surfaces must call governed Mycelis APIs for mutations.
4. No external client can bypass policy/approval/event audit paths.

Integration principle:
- "bring your interface" is allowed
- "bypass governance" is not allowed

---

## 8. Governance and Security Boundaries

1. All dynamic service changes are high-risk governed operations.
2. Remote service defaults are disabled unless explicitly approved.
3. All action invocations must carry provenance context (`run_id`, `team_id`, `agent_id`, `origin`).
4. Typed schema validation is mandatory before execution.
5. All side effects must emit persistent mission events before transport publish.

Team lifetime strategy:
- short-lived teams for bounded direct action plans
- long-lived teams for scheduled recurrence, sensor ingestion, and low-level IoT control loops
- operator request for repeat execution triggers promotion path from short-lived to long-lived team profile

---

## 9. Testing and Validation Matrix

### Unit
- adapter contract validation
- schema compile/validation tests
- risk classification tests

### Integration
- service add/probe/sync lifecycle
- adapter execution path (`mcp|openapi|python|a2a`)
- gateway handshake and scope enforcement tests

### API
- dynamic service endpoints
- universal action invoke endpoint
- backward compatibility for legacy MCP endpoints

### UI/E2E
- install local service and invoke action
- degraded recovery for each adapter class
- optional external client route proving governance preservation

---

## 10. Parallel Delivery Plan

- Team Forge: dynamic service registry + universal gateway API
- Team Helios: OpenAPI/JSON Schema normalization and action catalog
- Team Pyra: Python runtime manager and schema export
- Team Circuit: control-plane reliability + handshake/idempotency
- Team Sentinel: cross-adapter regression and release gates

---

## 11. Delivery Gates

Gate 1 - Protocol contracts finalized.
Gate 2 - MCP + OpenAPI adapters production-ready.
Gate 3 - Python manager operational with schema sync.
Gate 4 - A2A/ACP bridges validated for interoperability.
Gate 5 - Governance-preserving external client integrations validated.

## 12. Timeline Fit in V7 Development

This expansion is intentionally staged after core V7 scheduler/runs stabilization:
- Stage A: complete core scheduler + causal chain UX (current critical path).
- Stage B: ship universal action registry and dynamic service onboarding APIs.
- Stage C: add OpenAPI + Python adapters under governance controls.
- Stage D: pilot A2A/ACP bridges with strict scope and handshake enforcement.
- Stage E: enable remote actuation only under secure private mesh profile.

Rationale:
- preserves delivery momentum on existing V7 commitments
- introduces protocol expansion without destabilizing core mission execution
- keeps high-risk remote actuation behind explicit security gates

---

End of document.



