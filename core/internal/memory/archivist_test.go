package memory

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/cognitive"
)

func TestArchivist_GenerateSitRep(t *testing.T) {
	// 1. Setup Mock DB (Memory Service)
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	mem := &Service{db: db}

	// 2. Setup Mock Cognitive Router (LLM)
	// We need to inject a mock adapter
	cog, err := cognitive.NewRouter("", nil) // No config, no DB needed for this test part if we inject adapter manually
	if err != nil {
		t.Fatalf("Failed to init Cognitive Router: %v", err)
	}

	// Inject Mock Adapter for "architect" profile (default profile for SitRep)
	// We assume "architect" maps to a provider that we can override or we just override the router's logic?
	// The Router loads config. If file missing, it might be empty.
	// We can manually set the profile mapping and adapter.

	// Inject Mock Adapter for "architect" profile
	cog.Config = &cognitive.BrainConfig{
		Providers: map[string]cognitive.ProviderConfig{
			"mock-llm": {Type: "mock"},
		},
		Profiles: map[string]string{
			"architect": "mock-llm",
		},
	}

	// Inject Mock Adapter
	mockLLM := &cognitive.MockAdapter{
		FixedResponse: `{"summary": "Test Summary", "key_events": ["Event A"], "strategies": "None"}`,
	}
	cog.Adapters["mock-llm"] = mockLLM

	// 3. Init Archivist
	archivist := NewArchivist(mem, cog)

	// 4. Expectations

	// A. ListLogs Query
	// Matches `SELECT trace_id... FROM log_entries...`
	// We return 1 row
	rows := sqlmock.NewRows([]string{"trace_id", "timestamp", "level", "source", "intent", "message", "context"}).
		AddRow("trace-1", time.Now(), "INFO", "agent-1", "test", "Hello World", []byte("{}"))

	mock.ExpectQuery("SELECT trace_id, timestamp, level, source, intent, message, context FROM log_entries").
		WillReturnRows(rows)

	// B. Insert SitRep
	mock.ExpectExec("INSERT INTO sitreps").
		WithArgs("team-1", anyArg(), anyArg(), "Test Summary", anyArg(), "None").
		WillReturnResult(sqlmock.NewResult(1, 1))

	// 5. Execute
	err = archivist.GenerateSitRep(context.Background(), "team-1", 1*time.Hour)
	if err != nil {
		t.Errorf("GenerateSitRep failed: %v", err)
	}

	// 6. Verify
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}
}

func anyArg() interface{} {
	return sqlmock.AnyArg()
}
