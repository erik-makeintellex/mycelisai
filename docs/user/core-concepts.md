# Core Concepts

> A plain-language reference for every key term you'll encounter in Mycelis Cortex.

---

## Soma

Soma is your **executive intelligence** — the first agent that receives every message you type in Workspace. Soma reads your intent, decides which specialists to consult, and synthesises a coherent response. You always talk to Soma; Soma handles routing to the council internally.

Soma runs on the `admin` agent with up to **5 ReAct iterations** — meaning it can call multiple tools (research memory, check system status, consult specialists) before replying.

**When Soma appears:** The header in Workspace chat reads "Soma" with a cyan indicator. You never need to select it — it is always the default.

---

## The Council

The council is a set of **standing specialist agents** that Soma can consult when your message requires deeper expertise. Council members are always running and available.

| Member | Focus |
|--------|-------|
| **Architect** | System design, blueprints, technical planning |
| **Coder** | Code generation, file operations, debugging |
| **Creative** | Writing, narratives, visual concepts |
| **Sentry** | Security review, risk analysis, threat assessment |

Council consultation is **invisible to you by default** — Soma decides when to consult and synthesises the result into a single answer. A **Delegation Trace** card appears under Soma's reply showing which member contributed and a brief summary of their input.

You can also talk to a council member directly using the **Direct** button in the chat header — useful when you want a specialist's raw perspective without Soma's synthesis.

---

## Mission

A mission is **a defined unit of work** — a goal, a set of teams to execute it, and a lifecycle (draft → active → completed / failed). Missions are created through Soma chat or the LaunchCrew dialog.

A mission is a *template*. It describes what should happen. Actual execution happens in **Runs**.

---

## Run

A run is **a single execution instance** of a mission. Every time a mission executes, a new run record is created with a unique `run_id`. Runs have timelines — ordered sequences of events showing exactly what happened.

> Mission : Run = Class : Instance

You can view a run's timeline at `/runs/{run_id}` — accessible via the "Mission activated" link that appears in chat after confirmation.

---

## Blueprint / Proposal

When Soma detects that your message requires a **system mutation** (creating files, spawning teams, scheduling work, calling external services), it generates a **proposal** — a structured plan showing:

- What action will be taken
- Which tools will be called
- Any parameters or constraints

The proposal appears as a **ProposedActionBlock** in the chat. You must explicitly click **Confirm & Execute** (or **Launch Crew**) to activate it. Nothing executes automatically.

---

## Brain / Provider

A **Brain** is a configured inference provider — a connection to an LLM endpoint. Mycelis supports multiple brains simultaneously:

| Type | Examples |
|------|---------|
| Local | Ollama, vLLM, LM Studio |
| Commercial | OpenAI, Anthropic, Google |

Each **council role** (admin, architect, coder, etc.) is mapped to a brain via **profiles** in the Cognitive Matrix. You can switch which brain handles which role from `Settings → Matrix`.

A brain can be **enabled or disabled** at runtime. If no brain is available for a role, that agent enters a degraded state.

---

## Event

Every significant action in the system emits a **structured event** persisted to the database and published to NATS. Events form the run timeline.

| Event Type | When |
|---|---|
| `mission.started` | A run begins |
| `tool.invoked` | An agent calls a tool |
| `tool.completed` | A tool finishes successfully |
| `tool.failed` | A tool call errors |
| `agent.started` | An agent begins processing |
| `memory.stored` | A fact or artifact is stored |
| `memory.recalled` | Memory was queried |
| `artifact.created` | A file/document/code artifact was produced |
| `mission.completed` | Run finishes successfully |
| `mission.failed` | Run terminates with an error |

Events are viewable in the **Run Timeline** and in the **System → NATS** panel (Advanced Mode).

---

## Advanced Mode

**Advanced Mode** unlocks deeper operational panels that are hidden by default to keep the interface focused for everyday use.

Toggle it with the **eye icon** at the bottom of the navigation rail.

**Unlocks:**
- **System** nav item — health dashboard, NATS monitor, database inspector, debug tools
- **Neural Wiring** tab in Automations — low-level agent DAG editor

Advanced Mode state persists across sessions (stored in Zustand + localStorage).

---

## Trust Score

Every agent output carries a **trust score** (0.0–1.0). Outputs below the governance threshold are paused for human review in the **Approvals** queue rather than being automatically executed.

The default threshold is `0.7`. Commercial providers typically score `1.0` (fully trusted); local models default to `0.5`.

---

## NATS

NATS is the **message bus** connecting all agents and services internally. You don't interact with it directly, but the System → NATS panel (Advanced Mode) shows the live signal waterfall — every message flowing through the system in real time.

All agent-to-agent communication, heartbeats, governance signals, and event emissions travel over NATS topics.

---

## MCP Tools

**MCP (Model Context Protocol)** tools extend what agents can do. Built-in MCP tools include filesystem access, semantic memory, artifact rendering, and web research.

Additional servers can be installed from `Resources → Catalogue` or `Settings → Tools`. Each installed server exposes one or more tools that agents can call during ReAct loops.
