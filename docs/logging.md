# Logging Standard (V8.2)

> Status: Authoritative
> Last Updated: 2026-03-22
> Scope: Required logging and centralized review contract for all Mycelis services, Soma, and team-led execution

## 1. Purpose

Mycelis has to support two different consumers of runtime output:

1. Service operators reading raw process logs.
2. Agents and operators reading centralized, structured execution evidence.

This document governs the second surface. Raw `stdout` or `stderr` logs may stay service-local, but anything that Soma, meta-agentry, team leads, or later review automation should reason over must land in the centralized review model below.

## 2. Centralized Review Surfaces

### 2.1 Mission Events

- Storage: `mission_events`
- Transport: `swarm.mission.events.{run_id}`
- Purpose: persistent audit spine for run-linked execution, mutation, approvals, and lifecycle

Use mission events when the system needs durable causality, replay, audit, or governance review.

### 2.2 Operational Review Logs

- Storage: `log_entries`
- API: `GET /api/v1/memory/stream`
- Stream: SSE log broadcasts from the memory stream
- Purpose: centralized, review-oriented operational feed for Soma, meta-agentry, and team leads

This is the central review surface for:

- team status and result signals
- mirrored team telemetry summaries
- audit-oriented service actions
- bridged legacy runtime envelopes that still need centralized visibility

### 2.3 Team-Native Channels

- `swarm.team.{team_id}.signal.status`
- `swarm.team.{team_id}.signal.result`
- `swarm.team.{team_id}.telemetry`

Teams still own their local output channels. The rule is not "replace team channels." The rule is:

- teams publish on their own canonical subjects first
- centralized review mirrors the relevant output into `log_entries`
- Soma and central services inspect the centralized stream without stealing team ownership of local lanes

## 3. Required Separation of Concerns

### 3.1 Raw service logs

Allowed uses:

- startup diagnostics
- local failure detail
- process-level troubleshooting

These logs are useful for operators, but they are not the canonical agent-review surface.

### 3.2 Agent-reviewable logs

Required uses:

- cross-team review
- Soma orchestration review
- team lead supervision
- central meta-agentry analysis
- replayable operational summaries

These logs must use the centralized schema stored in `log_entries.context`.

## 4. Canonical Review Schema

The required schema is `protocol.OperationalLogContext` stored in `log_entries.context`.

Minimum required fields for centralized review:

- `schema_version`
- `review_scope`
- `audiences`
- `summary`
- `source_channel` when the log came from a bus subject
- `review_channels`
- `centralized_review`

Strongly recommended fields:

- `service`
- `component`
- `detail`
- `why_it_matters`
- `suggested_action`
- `source_kind`
- `payload_kind`
- `run_id`
- `team_id`
- `agent_id`
- `trace_id`
- `mission_event_id`
- `status`
- `tags`

## 5. Review Scope Rules

### 5.1 `service_local`

Use when the record is useful inside one service but not important for central orchestration.

### 5.2 `team_local`

Use when the record is primarily for a team and its lead, but may still be mirrored into the central review feed.

### 5.3 `central_review`

Default for operator-facing team output, cross-team execution updates, and anything Soma or meta-agentry should be able to inspect without polling each team directly.

### 5.4 `audit`

Use for governed or durable records that should align with mission-event review and governance reasoning.

## 6. Audience Rules

Canonical audiences:

- `soma`
- `meta_agentry`
- `team_lead`
- `team`
- `governance`

Default expectations:

- team-local execution output should normally target `team_lead` and also remain visible to `soma` through centralized review
- central orchestration and cross-team coordination should target `soma`, `meta_agentry`, and relevant `team_lead`
- governed mutation and approval evidence should include `governance`

## 7. Template for Agentry Review

Every centralized review log should be understandable when read without surrounding implementation context.

Preferred shape:

- `summary`: what happened in one sentence
- `detail`: compact supporting detail, status text, or payload summary
- `why_it_matters`: why Soma, meta-agentry, or a team lead should care
- `suggested_action`: what to inspect or do next if follow-up is needed
- `review_channels`: exact lanes where more detail can be found

Good example:

```json
{
  "schema_version": "v1",
  "review_scope": "central_review",
  "audiences": ["soma", "meta_agentry", "team_lead"],
  "service": "core",
  "component": "signal-bridge",
  "summary": "Creative Team Lead completed a review.",
  "detail": "Result payload captured from the team result lane.",
  "why_it_matters": "This team output is mirrored into centralized review so Soma and operational leads can inspect progress without leaving the shared channel model.",
  "source_kind": "system",
  "source_channel": "swarm.team.creative.signal.result",
  "payload_kind": "result",
  "run_id": "run-42",
  "team_id": "creative",
  "agent_id": "creative-lead",
  "centralized_review": true,
  "review_channels": [
    "memory.stream",
    "swarm.team.creative.signal.result",
    "swarm.mission.events.run-42"
  ],
  "tags": ["team-output", "result"]
}
```

## 8. Required Runtime Rules

1. Team outputs stay on canonical team subjects.
2. Team outputs intended for operator or orchestrator review must also be visible in centralized review.
3. Telemetry may remain high-volume on team telemetry subjects, but the summarized operational view must still be readable from centralized review when Soma or team leads need it.
4. Mission-critical or mutating execution must emit persistent mission events even when also mirrored into `log_entries`.
5. New centralized review producers must normalize through `protocol.OperationalLogContext`; do not invent ad-hoc JSON blobs in `log_entries.context`.

## 9. Event Taxonomy for Mission Events

Use `core/pkg/protocol/events.go` constants only. Do not invent ad-hoc event names.

Core families:

- Mission: `mission.started`, `mission.completed`, `mission.failed`, `mission.cancelled`
- Team: `team.spawned`, `team.stopped`
- Agent: `agent.started`, `agent.stopped`
- Tool: `tool.invoked`, `tool.completed`, `tool.failed`
- Artifact: `artifact.created`
- Memory: `memory.stored`, `memory.recalled`
- Orchestration: `trigger.fired`, `trigger.skipped`, `scheduler.tick`

## 10. Anti-Patterns

Disallowed:

- treating raw process logs as the primary agent-review surface
- inventing a second centralized logging schema outside `protocol.OperationalLogContext`
- publishing team-local output without any central review visibility when Soma or team leads need to inspect it later
- using telemetry subjects as the only source of operator review truth
- inventing event names outside `core/pkg/protocol/events.go`
- hardcoding `swarm.*` subjects outside canonical topic definitions

## 11. Agent Checklist Before Changing Logging

1. Read this file.
2. Read `core/pkg/protocol/events.go`.
3. Read `core/pkg/protocol/topics.go`.
4. Read `core/pkg/protocol/logging.go`.
5. Confirm how `mission_events` and `log_entries` are both affected.
6. Update focused tests for any new review-path behavior.

If operator-visible logging behavior changes, also update:

1. `README.md`
2. relevant tests
3. any architecture or ops docs that describe the affected surface
