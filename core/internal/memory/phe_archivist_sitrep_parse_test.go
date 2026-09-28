package memory

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/cognitive"
)

// TestPheArchivist_GenerateSitRep_JSONParseFailureReturnsHonestError is the
// negative/adversarial case for F22: when the model's response cannot be
// parsed as the expected sitrep JSON (even after fence extraction), the raw
// text must never be persisted as the summary. GenerateSitRep must fail
// honestly and write nothing.
func TestPheArchivist_GenerateSitRep_JSONParseFailureReturnsHonestError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}
	defer db.Close()

	mem := &Service{db: db}

	cog, err := cognitive.NewRouter("", nil)
	if err != nil {
		t.Fatalf("Failed to init Cognitive Router: %v", err)
	}
	cog.Config = &cognitive.BrainConfig{
		Providers: map[string]cognitive.ProviderConfig{
			"mock-llm": {Type: "mock", Enabled: true, ModelID: "mock-model"},
		},
		Profiles: map[string]string{
			"architect": "mock-llm",
		},
	}
	// Not valid JSON, and not fenced JSON either: extractJSON cannot recover it.
	cog.Adapters["mock-llm"] = &pheMockLLMAdapter{
		FixedResponse: "I cannot comply with structured output right now.",
	}

	archivist := NewArchivist(mem, cog)

	rows := sqlmock.NewRows([]string{"trace_id", "timestamp", "level", "source", "intent", "message", "context"}).
		AddRow("trace-1", time.Now(), "INFO", "agent-1", "test", "Hello World", []byte("{}"))
	mock.ExpectQuery("SELECT trace_id, timestamp, level, source, intent, message, context FROM log_entries").
		WillReturnRows(rows)

	err = archivist.GenerateSitRep(context.Background(), "team-1", 1*time.Hour)
	if err == nil {
		t.Fatal("expected an honest parse-failure error, got nil")
	}

	// No INSERT INTO sitreps expectation was set: sqlmock.ExpectationsWereMet
	// confirms the (only) queued expectation was the SELECT, i.e. no write
	// of the unparsable raw text happened.
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unexpected DB interaction: %v", err)
	}
}
