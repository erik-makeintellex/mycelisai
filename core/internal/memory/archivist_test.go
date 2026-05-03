package memory

import (
	"context"
	"fmt"
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
			"mock-llm": {Type: "mock", Enabled: true, ModelID: "mock-model"},
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

// ── Phase 5.0: Daemon Tests ──────────────────────────────────

func TestEventBuffer_ThresholdFlush(t *testing.T) {
	buf := newEventBuffer(5) // Low threshold for test

	for i := 0; i < 4; i++ {
		_, _, flush := buf.push("team-a", BufferedEvent{
			Source:    "agent-1",
			Signal:    "thought",
			Content:   fmt.Sprintf("Thinking step %d", i),
			Timestamp: time.Now(),
		})
		if flush {
			t.Fatalf("Buffer flushed early at event %d (threshold=5)", i)
		}
	}

	// 5th event triggers flush
	teamID, events, flush := buf.push("team-a", BufferedEvent{
		Source:    "agent-1",
		Signal:    "thought",
		Content:   "Thinking step 4",
		Timestamp: time.Now(),
	})
	if !flush {
		t.Fatal("Buffer did not flush at threshold")
	}
	if teamID != "team-a" {
		t.Errorf("Expected team-a, got %s", teamID)
	}
	if len(events) != 5 {
		t.Errorf("Expected 5 events, got %d", len(events))
	}
}

func TestEventBuffer_ArtifactFlush(t *testing.T) {
	buf := newEventBuffer(20)

	// Push 3 normal events
	for i := 0; i < 3; i++ {
		buf.push("team-b", BufferedEvent{
			Source: "agent-2", Signal: "thought", Content: "working",
			Timestamp: time.Now(),
		})
	}

	// Artifact triggers early flush
	_, events, flush := buf.push("team-b", BufferedEvent{
		Source: "agent-2", Signal: "artifact", Content: "Summary complete.",
		Timestamp: time.Now(),
	})
	if !flush {
		t.Fatal("Artifact signal did not trigger flush")
	}
	if len(events) != 4 {
		t.Errorf("Expected 4 events (3 thoughts + 1 artifact), got %d", len(events))
	}
}

func TestEventBuffer_Drain(t *testing.T) {
	buf := newEventBuffer(100)

	buf.push("team-a", BufferedEvent{Source: "a1", Signal: "thought", Timestamp: time.Now()})
	buf.push("team-b", BufferedEvent{Source: "b1", Signal: "thought", Timestamp: time.Now()})

	drained := buf.Drain()
	if len(drained) != 2 {
		t.Errorf("Expected 2 buckets, got %d", len(drained))
	}
	if len(drained["team-a"]) != 1 || len(drained["team-b"]) != 1 {
		t.Error("Drain returned incorrect bucket sizes")
	}

	// Buffer should be empty after drain
	drained2 := buf.Drain()
	if len(drained2) != 0 {
		t.Error("Buffer not empty after drain")
	}
}

func TestParseTraceEvent_CTSEnvelope(t *testing.T) {
	data := []byte(`{
		"meta": {"source_node": "paper-scanner", "timestamp": "2025-01-01T00:00:00Z"},
		"signal_type": "telemetry",
		"payload": {"content": "Downloading PDF"}
	}`)

	event := parseTraceEvent(data)
	if event == nil {
		t.Fatal("Failed to parse CTS envelope")
	}
	if event.Source != "paper-scanner" {
		t.Errorf("Expected source paper-scanner, got %s", event.Source)
	}
	if event.Content != "Downloading PDF" {
		t.Errorf("Expected content 'Downloading PDF', got %s", event.Content)
	}
}

func TestParseTraceEvent_GenericJSON(t *testing.T) {
	data := []byte(`{
		"meta": {"source_node": "summarizer", "capability_type": "digital"},
		"signal": "artifact",
		"payload": {"content": "Summary written to disk."}
	}`)

	event := parseTraceEvent(data)
	if event == nil {
		t.Fatal("Failed to parse generic trace JSON")
	}
	if event.Source != "summarizer" {
		t.Errorf("Expected source summarizer, got %s", event.Source)
	}
	if event.Signal != "artifact" {
		t.Errorf("Expected signal artifact, got %s", event.Signal)
	}
	if event.Content != "Summary written to disk." {
		t.Errorf("Expected 'Summary written to disk.', got %s", event.Content)
	}
}

func TestParseTraceEvent_Invalid(t *testing.T) {
	event := parseTraceEvent([]byte(`garbage`))
	if event != nil {
		t.Error("Expected nil for garbage input")
	}

	event = parseTraceEvent([]byte(`{"meta":{}}`))
	if event != nil {
		t.Error("Expected nil for empty source_node")
	}
}

func TestExtractJSON_CodeFences(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "clean JSON",
			input:    `{"summary": "test"}`,
			expected: `{"summary": "test"}`,
		},
		{
			name:     "markdown wrapped",
			input:    "```json\n{\"summary\": \"test\"}\n```",
			expected: `{"summary": "test"}`,
		},
		{
			name:     "no language tag",
			input:    "```\n{\"summary\": \"test\"}\n```",
			expected: `{"summary": "test"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractJSON(tt.input)
			if result != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, result)
			}
		})
	}
}
