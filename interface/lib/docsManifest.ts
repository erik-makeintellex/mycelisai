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
                description: "Active missions, drafts, trigger rules, approvals, scheduled runs",
            },
            {
                slug: "resources-guide",
                label: "Resources",
                path: "docs/user/resources.md",
                description: "Brains/providers, MCP tools, workspace filesystem, agent catalogue",
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
                slug: "v7-dev-state",
                label: "V7 Dev State",
                path: "V7_DEV_STATE.md",
                description: "Current authoritative map of what's done vs pending",
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
        ],
    },

];

/** Flat lookup map: slug → DocEntry. Used by the API route for validation. */
export const DOC_BY_SLUG: Map<string, DocEntry> = new Map(
    DOC_MANIFEST.flatMap((section) => section.docs).map((doc) => [doc.slug, doc])
);

