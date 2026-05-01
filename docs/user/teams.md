# Teams
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Soma-first team workflow: start with the smallest precise team, keep the lead visible, and split broad asks into several focused teams when needed.

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

The Teams area is where you review existing teams, open a team lead workspace, and create new teams through Soma.

Default rule:
- teams should be small, explicit, and focused on the output
- the ideal launch team is 3 people: Team Lead, Architect Prime, and one focused builder/developer role
- a single team should stay at 5 members or fewer
- broad requests should become several small teams or lane bundles coordinated by Soma and Council, not one large roster
- the root Soma workspace remains the main place to ask for team creation and orchestration

## Default Launch Shape

For most work, Soma should launch a team like this:

1. Team Lead: the user-facing counterpart who keeps the mission, status, and outputs clear.
2. Architect Prime: the planning role that shapes the approach, dependencies, and output contract.
3. Focused Builder: the role that produces the needed artifact, implementation, analysis, media prompt pack, data review, or delivery payload.

If the work truly needs more coverage, Soma can add up to two more targeted roles:

- Reviewer / Tester: verifies output quality, acceptance criteria, and cleanup.
- Domain Specialist: covers one distinct capability such as media direction, security, data analysis, or integration.

Soma should not launch a broad standing pool just because the request sounds ambitious. If more than 5 people seem useful, Soma should split the work into smaller coordinated teams and explain the lanes.

## Choosing The Right Workflow

Use the root Soma chat when you want the simplest path:

- "Soma, create a compact team to produce a customer-facing launch brief and keep the team to the smallest useful shape."
- "Soma, this is a broad product-review ask; split it into small lanes and tell me what each lane will output."
- "Soma, summarize what teams exist, what they are doing, and where their outputs are."

Use `Teams` when you want to inspect or manage existing teams:

- open the team lead workspace
- review retained outputs
- inspect member templates
- review or edit template role, model, and MCP/internal tool references
- check whether a team should be archived or kept

Use `Settings -> Connected Tools` or `Resources -> Connected Tools` when you need to confirm which tool refs, direct web search posture, or MCP servers are available before assigning them to a reusable agent template. Installed server cards should show the MCP structure, and Library is the reapply/edit path for curated server config.

Use `Create Team` when you want a guided setup instead of filling raw fields:

- describe the outcome
- review Soma's recommended compact shape
- confirm the lead, Architect Prime, and focused builder role
- decide whether the work should stay as one team or split into lanes
- confirm the visible output contract before launch

If you need the higher-level boundary between direct Soma, one context-rich agent, compact teams, and multi-lane orchestration, read [V8 Teamed Agentry Workflow Advantage](../architecture-library/V8_TEAMED_AGENTRY_WORKFLOW_ADVANTAGE.md).
If you want the operator-facing version, including how to keep plans durable through a reboot, read [Workflow Variants And Plan Memory](workflow-variants-and-plan-memory.md).

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
- inspect the current outputs and retained artifacts
- ask the lead to summarize what the team is doing
- use the lead to reach specialists only when needed
- ask for "what changed, what was produced, and what remains" when you need a fast state read

The team lead is the user-facing counterpart for that team, not a hidden extra member list.

## Team Creation

Use the guided team-creation workflow when you want Soma to shape the team for you.

Tell Soma:
- what outcome you need
- how broad the request is
- whether the work should stay as one compact team or split into multiple lanes
- what outputs you want visible at the end
- whether this is a temporary group whose logs can be reviewed and whose outputs should remain after closure

If the request is broad, expect Soma to recommend:
- several small teams
- a temporary workflow group
- a coordination plan over NATS and managed exchange

Good launch prompt:

```text
Create the smallest useful team for this outcome: draft an investor-ready product demo checklist, produce a one-page summary, and identify user-testing risks. Use a Team Lead, Architect Prime, and focused builder unless you can explain why a fourth specialist is necessary. Keep final outputs visible in chat and retained as artifacts.
```

## Useful Expectations For Testing

When testing team workflows, verify:
- compact team defaults are visible and ideally start at 3 members
- any team with 4 or 5 members explains the extra role
- no single team exceeds 5 members; broad work should split into several smaller lanes
- broad asks produce multiple small coordinated lanes instead of one huge team
- the team lead is the first visible operational counterpart
- retained outputs remain reviewable after a temporary team is closed
