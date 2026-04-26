from __future__ import annotations

import json
import subprocess
import time
from pathlib import Path
from typing import Callable


RelayLabels = dict[str, str] | None
DockerRun = Callable[[list[str], bool], subprocess.CompletedProcess[str]]


def docker_run(
    args: list[str],
    *,
    docker_command: Callable[..., list[str]],
    root_dir: Path,
    check: bool = True,
) -> subprocess.CompletedProcess[str]:
    result = subprocess.run(
        docker_command(*args, cwd=root_dir),
        capture_output=True,
        text=True,
    )
    if check and result.returncode != 0:
        detail = result.stderr.strip() or result.stdout.strip() or "unknown docker error"
        raise SystemExit(f"Docker command failed: {' '.join(args)} ({detail})")
    return result


def inspect_labels(
    *,
    relay_name: str,
    docker_runner: DockerRun,
) -> RelayLabels:
    result = docker_runner(
        ["inspect", relay_name, "--format", "{{json .Config.Labels}}"],
        False,
    )
    if result.returncode != 0:
        return None
    raw = result.stdout.strip()
    if not raw:
        return {}
    try:
        return json.loads(raw)
    except json.JSONDecodeError:
        return None


def stop(
    *,
    relay_name: str,
    docker_runner: DockerRun,
    docker_host_mode: Callable[[], str],
    running_in_wsl: Callable[[], bool],
):
    if not (docker_host_mode() == "wsl" or running_in_wsl()):
        return
    docker_runner(["rm", "-f", relay_name], False)


def ensure(
    target_host: str,
    target_port: int,
    relay_port: int,
    *,
    relay_name: str,
    relay_image: str,
    inspect_relay_labels: Callable[[], RelayLabels],
    stop_relay: Callable[[], None],
    docker_runner: DockerRun,
):
    labels = inspect_relay_labels()
    expected = {
        "mycelis.relay.target_host": target_host,
        "mycelis.relay.target_port": str(target_port),
        "mycelis.relay.listen_port": str(relay_port),
    }
    if labels and all(labels.get(key) == value for key, value in expected.items()):
        return

    stop_relay()
    relay_cmd = (
        "apk add --no-cache socat >/dev/null && "
        f"exec socat TCP-LISTEN:{relay_port},fork,reuseaddr,bind=0.0.0.0 "
        f"TCP:{target_host}:{target_port}"
    )
    result = docker_runner(
        [
            "run",
            "-d",
            "--rm",
            "--name",
            relay_name,
            "--network",
            "host",
            "--label",
            f"mycelis.relay.target_host={target_host}",
            "--label",
            f"mycelis.relay.target_port={target_port}",
            "--label",
            f"mycelis.relay.listen_port={relay_port}",
            relay_image,
            "sh",
            "-lc",
            relay_cmd,
        ],
        True,
    )
    if not result.stdout.strip():
        raise SystemExit("Failed to start the WSL Ollama relay container.")
    time.sleep(2)


def prepare_host(
    env_values: dict[str, str],
    *,
    docker_host_mode: Callable[[], str],
    running_in_wsl: Callable[[], bool],
    clean_env_value: Callable[[str], str],
    parse_network_endpoint: Callable[[str], tuple[str, int]],
    relay_port: Callable[[dict[str, str]], int],
    inspect_relay_labels: Callable[[], RelayLabels],
    wsl_http_available: Callable[[str], bool],
    ensure_relay: Callable[[str, int, int], None],
) -> dict[str, str]:
    values = dict(env_values)
    if not (docker_host_mode() == "wsl" or running_in_wsl()):
        return values

    configured = clean_env_value(
        values.get("MYCELIS_COMPOSE_OLLAMA_HOST", "http://host.docker.internal:11434")
    )
    configured_host, configured_port = parse_network_endpoint(configured)
    listen_port = relay_port(values)
    labels = inspect_relay_labels()
    if labels and labels.get("mycelis.relay.listen_port") == str(listen_port):
        target_port = labels.get("mycelis.relay.target_port", "")
        target_host = labels.get("mycelis.relay.target_host", "")
        if target_port == str(configured_port) and target_host in {configured_host, "127.0.0.1"}:
            print(
                "  WSL Ollama relay: "
                f"http://host.docker.internal:{listen_port} -> {target_host}:{target_port}"
            )
            values["MYCELIS_COMPOSE_OLLAMA_HOST"] = f"http://host.docker.internal:{listen_port}"
            return values

    relay_target_host = configured_host
    relay_target_port = configured_port
    if not wsl_http_available(configured):
        localhost_candidate = f"http://127.0.0.1:{configured_port}"
        if wsl_http_available(localhost_candidate):
            relay_target_host = "127.0.0.1"
        else:
            raise SystemExit(
                "WSL Docker could not reach the configured MYCELIS_COMPOSE_OLLAMA_HOST and "
                "no mirrored localhost Ollama fallback was reachable from WSL. "
                "Verify the Windows Ollama service is running and reachable from the Docker-owning WSL distro."
            )

    ensure_relay(relay_target_host, relay_target_port, listen_port)
    print(
        "  WSL Ollama relay: "
        f"http://host.docker.internal:{listen_port} -> {relay_target_host}:{relay_target_port}"
    )
    values["MYCELIS_COMPOSE_OLLAMA_HOST"] = f"http://host.docker.internal:{listen_port}"
    return values
