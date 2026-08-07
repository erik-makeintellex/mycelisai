# Core Concepts
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [User Docs Home](README.md)

Mycelis is organized around the work you want and the result you own. Runtime machinery stays behind **Details** or **Inspect** until it is useful.

## Workspace

A Workspace is the user-owned operating context in which you talk with Soma and return to work later. It can hold one Outcome or many Outcomes without changing how you interact.

The default authenticated route is the Soma workspace at `/dashboard`.

## Outcome

An Outcome is the durable product object created when meaningful work begins. It owns:

- deliverables
- active work
- proof
- recovery
- history
- continuity

The operator owns the Outcome. Soma owns the execution needed to produce it.

## Soma

Soma is the persistent operational counterpart for the Workspace. Ask Soma naturally for an answer, comparison, plan, retained output, team effort, scheduled lane, or recovery.

Soma should:

- answer directly when no execution is needed
- ask a useful follow-up when the goal or expected output is unclear
- propose the smallest useful execution shape for meaningful work
- request approval when policy or risk requires it
- keep the conversation available while asynchronous work continues
- report progress, completion, deliverables, proof, and recovery in user language

You do not need to choose agents, tools, or transport before asking for value.

## Adaptive Answer Depth

Soma separates answer depth from execution intent:

| Depth | Typical request | Expected response |
| --- | --- | --- |
| Quick result | "Give me the top five" | Compact list, table, links, or source box |
| Structured summary | "Compare these options" | Grouped findings and relevant differences |
| Decision brief | "What should we do?" | Recommendation, confidence, risks, and guidance |
| Execution proposal | "Create this and have a team review it" | Deliverables, execution shape, approval posture, and proof expectations |

When the request is ambiguous, Soma should use the lightest useful response and offer expansion.

## Governance

Governance protects meaningful or risky actions without turning ordinary conversation into a form.

A proposal should briefly state:

- what Soma understood
- what will be delivered
- which compact team or capability shape is needed
- whether approval is required

Approval creates durable governed work. Canceling or adjusting a proposal must not apply the mutation.

## WorkIntent And Execution

WorkIntent is the transitional record that converts conversation into governed execution. It is inspectable, but it is not another permanent object the user must manage.

The normal transition is:

```text
Conversation
-> WorkIntent
-> governed execution
-> Outcome
```

Execution can be one-shot, scheduled, continuous/service-style, project-shaped, or an extension to the Mycelis environment. The approved work must retain its expected output, stop/retry behavior, and proof requirements.

## Teams And Groups

Soma may use a direct capability, one specialist, or a compact team. Teams are execution mechanisms, not separate assistant identities.

Use **Groups** when you need to:

- inspect active or retained team work
- read the workflow log
- review outputs associated with the producing group
- message or steer an active lane
- archive or clear finished group records through governed controls

Temporary teams should expire or archive when their bounded work ends. Standing teams remain only when their continuing responsibility is useful.

## Deliverables

Deliverables are durable outputs such as documents, code, media, reports, project packages, or validated applications. A deliverable should be attributable to its Outcome and producing team.

Use **Open file**, **Open app**, **Open folder**, or **Open in Resources** when those actions are available. Planning notes and team-internal source material remain hidden unless you deliberately include them.

A package is not complete merely because a file exists. Required entrypoints, dependencies, expected behavior, and validation must pass before Mycelis presents it as verified.

## Outcome Health

Every Outcome uses the same operational vocabulary:

| State | Meaning |
| --- | --- |
| Healthy | Usable and does not need attention |
| Waiting | Needs approval, input, schedule, dependency, or another next action |
| Running | Work is active and can continue safely |
| Degraded | Some work failed or is uncertain, but trusted parts remain |
| Blocked | Soma cannot safely continue until something changes |
| Completed | A deliverable is retained with sufficient proof for revisit |
| Archived | Inactive but retained for history, proof, or outputs |

Health describes the Outcome. Proof remains a separate trust attribute.

## Proof And Recovery

Proof explains why a result can be trusted. Recovery explains how to proceed when it cannot.

The default UI should answer:

- what happened
- what remains trusted
- what is incomplete or uncertain
- which deliverable is available
- what Soma recommends next

Raw event payloads, runtime identifiers, and transport diagnostics belong behind **Details** or **Inspect**.

## Resources And Capabilities

**Resources** is where you revisit outputs and manage what Soma can use.

Key areas include:

- **Output Files** for retained deliverables and workspace folders
- **Capabilities** for readiness, catalog, access scope, servers, and repair
- **Exchange** for governed cross-team handoffs
- **Deployment Context** for approved long-lived source context
- **AI Engines** for model-provider posture
- **Role Library** for reusable worker profiles

A capability may be built in, an MCP server/tool, a local command, or another registered service. The UI should identify its source kind and access scope rather than presenting every item as MCP.

## Memory And Continuity

Memory supports recall. Continuity helps Soma resume long-running operational context. Neither is autonomous authority.

Authority remains with approved Outcomes, deliverables, proof, policies, audit records, and operator decisions. Use **Deployment Context** when a file or note should become approved long-lived source material rather than a one-run attachment.

## Activity And Runs

Most users should follow progress from Soma and the Outcome. **Activity** and **Runs** are Admin tools for deeper inspection, support, and audit.

A run is one execution instance supporting an Outcome. Its receipt summarizes what happened, what to trust, output references, proof, and recovery. The event timeline exists to reconstruct execution, not to replace the deliverable or Outcome view.

## Admin Tools

Turn on **Admin tools** from the left rail when you need Activity, deep Memory, System diagnostics, provider settings, or raw Inspect detail. Keep it off for the simplest Soma-first operating surface.

See also:

- [Using Soma Chat](soma-chat.md)
- [Teams](teams.md)
- [Resources](resources.md)
- [Memory](memory.md)
- [Governance & Trust](governance-trust.md)
- [Run Timeline](run-timeline.md)
