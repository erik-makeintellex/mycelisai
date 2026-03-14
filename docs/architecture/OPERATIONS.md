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
**Management scripting mandate:** App-tied management logic belongs in Python task modules. PowerShell is only a thin local wrapper when the host platform requires it.

---

## II. Task Automation

### Master Registry

**File:** `tasks.py` — 12 collections, 60+ @task functions

**Run from:** `scratch/` (project root where tasks.py lives)

**Invocation:** Always use `uv run inv <namespace>.<task>` (or `.\.venv\Scripts\inv.exe` from the project root) for real task execution.
Use `uvx --from invoke inv -l` only as a lightweight compatibility probe.
Do not use bare `uvx inv ...`.
Lifecycle tasks must not report success until Core `/healthz` is actually ready; open-port-only checks are insufficient.
Background Core startup must write to `workspace/logs/core-startup.log` so lifecycle failures have a deterministic diagnostic surface.
Lifecycle teardown must use bounded cleanup subprocesses and wait for ports to close before reporting success.

> **Tip:** If you `uv venv && .venv/Scripts/activate` (Windows) or `source .venv/bin/activate` (Linux), you can use `inv` directly.

### Core Tasks (`ops/core.py`)

| Command | Description |
|---------|-------------|
| `uv run inv core.build` | Compile Go binary + Docker image (returns TAG) |
| `uv run inv core.test` | `go test ./...` |
| `uv run inv core.run` | Run Core locally (foreground, blocking) |
| `uv run inv core.stop` | Kill running Core process (taskkill on Windows, pkill on Linux) |
| `uv run inv core.restart` | stop + run |
| `uv run inv core.smoke` | `go run ./cmd/smoke/main.go` (governance smoke tests) |
| `uv run inv core.clean` | `go clean`, remove bin/ |

### Interface Tasks (`ops/interface.py`)

| Command | Description |
|---------|-------------|
| `uv run inv interface.dev` | Start Turbopack dev server (stops existing first) |
| `uv run inv interface.install` | `npm install` |
| `uv run inv interface.build` | `npm run build` (production) |
| `uv run inv interface.lint` | `npm run lint` (ESLint) |
| `uv run inv interface.test` | `npm run test` (Vitest) |
| `uv run inv interface.test-coverage` | Vitest with V8 coverage |
| `uv run inv interface.e2e` | `npm run e2e` (Playwright owns Next.js server lifecycle; Invoke clears stale Interface listeners before/after; optional `--headed`, `--project=...`, `--spec=...`, `--live-backend`) |
| `uv run inv interface.stop` | Kill process on port 3000 |
| `uv run inv interface.clean` | rm -rf .next cache |
| `uv run inv interface.restart` | stop → clean → build → dev → check |
| `uv run inv interface.check` | Smoke-test: 9 pages for 200 status, no SSR errors, no hydration issues, no `bg-white`/`bg-zinc`/`bg-slate` leaks |

### Database Tasks (`ops/db.py`)

| Command | Description |
|---------|-------------|
| `uv run inv db.migrate` | Apply canonical forward SQL migrations (`001_init_memory.sql` + `*.up.sql`) |
| `uv run inv db.reset` | Drop + recreate + migrate |
| `uv run inv db.status` | List tables + row counts |
| `uv run inv db.create` | Create cortex database (if not exists) |

### Auth Tasks (`ops/auth.py`)

| Command | Description |
|---------|-------------|
| `uv run inv auth.dev-key` | Ensure a local `MYCELIS_API_KEY` exists and keep `.env.example` on a sample value |

### Lifecycle Tasks (`ops/lifecycle.py`)

| Command | Description |
|---------|-------------|
| `uv run inv lifecycle.up` | Idempotent bring-up: bridge -> dependencies -> Core (waits for `/healthz`; supports `--build` and `--frontend`) |
| `uv run inv lifecycle.down` | Clean teardown: Core -> Frontend -> port-forwards with bounded cleanup and port-close waits |
| `uv run inv lifecycle.status` | Dashboard for PostgreSQL, NATS, Core, Frontend, Ollama, and related PIDs |
| `uv run inv lifecycle.health` | Deep health probe against actual API endpoints with auth |
| `uv run inv lifecycle.restart` | Full local restart: down -> settle -> up (supports `--build` and `--frontend`) |
| `uv run inv lifecycle.memory-restart` | Destructive recovery path: down -> db.reset -> up -> health -> memory probes |

### Logging & Quality Gates (`ops/logging.py`, `ops/quality.py`)

| Command | Description |
|---------|-------------|
| `uv run inv logging.check-schema` | Verify runtime event literals map to declared event constants and docs coverage |
| `uv run inv logging.check-topics` | Fail when production code hardcodes `swarm.*` subjects outside allowed constants files |
| `uv run inv quality.max-lines --limit 350` | Enforce the hot-path `<=350 LOC` policy with temporary no-regression legacy caps |

### Kubernetes Tasks (`ops/k8s.py`)

| Command | Description |
|---------|-------------|
| `uv run inv k8s.init` | Create Kind cluster (handles Windows absolute paths) |
| `uv run inv k8s.up` | Canonical cluster bring-up: init -> deploy -> wait (PostgreSQL -> NATS -> Core API) |
| `uv run inv k8s.deploy` | Build Core, load Docker image, helm upgrade (injects secrets from .env) |
| `uv run inv k8s.wait` | Wait for rollout readiness gates (PostgreSQL -> NATS -> Core API) |
| `uv run inv k8s.bridge` | Port-forward NATS:4222, HTTP:8081←8080, PG:5432 |
| `uv run inv k8s.status` | Check Docker, Kind cluster, pod status, PVC status |
| `uv run inv k8s.recover` | Restart core + infra resources (core, NATS, PostgreSQL) |
| `uv run inv k8s.reset` | Full reset: delete cluster -> canonical bring-up (includes readiness wait) |

### Cognitive Tasks (`ops/cognitive.py`)

| Command | Description |
|---------|-------------|
| `uv run inv cognitive.install` | `uv sync` dependencies |
| `uv run inv cognitive.llm` | Start vLLM text server (GPU, configurable) |
| `uv run inv cognitive.media` | Start Diffusers media server (OpenAI-compatible) |
| `uv run inv cognitive.up` | Start full stack (vLLM + Media, handles Ctrl+C) |
| `uv run inv cognitive.stop` | Kill all cognitive processes |
| `uv run inv cognitive.status` | HTTP probes to check vLLM + Media health |

### Test Tasks (`ops/test.py`)

| Command | Description |
|---------|-------------|
| `uv run inv test.all` | Run ALL unit tests (Core + Interface) |
| `uv run inv test.coverage` | Go coverage + Vitest V8 coverage |
| `uv run inv test.e2e` | Alias for interface.e2e |

### CI Tasks (`ops/ci.py`)

| Command | Description |
|---------|-------------|
| `uv run inv ci.lint` | Go vet + Next.js lint (both must pass) |
| `uv run inv ci.test` | Go tests + Interface tests |
| `uv run inv ci.build` | Go binary + Next.js production build (no Docker) |
| `uv run inv ci.check` | Full pipeline: lint → test → build (with stage timers) |
| `uv run inv ci.baseline` | Strict delivery baseline covering gates, focused runtime tests, interface build/typecheck, and Vitest |
| `uv run inv ci.entrypoint-check` | Verify the supported invoke runner matrix and reject unsupported bare aliases |
| `uv run inv ci.toolchain-check` | Report toolchain versions and optionally enforce Go lock policy |
| `uv run inv ci.release-preflight` | Enforce release gate: clean tree + runner/toolchain checks + strict baseline |
| `uv run inv ci.deploy` | Build + Docker + K8s deploy |

### Other Tasks

| Module | Commands | Purpose |
|--------|----------|---------|
| `ops/proto_relay.py` | `proto.generate` | Protobuf codegen (Go + Python) |
| `ops/proto_relay.py` | `relay.test`, `relay.demo` | Python SDK tests + demo |
| `ops/device.py` | `device.boot --id=ghost-01` | Device simulation (NATS announce) |
| `ops/misc.py` | `clean.legacy` | Remove legacy Makefile/docker-compose |
| `ops/misc.py` | `team.sensors`, `team.output`, `team.test`, `team.architecture-sync`, `team.worktree-triage` | Python agent teams, central architect sync, and temporary dirty-worktree review/test triage |

Key coordination example: `uv run inv team.architecture-sync`

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
# Prerequisite (run once per reboot/reset):
uv run inv k8s.up       # Kind/namespace -> Helm deploy -> PostgreSQL -> NATS -> Core API

# Terminal 1: Infrastructure bridge
uv run inv k8s.bridge    # Port-forward NATS, API, Postgres

# Terminal 2: Backend
uv run inv core.run      # Go backend (foreground, port 8080 → bridged to 8081)

# Terminal 3: Frontend
uv run inv interface.dev # Next.js dev (Turbopack, port 3000)
```

Open [http://localhost:3000](http://localhost:3000)

### First-Time Setup

```bash
# 1. Configure secrets
cp .env.example .env
# Edit .env: POSTGRES_USER, POSTGRES_PASSWORD, POSTGRES_DB

# 2. Boot infrastructure in dependency order
uv run inv k8s.up        # Kind + Helm + readiness gates (PostgreSQL -> NATS -> Core API)

# 3. Open bridge (keep running)
uv run inv k8s.bridge

# 4. Initialize database
uv run inv db.migrate    # Apply canonical forward migrations (fresh reset verified on 2026-03-06)

# 5. Start backend
uv run inv core.run

# 6. Start frontend
uv run inv interface.install  # First time only
uv run inv interface.dev
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
| `MYCELIS_WORKSPACE` | `./workspace` (local) / `/data/workspace` (K8s) | Workspace sandbox root for manifested files and filesystem tools |
| `DATA_DIR` | `./workspace/artifacts` (local) / `/data/artifacts` (K8s) | Artifact storage root for file-backed outputs |
| `OLLAMA_HOST` | — | Override Ollama endpoint |
| `OPENAI_API_KEY` | — | OpenAI API key |
| `ANTHROPIC_API_KEY` | — | Anthropic API key |
| `GEMINI_API_KEY` | — | Gemini API key |

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
| **Core CI** | `core-ci.yaml` | `core/**`, `ops/core.py` | Go test + coverage, GolangCI-Lint v1.64.5, binary build |
| **Interface CI** | `interface-ci.yaml` | `interface/**`, `ops/interface.py` | ESLint, `tsc --noEmit`, Vitest, production build |
| **E2E CI** | `e2e-ci.yaml` | `interface/**`, `core/**` | Build Core + Next.js, start Core, Playwright-managed UI server, Chromium/Firefox/WebKit + mobile smoke |

**Trigger:** Push/PR to `main` and `develop`

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
9.  Spawn standing teams from YAML (admin, council, genesis)
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


