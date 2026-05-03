package server

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/triggers"
)

var triggerRuleColumns = []string{
	"id", "tenant_id", "name", "description", "event_pattern",
	"condition", "target_mission_id", "mode", "cooldown_seconds",
	"max_depth", "max_active_runs", "is_active", "last_fired_at",
	"created_at", "updated_at",
}

var triggerExecColumns = []string{
	"id", "rule_id", "event_id", "run_id", "status", "skip_reason", "executed_at",
}

// withTriggerStore wires a real triggers.Store (backed by sqlmock) onto the server.
func withTriggerStore(t *testing.T) (func(*AdminServer), sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock (triggers): %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return func(s *AdminServer) {
		s.Triggers = triggers.NewStore(db)
	}, mock
}
