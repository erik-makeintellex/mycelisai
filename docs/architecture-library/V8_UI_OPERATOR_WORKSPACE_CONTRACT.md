# V8 UI Operator Workspace Contract
> Navigation: [V8 UI/API Contract](V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md) | [V8.2 Soma UI Architecture Expression](V8_2_SOMA_UI_ARCHITECTURE_EXPRESSION.md) | [V8.2 Current State And Finalization PRD](V8_2_CURRENT_STATE_AND_FINALIZATION_PRD.md)

Status: canonical workspace detail.

## Target State

The target operator workspace is a single Soma-centered work window, not a page stack of architecture consoles.

The default window keeps these areas visible or one click away:

- expression input and outcome-specific starter prompts
- active work lane for queued, running, blocked, degraded, and output-ready work
- current output preview or workbench
- compact trust package for proof, recovery, and next action
- scoped context indicator for organization, team, deployment, or retained output when relevant

The operator should be able to say what they want, see how Soma framed the work, watch active execution, open the output, and inspect proof without needing to visit Teams, Resources, Runs, System, or logs first.

## Expression Contract

User expression is not handled as a raw query. Soma should frame it as governed work:

- outcome: what should exist when work is done
- output shape: answer, plan, file, project package, review, media, team deliverable, schedule, or deployment proof
- audience/use: who will use the output and why
- constraints: scope, exclusions, risk, permissions, style, data, and timing
- agentry posture: direct Soma, capability-backed execution, compact team, Council review, scheduler, or blocker
- acceptance proof: validation, retained output, run/proof link, review lineage, or recovery state
- continuation: one-time, active work, retry/review, or recurring

The UI should show that frame before raw runtime topology. Teams, MCP, bus subjects, provider details, and deployment roots appear only when they affect trust, authority, recovery, or operator control.

## First-Run Flow

The first-run path should teach that Mycelis gives the operator one governed Soma work surface, not disposable chat threads or an agent console.

Required visible concepts:
- Soma
- concrete things Soma can do now
- how work becomes an output
- where proof/recovery will appear
- starter template or empty-start choice when an AI Organization is useful
- AI Organization name/purpose as scoped context, not a separate assistant identity
- AI Engine Settings, Memory & Continuity, and Connected Tools as guided supporting concepts

Cold start must not imply prior work. Do not render completed-work or trust-package language until real work exists in the selected context.

## Organization Re-Entry

When organizations exist, default to the last active or selected organization. The primary action should open the Soma workspace in that context, not start an anonymous chat.

If Soma is universal in the current product lane, organization re-entry is a scoped context transition. The UI must not imply that a new Soma identity is created per organization.

## Soma Workspace

Soma workspace must support:
- expression framing
- direct answers
- proposal state for protected actions
- awaiting approval and cancellation clarity
- active execution state
- execution results
- blockers, degraded execution, and retry guidance
- retained output references
- activity/run visibility
- proof, audit, and recovery summary
- follow-up steering while work is active

The main workspace should not become a tall transcript with buried proof. The chat thread is the operating conversation; the active work lane and output workbench are where the user understands whether anything real is happening.

## Single-Window Layout

The target browser expression uses fixed work areas rather than whole-page sprawl:

- dashboard/home: Soma input, starter outcomes, active work, current output, compact trust summary
- Teams: active work list plus selected team/work-item detail; roster and raw bus details behind inspect
- Resources: type menu plus focused resource detail/workbench; server inventories and raw manifests behind inspect
- Runs: run list plus selected run proof/output/recovery detail
- System: cockpit summary plus Deployments detail for roots, endpoint posture, proof lane, and recovery
- Settings: menu/detail panels for auth, providers, access posture, and secret-reference configuration

Long route-level scroll is a design failure for primary work. Inner scrolling is allowed only inside bounded lists, detail panes, code/log viewers, and output previews.

## Browser And Mobile Expression

Desktop should keep expression, active work, output, and trust context in one coherent viewport.

Mobile should collapse navigation chrome and present:

- input and current response state first
- active work as a compact lane or sheet
- output/proof as a bottom sheet or detail route
- advanced details behind explicit inspect actions

Horizontal overflow is not enough to fail a mobile review; the release gate should also fail when the rail, dense cards, or oversized trust packages leave no practical work surface.

## Teams And Groups

Team/group surfaces should be compact and reviewable:
- broad asks become smaller lanes
- Team Lead focus stays visible
- retained outputs remain accessible after closure/archive
- communication/status/result history remains inspectable
- team creation is not treated as completed work when the user asked for a deliverable
- active teams can be inspected, steered, paused, resumed, retried, stopped, or archived without reading raw NATS traffic

## Advanced Boundaries

Advanced mode may show:
- provider policy
- response style
- inheritance/source details
- capability posture
- templates/config origin

It must not be required for a beginner to create and use an AI Organization.

Advanced surfaces should answer "why can I trust or recover this?" rather than "how much topology can we display?"
