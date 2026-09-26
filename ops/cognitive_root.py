"""Operator-facing reporting for the hosted variant's cognitive root provider.

Two views are kept strictly apart:

- configured (env): what .env / .env.compose / the shell / cognitive.yaml ask
  for (MYCELIS_ROOT_PROVIDER wins over root_provider). This is intent only.
- effective (Core): what a running Core reports on GET /api/v1/cognitive/status
  (root_provider, root_provider_model, and each profile's provider + source).
  Core's router is the single source of truth for routing.

`cognitive.status` and `compose.health` fail when the two disagree, so an
inert root (for example Compose not forwarding MYCELIS_ROOT_PROVIDER) cannot
pass as effective. This module never mutates config and never prints secrets.

PyYAML is not a default ops dependency, so core/config/cognitive.yaml (Core's
own tracked, 2-space-indented config) is read with a narrow line scan.
"""

import json
import os
import re
import urllib.request

from .config import API_HOST, API_PORT, ROOT_DIR

_YAML_TOP_LEVEL_ROOT_PROVIDER = re.compile(r"(?m)^root_provider:\s*(\S+)\s*$")
EXECUTION_PROFILES = ("admin", "architect", "chat", "coder", "creative", "overseer", "sentry")
CORE_STATUS_TIMEOUT_SECONDS = 12.0


def root_provider_env_values():
    """Merge .env, .env.compose, and the process env (process env wins)."""
    values = {}
    for name in (".env", ".env.compose"):
        path = ROOT_DIR / name
        if not path.exists():
            continue
        for raw_line in path.read_text(encoding="utf-8").splitlines():
            line = raw_line.strip()
            if not line or line.startswith("#") or "=" not in raw_line:
                continue
            key, value = raw_line.split("=", 1)
            cleaned = value.strip().strip('"').strip("'")
            if cleaned:
                values[key.strip()] = cleaned
    values.update({key: val for key, val in os.environ.items() if val})
    return values


def _yaml_provider_block(config_text, provider_key):
    block_match = re.search(
        rf"(?m)^  {re.escape(provider_key)}:\s*$\n((?:^(?:    .*)?$\n?)*)",
        config_text,
    )
    if not block_match:
        return {}
    fields = {}
    for field in ("model_id", "endpoint", "enabled"):
        field_match = re.search(rf"(?m)^    {field}:\s*(\S+)\s*$", block_match.group(1))
        if field_match:
            fields[field] = field_match.group(1)
    return fields


def normalize_provider_id(value):
    return (value or "").strip().lower().replace("_", "-")


def configured_root_provider(env_values=None):
    """Configured (env) root intent, or None when nothing asks for a root."""
    values = env_values if env_values is not None else root_provider_env_values()
    root_id = (values.get("MYCELIS_ROOT_PROVIDER", "") or "").strip()

    config_text = ""
    yaml_root = ""
    router_config_path = ROOT_DIR / "core" / "config" / "cognitive.yaml"
    if router_config_path.exists():
        config_text = router_config_path.read_text(encoding="utf-8")
        yaml_match = _YAML_TOP_LEVEL_ROOT_PROVIDER.search(config_text)
        if yaml_match:
            yaml_root = yaml_match.group(1).strip()

    source = "MYCELIS_ROOT_PROVIDER"
    if not root_id:
        root_id, source = yaml_root, "cognitive.yaml root_provider"
    if not root_id:
        return None

    provider_key = normalize_provider_id(root_id)
    provider_fields = _yaml_provider_block(config_text, provider_key) if config_text else {}
    env_prefix = f"MYCELIS_PROVIDER_{root_id.upper().replace('-', '_')}_"
    enabled_raw = values.get(f"{env_prefix}ENABLED") or provider_fields.get("enabled", "false")
    return {
        "provider_id": provider_key,
        "model_id": values.get(f"{env_prefix}MODEL_ID") or provider_fields.get("model_id", ""),
        "endpoint": values.get(f"{env_prefix}ENDPOINT") or provider_fields.get("endpoint", ""),
        "enabled": enabled_raw.strip().lower() in ("1", "true", "yes", "on"),
        "source": source,
    }


def core_status_url(values):
    host = values.get("MYCELIS_API_HOST") or API_HOST
    port = values.get("MYCELIS_API_PORT") or values.get("MYCELIS_COMPOSE_CORE_PORT") or API_PORT
    return f"http://{host}:{port}/api/v1/cognitive/status"


def fetch_core_status(values, timeout=CORE_STATUS_TIMEOUT_SECONDS):
    """Return (payload, error). The API key is sent, never printed."""
    api_key = values.get("MYCELIS_API_KEY", "")
    request = urllib.request.Request(core_status_url(values))
    if api_key:
        request.add_header("Authorization", f"Bearer {api_key}")
    try:
        with urllib.request.urlopen(request, timeout=timeout) as response:
            return json.loads(response.read().decode("utf-8")), ""
    except Exception as exc:  # noqa: BLE001 - reported to the operator, not raised
        return None, type(exc).__name__


def root_disagreements(configured, payload):
    """List configured-vs-effective root disagreements (empty when consistent)."""
    configured_id = normalize_provider_id(configured["provider_id"]) if configured else ""
    effective_id = normalize_provider_id(str(payload.get("root_provider") or ""))
    problems = []
    if configured_id != effective_id:
        problems.append(
            f"configured (env) root_provider={configured_id or '(unset)'} but Core "
            f"effective root_provider={effective_id or '(unset)'}; recreate core so it "
            "receives MYCELIS_ROOT_PROVIDER, or fix the configured value"
        )
    if effective_id:
        profiles = payload.get("profiles") or {}
        for profile in EXECUTION_PROFILES:
            source = str((profiles.get(profile) or {}).get("source", "unbound"))
            if source not in ("root", "override"):
                problems.append(f"profile {profile} source={source} although root_provider={effective_id}")
    return problems


def print_report(env_values=None, fetch=None):
    """Print configured (env) and effective (Core) root views.

    Returns the list of disagreements; an unreachable Core is a warning only.
    """
    values = env_values if env_values is not None else root_provider_env_values()
    configured = configured_root_provider(values)
    if configured:
        state = "enabled" if configured["enabled"] else "DISABLED"
        print(f"  Root Provider (configured, env): {configured['provider_id']} ({state}, via {configured['source']})")
        print(f"    Model (configured, env)      : {configured['model_id'] or '(not set)'}")
    else:
        print("  Root Provider (configured, env): none (MYCELIS_ROOT_PROVIDER/root_provider unset)")

    payload, error = (fetch or fetch_core_status)(values)
    if payload is None:
        print(f"  Root Provider (effective, Core): UNKNOWN - Core status unreachable ({error})")
        print("    [WARN] effective routing not verified; start Core and re-run")
        return []

    effective = str(payload.get("root_provider") or "")
    model = str(payload.get("root_provider_model") or "")
    print(f"  Root Provider (effective, Core): {effective or 'none'}{f' (model {model})' if model else ''}")
    for profile, binding in sorted((payload.get("profiles") or {}).items()):
        provider = binding.get("provider_id") or "-"
        print(f"    {profile:<10} {provider:<20} [{binding.get('source', 'unbound')}]")

    problems = root_disagreements(configured, payload)
    for problem in problems:
        print(f"  [FAIL] {problem}")
    return problems
