# Mycelis V8.3 - Soma User Experience Contract
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Index](ARCHITECTURE_LIBRARY_INDEX.md)

Status: ACTIVE  
Last Updated: 2026-06-14  
Owner: Mycelis Product Experience  
Scope: Normal operator experience with Soma-centered work, deliverables, proof, recovery, and optional advanced Inspect

## Purpose

This contract defines what a normal Mycelis operator should expect when working with Soma.

The operator is not expected to think like an architect, developer, AI engineer, workflow builder, model selector, agent manager, or tool registry maintainer. The operator should experience Mycelis as a trustworthy work counterpart:

```text
I tell Soma what I need.
Soma figures out the safe way to do it.
Soma asks for approval when appropriate.
Soma coordinates the work.
Soma delivers usable results.
Soma shows proof.
Soma helps me recover when something goes wrong.
```

The central user relationship is:

```text
User <-> Soma
```

It is not:

```text
User <-> Agents / Tools / MCP / Workflows / Models
```

The runtime system may involve teams, specialists, capabilities, tools, memory, policies, and infrastructure. Those details must appear through Soma in language and controls the operator can understand.

## Product Truth

Soma is the primary interface, continuity layer, work coordinator, execution planner, trust mediator, and recovery guide for the operator experience.

Soma is not a chatbot, search box, workflow canvas, model picker, tool registry, MCP console, or agent topology dashboard.

The user experience must make the operator feel like they are working with Soma to get meaningful work done safely. It must not make them feel like they are operating an AI system.

## User Mental Model

The default mental model is:

```text
Tell Soma what I need.
Soma understands the request.
Soma explains the plan.
Soma asks me only for meaningful choices or approvals.
Soma performs or coordinates the work.
Soma gives me a deliverable package.
Soma shows what happened, what is trusted, and what needs attention.
```

Every primary surface should reinforce this model.

## First Experience

The first meaningful prompt should be:

```text
What do you want Soma to do?
```

The first experience must not begin by asking the user to configure MCP servers, select agents, manage teams, choose models, build automation flows, or understand backend topology.

Good first-use request examples include:

- research competitors
- create a proposal
- build a browser game
- generate marketing assets
- analyze documents
- prepare a business report
- review architecture
- summarize customer feedback
- draft an implementation plan

The operator may later inspect or tune advanced runtime behavior, but the first path must begin with user intent and Soma's understanding of the work.

## Soma Understanding and Planning

After the user asks for work, Soma must show:

- what Soma understood
- what Soma plans to do
- what deliverable or outcome will be produced
- whether approval is needed
- what user input would improve the result
- what proof will be available
- what recovery path exists if execution fails or is incomplete

This can be concise. The goal is confidence, not process ceremony.

## Work States

The operator-facing work states are:

| State | Meaning |
| --- | --- |
| Running | Soma or a coordinated execution lane is actively working. |
| Waiting | Work is queued, paused, rate-limited, or awaiting an external condition. |
| Needs Review | A result or intermediate output is ready for user review. |
| Needs Approval | Work cannot safely proceed until the user approves a proposed action. |
| Completed | A usable result has been produced and retained. |
| Failed | Work could not complete as requested and requires recovery, retry, or acceptance of a partial result. |

The user cares about:

- what is being worked on
- what is finished
- what needs attention
- what can be opened, previewed, downloaded, trusted, retried, or inspected

They should not need to read raw logs, NATS traffic, MCP messages, model traces, or specialist chatter to answer those questions.

## Deliverables First

Deliverables are the primary product outcome. A response is not enough when the user asked for meaningful work.

Soma may answer simple questions directly, but substantive work requests must produce a retained deliverable package or an explicit recovery state explaining why no complete package exists.

A deliverable package should include:

- the usable output
- a short summary
- proof of what happened
- known limits or uncertainty
- recommended next actions
- recovery options when the result is partial or failed

## Deliverable Package Types

The UI and runtime must support these package types as first-class operator outcomes:

| Package type | Typical contents |
| --- | --- |
| Research Package | findings, sources, comparisons, confidence notes, gaps, and next research steps |
| Business Report Package | executive summary, analysis, tables or exhibits, recommendations, assumptions, and proof |
| Proposal Package | scope, value narrative, pricing or effort assumptions, timeline, risks, and exportable document |
| Browser Game Package | playable artifact, source files, preview link, controls, acceptance notes, and build or runtime proof |
| Code Package | changed files, implementation summary, test results, risks, and review notes |
| Media Package | generated or edited assets, prompts or source references when appropriate, previews, variants, and usage notes |
| Document Analysis Package | extracted findings, citations or locations, conflicts, open questions, and retained source references |
| Architecture Review Package | findings, affected areas, recommendations, risks, proof reviewed, and follow-up actions |
| Operations Package | runbook update, action log, current state, verification result, rollback or recovery path |

Package labels may be adapted to the user's language, but the underlying contract remains: the user receives a retained result with proof and recovery context.

## Trust and Proof

Every meaningful deliverable must help the user answer:

- what happened
- how it was done
- what evidence supports the result
- what is uncertain
- what was not completed
- what changed
- what can be trusted now
- what the next action should be

The default output actions are:

| Action | Purpose |
| --- | --- |
| Open | Enter the retained output or work surface. |
| Preview | Quickly inspect the result without leaving context. |
| Download | Export the result or artifact package. |
| Proof | See the evidence, checks, sources, logs summary, or verification trail. |
| Inspect | Open advanced execution detail when needed. |

Proof should be understandable before it is exhaustive. It may summarize detailed traces while preserving access to deeper inspection.

## Recovery

Recovery is part of the product promise, not an afterthought.

When work fails, partially succeeds, or produces uncertain results, Soma must explain:

- what failed or remained incomplete
- what still succeeded
- what output remains usable
- what proof is available
- what is not trusted
- whether retry is safe
- what alternative path Soma recommends

Recovery options should be plain and actionable:

- retry the same work
- retry with adjusted scope
- continue from the partial result
- ask the user for missing input
- fall back to another capability
- stop and preserve the proof

The user should never be stranded with raw backend failure text as the primary product response.

## Teams, Tools, and Capabilities

Teams and specialists may be shown when they clarify the work. They must not become the default user relationship.

Preferred operator language:

- Soma is coordinating research.
- Soma is preparing the report.
- Soma is asking for approval before changing files.
- Soma used search, files, and code capabilities.
- Soma needs review because one source conflicted with another.

Avoid making the operator manage internal structures unless they explicitly choose to inspect or configure them.

Capabilities may be described as:

- Files
- Search
- Research
- Media
- Code
- Business Apps
- Databases

Raw MCP servers, tool schemas, model routing, topology, bus subjects, traces, and specialist internals belong behind Inspect or advanced settings.

## Advanced Inspect

Inspect is optional and advanced. It exists for users who want deeper accountability, debugging, governance, or architecture visibility.

Inspect may show:

- runs
- steps
- tools
- teams
- capabilities
- logs
- traces
- approvals
- topology
- retained proof

Inspect must not be required for normal trust. A user should be able to understand the outcome, proof, and recovery path from the primary Soma surface.

## Continuity

Soma should preserve useful continuity across work:

- memory
- context
- preferences
- prior decisions
- retained outputs
- approval history
- recovery history

Continuity must not become hidden autonomy. Soma may remember and apply context, but significant actions still require clear user understanding and approval when appropriate.

The user should feel that Soma remembers the working relationship, while still making significant actions visible and understandable.

## UI Acceptance Standard

A Soma-centered UI slice is acceptable only when a normal operator can:

1. State what they need in ordinary work language.
2. See what Soma understood and plans to do.
3. Understand whether approval, review, or more input is needed.
4. Track work through Running, Waiting, Needs Review, Needs Approval, Completed, or Failed.
5. Open or preview the resulting deliverable package.
6. See proof, uncertainty, and next actions without opening raw logs.
7. Recover from failure or partial completion through clear choices.
8. Use Inspect only when they intentionally want advanced execution detail.

The slice is not acceptable if the operator must choose a model, wire a tool, interpret backend topology, read raw traces, or manually coordinate agents before receiving useful work.

## Final Standard

The product succeeds when the user can say:

```text
Soma understood what I needed, did meaningful work safely, delivered something usable, showed proof, and helped me recover where needed.
```

That is the user experience contract for Mycelis V8.3.
