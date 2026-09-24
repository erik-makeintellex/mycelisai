from __future__ import annotations

import subprocess
from pathlib import Path

import pytest

from ops import lifecycle_first_boot_framework as framework
from ops import lifecycle_first_boot_isolated as isolated


def fixture_at(root: Path, *, project: str = "mycelis-firstboot-0123456789abcdef") -> isolated.Fixture:
    root.mkdir(parents=True, exist_ok=True)
    (root / isolated.OWNER_MARKER).write_text(project, encoding="utf-8")
    (root / "output").mkdir()
    ports = {key: 25000 + index for index, key in enumerate(isolated.PORT_TARGETS)}
    environment = {"MYCELIS_OUTPUT_HOST_PATH": str(root / "output")}
    return isolated.Fixture(project, root, root / "docker-compose.yml", root / "fixture.env", ports, environment)


def test_framework_fixture_uses_only_owned_overlay_secrets_and_tls(tmp_path, monkeypatch):
    fixture = fixture_at(tmp_path / "mycelis-first-boot-b2")
    fixture.compose_file.write_text("services: {}\n", encoding="utf-8")
    monkeypatch.setattr(framework.os, "getuid", lambda: 1000)

    framework.attach_framework_fixture(fixture)

    overlay = fixture.root / framework.OVERLAY_NAME
    assert overlay.read_text(encoding="utf-8") == isolated.FRAMEWORK_OVERLAY_SOURCE.read_text(encoding="utf-8")
    assert fixture.environment["FRAMEWORK_RUNS_CORE_TOKEN"]
    assert fixture.environment["FRAMEWORK_RUNS_DB_PASSWORD"]
    assert "@framework-runs-postgres:5432/framework_runs?sslmode=disable" in fixture.environment["FRAMEWORK_RUNS_DATABASE_URL"]
    assert "MYCELIS_COMPOSE_FRAMEWORK_RUNS" not in fixture.environment
    for key in ("FRAMEWORK_RUNS_TLS_CERT_HOST_FILE", "FRAMEWORK_RUNS_TLS_KEY_HOST_FILE", "FRAMEWORK_RUNS_TLS_CA_HOST_FILE"):
        path = Path(fixture.environment[key])
        assert path.is_file() and fixture.root in path.parents
        assert path.stat().st_mode & 0o777 == 0o600
    assert fixture.env_file.stat().st_mode & 0o777 == 0o600
    result = subprocess.run(["openssl", "verify", "-CAfile", fixture.environment["FRAMEWORK_RUNS_TLS_CA_HOST_FILE"],
        fixture.environment["FRAMEWORK_RUNS_TLS_CERT_HOST_FILE"]], capture_output=True, text=True)
    assert result.returncode == 0
    certificate = subprocess.run(["openssl", "x509", "-in", fixture.environment["FRAMEWORK_RUNS_TLS_CERT_HOST_FILE"],
        "-noout", "-ext", "subjectAltName"], capture_output=True, text=True)
    assert certificate.returncode == 0 and "DNS:framework-runs-control" in certificate.stdout

    monkeypatch.setenv("FRAMEWORK_RUNS_CORE_TOKEN", "retained-token")
    monkeypatch.setenv("FRAMEWORK_RUNS_DB_PASSWORD", "retained-password")
    effective = isolated._compose_environment(fixture)
    assert effective["FRAMEWORK_RUNS_CORE_TOKEN"] == fixture.environment["FRAMEWORK_RUNS_CORE_TOKEN"]
    assert effective["FRAMEWORK_RUNS_DB_PASSWORD"] == fixture.environment["FRAMEWORK_RUNS_DB_PASSWORD"]


def test_framework_fixture_rejects_unowned_project_before_openssl(tmp_path, monkeypatch):
    fixture = fixture_at(tmp_path / "mycelis-first-boot-b2", project="mycelis-home")
    monkeypatch.setattr(framework, "_create_fixture_certificate", lambda _root: pytest.fail("openssl ran for unowned root"))
    with pytest.raises(SystemExit, match="project identity"):
        framework.attach_framework_fixture(fixture)


def test_framework_fixture_requires_host_uid_matching_container(tmp_path, monkeypatch):
    fixture = fixture_at(tmp_path / "mycelis-first-boot-b2")
    monkeypatch.setattr(framework.os, "getuid", lambda: 1001)
    monkeypatch.setattr(framework, "_create_fixture_certificate", lambda _root: pytest.fail("openssl ran for unreadable key"))
    with pytest.raises(SystemExit, match="UID 1000"):
        framework.attach_framework_fixture(fixture)


def test_opt_in_framework_proof_stays_inside_disposable_runner(tmp_path, monkeypatch):
    fixture = fixture_at(tmp_path / "mycelis-first-boot-b2")
    calls = []
    monkeypatch.setattr(isolated, "_make_fixture", lambda: fixture)
    monkeypatch.setattr(framework, "attach_framework_fixture", lambda _fixture: calls.append("attach"))
    monkeypatch.setattr(isolated, "_compose", lambda _fixture, *args, **_kwargs:
        subprocess.CompletedProcess(args, 0, "", ""))
    monkeypatch.setattr(isolated, "_run_proof", lambda *_args: calls.append("base"))
    monkeypatch.setattr(framework, "prove_framework_fixture", lambda _fixture, **_kwargs: calls.append("framework"))
    monkeypatch.setattr(isolated, "_repair_fixture_output_permissions", lambda _fixture: calls.append("repair"))

    isolated.run_isolated_first_boot(build=True, user_tables=("groups",), bootstrap_tables=("nodes",), framework_runs=True)

    assert calls == ["attach", "base", "framework", "repair"]
    assert not fixture.root.exists()


def test_no_framework_overlay_in_default_fixture(monkeypatch, tmp_path):
    fixture = fixture_at(tmp_path / "mycelis-first-boot-b2")
    monkeypatch.setattr(isolated, "_make_fixture", lambda: fixture)
    monkeypatch.setattr(framework, "attach_framework_fixture", lambda _fixture: pytest.fail("opt-in attached by default"))
    monkeypatch.setattr(isolated, "_compose", lambda _fixture, *args, **_kwargs:
        subprocess.CompletedProcess(args, 0, "", ""))
    monkeypatch.setattr(isolated, "_run_proof", lambda *_args: None)
    monkeypatch.setattr(isolated, "_repair_fixture_output_permissions", lambda _fixture: None)

    isolated.run_isolated_first_boot(build=False, user_tables=("groups",), bootstrap_tables=("nodes",))

    assert not fixture.root.exists()
