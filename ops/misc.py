from pathlib import Path

from invoke import Collection, task

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


@task(name="generated")
def clean_generated(c):
    """Remove repo-local generated artifacts that should not persist across host boundaries."""
    removed, missing = remove_repo_targets(GENERATED_ARTIFACT_TARGETS, ROOT_DIR)
    print("=== CLEAN GENERATED ===")
    print_cleanup_summary(removed, missing)
    print("Runtime data note:")
    print("  - workspace/docker-compose/data is intentionally untouched.")
    print("Workflow note:")
    print("  - keep heavy build/test artifacts in the WSL checkout; keep the Windows repo source-only.")


@task(name="reports")
def clean_reports(c):
    """Remove lightweight test/report artifacts without clearing install caches."""
    removed, missing = remove_repo_targets(REPORT_ARTIFACT_TARGETS, ROOT_DIR)
    print("=== CLEAN REPORTS ===")
    print_cleanup_summary(removed, missing)


@task(name="wsl-handoff")
def clean_wsl_handoff(c):
    """Reset cross-host generated artifacts before handing the repo off to WSL."""
    removed, missing = remove_repo_targets(WSL_HANDOFF_TARGETS, ROOT_DIR)
    print("=== CLEAN WSL HANDOFF ===")
    print_cleanup_summary(removed, missing)
    print("Next step:")
    print("  - use a WSL-native checkout for uv/npm/build/test/compose work.")


@task(name="windows-dev-residue")
def clean_windows_dev_residue(c):
    """Remove heavy repo-local artifacts from the Windows editing checkout."""
    if not is_windows():
        raise SystemExit(
            "clean.windows-dev-residue is Windows-only. Use clean.generated from the WSL checkout instead."
        )
    removed, missing = remove_repo_targets(GENERATED_ARTIFACT_TARGETS, ROOT_DIR)
    print("=== CLEAN WINDOWS DEV RESIDUE ===")
    print_cleanup_summary(removed, missing)
    print("Windows source-only reminder:")
    print("  - edit and commit here if needed, but run install/build/test/compose from the WSL checkout.")


@task(name="disk-status")
def clean_disk_status(c):
    """Report repo-local generated artifact usage and host-boundary cleanup guidance."""
    report = report_repo_targets(GENERATED_ARTIFACT_TARGETS, ROOT_DIR)
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
