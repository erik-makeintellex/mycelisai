# Soma Team and Channel Architecture V7
> Navigation: [Project README](../../README.md) | [Backend](BACKEND.md) | [NATS Signal Standard](NATS_SIGNAL_STANDARD_V7.md) | [V8 Runtime Contracts](../architecture-library/V8_RUNTIME_CONTRACTS.md)

Status: historical V7 migration input.

Use this file only for the V7 rationale behind team/channel behavior. Active subject taxonomy and runtime requirements live in Go protocol/topic constants, [Backend](BACKEND.md), [NATS Signal Standard](NATS_SIGNAL_STANDARD_V7.md), and V8 runtime docs.

## Retained V7 Signal

The useful V7 constraints are:
- separate operator-facing status/result signals from high-volume telemetry
- keep team, agent, run, source, channel, payload kind, and timestamp metadata explicit
- preserve request-reply boundaries for specialist/council calls
- keep device/sensor ingress separate until normalized
- make team process state inspectable to the operator
- persist mission events for mutating actions

## Current Subject Families

Preferred product subject families remain:
- `swarm.team.{team_id}.internal.command`
- `swarm.team.{team_id}.signal.status`
- `swarm.team.{team_id}.signal.result`
- `swarm.team.{team_id}.telemetry`
- `swarm.council.{agent_id}.request`
- `swarm.mission.events.{run_id}`
- `swarm.global.broadcast`

Runtime code should use canonical constants, not hardcoded literals.

## Channel Rules

- Web/API results normalize to the standard API envelope before UI consumption.
- IoT and sensor payloads identify device/feed origin.
- Telemetry is not reused as operator status.
- Mutating actions emit persistent mission events in addition to transient bus signals.
- Development-only subjects stay local and do not enter shared docs or UI flows.

## Migration Rule

When promoting V7 channel behavior, update:
- Go protocol/topic constants
- backend tests
- [Backend](BACKEND.md)
- [Testing](../TESTING.md)
- `.state/V8_DEV_STATE.md`

## Proof Rule

Any promoted channel must have metadata coverage and a test proving the UI receives normalized status/result data rather than raw bus payloads.
