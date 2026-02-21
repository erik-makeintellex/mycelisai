# Mycelis Cortex — Operations Manual

> **Load this doc when:** Deploying, configuring, testing, running CI, or managing infrastructure.
>
> **Related:** [Overview](OVERVIEW.md) | [Backend](BACKEND.md) | [Frontend](FRONTEND.md)

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

---

## II. Task Automation

### Master Registry

**File:** `tasks.py` — 12 collections, 60+ @task functions

**Run from:** `scratch/` (project root where tasks.py lives)

**Invocation:** Always use `uvx inv <namespace>.<task>` — never raw `inv` or `python -m invoke`

> **Tip:** If you `uv venv && .venv/Scripts/activate` (Windows) or `source .venv/bin/activate` (Linux), you can use `inv` directly.

### Core Tasks (`ops/core.py`)

| Command | Description |
|---------|-------------|
| `uvx inv core.build` | Compile Go binary + Docker image (returns TAG) |
| `uvx inv core.test` | `go test ./...` |
| `uvx inv core.run` | Run Core locally (foreground, blocking) |
| `uvx inv core.stop` | Kill running Core process (taskkill on Windows, pkill on Linux) |
| `uvx inv core.restart` | stop + run |
| `uvx inv core.smoke` | `go run ./cmd/smoke/main.go` (governance smoke tests) |
| `uvx inv core.clean` | `go clean`, remove bin/ |

### Interface Tasks (`ops/interface.py`)

| Command | Description |
|---------|-------------|
| `uvx inv interface.dev` | Start Turbopack dev server (stops existing first) |
| `uvx inv interface.install` | `npm install` |
| `uvx inv interface.build` | `npm run build` (production) |
| `uvx inv interface.lint` | `npm run lint` (ESLint) |
| `uvx inv interface.test` | `npm run test` (Vitest) |
| `uvx inv interface.test-coverage` | Vitest with V8 coverage |
| `uvx inv interface.e2e` | `npm run e2e` (Playwright, optional `--headed`) |
| `uvx inv interface.stop` | Kill process on port 3000 |
| `uvx inv interface.clean` | rm -rf .next cache |
| `uvx inv interface.restart` | stop → clean → build → dev → check |
| `uvx inv interface.check` | Smoke-test: 9 pages for 200 status, no SSR errors, no hydration issues, no `bg-white`/`bg-zinc`/`bg-slate` leaks |

### Database Tasks (`ops/db.py`)

| Command | Description |
|---------|-------------|
| `uvx inv db.migrate` | Apply all SQL migrations (idempotent) |
| `uvx inv db.reset` | Drop + recreate + migrate |
| `uvx inv db.status` | List tables + row counts |
| `uvx inv db.create` | Create cortex database (if not exists) |

### Kubernetes Tasks (`ops/k8s.py`)

| Command | Description |
|---------|-------------|
| `uvx inv k8s.init` | Create Kind cluster (handles Windows absolute paths) |
| `uvx inv k8s.deploy` | Build Core, load Docker image, helm upgrade (injects secrets from .env) |
| `uvx inv k8s.bridge` | Port-forward NATS:4222, HTTP:8081←8080, PG:5432 |
| `uvx inv k8s.status` | Check Docker, Kind cluster, pod status, PVC status |
| `uvx inv k8s.recover` | Restart NATS + neural-core deployments |
| `uvx inv k8s.reset` | Full reset: delete cluster → init → deploy |

### Cognitive Tasks (`ops/cognitive.py`)

| Command | Description |
|---------|-------------|
| `uvx inv cognitive.install` | `uv sync` dependencies |
| `uvx inv cognitive.llm` | Start vLLM text server (GPU, configurable) |
| `uvx inv cognitive.media` | Start Diffusers media server (OpenAI-compatible) |
| `uvx inv cognitive.up` | Start full stack (vLLM + Media, handles Ctrl+C) |
| `uvx inv cognitive.stop` | Kill all cognitive processes |
| `uvx inv cognitive.status` | HTTP probes to check vLLM + Media health |

### Test Tasks (`ops/test.py`)

| Command | Description |
|---------|-------------|
| `uvx inv test.all` | Run ALL unit tests (Core + Interface) |
| `uvx inv test.coverage` | Go coverage + Vitest V8 coverage |
| `uvx inv test.e2e` | Alias for interface.e2e |

### CI Tasks (`ops/ci.py`)

| Command | Description |
|---------|-------------|
| `uvx inv ci.lint` | Go vet + Next.js lint (both must pass) |
| `uvx inv ci.test` | Go tests + Interface tests |
| `uvx inv ci.build` | Go binary + Next.js production build (no Docker) |
| `uvx inv ci.check` | Full pipeline: lint → test → build (with stage timers) |
| `uvx inv ci.deploy` | Build + Docker + K8s deploy |

### Other Tasks

| Module | Commands | Purpose |
|--------|----------|---------|
| `ops/proto_relay.py` | `proto.generate` | Protobuf codegen (Go + Python) |
| `ops/proto_relay.py` | `relay.test`, `relay.demo` | Python SDK tests + demo |
| `ops/device.py` | `device.boot --id=ghost-01` | Device simulation (NATS announce) |
| `ops/misc.py` | `clean.legacy` | Remove legacy Makefile/docker-compose |
| `ops/misc.py` | `team.sensors`, `team.output`, `team.test` | Python agent teams |

### Config Module (`ops/config.py`)

**Purpose:** Cross-platform configuration utilities

**Constants:**
```python
CLUSTER_NAME = "mycelis-cluster"
NAMESPACE = "mycelis"
ROOT_DIR = project root
CORE_DIR, SDK_DIR = relative paths
API_HOST/PORT = localhost:8081 (env-overridable)
INTERFACE_HOST/PORT = localhost:3000 (env-overridable)
```

**Helpers:**
```python
is_windows() → bool
powershell(command: str) → str  # PowerShell with -NoProfile flag
```

**Environment Sanitization:** Clears Windows Store Python's global `VIRTUAL_ENV` on import.

---

## III. Development Workflow

### Local Development (3 terminals)

```bash
# Terminal 1: Infrastructure bridge
uvx inv k8s.bridge    # Port-forward NATS, API, Postgres

# Terminal 2: Backend
uvx inv core.run      # Go backend (foreground, port 8080 → bridged to 8081)

# Terminal 3: Frontend
uvx inv interface.dev # Next.js dev (Turbopack, port 3000)
```

Open [http://localhost:3000](http://localhost:3000)

### First-Time Setup

```bash
# 1. Configure secrets
cp .env.example .env
# Edit .env: POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB

# 2. Boot infrastructure
uvx inv k8s.reset     # Full cluster reset (Kind + Core + DB)

# 3. Open bridge (keep running)
uvx inv k8s.bridge

# 4. Initialize database
uvx inv db.migrate    # Apply all 21 migrations (idempotent)

# 5. Start backend
uvx inv core.run

# 6. Start frontend
uvx inv interface.install  # First time only
uvx inv interface.dev
```

### Configure Cognitive Engine

- **UI:** `/settings` → Cognitive Matrix tab → change provider routing, configure endpoints
- **MCP:** `/settings` → MCP Tools tab → install from curated library or manually
- **YAML:** Edit `core/config/cognitive.yaml` directly

---

## IV. Configuration System

### Configuration Files

| Config | Location | Managed Via |
|--------|----------|-------------|
| Cognitive (Bootstrap) | `core/config/cognitive.yaml` | UI (`/settings` → Matrix) or YAML |
| Standing Teams | `core/config/teams/*.yaml` | YAML (auto-loaded at startup) |
| MCP Servers | Database | UI (`/settings` → MCP Tools) or API |
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
| `MYCELIS_INTERFACE_HOST` | `localhost` | Next.js dev server host |
| `MYCELIS_INTERFACE_PORT` | `3000` | Next.js dev server port |
| `DB_HOST` | `mycelis-core-postgresql` | PostgreSQL host |
| `DB_PORT` | `5432` | PostgreSQL port |
| `DB_USER` | `mycelis` | PostgreSQL user |
| `DB_PASSWORD` | — | PostgreSQL password |
| `DB_NAME` | `cortex` | PostgreSQL database |
| `NATS_URL` | `nats://mycelis-core-nats:4222` | NATS broker |
| `PORT` | `8080` | Core HTTP listen port |
| `OLLAMA_HOST` | — | Override Ollama endpoint |
| `OPENAI_API_KEY` | — | OpenAI API key |
| `ANTHROPIC_API_KEY` | — | Anthropic API key |
| `GEMINI_API_KEY` | — | Gemini API key |

---

## V. Testing Strategy

### 5-Tier Verification

| Tier | Tool | Count | Speed | Command |
|------|------|-------|-------|---------|
| 1 | Go unit tests | ~112 tests | <5s | `uvx inv core.test` |
| 2 | Vitest component tests | ~114 tests | <10s | `uvx inv interface.test` |
| 3 | Playwright E2E | 12 spec files | 30s–2min | `uvx inv interface.e2e` |
| 4 | Integration tests | DB + NATS + LLM | varies | manual |
| 5 | Governance smoke | cmd/smoke | varies | `uvx inv core.smoke` |

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
- **E2E specs:** `e2e/specs/*.spec.ts` (12 files)
- **Config:** `playwright.config.ts`

### Smoke Test (Quality Gate)

`uvx inv interface.check` validates 9 pages:
- HTTP 200 status
- No SSR errors
- No hydration mismatches
- No `bg-white` / `bg-zinc-*` / `bg-slate-*` CSS leaks

---

## VI. CI/CD

### GitHub Actions Workflows

| Workflow | File | Trigger Paths | Checks |
|----------|------|--------------|--------|
| **Core CI** | `core-ci.yaml` | `core/**`, `ops/core.py` | Go test + coverage, GolangCI-Lint v1.64.5, binary build |
| **Interface CI** | `interface-ci.yaml` | `interface/**`, `ops/interface.py` | ESLint, `tsc --noEmit`, Vitest, production build |
| **E2E CI** | `e2e-ci.yaml` | `interface/**`, `core/**` | Build Core + Next.js, start servers, Playwright (Chromium) |

**Trigger:** Push/PR to `main` and `develop`

### Local CI

```bash
uvx inv ci.check    # Full pipeline: lint → test → build (with stage timers)
uvx inv ci.deploy   # Build + Docker + K8s deploy
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

1. Branch off `main`
2. Commit often (conventional commits)
3. Merge via squash

---

## VII. Deployment Architecture

### Kubernetes (Kind)

**Helm Chart:** `charts/mycelis-core/`

```yaml
replicaCount: 1
image: mycelis/core:latest (immutable TAG from core.build)
service: ClusterIP:8080
resources:
  requests: { cpu: 50m, memory: 64Mi }
  limits: { cpu: 200m, memory: 256Mi }
securityContext:
  runAsUser: 1000
  runAsNonRoot: true
  seccompProfile: RuntimeDefault
persistence: 2Gi (data/artifacts)
```

**Dependencies (Bitnami/NATS Helm charts):**
- PostgreSQL 12.12.10 (image: pgvector/pgvector:16-alpine, 1Gi persistence)
- NATS 1.2.4 (JetStream: 128Mi memory, 2Gi file)

**Helm Templates:** deployment.yaml, service.yaml, secrets.yaml, network-policy.yaml, data-pvc.yaml

### Docker

**Multi-stage build:**
- Stage 1: golang:1.26-alpine → compile
- Stage 2: alpine → copy binary
- Exposes port 8080
- ENTRYPOINT: `/app/server`

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
9.  Spawn standing teams from YAML (admin, council, genesis)
10. Start HTTP server (port 8080), register routes
11. Block on SIGINT
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
| Kind cluster cert chain breaks after Docker restart | Fix with `uvx inv k8s.reset` |
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
