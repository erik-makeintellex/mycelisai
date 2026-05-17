# Architecture Library Index
> Navigation: [Project README](../../README.md) | [Docs Home](../README.md)

> Status: Canonical
> Last Updated: 2026-05-16
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

Mycelis previously accumulated planning notes, strike-team plans, execution briefs, historical doctrine, and temporary runbooks alongside durable architecture docs. That made the active documentation surface noisy and easy to drift.

This index is intentionally narrow:
- active V8.2 authority stays here
- stale V7 topical documents and archive drafts are not retained
- `README.md` remains the primary inception document for active work
- [architecture/mycelis-architecture-v7.md](../../architecture/mycelis-architecture-v7.md) remains only a compatibility entrypoint for old references

## Layered Truth

- [README](../../README.md) is the primary inception document for active development.
- [V8.2 Full Production Architecture](../../architecture/v8-2.md) is the canonical full production architecture, full actuation target, and current B2+ delivery frame.
- [V8.2 Current State And Finalization PRD](V8_2_CURRENT_STATE_AND_FINALIZATION_PRD.md) is the architecture-team map of current implementation truth, concretization objects, finalization gaps, workstreams, risks, and exit gates.
- [V8.2 Finalization Concretization Contract](V8_2_FINALIZATION_CONCRETIZATION_CONTRACT.md) pins the first demo slice, schema fields, UI states, degraded lifecycle, deployment-trust owner, team prominence rule, and slice close-out template.
- [V8.2 Operational Embodiment Directive](V8_2_OPERATIONAL_EMBODIMENT_DIRECTIVE.md) constrains work away from doctrine expansion and toward visible, recoverable, trustworthy operation.
- [V8.2 Soma Team Interaction Contract](V8_2_SOMA_TEAM_INTERACTION_CONTRACT.md) defines how Soma, Council, operators, and runtime teams inspect, steer, recover, and archive active work without exposing raw orchestration topology as the default experience.
- [V8.1 Living Organization Architecture](V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md) is the foundation and compatibility baseline for the Soma-primary operator surface.
- [V8 Development State](../../.state/V8_DEV_STATE.md) is the live implementation scoreboard. Use its active snapshot and immediate next actions before reading dated historical boards.

## Canonical Documents

| Document | Load When | Focus |
| --- | --- | --- |
| [V8.2 Current State And Finalization PRD](V8_2_CURRENT_STATE_AND_FINALIZATION_PRD.md) | aligning architecture, product, and implementation teams | current state, ExecutionContract, ProofArtifact, UI response states, finalization target, blockers, workstreams, risks, and exit gates |
| [V8.2 Finalization Concretization Contract](V8_2_FINALIZATION_CONCRETIZATION_CONTRACT.md) | starting finalization implementation or reviewing whether a slice is concrete enough | first demo slice, ExecutionContract, ProofArtifact, CapabilityManifestState, confidence provenance, UI states, degraded lifecycle, deployment ownership, and close-out template |
| [V8.2 Operational Embodiment Directive](V8_2_OPERATIONAL_EMBODIMENT_DIRECTIVE.md) | deciding whether work should converge on visible operation rather than new doctrine | runs, outputs, proof, recovery, deployment trust, and confidence provenance preparation |
| [V8.2 Full Production Architecture](../../architecture/v8-2.md) | checking full B2+ target scope | Full Production Architecture (Canonical Target) for distributed execution, governed learning, deployment topology, and actuation scope |
| [V8 Runtime Contracts](V8_RUNTIME_CONTRACTS.md) | shaping runtime contracts | inception, kernel, council, provider policy, response classes, and continuity contracts |
| [V8 Config and Bootstrap Model](V8_CONFIG_AND_BOOTSTRAP_MODEL.md) | planning configuration sources, templates, bootstrap behavior, or inheritance | template vs instantiated organization, precedence rules, and V7-to-V8 migration truth |
| [V8 UI/API and Operator Experience Contract](V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md) | changing first-run flow, Soma workspace, advanced-mode boundaries, or screen/API mapping | canonical V8 operator PRD, source-of-truth layering, and UI/API contracts |
| [V8.2 Soma UI Architecture Expression](V8_2_SOMA_UI_ARCHITECTURE_EXPRESSION.md) | aligning Soma UI concept, component domains, active work, outputs, proof, teams, and scheduled work | ideal Soma operating loop, UI architecture domains, team execution lanes, and acceptance gates |
| [V8.2 Soma Team Interaction Contract](V8_2_SOMA_TEAM_INTERACTION_CONTRACT.md) | implementing or reviewing active team work, team steering, Council review, or team-output proof | canonical verbs, TeamInteraction, TeamWorkItem, TeamStatusEvent, TeamOutputRef, UI surfaces, recovery, and release gates |
| [V8 Directed Execution UI And Runtime Alignment Directive](V8_DIRECTED_EXECUTION_UI_RUNTIME_ALIGNMENT_DIRECTIVE.md) | reviewing whether UI/runtime feels like directed execution | current-state mapping, gaps, recommendations, execution plan, and validation standard |
| [V8 Directed Execution Delivery Plan](V8_DIRECTED_EXECUTION_DELIVERY_PLAN.md) | assigning directed-execution implementation work | team roster, work packages, waves, dependencies, validation matrix, and orchestration rules |
| [V8 Governed Execution Doctrine](V8_GOVERNED_EXECUTION_DOCTRINE.md) | checking governed-execution principles before implementation | accountable cognition, Event Spine truth, workspace visibility, and delivery compression |
| [V8 MVP Governed Execution Mission Plan](V8_MVP_GOVERNED_EXECUTION_MISSION_PLAN.md) | converting doctrine into executable MVP missions | mission scope, dependencies, governance implications, UI manifestation, events, and proof requirements |
| [V8.1 Living Organization Architecture](V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md) | checking foundation and compatibility baseline | loops, continuity, capabilities, bounded automations, and Soma-primary behavior |
| [V8 Compact Team Orchestration And Defaults](V8_COMPACT_TEAM_ORCHESTRATION_AND_DEFAULTS.md) | shaping team defaults or broad-ask splitting | compact team shape, minimal-team defaults, and Soma/Council coordination over NATS |
| [V8 Teamed Agentry Workflow Advantage](V8_TEAMED_AGENTRY_WORKFLOW_ADVANTAGE.md) | deciding when teams beat one strong agent | workflow variants and team-vs-single-agent boundaries |
| [V8 Workflow Variants And Reboot Proof Set](V8_WORKFLOW_VARIANTS_AND_REBOOT_PROOF_SET.md) | proving direct Soma, compact team, and multi-lane flows | comparative workflow proof and reboot-safe resume validation |
| [V8 Universal Soma And Context Model PRD](V8_UNIVERSAL_SOMA_AND_CONTEXT_MODEL_PRD.md) | working on Central Soma, context switching, scoped execution, or product-wide agentry standards | persistent Soma/Council identity, governed context model, and inspectable agency |
| [V8 Multi-User Identity And Soma Tenancy](V8_MULTI_USER_IDENTITY_AND_SOMA_TENANCY.md) | planning multi-user access or enterprise federation | SAML/SSO, break-glass admins, modular user management, and one-Soma tenancy |
| [V8.2 User Management And Enterprise Auth Module](V8_2_USER_MANAGEMENT_AND_ENTERPRISE_AUTH_MODULE.md) | defining auth providers, SSO, SAML, OIDC/OAuth, RBAC, or SCIM posture | pluggable authentication providers, Mycelis-owned authorization, approval authority, and enterprise validation |
| [V8 Memory Layer And Reflection Delivery Contract](V8_MEMORY_LAYER_AND_REFLECTION_DELIVERY_CONTRACT.md) | planning memory layers, reflection synthesis, or memory promotion | explicit memory lanes, candidate-first reflection, and promotion guardrails |
| [V8 Trusted Memory Arbitration And Team Vector Contract](V8_TRUSTED_MEMORY_ARBITRATION_AND_TEAM_VECTOR_CONTRACT.md) | defining trusted recall or memory arbitration | evidence anchors, precedence rules, trusted memory control plane, and bounded growth |
| [V8 Home Docker Compose Runtime](V8_HOME_DOCKER_COMPOSE_RUNTIME.md) | supporting the single-host Compose runtime | compose tasking, persistence, monitoring, and operator expectations |
| [V8 Self-Hosted Runtime Delivery Program](V8_SELF_HOSTED_RUNTIME_DELIVERY_PROGRAM.md) | coordinating release proof for self-hosted runtime | Compose proof, modular Kubernetes scale-up, ownership, cadence, and acceptance gates |
| [V8 Mycelis Search Capability Delivery Plan](V8_MYCELIS_SEARCH_CAPABILITY_DELIVERY_PLAN.md) | delivering owned search capability | local-source search, operator-owned local API search, optional SearXNG/Brave, and testing gates |
| [V8 Capability Manifest And Runtime Integration Standard](V8_CAPABILITY_MANIFEST_AND_RUNTIME_INTEGRATION_STANDARD.md) | adding or reviewing tools, APIs, scripts, plugins, or MCP integrations | capability manifests, permissions, run proof, output normalization, governance, and recovery |
| [V8 Enterprise Self-Hosted Kubernetes Delivery Plan](V8_ENTERPRISE_SELF_HOSTED_KUBERNETES_DELIVERY_PLAN.md) | turning Kubernetes into enterprise-compatible delivery | Rancher K3s, k3d validation, chart promotion, and acceptance gates |
| [V8 Compose Personal Owner Deployment Test Plan](V8_COMPOSE_PERSONAL_OWNER_DEPLOYMENT_TEST_PLAN.md) | validating personal-owner Compose deployments | data-plane startup, owner configuration, workflow proof, recovery, and evidence matrix |
| [V8 MVP Media, Team Output, And Template Registry](V8_MVP_MEDIA_TEAM_OUTPUT_AND_TEMPLATE_REGISTRY.md) | planning media/team-output demos or template registry work | user-output-first media proof, team-managed outputs, model-role routing, and conversation templates |
| [V8 UI Testing Agentry Product Contract](V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md) | defining browser validation expectations | Soma-first proof for answers, governed actions, continuity, and audit visibility |
| [V8 UI Team Full Test Set](V8_UI_TEAM_FULL_TEST_SET.md) | running the complete UI validation set | browser workflows, runtime proof, evidence packaging, and verdict rules |
| [Source File Size And Indexing Plan](SOURCE_FILE_SIZE_AND_INDEXING_PLAN.md) | reducing oversized source, test, docs, chart, or config files | 300-line ratchet, indexing rules, and decomposition priority |

## How To Use This Library

1. Start with [V8.2 Current State And Finalization PRD](V8_2_CURRENT_STATE_AND_FINALIZATION_PRD.md) when deciding what the architecture team should execute next.
2. Use [V8.2 Operational Embodiment Directive](V8_2_OPERATIONAL_EMBODIMENT_DIRECTIVE.md) to reject doctrine-only expansion that does not improve visible execution, proof, recovery, or trust.
3. Use [V8 Runtime Contracts](V8_RUNTIME_CONTRACTS.md), [V8 UI/API and Operator Experience Contract](V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md), and [V8 Capability Manifest And Runtime Integration Standard](V8_CAPABILITY_MANIFEST_AND_RUNTIME_INTEGRATION_STANDARD.md) as the default runtime/UI/capability trio.
4. Use [V8 Directed Execution Delivery Plan](V8_DIRECTED_EXECUTION_DELIVERY_PLAN.md) when assigning teams, waves, and validation gates.
5. Use [V8 UI Testing Agentry Product Contract](V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md) and [V8 UI Team Full Test Set](V8_UI_TEAM_FULL_TEST_SET.md) before accepting operator-facing slices.
6. Update [.state/V8_DEV_STATE.md](../../.state/V8_DEV_STATE.md) whenever implementation truth, validation evidence, or release posture changes.

Execution governance reminder:
- name the module boundary being advanced before implementation begins: runtime/deployment, memory/learning, team/workflow, capability/MCP, advanced UI, or governance/trust
- promote missing historical concepts into current V8.2 docs instead of restoring superseded V7 topical documents
- keep product-standard slices attached to browser proof, durable outputs, audit/proof visibility, and recovery behavior

## Supporting Specialized Docs

This library sits above specialized implementation references. Use these when the work is narrower than the canonical architecture set:

- [Overview](../architecture/OVERVIEW.md)
- [Backend](../architecture/BACKEND.md)
- [Frontend](../architecture/FRONTEND.md)
- [Operations](../architecture/OPERATIONS.md)

## Library Principles

- one durable concept per document
- operator outcome before internal complexity
- no temporary planning notes in the canonical library
- no stale archive drafts or superseded V7 topical docs as current authority
- README links only the few docs needed to orient new work
- tests, docs, and the in-app docs manifest must agree on the same authority set

## Canonical Outcome

The intended product is a governed execution system where:
- intent becomes a direct answer, a governed proposal, an execution result, or a blocker
- workflows can be one-shot, scheduled, persistent active, or event-driven
- every action preserves lineage across UI, API, NATS, persistence, proof, and recovery
- operators can understand what is happening without drowning in telemetry

Next:
- [V8.2 Current State And Finalization PRD](V8_2_CURRENT_STATE_AND_FINALIZATION_PRD.md)
