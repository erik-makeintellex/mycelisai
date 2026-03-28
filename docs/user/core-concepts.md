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
- generate images inline and persist them on request
- execute root-admin configuration work across the full platform (providers, policies, MCP, profiles, groups, and runtime settings), not only team creation

You do not need to manually pick Soma. It is the default route in Workspace chat.
On normal startup, Workspace opens with Soma already selected and starts the live stream automatically, so you should not need a recovery click just to begin.
For AI Organizations, the organization-wide AI Engine, Response Style, and Memory & Continuity posture chosen during setup become Soma's starting posture until you intentionally change them.
You can rename Soma from `Settings -> Profile -> Assistant Name`; the updated name appears across chat and operational UI labels.

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

## AI Engine / Provider

An AI engine is a curated model-provider posture used by the product.
Role routing is managed through AI engines and mission profiles.

Primary management surfaces:
- `Resources -> AI Engines` (Advanced mode)
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

Expected user-visible controls:
- global **Degraded Mode** banner actions (`Retry`, `Open Status`, and `Switch to Soma` only when you are not already on the Soma route)
- right-side **Status Drawer** with council reachability + service health
- `/system` **Quick Checks** with run + copy diagnostics actions

---

## Advanced Mode

Advanced Mode reveals deeper operational surfaces (for power operators).
Toggle from the rail footer (`Advanced: On/Off`).

Typical unlocks:
- System diagnostics depth
- Workflow Builder tab in Automations
- Workspace telemetry row (hidden in standard mode to keep chat-first layout)

---

## Users and Groups

`Settings -> People & Access` provides:
- user management elements (role, remote-provider allowance, active/disabled state)
- collaboration group management (goal/work-mode/team membership + group broadcast)

Group management is also available in `Automations -> Shared Teams`.

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

## Connected Tools

Connected tools extend agent capabilities (filesystem, fetch, memory, etc.).

Primary surfaces:
- `Resources -> Connected Tools` for server/tool visibility and install flows
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

---

## Current UI Reliability Baseline

The in-app docs and UI now assume these interaction guarantees:
1. no dead-end empty states on Automations
2. structured council error recovery in chat
3. global degraded visibility from any page
4. quick diagnostic checks directly from `/system`
