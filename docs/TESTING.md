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
- [MYC-RC Package Proof](#myc-rc-package-proof)
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
- Windows uses the **Two-Stage Windows Execution + Windows Rancher Desktop (`rancher-desktop`) Pipeline**: **Stage 1 (Windows Local Execution & Testing)** runs Go, Python, Vitest, and Playwright checks directly from the Windows workspace without launching WSL; **Stage 2 (Windows Rancher Desktop Containerization & Deployment)** uses Windows Rancher Desktop (`rancher-desktop` context at `https://127.0.0.1:6443`, `LoadBalancer` on `localhost:8080`, `local-path` storage, and `http://host.rancher-desktop.internal:11434/v1` AI bridge) to manage containerization and Helm deployment after local execution tests pass. Windows is the edit/review/push surface; WSL is the authoritative release-style proof checkout when optional WSL lanes are invoked.
- When WSL/Docker instability blocks that lane, use the Windows repo with managed `uv run inv interface.e2e ...` full-browser Chromium sessions as the active local product-proof workaround, and record WSL as an environment blocker rather than an application failure.
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
- **Multi-Form Soma Interaction & Anti-"Can't Deliver" Matrix:** Every Soma interaction change must pass the 3-layer interaction matrix (`core/internal/server/soma_interaction_matrix_test.go` covering 29 casual build, game, doc/script/checklist, team/multi-app/MCP, and poisoned-history prompts; `core/internal/swarm/agent_soma_interaction_test.go` covering `[REFERENTIAL REVIEW]` system promotion, direct-route tool-loop guards, and single-pass refusal recovery; and `interface/__tests__/dashboard/SomaInteractionMatrix.test.tsx` covering frontend answer, proposal, `+ Queue` auto-drain, and cold-start retry recovery).

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
- `/deployments` (standing/mission teams, groups, planned departments; `/teams` redirects here) and `/teams/create`
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
uv run inv compose.storage-health
uv run inv service.status --target=compose
uv run inv lifecycle.health
uv run inv quality.max-lines --limit 300
uv run inv logging.check-schema
uv run inv logging.check-topics
uv run inv release.status
uv run inv release.stage --tag=v0.8.3 --dry-run
uv run inv release.prune-branches --dry-run
```

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

## MYC-RC Package Proof

For `MYC-RC-000` through `MYC-RC-003`, testing must prove the platform change, not only route health.

- `MYC-RC-000`: docs/state/index agree on convergence authority, no stale nonexistent V8.3 links remain, and gates use `uv run inv ...`.
- `MYC-RC-001`: a Codex task package is produced only after proposal/confirm, creates mission run/events, writes retained artifacts, and exposes proof.
- `MYC-RC-002` / `MYC-RC-002B`: output refs normalize file/team/MCP/template/package outputs with open/reveal/proof/source lineage, and the UI presents one package card with Inspect details.
- `MYC-RC-003`: write/output proof records checksum or fingerprint, path-boundary verification, execution status, output ref, artifact linkage, and proof readback. Use [MYC-RC-003 Proof Envelope Delivery Plan](architecture-library/MYC_RC_003_PROOF_ENVELOPE_DELIVERY_PLAN.md) for the active owner map and Windows-managed proof lane.

Current focused `MYC-RC-001` / `MYC-RC-002` package-output proof:
- `go test ./internal/server`
- `npm test -- --run __tests__/dashboard/MissionControlChat.package-output.test.tsx __tests__/runs/RunTimeline.test.tsx __tests__/store/useCortexStore.confirm-proposal.execution.test.ts __tests__/store/useCortexStore.mission-chat-send.test.ts`
- `npx tsc --noEmit`
- `uv run inv interface.e2e --project=chromium --spec=e2e/specs/codex-package-retention.spec.ts`
- `uv run inv interface.e2e --project=chromium --workers=1`

`MYC-RC-002` adds explicit assertions that confirmed packages and Soma package envelopes carry `output_refs`, `delivery_package`, and `execution_summary`, including primary/proof roles and source run/contract lineage. Keep compatibility assertions for `artifacts` and legacy `package` fields until all consumers move to `OutputRef`.

Browser proof must include refresh/re-entry so retained package visibility is not a transient chat-only success. Manual acceptance is useful for exploration, but RC lock requires an automated focused browser spec or an explicit `BLOCKED` note naming the missing spec.

Current `MYC-RC-003A` proof adds backend assertions for package-file checksum, byte count, path-boundary verification, artifact identity, source run/contract, proof artifact linkage, and event payload proof metadata; store/component assertions for proof fields in confirm/chat normalization and Output Workbench; and Windows-managed Chromium as the local full-browser gate while WSL/Docker is environment-blocked.

## Tier 1: Backend Unit Tests

Run `uv run inv core.test`. Use focused Go package tests during implementation, then rerun the managed task before close-out.

## Tier 2: Frontend Unit Tests

Run `uv run inv interface.test` and `uv run inv interface.typecheck`. The Interface Vitest gate is intentionally sequential for deterministic full-page jsdom workflows.

## Tier 3: End-to-End Tests

Run `uv run inv interface.e2e`. Use `--server-mode=start` for production-start proof and `--live-backend` when the spec must hit a real Core backend. Invoke manages server lifecycle, browser cache, serial workers, and cleanup. `--spec` also accepts one quoted, space-separated list of spec paths to run several files in one invocation, e.g. `--spec="e2e/specs/a.spec.ts e2e/specs/b.spec.ts"`.
**Chromium-only policy for deep MVP specs:** mocked, no-live-backend specs proving MVP workflows (proposal/approve, team creation, workflow output retention/reload, error/recovery, codex package retention, groups, docs-and-runs, deployments, organization starters, config clone/apply) run Chromium only in CI; local ad hoc `--project=firefox`/`--project=webkit` runs remain available when a change plausibly affects another engine. GitHub CI runs this token-free set (no live backend, no Ollama) as the `mvp-browser` job, separate from `browser-smoke` (still scoped to `e2e/specs/homepage.spec.ts`):
```bash
uv run inv interface.e2e --project=chromium --workers=1 --server-mode=start --spec="e2e/specs/workflow-output.direct.spec.ts e2e/specs/proposals.spec.ts e2e/specs/team-creation.spec.ts e2e/specs/workflow-output.reload-review.spec.ts e2e/specs/error-scenarios.spec.ts e2e/specs/codex-package-retention.spec.ts e2e/specs/groups.spec.ts e2e/specs/docs-and-runs.spec.ts e2e/specs/deployments.spec.ts e2e/specs/organization-starters.spec.ts e2e/specs/config-apply.spec.ts"
```
**Windows e2e harness rules:** never run two `uv run inv interface.e2e` invocations at once (they share port 3100 and the webserver log, producing unrelated "worker process exited unexpectedly" cascades), and never run Vitest during an e2e run (its Interface process cleanup can kill the Vitest process mid-suite); specs that assert state after `page.reload()` must wait for the relevant network response with a bounded retry, because the throttled managed dev server can occasionally stall one reload round-trip (see `settings.spec.ts`).

`catalogue.spec.ts`, `wiring-active-delete.spec.ts`, and `wiring-edit.spec.ts` cover real MVP surfaces (Role Library and Automations "Workflow Builder" behind Advanced Mode) by seeding `mycelis-advanced-mode=true` via `page.addInitScript` and asserting the current `New Agent` and `Workflow Builder` controls directly.

**Visible Windows Chrome E2E Grading:** `interface/playwright.config.ts` supports `PLAYWRIGHT_CHANNEL=chrome` and `PLAYWRIGHT_SLOW_MO` for visible Windows Chrome E2E grading across `user-deliverable-journeys.spec.ts` (deliverable starters + sandboxed `<iframe sandbox="allow-scripts">` **`Play / Live Preview`**), `mvp-settings-and-modals.spec.ts`, and `v8-operator-first-ux.spec.ts`.

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

Run `uv run inv core.smoke`. Also run `logging.check-schema` and `logging.check-topics` when event, log, or NATS subject behavior changes.

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

Release-preflight lane presets are `baseline`, `runtime`, `service`, and `release`; use `uv run inv ci.release-preflight --lane=release` for the full release gate. Before promoting `dev` onto `main`, verify branch/tag/version readiness with `uv run inv release.status`, dry-run staging gates (`quality.max-lines --limit 300` + `git diff --check`) via `uv run inv release.stage --tag=v0.8.3 --dry-run`, and promote/tag/prune with `uv run inv release.stage --tag=v0.8.3` and `uv run inv release.prune-branches`.

Deployment proof contracts:
- Windows Docker Desktop Compose with the Windows browser on the same machine for rapid local proof.
- Kubernetes / Helm clustered deployment reached through the real ingress, remote host, IP, or hostname; the browser opens the UI through the same operator-facing address the delivered environment will actually use.
- when the validation target is clustered deployment, run `uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml`; Compose remains rapid local development/proof only.
- clustered Kubernetes proof: `uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml` is the static chart gate that must pass before any live cluster check.
- when the validation target is local Kubernetes, prefer `k3d`; fallback with `MYCELIS_K8S_BACKEND=kind`; on Windows Stage 2, `k8s.*` auto-detects Windows Rancher Desktop (`rancher-desktop` context at `https://127.0.0.1:6443`, `LoadBalancer` on `localhost:8080`, and `local-path` storage).
- use `MYCELIS_K8S_TEXT_ENDPOINT` and optional `MYCELIS_K8S_MEDIA_ENDPOINT`, an explicit reachable AI host instead of a chart-baked or localhost default, and keep it container-reachable instead of `localhost`, `127.0.0.1`, or `0.0.0.0`; Compose may relay `MYCELIS_COMPOSE_OLLAMA_HOST` through the WSL host, and Windows Rancher Desktop uses `http://host.rancher-desktop.internal:11434/v1`.
- `MYCELIS_K8S_VALUES_FILE` may select `charts/mycelis-core/values-k3d.yaml`, `charts/mycelis-core/values-enterprise.yaml`, or `charts/mycelis-core/values-enterprise-windows-ai.yaml` (or `charts/mycelis-core/values-rancher-desktop.yaml` for Windows Rancher Desktop).
- prove the deployed Core image can launch the curated `filesystem` stdio MCP server through `npx` and runtime workspace normalization for filesystem installs.

Guarded WSL tasks: `uv run inv wsl.status`, `uv run inv wsl.refresh`, `uv run inv wsl.validate`, `uv run inv wsl.cycle`. Clean Docker build context rules: `.dockerignore` excludes untracked `.state/*` dumps while keeping tracked `.state/V8_DEV_STATE.md` and `.state/V7_DEV_STATE.md`.

Release-proof sequencing rule: validate WSL git auth repair/report behavior for `wsl.refresh`; run `uv run inv wsl.validate` from the refreshed WSL proof checkout before trusting browser-gap or certification evidence; that task intentionally runs `ci.release-preflight --lane=runtime --no-e2e` first; keep the newly closed focused browser proof gaps green: `/runs` workflow depth and guided Soma retry/recovery both have focused Chromium proof in production `start` mode; rerun the broader headed Chromium certification pass only after the focused proof-hardening slice is committed and refreshed into WSL.

## Adding New Tests

Add tests where the risk lives:
- backend handler or service tests for API/runtime changes
- store/component tests for UI state and rendering contracts
- Playwright for operator-visible workflows
- live-backend Playwright for proxy/Core/runtime contracts
- docs-link and max-line gates for docs structure changes

Use status markers from the repo standard and keep evidence commands in the close-out.
