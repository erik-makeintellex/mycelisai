from __future__ import annotations

import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
README = ROOT / "README.md"
DOCS_MANIFEST = ROOT / "interface" / "lib" / "docsManifest.ts"
PRD_INDEX = ROOT / "architecture/mycelis-architecture-v7.md"
LEGACY_V7_DEV_STATE = ROOT / ".state/V7_DEV_STATE.md"
V8_DEV_STATE = ROOT / ".state/V8_DEV_STATE.md"
V8_BOOTSTRAP_MODEL = ROOT / "docs" / "architecture-library" / "V8_CONFIG_AND_BOOTSTRAP_MODEL.md"
V8_UI_API_CONTRACT = ROOT / "docs" / "architecture-library" / "V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md"
V8_1_LIVING_ARCHITECTURE = ROOT / "docs" / "architecture-library" / "V8_1_LIVING_ORGANIZATION_ARCHITECTURE.md"
V8_2_FULL_ARCHITECTURE = ROOT / "architecture/v8-2.md"
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
        "architecture/v8-2.md",
        "architecture/mycelis-architecture-v7.md",
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
        # retired core/internal/registry/README.md
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
        # retired core/internal/registry/README.md
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
    assert ".state/V8_DEV_STATE.md" in active_section
    assert "V8_RUNTIME_CONTRACTS.md" in active_section
    assert "_V7.md" not in active_section and ".state/V7_DEV_STATE.md" not in active_section, (
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
        ".state/V7_DEV_STATE.md",
        "interface/lib/docsManifest.ts",
        "docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md",
    ]

    missing = [ref for ref in required_refs if ref not in text]
    assert not missing, f"Fresh-agent review sequence is missing required references: {missing}"


def test_docs_readme_promotes_v8_navigation_before_legacy_v7_inputs():
    text = (ROOT / "docs" / "README.md").read_text(encoding="utf-8")

    required_v8_refs = [
        "**Active Development State**: `../.state/V8_DEV_STATE.md`",
        "**V8 Runtime Contracts**: `./architecture-library/V8_RUNTIME_CONTRACTS.md`",
        "**V8 Config and Bootstrap Model**: `./architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md`",
        "**V8 UI/API and Operator Experience Contract**: `./architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md`",
    ]
    missing = [ref for ref in required_v8_refs if ref not in text]
    assert not missing, f"docs/README is missing required V8 navigation references: {missing}"

    assert "Migration inputs and historical references:" in text
    assert text.index("**Active Development State**: `../.state/V8_DEV_STATE.md`") < text.index("**Legacy V7 Development State**: `../.state/V7_DEV_STATE.md`")


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
