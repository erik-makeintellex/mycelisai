# MCP Service Configuration (Local-First) V7

Version: `1.0`
Status: `Authoritative`
Last Updated: `2026-02-28`
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
5. Remote Service Exception Workflow
6. Library Authoring Standard
7. Runtime and Health Requirements
8. Governance and Risk Controls
9. Testing and Verification Matrix
10. Parallel Delivery Plan
11. Acceptance Gate

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

### Step 3 - Installation path validation
- Install through curated library API/UI path.
- Verify server registers in `mcp_servers`.
- Verify tools cache in `mcp_tools`.

### Step 4 - Runtime validation
- Call at least one tool via MCP executor path.
- Verify event emission (`tool.invoked`, `tool.completed|tool.failed`).
- Verify run timeline linkage.

### Step 5 - Recovery validation
- Simulate service unavailable state.
- Verify degraded messaging and retry/recovery actions.

---

## 5. Remote Service Exception Workflow

Use only when local execution is not feasible.

Required controls:
1. Explicit risk classification (`high` by default).
2. Governance confirmation on install/enable.
3. Host allowlist and protocol restrictions.
4. Clear user-visible remote boundary indicator.
5. Degraded fallback guidance if remote endpoint fails.

---

## 6. Library Authoring Standard

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

## 7. Runtime and Health Requirements

Each installed service must expose status through existing MCP surfaces:
- install state
- connection state (`connected|error|offline`)
- last error message
- tool count discovery

Global health integration:
- MCP service health contributes to system degraded state.
- Status drawer and degraded banner must provide actionable recovery for MCP failures.

---

## 8. Governance and Risk Controls

Risk classes:
- `low`: read/list style tools
- `medium`: local writes, local state mutation
- `high`: remote services, external data egress, install-time privilege surfaces

Policy requirements:
- library install remains curated path
- raw arbitrary install remains gated by security phase policy
- mutations still follow proposal + confirm where applicable

---

## 9. Testing and Verification Matrix

### Unit
- library parse and lookup tests
- `ToServerConfig` mapping tests
- local/remote validation helper tests

### Service Integration
- install/list/get/delete status tests
- tool caching and lookup tests
- duplicate name and not-found behavior tests

### Handler/API
- library install endpoint success/failure tests
- MCP list/tools endpoint response-shape tests
- remote exception path validation tests

### UI/E2E
- install local service from curated library
- call tool and verify run/timeline events
- simulate failure and verify degraded recovery actions

Recommended commands:

```bash
cd core && go test ./internal/mcp/ -count=1
cd core && go test ./internal/server/ -run TestHandleMCP -count=1
cd interface && npm run build
```

---

## 10. Parallel Delivery Plan

- Team Forge: service contract and library onboarding flow
- Team Circuit: degraded/recovery UX integration for MCP failures
- Team Helios: shared API normalization for MCP install/list/call paths
- Team Sentinel: regression suite and release evidence

---

## 11. Acceptance Gate

MCP service onboarding is production-ready only when:
1. service is documented in curated library with local-first defaults
2. install path works through curated API/UI flow
3. tool invocation emits traceable mission events
4. degraded/failure states have actionable recovery
5. test matrix evidence is attached in release artifacts

---

End of document.
