from __future__ import annotations

from pathlib import Path

from invoke import Context
import pytest

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


def test_isolated_first_boot_uses_only_fixture_runner(monkeypatch):
    calls = []
    monkeypatch.setattr(lifecycle, "down", lambda _c: pytest.fail("source lifecycle down called"))
    monkeypatch.setattr(db_tasks, "reset", lambda _c: pytest.fail("development database reset called"))
    monkeypatch.setattr(
        "ops.lifecycle_first_boot_isolated.run_isolated_first_boot",
        lambda **kwargs: calls.append(kwargs),
    )

    lifecycle.first_boot_proof.body(Context(), isolated=True, build=True)

    assert calls == [{
        "build": True,
        "user_tables": lifecycle_first_boot.CLEAN_FIRST_BOOT_USER_TABLES,
        "bootstrap_tables": lifecycle_first_boot.CLEAN_FIRST_BOOT_BOOTSTRAP_TABLES,
        "framework_runs": False,
    }]


def test_framework_runs_first_boot_requires_isolated_fixture(monkeypatch):
    monkeypatch.setattr(lifecycle, "down", lambda _c: pytest.fail("retained lifecycle touched"))
    monkeypatch.setattr(db_tasks, "reset", lambda _c: pytest.fail("retained database touched"))
    with pytest.raises(SystemExit, match="requires --isolated"):
        lifecycle.first_boot_proof.body(Context(), framework_runs=True)


def test_framework_runs_flag_uses_canonical_isolated_runner(monkeypatch):
    calls = []
    monkeypatch.setattr(
        "ops.lifecycle_first_boot_isolated.run_isolated_first_boot",
        lambda **kwargs: calls.append(kwargs),
    )
    lifecycle.first_boot_proof.body(Context(), isolated=True, framework_runs=True, build=True)
    assert calls == [{
        "build": True,
        "user_tables": lifecycle_first_boot.CLEAN_FIRST_BOOT_USER_TABLES,
        "bootstrap_tables": lifecycle_first_boot.CLEAN_FIRST_BOOT_BOOTSTRAP_TABLES,
        "framework_runs": True,
    }]


def test_framework_runs_certification_rejects_image_reuse(monkeypatch):
    monkeypatch.setattr(
        "ops.lifecycle_first_boot_isolated.run_isolated_first_boot",
        lambda **_kwargs: pytest.fail("stale worker image was accepted"),
    )
    with pytest.raises(SystemExit, match="requires --build"):
        lifecycle.first_boot_proof.body(Context(), isolated=True, framework_runs=True)


@pytest.mark.parametrize("kwargs", [{"frontend": False}, {"shutdown": False}])
def test_isolated_first_boot_requires_full_proof_and_cleanup(monkeypatch, kwargs):
    monkeypatch.setattr(
        "ops.lifecycle_first_boot_isolated.run_isolated_first_boot",
        lambda **_kwargs: pytest.fail("fixture started for an incomplete proof"),
    )
    with pytest.raises(SystemExit, match="requires frontend and shutdown"):
        lifecycle.first_boot_proof.body(Context(), isolated=True, **kwargs)
