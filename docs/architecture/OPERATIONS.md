# Mycelis Cortex - Operations Manual
> Navigation: [Project README](../../README.md) | [Overview](OVERVIEW.md) | [Backend](BACKEND.md) | [Frontend](FRONTEND.md) | [Testing](../TESTING.md)

This manual owns task and runtime operations. It links to [Local Development Workflow](../LOCAL_DEV_WORKFLOW.md) for setup details and [Testing](../TESTING.md) for evidence gates.

Implementation slices that change runtime, tasking, validation, API meaning, or operator behavior must review and update the owning docs in the same change rather than leaving docs drift for later cleanup.

Current GitHub Actions posture: workflows are intentionally manual-only through `workflow_dispatch`. Push and pull-request hosted runs are paused while local proof, Rancher Desktop K3s, WSL/Compose, and explicit manual workflow runs carry release discipline.

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

## I. Prerequisites

Operational lanes use:
- `uv` and Invoke for repo task execution
- Docker/Compose for the supported home-runtime stack
- Rancher Desktop K3s on Windows, or k3d/Kind plus kubectl/Helm for local Kubernetes
- Go, Node.js, and Python for source-mode development
- PostgreSQL, NATS, and a reachable AI endpoint for live runtime proof

Use `uv run inv ...` for real tasks. Use `uvx --from invoke inv -l` only as a compatibility probe.

## II. Task Automation

Task modules live under `ops/*.py` and are registered through `tasks.py`. App-tied management logic belongs in Python; PowerShell may only wrap host-specific calls.

### Master Registry

List tasks:

```bash
uv run inv -l
```

### Core Tasks (`ops/core.py`)

```bash
uv run inv core.test
uv run inv core.compile
uv run inv core.run
uv run inv core.stop
uv run inv core.restart
uv run inv core.package
uv run inv core.smoke
```

### Interface Tasks (`ops/interface.py`, `ops/interface_runtime.py`)

```bash
uv run inv interface.install
uv run inv interface.test
uv run inv interface.typecheck
uv run inv interface.build
uv run inv interface.e2e
uv run inv interface.check
uv run inv interface.stop
```

### Database Tasks (`ops/db.py`)

```bash
uv run inv db.create
uv run inv db.migrate
uv run inv db.reset
uv run inv db.status
```

`db.migrate` is forward-bootstrap aware: compatible schemas are not replayed as a cleanup mechanism. Use `db.reset` when you intentionally need a fresh rebuild.

### Auth Tasks (`ops/auth.py`)

```bash
uv run inv auth.dev-key
uv run inv auth.break-glass-key
uv run inv auth.posture
```

### Cache Tasks (`ops/cache.py`)

```bash
uv run inv cache.status
uv run inv cache.guard
uv run inv cache.clean
```

### Lifecycle Tasks (`ops/lifecycle.py`)

```bash
uv run inv lifecycle.up --frontend
uv run inv lifecycle.status
uv run inv lifecycle.health
uv run inv lifecycle.restart --frontend
uv run inv lifecycle.down
uv run inv lifecycle.memory-restart --frontend
```

`lifecycle.status` is the fast local snapshot. It reports process/port state and confirms Core through `/healthz` plus Ollama through `/api/tags` over loopback fallbacks. Use `lifecycle.health` for the deeper endpoint proof gate before claiming service readiness.

### Compose Tasks (`ops/compose.py`)

```bash
uv run inv compose.up --build --wait-timeout=240
uv run inv compose.status
uv run inv compose.health
uv run inv compose.warm-cognitive
uv run inv compose.migrate
uv run inv compose.storage-health
uv run inv compose.infra-up --wait-timeout=180
uv run inv compose.infra-health
uv run inv compose.down
```

Compose is the supported single-host runtime lane. `.env.compose` owns container topology; `.env` remains the secret source.
The WSL release proof health-gates each live browser spec with `compose.health` because the runner executes specs through separate WSL shell invocations.

### Kubernetes Tasks (`ops/k8s.py`)

```bash
uv run inv k8s.up
uv run inv k8s.status
uv run inv k8s.wait
uv run inv k8s.deploy
uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml
uv run inv k8s.reset
```

Prefer Rancher Desktop K3s on Windows and `k3d` on WSL/Linux; use `MYCELIS_K8S_BACKEND=kind` only for the explicit legacy path. `MYCELIS_K8S_BACKEND=rancher` targets the existing Rancher Desktop cluster and does not create or reset it.

### Cognitive Tasks (`ops/cognitive.py`)

```bash
uv run inv cognitive.install
uv run inv cognitive.up
uv run inv cognitive.status
uv run inv cognitive.stop
```

These are optional local GPU/helper tasks, not the default path for every host.

### Test Tasks (`ops/test.py`) And CI Tasks (`ops/ci.py`)

```bash
uv run inv test.all
uv run inv test.e2e
uv run inv ci.test
uv run inv ci.baseline
uv run inv ci.service-check
uv run inv ci.release-preflight --lane=release
uv run inv ci.entrypoint-check
uv run inv team.architecture-sync
```

Use `--live-backend` for browser proof that must hit a real Core backend.

### Logging & Quality Gates (`ops/logging.py`, `ops/quality.py`)

```bash
uv run inv logging.check-schema
uv run inv logging.check-topics
uv run inv quality.max-lines --limit 300
```

## III. Development Workflow

Choose the runtime lane first:
- Compose for supported home-runtime proof
- Rancher Desktop K3s or k3d/Helm for local cluster proof
- source lifecycle for implementation
- WSL proof checkout for guarded Compose release-style validation from Windows

Windows edit -> git push -> WSL refresh -> WSL validate is the guarded WSL Compose proof path:

```bash
uv run inv wsl.status
uv run inv wsl.refresh
uv run inv wsl.validate --lane=release
uv run inv wsl.validate --lane=release --headed-browser
```

The root and Interface install tasks use `npm ci` for Interface dependencies; release proof expects
dependency bootstrap to preserve a clean lockfile and fail before validation if it cannot.

### Deployment Guidance Across Host Architectures

- Windows x86_64: edit on Windows; prove through Compose, WSL Compose, or Kubernetes as the slice requires.
- Linux x86_64: use Compose for home-runtime and k3d/Helm for cluster proof.
- Linux arm64: prefer Compose or binary/control-node lanes with explicit remote AI.
- Mixed-architecture deployment rule: keep AI, database, NATS, Core, and Interface addresses explicit.

## IV. Configuration System

Configuration sources:
- `.env`: secrets and host-local runtime values
- `.env.compose`: Compose topology
- `core/config/cognitive.yaml`: provider profiles/routing
- `core/config/homepage.yaml`: portal copy/links
- `core/config/policy.yaml`: governance
- `core/config/templates/*.yaml`: bootstrap bundles/templates
- `core/config/teams/*.yaml`: standing team and legacy migration inputs
- Helm values files: cluster deployment shape

Provider/media env overrides are deployment-time infrastructure configuration only. Use `MYCELIS_PROVIDER_<PROVIDER_ID>_MODEL_ID`, `MYCELIS_PROVIDER_<PROVIDER_ID>_ENDPOINT`, `MYCELIS_PROFILE_<PROFILE>_PROVIDER`, and `MYCELIS_MEDIA_MODEL_ID`; the retired `MYCELIS_TEAM_PROVIDER_MAP` / `MYCELIS_AGENT_PROVIDER_MAP` must not return. `Bundle -> Instantiated Organization -> Inheritance -> Routing` remains authoritative, so do not treat env overrides as runtime organization behavior.

The Helm chart mounts this runtime config tree at `/core/config`.

Runtime bootstrap follows [V8 Config and Bootstrap Model](../architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md): template inputs become instantiated organizations through inheritance, precedence, and policy checks.

Startup should select a valid bundle, instantiate the runtime organization from it, fail closed if no valid bundle is available, and require `MYCELIS_BOOTSTRAP_TEMPLATE_ID` when multiple bundles are mounted. `lifecycle.up` ensures the `cortex` database exists before Core starts.

## V. Testing Strategy

Use [Testing](../TESTING.md) for gate details. Operational summary:
- backend: `uv run inv core.test`
- frontend: `uv run inv interface.test` and `interface.typecheck`
- browser: `uv run inv interface.e2e`
- live stack: `lifecycle.health` or `compose.health`
- release: `ci.baseline` or `wsl.validate --lane=release`
- visible live-window proof: `wsl.validate --lane=release --headed-browser` or focused `interface.e2e --headed --live-backend`

Runtime checks must start clean, verify readiness, run proof once services are healthy, and shut down unless a follow-on check needs them.

## VI. CI/CD

Local CI tasks:
- `ci.test`: Go + Interface tests
- `ci.baseline`: branch-readiness gate
- `ci.service-check`: currently running local stack
- `ci.release-preflight`: lane-aware release gate

GitHub CI proves repo health without hosted agentry. Live service/browser proof is local, WSL, Compose, or target-cluster evidence.

Invoke manages the Next.js server lifecycle for browser proof. Merge gates use the built production Interface server path, while CI also keeps a mocked Chromium homepage smoke. Use `uv run inv ci.service-check` for the currently running stack. `ci.release-preflight` supports `--lane=baseline|runtime|service|release`; `--lane=release` is the recommended full runtime/operator gate. Guarded WSL tasks are `wsl.status`, `wsl.refresh`, `wsl.validate`, `wsl.cycle`; add `--headed-browser` to `wsl.validate` or `wsl.cycle` when focused live-backend Playwright proof must open visible browser windows.
`uv run inv interface.check` includes a small retry loop for transient Windows socket-reuse errors after heavy browser proof, but persistent route failures still fail the task.

## VII. Deployment Architecture

### Kubernetes (Self-Hosted / Helm)

This chart is the target clustered deployment contract for self-hosted and enterprise Kubernetes. Use Helm with explicit values, secrets, persistent storage, ingress, and reachable AI endpoints. Use `uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml` for chart posture. local Kubernetes now prefers `k3d` when it is installed on WSL/Linux, prefers Rancher Desktop K3s on Windows, and falls back with `MYCELIS_K8S_BACKEND=kind`.

Open-standard resources include Deployment, Service, ServiceAccount, Secret, ConfigMap, PVC, Ingress, NetworkPolicy. `MYCELIS_K8S_VALUES_FILE` may select `values-k3d.yaml`, `values-enterprise.yaml`, or `values-enterprise-windows-ai.yaml`. `MYCELIS_K8S_TEXT_ENDPOINT` and `MYCELIS_K8S_TEXT_MODEL_ID` project provider endpoint/model overrides into Core, and configured text endpoints enable explicit AI egress in the chart NetworkPolicy.

### Docker

Docker Compose is rapid local development/proof only and must not become the production deployment standard. Use Compose for single-host self-hosted runtime. Keep `MYCELIS_COMPOSE_OLLAMA_HOST` container-reachable and avoid `localhost` unless it is meaningful inside the container path. use explicit reachable AI endpoints for deployed text or media engines instead of localhost assumptions; the Helm chart applies `MYCELIS_K8S_TEXT_ENDPOINT` through provider-specific env overrides.

use a reachable Windows IP or hostname such as `http://192.168.x.x:11434/v1`, not `localhost`; Compose can auto-start a WSL-host relay for `MYCELIS_COMPOSE_OLLAMA_HOST`, that relay is restartable, and `compose.warm-cognitive` proves the configured text model can complete a tiny chat before live browser proof starts.

the compose Core image includes Node/npm/npx so manual curated stdio MCP installs can launch from the shipped container; manual `filesystem` installs from the curated library are runtime-normalized to the configured `MYCELIS_WORKSPACE` root.

### Persistent Storage Contract

PostgreSQL plus pgvector owns durable memory/context. Output block storage is configured by `MYCELIS_OUTPUT_BLOCK_MODE` and `MYCELIS_OUTPUT_HOST_PATH` for local-hosted Compose, or PVC-backed cluster storage for Kubernetes. Native Core live-browser proof infers the host-visible workspace as `core/workspace` from `MYCELIS_WORKSPACE=./workspace`; K8s live-browser proof that checks backend-written files should use `PLAYWRIGHT_BACKEND_WORKSPACE_PROBE=k8s` so the assertion targets the Core pod workspace/PVC rather than host-only paths.

### Startup Sequence

Core startup depends on resolved config, database, NATS, provider posture, policy, and bootstrap bundle availability. Normal startup should fail closed when required bootstrap inputs are missing.

### Graceful Shutdown

Use task-owned teardown:

```bash
uv run inv lifecycle.down
uv run inv compose.down
```

### Readiness Probe

Use task health checks rather than raw port checks when claiming product readiness.

## VIII. Environment Gotchas

- Windows `OLLAMA_HOST=0.0.0.0` is a listen address, not a container-reachable endpoint.
- Docker-in-WSL may require `MYCELIS_WSL_DISTRO`.
- Compose may relay a Windows-hosted AI endpoint through WSL for bridge containers.
- Do not share generated environments across Windows and WSL.
- Browser proof that checks backend-written files infers `core/workspace` for repo-local native Core; Compose or split-checkout proof may still set `MYCELIS_BACKEND_WORKSPACE_ROOT`.
- K8s/PVC browser proof should use `PLAYWRIGHT_BACKEND_WORKSPACE_PROBE=k8s`, `PLAYWRIGHT_K8S_NAMESPACE`, `PLAYWRIGHT_K8S_CORE_SELECTOR`, and `PLAYWRIGHT_K8S_BACKEND_WORKSPACE_ROOT`.

## IX. Monitoring & Observability

Use:
- `lifecycle.status` / `lifecycle.health`
- `compose.status` / `compose.health`
- `k8s.status`
- structured Core logs
- NATS and PostgreSQL health probes
- activity/run timeline in the UI
Operator-facing errors should be normalized and human-readable; raw backend noise belongs in logs, not the default UI.
