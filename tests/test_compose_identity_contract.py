from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
COMPOSE_FILE = ROOT / "docker-compose.yml"


def _compose_service_block(text: str, name: str) -> str:
    import re

    match = re.search(rf"^  {name}:\n(.*?)(?=^  [a-z-]+:\n|^[a-z]|\Z)", text, re.MULTILINE | re.DOTALL)
    assert match, f"service {name} missing"
    return match.group(1)


def test_compose_web_secrets_are_required_and_never_fall_back_to_the_api_key():
    text = COMPOSE_FILE.read_text(encoding="utf-8")
    assert ":-${MYCELIS_API_KEY}" not in text
    core = _compose_service_block(text, "core")
    interface = _compose_service_block(text, "interface")
    forward = (
        "MYCELIS_WEB_IDENTITY_FORWARD_SECRET: ${MYCELIS_WEB_IDENTITY_FORWARD_SECRET:"
        "?set MYCELIS_WEB_IDENTITY_FORWARD_SECRET in .env (uv run inv auth.dev-key)}"
    )
    session = (
        "MYCELIS_WEB_SESSION_SECRET: ${MYCELIS_WEB_SESSION_SECRET:"
        "?set MYCELIS_WEB_SESSION_SECRET in .env (uv run inv auth.dev-key)}"
    )
    assert forward in core
    assert forward in interface
    assert session in interface
    # Core no longer reads the browser session secret.
    assert "MYCELIS_WEB_SESSION_SECRET" not in core
    for key in ("MYCELIS_LOCAL_ADMIN_USERNAME", "MYCELIS_LOCAL_ADMIN_PASSWORD", "MYCELIS_LOCAL_ADMIN_PASSWORD_SHA256"):
        assert f"      {key}: ${{{key}:-}}\n" in interface
    assert "MYCELIS_LOCAL_ADMIN_PASSWORD" not in core


def test_compose_profile_overrides_match_core_environment_contract():
    from pathlib import Path
    import re

    root = Path(__file__).resolve().parents[1]
    text = (root / "docker-compose.yml").read_text()
    overrides = re.findall(r"^      (MYCELIS_PROFILE_\w+):", text, re.MULTILINE)
    assert len(overrides) == 7
    for key, default in (
        ("MYCELIS_PROVIDER_VLLM_ENABLED", "false"),
        ("MYCELIS_PROVIDER_VLLM_ENDPOINT", "http://host.docker.internal:8000/v1"),
        ("MYCELIS_PROVIDER_VLLM_MODEL_ID", "qwen2.5-coder"),
        ("MYCELIS_TEXT_ENGINE_API_KEY", ""),
        ("MYCELIS_PROVIDER_VLLM_MAX_OUTPUT_TOKENS", "1024"),
        ("MYCELIS_PROVIDER_VLLM_TOKEN_BUDGET_PROFILE", "standard"),
    ):
        assert f"{key}: ${{{key}:-{default}}}" in text
    assert all(key.endswith("_PROVIDER") for key in overrides)
    core = (root / "core/internal/cognitive/env_overrides.go").read_text()
    assert 'strings.TrimSuffix(rawField, "_PROVIDER")' in core


def test_google_sso_is_runtime_interface_configuration_only():
    import yaml

    services = yaml.safe_load(COMPOSE_FILE.read_text())["services"]
    keys = (
        "MYCELIS_AUTH_GOOGLE_CLIENT_ID", "MYCELIS_AUTH_GOOGLE_CLIENT_SECRET",
        "MYCELIS_AUTH_GOOGLE_REDIRECT_URI", "MYCELIS_AUTH_GOOGLE_HOSTED_DOMAIN",
        "MYCELIS_AUTH_ALLOWED_DOMAINS", "MYCELIS_AUTH_ADMIN_EMAILS",
        "MYCELIS_PUBLIC_ORIGIN", "MYCELIS_WEB_COOKIE_SECURE",
    )
    for key in keys:
        assert services["interface"]["environment"][key] == f"${{{key}:-}}"
        for name, service in services.items():
            if name != "interface":
                assert key not in service.get("environment", {})
            build = service.get("build", {})
            if isinstance(build, dict):
                assert key not in build.get("args", {})


def test_google_secret_store_survives_compose_runtime_environment(tmp_path):
    from ops.compose_env import load_compose_env, compose_runtime_env

    (tmp_path / ".env").write_text(
        "MYCELIS_AUTH_GOOGLE_CLIENT_SECRET=fixture-current\n"
        "MYCELIS_AUTH_GOOGLE_CLIENT_ID=fixture-client\n"
        "MYCELIS_AUTH_GOOGLE_REDIRECT_URI=http://127.0.0.1:3000/auth/google/callback\n"
    )
    topology = tmp_path / ".env.compose"
    topology.write_text("MYCELIS_AUTH_GOOGLE_CLIENT_SECRET=fixture-stale\n")
    values = load_compose_env(topology, lambda: None)
    runtime = compose_runtime_env(
        values, lambda: "local", lambda: False, lambda v: v,
        tmp_path, str, environ={},
    )
    assert runtime["MYCELIS_AUTH_GOOGLE_CLIENT_SECRET"] == "fixture-current"
    assert runtime["MYCELIS_AUTH_GOOGLE_CLIENT_ID"] == "fixture-client"
    assert runtime["MYCELIS_AUTH_GOOGLE_REDIRECT_URI"].endswith("/auth/google/callback")
