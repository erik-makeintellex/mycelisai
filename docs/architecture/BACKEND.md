# Mycelis Cortex — Backend Specification

> **Load this doc when:** Working on Go code, APIs, database, NATS, or agent orchestration.
>
> **Related:** [Overview](OVERVIEW.md) | [Frontend](FRONTEND.md) | [Operations](OPERATIONS.md)

---

## I. Package Structure

### Entry Points (`cmd/`)

| Binary | File | Purpose |
|--------|------|---------|
| **server** | `cmd/server/main.go` | Main HTTP + NATS orchestration server |
| **probe** | `cmd/probe/main.go` | Health check utility |
| **signal_gen** | `cmd/signal_gen/main.go` | Signal generation tool |
| **smoke** | `cmd/smoke/main.go` | Smoke test suite |

### Public API (`pkg/`)

| Package | Files | Exports |
|---------|-------|---------|
| `protocol` | types.go, envelopes.go, manifest.go, blueprint.go, topics.go | CTSEnvelope, AgentManifest, MissionBlueprint, SignalType, 30+ topic constants |
| `pb` | swarm.pb.go, envelope.pb.go | MsgEnvelope (Protobuf), SCIP protocol |

### Private Implementation (`internal/` — 15 packages)

| Package | Files | Responsibility |
|---------|-------|---------------|
| **server** | admin.go, cognitive.go, mission.go, telemetry.go, governance.go, memory_search.go, proposals.go, artifacts.go, identity.go, registry.go, catalogue.go, mcp.go, provision.go, memory.go | HTTP handlers (50+ endpoints), `AdminServer` struct, `RegisterRoutes()` |
| **swarm** | soma.go, axon.go, agent.go, team.go, sensor_agent.go, internal_tools.go, activation.go, converter.go, seeds.go, tool_executor.go | Agent orchestration, ReAct loop, tool dispatch |
| **cognitive** | router.go, architect.go, openai.go, anthropic.go, google.go, discovery.go, types.go | LLM routing, provider adapters, token telemetry, embedding |
| **governance** | guard.go, policy.go | Policy engine, rule evaluation, approval queue |
| **memory** | service.go, archivist.go | Event log persistence, SitRep generation, LLM compression |
| **overseer** | engine.go | Zero-trust DAG orchestration, governance valve |
| **mcp** | service.go, pool.go, executor.go, library.go | MCP server registry, client lifecycle, tool invocation |
| **artifacts** | service.go | Agent output persistence |
| **catalogue** | service.go | Agent template CRUD |
| **bootstrap** | — | Node discovery |
| **router** | — | Input intent dispatcher |
| **signal** | stream.go | SSE streaming handler |
| **state** | — | State registry |
| **identity** | — | Auth stubs |
| **transport** | — | NATS wrapper |

### Go Dependencies (Direct)

| Package | Version | Purpose |
|---------|---------|---------|
| `github.com/jackc/pgx/v5` | 5.8.0 | PostgreSQL driver |
| `github.com/nats-io/nats.go` | 1.48.0 | NATS client |
| `github.com/nats-io/nats-server/v2` | 2.12.4 | Embedded NATS (testing) |
| `github.com/sashabaranov/go-openai` | 1.41.2 | OpenAI-compatible SDK |
| `github.com/mark3labs/mcp-go` | 0.43.2 | Model Context Protocol |
| `github.com/google/uuid` | 1.6.0 | UUID generation |
| `github.com/xeipuuv/gojsonschema` | 1.2.0 | JSON Schema validation |
| `github.com/DATA-DOG/go-sqlmock` | 1.5.2 | SQL mocking (tests) |
| `google.golang.org/protobuf` | 1.36.1 | Protocol Buffers |
| `gopkg.in/yaml.v3` | 3.0.1 | YAML config parsing |

---

## II. Swarm Orchestration

### Soma (Executive Cell) — `swarm/soma.go`

The user's proxy and team manager.

**Responsibilities:**
- Receives directives via `swarm.global.input.user`
- Spawns/manages Teams from YAML registry
- Wires composite tool executor (internal + MCP) to teams
- Delegates execution to Axon for monitoring
- **Exports:** `ListTeams()`, `ListAgents()`, `ActivateBlueprint()`, `DeactivateMission()`, `HandleCreateTeam()`

**Standing Teams (auto-spawned from YAML at startup):**

| Team ID | Name | Type | Members | Tools |
|---------|------|------|---------|-------|
| `admin-core` | Admin | action | 1 (admin agent) | 17 built-in tools, 5 ReAct iterations max |
| `council-core` | Council | cognitive | 4 (architect, coder, creative, sentry) | Per-member tool sets |
| `genesis-core` | Genesis | action | 2 (architect, commander) | Bootstrap tools |

### Axon (Messenger) — `swarm/axon.go`

Signal router and UI bridge.

- Subscribes to `swarm.team.*.internal.>` (all team chatter)
- Classifies signals: trigger, response, command, status
- Enriches with timestamps
- Broadcasts to SSE StreamHandler for real-time UI

### Agent (LLM Reasoning Node) — `swarm/agent.go`

**ReAct Loop:**
1. Receive trigger message or NATS request-reply
2. Build context: system_prompt + `InternalToolRegistry.BuildContext()` + thought_history + tool_descriptions
3. Call `cognitive.Router.Infer()` (LLM inference)
4. Parse JSON: `tool_call` or `final_response`
5. If `tool_call`: resolve (internal → MCP → error), execute, append to history, continue
6. If `final_response`: publish CTSEnvelope, break
7. Max iterations guard (default 5 for admin)

**Runtime Context Injection** (`BuildContext()` injects before every message):
1. Active teams roster (names, member counts)
2. Agent identity + NATS topology (trigger, respond, direct-address topics)
3. Cognitive engine (active providers, endpoints, models, profiles)
4. Installed MCP servers (names, transport, status, discovered tools)
5. Interaction protocol (pre-response checklist: recall, consult, delegate, MCP tools)

**Heartbeat:** `swarm.global.heartbeat` every 5 seconds
**Personal RPC:** Listens on `swarm.council.{agent_id}.request`

### SensorAgent (Poll-Based) — `swarm/sensor_agent.go`

- Configuration: endpoint URL, interval (default 60s), headers
- HTTP client timeout: 10s
- Publishes CTSEnvelopes with `TrustScore = 1.0` (sensor data is fact)
- Zero cognitive overhead — no LLM invocation

### Team — `swarm/team.go`

Agent container and task dispatcher.

- Spawns Agents (cognitive) or SensorAgents (poll-based) based on `Team.sensorConfigs`
- Subscribes to team input topics
- Routes triggers to agents via `swarm.team.{id}.internal.trigger`
- Publishes agent outputs to delivery topics
- May spawn nested sub-teams (hierarchy via `parent_id`)

### Internal Tool Registry — `swarm/internal_tools.go`

17 built-in tools:

| Tool | Purpose | Primary Users |
|------|---------|---------------|
| `consult_council` | Query a specific council member | Admin |
| `delegate_task` | Publish task to team trigger topic | Admin |
| `search_memory` | Vector semantic search on SitReps | Admin, Council |
| `list_teams` | Enumerate teams from Soma | Admin |
| `list_missions` | Query missions from DB | Admin |
| `get_system_status` | Aggregate runtime metrics | Admin |
| `list_available_tools` | Enumerate all tools | Council (Sentry) |
| `list_catalogue` | Browse agent templates | Council (Architect) |
| `generate_image` | Stable Diffusion via media endpoint | Council (Creative) |
| `remember` | Persistent fact storage | Any agent |
| `recall` | Retrieve stored facts | Any agent |
| `publish_signal` | Broadcast to NATS topic | Council, agents |
| `read_signals` | Subscribe to NATS topic | Council (Sentry) |
| `read_file` | Filesystem read access | Council |
| `write_file` | Filesystem write access | Council (Coder, Architect) |
| `store_artifact` | Persist agent output | Any agent |
| `research_for_blueprint` | Gather context for blueprint design | Admin |

### Composite Tool Executor — `swarm/tool_executor.go`

Unified routing:
1. Check internal tool registry → execute directly
2. If not found → check MCP pool → `mcp.Service.Call(server_id, tool_name, args)`
3. If not found → return error

### Blueprint Activation — `swarm/activation.go`, `converter.go`, `seeds.go`

- `ConvertBlueprintToManifests()` — mission-scoped IDs
- `ActivateBlueprint()` — DB persist + Soma spawn
- `seeds.go` — Symbiotic Seed (Gmail+Weather, no LLM required)

---

## III. Cognitive Layer

### Router — `cognitive/router.go`

**Config Priority:** YAML base → DB overlay → Env overrides

**Provider Types:**
| Provider ID | Type | Default Endpoint |
|-------------|------|-----------------|
| `ollama` | openai_compatible | `http://192.168.50.156:11434/v1` (LAN) |
| `vllm` | openai_compatible | `http://127.0.0.1:8000/v1` |
| `lmstudio` | openai_compatible | `http://127.0.0.1:1234/v1` |
| `production_gpt4` | openai | OpenAI API |
| `production_claude` | anthropic | Anthropic API |
| `production_gemini` | google | Gemini API |

**Profile → Provider Routing (all default → ollama):**
admin, architect, coder, creative, sentry, chat

**Token Telemetry:**
- `RecordTokens(n)` — cumulative counter + sliding window
- `TokenRate()` — tokens/second over 60s window
- Reported via `GET /api/v1/telemetry/compute`

**Media:** `http://127.0.0.1:8001/v1` (Stable Diffusion Diffusers, OpenAI-compatible)

### Adapters
| File | Provider | SDK |
|------|----------|-----|
| `openai.go` | OpenAI (GPT-4) | `go-openai` |
| `anthropic.go` | Anthropic (Claude) | Custom HTTP |
| `google.go` | Google (Gemini) | Custom HTTP |
| (shared) | OpenAI-compatible (Ollama, vLLM, LM Studio) | `go-openai` with custom base URL |

### Discovery — `cognitive/discovery.go`

- Probes all providers at startup with test inference
- Retries with exponential backoff
- Marks unavailable → WARN log, continues
- Dynamic probing: `GET /api/v1/cognitive/status`

### Architect — `cognitive/architect.go`

Meta-Architect: Intent (natural language) → MissionBlueprint (JSON) via LLM structured output.

---

## IV. Execution Pipelines

### Pipeline 1: Intent → Blueprint → Activation

```
User types intent in ArchitectChat
  │
  ▼
POST /api/v1/intent/negotiate
  │ Meta-Architect generates MissionBlueprint via LLM
  │ Frontend: blueprintToGraph() → ghost-draft nodes (50% opacity, dashed cyan)
  ▼
User reviews DAG, edits via WiringAgentEditor
  │
  ▼
POST /api/v1/intent/commit → commitAndActivate()
  │ 1. BEGIN DB transaction
  │ 2. INSERT missions (status='active')
  │ 3. INSERT teams (per BlueprintTeam)
  │ 4. INSERT service_manifests (per AgentManifest)
  │ 5. COMMIT
  ▼
Soma.ActivateBlueprint()
  │ 1. ConvertBlueprintToManifests() — mission-scoped IDs
  │ 2. Spawn Teams → spawn Agents/SensorAgents
  │ 3. Register NATS subscriptions, start heartbeat
  ▼
Ghost nodes solidify, status → ACTIVE
```

### Pipeline 2: Council Chat (Request-Reply)

```
POST /api/v1/council/{member}/chat  { "query": "..." }
  │ 1. Validate member via isCouncilMember() (checks Soma)
  │ 2. Build request envelope
  ▼
NATS: nc.Request("swarm.council.{member}.request", payload, 30s)
  │
  ▼
Agent ReAct loop → tools → final response
  │
  ▼
Wrap: ChatResponsePayload + TrustScore=0.5
  │ Includes: consultations[], tools_used[], source_node
  ▼
HTTP 200 JSON
```

### Pipeline 3: Agent ReAct Loop

```
Trigger on swarm.team.{team_id}.internal.trigger
  │
  ▼
LOOP (max_iterations):
  │ context = system_prompt + BuildContext() + thought_history + tools
  │ response = cognitive.Router.Infer(context)
  │
  ├─ final_response → publish CTSEnvelope, BREAK
  └─ tool_call → resolve → execute → append to history → CONTINUE
```

### Pipeline 4: Memory Archival & Compression

```
CTSEnvelope → swarm.team.{team_id}.telemetry
  │
  ├─ Memory.Service: INSERT log_entry, buffer event
  │
  ▼ Trigger: buffer ≥ 50 | artifact signal | 5min timer
  │
  Archivist.GenerateSitRep():
    1. Retrieve log_entries for team
    2. LLM compress → 3-sentence summary
    3. Embed via nomic-embed-text → 768-dim vector
    4. INSERT sitrep (content, context_vectors)
    5. DELETE processed log_entries
```

### Pipeline 5: Governance & Zero-Trust Actuation

```
Overseer receives CTSEnvelope
  │
  ├─ TrustScore >= 0.7: advance DAG (auto-execute)
  │
  └─ TrustScore < 0.7:
       GovernanceCallback → SSE → Zone D (GovernanceModal)
       Human: POST /governance/resolve/{id} → approve/deny
```

### Pipeline 6: SSE Real-Time Streaming

```
Frontend: EventSource → /api/v1/stream
Backend: signal.StreamHandler (verify http.Flusher, respect ctx.Done())
Axon routes signals → StreamHandler → data: {JSON}\n\n
Frontend: parse → Zustand streamLogs[] (cap 100)
  ├─ artifact → pendingArtifacts[]
  ├─ governance_halt → GovernanceModal
  └─ default → NatsWaterfall
```

---

## V. Data Contracts & Protocols

### 1. The CTS Envelope (Cortex Telemetry Standard)

```go
type CTSEnvelope struct {
    ID         string      `json:"id"`
    SourceNode string      `json:"source_node"`
    TeamID     string      `json:"team_id"`
    SignalType SignalType   `json:"signal_type"`
    TrustScore float64     `json:"trust_score"`    // 0.0–1.0
    Payload    interface{} `json:"payload"`
    CTSMeta    CTSMeta     `json:"cts_meta"`
}

type CTSMeta struct {
    SourceNodeID string    `json:"source_node_id"`
    Timestamp    time.Time `json:"timestamp"`
    TraceID      string    `json:"trace_id"`
}
```

**Signal Types:**
| Type | Purpose | Default TrustScore |
|------|---------|-------------------|
| `telemetry` | Agent work update | 0.5 |
| `task_complete` | DAG advancement | varies |
| `task_failed` | Error signal | varies |
| `error` | Exception | 0.0 |
| `heartbeat` | Agent alive | 1.0 |
| `governance_halt` | Low-trust pause | < threshold |
| `sensor_data` | SensorAgent ingress | 1.0 |
| `chat_response` | Council LLM output | 0.5 |

**Trust Defaults by Category:**
| Category | Score | Rationale |
|----------|-------|-----------|
| Sensory | 1.0 | Trusted fact |
| Cognitive | 0.5 | May hallucinate |
| Actuation | 0.8 | Moderate trust |
| Ledger | 1.0 | Immutable record |

### 2. The ChatResponsePayload

```go
type ChatResponsePayload struct {
    Response      string            `json:"response"`
    Consultations []ChatConsultation `json:"consultations,omitempty"`
    ToolsUsed     []string          `json:"tools_used,omitempty"`
    SourceNode    string            `json:"source_node"`
    TrustScore    float64           `json:"trust_score"`
}
```

### 3. The APIResponse Envelope

```go
type APIResponse struct {
    OK    bool        `json:"ok"`
    Data  interface{} `json:"data,omitempty"`
    Error string      `json:"error,omitempty"`
}
```

New endpoints use `respondAPIJSON()` / `respondAPIError()` helpers.

### 4. The MissionBlueprint

```go
type MissionBlueprint struct {
    MissionID    string                `json:"mission_id"`
    Intent       string                `json:"intent"`
    Teams        []BlueprintTeam       `json:"teams"`
    Constraints  []Constraint          `json:"constraints,omitempty"`
    Requirements []ResourceRequirement `json:"requirements,omitempty"`
}

type BlueprintTeam struct {
    Name   string          `json:"name"`
    Role   string          `json:"role"`
    Agents []AgentManifest `json:"agents"`
}

type ResourceRequirement struct {
    Type        string `json:"type"`   // mcp_server|api_key|env_var|credential
    Name        string `json:"name"`
    Description string `json:"description"`
    Required    bool   `json:"required"`
    Installed   bool   `json:"installed"`
}
```

### 5. The AgentManifest

```go
type AgentManifest struct {
    ID            string       `json:"id"`
    Role          string       `json:"role"`
    SystemPrompt  string       `json:"system_prompt"`
    Model         string       `json:"model"`
    Inputs        []string     `json:"inputs"`
    Outputs       []string     `json:"outputs"`
    Tools         []string     `json:"tools"`
    MaxIterations int          `json:"max_iterations"`
    Verification  Verification `json:"verification"`
}

type Verification struct {
    Strategy          string   `json:"strategy"`  // semantic|empirical
    Rubric            []string `json:"rubric"`
    ValidationCommand string   `json:"validation_command"`
}
```

### 6. The ProofEnvelope

```go
type ProofEnvelope struct {
    Artifact interface{} `json:"artifact"`
    Proof    Proof       `json:"proof"`
}

type Proof struct {
    Method      string `json:"method"`  // semantic|empirical|both
    Logs        string `json:"logs"`
    RubricScore string `json:"rubric_score"`
    Pass        bool   `json:"pass"`
}
```

### 7. The SitRep Schema (Archivist Output)

```json
{
  "contract_id": "archivist_v1_sitrep",
  "summary": "String (Max 3 sentences).",
  "key_events": [{ "signal": "string", "source": "string" }],
  "strategies_applied": ["string"]
}
```

### 8. API Graceful Degradation

| Component Missing | HTTP Status | Behavior |
|-------------------|-------------|----------|
| NATS | 503 | REST-only mode |
| LLM (all providers) | 502 | Inference fails gracefully |
| PostgreSQL | WARN log | In-memory only |
| MCP Pool | Skip | Internal tools only |
| SSE Handler | 503 | Stream unavailable |
| Governance Guard | WARN log | Allow all (security off) |
| Memory Archivist | WARN log | Archival skipped |

---

## VI. NATS Topic Architecture

### Global Control Plane

| Topic | Purpose | Frequency |
|-------|---------|-----------|
| `swarm.global.heartbeat` | Agent proof-of-life | 5s per agent |
| `swarm.global.announce` | System announcements | On events |
| `swarm.global.broadcast` | Mission Control → All Teams | On user action |
| `swarm.global.input.user` | User input ingress | On chat submit |
| `swarm.audit.trace` | Immutable audit log | Every action |

### Team Internal (per team)

| Topic Pattern | Purpose |
|---------------|---------|
| `swarm.team.{team_id}.internal.trigger` | Task dispatch to agents |
| `swarm.team.{team_id}.internal.response` | Agent task responses |
| `swarm.team.{team_id}.internal.command` | Control commands |
| `swarm.team.{team_id}.signal.status` | Status signals |
| `swarm.team.{team_id}.telemetry` | CTS envelopes (agent output + governance) |

### Wildcards

| Topic Pattern | Subscribers |
|---------------|-------------|
| `swarm.team.*.internal.>` | Axon (monitors all team chatter) |
| `swarm.team.*.telemetry` | Overseer (validates all CTS envelopes) |

### Council Request-Reply

| Topic Pattern | Purpose |
|---------------|---------|
| `swarm.council.{agent_id}.request` | Direct RPC to council member (30s timeout) |

### Sensor Data Ingress

| Topic Pattern | Purpose |
|---------------|---------|
| `swarm.data.email.>` | Email feeds |
| `swarm.data.weather.>` | Weather APIs |
| `swarm.data.mcp.>` | MCP tool outputs |
| `swarm.data.>` | Wildcard: all sensor data |

### Mission DAG

| Topic | Purpose |
|-------|---------|
| `swarm.mission.task` | Overseer task dispatch |

### Agent Output

| Topic Pattern | Purpose |
|---------------|---------|
| `swarm.agent.{agent_id}.output` | Individual agent stream |

**Rule:** All topics use constants from `pkg/protocol/topics.go` — never hardcode topic strings.

---

## VII. Database Schema

### 14+ Tables, 21 Migrations

#### Core Tables

**`users`** (Migration 002)
| Column | Type | Notes |
|--------|------|-------|
| id | UUID PK | |
| username | TEXT | |
| role | TEXT | admin\|user\|observer |
| settings | JSONB | |
| created_at | TIMESTAMPTZ | |

**`missions`** (Migrations 003, 010)
| Column | Type | Notes |
|--------|------|-------|
| id | UUID PK | |
| owner_id | UUID FK → users | |
| name | TEXT | |
| directive | TEXT | The intent |
| status | TEXT | draft\|active\|paused\|completed\|failed |
| activated_at | TIMESTAMPTZ | [010] When mission went active |
| created_at | TIMESTAMPTZ | |
| updated_at | TIMESTAMPTZ | |

**`teams`** (Migrations 002, 003, 007)
| Column | Type | Notes |
|--------|------|-------|
| id | UUID PK | |
| owner_id | UUID FK → users | |
| name | TEXT | |
| role | TEXT | |
| mission_id | UUID FK → missions ON DELETE CASCADE | |
| parent_id | UUID FK → teams ON DELETE SET NULL | Hierarchy |
| path | TEXT | Materialized path for tree traversal |
| type | TEXT | standing\|mission |
| created_at, updated_at | TIMESTAMPTZ | |

**`service_manifests`** (Migrations 002, 020)
| Column | Type | Notes |
|--------|------|-------|
| id | UUID PK | |
| team_id | UUID FK → teams ON DELETE CASCADE | [020] Added cascade |
| name | TEXT | |
| manifest | JSONB | AgentManifest config |
| status | TEXT | active\|paused\|failed |
| created_at, updated_at | TIMESTAMPTZ | |

#### Memory & Knowledge

**`log_entries`** (Migration 001) — id, team_id, event, content (JSONB), created_at

**`sitreps`** (Migration 008) — id, team_id FK, period_start, period_end, content (TEXT), context_vectors (vector(768)), created_at

**`working_memory`** (Migration 008) — id, team_id FK, context, created_at, expires_at (TTL)

**`agent_state`** (Migration 018) — agent_id (TEXT), key, value, mission_id FK (CASCADE), expires_at (TTL), PK(agent_id, key)

#### Output

**`artifacts`** (Migration 018) — id, mission_id FK, team_id FK, agent_id, trace_id, artifact_type (code\|document\|image\|audio\|data\|file\|chart), title, content_type (MIME), content, file_path, file_size_bytes, metadata (JSONB), trust_score, status (pending\|approved\|rejected\|archived), created_at

**`agent_catalogue`** (Migration 017) — id, name, role, system_prompt, model, tools (TEXT[]), inputs (TEXT[]), outputs (TEXT[]), verification_strategy, verification_rubric (TEXT[]), validation_command, created_at, updated_at

#### Infrastructure

**`nodes`** (Migrations 005, 011–015) — id (TEXT PK), type, status, last_seen, specs (JSONB)

**`cognitive_registry`** (Migration 006) — provider_id (TEXT PK), type, endpoint, model_id, api_key_env, status, last_checked

**`mcp_servers`** (Migration 016) — id, name, transport (stdio\|sse), command, args, env (JSONB), url, status, created_at

**`mcp_tools`** (Migration 016) — id, server_id FK → mcp_servers, name, description, input_schema (JSONB)

### Migration Index

| # | Purpose |
|---|---------|
| 001 | Event log (log_entries) |
| 002 | Core schema (users, teams, service_manifests) |
| 003 | Mission hierarchy + materialized path |
| 004 | Registry metadata |
| 005 | Hardware node discovery |
| 006 | LLM provider config (cognitive_registry) |
| 007 | Team fabric + root user seeding |
| 008 | Context engine (sitreps, working_memory, pgvector) |
| 009 | Local dev provider defaults |
| 010 | missions.status + missions.activated_at |
| 011–015 | Node/provider schema fixes |
| 016 | MCP server registry (mcp_servers, mcp_tools) |
| 017 | Agent catalogue (agent_catalogue) |
| 018 | Artifacts + agent state |
| 019 | Agent memory extensions |
| 020 | service_manifests → teams FK cascade fix |

---

## VIII. API Surface (50+ Endpoints)

### Identity & Users
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/v1/user/me` | Current user profile |
| PUT | `/api/v1/user/settings` | Update preferences |

### Chat & Council
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/v1/chat` | Admin agent chat (NATS request-reply only, no raw LLM fallback) |
| POST | `/api/v1/council/{member}/chat` | Direct council member query (CTS-enveloped response) |
| GET | `/api/v1/council/members` | List members (auto-discovered from Soma) |

### Mission Orchestration
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/v1/missions` | List missions |
| GET | `/api/v1/missions/{id}` | Mission detail |
| PUT | `/api/v1/missions/{id}/agents/{name}` | Update agent in active mission |
| DELETE | `/api/v1/missions/{id}/agents/{name}` | Remove agent from active mission |
| DELETE | `/api/v1/missions/{id}` | Delete mission (deactivate + DB) |
| POST | `/api/v1/intent/negotiate` | Generate blueprint via Meta-Architect |
| POST | `/api/v1/intent/commit` | Persist + activate mission |
| POST | `/api/v1/intent/seed/symbiotic` | Activate built-in sensors |

### Cognitive Engine
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/v1/cognitive/status` | Probe all providers dynamically |
| POST | `/api/v1/cognitive/infer` | Direct inference |
| GET | `/api/v1/cognitive/config` | Read config |
| PUT | `/api/v1/cognitive/profiles` | Update profile→provider routing |
| PUT | `/api/v1/cognitive/providers/{id}` | Update provider config |

### Telemetry & Trust
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/v1/stream` | SSE event stream |
| GET | `/api/v1/telemetry/compute` | Runtime metrics (goroutines, memory, tokens/sec, uptime) |
| GET | `/api/v1/trust/threshold` | Read AutoExecuteThreshold |
| PUT | `/api/v1/trust/threshold` | Update trust valve |

### Memory & RAG
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/v1/memory/stream` | Event log SSE |
| GET | `/api/v1/memory/search` | Vector semantic search (cosine) |
| GET | `/api/v1/memory/sitreps` | List SitReps |
| GET | `/api/v1/sensors` | List sensor subscriptions |

### Governance & Proposals
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/v1/governance/policy` | Read policy |
| PUT | `/api/v1/governance/policy` | Update policy |
| GET | `/api/v1/governance/pending` | List pending approvals |
| POST | `/api/v1/governance/resolve/{id}` | Approve/deny |
| GET | `/api/v1/proposals` | List team proposals |
| POST | `/api/v1/proposals/{id}/approve` | Approve proposal |
| POST | `/api/v1/proposals/{id}/reject` | Reject proposal |

### Teams
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/v1/teams` | List teams |
| GET | `/api/v1/teams/detail` | Aggregated team details |
| POST | `/api/v1/teams/{id}/connectors` | Install connector |
| GET | `/api/v1/teams/{id}/wiring` | Wiring diagram |

### MCP Management
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/v1/mcp/install` | Register server |
| GET | `/api/v1/mcp/servers` | List servers + tools |
| DELETE | `/api/v1/mcp/servers/{id}` | Uninstall |
| POST | `/api/v1/mcp/servers/{id}/tools/{tool}/call` | Invoke tool |
| GET | `/api/v1/mcp/tools` | List all tools |
| GET | `/api/v1/mcp/library` | Browse curated catalogue |
| POST | `/api/v1/mcp/library/install` | Install from library |

### Agent Catalogue
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/v1/catalogue/agents` | List templates |
| POST | `/api/v1/catalogue/agents` | Create |
| PUT | `/api/v1/catalogue/agents/{id}` | Update |
| DELETE | `/api/v1/catalogue/agents/{id}` | Delete |

### Artifacts
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/api/v1/artifacts` | List (filterable) |
| GET | `/api/v1/artifacts/{id}` | Detail |
| POST | `/api/v1/artifacts` | Store |
| PUT | `/api/v1/artifacts/{id}/status` | Update status |

### Provisioning & Registry
| Method | Path | Purpose |
|--------|------|---------|
| POST | `/api/v1/provision/draft` | Generate manifest |
| POST | `/api/v1/provision/deploy` | Deploy |
| GET | `/api/v1/registry/templates` | List templates |
| POST | `/api/v1/registry/templates` | Register |

### Health
| Method | Path | Purpose |
|--------|------|---------|
| GET | `/healthz` | Readiness probe |

---

## IX. Governance & Policy Engine

### Guard — `governance/guard.go`

**Rule Evaluation:**
1. Check intent pattern match (regex)
2. Check team/agent filter
3. Evaluate conditions (e.g., `amount > 50`)
4. Apply action: `ALLOW` | `DENY` | `REQUIRE_APPROVAL`

### Default Rules (`core/config/policy.yaml`)

| Rule | Intent Pattern | Action |
|------|---------------|--------|
| System Critical | `k8s.delete.*`, `system.shutdown` | REQUIRE_APPROVAL + DENY |
| Finance | `payment.create > $50` | REQUIRE_APPROVAL |
| IoT Safety | `firmware.update` | REQUIRE_APPROVAL |
| IoT Safety | `motor.set_speed > 8000 rpm` | DENY |
| Default | * | ALLOW |

**Approval Expiry:** 1 hour
**Storage:** In-memory (transient by design)

---

## X. MCP Integration

### Architecture (`internal/mcp/`)

| File | Purpose |
|------|---------|
| `service.go` | Registry + tool discovery |
| `pool.go` | Client lifecycle management |
| `executor.go` | Tool invocation |
| `library.go` | Curated server catalogue |

### Transport: stdio or SSE

### Curated Library (`core/config/mcp-library.yaml`)

| Category | Servers |
|----------|---------|
| Development | filesystem, github |
| Data/Search | postgres, sqlite, brave-search, fetch |
| Communication | slack |
| Media | stable-diffusion, flux, elevenlabs, replicate, comfyui, dall-e |
| Utilities | memory, puppeteer, sequential-thinking |

---

## XI. Startup & Shutdown

### Startup Sequence

```
1.  Load environment (ports, DB URL, NATS URL)
2.  Load config files (cognitive.yaml, policy.yaml)
3.  Connect PostgreSQL (retry 10x)
4.  Load Cognitive Router (provider discovery)
5.  Load Governance Guard (policy rules)
6.  Initialize Memory.Service
7.  Connect NATS (retry 10x, continue degraded if fail)
8.  Start Router (input dispatcher)
9.  Start Soma (team orchestration)
10. Start Axon (signal monitoring → SSE)
11. Start Overseer (DAG reconciliation)
12. Start Memory Subscription
13. Spawn standing teams from YAML (admin, council, genesis)
14. Spawn agents in each team
15. Start HTTP server (port 8080)
16. Register all routes
17. Block on SIGINT
```

### Graceful Shutdown

```
On SIGINT:
1. Drain NATS (in-flight messages)
2. Cancel all agent contexts
3. Stop HTTP server
4. Close database connection
5. Exit(0)
```
