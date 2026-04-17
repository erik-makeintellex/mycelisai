from __future__ import annotations

from pathlib import Path

import pytest

from ops import compose
from ops.config import docker_host_path


def test_compose_command_includes_env_file_and_project_name():
    cmd = compose._compose_command("ps")

    assert "compose" in cmd
    assert "--project-name" in cmd
    assert compose.COMPOSE_PROJECT in cmd
    assert "--env-file" in cmd
    assert docker_host_path(compose.COMPOSE_ENV_FILE) in cmd
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


def test_compose_effective_env_prefers_runtime_override(monkeypatch):
    monkeypatch.setenv("MYCELIS_COMPOSE_OLLAMA_HOST", "http://192.168.50.156:11434")

    values = compose._compose_effective_env(
        {"MYCELIS_COMPOSE_OLLAMA_HOST": "http://host.docker.internal:11434"}
    )

    assert values["MYCELIS_COMPOSE_OLLAMA_HOST"] == "http://192.168.50.156:11434"


def test_compose_runtime_env_passes_overrides_into_wsl(tmp_path, monkeypatch):
    monkeypatch.setattr(compose, "docker_host_mode", lambda: "wsl")
    monkeypatch.setattr(compose, "running_in_wsl", lambda: False)
    monkeypatch.setenv("MYCELIS_COMPOSE_OLLAMA_HOST", "http://192.168.50.156:11434")

    env = compose._compose_runtime_env(
        {
            "MYCELIS_COMPOSE_OLLAMA_HOST": "http://host.docker.internal:11434",
            "MYCELIS_OUTPUT_HOST_PATH": str(tmp_path),
            "MYCELIS_API_KEY": "test-key",
        }
    )

    assert env is not None
    assert env["MYCELIS_COMPOSE_OLLAMA_HOST"] == "http://192.168.50.156:11434"
    assert env["MYCELIS_OUTPUT_HOST_PATH"] == docker_host_path(tmp_path)
    assert "MYCELIS_COMPOSE_OLLAMA_HOST" in env["WSLENV"]
    assert "MYCELIS_OUTPUT_HOST_PATH" in env["WSLENV"]


def test_compose_runtime_env_passes_overrides_inside_direct_wsl_shell(tmp_path, monkeypatch):
    monkeypatch.setattr(compose, "docker_host_mode", lambda: "native")
    monkeypatch.setattr(compose, "running_in_wsl", lambda: True)
    monkeypatch.setenv("MYCELIS_COMPOSE_OLLAMA_HOST", "http://127.0.0.1:11435")

    env = compose._compose_runtime_env(
        {
            "MYCELIS_COMPOSE_OLLAMA_HOST": "http://host.docker.internal:11434",
            "MYCELIS_OUTPUT_HOST_PATH": str(tmp_path),
            "MYCELIS_API_KEY": "test-key",
        }
    )

    assert env is not None
    assert env["MYCELIS_COMPOSE_OLLAMA_HOST"] == "http://127.0.0.1:11435"
    assert env["MYCELIS_OUTPUT_HOST_PATH"] == str(tmp_path)
    assert "WSLENV" not in env or "MYCELIS_COMPOSE_OLLAMA_HOST" not in env.get("WSLENV", "")


def test_prepare_wsl_ollama_host_uses_reachable_configured_target(monkeypatch):
    calls: list[tuple[str, int, int]] = []

    monkeypatch.setattr(compose, "docker_host_mode", lambda: "wsl")
    monkeypatch.setattr(compose, "_inspect_wsl_ollama_relay_labels", lambda: None)
    monkeypatch.setattr(compose, "_wsl_http_available", lambda url: url == "http://192.168.50.156:11434/api/tags" or url == "http://192.168.50.156:11434")
    monkeypatch.setattr(
        compose,
        "_ensure_wsl_ollama_relay",
        lambda host, target_port, relay_port: calls.append((host, target_port, relay_port)),
    )

    values = compose._prepare_wsl_ollama_host(
        {"MYCELIS_COMPOSE_OLLAMA_HOST": "http://192.168.50.156:11434"}
    )

    assert values["MYCELIS_COMPOSE_OLLAMA_HOST"] == "http://host.docker.internal:11435"
    assert calls == [("192.168.50.156", 11434, 11435)]


def test_prepare_wsl_ollama_host_falls_back_to_wsl_localhost(monkeypatch):
    calls: list[tuple[str, int, int]] = []

    monkeypatch.setattr(compose, "docker_host_mode", lambda: "wsl")
    monkeypatch.setattr(compose, "_inspect_wsl_ollama_relay_labels", lambda: None)
    monkeypatch.setattr(
        compose,
        "_wsl_http_available",
        lambda url: url == "http://127.0.0.1:11434/api/tags" or url == "http://127.0.0.1:11434",
    )
    monkeypatch.setattr(
        compose,
        "_ensure_wsl_ollama_relay",
        lambda host, target_port, relay_port: calls.append((host, target_port, relay_port)),
    )

    values = compose._prepare_wsl_ollama_host(
        {"MYCELIS_COMPOSE_OLLAMA_HOST": "http://192.168.50.156:11434"}
    )

    assert values["MYCELIS_COMPOSE_OLLAMA_HOST"] == "http://host.docker.internal:11435"
    assert calls == [("127.0.0.1", 11434, 11435)]


def test_prepare_wsl_ollama_host_fails_when_target_and_localhost_are_unreachable(monkeypatch):
    monkeypatch.setattr(compose, "docker_host_mode", lambda: "wsl")
    monkeypatch.setattr(compose, "_inspect_wsl_ollama_relay_labels", lambda: None)
    monkeypatch.setattr(compose, "_wsl_http_available", lambda url: False)

    with pytest.raises(SystemExit) as excinfo:
        compose._prepare_wsl_ollama_host(
            {"MYCELIS_COMPOSE_OLLAMA_HOST": "http://192.168.50.156:11434"}
        )

    assert "could not reach the configured MYCELIS_COMPOSE_OLLAMA_HOST" in str(excinfo.value)


def test_prepare_wsl_ollama_host_reuses_existing_matching_relay(monkeypatch):
    monkeypatch.setattr(compose, "docker_host_mode", lambda: "wsl")
    monkeypatch.setattr(
        compose,
        "_inspect_wsl_ollama_relay_labels",
        lambda: {
            "mycelis.relay.listen_port": "11435",
            "mycelis.relay.target_host": "127.0.0.1",
            "mycelis.relay.target_port": "11434",
        },
    )
    monkeypatch.setattr(compose, "_wsl_http_available", lambda url: pytest.fail("should not probe WSL when relay already matches"))
    monkeypatch.setattr(compose, "_ensure_wsl_ollama_relay", lambda *args: pytest.fail("should not recreate relay"))

    values = compose._prepare_wsl_ollama_host(
        {"MYCELIS_COMPOSE_OLLAMA_HOST": "http://192.168.50.156:11434"}
    )

    assert values["MYCELIS_COMPOSE_OLLAMA_HOST"] == "http://host.docker.internal:11435"


def test_prepare_wsl_ollama_host_runs_inside_direct_wsl_shell(monkeypatch):
    calls: list[tuple[str, int, int]] = []

    monkeypatch.setattr(compose, "docker_host_mode", lambda: "native")
    monkeypatch.setattr(compose, "running_in_wsl", lambda: True)
    monkeypatch.setattr(compose, "_inspect_wsl_ollama_relay_labels", lambda: None)
    monkeypatch.setattr(
        compose,
        "_wsl_http_available",
        lambda url: url == "http://127.0.0.1:11434/api/tags" or url == "http://127.0.0.1:11434",
    )
    monkeypatch.setattr(
        compose,
        "_ensure_wsl_ollama_relay",
        lambda host, target_port, relay_port: calls.append((host, target_port, relay_port)),
    )

    values = compose._prepare_wsl_ollama_host(
        {"MYCELIS_COMPOSE_OLLAMA_HOST": "http://host.docker.internal:11434"}
    )

    assert values["MYCELIS_COMPOSE_OLLAMA_HOST"] == "http://host.docker.internal:11435"
    assert calls == [("127.0.0.1", 11434, 11435)]


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
    monkeypatch.setattr(compose, "_load_compose_env", lambda: {"DB_USER": "mycelis", "DB_NAME": "cortex"})
    monkeypatch.setattr(compose, "_compose_schema_bootstrapped", lambda env_values=None: False)
    monkeypatch.setattr(
        compose,
        "_run_compose",
        lambda args, check=True, env=None: commands.append(args) or type("Result", (), {"returncode": 0})(),
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


def test_run_compose_migrations_uses_configured_db_login(monkeypatch):
    commands: list[list[str]] = []
    files = [Path("001_init_memory.sql")]

    monkeypatch.setattr(compose.db_tasks, "_migration_files", lambda: files)
    monkeypatch.setattr(compose, "_load_compose_env", lambda: {"DB_USER": "owner", "DB_NAME": "ownerdb"})
    monkeypatch.setattr(compose, "_compose_schema_bootstrapped", lambda env_values=None: False)
    monkeypatch.setattr(
        compose,
        "_run_compose",
        lambda args, check=True, env=None: commands.append(args) or type("Result", (), {"returncode": 0})(),
    )

    compose._run_compose_migrations()

    assert "-U" in commands[0]
    assert "owner" in commands[0]
    assert "-d" in commands[0]
    assert "ownerdb" in commands[0]


def test_run_compose_migrations_skips_replay_when_schema_is_compatible(monkeypatch, capsys):
    monkeypatch.setattr(compose, "_load_compose_env", lambda: {"DB_USER": "mycelis", "DB_NAME": "cortex"})
    monkeypatch.setattr(compose, "_compose_schema_bootstrapped", lambda env_values=None: True)
    monkeypatch.setattr(compose, "_run_missing_compose_storage_migrations", lambda env_values: False)
    monkeypatch.setattr(
        compose.db_tasks,
        "_migration_files",
        lambda: pytest.fail("migration files should not be replayed for a compatible schema"),
    )

    compose._run_compose_migrations()

    out = capsys.readouterr().out
    assert "skipping forward migration replay" in out
    assert "compose.down --volumes" in out


def test_run_compose_migrations_applies_missing_late_storage_when_base_schema_is_compatible(monkeypatch, capsys):
    files = [Path("038_conversation_templates.up.sql")]
    commands: list[list[str]] = []

    monkeypatch.setattr(compose, "_load_compose_env", lambda: {"DB_USER": "mycelis", "DB_NAME": "cortex"})
    monkeypatch.setattr(compose, "_compose_schema_bootstrapped", lambda env_values=None: True)
    monkeypatch.setattr(
        compose,
        "_compose_storage_check_results",
        lambda env_values: [
            ("pgvector extension", True),
            ("conversation templates", False),
        ],
    )
    monkeypatch.setattr(compose.db_tasks, "_migration_files", lambda: files)
    monkeypatch.setattr(
        compose,
        "_run_compose",
        lambda args, check=True, env=None: commands.append(args) or type("Result", (), {"returncode": 0})(),
    )

    compose._run_compose_migrations()

    out = capsys.readouterr().out
    assert "Applying missing long-term storage migrations" in out
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
            "/migrations/038_conversation_templates.up.sql",
        )
    ]


def test_compose_schema_bootstrapped_requires_all_runtime_objects(monkeypatch):
    monkeypatch.setattr(compose.db_tasks, "SCHEMA_COMPATIBILITY_CHECKS", [("a", "sql"), ("b", "sql"), ("c", "sql")])
    monkeypatch.setattr(compose, "_compose_check_results", lambda checks, env_values: [("a", True), ("b", True), ("c", False)])

    assert compose._compose_schema_bootstrapped() is False


def test_compose_schema_bootstrapped_accepts_current_runtime_schema(monkeypatch):
    monkeypatch.setattr(compose.db_tasks, "SCHEMA_COMPATIBILITY_CHECKS", [("a", "sql"), ("b", "sql")])
    monkeypatch.setattr(compose, "_compose_check_results", lambda checks, env_values: [("a", True), ("b", True)])

    assert compose._compose_schema_bootstrapped() is True


def test_compose_up_orders_infra_then_migrations_then_app(monkeypatch):
    commands: list[list[str]] = []
    waits: list[tuple[str, int | str]] = []

    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(
        compose,
        "_run_compose",
        lambda args, check=True, env=None: commands.append(args) or type("Result", (), {"returncode": 0})(),
    )
    monkeypatch.setattr(compose, "_run_compose_migrations", lambda: waits.append(("migrate", 0)))
    monkeypatch.setattr(compose, "_wait_for_port", lambda port, label, timeout_seconds=60: waits.append((label, port)) or True)
    monkeypatch.setattr(
        compose,
        "_wait_for_postgres_ready",
        lambda timeout_seconds=90, env_values=None: waits.append(("PostgreSQL ready", timeout_seconds)) or True,
    )
    monkeypatch.setattr(
        compose,
        "_wait_for_http_ok",
        lambda url, label, timeout_seconds=60, headers=None: waits.append((label, url)) or True,
    )
    monkeypatch.setattr(compose, "_prepare_wsl_ollama_host", lambda env_values: env_values)
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


def test_compose_infra_up_starts_only_data_services(monkeypatch):
    commands: list[list[str]] = []
    waits: list[tuple[str, int]] = []

    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(
        compose,
        "_load_compose_env",
        lambda: {
            "MYCELIS_COMPOSE_POSTGRES_PORT": "15432",
            "MYCELIS_COMPOSE_NATS_PORT": "14222",
            "MYCELIS_COMPOSE_OLLAMA_HOST": "http://host.docker.internal:11434",
        },
    )
    monkeypatch.setattr(compose, "_validate_compose_env", lambda env_values: None)
    monkeypatch.setattr(
        compose,
        "_run_compose",
        lambda args, check=True, env=None: commands.append(args) or type("Result", (), {"returncode": 0})(),
    )
    monkeypatch.setattr(compose, "_wait_for_port", lambda port, label, timeout_seconds=60: waits.append((label, port)) or True)
    monkeypatch.setattr(
        compose,
        "_wait_for_postgres_ready",
        lambda timeout_seconds=90, env_values=None: waits.append(("PostgreSQL ready", timeout_seconds)) or True,
    )
    monkeypatch.setattr(compose, "_run_compose_migrations", lambda: pytest.fail("migrations should be opt-in for infra-up"))

    compose.infra_up.body(None, wait_timeout=120)

    assert commands == [compose._compose_command("up", "-d", "postgres", "nats")]
    assert waits == [
        ("PostgreSQL", 15432),
        ("PostgreSQL ready", 120),
        ("NATS", 14222),
    ]


def test_compose_infra_up_can_run_migrations_when_requested(monkeypatch):
    commands: list[list[str]] = []
    migrated: list[bool] = []

    monkeypatch.setattr(compose, "_require_compose_env_file", lambda: None)
    monkeypatch.setattr(compose, "_load_compose_env", lambda: {})
    monkeypatch.setattr(compose, "_validate_compose_env", lambda env_values: None)
    monkeypatch.setattr(
        compose,
        "_run_compose",
        lambda args, check=True, env=None: commands.append(args) or type("Result", (), {"returncode": 0})(),
    )
    monkeypatch.setattr(compose, "_wait_for_port", lambda port, label, timeout_seconds=60: True)
    monkeypatch.setattr(compose, "_wait_for_postgres_ready", lambda timeout_seconds=90, env_values=None: True)
    monkeypatch.setattr(compose, "_run_compose_migrations", lambda: migrated.append(True))

    compose.infra_up.body(None, wait_timeout=120, migrate=True)

    assert commands == [compose._compose_command("up", "-d", "postgres", "nats")]
    assert migrated == [True]


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
    assert "DB_PASSWORD=<from .env.compose; not printed>" in out
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


def test_validate_compose_env_rejects_loopback_ollama_host():
    with pytest.raises(SystemExit) as excinfo:
        compose._validate_compose_env({"MYCELIS_COMPOSE_OLLAMA_HOST": "http://127.0.0.1:11434"})

    assert "MYCELIS_COMPOSE_OLLAMA_HOST" in str(excinfo.value)
    assert "host.docker.internal" in str(excinfo.value)


def test_validate_compose_env_allows_host_gateway_alias(tmp_path):
    compose._validate_compose_env(
        {
            "MYCELIS_COMPOSE_OLLAMA_HOST": "http://host.docker.internal:11434",
            "MYCELIS_OUTPUT_BLOCK_MODE": "local_hosted",
            "MYCELIS_OUTPUT_HOST_PATH": str(tmp_path),
        }
    )


def test_validate_compose_env_allows_explicit_windows_lan_ollama_host(tmp_path):
    compose._validate_compose_env(
        {
            "MYCELIS_COMPOSE_OLLAMA_HOST": "http://192.168.50.156:11434",
            "MYCELIS_OUTPUT_BLOCK_MODE": "local_hosted",
            "MYCELIS_OUTPUT_HOST_PATH": str(tmp_path),
        }
    )


def test_validate_output_block_requires_local_hosted_path():
    with pytest.raises(SystemExit) as excinfo:
        compose._validate_output_block_config({"MYCELIS_OUTPUT_BLOCK_MODE": "local_hosted"})

    assert "MYCELIS_OUTPUT_HOST_PATH" in str(excinfo.value)


def test_validate_output_block_rejects_explicit_missing_local_hosted_path(tmp_path):
    missing = tmp_path / "missing-output"

    with pytest.raises(SystemExit) as excinfo:
        compose._validate_output_block_config(
            {
                "MYCELIS_OUTPUT_BLOCK_MODE": "local_hosted",
                "MYCELIS_OUTPUT_HOST_PATH": str(missing),
            }
        )

    assert "does not exist" in str(excinfo.value)


def test_validate_output_block_resolves_cross_platform_directory(tmp_path):
    output_dir = tmp_path / "Output Block With Spaces"
    output_dir.mkdir()

    compose._validate_output_block_config(
        {
            "MYCELIS_OUTPUT_BLOCK_MODE": "local_hosted",
            "MYCELIS_OUTPUT_HOST_PATH": f'"{output_dir}"',
        }
    )


def test_validate_output_block_creates_cluster_generated_default(tmp_path, monkeypatch):
    generated = tmp_path / "cluster-generated"
    monkeypatch.setattr(compose, "DEFAULT_OUTPUT_HOST_PATH", generated)

    compose._validate_output_block_config({"MYCELIS_OUTPUT_BLOCK_MODE": "cluster_generated"})

    assert generated.is_dir()


def test_validate_compose_env_creates_implicit_default_output_path(tmp_path, monkeypatch):
    generated = tmp_path / "implicit-output"
    monkeypatch.setattr(compose, "DEFAULT_OUTPUT_HOST_PATH", generated)

    compose._validate_compose_env({"MYCELIS_COMPOSE_OLLAMA_HOST": "http://host.docker.internal:11434"})

    assert generated.is_dir()


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
