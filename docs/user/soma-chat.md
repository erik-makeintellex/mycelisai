# Using Soma Chat

> Workspace-first interaction model: you send intent, Soma orchestrates execution.

---

## Overview

Open `Workspace` (`/dashboard`) and type naturally.
Soma receives every message and coordinates the rest.

```
You type -> Soma reasons (ReAct, up to 10 iterations)
         -> optional council consultation
         -> answer and/or governed proposal
```

---

## Sending Messages

1. Go to `Workspace` (`/dashboard`).
2. Enter your intent in the bottom input.
3. Press `Enter` or click send.

Live activity text indicates current steps (thinking, consulting, searching memory, invoking tools).

---

## Reading Responses

Soma responses can include:

1. **Primary answer**
- markdown text, code blocks, links, and tables

2. **Delegation Trace**
- compact cards showing which council members were consulted

3. **Proposal block (mutation paths)**
- explicit action preview with confirm/cancel

No mutation executes until you confirm.

---

## Council Failure Recovery

If a council call fails, Workspace shows a structured error card instead of a raw error.

The card includes:
- what failed
- likely cause
- next actions

Available actions:
- `Retry`
- `Switch to Soma`
- `Continue with Soma Only`
- `Copy Diagnostics`

This keeps recovery inline without retyping or page switching.

---

## Direct Council Access

To send directly to a specialist:

1. Click `Direct` in chat header.
2. Pick Architect, Coder, Creative, or Sentry.
3. Send your message.
4. Use `Soma` option to return to default orchestration.

---

## Launch Crew Flow

For multi-step execution:

1. Click `Launch Crew` in Workspace header.
2. Provide mission intent.
3. Review generated proposal.
4. Confirm execution.

On success, a system message includes a run link (`/runs/{run_id}`).

---

## Operational Helpers

While chatting, you can use:
- **Status Drawer** (global health visibility)
- **Degraded Mode Banner** actions
- **Focus Mode** (`F`) to prioritize chat height

---

## Good Prompting Practices

- be explicit about desired outputs
- reference recent context ("continue from step 2")
- review delegation trace to understand specialist contributions
- confirm only when proposal intent matches your goal

