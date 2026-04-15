from __future__ import annotations

from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
VALUES = ROOT / "charts" / "mycelis-core" / "values.yaml"
VALUES_K3D = ROOT / "charts" / "mycelis-core" / "values-k3d.yaml"
VALUES_ENTERPRISE = ROOT / "charts" / "mycelis-core" / "values-enterprise.yaml"
VALUES_ENTERPRISE_WINDOWS_AI = ROOT / "charts" / "mycelis-core" / "values-enterprise-windows-ai.yaml"
DEPLOYMENT = ROOT / "charts" / "mycelis-core" / "templates" / "deployment.yaml"
HELPERS = ROOT / "charts" / "mycelis-core" / "templates" / "_helpers.tpl"
INGRESS = ROOT / "charts" / "mycelis-core" / "templates" / "ingress.yaml"
SERVICE_ACCOUNT = ROOT / "charts" / "mycelis-core" / "templates" / "serviceaccount.yaml"


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
