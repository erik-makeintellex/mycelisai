from __future__ import annotations

import re
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
README = ROOT / "README.md"
DOCS_MANIFEST = ROOT / "interface" / "lib" / "docsManifest.ts"
PRD_INDEX = ROOT / "mycelis-architecture-v7.md"
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
