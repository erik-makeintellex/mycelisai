/**
 * Docs Manifest
 *
 * Curated registry of documentation files served by the in-app doc browser.
 * Every entry maps a URL-safe slug to a filesystem path (relative to the
 * project root, one level above the `interface/` directory).
 *
 * Security: the API route validates slugs against this manifest before
 * touching the filesystem — no arbitrary path traversal is possible.
 *
 * Adding a new doc:
 *   1. Add a DocEntry to the appropriate DocSection below.
 *   2. The slug becomes the URL: /docs?doc={slug}
 *   3. The path is relative to the scratch/ project root.
 */

export interface DocEntry {
    slug: string;
    label: string;
    /** Path relative to scratch/ project root (e.g. "README.md", "docs/API_REFERENCE.md") */
    path: string;
    /** Short one-line description shown in the sidebar tooltip */
    description?: string;
}

export interface DocSection {
    section: string;
    docs: DocEntry[];
}

export const DOC_MANIFEST: DocSection[] = [
    // ── User Guides ───────────────────────────────────────────────────────────
    {
        section: "User Guides",
        docs: [
            {
                slug: "core-concepts",
                label: "Core Concepts",
                path: "docs/user/core-concepts.md",
                description: "Soma, Council, Mission, Run, Brain, Event, Trust — plain-language glossary",
            },
            {
                slug: "system-status-recovery",
                label: "System Status & Recovery",
                path: "docs/user/system-status-recovery.md",
                description: "Global health signals, degraded recovery actions, and Quick Checks workflow",
            },
            {
                slug: "soma-chat",
                label: "Using Soma Chat",
                path: "docs/user/soma-chat.md",
                description: "Step-by-step: send messages, read delegation traces, confirm proposals",
            },
            {
                slug: "meta-agent-blueprint",
                label: "Meta-Agent & Blueprints",
                path: "docs/user/meta-agent-blueprint.md",
                description: "How Architect generates mission blueprints — teams, agents, tools, I/O contracts",
            },
            {
                slug: "run-timeline",
                label: "Run Timeline",
                path: "docs/user/run-timeline.md",
                description: "Reading execution timelines — event types, status, navigation",
            },
            {
                slug: "automations-guide",
                label: "Automations",
                path: "docs/user/automations.md",
                description: "Active automations, trigger rules, approvals, and advanced workflow tools",
            },
            {
                slug: "resources-guide",
                label: "Resources",
                path: "docs/user/resources.md",
                description: "Advanced tools, workspace files, AI engines, and reusable role definitions",
            },
            {
                slug: "memory-guide",
                label: "Memory",
                path: "docs/user/memory.md",
                description: "Semantic search, SitReps, artifacts, hot/warm/cold tiers",
            },
            {
                slug: "governance-trust",
                label: "Governance & Trust",
                path: "docs/user/governance-trust.md",
                description: "Trust scores, approval flows, policy rules, propose vs execute modes",
            },
        ],
    },

    // ── Getting Started ───────────────────────────────────────────────────────
    {
        section: "Getting Started",
        docs: [
            {
                slug: "readme",
                label: "Overview",
                path: "README.md",
                description: "Architecture, stack, commands, and current phase",
            },
            {
                slug: "local-dev",
                label: "Local Dev Workflow",
                path: "docs/LOCAL_DEV_WORKFLOW.md",
                description: "Setup, config reference, port map, troubleshooting",
            },
            {
                slug: "v8-dev-state",
                label: "V8 Dev State",
                path: "V8_DEV_STATE.md",
                description: "Live implementation scoreboard for current V8 delivery status, blockers, and evidence",
            },
            {
                slug: "v8-ui-workflow-verification",
                label: "V8 UI Workflow Verification",
                path: "docs/architecture-library/V8_UI_WORKFLOW_VERIFICATION_PLAN.md",
                description: "Release verification plan for the UI testing team, including workflows, expected API calls, and terminal states",
            },
            {
                slug: "v8-ui-testing-agentry-runbook",
                label: "V8 UI Testing Agentry Runbook",
                path: "docs/architecture-library/V8_UI_TESTING_AGENTRY_EXECUTION_RUNBOOK.md",
                description: "Step-by-step execution runbook for the UI testing agentry, including lane order, prompts, evidence packaging, and triage rules",
            },
            {
                slug: "v8-memory-continuity-rag-plan",
                label: "V8 Memory Continuity & RAG Plan",
                path: "docs/architecture-library/V8_MEMORY_CONTINUITY_AND_RAG_STRIKE_TEAM_PLAN.md",
                description: "Strike-team plan for scoped pgvector memory, temporary planning continuity, and trace-clean memory boundaries",
            },
            {
                slug: "v8-content-generation-collaboration-plan",
                label: "V8 Content Generation & Collaboration Plan",
                path: "docs/architecture-library/V8_CONTENT_GENERATION_AND_COLLABORATION_STRIKE_TEAM_PLAN.md",
                description: "Strike-team plan for inline content delivery, governed artifact generation, and policy-configurable specialist/model collaboration",
            },
            {
                slug: "v8-home-docker-compose-runtime",
                label: "V8 Home Docker Compose Runtime",
                path: "docs/architecture-library/V8_HOME_DOCKER_COMPOSE_RUNTIME.md",
                description: "Canonical single-host Docker Compose runtime for home-lab, demo, and partner-review use with managed env, health, and logging expectations",
            },
            {
                slug: "v8-true-mvp-finish-plan",
                label: "V8 True MVP Finish Plan",
                path: "docs/architecture-library/V8_TRUE_MVP_FINISH_PLAN.md",
                description: "Prioritized finish plan for getting Mycelis from release-candidate posture to a true MVP",
            },
            {
                slug: "v8-demo-product-plan",
                label: "V8 Demo Product Plan",
                path: "docs/architecture-library/V8_DEMO_PRODUCT_STRIKE_TEAM_PLAN.md",
                description: "Strike-team plan for making Mycelis feel like an obvious product for partner demos without removing advanced power",
            },
            {
                slug: "v8-demo-product-execution-brief",
                label: "V8 Demo Product Execution Brief",
                path: "docs/architecture-library/V8_DEMO_PRODUCT_EXECUTION_BRIEF.md",
                description: "First engaged-team deliverables for default surfaces, advanced preservation, golden-path demo, and UI verification",
            },
            {
                slug: "v8-demo-product-wording-drift",
                label: "V8 Demo Product Wording Drift",
                path: "docs/architecture-library/V8_DEMO_PRODUCT_WORDING_DRIFT_INVENTORY.md",
                description: "Prioritized wording-drift inventory for README, landing, organization entry, organization home, and continuity language",
            },
            {
                slug: "v8-demo-product-feature-retention",
                label: "V8 Demo Product Feature Retention",
                path: "docs/architecture-library/V8_DEMO_PRODUCT_FEATURE_RETENTION_MAP.md",
                description: "Retention map proving advanced power remains reachable while the default product story stays simple",
            },
            {
                slug: "v8-partner-demo-script",
                label: "V8 Partner Demo Script",
                path: "docs/architecture-library/V8_PARTNER_DEMO_SCRIPT.md",
                description: "Canonical partner/funder demo story proving product value, governance, continuity, and retained platform depth",
            },
            {
                slug: "v8-partner-demo-verification",
                label: "V8 Partner Demo Verification",
                path: "docs/architecture-library/V8_PARTNER_DEMO_VERIFICATION_CHECKLIST.md",
                description: "UI testing and release checklist for validating the partner demo as a product story rather than a narrow technical path",
            },
            {
                slug: "v7-dev-state",
                label: "V7 Dev State (Historical)",
                path: "V7_DEV_STATE.md",
                description: "Historical migration checkpoint retained as V8 input, not the live implementation scoreboard",
            },
        ],
    },

    // ── Soma Workflow Reference ───────────────────────────────────────────────
    {
        section: "Soma Workflow",
        docs: [
            {
                slug: "workflows",
                label: "Workflow Spec",
                path: "docs/WORKFLOWS.md",
                description: "10 user-facing workflow specifications (implementation authority)",
            },
            {
                slug: "swarm-operations",
                label: "Swarm Operations",
                path: "docs/SWARM_OPERATIONS.md",
                description: "Hierarchy, blueprints, activation, teams, tools, governance",
            },
            {
                slug: "council-chat-qa",
                label: "Council Chat QA",
                path: "docs/QA_COUNCIL_CHAT_API.md",
                description: "QA procedures and test cases for the council chat API",
            },
        ],
    },

    // ── Archive (Historical / Non-authoritative) ────────────────────────────
    {
        section: "Archive",
        docs: [
            {
                slug: "archive-index",
                label: "Archive Index",
                path: "docs/archive/README.md",
                description: "Historical documents only — not implementation authority",
            },
            {
                slug: "mvp-agentry",
                label: "MVP Agentry Plan (Archive)",
                path: "docs/archive/MVP_AGENTRY_PLAN.md",
                description: "Historical agentry chain map retained for context",
            },
            {
                slug: "v7-ui-verification",
                label: "V7 UI Verification (Archive)",
                path: "docs/archive/v7-step-01-ui.md",
                description: "Historical manual UI verification checklist for V7 Step 01 navigation",
            },
            {
                slug: "v7-ia-step01",
                label: "IA Step 01 (Archive)",
                path: "docs/archive/ia-v7-step-01.md",
                description: "Historical workflow-first navigation PRD for Step 01 implementation",
            },
            {
                slug: "v7-ui-optimal-workflow-prds",
                label: "UI Optimal Workflow PRDs (Archive)",
                path: "docs/archive/UI_OPTIMAL_WORKFLOW_PRDS_V7.md",
                description: "Historical planning PRDs superseded by current execution authority docs",
            },
            {
                slug: "v7-ui-engagement-actuation-review",
                label: "UI Engagement & Actuation Review (Archive)",
                path: "docs/archive/UI_OPTIMAL_ENGAGEMENT_ACTUATION_REVIEW_V7.md",
                description: "Historical UI engagement/actuation review retained for context",
            },
        ],
    },

    // ── API Reference ─────────────────────────────────────────────────────────
    {
        section: "API Reference",
        docs: [
            {
                slug: "api-reference",
                label: "API Reference",
                path: "docs/API_REFERENCE.md",
                description: "Full endpoint table — 80+ routes with request/response shapes",
            },
            {
                slug: "cognitive-architecture",
                label: "Cognitive Architecture",
                path: "docs/COGNITIVE_ARCHITECTURE.md",
                description: "Providers, profiles, matrix UI, embedding",
            },
            {
                slug: "logging-schema",
                label: "Logging Standard (V7)",
                path: "docs/logging.md",
                description: "Authoritative mission-events and memory-stream logging contract, taxonomy, and onboarding checklist",
            },
        ],
    },

    // ── Architecture ──────────────────────────────────────────────────────────
    {
        section: "Architecture",
        docs: [
            {
                slug: "architecture-library-index",
                label: "Architecture Library Index",
                path: "docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md",
                description: "Canonical modular index for target delivery, architecture, execution, UI, and testing guidance",
            },
            {
                slug: "target-deliverable-v7",
                label: "Target Deliverable V7",
                path: "docs/architecture-library/TARGET_DELIVERABLE_V7.md",
                description: "Full product end state, recurring-plan modes, phase framing, and success criteria",
            },
            {
                slug: "system-architecture-v7",
                label: "System Architecture V7",
                path: "docs/architecture-library/SYSTEM_ARCHITECTURE_V7.md",
                description: "Canonical runtime layers, persistence model, deployment posture, and bus/storage rules",
            },
            {
                slug: "execution-manifest-library-v7",
                label: "Execution And Manifest Library V7",
                path: "docs/architecture-library/EXECUTION_AND_MANIFEST_LIBRARY_V7.md",
                description: "Run lifecycle, manifest lifecycle, recurring-plan semantics, and activation rules",
            },
            {
                slug: "intent-manifestation-team-interaction-v7",
                label: "Intent To Manifestation + Team Interaction V7",
                path: "docs/architecture-library/INTENT_TO_MANIFESTATION_AND_TEAM_INTERACTION_V7.md",
                description: "Canonical intent-to-manifestation flow, module abstraction, and created-team interaction model",
            },
            {
                slug: "ui-operator-experience-v7",
                label: "UI And Operator Experience V7",
                path: "docs/architecture-library/UI_AND_OPERATOR_EXPERIENCE_V7.md",
                description: "Operator journeys, anti-information-swarm design rules, and intuitive UI targets",
            },
            {
                slug: "delivery-governance-testing-v7",
                label: "Delivery Governance And Testing V7",
                path: "docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md",
                description: "Delivery proof model, evidence requirements, and product-aligned testing expectations",
            },
            {
                slug: "next-execution-slices-v7",
                label: "Next Execution Slices V7",
                path: "docs/architecture-library/NEXT_EXECUTION_SLICES_V7.md",
                description: "Current working queue with scoped next slices and linked development/testing references",
            },
            {
                slug: "team-execution-global-state-protocol-v7",
                label: "Team Execution + Global State Protocol V7",
                path: "docs/architecture-library/TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md",
                description: "Multi-lane execution architecture, global-state maintenance rules, and deep-testing obligations",
            },
            {
                slug: "mvp-release-strike-team-plan-v7",
                label: "MVP Release Strike Team Plan V7",
                path: "docs/architecture-library/MVP_RELEASE_STRIKE_TEAM_PLAN_V7.md",
                description: "Active MVP lane ownership, communication cadence, and state-file update discipline",
            },
            {
                slug: "mvp-integration-toolship-execution-plan-v7",
                label: "MVP Integration + Toolship Plan V7",
                path: "docs/architecture-library/MVP_INTEGRATION_AND_TOOLSHIP_EXECUTION_PLAN_V7.md",
                description: "Canonical plan for AI interaction, internal toolship, and service-connection hardening to reach non-test-team MVP",
            },
            {
                slug: "ui-generation-testing-execution-plan-v7",
                label: "UI Generation + Testing Plan V7",
                path: "docs/architecture-library/UI_GENERATION_AND_TESTING_EXECUTION_PLAN_V7.md",
                description: "Deterministic UI generation contract, route-priority testing depth, and MVP UX gate criteria",
            },
            {
                slug: "v8-runtime-contracts",
                label: "V8 Runtime Contracts",
                path: "docs/architecture-library/V8_RUNTIME_CONTRACTS.md",
                description: "Canonical V8 contract shell for inception, kernel, council, provider policy, and continuity state",
            },
            {
                slug: "v8-config-bootstrap-model",
                label: "V8 Config and Bootstrap Model",
                path: "docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md",
                description: "Canonical V8 bootstrap + V7->V8 migration contract covering templates vs instantiated orgs, entry points, inheritance, and precedence",
            },
            {
                slug: "v8-ui-api-operator-experience-contract",
                label: "V8 UI/API/Operator Contract",
                path: "docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md",
                description: "Canonical V8 PRD for first-run, Soma-primary operator flow, advanced architecture/runtime boundaries, and screen-to-API mapping",
            },
            {
                slug: "v8-1-living-organization-architecture",
                label: "Current Release Architecture (V8.1)",
                path: "docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md",
                description: "Current bounded release architecture for Loop Profiles, Learning Loops, semantic continuity, Procedure / Skill Sets, promoted Agent Type/Response Contract inheritance, and bounded Automations visibility",
            },
            {
                slug: "v8-2-cross-repo-cleanup-release-plan",
                label: "V8.2 Cleanup + Release Structure Plan",
                path: "docs/architecture-library/V8_2_CROSS_REPO_CLEANUP_AND_RELEASE_STRUCTURE_PLAN.md",
                description: "Prioritized cleanup lanes, commit boundaries, owner roles, and release validation order for the current mixed worktree",
            },
            {
                slug: "v8-2-full-production-architecture",
                label: "Full Architecture (V8.2)",
                path: "v8-2.md",
                description: "Canonical full production architecture for distributed execution, governed learning, capability-backed execution, and full actuation scope",
            },
            {
                slug: "arch-overview",
                label: "Overview",
                path: "docs/architecture/OVERVIEW.md",
                description: "Philosophy, 4-layer anatomy, phases, roadmap",
            },
            {
                slug: "arch-backend",
                label: "Backend",
                path: "docs/architecture/BACKEND.md",
                description: "Go packages, APIs, DB schema, NATS, execution pipelines",
            },
            {
                slug: "arch-frontend",
                label: "Frontend",
                path: "docs/architecture/FRONTEND.md",
                description: "Routes, components, Zustand, design system",
            },
            {
                slug: "arch-operations",
                label: "Operations",
                path: "docs/architecture/OPERATIONS.md",
                description: "Deployment, config, testing, CI/CD",
            },
            {
                slug: "arch-ui-target-transaction-contract-v7",
                label: "UI Target + Transaction Contract",
                path: "docs/architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md",
                description: "Required UI terminal states, backend effects, and product-flow proof expectations",
            },
            {
                slug: "arch-memory-service",
                label: "Memory Service",
                path: "docs/architecture/DIRECTIVE_MEMORY_SERVICE.md",
                description: "State Engine design — event projection, pgvector, log_entries schema",
            },
            {
                slug: "v7-architecture-prd",
                label: "V7 Architecture PRD Index",
                path: "mycelis-architecture-v7.md",
                description: "Stable compatibility entrypoint pointing to the modular architecture library",
            },
            {
                slug: "v7-mcp-baseline",
                label: "V7 MCP Baseline",
                path: "docs/V7_MCP_BASELINE.md",
                description: "MVOS: filesystem, memory, artifact-renderer, fetch servers",
            },
            {
                slug: "arch-mcp-service-config-local-first",
                label: "MCP Service Config (Local-First)",
                path: "docs/architecture/MCP_SERVICE_CONFIGURATION_LOCAL_FIRST_V7.md",
                description: "Canonical process and configuration standard for adding MCP services with local-default posture",
            },
            {
                slug: "arch-universal-action-interface-v7",
                label: "Universal Action Interface V7",
                path: "docs/architecture/UNIVERSAL_ACTION_INTERFACE_V7.md",
                description: "Canonical universal action contracts, dynamic service API, and Python management interface",
            },
            {
                slug: "arch-agentry-template-marketplace-v7",
                label: "Template Marketplace + Custom",
                path: "docs/architecture/AGENTRY_TEMPLATE_MARKETPLACE_AND_CUSTOM_TEMPLATING_V7.md",
                description: "API and governance model for marketplace template acquisition (ClawHub-style) and tenant custom template publishing",
            },
            {
                slug: "arch-actualization-beyond-mcp-v7",
                label: "Actualization Beyond MCP V7",
                path: "docs/architecture/ACTUALIZATION_ARCHITECTURE_BEYOND_MCP_V7.md",
                description: "Multi-protocol actualization strategy across MCP, OpenAPI, A2A, ACP, and Python action management",
            },
            {
                slug: "arch-secure-gateway-remote-actuation-v7",
                label: "Secure Gateway + Remote Actuation",
                path: "docs/architecture/SECURE_GATEWAY_REMOTE_ACTUATION_PROFILE_V7.md",
                description: "Security baseline for self-hosted gateway patterns and remote actuation services",
            },
            {
                slug: "arch-hardware-interface-api-v7",
                label: "Hardware Interface API + Channels",
                path: "docs/architecture/HARDWARE_INTERFACE_API_AND_CHANNELS_V7.md",
                description: "Hardware interface control-plane APIs and direct channel support standards",
            },
            {
                slug: "arch-soma-symbiote-growth-host-actuation-v7",
                label: "Soma Symbiote + Host Actuation",
                path: "docs/architecture/SOMA_SYMBIOTE_GROWTH_AND_HOST_ACTUATION_V7.md",
                description: "Soma thought-profile contracts, learning-growth loop, and localhost host actuation model",
            },
            {
                slug: "arch-soma-team-channels",
                label: "Soma Team + Channel Architecture",
                path: "docs/architecture/SOMA_TEAM_CHANNEL_ARCHITECTURE_V7.md",
                description: "Canonical inter-team/process/MCP channel contracts and shared memory boundaries",
            },
            {
                slug: "arch-nats-signal-standard-v7",
                label: "NATS Signal Standard V7",
                path: "docs/architecture/NATS_SIGNAL_STANDARD_V7.md",
                description: "Canonical NATS subject families, source normalization, and product-vs-dev channel boundaries",
            },
            {
                slug: "arch-workflow-composer-delivery-v7",
                label: "Workflow Composer Delivery Plan V7",
                path: "docs/architecture/WORKFLOW_COMPOSER_DELIVERY_V7.md",
                description: "Airflow-style DAG workflow composer plan with team lanes, release gates, git discipline, and invoke task strategy",
            },
            {
                slug: "arch-soma-council-engagement-protocol-v7",
                label: "Soma-Council Engagement Protocol V7",
                path: "docs/architecture/SOMA_COUNCIL_ENGAGEMENT_PROTOCOL_V7.md",
                description: "Deterministic path-selection contract for internal tools, MCP, external APIs, and code-to-execution loops",
            },
            {
                slug: "arch-agent-source-instantiation-template-v7",
                label: "Agent Source Template V7",
                path: "docs/architecture/AGENT_SOURCE_INSTANTIATION_TEMPLATE_V7.md",
                description: "Standardized provider instantiation template with Ollama as default and governed overrides for OpenAI, Claude, Gemini, vLLM, and LM Studio",
            },
        ],
    },

    // ── Governance & Testing ──────────────────────────────────────────────────
    {
        section: "Governance & Testing",
        docs: [
            {
                slug: "governance",
                label: "Governance System",
                path: "docs/governance.md",
                description: "Policy enforcement, ALLOW/DENY/REQUIRE_APPROVAL, scenario configs",
            },
            {
                slug: "testing",
                label: "Testing",
                path: "docs/TESTING.md",
                description: "Unit, integration, smoke test protocols",
            },
            {
                slug: "v8-ui-testing-agentry-contract",
                label: "V8 UI Testing Agentry",
                path: "docs/architecture-library/V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md",
                description: "Canonical product-contract proof for Soma-first browser testing, governed actions, continuity, audit visibility, and trust recovery",
            },
            {
                slug: "v8-ui-testing-stabilization-strike-plan",
                label: "V8 UI Testing Strike Plan",
                path: "docs/architecture-library/V8_UI_TESTING_STABILIZATION_STRIKE_TEAM_PLAN.md",
                description: "Active strike-team ownership plan for stabilizing mocked browser proof, live governed-chat proof, and release hygiene",
            },
            {
                slug: "v8-release-platform-review",
                label: "V8 Release Platform Review",
                path: "docs/architecture-library/V8_RELEASE_PLATFORM_REVIEW_SECURITY_MONITORING_DEBUG.md",
                description: "Shared release-readiness review across governance/security, monitoring/ops, and debug/live-browser proof",
            },
        ],
    },

];

/** Flat lookup map: slug → DocEntry. Used by the API route for validation. */
export const DOC_BY_SLUG: Map<string, DocEntry> = new Map(
    DOC_MANIFEST.flatMap((section) => section.docs).map((doc) => [doc.slug, doc])
);
