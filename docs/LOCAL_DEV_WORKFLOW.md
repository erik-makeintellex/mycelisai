# Local Development Workflow
> Navigation: [Project README](../README.md) | [Docs Home](README.md)

Use this file for local setup and day-to-day runtime choices. Use [Operations](architecture/OPERATIONS.md) for task ownership and [Testing](TESTING.md) for evidence gates.

## TOC

- [Prerequisites](#prerequisites)
- [Deployment Method Selection](#deployment-method-selection)
- [Configuration Reference](#configuration-reference)
- [Recommended Host Paths](#recommended-host-paths)
- [WSL/Linux Proof Checkout Handoff](#wsllinux-proof-checkout-handoff)
- [Quick Start Paths](#quick-start-paths)
- [First-Time Setup](#first-time-setup)
- [Daily Startup Sequence](#daily-startup-sequence)
- [Port Map](#port-map)
- [Health Checks](#health-checks)
- [Troubleshooting](#troubleshooting)
- [Cognitive Engine Setup](#cognitive-engine-setup)
- [Binary Release Process](#binary-release-process)
- [Full Command Reference](#full-command-reference)

## Prerequisites

Install the toolchain needed for the runtime lane you will use:

| Tool | Purpose |
| --- | --- |
| Docker Engine / Docker CLI | Compose runtime and local Kubernetes container backend |
| Rancher Desktop | preferred Windows local Docker + K3s control plane |
| k3d | preferred WSL/Linux local Kubernetes backend |
| Kind | explicit legacy fallback via `MYCELIS_K8S_BACKEND=kind` |
| kubectl and Helm | cluster interaction and chart deploy/render |
| Go 1.26 | Core build/test |
| Node.js 20+ | Interface build/test |
| uv | Python environment and Invoke task runner |
| psql 16+ | database setup and migration checks |
| Ollama or compatible endpoint | local/self-hosted text inference |

Task runner contract: use `uv run inv ...` for real execution; use `uvx --from invoke inv -l` only as a compatibility probe; do not use bare `uvx inv ...`. Tasks are scoped to Mycelis tools, app services, data-plane dependencies, and proof lanes; host runtime lifecycle for WSL, Rancher Desktop, Docker Desktop, and OS VM repair stays outside the repo task runner.

## Deployment Method Selection

| Target | Use | Notes |
| --- | --- | --- |
| Single-host self-hosted runtime | Docker Compose | default user/demo/home-lab lane |
| Local Kubernetes proof on Windows | Rancher Desktop K3s | preferred commercial-release parity lane |
| Local Kubernetes proof on WSL/Linux | `k3d` | preferred chart validation lane |
| Enterprise/self-hosted cluster | Helm | use real ingress, secrets, storage, and policy controls |
| Small control node | packaged binary | keep AI on a reachable remote service |
| Source development | lifecycle tasks | best for implementation, not the operator deployment story |

Promoted chart presets:
- `charts/mycelis-core/values-k3d.yaml`
- `charts/mycelis-core/values-enterprise.yaml`
- `charts/mycelis-core/values-enterprise-windows-ai.yaml`

Set `MYCELIS_K8S_VALUES_FILE` before `uv run inv k8s.deploy` or `uv run inv k8s.up` to apply a preset.

## Deployment Guidance By Host Architecture

- Windows x86_64: use Windows editing plus Rancher Desktop Docker-compatible Compose, Rancher Desktop K3s proof, WSL Compose proof, or self-hosted Kubernetes.
- Linux x86_64: use Compose for home-runtime proof and k3d/Helm for cluster proof.
- Linux arm64: prefer Compose or binary/control-node lanes with remote AI where local model capacity is limited.
- Mixed-architecture deployments: keep the control plane, runtime, and AI endpoint addresses explicit instead of relying on localhost assumptions.

## Configuration Reference

Keep secrets in `.env`. Keep Compose topology and non-secret container shape in `.env.compose`.

Important files:
- `.env.example`: starting point for local secret/runtime values
- `.env.compose.example`: starting point for Compose topology
- `core/config/cognitive.yaml`: provider profiles and default routing
- `core/config/homepage.yaml`: deployer-editable portal copy/links
- `core/config/policy.yaml`: governance rules
- `core/config/templates/*.yaml`: bootstrap/template bundles
- `core/config/teams/*.yaml`: standing teams and legacy migration inputs
- `charts/mycelis-core/values*.yaml`: Helm deployment shape

Common runtime variables:
- `MYCELIS_API_KEY`: primary local-admin credential
- `MYCELIS_BREAK_GLASS_API_KEY`: optional recovery credential
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`: local Core database connection
- `NATS_URL`: Core NATS connection
- `MYCELIS_WORKSPACE`: workspace sandbox root
- `MYCELIS_COMPOSE_OLLAMA_HOST`: Compose-reachable text model endpoint
- `MYCELIS_K8S_TEXT_ENDPOINT`: Kubernetes/Helm text model endpoint
- `MYCELIS_K8S_TEXT_MODEL_ID`: Kubernetes/Helm text model override
- `MYCELIS_MEDIA_ENDPOINT`, `MYCELIS_MEDIA_MODEL_ID`: media provider overrides
- `MYCELIS_SEARCH_PROVIDER`, `MYCELIS_SEARXNG_ENDPOINT`, `MYCELIS_SEARCH_LOCAL_API_ENDPOINT`, `MYCELIS_SEARCH_MAX_RESULTS`: governed search posture; default native Core search is `local_sources`, while Compose defaults to self-hosted `searxng`

Provider override pattern:
- `MYCELIS_PROVIDER_<PROVIDER_ID>_MODEL_ID`
- `MYCELIS_PROVIDER_<PROVIDER_ID>_ENDPOINT`
- `MYCELIS_PROVIDER_<PROVIDER_ID>_ENABLED`
- `MYCELIS_PROVIDER_<PROVIDER_ID>_API_KEY_ENV`
- `MYCELIS_PROFILE_<PROFILE>_PROVIDER`

These env overrides are for deployment-time provider definition. They are not runtime organization behavior and do not replace bundle-defined truth. The retired `MYCELIS_TEAM_PROVIDER_MAP` / `MYCELIS_AGENT_PROVIDER_MAP` must not return.

Bootstrap truth stays:

```text
Bundle -> Instantiated Organization -> Inheritance -> Routing
```

The deployed Core image resolves those files from `/core/config`; the Helm chart mounts the runtime config volume there.

## Recommended Host Paths

Use separate generated dependency/runtime roots per host:
- Windows editing checkout: source, review, commits, pushes
- WSL proof checkout: release-style install/build/test/runtime proof
- Compose data/output roots: `workspace/docker-compose/...`
- repo-local tool/cache roots: inspect with `uv run inv cache.status`

Do not share `.venv`, `interface/node_modules`, `.next`, or generated runtime state across Windows and WSL.

## WSL/Linux Proof Checkout Handoff

Windows is the edit/review/git surface. WSL is the guarded Compose proof surface; Rancher Desktop K3s is the Windows local Kubernetes proof surface.
Guarded sequence:

```bash
uv run inv wsl.status
uv run inv wsl.refresh
uv run inv wsl.validate --lane=release
```

Rules:
- refresh the proof checkout from git
- do not copy source trees across the host boundary
- keep destructive source cleanup scoped to the dedicated WSL proof checkout
- use the Windows browser at `http://localhost:3000` when proving a WSL-hosted stack on the same machine
- use `uv run inv wsl.validate --lane=release --headed-browser` when release evidence needs visible live-window Playwright proof against the Compose-delivered UI
- use platform tooling, not Invoke, when the WSL distro or desktop container runtime itself needs shutdown, reset, or repair

## Quick Start Paths

Source-mode development:

```bash
uv run inv install
uv run inv auth.dev-key
uv run inv lifecycle.up --frontend
uv run inv lifecycle.health
```

Compose home runtime:

```bash
uv run inv compose.up --build --wait-timeout=240
uv run inv compose.status
uv run inv compose.health
```

Windows local Kubernetes with Rancher Desktop K3s:
```powershell
$env:MYCELIS_K8S_BACKEND="rancher"
$env:MYCELIS_K8S_VALUES_FILE="charts/mycelis-core/values-k3d.yaml"
$env:MYCELIS_K8S_TEXT_ENDPOINT="http://<windows-ai-host>:11434/v1"
$env:MYCELIS_K8S_TEXT_MODEL_ID="qwen3:8b"
uv run inv k8s.up
uv run inv k8s.bridge
uv run inv lifecycle.up --frontend
uv run inv lifecycle.health
```
WSL/Linux local Kubernetes with k3d uses the same `values-k3d.yaml` preset with `MYCELIS_K8S_BACKEND=k3d` or auto-detection when k3d is installed.

## First-Time Setup

1. Copy environment examples:

```bash
cp .env.example .env
cp .env.compose.example .env.compose
```

2. Stamp local credentials:

```bash
uv run inv auth.dev-key
```

3. Install dependencies:

```bash
uv run inv install
```

4. Choose one runtime lane from [Deployment Method Selection](#deployment-method-selection), then run the matching quick start.

Startup now instantiates the runtime organization only through a selected bundle. Core fails closed when no valid bundle exists in `core/config/templates/`. Use `MYCELIS_BOOTSTRAP_TEMPLATE_ID` when multiple bundles are mounted; no-bundle operation is not a normal startup path.

## Daily Startup Sequence

Source-mode:

```bash
uv run inv lifecycle.up --frontend
uv run inv lifecycle.status
uv run inv lifecycle.health
```

Compose:

```bash
uv run inv compose.up --wait-timeout=180
uv run inv compose.status
uv run inv compose.health
```

Shutdown:

```bash
uv run inv lifecycle.down
uv run inv compose.down
```

## Port Map

| Service | Default |
| --- | --- |
| Interface | `http://localhost:3000` |
| Core API | `http://localhost:8081` source, `8080` in cluster |
| PostgreSQL | `localhost:5432` |
| NATS | `localhost:4222` |
| Ollama | `http://127.0.0.1:11434` host-local |

## Health Checks

```bash
uv run inv lifecycle.status
uv run inv lifecycle.health
uv run inv compose.status
uv run inv compose.health
uv run inv compose.storage-health
uv run inv k8s.status
uv run inv interface.check
uv run inv cognitive.status
```

Treat `compose.health` as product availability proof: the text cognitive engine must be reachable before browser proof is valid.

For day-to-day Windows development, keep the Windows repo as the edit/review/push surface.

Use this as the guarded WSL Compose deployment-mimic proof path when Windows editing is ready for build, API, UI, runtime, or release-style validation:

```bash
uv run inv wsl.status
uv run inv wsl.refresh --branch <name>
uv run inv wsl.validate
uv run inv wsl.validate --lane=release --headed-browser
uv run inv wsl.cycle --branch <name>
uv run inv ci.release-preflight --lane=release
```

These tasks keep the WSL proof checkout git-backed and disposable instead of turning it into a second editing worktree. Always keep the WSL `mycelis-root` deployment checkout git-backed and disposable for Compose deployment-mimic proof; use Rancher Desktop K3s when the proof target is local Kubernetes/Helm parity.

Guarded commands: `uv run inv wsl.status`, `uv run inv wsl.refresh --branch <name>`, `uv run inv wsl.validate`, `uv run inv wsl.validate --headed-browser`, `uv run inv wsl.cycle --branch <name>`.

## Troubleshooting

- NATS or council responses disappear: run `uv run inv lifecycle.status`, then restart with `uv run inv lifecycle.restart --frontend`.
- Compose readiness times out on a first build: rerun with `uv run inv compose.up --build --wait-timeout=240`.
- Windows host has no native `docker`: the Compose task may use Docker inside WSL; set `MYCELIS_WSL_DISTRO` if the default distro is not the Docker host.
- Windows-hosted Ollama from WSL/Docker: keep `MYCELIS_COMPOSE_OLLAMA_HOST` pointed at the intended Windows service address; the task layer may relay it through WSL.
- Disk pressure: run `uv run inv cache.guard`, `uv run inv cache.status`, and `uv run inv clean.disk-status`.
- Stale local services: stop with `uv run inv lifecycle.down` before runtime or integration proof.

## Cognitive Engine Setup

Default local text inference is Ollama. Hosted or OpenAI-compatible providers are configured through `core/config/cognitive.yaml`, env overrides, or the Advanced AI Engine Settings UI.

Use explicit non-loopback endpoints for Compose/Kubernetes when the model service is outside the container/cluster:
- Compose: `MYCELIS_COMPOSE_OLLAMA_HOST=http://host.docker.internal:11434` or a reachable LAN host
- Kubernetes: `MYCELIS_K8S_TEXT_ENDPOINT=http://<host>:11434/v1` plus `MYCELIS_K8S_TEXT_MODEL_ID=<installed-model>` when the cluster should use a specific self-hosted model.

Optional local GPU helpers:

```bash
uv run inv cognitive.install
uv run inv cognitive.up
uv run inv cognitive.status
```

## Binary Release Process

Package Core with:

```bash
uv run inv core.package
```

Run release gates from [Testing](TESTING.md) before handoff.

## Full Command Reference

List current tasks with:

```bash
uv run inv -l
```

The task registry lives in Python under `tasks.py` and `ops/*.py`. App-tied management logic belongs there, not in PowerShell scripts.
