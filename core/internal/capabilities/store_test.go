package capabilities

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestStoreReplaceSnapshotPersistsManifestState(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC)
	store := NewStore(db)
	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO capability_manifests").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec("DELETE FROM capability_manifests").
		WithArgs("search:web_search").
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err = store.ReplaceSnapshot(context.Background(), Snapshot{
		GeneratedAt: now,
		Count:       1,
		Manifests: []Manifest{{
			ID:                  "search:web_search",
			DisplayName:         "Mycelis Search",
			Kind:                "search_capability",
			Source:              "searchcap",
			Status:              "enabled",
			RiskClass:           "medium-risk",
			Description:         "Search runtime capability.",
			DefaultAllowedRoles: []string{"soma"},
			AuditRequired:       true,
			DerivedAt:           now,
			UpdatedAt:           now,
		}},
	})
	if err != nil {
		t.Fatalf("ReplaceSnapshot: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func TestStoreListReturnsRuntimeVisibleState(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	now := time.Date(2026, 5, 17, 12, 0, 0, 0, time.UTC)
	rows := sqlmock.NewRows(capabilityManifestColumns()).
		AddRow(
			"search:web_search", "search:web_search", ManifestVersion, ManifestVersion,
			"Mycelis Search", "search_capability", "searchcap", "enabled", "healthy",
			"medium-risk", "Search runtime capability.", "Search runtime capability.",
			`["web_search"]`, `["soma"]`, `["soma"]`, true, false, "not_required",
			"SearchQuery", "ResearchSummary", "enabled", now,
			"surface_runtime_error_and_audit", "retry_with_audit", "required",
			"no_raw_secrets", "runtime_capability_search", `{}`, now, now,
		)
	mock.ExpectQuery("SELECT id, capability_id, version, manifest_version").
		WillReturnRows(rows)

	snap, err := NewStore(db).List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if snap.Count != 1 {
		t.Fatalf("Count = %d, want 1", snap.Count)
	}
	got := snap.Manifests[0]
	if got.CapabilityID != "search:web_search" || got.Health != "healthy" {
		t.Fatalf("unexpected manifest state: %+v", got)
	}
	if got.OutputSchemaRef != "ResearchSummary" {
		t.Fatalf("output_schema_ref = %q", got.OutputSchemaRef)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sql expectations: %v", err)
	}
}

func capabilityManifestColumns() []string {
	return []string{
		"id", "capability_id", "version", "manifest_version", "display_name",
		"kind", "source", "status", "health", "risk_class", "description",
		"purpose", "tool_refs", "default_allowed_roles", "allowed_roles",
		"audit_required", "approval_required", "approval_posture",
		"input_schema_ref", "output_schema_ref", "last_probe_status",
		"last_probe_at", "failure_posture", "recovery_posture", "audit_policy",
		"secret_ref_policy", "owner", "metadata", "derived_at", "updated_at",
	}
}
