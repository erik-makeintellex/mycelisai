package swarm

import "strings"

const directAnswerRouteMarker = "[DIRECT ANSWER ROUTE]"

// proposalPlanningBlockedTools contains tool names that must never execute
// during proposal generation. They can be proposed, but not run until the
// user has confirmed the action.
var proposalPlanningBlockedTools = map[string]struct{}{
	"broadcast":                  {},
	"activate_config_document":   {},
	"create_exchange_thread":     {},
	"create_team":                {},
	"delegate":                   {},
	"delegate_task":              {},
	"generate_blueprint":         {},
	"generate_image":             {},
	"load_deployment_context":    {},
	"local_command":              {},
	"promote_deployment_context": {},
	"publish_exchange_item":      {},
	"publish_signal":             {},
	"research_for_blueprint":     {},
	"remember":                   {},
	"save_cached_image":          {},
	"send_external_message":      {},
	"store_artifact":             {},
	"store_config_document":      {},
	"store_inception_recipe":     {},
	"summarize_conversation":     {},
	"temp_memory_clear":          {},
	"temp_memory_write":          {},
	"write_file":                 {},
}

// blocksProposalPlanningTool reports whether a tool is mutation-capable and
// therefore must be treated as intended action only during proposal planning.
func blocksProposalPlanningTool(toolName string) bool {
	_, ok := proposalPlanningBlockedTools[strings.TrimSpace(toolName)]
	return ok
}

// isDirectAnswerRoute identifies chat requests that may answer or use
// read-only tools but never receive mutation authority.
func isDirectAnswerRoute(input string) bool {
	return strings.Contains(strings.ToUpper(input), directAnswerRouteMarker)
}
