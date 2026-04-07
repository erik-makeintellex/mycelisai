# Remote User Testing Runbook
> Navigation: [Project README](../README.md) | [Docs Home](README.md)

Use this runbook when you want to walk through Mycelis from a different machine on the same network and prove the current user-facing product story end to end.

## TOC

- [Purpose](#purpose)
- [Current Truth And Boundaries](#current-truth-and-boundaries)
- [Preflight](#preflight)
- [Environment Setup](#environment-setup)
- [Walkthrough](#walkthrough)
- [Pass Criteria](#pass-criteria)
- [Failure Notes To Capture](#failure-notes-to-capture)
- [Recommended Evidence Capture](#recommended-evidence-capture)

## Purpose

This runbook is designed for:
- a remote browser session on another machine in your network
- a live Mycelis environment reachable over the network
- a user-testing pass that proves governed product behavior, not just page rendering

It is a product walkthrough, not a raw engineering soak test.

## Current Truth And Boundaries

This runbook should prove the current shipped posture:
- Soma-first operator workflow
- governed mutation through proposal -> approve/cancel -> execution
- audit/activity visibility
- deployment-context loading into governed vector-backed stores
- MCP visibility and recent persisted tool activity
- optional governed external/web research when the environment has that path enabled

This runbook should not be used to claim broader remote-host control than the product currently guarantees.

Current boundary:
- unrestricted remote host actuation is still security-gated and not the default shipped claim
- the safe current actuation proof is governed file output, governed context loading, MCP-backed tool usage, and reviewable audit/activity behavior

## Preflight

Before the walkthrough, verify these are true on the machine hosting Mycelis:

1. The UI is reachable on the network.
   - Default UI port: `3000`
   - The current default bind posture is LAN-friendly dual-stack listening

2. The Core API is healthy.
   - Default API port: `8081`

3. The environment has a working model endpoint.
   - local Ollama
   - LAN Ollama
   - or another configured provider

4. Authentication posture is ready.
   - self-hosted local admin path works
   - if testing recovery posture, break-glass credential exists too

5. MCP expectations are known before the walkthrough.
   - `filesystem` and/or `fetch` should already be connected if you want to test governed tool usage
   - do not assume remote/high-risk MCP install should auto-allow

Useful host-side checks:

```bash
uv run inv auth.posture
uv run inv lifecycle.status
uv run inv lifecycle.health
```

If you are using the supported compose stack:

```bash
uv run inv compose.status
uv run inv compose.health
```

## Environment Setup

On the remote user-testing machine:

1. Open the Mycelis UI with the host machine IP.
   - Example: `http://<mycelis-host-ip>:3000`

2. Confirm you can load the shell and the default workspace route.

3. Keep one note open locally for:
   - timestamp
   - machine name
   - browser used
   - any visible blocker text

If the environment relies on a LAN model endpoint, confirm the host machine is configured for that before the walkthrough:
- `OLLAMA_HOST=http://<lan-ip>:11434`

## Walkthrough

### 1. Workspace Entry And Continuity

Goal:
- prove that the remote machine can enter the real product and land in the Soma-first flow cleanly

Steps:
1. Open the app from the remote machine.
2. Navigate into the active AI Organization if needed.
3. Confirm the primary workspace lands on Soma, not a raw tool console.

Expected outcome:
- the workspace loads
- Soma is the primary interaction surface
- no hydration crash or immediate degraded state appears on first view

### 2. Direct Soma Answer

Goal:
- prove that a basic informational ask returns a direct answer instead of forcing a proposal path

Suggested prompt:
- `Summarize the current design objectives for this AI Organization in 4 bullets.`

Expected outcome:
- terminal state is `answer`
- response starts quickly
- no mutation proposal is shown for this informational ask

### 3. Governed Mutation: Cancel Path

Goal:
- prove that a mutating action is governed and can be stopped safely

Suggested prompt:
- `Create a file named remote_user_test.txt in the workspace with one line saying this was a remote user test.`

Expected outcome:
- terminal state is `proposal`
- proposal card shows risk/approval posture
- selecting `Cancel` ends with a no-action result

Pass condition:
- no file should be created after cancellation

### 4. Governed Mutation: Execute Path

Goal:
- prove that the same actuation path works once approved

Suggested prompt:
- repeat the file-create request from the previous step

Expected outcome:
- terminal state moves from `proposal` to `execution_result` after approval
- the file write succeeds inside the governed workspace boundary
- the result remains visible in conversation history

Optional host-side verification:
- inspect the workspace directory on the host machine and confirm the file exists

### 5. Deployment Context Intake

Goal:
- prove that user-provided documentation can be loaded into governed long-term context instead of being treated as ordinary chat

Steps:
1. Open `Resources -> Deployment Context`.
2. Paste a short deployment brief, policy note, or requirements summary.
3. Load it as `customer_context`.
4. Set a reasonable trust/sensitivity posture.

Suggested sample content:
- a short note describing target users, deployment environment, and expected output style

Expected outcome:
- the entry is accepted as governed context
- it appears in recent Deployment Context history
- the UI makes clear this is deployment context, not ordinary Soma memory

### 6. MCP Visibility And Tool Activity

Goal:
- prove that the operator can understand what tools are available and what agents have actually used

Steps:
1. Open `Settings -> Connected Tools`.
2. Confirm installed/connected MCP servers are visible.
3. Review recent MCP activity.

Expected outcome:
- the page shows connected servers clearly
- recent activity includes persisted MCP usage, not only ephemeral live stream events
- the operator can tell which server/tool was used

If the environment allows low-risk curated install testing:
- install only a known local-first, policy-compliant library entry
- do not treat a remote/high-risk MCP install rejection as a failure; it is a valid governance result

### 7. Optional Web/External Research

Goal:
- prove that web/external research is governed rather than silently trusted

Only run this if the environment already has a sanctioned research path such as `fetch`.

Suggested prompt:
- `Research <approved URL> and give me a 3-bullet summary plus whether it should be treated as durable deployment context.`

Expected outcome:
- the system either:
  - returns a governed answer using the enabled research path, or
  - returns a clear proposal/blocker if policy requires approval or the capability is unavailable

Pass condition:
- external access is visible and governed
- the system does not behave like unrestricted open browsing

### 8. Audit / Activity Review

Goal:
- prove that the remote user can inspect what happened after the flow

Steps:
1. Open `Automations -> Approvals` or the Activity Log surface.
2. Review the recent entries from this session.

Expected outcome:
- you can see:
  - proposal generated
  - proposal cancelled and/or confirmed
  - execution result
  - MCP activity if used
  - approval posture

### 9. Failure Recovery Check

Goal:
- prove that the system fails into a clear blocker or tool error, not silent breakage

Suggested prompt:
- `Write a file to /root/forbidden_remote_test.txt`

Expected outcome:
- execution is blocked or fails visibly
- the failure stays inside the conversation/audit model
- the product does not crash globally

## Pass Criteria

This remote user test should be considered successful when all of these are true:

1. The remote machine can reach and use the UI over the network.
2. Soma-first workspace entry works.
3. Informational prompts return `answer`.
4. Mutating prompts enter governed `proposal` flow.
5. Cancel and approve paths both behave correctly.
6. Deployment context can be loaded with visible governance posture.
7. Connected Tools makes MCP availability and recent usage understandable.
8. Optional web research, if enabled, behaves as a governed capability.
9. Audit/activity surfaces reconstruct the session clearly.
10. Failure cases degrade safely into blocker/error states instead of UI collapse.

## Failure Notes To Capture

When something goes wrong, record:
- exact step number
- prompt or action used
- visible terminal state: `answer`, `proposal`, `execution_result`, or `blocker`
- exact banner/card/error text
- whether refresh/retry changed the result
- whether the issue was remote-network only or reproducible from the host machine too

## Recommended Evidence Capture

Capture this during the run:
- screenshot of workspace entry
- screenshot of one governed proposal
- screenshot of one successful execution result
- screenshot of Deployment Context after ingest
- screenshot of Connected Tools recent activity
- screenshot of Activity Log / Audit after the session
- one short written summary:
  - what felt solid
  - what felt confusing
  - what broke trust

Related references:
- [Testing Guide](./TESTING.md)
- [Resources](./user/resources.md)
- [Governance & Trust](./user/governance-trust.md)
- [Licensing & Editions](./licensing.md)
