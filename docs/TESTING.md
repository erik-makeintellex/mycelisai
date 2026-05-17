# Verification & Testing Protocol
> Navigation: [Project README](../README.md) | [Docs Home](README.md)
Mycelis uses a five-tier validation model: backend unit tests, frontend component tests, browser workflows, integration tests, and governance/system smoke tests.

## TOC

- [Current Validation Contract](#current-validation-contract)
- [Thorough Release Testing Contract](#thorough-release-testing-contract)
- [User Interaction Delivery Gate](#user-interaction-delivery-gate)
- [Finalization Concretization Gate](#finalization-concretization-gate)
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
- GitHub Actions are manual-only through `workflow_dispatch`; source-mode local gates are first, infra-only PostgreSQL/NATS may support live proof, and full Docker/Compose/K8s app proof starts only after local run/build/test evidence is acceptable.
- Windows is the edit/review/push surface; WSL is the guarded Compose release-style proof checkout, while Rancher Desktop K3s is the Windows local Kubernetes/commercial-parity proof lane.
- `ci.service-check --live-backend` ensures the `cortex` database exists and proves the managed built server path when service/browser proof is required; `interface.check` retries transient Windows socket-reuse failures after heavy browser proof before treating a route as failed.
- Playwright starts/stops the managed Next.js app, can use the built production Interface server path, and covers `mobile-chromium`, `@axe-core/playwright`, `workspace-live-backend.spec.ts`, and `--live-backend` paths where relevant.

## Thorough Release Testing Contract

Use this source-first sequence when a slice changes the delivered operator workflow, runtime topology, governance behavior, retained outputs, AI provider posture, or release proof lane:

1. Source and contract proof from the Windows repo first: `uv run inv core.test`, Interface gates, docs tests, `api.delivery-proof` when Core is live, and `uv run inv quality.max-lines --limit 300`; capped files in `ops/quality_legacy_caps.txt` must match current counts.
2. Keep PostgreSQL and NATS available as infra-only dependencies when the local source stack needs real persistence or bus proof; only after source proof is acceptable, containerize Core/Interface for deployment-mimic proof from the git-refreshed WSL checkout:
   - `uv run inv wsl.refresh`
   - `uv run inv wsl.validate --lane=release`
3. Local Kubernetes proof when Helm/commercial-release parity is part of the risk: set `$env:MYCELIS_K8S_BACKEND="rancher"` on Windows Rancher Desktop K3s, use `$env:MYCELIS_K8S_VALUES_FILE="charts/mycelis-core/values-k3d.yaml"` as the shared local-Kubernetes preset for Rancher K3s and k3d, then run `uv run inv k8s.deploy`, `uv run inv k8s.wait --timeout=300`, `uv run inv k8s.bridge`, and a live backend GUI proof with `PLAYWRIGHT_BACKEND_WORKSPACE_PROBE=k8s`.
4. Visible live-window Playwright proof when the user-facing browser path is part of the risk:
   - `uv run inv wsl.validate --lane=release --headed-browser`
   - or, when Compose is already up and should stay running, `uv run inv interface.e2e --headed --live-backend --server-mode=external --project=chromium --spec=<focused-live-spec>`
5. Broader headed Chromium MVP certification after the clean release lane:
   - run the critical matrix in [V8 UI Team Browser Workflows](architecture-library/V8_UI_TEAM_BROWSER_WORKFLOWS.md) sequentially

## User Interaction Delivery Gate
Do not claim thorough release readiness from unit, type, or headless-only proof when the slice changes what the operator sees or approves.

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
- created teams start with the minimum viable roster, normally one accountable lead, and any UI/API expansion path explains the missing capability, owned task, proof expected, and removal point
- retained outputs survive refresh/reload
- AI-host failure produces a visible blocker and recovery restores the lane

Use [Remote User Testing](REMOTE_USER_TESTING.md) for human walkthrough proof and [V8 UI Team Full Test Set](architecture-library/V8_UI_TEAM_FULL_TEST_SET.md) for the full browser matrix.

## Finalization Concretization Gate
Every finalization slice must prove the concrete runtime contract it touches, not only that screens render or APIs return `200`.

| Contract | Required proof | Status |
| --- | --- | --- |
| Canonical MVP workflow | headed/live proof for Soma request -> proposal when required -> approval -> run -> output -> proof/audit -> revisit | `ACTIVE` |
| ExecutionContract | API/unit proof for contract id, execution shape, governance posture, required capabilities, expected output/proof, recovery/degradation, run linkage, version/timestamps | `IN_REVIEW` for confirm-action |
| ProofArtifact | API/UI proof for proof id, run id, status, evidence/output/audit refs, validation source, proof quality, degradation state, recovery options, confidence provenance | `IN_REVIEW` for confirm-action |
| TeamWorkItem/TeamInteraction/TeamStatusEvent | Go/API proof that team creation remains non-active (`new`/`briefed`), delegated work queues real work, retained deliverables move through queued/running/output-ready or degraded states, and run/contract/proof/output refs persist where available | `IN_REVIEW` for confirm-action wiring |
| UI response states | component/browser proof for direct answer, proposal, execution result, blocker, recovery state, degraded execution, awaiting approval, retry required, partial completion | `REQUIRED` |
| TeamWorkItem UI state | component/browser proof that Soma Active Work Lane prefers durable `/api/v1/teams/{id}/work` rows, renders projection fallback only as degraded/inspectable, and feeds retained `TeamOutputRef` records into the Output Workbench | `IN_REVIEW` |
| CapabilityManifestState | durable Go/SQL/API proof for capability id, health, probe status, risk, approval posture, allowed roles, input/output schemas, failure/recovery posture, audit/secret policy, owner, updated time, and manifest version; `db.migrate` compatibility now requires the V8.2 capability/proof/team-work tables; UI proof remains the next Connected Tools/Resources gate | `IN_REVIEW` for persistence/API, `REQUIRED` for UI proof |
| Deployment trust | `System -> Deployments` proof for deployment/execution/workspace roots, current commit, endpoint posture, runtime posture, proof lane, recovery state | `IN_REVIEW`; backend/API and mocked browser proof are available, local-source live proof uses already-running Core/Interface with infra-only PostgreSQL/NATS and skips honestly when that lane is not available |
Related active contracts: [V8 UI/API and Operator Experience Contract](architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md), [V8 Directed Execution Delivery Plan](architecture-library/V8_DIRECTED_EXECUTION_DELIVERY_PLAN.md), [V8 UI Testing Agentry Product Contract](architecture-library/V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md), and [V8.2 Current State And Finalization PRD](architecture-library/V8_2_CURRENT_STATE_AND_FINALIZATION_PRD.md). Current GUI proof status: live Soma governance, team execution, and playable-output flows are green through `soma-governance-live.spec.ts`, `team-execution-live.spec.ts`, and `team-output-content-live.spec.ts`. The exact first demo slice now has focused browser specs: `ui-finalization-browser-package-live.spec.ts` proves package metadata, proof opening, reload, and Groups output against a live backend, while `ui-finalization-browser-package-retry.spec.ts` proves the degraded/retry UI path with mocked failure.
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
  - <direct answer|proposal|execution result|blocker|recovery state|degraded execution|awaiting approval|retry required|partial completion>
- Runtime contract:
  - <ExecutionContract/ProofArtifact/CapabilityManifestState fields touched>
- Output/proof/event shape:
  - <output refs, proof refs, audit refs, emitted events>
- Recovery/degradation expectation:
  - <what succeeded, what failed, what remains trusted, what requires retry/operator attention>
- Evidence commands:
  - <unit/component/type/build/focused browser/live-backend/deployment docs gates>
```

No backend/API review is complete without a mapped UI target and evidence result.
## Clean Run Discipline

- Stop prior local services before runtime or integration tests: `uv run inv lifecycle.down`.
- Do not keep full Docker/K8s app stacks running during ordinary source work; use local run/build/test with infra-only PostgreSQL/NATS when needed, then intentionally bring up Compose/K8s for deployment proof.
- For Compose data-plane proof, use `uv run inv compose.infra-up`, `compose.infra-health`, and `compose.storage-health`.
- Inspect service ports/processes before runtime proof when prior runs may have left residue.
- Use `uv run inv lifecycle.status` for the fast process/endpoint snapshot; it checks Core through `/healthz` and Ollama through `/api/tags` across loopback fallbacks so transient TCP-only snapshots do not mark reachable services down.
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
uv run inv api.delivery-proof
uv run inv logging.check-schema
uv run inv logging.check-topics
```

`wsl.validate --lane=release` runs `compose.health` before each live browser spec because those specs execute through separate WSL shell invocations. Use `--headed-browser` on that same task when acceptance evidence must include visible live Playwright windows.

Focused live-backend examples:
```bash
uv run inv interface.e2e --live-backend --server-mode=start --spec=e2e/specs/soma-governance-live.spec.ts
uv run inv interface.e2e --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/search-provenance-live.spec.ts
uv run inv interface.e2e --live-backend --server-mode=start --spec=e2e/specs/workspace-live-backend.spec.ts
uv run inv interface.e2e --headed --live-backend --server-mode=start --project=chromium --spec=e2e/specs/groups-live-backend.spec.ts
uv run inv interface.e2e --headed --live-backend --server-mode=external --project=chromium --spec=e2e/specs/team-output-content-live.spec.ts
uv run inv interface.e2e --headed --live-backend --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/ui-finalization-browser-package-live.spec.ts
```
Finalization proof order after integration:
1. Run the mocked browser harness serially with `--project=chromium --workers=1`: `first-demo-success.spec.ts`, `ui-finalization-browser-package-retry.spec.ts`, `desktop-mobile-compression.spec.ts`, `system-deployments.spec.ts`, and `active-work-api.spec.ts`.
2. Run source gates: `uv run inv interface.test`, `uv run inv interface.typecheck`, `uv run inv interface.build`, docs tests, and focused backend tests for changed APIs.
3. Run live first-demo proof after Core/Interface are integrated: `ui-finalization-browser-package-live.spec.ts` with headed live backend.
4. Run local-source deployment/active-work proof only after Core and Interface are already running from source against the needed infra-only PostgreSQL/NATS services; do not start full Core/Interface Docker app stacks for this lane. `system-deployments.spec.ts` uses the live endpoint when `--live-backend`, `PLAYWRIGHT_LIVE_BACKEND=1`, or `PLAYWRIGHT_SYSTEM_DEPLOYMENTS_LIVE=1` is set. `active-work-api.spec.ts` writes one durable proof work item and reads it back only when `PLAYWRIGHT_TEAM_WORK_API=1` or `PLAYWRIGHT_ACTIVE_WORK_API_LIVE=1` is set, and it fails migrated-table regressions while skipping when Core/auth/PostgreSQL are unavailable.
Use `npm test -- teamWorkProjection OutputWorkbench useDurableTeamWork TeamsPage --maxWorkers=1` when Soma Active Work Lane, durable team-work projection, or retained team-output workbench mapping changes.
Use `uv run inv interface.e2e --headed --server-mode=external --project=chromium --workers=1 --spec=e2e/specs/soma-media-artifacts.spec.ts` when chat media rendering, saved-media downloads, or local storage-folder reveal controls change. The focused browser proof should show image previews, playable media or downloadable binary artifacts, saved paths, and an operator-visible control that opens the mounted storage location for generated content.
Use `npm test -- WorkspaceExplorer MCPToolRegistry MCPLibraryBrowser MCPServerCard --maxWorkers=1` from `interface/` when Connected Tools, filesystem MCP, Workspace Files browsing, or MCP output generation contracts change. Use `npm test -- DeploymentContextPanel ResourcesPage --maxWorkers=1` when Resources -> Deployment Context intake compression changes; proof should verify the content/classification/scope tabs, bounded loaded-context list, and unchanged POST payloads.
Use `cd core; go test ./internal/server -run "Test(ParsePlannedToolCall|HandleConfirmAction|ExecutePlannedToolCalls)" -count=1` when governed proposal planning or confirm-action replay changes. The focused backend proof should verify explicit `tool_ref` MCP plans retain the `mcp:server/tool` identity, execute through the MCP executor after approval, and return retained `mcp_tool_result` outputs instead of silently falling back to same-named internal tools. Add `TeamWork` to the focused `-run` filter when the confirmed action path changes team creation, delegation, retained output refs, or durable active-work state.
Use `uv run inv interface.e2e --project=chromium --workers=1 --server-mode=external --spec=e2e/specs/soma-proposal-mode.spec.ts` when proposal approval, confirm-action failure handling, degradation metadata, or failed-run recovery UI changes. The focused browser proof must show the failed run remains reviewable, the Operator trust package says operator attention is needed, recovery copy is visible, and success proof labels are absent.
Rancher Desktop K3s live-backend proof uses the local Interface server against the K3s Core bridge; `k8s.bridge` forwards in-cluster Core `:8080` to `MYCELIS_API_PORT` or `8081` by default, and the Interface proxy/live proof must use that same local port. When the spec checks backend-written files, prove the PVC-backed workspace through `kubectl`; see Operations for the full K3s env block.
```bash
uv run inv k8s.deploy
uv run inv k8s.wait --timeout=300
uv run inv k8s.bridge
$env:MYCELIS_API_HOST="127.0.0.1"
$env:MYCELIS_API_PORT="8081"
uv run inv interface.e2e --headed --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts --live-backend --workers=1 --server-mode=dev
uv run inv interface.e2e --headed --project=chromium --spec=e2e/specs/team-execution-live.spec.ts --live-backend --workers=1 --server-mode=external
```
Use `team-execution-live.spec.ts` when team execution, retained file/code outputs, group visibility, run-conversation proof, or team NATS proposal subjects change. The native Core proof path now infers `MYCELIS_BACKEND_WORKSPACE_ROOT=core/workspace` from `MYCELIS_WORKSPACE=./workspace`; Compose or split-checkout proof can still set `MYCELIS_BACKEND_WORKSPACE_ROOT` or `PLAYWRIGHT_BACKEND_WORKSPACE_ROOT` explicitly, and PVC-backed K8s proof should use `PLAYWRIGHT_BACKEND_WORKSPACE_PROBE=k8s`. The proof asks Soma for explicit expected output criteria, creates a runtime team, approves governed execution, writes a small browser-game HTML file through the backend workspace, confirms retained `team` and `code` outputs with the exact workspace-viewer href, opens the generated game from the retained output link, verifies the page title and initial score, clicks it in a browser page, checks the final score, checks `/api/v1/runs/{id}/conversation`, and verifies the team appears in `/groups`. Use `team-output-content-live.spec.ts` when the question is whether teams are producing meaningful content as specified: it creates multiple teams through Soma, approves their work, opens each retained browser output and local folder control in headed Chromium, verifies deliverable content, play-tests a code-graphics game score, checks run-conversation tool proof, and confirms the teams are reviewable in Groups.

When complex generated projects such as playable games are packaged as `project_package` outputs, focused proof should also verify the Operator trust package and Groups retained-output pane show the package title, entrypoint, folder, file list, validation/proof text, `Open Game`, and workspace `Storage` control. Compose live proof defaults `MYCELIS_WORKSPACE_REVEAL_DRY_RUN=1`, so `Storage` proves the retained mounted path without launching a host file explorer from the Core container. Unit/component proof can stay mocked with `npm test -- MissionControlChat.executionSummary GroupManagementPanel --maxWorkers=1`; backend proof should include confirmed `write_file` and `store_artifact` artifact-envelope paths with `go test ./internal/server ./internal/swarm ./internal/artifacts ./pkg/protocol -run "Test(BuildDirectChatExecutionSummary|ExtractToolOutputArtifacts|ArtifactType|ExecutionOutputsFromToolResults|ArtifactResultPayload|ExecutionOutputsFromArtifacts|ChatResponsePayload)" -count=1`; live proof should use the headed `team-output-content-live.spec.ts` or `team-execution-live.spec.ts` path to open and play the retained browser output.
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

Runtime-team proof should stay bounded. Use one named team for normal output/proof checks and at most three targeted teams when the test is explicitly about coordination. Do not use broad all-team broadcast as a default proof path; reserve it for broadcast/degradation tests. Clean up temporary runtime teams with:

```bash
curl -X DELETE http://127.0.0.1:8081/api/v1/teams/<team-id> -H "Authorization: Bearer $MYCELIS_API_KEY"
```

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
