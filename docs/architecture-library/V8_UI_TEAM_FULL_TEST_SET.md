# V8 UI Team Full Test Set
> Navigation: [Project README](../../README.md) | [Testing](../TESTING.md)

Status: ACTIVE.

This is the compact contract for the full V8 browser/operator test set. The older browser-workflow and live-proof split notes were removed; keep durable browser proof expectations here and detailed commands in [Testing](../TESTING.md).

## 1. Test Goal

The UI test set proves that:
- Mycelis reads as a product, not a dev console
- Soma is the persistent counterpart
- AI Organizations are governed work contexts
- direct answers and team-managed outputs remain distinct
- governed execution is trustworthy
- settings and continuity persist
- advanced depth remains reachable without polluting the default path

## 2. Supported Test Environments

### 2.1 Compose runtime

Use for home-runtime, partner-demo, and single-host product validation:

```bash
uv run inv compose.down --volumes
uv run inv compose.up --build --wait-timeout=240
uv run inv compose.status
uv run inv compose.health
uv run inv interface.check
```

### 2.2 Local Kubernetes runtime

Use only when cluster behavior matters. Prefer Rancher Desktop K3s on Windows, prefer k3d for WSL/Linux local cluster proof, and use Kind only when explicitly selected.

```bash
uv run inv lifecycle.down
uv run inv k8s.up
uv run inv lifecycle.up --frontend
uv run inv lifecycle.health
```

For Rancher Desktop K3s proof, set `MYCELIS_K8S_BACKEND=rancher`, use `charts/mycelis-core/values-k3d.yaml` as the shared local-Kubernetes preset, and use the K8s/PVC Playwright workspace probe when a browser spec asserts backend-written files.

### 2.3 Windows Self-Hosted Operator Runtime

Use when the browser/operator session is on Windows and the runtime is Compose or self-hosted Kubernetes with an explicit non-loopback AI endpoint.

## 3. Required Preflight Record

Record branch, commit SHA, date/time, runtime lane, browser, UI URL, live-backend vs stable mode, and whether the run was headed.

## 3A. Browser-Visible Certification Rule

For operator-facing certification, run critical Chromium proof in headed mode and do not fan out multiple managed Playwright invocations against the same workspace/port.

Critical command families:
- `uv run inv interface.e2e --project=chromium --workers=1`
- focused headed proof with `uv run inv interface.e2e --headed --live-backend --server-mode=external --project=chromium --workers=1 --spec=<focused-live-spec>`
- guarded WSL release proof with `uv run inv wsl.validate --lane=release --headed-browser`

For the guarded WSL release lane, `uv run inv wsl.validate --lane=release --headed-browser` runs the focused live-backend Compose proof with visible Playwright windows after the standard release validation sequence.

## 4. Execution Order

1. environment preflight
2. Central Soma and navigation proof
3. AI Organization entry and creation proof
4. organization workspace bootstrap proof
5. direct-answer and content-delivery proof
6. governed mutation and artifact proof
7. ExecutionContract, ProofArtifact, and UI response-state proof
8. continuity and refresh proof
9. settings persistence proof
10. advanced-surface boundary proof
11. audit/activity visibility proof
12. disruption, degradation, retry, and partial-completion proof
13. deployed live-backend proof, including `System -> Deployments` when deployment trust changed

## 5. Workflow Test Set

Core workflows are listed in [Testing](../TESTING.md#full-gui-coverage-matrix). Keep this contract stable for manifest/navigation references.

## 6. Deployed Live-Backend Proof

Live proof details live in [Testing](../TESTING.md#user-interaction-delivery-gate) and [Remote User Testing](../REMOTE_USER_TESTING.md).

## 7. Reporting Format

Report:
- runtime lane and UI URL
- branch/SHA
- commands run
- pass/fail results
- runtime objects verified: ExecutionContract, ProofArtifact, CapabilityManifestState, UIResponseState
- recovery/degradation result
- screenshots or recording locations
- blockers
- docs reviewed/updated

## 8. Final Verdict Format

Use:

```text
Verdict: PASS | FAIL | BLOCKED
Runtime lane:
Browser:
Evidence:
Blockers:
Follow-up:
```

## 9. Non-Negotiable Rule

Do not claim product/browser readiness from headless, API-only, or mocked evidence when the slice changes the delivered operator workflow, runtime topology, governance path, retained outputs, or provider availability.
