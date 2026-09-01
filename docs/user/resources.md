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
| Capabilities | What Soma can use now, what needs attention, and what can be requested |
| Exchange | Inspect managed channels, research/result threads, trust labels, and review posture |
| Deployment Context | Save files or notes Soma should reuse as long-lived, scoped source context |
| AI Engines | Global AI engine configuration and health |
| Worker Profiles | Ready-made teammates and Soma-guided custom profile creation |
The Resources page keeps these resource types in a persistent selector and renders the selected type inside a bounded work window. On phones and tablets, Resource types appear as one compact horizontally scrollable tab row so the selected work surface remains near the top of the screen. On desktop, the same choices use a vertical list-detail menu with short descriptions. Selection is recorded in the page URL, survives refresh, and works with browser Back. Long tool lists, workspace folders, exchange records, and provider forms scroll inside the selected panel rather than turning the whole page into one long operator path.
Retained group outputs use the same Outcome Health labels as Soma, Teams, and Runs. A selectable group with retained user output is Completed; the badge describes operational state while proof and source material remain separate details.
---

## Connected Tools

Connected tools are reviewed through `Resources -> Capabilities`: the user-facing question is what Soma can use, where that permission applies, what needs attention, and what can be requested. The default readiness view uses compact counts and origin summaries instead of capability inventory cards. Raw server structure stays behind Catalog details or Inspect.

For web access specifically, use `Resources -> Capabilities -> Soma research access`. That lane should show whether Soma currently has approved local data, mounted-source search, public-web search, and direct Soma Search available. When public web is already available, the lane should say Soma can search now and offer **Add URL reader MCP** only for explicit supplied-URL retrieval through tools such as `fetch`. When public web is missing, it should offer **Add web search provider** or **Set up web search** and guide the operator toward a safe search provider such as built-in web, SearXNG/local API, or Brave. Built-in Mycelis Soma Search does not depend on `fetch`; add or repair `fetch` when users need Soma or a team to retrieve a specific supplied URL. The default UI should say **Soma Search** or **Research**; raw tool IDs such as `web_search` belong in Inspect/API details only.

The web/search setup action should keep the user on the Capabilities surface and open the **Access** lane with the matching search-source form ready. Users should not have to discover that form manually after asking to add public web, mounted data, or private API search.

Search scope must stay honest without making ordinary asks feel like configuration work. Natural phrasing such as "can you search on this topic" or "search for this topic" runs the configured web search path; in the default release posture that means built-in token-free web search through `builtin_web`. Explicit local wording such as "search local sources", "search internal docs", "search mounted data", or "search retained context" uses governed `local_sources` instead. Only setup/status questions such as "can you search the web?" should return capability guidance. If the active provider is only `local_sources` and the user asks for web research, Soma should explain that public-web research is not configured and offer either web-source setup or a local/shared-source search instead. When a user asks for both internal/local context and public research, Soma requests `all`; if both boundaries are available, it searches approved local/mounted sources and public web together, and if only one side is available, the answer should say which source boundary was searched and which boundary is still missing.

Search sources also support client-owned data: private docs, customer portals, repositories, issue trackers, file stores, intranet search, or SaaS knowledge bases. Current Capabilities show configured sources Soma may use, including source boundary, endpoint/base URL or approved path, scope, auth scheme, sensitivity/trust, status, and recovery. Use **Add search source** to name the source, choose type, provide the base URL/endpoint when the source is external or API-backed, provide the approved path when it is a mounted folder or repository/code folder, choose auth scheme, select a secret reference, scope it to Everyone/Group/Host, and keep sensitivity/trust defaults visible. Operator-managed sources can be edited or removed from the same lane. The source list should say whether the source is ready for Soma now, ready once saved access is available, or only registered safely while another auth adapter is still needed. When Soma is asked to use a configured source, the runtime checks that the source is available, in scope for the current group or host, supported by a safe adapter, and able to resolve the referenced secret before searching it. Source records are PostgreSQL-backed and fail closed when durable persistence is unavailable. Raw tokens stay in `.env` or the configured secret backend. Code repositories and local code folders are a specialized `code_context` source type shown to users as **Repository or code folder** access. When registered, Soma can build a repository map from the approved boundary so it can answer questions such as "what touches this feature?", "what files should change?", or "which tests are likely affected?" without embedding the whole codebase into a prompt. The map is built by Mycelis from local source structure and content hashes; it is not a separate external service. In `Resources -> Capabilities -> Access -> Search sources`, code-context rows show the source name, Everyone/Group/Host scope, snapshot status and last mapped time, index status and last index time, plus plain Refresh or Repair guidance when the map is stale or broken. Raw snapshot refs, index refs, digests, and code-scope refs stay behind **Inspect technical refs**. Soma should cite file, symbol, and range refs where possible, label inferred impact separately from extracted parser facts, and ask for refresh or repair when the map is stale.

Local file access should use **named data mounts**, not arbitrary path guessing. Use Resources to list user-owned local folders or infrastructure shared folders that Soma may read, such as `workspace/client-docs`, a mounted network share, or a host-specific project data folder. Each mount needs a readable name, path/root, boundary description, mode such as read-only or read/write, scope (`Everyone`, `Group`, or `Host`), sensitivity/trust defaults, and recovery guidance. Once configured and in scope, mounted folders are live Soma search sources; Soma should name the mount when it uses files from it, and should not search or read unlisted paths just because they exist on the host.

## Outcome Templates
Outcome Templates help Soma ask only the few delivery questions that materially affect quality. They can define a minimum brief, defaults, expected output shape, validation expectations, capability/context references, and governance posture, then compile into the normal WorkIntent and Outcome journey. They are not saved chat prompts, automatic jobs, or a second execution system.
There are two equivalent authoring paths:
- Ask Soma to draft an Outcome Template, preview the resolved brief and effects, and save or activate it through governed configuration tools.
- Write YAML or JSON under `MYCELIS_CONFIG_ROOT` (by default `MYCELIS_WORKSPACE/config`) and ask Soma or the configuration API to preview, store, and activate that file.
Preview never persists or activates a document. Store creates an immutable revision. Activate changes the selected revision atomically and records who requested it; rollback explicitly reactivates a chosen prior revision. Running or approved work keeps the resolved template version and digest, so changing a template later does not rewrite historical authority. Configuration documents contain secret references only; put secret values in `.env` or the configured secret backend.
Conversation Templates remain separate: they render reusable non-executing asks that return to Soma chat. Use an Outcome Template when repeated work needs a stable delivery brief and validation shape.

---

## AI Engines

AI engines define the curated model posture available to admin operators.

Provider families:
- local/self-hosted (Ollama, vLLM, LM Studio)
- remote/commercial (OpenAI, Anthropic, Google)
- optional model gateway (LiteLLM), shown disabled until an administrator connects a reviewed proxy and scoped client key

LiteLLM is a transport option, not another Soma or team system. When configured, Mycelis still decides which profiles may use the provider, whether data may leave the organization, and whether approval is required. The gateway may route only inside that approved boundary; it must never turn local-only work into remote inference. The current release does not install or start LiteLLM, and ordinary users should see it only as a disabled or configured AI Engine rather than a new workflow surface.

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
- Capabilities should tell you what Soma can use now, what needs attention, and what can be requested.
- The default Overview tab should tell you what Soma can use now, what needs attention, and what can be requested without making large overview tiles dominate the first viewport.
- Servers should tell you whether any MCP servers are connected or whether the first step is **Add connector**. Its count is server inventory, not the number of capabilities Soma can use.
- **Add connector** should open the curated library, not a raw JSON/config paste box.
- Curated entries must name required environment variables without exposing secret values.
- After install or reapply, return to Servers and confirm the server card, tool list, capability binding, and recent MCP activity are visible.
- If the registry is unreachable, the UI should say that capability/tool readiness could not be confirmed and must not pretend the registry is truly empty.

Current baseline posture:
- curated library installs are the default path
- `filesystem` and `fetch` are common curated entries, not assumed bootstrap defaults in the supported runtime lanes
- `memory` remains curated install and should not be treated as equivalent to Mycelis-governed memory/context lanes
- `artifact-renderer` remains planned

Key outcome:
Operators should be able to determine "what Soma can currently use" directly from this tab. The default view should put web-access status and the compact capability overview ahead of examples or workflow education. The readiness summary uses small **Ready**, **Needs attention**, and **Available to add** controls plus compact origin summaries; it does not preview an arbitrary subset of capability names. The Catalog keeps capability name, purpose, availability, origin, risk, and approval posture first; output destinations, bindings, fallback behavior, and audit details belong behind **Inspect capability details** unless the capability needs attention. Command examples and workflow education may remain available as expandable guidance, but they should not crowd the main status readout.

Capability origin is separate from readiness. Overview summarizes four origins without listing every capability:
- **Built-in runtime**: implemented inside Mycelis; no MCP server is involved.
- **Host / CLI**: an allowlisted command or script exposed by the machine or container running Core. It is governed by Mycelis, but availability depends on that deployed host.
- **MCP**: supplied by an installed MCP server. Catalog labels include the server name when the runtime provides it.
- **Connector**: an external API, plugin, or provider connection that does not use MCP.
Local image generation is a provider connection, not an MCP requirement. For Pinokio Forge, launch Forge with `--api` enabled and configure its base URL, normally `http://127.0.0.1:7860`. Soma checks the stable Forge API before proposing image work. If the Forge UI is open but API mode is off, Soma should say exactly that in the Soma thread, include the next recovery step inline, and avoid creating a proposal, approval, team, failed run, or internal planning deliverable. The operational alert remains available for inspection, but the chat reply must be enough for a non-technical operator to know what to do next. The saved image under `groups/<team-id>/media` is the user deliverable. Team planning files remain internal and appear only through Workflow Log or Inspect.

Code-generated visual work is different. A request for a web page drawn with SVG, inline vector markup, HTML canvas, CSS art, a runnable browser game, app package, voxel-style interface, browser-native audio, or "no external assets" should stay on the app/package delivery path unless the user explicitly asks for generated image/media files. These requests do not require an image generator. Soma may still assign a visual/game specialist, but the expected deliverable is an openable package under `groups/<team-id>/generated/...`, not an image-provider output. That package must be validated like other interactive outputs before Soma calls it complete.

For an SVG or other code-rendered visual, opening successfully is not enough. The result should visibly express the requested subject and relationships, provide readable context and hierarchy, and make its promised interaction work. A minimal placeholder diagram or a control whose visible state does not change should be returned for repair rather than presented as a completed result.
Open **Catalog** for the full inventory. Filter it by origin to answer whether an action uses Mycelis runtime code, host/container tooling, or an MCP server. The inventory stays inside a bounded scrolling region and loads more entries deliberately so a large registry does not make the Resources page grow indefinitely.

Capability manifest expectation:
- every built-in MCP, external tool, local script, custom connector, or plugin must register as a governed capability before Soma or a team can use it
- every meaningful execution should create or attach to a run
- every meaningful output should normalize into Managed Exchange, a retained artifact/output, audit evidence, or a learning candidate before it is treated as durable product state
- raw MCP/custom output should not be the final unmanaged state

Capability permission groups support three configuration forms:
- `Everyone`: shared defaults Soma may use across the workspace after normal governance and approval
- `Group`: permissions limited to one Outcome, group, or collaboration lane
- `Host`: permissions limited to one deployment/runtime host

Under the hood these still save as MCP tool-set scopes (`all`, `group`, and `host`). When the same tool-set name exists at multiple layers, scoped runtime resolution should prefer the group or host layer first, then fall back to the shared `all` layer. This lets operators keep a default capability posture while adding narrower MCP access for a project lane or a particular host.

The Capabilities page opens as a focused readiness surface, not as one long MCP configuration document. Use the focus buttons to choose the current job: **Readiness** for web/search and compact origin posture, **Catalog** for the bounded, origin-filtered capability inventory, **Access** for sources/scopes/data, and **Inspect** for raw refs, provider bindings, workflow examples, and deeper technical evidence. Raw capability refs, output/write channels, provider bindings, and longer examples stay behind **Inspect capability details** or the **Inspect** focus.

Inside **Access**, choose the job you are doing instead of scrolling one mixed setup page:

- **Search sources**: governed places Soma can search, such as public web, local sources, mounted folders, code repositories, private APIs, or client-owned knowledge systems.
- **Live inputs**: buffered service, device, webhook, scheduler, database/event, or workflow feeds that Soma and teams may reference while working.
- **Service connections**: named tool/service access, connector scopes, and MCP/tool-set layering.

Live inputs should show only the source name, status, adapter kind, buffer mode, and scope by default. The operator can preview a small bounded buffer before routing the feed into Soma or a team. Raw ingress subjects, secret refs, and transport details stay behind **Inspect source**.

Capability permissions use **Everyone** for workspace defaults, **Group** for a collaboration lane, and **Host** for a target runtime host. Group and Host permissions require a target before they can be saved. Saved permission groups appear in the current-permissions list with a plain-language summary plus capability references for review.

Use **Common choices** when you do not want to type raw capability references. The current choices cover Workspace files, User data mounts, Web research, Team coordination, and Local host/media. Choosing one fills the matching capability refs and, when appropriate, nudges the form toward Group or Host scope so sensitive access does not accidentally become a workspace default. Advanced operators can still edit the refs before saving.

Review/edit expectation:
- the installed server card should expand into an MCP structure view with transport, status, command or endpoint, arguments, env/header references, discovered tools, and recent use
- the capability view should show manifest identity, input/output schema posture, risk, approval, availability, fallback, allowed roles, and output destinations
- secrets should appear only as references or redacted values; set or rotate values in `.env` or the configured secret backend
- use **Add connector** to install, reapply, or edit the curated server shape instead of pasting raw MCP config into the UI
- after changing structure or secrets, return to Servers and confirm the server card, tool list, and recent MCP activity match the expected shape

Current posture:
- curated library installs are the default path
- `/api/v1/mcp/library/apply` is the one-call API for applying a curated potential source: it returns `installed` with server/tools/governance when allowed, or `requires_approval` with the inspection report when a policy boundary is still required
- curated `filesystem` installs are repeat-safe and bind to the deployment workspace root, such as `/data/workspace` in the supported Docker Compose runtime
- local-first current-group configuration can install directly when policy allows
- remote or higher-risk entries can return an explicit approval boundary instead of silently installing
- credentialed external SaaS entries such as Slack, GitHub, hosted search, and hosted media should now be expected to require approval rather than behaving like low-risk local tools
- `brave-search` provides optional MCP-governed web search when installed with `BRAVE_API_KEY`; `fetch` retrieves explicit URLs for analysis, while built-in Mycelis `web_search` remains the default Soma search path when configured
- `Search provider details` in Inspect shows the active Soma search posture: the selected provider, whether Soma can search naturally from chat, whether approved local data/mounted sources or public web are supported, and whether the current path needs hosted Brave credentials
- Soma's Operator trust package also names the active search source boundary for Soma Search results, such as `Search source: Public web` or `Search source: Local Mycelis context`, so operators can distinguish public-web providers from retained Mycelis context
- web search does not have to depend on Brave tokens: `builtin_web` is the default token-free public web provider, `local_sources` remains the governed retained-context provider and falls back to bounded text search when embeddings are unavailable, `local_api` can call an operator-owned HTTP search endpoint, and the supported Compose release path can start SearXNG for public web search through an operator-owned endpoint
- client-owned search sources and local data mounts support persisted source records for readable name, source type, endpoint/base URL or local path, boundary, scope, auth or mount mode, secret reference when needed, sensitivity, trust defaults, live-vs-indexed mode, status, and recovery guidance; operator-managed records can be edited or removed, and `source_id` selection routes local-source, mounted-folder, repository/code-context, local-API, and SearXNG-compatible searches through scope/status guardrails before use; bearer/API-token env references can route through the local API adapter, while unsupported auth shapes stay registered-but-blocked until a safe adapter is available
- repository/code-context sources should show snapshot freshness, commit or content digest, index status, skipped-path diagnostics, and a plain **Refresh map** or **Repair source** action; raw graph details, parser output, and edge metadata stay behind Inspect
- when Soma searches configured sources, the answer should name the source boundary, cite or reference consulted sources where possible, and distinguish public web leads from private/customer/contextual sources before asking the operator to trust the result
- the same Capabilities surface should make the workflow legible end to end: choose **Add connector**, confirm the capability is available, and inspect recent persisted MCP activity plus live in-session usage showing which server/tool agents are using, including team, agent, and run labels when the runtime supplies them
- service MCPs such as PostgreSQL should be configured as named data-source connections. The Mycelis application database is system-owned and reserved by default; user/customer databases should be added as separate named connection profiles with secret references, readable purpose, and `Everyone`, `Group`, or `Host` scope so Soma can name which database it used and avoid confusing backend infrastructure with user data sources.
- the curated MCP library is now being standardized around the MCP registry `server.json` concepts so future registrations stay recognizable outside Mycelis too: each entry should carry a canonical server name, version, published package + transport metadata, repository/homepage metadata when known, and typed environment-variable declarations instead of only a local command block
- curated MCP install is repeat-safe by server name; reapplying an allowed entry updates and reconnects the existing server instead of creating duplicate registry state
- Capabilities should also make package-version policy visible instead of hiding it in install internals; the current library now also carries deployment-boundary and bundle-posture metadata, while the next interoperability slice should preserve enough metadata to round-trip against published `server.json` records without flattening Mycelis governance-specific fields
- enterprise packaging may later ship pinned supported bundle profiles for entries such as `filesystem`, `fetch`, `github`, `slack`, `postgres`, and `brave-search`, but free self-hosted deployments should still be able to install curated entries manually through the same governed path

Useful Soma prompts from this surface:
- `Search the web for "<topic>", summarize the strongest sources, and cite them.`
- `Use host data under workspace/shared-sources and list the files that shaped the answer.`
- `List the local data mounts Soma can read, then use the approved customer-docs mount for this research.`
- `Map this repository source, then tell me which files and tests are likely affected by changing the Resources capabilities page.`
- `Search the approved customer portal and company docs for the current onboarding policy, then tell me which source each claim came from.`
- `Add an authenticated search source for this internal docs URL using the configured token reference, limited to this group, and ask me to approve before Soma uses it.`
- `Register this local device feed as a latest-state input for the facilities group, then have Soma summarize changes every hour.`
- `Register this webhook as an append-log input for the support Outcome and route only review-worthy events to the team.`
- `Review current MCP servers, tools, and recent use, then tell me which agents should have which tools.`
- `Review the private-service or private-data boundary for this action, name the needed MCP server and .env variables, and ask me to confirm before enabling or assigning tools.`

When the request includes private services, credentials, production systems, customer/private data, or recurring tool behavior, Soma should use the protected interaction-template path: identify the matched theme, name the protection reason, confirm the scope, and then use the governed proposal path before action.

## Exchange

Exchange is the handoff review surface for evidence moving between Soma, teams, tools, and retained outputs. The default view should answer what was handed off most recently and whether it needs review. Work threads and source lanes remain available as focused tabs for advanced review, but the page should not force users to compare channels, threads, and items in three dense columns.

What you can inspect:
- recent handoffs and normalized outputs that another team or Soma may use next
- active work threads for planning, review, escalation, and learning
- source lanes/channels that explain where handoffs are allowed to move
- trust and sensitivity posture on outputs

Typical labels:
- `sensitivity_class`
- `trust_class`
- `review_required`
- capability-linked output context

Key outcome:
Operators should be able to answer "what entered the system, how trusted is it, and does it need review?" without reading raw logs.

Event-producing services and devices should appear as registered input sources, not as raw bus subjects. The setup path should name the source, choose its scope, choose a buffer mode, and bind it to an Outcome/group only when a team is expected to react. Fast sources should usually use latest-state or windowed summaries; audit-worthy callbacks should use append logs. Soma should tell the operator which buffer it used before asking a team to act. In the UI, these appear at `Resources -> Capabilities -> Access -> Live inputs`.

---

## Deployment Context

Deployment Context is the governed intake surface for files, notes, private/user-owned content, and deployment-shaping knowledge that should influence future Soma or team reasoning beyond one draft. In the product, this is the place to put **Context for Soma** when the material should persist, carry provenance, stay scoped, and remain inside an explicit trust boundary.

It is not the same as team working files, generated outputs, or team-shared execution memory. Team working files are current Outcome inputs/support material, generated outputs live in Output Files, and team continuity belongs in `AGENT_MEMORY`; Deployment Context is for governed source material you want Soma to reuse later. It is also not the same as Soma reading Mycelis help docs: curated docs lookup is read-only and citable.

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
- Soma and governed teams can recall allowed context during planning and answer generation without treating it as raw unrestricted web input
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

Output Files now starts with a **Group outputs** selector when retained group deliverables exist. The selector only lists groups that produced retained user-facing outputs through `/api/v1/groups/{id}/outputs`; that endpoint includes artifact rows and durable team-work `output_refs` for real deliverables. Groups with no delivered output stay out of the output picker so operators do not have to scan abandoned, planning-only, or internal-only lanes.

Resources treats delivered work and team working material differently. Final documents, packages, media, generated files, and other user-facing deliverables appear in the output picker whether they were stored as artifacts or returned as durable team output refs. Planning files, source/support files, proof notes, and handoffs stay available through the group's Workflow Log or explicit source/internal controls, but they do not make a group appear in the delivered-output picker by themselves.

When a group is cleared from `Groups`, retained files stay visible here unless the operator explicitly included retained output cleanup. Single-group clear and bulk-select clear both preserve files by default; their retained-output cleanup options are deliberate storage cleanup choices. If cleanup included the group outputs, the group workspace folder is removed and those artifact rows are archived, so the cleared group no longer appears in the curated output selector. Transient message-bus handoff data is not part of this retained-file list.

When an output group is selected, Resources links back to the same group in `Groups` for **Outputs**, **Workflow Log**, and **Message** review. Use those links when you need the collaboration history, group conversation, or workflow context behind a file; keep Resources focused on opening and inspecting the generated content itself.

Group workflow logs and chat-pipeline history are reviewed in `Groups -> Workflow Log`, not in `Resources -> Output Files`. Output Files should stay focused on durable artifacts that a user can open, download, preview, or reveal in the workspace. When you need the work history that led to an artifact, open the group and review its Workflow Log.

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
- `Preview` reads the selected generated file without leaving the Resources surface and offers **Ask Soma with this** when you want the selected file to ground the next Soma follow-up.
- `Create` keeps small handoff-folder and handoff-file writes available without making write controls the default browse path.

Output Files should read top-to-bottom: choose retained group output/source scope, optionally open the current folder, then browse, preview, or create from the full-width workspace panel below.

The upper access card includes **Open folder** for the current workspace path. Use it when an operator wants to grab generated files, media proof, project packages, or browser-game output from the local machine without decoding the storage configuration.

Retained output cards in Soma, Teams, and Groups open their primary file in the dedicated Mycelis output canvas and also expose **Open folder** when they carry a workspace path. Use **Back to Soma** to return to the exact conversation and Outcome context. Resources browsing, local folder access, proof, and technical references remain secondary ways to inspect the same retained object.

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
- `Resources -> Output Files` can link the selected group into `Groups` Outputs, Workflow Log, and Message review.
- `Resources -> Output Files` can narrow retained artifacts by contributor level before opening a file or package.
- `Resources -> Output Files -> Include team source files` switches from curated output artifacts to the selected group's workspace source folder.
- `Resources -> Output Files` can browse/read/write only under the governed workspace boundary and can open the current local folder through the workspace-confined reveal endpoint.
- `System -> Deployments` reports the deployment/workspace/artifact roots that explain where generated output will land.
- A retained demo output or project package opened from Soma/Teams/Groups resolves to the same workspace root family instead of a hidden process working directory; team-owned packages and media should be inside the selected group folder unless the operator explicitly chose another workspace path.
- A retained output or project package opened from Soma can be used as the source for the next Soma ask through **Reply**. Use this when the user wants an update, alternate, downstream generation, or another team to react to the delivered content without manually copying paths.
- A selected file in `Resources -> Output Files -> Preview` can be used for the next Soma ask through **Ask Soma with this**. The action opens Soma with the file title and workspace reference attached as one-shot continuation context; it does not approve execution or promote the file into long-term context.
- A saved entry in `Resources -> Deployment Context` can also use **Ask Soma with this** when the next chat turn should reference that durable context source directly. The context remains governed by its saved visibility, sensitivity, trust, and goal-scope metadata.

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

## Worker Profiles
Worker Profiles define reusable teammate types Soma may assign to governed work. The catalogue is a library of inert templates, not a list of running agents. Ready-made profiles are locked so the shipped safety and access posture stays inspectable. Use **Create with Soma** for a conversational draft, **Customize with Soma** from a ready-made profile, or paste/edit a Worker Profile YAML/JSON draft directly in Resources. The inline draft panel previews the exact document through the same ConfigDocument validator used by Soma, then hands the exact draft back to Soma when you choose save or activation.
Worker Profile YAML/JSON preview proves the family fields and exact id/version/digest without saving or activating anything. Activating a stored revision means it may be selected for future teams; activation does not start an agent, connect an agent API, open a provider session, or subscribe to the work bus. When an approved delivery needs that teammate type, team creation resolves the most specific eligible revision for the operator, workspace, organization, then built-in scope, creates the live member, and retains its exact profile/provider/backend/capability lineage. Profile `inputs` and `outputs` describe the teammate's expected brief fields and deliverables; they are not NATS subjects unless an advanced runtime field explicitly names a valid bus subject. Only selected members receive team-scoped runtime connections and subscriptions, which are released when the team stops. Existing teams do not change when a different revision is activated or rolled back. Legacy catalogue records remain API compatibility data and are not runtime activation authority; the primary UI does not present them as assignable profiles.
The natural Soma sequence is: ask to save the Worker Profile YAML or JSON, approve the compact save proposal, ask Soma to activate that profile, then ask Soma to create work using `this Worker Profile` or its name. An inline save, exact activation, or rollback naming a version is governed directly and does not depend on free-form model inference. A question such as `Should I activate this Worker Profile?` remains conversation, and a save request without a document lets Soma ask for or help create the missing content. Team creation starts in the background and leaves the conversation available. To return future teams to an earlier revision, tell Soma the exact version, for example `Roll back this Worker Profile to version alpha.` Existing teams retain the snapshot they started with.
Each profile may define:
- a plain purpose and role
- optional model preference, otherwise the workspace AI engine applies
- capability references such as approved research, file, media, or review access
- context sources and read/search/write posture
- whether Soma, the operator, or policy automation may select it and whether its default scope is Workspace, Outcome, or Team
- expected outputs and verification criteria

Profiles do not grant access by themselves. Runtime capability health, source scope, approval, secret, Outcome, and Execution Contract rules still apply. Ask Soma naturally to use a named profile, such as `Use the Research Specialist and Quality Reviewer`, or omit names and let Soma choose the smallest useful team. Teams receive a resolved profile snapshot at creation so later profile edits do not silently redefine running authority.
## Operational Guidance

Use `Resources` to answer these operator questions quickly:
1. Which AI engines are online?
2. Which tools are accessible right now?
3. How is external or research context classified and reviewed?
4. What deployment knowledge has been intentionally loaded into long-term context?
5. Are workspace file operations available?
6. Which ready-made worker profiles can Soma assign? Custom profiles are currently revisited through Soma or the ConfigDocument API rather than listed in the ready-made library.
