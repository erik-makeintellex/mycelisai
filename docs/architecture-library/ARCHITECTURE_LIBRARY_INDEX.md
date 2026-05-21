# Architecture Docs Index
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Status: Canonical
> Last Updated: 2026-05-17
> Purpose: Keep the active architecture surface small enough to use.

## TOC

- [Purpose](#purpose)
- [Canonical Set](#canonical-set)
- [Repo Docs](#repo-docs)
- [UI Docs](#ui-docs)
- [Cleanup Rule](#cleanup-rule)

## Purpose

This directory is no longer a holding area for every V8 plan, doctrine note, execution brief, or split detail page.

The active architecture library keeps only documents that are directly used for current delivery decisions. User-facing guidance belongs under `docs/user/`. Repo operations and implementation guidance belong under `docs/`, `docs/architecture/`, `ops/`, `core/`, and `interface/`. Historical evidence belongs in `.state/` or Git history, not as active documentation.

## Canonical Set

| Document | Use For |
| --- | --- |
| [V8.2 Current State And Finalization PRD](V8_2_CURRENT_STATE_AND_FINALIZATION_PRD.md) | current implementation truth, target finalization state, workstreams, blockers, exit gates |
| [V8.2 Finalization Delivery Plan](V8_2_FINALIZATION_DELIVERY_PLAN.md) | team sequence, GUI matrix, runtime gates, and orchestration rules |
| [V8.2 Finalization Concretization Contract](V8_2_FINALIZATION_CONCRETIZATION_CONTRACT.md) | first demo slice, ExecutionContract, ProofArtifact, UI states, degraded lifecycle, close-out template |
| [V8.2 Operational Embodiment Directive](V8_2_OPERATIONAL_EMBODIMENT_DIRECTIVE.md) | visible execution, durable outputs, proof, recovery, deployment trust |
| [V8.2 Soma UI Architecture Expression](V8_2_SOMA_UI_ARCHITECTURE_EXPRESSION.md) | ideal Soma interaction model, one-window UI architecture, active work, outputs, proof, teams |
| [V8.2 UI Research And Target Experience Review](V8_2_UI_RESEARCH_AND_TARGET_EXPERIENCE_REVIEW.md) | research-backed target UX, expression-to-output workbench architecture, default/Inspect rules, delivery plan |
| [V8.2 Soma Team Interaction Contract](V8_2_SOMA_TEAM_INTERACTION_CONTRACT.md) | talking with new or running teams, TeamWorkItem, TeamInteraction, TeamStatusEvent, recovery |
| [V8 Runtime Contracts](V8_RUNTIME_CONTRACTS.md) | instantiated organization runtime truth, Soma, Council, provider policy, continuity |
| [V8 Config and Bootstrap Model](V8_CONFIG_AND_BOOTSTRAP_MODEL.md) | bootstrap bundles, templates, inheritance, precedence, startup truth |
| [V8 UI/API and Operator Experience Contract](V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md) | operator terminology, response states, screen/API expectations, advanced boundary |
| [V8 Capability Manifest And Runtime Integration Standard](V8_CAPABILITY_MANIFEST_AND_RUNTIME_INTEGRATION_STANDARD.md) | governed capabilities, MCP/custom tools, output normalization, permission and recovery posture |
| [V8 UI Testing Agentry Product Contract](V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md) | browser-visible product proof standard |
| [V8 UI Team Full Test Set](V8_UI_TEAM_FULL_TEST_SET.md) | full GUI validation contract and verdict format |
| [Source File Size And Indexing Plan](SOURCE_FILE_SIZE_AND_INDEXING_PLAN.md) | repo-size hygiene and decomposition rules |

The full production architecture target lives outside this folder at [architecture/v8-2.md](../../architecture/v8-2.md). The live implementation scoreboard is [.state/V8_DEV_STATE.md](../../.state/V8_DEV_STATE.md).

## Repo Docs

Use these for implementation work:

- [README](../../README.md): repo inception, command contract, development rules
- [Docs Home](../README.md): organized navigation across user, repo, architecture, testing, and release docs
- [Operations](../architecture/OPERATIONS.md): task ownership, runtime lanes, deployment operations
- [Testing](../TESTING.md): validation policy and proof commands
- [API Reference](../API_REFERENCE.md): endpoint and payload behavior
- [Local Development Workflow](../LOCAL_DEV_WORKFLOW.md): setup, runtime choices, ports, troubleshooting
- [Ops README](../../ops/README.md): Invoke task implementation ownership

## UI Docs

Use these for product and in-app documentation:

- [User Docs Home](../user/README.md)
- [Using Soma Chat](../user/soma-chat.md)
- [Teams](../user/teams.md)
- [Resources](../user/resources.md)
- [Memory](../user/memory.md)
- [Governance & Trust](../user/governance-trust.md)
- [Settings And Access](../user/settings-access.md)
- [Authentication Modes](../user/auth-modes.md)
- [System Status & Recovery](../user/system-status-recovery.md)
- [Run Timeline](../user/run-timeline.md)

`interface/lib/docsManifest.ts` should expose only this curated set plus stable user/repo docs. Do not surface stale planning notes in the UI.

## Cleanup Rule

When a document is not directly referenced by README, Docs Home, the in-app docs manifest, an owning implementation doc, or an active test contract, delete it instead of preserving it as a stale active page. If a removed document contains a still-valid requirement, promote that requirement into the nearest canonical document above.
