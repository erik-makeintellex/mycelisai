# Hardware Interface API and Direct Channels V7

Version: `1.0`
Status: `Authoritative`
Last Updated: `2026-02-28`
Scope: Hardware interface onboarding, hardware-facing APIs, and direct channel communication standards

This document defines how users add hardware interfaces and how Mycelis supports both API-mediated and direct channel communication.

Default posture:
Companion Soma model: `docs/architecture/SOMA_SYMBIOTE_GROWTH_AND_HOST_ACTUATION_V7.md`.
- API-mediated control is preferred
- direct channel support is allowed for common protocols with governance and security controls

---

## Table of Contents

1. Purpose
2. Hardware Integration Modes
3. Hardware Interface API (Control Plane)
4. Direct Channel Support (Data/Actuation Plane)
5. Canonical Hardware Interface Contract
6. Governance and Safety Controls
7. Security Controls for Direct Channels
8. Observability and Diagnostics
9. Testing Matrix
10. Timeline Fit

---

## 1. Purpose

Users must be able to:
1. register a hardware interface in a governed way
2. connect via common direct channels when needed
3. route data/action through Soma direct path or manifested team path
4. trace all hardware interactions to run-linked events

---

## 2. Hardware Integration Modes

### Mode A - Interface API (Preferred)

Hardware is accessed through a managed interface service registered in Mycelis.

Benefits:
- unified schemas
- easier governance controls
- centralized health/diagnostics

### Mode B - Direct Channel (Supported)

Mycelis connects through protocol adapters for low-latency or legacy hardware paths.

Use cases:
- low-level IoT
- industrial controllers
- device fleets with native bus protocols

---

## 3. Hardware Interface API (Control Plane)

### 3.1 Required endpoints (new)

- `GET /api/v1/hardware/interfaces`
  - list interfaces and status
- `POST /api/v1/hardware/interfaces`
  - register interface definition (governed mutation)
- `PATCH /api/v1/hardware/interfaces/{id}`
  - update config, enable/disable, rotate metadata
- `DELETE /api/v1/hardware/interfaces/{id}`
  - remove interface
- `POST /api/v1/hardware/interfaces/{id}/probe`
  - connectivity and capability probe
- `POST /api/v1/hardware/interfaces/{id}/sync`
  - sync actions/capabilities from adapter

### 3.2 Action execution endpoints

- `GET /api/v1/hardware/actions`
- `GET /api/v1/hardware/actions/{action_id}`
- `POST /api/v1/hardware/actions/{action_id}/invoke`

These should map into the universal action gateway contract.

---

## 4. Direct Channel Support (Data/Actuation Plane)

### 4.1 Supported common channel classes

Initial common direct channels:
- `mqtt`
- `websocket`
- `serial`
- `modbus-tcp`
- `opc-ua`
- `canbus`
- `ble` (gateway-mediated where possible)

### 4.2 Channel profile requirements

Each direct channel profile must define:
- protocol and version
- endpoint/port/device address
- read/write capability class
- QoS/timeout/retry profile
- safety limits and allowed command classes

### 4.3 Routing requirement

Direct channels are routed through adapter workers that emit standard mission events and health signals.
No channel may bypass governance and audit rules.

---

## 5. Canonical Hardware Interface Contract

```json
{
  "interface_id": "hw.edge-gateway-01",
  "name": "Edge Gateway 01",
  "mode": "api|direct",
  "protocol": "mqtt|serial|modbus-tcp|opc-ua|canbus|websocket",
  "locality": "local|private_remote|remote",
  "capabilities": ["read_sensor", "set_output", "run_cycle"],
  "risk_level": "low|medium|high",
  "enabled": true,
  "heartbeat_sec": 15,
  "action_schema_ref": "..."
}
```

Invoke context must include:
- `run_id`
- `team_id`
- `agent_id`
- `origin`
- `idempotency_key`
- optional `approval_ref` for high-risk operations

---

## 6. Governance and Safety Controls

1. Hardware interface registration is a governed mutation.
2. High-risk hardware actions require proposal + approval reference.
3. Allowed command classes are explicit allowlists per interface.
4. Max parameter bounds are enforced via schema + policy.
5. Emergency halt path must exist for each actuation interface.

---

## 7. Security Controls for Direct Channels

1. Private-network posture by default for direct hardware channels.
2. Mutual authentication for remote/private mesh links.
3. Replay protections (`nonce`, TTL, idempotency).
4. Segment control-plane from telemetry-plane channels.
5. Strict credential handling and rotation for hardware endpoints.

---

## 8. Observability and Diagnostics

Each hardware interaction must emit:
- `tool.invoked`
- `tool.completed` or `tool.failed`
- hardware-specific event context (interface/protocol/channel)

Additional diagnostics:
- channel health (`connected|degraded|offline|error`)
- last heartbeat timestamp
- last command latency
- last fault code

---

## 9. Testing Matrix

### Unit
- interface config validation
- protocol adapter validation
- command-boundary policy checks

### Integration
- register/probe/sync interface lifecycle
- direct channel connect/disconnect/recovery
- command idempotency/replay protections

### API
- interface CRUD + invoke contract tests
- unauthorized and unsafe command rejection tests

### E2E
- one-off direct Soma hardware action
- manifested team hardware workflow
- scheduled repeat hardware routine
- degraded channel recovery with no context loss

---

## 10. Timeline Fit

This track fits in the existing wave model:
- Wave 2: hardware interface API scaffolds and schema contracts
- Wave 3: secure direct channel controls and private mesh validation
- Wave 4: productionized low-level IoT persistent-team pathways and scheduled-repeat support

Remote public actuation remains disabled until Wave 3 security gates pass.

---

End of document.
