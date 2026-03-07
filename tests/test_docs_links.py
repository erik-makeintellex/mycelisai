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
