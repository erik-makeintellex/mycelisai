from __future__ import annotations

import pytest

from ops import compose


def test_compose_infra_up_prints_connection_guidance(monkeypatch, capsys):
    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(
        compose,
        "_load_compose_env",
        lambda: {
            "MYCELIS_COMPOSE_POSTGRES_PORT": "15432",
            "MYCELIS_COMPOSE_NATS_PORT": "14222",
            "MYCELIS_COMPOSE_NATS_MONITOR_PORT": "18222",
            "DB_USER": "owner",
            "DB_PASSWORD": "secret",
            "DB_NAME": "ownerdb",
        },
    )
    monkeypatch.setattr(compose, "_validate_compose_env", lambda env_values: None)
    monkeypatch.setattr(
        compose,
        "_run_compose",
        lambda args, check=True, env=None: type("Result", (), {"returncode": 0})(),
    )
    monkeypatch.setattr(compose, "_wait_for_port", lambda port, label, timeout_seconds=60: True)
    monkeypatch.setattr(compose, "_wait_for_postgres_ready", lambda timeout_seconds=90, env_values=None: True)

    compose.infra_up.body(None, wait_timeout=120)

    out = capsys.readouterr().out
    assert "Core and Interface stay down" in out
    assert "DB_HOST=host.docker.internal" in out
    assert "DB_PORT=15432" in out
    assert "NATS_URL=nats://host.docker.internal:14222" in out
    assert "DB_USER=owner" in out
    assert "secret" not in out
    assert "DB_PASSWORD=<from .env; not printed>" in out
    assert "NATS monitor=http://127.0.0.1:18222/varz" in out


def test_compose_infra_health_checks_data_plane_only(monkeypatch, capsys):
    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(
        compose,
        "_load_compose_env",
        lambda: {
            "MYCELIS_COMPOSE_POSTGRES_PORT": "15432",
            "MYCELIS_COMPOSE_NATS_PORT": "14222",
            "MYCELIS_COMPOSE_NATS_MONITOR_PORT": "18222",
            "DB_USER": "owner",
            "DB_NAME": "ownerdb",
        },
    )
    monkeypatch.setattr(compose, "_validate_compose_env", lambda env_values: None)
    monkeypatch.setattr(compose, "_port_open", lambda port: port in {15432, 14222})
    monkeypatch.setattr(compose, "_wait_for_postgres_ready", lambda timeout_seconds=90, env_values=None: True)
    monkeypatch.setattr(compose, "_http_get", lambda url, timeout=3.0, headers=None: (200, "{}"))

    compose.infra_health.body(None)

    out = capsys.readouterr().out
    assert "Compose Data Plane Health" in out
    assert "PostgreSQL query" in out
    assert "NATS monitor" in out
    assert "Data plane healthy" in out
    assert "Core health" not in out
    assert "Frontend" not in out


def test_compose_infra_health_fails_without_nats_monitor(monkeypatch):
    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(compose, "_load_compose_env", lambda: {})
    monkeypatch.setattr(compose, "_validate_compose_env", lambda env_values: None)
    monkeypatch.setattr(compose, "_port_open", lambda port: True)
    monkeypatch.setattr(compose, "_wait_for_postgres_ready", lambda timeout_seconds=90, env_values=None: True)
    monkeypatch.setattr(compose, "_http_get", lambda url, timeout=3.0, headers=None: (0, "connection refused"))

    with pytest.raises(SystemExit) as excinfo:
        compose.infra_health.body(None)

    assert "NATS monitor did not answer" in str(excinfo.value) or excinfo.value.code == 1


def test_compose_storage_health_checks_long_term_storage(monkeypatch, capsys):
    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(compose, "_load_compose_env", lambda: {"DB_USER": "owner", "DB_NAME": "ownerdb"})
    monkeypatch.setattr(compose, "_validate_compose_env", lambda env_values: None)
    monkeypatch.setattr(
        compose,
        "_compose_storage_check_results",
        lambda env_values: [(label, True) for label, _sql in compose.COMPOSE_LONG_TERM_STORAGE_CHECKS],
    )

    compose.storage_health.body(None)

    out = capsys.readouterr().out
    assert "Long-Term Storage Health" in out
    assert "pgvector extension" in out
    assert "semantic context vectors" in out
    assert "conversation continuity" in out
    assert "managed exchange items" in out
    assert "Long-term storage ready" in out


def test_compose_storage_health_guides_migration_when_missing(monkeypatch, capsys):
    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(compose, "_load_compose_env", lambda: {})
    monkeypatch.setattr(compose, "_validate_compose_env", lambda env_values: None)
    monkeypatch.setattr(
        compose,
        "_compose_storage_check_results",
        lambda env_values: [
            ("pgvector extension", True),
            ("semantic context vectors", False),
        ],
    )

    with pytest.raises(SystemExit) as excinfo:
        compose.storage_health.body(None)

    out = capsys.readouterr().out
    assert excinfo.value.code == 1
    assert "uv run inv compose.migrate" in out
    assert "semantic context vectors" in out


def test_compose_storage_check_results_batches_queries(monkeypatch):
    captured: list[str] = []

    def fake_psql(sql, env_values):
        captured.append(sql)
        return type(
            "Result",
            (),
            {
                "returncode": 0,
                "stdout": "pgvector extension\tok\nsemantic context vectors\tmissing\n",
                "stderr": "",
            },
        )()

    monkeypatch.setattr(compose, "_run_compose_psql", fake_psql)

    assert compose._compose_storage_check_results({}) == [
        ("pgvector extension", True),
        ("semantic context vectors", False),
    ]
    assert len(captured) == 1
    assert "context_vectors" in captured[0]


def test_compose_up_rejects_tiny_wait_timeout(monkeypatch):
    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(compose, "_load_compose_env", lambda: {})
    monkeypatch.setattr(compose, "_validate_compose_env", lambda env_values: None)
    monkeypatch.setattr(compose, "_prepare_wsl_ollama_host", lambda env_values: env_values)

    with pytest.raises(SystemExit) as excinfo:
        compose.up.body(None, wait_timeout=29)

    assert "at least 30 seconds" in str(excinfo.value)


def test_compose_up_postgres_timeout_has_guidance(monkeypatch):
    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(compose, "_load_compose_env", lambda: {})
    monkeypatch.setattr(compose, "_validate_compose_env", lambda env_values: None)
    monkeypatch.setattr(compose, "_prepare_wsl_ollama_host", lambda env_values: env_values)
    monkeypatch.setattr(
        compose,
        "_run_compose",
        lambda args, check=True, env=None: type("Result", (), {"returncode": 0})(),
    )
    monkeypatch.setattr(compose, "_wait_for_port", lambda port, label, timeout_seconds=60: False)

    with pytest.raises(SystemExit) as excinfo:
        compose.up.body(None, wait_timeout=120)

    message = str(excinfo.value)
    assert "PostgreSQL did not become reachable within 120s" in message
    assert "compose.logs postgres" in message
    assert "compose.down --volumes" in message


def test_compose_up_prints_expectations(monkeypatch, capsys):
    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(compose, "_load_compose_env", lambda: {})
    monkeypatch.setattr(compose, "_validate_compose_env", lambda env_values: None)
    monkeypatch.setattr(compose, "_prepare_wsl_ollama_host", lambda env_values: env_values)
    monkeypatch.setattr(
        compose,
        "_run_compose",
        lambda args, check=True, env=None: type("Result", (), {"returncode": 0})(),
    )
    monkeypatch.setattr(compose, "_run_compose_migrations", lambda: None)
    monkeypatch.setattr(compose, "_wait_for_port", lambda port, label, timeout_seconds=60: True)
    monkeypatch.setattr(compose, "_wait_for_postgres_ready", lambda timeout_seconds=90, env_values=None: True)
    monkeypatch.setattr(compose, "_wait_for_http_ok", lambda url, label, timeout_seconds=60, headers=None: True)
    monkeypatch.setattr(compose.status, "body", lambda _c=None: None)

    compose.up.body(None, build=True, wait_timeout=240)

    out = capsys.readouterr().out
    assert "[1/4] Starting PostgreSQL and NATS..." in out
    assert "With --build, image preparation can take several minutes" in out
    assert "[4/4] Compose stack ready." in out
    assert "uv run inv compose.health" in out


def test_compose_down_forwards_volumes_flag(monkeypatch):
    commands: list[list[str]] = []
    stopped: list[bool] = []

    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(compose, "_run_compose", lambda args, check=True, env=None: commands.append(args) or type("Result", (), {"returncode": 0})())
    monkeypatch.setattr(compose, "_stop_wsl_ollama_relay", lambda: stopped.append(True))

    compose.down.body(None, volumes=True)

    assert commands == [compose._compose_command("down", "--volumes")]
    assert stopped == [True]


def test_compose_health_fails_when_text_engine_is_offline(monkeypatch, capsys):
    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(compose, "_load_compose_env", lambda: {"MYCELIS_API_KEY": "test-key"})
    monkeypatch.setattr(compose, "_prepare_wsl_ollama_host", lambda env_values: env_values)

    responses = {
        f"http://{compose.API_HOST}:{compose.API_PORT}/healthz": (200, "ok"),
        f"http://{compose.API_HOST}:{compose.API_PORT}/api/v1/templates": (200, "ok"),
        f"http://{compose.API_HOST}:{compose.API_PORT}/api/v1/brains": (200, "ok"),
        f"http://{compose.API_HOST}:{compose.API_PORT}/api/v1/telemetry/compute": (200, "ok"),
        f"http://{compose.INTERFACE_HOST}:{compose.INTERFACE_PORT}/": (200, "ok"),
        "http://127.0.0.1:8222/varz": (200, "ok"),
        f"http://{compose.API_HOST}:{compose.API_PORT}/api/v1/cognitive/status": (
            200,
            '{"text":{"status":"offline"},"media":{"status":"offline"}}',
        ),
    }
    monkeypatch.setattr(compose, "_http_get", lambda url, timeout=3.0, headers=None: responses[url])

    with pytest.raises(SystemExit):
        compose.health.body(None)

    out = capsys.readouterr().out
    assert "Cognitive Engine" in out
    assert "text=offline" in out
