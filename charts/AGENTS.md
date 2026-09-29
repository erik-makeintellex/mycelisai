# Charts (Helm)

> Navigation: [Repo AGENTS.md](../AGENTS.md) | [Deploy AGENTS.md](../deploy/AGENTS.md)

## Owns / does not own
- `charts/mycelis-core/**` (the Helm chart: templates, values, `Chart.lock`) -> `mycelis-platform-ops` (`.claude/agents/`, tracked roster).
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
- `charts/mycelis-core/config/templates/` holds byte-identical copies of the `core/config/templates/` boot bundles (`mycelis-runtime-core.yaml`, the default, and `mycelis-dev-swarm-optional.yaml`), enforced by `tests/test_k8s_config_parity.py`; change both copies together. The chart sets no `MYCELIS_BOOTSTRAP_TEMPLATE_ID`, so Core's default selects `mycelis-runtime-core`. `mycelis-runtime-core.yaml` is at 383 of 385 lines: a new standing team needs its own bundle, not growth there.
- Kubernetes cluster enforcement is not certified; do not treat a passing chart render as proof the policy holds live.
