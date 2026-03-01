# Universal Action Interface V7

Version: `1.0`
Status: `Authoritative`
Last Updated: `2026-02-28`
Scope: Universal action contract, dynamic service onboarding API, and Python management interface

This document defines how Mycelis exposes one universal action plane across MCP, OpenAPI-described services, and Python-bound capabilities.
Default posture is local/self-hosted first.
Companion landscape spec: `docs/architecture/ACTUALIZATION_ARCHITECTURE_BEYOND_MCP_V7.md` (multi-protocol expansion including A2A/ACP and OpenClaw-inspired control patterns).
Security profile: `docs/architecture/SECURE_GATEWAY_REMOTE_ACTUATION_PROFILE_V7.md`.
Hardware profile: `docs/architecture/HARDWARE_INTERFACE_API_AND_CHANNELS_V7.md`.
Soma growth profile: `docs/architecture/SOMA_SYMBIOTE_GROWTH_AND_HOST_ACTUATION_V7.md`.

---

## Table of Contents

1. Why Universal Action Interface
2. Design Principles
3. Action Plane Architecture
4. Canonical Action Contract
5. Dynamic Service Configuration API
6. Adapter Model (MCP/OpenAPI/Python)
7. Python Management Interface
8. Self-Hosted Platform Integration
9. Governance and Safety Boundaries
10. Testing and Validation Matrix
11. Parallel Delivery Plan
12. Delivery Gates
13. Timeline Fit and Rollout

---

## 1. Why Universal Action Interface

Mycelis currently has strong MCP and internal tool execution.
To scale channels without UI/API drift, we need one universal interface where:
- actions are declared once
- schemas are machine-readable (JSON Schema/OpenAPI)
- services can be added dynamically through governed APIs
- execution remains traceable to `run_id`

Outcome:
- MCP is preserved as a first-class adapter, not the only tool protocol.
- Python-native and API-native tools become composable under one contract.

---

## 2. Design Principles

1. Local-first and self-hosted by default.
2. One action contract regardless of backend adapter.
3. OpenAPI and JSON Schema as interoperability standards.
4. Governance before irreversible side effects.
5. Degraded states must be recoverable in-place.
6. Additive extensibility: new adapters without breaking existing MCP flows.

---

## 3. Action Plane Architecture

### 3.1 Layers

1. `Action Registry`
   - stores service + action metadata
   - validates schemas
   - enforces uniqueness and lifecycle state

2. `Action Gateway`
   - unified invoke endpoint
   - injects execution context (`run_id`, `team_id`, `agent_id`)
   - emits mission events

3. `Adapter Runtime`
   - MCP Adapter
   - OpenAPI Adapter
   - Python Adapter

4. `Observability + Governance`
   - event emission (`tool.invoked`, `tool.completed`, `tool.failed`)
   - policy checks and approval routing

### 3.2 Action Flow

1. User/agent selects action.
2. Gateway resolves action from registry.
3. Adapter executes target call.
3.1. Route selector chooses direct Soma execution or team manifestation path.
3.2. Team lifecycle policy applies (`ephemeral|persistent|auto`) with schedule/IoT promotion support.
4. Output normalized to universal result envelope.
5. Events persisted and streamed.

---

## 4. Canonical Action Contract

### 4.1 Action Definition

```json
{
  "action_id": "filesystem.read_file",
  "service_id": "mcp.filesystem",
  "adapter": "mcp|openapi|python",
  "display_name": "Read File",
  "description": "Read a workspace file",
  "input_schema": {},
  "output_schema": {},
  "risk_level": "low|medium|high",
  "idempotent": true,
  "locality": "local|remote",
  "enabled": true,
  "team_lifetime_hint": "ephemeral|persistent|auto",
  "version": "v1"
}
```

### 4.2 Universal Invoke Request

```json
{
  "arguments": {},
  "context": {
    "run_id": "",
    "team_id": "",
    "agent_id": "",
    "origin": "workspace|trigger|schedule|api|sensor",
    "governance_mode": "passive|approval_required|halted"
  }
}
```

### 4.3 Universal Invoke Response

```json
{
  "ok": true,
  "data": {
    "status": "success|error",
    "content": {},
    "adapter": "mcp|openapi|python",
    "duration_ms": 0
  },
  "error": "",
  "meta": {
    "run_id": "",
    "action_id": "",
    "service_id": "",
    "timestamp": ""
  }
}
```

---

## 5. Dynamic Service Configuration API

### 5.1 Service Registry Endpoints (New)

- `GET /api/v1/actions/services`
  - list registered services and health
- `POST /api/v1/actions/services`
  - add service dynamically (governed)
- `PATCH /api/v1/actions/services/{service_id}`
  - update config, enable/disable, rotate metadata
- `DELETE /api/v1/actions/services/{service_id}`
  - remove service
- `POST /api/v1/actions/services/{service_id}/probe`
  - runtime reachability and schema probe
- `POST /api/v1/actions/services/{service_id}/sync`
  - re-sync actions from provider

### 5.2 Action Endpoints (New)

- `GET /api/v1/actions`
  - list all actions with filters (`adapter`, `risk`, `locality`, `enabled`)
- `GET /api/v1/actions/{action_id}`
  - detail + schemas
- `POST /api/v1/actions/{action_id}/invoke`
  - universal execution path

### 5.3 Backward Compatibility

- Existing MCP endpoints remain supported.
- MCP registrations are mirrored into universal registry records.
- New UI surfaces should consume universal endpoints first.

---

## 6. Adapter Model (MCP/OpenAPI/Python)

### 6.1 MCP Adapter

- maps MCP tools into `ActionDefinition`
- uses existing ToolExecutorAdapter and pool
- preserves local-first MCP install policy

### 6.2 OpenAPI Adapter

- imports OpenAPI specs and derives callable actions
- request/response contracts mapped to JSON Schema
- supports auth metadata without exposing secrets in UI

### 6.3 Python Adapter

- executes registered Python actions through managed runtime
- action signatures exported as JSON Schema
- supports synchronous and async execution modes

---

## 7. Python Management Interface

A dedicated management interface is recommended for Python-bound action ecosystems.

### 7.1 Why Dedicated Python Interface

- Python dependency management differs from JS/Go runtimes
- isolation and reproducibility require explicit environment control
- frameworks like PydanticAI/CrewAI need controlled orchestration boundaries

### 7.2 Python Service Manager (Proposed)

- `Python Runtime Manager` service (self-hosted)
- environment strategy:
  - `uv`-managed per-service environment
  - pinned dependency lock
  - explicit runtime metadata in registry

### 7.3 Python Manager Endpoints (New)

- `GET /api/v1/python/services`
- `POST /api/v1/python/services`
- `POST /api/v1/python/services/{id}/install`
- `POST /api/v1/python/services/{id}/sync-actions`
- `GET /api/v1/python/services/{id}/health`

### 7.4 Framework Mapping

- PydanticAI: preferred for typed production action definitions
- CrewAI: supported for multi-agent workflows when exposed as actions
- Cline/IDE tools: treated as external client layer, not core execution runtime

### 7.5 Scheduled Repeat and Low-Level IoT Support

The universal action plane must support:
- converting one-off user actions into scheduled-repeat workflows when requested
- routing low-level IoT/watch-control objectives to persistent teams by default
- retaining direct Soma execution for safe one-off actions when team manifestation is unnecessary
- explicit operator override to force `ephemeral` or `persistent` profile

---

## 8. Self-Hosted Platform Integration

### 8.1 Control Surface Positioning

- Mycelis remains orchestration authority and governance plane.
- Open WebUI / LibreChat are optional client interfaces.
- Ollama remains primary local model runtime.

### 8.2 Integration Strategy

- expose Mycelis chat/action APIs with OpenAI-compatible compatibility where useful
- keep canonical action execution in Mycelis universal action gateway
- avoid dual authority for governance decisions

### 8.3 Local-Default Profiles

Reference default stack:
1. Ollama local
2. Mycelis Core + NATS + Postgres
3. MCP local services (filesystem/fetch/memory)
4. Optional Python runtime manager local
5. Optional Open WebUI/LibreChat as external UI clients

---

## 9. Governance and Safety Boundaries

1. Service add/update/delete are governed mutations.
2. High-risk actions require explicit confirmation.
3. Remote service onboarding defaults to high risk.
4. Action-level policies can deny/require approval/allow.
5. All action executions must emit mission events.

---

## 10. Testing and Validation Matrix

### Unit
- action schema validation
- adapter mapping tests (MCP/OpenAPI/Python)
- policy classification tests

### Integration
- dynamic service add/probe/sync lifecycle
- universal invoke execution path and event linkage
- Python manager environment lifecycle tests

### API
- endpoint contract tests for `/api/v1/actions/*`
- backward compatibility tests for existing MCP endpoints

### UI/E2E
- add local service from UI and invoke action
- degraded service recovery from status surfaces
- action detail schema rendering correctness

Recommended existing baseline commands:

```bash
cd core && go test ./internal/mcp/ -count=1
cd core && go test ./internal/server/ -run TestHandleMCP -count=1
cd interface && npm run build
```

---

## 11. Parallel Delivery Plan

- Team Forge: action registry + gateway contracts
- Team Helios: OpenAPI/JSON Schema normalization and UI adapters
- Team Circuit: reliability/degraded recovery UX
- Team Pyra: Python runtime manager and action sync
- Team Sentinel: cross-adapter regression gates

---

## 12. Delivery Gates

Gate 1 - Contract:
- universal action DTOs finalized
- API spec published

Gate 2 - Adapter:
- MCP adapter mapped into universal registry
- OpenAPI and Python adapters scaffolded

Gate 3 - Execution:
- universal invoke path operational with event linkage

Gate 4 - Governance:
- risk policy and approval flows validated

Gate 5 - Production:
- full test matrix green and docs in sync

## 13. Timeline Fit and Rollout

This architecture lands in rollout waves tied to current V7 delivery:
1. Wave 0: Team C scheduler completion + lifecycle contract wiring.
2. Wave 1: Team E/D UX completion for runs, chain, and degraded clarity.
3. Wave 2: universal action registry and service APIs scaffolded.
4. Wave 3: secure remote-actuation gateway controls enabled in private mesh mode only.
5. Wave 4: IoT long-lived pathways and scheduled-repeat promotion productionized.

Default behavior during rollout:
- direct Soma execution remains primary for low-risk one-off actions
- team manifestation occurs for complexity/risk/repeat/IoT profiles
- remote actuation remains disabled until Wave 3 security gates pass

---

End of document.



