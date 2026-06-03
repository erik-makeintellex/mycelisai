# Release Handoff
> Navigation: [Project README](../README.md) | [Docs Home](README.md) | [Testing](TESTING.md)

> Status: IN_REVIEW
> Last Updated: 2026-06-02
> Purpose: Current release-candidate handoff packet for operator proof, packaging, and follow-on validation.

## Current RC Result

- RC proof date: 2026-05-19
- Runtime proof commit: `4c821cfb` (`Hydrate Helm repos for release packaging`)
- Handoff/package label: `rc-v8.2-2026-05-19-final`
- Result: hosted `Full Release Candidate` is green with image publishing disabled.
- Proven source gates: repo hygiene/Python tests, Core tests/vet, Interface tests/build/typecheck, and Helm standards.
- Proven browser gate: authenticated Chromium homepage proof.
- Current local-source GUI gate: headed/live dashboard proof must show fresh Soma entry, proposal approval feedback, current retained output, visible output digest file/folder access, and Output-first review before promotion proof is trusted.
- Proven source API gate: hosted pgvector PostgreSQL/NATS, migrations, Core health, and `uv run inv api.delivery-proof --read-only`.
- Proven release artifacts: enterprise and enterprise Windows-AI Helm verification bundles plus Core binary archives for linux/amd64, linux/arm64, darwin/amd64, darwin/arm64, and windows/amd64.
- Hosted workflow maintenance: GitHub workflow actions have been updated to the Node 24-capable major lines, Interface CI now runs on Node.js 24, and Helm setup avoids the still-Node-20 `azure/setup-helm` action by installing pinned Helm 3 from `get.helm.sh` with checksum verification.

## Deployment Lanes

- Docker Compose: rapid local development, same-machine proof, home-lab, and demo runtime.
- Rancher Desktop K3s: preferred Windows local Kubernetes RC and commercial-parity proof lane.
- k3d: preferred WSL/Linux local Kubernetes validation lane.
- Enterprise Helm/Kubernetes: target clustered deployment contract with real ingress, storage, secrets, registry, and policy controls.
- WSL Compose: optional secondary deployment-mimic proof lane when the release packet needs that evidence.

## Rancher K3s RC Commands

Use a committed, clean tree and prepend Rancher Desktop CLI tools to `PATH` if `docker` or `rdctl` are not already resolved:

```powershell
$env:PATH="C:\Program Files\Rancher Desktop\resources\resources\win32\bin;$env:PATH"
$env:MYCELIS_K8S_BACKEND="rancher"
$env:MYCELIS_K8S_VALUES_FILE="charts/mycelis-core/values-k3d.yaml"
$env:MYCELIS_K8S_TEXT_ENDPOINT="http://<windows-ai-host>:11434/v1"
$env:MYCELIS_K8S_TEXT_MODEL_ID="qwen3:8b"
$env:MYCELIS_K8S_SEARCH_PROVIDER="searxng"
$env:MYCELIS_K8S_SEARXNG_ENDPOINT="http://<windows-ai-host>:8088"

uv run inv k8s.status
uv run inv k8s.deploy
uv run inv k8s.wait --timeout=300
uv run inv k8s.bridge
$env:MYCELIS_API_HOST="127.0.0.1"
$env:MYCELIS_API_PORT="8081"
uv run inv lifecycle.health
```

Pass criteria:
- Rancher K3s node is Ready.
- PostgreSQL, NATS, and Core are ready in namespace `mycelis`.
- Core bridge serves `/healthz` on `127.0.0.1:8081`.
- `/api/v1/search/status` reports `provider=searxng`, `online_allowed=true`, `approval_mode=notify`, and `disclosure_mode=notice_and_interpretation` when online search proof is in scope.
- `lifecycle.health` reports all backend endpoints healthy.

## Live GUI Proof Commands

Use K8s/PVC probing when browser specs assert backend-written workspace files:

```powershell
$env:PLAYWRIGHT_BACKEND_WORKSPACE_PROBE="k8s"
$env:PLAYWRIGHT_K8S_NAMESPACE="mycelis"
$env:PLAYWRIGHT_K8S_CORE_SELECTOR="app=mycelis-core"
$env:PLAYWRIGHT_K8S_BACKEND_WORKSPACE_ROOT="/data/workspace"

uv run inv interface.e2e --headed --live-backend --project=chromium --workers=1 --server-mode=dev --spec=e2e/specs/soma-governance-live.spec.ts
uv run inv interface.e2e --headed --live-backend --project=chromium --workers=1 --server-mode=external --spec=e2e/specs/dashboard-workbench-live-review.spec.ts
uv run inv interface.e2e --headed --live-backend --project=chromium --workers=1 --server-mode=dev --spec=e2e/specs/team-creation.spec.ts
uv run inv interface.e2e --headed --live-backend --project=chromium --workers=1 --server-mode=dev --spec=e2e/specs/groups-live-backend.spec.ts
uv run inv interface.e2e --headed --live-backend --project=chromium --workers=1 --server-mode=dev --spec=e2e/specs/workspace-live-backend.spec.ts
```

Current RC evidence:
- `team-execution-live.spec.ts`: passed
- `soma-governance-live.spec.ts`: 4 passed
- `team-creation.spec.ts`: passed
- `groups-live-backend.spec.ts`: passed
- `workspace-live-backend.spec.ts`: passed

## Regression Evidence

Committed-tree evidence for the RC proof included:
- `git show --check --stat --oneline HEAD`
- `uv run inv quality.max-lines --limit 300`
- `uv run pytest tests/test_docs_links.py tests/test_documentation_layout_contract.py tests/test_runtime_deploy_contract_text.py tests/test_k8s_deployment_standards_docs.py -q`
- `uv run pytest tests/test_k8s_tasks.py tests/test_k8s_chart_contract.py tests/test_k8s_standards_tasks.py tests/test_wsl_runtime_tasks.py tests/test_interface_e2e_tasks.py tests/test_quality_tasks.py -q`
- `uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-k3d.yaml`
- `uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml`
- `uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise-windows-ai.yaml`
- `uv run inv core.test`
- `uv run inv interface.typecheck`
- `uv run inv interface.build`
- `uv run inv interface.test`

## Packaging Commands

Use the docs-inclusive handoff label unless the exact runtime-proof commit must be packaged:

```powershell
$env:RELEASE_LABEL="v0.6.0-98f83b8"

uv run inv core.package --target-os=linux --target-arch=amd64 --version-tag=$env:RELEASE_LABEL
uv run inv core.package --target-os=linux --target-arch=arm64 --version-tag=$env:RELEASE_LABEL
uv run inv core.package --target-os=darwin --target-arch=amd64 --version-tag=$env:RELEASE_LABEL
uv run inv core.package --target-os=darwin --target-arch=arm64 --version-tag=$env:RELEASE_LABEL
uv run inv core.package --target-os=windows --target-arch=amd64 --version-tag=$env:RELEASE_LABEL

uv run inv k8s.deploy --verify-package --values-file=charts/mycelis-core/values-enterprise.yaml --package-output-dir=dist/helm/enterprise --release-label=$env:RELEASE_LABEL
uv run inv k8s.deploy --verify-package --values-file=charts/mycelis-core/values-enterprise-windows-ai.yaml --package-output-dir=dist/helm/enterprise-windows-ai --release-label=$env:RELEASE_LABEL
```

Generated artifact set:
- Core binary archives, manifests, and checksum files for `linux/amd64`, `linux/arm64`, `darwin/amd64`, `darwin/arm64`, and `windows/amd64`.
- Enterprise Helm verification bundle in `dist/helm/enterprise/`.
- Enterprise Windows-AI Helm verification bundle in `dist/helm/enterprise-windows-ai/`.

Helm verification prerequisites:
- `bitnami` repo: `https://charts.bitnami.com/bitnami`
- `nats` repo: `https://nats-io.github.io/k8s/helm/charts/`
- The hosted `Release Packaging`, `CI`, and `Full Release Candidate` workflows install `HELM_VERSION=v3.20.2` from the official Helm binary distribution with checksum verification, then add and update these Helm repos before verification so a clean GitHub runner can package without prior Helm repo state or Node-20 Helm action warnings.

## Optional WSL Compose Proof

Run this only when secondary deployment-mimic evidence is needed:

```powershell
uv run inv wsl.status
uv run inv wsl.refresh
uv run inv wsl.validate --lane=release
uv run inv wsl.validate --lane=release --headed-browser
```

## Known Notes

- Rancher Desktop K3s is local RC parity proof, not the production target cluster.
- Enterprise release still uses Helm/Kubernetes with real ingress, storage, secrets, registry, and policy controls.
- AI endpoints must be explicit and reachable from the cluster.
- Do not record raw secrets in release notes, logs, docs, or state files.
- Current packaging is artifact packaging only: no installer, SBOM, signing, or provenance attestation is claimed.
