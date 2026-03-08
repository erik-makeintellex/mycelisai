# Logging Standard (V7)

> Status: Authoritative
> Last Updated: 2026-03-05
> Scope: Required logging contract for all agents, services, and delivery teams

## 1. Why This Exists

Mycelis currently has two active telemetry surfaces:

1. V7 mission event spine (`mission_events` + run timelines).
2. Memory stream logs (`log_entries` via `/api/v1/memory/stream`).

This document standardizes how work must be logged so incoming agents can reason over logs without guessing format or source.

## 2. Canonical Sources

### 2.1 Persistent Execution Events (Primary for agent reasoning)

- Storage: `mission_events` table.
- Producer path: `core/internal/events/store.go` (`Emit` = DB-first, CTS publish best effort).
- Transport topic: `swarm.mission.events.{run_id}` (`TopicMissionEventsFmt`).
- Primary query surfaces:
  - `GET /api/v1/runs/{id}/events`
  - `GET /api/v1/runs/{id}/chain`

Use this source for execution lineage, causality, and run-level diagnostics.

### 2.2 Memory Stream Logs (Operational stream)

- Storage: `log_entries` table.
- Producer path: `core/internal/memory/service.go` (`Push` -> async `persist`).
- Primary query surface:
  - `GET /api/v1/memory/stream`

Use this source for lightweight operational signal feeds and UI stream rendering.

## 3. Required Event Taxonomy (Mission Events)

Use `core/pkg/protocol/events.go` constants only. Do not invent ad-hoc event names.

Core families:

- Mission: `mission.started`, `mission.completed`, `mission.failed`, `mission.cancelled`
- Team: `team.spawned`, `team.stopped`
- Agent: `agent.started`, `agent.stopped`
- Tool: `tool.invoked`, `tool.completed`, `tool.failed`
- Artifact: `artifact.created`
- Memory: `memory.stored`, `memory.recalled`
- Orchestration: `trigger.fired`, `trigger.skipped`, `scheduler.tick`

## 4. Required Field Contract

### 4.1 Mission event minimums

Every emission must include:

- `run_id` (non-empty)
- `event_type` (from constants)
- `severity` (`info|warn|error`)
- `source_agent` and/or `source_team` where available
- `payload` (map; use empty object when no details)

### 4.2 Payload keys by type

Use stable keys to keep agent parsing deterministic.

- `tool.*`: `tool_name`, `call_id`, `status`, `duration_ms`, `error` (if failed)
- `mission.*`: `mission_id`, `status`, `reason` (if non-success)
- `team.*`: `team_id`, `action`, `lifetime`
- `trigger.*`: `rule_id`, `matched`, `mode`, `decision`
- `scheduler.tick`: `schedule_id`, `due_count`, `dispatch_count`

## 5. Agent Onboarding Checklist (Mandatory Before Coding)

Any agent starting delivery work must do this first:

1. Read this file and `core/pkg/protocol/events.go`.
2. Verify topic constants in `core/pkg/protocol/topics.go`.
3. Confirm event persistence behavior in `core/internal/events/store.go`.
4. Confirm stream log behavior in `core/internal/memory/service.go`.
5. Run current test baseline before changing logging behavior.

If any logging change is made, agent must update:

1. `docs/logging.md`
2. relevant tests in `core/internal/events/*` or `core/internal/memory/*`
3. `README.md` and `V7_DEV_STATE.md` when behavior changes operator-facing outputs

## 6. Anti-Patterns (Disallowed)

- Hardcoded topic strings in handlers/services (must use constants).
- New event names not declared in `protocol/events.go`.
- Missing `run_id` on mission event emissions.
- Emitting CTS-only event signals without DB persistence.
- Unbounded free-form payload blobs for common event classes.

## 7. Near-Term Hardening Tasks

Add/standardize invoke tasks for logging quality gates:

- `uv run inv logging.check-schema`
  - validate event types and required keys across targeted tests/fixtures.
- `uv run inv logging.check-topics`
  - detect hardcoded `swarm.` topic strings outside `protocol/topics.go`.

Current note:
- `logging.check-coverage` is not a live invoke task yet. Coverage for new logging/event paths is enforced through focused tests plus `uv run inv ci.baseline` until a dedicated task is added.

These tasks are required before enabling broader autonomous agent execution.
