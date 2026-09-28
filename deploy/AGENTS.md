# Deploy (Compose)

> Navigation: [Repo AGENTS.md](../AGENTS.md) | [Charts AGENTS.md](../charts/AGENTS.md)

## Owns / does not own
- `deploy/compose/**` (Compose overlays) plus root `docker-compose.yml` -> `mycelis-platform-ops` (`.claude/agents/`, local-only and gitignored).
- See [`charts/AGENTS.md`](../charts/AGENTS.md) for the Helm/Kubernetes counterpart; a topology change here (ports, service names, env forwarding) is usually mirrored there.

## Contracts
- `.env` is the local secret store; `.env.compose` is topology and non-secret runtime shape only, and secret-like values from `.env` are authoritative over stale `.env.compose` values.
- The retained `mycelis-home-*` Compose stack and `vllm-node` are shared and already up; no slice agent starts, stops, or recreates them, and only one disposable `docker run --rm` Postgres is allowed for DB tests.
- The operator-approved Compose-only authenticated-Core control-peer exception is canonical (`.state/V8_DEV_STATE.md` "B2 Private Deployment Delivery Evidence"); do not extend that exception to Kubernetes without a new architecture decision.

## Gates
- Focused Python Compose contracts under `tests/test_compose_*.py` (see `tests/AGENTS.md`); run only the file(s) covering the overlay you changed.
- `uv run inv compose.health`, `compose.infra-health`, `compose.storage-health` for live verification — these are lead-run gates against the shared retained stack, not something a slice agent runs casually.

## Gotchas
- Docker builds can see `connection reset by peer` from `proxy.golang.org` during parallel `go mod download`; this is a known transient, not necessarily a config regression.
- `0.0.0.0` is a bind address, never a client/probe target; probe the configured reachable host (`127.0.0.1`, published port, or service DNS name) instead.
- Destructive proof runs only through `uv run inv lifecycle.first-boot-proof --isolated`, never ad hoc `docker compose down -v` or manual volume deletion.
