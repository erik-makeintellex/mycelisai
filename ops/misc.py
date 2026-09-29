from pathlib import Path

from invoke import Collection, task

from .cleanup_support import filter_active_runtime_targets, print_active_runtime_skip
from .config import ROOT_DIR
from .misc_support import (
    WORKTREE_BASELINE_INSTALLS,
    WORKTREE_REVIEW_TARGETS,
    build_worktree_triage as _build_worktree_triage,
    format_size_bytes,
    print_cleanup_summary,
    print_worktree_triage,
    remove_repo_targets,
    report_repo_targets,
)

GENERATED_ARTIFACT_RELATIVE_TARGETS = (
    ".venv",
    "interface/node_modules",
    "interface/.next",
    "workspace/tool-cache",
    "core/workspace/tool-cache",
    "interface/workspace/tool-cache",
    "interface/test-results",
    "interface/playwright-report",
    "interface/.playwright",
    "interface/tsconfig.tsbuildinfo",
    "interface/next-env.d.ts",
    ".pytest_cache",
    "core/bin",
)

SOURCE_TREE_CACHE_ROOTS = (
    ".",
    "agents",
    "cli",
    "cognitive",
    "framework_runs",
    "ops",
    "sdk/python",
    "tests",
)

REPORT_ARTIFACT_RELATIVE_TARGETS = (
    "interface/test-results",
    "interface/playwright-report",
    ".pytest_cache",
)

WSL_HANDOFF_RELATIVE_TARGETS = (
    ".venv",
    "interface/node_modules",
    "interface/.next",
)


def _source_tree_pycache_targets(root_dir: Path) -> tuple[Path, ...]:
    targets: set[Path] = set()
    root_cache = root_dir / "__pycache__"
    if root_cache.exists():
        targets.add(root_cache)
    for relative_root in SOURCE_TREE_CACHE_ROOTS[1:]:
        source_root = root_dir / relative_root
        if source_root.exists():
            targets.update(source_root.rglob("__pycache__"))
    return tuple(sorted(targets, key=lambda path: path.as_posix()))


def _generated_artifact_targets(root_dir: Path) -> tuple[Path, ...]:
    explicit = tuple(root_dir / path for path in GENERATED_ARTIFACT_RELATIVE_TARGETS)
    return explicit + _source_tree_pycache_targets(root_dir)


def _relative_targets(root_dir: Path, relative_targets: tuple[str, ...]) -> tuple[Path, ...]:
    return tuple(root_dir / path for path in relative_targets)


@task(name="generated")
def clean_generated(c):
    """Remove repo-local generated artifacts (build output, caches, reports) from the WSL checkout."""
    targets, skipped = filter_active_runtime_targets(_generated_artifact_targets(ROOT_DIR), ROOT_DIR)
    removed, missing = remove_repo_targets(tuple(targets), ROOT_DIR)
    print("=== CLEAN GENERATED ===")
    print_cleanup_summary(removed, missing)
    print_active_runtime_skip(skipped)
    print("Runtime data note:")
    print("  - workspace/docker-compose/data is intentionally untouched.")


@task(name="reports")
def clean_reports(c):
    """Remove lightweight test/report artifacts without clearing install caches."""
    report_targets = _relative_targets(ROOT_DIR, REPORT_ARTIFACT_RELATIVE_TARGETS)
    removed, missing = remove_repo_targets(report_targets, ROOT_DIR)
    print("=== CLEAN REPORTS ===")
    print_cleanup_summary(removed, missing)


@task(name="wsl-handoff")
def clean_wsl_handoff(c):
    """Reset cross-host generated artifacts before handing the repo off to WSL."""
    handoff_targets = _relative_targets(ROOT_DIR, WSL_HANDOFF_RELATIVE_TARGETS)
    targets, skipped = filter_active_runtime_targets(handoff_targets, ROOT_DIR)
    removed, missing = remove_repo_targets(tuple(targets), ROOT_DIR)
    print("=== CLEAN WSL HANDOFF ===")
    print_cleanup_summary(removed, missing)
    print_active_runtime_skip(skipped)
    print("Next step:")
    print("  - use a WSL-native checkout for uv/npm/build/test/compose work.")


@task(name="disk-status")
def clean_disk_status(c):
    """Report repo-local generated artifact usage and host-boundary cleanup guidance."""
    report = report_repo_targets(_generated_artifact_targets(ROOT_DIR), ROOT_DIR)
    total_bytes = sum(int(item["bytes"]) for item in report)

    print("=== CLEAN DISK STATUS ===")
    for item in report:
        presence = "present" if item["exists"] else "missing"
        print(
            f"  - {item['path']}: {presence} ({format_size_bytes(int(item['bytes']))})"
        )
    print(f"Repo-local generated total: {format_size_bytes(total_bytes)}")
    print("Storage boundary:")
    print("  - Heavy artifacts live in the WSL checkout, on the native Linux filesystem.")
    print("  - Docker image/volume usage and WSL VHD slack space are outside repo cleanup.")
    print("Low-disk reminder:")
    print("  - run clean.generated first, then `wsl --shutdown`, then compact the WSL VHD from an elevated PowerShell when needed.")

ns_clean = Collection("clean")
ns_clean.add_task(clean_generated)
ns_clean.add_task(clean_reports)
ns_clean.add_task(clean_wsl_handoff)
ns_clean.add_task(clean_disk_status)


@task(name="worktree-triage")
def worktree_triage(c):
    """Summarize dirty-worktree scope, install checks, and evidence commands.

    This is a local maintenance helper under ops/. It must not register,
    persist, or imply runtime teams inside core/config/teams or runtime
    registries.
    """
    print("=== WORKTREE TRIAGE ===")

    status = c.run("git status --porcelain", hide=True, warn=True)
    if status.exited != 0:
        raise SystemExit("WORKTREE TRIAGE FAILED: unable to read git status.")

    triage = _build_worktree_triage(status.stdout or "")
    print_worktree_triage(
        triage,
        review_targets=WORKTREE_REVIEW_TARGETS,
        baseline_installs=WORKTREE_BASELINE_INSTALLS,
    )


ns_team = Collection("team")
ns_team.add_task(worktree_triage)
