# V8 Compose Personal Owner Deployment Test Plan
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)

> Status: Canonical
> Last Updated: 2026-04-12
> Purpose: Define the near-enterprise validation plan for a personal owner running Mycelis locally with Docker Compose, starting PostgreSQL and NATS as an explicit data plane before launching Core and Interface against the configured services.

## TOC

- [Product Intent](#product-intent)
- [Deployment Shape](#deployment-shape)
- [Owner Configuration Contract](#owner-configuration-contract)
- [Team Execution Model For The Test](#team-execution-model-for-the-test)
- [Stage 0: Preconditions](#stage-0-preconditions)
- [Stage 1: Data Plane Only](#stage-1-data-plane-only)
- [Stage 2: Schema And Persistence](#stage-2-schema-and-persistence)
- [Long-Term Postgres Storage Contract](#long-term-postgres-storage-contract)
- [Stage 3: App Services Against Existing Data Plane](#stage-3-app-services-against-existing-data-plane)
- [Stage 4: Personal Owner First-Run Workflow](#stage-4-personal-owner-first-run-workflow)
- [Stage 5: Near-Enterprise Workflow Proof](#stage-5-near-enterprise-workflow-proof)
- [Stage 6: Security And Owner-Control Checks](#stage-6-security-and-owner-control-checks)
- [Stage 7: Observability And Recovery](#stage-7-observability-and-recovery)
- [Stage 8: Cleanup And Reset](#stage-8-cleanup-and-reset)
- [Automation Matrix](#automation-matrix)
- [Risks And Follow-Ups](#risks-and-follow-ups)

## Product Intent

The Docker Compose home-runtime path must support a personal owner who wants a local deployment with explicit control over where data lives and how Mycelis logs into that data plane.

This is not meant to replace Kubernetes or hosted enterprise deployment. It is the supported single-owner local path that should still be tested with enterprise discipline:
- explicit data-service ownership
- repeatable launch and teardown
- persisted PostgreSQL and NATS volumes
- visible Core and Interface health
- browser-level operator workflow validation
- owner-controlled secrets and output storage
- clear recovery steps when the data plane, app plane, or model provider is unavailable

## Deployment Shape

The supported baseline is one Compose project named `mycelis-home`.

Data plane:
- `postgres`: `pgvector/pgvector:pg16`
- `nats`: `nats:2.11-alpine` with JetStream and monitor endpoint
- named volumes: `postgres-data`, `nats-data`
- host ports from `.env.compose`: `MYCELIS_COMPOSE_POSTGRES_PORT`, `MYCELIS_COMPOSE_NATS_PORT`, `MYCELIS_COMPOSE_NATS_MONITOR_PORT`

App plane:
- `core`: Go runtime and API
- `interface`: Next.js UI
- output block mounted from `MYCELIS_OUTPUT_HOST_PATH`
- provider endpoint configured by `MYCELIS_COMPOSE_OLLAMA_HOST` or the broader provider/env surfaces

The first supported test path is:

```powershell
uv run inv compose.infra-up --wait-timeout=180
uv run inv compose.infra-health
uv run inv compose.migrate
uv run inv compose.storage-health
uv run inv compose.up --build --wait-timeout=240
uv run inv compose.health
```

`compose.infra-up` starts only PostgreSQL and NATS. It does not start Core or Interface.

## Owner Configuration Contract

The owner must configure `.env.compose` before launch.

Required owner-owned values:
- `MYCELIS_API_KEY`
- `MYCELIS_BREAK_GLASS_API_KEY`
- `POSTGRES_USER`
- `POSTGRES_PASSWORD`
- `POSTGRES_DB`
- `DB_USER`
- `DB_PASSWORD`
- `DB_NAME`
- `NATS_URL`
- `MYCELIS_OUTPUT_BLOCK_MODE`
- `MYCELIS_OUTPUT_HOST_PATH`

Same Compose project app containers should use:

```env
DB_HOST=postgres
DB_PORT=5432
NATS_URL=nats://nats:4222
```

Host-native clients should use:

```env
DB_HOST=127.0.0.1
DB_PORT=${MYCELIS_COMPOSE_POSTGRES_PORT}
NATS_URL=nats://127.0.0.1:${MYCELIS_COMPOSE_NATS_PORT}
```

Separate Compose app projects that connect through host-published ports should use:

```env
DB_HOST=host.docker.internal
DB_PORT=${MYCELIS_COMPOSE_POSTGRES_PORT}
NATS_URL=nats://host.docker.internal:${MYCELIS_COMPOSE_NATS_PORT}
```

Linux Docker Engine may require a host-gateway mapping or a shared external Docker network for the separate-app-project mode. The single `mycelis-home` Compose project remains the first supported path.

The automation must not print DB passwords. It may print the selected DB user, DB name, host/port, and NATS URL shape so the owner can configure clients without exposing secrets in logs.

## Team Execution Model For The Test

Use a compact test team:

1. Team Lead: owns the test run state, step-by-step checklist, and operator-facing findings.
2. Architect Prime: validates that the Compose deployment shape matches the product intent and data-plane/app-plane boundary.
3. Focused Developer / Tester: runs the task automation, browser checks, and persistence assertions.

Add a fourth reviewer only when release signoff needs separate security, docs, or platform review. Do not create a large standing test pool for this Compose workflow.

## Stage 0: Preconditions

Acceptance criteria:
- Docker Engine is running.
- `.env.compose` exists and is copied from `.env.compose.example`.
- owner has replaced example secrets before release-like testing.
- `MYCELIS_OUTPUT_HOST_PATH` exists when `MYCELIS_OUTPUT_BLOCK_MODE=local_hosted`.
- expected ports are free or intentionally remapped.
- no stale Core/Interface containers are running before data-plane-only testing.

Commands:

```powershell
uv run inv compose.status
```

If the stack is dirty and a fresh test is required:

```powershell
uv run inv compose.down
```

Use `uv run inv compose.down --volumes` only for an intentional destructive reset.

## Stage 1: Data Plane Only

Goal: start only PostgreSQL and NATS and expose their host ports for a later app launch.

Commands:

```powershell
uv run inv compose.infra-up --wait-timeout=180
uv run inv compose.infra-health
uv run inv compose.status
```

Acceptance criteria:
- `postgres` and `nats` are running.
- `core` and `interface` remain stopped.
- Postgres host port is reachable.
- Postgres accepts a query with configured `DB_USER` and `DB_NAME`.
- NATS host port is reachable.
- NATS monitor responds at `/varz`.
- task output shows connection guidance without printing DB password.

Failure recovery:
- `uv run inv compose.logs postgres`
- `uv run inv compose.logs nats`
- verify `MYCELIS_COMPOSE_POSTGRES_PORT`, `MYCELIS_COMPOSE_NATS_PORT`, and `MYCELIS_COMPOSE_NATS_MONITOR_PORT`

## Stage 2: Schema And Persistence

Goal: prove the data plane is usable before app launch.

Commands:

```powershell
uv run inv compose.migrate
uv run inv compose.migrate
uv run inv compose.storage-health
```

Acceptance criteria:
- canonical migrations apply against the Compose Postgres service.
- repeated migration run skips replay when schema compatibility checks pass.
- pgvector-backed tables remain available.
- `compose.storage-health` passes the long-term storage contract below.
- data survives `compose.down` without `--volumes`.
- data resets only when `compose.down --volumes` is intentionally used.

Suggested live proof:
- run migrations
- insert/read/delete a harmless sentinel using an existing test-safe table or temporary verification table
- stop/restart data plane without volumes
- confirm schema and sentinel behavior remain correct

## Long-Term Postgres Storage Contract

Mycelis does not use PostgreSQL as a generic connectivity dependency. It is the durable local substrate for long-horizon product memory and retained work.

`uv run inv compose.storage-health` must pass after migrations before the personal-owner path claims the following capabilities are available:

- pgvector extension: semantic vector recall foundation.
- `context_vectors`: durable RAG, deployment context, customer context, company knowledge, and reflection/context chunks.
- `agent_memories`: reviewed Soma/agent memory.
- `conversation_turns`: session continuity and replayable conversation evidence.
- `artifacts`: retained outputs, documents, files, media references, and review states.
- `temp_memory_channels`: restart-safe short-horizon working continuity.
- `collaboration_groups`: temporary and standing group persistence.
- `exchange_channels` and `exchange_items`: managed exchange retention for team/council communication and normalized tool/output records.
- `conversation_templates`: reusable Soma/Council/team asks and output contracts.

Acceptance criteria:
- storage-health checks use the configured `.env.compose` database user and database name.
- storage-health does not require Core or Interface to be running.
- missing storage tables produce migration guidance, not a generic database failure.
- `compose.migrate` may skip unsafe full forward replay when the base schema is already compatible, but it must still apply known missing late storage migrations such as `038_conversation_templates.up.sql` before storage-health is considered green.
- the operator can identify which persistent layer failed before launching the product UI.

## Stage 3: App Services Against Existing Data Plane

Goal: start Core and Interface after the data plane is already up.

Commands:

```powershell
uv run inv compose.up --build --wait-timeout=240
uv run inv compose.health
```

Acceptance criteria:
- Core connects to the existing Postgres and NATS services.
- Interface connects to Core.
- `compose.health` passes Core, Template Engine, Brains API, Telemetry, Frontend, NATS monitor, and text cognitive engine checks.
- if the text engine is unavailable, health fails with an actionable provider/setup blocker rather than a false-success result.

Note: `compose.up` may re-run the infra start step idempotently. That is acceptable for the single `mycelis-home` project. A future `compose.app-up` task can narrow this if separate app-plane testing becomes a release requirement.

## Stage 4: Personal Owner First-Run Workflow

Goal: validate the product as a real owner workflow, not just port availability.

Manual browser flow:
1. Open `http://localhost:3000`.
2. Enter the Soma workspace.
3. Ask: "What is the current state of this Mycelis deployment?"
4. Ask: "What teams exist and what are they doing?"
5. Open Settings and review local/admin identity posture.
6. Open Resources and review Connected Tools and Deployment Context.
7. Open Teams and start guided team creation.
8. Create a compact test team.
9. Confirm the team starts from Team Lead, Architect Prime, and focused builder unless a fourth or fifth role is justified.
10. Return to Soma and ask Soma to summarize the team output/state.

Acceptance criteria:
- Soma answers state questions without generic server-connectivity apologies.
- Settings and Resources load without raw backend noise.
- team creation stays compact and operator-readable.
- retained outputs remain visible through chat, team, group, or artifact surfaces.

## Stage 5: Near-Enterprise Workflow Proof

Workflow set:
- direct Soma answer: ordinary prose answer returns inline without unnecessary delegation.
- team-managed output: compact team produces a retained product-demo checklist or similar business artifact.
- governed mutation: proposal/confirm/cancel flow works and is auditable.
- deployment context: owner uploads or pastes a small private/company context item and Soma can use it within scope.
- MCP visibility: Connected Tools shows empty-state/library guidance or installed tools accurately.
- output block: renderable outputs preview inline; binary/download-only outputs show a clickable path or artifact reference.
- persistence: conversation continuity, organization/team state, context item, and retained output survive app restart.

Acceptance criteria:
- every workflow has a visible user outcome.
- every generated or retained output has a review path.
- no hidden large agent pool appears during team creation.
- data-plane state remains owned by the configured Compose services.

## Stage 6: Security And Owner-Control Checks

Scope: personal-owner confidence, not a full enterprise security audit.

Acceptance criteria:
- `.env.compose` stays untracked.
- example secrets are clearly examples and not release guidance.
- task output does not print DB password or API keys.
- NATS monitor host exposure is documented as local-only unless the owner intentionally exposes it.
- Postgres host-port exposure is understood as local-owner convenience, not hardened multi-tenant posture.
- output block path validation prevents accidental file mount surprises.
- MCP default bootstrap remains disabled unless owner explicitly configures it.
- CORS remains scoped to expected Interface origin.

## Stage 7: Observability And Recovery

Commands:

```powershell
uv run inv compose.status
uv run inv compose.infra-health
uv run inv compose.health
uv run inv compose.logs postgres
uv run inv compose.logs nats
uv run inv compose.logs core
uv run inv compose.logs interface
```

Acceptance criteria:
- infra-only health does not require Core or Interface.
- full health requires Core, Interface, and text cognitive engine.
- errors point to next commands and likely broken configuration.
- UI shows normalized errors rather than raw stack traces.

## Stage 8: Cleanup And Reset

Normal stop:

```powershell
uv run inv compose.down
```

Destructive fresh rebuild:

```powershell
uv run inv compose.down --volumes
```

Acceptance criteria:
- normal stop leaves named data volumes intact.
- destructive reset removes named data volumes only when explicitly requested.
- no unexpected Core/Interface containers remain after infra-only testing.
- ports `3000`, `8081`, `5432`, `4222`, and `8222` match the expected phase state.

## Automation Matrix

Automated now:
- `compose.infra-up` launches only `postgres` and `nats`.
- `compose.infra-health` validates Postgres port, Postgres query, NATS port, and NATS monitor.
- `compose.storage-health` validates long-term Postgres storage after migrations.
- compose task tests cover infra-only command ordering, migration opt-in, connection guidance, no password printing, infra-only health behavior, and long-term storage check behavior.
- docs-link tests must pass after this plan is indexed.

Live Docker gate:
- data-plane-only launch
- infra health
- migration idempotence
- long-term storage health
- full app launch against existing data plane
- full compose health

Manual browser proof:
- owner first-run workflow
- Soma state answer
- compact team creation
- deployment context use
- MCP visibility
- governance proposal
- retained output review

## Risks And Follow-Ups

- NATS auth is not currently part of the local Compose default.
- separate app Compose project against shared data services needs explicit host-endpoint or external-network testing before it becomes a first-class release promise.
- media engine remains optional/environment-specific; Compose can validate text/product workflow while recording media as a configured-provider follow-up.
- enterprise identity/SSO is outside the personal-owner Compose target and should not block this local deployment path.
- a future `compose.app-up` task may be useful if app-plane-only start becomes common in testing.
