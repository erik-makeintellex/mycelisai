# Soma Team and Channel Architecture V7

Version: `1.0`
Status: `Authoritative`
Last Updated: `2026-02-28`
Scope: Inter-team/process channels, MCP execution I/O, and shared memory/RAG contracts

This document defines the canonical channel architecture for workflow execution in V7.
Companion execution spec: `docs/architecture/UNIVERSAL_ACTION_INTERFACE_V7.md` (universal action contracts and dynamic service onboarding API).
Landscape extension: `docs/architecture/ACTUALIZATION_ARCHITECTURE_BEYOND_MCP_V7.md`.
Hardware integration profile: `docs/architecture/HARDWARE_INTERFACE_API_AND_CHANNELS_V7.md`.
If any workflow, API, or UI surface conflicts with this contract, this contract wins.

---

## Table of Contents

1. Purpose and Outcomes
2. Channel Taxonomy
3. Canonical Envelope Contracts
4. Team and Process Channel Architecture
5. MCP Execution Input/Output Architecture
6. Shared Memory and RAG Architecture for Soma and Teams
7. Channel Health and Recovery Architecture
8. UI Exposure Model by Operator Mode
9. Security, Isolation, and Governance Boundaries
10. Testing Matrix by Core Layer
11. Parallel Delivery Plan
12. Delivery Gates

---

## 1. Purpose and Outcomes

Mycelis must provide a predictable, governed execution fabric across:
- user and automation ingress
- inter-team coordination
- tool execution (internal + MCP)
- shared memory recall and writeback
- observability and recovery

Required outcomes:
1. Every meaningful action is traceable to `run_id`.
2. Every channel uses a documented envelope and health model.
3. Inter-team and MCP execution paths are explicit, not implicit.
4. Teams can share memory through controlled contracts, not ad hoc store access.
5. Degraded channels always expose recovery actions.

---

## 2. Channel Taxonomy

### 2.1 Ingress Channels

| Channel | Source | Primary API/Topic | Output |
| :--- | :--- | :--- | :--- |
| Workspace | Human operator | `POST /api/v1/council/admin/chat` | Chat response, proposals, run events |
| Direct council | Human operator | `POST /api/v1/council/{member}/chat` | Specialist response and tool trace |
| API | Programmatic caller | `POST /api/v1/chat` and related APIs | Run/event emission |
| Trigger | Event spine rule | `trigger.fired` workflow path | Child run creation |
| Schedule | Scheduler tick | `scheduler.tick` workflow path | Scheduled run execution |
| Sensor | External feed | `swarm.data.*` | CTS telemetry and mission routing |
| Low-level IoT | Device/edge control path | sensor + actuation adapters via action gateway | Long-lived monitoring/control team workflows with strict guardrails |

### 2.2 Process Channels

| Channel Type | Scope | Transport | Contract |
| :--- | :--- | :--- | :--- |
| Intra-team control | Team-local | NATS `swarm.team.{team_id}.internal.*` | CTS + mission event linkage |
| Inter-team routing | Cross-team | Mission event + trigger + NATS publish | Parent/child run chain |
| Council request-reply | Specialist RPC | `swarm.council.{agent_id}.request` | Structured chat payload |
| Global broadcast | Multi-team | `swarm.global.broadcast` | CTS envelope with provenance |

### 2.3 Execution Channels

| Channel | Dispatcher | Contract |
| :--- | :--- | :--- |
| Internal tools | InternalToolRegistry | Tool call/result events + run linkage |
| MCP tools | ToolExecutorAdapter + MCP pool | Tool call contract + MCP response normalization |
| Artifact channel | Artifact persistence + event spine | `artifact.created` event + artifact metadata |
| Memory channel | Memory service + embeddings | `memory.stored` and `memory.recalled` events |

---

## 3. Canonical Envelope Contracts

### 3.1 API Envelope

All new and migrated UI/API surfaces must normalize to:

```json
{
  "ok": true,
  "data": {},
  "error": "",
  "meta": {
    "channel": "workspace|trigger|schedule|api|sensor",
    "run_id": "",
    "team_id": "",
    "agent_id": "",
    "timestamp": ""
  }
}
```

Rules:
- UI accepts legacy raw payloads only through shared normalization helpers.
- Components never branch on raw/enveloped shape directly.

### 3.2 CTS Envelope

Transport envelope for NATS telemetry and real-time signals:
- `meta.source_node`
- `meta.timestamp`
- `signal_type`
- `payload`
- `mission_event_id` (when linked to persistent event)

### 3.3 Mission Event Envelope

Persistent audit record contract:
- `run_id`
- `tenant_id`
- `event_type`
- `severity`
- `source_agent`
- `source_team`
- `payload`
- `emitted_at`

Non-negotiable rule:
- Mutating actions must be represented in persistent mission events.

---

## 4. Team and Process Channel Architecture

### 4.1 Topic and Namespace Model

Canonical topic constants (no hardcoded subjects):
- `swarm.global.*`
- `swarm.team.{team_id}.internal.*`
- `swarm.team.{team_id}.telemetry`
- `swarm.council.{agent_id}.request`
- `swarm.mission.events.{run_id}`
- `swarm.agent.{agent_id}.interjection`

### 4.2 Team Process State Machine

Required execution state progression:
1. `draft`
2. `ready`
3. `running`
4. `degraded` (recoverable)
5. terminal: `completed|failed|halted|cancelled`

### 4.2.1 Team Lifetime Profiles

All manifested teams must declare a lifecycle profile:
- `ephemeral` (short-lived): scoped to a single run or bounded objective; terminates after completion/failure/timeout.
- `persistent` (long-lived): remains active for recurring schedules, event subscriptions, and low-level IoT watch/control loops.
- `auto`: Soma selects profile from risk, complexity, repeat intent, and channel class.

Promotion and demotion rules:
- user request for scheduled repeat promotes eligible ephemeral teams to persistent profile
- persistent teams without active schedules/subscriptions beyond threshold are candidates for graceful demotion to ephemeral
- lifecycle transitions must emit run-linked mission events for auditability

### 4.3 Inter-Team Contract

For inter-team handoff, payload metadata must include:
- `parent_run_id`
- `origin` (`workspace|trigger|schedule|api|sensor`)
- `intent_proof_id` when mutation-capable
- `governance_mode`

### 4.4 Process Guarantees

Required guarantees:
- idempotent inter-team activation by team ID
- duplicate event tolerance at subscriber edge
- deterministic replay by `run_id` timeline
- scheduled-repeat promotion path for user-requested recurrence (single action -> recurring plan -> persistent team)

---

## 5. MCP Execution Input/Output Architecture

### 5.1 Execution Path

1. Agent issues tool call.
2. Composite executor resolves internal vs MCP tool.
3. MCP adapter resolves server/tool from cache.
4. MCP pool executes call.
5. Result normalized and persisted as event(s).
6. Response returned to agent/UI with channel metadata.

### 5.2 MCP Call Contract

Required request metadata:
- `run_id`
- `team_id`
- `agent_id`
- `tool_name`
- `arguments`
- `governance_context`

Required response metadata:
- `status` (`success|error`)
- `content` (normalized text/structured payload)
- `duration_ms`
- `server_name`
- `tool_name`

### 5.3 MCP Safety Boundaries

- library-first install policy remains default
- raw install endpoints require explicit phase gate
- filesystem MCP remains sandbox-bound
- all MCP calls must emit `tool.invoked` plus completion/failure events

---

## 6. Shared Memory and RAG Architecture for Soma and Teams

### 6.1 Memory Layers

1. Working memory: in-run short context.
2. Structured memory: SitReps and artifacts.
3. Vector memory: semantic recall over embeddings.

### 6.2 Shared Team Memory Contract

Every memory write intended for reuse must carry:
- `tenant_id`
- `team_id`
- `agent_id`
- `run_id`
- `memory_kind` (`fact|strategy|artifact|recipe|summary`)
- `visibility` (`team|tenant|global`)
- `source` (`manual|tool|archivist|recipe`)

Every recall query must allow filters for:
- `tenant_id` (required)
- `team_id` (recommended default)
- optional `run_id` and `memory_kind`

### 6.3 Soma as Memory Coordinator

Soma responsibilities:
- issue memory-recall intents before high-complexity planning
- pass relevant recall context into spawned teams
- enforce visibility boundaries during cross-team retrieval

### 6.4 Required Team Memory API Track

Current APIs (available):
- `GET /api/v1/memory/search`
- `GET /api/v1/memory/sitreps`

Required v1 additions for stricter team contracts:
- `POST /api/v1/memory/team/{team_id}/store`
- `POST /api/v1/memory/team/{team_id}/query`
- `GET /api/v1/memory/team/{team_id}/policies`

Status: design-required for production-hard multi-team memory isolation.

---

## 7. Channel Health and Recovery Architecture

### 7.1 Health Dimensions

Global channel health is computed from:
- NATS connectivity
- SSE stream state
- database availability
- council reachability
- MCP pool connectivity
- scheduler/trigger readiness

### 7.2 Channel Status Model

Use only these statuses:
- `healthy`
- `degraded`
- `failure`
- `offline`
- `info`

### 7.3 Recovery Contract

Every degraded/failure state must expose:
1. What failed.
2. Likely cause.
3. Impact.
4. At least two recovery actions.
5. Diagnostic copy action.

---

## 8. UI Exposure Model by Operator Mode

### Basic Mode
- semantic channel labels
- no raw topic editing
- guided recovery only

### Guided Mode
- route templates
- impact previews
- rollback controls

### Expert Mode
- raw topic visibility
- filtered diagnostics
- advanced inspection tools

UI rule:
- channel complexity increases by mode, not by route.

---

## 9. Security, Isolation, and Governance Boundaries

1. Channel auth context must be present on all mutation paths.
2. Governance policy checks occur before irreversible side effects.
3. Memory visibility boundaries are enforced before vector recall results are returned.
4. Team-to-team delegation includes provenance and policy context.
5. No silent downgrade from governed execution to ungoverned execution.

---

## 10. Testing Matrix by Core Layer

### Layer A: Contracts and Types
- Envelope normalization tests
- DTO validation tests
- status mapping tests

### Layer B: Service and Adapter Integration
- MCP adapter success/failure/degraded tests
- memory store/query filter tests
- event linkage tests (`mission_event_id` and `run_id`)

### Layer C: API Integration
- ingress route tests (workspace/api/trigger/schedule)
- inter-team execution chain tests
- team memory API tests (when added)

### Layer D: UI and Operational E2E
- degraded banner recovery
- status drawer channel diagnostics
- inter-team run chain visibility
- MCP execution trace rendering

### Layer E: Reliability and Recovery
- forced NATS disconnect/reconnect
- SSE token expiry/recovery
- MCP server unavailable fallback
- memory retrieval timeout handling

---

## 11. Parallel Delivery Plan

### Team Atlas (Workflow Orchestration)
- Owns team lifecycle and inter-team process UX
- Delivers run-chain and handoff clarity

### Team Helios (Contracts + Memory)
- Owns shared envelope contracts and memory API design
- Delivers team-scoped memory policy and retrieval contract

### Team Circuit (Bus + Reliability)
- Owns channel health and diagnostics model
- Delivers retry/recovery flows and mode-aware bus visibility

### Team Forge (MCP Execution)
- Owns MCP call contracts, observability, and guardrails
- Delivers execution telemetry and governance mappings

### Team Sentinel (QA + Gates)
- Owns cross-lane validation and release evidence
- Enforces channel readiness gates

---

## 12. Delivery Gates

Gate 1 - Contract Gate:
- envelope and channel docs published
- adapters aligned

Gate 2 - Integration Gate:
- MCP + process + memory paths covered by integration tests

Gate 3 - UX Gate:
- channel state and recovery visible in all critical surfaces

Gate 4 - Reliability Gate:
- disconnect/reconnect and degraded paths proven

Gate 5 - Production Readiness:
- no critical channel without tests, diagnostics, and rollback path

---

End of document.

