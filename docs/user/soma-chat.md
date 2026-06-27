# Using Soma Chat
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Soma-first interaction model: tell Soma the outcome, review the proposal or answer, open the output, and keep proof/recovery visible.

## Start With Soma

Open `Soma` (`/dashboard`) and type naturally. Soma receives the request first, uses the current organization/workspace context, and should return either a direct answer, a governed proposal, a retained output, or a clear blocker/recovery state.

The dashboard is organized as a threaded workspace:

- `Quick actions`: pinned outcome buttons for repeated work, such as audits, briefs, or media packages. Use `Create action` to save a repeated Soma ask as a reusable action; Mycelis stores it through the conversation-template path when Core is available and keeps a local fallback when it is not. Saved actions still enter the Soma conversation so you can adjust risky or unclear work before execution.
- `Talk to Soma`: the primary conversation where you ask, approve, recover, and review.
- `Outcome Vault`: a secondary overlay drawer for saved results, work in progress, and anything that needs attention. It stays closed by default so Soma keeps the main workspace, then opens over the thread when you need delivery, recovery, or revisit detail.

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
- `Review current MCP servers, tools, and recent use, then tell me which agents should have which tools.`

## Understanding And Approval

Soma should infer the outcome, audience, output form, constraints, source boundary, and uncertainty before answering or proposing work. It should ask one concise clarifying question only when the missing detail would materially change the result.

Before creating teams, enabling MCP servers, assigning tools, changing capability bindings, using private services, or storing recurring behavior, Soma should:

1. Review the latest request and relevant prior context.
2. Name the action it infers.
3. Name missing capability, MCP, credential, or private-data boundaries.
4. Ask whether to proceed once or make the behavior recurring.
5. Use the governed proposal/approval path when mutation or execution is required.

Confirming with `yes`, `confirm`, `proceed`, `do it`, or `one time` should bind to the prior inferred action instead of starting a new unrelated request.

## Reading The Soma Workspace

The dashboard keeps Soma chat primary. A compact current-work lane above chat summarizes:

- current workflow state
- latest retained output
- review count
- unresolved recovery work when output is ready but some proof/work still needs attention
- the next action: `Review output` or `Review work`

Open the review panel when you need more detail. Its tabs keep dense information out of the main chat:

- `Work`: active, queued, degraded, or operator-needed items
- `Output`: retained files, packages, media, and folder actions
- `Trust`: what happened, evidence, run/proof links, and next step
- `Context`: tools, saved context, and setup cues

When output is ready and recovery is also present, Soma should say that plainly, keep the output openable, and point you to the Work tab for recovery.

Soma replies may also show small action-state cards inside the thread. These cards translate structured work state into user language such as `Approval sent`, `Execution started`, `Output ready`, or `Needs recovery` without exposing raw routing subjects or system payloads. When a handoff has a run receipt, the card should offer a plain `Open run receipt` link for proof and recovery review.

## Outputs

Soma responses can include:

1. **Primary answer**: markdown text, code blocks, links, and tables.
2. **Inline generated outputs**: images, audio, video, code, charts, briefs, data, documents, and media previews.
3. **Output package**: a retained file/app/package with `Open file`, `Open folder`, proof, and Resources re-entry.
4. **Proposal block**: a clear `Run this now?` confirmation for actions that execute or change something.
5. **Recovery/blocker card**: what failed, what remains trusted, what is not trusted, and what can safely happen next.
6. **Action-state card**: the current status, route, capability use, or next step for structured Soma work.

No mutation executes until you confirm. Risk, cost, resources, capability details, proof intent, and team/tool wiring should stay behind `Review run details` unless they require immediate attention.

Saved media and file outputs should appear in the same Soma output workbench with the latest output first, plain **Open file** and **Open folder** actions, visible workspace path, and collapsed verification details. Use `Resources -> Output Files` for broader browsing later.

## Teams And Groups

Root Soma is organization-wide. Focused team lanes keep chat, active work, retained outputs, and proof scoped together through the `Working in` picker.

Team defaults:

- generic teams start with one accountable lead
- explicit specialist-output requests may create a bounded specialist roster
- temporary specialists require a named missing capability, owned task, expected proof, and removal point
- broad work should split into smaller lead-owned lanes instead of one large roster

Groups are collaboration lanes. Use `Groups` when you want a temporary or standing lane with one selected-group workspace, Workflow Log, outputs, retained artifacts, and message/review context. The Workflow Log is the readable chat-pipeline view; it should not become raw bus logs or multiple little agent windows.

## Web, Search, And MCP

Soma is expected to execute, not just explain, when execution is available.

Preferred path:

1. Use internal capabilities.
2. Use `web_search` for search intent.
3. Use onboarded MCP tools when they are the shortest safe path, including `fetch` for explicit URLs and `brave-search` when that optional MCP server is installed.
4. Propose/confirm governed mutation paths.

Search behavior:

- if you ask whether Soma can search or make web requests, Soma should answer from current Mycelis Search capability status
- freshness-oriented prompts should call the configured Mycelis `web_search` capability before falling back to MCP-specific guidance
- `local_sources` is the default token-free provider and lets Soma search governed Mycelis context
- if semantic embeddings are unavailable, `local_sources` should fall back to bounded text search over retained Mycelis context
- when Soma uses `web_search`, the Operator trust package should show a source boundary such as `Search source: Local Mycelis context`
- the supported Compose path can use `searxng` with `MYCELIS_SEARXNG_ENDPOINT=http://searxng:8080`
- `local_api` uses `MYCELIS_SEARCH_LOCAL_API_ENDPOINT`
- `brave-search` requires `BRAVE_API_KEY`
- if a needed server or credential is missing, Soma should name the missing MCP server/env var and point you to `Resources -> Capabilities`

Read-only tool posture prompts such as `show me currently configured tools` should answer with current tool state and setup guidance, not create a runnable proposal. Prompts that enable, install, connect, assign, or bind tools remain governed mutation requests.

## Direct Drafting

If you ask for plain chat content such as a short letter, note, email, or message, Soma should answer directly in chat. It should not route that request through file tools, local commands, or council delegation unless you ask to save, inspect, execute, or hand off the work.

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
