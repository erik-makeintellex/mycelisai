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
    monkeypatch.setattr(compose, "_compose_schema_bootstrapped", lambda: False)
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


def test_run_compose_migrations_skips_replay_when_schema_is_compatible(monkeypatch, capsys):
    monkeypatch.setattr(compose, "_compose_schema_bootstrapped", lambda: True)
    monkeypatch.setattr(
        compose.db_tasks,
        "_migration_files",
        lambda: pytest.fail("migration files should not be replayed for a compatible schema"),
    )

    compose._run_compose_migrations()

    out = capsys.readouterr().out
    assert "skipping forward migration replay" in out
    assert "compose.down --volumes" in out


def test_compose_schema_bootstrapped_requires_all_runtime_objects(monkeypatch):
    responses = iter([True, True, False])
    monkeypatch.setattr(compose.db_tasks, "SCHEMA_COMPATIBILITY_CHECKS", [("a", "sql"), ("b", "sql"), ("c", "sql")])
    monkeypatch.setattr(compose, "_compose_query_succeeds", lambda sql: next(responses))

    assert compose._compose_schema_bootstrapped() is False


def test_compose_schema_bootstrapped_accepts_current_runtime_schema(monkeypatch):
    monkeypatch.setattr(compose.db_tasks, "SCHEMA_COMPATIBILITY_CHECKS", [("a", "sql"), ("b", "sql")])
    monkeypatch.setattr(compose, "_compose_query_succeeds", lambda sql: True)

    assert compose._compose_schema_bootstrapped() is True


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

    compose.up.body(None, build=False, wait_timeout=180)

    assert commands == [
        compose._compose_command("up", "-d", "postgres", "nats"),
        compose._compose_command("up", "-d", "core", "interface"),
    ]
    assert waits == [
        ("PostgreSQL", 5432),
        ("PostgreSQL ready", 180),
        ("NATS", 4222),
        ("migrate", 0),
        ("Core API", compose.API_PORT),
        ("Core health", f"http://{compose.API_HOST}:{compose.API_PORT}/healthz"),
        ("Frontend", compose.INTERFACE_PORT),
        ("Frontend", f"http://{compose.INTERFACE_HOST}:{compose.INTERFACE_PORT}/"),
        ("status", 0),
    ]


def test_compose_up_rejects_tiny_wait_timeout(monkeypatch):
    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(compose, "_load_compose_env", lambda: {})
    monkeypatch.setattr(compose, "_validate_compose_env", lambda env_values: None)

    with pytest.raises(SystemExit) as excinfo:
        compose.up.body(None, wait_timeout=29)

    assert "at least 30 seconds" in str(excinfo.value)


def test_compose_up_postgres_timeout_has_guidance(monkeypatch):
    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(compose, "_load_compose_env", lambda: {})
    monkeypatch.setattr(compose, "_validate_compose_env", lambda env_values: None)
    monkeypatch.setattr(
        compose,
        "_run_compose",
        lambda args, check=True: type("Result", (), {"returncode": 0})(),
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
    monkeypatch.setattr(
        compose,
        "_run_compose",
        lambda args, check=True: type("Result", (), {"returncode": 0})(),
    )
    monkeypatch.setattr(compose, "_run_compose_migrations", lambda: None)
    monkeypatch.setattr(compose, "_wait_for_port", lambda port, label, timeout_seconds=60: True)
    monkeypatch.setattr(compose, "_wait_for_postgres_ready", lambda timeout_seconds=90: True)
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

    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(compose, "_run_compose", lambda args, check=True: commands.append(args) or type("Result", (), {"returncode": 0})())

    compose.down.body(None, volumes=True)

    assert commands == [compose._compose_command("down", "--volumes")]


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
