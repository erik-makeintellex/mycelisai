# Worker Library / Hermes-Compatible Execution Source Map
> Navigation: [Architecture Docs Index](ARCHITECTURE_LIBRARY_INDEX.md) | [Canonical PRD](MYCELIS_CANONICAL_PRD.md)

> Status: Supporting implementation source map
> Last Updated: 2026-06-29
> Purpose: Define the focused worker-library execution package without creating a parallel architecture doctrine.

## Delivery Boundary

The worker library turns approved agentry decisions into executed work. It does not replace Soma, governance, NATS, teams, runs, MCP, or the Outcome model. The product calls one normalized worker interface, and configuration decides whether work executes through the default central backend or a Hermes-compatible backend.

Default execution remains `central`: centrally provisioned workers, centrally governed credentials, centrally controlled policy, centralized logs, centralized usage accounting, and operational responsibility owned by Mycelis.

Hermes compatibility is additive. An organization may configure `hermes_api` or a compatible `hermes_like` runtime, but the Mycelis control plane still owns run creation policy, org/project/user permissions, budget gates, audit schema, identity mapping, backend selection, approval routing, observability requirements, and fallback behavior.

## Primary Sources To Track

These are the primary-source surfaces the implementation should track before wiring a production Hermes backend:

| Source | Use | Current Confidence |
| --- | --- | --- |
| [NousResearch `hermes-agent` repository](https://github.com/NousResearch/hermes-agent) | Agent runtime, tools, skills, CLI, provider behavior, and API shape as implemented. | Confirm before Phase 2 endpoint wiring. |
| [Official Hermes Agent docs](https://hermes-agent.nousresearch.com/) | Product/runtime concepts, programmatic integration, provider runtime resolution, tools runtime, skills, API server, CLI/slash commands. | Confirm before Phase 2 endpoint wiring. |
| Hermes API Server docs | Health, capabilities, run submission, run state, event streaming, stop/cancel, approvals, final result retrieval. | Required for real `hermes_api`. |
| Hermes Programmatic Integration docs | Client lifecycle and embedding patterns. | Required for adapter ergonomics. |
| Hermes Provider Runtime Resolution docs | Backend/model/runtime selection behavior. | Required to avoid hardcoding a single protocol. |
| Hermes Tools Runtime docs | Tool execution, permission, and approval boundaries. | Required for security model. |
| Hermes Skills docs | Skill/package semantics and compatibility limits. | Useful for capabilities discovery. |
| Hermes CLI/slash-command docs | User/operator behavior only; do not model the worker API around CLI text. | Secondary for API design. |
| NVIDIA NemoHermes/NemoClaw docs | Hermes-like runtime patterns and compatibility boundaries. | No primary endpoint contract confirmed in this slice; track as optional compatibility research only. |
| Mycelis/OpenAI-compatible gateway references | Fallback mapping when only Responses or Chat Completions style protocol is available. | Use only when durable runs API is absent. |

## Confirmed Mycelis Contract

The internal package is `core/internal/workers`.

Public lifecycle:

```text
createRun -> accepted/run_id -> events/progress -> approval_needed? -> continue/deny -> completed|failed|cancelled -> result/audit
```

Required interface:

```text
CreateRun(request)
StreamRunEvents(runID)
GetRun(runID)
StopRun(runID)
SubmitApproval(runID, approval)
GetCapabilities()
HealthCheck()
```

Normalized objects:

- `WorkerConfig`
- `WorkerBackend`
- `WorkerRunRequest`
- `WorkerRunHandle`
- `WorkerEvent`
- `WorkerApprovalRequest`
- `WorkerApprovalDecision`
- `WorkerResult`
- `WorkerError`
- `WorkerCapabilities`
- `WorkerUsage`
- `WorkerAuditRecord`

Backend names:

- `central`
- `hermes_api`
- `hermes_like`

## Protocol Selection

The adapter must call health/capabilities before execution.

Protocol preference:

1. Hermes-style durable runs API
2. OpenAI-compatible Responses API when it supports stateful runs well enough
3. Chat Completions only for simple synchronous execution with no durable lifecycle, approval gates, or event stream

Phase 1 implements the durable-runs interface and `hermes_api` skeleton only. Phase 2 wires production Hermes endpoints after source confirmation.

## Configuration Shape

Default:

```json
{
  "execution": {
    "backend": "central"
  }
}
```

Hermes-compatible override:

```json
{
  "execution": {
    "backend": "hermes_api",
    "base_url": "https://hermes.example.test",
    "api_key_secret_ref": "secret://org/hermes/api",
    "capabilities_endpoint": "/v1/capabilities",
    "health_endpoint": "/health",
    "preferred_protocol": "runs_api",
    "session_key_strategy": "org_project_run",
    "approval_mode": "mycelis_control_plane",
    "event_stream_mode": "sse",
    "timeout_policy": {
      "connect_ms": 5000,
      "run_ms": 900000,
      "stream_ms": 900000
    },
    "tool_policy": {
      "allow_network": false,
      "allow_files": false,
      "allow_browser": false
    },
    "fallback_backend": "central"
  }
}
```

Selection levels:

- global default: central execution
- org override: org-hosted Hermes/Hermes-like runtime
- project override: project-specific backend
- run override: only when policy permits

## Security Model

- Credentials are referenced by secret IDs.
- The worker library never stores or logs raw API keys, OAuth tokens, payment credentials, card values, browser session secrets, `.env` values, or provider keys.
- Credential passthrough, if ever allowed, must use file references or secure handles and must be explicitly policy-approved.
- Approval requests are normalized and rendered by Mycelis before continuation.
- External runtimes execute only after Mycelis policy allows delegation.
- External backend logs and events are normalized before user/UI consumption.
- Unsupported features produce structured recoverable errors or controlled fallback, never silent degradation.

## Phase Plan

| Phase | Scope | Gate |
| --- | --- | --- |
| Phase 1 | Worker package, central backend, Hermes adapter skeleton, health/capability discovery, run/event/stop/approval normalization, mocked Hermes-compatible tests. | Unit tests with mocked central and Hermes-like backends. |
| Phase 2 | Wire documented Hermes endpoints for health, capabilities, submit, status, stream, stop, approval, result. | Integration proof against a real Hermes-compatible runtime or pinned mock matching official docs. |
| Phase 3 | Backend fallback and policy controls across org/project/run selection. | Policy tests for fallback vs fail-closed and audit records. |

## Acceptance Criteria

- The same agentry call can execute through central workers or Hermes-style workers by changing configuration only.
- The default configuration uses central execution.
- A Hermes-compatible backend can be registered with `base_url` and `api_key_secret_ref`.
- Capabilities are discovered before execution.
- Unsupported Hermes features degrade gracefully.
- Runs can be submitted, streamed/polled, cancelled, approved/denied, completed, failed, and audited through one normalized interface.
- No secrets appear in model-visible context, logs, transcripts, events, or audit text.
