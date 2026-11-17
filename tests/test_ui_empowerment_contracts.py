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
