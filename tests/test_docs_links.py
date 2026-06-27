from __future__ import annotations

import re
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
README = ROOT / "README.md"
DOCS_HOME = ROOT / "docs" / "README.md"
DOCS_MANIFEST = ROOT / "interface" / "lib" / "docsManifest.ts"
ARCH_INDEX = ROOT / "docs" / "architecture-library" / "ARCHITECTURE_LIBRARY_INDEX.md"
CANONICAL_PRD = ROOT / "docs" / "architecture-library" / "MYCELIS_CANONICAL_PRD.md"
V8_DEV_STATE = ROOT / ".state" / "V8_DEV_STATE.md"


def _local_links(path: Path) -> list[str]:
    text = path.read_text(encoding="utf-8")
    return re.findall(r"\]\(([^)]+)\)", text)


def _assert_links_resolve(path: Path) -> None:
    missing: list[str] = []
    for link in _local_links(path):
        if link.startswith(("http://", "https://", "mailto:", "/docs?doc=")):
            continue
        target = link.split("#", 1)[0]
        if not target or target.startswith("#"):
            continue
        resolved = (path.parent / target).resolve()
        if not resolved.exists():
            missing.append(link)
    assert not missing, f"{path.relative_to(ROOT)} contains broken local links: {missing}"


def test_readme_docs_home_and_architecture_links_resolve():
    for path in (README, DOCS_HOME, ARCH_INDEX, CANONICAL_PRD, ROOT / "architecture" / "README.md"):
        _assert_links_resolve(path)


def test_docs_manifest_paths_resolve_and_exposes_canonical_prd():
    text = DOCS_MANIFEST.read_text(encoding="utf-8")
    paths = re.findall(r'path:\s*"([^"]+)"', text)
    missing = [path for path in paths if not (ROOT / path).exists()]

    assert not missing, f"docsManifest contains broken paths: {missing}"
    assert 'slug: "mycelis-canonical-prd"' in text
    assert 'path: "docs/architecture-library/MYCELIS_CANONICAL_PRD.md"' in text


def test_readme_style_pages_expose_project_navigation_and_tocs():
    required = {
        README: "## README TOC",
        DOCS_HOME: "## Docs TOC",
        ARCH_INDEX: "## TOC",
        ROOT / "architecture" / "README.md": "# Architecture",
        ROOT / "ops" / "README.md": "## TOC",
        ROOT / "core" / "README.md": "## TOC",
        ROOT / "interface" / "README.md": "## TOC",
        ROOT / "cli" / "README.md": "## TOC",
        ROOT / "core" / "internal" / "registry" / "README.md": "## TOC",
    }

    missing: list[str] = []
    for path, heading in required.items():
        text = path.read_text(encoding="utf-8")
        if heading not in text:
            missing.append(f"{path.relative_to(ROOT)} missing `{heading}`")
        if path != README:
            top = "\n".join(text.splitlines()[:8])
            if "Project README" not in top or "Navigation:" not in top:
                missing.append(f"{path.relative_to(ROOT)} missing project navigation")

    assert not missing, "README-style pages are missing navigation or TOCs:\n" + "\n".join(missing)


def test_canonical_prd_covers_full_product_architecture_and_release_contract():
    text = CANONICAL_PRD.read_text(encoding="utf-8")
    required = [
        "Mycelis is a Soma-centered governed cognitive operating environment.",
        "The prime architecture rule is twofold",
        "protect confidence while making complexity disappear",
        "Ask\n-> Understand\n-> Approve\n-> Execute\n-> Deliver\n-> Trust\n-> Recover\n-> Revisit",
        "compact Quick Actions shelf",
        "large Talk to Soma thread as the primary canvas",
        "header Outcomes button that opens Outcome Vault on demand",
        "Explore",
        "Shape",
        "Execute",
        "OutcomeProject",
        "TeamRegistryEntry",
        "The Outcome never serves the runtime.",
        "Teams are autonomous execution mechanisms, never sovereign authorities",
        "Continuity vectors are not Soma's mind",
        "Capabilities are governed runtime objects",
        "available to all Soma work",
        "grouped for a capability set or environment",
        "targeted to a specific host/provider/tool endpoint",
        "WorkIntent",
        "ExecutionMode",
        "ExecutionContract",
        "No raw backend stack traces should reach the default UI.",
        "P0.1",
        "P0.10",
        "headed browser proof for actual user experience",
        "This PRD is the canonical architecture/product document.",
    ]
    missing = [snippet for snippet in required if snippet not in text]
    assert not missing, "Canonical PRD is missing required coverage: " + str(missing)


def test_active_navigation_points_to_single_architecture_prd():
    surfaces = {
        README: ["MYCELIS_CANONICAL_PRD.md", ".state/V8_DEV_STATE.md"],
        DOCS_HOME: ["MYCELIS_CANONICAL_PRD.md", "Architecture Docs Index"],
        ARCH_INDEX: ["MYCELIS_CANONICAL_PRD.md", "Do not restore split V7, V8.2, or V8.3 architecture documents"],
        V8_DEV_STATE: ["MYCELIS_CANONICAL_PRD.md", "canonical PRD alignment"],
    }

    missing: list[str] = []
    for path, snippets in surfaces.items():
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Active navigation is not aligned to the canonical PRD:\n" + "\n".join(missing)


def test_old_architecture_docs_are_deleted_not_archived_or_exposed():
    stale_paths = [
        ".state/V7_DEV_STATE.md",
        "architecture/mycelis-architecture-v7.md",
        "architecture/v8-2.md",
        "docs/architecture-library/V8_2_SOMA_TEAM_INTERACTION_CONTRACT.md",
        "docs/architecture-library/V8_2_SOMA_UI_ARCHITECTURE_EXPRESSION.md",
        "docs/architecture-library/V8_3_OPERATIONAL_EMBODIMENT_PRD.md",
        "docs/architecture-library/V8_3_RELEASE_ARCHITECTURE_DELIVERY_BRIEF.md",
        "docs/architecture-library/V8_3_RELEASE_RUNTIME_REFERENCE.md",
        "docs/architecture-library/V8_3_SOMA_USER_EXPERIENCE_CONTRACT.md",
        "docs/architecture-library/V8_3_UI_UX_ENGINEERING_IMPLEMENTATION_BRIEF.md",
        "docs/architecture-library/V8_3_MVP_UI_RUNTIME_DELIVERY_PLAN.md",
        "docs/architecture-library/V8_RUNTIME_CONTRACTS.md",
        "docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md",
        "docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md",
        "docs/architecture-library/V8_CAPABILITY_MANIFEST_AND_RUNTIME_INTEGRATION_STANDARD.md",
        "docs/architecture-library/V8_SECRET_STORAGE_AND_CREDENTIAL_BOUNDARY.md",
        "docs/architecture-library/V8_UI_TEAM_FULL_TEST_SET.md",
        "docs/architecture-library/V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md",
    ]

    present = [path for path in stale_paths if (ROOT / path).exists()]
    combined_navigation = "\n".join(
        [
            README.read_text(encoding="utf-8"),
            DOCS_HOME.read_text(encoding="utf-8"),
            DOCS_MANIFEST.read_text(encoding="utf-8"),
            ARCH_INDEX.read_text(encoding="utf-8"),
        ]
    )
    exposed = [path for path in stale_paths if path in combined_navigation]

    assert not present, "Superseded docs should be deleted: " + str(present)
    assert not exposed, "Superseded docs should not be exposed: " + str(exposed)


def test_docs_review_contract_remains_visible():
    required_snippets = {
        ROOT / "AGENTS.md": [
            "Every implementation slice that changes product behavior, runtime behavior, operator workflow, API contract, governance posture, or canonical terminology must include a documentation review in the same slice.",
            "docs/architecture-library/MYCELIS_CANONICAL_PRD.md",
        ],
        README: [
            "every implementation slice must include a docs review for the touched surface",
            "`docs/API_REFERENCE.md` when API behavior, payload meaning, or endpoint contract changes",
        ],
        ROOT / "docs" / "TESTING.md": [
            "Feature work is not done until relevant tests run against the final branch state",
            "Mycelis Canonical PRD",
        ],
        ROOT / "docs" / "architecture" / "OPERATIONS.md": [
            "Implementation slices that change runtime, tasking, validation, API meaning, or operator behavior must review and update the owning docs in the same change",
        ],
        ROOT / "ops" / "README.md": [
            "Task, runtime, or validation changes are not complete until the matching docs are reviewed and updated in the same slice.",
        ],
    }

    missing: list[str] = []
    for path, snippets in required_snippets.items():
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Docs-review/update contract is missing:\n" + "\n".join(missing)
