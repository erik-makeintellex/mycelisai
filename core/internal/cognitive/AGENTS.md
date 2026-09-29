# Core Cognitive (`core/internal/cognitive`)

> Navigation: [Core AGENTS.md](../../AGENTS.md) | [Repo AGENTS.md](../../../AGENTS.md)

## Owns / does not own
- Provider/profile routing, model configuration resolution, and the inference choke point.
- Owner: `mycelis-ai-runtime` (`.claude/agents/`, tracked roster) for provider-routing Go here plus `core/config/**` and `charts/mycelis-core/config/cognitive.yaml` (coordinate `cognitive.yaml` with platform-ops, since the chart copy must stay byte-identical to the source per the architecture-transition hygiene gate).
- Budget wiring in `core/internal/server/token_budgets.go` and `cognitive_profile_overrides.go` belongs to `mycelis-core-execution` (lead decision 2026-09-28: the B1 tail "Budget metering" target), not this package.

## Contracts
- `Router.InferWithContract` (`router.go`) is the single budget choke point: every inference call (Soma turn, council consult, team agents, D2 draft, agentry, memory recall, archivist, provisioning, architect) must go through it. Do not add a second inference entry point.
- [`B1_TOKEN_BUDGETS_CONTRACT.md`](../../../docs/architecture-library/B1_TOKEN_BUDGETS_CONTRACT.md) is the authority for budget/reservation/ledger behavior around this choke point.
- `finalizeInferenceResponse` (`router.go`) records provider-reported usage; `openai.go`, `anthropic.go` and `google.go` all map provider usage into `PromptTokens`/`CompletionTokens`/`TokensUsed` (verified on dev `38e74c3f`; the anthropic/google gap from the original B1 review is closed). A new provider adapter must do the same, or budgets under-count.
- Local-only data boundaries never fail over to a remote provider; no silent remote inference.

## Gates
- The Go module is `core/`; run from `core/` or use `go -C core`: `go -C core test -race -count=1 ./internal/cognitive/... -run '<pattern>'` and `go -C core vet ./internal/cognitive/...`. Verified 2026-09-28 against dev `38e74c3f`: `go -C core test -race -count=1 ./internal/cognitive/... -run NONE` prints `ok ... [no tests to run]`; `go -C core vet ./internal/cognitive/...` is clean.
- No capability-eval harness exists yet (`tests/evals/` is absent on dev); the ai-runtime role may add one there only after the lead approves the location.

## Gotchas
- The local model posture is Windows Ollama through the relay (next line); vLLM is an optional, currently stopped provider. If vLLM is re-enabled, Mycelis reads `MYCELIS_PROVIDER_VLLM_MODEL_ID`, and a model switch must be proven from inside the Core container (`host.docker.internal:8000/v1`), not just from the host.
- The live root is Windows Ollama (qwen3:14b) through the `mycelis-home-ollama-relay` container; `vllm-node` is stopped because WSL vLLM locked up. After a restart, run `uv run inv compose.warm-cognitive` before trusting any live inference proof. If vLLM is re-enabled, wait for `127.0.0.1:8000/v1/models` instead.
- `MaxOutputTokens` (`types.go`, default 1024 via `DefaultMaxTokensForBudget`) caps output per provider; committed values in `core/config/cognitive.yaml` are 1024 or 2048 depending on provider (`vllm`: 1024) — read the active config (`uv run inv cognitive.*` status task) before assuming a budget.
