# Soma Symbiote Growth and Host Actuation Architecture V7
> Navigation: [Project README](../../README.md) | [Architecture Index](../architecture-library/ARCHITECTURE_LIBRARY_INDEX.md) | [V8 Runtime Contracts](../architecture-library/V8_RUNTIME_CONTRACTS.md)

Status: historical V7 migration input.

This document no longer owns active Soma growth or host-actuation behavior. Current authority lives in:
- [V8 Runtime Contracts](../architecture-library/V8_RUNTIME_CONTRACTS.md)
- [V8.1 Living Organization Architecture](../architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md)
- [V8.2 Production Architecture Target](../../architecture/v8-2.md)
- [V8 Config and Bootstrap Model](../architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md)

## Retained V7 Signal

V7 established several ideas that remain useful as migration constraints:
- Soma is not a disposable chat persona; it is the operator-facing coordinating counterpart.
- Learning must move through reviewable memory and continuity paths.
- Host actuation must be local-first, governed, and reversible where possible.
- Model/provider health must be visible before runtime proof is trusted.
- Mutating action needs approval, audit, and persistent event evidence.

## Current Translation

Translate V7 terms this way:

| V7 term | Current owner |
| --- | --- |
| Thought profile | Soma Kernel / response contract in V8 runtime docs |
| Decision policy | Governance and proposal flow |
| Symbiote growth | reviewed Memory & Continuity |
| Host actuation adapter | governed capability/tool execution |
| Local Ollama contract | deployment/provider policy in ops and config docs |

## Compatibility Rules

- Do not add new canonical host-actuation subjects here.
- Do not add raw secrets, run logs, or local machine snapshots.
- Promote any active runtime requirement into V8 docs and `.state/V8_DEV_STATE.md`.
- Keep host/device/sensor signal origins explicit until normalized.

## Testing Memory

When reviving behavior from this V7 input, prove:
- provider health is visible and recoverable
- mutating actions use proposals/approvals
- retained outputs and memory survive refresh/restart
- operator errors are normalized and human-readable
- runtime signals include required metadata

## Close-Out Rule

If a slice cites this file, it must also name the V8 document that now owns the promoted requirement.
