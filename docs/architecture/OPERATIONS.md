# Mycelis Cortex — Operations Manual
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> **Load this doc when:** Deploying, configuring, testing, running CI, or managing infrastructure.
>
> **Related:** [Overview](OVERVIEW.md) | [Backend](BACKEND.md) | [Frontend](FRONTEND.md)

## TOC

- [I. Prerequisites](#i-prerequisites)
- [II. Task Automation](#ii-task-automation)
- [III. Development Workflow](#iii-development-workflow)
- [IV. Configuration System](#iv-configuration-system)
- [V. Testing Strategy](#v-testing-strategy)
- [VI. CI/CD](#vi-cicd)
- [VII. Deployment Architecture](#vii-deployment-architecture)
- [VIII. Environment Gotchas](#viii-environment-gotchas)
- [IX. Monitoring & Observability](#ix-monitoring--observability)

---

## I. Prerequisites

| Tool | Version | Purpose |
|------|---------|---------|
| [uv](https://github.com/astral-sh/uv) | latest | Python package manager (replaces pip/pipx) |
| [Docker](https://www.docker.com/) | latest | Container runtime |
| [Go](https://go.dev/) | 1.26 | Backend compilation |
| [Node.js](https://nodejs.org/) | >=20 | Frontend build |
| [kubectl](https://kubernetes.io/docs/tasks/tools/) | >=1.35 | Kubernetes CLI (1.32 has port-forward bug with K8s 1.33) |

**Cross-platform mandate:** All tasks must work on Windows AND Linux. Use `is_windows()` from `ops/config.py` for OS branching. Never use bash/cmd wrapper scripts.
**Management scripting mandate:** App-tied management logic belongs in Python task modules. PowerShell is only a thin local wrapper when the host platform requires it.

---

## II. Task Automation

### Master Registry

**File:** `tasks.py` — 18 namespaces, 78 exported tasks in the root invoke surface

**Run from:** `scratch/` (project root where tasks.py lives)

**Invocation:** Always use `uv run inv <namespace>.<task>` for real task execution.
Use `uvx --from invoke inv -l` only as a lightweight compatibility probe.
Do not use bare `uvx inv ...`.
Lifecycle tasks must not report success until Core `/healthz` is actually ready; open-port-only checks are insufficient.
Background Core startup must write to `workspace/logs/core-startup.log` so lifecycle failures have a deterministic diagnostic surface.
Lifecycle teardown must use bounded cleanup subprocesses, sweep repo-local Interface worker residue, and wait for ports to close before reporting success.
Interface-focused Invoke and CI tasks must execute from the `interface/` working directory and reuse the same `npm`/`node` entrypoints on Windows and Linux rather than depending on host-shell `cd ... &&` wrappers.
Implementation slices that change runtime, tasking, validation, API meaning, or operator behavior must review and update the owning docs in the same change rather than leaving docs drift for later cleanup.

> **Tip:** After activating the project virtualenv yourself, `inv` can still be used as a convenience, but the canonical repo contract remains `uv run inv ...`.

### Core Tasks (`ops/core.py`)

| Command | Description |
|---------|-------------|
| `uv run inv core.compile` | Compile the repo-local Go binary only |
| `uv run inv core.build` | Compile Go binary + Docker image (returns immutable TAG) |
| `uv run inv core.test` | `go test ./...` |
| `uv run inv core.run` | Run Core locally (foreground, blocking) |
| `uv run inv core.stop` | Kill running Core process (taskkill on Windows, pkill on Linux) |
| `uv run inv core.restart` | stop + run |
| `uv run inv core.smoke` | `go run ./cmd/smoke/main.go` (governance smoke tests) |
| `uv run inv core.package` | Cross-compile and package a versioned Core binary archive under `dist/` |
| `uv run inv core.clean` | `go clean`, remove bin/ |

### Interface Tasks (`ops/interface.py`, `ops/interface_runtime.py`)

| Command | Description |
|---------|-------------|
| `uv run inv interface.dev` | Start the Invoke-managed webpack-backed Next.js dev server (stops existing first) |
| `uv run inv interface.install` | `npm install` |
| `uv run inv interface.build` | `npm run build` (production, with one managed retry after a stale repo-local Next build lock or stale `.next/standalone` cleanup lock) |
| `uv run inv interface.lint` | `npm run lint` (ESLint) |
| `uv run inv interface.test` | `npm run test` (Vitest) |
| `uv run inv interface.typecheck` | `npx tsc --noEmit` |
| `uv run inv interface.test-coverage` | Vitest with V8 coverage |
| `uv run inv interface.e2e` | `npm run e2e` (Invoke manages the Next.js server lifecycle via managed `dev` mode by default for stable mocked browser proof; use `--server-mode=start` for the built production Interface server path or live-backend proof. It still uses repo-managed Playwright browsers, defaults to `--workers=1` for repeatability, clears stale Interface listeners and repo-local worker residue before/after, clears an orphaned `interface/.next/dev/lock` only when no repo-local Next worker remains, retries once after stale Next build locks, stale `.next/standalone` cleanup locks, or incomplete built-server packaging, and fails if it cannot bring up and hold its own managed UI server; optional `--headed`, `--project=...`, `--spec=...`, `--live-backend`) |
| `uv run inv interface.stop` | Kill the repo-local Interface server on port 3000 |
| `uv run inv interface.clean` | rm -rf .next cache |
| `uv run inv interface.restart` | stop → clean → build → dev → check |
| `uv run inv interface.check` | Smoke-test: 9 pages for 200 status, no SSR errors, no hydration issues, no `bg-white`/`bg-zinc`/`bg-slate` leaks |

`ops/interface.py` remains the stable Invoke entrypoint. Runtime/browser orchestration now lives in `ops/interface_runtime.py`, shared environment and command helpers live in `ops/interface_env.py`, and repo-local process matching hints live in `ops/interface_process_support.py`.

### Database Tasks (`ops/db.py`)

| Command | Description |
|---------|-------------|
| `uv run inv db.migrate` | Apply canonical forward SQL migrations (`001_init_memory.sql` + `*.up.sql`) only when the target schema is not yet compatible with the current runtime; otherwise skip replay and point to `db.reset` |
| `uv run inv db.reset` | Drop + recreate + migrate |
| `uv run inv db.status` | List tables + row counts |
| `uv run inv db.create` | Create cortex database (if not exists) |

### Auth Tasks (`ops/auth.py`)

| Command | Description |
|---------|-------------|
| `uv run inv auth.dev-key` | Ensure a local `MYCELIS_API_KEY` exists and keep `.env.example` on a sample value |
| `uv run inv auth.break-glass-key` | Ensure a dedicated `MYCELIS_BREAK_GLASS_API_KEY` exists for explicit self-hosted recovery posture |
| `uv run inv auth.posture` | Print the current local-admin and break-glass auth posture from `.env` or `.env.compose` |

### Cache Tasks (`ops/cache.py`)

| Command | Description |
|---------|-------------|
| `uv run inv cache.status` | Report repo-managed cache sizes plus configured user-cache roots |
| `uv run inv cache.guard` | Fail fast when the repo/cache volume does not have enough free space for heavy build/test churn |
| `uv run inv cache.clean` | Clear repo-managed tool caches and local build artifacts under `workspace/tool-cache` |
| `uv run inv cache.clean --user` | Also clear user-level tool caches configured by `cache.apply-user-policy` |
| `uv run inv cache.apply-user-policy` | Persist Windows user cache env vars so pip/npm/go/uv stop defaulting back to `C:` |

### Suggested Development Build Configuration

- Windows:
  - keep the repo and `workspace/tool-cache` on a drive with real headroom when possible
  - run `uv run inv cache.apply-user-policy` on first setup if the user profile lives on a smaller `C:` volume
  - treat Docker Desktop / WSL storage as a separate disk budget from repo-managed build/test caches
- Linux/macOS:
  - keep `MYCELIS_PROJECT_CACHE_ROOT` on the volume intended for repeated builds if the default workspace disk is constrained
  - set user-level cache roots only when tool defaults would otherwise refill a small home/root volume
- Cross-platform:
  - prefer Invoke-managed build/test/browser tasks over raw tool commands so cache roots, browser binaries, telemetry suppression, and worker cleanup stay consistent
  - use `uv run inv cache.status` before large validation runs when free space is tight, then `uv run inv cache.clean` as the first repo-safe reclaim path
  - use `uv run inv cache.guard` when you want a hard preflight on repo/cache disk headroom; it covers the repo/cache volume and intentionally does not pretend to measure Docker daemon image-layer storage

### Deployment Guidance Across Host Architectures

- Windows x86_64:
  - supported as the primary development/operator workstation
  - best used for local iteration, UI work, local validation, and Docker Desktop / Kind-backed development
  - keep repo caches off constrained system volumes and do not treat the workstation image/layout as the canonical production host shape
- Linux x86_64:
  - preferred for longer-running local services, CI-like hosts, and container/Helm-driven deployment targets
  - build binaries and container images for the target Linux host instead of assuming desktop-generated artifacts are portable
  - use the standard env override contract to stamp provider endpoints, model ids, and profile defaults per environment
- Linux arm64:
  - valid for lighter edge/control-host roles, Raspberry Pi style control nodes, and remote-provider-connected deployments
  - keep local service expectations modest unless the ARM host has been explicitly validated for heavier workloads
  - prefer remote Ollama or hosted providers for smaller boards rather than forcing heavyweight local inference onto them
- Mixed-architecture deployment rule:
  - runtime organization truth must stay bundle-driven on every host
  - deployment-time env overrides may vary by environment, but they must not become a replacement for instantiated-organization routing
  - do not reintroduce per-host team/agent env-map routing to compensate for hardware differences

### Lifecycle Tasks (`ops/lifecycle.py`)

| Command | Description |
|---------|-------------|
| `uv run inv lifecycle.up` | Idempotent bring-up: bridge -> dependencies -> Core (waits for `/healthz`; supports `--build` and `--frontend`) |
| `uv run inv lifecycle.down` | Clean teardown: Core -> Frontend -> repo-local Interface worker sweep -> compiled Go service sweep -> port-forwards with bounded cleanup and port-close waits |
| `uv run inv lifecycle.status` | Dashboard for PostgreSQL, NATS, Core, Frontend, Ollama, compiled Go leftovers, and related PIDs |
| `uv run inv lifecycle.health` | Deep health probe against actual API endpoints with auth |
| `uv run inv lifecycle.restart` | Full local restart: down -> settle -> up (supports `--build` and `--frontend`) |
| `uv run inv lifecycle.memory-restart` | Destructive recovery path: down -> db.reset -> up -> health -> memory probes |
| `uv run inv lifecycle.up` | Also ensures the `cortex` database exists before Core starts so bootstrap listeners do not crash after a fresh bridge or cluster reset |

### Compose Tasks (`ops/compose.py`)

| Command | Description |
|---------|-------------|
| `uv run inv compose.infra-up` | Start only the Compose data plane (`postgres` + `nats`), leave Core/Interface down, wait for readiness, and print owner-facing DB/NATS connection settings; use `--migrate` only when schema bootstrap is intentionally needed |
| `uv run inv compose.infra-health` | Probe only the Compose data plane: PostgreSQL port/query readiness, NATS port, and NATS monitor, without checking Core or Interface |
| `uv run inv compose.storage-health` | Probe post-migration Compose PostgreSQL long-term storage: pgvector plus semantic context vectors, durable memory, conversation continuity, artifacts, temporary continuity, collaboration groups, managed exchange, and conversation templates |
| `uv run inv compose.up` | Managed Docker Compose bring-up: postgres + nats -> compatibility-aware canonical forward migrations -> core + interface, with numbered stage output, host readiness checks, and optional `--wait-timeout=<seconds>` |
| `uv run inv compose.down` | Stop the compose stack (`--volumes` for a truly fresh rebuild) |
| `uv run inv compose.migrate` | Apply canonical forward migrations through the PostgreSQL compose service when the compose schema is not already compatible with the current runtime |
| `uv run inv compose.status` | Show compose service state plus host-port reachability |
| `uv run inv compose.health` | Deep health probe for core, text inference availability, frontend, and NATS monitor |
| `uv run inv compose.logs` | Tail compose logs for the full stack or a single service |

Compose runtime guardrails:
- `.env.compose` is the supported env contract for the home-runtime path; do not reuse Kind/bridge `OLLAMA_HOST` values blindly
- use `MYCELIS_COMPOSE_OLLAMA_HOST` for the home-runtime AI engine path so host-level `OLLAMA_HOST` bind settings do not override the compose runtime
- the Compose stack maps `MYCELIS_COMPOSE_OLLAMA_HOST` into provider-specific runtime endpoint overrides inside Core, so deployed execution follows the same explicit provider contract as Helm instead of relying on a global host rewrite
- loopback compose Ollama values (`localhost`, `127.0.0.1`, `0.0.0.0`) are invalid for the Core container and are rejected by the compose task layer
- on Windows hosts without a native `docker` binary, the compose task layer can execute Docker through WSL instead and translates compose/output-block host paths for that runtime; set `MYCELIS_WSL_DISTRO` when the Docker-owning distro is not the default
- on that Windows + WSL Docker path, `compose.up` and `compose.health` can auto-start a WSL-host relay for `MYCELIS_COMPOSE_OLLAMA_HOST`, including when the tasks run directly inside the Docker-owning WSL distro, so bridge containers keep using `host.docker.internal` even when the Windows LAN IP is not directly reachable from Docker-in-WSL
- `compose.infra-up` is the supported personal-owner data-plane preflight when the operator wants PostgreSQL/NATS up and exposed before deciding how Core/Interface should connect
- `compose.storage-health` is the post-migration gate for the long-term Postgres store; it should pass before a personal-owner workflow claims semantic memory, deployment context, retained outputs, managed exchange, or conversation continuity are available
- when the base compose schema is already compatible, `compose.migrate` skips unsafe full replay and can apply known missing late storage migrations before `compose.storage-health` runs
- `compose.up` and `compose.migrate` now follow the same compatibility-aware posture as `db.migrate`: once the compose `cortex` schema already has the required late-runtime tables and columns, the tasks skip forward replay and leave reset/rebuild work to `compose.down --volumes`
- `compose.up` prints deterministic step numbers, stage expectations, and recovery guidance on timeout so both operators and agent callers can follow the same bring-up contract
- prefer `uv run inv compose.up --build --wait-timeout=240` on a fresh or slower host where image build and first readiness can legitimately take longer than the default wait window
- the slim compose Core image disables default npm-backed MCP auto-bootstrap by default to keep startup logs honest; manual/external MCP connectivity remains supported

### Clean Run Discipline for Runtime and Integration Checks

- Before any runtime or integration-style test, stop prior local services using the repo lifecycle task path. Use `uv run inv lifecycle.down` unless a narrower repo task is the safer equivalent for the slice.
- Verify ports and processes are clear for the services involved in the check. At minimum review the Core API port, NATS, PostgreSQL, and Ollama when the slice depends on them, using repo ops tasks such as `uv run inv lifecycle.status` or OS-level port/process tools.
- Detect compiled Go binaries before starting the check. Inspect for repo-local command lines or binary paths plus listeners on declared dev/test ports, terminate them through lifecycle/task helpers when found, and never assume they belong to the current slice.
- Treat repo-local Interface workers as the same cleanup surface. Build/test/browser runs must not leave `node.exe` helpers from `.next`, Vitest, or Playwright behind after the command exits.
- Merge-readiness browser gates should bias toward repeatability. The managed readiness tasks now run Playwright with reduced worker counts on the local host path: `uv run inv ci.baseline` uses `--workers=1` against the built production Interface server path, while `uv run inv ci.service-check --live-backend` stays at `--workers=1`, restores the local bridge/core stack before the live governed Soma proof when that bridge is absent, and reuses an already-initialized `cortex` schema instead of replaying non-idempotent migrations every run.
- Live browser specs that assert backend-written files should bind to the backend's real workspace root. When the browser spec runs from a different checkout or worktree than the live Core backend, set `MYCELIS_BACKEND_WORKSPACE_ROOT` (or `PLAYWRIGHT_BACKEND_WORKSPACE_ROOT`) to the backend's actual workspace directory before running `soma-governance-live.spec.ts`, such as `core/workspace` for a repo-local Core process or `workspace/docker-compose/data/workspace` for the supported compose stack.
- Start only the minimal services required for the specific check. Prefer the narrowest path that matches the validation target, such as Helm render only, bootstrap/unit coverage only, Core-only, or a bounded local stack bring-up.
- Run the test or validation command once the required services are confirmed ready.
- Shut services down immediately after the check unless the slice explicitly requires them left running for a follow-on validation step.
- Agents must never stack runs on top of unknown existing processes.

### Compiled Go Service Cleanup Before Tests

Local Go binaries can survive outside normal lifecycle orchestration after `go build`, `go run`, or manual binary execution. That means the clean-run contract must cover more than containers and bridges.

Typical binaries/process shapes to clear:
- core server
- relays / bridges when a local Go service provides them
- bootstrap helpers
- any repo-local Go service launched through `go run ./cmd/...` or `core/bin/*`

Pre-test cleanup must distinguish:
- local compiled Go services
- containerized dependencies
- test-managed ephemeral services

Stopping containers is necessary but not sufficient. The operator or agent must also inspect local processes for stray compiled Go services and terminate them before the next runtime or integration check begins. If process inspection fails, treat the local environment as unverified and stop before running the check.

### Logging & Quality Gates (`ops/logging.py`, `ops/quality.py`)

| Command | Description |
|---------|-------------|
| `uv run inv logging.check-schema` | Verify runtime event literals map to declared event constants and docs coverage |
| `uv run inv logging.check-topics` | Fail when production code hardcodes `swarm.*` subjects outside allowed constants files |
| `uv run inv quality.max-lines --limit 350` | Enforce the hot-path `<=350 LOC` policy with temporary no-regression legacy caps |

### Kubernetes Tasks (`ops/k8s.py`)

| Command | Description |
|---------|-------------|
| `uv run inv k8s.init` | Create the preferred local Kubernetes cluster (`k3d` when available, Kind fallback) |
| `uv run inv k8s.up` | Canonical local-cluster bring-up: init -> deploy -> wait (PostgreSQL -> NATS -> Core API) |
| `uv run inv k8s.deploy` | Build Core, load or import the Docker image into the active local backend, helm upgrade (injects secrets from .env) |
| `uv run inv k8s.wait` | Wait for rollout readiness gates (PostgreSQL -> NATS -> Core API) |
| `uv run inv k8s.bridge` | Port-forward NATS:4222, HTTP:8080, PG:5432 and verify the local forwards actually bind before reporting success |
| `uv run inv k8s.status` | Check Docker, the preferred local Kubernetes backend, pod status, PVC status |
| `uv run inv k8s.recover` | Restart core + infra resources (core, NATS, PostgreSQL), then wait for readiness and fail closed if the cluster is unreachable |
| `uv run inv k8s.reset` | Full reset: delete cluster -> canonical bring-up (includes readiness wait) |

Kubernetes operator contract:
- local Kubernetes now prefers `k3d` when it is installed; set `MYCELIS_K8S_BACKEND=kind` when you intentionally need the older Kind workflow
- use explicit reachable AI endpoints for deployed text or media engines instead of localhost assumptions
- `uv run inv k8s.deploy` accepts `MYCELIS_K8S_TEXT_ENDPOINT` and `MYCELIS_K8S_MEDIA_ENDPOINT` from the shell or `.env` and forwards them into the Helm chart as operator-owned runtime config
- `uv run inv k8s.deploy` also accepts `MYCELIS_K8S_VALUES_FILE`, resolves repo-relative preset paths, and fails fast if the requested values file does not exist
- the `values-enterprise-windows-ai` preset is intentionally fail-closed: `uv run inv k8s.deploy` / `uv run inv k8s.up` require `MYCELIS_K8S_TEXT_ENDPOINT` to point at the real Windows-hosted AI service before rollout begins
- the Helm chart applies `MYCELIS_K8S_TEXT_ENDPOINT` through provider-specific env overrides (`MYCELIS_PROVIDER_<PROVIDER_ID>_ENDPOINT`) so deployed providers can target a Windows-hosted or otherwise external self-hosted AI service without editing chart source
- for the current GPU-attached Windows-host topology, use a reachable Windows IP or hostname such as `http://192.168.x.x:11434/v1`, not `localhost`

### Cognitive Tasks (`ops/cognitive.py`)

| Command | Description |
|---------|-------------|
| `uv run inv cognitive.install` | Install optional local vLLM/Diffusers dependencies |
| `uv run inv cognitive.llm` | Start optional local vLLM text server (GPU, configurable) |
| `uv run inv cognitive.media` | Start optional local Diffusers media server (OpenAI-compatible) |
| `uv run inv cognitive.up` | Start the optional local engine stack (vLLM + Media, handles Ctrl+C) |
| `uv run inv cognitive.stop` | Kill optional local cognitive processes |
| `uv run inv cognitive.status` | HTTP probes to check optional local vLLM + Media health |

These tasks are not part of the supported default Core + Interface runtime path. The default `uv run inv install` path now targets the supported stack only; use `uv run inv install --optional-engines` or `uv run inv cognitive.install` when you explicitly want local engine helpers.
The repo-local `cognitive.*` helper lane is intended for supported Linux GPU hosts; on Windows, keep Ollama local or point the `vllm` provider at a remote OpenAI-compatible server instead.

### Test Tasks (`ops/test.py`)

| Command | Description |
|---------|-------------|
| `uv run inv test.all` | Run ALL unit tests (Core + Interface), normalizing task failures to a single nonzero exit for scripting/CI use |
| `uv run inv test.coverage` | Go coverage + Vitest V8 coverage |
| `uv run inv test.e2e` | Alias for `interface.e2e`, including `--workers` and `--server-mode` controls |

### CI Tasks (`ops/ci.py`)

| Command | Description |
|---------|-------------|
| `uv run inv ci.lint` | Go vet + Next.js lint (both must pass) |
| `uv run inv ci.test` | Go tests + Interface tests |
| `uv run inv ci.build` | Go binary + Next.js production build (no Docker) |
| `uv run inv ci.check` | Full pipeline: lint → test → build (with stage timers) |
| `uv run inv ci.baseline` | Strict delivery baseline covering gates, focused runtime tests, interface build/typecheck, Vitest, and the shared `interface.e2e` path by default (use `--no-e2e` only for intentionally narrower local debugging) |
| `uv run inv ci.service-check` | Verify the currently running local stack with `lifecycle.health`, plus optional live-backend governed Soma browser proof |
| `uv run inv ci.entrypoint-check` | Verify the supported invoke runner matrix and reject unsupported bare aliases |
| `uv run inv ci.toolchain-check` | Report toolchain versions and optionally enforce Go lock policy |
| `uv run inv ci.release-preflight` | Enforce release gate: clean tree + runner/toolchain checks + strict baseline, with optional `--runtime-posture`, `--service-health`, and `--live-backend` proof for deployment/runtime changes; `--runtime-posture` reads process env plus `.env.compose` / `.env` and fails when no explicit supported AI endpoint contract is configured |
| `uv run inv ci.deploy` | Build + Docker + K8s deploy |

### Other Tasks

| Module | Commands | Purpose |
|--------|----------|---------|
| `ops/proto_relay.py` | `proto.generate` | Protobuf codegen (Go + Python) |
| `ops/proto_relay.py` | `relay.test`, `relay.demo` | Python SDK tests + demo |
| `ops/device.py` | `device.boot --id=ghost-01` | Device simulation (NATS announce) |
| `ops/cache.py` | `cache.status`, `cache.clean`, `cache.apply-user-policy` | Managed cache reporting, cleanup, and Windows user-policy stamping |
| `.dockerignore` | root Docker build-context exclusions | Excludes Interface build/test outputs so repo-root Docker builds do not ingest stale `.next`, coverage, or browser artifacts |
| `ops/misc.py` | `clean.legacy` | Remove legacy Makefiles |
| `ops/misc.py` | `team.architecture-sync`, `team.worktree-triage` | central architect sync plus local worktree/task triage helpers |

Key coordination example: `uv run inv team.architecture-sync`

### Config Module (`ops/config.py`)

**Purpose:** Cross-platform configuration utilities

**Constants:**
```python
CLUSTER_NAME = "mycelis-cluster"
NAMESPACE = "mycelis"
ROOT_DIR = project root
CORE_DIR, SDK_DIR = relative paths
WORKSPACE_DIR = project workspace root
PROJECT_CACHE_ROOT = workspace-managed cache root
API_HOST/PORT = localhost:8081 (env-overridable)
INTERFACE_HOST/PORT = 127.0.0.1:3000 local probe/base URL (env-overridable)
INTERFACE_BIND_HOST = :: default dual-stack UI bind host for localhost + LAN reachability (env-overridable)
```

**Helpers:**
```python
is_windows() → bool
powershell(command: str) → str  # PowerShell with -NoProfile flag
```

**Environment Sanitization:** Clears Windows Store Python's global `VIRTUAL_ENV` on import.

---

## III. Development Workflow

### Deployment Method Selection

Choose the runtime by target environment, not by whichever local helper happens to be installed first.

- Docker Compose:
  - default single-host self-hosted runtime for home-lab, demo, and personal-owner deployment
  - recommended easiest full-stack bring-up on WSL2, Linux, and macOS
- local Kubernetes with `k3d`:
  - preferred repo-local validation lane for Helm, readiness, PVC, ingress, and other cluster behavior
  - not the default choice when the real target is only one host
- enterprise self-hosted Kubernetes:
  - same Helm contract on a customer or enterprise cluster with real registry, secret, ingress, and storage values
  - treat `k3d` as the preflight lane for that chart, not as the production target
- edge or small-node deployment:
  - packaged binary or node-attached service on a smaller Linux host
  - keep AI remote unless the smaller host has been explicitly validated for local inference
- developer source mode:
  - implementation lane for active code changes
  - not the deployment story to hand to operators

AI endpoint rule:
- when a Windows GPU host runs Ollama or another self-hosted AI service, point Compose or Kubernetes at the reachable Windows IP or hostname instead of `localhost`
- use `charts/mycelis-core/values-k3d.yaml` for local `k3d` validation, `charts/mycelis-core/values-enterprise.yaml` for the baseline enterprise self-hosted posture, and `charts/mycelis-core/values-enterprise-windows-ai.yaml` when the cluster should target a Windows-hosted AI endpoint
- set `MYCELIS_K8S_VALUES_FILE` before `uv run inv k8s.deploy` or `uv run inv k8s.up` to apply one of those promoted presets without editing the chart

### Docker Compose Home Runtime

```bash
cp .env.compose.example .env.compose
# set MYCELIS_API_KEY and adjust MYCELIS_COMPOSE_OLLAMA_HOST if needed

uv run inv compose.up --build --wait-timeout=240
uv run inv compose.status
uv run inv compose.health
```

This is the supported single-host runtime for home-lab and demo use when local Kubernetes is unnecessary.
It is also the recommended easiest full-stack bring-up path on WSL2, Linux, and macOS.

### k3d / Local Kubernetes Development

```bash
# Prerequisite (run once per reboot/reset):
uv run inv k8s.up       # preferred local backend (`k3d` when available) -> Helm deploy -> PostgreSQL -> NATS -> Core API
# optional legacy fallback:
#   MYCELIS_K8S_BACKEND=kind uv run inv k8s.up

# Managed local bridge + backend + frontend
uv run inv lifecycle.up --frontend

# Database bootstrap when needed
uv run inv db.migrate

# Verification
uv run inv lifecycle.status
uv run inv lifecycle.health
```

Open [http://localhost:3000](http://localhost:3000)

Use `uv run inv k8s.bridge` only when you intentionally need a manual long-running port-forward outside the managed lifecycle path.

### Cross-Host Working Copy Rule

Do not treat one generated repo environment as portable across Windows and WSL/Linux/macOS.

If you switch host environment for the same checkout, recreate:
- `.venv`
- `interface/node_modules`
- `interface/.next`

Safest posture:
- keep a separate clone or worktree per host environment when you regularly use both Windows and WSL

### First-Time Setup

```bash
# 1. Configure secrets
cp .env.example .env
# Edit .env: POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB

# 2. Boot infrastructure in dependency order
uv run inv k8s.up        # preferred local backend (`k3d` when available) + Helm + readiness gates

# 3. Install frontend deps (first time only)
uv run inv interface.install  # First time only

# 4. Bring up managed local services
uv run inv lifecycle.up --frontend

# 5. Initialize database
uv run inv db.migrate    # Apply canonical forward migrations if the schema is not yet runtime-compatible

# 6. Verify the stack
uv run inv lifecycle.status
uv run inv lifecycle.health
```

### Configure Cognitive Engine

- **UI:** `/settings` → AI Engines (Advanced mode) → change provider routing, configure endpoints
- **Tools:** `/settings` → Connected Tools (Advanced mode) → install from curated library or manually
- **YAML:** Edit `core/config/cognitive.yaml` directly

---

## IV. Configuration System

### Configuration Files

| Config | Location | Managed Via |
|--------|----------|-------------|
| Cognitive (Bootstrap) | `core/config/cognitive.yaml` | UI (`/settings` → AI Engines, Advanced mode) or YAML |
| Bootstrap Templates | `core/config/templates/*.yaml` | YAML (startup-selected transitional V8 migration bundles that instantiate the runtime organization; Task 005 bridge layer). In the deployed Core image, the chart mounts this runtime config tree at `/core/config` so the container workdir and bootstrap bundle lookup stay aligned. |
| Standing Teams | `core/config/teams/*.yaml` | YAML (transitional migration inputs mirrored for compatibility packaging; normal startup now requires a bootstrap bundle and does not read them directly) |
| MCP Servers | Database | UI (`/settings` → Connected Tools, Advanced mode) or API |
| Governance Policy | `core/config/policy.yaml` | UI (`/approvals` → Policy tab) or YAML |
| MCP Library | `core/config/mcp-library.yaml` | YAML (curated registry) |
| Environment | `.env` | Manual |

### Team YAML Files

| File | Team ID | Members |
|------|---------|---------|
| `teams/admin.yaml` | `admin-core` | admin (17 tools, 5 iterations) |
| `teams/council.yaml` | `council-core` | architect, coder, creative, sentry |
| `teams/genesis.yaml` | `genesis-core` | architect, commander |
| `teams/telemetry.yaml` | telemetry | monitoring agents |

Adding a new council member: add to YAML → restart → `GET /api/v1/council/members` auto-discovers.

### Cognitive Configuration (`cognitive.yaml`)

```yaml
providers:
  vllm:          # openai_compatible @ http://127.0.0.1:8000/v1
  ollama:        # openai_compatible @ http://192.168.50.156:11434/v1 (LAN)
  lmstudio:      # openai_compatible @ http://127.0.0.1:1234/v1
  production_gpt4:   # openai (env: OPENAI_API_KEY)
  production_claude: # anthropic (env: ANTHROPIC_API_KEY)
  production_gemini: # google (env: GEMINI_API_KEY)

profiles:
  admin → ollama
  architect → ollama
  coder → ollama
  creative → ollama
  sentry → ollama
  chat → ollama

media:
  endpoint: http://127.0.0.1:8001/v1  # Stable Diffusion Diffusers
```

Provider auth quick reference:
- `openai` providers use Bearer auth and should normally use `api_key_env: OPENAI_API_KEY`
- `anthropic` providers use `x-api-key` plus `anthropic-version` and should use `api_key_env: ANTHROPIC_API_KEY`
- `google` providers use `x-goog-api-key` and should use `api_key_env: GEMINI_API_KEY`
- `openai_compatible` providers share the OpenAI-compatible client path; Ollama can ignore the placeholder key, while vLLM can enforce it when started with `--api-key`

Local engine defaults:
- `ollama` -> `http://127.0.0.1:11434/v1`
- `vllm` -> `http://127.0.0.1:8000/v1`
- `lmstudio` -> `http://127.0.0.1:1234/v1`

### Governance Policy (`policy.yaml`)

```yaml
rules:
  - id: system-critical
    intent: "k8s.delete.*|system.shutdown"
    action: REQUIRE_APPROVAL
  - id: finance
    intent: "payment.create"
    conditions: ["amount > 50"]
    action: REQUIRE_APPROVAL
  - id: iot-safety
    intent: "firmware.update"
    action: REQUIRE_APPROVAL
  - id: iot-speed
    intent: "motor.set_speed"
    conditions: ["value > 8000"]
    action: DENY
defaults:
  default_action: ALLOW
  approval_expiry: 1h
```

### Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `MYCELIS_API_HOST` | `localhost` | Core API host |
| `MYCELIS_API_PORT` | `8081` | Core API port (bridge forward) |
| `MYCELIS_INTERFACE_HOST` | `127.0.0.1` | Local UI probe/base host used by lifecycle checks and browser tooling |
| `MYCELIS_INTERFACE_BIND_HOST` | `::` | Next.js bind host; defaults to dual-stack listening so IPv4 localhost, IPv6 localhost, and LAN clients can all reach the UI |
| `MYCELIS_INTERFACE_PORT` | `3000` | Next.js dev server port |
| `DB_HOST` | `mycelis-core-postgresql` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `mycelis` | PostgreSQL user |
| `DB_PASSWORD` | — | PostgreSQL password |
| `DB_NAME` | `cortex` | PostgreSQL database |
| `NATS_URL` | `nats://mycelis-core-nats:4222` | NATS broker |
| `PORT` | `8080` | Core HTTP listen port |
| `MYCELIS_OUTPUT_BLOCK_MODE` / `outputBlock.mode` | `local_hosted` (Compose) / `cluster_generated` (Helm) | Output-block ownership mode for generated artifacts, media, and workspace files |
| `MYCELIS_OUTPUT_HOST_PATH` | `./workspace/docker-compose/data` (Compose) | Host directory mounted into Core as `/data` when `MYCELIS_OUTPUT_BLOCK_MODE=local_hosted` |
| `MYCELIS_WORKSPACE` | `./workspace` (local) / `/data/workspace` (K8s) | Workspace sandbox root for manifested files and filesystem tools |
| `DATA_DIR` | `./workspace/artifacts` (local) / `/data/artifacts` (K8s) | Artifact storage root for file-backed outputs |
| `MYCELIS_PROJECT_CACHE_ROOT` | `./workspace/tool-cache` | Repo-managed cache root used by Invoke tasks for uv/pip/npm/go/python bytecode |
| `MYCELIS_USER_CACHE_ROOT` | platform-specific | Optional user cache root stamped by `uv run inv cache.apply-user-policy` on Windows |
| `UV_CACHE_DIR` | managed by task env or user policy | uv download/build cache location |
| `PIP_CACHE_DIR` | managed by task env or user policy | pip wheel/http cache location |
| `NPM_CONFIG_CACHE` | managed by task env or user policy | npm/npx cache location |
| `GOCACHE` | managed by task env or user policy | Go build cache location |
| `GOMODCACHE` | managed by task env or user policy | Go module cache location |
| `PLAYWRIGHT_BROWSERS_PATH` | managed by task env or user policy | Playwright browser binary cache location |
| `NEXT_TELEMETRY_DISABLED` | `1` in task-managed Interface runs | Prevents Next telemetry writes during local build/test/browser execution |
| `MYCELIS_COMPOSE_OLLAMA_HOST` | — | Compose-side text AI host that is projected into provider-specific endpoint overrides for the Core container |
| `MYCELIS_PROVIDER_<PROVIDER_ID>_MODEL_ID` | — | Override provider model selection at startup/runtime config load |
| `MYCELIS_PROVIDER_<PROVIDER_ID>_ENDPOINT` | — | Override provider endpoint from automation tooling |
| `MYCELIS_PROVIDER_<PROVIDER_ID>_ENABLED` | — | Enable/disable a provider via env (`true` / `false`) |
| `MYCELIS_PROVIDER_<PROVIDER_ID>_TYPE` | — | Define provider type for env-defined providers |
| `MYCELIS_PROVIDER_<PROVIDER_ID>_API_KEY` | — | Direct provider API key override |
| `MYCELIS_PROVIDER_<PROVIDER_ID>_API_KEY_ENV` | — | Provider API key env indirection |
| `MYCELIS_PROVIDER_<PROVIDER_ID>_TOKEN_BUDGET_PROFILE` | — | Override the provider output-budget preset (`conservative`, `standard`, `extended`, `deep`) |
| `MYCELIS_PROVIDER_<PROVIDER_ID>_MAX_OUTPUT_TOKENS` | — | Override the provider max output budget directly |
| `MYCELIS_PROFILE_<PROFILE>_PROVIDER` | — | Route a profile to a provider via env |
| `MYCELIS_MEDIA_ENDPOINT` | — | Override media/image endpoint |
| `MYCELIS_MEDIA_MODEL_ID` | — | Override media/image model id |
| `MYCELIS_MEDIA_API_KEY_ENV` | — | Point a hosted or protected media provider at an env var containing the API key |
| `MYCELIS_LOCAL_ADMIN_USERNAME` | `admin` | Primary self-hosted local-admin principal name behind `MYCELIS_API_KEY` |
| `MYCELIS_LOCAL_ADMIN_USER_ID` | `00000000-0000-0000-0000-000000000000` | Stable internal id for the primary self-hosted local-admin principal |
| `MYCELIS_BREAK_GLASS_API_KEY` | — | Optional dedicated break-glass credential for self-hosted hybrid/federated recovery |
| `MYCELIS_BREAK_GLASS_USERNAME` | `recovery-admin` | Break-glass recovery principal name |
| `MYCELIS_BREAK_GLASS_USER_ID` | `00000000-0000-0000-0000-000000000001` | Stable internal id for the break-glass recovery principal |
| `OPENAI_API_KEY` | — | OpenAI API key |
| `ANTHROPIC_API_KEY` | — | Anthropic API key |
| `GEMINI_API_KEY` | — | Gemini API key |

Deployment automation rule:
- use `MYCELIS_PROVIDER_<PROVIDER_ID>_*`, `MYCELIS_PROFILE_<PROFILE>_PROVIDER`, and `MYCELIS_MEDIA_*` when automation tools need to stamp environment-specific cognitive config
- do not use the retired `MYCELIS_TEAM_PROVIDER_MAP` / `MYCELIS_AGENT_PROVIDER_MAP` env maps; provider routing now comes from provider config plus instantiated organization policy
- treat env overrides as deployment-time infrastructure configuration only: provider definitions, profile defaults, and environment-specific endpoint/model wiring
- token consumption must stay configuration-owned: set `token_budget_profile` and `max_output_tokens` in `core/config/cognitive.yaml`, override them with `MYCELIS_PROVIDER_<PROVIDER_ID>_TOKEN_BUDGET_PROFILE` / `MYCELIS_PROVIDER_<PROVIDER_ID>_MAX_OUTPUT_TOKENS` during deployment automation, or manage them in `/settings` -> `AI Engines` (Advanced mode)
- safe preset ranges are `conservative=512`, `standard=1024`, `extended=2048`, and `deep=4096`; treat `standard` as the normal local default and opt into larger budgets intentionally
- `MYCELIS_API_KEY` remains the normal local-admin credential for current self-hosted operation; use `MYCELIS_BREAK_GLASS_API_KEY` only for an explicit recovery path, not as a second everyday token
- local-hosted output storage must provide a real `MYCELIS_OUTPUT_HOST_PATH`; the Compose task resolves it with Python `pathlib` before startup so platform-specific paths are normalized and missing directories fail early instead of being silently mis-mounted
- do not treat env overrides as runtime organization behavior or team/role routing control
- runtime truth remains `Bundle -> Instantiated Organization -> Inheritance -> Routing`

---

## V. Testing Strategy

### 5-Tier Verification

| Tier | Tool | Count | Speed | Command |
|------|------|-------|-------|---------|
| 1 | Go unit tests | ~112 tests | <5s | `uv run inv core.test` |
| 2 | Vitest component tests | ~114 tests | <10s | `uv run inv interface.test` |
| 3 | Playwright E2E | 17 spec files + axe accessibility baseline | 30s–2min | `uv run inv interface.e2e` |
| 4 | Integration tests | DB + NATS + LLM | varies | manual |
| 5 | Governance smoke | cmd/smoke | varies | `uv run inv core.smoke` |

### Go Test Patterns

- Partial server construction via `newTestServer(opts...)`
- SQL mocking with `go-sqlmock`
- Path params via `r.PathValue()`
- Nil guards for optional infrastructure
- **Key test files:** `server/{governance,telemetry,mission,mcp,memory_search,proposals,identity,catalogue,artifacts,cognitive}_test.go`
- **Swarm tests:** `swarm/{soma,team,workflow}_test.go`
- **Cognitive tests:** `cognitive/{cognitive,middleware,discovery}_test.go`
- **Other:** `memory/archivist_test.go`, `overseer/engine_test.go`, `governance/{guard,governance}_test.go`

### Frontend Test Patterns

- **Setup:** `__tests__/setup.ts` — mockFetch, MockEventSource, next/navigation mock
- **ReactFlow mock:** `__tests__/mocks/reactflow.ts` — ResizeObserver, components, hooks, enums
- **Store testing:** `useCortexStore.setState()` for pre-seeded state
- **Key test files:** `__tests__/{shell,dashboard}/*.test.tsx`
- **E2E specs:** `e2e/specs/*.spec.ts` (17 files, including mobile and accessibility slices)
- **Config:** `playwright.config.ts`

### Smoke Test (Quality Gate)

`uv run inv interface.check` validates 9 pages:
- HTTP 200 status
- No SSR errors
- No hydration mismatches
- No `bg-white` / `bg-zinc-*` / `bg-slate-*` CSS leaks

---

## VI. CI/CD

### GitHub Actions Workflows

| Workflow | File | Trigger Paths | Checks |
|----------|------|--------------|--------|
| **Core CI** | `core-ci.yaml` | `core/**`, `ops/core.py`, `ops/config.py`, `tasks.py`, `pyproject.toml`, `uv.lock` | workflow-native Python/uv + Go bootstrap, `uv run inv core.test`, direct coverage capture, GolangCI-Lint v1.64.5, `uv run inv core.compile` |
| **Interface CI** | `interface-ci.yaml` | `interface/**`, `ops/interface.py`, `ops/interface_runtime.py`, `ops/interface_env.py`, `ops/interface_process_support.py`, `ops/config.py`, `.npmrc`, `tasks.py`, `pyproject.toml`, `uv.lock` | workflow-native Python/uv + Node bootstrap, `npm ci`, then `uv run inv interface.lint`, `uv run inv interface.typecheck`, `uv run inv interface.test`, and `uv run inv interface.build` |
| **E2E CI** | `e2e-ci.yaml` | `interface/**`, `ops/interface.py`, `ops/interface_runtime.py`, `ops/interface_env.py`, `ops/interface_process_support.py`, `ops/config.py`, `.npmrc`, `tasks.py`, `pyproject.toml`, `uv.lock` | workflow-native Python/uv + Node bootstrap, Playwright browser install, `uv run inv interface.build`, then the stable invoke-managed Chromium/Firefox/WebKit + mobile smoke browser matrix via `uv run inv interface.e2e` |
| **Release Binaries** | `release-binaries.yaml` | tag push `v*` or manual dispatch | workflow-native Python/uv + Go bootstrap, then matrix packaging through `uv run inv core.package` and GitHub release asset upload |

**Trigger:** `pull_request` to `main` and `develop`; push-triggered GitHub pipeline runs are intentionally paused until the initial release-ready gate reopens. Container/image workflows are manual-only via `workflow_dispatch`.

### Local CI

```bash
uv run inv ci.check    # Full pipeline: lint → test → build (with stage timers)
uv run inv ci.deploy   # Build + Docker + K8s deploy
```

### Branching Strategy

Trunk-based development with ephemeral feature branches.

| Type | Prefix | Example |
|------|--------|---------|
| Production | `main` | `main` |
| Feature | `feat/` | `feat/neural-router` |
| Fix | `fix/` | `fix/memory-leak` |
| Chore | `chore/` | `chore/infra-reset` |
| Docs | `docs/` | `docs/api-spec` |

Branch rules:
1. Start every non-hotfix change from a fresh branch off `main`.
2. Keep branch scope single-purpose (one feature/fix lane per branch).
3. Merge via squash after tests/docs are complete.

Commit rules (required):
1. Use conventional commit subject lines: `feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `chore:`.
2. Add commit commentary in the body for non-trivial changes:
   - `Why:` reason for change
   - `What:` key files/surfaces changed
   - `Validation:` exact tests/checks run
3. Do not commit without at least one validation command unless the change is docs-only.

Commit template:

```text
feat(scope): short summary

Why:
- ...

What:
- ...

Validation:
- ...
```

---

## VII. Deployment Architecture

### Kubernetes (Self-Hosted / Helm)

**Helm Chart:** `charts/mycelis-core/`

This chart is the shared contract for both local `k3d` validation and enterprise self-hosted Kubernetes deployment.
Local `k3d` is the validation backend; the enterprise target is the real customer or internal cluster.

```yaml
replicaCount: 1
image: mycelis/core:<immutable-tag> (from core.build)
service: ClusterIP:8080
resources:
  requests: { cpu: 50m, memory: 64Mi }
  limits: { cpu: 200m, memory: 256Mi }
securityContext:
  runAsUser: 1000
  runAsNonRoot: true
  seccompProfile: RuntimeDefault
persistence: 2Gi (data/artifacts)
probes:
  startup/readiness/liveness: GET /healthz
```

**Dependencies (Bitnami/NATS Helm charts):**
- PostgreSQL 12.12.10 (image: pgvector/pgvector:16-alpine, 1Gi persistence)
- NATS 1.2.4 (JetStream: 128Mi memory, 2Gi file)

**Helm Templates:** deployment.yaml, service.yaml, secrets.yaml, network-policy.yaml, data-pvc.yaml

**Probe contract:** `mycelis-core` now declares startup/readiness/liveness probes on `/healthz`, so rollout readiness in `k8s.wait`/`k8s.up` reflects actual API health, not just container process start.

### Docker

**Multi-stage build:**
- Stage 1: golang:1.26-alpine → compile
- Stage 2: alpine → copy binary
- Exposes port 8080
- ENTRYPOINT: `/app/server`

### Persistent Storage Contract

- Kubernetes mounts the data PVC at `/data`.
- `MYCELIS_WORKSPACE` must point at `/data/workspace` for filesystem MCP access and manifested file output.
- `DATA_DIR` must point at `/data/artifacts` for artifact blob/file storage.
- Core prepares both paths at startup so mounted storage is immediately writable for manifestation requests.

### Startup Sequence

```
1.  Load environment (ports, DB URL, NATS URL)
2.  Load config files (cognitive.yaml, policy.yaml)
3.  Connect PostgreSQL (retry 10x)
4.  Load Cognitive Router (provider discovery)
5.  Load Governance Guard (policy rules)
6.  Initialize Memory.Service
7.  Connect NATS (retry 10x, continue degraded if fail)
8.  Start Router → Soma → Axon → Overseer → Memory Subscription
9.  Resolve the selected startup bootstrap bundle, instantiate the runtime organization from it, and build startup teams from that instantiated object; fail closed if no valid bundle is available, and require `MYCELIS_BOOTSTRAP_TEMPLATE_ID` when multiple bundles are mounted
10. Run bounded MCP bootstrap/reconnect work so one hung server cannot block boot indefinitely
11. Start HTTP server (port 8080), register routes
12. Block on SIGINT
```

### Graceful Shutdown

```
On SIGINT:
1. Drain NATS (in-flight messages)
2. Cancel all agent contexts
3. Stop HTTP server
4. Close database connection
5. Exit(0)
```

### Readiness Probe

`GET /healthz` — returns 200 when server is ready

---

## VIII. Environment Gotchas

| Issue | Fix |
|-------|-----|
| Windows Store Python sets global `VIRTUAL_ENV` | `ops/config.py` sanitizes on import |
| PowerShell profile has PSReadLine error | All PS calls use `-NoProfile` via `powershell()` helper |
| PTY not available on Windows | Use `pty=not is_windows()` |
| kubectl v1.32 ↔ K8s v1.33 port-forward bug | Upgraded kubectl to v1.35 via scoop |
| Kind cluster cert chain breaks after Docker restart | Fix with `uv run inv k8s.reset` |
| Ollama on LAN (not localhost) | Use `192.168.50.156:11434` |
| Ollama discovery timeout | 10s (Kind → Host LAN is slow) |
| `shutil`/`os.path` for file ops in tasks | Never use shell commands for file operations |

---

## IX. Monitoring & Observability

### Built-in Metrics

```
GET /api/v1/telemetry/compute:
  - goroutines (runtime.NumGoroutine)
  - memory (Alloc, TotalAlloc, Sys, GC pause)
  - token_rate (tokens/second from cognitive Router)
  - uptime (server duration)
```

### NATS Monitoring

- All team internal traffic captured by Axon → SSE stream
- Signal classification: trigger, response, command, status
- Real-time via NatsWaterfall component

### Logging

- Structured logs via stdlib `log` package
- Topics: bootstrap, identity, NATS, DB, cognitive
- All signals include trace_id for correlation

### Cognitive Health

```
GET /api/v1/cognitive/status:
  - Probes all openai_compatible providers dynamically
  - Reports: provider_id, type, endpoint, model_id, status, last_checked
```


Project-owned config backstops:
- root `.npmrc` anchors direct npm/npx cache usage inside `workspace/tool-cache/npm` and trims low-value cache churn
- `pyproject.toml` routes pytest cache metadata into `workspace/tool-cache/pytest`
- `interface/package.json` keeps `npm run test` non-watch by default; use `npm run test:watch` only when an interactive watch session is explicitly intended
