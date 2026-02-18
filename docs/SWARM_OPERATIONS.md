# Swarm Operations

> Back to [README](../README.md) | See also: [API Reference](API_REFERENCE.md) | [Cognitive Architecture](COGNITIVE_ARCHITECTURE.md)

Mycelis uses a **Blueprint → Manifest → Activation** pipeline. Users define what they want in natural language; the system decomposes it into teams and agents, then activates them as live processes on the NATS bus.

## Hierarchy

```
Admin (generic user proxy — always on)
  └── Council Member (specialist: architect, coder, creative, sentry — always on)
        └── Agentry Team (1+ agents — mission-scoped, created on demand)
```

- **Admin** handles Mission Control chat and routes directives.
- **Council Members** are standing specialists available for broadcast and consultation.
- **Agentry Teams** are spawned per-mission with specific roles, inputs, and outputs.

## Standing Teams (Always-On)

Defined as YAML manifests in `core/config/teams/`. Loaded by `Registry.LoadManifests()` at `Soma.Start()`. No blueprint generation required.

| Team | ID | Members | Tools | Purpose |
| :--- | :--- | :--- | :--- | :--- |
| **Admin** | `admin-core` | admin (max 5 ReAct iterations) | consult_council, delegate_task, search_memory, list_teams, list_missions, get_system_status, list_available_tools, list_catalogue, generate_image, remember, recall, publish_signal, read_signals, read_file, write_file | Orchestrator — routes user chat, consults council, delegates to teams, senses bus |
| **Council** | `council-core` | architect, coder, creative, sentry | Per-member: search_memory, store_artifact, remember, recall, read_file, write_file, publish_signal, read_signals + role-specific (generate_blueprint, list_catalogue, generate_image, list_available_tools) | Standing specialists, individually addressable via `swarm.council.<id>.request` |

To add a standing team: create a YAML file in `core/config/teams/` following the manifest schema:

```yaml
id: my-team-id
name: My Team
type: action
description: "Purpose of this team"
members:
  - id: agent-one
    role: cognitive
    system_prompt: "You are a specialist in..."
    model: ""  # blank = use default
inputs:
  - swarm.global.broadcast
deliveries:
  - swarm.team.my-team-id.signal.status
```

## Runtime Context Injection

Every agent receives **live system state** injected into its system prompt before processing any message. This is built by `InternalToolRegistry.BuildContext()` and includes:

1. **Active Teams** — roster with member IDs and counts
2. **Agent Identity & NATS Topology** — agent ID, team ID, trigger/respond bus subjects, direct address topic, team inputs/deliveries
3. **Cognitive Engine** — active providers with endpoints, models, and profile routing
4. **Installed MCP Servers** — server names, transport, status, and discovered tools
5. **Interaction Protocol** — pre-response checklist (recall, consult, delegate, MCP tools) and post-response protocol (remember, store artifact, report)

This ensures agents are never "context-blind" — they know what teams exist, what tools are available, and how to reach other agents.

## Internal Tool Registry

17 built-in tools available to agents without external MCP servers:

| Tool | Description | Primary Users |
| :--- | :--- | :--- |
| `consult_council` | NATS request-reply to a council member (30s timeout) | Admin |
| `delegate_task` | Publish task to a team's trigger topic | Admin |
| `search_memory` | Semantic vector search over SitReps | Admin, Council |
| `list_teams` | Active team roster with member counts | Admin |
| `list_missions` | Active missions from DB | Admin |
| `get_system_status` | Goroutines, heap, LLM tokens/sec | Admin |
| `list_available_tools` | All internal + MCP tools | Admin, Council |
| `generate_blueprint` | Meta-Architect intent → blueprint | Architect |
| `list_catalogue` | Agent catalogue templates | Architect, Admin |
| `remember` | Store fact/preference/goal (RDBMS + vector) | All |
| `recall` | Semantic + keyword recall of stored memories | All |
| `store_artifact` | Persist code/doc/data artifact | All |
| `publish_signal` | Publish to any NATS topic | All |
| `read_signals` | Subscribe and collect NATS messages | All |
| `read_file` | Read local filesystem (32KB max) | All |
| `write_file` | Write to local filesystem | All |
| `generate_image` | Diffusers media engine (SDXL) | Creative, Admin |

The `CompositeToolExecutor` wraps both `InternalToolRegistry` and `mcp.ToolExecutorAdapter` behind a single `MCPToolExecutor` interface — internal tools are tried first, MCP falls back.

## Agent Blueprints (`/catalogue`)

Before building missions, populate the Agent Catalogue with reusable templates.

Each **AgentTemplate** defines:

| Field | Purpose |
| :--- | :--- |
| `name` | Human-readable identifier (e.g., "Email Analyst") |
| `role` | Category — `cognitive`, `sensory`, `actuation`, or `ledger` |
| `system_prompt` | LLM persona instructions |
| `model` | Model override (blank = use profile default from `cognitive.yaml`) |
| `tools` | MCP tool names this agent can invoke |
| `inputs` | NATS topics this agent listens to |
| `outputs` | NATS topics this agent publishes to |
| `verification_strategy` | `semantic` (LLM-graded) or `empirical` (script-validated) |
| `verification_rubric` | Grading criteria for proof-of-work |

**CRUD via UI:** Navigate to `/catalogue` → create, edit, or delete agent cards via the editor drawer.
**CRUD via API:** `GET/POST /api/v1/catalogue/agents`, `PUT/DELETE /api/v1/catalogue/agents/{id}`.

## Negotiate a Mission Blueprint (`/wiring`)

The **ArchitectChat** translates natural-language intent into a structured **MissionBlueprint**.

1. Navigate to `/wiring` (or `/architect`).
2. Type a mission goal: *"Build a team that monitors weather data and sends email summaries."*
3. The Meta-Architect decomposes the intent into teams, agents, and constraints.
4. The blueprint renders as a **ghost-draft DAG** on the CircuitBoard (ReactFlow canvas).

**Blueprint structure:**

```
MissionBlueprint
├── mission_id: "mission-<timestamp>"
├── intent: "Monitor weather and send email summaries"
├── teams[]
│   ├── Team: "Weather Sensors"
│   │   └── agents[]: weather-poller (sensory), weather-parser (cognitive)
│   └── Team: "Email Dispatch"
│       └── agents[]: email-sender (actuation)
└── constraints[]
```

## Instantiate the Swarm

1. Click **INSTANTIATE SWARM** (green button, bottom-center of CircuitBoard).
2. The system executes `commitAndActivate()`:
   - **Persist:** Inserts mission, teams, and agent manifests into the database.
   - **Activate:** Soma converts the blueprint into `TeamManifest` objects and spawns live teams.
3. Ghost-draft nodes solidify — status dots turn green, badge changes to `ACTIVE`.

| Step | System Action |
| :--- | :--- |
| Blueprint → Manifests | `ConvertBlueprintToManifests()` — generates team IDs, aggregates topics |
| Manifest → Team | `NewTeam(manifest)` — creates team with NATS subscriptions |
| Team → Agents | For each member: spawn `Agent` (cognitive) or `SensorAgent` (poll-based) |
| Agent subscriptions | Team trigger bus + personal `swarm.council.<id>.request` |
| Heartbeat | Protobuf `MsgEnvelope` to `swarm.global.heartbeat` every 5s |

## Team I/O Contracts

```
External Signal → Team.Inputs[] → handleTrigger()
    → internal bus: swarm.team.<id>.internal.trigger
    → All agents process (LLM inference or HTTP poll)
    → Agent responds: swarm.team.<id>.internal.respond
    → handleResponse() → Team.Deliveries[] (external output)
```

Cross-team wiring: connect one team's deliveries to another team's inputs.

## Agent Execution Modes

| Mode | Trigger | Processing | Output | Trust |
| :--- | :--- | :--- | :--- | :--- |
| **Cognitive** | NATS message (push) | LLM `InferWithContract()` + ReAct tool loop (configurable max iterations) | Response text → team respond bus | Variable (0.5–1.0) |
| **Sensor** | Timer (poll interval, default 60s) | HTTP fetch or heartbeat-only | `CTSEnvelope` → agent output topics | 1.0 (fully trusted) |

**ReAct Tool Loop:**
1. LLM receives trigger + system prompt (with runtime context + tools block).
2. If response contains `{"tool_call": {...}}`: resolve via `CompositeToolExecutor`, execute, re-infer.
3. Repeat up to `MaxIterations` (default 3, Admin gets 5).
4. Final response published to team respond bus.

## Global Broadcast

- **Toggle mode:** Click the Megaphone icon → header switches to `BROADCAST`.
- **Prefix shortcut:** Type `/all <message>` in Admin mode.
- **API:** `POST /api/v1/swarm/broadcast` with `{"content": "...", "source": "mission-control"}`.

Broadcasts to each team's `swarm.team.<id>.internal.trigger` topic.

## Governance & Artifacts

1. Artifact wrapped in `CTSEnvelope` with `TrustScore`.
2. If `TrustScore < AutoExecuteThreshold`: Overseer halts, emits `governance_halt` via SSE.
3. **DeliverablesTray** appears with pending artifacts.
4. User reviews in **GovernanceModal** → Approve & Dispatch or Reject & Rework.

## Symbiotic Seed (Quick Start)

For testing without LLM negotiation:

```bash
curl -X POST http://localhost:8081/api/v1/intent/seed/symbiotic \
  -H 'Content-Type: application/json'
```

Creates a pre-built Gmail + Weather sensor mission with `SensorAgent` instances.

## Bus Topology

| Level | Pattern | Purpose |
| :--- | :--- | :--- |
| Global | `swarm.global.*` | Public, guarded, high-level (user input, broadcast, heartbeat) |
| Team internal | `swarm.team.<id>.internal.*` | Private team chatter (trigger, respond) |
| Team signal | `swarm.team.<id>.signal.*` | Public team outputs |
| Council | `swarm.council.<id>.request` | Personal request-reply (direct addressing) |
| Audit | `swarm.audit.trace` | Enforced logging of all traffic |
