from __future__ import annotations

from pathlib import Path

from invoke import Collection, task

from .config import ROOT_DIR

DEFAULT_SOURCE_PATHS = (
    "core,interface,ops,tests,scripts,agents,cli,sdk/python/src,"
    "cognitive/src,docs,charts,k8s,architecture,proto,README.md,"
    "AGENTS.md,pyproject.toml"
)
DEFAULT_HOT_PATHS = DEFAULT_SOURCE_PATHS
DEFAULT_EXTENSIONS = {
    ".go",
    ".py",
    ".ts",
    ".tsx",
    ".js",
    ".jsx",
    ".md",
    ".sql",
    ".yaml",
    ".yml",
    ".json",
    ".toml",
    ".proto",
    ".sh",
    ".ps1",
    ".tpl",
}
DEFAULT_EXCLUDE_DIRS = {
    ".git",
    ".next",
    ".venv",
    "node_modules",
    "dist",
    "build",
    "coverage",
    "test-results",
    "playwright-report",
}
GENERATED_OR_LOCK_PATHS = {
    "core/go.sum": "Go module checksum lockfile",
    "scripts/qa/go.sum": "Go module checksum lockfile",
    "interface/package-lock.json": "npm lockfile",
    "uv.lock": "uv lockfile",
    "charts/mycelis-core/Chart.lock": "Helm dependency lockfile",
}
GENERATED_SUFFIXES = (
    ".pb.go",
    "_pb2.py",
    "_pb2_grpc.py",
)
LEGACY_CAPS_PATH = ROOT_DIR / "ops" / "quality_legacy_caps.txt"


def _parse_paths(paths: str) -> list[Path]:
    result: list[Path] = []
    for raw in paths.split(","):
        value = raw.strip()
        if not value:
            continue
        path = ROOT_DIR / value
        if path.exists():
            result.append(path)
    return result


def _should_skip(path: Path) -> bool:
    if any(part in DEFAULT_EXCLUDE_DIRS for part in path.parts):
        return True
    rel = path.relative_to(ROOT_DIR).as_posix() if path.is_absolute() else path.as_posix()
    if rel in GENERATED_OR_LOCK_PATHS:
        return True
    name = path.name
    if name.endswith(GENERATED_SUFFIXES):
        return True
    return False


def _iter_source_files(paths: list[Path]) -> list[Path]:
    files: list[Path] = []
    for root in paths:
        if root.is_file():
            if root.suffix in DEFAULT_EXTENSIONS and not _should_skip(root):
                files.append(root)
            continue

        for path in root.rglob("*"):
            if not path.is_file():
                continue
            if path.suffix not in DEFAULT_EXTENSIONS:
                continue
            if _should_skip(path):
                continue
            files.append(path)
    return files


def _line_count(path: Path) -> int:
    return len(path.read_text(encoding="utf-8").splitlines())


def _path_is_under(path: Path, root: Path) -> bool:
    try:
        path.relative_to(root)
        return True
    except ValueError:
        return False


def _load_legacy_caps(path: Path = LEGACY_CAPS_PATH) -> dict[str, int]:
    caps: dict[str, int] = {}
    if not path.exists():
        return caps

    for raw in path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line or line.startswith("#"):
            continue
        if "=" not in line:
            raise SystemExit(f"QUALITY CHECK FAILED: malformed legacy cap entry: {line}")
        key, value = line.split("=", 1)
        rel = key.strip().replace("\\", "/")
        if not rel:
            raise SystemExit("QUALITY CHECK FAILED: legacy cap entry has an empty path.")
        if rel in caps:
            raise SystemExit(f"QUALITY CHECK FAILED: duplicate legacy cap entry: {rel}")
        try:
            cap = int(value.strip())
        except ValueError:
            raise SystemExit(f"QUALITY CHECK FAILED: legacy cap entry has a non-integer cap: {line}") from None
        if cap <= 0:
            raise SystemExit(f"QUALITY CHECK FAILED: legacy cap entry must be positive: {line}")
        caps[rel] = cap
    return caps


@task(
    help={
        "limit": "Maximum allowed lines per file (default: 300).",
        "paths": "Comma-separated paths to scan (default: main source tree).",
        "strict": "Ignore legacy caps and fail on every over-limit file.",
    }
)
def max_lines(_c, limit=300, paths=DEFAULT_SOURCE_PATHS, strict=False):
    """
    Enforce maximum file length with temporary no-regression caps for legacy files.
    """
    roots = _parse_paths(paths)
    if not roots:
        raise SystemExit("QUALITY CHECK FAILED: no valid paths provided.")

    caps = {} if strict else _load_legacy_caps()
    files = _iter_source_files(roots)
    violations: list[str] = []
    legacy_ok: list[str] = []
    loose_caps: list[str] = []
    retired_caps: list[str] = []
    stale_caps: list[str] = []

    if not strict:
        for rel in sorted(caps):
            cap_path = ROOT_DIR / rel
            if any(_path_is_under(cap_path, root) for root in roots) and not cap_path.exists():
                stale_caps.append(rel)

    for path in files:
        try:
            rel = path.relative_to(ROOT_DIR).as_posix()
        except ValueError:
            rel = path.as_posix()
        count = _line_count(path)
        cap = caps.get(rel)

        if cap is not None and count <= limit:
            retired_caps.append(f"{rel}: {count} lines (remove legacy-cap={cap})")
            continue

        if count <= limit:
            continue

        if strict or cap is None or count > cap:
            detail = f"{rel}: {count} lines (limit={limit}"
            if cap is not None and not strict:
                detail += f", legacy-cap={cap}"
            detail += ")"
            violations.append(detail)
        elif cap > count:
            loose_caps.append(f"{rel}: {count} lines (lower legacy-cap={cap})")
        else:
            legacy_ok.append(f"{rel}: {count}/{cap}")

    if legacy_ok:
        print("Legacy files still over global limit but within temporary caps:")
        for item in legacy_ok:
            print(f"  {item}")

    if violations:
        print("Max-line violations:")
        for item in violations:
            print(f"  {item}")
        raise SystemExit("QUALITY CHECK FAILED: file length violations found.")

    if loose_caps:
        print("Loose legacy cap entries:")
        for item in loose_caps:
            print(f"  {item}")
        raise SystemExit("QUALITY CHECK FAILED: legacy max-line caps must match current line counts.")

    if retired_caps:
        print("Retired legacy cap entries:")
        for item in retired_caps:
            print(f"  {item}")
        raise SystemExit("QUALITY CHECK FAILED: legacy max-line caps for files under the limit must be removed.")

    if stale_caps:
        print("Stale legacy cap entries:")
        for item in stale_caps:
            print(f"  {item}")
        raise SystemExit("QUALITY CHECK FAILED: stale legacy max-line caps found.")

    print(f"QUALITY CHECK PASSED: scanned {len(files)} files, limit={limit}.")


ns = Collection("quality")
ns.add_task(max_lines, name="max-lines")
