from __future__ import annotations

from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
CANONICAL_PRD = ROOT / "docs" / "architecture-library" / "MYCELIS_CANONICAL_PRD.md"
ARCH_INDEX = ROOT / "docs" / "architecture-library" / "ARCHITECTURE_LIBRARY_INDEX.md"
DOCS_HOME = ROOT / "docs" / "README.md"
DOCS_MANIFEST = ROOT / "interface" / "lib" / "docsManifest.ts"
README = ROOT / "README.md"
V8_DEV_STATE = ROOT / ".state" / "V8_DEV_STATE.md"


def test_canonical_prd_layers_threaded_workspace_mandate():
    text = CANONICAL_PRD.read_text(encoding="utf-8")

    required = [
        "The first authenticated surface is the Soma workspace.",
        "compact Quick Actions shelf",
        "large Talk to Soma thread as the primary canvas",
        "header Outcomes button that opens Outcome Vault on demand",
        "No raw backend stack traces should reach the default UI.",
        "NATS, the current EventSource stream, or a future WebSocket bridge should produce typed thread events",
        "Ask\n-> Understand\n-> Approve\n-> Execute\n-> Deliver\n-> Trust\n-> Recover\n-> Revisit",
        "The dashboard should keep the composer reachable",
    ]

    missing = [snippet for snippet in required if snippet not in text]
    assert not missing, "Canonical PRD is missing required UI/UX mandate text: " + ", ".join(missing)


def test_canonical_prd_is_registered_in_active_docs():
    references = {
        ARCH_INDEX: "MYCELIS_CANONICAL_PRD.md",
        DOCS_HOME: "MYCELIS_CANONICAL_PRD.md",
        DOCS_MANIFEST: "mycelis-canonical-prd",
        README: "MYCELIS_CANONICAL_PRD.md",
        V8_DEV_STATE: "MYCELIS_CANONICAL_PRD.md",
    }

    missing = [
        f"{path.relative_to(ROOT)} missing `{snippet}`"
        for path, snippet in references.items()
        if snippet not in path.read_text(encoding="utf-8")
    ]

    assert not missing, "Canonical PRD is not registered everywhere required:\n" + "\n".join(missing)
