# Teams
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Soma-first team workflow: start compact, keep the lead visible, and split broad asks into several focused teams when needed.

---

## Overview

The Teams area is where you review existing teams, open a team lead workspace, and create new teams through Soma.

Default rule:
- teams should be small and focused unless the request is genuinely broad
- broad requests should usually become several small teams or lane bundles coordinated by Soma and Council
- the root Soma workspace remains the main place to ask for team creation and orchestration

## What A Good Team Looks Like

Most teams should have:
- one clear lead
- a small specialist set
- a narrow mission
- readable outputs

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

## Working With A Team Lead

When you open a team:
- start with the team lead
- inspect the current outputs and retained artifacts
- ask the lead to summarize what the team is doing
- use the lead to reach specialists only when needed

The team lead is the user-facing counterpart for that team, not a hidden extra member list.

## Team Creation

Use the guided team-creation workflow when you want Soma to shape the team for you.

Tell Soma:
- what outcome you need
- how broad the request is
- whether the work should stay as one compact team or split into multiple lanes
- what outputs you want visible at the end

If the request is broad, expect Soma to recommend:
- several small teams
- a temporary workflow group
- a coordination plan over NATS and managed exchange

## Useful Expectations For Testing

When testing team workflows, verify:
- compact team defaults are visible and reasonable
- broad asks produce multiple small coordinated lanes instead of one huge team
- the team lead is the first visible operational counterpart
- retained outputs remain reviewable after a temporary team is closed

