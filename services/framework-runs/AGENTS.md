# Framework Runs Service (`services/framework-runs`)

> Navigation: [Repo AGENTS.md](../../AGENTS.md) | [Core AGENTS.md](../../core/AGENTS.md)

## Owns / does not own
- `services/framework-runs/**`: a separate Go module (`go.mod` declares `github.com/mycelis/framework-runs`, not `github.com/mycelis/core`) with its own `cmd/framework-runs`, `internal/{auth,config,controller,executor,httpapi,journal,protocol}`, `migrations/`, and `Dockerfile` -> `mycelis-core-execution` (`.claude/agents/`, local-only and gitignored; the role file names this module explicitly).
- This is the confined Runs HTTP client/control service from the P0.10 B2 delivery, not the Core monolith; do not fold its authority logic into `core/internal/runs/**` without an explicit architecture decision.

## Contracts
- It is a separate Go module deliberately: authenticated, bounded readiness (`production_ready=false` until executor/dispatch is certified), scoped secret checks, and its own PostgreSQL store — it does not share Core's schema or process.
- The operator-approved Compose-only authenticated-Core control-peer exception (`.state/V8_DEV_STATE.md` "B2 Private Deployment Delivery Evidence") is canonical for this service; Kubernetes retains stricter directional policy and is not certified.
- No executor or candidate uploader is enabled by the current delivery; do not add live dispatch/execution here without the P0.10 C contract (identity/grant/invocation mapping) being accepted first.

## Gates
- Separate Go module, its own gate: `go -C services/framework-runs test -race ./...`. Verified 2026-09-28 against dev `38e74c3f`: every package prints `ok` (or `[no test files]` for `cmd/framework-runs`'s empty leaves and `migrations`); no failures.
- Not covered by `uv run inv core.test` (that gate is the `core/` Go module only) and not scanned by `uv run inv quality.max-lines` (`services/**` is outside `DEFAULT_SOURCE_PATHS` in `ops/quality.py`) — don't assume either gate exercises this module.
- Deployment/packaging proof (Dockerfile build, isolated first boot including this service) is a lead-run gate; see `deploy/AGENTS.md`.

## Gotchas
- Its `Dockerfile` builds `FROM scratch` with `CGO_ENABLED=0`; a change that requires cgo (for example a different PostgreSQL driver than `jackc/pgx/v5`) breaks the image, not just the Go build.
- `production_ready=false` is a real, load-bearing readiness value, not a placeholder to silence — do not flip it without the P0.10 C durability/control/validation work landing first (root `AGENTS.md` "no placeholder or fake success").
- Because this module is outside `quality.max-lines` and `core.test`, a regression here will not surface in the Core-focused gates other writers run; always run its own `go -C services/framework-runs test -race ./...` when you touch it.
