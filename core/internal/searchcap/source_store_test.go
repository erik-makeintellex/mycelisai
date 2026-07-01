package searchcap

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSourceStoreListAndCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	store := NewSourceStore(db)
	rows := sqlmock.NewRows([]string{
		"id", "name", "provider", "source_type", "endpoint", "scope_kind", "scope_ref",
		"boundary", "auth_scheme", "secret_ref", "mode", "sensitivity_class",
		"trust_class", "status", "recovery",
	}).AddRow(
		"research_api", "Research API", "local_api", "local_api", "https://search.example.test/api",
		"group", "research", "Approved research index", "api_token", "SEARCH_API_KEY",
		"live", "governed", "bounded_internal", "available", "",
	)
	mock.ExpectQuery("SELECT id, name, provider, source_type").WillReturnRows(rows)

	sources, err := store.List(context.Background())
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(sources) != 1 || sources[0].BaseURL != "https://search.example.test/api" || sources[0].SecretRef != "SEARCH_API_KEY" {
		t.Fatalf("sources = %+v", sources)
	}

	mock.ExpectExec("INSERT INTO search_sources").
		WithArgs(
			"research_api", "Research API", "local_api", "local_api", "https://search.example.test/api",
			"group", "research", "Approved research index", "api_token", "SEARCH_API_KEY",
			"live", "governed", "bounded_internal", "available", "",
		).
		WillReturnResult(sqlmock.NewResult(0, 1))

	created, err := store.Create(context.Background(), Source{
		ID: "research_api", Name: "Research API", Provider: "local_api", SourceType: "local_api",
		Endpoint: "https://search.example.test/api", BaseURL: "https://search.example.test/api",
		ScopeKind: "group", ScopeRef: "research", Boundary: "Approved research index",
		AuthScheme: "api_token", SecretRef: "SEARCH_API_KEY", Mode: "live",
		SensitivityClass: "governed", TrustClass: "bounded_internal", Status: "available",
	})
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if created.SecretRef != "SEARCH_API_KEY" {
		t.Fatalf("created = %+v", created)
	}

	mock.ExpectExec("UPDATE search_sources").
		WithArgs(
			"research_api", "Research API v2", "local_api", "local_api", "https://search.example.test/v2",
			"group", "research", "Approved research index v2", "none", "",
			"live", "governed", "bounded_internal", "available", "",
		).
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := store.Update(context.Background(), Source{
		ID: "research_api", Name: "Research API v2", Provider: "local_api", SourceType: "local_api",
		Endpoint: "https://search.example.test/v2", ScopeKind: "group", ScopeRef: "research",
		Boundary: "Approved research index v2", AuthScheme: "none", Mode: "live",
		SensitivityClass: "governed", TrustClass: "bounded_internal", Status: "available",
	}); err != nil {
		t.Fatalf("Update: %v", err)
	}

	mock.ExpectExec("DELETE FROM search_sources").
		WithArgs("research_api").
		WillReturnResult(sqlmock.NewResult(0, 1))
	if err := store.Delete(context.Background(), "research_api"); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestServiceUseSourceStoreLoadsPersistedSources(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	rows := sqlmock.NewRows([]string{
		"id", "name", "provider", "source_type", "endpoint", "scope_kind", "scope_ref",
		"boundary", "auth_scheme", "secret_ref", "mode", "sensitivity_class",
		"trust_class", "status", "recovery",
	}).AddRow(
		"docs", "Approved docs", "knowledge_collection", "knowledge_collection", "",
		"all", "", "Approved docs only", "none", "", "live", "governed",
		"trusted_internal", "available", "",
	)
	mock.ExpectQuery("SELECT id, name, provider, source_type").WillReturnRows(rows)

	svc := NewService(Config{Provider: ProviderDisabled}, nil, nil)
	if err := svc.UseSourceStore(context.Background(), NewSourceStore(db)); err != nil {
		t.Fatalf("UseSourceStore: %v", err)
	}
	sources := svc.ListSources()
	if len(sources) != 1 || sources[0].Name != "Approved docs" {
		t.Fatalf("sources = %+v", sources)
	}
}
