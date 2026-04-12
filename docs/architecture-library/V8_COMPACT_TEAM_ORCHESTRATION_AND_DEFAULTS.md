# V8 Compact Team Orchestration And Defaults
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [Architecture Library Index](ARCHITECTURE_LIBRARY_INDEX.md)

> Status: Canonical
> Last Updated: 2026-04-12
> Purpose: Define the compact-team default for Mycelis so ordinary asks produce small, focused teams while broad asks are split into several smaller coordinated teams or lanes orchestrated by Soma with Council help over NATS and the managed exchange framework.

## TOC

- [Why This Exists](#why-this-exists)
- [Compact By Default](#compact-by-default)
- [Broad Ask Orchestration](#broad-ask-orchestration)
- [Team Shape Guidance](#team-shape-guidance)
- [NATS And Exchange Coordination](#nats-and-exchange-coordination)
- [Operator-Facing Output Rules](#operator-facing-output-rules)
- [Testing Expectations](#testing-expectations)

## Why This Exists

Mycelis should not default to building large standing teams for ordinary requests.

Most operator asks are best served by a small team with a clear lead, a small specialist set, and a short execution path. Large team rosters are harder to inspect, harder to test, and easier to misuse as a substitute for good orchestration.

This document sets the default posture for team creation and broader execution shaping:
- compact teams are the default
- broad asks are decomposed into several small teams or lanes
- Soma remains the root orchestrator
- Council provides specialist review and planning help
- NATS and the managed exchange framework carry coordination and observability

## Compact By Default

Default team shape:
- 1 focused Team Lead
- 1 Architect Prime
- 1 focused builder/developer role matched to the requested output
- up to 2 additional targeted specialists only when the mission output clearly requires distinct capability coverage
- a bounded output contract
- a clear completion path
- a visible archive or retained-output outcome when the work is done

Default guidance:
- start from the 3-person default whenever it can reasonably deliver the requested outcome
- prefer reusable templates over inventing extra members
- keep the team roster readable in the UI
- keep the execution path narrow enough that a user can understand who is doing what
- keep any single launched team at 5 members or fewer

Avoid by default:
- large monolithic teams
- sprawling member lists with overlapping roles
- teams that exist only because a prompt was broad
- hidden internal sub-teams that the operator cannot reason about
- single-team launches with more than 5 members

The practical rule is simple:
- if the request is ordinary, launch the smallest useful 3-person team
- if the request is broad, split the work into several compact teams and coordinate them

## Broad Ask Orchestration

When a request is broad, Mycelis should not create one oversized team.

Instead, Soma should:
1. identify the outcome boundary
2. split the work into a few smaller delivery lanes
3. assign each lane a focused purpose
4. route specialist review through Council when it improves quality or trust
5. coordinate the lanes over NATS and the managed exchange surface
6. keep the operator able to inspect each lane separately

Examples:
- a product-launch request may become a planning lane, a marketing lane, and a review lane
- a research-heavy request may become an intake lane, an analysis lane, and a synthesis lane
- a media-heavy request may become a prompt/planning lane, a generation lane, and a review or publish lane

This is intentionally not a giant team with every discipline in one roster. The orchestration layer should do the work of coordination.

## Team Shape Guidance

Suggested size guidance:
- 3 members: normal ideal default
- 4 to 5 members: allowed only when the output needs a distinct reviewer, domain specialist, or second focused builder
- over 5 members: do not launch as one team; trigger decomposition into multiple compact teams or lanes

Recommended specialization pattern:
- Team Lead: owns the lane and the operator-facing summary
- Architect Prime: turns the operator request into a practical plan, dependencies, and output contract
- Focused Builder: creates the artifact, implementation payload, analysis, media direction, website draft, data review, or other requested deliverable
- Optional Reviewer / Tester: checks acceptance criteria, quality, cleanup, and retained outputs
- Optional Domain Specialist: covers one distinct capability such as media direction, security, data analysis, or integration

When creating templates:
- prefer role templates that can be reused across many compact teams
- avoid baking a large permanent roster into the template
- make the team lead visible as the first operational counterpart
- force any template over 5 members through decomposition or explicit architecture review before it becomes product default

## NATS And Exchange Coordination

Broad execution should be visible as coordinated small-team work, not silent oversized orchestration.

The coordination model should use:
- NATS for directed team inputs, status, and bounded result propagation
- managed exchange for durable, inspectable outputs and reviewable artifacts
- Soma as the root orchestrator that can coordinate the overall plan
- Council as the specialist review layer when a lane needs deeper validation

Coordination expectations:
- each small team should have a clear boundary and a readable output contract
- cross-team handoffs should be explicit
- team-to-team communication should remain observable through the governed bus and exchange surfaces
- archived work should keep retained outputs visible even when the temporary team is removed

## Operator-Facing Output Rules

The UI should make the compact-team story legible.

Operator should see:
- a short team summary rather than a giant roster dump
- the focused team lead as the main touchpoint
- the reason a request was split into multiple lanes when it is broad
- the outputs each lane is expected to produce
- the coordination path between Soma, Council, and the lanes

The UI should not imply:
- that more members is automatically better
- that a broad request must become one giant team
- that the operator needs to micromanage every specialist to get work done

## Testing Expectations

This compact-team rule must be proven at multiple layers:

- backend contract tests should verify ordinary asks stay compact and broad asks choose a multi-team coordination shape
- UI tests should verify the creation workflow teaches compact default teams and shows the broad-ask split into smaller lanes
- NATS/exchange observability tests should verify that multiple compact teams can coordinate and remain inspectable through governed signals and outputs
- browser workflow tests should verify a user can create a focused team, inspect the lead, and see broad requests broken into several small team bundles instead of one oversized roster
- documentation tests and user walkthroughs should show ordinary team creation as Team Lead + Architect Prime + focused builder, with 4th/5th roles justified by output need

The first pass should fail if:
- the default team shape exceeds the compact threshold without an explicit broad-ask reason
- any single team launch exceeds 5 members instead of decomposing
- broad asks silently create one giant team instead of multiple lanes
- coordination cannot be observed in the bus or exchange view
- the UI obscures the team lead or output contract
