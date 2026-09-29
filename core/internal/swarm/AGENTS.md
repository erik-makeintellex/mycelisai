# Core Swarm (`core/internal/swarm`)

> Navigation: [Core AGENTS.md](../../AGENTS.md) | [Repo AGENTS.md](../../../AGENTS.md)

## Owns / does not own
- Agent dispatch, tool execution/scope resolution, team-work lifecycle, ReAct loop control.
- Owner: `mycelis-core-execution` (`.claude/agents/`, tracked roster), alongside `core/internal/dispatchoutbox/**`, `core/internal/outputvalidation/**`, `core/internal/runs/**`, and the matching `core/internal/server` dispatch/templates files (see `core/internal/server/AGENTS.md`).
- Tool-scope/toolset authority (`tool_scope.go`, `tool_executor.go`) is dual-relevant to MCP configuration: coordinate with `mycelis-core-authority` through the lead before changing how `mcp:` refs are resolved, since [`MCPA_MCP_CONFIGURATION_AUTHORITY_CONTRACT.md`](../../../docs/architecture-library/MCPA_MCP_CONFIGURATION_AUTHORITY_CONTRACT.md) governs the toolset write path in `core/internal/server`.

## Contracts
- Use canonical NATS subject constants from the `protocol` package; never hardcode `swarm.*` literals (root `AGENTS.md` "NATS Signal Standard" has the full subject family list and required metadata: `run_id`, `team_id`, `agent_id`, `source_kind`, `source_channel`, `payload_kind`, `timestamp`).
- Nothing leaves the request boundary before approval, WorkIntent, ExecutionContract, ownership, run and idempotent dispatch intent commit.
- Events are untrusted and replayable: claim durable receipt and advance cursor in one transaction; handle gap/duplicate/stale/wrong-scope/competing-consumer cases explicitly.
- Worker `completed` status is candidate evidence only, never trusted completion by itself ([`TRUTHFUL_DELIVERY_AND_NEXT_ARCHITECTURE.md`](../../../docs/architecture-library/TRUTHFUL_DELIVERY_AND_NEXT_ARCHITECTURE.md)).

## Gates
- The Go module is `core/`; run from `core/` or use `go -C core`: `go -C core test -race -count=1 ./internal/swarm/... -run '<pattern>'` and `go -C core vet ./internal/swarm/...`. Verified 2026-09-28 against dev `38e74c3f`: `go -C core test -race -count=1 ./internal/swarm/... -run NONE` prints `ok ... [no tests to run]`; `go -C core vet ./internal/swarm/...` is clean.
- Crash-window and response-loss behavior needs real PostgreSQL and counted external calls under `--race`.

## Gotchas
- Toolset expansion is cached per agent scope instance (`tool_scope.go` field `resolved`; the resolver itself is `mcp.ToolSetService.ResolveRefs` in `core/internal/mcp/toolsets.go`, not this package): narrowing a toolset does not reach live agents (open MCPA item "Live-agent toolset revocation"). A fix here needs a cache-invalidation story, not just a data change.
- ReAct loop bounds: `protocol.DefaultMaxIterations = 3` (`core/pkg/protocol/manifest.go`, overridable per manifest via `max_iterations`); internal-tool agents default to 6 (`internal_tools_support.go`); `agent_tool_loop.go` applies its own result-contract loop limit and `inferenceStopped` for early exit. Don't raise any of these without an approved contract.
- `swarm/` and `dispatchoutbox/`/`runs/` are split by concern, not by directory depth; check `core/internal/server` for the matching HTTP-side file before assuming a change is swarm-only.
