# User Acceptance Runbook
> Navigation: [Project README](../README.md) | [Docs Home](README.md) | [Testing](TESTING.md) | [Canonical PRD](architecture-library/MYCELIS_CANONICAL_PRD.md)

Use this runbook for human validation of the delivered Mycelis experience. It applies to same-machine development review, another browser or device, and release deployment certification.

## Purpose

Acceptance proves the Trusted Outcome Journey through the address a user actually opens:

```text
Ask
-> Understand
-> Approve
-> Execute
-> Deliver
-> Trust
-> Recover
-> Revisit
```

The Soma-first operator workflow does not pass merely because endpoints respond or buttons can be clicked. A new user must understand what to do, see useful progress, open the result, and know what remains trusted later.

## Record Before Testing

Record:

- branch and commit SHA
- local date/time
- runtime lane
- UI address used by the tester
- Core/API address when separately exposed
- AI endpoint host posture without tokens
- browser, operating system, and viewport/device
- whether test-owned retained fixtures were cleaned first

## Runtime Lanes

| Lane | Purpose |
| --- | --- |
| Source development | Dockerized PostgreSQL/NATS with local Core/Interface for rapid product review |
| Full Compose | Packaged single-host release proof |
| Kubernetes | Helm, ingress, storage, secret, and clustered-runtime proof |
| Remote operator | Real hostname/IP/ingress from another browser or device |

WSL is used only when it provides distinct release evidence. It is not the normal source-development application host.

## Preflight

For normal development:

```powershell
git status --short --branch
uv run inv compose.infra-health
uv run inv lifecycle.status
uv run inv lifecycle.health
```

For full Compose:

```powershell
uv run inv compose.status
uv run inv compose.health
```

For Kubernetes, run the selected Helm standards check plus cluster status/readiness and operator-address probes.

Before a fresh-user acceptance pass, use only owner-tagged QA cleanup. Do not infer test ownership from names or delete shared NATS/database state.

## Interaction Review Standard

At every step inspect:

- first-glance purpose and next action
- plain-language copy
- layout density and visual hierarchy
- primary scroll ownership
- text-field reachability and growth
- panel, drawer, and keyboard overlap
- desktop, tablet, and compact-phone behavior
- loading, empty, blocked, degraded, retry, and completed states
- console, hydration, network, and page errors

Do not accept a technically functional screen that is confusing, crowded, clipped, or dependent on runtime vocabulary.

Human-first acceptance is performed before Playwright scrolls, focuses, fills, or clicks for the user:

- the Soma composer is visible in the untouched first viewport
- one Workspace/Outcome context and no more than one background-work entry point compete with the thread
- no drawer opens automatically or reduces the conversation width
- ordinary user surfaces do not expose agent rosters, NATS, MCP, run ids, execution contracts, event chains, or message-bus language
- delayed work leaves the composer enabled; a steering message is acknowledged and remains correlated to the same Outcome
- completion contains one concise summary and one primary `Open output` action
- substantial output opens into a dedicated readable/operable surface, and Back restores the originating Soma conversation and Outcome
- a novice can locate Ask, active work, and a completed deliverable without instruction

## Trusted Outcome Walkthrough

### 1. Sign In And Enter Soma

Sign in through the configured local or SSO path.

Expected:

- successful authentication lands at `/dashboard`
- Soma is the dominant operating surface
- starter asks are compact and optional
- Outcomes and review panels do not squeeze the thread by default
- the composer is immediately reachable

### 2. Ask For A Direct Answer

Ask a simple non-mutating question and then request a compact list or table.

Expected:

- Soma answers without forcing a proposal
- response depth matches the request
- sources or trust context appear when needed
- the answer does not expose raw provider or transport noise

### 3. Shape Meaningful Work

Ask Soma to create or review a durable result. Include the expected output type and how it should be usable.

Expected:

- Soma can discuss and refine the request before execution
- the proposed team/capability shape is the smallest useful one
- the proposal summarizes intended work and deliverables in conversational form
- risk and approval posture are understandable

### 4. Cancel And Adjust Conversationally

Type `cancel` for one governed proposal, then describe the adjustment in the normal composer and ask again.

Expected:

- cancellation applies no mutation
- the conversation remains intact
- adjustment does not require rebuilding the request from scratch

### 5. Approve And Continue Talking

Type `approve` or `go ahead` for the revised work.

Expected:

- visible queued/running feedback appears immediately
- one clear **Approve** or **Start** button is present when execution is pending; no separate adjust, team-routing, or run-navigation button is required
- Soma remains available while NATS-backed work continues
- progress is correlated to the same Outcome
- the UI does not navigate the user into a raw event chain as the primary experience

### 6. Receive And Open The Deliverable

Wait for terminal feedback without manually polling hidden runtime pages.

Expected:

- Soma reports completion or truthful degradation
- the completion message briefly explains what was delivered
- one primary `Open output` action opens the file, app, document, media item, dataset, report, or package in a dedicated surface
- folder access, Resources, proof, versions, download, and technical references remain secondary
- group-owned content is isolated under `groups/<team-id>/generated/...`
- package entrypoint, local dependencies, and expected interaction were validated

### 7. Review Trust

Open proof or receipt details deliberately.

Expected:

- Outcome health uses the canonical vocabulary
- proof explains why the result is trusted
- runtime IDs and raw events stay behind Details or Inspect
- missing output or validation creates recovery, not false completion

### 8. Exercise Recovery

Use a safe injected or existing degraded case.

Expected:

- the UI states what failed, what remains trusted, and what can continue
- recovery retains the approved contract and context
- uncertain external mutation is not silently retried
- the user returns to Soma or the owning Outcome after recovery

### 9. Revisit Across Surfaces

Reload and revisit the Outcome from Soma, Groups, and Resources.

Expected:

- the deliverable and producing group remain attributable
- completed history does not inflate active-review counts
- replying to an output can create an update, alternative, or follow-on request
- selected files or context can return to Soma without becoming unrestricted permanent memory

### 10. Review Supporting Routes

Check Groups, Resources, Docs, and guided Settings. Turn on Admin tools only for Activity, Runs, deep Memory, System, or raw capability inspection.

Expected:

- each route has one obvious purpose
- tabs/list-detail/overlays replace long compressed columns
- Help content matches visible labels and workflows
- infrastructure depth remains optional

## Workflow Shape And Restart Acceptance

Use one release-readiness objective to prove that Soma chooses the smallest useful
execution shape and that durable work survives restart:

```text
Prepare a self-hosted release-readiness package for a Windows operator lane that
uses a Windows GPU host for AI. I need a quick recommendation, a deployable
validation plan, and a reviewable package I can resume after a reboot.
```

Exercise these four variants:

1. Ask for the shortest practical recommendation. Soma should answer directly,
   create no team or workflow group, recommend the supported Docker Compose lane
   first with an explicit Windows AI endpoint, and keep Kubernetes as the modular
   scale-up proof lane.
2. Ask for the smallest useful team with a named lead, validation checklist,
   deployment recommendation, risk review, and retained release-readiness package.
3. Ask Soma to split broader work into compact planning, deployment-validation,
   and review lanes. Each lane must have a visible purpose, owner, output contract,
   and retained output.
4. After variant 2 or 3 retains an output, restart the environment, reopen the same
   Workspace, and ask: `Resume the release-readiness work from the retained package
   and show me what is already done, what remains, and which lane or lead owns the
   next step.` The same package and producing team/group must remain reviewable;
   Soma must resume from retained evidence rather than inventing a fresh answer.

The automated counterparts are:

- `interface/e2e/specs/workflow-output.direct.spec.ts`
- `interface/e2e/specs/workflow-output.compact-team.spec.ts`
- `interface/e2e/specs/workflow-output.multi-lane.spec.ts`
- `interface/e2e/specs/workflow-output.reload-review.spec.ts`

Capture the prompt, execution shape, visible owner/output contract, retained output,
and before/after-restart evidence. A pass requires the direct answer to resist
unnecessary orchestration, the compact and multi-lane variants to remain legible,
and restart review to preserve attribution and the next owner.

## Cross-Device Matrix

Repeat the primary journey at minimum on:

- compact phone
- tablet or narrow laptop
- standard laptop/desktop
- wide desktop

Verify navigation, peer tabs, overlays, composer, approval, deliverables, recovery, Docs, and Settings with pointer and keyboard equivalents where applicable.

## Release Certification

Run the committed release candidate through:

```powershell
uv run inv ci.release-preflight --lane=release
```

Use WSL only when it supplies distinct deployment-mimic evidence. Keep the Windows root repo as the dev/staging worktree and use a clean WSL deployment-mimic checkout refreshed from git as the validation host. The guarded tasks are:

```powershell
uv run inv wsl.status
uv run inv wsl.refresh --branch <name>
uv run inv wsl.validate --lane=release
```

For a same-machine WSL or Compose proof, verify `http://localhost:3000` from the Windows side with both a simple HTTP probe and a real browser launch. If the first request warms a cold runtime, classify it as `cold_start_first_request` instead of a clean first-pass success. Do not silently relabel the run as a clean first-pass success. Record whether the issue is a `cold_start_first_request`, a steady-state regression, or an environment/setup gap.

Deployment-context loading, capability/MCP visibility, and recent persisted tool activity are tested only when relevant to the approved Outcome. They remain supporting evidence, not setup steps every new user must complete.

## Pass Criteria

The run passes only when:

- the real delivered UI address was used
- the full Trusted Outcome Journey was understandable
- asynchronous execution did not block continued Soma conversation
- deliverables opened and matched the approved output contract
- proof and recovery were truthful
- refresh/revisit preserved the Outcome
- no blocking console, hydration, page, overlap, overflow, or unreachable-control errors occurred

## Evidence Packet

Retain:

- runtime lane, commit, and UI address
- viewport/device list
- screenshots or short recordings of the major journey steps
- console/page-error summary
- commands and suites run
- pass/fail notes
- blockers, trusted remainder, and recovery action

Do not retain credentials, unrelated user data, or unowned test fixtures.
