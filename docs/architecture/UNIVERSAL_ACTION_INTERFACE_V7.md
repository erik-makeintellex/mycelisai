# Universal Action Interface V7
> Navigation: [Project README](../../README.md) | [Backend](BACKEND.md) | [V8 Runtime Contracts](../architecture-library/V8_RUNTIME_CONTRACTS.md) | [V8 UI/API Contract](../architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md)

Status: historical V7 migration input.

The active action/capability model is governed through V8 runtime contracts, Connected Tools/MCP behavior, proposal/approval flow, and capability policy. This V7 file remains as rationale for why actions need a universal, auditable shape.

## Retained V7 Signal

V7 established that Mycelis actions should be:
- described before execution
- invoked through typed requests
- bounded by policy
- adaptable across MCP, OpenAPI, Python/internal tools, and future service interfaces
- visible to operators as governed capabilities, not raw implementation hooks
- tested across unit, integration, API, and UI layers

## Current Translation

| V7 concept | Current owner |
| --- | --- |
| Action plane | capability/tool execution and MCP integration |
| Universal invoke request | typed tool/action request payloads |
| Universal invoke response | bounded execution result / blocker payloads |
| Service registry | Resources / Connected Tools and backend registry |
| Template marketplace | bootstrap/template and resource registry docs |
| Python manager | Python task automation only, not product runtime action logic |

## Active Rules

- Product actions that mutate state must use governance/proposal pathways.
- Tool execution must emit audit-ready evidence.
- UI must show normalized result or blocker states.
- Raw secrets must never appear in action definitions, logs, docs, or UI.
- App-tied management behavior belongs in Go runtime or Python task modules as defined by repo language ownership.

## Promotion Rule

When a V7 action idea becomes active work, update the owning V8 runtime/UI docs, backend contracts, tests, and `.state/V8_DEV_STATE.md`. Do not grow this historical file.

Managed interface build command: `uv run inv interface.build`.
