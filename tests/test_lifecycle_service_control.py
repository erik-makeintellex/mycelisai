from invoke import Context

from ops import lifecycle
from ops import lifecycle_infra


def _stub_local_shutdown(monkeypatch) -> None:
    from ops import interface as interface_tasks

    monkeypatch.setattr(lifecycle, "_kill_port", lambda *_args, **_kwargs: False)
    monkeypatch.setattr(lifecycle, "_wait_for_port_closed", lambda *_args, **_kwargs: True)
    monkeypatch.setattr(lifecycle, "_kill_compiled_go_services", lambda: [])
    monkeypatch.setattr(lifecycle, "_run_best_effort", lambda *_args, **_kwargs: None)
    monkeypatch.setattr(lifecycle, "_port_open", lambda *_args, **_kwargs: False)
    monkeypatch.setattr(lifecycle, "is_windows", lambda: False)
    monkeypatch.setattr(interface_tasks, "_cleanup_repo_local_interface_processes", lambda: [])


def test_lifecycle_down_retains_compose_data_plane_by_default(monkeypatch, capsys):
    _stub_local_shutdown(monkeypatch)
    monkeypatch.setenv("MYCELIS_DEV_INFRA_MODE", "compose")
    monkeypatch.setattr(
        lifecycle_infra,
        "stop_compose_data_plane",
        lambda _context: (_ for _ in ()).throw(AssertionError("data plane should remain running")),
    )

    lifecycle.down.body(Context(), include_data_plane=False)

    output = capsys.readouterr().out
    assert "Compose PostgreSQL and NATS remain running for reuse" in output
    assert "All services stopped" not in output


def test_lifecycle_down_can_stop_compose_data_plane(monkeypatch, capsys):
    _stub_local_shutdown(monkeypatch)
    monkeypatch.setenv("MYCELIS_DEV_INFRA_MODE", "compose")
    calls: list[Context] = []
    monkeypatch.setattr(lifecycle_infra, "stop_compose_data_plane", calls.append)

    lifecycle.down.body(Context(), include_data_plane=True)

    assert len(calls) == 1
    assert "Compose data plane stopped" in capsys.readouterr().out
