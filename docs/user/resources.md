# Resources
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Operator support surface for generated output files, capability readiness, deployment context intake, managed exchange, AI engine setup, and reusable role definitions.

---

## Overview

Open `/resources` directly from the main rail when you need generated files, capability readiness, workspace roots, or connected-tool posture.

Current resource menu:

| Resource Type | Purpose |
|-----|---------|
| Output Files | Open generated content folders and browse filesystem MCP-backed files inside workspace boundary |
| Capabilities | What Soma can use now, what needs repair, and what can be requested |
| Exchange | Inspect managed channels, research/result threads, trust labels, and review posture |
| Deployment Context | Load governed private/user records, customer context, approved company knowledge, Soma operating context, and reflection/synthesis observations into separate governed context lanes |
| AI Engines | Global AI engine configuration and health |
| Role Library | Reusable specialist-role definitions |

The Resources page keeps these resource types in a persistent menu and renders the selected type inside a bounded work window. Long tool lists, workspace folders, exchange records, and provider forms should scroll inside that selected panel rather than turning the whole page into one long operator path.

---

## AI Engines

AI engines define the curated model posture available to admin operators.

Provider families:
- local/self-hosted (Ollama, vLLM, LM Studio)
- remote/commercial (OpenAI, Anthropic, Google)

What you can do:
- add/edit/delete providers
- probe health
- manage role routing through profiles

---

## Capabilities

MCP servers, custom connectors, local scripts, external APIs, and future plugins expose capabilities Soma and teams can invoke during execution.

The Dashboard readiness strip summarizes search/tool posture for Soma, but
Resources is the primary place to inspect what Soma can use, repair missing capability, request MCP servers, check web search readiness, and review recent tool activity.

New-user readiness checks:
- Capabilities should tell you what Soma can use now, what needs repair, and what can be requested.
- Installed should tell you whether any MCP servers are connected or whether the first step is **Add MCP Server**.
- **Add MCP Server** should open the curated library, not a raw JSON/config paste box.
- Curated entries must name required environment variables without exposing secret values.
- After install or reapply, return to Installed and confirm the server card, tool list, capability binding, and recent MCP activity are visible.
- If the registry is unreachable, the UI should say that capability/tool readiness could not be confirmed and must not pretend the registry is truly empty.

Current baseline posture:
- curated library installs are the default path
- `filesystem` and `fetch` are common curated entries, not assumed bootstrap defaults in the supported runtime lanes
- `memory` remains curated install and should not be treated as equivalent to Mycelis-governed memory/context lanes
- `artifact-renderer` remains planned

Key outcome:
Operators should be able to determine "what Soma can currently use" directly from this tab, including what each capability is for, whether it is available, what risk level it has, what outputs it can produce, and where its results appear.

Capability manifest expectation:
- every built-in MCP, external tool, local script, custom connector, or plugin must register as a governed capability before Soma or a team can use it
- every meaningful execution should create or attach to a run
- every meaningful output should normalize into Managed Exchange, a retained artifact/output, audit evidence, or a learning candidate before it is treated as durable product state
- raw MCP/custom output should not be the final unmanaged state

MCP tool-set layering supports three configuration forms:
- `all`: a shared tool set available across the organization, used as the fallback for plain `toolset:<name>` references
- `group`: a grouped tool set targeted to a collaboration/team/group lane through `scope_ref`
- `host`: a host-targeted tool set for a specific deployment/runtime host through `scope_ref`

When the same tool-set name exists at multiple layers, scoped runtime resolution should prefer the group or host layer first, then fall back to the shared `all` layer. This lets operators keep a default capability posture while adding narrower MCP access for a project lane or a particular host.

The Capabilities page now exposes this as **MCP access layers**. Use **All** for organization-wide defaults, **Group** for a collaboration lane, and **Host** for a target runtime host. Group and Host layers require a target id before they can be saved, and saved layers should appear in the existing-layers list with their tool references visible for review.

Review/edit expectation:
- the installed server card should expand into an MCP structure view with transport, status, command or endpoint, arguments, env/header references, discovered tools, and recent use
- the capability view should show manifest identity, input/output schema posture, risk, approval, availability, fallback, allowed roles, and output destinations
- secrets should appear only as references or redacted values; set or rotate values in `.env` or the configured secret backend
- use **Add MCP Server** to install, reapply, or edit the curated server shape instead of pasting raw MCP config into the UI
- after changing structure or secrets, return to Installed and confirm the server card, tool list, and recent MCP activity match the expected shape

Current posture:
- curated library installs are the default path
- `/api/v1/mcp/library/apply` is the one-call API for applying a curated potential source: it returns `installed` with server/tools/governance when allowed, or `requires_approval` with the inspection report when a policy boundary is still required
- curated `filesystem` installs are repeat-safe and bind to the deployment workspace root, such as `/data/workspace` in the supported Docker Compose runtime
- local-first current-group configuration can install directly when policy allows
- remote or higher-risk entries can return an explicit approval boundary instead of silently installing
- credentialed external SaaS entries such as Slack, GitHub, hosted search, and hosted media should now be expected to require approval rather than behaving like low-risk local tools
- `brave-search` provides governed web search when installed with `BRAVE_API_KEY`; `fetch` retrieves explicit URLs for analysis, and together they form the default curated research toolset without making web access unrestricted trust
- `Mycelis Search Capability` shows the active Soma search posture directly in Capabilities: the selected provider, whether Soma can call `web_search`, whether local shared sources or public web are supported, and whether the current path needs hosted Brave credentials
- Soma's Operator trust package also names the active search source boundary for `web_search` results, such as `Search source: Local Mycelis context`, so operators can distinguish retained Mycelis context from public-web providers
- self-hosted search does not have to depend on Brave tokens: `local_sources` is the default token-free provider for governed Mycelis context and falls back to bounded text search when embeddings are unavailable, `local_api` can call an operator-owned HTTP search endpoint, and the supported Compose release path starts SearXNG for public web search through an operator-owned endpoint
- the same Capabilities surface should make the workflow legible end to end: choose **Add MCP Server**, confirm the capability is available, and inspect recent persisted MCP activity plus live in-session usage showing which server/tool agents are using, including team, agent, and run labels when the runtime supplies them
- the curated MCP library is now being standardized around the MCP registry `server.json` concepts so future registrations stay recognizable outside Mycelis too: each entry should carry a canonical server name, version, published package + transport metadata, repository/homepage metadata when known, and typed environment-variable declarations instead of only a local command block
- curated MCP install is repeat-safe by server name; reapplying an allowed entry updates and reconnects the existing server instead of creating duplicate registry state
- Capabilities should also make package-version policy visible instead of hiding it in install internals; the current library now also carries deployment-boundary and bundle-posture metadata, while the next interoperability slice should preserve enough metadata to round-trip against published `server.json` records without flattening Mycelis governance-specific fields
- enterprise packaging may later ship pinned supported bundle profiles for entries such as `filesystem`, `fetch`, `github`, `slack`, `postgres`, and `brave-search`, but free self-hosted deployments should still be able to install curated entries manually through the same governed path

Useful Soma prompts from this surface:
- `Search the web for "<topic>", summarize the strongest sources, and cite them.`
- `Use host data under workspace/shared-sources and list the files that shaped the answer.`
- `Review current MCP servers, tools, and recent use, then tell me which agents should have which tools.`
- `Review the private-service or private-data boundary for this action, name the needed MCP server and .env variables, and ask me to confirm before enabling or assigning tools.`

When the request includes private services, credentials, production systems, customer/private data, or recurring tool behavior, Soma should use the protected interaction-template path: identify the matched theme, name the protection reason, confirm the scope, and then use the governed proposal path before action.

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

## Output Files

Output Files is the default `/resources` view and uses the `filesystem` MCP server directly from Resources.
The browser starts at the MCP-safe `workspace` root rather than the Core
process working directory, so ordinary browse/read/write actions stay inside
the configured mounted data boundary.

Output Files now starts with a **Group outputs** selector when retained group
artifacts exist. The selector only lists groups that produced retained
user-facing outputs through `/api/v1/groups/{id}/outputs`; groups with no
deliverable artifacts stay out of the output picker so operators do not have to
scan abandoned or internal-only lanes.

Group workflow logs and chat-pipeline history are reviewed in `Groups ->
Workflow Log`, not in `Resources -> Output Files`. Output Files should stay
focused on durable artifacts that a user can open, download, preview, or reveal
in the workspace. When you need the work history that led to an artifact, open
the group and review its Workflow Log.

By default, selecting a group opens its retained output artifacts, such as final
documents, packages, media, or generated files. Team-generated working files
used to build the final deliverable stay hidden from this curated output list.
Use **Include team source files** only when an operator needs to inspect the
group workspace folder itself, including intermediate files under the same
`workspace_folder`.

When a retained group has many outputs, the group-output selector splits the
artifact cards by contributor level. Use **All**, **Team lead**, **Coders**,
**Review**, **Media**, or **Other** to narrow the visible artifacts without
leaving the selected group. Mycelis uses artifact metadata such as role/agent
level when present and falls back to the artifact agent id, title, and type.

The workspace explorer is organized around three operator steps:
- `Find outputs` lists retained files and folders and opens file selections into preview.
- `Preview` reads the selected generated file without leaving the Resources surface.
- `Create` keeps small handoff-folder and handoff-file writes available without making write controls the default browse path.

The top of the panel includes **Open folder** for the current workspace path.
Use it when an operator wants to grab generated files, media proof, project
packages, or browser-game output from the local machine without decoding the
storage configuration. Retained output cards in Soma, Teams, and Groups should
also expose **Open folder** when they carry a workspace path.

Output locations:
- generated files, project packages, browser games, and filesystem MCP writes land under `MYCELIS_WORKSPACE`
- standing or Soma-created group deliverables should land under `MYCELIS_WORKSPACE/groups/...`; the Groups detail pane shows the exact `workspace_folder` and an `Open folder` action
- file-backed artifacts and cached media land under `MYCELIS_ARTIFACT_ROOT`
- `DATA_DIR` remains a legacy alias for artifact storage and should match `MYCELIS_ARTIFACT_ROOT` until older paths are removed

For local source development, the default readable shape is:

```text
MYCELIS_WORKSPACE=./workspace
MYCELIS_ARTIFACT_ROOT=./workspace/artifacts
```

Use `System -> Deployments` to confirm the runtime is reporting the same workspace root and artifact root that you expect on disk.

New-user proof should verify both sides of this boundary:
- `Resources -> Capabilities` shows whether `filesystem` is installed and connected.
- `Resources -> Output Files` lists only groups with retained user-facing output in the group selector.
- `Resources -> Output Files` can narrow retained artifacts by contributor level before opening a file or package.
- `Resources -> Output Files -> Include team source files` switches from curated output artifacts to the selected group's workspace source folder.
- `Resources -> Output Files` can browse/read/write only under the governed workspace boundary and can open the current local folder through the workspace-confined reveal endpoint.
- `System -> Deployments` reports the deployment/workspace/artifact roots that explain where generated output will land.
- A retained demo output or project package opened from Soma/Teams/Groups resolves to the same workspace root family instead of a hidden process working directory; team-owned packages and media should be inside the selected group folder unless the operator explicitly chose another workspace path.

Supported operator actions:
- browse directories (`list_directory`)
- read files (`read_text_file`)
- create directories (`create_directory`)
- write files (`write_file`)

Operational behavior:
- if `filesystem` is not installed or not connected, explorer shows actionable recovery controls
- the recovery state keeps two paths visible: **Open Capabilities** to repair or install filesystem MCP, and **View storage roots** to confirm where generated output is mounted while MCP recovers
- all tool calls run through the same API request contract used by other resource channels: `{"arguments": {...}}`
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
