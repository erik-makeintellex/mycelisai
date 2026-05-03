import shutil
from pathlib import Path

WORKTREE_REVIEW_TARGETS = (
    "README.md",
    ".state/V8_DEV_STATE.md",
    "docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md",
    "docs/architecture-library/V8_RUNTIME_CONTRACTS.md",
    "docs/architecture-library/V8_CONFIG_AND_BOOTSTRAP_MODEL.md",
    "docs/architecture-library/V8_UI_API_AND_OPERATOR_EXPERIENCE_CONTRACT.md",
    "docs/LOCAL_DEV_WORKFLOW.md",
    "docs/architecture/OPERATIONS.md",
    "docs/TESTING.md",
    "ops/README.md",
)

WORKTREE_BASELINE_INSTALLS = ("uv run inv install",)

WORKTREE_BASELINE_COMMANDS = (
    "uv run inv ci.entrypoint-check",
    "uv run inv ci.baseline",
)

WORKTREE_AREA_RULES = (
    {
        "name": "Core runtime",
        "prefixes": ("core/", "proto/"),
        "installs": ("cd core && go mod download",),
        "commands": ("uv run inv core.test", "uv run inv core.compile"),
    },
    {
        "name": "Interface",
        "prefixes": ("interface/",),
        "installs": ("uv run inv interface.install",),
        "commands": (
            "uv run inv interface.test",
            "uv run inv interface.typecheck",
            "uv run inv interface.build",
        ),
    },
    {
        "name": "Cognitive",
        "prefixes": ("cognitive/",),
        "installs": ("uv run inv cognitive.install",),
        "commands": (),
    },
    {
        "name": "Python automation",
        "prefixes": ("agents/", "cli/", "ops/", "scripts/", "sdk/python/", "tests/"),
        "exact_paths": ("pyproject.toml", "tasks.py"),
        "installs": ("uv sync --all-packages --dev",),
        "commands": (
            "$env:PYTHONPATH='.'; uv run pytest tests/test_core_tasks.py tests/test_ci_tasks.py tests/test_interface_tasks.py tests/test_interface_e2e_tasks.py tests/test_interface_command_tasks.py tests/test_k8s_tasks.py tests/test_lifecycle_tasks.py tests/test_misc_tasks.py -q",
            "uv run inv ci.build",
        ),
    },
    {
        "name": "Docs and state",
        "prefixes": ("docs/",),
        "exact_paths": (
            "README.md",
            ".state/V7_DEV_STATE.md",
            ".state/V8_DEV_STATE.md",
            "architecture/mycelis-architecture-v7.md",
        ),
        "installs": (),
        "commands": ("$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q",),
    },
    {
        "name": "Infra and deploy",
        "prefixes": ("charts/", "deploy/", "k8s/"),
        "installs": (),
        "commands": (),
    },
)

def architecture_sync_directives():
    messages = {
        "prime-architect": "Central architecture directive: keep the next-target workflow aligned to strict gate order. "
        "P0 remains the active phase. Require concrete test evidence before any phase advancement, "
        "coordinate development and AGUI work to the target goals, and reply with a concise execution brief. "
        "Do not use tools for this sync. Respond in plain text with at most 6 short lines.",
        "prime-development": "Development directive: focus on P0 closure, memory-restart reliability, logging standardization, "
        "error-handling normalization, and no-regression verification. Go remains the primary implementation "
        "language for backend/runtime work. Python is limited to tasks, management scripting, and tests. "
        "Reply with the top implementation/testing priorities. Do not use tools for this sync. "
        "Respond in plain text with at most 5 short lines.",
        "agui-design-architect": "AGUI directive: align base UI updates to architecture truth. Prioritize workflow-composer onboarding, "
        "gate-state visibility, system status, team roster visibility, and operator-safe error presentation. "
        "Do not invent client-side workflow semantics that diverge from backend gates. "
        "Reply with the top UI architecture priorities. Do not use tools for this sync. "
        "Respond in plain text with at most 5 short lines.",
    }
    return {
        team_id: {
            "command_subject": f"swarm.team.{team_id}.internal.command",
            "reply_subjects": (f"swarm.team.{team_id}.signal.status", f"swarm.team.{team_id}.signal.result"),
            "message": message,
        }
        for team_id, message in messages.items()
    }

def repo_relative(path: Path, root_dir: Path) -> str:
    try:
        return str(path.resolve().relative_to(root_dir.resolve())).replace("\\", "/")
    except ValueError:
        return str(path)


def assert_repo_managed_target(path: Path, root_dir: Path) -> Path:
    resolved = path.resolve()
    try:
        resolved.relative_to(root_dir.resolve())
    except ValueError as exc:
        raise SystemExit(f"CLEANUP FAILED: refusing to touch non-repo path: {path}") from exc
    return resolved

def artifact_size_bytes(path: Path) -> int:
    if not path.exists():
        return 0
    if path.is_file():
        return path.stat().st_size

    total = 0
    for child in path.rglob("*"):
        if child.is_file():
            try:
                total += child.stat().st_size
            except OSError:
                continue
    return total

def format_size_bytes(size_bytes: int) -> str:
    units = ("B", "KB", "MB", "GB", "TB")
    value = float(size_bytes)
    for unit in units:
        if value < 1024 or unit == units[-1]:
            return f"{value:.1f} {unit}"
        value /= 1024
    return f"{size_bytes} B"

def remove_repo_targets(
    targets: tuple[Path, ...], root_dir: Path
) -> tuple[list[str], list[str]]:
    removed: list[str] = []
    missing: list[str] = []
    for target in targets:
        managed_target = assert_repo_managed_target(target, root_dir)
        label = repo_relative(target, root_dir)
        if not managed_target.exists():
            missing.append(label)
            continue
        if managed_target.is_file():
            managed_target.unlink()
        else:
            shutil.rmtree(managed_target)
        removed.append(label)
    return removed, missing

def report_repo_targets(
    targets: tuple[Path, ...], root_dir: Path
) -> list[dict[str, object]]:
    report: list[dict[str, object]] = []
    for target in targets:
        managed_target = assert_repo_managed_target(target, root_dir)
        exists = managed_target.exists()
        report.append(
            {
                "path": repo_relative(target, root_dir),
                "exists": exists,
                "bytes": artifact_size_bytes(managed_target) if exists else 0,
            }
        )
    return report


def print_cleanup_summary(removed: list[str], missing: list[str]) -> None:
    if removed:
        print("Removed:")
        for path in removed:
            print(f"  - {path}")
    else:
        print("Removed:")
        print("  - none")

    if missing:
        print("Already clean:")
        for path in missing:
            print(f"  - {path}")


def _unique_strings(items):
    ordered = []
    seen = set()
    for item in items:
        if item in seen:
            continue
        seen.add(item)
        ordered.append(item)
    return ordered


def _normalize_git_path(path: str) -> str:
    normalized = path.strip()
    if " -> " in normalized:
        normalized = normalized.split(" -> ", 1)[1]
    return normalized.replace("\\", "/")


def _parse_porcelain_entries(status_output: str):
    entries = []
    for raw_line in status_output.splitlines():
        if not raw_line.strip() or len(raw_line) < 4:
            continue
        entries.append(
            {"status": raw_line[:2].strip() or "??", "path": _normalize_git_path(raw_line[3:])}
        )
    return entries


def _match_worktree_rule(path: str):
    for rule in WORKTREE_AREA_RULES:
        if path in rule.get("exact_paths", ()):
            return rule
        if any(path.startswith(prefix) for prefix in rule.get("prefixes", ())):
            return rule
    return {"name": "Unclassified", "installs": (), "commands": ()}


def build_worktree_triage(status_output: str):
    entries = _parse_porcelain_entries(status_output)
    area_counts = {}
    priority_installs = []
    recommended_commands = list(WORKTREE_BASELINE_COMMANDS)

    for entry in entries:
        rule = _match_worktree_rule(entry["path"])
        entry["area"] = rule["name"]
        area_counts[rule["name"]] = area_counts.get(rule["name"], 0) + 1
        priority_installs.extend(rule.get("installs", ()))
        recommended_commands.extend(rule.get("commands", ()))

    return {
        "entries": entries,
        "areas": [
            {"name": name, "count": count}
            for name, count in sorted(area_counts.items())
        ],
        "priority_installs": _unique_strings(priority_installs),
        "recommended_commands": _unique_strings(recommended_commands),
    }


def print_worktree_triage(
    triage,
    windows_host: bool,
    review_targets=WORKTREE_REVIEW_TARGETS,
    baseline_installs=WORKTREE_BASELINE_INSTALLS,
) -> None:
    entries = triage["entries"]

    print(f"Working tree: dirty ({len(entries)} changed path(s))" if entries else "Working tree: clean")
    print("\nExpected review targets:")
    for target in review_targets:
        print(f"  - {target}")

    print("\nDependency reset:")
    for command in baseline_installs:
        print(f"  - {command}")

    if windows_host:
        print("\nWindows host note:")
        print("  - treat this checkout as source-only and run heavy validation from the WSL checkout.")

    print("\nPriority install checks:")
    if triage["priority_installs"]:
        for command in triage["priority_installs"]:
            print(f"  - {command}")
    else:
        print("  - none triggered by current paths")

    print("\nRecommended evidence commands:")
    for command in triage["recommended_commands"]:
        print(f"  - {command}")

    if entries:
        print("\nArea breakdown:")
        for area in triage["areas"]:
            print(f"  - {area['name']}: {area['count']} path(s)")

        print("\nChanged paths:")
        preview = entries[:20]
        for entry in preview:
            print(f"  - [{entry['status']}] {entry['path']} ({entry['area']})")
        if len(entries) > len(preview):
            print(f"  - ... and {len(entries) - len(preview)} more")
    else:
        print("\nChanged paths:")
        print("  - none")
