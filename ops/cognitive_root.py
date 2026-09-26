"""Read-only reporting for the hosted variant's cognitive root provider.

Mirrors core/internal/cognitive's MYCELIS_ROOT_PROVIDER/root_provider
precedence (env overrides cognitive.yaml) for operator-facing status output
(ops/cognitive.py status, cognitive.status). It never mutates config and
never duplicates Go's real per-profile override resolution — Core's router is
the single source of truth for what a profile actually resolves to.

PyYAML is only a dependency of the optional cognitive/ workspace member (the
local vLLM/Diffusers helper), not of the default ops/dev toolchain, so this
helper must not require `import yaml` to succeed. core/config/cognitive.yaml
is Core's own tracked, 2-space-indented config, not arbitrary user YAML, so a
targeted line scan for a handful of known scalar fields is sufficient here.
"""

import os
import re

from .config import ROOT_DIR

_YAML_TOP_LEVEL_ROOT_PROVIDER = re.compile(r"(?m)^root_provider:\s*(\S+)\s*$")


def root_provider_env_values():
    """Merge .env, .env.compose, and the process env for root-provider lookups
    (the shell/process env wins over the tracked files, matching Core's own
    env-override precedence)."""
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
    """Extract the (2-space-indented) provider block for provider_key out of
    core/config/cognitive.yaml's providers: map using a narrow line scan."""
    block_match = re.search(
        rf"(?m)^  {re.escape(provider_key)}:\s*$\n((?:^(?:    .*)?$\n?)*)",
        config_text,
    )
    if not block_match:
        return {}
    block = block_match.group(1)
    fields = {}
    for field in ("model_id", "endpoint", "enabled"):
        field_match = re.search(rf"(?m)^    {field}:\s*(\S+)\s*$", block)
        if field_match:
            fields[field] = field_match.group(1)
    return fields


def effective_root_provider(env_values=None):
    """Report the effective root provider/model for the hosted variant's root
    model posture. An explicit MYCELIS_ROOT_PROVIDER (shell, .env, or
    .env.compose) overrides core/config/cognitive.yaml's root_provider.
    Returns None when no root provider is configured anywhere (today's
    default, unchanged behavior)."""
    values = env_values if env_values is not None else root_provider_env_values()
    root_id = (values.get("MYCELIS_ROOT_PROVIDER", "") or "").strip()

    yaml_root = ""
    config_text = ""
    router_config_path = ROOT_DIR / "core" / "config" / "cognitive.yaml"
    if router_config_path.exists():
        config_text = router_config_path.read_text(encoding="utf-8")
        yaml_match = _YAML_TOP_LEVEL_ROOT_PROVIDER.search(config_text)
        if yaml_match:
            yaml_root = yaml_match.group(1).strip()

    source = "MYCELIS_ROOT_PROVIDER"
    if not root_id:
        root_id = yaml_root
        source = "cognitive.yaml root_provider"
    if not root_id:
        return None

    provider_key = root_id.lower().replace("_", "-")
    provider_fields = _yaml_provider_block(config_text, provider_key) if config_text else {}
    env_prefix = f"MYCELIS_PROVIDER_{root_id.upper().replace('-', '_')}_"

    model_id = values.get(f"{env_prefix}MODEL_ID") or provider_fields.get("model_id", "")
    endpoint = values.get(f"{env_prefix}ENDPOINT") or provider_fields.get("endpoint", "")
    enabled_raw = values.get(f"{env_prefix}ENABLED")
    if enabled_raw is not None:
        enabled = enabled_raw.strip().lower() in ("1", "true", "yes", "on")
    else:
        enabled = provider_fields.get("enabled", "false").strip().lower() in ("1", "true", "yes", "on")

    return {
        "provider_id": provider_key,
        "model_id": model_id,
        "endpoint": endpoint,
        "enabled": enabled,
        "source": source,
    }
