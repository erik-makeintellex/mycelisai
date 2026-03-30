from __future__ import annotations

from pathlib import Path

import pytest

from ops import compose


def test_compose_command_includes_env_file_and_project_name():
    cmd = compose._compose_command("ps")

    assert cmd[:6] == [
        "docker",
        "compose",
        "--project-name",
        compose.COMPOSE_PROJECT,
        "--env-file",
        str(compose.COMPOSE_ENV_FILE),
    ]
    assert cmd[-1] == "ps"


def test_load_compose_env_parses_key_values(tmp_path, monkeypatch):
    env_file = tmp_path / ".env.compose"
    env_file.write_text(
        "\n".join(
            [
                "# comment",
                "MYCELIS_API_KEY=test-key",
                "DB_HOST=postgres",
                "",
            ]
        ),
        encoding="utf-8",
    )
    monkeypatch.setattr(compose, "COMPOSE_ENV_FILE", env_file)

    assert compose._load_compose_env() == {
        "MYCELIS_API_KEY": "test-key",
        "DB_HOST": "postgres",
    }


def test_require_compose_env_file_has_clear_guidance(tmp_path, monkeypatch):
    missing = tmp_path / ".env.compose"
    example = tmp_path / ".env.compose.example"
    example.write_text("MYCELIS_API_KEY=x\n", encoding="utf-8")
    monkeypatch.setattr(compose, "COMPOSE_ENV_FILE", missing)
    monkeypatch.setattr(compose, "COMPOSE_ENV_EXAMPLE", example)

    with pytest.raises(SystemExit) as excinfo:
        compose._require_compose_env_file()

    assert ".env.compose.example" in str(excinfo.value)
    assert "MYCELIS_API_KEY" in str(excinfo.value)


def test_run_compose_migrations_executes_canonical_files(monkeypatch):
    commands: list[list[str]] = []
    files = [Path("001_init_memory.sql"), Path("002_extra.up.sql")]

    monkeypatch.setattr(compose.db_tasks, "_migration_files", lambda: files)
    monkeypatch.setattr(
        compose,
        "_run_compose",
        lambda args, check=True: commands.append(args) or type("Result", (), {"returncode": 0})(),
    )

    compose._run_compose_migrations()

    assert commands == [
        compose._compose_command(
            "exec",
            "-T",
            "postgres",
            "psql",
            "-v",
            "ON_ERROR_STOP=1",
            "-h",
            "127.0.0.1",
            "-U",
            "mycelis",
            "-d",
            "cortex",
            "-f",
            "/migrations/001_init_memory.sql",
        ),
        compose._compose_command(
            "exec",
            "-T",
            "postgres",
            "psql",
            "-v",
            "ON_ERROR_STOP=1",
            "-h",
            "127.0.0.1",
            "-U",
            "mycelis",
            "-d",
            "cortex",
            "-f",
            "/migrations/002_extra.up.sql",
        ),
    ]


def test_compose_up_orders_infra_then_migrations_then_app(monkeypatch):
    commands: list[list[str]] = []
    waits: list[tuple[str, int | str]] = []

    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(
        compose,
        "_run_compose",
        lambda args, check=True: commands.append(args) or type("Result", (), {"returncode": 0})(),
    )
    monkeypatch.setattr(compose, "_run_compose_migrations", lambda: waits.append(("migrate", 0)))
    monkeypatch.setattr(compose, "_wait_for_port", lambda port, label, timeout_seconds=60: waits.append((label, port)) or True)
    monkeypatch.setattr(compose, "_wait_for_postgres_ready", lambda timeout_seconds=90: waits.append(("PostgreSQL ready", timeout_seconds)) or True)
    monkeypatch.setattr(
        compose,
        "_wait_for_http_ok",
        lambda url, label, timeout_seconds=60, headers=None: waits.append((label, url)) or True,
    )
    monkeypatch.setattr(compose.status, "body", lambda _c=None: waits.append(("status", 0)))

    compose.up.body(None, build=False)

    assert commands == [
        compose._compose_command("up", "-d", "postgres", "nats"),
        compose._compose_command("up", "-d", "core", "interface"),
    ]
    assert waits == [
        ("PostgreSQL", 5432),
        ("PostgreSQL ready", 90),
        ("NATS", 4222),
        ("migrate", 0),
        ("Core API", compose.API_PORT),
        ("Core health", f"http://{compose.API_HOST}:{compose.API_PORT}/healthz"),
        ("Frontend", compose.INTERFACE_PORT),
        ("Frontend", f"http://{compose.INTERFACE_HOST}:{compose.INTERFACE_PORT}/"),
        ("status", 0),
    ]


def test_validate_compose_env_rejects_loopback_ollama_host():
    with pytest.raises(SystemExit) as excinfo:
        compose._validate_compose_env({"MYCELIS_COMPOSE_OLLAMA_HOST": "http://127.0.0.1:11434"})

    assert "MYCELIS_COMPOSE_OLLAMA_HOST" in str(excinfo.value)
    assert "host.docker.internal" in str(excinfo.value)


def test_validate_compose_env_allows_host_gateway_alias():
    compose._validate_compose_env({"MYCELIS_COMPOSE_OLLAMA_HOST": "http://host.docker.internal:11434"})


def test_compose_health_fails_when_text_engine_is_offline(monkeypatch, capsys):
    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(compose, "_load_compose_env", lambda: {"MYCELIS_API_KEY": "test-key"})

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
