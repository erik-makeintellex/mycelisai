# System Status & Recovery
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md) | [User Docs Home](README.md)

Use product-level recovery first. Open infrastructure detail only when the visible Outcome or Soma alert cannot explain the safe next action.

For an interactive generated package, a saved folder or visible `Output ready` card is not final validation. If the approved control does not visibly change the marked app surface, Soma must keep the work in review or recovery and ask the producing team to repair it; structural readback or runtime fallback cannot override that browser result.

## User-Facing Recovery

When work degrades, Mycelis should explain:

- what failed or became uncertain
- what remains trusted
- which deliverable or proof is still available
- whether retry is safe
- what needs operator attention

Start from the alert, Outcome review, or Soma thread. Use the offered **Retry**, **Recover**, **Review output**, or **Open details** action rather than recreating the request from memory.

### When an external result is uncertain

If Mycelis handed work to an external system but did not receive a final result, it must not retry the mutation blindly. Open the affected work item and record what you observed in the external system:

- **Completed there**: the external change exists. Mycelis closes the work as operator-attested and does not replay it.
- **Did not complete**: the change does not exist. Return to Soma to create and approve a new attempt.
- **Still cannot tell**: keep the work degraded until better evidence is available.

Add a short observation and, when available, an evidence reference such as a receipt, external record ID, or trusted URL. This verification is retained with the work history. Submitting it never starts another run automatically.

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

Normal development uses the Docker `pgvector/pgvector:pg16` PostgreSQL server and Dockerized NATS while Core and Interface run locally. Relational and vector data share the PostgreSQL service's `postgres-data` volume. Local Core connects to the configured published port (`127.0.0.1:15432` by default); Compose Core connects to `postgres:5432`.

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

`lifecycle.down` leaves Dockerized PostgreSQL and NATS running. Use `uv run inv lifecycle.down --include-data-plane` to stop the local app and data plane together without deleting volumes, or `uv run inv compose.down` when only the dependency containers need to stop. Ollama and host runtimes remain independently managed.

A host `psql` binary may be used as a client against the published Docker port. A native host PostgreSQL server and `MYCELIS_DEV_INFRA_MODE=native` are unsupported recovery paths.

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
