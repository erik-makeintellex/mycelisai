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
- `/groups` is for retained outputs, archived collaboration records, and reviewable temporary-workflow history after the active collaboration is closed.
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

- review the Active Work Lane to see whether a team is new, queued, running, output-ready, degraded, paused, or waiting on the operator
- treat the Dashboard Active Work lane as an attention-first slice; use `Teams` for the full durable backlog
- use the Dashboard `Work contexts` strip when you want to switch Soma into a specific team's focused chat/output/proof lane without leaving the main workbench
- open the team lead workspace
- review current outputs while the team is active
- use Ask Team or Respond on a durable active-work row to request a bounded follow-on output or supply missing input without opening raw bus details
- use Inspect, Steer, Start, Pause, Resume, or Archive controls when they are enabled for the current team state
- inspect member templates
- review or edit template role, model, and MCP/internal tool references
- check whether a team should be archived or kept

When a team has just been created and no delegated work item exists yet, the Dashboard shows a first-deliverable launcher instead of treating the team shell as active work. Choose a starter such as `Build playable prototype`, `Write design brief`, or `Draft delivery plan`; Soma places the bounded ask in the chat input for review, then your send creates the governed work item that can run, produce output, and attach proof.

Use `Groups` when you want to review retained outputs or collaboration records after a temporary workflow has been archived. Standing groups and Soma-created runtime-team groups also have a dedicated workspace folder under `MYCELIS_WORKSPACE/groups/...`, visible from the group detail pane with an `Open folder` action.

Use `Automations` when you want event rules to actuate work, route proposals, require approval before execution, or author propose-only Schedule Rules for reviewable cadence.

Use `Settings -> Connected Tools` or `Resources -> Connected Tools` when you need to confirm which tool refs, direct web search posture, or MCP servers are available before assigning them to a reusable agent template. Installed server cards should show the MCP structure, and Library is the reapply/edit path for curated server config.

If Soma recommends tools that are not installed yet, it should walk you through the enablement path before launch: name the missing MCP server, name required `.env` variables without exposing secret values, point to the Connected Tools Library, and then bind the resulting tool refs to the team or reusable member template after you confirm.

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

After a temporary collaboration is archived, use `/groups` to review the retained output package and collaboration record.

For permanent or standing groups, keep deliverables inside the group folder instead of the general output folders. Soma-owned team media defaults to `groups/<team-id>/media`, and Soma-owned team project packages default to `groups/<team-id>/generated/...`. Explicit operator paths are still respected when you intentionally name a different workspace-confined target.

## Team Creation

Use the guided team-creation workflow when you want Soma to shape the team for you.

Tell Soma:
- what outcome you need
- how broad the request is
- whether the work should stay as one compact team or split into multiple lanes
- what outputs you want visible at the end
- whether this is a temporary group whose logs can be reviewed and whose outputs should remain after closure

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
- any temporary specialist explains the missing capability, owned task, proof expected, and removal point
- broad work splits into several smaller lead-owned lanes instead of starting a large roster
- broad asks produce multiple small coordinated lanes instead of one huge team
- team-only creation does not imply the team is already executing; the Dashboard first-deliverable launcher should seed a bounded Soma ask and leave final send/approval with the operator
- the team lead is the first visible operational counterpart
- Dashboard Active Work remains capped and points to `/teams` for the full durable backlog
- Ask Team or Respond creates a durable follow-on work item, then visibly returns output-ready or degraded state
- degraded team asks name timeout/offline/unreadable-response proof, recovery options, and what remains trusted
- raw input/delivery subjects, models, prompts, and tool ids stay behind Advanced/Inspect instead of default team cards
- retained outputs remain reviewable in `/groups` after a temporary collaboration is closed
- event-driven actuation is configured through `/automations` Trigger Rules, not through the team workspace itself
