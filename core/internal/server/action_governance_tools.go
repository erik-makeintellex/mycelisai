package server

import (
	"fmt"
	"strings"
)

func approvalRank(risk string) int {
	switch strings.TrimSpace(strings.ToLower(risk)) {
	case "high", "high-risk":
		return 3
	case "medium", "medium-risk":
		return 2
	default:
		return 1
	}
}

func maxRisk(left, right string) string {
	if approvalRank(left) >= approvalRank(right) {
		return left
	}
	return right
}

func capabilityForPlannedTool(name string) string {
	switch strings.TrimSpace(name) {
	case "write_file":
		return "file_output"
	case "generate_blueprint":
		return "planning"
	case "load_deployment_context", "promote_deployment_context", "remember", "summarize_conversation":
		return "learning"
	case "delegate":
		return "review"
	case "publish_signal", "broadcast":
		return "tool_execution"
	default:
		return ""
	}
}

func capabilityRiskForTool(name string, arguments map[string]any) string {
	switch strings.TrimSpace(name) {
	case "publish_signal", "broadcast", "promote_deployment_context":
		return "high"
	case "load_deployment_context":
		switch strings.TrimSpace(fmt.Sprint(arguments["knowledge_class"])) {
		case "company_knowledge", "soma_operating_context", "user_private_context", "reflection_synthesis":
			return "high"
		default:
			return "medium"
		}
	case "write_file", "delegate", "remember", "summarize_conversation":
		return "medium"
	default:
		return "low"
	}
}

func estimateActionCost(name string, arguments map[string]any) float64 {
	switch strings.TrimSpace(name) {
	case "publish_signal", "broadcast":
		return 1.5
	case "write_file":
		if content, ok := arguments["content"].(string); ok && len(content) > 500 {
			return 0.75
		}
		return 0.35
	case "load_deployment_context":
		switch strings.TrimSpace(fmt.Sprint(arguments["knowledge_class"])) {
		case "soma_operating_context", "user_private_context", "reflection_synthesis":
			return 0.95
		}
		if content, ok := arguments["content"].(string); ok && len(content) > 2000 {
			return 0.8
		}
		return 0.45
	case "promote_deployment_context":
		if content, ok := arguments["content"].(string); ok && len(content) > 2000 {
			return 0.9
		}
		return 0.55
	case "delegate":
		return 0.6
	case "generate_blueprint":
		return 0.2
	default:
		return 0.1
	}
}

func externalDataUseForTool(name string, arguments map[string]any) bool {
	switch strings.TrimSpace(name) {
	case "publish_signal", "broadcast":
		return true
	case "load_deployment_context", "promote_deployment_context":
		return strings.TrimSpace(fmt.Sprint(arguments["source_kind"])) == "web_research"
	default:
		return false
	}
}
