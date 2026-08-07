# System Status & Recovery
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [User Docs Home](README.md)

Use product-level recovery first. Open infrastructure detail only when the visible Outcome or Soma alert cannot explain the safe next action.

## User-Facing Recovery

When work degrades, Mycelis should explain:

- what failed or became uncertain
- what remains trusted
- which deliverable or proof is still available
- whether retry is safe
- what needs operator attention

Start from the alert, Outcome review, or Soma thread. Use the offered **Retry**, **Recover**, **Review output**, or **Open details** action rather than recreating the request from memory.

## Global Health Signals

The application can expose:

- a degraded-mode banner for critical dependency failures
- a status drawer opened from the rail or degraded banner
- System checks behind **Admin tools**

Normal connection warm-up should not be presented as a failure. Persistent API, event-stream, NATS, database, provider, or capability failures should be visible and actionable.

## Status Drawer

Use the status drawer to answer:

1. Is the problem limited to one dependency or affecting the whole workspace?
2. Is Core reachable?
3. Is the live event stream connected?
4. Is the required model/provider available?
5. Has recovery already restored the dependency?

The drawer is a quick health view, not the place to manage Outcomes or deliverables.

## Degraded Mode Banner

The banner appears when a critical dependency remains degraded. Available actions can include **Retry** and **Open Status**. It clears after health recovers.

If ordinary conversation still works, keep using Soma while background execution recovers. A degraded subsystem should not freeze unrelated interaction.

## System Checks

Turn on **Admin tools**, open `/system`, and use System checks for targeted diagnostics such as:

- NATS connection
- database reachability
- event-stream state
- trigger/scheduler readiness
- provider or service health

Use copied diagnostics for support without exposing secret values.

The **Deployments** tab reports deployment, execution, workspace, and artifact roots plus version and health posture. Missing values should remain `unknown`; the UI must not guess.

## Output And Data Roots

- `workspace_root` is the configured `MYCELIS_WORKSPACE`. Generated files, group packages, and filesystem capability writes belong here.
- `artifact_root` is `MYCELIS_ARTIFACT_ROOT`. `DATA_DIR` remains a compatibility fallback for artifact/cache storage.

If these roots do not match the intended mounted or host folder, fix configuration before asking Soma to create more retained work.

## Default Development Recovery

Normal development uses PostgreSQL and NATS in Rancher Desktop's Docker engine while Core and Interface run locally.

Check the reusable data plane and app first:

```powershell
uv run inv compose.infra-health
uv run inv lifecycle.status
uv run inv lifecycle.health
```

Restart only the local application processes when needed:

```powershell
uv run inv lifecycle.down
uv run inv lifecycle.up --frontend
uv run inv lifecycle.health
```

`lifecycle.down` leaves Dockerized PostgreSQL and NATS running. Use `uv run inv compose.down` only when intentionally stopping the development data plane.

The host-service fallback is explicit: set `MYCELIS_DEV_INFRA_MODE=native` and use `native-infra.*` only when Dockerized dependencies are intentionally unavailable.

## Full Compose Recovery

Use this lane only when validating the packaged single-host application:

```powershell
uv run inv compose.status
uv run inv compose.health
```

If the committed release candidate needs a rebuild, run:

```powershell
uv run inv compose.up --build --wait-timeout=240
uv run inv compose.health
```

Do not rebuild the full application stack during ordinary source iteration.

## Kubernetes Recovery

For Rancher Desktop K3s or another supported cluster, recover the Mycelis workload and bridge with the narrow Kubernetes tasks:

```powershell
uv run inv k8s.up
uv run inv k8s.status
uv run inv k8s.wait
uv run inv k8s.bridge
uv run inv lifecycle.health
```

Repository tasks do not repair Rancher Desktop, Docker Desktop, WSL, or the host VM. Repair the host runtime through its platform controls, then rerun the narrow Mycelis readiness task.

## Escalation Information

Capture:

- exact user action
- visible alert or Outcome health
- affected deliverable or run link
- browser console/network symptom when available
- Core or service symptom when available
- approximate timestamp
- whether retry or refresh recovered

Never include raw credentials, tokens, or secret environment values.
