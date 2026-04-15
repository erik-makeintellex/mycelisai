from __future__ import annotations

from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]
README = ROOT / "README.md"
TESTING = ROOT / "docs" / "TESTING.md"
OPERATIONS = ROOT / "docs" / "architecture" / "OPERATIONS.md"
DEPLOYMENT_METHODS = ROOT / "docs" / "user" / "deployment-methods.md"
DOCKER_COMPOSE = ROOT / "docker-compose.yml"


def test_compose_testing_contract_points_to_explicit_non_loopback_ai_hosts():
    snippets = [
        (
            README,
            [
                "use a reachable host/IP like `http://192.168.x.x:11434/v1`, not `localhost`",
                "point it at a host-reachable endpoint such as `http://host.docker.internal:11434`",
                "auto-start a WSL-host relay for the AI endpoint when needed",
            ],
        ),
        (
            TESTING,
            [
                "use `MYCELIS_K8S_TEXT_ENDPOINT` and optional `MYCELIS_K8S_MEDIA_ENDPOINT`",
                "explicit reachable AI host instead of a chart-baked or localhost default",
                "keep it container-reachable instead of `localhost`, `127.0.0.1`, or `0.0.0.0`",
                "may relay `MYCELIS_COMPOSE_OLLAMA_HOST` through the WSL host",
            ],
        ),
        (
            OPERATIONS,
            [
                "use explicit reachable AI endpoints for deployed text or media engines instead of localhost assumptions",
                "the Helm chart applies `MYCELIS_K8S_TEXT_ENDPOINT` through provider-specific env overrides",
                "use a reachable Windows IP or hostname such as `http://192.168.x.x:11434/v1`, not `localhost`",
                "can auto-start a WSL-host relay for `MYCELIS_COMPOSE_OLLAMA_HOST`",
            ],
        ),
    ]

    missing: list[str] = []
    for path, required_snippets in snippets:
        text = path.read_text(encoding="utf-8")
        for snippet in required_snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Compose/testing contract is missing explicit non-loopback AI endpoint guidance:\n" + "\n".join(missing)


def test_compose_runtime_maps_ai_host_into_provider_overrides():
    text = DOCKER_COMPOSE.read_text(encoding="utf-8")

    required_snippets = [
        "MYCELIS_PROVIDER_OLLAMA_ENDPOINT: ${MYCELIS_COMPOSE_OLLAMA_HOST:-http://host.docker.internal:11434}/v1",
        "MYCELIS_PROVIDER_LOCAL_OLLAMA_DEV_ENDPOINT: ${MYCELIS_COMPOSE_OLLAMA_HOST:-http://host.docker.internal:11434}/v1",
        'MYCELIS_PROVIDER_LOCAL_OLLAMA_DEV_ENABLED: "true"',
        "MYCELIS_PROVIDER_LOCAL_SOVEREIGN_ENDPOINT: ${MYCELIS_COMPOSE_OLLAMA_HOST:-http://host.docker.internal:11434}/v1",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, "docker-compose.yml is missing provider endpoint overrides:\n" + "\n".join(missing)
    assert "      OLLAMA_HOST:" not in text, "docker-compose.yml should not inject legacy OLLAMA_HOST into Core"


def test_k8s_docs_prefer_k3d_with_kind_fallback():
    snippets = [
        (
            README,
            [
                "prefer `k3d` as the local Kubernetes backend when it is available",
                "MYCELIS_K8S_BACKEND=kind",
            ],
        ),
        (
            TESTING,
            [
                "when the validation target is local Kubernetes, prefer `k3d`",
                "`MYCELIS_K8S_BACKEND=kind`",
            ],
        ),
        (
            OPERATIONS,
            [
                "local Kubernetes now prefers `k3d` when it is installed",
                "MYCELIS_K8S_BACKEND=kind",
            ],
        ),
    ]

    missing: list[str] = []
    for path, required_snippets in snippets:
        text = path.read_text(encoding="utf-8")
        for snippet in required_snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "k3d local-Kubernetes contract is missing from active docs:\n" + "\n".join(missing)


def test_user_docs_explain_deployment_method_selection_by_target_environment():
    text = DEPLOYMENT_METHODS.read_text(encoding="utf-8")

    required_snippets = [
        "Docker Compose",
        "Local Kubernetes With k3d",
        "Enterprise Self-Hosted Kubernetes",
        "Edge Or Small Node Deployments",
        "Developer source mode is not a deployment method",
        "MYCELIS_COMPOSE_OLLAMA_HOST=http://<windows-ai-host>:11434",
        "MYCELIS_K8S_TEXT_ENDPOINT=http://<windows-ai-host>:11434/v1",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, "Deployment method selection doc is missing required guidance:\n" + "\n".join(missing)


def test_k8s_docs_cover_promoted_values_file_contract():
    snippets = [
        (
            README,
            [
                "MYCELIS_K8S_VALUES_FILE",
                "charts/mycelis-core/values-k3d.yaml",
                "charts/mycelis-core/values-enterprise.yaml",
                "charts/mycelis-core/values-enterprise-windows-ai.yaml",
            ],
        ),
        (
            TESTING,
            [
                "MYCELIS_K8S_VALUES_FILE",
                "charts/mycelis-core/values-k3d.yaml",
                "charts/mycelis-core/values-enterprise.yaml",
                "charts/mycelis-core/values-enterprise-windows-ai.yaml",
            ],
        ),
        (
            OPERATIONS,
            [
                "MYCELIS_K8S_VALUES_FILE",
                "values-k3d.yaml",
                "values-enterprise.yaml",
                "values-enterprise-windows-ai.yaml",
            ],
        ),
    ]

    missing: list[str] = []
    for path, required_snippets in snippets:
        text = path.read_text(encoding="utf-8")
        for snippet in required_snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Kubernetes preset-values contract is missing from active docs:\n" + "\n".join(missing)
