# V7 MCP Baseline Operating Profile
## Default, Configured, and Immediately Usable Capability Set

**Version:** V7.0
**Authority:** Principal Architect
**Purpose:** Enable meaningful work immediately after installation — beyond chat — with safe, sandboxed, preconfigured MCP capabilities.

---

## 0. Design Principles

1. Local-first — no remote dependency required
2. Sandbox-first — workspace boundary enforced
3. No arbitrary command execution
4. All tools emit MissionEventEnvelope events
5. All tools are scope-restricted by default
6. All mutations require proposal + confirm

---

## 1. Default MCP Servers (Preinstalled + Enabled)

### 1.1 Filesystem MCP (Sandboxed)

**Purpose:** Enable artifact creation and structured file output.

**Root Path:** `/workspace` — no absolute path access, no traversal beyond workspace.

**Allowed Operations (v1):**
- `read_file` (workspace only)
- `write_file` (workspace only)
- `list_dir` (workspace only)
- `create_dir` (workspace only)
- `append_file` (workspace only)

**Disallowed:**
- chmod, rm -rf outside workspace, symlink creation, absolute paths, shell execution

**Auto-Emit Events:**
- `tool.invoked` — on every operation
- `tool.completed` — on success
- `artifact.created` — on write operations
- `tool.failed` — on sandbox violation

### 1.2 Memory MCP (Semantic Store)

**Purpose:** Enable recall and knowledge persistence.

**Backed by:** Postgres + pgvector

**Default Behavior:**
- All stored memory scoped to `tenant_id`
- Mission `run_id` linked to `memory.stored` events
- Recall emits `memory.recalled` event

**Operations:**
- `store_fact` — persist a knowledge fragment
- `store_document` — persist a larger document
- `semantic_search` — cosine distance vector search
- `retrieve_by_id` — exact lookup

### 1.3 Artifact Renderer MCP

**Purpose:** Render structured outputs inline in the chat/timeline.

**Supported formats:**
- Markdown
- JSON (collapsible)
- HTML-safe fragments
- Tables (GFM)
- Image references (URL or base64)

**Events:**
- `artifact.created` — on render request
- `artifact.rendered` — on successful display

### 1.4 Research MCP (Safe Fetch + Extraction)

**Purpose:** Allow controlled web fetch for research tasks.

**Restrictions:**
- HTTP GET only
- Domain allowlist configurable (default: open)
- Max response size limit (configurable, default 1MB)
- No arbitrary script execution
- No POST/PUT/DELETE

**Events:**
- `tool.invoked` — on fetch
- `artifact.created` — if report generated from fetched content

---

## 2. Default Provider Configuration

### 2.1 Local Ollama (Enabled)
- Default inference engine
- All roles allowed
- No data boundary warning
- Endpoint: `http://127.0.0.1:11434` (configurable via `OLLAMA_HOST`)

### 2.2 Remote Providers (Disabled by Default)
- Claude, OpenAI, Gemini — must be manually enabled
- Confirmation modal required on first enable
- UI displays REMOTE boundary indicator on every message
- All remote usage logged with `provider_id` in MissionEventEnvelope

---

## 3. Default Trigger Capabilities

Trigger system ships functional with:

- **Event-based triggers** — fire on specific event types
- **Completion triggers** — fire on `mission.completed`
- **Cooldown** — minimum interval between firings
- **Recursion guard** — max depth limit (default 5, hard ceiling 10)
- **Concurrency guard** — max active runs per target (default 1)

**Default mode:** `propose` — all triggers generate proposals, never auto-execute without explicit policy.

---

## 4. Scheduling Engine

Enabled by default as a core capability (not MCP — user-visible in Automations panel).

**User operations:**
- Schedule mission (cron expression)
- Set interval
- Pause / Resume
- View scheduled runs

**All scheduled executions emit:**
- `mission.started`
- `mission.completed` / `mission.failed`
- Full event timeline per run

**Safety:**
- `max_active_runs` enforced (default 1)
- Scheduler suspends when NATS offline
- Minimum interval: 30 seconds

---

## 5. Default User-Capable Workflows

Immediately after install, a user must be able to:

### Workflow A: Create a Document Project
**User:** "Generate a product spec and save it to workspace/spec.md."

**Flow:**
1. Proposal generated (mutation detected: file write)
2. User confirms
3. Filesystem MCP writes file
4. `artifact.created` event emitted
5. Timeline shows: `tool.invoked` → `artifact.created` → `mission.completed`

### Workflow B: Research + Generate Variant Media
**User:** "Research AI trends and generate three LinkedIn posts."

**Flow:**
1. Research MCP fetches sources
2. Report artifact created
3. Variant generation produces 3 text artifacts
4. Trigger rule may auto-propose scheduling for recurring research

### Workflow C: Schedule Recurring Summary
**User:** "Every morning summarize my workspace reports."

**Flow:**
1. Proposal generated (schedule creation)
2. User confirms
3. Scheduler registers cron job
4. `next_run_at` set
5. Each run produces its own timeline

### Workflow D: Mission Chaining via Trigger
**User:** Creates trigger: "When research completes, generate a summary report."

**Flow:**
1. Research mission completes → `mission.completed` event
2. Trigger evaluates → matches
3. Summary mission proposed (default propose mode)
4. User confirms → summary run created with `parent_run_id`
5. Causal chain visible in ViewChain UI

---

## 6. Default Workspace Structure

```
/workspace
  /projects    — user project files
  /research    — fetched research outputs
  /artifacts   — generated artifacts (charts, reports, media)
  /reports     — structured reports
  /exports     — exported data files
```

Created automatically on first server boot. Filesystem MCP enforces boundary — `validateToolPath()` prevents escape via `..`, symlinks, or absolute paths.

---

## 7. Tool Risk Classification & Confirm Requirements

| Risk Level | Tools | Confirm Required |
| :--- | :--- | :--- |
| **Low** | `read_file`, `list_dir`, `semantic_search`, `retrieve_by_id`, `render_artifact` | No |
| **Medium** | `write_file`, `create_dir`, `append_file`, `store_fact`, `store_document`, scheduling | Yes |
| **High** | Remote provider enable, trigger rule creation, MCP server install, domain allowlist change | Always |

---

## 8. Event Emission Requirements

**Every MCP tool must emit MissionEventEnvelope events:**

| Event | When |
| :--- | :--- |
| `tool.invoked` | Before execution begins |
| `tool.completed` | After successful execution |
| `tool.failed` | On error or sandbox violation |
| `artifact.created` | When a file/output is produced |
| `memory.stored` | When memory is persisted |
| `memory.recalled` | When memory is retrieved |

Events are persisted BEFORE CTS publish. If NATS offline, events still persist (degraded mode).

---

## 9. Implementation Requirements Per Team

### Team A (Event Spine)
- Wire Filesystem MCP to event spine (`tool.invoked`, `tool.completed`, `artifact.created`)
- Wire Memory MCP to emit `memory.stored`, `memory.recalled` with `run_id`
- Link artifact pipeline to `run_id` in MissionEventEnvelope
- All tools log `audit_event_id`

### Team B (Trigger Engine)
- Enable trigger rule creation via API
- Default `mode = propose` enforced
- Trigger evaluation emits `trigger.rule.evaluated` and `trigger.fired`

### Team C (Scheduler)
- Enable scheduler with guardrails (max_active_runs, NATS-offline suspension)
- Scheduled runs produce `mission_run` + full event timeline

### Team D (Navigation)
- Add "Workspace" file explorer under Resources panel
- Display artifact list per run in Run Timeline
- Allow user to open/preview workspace files inline

### Team E (Timeline UI)
- Render `tool.invoked` / `tool.completed` / `artifact.created` events in RunTimeline
- Artifact events show file path, size, content preview

---

## 10. What Is Explicitly NOT Included in V7

- Shell MCP (command execution)
- Docker/Kubernetes control
- OS-level commands
- Git push/pull automation
- External secrets injection
- Arbitrary MCP install without library approval

These belong in V8+ under enterprise control profiles.

---

## 11. Acceptance Criteria

V7 MCP Baseline is considered operational when:

- [ ] Install → Login → User can immediately create artifacts (no setup beyond starting Ollama)
- [ ] No manual MCP configuration required
- [ ] Files written only to `/workspace` (sandbox enforced)
- [ ] All tool operations appear in Run Timeline
- [ ] Scheduled missions run safely with guardrails
- [ ] Triggered missions visible and inspectable via Causal Chain
- [ ] No blank panels — degraded messaging when services unavailable
- [ ] Remote providers remain disabled until explicitly enabled with confirmation
- [ ] All MCP tools emit proper MissionEventEnvelope events

---

End of MCP Baseline Operating Profile.
