package triggers

var ruleColumns = []string{
	"id", "tenant_id", "name", "description", "trigger_kind", "event_pattern",
	"condition", "target_mission_id", "mode", "cooldown_seconds",
	"max_depth", "max_active_runs", "is_active", "last_fired_at",
	"schedule_interval_seconds", "next_run_at", "proof_expectations", "recovery_behavior",
	"created_at", "updated_at",
}

var execColumns = []string{
	"id", "rule_id", "event_id", "run_id", "status", "skip_reason", "executed_at",
}
