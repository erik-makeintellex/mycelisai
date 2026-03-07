# Using Soma Chat

> Workspace-first interaction model: you send intent, Soma orchestrates execution.

---

## Overview

Open `Workspace` (`/dashboard`) and type naturally.
Soma receives every message and coordinates the rest.
Soma operates as a symbiote execution partner: it should execute and deliver outcomes,
not instruct you step-by-step on how to do the work manually.

Display-name customization:
- open `Settings -> Profile`
- set **Assistant Name**
- save to update Workspace/status labels that normally show "Soma"

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

4. **Inline image outputs**
- if a response generates an image, it is rendered directly in chat
- generated images are cache-first and expire after 60 minutes unless saved
- use the inline `Save` action or ask Soma to save it (for example: "save this image to saved-media")

---

## Execution-First Contract

Soma is expected to execute, not just explain, whenever execution is available.

Preferred path:
1. use internal capabilities
2. use onboarded MCP tools when they are the shortest safe path
3. propose/confirm for governed mutation paths

If a tool call fails, Soma should recover inline (retry, reroute, or proposal fallback) without making you retype the request.

Direct drafting behavior:
- if you ask for plain chat content such as a short letter, email, note, or message,
  Soma should answer with the text directly in chat
- it should not route that request through file tools, local commands, or council delegation
  unless you explicitly ask to save, inspect, execute, or hand the work off

Execution guardrail:
- if Soma responds with planning-only language (for example "Step 1" / "we need to delegate") on an actionable request,
  the runtime triggers one policy-correction pass to force a tool call or a concrete blocker response.

Root-admin configuration behavior:
- if you ask Soma to configure Mycelis, it should execute against the relevant configuration surface
  (brains/providers, profiles, governance policy, MCP, users/groups, runtime settings) rather than
  limiting itself to "create a new team" flows.
- governed mutations still use proposal/confirm gates where required.

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

When crew creation is needed, Soma should build a lean team oriented to:
1. plan
2. develop
3. verify
4. deliver

---

## Operational Helpers

While chatting, you can use:
- **Status Drawer** (global health visibility)
- **Degraded Mode Banner** actions
- **Focus Mode** (`F`) to prioritize chat height
- **Advanced Mode** toggle (Settings footer) to show/hide high-density telemetry surfaces

---

## Good Prompting Practices

- be explicit about desired outputs
- reference recent context ("continue from step 2")
- review delegation trace to understand specialist contributions
- confirm only when proposal intent matches your goal
