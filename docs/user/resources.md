# Resources

> Configure model providers, MCP capability surfaces, workspace visibility, and agent capabilities.

---

## Overview

Open `/resources`.

Current tabs:

| Tab | Purpose |
|-----|---------|
| Brains | Provider configuration and health |
| MCP Tools | Installed MCP servers and tool capability visibility |
| Workspace Explorer | Filesystem MCP-backed browsing and file operations inside workspace boundary |
| Capabilities | Agent catalogue/templates |

---

## Brains

Brains define available model providers for orchestration roles.

Provider families:
- local/self-hosted (Ollama, vLLM, LM Studio)
- remote/commercial (OpenAI, Anthropic, Google)

What you can do:
- add/edit/delete providers
- probe health
- manage role routing through profiles

---

## MCP Tools

MCP servers expose tool capabilities agents can invoke during execution.

Baseline status (V7):
- `filesystem`: bootstrap default
- `fetch`: bootstrap default
- `memory`: curated install (available, not default bootstrap)
- `artifact-renderer`: planned

Key outcome:
Operators should be able to determine "what the system can access" directly from this tab.

---

## Workspace Explorer

Workspace Explorer now uses the `filesystem` MCP server directly from Resources.

Supported operator actions:
- browse directories (`list_dir`)
- read files (`read_file`)
- create directories (`create_dir`)
- write files (`write_file`)

Operational behavior:
- if `filesystem` is not installed or not connected, explorer shows actionable recovery controls
- all tool calls run through the same API request contract used by other resource channels
- workspace boundaries still apply (sandboxed filesystem rules)

---

## Capabilities

Capabilities tab is the catalogue surface for reusable agent definitions/templates.

Typical template fields:
- role
- model/provider expectations
- allowed tools
- input/output contracts
- validation strategy

---

## Operational Guidance

Use `Resources` to answer these operator questions quickly:
1. Which providers are online?
2. Which tools are accessible right now?
3. Are workspace file operations available?
4. Which capability templates are available for wiring?
