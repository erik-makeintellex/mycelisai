# Mycelis Service Network (Gen-5)
> **Codename**: "The Conscious Organism"
> **Standard**: [Service Release 1.0](#-service-development--release-standard)
> **Status**: Phase 6 (Expansion & Intelligence)
> **Current Identity**: `v0.6.0` (See [VERSION](VERSION))

## 🧠 The Vision
Mycelis is not a distributed system; it is a single, cohesive cybernetic organism extending the user's will. It adheres to the **V5 Blueprint** (`mycelist_v5_arch.md`) and enforces strict **Immutable Identity** for all cells.

---

## 📜 Service Development & Release Standard
**Version:** 1.0 | **Status:** ENFORCED

### 1. The "Immutable Identity" Strategy
**Objective:** Eliminate ambiguity. Every running container must be traceable to a specific commit SHA. The tag `:latest` is strictly forbidden in production.

#### 1.1 The Format
All built artifacts adhere to the **Hybrid Versioning Standard**:
`v{MAJOR}.{MINOR}.{PATCH}-{SHORT_SHA}`

* **Source of Truth:** The `VERSION` file in the root controls the Semantic Version (Human Intent).
* **Precision:** The Git Short SHA (7 chars) ensures every commit produces a unique artifact (Machine State).

**Example:**
* `VERSION` file: `0.6.0`
* Current Commit: `42f9958`
* **Resulting Tag:** `v0.6.0-42f9958`

#### 1.2 Implementation Rules
1.  **Local Dev:** It is permissible to tag `latest` for local debugging *only* (The build system warns you about this).
2.  **Cluster Deployment:** The build system (`ops/`) automatically calculates and injects the specific tag into the Helm Release.
3.  **Rollbacks:** Rolling back is deterministic—simply point Helm to the previous unique tag.

---

## 🛠️ Build System Specification (`ops/`)
We use `uv` and `invoke` to enforce this standard. The logic is abstracted in `ops/`.

### 2.1 Component: `ops/version.py`
Calculates the identity by combining `VERSION` + `git rev-parse --short HEAD`.

### 2.2 Task: Build (`inv core.build`)
* **Input:** Auto-detects version.
* **Action:**
    1.  Calculates `TAG`.
    2.  Builds Docker Image: `mycelis/core:{TAG}`.
    3.  *Warning:* Tags `latest` locally for convenience but warns against pushing it.
* **Output:** Returns `TAG` string.

### 2.3 Task: Deploy (`inv k8s.deploy`)
* **Action:**
    1.  Calls `core.build` to get the fresh `TAG`.
    2.  Loads `mycelis/core:{TAG}` into Kind.
    3.  Executes `helm upgrade ... --set image.tag={TAG}`.
* **Result:** The cluster updates to the exact code on your disk.

---

## 3. Infrastructure Specification (Consolidated Stack)
To minimize "Cluster Sprawl", we enforce a **Single Persistence** policy.

### 3.1 The "Hippocampus" (PostgreSQL)
A single PostgreSQL instance handles all long-term memory (State + Vectors).

* **Service:** `mycelis-pg` (via `mycelis-core` chart dependency).
* **Configuration:**
    * **Resources:** `100m/256Mi` (Request) -> `500m/512Mi` (Limit).
    * **Persistence:** 1GB PVC (Local Path).
* **Pinning:** Strictly pinned to `bitnami/postgresql:16.1.0-debian-11-r12`.

### 3.2 External Dependency Pinning
All external images in `values.yaml` must be pinned to a specific digest or patch version.

| Service | Image | Required Tag | Status |
| :--- | :--- | :--- | :--- |
| **Database** | `bitnami/postgresql` | `16.1.0-debian-11-r12` | **PINNED** |
| **Message Bus** | `nats` | `2.10.7-alpine3.19` | **PINNED** |
| **Logic (LLM)** | `ollama/ollama` | `0.1.20` | **PINNED** (External) |

---

## ⚡ Developer Workflow

### To Release a Feature:
1.  **Commit Code.**
2.  *(Optional)* Inspect/Bump `VERSION` file.
3.  **Run:** `uv run inv k8s.deploy`
    * *System automates:* Tagging -> Building -> Loading -> Helm Upgrading.
4.  **Verify:** `uv run inv k8s.status` -> Check that `IMAGE` column matches your new SHA.

### To Rollback:
1.  **Find previous tag:** `docker images | grep mycelis/core`
2.  **Run:** `helm upgrade ... --set image.tag=v0.6.0-OLDSHA`

---

## ⚡ Quick Start
```bash
# 1. Initialize Body (Infra)
uv run inv k8s.init

# 2. Build & Deploy Brain (Standard 1.0)
uv run inv k8s.deploy

# 3. Check Vital Signs
uv run inv k8s.status

# 4. Neural Bridge & UI
uv run inv k8s.bridge
uv run inv interface.dev
```
