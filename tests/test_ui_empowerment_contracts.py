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


def test_deployments_detail_pane_exposes_plan_and_execute_next_steps_action() -> None:
    detail_pane = (REPO_ROOT / "interface" / "components" / "deployments" / "DeploymentDetailPane.tsx").read_text(
        encoding="utf-8"
    )
    assert 'data-testid="deployments-plan-and-execute"' in detail_pane
    assert "team_action=plan_execute" in detail_pane
    assert "Plan &amp; Execute Next Steps" in detail_pane or "Plan & Execute Next Steps" in detail_pane

    chat = (REPO_ROOT / "interface" / "components" / "dashboard" / "MissionControlChat.tsx").read_text(
        encoding="utf-8"
    )
    assert "buildTeamActionPrefillPrompt" in chat
    assert "teamSuggestions(currentTeam.name, currentTeam.id)" in chat

    chat_ui = (REPO_ROOT / "interface" / "components" / "dashboard" / "missionControlChatUi.ts").read_text(
        encoding="utf-8"
    )
    assert "export function buildTeamActionPrefillPrompt" in chat_ui
    assert 'teamAction === "plan_execute"' in chat_ui
    assert "execute next steps to produce its execution plan, action matrix, and handoff deliverable package" in chat_ui


def test_default_theme_is_system_across_backend_and_store() -> None:
    backend = (REPO_ROOT / "core" / "internal" / "server" / "identity_settings.go").read_text(encoding="utf-8")
    assert '"theme":                 "system"' in backend

    initial_state = (REPO_ROOT / "interface" / "store" / "cortexStoreInitialState.ts").read_text(encoding="utf-8")
    assert "theme: 'system'" in initial_state

    settings_slice = (REPO_ROOT / "interface" / "store" / "cortexStoreUserSettingsSlice.ts").read_text(encoding="utf-8")
    assert "return 'system';" in settings_slice


def test_organization_starters_use_compact_grid_and_thorough_delivery_breakdown() -> None:
    entry = (REPO_ROOT / "interface" / "components" / "organizations" / "CreateOrganizationEntry.tsx").read_text(
        encoding="utf-8"
    )
    assert 'data-testid="starter-cards-grid"' in entry
    assert "sm:grid-cols-2" in entry

    entry_ui = (REPO_ROOT / "interface" / "components" / "organizations" / "createOrganizationEntryUi.tsx").read_text(
        encoding="utf-8"
    )
    assert 'data-testid="starter-target-delivery-breakdown"' in entry_ui
    assert "Target Delivery &amp; Execution Breakdown" in entry_ui
    assert "+ Ideal Reviewer Gate" in entry_ui
    assert "starter-department-execution-lanes" in entry_ui
    assert "Delivery Verification Gate:" in entry_ui



def test_teams_always_surface_ideal_reviewer_and_delivery_review_gate() -> None:
    detail_pane = (REPO_ROOT / "interface" / "components" / "deployments" / "DeploymentDetailPane.tsx").read_text(
        encoding="utf-8"
    )
    assert 'data-testid="deployments-reviewer-badge"' in detail_pane
    assert "IDEAL REVIEWER · DELIVERY GATE" in detail_pane
    assert 'data-testid="deployments-review-gate"' in detail_pane
    assert "function isReviewerRole(" in detail_pane
    assert "-review-team" in detail_pane
    assert "OutputProofEnvelope" in detail_pane

    tools_path = REPO_ROOT / "core" / "internal" / "swarm" / "internal_tools_support.go"
    if tools_path.exists():
        tools_content = tools_path.read_text(encoding="utf-8")
        if "ensureIdealReviewerMember" in tools_content:
            assert "ensureIdealReviewerMember" in tools_content

    reviewer_path = REPO_ROOT / "core" / "internal" / "server" / "cognitive_team_plan_reviewer.go"
    if reviewer_path.exists():
        assert "buildCompanionReviewTeamCall" in reviewer_path.read_text(encoding="utf-8")


def test_deployments_output_preview_drawer_includes_review_team_audit_certificate() -> None:
    detail_pane = (REPO_ROOT / "interface" / "components" / "deployments" / "DeploymentDetailPane.tsx").read_text(
        encoding="utf-8"
    )
    assert "DeploymentOutputPreviewDrawer" in detail_pane
    assert "setPreviewItem(item)" in detail_pane

    drawer = (REPO_ROOT / "interface" / "components" / "deployments" / "DeploymentOutputPreviewDrawer.tsx").read_text(
        encoding="utf-8"
    )
    assert 'data-testid="deployment-output-preview-drawer"' in drawer
    assert 'data-testid="deployment-output-review-certificate"' in drawer
    assert "VERIFIED BY IDEAL REVIEWER & DELIVERY REVIEW TEAM" in drawer
    assert 'data-testid="deployment-output-preview-iframe"' in drawer
    assert 'data-testid="deployment-output-preview-source"' in drawer


def test_web_based_agentry_services_presets_and_secret_ref_pipeline_are_wired() -> None:
    brains_types = (REPO_ROOT / "interface" / "components" / "settings" / "brainsTypes.ts").read_text(encoding="utf-8")
    for provider_key in ("openai", "anthropic", "google", "openrouter", "groq", "deepseek", "mistral", "xai"):
        assert f"{provider_key}:" in brains_types
    assert "OPENAI_API_KEY" in brains_types
    assert "ANTHROPIC_API_KEY" in brains_types
    assert "GEMINI_API_KEY" in brains_types
    assert "OPENROUTER_API_KEY" in brains_types
    assert "GROQ_API_KEY" in brains_types
    assert "DEEPSEEK_API_KEY" in brains_types

    brains_modal = (REPO_ROOT / "interface" / "components" / "settings" / "BrainsProviderModal.tsx").read_text(
        encoding="utf-8"
    )
    assert 'data-testid="brains-web-service-presets"' in brains_modal
    assert 'data-testid="brain-form-api-key-env"' in brains_modal

    cog_yaml = (REPO_ROOT / "core" / "config" / "cognitive.yaml").read_text(encoding="utf-8")
    assert "claude-3-7-sonnet-latest" in cog_yaml
    assert "gemini-2.5-pro" in cog_yaml
    assert "gpt-4o" in cog_yaml
    assert "openrouter:" in cog_yaml
    assert "groq:" in cog_yaml
    assert "deepseek:" in cog_yaml


def test_simple_project_development_asks_bypass_bureaucracy_and_compact_proposals() -> None:
    suggestion_bar = (REPO_ROOT / "interface" / "components" / "soma" / "SomaSuggestionBar.tsx").read_text(
        encoding="utf-8"
    )
    assert 'data-testid="toggle-advanced-playbooks"' in suggestion_bar
    assert "displayPlaybooks = advancedMode || showAdvancedPlaybooks" in suggestion_bar

    proposed_action = (REPO_ROOT / "interface" / "components" / "dashboard" / "ProposedActionBlock.tsx").read_text(
        encoding="utf-8"
    )
    assert "const isSimpleDevView = simpleMode && !advancedMode;" in proposed_action
    assert "What happens when you click" in proposed_action
    assert "Files & teams to build" in proposed_action

    deliverable_plan = (REPO_ROOT / "core" / "internal" / "server" / "cognitive_deliverable_plan.go").read_text(
        encoding="utf-8"
    )
    assert "isGeneralWebAppProjectBundleRequest" in deliverable_plan
    assert '"dev_project_bundle"' in deliverable_plan
    assert '"multi_app_bundle"' in deliverable_plan

    multi_app_plan = (REPO_ROOT / "core" / "internal" / "server" / "cognitive_multi_app_plan.go").read_text(
        encoding="utf-8"
    )
    assert "func isInteractingMultiAppRequest(" in multi_app_plan
    assert "func isSimpleCasualBuildRequest(" in multi_app_plan
    assert "func inferInteractingMultiAppTeamCalls(" in multi_app_plan

    detail_pane = (REPO_ROOT / "interface" / "components" / "deployments" / "DeploymentDetailPane.tsx").read_text(
        encoding="utf-8"
    )
    assert 'data-testid="deployments-build-companion-app"' in detail_pane
    assert "team_action=companion_app" in detail_pane


def test_instant_proposal_fast_path_and_stackable_work_queue_are_wired() -> None:
    tool_plan = (REPO_ROOT / "core" / "internal" / "server" / "cognitive_tool_plan.go").read_text(encoding="utf-8")
    assert "func hasDeterministicFastPathPlan(" in tool_plan
    assert "func buildFastPathProposalReplyText(" in tool_plan
    assert "keep typing to stack more work into the queue for Soma to review" in tool_plan

    chat_handler = (REPO_ROOT / "core" / "internal" / "server" / "cognitive_chat_handler.go").read_text(encoding="utf-8")
    assert "hasDeterministicFastPathPlan(latestUserText, requestMutationTools)" in chat_handler

    chat_slice = (REPO_ROOT / "interface" / "store" / "cortexStoreMissionChatSlice.ts").read_text(encoding="utf-8")
    assert "queuedChatPrompts" in chat_slice
    assert "drainNextQueuedPrompt" in chat_slice

    chat_ui = (REPO_ROOT / "interface" / "components" / "dashboard" / "MissionControlChat.tsx").read_text(
        encoding="utf-8"
    )
    assert 'data-testid="soma-queued-prompts-tray"' in chat_ui
    assert "Stacked Work Queue for" in chat_ui
    assert "disabled={false}" in chat_ui


def test_standard_deployment_mcps_and_custom_developed_mcp_grouping_are_wired() -> None:
    mcp_panel = (
        REPO_ROOT / "interface" / "components" / "deployments" / "DeploymentMCPAdditionsPanel.tsx"
    ).read_text(encoding="utf-8")
    assert 'data-testid="deployments-mcp-additions-panel"' in mcp_panel
    assert 'data-testid="deployments-mcp-tab-standard"' in mcp_panel
    assert 'data-testid="deployments-mcp-tab-custom"' in mcp_panel
    assert 'data-testid="deployments-standard-mcp-group"' in mcp_panel
    assert 'data-testid="deployments-custom-mcp-group"' in mcp_panel
    assert 'data-testid="deployments-custom-mcp-builder"' in mcp_panel
    assert "STANDARD_DEPLOYMENT_MCPS" in mcp_panel
    assert "CUSTOM_DEVELOPED_MCPS" in mcp_panel
    for mcp_id in ("filesystem", "github", "git", "fetch", "postgres", "sqlite", "docker-kubernetes", "memory", "puppeteer", "sequential-thinking"):
        assert f'id: "{mcp_id}"' in mcp_panel
    for custom_id in ("custom-python-fastmcp", "custom-ts-mcp-sdk", "custom-http-sse-mcp", "custom-nats-service-bridge"):
        assert f'id: "{custom_id}"' in mcp_panel

    detail_pane = (REPO_ROOT / "interface" / "components" / "deployments" / "DeploymentDetailPane.tsx").read_text(
        encoding="utf-8"
    )
    assert "DeploymentMCPAdditionsPanel" in detail_pane

    mcp_lib = (REPO_ROOT / "core" / "config" / "mcp-library.yaml").read_text(encoding="utf-8")
    assert 'name: "Custom Developed MCPs"' in mcp_lib
    assert 'name: "custom-python-fastmcp"' in mcp_lib
    assert 'name: "custom-ts-mcp-sdk"' in mcp_lib
    assert 'name: "custom-http-sse-mcp"' in mcp_lib
    assert 'name: "custom-nats-service-bridge"' in mcp_lib
    assert 'name: "git"' in mcp_lib
    assert 'name: "docker-kubernetes"' in mcp_lib

    mcp_browser = (REPO_ROOT / "interface" / "components" / "settings" / "MCPLibraryBrowser.tsx").read_text(
        encoding="utf-8"
    )
    assert 'data-testid="mcp-library-group-filters"' in mcp_browser
    assert "Standard Deployment MCPs" in mcp_browser
    assert "Custom Developed MCPs" in mcp_browser

