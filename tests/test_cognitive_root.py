from __future__ import annotations

from ops import cognitive_root


def test_effective_root_provider_unset_returns_none(monkeypatch, tmp_path):
    monkeypatch.setattr(cognitive_root, "ROOT_DIR", tmp_path)
    assert cognitive_root.effective_root_provider(env_values={}) is None


def test_effective_root_provider_reads_cognitive_yaml_default(monkeypatch, tmp_path):
    monkeypatch.setattr(cognitive_root, "ROOT_DIR", tmp_path)
    config_dir = tmp_path / "core" / "config"
    config_dir.mkdir(parents=True)
    (config_dir / "cognitive.yaml").write_text(
        """
root_provider: vllm
providers:
  vllm:
    type: openai_compatible
    endpoint: http://host.docker.internal:8000/v1
    model_id: Qwen/Qwen2.5-Coder-14B-Instruct-AWQ
    enabled: true
""",
        encoding="utf-8",
    )

    root = cognitive_root.effective_root_provider(env_values={})

    assert root == {
        "provider_id": "vllm",
        "model_id": "Qwen/Qwen2.5-Coder-14B-Instruct-AWQ",
        "endpoint": "http://host.docker.internal:8000/v1",
        "enabled": True,
        "source": "cognitive.yaml root_provider",
    }


def test_effective_root_provider_env_override_wins_over_yaml(monkeypatch, tmp_path):
    monkeypatch.setattr(cognitive_root, "ROOT_DIR", tmp_path)
    config_dir = tmp_path / "core" / "config"
    config_dir.mkdir(parents=True)
    (config_dir / "cognitive.yaml").write_text(
        """
root_provider: ollama
providers:
  ollama:
    type: ollama
    endpoint: http://127.0.0.1:11434/v1
    model_id: qwen2.5-coder:7b
    enabled: true
""",
        encoding="utf-8",
    )

    root = cognitive_root.effective_root_provider(
        env_values={
            "MYCELIS_ROOT_PROVIDER": "vllm",
            "MYCELIS_PROVIDER_VLLM_ENDPOINT": "http://host.docker.internal:8000/v1",
            "MYCELIS_PROVIDER_VLLM_MODEL_ID": "Qwen/Qwen2.5-Coder-14B-Instruct-AWQ",
            "MYCELIS_PROVIDER_VLLM_ENABLED": "true",
        }
    )

    assert root["provider_id"] == "vllm"
    assert root["model_id"] == "Qwen/Qwen2.5-Coder-14B-Instruct-AWQ"
    assert root["endpoint"] == "http://host.docker.internal:8000/v1"
    assert root["enabled"] is True
    assert root["source"] == "MYCELIS_ROOT_PROVIDER"


def test_effective_root_provider_disabled_reports_disabled(monkeypatch, tmp_path):
    monkeypatch.setattr(cognitive_root, "ROOT_DIR", tmp_path)
    config_dir = tmp_path / "core" / "config"
    config_dir.mkdir(parents=True)
    (config_dir / "cognitive.yaml").write_text(
        """
root_provider: vllm
providers:
  vllm:
    type: openai_compatible
    endpoint: http://host.docker.internal:8000/v1
    model_id: Qwen/Qwen2.5-Coder-14B-Instruct-AWQ
    enabled: false
""",
        encoding="utf-8",
    )

    root = cognitive_root.effective_root_provider(env_values={})

    assert root["enabled"] is False


def test_effective_root_provider_missing_config_file_returns_none(monkeypatch, tmp_path):
    monkeypatch.setattr(cognitive_root, "ROOT_DIR", tmp_path)
    root = cognitive_root.effective_root_provider(
        env_values={"MYCELIS_ROOT_PROVIDER": "vllm"}
    )
    # No cognitive.yaml present: the provider id resolves, but nothing is
    # known about it beyond what env vars supplied (none here), so the
    # reported model/endpoint are blank rather than fabricated.
    assert root["provider_id"] == "vllm"
    assert root["model_id"] == ""
    assert root["endpoint"] == ""
    assert root["enabled"] is False


def test_root_provider_env_values_shell_wins_over_dotenv_files(monkeypatch, tmp_path):
    monkeypatch.setattr(cognitive_root, "ROOT_DIR", tmp_path)
    (tmp_path / ".env.compose").write_text(
        "MYCELIS_ROOT_PROVIDER=ollama\n", encoding="utf-8"
    )
    monkeypatch.setenv("MYCELIS_ROOT_PROVIDER", "vllm")

    values = cognitive_root.root_provider_env_values()

    assert values["MYCELIS_ROOT_PROVIDER"] == "vllm"
