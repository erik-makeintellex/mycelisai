package swarm

import "strings"

// proposalPlanningBlockedTools contains tool names that must never execute
// during proposal generation. They can be proposed, but not run until the
// user has confirmed the action.
var proposalPlanningBlockedTools = map[string]struct{}{
	"broadcast":                  {},
	"create_exchange_thread":     {},
	"create_team":                {},
	"delegate_task":              {},
	"generate_image":             {},
	"load_deployment_context":    {},
	"local_command":              {},
	"promote_deployment_context": {},
	"publish_exchange_item":      {},
	"publish_signal":             {},
	"remember":                   {},
	"save_cached_image":          {},
	"send_external_message":      {},
	"store_artifact":             {},
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
