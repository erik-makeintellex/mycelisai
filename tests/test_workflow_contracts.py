from pathlib import Path


ROOT = Path(__file__).resolve().parents[1]


def _read(path: str) -> str:
    return (ROOT / path).read_text(encoding="utf-8")


def test_core_ci_watches_uv_lock():
    workflow = _read(".github/workflows/core-ci.yaml")

    assert '- "uv.lock"' in workflow


def test_frontend_workflows_watch_root_npmrc_and_uv_lock():
    interface_workflow = _read(".github/workflows/interface-ci.yaml")
    e2e_workflow = _read(".github/workflows/e2e-ci.yaml")

    for workflow in (interface_workflow, e2e_workflow):
        assert '- ".npmrc"' in workflow
        assert '- "uv.lock"' in workflow


def test_e2e_workflow_contract_matches_stable_matrix():
    workflow = _read(".github/workflows/e2e-ci.yaml")
    operations_doc = _read("docs/architecture/OPERATIONS.md")

    assert '- "core/**"' not in workflow
    assert "uv run inv interface.e2e" in workflow
    assert (
        "then the stable invoke-managed Chromium/Firefox/WebKit + mobile smoke browser matrix via `uv run inv interface.e2e`"
        in operations_doc
    )
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


def test_soma_web_capability_contract_uses_governed_mcp_search_and_fetch():
    admin = _read("core/config/teams/admin.yaml")
    council = _read("core/config/teams/council.yaml")
    template = _read("core/config/templates/v8-migration-standing-team-bridge.yaml")
    library = _read("core/config/mcp-library.yaml")
    resources_doc = _read("docs/user/resources.md")
    soma_doc = _read("docs/user/soma-chat.md")

    for manifest in (admin, template):
        assert "mcp:fetch/*" in manifest
        assert "mcp:brave-search/*" in manifest
        assert "BRAVE_API_KEY" in manifest
        assert "You can perform web search or URL\n      retrieval only through installed governed tools" in manifest or "You can perform web search or URL\n          retrieval only through installed governed tools" in manifest
        assert "Browse the internet, search the web, or access URLs directly" not in manifest

    assert "mcp:brave-search/*" in council
    assert "Prefer `brave-search` for search and `fetch` for supplied" in council

    assert 'name: "brave-search"' in library
    assert 'tool_set: "research"' in library
    assert "BRAVE_API_KEY" in library

    assert "`brave-search` provides governed web search" in resources_doc
    assert "`brave-search` for governed web search and `fetch` for explicit URL retrieval" in soma_doc


def test_mycelis_search_delivery_plan_assigns_teams_and_testing_gates():
    plan = _read("docs/architecture-library/V8_MYCELIS_SEARCH_CAPABILITY_DELIVERY_PLAN.md")
    index = _read("docs/architecture-library/ARCHITECTURE_LIBRARY_INDEX.md")
    docs_home = _read("docs/README.md")
    manifest = _read("interface/lib/docsManifest.ts")
    state = _read("V8_DEV_STATE.md")

    required_teams = [
        "Architecture Lead",
        "Runtime Development",
        "Data/Memory Development",
        "Interface Development",
        "Ops/Runtime Delivery",
        "Validation",
    ]
    for team in required_teams:
        assert team in plan

    for required in [
        "Soma -> Mycelis Search API -> local_sources | searxng | brave | disabled",
        "MYCELIS_SEARCH_PROVIDER=disabled|local_sources|searxng|brave",
        "MYCELIS_SEARXNG_ENDPOINT=http://searxng:8080",
        "asking \"can you search the web?\" returns capability status, not a blanket no",
        "local-source search works without Brave or any hosted token",
    ]:
        assert required in plan

    canonical_path = "docs/architecture-library/V8_MYCELIS_SEARCH_CAPABILITY_DELIVERY_PLAN.md"
    assert "V8 Mycelis Search Capability Delivery Plan" in index
    assert "V8_MYCELIS_SEARCH_CAPABILITY_DELIVERY_PLAN.md" in docs_home
    assert canonical_path in manifest
    assert canonical_path in state


def test_active_architecture_docs_use_managed_interface_build_command():
    mcp_doc = _read("docs/architecture/MCP_SERVICE_CONFIGURATION_LOCAL_FIRST_V7.md")
    action_doc = _read("docs/architecture/UNIVERSAL_ACTION_INTERFACE_V7.md")

    for text in (mcp_doc, action_doc):
        assert "uv run inv interface.build" in text
        assert "cd interface && npm run build" not in text


def test_release_binaries_workflow_uses_core_package_task():
    workflow = _read(".github/workflows/release-binaries.yaml")

    assert "workflow_dispatch" in workflow
    assert 'tags: ["v*"]' in workflow
    assert 'id: release' in workflow
    assert "uv run inv core.package" in workflow
    assert "--version-tag=${{ steps.release.outputs.label }}" in workflow
    assert "softprops/action-gh-release" in workflow


def test_release_workflow_verifies_enterprise_packaging_before_optional_image_publish():
    workflow = _read(".github/workflows/release.yaml")

    assert "verify-enterprise-packaging" in workflow
    assert "azure/setup-helm@v4" in workflow
    assert "uv run inv k8s.deploy" in workflow
    assert "--verify-package" in workflow
    assert "values-enterprise.yaml" in workflow
    assert "values-enterprise-windows-ai.yaml" in workflow
    assert "actions/upload-artifact@v4" in workflow
    assert "if: inputs.publish_images == true" in workflow
