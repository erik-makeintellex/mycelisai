# MASTER STATE DEFINITION: MYCELIS CORTEX (V6.1)
**Codename:** "The Conscious Organism"
**Date:** February 4, 2026
**Status:** ACTIVE TRANSITION
**Authored By:** The Systems Architect (Tier 1)

---

## 1. EXECUTIVE SUMMARY
We are pivoting from a "Flat Agent Framework" to a **"Recursive Swarm Operating System."**
The goal is to build the "Operating System for the Physical Singularity"—a platform capable of orchestrating not just chatbots, but complex, hierarchical organizations of software and hardware (IoT/Robotics) through a strategic AI interface.

### The Delta (Current vs. Target)

| Aspect | **Current State (V6.0)** | **Target State (V6.1)** |
| :--- | :--- | :--- |
| **Organization** | Flat List of Teams. | **Recursive Hierarchy** (Mission > Division > Squad). |
| **Strategy** | User manually creates agents. | **Consular Negotiation:** AI "Chief of Staff" proposes & refines architectures. |
| **Connectivity** | Ad-hoc Python scripts. | **Registry:** Standardized Connectors (IoT/API) & Blueprints. |
| **Security** | Standard Docker Container. | **Iron Dome:** NSA-Hardened, Read-Only Root, Zero Trust Network. |
| **Interface** | Passive Dashboard. | **Command Console:** Active "Aero-Light" Terminal & Split-View Negotiation. |
| **Cognition** | Single Model Dependency. | **Universal Adapter:** Hybrid Routing (Local Qwen + Cloud GPT-4). |

---

## 2. DETAILED ASPECT SPECIFICATIONS

### ASPECT 1: THE FOUNDATION (DATA & HIERARCHY)
**Goal:** Enable deep organizational nesting and "Mission-Aware" context.

* **Current State:**
    * `teams` table is a flat list.
    * Agents have no awareness of a "Grand Strategy."
* **Target Requirements:**
    * **The Mission Root:** A "Genesis" container defining the *Prime Directive* (e.g., "Symbiotic Swarm").
    * **Recursive Teams:** Teams can own other Teams (`parent_id`).
    * **Context Cascade:** Agents inherit directives from their entire ancestry tree.
    * **Schema:**
        ```sql
        CREATE TABLE missions (id UUID, directive TEXT, ...);
        ALTER TABLE teams ADD COLUMN parent_id UUID, path TEXT;
        ```

### ASPECT 2: THE BRAIN (STRATEGY & COGNITION)
**Goal:** Move from "Text Generation" to "Strategic Orchestration."

* **Current State:**
    * `cognitive` package supports basic LLM calls.
    * **Phase 20/21 Complete:** Universal Adapter & Grading System (Tier S/A/B) are live.
* **Target Requirements:**
    * **The Consular Layer:** A "Tier 1.5" service that sits between User and Provisioning.
    * **Negotiation Protocol:**
        1.  **Draft:** Consul proposes a hierarchy based on user intent.
        2.  **Refine:** User critiques the draft (Chat + Graph).
        3.  **Finalize:** Consul commits the structure to Reality.
    * **Mission Archetypes:** "One-Click" deployment of complex stacks (e.g., "IoT Swarm", "Startup Team") defined in JSON templates.

### ASPECT 3: THE BODY (PROVISIONING & REGISTRY)
**Goal:** Standardize how the system touches the world.

* **Current State:**
    * **Phase 19:** Provisioning Engine ("The Forge") is code-complete but blocked by network testing.
    * No standardized way to add "Weather" or "Serial Port" inputs.
* **Target Requirements:**
    * **The Registry:** A database of "Connectors" (Data Sources) and "Blueprints" (Agent Roles).
    * **Wiring API:**
        * `POST /connectors`: Install a "Weather Poller" container.
        * `PATCH /wiring`: Connect `weather_topic` to `agent_input`.
    * **Iron Dome (Phase 11):**
        * **Immutable Runtime:** Agents run as User 1000 with `readOnlyRootFilesystem`.
        * **Zero Trust:** Default `Deny-All` Network Policy. Only NATS/DNS allowed.

### ASPECT 4: THE FACE (INTERFACE)
**Goal:** Pivot from "Passive Observation" to "Active Command."

* **Current State:**
    * Basic Dashboard scaffolding.
    * "Mission Control" naming (Vague).
* **Target Requirements:**
    * **Theme:** "Aero-Light" (Clinical/Professional, `Zinc-50`).
    * **Root View (`/`):** **Command Console**.
        * *Status Bar:* UPLINK (NATS), BRAIN (Provider), CAPACITY (Agents).
        * *Central Feed:* List of **Active Sessions** (Drafting, Running, Paused).
        * *Input:* "Omni-Command" terminal (`>_`).
    * **Consular View (`/genesis`):** Split-screen (Chat Left / Graph Right) for negotiating drafts.
    * **Orchestration View (`/orchestration`):** The "Loom" graph for wiring and monitoring.

---

## 3. EXECUTION REGISTER (DELIVERY TRACKING)

This register tracks the **True Delivery** of the target state. A task is only "Complete" when the **Proof Command** succeeds.

### 🟥 BACKEND (CORE)
| ID | Deliverable | Target State | Proof Command (Verification) | Status |
| :--- | :--- | :--- | :--- | :--- |
| **D-1.1** | **Hierarchy Schema** | `missions` + recursive `teams` | `INSERT INTO teams (parent_id) VALUES (non_existent)` -> Fails FK. | 🔴 Pending |
| **D-2.1** | **Registry API** | Connector/Blueprint CRUD | `curl POST /api/registry/templates` -> 201 Created. | 🔴 Pending |
| **D-C.1** | **Consul Logic** | Negotiation/Drafting Service | Send "Add Node" chat -> LLM calls `AddNode` tool -> Draft updates. | 🔴 Pending |
| **D-11.1** | **Secure Image** | Non-Root Docker Image | `docker inspect image` -> User is `1000`. | 🔴 Pending |

### 🟦 FRONTEND (INTERFACE)
| ID | Deliverable | Target State | Proof Command (Verification) | Status |
| :--- | :--- | :--- | :--- | :--- |
| **D-3.1** | **Genesis Wizard** | Tabula Rasa "First Run" | Clear DB -> Load UI -> Wizard Modal appears. | 🔴 Pending |
| **D-9.2** | **Command Input** | Active Terminal on Root | Load Page -> Input is focused -> `Enter` POSTs to API. | 🔴 Pending |
| **D-C.2** | **Negotiation UI** | Split Chat/Graph View | Chat "Add Vision Team" -> Node appears on Graph instantly. | 🔴 Pending |

### 🟩 INFRASTRUCTURE (OPS)
| ID | Deliverable | Target State | Proof Command (Verification) | Status |
| :--- | :--- | :--- | :--- | :--- |
| **D-11.3** | **Network Policy** | Zero Trust (Deny All) | `kubectl exec agent -- curl google.com` -> Timeout. | 🔴 Pending |
| **D-2.4** | **Mock Source** | Runtime Connector | `kubectl get pods` shows `mock-source` running after API call. | 🔴 Pending |

---

## 4. IMMEDIATE ACTION PLAN

**The Blockers:** We cannot build the UI or the Consul without the **Data Foundation** (Phase 1).

1.  **Execute Phase 1 (Data):** Apply `003_mission_hierarchy.up.sql`.
2.  **Execute Phase 19 (Fix):** Resolve network blocking to finalize Provisioning Engine.
3.  **Execute Phase 2.5 (Registry):** Apply `004_registry.up.sql` to support Connectors.
4.  **Execute Phase 3 (Genesis):** Build the Frontend Wizard to create the first Mission.