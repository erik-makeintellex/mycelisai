# Automations

> Workflow control surface for trigger rules, approvals, and advanced workflow tools.

---

## Overview

Open `/automations`.

Current tabs:

| Tab | Purpose |
|-----|---------|
| Active Automations | Actionable hub (available now + coming soon scheduler state) |
| Trigger Rules | Event-driven mission routing rules |
| Approvals | Governance review queue and policy controls |
| Shared Teams (Advanced) | Operational shared-team visibility |
| Workflow Builder (Advanced) | Lower-level workflow structure editing |

---

## Active Automations (Automation Hub)

This tab is intentionally non-empty, even when scheduler is not active.

Layout:
- **Available Now** cards (Trigger Rules, Approvals, plus advanced Shared Teams and Workflow Builder when Advanced mode is on)
- **Coming Soon** scheduler panel (roadmap state, not an error)
- primary CTA: **Set Up Your First Automation Chain**

Default chain guidance:
1. Create Trigger
2. Set Propose Mode
3. Route to Team
4. Review Approval
5. Execute

### Team Instantiation Wizard

Use **Open Wizard** in Active Automations to launch a guided flow:
1. Objective
2. Profile
3. Readiness
4. Launch

The wizard is the fastest path when you want a new team scaffolded from intent instead of manual setup.

---

## Trigger Rules

Trigger rules evaluate mission/event activity and decide whether to route new work.

Typical fields:
- name
- event pattern/type
- target mission
- mode (`propose` or `execute` when policy allows)
- cooldown/depth/concurrency guards

Safety posture:
- start in `propose`
- move to automated execution only after validation

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

## Shared Teams (Advanced Mode)

Shared Teams surfaces operational readiness:
- online agent count
- heartbeat recency
- health indicators
- quick actions (chat/runs/wiring/logs)

Use this tab for fast diagnosis when execution paths degrade.

Quick actions on each team card:
- Open chat
- View runs
- View wiring
- View logs

---

## Workflow Builder (Advanced Mode)

Advanced Mode enables the Workflow Builder tab.
This is the graph-level authoring and editing surface for advanced workflow structure.

---

## Scheduler Status

Scheduler capabilities are represented as planned/coming-soon where applicable.
This should appear as roadmap context, not a dead error state.

---

## What "Healthy" Looks Like

For the current V7 UX baseline, `/automations` should show:
- Active Automations tab content with the Automation Hub
- **Set Up Your First Automation Chain** primary CTA
- **Open Wizard** control in the Guided Setup panel

If those are missing, verify you are on the current build and then refresh the app session.
