# Documentation
> Navigation: [Project README](../README.md)

Use this page as the clean navigation layer between user guidance, developer guidance, testing guidance, and retained compatibility references.

## Docs TOC

- [User Guidance](#user-guidance)
- [Agent and Developer Guidance](#agent-and-developer-guidance)
- [Product and Licensing Guidance](#product-and-licensing-guidance)
- [Testing and Release Guidance](#testing-and-release-guidance)
- [Compatibility Boundary](#compatibility-boundary)

## User Guidance

These are the best entry points for someone using Mycelis through the product:

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

## Agent and Developer Guidance

These are the active authority surfaces for contributors changing or reviewing the repo:

- **Repository Entry Point**: `../README.md`
- **Repository Standards**: `../AGENTS.md`
- **Active Development State**: `../.state/V8_DEV_STATE.md`
- **Architecture Library Index**: `./architecture-library/ARCHITECTURE_LIBRARY_INDEX.md`
- **V8.2 Current State And Finalization PRD**: `./architecture-library/V8_2_CURRENT_STATE_AND_FINALIZATION_PRD.md`
- **V8.2 Production Architecture Target / B2+ Delivery Frame**: `../architecture/v8-2.md`
- **V8.2 Operational Embodiment Directive**: `./architecture-library/V8_2_OPERATIONAL_EMBODIMENT_DIRECTIVE.md`
- **V8 Runtime Contracts**: `./architecture-library/V8_RUNTIME_CONTRACTS.md`
- **V8 Config and Bootstrap Model**: `./architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`
- **V8 UI/API and Operator Experience Contract**: `./architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md`
- **V8 Directed Execution UI And Runtime Alignment Directive**: `./architecture-library/V8_DIRECTED_EXECUTION_UI_RUNTIME_ALIGNMENT_DIRECTIVE.md`
- **V8 Directed Execution Delivery Plan**: `./architecture-library/V8_DIRECTED_EXECUTION_DELIVERY_PLAN.md`
- **V8 Capability Manifest And Runtime Integration Standard**: `./architecture-library/V8_CAPABILITY_MANIFEST_AND_RUNTIME_INTEGRATION_STANDARD.md`
- **V8 Mycelis Search Capability Delivery Plan**: `./architecture-library/V8_MYCELIS_SEARCH_CAPABILITY_DELIVERY_PLAN.md`
- **V8 Memory Layer + Reflection Contract**: `./architecture-library/V8_MEMORY_LAYER_AND_REFLECTION_DELIVERY_CONTRACT.md`
- **V8 Trusted Memory Arbitration Contract**: `./architecture-library/V8_TRUSTED_MEMORY_ARBITRATION_AND_TEAM_VECTOR_CONTRACT.md`
- **V8.2 User Management And Enterprise Auth Module**: `./architecture-library/V8_2_USER_MANAGEMENT_AND_ENTERPRISE_AUTH_MODULE.md`
- **Operations**: `./architecture/OPERATIONS.md`
- **Local Dev Workflow**: `./LOCAL_DEV_WORKFLOW.md`
- **Logging Standard**: `./logging.md`

## Product and Licensing Guidance

Use these when the goal is product packaging, edition posture, or the paid-vs-self-hosted boundary:

- **Licensing & Editions**: `./licensing.md`
- **Governance System**: `./governance.md`
- **Governance & Trust**: `./user/governance-trust.md`
- **V8 Multi-User Identity + Soma Tenancy**: `./architecture-library/V8_MULTI_USER_IDENTITY_AND_SOMA_TENANCY.md`
- **V8.2 User Management And Enterprise Auth Module**: `./architecture-library/V8_2_USER_MANAGEMENT_AND_ENTERPRISE_AUTH_MODULE.md`
- **V8 Trusted Memory Arbitration Contract**: `./architecture-library/V8_TRUSTED_MEMORY_ARBITRATION_AND_TEAM_VECTOR_CONTRACT.md`

## Testing and Release Guidance

Use these when the goal is verification, release proof, or workflow-complete validation:

- **Release Handoff**: `./RELEASE_HANDOFF.md`
- **Testing Guide**: `./TESTING.md`
- **Remote User Testing Runbook**: `./REMOTE_USER_TESTING.md`
- **V8 Compose Personal Owner Deployment Test Plan**: `./architecture-library/V8_COMPOSE_PERSONAL_OWNER_DEPLOYMENT_TEST_PLAN.md`
- **V8 UI Testing Agentry Product Contract**: `./architecture-library/V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md`
- **V8 UI Team Full Test Set**: `./architecture-library/V8_UI_TEAM_FULL_TEST_SET.md`
- **V8 MVP Media, Team Output, And Template Registry**: `./architecture-library/V8_MVP_MEDIA_TEAM_OUTPUT_AND_TEMPLATE_REGISTRY.md`
- **Governance System**: `./governance.md`

## Compatibility Boundary

The active documentation surface no longer keeps superseded V7 doctrine files or archive drafts as readable authority. Historical state remains only where it is needed to explain migration evidence:

- **V7 Architecture PRD Compatibility Entry Point**: `../architecture/mycelis-architecture-v7.md`
- **Legacy V7 Development State**: `../.state/V7_DEV_STATE.md`

Guidance rules:
- user guidance should stay focused on using the product, not implementation internals
- agent/developer guidance should point to active V8.2 authority before retained compatibility evidence
- testing guidance should point to durable verification contracts rather than temporary execution notes
- V8.2/B2+-aligned work should prove operational embodiment through visible execution, durable outputs, recoverable runs, deployment trust, and one excellent canonical MVP workflow before expanding new doctrine
- any slice that changes runtime, validation, API meaning, or memory/governance posture must review and update the matching docs set in the same change
