from __future__ import annotations

from contextlib import contextmanager
from dataclasses import dataclass
from pathlib import Path

from invoke import Context

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
    def __init__(self):
        super().__init__()
        self.commands: list[str] = []

    def run(self, command: str, **_kwargs) -> FakeResult:
        self.commands.append(command)
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
        "search:\n",
        encoding="utf-8",
    )
    for relative, text in {
        "deployment.yaml": "apiVersion: apps/v1\nkind: Deployment\nserviceAccountName:\nimagePullSecrets:\nstartupProbe:\nreadinessProbe:\nlivenessProbe:\nsecurityContext:\nvolumeMounts:\n{{- if .Values.env }}\n{{- toYaml .Values.env | nindent 12 }}\n",
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
