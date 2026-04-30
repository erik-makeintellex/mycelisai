from __future__ import annotations

from pathlib import Path

from .config import ROOT_DIR

INTERFACE_DIR = ROOT_DIR / "interface"


def _process_path_variants(path: Path) -> set[str]:
    normalized = {
        str(path.resolve()).lower().replace("\\", "/"),
        str(path).lower().replace("\\", "/"),
    }
    extra: set[str] = set()
    for entry in normalized:
        if entry.startswith("/mnt/") and len(entry) > 6 and entry[5].isalpha() and entry[6] == "/":
            extra.add(f"{entry[5]}:{entry[6:]}")
    normalized.update(extra)
    return {entry.rstrip("/") for entry in normalized}


_INTERFACE_PROCESS_PATH_HINTS = tuple(
    variant
    for path in {
        INTERFACE_DIR / ".next",
        INTERFACE_DIR / "node_modules",
        INTERFACE_DIR / "scripts",
        INTERFACE_DIR / "playwright-report",
        INTERFACE_DIR / "test-results",
    }
    for variant in _process_path_variants(path)
)

_INTERFACE_PROCESS_COMMAND_HINTS = (
    "/.next/dev/build/postcss.js",
    "/.next/standalone/server.js",
    "/.next/standalone/interface/server.js",
    "/next/dist/bin/next",
    "/next/dist/server/lib/start-server.js",
    "/dist/server/lib/start-server.js",
    "./node_modules/next/dist/bin/next",
    "/scripts/playwright-webserver.mjs",
    "./scripts/playwright-webserver.mjs",
    "/node_modules/.bin/vitest",
    "./node_modules/.bin/vitest",
    "/node_modules/vitest/",
    "/vitest/vitest.mjs",
    "/playwright/",
)
