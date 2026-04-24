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
LOCAL_DEV_WORKFLOW = ROOT / "docs" / "LOCAL_DEV_WORKFLOW.md"
REMOTE_USER_TESTING = ROOT / "docs" / "REMOTE_USER_TESTING.md"
API_REFERENCE = ROOT / "docs" / "API_REFERENCE.md"
BACKEND_ARCH = ROOT / "docs" / "architecture" / "BACKEND.md"
DOCKER_COMPOSE = ROOT / "docker-compose.yml"
CORE_DOCKERFILE = ROOT / "core" / "Dockerfile"
V8_DEV_STATE = ROOT / "V8_DEV_STATE.md"


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
                "Linux server/self-hosted release",
                "open `http://localhost:3000` from the Windows browser",
                "open `http://<server-hostname-or-ip>:3000` from the operator machine",
            ],
        ),
        (
            TESTING,
            [
                "Windows Docker Desktop and same-machine WSL-hosted stacks use the Windows browser with `http://localhost:3000` as the first operator path",
                "Linux self-hosted server or cluster reached through the real remote host/IP/hostname",
                "the browser opens the UI through the same operator-facing address the delivered environment will actually use",
            ],
        ),
        (
            DEPLOYMENT_METHODS,
            [
                "Supported user access lanes:",
                "Windows Docker Desktop",
                "Windows + WSL Docker",
                "Linux server/self-hosted release",
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


def test_release_preflight_docs_prefer_lane_preset():
    snippets = [
        (README, ["uv run inv ci.release-preflight --lane=release"]),
        (
            TESTING,
            [
                "uv run inv ci.release-preflight --lane=release",
                "lane presets are `baseline`, `runtime`, `service`, and `release`",
            ],
        ),
        (
            OPERATIONS,
            [
                "--lane=baseline|runtime|service|release",
                "`--lane=release` is the recommended full runtime/operator gate",
            ],
        ),
        (LOCAL_DEV_WORKFLOW, ["uv run inv ci.release-preflight --lane=release"]),
        (REMOTE_USER_TESTING, ["uv run inv ci.release-preflight --lane=release"]),
    ]

    missing: list[str] = []
    for path, required_snippets in snippets:
        text = path.read_text(encoding="utf-8")
        for snippet in required_snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Release-preflight lane contract is missing from active docs:\n" + "\n".join(missing)


def test_remote_user_testing_doc_covers_wsl_deployment_mimic_windows_browser_and_cold_start_notes():
    text = REMOTE_USER_TESTING.read_text(encoding="utf-8")

    required_snippets = [
        "clean WSL deployment-mimic checkout refreshed from git as the validation host",
        "Windows root repo as the dev/staging worktree",
        "verify `http://localhost:3000` from the Windows side with both a simple HTTP probe and a real browser launch",
        "classify it as `cold_start_first_request` instead of a clean first-pass success",
        "do not silently relabel the run as a clean first-pass success",
        "whether the issue is a `cold_start_first_request`, a steady-state regression, or an environment/setup gap",
    ]

    missing = [snippet for snippet in required_snippets if snippet not in text]
    assert not missing, "Remote user testing doc is missing WSL-proof/browser/cold-start guidance:\n" + "\n".join(missing)


def test_active_docs_reference_guarded_wsl_handoff_tasks():
    snippets = [
        (
            README,
            [
                "`uv run inv wsl.status`",
                "`uv run inv wsl.refresh`",
                "`uv run inv wsl.validate`",
                "`uv run inv wsl.cycle`",
            ],
        ),
        (
            TESTING,
            [
                "`uv run inv wsl.status`",
                "`uv run inv wsl.refresh`",
                "`uv run inv wsl.validate`",
                "`uv run inv wsl.cycle`",
            ],
        ),
        (
            OPERATIONS,
            [
                "`wsl.status`, `wsl.refresh`, `wsl.validate`, `wsl.cycle`",
            ],
        ),
        (
            LOCAL_DEV_WORKFLOW,
            [
                "`uv run inv wsl.status`",
                "`uv run inv wsl.refresh --branch <name>`",
                "`uv run inv wsl.validate`",
                "`uv run inv wsl.cycle --branch <name>`",
            ],
        ),
        (
            REMOTE_USER_TESTING,
            [
                "use a clean WSL deployment-mimic checkout refreshed from git as the validation host",
            ],
        ),
    ]

    missing: list[str] = []
    for path, required_snippets in snippets:
        text = path.read_text(encoding="utf-8")
        for snippet in required_snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Guarded WSL handoff/proof tasks are missing from active docs:\n" + "\n".join(missing)


def test_release_proof_sequence_keeps_wsl_validate_before_browser_certification():
    snippets = [
        (
            TESTING,
                [
                    "Release-proof sequencing rule:",
                    "validate WSL git auth repair/report behavior for `wsl.refresh`",
                    "run `uv run inv wsl.validate` from the refreshed WSL proof checkout before trusting browser-gap or certification evidence",
                    "that task intentionally runs `ci.release-preflight --lane=runtime --no-e2e` first",
                    "close focused browser proof gaps next, currently `/runs` workflow depth and guided Soma retry/recovery",
                    "rerun the broader headed Chromium certification pass only after the focused proof-hardening slice is committed and refreshed into WSL",
                ],
            ),
            (
                V8_DEV_STATE,
                [
                    "guarded `uv run inv wsl.validate --lane=release` path from the refreshed `mother-brain` checkout",
                    "maps `--lane=service` and `--lane=release` to `ci.release-preflight --lane=runtime --no-e2e`",
                    "WSL `wsl.validate` from the refreshed proof checkout, deeper `/runs` workflow proof, guided retry/recovery, and broader headed certification from committed state",
                    "run `uv run inv wsl.validate` from the refreshed WSL proof checkout before accepting browser-gap or certification evidence",
                    "live MCP workflow correlation is now green from the refreshed WSL proof checkout",
                    "run `wsl.validate` from the refreshed proof checkout, keep the new `/runs` browser workflow proof green, unskip and keep green the guided Soma retry/recovery lane, then rerun the broader headed Chromium certification pass from committed state",
                ],
            ),
    ]

    missing: list[str] = []
    for path, required_snippets in snippets:
        text = path.read_text(encoding="utf-8")
        for snippet in required_snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Release-proof sequencing docs drifted from the active Slice 4 order:\n" + "\n".join(missing)


def test_windows_edit_wsl_proof_contract_does_not_turn_wsl_into_day_to_day_worktree():
    text = LOCAL_DEV_WORKFLOW.read_text(encoding="utf-8")

    required_snippets = [
        "For day-to-day Windows development, keep the Windows repo as the edit/review/push surface.",
        "Use this as the canonical deployment-mimic proof path when Windows editing is ready for authoritative build, API, UI, runtime, or release-style validation:",
        "keep the WSL `mother-brain` checkout git-backed and disposable for deployment-mimic proof",
        "These tasks keep the WSL proof checkout git-backed and disposable instead of turning it into a second editing worktree.",
    ]
    missing = [snippet for snippet in required_snippets if snippet not in text]

    assert not missing, "Local dev workflow is missing the Windows-edit/WSL-proof contract:\n" + "\n".join(missing)
    assert "prefer a WSL worktree plus the Compose path above" not in text, (
        "docs/LOCAL_DEV_WORKFLOW.md must not make a WSL worktree the day-to-day Windows edit surface"
    )


def test_identity_docs_describe_deploy_owned_people_access_contract():
    snippets = [
        (
            API_REFERENCE,
            [
                "deploy-owned People & Access contract surfaced read-only",
                "PUT ignores/preserves those deploy-owned fields instead of persisting them",
            ],
        ),
        (
            BACKEND_ARCH,
            [
                "deploy-owned People & Access posture surfaced read-only",
                "PUT ignores/preserves those deploy-owned fields instead of persisting them",
            ],
        ),
        (
            LICENSING,
            [
                "deploy-owned edition/auth posture now resolves from env or a deployment-contract file",
                "settings writes do not persist or override",
            ],
        ),
        (
            GOVERNANCE_TRUST,
            [
                "deploy-owned review state, not an ordinary user preference",
            ],
        ),
    ]

    missing: list[str] = []
    for path, required_snippets in snippets:
        text = path.read_text(encoding="utf-8")
        for snippet in required_snippets:
            if snippet not in text:
                missing.append(f"{path.relative_to(ROOT)} missing `{snippet}`")

    assert not missing, "Deploy-owned People & Access contract docs are missing required wording:\n" + "\n".join(missing)
