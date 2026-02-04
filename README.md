# Mycelis Cortex V6

**A Multi-Tenant, Cognitive Swarm Orchestrator.**

Mycelis is a "Neural Organism" that orchestrates AI agents to solve complex tasks. V6 introduces the **Cortex Command Console**, a high-density, rigorous operational environment for managing cognitive resources.

## 🏗️ Architecture

- **Tier 1: Core (Go + Postgres)**
  - Use `core` service for Identity, Governance, and Cognitive Routing.
  - **CQA (Cognitive Quality Assurance):** Enforces strict timeouts and schema validation on all LLM calls.
  - **SCIP (Standardized Cybernetic Interchange Protocol):** Protobuf-based message envelope for all agent communication.

- **Tier 2: Nervous System (NATS JetStream)**
  - Real-time event bus (`scip.>`) connecting the Core to the Swarm.

- **Tier 3: The Cortex (Next.js)**
  - **Aero-Light Theme:** High-contrast, strictly typed command console.
  - **Cognitive Matrix:** Control panel for routing prompts.
  - **Telemetry Dashboard:** Real-time observability (Logs, Agents) at `/dashboard`.
  - **Registry & Marketplace:** Catalog for installing Connectors (`/marketplace`) and Blueprints.
  - **Neural Wiring:** Visual signal graph (`/wiring`) of input/output data flows.

## 📚 Documentation Hub
| Context | Resource | Description |
| :--- | :--- | :--- |
| **Architecture** | [The Archivist](docs/architecture/DIRECTIVE_ARCHIVIST.md) | Event Store & Memory Architecture Specs. |
| **Governance** | [Gatekeeper Protocol](docs/governance.md) | Policy enforcement, approvals, and security boundaries. |
| **Registry** | [The Registry](core/internal/registry/README.md) | Connector Marketplace and Wiring Graph specs. |
| **Telemetry** | [Logging Schema](docs/logging.md) | SCIP Log structure and centralized observability. |
| **Testing** | [Verification Suite](docs/TESTING.md) | Unit, Integration, and Smoke Testing protocols. |
| **AI Providers** | [Provider Guide](docs/PROVIDERS.md) | LLM configuration (Ollama, OpenAI, Anthropic). |
| **Core API** | [Core Specs](core/README.md) | Go Service architecture and internal packages. |
| **CLI** | [Synaptic Injector](cli/README.md) | `myc` command-line tool usage. |
| **Interface** | [Cortex UI](interface/README.md) | Next.js Frontend architecture. |

### 🔑 Key Configurations
- **Brain (Models):** `core/config/brain.yaml`
- **Policy (Rules):** `core/config/policy.yaml`
- **Secrets:** `.env` (See `.env.example`)


## 🧠 Cognitive Architecture (Default)

Mycelis V6 defaults to a **Single Local Model** architecture for privacy and air-gapped readiness.
- **Default Model:** `qwen2.5-coder:7b-instruct` (via Ollama).
- **Fallback:** None (Strict Reliability).

### 🛠️ Developer Orchestration
Run the following commands from the root directory:

| Command | Description |
| :--- | :--- |
| `inv core.build` | Compiles the Go binary and builds the Docker image. |
| `inv core.test` | Runs Unit Tests. |
| `inv core.run` | **(New)** Runs the Core locally (Native). |
| `inv core.restart` | **(New)** Restarts the Core (Kill + Run). |
| `inv core.smoke` | **(New)** Runs Governance Smoke Tests against local Core. |
| `inv k8s.reset` | **(New)** Full Infrastructure Reset (Teardown + Init + Deploy). |
| `inv k8s.status` | Checks the status of the Kubernetes cluster. |
| `inv k8s.deploy` | Deploys the Helm chart to the local cluster. |
| `inv k8s.bridge` | Opens ports for NATS, API, and Postgres. |

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
uv run inv k8s.reset

# Open Development Bridge (Required for CLI/UI)
uv run inv k8s.bridge
```

### 2. Bootstrap Identity (The First Login)
Since V6 enforces RBAC, you must create an Admin user to access the console.
*(Ensure `inv k8s.bridge` is running if using local CLI)*
```bash
# Create the first admin user
uv run python cli/main.py admin-create "admin"
```
*Copy the Session Token output!*

### 3. Launch the Cortex Console
```bash
cd interface
npm run dev
```
Open [http://localhost:3000](http://localhost:3000).

### 4. Configure the Brain
Edit `core/config/brain.yaml` to define your Model Matrix.
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
# Run Unit Tests (Logic)
cd core
go test ./internal/cognitive/...

# Run Integration Tests (Real Ollama)
go test -v -tags=integration ./tests/...
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

### Protocol:
1.  **Branch off `main`:** Always start fresh.
2.  **Commit Often:** Use conventional commits (e.g., `feat: add gatekeeper`).
3.  **Merge via Squash:** Keep history linear and clean.
