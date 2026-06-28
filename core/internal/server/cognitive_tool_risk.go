package server

// chatToolRisk estimates risk level from mutation tools used.
func chatToolRisk(tools []string) string {
	for _, t := range tools {
		if t == "publish_signal" || t == "broadcast" {
			return "high"
		}
		if t == "generate_blueprint" || t == "delegate" || t == "create_team" || t == "delegate_task" || t == "write_file" {
			return "medium"
		}
		if t == "promote_deployment_context" {
			return "high"
		}
	}
	return "low"
}
