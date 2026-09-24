from __future__ import annotations

import subprocess
from dataclasses import replace
from pathlib import Path

import pytest

from ops import lifecycle_first_boot_isolated as isolated


def fixture_at(root: Path, *, project: str = "mycelis-firstboot-0123456789abcdef") -> isolated.Fixture:
    root.mkdir(parents=True, exist_ok=True)
    (root / isolated.OWNER_MARKER).write_text(project, encoding="utf-8")
    (root / "output").mkdir(exist_ok=True)
    ports = {key: 25000 + index for index, key in enumerate(isolated.PORT_TARGETS)}
    environment = {"MYCELIS_OUTPUT_HOST_PATH": str(root / "output"), "MYCELIS_API_KEY": "fixture-key"}
    return isolated.Fixture(project, root, root / "docker-compose.yml", root / "fixture.env", ports, environment)


def test_generated_compose_binds_every_published_port_to_loopback():
    source = isolated.COMPOSE_SOURCE.read_text(encoding="utf-8")
    rendered = isolated._fixture_compose_text(source)

    assert rendered.count('"127.0.0.1:${MYCELIS_COMPOSE_') == len(isolated.PORT_TARGETS)
    assert "build:\n      context: ." in rendered
    assert "postgres-data:/var/lib/postgresql/data" in rendered
    with pytest.raises(SystemExit, match="port mapping"):
        isolated._fixture_compose_text(source.replace("${MYCELIS_COMPOSE_CORE_PORT:-8081}:8080", "8081:8080"))


def test_fixture_rejects_retained_identity_root_and_port(tmp_path):
    fixture = fixture_at(tmp_path / "mycelis-first-boot-case")
    isolated._validate_fixture(fixture)

    with pytest.raises(SystemExit, match="project identity"):
        isolated._validate_fixture(fixture_at(fixture.root, project="mycelis-home"))
    (fixture.root / isolated.OWNER_MARKER).write_text(fixture.project, encoding="utf-8")
    fixture.ports["MYCELIS_COMPOSE_CORE_PORT"] = 8081
    with pytest.raises(SystemExit, match="overlaps a default"):
        isolated._validate_fixture(fixture)


def test_fixture_environment_ignores_retained_compose_overrides(monkeypatch, tmp_path):
    fixture = fixture_at(tmp_path / "mycelis-first-boot-case")
    monkeypatch.setenv("MYCELIS_API_KEY", "retained-key")
    monkeypatch.setenv("DB_HOST", "retained-postgres")
    monkeypatch.setenv("COMPOSE_PROJECT_NAME", "mycelis-home")
    monkeypatch.setenv("DOCKER_HOST", "unix:///tmp/test-docker.sock")

    result = isolated._compose_environment(fixture)

    assert result["MYCELIS_API_KEY"] == "fixture-key"
    assert "DB_HOST" not in result
    assert "COMPOSE_PROJECT_NAME" not in result
    assert result["DOCKER_HOST"] == "unix:///tmp/test-docker.sock"


def test_compose_command_targets_only_named_fixture(monkeypatch, tmp_path):
    fixture = fixture_at(tmp_path / "mycelis-first-boot-case")
    captured = {}

    def fake_run(command, **kwargs):
        captured.update(command=command, environment=kwargs["env"])
        return subprocess.CompletedProcess(command, 0, "", "")

    monkeypatch.setattr(isolated.subprocess, "run", fake_run)
    isolated._compose(fixture, "down", "--volumes")

    command = captured["command"]
    assert command[:4] == ["docker", "compose", "--project-name", fixture.project]
    assert command[-2:] == ["down", "--volumes"]
    assert str(fixture.compose_file) in command
    assert "mycelis-home" not in command


def test_fixture_cleanup_after_proof_failure(monkeypatch, tmp_path):
    fixture = fixture_at(tmp_path / "mycelis-first-boot-case")
    calls = []
    monkeypatch.setattr(isolated, "_make_fixture", lambda: fixture)

    def fake_compose(_fixture, *args, **_kwargs):
        calls.append(args)
        return subprocess.CompletedProcess(args, 0, "", "")

    monkeypatch.setattr(isolated, "_compose", fake_compose)
    monkeypatch.setattr(isolated, "_repair_fixture_output_permissions", lambda _fixture: calls.append(("repair",)))
    monkeypatch.setattr(isolated, "_run_proof", lambda *_args: (_ for _ in ()).throw(RuntimeError("probe failed")))

    with pytest.raises(RuntimeError, match="probe failed"):
        isolated.run_isolated_first_boot(build=False, user_tables=("groups",), bootstrap_tables=("nodes",))

    assert calls == [
        ("config", "--quiet"), ("ps", "-q"), ("stop", "core", "interface"),
        ("repair",), ("down", "--volumes", "--remove-orphans", "--rmi", "local"),
    ]
    assert not fixture.root.exists()


def test_failed_fixture_cleanup_fails_proof_and_preserves_recovery_files(monkeypatch, tmp_path):
    fixture = fixture_at(tmp_path / "mycelis-first-boot-case")
    monkeypatch.setattr(isolated, "_make_fixture", lambda: fixture)
    monkeypatch.setattr(isolated, "_run_proof", lambda *_args: None)

    def fake_compose(_fixture, *args, **_kwargs):
        code = 1 if args[0] == "down" else 0
        return subprocess.CompletedProcess(args, code, "", "")

    monkeypatch.setattr(isolated, "_compose", fake_compose)
    monkeypatch.setattr(isolated, "_repair_fixture_output_permissions", lambda _fixture: None)
    with pytest.raises(SystemExit, match="Fixture cleanup failed"):
        isolated.run_isolated_first_boot(build=False, user_tables=("groups",), bootstrap_tables=("nodes",))

    assert fixture.root.exists()


def test_failed_permission_repair_retains_fixture_image_and_files(monkeypatch, tmp_path):
    fixture = fixture_at(tmp_path / "mycelis-first-boot-case")
    calls = []
    monkeypatch.setattr(isolated, "_make_fixture", lambda: fixture)
    monkeypatch.setattr(isolated, "_run_proof", lambda *_args: None)

    def fake_compose(_fixture, *args, **_kwargs):
        calls.append(args)
        return subprocess.CompletedProcess(args, 0, "", "")

    monkeypatch.setattr(isolated, "_compose", fake_compose)
    monkeypatch.setattr(
        isolated, "_repair_fixture_output_permissions",
        lambda _fixture: (_ for _ in ()).throw(SystemExit("repair failed")),
    )

    with pytest.raises(SystemExit, match="Fixture cleanup failed.*repair failed"):
        isolated.run_isolated_first_boot(build=False, user_tables=("groups",), bootstrap_tables=("nodes",))

    assert calls[-1] == ("down", "--volumes", "--remove-orphans")
    assert fixture.root.exists()


def test_output_repair_uses_only_fixture_image_and_networkless_mount(monkeypatch, tmp_path):
    fixture = fixture_at(tmp_path / "mycelis-first-boot-case")
    image = "sha256:" + "a" * 64
    monkeypatch.setattr(isolated, "_compose", lambda *_args, **_kwargs: subprocess.CompletedProcess([], 0, image + "\n", ""))
    commands = []

    def fake_run(command, **_kwargs):
        commands.append(command)
        return subprocess.CompletedProcess(command, 0, "", "")

    monkeypatch.setattr(isolated.subprocess, "run", fake_run)
    isolated._repair_fixture_output_permissions(fixture)

    assert len(commands) == 1
    command = commands[0]
    assert command[:2] == ["docker", "run"]
    assert "--network=none" in command
    assert "--pull=never" in command
    assert image in command
    assert f"type=bind,source={fixture.root / 'output'},target=/data" in command
    assert command[-3:] == ["-R", "a+rwX", "/data"]


def test_output_repair_rejects_success_without_host_writability(monkeypatch, tmp_path):
    fixture = fixture_at(tmp_path / "mycelis-first-boot-case")
    image = "sha256:" + "a" * 64
    monkeypatch.setattr(isolated, "_compose", lambda *_args, **_kwargs: subprocess.CompletedProcess([], 0, image, ""))
    monkeypatch.setattr(
        isolated.subprocess, "run",
        lambda command, **_kwargs: subprocess.CompletedProcess(command, 0, "", ""),
    )
    monkeypatch.setattr(isolated.os, "access", lambda *_args: False)

    with pytest.raises(SystemExit, match="remains unwritable after repair"):
        isolated._repair_fixture_output_permissions(fixture)


def test_output_repair_refuses_symlinked_output_before_docker(monkeypatch, tmp_path):
    fixture = fixture_at(tmp_path / "mycelis-first-boot-case")
    (fixture.root / "output").rmdir()
    (fixture.root / "output").symlink_to(tmp_path, target_is_directory=True)
    monkeypatch.setattr(isolated, "_compose", lambda *_args, **_kwargs: pytest.fail("Compose called for escaped output"))

    with pytest.raises(SystemExit, match="escaped its owned root"):
        isolated._repair_fixture_output_permissions(fixture)


def test_no_cleanup_of_unowned_project_on_preflight_failure(monkeypatch, tmp_path):
    fixture = fixture_at(tmp_path / "mycelis-first-boot-case", project="mycelis-home")
    monkeypatch.setattr(isolated, "_make_fixture", lambda: fixture)
    monkeypatch.setattr(isolated, "_compose", lambda *_args, **_kwargs: pytest.fail("Compose called before guard"))

    with pytest.raises(SystemExit, match="project identity"):
        isolated.run_isolated_first_boot(build=False, user_tables=("groups",), bootstrap_tables=("nodes",))

    assert not fixture.root.exists()


def test_preflight_never_deletes_repository_root(monkeypatch, tmp_path):
    fixture = fixture_at(tmp_path / "mycelis-first-boot-case")
    fixture = replace(fixture, root=isolated.ROOT_DIR)
    monkeypatch.setattr(isolated, "_make_fixture", lambda: fixture)
    monkeypatch.setattr(isolated, "_compose", lambda *_args, **_kwargs: pytest.fail("Compose called before root guard"))

    with pytest.raises(SystemExit, match="outside the repository"):
        isolated.run_isolated_first_boot(build=False, user_tables=("groups",), bootstrap_tables=("nodes",))

    assert isolated.ROOT_DIR.exists()


def test_preflight_does_not_delete_unowned_lookalike_directory(monkeypatch, tmp_path):
    fixture = fixture_at(tmp_path / "mycelis-first-boot-case")
    (fixture.root / isolated.OWNER_MARKER).write_text("someone-else", encoding="utf-8")
    monkeypatch.setattr(isolated, "_make_fixture", lambda: fixture)

    with pytest.raises(SystemExit, match="outside the repository"):
        isolated.run_isolated_first_boot(build=False, user_tables=("groups",), bootstrap_tables=("nodes",))

    assert fixture.root.exists()
