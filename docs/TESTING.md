# Verification & Testing Protocol
> Navigation: [Project README](../README.md) | [Docs Home](README.md)

Mycelis uses a five-tier validation model: backend unit tests, frontend component tests, browser workflows, integration tests, and governance/system smoke tests.

## TOC

- [Current Validation Contract](#current-validation-contract)
- [Thorough Release Testing Contract](#thorough-release-testing-contract)
- [User Interaction Delivery Gate](#user-interaction-delivery-gate)
- [Target-Action Testing Matrix](#target-action-testing-matrix)
- [Full GUI Coverage Matrix](#full-gui-coverage-matrix)
- [Backend/API -> UI Target Plan](#backendapi---ui-target-plan)
- [Clean Run Discipline](#clean-run-discipline)
- [Quick Reference](#quick-reference)
- [Product Delivery Proof](#product-delivery-proof)
- [Tier 1: Backend Unit Tests](#tier-1-backend-unit-tests)
- [Tier 2: Frontend Unit Tests](#tier-2-frontend-unit-tests)
- [Tier 3: End-to-End Tests](#tier-3-end-to-end-tests)
- [Tier 4: Integration Tests](#tier-4-integration-tests)
- [Tier 5: Governance Smoke Tests](#tier-5-governance-smoke-tests)
- [Memory Restart Validation](#memory-restart-validation)
- [CI Pipelines](#ci-pipelines)
- [Adding New Tests](#adding-new-tests)

## Current Validation Contract

- Feature work is not done until relevant tests run against the final branch state.
- feature work is also not done until the touched docs are reviewed and updated where meaning changed
- end-of-slice reporting should name both the evidence commands run and the docs updated or reviewed unchanged for the touched scope
- Use `uv run inv ...` for real task execution.
- `uv run inv ci.baseline` is the default branch-readiness gate.
- Use `uv run inv ci.baseline --no-e2e` only for intentionally narrower debugging.
- GitHub CI proves codebase health without hosted agentry; live AI/service proof remains local, WSL, Compose, or target-cluster evidence.
- Windows is the edit/review/push surface; WSL is the guarded Compose release-style proof checkout, while Rancher Desktop K3s is the Windows local Kubernetes/commercial-parity proof lane.
- `ci.service-check --live-backend` ensures the `cortex` database exists and proves the managed built server path when service/browser proof is required.
- Playwright starts/stops the managed Next.js app, can use the built production Interface server path, and covers `mobile-chromium`, `@axe-core/playwright`, `workspace-live-backend.spec.ts`, and `--live-backend` paths where relevant.

## Thorough Release Testing Contract

Use this sequence when a slice changes the delivered operator workflow, runtime topology, governance behavior, retained outputs, AI provider posture, or release proof lane:

1. Source and contract proof from the Windows repo: `uv run inv core.test`, `uv run inv interface.test`, `uv run inv interface.typecheck`, `uv run inv interface.build`, docs tests, and `uv run inv quality.max-lines --limit 300`.
2. Deployment-mimic proof from the git-refreshed WSL checkout:
   - `uv run inv wsl.refresh`
   - `uv run inv wsl.validate --lane=release`
3. Local Kubernetes proof when Helm/commercial-release parity is part of the risk: set `$env:MYCELIS_K8S_BACKEND="rancher"` on Windows Rancher Desktop K3s, use `$env:MYCELIS_K8S_VALUES_FILE="charts/mycelis-core/values-k3d.yaml"` as the shared local-Kubernetes preset for Rancher K3s and k3d, then run `uv run inv k8s.deploy`, `uv run inv k8s.wait --timeout=300`, `uv run inv k8s.bridge`, and a live backend GUI proof with `PLAYWRIGHT_BACKEND_WORKSPACE_PROBE=k8s`.
4. Visible live-window Playwright proof when the user-facing browser path is part of the risk:
   - `uv run inv wsl.validate --lane=release --headed-browser`
   - or, when Compose is already up and should stay running, `uv run inv interface.e2e --headed --live-backend --server-mode=external --project=chromium --spec=<focused-live-spec>`
5. Broader headed Chromium MVP certification after the clean release lane:
   - run the critical matrix in [V8 UI Team Browser Workflows](architecture-library/V8_UI_TEAM_BROWSER_WORKFLOWS.md) sequentially

Do not claim thorough release readiness from unit, type, or headless-only proof when the slice changes what the operator sees or approves.

## User Interaction Delivery Gate

Supported user proof lanes:
- Windows Docker-compatible Compose through Rancher Desktop or Windows Docker Desktop Compose with Windows browser at `http://localhost:3000`
- Windows Rancher Desktop K3s with a local Interface server and Windows browser
- WSL-hosted Compose with Windows browser at `http://localhost:3000`
- second-machine or Linux-server proof through a real host/IP/hostname
- self-hosted Kubernetes/Helm through real ingress

Every accepted user-interaction proof must verify:
- the browser uses the same operator-facing address the delivered environment will use
- Soma returns a direct `answer` for a non-mutating prompt
- a mutating prompt enters `proposal` and can be approved or cancelled
- guided team creation or a temporary workflow lane completes and remains reviewable
- retained outputs survive refresh/reload
- AI-host failure produces a visible blocker and recovery restores the lane

Use [Remote User Testing](REMOTE_USER_TESTING.md) for human walkthrough proof and [V8 UI Team Full Test Set](architecture-library/V8_UI_TEAM_FULL_TEST_SET.md) for the full browser matrix.

## Target-Action Testing Matrix

| Target action | Required proof | Status |
| --- | --- | --- |
| Soma-first Team Expression and module binding | component/store coverage, adapter integration, product-flow proof | `ACTIVE` |
| Created-team workspace and channel inspector | selected-team inspector, command/status/result integration, browser proof | `ACTIVE` |
| Scheduler recurring execution | schedule CRUD/tick tests, restart persistence, recurring-state UI tests | `NEXT` |
| Causal chain operator UI | `/runs/[id]/chain` page/API/error coverage and focused browser lineage proof | `IN_REVIEW` |

Related migration inputs:
- [Intent To Manifestation And Team Interaction V7](architecture-library/INTENT_TO_MANIFESTATION_AND_TEAM_INTERACTION_V7.md)
- [Team Execution And Global State Protocol V7](architecture-library/TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md)
- [UI Target And Transaction Contract V7](architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md)

## Full GUI Coverage Matrix

Current browser workflow details live in [V8 UI Team Browser Workflows](architecture-library/V8_UI_TEAM_BROWSER_WORKFLOWS.md). Keep this document as the validation policy entrypoint rather than a route-by-route duplicate.

Minimum route families under active proof:
- `/dashboard` and AI Organization re-entry
- `/organizations/[id]` Soma-primary workspace
- `/groups` temporary and standing collaboration
- `/teams` and `/teams/create`
- `/resources` and Connected Tools
- `/memory`, `/system`, `/settings`
- `/runs`, `/runs/[id]`, `/runs/[id]/chain`
- legacy redirect routes

## Backend/API -> UI Target Plan

When backend/API behavior changes, attach this block before review:

```md
Backend/API -> UI Target Plan
- Backend/API change:
  - <route/payload/runtime contract delta>
- UI surfaces impacted:
  - <page/component/store paths>
- Expected terminal state(s):
  - <answer|proposal|execution_result|blocker>
- Failure/recovery expectation:
  - <timeout/degraded/rejection behavior shown to operator>
- Evidence commands:
  - uv run inv core.test
  - uv run inv interface.test
  - uv run inv interface.typecheck
  - uv run inv interface.build
  - <focused playwright command(s)>
  - <live-backend playwright command when proxy/core contract changed>
```

No backend/API review is complete without a mapped UI target and evidence result.

## Clean Run Discipline

- Stop prior local services before runtime or integration tests: `uv run inv lifecycle.down`.
- For Compose rebuild proof, use `uv run inv compose.down --volumes`.
- For Compose data-plane proof, use `uv run inv compose.infra-up`, `compose.infra-health`, and `compose.storage-health`.
- Inspect service ports/processes before runtime proof when prior runs may have left residue.
- Treat repo-local Interface workers as cleanup targets on Windows.
- Run one managed Playwright invocation at a time for a workspace and port.
- Start only the services required for the check.
- Shut services down after the check unless a follow-on validation needs them alive.
- Windows task loads prepend standard local tool bins, including Rancher Desktop and Chocolatey, and embedded NATS unit tests must bind package-unique low loopback ports outside the Windows dynamic client-port range instead of `Port: -1`.

Compiled Go services started with `go build`, `go run`, or direct binaries can outlive tests. Runtime validation must account for repo-local binaries, containerized dependencies, and test-managed services.

## Quick Reference

```bash
uv run inv core.test
uv run inv interface.test
uv run inv interface.typecheck
uv run inv interface.build
uv run inv interface.e2e
uv run inv core.smoke
uv run inv ci.test
uv run inv ci.baseline
uv run inv ci.release-preflight --lane=release
uv run inv wsl.validate --lane=release --headed-browser
uv run inv compose.health
uv run inv compose.warm-cognitive
uv run inv compose.storage-health
uv run inv lifecycle.health
uv run inv quality.max-lines --limit 300
uv run inv logging.check-schema
uv run inv logging.check-topics
```

`wsl.validate --lane=release` runs `compose.health` before each live browser spec because those specs execute through separate WSL shell invocations. Use `--headed-browser` on that same task when acceptance evidence must include visible live Playwright windows.

Focused live-backend examples:

```bash
uv run inv interface.e2e --live-backend --server-mode=start --spec=e2e/specs/soma-governance-live.spec.ts
uv run inv interface.e2e --live-backend --server-mode=start --spec=e2e/specs/workspace-live-backend.spec.ts
uv run inv interface.e2e --headed --live-backend --server-mode=start --project=chromium --spec=e2e/specs/groups-live-backend.spec.ts
uv run inv interface.e2e --headed --live-backend --server-mode=external --project=chromium --spec=e2e/specs/team-execution-live.spec.ts
```

Rancher Desktop K3s live-backend proof uses the local Interface server against the K3s Core bridge; `k8s.bridge` forwards in-cluster Core `:8080` to `MYCELIS_API_PORT` or `8081` by default, and the Interface proxy/live proof must use that same local port. When the spec checks backend-written files, prove the PVC-backed workspace through `kubectl`:

```bash
$env:MYCELIS_K8S_BACKEND="rancher"
$env:MYCELIS_K8S_VALUES_FILE="charts/mycelis-core/values-k3d.yaml"
$env:MYCELIS_K8S_TEXT_ENDPOINT="http://<windows-ai-host>:11434/v1"
$env:MYCELIS_K8S_TEXT_MODEL_ID="qwen3:8b"
$env:MYCELIS_K8S_SEARCH_PROVIDER="searxng"; $env:MYCELIS_K8S_SEARXNG_ENDPOINT="http://<windows-ai-host>:8088"
uv run inv k8s.deploy
uv run inv k8s.wait --timeout=300
uv run inv k8s.bridge
$env:MYCELIS_API_HOST="127.0.0.1"
$env:MYCELIS_API_PORT="8081"
$env:PLAYWRIGHT_BACKEND_WORKSPACE_PROBE="k8s"
$env:PLAYWRIGHT_K8S_NAMESPACE="mycelis"
$env:PLAYWRIGHT_K8S_CORE_SELECTOR="app=mycelis-core"
$env:PLAYWRIGHT_K8S_BACKEND_WORKSPACE_ROOT="/data/workspace"
uv run inv interface.e2e --headed --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts --live-backend --workers=1 --server-mode=dev
uv run inv interface.e2e --headed --project=chromium --spec=e2e/specs/team-execution-live.spec.ts --live-backend --workers=1 --server-mode=external
```

Use `team-execution-live.spec.ts` when team execution, retained file/code outputs, group visibility, run-conversation proof, or team NATS proposal subjects change. The proof asks Soma for explicit expected output criteria, creates a runtime team, approves governed execution, writes a small browser-game HTML file through the backend workspace, verifies the PVC-backed artifact with `PLAYWRIGHT_BACKEND_WORKSPACE_PROBE=k8s`, confirms retained `team` and `code` outputs with the exact workspace-viewer href, opens the generated game from the retained output link, verifies the page title and initial score, clicks it in a browser page, checks the final score, checks `/api/v1/runs/{id}/conversation`, and verifies the team appears in `/groups`.

## Product Delivery Proof

For product-facing work, include relevant unit/component tests, focused browser proof for the user workflow, live-backend proof when Core/proxy/runtime contracts changed, docs review, and explicit pass/fail evidence in close-out.

For output block, media readiness, and team-managed review, use:
- `uv run inv compose.up --build --wait-timeout=240`
- `uv run inv compose.health`
- `uv run inv interface.e2e --headed --project=chromium --spec=e2e/specs/v8-ui-testing-agentry.spec.ts`
- `uv run inv interface.e2e --headed --project=chromium --spec=e2e/specs/team-creation.spec.ts`

If the media engine is offline, record a blocker instead of treating missing media as passed.

## Tier 1: Backend Unit Tests

Run:

```bash
uv run inv core.test
```

Use focused Go package tests during implementation, then rerun the managed task before close-out.

## Tier 2: Frontend Unit Tests

Run:

```bash
uv run inv interface.test
uv run inv interface.typecheck
```

The Interface Vitest gate is intentionally sequential for deterministic full-page jsdom workflows.

## Tier 3: End-to-End Tests

Run:

```bash
uv run inv interface.e2e
```

Use `--server-mode=start` for production-start proof and `--live-backend` when the spec must hit a real Core backend. Invoke manages server lifecycle, browser cache, serial workers, and cleanup.

## Tier 4: Integration Tests

Use integration proof only when the slice depends on live external services, runtime orchestration, persistent storage, or provider behavior. Start the minimal stack, prove readiness, run the check, and shut down.

Common gates:

```bash
uv run inv lifecycle.health
uv run inv compose.health
uv run inv compose.storage-health
uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml
```

## Tier 5: Governance Smoke Tests

Run:

```bash
uv run inv core.smoke
```

Also run `logging.check-schema` and `logging.check-topics` when event, log, or NATS subject behavior changes.

## Memory Restart Validation

Run when memory, continuity, retained outputs, semantic storage, or restart behavior changes:

```bash
uv run inv lifecycle.memory-restart --frontend
```

For Compose long-term storage, pair migrations with:

```bash
uv run inv compose.storage-health
```

## CI Pipelines

Primary local gates:
- `uv run inv ci.test`
- `uv run inv ci.baseline`
- `uv run inv ci.service-check`
- `uv run inv ci.release-preflight --lane=release`

WSL release-style proof:
```bash
uv run inv wsl.refresh
uv run inv wsl.validate --lane=release
```

The managed install path uses `npm ci` for Interface dependencies, so WSL and CI-style proof
checkouts must not dirty `interface/package-lock.json` during bootstrap.

Release-preflight lane presets are `baseline`, `runtime`, `service`, and `release`; use `uv run inv ci.release-preflight --lane=release` for the full release gate.

Deployment proof contracts:
- Windows Docker-compatible Compose through Rancher Desktop or Windows Docker Desktop Compose with the Windows browser on the same machine for rapid local proof.
- Windows Rancher Desktop K3s with `MYCELIS_K8S_BACKEND=rancher` for local commercial-release parity proof.
- Kubernetes / Helm clustered deployment reached through the real ingress, remote host, IP, or hostname; the browser opens the UI through the same operator-facing address the delivered environment will actually use.
- when the validation target is clustered deployment, run `uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml`; Compose remains rapid local development/proof only.
- when the validation target is local Kubernetes, prefer `k3d` on WSL/Linux, prefer Rancher Desktop K3s on Windows, and fallback with `MYCELIS_K8S_BACKEND=kind`.
- use `MYCELIS_K8S_TEXT_ENDPOINT` and optional `MYCELIS_K8S_MEDIA_ENDPOINT`; optional `MYCELIS_K8S_TEXT_MODEL_ID` selects the installed model. Use an explicit reachable AI host instead of a chart-baked or localhost default, and keep it container-reachable instead of `localhost`, `127.0.0.1`, or `0.0.0.0`; Compose may relay `MYCELIS_COMPOSE_OLLAMA_HOST` through the WSL host.
- `MYCELIS_K8S_VALUES_FILE` may select `charts/mycelis-core/values-k3d.yaml`, `charts/mycelis-core/values-enterprise.yaml`, or `charts/mycelis-core/values-enterprise-windows-ai.yaml`; `values-k3d.yaml` is the shared local-Kubernetes preset for Rancher Desktop K3s on Windows and k3d on WSL/Linux.
- prove the deployed Core image can launch the curated `filesystem` stdio MCP server through `npx` and runtime workspace normalization for filesystem installs.

Guarded WSL tasks: `uv run inv wsl.status`, `uv run inv wsl.refresh`, `uv run inv wsl.validate`, `uv run inv wsl.cycle`.

Release-proof sequencing rule: validate WSL git auth repair/report behavior for `wsl.refresh`; run `uv run inv wsl.validate` from the refreshed WSL proof checkout before trusting browser-gap or certification evidence; that task intentionally runs `ci.release-preflight --lane=runtime --no-e2e` first, then Compose health/storage and `compose.warm-cognitive` before live browser proof; keep the newly closed focused browser proof gaps green: `/runs` workflow depth and guided Soma retry/recovery both have focused Chromium proof in production `start` mode; rerun the broader headed Chromium certification pass only after the focused proof-hardening slice is committed and refreshed into WSL.

## Adding New Tests

Add tests where the risk lives:
- backend handler or service tests for API/runtime changes
- store/component tests for UI state and rendering contracts
- Playwright for operator-visible workflows
- live-backend Playwright for proxy/Core/runtime contracts
- docs-link and max-line gates for docs structure changes

Use status markers from the repo standard and keep evidence commands in the close-out.
