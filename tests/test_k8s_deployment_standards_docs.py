from __future__ import annotations

from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
README = ROOT / "README.md"
TESTING = ROOT / "docs" / "TESTING.md"
OPERATIONS = ROOT / "docs" / "architecture" / "OPERATIONS.md"
DEPLOYMENT_METHODS = ROOT / "docs" / "user" / "deployment-methods.md"


def test_k8s_docs_define_clustered_open_standards_as_target_deployment():
    snippets = [
        (
            README,
            [
                "Kubernetes / Helm: target self-hosted and enterprise deployment contract using standard Kubernetes resources",
                "Docker Compose: rapid local development, demo, and same-machine proof runtime; it is not the target clustered deployment contract",
                "uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml",
                "Deployment, Service, ServiceAccount, Secret, ConfigMap, PVC, Ingress, NetworkPolicy",
            ],
        ),
        (
            OPERATIONS,
            [
                "This chart is the target clustered deployment contract for self-hosted and enterprise Kubernetes.",
                "Docker Compose is rapid local development/proof only and must not become the production deployment standard.",
                "uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml",
            ],
        ),
        (
            DEPLOYMENT_METHODS,
            [
                "Use this path first for rapid iteration, not as the target clustered deployment contract.",
                "Open-standard chart gate:",
                "Deployment, Service, ServiceAccount, Secret, ConfigMap, PVC, Ingress, NetworkPolicy",
            ],
        ),
        (
            TESTING,
            [
                "when the validation target is clustered deployment, run `uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml`; Compose remains rapid local development/proof only",
                "clustered Kubernetes proof: `uv run inv k8s.standards --helm --values-file=charts/mycelis-core/values-enterprise.yaml`",
            ],
        ),
    ]

    missing: list[str] = []
    for path, required_snippets in snippets:
        text = path.read_text(encoding="utf-8")
        for snippet in required_snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Clustered Kubernetes open-standards target docs are missing:\n" + "\n".join(missing)
