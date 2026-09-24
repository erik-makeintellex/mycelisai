"""Opt-in Runs controller proof inside the canonical disposable first-boot project."""

from __future__ import annotations

import base64
import ipaddress
import json
import os
import re
import secrets
import subprocess
from pathlib import Path
from urllib.parse import quote

from .config import ROOT_DIR, docker_command, docker_host_path
from . import lifecycle_first_boot_isolated as isolated

OVERLAY_NAME = "framework-runs.yml"
RUN_ID = "b2-fixture-sentinel"
RUN_TABLES = ("runs", "run_events", "run_commands", "run_approvals", "candidate_manifests")

CORE_PROBE_JS = r"""
const https = require('https');
const mode = process.argv[1];
const ca = Buffer.from(process.argv[2] || '', 'base64');
const run = 'b2-fixture-sentinel';
const create = {run_id: run, intent: 'Produce candidate', correlation: {
  run_id: run, intent_proof_id: 'proof-1', execution_contract_id: 'contract-1',
  work_item_id: 'work-1', idempotency_key: 'idem:' + run, source_kind: 'web_api',
  source_channel: 'api.intent', payload_kind: 'command', graph_revision: 'graph-1'}};
const body = mode === 'create' ? JSON.stringify(create) : '';
const headers = {};
if (mode !== 'no-token') headers.Authorization = 'Bearer ' +
  (mode === 'wrong-token' ? 'incorrect-credential' : process.env.MYCELIS_WORKER_API_KEY || '');
if (body) { headers['Content-Type'] = 'application/json'; headers['Content-Length'] = Buffer.byteLength(body); }
const path = mode === 'create' ? '/v1/runs' : mode === 'get' ? '/v1/runs/' + run : '/health';
const req = https.request({hostname: 'framework-runs-control', port: 8091, path,
  method: body ? 'POST' : 'GET', ca: mode === 'bad-trust' ? undefined : ca,
  headers, timeout: 2000}, res => {
    let text = '';
    res.on('data', chunk => { if (text.length < 8192) text += chunk.toString(); });
    res.on('end', () => {
      let value = {}; try { value = JSON.parse(text); } catch {}
      console.log(JSON.stringify({status: res.statusCode, body: value}));
    });
  });
req.on('timeout', () => req.destroy(new Error('timeout')));
req.on('error', error => { console.log(JSON.stringify({transport_error: error.code || 'error'})); process.exitCode = 2; });
req.end(body);
"""

WORKER_NETWORK_JS = r"""
const mode = process.argv[1];
if (mode === 'core-no-token') {
  const http = require('http');
  const req = http.get('http://core:8080/api/v1/templates', {timeout: 2000}, res => {
    res.resume(); res.on('end', () => console.log(JSON.stringify({status: res.statusCode})));
  });
  req.on('timeout', () => req.destroy(new Error('timeout')));
  req.on('error', error => { console.log(JSON.stringify({transport_error: error.code || 'error'})); process.exitCode = 2; });
} else if (mode === 'egress-route') {
  const fs = require('fs');
  const lines = fs.readFileSync('/proc/net/route', 'utf8').trim().split('\n').slice(1);
  console.log(JSON.stringify({default_route: lines.some(line => line.trim().split(/\s+/)[1] === '00000000')}));
} else { process.exitCode = 2; }
"""

TCP_PROBE_JS = r"""
const net = require('net');
const socket = net.createConnection({host: process.argv[1], port: Number(process.argv[2]), timeout: 2000});
socket.on('connect', () => { console.log(JSON.stringify({connected: true})); socket.destroy(); });
socket.on('timeout', () => socket.destroy(new Error('timeout')));
socket.on('error', error => console.log(JSON.stringify({connected: false, error: error.code || 'error'})));
"""


def _run_quiet(command: list[str], *, cwd: Path, timeout: int = 30) -> None:
    result = subprocess.run(command, cwd=cwd, capture_output=True, text=True, timeout=timeout)
    if result.returncode != 0:
        raise SystemExit("Disposable Runs TLS certificate generation failed; no external certificate was used.")


def _create_fixture_certificate(root: Path) -> tuple[Path, Path, Path]:
    directory = root / "framework-runs-tls"
    directory.mkdir(mode=0o700)
    ca_key, ca_cert = directory / "ca.key", directory / "ca.crt"
    server_key, server_csr = directory / "tls.key", directory / "server.csr"
    server_cert, extension = directory / "tls.crt", directory / "server.ext"
    extension.write_text(
        "subjectAltName=DNS:framework-runs-control\n"
        "extendedKeyUsage=serverAuth\nkeyUsage=digitalSignature,keyEncipherment\n",
        encoding="ascii",
    )
    _run_quiet(["openssl", "req", "-x509", "-newkey", "rsa:2048", "-nodes",
        "-keyout", str(ca_key), "-out", str(ca_cert), "-days", "1",
        "-subj", "/CN=mycelis-firstboot-runs-ca",
        "-addext", "basicConstraints=critical,CA:TRUE",
        "-addext", "keyUsage=critical,keyCertSign,cRLSign"], cwd=root)
    _run_quiet(["openssl", "req", "-newkey", "rsa:2048", "-nodes",
        "-keyout", str(server_key), "-out", str(server_csr),
        "-subj", "/CN=framework-runs-control"], cwd=root)
    _run_quiet(["openssl", "x509", "-req", "-in", str(server_csr),
        "-CA", str(ca_cert), "-CAkey", str(ca_key), "-CAcreateserial",
        "-out", str(server_cert), "-days", "1", "-extfile", str(extension)], cwd=root)
    for path in (ca_key, ca_cert, server_key, server_cert, server_csr, extension):
        path.chmod(0o600)
    return server_cert, server_key, ca_cert


def attach_framework_fixture(fixture: isolated.Fixture) -> None:
    isolated._validate_fixture(fixture)
    if not isolated.FRAMEWORK_OVERLAY_SOURCE.is_file():
        raise SystemExit("Canonical framework-runs Compose overlay is unavailable.")
    if not hasattr(os, "getuid") or os.getuid() != 1000:
        raise SystemExit("Runs fixture TLS key must be owned by the container's UID 1000.")
    cert, key, ca = _create_fixture_certificate(fixture.root)
    token = secrets.token_urlsafe(48)
    password = secrets.token_urlsafe(32)
    fixture.environment.update({
        "FRAMEWORK_RUNS_CORE_TOKEN": token,
        "FRAMEWORK_RUNS_DB_USER": "framework_runs",
        "FRAMEWORK_RUNS_DB_PASSWORD": password,
        "FRAMEWORK_RUNS_DB_NAME": "framework_runs",
        "FRAMEWORK_RUNS_DATABASE_URL": (
            "postgres://framework_runs:" + quote(password, safe="") +
            "@framework-runs-postgres:5432/framework_runs?sslmode=disable"
        ),
        "FRAMEWORK_RUNS_TLS_CERT_HOST_FILE": docker_host_path(cert),
        "FRAMEWORK_RUNS_TLS_KEY_HOST_FILE": docker_host_path(key),
        "FRAMEWORK_RUNS_TLS_CA_HOST_FILE": docker_host_path(ca),
    })
    (fixture.root / OVERLAY_NAME).write_text(
        isolated.FRAMEWORK_OVERLAY_SOURCE.read_text(encoding="utf-8"), encoding="utf-8",
    )
    fixture.env_file.write_text(
        "\n".join(f"{name}={value}" for name, value in sorted(fixture.environment.items())) + "\n",
        encoding="utf-8",
    )
    fixture.env_file.chmod(0o600)


def _runs_psql(fixture: isolated.Fixture, sql: str) -> str:
    result = isolated._compose(fixture, "exec", "-T", "framework-runs-postgres",
        "psql", "-X", "-A", "-t", "-v", "ON_ERROR_STOP=1",
        "-U", "framework_runs", "-d", "framework_runs", "-c", sql)
    return result.stdout.strip()


def _worker_probe(fixture: isolated.Fixture) -> bool:
    return isolated._compose(fixture, "exec", "-T", "framework-runs",
        "/framework-runs", "probe", check=False).returncode == 0


def prepare_framework_fixture(fixture: isolated.Fixture, *, build: bool) -> None:
    isolated._compose(fixture, "up", "-d", "framework-runs-postgres")
    isolated._wait_until("Runs fixture PostgreSQL", lambda: isolated._compose(
        fixture, "exec", "-T", "framework-runs-postgres", "pg_isready",
        "-U", "framework_runs", "-d", "framework_runs", check=False,
    ).returncode == 0)
    if build:
        isolated._compose(fixture, "build", "framework-runs")
    isolated._compose(fixture, "up", "-d", "--no-build", "framework-runs")
    isolated._wait_until("authenticated Runs controller readiness", lambda: _worker_probe(fixture))
    if _runs_psql(fixture, "SELECT schema_version FROM framework_runs_schema_metadata;") != "1":
        raise SystemExit("Runs fixture canonical journal schema is absent.")
    if any(int(_runs_psql(fixture, f"SELECT COUNT(*) FROM {table};")) != 0 for table in RUN_TABLES):
        raise SystemExit("Runs fixture journal was not empty at first boot.")


def _core_probe(fixture: isolated.Fixture, mode: str) -> dict:
    ca = Path(fixture.root / "framework-runs-tls" / "ca.crt").read_bytes()
    result = isolated._compose(fixture, "exec", "-T", "core", "node", "-e",
        CORE_PROBE_JS, mode, base64.b64encode(ca).decode("ascii"), check=False)
    try:
        return json.loads(result.stdout.strip())
    except json.JSONDecodeError as exc:
        raise SystemExit(f"Core namespace Runs probe {mode} did not return a bounded result.") from exc


def _expect_probe(fixture: isolated.Fixture, mode: str, *, status: int = 0, error: bool = False) -> dict:
    result = _core_probe(fixture, mode)
    if (error and not result.get("transport_error")) or (not error and result.get("status") != status):
        raise SystemExit(f"Core namespace Runs probe {mode} did not meet the expected transport/auth state.")
    return result


def _fixture_image(fixture: isolated.Fixture, service: str) -> str:
    image = isolated._compose(fixture, "images", "-q", service).stdout.strip()
    if not re.fullmatch(r"(?:sha256:)?[0-9a-f]{64}", image):
        raise SystemExit(f"Fixture {service} image identity is unavailable.")
    return image


def _fixture_container(fixture: isolated.Fixture, service: str) -> str:
    container = isolated._compose(fixture, "ps", "-q", service).stdout.strip()
    if not re.fullmatch(r"[0-9a-f]{64}", container):
        raise SystemExit(f"Fixture {service} container identity is unavailable.")
    return container


def _worker_network_probe(fixture: isolated.Fixture, mode: str, target: str = "") -> dict:
    command = docker_command(
        "run", "--rm", "--pull=never", "--read-only", "--cap-drop=ALL",
        "--security-opt=no-new-privileges", "--network", "container:" + _fixture_container(fixture, "framework-runs"),
        "--entrypoint", "node", _fixture_image(fixture, "core"), "-e", WORKER_NETWORK_JS, mode, target,
        cwd=ROOT_DIR,
    )
    result = subprocess.run(command, cwd=ROOT_DIR, env=isolated._compose_environment(fixture),
        capture_output=True, text=True, timeout=20)
    if result.returncode != 0:
        raise SystemExit(f"Runs network namespace probe {mode} failed before reporting a network result.")
    try:
        return json.loads(result.stdout.strip())
    except json.JSONDecodeError as exc:
        raise SystemExit(f"Runs network namespace probe {mode} did not return a bounded result.") from exc


def _tcp_probe(fixture: isolated.Fixture, namespace_service: str, host: str, port: int) -> dict:
    if namespace_service == "core":
        result = isolated._compose(fixture, "exec", "-T", "core", "node", "-e",
            TCP_PROBE_JS, host, str(port), check=False)
    else:
        command = docker_command(
            "run", "--rm", "--pull=never", "--read-only", "--cap-drop=ALL",
            "--security-opt=no-new-privileges", "--network", "container:" + _fixture_container(fixture, namespace_service),
            "--entrypoint", "node", _fixture_image(fixture, "core"), "-e", TCP_PROBE_JS, host, str(port),
            cwd=ROOT_DIR,
        )
        result = subprocess.run(command, cwd=ROOT_DIR, env=isolated._compose_environment(fixture),
            capture_output=True, text=True, timeout=20)
    if result.returncode != 0:
        raise SystemExit(f"Fixture {namespace_service} TCP probe failed before connection testing.")
    try:
        value = json.loads(result.stdout.strip())
    except json.JSONDecodeError as exc:
        raise SystemExit(f"Fixture {namespace_service} TCP probe returned no connection result.") from exc
    if value.get("connected") is False and value.get("error") not in {
        "ECONNREFUSED", "ETIMEDOUT", "EHOSTUNREACH", "ENETUNREACH", "timeout",
    }:
        raise SystemExit(f"Fixture {namespace_service} TCP probe failed for a non-network reason.")
    return value


def _service_ip(fixture: isolated.Fixture, service: str, network: str) -> str:
    command = docker_command("inspect", "--format",
        '{{(index .NetworkSettings.Networks "' + network + '").IPAddress}}',
        _fixture_container(fixture, service), cwd=ROOT_DIR)
    result = subprocess.run(command, cwd=ROOT_DIR, env=isolated._compose_environment(fixture),
        capture_output=True, text=True, timeout=10)
    ip = result.stdout.strip()
    try:
        parsed = ipaddress.ip_address(ip)
    except ValueError as exc:
        raise SystemExit(f"Fixture {service} network IP was unavailable.") from exc
    if result.returncode != 0 or parsed.version != 4:
        raise SystemExit(f"Fixture {service} network IP was unavailable.")
    return ip


def prove_framework_fixture(fixture: isolated.Fixture, *, build: bool) -> None:
    del build  # Image build and first boot are owned by prepare_framework_fixture.
    print("  Runs image: " + _fixture_image(fixture, "framework-runs"))
    ready = _expect_probe(fixture, "ready", status=200)
    body = ready.get("body", {})
    if body.get("healthy") is not True or body.get("controller_ready") is not True or body.get("production_ready") is not False:
        raise SystemExit("Runs readiness confused controller availability with executor readiness.")
    _expect_probe(fixture, "no-token", status=401)
    _expect_probe(fixture, "wrong-token", status=401)
    bad_trust = _expect_probe(fixture, "bad-trust", error=True)
    if bad_trust.get("transport_error") not in {
        "UNABLE_TO_VERIFY_LEAF_SIGNATURE", "SELF_SIGNED_CERT_IN_CHAIN",
        "DEPTH_ZERO_SELF_SIGNED_CERT", "ERR_TLS_CERT_ALTNAME_INVALID",
    }:
        raise SystemExit("Bad-trust probe failed for a reason other than TLS verification.")
    create = _expect_probe(fixture, "create", status=503)
    if create.get("body", {}).get("error", {}).get("code") != "unready":
        raise SystemExit("Protocol-only create did not fail for unavailable executor.")
    if _runs_psql(fixture, "SELECT COUNT(*) FROM runs;") != "0":
        raise SystemExit("Protocol-only Runs create persisted work without an executor.")

    sentinel = (
        "INSERT INTO runs (run_id,idempotency_key,request_digest,request_json,snapshot_json,"
        "status,version,next_sequence,created_at,updated_at) VALUES ("
        f"'{RUN_ID}','idem:{RUN_ID}',repeat('a',64),"
        f"'{{\"run_id\":\"{RUN_ID}\"}}'::jsonb,"
        f"'{{\"run_id\":\"{RUN_ID}\",\"status\":\"accepted\",\"version\":1}}'::jsonb,"
        "'accepted',1,2,NOW(),NOW());"
    )
    _runs_psql(fixture, sentinel)  # SQL-only volume sentinel, not an accepted API run.
    old_container = _fixture_container(fixture, "framework-runs")
    isolated._compose(fixture, "up", "-d", "--no-deps", "--force-recreate", "--no-build", "framework-runs")
    isolated._wait_until("Runs controller after container recreation", lambda: _worker_probe(fixture))
    if _fixture_container(fixture, "framework-runs") == old_container:
        raise SystemExit("Runs worker container was not recreated.")
    if _runs_psql(fixture, f"SELECT COUNT(*) FROM runs WHERE run_id='{RUN_ID}';") != "1":
        raise SystemExit("Runs journal sentinel was lost after container recreation.")
    _expect_probe(fixture, "get", status=200)

    data_ip = _service_ip(fixture, "framework-runs", fixture.project + "_framework-runs-data")
    if _tcp_probe(fixture, "framework-runs-postgres", "127.0.0.1", 5432).get("connected") is not True:
        raise SystemExit("Runs DB namespace TCP control did not reach its own healthy database.")
    if _tcp_probe(fixture, "framework-runs-postgres", data_ip, 8091).get("connected") is not False:
        raise SystemExit("Runs API accepted TCP on its data-network interface.")
    no_core_token = _worker_network_probe(fixture, "core-no-token")
    if no_core_token.get("status") != 401:
        raise SystemExit("Runs namespace reached Core without the required Core credential boundary.")
    route = _worker_network_probe(fixture, "egress-route")
    if route.get("default_route") is not False:
        raise SystemExit("Runs namespace has a default egress route.")
    nats_ip = _service_ip(fixture, "nats", fixture.project + "_default")
    if _tcp_probe(fixture, "core", nats_ip, 4222).get("connected") is not True:
        raise SystemExit("Fixture Core namespace could not reach the healthy control NATS target.")
    if _tcp_probe(fixture, "framework-runs", nats_ip, 4222).get("connected") is not False:
        raise SystemExit("Runs namespace reached the fixture's separate default network.")

    old_database_container = _fixture_container(fixture, "framework-runs-postgres")
    isolated._compose(fixture, "stop", "framework-runs-postgres")
    isolated._wait_until("Runs HTTP 503 while DB is down",
        lambda: _core_probe(fixture, "ready").get("status") == 503, timeout=40)
    isolated._compose(fixture, "up", "-d", "--no-deps", "--force-recreate", "framework-runs-postgres")
    isolated._wait_until("Runs database recovered", lambda: isolated._compose(
        fixture, "exec", "-T", "framework-runs-postgres", "pg_isready",
        "-U", "framework_runs", "-d", "framework_runs", check=False,
    ).returncode == 0)
    isolated._wait_until("Runs authenticated readiness recovered", lambda: _worker_probe(fixture))
    if _fixture_container(fixture, "framework-runs-postgres") == old_database_container:
        raise SystemExit("Runs database container was not recreated.")
    if _runs_psql(fixture, f"SELECT COUNT(*) FROM runs WHERE run_id='{RUN_ID}';") != "1":
        raise SystemExit("Runs journal sentinel was lost after database container recreation.")
    _expect_probe(fixture, "get", status=200)
    print("  [OK] Protocol-only Runs TLS/auth, isolation, SQL-seeded journal and DB recreation, outage and recovery")
