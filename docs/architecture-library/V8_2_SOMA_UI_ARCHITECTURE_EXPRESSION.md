# V8.2 Soma UI Architecture Expression
> Navigation: [Project README](../../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)

Status: canonical target-expression document.

Purpose: define the ideal operator experience and UI architecture posture for Soma as Mycelis moves from architecture description into operational embodiment.

This document owns the expression model. Research-backed target UX, external calibration, information architecture, and delivery order should be promoted here or into the V8.3 embodiment docs instead of being retained as separate research notes.

## Core Idea

The UI is not a dashboard around an AI system. The UI is the operator's working relationship with Soma:

```text
Ask for meaningful work
-> understand what Soma plans to do
-> approve meaningful change when needed
-> watch execution progress
-> review durable outputs
-> inspect proof and recovery
-> steer the next step
```

Teams, tools, runs, capabilities, memory, schedules, deployment roots, logs, and audits are operational contexts behind that loop. They explain what Soma can use, what happened, what remains trusted, and what can happen next.

## Enterprise Interaction Posture

Preserve these enterprise-grade patterns:

- one clear assistant/work surface instead of a maze of admin consoles
- connected tools described by what Soma can do, not installed-server trivia
- role and use-case workflows that make expert systems approachable
- integrations and secrets handled as governed configuration, never pasted into chat
- scheduled or recurring work presented as accountable future actions with proof
- action history, provenance, and recovery visible without forcing logs into the default path

## Product Promise

The default UI must make this sentence true for a new user:

```text
I ask Soma for work, I see whether it is answering, proposing, running, blocked, or done, and I can open the output and proof.
```

If a screen cannot support that promise, it belongs in a secondary or advanced surface.

## Operator Expression Model

An operator expression is not a database query, search string, or chat prompt to be answered and forgotten. It is a request to evoke governed work: a desired future output, its purpose, the constraints around producing it, the agentry needed to make it real, and the proof required to trust it later.

The UI should help Soma convert expression into a compact expression frame before topology appears:

| Field | User-facing meaning |
| --- | --- |
| Outcome | What should exist when work is done. |
| Output shape | File, project package, review, plan, media, report, scheduled rule, deployment proof, or team deliverable. |
| Audience and use | Who the output is for and how it will be used. |
| Constraints | Must-have scope, exclusions, risk boundaries, deadline, style, data, and permissions. |
| Agentry posture | Direct Soma, capability-backed work, compact team, Council review, scheduler, or blocker. |
| Acceptance proof | How the user will know the result is complete, valid, and recoverable. |
| Continuation | Whether this is one-time work, active team work, review/retry, or scheduled recurrence. |

Soma should reflect the frame in operator language:

```text
Output: playable browser game package
Use: reviewable project starter
Agentry: compact game team plus file-writing capability
Proof: README, playable entrypoint, validation status, retained run link
Next: approve, revise scope, or start work
```

This is the interaction contract that prevents Mycelis from feeling like a prompt box over hidden machinery. The operator should see the work being shaped before they are asked to understand the machinery.

## Output Conscription

Outputs are not passive attachments. The requested output conscripts the right execution posture:

- A direct answer conscripts Soma reasoning and optional source/proof disclosure.
- A project package conscripts file/artifact capabilities, validation, preview/open controls, and retained proof.
- A review conscripts evidence, criteria, findings, and change recommendations.
- A team deliverable conscripts a `TeamWorkItem`, active work state, team status events, outputs, and recovery.
- A scheduled result conscripts cadence, approval posture, next-run state, cooldown, proof, and failure recovery.
- A deployment proof conscripts roots, endpoint posture, current commit/image/chart, health checks, and recovery actions.

The output type should determine the visible UI state, available controls, and required proof. Users should not have to say "create a run, call this capability, make a proof artifact"; they should say the work they want, and Soma should present the governed output contract.

## Research Calibration

Microsoft HAX, Google PAIR, IBM Carbon for AI, IBM explainability guidance, and VIKTOR-style app building all reinforce the same UI lesson: start with concrete user value, show capability and limits early, keep correction/control easy, and explain consequences through proof and recovery instead of topology. Research calibration informs the product expression; it does not justify more default surfaces.

## Operator Journey

### 1. Start

The first screen answers:

- What can I ask Soma to do?
- What can Soma currently use?
- What will I see when work happens?

A cold-start surface must not imply hidden prior work. Do not show "Soma just did this" until Soma actually did something in the current visible thread or selected context.

### 2. Ask

The input is the dominant control. Starter prompts should be concrete outcomes:

- build a playable prototype
- review a document or plan
- create a reusable report
- connect or check a tool
- schedule a recurring review

The default path should not require users to understand teams, NATS, MCP, or proof classes before they receive value.

### 3. Understand

Every Soma response maps to one UI response state:

- direct answer
- proposal
- awaiting approval
- running
- execution result
- partial completion
- degraded execution
- blocker
- retry required
- recovery state

These states are UI contracts, not copy variants.

### 4. Execute

Active execution needs a visible work lane:

```text
Queued -> Running -> Output ready -> Proof retained
```

For team-backed work, distinguish team created, first work item queued, team active, output ready, operator input needed, and degraded or timed-out work. A team existing is not the same thing as a team working.

### 5. Review

Outputs are product objects. The UI privileges title, type, preview/open action, storage or reconstruction details, owning run, proof quality, and next action.

The output view should be stronger than the chat transcript. Chat is the operating conversation; outputs are the durable deliverables.

### 6. Trust

Trust surfaces answer:

- What did Soma understand?
- What changed?
- What capability, team, source, model, schedule, or deployment context mattered?
- What output exists?
- What proof exists?
- What failed, if anything?
- What can safely continue?
- What does the operator do next?

The default card shows outcome, outputs, proof, and next step. Full lineage belongs behind inspectable disclosure.

## UI Architecture Domains

### Soma Operating Surface

Owns input, starter prompts, response-state rendering, current workflow summary, active execution, current output summary, trust summary, and follow-up steering.

Target expression: one reusable Soma operating model across dashboard and organization workspaces.

### Output Workbench

Owns retained file, media, plan, review, and project-package views; open/download/reveal controls; preview and reconstruction metadata; output-to-run linkage; and output-to-team linkage.

Target expression: outputs are first-class review surfaces, not attachments buried in chat.

### Active Work Lane

Owns queued/running/blocked/done state, team or capability work-item identity, progress events, timeout/degradation state, and operator-needed prompts. Soma home shows an attention-first recent slice; Teams owns full durable backlog management and inspection.

### Current Work Lane

Owns the compact Dashboard summary of selected workflow, active task posture, latest retained output access, and next review action. It is a summary of existing Active Work and Output Workbench state, not a separate runtime object or revived pre-chat output dock.

Target expression: users can work with teams while they are still active.

### Trust Package

Owns intent, understanding, governance posture, capability/team/source use, proof, audit/recovery, confidence provenance fields, and next step.

Target expression: compact by default, complete on inspection.

Trust package detail should open into a reusable trust drawer that can explain a proposal, execution result, failed run, automation cycle, tool use, deployment check, or governed context intake with the same chain:

```text
Intent -> Decision -> Authority -> Capability -> Execution -> Output -> Proof -> Audit -> Recovery
```

### Capabilities And Tools

Owns what Soma can use, health, risk, permission posture, schemas, output shapes, recovery behavior, and recent use.

Target expression: connected tools answer "what Soma can do safely" before "which servers exist." Capability cards should lead with purpose, source, risk, approval, allowed roles, secret references, health, outputs, audit, and fallback. Server details are drill-down.

### System And Deployment Trust

Owns runtime health, deployment root, execution root, workspace/artifact root, current commit, endpoint posture, proof lane, and recovery action.

Target expression: self-hosted operators can trust the installation without reading container logs first.

`System -> Deployments` is the enterprise trust home for Compose/Kubernetes lane, current commit/image/chart, endpoint posture, AI/search/filesystem readiness, workspace/artifact roots, proof lane, and recovery commands.

## State Architecture

Default state views show the operator ask, Soma summary, decision, risk, current-work summary, active work, output card, proof link, degradation next step, and schedule next run. Default activity and team surfaces summarize recent evidence instead of streaming raw logs or rosters. Inspect views reveal original request, assumptions, contracts, resources, policies, events, team/capability details, storage paths, schemas, evidence, audit refs, validation source, retryability, uncertainty, cooldowns, and history.

Security rule: UI may show secret reference names, redacted readiness, and rotation guidance, never raw secret values.

## Team Execution Architecture

Teams should be visible only when they help the operator understand or steer work.

Default team behavior:

- Soma creates compact teams only when the ask benefits from persistent ownership.
- New teams start with one accountable lead unless the work proves a missing capability.
- Team creation should usually pair with the first concrete deliverable or a clear "start work" next step.
- Team communication appears as operator-readable events, not raw bus traffic.
- Temporary teams need stop/archive/retry behavior and proof.

Target active-team view:

```text
Team
Current objective
Active work item
Recent events
Outputs
Needs operator
Stop/archive
```

## Scheduled Work Architecture

Scheduled work must be governed execution, not a calendar label.

Minimum production shape: rule name, owner, cadence or trigger, next run, cooldown, approval posture, capability/team scope, proof expectations, last result, and recovery behavior.

Schedule language stays out of default UX until these fields exist.

## Execution Teams

Soma Experience, UI Architecture, Governance And Trust, Runtime And Capability, and QA And Embodiment own the first production path together: compress default UX, isolate response-state rendering, make proof understandable, persist runtime state, and prove cold-start, active-team, output-proof, recovery, mobile, and scheduled-work behavior. Enterprise role paths should become explicit over time: owner/admin configuration, operator execution, reviewer approval/audit, platform deployment proof, and security capability/secret posture.

## Acceptance Gates

A UI slice is acceptable only if it improves at least one of:

- visible execution
- operator trust
- durable output review
- recovery clarity
- deployment reality
- Soma as singular operating surface
- conceptual compression

Required evidence for behavior changes:

- focused component tests for state rendering
- Playwright proof for the touched operator workflow
- live backend proof when runtime state changes
- docs/state update when terminology or product behavior changes
- no new placeholder or test-only product surfaces

## Immediate Finalization Decisions

1. The default dashboard must not show a post-action trust package before a real action exists.
2. Active team work needs an operator-readable lifecycle surface.
3. Output review must become stronger than chat transcript review.
4. Connected Tools should read as "what Soma can use" before "installed MCP servers."
5. Scheduling must return as governed cadence only when next-run, proof, cooldown, and recovery are real.
6. Advanced topology stays inspectable but secondary.
7. Deployment trust needs a real `System -> Deployments` surface before enterprise delivery claims are complete.

## Final Principle

The ideal UI does not make Mycelis look simpler by hiding truth. It makes Mycelis feel simpler by showing the right truth at the right time:

```text
What Soma will do.
What Soma is doing.
What Soma produced.
Why the result can be trusted.
What the operator can do next.
```
