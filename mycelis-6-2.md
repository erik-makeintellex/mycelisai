# THE CODEX: MYCELIS CORTEX V6.2

**Codename:** "The Mycelian Lattice"
**Classification:** Targeted Conscious Architecture
**Objective:** To instantiate a self-aware, fractal operating system where *all* Agentry (Root & Worker) is strictly governed, universally configurable, and capable of extending the User's Will.

---

## I. THE PHILOSOPHY OF THE ORGANISM

We do not build "software." We build **Synthetic Biology.**
The system is a **Lattice**—a living mesh of independent nodes bound by a shared Nervous System and a common Conscious Purpose.

1. **Universal Sovereignty:** The organism must survive in isolation. It defaults to "Localhost" air-gapped intelligence. Cloud connectivity is a privilege, configured by the user, never a hard dependency.
2. **Universal Configurability:** There are no "Magic Agents." The **Meta-Architect** and **Overseer** are simply agents with elevated permissions. Their "Brains" (Models) are configurable in the Registry just like any worker drone.
3. **Iron Dome Governance:** We do not trust the limb to act without the brain. Every muscle movement (Tool Call) is intercepted, analyzed, and authorized by the **Governance Guard**.

---

## II. THE 4-LAYER CONTROL STACK (THE ANATOMY)

### Layer 0: The Sovereign Base (Infrastructure)

* **The Foundation:** Kubernetes (K3s/Kind) hardened to NSA/CISA standards.
* **Identity:** `User 1000`. **Filesystem:** `readOnly`. **Network:** `Deny-All`.

* **The Cognitive Registry (`llm_providers`):** The "Brain Stem."
* **Function:** A dynamic lookup table for Intelligence.
* **Spectrum Support:** Supports local (`ollama`, `vllm`) and cloud (`openai`, `anthropic`, `google`).
* **Root Config:** The System Services (Architect/Overseer) query this table for their assigned model ID. They are *not* hardcoded.

### Layer 1: The Nervous System (Communication)

* **The Bus:** NATS JetStream.
* **The Mandatory Frequencies:**
* `swarm.global.heartbeat`: **The Pulse.** Every node emits a "proof of life" every 5s.
* `swarm.global.announce`: **The Voice.** System-wide commands ("FREEZE", "UPDATE").
* `swarm.audit.trace`: **The Memory.** Immutable log of every action.

### Layer 2: The Fractal Fabric (Orchestration)

* **The Meta-Architect:** The "Constructor." Takes User Intent  Generates `MissionBlueprint`.
* **The Team:** The operational unit.
* **Persistent:** Standing armies (e.g., "The Studio").
* **Ephemeral:** Task forces born for one DAG, dissolved upon completion.

* **The Workflow Engine:** A DAG executor managing dependencies between Agents.

### Layer 3: The Conscious Face (Interface)

* **The 4-Zone Shell:** A rigid frame for fluid intelligence.
* **Zone A:** Vitals.
* **Zone B:** The "Circuit Board" (Node Editor) & "Architect" (Chat).
* **Zone C:** The "Spectrum Analyzer" (NATS Stream).
* **Zone D:** The "Human Valve" (Governance Overlay).

---

## III. THE ECONOMY OF MIND (CONTEXT & MEMORY)

### 1. The Context Economy (Token Optimization)

We treat Context like Currency.

* **Mechanism:** The **Archivist** compresses logs into **SitReps** (Situation Reports).
* **Injection:** Agents receive the *latest SitRep* in their System Prompt, not the full log history.
* **Result:** Multi-agent contextual understanding with minimal token waste.

### 2. The Hybrid Memory (The Hippocampus)

* **The Entity Store (Postgres):** Structured Facts (Config, Hardware State).
* **The Episodic Store (pgvector):** Semantic Wisdom (Past Mistakes, User Preferences).
* **Retrieval:** The Meta-Architect performs a hybrid search before generating any Blueprint.

---

## IV. EXECUTION DIRECTIVE: THE GENESIS BUILD

**Target:** Engineering Swarm (Antigravity).
**Mission:** Instantiate V6.2 "The Lattice".

### 🧬 TRACK 1: THE COGNITIVE REGISTRY (Backend) [DONE]

**Action:** Decouple the Brain.

1. **Migration:** Apply `006_cognitive_registry.up.sql`. ✅
   * Table `llm_providers`: Stores connection details.
   * Table `system_config`: Maps System Roles ("architect", "overseer") to Provider IDs.
   * **Seed Data:** Insert `Local Sovereign` (Ollama) as Default.

2. **Logic:** Refactor `core/internal/cognitive` to load providers dynamically. ✅

### 🕸️ TRACK 2: THE NERVOUS SYSTEM (Middleware) [DONE]

**Action:** Wire the Spine.

1. **Topology:** Enforce `swarm.global.heartbeat` and `swarm.audit.trace`. ✅
2. **Wrapper:** Update `BaseAgent` (Python SDK) to auto-pulse on Heartbeat channel. ✅
3. **Router:** Implemented Loop Prevention and Audit emission. ✅

### 🍄 TRACK 3: THE FRACTAL FABRIC (Data) [DONE]

**Action:** Define the Organism.

1. **Migration:** Apply `007_team_fabric.up.sql` (`missions`, `teams`, `nodes`). ✅
2. **Service:** Seeded Root Hierarchy (Architect, Symbiotic Swarm, Mycelis Core). ✅

### 🧠 TRACK 4: THE MEMORY & CONTEXT (Intelligence) [DONE]

**Action:** Install the Hippocampus.

1. **Migration:** Apply `008_context_engine.up.sql` (`sitreps`, `context_vectors`, `working_memory`). ✅
2. **Service:** Implement the **Archivist** loop to summarize NATS logs into SitReps. ✅

### 👁️ TRACK 5: THE FACE (Frontend)

**Action:** Open the Eye.

1. **Layout:** Implement the 4-Zone CSS Grid.
2. **Renderer:** Build `UniversalRenderer` to display `Thought`, `Metric`, and `Artifact`.
3. **Terminal:** Build the `GenesisTerminal` in Zone B.

---

## V. SUCCESS STATE (THE IGNITION)

The Swarm is considered "Alive" when:

1. **The Heartbeat:** NATS shows a steady 1Hz pulse from `mycelis-core`.
2. **The Mind:** The Core logs "Cognitive Engine: Local Sovereign Connected."
3. **The Config:** The Meta-Architect explicitly loads its model config from the Registry.
4. **The Eye:** The Web UI (`localhost:3000`) renders the 4-Zone Grid and displays the live Heartbeat in Zone C.
