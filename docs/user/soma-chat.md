# Using Soma Chat
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Workspace-first interaction model: you send intent, Soma orchestrates execution.

---

## Overview

Open `Workspace` (`/dashboard`) and type naturally.
Soma receives every message and coordinates the rest.
Soma operates as a symbiote execution partner: it should execute and deliver outcomes,
not instruct you step-by-step on how to do the work manually.
Workspace should open with Soma already active, so you should not need to manually switch to Soma just to begin a normal session.

Display-name customization:
- open `Settings -> Profile`
- set **Assistant Name**
- save to update Workspace/status labels that normally show "Soma"

```
You type -> Soma reasons (ReAct, up to 10 iterations)
         -> optional council consultation
         -> answer and/or governed proposal
```

---

## Sending Messages

1. Go to `Workspace` (`/dashboard`).
2. Enter your intent in the bottom input.
3. Press `Enter` or click send.

Concrete central Soma phrases:

- `Search the web for the latest changes in self-hosted AI agent products, summarize the top findings, and cite the sources you used.`
- `Create a small temporary team to review the release-readiness risks, assign the right specialists, and bring the retained output back here for approval.`
- `Ask the active delivery teams for current blockers, compare their answers, and tell me which workflow needs attention first.`
- `Use the host data under workspace/shared-sources to answer this question, and tell me which files or records shaped the answer.`
- `Review the current MCP tool structure, tell me which tools are available to Soma and teams, and recommend what should be connected next.`
- `Review my latest request, match it to related prior commands and tool metadata, tell me the action you infer, and ask me to confirm before you execute.`
- `Create the smallest useful team for this outcome, have Council help choose the specialists and outputs, include target MCP tools, and walk me through any missing MCP enablement.`

Live activity text indicates current steps (thinking, consulting, searching memory, invoking tools).

## Referential Review And Confirmation

Soma should treat action-oriented requests as contextual, not isolated.

Before creating teams, enabling MCP servers, assigning tools, or changing capability bindings, Soma should:

1. review the latest user request
2. match related prior conversation turns and command/action metadata
3. review registered MCP servers, available tools, and relevant installable library entries
4. state the action it infers
5. ask whether to proceed once or make the behavior recurring for that workflow

When you confirm with wording such as `yes`, `confirm`, `proceed`, `do it`, or `one time`, Soma should bind that confirmation to the prior inferred action and continue through the governed proposal/execution path.

For team manifestation, Soma should use Council or specialist context to choose the smallest useful member set, expected outputs, and target MCP/tool bindings. If a needed MCP server is missing, Soma should name the server and required `.env` variable, then point you to `Resources -> Connected Tools` to install, reapply, or enable it.

## Protected Interaction Templates

Some words should map to protected Soma behavior instead of ordinary chat.

The basic theme layer recognizes common operator language:

- team words: `create team`, `specialists`, `members`, `lanes`
- tool words: `MCP`, `tools`, `web search`, `GitHub`, `fetch`, `host data`
- private-service words: `private service`, `production service`, `private API`, `token`, `credential`
- private-data words: `customer data`, `deployment context`, `company knowledge`, `sensitive`, `confidential`
- recurring words: `always`, `every time`, `standing behavior`, `template`, `reuse`

The template layer combines those themes with protection rules. Examples:

- compact team with target tools
- private service or credentialed action
- private data review
- reusable protected interaction
- referential review before action

Soma may explain these requests directly, but it must not create teams, bind tools, use credentials/private data, or store recurring behavior until it has confirmed the inferred action and the relevant proposal/approval path allows it.

Good protected prompts:

- `Use the private service only after you name the target service, credential boundary, allowed action, and approval needed.`
- `Review customer data from deployment context, tell me the visibility scope, and ask before retaining any output.`
- `Make this a reusable conversation template only after you summarize the trigger phrase, protected scope, and approval posture.`

AI Organization home adds a guided Soma entrypoint:
- type a team or delivery request and choose `Start team design`, or
- leave the field blank and use `Run a quick strategy check` to trigger an immediate first-pass review
- ask Soma to create or reshape teams from the root organization workspace before dropping into a narrower lane
- use `AI Engine Settings` from the same root workspace when an admin needs to set one shared model for everyone or detected output-type models for planning, research, code, and vision work

If you leave the organization workspace and come back later, the current guided Soma draft and the last successful guidance for that organization should still be there.

Team workspaces are different from the root Soma workspace:
- the root workspace is Soma-first and organization-wide
- the root Soma home also carries a live interaction stream so an admin can review active team output without leaving the main chat surface
- the root Soma chat should stay in a bounded panel and scroll message history internally, so long conversations do not push the live stream and review panels farther down the page
- that stream can be filtered by multiple teams and by available output aspects such as status, results, artifacts, tools, governance, and errors
- the `Teams` page is where admins review available teams, open focused lead workspaces, and define the reusable member templates Soma should apply when creating new team members
- the dedicated guided team-creation workflow now lives at `/teams/create`, so detailed team creation is a step-by-step Soma lane instead of a dense roster-side form
- when Soma returns a native execution path there, the same lane can now launch a temporary workflow group directly, hand you into `Groups`, and continue through archive/closure while keeping retained outputs reviewable
- team defaults should start from 3 precise roles: Team Lead, Architect Prime, and the focused builder/developer role needed for the requested output
- a single team should stay at 5 members or fewer; if more roles seem useful, Soma should split the work into several small teams or lanes coordinated by Soma and Council over NATS and managed exchange
- entering a created team should center the team's focused lead entity first
- that team lead can still coordinate back through Soma using scoped memory, RAG retrieval, and broader organization context when needed
- the team's lead and specialists inherit the organization output-model policy unless an admin changes it

Groups are a separate workflow surface:
- use `Groups` when you want a temporary or standing collaboration lane without cluttering the root Soma home
- the Groups screen uses a compact list/detail layout: select a group from the left rail, then use the main pane for lane data, broadcast/review, outputs, and retained artifacts
- each group should expose the focused lead context, runtime/config posture such as work mode, approval policy, capabilities, and current model-inheritance status, plus output/contributing-lead summaries
- Soma remains the root reviewer that can summarize groups, pull forward outputs, and route you into the right team lead when needed

---

## Output Model Routing

Admins can configure how output models are assigned inside an AI Organization.

Available modes:
- `One model for everyone`: all team members use the same default model
- `Detected by output type`: team leads and specialists inherit the best-fit model for the kind of work they are doing

Current self-hosted starting points shown in product:
- `Qwen3 8B`
- `Llama 3.1 8B`
- `Qwen3 14B` when installed and latency is acceptable
- `Qwen2.5 Coder 14B` or `DeepSeek Coder V2 16B` for heavier code / website generation lanes
- `BGE-M3` and `nomic-embed-text` remain retrieval/embedding candidates, not chat-output models

Current detected output-type defaults:
- general text -> `Qwen3 8B`
- research and reasoning -> `Llama 3.1 8B`
- code generation -> `Qwen2.5 Coder 7B`
- vision analysis -> `LLaVA 7B`

When an admin has not pinned a specific output type, Soma should choose from installed self-hosted models using explicit criteria: match the detected output type first, prefer higher-capacity local models when latency and memory allow, and be honest about engine boundaries. Ollama vision/text models can help plan, code, critique, or review media, but pixel or voice generation still requires the configured media engine.

This routing is durable organization policy, so ordinary user chats should not silently rewrite it. Soma should ask the owner/admin before reviewing potential model behavior for a requested output or changing saved routing.

---

## Reading Responses

Soma responses can include:

1. **Primary answer**
- markdown text, code blocks, links, and tables

2. **Inline generated outputs**
- images, audio, video, code, charts, briefs, data payloads, and documents can appear directly in the same Soma conversation turn
- these outputs may be generated by Soma directly or by a specialist/council path Soma consulted for you
- the operator should not need to leave the conversation or inspect team-delivery lanes just to review the generated result
- non-binary outputs should render directly in chat whenever possible
- binary or saved outputs should show a visible saved path or download link in the same conversation turn

3. **Delegation Trace**
- compact cards showing which council members were consulted

4. **Proposal block (mutation paths)**
- explicit action preview with confirm/cancel, including whether the task runs once, is scheduled, remains active as monitoring, or connects to a current-team/multi-team NATS lane

No mutation executes until you confirm.

6. **Memory-backed continuity**
- central Soma chat uses a scoped session id so backend conversation turns can be persisted and replayed for that same workspace/session when available
- this session continuity is separate from semantic memory promotion: it helps Soma keep a conversation thread coherent, but it does not automatically promote raw interaction content into durable Soma memory or reflection memory
- durable memory stays available for scoped recall when it has been intentionally promoted
- ordinary draft planning and return-visit continuity can stay useful without automatically becoming long-term semantic memory

5. **Inline image outputs**
- if a response generates an image, it is rendered directly in chat
- generated images are cache-first and expire after 60 minutes unless saved
- use the inline `Save` action or ask Soma to save it (for example: "save this image to saved-media")

---

## Execution-First Contract

Soma is expected to execute, not just explain, whenever execution is available.

Preferred path:
1. use internal capabilities
2. use internal capabilities such as `web_search` for search intent, then onboarded MCP tools when they are the shortest safe path, including `fetch` for explicit URL retrieval and `brave-search` when that optional MCP server is installed
3. propose/confirm for governed mutation paths

If a tool call fails, Soma should recover inline (retry, reroute, or proposal fallback) without making you retype the request.

Web/research behavior:
- if you ask Soma whether it can search or make web requests, Soma should answer from current Mycelis Search capability status instead of falling back to provider boilerplate
- if you ask a freshness-oriented question such as latest news, current updates, or recent releases, Soma should call the configured Mycelis `web_search` capability directly before falling back to MCP-specific guidance
- `local_sources` lets Soma search governed user-shared and deployment context without hosted search tokens
- the supported Compose release path starts a self-hosted SearXNG service by default, so `MYCELIS_SEARCH_PROVIDER=searxng` and `MYCELIS_SEARXNG_ENDPOINT=http://searxng:8080` let Soma use public web search without Brave tokens
- `searxng` can also point at another operator-owned endpoint when configured through `MYCELIS_SEARXNG_ENDPOINT`
- `local_api` lets Soma use an operator-owned HTTP search endpoint when configured through `MYCELIS_SEARCH_PROVIDER=local_api` and `MYCELIS_SEARCH_LOCAL_API_ENDPOINT`
- if `brave-search` is installed and configured with `BRAVE_API_KEY`, Soma and web-capable specialists may use it for search
- if `fetch` is installed and you provide a URL, Soma and web-capable specialists may use it to retrieve page content
- if the needed server or credential is missing, Soma should name the missing MCP server/env var and point you to Connected Tools instead of claiming web requests are impossible

Direct drafting behavior:
- if you ask for plain chat content such as a short letter, email, note, or message,
  Soma should answer with the text directly in chat
- it should not route that request through file tools, local commands, or council delegation
  unless you explicitly ask to save, inspect, execute, or hand the work off
- if you ask `what is your current state` or `what teams currently exist`,
  Soma should answer from current runtime and team state directly instead of falling back to a generic provider apology

Execution guardrail:
- if Soma responds with planning-only language (for example "Step 1" / "we need to delegate") on an actionable request,
  the runtime triggers one policy-correction pass to force a tool call or a concrete blocker response.

Root-admin configuration behavior:
- if you ask Soma to configure Mycelis, it should execute against the relevant configuration surface
  (brains/providers, profiles, governance policy, MCP, users/groups, runtime settings) rather than
  limiting itself to "create a new team" flows.
- governed mutations still use proposal/confirm gates where required.

---

## Council Failure Recovery

If a council call fails, Workspace shows a structured error card instead of a raw error.
The same rule applies to the central Soma path:
- no raw `500` strings as the visible answer
- no raw `tool_call` JSON
- no raw structured runtime envelopes in the main conversation body

The card includes:
- what failed
- likely cause
- next actions

Available actions:
- `Retry`
- `Switch to Soma`
- `Continue with Soma Only`
- `Copy Diagnostics`

This keeps recovery inline without retyping or page switching.
`Switch to Soma` is specifically a recovery path for when you were using a direct specialist route and want to return to the default orchestration path.

---

## Direct Council Access

To send directly to a specialist:

1. Click `Direct` in chat header.
2. Pick Architect, Coder, Creative, or Sentry.
3. Send your message.
4. Use `Soma` option to return to default orchestration.

---

## Launch Crew Flow

For multi-step execution:

1. Open the AI Organization workspace you want Soma to coordinate.
2. Click `Create teams with Soma`.
3. Click `Open crew launcher`.
4. Provide mission intent.
5. Review the generated proposal or blocker recovery guidance.
6. Confirm execution when Soma returns a proposal.

On success, a system message includes a run link (`/runs/{run_id}`).

When crew creation is needed, Soma should start from a lean 3-role shape:

1. Team Lead: owns the operator-facing state, handoffs, and final output summary.
2. Architect Prime: shapes the plan, dependencies, and acceptance criteria.
3. Focused Builder: produces the requested artifact, implementation, media direction, website draft, data review, or delivery payload.

Soma can add a 4th or 5th role only when the output needs a separate reviewer/tester, domain specialist, or second focused builder. If the work needs more than 5 people, Soma should split the request into multiple compact lanes and explain what each lane will produce.

Good examples:
- "Create the smallest useful team to turn this customer brief into a one-page proposal and keep outputs visible in chat."
- "This is broad: split the product review into planning, build, and review lanes, with each lane capped at a compact team."
- "Show me which team lead owns each output and where I can review the retained artifacts."

When specialized output is needed, Soma should prefer to:
1. plan the need at the root workspace
2. shape the right team or specialist lane
3. let the team's lead inherit the configured output model policy
4. only override delivery routing through governed admin configuration, not ad hoc user chat

---

## Operational Helpers

While chatting, you can use:
- **Status Drawer** (global health visibility)
- **Degraded Mode Banner** actions
- **Focus Mode** (`F`) to prioritize chat height
- **Advanced Mode** toggle (Settings footer) to show/hide high-density telemetry surfaces

---

## Good Prompting Practices

- be explicit about desired outputs
- reference recent context ("continue from step 2")
- review delegation trace to understand specialist contributions
- confirm only when proposal intent matches your goal
