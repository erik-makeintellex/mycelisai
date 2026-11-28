package server

import (
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func isInteractingMultiAppRequest(text string) bool {
	lower := strings.ToLower(strings.TrimSpace(text))
	if lower == "" {
		return false
	}
	hasCompanionCue := requestContainsAny(lower, []string{
		"another app", "another application", "second app", "second application",
		"companion app", "companion application", "two apps", "two applications",
		"2 apps", "2 applications", "another team", "second team",
		"have a team work on another", "team work on another",
	})
	hasInteractionCue := requestContainsAny(lower, []string{
		"interact", "interacts", "interacting", "talk to", "talks to",
		"connect to", "connected to", "communicate with", "communicates with",
		"sync with", "syncs with", "integrate with", "integrates with",
		"initial ask", "first app", "primary app", "each other", "together",
	})
	if hasCompanionCue && (hasInteractionCue || hasDeliverableCreationIntent(lower) || hasTeamCreationIntent(lower)) {
		return true
	}
	return hasInteractionCue && requestContainsAny(lower, []string{"app", "application", "project", "service"}) &&
		requestContainsAny(lower, []string{"another", "second", "companion", "two", "both"})
}

func isExplicitWebAppOrArtifactTarget(lower string) bool {
	return requestContainsAny(lower, []string{
		"app", "application", "website", "web app", "landing page", "calculator", "timer",
		"stopwatch", "todo", "to-do", "tracker", "board", "kanban", "widget", "converter",
		"form", "simulator", "visualizer", "frontend", "portal", "prototype",
		"game", "script", "checklist", "runbook", "roadmap", "markdown", "document",
	})
}

func isSimpleCasualBuildRequest(lower string) bool {
	if lower == "" || namedFilePattern.MatchString(lower) {
		return false
	}
	if requestContainsAny(lower, []string{"mcp", "mcps", "target mcp", "web search", "private service", "api token"}) {
		return false
	}
	if hasTeamCreationIntent(lower) && !isExplicitWebAppOrArtifactTarget(lower) {
		return false
	}
	casualVerbs := []string{
		"make me", "build me", "create me", "give me", "put together",
		"spin up", "whip up", "set up", "start a", "i want a", "i need a", "i'd like a", "id like a",
		"can we make", "can you make", "can you build", "can you create", "could you make", "could you build", "could you create",
		"please make", "please build", "please create", "help me make", "help me build", "help me create",
		"let's make", "let's build", "let's create", "lets make", "lets build", "lets create",
	}
	casualTargets := []string{
		"app", "application", "project", "game", "site", "website", "page",
		"tool", "widget", "tracker", "calculator", "timer", "stopwatch", "todo", "to-do", "board", "kanban",
		"chat", "store", "shop", "cart", "api", "backend", "frontend", "service", "companion",
		"checklist", "runbook", "roadmap", "guide", "report", "script", "utility", "notes",
	}
	return requestContainsAny(lower, casualVerbs) && requestContainsAny(lower, casualTargets)
}

func resolveInteractingAppSlugs(goal string) (string, string) {
	lower := strings.ToLower(goal)
	keywords := []string{
		"storefront", "checkout", "inventory", "analytics", "order", "payment",
		"todo", "tracker", "calculator", "timer", "kanban", "chat", "dashboard",
		"booking", "catalog", "admin", "customer", "mobile", "api", "backend", "frontend",
	}
	var matched []string
	for _, kw := range keywords {
		if strings.Contains(lower, kw) {
			matched = append(matched, kw+"-app")
		}
	}
	primary := "primary-core-app"
	companion := "companion-connected-app"
	if len(matched) >= 1 {
		primary = matched[0]
	}
	if len(matched) >= 2 && matched[1] != primary {
		companion = matched[1]
	} else if primary != "primary-core-app" {
		companion = strings.TrimSuffix(primary, "-app") + "-companion-app"
	}
	return primary, companion
}

func inferInteractingMultiAppTeamCalls(latestRequest string) []protocol.PlannedToolCall {
	primarySlug, companionSlug := resolveInteractingAppSlugs(latestRequest)
	primaryTeamID := primarySlug + "-team"
	companionTeamID := companionSlug + "-team"
	reviewTeamID := primarySlug + "-integration-review-team"
	primaryTitle := titleCaseTeamID(primarySlug)
	companionTitle := titleCaseTeamID(companionSlug)

	primaryOut := fmt.Sprintf(protocol.TopicTeamSignalResult, primaryTeamID)
	companionOut := fmt.Sprintf(protocol.TopicTeamSignalResult, companionTeamID)

	return []protocol.PlannedToolCall{
		{
			Name: "create_team",
			Arguments: map[string]any{
				"team_id":     primaryTeamID,
				"name":        primaryTitle + " Team",
				"description": fmt.Sprintf("Builds the primary application (%s) and publishes state/API events on %s for companion apps.", primaryTitle, primaryOut),
				"inputs":      []string{fmt.Sprintf(protocol.TopicTeamInternalCommand, primaryTeamID), companionOut, protocol.TopicGlobalBroadcast},
				"deliveries":  []string{primaryOut, fmt.Sprintf(protocol.TopicTeamSignalStatus, primaryTeamID)},
				"members": map[string]any{
					"architect":         "council-architect",
					"developer":         "prime-development-agent",
					"designer":          "agui-design-architect-agent",
					"delivery-reviewer": "council-sentry",
				},
				"tools": []string{"write_file", "publish_signal", "read_file"},
			},
		},
		{
			Name: "create_team",
			Arguments: map[string]any{
				"team_id":     companionTeamID,
				"name":        companionTitle + " Team",
				"description": fmt.Sprintf("Builds the companion application (%s) that interacts with %s via %s.", companionTitle, primaryTitle, primaryOut),
				"inputs":      []string{primaryOut, fmt.Sprintf(protocol.TopicTeamInternalCommand, companionTeamID)},
				"deliveries":  []string{companionOut, fmt.Sprintf(protocol.TopicTeamSignalStatus, companionTeamID)},
				"members": map[string]any{
					"integrator":        "prime-development-agent",
					"ui-developer":      "agui-design-architect-agent",
					"delivery-reviewer": "council-sentry",
				},
				"tools": []string{"write_file", "publish_signal", "read_file"},
			},
		},
		{
			Name: "create_team",
			Arguments: map[string]any{
				"team_id":     reviewTeamID,
				"name":        primaryTitle + " & " + companionTitle + " — Integration Review Team",
				"description": fmt.Sprintf("Dedicated cross-application review team verifying the API contract and live interaction between %s and %s.", primaryTitle, companionTitle),
				"inputs":      []string{primaryOut, companionOut, fmt.Sprintf(protocol.TopicTeamInternalCommand, reviewTeamID)},
				"deliveries":  []string{fmt.Sprintf(protocol.TopicTeamSignalResult, reviewTeamID), fmt.Sprintf(protocol.TopicTeamSignalStatus, reviewTeamID)},
				"members": map[string]any{
					"lead-reviewer":        "council-sentry",
					"architecture-auditor": "council-architect",
					"qa-verifier":          "agui-design-architect-agent",
				},
				"tools": []string{"read_file", "publish_signal"},
			},
		},
	}
}

func interactingMultiAppBundleSpecs(goal string) ([]deliverableBundleFileSpec, string) {
	primarySlug, companionSlug := resolveInteractingAppSlugs(goal)
	primaryTitle := titleCaseTeamID(primarySlug)
	companionTitle := titleCaseTeamID(companionSlug)
	primaryTeamID := primarySlug + "-team"
	companionTeamID := companionSlug + "-team"
	primaryResult := fmt.Sprintf(protocol.TopicTeamSignalResult, primaryTeamID)
	companionResult := fmt.Sprintf(protocol.TopicTeamSignalResult, companionTeamID)
	suiteSlug := strings.TrimSuffix(primarySlug, "-app") + "-and-" + companionSlug

	return []deliverableBundleFileSpec{
		{
			name:        "index.html",
			title:       primaryTitle + " ↔ " + companionTitle + " — Live Interacting Dual-App Workbench (HTML5)",
			contentType: "text/html",
			content:     interactingMultiAppHTML(primaryTitle, companionTitle, primaryTeamID, companionTeamID, goal),
		},
		{
			name:        "app_interaction_contract.json",
			title:       "Cross-Application Event & API Contract (JSON)",
			contentType: "application/json",
			content: "{\n" +
				"  \"suite_id\": " + strconvQuoteSafe(suiteSlug) + ",\n" +
				"  \"primary_app\": {\n" +
				"    \"app_id\": " + strconvQuoteSafe(primarySlug) + ",\n" +
				"    \"owner_team\": " + strconvQuoteSafe(primaryTeamID) + ",\n" +
				"    \"publishes_on\": " + strconvQuoteSafe(primaryResult) + ",\n" +
				"    \"listens_on\": " + strconvQuoteSafe(companionResult) + "\n" +
				"  },\n" +
				"  \"companion_app\": {\n" +
				"    \"app_id\": " + strconvQuoteSafe(companionSlug) + ",\n" +
				"    \"owner_team\": " + strconvQuoteSafe(companionTeamID) + ",\n" +
				"    \"listens_on\": " + strconvQuoteSafe(primaryResult) + ",\n" +
				"    \"publishes_on\": " + strconvQuoteSafe(companionResult) + "\n" +
				"  },\n" +
				"  \"review_gate_team\": " + strconvQuoteSafe(primarySlug+"-integration-review-team") + ",\n" +
				"  \"shared_events\": [\"item.created\", \"state.synced\", \"ack.received\"]\n" +
				"}\n",
		},
		{
			name:        "multi_team_architecture.md",
			title:       "Multi-Team & Cross-App Integration Guide",
			contentType: "text/markdown",
			content: "# " + primaryTitle + " ↔ " + companionTitle + " Integration Guide\n\n## Goal\n" + goal + "\n\n" +
				"## Coordinated Teams\n" +
				"1. **" + primaryTitle + " Team (`" + primaryTeamID + "`):** Owns the primary application and emits live events on `" + primaryResult + "`.\n" +
				"2. **" + companionTitle + " Team (`" + companionTeamID + "`):** Owns the interacting companion application, subscribes to `" + primaryTeamID + "` output events, and sends acknowledgments back.\n" +
				"3. **Integration Review Team (`" + primarySlug + "-integration-review-team`):** Certifies contract compatibility and SHA-256 `OutputProofEnvelope` delivery.\n",
		},
	}, suiteSlug
}

func interactingMultiAppHTML(primaryTitle, companionTitle, primaryTeamID, companionTeamID, goal string) string {
	safeGoal := strings.ReplaceAll(strings.ReplaceAll(goal, "<", "&lt;"), ">", "&gt;")
	return `<!DOCTYPE html>
<html lang="en">
<head>
<meta charset="UTF-8">
<title>` + primaryTitle + ` ↔ ` + companionTitle + `</title>
<style>
  body { font-family: system-ui, sans-serif; background: #090d16; color: #f8fafc; margin: 0; padding: 1.25rem; }
  h1 { margin: 0 0 0.25rem; font-size: 1.2rem; color: #38bdf8; }
  p.sub { margin: 0 0 1rem; font-size: 0.8rem; color: #94a3b8; }
  .grid { display: grid; grid-template-columns: 1fr 1fr; gap: 1rem; }
  .pane { background: #111827; border: 1px solid #334155; border-radius: 10px; padding: 1rem; }
  .pane h2 { margin: 0 0 0.25rem; font-size: 0.98rem; color: #e2e8f0; }
  .team-tag { display: inline-block; font-family: monospace; font-size: 0.7rem; padding: 0.15rem 0.45rem; border-radius: 4px; background: #1e293b; color: #38bdf8; margin-bottom: 0.75rem; }
  .row { display: flex; gap: 0.4rem; margin-bottom: 0.75rem; }
  input { flex: 1; padding: 0.45rem 0.6rem; border-radius: 6px; border: 1px solid #475569; background: #020617; color: #f8fafc; font-size: 0.82rem; }
  button { padding: 0.45rem 0.75rem; border-radius: 6px; border: 1px solid #38bdf8; background: #0284c7; color: #fff; font-weight: 600; cursor: pointer; font-size: 0.8rem; }
  ul { list-style: none; padding: 0; margin: 0; display: flex; flex-direction: column; gap: 0.4rem; }
  li { display: flex; justify-content: space-between; align-items: center; padding: 0.45rem 0.6rem; border-radius: 6px; background: #020617; border: 1px solid #1e293b; font-size: 0.8rem; }
  .badge { font-size: 0.68rem; padding: 0.1rem 0.4rem; border-radius: 999px; background: #065f46; color: #a7f3d0; }
</style>
</head>
<body>
  <h1>` + primaryTitle + ` &harr; ` + companionTitle + ` (Live Interacting Apps)</h1>
  <p class="sub">` + safeGoal + `</p>
  <div class="grid">
    <div class="pane" data-testid="primary-app-pane">
      <h2>1. ` + primaryTitle + `</h2>
      <span class="team-tag">Owner: ` + primaryTeamID + `</span>
      <div class="row">
        <input id="primary-input" placeholder="Send event to ` + companionTitle + `..." />
        <button id="send-to-companion-btn" type="button" onclick="sendEvent()">Publish &rarr;</button>
      </div>
      <ul id="primary-log">
        <li><span>Initial ask ready &amp; broadcasting</span><span class="badge">Synced</span></li>
      </ul>
    </div>
    <div class="pane" data-testid="companion-app-pane">
      <h2>2. ` + companionTitle + `</h2>
      <span class="team-tag">Owner: ` + companionTeamID + ` (Listens on ` + primaryTeamID + `)</span>
      <div class="row">
        <button id="ack-primary-btn" type="button" onclick="ackPrimary()">&larr; Send Companion Ack</button>
      </div>
      <ul id="companion-log">
        <li><span>Connected to ` + primaryTitle + ` event stream</span><span class="badge">Listening</span></li>
      </ul>
    </div>
  </div>
  <script>
    function sendEvent() {
      const input = document.getElementById('primary-input');
      const val = (input.value || 'Sync payload #' + Math.floor(Math.random()*100)).trim();
      const pLi = document.createElement('li');
      pLi.innerHTML = '<span>Sent: ' + val + '</span><span class="badge">Published</span>';
      document.getElementById('primary-log').prepend(pLi);
      const cLi = document.createElement('li');
      cLi.innerHTML = '<span>Received from Primary: ' + val + '</span><span class="badge">Processed</span>';
      document.getElementById('companion-log').prepend(cLi);
      input.value = '';
    }
    function ackPrimary() {
      const pLi = document.createElement('li');
      pLi.innerHTML = '<span>Companion Ack received</span><span class="badge">Verified</span>';
      document.getElementById('primary-log').prepend(pLi);
    }
  </script>
</body>
</html>`
}
