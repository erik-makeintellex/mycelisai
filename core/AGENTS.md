# Core (Go)

> Navigation: [Repo AGENTS.md](../AGENTS.md) | [Canonical PRD](../docs/architecture-library/MYCELIS_CANONICAL_PRD.md)

## Owns / does not own
- `core/**`: Go runtime, orchestration, APIs, NATS integrations, persistence-facing logic (root `AGENTS.md` "Language Ownership").
- No single writer owns all of `core/`; roster roles split it by sub-path (role files under `.claude/agents/`, local-only and gitignored — ask the lead for the current text if you can't see it):
  - `core/internal/identity/**`, `core/internal/invocation/**`, `core/internal/workerauthority/**`, `core/internal/mcp/**`, `core/internal/comms/**`, `core/internal/deploymentcontext/**`, `core/internal/server/auth*.go|identity*.go|groups_auth*.go|team_ownership*.go|invocations*.go|admin_routes.go` (auth wiring), `mcp*.go`, `organization*.go`, `groups*.go` (non-auth), `deployment*.go` -> `mycelis-core-authority` (lead decision 2026-09-28: MCP config/scope gates and comms/deployment/membership authority both pair with the identity/ownership surface this role already owns).
  - `core/internal/server/*dispatch*|templates_confirm_action*.go|templates_worker_execution*.go|templates_execution*.go|worker_event_projection*.go|framework_worker_authority*.go|team_work_*|cognitive*.go|soma*.go|mission*.go|outcome*.go|review*.go|qa*.go|output*.go|workspace*.go|artifacts*.go|memory*.go|token_budgets.go`, `core/internal/workers/**`, `core/internal/swarm/**`, `core/internal/dispatchoutbox/**`, `core/internal/outputvalidation/**`, `core/internal/runs/**`, `core/internal/memory/**`, `core/internal/agentry/**` -> `mycelis-core-execution` (lead decision 2026-09-28: Soma-turn, proposal, execution-request and memory paths).
  - `services/framework-runs/**` (separate Go module; gate `go -C services/framework-runs test -race ./...`) -> `mycelis-core-execution`.
  - `core/config/**` (cognitive, soma-commands, profiles), provider-routing Go under `core/internal/cognitive/**`, `brains*.go`, `cognitive_status_*.go`/`cognitive_config_*.go`, `capabilities.go` (provider/config semantics; `mycelis-core-execution` writes handler plumbing when only routing changes, not provider posture) -> `mycelis-ai-runtime`.
  - `core/migrations/**` -> `mycelis-schema`, the sole schema writer; see `core/migrations/AGENTS.md`.
  - `proto/**` -> `mycelis-core-execution` (lead decision 2026-09-28; see `proto/AGENTS.md`).
  - See the sub-area files for `core/internal/server`, `core/internal/cognitive`, `core/internal/swarm`.
- `mycelis-architect` (read-only, freezes contracts) and `mycelis-security-qa` (read-only, independent QA) never write here.

## Contracts
- [`docs/architecture-library/MYCELIS_CANONICAL_PRD.md`](../docs/architecture-library/MYCELIS_CANONICAL_PRD.md) is the single product/architecture authority.
- [`B1_TOKEN_BUDGETS_CONTRACT.md`](../docs/architecture-library/B1_TOKEN_BUDGETS_CONTRACT.md), [`G4_E10_DURABLE_INVOCATION.md`](../docs/architecture-library/G4_E10_DURABLE_INVOCATION.md), [`MCPA_MCP_CONFIGURATION_AUTHORITY_CONTRACT.md`](../docs/architecture-library/MCPA_MCP_CONFIGURATION_AUTHORITY_CONTRACT.md), [`MCPS_OPERATOR_MCP_SCOPE_CONTRACT.md`](../docs/architecture-library/MCPS_OPERATOR_MCP_SCOPE_CONTRACT.md) govern their named slices.
- Invariants: Core is the only authorization authority (Human/Soma -> BFF -> Core -> authority -> execution); models/NATS/MCP discovery/memory/prompts never grant permission; no placeholder or fake success (`respondBlocker` in `core/internal/server/blocker_copy.go` is the fail-closed pattern); use `protocol` subject constants, never hardcoded `swarm.*` literals; every identity/authority mutation writes `identity_audit_events` in the same transaction.
- Schema changes are never made directly by other Core writers; request DDL from `mycelis-schema` through the lead.

## Gates
- Targeted (the Go module is `core/`; run from `core/` or use `go -C core` — put the package path before `-run`, never inside the `-run` pattern): `go -C core test -race -count=1 ./internal/<pkg>/... -run '<pattern>'` and `go -C core vet ./internal/<pkg>/...` for the package you touched. Verified 2026-09-28 against dev `38e74c3f`: `go -C core test -race -count=1 ./internal/cognitive/... -run NONE`, `./internal/server/...` and `./internal/swarm/...` each print `ok ... [no tests to run]`; `go -C core vet` is clean on all three; `go -C services/framework-runs test -race ./... -run NONE` also prints `ok`/`[no test files]` for every package.
- Real PostgreSQL DSN required for persistence tests (for example `MYCELIS_INVOCATION_TEST_DSN`); the current schema (`core/migrations/001_current_schema.sql`) must be applied to that DSN first.
- Lead-only: `uv run inv core.test` (full suite), `uv run inv lifecycle.first-boot-proof` for startup/schema/persistence changes.

## Gotchas
- Authored `.go` files target 350 lines, hard cap 385 (`uv run inv quality.max-lines`); check `ops/quality_legacy_caps.txt` for existing exceptions before splitting a file you didn't grow.
- `core/internal/server` alone holds 500+ files; `rg`/grep before assuming a handler doesn't exist.
- Parallel slices can add same-named Go test helpers in one package; `go vet` at merge is what catches it, not CI on your branch. Prefix new helpers with your packet's helper prefix.
- `gofmt` is clean on dev (verified 2026-09-28: `gofmt -l ./internal ./cmd ./pkg` from `core/` lists 0 files); run `gofmt -l` on the files you touched and fix any drift you introduce.
