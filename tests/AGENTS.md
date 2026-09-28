# Tests (Python)

> Navigation: [Repo AGENTS.md](../AGENTS.md)

## Owns / does not own
- `tests/**` is a shared Python test/contract area; ownership follows the code area each file exercises, not one roster role (`.claude/agents/`, local-only and gitignored). Precedence: the more specific prefix wins over a general one.
  - `test_docs_links.py`, `test_documentation_layout_contract.py`, `test_canonical_workspace_docs.py`, `test_trusted_outcome_docs.py`, `test_user_help_docs.py` -> `mycelis-docs-steward`.
  - `test_db_*.py` and other schema-integrity/installer-compatibility tests -> `mycelis-schema` (wins over the `test_*_tasks.py` default below).
  - `test_cognitive_*.py`, `test_litellm_config_contract.py` -> `mycelis-ai-runtime` (wins over the `test_*_tasks.py` default below).
  - `test_framework_runs_*.py`, `test_compose_framework_runs.py` -> `mycelis-core-execution` (`services/framework-runs/**`, `framework_runs/`).
  - `test_auth_*.py` -> `mycelis-core-authority` where it exercises Core auth behavior, or `mycelis-platform-ops` where it exercises the `auth.dev-key` task.
  - `test_workflow_contracts.py` guards `.github/workflows/**`, which is lead-only regardless of who edits this test.
  - Everything else, including `test_*_tasks.py`, `test_compose_*.py`, `test_ci_*.py`, `test_k8s_*.py`, `test_lifecycle_*.py`, `test_cache_tasks.py`, `test_cleanup_tasks.py`, `test_core_dockerfile_resilience.py`, `test_search_runtime_config.py`, `test_media_gateway.py`, and `test_tooling_hygiene.py`, defaults to `mycelis-platform-ops`.
- Go tests live beside their package under `core/internal/<pkg>/*_test.go`, not here; this folder is Python-only test/task-contract coverage plus the docs contracts. `agents/tests/` is also a registered pytest testpath (`pyproject.toml`) outside this directory.

## Contracts
- Test-first (root `AGENTS.md` "Evidence And Completion"): reproduce a failure here as a test before fixing the code it targets.
- Python only through `uv`: `uv run pytest -q tests/<file>`, never bare `pytest`/`python -m pytest`.
- Evidence over confidence: report only tests you actually executed, with pass/fail/skip counts; fixture proof is not live proof.

## Gates
- Run the single file(s) relevant to your change, for example `uv run pytest -q tests/test_compose_tasks.py`; do not run the full `tests/` directory or `uv run inv core.test` from a slice — that is a lead-only gate.
- `git diff --check` for whitespace.

## Gotchas
- Several files here gate docs contracts, not task contracts (`test_docs_links.py`, `test_documentation_layout_contract.py`) — a docs-only change can still fail here if a link breaks or a required file moves.
- `test_product_architecture_library_has_one_prd_and_scoped_supporting_docs` (inside `test_docs_links.py`) asserts the exact file list of `docs/architecture-library/`; adding a file there without updating this test fails the gate.
