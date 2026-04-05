# Resources

> Advanced-mode support surface for connected tools, deployment context intake, managed exchange, workspace files, AI engine setup, and reusable role definitions.

---

## Overview

Open `/resources` after turning on Advanced mode.

Current tabs:

| Tab | Purpose |
|-----|---------|
| Connected Tools | Installed MCP servers and tool capability visibility |
| Exchange | Inspect managed channels, research/result threads, trust labels, and review posture |
| Deployment Context | Load governed customer context and approved company knowledge into a separate vector-backed context store for Soma |
| Workspace Files | Filesystem MCP-backed browsing and file operations inside workspace boundary |
| AI Engines | Global AI engine configuration and health |
| Role Library | Reusable specialist-role definitions |

---

## AI Engines

AI engines define the curated model posture available to advanced operators.

Provider families:
- local/self-hosted (Ollama, vLLM, LM Studio)
- remote/commercial (OpenAI, Anthropic, Google)

What you can do:
- add/edit/delete providers
- probe health
- manage role routing through profiles

---

## Connected Tools

MCP servers expose tool capabilities agents can invoke during execution.

Baseline status (V7):
- `filesystem`: bootstrap default
- `fetch`: bootstrap default
- `memory`: curated install (available, not default bootstrap)
- `artifact-renderer`: planned

Key outcome:
Operators should be able to determine "what the system can access" directly from this tab.

Current posture:
- curated library installs are the default path
- local-first current-group configuration can install directly when policy allows
- remote or higher-risk entries can return an explicit approval boundary instead of silently installing
- `fetch`/research capability is how governed external context can be added without treating web access as unrestricted trust

## Exchange

Exchange is the inspectable context-security surface for advanced operators.

What you can inspect:
- normalized channels such as research/result outputs
- active review threads
- recent exchange items
- trust and sensitivity posture on outputs

Typical labels:
- `sensitivity_class`
- `trust_class`
- `review_required`
- capability-linked output context

Key outcome:
Operators should be able to answer "what entered the system, how trusted is it, and does it need review?" without reading raw logs.

---

## Deployment Context

Deployment Context is the governed intake surface for deployment-shaping knowledge that should influence future Soma reasoning without being treated as ordinary Soma memory.

Typical inputs:
- customer deployment notes
- architecture briefs
- provider and MCP constraints
- security policies
- curated external research or handoff documents
- approved company-authored playbooks or guidance

Operational behavior:
- every load creates a durable document artifact plus vector-backed chunks in the governed context store
- each entry carries `knowledge_class`, visibility, sensitivity, trust, and provenance metadata
- `knowledge_class=customer_context` is for operator/customer-provided material
- `knowledge_class=company_knowledge` is for approved company-authored guidance only
- promotion from customer context into company knowledge should happen through a governed approval path with lineage preserved, not by rewriting the original entry in place
- Soma, Council, and teams can recall this context during planning and answer generation without treating it as raw unrestricted web input
- use `source_kind=web_research` or a stricter trust/sensitivity class when the content came from external sources

Key outcome:
Operators should be able to answer "what governed context did we intentionally load into Soma, which store did it enter, and under what trust boundary?" from one surface.

---

## Workspace Files

Workspace Files uses the `filesystem` MCP server directly from Resources.

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

## Role Library

Role Library is the catalogue surface for reusable specialist definitions and templates.

Typical template fields:
- role
- model/provider expectations
- allowed tools
- input/output contracts
- validation strategy

---

## Operational Guidance

Use `Resources` to answer these operator questions quickly:
1. Which AI engines are online?
2. Which tools are accessible right now?
3. How is external or research context classified and reviewed?
4. What deployment knowledge has been intentionally loaded into long-term context?
5. Are workspace file operations available?
6. Which role definitions are available for advanced workflow work?
