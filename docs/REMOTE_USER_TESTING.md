# Remote User Testing Runbook
> Navigation: [Project README](../README.md) | [Docs Home](README.md) | [Testing](TESTING.md)

Use this runbook when a human tester validates Mycelis from a different browser, host, or operator machine than the runtime/control checkout.

## TOC

- [Purpose](#purpose)
- [Current Truth And Boundaries](#current-truth-and-boundaries)
- [Preflight](#preflight)
- [Environment Setup](#environment-setup)
- [Windows Self-Hosted Operator Lane](#windows-self-hosted-operator-lane)
- [Walkthrough](#walkthrough)
- [Pass Criteria](#pass-criteria)
- [Initial Release Handoff](#initial-release-handoff)
- [Failure Notes To Capture](#failure-notes-to-capture)
- [Recommended Evidence Capture](#recommended-evidence-capture)

## Purpose

Remote user testing proves the delivered operator path, not only local code health. The tester should reach the UI through the same address a real user will use, interact through Soma, and verify governance, retained outputs, and recovery from the browser.

Use [Testing](TESTING.md) for engineering gates and [Mycelis Canonical PRD](architecture-library/MYCELIS_CANONICAL_PRD.md) for the full browser matrix.

## Current Truth And Boundaries

Accepted runtime lanes:
- Docker Compose on the same Windows machine as the tester
- Docker-in-WSL Compose with Windows browser on the same machine
- Compose or Kubernetes on another host reached by IP/hostname
- self-hosted Kubernetes through real ingress

Required topology truth:
- AI endpoint is explicit and reachable from the runtime
- the browser address is the delivered operator address
- no proof depends on hidden localhost shortcuts unless the delivered lane is same-machine `http://localhost:3000`

Soma-first operator workflow is the target user path.

Remote proof should include deployment-context loading into governed vector-backed stores when a workflow depends on retained host/deployment context. Include MCP visibility and recent persisted tool activity. The safe current actuation proof is governed file output, governed context loading, MCP-backed tool usage, and reviewable audit/activity behavior.

Use a clean WSL deployment-mimic checkout refreshed from git as the validation host, with the Windows root repo as the dev/staging worktree. For WSL proof, use a clean WSL deployment-mimic checkout refreshed from git as the validation host and run `uv run inv wsl.refresh`, `uv run inv wsl.validate`, and `uv run inv ci.release-preflight --lane=release`.

For Windows browser proof, verify `http://localhost:3000` from the Windows side with both a simple HTTP probe and a real browser launch. If the first browser request warms a cold Next.js/Compose path, classify it as `cold_start_first_request` instead of a clean first-pass success; do not silently relabel the run as a clean first-pass success. Record whether the issue is a `cold_start_first_request`, a steady-state regression, or an environment/setup gap.

## Preflight

Record:
- branch and commit SHA
- local date/time
- runtime lane: `compose`, `wsl-compose`, `kubernetes`, or `remote-host`
- UI URL used by the tester
- Core/API URL if separately exposed
- AI endpoint host/IP, not secret tokens
- browser and OS

Run one matching readiness set before inviting the tester:

```bash
uv run inv compose.status
uv run inv compose.health
```

or:

```bash
uv run inv lifecycle.status
uv run inv lifecycle.health
```

or the target-cluster equivalent plus:

```bash
uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml
```

## Environment Setup

For Compose:
- configure `.env.compose`
- set `MYCELIS_COMPOSE_OLLAMA_HOST` to a runtime-reachable endpoint
- run `uv run inv compose.up --build --wait-timeout=240`
- confirm `uv run inv compose.health` passes

For Kubernetes:
- set `MYCELIS_K8S_TEXT_ENDPOINT`
- optionally set `MYCELIS_K8S_MEDIA_ENDPOINT`
- select a values preset with `MYCELIS_K8S_VALUES_FILE`
- deploy through `uv run inv k8s.up` or the target-cluster Helm process

For source-mode proof:
- use the WSL proof checkout for release-style validation
- use the Windows browser for same-machine WSL-hosted UI proof

## Windows Self-Hosted Operator Lane

This lane proves the product as a Windows operator actually uses it:
- browser runs on Windows
- UI is reached from Windows at the delivered address
- runtime is Windows/Rancher Desktop Docker-compatible Compose, Docker-in-WSL Compose, Rancher Desktop K3s, or self-hosted Kubernetes
- AI engine is an explicit Windows host/IP or equivalent self-hosted service

Same-machine proof starts at:

```text
http://localhost:3000
```

Second-machine proof uses the reachable hostname or IP. Do not record raw secrets in evidence.

## Walkthrough

### 1. Workspace Entry And Continuity

Open the UI. Confirm the tester can create or re-enter an AI Organization and land in a Soma-primary workspace.

Expected: the product reads as an AI Organization workspace, not a raw chat box or dev console.

### 2. Direct Soma Answer

Ask a non-mutating question.

Expected: Soma returns an `answer` without asking for mutation approval.

### 3. Soma Creates Or Refines A Team

Ask Soma for a focused team or temporary workflow lane.

Expected: the team/lane shape is compact, reviewable, and routed through governed UI state.

### 3a. Groups Workspace

Open the groups/temporary workflow surface.

Expected: retained outputs, status, and review affordances remain visible.

### 4. Team Creation And Team Lead Focus

Run the guided team-creation path.

Expected: the Team Lead, purpose, and compact-default behavior are clear.

### 4a. Output Model Routing

Inspect AI Engine/response-style behavior only if the slice changes provider routing or response contracts.

Expected: effective settings are visible without exposing secrets.

### 4b. Output Block And Media Readiness

Create or review an output artifact.

Expected: generated files are linked/downloadable from the configured output block. If media is unavailable, the UI shows a clear blocker.

### 5. Governed Mutation: Cancel Path

Ask for a protected/mutating action and cancel it.

Expected: proposal state is visible, cancellation is clean, and no mutation is applied.

### 6. Governed Mutation: Execute Path

Repeat a protected action and approve it.

Expected: execution state and result are visible, auditable, and retained.

### 7. Deployment Context Intake

Provide deployment context only when needed for the tested slice.

Expected: private context is treated as governed input, not leaked as raw backend noise.

### 8. MCP Visibility And Tool Activity

Open Resources/Capabilities when MCP behavior is in scope.

Expected: registry/library/activity state is understandable and actions remain governed.

### 9. Optional Web/External Research

Run only when search/research is in scope.

Expected: configured provider posture is visible, and unreachable endpoints block clearly.

### 10. Audit / Activity Review

Open activity or run timeline.

Expected: direct answers, proposals, executions, and retained artifacts are reviewable.

### 11. Failure Recovery Check

Temporarily break or simulate AI endpoint failure only if safe for the environment.

Expected: blocker appears, recovery restores the same workflow after endpoint health returns.

## Pass Criteria

The run passes only when:
- the tester used the real delivered UI address
- Soma direct answer, governed proposal, approval/cancel, and retention paths worked
- output artifacts were visible or a truthful blocker was recorded
- the AI endpoint was explicit and reachable from runtime
- failures surfaced in user-readable form

## Initial Release Handoff

Provide:
- UI URL
- runtime lane
- AI endpoint host posture
- evidence commands run
- screenshots or short recordings
- pass/fail notes
- known blockers and recovery steps

Follow [Local Development Workflow](LOCAL_DEV_WORKFLOW.md) for host-specific bring-up.

## Failure Notes To Capture

Capture:
- exact user action
- visible UI state
- browser console/network symptom when available
- Core/log symptom when available
- runtime lane and UI URL
- whether refresh/retry recovered

## Recommended Evidence Capture

Minimum:
- dashboard or organization entry screenshot
- Soma direct answer screenshot
- proposal/cancel screenshot
- proposal/execute result screenshot
- retained output or blocker screenshot
- activity/run timeline screenshot

References:
- [Testing](TESTING.md)
- [Local Development Workflow](LOCAL_DEV_WORKFLOW.md)
- [Mycelis Canonical PRD](architecture-library/MYCELIS_CANONICAL_PRD.md)
