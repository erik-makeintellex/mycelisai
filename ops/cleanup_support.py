import os
import sys
from pathlib import Path

from .misc_support import repo_relative


def filter_active_runtime_targets(
    targets: tuple[Path, ...], root_dir: Path
) -> tuple[list[Path], list[str]]:
    active_locations = set(Path(sys.executable).absolute().parents)
    active_locations.add(Path(sys.prefix).resolve(strict=False))
    if virtual_env := os.environ.get("VIRTUAL_ENV"):
        active_locations.add(Path(virtual_env).resolve(strict=False))
    kept: list[Path] = []
    skipped: list[str] = []
    for target in targets:
        managed_target = target.resolve(strict=False)
        if managed_target in active_locations:
            skipped.append(repo_relative(target, root_dir))
            continue
        kept.append(target)
    return kept, skipped


def print_active_runtime_skip(skipped: list[str]) -> None:
    if not skipped:
        return
    print("Skipped active runtime:")
    for path in skipped:
        print(f"  - {path}")
    print(
        "  Reason: this task is running from that environment; "
        "remove it from an external shell if needed."
    )
