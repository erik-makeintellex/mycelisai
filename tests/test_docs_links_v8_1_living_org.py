from __future__ import annotations

from tests.test_docs_links import (
    DOCS_MANIFEST,
    README,
    ROOT,
    V8_1_LIVING_ARCHITECTURE,
    V8_2_FULL_ARCHITECTURE,
    V8_DEV_STATE,
    V8_UI_API_CONTRACT,
)


def test_v8_docs_cover_managed_exchange_foundation():
    expectations = {
        README: [
            "managed exchange foundation",
            "channels, threads, schemas, and normalized outputs",
        ],
        V8_1_LIVING_ARCHITECTURE: [
            "### 5.8 Managed exchange foundation",
            "named channels for work, review, learning, and normalized tool output",
            "structured threads for planning, work, review, escalation, and learning",
        ],
        V8_UI_API_CONTRACT: [
            "managed exchange items and artifact schemas",
            "inspect-only managed exchange surfaces for Channels, Threads, and Recent Artifacts",
        ],
        V8_2_FULL_ARCHITECTURE: [
            "Managed Exchange",
            "channels, typed fields, artifact schemas, and inter-agent threads become the governed exchange substrate",
        ],
    }

    missing: list[str] = []
    for path, snippets in expectations.items():
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Managed exchange foundation docs are out of sync:\n" + "\n".join(missing)


def test_execution_governance_docs_reference_v8_migration_contract():
    governance_text = (ROOT / "docs" / "architecture-library" / "DELIVERY_GOVERNANCE_AND_TESTING_V7.md").read_text(encoding="utf-8")
    team_text = (ROOT / "docs" / "architecture-library" / "TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md").read_text(encoding="utf-8")

    required = [
        (governance_text, "## V8 Migration Alignment"),
        (governance_text, "docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md"),
        (governance_text, "Template ≠ instantiated organization"),
        (team_text, "V8 Migration Workstreams"),
        (team_text, "template -> instantiation -> inheritance -> precedence"),
        (team_text, ".state/V8_DEV_STATE.md"),
    ]

    missing = []
    for text, snippet in required:
        if snippet not in text:
            missing.append(snippet)

    assert not missing, (
        "Execution/governance docs are missing V8 migration references: "
        f"{missing}"
    )


def test_v8_bundle_startup_docs_reflect_fail_closed_contract():
    surfaces = {
        README: [
            "normal startup fails closed unless a valid bootstrap bundle is present",
            "`MYCELIS_BOOTSTRAP_TEMPLATE_ID`",
        ],
        ROOT / "docs" / "LOCAL_DEV_WORKFLOW.md": [
            "Startup now instantiates the runtime organization only through a selected bundle.",
            "Core fails closed when no valid bundle exists in `core/config/templates/`",
            "`MYCELIS_BOOTSTRAP_TEMPLATE_ID`",
            "not a normal startup path",
        ],
        ROOT / "docs" / "architecture" / "OPERATIONS.md": [
            "instantiate the runtime organization from it",
            "fail closed if no valid bundle is available",
            "require `MYCELIS_BOOTSTRAP_TEMPLATE_ID` when multiple bundles are mounted",
        ],
        V8_DEV_STATE: [
            "startup now instantiates runtime organization truth directly from self-contained bundle data",
            "retired the remaining no-bundle bootstrap fallback",
            "`MYCELIS_BOOTSTRAP_TEMPLATE_ID` must be set whenever more than one bundle is present",
        ],
    }

    missing: list[str] = []
    for path, snippets in surfaces.items():
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "V8 bundle startup docs are missing required fail-closed contract snippets:\n" + "\n".join(missing)

    local_workflow_text = (ROOT / "docs" / "LOCAL_DEV_WORKFLOW.md").read_text(encoding="utf-8")
    assert "when no startup bundle is configured" not in local_workflow_text, (
        "docs/LOCAL_DEV_WORKFLOW.md still describes a no-bundle startup fallback"
    )


def test_v8_ui_api_contract_is_indexed_exposed_and_complete():
    index_text = (ROOT / "docs" / "architecture-library" / "ARCHITECTURE_LIBRARY_INDEX.md").read_text(encoding="utf-8")
    manifest_text = DOCS_MANIFEST.read_text(encoding="utf-8")
    contract_text = V8_UI_API_CONTRACT.read_text(encoding="utf-8")

    required_index_refs = [
        "V8 UI/API and Operator Experience Contract",
        "V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md",
    ]
    missing_index_refs = [snippet for snippet in required_index_refs if snippet not in index_text]
    assert not missing_index_refs, (
        "Architecture library index is missing the V8 UI/API contract references: "
        f"{missing_index_refs}"
    )

    required_manifest_refs = [
        'slug: "v8-ui-api-operator-experience-contract"',
        'path: "docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md"',
    ]
    missing_manifest_refs = [snippet for snippet in required_manifest_refs if snippet not in manifest_text]
    assert not missing_manifest_refs, (
        "docsManifest is missing the V8 UI/API contract entry: "
        f"{missing_manifest_refs}"
    )

    required_contract_snippets = [
        "The default product experience must feel like creating and operating an AI Organization.",
        "## 2. Canonical terminology",
        "AI Organization",
        "Soma",
        "Team Lead",
        "Advisors",
        "Departments",
        "Specialists",
        "AI Engine Settings",
        "Memory & Continuity",
        "### 3.3 Two explicit UX and control layers",
        "### 3.4 Source-of-truth layers",
        "### 3.5 Primary orchestration flow",
        "Default Operator Surface:",
        "Advanced Architecture / Runtime Surface:",
        "deployment/env overrides must not replace bundle-defined runtime organization truth",
        "### 4.1 First-run flow",
        "### 4.2 Create AI Organization flow",
        "### 4.3 Choose template vs empty start",
        "### 4.4 AI Organization home and header contract",
        "### 4.5 Soma-primary workspace behavior",
        "### 4.6 Advisor, Department, and Specialist visibility rules",
        "### 4.7 Advanced-mode boundaries",
        "## 5. API/UI contract mapping by screen and action",
        "The product must not open into a raw assistant conversation as its primary identity.",
        "Open Soma Workspace",
    ]
    missing_contract_snippets = [snippet for snippet in required_contract_snippets if snippet not in contract_text]
    assert not missing_contract_snippets, (
        "V8 UI/API contract is missing required operator-flow or terminology coverage: "
        f"{missing_contract_snippets}"
    )


def test_v8_runtime_contracts_cover_semantic_continuity_and_learning_layers():
    text = (ROOT / "docs" / "architecture-library" / "V8_RUNTIME_CONTRACTS.md").read_text(encoding="utf-8")
    required_snippets = [
        "Learning Loops, semantic continuity, Procedure / Skill Sets",
        "semantic continuity recall",
        "reviewed memory promotion inputs",
        "procedure and skill retrieval for type-bound specialization",
        "pgvector-backed semantic continuity substrate provides event, action, and result semantic indexing",
        "Soma Kernel interprets and orchestrates continuity using semantic recall, but it does not become the memory substrate itself",
        "Loops generate candidates, perform review, and route promotion; they never silently rewrite continuity state.",
        "raw memory",
        "reviewed memory",
        "promoted memory",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, (
        "V8 runtime contracts are missing semantic continuity or learning-layer coverage: "
        f"{missing}"
    )


def test_v8_1_living_architecture_is_indexed_exposed_and_complete():
    index_text = (ROOT / "docs" / "architecture-library" / "ARCHITECTURE_LIBRARY_INDEX.md").read_text(encoding="utf-8")
    manifest_text = DOCS_MANIFEST.read_text(encoding="utf-8")
    architecture_text = V8_1_LIVING_ARCHITECTURE.read_text(encoding="utf-8")

    required_index_refs = [
        "V8.1 Living Organization Architecture",
        "V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md",
    ]
    missing_index_refs = [snippet for snippet in required_index_refs if snippet not in index_text]
    assert not missing_index_refs, (
        "Architecture library index is missing the V8.1 living architecture references: "
        f"{missing_index_refs}"
    )

    required_manifest_refs = [
        'slug: "v8-1-living-organization-architecture"',
        'path: "docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md"',
    ]
    missing_manifest_refs = [snippet for snippet in required_manifest_refs if snippet not in manifest_text]
    assert not missing_manifest_refs, (
        "docsManifest is missing the V8.1 living architecture entry: "
        f"{missing_manifest_refs}"
    )

    required_architecture_snippets = [
        "Loop Profiles as the bounded execution layer",
        "Runtime Capabilities as the bounded action layer",
        "Learning Loops as the bounded candidate-capture and promotion-review layer",
        "Memory Promotion and Semantic Continuity as the pgvector-backed recall substrate",
        "Procedure / Skill Sets as reviewed specialist memory bound to Agent Type Profiles",
        "### 5.1 Loop Profiles",
        "### 5.2 Runtime Capabilities",
        "### 5.3 Response Contract",
        "### 5.4 Agent Type Profiles",
        "### 5.5 Memory Promotion and Semantic Continuity",
        "### 5.6 Procedure / Skill Sets",
        "### 5.7 Layering clarification",
        "Learning Loop",
        "pgvector",
        "raw memory",
        "reviewed memory",
        "promoted memory",
        "no silent self-rewrite",
        "### 8.2 Automations surface",
        "Automations",
        "Watchers",
        "Reviews",
        "### 8.3 Advanced UI boundaries",
        "organization defaults and inheritance visibility",
        "deployment/env influence",
        "### 11.2 First shippable state",
        "loops exist as configuration and inspectable architecture, not broad execution",
        "capabilities are defined but not fully exercised",
        "learning continuity architecture is defined even when raw/reviewed/promoted memory promotion is not fully implemented yet",
        "the system remains safe and inspectable",
    ]
    missing_architecture_snippets = [snippet for snippet in required_architecture_snippets if snippet not in architecture_text]
    assert not missing_architecture_snippets, (
        "V8.1 living architecture doc is missing required architecture or release-contract coverage: "
        f"{missing_architecture_snippets}"
    )
