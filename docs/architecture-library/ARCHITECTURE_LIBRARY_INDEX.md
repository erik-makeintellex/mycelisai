# Architecture Library Index

> Status: Canonical
> Last Updated: 2026-04-04
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

## Layered Truth

- [README](../../README.md) is the primary inception document for active development.
- [V8.1 Living Organization Architecture](V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md) is the current release architecture.
- [V8.2 Full Production Architecture](../../v8-2.md) is the canonical full production architecture and full actuation target.
- [V8 Development State](../../V8_DEV_STATE.md) is the live implementation scoreboard.

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
| [V8 Runtime Contracts](V8_RUNTIME_CONTRACTS.md) | defining the new V8 runtime contract surface during migration | contract shells for inception, kernel, council, provider policy, and continuity state |
| [V8 Config and Bootstrap Model](V8_CONFIG_AND_BOOTSTRAP_MODEL.md) | planning how V8 concepts enter the system through config and bootstrap behavior | configuration sources, templates, instantiation entry points, bootstrap resolution, scope inheritance, precedence rules, and the V7-to-V8 bootstrap migration contract |
| [V8 UI/API and Operator Experience Contract](V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md) | working on first-run flow, AI Organization creation, Soma-primary workspace, advanced architecture/runtime boundaries, or screen-to-API mapping | canonical V8 operator PRD, anti-generic-chat UX guardrails, default-vs-advanced surface rules, managed exchange visibility boundaries, security-label inspection boundaries, source-of-truth layering, and API/UI contracts |
| [V8.1 Living Organization Architecture](V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md) | defining the current bounded release architecture for loop profiles, learning loops, semantic continuity, procedure/skill memory, managed exchange, runtime capabilities, promoted response/style inheritance, or Automations visibility | Current Release Architecture for persistent execution, policy-bounded automation, semantic continuity, managed exchange foundations, exchange security foundations, Agent Type Profile runtime truth, and the first bounded Automations posture |
| [V8 UI Testing Agentry Product Contract](V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md) | defining what the UI testing agentry must prove before a workspace or organization change is ready for review | canonical Soma-first browser proof contract for direct answers, governed proposals, approval/cancel behavior, continuity, audit visibility, and operator trust |
| [V8 UI Testing Agentry Execution Runbook](V8_UI_TESTING_AGENTRY_EXECUTION_RUNBOOK.md) | running the UI testing agentry step by step and returning evidence in a reliably triageable format | operational runbook for environment prep, lane order, workflow prompts, expected terminal states, evidence capture, and failure classification |
| [V8 UI Workflow Verification Plan](V8_UI_WORKFLOW_VERIFICATION_PLAN.md) | coordinating the exact workflow-by-workflow release verification pass for the UI testing team | prioritized browser verification plan with expected routes, API calls, terminal states, and evidence requirements |
| [V8 UI Team Full Test Set](V8_UI_TEAM_FULL_TEST_SET.md) | running the complete UI team validation set across Central Soma, AI Organization flows, settings persistence, governed execution, and deployed runtime proof | canonical full test matrix for browser workflows, compose/kind environments, evidence packaging, and final verdict rules |
| [V8 UI Testing Stabilization Strike Team Plan](V8_UI_TESTING_STABILIZATION_STRIKE_TEAM_PLAN.md) | coordinating mocked browser proof, live governed-chat repair, and UI testing release hygiene | active strike-team ownership plan for the current UI-testing stabilization effort |
| [V8 Memory Continuity And RAG Strike Team Plan](V8_MEMORY_CONTINUITY_AND_RAG_STRIKE_TEAM_PLAN.md) | coordinating team-scoped vector recall, temporary planning continuity, and trace-clean memory boundaries | canonical strike-team plan for durable pgvector memory, temporary continuity channels, trace separation, ownership, and release proof |
| [V8 Content Generation And Collaboration Strike Team Plan](V8_CONTENT_GENERATION_AND_COLLABORATION_STRIKE_TEAM_PLAN.md) | coordinating inline content delivery, governed artifact generation, and policy-configurable specialist/model collaboration | canonical strike-team plan for answer-vs-proposal behavior, generated artifact proof, collaboration posture, and content-value delivery |
| [V8 Approval and Product Trust Strike Team Plan](V8_APPROVAL_AND_PRODUCT_TRUST_STRIKE_TEAM_PLAN.md) | coordinating approval/auth-style UX simplification, content-value delivery, and governed product trust without removing advanced depth | canonical strike-team plan for user-first proposal surfaces, details-on-demand execution metadata, content/artifact result return, and release acceptance gates |
| [V8 Universal Soma And Context Model PRD](V8_UNIVERSAL_SOMA_AND_CONTEXT_MODEL_PRD.md) | correcting the product model so Soma and Council stay universal while organizations, deployments, teams, and logs remain governed contexts | canonical PRD for Central Soma, context switching, scoped execution, preserved advanced depth, and the transition away from “one Soma per organization” |
| [V8 Home Docker Compose Runtime](V8_HOME_DOCKER_COMPOSE_RUNTIME.md) | supporting the single-host Docker Compose runtime for home-lab, demo, and partner-review use | canonical contract for env separation, compose tasking, persistence, monitoring, and operator expectations without Kind |
| [V8 True MVP Finish Plan](V8_TRUE_MVP_FINISH_PLAN.md) | coordinating the final cross-lane push from release-candidate posture to a true MVP | prioritized finish plan for product value delivery, reliability, settings completion, demo readiness, and release proof |
| [V8 Demo Product Strike Team Plan](V8_DEMO_PRODUCT_STRIKE_TEAM_PLAN.md) | coordinating product simplification, demo legibility, feature preservation, and partner-ready workflow proof | canonical strike-team plan for making Mycelis read as an obvious product without removing Soma capability or advanced power |
| [V8 Demo Product Execution Brief](V8_DEMO_PRODUCT_EXECUTION_BRIEF.md) | executing the first engaged-team deliverables for default/advanced surface inventory, feature preservation, golden-path demo, and UI verification | canonical execution brief for turning the strike-team plan into immediate product-facing outputs |
| [V8 Demo Product Wording Drift Inventory](V8_DEMO_PRODUCT_WORDING_DRIFT_INVENTORY.md) | reviewing README, landing, organization setup, organization home, and docs for product-language drift | canonical wording audit for `Memory & Continuity`, retained-knowledge language, and product-legibility cleanup priorities |
| [V8 Demo Product Feature Retention Map](V8_DEMO_PRODUCT_FEATURE_RETENTION_MAP.md) | proving that advanced power remains reachable while the default product story becomes simpler | canonical retained-home map for advanced routes, governed chat power, runtime depth, and partner-demo simplification boundaries |
| [V8 MVP Investor Demo and Governed Capability Plan](V8_MVP_INVESTOR_DEMO_AND_GOVERNED_CAPABILITY_PLAN.md) | preparing the investor-ready workflow story across Soma, governed execution, MCP expansion, web research, and context-security controls | canonical investor demo plan for product value, connected tools, securable external capability, and proof requirements |
| [V8 Partner Demo Script](V8_PARTNER_DEMO_SCRIPT.md) | running the canonical partner/funder demo without flattening the platform into a narrow showcase | canonical demo sequence for product value, governed execution, continuity, and optional advanced reveal |
| [V8 Partner Demo Verification Checklist](V8_PARTNER_DEMO_VERIFICATION_CHECKLIST.md) | verifying that the partner demo actually proves product value, trust, structure, and retained depth | product-focused QA checklist for the demo lane, distinct from the broader release verification plan |
| [V8.2 Cross-Repo Cleanup and Release Structure Plan](V8_2_CROSS_REPO_CLEANUP_AND_RELEASE_STRUCTURE_PLAN.md) | coordinating git hygiene, commit-lane packaging, release cleanup, cross-team ownership, and final validation order | prioritized cleanup lanes, commit boundaries, owner assignments, validation gates, and release-structure rules for the current mixed worktree |
| [V8 Release Platform Review: Security, Monitoring, and Debug](V8_RELEASE_PLATFORM_REVIEW_SECURITY_MONITORING_DEBUG.md) | aligning release truth across governance/security, monitoring/ops, and debug/live-browser proof | canonical shared review surface for platform release readiness, operator checklist, residual risks, and matching-doc requirements |
| [V8 Dev Team Execution and UI Engagement Plan](V8_DEV_TEAM_EXECUTION_AND_UI_ENGAGEMENT_PLAN.md) | coordinating the dev team around true-MVP execution order, ownership, UI testing expectations, and engagement-testing expectations | canonical cross-team execution plan for active dev lanes, committed-slice discipline, UI/browser proof, and partner-demo engagement validation |
| [V8 Ask-Class, Agent-Type, and Output Contract Plan](V8_ASK_CLASS_AGENT_TYPE_OUTPUT_CONTRACT_PLAN.md) | coordinating one shared machine-readable contract for ask classification, specialist targeting, and output-mode assertion | canonical plan for consolidating ask->agent->output truth across runtime, UI, docs, and tests without weakening governed execution |
| [V8 User Workflow Execution and Verification Plan](V8_USER_WORKFLOW_EXECUTION_AND_VERIFICATION_PLAN.md) | coordinating the team around real user workflows, user-shaped delivery structure, MCP usability, and workflow-complete proof | canonical execution plan for outcome-based workflow testing, neutral trust language, user-defined team shaping, and engine-governed rather than intent-governed product behavior |
| [V8 User Intent Workflow Review and Instantiation Plan](V8_USER_INTENT_WORKFLOW_REVIEW_AND_INSTANTIATION_PLAN.md) | coordinating one canonical review and proof contract for direct intent handling, team instantiation, and external workflow-contract instantiation | canonical plan for separating native Mycelis team manifestation from external workflow services such as `n8n`, while requiring visible target-output proof for both |
| [V8 Full Testing Action Plan](V8_FULL_TESTING_ACTION_PLAN.md) | running one release-style test pass across repo baseline, stable browser proof, live governed-browser proof, compose-aware runtime proof, and final doc/state sync | canonical ordered full-testing runbook for release review, agent execution, and workflow-complete evidence capture |
| [V8.2 Full Production Architecture](../../v8-2.md) | defining the full distributed, learning, capability-enabled, managed-exchange, and actuation target beyond the current release | Full Production Architecture (Canonical Target) for distributed execution, governed learning, managed exchange, permissioned trust boundaries, capability-backed execution, and full actuation scope |

### V8 migration reminder

- V7 docs in this table stay authoritative migration inputs until replaced, but they do not override the V8 bootstrap contract.
- `docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md` is the canonical V7->V8 migration and bootstrap source; every YAML/runtime/DB/operator asset must be translated through it before shaping a running organization.
- `docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md` extends the V8 contract library with Loop Profiles, Learning Loops, Runtime Capabilities, semantic continuity, Procedure / Skill Sets, promoted Response Contract inheritance, promoted Agent Type Profiles, and the first bounded Automations operator surface.
- `../../v8-2.md` is the canonical full production architecture and full actuation target; it is not the current release contract.
- `Template ≠ instantiated organization`; templates remain reusable blueprints while instantiated Inceptions inherit, override, and resolve configuration through the V8 flow.

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
12. Use [V8 Runtime Contracts](V8_RUNTIME_CONTRACTS.md) when defining the new canonical V8 runtime contracts during migration.
13. Use [V8 Config and Bootstrap Model](V8_CONFIG_AND_BOOTSTRAP_MODEL.md) when planning how V8 contracts enter the system through configuration, templates, organization entry points, bootstrap resolution, and precedence rules.
14. Use [V8 UI/API and Operator Experience Contract](V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md) when defining first-run flow, AI Organization creation, Soma-primary workspace behavior, role visibility, advanced-mode boundaries, source-of-truth layering, and screen-to-API mapping.
15. Use [V8.1 Living Organization Architecture](V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md) when defining the current bounded release architecture for Loop Profiles, Learning Loops, semantic continuity, managed exchange foundations, Procedure / Skill Sets, Runtime Capabilities, Agent Type runtime truth, promoted Response Style inheritance, and the read-only Automations surface.
16. Use [V8 UI Testing Agentry Product Contract](V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md) when defining what the browser/testing agentry must prove for Soma-first entry, governed actions, continuity, audit visibility, and trust recovery.
17. Use [V8 UI Testing Agentry Execution Runbook](V8_UI_TESTING_AGENTRY_EXECUTION_RUNBOOK.md) when the testing agentry needs exact step-by-step run order, prompts, evidence handling, result formatting, and failure classification rules.
18. Use [V8 UI Workflow Verification Plan](V8_UI_WORKFLOW_VERIFICATION_PLAN.md) when executing the release verification pass across AI Organization entry, Soma direct-answer flow, governed mutation, approvals/audit, settings changes, and route reachability.
19. Use [V8 UI Team Full Test Set](V8_UI_TEAM_FULL_TEST_SET.md) when the UI team needs one complete test matrix covering Central Soma, AI Organization creation, settings persistence, governed execution, advanced boundaries, and deployed runtime proof.
20. Use [V8 UI Testing Stabilization Strike Team Plan](V8_UI_TESTING_STABILIZATION_STRIKE_TEAM_PLAN.md) when coordinating the active mocked-browser, live-governed-chat, and release-hygiene UI stabilization effort.
21. Use [V8 Memory Continuity And RAG Strike Team Plan](V8_MEMORY_CONTINUITY_AND_RAG_STRIKE_TEAM_PLAN.md) when coordinating team-scoped vector recall, temporary planning continuity, trace-clean memory boundaries, and release proof for semantic continuity changes.
22. Use [V8 Content Generation And Collaboration Strike Team Plan](V8_CONTENT_GENERATION_AND_COLLABORATION_STRIKE_TEAM_PLAN.md) when coordinating content delivery posture, generated artifact proof, media/file generation behavior, and policy-configurable specialist/model collaboration.
23. Use [V8 Approval and Product Trust Strike Team Plan](V8_APPROVAL_AND_PRODUCT_TRUST_STRIKE_TEAM_PLAN.md) when coordinating user-first proposal surfaces, content/artifact value delivery, details-on-demand governance metadata, and release acceptance for approval UX changes.
24. Use [V8 Universal Soma And Context Model PRD](V8_UNIVERSAL_SOMA_AND_CONTEXT_MODEL_PRD.md) when correcting product identity around one persistent Soma/Council pair, Central Soma home, context switching, scoped execution, and the transition away from org-scoped Soma language.
25. Use [V8 Home Docker Compose Runtime](V8_HOME_DOCKER_COMPOSE_RUNTIME.md) when supporting the single-host Docker Compose runtime for home-lab, demo, and partner-review use without losing observability or compose-task discipline.
26. Use [V8 True MVP Finish Plan](V8_TRUE_MVP_FINISH_PLAN.md) when coordinating the final MVP push across product value delivery, settings completion, demo readiness, and release proof.
27. Use [V8 Demo Product Strike Team Plan](V8_DEMO_PRODUCT_STRIKE_TEAM_PLAN.md) when coordinating product simplification, demo readiness, feature preservation, workflow proof, and partner/funder-facing legibility.
28. Use [V8 Demo Product Execution Brief](V8_DEMO_PRODUCT_EXECUTION_BRIEF.md) when executing the first engaged-team deliverables for default-vs-advanced surfaces, feature preservation, the golden-path partner demo, and the product-focused UI verification checklist.
29. Use [V8 Demo Product Wording Drift Inventory](V8_DEMO_PRODUCT_WORDING_DRIFT_INVENTORY.md) when aligning README, landing, organization setup, organization home, and retention/governance language to the current product story without blindly renaming deep internal contracts.
30. Use [V8 Demo Product Feature Retention Map](V8_DEMO_PRODUCT_FEATURE_RETENTION_MAP.md) when proving that advanced routes, governed mutation depth, memory inspection, system diagnostics, and broadcast/specialist power remain reachable while the default demo story stays simple.
31. Use [V8 MVP Investor Demo and Governed Capability Plan](V8_MVP_INVESTOR_DEMO_AND_GOVERNED_CAPABILITY_PLAN.md) when preparing an investor-ready story across Soma’s initial value, governed mutation, MCP-powered input/output, securable web research, and inspectable context-security controls.
32. Use [V8 Partner Demo Script](V8_PARTNER_DEMO_SCRIPT.md) when preparing the canonical technical-partner/funder demonstration and keeping the story centered on product value rather than infrastructure depth.
33. Use [V8 Partner Demo Verification Checklist](V8_PARTNER_DEMO_VERIFICATION_CHECKLIST.md) when the UI testing and release teams need to prove that the partner demo shows product value, governed trust, structure, continuity, preserved advanced depth, and governable external capability.
34. Use [V8.2 Cross-Repo Cleanup and Release Structure Plan](V8_2_CROSS_REPO_CLEANUP_AND_RELEASE_STRUCTURE_PLAN.md) when coordinating git hygiene, commit boundaries, cleanup lanes, release structure, and final validation order across a mixed worktree.
35. Use [V8 Release Platform Review: Security, Monitoring, and Debug](V8_RELEASE_PLATFORM_REVIEW_SECURITY_MONITORING_DEBUG.md) when aligning release truth across security/governance, monitoring/ops, and debug/live-browser proof.
36. Use [V8 Dev Team Execution and UI Engagement Plan](V8_DEV_TEAM_EXECUTION_AND_UI_ENGAGEMENT_PLAN.md) when coordinating cross-team execution order, dev-lane ownership, UI/browser proof, engagement testing, and committed-slice discipline for the active MVP push.
37. Use [V8 Ask-Class, Agent-Type, and Output Contract Plan](V8_ASK_CLASS_AGENT_TYPE_OUTPUT_CONTRACT_PLAN.md) when consolidating ask classification, specialist targeting, runtime output assertion, and shared test proof across Soma, council, and governed execution surfaces.
38. Use [V8 User Workflow Execution and Verification Plan](V8_USER_WORKFLOW_EXECUTION_AND_VERIFICATION_PLAN.md) when validating that Mycelis supports user-shaped delivery, standards-compliant MCP usability, and workflow-complete operator outcomes rather than only interaction-level success.
39. Use [V8 User Intent Workflow Review and Instantiation Plan](V8_USER_INTENT_WORKFLOW_REVIEW_AND_INSTANTIATION_PLAN.md) when deciding whether an ask should stay direct, instantiate a managed team, or instantiate an external workflow contract such as `n8n`, and when defining proof for target-output delivery such as bounded image generation.
40. Use [V8 Full Testing Action Plan](V8_FULL_TESTING_ACTION_PLAN.md) when you need one ordered release-style run across repo baseline, live service proof, compose-aware runtime proof, focused workflow suites, and final docs/state verification.
41. Use [V8.2 Full Production Architecture](../../v8-2.md) when checking whether a surface belongs to the canonical full production, managed exchange, and actuation target rather than the current release.

Execution governance reminder:
- `NEXT_EXECUTION_SLICES_V7.md`, `DELIVERY_GOVERNANCE_AND_TESTING_V7.md`, and `TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md` now describe V8 migration slices; apply the V8 bootstrap pipeline (`docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`) and update `V8_DEV_STATE.md` whenever those docs change.

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
