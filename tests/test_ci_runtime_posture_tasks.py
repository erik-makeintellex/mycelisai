import os

import pytest

from ops import ci
from tests.ci_task_support import FakeContext


def test_runtime_posture_check_fails_when_no_explicit_endpoint_is_configured(monkeypatch, tmp_path):
    headroom_calls: list[dict[str, object]] = []
    monkeypatch.setattr(
        ci.cache_tasks,
        "ensure_disk_headroom",
        lambda **kwargs: headroom_calls.append(kwargs),
    )
    monkeypatch.setattr(ci, "ROOT_DIR", tmp_path)
    for env_name in (
        "MYCELIS_COMPOSE_OLLAMA_HOST",
        "MYCELIS_K8S_TEXT_ENDPOINT",
        "MYCELIS_K8S_MEDIA_ENDPOINT",
        "MYCELIS_PROVIDER_LOCAL_OLLAMA_DEV_ENDPOINT",
    ):
        monkeypatch.delenv(env_name, raising=False)

    with pytest.raises(SystemExit) as excinfo:
        ci._runtime_posture_check(FakeContext({}))

    assert headroom_calls == [{"min_free_gb": 12, "reason": "release preflight posture"}]
    assert "no explicit AI endpoint configured" in str(excinfo.value)


def test_runtime_posture_check_reads_compose_env_file_when_process_env_is_empty(monkeypatch, tmp_path):
    monkeypatch.setattr(ci.cache_tasks, "ensure_disk_headroom", lambda **_kwargs: None)
    monkeypatch.setattr(ci, "ROOT_DIR", tmp_path)
    for env_name in (
        "MYCELIS_COMPOSE_OLLAMA_HOST",
        "MYCELIS_K8S_TEXT_ENDPOINT",
        "MYCELIS_K8S_MEDIA_ENDPOINT",
        "MYCELIS_PROVIDER_LOCAL_OLLAMA_DEV_ENDPOINT",
    ):
        monkeypatch.delenv(env_name, raising=False)

    (tmp_path / ".env.compose").write_text(
        "MYCELIS_COMPOSE_OLLAMA_HOST=http://192.168.50.156:11434\n",
        encoding="utf-8",
    )

    probe_urls: list[str] = []

    def fake_probe(url: str, timeout: float = 3.0):
        probe_urls.append(url)
        return 200, "{}"

    monkeypatch.setattr(ci, "_probe_http_endpoint", fake_probe)

    ci._runtime_posture_check(FakeContext({}))

    assert probe_urls == ["http://192.168.50.156:11434/api/tags"]


def test_runtime_posture_check_includes_provider_specific_endpoint_from_env_file(monkeypatch, tmp_path):
    monkeypatch.setattr(ci.cache_tasks, "ensure_disk_headroom", lambda **_kwargs: None)
    monkeypatch.setattr(ci, "ROOT_DIR", tmp_path)
    for env_name in (
        "MYCELIS_COMPOSE_OLLAMA_HOST",
        "MYCELIS_K8S_TEXT_ENDPOINT",
        "MYCELIS_K8S_MEDIA_ENDPOINT",
        "MYCELIS_PROVIDER_LOCAL_OLLAMA_DEV_ENDPOINT",
        "MYCELIS_PROVIDER_TEAMLEAD_OLLAMA_ENDPOINT",
    ):
        monkeypatch.delenv(env_name, raising=False)

    (tmp_path / ".env").write_text(
        "MYCELIS_PROVIDER_TEAMLEAD_OLLAMA_ENDPOINT=http://192.168.50.157:11434/v1\n",
        encoding="utf-8",
    )

    probe_urls: list[str] = []

    def fake_probe(url: str, timeout: float = 3.0):
        probe_urls.append(url)
        return 401, "unauthorized"

    monkeypatch.setattr(ci, "_probe_http_endpoint", fake_probe)

    ci._runtime_posture_check(FakeContext({}))

    assert probe_urls == ["http://192.168.50.157:11434/v1/models"]


def test_runtime_posture_check_rejects_loopback_ai_endpoint(monkeypatch):
    monkeypatch.setattr(ci.cache_tasks, "ensure_disk_headroom", lambda **_kwargs: None)
    monkeypatch.setenv("MYCELIS_COMPOSE_OLLAMA_HOST", "http://127.0.0.1:11434")

    with pytest.raises(SystemExit):
        ci._runtime_posture_check(FakeContext({}))


def test_runtime_posture_check_probes_compose_ai_endpoint_with_fallback(monkeypatch):
    monkeypatch.setattr(ci.cache_tasks, "ensure_disk_headroom", lambda **_kwargs: None)
    monkeypatch.setenv("MYCELIS_COMPOSE_OLLAMA_HOST", "http://10.0.0.5:11434")
    monkeypatch.setattr(ci, "running_in_wsl", lambda: False)

    probe_urls: list[str] = []

    def fake_probe(url: str, timeout: float = 3.0):
        probe_urls.append(url)
        if url.endswith("/api/tags"):
            return 404, "not found"
        return 401, "unauthorized"

    monkeypatch.setattr(ci, "_probe_http_endpoint", fake_probe)

    ci._runtime_posture_check(FakeContext({}))

    assert probe_urls == [
        "http://10.0.0.5:11434/api/tags",
        "http://10.0.0.5:11434/v1/models",
    ]


def test_runtime_posture_check_probes_wsl_localhost_mirror_for_host_docker_internal(monkeypatch):
    monkeypatch.setattr(ci.cache_tasks, "ensure_disk_headroom", lambda **_kwargs: None)
    monkeypatch.setenv("MYCELIS_COMPOSE_OLLAMA_HOST", "http://host.docker.internal:11434")
    monkeypatch.setattr(ci, "running_in_wsl", lambda: True)

    probe_urls: list[str] = []

    def fake_probe(url: str, timeout: float = 3.0):
        probe_urls.append(url)
        if url.startswith("http://host.docker.internal:11434/"):
            return 0, "connection refused"
        return 200, "{}"

    monkeypatch.setattr(ci, "_probe_http_endpoint", fake_probe)

    ci._runtime_posture_check(FakeContext({}))

    assert probe_urls == [
        "http://host.docker.internal:11434/api/tags",
        "http://host.docker.internal:11434/v1/models",
        "http://127.0.0.1:11434/api/tags",
    ]


def test_runtime_posture_check_probes_k8s_ai_endpoint(monkeypatch, tmp_path):
    monkeypatch.setattr(ci.cache_tasks, "ensure_disk_headroom", lambda **_kwargs: None)
    monkeypatch.setattr(ci, "ROOT_DIR", tmp_path)
    monkeypatch.delenv("MYCELIS_COMPOSE_OLLAMA_HOST", raising=False)
    monkeypatch.delenv("MYCELIS_K8S_MEDIA_ENDPOINT", raising=False)
    monkeypatch.delenv("MYCELIS_PROVIDER_LOCAL_OLLAMA_DEV_ENDPOINT", raising=False)
    for key in list(os.environ):
        if key.startswith("MYCELIS_PROVIDER_") and key.endswith("_ENDPOINT"):
            monkeypatch.delenv(key, raising=False)
    monkeypatch.setenv("MYCELIS_K8S_TEXT_ENDPOINT", "http://10.0.0.6:11434/v1")

    probe_urls: list[str] = []

    def fake_probe(url: str, timeout: float = 3.0):
        probe_urls.append(url)
        return 401, "unauthorized"

    monkeypatch.setattr(ci, "_probe_http_endpoint", fake_probe)

    ci._runtime_posture_check(FakeContext({}))

    assert probe_urls == [
        "http://10.0.0.6:11434/v1/models",
    ]
