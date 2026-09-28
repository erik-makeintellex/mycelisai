# Charts (Helm)

> Navigation: [Repo AGENTS.md](../AGENTS.md) | [Deploy AGENTS.md](../deploy/AGENTS.md)

## Owns / does not own
- `charts/mycelis-core/**` (the Helm chart: templates, values, `Chart.lock`) -> `mycelis-platform-ops` (`.claude/agents/`, local-only and gitignored).
- `charts/mycelis-core/config/cognitive.yaml` specifically must stay coordinated with `mycelis-ai-runtime`, since it mirrors `core/config/cognitive.yaml` and the architecture-transition hygiene gate requires packaged runtime config copies to be byte-identical to their canonical source.
- See [`deploy/AGENTS.md`](../deploy/AGENTS.md) for the Compose counterpart of packaging; the two are cross-linked because a topology change to one usually needs the same change reasoned through for the other.

## Contracts
- Helm workload/Secret/NetworkPolicy references remain source/render-tested, not live-cluster-certified: "Disabled Helm package is source/render-tested, not live cluster-certified" (`.state/V8_DEV_STATE.md` "B2 Private Deployment Delivery Evidence").
- Committed chart values use secret references only, never raw secrets; `.env`/`.env.compose` remain the runtime source of truth for anything sensitive.
- Kubernetes retains stricter directional policy than the Compose-only authenticated-Core exception the canonical PRD records (root `AGENTS.md` "Runtime Config And Proof Boundary").

## Gates
- `uv run inv k8s.standards` (`helm lint` plus `helm template` render, `ops/k8s_standards.py`) and `uv run pytest -q tests/test_k8s_chart_contract.py tests/test_k8s_config_parity.py tests/test_framework_worker_chart.py tests/test_k8s_standards_tasks.py` for the touched chart surface.
- `uv run inv quality.max-lines` covers `charts/**` `.yaml`/`.tpl` files.

## Gotchas
- `charts/mycelis-core/config/templates/v8-migration-standing-team-bridge.yaml` is capped at 533 lines in `ops/quality_legacy_caps.txt` (matching the `core/config/templates/` copy); don't grow it further, and don't "fix" the cap without also fixing the `core/config/` sibling.
- Kubernetes cluster enforcement is not certified; do not treat a passing chart render as proof the policy holds live.
