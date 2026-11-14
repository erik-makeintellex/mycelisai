from __future__ import annotations

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
    (tmp_path / ".env").write_text("MYCELIS_API_KEY=secret-key\nDB_PASSWORD=secret-password\nDB_HOST=local-host\n", encoding="utf-8")
    env_file.write_text("# comment\nMYCELIS_API_KEY=compose-ignored\nDB_PASSWORD=compose-ignored\nDB_HOST=postgres\n", encoding="utf-8")
    monkeypatch.setattr(compose, "COMPOSE_ENV_FILE", env_file)

    assert compose._load_compose_env() == {
        "MYCELIS_API_KEY": "secret-key",
        "DB_PASSWORD": "secret-password",
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
    (tmp_path / ".env").write_text("MYCELIS_API_KEY=x\n", encoding="utf-8")
    example.write_text("MYCELIS_COMPOSE_OLLAMA_HOST=http://host.docker.internal:11434\n", encoding="utf-8")
    monkeypatch.setattr(compose, "COMPOSE_ENV_FILE", missing)
    monkeypatch.setattr(compose, "COMPOSE_ENV_EXAMPLE", example)

    with pytest.raises(SystemExit) as excinfo:
        compose._require_compose_env_file()

    assert ".env.compose.example" in str(excinfo.value)
    assert "keep secrets in .env" in str(excinfo.value)


def test_validate_compose_env_rejects_loopback_ollama_host():
    with pytest.raises(SystemExit) as excinfo:
        compose._validate_compose_env({"MYCELIS_COMPOSE_OLLAMA_HOST": "http://127.0.0.1:11434"})

    assert "MYCELIS_COMPOSE_OLLAMA_HOST" in str(excinfo.value)
    assert "host.docker.internal" in str(excinfo.value)


def test_validate_compose_env_allows_host_gateway_alias(tmp_path):
    compose._validate_compose_env({
        "MYCELIS_COMPOSE_OLLAMA_HOST": "http://host.docker.internal:11434",
        "MYCELIS_OUTPUT_BLOCK_MODE": "local_hosted",
        "MYCELIS_OUTPUT_HOST_PATH": str(tmp_path),
    })


def test_validate_compose_env_allows_explicit_windows_lan_ollama_host(tmp_path):
    compose._validate_compose_env({
        "MYCELIS_COMPOSE_OLLAMA_HOST": "http://192.168.50.156:11434",
        "MYCELIS_OUTPUT_BLOCK_MODE": "local_hosted",
        "MYCELIS_OUTPUT_HOST_PATH": str(tmp_path),
    })


def test_validate_output_block_requires_local_hosted_path():
    with pytest.raises(SystemExit) as excinfo:
        compose._validate_output_block_config({"MYCELIS_OUTPUT_BLOCK_MODE": "local_hosted"})
    assert "MYCELIS_OUTPUT_HOST_PATH" in str(excinfo.value)


def test_validate_output_block_rejects_explicit_missing_local_hosted_path(tmp_path):
    missing = tmp_path / "missing-output"
    with pytest.raises(SystemExit) as excinfo:
        compose._validate_output_block_config({
            "MYCELIS_OUTPUT_BLOCK_MODE": "local_hosted",
            "MYCELIS_OUTPUT_HOST_PATH": str(missing),
        })
    assert "does not exist" in str(excinfo.value)


def test_validate_output_block_resolves_cross_platform_directory(tmp_path):
    output_dir = tmp_path / "Output Block With Spaces"
    output_dir.mkdir()
    compose._validate_output_block_config({
        "MYCELIS_OUTPUT_BLOCK_MODE": "local_hosted",
        "MYCELIS_OUTPUT_HOST_PATH": f'"{output_dir}"',
    })


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
