# Architecture Library Index
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Status: Canonical
> Last Updated: 2026-04-27
> Purpose: Maintain one precise, durable documentation surface for product target, runtime structure, UI contract, and delivery proof.

## TOC

- [Why This Library Exists](#why-this-library-exists)
- [Layered Truth](#layered-truth)
- [Canonical Documents](#canonical-documents)
- [How To Use This Library](#how-to-use-this-library)
- [Supporting Specialized Docs](#supporting-specialized-docs)
- [Library Principles](#library-principles)
- [Canonical Outcome](#canonical-outcome)

## Why This Library Exists

Mycelis previously accumulated planning notes, strike-team plans, execution briefs, and temporary runbooks alongside its durable architecture docs. That made the active documentation surface noisy and easy to drift.

This index is now intentionally narrower:
- precise reference docs stay here
- temporary planning notes do not
- `README.md` remains the primary inception document for active work
- [mycelis-architecture-v7.md](../../mycelis-architecture-v7.md) remains the stable PRD compatibility entrypoint

## Layered Truth

- [README](../../README.md) is the primary inception document for active development.
- [V8.1 Living Organization Architecture](V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md) is the current release architecture.
- [V8.2 Full Production Architecture](../../v8-2.md) is the canonical full production architecture and full actuation target.
- [V8 Development State](../../V8_DEV_STATE.md) is the live implementation scoreboard.

V8.2 work may proceed before V8.1 release lock when it is modular and proof-producing. Keep each slice attached to one owning contract area, keep non-promoted V8.2 capability out of the default V8.1 operator surface, and update `V8_DEV_STATE.md` when a module boundary changes status.

## Canonical Documents

| Document | Load When | Focus |
| --- | --- | --- |
| [Target Deliverable V7](TARGET_DELIVERABLE_V7.md) | defining end state, phases, and product outcomes | target delivery, success criteria, and phase framing |
| [System Architecture V7](SYSTEM_ARCHITECTURE_V7.md) | working on backend, runtime, storage, NATS, or deployment | platform layers, persistence, runtime boundaries, and deployment posture |
| [Execution And Manifest Library V7](EXECUTION_AND_MANIFEST_LIBRARY_V7.md) | working on workflows, runs, schedules, active plans, or manifests | execution lifecycle, manifest lifecycle, and recurring behavior |
| [Intent To Manifestation And Team Interaction V7](INTENT_TO_MANIFESTATION_AND_TEAM_INTERACTION_V7.md) | integrating Soma-first manifestation and created-team interaction | intent flow, module abstraction, and team interaction contract |
| [UI And Operator Experience V7](UI_AND_OPERATOR_EXPERIENCE_V7.md) | working on UI, UX, onboarding, or operator terminology | operator journeys, simplification rules, and anti-swarm guidance |
| [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md) | defining acceptance criteria, delivery gates, or test evidence | delivery proof model, evidence expectations, and product-aligned testing |
| [Team Execution And Global State Protocol V7](TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md) | coordinating multi-lane execution and state maintenance | execution discipline, state-file rules, and deep-testing obligations |
| [V8 Runtime Contracts](V8_RUNTIME_CONTRACTS.md) | shaping new V8 runtime contracts | inception, kernel, council, provider policy, and continuity contracts |
| [V8 Config and Bootstrap Model](V8_CONFIG_AND_BOOTSTRAP_MODEL.md) | planning configuration sources, templates, bootstrap behavior, or inheritance | V7-to-V8 bootstrap migration, template vs instantiated organization, and precedence rules |
| [V8 UI/API and Operator Experience Contract](V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md) | changing first-run flow, AI Organization creation, Soma-primary workspace, or screen-to-API mapping | canonical V8 operator PRD, source-of-truth layering, and UI/API contracts |
| [V8.1 Living Organization Architecture](V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md) | defining the current release architecture | current release architecture for loops, continuity, capabilities, and bounded automations |
| [V8 Compact Team Orchestration And Defaults](V8_COMPACT_TEAM_ORCHESTRATION_AND_DEFAULTS.md) | defining how team creation should stay compact by default | compact team shape guidance, broad-ask splitting, and Soma/Council orchestration over NATS |
| [V8 Teamed Agentry Workflow Advantage](V8_TEAMED_AGENTRY_WORKFLOW_ADVANTAGE.md) | deciding when a team actually beats one strong context-rich agent | workflow variants, team-vs-single-agent boundary, and complex patterns where teamed agentry wins |
| [V8 Workflow Variants And Reboot Proof Set](V8_WORKFLOW_VARIANTS_AND_REBOOT_PROOF_SET.md) | demonstrating the same objective across direct Soma, compact team, and multi-lane workflow shapes | comparative workflow proof and reboot-safe resume validation |
| [V8 Universal Soma And Context Model PRD](V8_UNIVERSAL_SOMA_AND_CONTEXT_MODEL_PRD.md) | working on Central Soma, context switching, or scoped execution | persistent Soma/Council identity and governed context model |
| [V8 Multi-User Identity And Soma Tenancy](V8_MULTI_USER_IDENTITY_AND_SOMA_TENANCY.md) | planning multi-user access, enterprise federation, or shared Soma identity | SAML/SSO, break-glass admins, modular user management, and one-Soma tenancy |
| [V8.2 User Management And Enterprise Auth Module](V8_2_USER_MANAGEMENT_AND_ENTERPRISE_AUTH_MODULE.md) | defining enterprise-ready users, roles, SSO, SAML, OIDC/OAuth, Entra ID, Google Workspace, GitHub, or future SCIM | pluggable authentication providers, Mycelis-owned authorization, edition strategy, RBAC, approval authority, and enterprise auth validation |
| [V8 Memory Layer And Reflection Delivery Contract](V8_MEMORY_LAYER_AND_REFLECTION_DELIVERY_CONTRACT.md) | planning memory layers, reflection synthesis, exchange-backed learning candidates, or memory promotion | explicit `SOMA_MEMORY`, `AGENT_MEMORY`, `PROJECT_MEMORY`, `REFLECTION_MEMORY`, candidate-first reflection, and promotion guardrails |
| [V8 Trusted Memory Arbitration And Team Vector Contract](V8_TRUSTED_MEMORY_ARBITRATION_AND_TEAM_VECTOR_CONTRACT.md) | defining Soma personal memory, team-shared vector memory, governed swarm doctrine, anchor verification, or memory arbitration | trusted memory control plane, evidence anchors, precedence rules, and bounded-growth policy |
| [V8 Home Docker Compose Runtime](V8_HOME_DOCKER_COMPOSE_RUNTIME.md) | supporting the single-host compose runtime | compose tasking, persistence, monitoring, and operator expectations |
| [V8 Self-Hosted Runtime Delivery Program](V8_SELF_HOSTED_RUNTIME_DELIVERY_PROGRAM.md) | coordinating the compact delivery team for real deployment work while V8.2 runtime modules continue | release-gating Compose proof, modular Kubernetes scale-up work, team ownership, management cadence, and acceptance gates |
| [V8 Mycelis Search Capability Delivery Plan](V8_MYCELIS_SEARCH_CAPABILITY_DELIVERY_PLAN.md) | putting the architecture/runtime/interface/validation teams together around owned search capability | Mycelis Search API, local-source search, optional SearXNG, optional Brave, and capability/MCP delivery gates |
| [V8 Enterprise Self-Hosted Kubernetes Delivery Plan](V8_ENTERPRISE_SELF_HOSTED_KUBERNETES_DELIVERY_PLAN.md) | turning the Helm/Kubernetes lane into an enterprise-compatible deployment contract | local `k3d` validation, chart promotion rules, compact team ownership, and acceptance gates |
| [V8 Compose Personal Owner Deployment Test Plan](V8_COMPOSE_PERSONAL_OWNER_DEPLOYMENT_TEST_PLAN.md) | validating a personal-owner Compose deployment from data-plane-only startup through near-enterprise product workflows | data-plane-only Postgres/NATS launch, owner configuration, app bring-up, workflow proof, recovery, and evidence matrix |
| [V8 MVP Media, Team Output, And Template Registry](V8_MVP_MEDIA_TEAM_OUTPUT_AND_TEMPLATE_REGISTRY.md) | planning user-output-first media/team demos, Ollama role routing, or DB-backed conversation templates | model-role routing, media-engine boundary, team-managed output proof, and conversation-template registry |
| [V8 UI Testing Agentry Product Contract](V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md) | defining what browser validation must prove | Soma-first browser proof contract for answers, governed actions, continuity, and audit visibility |
| [V8 UI Team Full Test Set](V8_UI_TEAM_FULL_TEST_SET.md) | running the complete UI validation set | browser workflows, compose/runtime proof, evidence packaging, and verdict rules |
| [V8.2 Full Production Architecture](../../v8-2.md) | checking whether a surface belongs to the full target rather than the current release | Full Production Architecture (Canonical Target) for distributed execution, governed learning, and actuation scope |

## How To Use This Library

1. Start with [Target Deliverable V7](TARGET_DELIVERABLE_V7.md) to confirm the intended end state.
2. Use [System Architecture V7](SYSTEM_ARCHITECTURE_V7.md) to align runtime and deployment assumptions.
3. Use [Execution And Manifest Library V7](EXECUTION_AND_MANIFEST_LIBRARY_V7.md) for workflows, activation, schedules, and always-on plans.
4. Use [Intent To Manifestation And Team Interaction V7](INTENT_TO_MANIFESTATION_AND_TEAM_INTERACTION_V7.md) for the canonical path from intent through manifestation and created-team operation.
5. Use [UI And Operator Experience V7](UI_AND_OPERATOR_EXPERIENCE_V7.md) before changing operator-facing surfaces.
6. Use [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md) before implementing or accepting a slice.
7. Use [Team Execution And Global State Protocol V7](TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md) when coordinating lane ownership, state-file updates, and deep-testing coverage.
8. Use [V8 Runtime Contracts](V8_RUNTIME_CONTRACTS.md) when defining the new canonical V8 runtime contracts during migration.
9. Use [V8 Config and Bootstrap Model](V8_CONFIG_AND_BOOTSTRAP_MODEL.md) when translating V7 assumptions through the template -> instantiation -> inheritance -> precedence pipeline.
10. Use [V8 UI/API and Operator Experience Contract](V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md) when defining first-run flow, Soma-primary workspace behavior, advanced-mode boundaries, or screen-to-API mapping.
11. Use [V8.1 Living Organization Architecture](V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md) when checking the current release architecture.
12. Use [V8 Compact Team Orchestration And Defaults](V8_COMPACT_TEAM_ORCHESTRATION_AND_DEFAULTS.md) when shaping team defaults, broad-ask decomposition, or Soma/Council coordination over NATS.
13. Use [V8 Teamed Agentry Workflow Advantage](V8_TEAMED_AGENTRY_WORKFLOW_ADVANTAGE.md) when deciding whether a request should stay direct Soma, become a compact team, or split into several coordinated lanes.
14. Use [V8 Workflow Variants And Reboot Proof Set](V8_WORKFLOW_VARIANTS_AND_REBOOT_PROOF_SET.md) when you need one concrete demonstration set that compares direct Soma, compact teams, multi-lane workflows, and reboot-safe resume behavior.
15. Use [V8 Universal Soma And Context Model PRD](V8_UNIVERSAL_SOMA_AND_CONTEXT_MODEL_PRD.md) when reasoning about one persistent Soma/Council pair across organizations and deployments.
16. Use [V8 Multi-User Identity And Soma Tenancy](V8_MULTI_USER_IDENTITY_AND_SOMA_TENANCY.md) when reasoning about enterprise federation, local break-glass admins, or one shared Soma persona across many users.
17. Use [V8.2 User Management And Enterprise Auth Module](V8_2_USER_MANAGEMENT_AND_ENTERPRISE_AUTH_MODULE.md) when defining auth providers, enterprise SSO, SAML, OIDC/OAuth, internal RBAC, approval authority, or user-context policy.
18. Use [V8 Memory Layer And Reflection Delivery Contract](V8_MEMORY_LAYER_AND_REFLECTION_DELIVERY_CONTRACT.md) when defining memory-layer boundaries, reflection candidates, or promotion rules.
19. Use [V8 Trusted Memory Arbitration And Team Vector Contract](V8_TRUSTED_MEMORY_ARBITRATION_AND_TEAM_VECTOR_CONTRACT.md) when defining Soma personal continuity, team-shared vector memory, governed swarm doctrine, or trusted memory arbitration.
20. Use [V8 Home Docker Compose Runtime](V8_HOME_DOCKER_COMPOSE_RUNTIME.md) when validating the supported single-host compose path.
21. Use [V8 Self-Hosted Runtime Delivery Program](V8_SELF_HOSTED_RUNTIME_DELIVERY_PROGRAM.md) when coordinating the active compact team delivering Compose release proof and modular self-hosted Kubernetes runtime truth.
22. Use [V8 Mycelis Search Capability Delivery Plan](V8_MYCELIS_SEARCH_CAPABILITY_DELIVERY_PLAN.md) when coordinating owned search capability, local-source search, optional SearXNG, optional hosted search, and capability/MCP testing gates.
23. Use [V8 Enterprise Self-Hosted Kubernetes Delivery Plan](V8_ENTERPRISE_SELF_HOSTED_KUBERNETES_DELIVERY_PLAN.md) when coordinating `k3d` local validation, enterprise chart readiness, and promoted cluster deployment rules.
24. Use [V8 Compose Personal Owner Deployment Test Plan](V8_COMPOSE_PERSONAL_OWNER_DEPLOYMENT_TEST_PLAN.md) when validating data-plane-only Postgres/NATS startup, owner endpoint/credential configuration, and near-enterprise Compose workflow proof.
25. Use [V8 MVP Media, Team Output, And Template Registry](V8_MVP_MEDIA_TEAM_OUTPUT_AND_TEMPLATE_REGISTRY.md) when proving user-output-first media delivery, team-managed outputs, Ollama role routing, or DB-backed conversation templates.
26. Use [V8 UI Testing Agentry Product Contract](V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md) when defining browser expectations for Soma-first flows.
27. Use [V8 UI Team Full Test Set](V8_UI_TEAM_FULL_TEST_SET.md) when you need one complete UI/browser validation set.
28. Use [V8.2 Full Production Architecture](../../v8-2.md) when checking whether a surface belongs to the canonical full production target rather than the current release.

Execution governance reminder:
- [Delivery Governance And Testing V7](DELIVERY_GOVERNANCE_AND_TESTING_V7.md) and [Team Execution And Global State Protocol V7](TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md) remain authoritative migration inputs for delivery discipline.
- Apply the V8 bootstrap pipeline from [V8 Config and Bootstrap Model](V8_CONFIG_AND_BOOTSTRAP_MODEL.md) whenever configuration, templates, or runtime state are involved.
- Update [V8_DEV_STATE.md](../../V8_DEV_STATE.md) whenever release posture or validation truth changes.
- For V8.2-aligned slices, name the module boundary being advanced before implementation begins: runtime/deployment, memory/learning, team/workflow, capability/MCP, advanced UI, or governance/trust.

## Supporting Specialized Docs

This library sits above specialized implementation references. Use these when the work is narrower than the canonical architecture set:

- [Overview](../architecture/OVERVIEW.md)
- [Backend](../architecture/BACKEND.md)
- [Frontend](../architecture/FRONTEND.md)
- [Operations](../architecture/OPERATIONS.md)
- [NATS Signal Standard V7](../architecture/NATS_SIGNAL_STANDARD_V7.md)
- [UI Target And Transaction Contract V7](../architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md)
- [Workflow Composer Delivery V7](../architecture/WORKFLOW_COMPOSER_DELIVERY_V7.md)

## Library Principles

- one durable concept per document
- operator outcome before internal complexity
- no temporary planning notes in the canonical library
- README links only the few docs needed to orient new work
- tests, docs, and the in-app docs manifest must agree on the same authority set

## Canonical Outcome

The intended product is a governed execution system where:
- intent becomes a direct answer, a governed proposal, an execution result, or a blocker
- workflows can be one-shot, scheduled, persistent active, or event-driven
- every action preserves lineage across UI, API, NATS, and persistence
- operators can understand what is happening without drowning in telemetry

Next:
- [Target Deliverable V7](TARGET_DELIVERABLE_V7.md)
