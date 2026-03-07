from __future__ import annotations

import re
from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
README = ROOT / "README.md"
DOCS_MANIFEST = ROOT / "interface" / "lib" / "docsManifest.ts"


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
