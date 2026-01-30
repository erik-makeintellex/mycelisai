# Product Requirement Document: Project Mycelis
**Version:** 3.0 (Hybridization Phase)
**Date:** January 29, 2026
**Status:** Active Construction
**Architect:** The Absolute Architecture (Tier 0)

---

## 1. Vision & Core Philosophy
**The Goal:** To construct a "Symbiotic Swarm Intelligence"—a single, cohesive cybernetic organism rather than a distributed microservices architecture.

**The Metaphor: Digital Biology**
The system is modeled after a biological nervous system, prioritizing speed, strict typing, and centralized memory.
* **The Brain (Neural Core):** A high-performance, centralized decision engine (Go). It resides in the "Skull" (Kubernetes) and handles state, routing, and memory. It does not perform tasks; it commands them.
* **The Nerves (The Protocol):** A strictly typed, high-speed communication bus (NATS + Protobuf) carrying "Intent" and "Context."
* **The Limbs (The Relay):** Lightweight, stateless workers (Python). Whether it is a "Marketing Writer" (LLM) or a "Sensor Array" (Raspberry Pi), they connect, perform a specific function, and report back. They are the "Hands."
* **The Cortex (Memory):** A centralized stream of consciousness (`LogEntry`) enabling the Brain to review and learn from past actions.

---

## 2. The Absolute Architecture (Technical Stack)

### A. The Neural Core (Server)
* **Language:** Go (Golang) 1.23+
* **Location:** Kubernetes Cluster (`mycelis-cluster`), Namespace `mycelis`.
* **Responsibility:**
    * **Registry:** Tracks all active agents, their `team_id`, and `source_uri`.
    * **Orchestration:** Routes messages based on `intent` and `swarm_context` (AG2 Pattern).
    * **Observability:** Wraps all NATS events into structured `LogEntry` objects for the Cortex.

### B. The Relay (Client SDK)
* **Language:** Python (`sdk/python/relay`).
* **Role:** The "Universal Adapter" for AI models, sensors, and legacy scripts.
* **Key Features:**
    * **Identity:** Defaults to `source="swarm:base"` (Generic) unless specialized (e.g., `source="qwen-72b"`).
    * **Team Awareness:** Auto-subscribes to `swarm.team.{team_id}.>` upon connection.
    * **Context Injection:** Enforces packing the `swarm_context` (Shared Memory) into every message envelope.
    * **Dynamic Binding:** Supports `subscribe()` for ephemeral topics (Ad-Hoc/Temp Teams).

### C. The Nerve Network (Protocol)
* **Format:** Protocol Buffers (`proto/swarm/v1/swarm.proto`).
* **Transport:** NATS (Subject-based messaging).
* **Critical Schemas:**
    * `MsgEnvelope`: Contains `team_id` (Routing) and `swarm_context` (State/Context Variables).
    * `LogEntry`: The "Cortex Memory" (Timestamp, TraceID, Agent, Intent, Result, ContextSnapshot).
    * `AgentProfile`: Includes `source_uri` to identify capabilities.

### D. Operations (The "Split-Lifecycle")
* **Tooling:** Python `invoke` (`tasks.py`). **Legacy `Makefile` is deprecated.**
* **Platform:** Cross-platform compatible (Windows/Linux/Mac).
* **Workflows:**
    * **Infrastructure (`inv k8s.init`):** Deploys stable services (NATS, Databases). Runs rarely.
    * **Application (`inv k8s.deploy`):** Builds and redeploys the Go Core/Relays. Runs frequently.
    * **Development (`inv k8s.bridge`):** Tunnels cluster ports (NATS:4222, HTTP:8080) to localhost for testing.

---

## 3. The "Architect Agent" Protocol
This project utilizes **Agent-Driven Development**. The User acts as Executive; the AI acts as Architect/Builder.

### The Persona Protocol
The AI must adopt specific "Specialist" personas to constrain scope and tone:
* `ACT AS: spec:arch:01`: **System Architect**. High-level design, consistency checks, PRD enforcement.
* `ACT AS: spec:devops:01`: **Tooling Specialist**. `tasks.py`, Docker, Kubernetes, CI/CD.
* `ACT AS: spec:proto:01`: **Data Specialist**. `swarm.proto`, Schema definitions.
* `ACT AS: spec:python:01`: **SDK Specialist**. Python `RelayClient`, library implementation.
* `ACT AS: spec:qa:01`: **Test Engineer**. `pytest`, `go test`, Integration scripts.

### The Execution Loop
1.  **User Intent:** User defines a goal.
2.  **Architect Plan:** AI proposes an "Execution Block" (Markdown) detailing the exact task for a specialist.
3.  **Generation:** AI generates the *exact code* or *shell commands* required.

---

## 4. Session Restoration Protocol
**Instructions for New AI Sessions:**
You are picking up the role of **Lead Architect** for Project Mycelis. This PRD describes the *intended* state. Development is fluid; reality may differ from the plan.

**To synchronize your internal context, you MUST execute the following "Gap Analysis" immediately:**

1.  **Read the Map:** Request `ls -R` to verify physical structure.
2.  **Verify the Build:** Request `cat tasks.py` to confirm the `invoke` system is active and namespaces (`proto`, `k8s`) exist.
3.  **Check the DNA:** Request `cat proto/swarm/v1/swarm.proto`. **CRITICAL:** Verify it contains `message LogEntry` and `google.protobuf.Struct swarm_context`. If missing, the Protocol Upgrade is incomplete.
4.  **Check the Relay:** Request `cat sdk/python/src/relay/client.py`. **CRITICAL:** Verify the `RelayClient` class exists and supports `source_uri` and `subscribe`.

**How to Ask for Updates:**
> "I have reviewed the Mycelis PRD v3. To align with the current deployment state, please provide the file listing and the contents of `tasks.py` and `swarm.proto`."

---

## 5. Current Implementation Status (Snapshot)
* **Infrastructure:** **ONLINE** (Kind Cluster + NATS).
* **Build System:** **ONLINE** (Migrated to `invoke`).
* **Protocol:** **ONLINE** (Verified `swarm.proto` with `LogEntry` & `swarm_context`).
* **Python SDK:** **ONLINE** (RelayClient Verified & Tested).
* **Legacy Code:** **ARCHIVED** (Moved to `legacy_archive/`).