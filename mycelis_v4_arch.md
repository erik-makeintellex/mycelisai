# Product Requirement Document: Project Mycelis
**Version:** 4.0 (The Conscious Organism)
**Date:** February 1, 2026
**Status:** Active Blueprint (Phase 4/5)
**Architect:** The Absolute Architecture (Tier 0)

---

## 1. Core Philosophy: Digital Biology
**The Concept:** Mycelis is not a distributed system; it is a single cohesive cybernetic organism. It extends the user's will into the digital and physical realms.

**The Mandate:**
1.  **Unity:** A drone sensor (IoT) and a bank transaction (API) are identical signals ("Impulses") requiring processing.
2.  **Memory:** Nothing is lost. If a limb is severed (network loss), it remembers its state until reconnected.
3.  **Hierarchy:** The Core (Mother Brain) commands. The Limbs obey. The User rules.
4.  **Safety:** The Organism must not harm itself or its creator. This is enforced by the **Governance Layer** (Gatekeeper).

---

## 2. Anatomy: The Physical Structure

### A. The Brain: Neural Core (Server)
* **Type:** Go Binary (Kubernetes `mycelis-cluster`).
* **Role:** The centralized decision engine and "Source of Truth."
* **New Capability: The Gatekeeper (Conscience).**
    * It no longer just routes; it *evaluates*.
    * **Policy Engine:** Loads `core/policy/policy.yaml`. Intercepts intents based on risk (e.g., "Spend > $50").
    * **Arbitration:** Holds high-risk tasks in a "Pending Synapse" buffer until the User approves via the Interface.

### B. The Limbs: Relay Agents (Python SDK)
* **Type:** Python `RelayClient`.
* **Role:** The "Universal Adapter" for complex, logic-heavy tasks (LLMs, Data Processing).
* **New Capability: The Black Box (Muscle Memory).**
    * **Store-and-Forward:** Integrated SQLite buffer (`agent_buffer.db`).
    * **Resilience:** If NATS is unreachable, logs/events are saved locally. They are "replayed" upon reconnection with original timestamps.

### C. The Satellites: Neural Extensions (Go Edge)
* **Type:** Static Go Binary (`mycelis-sat`).
* **Role:** Lightweight, "bare metal" I/O for Edge Devices (Raspberry Pi Zero).
* **Optimization:** <10MB size. Direct hardware mapping (GPIO <-> NATS). No Python overhead.
* **Function:** Acts as a remote sensory organ. It does not "think"; it reports raw data to the Core for LLM processing.

### D. The Visual Cortex: User Interface
* **Type:** Next.js / React Application (`interface/`).
* **Role:** The "God View" for the User.
* **Features:**
    * **Live Stream:** Visualizes `LogEntry` flow (The Matrix view).
    * **Approval Deck:** The interface for the Gatekeeper. Shows "Pending Proposals" (e.g., "Allow Marketing Team to Tweet?").
    * **Registry Map:** Visualization of all active Limbs and Satellites.

---

## 3. Physiology: The Data Flow

### A. The Universal Event Model (The Impulse)
All data, regardless of source, is normalized into the `LogEntry` and `MsgEnvelope` schema.

| Signal Type | Source | Normalized Intent | Context |
| :--- | :--- | :--- | :--- |
| **Physical** | `swarm:sat:drone_01` | `intent="collision_warning"` | `{"velocity": 12.5, "obstacle": "tree"}` |
| **Financial** | `swarm:api:stripe` | `intent="high_spend_alert"` | `{"amount": 150.00, "merchant": "aws"}` |
| **Social** | `swarm:bot:twitter` | `intent="sentiment_spike"` | `{"topic": "mycelis", "score": 0.98}` |

### B. The Memory Stream (Cortex)
* **Requirement:** Consistency across disconnection.
* **Traceability:** Every action includes a `trace_id`. We must be able to trace a "Tweet" back to the "Sensor Event" that triggered it.

---

## 4. Psychology: Governance & Safety

### A. The Policy Configuration (`policy.yaml`)
A strict set of rules defining the organism's autonomy levels.

```yaml
policies:
  - name: "Autonomous Reflexes"
    teams: ["sensors", "telemetry"]
    action: ALLOW_ALL

  - name: "Financial Safety"
    teams: ["finance"]
    rules:
      - condition: "amount > 50.00"
        action: REQUIRE_APPROVAL
        message: "High value transaction detected."

  - name: "Reputation Guard"
    teams: ["marketing", "social"]
    rules:
      - intent: "publish"
        action: REQUIRE_APPROVAL