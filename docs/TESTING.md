# Verification & Testing Protocol
> Navigation: [Project README](../README.md) | [Docs Home](README.md)

Mycelis uses a five-tier validation model: backend unit tests, frontend component tests, browser workflows, integration tests, and governance/system smoke tests.

## TOC

- [Current Validation Contract](#current-validation-contract)
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
- Windows is the edit/review/push surface; WSL is the authoritative release-style proof checkout.
- `ci.service-check --live-backend` ensures the `cortex` database exists and proves the managed built server path when service/browser proof is required.
- Playwright starts/stops the managed Next.js app, can use the built production Interface server path, and covers `mobile-chromium`, `@axe-core/playwright`, `workspace-live-backend.spec.ts`, and `--live-backend` paths where relevant.

## User Interaction Delivery Gate

Supported user proof lanes:
- Windows Docker Desktop Compose with Windows browser at `http://localhost:3000`
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
uv run inv compose.health
uv run inv compose.warm-cognitive
uv run inv compose.storage-health
uv run inv lifecycle.health
uv run inv quality.max-lines --limit 300
uv run inv logging.check-schema
uv run inv logging.check-topics
```

`wsl.validate --lane=release` runs `compose.health` before each live browser spec because those specs execute through separate WSL shell invocations.

Focused live-backend examples:

```bash
uv run inv interface.e2e --live-backend --server-mode=start --spec=e2e/specs/soma-governance-live.spec.ts
uv run inv interface.e2e --live-backend --server-mode=start --spec=e2e/specs/workspace-live-backend.spec.ts
uv run inv interface.e2e --headed --live-backend --server-mode=start --project=chromium --spec=e2e/specs/groups-live-backend.spec.ts
```

## Product Delivery Proof

For product-facing work, include:
- relevant unit/component tests
- focused browser proof for the user workflow
- live-backend proof when Core/proxy/runtime contracts changed
- docs review
- explicit pass/fail evidence in close-out

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
- Windows Docker Desktop Compose with the Windows browser on the same machine for rapid local proof.
- Kubernetes / Helm clustered deployment reached through the real ingress, remote host, IP, or hostname; the browser opens the UI through the same operator-facing address the delivered environment will actually use.
- when the validation target is clustered deployment, run `uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml`; Compose remains rapid local development/proof only.
- clustered Kubernetes proof: `uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml`.
- when the validation target is local Kubernetes, prefer `k3d`; fallback with `MYCELIS_K8S_BACKEND=kind`.
- use `MYCELIS_K8S_TEXT_ENDPOINT` and optional `MYCELIS_K8S_MEDIA_ENDPOINT`, an explicit reachable AI host instead of a chart-baked or localhost default, and keep it container-reachable instead of `localhost`, `127.0.0.1`, or `0.0.0.0`; Compose may relay `MYCELIS_COMPOSE_OLLAMA_HOST` through the WSL host.
- `MYCELIS_K8S_VALUES_FILE` may select `charts/mycelis-core/values-k3d.yaml`, `charts/mycelis-core/values-enterprise.yaml`, or `charts/mycelis-core/values-enterprise-windows-ai.yaml`.
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
