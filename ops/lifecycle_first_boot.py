from __future__ import annotations

import json
import os
import shutil
import socket
import urllib.error
import urllib.request
from pathlib import Path

from invoke import task

from . import db as db_tasks
from . import lifecycle_infra
from .config import CORE_DIR, ROOT_DIR

CLEAN_FIRST_BOOT_USER_TABLES = (
    "artifacts",
    "collaboration_groups",
    "context_vectors",
    "conversation_summaries",
    "conversation_turns",
    "groups",
    "mission_runs",
    "outcome_projects",
    "proof_artifacts",
    "runtime_team_manifests",
)
CLEAN_FIRST_BOOT_BOOTSTRAP_TABLES = (
    "exchange_capability_registry",
    "exchange_channels",
    "exchange_field_registry",
    "exchange_schema_registry",
    "mcp_servers",
    "mcp_tool_sets",
    "mcp_tools",
    "nodes",
)
CLEAN_FIRST_BOOT_WORKSPACE_DIRS = ("groups", "generated", "artifacts", "saved-media")


def _load_env():
    from dotenv import load_dotenv

    load_dotenv(str(ROOT_DIR / ".env"), override=True)


def _wait_for_port(port: int, label: str, timeout: float = 30, host: str = "127.0.0.1") -> bool:
    import time

    deadline = time.time() + timeout
    while time.time() < deadline:
        try:
            with socket.create_connection((host, port), timeout=1.0):
                print(f"  [OK] {label} reachable on {host}:{port}")
                return True
        except OSError:
            time.sleep(1.0)
    return False


def _http_get(url: str, timeout: float = 5.0) -> tuple[int, str]:
    try:
        with urllib.request.urlopen(url, timeout=timeout) as resp:
            return resp.status, resp.read().decode("utf-8", errors="replace")
    except urllib.error.HTTPError as exc:
        return exc.code, str(exc)
    except Exception as exc:
        return 0, str(exc)


def _backend_workspace_root() -> Path:
    _load_env()
    workspace = os.environ.get("MYCELIS_WORKSPACE", "./workspace").strip() or "./workspace"
    path = Path(workspace)
    if path.is_absolute():
        return path.resolve()
    return (CORE_DIR / path).resolve()


def _remove_child_dir(root: Path, child: str):
    target = (root / child).resolve()
    if root not in target.parents:
        raise SystemExit(f"Refusing to clean path outside workspace root: {target}")
    if not target.exists():
        return
    if not target.is_dir():
        raise SystemExit(f"Refusing to remove non-directory workspace path: {target}")
    shutil.rmtree(target)


def _clean_first_boot_workspace_roots() -> Path:
    workspace_root = _backend_workspace_root()
    workspace_root.mkdir(parents=True, exist_ok=True)
    for child in CLEAN_FIRST_BOOT_WORKSPACE_DIRS:
        _remove_child_dir(workspace_root, child)
    return workspace_root


def _psql_count_map(table_names: tuple[str, ...]) -> dict[str, int]:
    selects = [f"SELECT '{name}' AS name, COUNT(*)::bigint AS count FROM {name}" for name in table_names]
    result = db_tasks._run_psql(sql=" UNION ALL ".join(selects) + " ORDER BY name;")
    db_tasks._emit_psql_output(result)
    if result.returncode != 0:
        raise SystemExit("Clean first-boot proof failed while reading product-state tables.")
    counts: dict[str, int] = {}
    for line in result.stdout.splitlines():
        line = line.strip()
        if not line or "|" not in line:
            continue
        name, raw_count = line.split("|", 1)
        raw_count = raw_count.strip()
        if not raw_count.lstrip("-").isdigit():
            continue
        try:
            counts[name.strip()] = int(raw_count)
        except ValueError as exc:
            raise SystemExit(f"Unexpected row-count output: {line}") from exc
    missing = [name for name in table_names if name not in counts]
    if missing:
        raise SystemExit("Clean first-boot proof missed table counts: " + ", ".join(missing))
    return counts


def _assert_clean_first_boot_user_tables(label: str) -> dict[str, int]:
    counts = _psql_count_map(CLEAN_FIRST_BOOT_USER_TABLES)
    dirty = {name: count for name, count in counts.items() if count != 0}
    if dirty:
        rendered = ", ".join(f"{name}={count}" for name, count in dirty.items())
        raise SystemExit(f"Clean first-boot proof failed after {label}: user state exists ({rendered}).")
    print(f"  [OK] No user product state after {label}")
    return counts


def _bootstrap_counts() -> dict[str, int]:
    return _psql_count_map(CLEAN_FIRST_BOOT_BOOTSTRAP_TABLES)


def _assert_bootstrap_counts_stable(before: dict[str, int], after: dict[str, int]):
    changed = {
        name: (before.get(name, 0), after.get(name, 0))
        for name in sorted(set(before) | set(after))
        if before.get(name, 0) != after.get(name, 0)
    }
    if changed:
        rendered = ", ".join(f"{name}: {old}->{new}" for name, (old, new) in changed.items())
        raise SystemExit(f"Clean first-boot proof failed: bootstrap row counts changed after restart ({rendered}).")
    print("  [OK] Bootstrap row counts are stable after restart")


def _assert_jetstream_empty():
    monitor_port = int(os.environ.get("MYCELIS_COMPOSE_NATS_MONITOR_PORT", "8222"))
    code, body = _http_get(f"http://127.0.0.1:{monitor_port}/jsz?streams=1")
    if code != 200:
        raise SystemExit(f"Clean first-boot proof failed: NATS monitor /jsz returned {code}.")
    try:
        payload = json.loads(body)
    except json.JSONDecodeError as exc:
        raise SystemExit("Clean first-boot proof failed: NATS monitor returned invalid JSON.") from exc
    streams = int(payload.get("streams", 0) or 0)
    messages = int(payload.get("messages", 0) or 0)
    if streams != 0 or messages != 0:
        raise SystemExit(
            f"Clean first-boot proof failed: NATS JetStream is not empty (streams={streams}, messages={messages})."
        )
    print("  [OK] NATS JetStream has no retained streams or messages")


@task(
    help={
        "build": "Build the Go binary before first startup (default: False).",
        "frontend": "Also start frontend during proof (default: True).",
        "shutdown": "Stop local app services after proof; data plane remains running (default: True).",
    }
)
def first_boot_proof(c, build=False, frontend=True, shutdown=True):
    """
    Prove a clean deployment can first-boot without historical product state.
    """
    from . import lifecycle

    print("=== Mycelis Clean First-Boot Proof ===\n")
    print("[1/8] Stopping local app services...")
    lifecycle.down(c)
    print()

    print("[2/8] Ensuring Dockerized data plane...")
    lifecycle._ensure_bridge()
    db_host, db_port = lifecycle_infra.database_endpoint(ROOT_DIR)
    if not _wait_for_port(db_port, "PostgreSQL", host=db_host):
        raise SystemExit(f"Clean first-boot proof failed: PostgreSQL is not reachable at {db_host}:{db_port}.")
    if not _wait_for_port(4222, "NATS"):
        raise SystemExit("Clean first-boot proof failed: NATS is not reachable on :4222.")
    print()

    print("[3/8] Resetting PostgreSQL/pgvector and generated workspace roots...")
    db_tasks.reset(c)
    workspace_root = _clean_first_boot_workspace_roots()
    print(f"  [OK] Workspace generated roots cleared under {workspace_root}")
    _assert_clean_first_boot_user_tables("database reset")
    _assert_jetstream_empty()
    print()

    print("[4/8] Starting candidate stack from clean state...")
    lifecycle.up(c, frontend=frontend, build=build)
    print()

    print("[5/8] Running health and empty-state probes...")
    lifecycle.health(c)
    _assert_clean_first_boot_user_tables("first boot")
    first_bootstrap = _bootstrap_counts()
    _assert_jetstream_empty()
    print()

    print("[6/8] Restarting Core/Interface to prove idempotent bootstrap...")
    lifecycle.down(c)
    print()
    lifecycle.up(c, frontend=frontend, build=False)
    print()

    print("[7/8] Rechecking health, user state, and bootstrap stability...")
    lifecycle.health(c)
    _assert_clean_first_boot_user_tables("restart")
    _assert_bootstrap_counts_stable(first_bootstrap, _bootstrap_counts())
    _assert_jetstream_empty()
    print()

    if shutdown:
        print("[8/8] Stopping local app services; data plane remains reusable...")
        lifecycle.down(c)
    else:
        print("[8/8] Leaving local app services running for follow-on proof.")

    print("\nCLEAN FIRST-BOOT PROOF READY.")
