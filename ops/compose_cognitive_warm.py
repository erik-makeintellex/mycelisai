from __future__ import annotations

import json
import urllib.error
import urllib.request

from invoke import task

from . import compose_env
from . import compose_wsl_relay
from .config import ROOT_DIR, docker_command, docker_host_mode, running_in_wsl


COMPOSE_ENV_FILE = ROOT_DIR / ".env.compose"
COMPOSE_ENV_EXAMPLE = ROOT_DIR / ".env.compose.example"
COMPOSE_PROJECT = "mycelis-home"
DEFAULT_OUTPUT_HOST_PATH = ROOT_DIR / "workspace" / "docker-compose" / "data"
DEFAULT_WSL_OLLAMA_RELAY_PORT = 11435
WSL_OLLAMA_RELAY_IMAGE = "alpine:3.21"
WSL_OLLAMA_RELAY_NAME = f"{COMPOSE_PROJECT}-ollama-relay"


def _require_compose_env_file():
    return compose_env.require_compose_env_file(COMPOSE_ENV_FILE, COMPOSE_ENV_EXAMPLE)


def _load_compose_env() -> dict[str, str]:
    return compose_env.load_compose_env(COMPOSE_ENV_FILE, _require_compose_env_file)


def _compose_effective_env(env_values: dict[str, str] | None = None) -> dict[str, str]:
    return compose_env.compose_effective_env(env_values, _load_compose_env)


def _validate_output_block_config(env_values: dict[str, str]):
    return compose_env.validate_output_block_config(env_values, DEFAULT_OUTPUT_HOST_PATH)


def _validate_compose_env(env_values: dict[str, str]):
    return compose_env.validate_compose_env(env_values, _validate_output_block_config)


def _docker_run(args: list[str], check: bool = True):
    return compose_wsl_relay.docker_run(
        args,
        docker_command=docker_command,
        root_dir=ROOT_DIR,
        check=check,
    )


def _inspect_wsl_ollama_relay_labels() -> dict[str, str] | None:
    return compose_wsl_relay.inspect_labels(
        relay_name=WSL_OLLAMA_RELAY_NAME,
        docker_runner=_docker_run,
    )


def _stop_wsl_ollama_relay():
    compose_wsl_relay.stop(
        relay_name=WSL_OLLAMA_RELAY_NAME,
        docker_runner=_docker_run,
        docker_host_mode=docker_host_mode,
        running_in_wsl=running_in_wsl,
    )


def _ensure_wsl_ollama_relay(target_host: str, target_port: int, relay_port: int):
    compose_wsl_relay.ensure(
        target_host,
        target_port,
        relay_port,
        relay_name=WSL_OLLAMA_RELAY_NAME,
        relay_image=WSL_OLLAMA_RELAY_IMAGE,
        inspect_relay_labels=_inspect_wsl_ollama_relay_labels,
        stop_relay=_stop_wsl_ollama_relay,
        docker_runner=_docker_run,
    )


def _prepare_wsl_ollama_host(env_values: dict[str, str]) -> dict[str, str]:
    return compose_wsl_relay.prepare_host(
        env_values,
        docker_host_mode=docker_host_mode,
        running_in_wsl=running_in_wsl,
        clean_env_value=compose_env.clean_env_value,
        parse_network_endpoint=compose_env.parse_network_endpoint,
        relay_port=lambda values: compose_env.wsl_ollama_relay_port(values, DEFAULT_WSL_OLLAMA_RELAY_PORT),
        inspect_relay_labels=_inspect_wsl_ollama_relay_labels,
        wsl_http_available=lambda url: compose_env.wsl_http_available(
            url,
            (lambda *args: list(args)) if running_in_wsl() else compose_env.wsl_exec_command,
        ),
        ensure_relay=_ensure_wsl_ollama_relay,
    )


def _compose_ollama_probe_base_url(env_values: dict[str, str]) -> str:
    configured = compose_env.clean_env_value(
        env_values.get("MYCELIS_COMPOSE_OLLAMA_HOST", "http://host.docker.internal:11434")
    )
    host, port = compose_env.parse_network_endpoint(configured)
    if (docker_host_mode() == "wsl" or running_in_wsl()) and host == "host.docker.internal":
        return f"http://127.0.0.1:{port}"
    return configured.rstrip("/")


def _configured_ollama_model_id() -> str:
    try:
        import yaml

        config = yaml.safe_load((ROOT_DIR / "core" / "config" / "cognitive.yaml").read_text(encoding="utf-8")) or {}
        provider = (config.get("providers") or {}).get("ollama") or {}
        model_id = str(provider.get("model_id") or "").strip()
        if model_id:
            return model_id
    except Exception:
        pass
    return "qwen3:14b"


@task
def warm_cognitive(c):
    """Warm the Compose text model through the same endpoint Core uses."""
    del c
    _require_compose_env_file()
    env_values = _compose_effective_env()
    _validate_compose_env(env_values)
    env_values = _prepare_wsl_ollama_host(env_values)
    model_id = _configured_ollama_model_id()
    warm_url = f"{_compose_ollama_probe_base_url(env_values)}/v1/chat/completions"
    payload = {
        "model": model_id,
        "messages": [{"role": "user", "content": "/no_think Reply READY."}],
        "stream": False,
        "max_tokens": 8,
    }
    print("=== Mycelis Compose Cognitive Warm-Up ===\n")
    print(f"  model: {model_id}")
    print(f"  endpoint: {warm_url}")
    try:
        request = urllib.request.Request(
            warm_url,
            data=json.dumps(payload).encode("utf-8"),
            headers={"Content-Type": "application/json"},
            method="POST",
        )
        with urllib.request.urlopen(request, timeout=240) as response:
            body = response.read().decode("utf-8", errors="replace")
            if response.status != 200:
                raise SystemExit(f"Cognitive warm-up failed: {warm_url} returned HTTP {response.status}.")
    except urllib.error.HTTPError as exc:
        detail = exc.read().decode("utf-8", errors="replace")
        raise SystemExit(f"Cognitive warm-up failed: {warm_url} returned HTTP {exc.code}: {detail}") from exc
    except Exception as exc:
        raise SystemExit(f"Cognitive warm-up failed: {exc}") from exc
    try:
        decoded = json.loads(body)
    except json.JSONDecodeError as exc:
        raise SystemExit(f"Cognitive warm-up failed: invalid JSON response ({exc}).") from exc
    if not decoded.get("choices"):
        raise SystemExit("Cognitive warm-up failed: response did not include choices.")
    print("  [OK] text model completed a warm-up chat response.")
