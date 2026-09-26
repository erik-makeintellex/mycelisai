"""Disposable full-Compose proof for the canonical first-boot task."""

from __future__ import annotations

import json
import os
import re
import secrets
import shutil
import socket
import subprocess
import tempfile
import time
import urllib.error
import urllib.request
from dataclasses import dataclass
from pathlib import Path

from .config import ROOT_DIR, docker_command, docker_host_path
from .lifecycle_first_boot_readiness import (
    redact,
    retry_transient,
    secret_values,
    wait_for_isolated_postgres_ready,
)
from .lifecycle_first_boot_state import assert_empty_user_state, assert_nats_empty, assert_schema_compatible, count_tables


COMPOSE_SOURCE = ROOT_DIR / "docker-compose.yml"
FRAMEWORK_OVERLAY_SOURCE = ROOT_DIR / "deploy" / "compose" / "framework-runs.yml"
PORT_TARGETS = {
    "MYCELIS_COMPOSE_POSTGRES_PORT": "5432",
    "MYCELIS_COMPOSE_NATS_PORT": "4222",
    "MYCELIS_COMPOSE_NATS_MONITOR_PORT": "8222",
    "MYCELIS_COMPOSE_SEARXNG_PORT": "8080",
    "MYCELIS_COMPOSE_CORE_PORT": "8080",
    "MYCELIS_COMPOSE_INTERFACE_PORT": "3000",
}
DEFAULT_PORTS = {5432, 15432, 4222, 8222, 8088, 8081, 3000}
PROJECT_PATTERN = re.compile(r"mycelis-firstboot-[0-9a-f]{16}\Z")
COMPOSE_VARIABLE = re.compile(r"\$\{([A-Za-z_][A-Za-z_0-9]*)")
OWNER_MARKER = ".mycelis-first-boot-owner"


@dataclass(frozen=True)
class Fixture:
    project: str
    root: Path
    compose_file: Path
    env_file: Path
    ports: dict[str, int]
    environment: dict[str, str]


def _available_port(excluded: set[int]) -> int:
    for _ in range(32):
        with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as listener:
            listener.bind(("127.0.0.1", 0))
            port = listener.getsockname()[1]
        if port not in excluded:
            return port
    raise SystemExit("Could not reserve distinct fixture host ports.")


def _fixture_environment(ports: dict[str, int], root: Path) -> dict[str, str]:
    password = secrets.token_urlsafe(32)
    api_key = secrets.token_urlsafe(48)
    return {
        **{key: str(value) for key, value in ports.items()},
        "POSTGRES_USER": "mycelis",
        "POSTGRES_DB": "cortex",
        "POSTGRES_PASSWORD": password,
        "DB_USER": "mycelis",
        "DB_NAME": "cortex",
        "DB_HOST": "postgres",
        "DB_PORT": "5432",
        "DB_PASSWORD": password,
        "NATS_URL": "nats://nats:4222",
        "MYCELIS_API_KEY": api_key,
        "MYCELIS_WEB_SESSION_SECRET": secrets.token_urlsafe(48),
        "MYCELIS_WEB_IDENTITY_FORWARD_SECRET": secrets.token_urlsafe(48),
        "MYCELIS_NATS_SERVICE_ID": "firstboot-" + root.name.removeprefix("mycelis-first-boot-"),
        "MYCELIS_OUTPUT_HOST_PATH": docker_host_path(root / "output"),
        "MYCELIS_WORKSPACE": "/data/workspace",
        "MYCELIS_ARTIFACT_ROOT": "/data/artifacts",
        "DATA_DIR": "/data/artifacts",
        "MYCELIS_COMPOSE_OLLAMA_ENABLED": "false",
        "MYCELIS_DISABLE_DEFAULT_MCP_BOOTSTRAP": "true",
        "CORS_ORIGIN": f"http://127.0.0.1:{ports['MYCELIS_COMPOSE_INTERFACE_PORT']}",
    }


def _fixture_compose_text(source: str) -> str:
    """Keep canonical services/builds, binding only published ports to loopback."""
    text = source
    for key, container_port in PORT_TARGETS.items():
        pattern = re.compile(
            rf'(?m)^([ \t]*- ")\$\{{{key}:-[0-9]+\}}:{container_port}("[ \t]*)$'
        )
        text, count = pattern.subn(rf'\g<1>127.0.0.1:${{{key}}}:{container_port}\g<2>', text)
        if count != 1:
            raise SystemExit(f"Canonical Compose port mapping for {key} changed; isolated proof refused.")
    return text


def _make_fixture() -> Fixture:
    root = Path(tempfile.mkdtemp(prefix="mycelis-first-boot-")).resolve()
    try:
        project = "mycelis-firstboot-" + secrets.token_hex(8)
        excluded = set(DEFAULT_PORTS)
        ports: dict[str, int] = {}
        for key in PORT_TARGETS:
            port = _available_port(excluded)
            ports[key] = port
            excluded.add(port)
        values = _fixture_environment(ports, root)
        compose_file = root / "docker-compose.yml"
        env_file = root / "fixture.env"
        compose_file.write_text(_fixture_compose_text(COMPOSE_SOURCE.read_text(encoding="utf-8")), encoding="utf-8")
        env_file.write_text("\n".join(f"{key}={value}" for key, value in sorted(values.items())) + "\n", encoding="utf-8")
        env_file.chmod(0o600)
        marker = root / OWNER_MARKER
        marker.write_text(project, encoding="utf-8")
        marker.chmod(0o600)
        (root / "output").mkdir(mode=0o700)
    except BaseException:
        shutil.rmtree(root)
        raise
    return Fixture(project, root, compose_file, env_file, ports, values)


def _validate_fixture(fixture: Fixture) -> None:
    root = fixture.root.resolve()
    if not PROJECT_PATTERN.fullmatch(fixture.project) or fixture.project == "mycelis-home":
        raise SystemExit("Isolated first-boot project identity is invalid.")
    if not _owned_fixture_root(root, fixture.project):
        raise SystemExit("Isolated first-boot root must be outside the repository.")
    if set(fixture.ports) != set(PORT_TARGETS) or len(set(fixture.ports.values())) != len(PORT_TARGETS):
        raise SystemExit("Isolated first-boot ports are missing or duplicated.")
    if any(port in DEFAULT_PORTS or not 1024 < port < 65536 for port in fixture.ports.values()):
        raise SystemExit("Isolated first-boot port overlaps a default target.")
    if fixture.environment.get("MYCELIS_OUTPUT_HOST_PATH") != docker_host_path(root / "output"):
        raise SystemExit("Isolated first-boot output path is outside its fixture.")
    if fixture.compose_file.resolve() != root / "docker-compose.yml" or fixture.env_file.resolve() != root / "fixture.env":
        raise SystemExit("Isolated first-boot files are outside their fixture.")


def _owned_fixture_root(root: Path, project: str) -> bool:
    root = root.resolve()
    if not (root.name.startswith("mycelis-first-boot-") and root != ROOT_DIR
            and root not in ROOT_DIR.parents and ROOT_DIR not in root.parents):
        return False
    try:
        return (root / OWNER_MARKER).read_text(encoding="utf-8") == project
    except OSError:
        return False


def _compose_environment(fixture: Fixture) -> dict[str, str]:
    # Shell values normally override --env-file. Strip every interpolation key
    # before setting the fixture's exact values, while preserving Docker access.
    source = COMPOSE_SOURCE.read_text(encoding="utf-8")
    overlay = fixture.root / "framework-runs.yml"
    if overlay.exists():
        source += "\n" + overlay.read_text(encoding="utf-8")
    variables = set(COMPOSE_VARIABLE.findall(source)) | {"COMPOSE_PROJECT_NAME"}
    environment = {key: value for key, value in os.environ.items() if key not in variables}
    environment.update(fixture.environment)
    return environment


def _compose(fixture: Fixture, *args: str, check: bool = True) -> subprocess.CompletedProcess[str]:
    compose_files = ["-f", docker_host_path(fixture.compose_file)]
    overlay = fixture.root / "framework-runs.yml"
    if overlay.exists():
        compose_files.extend(["-f", docker_host_path(overlay)])
    command = docker_command(
        "compose", "--project-name", fixture.project,
        "--project-directory", docker_host_path(ROOT_DIR),
        "--env-file", docker_host_path(fixture.env_file),
        *compose_files, *args, cwd=ROOT_DIR,
    )
    result = subprocess.run(command, cwd=ROOT_DIR, env=_compose_environment(fixture), capture_output=True, text=True)
    if check and result.returncode != 0:
        detail = redact(result.stderr.strip() or result.stdout.strip(), secret_values(fixture.environment))
        raise SystemExit(f"Isolated Compose {args[0]} failed for {fixture.project}: {detail or '(no output captured)'}")
    return result


def _wait_until(label: str, probe, timeout: float = 120) -> None:
    deadline = time.monotonic() + timeout
    while time.monotonic() < deadline:
        try:
            if probe():
                print(f"  [OK] {label}")
                return
        except (OSError, urllib.error.URLError, json.JSONDecodeError):
            pass
        time.sleep(1)
    raise SystemExit(f"Isolated first-boot proof timed out waiting for {label}.")


def _http_ok(url: str, *, api_key: str = "") -> bool:
    headers = {"Authorization": f"Bearer {api_key}"} if api_key else {}
    request = urllib.request.Request(url, headers=headers)
    with urllib.request.urlopen(request, timeout=5) as response:
        response.read(65536)
        return response.status == 200


def _psql_exec(fixture: Fixture, *args: str, check: bool = True) -> subprocess.CompletedProcess[str]:
    return _compose(
        fixture, "exec", "-T", "postgres", "psql", "-X", "-A", "-t",
        "-v", "ON_ERROR_STOP=1", "-U", "mycelis", "-d", "cortex", *args, check=check,
    )


def _psql(fixture: Fixture, *args: str) -> str:
    return _psql_exec(fixture, *args).stdout.strip()


def _assert_http_ready(fixture: Fixture) -> None:
    core_port = fixture.ports["MYCELIS_COMPOSE_CORE_PORT"]
    interface_port = fixture.ports["MYCELIS_COMPOSE_INTERFACE_PORT"]
    core = f"http://127.0.0.1:{core_port}"
    _wait_until("Core /healthz", lambda: _http_ok(core + "/healthz"))
    _wait_until("Core authenticated templates read", lambda: _http_ok(core + "/api/v1/templates", api_key=fixture.environment["MYCELIS_API_KEY"]))
    _wait_until("Interface /", lambda: _http_ok(f"http://127.0.0.1:{interface_port}/"))


def _run_proof(fixture: Fixture, build: bool, user_tables: tuple[str, ...], bootstrap_tables: tuple[str, ...]) -> None:
    _compose(fixture, "up", "-d", "postgres", "nats")
    fixture_secrets = secret_values(fixture.environment)
    # Postgres briefly stops/restarts internally after its own init phase; a lone
    # pg_isready can see that temporary server, so require a real SELECT 1 and
    # retry only transient exec failures on the schema-install exec after it.
    wait_for_isolated_postgres_ready(
        pg_isready=lambda: _compose(fixture, "exec", "-T", "postgres", "pg_isready", "-U", "mycelis", "-d", "cortex", check=False),
        select_one=lambda: _psql_exec(fixture, "-c", "SELECT 1", check=False),
        secrets=fixture_secrets,
    )
    print("  [OK] fixture PostgreSQL")
    _wait_until("fixture NATS monitor", lambda: _http_ok(
        f"http://127.0.0.1:{fixture.ports['MYCELIS_COMPOSE_NATS_MONITOR_PORT']}/healthz"
    ))
    retry_transient(
        "Isolated first-boot schema install",
        lambda: _psql_exec(fixture, "-f", "/migrations/001_current_schema.sql", check=False),
        secrets=fixture_secrets,
    )
    psql = lambda *a: _psql(fixture, *a)  # noqa: E731 - bound runner for the state-assertion module
    nats_port = fixture.ports["MYCELIS_COMPOSE_NATS_MONITOR_PORT"]
    assert_schema_compatible(psql)
    assert_empty_user_state(psql, user_tables, "schema install")
    assert_nats_empty(nats_port)

    if (fixture.root / "framework-runs.yml").exists():
        from .lifecycle_first_boot_framework import prepare_framework_fixture

        prepare_framework_fixture(fixture, build=build)

    if build:
        _compose(fixture, "build", "core", "interface")
    _compose(fixture, "up", "-d", "--no-build", "core", "interface")
    _assert_http_ready(fixture)
    assert_empty_user_state(psql, user_tables, "first boot")
    first_bootstrap = count_tables(psql, bootstrap_tables)
    assert_nats_empty(nats_port)

    _compose(fixture, "restart", "core", "interface")
    _assert_http_ready(fixture)
    assert_empty_user_state(psql, user_tables, "restart")
    if count_tables(psql, bootstrap_tables) != first_bootstrap:
        raise SystemExit("Isolated first-boot bootstrap row counts changed after restart.")
    assert_nats_empty(nats_port)
    print("ISOLATED CLEAN FIRST-BOOT CHECKS PASSED; cleaning fixture...")


def _repair_fixture_output_permissions(fixture: Fixture) -> None:
    _validate_fixture(fixture)
    output = fixture.root / "output"
    if output.is_symlink() or not output.is_dir() or output.resolve() != fixture.root.resolve() / "output":
        raise SystemExit("Fixture output path is missing or escaped its owned root.")
    image_result = _compose(fixture, "images", "-q", "core", check=False)
    if image_result.returncode != 0:
        raise SystemExit("Could not identify the fixture Core image for output cleanup.")
    image = image_result.stdout.strip()
    if not image:
        return  # Core never built; it could not have created output directories.
    if not re.fullmatch(r"(?:sha256:)?[0-9a-f]{64}", image):
        raise SystemExit("Fixture Core image identity is malformed; output cleanup refused.")
    command = docker_command(
        "run", "--rm", "--pull=never", "--network=none", "--user=0:0",
        "--mount", f"type=bind,source={docker_host_path(output)},target=/data",
        "--entrypoint=/bin/chmod", image, "-R", "a+rwX", "/data", cwd=ROOT_DIR,
    )
    result = subprocess.run(command, cwd=ROOT_DIR, env=_compose_environment(fixture), capture_output=True, text=True)
    if result.returncode != 0:
        raise SystemExit("Fixture-owned output permission repair failed; project image was retained for recovery.")
    errors: list[OSError | str] = []
    for directory, _, _ in os.walk(output, onerror=errors.append):
        if not os.access(directory, os.W_OK | os.X_OK):
            errors.append(directory)
    if errors:
        raise SystemExit("Fixture output remains unwritable after repair; project image was retained for recovery.")

def run_isolated_first_boot(*, build: bool, user_tables: tuple[str, ...], bootstrap_tables: tuple[str, ...], framework_runs: bool = False) -> None:
    if framework_runs and not build:
        raise SystemExit("Framework Runs isolated proof requires --build.")
    fixture = _make_fixture()
    armed = False
    cleaned = False
    repair_error = None
    try:
        _validate_fixture(fixture)
        if framework_runs:
            from .lifecycle_first_boot_framework import attach_framework_fixture

            attach_framework_fixture(fixture)
        _compose(fixture, "config", "--quiet")
        if _compose(fixture, "ps", "-q").stdout.strip():
            raise SystemExit("Isolated first-boot project already owns containers; refusing to reuse it.")
        print(f"=== Isolated Mycelis First-Boot Proof ({fixture.project}) ===")
        print("Fixture ports: " + ", ".join(f"{key}={port}" for key, port in fixture.ports.items()))
        armed = True
        _run_proof(fixture, build, user_tables, bootstrap_tables)
        if framework_runs:
            from .lifecycle_first_boot_framework import prove_framework_fixture

            prove_framework_fixture(fixture, build=build)
    finally:
        if armed:
            stopped = _compose(fixture, "stop", "core", "interface", check=False)
            if stopped.returncode == 0:
                try:
                    _repair_fixture_output_permissions(fixture)
                except (OSError, SystemExit) as exc:
                    repair_error = exc
            else:
                repair_error = SystemExit("Fixture app containers did not stop before output cleanup.")
            down_args = ("down", "--volumes", "--remove-orphans")
            if repair_error is None:
                down_args += ("--rmi", "local")
            result = _compose(fixture, *down_args, check=False)
            cleaned = result.returncode == 0 and repair_error is None
        if (cleaned or not armed) and _owned_fixture_root(fixture.root, fixture.project):
            try:
                shutil.rmtree(fixture.root)
            except OSError as exc:
                raise SystemExit(
                    f"Fixture project stopped, but output cleanup failed at {fixture.root}; "
                    "fixture files remain for scoped recovery."
                ) from exc
        if armed and not cleaned:
            raise SystemExit(
                f"Fixture cleanup failed ({repair_error or 'Compose down failed'}); inspect only project {fixture.project}. "
                f"Fixture files remain at {fixture.root}."
            )
    print("ISOLATED CLEAN FIRST-BOOT PROOF READY.")
