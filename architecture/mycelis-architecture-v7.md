# Mycelis V7 PRD Index
> Navigation: [Project README](../README.md) | [Docs Home](../docs/README.md)

> Status: Canonical compatibility entrypoint
> Last Updated: 2026-03-07
> Purpose: Preserve the stable root PRD path while moving detailed planning authority into the modular architecture library.

## TOC

- [Why This File Changed](#why-this-file-changed)
- [Canonical Planning Surface](#canonical-planning-surface)
- [Core Product Position](#core-product-position)
- [Supporting Detailed Authorities](#supporting-detailed-authorities)
- [Current Delivery Framing](#current-delivery-framing)

## Why This File Changed

`architecture/mycelis-architecture-v7.md` previously acted as a massive all-in-one planning file.

That made it difficult to:
- keep target state current
- update UI concepts without creating more information swarm
- separate architecture from delivery proof
- maintain recurring-plan concepts without restating the whole platform

The detailed authority is now split into the modular library under `docs/architecture-library/`.

## Canonical Planning Surface

Start here:
- [Architecture Library Index](../docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md)

Primary detailed documents:
- [Target Deliverable V7](../docs/architecture-library/TARGET_DELIVERABLE_V7.md)
- [System Architecture V7](../docs/architecture-library/SYSTEM_ARCHITECTURE_V7.md)
- [Execution And Manifest Library V7](../docs/architecture-library/EXECUTION_AND_MANIFEST_LIBRARY_V7.md)
- [UI And Operator Experience V7](../docs/architecture-library/UI_AND_OPERATOR_EXPERIENCE_V7.md)
- [Delivery Governance And Testing V7](../docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md)
- [Team Execution And Global State Protocol V7](../docs/architecture-library/TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md)

## Core Product Position

Mycelis is targeting a governed execution system where:
- direct intent can return an answer
- governed intent can return a proposal
- executable intent can return a visible result
- blocked intent returns a concrete recovery path

The platform must support plans that are:
- `one_shot`
- `scheduled`
- `persistent_active`
- `event_driven`

The UI must help operators accomplish those outcomes without degenerating into a raw information swarm.

## Supporting Detailed Authorities

Use these specialized references alongside the modular library:
- [Overview](../docs/architecture/OVERVIEW.md)
- [Backend](../docs/architecture/BACKEND.md)
- [Frontend](../docs/architecture/FRONTEND.md)
- [Operations](../docs/architecture/OPERATIONS.md)
- [NATS Signal Standard V7](../docs/architecture/NATS_SIGNAL_STANDARD_V7.md)
- [UI Target And Transaction Contract V7](../docs/architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md)
- [Workflow Composer Delivery V7](../docs/architecture/WORKFLOW_COMPOSER_DELIVERY_V7.md)
- [Team Execution And Global State Protocol V7](../docs/architecture-library/TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md)

## Current Delivery Framing

The canonical phase order remains:
- `P0` operational foundation
- `P1` logging, error handling, and cleanup
- `P2` meta-agent-owned manifest pipeline
- `P3` workflow-composer onboarding and execution-facing UI
- `P4` release hardening

The detailed target and acceptance language now lives in:
- [Target Deliverable V7](../docs/architecture-library/TARGET_DELIVERABLE_V7.md)
- [Delivery Governance And Testing V7](../docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md)
