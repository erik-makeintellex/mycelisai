from __future__ import annotations

import ops.config
from ops.k8s_rancher import (
    DEFAULT_RANCHER_CONTEXT,
    DEFAULT_RANCHER_SERVER_URL,
    MAX_RETAINED_IMAGE_VERSIONS,
    prune_old_mycelis_images,
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


def test_db_require_postgres_auto_bridges_localhost_when_port_closed(monkeypatch):
    import ops.db
    import ops.k8s_helpers

    monkeypatch.setattr(ops.db, "_dsn", lambda dbname=None: ("127.0.0.1", "5432", "mycelis", "password", "postgres"))
    psql_calls = []

    def fake_psql(sql=None, file=None, dbname=None):
        psql_calls.append((sql, dbname))
        return 1 if len(psql_calls) == 1 else 0

    forwarded = []
    monkeypatch.setattr(ops.db, "_psql", fake_psql)
    monkeypatch.setattr(ops.k8s_helpers, "_port_open", lambda port, host="127.0.0.1", timeout=1.0: False)
    monkeypatch.setattr(
        ops.k8s_helpers,
        "_start_port_forward_detached",
        lambda service, forward: forwarded.append((service, forward)),
    )
    monkeypatch.setattr(
        ops.k8s_helpers,
        "_wait_for_local_port",
        lambda port, label, timeout_seconds=15, interval_seconds=1.0: True,
    )

    ops.db._require_postgres("postgres")
    assert len(psql_calls) == 2
    assert forwarded == [("svc/mycelis-core-postgresql", "5432:5432")]


def test_k8s_up_defaults_to_with_interface_on_rancher_desktop(monkeypatch):
    import ops.k8s

    recorded = {}
    monkeypatch.setattr(ops.k8s, "_k8s_backend", lambda: "rancher-desktop")
    monkeypatch.setattr(ops.k8s, "init", lambda c: None)
    monkeypatch.setattr(
        ops.k8s,
        "deploy",
        lambda c, with_interface=False: recorded.setdefault("with_interface", with_interface),
    )
    monkeypatch.setattr(ops.k8s, "wait", lambda c, timeout=180: None)
    monkeypatch.delenv("MYCELIS_K8S_WITH_INTERFACE", raising=False)

    ops.k8s.up.body(object())
    assert recorded["with_interface"] is True


def test_prune_old_mycelis_images_retains_at_most_two_versions_per_repo(monkeypatch):
    import subprocess
    import ops.k8s_rancher as kr

    assert MAX_RETAINED_IMAGE_VERSIONS == 2
    sample_output = "\n".join(
        [
            "mycelis/core\tv1.2.1-newest",
            "mycelis/core\tv1.2.0-prev",
            "mycelis/core\tv1.1.9-old1",
            "mycelis/core\tv1.1.8-old2",
            "mycelis/interface\tv1.2.1-newest",
            "mycelis/interface\tv1.2.0-prev",
            "mycelis/interface\tv1.1.9-old1",
            "postgres\t15-alpine",
        ]
    )
    commands: list[list[str]] = []

    def fake_run(cmd, **kwargs):
        commands.append(list(cmd))
        if cmd[:2] == ["docker", "images"]:
            return subprocess.CompletedProcess(cmd, 0, stdout=sample_output, stderr="")
        return subprocess.CompletedProcess(cmd, 0, stdout="", stderr="")

    monkeypatch.setattr(kr.subprocess, "run", fake_run)
    removed = prune_old_mycelis_images("v1.2.1-newest")

    assert removed == [
        "mycelis/core:v1.1.9-old1",
        "mycelis/core:v1.1.8-old2",
        "mycelis/interface:v1.1.9-old1",
    ]
    assert any(cmd[:3] == ["docker", "rmi", "-f"] for cmd in commands)
    assert any(cmd[:3] == ["docker", "image", "prune"] for cmd in commands)
