/**
 * Docs Manifest
 *
 * Curated registry of documentation files served by the in-app doc browser.
 * Every entry maps a URL-safe slug to a filesystem path relative to the
 * project root, one level above the `interface/` directory.
 *
 * This manifest intentionally exposes active, durable documentation only.
 * Superseded archive drafts and old topical V7 docs do not belong in the
 * operator-facing documentation surface.
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
            { slug: "user-docs-home", label: "User Docs Home", path: "docs/user/README.md", description: "Operator-first entry point for Soma, teams, resources, memory, and recovery workflows" },
            { slug: "deployment-method-selection", label: "Deployment Method Selection", path: "docs/user/deployment-methods.md", description: "Choose Compose, Kubernetes, enterprise self-hosted, or edge deployment lanes" },
            { slug: "core-concepts", label: "Core Concepts", path: "docs/user/core-concepts.md", description: "Soma, Council, Mission, Run, Brain, Event, and Trust in operator language" },
            { slug: "soma-chat", label: "Using Soma Chat", path: "docs/user/soma-chat.md", description: "Concrete Soma prompts, outputs, delegation traces, governed proposals, and recovery" },
            { slug: "workflow-variants-plan-memory", label: "Workflow Variants + Plan Memory", path: "docs/user/workflow-variants-and-plan-memory.md", description: "Direct Soma, compact teams, multi-lane workflows, and durable plan memory" },
            { slug: "teams-guide", label: "Teams", path: "docs/user/teams.md", description: "Active team work, compact defaults, broad-ask splitting, and lead-centered workflows" },
            { slug: "automations-guide", label: "Automations", path: "docs/user/automations.md", description: "Event trigger rules, schedules, mission profiles, and approvals around automated work" },
            { slug: "resources-guide", label: "Resources", path: "docs/user/resources.md", description: "Connected Tools, MCP structure review, workspace files, AI engines, and governed context" },
            { slug: "memory-guide", label: "Memory", path: "docs/user/memory.md", description: "Trusted recall, memory lanes, governed context, and continuity rules" },
            { slug: "governance-trust", label: "Governance & Trust", path: "docs/user/governance-trust.md", description: "Approval posture, risk classes, audit visibility, and trusted-memory precedence" },
            { slug: "settings-access", label: "Settings And Access", path: "docs/user/settings-access.md", description: "Profile, access posture, auth providers, and connected-tool/search boundaries" },
            { slug: "auth-modes", label: "Authentication Modes", path: "docs/user/auth-modes.md", description: "Local owner auth, break-glass recovery, OIDC/OAuth, SAML, Entra ID, Google Workspace, GitHub, and future SCIM posture" },
            { slug: "system-status-recovery", label: "System Status & Recovery", path: "docs/user/system-status-recovery.md", description: "Health signals, degraded recovery actions, and System Checks workflow" },
            { slug: "run-timeline", label: "Run Timeline", path: "docs/user/run-timeline.md", description: "Execution timelines, status changes, and run navigation paths" },
            { slug: "licensing-editions", label: "Licensing & Editions", path: "docs/licensing.md", description: "Product-edition posture for self-hosted, enterprise, and hosted layering" },
        ],
    },
    {
        section: "Getting Started",
        docs: [
            { slug: "docs-home", label: "Docs Home", path: "docs/README.md", description: "Clean navigation layer for user, developer, testing, release, and compatibility docs" },
            { slug: "readme", label: "Overview", path: "README.md", description: "Primary development-swarm inception document and layered truth summary" },
            { slug: "local-dev", label: "Local Dev Workflow", path: "docs/LOCAL_DEV_WORKFLOW.md", description: "Setup, config reference, port map, and troubleshooting guidance" },
            { slug: "remote-user-testing", label: "Remote User Testing", path: "docs/REMOTE_USER_TESTING.md", description: "Networked user testing of Soma, governance, MCP visibility, and safe actuation paths" },
            { slug: "release-handoff", label: "Release Handoff", path: "docs/RELEASE_HANDOFF.md", description: "Current release handoff, deployment lanes, GUI proof commands, and packaging commands" },
            { slug: "v8-dev-state", label: "V8 Dev State", path: ".state/V8_DEV_STATE.md", description: "Live implementation scoreboard for current V8 delivery status, blockers, and evidence" },
        ],
    },
    {
        section: "API Reference",
        docs: [
            { slug: "api-reference", label: "API Reference", path: "docs/API_REFERENCE.md", description: "Endpoint table with request and response shapes" },
            { slug: "cognitive-architecture", label: "Cognitive Architecture", path: "docs/COGNITIVE_ARCHITECTURE.md", description: "Providers, profiles, embeddings, and AI engine architecture" },
            { slug: "logging-schema", label: "Logging Standard", path: "docs/logging.md", description: "Mission-events and memory-stream logging contract and taxonomy" },
        ],
    },
    {
        section: "Architecture",
        docs: [
            { slug: "architecture-library-index", label: "Architecture Library Index", path: "docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md", description: "Canonical modular index for V8.2 architecture, UI, runtime, and testing authority" },
            { slug: "v8-2-current-state-finalization-prd", label: "V8.2 Current State PRD", path: "docs/architecture-library/V8_2_CURRENT_STATE_AND_FINALIZATION_PRD.md", description: "Architecture-team PRD for current truth, finalization target, workstreams, risks, and exit gates" },
            { slug: "v8-2-operational-embodiment-directive", label: "V8.2 Operational Embodiment", path: "docs/architecture-library/V8_2_OPERATIONAL_EMBODIMENT_DIRECTIVE.md", description: "Directive for visible execution, durable outputs, recoverable runs, and the canonical MVP workflow" },
            { slug: "v8-2-full-production-architecture", label: "Full Architecture (V8.2)", path: "architecture/v8-2.md", description: "Canonical full production architecture for distributed execution, governed learning, and actuation" },
            { slug: "v8-runtime-contracts", label: "V8 Runtime Contracts", path: "docs/architecture-library/V8_RUNTIME_CONTRACTS.md", description: "Inception, kernel, council, provider policy, response classes, and continuity contracts" },
            { slug: "v8-config-bootstrap-model", label: "V8 Config and Bootstrap Model", path: "docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md", description: "Template, instantiation, inheritance, precedence, and migration truth" },
            { slug: "v8-ui-api-operator-experience-contract", label: "V8 UI/API/Operator Contract", path: "docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md", description: "Canonical operator workflow contract for AI Organization creation and Soma-first behavior" },
            { slug: "v8-2-soma-ui-architecture-expression", label: "V8.2 Soma UI Expression", path: "docs/architecture-library/V8_2_SOMA_UI_ARCHITECTURE_EXPRESSION.md", description: "Ideal Soma operating loop, UI architecture domains, active work, outputs, proof, teams, and acceptance gates" },
            { slug: "v8-2-soma-team-interaction-contract", label: "V8.2 Soma Team Interaction", path: "docs/architecture-library/V8_2_SOMA_TEAM_INTERACTION_CONTRACT.md", description: "Canonical contract for Soma, Council, operators, and runtime teams to inspect, steer, recover, and archive active work" },
            { slug: "v8-directed-execution-ui-runtime-alignment", label: "V8 Directed Execution UI", path: "docs/architecture-library/V8_DIRECTED_EXECUTION_UI_RUNTIME_ALIGNMENT_DIRECTIVE.md", description: "UI/runtime alignment for Soma-centered directed execution, outputs, proof, and governance" },
            { slug: "v8-directed-execution-delivery-plan", label: "V8 Directed Execution Delivery", path: "docs/architecture-library/V8_DIRECTED_EXECUTION_DELIVERY_PLAN.md", description: "Team execution plan for directed-execution packages, waves, dependencies, and gates" },
            { slug: "v8-capability-manifest-runtime-integration", label: "V8 Capability Manifest", path: "docs/architecture-library/V8_CAPABILITY_MANIFEST_AND_RUNTIME_INTEGRATION_STANDARD.md", description: "Governed capability manifests, permissions, run proof, output normalization, and recovery" },
            { slug: "v8-1-living-organization-architecture", label: "V8.1 Living Organization Baseline", path: "docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md", description: "Compatibility baseline for loops, continuity, capabilities, and bounded automations" },
            { slug: "v8-governed-execution-doctrine", label: "V8 Governed Execution Doctrine", path: "docs/architecture-library/V8_GOVERNED_EXECUTION_DOCTRINE.md", description: "Accountable cognition, Event Spine truth, workspace visibility, and delivery compression" },
            { slug: "v8-mvp-governed-execution-mission-plan", label: "V8 MVP Governed Execution Missions", path: "docs/architecture-library/V8_MVP_GOVERNED_EXECUTION_MISSION_PLAN.md", description: "Executable MVP mission plan for governed execution, events, UI manifestation, and proof" },
            { slug: "v8-compact-team-orchestration-defaults", label: "V8 Compact Team Orchestration", path: "docs/architecture-library/V8_COMPACT_TEAM_ORCHESTRATION_AND_DEFAULTS.md", description: "Compact-team defaults, broad-ask splitting, and Soma/Council coordination over NATS" },
            { slug: "v8-teamed-agentry-workflow-advantage", label: "V8 Teamed Agentry Advantage", path: "docs/architecture-library/V8_TEAMED_AGENTRY_WORKFLOW_ADVANTAGE.md", description: "Workflow variants and boundaries between direct Soma, compact teams, and multi-lane agentry" },
            { slug: "v8-workflow-variants-reboot-proof-set", label: "V8 Workflow Variants Proof Set", path: "docs/architecture-library/V8_WORKFLOW_VARIANTS_AND_REBOOT_PROOF_SET.md", description: "Comparative proof for direct Soma, compact teams, multi-lane workflows, and reboot resume" },
            { slug: "v8-universal-soma-context-model", label: "V8 Universal Soma Context Model", path: "docs/architecture-library/V8_UNIVERSAL_SOMA_AND_CONTEXT_MODEL_PRD.md", description: "One persistent Soma/Council pair with scoped execution across contexts" },
            { slug: "v8-multi-user-identity-soma-tenancy", label: "V8 Multi-User Identity + Soma Tenancy", path: "docs/architecture-library/V8_MULTI_USER_IDENTITY_AND_SOMA_TENANCY.md", description: "SAML/SSO, break-glass admins, modular IAM, and one shared Soma persona" },
            { slug: "v8-2-user-management-enterprise-auth", label: "V8.2 User Management + Enterprise Auth", path: "docs/architecture-library/V8_2_USER_MANAGEMENT_AND_ENTERPRISE_AUTH_MODULE.md", description: "Enterprise auth providers, SSO/SAML/OIDC, RBAC, approval authority, and SCIM posture" },
            { slug: "v8-memory-layer-reflection-delivery", label: "V8 Memory Layer + Reflection", path: "docs/architecture-library/V8_MEMORY_LAYER_AND_REFLECTION_DELIVERY_CONTRACT.md", description: "Memory lanes, reflection candidates, and promotion guardrails" },
            { slug: "v8-trusted-memory-arbitration-team-vector-contract", label: "V8 Trusted Memory Arbitration", path: "docs/architecture-library/V8_TRUSTED_MEMORY_ARBITRATION_AND_TEAM_VECTOR_CONTRACT.md", description: "Trusted memory control plane, evidence anchors, precedence, and bounded growth" },
            { slug: "v8-home-docker-compose-runtime", label: "V8 Home Docker Compose Runtime", path: "docs/architecture-library/V8_HOME_DOCKER_COMPOSE_RUNTIME.md", description: "Single-host Docker Compose runtime for home-lab, demo, and partner review" },
            { slug: "v8-self-hosted-runtime-delivery-program", label: "V8 Self-Hosted Runtime Delivery", path: "docs/architecture-library/V8_SELF_HOSTED_RUNTIME_DELIVERY_PROGRAM.md", description: "Compose release proof, Kubernetes scale-up, ownership, cadence, and acceptance gates" },
            { slug: "v8-mycelis-search-capability-delivery", label: "V8 Mycelis Search Capability", path: "docs/architecture-library/V8_MYCELIS_SEARCH_CAPABILITY_DELIVERY_PLAN.md", description: "Owned local-source search, local API search, optional SearXNG/Brave, and capability gates" },
            { slug: "v8-enterprise-self-hosted-kubernetes-delivery-plan", label: "V8 Enterprise Self-Hosted Kubernetes", path: "docs/architecture-library/V8_ENTERPRISE_SELF_HOSTED_KUBERNETES_DELIVERY_PLAN.md", description: "Enterprise Kubernetes delivery with Rancher K3s, k3d validation, and chart promotion" },
            { slug: "v8-compose-personal-owner-test-plan", label: "V8 Compose Personal Owner Test Plan", path: "docs/architecture-library/V8_COMPOSE_PERSONAL_OWNER_DEPLOYMENT_TEST_PLAN.md", description: "Personal-owner Compose deployment from data-plane startup through workflow proof" },
            { slug: "v8-mvp-media-team-output-template-registry", label: "V8 MVP Media + Team Output", path: "docs/architecture-library/V8_MVP_MEDIA_TEAM_OUTPUT_AND_TEMPLATE_REGISTRY.md", description: "User-output-first media/team proof, model-role routing, and conversation templates" },
            { slug: "arch-overview", label: "Overview", path: "docs/architecture/OVERVIEW.md", description: "Current architecture overview for the V8.2 operational embodiment target" },
            { slug: "arch-backend", label: "Backend", path: "docs/architecture/BACKEND.md", description: "Go packages, APIs, DB schema, NATS, and execution pipelines" },
            { slug: "arch-frontend", label: "Frontend", path: "docs/architecture/FRONTEND.md", description: "Routes, components, Zustand, and design system" },
            { slug: "arch-operations", label: "Operations", path: "docs/architecture/OPERATIONS.md", description: "Deployment, config, testing, and CI/CD guidance" },
            { slug: "v7-architecture-prd", label: "PRD Compatibility Index", path: "architecture/mycelis-architecture-v7.md", description: "Stable compatibility entrypoint for old references; current authority points to V8.2 docs" },
        ],
    },
    {
        section: "Governance & Testing",
        docs: [
            { slug: "governance", label: "Governance System", path: "docs/governance.md", description: "Policy enforcement, approval posture, and audit-linked governance model" },
            { slug: "testing", label: "Testing", path: "docs/TESTING.md", description: "Unit, integration, browser, and release validation guidance" },
            { slug: "v8-ui-testing-agentry-contract", label: "V8 UI Testing Agentry", path: "docs/architecture-library/V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md", description: "Browser proof contract for Soma-first flows, governed actions, and trust recovery" },
            { slug: "v8-ui-team-full-test-set", label: "V8 UI Team Full Test Set", path: "docs/architecture-library/V8_UI_TEAM_FULL_TEST_SET.md", description: "Full UI validation set for browser workflows, runtime proof, and final verdict rules" },
        ],
    },
];

/** Flat lookup map: slug -> DocEntry. Used by the API route for validation. */
export const DOC_BY_SLUG: Map<string, DocEntry> = new Map(
    DOC_MANIFEST.flatMap((section) => section.docs).map((doc) => [doc.slug, doc])
);
