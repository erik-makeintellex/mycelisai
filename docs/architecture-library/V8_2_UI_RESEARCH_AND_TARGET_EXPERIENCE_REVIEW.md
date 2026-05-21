# V8.2 UI Research And Target Experience Review
> Navigation: [Project README](../../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md) | [Soma UI Architecture Expression](V8_2_SOMA_UI_ARCHITECTURE_EXPRESSION.md)

> Status: canonical UX review
> Last Updated: 2026-05-20
> Purpose: translate current human-AI UX research and enterprise workflow references into the target Mycelis operator experience.

## Summary

Mycelis should not feel like a chatbot, agent dashboard, event log, or topology browser. The target experience is a Soma-centered workbench where the operator expresses a desired outcome, Soma turns that expression into a governed execution contract, work happens visibly, and durable outputs/proof remain easy to revisit.

The researched pattern is consistent across human-AI UX guidance:

- explain the benefit before the technology
- preserve user control where automation may fail
- reveal AI presence and proof without flooding the default path
- show the status, consequence, and recovery path of work
- keep feedback, correction, and steering close to the output
- make advanced/runtime detail inspectable, not primary

For Mycelis, this means the primary UI object is not "chat". It is an expression-to-output workspace.

## Research Inputs

| Source | Relevant Lesson For Mycelis |
| --- | --- |
| [Microsoft HAX Toolkit](https://www.microsoft.com/en-us/haxtoolkit/) and [Guidelines for Human-AI Interaction](https://www.microsoft.com/en-us/haxtoolkit/ai-guidelines/) | Plan AI behavior across initial interaction, normal interaction, wrong-system behavior, and use over time. Use failure planning as part of design, not support cleanup. |
| [Google PAIR Mental Models](https://pair.withgoogle.com/guidebook-v2/chapter/mental-models/) | Introduce AI in stages, explain user benefit over technology, and give non-dead-end fallback when AI cannot complete work. |
| [Google PAIR Feedback + Control](https://pair.withgoogle.com/guidebook-v2/chapters/feedback-controls/) | Balance automation with user control; feedback must make sense to users and connect to visible future value. |
| [Google PAIR Explainability + Trust](https://pair.withgoogle.com/chapter/explainability-trust) | Trust depends on the user understanding capability, limits, data use, and confidence at the right level of detail. |
| [Google PAIR Errors + Graceful Failure](https://pair.withgoogle.com/chapter/errors-failing/) | Failure is normal in AI systems; error states need diagnosis, recovery, and an alternative path. |
| [IBM Carbon for AI](https://carbondesignsystem.com/guidelines/carbon-for-ai/) | AI presence and generated content should be identifiable; explanations should be integrated into workflow and shown when needed or requested. |
| [IBM Explainability](https://www.ibm.com/design/ai/ethics/explainability) | Users need ongoing ways to ask why an AI is doing what it is doing; decision processes need review records. |
| [VIKTOR App Builder](https://docs.viktor.ai/docs/app-builder) and [VIKTOR app structure](https://docs.viktor.ai/docs/getting-started/fundamentals/basic-app-structure/) | Complex expert workflows become approachable when input, preview/results, save/revisit, and deployment/share actions are held in one editor/workspace mental model. |
| [VIKTOR Home update](https://docs.viktor.ai/docs/whats-new/) | Enterprise platforms reduce navigation complexity by adapting home surfaces around role and activity, such as building, using, and personalized work. |

## Mycelis Interpretation

### The User Is Not Querying A Database

The user expression is an act of work definition:

```text
I want an outcome.
I need it shaped for a purpose.
I want Soma to decide the safe execution posture.
I want to see output, proof, and recovery.
I want to steer without learning internal topology first.
```

That expression must be captured as:

- outcome
- output shape
- audience/use
- constraints
- agentry posture
- capability and governance boundary
- acceptance proof
- continuation mode

The UI should help the user refine that expression only when it improves execution. It should not ask for execution topology before value is visible.

### Soma Is The Operating Surface

Soma is the persistent interface to work. Teams, Council, capabilities, runs, memory, outputs, schedules, and logs are not peer identities in the default UI. They are scoped operational contexts.

Default phrasing should be:

- "Soma will create..."
- "Soma needs approval because..."
- "Soma is waiting on..."
- "The team is producing..."
- "The output is ready..."
- "Proof is incomplete because..."

Default phrasing should avoid:

- "NATS subject..."
- "agent roster..."
- "payload..."
- "tool_call..."
- "runtime topology..."
- "server log..."

Those details stay available through Inspect/Advanced when they explain authority, proof, failure, or deployment posture.

## Target Information Architecture

### 1. Soma Workbench

The main workspace should hold four regions in one browser window:

| Region | Purpose | Default Content |
| --- | --- | --- |
| Expression | Capture the desired outcome and constraints. | Natural-language ask, structured expression frame, starter outcomes. |
| Active Work | Show what is running, blocked, degraded, awaiting input, or ready. | `TeamWorkItem`, schedule work, proposal state, compact progress. |
| Output Workbench | Review durable results. | Output preview, files/packages/media/reports, validation, open/download/reuse. |
| Trust Package | Explain why the user can trust or recover. | proof link, audit refs, validation source, degradation, next action. |

The workbench should avoid page-level scroll as the primary behavior. Each region can have internal scrolling only after it is capped, labeled, and useful.

### 2. Soma Response State Cards

Every Soma response should render one primary state:

- direct answer
- proposal
- awaiting approval
- running
- output ready
- partial completion
- degraded execution
- blocker
- recovery required

The state card must answer:

- What is happening?
- Why this posture?
- What changed or will change?
- What output/proof exists?
- What can I do next?

### 3. Output-Led Agentry

Agentry should be conscripted by the output:

| User Wants | Default Agentry | UI Control |
| --- | --- | --- |
| quick answer | Soma direct | answer with source/proof inspect |
| project package | compact execution lane | files, entrypoint, README, validation, proof |
| review/report | evidence lane | findings, sources, criteria, downloadable report |
| complex work | active team | team work item, ask/respond/recover controls |
| recurring work | governed schedule | cadence, next run, cooldown, approval posture |
| deployment proof | system trust lane | roots, endpoint posture, commit/image/chart, recovery |

The operator does not need to choose "team" first. The requested output should make the needed execution posture obvious.

### 4. Teams As Active Work, Not A Roster Center

Team UI should lead with:

- active work item
- accountable lead
- current state
- expected output
- last readable event
- ask/respond/recover controls
- proof/output refs

The roster, model, prompts, tools, topics, and bus details belong behind Inspect. This keeps growing agentry from becoming a growing UI problem.

### 5. Runs As Proof Timelines, Not Log Streams

Run pages should show:

- intent
- approval/governance posture
- major state transitions
- output refs
- proof/audit refs
- degradation/recovery

Raw event payloads, internal bus subjects, and provider traces stay behind explicit Inspect controls. Long run conversation turns should be clamped with "Show full turn" so the timeline remains navigable.

### 6. Activity As Operations Summary

Activity should be a concise operational pulse:

- recent outputs
- errors/degradation
- approvals/governance
- capability use
- active team status
- deployment/system posture

It should not default to a bus terminal. Advanced activity can expose raw signal detail when the operator is diagnosing system behavior.

### 7. Resources As Capability Catalog

Resources should answer:

```text
What can Soma use?
Is it healthy?
What risk class is it?
What approval does it need?
What outputs does it produce?
What happens if it fails?
```

MCP/server/tool installation details are secondary to capability meaning. The target pattern is menu/detail panes by resource type, with local workspace, connected tools, deployment context, and library entries separated.

### 8. Schedules As Future Work

Scheduled events should not be hidden automation config. They are future work commitments:

- what Soma/team will do
- when it will run
- what it can change
- approval mode
- expected output/proof
- last run
- next run
- failure/retry posture

This belongs near Active Work and Automations, not buried as scheduler internals.

## Target Interaction Flow

```text
1. User expresses desired output.
2. Soma frames the expression: outcome, output, use, constraints, proof.
3. Soma picks response state: answer, proposal, active work, blocker, recovery.
4. If governed, proposal summarizes consequence and approval posture.
5. Execution creates or updates active work.
6. Output appears in the Output Workbench.
7. Trust Package links proof, audit, validation, and recovery.
8. User can steer, ask the team, schedule continuation, or inspect detail.
```

This is the target shape for every primary workflow.

## Design Rules

1. Default screens summarize; Inspect screens explain.
2. Keep one primary work window for expression, active work, output, and trust.
3. Do not show raw JSON, bus topics, model prompts, server logs, or rosters as default content.
4. Show AI presence and generated/proposed content consistently, but do not decorate with AI styling where no AI action occurred.
5. Make every wait state say what is happening and what the operator can do.
6. Make every degraded state say what failed, what remains trusted, and what can be retried.
7. Treat feedback and steering as work controls, not afterthought reactions.
8. Let users complete or recover work without understanding internal topology.
9. Make evidence durable enough for reload, audit, and handoff.
10. Keep the advanced surface powerful but clearly secondary.

## Delivery Plan

### P0 - Main Workbench

- Combine Soma expression, Active Work, Output Workbench, and Trust Package into one bounded workbench layout.
- Keep chat transcript secondary to current state and durable outputs.
- Add structured expression frame preview for outcome, output shape, constraints, proof, and continuation.

### P0 - State Card Standard

- Build a shared `SomaStateCard` contract for direct answer, proposal, awaiting approval, running, output ready, degraded, blocker, and recovery.
- Use consistent chips, actions, and proof rows across dashboard, organization workspace, teams, and runs.

### P0 - Output Workbench

- Promote outputs above logs.
- Support project package, report, file, media, MCP result, deployment proof, and schedule output shapes.
- Cap visible files/proof refs and provide Inspect for full evidence.

### P1 - Team Work Surface

- Make `TeamWorkItem` the primary team UI unit.
- Keep team member roster summarized by role/status.
- Add active-team conversation/steering in the work item, not in a separate agentry console.

### P1 - Runs And Activity Compression

- Make run timelines proof-first.
- Clamp long turns.
- Keep raw payloads behind Inspect.
- Keep Activity as an operational pulse, not default diagnostics.

### P1 - Scheduled Work Surface

- Add schedule cards with next run, output contract, approval posture, proof expectation, and recovery state.
- Let Soma propose, edit, pause, resume, and recover scheduled work through the same governance pattern.

### P2 - Enterprise Control Plane

- Make identity, roles, SSO, provider posture, capabilities, deployment trust, and audit navigable through focused menu/detail panes.
- Keep admin pages dense but bounded; every long list needs filtering, grouping, and internal scrolling.

## Acceptance Criteria

For each UI delivery slice:

- A new user can state what Soma is doing without reading logs.
- The screen has one obvious primary action.
- The output or next proof object is visible above raw runtime detail.
- The user can steer, retry, recover, or inspect from the current state.
- Advanced/runtime detail is still available without being default.
- Desktop and mobile have no horizontal overflow.
- Browser proof verifies the actual interaction, not only route rendering.
- Docs and state name the default/Inspect boundary changed.

## Open Questions

- Should the expression frame be explicitly editable before approval, or inferred and corrected through proposal revision?
- Should the main workbench default to "current work" or "latest output" after a successful run?
- How should confidence provenance be shown without introducing misleading numeric scores?
- Which team status events are operator-facing, and which remain audit-only?
- What minimum schedule authoring UI is enough for first enterprise review?
