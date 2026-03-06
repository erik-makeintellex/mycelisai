# NATS Signal Standard V7

Version: `1.0`
Status: `Authoritative`
Last Updated: `2026-03-06`
Scope: NATS subject families, source normalization, signal metadata, and product-vs-dev channel boundaries

This document defines the canonical signal standard for Mycelis NATS traffic across workspace UI, web APIs, automation triggers, sensors, IoT paths, internal tools, and system/runtime services.
If a team manifest, runtime feature, UI surface, or test fixture conflicts with this standard, this standard wins.

Companion topology contract: `docs/architecture/SOMA_TEAM_CHANNEL_ARCHITECTURE_V7.md`.
Companion execution contract: `docs/architecture/UNIVERSAL_ACTION_INTERFACE_V7.md`.
Companion operations/testing contract: `docs/architecture/OPERATIONS.md` and `docs/TESTING.md`.

---

## 1. Goals

Required outcomes:
1. Every meaningful product signal identifies its source class and intended audience.
2. Team input, operator status, machine telemetry, and execution results do not share ambiguous subjects.
3. Web/API responses and IoT/sensor streams are normalized before they meet operator-facing UI flows.
4. Development-only infrastructure experimentation stays outside canonical orchestration.

---

## 2. Signal Classes

| Class | Purpose | Consumer | Persistence |
| :--- | :--- | :--- | :--- |
| `internal.command` | Directed team control input | team runtime | transient |
| `signal.status` | Compact operator-readable state updates | UI, operators, coordinators | transient, may mirror mission events |
| `signal.result` | Bounded execution outcomes and summaries | UI, automation callers | transient, may mirror mission events |
| `telemetry` | High-volume machine metrics or trace detail | observability and recovery tooling | transient |
| `request` / `reply` | Request-reply specialist or tool execution flow | targeted caller | transient |
| `mission.events` | Run-linked audit fanout | timeline, replay, recovery | backed by persistent mission events |

Rules:
- `signal.status` is not a telemetry dump.
- `signal.result` is not a substitute for persistence.
- `telemetry` is not an operator-facing result channel.
- Request-reply subjects are not shared event streams.
- In runtime, operator-facing `signal.status` and `signal.result` subjects are wrapped with standardized metadata before publish.

---

## 3. Required Metadata

Every governed product signal must carry:
- `timestamp`
- `source_kind`
- `source_channel`
- `payload_kind`

Execution-linked signals must also carry:
- `run_id`
- `team_id` when team-scoped
- `agent_id` when agent-scoped

Recommended metadata:
- `mission_event_id` when a persistent event already exists
- `tenant_id`
- `governance_mode`
- `origin`
- `device_id` for IoT/device-originated payloads

Canonical `source_kind` values:
- `workspace_ui`
- `web_api`
- `automation_trigger`
- `scheduler`
- `sensor`
- `iot`
- `internal_tool`
- `mcp`
- `system`

Canonical `payload_kind` values:
- `command`
- `status`
- `result`
- `telemetry`
- `event`
- `artifact`
- `error`

---

## 4. Canonical Subject Families

| Subject | Contract |
| :--- | :--- |
| `swarm.team.{team_id}.internal.command` | Team-directed input and control messages |
| `swarm.team.{team_id}.signal.status` | Concise team health and phase-state updates |
| `swarm.team.{team_id}.signal.result` | Bounded team outputs intended for operators or orchestrators |
| `swarm.team.{team_id}.telemetry` | High-volume machine telemetry and trace detail |
| `swarm.council.{agent_id}.request` | Request-reply specialist calls |
| `swarm.mission.events.{run_id}` | Run-linked fanout keyed to persistent mission events |
| `swarm.global.broadcast` | Governed global broadcast |

Subject rules:
- Use protocol constants for product subjects in Go code.
- Team command input should land on `.internal.command`, not generic wildcard subjects.
- Team status and result subjects must stay distinct even when they share the same producer.
- New product subject families require architecture review before they become authoritative.
- Current runtime preserves legacy internal worker subjects (`.internal.trigger`, `.internal.response`) behind the team boundary, but product-facing orchestration should target `.internal.command`, `.signal.status`, and `.signal.result`.

---

## 5. Source Normalization Rules

### 5.1 Workspace UI and Web API

- Operator-facing API responses must normalize to the standard API envelope before UI state consumes them.
- Bus signals derived from web/API work must declare `source_kind=workspace_ui` or `source_kind=web_api`.
- UI should consume normalized status/result channels, not raw telemetry subjects.
- Frontend stream consumers should normalize legacy signal objects and standardized signal envelopes through one shared boundary helper before rendering operator surfaces.

### 5.2 Automation, Scheduler, and Internal Tooling

- Trigger and scheduler emissions must identify `source_kind=automation_trigger` or `source_kind=scheduler`.
- Internal tool and MCP outputs must declare `source_kind=internal_tool` or `source_kind=mcp`.
- Mutating actions must emit persistent mission events in addition to transient NATS signals.

### 5.3 Sensor and IoT Paths

- Device, feed, or adapter-originated messages must declare `source_kind=sensor` or `source_kind=iot`.
- Sensor and IoT payloads must carry origin identity such as `device_id`, `sensor_id`, or adapter metadata.
- Raw sensor/IoT streams must not be exposed directly as operator result channels until normalized.
- Long-lived monitoring/control loops may stay on telemetry-class subjects, but operator-facing summaries belong on `signal.status` or `signal.result`.

---

## 6. Logging and Error Handling

- Signal-handling code must log the component, subject, and source kind when rejecting or degrading a message.
- Invalid or incomplete governed payloads must fail fast with actionable errors; do not emit false-success status.
- Operator-facing status messages should summarize degraded state without dumping raw bridge/runtime noise.
- Persistent `mission_events` remain the audit source of truth for mutating or materially significant work.

---

## 7. Development-Only Channel Boundary

Infrastructure-development and experimentation channels are not part of the canonical orchestration surface.

Rules:
- Do not add infrastructure-dev subjects to `pkg/protocol/topics.go` unless they have passed architecture review.
- Do not publish infrastructure-dev subject families in standing team manifests or operator UI wiring as if they were product channels.
- Keep temporary infrastructure-dev channels local to development tasks, fixtures, or isolated experiments.
- Promotion from dev-only to product channel requires documentation updates, protocol constants, tests, and explicit acceptance evidence.

---

## 8. Testing and Gate Expectations

Minimum coverage for any new product channel:
1. Subject naming and constant coverage in runtime code.
2. Envelope or metadata validation for source classification.
3. Logging/error-path coverage for malformed payloads.
4. Documentation and docs-manifest updates for authoritative new contracts.
5. `uv run inv ci.entrypoint-check` plus targeted task/runtime tests when management surfaces change.

Recommended verification commands:
- `uv run inv ci.entrypoint-check`
- `uv run inv logging.check-topics`
- `$env:PYTHONPATH='.'; uv run pytest tests/test_ci_tasks.py tests/test_logging_tasks.py -q`

---

## 9. Promotion Rule

No new channel variant becomes canonical until:
- the subject family is documented here or in the companion channel architecture doc
- the runtime uses constants instead of literals
- operator-facing consumers know whether the payload is command, status, result, or telemetry
- the change is reflected in `README.md`, `V7_DEV_STATE.md`, and the in-app docs manifest
