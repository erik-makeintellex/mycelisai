# Teams
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Soma-first team workflow for active team work: start lead-owned by default, and create explicit specialist rosters only when the requested output already names distinct roles.

---

## TOC

- [Overview](#overview)
- [Default Launch Shape](#default-launch-shape)
- [Choosing The Right Workflow](#choosing-the-right-workflow)
- [What A Good Team Looks Like](#what-a-good-team-looks-like)
- [When Soma Should Split Work](#when-soma-should-split-work)
- [Working With A Team Lead](#working-with-a-team-lead)
- [Team Creation](#team-creation)
- [Useful Expectations For Testing](#useful-expectations-for-testing)

## Overview

The Teams area is where you review active teams, open a team lead workspace, and create new teams through Soma.

Use the adjacent workstream surfaces deliberately:
- `/teams` is for active team work, team lead workspaces, team shape, and current execution posture.
- `/groups` is for focused collaboration lanes, workflow-log review, retained outputs, archived collaboration records, and temporary-workflow history.
- `/automations` is for trigger rules that react to events, plus approvals around automated actuation.

Default rule:
- teams should be small, explicit, and focused on the output
- the ideal launch team is one accountable lead
- a single team should stay lead-only at creation unless the operator explicitly asks for named specialist roles tied to a retained deliverable
- broad requests should become a few lead-only teams or lane bundles, not one large roster
- add temporary specialists only after the operator or team lead names the missing capability, owned task, expected proof, and removal point
- the root Soma workspace remains the main place to ask for team creation and orchestration

## Default Launch Shape

For most work, Soma should launch one accountable lead first.

The lead is the user-facing counterpart who keeps the mission, status, and outputs clear. If the work truly needs more coverage, the operator can add a member deliberately or the temporary team lead can request one temporary specialist with:

- the missing capability
- the owned task
- the proof expected
- the removal point

Soma should not launch a broad standing pool just because the request sounds ambitious. If multiple capabilities seem useful, Soma should split the work into smaller lead-owned lanes and explain the lane outputs.

Explicit specialist-output requests are different. If the operator asks for a concrete team like "an artist, a character designer, and someone who writes the lines" and also asks for a retained output such as a comic page or image, Soma may create a bounded specialist delivery roster in the same governed proposal. That roster must still have one accountable lead, named specialist roles, expected output/proof, local/private media posture where relevant, and retained artifacts.

## Choosing The Right Workflow

Use the root Soma chat when you want the simplest path:

- "Soma, create a compact team to produce a customer-facing launch brief and keep the team to the smallest useful shape."
- "Soma, this is a broad product-review ask; split it into small lanes and tell me what each lane will output."
- "Soma, summarize what teams exist, what they are doing, and where their outputs are."
- "Soma, review this request against recent commands, infer the team/action you think I want, include target MCP tools, and ask me to confirm before you launch it."

Use `Teams` when you want to inspect or manage existing teams:

- read the compact Outcome Health badge first: Healthy, Waiting, Running, Degraded, Blocked, Completed, or Archived; open details only when you need the underlying team lifecycle state

- review the Active Work Lane to see whether a team is new, queued, running, output-ready, degraded, paused, or waiting on the operator
- use `/teams?view=work` when arriving from the Dashboard review panel; this focused Review Queue starts with counts for work needing a decision, ready output, work still running, and items that can be cleared
- in Review Work, each row should answer `Reason`, `Trust`, and `Move` before the decision actions so the operator can decide whether to inspect, respond, recover, or clear it without reading the whole team setup page first
- treat the Dashboard Active Work lane as an attention-first slice; its API projection includes operator decisions, recovery, and ready deliverables without counting ordinary queued/running progress as review; use `Teams` for the full durable backlog
- use the Dashboard current-work lane for the quickest read of focused workflow, active task posture, latest output, and next review action
- use the Dashboard `Working in` picker when you want to switch Soma between `Soma root` and a specific team's focused chat/output/proof lane without leaving the main workbench
- when a focused team has retained outputs and no active work needs attention, use the current-work lane or Work panel for immediate open/open-folder access; when work is queued, running, degraded, or waiting on the operator, the lane keeps Work as the primary next action while preserving latest-output access
- focused team chat stays scoped to that team for conversation continuity, proposals, and team bus wiring while root Soma remains the cross-team reviewer when no team is selected
- open the team lead workspace
- review current outputs while the team is active
- use Ask Team or Respond on a durable active-work row to queue a bounded follow-on output or supply missing input without opening raw bus details
- use `Open details`, `Reply to team`, `Ask for changes`, `Start task`, `Pause`, `Resume`, `Retry recovery`, or `Clear from review` when those controls are enabled for the current team state
- inspect member templates
- review or edit template role, model, and MCP/internal tool references
- check whether a team should be archived or kept

When a team has just been created and no delegated work item exists yet, the Dashboard shows a first-deliverable launcher instead of treating the team shell as active work. Choose a starter such as `Build playable prototype`, `Write design brief`, or `Draft delivery plan`; Soma places the bounded ask in the chat input for review, then your send creates the governed work item that can run, produce output, and attach proof.

Ask Team is non-blocking. When you queue a follow-on ask, the row should close the form, show a queued work item immediately, keep the workspace usable, and refresh Active Work while the team moves through `running`, optional `reviewing`, `output_ready`, or `degraded`. Async asks record durable dispatch state so a published command is not invisible. Correlated status/result signals return to the original row. Retained output refs stay attached to that team context, but an interactive package in `reviewing` is only a candidate: Core checks the exact retained digest in a browser before attaching runtime proof or completing its run. A pass advances to `output_ready`; a failed or unavailable check advances to `degraded` with a concise repair request for Soma and the same team. Missing retained outputs also degrade rather than claim completion. Durable output refs store workspace-confined paths, not viewer URLs, so the UI can derive **Open**, **Open folder**, and Resources actions consistently. If the bus or worker lane is unavailable, the ask remains durable and explains recovery instead of leaving the operator waiting on a browser request.

`Clear from review` archives a durable work item so it leaves active review queues while retained outputs, proof refs, audit refs, and history remain inspectable. Use it for stale failed proposals or old test data after confirming nothing useful is waiting to be recovered.

Use `Groups` when you want to review a collaboration lane without opening every team surface. The selected group includes a **Workflow Log** tab that combines the group brief, lifecycle recommendation, attached team-work rows, retained output cues, latest broadcast result, and bus/recovery signal into one operator-readable stream. It is workflow context, not a final deliverable folder and not raw NATS/bus logs. Group workspace tabs keep the selected group and panel in the URL, so an operator can return directly to `overview`, `workflow`, `outputs`, `message`, `settings`, or `create` during review handoff.

Groups opens on **Current** so completed and archived history does not crowd active work. Open **Filters** to switch between **Current**, **Completed**, and **Archived**. Completed means an expired temporary collaboration that has not yet been archived; Archived means a cleared retained record. Completed history can be bounded by age. Route-selected records remain inspectable even when they sit outside the current filter.

On phones and tablets, Groups shows one job at a time. Start in the group-record list, choose a group to open its workspace, and use **All groups** or browser Back to return to the list. The workspace sections remain a horizontal tab strip instead of compressing the record list and detail pane beside each other. Desktop keeps the list and selected workspace visible together.

The **Create** tab is sectioned as **Basics**, **Policy**, **People**, and **Advanced** so the operator can define one part of the collaboration lane at a time instead of reading a compressed multi-column form. Start with the name and goal, then add work mode/approval posture, team or member ids, and only then any workspace/coordinator detail that matters.

The **Outputs** tab is curated for user-facing deliverables. It hides planning, proof, source/support files, and team handoff records by default so a planned team does not look like it delivered real work. Use the include-internal checkbox when you intentionally need to inspect planning records such as `TEAM_EVOCATION.md`, proof files, research handoffs, or source material. A group labeled **Planned only** has retained working material but still needs a delivered output before it should be treated as complete. When you hand a file to a team, say whether it is a one-run draft/input, standing context for that team, or a final output target; Soma should keep those roles separate in the work item.

Group Outputs use the same Outcome Health vocabulary as Soma and Runs: a group with retained user deliverables is Completed, an empty active output lane is Waiting, and a cleared group is Archived. Proof remains visible separately and does not rename a completed deliverable as Healthy.

Standing groups and Soma-created runtime-team groups also have a dedicated workspace folder under `MYCELIS_WORKSPACE/groups/...`, visible from the group detail pane with an `Open folder` action.

Use **Clear group** when a group is done, stale, or test-only. Clearing archives the group so it leaves active/review lanes. Message-bus handoff data is transient and does not need separate cleanup. Retained output files stay available by default; choose **Also remove retained output files** only when the durable files in that group workspace should be removed too. When retained files are included, Mycelis removes the group workspace folder under `MYCELIS_WORKSPACE/groups/...` and archives the group's output artifact rows so the cleared deliverables no longer appear in Resources group outputs.

For repeated cleanup, use **Select** in the Groups record rail. Select mode lets an operator choose multiple active groups from the currently filtered list and clear them together. Bulk clear keeps retained output files by default and is meant for old test lanes, completed temporary lanes, or stale active records that no longer need attention. Turn on **Also delete retained output files** only when the selected group folders and curated output artifacts should be removed too. Archived groups are shown for review but are not selectable for bulk active-lane actions.

When a group is meant to react to service, device, API, or sensor traffic, register that traffic as an input source first. Soma or the group should describe whether it needs an append log, latest-state view, or windowed summary. Teams should work from those buffered source references instead of raw real-time traffic, so fast streams remain useful without overwhelming agentry or hiding proof.

Use `Automations` when you want event rules to actuate work, route proposals, require approval before execution, or author propose-only Schedule Rules for reviewable cadence.

Use `Resources -> Worker Profiles` to inspect ready-made teammates or create scoped profiles for recurring work. Name a profile in a Soma request when you want that teammate specifically, or let Soma choose the smallest useful set. Use `Resources -> Capabilities` to confirm the access those profiles may request; a profile never bypasses capability health, source scope, approval, or secret policy.

If Soma recommends tools that are not installed yet, it should walk you through the enablement path before launch: name the missing MCP server, name required `.env` variables without exposing secret values, point to Capabilities and the MCP library, and then bind the resulting tool refs to the team or reusable member template after you confirm.

Use `Create Team` when you want a guided setup instead of filling raw fields:

- describe the outcome
- review Soma's recommended lead-only start
- confirm whether any temporary specialist is justified yet
- decide whether the work should stay as one team or split into lanes
- confirm the visible output contract before launch

If you need the higher-level boundary between direct Soma, one context-rich agent, compact teams, and multi-lane orchestration, read [Workflow Variants And Plan Memory](workflow-variants-and-plan-memory.md).

## What A Good Team Looks Like

Most teams should have:
- one clear lead
- a small specialist set
- a narrow mission
- readable outputs
- a named output contract such as "brief", "test plan", "image prompt pack", "website draft", "data review", "implementation patch", or "release checklist"

That keeps the team easy to inspect and easy to test.

Avoid teams that grow into a giant roster unless the work is so broad that splitting it would make the workflow less clear.

## When Soma Should Split Work

Soma should split a request into several smaller teams when the request spans:
- planning and implementation
- research and delivery
- media generation and review
- multiple departments or product areas
- multiple outputs that would be clearer as separate lanes

In those cases, each lane should keep its own lead and output contract while Soma coordinates the whole set.

Example broad split:
- Planning lane: Team Lead, Architect Prime, and focused researcher.
- Build lane: Team Lead, Architect Prime, and focused builder.
- Review lane: Team Lead, focused reviewer, and domain specialist.

Each lane stays inspectable, and Soma coordinates handoffs over managed exchange and NATS rather than hiding a large internal pool.

## Working With A Team Lead

When you open a team:
- start with the team lead
- inspect the current outputs and active artifacts
- ask the lead to summarize what the team is doing
- use the lead to reach specialists only when needed
- ask for "what changed, what was produced, and what remains" when you need a fast state read

The team lead is the user-facing counterpart for that team, not a hidden extra member list.

After a temporary collaboration is archived, use `/groups` to review the workflow log, retained output package, and collaboration record.

For every team, keep deliverables inside the team group folder instead of the general output folders. Team creation prepares `groups/<team-id>/planning`, `groups/<team-id>/source`, and `groups/<team-id>/generated`; restoring an existing runtime team repairs those folders when they are missing. Soma-owned team media defaults to `groups/<team-id>/media`, and Soma-owned team project packages default to `groups/<team-id>/generated/...`. The general generated folder is appropriate only for explicit unscoped one-file output or short-lived handoff material. Explicit operator paths are still respected when you intentionally name a different workspace-confined target. Long-running teams should also name approved input mounts or Deployment Context sources they may reuse so they do not treat every old working file as current truth.

A package team is complete only when it returns a retained project package under `groups/<team-id>/generated/...` with an entrypoint and local support files that physically exist. If the request requires a playable or interactive workflow, the package must also expose and pass the requested primary interaction; opening static HTML is not enough. A loose `.py`/`.js` file, copied request text, planning brief, missing parent/dependency, or non-responsive control is not a delivered application. Valid completion returns to the Soma conversation as `Work complete`, names the retained deliverable, and provides a direct open action; run and transport evidence remain available through Details/Inspect.

Active Work retains the approved execution mode and its lifecycle guidance. Open `Advanced inspect` when you need to verify whether the work is one-shot, scheduled, a continuing service, a project, or a Soma extension and which stop, retry, and recovery operations apply. Those labels describe the governed control path; they do not let a team bypass approval or policy.

## Team Creation

Use the guided team-creation workflow when you want Soma to shape the team for you.

Tell Soma:
- what outcome you need
- how broad the request is
- whether the work should stay as one compact team or split into multiple lanes
- what outputs you want visible at the end
- what source files, mounted folders, or saved context the team may use
- whether those sources are one-time handoff material or long-lived context
- whether this is a temporary group whose logs can be reviewed and whose outputs should remain after closure
- whether the group workflow log should show only final user-facing outputs or include deeper team-work/source context during review

If the request is broad, expect Soma to recommend:
- a few lead-only teams
- a temporary workflow group
- a coordination plan over NATS and managed exchange
- target MCP/tool bindings for each lane, plus missing-tool setup steps when needed

Good launch prompt:

```text
Create the smallest useful team for this outcome: draft an investor-ready product demo checklist, produce a one-page summary, and identify user-testing risks. Use a Team Lead, Architect Prime, and focused builder unless you can explain why a fourth specialist is necessary. Keep final outputs visible in chat and retained as artifacts.
```

## Useful Expectations For Testing

When testing team workflows, verify:
- compact team defaults are visible and start lead-only
- explicit specialist-output requests preserve the requested roles instead of collapsing them into a ceremonial single-member shell
- media team requests produce or degrade a retained media deliverable with output/proof references
- P0 generated-game proof uses a natural Soma/team request, not a pasted finished HTML file; the retained package should open as a browser app with code-generated graphics, movement, collision, hazards/enemies, health, key, locked door, win/fail states, restart, support files (`README.md`, `PROOF.md`, `project-package.json`), and a Resources link to the package folder
- any temporary specialist explains the missing capability, owned task, proof expected, and removal point
- broad work splits into several smaller lead-owned lanes instead of starting a large roster
- broad asks produce multiple small coordinated lanes instead of one huge team
- team-only creation does not imply the team is already executing; the Dashboard first-deliverable launcher should seed a bounded Soma ask and leave final send/approval with the operator
- the team lead is the first visible operational counterpart
- Dashboard Active Work remains capped and points to `/teams` for the full durable backlog
- archived or cleared work does not appear in the Dashboard review queue or `/teams?view=work`, but retained history remains available outside the active review lane
- Review Work shows the queue summary and concise `Reason` / `Trust` / `Move` labels before team context, setup, or roster content
- Dashboard current-work lane shows one obvious next action while keeping any latest output openable
- Ask Team or Respond creates a durable follow-on work item, shows queued state immediately, keeps the UI usable, then visibly returns output-ready or degraded state
- degraded team asks name timeout/offline/unreadable-response proof, recovery options, and what remains trusted
- raw input/delivery subjects, models, prompts, and tool ids stay behind Advanced/Inspect instead of default team cards
- retained outputs remain reviewable in `/groups` after a temporary collaboration is closed
- the group **Workflow Log** shows work items, run/proof cues, retained output cues, and recovery/degraded fetch guidance without exposing raw bus subjects by default
- group tab changes are deliberate clicks or keyboard tab navigation, and the current group/panel can be reopened from the URL during handoff
- completed, stale, or test groups have a visible **Clear group** action, and retained output files are removed only when the operator explicitly includes them
- event-driven actuation is configured through `/automations` Trigger Rules, not through the team workspace itself
