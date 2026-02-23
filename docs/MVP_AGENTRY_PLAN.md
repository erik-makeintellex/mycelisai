# Mycelis MVP Agentry Plan
## Agent Partner -> Meta -> Council -> Plan -> Execute

> **Date:** 2026-02-22  
> **Owner:** Principal Architect  
> **Goal:** Get the full agentry chain working end-to-end, then grow features from that verified foundation.

---

## System Map (Current Reality)

```
User
 +-- Workspace Chat (frontend)
      +-- POST /api/v1/council/admin/chat
           +-- Go AuthMiddleware (API key)
                +-- HandleCouncilChat
                     +-- NATS: swarm.council.admin.request
                          +-- Soma (admin agent)
                               +-- Internal Tools (read_file, write_file, remember, recall...)
                               +-- consult_council -> NATS: swarm.council.{member}.request
                               |    +-- Architect: generate_blueprint, list_catalogue
                               |    +-- Coder: read_file, write_file, store_artifact
                               |    +-- Creative: generate_image, write_file
                               |    +-- Sentry: security review, read_signals
                               +-- research_for_blueprint -> list_missions + list_catalogue + get_system_status
                               +-- generate_blueprint -> MissionBlueprint JSON
                                    +-- POST /api/v1/intent/commit (user confirms)
                                         +-- Soma.ActivateBlueprint() -> spawn team -> agents execute
```

**Provider chain:**  
All agents -> `cognitive.Router` -> profile lookup -> `ollama` adapter -> `http://127.0.0.1:11434/v1` -> model inference

---
## Sprint 0 - Foundation Verified (Do Now)

**What:** Ensure all services start cleanly and the auth chain works.

### 0.1 Services (manual, two terminals)

```bash
# Terminal A - Core API (blocking)
inv core.build && inv core.run

# Terminal B - Interface (restart to pick up .env.local)
inv interface.stop && inv interface.dev
```

**Verify:**
```bash
inv lifecycle.status
# Expected: PostgreSQL UP, NATS UP, Core API UP, Frontend UP, Ollama UP
```

### 0.2 Auth Chain Verified

Auth middleware (`interface/middleware.ts`) injects `Authorization: Bearer` from `MYCELIS_API_KEY`.  
`interface/.env.local` provides the key to the Next.js process.  
`ops/interface.py dev()` now loads `.env` before starting.  
**Status: FIXED.**

### 0.3 Ollama Model Verification WARNING CRITICAL

`cognitive.yaml` specifies `model_id: "qwen2.5-coder:7b"`. This MUST match an installed model:

```bash
# Check installed models
ollama list

# If qwen2.5-coder:7b is missing, pull it:
ollama pull qwen2.5-coder:7b

# OR update cognitive.yaml to match what IS installed:
# model_id: "llama3.2"  (or whatever ollama list shows)
```

**File:** `core/config/cognitive.yaml` -- `providers.ollama.model_id`

### 0.4 MCP Bootstrap Verification

On first `inv core.run`, `BootstrapDefaults()` auto-installs `filesystem` and `fetch` MCP servers.  
Check core startup logs for:
```
[mcp] bootstrap: installing filesystem
[mcp] bootstrap: filesystem connected and tools discovered
[mcp] bootstrap: installing fetch
[mcp] bootstrap: fetch connected and tools discovered
```
**Requires:** `npx` available on PATH (Node.js installed).

### 0.5 First Chat Test

Open `http://localhost:3000` -> Workspace -> type:  
`"Hello Soma, who are you and what tools do you have?"`

Expected: Soma introduces itself and calls `list_available_tools`.

---

## Sprint 1 - Agentry Chain: Full Smoke Test (1 day)

**Goal:** Confirm every link in the chain works. Fix anything broken.

### 1.1 Soma -> Council Round-Trip

Test: `"Ask the Architect to outline what a research-to-report mission would look like."`

Expected:
1. Soma calls `consult_council member="council-architect"`
2. NATS: `swarm.council.council-architect.request`
3. Architect responds with design
4. Soma returns synthesis to user

**Check:** Core logs show two NATS round-trips.

### 1.2 Mission Blueprint -> Confirm -> Execute

Test: `"Create a mission to research AI trends and write a summary to workspace/trends.md"`

Expected flow:
1. Soma calls `research_for_blueprint` -> gathers context
2. Soma calls `generate_blueprint` -> returns `MissionBlueprint` JSON
3. Response arrives with `template_id: "chat-to-proposal"` + `confirm_token`
4. UI shows `ProposedActionBlock` with "Confirm & Launch" button
5. User clicks confirm -> `POST /api/v1/intent/confirm-action`
6. Mission activates -> agents execute -> writes to `workspace/trends.md`

**Check:** `workspace/trends.md` exists after execution.

### 1.3 Internal File Tools

Test: `"Read the file workspace/trends.md and summarize it."`

Expected: Soma calls `read_file path="trends.md"` -> returns content.

### 1.4 Memory Round-Trip

Test: `"Remember that I prefer concise bullet-point summaries."`  
Then: `"What are my formatting preferences?"`

Expected: Soma calls `remember`, then `recall` on second query.

---

## Sprint 2 - Provider Selection Experience (2 days)

**Goal:** Make it obvious and easy to switch between Ollama, vLLM, and cloud providers.

### 2.1 Cognitive YAML - Model Accuracy (immediate)

Update `core/config/cognitive.yaml` to reflect what's actually available:

```yaml
providers:
  ollama:
    type: "openai_compatible"
    endpoint: "http://127.0.0.1:11434/v1"
    model_id: "qwen2.5-coder:7b"   # MUST match `ollama list` output
    api_key: "ollama"
    location: "local"
    data_boundary: "local_only"
    enabled: true
```

**Action:** After running `ollama list`, update `model_id` to match installed model.

### 2.2 Provider Status in Workspace Header (frontend)

**File:** `interface/components/dashboard/MissionControl.tsx`

Add a live cognitive status pill in the header -- shows active model and provider:

```
[ Ollama . qwen2.5-coder:7b . LOCAL ]
```

Data: `GET /api/v1/cognitive/status` -> probe result -> show in TelemetryRow or header.

**New component:** `interface/components/dashboard/ProviderPill.tsx`

### 2.3 Remote Provider Enable Flow (Resources -> Brains)

Already built: `interface/components/settings/BrainsPage.tsx` with toggle + data boundary.

**Gaps to fill:**
- API key input field in `RemoteEnableModal.tsx` -- when enabling `production_claude` or `production_gpt4`, prompt for API key
- Store key in environment (write to `.env` file via Go handler) or session env var
- Show LEAVES ORG warning prominently before enable

**New backend endpoint:** `PUT /api/v1/cognitive/providers/{id}/key` -- sets `AuthKey` in-memory (not persisted to yaml for security).

### 2.4 Per-Profile Routing (Brains tab)

Show which council role uses which provider, allow reassignment:

```
Profile    Provider          Model
----------------------------------------------
admin      ollama            qwen2.5-coder:7b
architect  ollama            qwen2.5-coder:7b
coder      ollama            qwen2.5-coder:7b
creative   ollama            qwen2.5-coder:7b
sentry     ollama            qwen2.5-coder:7b
```

Dropdown per row: select provider. Calls `PUT /api/v1/cognitive/profiles`.  
Already implemented on backend. **Frontend: add profile assignment UI to BrainsPage.**

### 2.5 Multi-Provider Routing (Advanced)

When multiple providers are enabled, allow per-agent override:
- Architect uses `production_claude` (complex reasoning)
- Coder uses `ollama` (local, fast)
- Sentry uses `ollama` (data stays local)

**Config:** `cognitive.yaml profiles` + agent manifest `model` field override.

---

## Sprint 3 - V7 Event Spine (Team A) [3 days backend]

**Goal:** Every mission action is observable. Foundation for automation.

### 3.1 DB Migrations

```sql
-- 023_mission_runs.up.sql
CREATE TABLE mission_runs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    mission_id VARCHAR(255) NOT NULL,
    tenant_id VARCHAR(255) NOT NULL DEFAULT 'default',
    status VARCHAR(50) NOT NULL DEFAULT 'pending', -- pending|running|completed|failed
    run_depth INT NOT NULL DEFAULT 0,
    parent_run_id UUID REFERENCES mission_runs(id),
    trigger_rule_id UUID,
    started_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    completed_at TIMESTAMPTZ,
    metadata JSONB
);

-- 024_mission_events.up.sql
CREATE TABLE mission_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    run_id UUID NOT NULL REFERENCES mission_runs(id),
    tenant_id VARCHAR(255) NOT NULL DEFAULT 'default',
    event_type VARCHAR(100) NOT NULL,  -- tool.invoked, tool.completed, mission.started, etc.
    severity VARCHAR(20) NOT NULL DEFAULT 'info',
    source_agent VARCHAR(255),
    source_team VARCHAR(255),
    payload JSONB,
    audit_event_id UUID,
    emitted_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX ON mission_events(run_id, emitted_at);
```

### 3.2 New Files

| File | Purpose |
|------|---------|
| `core/pkg/protocol/events.go` | `MissionEventEnvelope`, `EventType` constants (17 types) |
| `core/internal/events/store.go` | `Emit()`, `EmitAndPublish()`, `GetRunTimeline()`, `GetChain()` |
| `core/internal/runs/manager.go` | `CreateRun()`, `UpdateRunStatus()`, `GetRun()`, `ListRunsForMission()` |
| `core/internal/server/runs.go` | HTTP handlers for run/events/chain endpoints |

### 3.3 Wire Into Existing Code

- `soma.go` -> `ActivateBlueprint()` creates `mission_run` before spawning team
- `agent.go` -> `processMessageStructured()` emits `tool.invoked`/`tool.completed`
- `mission.go` -> `commitAndActivate()` calls `runs.CreateRun()`
- `admin.go` -> register 2 new routes

### 3.4 Degraded Mode

If NATS offline: persist event to DB, log warning, skip CTS publish. No panic.

**Done when:** `GET /api/v1/runs/{id}/events` returns timeline. `inv core.test` passes.

---

## Sprint 4 - MCP Baseline: Workspace + Fetch (2 days)

**Goal:** Agents can create files in workspace, fetch web content, and user can browse results.

### 4.1 Filesystem MCP (Auto-Installed)

**Already implemented:** `BootstrapDefaults()` in `mcp/service.go` auto-installs `filesystem` on first boot.

**Gap:** Requires `npx` on PATH. If `npx` unavailable:
- Fallback to internal `read_file`/`write_file` tools (already working)
- Log clear error: `[mcp] bootstrap: filesystem failed -- is node/npx installed?`

**Enhancement:** Create workspace subdirectories on first boot:
```
workspace/
  projects/    <- user project files
  research/    <- fetched research outputs  
  artifacts/   <- generated artifacts
  reports/     <- structured reports
  exports/     <- exported data files
```

**File:** `core/internal/mcp/service.go` -> `BootstrapDefaults()` -> add `os.MkdirAll` for subdirs.

### 4.2 Fetch MCP (Auto-Installed)

**Already implemented:** `BootstrapDefaults()` installs `@modelcontextprotocol/server-fetch`.

**Gap:** Agents need `fetch` in their tool list to use it. Currently agents use internal tools only.

**Config fix:** Add `fetch_url` to relevant agent tool lists in `council.yaml`:
```yaml
# council-coder tools addition:
- fetch_url   # for documentation lookup
```

**Backend:** MCP tool calls flow through `CompositeToolExecutor` -> `mcp.ClientPool.CallTool()`.

### 4.3 Workspace File Browser UI

**File:** `interface/app/(app)/resources/page.tsx` -> Workspace tab (currently DegradedState)

New component: `interface/components/workspace/WorkspaceExplorer.tsx`

```
/workspace
+-- projects/
+-- research/
|   +-- ai-trends-2026.md  [View] [Download]
+-- artifacts/
+-- reports/
```

**Backend:** New endpoints:
```
GET /api/v1/workspace/files          -> list workspace tree
GET /api/v1/workspace/files/{path}   -> read file content (workspace-sandboxed)
```

**Priority:** After Sprint 3 (needs run_id for file-to-run association).

---

## Sprint 5 - Trigger Engine + Scheduler (Teams B + C) [3 days]

**Goal:** Missions can chain and recur automatically.

### 5.1 Trigger Engine (Team B)

After Sprint 3. Missions fire on `mission.completed`, `tool.completed`, etc.

```
trigger_rules:
  if event_type == "mission.completed"
  and source_mission == "research-mission"
  -> propose("summary-mission")
```

Guards: cooldown_ms, max_depth (10 hard ceiling), max_active_runs (1 default).

**UI:** Automations -> Trigger Rules tab (currently DegradedState).

### 5.2 Scheduler (Team C)

After Sprint 3. Cron-based mission activation via `scheduled_missions` table.

```
schedule:
  mission_id: "research-mission"
  cron: "0 9 * * 1-5"   # 9am weekdays
  max_active_runs: 1
```

**UI:** Automations -> Active Automations tab (currently DegradedState).

---

## Sprint 6 - Run Timeline UI (Team E) [2 days frontend]

**Goal:** User can see exactly what happened in every execution.

After Sprint 3 APIs are live.

**Components:**
- `RunTimeline.tsx` -- vertical event list, color-coded by type
- `EventCard.tsx` -- expandable detail: tool args, output, timing
- `ViewChain.tsx` -- parent -> event -> trigger -> child run tree
- Route `/runs/{id}` and `/runs/{id}/chain`

**Zustand additions:** `activeRunId`, `runTimeline`, `runChain`, `fetchRunTimeline`, `fetchRunChain`

---

## Provider Selection Priority Matrix

| Provider | Status | When to Use | Data Boundary |
|----------|--------|------------|---------------|
| Ollama (local) | ENABLED | Default -- all agents | LOCAL ONLY |
| vLLM (self-hosted) | enabled | High-throughput workloads | LOCAL ONLY |
| LM Studio | enabled | Desktop dev setup | LOCAL ONLY |
| OpenAI GPT-4 | disabled | Complex reasoning, requires API key | LEAVES ORG |
| Anthropic Claude | disabled | Long context, nuanced tasks, API key required | LEAVES ORG |
| Google Gemini | disabled | Multimodal tasks, API key required | LEAVES ORG |

**To enable a remote provider:**
1. Resources -> Brains tab -> toggle provider
2. Enter API key in modal
3. System stores key in memory (session only) or `.env`
4. Select which profiles/agents use it

**API Key Storage Decision (to implement):**
- Option A: In-memory only (session, more secure, lost on restart) 
- Option B: Write to `.env` file via Go handler (persists, needs file write permission)
- Option C: Store encrypted in DB (best, requires encryption key)

**Recommendation:** Option C for production. Option A for MVP (least surprise, most secure).

---

## Critical Path to MVP Release

```
Sprint 0: Services verified, auth working, Ollama model confirmed
    |
Sprint 1: Full agentry chain smoke-tested (Soma <-> Council <-> Tools)  
    |
Sprint 2: Provider selection visible and switchable
    |
Sprint 3: Event Spine (observable execution, audit trail)
    |
Sprint 4: Workspace explorer + MCP filesystem confirmed working
    |
Sprint 5: Triggers + Scheduler (automation)
    |
Sprint 6: Run Timeline UI (visual history)
    |
MVP RELEASE: Full agent partner experience
```

**MVP Release Definition (minimum):**
- [x] User can chat with Soma persistently (memory across sessions)
- [x] Soma can consult all 4 council members
- [x] Soma generates mission blueprints with confirm flow
- [x] Missions execute and produce workspace artifacts
- [x] User can switch between providers (local <-> cloud)
- [ ] Every execution is observable (Sprint 3)
- [ ] Workspace files browsable (Sprint 4)
- [ ] Automations possible (Sprint 5)
- [ ] Execution history visual (Sprint 6)

---

## Immediate Actions (ordered)

1. **Run `ollama list`** -- verify model name matches `cognitive.yaml`
2. **Update `cognitive.yaml` model_id** if needed
3. **Terminal A:** `inv core.build && inv core.run`
4. **Terminal B:** `inv interface.stop && inv interface.dev`
5. **Test chat** -- confirm Soma responds
6. **Check core logs** -- verify MCP bootstrap (filesystem + fetch)
7. **Test mission flow** -- research AI trends and write to workspace/research/trends.md
8. **Confirm workspace file exists** after execution
9. **Begin Sprint 3 coding** -- Event Spine migrations + store.go

---

End of MVP Agentry Plan.