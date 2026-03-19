from __future__ import annotations

import re
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
README = ROOT / "README.md"
DOCS_MANIFEST = ROOT / "interface" / "lib" / "docsManifest.ts"
PRD_INDEX = ROOT / "mycelis-architecture-v7.md"
NEXT_EXECUTION_SLICES = ROOT / "docs" / "architecture-library" / "NEXT_EXECUTION_SLICES_V7.md"
DEV_STATE = ROOT / "V7_DEV_STATE.md"
V8_DEV_STATE = ROOT / "V8_DEV_STATE.md"
V8_BOOTSTRAP_MODEL = ROOT / "docs" / "architecture-library" / "V8_CONFIG_AND_BOOTSTRAP_MODEL.md"
V8_UI_API_CONTRACT = ROOT / "docs" / "architecture-library" / "V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md"
V8_1_LIVING_ARCHITECTURE = ROOT / "docs" / "architecture-library" / "V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md"
SOMA_COUNCIL_PROTOCOL = ROOT / "docs" / "architecture" / "SOMA_COUNCIL_ENGAGEMENT_PROTOCOL_V7.md"
UI_OPERATOR_EXPERIENCE = ROOT / "docs" / "architecture-library" / "UI_AND_OPERATOR_EXPERIENCE_V7.md"
CANONICAL_DOCS = [
    README,
    PRD_INDEX,
    ROOT / "docs" / "TESTING.md",
    ROOT / "docs" / "architecture" / "OPERATIONS.md",
    ROOT / "ops" / "README.md",
    ROOT / "docs" / "user" / "system-status-recovery.md",
    ROOT / "docs" / "architecture" / "WORKFLOW_COMPOSER_DELIVERY_V7.md",
]


def test_readme_local_links_resolve():
    text = README.read_text(encoding="utf-8")
    links = re.findall(r"\]\(([^)]+)\)", text)
    missing: list[str] = []

    for link in links:
        if link.startswith(("http://", "https://", "mailto:", "/docs?doc=")):
            continue
        target = link.split("#", 1)[0]
        if not target or target.startswith("#"):
            continue
        if not (ROOT / target).exists():
            missing.append(link)

    assert not missing, f"README contains broken local links: {missing}"


def test_docs_manifest_paths_resolve():
    text = DOCS_MANIFEST.read_text(encoding="utf-8")
    paths = re.findall(r'path:\s*"([^"]+)"', text)
    missing = [path for path in paths if not (ROOT / path).exists()]

    assert not missing, f"docsManifest contains broken paths: {missing}"


def test_docs_manifest_exposes_required_canonical_docs():
    text = DOCS_MANIFEST.read_text(encoding="utf-8")
    required_paths = [
        "docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md",
        "docs/architecture-library/TARGET_DELIVERABLE_V7.md",
        "docs/architecture-library/SYSTEM_ARCHITECTURE_V7.md",
        "docs/architecture-library/EXECUTION_AND_MANIFEST_LIBRARY_V7.md",
        "docs/architecture-library/UI_AND_OPERATOR_EXPERIENCE_V7.md",
        "docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md",
        "docs/architecture-library/NEXT_EXECUTION_SLICES_V7.md",
        "docs/architecture-library/V8_RUNTIME_CONTRACTS.md",
        "docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md",
        "docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md",
        "docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md",
        "mycelis-architecture-v7.md",
    ]

    missing = [path for path in required_paths if path not in text]
    assert not missing, f"docsManifest is missing required canonical docs: {missing}"


def _slugify_heading(heading: str) -> str:
    slug = heading.strip().lower()
    slug = re.sub(r"[^\w\s-]", "", slug)
    slug = re.sub(r"\s+", "-", slug)
    slug = re.sub(r"-+", "-", slug)
    return slug.strip("-")


def test_readme_has_structured_toc_with_live_heading_targets():
    text = README.read_text(encoding="utf-8")
    assert "## README TOC" in text, "README must expose a structured top-level TOC"

    headings = re.findall(r"^##+\s+(.+)$", text, flags=re.MULTILINE)
    heading_slugs = {_slugify_heading(heading) for heading in headings}

    toc_match = re.search(r"^## README TOC\s*\n(.*?)(?=^## )", text, flags=re.MULTILINE | re.DOTALL)
    assert toc_match, "README TOC block must sit directly under the title section"

    toc_links = re.findall(r"\]\(#([^)]+)\)", toc_match.group(1))
    assert toc_links, "README TOC must contain anchor links"

    missing = [anchor for anchor in toc_links if anchor not in heading_slugs]
    assert not missing, f"README TOC contains anchors without matching headings: {missing}"


def test_readme_has_fresh_agent_review_sequence():
    text = README.read_text(encoding="utf-8")
    assert "## Fresh Agent Start Here" in text, "README must expose a fresh-agent review section near the top"

    required_refs = [
        "AGENTS.md",
        "docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md",
        "docs/architecture-library/TARGET_DELIVERABLE_V7.md",
        "docs/architecture-library/EXECUTION_AND_MANIFEST_LIBRARY_V7.md",
        "docs/architecture-library/UI_AND_OPERATOR_EXPERIENCE_V7.md",
        "docs/architecture/UI_TARGET_AND_TRANSACTION_CONTRACT_V7.md",
        "docs/architecture/OPERATIONS.md",
        "docs/TESTING.md",
        "V7_DEV_STATE.md",
        "interface/lib/docsManifest.ts",
        "docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md",
    ]

    missing = [ref for ref in required_refs if ref not in text]
    assert not missing, f"Fresh-agent review sequence is missing required references: {missing}"


def test_readme_has_feature_status_standard():
    text = README.read_text(encoding="utf-8")
    assert "## Feature Status Standard" in text, "README must define the canonical feature-status markers"

    required_markers = ["`REQUIRED`", "`NEXT`", "`ACTIVE`", "`IN_REVIEW`", "`COMPLETE`", "`BLOCKED`"]
    missing = [marker for marker in required_markers if marker not in text]
    assert not missing, f"Feature status standard is missing markers: {missing}"


def test_canonical_docs_do_not_ship_executable_bare_uvx_inv_examples():
    offenders: list[str] = []

    for path in CANONICAL_DOCS:
        for lineno, line in enumerate(path.read_text(encoding="utf-8").splitlines(), start=1):
            if re.match(r"^\s*uvx inv\b", line):
                offenders.append(f"{path.relative_to(ROOT)}:{lineno}:{line.strip()}")

    assert not offenders, "Canonical docs contain executable bare uvx inv examples:\n" + "\n".join(offenders)


def test_canonical_task_docs_cover_required_invoke_commands():
    expected = {
        README: [
            "uv run inv ci.entrypoint-check",
            "uv run inv lifecycle.memory-restart",
            "uv run inv team.architecture-sync",
        ],
        ROOT / "docs" / "architecture" / "OPERATIONS.md": [
            "uv run inv auth.dev-key",
            "uv run inv lifecycle.memory-restart",
            "uv run inv logging.check-schema",
            "uv run inv quality.max-lines --limit 350",
            "uv run inv ci.entrypoint-check",
            "uv run inv team.architecture-sync",
            "--live-backend",
        ],
        ROOT / "ops" / "README.md": [
            "uv run inv core.build",
            "uv run inv auth.dev-key",
            "uv run inv lifecycle.health",
            "uv run inv ci.entrypoint-check",
            "uv run inv team.architecture-sync",
        ],
    }

    missing: list[str] = []
    for path, commands in expected.items():
        text = path.read_text(encoding="utf-8")
        for command in commands:
            if command not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{command}`")

    assert not missing, "Canonical task docs are missing required invoke commands:\n" + "\n".join(missing)


def test_playwright_docs_reflect_owned_server_and_browser_matrix():
    expectations = {
        README: [
            "Playwright owns the Next.js server lifecycle",
            "chromium firefox webkit",
        ],
        ROOT / "docs" / "TESTING.md": [
            "Playwright starts/stops the Next.js server",
            "mobile-chromium",
            "@axe-core/playwright",
            "workspace-live-backend.spec.ts",
            "--live-backend",
        ],
        ROOT / "docs" / "architecture" / "OPERATIONS.md": [
            "Playwright owns Next.js server lifecycle",
            "Chromium/Firefox/WebKit + mobile smoke",
        ],
    }

    missing: list[str] = []
    for path, snippets in expectations.items():
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

        if "requires running servers" in text:
            missing.append(f"{path.relative_to(ROOT)} still says Playwright requires running servers")

    assert not missing, "Playwright docs are missing required lifecycle/browser guidance:\n" + "\n".join(missing)


def test_prd_index_points_to_modular_architecture_library():
    text = PRD_INDEX.read_text(encoding="utf-8")
    required_links = [
        "docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md",
        "docs/architecture-library/TARGET_DELIVERABLE_V7.md",
        "docs/architecture-library/SYSTEM_ARCHITECTURE_V7.md",
        "docs/architecture-library/EXECUTION_AND_MANIFEST_LIBRARY_V7.md",
        "docs/architecture-library/UI_AND_OPERATOR_EXPERIENCE_V7.md",
        "docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md",
        "docs/architecture-library/NEXT_EXECUTION_SLICES_V7.md",
    ]

    missing = [link for link in required_links if link not in text]
    assert not missing, f"PRD index is missing modular architecture links: {missing}"


def test_next_execution_slices_follow_canonical_priority_order():
    text = NEXT_EXECUTION_SLICES.read_text(encoding="utf-8")
    required_snippets = [
        "## Commitment Logic For This Queue",
        "## Slice 1: Launch Crew And Workflow Onboarding",
        "## Slice 2: P1 Logging, Error Handling, And Execution Feedback",
        "## Slice 3: Prime-Development Reply Reliability",
        "## Slice 4: P1 Hot-Path Cleanup",
        "## Slice 5: Manifest Pipeline Preparation",
        "1. operator-facing execution clarity",
        "2. operator-facing error and recovery clarity",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, f"Next execution slices doc is missing required queue structure: {missing}"


def test_dev_state_uses_delivery_program_snapshot():
    text = DEV_STATE.read_text(encoding="utf-8")
    required_snippets = [
        "### Delivery Program Snapshot",
        "P0  Operational foundation and gate discipline",
        "P1  Logging, error handling, and hot-path cleanup",
        "### Active Queue In Canonical Order",
        "Slice 1  Launch Crew and workflow onboarding execution contract",
        "Slice 2  P1 logging, error handling, and execution feedback",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, f"Dev state is missing delivery-program framing: {missing}"


def test_ui_operator_experience_covers_direct_first_and_theme_simplification_contract():
    text = UI_OPERATOR_EXPERIENCE.read_text(encoding="utf-8")
    required_snippets = [
        "### 3.5 Soma-First conversation economy rule",
        "Workspace chat should default to Soma-only execution for normal interaction.",
        "### 3.6 Theme simplification rule",
        "diagnostics are progressive-disclosure, not always-on density",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, (
        "UI/operator experience doc is missing direct-first or theme-simplification contract snippets: "
        f"{missing}"
    )


def test_soma_council_protocol_defines_consultation_trigger_modes():
    text = SOMA_COUNCIL_PROTOCOL.read_text(encoding="utf-8")
    required_snippets = [
        "### 2.1 Consultation trigger policy (token economy)",
        "Council consultation is required only when at least one condition is true:",
        "- `none`: no council call, Soma responds directly",
        "- `targeted`: one specialist selected by need",
        "- `full_council`: only for explicitly broad architectural/program decisions",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, f"Soma-council protocol is missing consultation trigger contract snippets: {missing}"


def test_dev_state_tracks_active_slice2_theme_and_routing_priorities():
    text = DEV_STATE.read_text(encoding="utf-8")
    required_snippets = [
        "1. `ACTIVE` Slice 2 sub-track: Workspace theme simplification and conversation-first hierarchy",
        "2. `ACTIVE` Slice 2 sub-track: Soma direct-first routing economy and consultation trigger discipline",
        "Where theme/routing are now encoded in architecture planning:",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, f"Dev state is missing active Slice 2 theme/routing priorities: {missing}"


def test_canonical_surfaces_do_not_expose_purged_planning_docs():
    purged_paths = [
        "docs/UI_FRAMEWORK_V7.md",
        "docs/UI_ELEMENTS_PLANNING_V7.md",
        "docs/ui-delivery/PARALLEL_DELIVERY_BOARD.md",
        "docs/ui-delivery/TEAM_ABCQ_EXECUTION_BOARD.md",
        "docs/product/UI_WORKFLOW_INSTANTIATION_AND_BUS_PLAN_V7.md",
        "docs/product/SOMA_EXTENSION_OF_SELF_PRD_V7.md",
    ]

    surfaces = {
        README: README.read_text(encoding="utf-8"),
        DOCS_MANIFEST: DOCS_MANIFEST.read_text(encoding="utf-8"),
        DEV_STATE: DEV_STATE.read_text(encoding="utf-8"),
        ROOT / "docs" / "README.md": (ROOT / "docs" / "README.md").read_text(encoding="utf-8"),
    }

    offenders: list[str] = []
    for path, text in surfaces.items():
        for purged in purged_paths:
            if purged in text:
                offenders.append(f"{path.relative_to(ROOT)} still references `{purged}`")

    assert not offenders, "Canonical surfaces still reference purged planning docs:\n" + "\n".join(offenders)


def test_v8_bootstrap_model_template_section_defines_blueprint_contract():
    text = V8_BOOTSTRAP_MODEL.read_text(encoding="utf-8")
    required_snippets = [
        "## Template and instantiation entry points",
        "Template = reusable blueprint",
        "Inception / AI Organization = actual instantiated organization",
        "Templates are reusable organization blueprints, not just UI presets.",
        "- organization type",
        "- default Team Lead / Soma Kernel posture",
        "- default Advisors / Council composition",
        "- default Departments / Teams",
        "- default Specialists / Agents",
        "- default AI Engine Settings / provider policy",
        "- default Memory & Personality settings",
        "- optional beginner-facing labels and descriptions",
        "By default, a template does not contain:",
        "- live runtime state",
        "- execution history",
        "- per-run outcomes",
        "- user-specific secrets",
        "- starter templates for beginners",
        "- domain templates such as Research, Engineering, and Marketing",
        "- executive templates such as CTO, COO, and Product Lead",
        "- personal / continuity templates",
        "- empty / minimal template",
        "### Template validation expectations",
        "Templates should be treated as governed blueprint inputs, not free-form UI metadata blobs.",
        "- declare enough structure to identify its organization type and intended operating posture",
        "- organize defaults along the same conceptual scopes used by bootstrap resolution",
        "- separate beginner-facing labels from runtime-shaping defaults",
        "- be resolvable into an instantiated organization without inventing a second hidden template model",
        "1. create from template",
        "2. create empty",
        "3. create from config/API",
        "4. clone and modify an existing template later",
        "1. the template supplies defaults",
        "2. the instantiated organization becomes its own object after creation",
        "3. later edits to the template do not silently rewrite existing organizations unless that behavior is explicitly designed and governed",
        "A template can:",
        "- define team defaults",
        "- define agent defaults",
        "- define optional advanced overrides",
        "Beginner UI should mainly show:",
        "- template name",
        "- purpose",
        "- a simple summary",
        "Advanced panels may expose:",
        "- internal structure",
        "- routing posture",
        "- scoped defaults",
        "### Target-delivery implication",
        "Templates must support target delivery rather than acting as decorative presets.",
        "- templates are bootstrap inputs for real organizations that should be capable of participating in the governed execution platform",
        "- instantiated organizations created from templates must still resolve into runtime behavior that can produce target product outcomes such as `answer`, `proposal`, `execution_result`, and `blocker`",
        "- the template system should not become a second planning-only layer that is disconnected from execution, governance, or operator-visible delivery behavior",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, (
        "V8 bootstrap model is missing required template/instantiation contract snippets: "
        f"{missing}"
    )


def test_v8_dev_state_tracks_bootstrap_completion_and_validation_pass():
    text = V8_DEV_STATE.read_text(encoding="utf-8")
    required_snippets = [
        "### 17. Template and instantiation entry points definition",
        "the `Migration from V7 bootstrap assumptions` section now explains how fixed V7 startup assumptions collapse into explicit V8 configuration sources",
        "Task 004  Config and bootstrap model planning                       [COMPLETE]",
        "Task 008  Planning-integration validation pass                      [COMPLETE]",
        "Task 009  Next-execution/governance guidance migration              [NEXT]",
        "`COMPLETE` run the planning-integration validation pass so README, the architecture-library index, docs manifests, and doc-tests all confirm the new V7-to-V8 bootstrap migration contract.",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, (
        "V8 dev state is missing the completed bootstrap plan or the new validation-pass next step: "
        f"{missing}"
    )


def test_v8_bootstrap_model_keeps_template_contract_linked_to_target_delivery():
    bootstrap_text = V8_BOOTSTRAP_MODEL.read_text(encoding="utf-8")
    target_text = (ROOT / "docs" / "architecture-library" / "TARGET_DELIVERABLE_V7.md").read_text(encoding="utf-8")

    required_target_outcomes = ["`answer`", "`proposal`", "`execution_result`", "`blocker`"]
    missing_target_outcomes = [marker for marker in required_target_outcomes if marker not in target_text]
    assert not missing_target_outcomes, (
        "Target deliverable doc is missing required product-outcome markers: "
        f"{missing_target_outcomes}"
    )

    required_bootstrap_linkage = [
        "Templates must support target delivery rather than acting as decorative presets.",
        "`answer`, `proposal`, `execution_result`, and `blocker`",
    ]
    missing_bootstrap_linkage = [snippet for snippet in required_bootstrap_linkage if snippet not in bootstrap_text]
    assert not missing_bootstrap_linkage, (
        "V8 bootstrap model is missing required target-delivery linkage for templates: "
        f"{missing_bootstrap_linkage}"
    )


def test_v8_bootstrap_model_defines_v7_to_v8_migration_contract():
    text = V8_BOOTSTRAP_MODEL.read_text(encoding="utf-8")
    required_snippets = [
        "## Migration from V7 bootstrap assumptions",
        "canonical V7->V8 migration contract",
        "V7 never exposed a single declarative bootstrap contract.",
        "**YAML manifests and sidecar config files**",
        "**Runtime configuration** (env vars, `.env`, CLI arguments)",
        "**Database state** produced by ad hoc migrations or interactive setup scripts",
        "**Operator flows** (UI wizards, CLI prompts) that mutated standing-team rows directly",
        "standing team definitions were hydrated automatically at process start if the database was empty",
        "Soma + Council roles were fixed to a single canonical lineup",
        "runtime state (runs, manifests, NATS registrations) doubled as bootstrap inputs",
        "1. **Templates** provide reusable blueprints but remain separate from instantiated organizations (`Template ≠ instantiated organization`).",
        "2. **Declarative configuration artifacts** (files, APIs, automation payloads) describe organization inputs",
        "3. **Operator flows** submit explicit organization creation or update intents that feed the same instantiation pipeline",
        "4. **Runtime/persistent state** supplies lineage and continuity only after an organization exists",
        "V7 YAML and manifest assets remain valid migration inputs, but they must be translated into V8 template/config packages before use.",
        "Auto-hydrating standing-team rows at process start.",
        "Treating last-run database state as the bootstrap plan.",
        "V8 keeps prior assets useful, but only after they conform to the explicit template + instantiation + inheritance + precedence pipeline.",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, (
        "V8 bootstrap model migration section is missing required V7-to-V8 contract details: "
        f"{missing}"
    )


def test_execution_governance_docs_reference_v8_migration_contract():
    next_text = (ROOT / "docs" / "architecture-library" / "NEXT_EXECUTION_SLICES_V7.md").read_text(encoding="utf-8")
    governance_text = (ROOT / "docs" / "architecture-library" / "DELIVERY_GOVERNANCE_AND_TESTING_V7.md").read_text(encoding="utf-8")
    team_text = (ROOT / "docs" / "architecture-library" / "TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md").read_text(encoding="utf-8")

    required = [
        (next_text, "## V8 Migration Alignment"),
        (next_text, "template -> instantiation -> inheritance -> precedence"),
        (next_text, "V8_DEV_STATE.md"),
        (governance_text, "## V8 Migration Alignment"),
        (governance_text, "docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md"),
        (governance_text, "Template ≠ instantiated organization"),
        (team_text, "V8 Migration Workstreams"),
        (team_text, "template -> instantiation -> inheritance -> precedence"),
        (team_text, "V8_DEV_STATE.md"),
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
        "Team Lead",
        "Advisors",
        "Departments",
        "Specialists",
        "AI Engine Settings",
        "Memory & Personality",
        "### 4.1 First-run flow",
        "### 4.2 Create AI Organization flow",
        "### 4.3 Choose template vs empty start",
        "### 4.4 AI Organization home and header contract",
        "### 4.5 Team Lead-first workspace behavior",
        "### 4.6 Advisor, Department, and Specialist visibility rules",
        "### 4.7 Advanced-mode boundaries",
        "## 5. API/UI contract mapping by screen and action",
        "The product must not open into a raw assistant conversation as its primary identity.",
        "Open Team Lead Workspace",
    ]
    missing_contract_snippets = [snippet for snippet in required_contract_snippets if snippet not in contract_text]
    assert not missing_contract_snippets, (
        "V8 UI/API contract is missing required operator-flow or terminology coverage: "
        f"{missing_contract_snippets}"
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
        "### 5.1 Loop Profiles",
        "### 5.2 Runtime Capabilities",
        "### 5.3 Response Contract",
        "### 5.4 Agent Type Profiles",
        "### 8.2 Automations surface",
        "Automations",
        "Watchers",
        "Reviews",
        "### 11.2 First shippable state",
        "loops exist as configuration and inspectable architecture, not broad execution",
        "capabilities are defined but not fully exercised",
        "the system remains safe and inspectable",
    ]
    missing_architecture_snippets = [snippet for snippet in required_architecture_snippets if snippet not in architecture_text]
    assert not missing_architecture_snippets, (
        "V8.1 living architecture doc is missing required architecture or release-contract coverage: "
        f"{missing_architecture_snippets}"
    )
