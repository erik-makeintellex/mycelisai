# Operations & Build System
> **Role**: The Builder
> **Language**: Python (Invoke)
> **Path**: `ops/`

## Runner Contract

- Real task execution uses `uv run inv <namespace>.<task>`.
- Compatibility probe only: `uvx --from invoke inv -l`.
- Do not use bare `uvx inv ...`.
- App-tied management logic stays in Python task modules; PowerShell is wrapper-only when the host needs it.

## 🛠️ Components
This directory contains the logic for the **Service Release Standard 1.0**.

### `version.py` (Identity)
Calculates the **Immutable Tag**: `v{SEMVER}-{SHA}`.
- Source: `../VERSION` file.
- Git: `git rev-parse --short HEAD`.

### `k8s.py` (Deployment)
Handles the atomic deployment to Kubernetes (Kind).
- **Init**: `uv run inv k8s.init` (Infra).
- **Deploy**: `uv run inv k8s.deploy` (Core).
- **Status**: `uv run inv k8s.status` (Health).
- Chart/runtime config alignment: the deployed Core image resolves startup config from `/core/config`, so the chart mount path and container workdir must stay in sync for bootstrap bundles to load.

### `core.py` (Compilation)
Handles Go compilation and Docker image building.
- **Build**: `uv run inv core.build` (Returns Tag).

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

## Clean Run Discipline for Runtime and Integration Checks

- Before any runtime or integration-style test, stop prior local services using the repo lifecycle task path. Use `uv run inv lifecycle.down` unless a narrower repo task is the safer equivalent for the slice.
- Verify ports and processes are clear for the services involved in the check. At minimum review the Core API port, NATS, PostgreSQL, and Ollama when the slice depends on them, using repo ops tasks such as `uv run inv lifecycle.status` or OS-level port/process tools.
- Detect running compiled Go services before the test begins. Check repo-local command lines or binary paths plus any processes bound to declared dev/test ports; if found, terminate them with the lifecycle/task helpers and never assume they belong to the current run.
- Detect repo-local Interface worker residue before and after browser/build/test runs. Windows `node.exe` children from `.next`, Vitest, or Playwright count as leaked dev state and should be swept by the task wrappers.
- Treat merge-readiness browser proof as a stability-sensitive path. `uv run inv ci.baseline` now calls Playwright with `--workers=1` against the built `next start` server, and `uv run inv ci.service-check --live-backend` keeps `--workers=1` while restoring the local bridge/core stack before the live proof when needed. The live service check now reuses an already-initialized `cortex` schema instead of replaying non-idempotent migrations on every run, so the managed gate prefers repeatable results over peak local parallelism without turning database state into false failure noise.
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

### `misc.py` (Team Coordination)
Central architect sync path and utility task surfaces.
- **Architecture Sync**: `uv run inv team.architecture-sync`

## ⚡ Directives
- **Never tag `latest`** for production.
- **Always pinning** dependencies in `charts/`.
- **Prefer repo-managed caches** under `workspace/tool-cache` for Invoke-driven work so local validation does not silently refill `C:`.
