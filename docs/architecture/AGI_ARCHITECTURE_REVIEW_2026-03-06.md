# AGI Architecture Review - 2026-03-06

## Purpose

This document is the current architecture review addendum for the active system.
It records runtime truths that were validated during the latest delivery and
operations hardening work and should be treated as a corrective overlay for the
existing architecture set until those base documents are fully folded forward.

Review mode for this pass used standard communication fallback. The live NATS
team bus is available, but this review record is written directly into the repo
so the architecture state remains durable and auditable.

## Reviewed Scope

- `docs/architecture/OVERVIEW.md`
- `docs/architecture/OPERATIONS.md`
- `docs/architecture/NEXT_TARGET_GATED_DELIVERY_PROGRAM.md`
- `docs/architecture/SOMA_TEAM_CHANNEL_ARCHITECTURE_V7.md`
- `V7_DEV_STATE.md`

## Current Runtime Truths

### 1. Operator entrypoint contract

The supported local operator entrypoints are:

- `uv run inv ...`
- `.\.venv\Scripts\inv.exe ...`

`uvx inv ...` is not the standard operator contract for this repo and must not
be treated as the documented default path for lifecycle and database workflows.

### 2. Memory restart sequencing

`lifecycle.memory-restart` must not call `db.reset` after tearing down local
port-forwards without first restoring PostgreSQL connectivity.

Required sequence:

1. Stop stack components.
2. Re-establish the PostgreSQL bridge.
3. Run `db.reset`.
4. Apply migrations.
5. Bring the remaining services back up.

This is a correctness requirement, not an optimization. Running reset against a
dead `127.0.0.1:5432` bridge yields connection refused failures and misleading
operator output.

### 3. Database reset behavior

Database reset flows must fail fast when PostgreSQL is unavailable. They must
not print successful recreation or readiness messages if `psql` connectivity
has not been established.

### 4. Windows shutdown behavior

Windows shutdown paths may surface `taskkill` timeout behavior during lifecycle
teardown. This must be handled as an operational edge case and must not cause
teardown to abort if the target process or port has already been cleared.

### 5. NATS bridge behavior

Transient `kubectl port-forward` errors on `:4222` can appear during startup,
including `portforward.go:404` with a local socket reset. A single event of
this type is not by itself evidence of a NATS or JetStream outage.

Treat it as actionable only if one of the following is true:

- the error repeats continuously
- runtime health reports NATS offline
- messaging features demonstrably fail

### 6. Team topology

The current runtime team set includes:

- `prime-architect`
- `prime-development`
- `agui-design-architect`

These teams are live on the NATS bus, but they are runtime-instantiated rather
than manifest-backed. They will not survive a Core restart unless persisted as
standing manifests.

### 7. Team communication model

The team transport is NATS-first when available. Team input subjects are for
one-way triggering. Observability and collaboration flow through status and
signal subjects rather than request-reply semantics.

Architecture and operational docs must avoid implying that a command published
to a team input subject will behave like a synchronous RPC.

### 8. Gated delivery policy

The active next-target workflow is gated and phase-ordered:

- `P0`
- `P1`
- `P2`
- `P3`
- `P4`

No downstream phase may start before the current phase gate passes. This gate
discipline is now part of the delivery architecture and should be reflected in
planning and execution docs.

### 9. Management scripting standard

Cross-platform app management workflows should be implemented in Python, not
PowerShell.

PowerShell is acceptable only as a thin local usage layer when a Windows host
needs an invocation wrapper. It must not become the source of truth for
management logic that is tied to the application lifecycle, environment setup,
or delivery controls.

## Architecture Corrections Required In Base Docs

### `docs/architecture/OVERVIEW.md`

Must explicitly state:

- the repo-standard operator entrypoints
- that live teams can be runtime-only unless backed by manifests
- that phased workflow delivery is currently gate-enforced

### `docs/architecture/OPERATIONS.md`

Must explicitly state:

- the required PostgreSQL bridge restoration before `db.reset`
- fail-fast expectations for database reset flows
- Windows teardown timeout handling expectations
- interpretation guidance for transient NATS port-forward errors
- Python-first management scripting for cross-platform operator workflows

### `docs/architecture/SOMA_TEAM_CHANNEL_ARCHITECTURE_V7.md`

Must explicitly state:

- runtime teams versus manifest-backed teams
- one-way trigger semantics for team command subjects
- the need for status-topic observability in multi-agent coordination

### `docs/architecture/NEXT_TARGET_GATED_DELIVERY_PROGRAM.md`

Must remain the controlling record for:

- strict `P0 -> P1 -> P2 -> P3 -> P4` execution order
- gate-pass enforcement
- evidence requirements before phase advancement

### `V7_DEV_STATE.md`

Must track the still-open operational risk:

- Core startup timing remains the main flaky point during full
  `lifecycle.memory-restart` execution after the database and bridge defects
  were corrected

## Accepted Review Outcome

The architecture is directionally coherent, but the docs needed correction in
three places:

- runtime operations behavior was more accurate in code than in docs
- team lifecycle and transport semantics were not durable in the docs
- gated delivery policy needed to be treated as architecture, not just process

This review closes those gaps at the addendum level and should be folded into
the base documents as the next documentation pass.

## Required Follow-On Work

1. Persist the live team set as manifest-backed teams if they are intended to
   survive restarts.
2. Fold this addendum into the base architecture docs listed above.
3. Continue P0 until the full memory restart path passes without Core startup
   timing regression.
