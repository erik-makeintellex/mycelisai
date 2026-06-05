# Operations & Build System
> Navigation: [Project README](../README.md) | [Docs Home](../docs/README.md)

> **Role**: The Builder
> **Language**: Python (Invoke)
> **Path**: `ops/`

## TOC

- [Runner Contract](#runner-contract)
- [Components](#components)
- [Clean Run Discipline for Runtime and Integration Checks](#clean-run-discipline-for-runtime-and-integration-checks)
- [Compiled Go Service Cleanup Before Tests](#compiled-go-service-cleanup-before-tests)
- [Directives](#directives)

## Runner Contract

- Real task execution uses `uv run inv <namespace>.<task>`.
- Compatibility probe only: `uvx --from invoke inv -l`.
- Do not use bare `uvx inv ...`.
- App-tied management logic stays in Python task modules; PowerShell is wrapper-only when the host needs it.
- Task, runtime, or validation changes are not complete until the matching docs are reviewed and updated in the same slice.
- GitHub Actions workflows are manual-only through `workflow_dispatch` for the current release-readiness push; `Full Release Candidate` is the umbrella hosted release lane, while CI/source/API/build/package workflows remain narrower manual proofs.
- Treat source-mode local run/build/test with native PostgreSQL/NATS as the first acceptance lane. Bring up full Compose, Rancher K3s, WSL/Compose, or target-cluster app proof only after local evidence is acceptable.
- Scope tasks around needed tools and Mycelis services. Do not add repo tasks that manage whole host environments such as terminating WSL distros, resetting Rancher Desktop VMs, or repairing Docker Desktop itself.

## Components
This directory contains the logic for the **Service Release Standard 1.0**.

Recommended host posture:
- Windows repo: canonical editing, review, and git-push surface for active development
- WSL `mycelis-root` deployment checkout: guarded WSL Compose proof checkout for install, build, API/UI test, Compose runtime, and deployment-mimic validation
- Windows native: run Core/Interface from source against Windows PostgreSQL and a local NATS server first; use Compose or Kubernetes only after local evidence is acceptable
- Linux GPU hosts: optional `cognitive.*` helpers are appropriate only when you intentionally want local vLLM/Diffusers
- if you switch a repo between Windows and WSL/Linux/macOS, recreate host-specific generated surfaces such as `.venv`, `interface/node_modules`, and `interface/.next`

WSL Compose proof-checkout contract:
- use the Windows repo to edit, review, and push; use Rancher Desktop K3s for Windows local Kubernetes parity and WSL when guarded Compose proof is required
- after a Windows-side commit/push, refresh the WSL proof checkout from git before running WSL Compose evidence
- keep destructive cleanup such as `git reset --hard` and source `git clean` scoped to the dedicated WSL proof checkout; `wsl.refresh` preserves generated `workspace/tool-cache`, `workspace/logs`, and `workspace/docker-compose` roots so permission-owned runtime/cache mounts do not block source refresh
- when the runtime is hosted from WSL on the same Windows machine, the required operator-facing browser path is the Windows browser at `http://localhost:3000`
- use `uv run inv wsl.status`, `uv run inv wsl.refresh`, `uv run inv wsl.validate`, and `uv run inv wsl.cycle` when you want the guarded Windows-dev -> WSL-proof task path instead of a manual handoff
- `uv run inv wsl.refresh` runs WSL git fetch noninteractively, tries a repo-local Git Credential Manager helper repair for GitHub HTTPS remotes when Git for Windows is visible from WSL, preserves generated WSL runtime/cache roots during source cleanup, and otherwise fails before reset/clean with SSH/HTTPS auth guidance; keep the boundary commit/push/fetch based instead of copying source trees
- `uv run inv wsl.validate` now bootstraps `.env.compose` from `.env.compose.example` when the clean WSL proof checkout has no local compose env yet, creates the configured Compose output-block host path when needed, sanitizes the Linux-side PATH so Windows Docker credential helpers cannot hijack Docker pulls, uses local Linux probes when the Ollama relay is prepared inside WSL, then runs Compose-safe release-preflight, Compose health/storage proof, a `compose.warm-cognitive` text-model warm-up, focused live-backend browser workflows against the Compose-delivered UI, and the Windows GUI probe; `--lane=service` and `--lane=release` keep the runtime preflight step and let this task own the Compose service/browser proof, and `--headed-browser` adds visible Playwright windows for those focused live specs when browser-visible release evidence is required
- WSL tasking is deliberately proof-checkout scoped. If the distro, Rancher Desktop, or Docker host is unhealthy, fix the host tool outside the repo task runner and then rerun the narrow validation task.

Deployment selection rule:
- local source-mode run/build/test with native PostgreSQL/NATS is the default development and review lane
- Docker Compose is the rapid local development and same-machine proof runtime
- Rancher Desktop K3s is the preferred Windows local Kubernetes validation lane for Helm behavior
- `k3d` is the preferred WSL/Linux local Kubernetes validation lane for Helm behavior
- Kubernetes / Helm is the target clustered deployment contract for self-hosted and enterprise environments
- enterprise self-hosted Kubernetes uses the same Helm chart with real cluster values
- packaged binaries fit small Linux nodes or edge/control-host roles that should point at a remote AI service
- promoted chart presets now live at `charts/mycelis-core/values-k3d.yaml`, `charts/mycelis-core/values-enterprise.yaml`, and `charts/mycelis-core/values-enterprise-windows-ai.yaml`

### Root `install` Task
`uv run inv install` now installs the supported default Core + Interface stack only.

- Use `uv run inv install --optional-engines` when you also want the local `cognitive/` extras.
- Use `uv run inv cognitive.install` if you want only the optional local engine dependencies.
- Interface dependencies are installed with `npm ci` so clean WSL/CI-style checkouts do not rewrite
  `interface/package-lock.json` during proof bootstrap.
- The supported install path now also provisions the managed Playwright Chromium binary so fresh checkouts can run the repo-owned browser proof lane without a separate manual browser install step.
- The root `uv.lock` is tracked for reproducible Python automation; update it intentionally with `uv lock` after dependency changes and verify with `uv lock --check` before release handoff.

### `version.py` (Identity)
Calculates the **Immutable Tag**: `v{SEMVER}-{SHA}`.
- Source: `../VERSION` file.
- Git: `git rev-parse --short HEAD`.

### `k8s.py` (Deployment)
Handles standards-first Kubernetes/Helm deployment automation, with Rancher Desktop K3s supported for Windows local validation, `k3d` preferred for WSL/Linux local cluster validation, and Kind retained only as an older fallback.
- **Init**: `uv run inv k8s.init` (Infra).
- **Deploy**: `uv run inv k8s.deploy` (Core).
- **Standards**: `uv run inv k8s.standards` verifies the static open-standard chart contract; add `--helm --values-file=<path>` to run offline Helm lint/template checks with vendored chart dependencies.
- **Status**: `uv run inv k8s.status` (Health).
- **Bridge**: `uv run inv k8s.bridge` now verifies the local PostgreSQL/NATS/Core port-forwards actually bind before reporting success. Core forwards from in-cluster `:8080` to the repo-local API port, `MYCELIS_API_PORT` or `8081` by default.
- **Recover**: `uv run inv k8s.recover` now fails closed when the cluster is unreachable and waits for rollout readiness before claiming recovery.
- **Backend selection**: local Kubernetes supports `MYCELIS_K8S_BACKEND=rancher` for Rancher Desktop K3s, prefers `k3d` on WSL/Linux when available, and accepts `MYCELIS_K8S_BACKEND=kind` for the older Kind workflow. Rancher mode requires an existing reachable Rancher Desktop cluster and skips image import because Rancher K3s shares the local Docker engine.
- **Windows CLI path loading**: repo task startup prepends standard local tool bins, including Rancher Desktop's `resources/resources/win32/bin` and Chocolatey, so `uv run inv k8s.*` and Core image builds can find `docker`, `kubectl`, and related CLIs from fresh shells.
- **External AI endpoint contract**: `k8s.deploy` accepts `MYCELIS_K8S_TEXT_ENDPOINT`, optional `MYCELIS_K8S_TEXT_MODEL_ID`, and `MYCELIS_K8S_MEDIA_ENDPOINT`, forwarding them into Helm so deployed providers can target a reachable external AI host without editing chart source.
- **External search endpoint contract**: `k8s.deploy` accepts `MYCELIS_K8S_SEARCH_PROVIDER`, `MYCELIS_K8S_SEARXNG_ENDPOINT`, and `MYCELIS_K8S_SEARCH_LOCAL_API_ENDPOINT`; configured online search is notify-only/no-confirm and responses disclose provider/path plus external-result interpretation.
- **Preset values contract**: `k8s.deploy` also accepts `MYCELIS_K8S_VALUES_FILE`; repo-relative paths such as `charts/mycelis-core/values-k3d.yaml` are resolved from the repo root and the task fails fast if the requested values file does not exist.
- **Verification packaging**: `k8s.deploy --verify-package` is the release-packaging path for promoted values files; it renders, lints, packages, and writes manifest/checksum artifacts under `dist/helm/` without contacting a cluster.
- **Open-standard surfaces**: the standards gate keeps the chart on standard Kubernetes resources: Deployment, Service, ServiceAccount, Secret, ConfigMap, PVC, Ingress, NetworkPolicy, probes, non-root pod/container security context, image pull/digest posture, and cluster-managed output storage.
- **Enterprise Windows-AI guardrail**: the `values-enterprise-windows-ai` preset fails closed unless `MYCELIS_K8S_TEXT_ENDPOINT` is set to the real Windows GPU host endpoint; the preset also declares `qwen3:8b` as the reviewable self-hosted model default and enables explicit AI egress.
- Chart/runtime config alignment: the deployed Core image resolves startup config from `/core/config`, so the chart mount path and container workdir must stay in sync for bootstrap bundles to load.

### `compose.py` (Home Runtime)
Handles the rapid Docker Compose single-host runtime for development, same-machine proof, home-lab experiments, and demo use. Compose is not the target clustered deployment contract.
- **Infra Up**: `uv run inv compose.infra-up` (postgres + nats only, Core/Interface stay down, readiness checks + owner-facing connection settings; add `--migrate` only when schema bootstrap is intentionally needed)
- **Infra Health**: `uv run inv compose.infra-health` (PostgreSQL port/query readiness, NATS port, and NATS monitor only; no Core/UI health checks)
- **Storage Health**: `uv run inv compose.storage-health` (post-migration PostgreSQL long-term storage gate for pgvector, semantic context vectors, durable memory, conversation continuity, artifacts, managed exchange, collaboration groups, and templates)
- **Warm Cognitive**: `uv run inv compose.warm-cognitive` (warms the configured Compose text model through the same Ollama endpoint Core uses before live browser proof)
- **Up**: `uv run inv compose.up` (postgres + nats -> migrate -> core + interface, with numbered stage output and optional `--wait-timeout=<seconds>`)
- Compose `up` and `migrate` now behave like the main `db.migrate` contract: they bootstrap forward only when the compose `cortex` schema is compatible through the V8.2 capability/proof/team-work tables and collaboration-group workspace-folder schema, and they point to `uv run inv compose.down --volumes` for a truly fresh replay.
- **Down**: `uv run inv compose.down`
- **Health**: `uv run inv compose.health`
- **Status**: `uv run inv compose.status`
- **Logs**: `uv run inv compose.logs`
- Compose uses `.env.compose` so host/container assumptions stay separate from the local-Kubernetes `.env` path.
- Compose uses `MYCELIS_COMPOSE_OLLAMA_HOST` instead of raw `OLLAMA_HOST` so host-machine Ollama bind settings cannot override the container runtime accidentally, and maps that value into provider-specific endpoint overrides inside Core.
- Compose passes `MYCELIS_SEARCH_PROVIDER`, `MYCELIS_SEARXNG_ENDPOINT`, `MYCELIS_SEARCH_LOCAL_API_ENDPOINT`, and `MYCELIS_SEARCH_MAX_RESULTS` into Core for governed search; use `local_sources` for token-free governed local-source search, `searxng` for operator-owned metasearch, or `local_api` for an operator-owned HTTP search endpoint instead of treating Brave as mandatory.
- Compose rejects loopback compose Ollama values because `localhost`, `127.0.0.1`, and `0.0.0.0` point back at the Core container instead of the operator host.
- Compose `infra-up` is the data-plane-only preflight for personal-owner deployments where PostgreSQL/NATS should be reachable before app services are launched.
- Compose `storage-health` is the matching post-migration proof that the long-term Mycelis Postgres store is present before claiming RAG, retained outputs, or continuity are available.
- Compose `migrate` skips unsafe full replay on compatible volumes but still applies known missing late storage migrations so `storage-health` can close the long-term store gate.
- Compose validates output block mounting: use `MYCELIS_OUTPUT_BLOCK_MODE=local_hosted` plus `MYCELIS_OUTPUT_HOST_PATH=<host-directory>` when a local or Pinokio/media-hosted output directory should be mounted into Core as `/data`. The task resolves the host path with Python `pathlib` across Windows, Linux, and macOS before Docker starts.
- Runtime output roots inside Core are explicit: `MYCELIS_WORKSPACE` is where generated files, project packages, browser games, and filesystem MCP writes land; `MYCELIS_ARTIFACT_ROOT` is where file-backed artifact/cache data lands. `DATA_DIR` is still honored as a legacy artifact-root alias and should match `MYCELIS_ARTIFACT_ROOT` in Compose proof.
- Compose defaults `MYCELIS_WORKSPACE_REVEAL_DRY_RUN=1` because Core runs inside a container. Browser `Storage` controls should still return a trusted mounted workspace path during proof, while native Core or a host-action bridge can opt into actual desktop folder opening.
- On Windows, when native `docker` is absent but WSL Docker is available, `compose.*` uses the WSL Docker CLI automatically and translates compose/output-block paths for that runtime. Set `MYCELIS_WSL_DISTRO` when the Docker-owning distro is not the default.
- On that WSL Docker path, `compose.up` and `compose.health` can auto-start a small restartable WSL-host relay for `MYCELIS_COMPOSE_OLLAMA_HOST`, including when you run the tasks directly inside the Docker-owning WSL distro, so the Core container can still reach a Windows-hosted Ollama service through `host.docker.internal` when bridge containers cannot dial the Windows LAN IP directly.
- Compose now emits deterministic stage expectations and timeout guidance so humans and agent-run callers can tell what should happen next and which recovery command to run if a stage stalls.
- Use `uv run inv compose.up --build --wait-timeout=240` on a fresh or slower machine when image build and first readiness can legitimately take longer than the default window.
- `compose.health` is a usable-product gate for the home runtime, so it fails when text inference is offline even if the API still responds.
- The compose Core image includes Node/npm/npx so manual curated stdio MCP installs can launch from the shipped container; default npm-backed MCP auto-bootstrap still stays disabled by default to keep startup logs honest.
- Manual `filesystem` installs from the curated library are runtime-normalized to the configured `MYCELIS_WORKSPACE` root, which is `/data/workspace` in the supported Compose output block.

### `native_infra.py` (Source-Mode Data Plane)
Owns the narrow Windows/source-mode dependency path for development and testing without Docker.
- **Install NATS**: `uv run inv native-infra.install-nats` installs `nats-server` with Go into the local Go bin directory. The pinned default is `v2.14.0`.
- **Up**: `uv run inv native-infra.up` bootstraps the configured PostgreSQL app role/database and starts local NATS with JetStream plus the HTTP monitor.
- **Status**: `uv run inv native-infra.status` checks PostgreSQL port/query readiness, NATS port readiness, and the NATS monitor endpoint.
- **Bootstrap PostgreSQL**: `uv run inv native-infra.bootstrap-postgres` uses `POSTGRES_USER` / `POSTGRES_PASSWORD` from `.env` to create or update `DB_USER`, `DB_PASSWORD`, and `DB_NAME`.
- **Start NATS**: `uv run inv native-infra.start-nats` starts only NATS when PostgreSQL is already ready.
- `lifecycle.up` defaults to `MYCELIS_DEV_INFRA_MODE=native`, so it no longer tries Kubernetes port-forwards during ordinary source-mode development. Set `MYCELIS_DEV_INFRA_MODE=k8s` only for explicit clustered bridge proof.

### `core.py` (Compilation)
Handles Go compilation and Docker image building.
- **Compile**: `uv run inv core.compile` (repo-local binary only).
- **Package**: `uv run inv core.package` (versioned cross-target binary archive under `dist/`, plus manifest/checksum sidecars).
- **Build**: `uv run inv core.build` (Returns immutable image tag; no `latest` aliasing).

### `auth.py` (Local Operator Auth)
Keeps local API-key development access aligned.
- **Dev Key**: `uv run inv auth.dev-key`
- **Break-Glass Key**: `uv run inv auth.break-glass-key`
- **Posture**: `uv run inv auth.posture`

### `mycelis_api.py` (Self-Use API Proof)
Uses the live Core API as a delivery tool instead of only testing it from the outside.
- **Delivery Proof**: `uv run inv api.delivery-proof`
- The task refreshes capabilities, reads System -> Deployments trust, creates a bounded `TeamWorkItem`, attaches a `TeamInteraction`, and reads the work item back.
- It defaults to `MYCELIS_API_BASE_URL` or `MYCELIS_API_HOST`/`MYCELIS_API_PORT` from the shell or repo `.env`, and uses `MYCELIS_API_KEY` from the same sources.

### `cache.py` (Cache Hygiene)
Keeps project and user-level tool caches off the system drive hot path and easy to prune.
- **Status**: `uv run inv cache.status`
- **Guard**: `uv run inv cache.guard`
- **Clean**: `uv run inv cache.clean`
- **Clean User Cache Too**: `uv run inv cache.clean --user`
- **Apply Windows User Policy**: `uv run inv cache.apply-user-policy`
- Project-owned backstops: root `.npmrc` keeps direct npm/npx cache local to `workspace/tool-cache`, pytest cache metadata lives in `workspace/tool-cache/pytest`, and task-managed browser runs export `PLAYWRIGHT_BROWSERS_PATH`
- Suggested platform posture: on Windows, stamp the user-level cache env vars early if `C:` is the small drive; on Linux/macOS, move project/user cache roots only when the default workspace or home volume is the wrong place for repeated build churn
- Heavy repo-managed build/test paths now run a disk-headroom preflight before large local churn; that guard covers the repo/cache volume and leaves Docker daemon / WSL image-layer storage as a separate budget to monitor.

### `logging.py` (Logging Gates)
Enforces logging contract quality checks before delivery.
- **Schema**: `uv run inv logging.check-schema` (event taxonomy + docs coverage)
- **Topics**: `uv run inv logging.check-topics` (no hardcoded `swarm.*` outside constants)

### `quality.py` (Code Hygiene Gates)
Enforces max-lines policy across the main source tree with temporary no-regression caps for legacy oversized files. Stale cap entries for deleted files fail the gate so cleanup cannot leave old exceptions behind.
- **Max Lines**: `uv run inv quality.max-lines --limit 300`

### `lifecycle.py` (Local Stack Control)
Owns deterministic local bring-up, teardown, and deep health checks.
- **Up**: `uv run inv lifecycle.up --frontend`
- **Down**: `uv run inv lifecycle.down`
- **Health**: `uv run inv lifecycle.health`
- **Memory Restart**: `uv run inv lifecycle.memory-restart --frontend`
- `lifecycle.up` now ensures the `cortex` database exists before Core starts, so the bootstrap listener does not crash when a fresh bridge comes up after a reboot or cluster reset
- `lifecycle.down` now treats repo-local Interface worker residue as part of the teardown contract, not just bound ports
- `lifecycle.status` reports a quick service snapshot and validates Core through `/healthz` plus Ollama through `/api/tags` across accepted loopback hosts so endpoint-reachable services are not reported down from a single TCP miss
- `lifecycle.health` and `compose.health` allow the cognitive status endpoint a longer client window than its internal provider probes, so slow or degraded local AI endpoints produce operator-readable health evidence instead of edge timeouts
- local tasking targets the bridged Core API port by default (`localhost:8081` unless `MYCELIS_API_PORT` overrides it)
- local Interface tasking now binds the UI broadly by default (`[::]:3000`) while probing it through `127.0.0.1:3000`; override with `MYCELIS_INTERFACE_BIND_HOST`, `MYCELIS_INTERFACE_HOST`, and `MYCELIS_INTERFACE_PORT` when a host needs a different split
- frontend teardown falls back to matching both `next dev` and `next start` command lines, so built UI servers do not survive outside the lifecycle contract

### `interface.py` + `interface_runtime.py` (Frontend Build And Browser Tasks)
- **Install**: `uv run inv interface.install`
- `interface.install` now provisions npm dependencies plus the managed Playwright Chromium binary used by `interface.e2e`
- **Type Check**: `uv run inv interface.typecheck`
- **Build**: `uv run inv interface.build`
- **Test**: `uv run inv interface.test` (Vitest runs test files sequentially for deterministic full-suite proof)
- **E2E**: `uv run inv interface.e2e`
- **Broad managed UI certification**: `uv run inv interface.e2e --server-mode=start --project=chromium --workers=1`
- `ops/interface.py` is the stable Invoke entrypoint; the task implementation lives in `ops/interface_runtime.py`, with shared command/env helpers in `ops/interface_env.py` and process matching hints in `ops/interface_process_support.py`
- `uv run inv interface.e2e` now defaults to managed `dev` mode for stable mocked browser proof. Use `--server-mode=start` when you need the built `next start` path for stricter or live-backend proof; `uv run inv interface.build` still retries once after stale repo-local Next build locks, stale `.next/standalone` cleanup locks, incomplete built-server packaging, or transient missing `.next/types` output before failing, start-mode E2E inherits that same recovery behavior, Windows managed Playwright servers bind to `127.0.0.1` for stable loopback readiness/browser navigation, and managed `dev` runs clear an orphaned `interface/.next/dev/lock` only when no repo-local Next worker remains.
- Live backend browser specs that assert filesystem side effects now infer `MYCELIS_BACKEND_WORKSPACE_ROOT` from the loaded `.env`/process `MYCELIS_WORKSPACE` for native Core: absolute roots are used directly, while repo-local `./workspace` maps to `core/workspace`. Set `MYCELIS_BACKEND_WORKSPACE_ROOT` or `PLAYWRIGHT_BACKEND_WORKSPACE_ROOT` only when the spec checkout and the running Core checkout differ, such as a supported Compose host workspace under `workspace/docker-compose/data/workspace`. For K8s/PVC proof, set `PLAYWRIGHT_BACKEND_WORKSPACE_PROBE=k8s` plus `PLAYWRIGHT_K8S_NAMESPACE`, `PLAYWRIGHT_K8S_CORE_SELECTOR`, and `PLAYWRIGHT_K8S_BACKEND_WORKSPACE_ROOT` so the live spec checks the Core pod workspace with `kubectl`.
- Runtime file/tool requests may also use `workspace/...` as a friendly alias for the configured workspace root; the backend normalizes that prefix away so Compose-backed `/data/workspace` and repo-local `./workspace` do not produce doubled `workspace/workspace/...` paths.

### `test.py` (Root Test Aliases)
- **All**: `uv run inv test.all`
- **Coverage**: `uv run inv test.coverage`
- **E2E Alias**: `uv run inv test.e2e`
- `uv run inv test.e2e` mirrors the same managed browser options as `uv run inv interface.e2e`, including `--workers` and `--server-mode=dev|start|external`.

### `cognitive.py` (Optional Local Engine Helpers)
- **Install**: `uv run inv cognitive.install`
- **LLM**: `uv run inv cognitive.llm`
- **Media**: `uv run inv cognitive.media`
- **Media Gateway**: `uv run inv cognitive.media-gateway`
- **Up**: `uv run inv cognitive.up`
- **Stop**: `uv run inv cognitive.stop`
- **Status**: `uv run inv cognitive.status`
- These are optional local helpers for vLLM/Diffusers experimentation, not part of the supported default Core + Interface runtime contract.
- The repo-local vLLM/Diffusers helper lane is intended for supported Linux GPU hosts; on Windows, use Ollama locally for text and `cognitive.media-gateway` for Pinokio-hosted Forge/AUTOMATIC1111 or ComfyUI media generation.

## Clean Run Discipline for Runtime and Integration Checks

- Before any runtime or integration-style test, stop prior local services using the repo lifecycle task path. Use `uv run inv lifecycle.down` unless a narrower repo task is the safer equivalent for the slice.
- For the supported home-runtime stack, `uv run inv compose.down --volumes` is the clean reset path before runtime/browser proof.
- Verify ports and processes are clear for the services involved in the check. At minimum review the Core API port, NATS, PostgreSQL, and Ollama when the slice depends on them, using repo ops tasks such as `uv run inv lifecycle.status` or OS-level port/process tools.
- Detect running compiled Go services before the test begins. Check repo-local command lines or binary paths plus any processes bound to declared dev/test ports; if found, terminate them with the lifecycle/task helpers and never assume they belong to the current run.
- Detect repo-local Interface worker residue before and after browser/build/test runs. Windows `node.exe` children from `.next`, Vitest, or Playwright count as leaked dev state and should be swept by the task wrappers.
- Treat browser proof as a stability-sensitive path by default. `uv run inv interface.e2e` now uses managed `dev` mode and `--workers=1` for stable mocked browser proof unless you explicitly switch to `--server-mode=start`, which refreshes the production bundle before the managed server runs and keeps the stricter or live-backend path aligned with the same cleanup/retry behavior. On Windows, managed browser proof binds the UI server to IPv4 loopback (`127.0.0.1`) even when normal dev servers may bind more broadly. `uv run inv ci.baseline` keeps the same low-parallelism posture, and `uv run inv ci.service-check --live-backend` stays at `--workers=1` while restoring the local bridge/core stack before the live proof when needed. The managed browser task must now own a clean UI server for the run; if the managed Next process exits or a stale listener blocks the port, the task fails instead of silently borrowing the wrong server. Keep managed `interface.build`, `interface.test`, and managed `interface.e2e` serial for a workspace and port because they share Next/Vitest workers and server ownership. The live service check now reuses an already-initialized `cortex` schema instead of replaying non-idempotent migrations on every run, so the managed gate prefers repeatable results over peak local parallelism without turning database state into false failure noise.
- `uv run inv interface.check` retries transient Windows socket-reuse errors after heavy Playwright runs before failing a route check; persistent HTTP/SSR/runtime page failures still fail immediately.
- `uv run inv db.migrate` is intentionally a forward-bootstrap helper now, not a replay-everything hammer. Once the schema is compatible through the V8.2 capability/proof/team-work tables and collaboration-group workspace-folder schema, it skips replay and points operators to `uv run inv db.reset` for a clean rebuild path. For fresh Soma/team proof without rebuilding schema, first run `uv run inv db.clear-runtime-context` to inspect volatile context counts, then add `--yes` when stale conversations, team work, run/proof handshakes, or temp memory should be cleared; long-memory vectors are kept unless `--include-memory-vectors` is supplied.
- Start only the minimal services required for the specific check. Prefer the narrowest path that matches the validation target, such as Helm render only, bootstrap/unit coverage only, Core-only, or a bounded local stack bring-up.
- Run the test or validation command once the required services are confirmed ready.
- Shut services down immediately after the check unless the slice explicitly requires them left running for a follow-on validation step.
- Agents must never stack runs on top of unknown existing processes.

### Compiled Go Service Cleanup Before Tests

Locally built or manually started Go binaries can survive after the test that launched them. `go build`, `go run`, and direct `core/bin/*` execution all count as part of the cleanup surface even when containers and bridges are already down.

Typical binaries/process shapes to clear:
- core server
- relays / bridges when they run as Go services
- bootstrap helpers
- any Go-based local services used by the repo, including `go run ./cmd/server`, `go run ./cmd/probe`, `go run ./cmd/signal_gen`, and similar helper processes

Cleanup must cover:
- local compiled Go services
- containerized dependencies
- test-managed ephemeral services

Stopping containers alone is not enough. The cleanup pass must also inspect and terminate stray compiled Go binaries before the next runtime or integration test. If process inspection fails, treat the environment as unverified and stop before running the check.

### `ci.py` (Delivery Gates)
Delivery-focused validation, runner checks, and release preflight.
- **Test**: `uv run inv ci.test` (Go tests + blocking Vitest run)
- **Entrypoint Check**: `uv run inv ci.entrypoint-check`
- **Baseline**: `uv run inv ci.baseline` (includes Playwright by default; use `--no-e2e` only for intentionally narrower local debugging)
- **Service Check**: `uv run inv ci.service-check --live-backend`
- **Release Preflight**: `uv run inv ci.release-preflight --lane=release`
- **Lane presets**: `baseline`, `runtime`, `service`, `release` (legacy flags still supported for custom proof)
- **Runtime Posture Gate**: `--runtime-posture` adds a 12 GiB disk-headroom check, reads explicit AI endpoints from process env plus `.env.compose` / `.env`, fails closed when no supported non-loopback endpoint contract is configured before baseline proof runs, and mirrors `host.docker.internal` through WSL localhost when the probe itself is running inside WSL before Compose relay startup.
- Interface-facing CI steps now perform the same repo-local worker cleanup after `build`, `tsc`, `vitest`, and Playwright runs, and they execute from the `interface/` working directory so Windows and Linux share the same `npm`/`node` task path
- GitHub validation workflows should keep dependency/bootstrap steps workflow-native (`actions/setup-*`, `npm ci`, Playwright browser install), then hand real build/test execution back to the same `uv run inv ...` task surfaces so local and CI validation stay aligned
- Hosted GitHub validation uses Node 24-capable action majors and Node.js 24 for Interface lanes. Helm setup is a checksum-verified pinned Helm 3 download instead of `azure/setup-helm@v4` because that action still runs on Node 20.
- GitHub CI is manual-only and offers targeted repo hygiene, Core, Interface, authenticated browser smoke, and Helm standards lanes; live agentry/runtime proof remains in local/WSL release lanes unless the manual Source API Proof or Full Release Candidate workflow is explicitly dispatched.
- Delivery reporting should include the commands run plus the docs changed and the touched docs reviewed unchanged.

### `misc.py` (Team Coordination)
Central architect sync path and utility task surfaces.
- **Architecture Sync**: `uv run inv team.architecture-sync`
- **Worktree Triage**: `uv run inv team.worktree-triage`

## Directives
- `uv run inv interface.build` retries once after stale repo-local Next build locks, stale `.next/standalone` cleanup locks, incomplete built-server packaging, or transient missing `.next/types` output before failing so Windows cleanup residue and generated-output races do not masquerade as product regressions.
- **Never tag `latest`** for production.
- **Always pinning** dependencies in `charts/`.
- **Prefer repo-managed caches** under `workspace/tool-cache` for Invoke-driven work so local validation does not silently refill `C:`.
- **Do not land implementation drift against the docs stack.** Review and update README, state, testing/ops docs, API reference, and owning canonical or user docs whenever the slice changes their meaning.
