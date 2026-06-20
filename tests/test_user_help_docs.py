from __future__ import annotations

import re
from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
DOCS_MANIFEST = ROOT / "interface" / "lib" / "docsManifest.ts"
SOMA_CHAT = ROOT / "docs" / "user" / "soma-chat.md"


def _manifest_section(text: str, section: str) -> str:
    match = re.search(
        rf'section:\s*"{re.escape(section)}".*?docs:\s*\[(.*?)\]\s*,\s*\}}',
        text,
        flags=re.DOTALL,
    )
    assert match, f"docs manifest is missing `{section}`"
    return match.group(1)


def test_user_help_start_here_stays_operator_first():
    manifest = DOCS_MANIFEST.read_text(encoding="utf-8")
    start_here = _manifest_section(manifest, "Start Here")
    advanced = _manifest_section(manifest, "Advanced User Surfaces")

    for slug in ["user-docs-home", "soma-chat", "teams-guide", "resources-guide"]:
        assert slug in start_here
    assert "meta-agent-blueprint" not in start_here
    assert "meta-agent-blueprint" in advanced


def test_soma_chat_doc_matches_current_outcome_workspace():
    text = SOMA_CHAT.read_text(encoding="utf-8")

    for required in [
        "You ask -> Soma understands -> optional proposal -> execution -> output/proof/recovery -> revisit",
        "When output is ready and recovery is also present",
        "`web_search` for search intent",
        "MYCELIS_SEARCH_LOCAL_API_ENDPOINT",
        "Search source: Local Mycelis context",
        "fall back to bounded text search",
    ]:
        assert required in text

    assert "raw `tool_call` JSON" in text
    assert "Blueprints And Mission Planning" not in _manifest_section(
        DOCS_MANIFEST.read_text(encoding="utf-8"),
        "Start Here",
    )
