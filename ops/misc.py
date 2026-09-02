from pathlib import Path

from invoke import Collection, task

from .cleanup_support import filter_active_runtime_targets, print_active_runtime_skip
from .ci_pipeline import _console_safe
from .config import ROOT_DIR, is_windows
from .misc_support import (
    WORKTREE_BASELINE_INSTALLS,
    WORKTREE_REVIEW_TARGETS,
    architecture_sync_directives as _architecture_sync_directives,
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
    """Remove repo-local generated artifacts that should not persist across host boundaries."""
    targets, skipped = filter_active_runtime_targets(_generated_artifact_targets(ROOT_DIR), ROOT_DIR)
    removed, missing = remove_repo_targets(tuple(targets), ROOT_DIR)
    print("=== CLEAN GENERATED ===")
    print_cleanup_summary(removed, missing)
    print_active_runtime_skip(skipped)
    print("Runtime data note:")
    print("  - workspace/docker-compose/data is intentionally untouched.")
    print("Workflow note:")
    print("  - keep heavy build/test artifacts in the WSL checkout; keep the Windows repo source-only.")


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


@task(name="windows-dev-residue")
def clean_windows_dev_residue(c):
    """Remove heavy repo-local artifacts from the Windows editing checkout."""
    if not is_windows():
        raise SystemExit(
            "clean.windows-dev-residue is Windows-only. Use clean.generated from the WSL checkout instead."
        )
    targets, skipped = filter_active_runtime_targets(_generated_artifact_targets(ROOT_DIR), ROOT_DIR)
    removed, missing = remove_repo_targets(tuple(targets), ROOT_DIR)
    print("=== CLEAN WINDOWS DEV RESIDUE ===")
    print_cleanup_summary(removed, missing)
    print_active_runtime_skip(skipped)
    print("Windows source-only reminder:")
    print("  - edit and commit here if needed, but run install/build/test/compose from the WSL checkout.")


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
    print("  - Windows should stay source-only; heavy artifacts belong in the WSL checkout.")
    print("  - Docker image/volume usage and WSL VHD slack space are outside repo cleanup.")
    print("Low-disk reminder:")
    print("  - run clean.generated first, then `wsl --shutdown`, then compact the WSL VHD from an elevated PowerShell when needed.")

ns_clean = Collection("clean")
ns_clean.add_task(clean_generated)
ns_clean.add_task(clean_reports)
ns_clean.add_task(clean_wsl_handoff)
ns_clean.add_task(clean_windows_dev_residue)
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
        is_windows(),
        review_targets=WORKTREE_REVIEW_TARGETS,
        baseline_installs=WORKTREE_BASELINE_INSTALLS,
    )


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


def _format_sync_output(subject: str, message: str, encoding: str | None = None) -> str:
    return _console_safe(f"[reply] {subject}: {_format_sync_reply(message)}", encoding)


@task(name="architecture-sync")
def architecture_sync(c, timeout=12):
    """Synchronize architect, development, and AGUI teams over the NATS bus."""
    import socket
    import time

    directives = _architecture_sync_directives()

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
            print(_format_sync_output(subject, message))

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
