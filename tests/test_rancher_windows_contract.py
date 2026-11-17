from __future__ import annotations

import ops.config
from ops.k8s_rancher import (
    DEFAULT_RANCHER_CONTEXT,
    DEFAULT_RANCHER_SERVER_URL,
    rancher_target_context,
    rancher_target_url,
)


def test_docker_host_mode_defaults_to_native_on_windows_without_wsl_probe(monkeypatch):
    ops.config.docker_host_mode.cache_clear()
    try:
        monkeypatch.setattr(ops.config, "is_windows", lambda: True)
        monkeypatch.setattr(
            ops.config,
            "wsl_has_command",
            lambda _cmd: (_ for _ in ()).throw(AssertionError("should not probe WSL in auto mode")),
        )

        monkeypatch.delenv("MYCELIS_DOCKER_HOST", raising=False)
        ops.config.docker_host_mode.cache_clear()
        assert ops.config.docker_host_mode() == "native"

        monkeypatch.setenv("MYCELIS_DOCKER_HOST", "auto")
        ops.config.docker_host_mode.cache_clear()
        assert ops.config.docker_host_mode() == "native"
    finally:
        ops.config.docker_host_mode.cache_clear()


def test_rancher_target_defaults_and_env_overrides(monkeypatch):
    monkeypatch.delenv("MYCELIS_RANCHER_URL", raising=False)
    monkeypatch.delenv("MYCELIS_K8S_SERVER", raising=False)
    assert DEFAULT_RANCHER_SERVER_URL == "https://127.0.0.1:6443"
    assert rancher_target_url() == "https://127.0.0.1:6443"

    monkeypatch.setenv("MYCELIS_K8S_SERVER", "https://127.0.0.1:16443")
    assert rancher_target_url() == "https://127.0.0.1:16443"

    monkeypatch.setenv("MYCELIS_RANCHER_URL", "https://127.0.0.1:26443")
    assert rancher_target_url() == "https://127.0.0.1:26443"

    monkeypatch.delenv("MYCELIS_K8S_CONTEXT", raising=False)
    monkeypatch.delenv("MYCELIS_RANCHER_CONTEXT", raising=False)
    assert DEFAULT_RANCHER_CONTEXT == "rancher-desktop"
    assert rancher_target_context() == "rancher-desktop"

    monkeypatch.setenv("MYCELIS_RANCHER_CONTEXT", "custom-rancher-ctx")
    assert rancher_target_context() == "custom-rancher-ctx"

    monkeypatch.setenv("MYCELIS_K8S_CONTEXT", "custom-k8s-ctx")
    assert rancher_target_context() == "custom-k8s-ctx"


def test_agents_md_codifies_two_stage_windows_rancher_contract():
    agents_md = (ops.config.ROOT_DIR / "AGENTS.md").read_text(encoding="utf-8")
    assert "Windows-Only Local Execution + Windows Rancher Desktop Deployment" in agents_md
    assert "Stage 1 Local Execution & Testing" in agents_md
    assert "Stage 2 Containerization & Cluster Deployment" in agents_md


def test_helm_chart_and_k8s_helpers_support_optional_interface_deployment():
    values_yaml = (ops.config.ROOT_DIR / "charts" / "mycelis-core" / "values.yaml").read_text(encoding="utf-8")
    assert "interface:" in values_yaml
    assert "repository: mycelis/interface" in values_yaml

    interface_tmpl = (
        ops.config.ROOT_DIR / "charts" / "mycelis-core" / "templates" / "interface-deployment.yaml"
    ).read_text(encoding="utf-8")
    assert "name: {{ .Chart.Name }}-interface" in interface_tmpl
    assert "name: MYCELIS_API_HOST" in interface_tmpl
    assert "name: MYCELIS_API_PORT" in interface_tmpl
    assert "name: MYCELIS_API_KEY" in interface_tmpl

    k8s_helpers = (ops.config.ROOT_DIR / "ops" / "k8s_helpers.py").read_text(encoding="utf-8")
    assert "MYCELIS_K8S_WITH_INTERFACE" in k8s_helpers
    assert "--set interface.enabled=true" in k8s_helpers

