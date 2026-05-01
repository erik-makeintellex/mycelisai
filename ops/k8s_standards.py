from __future__ import annotations

import shlex
from pathlib import Path

from invoke import task

from .config import NAMESPACE, ROOT_DIR, is_windows
from .packaging import resolve_repo_path


K8S_OPEN_STANDARD_REQUIRED_FILES = (
    "Chart.yaml",
    "values.yaml",
    "templates/deployment.yaml",
    "templates/service.yaml",
    "templates/serviceaccount.yaml",
    "templates/secrets.yaml",
    "templates/configmap-config.yaml",
    "templates/data-pvc.yaml",
    "templates/ingress.yaml",
    "templates/network-policy.yaml",
)


K8S_OPEN_STANDARD_REQUIRED_SNIPPETS = {
    "values.yaml": (
        "serviceAccount:",
        "imagePullSecrets:",
        "ingress:",
        "networkPolicy:",
        "outputBlock:",
        "mode: cluster_generated",
        "coreAuth:",
        "existingSecret:",
        "probes:",
        "startup:",
        "readiness:",
        "liveness:",
        "podSecurityContext:",
        "securityContext:",
        "seccompProfile:",
        "readOnlyRootFilesystem: true",
        "capabilities:",
        'drop: ["ALL"]',
        "storageClassName:",
        "search:",
    ),
    "templates/deployment.yaml": (
        "apiVersion: apps/v1",
        "kind: Deployment",
        "serviceAccountName:",
        "imagePullSecrets:",
        "startupProbe:",
        "readinessProbe:",
        "livenessProbe:",
        "securityContext:",
        "volumeMounts:",
        "{{- if .Values.env }}",
        "{{- toYaml .Values.env | nindent 12 }}",
    ),
    "templates/service.yaml": (
        "apiVersion: v1",
        "kind: Service",
        "type: {{ .Values.service.type }}",
    ),
    "templates/serviceaccount.yaml": (
        "apiVersion: v1",
        "kind: ServiceAccount",
    ),
    "templates/secrets.yaml": (
        "apiVersion: v1",
        "kind: Secret",
    ),
    "templates/configmap-config.yaml": (
        "apiVersion: v1",
        "kind: ConfigMap",
    ),
    "templates/data-pvc.yaml": (
        "apiVersion: v1",
        "kind: PersistentVolumeClaim",
    ),
    "templates/ingress.yaml": (
        "apiVersion: networking.k8s.io/v1",
        "kind: Ingress",
    ),
    "templates/network-policy.yaml": (
        "apiVersion: networking.k8s.io/v1",
        "kind: NetworkPolicy",
    ),
}


def _chart_dir() -> Path:
    return ROOT_DIR / "charts" / "mycelis-core"


def _shell_quote(value: str | Path) -> str:
    text = str(value)
    if is_windows():
        return '"' + text.replace('"', '\\"') + '"'
    return shlex.quote(text)


def _resolve_k8s_values_file(explicit_path: str = "") -> Path | None:
    raw_value = explicit_path.strip() or ""
    if not raw_value:
        return None

    resolved = resolve_repo_path(raw_value, root_dir=ROOT_DIR)
    if not resolved.exists():
        raise SystemExit(f"MYCELIS_K8S_VALUES_FILE does not exist: {resolved}")
    return resolved


def _run_k8s_open_standard_static_checks() -> None:
    chart_dir = _chart_dir()
    missing_files = [
        relative_path
        for relative_path in K8S_OPEN_STANDARD_REQUIRED_FILES
        if not (chart_dir / relative_path).exists()
    ]
    if missing_files:
        raise SystemExit(
            "Kubernetes standards check failed: missing chart files:\n  - "
            + "\n  - ".join(missing_files)
        )

    missing_snippets: list[str] = []
    for relative_path, snippets in K8S_OPEN_STANDARD_REQUIRED_SNIPPETS.items():
        text = (chart_dir / relative_path).read_text(encoding="utf-8")
        for snippet in snippets:
            if snippet not in text:
                missing_snippets.append(f"{relative_path}: {snippet}")

    values_text = (chart_dir / "values.yaml").read_text(encoding="utf-8")
    if "tag: latest" in values_text:
        missing_snippets.append("values.yaml: chart must not default image.tag to latest")
    if "hostPath:" in values_text and "mode: cluster_generated" not in values_text:
        missing_snippets.append("values.yaml: hostPath must not be the default clustered output mode")

    if missing_snippets:
        raise SystemExit(
            "Kubernetes standards check failed: chart is missing required open-standard surfaces:\n  - "
            + "\n  - ".join(missing_snippets)
        )


@task(
    help={
        "values_file": "Optional values file for Helm lint/template. Defaults to values-enterprise.yaml when present, otherwise values.yaml.",
        "helm": "Also run offline Helm lint and template checks against vendored chart dependencies. Static chart standards checks always run.",
    }
)
def standards(c, values_file="", helm=False):
    """Verify the Kubernetes/Helm chart keeps an open-standard clustered deployment contract."""
    print("=== Kubernetes Open Standards Gate ===")
    print("Target posture: clustered Kubernetes via Helm; Compose is rapid local development only.")
    _run_k8s_open_standard_static_checks()
    print("Static chart contract: OK")

    if not helm:
        print("Helm execution: skipped (pass --helm to run lint/template checks with vendored chart dependencies).")
        return

    resolved_values_file = _resolve_k8s_values_file(values_file)
    if not resolved_values_file:
        enterprise_values = _chart_dir() / "values-enterprise.yaml"
        resolved_values_file = enterprise_values if enterprise_values.exists() else _chart_dir() / "values.yaml"

    quoted_chart_dir = _shell_quote(_chart_dir())
    quoted_values_file = _shell_quote(resolved_values_file)
    print(f"Helm values file: {resolved_values_file}")
    c.run(f"helm lint {quoted_chart_dir} --values {quoted_values_file}")
    c.run(f"helm template mycelis-core {quoted_chart_dir} --namespace {NAMESPACE} --values {quoted_values_file}", hide=True)
    print("Helm lint/template: OK")
