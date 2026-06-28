# Mycelis Canonical PRD
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Index](ARCHITECTURE_LIBRARY_INDEX.md)

> Status: Canonical
> Last Updated: 2026-06-28
> Purpose: Single source of product, architecture, UX, runtime, and MVP delivery truth for Mycelis.
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

- compact Quick Actions shelf as a bounded grid for pinned repeatable Soma asks without a visible horizontal scrollbar
- large Talk to Soma thread as the primary canvas
- no separate dashboard headline band above the thread; status and governance live in the Soma header
- quiet current-work strip only when there is meaningful work state
- header Outcomes button that opens Outcome Vault on demand
- no default right rail squeezing Soma
- no setup, identity, topology, or environment stack below the chat
- Details and Inspect controls for depth, not always-visible technical panels

The empty Soma thread should not be a stack of starter action cards. It should behave like the beginning of a conversation: one plain prompt for the outcome, one short cue that Soma can help shape the path, and optional quoted examples of user asks. Those examples are readable language, not buttons, hidden launchers, or a predefined workflow menu.

The dashboard should keep the composer reachable at common desktop, laptop, tablet, and mobile viewports. Long content belongs inside bounded panes, overlays, tabs, or detail drawers rather than growing the whole page. Default work/output summaries should not expose file paths, proof internals, or stacked cards before the operator asks for detail; the primary surface should show the title, safe action, and review entry point.

Quick Actions are saved conversational accelerators, not autonomous triggers. Button Studio should persist reusable Soma asks through the conversation-template path, keep a local fallback only for resilience, and run saved actions by sending the rendered prompt back into the Soma thread so understanding, approval, proof, and recovery stay intact.

## Conversation And Governance

Soma must support natural exploratory conversation before execution. Users can ask questions, refine goals, co-architect requirements, compare options, or shape an idea before asking Soma to run anything. Soma is the persistent operational identity of the workspace: it understands intent, shapes Outcomes, chooses execution strategies, coordinates capabilities, preserves continuity, maintains trust, guides recovery, and keeps long-running context. The user should increasingly think "I work through Soma," not "I operate an AI platform."

Conversational phases:
| Phase | User Experience | Runtime Meaning |
| --- | --- | --- |
| Explore | "Help me think this through." | No execution contract yet; Soma may ask questions, draft options, or reason over allowed context. |
| Shape | "Turn this into a plan." | Soma forms WorkIntent, output shape, constraints, execution mode, and approval posture. |
| Execute | "Run/build/schedule/start this." | Soma creates or uses an ExecutionContract, starts governed work, and updates the thread. |

Governance is mandatory for mutation, durable execution, risky tool use, team/project instantiation, schedules, service mode, and Soma self-extension. The visible governance experience should be a small conversational pause, not a large compliance panel: Soma gives a 1-3 sentence summary, a short bullet list of the intended team/work/output, and one approval choice while NATS/team routing, run proof, and recovery details stay available behind Details.

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

Every user-facing output package should expose:

- clear title and outcome state
- primary open action for the deliverable
- folder or data-root access where safe
- conversational reply action that returns the output reference to Soma for updates, alternates, downstream generation, or team handoff
- proof or receipt link
- recovery state if degraded
- source/intermediate-output visibility only as an opt-in

Outcome Vault is the persistent delivery/revisit concept, but it should open as an overlay by default. It should show saved results, work in progress, scheduled/service work, and recovery items without permanently taking layout width from Soma.

## Projects Teams And Capability Use

Soma is singular. Organizations, groups, projects, deployments, teams, tools, and outputs are scoped operational contexts, not separate assistant identities.

Complex work may create an OutcomeProject and a TeamRegistryEntry. The user should see the outcome and team purpose, not a pile of agent internals. Minimal teams are preferred. Temporary teams should expire or archive unless they produce durable user-facing outputs. Operators may explicitly name a team, but when they do not, Soma must infer the expected readable team name from intent, such as Temporary Game Delivery Team, Standing Content Steward Team, Media Generation Team, or Mixed Output Team. Generated team ids must mirror that purpose with obvious prefixes such as `temp-`, `standing-`, `game-delivery-team`, `media-generation-team`, or `mixed-output-team`, ending in a short uuid suffix rather than timestamp-like numbers. Team-created intermediate files belong in source/support folders and should be hidden from user output lists unless the user chooses to include team-source outputs. When Soma coordinates builder, watcher, and transaction teams, source content and requested output targets must remain distinct: a request to watch or react to one folder/file while saving another file must retain the saved target as the user-facing output.

Hard domain tests, such as asking Soma to generate a substantial action game, media package, commercial data tool, deployable app, or mixed text/media/code output, are probes of generic complex-output orchestration. Mycelis must not become a game engine, media engine, website generator, or code framework by hardcoding one domain's implementation path. Instead, complex work should trigger a team-evocation phase that is never terminal by itself: Soma researches available external/current context or local sources when possible, exposes search/tool boundaries when research is unavailable, consults council/review before implementation, retains a research/council handoff, delegates implementation to the evoked delivery team, chooses the output format and stack that fit the operator's deployment target, and defines the smallest useful lead/specialist team with proof gates. Team plans must carry a content contract for the requested work type. Game-like probes require playable controls, rendered loop, collision/bounds, objective, win/fail, restart, appropriate audio when requested, direct launch/view access, and headed play-through proof; media requires prompt/constraint fit, saved artifact, provider boundary, and review notes; text requires requested structure, readable claims/assumptions, reopenable file output, and proof. If validation finds a defect in a requested output, the repair request should be reported back through Soma so Soma can preserve continuity, coordinate the producing team or a follow-up team, and keep the fix visible in the Outcome history instead of silently editing the artifact outside the conversation. Repeated temporary teams must not collapse into an older group lane merely because they share a display name; output ownership follows the actual team id and Outcome. Teams are autonomous execution mechanisms, never sovereign authorities: authority flows Operator -> Policy -> Outcome -> ExecutionContract -> Team -> Capability.

Cross-functional delivery must also be generic. A game team handing evidence to a marketing team is only one stress case. The same pattern applies when an app delivery team hands usage proof to launch marketing, a media team hands asset examples to a campaign team, a data team hands validation notes to an analyst, or a documentation team hands release facts to support. Soma should coordinate source and downstream teams through the Outcome: the source team improves or produces the deliverable, retains proof examples in its group workspace, writes a concise handoff, and notifies the downstream team. The downstream team must ground claims, campaign copy, review notes, support instructions, or follow-up work in that retained evidence rather than inventing unsupported assertions. User-facing output lists should show the final deliverables by Outcome and producing team, with source/intermediate evidence available through an opt-in detail path.

Capabilities are governed runtime objects. MCP servers, local scripts, APIs, filesystem access, media engines, search, and generated app builders should be presented as capabilities Soma can use. The user-facing question is "What can Soma use, and what needs repair?" not "Which server topology is installed?"

Default capability and service-inventory answers must use user language. When a user asks `list of services?`, `what can Soma use?`, or similar, Soma should summarize available workspace services, AI engine posture, storage/output access, team coordination, memory/context, status checks, and any repair-needed capability. It should not list raw internal tool names, MCP server status strings, subjects, IDs, or topology unless the user explicitly asks for a technical inventory such as `show internal tool names` or `debug MCP status`.

Capability configuration must support three scopes:

- available to all Soma work
- grouped for a capability set or environment
- targeted to a specific host/provider/tool endpoint

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
| Files | Workspace root | Store generated content under governed workspace/project/group folders. |
| Capabilities | Core manifests | Register risk, permission, scope, output types, and recovery behavior. |
| UI state | Interface | Render typed events as rich cards, not raw logs or stack traces. |

Execution modes must distinguish one-shot tasks, scheduled tasks, long-running service/watch tasks, multi-team project delivery, and Soma self-extension. Each mode needs stop/pause/retry/recover semantics. Continuity vectors are not Soma's mind, source of truth, or autonomous authority; they preserve long-running operational context while authority remains with approved Outcomes, deliverables, proof artifacts, run receipts, policies, audit, and operator decisions.

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

Default navigation:

- Soma: ask, shape, approve, execute, review, recover, revisit
- Groups: focused collaboration lanes and retained team outputs
- Resources: generated outputs, workspace folders, capabilities, and connected tools
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

| Priority | Slice | Status | Acceptance |
| --- | --- | --- | --- |
| P0.1 | Threaded Soma dashboard | IN_REVIEW | Compact quick actions, primary chat, no default Vault rail, no setup stack, reachable composer. |
| P0.2 | Natural governance cards | IN_REVIEW | Proposal/running/done/blocked cards are small, conversational, proof-linked, and keep deep recovery detail behind explicit review. |
| P0.3 | WorkIntent and ExecutionMode | ACTIVE | One-shot, scheduled, service, project, and self-extension modes have typed contracts. |
| P0.4 | Bus handoff and started feedback | IN_REVIEW | Approval or quick action immediately creates visible started state with correlation. |
| P0.5 | OutcomeProject and TeamRegistry | IN_REVIEW | Confirmed work writes durable project/team ownership and Vault summaries. |
| P0.6 | Output packages and Vault | IN_REVIEW | Deliverables open cleanly; source/intermediate outputs are opt-in. |
| P0.7 | Capability settings | NEXT | Capabilities can be all-work, grouped, or targeted-host scoped with repair paths. |
| P0.8 | Run receipts and recovery | IN_REVIEW | Receipts explain outcome, proof, failure, trusted state, and next safe action. |
| P0.9 | Full journey proof | IN_REVIEW | Headed and headless proof cover ask through revisit. |
| P0.10 | Docs cleanup | ACTIVE | This PRD is the single architecture authority; stale docs are deleted. |

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
