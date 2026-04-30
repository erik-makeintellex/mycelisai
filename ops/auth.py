from __future__ import annotations

import secrets
import string
from pathlib import Path

from invoke import Collection, task

from .config import ROOT_DIR

KEY_NAME = "MYCELIS_API_KEY"
BREAK_GLASS_KEY_NAME = "MYCELIS_BREAK_GLASS_API_KEY"
LOCAL_ADMIN_USERNAME_NAME = "MYCELIS_LOCAL_ADMIN_USERNAME"
LOCAL_ADMIN_USER_ID_NAME = "MYCELIS_LOCAL_ADMIN_USER_ID"
BREAK_GLASS_USERNAME_NAME = "MYCELIS_BREAK_GLASS_USERNAME"
BREAK_GLASS_USER_ID_NAME = "MYCELIS_BREAK_GLASS_USER_ID"

SAMPLE_VALUE = "mycelis-dev-key-change-in-prod"
BREAK_GLASS_SAMPLE_VALUE = "mycelis-break-glass-key-change-in-prod"
LOCAL_ADMIN_SAMPLE_USERNAME = "admin"
LOCAL_ADMIN_SAMPLE_USER_ID = "00000000-0000-0000-0000-000000000000"
BREAK_GLASS_SAMPLE_USERNAME = "recovery-admin"
BREAK_GLASS_SAMPLE_USER_ID = "00000000-0000-0000-0000-000000000001"
ENV_PATH = ROOT_DIR / ".env"
ENV_EXAMPLE_PATH = ROOT_DIR / ".env.example"
ENV_COMPOSE_EXAMPLE_PATH = ROOT_DIR / ".env.compose.example"


def _generate_dev_key(length: int = 40, prefix: str = "mycelis-dev-") -> str:
    alphabet = string.ascii_letters + string.digits
    suffix = "".join(secrets.choice(alphabet) for _ in range(length))
    return f"{prefix}{suffix}"


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


def _sync_auth_example(path: Path) -> None:
    if not path.exists():
        return
    _upsert_env_value(path, KEY_NAME, SAMPLE_VALUE)
    _upsert_env_value(path, BREAK_GLASS_KEY_NAME, BREAK_GLASS_SAMPLE_VALUE)
    _upsert_env_value(path, LOCAL_ADMIN_USERNAME_NAME, LOCAL_ADMIN_SAMPLE_USERNAME)
    _upsert_env_value(path, LOCAL_ADMIN_USER_ID_NAME, LOCAL_ADMIN_SAMPLE_USER_ID)
    _upsert_env_value(path, BREAK_GLASS_USERNAME_NAME, BREAK_GLASS_SAMPLE_USERNAME)
    _upsert_env_value(path, BREAK_GLASS_USER_ID_NAME, BREAK_GLASS_SAMPLE_USER_ID)


def _sync_auth_examples() -> None:
    _sync_auth_example(ENV_EXAMPLE_PATH)


def _inspect_auth_posture(path: Path) -> dict[str, str]:
    return {
        KEY_NAME: _read_env_value(path, KEY_NAME),
        BREAK_GLASS_KEY_NAME: _read_env_value(path, BREAK_GLASS_KEY_NAME),
        LOCAL_ADMIN_USERNAME_NAME: _read_env_value(path, LOCAL_ADMIN_USERNAME_NAME),
        LOCAL_ADMIN_USER_ID_NAME: _read_env_value(path, LOCAL_ADMIN_USER_ID_NAME),
        BREAK_GLASS_USERNAME_NAME: _read_env_value(path, BREAK_GLASS_USERNAME_NAME),
        BREAK_GLASS_USER_ID_NAME: _read_env_value(path, BREAK_GLASS_USER_ID_NAME),
    }


def _auth_posture_warnings(posture: dict[str, str]) -> list[str]:
    warnings: list[str] = []
    break_glass_present = any(
        posture[name].strip()
        for name in (
            BREAK_GLASS_KEY_NAME,
            BREAK_GLASS_USERNAME_NAME,
            BREAK_GLASS_USER_ID_NAME,
        )
    )
    break_glass_complete = all(
        posture[name].strip()
        for name in (
            BREAK_GLASS_KEY_NAME,
            BREAK_GLASS_USERNAME_NAME,
            BREAK_GLASS_USER_ID_NAME,
        )
    )
    if break_glass_present and not break_glass_complete:
        warnings.append(
            "partial break-glass config: set MYCELIS_BREAK_GLASS_API_KEY, "
            "MYCELIS_BREAK_GLASS_USERNAME, and MYCELIS_BREAK_GLASS_USER_ID together"
        )
    if (
        posture[KEY_NAME].strip()
        and posture[BREAK_GLASS_KEY_NAME].strip()
        and posture[KEY_NAME].strip() == posture[BREAK_GLASS_KEY_NAME].strip()
    ):
        warnings.append("break-glass API key matches MYCELIS_API_KEY; use a distinct recovery credential")
    if (
        posture[LOCAL_ADMIN_USERNAME_NAME].strip()
        and posture[BREAK_GLASS_USERNAME_NAME].strip()
        and posture[LOCAL_ADMIN_USERNAME_NAME].strip() == posture[BREAK_GLASS_USERNAME_NAME].strip()
        and posture[LOCAL_ADMIN_USER_ID_NAME].strip()
        and posture[BREAK_GLASS_USER_ID_NAME].strip()
        and posture[LOCAL_ADMIN_USER_ID_NAME].strip() == posture[BREAK_GLASS_USER_ID_NAME].strip()
    ):
        warnings.append("break-glass principal duplicates the primary local admin identity")
    return warnings


def _print_auth_posture(path: Path, label: str) -> None:
    posture = _inspect_auth_posture(path)
    print(f"{label}: {path}")
    print(f"  primary_key: {_mask_secret(posture[KEY_NAME]) if posture[KEY_NAME] else '(missing)'}")
    print(f"  break_glass_key: {_mask_secret(posture[BREAK_GLASS_KEY_NAME]) if posture[BREAK_GLASS_KEY_NAME] else '(missing)'}")
    print(f"  local_admin: {posture[LOCAL_ADMIN_USERNAME_NAME] or 'admin'} / {posture[LOCAL_ADMIN_USER_ID_NAME] or LOCAL_ADMIN_SAMPLE_USER_ID}")
    print(f"  break_glass: {posture[BREAK_GLASS_USERNAME_NAME] or BREAK_GLASS_SAMPLE_USERNAME} / {posture[BREAK_GLASS_USER_ID_NAME] or BREAK_GLASS_SAMPLE_USER_ID}")
    for warning in _auth_posture_warnings(posture):
        print(f"  warning: {warning}")


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

    _sync_auth_examples()

    visible = key if show else _mask_secret(key)
    print(f"{KEY_NAME}: {visible}")
    print(f"Action: {action}")
    print("Next: restart services to apply auth key changes:")
    print("  uv run inv lifecycle.restart")


@task(
    help={
        "rotate": "Rotate and replace the break-glass key in .env even if one already exists.",
        "show": "Print the full key value (default is masked).",
        "value": "Use an explicit key value instead of generating one.",
    }
)
def break_glass_key(_c, rotate=False, show=False, value=""):
    """
    Ensure a local MYCELIS_BREAK_GLASS_API_KEY exists for explicit recovery posture.
    """
    if not ENV_PATH.exists():
        raise SystemExit("Missing .env. Copy .env.example to .env first.")

    explicit_value = value.strip()
    existing = _read_env_value(ENV_PATH, BREAK_GLASS_KEY_NAME)
    action = "kept existing"

    if explicit_value:
        key = explicit_value
        _upsert_env_value(ENV_PATH, BREAK_GLASS_KEY_NAME, key)
        action = "set explicit value"
    elif not existing:
        key = _generate_dev_key(prefix="mycelis-break-glass-")
        _upsert_env_value(ENV_PATH, BREAK_GLASS_KEY_NAME, key)
        action = "generated new break-glass key"
    elif rotate:
        key = _generate_dev_key(prefix="mycelis-break-glass-")
        _upsert_env_value(ENV_PATH, BREAK_GLASS_KEY_NAME, key)
        action = "rotated break-glass key"
    else:
        key = existing

    _sync_auth_examples()

    visible = key if show else _mask_secret(key)
    print(f"{BREAK_GLASS_KEY_NAME}: {visible}")
    print(f"Action: {action}")
    print("Next: restart services to apply break-glass auth key changes:")
    print("  uv run inv lifecycle.restart")


@task(help={"compose": "Inspect .env.compose instead of .env."})
def posture(_c, compose=False):
    """
    Print the current local-admin and break-glass auth posture.
    """
    path = ENV_COMPOSE_EXAMPLE_PATH.parent / ".env.compose" if compose else ENV_PATH
    label = ".env.compose" if compose else ".env"
    if not path.exists():
        raise SystemExit(f"Missing {label}. Copy the matching example file first.")
    _print_auth_posture(path, label)


ns = Collection("auth")
ns.add_task(dev_key, name="dev-key")
ns.add_task(break_glass_key, name="break-glass-key")
ns.add_task(posture, name="posture")
