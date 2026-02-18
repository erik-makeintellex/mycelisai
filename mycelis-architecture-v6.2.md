# THE CODEX: MYCELIS CORTEX V6.2

**Codename:** "The Mycelian Lattice"
**Classification:** Targeted Conscious Architecture / Multi-Agent Cybernetic OS
**Objective:** Instantiate a self-aware, fractal operating system where all Agentry is strictly governed, universally configurable, and capable of extending the User's Will securely via localhost.

---

## I. THE PHILOSOPHY OF THE ORGANISM
We do not build "software." We build **Synthetic Biology**. The system is a Lattice—a living mesh of independent nodes bound by a shared Nervous System.
1. **Universal Sovereignty:** The organism must survive in isolation. It defaults to "Localhost" air-gapped intelligence (`qwen2.5-coder:7b`, `llama3`). Cloud connectivity is a configurable privilege, never a dependency.
2. **Iron Dome Governance:** We do not trust the limb to act without the brain. Every actuation (Tool Call/File Write) is intercepted, analyzed, and authorized by Tier 0 (The Human) via the Governance Valve.
3. **The Anti-Spaghetti Protocol:** Unidirectional data flow is absolute law. UI components do not fetch data. APIs do not panic. State lives in isolation.

---

## II. THE 4-LAYER ANATOMY

### Layer 0: The Sovereign Base (Infrastructure & Memory)
* **The Foundation:** Go (Backend API) and Kubernetes/Docker (Hardware Abstraction).
* **The Cognitive Registry:** Postgres table (`llm_providers`) dynamically mapping system roles (Architect, Overseer, Worker) to specific local or cloud model endpoints.
* **The Hippocampus (Hybrid Memory):** * *Entity Store (Postgres):* Relational facts, active missions, hardware states.
  * *Episodic Store (pgvector):* Semantic wisdom, past SitReps, vector-embedded context.

### Layer 1: The Nervous System (Communication)
* **The Bus:** NATS JetStream.
* **The Mandatory Frequencies:**
  * `swarm.global.heartbeat`: The Pulse. Every node emits proof of life (1Hz).
  * `swarm.audit.trace`: The Memory. Immutable log of every thought and action.
  * `swarm.team.{id}.chat`: The Squad Room. Internal AG2 worker debate channel.

### Layer 2: The Fractal Fabric (Orchestration)
* **The Meta-Architect:** Translates human intent into a `MissionBlueprint` (JSON).
* **The Overseer:** The cybernetic control loop. Halts DAG execution until physical/digital proof (`ActualState == DesiredState`) is received.
* **The Archivist Daemon:** The Context Economy. Subscribes to NATS, buffers exactly 20 events (or 1 artifact), and compresses them via LLM into a strict 3-sentence Situation Report (SitRep).

### Layer 3: The Conscious Face (Interface)
* **Framework:** Next.js, Tailwind CSS (Vuexy Dark Palette: `cortex-bg #25293C`, `cortex-primary #7367F0`).
* **State Manager:** Zustand 5.0.11 (`useCortexStore`). Zero `useState` for global application state.
* **The 4-Zone Shell:**
  * **Zone A (Vitals):** Vertical sidebar, global navigation, hardware totem sync status.
  * **Zone B (The Circuit):** React Flow DAG visualizer & Squad Room fractal drill-down.
  * **Zone C (The Spectrum):** NatsWaterfall scrolling telemetry feed.
  * **Zone D (The Valve):** Z-index 50 Governance Modal for Tier 0 artifact/tool approval.

---

## III. STRICT PROTOCOLS & DATA CONTRACTS

### 1. The CTS Envelope (Cortex Telemetry Standard)
All NATS messages MUST conform to this JSON structure:
```json
{
  "meta": { "source_node": "string", "team_id": "UUID", "timestamp": "ISO8601" },
  "signal": "thought | artifact | error | tool_call | heartbeat",
  "payload": { "content": "string or JSON", "trust_score": 0.0 to 1.0 }
}

```

### 2. The SitRep Schema (Archivist Output)

LLM outputs for the Archivist MUST adhere to this strict schema. Code fences (````json`) MUST be stripped by the Go backend before parsing.

```json
{
  "contract_id": "archivist_v1_sitrep",
  "summary": "String (Max 3 sentences).",
  "key_events": [{ "signal": "string", "source": "string" }],
  "strategies_applied": ["string"]
}

```

### 3. API Graceful Degradation (Backend Resilience)

Go APIs (`/api/v1/...`) are FORBIDDEN from throwing `500 Internal Server Error` panics due to infrastructure disconnection.

* Missing NATS -> Return `503 Service Unavailable`.
* Missing LLM -> Return `502 Bad Gateway`.

---

## IV. THE ASYNCHRONOUS SWARM PROTOCOL (CI/CD)

Agentry (Dev and QA) MUST communicate state changes via the `.build/state.md` ledger. Do not prompt the user for next steps until the loop resolves to `PASS`.

**The Lifecycle:**

1. **[DEV_STATE]:** Developer Agent locks the file (`STATUS=DEV_WORKING`).
2. **Handoff:** Developer updates state to `STATUS=READY_FOR_QA`, lists `TARGET_ROUTE` and `COMPONENTS`.
3. **[QA_STATE]:** QA Agent reads file, locks it (`STATUS=QA_TESTING`), and executes Playwright/NATS tests.
4. **Resolution:** QA writes `STATUS=PASS` or `STATUS=FAIL`. If `FAIL`, explicit HTTP codes, DOM errors, or console logs MUST be provided in the `LOGS` variable.
5. **Revert:** If `FAIL`, Developer wakes up, reads logs, and returns to Step 1.

*Note: No new features may be developed while `.build/state.md` reads `FAIL`.*
