# Mycelis Service Network (Gen-3)

> **Current Architecture**: "Absolute Architecture" (Phase 3)
> **Agent Context**: This file is the Source of Truth for Project Structure and Tooling.

## 📚 Table of Contents

- [Architecture](#-architecture-the-absolute-standard)
- [Tooling & Standards](#-tooling--standards)
- [Quick Start](#quick-start-invoke-workflow)
- [The Relay (SDK)](#-the-relay-python-sdk)
- [Cortex Memory](#-cortex-memory-logging)
- [Governance & Policy](#-governance--policy)
- [Directory Structure](#-directory-structure)

## 📌 Quick Links
- [System Architecture (Detailed)](architecture.md)
- [Product Vision (PRD v3)](myclis_v3_arch.md)
- [Governance Documentation](docs/governance.md)
- [Logging Schema](docs/logging.md)
- [UI Design Spec](ui_design_spec.md)

---

## 🏗️ Architecture: The Absolute Standard

We have pivoted to a strict segregation of concerns:

| Component | Tech | Path | Description |
| :--- | :--- | :--- | :--- |
| **Neural Core** | Go | `core/` | The central brain. Manages state, routing, and swarm coherence. |
| **Relay SDK** | Python | `sdk/python` | The universal connector for Agents. Stateless & Team-Aware. |
| **Contracts** | Protobuf | `proto/` | The LAW. `swarm.proto` defines all inter-service communication. |
| **Memory** | NATS | `cortex.logs` | Centralized, strict-schema logging (`LogEntry`). |

---

## 🛠️ Tooling & Standards

We enforce strict tooling to ensure deterministic environments.

-   **Task Runner**: **Invoke** (`tasks.py`). *Makefiles are banned.*
-   **Dependency Manager**: **`uv`** (Python) and **`go mod`** (Go).
-   **Cluster**: Kind (Kubernetes in Docker).

### Quick Start (Invoke Workflow)
You only need `uv` installed. The runner handles the rest.

```bash
# 0. Install Invoke
uv tool install invoke

# 1. Start Infrastructure (Kind + NATS + Postgres)
inv k8s.init

# 2. Generate Contracts (Protobuf -> Go/Python)
inv proto.generate

# 3. Build & Deploy Core (Helm Chart -> Kind)
inv k8s.deploy

# 4. Check Status (NEW)
inv k8s.status

# 5. Open Development Bridge (NATS:4222, HTTP:8080)
inv k8s.bridge

# 6. Verify Stack
inv core.test
inv relay.test

# 7. Recovery (Restart Deployments)
inv k8s.recover
```

---

## 🔌 The Relay (Python SDK)

The **RelayClient** (`sdk/python/src/relay/client.py`) is the only authorized way for Python code to interact with the Swarm.

### Connecting a Service
```python
from relay.client import RelayClient
from swarm.v1 import swarm_pb2
import asyncio

async def main():
    # 1. Connect to Swarm
    relay = RelayClient(
        agent_id="my-agent-01", 
        team_id="data-proc"
    )
    await relay.connect()

    # 2. Subscribe using Strict Types
    async def on_task(envelope: swarm_pb2.MsgEnvelope):
        print(f"Task Received: {envelope.event.event_type}")
        
        # 3. Emit Result
        await relay.send_event(
            event_type="task.complete",
            data={"status": "done"},
            context={"processor": "gpu-01"}
        )

    await relay.subscribe("swarm.team.data-proc.>", on_task)

    # Keep alive
    while True: await asyncio.sleep(1)

if __name__ == "__main__":
    asyncio.run(main())
```

---

## 🧠 Cortex Memory (Logging)

No more text files. We use a **Structured Log Event Stream**.

-   **Schema**: Defined in `proto/swarm/v1/swarm.proto` as `LogEntry`.
-   **Traceability**: `trace_id` and `span_id` are mandatory.
-   **Context**: Logs capture the `swarm_context` (State Snapshot) at the moment of emission.
-   **Transport**: All logs flow to `cortex.logs` on NATS.

> 📘 **Documentation**: See [docs/logging.md](docs/logging.md) for the detailed `LogEntry` schema and parsing guide.

---

## 🛡️ Governance & Policy

Messages are filtered by the **Gatekeeper** before routing.
Rules are defined in `core/policy/policy.yaml`.

> 📘 **Documentation**: See [docs/governance.md](docs/governance.md) for configuration examples (IoT vs. Agents).

---

## 📂 Directory Structure

```text
/
├── charts/                # [Helm] Mycelis Core Chart
├── core/                  # [Go] The Neural Core
│   ├── cmd/server/        # Entrypoint
│   └── internal/state/    # In-Memory Agent Registry
├── ops/                   # [Python] Build System Tasks
├── proto/                 # [Protobuf] The Law
│   └── swarm/v1/          # swarm.proto (Contracts)
├── sdk/                   # [Python] The Relay
│   └── python/            # Source-Aware Client
└── tasks.py               # [Invoke] Entrypoint (wraps ops/)
```

> **Note**: `api/`, `runner/`, and `ui/` are LEGACY and deprecated. Do not use them for new development.

---

## 🤖 Agent Directives

**If you are an Agent working on this repo:**
1.  **Strict Mode**: If it's not in `tasks.py`, it doesn't exist.
2.  **No Makefiles**: Do not suggest or create Makefiles.
3.  **Python == Relay**: All Python code must use `RelayClient`.

**Active Specialists:**
*   `spec:arch:01` - System Architect
*   `spec:golang:01` - Backend Engineer
