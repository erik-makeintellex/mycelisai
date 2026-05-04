# V8 Runtime Provider And Continuity
> Navigation: [V8 Runtime Contracts](V8_RUNTIME_CONTRACTS.md) | [V8 Config and Bootstrap Model](V8_CONFIG_AND_BOOTSTRAP_MODEL.md)

Status: canonical detail for provider policy and continuity state.

## Provider Policy

Provider Policy governs model/media routing for an instantiated organization and its lower scopes.

Inputs may come from:
- bootstrap bundles/templates
- static config
- database/runtime state
- operator input
- deployment/env overrides

Deployment/env overrides configure endpoints, credentials references, and profile defaults. They do not become runtime organization truth by themselves.

## Scope Layers

Provider policy can apply at:
- Inception
- Soma Kernel
- Council/advisor role
- team/department
- agent/specialist

Lower scopes may refine provider selection only where higher policy permits it.

## Supported Provider Classes

Current provider classes include:
- local Ollama/OpenAI-compatible text endpoints
- hosted OpenAI-compatible endpoints
- Anthropic
- Google/Gemini
- optional media providers
- future self-hosted providers

Secrets must be referenced through env/secret stores, not embedded in docs, UI, logs, or committed config.

## Continuity State

Continuity includes:
- organization identity
- Soma identity and response posture
- reviewed memory
- retained outputs/artifacts
- recent activity/run context
- response-style defaults
- continuity settings that survive restart/refresh

Continuity does not include:
- one-off execution context as inherited config
- raw transient telemetry
- unreviewed private data silently promoted to memory

## Failure Behavior

Provider or continuity failures should surface as:
- `blocker`
- degraded but explicit read-only state
- retry/recovery guidance
- audit-visible failure event when tied to a run

Raw provider errors should not leak into operator-facing UI.
