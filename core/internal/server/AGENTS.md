# Core HTTP Server (`core/internal/server`)

> Navigation: [Core AGENTS.md](../../AGENTS.md) | [Repo AGENTS.md](../../../AGENTS.md)

## Owns / does not own
- HTTP handlers, route registration (`admin_routes.go`), request/response envelopes, and the blocker pattern for this 500+ file package.
- Split by roster role (`.claude/agents/`, local-only and gitignored), not one writer. Lead ownership decisions, 2026-09-28:
  - `auth*.go`, `identity*.go`, `groups_auth*.go`, `team_ownership*.go`, `invocations*.go`, `admin_routes.go` (auth wiring only), `mcp*.go`, `organization*.go`, `groups*.go` (non-auth), `deployment*.go` -> `mycelis-core-authority`. Every MCP slice to date (MCPS W1, MCPA-H, MCPA W1, RED, MCPL W1) was core-authority, and these files are the scope/authority gates; organization/groups/deployment are ownership and deployment-context authority (M1/M2) and pair with `team_ownership*`/`groups_auth*`.
  - `*dispatch*`, `templates_confirm_action*.go`, `templates_worker_execution*.go`, `templates_execution*.go`, `worker_event_projection*.go`, `framework_worker_authority*.go`, `team_work_*`, `cognitive*.go` (the 50+ Soma-turn/proposal files, not `cognitive_status_*`/`cognitive_config_*` below), `soma*.go`, `mission*.go`, `outcome*.go`, `review*.go`, `qa*.go`, `output*.go`, `workspace*.go`, `artifacts*.go`, `memory*.go`, `token_budgets.go` -> `mycelis-core-execution`. These are Soma-turn, proposal, execution-request and budget paths (MEM and the B1 tail "Budget metering" targets are both core-execution).
  - `brains*.go`, `cognitive_status_*.go`, `cognitive_config_*.go`, `capabilities.go` -> `mycelis-ai-runtime` for provider/config semantics; `mycelis-core-execution` writes handler plumbing only when routing changes, never provider posture.
  - `catalogue.go` (worker profiles and team composition) → `mycelis-core-execution`; `code_context.go` (context retrieval for models) → `mycelis-ai-runtime` (lead decision 2026-09-28).
  - Any remaining file not named above has no standing writer; check `.state/V8_DEV_STATE.md` "Delivery Targets And Teams" for the active packet before editing, and coordinate through the lead if two slices touch the same file.

## Contracts
- [`MCPA_MCP_CONFIGURATION_AUTHORITY_CONTRACT.md`](../../../docs/architecture-library/MCPA_MCP_CONFIGURATION_AUTHORITY_CONTRACT.md) and [`MCPS_OPERATOR_MCP_SCOPE_CONTRACT.md`](../../../docs/architecture-library/MCPS_OPERATOR_MCP_SCOPE_CONTRACT.md) govern every `mcp*.go` route.
- [`B1_TOKEN_BUDGETS_CONTRACT.md`](../../../docs/architecture-library/B1_TOKEN_BUDGETS_CONTRACT.md) governs `token_budgets.go` and any route that calls `Router.InferWithContract`.
- Fail-closed audit precedent: `createAuditEvent` (`templates_audit.go`) returns an empty id and an error on failure; callers must 503 on an empty audit id (see `token_budgets.go`), never proceed silently.
- The blocker envelope (`blocker_copy.go`, `respondBlocker`) is the only accepted "can't do this yet" response shape; no route may fake a success payload instead.
- Default-deny once a route's scope requirement is accepted; `AuthMiddleware` (`cmd/server/main.go`, `auth.go`) only resolves identity, it does not authorize — each handler must check its own scope.

## Gates
- The Go module is `core/`; run from `core/` or use `go -C core` — put the package path before `-run`, never inside the `-run` pattern: `go -C core test -race -count=1 ./internal/server/... -run '<TestPattern>'` and `go -C core vet ./internal/server/...`. Verified 2026-09-28 against dev `38e74c3f`: `go -C core test -race -count=1 ./internal/server/... -run NONE` prints `ok ... [no tests to run]`; `go -C core vet ./internal/server/...` is clean.
- Auth/identity changes with persistence need a real PostgreSQL DSN under `--race`.

## Gotchas
- MCP routes are scope-gated (MCPS/MCPA, closed `2a7f7699`; MCPL `cb0aa476`, verified on dev `38e74c3f`): direct tool calls go through `authorizeMCPToolCall` (`mcp_call_authority.go`, checks `hasScope(identity, mcpDirectReadScope)`), and config/toolset writes go through `requireRootAdminScope(..., scopeMCPConfigWrite)` (`mcp_config_authority.go`). A new `mcp*` route must reuse these gates, not add its own. Open follow-ups are listed in the MCPS contract's closed-follow-ups section and the MCPL merge body.
- `mcp:write` is deliberately not usable as a scope name because it also parses as a valid agent tool ref (`mcp_config_authority.go:25`, `mcp/toolref.go` `ParseToolRef`; see the MCPA contract "mcp:write is also a valid agent tool ref") — don't reuse that string for a new scope name without checking the parser.
- This directory alone is at the file-count scale where `rg`/grep before editing is mandatory; do not read it file-by-file.
