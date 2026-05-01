import shutil
from pathlib import Path

from invoke import Collection, task

from .config import ROOT_DIR, is_windows

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

WORKTREE_BASELINE_INSTALLS = (
    "uv run inv install",
)

WORKTREE_BASELINE_COMMANDS = (
    "uv run inv ci.entrypoint-check",
    "uv run inv ci.baseline",
)

GENERATED_ARTIFACT_TARGETS = (
    ROOT_DIR / ".venv",
    ROOT_DIR / "interface" / "node_modules",
    ROOT_DIR / "interface" / ".next",
    ROOT_DIR / "workspace" / "tool-cache",
    ROOT_DIR / "interface" / "test-results",
    ROOT_DIR / "interface" / "playwright-report",
    ROOT_DIR / ".pytest_cache",
    ROOT_DIR / "core" / "bin",
)

REPORT_ARTIFACT_TARGETS = (
    ROOT_DIR / "interface" / "test-results",
    ROOT_DIR / "interface" / "playwright-report",
    ROOT_DIR / ".pytest_cache",
)

WSL_HANDOFF_TARGETS = (
    ROOT_DIR / ".venv",
    ROOT_DIR / "interface" / "node_modules",
    ROOT_DIR / "interface" / ".next",
)

WORKTREE_AREA_RULES = (
    {
        "name": "Core runtime",
        "prefixes": ("core/", "proto/"),
        "installs": ("cd core && go mod download",),
        "commands": (
            "uv run inv core.test",
            "uv run inv core.compile",
        ),
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
        "exact_paths": ("README.md", ".state/V7_DEV_STATE.md", ".state/V8_DEV_STATE.md", "architecture/mycelis-architecture-v7.md"),
        "installs": (),
        "commands": (
            "$env:PYTHONPATH='.'; uv run pytest tests/test_docs_links.py -q",
        ),
    },
    {
        "name": "Infra and deploy",
        "prefixes": ("charts/", "deploy/", "k8s/"),
        "installs": (),
        "commands": (),
    },
)

def _repo_relative(path: Path) -> str:
    try:
        return str(path.resolve().relative_to(ROOT_DIR.resolve())).replace("\\", "/")
    except ValueError:
        return str(path)


def _assert_repo_managed_target(path: Path) -> Path:
    resolved = path.resolve()
    try:
        resolved.relative_to(ROOT_DIR.resolve())
    except ValueError as exc:
        raise SystemExit(f"CLEANUP FAILED: refusing to touch non-repo path: {path}") from exc
    return resolved


def _artifact_size_bytes(path: Path) -> int:
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


def _format_size_bytes(size_bytes: int) -> str:
    units = ("B", "KB", "MB", "GB", "TB")
    value = float(size_bytes)
    for unit in units:
        if value < 1024 or unit == units[-1]:
            return f"{value:.1f} {unit}"
        value /= 1024
    return f"{size_bytes} B"


def _remove_repo_targets(targets: tuple[Path, ...]) -> tuple[list[str], list[str]]:
    removed: list[str] = []
    missing: list[str] = []
    for target in targets:
        managed_target = _assert_repo_managed_target(target)
        label = _repo_relative(target)
        if not managed_target.exists():
            missing.append(label)
            continue
        if managed_target.is_file():
            managed_target.unlink()
        else:
            shutil.rmtree(managed_target)
        removed.append(label)
    return removed, missing


def _report_repo_targets(targets: tuple[Path, ...]) -> list[dict[str, object]]:
    report: list[dict[str, object]] = []
    for target in targets:
        managed_target = _assert_repo_managed_target(target)
        exists = managed_target.exists()
        report.append(
            {
                "path": _repo_relative(target),
                "exists": exists,
                "bytes": _artifact_size_bytes(managed_target) if exists else 0,
            }
        )
    return report


def _print_cleanup_summary(removed: list[str], missing: list[str]) -> None:
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


# -- CLEAN --
@task
def legacy(c):
    """Remove legacy build files."""
    legacy_files = [
        ROOT_DIR / "Makefile",
        ROOT_DIR / "Makefile.legacy",
    ]
    for p in legacy_files:
        if p.exists():
            p.unlink()
            print(f"Removed {p}")


@task(name="generated")
def clean_generated(c):
    """Remove repo-local generated artifacts that should not persist across host boundaries."""
    removed, missing = _remove_repo_targets(GENERATED_ARTIFACT_TARGETS)
    print("=== CLEAN GENERATED ===")
    _print_cleanup_summary(removed, missing)
    print("Runtime data note:")
    print("  - workspace/docker-compose/data is intentionally untouched.")
    print("Workflow note:")
    print("  - keep heavy build/test artifacts in the WSL checkout; keep the Windows repo source-only.")


@task(name="reports")
def clean_reports(c):
    """Remove lightweight test/report artifacts without clearing install caches."""
    removed, missing = _remove_repo_targets(REPORT_ARTIFACT_TARGETS)
    print("=== CLEAN REPORTS ===")
    _print_cleanup_summary(removed, missing)


@task(name="wsl-handoff")
def clean_wsl_handoff(c):
    """Reset cross-host generated artifacts before handing the repo off to WSL."""
    removed, missing = _remove_repo_targets(WSL_HANDOFF_TARGETS)
    print("=== CLEAN WSL HANDOFF ===")
    _print_cleanup_summary(removed, missing)
    print("Next step:")
    print("  - use a WSL-native checkout for uv/npm/build/test/compose work.")


@task(name="windows-dev-residue")
def clean_windows_dev_residue(c):
    """Remove heavy repo-local artifacts from the Windows editing checkout."""
    if not is_windows():
        raise SystemExit(
            "clean.windows-dev-residue is Windows-only. Use clean.generated from the WSL checkout instead."
        )
    removed, missing = _remove_repo_targets(GENERATED_ARTIFACT_TARGETS)
    print("=== CLEAN WINDOWS DEV RESIDUE ===")
    _print_cleanup_summary(removed, missing)
    print("Windows source-only reminder:")
    print("  - edit and commit here if needed, but run install/build/test/compose from the WSL checkout.")


@task(name="disk-status")
def clean_disk_status(c):
    """Report repo-local generated artifact usage and host-boundary cleanup guidance."""
    report = _report_repo_targets(GENERATED_ARTIFACT_TARGETS)
    total_bytes = sum(int(item["bytes"]) for item in report)

    print("=== CLEAN DISK STATUS ===")
    for item in report:
        presence = "present" if item["exists"] else "missing"
        print(
            f"  - {item['path']}: {presence} ({_format_size_bytes(int(item['bytes']))})"
        )
    print(f"Repo-local generated total: {_format_size_bytes(total_bytes)}")
    print("Storage boundary:")
    print("  - Windows should stay source-only; heavy artifacts belong in the WSL checkout.")
    print("  - Docker image/volume usage and WSL VHD slack space are outside repo cleanup.")
    print("Low-disk reminder:")
    print("  - run clean.generated first, then `wsl --shutdown`, then compact the WSL VHD from an elevated PowerShell when needed.")


ns_clean = Collection("clean")
ns_clean.add_task(legacy)
ns_clean.add_task(clean_generated)
ns_clean.add_task(clean_reports)
ns_clean.add_task(clean_wsl_handoff)
ns_clean.add_task(clean_windows_dev_residue)
ns_clean.add_task(clean_disk_status)

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
        if not raw_line.strip():
            continue
        if len(raw_line) < 4:
            continue
        entries.append(
            {
                "status": raw_line[:2].strip() or "??",
                "path": _normalize_git_path(raw_line[3:]),
            }
        )
    return entries


def _match_worktree_rule(path: str):
    for rule in WORKTREE_AREA_RULES:
        if path in rule.get("exact_paths", ()):
            return rule
        if any(path.startswith(prefix) for prefix in rule.get("prefixes", ())):
            return rule
    return {
        "name": "Unclassified",
        "installs": (),
        "commands": (),
    }


def _build_worktree_triage(status_output: str):
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

    areas = [
        {"name": name, "count": count}
        for name, count in sorted(area_counts.items())
    ]

    return {
        "entries": entries,
        "areas": areas,
        "priority_installs": _unique_strings(priority_installs),
        "recommended_commands": _unique_strings(recommended_commands),
    }


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
    entries = triage["entries"]

    if entries:
        print(f"Working tree: dirty ({len(entries)} changed path(s))")
    else:
        print("Working tree: clean")

    print("\nExpected review targets:")
    for target in WORKTREE_REVIEW_TARGETS:
        print(f"  - {target}")

    print("\nDependency reset:")
    for command in WORKTREE_BASELINE_INSTALLS:
        print(f"  - {command}")

    if is_windows():
        print("\nWindows host note:")
        print("  - treat this checkout as source-only and run heavy validation from the WSL checkout.")

    if triage["priority_installs"]:
        print("\nPriority install checks:")
        for command in triage["priority_installs"]:
            print(f"  - {command}")
    else:
        print("\nPriority install checks:")
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


def _read_nats_line(sock) -> str:
    data = bytearray()
    while True:
        chunk = sock.recv(1)
        if not chunk:
            raise ConnectionError("NATS connection closed")
        data.extend(chunk)
        if data.endswith(b"\r\n"):
            return data[:-2].decode("utf-8", errors="replace")


def _drain_nats_messages(sock, timeout_seconds: float) -> list[tuple[str, str]]:
    import socket
    import time

    messages: list[tuple[str, str]] = []
    deadline = time.time() + timeout_seconds
    while time.time() < deadline:
        try:
            sock.settimeout(max(0.1, deadline - time.time()))
            line = _read_nats_line(sock)
        except socket.timeout:
            break
        if not line:
            continue
        if line == "PING":
            sock.sendall(b"PONG\r\n")
            continue
        if not line.startswith("MSG "):
            continue

        parts = line.split()
        subject = parts[1]
        payload_len = int(parts[-1])
        payload = bytearray()
        while len(payload) < payload_len:
            chunk = sock.recv(payload_len - len(payload))
            if not chunk:
                raise ConnectionError("NATS connection closed during payload read")
            payload.extend(chunk)
        trailer = sock.recv(2)
        if trailer != b"\r\n":
            raise ConnectionError("Invalid NATS message trailer")
        messages.append((subject, payload.decode("utf-8", errors="replace")))

    return messages


def _format_sync_reply(message: str) -> str:
    import json

    text = message.strip()
    if not text:
        return text

    try:
        payload = json.loads(text)
    except json.JSONDecodeError:
        return text

    if isinstance(payload, dict):
        reply_text = payload.get("text")
        if isinstance(reply_text, str) and reply_text.strip():
            return reply_text.strip()
        nested = payload.get("payload")
        if isinstance(nested, str) and nested.strip():
            return nested.strip()
        if isinstance(nested, dict):
            nested_text = nested.get("text")
            if isinstance(nested_text, str) and nested_text.strip():
                return nested_text.strip()
    return text


@task(name="architecture-sync")
def architecture_sync(c, timeout=12):
    """Synchronize architect, development, and AGUI teams over the NATS bus."""
    import socket
    import time

    directives = {
        "prime-architect": {
            "command_subject": "swarm.team.prime-architect.internal.command",
            "reply_subjects": [
                "swarm.team.prime-architect.signal.status",
                "swarm.team.prime-architect.signal.result",
            ],
            "message": (
                "Central architecture directive: keep the next-target workflow aligned to strict gate order. "
                "P0 remains the active phase. Require concrete test evidence before any phase advancement, "
                "coordinate development and AGUI work to the target goals, and reply with a concise execution brief. "
                "Do not use tools for this sync. Respond in plain text with at most 6 short lines."
            ),
        },
        "prime-development": {
            "command_subject": "swarm.team.prime-development.internal.command",
            "reply_subjects": [
                "swarm.team.prime-development.signal.status",
                "swarm.team.prime-development.signal.result",
            ],
            "message": (
                "Development directive: focus on P0 closure, memory-restart reliability, logging standardization, "
                "error-handling normalization, and no-regression verification. Go remains the primary implementation "
                "language for backend/runtime work. Python is limited to tasks, management scripting, and tests. "
                "Reply with the top implementation/testing priorities. Do not use tools for this sync. "
                "Respond in plain text with at most 5 short lines."
            ),
        },
        "agui-design-architect": {
            "command_subject": "swarm.team.agui-design-architect.internal.command",
            "reply_subjects": [
                "swarm.team.agui-design-architect.signal.status",
                "swarm.team.agui-design-architect.signal.result",
            ],
            "message": (
                "AGUI directive: align base UI updates to architecture truth. Prioritize workflow-composer onboarding, "
                "gate-state visibility, system status, team roster visibility, and operator-safe error presentation. "
                "Do not invent client-side workflow semantics that diverge from backend gates. "
                "Reply with the top UI architecture priorities. Do not use tools for this sync. "
                "Respond in plain text with at most 5 short lines."
            ),
        },
    }

    print("=== Team Architecture Sync ===")
    print("Transport: NATS")
    print("Role: central architect")
    print("Target: keep architect, development, and AGUI teams aligned to current goals and testing gates.\n")

    sock = socket.create_connection(("127.0.0.1", 4222), timeout=5)
    try:
        info_line = _read_nats_line(sock)
        if not info_line.startswith("INFO "):
            raise RuntimeError(f"Unexpected NATS handshake: {info_line}")

        sock.sendall(b"CONNECT {\"verbose\":false,\"pedantic\":false}\r\n")

        sid = 1
        for config in directives.values():
            for subject in config["reply_subjects"]:
                sock.sendall(f"SUB {subject} {sid}\r\n".encode("utf-8"))
                sid += 1

        sock.sendall(b"PING\r\n")
        if _read_nats_line(sock) != "PONG":
            raise RuntimeError("NATS did not acknowledge the subscription flush")

        for team_id, config in directives.items():
            payload = config["message"].encode("utf-8")
            sock.sendall(
                f"PUB {config['command_subject']} {len(payload)}\r\n".encode("utf-8")
            )
            sock.sendall(payload + b"\r\n")
            print(f"Dispatched architecture directive to {team_id}")

        sock.sendall(b"PING\r\n")
        if _read_nats_line(sock) != "PONG":
            raise RuntimeError("NATS did not acknowledge the publish flush")

        acknowledgements: list[tuple[str, str]] = []
        responded_teams: set[str] = set()
        deadline = time.time() + float(timeout)
        while time.time() < deadline and len(responded_teams) < len(directives):
            new_messages = _drain_nats_messages(sock, 0.8)
            acknowledgements.extend(new_messages)
            for subject, _ in new_messages:
                for team_id, config in directives.items():
                    if subject in config["reply_subjects"]:
                        responded_teams.add(team_id)
            time.sleep(0.2)

        for subject, message in acknowledgements:
            print(f"[reply] {subject}: {_format_sync_reply(message)}")

        missing = [
            team_id
            for team_id in directives
            if team_id not in responded_teams
        ]
        if missing:
            print("\nMissing team replies:")
            for team_id in missing:
                print(f"  - {team_id}")
        else:
            print("\nAll teams replied inside the sync window.")
    finally:
        try:
            sock.sendall(b"QUIT\r\n")
        except OSError:
            pass
        sock.close()

ns_team = Collection("team")
ns_team.add_task(architecture_sync)
ns_team.add_task(worktree_triage)
