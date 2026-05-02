# Mycelis Workflow Index
> Navigation: [Project README](../README.md) | [Docs Home](README.md)

> Status: Current routing index
> Purpose: Point operators and contributors to the active workflow authority without preserving the old monolithic UI specification.

This file is no longer the implementation authority for every screen, API call, or state transition. The old monolith mixed current Soma-first behavior with legacy catalogue, wiring, dashboard, and database planning detail. Current workflow meaning now lives in focused user docs and canonical architecture contracts.

Use this page when an existing link still lands on `docs/WORKFLOWS.md` and you need the right current document.

## TOC

- [Current Authority](#current-authority)
- [Workflow Routes](#workflow-routes)
- [Removed Monolith Sections](#removed-monolith-sections)
- [Validation Expectations](#validation-expectations)

## Current Authority

Use these documents instead of rebuilding a large workflow spec here:

| Need | Current source |
| --- | --- |
| Daily operator entrypoint | [User Docs Home](user/README.md) |
| Soma chat, protected prompts, generated outputs, and recovery | [Using Soma Chat](user/soma-chat.md) |
| Choosing direct Soma, one agent, compact teams, or multi-lane work | [Workflow Variants And Plan Memory](user/workflow-variants-and-plan-memory.md) |
| Team creation, lead-centered workspaces, and lane splitting | [Teams](user/teams.md) |
| MCP, Connected Tools, workspace files, AI engines, and deployment context | [Resources](user/resources.md) |
| Memory, continuity, retained knowledge, and reflection boundaries | [Memory](user/memory.md) |
| Approval posture, risk classes, and audit visibility | [Governance And Trust](user/governance-trust.md) |
| Run review, activity, and execution timelines | [Run Timeline](user/run-timeline.md) |
| API routes and payload shapes | [API Reference](API_REFERENCE.md) |
| Canonical V8 UI/API operator contract | [V8 UI/API/Operator Contract](architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md) |
| Workflow variants proof set | [V8 Workflow Variants Proof Set](architecture-library/V8_WORKFLOW_VARIANTS_AND_REBOOT_PROOF_SET.md) |
| Delivery proof and release validation | [Testing](TESTING.md) |

## Workflow Routes

The current product center is Soma-first:

1. Start in the Soma workspace or an AI Organization home.
2. Describe the outcome in natural language.
3. Let Soma answer directly, propose a compact team, or split broad work into lanes.
4. Confirm governed actions before mutation, private-data use, credentialed service use, recurring behavior, or durable memory promotion.
5. Review visible outputs in the chat response, team lead surface, Groups, retained artifacts, Activity, or Run Timeline.

Use the smallest workflow that preserves clarity and recovery:

| Workflow shape | Use when | Primary docs |
| --- | --- | --- |
| Direct Soma | One answer, recommendation, draft, or bounded explanation is enough. | [Using Soma Chat](user/soma-chat.md) |
| One context-rich agent | One complex output needs depth more than role separation. | [Workflow Variants And Plan Memory](user/workflow-variants-and-plan-memory.md) |
| Compact team | Planning, production, and review should be visible and resumable. | [Teams](user/teams.md) |
| Multi-lane workflow | Broad work needs separate outputs, handoffs, or parallel review. | [Workflow Variants And Plan Memory](user/workflow-variants-and-plan-memory.md) |
| Advanced blueprint or graph review | The operator intentionally needs mission/graph-level inspection. | [Meta-Agent And Blueprints](user/meta-agent-blueprint.md) |
| Connected tool review | The operator needs to inspect MCP/tool capability posture. | [Resources](user/resources.md) |
| Governance review | The operator needs approval, risk, audit, or protected-action context. | [Governance And Trust](user/governance-trust.md) |
| Runtime review | The operator needs run history, bus summaries, or active workflow activity. | [Run Timeline](user/run-timeline.md) |

## Removed Monolith Sections

The previous file contained detailed sections for:

- Agent Catalogue
- MCP Tool Registry
- Team Builder
- Mission Authoring
- Mission Monitoring
- Governance
- Direct Chat
- Settings
- Team Actuation and Output Viewer
- Three-Tier Memory
- cross-cutting Zustand, SSE, navigation, API, and database implementation notes

Those sections were not preserved here because they duplicated or contradicted newer Soma-first user guidance, API reference material, architecture contracts, and in-app docs. Legacy planning detail belongs in archived historical docs when it is needed for migration context, not in the active `/docs` workflow entry.

## Validation Expectations

Workflow-affecting slices should still review the owning docs in the same change:

- `README.md`
- `.state/V8_DEV_STATE.md`
- this index when a workflow entrypoint changes
- the focused user or architecture doc that owns the changed behavior
- `interface/lib/docsManifest.ts` when in-app docs visibility changes
- `docs/API_REFERENCE.md` when endpoint or payload meaning changes
- `docs/TESTING.md`, `docs/architecture/OPERATIONS.md`, and `ops/README.md` when task-running or validation behavior changes

For proof, keep the docs link tests green and use the focused workflow/browser tests named by the owning doc. Do not grow this file back into a screen-by-screen monolith.
