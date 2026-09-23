from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
COMPOSE_FILE = ROOT / "docker-compose.yml"


def test_compose_shares_web_identity_secrets_between_interface_and_core():
    text = COMPOSE_FILE.read_text(encoding="utf-8")

    session_secret = "MYCELIS_WEB_SESSION_SECRET: ${MYCELIS_WEB_SESSION_SECRET:-${MYCELIS_API_KEY}}"
    forward_secret = (
        "MYCELIS_WEB_IDENTITY_FORWARD_SECRET: "
        "${MYCELIS_WEB_IDENTITY_FORWARD_SECRET:-${MYCELIS_API_KEY}}"
    )

    assert text.count(session_secret) == 2
    assert text.count(forward_secret) == 2


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
