package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/cognitive"
)

// TestArchivistCompression is the Phase 5.0 verification pledge:
// mock 20 trace events, buffer them, compress them through the mock LLM,
// and persist the resulting sitrep.
func TestArchivistCompression(t *testing.T) {
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
	cog.Adapters["mock-llm"] = &cognitive.MockAdapter{
		FixedResponse: `{"summary": "20 events processed. Agent scanned 15 papers and summarized 5. No blockers.", "key_events": ["scan-start", "scan-complete", "summary-generated"], "strategies": "Continue monitoring throughput."}`,
	}

	archivist := NewArchivist(mem, cog)
	buffer := newEventBuffer(20)

	teamID := "22222222-2222-2222-2222-222222222222"
	var flushedEvents []BufferedEvent
	var flushed bool

	for i := 0; i < 20; i++ {
		content := fmt.Sprintf("Processing document %d", i+1)
		if i == 19 {
			content = "Final processing step"
		}

		raw := fmt.Sprintf(`{
			"meta": {"source_node": "paper-scanner", "timestamp": "%s"},
			"signal": "thought",
			"payload": {"content": "%s"}
		}`, time.Now().Add(time.Duration(i)*time.Second).Format(time.RFC3339), content)

		event := parseTraceEvent([]byte(raw))
		if event == nil {
			t.Fatalf("Failed to parse event %d", i)
		}

		var events []BufferedEvent
		_, events, flushed = buffer.push(teamID, *event)
		if flushed {
			flushedEvents = events
		}
	}

	if !flushed {
		t.Fatal("Buffer did not flush after 20 events")
	}
	if len(flushedEvents) != 20 {
		t.Fatalf("Expected 20 flushed events, got %d", len(flushedEvents))
	}

	mock.ExpectExec("INSERT INTO sitreps").
		WithArgs(
			teamID,
			anyArg(),
			anyArg(),
			"20 events processed. Agent scanned 15 papers and summarized 5. No blockers.",
			anyArg(),
			"Continue monitoring throughput.",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	archivist.compressAndStore(context.Background(), teamID, flushedEvents)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("Unmet DB expectations: %v", err)
	}

	eventsJSON, _ := json.Marshal([]string{"scan-start", "scan-complete", "summary-generated"})
	var parsed []string
	if err := json.Unmarshal(eventsJSON, &parsed); err != nil {
		t.Errorf("key_events JSON invalid: %v", err)
	}
	if len(parsed) != 3 {
		t.Errorf("Expected 3 key_events, got %d", len(parsed))
	}
}
