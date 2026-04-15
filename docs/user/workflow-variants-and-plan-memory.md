# Workflow Variants And Plan Memory
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [User Docs Home](README.md)

Use this page when you want to decide whether a request should stay as a direct Soma ask, become a compact team, or split into several lanes, and when you want the plan to survive a full environment reboot.

## TOC

- [Choose The Smallest Useful Workflow](#choose-the-smallest-useful-workflow)
- [Variant 1: Direct Soma](#variant-1-direct-soma)
- [Variant 2: One Context-Rich Agent](#variant-2-one-context-rich-agent)
- [Variant 3: Compact Team](#variant-3-compact-team)
- [Variant 4: Multi-Lane Workflow](#variant-4-multi-lane-workflow)
- [When Teams Truly Win](#when-teams-truly-win)
- [How To Keep A Plan Through A Reboot](#how-to-keep-a-plan-through-a-reboot)
- [Good Operator Habit](#good-operator-habit)

## Choose The Smallest Useful Workflow

Use the smallest workflow that changes the outcome for the better.

That usually means:
- direct Soma for one good answer
- one context-rich agent for one complex output
- a compact team for planning, making, and reviewing
- a multi-lane workflow for broad work with distinct outputs or handoffs

The goal is not to create more agents.
The goal is to choose the shape that stays clear, inspectable, and resumable.

## Variant 1: Direct Soma

Best for:
- one answer
- one recommendation
- one short draft
- one bounded explanation

Use it when:
- you do not need a managed delivery lane
- the work has one main output
- restarting the session would not be costly

## Variant 2: One Context-Rich Agent

Best for:
- one complex brief
- one architecture recommendation
- one long synthesis pass
- one patch or rewrite in a tight scope

Use it when:
- the work still has one main reasoning line
- the main challenge is context depth, not role separation

## Variant 3: Compact Team

Best for:
- one output package with planning plus production
- implementation plus test or review
- launch brief plus risk check
- media concept plus quality review

Typical shape:
- Team Lead
- Architect Prime
- Focused Builder
- optional Reviewer or Domain Specialist

Use it when:
- planning, making, and checking should be separate
- the operator should see who owns the result
- you want retained outputs and a clearer resume path

## Variant 4: Multi-Lane Workflow

Best for:
- broad asks
- several outputs
- explicit handoffs
- work that should keep separate planning, build, and review lanes

Examples:
- research lane, synthesis lane, review lane
- planning lane, implementation lane, validation lane
- concept lane, generation lane, approval lane

Use it when one giant thread or one giant team would become harder to understand than the work itself.

## When Teams Truly Win

Teams show a real advantage when:
- different roles need different incentives
- several lines of work can run in parallel
- one lane should create while another should verify
- the work has several deliverables
- the operator needs visible handoffs and recovery
- you want to resume after interruption without reconstructing everything from chat

Teams are not automatically better.
They are better when the structure of the work matters.

## How To Keep A Plan Through A Reboot

If the plan matters enough that you would not want to lose it after a full environment reboot, do not leave it only in chat.

Use at least one durable surface:

- **Temporary continuity** for restart-safe working checkpoints
- **Retained artifacts** for plan summaries, checklists, briefs, and output contracts
- **Run Timeline** for the execution trace and resume context
- **Temporary workflow groups** when you want lane ownership and archived output review
- **Conversation templates** when the ask shape should be reused later
- **Governed memory or deployment context** only when the material should influence future work beyond this one mission

Simple rule:
- working state belongs in temporary continuity
- deliverables belong in retained artifacts
- reusable lessons belong in the right memory layer

## Good Operator Habit

When a plan becomes important, ask Soma for a durable package such as:

```text
Turn this into a compact execution plan with named outputs, acceptance criteria, and a retained artifact summary so we can resume after a reboot.
```

For broader work:

```text
Split this into planning, build, and review lanes. Keep each lane compact, retain the output contract for each lane, and make the plan resumable if the environment restarts.
```

Helpful related docs:
- [Teams](teams.md)
- [Memory](memory.md)
- [Run Timeline](run-timeline.md)
- [Meta-Agent & Blueprints](meta-agent-blueprint.md)
- [V8 Teamed Agentry Workflow Advantage](../architecture-library/V8_TEAMED_AGENTRY_WORKFLOW_ADVANTAGE.md)
