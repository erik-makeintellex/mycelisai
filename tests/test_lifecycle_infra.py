from invoke import Context
import pytest

from ops import compose, lifecycle


def test_dev_infra_mode_defaults_to_compose(monkeypatch, tmp_path):
    monkeypatch.delenv("MYCELIS_DEV_INFRA_MODE", raising=False)

    assert lifecycle.lifecycle_infra.dev_infra_mode(tmp_path) == "compose"


def test_dev_infra_mode_rejects_native_host_services(monkeypatch, tmp_path):
    monkeypatch.setenv("MYCELIS_DEV_INFRA_MODE", "native")

    with pytest.raises(SystemExit, match="native host services are unsupported"):
        lifecycle.lifecycle_infra.dev_infra_mode(tmp_path)


def test_ensure_bridge_uses_compose_data_plane_without_building_apps(monkeypatch):
    calls: list[tuple[object, int, bool]] = []
    monkeypatch.setenv("MYCELIS_DEV_INFRA_MODE", "compose")
    monkeypatch.setattr(
        compose.infra_up,
        "body",
        lambda context, wait_timeout=180, migrate=False: calls.append((context, wait_timeout, migrate)),
    )

    lifecycle._ensure_bridge()

    assert calls == [(None, 180, False)]


def test_compose_down_cleanup_leaves_data_plane_ports_managed_by_docker(monkeypatch):
    monkeypatch.setenv("MYCELIS_DEV_INFRA_MODE", "compose")
    monkeypatch.setattr(lifecycle, "_port_open", lambda *args, **kwargs: True)

    assert lifecycle._remaining_managed_services() == ["Core API", "Frontend"]
    assert lifecycle._service_keys_by_label(["PostgreSQL", "NATS", "Core API", "Frontend"]) == [
        "core",
        "frontend",
    ]


def test_status_reports_compose_data_plane_with_local_apps(monkeypatch, capsys):
    monkeypatch.setenv("MYCELIS_DEV_INFRA_MODE", "compose")
    monkeypatch.setattr(lifecycle, "_port_open", lambda *args, **kwargs: False)
    monkeypatch.setattr(lifecycle, "_http_get", lambda *args, **kwargs: (0, "offline"))
    monkeypatch.setattr(lifecycle, "_list_compiled_go_service_processes", lambda: [])
    monkeypatch.setattr(lifecycle.lifecycle_infra, "database_endpoint", lambda _root: ("127.0.0.1", 15432))

    lifecycle.status.body(Context())

    output = capsys.readouterr().out
    assert "Dev infra mode  : compose" in output
    assert "Development     : Docker PostgreSQL/NATS + local Core/Interface" in output
    assert "Full containers : explicit proof via compose.up; Kubernetes via k8s.*" in output
    assert "PostgreSQL      : DOWN  [127.0.0.1:15432]" in output
