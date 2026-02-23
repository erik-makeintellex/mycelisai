# Resources

> Configure the intelligence infrastructure: brains (LLM providers), tools (MCP servers), workspace, and the agent catalogue.

---

## Overview

The **Resources** page (`/resources`) has four tabs:

| Tab | Purpose |
|-----|---------|
| **Brains** | LLM provider configuration and health |
| **Tools** | Installed MCP tool servers and their capabilities |
| **Workspace** | Files and artifacts in the managed workspace |
| **Catalogue** | Agent definition templates |

---

## Brains

Brains are the **inference providers** that power your agents. Each council role (admin, architect, coder, creative, sentry) is mapped to a brain via a profile in the Cognitive Matrix.

### Supported Providers

| Type | Examples |
|------|---------|
| **Local (self-hosted)** | Ollama, vLLM, LM Studio |
| **Commercial** | OpenAI (GPT-4), Anthropic (Claude), Google (Gemini) |

Any provider that speaks the OpenAI-compatible API format can be pointed at any host on your network — useful for running models on a separate GPU machine.

### Health Status

Each brain card shows live health:
- **Online** (green) — endpoint reachable, model responding
- **Offline** (red) — endpoint unreachable
- **Error** (amber) — reachable but returning errors

The system probes all brains every 15 seconds.

### Enabling / Disabling a Brain

Toggle the switch on any brain card to enable or disable it. A disabled brain is excluded from routing — the profile will fall back to the next available brain, or agents will enter a degraded state if no fallback exists.

### Cognitive Matrix

The **Matrix** button opens the full profile-routing table:

```
Profile     →    Provider
─────────────────────────
admin       →    ollama
architect   →    ollama
coder       →    vllm
creative    →    ollama
sentry      →    ollama
```

Click any cell to change which brain handles that role. Changes take effect immediately for new requests; in-flight conversations are not affected.

---

## Tools (MCP Servers)

**MCP (Model Context Protocol)** servers extend what agents can do. Installed servers expose tools that agents call during their ReAct loops.

### Built-in MCP Servers

These are pre-installed and always available:

| Server | Tools | Purpose |
|--------|-------|---------|
| `filesystem` | read_file, write_file, list_dir, create_dir, append_file | Sandboxed workspace I/O (`/workspace` only) |
| `memory` | store_fact, semantic_search, retrieve_by_id | Persistent semantic memory (pgvector) |
| `artifact-renderer` | render_artifact | Structured output (markdown, tables, JSON, images) |
| `fetch` | fetch_url | Controlled HTTP GET research (domain allowlist) |

### Installing Additional Servers

1. Click **+ Install Server**
2. Choose transport:
   - **stdio** — local process (command + arguments)
   - **sse** — HTTP server (URL)
3. Provide configuration details
4. Click **Install** — the server connects and discovers its tools automatically

### Testing a Tool

Expand any server card → find a tool → click **Test** → fill in the input form → click **Run**. The tool executes and the result is shown inline. This is a live test — results are real.

### Removing a Server

Click the trash icon on a server card. Any agents with tools from that server will show a "Tool unavailable" warning.

---

## Workspace

The managed workspace is a sandboxed filesystem at `/workspace` that agents can read and write (via the `filesystem` MCP server).

### Directory Structure

```
/workspace
  /projects    — code and project files
  /research    — fetched documents and notes
  /artifacts   — generated outputs (code, docs, data)
  /reports     — structured mission reports
  /exports     — files staged for external use
```

The Workspace tab shows the current contents with:
- File tree navigation
- Preview pane (text, markdown, JSON rendered inline)
- Download button for individual files
- Creation timestamp and size

Files created by agents (via `write_file` or `store_artifact`) appear here automatically.

---

## Catalogue

The catalogue is a library of **agent definition templates** — reusable blueprints that define what an agent is before it is assigned to a team.

### Agent Template Fields

| Field | Description |
|-------|-------------|
| **Name** | Unique identifier |
| **Role** | Semantic role (researcher, writer, reviewer, etc.) |
| **System Prompt** | The agent's operational directive |
| **Model** | Override the default profile routing (optional) |
| **Tools** | Which tools this agent can call |
| **Inputs** | NATS topic patterns this agent listens on |
| **Outputs** | NATS topic patterns this agent publishes to |
| **Verification** | How outputs are validated before trust scoring |

### Using a Template

Templates are instantiated when a mission blueprint references them. The Wiring canvas can drag agent templates from the catalogue onto the circuit board to manually compose teams.

### Creating a New Template

1. Click **+ New Agent**
2. Fill in the template form
3. Click **Save** — the template is immediately available for team composition

Templates are versioned. Editing a template does not affect running agents that were instantiated from it.
