package server

import (
	"net/http"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/internal/memory"
)

func TestHandleDeploymentContext_PostNormalizesAdminSomaDefaults(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mem := memory.NewServiceWithDB(db)
	artSvc := artifacts.NewService(db, "/data/artifacts")
	s := &AdminServer{
		Artifacts: artSvc,
		Mem:       mem,
		Cognitive: newDeploymentContextBrain(),
	}

	now := time.Now()
	mock.ExpectQuery("INSERT INTO artifacts").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(),
			"soma", sqlmock.AnyArg(),
			artifacts.ArtifactType("document"), "Soma Identity Guidance", "text/markdown",
			"Keep shared Soma behavior steady and concise.", sqlmock.AnyArg(), sqlmock.AnyArg(),
			metadataContains{
				"knowledge_class":    "soma_operating_context",
				"source_kind":        "user_note",
				"source_label":       "admin guidance",
				"visibility":         "global",
				"sensitivity_class":  "restricted",
				"trust_class":        "trusted_internal",
				"soma_context_kind":  "identity",
				"output_specificity": "balanced",
				"tags":               []string{"shared-output-specificity", "soma-operating-context"},
			},
			sqlmock.AnyArg(), "approved",
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow("cccccccc-cccc-cccc-cccc-cccccccccccc", now))
	mock.ExpectExec("INSERT INTO context_vectors").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{
		"knowledge_class": "soma_operating_context",
		"title": "Soma Identity Guidance",
		"content": "Keep shared Soma behavior steady and concise.",
		"source_kind": "user_document",
		"visibility": "private",
		"sensitivity_class": "role_scoped",
		"trust_class": "user_provided",
		"soma_context_kind": "identity",
		"output_specificity": "balanced"
	}`
	rr := doRequest(t, http.HandlerFunc(s.HandleDeploymentContext), http.MethodPost, "/api/v1/memory/deployment-context", body)
	assertStatus(t, rr, http.StatusCreated)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["knowledge_class"] != "soma_operating_context" {
		t.Fatalf("expected soma_operating_context knowledge class, got %+v", resp)
	}
	if resp["source_kind"] != "user_note" {
		t.Fatalf("expected source_kind=user_note, got %+v", resp)
	}
	if resp["source_label"] != "admin guidance" {
		t.Fatalf("expected admin guidance source label, got %+v", resp)
	}
	if resp["visibility"] != "global" {
		t.Fatalf("expected visibility=global, got %+v", resp)
	}
	if resp["sensitivity_class"] != "restricted" {
		t.Fatalf("expected sensitivity_class=restricted, got %+v", resp)
	}
	if resp["trust_class"] != "trusted_internal" {
		t.Fatalf("expected trust_class=trusted_internal, got %+v", resp)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestHandleDeploymentContext_PostStoresUserPrivateGoalContext(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mem := memory.NewServiceWithDB(db)
	artSvc := artifacts.NewService(db, "/data/artifacts")
	s := &AdminServer{
		Artifacts: artSvc,
		Mem:       mem,
		Cognitive: newDeploymentContextBrain(),
	}

	now := time.Now()
	mock.ExpectQuery("INSERT INTO artifacts").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(),
			"admin", sqlmock.AnyArg(),
			artifacts.ArtifactType("document"), "Personal Finance Notes", "text/markdown",
			"Q2 savings goals and invoice timing.", sqlmock.AnyArg(), sqlmock.AnyArg(),
			metadataContains{
				"knowledge_class":   "user_private_context",
				"source_kind":       "finance_record",
				"source_label":      "private user content",
				"visibility":        "private",
				"sensitivity_class": "restricted",
				"trust_class":       "user_provided",
				"content_domain":    "finance",
				"target_goal_sets":  []string{"tax-planning", "cash-flow"},
				"tags":              []string{"finance", "user-private-context"},
			},
			sqlmock.AnyArg(), "approved",
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow("dddddddd-dddd-dddd-dddd-dddddddddddd", now))
	mock.ExpectExec("INSERT INTO context_vectors").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{
		"knowledge_class": "user_private_context",
		"title": "Personal Finance Notes",
		"content": "Q2 savings goals and invoice timing.",
		"source_kind": "finance_record",
		"visibility": "global",
		"sensitivity_class": "role_scoped",
		"content_domain": "finance",
		"target_goal_sets": ["tax-planning", "cash-flow"]
	}`
	rr := doRequest(t, http.HandlerFunc(s.HandleDeploymentContext), http.MethodPost, "/api/v1/memory/deployment-context", body)
	assertStatus(t, rr, http.StatusCreated)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["knowledge_class"] != "user_private_context" {
		t.Fatalf("expected user_private_context knowledge class, got %+v", resp)
	}
	if resp["visibility"] != "private" {
		t.Fatalf("expected visibility=private, got %+v", resp)
	}
	if resp["sensitivity_class"] != "restricted" {
		t.Fatalf("expected sensitivity_class=restricted, got %+v", resp)
	}
	if resp["content_domain"] != "finance" {
		t.Fatalf("expected content_domain=finance, got %+v", resp)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}

func TestHandleDeploymentContext_PostStoresReflectionSynthesisDefaults(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	mem := memory.NewServiceWithDB(db)
	artSvc := artifacts.NewService(db, "/data/artifacts")
	s := &AdminServer{
		Artifacts: artSvc,
		Mem:       mem,
		Cognitive: newDeploymentContextBrain(),
	}

	now := time.Now()
	mock.ExpectQuery("INSERT INTO artifacts").
		WithArgs(
			sqlmock.AnyArg(), sqlmock.AnyArg(),
			"admin", sqlmock.AnyArg(),
			artifacts.ArtifactType("document"), "Investor Workflow Shift", "text/markdown",
			"The user trajectory shifted toward investor-ready team-managed media output demos.", sqlmock.AnyArg(), sqlmock.AnyArg(),
			metadataContains{
				"knowledge_class":   "reflection_synthesis",
				"source_kind":       "synthesis_note",
				"source_label":      "reflection synthesis",
				"visibility":        "private",
				"sensitivity_class": "restricted",
				"trust_class":       "trusted_internal",
				"content_domain":    "reflection",
				"reflection_kind":   "synthesis_note",
				"target_goal_sets":  []string{"investor-review"},
				"tags":              []string{"reflection-synthesis-memory", "synthesis_note"},
			},
			sqlmock.AnyArg(), "approved",
		).
		WillReturnRows(sqlmock.NewRows([]string{"id", "created_at"}).
			AddRow("eeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee", now))
	mock.ExpectExec("INSERT INTO context_vectors").
		WithArgs(sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	body := `{
		"knowledge_class": "reflection_synthesis",
		"title": "Investor Workflow Shift",
		"content": "The user trajectory shifted toward investor-ready team-managed media output demos.",
		"target_goal_sets": ["investor-review"]
	}`
	rr := doRequest(t, http.HandlerFunc(s.HandleDeploymentContext), http.MethodPost, "/api/v1/memory/deployment-context", body)
	assertStatus(t, rr, http.StatusCreated)

	var resp map[string]any
	assertJSON(t, rr, &resp)
	if resp["knowledge_class"] != "reflection_synthesis" {
		t.Fatalf("expected reflection_synthesis knowledge class, got %+v", resp)
	}
	if resp["source_kind"] != "synthesis_note" {
		t.Fatalf("expected source_kind=synthesis_note, got %+v", resp)
	}
	if resp["visibility"] != "private" {
		t.Fatalf("expected visibility=private, got %+v", resp)
	}
	if resp["sensitivity_class"] != "restricted" {
		t.Fatalf("expected sensitivity_class=restricted, got %+v", resp)
	}
	if resp["trust_class"] != "trusted_internal" {
		t.Fatalf("expected trust_class=trusted_internal, got %+v", resp)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expectations: %v", err)
	}
}
