# Mycelis Core (Tier 1)
> Navigation: [Project README](../README.md) | [Docs Home](../docs/README.md)
> **Role**: The Cognitive & State Engine
> **Language**: Go
> **Path**: `core/`

## TOC

- [Responsibilities](#-responsibilities)
- [Structure](#-structure)
- [Development](#-development)

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
uv run inv core.test

# Build Binary (local)
uv run inv core.compile
```
