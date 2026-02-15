# Mycelis Cortex V6.2 (The Lattice)

**The Recursive Swarm Operating System.**

> [!IMPORTANT]
> **MASTER STATE AUTHORITY**
> The Single Source of Truth for the project state, architectural delta, and immediate goals is:
> **[mycelis-6-2.md](mycelis-6-2.md)**
>
> **Agents & Humans:** Always consult the Master State document before making architectural decisions. This README provides general usage instructions, but the Codex defines "What is True".

Mycelis is a "Neural Organism" that orchestrates AI agents to solve complex tasks. V6.2 introduces the **Fractal Fabric**, **Cognitive Registry**, and **Context Economy**.

## 🏗️ Architecture

- **Tier 1: Core (Go + Postgres + pgvector) [VERIFIED]**
  - Use `core` service for Identity, Governance, and Cognitive Routing.
  - **Cognitive Registry:** Database-backed (`llm_providers`) lookup for AI models, decoupling logic from config.
  - **Guard:** The Governance Engine (formerly Gatekeeper) enforcing policy on every message.
  - **Bootstrap Service:** Listens for Hardware Announcements and Heartbeats (`005_nodes`).
  - **Archivist:** The Context Engine. Summarizes logs into **SitReps** (Situation Reports) for efficient Agentry.

- **Tier 2: Nervous System (NATS JetStream)**
  - **Heartbeat:** Global 1Hz pulse (`swarm.global.heartbeat`) from all nodes.
  - **Audit Trace:** Enforced logging (`swarm.audit.trace`) of all traffic.
  - **SCIP:** Standardized Cybernetic Interchange Protocol (Protobuf).

- **Tier 2.5: The Fractal Fabric (Data)**
  - **Teams & Missions:** Recursive hierarchy (`007_team_fabric`) defining ownership and purpose.
  - **Registry:** Centralized database for Connectors and Blueprints.
  - **Iron Dome:** Security layer enforcing NSA/CIS Hardening standards (User 10001, ReadOnly FS).

- **Tier 3: The Face (Next.js + Zustand)**
  - **Mission Control (`/`):** Panopticon dashboard — TelemetryRow (4 live compute sparkline cards polling `/api/v1/telemetry/compute`), PriorityStream (filtered governance/error/artifact NATS feed), MissionsPanel (active missions), ActivityStream (full SSE feed). Header routes: NEW MISSION → `/wiring`, Settings → `/settings`.
  - **The Shell Layout (`/wiring`):** A rigid frame for fluid intelligence.
    - **Zone A (Vitals):** System health, heartbeat, and resource metrics.
    - **Zone B (The Circuit):** Workspace split-pane — ArchitectChat (intent negotiation + TrustSlider + BlueprintDrawer) + CircuitBoard (ReactFlow DAG) or SquadRoom (fractal drill-down).
    - **Spectrum (Bottom Panel):** NatsWaterfall — collapsible real-time SSE stream visualization powered by Zustand `streamLogs[]`.
    - **Zone D (The Valve):** Human-in-the-loop Governance overlay — `GovernanceModal.tsx` with two-column review (output + proof-of-work), APPROVE/REJECT controls. Trust Economy: envelopes with `TrustScore < AutoExecuteThreshold` are halted by the Overseer Governance Valve and routed to Zone D via SSE `governance_halt` signals.
    - **Deliverables Tray:** Bottom-docked `DeliverablesTray.tsx` showing pending `CTSEnvelope` artifacts intercepted from SSE. Pulsing green glow signals human action needed.
  - **Trust Economy:** Autonomy Threshold slider (0.0–1.0) in ArchitectChat controls the `AutoExecuteThreshold`. High-trust envelopes bypass human approval; low-trust halts for governance. Synced to backend via `PUT /api/v1/trust/threshold`.
  - **Blueprint Library:** Slide-out drawer in ArchitectChat for saving, importing (JSON), exporting, and loading mission topologies into the ReactFlow canvas.
  - **Node Iconography:** AgentNode border-left accent by category — Cognitive (purple), Sensory (cyan), Actuation (green), Ledger (muted). Trust score badge visible per-node.
  - **Fractal Navigation:** Double-click team group nodes to drill into SquadRoom sub-views (internal debate feed + proof-of-work artifacts).
  - **State Fabric:** Zustand 5.0.11 atomic store (`useCortexStore`) — strict unidirectional data flow, no useState for API/graph state. Trust state: `trustThreshold`, `setTrustThreshold()`. Blueprint state: `savedBlueprints[]`, `loadBlueprint()`, `toggleBlueprintDrawer()`.
  - **Live Telemetry:** SSE stream (`/api/v1/stream`) dispatches signals to both the waterfall and individual ReactFlow agent nodes (activity ring + thought bubble). Compute telemetry: `GET /api/v1/telemetry/compute` (goroutines, heap, system memory, LLM tokens/sec).
  - **Visual Protocol:** Vuexy Dark — `cortex-bg` (#25293C), `cortex-surface` (#2F3349), `cortex-primary` (#7367F0), `cortex-success` (#28C76F). Zero zinc/slate classes in active routes.

## 📚 Documentation Hub

| Context | Resource | Description |
| :--- | :--- | :--- |
| **Architecture** | [Memory Specs](docs/architecture/DIRECTIVE_MEMORY_SERVICE.md) | Event Store & Memory Architecture Specs. |
| **Governance** | [Guard Protocol](docs/governance.md) | Policy enforcement, approvals, and security boundaries. |
| **Registry** | [The Registry](core/internal/registry/README.md) | Connector Marketplace and Wiring Graph specs. |
| **Telemetry** | [Logging Schema](docs/logging.md) | SCIP Log structure and centralized observability. |
| **Testing** | [Verification Suite](docs/TESTING.md) | Unit, Integration, and Smoke Testing protocols. |
| **AI Providers** | [Provider Guide](docs/PROVIDERS.md) | LLM configuration (Ollama, OpenAI, Anthropic). |
| **Core API** | [Core Specs](core/README.md) | Go Service architecture and internal packages. |
| **CLI** | [Synaptic Injector](cli/README.md) | `myc` command-line tool usage. |
| **Interface** | [Cortex UI](interface/README.md) | Next.js Frontend architecture. |

### 🔑 Key Configurations

- **Cognitive (Bootstrap):** `core/config/cognitive.yaml` (Overrides DB if needed)
- **Cognitive (Dynamic):** `llm_providers` table (Managed via UI/SQL)
- **Policy (Rules):** `core/config/policy.yaml`
- **Secrets:** `.env` (See `.env.example`)

## 🧠 Cognitive Architecture (Default)

Mycelis V6 defaults to a **Single Local Model** architecture for privacy and air-gapped readiness, but supports granular overrides per Agent.

- **Default Model:** `qwen2.5-coder:7b-instruct` (via Ollama).
- **Fallback:** None (Strict Reliability).
- **Agent Overrides:**
  - **System Prompt:** Custom instructions defining the agent's persona.
  - **Model Profile:** Specific model ID (e.g., `llama3.2`, `gpt-4o`) for specialized tasks.

### 🛠️ Developer Orchestration

**Prerequisites:** [uv](https://github.com/astral-sh/uv) (for Python/Node management) and [Docker](https://www.docker.com/).

Run the following commands from the root directory:

| Command | Description |
| :--- | :--- |
| **Core** | |
| `inv core.build` | Compiles the Go binary and builds the Docker image. |
| `inv core.test` | Runs Unit Tests. |
| `inv core.run` | Runs the Core locally (Native). |
| `inv core.restart` | Restarts the Core (Kill + Run). |
| `inv core.smoke` | Runs Governance Smoke Tests against local Core. |
| **Interface** | |
| `inv interface.dev` | Start Next.js dev server (Turbopack). |
| `inv interface.build` | Production build. |
| `inv interface.lint` | ESLint check. |
| `inv interface.test` | Run Vitest unit tests. |
| `inv interface.check` | Smoke-test running server (fetches pages, checks for errors). |
| `inv interface.stop` | Kill dev server by port. |
| `inv interface.clean` | Clear `.next` build cache. |
| `inv interface.restart` | Full restart: stop → clean → build → dev → check. |
| **Database** | |
| `inv db.migrate` | Apply all SQL migrations to cortex (idempotent). |
| `inv db.reset` | Drop + recreate cortex DB, then run all migrations. |
| `inv db.create` | Create the cortex database if it doesn't exist. |
| `inv db.status` | Show tables in the cortex database. |
| **Infrastructure** | |
| `inv k8s.reset` | Full Infrastructure Reset (Teardown + Init + Deploy). |
| `inv k8s.status` | Checks the status of the Kubernetes cluster. |
| `inv k8s.deploy` | Deploys the Helm chart to the local cluster. |
| `inv k8s.bridge` | Opens ports for NATS, API, and Postgres. |
| `inv device.boot` | Simulates a hardware node announcement via NATS. |
| **Testing** | |
| `inv test.all` | Run all tests (Core + Interface). |
| `inv team.test` | Run Python agent unit tests. |

> [!TIP]
> **Pro Tip:** If you run `uv venv` and activate your virtual environment, you can run `inv` directly without `uv run`.

> [!NOTE]
> **Migrations:** Use `uvx inv db.migrate` (idempotent) or `uvx inv db.reset` (destructive). Migrations live in `core/migrations/*.sql` and are applied in lexicographic order via `psql` over the bridge.

### Hardware Grading

| Tier | RAM | Supported Models | Use Case |
| :--- | :--- | :--- | :--- |
| **Tier 1 (Min)** | 16 GB | 7B Models (Q4) | Basic Coding, CLI |
| **Tier 2 (Rec)** | 32 GB | 14B - 32B Models | Complex Architecture, Deep Reasoning |
| **Tier 3 (Ultra)** | 64 GB+ | 70B+ or Multi-Model | **Enterprise Core** (Current Dev Host) |

*Note: The system auto-detects resources but defaults to the 7B model for speed.*

## 🚀 Getting Started

### 1. Configure Secrets

Ensure your `.env` file exists and contains the necessary credentials:

```bash
cp .env.example .env
# Edit .env to set POSTGRES_USER, POSTGRES_PASSWORD, etc.
```

### 2. Boot the Infrastructure

```bash
# Full System Reset (Cluster + Core + DB)
uvx inv k8s.reset
```

### 3. Open the Development Bridge (Terminal 1)

Port-forwards PostgreSQL, NATS, and HTTP from the Kind cluster to localhost. Keep this running.

```bash
uvx inv k8s.bridge
```

### 4. Initialize the Database

```bash
uvx inv db.migrate    # Apply all migrations (idempotent)
# or: uvx inv db.reset   # Full drop + recreate + migrate
```

### 5. Start the Core Server (Terminal 2)

```bash
uvx inv core.run      # Stops any existing instance, then starts
```

### 6. Launch the Cortex Console (Terminal 3)

```bash
uvx inv interface.install   # First time: install npm dependencies
uvx inv interface.dev       # Start dev server on port 3000
```

Open [http://localhost:3000](http://localhost:3000).

### 4. Configure the Cognitive Engine

Edit `core/config/cognitive.yaml` to define your Model Matrix.

- **Profiles:** `sentry`, `architect`, `coder`.
- **Policies:** Set `timeout_ms` and `max_retries` per profile.

## 🛠️ Developer Tools

### CLI (`myc`)

- `myc snoop`: Decode SCIP traffic in real-time.
- `myc inject <intent> <payload>`: Send signals to the swarm.
- `myc think <prompt> --profile=coder`: Test the cognitive router.

### Protobuf Generation

If you modify `proto/envelope.proto`:

```bash
uv run inv proto.generate
```

## 🧪 Verification

Mycelis uses a 2-Tier testing strategy (Mocked Unit + Real Integration).
See [docs/TESTING.md](docs/TESTING.md) for full details.

```bash
# Run All Core Unit Tests (Logic)
cd core
go test ./...

# Run Specific Package Tests
go test -v ./internal/bootstrap/... # Heartbeats
go test -v ./internal/memory/...    # Archivist/SitReps
go test -v ./internal/cognitive/... # LLM Router

# Run Integration Tests (Real Ollama)
# Run Integration Tests (Real Ollama)
# Requires OLLAMA_HOST to be set
$env:OLLAMA_HOST="http://localhost:11434"; go test -v -tags=integration ./tests/...
```

## 🌲 Branching Strategy

Mycelis follows a strict **Trunk-Based Development** workflow with ephemeral feature branches.

| Branch Type | Prefix | Example | Description |
| :--- | :--- | :--- | :--- |
| **Production** | `main` | `main` | Stable, deployable code. All PRs must pass CI. |
| **Feature** | `feat/` | `feat/neural-router` | New capabilities or substantial changes. |
| **Fix** | `fix/` | `fix/memory-leak` | patches for bugs or regression. |
| **Chore** | `chore/` | `chore/infra-reset` | Maintenance, dependencies, or refactoring. |
| **Documentation** | `docs/` | `docs/api-spec` | Documentation-only updates. |

### Protocol

1. **Branch off `main`:** Always start fresh.
2. **Commit Often:** Use conventional commits (e.g., `feat: add gatekeeper`).
3. **Merge via Squash:** Keep history linear and clean.
