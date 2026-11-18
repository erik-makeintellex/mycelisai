from __future__ import annotations

from pathlib import Path

REPO_ROOT = Path(__file__).resolve().parents[1]


def test_soma_suggestion_bar_includes_orchestration_and_council_playbooks() -> None:
    content = (REPO_ROOT / "interface" / "components" / "soma" / "SomaSuggestionBar.tsx").read_text(encoding="utf-8")
    assert "ORCHESTRATION_PLAYBOOK_STARTERS" in content
    assert "Council Architecture & Security Review" in content
    assert "Multi-Team Mission Blueprint" in content
    assert "Trusted-Memory & Doctrine Audit" in content
    assert "Service & Tool Capability Plan" in content
    assert "orchestration-playbook-starter-" in content


def test_deployments_detail_pane_exposes_soma_team_engagement_and_config_access() -> None:
    content = (REPO_ROOT / "interface" / "components" / "deployments" / "DeploymentDetailPane.tsx").read_text(
        encoding="utf-8"
    )
    assert 'data-testid="deployments-engage-soma"' in content
    assert "/dashboard?team_id=" in content
    assert "Work with Team in Soma" in content
    assert 'data-testid="deployments-view-config"' in content


def test_resources_page_includes_deliverables_tab_for_seamless_switching() -> None:
    content = (REPO_ROOT / "interface" / "app" / "(app)" / "resources" / "page.tsx").read_text(encoding="utf-8")
    assert "deliverables" in content
    assert 'label="Deliverables"' in content
    assert "if (!advancedMode && !tabParam)" in content


def test_soma_suggestion_bar_defines_three_progressive_engagement_tiers() -> None:
    content = (REPO_ROOT / "interface" / "components" / "soma" / "SomaSuggestionBar.tsx").read_text(encoding="utf-8")
    assert "export const DEFAULT_SOMA_SUGGESTIONS" in content
    assert "export const POPULAR_DELIVERABLE_STARTERS" in content
    assert "export const ORCHESTRATION_PLAYBOOK_STARTERS" in content
    assert "Start with Soma" in content
    assert "Popular deliverable starters" in content
    assert "Advanced orchestration & council playbooks" in content


def test_team_generated_content_access_surfaces_are_wired() -> None:
    workbench = (REPO_ROOT / "interface" / "components" / "soma" / "OutputWorkbench.tsx").read_text(encoding="utf-8")
    assert "output-workbench-open-team" in workbench
    assert "output-workbench-open-resources" in workbench
    assert "output-workbench-inline-iframe" not in workbench

    detail_pane = (REPO_ROOT / "interface" / "components" / "deployments" / "DeploymentDetailPane.tsx").read_text(
        encoding="utf-8"
    )
    assert "deployments-open-deliverables" in detail_pane
    assert "deployments-team-outputs" in detail_pane
    assert "deployments-team-channels" in detail_pane
    assert "deployments-chain-output" in detail_pane

    durable_files = (
        REPO_ROOT / "interface" / "components" / "resources" / "DurableFilesOperatorView.tsx"
    ).read_text(encoding="utf-8")
    assert "deliverables-producer-filter" in durable_files


def test_run_chain_node_links_to_team_workspace_and_outputs() -> None:
    content = (REPO_ROOT / "interface" / "components" / "runs" / "RunChainNode.tsx").read_text(encoding="utf-8")
    assert "chain-node-open-team-" in content
    assert "/deployments?team=" in content



