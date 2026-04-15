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
 * This manifest intentionally exposes precise, durable documentation only.
 * Temporary planning notes, strike-team plans, runbooks, and execution briefs
 * do not belong in the canonical in-app documentation surface.
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
    {
        section: "User Guides",
        docs: [
            {
                slug: "user-docs-home",
                label: "User Docs Home",
                path: "docs/user/README.md",
                description: "Operator-first entry point for Soma, teams, resources, memory, and recovery workflows",
            },
            {
                slug: "deployment-method-selection",
                label: "Deployment Method Selection",
                path: "docs/user/deployment-methods.md",
                description: "Pick Docker Compose, local k3d, enterprise self-hosted Kubernetes, or edge deployment by target environment",
            },
            {
                slug: "core-concepts",
                label: "Core Concepts",
                path: "docs/user/core-concepts.md",
                description: "Soma, Council, Mission, Run, Brain, Event, and Trust in operator language",
            },
            {
                slug: "system-status-recovery",
                label: "System Status & Recovery",
                path: "docs/user/system-status-recovery.md",
                description: "Health signals, degraded recovery actions, and Quick Checks workflow",
            },
            {
                slug: "soma-chat",
                label: "Using Soma Chat",
                path: "docs/user/soma-chat.md",
                description: "Send messages, read delegation traces, and confirm governed proposals",
            },
            {
                slug: "workflow-variants-plan-memory",
                label: "Workflow Variants + Plan Memory",
                path: "docs/user/workflow-variants-and-plan-memory.md",
                description: "Choose between direct Soma, compact teams, and multi-lane workflows, and keep important plans durable across reboots",
            },
            {
                slug: "meta-agent-blueprint",
                label: "Meta-Agent & Blueprints",
                path: "docs/user/meta-agent-blueprint.md",
                description: "How Architect generates mission blueprints across teams, tools, and I/O contracts",
            },
            {
                slug: "run-timeline",
                label: "Run Timeline",
                path: "docs/user/run-timeline.md",
                description: "Reading execution timelines, status changes, and navigation paths",
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
                description: "Advanced tools, workspace files, AI engines, and governed private/deployment/reflection context",
            },
            {
                slug: "teams-guide",
                label: "Teams",
                path: "docs/user/teams.md",
                description: "Compact team defaults, broad-ask splitting, and lead-centered team workflows",
            },
            {
                slug: "memory-guide",
                label: "Memory",
                path: "docs/user/memory.md",
                description: "Semantic search, retained knowledge, private/deployment/reflection context, and continuity rules",
            },
            {
                slug: "governance-trust",
                label: "Governance & Trust",
                path: "docs/user/governance-trust.md",
                description: "Approval posture, risk classes, audit visibility, and operator control",
            },
            {
                slug: "licensing-editions",
                label: "Licensing & Editions",
                path: "docs/licensing.md",
                description: "Product-edition posture for self-hosted release, enterprise add-ons, and hosted control-plane layering",
            },
        ],
    },
    {
        section: "Getting Started",
        docs: [
            {
                slug: "docs-home",
                label: "Docs Home",
                path: "docs/README.md",
                description: "Clean navigation layer between user guidance, developer guidance, testing guidance, and historical references",
            },
            {
                slug: "readme",
                label: "Overview",
                path: "README.md",
                description: "Primary development-swarm inception document and layered truth summary",
            },
            {
                slug: "local-dev",
                label: "Local Dev Workflow",
                path: "docs/LOCAL_DEV_WORKFLOW.md",
                description: "Setup, config reference, port map, and troubleshooting guidance",
            },
            {
                slug: "remote-user-testing",
                label: "Remote User Testing",
                path: "docs/REMOTE_USER_TESTING.md",
                description: "Walkthrough for networked user testing of Soma, governance, MCP visibility, and safe actuation paths",
            },
            {
                slug: "v8-dev-state",
                label: "V8 Dev State",
                path: "V8_DEV_STATE.md",
                description: "Live implementation scoreboard for current V8 delivery status, blockers, and evidence",
            },
            {
                slug: "v7-dev-state",
                label: "V7 Dev State (Historical)",
                path: "V7_DEV_STATE.md",
                description: "Historical migration checkpoint retained as legacy input, not live authority",
            },
        ],
    },
    {
        section: "Soma Workflow",
        docs: [
            {
                slug: "workflows",
                label: "Workflow Spec",
                path: "docs/WORKFLOWS.md",
                description: "User-facing workflow specifications and implementation authority",
            },
            {
                slug: "swarm-operations",
                label: "Swarm Operations",
                path: "docs/SWARM_OPERATIONS.md",
                description: "Hierarchy, blueprints, activation, teams, tools, and governance surfaces",
            },
            {
                slug: "council-chat-qa",
                label: "Council Chat QA",
                path: "docs/QA_COUNCIL_CHAT_API.md",
                description: "QA procedures and test cases for the council chat API",
            },
        ],
    },
    {
        section: "API Reference",
        docs: [
            {
                slug: "api-reference",
                label: "API Reference",
                path: "docs/API_REFERENCE.md",
                description: "Endpoint table with request and response shapes",
            },
            {
                slug: "cognitive-architecture",
                label: "Cognitive Architecture",
                path: "docs/COGNITIVE_ARCHITECTURE.md",
                description: "Providers, profiles, embeddings, and AI engine architecture",
            },
            {
                slug: "logging-schema",
                label: "Logging Standard (V7)",
                path: "docs/logging.md",
                description: "Mission-events and memory-stream logging contract and taxonomy",
            },
        ],
    },
    {
        section: "Architecture",
        docs: [
            {
                slug: "architecture-library-index",
                label: "Architecture Library Index",
                path: "docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md",
                description: "Canonical modular index for target delivery, architecture, UI, and testing authority",
            },
            {
                slug: "target-deliverable-v7",
                label: "Target Deliverable V7",
                path: "docs/architecture-library/TARGET_DELIVERABLE_V7.md",
                description: "Product end state, phase framing, and success criteria",
            },
            {
                slug: "system-architecture-v7",
                label: "System Architecture V7",
                path: "docs/architecture-library/SYSTEM_ARCHITECTURE_V7.md",
                description: "Runtime layers, persistence model, deployment posture, and bus/storage rules",
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
                description: "Intent-to-manifestation flow, module abstraction, and created-team interaction model",
            },
            {
                slug: "ui-operator-experience-v7",
                label: "UI And Operator Experience V7",
                path: "docs/architecture-library/UI_AND_OPERATOR_EXPERIENCE_V7.md",
                description: "Operator journeys, anti-information-swarm rules, and intuitive UI targets",
            },
            {
                slug: "delivery-governance-testing-v7",
                label: "Delivery Governance And Testing V7",
                path: "docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md",
                description: "Delivery proof model, evidence requirements, and product-aligned testing expectations",
            },
            {
                slug: "team-execution-global-state-protocol-v7",
                label: "Team Execution + Global State Protocol V7",
                path: "docs/architecture-library/TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md",
                description: "Multi-lane execution discipline, state maintenance rules, and deep-testing obligations",
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
                description: "Canonical V7-to-V8 bootstrap, template, instantiation, inheritance, and precedence contract",
            },
            {
                slug: "v8-ui-api-operator-experience-contract",
                label: "V8 UI/API/Operator Contract",
                path: "docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md",
                description: "Canonical operator workflow contract for AI Organization creation and Soma-first runtime behavior",
            },
            {
                slug: "v8-1-living-organization-architecture",
                label: "Current Release Architecture (V8.1)",
                path: "docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md",
                description: "Current bounded release architecture for loops, continuity, capabilities, and bounded automations",
            },
            {
                slug: "v8-compact-team-orchestration-defaults",
                label: "V8 Compact Team Orchestration",
                path: "docs/architecture-library/V8_COMPACT_TEAM_ORCHESTRATION_AND_DEFAULTS.md",
                description: "Compact-team defaults, broad-ask splitting, and Soma/Council orchestration over NATS",
            },
            {
                slug: "v8-teamed-agentry-workflow-advantage",
                label: "V8 Teamed Agentry Advantage",
                path: "docs/architecture-library/V8_TEAMED_AGENTRY_WORKFLOW_ADVANTAGE.md",
                description: "Workflow variants, reboot-safe plan continuity, and the boundary between direct Soma, compact teams, and multi-lane coordinated agentry",
            },
            {
                slug: "v8-universal-soma-context-model",
                label: "V8 Universal Soma Context Model",
                path: "docs/architecture-library/V8_UNIVERSAL_SOMA_AND_CONTEXT_MODEL_PRD.md",
                description: "Canonical PRD for one persistent Soma/Council pair with scoped execution across contexts",
            },
            {
                slug: "v8-multi-user-identity-soma-tenancy",
                label: "V8 Multi-User Identity + Soma Tenancy",
                path: "docs/architecture-library/V8_MULTI_USER_IDENTITY_AND_SOMA_TENANCY.md",
                description: "Target contract for SAML/SSO, break-glass admins, modular IAM, and one shared Soma persona",
            },
            {
                slug: "v8-memory-layer-reflection-delivery",
                label: "V8 Memory Layer + Reflection",
                path: "docs/architecture-library/V8_MEMORY_LAYER_AND_REFLECTION_DELIVERY_CONTRACT.md",
                description: "Explicit Soma/agent/project/reflection memory layers, candidate-first reflection, and promotion guardrails",
            },
            {
                slug: "v8-home-docker-compose-runtime",
                label: "V8 Home Docker Compose Runtime",
                path: "docs/architecture-library/V8_HOME_DOCKER_COMPOSE_RUNTIME.md",
                description: "Single-host Docker Compose runtime for home-lab, demo, and partner review",
            },
            {
                slug: "v8-self-hosted-runtime-delivery-program",
                label: "V8 Self-Hosted Runtime Delivery",
                path: "docs/architecture-library/V8_SELF_HOSTED_RUNTIME_DELIVERY_PROGRAM.md",
                description: "Compact delivery-team contract for deployable Compose and self-hosted Kubernetes runtime work",
            },
            {
                slug: "v8-enterprise-self-hosted-kubernetes-delivery-plan",
                label: "V8 Enterprise Self-Hosted Kubernetes",
                path: "docs/architecture-library/V8_ENTERPRISE_SELF_HOSTED_KUBERNETES_DELIVERY_PLAN.md",
                description: "Enterprise-compatible Kubernetes delivery plan with local k3d validation, chart promotion rules, and acceptance gates",
            },
            {
                slug: "v8-compose-personal-owner-test-plan",
                label: "V8 Compose Personal Owner Test Plan",
                path: "docs/architecture-library/V8_COMPOSE_PERSONAL_OWNER_DEPLOYMENT_TEST_PLAN.md",
                description: "PRD-style test plan for personal-owner Compose deployment from data-plane-only startup through workflow proof",
            },
            {
                slug: "v8-mvp-media-team-output-template-registry",
                label: "V8 MVP Media + Team Output",
                path: "docs/architecture-library/V8_MVP_MEDIA_TEAM_OUTPUT_AND_TEMPLATE_REGISTRY.md",
                description: "User-output-first media/team proof, Ollama role routing, and conversation-template registry plan",
            },
            {
                slug: "v8-2-full-production-architecture",
                label: "Full Architecture (V8.2)",
                path: "v8-2.md",
                description: "Canonical full production architecture for distributed execution, governed learning, and actuation",
            },
            {
                slug: "arch-overview",
                label: "Overview",
                path: "docs/architecture/OVERVIEW.md",
                description: "Philosophy, 4-layer anatomy, phases, and roadmap",
            },
            {
                slug: "arch-backend",
                label: "Backend",
                path: "docs/architecture/BACKEND.md",
                description: "Go packages, APIs, DB schema, NATS, and execution pipelines",
            },
            {
                slug: "arch-frontend",
                label: "Frontend",
                path: "docs/architecture/FRONTEND.md",
                description: "Routes, components, Zustand, and design system",
            },
            {
                slug: "arch-operations",
                label: "Operations",
                path: "docs/architecture/OPERATIONS.md",
                description: "Deployment, config, testing, and CI/CD guidance",
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
                description: "State engine design, pgvector usage, and directive-memory schema",
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
                description: "Filesystem, memory, artifact-renderer, and fetch servers baseline",
            },
            {
                slug: "arch-mcp-service-config-local-first",
                label: "MCP Service Config (Local-First)",
                path: "docs/architecture/MCP_SERVICE_CONFIGURATION_LOCAL_FIRST_V7.md",
                description: "Local-default standard for adding MCP services with governed configuration",
            },
            {
                slug: "arch-universal-action-interface-v7",
                label: "Universal Action Interface V7",
                path: "docs/architecture/UNIVERSAL_ACTION_INTERFACE_V7.md",
                description: "Universal action contracts, dynamic service API, and Python management interface",
            },
            {
                slug: "arch-agentry-template-marketplace-v7",
                label: "Template Marketplace + Custom",
                path: "docs/architecture/AGENTRY_TEMPLATE_MARKETPLACE_AND_CUSTOM_TEMPLATING_V7.md",
                description: "Marketplace acquisition and tenant-owned custom template publishing model",
            },
            {
                slug: "arch-actualization-beyond-mcp-v7",
                label: "Actualization Beyond MCP V7",
                path: "docs/architecture/ACTUALIZATION_ARCHITECTURE_BEYOND_MCP_V7.md",
                description: "Multi-protocol architecture across MCP, OpenAPI, A2A, ACP, and Python action management",
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
                description: "Hardware control-plane APIs and direct channel support standards",
            },
            {
                slug: "arch-soma-symbiote-growth-host-actuation-v7",
                label: "Soma Symbiote + Host Actuation",
                path: "docs/architecture/SOMA_SYMBIOTE_GROWTH_AND_HOST_ACTUATION_V7.md",
                description: "Soma thought-profile contracts, learning-growth loop, and localhost actuation model",
            },
            {
                slug: "arch-soma-team-channels",
                label: "Soma Team + Channel Architecture",
                path: "docs/architecture/SOMA_TEAM_CHANNEL_ARCHITECTURE_V7.md",
                description: "Inter-team and MCP channel contracts plus shared memory boundaries",
            },
            {
                slug: "arch-nats-signal-standard-v7",
                label: "NATS Signal Standard V7",
                path: "docs/architecture/NATS_SIGNAL_STANDARD_V7.md",
                description: "Canonical subject families, source normalization, and product-vs-dev channel boundaries",
            },
            {
                slug: "arch-workflow-composer-delivery-v7",
                label: "Workflow Composer Delivery Plan V7",
                path: "docs/architecture/WORKFLOW_COMPOSER_DELIVERY_V7.md",
                description: "DAG workflow composer delivery contract, release gates, and invoke task strategy",
            },
            {
                slug: "arch-soma-council-engagement-protocol-v7",
                label: "Soma-Council Engagement Protocol V7",
                path: "docs/architecture/SOMA_COUNCIL_ENGAGEMENT_PROTOCOL_V7.md",
                description: "Path-selection contract for internal tools, MCP, external APIs, and code-to-execution loops",
            },
            {
                slug: "arch-agent-source-instantiation-template-v7",
                label: "Agent Source Template V7",
                path: "docs/architecture/AGENT_SOURCE_INSTANTIATION_TEMPLATE_V7.md",
                description: "Provider instantiation template with Ollama default and governed overrides",
            },
        ],
    },
    {
        section: "Governance & Testing",
        docs: [
            {
                slug: "governance",
                label: "Governance System",
                path: "docs/governance.md",
                description: "Policy enforcement, approval posture, and audit-linked governance model",
            },
            {
                slug: "testing",
                label: "Testing",
                path: "docs/TESTING.md",
                description: "Unit, integration, browser, and release validation guidance",
            },
            {
                slug: "v8-ui-testing-agentry-contract",
                label: "V8 UI Testing Agentry",
                path: "docs/architecture-library/V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md",
                description: "Canonical browser proof contract for Soma-first flows, governed actions, and trust recovery",
            },
            {
                slug: "v8-ui-team-full-test-set",
                label: "V8 UI Team Full Test Set",
                path: "docs/architecture-library/V8_UI_TEAM_FULL_TEST_SET.md",
                description: "Full UI validation set for browser workflows, runtime proof, and final verdict rules",
            },
        ],
    },
    {
        section: "Archive",
        docs: [
            {
                slug: "archive-index",
                label: "Archive Index",
                path: "docs/archive/README.md",
                description: "Historical documents only and not active implementation authority",
            },
        ],
    },
];

/** Flat lookup map: slug → DocEntry. Used by the API route for validation. */
export const DOC_BY_SLUG: Map<string, DocEntry> = new Map(
    DOC_MANIFEST.flatMap((section) => section.docs).map((doc) => [doc.slug, doc])
);
