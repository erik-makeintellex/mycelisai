# Repository Standards

This repository is Go-first for product/runtime work and Python-first for management automation.

## Language Ownership

- Go owns core runtime, orchestration, APIs, NATS integrations, and persistence-facing backend logic.
- TypeScript owns the interface, in-app docs browser, and operator-facing workflow surfaces.
- Python owns app management tasks, operator automation, CI task orchestration, and repo-local test harnesses.
- SQL owns schema and migration contracts.
- PowerShell is allowed only as a thin host wrapper when the local platform requires it. App-tied management logic must not live in PowerShell scripts.

## Task Runner Contract

- Use `uv run inv ...` for real task execution.
- Use `uvx --from invoke inv -l` only as a compatibility probe.
- Do not use bare `uvx inv ...`.

## NATS Signal Standard

- Use canonical subject constants from Go protocol/topic definitions for product subjects. Do not hardcode `swarm.*` literals in runtime code.
- Every bus payload that represents product behavior must declare enough metadata to identify source, scope, and intended consumer.

Required metadata for governed product signals:
- `run_id` when the signal is execution-linked
- `team_id` when team-scoped
- `agent_id` when agent-scoped
- `source_kind`
- `source_channel`
- `payload_kind`
- `timestamp`

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

Preferred subject families:
- `swarm.team.{team_id}.internal.command` for directed team input
- `swarm.team.{team_id}.signal.status` for concise operator-readable status
- `swarm.team.{team_id}.signal.result` for bounded execution outcomes
- `swarm.team.{team_id}.telemetry` for high-volume machine telemetry
- `swarm.council.{agent_id}.request` for request-reply specialist calls
- `swarm.mission.events.{run_id}` for run-linked fanout
- `swarm.global.broadcast` for governed fanout

Channel rules:
- Web/API results must normalize to the standard API envelope before UI consumption.
- IoT and sensor payloads must identify device/feed origin and stay separated from operator-facing result channels until normalized.
- High-volume telemetry must not be reused as operator status or workflow-result channels.
- Mutating actions must emit persistent mission events in addition to transient bus signals.

## Infrastructure Development Channel Boundary

- Infrastructure-development or experimentation subjects are local-only and must not be committed as canonical orchestration channels.
- Do not add development-only infrastructure subjects to shared architecture docs, protocol constants, standing manifests, or operator UI flows unless they are intentionally promoted through architecture review.
- If temporary infrastructure-dev subjects are needed for local work, keep them out of the authoritative channel taxonomy and out of persisted workflow orchestration.

## Logging and Error Handling

- Go runtime logs should be structured and component-identified.
- Python task output should be operator-readable, fail fast on broken prerequisites, and avoid false-success messaging.
- UI surfaces should show normalized error states, not raw backend noise.
