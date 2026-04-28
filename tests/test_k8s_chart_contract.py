from __future__ import annotations

from contextlib import contextmanager
from dataclasses import dataclass
import json
from pathlib import Path
import shlex

from invoke import Context
import pytest

from ops import k8s

ROOT = Path(__file__).resolve().parents[1]
VALUES = ROOT / "charts" / "mycelis-core" / "values.yaml"
VALUES_K3D = ROOT / "charts" / "mycelis-core" / "values-k3d.yaml"
VALUES_ENTERPRISE = ROOT / "charts" / "mycelis-core" / "values-enterprise.yaml"
VALUES_ENTERPRISE_WINDOWS_AI = ROOT / "charts" / "mycelis-core" / "values-enterprise-windows-ai.yaml"
CHART_LOCK = ROOT / "charts" / "mycelis-core" / "Chart.lock"
DEPLOYMENT = ROOT / "charts" / "mycelis-core" / "templates" / "deployment.yaml"
HELPERS = ROOT / "charts" / "mycelis-core" / "templates" / "_helpers.tpl"
INGRESS = ROOT / "charts" / "mycelis-core" / "templates" / "ingress.yaml"
SERVICE_ACCOUNT = ROOT / "charts" / "mycelis-core" / "templates" / "serviceaccount.yaml"


@dataclass
class FakeResult:
    exited: int = 0
    stdout: str = ""
    stderr: str = ""

    @property
    def ok(self) -> bool:
        return self.exited == 0


class FakeContext(Context):
    def __init__(self, chart_version: str):
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


def test_chart_exposes_first_enterprise_k8s_override_surfaces():
    values_text = VALUES.read_text(encoding="utf-8")
    required_value_blocks = [
        "serviceAccount:",
        "imagePullSecrets:",
        "nodeSelector:",
        "tolerations:",
        "affinity:",
        "topologySpreadConstraints:",
        "ingress:",
        "digest:",
    ]

    missing = [block for block in required_value_blocks if block not in values_text]
    assert not missing, "values.yaml is missing enterprise-Kubernetes override surfaces:\n" + "\n".join(missing)
    assert "tag: latest" not in values_text, "values.yaml should not default the chart image tag to `latest`"


def test_chart_templates_project_enterprise_k8s_surfaces_into_workload():
    deployment_text = DEPLOYMENT.read_text(encoding="utf-8")
    required_deployment_snippets = [
        'serviceAccountName: {{ $serviceAccountName }}',
        "imagePullSecrets:",
        "nodeSelector:",
        "tolerations:",
        "affinity:",
        "topologySpreadConstraints:",
        'image: "{{ include "mycelis-core.image" . }}"',
    ]
    missing = [snippet for snippet in required_deployment_snippets if snippet not in deployment_text]
    assert not missing, "deployment.yaml is missing projected enterprise-Kubernetes workload controls:\n" + "\n".join(missing)

    helpers_text = HELPERS.read_text(encoding="utf-8")
    assert 'define "mycelis-core.image"' in helpers_text
    assert "@%s" in helpers_text, "image helper should support digest-based image references"
    assert 'define "mycelis-core.serviceAccountName"' in helpers_text


def test_chart_projects_search_provider_env_into_workload():
    values_text = VALUES.read_text(encoding="utf-8")
    deployment_text = DEPLOYMENT.read_text(encoding="utf-8")

    required_value_snippets = [
        "search:",
        "provider: disabled",
        'searxngEndpoint: ""',
        'localApiEndpoint: ""',
        "maxResults: 8",
    ]
    missing_values = [snippet for snippet in required_value_snippets if snippet not in values_text]
    assert not missing_values, "values.yaml is missing search provider values:\n" + "\n".join(missing_values)

    required_deployment_snippets = [
        "MYCELIS_SEARCH_PROVIDER",
        "MYCELIS_SEARXNG_ENDPOINT",
        "MYCELIS_SEARCH_LOCAL_API_ENDPOINT",
        "MYCELIS_SEARCH_MAX_RESULTS",
        'value: {{ default "disabled" .provider | quote }}',
        'value: {{ default "" .searxngEndpoint | quote }}',
        'value: {{ default "" .localApiEndpoint | quote }}',
        "value: {{ default 8 .maxResults | quote }}",
    ]
    missing_deployment = [snippet for snippet in required_deployment_snippets if snippet not in deployment_text]
    assert not missing_deployment, "deployment.yaml is missing search provider env projection:\n" + "\n".join(missing_deployment)


def test_chart_adds_ingress_and_serviceaccount_templates():
    ingress_text = INGRESS.read_text(encoding="utf-8")
    assert "kind: Ingress" in ingress_text
    assert "ingress.hosts must be set when ingress.enabled=true" in ingress_text

    service_account_text = SERVICE_ACCOUNT.read_text(encoding="utf-8")
    assert "kind: ServiceAccount" in service_account_text
    assert ".Values.serviceAccount.create" in service_account_text


def test_chart_presets_cover_local_k3d_and_enterprise_postures():
    checks = [
        (
            VALUES_K3D,
            [
                "serviceAccount:\n  create: false",
                "coreAuth:\n  apiKey: k3d-local-dev-key",
                "ingress:\n  enabled: false",
                "storageClassName: local-path",
                "storageClass: local-path",
            ],
        ),
        (
            VALUES_ENTERPRISE,
            [
                "serviceAccount:\n  create: true",
                "image:\n  repository: registry.example.com/mycelis/core",
                "imagePullSecrets:\n  - name: mycelis-registry",
                "ingress:\n  enabled: true",
                "coreAuth:\n  existingSecret: mycelis-core-auth",
                "storageClassName: standard",
                "storageClass: standard",
            ],
        ),
        (
            VALUES_ENTERPRISE_WINDOWS_AI,
            [
                "Enterprise self-hosted Kubernetes preset with a Windows-hosted AI endpoint.",
                "serviceAccount:\n  create: true",
                "imagePullSecrets:\n  - name: mycelis-registry",
                "ingress:\n  enabled: true",
                "coreAuth:\n  existingSecret: mycelis-core-auth",
                'textEndpoint: "http://<windows-ai-host>:11434/v1"',
                "storageClassName: standard",
                "storageClass: standard",
            ],
        ),
    ]

    missing: list[str] = []
    for path, snippets in checks:
        text = path.read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Chart preset values are missing the expected deployment posture choices:\n" + "\n".join(missing)


def test_chart_lock_pins_release_packaging_dependencies():
    chart_lock_text = CHART_LOCK.read_text(encoding="utf-8")

    assert "digest:" in chart_lock_text
    assert "name: postgresql" in chart_lock_text
    assert "name: nats" in chart_lock_text


def test_verify_package_mode_requires_explicit_values_file():
    with pytest.raises(SystemExit, match="A Helm values file is required for --verify-package"):
        k8s.deploy.body(Context(), verify_package=True)


def test_verify_package_mode_writes_enterprise_release_artifacts(monkeypatch, tmp_path: Path):
    chart_dir = tmp_path / "charts" / "mycelis-core"
    chart_dir.mkdir(parents=True)
    (chart_dir / "Chart.yaml").write_text(
        "apiVersion: v2\n"
        "name: mycelis-core\n"
        "version: 0.1.0\n",
        encoding="utf-8",
    )
    (chart_dir / "values-enterprise.yaml").write_text("replicaCount: 1\n", encoding="utf-8")

    ctx = FakeContext(chart_version="0.1.0")
    monkeypatch.setattr(k8s, "ROOT_DIR", tmp_path)

    manifest_path = k8s.deploy.body(
        ctx,
        verify_package=True,
        values_file="charts/mycelis-core/values-enterprise.yaml",
        release_label="manual/42",
        package_output_dir="dist/helm/enterprise",
    )

    manifest = json.loads(Path(manifest_path).read_text(encoding="utf-8"))
    assert manifest["artifact_kind"] == "enterprise_helm_package"
    assert manifest["preset"] == "enterprise"
    assert manifest["release_label"] == "manual/42"
    assert manifest["status"] == "scaffold"
    assert manifest["values_file"] == "charts/mycelis-core/values-enterprise.yaml"
    assert (tmp_path / manifest["chart_archive_path"]).exists()
    assert (tmp_path / manifest["rendered_bundle_path"]).exists()
    assert any(command.startswith("helm dependency build ") for command in ctx.commands)
    assert any(command.startswith("helm lint ") for command in ctx.commands)
    assert any(command.startswith("helm template ") for command in ctx.commands)
    assert any(command.startswith("helm package ") for command in ctx.commands)
