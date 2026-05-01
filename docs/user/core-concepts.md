# Core Concepts
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

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
At the root organization workspace, admins should ask Soma to create, reshape, and coordinate teams rather than bypassing Soma to manually assemble the operating structure first.
You can rename Soma from `Settings -> Profile -> Assistant Name`; the updated name appears across chat and operational UI labels.
Root admins can also configure output-model routing for the organization so all team members either share one default model or inherit detected models by output type.

---

## Team Leads

Created teams are entered through a focused lead counterpart, not through a generic shared chat bucket.

Team Leads:
- own the immediate lane context for that team
- work from scoped team memory and delivery context first
- can coordinate back through Soma when broader organization context, RAG retrieval, or cross-team direction is needed
- inherit the organization's default output model unless the admin has enabled detected output-type routing for planning, research, code, or vision work

Use Soma at the root workspace when you want to:
- create or manage teams
- reshape organization structure
- coordinate across multiple teams

Use a Team Lead workspace when you want to:
- work inside one team lane
- review that lane's inputs, deliveries, and agent roster
- keep interaction focused before escalating back to Soma

Use `Teams` when you want to:
- review which teams currently exist
- open a specific lead workspace
- define or edit the reusable member templates Soma should use when specializing future teams
- decide when a certain kind of work should prefer a specific specialist role, model, toolset, or output contract
- enter the guided team-creation workflow at `/teams/create` instead of building a new team from a raw field list
- launch a temporary workflow group directly from Soma's guided execution path when the design should move immediately into a bounded delivery lane

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

For AI Organizations, admins can also assign output models by delivery type:
- `single_model`: one default model for all team members
- `detected_output_types`: a shared default plus specialized models for general text, research and reasoning, code generation, and vision analysis

Current self-hosted starting points in the product:
- `Qwen3 8B` for strong general text delivery
- `Llama 3.1 8B` for research and reasoning
- `Qwen2.5 Coder 7B` for code generation
- `LLaVA 7B` for vision analysis

The model inventory can include several models on the same Ollama, vLLM, or LM Studio host. When the admin has not pinned a model for a requested output, Soma should prefer installed self-hosted models that match the detected output type, use larger local candidates such as `Qwen3 14B`, `Qwen2.5 Coder 14B`, or `DeepSeek Coder V2 16B` when the host and latency budget fit, and ask the owner/admin before running a model-behavior review or changing the saved routing policy.

Ollama text and vision models do not automatically mean image or voice generation is configured. Soma can use them to plan prompts, write website/code artifacts, or critique images; actual pixel/audio output needs the configured media engine.

Primary management surfaces:
- `Resources -> AI Engines` (Advanced mode)
- `Settings -> Profiles`
- `AI Organization -> AI Engine Settings`

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
- environment ownership, identity posture, and shared-Soma governance posture
- deploy-owned edition/auth fields as read-only release posture unless the current edition enables management actions

See [Settings And Access](settings-access.md) for the current operator contract around Profile, People & Access, Auth Providers, connected-tool boundaries, and access-denied recovery.

Collaboration groups now have their own dedicated workflow surface:
- `Groups` for standing and temporary group creation, plus archived temporary-group review after closure
- compact group selection in a left rail, with the selected group's data, config posture, broadcast/review workflow, and retained outputs shown in the main panes
- focused group review, broadcast while active, output/contributing-lead summaries, and retained output visibility after archive
- quick entry into the attached team-lead lanes

The root Soma home also includes a filtered live interaction stream:
- review active team output from the admin surface
- filter by multiple teams
- filter by available activity aspects such as status, results, artifacts, tools, governance, and errors

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

Current baseline profile:
- curated library installs are the supported default path
- `filesystem` and `fetch` are common curated entries, not assumed bootstrap defaults in supported runtime lanes
- `memory` is curated install and remains distinct from Mycelis-governed memory/context layers
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
