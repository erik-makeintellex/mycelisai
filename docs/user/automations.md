# Automations
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Workflow control surface for event trigger rules and governance approvals.

---

## Overview

Open `/automations`.

Current ownership:

- `/teams` is for active team work and team lead workspaces.
- `/groups` is for retained outputs, archived collaboration records, and reviewable temporary-workflow history.
- `/automations` is for event trigger rules, propose-only schedule rules, and approvals around automation behavior.

Current tabs:

| Tab | Purpose |
|-----|---------|
| Active Automations | Actionable hub for current automation setup, trigger-rule entry, and governance review |
| Trigger Rules | Event-driven actuation rules |
| Schedule Rules | Propose-only cadence rules with cooldown, proof expectations, recovery text, and next-run state |
| Approvals | Governance review queue and policy controls |
| Workflow Builder (Advanced) | Lower-level workflow structure editing |

---

## Active Automations (Automation Hub)

This tab is intentionally non-empty. It should orient operators toward the automation surfaces that exist now, rather than presenting scheduler capability as a future placeholder.

Layout:
- **Trigger Rules** for event-driven actuation setup
- **Schedule Rules** for cadence proposals that remain reviewable before execution
- **Approvals** for governed automation decisions
- **Workflow Builder** when Advanced mode is on
- primary CTA for building a governed automation chain

Default chain guidance:
1. Create Trigger
2. Set Propose Mode
3. Route to Team
4. Review Approval
5. Execute

### Mission Profile Wizard

Use **Open Wizard** in Active Automations to launch a guided flow:
1. Objective
2. Profile
3. Readiness
4. Launch

The wizard is the fastest path when you want a governed mission profile scaffolded from intent instead of manual setup.

---

## Trigger Rules

Trigger rules evaluate mission events and other supported signals, then decide whether to route new work or request approval before actuation.

Typical fields:
- name
- event pattern/type
- target mission
- mode (`propose` or `auto_execute` when policy allows)
- cooldown/depth/concurrency guards

Safety posture:
- start in `propose`
- move to automated execution only after validation
- keep active team operation visible in `/teams`
- keep retained output and collaboration records visible in `/groups`

---

## Approvals

Approvals is the governance queue.

You review:
- low-trust or governed mutation proposals
- propose-mode trigger outcomes
- manual mutation proposals from Workspace flows

Core actions:
- approve
- reject
- inspect structured payload details

---

## Workflow Builder (Advanced Mode)

Advanced Mode enables the Workflow Builder tab.
This is the graph-level authoring and editing surface for advanced workflow structure.

---

## Scheduler Status

Cadence authoring is now present as propose-only Schedule Rules. A schedule rule records a rule name, target mission, cadence interval, next proposal time, cooldown, proof expectations, and recovery behavior. Scheduler ticks can record a proposed cadence outcome and update the next run, but this first production slice does not autonomously execute the target mission. Operators should treat Schedule Rules as reviewable cadence intent until the approval/execution trust path is promoted.

---

## What "Healthy" Looks Like

For the current V8 UX baseline, `/automations` should show:
- Active Automations tab content with the Automation Hub
- a clear path into Trigger Rules
- a clear path into Schedule Rules
- a clear path into Approvals
- no misplaced team-readiness tab
- no roadmap placeholder copy for scheduler behavior

If those are missing, verify you are on the current build and then refresh the app session.
