from __future__ import annotations

from pathlib import Path
import re
import subprocess

from .config import ROOT_DIR

SEMVER_PATTERN = re.compile(r"^\d+\.\d+\.\d+$")
RELEASE_TAG_PATTERN = re.compile(r"^v\d+\.\d+\.\d+(?:-rc\d+)?$")


def read_semver(root_dir: Path = ROOT_DIR) -> str:
    """Read and validate the canonical semantic version from VERSION."""
    version_file = Path(root_dir) / "VERSION"
    if not version_file.exists():
        raise FileNotFoundError("VERSION file missing in project root")
    semver = version_file.read_text(encoding="utf-8").strip()
    if not SEMVER_PATTERN.match(semver):
        raise ValueError(f"Invalid semantic version in VERSION: {semver!r}")
    return semver


def validate_release_tag(tag: str) -> str:
    """Validate an annotated release or RC tag (e.g. v0.8.3 or v0.8.3-rc1)."""
    cleaned = (tag or "").strip()
    if not RELEASE_TAG_PATTERN.match(cleaned):
        raise ValueError(
            f"Invalid release tag {tag!r}; expected v<MAJOR>.<MINOR>.<PATCH> or v<MAJOR>.<MINOR>.<PATCH>-rc<N>."
        )
    return cleaned


def get_version(c=None) -> str:
    """Calculate the immutable build identity tag: v{MAJOR}.{MINOR}.{PATCH}-{SHORT_SHA}."""
    semver = read_semver(ROOT_DIR)
    if c is not None:
        res = c.run("git rev-parse --short HEAD", hide=True)
        sha = res.stdout.strip()
    else:
        sha = (
            subprocess.check_output(["git", "rev-parse", "--short", "HEAD"], cwd=str(ROOT_DIR))
            .decode("utf-8")
            .strip()
        )
    return f"v{semver}-{sha}"
