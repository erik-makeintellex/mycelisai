# Automations

> Schedule missions, define trigger rules, manage proposals and team approvals.

---

## Overview

The **Automations** page (`/automations`) consolidates everything related to *when* and *how* missions execute automatically. It has five tabs:

| Tab | Purpose |
|-----|---------|
| **Active** | Currently running and recently completed missions |
| **Drafts** | Saved but not yet activated mission blueprints |
| **Triggers** | Rules that fire missions in response to events |
| **Approvals** | Proposals waiting for human sign-off |
| **Teams** | Standing teams and their current status |

---

## Active Missions

Shows missions currently executing or recently completed.

- Each card displays: mission name, status, start time, run count
- Click a mission to open its most recent run timeline
- **Status indicators:** `running` (pulsing cyan), `completed` (green), `failed` (red)

---

## Drafts

Mission blueprints that have been saved but not activated.

- **Edit**: modify the blueprint (name, goal, team configuration)
- **Activate**: promote to an active mission
- **Delete**: remove the draft permanently

Drafts are created when you negotiate a mission in the Wiring canvas and save without activating, or when Soma generates a blueprint that you choose to defer.

---

## Trigger Rules

Triggers allow missions to fire automatically in response to system events.

### Creating a Trigger

1. Click **+ New Trigger**
2. Configure the rule:

| Field | Description |
|-------|-------------|
| **Name** | Human-readable identifier |
| **Event Type** | The event that fires this trigger (e.g., `mission.completed`, `artifact.created`) |
| **Source Filter** | Optional — only fire if the event came from a specific mission or agent |
| **Target Mission** | Which mission to activate when the rule fires |
| **Mode** | `propose` (requires human approval) or `execute` (fires automatically) |
| **Max Depth** | Recursion guard — prevent trigger chains from looping (default: 3) |

### Trigger Modes

- **Propose** (default, safe): When the rule matches, a proposal is created in the Approvals tab — you review and confirm before anything runs.
- **Execute** (automatic): The target mission fires immediately when the event matches. Requires explicit policy allowance. Use with caution.

### Trigger Safety Guards

The trigger engine enforces two automatic guards regardless of mode:

- **Max Depth**: Prevents recursive trigger chains from running indefinitely
- **Max Active Runs**: Prevents runaway concurrency if triggers fire rapidly

---

## Approvals

Proposals waiting for your review before execution. These come from:

- **Governance halts** — agent outputs that fell below the trust threshold
- **Propose-mode triggers** — triggers that matched but require sign-off
- **Manual proposals** — Soma-generated proposals you haven't confirmed yet

### Reviewing a Proposal

Each proposal card shows:
- **Action description** — what will happen
- **Source** — which agent or trigger generated it
- **Risk level** — Low / Medium / High based on action type
- **Payload preview** — the structured action details

Actions:
- **Approve** — confirms the proposal and activates it
- **Reject** — discards the proposal and logs the decision
- **View details** — expands the full proposal JSON

### Policy Management

The **Policy** sub-section of Approvals lets you configure governance rules:
- Trust thresholds per agent role
- Auto-execute allowances for specific action types
- Per-team risk profiles

---

## Teams

Standing teams running in the background.

| Column | Description |
|--------|-------------|
| **Name** | Team identifier |
| **Status** | `active`, `idle`, `offline` |
| **Members** | Agent count |
| **Last Active** | Timestamp of most recent activity |

Standing teams (Admin, Council) are always running while the Core service is up. Mission-scoped teams appear here during execution and disappear when the run completes.

---

## Scheduled Missions (Advanced)

Schedule recurring mission executions using cron syntax:

1. Open a mission in Active or Drafts
2. Click **Schedule**
3. Enter a cron expression (e.g., `0 9 * * 1-5` for weekdays at 9am)
4. The scheduler creates runs at the specified intervals

The scheduler is **NATS-aware** — it suspends automatically if the message bus goes offline and resumes when connectivity is restored. Missed runs are not backfilled.
