from __future__ import annotations

import hashlib
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
SESSION_SECRET_NAME = "MYCELIS_WEB_SESSION_SECRET"
FORWARD_SECRET_NAME = "MYCELIS_WEB_IDENTITY_FORWARD_SECRET"
LOCAL_PASSWORD_NAME = "MYCELIS_LOCAL_ADMIN_PASSWORD"
LOCAL_PASSWORD_SHA256_NAME = "MYCELIS_LOCAL_ADMIN_PASSWORD_SHA256"
WEB_SECRET_NAMES = (SESSION_SECRET_NAME, FORWARD_SECRET_NAME)
MIN_SECRET_BYTES = 32

SAMPLE_VALUE = "mycelis-dev-key-change-in-prod"
BREAK_GLASS_SAMPLE_VALUE = "mycelis-break-glass-key-change-in-prod"
LOCAL_ADMIN_SAMPLE_USERNAME = "admin"
LOCAL_ADMIN_SAMPLE_USER_ID = "00000000-0000-0000-0000-000000000000"
BREAK_GLASS_SAMPLE_USERNAME = "recovery-admin"
BREAK_GLASS_SAMPLE_USER_ID = "00000000-0000-0000-0000-000000000001"
WEB_SECRET_SAMPLES = {
    SESSION_SECRET_NAME: "mycelis-web-session-secret-change-in-prod",
    FORWARD_SECRET_NAME: "mycelis-web-identity-forward-secret-change-in-prod",
}
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
    for name, sample in WEB_SECRET_SAMPLES.items():
        _upsert_env_value(path, name, sample)


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
        SESSION_SECRET_NAME: _read_env_value(path, SESSION_SECRET_NAME),
        FORWARD_SECRET_NAME: _read_env_value(path, FORWARD_SECRET_NAME),
        LOCAL_PASSWORD_NAME: _read_env_value(path, LOCAL_PASSWORD_NAME),
        LOCAL_PASSWORD_SHA256_NAME: _read_env_value(path, LOCAL_PASSWORD_SHA256_NAME),
    }


def _web_secret_problem(name: str, values: dict[str, str]) -> str:
    """Name what is wrong with a web secret without echoing any value."""
    value = values.get(name, "").strip()
    if not value:
        return f"{name} is missing; run uv run inv auth.dev-key"
    if len(value.encode("utf-8")) < MIN_SECRET_BYTES:
        return f"{name} must be at least {MIN_SECRET_BYTES} bytes; run uv run inv auth.dev-key"
    others = [KEY_NAME, BREAK_GLASS_KEY_NAME] + [other for other in WEB_SECRET_NAMES if other != name]
    reused = [other for other in others if values.get(other, "").strip() == value]
    if reused:
        return f"{name} must differ from {', '.join(reused)}; run uv run inv auth.dev-key"
    return ""


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
    for name in WEB_SECRET_NAMES:
        problem = _web_secret_problem(name, posture)
        if problem:
            warnings.append(problem)
    if not posture[LOCAL_PASSWORD_NAME].strip() and not posture[LOCAL_PASSWORD_SHA256_NAME].strip():
        warnings.append(
            f"set {LOCAL_PASSWORD_SHA256_NAME} (preferred) or {LOCAL_PASSWORD_NAME}; "
            f"{KEY_NAME} is no longer accepted as the local admin password"
        )
    elif posture[LOCAL_PASSWORD_NAME].strip() and posture[LOCAL_PASSWORD_NAME].strip() in {
        posture[name].strip() for name in (KEY_NAME, BREAK_GLASS_KEY_NAME, *WEB_SECRET_NAMES) if posture[name].strip()
    }:
        warnings.append(f"{LOCAL_PASSWORD_NAME} must not reuse an API key or web secret")
    plain, digest = posture[LOCAL_PASSWORD_NAME].strip(), posture[LOCAL_PASSWORD_SHA256_NAME].strip().lower()
    if plain and digest and _sha256_hex(plain) != digest:
        warnings.append(f"{LOCAL_PASSWORD_NAME} does not match {LOCAL_PASSWORD_SHA256_NAME}; sign-in uses the hash. Run uv run inv auth.dev-key --admin-password=sync")
    return warnings


def _sha256_hex(value: str) -> str:
    return hashlib.sha256(value.encode("utf-8")).hexdigest()


def _ensure_web_secrets(path: Path) -> list[tuple[str, str]]:
    """Generate missing, short or reused web secrets. Returns (name, action) pairs, never values."""
    results: list[tuple[str, str]] = []
    for name in WEB_SECRET_NAMES:
        values = _inspect_auth_posture(path)
        if _web_secret_problem(name, values):
            _upsert_env_value(path, name, secrets.token_urlsafe(48))
            results.append((name, "generated" if not values[name].strip() else "replaced (was missing, short or reused)"))
        else:
            results.append((name, "kept existing"))
    return results


def _print_auth_posture(path: Path, label: str) -> None:
    posture = _inspect_auth_posture(path)
    print(f"{label}: {path}")
    print(f"  primary_key: {_mask_secret(posture[KEY_NAME]) if posture[KEY_NAME] else '(missing)'}")
    print(f"  break_glass_key: {_mask_secret(posture[BREAK_GLASS_KEY_NAME]) if posture[BREAK_GLASS_KEY_NAME] else '(missing)'}")
    print(f"  local_admin: {posture[LOCAL_ADMIN_USERNAME_NAME] or 'admin'} / {posture[LOCAL_ADMIN_USER_ID_NAME] or LOCAL_ADMIN_SAMPLE_USER_ID}")
    print(f"  break_glass: {posture[BREAK_GLASS_USERNAME_NAME] or BREAK_GLASS_SAMPLE_USERNAME} / {posture[BREAK_GLASS_USER_ID_NAME] or BREAK_GLASS_SAMPLE_USER_ID}")
    for name in (*WEB_SECRET_NAMES, LOCAL_PASSWORD_SHA256_NAME, LOCAL_PASSWORD_NAME):
        print(f"  {name}: {'set' if posture[name].strip() else '(missing)'}")
    for warning in _auth_posture_warnings(posture):
        print(f"  warning: {warning}")


@task(
    help={
        "rotate": "Rotate and replace the key in .env even if one already exists.",
        "show": "Print the full key value (default is masked).",
        "value": "Use an explicit key value instead of generating one.",
        "admin_password": "Local admin for e2e-from-.env: sync | generate | <password> (generate/<password> change your sign-in).",
    }
)
def dev_key(_c, rotate=False, show=False, value="", admin_password=""):
    """
    Ensure MYCELIS_API_KEY and distinct web session/forward secrets exist in .env.
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

    web_secret_actions = _ensure_web_secrets(ENV_PATH)
    _sync_auth_examples()

    visible = key if show else _mask_secret(key)
    print(f"{KEY_NAME}: {visible}")
    print(f"Action: {action}")
    for name, web_action in web_secret_actions:
        shown = f" = {_read_env_value(ENV_PATH, name)}" if show else ""
        print(f"{name}: {web_action}{shown}")
    posture = _inspect_auth_posture(ENV_PATH)
    if not posture[LOCAL_PASSWORD_NAME].strip() and not posture[LOCAL_PASSWORD_SHA256_NAME].strip():
        print(f"Local sign-in stays disabled until {LOCAL_PASSWORD_SHA256_NAME} (preferred) or {LOCAL_PASSWORD_NAME} is set in .env.")
    if admin_password.strip():
        _ensure_local_admin(admin_password.strip(), show)
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
    Print the current local-admin and break-glass auth posture. Use before auth.dev-key/auth.break-glass-key to check what is already set.
    """
    path = ENV_COMPOSE_EXAMPLE_PATH.parent / ".env.compose" if compose else ENV_PATH
    label = ".env.compose" if compose else ".env"
    if not path.exists():
        raise SystemExit(f"Missing {label}. Copy the matching example file first.")
    _print_auth_posture(path, label)


def _ensure_local_admin(mode: str, show: bool) -> None:
    """Keep the local admin plaintext and its SHA-256 consistent in .env so e2e signs in from
    .env. mode: "sync" (re-derive the hash from an existing plaintext), "generate" (new
    password), or an explicit password. generate/explicit change the owner's sign-in."""
    posture = _inspect_auth_posture(ENV_PATH)
    reserved = {posture[n].strip() for n in (KEY_NAME, BREAK_GLASS_KEY_NAME, *WEB_SECRET_NAMES) if posture[n].strip()}
    changing = mode not in ("", "sync")
    plain = (_generate_dev_key(prefix="mycelis-admin-") if mode == "generate" else mode) if changing else posture[LOCAL_PASSWORD_NAME].strip()
    if not plain:
        print(f"{LOCAL_PASSWORD_NAME} is empty ({LOCAL_PASSWORD_SHA256_NAME} is {'set' if posture[LOCAL_PASSWORD_SHA256_NAME].strip() else 'missing'}); "
              "e2e cannot sign in from .env. Use --admin-password=generate or --admin-password=<password>. Local admin unchanged.")
        return
    if plain in reserved:
        raise SystemExit(f"{LOCAL_PASSWORD_NAME} must not reuse an API key or web secret.")
    action = "set" if changing else ("kept" if _sha256_hex(plain) == posture[LOCAL_PASSWORD_SHA256_NAME].strip().lower() else "hash re-derived from plaintext")
    _upsert_env_value(ENV_PATH, LOCAL_PASSWORD_NAME, plain)
    _upsert_env_value(ENV_PATH, LOCAL_PASSWORD_SHA256_NAME, _sha256_hex(plain))
    print(f"{LOCAL_PASSWORD_NAME}: {plain if show else '(hidden; --show to print)'}  action: {action}")
    print("Then: uv run inv compose.up (applies the hash); uv run inv interface.e2e reads the password from .env.")


ns = Collection("auth")
ns.add_task(dev_key, name="dev-key")
ns.add_task(break_glass_key, name="break-glass-key")
ns.add_task(posture, name="posture")
