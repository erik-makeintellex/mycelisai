# Local Development Workflow

> **Quick Start:** Already set up? Jump to [Daily Startup](#daily-startup-sequence).

## Prerequisites

| Tool | Version | Install | Purpose |
|:--|:--|:--|:--|
| Docker Desktop | Latest | [docker.com](https://www.docker.com/) | Container runtime for Kind cluster |
| Kind | Latest | `scoop install kind` / `brew install kind` | Local Kubernetes cluster |
| kubectl | **v1.35+** | `scoop install kubectl` / `brew install kubectl` | K8s CLI (v1.32 has port-forward bugs) |
| Helm | v3+ | `scoop install helm` / `brew install helm` | Chart deployment |
| Go | 1.26 | [go.dev](https://go.dev/) | Backend compiler |
| Node.js | 20+ | [nodejs.org](https://nodejs.org/) | Frontend runtime |
| uv | Latest | `pip install uv` / `pipx install uv` | Python environment + task runner |
| psql | 16+ | Comes with PostgreSQL or standalone | Database migrations |
| Ollama | Latest | [ollama.com](https://ollama.com/) | Local LLM inference |

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
| **Next.js Proxy** | | |
| `MYCELIS_API_HOST` | `localhost` | Backend host for Next.js rewrite proxy |
| `MYCELIS_API_PORT` | `8081` | Backend port for Next.js rewrite proxy |
| **Commercial LLM (Optional)** | | |
| `OPENAI_API_KEY` | — | OpenAI API key |
| `ANTHROPIC_API_KEY` | — | Anthropic API key |
| `GEMINI_API_KEY` | — | Google Gemini API key |
| **GitHub** | | |
| `GH_TOKEN` | — | GitHub PAT for integrations |

> **Windows gotcha:** If Ollama is installed locally, it sets `OLLAMA_HOST=0.0.0.0` as a **Windows User environment variable** (listen address). The Go server detects this and skips endpoint patching. Your `.env` value will be used via the YAML config instead.

### `core/config/cognitive.yaml` — LLM Provider Routing

Defines which LLM providers are available and which profiles route to them.

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

**Change provider routing:** Edit the `profiles` section or use the UI at `/settings` → **Cognitive Matrix**.

### `core/config/teams/*.yaml` — Standing Teams

| File | Team | Agents |
|:--|:--|:--|
| `admin.yaml` | Admin | admin (17 tools, 5 ReAct iterations) |
| `council.yaml` | Council | council-architect, council-coder, council-creative, council-sentry |
| `genesis.yaml` | Genesis Core | architect, commander |
| `telemetry.yaml` | Telemetry Core | observer |

Add a YAML file, restart the server, and the new team is automatically loaded.

### `core/config/policy.yaml` — Governance Rules

Defines approval thresholds, deny rules, and safety constraints.

## First-Time Setup

Run all commands from `scratch/` (project root where `tasks.py` lives).

```bash
# 1. Configure environment
cp .env.example .env
# Edit .env — set OLLAMA_HOST, DB credentials, etc.

# 2. Create Kind cluster + deploy infrastructure (PostgreSQL, NATS)
uvx inv k8s.reset

# 3. Open the development bridge (port-forwards)
# This needs its own terminal — it stays running
uvx inv k8s.bridge

# 4. Initialize the database
uvx inv db.migrate        # Apply all 21 migrations (idempotent)

# 5. Build the Go backend
uvx inv core.build        # Compile binary + Docker image

# 6. Start the backend (new terminal — stays running)
uvx inv core.run

# 7. Install frontend dependencies (first time only)
uvx inv interface.install

# 8. Start the frontend (new terminal — stays running)
uvx inv interface.dev
```

Open [http://localhost:3000](http://localhost:3000).

## Daily Startup Sequence

Four terminals, in order:

### Terminal 1: Development Bridge

Port-forwards NATS, PostgreSQL, and the in-cluster API from Kind to localhost.

```bash
uvx inv k8s.bridge
```

> On Windows, this opens 3 new CMD windows (one per port-forward). Keep them open.

### Terminal 2: Go Backend

```bash
uvx inv core.run     # Foreground, blocking. Ctrl+C to stop.
```

### Terminal 3: Next.js Frontend

```bash
uvx inv interface.dev
```

### Terminal 4: Commands

Free terminal for running tasks:

```bash
uvx inv db.status        # Check database tables
uvx inv interface.check  # Smoke-test running pages
uvx inv core.test        # Run Go unit tests
```

## Port Map

| Service | Port | Source | Protocol |
|:--|:--|:--|:--|
| Next.js (frontend) | **3000** | Native | HTTP |
| Go Core (backend) | **8081** | Native | HTTP |
| PostgreSQL | **5432** | Kind bridge | TCP |
| NATS | **4222** | Kind bridge | TCP |
| Ollama | **11434** | Local / LAN | HTTP |
| vLLM (optional) | **8000** | Native | HTTP |
| Media Server (optional) | **8001** | Native | HTTP |
| Kind API Server | **6443** | Docker | TCP |

### Architecture: How Requests Flow

```
Browser (3000) → Next.js rewrite proxy → Go backend (8081) → NATS (4222) → Agent
                                       → PostgreSQL (5432)
                                       → Ollama (11434) via Cognitive Router
```

The Next.js `next.config.ts` rewrites `/api/*` to `http://{MYCELIS_API_HOST}:{MYCELIS_API_PORT}/api/*`.

## Health Checks

Quick commands to verify each service:

```bash
# Kind cluster + pods
uvx inv k8s.status

# Backend health
curl http://localhost:8081/healthz

# NATS (TCP check — NATS doesn't speak HTTP)
# Windows:
powershell -NoProfile -Command "Test-NetConnection localhost -Port 4222 -InformationLevel Quiet"
# Linux:
nc -z localhost 4222

# PostgreSQL
uvx inv db.status

# Frontend
curl -s -o /dev/null -w "%{http_code}" http://localhost:3000/dashboard

# Ollama models
curl http://127.0.0.1:11434/api/tags

# Full smoke test (9 pages, checks for 404s, 500s, hydration errors)
uvx inv interface.check
```

## Troubleshooting

### "Council member admin did not respond: nats: connection closed"

**Cause:** NATS port-forward died. This is the most common failure.

**Fix:** Restart the bridge, then restart the backend:

```bash
# Quick fix — just NATS:
kubectl port-forward -n mycelis svc/mycelis-core-nats 4222:4222

# Or restart full bridge:
uvx inv k8s.bridge

# Then restart backend to reconnect:
uvx inv core.restart
```

### "Agent unavailable — no cognitive engine"

**Cause:** Ollama/LLM provider unreachable. The backend connected to NATS but can't reach any LLM.

**Fix:**
1. Check Ollama is running: `curl http://127.0.0.1:11434/`
2. Check `.env` has `OLLAMA_HOST=http://127.0.0.1:11434` (not `0.0.0.0`)
3. Check `core/config/cognitive.yaml` — `ollama.endpoint` should point to the reachable address
4. Restart backend: `uvx inv core.restart`

### "ModuleNotFoundError: No module named 'dotenv'"

**Cause:** Using `uvx --from invoke inv` instead of `uv run inv`. The `uvx` ephemeral env doesn't include project deps.

**Fix:** Use `uv run inv` (uses project venv with all dependencies) or run `uv sync` first.

### Kind cluster cert chain broken

**Cause:** Docker Desktop restart invalidates Kind cluster certificates.

**Fix:**
```bash
uvx inv k8s.reset   # Full teardown + reinit + deploy
```

### Port already in use

```bash
uvx inv core.stop       # Kill Go backend (port 8081)
uvx inv interface.stop  # Kill Next.js (port 3000)
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
uvx inv cognitive.install    # Install vLLM + deps
uvx inv cognitive.llm        # Start on port 8000
```

Then change profiles in `cognitive.yaml` to `"vllm"`.

## Full Command Reference

| Command | Description |
|:--|:--|
| **Core** | |
| `uvx inv core.build` | Compile Go binary + Docker image |
| `uvx inv core.test` | Go unit tests (`go test ./...`) |
| `uvx inv core.run` | Start backend (foreground) |
| `uvx inv core.stop` | Kill backend |
| `uvx inv core.restart` | Stop + Run |
| `uvx inv core.smoke` | Governance smoke tests |
| **Interface** | |
| `uvx inv interface.dev` | Start Next.js dev server |
| `uvx inv interface.build` | Production build |
| `uvx inv interface.test` | Vitest unit tests |
| `uvx inv interface.e2e` | Playwright E2E tests |
| `uvx inv interface.check` | Smoke-test running pages |
| `uvx inv interface.stop` | Kill dev server |
| `uvx inv interface.clean` | Clear `.next` cache |
| `uvx inv interface.restart` | Full restart cycle |
| **Database** | |
| `uvx inv db.migrate` | Apply all migrations (idempotent) |
| `uvx inv db.reset` | Drop + recreate + migrate |
| `uvx inv db.status` | Show tables |
| **Infrastructure** | |
| `uvx inv k8s.init` | Create Kind cluster |
| `uvx inv k8s.deploy` | Helm deploy to cluster |
| `uvx inv k8s.bridge` | Port-forward NATS, PG, API |
| `uvx inv k8s.status` | Cluster health check |
| `uvx inv k8s.reset` | Full teardown + redeploy |
| **CI** | |
| `uvx inv ci.check` | Full pipeline: lint → test → build |
| **Cognitive** | |
| `uvx inv cognitive.up` | Start vLLM + Media servers |
| `uvx inv cognitive.stop` | Kill cognitive processes |
| `uvx inv cognitive.status` | Health check providers |
