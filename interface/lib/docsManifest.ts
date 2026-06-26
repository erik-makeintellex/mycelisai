/**
 * Docs Manifest
 *
 * Curated registry of documentation files served by the in-app doc browser.
 * This surface is intentionally small: user guides, repo operating docs, and
 * the active architecture contracts only. Stale planning notes and superseded
 * version docs do not belong in the operator-facing documentation surface.
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
        section: "Start Here",
        docs: [
            { slug: "user-docs-home", label: "Use Mycelis", path: "docs/user/README.md", description: "Operator-first entry point for Soma, teams, resources, outputs, and recovery workflows" },
            { slug: "soma-chat", label: "Using Soma Chat", path: "docs/user/soma-chat.md", description: "Concrete Soma prompts, outputs, delegation traces, governed proposals, and recovery" },
            { slug: "workflow-variants-plan-memory", label: "Workflow Variants + Plan Memory", path: "docs/user/workflow-variants-and-plan-memory.md", description: "Direct Soma, compact teams, multi-lane workflows, and durable plan memory" },
            { slug: "teams-guide", label: "Teams", path: "docs/user/teams.md", description: "Active team work, compact defaults, broad-ask splitting, and lead-centered workflows" },
            { slug: "resources-guide", label: "Outputs And Resources", path: "docs/user/resources.md", description: "Output files, group artifacts, capabilities, MCP structure review, AI engines, and governed context" },
        ],
    },
    {
        section: "Trust And Setup",
        docs: [
            { slug: "core-concepts", label: "Core Concepts", path: "docs/user/core-concepts.md", description: "Soma, teams, memory, governance, runs, and trust in operator language" },
            { slug: "memory-guide", label: "Memory", path: "docs/user/memory.md", description: "Trusted recall, memory lanes, governed context, and continuity rules" },
            { slug: "governance-trust", label: "Governance & Trust", path: "docs/user/governance-trust.md", description: "Approval posture, risk classes, audit visibility, and trusted-memory precedence" },
            { slug: "settings-access", label: "Settings And Access", path: "docs/user/settings-access.md", description: "Profile, access posture, auth providers, and connected-tool/search boundaries" },
            { slug: "auth-modes", label: "Authentication Modes", path: "docs/user/auth-modes.md", description: "Local owner auth, break-glass recovery, OIDC/OAuth, SAML, Entra ID, Google Workspace, GitHub, and future SCIM posture" },
            { slug: "deployment-method-selection", label: "Deployment Method Selection", path: "docs/user/deployment-methods.md", description: "Choose Compose, Kubernetes, enterprise self-hosted, or edge deployment lanes" },
            { slug: "system-status-recovery", label: "System Status & Recovery", path: "docs/user/system-status-recovery.md", description: "Health signals, degraded recovery actions, and System Checks workflow" },
            { slug: "run-timeline", label: "Run Timeline", path: "docs/user/run-timeline.md", description: "Execution timelines, status changes, and run navigation paths" },
        ],
    },
    {
        section: "Advanced User Surfaces",
        docs: [
            { slug: "automations-guide", label: "Automations", path: "docs/user/automations.md", description: "Event trigger rules, schedules, mission profiles, and approvals around automated work" },
            { slug: "meta-agent-blueprint", label: "Blueprints And Mission Planning", path: "docs/user/meta-agent-blueprint.md", description: "Advanced blueprint and mission-planning language for graph-level planning" },
        ],
    },
    {
        section: "Contributor Docs",
        docs: [
            { slug: "docs-home", label: "Docs Home", path: "docs/README.md", description: "Clean navigation layer for user, developer, testing, release, and compatibility docs" },
            { slug: "readme", label: "Repository Overview", path: "README.md", description: "Primary development-swarm inception document and command contract" },
            { slug: "local-dev", label: "Local Dev Workflow", path: "docs/LOCAL_DEV_WORKFLOW.md", description: "Setup, config reference, port map, and troubleshooting guidance" },
            { slug: "operations", label: "Operations", path: "docs/architecture/OPERATIONS.md", description: "Task ownership, lifecycle, Compose, Kubernetes, CI, and release-lane sequencing" },
            { slug: "testing", label: "Testing", path: "docs/TESTING.md", description: "Unit, integration, browser, and release validation guidance" },
            { slug: "api-reference", label: "API Reference", path: "docs/API_REFERENCE.md", description: "Endpoint table with request and response shapes" },
            { slug: "cognitive-architecture", label: "Cognitive Architecture", path: "docs/COGNITIVE_ARCHITECTURE.md", description: "Provider routing, AI engines, local media gateway, and model/embedding configuration" },
            { slug: "release-handoff", label: "Release Handoff", path: "docs/RELEASE_HANDOFF.md", description: "Current release handoff, deployment lanes, GUI proof commands, and packaging commands" },
            { slug: "licensing-editions", label: "Licensing & Editions", path: "docs/licensing.md", description: "Product-edition posture for self-hosted, enterprise, and hosted layering" },
            { slug: "governance", label: "Governance System", path: "docs/governance.md", description: "Policy enforcement, approval posture, and audit-linked governance model" },
            { slug: "logging-schema", label: "Logging Standard", path: "docs/logging.md", description: "Mission-events and memory-stream logging contract and taxonomy" },
        ],
    },
    {
        section: "Architecture Review",
        docs: [
            { slug: "architecture-index", label: "Architecture Docs Index", path: "docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md", description: "Curated active architecture set; stale planning notes are excluded" },
            { slug: "mycelis-canonical-prd", label: "Mycelis Canonical PRD", path: "docs/architecture-library/MYCELIS_CANONICAL_PRD.md", description: "Single source for product thesis, UX, runtime architecture, governance, outcomes, capabilities, recovery, MVP scope, P0 delivery, and release gates" },
            { slug: "arch-overview", label: "Architecture Overview", path: "docs/architecture/OVERVIEW.md", description: "Current implementation overview aligned to the canonical PRD" },
            { slug: "arch-backend", label: "Backend", path: "docs/architecture/BACKEND.md", description: "Go packages, APIs, DB schema, NATS, and execution pipelines" },
            { slug: "arch-frontend", label: "Frontend", path: "docs/architecture/FRONTEND.md", description: "Routes, components, Zustand, and design system" },
        ],
    },
];

/** Flat lookup map: slug -> DocEntry. Used by the API route for validation. */
export const DOC_BY_SLUG: Map<string, DocEntry> = new Map(
    DOC_MANIFEST.flatMap((section) => section.docs).map((doc) => [doc.slug, doc])
);
