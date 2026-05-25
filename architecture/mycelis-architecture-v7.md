# Mycelis PRD Compatibility Index
> Navigation: [Project README](../README.md) | [Docs Home](../docs/README.md)

> Status: Compatibility entrypoint
> Last Updated: 2026-05-24
> Purpose: Preserve the stable historical PRD path while pointing active work to the current V8.3 embodiment and V8.2 architecture baseline.

## TOC

- [Why This File Exists](#why-this-file-exists)
- [Current Authority](#current-authority)
- [Compatibility Boundary](#compatibility-boundary)
- [Current Product Position](#current-product-position)
- [Delivery Framing](#delivery-framing)

## Why This File Exists

`architecture/mycelis-architecture-v7.md` is a retained path for older references and agent onboarding. It is not the place to grow new doctrine, topology language, or detailed product plans.

The prior monolithic V7 documents have been removed from active documentation because they were superseded by current V8/V8.3 contracts. New work should enter through the current modular library.

## Current Authority

Start here:
- [Architecture Library Index](../docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md)
- [V8.3 Operational Embodiment PRD](../docs/architecture-library/V8_3_OPERATIONAL_EMBODIMENT_PRD.md)
- [V8.3 Dev Agentry Operational Directive](../docs/architecture-library/V8_3_DEV_AGENTRY_OPERATIONAL_DIRECTIVE.md)
- [V8.3 Multi-Agentry Steering Doctrine](../docs/architecture-library/V8_3_MULTI_AGENTRY_STEERING_DOCTRINE.md)
- [V8.2 Full Production Architecture](v8-2.md)
- [V8 Runtime Contracts](../docs/architecture-library/V8_RUNTIME_CONTRACTS.md)
- [V8 UI/API and Operator Experience Contract](../docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md)
- [V8 Capability Manifest And Runtime Integration Standard](../docs/architecture-library/V8_CAPABILITY_MANIFEST_AND_RUNTIME_INTEGRATION_STANDARD.md)
- [V8.2 Soma UI Architecture Expression](../docs/architecture-library/V8_2_SOMA_UI_ARCHITECTURE_EXPRESSION.md)
- [V8.2 Soma Team Interaction Contract](../docs/architecture-library/V8_2_SOMA_TEAM_INTERACTION_CONTRACT.md)
- [V8 Development State](../.state/V8_DEV_STATE.md)

Supporting implementation references:
- [Overview](../docs/architecture/OVERVIEW.md)
- [Backend](../docs/architecture/BACKEND.md)
- [Frontend](../docs/architecture/FRONTEND.md)
- [Operations](../docs/architecture/OPERATIONS.md)
- [Testing](../docs/TESTING.md)

## Compatibility Boundary

Historical V7 development state remains in [.state/V7_DEV_STATE.md](../.state/V7_DEV_STATE.md) as migration evidence only. It should not be used as product authority when it conflicts with current V8/V8.3 docs.

Superseded archive drafts and old V7/V8 topical documents are intentionally deleted rather than retained as stale reference material. If a missing concept is still needed, promote it into the appropriate current V8/V8.3 contract instead of restoring an old document.

## Current Product Position

Mycelis is a Soma-centered governed cognitive operating environment.

The operator experience should collapse into:
- I tell Soma what I want.
- Soma directs the work safely.
- I can see what happened.
- I can trust the result.

Current V8.3 work prioritizes operational embodiment: async execution, visible runs, durable outputs, inspectable proof, executable recovery, deployment trust, and confidence provenance preparation.

## Delivery Framing

Every implementation slice should answer:
- Does this improve embodiment?
- Does this improve operator trust?
- Does this improve visible execution?
- Does this improve recoverability?
- Does this improve deployment reality?
- Does this reduce conceptual density?
- Does this strengthen Soma as the singular operating surface?

The active finalization roadmap lives in [V8.3 Operational Embodiment PRD](../docs/architecture-library/V8_3_OPERATIONAL_EMBODIMENT_PRD.md) and the implementation scoreboard lives in [V8 Development State](../.state/V8_DEV_STATE.md).
