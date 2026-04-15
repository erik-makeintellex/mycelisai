from __future__ import annotations

from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
VALUES = ROOT / "charts" / "mycelis-core" / "values.yaml"
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
