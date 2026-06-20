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
            { slug: "meta-agent-blueprint", label: "Blueprints And Mission Planning", path: "docs/user/meta-agent-blueprint.md", description: "Advanced blueprint and mission-planning language for graph-level planning" },
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
            { slug: "automations-guide", label: "Automations", path: "docs/user/automations.md", description: "Event trigger rules, schedules, mission profiles, and approvals around automated work" },
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
            { slug: "v8-3-release-architecture-delivery-brief", label: "V8.3 Release Architecture Brief", path: "docs/architecture-library/V8_3_RELEASE_ARCHITECTURE_DELIVERY_BRIEF.md", description: "Compact shareable architecture document for outcome-centric MVP-to-release execution, delivery lanes, P0 train, proof, and acceptance standard" },
            { slug: "v8-3-autonomy-control-architecture", label: "V8.3 Autonomy Control", path: "docs/architecture-library/V8_3_AUTONOMY_CONTROL_ARCHITECTURE.md", description: "Control-first boundary for future autonomous, adaptive, and self-improving behavior without expanding the MVP scope" },
            { slug: "v8-3-operational-embodiment-prd", label: "V8.3 Embodiment PRD", path: "docs/architecture-library/V8_3_OPERATIONAL_EMBODIMENT_PRD.md", description: "Outcome-primary release plan for async runtime, recovery actions, ExpressionFrame, local media proof, and new-user GUI gates" },
            { slug: "v8-3-product-manifestation-review", label: "V8.3 Product Manifestation Review", path: "docs/architecture-library/V8_3_PRODUCT_MANIFESTATION_REVIEW.md", description: "Outcome-value review of major subsystems, visibility boundaries, MVP classification, and adoption risk" },
            { slug: "v8-3-soma-user-experience-contract", label: "V8.3 Soma UX Contract", path: "docs/architecture-library/V8_3_SOMA_USER_EXPERIENCE_CONTRACT.md", description: "Operator-first contract for working through Soma, outcomes, deliverables, proof, recovery, capabilities, and optional Inspect" },
            { slug: "v8-3-mvp-ui-runtime-delivery-plan", label: "V8.3 MVP UI Runtime Delivery", path: "docs/architecture-library/V8_3_MVP_UI_RUNTIME_DELIVERY_PLAN.md", description: "Executable plan for outcome workspace, run receipts, output packages, capability catalog, recovery queue, and Trusted Outcome Journey proof" },
            { slug: "v8-new-user-acceptance-matrix", label: "V8 New-User Acceptance", path: "docs/architecture-library/V8_NEW_USER_ACCEPTANCE_MATRIX.md", description: "Fresh-operator acceptance matrix for login, Soma, active work, outputs, proof, recovery, and retained artifact re-entry" },
            { slug: "v8-ui-api-operator-experience-contract", label: "V8 UI/API/Operator Contract", path: "docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md", description: "Canonical operator workflow contract for Soma-first behavior" },
            { slug: "v8-capability-manifest-runtime-integration", label: "V8 Capability Manifest", path: "docs/architecture-library/V8_CAPABILITY_MANIFEST_AND_RUNTIME_INTEGRATION_STANDARD.md", description: "Governed capability manifests, permissions, run proof, output normalization, and recovery" },
            { slug: "v8-secret-storage-credential-boundary", label: "V8 Secret Storage Boundary", path: "docs/architecture-library/V8_SECRET_STORAGE_AND_CREDENTIAL_BOUNDARY.md", description: "Secret references, runtime resolution, UI exposure, audit/proof boundaries, and rotation posture" },
            { slug: "v8-runtime-contracts", label: "V8 Runtime Contracts", path: "docs/architecture-library/V8_RUNTIME_CONTRACTS.md", description: "Instantiated organization runtime truth, Soma, Council, provider policy, and continuity contracts" },
            { slug: "v8-config-bootstrap-model", label: "V8 Config and Bootstrap Model", path: "docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md", description: "Template, instantiation, inheritance, precedence, and migration truth" },
            { slug: "v8-2-full-production-architecture", label: "Full Architecture (V8.2)", path: "architecture/v8-2.md", description: "Canonical full production architecture for distributed execution, governed learning, and actuation" },
            { slug: "v8-2-soma-ui-architecture-expression", label: "V8.2 Soma UI Expression", path: "docs/architecture-library/V8_2_SOMA_UI_ARCHITECTURE_EXPRESSION.md", description: "Ideal Soma operating loop, UI architecture domains, active work, outputs, proof, teams, and acceptance gates" },
            { slug: "v8-2-soma-team-interaction-contract", label: "V8.2 Soma Team Interaction", path: "docs/architecture-library/V8_2_SOMA_TEAM_INTERACTION_CONTRACT.md", description: "Contract for Soma, Council, operators, and runtime teams to inspect, steer, recover, and archive active work" },
            { slug: "v8-ui-team-full-test-set", label: "V8 UI Team Full Test Set", path: "docs/architecture-library/V8_UI_TEAM_FULL_TEST_SET.md", description: "Full UI validation contract for browser workflows, runtime proof, and verdict rules" },
            { slug: "arch-overview", label: "Architecture Overview", path: "docs/architecture/OVERVIEW.md", description: "Current architecture overview for the V8.3 embodiment target and V8.2 baseline" },
            { slug: "arch-backend", label: "Backend", path: "docs/architecture/BACKEND.md", description: "Go packages, APIs, DB schema, NATS, and execution pipelines" },
            { slug: "arch-frontend", label: "Frontend", path: "docs/architecture/FRONTEND.md", description: "Routes, components, Zustand, and design system" },
            { slug: "v7-architecture-prd", label: "PRD Compatibility Index", path: "architecture/mycelis-architecture-v7.md", description: "Compatibility entrypoint for old references; current authority points to V8.3 embodiment docs and the V8.2 baseline" },
        ],
    },
];

/** Flat lookup map: slug -> DocEntry. Used by the API route for validation. */
export const DOC_BY_SLUG: Map<string, DocEntry> = new Map(
    DOC_MANIFEST.flatMap((section) => section.docs).map((doc) => [doc.slug, doc])
);
