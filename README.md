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
  - **Cognitive Matrix:** Control panel for routing prompts to Local (Ollama) or Cloud (OpenAI) models.

## 🚀 Getting Started

### 1. Boot the Infrastructure
```bash
# Start NATS, Postgres, and Core in Kubernetes (KinD)
uv run inv k8s.up
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
