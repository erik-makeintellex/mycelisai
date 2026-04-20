# Remote User Testing Runbook
> Navigation: [Project README](../README.md) | [Docs Home](README.md)

Use this runbook when you want to walk through Mycelis from a different machine on the same network and prove the current user-facing product story end to end.

## TOC

- [Purpose](#purpose)
- [Current Truth And Boundaries](#current-truth-and-boundaries)
- [Preflight](#preflight)
- [Environment Setup](#environment-setup)
- [Windows Self-Hosted Operator Lane](#windows-self-hosted-operator-lane)
- [Walkthrough](#walkthrough)
- [Pass Criteria](#pass-criteria)
- [Failure Notes To Capture](#failure-notes-to-capture)
- [Recommended Evidence Capture](#recommended-evidence-capture)

## Purpose

This runbook is designed for:
- a remote browser session on another machine in your network
- a live Mycelis environment reachable over the network
- a user-testing pass that proves governed product behavior, not just page rendering

Use [V8 Workflow Variants And Reboot Proof Set](./architecture-library/V8_WORKFLOW_VARIANTS_AND_REBOOT_PROOF_SET.md) when you specifically want to compare direct Soma, compact-team, and multi-lane workflow behavior for the same objective and verify the resume path after a full reboot.

It is a product walkthrough, not a raw engineering soak test.

## Current Truth And Boundaries

This runbook should prove the current shipped posture:
- Soma-first operator workflow
- Soma as the root admin workspace for creating and shaping teams
- a dedicated Groups workspace for temporary and standing collaboration lanes
- team-specific follow-through through a focused Team Lead workspace once a lane is selected
- governed mutation through proposal -> approve/cancel -> execution
- audit/activity visibility
- deployment-context loading into governed vector-backed stores
- admin-visible output-model routing for team delivery types
- MCP visibility and recent persisted tool activity

Trusted memory boundary:
- governed doctrine and deterministic evidence outrank team-shared `AGENT_MEMORY`
- team-shared `AGENT_MEMORY` outranks Soma personal continuity for team execution
- Soma personal continuity may still shape local style or relationship preference inside scope
- optional governed external/web research when the environment has that path enabled
- the same user journey must also work from a Windows browser against a self-hosted runtime while the AI engine runs on a Windows GPU host addressed by explicit IP or hostname

This runbook should not be used to claim broader remote-host control than the product currently guarantees.

Current boundary:
- unrestricted remote host actuation is still security-gated and not the default shipped claim
- the safe current actuation proof is governed file output, governed context loading, MCP-backed tool usage, and reviewable audit/activity behavior

## Preflight

Before the walkthrough, verify these are true on the machine hosting Mycelis:

1. The UI is reachable on the network.
   - Default UI port: `3000`
   - The current default bind posture is LAN-friendly dual-stack listening
   - If the stack is running inside WSL on the same Windows machine the operator is using, prove the Windows browser path through `http://localhost:3000` first before treating LAN reachability as the only valid access path

2. The Core API is healthy.
   - Default API port: `8081`

3. The environment has a working model endpoint.
   - local Ollama
   - LAN Ollama
   - or another configured provider
   - for Windows self-hosted validation, record the explicit Windows GPU host IP or hostname used by the model service and do not validate against loopback

4. Authentication posture is ready.
   - self-hosted local admin path works
   - if testing recovery posture, break-glass credential exists too

5. MCP expectations are known before the walkthrough.
   - `filesystem` and/or `fetch` should already be connected if you want to test governed tool usage
   - do not assume remote/high-risk MCP install should auto-allow

6. Output storage expectations are known before the walkthrough.
   - if testing `local_hosted` output storage, confirm the host path exists before Compose bring-up
   - if testing cluster-generated output storage, confirm the chart/runtime is using the PVC-backed default

7. Media readiness is known before the walkthrough.
   - if no media provider is configured, capture the missing-provider blocker explicitly
   - if a media provider is configured, be ready to prove render/save/download behavior from the browser

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

1. Open the Mycelis UI with the operator-facing host path.
   - Same Windows machine talking to a WSL-hosted stack: `http://localhost:3000`
   - Different machine on the same network: `http://<mycelis-host-ip>:3000`

2. Confirm you can load the shell and the default workspace route.

3. Keep one note open locally for:
   - timestamp
   - machine name
   - browser used
   - any visible blocker text

If the environment relies on a LAN model endpoint, confirm the host machine is configured for that before the walkthrough:
- `OLLAMA_HOST=http://<lan-ip>:11434`

If the environment relies on a Windows GPU inference host, confirm the browser-facing deployment references an explicit host or IP before the walkthrough:
- `MYCELIS_COMPOSE_OLLAMA_HOST=http://<windows-ai-host>:11434`
- or the equivalent provider-specific endpoint override used by the deployment

## Windows Self-Hosted Operator Lane

Use this lane when the operator is on Windows and the product is running as a self-hosted deployment:

1. Open the UI from the Windows browser using the real operator-facing address.
   - Same machine, WSL-hosted stack: `http://localhost:3000`
   - Different machine or explicit LAN proof: use the network-reachable host name or IP
2. Confirm the root Soma workspace loads with a healthy runtime and a direct `answer` path for informational prompts.
3. Confirm a mutating prompt enters `proposal` and can be approved or cancelled.
4. Confirm guided team creation or temporary workflow launch completes without forcing a local-only dev shortcut.
5. Confirm archived temporary groups retain their outputs and stay reviewable after refresh.
6. Confirm a missing or unreachable Windows AI host produces a visible blocker, not a silent fallback to `localhost`.
7. Restore the AI host and confirm the same browser session can continue working without a full reinstallation or desktop-only reset.

## Walkthrough

### 1. Workspace Entry And Continuity

Goal:
- prove that the remote machine can enter the real product and land in the Soma-first flow cleanly

Steps:
1. Open the app from the remote machine.
2. Navigate into the active AI Organization if needed.
3. Confirm the primary workspace lands on Soma, not a raw tool console.
4. If the host was just rebuilt or upgraded, do one hard refresh before starting the scripted prompts so old conversation state does not masquerade as a live runtime failure.

Expected outcome:
- the workspace loads
- Soma is the primary interaction surface
- the root workspace makes it clear that admin can ask Soma to shape or create teams from here
- a live interaction stream is visible on the Soma home and can be filtered by multiple teams and output aspects
- no hydration crash or immediate degraded state appears on first view

### 2. Direct Soma Answer

Goal:
- prove that a basic informational ask returns a direct answer instead of forcing a proposal path

Suggested prompts:
- `what is your current state`
- `what teams currently exist`
- `Summarize the current design objectives for this AI Organization in 4 bullets.`

Expected outcome:
- terminal state is `answer`
- response starts quickly
- the first two prompts return a deterministic runtime/roster summary instead of a generic provider apology
- no mutation proposal is shown for this informational ask

### 3. Soma Creates Or Refines A Team

Goal:
- prove that the root Soma workspace can be used to create or reshape a team without forcing the user into a low-level infrastructure view

Suggested prompt:
- `Create a marketing team focused on campaign planning, launch messaging, and asset review.`

Expected outcome:
- Soma responds in a team-shaping posture
- the result clearly frames the requested team as a governed working lane, not just a loose chat answer
- the user can move from Soma's root workspace into a more focused team view afterward

### 3a. Groups Workspace

Goal:
- prove that collaboration groups have a dedicated workflow lane instead of crowding the root admin home

Steps:
1. Open `Groups`.
2. Confirm the page clearly separates standing and temporary groups.
3. Create or review a temporary group.
4. Confirm the selected group shows:
   - focused lead/work mode
   - recent outputs
   - broadcast controls
   - quick links back to Soma home and attached team leads
5. Archive the temporary group.
6. Confirm it moves from `Temporary groups` to `Archived temporary groups`.
7. Confirm retained outputs remain reviewable and downloadable after archive.
8. Confirm the group review surface shows how many outputs and contributing leads are currently attached to that lane.
8. Confirm new broadcast is no longer available from the archived group view.

Expected outcome:
- groups are managed in their own graceful interface
- temporary groups read as time-bounded working lanes, not permanent teams
- outputs are reviewable from the group lane without dropping into raw logs by default
- archived temporary groups remain reviewable after closure instead of disappearing into raw logs
- retained outputs stay available without treating archived groups like active coordination lanes
- the group review surface gives a quick count of outputs and contributing leads before the operator reads individual artifacts

### 4. Team Creation And Team Lead Focus

Goal:
- prove that team creation happens through a guided workflow first, and then team work happens through a focused lead instead of a generic global surface

Steps:
1. Open `Teams`.
2. Confirm the page shows both:
   - the available team roster and lead-entry surface
   - the reusable team-member templates Soma can use when specializing new teams
3. Open `Open guided team creation`.
4. Confirm the workflow explains:
   - organization context
   - expected outputs
   - guided Soma handoff
5. Use one of the guided starter prompts or enter a custom team request.
6. Confirm Soma returns a team-shaping response with visible next steps or execution-path framing.
7. If a native execution path is suggested, use `Create temporary workflow group`.
8. Confirm the success state links into `Groups` for the newly created workflow group.
9. Open that group and confirm it is already selected in the Groups workspace.
10. Confirm the selected group shows multiple outputs or at least a clear output/contributing-lead summary when outputs exist.
11. Archive the temporary workflow group.
12. Confirm retained outputs remain reviewable/downloadable after archive and that broadcast controls are no longer available.
13. Return to `Teams`.
14. Select the created or existing team.
15. Confirm the page frames that lane around a focused Team Lead counterpart.
16. Ask a short team-specific question such as `Summarize this team's job in one sentence.`

Expected outcome:
- the Teams page reads as the admin surface for team specialization defaults and lead-entry, not the place to manually assemble a raw team form
- the detailed creation flow happens on its own page instead of being buried in the roster
- guided team creation can launch a temporary workflow group directly instead of forcing the user back into raw group fields
- the newly launched group can be opened directly in the Groups workspace
- the launched temporary workflow can be archived while preserving retained outputs for later review
- live-backend validation should also prove that backend-stored outputs appear in the same retained group review surface after temporary-lane archive
- the workspace or drawer clearly identifies a focused Team Lead / lead counterpart
- team-specific prompts refer to the selected team first
- the answer is scoped to that team, not the full global roster

### 4a. Output Model Routing

Goal:
- prove that an admin can set a shared default model or detected output-type models for team delivery without editing backend config files

Steps:
1. Return to the root AI Organization workspace.
2. Open `AI Engine Settings`.
3. Confirm the output-model routing panel loads.
4. Review the recommended self-hosted starting points.
5. Switch between:
   - one shared model for everyone
   - detected output-type routing
6. Save a detected routing example such as:
   - general text -> `Qwen3 8B`
   - research and reasoning -> `Llama 3.1 8B`
   - code generation -> `Qwen2.5 Coder 7B`
   - vision analysis -> `LLaVA 7B`
7. Return to the team view and confirm the team summary shows the detected effective model for that lane.

Expected outcome:
- the panel lists installed local models and recommended self-hosted starting points
- the admin can save either routing mode without touching YAML directly
- team-facing summaries reflect the detected effective model after save
- ordinary non-admin interaction is not framed as the place to rewrite shared output-model policy

### 4b. Output Block And Media Readiness

Goal:
- prove that retained outputs are written to the correct storage posture and that media generation is either live or clearly blocked

Steps:
1. Confirm the current output-block mode is visible in the setup or runtime notes.
2. If the run is `local_hosted`, verify the host path exists before you start the browser walkthrough.
3. If the run is `cluster_generated`, verify the storage path is presented as cluster-managed/PVC-backed.
4. Run the health check and confirm whether the media engine is online.
5. If media is online, run the headed browser proof that renders generated media artifacts and exposes save/download paths.
6. If media is offline, capture the blocker text and stop the run as blocked instead of treating it as a pass.
7. Use a team-managed flow such as guided team creation and temporary-group review to confirm retained outputs are still reviewable after archive.

Expected outcome:
- local-hosted output storage is tied to a real host directory and reads as a self-hosted exception
- cluster-generated output storage reads as the default managed path
- media readiness is explicit, not implied
- team-managed outputs stay reviewable after temporary group closure

### 5. Governed Mutation: Cancel Path

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

### 6. Governed Mutation: Execute Path

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

### 7. Deployment Context Intake

Goal:
- prove that user-provided documentation can be loaded into governed long-term context instead of being treated as ordinary chat

Steps:
1. Open `Resources -> Deployment Context`.
2. Paste a short deployment brief, policy note, or requirements summary.
3. Load it as either `user_private_context` for private records tied to a target goal set, `customer_context` for operator/customer deployment material, or `reflection_synthesis` for a distilled lesson/pattern/trajectory-shift note.
4. Set a reasonable trust/sensitivity posture.

Suggested sample content:
- a short note describing target users, deployment environment, and expected output style

Expected outcome:
- the entry is accepted as governed context
- it appears in recent Deployment Context history
- the UI makes clear this is governed context, not ordinary Soma memory
- it is also clear that loading governed context does not silently create team-shared `AGENT_MEMORY`
- private context shows private/restricted posture and target-goal metadata when provided
- reflection/synthesis context shows a synthesis-oriented source kind such as `lesson`, `inferred_pattern`, `contradiction`, `trajectory_shift`, or `meta_observation`
- it is clear that this context can support Soma and team leads without becoming ungoverned shared chat history

### 8. MCP Visibility And Tool Activity

Goal:
- prove that the operator can understand what tools are available and what agents have actually used

Steps:
1. Open `Resources -> Connected Tools`.
2. Confirm installed/connected MCP servers are visible.
3. Review recent MCP activity.

Expected outcome:
- the page shows connected servers clearly
- recent activity includes persisted MCP usage, not only ephemeral live stream events
- the operator can tell which server/tool was used

If the environment allows low-risk curated install testing:
- install only a known local-first, policy-compliant library entry
- do not treat a remote/high-risk MCP install rejection as a failure; it is a valid governance result

### 9. Optional Web/External Research

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

### 10. Audit / Activity Review

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

### 11. Failure Recovery Check

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
3. Soma can be used from the root workspace to shape or create a team.
4. Selecting a team shifts the interaction into a clearly team-focused lead workspace.
5. Admins can configure organization output-model routing from the root Soma workspace.
6. Informational prompts return `answer`.
7. Mutating prompts enter governed `proposal` flow.
8. Cancel and approve paths both behave correctly.
9. Deployment context can be loaded with visible governance posture.
10. Connected Tools makes MCP availability and recent usage understandable.
11. Optional web research, if enabled, behaves as a governed capability.
12. Audit/activity surfaces reconstruct the session clearly.
13. Failure cases degrade safely into blocker/error states instead of UI collapse.
14. The same walkthrough works from Windows against a self-hosted runtime with an explicit non-loopback AI host.

## Initial Release Handoff

Use this shorter sequence when you are validating a fresh checkout on another machine before a first release handoff:

1. Clone or update the repo on the second machine.
2. Follow [Local Development Workflow](./LOCAL_DEV_WORKFLOW.md) for the host you are using.
3. Start the supported runtime (`uv run inv compose.up --build` on WSL/Linux/macOS, or the supported self-hosted runtime path with an explicit non-loopback AI endpoint on Windows or another host).
4. Run `uv run inv ci.release-preflight --runtime-posture --service-health --live-backend`.
5. Run the remote walkthrough in this document from the second machine.
6. Confirm the current release blockers are named in `V8_DEV_STATE.md` before you declare the release ready.

Initial-release gate rule:
- treat `V8_DEV_STATE.md` as the live blocker board
- do not promote a release if the remote walkthrough finds a fresh Soma chat, proposal/confirm, guided team creation, context intake, or artifact/output review defect
- record media-engine gaps as environment notes unless the target machine is explicitly configured for live media generation

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
- [Development Workflow](./LOCAL_DEV_WORKFLOW.md)
