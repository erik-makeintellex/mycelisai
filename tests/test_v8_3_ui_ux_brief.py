from __future__ import annotations

from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
UI_UX_BRIEF = ROOT / "docs" / "architecture-library" / "V8_3_UI_UX_ENGINEERING_IMPLEMENTATION_BRIEF.md"
ARCH_INDEX = ROOT / "docs" / "architecture-library" / "ARCHITECTURE_LIBRARY_INDEX.md"
DOCS_HOME = ROOT / "docs" / "README.md"
RELEASE_BRIEF = ROOT / "docs" / "architecture-library" / "V8_3_RELEASE_ARCHITECTURE_DELIVERY_BRIEF.md"
DOCS_MANIFEST = ROOT / "interface" / "lib" / "docsManifest.ts"
README = ROOT / "README.md"
V8_DEV_STATE = ROOT / ".state" / "V8_DEV_STATE.md"


def test_v8_3_ui_ux_brief_layers_threaded_workspace_mandate():
    text = UI_UX_BRIEF.read_text(encoding="utf-8")

    required = [
        "Dashboard becomes the Soma threaded workspace.",
        "OperationalAlertFrame",
        "SomaActionShelf",
        "Outcome Vault",
        "Capability-permission settings cards",
        "Typed thread-event adapter",
        "WebSocket bridge from event bus to frontend state",
        "Ask\n-> Understand\n-> Approve\n-> Execute\n-> Deliver\n-> Trust\n-> Recover\n-> Revisit",
        "Do not accept a UI slice that makes the user learn agents, MCP, NATS, workflows, runs, topology, or infrastructure",
    ]

    missing = [snippet for snippet in required if snippet not in text]
    assert not missing, "UI/UX implementation brief is missing required mandate text: " + ", ".join(missing)


def test_v8_3_ui_ux_brief_is_registered_in_active_docs():
    references = {
        ARCH_INDEX: "V8_3_UI_UX_ENGINEERING_IMPLEMENTATION_BRIEF.md",
        DOCS_HOME: "V8_3_UI_UX_ENGINEERING_IMPLEMENTATION_BRIEF.md",
        RELEASE_BRIEF: "V8.3 UI/UX Engineering Implementation Brief",
        DOCS_MANIFEST: "v8-3-ui-ux-engineering-implementation-brief",
        README: "V8_3_UI_UX_ENGINEERING_IMPLEMENTATION_BRIEF.md",
        V8_DEV_STATE: "V8.3 UI/UX Engineering Implementation Brief",
    }

    missing = [
        f"{path.relative_to(ROOT)} missing `{snippet}`"
        for path, snippet in references.items()
        if snippet not in path.read_text(encoding="utf-8")
    ]

    assert not missing, "UI/UX implementation brief is not registered everywhere required:\n" + "\n".join(missing)
