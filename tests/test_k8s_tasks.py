from __future__ import annotations

from contextlib import contextmanager
from dataclasses import dataclass
from pathlib import Path

from invoke import Context
import pytest

from ops import k8s


@dataclass
class FakeResult:
    exited: int = 0
    stdout: str = ""
    stderr: str = ""

    @property
    def ok(self) -> bool:
        return self.exited == 0


class FakeContext(Context):
    def __init__(self, command_results: dict[str, FakeResult] | None = None):
        super().__init__()
        self.command_results = command_results or {}
        self.commands: list[str] = []

    def run(self, command: str, **_kwargs) -> FakeResult:
        self.commands.append(command)
        return self.command_results.get(command, FakeResult())

    @contextmanager
    def cd(self, _path: str):
        yield


def test_deploy_uses_core_build_task_body(monkeypatch):
    ctx = FakeContext()
    build_calls: list[str] = []

    monkeypatch.setattr(k8s, "_k8s_backend", lambda: "kind")
    monkeypatch.setattr(k8s.core_build, "body", lambda _ctx: build_calls.append("build") or "v0.1.0-deadbee")
    monkeypatch.setenv("POSTGRES_USER", "mycelis")
    monkeypatch.setenv("POSTGRES_PASSWORD", "password")
    monkeypatch.setenv("POSTGRES_DB", "cortex")
    monkeypatch.setenv("MYCELIS_API_KEY", "dev-key")

    k8s.deploy.body(ctx)

    assert build_calls == ["build"]
    assert any("kind load docker-image mycelis/core:v0.1.0-deadbee" in command for command in ctx.commands)
    assert any("--set image.tag=v0.1.0-deadbee" in command for command in ctx.commands)


def test_deploy_uses_k3d_image_import_when_k3d_backend_selected(monkeypatch, capsys):
    ctx = FakeContext()

    monkeypatch.setattr(k8s, "_k8s_backend", lambda: "k3d")
    monkeypatch.setattr(k8s.core_build, "body", lambda _ctx: "v0.1.0-deadbee")
    monkeypatch.setenv("POSTGRES_USER", "mycelis")
    monkeypatch.setenv("POSTGRES_PASSWORD", "password")
    monkeypatch.setenv("POSTGRES_DB", "cortex")
    monkeypatch.setenv("MYCELIS_API_KEY", "dev-key")

    k8s.deploy.body(ctx)
    output = capsys.readouterr().out

    assert "Deployment posture: k3d validation" in output
    assert any("k3d image import mycelis/core:v0.1.0-deadbee -c mycelis-cluster" in command for command in ctx.commands)
    assert any("--set image.tag=v0.1.0-deadbee" in command for command in ctx.commands)


def test_deploy_includes_explicit_ai_endpoint_overrides(monkeypatch, tmp_path: Path, capsys):
    ctx = FakeContext()
    values_file = tmp_path / "values-enterprise-windows-ai.yaml"
    values_file.write_text(
        "ai:\n"
        '  textEndpoint: "http://<windows-ai-host>:11434/v1"\n'
        '  mediaEndpoint: ""\n',
        encoding="utf-8",
    )

    monkeypatch.setattr(k8s, "_k8s_backend", lambda: "k3d")
    monkeypatch.setattr(k8s, "ROOT_DIR", tmp_path)
    monkeypatch.setattr(k8s.core_build, "body", lambda _ctx: "v0.1.0-deadbee")
    monkeypatch.setenv("POSTGRES_USER", "mycelis")
    monkeypatch.setenv("POSTGRES_PASSWORD", "password")
    monkeypatch.setenv("POSTGRES_DB", "cortex")
    monkeypatch.setenv("MYCELIS_API_KEY", "dev-key")
    monkeypatch.setenv("MYCELIS_K8S_TEXT_ENDPOINT", "http://192.168.50.156:11434/v1")
    monkeypatch.setenv("MYCELIS_K8S_MEDIA_ENDPOINT", "http://192.168.50.156:8001/v1")
    monkeypatch.setenv("MYCELIS_K8S_VALUES_FILE", values_file.name)

    k8s.deploy.body(ctx)
    output = capsys.readouterr().out

    assert "Deployment posture: enterprise self-hosted with Windows-hosted AI" in output
    helm_command = next(command for command in ctx.commands if command.startswith("helm upgrade --install"))
    assert f"--set-string ai.textEndpoint={k8s._shell_quote('http://192.168.50.156:11434/v1')}" in helm_command
    assert f"--set-string ai.mediaEndpoint={k8s._shell_quote('http://192.168.50.156:8001/v1')}" in helm_command
    assert f"--values {k8s._shell_quote(values_file.resolve())}" in helm_command


def test_deploy_requires_explicit_windows_ai_endpoint_for_enterprise_preset(monkeypatch, tmp_path: Path):
    ctx = FakeContext()
    values_file = tmp_path / "values-enterprise-windows-ai.yaml"
    values_file.write_text(
        "ai:\n"
        '  textEndpoint: "http://<windows-ai-host>:11434/v1"\n'
        '  mediaEndpoint: ""\n',
        encoding="utf-8",
    )

    monkeypatch.setattr(k8s, "_k8s_backend", lambda: "k3d")
    monkeypatch.setattr(k8s, "ROOT_DIR", tmp_path)
    monkeypatch.setattr(k8s.core_build, "body", lambda _ctx: "v0.1.0-deadbee")
    monkeypatch.setenv("POSTGRES_USER", "mycelis")
    monkeypatch.setenv("POSTGRES_PASSWORD", "password")
    monkeypatch.setenv("POSTGRES_DB", "cortex")
    monkeypatch.setenv("MYCELIS_API_KEY", "dev-key")
    monkeypatch.setenv("MYCELIS_K8S_VALUES_FILE", values_file.name)

    with pytest.raises(SystemExit, match="MYCELIS_K8S_TEXT_ENDPOINT must be set"):
        k8s.deploy.body(ctx)


def test_deploy_includes_repo_relative_values_file(monkeypatch, tmp_path: Path):
    ctx = FakeContext()
    values_file = tmp_path / "preset files" / "k8s-values.yaml"
    values_file.parent.mkdir(parents=True)
    values_file.write_text("replicaCount: 1\n", encoding="utf-8")

    monkeypatch.setattr(k8s, "_k8s_backend", lambda: "k3d")
    monkeypatch.setattr(k8s, "ROOT_DIR", tmp_path)
    monkeypatch.setattr(k8s.core_build, "body", lambda _ctx: "v0.1.0-deadbee")
    monkeypatch.setenv("POSTGRES_USER", "mycelis")
    monkeypatch.setenv("POSTGRES_PASSWORD", "password")
    monkeypatch.setenv("POSTGRES_DB", "cortex")
    monkeypatch.setenv("MYCELIS_API_KEY", "dev-key")
    monkeypatch.setenv("MYCELIS_K8S_VALUES_FILE", "preset files/k8s-values.yaml")

    k8s.deploy.body(ctx)

    helm_command = next(command for command in ctx.commands if command.startswith("helm upgrade --install"))
    expected_path = k8s._shell_quote(values_file.resolve())
    assert f"--values {expected_path}" in helm_command


def test_deploy_fails_fast_when_values_file_is_missing(monkeypatch, tmp_path: Path):
    ctx = FakeContext()

    monkeypatch.setattr(k8s, "_k8s_backend", lambda: "k3d")
    monkeypatch.setattr(k8s, "ROOT_DIR", tmp_path)
    monkeypatch.setattr(k8s.core_build, "body", lambda _ctx: "v0.1.0-deadbee")
    monkeypatch.setenv("POSTGRES_USER", "mycelis")
    monkeypatch.setenv("POSTGRES_PASSWORD", "password")
    monkeypatch.setenv("POSTGRES_DB", "cortex")
    monkeypatch.setenv("MYCELIS_API_KEY", "dev-key")
    monkeypatch.setenv("MYCELIS_K8S_VALUES_FILE", "preset files/missing.yaml")

    with pytest.raises(SystemExit, match="MYCELIS_K8S_VALUES_FILE does not exist"):
        k8s.deploy.body(ctx)


def test_k8s_backend_prefers_k3d_when_available(monkeypatch):
    monkeypatch.delenv("MYCELIS_K8S_BACKEND", raising=False)
    monkeypatch.setattr(
        k8s.shutil,
        "which",
        lambda name: f"/usr/bin/{name}" if name in {"k3d", "kind"} else None,
    )

    assert k8s._k8s_backend() == "k3d"


def test_init_creates_k3d_cluster_when_k3d_backend_selected(monkeypatch):
    ctx = FakeContext()
    monkeypatch.setattr(k8s, "_k8s_backend", lambda: "k3d")
    monkeypatch.setattr(k8s, "_cluster_exists", lambda _c: False)

    k8s.init.body(ctx)

    assert "k3d cluster create mycelis-cluster" in ctx.commands


def test_init_reads_and_writes_kind_config_under_root_dir(monkeypatch, tmp_path: Path):
    ctx = FakeContext()
    monkeypatch.setattr(k8s, "_k8s_backend", lambda: "kind")
    monkeypatch.setattr(k8s, "_cluster_exists", lambda _c: False)
    monkeypatch.setattr(k8s, "ROOT_DIR", tmp_path)

    (tmp_path / "kind-config.yaml").write_text(
        "mounts:\n  - hostPath: ./ops\n  - hostPath: ./logs\n",
        encoding="utf-8",
    )

    k8s.init.body(ctx)

    generated = tmp_path / "kind-config.gen.yaml"
    assert generated.exists()
    generated_text = generated.read_text(encoding="utf-8")
    assert str(tmp_path / "ops").replace("\\", "/") in generated_text
    assert str(tmp_path / "logs").replace("\\", "/") in generated_text
    assert any(str(generated) in command for command in ctx.commands)


def test_reset_prints_managed_status_command(monkeypatch, capsys):
    ctx = FakeContext()
    monkeypatch.setattr(k8s, "_k8s_backend", lambda: "kind")
    monkeypatch.setattr(k8s, "up", lambda _ctx: ctx.commands.append("up"))

    k8s.reset.body(ctx)

    output = capsys.readouterr().out
    assert "uv run inv k8s.status" in output
    assert "Run 'inv k8s.status' to verify." not in output


def test_reset_uses_k3d_cluster_delete_when_k3d_backend_selected(monkeypatch):
    ctx = FakeContext()
    monkeypatch.setattr(k8s, "_k8s_backend", lambda: "k3d")
    monkeypatch.setattr(k8s, "up", lambda _ctx: ctx.commands.append("up"))

    k8s.reset.body(ctx)

    assert "k3d cluster delete mycelis-cluster" in ctx.commands


def test_cluster_exists_parses_kind_get_clusters_directly(monkeypatch):
    ctx = FakeContext({"kind get clusters": FakeResult(stdout="other\nmycelis-cluster\n")})
    monkeypatch.setattr(k8s, "_k8s_backend", lambda: "kind")

    assert k8s._cluster_exists(ctx)

    assert ctx.commands == ["kind get clusters"]


def test_cluster_exists_parses_k3d_cluster_list_directly(monkeypatch):
    ctx = FakeContext(
        {
            "k3d cluster list": FakeResult(
                stdout="NAME SERVERS AGENTS LOADBALANCER\nmycelis-cluster running 1 1\n"
            )
        }
    )
    monkeypatch.setattr(k8s, "_k8s_backend", lambda: "k3d")

    assert k8s._cluster_exists(ctx)

    assert ctx.commands == ["k3d cluster list"]


def test_status_skips_cluster_checks_when_docker_is_down(capsys):
    class DownContext(FakeContext):
        def run(self, command: str, **_kwargs):
            self.commands.append(command)
            if command == "docker info":
                raise RuntimeError("docker unavailable")
            return FakeResult()

    ctx = DownContext()

    k8s.status.body(ctx)

    output = capsys.readouterr().out
    assert "Docker: NOT Running." in output
    assert "Local Kubernetes Cluster: SKIPPED (Docker down)" in output
    assert "Pod Status: SKIPPED" in output
    assert "Persistence (PVC) Status: SKIPPED" in output
    assert ctx.commands == ["docker info"]


def test_recover_fails_when_restart_signal_cannot_be_sent(monkeypatch):
    ctx = FakeContext({
        "kubectl cluster-info": FakeResult(),
        "kubectl rollout restart deployment/mycelis-core -n mycelis": FakeResult(exited=1, stderr="boom"),
    })
    monkeypatch.setattr(k8s, "_resource_exists", lambda _c, _resource: True)
    monkeypatch.setattr(k8s, "wait", lambda *_args, **_kwargs: None)

    with pytest.raises(SystemExit, match="unable to send restart signal"):
        k8s.recover.body(ctx)
