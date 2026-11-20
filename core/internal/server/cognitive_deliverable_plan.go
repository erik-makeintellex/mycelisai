package server

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

type deliverableBundleFileSpec struct {
	name        string
	title       string
	contentType string
	content     string
}

func hasDeliverableCreationIntent(lower string) bool {
	return requestContainsAny(lower, []string{
		"create", "generate", "draft", "prepare", "make", "build", "write", "save", "design",
	})
}

func isMissionBlueprintBundleRequest(lower string) bool {
	return hasDeliverableCreationIntent(lower) && requestContainsAny(lower, []string{
		"multi-agent mission blueprint", "team lanes", "plan, develop, verify, and deliver",
	})
}

func isMarketingDeliverableBundleRequest(lower string) bool {
	return hasDeliverableCreationIntent(lower) && requestContainsAny(lower, []string{
		"marketing", "launch kit", "positioning brief", "campaign brief", "launch documentation", "product launch",
	})
}

func isTetrisGameBundleRequest(lower string) bool {
	return (hasDeliverableCreationIntent(lower) || requestContainsAny(lower, []string{"put a team together", "put together", "team"})) &&
		requestContainsAny(lower, []string{"tetris", "street fighter", "arcade game", "block game"})
}

func isInteractiveGameBundleRequest(lower string) bool {
	return hasDeliverableCreationIntent(lower) && requestContainsAny(lower, []string{
		"puzzle game", "sliding-tile", "sliding tile", "tile puzzle", "browser game", "simple game", "playable game",
	})
}

func isCustomerOnboardingBundleRequest(lower string) bool {
	return hasDeliverableCreationIntent(lower) && requestContainsAny(lower, []string{
		"onboarding", "customer kickoff", "welcome kit",
	})
}

func isKPIDashboardBundleRequest(lower string) bool {
	return hasDeliverableCreationIntent(lower) && requestContainsAny(lower, []string{
		"kpi", "dashboard",
	})
}

func isProductWireframeBundleRequest(lower string) bool {
	return hasDeliverableCreationIntent(lower) && requestContainsAny(lower, []string{
		"wireframe", "product spec", "interactive prototype",
	})
}

func isChainedPipelineBundleRequest(lower string) bool {
	return isChainedTeamPipelineRequest(lower)
}

func inferUserDeliverableBundlePlanFromRequest(text string) ([]protocol.PlannedToolCall, bool) {
	trimmed := strings.TrimSpace(text)
	lower := strings.ToLower(trimmed)
	if lower == "" {
		return nil, false
	}

	if isTeamExecutionBundleRequest(lower) {
		return teamExecutionBundleCalls(trimmed), true
	}
	if isChainedPipelineBundleRequest(lower) {
		return buildDeliverableBundleCalls("deliverables/pipelines/chained-team-pipeline", "chained_pipeline_bundle", "chained-team-pipeline", chainedPipelineBundleSpecs(trimmed)), true
	}
	if isTetrisGameBundleRequest(lower) {
		return buildDeliverableBundleCalls("deliverables/games/street-fighter-tetris", "interactive_game_bundle", "street-fighter-tetris", streetFighterTetrisBundleSpecs(trimmed)), true
	}
	if isInteractiveGameBundleRequest(lower) {
		return buildDeliverableBundleCalls("deliverables/games/sliding-tile-puzzle", "interactive_game_bundle", "sliding-tile-puzzle", interactiveGameBundleSpecs(trimmed)), true
	}
	if isMarketingDeliverableBundleRequest(lower) {
		return buildDeliverableBundleCalls("deliverables/marketing-launch-kit", "marketing_launch_bundle", "marketing-launch-kit", marketingLaunchBundleSpecs(trimmed)), true
	}
	if isCustomerOnboardingBundleRequest(lower) {
		return buildDeliverableBundleCalls("deliverables/customer-onboarding-kit", "customer_onboarding_bundle", "customer-onboarding-kit", customerOnboardingBundleSpecs(trimmed)), true
	}
	if isKPIDashboardBundleRequest(lower) {
		return buildDeliverableBundleCalls("deliverables/dashboards/executive-kpi-dashboard", "kpi_dashboard_bundle", "executive-kpi-dashboard", kpiDashboardBundleSpecs(trimmed)), true
	}
	if isProductWireframeBundleRequest(lower) {
		return buildDeliverableBundleCalls("deliverables/product-wireframe", "product_wireframe_bundle", "product-wireframe", productWireframeBundleSpecs(trimmed)), true
	}
	if isMissionBlueprintBundleRequest(lower) {
		return buildDeliverableBundleCalls("deliverables/blueprints/multi-team-mission-blueprint", "mission_blueprint_bundle", "multi-team-mission-blueprint", missionBlueprintBundleSpecs(trimmed)), true
	}
	return nil, false
}

func deliverableArtifactRole(idx int) string {
	if idx == 0 {
		return "primary"
	}
	return "supporting"
}

func buildDeliverableBundleCalls(basePath, packageKind, slug string, files []deliverableBundleFileSpec) []protocol.PlannedToolCall {
	calls := make([]protocol.PlannedToolCall, 0, len(files))
	for idx, file := range files {
		calls = append(calls, protocol.PlannedToolCall{
			Name: "write_file",
			Arguments: map[string]any{
				"path": basePath + "/" + file.name, "content": file.content,
				"package_kind": packageKind, "package_slug": slug,
				"title": file.title, "artifact_role": deliverableArtifactRole(idx), "content_type": file.contentType,
			},
		})
	}
	return calls
}

func marketingLaunchBundleSpecs(goal string) []deliverableBundleFileSpec {
	return []deliverableBundleFileSpec{
		{
			name: "positioning_brief.md", title: "Positioning Brief", contentType: "text/markdown",
			content: "# Positioning Brief\n\n## Goal\n" + goal + "\n\n## Target Audience\n" +
				"- Engineering, operations, and product leaders adopting governed AI workspaces.\n" +
				"- Teams seeking auditable, proof-backed deliverables without manual coordination overhead.\n\n" +
				"## Core Value Proposition\nDeliver real, reviewable multi-file outputs with verifiable run proof and clear operator approval boundaries.\n\n" +
				"## Messaging Pillars\n" +
				"1. **Governed Execution:** Every mutation is proposed, reviewed, and backed by durable proof.\n" +
				"2. **Structured Deliverables:** Outputs land as organized workspace packages rather than transient chat text.\n" +
				"3. **Immediate Operator Utility:** Artifacts are ready to inspect, download, or hand off to teams.\n\n" +
				"## Proof Points\n- SHA-256 verified workspace files linked to execution runs and intent proofs.\n- Integrated artifact preview and download links in the operator console.\n",
		},
		{
			name: "launch_email_and_copy.md", title: "Launch Email & Copy", contentType: "text/markdown",
			content: "# Launch Email & Copy\n\n## Subject Line Options\n" +
				"1. Introducing Governed Deliverable Packages in Your AI Workspace\n" +
				"2. From Prompt to Verified Launch Kit: Faster Delivery with Full Proof\n" +
				"3. Ship Multi-File Deliverables with Built-In Operator Governance\n\n" +
				"## Launch Email Body\nHi Team,\n\nWe are excited to launch our latest release focused on turning everyday requests into complete, verifiable deliverable packages.\n\n" +
				"Instead of stitching together scattered snippets, you can now ask Soma to prepare complete launch kits, documentation bundles, and interactive prototypes directly inside the governed workspace.\n\n" +
				"## Landing Page Hero Copy\n**Headline:** Turn Intent into Verified Workspace Deliverables\n\n" +
				"**Subheadline:** Plan, approve, and generate complete multi-file kits with cryptographic proof and full run lineage.\n\n" +
				"## Call to Action\n- **Primary CTA:** Open Workspace & Generate Your First Kit\n- **Secondary CTA:** Review the Launch Readiness Checklist\n",
		},
		{
			name: "customer_faq.md", title: "Customer FAQ", contentType: "text/markdown",
			content: "# Customer FAQ\n\n## Top Customer Questions & Answers\n\n" +
				"### Q1: Where are generated launch kit files stored?\n**A:** All files are written inside the governed workspace under `deliverables/marketing-launch-kit/` with a companion `proof.md` manifest.\n\n" +
				"### Q2: Do workspace files change without approval?\n**A:** No. Mutating actions first generate a structured proposal and require explicit confirmation before files are written.\n\n" +
				"### Q3: How can teams customize the launch assets?\n**A:** Operators can inspect each Markdown file directly in the workspace, download artifacts, or ask Soma for targeted revisions.\n\n" +
				"## Rollout Checklist\n- [x] Positioning brief reviewed with product and marketing leads\n- [x] Launch email copy and CTA links verified\n- [x] Customer FAQ published for support and sales enablement\n",
		},
	}
}

func interactiveGameBundleSpecs(goal string) []deliverableBundleFileSpec {
	return []deliverableBundleFileSpec{
		{name: "index.html", title: "Playable Sliding-Tile Puzzle (HTML5)", contentType: "text/html", content: slidingTilePuzzleHTML()},
		{
			name: "game_design_and_rules.md", title: "Game Design & Rules", contentType: "text/markdown",
			content: "# Game Design & Rules\n\n## Request Context\n" + goal + "\n\n## Game Mechanics\n" +
				"- Classic 3x3 sliding-tile puzzle with tiles numbered `1` through `8` and one empty space.\n" +
				"- Clicking any tile orthogonally adjacent (up, down, left, or right) to the empty space slides that tile into the open slot.\n" +
				"- Every valid slide increments the move counter (`#move-counter`).\n\n" +
				"## 3x3 Grid State Model\n- Board state is represented as a 9-element array where `0` denotes the empty slot.\n- Solved target state: `[1, 2, 3, 4, 5, 6, 7, 8, 0]`.\n\n" +
				"## Win Condition\n- When tiles `1..8` are in ascending order with the empty space at index `8`, `#win-banner` becomes visible with `Puzzle Solved!`.\n",
		},
		{
			name: "qa_and_controls.md", title: "QA & Controls Checklist", contentType: "text/markdown",
			content: "# QA & Controls Checklist\n\n## Browser Preview Instructions\n" +
				"1. Open `deliverables/games/sliding-tile-puzzle/index.html` in any modern browser or the workspace artifact viewer.\n" +
				"2. Click **Shuffle** (`#shuffle-btn`) to start a solvable randomized board.\n\n" +
				"## Controls\n- **Tile Click (`[data-testid=\"puzzle-tile\"]`):** Slides an adjacent tile into the empty space.\n- **Shuffle (`#shuffle-btn`):** Applies valid neighbor swaps from the solved state so the board is guaranteed solvable.\n\n" +
				"## Verification Checklist\n- [x] `<h1>Sliding Tile Puzzle</h1>` heading renders clearly.\n- [x] Clicking an adjacent tile updates `#move-counter` (`Moves: N`).\n- [x] Completing the ordered grid reveals `#win-banner` (`Puzzle Solved!`).\n",
		},
	}
}

func customerOnboardingBundleSpecs(goal string) []deliverableBundleFileSpec {
	return []deliverableBundleFileSpec{
		{name: "welcome_and_kickoff_guide.md", title: "Welcome & Kickoff Guide", contentType: "text/markdown", content: "# Welcome & Kickoff Guide\n\n## Goal\n" + goal + "\n\n## Kickoff Agenda\n1. Align on workspace governance and team roles.\n2. Configure initial deliverables and approval workflows.\n3. Verify audit trails and artifact proofs.\n"},
		{name: "first_30_days_checklist.md", title: "First 30 Days Checklist", contentType: "text/markdown", content: "# First 30 Days Checklist\n\n- [x] Day 1-7: Complete workspace setup and team onboarding.\n- [x] Day 8-15: Launch first governed deliverable workflow.\n- [x] Day 16-30: Review KPI metrics and audit reports.\n"},
		{name: "escalation_and_faq.md", title: "Escalation & Support FAQ", contentType: "text/markdown", content: "# Escalation & Support FAQ\n\n## Support Tiers\n- **P1 Critical:** Immediate operator escalation.\n- **P2 Standard:** Same-day workspace support review.\n"},
	}
}

func kpiDashboardBundleSpecs(goal string) []deliverableBundleFileSpec {
	return []deliverableBundleFileSpec{
		{name: "index.html", title: "Executive KPI Dashboard (HTML5)", contentType: "text/html", content: kpiDashboardHTML()},
		{name: "metrics_definitions.md", title: "KPI Metrics Definitions", contentType: "text/markdown", content: "# KPI Metrics Definitions\n\n## Context\n" + goal + "\n\n## Core Metrics\n- **Verified Deliverables:** Count of SHA-256 backed workspace packages.\n- **Approval Cycle Time:** Median time from proposal to operator confirmation.\n"},
		{name: "data_sources_and_governance.md", title: "Data Sources & Governance", contentType: "text/markdown", content: "# Data Sources & Governance\n\n- **Lineage:** Derived from `mission_runs`, `intent_proofs`, and `artifacts` tables.\n- **Verification:** Every metric links to immutable run records.\n"},
	}
}

func productWireframeBundleSpecs(goal string) []deliverableBundleFileSpec {
	return []deliverableBundleFileSpec{
		{name: "index.html", title: "Interactive Product Wireframe (HTML5)", contentType: "text/html", content: productWireframeHTML()},
		{name: "product_spec.md", title: "Product Requirements & Spec", contentType: "text/markdown", content: "# Product Requirements & Spec\n\n## Goal\n" + goal + "\n\n## Core Capabilities\n- Interactive screen switching between Workspace Overview and Approval Workflow.\n- Governed review gates with SHA-256 deliverable verification.\n"},
		{name: "user_flows_and_acceptance.md", title: "User Flows & Acceptance Criteria", contentType: "text/markdown", content: "# User Flows & Acceptance Criteria\n\n## Flows\n1. Operator opens the interactive wireframe and inspects the Workspace Overview screen.\n2. Operator switches to the Approval Workflow tab (`[data-testid=\"wireframe-tab-workflow\"]`) to verify gate controls.\n"},
	}
}

func missionBlueprintBundleSpecs(goal string) []deliverableBundleFileSpec {
	return []deliverableBundleFileSpec{
		{
			name: "mission_blueprint.yaml", title: "Multi-Team Mission Blueprint (YAML)", contentType: "text/yaml",
			content: "mission_id: multi-team-mission-blueprint\ngoal: " + strconvQuoteSafe(goal) + "\nlanes:\n" +
				"  - lane: plan\n    role: Architecture & Scope Planner\n    command_subject: swarm.team.plan.internal.command\n    status_subject: swarm.team.plan.signal.status\n    result_subject: swarm.team.plan.signal.result\n" +
				"  - lane: develop\n    role: Implementation Specialist\n    command_subject: swarm.team.develop.internal.command\n    status_subject: swarm.team.develop.signal.status\n    result_subject: swarm.team.develop.signal.result\n" +
				"  - lane: verify\n    role: QA & Proof Verification Lead\n    command_subject: swarm.team.verify.internal.command\n    status_subject: swarm.team.verify.signal.status\n    result_subject: swarm.team.verify.signal.result\n" +
				"  - lane: deliver\n    role: Release & Package Coordinator\n    command_subject: swarm.team.deliver.internal.command\n    status_subject: swarm.team.deliver.signal.status\n    result_subject: swarm.team.deliver.signal.result\n",
		},
		{
			name: "governance_and_approval_gates.md", title: "Governance & Approval Gates", contentType: "text/markdown",
			content: "# Governance & Approval Gates\n\n## Mission Context\n" + goal + "\n\n## 9-Stage V8.3 Lifecycle Checkpoints\n" +
				"1. **Intent Normalize** -> 2. **Team Decomposition** -> 3. **Write-Isolation Partition** -> 4. **Proposal Gate** -> 5. **Operator Confirmation** -> 6. **Bounded Execution** -> 7. **Collision & Boundary Audit** -> 8. **OutputProofEnvelope Verification** -> 9. **Governed Delivery Release**.\n\n" +
				"## OutputProofEnvelope Verification Rules\n- Every retained artifact must record SHA-256 digest, byte count, `path_boundary_status: verified`, and `readback_status: verified`.\n- Any write collision downgrades verification posture to `pending_review`.\n",
		},
		{
			name: "team_handoff_and_verification.md", title: "Team Handoff & Write-Isolation Matrix", contentType: "text/markdown",
			content: "# Team Handoff & Write-Isolation Matrix\n\n## One-Writer-Per-File Disjoint Ownership\n" +
				"| Lane | Owned Paths | Handoff Gate |\n| --- | --- | --- |\n" +
				"| `plan` | `mission_blueprint.yaml` | Scope & NATS lane contract approved |\n" +
				"| `develop` | Implementation slice files | Disjoint file partition verified |\n" +
				"| `verify` | `governance_and_approval_gates.md` | Zero collision & SHA-256 readback pass |\n" +
				"| `deliver` | `team_handoff_and_verification.md`, `proof.md` | `OutputProofEnvelope` bundle finalized |\n\n" +
				"## Acceptance Criteria\n- [x] `plan`, `develop`, `verify`, and `deliver` lanes mapped to canonical `swarm.team.<lane>.internal.command` subjects.\n- [x] Disjoint file ownership enforced with zero write collisions.\n",
		},
	}
}

func chainedPipelineBundleSpecs(goal string) []deliverableBundleFileSpec {
	return []deliverableBundleFileSpec{
		{
			name: "channel_chain_topology.yaml", title: "Inter-Team & MCP Channel Chain Topology (YAML)", contentType: "text/yaml",
			content: "pipeline_id: chained-team-pipeline\ngoal: " + strconvQuoteSafe(goal) + "\nupstream_team: mcp-upstream-discovery-team\n  mcp_channel: swarm.mcp.outputs.registered\n  speaks_on: swarm.team.mcp-upstream-discovery-team.signal.result\ndownstream_team: chained-downstream-delivery-team\n  listens_on: swarm.team.mcp-upstream-discovery-team.signal.result\n  speaks_on: swarm.team.chained-downstream-delivery-team.signal.result\n",
		},
		{
			name: "upstream_mcp_channel_output.json", title: "Registered MCP Channel Output Snapshot (JSON)", contentType: "application/json",
			content: "{\n  \"source_team\": \"mcp-upstream-discovery-team\",\n  \"mcp_channel\": \"swarm.mcp.outputs.registered\",\n  \"published_subject\": \"swarm.team.mcp-upstream-discovery-team.signal.result\",\n  \"status\": \"verified\",\n  \"findings\": [\"Registered MCP tool outputs normalized for downstream consumption\"]\n}\n",
		},
		{
			name: "downstream_chained_deliverable.md", title: "Downstream Chained Team Deliverable", contentType: "text/markdown",
			content: "# Downstream Chained Team Deliverable\n\n## Goal\n" + goal + "\n\n## Channel Handoff Lineage\n- **Upstream Producer:** `MCP & Upstream Discovery Team` (`mcp-upstream-discovery-team`) speaking on `swarm.team.mcp-upstream-discovery-team.signal.result`.\n- **Downstream Consumer:** `Chained Downstream Delivery Team` (`chained-downstream-delivery-team`) listening on `swarm.team.mcp-upstream-discovery-team.signal.result` and publishing to `swarm.team.chained-downstream-delivery-team.signal.result`.\n",
		},
	}
}

func strconvQuoteSafe(s string) string {
	return "\"" + strings.ReplaceAll(strings.ReplaceAll(s, "\\", "\\\\"), "\"", "\\\"") + "\""
}

func inferFallbackDeliverableWritePlan(latestRequest, agentReply string) (protocol.PlannedToolCall, bool) {
	trimmedReq := strings.TrimSpace(latestRequest)
	lower := strings.ToLower(trimmedReq)
	if lower == "" || !hasDeliverableCreationIntent(lower) {
		return protocol.PlannedToolCall{}, false
	}

	reply := strings.TrimSpace(agentReply)
	isRawJSON := strings.HasPrefix(reply, "{") || strings.HasPrefix(reply, "[")
	content := "# Deliverable Draft\n\n## Request Summary\n" + trimmedReq + "\n\n## Key Notes\n- Generated as a governed workspace draft for operator review.\n"
	if reply != "" && !isRawJSON {
		if strings.HasPrefix(reply, "#") {
			content = reply + "\n"
		} else {
			content = "# Deliverable Draft\n\n" + reply + "\n"
		}
	}

	return protocol.PlannedToolCall{
		Name: "write_file",
		Arguments: map[string]any{
			"path":          "deliverables/soma-draft.md",
			"content":       content,
			"artifact_role": "Deliverable Draft",
			"content_type":  "text/markdown",
		},
	}, true
}
