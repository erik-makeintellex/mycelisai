# MCP Service Configuration (Local-First) V7

Version: `1.1`
Status: `Authoritative`
Last Updated: `2026-03-01`
Scope: How to add and operate MCP services with local serving as the default posture

This document defines the standard onboarding and configuration process for MCP services.
Default stance is local-first. Remote MCP is opt-in and treated as higher risk.
Companion extension spec: `docs/architecture/UNIVERSAL_ACTION_INTERFACE_V7.md` (MCP as one adapter within universal action plane).

---

## Table of Contents

1. Purpose
2. Local-First Policy
3. Canonical MCP Service Spec
4. Local Service Onboarding Workflow
5. Policy Alignment Inspection Gate
6. Remote Service Exception Workflow
7. Library Authoring Standard
8. Runtime and Health Requirements
9. Governance and Risk Controls
10. Testing and Verification Matrix
11. Parallel Delivery Plan
12. Acceptance Gate

---

## 1. Purpose

When adding a new MCP service, operators and developers must have one repeatable process that:
- defaults to local execution
- is safe by default
- is testable end-to-end
- is observable in runs, events, and diagnostics

---

## 2. Local-First Policy

1. Default transport is `stdio` with a locally executed command.
2. If `sse` transport is used by default, URL must be loopback (`127.0.0.1` or `localhost`).
3. Remote hosts are not defaultable in baseline profiles.
4. Remote MCP service onboarding requires explicit governance confirmation.
5. All MCP service additions must define risk level and recovery behavior.

---

## 3. Canonical MCP Service Spec

Every MCP service definition must include these fields:

```yaml
name: "service-name"
description: "What it does"
transport: "stdio" # default
command: "npx"
args: ["-y", "@scope/server-package"]
env:
  OPTIONAL_TOKEN: ""
tags: ["category", "local"]
tool_set: "optional-toolset"
```

Rules:
- `transport` allowed: `stdio|sse`
- `command` and `args` required for `stdio`
- `url` required for `sse`
- `env` keys documented with safe defaults
- `name` must be unique across library entries

---

## 4. Local Service Onboarding Workflow

### Step 1 - Library entry (authoring)
- Add service to `core/config/mcp-library.yaml`.
- Mark tags to indicate local-first usage where applicable.
- Provide empty-string placeholders for required secrets.

### Step 2 - Config validation
- Validate YAML parse and service lookup behavior.
- Validate generated `ServerConfig` mapping.

### Step 3 - Policy alignment inspection (mandatory)
- Evaluate service against policy stack:
  - Soma targeted safety rules (global baseline)
  - deployment defaults (org/environment policy)
  - user-added policy overlays (tenant/user overrides)
- Produce an inspection report before install:
  - `service_name`, `source`, `risk_level`, `required_scopes`, `network_locality`, `secrets_declared`
  - `decision` (`allow|require_approval|deny`)
  - `reasons[]` and required remediation if denied
- If inspection result is `deny`, install is blocked.
- If inspection result is `require_approval`, installation remains pending until approved.

### Step 4 - Installation path validation
- Install through curated library API/UI path.
- Verify server registers in `mcp_servers`.
- Verify tools cache in `mcp_tools`.

### Step 5 - Runtime validation
- Call at least one tool via MCP executor path.
- Verify event emission (`tool.invoked`, `tool.completed|tool.failed`).
- Verify run timeline linkage.

### Step 6 - Recovery validation
- Simulate service unavailable state.
- Verify degraded messaging and retry/recovery actions.

---

## 5. Policy Alignment Inspection Gate

Every library/service intake must resolve this precedence order:
1. `Soma baseline policy` (non-overridable deny rules)
2. `Deployment defaults` (org defaults)
3. `User overlays` (tenant/user additions)

Resolution rules:
- overlay can tighten restrictions, but cannot bypass non-overridable baseline denies
- final resolved policy is attached to install event and audit trail
- effective policy hash must be persisted for reproducible reviews

Minimum checks:
1. capability scope classification (`read|write|execute|network|credential`)
2. locality classification (`local|remote`)
3. sandbox profile compatibility
4. secret handling contract (declared vars only, no hidden pulls)
5. rollback compatibility (disable/uninstall path proven)

---

## 6. Remote Service Exception Workflow

Use only when local execution is not feasible.

Required controls:
1. Explicit risk classification (`high` by default).
2. Governance confirmation on install/enable.
3. Host allowlist and protocol restrictions.
4. Clear user-visible remote boundary indicator.
5. Degraded fallback guidance if remote endpoint fails.

---

## 7. Library Authoring Standard

`core/config/mcp-library.yaml` conventions:
- group under the most specific category
- include short operational description
- include tags for discoverability
- prefer local endpoints in env defaults (`127.0.0.1`)
- keep secret defaults blank (`""`)

Example local SSE (allowed default):

```yaml
name: "local-vision"
description: "Local vision inference bridge"
transport: sse
url: "http://127.0.0.1:9101/sse"
tags: ["vision", "local"]
```

Example remote SSE (exception path):

```yaml
name: "remote-knowledge"
description: "Hosted knowledge MCP bridge"
transport: sse
url: "https://mcp.example.com/sse"
tags: ["knowledge", "remote"]
```

---

## 8. Runtime and Health Requirements

Each installed service must expose status through existing MCP surfaces:
- install state
- connection state (`connected|error|offline`)
- last error message
- tool count discovery

Global health integration:
- MCP service health contributes to system degraded state.
- Status drawer and degraded banner must provide actionable recovery for MCP failures.

---

## 9. Governance and Risk Controls

Risk classes:
- `low`: read/list style tools
- `medium`: local writes, local state mutation
- `high`: remote services, external data egress, install-time privilege surfaces

Policy requirements:
- library install remains curated path
- raw arbitrary install remains gated by security phase policy
- mutations still follow proposal + confirm where applicable
- policy evaluation must include effective defaults + user overlays before enablement

---

## 10. Testing and Verification Matrix

### Unit
- library parse and lookup tests
- `ToServerConfig` mapping tests
- local/remote validation helper tests
- policy resolution tests (baseline + defaults + user overlay precedence)
- deny/approval/allow classification tests for intake

### Service Integration
- install/list/get/delete status tests
- tool caching and lookup tests
- duplicate name and not-found behavior tests
- install blocked when policy inspection returns `deny`
- install pending when policy inspection returns `require_approval`

### Handler/API
- library install endpoint success/failure tests
- MCP list/tools endpoint response-shape tests
- remote exception path validation tests

### UI/E2E
- install local service from curated library
- call tool and verify run/timeline events
- simulate failure and verify degraded recovery actions
- inspect report is visible before install confirmation
- user policy overlay changes effective decision in preview

Recommended commands:

```bash
cd core && go test ./internal/mcp/ -count=1
cd core && go test ./internal/server/ -run TestHandleMCP -count=1
uv run inv interface.build
```

---

## 11. Parallel Delivery Plan

- Team Forge: service contract and library onboarding flow
- Team Circuit: degraded/recovery UX integration for MCP failures
- Team Helios: shared API normalization for MCP install/list/call paths
- Team Sentinel: regression suite and release evidence

---

## 12. Acceptance Gate

MCP service onboarding is production-ready only when:
1. service is documented in curated library with local-first defaults
2. intake inspection passes resolved policy checks (Soma baseline + defaults + user overlay)
3. install path works through curated API/UI flow
4. tool invocation emits traceable mission events
5. degraded/failure states have actionable recovery
6. test matrix evidence is attached in release artifacts

---

End of document.
