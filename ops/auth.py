from __future__ import annotations

import secrets
import string
from pathlib import Path

from invoke import Collection, task

from .config import ROOT_DIR

KEY_NAME = "MYCELIS_API_KEY"
SAMPLE_VALUE = "mycelis-dev-key-change-in-prod"
ENV_PATH = ROOT_DIR / ".env"
ENV_EXAMPLE_PATH = ROOT_DIR / ".env.example"


def _generate_dev_key(length: int = 40) -> str:
    alphabet = string.ascii_letters + string.digits
    suffix = "".join(secrets.choice(alphabet) for _ in range(length))
    return f"mycelis-dev-{suffix}"


def _read_env_value(path: Path, key: str) -> str:
    if not path.exists():
        return ""

    for raw in path.read_text(encoding="utf-8").splitlines():
        line = raw.strip()
        if not line or line.startswith("#") or "=" not in raw:
            continue
        name, value = raw.split("=", 1)
        if name.strip() == key:
            return value.strip().strip('"').strip("'")
    return ""


def _upsert_env_value(path: Path, key: str, value: str) -> None:
    lines: list[str] = []
    if path.exists():
        lines = path.read_text(encoding="utf-8").splitlines()

    replaced = False
    for idx, raw in enumerate(lines):
        line = raw.strip()
        if not line or line.startswith("#") or "=" not in raw:
            continue
        name, _ = raw.split("=", 1)
        if name.strip() == key:
            lines[idx] = f"{key}={value}"
            replaced = True
            break

    if not replaced:
        if lines and lines[-1].strip():
            lines.append("")
        lines.append(f"{key}={value}")

    path.write_text("\n".join(lines).rstrip() + "\n", encoding="utf-8")


def _mask_secret(value: str) -> str:
    if len(value) <= 10:
        return "*" * len(value)
    return f"{value[:8]}...{value[-4:]}"


@task(
    help={
        "rotate": "Rotate and replace the key in .env even if one already exists.",
        "show": "Print the full key value (default is masked).",
        "value": "Use an explicit key value instead of generating one.",
    }
)
def dev_key(_c, rotate=False, show=False, value=""):
    """
    Ensure a local MYCELIS_API_KEY exists and keep .env.example on a sample value.
    """
    if not ENV_PATH.exists():
        raise SystemExit("Missing .env. Copy .env.example to .env first.")

    explicit_value = value.strip()
    existing = _read_env_value(ENV_PATH, KEY_NAME)
    action = "kept existing"

    if explicit_value:
        key = explicit_value
        _upsert_env_value(ENV_PATH, KEY_NAME, key)
        action = "set explicit value"
    elif not existing:
        key = _generate_dev_key()
        _upsert_env_value(ENV_PATH, KEY_NAME, key)
        action = "generated new key"
    elif rotate:
        key = _generate_dev_key()
        _upsert_env_value(ENV_PATH, KEY_NAME, key)
        action = "rotated key"
    else:
        key = existing

    if ENV_EXAMPLE_PATH.exists():
        _upsert_env_value(ENV_EXAMPLE_PATH, KEY_NAME, SAMPLE_VALUE)

    visible = key if show else _mask_secret(key)
    print(f"{KEY_NAME}: {visible}")
    print(f"Action: {action}")
    print("Next: restart services to apply auth key changes:")
    print("  uv run inv lifecycle.restart")


ns = Collection("auth")
ns.add_task(dev_key, name="dev-key")
