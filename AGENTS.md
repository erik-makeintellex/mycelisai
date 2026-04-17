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
- When invoke task behavior or task names change, update `README.md`, `docs/TESTING.md`, `docs/architecture/OPERATIONS.md`, `ops/README.md`, and any affected in-app docs surface in `interface/lib/docsManifest.ts` in the same slice.

## README Navigation Contract

- Keep a structured `## README TOC` near the top of `README.md`.
- When adding, removing, or renaming major README sections, update the TOC links in the same change.
- Treat the README TOC as the stable navigation contract that future development agents should use before scanning the full file.

## Canonical Docs Location

- Put new canonical planning, target-delivery, UI-target, execution-model, and delivery-governance docs under `docs/architecture-library/`.
- Treat `mycelis-architecture-v7.md` as the stable PRD index and compatibility entrypoint, not the place to grow another giant monolithic spec.
- If a canonical doc is meant to be readable in the in-app `/docs` page, add or update its entry in `interface/lib/docsManifest.ts` in the same change.

## Documentation Synchronization Contract

- Every implementation slice that changes product behavior, runtime behavior, operator workflow, API contract, governance posture, or canonical terminology must include a documentation review in the same slice.
- Update the owning docs in the same change whenever meaning changed, not later as cleanup.
- At minimum review `README.md`, `V8_DEV_STATE.md`, the owning canonical/user/ops docs for the touched surface, and any affected in-app docs entry in `interface/lib/docsManifest.ts`.
- When API behavior or payload meaning changes, review `docs/API_REFERENCE.md` in the same slice.
- When testing or task-running behavior changes, review `docs/TESTING.md`, `docs/architecture/OPERATIONS.md`, and `ops/README.md` in the same slice.
- Slice close-out should explicitly report which docs changed and which touched docs were reviewed but left unchanged.

## Feature Status Standard

- Use these canonical status markers in planning and state docs: `REQUIRED`, `NEXT`, `ACTIVE`, `IN_REVIEW`, `COMPLETE`, `BLOCKED`.
- Preferred meanings:
  - `REQUIRED`: must exist for target delivery or gate pass, but not started/ready yet
  - `NEXT`: highest-priority upcoming implementation slice
  - `ACTIVE`: currently being worked
  - `IN_REVIEW`: implemented and awaiting validation/review/gate decision
  - `COMPLETE`: accepted and delivered
  - `BLOCKED`: cannot advance until a named dependency or defect is resolved
- Avoid inventing synonymous markers like "in progress", "done-ish", or "pending review" when one of the canonical markers fits.

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
