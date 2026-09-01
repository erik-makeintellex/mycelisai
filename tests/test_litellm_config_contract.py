from __future__ import annotations

from pathlib import Path

import yaml


ROOT = Path(__file__).resolve().parents[1]
CORE_CONFIG = ROOT / "core" / "config" / "cognitive.yaml"
CHART_CONFIG = ROOT / "charts" / "mycelis-core" / "config" / "cognitive.yaml"
CHART_VALUES = ROOT / "charts" / "mycelis-core" / "values.yaml"
CHART_DEPLOYMENT = ROOT / "charts" / "mycelis-core" / "templates" / "deployment.yaml"
COMPOSE = ROOT / "docker-compose.yml"
ENV_EXAMPLE = ROOT / ".env.example"


def _provider(path: Path) -> dict[str, object]:
    config = yaml.safe_load(path.read_text(encoding="utf-8"))
    return config["providers"]["litellm"]


def test_litellm_uses_disabled_openai_compatible_provider_contract():
    core_provider = _provider(CORE_CONFIG)
    chart_provider = _provider(CHART_CONFIG)

    for provider in (core_provider, chart_provider):
        assert provider["type"] == "openai_compatible"
        assert provider["model_id"] == "mycelis-default"
        assert provider["api_key"] == ""
        assert provider["api_key_env"] == "LITELLM_PROXY_API_KEY"
        assert provider["enabled"] is False
        assert provider["location"] == "remote"
        assert provider["data_boundary"] == "leaves_org"
        assert provider["usage_policy"] == "require_approval"

    assert core_provider["endpoint"] == "http://127.0.0.1:4000/v1"
    assert chart_provider["endpoint"] == "http://litellm:4000/v1"


def test_litellm_example_env_is_secret_safe_and_disabled():
    env_example = ENV_EXAMPLE.read_text(encoding="utf-8")

    assert "LITELLM_PROXY_API_KEY=\n" in env_example
    assert "LITELLM_" + "MASTER_KEY" not in env_example
    assert "MYCELIS_PROVIDER_LITELLM_ENABLED=false" in env_example
    assert "MYCELIS_PROVIDER_LITELLM_ENDPOINT=http://127.0.0.1:4000/v1" in env_example
    assert "MYCELIS_PROVIDER_LITELLM_MODEL_ID=mycelis-default" in env_example


def test_chart_requires_external_litellm_endpoint_and_existing_secret_to_enable():
    values = yaml.safe_load(CHART_VALUES.read_text(encoding="utf-8"))
    deployment = CHART_DEPLOYMENT.read_text(encoding="utf-8")
    litellm = values["ai"]["litellm"]

    assert litellm == {
        "enabled": False,
        "endpoint": "",
        "modelId": "mycelis-default",
        "existingSecret": "",
        "secretKey": "litellm-proxy-api-key",
    }

    required_snippets = [
        "if .Values.ai.litellm.enabled",
        "MYCELIS_PROVIDER_LITELLM_ENDPOINT",
        "ai.litellm.endpoint is required when LiteLLM is enabled",
        "MYCELIS_PROVIDER_LITELLM_MODEL_ID",
        "MYCELIS_PROVIDER_LITELLM_API_KEY_ENV",
        "MYCELIS_PROVIDER_LITELLM_ENABLED",
        "name: LITELLM_PROXY_API_KEY",
        "ai.litellm.existingSecret is required when LiteLLM is enabled",
        "ai.litellm.secretKey is required when LiteLLM is enabled",
    ]
    missing = [snippet for snippet in required_snippets if snippet not in deployment]
    assert not missing, "LiteLLM chart activation contract is incomplete:\n" + "\n".join(missing)


def test_first_litellm_conformance_slice_does_not_install_gateway_service():
    compose = yaml.safe_load(COMPOSE.read_text(encoding="utf-8"))

    assert "litellm" not in compose["services"]
    assert not list((ROOT / "charts" / "mycelis-core" / "templates").glob("*litellm*"))


def test_core_facing_contract_excludes_proxy_administration_key():
    administration_key_env = "LITELLM_" + "MASTER_KEY"

    for path in (CORE_CONFIG, CHART_CONFIG, CHART_VALUES, CHART_DEPLOYMENT, ENV_EXAMPLE):
        assert administration_key_env not in path.read_text(encoding="utf-8")
