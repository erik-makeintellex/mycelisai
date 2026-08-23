from __future__ import annotations

from pathlib import Path

from invoke import Context

from ops import db as db_tasks
from ops import lifecycle
from ops import lifecycle_first_boot


def test_first_boot_proof_resets_restarts_and_checks_clean_state(monkeypatch):
    events: list[str] = []
    bootstrap_snapshots = iter([
        {"mcp_servers": 2, "mcp_tools": 14, "nodes": 11},
        {"mcp_servers": 2, "mcp_tools": 14, "nodes": 11},
    ])

    monkeypatch.setattr(lifecycle, "down", lambda _c: events.append("down"))
    monkeypatch.setattr(lifecycle, "_ensure_bridge", lambda: events.append("bridge"))
    monkeypatch.setattr(
        lifecycle_first_boot,
        "_wait_for_port",
        lambda port, label, timeout=30, interval=1.0, host="127.0.0.1": events.append(f"wait:{label}") or True,
    )
    monkeypatch.setattr(lifecycle_first_boot.lifecycle_infra, "database_endpoint", lambda _root: ("127.0.0.1", 15432))
    monkeypatch.setattr(db_tasks, "reset", lambda _c: events.append("db.reset"))
    monkeypatch.setattr(
        lifecycle_first_boot,
        "_clean_first_boot_workspace_roots",
        lambda: events.append("workspace.clean") or Path("E:/mycelis/core/workspace"),
    )
    monkeypatch.setattr(
        lifecycle_first_boot,
        "_assert_clean_first_boot_user_tables",
        lambda label: events.append(f"user.empty:{label}") or {},
    )
    monkeypatch.setattr(lifecycle_first_boot, "_assert_jetstream_empty", lambda: events.append("nats.empty"))
    monkeypatch.setattr(lifecycle, "up", lambda _c, frontend=False, build=False: events.append(f"up:{build}:{frontend}"))
    monkeypatch.setattr(lifecycle, "health", lambda _c: events.append("health"))
    monkeypatch.setattr(lifecycle_first_boot, "_bootstrap_counts", lambda: next(bootstrap_snapshots))
    monkeypatch.setattr(
        lifecycle_first_boot,
        "_assert_bootstrap_counts_stable",
        lambda before, after: events.append(f"bootstrap.stable:{before == after}"),
    )

    lifecycle.first_boot_proof.body(Context(), build=True, frontend=True, shutdown=True)

    assert events == [
        "down",
        "bridge",
        "wait:PostgreSQL",
        "wait:NATS",
        "db.reset",
        "workspace.clean",
        "user.empty:database reset",
        "nats.empty",
        "up:True:True",
        "health",
        "user.empty:first boot",
        "nats.empty",
        "down",
        "up:False:True",
        "health",
        "user.empty:restart",
        "bootstrap.stable:True",
        "nats.empty",
        "down",
    ]


def test_first_boot_proof_can_leave_services_running(monkeypatch):
    events: list[str] = []
    monkeypatch.setattr(lifecycle, "down", lambda _c: events.append("down"))
    monkeypatch.setattr(lifecycle, "_ensure_bridge", lambda: None)
    monkeypatch.setattr(lifecycle_first_boot, "_wait_for_port", lambda *args, **kwargs: True)
    monkeypatch.setattr(lifecycle_first_boot.lifecycle_infra, "database_endpoint", lambda _root: ("127.0.0.1", 15432))
    monkeypatch.setattr(db_tasks, "reset", lambda _c: None)
    monkeypatch.setattr(lifecycle_first_boot, "_clean_first_boot_workspace_roots", lambda: Path("E:/mycelis/core/workspace"))
    monkeypatch.setattr(lifecycle_first_boot, "_assert_clean_first_boot_user_tables", lambda _label: {})
    monkeypatch.setattr(lifecycle_first_boot, "_assert_jetstream_empty", lambda: None)
    monkeypatch.setattr(lifecycle, "up", lambda _c, frontend=False, build=False: None)
    monkeypatch.setattr(lifecycle, "health", lambda _c: None)
    monkeypatch.setattr(lifecycle_first_boot, "_bootstrap_counts", lambda: {})
    monkeypatch.setattr(lifecycle_first_boot, "_assert_bootstrap_counts_stable", lambda _before, _after: None)

    lifecycle.first_boot_proof.body(Context(), build=False, frontend=False, shutdown=False)

    assert events == ["down", "down"]
