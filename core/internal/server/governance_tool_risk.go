package server

import "strings"

// registeredInternalToolRisk is the explicit risk class of every internal tool
// registered by swarm.NewInternalToolRegistry (A2b C1b). It is the only source
// that may rate a tool below high: `mcp:*`, `toolset:*`, unknown names and case
// variants are not in it and classify as high. A test keeps it equal to the
// runtime registry, so a new tool must be classified here explicitly.
var registeredInternalToolRisk = map[string]string{
	// read-only lookups
	"get_system_status": "low", "list_available_tools": "low", "list_missions": "low", "list_teams": "low",
	"list_catalogue": "low", "generate_blueprint": "low", "list_docs": "low", "list_exchange_channels": "low", "list_exchange_threads": "low",
	"preview_config_document": "low", "read_doc": "low", "read_file": "low", "read_handoff_input": "low", "read_signals": "low",
	"recall": "low", "recall_inception_recipes": "low", "search_docs": "low", "search_exchange_items": "low",
	"search_memory": "low", "temp_memory_read": "low",
	"code_context.query": "low", "code_context.explain": "low", "code_context.impact": "low",
	// governed writes, orchestration, and external reads
	"web_search": "medium", "research_for_blueprint": "medium", "consult_council": "medium",
	"create_team": "medium", "delegate_task": "medium", "hand_off": "medium", "generate_image": "medium",
	"activate_config_document": "medium", "store_config_document": "medium", "create_exchange_thread": "medium",
	"publish_exchange_item": "medium", "instantiate_conversation_template": "medium", "load_deployment_context": "medium",
	"remember": "medium", "save_cached_image": "medium", "store_artifact": "medium", "store_conversation_template": "medium",
	"store_inception_recipe": "medium", "summarize_conversation": "medium", "temp_memory_clear": "medium",
	"temp_memory_write": "medium", "write_file": "medium",
	// external effects and host execution
	"broadcast": "high", "publish_signal": "high", "promote_deployment_context": "high",
	"local_command": "high", "send_external_message": "high",
}

// capabilityRiskForTool is the one fail-closed tool classifier shared by chat,
// council and blueprint approval: the name is trimmed and looked up exactly
// (case-sensitive) in registeredInternalToolRisk; anything else is high. A
// registered tool may still be raised by its arguments.
func capabilityRiskForTool(name string, arguments map[string]any) string {
	trimmed := strings.TrimSpace(name)
	base, ok := registeredInternalToolRisk[trimmed]
	if !ok {
		return "high"
	}
	return maxRisk(base, argumentToolRisk(trimmed, arguments))
}
