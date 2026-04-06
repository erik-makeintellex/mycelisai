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
    assert "uv run inv core.package" in workflow
    assert "softprops/action-gh-release" in workflow
