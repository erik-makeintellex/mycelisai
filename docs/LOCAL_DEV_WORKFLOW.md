# Local Development Workflow
> Navigation: [Project README](../README.md) | [Docs Home](README.md)

> **Quick Start:** Already set up? Jump to [Daily Startup](#daily-startup-sequence).
> **Invoke Contract:** Run project tasks with `uv run inv ...`.
> Use `uvx --from invoke inv -l` only as a lightweight compatibility probe.
> Do not use bare `uvx inv ...`.
> **Management Scripting:** App-tied management logic belongs in Python task modules. PowerShell is a host wrapper only when the local platform requires it.
> **Lifecycle Readiness:** `uv run inv lifecycle.up ...` now waits for Core `/healthz` readiness and fails fast if the API never becomes healthy.
> **Lifecycle Teardown:** `uv run inv lifecycle.down` now uses bounded cleanup timeouts and waits for Core/Frontend ports to close before reporting success.
> **Startup Diagnostics:** Background Core startup output is captured in `workspace/logs/core-startup.log`.

## TOC

- [Prerequisites](#prerequisites)
- [Deployment Method Selection](#deployment-method-selection)
- [Configuration Reference](#configuration-reference)
- [Recommended Host Paths](#recommended-host-paths)
- [WSL/Linux Codex Handoff](#wsllinux-codex-handoff)
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

| Tool | Version | Install | Purpose |
|:--|:--|:--|:--|
| Docker Engine / Docker CLI | Latest | Docker Desktop, native Linux Docker, or Docker inside WSL | Container runtime for Compose or local Kubernetes bring-up |
| k3d | Latest | `brew install k3d` or package manager of choice | Preferred local Kubernetes cluster in Docker |
| Kind | Latest | `scoop install kind` / `brew install kind` | Legacy local Kubernetes fallback when `MYCELIS_K8S_BACKEND=kind` is required |
| kubectl | **v1.35+** | `scoop install kubectl` / `brew install kubectl` | K8s CLI (v1.32 has port-forward bugs) |
| Helm | v3+ | `scoop install helm` / `brew install helm` | Chart deployment |
| Go | 1.26 | [go.dev](https://go.dev/) | Backend compiler |
| Node.js | 20+ | [nodejs.org](https://nodejs.org/) | Frontend runtime |
| uv | Latest | `pip install uv` / `pipx install uv` | Python environment + task runner |
| psql | 16+ | Comes with PostgreSQL or standalone | Database migrations |
| Ollama | Latest | [ollama.com](https://ollama.com/) | Local LLM inference |

Platform-first guidance:
- WSL2/Linux/macOS should usually start with the Docker Compose path because it is the easiest full-stack bring-up and avoids the extra Kind/bridge layer.
- when you do need local Kubernetes, prefer `k3d`; use `MYCELIS_K8S_BACKEND=kind` only when you intentionally need the older Kind flow
- Windows native is still a supported operator/development path, but the durable runtime stories are Compose and self-hosted Kubernetes rather than a Windows-only Docker Desktop assumption.
- Optional repo-local `cognitive.*` helpers are for supported Linux GPU hosts; they are not the default setup path on Windows or macOS.

## Deployment Method Selection

Choose the runtime by the environment you are targeting:

| Target environment | Choose this path | Why |
| :-- | :-- | :-- |
| Single-host self-hosted runtime | Docker Compose | Easiest supported full-stack bring-up for one machine, demos, and home-lab use |
| Local Kubernetes or Helm validation | `k3d` | Best repo-local proof for chart behavior before a real cluster |
| Enterprise or customer-managed cluster | Helm on self-hosted Kubernetes | Best fit for real ingress, secrets, registry, storage, and policy controls |
| Small node or Raspberry Pi style control host | Packaged binary or node-attached service | Keeps the runtime lighter and lets AI stay remote |
| Active code changes | Source-mode lifecycle tasks | Good for implementation work, but not the deployment target to recommend to operators |

Deployment selection rules:
- use Docker Compose for the normal single-host self-hosted story
- use `k3d` when the goal is local Kubernetes behavior, not just "run the app somehow"
- use the Helm chart for enterprise self-hosted Kubernetes and treat `k3d` as the preflight lane for that chart
- use the binary path for edge/control-node deployments and point it at a reachable remote AI service
- when the AI service lives on a Windows GPU host, keep Compose or Kubernetes pointed at the reachable Windows IP or hostname rather than `localhost`

Promoted Kubernetes preset files:
- `charts/mycelis-core/values-k3d.yaml` for local `k3d` validation
- `charts/mycelis-core/values-enterprise.yaml` for an enterprise-shaped self-hosted cluster
- `charts/mycelis-core/values-enterprise-windows-ai.yaml` for an enterprise-shaped self-hosted cluster that points at a Windows-hosted AI service
- set `MYCELIS_K8S_VALUES_FILE` before `uv run inv k8s.deploy` or `uv run inv k8s.up` to apply one of those presets without editing the chart

## Configuration Reference

All configuration lives in **three places** — the `.env` file, `cognitive.yaml`, and team YAML files.

### `.env` (Project Root)

The `.env` file is the **single source** for all runtime configuration. Copy from `.env.example` on first setup.

```bash
cp .env.example .env
```

| Variable | Default | Description |
|:--|:--|:--|
| **Cognitive Engine** | | |
| `OLLAMA_HOST` | `http://127.0.0.1:11434` | Ollama API endpoint. Use `127.0.0.1` for local, or LAN IP for remote. |
| `MYCELIS_PROVIDER_<PROVIDER_ID>_MODEL_ID` | — | Override a provider model at startup/runtime config load time. Example: `MYCELIS_PROVIDER_LOCAL_OLLAMA_DEV_MODEL_ID=qwen3:8b` |
| `MYCELIS_PROVIDER_<PROVIDER_ID>_ENDPOINT` | — | Override a provider endpoint. Example: `MYCELIS_PROVIDER_LOCAL_OLLAMA_DEV_ENDPOINT=http://192.168.50.156:11434/v1` |
| `MYCELIS_PROVIDER_<PROVIDER_ID>_ENABLED` | — | Enable/disable a provider from automation tools using `true` / `false` |
| `MYCELIS_PROVIDER_<PROVIDER_ID>_TYPE` | — | Define a provider type from env for deployment-created providers (`ollama`, `openai_compatible`, `anthropic`, `google`) |
| `MYCELIS_PROVIDER_<PROVIDER_ID>_API_KEY` | — | Override a provider API key directly when automation tooling must inject it |
| `MYCELIS_PROVIDER_<PROVIDER_ID>_API_KEY_ENV` | — | Point a provider at another env var containing the API key |
| `MYCELIS_PROVIDER_<PROVIDER_ID>_TOKEN_BUDGET_PROFILE` | — | Override the provider output-budget preset (`conservative`, `standard`, `extended`, `deep`) |
| `MYCELIS_PROVIDER_<PROVIDER_ID>_MAX_OUTPUT_TOKENS` | — | Override the provider max output budget directly |
| `MYCELIS_PROFILE_<PROFILE>_PROVIDER` | — | Route a profile to a provider from env. Example: `MYCELIS_PROFILE_CHAT_PROVIDER=local_ollama_dev` |
| `MYCELIS_MEDIA_ENDPOINT` | — | Override the image/media engine endpoint |
| `MYCELIS_MEDIA_MODEL_ID` | — | Override the image/media model id |
| `MYCELIS_MEDIA_API_KEY_ENV` | — | Point a hosted or protected media provider at an env var containing the API key |
| **Database** | | |
| `DB_HOST` | `127.0.0.1` | PostgreSQL host (localhost via bridge, or K8s service name in-cluster) |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `mycelis` | Database user |
| `DB_PASSWORD` | `password` | Database password |
| `DB_NAME` | `cortex` | Database name |
| `POSTGRES_USER` | `mycelis` | Helm chart injection (must match DB_USER) |
| `POSTGRES_PASSWORD` | `password` | Helm chart injection (must match DB_PASSWORD) |
| `POSTGRES_DB` | `cortex` | Helm chart injection (must match DB_NAME) |
| **Core Server** | | |
| `PORT` | `8081` | HTTP listen port (8081 local, 8080 in-cluster) |
| `NATS_URL` | `nats://127.0.0.1:4222` | NATS connection string |
| `LOG_LEVEL` | `debug` | Log verbosity |
| `CORS_ORIGIN` | `http://localhost:3000` | Allowed CORS origin (Next.js dev server) |
| `DATA_DIR` | `/data/artifacts` | Agent artifact storage path |
| `MYCELIS_WORKSPACE` | `./workspace` | Workspace sandbox root for manifested files and file tools. In Kubernetes use `/data/workspace`. |
| `MYCELIS_LOCAL_ADMIN_USERNAME` | `admin` | Display/identity name for the primary self-hosted local admin principal behind `MYCELIS_API_KEY` |
| `MYCELIS_LOCAL_ADMIN_USER_ID` | `00000000-0000-0000-0000-000000000000` | Stable internal id for the primary self-hosted local admin principal |
| `MYCELIS_BREAK_GLASS_API_KEY` | — | Optional dedicated break-glass API key for self-hosted recovery when using hybrid or federated posture |
| `MYCELIS_BREAK_GLASS_USERNAME` | `recovery-admin` | Display/identity name for the break-glass recovery principal |
| `MYCELIS_BREAK_GLASS_USER_ID` | `00000000-0000-0000-0000-000000000001` | Stable internal id for the break-glass recovery principal |
| **Next.js Proxy** | | |
| `MYCELIS_API_HOST` | `localhost` | Backend host for Next.js rewrite proxy |
| `MYCELIS_API_PORT` | `8081` | Backend port for Next.js rewrite proxy |
| `MYCELIS_INTERFACE_HOST` | `127.0.0.1` | Local UI probe/base host used by browser automation and lifecycle checks |
| `MYCELIS_INTERFACE_BIND_HOST` | `::` | UI bind host; defaults to dual-stack listening so IPv4 localhost, IPv6 localhost, and LAN clients can reach the UI |
| `MYCELIS_INTERFACE_PORT` | `3000` | UI port |
| **Commercial LLM (Optional)** | | |
| `OPENAI_API_KEY` | — | OpenAI API key |
| `ANTHROPIC_API_KEY` | — | Anthropic API key |
| `GEMINI_API_KEY` | — | Google Gemini API key |
| **GitHub** | | |
| `GH_TOKEN` | — | GitHub PAT for integrations |

> **Windows gotcha:** If Ollama is installed locally, it sets `OLLAMA_HOST=0.0.0.0` as a **Windows User environment variable** (listen address). The Go server detects this and skips endpoint patching. Your `.env` value will be used via the YAML config instead.

> **Deployment automation rule:** Prefer `MYCELIS_PROVIDER_<PROVIDER_ID>_*`, `MYCELIS_PROFILE_<PROFILE>_PROVIDER`, and `MYCELIS_MEDIA_*` for automation-driven cognitive config overrides. Do not rely on the retired `MYCELIS_TEAM_PROVIDER_MAP` / `MYCELIS_AGENT_PROVIDER_MAP` env maps.
>
> **Architecture boundary:** These env overrides are for deployment-time provider definition, environment-specific endpoint/model configuration, and profile default wiring only. They are not runtime organization behavior, do not control team/role routing directly, and do not replace bundle-defined truth.
>
> **Identity boundary:** `MYCELIS_API_KEY` remains the primary local-admin credential. `MYCELIS_BREAK_GLASS_API_KEY` is optional and should be used only for explicit self-hosted recovery posture, not as the normal everyday token.

### `core/config/cognitive.yaml` — LLM Provider Routing

Defines which LLM providers are available and which profiles route to them.

Python dependency resolution is pinned by the tracked root `uv.lock`. Regenerate it intentionally with `uv lock` after changing `pyproject.toml`, then run `uv lock --check` before release handoff.

```yaml
providers:
  ollama:
    type: "openai_compatible"
    endpoint: "http://127.0.0.1:11434/v1"   # <-- Change for remote Ollama
    model_id: "qwen2.5-coder:7b"
    api_key: "ollama"
  vllm:
    type: "openai_compatible"
    endpoint: "http://127.0.0.1:8000/v1"
    model_id: "qwen2.5-coder"
    api_key: "mycelis-local"
  # ... production_gpt4, production_claude, production_gemini

profiles:
  admin: "ollama"     # Which provider handles admin chat
  architect: "ollama" # Which provider handles architect tasks
  coder: "ollama"     # Which provider handles coding
  creative: "ollama"  # Which provider handles creative tasks
  sentry: "ollama"    # Which provider handles security review
  chat: "ollama"      # Which provider handles general chat
```

**Change provider routing:** Edit the `profiles` section or use the UI at `/settings` → **AI Engines** (Advanced mode).
By default, startup probes focus on `ollama`. Additional backends should be explicitly enabled and profile-routed before Mycelis attempts startup connectivity checks.

Token budget guidance:
- `conservative`: 512 max output tokens for terse, low-cost work
- `standard`: 1024 max output tokens for everyday local agentry and UI-facing interaction
- `extended`: 2048 max output tokens for longer coding, planning, or hosted-provider usage
- `deep`: 4096 max output tokens for complex reasoning where the host and provider budget support it

Management rule:
- keep token budgets in `cognitive.yaml`, env overrides, or `/settings` -> **AI Engines** (Advanced mode)
- prefer the preset profile first, then override `max_output_tokens` only when a provider needs a specific cap
- safe defaults should stay bounded; do not assume every provider should default to deep output

Automation override examples:

```bash
MYCELIS_PROVIDER_LOCAL_OLLAMA_DEV_MODEL_ID=qwen3:8b
MYCELIS_PROVIDER_LOCAL_OLLAMA_DEV_ENABLED=true
MYCELIS_PROVIDER_LOCAL_OLLAMA_DEV_TOKEN_BUDGET_PROFILE=standard
MYCELIS_PROVIDER_LOCAL_OLLAMA_DEV_MAX_OUTPUT_TOKENS=1024
MYCELIS_PROFILE_CHAT_PROVIDER=local_ollama_dev
MYCELIS_PROFILE_CODER_PROVIDER=local_ollama_dev
```

Runtime truth still follows:

```text
Bundle -> Instantiated Organization -> Inheritance -> Routing
```

Hosted-provider auth inventory:

| Provider | Auth env | Runtime auth posture |
| :--- | :--- | :--- |
| OpenAI | `OPENAI_API_KEY` | Bearer auth against `https://api.openai.com/v1` |
| Anthropic | `ANTHROPIC_API_KEY` | `x-api-key` plus Anthropic API version header |
| Gemini | `GEMINI_API_KEY` | `x-goog-api-key` header against Gemini REST endpoints |

Local-provider switching inventory:

| Provider | Default endpoint | Default posture |
| :--- | :--- | :--- |
| Ollama | `http://127.0.0.1:11434/v1` | shipped default |
| vLLM | `http://127.0.0.1:8000/v1` | optional local high-throughput server |
| LM Studio | `http://127.0.0.1:1234/v1` | optional desktop GUI local server |

If you change local engines:
- keep `endpoint`, `model_id`, `enabled`, and profile routing aligned together
- when using repo-local vLLM helpers, keep `cognitive/config/engine.yaml` and `core/config/cognitive.yaml` pointing at the same host, port, and API key
- use `/settings` -> **AI Engines** for the product-facing inventory, or edit `core/config/cognitive.yaml` directly for file-backed local control

## Deployment Guidance By Host Architecture

- Windows x86_64:
  - best fit for local development, operator workflow iteration, and desktop Docker Desktop / local-Kubernetes usage
  - move tool caches off a tight `C:` volume early with `uv run inv cache.apply-user-policy`
  - use remote Ollama or hosted providers when the desktop is not meant to be the long-running inference host
- Linux x86_64:
  - preferred for longer-running Core, Postgres, NATS, and container/Helm-driven environments
  - keep repo-managed caches and workspace data on the volume intended for repeated build/test/deploy cycles
  - use the same env override contract for deployment automation rather than forking config files per host
- Linux arm64:
  - appropriate for lighter control-host, edge-host, or remote-provider-connected setups
  - do not assume local heavyweight Ollama/model serving is a good fit on smaller ARM boards; prefer remote provider endpoints unless the host has been validated
  - keep the bootstrap bundle and instantiated-organization path identical to other hosts so architecture truth does not fork by platform
- Mixed-architecture deployments:
  - build binaries/images for the target architecture instead of reusing desktop-local artifacts blindly
  - keep deployment concerns in env/files/config, while runtime behavior still comes from `Bundle -> Instantiated Organization -> Inheritance -> Routing`
  - use `MYCELIS_PROVIDER_<PROVIDER_ID>_*`, `MYCELIS_PROFILE_<PROFILE>_PROVIDER`, and `MYCELIS_MEDIA_*` to stamp environment-specific provider wiring without changing organizational routing rules

## Recommended Host Paths

### WSL2 / Linux / macOS

This is the recommended easiest path for a fresh full-stack setup:

```bash
cp .env.compose.example .env.compose
uv run inv install
uv run inv compose.up --build
uv run inv compose.health
```

Use this path when you want:
- the easiest supported full-stack bring-up
- a single-host runtime for development, demos, or home-lab use
- to avoid local Kubernetes/bridge complexity unless you are specifically working on Kubernetes behavior

### Windows Native

This remains a supported main development path:

```powershell
Copy-Item .env.example .env
uv run inv install
uv run inv k8s.up
uv run inv lifecycle.up --frontend
uv run inv lifecycle.health
```

Use this path when you want:
- the normal Windows desktop development experience
- direct validation of the local Kubernetes path (`k3d` preferred, `MYCELIS_K8S_BACKEND=kind` fallback)
- a Windows operator/browser station that reaches a self-hosted runtime and an explicit Windows-hosted AI endpoint

### Cross-Host Working Copy Rule

Do not share one long-lived generated environment across Windows and WSL/Linux/macOS toolchains.

If you switch host environments:
- recreate `.venv`
- recreate `interface/node_modules`
- recreate `interface/.next`

## WSL/Linux Codex Handoff

Use this as the canonical resume path when active development moves to WSL/Linux/macOS:

1. Start from a WSL-native clone or worktree, not a mixed Windows-generated environment.
2. Recreate `.venv`, `interface/node_modules`, and `interface/.next` if the checkout was ever used from Windows.
3. Copy `.env.compose.example` to `.env.compose` and set `MYCELIS_COMPOSE_OLLAMA_HOST` deliberately for the real model host.
4. Run `uv run inv auth.posture --compose`.
5. Run `uv run inv install`.
6. For personal-owner or data-plane-first validation, use `uv run inv compose.infra-up --wait-timeout=180`, `uv run inv compose.infra-health`, `uv run inv compose.migrate`, and `uv run inv compose.storage-health`.
7. For the normal full-stack path, use `uv run inv compose.up --build --wait-timeout=240`, then `uv run inv compose.health`.
8. Use `uv run inv ci.baseline`, `uv run inv ci.service-check --live-backend`, and `uv run inv ci.release-preflight --runtime-posture --service-health --live-backend` as the canonical validation gate; the runtime-posture step reads process env plus `.env.compose` / `.env` and fails closed when no explicit supported AI endpoint is configured.

WSL/Linux notes:
- treat `uv run inv ...` as the only normal execution path; raw `npx`, direct Playwright, and PowerShell wrappers are fallback troubleshooting paths only
- when Windows no longer has a native `docker` binary, `uv run inv compose.*` can drive Docker through WSL automatically; set `MYCELIS_WSL_DISTRO` if the default WSL distro is not the one that owns Docker
- on that Windows + WSL Docker path, `compose.up` and `compose.health` can auto-start a WSL-host relay for `MYCELIS_COMPOSE_OLLAMA_HOST` so bridge containers can still reach a Windows-hosted Ollama service through `host.docker.internal`
- use `MYCELIS_BACKEND_WORKSPACE_ROOT=workspace/docker-compose/data/workspace` for Compose-backed live browser specs when the spec checkout differs from the running backend worktree
- `psql` is required for `db.*` tasks, but pure `compose.*` workflows run migrations and health checks through the Postgres container
- `host.docker.internal` is usually correct for Docker Desktop + WSL, but native Linux Docker hosts may need another hostname or reachable service address instead

Safest posture:
- use a separate clone or worktree per host environment when you actively move between Windows and WSL

### `core/config/templates/*.yaml` — Bootstrap Template Bundles

| File | Purpose |
|:--|:--|
| `v8-migration-standing-team-bridge.yaml` | Transitional V8 migration bundle that carries standing-team org/team/agent data as a self-contained startup bundle |

These bundles are the first implementation bridge from the V8 bootstrap model into runtime-readable config.
Startup now instantiates the runtime organization only through a selected bundle.
The deployed Core image resolves those files from `/core/config`, so the Helm chart mounts the runtime config volume there and the container workdir must match that contract for bootstrap startup to succeed.
Core fails closed when no valid bundle exists in `core/config/templates/`, and operators must set `MYCELIS_BOOTSTRAP_TEMPLATE_ID` whenever more than one bundle is present. `core/config/teams/*.yaml` remains migration/reference input, not a normal startup path.

### `core/config/teams/*.yaml` — Standing Teams

| File | Team | Agents |
|:--|:--|:--|
| `admin.yaml` | Admin | admin (17 tools, 5 ReAct iterations) |
| `council.yaml` | Council | council-architect, council-coder, council-creative, council-sentry |
| `prime-architect.yaml` | Prime Architect | architecture lead for gated delivery and review flows |
| `prime-development.yaml` | Prime Development | implementation lead for delivery execution and testing follow-through |
| `agui-design-architect.yaml` | AGUI Design Architect | workflow-composer and UI/UX architecture lead |
| `genesis.yaml` | Genesis Core | architect, commander |
| `telemetry.yaml` | Telemetry Core | observer |

These YAML files are transitional migration/reference inputs mirrored for packaging and migration work, but they are not part of the normal startup path.
Canonical bus signal conventions for these teams live in `docs/architecture/NATS_SIGNAL_STANDARD_V7.md`.

### `core/config/policy.yaml` — Governance Rules

Defines approval thresholds, deny rules, and safety constraints.

## Quick Start Paths

Use the shortest path that matches your host:

- WSL2/Linux/macOS:
  1. `cp .env.compose.example .env.compose`
  2. `uv run inv install`
  3. `uv run inv compose.up --build`
  4. `uv run inv compose.health`
  5. open `http://localhost:3000`
- Windows native:
  1. `copy .env.example .env`
  2. `uv run inv install`
  3. `uv run inv k8s.up`
  4. `uv run inv lifecycle.up --frontend`
  5. `uv run inv lifecycle.health`
  6. open `http://localhost:3000`

## First-Time Setup

Run all commands from `scratch/` (project root where `tasks.py` lives).

Choose one supported local runtime:
- `k3d/local Kubernetes` for chart and cluster parity
- `Docker Compose` for home-lab, single-host, and partner-demo use

Docker Compose rule:
- keep `.env.compose` separate from `.env`
- use `MYCELIS_COMPOSE_OLLAMA_HOST` in `.env.compose` so Windows/macOS host-level `OLLAMA_HOST` bind settings do not override the container runtime
- `MYCELIS_COMPOSE_OLLAMA_HOST` must resolve from inside the Core container, so use `http://host.docker.internal:11434` or another reachable service hostname instead of `localhost`
- if Docker is running inside WSL and Ollama is on the same Windows host, keep `MYCELIS_COMPOSE_OLLAMA_HOST` pointed at the intended Windows service address; the task layer can relay that through the WSL host when bridge containers cannot reach the Windows LAN IP directly

### Option A: k3d / Local Kubernetes

```bash
# 1. Configure environment
cp .env.example .env
# Edit .env — set DB credentials, API key, and any optional runtime overrides.
# Optional promoted preset:
#   export MYCELIS_K8S_VALUES_FILE=charts/mycelis-core/values-k3d.yaml
# For an external Windows-hosted text model service:
#   export MYCELIS_K8S_TEXT_ENDPOINT=http://192.168.x.x:11434/v1
# Optional external media endpoint:
#   export MYCELIS_K8S_MEDIA_ENDPOINT=http://192.168.x.x:8001/v1

# 2. Bring up cluster services in dependency order
uv run inv k8s.up
# Optional legacy fallback:
#   MYCELIS_K8S_BACKEND=kind uv run inv k8s.up
# Order enforced by task:
#   preferred local backend (`k3d` when available) -> Helm deploy -> PostgreSQL ready -> NATS ready -> Core API ready

# 3. Install frontend dependencies (first time only)
uv run inv interface.install

# 4. Bring up the managed local bridge + backend + frontend
uv run inv lifecycle.up --frontend

# 5. Initialize the database
uv run inv db.migrate        # Apply canonical forward migrations (001_init_memory.sql + *.up.sql)

# 6. Verify the stack
uv run inv lifecycle.status
uv run inv lifecycle.health
```

Open [http://localhost:3000](http://localhost:3000).

Enterprise preflight render examples:

```bash
helm template mycelis-core charts/mycelis-core -f charts/mycelis-core/values-enterprise.yaml
helm template mycelis-core charts/mycelis-core -f charts/mycelis-core/values-enterprise-windows-ai.yaml
```

### Option B: Docker Compose Home Runtime

```bash
# 1. Configure compose-specific environment
cp .env.compose.example .env.compose
# Edit .env.compose — set MYCELIS_API_KEY and adjust MYCELIS_COMPOSE_OLLAMA_HOST if needed

# 2. Bring up the full single-host stack
uv run inv compose.up --build
```

Compose bring-up order is managed for you:
- PostgreSQL + NATS first
- canonical migrations through the PostgreSQL container
- Core + Interface second
- host health verification last

## Binary Release Process

Use these commands when you need a distributable Core binary instead of a local dev build:

```bash
uv run inv core.package
uv run inv core.package --target-os=windows --target-arch=amd64 --version-tag=v0.1.0
```

What it does:
- cross-compiles a versioned Core binary
- writes a versioned archive under `dist/`
- includes a small release `README.txt` beside the binary
- pair this with `uv run inv ci.release-preflight --service-health --live-backend` before you hand the binary to another machine for release checkout testing
- use [Remote User Testing](./REMOTE_USER_TESTING.md) for the second-machine walkthrough sequence and [Testing](./TESTING.md) for the authoritative gate order

Canonical automation:
- `.github/workflows/release-binaries.yaml`
- runs on `v*` tag pushes
- supports manual dispatch for one-off packaging

## Daily Startup Sequence

### Docker Compose Alternative

For the home-runtime path, the canonical daily cycle is:

```bash
uv run inv compose.up
uv run inv compose.status
uv run inv compose.health
```

Shutdown:

```bash
uv run inv compose.down
```

k3d/local Kubernetes daily path:

```bash
uv run inv k8s.up
uv run inv lifecycle.up --frontend
uv run inv db.migrate
uv run inv lifecycle.status
uv run inv lifecycle.health
```

Use `uv run inv k8s.bridge` only when you intentionally need a manual long-running port-forward outside the managed lifecycle path.

Free terminal for running tasks:

```bash
uv run inv db.status        # Check database tables
uv run inv interface.check  # Smoke-test running pages
uv run inv core.test        # Run Go unit tests
```

## Port Map

| Service | Port | Source | Protocol |
|:--|:--|:--|:--|
| Next.js (frontend) | **3000** | Native | HTTP |
| Go Core (backend) | **8081** | Native | HTTP |
| PostgreSQL | **5432** | Local-Kubernetes bridge or Compose publish | TCP |
| NATS | **4222** | Local-Kubernetes bridge or Compose publish | TCP |
| NATS Monitor | **8222** | Compose publish | HTTP |
| Ollama | **11434** | Local / LAN | HTTP |
| vLLM (optional) | **8000** | Native | HTTP |
| Media Server (optional) | **8001** | Native | HTTP |
| Local Kubernetes API Server | **6443** | Docker | TCP |

### Architecture: How Requests Flow

```
Browser (3000) → Next.js rewrite proxy → Go backend (8081) → NATS (4222) → Agent
                                       → PostgreSQL (5432)
                                       → Ollama (11434) via Cognitive Router
```

Persistent storage contract:
- local development typically uses `MYCELIS_WORKSPACE=./workspace`
- Kubernetes should mount the PVC at `/data`
- manifested files and MCP filesystem access should use `MYCELIS_WORKSPACE=/data/workspace`
- artifact blobs should use `DATA_DIR=/data/artifacts`
- Compose output blocks use `MYCELIS_OUTPUT_BLOCK_MODE=local_hosted` with `MYCELIS_OUTPUT_HOST_PATH` pointing at the host directory mounted as Core `/data`; use `cluster_generated` for chart/PVC-owned cluster output storage
- when `local_hosted` is selected, create the host directory first. The Invoke compose task resolves the path with Python `pathlib`, including `~`, environment variables, Windows drive paths, Linux paths, macOS paths, and paths with spaces

The Next.js `next.config.ts` rewrites `/api/*` to `http://{MYCELIS_API_HOST}:{MYCELIS_API_PORT}/api/*`.

## Health Checks

Quick commands to verify each service:

```bash
# Local Kubernetes cluster + pods
uv run inv k8s.status

# Compose stack
uv run inv compose.status
uv run inv compose.health

# Backend health
curl http://localhost:8081/healthz

# NATS (TCP check — NATS doesn't speak HTTP)
# Windows:
powershell -NoProfile -Command "Test-NetConnection localhost -Port 4222 -InformationLevel Quiet"
# Linux:
nc -z localhost 4222

# PostgreSQL
uv run inv db.status

# Frontend
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/dashboard

# Ollama models
curl http://127.0.0.1:11434/api/tags

# Full smoke test (9 pages, checks for 404s, 500s, hydration errors)
uv run inv interface.check
```

## Troubleshooting

### "Council member admin did not respond: nats: connection closed"

**Cause:** NATS port-forward died. This is the most common failure.

**Fix:** Restart the bridge, then restart the backend:

```bash
# Quick fix — just NATS:
kubectl port-forward -n mycelis svc/mycelis-core-nats 4222:4222

# Or restart full bridge:
uv run inv k8s.bridge

# Then restart backend to reconnect:
uv run inv core.restart
```

### "Agent unavailable — no cognitive engine"

**Cause:** Ollama/LLM provider unreachable. The backend connected to NATS but can't reach any LLM.

**Fix:**
1. Check Ollama is running: `curl http://127.0.0.1:11434/`
2. Check `.env` has `OLLAMA_HOST=http://127.0.0.1:11434` (not `0.0.0.0`)
3. Check `core/config/cognitive.yaml` — `ollama.endpoint` should point to the reachable address
4. Restart backend: `uv run inv core.restart`

### "ModuleNotFoundError: No module named 'dotenv'"

**Cause:** Running a dependency-heavy task outside the project invoke environment.

**Fix:** Use `uv run inv` so the task runs inside the project-managed environment with the expected dependencies.

### "STACK UP FAILED: Core did not open its port in time"

**Cause:** Core exited or stalled before binding `:8081`.

**Fix:**
1. Inspect `workspace/logs/core-startup.log`
2. If the log points to MCP bootstrap/connect issues, rerun after fixing the failing server config
3. Recheck with `uv run inv lifecycle.status`

### Kind fallback cluster cert chain broken

**Cause:** Docker Desktop restart can invalidate Kind cluster certificates when you are intentionally using the Kind fallback backend.

**Fix:**
```bash
uv run inv k8s.reset   # Full teardown + reinit + deploy
```

### Port already in use

```bash
uv run inv core.stop       # Kill Go backend (port 8081)
uv run inv interface.stop  # Kill Next.js (port 3000)
```

### UnicodeEncodeError in invoke output (Windows)

**Cause:** Go server output contains emoji characters (✅🧠⚡) that Windows cp1252 console can't encode. Invoke's stdout handler crashes.

**Impact:** The server is still running — only the invoke wrapper's output display fails.

**Workaround:** Run the server directly if invoke crashes:
```bash
cd core
set OLLAMA_HOST=http://127.0.0.1:11434
set DB_HOST=127.0.0.1
set PORT=8081
set NATS_URL=nats://127.0.0.1:4222
bin\server.exe
```

### kubectl port-forward version mismatch

**Cause:** kubectl v1.32 has bugs with K8s v1.33 port-forwarding.

**Fix:** Upgrade kubectl to v1.35+:
```bash
scoop update kubectl   # Windows
brew upgrade kubectl   # macOS
```

## Cognitive Engine Setup

### Option A: Local Ollama (Recommended)

1. Install Ollama from [ollama.com](https://ollama.com/)
2. Pull models:
   ```bash
   ollama pull qwen2.5-coder:7b-instruct    # Code generation
   ollama pull nomic-embed-text              # Embeddings (RAG)
   ```
3. Set in `.env`: `OLLAMA_HOST=http://127.0.0.1:11434`
4. Set in `cognitive.yaml`: `ollama.endpoint: "http://127.0.0.1:11434/v1"`

### Option B: Remote Ollama (LAN)

Same as above but use the remote machine's IP:
```
OLLAMA_HOST=http://192.168.x.x:11434
```

### Option C: vLLM (GPU inference)

```bash
uv run inv install --optional-engines    # Or run cognitive.install directly for engine-only extras
uv run inv cognitive.install    # Install vLLM + deps
uv run inv cognitive.llm        # Start on port 8000
```

Then change profiles in `cognitive.yaml` to `"vllm"`.

vLLM contract notes:
- the repo-local helper starts the OpenAI-compatible server on `http://127.0.0.1:8000/v1`
- the helper uses the API key from `cognitive/config/engine.yaml` (`mycelis-local` by default)
- `core/config/cognitive.yaml` should keep the same key in the `vllm.api_key` field unless you change the helper config too
- repo-local `cognitive.*` helpers are intended for supported Linux GPU hosts; on Windows, use Ollama locally or target a remote OpenAI-compatible vLLM endpoint instead

Example switch:

```yaml
providers:
  vllm:
    type: "openai_compatible"
    endpoint: "http://127.0.0.1:8000/v1"
    model_id: "qwen2.5-coder"
    api_key: "mycelis-local"
    enabled: true

profiles:
  coder: "vllm"
  architect: "vllm"
  chat: "ollama"
```

## Full Command Reference

| Command | Description |
|:--|:--|
| **Core** | |
| `uv run inv core.compile` | Compile Go binary only |
| `uv run inv core.package` | Package a versioned Core binary archive under `dist/` |
| `uv run inv core.build` | Compile Go binary + Docker image |
| `uv run inv core.test` | Go unit tests (`go test ./...`) |
| `uv run inv core.run` | Start backend (foreground) |
| `uv run inv core.stop` | Kill backend |
| `uv run inv core.restart` | Stop + Run |
| `uv run inv core.smoke` | Governance smoke tests |
| **Interface** | |
| `uv run inv interface.dev` | Start Next.js dev server |
| `uv run inv interface.build` | Production build |
| `uv run inv interface.test` | Vitest unit tests |
| `uv run inv interface.typecheck` | TypeScript type check |
| `uv run inv interface.e2e` | Playwright E2E tests (self-managed Next.js server lifecycle; add `--live-backend` for real Core-backed UI flows) |
| `uv run inv interface.check` | Smoke-test running pages |
| `uv run inv interface.stop` | Kill dev server |
| `uv run inv interface.clean` | Clear `.next` cache |
| `uv run inv interface.restart` | Full restart cycle |
| **Database** | |
| `uv run inv db.migrate` | Apply canonical forward migrations (`001_init_memory.sql` + `*.up.sql`) only when the target schema is not already initialized; otherwise skip replay and use `db.reset` for a clean rebuild |
| `uv run inv db.reset` | Drop + recreate + migrate |
| `uv run inv db.status` | Show tables |
| **Infrastructure** | |
| `uv run inv k8s.init` | Create preferred local Kubernetes cluster (`k3d` when available, Kind fallback) |
| `uv run inv k8s.up` | Canonical cluster bring-up: init → deploy → wait |
| `uv run inv k8s.deploy` | Helm deploy to cluster |
| `uv run inv k8s.wait` | Wait for readiness gates (PostgreSQL → NATS → Core API) |
| `uv run inv k8s.bridge` | Port-forward NATS, PG, API |
| `uv run inv k8s.status` | Local Kubernetes health check |
| `uv run inv k8s.recover` | Restart core + infra resources (core, NATS, PostgreSQL) |
| `uv run inv k8s.reset` | Full teardown + canonical bring-up (includes readiness wait) |
| `uv run inv compose.infra-up` | Start only Compose PostgreSQL + NATS for personal-owner data-plane preflight |
| `uv run inv compose.infra-health` | Probe only Compose PostgreSQL/NATS without requiring Core or Interface |
| `uv run inv compose.storage-health` | Probe post-migration Compose PostgreSQL long-term storage for pgvector, memory/context, artifacts, exchange, and continuity |
| `uv run inv compose.up` | Managed Docker Compose bring-up: postgres/nats -> migrate -> core/interface |
| `uv run inv compose.down` | Stop the Docker Compose stack |
| `uv run inv compose.status` | Compose service + host-port status |
| `uv run inv compose.health` | Deep health probe for the compose runtime |
| `uv run inv compose.logs` | Tail compose logs for all services or one service |
| **CI** | |
| `uv run inv ci.check` | Full pipeline: lint → test → build |
| **Cognitive** | |
| `uv run inv cognitive.up` | Start optional local vLLM + Media servers |
| `uv run inv cognitive.stop` | Kill optional local cognitive processes |
| `uv run inv cognitive.status` | Health check optional local providers |


