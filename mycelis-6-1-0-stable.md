# PRODUCT REQUIREMENTS DOCUMENT: MYCELIS CORTEX V6.1
**Codename:** "The Conscious Organism"
**Version:** 6.1.0 (Target State)
**Authored By:** Systems Architect (Tier 1)
**Scope:** Infrastructure, Backend Core, Interface, and Governance.

---

## 1. EXECUTIVE SUMMARY
**The Vision:** Mycelis Cortex is not an "Agent Framework." It is a **Sovereign Operating System for the Physical Singularity**. It bridges the gap between High-Level Human Intent (Strategy) and Low-Level Hardware Execution (IoT/Robotics).

**The Pivot:** We are moving from a "Flat Chatbot" model to a **"Recursive Swarm"** model.
* **Recursive:** Missions can own Sub-Missions. Teams can own Sub-Teams.
* **Physical:** The system is aware of "Ghost Nodes" (Hardware) and "Virtual Services" (MCP).
* **Strategic:** The User negotiates with a "Consul" (Chief of Staff) rather than micro-managing Python scripts.

---

## 2. INFRASTRUCTURE & SECURITY ("THE IRON DOME")
**Standard:** NSA/CISA Kubernetes Hardening & CIS Benchmarks.

### 2.1. The Universal Security Paradigm
Every container deployed by the "Provisioning Engine" (The Forge) MUST adhere to this strict manifest:
* **User Identity:** `runAsUser: 1000`, `runAsGroup: 3000`. Root (`0`) is FORBIDDEN.
* **Filesystem:** `readOnlyRootFilesystem: true`. Writable layers are strictly limited to `/tmp` (EmptyDir).
* **Capabilities:** `drop: ["ALL"]`. No privileged escalation.
* **Network (Zero Trust):** Default `Deny-All` NetworkPolicy.
    * *Allow:* Egress to NATS (`4222`), K8s API (`443` - Internal), and DNS (`53`).
    * *Block:* All public Internet egress unless explicitly whitelisted via Governance Guard.

### 2.2. Connectivity Architecture (Dual-Plane)
* **The Nervous System (NATS JetStream):**
    * Used for high-frequency, asynchronous Agent-to-Agent communication.
    * Topic Standard: `swarm.{mission}.{team}.{signal}`.
* **The Tool Plane (MCP - Model Context Protocol):**
    * Used for connecting the Swarm to external Tools (Stripe, GitHub, Local Files).
    * The Core acts as an **MCP Host**, mounting external servers and exposing them to Agents as native functions.

---

## 3. CORE LOGIC ARCHITECTURE ("THE BRAIN")

### 3.1. The Cognitive Engine (`core/internal/cognitive`)
* **Role:** The Router and Reasoning Unit.
* **Capability:** **Universal Adapter.** dynamically routes prompts between Local Models (Qwen/Llama via Ollama) and Cloud Models (GPT-4/Claude) based on the `Mission.Priority`.

### 3.2. Cortex Memory (`core/internal/memory`)
* **Role:** Long-term persistence and Context Retrieval.
* **Components:**
    * **SQL (Postgres):** Structured data (Missions, Nodes, Logs).
    * **Vector (pgvector):** Semantic search for "Thought History."
* **The "Archivist" Pattern:** A background service that watches NATS logs and asynchronously indexes them into Cortex Memory.

### 3.3. The Governance Guard (`core/internal/governance`)
* **Role:** The "Gatekeeper" of Action.
* **Logic:** Intercepts every **Tool Call** (e.g., "Delete File", "Transfer Funds").
    * *Low Risk:* Auto-Approve.
    * *High Risk:* Halt execution. Trigger **Zone D (Decision Overlay)** in UI.

### 3.4. Consular Operations (`core/internal/consul`)
* **Role:** The Strategic Layer.
* **Function:** Manages the **Negotiation Protocol**. It maintains the `OrganizationDraft` (state in Redis) while the user edits the team structure, only committing to Postgres upon "Finalization."

---

## 4. INTERFACE ARCHITECTURE ("THE STANDARD OPERATING VIEW")
**Philosophy:** "Generative Content within a Rigid Frame."

### 4.1. The 4 Immutable Zones
To ensure stability while allowing infinite agent variety, the UI is divided into four zones.

| Zone | Name | Role | Behavior |
| :--- | :--- | :--- | :--- |
| **A** | **Command Rail** | Navigation & Vitals | Fixed Left Vertical. Never scrolls. Shows Identity & System Heartbeat. |
| **B** | **Active Workspace** | The Task Canvas | **Dynamic.** Renders the *current* tool (Terminal, Graph, Doc). |
| **C** | **Activity Stream** | The Black Box | Fixed Right Vertical. Immutable feed of Thoughts/Actions/Events. |
| **D** | **Decision Overlay** | Governance | Modal/Toast. Blocks interaction for critical human approvals. |

### 4.2. The Universal Renderer
The Frontend is "Dumb." It does not know what a "Drone" is. It only knows how to render **Universal Envelopes**.
* **`type: "thought"`**: Collapsible reasoning traces.
* **`type: "metric"`**: Sparklines/Pills for numbers (Battery, Cash).
* **`type: "artifact"`**: Cards for Files/Images/Code.

---

## 5. DETAILED UI SPECIFICATIONS (PER PAGE)

### 5.1. The Genesis Terminal (Route: `/genesis`)
* **User Intent:** "I want to initialize a new Swarm or Mission."
* **Zone B Configuration:** **Split View (Generative Chat + Topology Graph).**

#### Element A: The Pathfinder Deck (Initial State)
* **Concept:** Three large interactive cards to guide the user's start.
* **Cards:** "Architect" (Manual), "Commander" (Template), "Explorer" (Demo).
* **API:** None (Static Route).

#### Element B: The Resource Grid (The "Ghost" Visualizer)
* **Concept:** A visual map of "Unassigned Matter" (Nodes) vs "Assigned Spirit" (Agents).
* **Interaction:** Drag-and-Drop a "Ghost Node" (from the edge) onto a "Mission Node" (center) to link them.
* **Data Source:** `GET /api/v1/nodes/pending` (Polled).
    * *Returns:* `[{ "id": "rpi-01", "status": "pending", "capabilities": ["gpio"] }]`.
* **User Empowerment:** The user feels like a "God" bestowing purpose on inert matter.

#### Element C: The Consular Chat (The Negotiator)
* **Concept:** Natural Language interface to modify the Graph.
* **API:** `POST /api/v1/consul/negotiate`.
    * *Input:* "Add a Sentry Team."
    * *Output:* Updated `OrganizationDraft` JSON.
* **GenUI Behavior:** The Chat triggers a re-render of the Graph in real-time.

---

### 5.2. The Command Console (Route: `/command`)
* **User Intent:** "Active Monitoring & High-Level Direction."
* **Zone B Configuration:** **The "Loom" (Orchestration Graph).**

#### Element A: The Omni-Input (Terminal Bar)
* **Concept:** A CLI-like input at the bottom of Zone B.
* **Capabilities:** Slash Commands (`/deploy`, `/halt`), Natural Language ("Scan the perimeter").
* **API:** `POST /api/v1/command/exec`.

#### Element B: The Live Topology (The Graph)
* **Concept:** Visualization of the active NATS mesh.
* **Visuals:** Nodes pulse when publishing messages. Links glow based on traffic volume.
* **Interaction:** Clicking a Node opens its **Context Inspector** (Memory, Logs, Config).

---

## 6. PROTOCOLS & DATA STANDARDS

### 6.1. The Bootstrap Protocol ("The Handshake")
How a new Node (Client) joins the Swarm.

1.  **Boot:** Node reads local file `/etc/mycelis/identity.yaml` (UUID + Cert).
2.  **Connect:** Node establishes TLS connection to NATS (`swarm.bootstrap`).
3.  **Announce:** Node publishes Identity + Capabilities.
4.  **Wait:** Node listens on `swarm.node.{uuid}.config`.
5.  **Adoption:** User (via Consul) approves Node. Core publishes `Config` (Mission Context).
6.  **Activate:** Node applies Config and switches to "Active" mode.

### 6.2. The Data Envelope Standard (CTS - Cortex Telemetry Standard)
All Agents MUST output data in this format to be renderable by Zone C.

```json
{
  "meta": {
    "id": "uuid",
    "timestamp": "iso8601",
    "source_node": "hardware-id",
    "source_agent": "agent-role-id"
  },
  "type": "metric", // or 'thought', 'artifact', 'log'
  "payload": {
    "label": "CPU Temp",
    "value": 45.2,
    "unit": "C"
  }
}

```

---

## 7. EXECUTION REGISTER (DEVELOPER TARGETS)

### Priority 1: The Foundation (Backend)

* [ ] **Refactor Terminology:** Rename `hippocampus` -> `memory`.
* [ ] **Iron Dome Audit:** Verify `core/deploy` manifests match the Security Paradigm.
* [ ] **Bootstrap Service:** Implement the `swarm.bootstrap.announce` listener.

### Priority 2: The Face (Frontend)

* [ ] **App Shell:** Implement the 4-Zone Layout (CSS Grid).
* [ ] **Universal Renderer:** Build `ThoughtCard`, `MetricPill`, `ArtifactCard`.
* [ ] **Genesis Terminal:** Build the "Pathfinder" Cards and "Resource Grid."

### Priority 3: The Intelligence (Cognitive)

* [ ] **Consul Service:** Connect the LLM to the `OrganizationDraft` state machine.
* [ ] **MCP Bridge:** Implement the server capability to mount external tools.