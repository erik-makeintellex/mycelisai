# Product Requirement Document: Project Mycelis
**Version:** 6.0 (The Cortex Architecture)
**Date:** February 1, 2026
**Status:** Active Blueprint (Phase 6: Expansion & Hardening)
**Host System:** "MOTHER-BRAIN" (Windows 11 / Local Intelligence)
**Architect:** The Absolute Architecture (Tier 0)

---

## 1. Vision & Core Philosophy: Digital Biology
**The Concept:** Mycelis is not a distributed system; it is a single, cohesive cybernetic organism. It extends the user's will into the digital and physical realms.

**The Mandate:**
1.  **Symbiosis:** The system is an extension of the User (Tier 0). It acts *as* you, not just *for* you.
2.  **Unity:** A drone sensor (IoT) and a bank transaction (API) are identical signals ("Impulses") requiring processing.
3.  **Memory:** Nothing is lost. If a limb is severed (network loss), it remembers its state until reconnected.
4.  **Hierarchy:** The User commands. The Brain (Core) routes. The Limbs (Agents) obey.
5.  **Safety:** The Organism must not harm itself or its creator. This is enforced by the **Governance Layer** (The Conscience).

---

## 2. Anatomy: The Architecture Stack

### Tier 0: The Decider (User Context)
* **Role:** The source of Intent.
* **Interface:** The **Cortex Viewer** (Visual Dashboard) and **Synaptic Injector** (CLI).
* **Authority:** Absolute override capability via the "Global Freeze" (Kill Switch).

### Tier 1: The Brain (Neural Core)
* **Implementation:** Go Binary (Hardened, Non-Root) running in `mycelis-cluster`.
* **The "State Engine" Evolution:** The Core is no longer just a router; it is the keeper of state.
    * **The Archivist:** A sub-system that subscribes to `swarm.>` and persists verified `LogEntry` data to PostgreSQL.
    * **The Gatekeeper:** Intercepts high-risk intents (`amount > $50`) based on `policy.yaml`.
    * **The Cognitive Loop:** Connects to **Local Ollama** (`qwen2.5-coder:32b`) to analyze complex data or summarize logs.

### Tier 2: The Memory (Hippocampus)
* **Technology:** PostgreSQL (Bitnami) with `pgvector` enabled.
* **Role:** The Single Source of Truth.
    * **Relational:** Event logs, Governance requests, Registry state.
    * **Vector:** Semantic search of "Thoughts" and "Context."
* **Deployment:** In-Cluster (Consolidated Persistence).

### Tier 3: The Nerves (Protocol Layer)
* **Technology:** NATS (Subject-based Messaging) + Protocol Buffers (`swarm.proto`).
* **Standard:** The **Universal Event Model**.
    * `MsgEnvelope`: Contains `trace_id`, `swarm_context` (Shared State), and `payload`.
    * `LogEntry`: The verifiable history of every thought and action.

### Tier 4: The Limbs (Agent Swarm)
* **The Relay (Python SDK):**
    * **Role:** Complex logic, LLM interaction, API integrations.
    * **Resilience:** **"Black Box" Buffer**. Uses SQLite (`agent_buffer.db`) to store events when disconnected, replaying them upon reconnection.
* **The Satellites (Go Edge):**
    * **Role:** "Bare Metal" I/O for Raspberry Pi Zero / Micro-controllers.
    * **Implementation:** Static Binary (`mycelis-sat`). Maps GPIO <-> NATS directly.

---

## 3. The Synaptic Bridge: Cluster API
The Backend must provide a robust API for the UI to be effective. The UI is simply a consumer of this API.

**Base URL:** `http://localhost:8080/api/v1`

| Endpoint | Method | Purpose | Response |
| :--- | :--- | :--- | :--- |
| `/registry` | `GET` | "Who is alive?" | `[{id: "sensor-01", status: "IDLE", last_seen: "2s"}]` |
| `/memory/stream` | `GET` | "What happened?" (History) | `[{timestamp: "...", level: "WARN", msg: "..."}]` |
| `/governance/pending` | `GET` | "What needs approval?" | `[{id: "req-123", intent: "spend", context: {...}}]` |
| `/command` | `POST` | "Do this." (Injection) | `{"ack_id": "...", status: "SENT"}` |
| `/cognitive/think` | `POST` | "Analyze this." (LLM) | `{"thought": "Based on logs, drone is stuck."}` |

---

## 4. The Visual Cortex: UI Specification (Ultrathink)
**Theme:** Slate-900 / Cyber-Organism / Glassmorphism.
**Goal:** High-contrast, data-dense "Glass Cockpit."

### A. The HUD (Heads-Up Display)
* **Vitality Monitor:** A literal EKG-style line graph (Canvas) showing message throughput. Flatline = Death.
* **Token Consumption:** Real-time counter of LLM tokens used (Cost awareness).
* **Global Freeze:** A red, guarded switch. "Emergency Stop." (Broadcasts `swarm.global.freeze`).

### B. The Neural Grid (Agent Registry)
* **Visual:** A **Force-Directed Graph** or Hex Grid, not a list.
* **States:**
    * **Green Node:** Healthy/Idle.
    * **Pulse Animation:** actively processing a task.
    * **Red Border:** Disconnected (or using Offline Buffer).
* **Interaction:** Clicking a node opens the "Synapse Inspector" (Live log tail for *just* that agent).

### C. The Governance Deck (The "Conscience")
* **Layout:** "Card Stack" view.
* **Cards:**
    * **Front:** "High Spend Alert ($150)." Visual "Risk Level" borders (Red for Finance, Amber for Social).
    * **Back (Flip):** Full JSON Payload & Trace Context.
* **Action:**
    * `[A] Approve` (Keyboard Shortcut).
    * `[D] Deny`.

### D. The Synaptic Injector (Integrated CLI)
* **Location:** Bottom fixed pane. Always visible.
* **Visual:** Modeled after VS Code Terminal.
* **Function:** A **Context-Aware REPL**.
    * User types: `ask finance`...
    * UI suggests: `...report_daily_spend` (Auto-complete based on Agent capabilities).
    * **Execution:** Sends to `POST /api/v1/command`. Displays the *Agent's textual response* directly in the console.

---

## 5. Security & Infrastructure Standard ("Hardened Hull")
To comply with **Monokle** and **Production Standards**:

1.  **Immutable Identity:**
    * **Tagging:** All deployments must use `v{SEMVER}-{SHA}`.
    * **Ban:** The `:latest` tag is forbidden in the cluster.
2.  **Container Security:**
    * `runAsNonRoot: true` (User 10001).
    * `readOnlyRootFilesystem: true`.
    * `capabilities: drop ["ALL"]`.
    * `allowPrivilegeEscalation: false`.
3.  **Resource Hygiene:**
    * All containers MUST have `requests` (guaranteed) and `limits` (ceiling).
    * *Standard:* Request 50m CPU / 64Mi RAM.
4.  **Monitoring:**
    * **Internal:** Core/Agents only.
    * **External:** Monokle, OpenLens, K9s, and Cortex UI running on Host. No in-cluster Grafana.

---

## 6. Execution Roadmap & Verification

### Phase 1: The Foundation (Infrastructure) - *IN PROGRESS*
* **Task:** Implement "Immutable Identity" (Tagging) and "Hippocampus" (Postgres).
* **Verification:** `inv k8s.status` shows specific SHA tags. Monokle shows 0 security warnings.

### Phase 2: The Memory (Backend State) - *NEXT*
* **Task:** Build "The Archivist" in Go Core to save NATS logs to Postgres.
* **Task:** Expose `GET /api/v1/memory/stream`.
* **Verification:** Restart the cluster. Query API. Confirm old logs are returned (Persistence check).

### Phase 3: The Resilience (Agent Upgrade)
* **Task:** Add SQLite "Black Box" to Python `RelayClient`.
* **Verification:** "The Disconnection Drill" (Kill NATS -> Gen Logs -> Restore -> Verify Delivery).

### Phase 4: The Vision (UI Implementation)
* **Task:** Build Next.js "Cortex" connecting to the stable Cluster API.
* **Verification:** "The Rogue Agent Test" (Trigger Spend -> UI Card Appears -> Click Approve -> Agent Executes).

---

## 7. Agent Verification Protocol
*Instructions for future AI Agents joining this session.*

To confirm you are aligned with this PRD, perform this **Gap Analysis**:

1.  **Check Identity:** `cat VERSION` and `cat ops/version.py`.
    * *Expect:* Existence of versioning logic.
2.  **Check Memory:** `helm list` or `kubectl get pods`.
    * *Expect:* `mycelis-pg` (PostgreSQL) is present.
3.  **Check API:** `curl http://localhost:8080/api/v1/registry` (if Bridge is active).
    * *Expect:* JSON response of agents.
4.  **Check Security:** `cat charts/mycelis-core/values.yaml`.
    * *Expect:* `securityContext` with `runAsNonRoot: true`.

**If any file is missing or conflicts with this doc, STOP and request remediation.**