# System Status & Recovery

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

The banner appears when critical dependencies degrade (for example SSE/NATS/council connectivity).

Available recovery actions:
- `Retry`
- `Switch to Soma`
- `Open Status`

The banner auto-clears once health recovers.

---

## System Quick Checks (`/system`)

Quick Checks lets you run targeted validations:
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
2. Run relevant Quick Checks in `/system`.
3. Use `Retry` from banner or affected surface.
4. If council-specific, switch route to Soma and continue.
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
uv run inv k8s.bridge
uv run inv db.migrate
uv run inv lifecycle.up --frontend
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
