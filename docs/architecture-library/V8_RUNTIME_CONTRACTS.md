# V8 Runtime Contracts
> Navigation: [Project README](../../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)

Status: canonical V8 runtime contract.

This document owns the active runtime contract set. Earlier split runtime, V8.1 baseline detail, and V8.2 planning docs have been removed from the active library; preserve still-needed requirements here, in [V8.3 Operational Embodiment PRD](V8_3_OPERATIONAL_EMBODIMENT_PRD.md), or in [V8 Config and Bootstrap Model](V8_CONFIG_AND_BOOTSTRAP_MODEL.md).

## Purpose

V8 runtime contracts define what the system must support after templates/configuration become an instantiated organization. Bootstrap explains how inputs enter the system; runtime contracts explain the live concepts that Core, UI, policy, memory, and tests must preserve.

## Inception

An Inception is the instantiated AI Organization runtime object. It is not a template, not a provider session, and not a disposable chat thread.

Required behavior:
- owns organization identity and purpose
- anchors Soma, council/advisors, teams/departments, specialists, policy, provider posture, memory, and continuity
- receives bootstrap-resolved defaults
- persists as runtime truth

Boundaries:
- templates remain reusable blueprints
- execution results are runtime state, not inherited config
- V7 YAML assets must be translated through the bootstrap model

## Soma Kernel

The Soma Kernel is the runtime contract for Soma as the primary operator-facing coordinator.

Responsibilities:
- receive operator intent
- answer directly when safe
- route to council/team/specialist paths when needed
- propose protected or mutating actions
- preserve continuity and response contract expectations
- surface blockers in user-readable form

Boundaries:
- Soma is not a raw model-picker shell
- Soma does not bypass policy
- Soma does not turn every prompt into a standing team

## Central Council

Central Council is the advisory/governance support layer beneath Soma. It may contain advisor roles, specialist request-reply paths, and policy-aware reasoning helpers.

Responsibilities:
- provide bounded advice
- support routing, critique, planning, and review
- stay subordinate to Soma in the default UX
- use normalized request-reply contracts

Boundaries:
- not the default front door
- not a flat multi-bot messenger
- not a way to bypass provider or capability policy

## Provider Policy

Provider Policy defines which model/media providers and profiles may serve each runtime scope.

Rules:
- environment overrides configure deployment endpoints and profile defaults
- instantiated organization truth controls runtime behavior
- lower scopes may specialize only where higher policy allows
- secret values stay in `.env` or secret stores
- unavailable providers produce blocker/degraded states

## Identity and Continuity State

Identity and continuity state preserve organization identity, Soma posture, reviewed memory, retained outputs, response style, and continuity defaults.

Learning Loops, semantic continuity, Procedure / Skill Sets define the current learning-layer vocabulary. semantic continuity recall uses reviewed memory promotion inputs and procedure and skill retrieval for type-bound specialization.

The pgvector-backed semantic continuity substrate provides event, action, and result semantic indexing. Soma Kernel interprets and orchestrates continuity using semantic recall, but it does not become the memory substrate itself. Loops generate candidates, perform review, and route promotion; they never silently rewrite continuity state.

Memory states are raw memory, reviewed memory, and promoted memory.

Rules:
- identity does not inherit as generic config
- transient conversation state is not configuration
- memory promotion must be reviewable
- continuity must survive refresh/restart when the feature claims persistence

## Runtime Relationship

```text
Template/config sources
  -> Bootstrap resolution
  -> Inception
  -> Soma Kernel
  -> Council / teams / specialists
  -> governed execution, memory, activity, and retained outputs
```

## Team Work Signals

Durable team work is an operator-visible runtime object, not only a bus message.

Rules:
- bounded team asks create a `TeamWorkItem` before any worker response exists
- async asks publish governed command envelopes with `work_item_id`, `team_id`, expected outputs, expected proof, and source context
- team command/result/status paths must preserve `work_item_id` so `swarm.team.{team_id}.signal.status` and `swarm.team.{team_id}.signal.result` can update the original Active Work row
- correlated result payloads should include retained `outputs[]` or normalized `output_refs[]`; the runtime projects real workspace paths, app URLs, proof refs, and audit refs onto the original work item instead of fabricating storage locations
- status-only teams may still advance a queued item to `output_ready` when their correlated status payload declares `state=output_ready`
- explicit degraded states in result/status payloads must remain degraded and carry the degradation reason into recovery posture
- uncorrelated team signals are ignored by Active Work projection instead of mutating the wrong work item

## Migration Note

Historical content remains migration evidence only. Promote active requirements into this document, bootstrap docs, UI/API docs, and `.state/V8_DEV_STATE.md`.
