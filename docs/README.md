# Documentation
> Navigation: [Project README](../README.md)

Use this page as the repo documentation map. It separates product docs, repo/operator docs, architecture contracts, and retained compatibility evidence.

## Docs TOC

- [User Guidance](#user-guidance)
- [Repo Guidance](#repo-guidance)
- [Architecture Contracts](#architecture-contracts)
- [Testing And Release](#testing-and-release)
- [Agent And Maintainer Notes](#agent-and-maintainer-notes)
- [Compatibility Boundary](#compatibility-boundary)

## User Guidance

These are the best entry points for someone using Mycelis through the product or in-app docs browser:

- **User Docs Home**: `./user/README.md`
- **Deployment Method Selection**: `./user/deployment-methods.md`
- **Core Concepts**: `./user/core-concepts.md`
- **Using Soma Chat**: `./user/soma-chat.md`
- **Workflow Variants And Plan Memory**: `./user/workflow-variants-and-plan-memory.md`
- **Teams**: `./user/teams.md`
- **Governance & Trust**: `./user/governance-trust.md`
- **Automations**: `./user/automations.md`
- **Resources**: `./user/resources.md`
- **Memory**: `./user/memory.md`
- **Settings And Access**: `./user/settings-access.md`
- **Authentication Modes**: `./user/auth-modes.md`
- **System Status & Recovery**: `./user/system-status-recovery.md`
- **Run Timeline**: `./user/run-timeline.md`

## Repo Guidance

These are the active authority surfaces for contributors changing or reviewing the repo:

- **Repository Entry Point**: `../README.md`
- **Operations**: `./architecture/OPERATIONS.md`
- **Local Dev Workflow**: `./LOCAL_DEV_WORKFLOW.md`
- **Cognitive Architecture**: `./COGNITIVE_ARCHITECTURE.md`
- **API Reference**: `./API_REFERENCE.md`
- **Logging Standard**: `./logging.md`
- **Ops README**: `../ops/README.md`
- **Core README**: `../core/README.md`
- **Interface README**: `../interface/README.md`

## Architecture Contracts

The active architecture library is intentionally small. It is not a holding area for old version notes, doctrine fragments, or temporary execution plans.

- **Architecture Docs Index**: `./architecture-library/ARCHITECTURE_LIBRARY_INDEX.md`
- **V8.2 Production Architecture Target / B2+ Delivery Frame**: `../architecture/v8-2.md`
- **V8.3 Operational Embodiment PRD**: `./architecture-library/V8_3_OPERATIONAL_EMBODIMENT_PRD.md`
- **V8.3 Release Architecture Delivery Brief**: `./architecture-library/V8_3_RELEASE_ARCHITECTURE_DELIVERY_BRIEF.md`
- **V8.3 Product Manifestation Architecture Review**: `./architecture-library/V8_3_PRODUCT_MANIFESTATION_REVIEW.md`
- **V8.3 Soma User Experience Contract**: `./architecture-library/V8_3_SOMA_USER_EXPERIENCE_CONTRACT.md`
- **V8.3 MVP UI Runtime Delivery Plan**: `./architecture-library/V8_3_MVP_UI_RUNTIME_DELIVERY_PLAN.md`
- **V8 New-User Acceptance Matrix**: `./architecture-library/V8_NEW_USER_ACCEPTANCE_MATRIX.md`
- **V8.2 Soma UI Architecture Expression**: `./architecture-library/V8_2_SOMA_UI_ARCHITECTURE_EXPRESSION.md`
- **V8.2 Soma Team Interaction Contract**: `./architecture-library/V8_2_SOMA_TEAM_INTERACTION_CONTRACT.md`
- **V8 Runtime Contracts**: `./architecture-library/V8_RUNTIME_CONTRACTS.md`
- **V8 Config and Bootstrap Model**: `./architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`
- **V8 UI/API and Operator Experience Contract**: `./architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md`
- **V8 Capability Manifest And Runtime Integration Standard**: `./architecture-library/V8_CAPABILITY_MANIFEST_AND_RUNTIME_INTEGRATION_STANDARD.md`
- **V8 Secret Storage And Credential Boundary**: `./architecture-library/V8_SECRET_STORAGE_AND_CREDENTIAL_BOUNDARY.md`
- **V8 UI Team Full Test Set**: `./architecture-library/V8_UI_TEAM_FULL_TEST_SET.md`

## Testing And Release

Use these when the goal is verification, release proof, or workflow-complete validation:

- **Testing Guide**: `./TESTING.md`
- **Remote User Testing Runbook**: `./REMOTE_USER_TESTING.md`
- **V8 New-User Acceptance Matrix**: `./architecture-library/V8_NEW_USER_ACCEPTANCE_MATRIX.md`
- **V8 UI Testing Product Contract**: `./architecture-library/V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md`
- **Release Handoff**: `./RELEASE_HANDOFF.md`
- **Governance System**: `./governance.md`
- **Licensing & Editions**: `./licensing.md`

## Agent And Maintainer Notes

These files are useful for Codex, maintainers, and architecture reviewers, but they should not be treated as ordinary user-facing product docs:

- **Repository Standards**: `../AGENTS.md`
- **Active Development State**: `../.state/V8_DEV_STATE.md`
- **Legacy V7 Development State**: `../.state/V7_DEV_STATE.md`
- **V8.3 Dev Agentry Operational Directive**: `./architecture-library/V8_3_DEV_AGENTRY_OPERATIONAL_DIRECTIVE.md`
- **V8.3 Multi-Agentry Steering Doctrine**: `./architecture-library/V8_3_MULTI_AGENTRY_STEERING_DOCTRINE.md`

Keep these out of the normal operator path unless the task is repo maintenance, architecture review, or delivery-state inspection.

## Compatibility Boundary

The active documentation surface does not retain superseded V7 or older V8 topical docs as readable authority. Historical state remains only where it is needed to explain migration evidence:

- **V7 Architecture PRD Compatibility Entry Point**: `../architecture/mycelis-architecture-v7.md`
- **Legacy V7 Development State**: `../.state/V7_DEV_STATE.md`

Guidance rules:
- user guidance should stay focused on using the product, visible outputs, and recovery paths, not implementation internals
- repo guidance should point to operating and implementation contracts, not old planning notes
- architecture guidance should remain small, current, and directly referenced
- agent/maintainer notes should stay out of the default in-app operator docs unless they are intentionally surfaced for architecture review
- testing guidance should point to durable verification contracts rather than temporary execution notes
- if a removed doc contains a still-needed requirement, promote that requirement into the nearest active contract instead of restoring the old doc
