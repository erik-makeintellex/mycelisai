# Run Timeline

> Every mission execution produces a run. The timeline shows exactly what happened, in order.

---

## What Is a Run?

When you confirm a proposal in Workspace chat, Mycelis creates a **run** — a single execution instance of the activated mission. Runs have:

- A unique `run_id` (e.g., `abc1234-...`)
- A start timestamp
- A status: `running`, `completed`, or `failed`
- An ordered list of **events**

---

## Opening the Timeline

After confirming a proposal, a pill appears in the chat:

```
⚡ Mission activated — abc1234... →
```

Click it to navigate to `/runs/{run_id}`. You can also navigate directly if you have the run ID.

---

## Reading the Timeline

The timeline is a vertical sequence of event cards, oldest at top, newest at bottom. Auto-refreshes every 5 seconds while the run is active.

### Event Card Anatomy

```
● tool.invoked         [admin]         2s ago
  write_file → /workspace/projects/parser.py
  ▼ (expand for full payload)
```

| Element | Description |
|---------|-------------|
| **Dot color** | Event severity / type (see table below) |
| **Event type** | What happened |
| **Source agent** | Which agent emitted this event |
| **Timestamp** | Relative time (hover for absolute) |
| **Summary** | Most useful payload field (tool name, file path, error) |
| **Expand chevron** | Click to see full JSON payload |

### Event Type Colors

| Event | Color | Meaning |
|-------|-------|---------|
| `mission.started` | Green | Run began |
| `mission.completed` | Green | Run finished successfully |
| `mission.failed` | Red | Run terminated with error |
| `tool.invoked` | Cyan | Agent called a tool |
| `tool.completed` | Blue | Tool finished |
| `tool.failed` | Red | Tool errored |
| `agent.started` | Muted | Agent began processing |
| `memory.stored` | Amber | Fact or artifact stored |
| `memory.recalled` | Amber | Memory was queried |
| `artifact.created` | Amber | File/document produced |

---

## Run Status

The header shows current run status:

| Status | Display | Meaning |
|--------|---------|---------|
| `running` | `● running` (pulsing) | Execution in progress |
| `completed` | `✓ completed` | Finished successfully |
| `failed` | `✗ failed` | Terminated with error |

While `running`, the timeline polls for new events every 5 seconds automatically.

---

## Navigation

- **← Workspace** link in the header returns to Mission Control
- The run URL (`/runs/{run_id}`) is bookmarkable and shareable — anyone with access can view the same timeline
- Each event card's expanded JSON is a complete audit record

---

## Common Patterns

### Tool Chain
A typical file-writing task produces:
```
mission.started
agent.started    [admin]
tool.invoked     [admin]  write_file → parser.py
tool.completed   [admin]  write_file ✓  (234 bytes)
artifact.created [admin]  parser.py
mission.completed
```

### Council Consultation
When Soma consults the Coder:
```
mission.started
agent.started    [admin]
tool.invoked     [admin]  consult_council → council-coder
agent.started    [council-coder]
tool.invoked     [council-coder]  write_file → impl.py
tool.completed   [council-coder]  write_file ✓
artifact.created [council-coder]  impl.py
tool.completed   [admin]   consult_council ✓
mission.completed
```

### Memory Operations
```
tool.invoked   [admin]  search_memory → "CSV parsing patterns"
memory.recalled [admin]  3 results found
tool.invoked   [admin]  remember → key: "user_prefers_type_hints"
memory.stored  [admin]  stored successfully
```
