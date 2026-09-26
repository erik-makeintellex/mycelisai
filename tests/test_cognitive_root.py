from __future__ import annotations

from ops import cognitive_root


def test_configured_root_provider_unset_returns_none(monkeypatch, tmp_path):
    monkeypatch.setattr(cognitive_root, "ROOT_DIR", tmp_path)
    assert cognitive_root.configured_root_provider(env_values={}) is None


def test_configured_root_provider_reads_cognitive_yaml_default(monkeypatch, tmp_path):
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

    root = cognitive_root.configured_root_provider(env_values={})

    assert root == {
        "provider_id": "vllm",
        "model_id": "Qwen/Qwen2.5-Coder-14B-Instruct-AWQ",
        "endpoint": "http://host.docker.internal:8000/v1",
        "enabled": True,
        "source": "cognitive.yaml root_provider",
    }


def test_configured_root_provider_env_override_wins_over_yaml(monkeypatch, tmp_path):
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

    root = cognitive_root.configured_root_provider(
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


def test_configured_root_provider_disabled_reports_disabled(monkeypatch, tmp_path):
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

    root = cognitive_root.configured_root_provider(env_values={})

    assert root["enabled"] is False


def test_configured_root_provider_missing_config_file_returns_none(monkeypatch, tmp_path):
    monkeypatch.setattr(cognitive_root, "ROOT_DIR", tmp_path)
    root = cognitive_root.configured_root_provider(
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


ROOT_PAYLOAD = {
    "root_provider": "vllm",
    "root_provider_model": "Qwen/Qwen2.5-Coder-14B-Instruct-AWQ",
    "profiles": {
        name: {"provider_id": "vllm", "source": "root"}
        for name in cognitive_root.EXECUTION_PROFILES
    },
}


def test_root_disagreements_consistent_root_passes(monkeypatch, tmp_path):
    monkeypatch.setattr(cognitive_root, "ROOT_DIR", tmp_path)
    configured = cognitive_root.configured_root_provider({"MYCELIS_ROOT_PROVIDER": "vllm"})
    payload = {**ROOT_PAYLOAD, "profiles": {**ROOT_PAYLOAD["profiles"], "coder": {"provider_id": "ollama", "source": "override"}}}
    assert cognitive_root.root_disagreements(configured, payload) == []


def test_root_disagreements_flags_inert_root(monkeypatch, tmp_path):
    # The S6 defect: env says vllm, Core never received it.
    monkeypatch.setattr(cognitive_root, "ROOT_DIR", tmp_path)
    configured = cognitive_root.configured_root_provider({"MYCELIS_ROOT_PROVIDER": "vllm"})
    problems = cognitive_root.root_disagreements(configured, {"text": {"status": "online"}})
    assert len(problems) == 1
    assert "configured (env) root_provider=vllm" in problems[0]
    assert "effective root_provider=(unset)" in problems[0]


def test_root_disagreements_flags_unexpected_effective_root_and_default_profiles(monkeypatch, tmp_path):
    monkeypatch.setattr(cognitive_root, "ROOT_DIR", tmp_path)
    payload = {**ROOT_PAYLOAD, "profiles": {**ROOT_PAYLOAD["profiles"], "chat": {"provider_id": "ollama", "source": "default"}}}
    problems = cognitive_root.root_disagreements(None, payload)
    assert any("configured (env) root_provider=(unset)" in p for p in problems)
    assert any("profile chat source=default" in p for p in problems)


def test_root_disagreements_unset_everywhere_passes(monkeypatch, tmp_path):
    monkeypatch.setattr(cognitive_root, "ROOT_DIR", tmp_path)
    assert cognitive_root.root_disagreements(None, {"profiles": {"chat": {"source": "default"}}}) == []


def test_print_report_labels_configured_and_effective(monkeypatch, tmp_path, capsys):
    monkeypatch.setattr(cognitive_root, "ROOT_DIR", tmp_path)
    problems = cognitive_root.print_report(
        {"MYCELIS_ROOT_PROVIDER": "vllm", "MYCELIS_API_KEY": "never-printed"},
        fetch=lambda values: (ROOT_PAYLOAD, ""),
    )
    out = capsys.readouterr().out
    assert problems == []
    assert "Root Provider (configured, env): vllm" in out
    assert "Root Provider (effective, Core): vllm (model Qwen/Qwen2.5-Coder-14B-Instruct-AWQ)" in out
    assert "[root]" in out
    assert "never-printed" not in out


def test_print_report_unreachable_core_warns_without_failing(monkeypatch, tmp_path, capsys):
    monkeypatch.setattr(cognitive_root, "ROOT_DIR", tmp_path)
    problems = cognitive_root.print_report({"MYCELIS_ROOT_PROVIDER": "vllm"}, fetch=lambda values: (None, "URLError"))
    out = capsys.readouterr().out
    assert problems == []
    assert "UNKNOWN - Core status unreachable (URLError)" in out
    assert "[WARN]" in out


def test_core_status_url_prefers_compose_core_port(monkeypatch):
    monkeypatch.delenv("MYCELIS_API_PORT", raising=False)
    url = cognitive_root.core_status_url({"MYCELIS_COMPOSE_CORE_PORT": "9999"})
    assert url.endswith(":9999/api/v1/cognitive/status")


def test_compose_health_fails_when_effective_root_differs(monkeypatch, tmp_path):
    from ops import compose_probe

    monkeypatch.setattr(cognitive_root, "ROOT_DIR", tmp_path)
    failures: list[str] = []
    body = '{"text":{"status":"online"},"media":{"status":"offline"}}'
    compose_probe._append_cognitive_health_failures("url", body, failures, {"MYCELIS_ROOT_PROVIDER": "vllm"})
    assert any(f.startswith("Root Provider: configured (env) root_provider=vllm") for f in failures)

    failures = []
    import json

    compose_probe._append_cognitive_health_failures(
        "url", json.dumps({**ROOT_PAYLOAD, "text": {"status": "online"}}), failures, {"MYCELIS_ROOT_PROVIDER": "vllm"}
    )
    assert failures == []
