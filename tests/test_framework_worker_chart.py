"""Render-only checks for the disabled-by-default Runs Helm package."""

from __future__ import annotations

import shutil
import subprocess
from pathlib import Path

import pytest
import yaml


CHART = Path(__file__).resolve().parents[1] / "charts" / "mycelis-core"
DIGEST = "sha256:" + "a" * 64


def helm_template(*extra: str) -> subprocess.CompletedProcess[str]:
    helm = shutil.which("helm")
    if not helm:
        pytest.skip("pinned Helm binary is not on PATH")
    return subprocess.run(
        [helm, "template", "b2fixture", str(CHART), *extra],
        capture_output=True,
        text=True,
        check=False,
        timeout=30,
    )


def enabled_args(*extra: str) -> tuple[str, ...]:
    return (
        "--set", "frameworkWorker.enabled=true",
        "--set-string", f"frameworkWorker.image.digest={DIGEST}",
        "--set-string", "frameworkWorker.auth.existingSecret=runs-auth",
        "--set-string", "frameworkWorker.database.existingSecret=runs-db",
        "--set-string", "frameworkWorker.database.egressCIDRs[0]=10.42.0.5/32",
        "--set-string", "frameworkWorker.tls.existingSecret=runs-tls",
        "--set-string", "frameworkWorker.tls.serverName=framework-runs.b2.svc.cluster.local",
        *extra,
    )


def test_worker_is_absent_from_default_chart():
    result = helm_template()
    assert result.returncode == 0, result.stderr
    assert not any(
        "framework-runs" in doc.get("metadata", {}).get("name", "")
        for doc in yaml.safe_load_all(result.stdout) if isinstance(doc, dict)
    )


def test_worker_render_has_scoped_secret_network_and_probe_contract():
    result = helm_template(*enabled_args())
    assert result.returncode == 0, result.stderr
    docs = {
        (doc["kind"], doc["metadata"]["name"]): doc
        for doc in yaml.safe_load_all(result.stdout) if isinstance(doc, dict)
    }
    deployment = docs[("Deployment", "b2fixture-framework-runs")]
    container = deployment["spec"]["template"]["spec"]["containers"][0]
    assert container["image"].endswith("@" + DIGEST)
    assert container["readinessProbe"]["exec"]["command"] == ["/framework-runs", "probe"]
    assert container["livenessProbe"]["tcpSocket"]["port"] == 8091
    assert deployment["spec"]["template"]["spec"]["automountServiceAccountToken"] is False
    assert container["securityContext"]["readOnlyRootFilesystem"] is True
    volumes = deployment["spec"]["template"]["spec"]["volumes"]
    assert {item["secret"]["secretName"] for item in volumes if "secret" in item} == {
        "runs-auth", "runs-db", "runs-tls",
    }
    assert ("Service", "b2fixture-framework-runs") in docs
    policy = docs[("NetworkPolicy", "b2fixture-framework-runs")]["spec"]
    assert policy["podSelector"]["matchLabels"]["app.kubernetes.io/name"] == "framework-runs"
    assert policy["egress"][1]["to"] == [{"ipBlock": {"cidr": "10.42.0.5/32"}}]
    assert policy["ingress"][0]["ports"] == [{"protocol": "TCP", "port": 8091}]


@pytest.mark.parametrize(
    ("override", "message"),
    [
        (("frameworkWorker.image.digest=latest",), "sha256 image digest"),
        (("frameworkWorker.database.egressCIDRs[0]=0.0.0.0/0",), "exact IPv4 /32"),
        (("frameworkWorker.database.egressCIDRs[0]=::/0",), "exact IPv4 /32"),
        (("frameworkWorker.database.egressCIDRs[0]=010.42.0.5/32",), "exact IPv4 /32"),
        (("frameworkWorker.auth.existingSecret=mycelis-core-core-auth", "frameworkWorker.auth.tokenKey=mycelis-api-key"), "Core API credential"),
        (("coreAuth.existingSecret=shared-core-key", "frameworkWorker.auth.existingSecret=shared-core-key", "frameworkWorker.auth.tokenKey=mycelis-api-key"), "Core API credential"),
    ],
)
def test_worker_chart_rejects_unsafe_enablement(override, message):
    args = tuple(item for value in override for item in ("--set-string", value))
    result = helm_template(*enabled_args(*args))
    assert result.returncode != 0
    assert message in result.stderr
