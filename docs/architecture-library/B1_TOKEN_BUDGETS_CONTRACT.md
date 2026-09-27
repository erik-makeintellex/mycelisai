# B1 Token Budgets Contract

Owner: AI runtime and Core authority. Product authority remains the [canonical PRD](MYCELIS_CANONICAL_PRD.md).

> Frozen contract for per-execution token budgets (owner-approved 2026-09-27). Implementation branch: `feature/token-budgets`. Related: [Truthful Delivery And Next Architecture](TRUTHFUL_DELIVERY_AND_NEXT_ARCHITECTURE.md).


Worktree `token-budgets`, branch `feature/token-budgets`, base `dev` @9f6c8d98 (rebase after M1, PH-D, H1 and U1 merge). Slice id `B1`. Owner request 2026-09-27: budgets resolved from the model in use or a default, modifiable per agentry execution, with hard stops.

## Verified today (source facts)
- Choke point exists: every inference goes through `Router.InferWithContract` (cognitive/router.go:224). Callers: swarm agents via `inferWithExecutionBounds` (swarm/agent_inference_bounds.go:29; covers the Soma turn = `admin` council member, council consults (server/cognitive_council.go:122), and team agents), the D2 draft (server/cognitive_write_file_draft.go:116), agentry (agentry/runner.go:54,166), memory recall tool (swarm/internal_tools_handlers_memory_recall.go:117), archivist, provisioning, and architect.
- Accounting: `finalizeInferenceResponse` records provider-reported `TokensUsed` into global atomics only (router.go:282-287, `RecordTokens` :45). There is no per-run, agent or team ledger.
- **Gap:** only `openai.go:134-136` fills usage. `anthropic.go:66-75` and `google.go:73-87` never parse usage, so hosted calls report 0 tokens today.
- Output caps today: `ProviderConfig.TokenBudgetProfile` and `MaxOutputTokens` (types.go:131-132, 163-199). The local qwen3:14b root uses 2048.
- Correlation {RunID, TeamID, AgentID} is set for swarm agents (swarm/agent_processing.go:273). D2 and agentry send none.
- ReAct: `loopLimit` (swarm/agent_tool_loop.go:88; DefaultMaxIterations=3, internal tools 6), plus correction re-infers. `inferenceStopped` (:30) exists. The team response carries `Availability.Code` (swarm/agent_bus.go:134-140).
- D2 cap: `writeFileDraftMaxPerTurn = 3` (cognitive_write_file_draft.go:22), with per-draft and turn time bounds (:17, :23).
- Storage precedent: `config_documents` (schema:2101) plus activations/history. Built-ins are seeded only by `SeedBuiltInRevisions` (configdocuments/builtin_seed.go:124) from `config/documents/templates` (cmd/server/startup_config_documents.go:15). Kinds live at protocol/config_documents.go:21-23.
- Authority: `requireRootAdminScope` (server/groups_auth.go:27) and `cognitiveWriteScope` (server/cognitive_profile_overrides.go:18; handler :74). A2b `approverCostCeiling = 5.0` (server/governance_approver_tier.go:69). Blockers: `respondBlocker` (server/blocker_copy.go:118).

## Decisions
**D1. Units.** An *execution* is one agent turn (the Soma turn, a team agent turn, a council consult), one D2 drafting pass per chat turn, one agentry `Run`, or one system call (archivist, provisioning, architect, recall). A *run* is every execution sharing a RunID. A *period* is a UTC day per team and per agent. The tokens counted are prompt plus completion as reported by the provider.

**D2. Resolution (most specific wins, per field):** agent → team → profile → model class → global. Fields are `per_execution`, `per_run`, `per_team_day`, `per_agent_day`, and `warn_pct`. The class comes from a new optional `ProviderConfig.budget_class`. If that is unset, `normalizedDataBoundary` decides: `local_only` → `local_large`, `leaves_org` → `hosted_premium` (it fails toward tighter). System executions use only per_execution and are excluded from team/agent day caps.

**D3. Defaults (built-in seed `config/documents/templates/token-budget-defaults.yaml`, new kind `TokenBudgetPolicy`):**
| class | per_execution | per_run | per_team_day | per_agent_day |
|---|---|---|---|---|
| local_large (qwen3:14b root, in use) | 64,000 | 256,000 | 2,000,000 | 1,000,000 |
| local_small | 24,000 | 96,000 | 1,000,000 | 500,000 |
| hosted_standard | 32,000 | 128,000 | 300,000 | 150,000 |
| hosted_premium | 24,000 | 96,000 | 150,000 | 75,000 |
| global (fallback) | 32,000 | 128,000 | 500,000 | 250,000 |
`warn_pct` is 80 everywhere. Rationale for 64k: about 8 calls per turn (initial, policy correction, up to 6 loop iterations, contract correction), each ~6k prompt plus ≤2k completion.

**D4. Enforcement (one choke point).** A new `cognitive/budget.go` holds the resolver and `cognitive/budget_meter.go` holds `ExecutionMeter`. The meter travels in `context.Context` (`cognitive.WithExecutionMeter`) and is created at unit entry: swarm per-message processing, the D2 turn, agentry `Run`. `InferWithContract` is the only place that checks and charges:
- **Before the call:** resolve limits and compute `remaining` as the minimum across the execution, run, team-day and agent-day scopes. If `remaining < 256`, refuse with `ErrTokenBudgetExhausted{Scope, Used, Limit, ResetsAt}` and make no provider call. Otherwise clamp `opts.MaxTokens = min(provider MaxOutputTokens, remaining)`.
- **After the call:** charge the reported usage in `finalizeInferenceResponse` and keep `RecordTokens`.
- A call with no meter in ctx gets a single-call meter that is still charged to the period scopes via Correlation.
- Known overshoot: at most one call's prompt. This is documented, not hidden.

**D5. Exhaustion is honest.**
- Swarm: `ErrTokenBudgetExhausted` becomes `ProcessResult.Availability{Code:"token_budget_exhausted"}` plus `inferenceStopped=true`. Any text already produced is returned labelled partial. The execution summary status is `stopped_budget` and never completed or verified.
- HTTP surfaces use `respondBlocker(429, "token_budget_exhausted")` with data {scope, used, limit, resets_at, recommended_action}. The user copy is "This work stopped because it reached its token budget." For an admin, the action names the override path.
- Crossing `warn_pct` emits one `token_budget_warning` mission event per scope per execution or day.

**D6. Unknown usage.** Parse usage in the anthropic adapter (`usage.input_tokens/output_tokens`) and the google adapter (`usageMetadata.promptTokenCount/candidatesTokenCount/totalTokenCount`). If a provider still reports none, charge the clamped `opts.MaxTokens` (a configured reservation, not a text estimate) and write `usage_reported=false`. The UI then shows "at least N".

**D7. Ledger (DB, append-only; the meter itself is in-memory per execution).** Append this block to `core/migrations/001_current_schema.sql` before `COMMIT;`:
```sql
-- BEGIN TOKEN_USAGE_LEDGER_EXTENSION
CREATE TABLE IF NOT EXISTS token_usage_ledger (
    id BIGSERIAL PRIMARY KEY,
    tenant_id TEXT NOT NULL DEFAULT 'default',
    occurred_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    execution_id TEXT NOT NULL,
    execution_kind TEXT NOT NULL CHECK (execution_kind IN ('soma_turn','agent_turn','council_consult','draft','agentry','system')),
    run_id TEXT NOT NULL DEFAULT '', team_id TEXT NOT NULL DEFAULT '', agent_id TEXT NOT NULL DEFAULT '',
    provider_id TEXT NOT NULL, model_id TEXT NOT NULL, budget_class TEXT NOT NULL,
    prompt_tokens INT NOT NULL DEFAULT 0, completion_tokens INT NOT NULL DEFAULT 0, total_tokens INT NOT NULL,
    usage_reported BOOLEAN NOT NULL, outcome TEXT NOT NULL CHECK (outcome IN ('charged','refused_exhausted'))
);
CREATE INDEX IF NOT EXISTS idx_token_ledger_team_day ON token_usage_ledger(tenant_id, team_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_token_ledger_agent_day ON token_usage_ledger(tenant_id, agent_id, occurred_at DESC);
CREATE INDEX IF NOT EXISTS idx_token_ledger_run ON token_usage_ledger(tenant_id, run_id) WHERE run_id <> '';
-- END TOKEN_USAGE_LEDGER_EXTENSION
```
- Period totals are read once per scope and then kept as in-memory write-through counters.
- If the DB is unavailable, period caps run on since-boot counters and every surface labels them "since restart" (owner Q2). There is no purge in B1.

**D8. Overrides API** (new `server/token_budgets.go`, routes in `admin_routes.go`):
- `GET /api/v1/cognitive/budgets`: effective policy. Any authenticated user may read it; override provenance is shown to admins only.
- `GET /api/v1/cognitive/budgets/usage?team_id|agent_id|run_id`: {used, limit, remaining, warn, period, usage_reported}.
- `PUT /api/v1/cognitive/budgets/overrides/{level}/{ref}` and `DELETE` on the same path, with `level ∈ agent|team|profile|class`. Both require `requireRootAdminScope(w,r,cognitiveWriteScope)`.
- Validation: integers in [1,024, 5,000,000]. There is no "unlimited" value, and `per_execution ≤ per_run ≤ per_team_day` is enforced after merging.
- Storage: each change writes a new operator-scope `TokenBudgetPolicy` revision (source `api`) and activates it in one transaction, like the S6 override pattern. Audit is the activation history plus a governed mission event `token_budget_changed` {actor, level, ref, before, after}.
- Budgets never change A2b tiers: cost > 5.0 still needs an approver, and a budget stop is not an approval request.

**D9. D2 mapping.** Remove `writeFileDraftMaxPerTurn`. The drafting pass runs under a `draft` meter charged to agent `admin`, and Correlation is now set.
- Preflight: if n × provider `MaxOutputTokens` exceeds remaining, return the `token_budget_exhausted` blocker with copy "can draft K of N" and propose nothing.
- Exhaustion mid-pass also returns the blocker and proposes nothing.
- The time bounds (:17, :23) stay.

**D10. UI summary.** Examples: "This team used 12k of 64k tokens this run · 180k of 2M today." and "Soma stopped: token budget reached."
- Not U1-dirty (edit now): `interface/lib/blockerCopy.ts` (add the code), new `interface/lib/tokenBudgets.ts` (client), new `interface/components/settings/TokenBudgetsPanel.tsx` mounted under Advanced in `interface/components/settings/BrainsPage.tsx` (557 lines, legacy cap: net-zero mount), new `interface/components/teams/TeamTokenUsage.tsx`.
- U1-owned (mount only after U1 merges): `components/teams/TeamDetailDrawer.tsx` (team usage line) and `components/soma/ExecutionSummaryCardModel.ts` (`stopped_budget` status and partial label).

## Test-first acceptance (-race; fake adapters report exact usage)
1. A run that goes over per_execution stops with `token_budget_exhausted`, the provider is not called once remaining < 256, the status is `stopped_budget`, and no completed/verified text appears.
2. Defaults resolve by class: local_only → local_large 64k and leaves_org → hosted_premium 24k. An explicit `budget_class` wins.
3. A team override beats the class default, an agent override beats the team, and DELETE restores the default.
4. PUT by a non-admin → 403 `admin_required`. An admin without `cognitive:write` → 403 with required_scope. Adversarial cases → 400: 0, negative, above max, per_execution > per_run, an unknown level, and a spoofed actor in the body (ignored).
5. Ledger totals equal the sum of adapter-reported usage for openai, anthropic and google fixtures. Missing usage → reservation charged with `usage_reported=false`.
6. The 80% warning emits exactly one event. MaxTokens clamps to remaining.
7. D2: preflight refusal, and mid-pass exhaustion → blocker with no proposal. The old constant is gone.
8. With the DB down, period caps use since-boot counters, and the usage API reports `period:"since_restart"`.
9. Interface: blockerCopy snapshot for user and admin, the panel renders the effective table and hides override controls for non-admins, and the team usage line formats k/M.
Gates: targeted -race, `uv run inv core.test`, `interface.test`, `interface.typecheck`, max-lines, diff --check, and the docs-links test.

**Live probe J6 "budget hit → honest stop"** (lead runs it; needs the stack):
1. As admin, PUT a team override with per_execution=2,000.
2. Give that team a multi-step task.
3. Expect the blocker copy, no "Result verified", usage used ≥ 2,000 of 2,000, and a `token_budget_changed` plus a warning event in audit.
4. DELETE the override and rerun: the run completes.

## Writers (max 2)
- **W1 Go (Opus), about 1,000 lines including tests, all new files ≤385.** Owns: cognitive/{budget,budget_meter}.go, adapter usage parsing, a ≤15-line hook in router.go, the protocol kind with its validation and compile, the seed yaml, the schema block, server/token_budgets.go with routes, the swarm/D2/agentry meter wiring, and the tests.
- **W2 interface (Sonnet), about 350 lines,** starting after W1's API is frozen. The two U1 mounts wait for the U1 merge.
- Docs: API_REFERENCE, OPERATIONS (net-zero), user docs for settings and teams, docsManifest if a doc is added, and the Area Contracts block.
- Do not touch: `.github/workflows`, the A2b tier code, or U1 files before the U1 merge.

## Owner questions (recommended default in bold)
- Q1. Are the local_large (14B root) defaults right: 64k per execution, 256k per run, 2M per team per day? **Yes, as tabled.**
- Q2. When the DB is down, should period caps run on "since restart" counters or fail closed? **Since-restart counters, labelled.**
- Q3. Should exhaustion offer a one-time admin "continue with more tokens" grant? **No in B1; the override-and-retry path only.**

## Owner answers (2026-09-27)
- Q1: local_large defaults 64k per execution, 256k per run, 2M per team per day (as tabled).
- Q2: with the DB down, period caps use since-restart counters and are labelled as such.
- Q3: no one-time top-up in B1; the admin overrides and retries.
