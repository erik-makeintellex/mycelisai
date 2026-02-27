# Core Concepts

> Plain-language glossary for the terms you see in Workspace, Automations, Resources, and System.

---

## Soma

Soma is your executive orchestrator. Every Workspace message goes to Soma first.

Soma can:
- reason through your request (ReAct loop, up to 10 iterations)
- consult council specialists
- call internal and MCP tools
- return an answer or a governed proposal

You do not need to manually pick Soma. It is the default route in Workspace chat.

---

## Council

The council is a standing specialist set that Soma can consult:

| Member | Focus |
|--------|-------|
| Architect | system design and planning |
| Coder | implementation and debugging |
| Creative | writing and ideation |
| Sentry | security and risk review |

Consultations appear as Delegation Trace cards below Soma replies.
If you want direct specialist output, use the `Direct` selector in the chat header.

---

## Mission

A mission is a defined work objective and execution plan.
It is the "what" to execute.

---

## Run

A run is one concrete execution instance of a mission, with a unique `run_id`.
Runs are the audit anchor for:
- timeline events
- trigger chains
- artifacts

Mission : Run = definition : execution instance.

---

## Proposal

When your request requires mutation (files, teams, schedules, external actions), Soma returns a proposal block.

Proposals show:
- intended action
- tool path
- risk context
- explicit confirm/cancel controls

No mutation executes until confirmation.

---

## Brain / Provider

A brain is a configured model provider endpoint (local or remote).
Role routing is managed through brains and mission profiles.

Primary management surfaces:
- `Resources -> Brains`
- `Settings -> Profiles`

---

## Event

Every important step emits a structured event.
Events are persisted and used to build run timelines and causal traces.

Common examples:
- `mission.started`
- `tool.invoked`
- `tool.completed`
- `tool.failed`
- `mission.completed`
- `mission.failed`

---

## Operational Status UX (Gate A)

V7 includes global operator-recovery UX:

- **Degraded Mode Banner**: appears when critical subsystems degrade
- **Status Drawer**: global health panel (open via ribbon or floating status action)
- **Structured Council Error Card**: retry/reroute/copy diagnostics in chat
- **Focus Mode (`F`)**: collapse ops panel while keeping critical status strip
- **System Quick Checks**: run targeted checks from `/system`

These are designed to keep workflows recoverable without page switching.

---

## Advanced Mode

Advanced Mode reveals deeper operational surfaces (for power operators).
Toggle from the rail footer (`Advanced: On/Off`).

Typical unlocks:
- System diagnostics depth
- Neural Wiring tab in Automations

---

## Trust Score and Governance

Agent outputs include trust context.
Lower-trust mutation paths are routed to approvals instead of auto-executed.

See:
- `Automations -> Approvals`
- `docs/user/governance-trust.md`

---

## NATS

NATS is the internal event spine and signaling bus.
Agents, triggers, runtime health, and mission events all depend on it.

If NATS degrades, UI enters degraded mode and offers recovery actions.

---

## MCP Tools

MCP extends agent capabilities (filesystem, fetch, memory, etc.).

Primary surfaces:
- `Resources -> MCP Tools` for server/tool visibility and install flows
- `Workspace` for actual tool usage via agent execution

Baseline profile (V7):
- `filesystem` and `fetch` are bootstrap defaults
- `memory` is curated install
- `artifact-renderer` remains planned

---

## Standardized Resource API Contract

Mycelis resource surfaces are converging on one API envelope pattern:
- payloads may arrive as either `{ ok, data, error }` or raw data
- UI store/actions normalize both into typed state before rendering
- component-level parsing is minimized so new channels can reuse the same contract

Why this matters:
- adding new AI resource channels (tools, services, hardware, future RAG paths) does not require bespoke parsing logic per screen
- degraded/error behavior stays consistent across Workspace, Resources, and System surfaces
