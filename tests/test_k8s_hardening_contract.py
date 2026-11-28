from pathlib import Path

CHART_DIR = Path(__file__).resolve().parents[1] / "charts" / "mycelis-core"


def test_searxng_chart_has_no_plaintext_secret_or_latest_tag():
    values = (CHART_DIR / "values.yaml").read_text(encoding="utf-8")
    tmpl = (CHART_DIR / "templates" / "searxng-deployment.yaml").read_text(encoding="utf-8")

    assert "mycelis-k8s-searxng-search-secret" not in tmpl
    assert "searxng/searxng:latest" not in tmpl
    assert "searxng/searxng:latest" not in values
    assert "secretKeyRef:" in tmpl
    assert "allowPrivilegeEscalation: false" in tmpl
    assert "readinessProbe:" in tmpl
    assert "livenessProbe:" in tmpl


def test_interface_chart_enforces_security_context_resources_and_probes():
    tmpl = (CHART_DIR / "templates" / "interface-deployment.yaml").read_text(encoding="utf-8")

    assert "serviceAccountName:" in tmpl
    assert "runAsNonRoot: true" in tmpl
    assert "allowPrivilegeEscalation: false" in tmpl
    assert "resources:" in tmpl
    assert "readinessProbe:" in tmpl
    assert "livenessProbe:" in tmpl
