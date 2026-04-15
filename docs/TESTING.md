# Verification & Testing Protocol
> Navigation: [Project README](../README.md) | [Docs Home](README.md)

Mycelis employs a **5-Tier Testing Strategy** covering backend handlers, frontend components, end-to-end flows, integration tests, and governance smoke tests.

Current validation contract:
- feature work is not done until the relevant tests are rerun against the final branch state
- `uv run inv ci.baseline` is the default branch-readiness gate and now includes Playwright by default
- use `uv run inv ci.baseline --no-e2e` only for intentionally narrower local debugging
- use `uv run inv ci.service-check` to verify the currently running local stack through lifecycle health; the live-backend variant restores the local bridge/core stack, ensures the `cortex` database exists, reuses the `cortex` schema only when it is already compatible with the current runtime, and otherwise bootstraps the database before proving the browser contract against the managed built server / built production server path
- when the validation target is the supported home-runtime stack, use `uv run inv compose.up`, `uv run inv compose.status`, and `uv run inv compose.health` before browser proof instead of assuming Kind/bridge is the only real local environment; compose bring-up now mirrors `db.migrate` by skipping forward replay when the compose `cortex` schema is already compatible with the current runtime, emits numbered stage output with operator expectations, and accepts `--wait-timeout=<seconds>` for slower rebuild hosts
- when the validation target is a personal-owner Compose deployment with user-configured data services, use `uv run inv compose.infra-up --wait-timeout=180` first to start only PostgreSQL/NATS, verify the printed DB/NATS connection settings, run `uv run inv compose.infra-health`, then run `uv run inv compose.storage-health` after migrations so the long-term pgvector/Postgres memory and artifact store is proven before Core/Interface workflow proof
- in the supported home-runtime stack, `.env.compose` must keep container-host assumptions separate from `.env`; use `MYCELIS_COMPOSE_OLLAMA_HOST` there, and keep it container-reachable instead of `localhost`, `127.0.0.1`, or `0.0.0.0`
- on Windows hosts without a native `docker` binary, the compose task layer may execute Docker through WSL instead; keep the Docker daemon in a WSL distro that can reach the repo filesystem, and set `MYCELIS_WSL_DISTRO` when the default distro is not the correct Docker host
- on that Windows + WSL Docker path, the compose task layer may relay `MYCELIS_COMPOSE_OLLAMA_HOST` through the WSL host so the Core container can still reach a Windows-hosted Ollama service even when the Windows LAN IP is not directly reachable from bridge containers
- when the validation target is the Helm/self-hosted Kubernetes path, use `MYCELIS_K8S_TEXT_ENDPOINT` and optional `MYCELIS_K8S_MEDIA_ENDPOINT` to prove the deployment targets an explicit reachable AI host instead of a chart-baked or localhost default
- when the validation target is the Windows self-hosted operator lane, prove the field topology directly: a Windows browser/client opens the UI over the network, the runtime runs in Compose or self-hosted Kubernetes, and the AI engine lives on a Windows GPU host reached by explicit IP or hostname rather than `localhost`
- use `uv run inv ci.release-preflight --service-health --live-backend` when a branch changes proxy/runtime/service contracts and needs both clean-tree proof and live service/browser evidence
- when live browser proof asserts backend-written files from a different worktree than the running Core backend, set `MYCELIS_BACKEND_WORKSPACE_ROOT` (or `PLAYWRIGHT_BACKEND_WORKSPACE_ROOT`) to the backend's actual workspace root before running the spec, such as `core/workspace` for a repo-local Core process or `workspace/docker-compose/data/workspace` for the supported compose stack
- docs, tasks, and release language must stay synchronized with the actual validation gate in the same slice
- team-creation and orchestration changes must prove the compact-default rule: small teams by default, broad asks split into several smaller lanes, and the coordination path remains visible through Soma/Council/NATS rather than being hidden in a giant roster

Canonical full-gate references:
- use this document for the ordered release-style testing pass across repo baseline, stable browser proof, live service/browser proof, and compose-aware runtime proof
- use `docs/REMOTE_USER_TESTING.md` when the goal is a human walkthrough from a different machine on the same network rather than a repo-local engineering gate
- use `docs/architecture-library/V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md` for the Soma-first browser contract
- use `docs/architecture-library/V8_UI_TEAM_FULL_TEST_SET.md` for the full UI/browser validation set

## Target-Action Testing Matrix (Intent -> Manifestation)

The same target actions defined in architecture docs must have matching test actions.

| Target action | Required test actions | Status |
| --- | --- | --- |
| Soma-first Team Expression + module binding | Vitest component/store coverage for expression editing + module binding render states; integration coverage for normalized adapter payloads; product-flow proof of `proposal` -> `execution_result` | `ACTIVE` |
| Created-team workspace + channel inspector | UI tests for communication filters; integration tests for created-team command -> `signal.status`/`signal.result`; product-flow proof for interject/reroute/pause-resume controls | `REQUIRED` |
| Scheduler recurring execution (`scheduled_missions`) | backend schedule CRUD/tick tests, restart/rehydration persistence tests, recurring-state UI tests | `NEXT` |
| Causal chain operator UI | `/runs/[id]/chain` page/component tests + chain API mapping/error-state integration tests | `REQUIRED` |

Cross-reference:
- `docs/architecture-library/INTENT_TO_MANIFESTATION_AND_TEAM_INTERACTION_V7.md`
- `docs/architecture-library/TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md`
- `docs/architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md`

## Full GUI Coverage Matrix

This matrix is route-driven and code-verified against `interface/app/**`, `interface/__tests__/pages/**`, and `interface/e2e/specs/**`.

| GUI surface | Unit/component coverage | Playwright coverage | Status |
| --- | --- | --- | --- |
| `/organizations/[id]` Soma-primary AI Organization workspace | `OrganizationPage.test.tsx`, organization/workspace/store suites | `v8-organization-entry.spec.ts`, `v8-ui-testing-agentry.spec.ts`, plus live-backend proof via `soma-governance-live.spec.ts` for governed Soma execution and `workspace-live-backend.spec.ts` for proxy/status continuity when those contracts change | `ACTIVE` |
| `/dashboard` AI Organization re-entry and status overview | `DashboardPage.test.tsx`, dashboard/store suites | `missions.spec.ts`, `navigation.spec.ts`, accessibility baseline | `ACTIVE` |
| `/groups` standing + temporary collaboration workspace | `GroupsPage.test.tsx`, `GroupManagementPanel.test.tsx` | `groups.spec.ts`, plus live-backend retained-output aggregation proof via `groups-live-backend.spec.ts` | `ACTIVE` |
| `/automations` | `AutomationsPage.test.tsx`, automations component suites | `layout.spec.ts`, `proposals.spec.ts` | `ACTIVE` |
| `/resources` (+ redirects from `/catalogue`, `/marketplace`) | `ResourcesPage.test.tsx`, redirect page tests | `catalogue.spec.ts` (partial), `mcp-connected-tools.spec.ts` for Connected Tools registry/library/activity visibility and empty-state activation guidance | `ACTIVE` |
| `/memory` | `MemoryPage.test.tsx`, memory component suites | `memory.spec.ts` (live-backend-gated via `PLAYWRIGHT_LIVE_BACKEND`) | `ACTIVE` |
| `/system` (+ redirects from `/telemetry`, `/matrix`) | `SystemPage.test.tsx`, redirect page tests | route-level smoke remains unit-first; live-backend/browser depth is still selective | `ACTIVE` |
| `/settings` (+ `/settings/tools`) | `SettingsPage.test.tsx`, settings component suites, `MCPToolRegistry.test.tsx`, `MCPLibraryBrowser.test.tsx` | `settings.spec.ts` for guided settings/profile/access/theme; Connected Tools browser depth now lives on `/resources?tab=tools` in `mcp-connected-tools.spec.ts` | `ACTIVE` |
| `/runs`, `/runs/[id]` | run component suites (`RunDetailPage`, timeline, cards) + `RunsPage.test.tsx` | browser depth intentionally secondary to the MVP route gate | `ACTIVE` |
| `/docs` in-app browser | `DocsPage.test.tsx` | `docs-and-runs.spec.ts` docs manifest/render smoke | `ACTIVE` |
| `/teams` team roster + specialization hub | `TeamsPage.test.tsx`, `pages/TeamsPage.test.tsx` | `teams.spec.ts` | `ACTIVE` |
| `/teams/create` guided team-creation workflow | `TeamCreationPage.test.tsx`, `TeamLeadInteractionPanel.test.tsx`, `pages/CreateTeamPage.test.tsx` | `team-creation.spec.ts` | `ACTIVE` |
| Legacy redirect routes (`/wiring`, `/architect`, `/approvals`, etc.) | page redirect tests present | indirect via workflow-parent specs | `COMPLETE` |

### Team-Creation / Compact-Default Test Contract

When the work changes team defaults or team shaping behavior, add proof across four layers:

- backend contract: verify the guidance response or routing contract keeps ordinary requests compact and splits broad asks into multiple smaller lanes instead of one giant team
- UI workflow: verify the creation surface explains the compact default and shows the broad-ask split clearly enough for an operator to understand the plan
- NATS/exchange observability: verify that multi-team orchestration stays inspectable through governed status/result/review surfaces and that lane handoffs remain visible
- browser workflow: verify a real headed browser path can create a focused team, review the lead, and confirm the broad-ask case becomes multiple compact lanes with retained outputs rather than a monolithic roster

Recommended evidence targets:
- `uv run inv core.test`
- `cd interface; npx vitest run __tests__/teams/TeamCreationPage.test.tsx __tests__/organizations/TeamLeadInteractionPanel.test.tsx --reporter=dot`
- `cd interface; npx tsc --noEmit`
- `uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/team-creation.spec.ts`
- a live-backend browser proof if the slice changes runtime coordination or retained-output behavior

Immediate test additions required for stronger full-stack confidence:
1. `COMPLETE` add a focused Connected Tools browser proof for MCP registry/library/activity visibility beyond component coverage.
2. `COMPLETE` add a browser proof for direct Soma output vs team-managed output package delivery.
3. `COMPLETE` add browser proof for media artifact rendering/save/download, or a precise missing-media-engine blocker when no media engine is configured.
4. `NEXT` add live-backend browser proof that a Soma/team workflow can use an MCP-backed capability and surface matching recent MCP activity.
5. `NEXT` unskip and keep green the guided Soma retry/recovery browser scenario so first-run failure handling stays proven.
6. `NEXT` expand `/docs` coverage to include markdown internal-link traversal and manifest/read failure fallback branches.

### Tester Proof Slice: Output Block, Media Readiness, And Team-Managed Review

Use this compact slice when you need to prove the current media/output lane without expanding into backend runtime changes:

1. **Local-hosted output block**
   - Set `MYCELIS_OUTPUT_BLOCK_MODE=local_hosted`.
   - Set `MYCELIS_OUTPUT_HOST_PATH` to an existing host directory before Compose bring-up.
   - Run `uv run inv compose.up --build`, then `uv run inv compose.health`.
   - In the browser, confirm team or Soma-generated file outputs resolve to download links that point at the mounted output location and the UI presents them as locally hosted artifacts.

2. **Cluster-generated output block**
   - Use the chart/runtime default `cluster_generated` mode.
   - Render or deploy the chart with PVC-backed output storage.
   - Confirm the UI still exposes downloadable retained outputs, but the operator path is treated as cluster-managed storage rather than a host-mounted directory.

3. **Live media provider readiness**
   - Run `uv run inv lifecycle.health` first.
   - If the media engine is offline, record the blocker explicitly instead of treating the run as passed.
   - If a media provider is available, use the headed Chromium proof in `interface/e2e/specs/v8-ui-testing-agentry.spec.ts` to validate generated media artifact rendering, save, and download behavior.

4. **Team-managed output review**
   - Run the guided team creation path in `interface/e2e/specs/team-creation.spec.ts`.
   - Confirm the temporary workflow group can be created, archived, and reviewed with retained outputs still visible.
   - Pair that with `interface/e2e/specs/v8-ui-testing-agentry.spec.ts` to prove direct Soma answers stay distinct from team-managed output packages.

### Windows Self-Hosted Validation Lane

Use this lane when you need repeatable proof that the self-hosted product works the way a Windows operator actually uses it:

- the browser session runs on Windows
- the UI is reached over the network, not through a local-only shortcut
- the runtime is Compose or self-hosted Kubernetes
- the AI engine runs on a Windows GPU host or equivalent self-hosted service reached by explicit IP or hostname

Required setup:

1. Set the runtime endpoint explicitly in the compose or Helm environment, for example `MYCELIS_COMPOSE_OLLAMA_HOST=http://<windows-ai-host>:11434` or the provider-specific endpoint override used by the deployment.
   - On Windows + WSL Docker, `compose.up` may relay that endpoint through `host.docker.internal:<relay-port>` on the WSL host as an implementation detail; the operator-facing contract is still the explicit Windows AI host.
2. Bring the stack up with the supported runtime task path.
3. Confirm `compose.status` or the relevant Kubernetes health proof shows the UI and backend are healthy before browser work begins.
4. Record the Windows host address used for the model service and the browser URL used for the UI.

Required proof sequence:

1. Open the UI from Windows using the network-reachable host name or IP.
2. Confirm Soma returns a direct `answer` for a non-mutating prompt.
3. Confirm a mutating prompt enters `proposal` and can be approved or cancelled.
4. Confirm guided team creation reaches a real team or temporary workflow lane.
5. Confirm retained outputs remain available after the lane is archived or closed.
6. Confirm the run still works after a browser reload or fresh session.
7. Confirm the operator can see a clear blocker if the AI host becomes unreachable, and that recovery restores the normal flow once the endpoint is reachable again.

Recommended evidence targets:

- `uv run inv compose.up --build --wait-timeout=240`
- `uv run inv compose.status`
- `uv run inv compose.health`
- `uv run inv interface.e2e --headed --live-backend --server-mode=start --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts`
- `uv run inv interface.e2e --headed --live-backend --server-mode=start --project=chromium --spec=e2e/specs/team-creation.spec.ts`
- `uv run inv interface.e2e --headed --live-backend --server-mode=start --project=chromium --spec=e2e/specs/groups-live-backend.spec.ts`
- `uv run inv interface.e2e --headed --project=chromium --spec=e2e/specs/v8-ui-testing-agentry.spec.ts`

Pass condition:

- the browser path works from Windows
- the runtime is self-hosted and not Docker Desktop-dependent
- the AI engine endpoint is explicit and non-loopback
- the operator can complete a normal user journey, a governed mutation, and a retention/recovery check in the same lane
- failure of the AI host is visible and recoverable instead of being hidden by browser-only success

7. `NEXT` deepen `/runs` and `/runs/[id]` browser coverage for interjection path, terminal status transitions, and retry/error states.

Canonical UI testing agentry contract:
- `docs/architecture-library/V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md`
- Stable operator-flow browser proof: `interface/e2e/specs/v8-ui-testing-agentry.spec.ts`
- Live backend governed-chat proof: `interface/e2e/specs/soma-governance-live.spec.ts`

## Backend/API -> UI Target Plan (Required)

When backend/API behavior changes, attach this plan block in the same slice before review:

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
  - <focused playwright command(s) for impacted UI path>
  - <live-backend playwright command when proxy/core contract changed>
```

Minimum policy:
- no backend/API review without a mapped UI target plan
- no `COMPLETE` status without executed evidence commands and pass/fail results

## Clean Run Discipline for Runtime and Integration Checks

- Before any runtime or integration-style test, stop prior local services using the repo lifecycle task path. Use `uv run inv lifecycle.down` unless a narrower repo task is the safer equivalent for the slice.
- For Docker Compose runtime proof, use `uv run inv compose.down --volumes` as the clean reset equivalent so PostgreSQL, NATS, and Core state are truly rebuilt instead of inherited from a prior container run.
- For personal-owner Compose data-plane proof, use `uv run inv compose.infra-up --wait-timeout=180` and `uv run inv compose.infra-health` before full app launch, and expect PostgreSQL/NATS to be `UP` while Core API and Frontend remain `DOWN`.
- For personal-owner long-term storage proof, run `uv run inv compose.migrate` and `uv run inv compose.storage-health` before claiming semantic memory, deployment context, retained artifacts, managed exchange, or conversation continuity are available.
- For partially migrated Compose volumes, `compose.migrate` should not replay older unsafe migrations when the base schema is already compatible; it should apply only known missing late storage migrations and then rely on `compose.storage-health` for proof.
- For Docker Compose runtime proof, treat `uv run inv compose.up` and `uv run inv compose.migrate` as forward-bootstrap helpers, not replay-everything hammers. Once the compose `cortex` schema is already compatible with the current runtime, they intentionally skip replay and point to `uv run inv compose.down --volumes` when a truly fresh rebuild is required.
- For Docker Compose runtime proof, prefer `uv run inv compose.up --build --wait-timeout=240` on a fresh host or first rebuild so readiness checks stay aligned with slower image-build timelines instead of failing early at the caller layer.
- For Docker Compose runtime proof, treat `uv run inv compose.health` as a product-availability gate, not only a port gate: the text cognitive engine must be online before browser proof is considered valid.
- Verify ports and processes are clear for the services involved in the check. At minimum review the Core API port, NATS, PostgreSQL, and Ollama when the slice depends on them, using repo ops tasks such as `uv run inv lifecycle.status` or OS-level port/process tools.
- Detect running compiled binaries with process inspection before the test begins. Look for repo-local command lines or binary paths plus any processes bound to declared dev/test ports; if found, terminate them with the lifecycle/task helpers and never assume they belong to the current run.
- Treat repo-local Interface worker residue as part of the same cleanup surface. On Windows in particular, `next`, `vitest`, `playwright`, and generated `.next/dev/build/postcss.js` workers can survive after the owning command exits unless the task wrapper sweeps them.
- Merge-readiness browser gates use the managed low-parallelism path: `uv run inv ci.baseline` forces Playwright to `--workers=1` and runs against the built production Interface server path, while `uv run inv ci.service-check --live-backend` stays serial at `--workers=1`, restores the local bridge/core stack, and only skips `db.migrate` when the `cortex` schema is already compatible with the current runtime, so full-stack proof stays repeatable under local host load without turning the baseline into an impractical wall-clock run.
- Start only the minimal services required for the specific check. Prefer the narrowest path that matches the validation target, such as Helm render only, bootstrap/unit coverage only, Core-only, or a bounded local stack bring-up.
- Run the test or validation command once the required services are confirmed ready.
- Shut services down immediately after the check unless the slice explicitly requires them left running for a follow-on validation step.
- Agents must never stack runs on top of unknown existing processes.

### Compiled Go Service Cleanup Before Tests

Go services started through `go build`, `go run`, or direct dev binaries can outlive the test that launched them and remain running outside the normal container or bridge lifecycle. Clean-run validation must treat these binaries as first-class cleanup targets before any runtime or integration-style test.

Typical binaries/process shapes to check:
- core server
- relays / bridges when implemented as Go services
- bootstrap helpers
- any Go-based local services used by the repo, including `go run ./cmd/server`, `go run ./cmd/probe`, `go run ./cmd/signal_gen`, and similar dev/test helpers

Cleanup must distinguish and handle all three classes:
- local compiled Go services
- containerized or bridged dependencies
- test-managed ephemeral services

Stopping containers or port-forwards alone is not enough. The pre-test cleanup pass must also inspect process tables for stray Go binaries and kill them when found. If process inspection fails, treat the environment as unverified and do not continue with runtime or integration-style tests until the inspection path is healthy again.

## Quick Reference

```bash
# Run from scratch/ root.
# Primary runner: uv run inv ...
# Compatibility probe: uvx --from invoke inv -l
# Unsupported bare alias: uvx inv ...
uv run inv core.test             # Go unit tests (all packages)
uv run inv install               # Default Core + Interface dependency install (add --optional-engines only when local vLLM/Diffusers helpers are needed)
uv run inv auth.dev-key          # Ensure the primary local-admin credential exists in .env
uv run inv auth.break-glass-key  # Stamp a separate break-glass recovery credential when self-hosted recovery is part of the slice
uv run inv auth.posture          # Inspect local-admin vs break-glass auth posture before runtime/browser proof
uv run inv core.package          # Versioned Core binary archive for release handoff (not part of the default validation gate)
uv run inv interface.test        # Vitest unit tests (jsdom)
uv run inv interface.typecheck   # TypeScript typecheck through the managed Interface task path
uv run inv interface.e2e         # Playwright E2E tests (defaults to managed dev mode for stable mocked browser proof; use --server-mode=start for built/start-mode or live-backend proof. Invoke manages the server lifecycle, uses serial workers by default, keeps browser cache repo-managed, cleans repo-local UI workers, and fails if it cannot own a clean managed UI server)
uv run inv test.e2e              # Root alias for interface.e2e with the same --workers / --server-mode controls
uv run inv interface.e2e --live-backend --server-mode=start --spec=e2e/specs/soma-governance-live.spec.ts   # Real governed Soma chat + approval contract
uv run inv interface.e2e --live-backend --server-mode=start --spec=e2e/specs/workspace-live-backend.spec.ts  # Real workspace proxy/status continuity contract
uv run inv interface.e2e --headed --live-backend --server-mode=start --project=chromium --spec=e2e/specs/groups-live-backend.spec.ts  # Real temporary group + backend-stored retained output review
MYCELIS_BACKEND_WORKSPACE_ROOT=core/workspace uv run inv interface.e2e --live-backend --server-mode=start --spec=e2e/specs/soma-governance-live.spec.ts  # Repo-local Core: set explicitly when spec checkout != backend checkout
MYCELIS_BACKEND_WORKSPACE_ROOT=workspace/docker-compose/data/workspace uv run inv interface.e2e --live-backend --server-mode=start --spec=e2e/specs/soma-governance-live.spec.ts  # Supported compose stack: use the compose workspace root when the spec checkout differs
uv run inv core.smoke            # Governance smoke tests
uv run inv ci.test               # Blocking Go + Vitest validation
uv run inv interface.check       # HTTP smoke test against the running Interface server
uv run inv cache.status          # Inspect managed repo/user cache roots before large local validation runs
uv run inv cache.clean           # Prune repo-managed caches and build artifacts when disk pressure returns
uv run inv logging.check-schema  # Event schema + docs coverage gate
uv run inv logging.check-topics  # Hardcoded swarm topic gate
uv run inv quality.max-lines --limit 350  # Hot-path max-lines gate with legacy caps
uv run inv lifecycle.memory-restart --frontend          # Full memory reset + post-restart memory probes
uv run inv ci.entrypoint-check   # Verify uv / uvx runner matrix
uv run inv ci.baseline           # Canonical strict baseline (docs/logging/topics/line gates + core + interface + Playwright by default)
uv run inv ci.baseline --no-e2e  # Narrower local debug path when browser proof is intentionally skipped
uv run inv ci.service-check      # Running-stack health proof against local services
uv run inv compose.infra-up --wait-timeout=180      # Personal-owner Compose data-plane preflight: postgres + nats only, no Core/UI
uv run inv compose.infra-health      # Data-plane-only health proof: PostgreSQL + NATS without Core/UI checks
uv run inv compose.storage-health    # Post-migration long-term storage proof: pgvector, memory/context, artifacts, exchange, continuity
uv run inv compose.up --build --wait-timeout=240    # Supported home-runtime bring-up without Kind on a fresh or slower host
uv run inv compose.health        # Deep health proof for the compose stack
uv run inv ci.release-preflight --service-health --live-backend  # Clean-tree + baseline + live service/browser proof
```

Runner matrix:
- `uv run inv ...` is the supported path for real task execution and testing.
- `uvx --from invoke inv -l` is a lightweight compatibility probe only.
- `uvx inv ...` is expected to fail and is checked as a negative control by `uv run inv ci.entrypoint-check`.
- Invoke-managed Node/Go/Python validation now routes caches through `workspace/tool-cache` by default; on Windows, `uv run inv cache.apply-user-policy` persists the same posture for direct per-user tool usage.
- Project-owned backstops keep direct commands aligned too: root `.npmrc` anchors npm/npx cache under `workspace/tool-cache/npm`, pytest uses `workspace/tool-cache/pytest`, and Invoke-managed browser runs export `PLAYWRIGHT_BROWSERS_PATH` plus `NEXT_TELEMETRY_DISABLED=1`.
- Invoke-managed Interface and CI validation now runs from the `interface/` working directory with the same `npm`/`node` entrypoints on Windows and Linux, so browser/build/test cleanup does not depend on shell-specific `cd ... &&` wrappers.

Signal/channel standard:
- When tests touch NATS channel behavior, use the canonical subject families and source metadata defined in `docs/architecture/NATS_SIGNAL_STANDARD_V7.md`.
- Development-only infrastructure subjects are not part of product orchestration and should stay out of authoritative runtime tests unless the test is explicitly exercising dev-only behavior.
- Channel-private relay contract: `publish_signal` may emit `privacy_mode=reference` payloads while persisting full private payloads to checkpoint channels; relaunch recovery must use `read_signals` with `latest_only=true`.
- Current focused runtime check: `cd core && go test ./internal/swarm ./pkg/protocol -count=1`
- Current focused bootstrap migration check: `cd core && go test ./cmd/server ./internal/bootstrap ./internal/swarm -count=1`
- Current focused chart/bootstrap render check: `helm template mycelis-core charts/mycelis-core`
- Current focused Kubernetes AI-endpoint render check: `helm template mycelis-core charts/mycelis-core --set-string ai.textEndpoint=http://192.168.50.156:11434/v1 --set-string ai.mediaEndpoint=http://192.168.50.156:8001/v1`
- Current focused toolship metadata check: `cd core && go test ./internal/swarm -run "TestHandleDelegateTask_PublishesToInternalCommand|TestHandlePublishSignal_WrapsCanonicalStatusSubject|TestHandlePublishSignal_PrivateReferenceAndCheckpoint|TestHandleReadSignals_LatestOnlyReturnsCheckpoint|TestTeam_TriggerLogic_UnwrapsCommandEnvelope|TestAgentPublishToolBusSignal_StatusChannelForMCP|TestAgentPublishToolBusSignal_ResultChannelForMCP|TestAgentPublishToolBusSignal_PersistsLatestCheckpoint" -count=1`
- Current focused agent parsing/preflight check: `cd core && go test ./internal/swarm -run "TestParseConversationPayload_|TestParseToolCall|TestAutofillToolArguments|TestShouldCouncilPreflight|TestCouncilPreflightMember" -count=1`
- Current focused stream-normalization check: `cd interface && npx vitest run __tests__/lib/signalNormalize.test.ts __tests__/store/useCortexStore.test.ts --reporter=dot`
- Current focused docs/runs page check: `cd interface && npx vitest run __tests__/pages/DocsPage.test.tsx __tests__/pages/RunsPage.test.tsx __tests__/runs/RunDetailPage.test.tsx --reporter=dot`
- Current focused docs/runs browser check: `cd interface && npx playwright test e2e/specs/docs-and-runs.spec.ts --project=chromium`
- Current focused store-utils check: `cd interface && npx vitest run __tests__/store/cortexStoreUtils.test.ts __tests__/store/useCortexStore.test.ts --reporter=dot`
- Current focused Workspace chat contract check: `cd interface && npx vitest run __tests__/dashboard/MissionControlChat.test.tsx __tests__/lib/labels.test.ts --reporter=dot`
- Current focused execution feedback check: `cd interface && npx vitest run __tests__/dashboard/CouncilCallErrorCard.test.tsx __tests__/dashboard/DegradedModeBanner.test.tsx --reporter=dot`
- Current focused Workspace failure-model check: `cd interface && npx vitest run __tests__/lib/missionChatFailure.test.ts __tests__/dashboard/CouncilCallErrorCard.test.tsx __tests__/dashboard/DegradedModeBanner.test.tsx __tests__/dashboard/StatusDrawer.test.tsx __tests__/dashboard/MissionControlChat.test.tsx __tests__/store/useCortexStore.test.ts --reporter=dot`
- Current focused Launch Crew contract check: `cd interface && npx vitest run __tests__/workspace/LaunchCrewModal.test.tsx __tests__/store/useCortexStore.test.ts --reporter=dot`
- Current focused Launch Crew browser proof: `uv run inv interface.e2e --live-backend --server-mode=start --project=chromium --spec=e2e/specs/proposals.spec.ts` (live AI Organization workspace entry, proposal outcome, normalized blocker recovery, and confirm UI)
- Current focused Launch Crew live confirm proof: the same spec uses a real `/api/v1/chat` generated proposal token plus real `/api/v1/intent/confirm-action` durable execution proof, and removes generated QA files after the run.
- Current focused groups workspace contract check: `cd interface && npx vitest run __tests__/pages/GroupsPage.test.tsx __tests__/teams/GroupManagementPanel.test.tsx --reporter=dot`
- Current focused groups workspace browser proof: `uv run inv interface.e2e --project=chromium --spec=e2e/specs/groups.spec.ts`
- Current live-backend retained group-output proof: `uv run inv interface.e2e --headed --live-backend --server-mode=start --project=chromium --spec=e2e/specs/groups-live-backend.spec.ts`
- Current guided team-creation launch-path contract check: `cd interface && npx vitest run __tests__/organizations/TeamLeadInteractionPanel.test.tsx __tests__/teams/GroupManagementPanel.test.tsx __tests__/pages/GroupsPage.test.tsx --reporter=dot`
- Current guided team-creation lifecycle browser proof: `uv run inv interface.e2e --headed --project=chromium --spec=e2e/specs/team-creation.spec.ts` (guided design -> launch temporary workflow group -> open Groups -> review multiple outputs -> archive temporary lane -> retained-output review)
- Current focused teams workspace contract check: `cd interface && npx vitest run __tests__/pages/TeamsPage.test.tsx __tests__/pages/CreateTeamPage.test.tsx __tests__/teams/TeamsPage.test.tsx __tests__/teams/TeamCreationPage.test.tsx __tests__/organizations/TeamLeadInteractionPanel.test.tsx --reporter=dot`
- Current focused teams workspace browser proof: `uv run inv interface.e2e --project=chromium --spec=e2e/specs/teams.spec.ts`
- Current focused guided team-creation browser proof: `uv run inv interface.e2e --project=chromium --spec=e2e/specs/team-creation.spec.ts`
- Current focused UI testing agentry browser proof: `uv run inv interface.e2e --project=chromium --spec=e2e/specs/v8-ui-testing-agentry.spec.ts`
- Current focused Windows self-hosted browser proof: `MYCELIS_COMPOSE_OLLAMA_HOST=http://<windows-ai-host>:11434 uv run inv compose.up --build --wait-timeout=240` followed by `uv run inv interface.e2e --headed --live-backend --server-mode=start --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts`
- Current focused direct-vs-team and media-output browser proof: `uv run inv interface.e2e --project=chromium --workers=1 --spec=e2e/specs/v8-ui-testing-agentry.spec.ts`; if the Windows Invoke-managed server wrapper is unhealthy, use a single already-started Interface listener and run the PowerShell fallback from `interface/`: `$env:PLAYWRIGHT_SKIP_WEBSERVER='1'; $env:PLAYWRIGHT_PORT='3100'; npx playwright test e2e/specs/v8-ui-testing-agentry.spec.ts --project=chromium --workers=1 --grep "distinguishes|renders generated" --timeout=60000`.
- Current focused Connected Tools MCP browser proof: `uv run inv interface.e2e --project=chromium --workers=1 --spec=e2e/specs/mcp-connected-tools.spec.ts`; if the Windows Invoke-managed server wrapper is unhealthy, use a single already-started Interface listener and run the PowerShell fallback from `interface/`: `$env:PLAYWRIGHT_SKIP_WEBSERVER='1'; $env:PLAYWRIGHT_PORT='3100'; npx playwright test e2e/specs/mcp-connected-tools.spec.ts --project=chromium --workers=1 --timeout=60000`.
- Current focused UI testing agentry live governance proof: `uv run inv interface.e2e --live-backend --server-mode=start --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts`
- Current focused team-sync contract check: `$env:PYTHONPATH='.'; uv run pytest tests/test_misc_tasks.py -q`
- Current focused README navigation check: `$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q`
- Current docs/task drift rule: canonical docs must not contain executable bare `uvx inv ...` examples outside explicit negative-control guidance.

UI delivery contract:
- Use `docs/architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md` as the authoritative map for UI terminal states and backend transaction expectations.
- A UI test is incomplete if it proves rendering but does not prove the backend effect or intentionally blocked state behind that render.

---

## Product Delivery Proof (Required)

For execution-facing UI work, tests must prove product behavior, not only component mechanics.

Every changed UI path must document and test:
1. the initiating user interaction
2. the expected terminal UI state: `answer`, `proposal`, `execution_result`, or `blocker`
3. the backend effect caused by the frontend: HTTP call, DB mutation, run/event creation, or NATS interaction
4. the failure path and recovery affordance

Minimum proof requirements by path:

| UI Path | UI Proof | Backend/Transaction Proof |
| --- | --- | --- |
| Workspace / Soma chat | response lands in one valid terminal state, not planning-only output | `/api/v1/chat` call occurs and returned payload is classified correctly |
| Direct council chat | specialist answer or structured blocker card renders | `/api/v1/council/{member}/chat` path is exercised and timeout/failure behavior is mapped |
| Prime-team sync | architecture directives publish to canonical team lanes and operator-visible replies are collected | `swarm.team.{team}.internal.command` publishes and `signal.status`/`signal.result` replies are observed |
| Launch Crew / guided manifestation | proposal or activation result is visible | proposal/confirm endpoints produce identifiers and mutation state |
| Workflow composer | invalid graph blocks, valid graph proposes or activates | validate/compile/activate endpoints called with expected payloads |
| Runs / timeline / chain | operator can inspect run state and outcome | run/event/chain/conversation queries return and are rendered consistently |
| System / degraded mode | recovery guidance is visible, not only colored status | health/status responses map to explicit degraded-state actions |

Required test layers for UI-affecting delivery:
- component tests: terminal state rendering and action affordances
- integration tests: request/response mapping between UI and backend
- product-flow tests: user journey reaches a real outcome
- backend transaction tests: route/DB/NATS side effects match UI claims
- failure tests: timeout, rejection, degraded dependency, retry/reroute flow

For operator-facing blocker work, a passing test set must prove:
- the store preserves raw diagnostics
- the shared failure model classifies the error once
- Workspace blocker card, degraded banner, and status drawer all render from that shared model
- the final chat surface never exposes raw tool-call JSON, raw structured runtime envelopes, or raw admin/council transport strings as the visible answer

Disallowed testing posture:
- proving only that a button renders
- proving only that a fetch function was called without validating resulting user state
- treating planning-only content as a passing success state
- validating backend code in isolation when the UI contract is the feature being changed

---

## Tier 1: Backend Unit Tests (Go)

**Goal:** Verify API handler logic, nil guards, validation, and SQL interactions.
**Speed:** < 5s (mocked DB via `go-sqlmock`).

### Location

| File | Coverage |
|------|----------|
| `core/internal/server/governance_test.go` | Policy CRUD, pending approvals, resolve approve/reject |
| `core/internal/server/telemetry_test.go` | Runtime telemetry, trust threshold GET/PUT/range |
| `core/internal/server/mission_test.go` | Mission CRUD, intent commit TX, negotiate, sensor configs, blueprint extraction |
| `core/internal/server/mcp_test.go` | MCP install validation, list, delete, tool call, library |
| `core/internal/server/mcp_toolsets_test.go` | MCP tool set list/create/update/delete paths + nil guards |
| `core/internal/server/memory_search_test.go` | Memory search, sitreps, sensors |
| `core/internal/server/proposals_test.go` | Proposal CRUD, approve/reject, conflict detection |
| `core/internal/server/identity_test.go` | Identity, teams, settings |
| `core/internal/server/catalogue_test.go` | Agent catalogue CRUD |
| `core/internal/server/artifacts_test.go` | Artifact storage |
| `core/internal/server/cognitive_test.go` | Cognitive router config |
| `core/internal/server/testhelpers_test.go` | Shared test helpers (`newTestServer`, `withDB`, `withGuard`, `doRequest`, `assertStatus`, `assertJSON`) |
| `core/internal/cognitive/middleware_test.go` | LLM retry, schema validation, timeouts |

### Patterns

- **Partial server construction:** Use `newTestServer(opts...)` with option functions (`withDB`, `withGuard`) to build only the dependencies a handler needs.
- **SQL mocking:** `go-sqlmock` for all DB interactions. Use `ExpectQuery`/`ExpectExec` with regex patterns.
- **Path params:** Use `http.NewServeMux` with Go 1.22+ route patterns (`"GET /path/{id}"`) for `r.PathValue()` support.
- **Nil guards:** Every handler that touches optional infrastructure (Guard, Overseer, MCP, Cognitive, Mem) has a nil-guard test returning 503.

### Running

```bash
uv run inv core.test                          # All packages
go test -v ./internal/server/...           # Server handlers only
go test -v -run TestHandleGovernance ./internal/server/...  # Single test pattern
go test -v ./internal/mcp/ -count=1        # MCP service/library/executor/toolset suites
go test -v -run TestHandleMCP ./internal/server/... -count=1
go test -v -run TestHandleUpdateToolSet ./internal/server/... -count=1
go test -v -run TestScoped ./internal/swarm/... -count=1
```

---

## Tier 2: Frontend Unit Tests (Vitest)

**Goal:** Verify component rendering, store interactions, UI transaction mapping, and terminal delivery states in jsdom.
**Speed:** < 10s.

### Location

| Directory | Coverage |
|-----------|----------|
| `interface/__tests__/shell/` | ShellLayout, ZoneA_Rail, GovernanceModal |
| `interface/__tests__/dashboard/` | MissionControlChat, ProposedActionBlock, StatusDrawer, ManifestationPanel, CouncilCallErrorCard, and degraded-state support cards |
| `interface/__tests__/pages/` | Dashboard, Organization, Teams, Automations, Resources, Settings, Docs, System, Runs, and remaining legacy redirect route coverage |
| `interface/__tests__/store/` | Cortex store contract, helpers, delivery-state transitions |
| `interface/__tests__/workspace/` | LaunchCrewModal, CircuitBoard, ArchitectChat, DeliverablesTray, BlueprintDrawer, TrustSlider |

### Infrastructure

| File | Purpose |
|------|---------|
| `interface/__tests__/setup.ts` | Global setup: `mockFetch()`, `MockEventSource`, `next/navigation` mock |
| `interface/__tests__/mocks/reactflow.ts` | ReactFlow jsdom mock (ResizeObserver, components, hooks, enums) |
| `interface/vitest.config.ts` | jsdom environment, `@/` alias, excludes `e2e/**` |

### Patterns

- **API mocking:** Use `mockFetch` from `setup.ts` — no MSW needed. Call `mockFetch.mockResolvedValueOnce({...})`.
- **Store state:** Set Zustand state directly via `useCortexStore.setState({...})`.
- **ReactFlow components:** Import mock via `vi.mock('reactflow', () => import('../mocks/reactflow'))`.
- **Next.js navigation:** `usePathname` and `useRouter` mocked globally in `setup.ts`.
- **Terminal state assertions:** For execution-facing surfaces, assert the final delivery state (`answer`, `proposal`, `execution_result`, `blocker`) instead of only intermediate loading behavior.
- **Transaction assertions:** When a component triggers an API call, assert the expected request target/payload and the resulting user-visible outcome.

### Running

```bash
uv run inv interface.test                     # All Vitest tests
npx vitest run --reporter=verbose          # Verbose output (from interface/)
npx vitest run __tests__/shell/            # Single directory
```

---

## Tier 3: End-to-End Tests (Playwright)

**Goal:** Verify full user journeys through the running application.
**Speed:** 30s-2min (Invoke manages the Interface server for default Playwright runs; start Core separately only for live-backend specs).

For execution-facing UI work, Playwright coverage should prefer user stories with real closure:
- direct answer returned
- proposal created and confirmable
- run created and inspectable
- structured blocker with recovery path
- browser-visible operator proof for the critical Soma path should use a headed Chromium run before a slice is called ready
- runtime-state prompts such as `what is your current state` and `what teams currently exist` should return a deterministic direct answer on the live `/api/v1/chat` route instead of weak provider fallback text

### Location

| Spec | Coverage |
|------|----------|
| `interface/e2e/specs/missions.spec.ts` | Dashboard load, default navigation, organization-entry actions |
| `interface/e2e/specs/governance.spec.ts` | Approvals page, policy tab, pending section |
| `interface/e2e/specs/catalogue.spec.ts` | Catalogue page, agent cards, create button |
| `interface/e2e/specs/settings.spec.ts` | Settings guided setup, profile/access/theme persistence; MCP registry browser proof is still a required focused addition |
| `interface/e2e/specs/layout.spec.ts` | Shell structure, zone rendering |
| `interface/e2e/specs/navigation.spec.ts` | Route transitions, active states |
| `interface/e2e/specs/trust_economy.spec.ts` | Automations approvals reachability |
| `interface/e2e/specs/telemetry.spec.ts` | Legacy raw-endpoint probe (skipped in default MVP gate) |
| `interface/e2e/specs/memory.spec.ts` | Memory explorer, search |
| `interface/e2e/specs/proposals.spec.ts` | Proposal CRUD flow |
| `interface/e2e/specs/v8-ui-testing-agentry.spec.ts` | Stable V8 operator-flow proof for Soma-first entry, continuity, cold-start recovery, governed mutation/cancel, audit visibility, and oversized content handling |
| `interface/e2e/specs/groups.spec.ts` | Standing vs temporary vs archived temporary groups, output/contributing-lead summaries, retained output review, and broadcast workflow |
| `interface/e2e/specs/groups-live-backend.spec.ts` | Live-backend temporary group creation, backend-stored output aggregation, archive/closure, and retained-output review |
| `interface/e2e/specs/team-creation.spec.ts` | Guided team creation, Soma execution-path return, one-click temporary workflow group launch, archive/closure, and retained-output review |
| `interface/e2e/specs/teams.spec.ts` | Teams hub, roster, lead-entry links, guided-creation handoff |
| `interface/e2e/specs/wiring-edit.spec.ts` | Neural wiring, agent edit/delete |
| `interface/e2e/specs/v7-operational-ux.spec.ts` | Legacy V7 operator UX probe (skipped in default MVP gate) |
| `interface/e2e/specs/mobile.spec.ts` | Mobile landing-page smoke coverage under the dedicated mobile Playwright project |
| `interface/e2e/specs/accessibility.spec.ts` | Axe-backed accessibility baseline for key operator surfaces |
| `interface/e2e/specs/workspace-live-backend.spec.ts` | Real Workspace contract coverage against live `/api/v1/services/status` and `/api/v1/council/members` traffic |

Focused additions planned for the MVP media/team-output lane:

- `interface/e2e/specs/mcp-connected-tools.spec.ts`: Connected Tools registry/library/activity visibility and high-risk install posture.
- `interface/e2e/specs/media-output.spec.ts`: mocked image/audio/file/document artifacts, visible `Artifact result`, save action, saved path, and download references.
- `interface/e2e/specs/direct-vs-team-output.spec.ts`: direct inline Soma answer vs team-managed deliverable package with target outputs and retained output framing.
- live media extension to `soma-governance-live.spec.ts` or a gated `media-output-live.spec.ts`: real artifact display/download when the media engine exists, or a precise missing-engine blocker when it does not.
- compose output-block proof: when testing a local or Pinokio/media-hosted output directory, set `MYCELIS_OUTPUT_BLOCK_MODE=local_hosted` and `MYCELIS_OUTPUT_HOST_PATH` to the host directory mounted as Core `/data`; the Compose task resolves the path with Python `pathlib` and should fail early if the directory is missing or not a directory. For chart/Kubernetes proof, keep the output block on the cluster-managed PVC and treat host-path mounts as an explicit self-hosted exception.

### Configuration

- **Config:** `interface/playwright.config.ts`
- **Base URL:** `http://127.0.0.1:3000` by default for local browser traffic (`INTERFACE_HOST` / `INTERFACE_PORT` override supported)
- **Bind Host:** the managed Next.js server binds to `[::]:3000` by default so IPv4 localhost, IPv6 localhost, and LAN clients can all reach the UI (`INTERFACE_BIND_HOST` / `MYCELIS_INTERFACE_BIND_HOST` override supported)
- **Browser Projects:** `chromium`, `firefox`, `webkit`, `mobile-chromium`
- **Server Lifecycle:** `uv run inv interface.e2e` starts/stops the managed Next.js app and now defaults to the managed `dev` server for stable mocked browser proof; use `--server-mode=start` when you need the built production Interface server path for stricter or live-backend proof. Start-mode runs refresh the production bundle before the managed server launches and retry once after a stale repo-local Next build lock, a stale `.next/standalone` cleanup lock, or incomplete built-server packaging
- **Task Cleanup:** `uv run inv interface.e2e` stops any stale listener on `:3000` before and after each run, launches a managed local Next.js server, defaults Playwright to `--workers=1`, sweeps repo-local Next/Vitest/Playwright worker residue, clears an orphaned `interface/.next/dev/lock` only when no repo-local Next worker remains, and fails closed if the managed server exits or a stale port prevents it from owning the browser target
- **Managed Browsers:** default task runs expect Playwright browser binaries under `workspace/tool-cache/playwright`
- **Live Backend Mode:** `uv run inv interface.e2e --live-backend --server-mode=start ...` loads proxy auth env and enables specs that require a real Core backend
- **Browser-Visible UX Pass:** use `--headed --project=chromium --server-mode=start` for the operator-facing certification sweep so the same run exercises the real browser window, current production bundle, and actual rendered recovery states
- **Accessibility Gate:** `@axe-core/playwright` is a required dev dependency; accessibility specs must fail when violated, not skip because the package is missing
- **Dark mode compliance:** Every spec includes `no bg-white` assertion

### Running

```bash
# Core is only required for specs that hit the real backend instead of route stubs.
uv run inv core.run          # Optional: start live backend coverage in a separate terminal

# Run E2E tests
uv run inv interface.e2e                     # All specs
uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/v8-ui-testing-agentry.spec.ts
uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/v8-organization-entry.spec.ts
uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/teams.spec.ts
uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/team-creation.spec.ts
uv run inv interface.e2e --headed --project=chromium --server-mode=start --spec=e2e/specs/settings.spec.ts
uv run inv interface.e2e --live-backend --server-mode=start --spec=e2e/specs/workspace-live-backend.spec.ts
uv run inv interface.e2e --headed --live-backend --server-mode=start --project=chromium --spec=e2e/specs/groups-live-backend.spec.ts
uv run inv interface.e2e --headed --live-backend --server-mode=start --project=chromium --spec=e2e/specs/soma-governance-live.spec.ts
uv run inv interface.e2e --project=firefox
uv run inv interface.e2e --project=mobile-chromium --spec=e2e/specs/mobile.spec.ts
npx playwright test --project=chromium    # From interface/
npx playwright test e2e/specs/missions.spec.ts  # Single spec
npx playwright test e2e/specs/accessibility.spec.ts --project=chromium
npx playwright show-report                # View HTML report
```

---

## Tier 4: Integration Tests (Go)

**Goal:** Verify the real model (Ollama/OpenAI) understands prompts and schemas.
**Speed:** 1s-30s (depends on model).
**Build Tag:** `//go:build integration` (skipped by default `go test`)

### Location

- `core/tests/agent_interaction_test.go`

### Running

```bash
# Requires: Ollama running + model pulled
# ollama pull qwen2.5-coder:7b-instruct

# Windows (PowerShell)
$env:OLLAMA_HOST="http://192.168.50.156:11434"; go test -v -tags=integration ./tests/...

# Linux/Mac
OLLAMA_HOST=http://192.168.50.156:11434 go test -v -tags=integration ./tests/...
```

---

## Tier 5: Governance Smoke Tests (System)

**Goal:** Verify the Gatekeeper blocks dangerous actions and routes approvals correctly.

### Protocol

1. Start the Core: `uv run inv core.run`
2. Inject poison: Send a message with intent `k8s.delete.pod`
3. Verify block: Check logs for "Gatekeeper DENIED"
4. Inject require approval: Send `payment.create` with amount `100`
5. Verify inbox: Check `/approvals` for the pending request

### Running

```bash
uv run inv core.smoke
```

---

## Memory Restart Validation

Use this when memory path behavior is suspect (stale stream, sitrep/read model drift, failed local state recovery).

```bash
uv run inv lifecycle.memory-restart --build --frontend
```

Expected command outcomes:
- stack teardown/restart completes
- forward-only migration set applies cleanly (`001_init_memory.sql` + `*.up.sql`)
- health probe passes
- memory probes return HTTP 200:
  - `/api/v1/memory/stream`
  - `/api/v1/memory/sitreps?limit=1`
- if Core fails before binding `:8081`, inspect `workspace/logs/core-startup.log`

---

## CI Pipelines

Three GitHub Actions workflows remain available for PR-time and manual validation while push-triggered pipeline runs are intentionally disabled until the initial release-readiness gate is accepted:

| Workflow | File | What it does |
|----------|------|-------------|
| **Core CI** | `.github/workflows/core-ci.yaml` | Workflow-native Python/uv + Go bootstrap, `uv run inv core.test`, coverage, GolangCI-Lint v1.64.5, and `uv run inv core.compile` |
| **Interface CI** | `.github/workflows/interface-ci.yaml` | Workflow-native Python/uv + Node bootstrap, `npm ci`, then `uv run inv interface.lint`, `uv run inv interface.typecheck`, `uv run inv interface.test`, and `uv run inv interface.build` |
| **E2E CI** | `.github/workflows/e2e-ci.yaml` | Workflow-native Python/uv + Node bootstrap, Playwright browser install, `uv run inv interface.build`, then the stable invoke-managed browser matrix via `uv run inv interface.e2e` |

Current trigger posture:

- `core-ci.yaml`, `interface-ci.yaml`, and `e2e-ci.yaml` run on `pull_request` to `main` and `develop`
- `dev-build.yaml` and `release.yaml` are manual-only via `workflow_dispatch`
- direct push-triggered GitHub pipeline runs are paused until the initial release-ready boundary is explicitly reopened

### CI Checks

- **Go:** `go test -v -coverprofile=coverage.out ./...`
- **Go Lint:** GolangCI-Lint v1.64.5
- **TypeScript:** `uv run inv interface.typecheck` (strict type checking)
- **Frontend Lint:** `uv run inv interface.lint`
- **Frontend Tests:** `uv run inv interface.test`
- **E2E:** stable Playwright browser matrix (`chromium`, `firefox`, `webkit`, `mobile-chromium`) via `uv run inv interface.e2e` in managed `dev` mode, with `--server-mode=start` reserved for stricter or live-backend proof and the live path otherwise kept in `ci.service-check` / `ci.release-preflight`

---

## Adding New Tests

### UI Delivery Test Checklist

Before adding or modifying an execution-facing UI feature, answer these in the test file or PR notes:
1. what user interaction starts the flow?
2. what terminal UI state is expected?
3. what backend transaction proves the UI actually caused the intended effect?
4. what failure state should the user see?
5. what recovery action should remain available?

### Backend Handler Test

1. Create `core/internal/server/<handler>_test.go`
2. Import test helpers: `newTestServer`, `withDB`, `withGuard`, `doRequest`, `assertStatus`, `assertJSON`
3. Build a minimal server with only required dependencies
4. Use `go-sqlmock` for database expectations
5. Test: happy path, validation errors (400), nil guards (503), not found (404)

```go
func TestHandleMyEndpoint(t *testing.T) {
    dbOpt, mock := withDB(t)
    s := newTestServer(dbOpt)

    mock.ExpectQuery(`SELECT .+ FROM my_table`).
        WillReturnRows(sqlmock.NewRows([]string{"id", "name"}).AddRow("uuid", "test"))

    rr := doRequest(t, s.handleMyEndpoint, "GET", "/my-endpoint", "")
    assertStatus(t, rr, http.StatusOK)
}
```

### Frontend Component Test

1. Create `interface/__tests__/<area>/<Component>.test.tsx`
2. Mock external dependencies (`vi.mock(...)`)
3. Set Zustand state if needed (`useCortexStore.setState({...})`)
4. Use `mockFetch` for API calls

```tsx
import { render, screen } from '@testing-library/react'
import { MyComponent } from '@/components/area/MyComponent'

vi.mock('reactflow', () => import('../mocks/reactflow'))

test('renders content', () => {
    render(<MyComponent />)
    expect(screen.getByText('Expected')).toBeInTheDocument()
})
```

### E2E Spec

1. Create `interface/e2e/specs/<feature>.spec.ts`
2. Use `page.goto()` + `waitForLoadState('domcontentloaded')` for deterministic hydration checks (avoid long-lived stream flake from `networkidle`)
3. Always include dark mode compliance check (`no bg-white`)

```typescript
import { test, expect } from '@playwright/test'

test.describe('Feature Page', () => {
    test.beforeEach(async ({ page }) => {
        await page.goto('/feature')
        await page.waitForLoadState('domcontentloaded')
    })

    test('page loads without errors', async ({ page }) => {
        await expect(page.locator('text=Feature')).toBeVisible()
    })

    test('no bg-white leak', async ({ page }) => {
        const body = await page.content()
        expect(body).not.toContain('bg-white')
    })
})
```
