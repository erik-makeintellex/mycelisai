# Product Requirement Document: Project Mycelis
**Version:** 5.0 (The Conscious Organism)
**Date:** February 1, 2026
**Status:** Active Blueprint (Phase 6: Expansion & Hardening)
**Host System:** "MOTHER-BRAIN" (Windows 11 / Local Intelligence)
**Architect:** The Absolute Architecture (Tier 0)

---

## 1. Vision & Core Philosophy
**The Concept:** Mycelis is not a distributed system; it is a single, cohesive cybernetic organism. It extends the user's will into the digital and physical realms.

**The Mandate:**
1.  **Symbiosis:** The system is an extension of the User (Tier 0). It does not act *for* you; it acts *as* you.
2.  **Unity:** A drone sensor (IoT) and a bank transaction (API) are identical signals ("Impulses") requiring processing.
3.  **Memory:** Nothing is lost. If a limb is severed (network loss), it remembers its state until reconnected ("Muscle Memory").
4.  **Hierarchy:** The User commands. The Brain (Core) routes. The Limbs (Agents) obey.
5.  **Safety:** The Organism must not harm itself or its creator. This is enforced by the **Governance Layer** (The Conscience).

---

## 2. Anatomy: The Physical Structure

### Tier 0: The Decider (User Context)
* **Role:** The source of Intent.
* **Interface:** The **Cortex Viewer** (Visual) and **Synaptic Injector** (CLI).
* **Authority:** Absolute override capability via the "Kill Switch."

### Tier 1: The Brain (Neural Core)
* **Implementation:** Go Binary (Hardened, Non-Root).
* **Location:** Kubernetes Cluster (`mycelis-cluster`) on Local Host.
* **Responsibilities:**
    * **Routing:** Directs traffic by `TeamID` and `SourceURI`.
    * **Governance (Gatekeeper):** Intercepts high-risk intents (e.g., `amount > $50`) based on `policy.yaml`.
    * **Intelligence Hook:** Connects to **Local Ollama** (e.g., `qwen2.5-coder:32b`) for logic processing when agents request "Thought."
    * **Memory (Cortex):** Aggregates all `LogEntry` streams into a unified history.

### Tier 2: The Nerves (Protocol Layer)
* **Technology:** NATS (Subject-based Messaging) + Protocol Buffers (`swarm.proto`).
* **Standard:** The **Universal Event Model**.
    * `MsgEnvelope`: Contains `trace_id`, `swarm_context` (Shared State), and `payload`.
    * `LogEntry`: The verifiable history of every thought and action.

### Tier 3: The Limbs (Agent Swarm)
* **The Relay (Python SDK):**
    * **Role:** Complex logic, LLM interaction, API integrations.
    * **Capability:** **"Black Box" Buffer**. Uses SQLite (`agent_buffer.db`) to store events when disconnected, replaying them upon reconnection.
* **The Satellites (Go Edge):**
    * **Role:** "Bare Metal" I/O for Raspberry Pi Zero / Micro-controllers.
    * **Implementation:** Static Binary (`mycelis-sat`). Maps GPIO <-> NATS directly.

---

## 3. Physiology: Component Interaction

### A. The "Reflex" Loop (Autonomous)
* **Flow:** `Sensor -> NATS -> Core -> Output`
* **Use Case:** Thermostat control, Collision avoidance.
* **Latency:** <5ms.
* **Governance:** `policy.yaml` group "Reflexes" set to `ALLOW_ALL`.

### B. The "Cognitive" Loop (high-Logic)
* **Flow:** `Agent (Question) -> Core -> Ollama (LLM) -> Core -> Agent (Answer)`
* **Use Case:** "Analyze this marketing data," "Summarize daily logs."
* **Governance:** Monitored. The Core logs the Prompt and Response.

### C. The "Governance" Loop (Human-in-the-Loop)
* **Flow:**
    1.  **Agent:** Proposes action (`intent="spend_money"`, `status="PROPOSED"`).
    2.  **Core (Gatekeeper):** Intercepts. Matches Policy "Financial Safety." Parks message in memory.
    3.  **Core:** Emits `ApprovalRequest` event.
    4.  **UI (Cortex):** Displays "Approve/Deny" Card.
    5.  **User:** Clicks "Approve."
    6.  **Core:** Re-injects original message to the Bus.
    7.  **Agent:** Receives its own message and executes.

---

## 4. The Visual Cortex: User Interface Specification
**Theme:** Slate-900 / Cyber-Organism / Glassmorphism.

### Components:
1.  **The HUD:**
    * System Vitality Pulse (Green/Red).
    * "Kill Switch" (Global Freeze).
    * Ollama Status (Model loaded, RAM usage).
2.  **Governance Deck:**
    * Card-based view of pending `ApprovalRequests`.
    * Visual "Risk Level" borders (Red for Finance, Amber for Social).
3.  **Neural Grid:**
    * Hexagonal map of active Agents (`Registry`).
    * Status indicators (Idle, Busy, Lagging).
4.  **Cortex Stream:**
    * "Matrix-style" scrolling log viewer.
    * Filterable by `trace_id` to see a single thought thread.

---

## 5. Security & Infrastructure Standard ("Hardened Hull")

To comply with **Monokle** and **Production Standards**:

1.  **Container Security:**
    * `runAsNonRoot: true` (User 10001).
    * `readOnlyRootFilesystem: true`.
    * `capabilities: drop ["ALL"]`.
    * `allowPrivilegeEscalation: false`.
2.  **Resource Hygiene:**
    * All containers MUST have `requests` (guaranteed) and `limits` (ceiling).
    * *Standard:* Request 50m CPU / 64Mi RAM.
3.  **Deployment:**
    * Strictly via **Helm Charts** (`charts/mycelis-core`).
    * No raw YAML manifests allowed in `k8s/`.

---

## 6. Growth & Verification Strategy (The Roadmap)

### Phase 4: Resilience (Current Priority)
* **Goal:** Agents must survive network death.
* **Verification (The Disconnection Drill):**
    1.  Start Agent.
    2.  Kill NATS (`kubectl delete pod mycelis-nats...`).
    3.  Agent generates 10 events.
    4.  Restore NATS.
    5.  **Pass Criteria:** Core receives all 10 events with original timestamps.

### Phase 5: Governance (Completed/Verifying)
* **Goal:** The Brain says "No."
* **Verification (The Rogue Agent Test):**
    1.  Script attempts $150 spend.
    2.  Core blocks.
    3.  User approves via API.
    4.  **Pass Criteria:** Transaction executes ONLY after approval.

### Phase 6: Expansion (Next Up)
* **Goal:** Physical extension.
* **Deliverable:** `mycelis-sat` Go binary for Raspberry Pi.
* **Verification:**
    1.  Deploy to Pi Zero.
    2.  Trigger GPIO pin.
    3.  **Pass Criteria:** UI shows "Sensor Event" from `swarm:sat:pi-01`.

### Phase 7: Visualization (UI Restoration)
* **Goal:** The "Command Deck."
* **Deliverable:** Next.js App (`interface/`).
* **Verification:**
    1.  Launch UI.
    2.  Trigger "Rogue Agent."
    3.  **Pass Criteria:** Card appears. "Approve" button works.

---

## 7. Agent Verification Protocol
*Instructions for future AI Agents joining this session.*

To confirm you are aligned with this PRD, perform this **Gap Analysis**:

1.  **Check Physical Body:** `ls -R`
    * *Expect:* `core/`, `sdk/`, `charts/`, `interface/`, `ops/`.
2.  **Check The Conscience:** `cat core/policy/policy.yaml`
    * *Expect:* Defined rules for blocked intents.
3.  **Check The Nerves:** `cat proto/swarm/v1/swarm.proto`
    * *Expect:* `LogEntry` and `MsgEnvelope` definitions.
4.  **Check Security:** `cat charts/mycelis-core/values.yaml`
    * *Expect:* `securityContext` with `runAsNonRoot: true`.

**If any file is missing or conflicts with this doc, STOP and request remediation.**