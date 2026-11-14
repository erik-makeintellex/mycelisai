from __future__ import annotations

from pathlib import Path

ROOT = Path(__file__).resolve().parents[1]
README = ROOT / "README.md"
TESTING = ROOT / "docs" / "TESTING.md"
OPERATIONS = ROOT / "docs" / "architecture" / "OPERATIONS.md"
DEPLOYMENT_METHODS = ROOT / "docs" / "user" / "deployment-methods.md"
LICENSING = ROOT / "docs" / "licensing.md"
RESOURCES = ROOT / "docs" / "user" / "resources.md"
CORE_CONCEPTS = ROOT / "docs" / "user" / "core-concepts.md"
GOVERNANCE_TRUST = ROOT / "docs" / "user" / "governance-trust.md"
DOCKER_COMPOSE = ROOT / "docker-compose.yml"
CORE_DOCKERFILE = ROOT / "core" / "Dockerfile"


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


def test_compose_core_image_supports_curated_stdio_mcp_launch():
    dockerfile = CORE_DOCKERFILE.read_text(encoding="utf-8")
    required_dockerfile_snippets = [
        "FROM node:22-bookworm-slim",
        "Curated stdio MCP servers are launched through npm/npx",
    ]
    missing_dockerfile = [snippet for snippet in required_dockerfile_snippets if snippet not in dockerfile]
    assert not missing_dockerfile, "core/Dockerfile is missing stdio MCP runtime support:\n" + "\n".join(missing_dockerfile)

    snippets = [
        (
            README,
            [
                "the supported Core container image includes Node/npm/npx for curated stdio MCP servers",
                "manual `filesystem` library install must be able to launch and bind to the configured `/data/workspace` output block",
            ],
        ),
        (
            OPERATIONS,
            [
                "the compose Core image includes Node/npm/npx so manual curated stdio MCP installs can launch from the shipped container",
                "manual `filesystem` installs from the curated library are runtime-normalized to the configured `MYCELIS_WORKSPACE` root",
            ],
        ),
        (
            TESTING,
            [
                "the deployed Core image can launch the curated `filesystem` stdio MCP server through `npx`",
                "runtime workspace normalization for filesystem installs",
            ],
        ),
    ]

    missing: list[str] = []
    for path, required_snippets in snippets:
        text = path.read_text(encoding="utf-8")
        for snippet in required_snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Curated stdio MCP runtime packaging contract is missing from active docs:\n" + "\n".join(missing)


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
        "Docker Compose when you want rapid local development, same-machine proof, or a quick demo loop",
        "not as the target clustered deployment contract",
        "MYCELIS_COMPOSE_OLLAMA_HOST=http://<windows-ai-host>:11434",
        "MYCELIS_K8S_TEXT_ENDPOINT=http://<windows-ai-host>:11434/v1",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, "Deployment method selection doc is missing required guidance:\n" + "\n".join(missing)


def test_active_docs_cover_supported_user_access_lanes():
    snippets = [
        (
            README,
            [
                "Windows Docker Desktop Compose",
                "Windows + WSL Docker Compose",
                "Kubernetes / Helm clustered deployment",
                "open `http://localhost:3000` from the Windows browser",
                "prove the real ingress/hostname/IP from the operator machine",
            ],
        ),
        (
            TESTING,
            [
                "Windows Docker Desktop Compose with the Windows browser on the same machine for rapid local proof",
                "Kubernetes / Helm clustered deployment reached through the real ingress, remote host, IP, or hostname",
                "the browser opens the UI through the same operator-facing address the delivered environment will actually use",
            ],
        ),
        (
            DEPLOYMENT_METHODS,
            [
                "Supported user access lanes:",
                "Windows Docker Desktop",
                "Windows + WSL Docker",
                "Kubernetes / Helm clustered deployment",
            ],
        ),
    ]

    missing: list[str] = []
    for path, required_snippets in snippets:
        text = path.read_text(encoding="utf-8")
        for snippet in required_snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Supported user access lanes are missing from active docs:\n" + "\n".join(missing)


def test_licensing_docs_define_release_enterprise_and_hosted_admin_boundaries():
    snippets = [
        (
            LICENSING,
            [
                "local named users and manual local role/group administration",
                "optional local break-glass recovery principal",
                "SAML and/or OIDC federation",
                "optional SCIM lifecycle sync",
                "Hosted admin control plane",
                "local break-glass recovery is part of self-hosted recovery posture",
                "promoted enterprise curated set: `fetch`, `github`, `slack`, `postgres`, `brave-search`",
            ],
        ),
        (
            README,
            [
                "hosted admin control plane",
                "full enterprise multi-user IAM, federated SAML/OIDC/SSO, optional lifecycle sync, and delegated enterprise admin/recovery flows",
            ],
        ),
        (
            GOVERNANCE_TRUST,
            [
                "hosted admin control plane",
                "local break-glass recovery remains part of the self-hosted posture",
            ],
        ),
    ]

    missing: list[str] = []
    for path, required_snippets in snippets:
        text = path.read_text(encoding="utf-8")
        for snippet in required_snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Licensing/edition boundary docs are missing required contract snippets:\n" + "\n".join(missing)


def test_connected_tools_docs_no_longer_claim_bootstrap_defaults_for_filesystem_and_fetch():
    resources_text = RESOURCES.read_text(encoding="utf-8")
    core_concepts_text = CORE_CONCEPTS.read_text(encoding="utf-8")

    forbidden_snippets = [
        "`filesystem`: bootstrap default",
        "`fetch`: bootstrap default",
        "`filesystem` and `fetch` are bootstrap defaults",
    ]

    still_present = []
    for snippet in forbidden_snippets:
        if snippet in resources_text:
            still_present.append(f"docs/user/resources.md still contains `{snippet}`")
        if snippet in core_concepts_text:
            still_present.append(f"docs/user/core-concepts.md still contains `{snippet}`")

    assert not still_present, "Connected Tools docs still describe deprecated bootstrap defaults:\n" + "\n".join(still_present)
