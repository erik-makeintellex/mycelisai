# Meta-Agent & Blueprint Planning

> How Mycelis translates natural language intent into a structured mission — agents, teams, tools, and I/O contracts.

---

## What Is the Meta-Agent?

The **Architect** council member acts as the system's meta-agent — a specialist whose job is not to execute tasks directly, but to **design the execution plan** for other agents and teams.

When Soma receives a request that requires spinning up a crew (multiple agents, file I/O, API calls, chained steps), it consults the Architect via `consult_council`. The Architect then calls `generate_blueprint` to produce a structured mission plan.

```
You → Soma → consult_council("architect")
               ↓
          Architect.generate_blueprint(intent)
               ↓
          MissionBlueprint (teams, agents, tools, I/O contracts)
               ↓
          Soma synthesizes + proposes
               ↓
          ProposedActionBlock in chat
```

---

## The Blueprint Structure

A blueprint is a structured JSON document describing a complete mission:

```json
{
  "mission_id": "m-abc1234",
  "goal": "Build a Python CSV parser with tests",
  "teams": [
    {
      "name": "dev-team",
      "role": "implementation",
      "agents": [
        {
          "name": "coder-1",
          "role": "coder",
          "model": "qwen2.5-coder:7b",
          "tools": ["write_file", "read_file", "store_artifact"],
          "system_prompt": "You are a senior Python developer...",
          "inputs": ["swarm.team.dev-team.internal.trigger"],
          "outputs": ["swarm.team.dev-team.internal.respond"]
        }
      ]
    },
    {
      "name": "review-team",
      "role": "verification",
      "agents": [
        {
          "name": "sentry-1",
          "role": "sentry",
          "tools": ["read_file", "search_memory"],
          "inputs": ["swarm.team.review-team.internal.trigger"],
          "outputs": ["swarm.team.review-team.internal.respond"]
        }
      ]
    }
  ],
  "constraints": {
    "max_iterations": 5,
    "trust_threshold": 0.7,
    "require_approval": ["write_file"]
  }
}
```

### Blueprint Fields

| Field | Description |
|-------|-------------|
| `mission_id` | Unique identifier (auto-generated) |
| `goal` | Plain-language description of the mission objective |
| `teams[]` | One or more teams to execute the mission |
| `teams[].role` | Team's function: implementation, research, review, monitoring, etc. |
| `teams[].agents[]` | Agent definitions — role, model, tools, NATS I/O |
| `agents[].tools` | Which tools this agent can invoke in its ReAct loop |
| `agents[].inputs` | NATS topics this agent listens on for triggers |
| `agents[].outputs` | NATS topics this agent publishes results to |
| `constraints` | Mission-level safety constraints |

---

## Triggering Blueprint Generation

### Via Soma Chat (Natural Language)

The most common path — just describe what you want:

```
"Build me a data pipeline that fetches weather data,
 cleans it, and stores a daily report in /workspace/reports"
```

Soma detects this requires a multi-agent setup and delegates to the Architect. The Architect calls `generate_blueprint` with your intent, returns a structured plan, and Soma proposes it as a `ProposedActionBlock`.

**When blueprint generation is triggered:**
- Multi-step tasks (more than 2 sequential operations)
- Tasks requiring specialized agents (coder + reviewer, researcher + writer)
- Any task mentioning "team", "crew", "pipeline", or "build"
- Tasks with explicit verification requirements

### Via LaunchCrew Modal

For structured intent before starting a conversation:

1. Click **Launch Crew** in the Workspace header
2. Describe the mission goal in the Step 1 text field — be specific about:
   - What needs to be produced
   - What tools or data sources are involved
   - Any constraints (language, file paths, output format)
3. Soma routes directly to blueprint generation — no pre-conversation needed

### Via API (Direct)

Call the negotiate endpoint directly:

```http
POST /api/v1/intent/negotiate
Authorization: Bearer mycelis-dev-key-change-in-prod
Content-Type: application/json

{
  "intent": "Build a Python CSV parser with quoted field support and unit tests",
  "context": {}
}
```

Response is a full `MissionBlueprint` JSON document.

---

## Reviewing a Blueprint Before Launch

The `ProposedActionBlock` in chat shows a summary. For the full blueprint JSON:

1. Click the **OrchestrationInspector** (expand chevron below the proposal)
2. The full blueprint is visible as structured JSON
3. You can see exactly which agents will be spawned, which tools they have, and what NATS topics they connect to

**The Wiring canvas** (`/automations?tab=wiring`, Advanced Mode) renders the blueprint as a visual DAG — nodes for each agent, edges for data flow between teams. You can modify the graph manually before activating.

---

## Agent Catalogue & Reuse

The Architect draws from the **Agent Catalogue** when generating blueprints — reusing proven templates rather than inventing agents from scratch:

- `GET /api/v1/catalogue/agents` — list all available templates
- Templates define: role, default system prompt, tool bindings, verification strategy
- The Architect matches catalogue templates to roles in the blueprint goal
- If no good match exists, a new agent definition is generated inline

To improve blueprint quality for recurring task types, add good agent templates to the catalogue (`/resources?tab=catalogue`). The Architect will use them automatically.

---

## Blueprint → Activation Pipeline

Once you confirm a blueprint proposal:

```
Confirm & Execute
    ↓
POST /api/v1/intent/confirm-action { confirm_token }
    ↓
HandleConfirmAction → ActivateBlueprint
    ↓
ConvertBlueprintToManifests() — blueprint → TeamManifest[]
    ↓
NewTeam(manifest) — NATS subscriptions registered
    ↓
Spawn Agent goroutines (cognitive or sensor mode)
    ↓
mission_run created → run_id returned
    ↓
mission.started event emitted → Run Timeline visible
```

After activation, teams appear in **Automations → Teams** and agent heartbeats appear in the **System → NATS** waterfall (Advanced Mode).

---

## Tips for Better Blueprints

- **Name the output explicitly**: "produce a file at `/workspace/reports/summary.md`" generates better agent tool bindings than "write a report"
- **Specify the stack**: "Python 3.12, type hints, pytest" constrains the Coder's output style
- **Mention verification**: "the code must pass linting" triggers a Sentry reviewer agent
- **Reference past work**: "similar to the CSV parser from last week" — the Architect searches memory for prior blueprints and reuses their structure
- **Limit scope**: Large vague goals produce large vague blueprints. Break big missions into focused 2–3 agent teams
