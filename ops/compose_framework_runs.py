"""Fail-closed selection and validation for the opt-in Runs Compose overlay."""

from __future__ import annotations

import os
import subprocess
from pathlib import Path
from urllib.parse import parse_qs, unquote, urlsplit


OPT_IN_KEY = "MYCELIS_COMPOSE_FRAMEWORK_RUNS"
SECRET_INPUT_KEYS = (
    "FRAMEWORK_RUNS_CORE_TOKEN",
    "FRAMEWORK_RUNS_DB_USER",
    "FRAMEWORK_RUNS_DB_PASSWORD",
    "FRAMEWORK_RUNS_DB_NAME",
    "FRAMEWORK_RUNS_DATABASE_URL",
)
TLS_HOST_FILE_KEYS = (
    "FRAMEWORK_RUNS_TLS_CERT_HOST_FILE",
    "FRAMEWORK_RUNS_TLS_KEY_HOST_FILE",
    "FRAMEWORK_RUNS_TLS_CA_HOST_FILE",
)


def enabled(environ: dict[str, str] | None = None) -> bool:
    raw = (environ if environ is not None else os.environ).get(OPT_IN_KEY, "").strip().lower()
    if raw in {"", "0", "false", "no", "off"}:
        return False
    if raw in {"1", "true", "yes", "on"}:
        return True
    raise SystemExit(f"{OPT_IN_KEY} must be explicitly true or false.")


def overlay_file(base_compose_file: Path) -> Path:
    path = base_compose_file.parent / "deploy" / "compose" / "framework-runs.yml"
    if not path.is_file():
        raise SystemExit("Opt-in framework Runs Compose overlay is missing.")
    return path


def _env_values(path: Path) -> dict[str, str]:
    if not path.is_file():
        return {}
    return {
        line.split("=", 1)[0].strip(): line.split("=", 1)[1].strip()
        for line in path.read_text(encoding="utf-8").splitlines()
        if line.strip() and not line.lstrip().startswith("#") and "=" in line
    }


def validate(values: dict[str, str], *, root_dir: Path) -> None:
    if not enabled():
        return
    source_values = _env_values(root_dir / ".env")
    if any(not source_values.get(key, "").strip() or values.get(key) != source_values[key] for key in SECRET_INPUT_KEYS):
        raise SystemExit("Runs opt-in requires scoped token and dedicated database credentials/URL in .env.")
    if any(
        value != value.strip() or value.startswith(("'", '"')) or value.endswith(("'", '"'))
        for value in (values[key] for key in SECRET_INPUT_KEYS)
    ):
        raise SystemExit("Runs scoped credentials must be unquoted canonical .env values.")
    token = values["FRAMEWORK_RUNS_CORE_TOKEN"]
    if len(token) < 32 or token == values.get("MYCELIS_API_KEY", "").strip().strip('"\''):
        raise SystemExit("Runs service token must be canonical, scoped, and at least 32 bytes.")
    if values["FRAMEWORK_RUNS_DB_PASSWORD"] == values.get("DB_PASSWORD", "").strip().strip('"\''):
        raise SystemExit("Runs database credential must be separate from Core PostgreSQL.")

    try:
        url = urlsplit(values["FRAMEWORK_RUNS_DATABASE_URL"])
        query = parse_qs(url.query, strict_parsing=True)
        matches = (
            url.scheme in {"postgres", "postgresql"}
            and url.hostname == "framework-runs-postgres"
            and url.port == 5432
            and unquote(url.username or "") == values["FRAMEWORK_RUNS_DB_USER"]
            and unquote(url.password or "") == values["FRAMEWORK_RUNS_DB_PASSWORD"]
            and unquote(url.path.lstrip("/")) == values["FRAMEWORK_RUNS_DB_NAME"]
            and not url.fragment
            and query == {"sslmode": ["disable"]}
        )
    except (ValueError, TypeError):
        matches = False
    if not matches:
        raise SystemExit("Runs database URL must match its dedicated Compose database and credentials.")

    for key in TLS_HOST_FILE_KEYS:
        raw = values.get(key, "").strip()
        if not raw:
            raise SystemExit(f"Runs opt-in requires {key}.")
        path = Path(os.path.expandvars(os.path.expanduser(raw))).resolve(strict=False)
        if not path.is_file():
            raise SystemExit(f"Runs TLS input {key} must name an existing regular file.")


def runtime_values(
    values: dict[str, str], *, docker_host_path, host_mode: str,
) -> dict[str, str]:
    if not enabled():
        return values
    converted = dict(values)
    for key in TLS_HOST_FILE_KEYS:
        raw = converted.get(key, "").strip()
        if not raw:
            continue
        path = Path(os.path.expandvars(os.path.expanduser(raw))).resolve(strict=False)
        converted[key] = docker_host_path(path) if host_mode == "wsl" else str(path)
    return converted


def verify_health(values: dict[str, str], *, compose_command, runtime_env) -> None:
    if not enabled():
        return
    for service, argv in (
        ("framework-runs-postgres", ["pg_isready", "-U", values["FRAMEWORK_RUNS_DB_USER"], "-d", values["FRAMEWORK_RUNS_DB_NAME"]]),
        ("framework-runs", ["/framework-runs", "probe"]),
    ):
        command = compose_command("exec", "-T", service, *argv)
        try:
            result = subprocess.run(command, capture_output=True, text=True, env=runtime_env(), timeout=15)
        except subprocess.TimeoutExpired:
            raise SystemExit(f"Compose framework Runs health timed out: {service}.") from None
        if result.returncode != 0:
            raise SystemExit(f"Compose framework Runs health failed: {service} is not ready.")
        print(f"  [OK] {service} authenticated readiness" if service == "framework-runs" else f"  [OK] {service} readiness")
