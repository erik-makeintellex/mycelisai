# PRODUCT REQUIREMENTS DOCUMENT: MYCELIS CORTEX (V6)
**Version:** 6.1.0-STABLE
**Codename:** "The Conscious Organism"
**Date:** February 2, 2026
**Status:** IMPLEMENTATION (PHASES 1-19)
**Authored By:** The Systems Architect (Tier 1)

---

## 1. EXECUTIVE SUMMARY
Mycelis Cortex V6 is a **Multi-Tenant, Cognitive Swarm Orchestrator**. It transcends traditional "Chatbot" or "Agent" frameworks by establishing a persistent, secure, and governable environment where AI entities operate as long-running microservices.

**Core Capabilities:**
* **Hybrid Cognition:** Dynamic routing between Local Models (Qwen) and Cloud Models (GPT-4) via the **Cognitive Routing Service**.
* **Physical Agency:** Direct interaction with local files and remote systems via the **Hybrid MCP Service Mesh**.
* **Automated Provisioning:** Conversion of natural language intent into immutable **Service Manifests** via the **Provisioning Engine**.
* **Strict Governance:** NSA-grade security hardening and "Human-in-the-Loop" decision gates.

---

## 2. SYSTEM ARCHITECTURE (THE 4-TIER MODEL)

### Tier 0: The Authority (Identity & Control)
* **Role:** Human Operators & Admins.
* **Interface:**
    * **The Cortex:** Next.js Web Console ("Aero-Light" Professional Theme).
    * **The CLI (`myc`):** Headless bridge and MCP host.
* **State:** RBAC backed by Postgres (Users, Teams, Roles).

### Tier 1: The Core (Brain & State)
* **Service:** `mycelis-core` (Go 1.22+).
* **Responsibility:** Orchestration, Provisioning, Governance.
* **Components:**
    * **Cognitive Router:** Maps Intent Profiles (e.g., "Architect") to Model Providers.
    * **Provisioning Engine:** Generates and validates Service Manifests.
    * **CQA Middleware:** Enforces time limits and schema validation on cognitive transactions.

### Tier 2: The Nervous System (Transport)
* **Service:** NATS JetStream.
* **Protocol:** **SCIP** (Standardized Cybernetic Interchange Protocol).
    * *Format:* Protobuf (`SignalEnvelope`).
    * *Data Types:* `TEXT_UTF8`, `IMAGE_BINARY`, `TENSOR_FLOAT`.
* **Topology:**
    * `swarm.cognitive.>` (LLM Requests)
    * `swarm.mcp.>` (Tool Execution)
    * `swarm.telemetry.>` (Logs/Metrics)

### Tier 3: The Extremities (Agents & Tools)
* **Agents:** Standardized Containerized Services (Python/Go).
* **MCP Service Mesh:**
    * **Local Bridge:** CLI-hosted tools on User Host.
    * **Central Hub:** Kubernetes-hosted shared tools.

---

## 3. INTERACTION & CONNECTIVITY ARCHITECTURE
**Objective:** Define how Human Intent becomes System Actuation.

### 3.1. The Interaction Pipeline (Intent to Manifest)
1.  **Input:** User provides intent via Cortex UI ("Monitor the production database for slow queries").
2.  **Cognitive Processing:** The **Provisioning Engine** invokes the `architect` profile.
    * *Constraint:* "Map intent to available MCP Tools and NATS Topics."
3.  **Artifact Generation:** The Model generates a **Service Manifest (JSON)**.
    ```json
    {
      "name": "svc-db-monitor-01",
      "runtime": { "schedule": "*/1 * * * *" },
      "connectivity": {
        "subscriptions": ["swarm.telemetry.db.slow_log"],
        "publications": ["swarm.alerts.high"]
      },
      "permissions": ["mcp.postgres.read"]
    }
    ```
4.  **Validation:** The Core validates the Manifest against the schema and resource quotas.

### 3.2. The Connectivity Actuation (Manifest to Wire)
1.  **Deployment:** The Core deploys a Kubernetes Pod for the Agent.
2.  **Injection:** Connectivity rules are injected as Environment Variables.
    * `ENV_NATS_SUB_01=swarm.telemetry.db.slow_log`
    * `ENV_MCP_ACCESS=postgres.read`
3.  **Actuation:** On boot, the Agent SDK reads these variables and:
    * Establishes the NATS Subscription.
    * Authenticates with the MCP Mesh.
    * *Result:* The wiring is established **before** business logic executes.

---

## 4. FEATURE SPECIFICATIONS

### 4.1. Cognitive Routing Service
* **Goal:** Decouple Logic from Model Providers.
* **Configuration:** `config/brain.yaml`.
* **Profiles:**
    * `sentry`: Low latency (<2s). Target: **Local Qwen**.
    * `architect`: High reasoning, strict JSON output. Target: **Qwen/GPT-4**.
    * `creative`: High temperature. Target: **Claude/GPT-4**.

### 4.2. Cognitive Quality Assurance (CQA)
* **Goal:** Enforce Service Level Agreements (SLAs) on LLM interactions.
* **Logic:**
    * **Timeout enforcement:** Hard kill on slow responses.
    * **Schema Validation:** Retry loop if output is not valid JSON.
    * **Context Compression:** Strip non-essential tokens before transmission.

### 4.3. Hybrid MCP Service Mesh
* **Goal:** Universal Tool Access (Local + Remote).
* **Discovery Protocol:**
    * Bridges publish `swarm.mcp.announce` to NATS.
    * Core maintains a dynamic `ToolRegistry`.
* **Ambiguity Resolution:**
    * If multiple tools have the same name (e.g., `read_file`), the **Governance Engine** triggers a user inquiry via the UI.

---

## 5. DATA MODEL (POSTGRES SCHEMA)

### Identity & RBAC
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    username TEXT UNIQUE,
    role TEXT DEFAULT 'operator' -- admin, operator, observer
);

CREATE TABLE teams (
    id UUID PRIMARY KEY,
    owner_id UUID REFERENCES users(id),
    settings JSONB -- { "model_defaults": {...} }
);

```

### Service Registry

```sql
CREATE TABLE service_manifests (
    id UUID PRIMARY KEY,
    team_id UUID REFERENCES teams(id),
    name TEXT,
    manifest JSONB, -- The immutable JSON config
    status TEXT -- active, paused, error
);

```

---

## 6. SECURITY STANDARDS (OPERATION IRON DOME)

**Status:** PHASE 11 (PENDING)
**Standard:** NSA Kubernetes Hardening & CIS Docker Benchmark.

1. **Zero Trust Networking:**
* **Default:** Deny All.
* **Rule:** Agents can *only* talk to NATS and DNS. Direct Egress is blocked.


2. **Immutable Infrastructure:**
* Agents run with `readOnlyRootFilesystem: true`.
* State is persisted *only* to Postgres or S3.


3. **Non-Root Execution:**
* Enforced `runAsUser: 1000` via SecurityContext.
* Capabilities `DROP ALL`.



---

## 7. EXECUTION ROADMAP

| Phase | Module | Status | Objective |
| --- | --- | --- | --- |
| **8** | **Identity Layer** | 🟢 Done | Users, Teams, Settings DB. |
| **9** | **Professional UI** | 🟡 In-Flight | "Aero-Light" Console & Interactive Operator. |
| **10** | **MCP Mesh** | 🟡 In-Flight | CLI Bridge & NATS Routing. |
| **11** | **Iron Dome** | 🔴 **NEXT** | NSA Security Hardening. |
| **19** | **Provisioning** | 🔴 **NEXT** | "The Forge" (Intent-to-Manifest). |
