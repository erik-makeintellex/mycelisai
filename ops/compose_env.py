import os
import shlex
import socket
import subprocess
from pathlib import Path
from urllib.parse import urlparse


COMPOSE_RUNTIME_OVERRIDE_KEYS = {
    "CORS_ORIGIN",
    "DATA_DIR",
    "DB_HOST",
    "DB_NAME",
    "DB_PASSWORD",
    "DB_PORT",
    "DB_USER",
    "MYCELIS_API_KEY",
    "MYCELIS_BOOTSTRAP_TEMPLATE_ID",
    "MYCELIS_COMPOSE_CORE_PORT",
    "MYCELIS_COMPOSE_INTERFACE_PORT",
    "MYCELIS_COMPOSE_NATS_MONITOR_PORT",
    "MYCELIS_COMPOSE_NATS_PORT",
    "MYCELIS_COMPOSE_OLLAMA_HOST",
    "MYCELIS_COMPOSE_POSTGRES_PORT",
    "MYCELIS_COMPOSE_WSL_OLLAMA_RELAY_PORT",
    "MYCELIS_DISABLE_DEFAULT_MCP_BOOTSTRAP",
    "MYCELIS_OUTPUT_BLOCK_MODE",
    "MYCELIS_OUTPUT_HOST_PATH",
    "MYCELIS_SEARCH_MAX_RESULTS",
    "MYCELIS_SEARCH_LOCAL_API_ENDPOINT",
    "MYCELIS_SEARCH_PROVIDER",
    "MYCELIS_SEARXNG_ENDPOINT",
    "MYCELIS_WORKSPACE",
    "NATS_URL",
    "POSTGRES_DB",
    "POSTGRES_PASSWORD",
    "POSTGRES_USER",
}

OUTPUT_BLOCK_MODES = {"local_hosted", "cluster_generated"}


def compose_command(
    root_dir: Path,
    compose_project: str,
    compose_env_file: Path,
    compose_file: Path,
    docker_command,
    docker_host_path,
    *args: str,
) -> list[str]:
    return docker_command(
        "compose",
        "--project-name",
        compose_project,
        "--env-file",
        docker_host_path(compose_env_file),
        "-f",
        docker_host_path(compose_file),
        *args,
        cwd=root_dir,
    )


def require_compose_env_file(compose_env_file: Path, compose_env_example: Path):
    if compose_env_file.exists():
        return
    raise SystemExit(
        f"Missing {compose_env_file.name}. Copy {compose_env_example.name} to "
        f"{compose_env_file.name} and set MYCELIS_API_KEY before running compose tasks."
    )


def load_compose_env(compose_env_file: Path, require_env_file) -> dict[str, str]:
    require_env_file()
    values: dict[str, str] = {}
    for raw_line in compose_env_file.read_text(encoding="utf-8").splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#") or "=" not in line:
            continue
        key, value = line.split("=", 1)
        values[key.strip()] = value.strip()
    return values


def clean_env_value(value: str) -> str:
    return value.strip().strip('"').strip("'")


def compose_effective_env(
    env_values: dict[str, str] | None,
    load_env,
    *,
    environ: dict[str, str] | None = None,
) -> dict[str, str]:
    source_environ = environ if environ is not None else os.environ
    values = dict(env_values or load_env())
    for key in COMPOSE_RUNTIME_OVERRIDE_KEYS:
        override = source_environ.get(key)
        if override is None:
            continue
        cleaned = clean_env_value(override)
        if cleaned:
            values[key] = cleaned
    return values


def wsl_exec_command(*args: str, environ: dict[str, str] | None = None) -> list[str]:
    source_environ = environ if environ is not None else os.environ
    command = ["wsl.exe"]
    distro = source_environ.get("MYCELIS_WSL_DISTRO", "").strip()
    if distro:
        command.extend(["-d", distro])
    command.extend(["--exec", *args])
    return command


def port_open(port: int, host: str = "127.0.0.1", timeout: float = 1.0) -> bool:
    try:
        with socket.create_connection((host, port), timeout=timeout):
            return True
    except (ConnectionRefusedError, TimeoutError, OSError):
        return False


def normalize_bool(value: str) -> bool:
    return value.strip().lower() in {"1", "true", "yes", "on"}


def looks_like_container_loopback(url: str) -> bool:
    candidate = url.strip().lower()
    for prefix in ("http://", "https://"):
        if candidate.startswith(prefix):
            candidate = candidate[len(prefix):]
            break
    candidate = candidate.split("/", 1)[0]
    host = candidate.split(":", 1)[0]
    return host in {"127.0.0.1", "localhost", "0.0.0.0"}


def resolve_host_path(path_value: str) -> Path:
    raw_path = clean_env_value(path_value)
    expanded = os.path.expandvars(os.path.expanduser(raw_path))
    return Path(expanded).resolve(strict=False)


def parse_network_endpoint(url: str) -> tuple[str, int]:
    parsed = urlparse(url.strip())
    if parsed.scheme not in {"http", "https"} or not parsed.hostname:
        raise SystemExit(
            "Invalid .env.compose MYCELIS_COMPOSE_OLLAMA_HOST: "
            f"{url}. Use an http(s) endpoint such as http://host.docker.internal:11434."
        )
    port = parsed.port or (443 if parsed.scheme == "https" else 80)
    return parsed.hostname, port


def wsl_http_available(url: str, wsl_exec, run=subprocess.run) -> bool:
    probe = url.rstrip("/") + "/api/tags"
    try:
        result = run(
            wsl_exec(
                "sh",
                "-lc",
                f"curl -fsS --max-time 5 -o /dev/null {shlex.quote(probe)}",
            ),
            capture_output=True,
            text=True,
            timeout=15,
        )
    except subprocess.TimeoutExpired:
        return False
    return result.returncode == 0


def wsl_ollama_relay_port(env_values: dict[str, str], default_port: int) -> int:
    raw = clean_env_value(
        env_values.get("MYCELIS_COMPOSE_WSL_OLLAMA_RELAY_PORT", str(default_port))
    )
    try:
        return int(raw)
    except ValueError as exc:
        raise SystemExit(
            "Invalid .env.compose MYCELIS_COMPOSE_WSL_OLLAMA_RELAY_PORT: "
            f"{raw!r} must be an integer port."
        ) from exc


def validate_output_block_config(
    env_values: dict[str, str],
    default_output_host_path: Path,
    *,
    output_block_modes: set[str] = OUTPUT_BLOCK_MODES,
):
    mode_explicit = "MYCELIS_OUTPUT_BLOCK_MODE" in env_values
    mode = clean_env_value(env_values.get("MYCELIS_OUTPUT_BLOCK_MODE", "local_hosted")).lower()
    if mode not in output_block_modes:
        raise SystemExit(
            "Invalid .env.compose MYCELIS_OUTPUT_BLOCK_MODE: "
            f"{mode}. Use one of: {', '.join(sorted(output_block_modes))}."
        )

    path_explicit = "MYCELIS_OUTPUT_HOST_PATH" in env_values
    raw_path = clean_env_value(env_values.get("MYCELIS_OUTPUT_HOST_PATH", ""))
    if not raw_path:
        if mode == "local_hosted" and mode_explicit:
            raise SystemExit(
                "MYCELIS_OUTPUT_HOST_PATH is required when MYCELIS_OUTPUT_BLOCK_MODE=local_hosted. "
                "Set it to the host directory Docker should mount as Core /data."
            )
        raw_path = str(default_output_host_path)

    host_path = resolve_host_path(raw_path)
    if host_path.exists() and not host_path.is_dir():
        raise SystemExit(
            "Invalid .env.compose MYCELIS_OUTPUT_HOST_PATH: "
            f"{host_path} exists but is not a directory."
        )
    if not host_path.exists():
        if mode == "local_hosted" and (mode_explicit or path_explicit):
            raise SystemExit(
                "Invalid .env.compose MYCELIS_OUTPUT_HOST_PATH: "
                f"{host_path} does not exist. Create the directory first so Docker mounts the intended output block."
            )
        host_path.mkdir(parents=True, exist_ok=True)


def validate_compose_env(env_values: dict[str, str], validate_output_block):
    ollama_host = env_values.get("MYCELIS_COMPOSE_OLLAMA_HOST", "http://host.docker.internal:11434").strip()
    if ollama_host and looks_like_container_loopback(ollama_host):
        raise SystemExit(
            "Invalid .env.compose MYCELIS_COMPOSE_OLLAMA_HOST for Docker Compose: "
            f"{ollama_host}. Use a host-reachable address such as "
            "http://host.docker.internal:11434 or another container/service hostname."
        )
    validate_output_block(env_values)


def compose_runtime_env(
    env_values: dict[str, str] | None,
    docker_host_mode,
    running_in_wsl,
    effective_env,
    default_output_host_path: Path,
    docker_host_path,
    *,
    environ: dict[str, str] | None = None,
) -> dict[str, str] | None:
    host_mode = docker_host_mode()
    wsl_shell = running_in_wsl()
    if host_mode != "wsl" and not wsl_shell:
        return None

    values = effective_env(env_values)
    raw_host_path = clean_env_value(values.get("MYCELIS_OUTPUT_HOST_PATH", ""))
    resolved_host_path = resolve_host_path(raw_host_path) if raw_host_path else default_output_host_path

    env = dict(environ if environ is not None else os.environ)
    passthrough_keys = sorted(values.keys())
    for key in passthrough_keys:
        env[key] = values[key]
    env["MYCELIS_OUTPUT_HOST_PATH"] = (
        docker_host_path(resolved_host_path) if host_mode == "wsl" else str(resolved_host_path)
    )
    if host_mode == "wsl":
        passthrough_entries = [entry for entry in env.get("WSLENV", "").split(":") if entry]
        if "MYCELIS_OUTPUT_HOST_PATH" not in passthrough_keys:
            passthrough_keys.append("MYCELIS_OUTPUT_HOST_PATH")
        for key in passthrough_keys:
            if key not in passthrough_entries:
                passthrough_entries.append(key)
        env["WSLENV"] = ":".join(passthrough_entries)
    return env
