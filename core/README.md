# Mycelis Core (Tier 1)
> **Role**: The Cognitive & State Engine
> **Language**: Go
> **Path**: `core/`

## 🧠 Responsibilities
1.  **Routing**: Delivers `MsgEnvelope` between Agents (NATS).
2.  **Governance**: Enforces `policy.yaml` (The Guard).
3.  **Memory**: Persists logs to PostgreSQL (Memory Service).
4.  **Intelligence**: Connects to LLMs for cognitive loops (Cognitive Engine).

## 🏗️ Structure
- `cmd/server/`: Main entrypoint.
- `internal/router/`: NATS subscription logic.
- `internal/governance/`: Policy enforcement engine.
- `internal/state/`: In-memory Registry.

## 🚀 Development
```bash
# Run Unit Tests
inv core.test

# Build Binary (local)
inv core.build
```
