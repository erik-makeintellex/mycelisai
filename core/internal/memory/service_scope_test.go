package memory

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestSemanticSearchWithOptions_TeamScoped(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	svc := NewServiceWithDB(db)
	now := time.Now()
	rows := sqlmock.NewRows([]string{"id", "content", "metadata", "score", "created_at"}).
		AddRow("vec-1", "team memory", `{"team_id":"alpha","visibility":"team","type":"agent_memory"}`, 0.91, now)

	mock.ExpectQuery("SELECT id, content, metadata, 1 - \\(embedding <=> \\$1::vector\\) AS score, created_at").
		WithArgs(sqlmock.AnyArg(), "default", "agent_memory", "alpha", 5).
		WillReturnRows(rows)

	results, err := svc.SemanticSearchWithOptions(context.Background(), []float64{0.1, 0.2}, SemanticSearchOptions{
		Limit:       5,
		TenantID:    "default",
		TeamID:      "alpha",
		Types:       []string{"agent_memory"},
		AllowGlobal: true,
	})
	if err != nil {
		t.Fatalf("SemanticSearchWithOptions: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("results len = %d, want 1", len(results))
	}
	if results[0].Metadata["team_id"] != "alpha" {
		t.Fatalf("team_id = %v, want alpha", results[0].Metadata["team_id"])
	}
}
