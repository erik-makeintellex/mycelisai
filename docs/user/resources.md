# Resources
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Advanced-mode support surface for connected tools, deployment context intake, managed exchange, workspace files, AI engine setup, and reusable role definitions.

---

## Overview

Open `/resources` after turning on Advanced mode.

Current tabs:

| Tab | Purpose |
|-----|---------|
| Connected Tools | Installed MCP servers and tool capability visibility |
| Exchange | Inspect managed channels, research/result threads, trust labels, and review posture |
| Deployment Context | Load governed private/user records, customer context, approved company knowledge, Soma operating context, and reflection/synthesis observations into separate governed context lanes |
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

Current baseline posture:
- curated library installs are the default path
- `filesystem` and `fetch` are common curated entries, not assumed bootstrap defaults in the supported runtime lanes
- `memory` remains curated install and should not be treated as equivalent to Mycelis-governed memory/context lanes
- `artifact-renderer` remains planned

Key outcome:
Operators should be able to determine "what the system can access" directly from this tab.

Current posture:
- curated library installs are the default path
- `/api/v1/mcp/library/apply` is the one-call API for applying a curated potential source: it returns `installed` with server/tools/governance when allowed, or `requires_approval` with the inspection report when a policy boundary is still required
- local-first current-group configuration can install directly when policy allows
- remote or higher-risk entries can return an explicit approval boundary instead of silently installing
- credentialed external SaaS entries such as Slack, GitHub, hosted search, and hosted media should now be expected to require approval rather than behaving like low-risk local tools
- `fetch`/research capability is how governed external context can be added without treating web access as unrestricted trust
- the same Connected Tools surface should make the workflow legible end to end: add from the curated library, confirm the server is connected, and inspect recent persisted MCP activity plus live in-session usage showing which server/tool agents are using, including team, agent, and run labels when the runtime supplies them
- the curated MCP library is now being standardized around the MCP registry `server.json` concepts so future registrations stay recognizable outside Mycelis too: each entry should carry a canonical server name, version, published package + transport metadata, repository/homepage metadata when known, and typed environment-variable declarations instead of only a local command block
- curated MCP install is repeat-safe by server name; reapplying an allowed entry updates and reconnects the existing server instead of creating duplicate registry state
- Connected Tools should also make package-version policy visible instead of hiding it in install internals; the current library now also carries deployment-boundary and bundle-posture metadata, while the next interoperability slice should preserve enough metadata to round-trip against published `server.json` records without flattening Mycelis governance-specific fields
- enterprise packaging may later ship pinned supported bundle profiles for entries such as `filesystem`, `fetch`, `github`, `slack`, `postgres`, and `brave-search`, but free self-hosted deployments should still be able to install curated entries manually through the same governed path

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

Deployment Context is the governed intake surface for private/user-owned content and deployment-shaping knowledge that should influence future Soma reasoning without being treated as ordinary Soma memory.

It is not the same as team-shared execution memory. Team-shared continuity belongs in `AGENT_MEMORY`, while Deployment Context is for governed source material and promoted doctrine.

Typical inputs:
- private records or diary/journal notes the user explicitly wants Soma to use
- finance, legal, health, household, or business references tied to target goal sets
- customer deployment notes
- architecture briefs
- provider and MCP constraints
- security policies
- curated external research or handoff documents
- approved company-authored playbooks or guidance
- reflection/synthesis observations such as distilled lessons, inferred patterns, contradictions, shifts in user trajectory, and meta-observations about what is changing over time

Operational behavior:
- every load creates a durable document artifact plus vector-backed chunks in governed context lanes within the shared recall substrate
- each entry carries `knowledge_class`, visibility, sensitivity, trust, and provenance metadata
- uploaded text files are read into the same governed intake contract as pasted content
- `knowledge_class=user_private_context` is for private user-owned records, diary entries, finance notes, and other sensitive references; it defaults to private visibility, restricted sensitivity, and explicit goal-set metadata
- `knowledge_class=customer_context` is for operator/customer-provided material
- `knowledge_class=company_knowledge` is for approved company-authored guidance only
- `knowledge_class=soma_operating_context` is for root-admin or delegated-owner guidance that shapes shared Soma behavior across users
- `knowledge_class=reflection_synthesis` is the promotion target for distilled lessons, inferred patterns, contradictions, trajectory shifts, and meta-observations; agent-driven reflection should start as a Managed Exchange `LearningCandidate` with classification, confidence, and review posture before it is promoted
- team-shared execution memory should stay in scoped `AGENT_MEMORY`; loading a document here does not make it team memory by default
- promotion from customer context into company knowledge should happen through a governed approval path with lineage preserved, not by rewriting the original entry in place
- Soma operating context is stricter than ordinary deployment intake: it is normalized into admin guidance, stays globally scoped, and is intended for durable shared output/identity/stance shaping rather than personal chat preferences
- reflection/synthesis context is separate from Soma memory and from user-private/customer/company lanes so Soma can reason about what is changing over time without mixing those meta-observations into raw source material
- Soma, Council, and teams can recall allowed context during planning and answer generation without treating it as raw unrestricted web input
- private user context is only intended to enter agent work when its visibility/scope and target goal sets match the user’s request; it is not company knowledge and should not be promoted silently
- use `source_kind=web_research` or a stricter trust/sensitivity class when the content came from external sources

Key outcome:
Operators should be able to answer "what governed context did we intentionally load into Soma, which store did it enter, what target goals can use it, and under what trust boundary?" from one surface.

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
