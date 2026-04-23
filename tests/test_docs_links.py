from __future__ import annotations

import re
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
README = ROOT / "README.md"
DOCS_MANIFEST = ROOT / "interface" / "lib" / "docsManifest.ts"
PRD_INDEX = ROOT / "mycelis-architecture-v7.md"
LEGACY_V7_DEV_STATE = ROOT / "V7_DEV_STATE.md"
V8_DEV_STATE = ROOT / "V8_DEV_STATE.md"
V8_BOOTSTRAP_MODEL = ROOT / "docs" / "architecture-library" / "V8_CONFIG_AND_BOOTSTRAP_MODEL.md"
V8_UI_API_CONTRACT = ROOT / "docs" / "architecture-library" / "V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md"
V8_1_LIVING_ARCHITECTURE = ROOT / "docs" / "architecture-library" / "V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md"
V8_2_FULL_ARCHITECTURE = ROOT / "v8-2.md"
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


def test_active_docs_expose_project_readme_navigation():
    manifest_text = DOCS_MANIFEST.read_text(encoding="utf-8")
    manifest_paths = re.findall(r'path:\s*"([^"]+\.md)"', manifest_text)
    required_paths = sorted(set(manifest_paths + ["docs/README.md", "ops/README.md"]))

    missing_navigation: list[str] = []
    for relative_path in required_paths:
        if relative_path == "README.md":
            continue
        path = ROOT / relative_path
        text = path.read_text(encoding="utf-8")
        top = "\n".join(text.splitlines()[:8])
        if "Project README" not in top or "Navigation:" not in top:
            missing_navigation.append(relative_path)

    assert not missing_navigation, (
        "Active docs are missing return navigation to the project README: "
        f"{missing_navigation}"
    )


def test_docs_manifest_exposes_required_canonical_docs():
    text = DOCS_MANIFEST.read_text(encoding="utf-8")
    required_paths = [
        "docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md",
        "docs/architecture-library/TARGET_DELIVERABLE_V7.md",
        "docs/architecture-library/SYSTEM_ARCHITECTURE_V7.md",
        "docs/architecture-library/EXECUTION_AND_MANIFEST_LIBRARY_V7.md",
        "docs/architecture-library/UI_AND_OPERATOR_EXPERIENCE_V7.md",
        "docs/architecture-library/DELIVERY_GOVERNANCE_AND_TESTING_V7.md",
        "docs/architecture-library/TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md",
        "docs/architecture-library/V8_RUNTIME_CONTRACTS.md",
        "docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md",
        "docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md",
        "docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md",
        "v8-2.md",
        "mycelis-architecture-v7.md",
    ]

    missing = [path for path in required_paths if path not in text]
    assert not missing, f"docsManifest is missing required canonical docs: {missing}"


def test_superseded_root_v8_1_draft_is_archived_not_loose():
    root_draft = ROOT / "v8-1.md"
    archived_draft = ROOT / "docs" / "archive" / "drafts" / "v8-1.md"

    assert not root_draft.exists(), "Superseded v8-1.md must not remain loose at repo root"
    assert archived_draft.exists(), "Superseded v8-1.md must be archived under docs/archive/drafts/"

    archived_text = archived_draft.read_text(encoding="utf-8")
    required_snippets = [
        "# Superseded Draft: V8.1 PRD",
        "Superseded By: `docs/architecture-library/V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md`",
        "This file is preserved only as historical drafting context. It is not active implementation authority.",
    ]
    missing = [snippet for snippet in required_snippets if snippet not in archived_text]
    assert not missing, f"Archived v8-1 draft is missing superseded markers: {missing}"


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


def test_readme_style_pages_expose_tocs():
    required_tocs = {
        README: "## README TOC",
        ROOT / "docs" / "README.md": "## Docs TOC",
        ROOT / "ops" / "README.md": "## TOC",
        ROOT / "docs" / "archive" / "README.md": "## TOC",
        ROOT / "core" / "README.md": "## TOC",
        ROOT / "interface" / "README.md": "## TOC",
        ROOT / "cli" / "README.md": "## TOC",
        ROOT / "core" / "internal" / "registry" / "README.md": "## TOC",
        PRD_INDEX: "## TOC",
        ROOT / "docs" / "architecture-library" / "ARCHITECTURE_LIBRARY_INDEX.md": "## TOC",
    }

    missing: list[str] = []
    for path, heading in required_tocs.items():
        text = path.read_text(encoding="utf-8")
        if heading not in text:
            missing.append(f"{path.relative_to(ROOT)} missing `{heading}`")

    assert not missing, "README-style pages are missing TOC blocks:\n" + "\n".join(missing)


def test_repo_readmes_expose_project_navigation():
    readmes = [
        ROOT / "docs" / "README.md",
        ROOT / "ops" / "README.md",
        ROOT / "docs" / "archive" / "README.md",
        ROOT / "core" / "README.md",
        ROOT / "interface" / "README.md",
        ROOT / "cli" / "README.md",
        ROOT / "core" / "internal" / "registry" / "README.md",
        PRD_INDEX,
        ROOT / "docs" / "architecture-library" / "ARCHITECTURE_LIBRARY_INDEX.md",
    ]

    missing: list[str] = []
    for path in readmes:
        text = path.read_text(encoding="utf-8")
        top = "\n".join(text.splitlines()[:8])
        if "Project README" not in top or "Navigation:" not in top:
            missing.append(str(path.relative_to(ROOT)))

    assert not missing, "README-style pages are missing project navigation:\n" + "\n".join(missing)


def test_archive_readme_keeps_v7_material_as_migration_input_not_active_authority():
    text = (ROOT / "docs" / "archive" / "README.md").read_text(encoding="utf-8")
    active_section_match = re.search(
        r"## Active Implementation Sources\s*\n(.*?)(?=^## )",
        text,
        flags=re.MULTILINE | re.DOTALL,
    )
    assert active_section_match, "Archive README must keep an explicit active-source pointer section"

    active_section = active_section_match.group(1)
    assert "V8_DEV_STATE.md" in active_section
    assert "V8_RUNTIME_CONTRACTS.md" in active_section
    assert "_V7.md" not in active_section and "V7_DEV_STATE.md" not in active_section, (
        "docs/archive/README.md must not elevate V7 docs as active implementation sources"
    )

    required_migration_input_snippets = [
        "## Migration Inputs",
        "V7-labeled architecture-library docs are historical migration inputs, not active implementation authorities",
    ]
    missing = [snippet for snippet in required_migration_input_snippets if snippet not in text]
    assert not missing, "Archive README is missing V7 migration-input framing:\n" + "\n".join(missing)


def test_interface_readme_names_aero_light_as_default_theme_not_midnight():
    text = (ROOT / "interface" / "README.md").read_text(encoding="utf-8")

    required_snippets = [
        "Utility-first aero-light default theme, midnight alternate, custom animations",
        "The default product theme is `aero-light`; `midnight-cortex` remains an alternate.",
    ]
    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, "Interface README is missing current theme-default language:\n" + "\n".join(missing)

    forbidden_snippets = [
        "Utility-first midnight theme",
        "Midnight Cortex palette",
    ]
    present = [snippet for snippet in forbidden_snippets if snippet in text]
    assert not present, "Interface README still describes midnight as the default theme:\n" + "\n".join(present)


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


def test_docs_readme_promotes_v8_navigation_before_legacy_v7_inputs():
    text = (ROOT / "docs" / "README.md").read_text(encoding="utf-8")

    required_v8_refs = [
        "**Active Development State**: `../V8_DEV_STATE.md`",
        "**V8 Runtime Contracts**: `./architecture-library/V8_RUNTIME_CONTRACTS.md`",
        "**V8 Config and Bootstrap Model**: `./architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`",
        "**V8 UI/API and Operator Experience Contract**: `./architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md`",
    ]
    missing = [ref for ref in required_v8_refs if ref not in text]
    assert not missing, f"docs/README is missing required V8 navigation references: {missing}"

    assert "Migration inputs and historical references:" in text
    assert text.index("**Active Development State**: `../V8_DEV_STATE.md`") < text.index("**Legacy V7 Development State**: `../V7_DEV_STATE.md`")


def test_readme_has_feature_status_standard():
    text = README.read_text(encoding="utf-8")
    assert "## Feature Status Standard" in text, "README must define the canonical feature-status markers"

    required_markers = ["`REQUIRED`", "`NEXT`", "`ACTIVE`", "`IN_REVIEW`", "`COMPLETE`", "`BLOCKED`"]
    missing = [marker for marker in required_markers if marker not in text]
    assert not missing, f"Feature status standard is missing markers: {missing}"


def test_readme_exposes_layered_architecture_truth():
    text = README.read_text(encoding="utf-8")

    required_sections = [
        "## Final Production Architecture (V8.2)",
        "## Current Release Target (V8.1)",
        "## Current Implementation State",
    ]
    missing_sections = [section for section in required_sections if section not in text]
    assert not missing_sections, f"README is missing required layered-truth sections: {missing_sections}"

    required_snippets = [
        "v8-2.md",
        "full actuation architecture",
        "V8.1 is the current release target",
        "V8.2 is the full production target",
        "Actual implementation state lives in [V8_DEV_STATE.md](V8_DEV_STATE.md).",
    ]
    missing_snippets = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing_snippets, f"README is missing required layered-truth snippets: {missing_snippets}"


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


def test_v8_docs_cover_managed_exchange_security_foundation():
    expectations = {
        README: [
            "managed exchange is permissioned",
            "normalization into managed exchange does not imply unrestricted trust",
            "free-node release now includes foundational security boundaries",
        ],
        V8_1_LIVING_ARCHITECTURE: [
            "channels, threads, and exchange items carry explicit readers, writers, reviewers, participants, sensitivity classes, and downstream allowed-consumer metadata",
            "capability-producing outputs carry a capability id, risk class, trust class, and audit-ready publication metadata",
            "normalization into exchange does not imply unrestricted trust",
        ],
        V8_UI_API_CONTRACT: [
            "inspect-only trust, sensitivity, and review labels for managed exchange items where they matter operationally",
            "managed exchange security labels stay inspectable in advanced surfaces without leaking forbidden/internal-only security implementation detail into the default Soma-first UX",
        ],
        V8_2_FULL_ARCHITECTURE: [
            "managed exchange remains permissioned",
            "normalization into exchange does not imply unrestricted trust",
        ],
        V8_DEV_STATE: [
            "managed exchange security foundation now exists",
            "enterprise identity, approval workflows, and multi-user access management remain deferred",
        ],
    }

    missing: list[str] = []
    for path, snippets in expectations.items():
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Managed exchange security foundation docs are out of sync:\n" + "\n".join(missing)


def test_readme_defines_default_and_advanced_surface_contract():
    text = README.read_text(encoding="utf-8")

    required_sections = [
        "## Default And Advanced Surfaces",
    ]
    missing_sections = [section for section in required_sections if section not in text]
    assert not missing_sections, f"README is missing the default/advanced surface section: {missing_sections}"

    required_snippets = [
        "Default Operator Surface:",
        "Advanced Architecture / Runtime Surface:",
        "source-of-truth layers remain separate:",
        "the advanced architecture/runtime surface is now defined as a contract, but it is not fully implemented yet",
        "the default UX must stay simple and intent-first",
        "the advanced architecture/runtime surface must stay separate, make inheritance legible, and make config origin legible",
    ]
    missing_snippets = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing_snippets, (
        "README is missing required default-vs-advanced surface guidance: "
        f"{missing_snippets}"
    )


def test_readme_declares_development_contract():
    text = README.read_text(encoding="utf-8")

    required_sections = [
        "## Development Contract",
    ]
    missing_sections = [section for section in required_sections if section not in text]
    assert not missing_sections, f"README is missing the development contract section: {missing_sections}"

    required_snippets = [
        "A slice is not complete unless:",
        "- tests pass",
        "- documentation is updated where meaning changed",
        "- architecture alignment is verified across the layered truth surfaces",
        "`README.md` is the primary architecture inception document",
        "`v8-2.md` is the canonical full architecture",
        "`V8_DEV_STATE.md` is the source of actual implementation truth",
        "all slices must update these surfaces when implementation, release posture, or target meaning changes",
        "end-of-slice reporting must explicitly state which tests ran, which docs changed, and which scoped docs were reviewed but left unchanged",
    ]
    missing_snippets = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing_snippets, f"README is missing required development-contract guidance: {missing_snippets}"


def test_docs_cover_env_override_contract_for_deployment_automation():
    expectations = {
        README: [
            "MYCELIS_PROVIDER_<PROVIDER_ID>_*",
            "MYCELIS_PROFILE_<PROFILE>_PROVIDER",
            "MYCELIS_MEDIA_*",
            "retired `MYCELIS_TEAM_PROVIDER_MAP` / `MYCELIS_AGENT_PROVIDER_MAP`",
            "Bundle -> Instantiated Organization -> Inheritance -> Routing",
        ],
        ROOT / "docs" / "LOCAL_DEV_WORKFLOW.md": [
            "MYCELIS_PROVIDER_<PROVIDER_ID>_MODEL_ID",
            "MYCELIS_PROVIDER_<PROVIDER_ID>_ENDPOINT",
            "MYCELIS_PROFILE_<PROFILE>_PROVIDER",
            "MYCELIS_MEDIA_ENDPOINT",
            "retired `MYCELIS_TEAM_PROVIDER_MAP` / `MYCELIS_AGENT_PROVIDER_MAP`",
            "Bundle -> Instantiated Organization -> Inheritance -> Routing",
        ],
        ROOT / "docs" / "architecture" / "OPERATIONS.md": [
            "MYCELIS_PROVIDER_<PROVIDER_ID>_MODEL_ID",
            "MYCELIS_PROVIDER_<PROVIDER_ID>_ENDPOINT",
            "MYCELIS_PROFILE_<PROFILE>_PROVIDER",
            "MYCELIS_MEDIA_MODEL_ID",
            "retired `MYCELIS_TEAM_PROVIDER_MAP` / `MYCELIS_AGENT_PROVIDER_MAP`",
            "Bundle -> Instantiated Organization -> Inheritance -> Routing",
        ],
    }

    missing: list[str] = []
    for path, snippets in expectations.items():
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Docs are missing the deployment env-override contract:\n" + "\n".join(missing)


def test_docs_keep_env_overrides_separate_from_runtime_organization_truth():
    expectations = {
        README: [
            "env overrides are deployment-time infrastructure wiring, not runtime organization behavior",
            "do not replace bundle-defined runtime organization truth",
        ],
        ROOT / "docs" / "LOCAL_DEV_WORKFLOW.md": [
            "These env overrides are for deployment-time provider definition",
            "They are not runtime organization behavior",
            "do not replace bundle-defined truth",
        ],
        ROOT / "docs" / "architecture" / "OPERATIONS.md": [
            "deployment-time infrastructure configuration only",
            "do not treat env overrides as runtime organization behavior",
        ],
    }

    missing: list[str] = []
    for path, snippets in expectations.items():
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Docs are missing the env/runtime architecture boundary:\n" + "\n".join(missing)


def test_docs_cover_deployment_guidance_across_host_architectures():
    expectations = {
        README: [
            "Deployment guidance by host architecture:",
            "Windows x86_64",
            "Linux x86_64",
            "Linux arm64",
            "Mixed-architecture deployments",
        ],
        ROOT / "docs" / "LOCAL_DEV_WORKFLOW.md": [
            "## Deployment Guidance By Host Architecture",
            "Windows x86_64",
            "Linux x86_64",
            "Linux arm64",
            "Mixed-architecture deployments",
        ],
        ROOT / "docs" / "architecture" / "OPERATIONS.md": [
            "### Deployment Guidance Across Host Architectures",
            "Windows x86_64",
            "Linux x86_64",
            "Linux arm64",
            "Mixed-architecture deployment rule",
        ],
    }

    missing: list[str] = []
    for path, snippets in expectations.items():
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Docs are missing deployment guidance across host architectures:\n" + "\n".join(missing)


def test_docs_cover_core_runtime_config_mount_alignment():
    expectations = {
        README: [
            "deployed Core image resolves runtime config from `/core/config`",
        ],
        ROOT / "docs" / "LOCAL_DEV_WORKFLOW.md": [
            "deployed Core image resolves those files from `/core/config`",
            "Helm chart mounts the runtime config volume there",
        ],
        ROOT / "docs" / "architecture" / "OPERATIONS.md": [
            "mounts this runtime config tree at `/core/config`",
        ],
        ROOT / "ops" / "README.md": [
            "startup config from `/core/config`",
        ],
    }

    missing: list[str] = []
    for path, snippets in expectations.items():
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Docs are missing the Core runtime config mount alignment:\n" + "\n".join(missing)


def test_docs_cover_lifecycle_up_database_bootstrap_alignment():
    expectations = {
        ROOT / "docs" / "TESTING.md": [
            "ensures the `cortex` database exists",
            "managed built server",
        ],
        ROOT / "docs" / "architecture" / "OPERATIONS.md": [
            "ensures the `cortex` database exists before Core starts",
        ],
        ROOT / "ops" / "README.md": [
            "ensures the `cortex` database exists before Core starts",
        ],
    }

    missing: list[str] = []
    for path, snippets in expectations.items():
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Docs are missing lifecycle/database bootstrap alignment:\n" + "\n".join(missing)


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
            "uv run inv ci.service-check",
            "uv run inv lifecycle.memory-restart",
            "uv run inv team.architecture-sync",
        ],
        ROOT / "docs" / "architecture" / "OPERATIONS.md": [
            "uv run inv auth.dev-key",
            "uv run inv auth.break-glass-key",
            "uv run inv auth.posture",
            "uv run inv core.compile",
            "uv run inv interface.typecheck",
            "uv run inv lifecycle.memory-restart",
            "uv run inv logging.check-schema",
            "uv run inv quality.max-lines --limit 350",
            "uv run inv ci.service-check",
            "uv run inv ci.entrypoint-check",
            "uv run inv team.architecture-sync",
            "--live-backend",
        ],
        ROOT / "ops" / "README.md": [
            "uv run inv core.build",
            "uv run inv core.compile",
            "uv run inv interface.typecheck",
            "uv run inv auth.dev-key",
            "uv run inv auth.break-glass-key",
            "uv run inv auth.posture",
            "uv run inv lifecycle.health",
            "uv run inv ci.service-check",
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
            "owns the local Next.js server lifecycle",
            "built production Interface server path",
            "chromium firefox webkit",
            "uv run inv ci.baseline` now includes Playwright by default",
            "uv run inv ci.service-check",
        ],
        ROOT / "docs" / "TESTING.md": [
            "starts/stops the managed Next.js app",
            "built production Interface server path",
            "mobile-chromium",
            "@axe-core/playwright",
            "workspace-live-backend.spec.ts",
            "--live-backend",
            "uv run inv ci.baseline --no-e2e",
            "uv run inv ci.service-check",
        ],
        ROOT / "docs" / "architecture" / "OPERATIONS.md": [
            "Invoke manages the Next.js server lifecycle",
            "built production Interface server path",
            "Chromium/Firefox/WebKit + mobile smoke",
            "uv run inv ci.service-check",
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
        "docs/architecture-library/TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md",
    ]

    missing = [link for link in required_links if link not in text]
    assert not missing, f"PRD index is missing modular architecture links: {missing}"


def test_architecture_library_index_omits_temporary_plan_docs():
    text = (ROOT / "docs" / "architecture-library" / "ARCHITECTURE_LIBRARY_INDEX.md").read_text(encoding="utf-8")
    forbidden_refs = [
        "NEXT_EXECUTION_SLICES_V7.md",
        "MVP_RELEASE_STRIKE_TEAM_PLAN_V7.md",
        "MVP_INTEGRATION_AND_TOOLSHIP_EXECUTION_PLAN_V7.md",
        "UI_GENERATION_AND_TESTING_EXECUTION_PLAN_V7.md",
        "V8_FULL_TESTING_ACTION_PLAN.md",
        "V8_UI_TESTING_AGENTRY_EXECUTION_RUNBOOK.md",
        "V8_UI_WORKFLOW_VERIFICATION_PLAN.md",
        "V8_STRUCTURED_TEAM_ASKS_AND_LANE_ROUTING_PLAN.md",
        "V8_GOVERNED_DEPLOYMENT_CONTEXT_AND_RAG_PLAN.md",
        "V8_RELEASE_PLATFORM_REVIEW_SECURITY_MONITORING_DEBUG.md",
    ]

    present = [ref for ref in forbidden_refs if ref in text]
    assert not present, f"Architecture library index still references temporary plan docs: {present}"


def test_legacy_v7_dev_state_uses_delivery_program_snapshot():
    text = LEGACY_V7_DEV_STATE.read_text(encoding="utf-8")
    required_snippets = [
        "### Delivery Program Snapshot",
        "P0  Operational foundation and gate discipline",
        "P1  Logging, error handling, and hot-path cleanup",
        "### Active Queue In Canonical Order",
        "Slice 1  Launch Crew and workflow onboarding execution contract",
        "Slice 2  P1 logging, error handling, and execution feedback",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, f"Legacy V7 dev state is missing delivery-program framing: {missing}"


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


def test_legacy_v7_dev_state_tracks_active_slice2_theme_and_routing_priorities():
    text = LEGACY_V7_DEV_STATE.read_text(encoding="utf-8")
    required_snippets = [
        "1. `ACTIVE` Slice 2 sub-track: Workspace theme simplification and conversation-first hierarchy",
        "2. `ACTIVE` Slice 2 sub-track: Soma direct-first routing economy and consultation trigger discipline",
        "Where theme/routing are now encoded in architecture planning:",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, f"Legacy V7 dev state is missing active Slice 2 theme/routing priorities: {missing}"


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
        LEGACY_V7_DEV_STATE: LEGACY_V7_DEV_STATE.read_text(encoding="utf-8"),
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
    governance_text = (ROOT / "docs" / "architecture-library" / "DELIVERY_GOVERNANCE_AND_TESTING_V7.md").read_text(encoding="utf-8")
    team_text = (ROOT / "docs" / "architecture-library" / "TEAM_EXECUTION_AND_GLOBAL_STATE_PROTOCOL_V7.md").read_text(encoding="utf-8")

    required = [
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


def test_v8_2_full_architecture_is_indexed_exposed_and_canonical():
    index_text = (ROOT / "docs" / "architecture-library" / "ARCHITECTURE_LIBRARY_INDEX.md").read_text(encoding="utf-8")
    manifest_text = DOCS_MANIFEST.read_text(encoding="utf-8")
    architecture_text = V8_2_FULL_ARCHITECTURE.read_text(encoding="utf-8")

    required_index_refs = [
        "V8.2 Full Production Architecture",
        "../../v8-2.md",
        "Full Production Architecture (Canonical Target)",
    ]
    missing_index_refs = [snippet for snippet in required_index_refs if snippet not in index_text]
    assert not missing_index_refs, (
        "Architecture library index is missing the V8.2 full-architecture references: "
        f"{missing_index_refs}"
    )

    required_manifest_refs = [
        'slug: "v8-2-full-production-architecture"',
        'label: "Full Architecture (V8.2)"',
        'path: "v8-2.md"',
    ]
    missing_manifest_refs = [snippet for snippet in required_manifest_refs if snippet not in manifest_text]
    assert not missing_manifest_refs, (
        "docsManifest is missing the V8.2 full-architecture entry: "
        f"{missing_manifest_refs}"
    )

    required_architecture_snippets = [
        "canonical full production architecture",
        "distributed, learning, capability-enabled, and acting system",
        "distributed execution target state",
        "learning, capability, and actuation target state",
        "Not all elements described here are implemented yet.",
        "V8.1 remains the current bounded release target",
        "default operator experience remains Soma-primary, intent-first, and simple by default",
        "advanced architecture/runtime control exists as a separate non-default layer for deep users",
    ]
    missing_architecture_snippets = [snippet for snippet in required_architecture_snippets if snippet not in architecture_text]
    assert not missing_architecture_snippets, (
        "V8.2 full architecture doc is missing required canonical-target coverage: "
        f"{missing_architecture_snippets}"
    )


def test_layered_docs_align_on_advanced_surface_contract():
    surfaces = {
        README: [
            "Advanced Architecture / Runtime Surface:",
            "source-of-truth layers remain separate:",
        ],
        V8_UI_API_CONTRACT: [
            "Advanced Architecture / Runtime Surface:",
            "The product must keep these layers distinct:",
            "advanced UI may explain deployment/env influence, but it must not become a second source of runtime truth",
        ],
        V8_1_LIVING_ARCHITECTURE: [
            "### 8.3 Advanced UI boundaries",
            "organization defaults and inheritance visibility",
            "sensitive deployment secrets stay file/env/config driven",
        ],
        V8_2_FULL_ARCHITECTURE: [
            "advanced architecture/runtime control exists as a separate non-default layer for deep users",
        ],
        V8_DEV_STATE: [
            "`NEXT` advanced architecture/runtime configuration remains a separate contract and implementation lane",
        ],
    }

    missing: list[str] = []
    for path, snippets in surfaces.items():
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Layered docs are missing the advanced-surface contract:\n" + "\n".join(missing)


def test_v8_docs_keep_current_release_and_full_target_distinct():
    index_text = (ROOT / "docs" / "architecture-library" / "ARCHITECTURE_LIBRARY_INDEX.md").read_text(encoding="utf-8")
    v8_1_text = V8_1_LIVING_ARCHITECTURE.read_text(encoding="utf-8")
    state_text = V8_DEV_STATE.read_text(encoding="utf-8")

    required_index_snippets = [
        "current release architecture",
        "canonical full production architecture and full actuation target",
    ]
    missing_index_snippets = [snippet for snippet in required_index_snippets if snippet not in index_text]
    assert not missing_index_snippets, (
        "Architecture library index is missing the current-vs-target distinction: "
        f"{missing_index_snippets}"
    )

    required_v8_1_snippets = [
        "This is the current release architecture.",
        "canonical full production architecture and full actuation target beyond the current release",
    ]
    missing_v8_1_snippets = [snippet for snippet in required_v8_1_snippets if snippet not in v8_1_text]
    assert not missing_v8_1_snippets, (
        "V8.1 living architecture doc is missing the current-release distinction: "
        f"{missing_v8_1_snippets}"
    )

    required_state_snippets = [
        "Development is progressing toward the V8.2 full production target.",
        "V8.1 is the current bounded release target.",
        "V8.2 is the distributed, learning, capability-enabled, and actuation target beyond the current release.",
    ]
    missing_state_snippets = [snippet for snippet in required_state_snippets if snippet not in state_text]
    assert not missing_state_snippets, (
        "V8 dev state is missing the current-release vs full-target distinction: "
        f"{missing_state_snippets}"
    )


def test_v8_dev_state_declares_architecture_synchronization_rule():
    text = V8_DEV_STATE.read_text(encoding="utf-8")

    required_snippets = [
        "## Architecture Synchronization Rule",
        "Every slice must:",
        "- update state",
        "- verify README alignment",
        "- verify V8.2 alignment",
        "a slice is not complete unless tests pass, documentation is updated where meaning changed, and architecture alignment is verified",
        "`COMPLETE` records accepted delivered work",
        "`ACTIVE` records work in progress",
        "`NEXT` records the next committed follow-on slices",
        "slice close-out should explicitly report tests run, docs changed, and docs reviewed unchanged for the touched scope",
    ]
    missing_snippets = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing_snippets, (
        "V8 dev state is missing the architecture synchronization rule: "
        f"{missing_snippets}"
    )


def test_active_docs_require_docs_review_as_part_of_slice_completion():
    required_snippets = {
        ROOT / "AGENTS.md": [
            "Every implementation slice that changes product behavior, runtime behavior, operator workflow, API contract, governance posture, or canonical terminology must include a documentation review in the same slice.",
            "When API behavior or payload meaning changes, review `docs/API_REFERENCE.md` in the same slice.",
        ],
        README: [
            "every implementation slice must include a docs review for the touched surface, even when the result is \"reviewed, no content change required\"",
            "`docs/API_REFERENCE.md` when API behavior, payload meaning, or endpoint contract changes",
        ],
        ROOT / "docs" / "TESTING.md": [
            "feature work is also not done until the touched docs are reviewed and updated where meaning changed",
            "end-of-slice reporting should name both the evidence commands run and the docs updated or reviewed unchanged for the touched scope",
        ],
        ROOT / "docs" / "architecture" / "OPERATIONS.md": [
            "Implementation slices that change runtime, tasking, validation, API meaning, or operator behavior must review and update the owning docs in the same change rather than leaving docs drift for later cleanup.",
        ],
        ROOT / "ops" / "README.md": [
            "Task, runtime, or validation changes are not complete until the matching docs are reviewed and updated in the same slice.",
            "Delivery reporting should include the commands run plus the docs changed and the touched docs reviewed unchanged.",
        ],
    }

    missing: list[str] = []
    for path, snippets in required_snippets.items():
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, (
        "Active docs are missing the docs-review/update contract for implementation slices:\n"
        + "\n".join(missing)
    )


def test_v8_2_declares_development_alignment():
    text = V8_2_FULL_ARCHITECTURE.read_text(encoding="utf-8")

    required_snippets = [
        "## Development Alignment",
        "V8.2 is the full target.",
        "Implementation progresses incrementally from the current V8.1 release toward this architecture.",
        "Documentation must always reflect:",
        "- what is implemented",
        "- what is planned",
        "no silent divergence is allowed between implementation, release contract, and full architecture target",
    ]
    missing_snippets = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing_snippets, (
        "V8.2 full architecture doc is missing the development-alignment contract: "
        f"{missing_snippets}"
    )
