# Ops (Python task runner and platform automation)

> Navigation: [Repo AGENTS.md](../AGENTS.md) | [Ops README](README.md)

## Owns / does not own
- `ops/**`, `tasks.py`, `docker-compose.yml`, `deploy/**`, `charts/**`, `.github/workflows/**`, `pyproject.toml`/`uv.lock` (only when a slice needs it) -> `mycelis-platform-ops` (`.claude/agents/`, tracked roster).
- The "Task runner tightening" delivery target is CLOSED (see `.state/V8_DEV_STATE.md` "Delivery Targets And Teams"). Task-name or behavior changes to `tasks.py`, `ops/*.py`, `README.md`, `ops/README.md`, `docs/TESTING.md`, and `docs/architecture/OPERATIONS.md` follow the docs-sync rule in the root `AGENTS.md` "Task Runner Contract".
- `ops/db*.py` is schema-installer territory and is shared with `mycelis-schema` only where installer behavior is involved.
- `.github/workflows/**` is never edited by a non-lead agent regardless of role (root `AGENTS.md` worktree safety rule extends here: CI workflow changes are a lead-only action in practice even though the role file lists the path).

## Contracts
- Task runner contract (root `AGENTS.md` "Task Runner Contract"): `uv run inv ...` only, no new Invoke aliases, keep the registered task surface within the budget pinned in `tests/test_tasks_root.py` (`test_root_task_surface_stays_within_operator_budget`; lower it whenever tasks are removed). **Get the live count and names from `uv run inv -l` (or `uvx --from invoke inv -l` as a compatibility probe only) rather than from this file or any other document.** Windows is only a host (Ollama); every task runs from the WSL `dev` checkout, and no task drives a second checkout or publishes raw to `swarm.*` subjects.
- Committed config (`.env.compose`, Compose/Helm values) uses secret references only; `.env` remains the sole secret store and is never printed or committed.
- Destructive proof only through `lifecycle.first-boot-proof --isolated`; never delete retained volumes, data-plane state, or external containers (vLLM, Open WebUI, NATS, PostgreSQL data plane) from a slice.
- Live proof (`compose.health`, `test.probe`) must run on current app images: `ops/compose_images.py` fails on a core/interface container whose image id differs from its tag; `test.probe --allow-stale` is the only override.

## Gates
- Python tests for this area live under `tests/` (see [`tests/AGENTS.md`](../tests/AGENTS.md)); run the specific file(s) covering the task you changed, for example `uv run pytest -q tests/test_compose_tasks.py` — get the exact file name from `tests/` rather than assuming.
- `git diff --check` for whitespace; `uv run inv quality.max-lines` for the line-cap gate (default scope in `ops/quality.py` `DEFAULT_SOURCE_PATHS`: `core, interface, ops, tests, agents, sdk/python/src, cognitive/src, docs, charts, architecture, proto, README.md, AGENTS.md, pyproject.toml`; `deploy/**` and `services/**` are **not** scanned — see `ops/quality_legacy_caps.txt` for existing exceptions in the scanned paths).
- Python only through `uv`: `uv run python ...`, `uv run pytest ...`, `uv run inv ...`, `uvx <tool>`. Never bare `python3`/`python`/`pip`.

## Gotchas
- Discover configured targets (ports, hosts, Docker owner, provider endpoints) from `.env`/`.env.compose`/task defaults instead of assuming a host; see root `AGENTS.md` "Runtime Config And Proof Boundary" and "Configured Service Target Standard" for the full discovery order.
- A writer in this area never starts, stops, or recreates the shared `mycelis-home-*` stack or `vllm-node` (stopped; the root is Windows Ollama); the only allowed container is one disposable `docker run --rm` Postgres for DB tests.
- WSL shuts down its VM when no `wsl.exe` session is active, and containers restart on the next boot — a "service down" symptom may just be an idle VM, not a real regression.
