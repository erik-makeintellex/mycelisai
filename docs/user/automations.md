# Automations

> Workflow control surface for triggers, approvals, teams, and advanced wiring.

---

## Overview

Open `/automations`.

Current tabs:

| Tab | Purpose |
|-----|---------|
| Active Automations | Actionable hub (available now + coming soon scheduler state) |
| Draft Blueprints | Deferred/placeholder workflow guidance |
| Trigger Rules | Event-driven mission routing rules |
| Approvals | Governance review queue and policy controls |
| Teams | Operational team visibility |
| Neural Wiring (Advanced) | Low-level wiring canvas and graph editing |

---

## Active Automations (Automation Hub)

This tab is intentionally non-empty, even when scheduler is not active.

Layout:
- **Available Now** cards (Trigger Rules, Approvals, Teams, Neural Wiring when advanced)
- **Coming Soon** scheduler panel (roadmap state, not an error)
- primary CTA: **Set Up Your First Automation Chain**

Default chain guidance:
1. Create Trigger
2. Set Propose Mode
3. Route to Team
4. Review Approval
5. Execute

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

## Teams

Teams tab surfaces operational readiness:
- online agent count
- heartbeat recency
- health indicators
- quick actions (chat/runs/wiring/logs)

Use this tab for fast diagnosis when execution paths degrade.

---

## Neural Wiring (Advanced Mode)

Advanced Mode enables the Neural Wiring tab.
This is the graph-level authoring/editing surface for agent/team wiring and execution topology.

---

## Scheduler Status

Scheduler capabilities are represented as planned/coming-soon where applicable.
This should appear as roadmap context, not a dead error state.

