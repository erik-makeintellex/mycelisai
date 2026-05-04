# V8 Runtime Inception And Kernel
> Navigation: [V8 Runtime Contracts](V8_RUNTIME_CONTRACTS.md)

Status: canonical detail for Inception and Soma Kernel.

## Inception Contract

An Inception is the live AI Organization. It is the object operators create, reopen, configure, and rely on.

Required fields conceptually include:
- stable organization id
- display name and purpose
- Soma identity/profile
- council/advisor composition
- team/department defaults
- provider policy scope
- memory and continuity defaults
- governance posture
- creation/source metadata

## Inception Boundaries

An Inception is not:
- a template
- a YAML file
- a raw model session
- a single run
- a transient UI tab

Templates can create or seed Inceptions, but only instantiated organizations enter runtime resolution.

## Soma Kernel Contract

Soma is the default operator counterpart for an Inception. Soma owns interaction continuity, safe routing, proposal posture, and visible recovery.

Required Soma behavior:
- answer ordinary non-mutating prompts directly
- ask for confirmation/proposal on protected actions
- route broad work into compact teams or workflow lanes
- keep action/results reviewable
- surface unavailable dependencies as blockers
- preserve identity across refresh/re-entry

## Council Relationship

Council/advisors support Soma. They may handle critique, planning, specialist calls, and policy-aware guidance, but they are not the default UI front door.

## Runtime Position

```text
Inception
  -> Soma Kernel
  -> Council/advisors
  -> Teams/departments/specialists
  -> Runs, activity, memory, retained outputs
```

## Testing Implications

Proof must cover:
- AI Organization creation/re-entry
- Soma direct answer
- protected proposal path
- compact team or workflow lane creation
- retained output review
- recovery from model/runtime failure
