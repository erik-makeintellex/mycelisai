# Architecture Library Index

> Status: Canonical
> Last Updated: 2026-03-10
> Purpose: Replace monolithic planning with a modular, cross-linked architecture and target-delivery library.

## Why This Library Exists

Mycelis had accumulated large planning documents that mixed target state, implementation detail, UI concepts, and delivery gates into a single place. That made updates expensive and encouraged drift.

This library is the new canonical planning surface for ongoing development.

Use it to answer:
- what the product is supposed to become
- how the system is structured
- how execution and recurring plans work
- how the UI should feel and behave
- how delivery is proven in tests and gates

The compatibility PRD entrypoint remains [mycelis-architecture-v7.md](../../mycelis-architecture-v7.md), but the detailed authority now lives here.

## Canonical Documents

| Document | Load When | Focus |
| --- | --- | --- |
| [Target Deliverable V7](TARGET_DELIVERABLE_V7.md) | defining end state, phases, and product outcomes | full target delivery, plan modes, success criteria |
| [System Architecture V7](SYSTEM_ARCHITECTURE_V7.md) | working on backend/runtime/storage/NATS/deployment | platform layers, contracts, deployment model |
| [Execution And Manifest Library V7](EXECUTION_AND_MANIFEST_LIBRARY_V7.md) | working on workflows, runs, schedules, active plans, team manifests | execution lifecycle, manifest lifecycle, recurring behavior |
| [Intent To Manifestation And Team Interaction V7](INTENT_TO_MANIFESTATION_AND_TEAM_INTERACTION_V7.md) | integrating Soma-first manifestation, module bindings, and created-team interaction | intent flow contract, module abstraction, team communication UX |
| [UI And Operator Experience V7](UI_AND_OPERATOR_EXPERIENCE_V7.md) | working on UI, UX, interaction flows, onboarding | intuitive operator journeys, anti-swarm rules, terminal states |
| [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md) | planning slices, acceptance criteria, test proof | delivery gates, evidence, product-aligned testing |
| [Next Execution Slices V7](NEXT_EXECUTION_SLICES_V7.md) | choosing the next concrete implementation slice | working queue with scoped files, development refs, and testing refs |
| [Team Execution And Global State Protocol V7](TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md) | coordinating multi-lane execution and state discipline | team execution architecture, state-file maintenance, deep-testing obligations |
| [MVP Release Strike Team Plan V7](MVP_RELEASE_STRIKE_TEAM_PLAN_V7.md) | running coordinated MVP delivery lanes | strike-team ownership, communication cadence, and state-file discipline |
| [MVP Integration And Toolship Execution Plan V7](MVP_INTEGRATION_AND_TOOLSHIP_EXECUTION_PLAN_V7.md) | executing AI + internal toolship + service-connection hardening for non-test users | integration phases, acceptance gates, deep-testing matrix, and exit criteria |
| [UI Generation And Testing Execution Plan V7](UI_GENERATION_AND_TESTING_EXECUTION_PLAN_V7.md) | executing deterministic UI build flow and deep route-level test proof | generation contracts, route coverage priorities, and MVP UI gate criteria |

## How To Use This Library

1. Start with [Target Deliverable V7](TARGET_DELIVERABLE_V7.md) to confirm the intended end state.
2. Use [System Architecture V7](SYSTEM_ARCHITECTURE_V7.md) to align runtime and deployment assumptions.
3. Use [Execution And Manifest Library V7](EXECUTION_AND_MANIFEST_LIBRARY_V7.md) for workflows, activation, schedules, and always-on plans.
4. Use [Intent To Manifestation And Team Interaction V7](INTENT_TO_MANIFESTATION_AND_TEAM_INTERACTION_V7.md) for the canonical path from intent through manifestation and created-team operation.
5. Use [UI And Operator Experience V7](UI_AND_OPERATOR_EXPERIENCE_V7.md) before changing operator-facing surfaces.
6. Use [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md) before implementing or accepting a slice.
7. Use [Next Execution Slices V7](NEXT_EXECUTION_SLICES_V7.md) when deciding what to do next or which docs/tests to load for a slice.
8. Use [Team Execution And Global State Protocol V7](TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md) when coordinating lane ownership, state-file updates, and deep-testing coverage.
9. Use [MVP Release Strike Team Plan V7](MVP_RELEASE_STRIKE_TEAM_PLAN_V7.md) when executing the active MVP release push and assigning lane responsibilities.
10. Use [MVP Integration And Toolship Execution Plan V7](MVP_INTEGRATION_AND_TOOLSHIP_EXECUTION_PLAN_V7.md) when delivering AI/toolship/service integration slices toward non-test-team usability.
11. Use [UI Generation And Testing Execution Plan V7](UI_GENERATION_AND_TESTING_EXECUTION_PLAN_V7.md) when defining new UI surfaces or raising route-level browser test depth.

## Supporting Specialized Docs

This library does not replace specialized implementation references. It sits above them.

Use these existing docs when needed:
- [Overview](../architecture/OVERVIEW.md)
- [Backend](../architecture/BACKEND.md)
- [Frontend](../architecture/FRONTEND.md)
- [Operations](../architecture/OPERATIONS.md)
- [NATS Signal Standard V7](../architecture/NATS_SIGNAL_STANDARD_V7.md)
- [UI Target And Transaction Contract V7](../architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md)
- [Workflow Composer Delivery V7](../architecture/WORKFLOW_COMPOSER_DELIVERY_V7.md)

## Library Principles

The library is optimized for delivery, not archival completeness.

- one concept per document
- operator outcome before internal complexity
- recurring and always-on behavior treated as first-class
- no planning-only UI outcomes
- docs must be maintainable by future agents without reading a giant file
- README links only the few docs needed to orient new work

## Canonical Outcome

The intended product is a governed execution system where:
- intent becomes a direct answer, a governed proposal, an execution result, or a blocker
- workflows can be one-shot, scheduled, persistent active, or event-driven
- every action preserves lineage across UI, API, NATS, and persistence
- operators can understand what is happening without drowning in telemetry

Next:
- [Target Deliverable V7](TARGET_DELIVERABLE_V7.md)
