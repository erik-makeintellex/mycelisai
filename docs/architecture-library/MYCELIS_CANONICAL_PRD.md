# Mycelis Canonical PRD
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Index](ARCHITECTURE_LIBRARY_INDEX.md)
> Status: Canonical | Last Updated: 2026-07-13 | Purpose: Single source of product, architecture, UX, runtime, and MVP delivery truth for Mycelis.
## Product Thesis
Mycelis is a Soma-centered governed cognitive operating environment. It is not an agent console, chatbot shell, MCP registry, or workflow dashboard. The product value is that a person can talk with Soma, shape meaningful work, approve governed execution, receive durable outputs, inspect proof, recover failures, and revisit the outcome later without learning infrastructure vocabulary. The prime architecture rule is twofold: every decision must be technically correct and must make the system easier to trust without exposing unnecessary complexity.
The default product language is:
```text
I tell Soma what I want.
Soma helps shape it.
Soma directs the work safely.
I can see what happened.
I can trust or recover the result later.
```
The architecture exists to protect confidence while making complexity disappear. Runs, agents, groups, capabilities, tools, continuity vectors, NATS, schemas, receipts, audit, and deployment topology serve Soma and the Outcome; they are supporting machinery, not default user concepts. When the architecture succeeds, users do not admire the runtime. They trust Soma.

## Release Goal
The V8.3 release target is operational embodiment: prove Mycelis through visible execution, durable deliverables, recoverable work, understandable trust, and clean deployment reality. The risk is no longer insufficient architecture. The risk is doctrine expansion without product proof.

Release success means a non-technical user can complete the journey from ask to trusted revisit without needing to understand agents, MCP, workflows, runs, topology, or infrastructure. A technical user can still inspect proof and runtime detail when needed. If runtime correctness improves but user trust or usability declines, the architecture moved in the wrong direction.

Architecture stability rule: the architecture is stable, not frozen. Future work should extend, refine, harden, and express the current Outcome/Soma/execution-spine model instead of replacing it or creating parallel concepts. Outcome is the organizing identity across Soma conversations, deliverables, active work, recovery, proof, history, and continuity; the existing Soma, Groups, Resources, Runs, Recovery, and Administration surfaces should become progressively Outcome-aware rather than Outcome-replaced. Runtime machinery translates machine work into user work, so ExecutionContracts, runs, teams, capabilities, MCP, NATS, vector retrieval, storage, event routing, autonomy, adaptive teams, providers, and new capability types should compose through the existing spine and remain primarily behind Details or Inspect unless surfacing them directly improves Outcome, Deliverable, Trust, Recovery, or Continuity value.
## Trusted Outcome Journey
All P0 work is judged through this journey:

```text
Ask
-> Understand
-> Approve
-> Execute
-> Deliver
-> Trust
-> Recover
-> Revisit
```

Subsystems matter only when they strengthen the journey. Output Packages strengthen Deliver. Run Receipts strengthen Trust. Recovery Queue strengthens Recover. Review Inbox strengthens Understand and Approve. Capability Catalog strengthens Trust. Resources, Groups, and Vault strengthen Revisit.

The release question is not "Did we finish the subsystem?" The release question is "Can the user complete the journey and trust the result later?" Every screen should strengthen one journey step; every subsystem should exist only because it improves this journey.
## Primary User Experience

The first authenticated surface is the Soma workspace. It should feel like a focused threaded workspace, not a dense admin console.

Required first-viewport composition:
- compact Quick Actions shelf as a bounded row/wrap of pinned repeatable Soma asks without a visible horizontal scrollbar
- large Talk to Soma thread as the primary canvas
- no separate dashboard headline band above the thread; status and governance live in the Soma header
- focused work/team context inside the Soma header when needed, not as another full-width pre-chat panel
- quiet current-work strip only when there is meaningful work state
- header Outcomes button that opens Outcome Vault on demand
- no default right rail squeezing Soma
- no setup, identity, topology, or environment stack below the chat
- Details and Inspect controls for depth, not always-visible technical panels

The empty Soma thread should not be a stack of starter action cards. It should behave like the beginning of a conversation: one plain prompt for the outcome, one short cue that Soma can help shape the path, and optional quoted examples of user asks. Those examples are readable language in a compact quote strip, not buttons, hidden launchers, framed cards, or a predefined workflow menu.

The dashboard should keep the composer reachable at common desktop, laptop, tablet, and mobile viewports. The left navigation rail must be collapsible on desktop so Soma has more horizontal workspace when the operator is reading outputs, tables, or generated plans. Long content belongs inside bounded panes, overlays, tabs, or detail drawers rather than growing the whole page. Default work/output summaries should not expose file paths, proof internals, or stacked cards before the operator asks for detail; the primary surface should show the title, safe action, and review entry point.

Quick Actions are saved conversational accelerators, not autonomous triggers. Button Studio should persist reusable Soma asks through the conversation-template path, keep a local fallback only for resilience, and run saved actions by sending the rendered prompt back into the Soma thread so understanding, approval, proof, and recovery stay intact.

## Conversation And Governance

Soma must support natural exploratory conversation before execution. Users can ask questions, refine goals, co-architect requirements, compare options, or shape an idea before asking Soma to run anything. Soma is the persistent operational identity of the workspace: it understands intent, shapes Outcomes, chooses execution strategies, coordinates capabilities, preserves continuity, maintains trust, guides recovery, and keeps long-running context. The user should increasingly think "I work through Soma," not "I operate an AI platform."
Soma must infer response depth separately from execution intent. A user asking for a quick list, source box, compact table, summary, comparison, or recommendation is choosing answer density, not approving work. The canonical depths are `quick_box`, `structured_summary`, `decision_brief`, and `execution_proposal`. Soma should default to the lightest useful response and offer expansion rather than turning every request into a decision brief or work proposal. The Soma thread consumes this as presentation guidance: answer depths stay conversational with compact markdown/table rendering and no approval controls, neutral status cards, tool-chip stacks, or execution receipts unless a real execution result, retained output, blocker, or recovery state exists. `execution_proposal` is not enough by itself to create approval controls; approval controls appear only when the runtime returns a governed proposal.
Conversational phases:
| Phase | User Experience | Runtime Meaning |
| --- | --- | --- |
| Explore | "Help me think this through." | No execution contract yet; Soma may ask questions, draft options, or reason over allowed context. |
| Shape | "Turn this into a plan." | Soma forms WorkIntent, output shape, constraints, execution mode, and approval posture. |
| Execute | "Run/build/schedule/start this." | Soma creates or uses an ExecutionContract, starts governed work, and updates the thread. |

Governance is mandatory for mutation, durable execution, risky tool use, team/project instantiation, schedules, service mode, and Soma self-extension regardless of requested response depth. A deeper explanation, expanded details, or decision brief never grants execution authority. The visible governance experience should be a small conversational pause, not a large compliance panel: Soma gives a 1-3 sentence summary, a short bullet list of the intended team/work/output, and one approval choice while NATS/team routing, run proof, and recovery details stay available behind Details.

Default approval frame:

```text
Soma
I can start that.
Approve this?
I will hand this to the work bus after approval and keep this thread open.

- Shape the project workspace.
- Hand the work to the right team.
- Save the deliverable to Outcomes.

[Approve] [Adjust] [Details]
```

After approval, Soma should immediately acknowledge the handoff:

```text
Started.
I handed this to the work bus and saved the receipt.
You can keep talking here while updates arrive.
```

## Outcome Vault

The defining product abstraction is the Outcome. Deliverables, projects, operations, proof, recovery, history, continuity, and active lanes belong to Outcomes. Runs, teams, capabilities, transport, storage, and event correlation support Outcomes. The Outcome never serves the runtime. Outputs are durable product objects, not transient chat text. They may include apps, files, plans, reports, media, reviews, proof bundles, deployment results, or retained learning candidates.

Every user-facing output package should expose clear title/state, primary open action, safe folder or data-root access, conversational reply action for updates/alternates/downstream team handoff, proof or receipt link, degraded recovery state, and source/intermediate-output visibility only as an opt-in. File/content handoff has three product meanings: a current Outcome input for this team or draft, a delivered output to retain and revisit, or a long-lived context source Soma may use again under scope/trust rules. Soma should name which meaning it inferred before governed execution, and Output/Reply must carry typed continuation context so updates, forks, downstream routes, and ordinary follow-ups preserve Outcome/source/proof identity without relying only on natural-language quotes.

Output ownership must distinguish final user deliverables from team working material. The canonical output classes are `user_deliverable`, `planning`, `internal_handoff`, `proof`, and `source_material`. Outcome Vault, Resources output pickers, and the default Groups Outputs tab show `user_deliverable` records by default. Planning files such as `TEAM_EVOCATION.md`, council/research handoffs, proof notes, support files, and source records remain inspectable through Workflow Log, Details, or an explicit include-internal control, but they must not make a group appear complete or delivered on their own.

Soma output plans must name the expected output shape before details: table/report, document, app/package, code/script, media, dataset, or mixed output. The visible plan should explain the minimum support/proof needed for that shape in user language. Table-like information should render as actual compact tables in the Soma thread, not as monospaced aligned text.

Outcome Vault is the persistent delivery/revisit concept, but it should open as an overlay by default. It should show saved results, work in progress, scheduled/service work, and recovery items without permanently taking layout width from Soma.

## Projects Teams And Capability Use

Soma is singular. Organizations, groups, projects, deployments, teams, tools, and outputs are scoped operational contexts, not separate assistant identities.

Complex work may create an OutcomeProject and a TeamRegistryEntry. The user should see the outcome and team purpose, not a pile of agent internals. Minimal teams are preferred. Temporary teams should expire or archive unless they produce durable user-facing outputs. Operators may explicitly name a team, but when they do not, Soma must infer the expected readable team name from intent, such as Temporary Game Delivery Team, Standing Content Steward Team, Media Generation Team, or Mixed Output Team. Generated team ids must mirror that purpose with obvious prefixes such as `temp-`, `standing-`, `game-delivery-team`, `media-generation-team`, or `mixed-output-team`, ending in a short uuid suffix rather than timestamp-like numbers. Team-created intermediate files belong in source/support folders and should be hidden from user output lists unless the user chooses to include team-source outputs. When Soma coordinates builder, watcher, and transaction teams, source content and requested output targets must remain distinct: a request to watch or react to one folder/file while saving another file must retain the saved target as the user-facing output. Long-running teams must declare which context sources they may read, which generated folders they own, and which outputs are final deliverables versus working material.

Hard domain tests, such as asking Soma to generate a substantial action game, media package, commercial data tool, deployable app, or mixed text/media/code output, are probes of generic complex-output orchestration. Mycelis must not become a game engine, media engine, website generator, or code framework by hardcoding one domain's implementation path. Instead, complex work should trigger a team-evocation phase that is never terminal by itself: Soma researches available external/current context or local sources when possible, exposes search/tool boundaries when research is unavailable, consults council/review before implementation, retains a research/council handoff, delegates implementation to the evoked delivery team, chooses the output format and stack that fit the operator's deployment target, and defines the smallest useful lead/specialist team with proof gates. Team plans must carry a content contract for the requested work type. Application-package probes require a direct launch/open path, primary workflow validation, data/table usability when relevant, usage notes, runtime/browser error review, and a follow-up path through Soma; game-like probes add playable controls, rendered loop, collision/bounds, objective, win/fail, restart, audio when requested, and headed play-through proof; media requires prompt/constraint fit, saved artifact, provider boundary, and review notes; text requires requested structure, readable claims/assumptions, reopenable file output, and proof. If validation finds a defect in a requested output, the repair request should be reported back through Soma so Soma can preserve continuity, coordinate the producing team or a follow-up team, and keep the fix visible in the Outcome history instead of silently editing the artifact outside the conversation. Repeated temporary teams must not collapse into an older group lane merely because they share a display name; output ownership follows the actual team id and Outcome. Teams are autonomous execution mechanisms, never sovereign authorities: authority flows Operator -> Policy -> Outcome -> ExecutionContract -> Team -> Capability.

Cross-functional delivery must also be generic. A game team handing evidence to a marketing team is only one stress case. The same pattern applies when an app delivery team hands usage proof to launch marketing, a media team hands asset examples to a campaign team, a data team hands validation notes to an analyst, or a documentation team hands release facts to support. Soma should coordinate source and downstream teams through the Outcome: the source team improves or produces the deliverable, retains proof examples in its group workspace, writes a concise handoff, and notifies the downstream team. The downstream team must ground claims, campaign copy, review notes, support instructions, or follow-up work in that retained evidence rather than inventing unsupported assertions. User-facing output lists should show the final deliverables by Outcome and producing team, with source/intermediate evidence available through an opt-in detail path.

Capabilities are governed runtime objects. MCP servers, local scripts, APIs, filesystem access, user-owned data mounts, infrastructure shared folders, media engines, search, and generated app builders should be presented as capabilities Soma can use. The user-facing question is "What can Soma use, and what needs repair?" not "Which server topology is installed?"

Default capability and service-inventory answers must use user language. When a user asks `list of services?`, `what can Soma use?`, or similar, Soma should summarize available workspace services, AI engine posture, storage/output access, team coordination, memory/context, status checks, and any repair-needed capability. It should not list raw internal tool names, MCP server status strings, subjects, IDs, or topology unless the user explicitly asks for a technical inventory such as `show internal tool names` or `debug MCP status`.

Capability configuration must support three scopes:

- available to all Soma work
- grouped for a capability set or environment
- targeted to a specific host/provider/tool endpoint

The default configuration path should start from common capability choices rather than raw tool references. Operators should be able to choose readable intents such as Workspace files, User data mounts, Web research, Team coordination, or Local host/media, then set whether the permission applies to everyone, one Outcome/group lane, or one host. Raw MCP/tool refs remain visible for inspection and advanced editing, but they are not the first thing a user must understand.

Built-in Soma commands are runtime capabilities, not hidden prompt affordances. Their handler execution stays in Go, but their user-facing title, quoted ask language, scope, governance posture, delivery shape, proof expectation, recovery posture, and UI label must come from command manifests under `core/config/soma-commands/`. The runtime must validate those manifests against the registered handler set so new or removed built-ins cannot silently drift away from user-facing capability language. Capability and service-inventory surfaces should use this metadata first and expose raw handler names only in Inspect or explicit technical inventory requests.

Search is a governed capability family, not only public web search. Mycelis must support a Search Source Registry where operators can add sources Soma may search when allowed: built-in Mycelis/local search; public web through self-hosted SearXNG, operator-owned local APIs, or optional hosted providers; explicit URL retrieval through governed fetch; authenticated URL/API sources such as docs sites, customer portals, SaaS knowledge bases, issue trackers, repositories, file stores, mounted user folders, infrastructure shared folders, or internal search endpoints; and future dedicated connectors such as GitHub, Slack, Notion, Confluence, SharePoint, Google Drive, Postgres, CRM, accounting, or ticketing systems.

Authenticated sources must use secret references, not raw tokens in UI, logs, state files, docs, or capability manifests. The first supported shape is bearer/API-token env references for local API search sources, applied as an Authorization bearer header only after scope/status checks pass. Service-required query/header token placement, OAuth2/client-credential metadata, and non-env secret backends remain follow-on adapters. Each source needs a plain name, source type, endpoint, domain/path boundary, auth scheme, secret reference, scope (`Everyone`, `Group`, or `Host`), sensitivity/trust defaults, index/live-search mode, and recovery posture. Soma should name configured sources used, cite or reference them in the trust package, and ask for approval before searching sensitive/private sources when policy requires it. Search scope is part of trust: ordinary search phrasing should use the configured web path by default, local/internal/shared asks request local-source scope, and mixed local-plus-web requests should search both approved local/mounted sources and public web when both boundaries are available. Partial mixed coverage must state which boundary was searched and which boundary is missing or partial. Local-source search must never be presented as public research.

Reticulum is an install-time and future capability substrate, not a default user concept. The supported install now provides `rns` import access and verifies `uvx --from rns rnstatus --help`; next Reticulum capabilities should be prioritized as status/health (`rnstatus`/daemon), LXMF-compatible messaging and offline notifications, retained-output/file sync, Nomad/RNS page publishing and browsing, then governed remote shell/admin, monitoring, and low-priority voice/live chat; each must enter Resources as a governed capability with scope, proof, recovery, and clear boundaries before Soma can use it.

## Runtime Architecture

The canonical execution spine is:
```text
Intent
-> Soma understanding
-> WorkIntent
-> ExecutionMode
-> ExecutionContract
-> Governed Run or Bus Handoff
-> Capability or Team Invocation
-> Output Package
-> Proof / Recovery / Revisit
```

Core runtime responsibilities:

| Area | Owner | Requirement |
| --- | --- | --- |
| Conversation | Interface + Core | Preserve natural thread state, compact cards, and typed results. |
| Governance | Core | Persist proposal, approval, execution contract, and audit boundary. |
| Work handoff | Core + NATS | Correlate run, work item, team, project, capability, and output references. |
| Persistence | PostgreSQL | Store projects, teams, work items, interactions, runs, receipts, outputs, and recovery. |
| Files | Workspace root and named mounts | Store generated content under governed workspace/project/group folders; read user-owned or infrastructure shared folders only through named mount records with scope, boundary, mode, and proof; distinguish one-run source files, retained deliverables, and durable context sources. |
| Capabilities | Core manifests | Register risk, permission, scope, output types, and recovery behavior. |
| UI state | Interface | Render typed events as rich cards, not raw logs or stack traces. |

Execution modes must distinguish one-shot tasks, scheduled tasks, long-running service/watch tasks, multi-team project delivery, and Soma self-extension. Each mode needs stop/pause/retry/recover semantics. The worker library is the focused execution package that lets approved agentry decisions become executed work without adding a new orchestration layer. Its default backend is `central`: Mycelis-owned workers, credentials, policies, logs, usage accounting, and operational responsibility. Hermes-compatible execution is additive through an adapter backend (`hermes_api` or compatible `hermes_like`) selected by configuration after Mycelis policy permits delegation. The normalized lifecycle is create run, stream/poll events, request approval if needed, continue or deny, complete/fail/cancel, then retain result, usage, and audit. Continuity vectors are not Soma's mind, source of truth, or autonomous authority; they preserve long-running operational context while authority remains with approved Outcomes, deliverables, proof artifacts, run receipts, policies, audit, and operator decisions. Soma may read a curated Core-owned documentation surface for help, API, testing, and architecture answers; this access is read-only, citable by slug/path, distinct from memory, Deployment Context, and vector continuity, and never promotes docs into durable context or doctrine without a governed Resources/Deployment Context action.

## API And Event Contracts

The API should normalize all user-facing results into standard envelopes with status, output references, proof references, recovery hints, and target references. UI code should not infer trust from raw text.

Required references on meaningful work:

- `run_id` when execution-linked
- `team_id` when team-scoped
- `project_id` or outcome reference for durable output ownership
- `work_item_id` for review/recovery queues
- `capability_id` or capability name when tools are used
- `source_kind`, `source_channel`, `payload_kind`, and timestamp for bus payloads

NATS, the current EventSource stream, or a future WebSocket bridge should produce typed thread events such as started, progress, proposal, needs approval, output ready, blocked, recovered, and archived. The UI must render an immediate "Approval sent" or "Started" card after a quick action or approval instead of waiting for completion. These in-thread cards should read as compact conversational annotations, not diagnostic panels; they should avoid duplicating the plain event text when structured event data already exists. Only normalized `thread_event` payloads belong in the Soma thread by default; raw NATS subjects, runtime envelopes, and stack traces remain behind Inspect/Details.

## Settings And Capability Configuration

Settings should feel like application permissions and deployment readiness, not topology management.

Required settings model:

- profile, auth, role, and access posture in Settings/System, not on the default dashboard
- AI engine readiness with simple connected/degraded/needs setup states
- capability cards such as "Connect filesystem", "Connect search", "Connect media engine", or "Connect accounting software"
- grouped capability sets for environments
- targeted host/provider configuration for specific MCP or API endpoints
- searchable data-source and mount cards such as "Connect company docs", "Connect GitHub repository", "Connect customer portal", "Connect internal search API", or "Add local data folder"
- raw server auth, vector index, topology, and schema details behind Inspect

## Trust Recovery And Confidence

Failure is normal. The system must answer:

- what failed
- what remains trusted
- what proof is invalid
- what can continue safely
- what requires retry
- what requires operator attention
- what uncertainty is exposed

No raw backend stack traces should reach the default UI. Backend failures, MCP timeouts, provider outages, malformed outputs, and unavailable tools should render as Operational Alert Cards with plain choices such as Retry, Adjust, Connect, Skip, Keep partial result, or Open details.

Confidence provenance is an emerging layer. The architecture should prepare for validation source, evidence strength, cross-model agreement, review lineage, and proof quality without overbuilding scores before the MVP journey works.

## Information Architecture

Preserve the existing navigation and make it progressively Outcome-aware:

- Soma: ask, shape, approve, execute, review, recover, and revisit Outcome work
- Groups: focused collaboration lanes, active work, and retained team outputs
- Resources: generated deliverables, workspace folders, capabilities, and connected tools
- Docs: user help and contributor docs

Admin/deep navigation:

- Activity/Runs
- Memory
- System
- Settings
- Inspect/details surfaces

Documentation and UI should use tabs, list/detail layouts, overlays, and bounded panes for deep content. Avoid page-length stacks of unrelated cards.

## MVP Scope

MVP is complete when one canonical workflow feels excellent:

```text
User asks Soma to create or review meaningful work
-> Soma explores and shapes the request
-> Soma proposes execution mode and expected outcome
-> user approves when execution is meaningful
-> Soma starts work and shows visible handoff
-> owned work creates durable output
-> output appears in thread and Vault/Resources
-> proof and recovery are visible on demand
-> user can return later and trust what happened
```

Non-goals for MVP:

- marketplace abstraction
- broad recursive autonomy
- user-facing topology management
- multiple assistant identities
- raw MCP server administration as the main experience
- architecture docs proliferating into separate doctrine systems

## P0 Delivery Plan
| Priority | Slice | Status | Journey Step | Acceptance |
| --- | --- | --- | --- | --- |
| P0.1 | Threaded Soma workspace | IN_REVIEW | Ask, Understand | Compact quick asks, primary chat, no default Vault rail, no setup stack, collapsible rail, reachable expanding composer, and headed route sweep without console/page errors. |
| P0.2 | Natural governance cards | IN_REVIEW | Approve, Trust | Proposal/running/done/blocked cards stay small and conversational, show one obvious next action, and keep risk/proof/recovery detail behind Details or Inspect. |
| P0.3 | WorkIntent and ExecutionMode | ACTIVE | Understand, Execute, Trust | One-shot, scheduled, service/watch, project, and Soma self-extension modes have typed contracts, stop/retry/recover semantics, output-shape expectations, validation/launch hints, and approval posture without expanding the default approval card. Confirmed runs and durable team-work refs must retain the same output contract so proof/recovery can compare delivered work to the approved expectation, including treating deliverable completion without retained output refs as recoverable rather than trusted output. |
| P0.4 | Bus handoff and started feedback | IN_REVIEW | Execute, Recover, Trust | Approval or quick action immediately creates visible started state, correlation, durable work linkage, expected-output/proof context in run-linked events, proof-linked retained output refs when async team results arrive, and actionable degradation when NATS/team dispatch or retained-output completion cannot complete. |
| P0.5 | OutcomeProject and TeamRegistry | IN_REVIEW | Deliver, Revisit | Confirmed work writes durable project/team ownership, Vault summaries, target refs, and producing-team identity without exposing agent internals by default. |
| P0.6 | Output packages, Resources, and Vault | IN_REVIEW | Deliver, Revisit | Deliverables open cleanly from Soma, Groups, Resources, and Vault. Artifact rows and durable team `output_refs` both appear as retained deliverables; planning/source/internal material remains opt-in; selected files/context can re-enter Soma as one-shot continuation context. |
| P0.7 | Capability settings | IN_REVIEW | Trust, Recover | Capabilities default to a readiness view that says what Soma can use, what needs repair, and what can be requested; all-work, grouped, targeted-host scopes, inspectable refs, raw bindings, and examples stay behind deliberate configure/inspect controls. |
| P0.7a | Search and data-source registry | IN_REVIEW | Ask, Trust | Search status and Resources show public web, approved local/mounted data, private/API sources, and named mounts; persisted sources carry endpoint/path, scope, boundary, auth or mount mode, secret-ref when needed, sensitivity, trust, and recovery metadata; mixed local-plus-web asks state coverage clearly. |
| P0.8 | Run receipts and recovery | IN_REVIEW | Trust, Recover | Receipts explain outcome, proof, failure, trusted state, invalidated proof, next safe action, and recovery/clear controls without leading with raw logs. |
| P0.9 | Full journey proof | IN_REVIEW | Ask through Revisit | Headed and headless proof cover ask, understand, approve, execute, deliver, trust, recover, and revisit; latest GUI sweep covers primary/admin routes and critical workflows with no console/page errors or horizontal overflow. |
| P0.10 | Worker execution library | ACTIVE | Execute, Deliver | Agentry can call one worker interface while central execution remains default and Hermes-compatible execution stays adapter-based with retained outputs, usage, audit, and recovery. |
| P0.11 | Docs cleanup and release discipline | ACTIVE | Trust, Revisit | This PRD remains the single architecture authority, old doctrine is deleted instead of archived into active docs, user help matches current UI, and implementation slices update docs/state in the same change. |
| P0.12 | Release hygiene and promotion proof | NEXT | Trust | Keep main/dev branches clean, commit coherent tested slices, run release proof from committed state, and verify PostgreSQL/NATS-backed runtime paths rather than stale local/test-only state. |

## Testing And Release Gates

Visible UI changes require both functional tests and live user-experience review. The reviewer must inspect layout density, scroll behavior, text-field reachability, panel overlap, card size, plain-language copy, and whether the screen matches the target Soma workspace concept.

Required proof lanes:

- unit tests for typed state, cards, projections, and API adapters
- Go tests for runtime, persistence, governance, and event correlation
- docs tests for live links and canonical PRD coverage
- Playwright headless proof for repeatability
- headed browser proof for actual user experience
- release preflight from a clean committed state before production deployment

## Documentation Contract

This PRD is the canonical architecture/product document. Keep support docs, but do not recreate split doctrine.

Allowed supporting docs:

- `README.md` for repo entry, command contract, and contributor navigation
- `.state/V8_DEV_STATE.md` for active implementation state
- `docs/README.md` for docs navigation
- `docs/user/*` for operator help
- `docs/API_REFERENCE.md` for API behavior
- `docs/TESTING.md` for validation
- `docs/architecture/OPERATIONS.md`, `BACKEND.md`, `FRONTEND.md`, and `OVERVIEW.md` for implementation support
- `ops/README.md`, `core/README.md`, and `interface/README.md` for owned subsystem operations

Removed architecture details must be promoted here if still current. Otherwise they should be deleted and left to Git history.
