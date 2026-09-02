package journal

import (
	"strings"
	"testing"

	"github.com/mycelis/framework-runs/migrations"
)

func TestCurrentSchemaOwnsOnlyRunsServiceState(t *testing.T) {
	for _, table := range expectedTables {
		if !strings.Contains(migrations.CurrentSchema, "CREATE TABLE "+table) {
			t.Errorf("current schema omits %s", table)
		}
	}
	for _, forbidden := range []string{"worker_run_bindings", "worker_event_receipts", "missions", "outcomes", "nats"} {
		if strings.Contains(strings.ToLower(migrations.CurrentSchema), forbidden) {
			t.Errorf("current schema crosses Core ownership with %q", forbidden)
		}
	}
	for _, invariant := range []string{
		"run_events_immutable", "candidate_manifests_immutable",
		"run_commands_one_pending_per_run", "runs_idempotency_key_key",
	} {
		if invariant == "runs_idempotency_key_key" {
			if !strings.Contains(migrations.CurrentSchema, "idempotency_key VARCHAR(128) NOT NULL UNIQUE") {
				t.Errorf("schema omits %s invariant", invariant)
			}
			continue
		}
		if !strings.Contains(migrations.CurrentSchema, invariant) {
			t.Errorf("schema omits %s invariant", invariant)
		}
	}
}
