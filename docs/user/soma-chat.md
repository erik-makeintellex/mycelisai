# Using Soma Chat
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Soma-first interaction model: tell Soma the outcome, review the proposal or answer, open the output, and keep proof/recovery visible.

## Start With Soma

Open `Soma` (`/dashboard`) and type naturally. Soma receives the request first, uses the current organization/workspace context, and should return either a direct answer, a governed proposal, a retained output, or a clear blocker/recovery state.

Opening `/dashboard?fresh=1` starts a fresh conversation presentation; it does not erase or conceal retained Outcomes. The default Outcomes attention count is bounded to operator decisions, recovery, and ready deliverables. Ordinary queued/running work remains available through its Outcome or team progress view without appearing as review debt.

The dashboard is organized as a threaded workspace:

- `Talk to Soma`: the primary visible heading and conversation where you ask, approve, recover, and review.
- `Outcome Vault`: a secondary overlay drawer for saved results, work in progress, and anything that needs attention. It stays closed by default so Soma keeps the main workspace, then opens over the thread when you need delivery, recovery, or revisit detail. On a phone it becomes a full-width sheet; on larger screens it stays a bounded right-side drawer without narrowing the conversation. Close it with its visible control, the shaded backdrop, or Escape. Keyboard focus stays inside while it is open and returns to `Outcomes` when it closes.

When the conversation is empty, Soma should help you enter naturally instead of presenting a stack of action cards. The empty thread should briefly cue the pattern: ask for the outcome, let Soma shape the path, and approve only when work should run. Example asks are shown as quoted language only; they are not buttons or a separate workflow menu.

The default dashboard hides engine trace details for ordinary answers. Source/model badges, tool chips, consultation traces, and raw capability labels belong in advanced views, proof/review panels, Activity, or Inspect. The chat thread should show the answer first, then surface compact proposal, blocker, receipt, or recovery cards only when they change what the user can safely do next.

The dashboard should not require you to scroll through setup panels before using Soma. Sign-in, role, provider, and scope details are available through Settings/System or proof details when you need to inspect them.

Basic path:

```text
You ask -> Soma understands -> optional proposal -> execution -> output/proof/recovery -> revisit
```

Display-name customization lives in `Settings -> Profile -> Assistant Name`.

## Good First Prompts

- `Give me a short readiness check for this Mycelis environment.`
- `Create a retained demo package I can open in the browser and include proof I can revisit.`
- `Search the web for the latest changes in self-hosted AI agent products, summarize the top findings, and cite the sources you used.`
- `Create the smallest useful team for this outcome and bring the retained output back here.`
- `Ask the active delivery teams for current blockers and tell me which workflow needs attention first.`
- `Use the host data under workspace/shared-sources and tell me which files shaped the answer.`
- `Use the Mycelis docs to explain how Deployment Context differs from Memory, and cite the docs you used.`
- `Tell me what Soma can currently use, what needs attention, and what I should connect next.`

## Understanding And Approval

Soma should infer the outcome, audience, output form, constraints, source boundary, and uncertainty before answering or proposing work. It should ask one concise clarifying question only when the missing detail would materially change the result.

Soma should choose the lightest useful answer depth for the ask. A request for a quick list, links, a table, a summary, or a recommendation changes how much detail Soma returns; it does not mean work should run. If you want more, ask Soma to expand, turn the answer into a brief, or turn it into governed work.

Lightweight answers stay inside the normal conversation. A quick table, summary, or decision brief should not add approval buttons, tool-chip stacks, or run receipts unless Soma is actually proposing or reporting work. Ask `turn this into work` when you want Soma to move from answer to execution.

When work will run once, on a schedule, as a continuing service, as a project, or as an extension to Soma, open `Details` to check how it can be stopped, retried, or recovered. These controls remain attached to the approved work after handoff and reload. They stay out of the default approval pause so the conversation remains readable.

Before creating teams, enabling MCP servers, assigning tools, changing capability bindings, using private services, or storing recurring behavior, Soma should:

1. Review the latest request and relevant prior context.
2. Name the action it infers.
3. Name missing capability, MCP, credential, or private-data boundaries.
4. Ask whether to proceed once or make the behavior recurring.
5. Use the governed proposal/approval path when mutation or execution is required.

Confirming with `yes`, `confirm`, `proceed`, `do it`, or `one time` should bind to the prior inferred action instead of starting a new unrelated request.

## Reading The Soma Workspace

The dashboard keeps Soma chat primary. When meaningful work exists, one compact background-work indicator shows the Outcome health and whether anything needs you. Opening it reveals a bounded work list over the conversation; it never narrows Soma or turns the page into a dashboard.

File paths and detailed proof stay out of the default strip unless needed. Open the review panel when you need more detail. Its tabs keep dense information out of the main chat:

- `Work`: active, queued, degraded, or operator-needed items
- `Output`: retained files, packages, media, and folder actions
- `Trust`: what happened, evidence, run/proof links, and next step
- `Context`: tools, saved context, and setup cues

The review panel overlays the conversation instead of resizing it. It is for locating and triaging work, not for operating a substantial deliverable. On compact screens it uses an opaque full-width sheet so background text cannot compete with review content. Exact package-reference duplicates must not appear as a second output, and supporting files remain available under Details or Inspect.

When a document, application, report, dataset, media item, or package becomes the object of attention, **Open output** promotes it into a dedicated Mycelis surface with enough room to read, operate, compare, or validate it instead of exposing a raw workspace-file endpoint. Completion shows one short summary and one primary **Open output** action. Folder access, Resources, proof, versions, download, and technical references are secondary. **Back to Soma** returns to the originating conversation, query, and Outcome context.

When output is ready and recovery is also present, Soma should say that plainly, keep the output openable, and point you to the Work tab for recovery.

Soma replies may also show one small action-state card inside the thread. Consecutive machine updates for the same work collapse into the latest state, while the full event history remains available through proof and inspection. The card uses user language such as `Work started`, `Output ready`, or `Soma needs your direction`; technical provider, tool, status, and routing detail stays under `What happened`. The composer is the continuation path, so status events do not add repeated continuation buttons. `Work started` means the approved plan is durably queued or running; it is not completion proof. When work stops, Soma explains that no usable result was produced and suggests plain conversational choices such as trying again, using another available service, or changing the request.

## Outputs

Soma responses can include:

1. **Primary answer**: markdown text, code blocks, links, and tables. Table-like data should render as a real table, not as pasted aligned text. Compact labels such as `Quick answer`, `Summary`, or `Decision brief` may appear only to clarify answer depth.
2. **Inline generated outputs**: images, audio, video, code, charts, briefs, data, documents, and media previews.
3. **Output package**: a retained file/app/package with `Open file`, `Open folder`, proof, and Resources re-entry.
4. **Proposal quote**: a compact summary and short work list for actions that execute or change something. Reply `approve`, `go ahead`, or `start` in the normal composer to begin; reply `cancel` to cancel; otherwise tell Soma what to change.
5. **Recovery/blocker card**: a compact trust boundary in the thread, with what failed, what remains trusted, what is not trusted, and what can safely happen next behind `Details and proof`.
6. **Action-state card**: the current status, route, capability use, or next step for structured Soma work.

No mutation executes until you confirm. Only a bounded approval reply resolves the pending proposal. A qualified reply such as `approve after changing the title` remains ordinary conversation so Soma can revise it safely. Opening `Details`, asking for more explanation, or requesting a deeper brief is not approval. Risk, cost, resources, capability details, proof intent, and team/tool wiring should stay behind `Details` unless they require immediate attention.

Saved media and file outputs should appear in the same Soma output workbench with the latest output first, plain **Open file** and **Open folder** actions, visible workspace path, and collapsed verification details. Use `Resources -> Output Files` for broader browsing later.

When Soma is planning or reporting work, the visible plan should name the expected output shape first: table/report, app/package, code/script, media, document, dataset, or mixed output. App/package work should include a direct open path, usage notes, validation status, folder access, proof, and a way to ask Soma for follow-up changes without forcing the operator to read internal team/tool topology.

Proposal details may show an `Expected output` cue. This is not a separate action to approve; it is Soma's contract for what the team or capability must bring back, such as an app/package with an openable entrypoint, a table, a document, media, code, or a dataset. The same expected output follows the approval into the run receipt and team handoff, so the work can be reviewed later against what Soma said it would deliver. Outcome-language requests for a retained app, package, executable, or playable multi-file product should produce a bounded delivery-team proposal without requiring you to name the team yourself; a request for one exact file remains direct. When an asynchronous team returns an interactive package, Soma first reports that the retained candidate is being checked and stays available for conversation while Core validates it. `Work complete` with **Open app** or **Open output** appears only after the exact retained package passes its approved browser workflow. A missing attachment, entrypoint, dependency, non-responsive control, page error, failed local asset, or unavailable validator becomes recovery and asks the same team to repair or regenerate the deliverable. Packages live under the producing team's `groups/<team-id>/generated/...` folder. The main approval card stays short, while output shape, launch hint, validation, bus/team wiring, and proof expectations stay behind `Details`.

Trusted compact receipts in the Soma thread expose the output title, Outcome health, a short completion summary, and one primary **Open output** action. Use Details for folder access, Resources, proof, versions, download, and technical references. Use the normal conversation to ask for changes; Soma keeps the delivered output and Outcome identity attached to that follow-up.

Use **Reply** on a delivered output or project package when you want Soma to keep that exact output as context for the next request. Reply does not execute work by itself. It keeps the output visible, shows a compact `Continuing from` indicator, leaves the composer ready for your natural follow-up, and sends typed continuation context with the output title, workspace reference, and proof id when available. You can ask to update it, make an alternate version, generate downstream material from it, inspect it, or route it to another team; Soma classifies that follow-up separately from whether approval is required. If a file should become reusable long-term source material rather than just a one-request handoff, ask Soma to save it as governed context or use `Resources -> Deployment Context`.

Resources can start the same one-shot continuation flow. Use **Ask Soma with this** from `Resources -> Output Files -> Preview` when a selected workspace file should ground the next message, or from a saved Deployment Context entry when durable governed context should be referenced in the next turn. Soma opens with a compact `Continuing from` indicator and waits for your natural instruction; the handoff is context only, not approval to mutate, run, or promote data.

## Teams And Groups

Root Soma is organization-wide. When a Groups or Outcome link opens a focused conversation, the header quietly names that context. Ask Soma to change, combine, or leave that context; team selection and routing are not required controls in the default conversation.

Team defaults:

- generic teams start with one accountable lead
- explicit specialist-output requests may create a bounded specialist roster
- temporary specialists require a named missing capability, owned task, expected proof, and removal point
- broad work should split into smaller lead-owned lanes instead of one large roster

Groups are collaboration lanes. Use `Groups` when you want a temporary or standing lane with one selected-group workspace, Workflow Log, outputs, retained artifacts, and message/review context. The Workflow Log is the readable chat-pipeline view; it should not become raw bus logs or multiple little agent windows.

When Soma evokes teams, planning and handoff files are working context, not final delivery. Ask Soma to continue until the group has a real user-facing output, then use Workflow Log or internal/source review only when you need the evidence trail behind that result. For long-running teams, tell Soma which files or mounted sources are standing context, which folder the team may watch or reuse, and which folder/file should count as the delivered output.

For multi-team work, ask in outcome language rather than agent wiring. For example: `Improve this app, retain examples of what changed, then let marketing create launch copy from those examples.` Soma should keep the source team, evidence examples, handoff, and downstream team output connected to the same Outcome. This applies across apps, games, media, documents, reports, and data products; games are only a difficult proof case, not a special product mode.

## Web, Search, And MCP

Soma is expected to execute, not just explain, when execution is available.

Preferred path:

1. Use internal capabilities.
2. Use built-in Mycelis `web_search` for search intent when configured.
3. Use onboarded MCP tools when they are the shortest safe path, including `fetch` for explicit user-supplied URLs and `brave-search` when that optional MCP server is installed.
4. Propose/confirm governed mutation paths.

Search behavior:

- if you ask whether Soma can search or make web requests, Soma should answer from current Mycelis Search capability status
- freshness-oriented prompts should call the configured Mycelis `web_search` capability before falling back to MCP-specific guidance
- `builtin_web` is the default token-free provider for ordinary web search
- `local_sources` lets Soma search governed Mycelis context when the user asks for local, retained, internal, mounted, or shared sources
- mixed local-plus-web prompts should search both approved local/mounted sources and the configured public-web path when both are available; if only one boundary is available, Soma should keep the useful result but show a coverage warning
- if semantic embeddings are unavailable, `local_sources` should fall back to bounded text search over retained Mycelis context
- when Soma uses `web_search`, the Operator trust package should show a source boundary such as `Search source: Public web` for public research or `Search source: Local Mycelis context` for retained/local data
- the supported Compose path can use `searxng` with `MYCELIS_SEARXNG_ENDPOINT=http://searxng:8080`
- `local_api` uses `MYCELIS_SEARCH_LOCAL_API_ENDPOINT`
- `brave-search` requires `BRAVE_API_KEY`
- `fetch` is useful for retrieving a specific URL the user supplies, but it is not required for built-in Mycelis `web_search`
- if a needed search provider, MCP server, or credential is missing, Soma should name the missing provider/server/env var and point you to `Resources -> Capabilities -> Access` for web/search sources or the relevant capability lane for other tools

Read-only tool posture prompts such as `show me currently configured tools` should answer with current tool state and setup guidance, not create a runnable proposal. Prompts that enable, install, connect, assign, or bind tools remain governed mutation requests.

Plain service-inventory prompts such as `list of services?` should answer in user language: Soma workspace, AI engine, storage, team coordination, memory/context, status checks, and any repair-needed capability. Raw internal tool names, MCP server status, and tool IDs should appear only when you explicitly ask for technical inventory, such as `show internal tool names` or `debug MCP status`.

## Docs, Context, And Memory

Soma can read a curated set of Mycelis help, API, testing, and architecture docs through read-only documentation tools. This lets Soma answer questions like `which docs explain web access setup?` or `use the canonical PRD to explain the current Outcome model` with slug/path citations instead of guessing.

Docs access is not the same as memory. Reading a doc does not promote it into durable memory, team-shared execution memory, or governed context. It is a citable help/source lookup.

Use these boundaries:

- ask Soma to use Mycelis docs when you want documentation explained or cited
- use Reply or a team handoff when a file should guide this next draft or follow-up only
- use `Resources -> Output Files -> Preview -> Ask Soma with this` when you found a workspace file outside the current thread and want it attached to the next Soma turn
- use `Resources -> Deployment Context` when you want private, customer, company, or operating-context material to influence future Soma reasoning under a governed trust boundary
- use Memory when you want to inspect retained recall, prior facts, SitReps, or continuity records already stored by the system
- ask Soma to name which docs or context sources shaped an answer when trust matters

## Direct Drafting

If you ask for plain chat content such as a short letter, note, email, or message, Soma should answer directly in chat. It should not route that request through file tools, local commands, or team delegation unless you ask to save, inspect, execute, or hand off the work.

If you ask `what is your current state` or `what teams currently exist`, Soma should answer from current runtime and team state rather than giving a generic provider apology.

## Recovery

If execution fails, Soma should recover inline without making you retype the request. Recovery cards should avoid raw `500`, raw `tool_call` JSON, and raw runtime envelopes in the main conversation.

Useful actions:

- `Retry`
- `Switch to Soma`, when returning from a direct specialist route
- `Continue with Soma Only`
- `Copy Diagnostics`
- `Clear from review`, when stale work should leave active queues while retaining history

## Operational Helpers

- `Resources -> Capabilities`: what Soma can use, repair, or request
- `Resources -> Output Files`: retained generated content and workspace folders
- `Groups`: collaboration lanes, workflow logs, and group outputs
- `Teams`: active work, team lead workspaces, and reusable member templates
- `System -> Deployments`: runtime/workspace/artifact roots
- `Advanced Mode`: high-density admin/telemetry routes when needed

## Good Prompting Practices

- name the desired output
- say whether you want a direct answer, retained file/package, or team output
- confirm only when the proposal intent matches your goal
- ask Soma to include proof and recovery notes for deliverables you will revisit later
