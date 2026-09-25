# System Status & Recovery
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> How to read system health and recover quickly without leaving your workflow.

---

## Global Health Signals

From any page, use:
- top status strip
- degraded mode banner (when present)
- status drawer

These provide immediate visibility into:
- council reachability
- NATS state
- SSE stream state
- governance state
- active mission count

---

## Status Drawer

Open the drawer from:
- status strip click
- floating status button
- degraded banner action

Use it to answer:
1. Is the issue local (one member/service) or systemic?
2. Which subsystem failed first?
3. Is recovery already in progress?

---

## Degraded Mode Banner

The banner appears when critical dependencies actually degrade (for example SSE/NATS/council connectivity after startup has had a chance to connect).
Normal startup while the live stream is still connecting should not be shown as degraded mode.

Available recovery actions:
- `Retry`
- `Switch to Soma` when direct specialist routing was active
- `Open Status`

The banner auto-clears once health recovers.

---

## "Core API unreachable"

When the interface cannot reach the Mycelis Core API at all (a network error or a 5xx from the proxy, not just one degraded service), the degraded banner, the Soma header, the causal summary, and the Status drawer all say **"Core API unreachable"** together — no surface claims readiness or completed work while Core is down. Groups and Runs show a load-failure state with `Retry` instead of a misleading empty list.

To recover, in a terminal:

```bash
uv run inv lifecycle.up --frontend
uv run inv lifecycle.status
```

Start the runtime with the first command, then confirm with the second. See [Local Development Workflow](../LOCAL_DEV_WORKFLOW.md#troubleshooting) for other Core startup issues.

---

## System Checks (`/system`, Advanced mode)

System Checks lets you run targeted validations:
- NATS connected
- Database reachable
- SSE stream live
- Trigger engine active
- Scheduler state

Each row supports:
- `Run Check`
- `Copy` diagnostics snippet
- last checked timestamp

Use copied snippets in support/debug threads to share precise status context.

---

## Recommended Recovery Sequence

1. Open Status Drawer and identify first failing subsystem.
2. Run relevant System Checks in `/system`.
3. Use `Retry` from banner or affected surface.
4. If the issue is tied to a direct specialist route, switch back to Soma and continue.
5. Confirm banner clears and checks return healthy/degraded as expected.

---

## Full Local Reset (Fresh Deployment)

If runtime state is stale and normal retries do not recover:

```bash
uv run inv lifecycle.down
uv run inv k8s.reset
uv run inv lifecycle.up --build --frontend
uv run inv lifecycle.health
```

For normal startup without deleting the cluster, use the canonical cluster sequence first:

```bash
uv run inv k8s.up
uv run inv lifecycle.up --frontend
uv run inv db.migrate
uv run inv lifecycle.status
uv run inv lifecycle.health
```

Then reload `/dashboard` and re-check:
- degraded banner state
- status drawer service rows
- `/system` quick checks

---

## When To Escalate

Escalate to ops/dev when:
- SSE stays offline after multiple retries
- NATS remains disconnected
- council calls fail with repeated 5xx
- quick-check timestamps update but statuses do not recover

Include:
- copied Quick Check snippet(s)
- failing route
- approximate timestamp
- action attempted
