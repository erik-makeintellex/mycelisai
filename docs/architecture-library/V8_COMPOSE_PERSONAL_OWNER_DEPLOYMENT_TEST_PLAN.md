# V8 Compose Personal Owner Deployment Test Plan
> Navigation: [Project README](../../README.md) | [Testing](../TESTING.md) | [Local Development Workflow](../LOCAL_DEV_WORKFLOW.md)

Status: ACTIVE Compose proof plan.

## Purpose

Prove the personal-owner Docker Compose deployment lane: single-host runtime, explicit AI endpoint, durable storage, retained outputs, and browser-visible operator workflows.

## Preconditions

- `.env` exists and contains secrets.
- `.env.compose` exists and contains Compose topology.
- `MYCELIS_COMPOSE_OLLAMA_HOST` points to a container-reachable text model endpoint.
- Docker is available either natively or through the configured WSL Docker host.
- Output storage is configured when local-hosted artifacts are in scope.

## Bring-Up

```bash
uv run inv compose.down --volumes
uv run inv compose.up --build --wait-timeout=240
uv run inv compose.status
uv run inv compose.health
```

For data-plane only proof:

```bash
uv run inv compose.infra-up --wait-timeout=180
uv run inv compose.infra-health
uv run inv compose.migrate
uv run inv compose.storage-health
```

## Required Proof

Prove:
- PostgreSQL and NATS are healthy
- Core and Interface are healthy
- text provider is reachable from Core
- storage health covers pgvector and durable Mycelis tables
- browser can reach the delivered UI address
- Soma direct answer works
- governed proposal cancel/execute works
- retained outputs survive refresh

## Browser Evidence

Use focused proof from [V8 UI Team Live Proof](V8_UI_TEAM_LIVE_PROOF.md):

```bash
uv run inv interface.e2e --headed --live-backend --server-mode=start --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts
uv run inv interface.e2e --headed --live-backend --server-mode=start --project=chromium --spec=e2e/specs/groups-live-backend.spec.ts
```

## Failure Handling

Record as blockers:
- unreachable AI endpoint
- storage health failure
- retained output mount failure
- raw backend error in UI
- browser path differs from delivered operator address

## Close-Out

Report runtime lane, UI URL, AI endpoint posture, commands, pass/fail result, screenshots/traces, and docs reviewed.
