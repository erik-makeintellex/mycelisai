from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


def _read(path: str) -> str:
    return (ROOT / path).read_text(encoding="utf-8")


def test_ci_workflow_runs_token_free_codebase_gates():
    workflow = _read(".github/workflows/ci.yaml")
    operations_doc = _read("docs/architecture/OPERATIONS.md")

    for expected in [
        "Repo Hygiene, Docs, And Ops Tests",
        "Go Core",
        "Interface Unit And Build",
        "Browser Smoke Without Live Agentry",
        "Helm And Kubernetes Standards",
        "uv run inv quality.max-lines --limit 300",
        "go test ./... -count=1 -p 1",
        "uv run inv interface.test",
        "uv run inv interface.build",
        "uv run inv interface.typecheck",
        "uv run inv interface.e2e",
        "browser_spec",
        "e2e/specs/homepage.spec.ts",
        "uv run inv k8s.standards",
        "MYCELIS_SEARCH_PROVIDER: \"disabled\"",
        "workflow_dispatch",
    ]:
        assert expected in workflow

    assert "hosted agentry" in operations_doc
    assert "core-ci.yaml" not in operations_doc
    assert "interface-ci.yaml" not in operations_doc
    assert "e2e-ci.yaml" not in operations_doc
    assert "start Core, then `uv run inv interface.e2e --live-backend`" not in operations_doc


def test_user_workflow_specs_match_current_shared_trial_expectations():
    manual_plan = _read("tests/ui/browser_qa_workflow_variants_reboot.md")
    remote_testing = _read("docs/REMOTE_USER_TESTING.md")

    workflow_specs = {
        "direct": _read("interface/e2e/specs/workflow-output.direct.spec.ts"),
        "compact": _read("interface/e2e/specs/workflow-output.compact-team.spec.ts"),
        "multi_lane": _read("interface/e2e/specs/workflow-output.multi-lane.spec.ts"),
        "reload_review": _read("interface/e2e/specs/workflow-output.reload-review.spec.ts"),
    }

    for spec_path in [
        "interface/e2e/specs/workflow-output.direct.spec.ts",
        "interface/e2e/specs/workflow-output.compact-team.spec.ts",
        "interface/e2e/specs/workflow-output.multi-lane.spec.ts",
        "interface/e2e/specs/workflow-output.reload-review.spec.ts",
    ]:
        assert spec_path in manual_plan

    assert "Soma-first operator workflow" in remote_testing
    assert "deployment-context loading into governed vector-backed stores" in remote_testing
    assert "MCP visibility and recent persisted tool activity" in remote_testing
    assert "safe current actuation proof is governed file output, governed context loading, MCP-backed tool usage, and reviewable audit/activity behavior" in remote_testing

    assert "supported Docker Compose lane" in manual_plan
    assert "Kubernetes is framed as the modular scale-up proof lane" in manual_plan
    assert "Use the supported Docker Compose lane first with an explicit Windows AI endpoint" in workflow_specs["direct"]
    assert "Keep Kubernetes as the modular scale-up proof lane" in workflow_specs["direct"]
    assert "Use the self-hosted Kubernetes lane with an explicit Windows AI endpoint" not in workflow_specs["direct"]

    assert "Create temporary workflow group" in workflow_specs["compact"]
    assert "Archive temporary group" in workflow_specs["compact"]
    assert "Validation checklist" in workflow_specs["compact"]
    assert "Risk review" in workflow_specs["compact"]

    assert "Planning lane package" in workflow_specs["multi_lane"]
    assert "Validation lane checklist" in workflow_specs["multi_lane"]
    assert "Review lane summary" in workflow_specs["multi_lane"]

    assert "Resume the release-readiness work from the retained package" in workflow_specs["reload_review"]
    assert "Already done: planning lane package, validation checklist, and review summary are retained" in workflow_specs["reload_review"]


def test_soma_web_capability_contract_uses_governed_mycelis_search_and_fetch():
    admin = _read("core/config/teams/admin.yaml")
    council = _read("core/config/teams/council.yaml")
    template = _read("core/config/templates/v8-migration-standing-team-bridge.yaml")
    chart_admin = _read("charts/mycelis-core/config/teams/admin.yaml")
    chart_council = _read("charts/mycelis-core/config/teams/council.yaml")
    chart_template = _read("charts/mycelis-core/config/templates/v8-migration-standing-team-bridge.yaml")
    library = _read("core/config/mcp-library.yaml")
    resources_doc = _read("docs/user/resources.md")
    soma_doc = _read("docs/user/soma-chat.md")

    for manifest in (admin, template, chart_admin, chart_template):
        assert "web_search" in manifest
        assert "mcp:fetch/*" in manifest
        assert "mcp:brave-search/*" in manifest
        assert "For search intent, use `web_search` first" in manifest
        assert "Brave Search" not in manifest
        assert "Browse the internet, search the web, or access URLs directly" not in manifest
        assert "ephemeral web code first" not in manifest

    for manifest in (council, chart_council):
        assert "mcp:brave-search/*" in manifest
        assert "governed `web_search` capability for search intent" in manifest
        assert "ephemeral web research/retrieval code paths first" not in manifest

    assert 'name: "brave-search"' in library
    assert 'tool_set: "research"' in library
    assert "BRAVE_API_KEY" in library

    assert "`brave-search` provides governed web search" in resources_doc
    assert "`web_search` for search intent" in soma_doc
    assert "MYCELIS_SEARCH_LOCAL_API_ENDPOINT" in soma_doc


def test_search_fallback_and_provenance_docs_match_runtime_shape():
    api_doc = _read("docs/API_REFERENCE.md")
    resources_doc = _read("docs/user/resources.md")
    soma_doc = _read("docs/user/soma-chat.md")

    assert "data.metadata.semantic_fallback" in api_doc
    assert "`metadata.semantic_fallback`" not in api_doc
    assert "capability_use.reason" in api_doc
    assert "Search source: Local Mycelis context" in resources_doc
    assert "Search source: Local Mycelis context" in soma_doc
    assert "falls back to bounded text search" in resources_doc
    assert "fall back to bounded text search" in soma_doc


def test_mycelis_search_contract_lives_in_user_api_and_capability_docs():
    index = _read("docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md")
    docs_home = _read("docs/README.md")
    manifest = _read("interface/lib/docsManifest.ts")
    resources_doc = _read("docs/user/resources.md")
    soma_doc = _read("docs/user/soma-chat.md")
    api_doc = _read("docs/API_REFERENCE.md")
    capability_doc = _read("docs/architecture-library/V8_CAPABILITY_MANIFEST_AND_RUNTIME_INTEGRATION_STANDARD.md")

    assert "V8 Mycelis Search Capability Delivery Plan" not in index
    assert "V8_MYCELIS_SEARCH_CAPABILITY_DELIVERY_PLAN.md" not in docs_home
    assert "V8_MYCELIS_SEARCH_CAPABILITY_DELIVERY_PLAN.md" not in manifest

    for required in [
        "local_sources",
        "searxng",
        "local_api",
        "brave",
        "disabled",
        "MYCELIS_SEARXNG_ENDPOINT=http://searxng:8080",
        "MYCELIS_SEARCH_LOCAL_API_ENDPOINT",
        "Current Mycelis Search provider posture",
        "capability",
    ]:
        combined = "\n".join([resources_doc, soma_doc, api_doc, capability_doc])
        assert required in combined


def test_active_capability_docs_use_current_governed_resource_contract():
    capability_doc = _read("docs/architecture-library/V8_CAPABILITY_MANIFEST_AND_RUNTIME_INTEGRATION_STANDARD.md")
    resources_doc = _read("docs/user/resources.md")

    required = [
        (capability_doc, "Every execution path in Mycelis must be manageable, inspectable, governed, and reusable."),
        (capability_doc, "manifest"),
        (capability_doc, "approval"),
        (resources_doc, "Connected Tools"),
        (resources_doc, "MCP"),
    ]
    missing = [snippet for text, snippet in required if snippet not in text]
    assert not missing, "Active capability/resource docs are missing governed-resource contract snippets: " + str(missing)


def test_testing_docs_cover_finalization_concretization_contracts():
    testing = _read("docs/TESTING.md")
    ui_contract = _read("docs/architecture-library/V8_UI_TESTING_AGENTRY_PRODUCT_CONTRACT.md")
    full_set = _read("docs/architecture-library/V8_UI_TEAM_FULL_TEST_SET.md")

    required = [
        (testing, "## Finalization Concretization Gate"),
        (testing, "ExecutionContract"),
        (testing, "ProofArtifact"),
        (testing, "CapabilityManifestState"),
        (testing, "UI response states"),
        (testing, "System -> Deployments"),
        (ui_contract, "Evidence for degraded execution must state what succeeded"),
        (ui_contract, "ExecutionContract, ProofArtifact, or UI response-state fields are absent"),
        (full_set, "runtime objects verified: ExecutionContract, ProofArtifact, CapabilityManifestState, UIResponseState"),
    ]
    missing = [snippet for text, snippet in required if snippet not in text]
    assert not missing, "Testing docs are missing concretization proof requirements: " + str(missing)


def test_release_binaries_workflow_uses_core_package_task():
    workflow = _read(".github/workflows/release-binaries.yaml")

    assert "workflow_dispatch" in workflow
    assert "target:" in workflow
    assert "publish_release_assets" in workflow
    assert 'tags: ["v*"]' not in workflow
    assert 'id: release' in workflow
    assert "uv run inv core.package" in workflow
    assert "--version-tag=${{ steps.release.outputs.label }}" in workflow
    assert "softprops/action-gh-release" in workflow


def test_release_workflow_verifies_enterprise_packaging_before_optional_image_publish():
    workflow = _read(".github/workflows/release.yaml")

    assert "verify-enterprise-packaging" in workflow
    assert "package_profile" in workflow
    assert "HELM_VERSION: \"v3.20.2\"" in workflow
    assert "https://get.helm.sh/${tarball}.sha256sum" in workflow
    assert "sha256sum -c" in workflow
    assert "helm repo add bitnami https://charts.bitnami.com/bitnami" in workflow
    assert "helm repo add nats https://nats-io.github.io/k8s/helm/charts/" in workflow
    assert "helm repo update" in workflow
    assert "uv run inv k8s.deploy" in workflow
    assert "--verify-package" in workflow
    assert "values-enterprise.yaml" in workflow
    assert "values-enterprise-windows-ai.yaml" in workflow
    assert "actions/upload-artifact@v7" in workflow
    assert "docker/build-push-action@v7" in workflow
    assert "if: inputs.publish_images == true" in workflow


def test_manual_source_api_proof_workflow_uses_hosted_infra_and_mycelis_api():
    workflow = _read(".github/workflows/source-api-proof.yaml")
    testing = _read("docs/TESTING.md")

    for expected in [
        "workflow_dispatch",
        "proof_mode",
        "pgvector/pgvector:pg16",
        "nats:2-alpine",
        "uv run inv db.migrate",
        "uv run inv core.compile",
        "./bin/server",
        "uv run inv api.delivery-proof --read-only",
        "uv run inv api.delivery-proof",
        "PLAYWRIGHT_ACTIVE_WORK_API_LIVE",
        "PLAYWRIGHT_SKIP_AUTH_SETUP",
        "active-work-api.spec.ts",
    ]:
        assert expected in workflow

    assert "Source API Proof" in testing
    assert "hosted pgvector PostgreSQL/NATS" in testing
