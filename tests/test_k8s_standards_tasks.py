from __future__ import annotations

from contextlib import contextmanager
from dataclasses import dataclass
import json
from pathlib import Path
import shlex

from invoke import Context
import pytest

from ops import k8s_standards


@dataclass
class FakeResult:
    exited: int = 0
    stdout: str = ""
    stderr: str = ""

    @property
    def ok(self) -> bool:
        return self.exited == 0


class FakeContext(Context):
    def __init__(self, chart_version: str = "0.1.0"):
        super().__init__()
        self.chart_version = chart_version
        self.commands: list[str] = []

    def run(self, command: str, **_kwargs) -> FakeResult:
        self.commands.append(command)
        if " > " in command:
            rendered_target = Path(shlex.split(command.rsplit(" > ", 1)[1])[0])
            rendered_target.parent.mkdir(parents=True, exist_ok=True)
            rendered_target.write_text("kind: Deployment\n", encoding="utf-8")
        if command.startswith("helm package "):
            destination = Path(shlex.split(command.split("--destination ", 1)[1])[0])
            destination.mkdir(parents=True, exist_ok=True)
            (destination / f"mycelis-core-{self.chart_version}.tgz").write_bytes(b"chart")
        return FakeResult()

    @contextmanager
    def cd(self, _path: str):
        yield


def _write_minimal_chart(chart_dir: Path) -> None:
    templates = chart_dir / "templates"
    templates.mkdir(parents=True)
    (chart_dir / "Chart.yaml").write_text("apiVersion: v2\nname: mycelis-core\nversion: 0.1.0\n", encoding="utf-8")
    (chart_dir / "values.yaml").write_text(
        "serviceAccount:\n"
        "imagePullSecrets: []\n"
        "ingress:\n"
        "networkPolicy:\n"
        "outputBlock:\n"
        "  mode: cluster_generated\n"
        "coreAuth:\n"
        "  existingSecret: mycelis-core-auth\n"
        "probes:\n"
        "startup:\n"
        "readiness:\n"
        "liveness:\n"
        "podSecurityContext:\n"
        "securityContext:\n"
        "  readOnlyRootFilesystem: true\n"
        "  capabilities:\n"
        '    drop: ["ALL"]\n'
        "seccompProfile:\n"
        "storageClassName:\n"
        "search:\n"
        "oidc:\n",
        encoding="utf-8",
    )
    for relative, text in {
        "deployment.yaml": "apiVersion: apps/v1\nkind: Deployment\nserviceAccountName:\nimagePullSecrets:\nstartupProbe:\nreadinessProbe:\nlivenessProbe:\nsecurityContext:\nvolumeMounts:\nMYCELIS_OIDC_ENABLED\n{{- if .Values.env }}\n{{- toYaml .Values.env | nindent 12 }}\n",
        "service.yaml": "apiVersion: v1\nkind: Service\ntype: {{ .Values.service.type }}\n",
        "serviceaccount.yaml": "apiVersion: v1\nkind: ServiceAccount\n",
        "secrets.yaml": "apiVersion: v1\nkind: Secret\n",
        "configmap-config.yaml": "apiVersion: v1\nkind: ConfigMap\n",
        "data-pvc.yaml": "apiVersion: v1\nkind: PersistentVolumeClaim\n",
        "ingress.yaml": "apiVersion: networking.k8s.io/v1\nkind: Ingress\n",
        "network-policy.yaml": "apiVersion: networking.k8s.io/v1\nkind: NetworkPolicy\n",
    }.items():
        (templates / relative).write_text(text, encoding="utf-8")


def test_k8s_standards_static_gate_passes_for_chart_contract(capsys):
    ctx = FakeContext()

    k8s_standards.standards.body(ctx)

    output = capsys.readouterr().out
    assert "Kubernetes Open Standards Gate" in output
    assert "Static chart contract: OK" in output
    assert "Compose is rapid local development only" in output
    assert ctx.commands == []


def test_k8s_standards_helm_gate_runs_lint_and_template(monkeypatch, tmp_path: Path):
    ctx = FakeContext()
    chart_dir = tmp_path / "charts" / "mycelis-core"
    _write_minimal_chart(chart_dir)

    monkeypatch.setattr(k8s_standards, "ROOT_DIR", tmp_path)

    k8s_standards.standards.body(ctx, helm=True)

    assert not any(command.startswith("helm dependency build ") for command in ctx.commands)
    assert any(command.startswith("helm lint ") for command in ctx.commands)
    assert any(command.startswith("helm template ") for command in ctx.commands)


def test_k8s_standards_fails_when_oidc_values_missing(monkeypatch, tmp_path: Path):
    ctx = FakeContext()
    chart_dir = tmp_path / "charts" / "mycelis-core"
    _write_minimal_chart(chart_dir)
    values_path = chart_dir / "values.yaml"
    values_path.write_text(values_path.read_text(encoding="utf-8").replace("oidc:\n", ""), encoding="utf-8")
    monkeypatch.setattr(k8s_standards, "ROOT_DIR", tmp_path)

    with pytest.raises(SystemExit, match="values.yaml: oidc:"):
        k8s_standards.standards.body(ctx)


def test_k8s_standards_fails_when_oidc_deployment_env_missing(monkeypatch, tmp_path: Path):
    ctx = FakeContext()
    chart_dir = tmp_path / "charts" / "mycelis-core"
    _write_minimal_chart(chart_dir)
    deployment_path = chart_dir / "templates" / "deployment.yaml"
    deployment_path.write_text(
        deployment_path.read_text(encoding="utf-8").replace("MYCELIS_OIDC_ENABLED\n", ""),
        encoding="utf-8",
    )
    monkeypatch.setattr(k8s_standards, "ROOT_DIR", tmp_path)

    with pytest.raises(SystemExit, match="templates/deployment.yaml: MYCELIS_OIDC_ENABLED"):
        k8s_standards.standards.body(ctx)


def test_verify_and_package_chart_writes_manifest_and_checksums(tmp_path: Path):
    chart_dir = tmp_path / "charts" / "mycelis-core"
    _write_minimal_chart(chart_dir)
    values_enterprise = chart_dir / "values-enterprise.yaml"
    values_enterprise.write_text("replicaCount: 1\n", encoding="utf-8")

    ctx = FakeContext(chart_version="0.1.0")
    manifest_path = k8s_standards._verify_and_package_chart(
        ctx,
        values_file=values_enterprise,
        release_label="v0.8.6",
        package_output_dir="dist/helm/enterprise",
        root_dir=tmp_path,
    )

    manifest = json.loads(manifest_path.read_text(encoding="utf-8"))
    assert manifest["artifact_kind"] == "enterprise_helm_package"
    assert manifest["verification_status"] == "verified"
    assert "OIDC/RBAC enterprise contract surfaces" in manifest["notes"][0]
    assert (tmp_path / manifest["rendered_bundle_checksum_path"]).exists()
    assert (tmp_path / manifest["chart_archive_checksum_path"]).exists()


def test_k8s_backend_prefers_rancher_desktop_when_rdctl_and_context_active(monkeypatch):
    from ops import k8s

    monkeypatch.delenv("MYCELIS_K8S_BACKEND", raising=False)
    monkeypatch.setattr(k8s, "_ensure_rancher_desktop_path", lambda: [])
    monkeypatch.setattr(k8s, "_rancher_desktop_context_available", lambda: True)
    monkeypatch.setattr(
        k8s.shutil,
        "which",
        lambda name: f"/usr/bin/{name}" if name in {"rdctl", "k3d", "kind"} else None,
    )

    assert k8s._k8s_backend() == "rancher-desktop"


def test_k8s_backend_accepts_rancher_alias(monkeypatch):
    from ops import k8s

    monkeypatch.setenv("MYCELIS_K8S_BACKEND", "rancher")
    monkeypatch.setattr(k8s, "_ensure_rancher_desktop_path", lambda: [])
    monkeypatch.setattr(k8s.shutil, "which", lambda name: f"/usr/bin/{name}" if name == "rdctl" else None)

    assert k8s._k8s_backend() == "rancher-desktop"


def test_k8s_rancher_desktop_init_deploy_and_reset(monkeypatch, capsys):
    from ops import k8s

    ctx = FakeContext()
    monkeypatch.setattr(k8s, "_k8s_backend", lambda: "rancher-desktop")
    monkeypatch.setattr(k8s.core_build, "body", lambda _ctx: "v0.8.7-rancher")
    monkeypatch.setenv("POSTGRES_USER", "mycelis")
    monkeypatch.setenv("POSTGRES_PASSWORD", "password")
    monkeypatch.setenv("POSTGRES_DB", "cortex")
    monkeypatch.setenv("MYCELIS_API_KEY", "dev-key")

    k8s.init.body(ctx)
    assert "kubectl config use-context rancher-desktop" in ctx.commands

    k8s.deploy.body(ctx)
    output = capsys.readouterr().out
    assert "Rancher Desktop Moby runtime active: image mycelis/core:v0.8.7-rancher shared with cluster." in output
    assert not any("k3d image import" in c for c in ctx.commands)
    assert not any("kind load docker-image" in c for c in ctx.commands)

    monkeypatch.setattr(k8s, "up", lambda _c: ctx.commands.append("up"))
    k8s.reset.body(ctx)
    assert "helm uninstall mycelis-core -n mycelis" in ctx.commands
    assert "kubectl delete namespace mycelis --ignore-not-found" in ctx.commands
