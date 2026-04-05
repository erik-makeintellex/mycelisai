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

## Components
This directory contains the logic for the **Service Release Standard 1.0**.

### Root `install` Task
`uv run inv install` now installs the supported default Core + Interface stack only.

- Use `uv run inv install --optional-engines` when you also want the local `cognitive/` extras.
- Use `uv run inv cognitive.install` if you want only the optional local engine dependencies.

### `version.py` (Identity)
Calculates the **Immutable Tag**: `v{SEMVER}-{SHA}`.
- Source: `../VERSION` file.
- Git: `git rev-parse --short HEAD`.

### `k8s.py` (Deployment)
Handles the atomic deployment to Kubernetes (Kind).
- **Init**: `uv run inv k8s.init` (Infra).
- **Deploy**: `uv run inv k8s.deploy` (Core).
- **Status**: `uv run inv k8s.status` (Health).
- **Bridge**: `uv run inv k8s.bridge` now verifies the local PostgreSQL/NATS/Core port-forwards actually bind before reporting success.
- **Recover**: `uv run inv k8s.recover` now fails closed when the cluster is unreachable and waits for rollout readiness before claiming recovery.
- Chart/runtime config alignment: the deployed Core image resolves startup config from `/core/config`, so the chart mount path and container workdir must stay in sync for bootstrap bundles to load.

### `compose.py` (Home Runtime)
Handles the supported Docker Compose single-host runtime for home-lab and demo use.
- **Up**: `uv run inv compose.up` (postgres + nats -> migrate -> core + interface)
- Compose `up` and `migrate` now behave like the main `db.migrate` contract: they bootstrap forward only when the compose `cortex` schema is not already compatible with the current runtime, and they point to `uv run inv compose.down --volumes` for a truly fresh replay.
- **Down**: `uv run inv compose.down`
- **Health**: `uv run inv compose.health`
- **Status**: `uv run inv compose.status`
- **Logs**: `uv run inv compose.logs`
- Compose uses `.env.compose` so host/container assumptions stay separate from the Kind/bridge `.env` path.
- Compose uses `MYCELIS_COMPOSE_OLLAMA_HOST` instead of raw `OLLAMA_HOST` so host-machine Ollama bind settings cannot override the container runtime accidentally.
- Compose rejects loopback compose Ollama values because `localhost`, `127.0.0.1`, and `0.0.0.0` point back at the Core container instead of the operator host.
- `compose.health` is a usable-product gate for the home runtime, so it fails when text inference is offline even if the API still responds.
- The slim compose Core image disables default npm-backed MCP auto-bootstrap by default to keep startup logs honest; manual/external MCP connectivity remains supported.

### `core.py` (Compilation)
Handles Go compilation and Docker image building.
- **Compile**: `uv run inv core.compile` (repo-local binary only).
- **Build**: `uv run inv core.build` (Returns immutable image tag; no `latest` aliasing).

### `auth.py` (Local Operator Auth)
Keeps local API-key development access aligned.
- **Dev Key**: `uv run inv auth.dev-key`

### `cache.py` (Cache Hygiene)
Keeps project and user-level tool caches off the system drive hot path and easy to prune.
- **Status**: `uv run inv cache.status`
- **Clean**: `uv run inv cache.clean`
- **Clean User Cache Too**: `uv run inv cache.clean --user`
- **Apply Windows User Policy**: `uv run inv cache.apply-user-policy`
- Project-owned backstops: root `.npmrc` keeps direct npm/npx cache local to `workspace/tool-cache`, pytest cache metadata lives in `workspace/tool-cache/pytest`, and task-managed browser runs export `PLAYWRIGHT_BROWSERS_PATH`
- Suggested platform posture: on Windows, stamp the user-level cache env vars early if `C:` is the small drive; on Linux/macOS, move project/user cache roots only when the default workspace or home volume is the wrong place for repeated build churn

### `logging.py` (Logging Gates)
Enforces logging contract quality checks before delivery.
- **Schema**: `uv run inv logging.check-schema` (event taxonomy + docs coverage)
- **Topics**: `uv run inv logging.check-topics` (no hardcoded `swarm.*` outside constants)

### `quality.py` (Code Hygiene Gates)
Enforces max-lines policy on hot-path files with temporary no-regression caps.
- **Max Lines**: `uv run inv quality.max-lines --limit 350`

### `lifecycle.py` (Local Stack Control)
Owns deterministic local bring-up, teardown, and deep health checks.
- **Up**: `uv run inv lifecycle.up --frontend`
- **Down**: `uv run inv lifecycle.down`
- **Health**: `uv run inv lifecycle.health`
- **Memory Restart**: `uv run inv lifecycle.memory-restart --frontend`
- `lifecycle.up` now ensures the `cortex` database exists before Core starts, so the bootstrap listener does not crash when a fresh bridge comes up after a reboot or cluster reset
- `lifecycle.down` now treats repo-local Interface worker residue as part of the teardown contract, not just bound ports
- local tasking targets the bridged Core API port by default (`localhost:8081` unless `MYCELIS_API_PORT` overrides it)
- local Interface tasking now binds the UI broadly by default (`[::]:3000`) while probing it through `127.0.0.1:3000`; override with `MYCELIS_INTERFACE_BIND_HOST`, `MYCELIS_INTERFACE_HOST`, and `MYCELIS_INTERFACE_PORT` when a host needs a different split
- frontend teardown falls back to matching both `next dev` and `next start` command lines, so built UI servers do not survive outside the lifecycle contract

### `interface.py` (Frontend Build And Browser Tasks)
- **Install**: `uv run inv interface.install`
- **Type Check**: `uv run inv interface.typecheck`
- **Build**: `uv run inv interface.build`
- **Test**: `uv run inv interface.test`
- **E2E**: `uv run inv interface.e2e`
- `uv run inv interface.e2e` now defaults to managed `dev` mode for stable mocked browser proof. Use `--server-mode=start` when you need the built `next start` path for stricter or live-backend proof; `uv run inv interface.build` still retries once after a stale repo-local Next build lock before failing, and start-mode E2E inherits that same recovery behavior.
- Live backend browser specs that assert filesystem side effects may need `MYCELIS_BACKEND_WORKSPACE_ROOT` (or `PLAYWRIGHT_BACKEND_WORKSPACE_ROOT`) when the spec checkout and the running Core checkout differ; use the backend's actual workspace root, such as `core/workspace` for a repo-local Core process or `workspace/docker-compose/data/workspace` for the supported compose stack.

### `cognitive.py` (Optional Local Engine Helpers)
- **Install**: `uv run inv cognitive.install`
- **LLM**: `uv run inv cognitive.llm`
- **Media**: `uv run inv cognitive.media`
- **Up**: `uv run inv cognitive.up`
- **Stop**: `uv run inv cognitive.stop`
- **Status**: `uv run inv cognitive.status`
- These are optional local helpers for vLLM/Diffusers experimentation, not part of the supported default Core + Interface runtime contract.
- The repo-local `cognitive.*` helper lane is intended for supported Linux GPU hosts; on Windows, use Ollama locally or point `core/config/cognitive.yaml` at a remote OpenAI-compatible vLLM server instead.

## Clean Run Discipline for Runtime and Integration Checks

- Before any runtime or integration-style test, stop prior local services using the repo lifecycle task path. Use `uv run inv lifecycle.down` unless a narrower repo task is the safer equivalent for the slice.
- For the supported home-runtime stack, `uv run inv compose.down --volumes` is the clean reset path before runtime/browser proof.
- Verify ports and processes are clear for the services involved in the check. At minimum review the Core API port, NATS, PostgreSQL, and Ollama when the slice depends on them, using repo ops tasks such as `uv run inv lifecycle.status` or OS-level port/process tools.
- Detect running compiled Go services before the test begins. Check repo-local command lines or binary paths plus any processes bound to declared dev/test ports; if found, terminate them with the lifecycle/task helpers and never assume they belong to the current run.
- Detect repo-local Interface worker residue before and after browser/build/test runs. Windows `node.exe` children from `.next`, Vitest, or Playwright count as leaked dev state and should be swept by the task wrappers.
- Treat browser proof as a stability-sensitive path by default. `uv run inv interface.e2e` now uses managed `dev` mode and `--workers=1` for stable mocked browser proof unless you explicitly switch to `--server-mode=start`, which refreshes the production bundle before the managed server runs and keeps the stricter or live-backend path aligned with the same cleanup/retry behavior. `uv run inv ci.baseline` keeps the same low-parallelism posture, and `uv run inv ci.service-check --live-backend` stays at `--workers=1` while restoring the local bridge/core stack before the live proof when needed. The managed browser task must now own a clean UI server for the run; if the managed Next process exits or a stale listener blocks the port, the task fails instead of silently borrowing the wrong server. The live service check now reuses an already-initialized `cortex` schema instead of replaying non-idempotent migrations on every run, so the managed gate prefers repeatable results over peak local parallelism without turning database state into false failure noise.
- `uv run inv db.migrate` is intentionally a forward-bootstrap helper now, not a replay-everything hammer. Once the schema is already compatible with the current runtime, it skips replay and points operators to `uv run inv db.reset` for a clean rebuild path.
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
- **Release Preflight**: `uv run inv ci.release-preflight --strict-toolchain --service-health --live-backend`
- Interface-facing CI steps now perform the same repo-local worker cleanup after `build`, `tsc`, `vitest`, and Playwright runs, and they execute from the `interface/` working directory so Windows and Linux share the same `npm`/`node` task path
- GitHub validation workflows should keep dependency/bootstrap steps workflow-native (`actions/setup-*`, `npm ci`, Playwright browser install), then hand real build/test execution back to the same `uv run inv ...` task surfaces so local and CI validation stay aligned
- Push-triggered GitHub pipeline runs are intentionally paused until the initial release-ready gate reopens; use local `uv run inv ...` proof plus PR/manual workflow runs as the active validation path

### `misc.py` (Team Coordination)
Central architect sync path and utility task surfaces.
- **Architecture Sync**: `uv run inv team.architecture-sync`
- **Worktree Triage**: `uv run inv team.worktree-triage`

## Directives
- **Never tag `latest`** for production.
- **Always pinning** dependencies in `charts/`.
- **Prefer repo-managed caches** under `workspace/tool-cache` for Invoke-driven work so local validation does not silently refill `C:`.
