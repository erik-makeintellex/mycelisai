from __future__ import annotations

from pathlib import Path

from invoke import Context

from ops import db as db_tasks
from ops import lifecycle


def test_core_council_ready_requires_success_response(monkeypatch):
    calls: list[tuple[str, dict[str, str] | None]] = []
    monkeypatch.setenv("MYCELIS_API_KEY", "test-key")
    monkeypatch.setattr(
        lifecycle,
        "_http_get",
        lambda url, timeout=5.0, headers=None: calls.append((url, headers))
        or (200, '{"ok":true}'),
    )

    assert lifecycle._core_council_ready(timeout=1, interval=0.01)
    assert calls == [
        (
            f"http://{lifecycle.API_HOST}:{lifecycle.API_PORT}/api/v1/council/members",
            {"Authorization": "Bearer test-key"},
        )
    ]


def test_up_frontend_uses_shared_interface_launcher(monkeypatch):
    events: list[str] = []

    monkeypatch.setattr(Path, "exists", lambda self: True)
    monkeypatch.setattr(lifecycle, "_ensure_bridge", lambda: None)
    monkeypatch.setattr(lifecycle, "_wait_for_http_ok", lambda *args, **kwargs: True)
    monkeypatch.setattr(lifecycle, "_core_council_ready", lambda timeout=10, interval=1.0: True)
    monkeypatch.setattr(db_tasks.create, "body", lambda _c: None)
    monkeypatch.setattr(lifecycle, "_start_core_background", lambda: True)

    def fake_wait_for_port(port, label, timeout=30, interval=1.0):
        events.append(f"wait:{port}:{label}")
        return True

    def fake_port_open(port: int, host: str = "127.0.0.1", timeout: float = 1.0) -> bool:
        if port in (5432, 4222):
            return True
        if port in (lifecycle.API_PORT, lifecycle.INTERFACE_PORT):
            return False
        return False

    monkeypatch.setattr(lifecycle, "_wait_for_port", fake_wait_for_port)
    monkeypatch.setattr(lifecycle, "_port_open", fake_port_open)

    from ops import interface as interface_tasks

    monkeypatch.setattr(interface_tasks, "interface_task_env", lambda extra=None: {"TEST_ENV": "1"})
    monkeypatch.setattr(
        interface_tasks,
        "start_dev_server_detached",
        lambda env=None, host=lifecycle.INTERFACE_BIND_HOST, port=lifecycle.INTERFACE_PORT: events.append(
            f"frontend:{host}:{port}:{env.get('TEST_ENV') if env else ''}"
        ),
    )

    lifecycle.up.body(Context(), frontend=True, build=False)

    assert f"frontend:{lifecycle.INTERFACE_BIND_HOST}:{lifecycle.INTERFACE_PORT}:1" in events
    assert f"wait:{lifecycle.INTERFACE_PORT}:Frontend" in events


def test_up_with_build_uses_core_compile_task_body(monkeypatch):
    compile_calls: list[str] = []

    monkeypatch.setattr(lifecycle, "_ensure_bridge", lambda: None)
    monkeypatch.setattr(lifecycle, "_wait_for_port", lambda *args, **kwargs: True)
    monkeypatch.setattr(lifecycle, "_wait_for_http_ok", lambda *args, **kwargs: True)
    monkeypatch.setattr(lifecycle, "_core_council_ready", lambda timeout=10, interval=1.0: True)
    monkeypatch.setattr(
        lifecycle,
        "_port_open",
        lambda port, host="127.0.0.1", timeout=1.0: False if port == lifecycle.API_PORT else True,
    )
    monkeypatch.setattr(db_tasks.create, "body", lambda _c: None)
    monkeypatch.setattr(lifecycle, "_start_core_background", lambda: True)

    from ops import core as core_tasks

    monkeypatch.setattr(core_tasks.compile, "body", lambda _c: compile_calls.append("compile"))

    lifecycle.up.body(Context(), frontend=False, build=True)

    assert compile_calls == ["compile"]
