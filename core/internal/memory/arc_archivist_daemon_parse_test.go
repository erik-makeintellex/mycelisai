package memory

import (
	"bytes"
	"context"
	"log"
	"strings"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/cognitive"
)

// arcNewCompressionArchivist builds an Archivist wired to a mock cognitive
// Router whose "architect" profile returns FixedResponse, plus a sqlmock DB.
// Callers get the mock back to set/verify DB expectations.
func arcNewCompressionArchivist(t *testing.T, fixedResponse string) (*Archivist, sqlmock.Sqlmock, func()) {
	t.Helper()

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("Failed to mock DB: %v", err)
	}

	mem := &Service{db: db}

	cog, err := cognitive.NewRouter("", nil)
	if err != nil {
		db.Close()
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
	cog.Adapters["mock-llm"] = &pheMockLLMAdapter{FixedResponse: fixedResponse}

	archivist := NewArchivist(mem, cog)
	return archivist, mock, func() { db.Close() }
}

func arcBufferedEvents(teamID string) []BufferedEvent {
	now := time.Now()
	return []BufferedEvent{
		{Source: "agent-1", Signal: "thought", Content: "did some work", Timestamp: now},
		{Source: "agent-1", Signal: "output", Content: "finished the task", Timestamp: now.Add(time.Second)},
	}
}

// TestArcCompressAndStore_NonJSONResponse_NeverStoresRawText is the negative/
// adversarial case: the model returns prose with no JSON at all. The daemon
// compression path (compressAndStore) must not persist that raw text as the
// sitrep summary, and must not report success by writing a row anyway. This
// mirrors PH-E's F22 fix to GenerateSitRep (archivist.go), which the same
// commit's follow-ups flagged as still missing from archivist_daemon.go.
func TestArcCompressAndStore_NonJSONResponse_NeverStoresRawText(t *testing.T) {
	archivist, mock, closeDB := arcNewCompressionArchivist(t, "I cannot comply with structured output right now.")
	defer closeDB()

	teamID := "arc-team-nonjson"
	events := arcBufferedEvents(teamID)

	// No INSERT expectation is registered at all: a fixed, honest
	// parse-failure path must never attempt to persist anything, so
	// ExpectationsWereMet trivially holds either way. What actually
	// distinguishes "stored raw text" from "skipped honestly" is whether
	// compressAndStore even attempts the DB write, which we observe via
	// the log output: the buggy path logs "Storing raw" and then a DB
	// write attempt ("DB write failed", since no expectation matches);
	// the fixed path must log the parse failure and stop before any DB
	// call.
	var logBuf bytes.Buffer
	prevOut := log.Writer()
	prevFlags := log.Flags()
	log.SetOutput(&logBuf)
	defer func() {
		log.SetOutput(prevOut)
		log.SetFlags(prevFlags)
	}()

	archivist.compressAndStore(context.Background(), teamID, events)

	logged := logBuf.String()
	if strings.Contains(logged, "Storing raw") {
		t.Errorf("compressAndStore stored raw model text as the summary; log: %s", logged)
	}
	if strings.Contains(logged, "DB write failed") {
		t.Errorf("compressAndStore attempted a DB write after a parse failure; log: %s", logged)
	}
	if !strings.Contains(logged, "parse") {
		t.Errorf("expected an honest parse-failure log message, got: %s", logged)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unexpected DB interaction: %v", err)
	}
}

// TestArcCompressAndStore_FencedJSONResponse_ExtractsAndStores is the
// positive case: JSON wrapped in a leading markdown fence (the shape
// extractJSON is built to strip) must be recognized and its "summary" field
// persisted -- not the raw fenced text.
func TestArcCompressAndStore_FencedJSONResponse_ExtractsAndStores(t *testing.T) {
	fenced := "```json\n" +
		`{"summary": "Team made steady progress.", "key_events": ["step1"], "strategies": "Keep going."}` +
		"\n```"
	archivist, mock, closeDB := arcNewCompressionArchivist(t, fenced)
	defer closeDB()

	teamID := "arc-team-fenced"
	events := arcBufferedEvents(teamID)

	mock.ExpectExec("INSERT INTO sitreps").
		WithArgs(
			teamID,
			anyArg(),
			anyArg(),
			"Team made steady progress.",
			anyArg(),
			"Keep going.",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	archivist.compressAndStore(context.Background(), teamID, events)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}

// TestArcCompressAndStore_JSONWrappedInProse_HonestFailure covers JSON
// preceded by prose (a shape the existing extractJSON helper is not built to
// recover, since it only strips a fence that starts the text). The daemon
// must not fall back to storing that raw text as the summary: it must fail
// honestly and skip the write, same as the plain non-JSON case.
func TestArcCompressAndStore_JSONWrappedInProse_HonestFailure(t *testing.T) {
	prosePrefixed := "Here is the SitRep:\n```json\n" +
		`{"summary": "Team made steady progress.", "key_events": ["step1"], "strategies": "Keep going."}` +
		"\n```\nLet me know if you need more."
	archivist, mock, closeDB := arcNewCompressionArchivist(t, prosePrefixed)
	defer closeDB()

	teamID := "arc-team-prose-prefixed"
	events := arcBufferedEvents(teamID)

	// No INSERT expectation registered: compressAndStore must not persist
	// the raw prose-wrapped text as the summary.
	archivist.compressAndStore(context.Background(), teamID, events)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unexpected DB interaction: %v", err)
	}
}

// TestArcCompressAndStore_PlainJSONResponse_ExtractsAndStores is the
// positive baseline: a response that is exactly JSON with no fence or prose
// must parse and persist its "summary" field.
func TestArcCompressAndStore_PlainJSONResponse_ExtractsAndStores(t *testing.T) {
	raw := `{"summary": "All quiet.", "key_events": [], "strategies": "None needed."}`
	archivist, mock, closeDB := arcNewCompressionArchivist(t, raw)
	defer closeDB()

	teamID := "arc-team-plain"
	events := arcBufferedEvents(teamID)

	mock.ExpectExec("INSERT INTO sitreps").
		WithArgs(
			teamID,
			anyArg(),
			anyArg(),
			"All quiet.",
			anyArg(),
			"None needed.",
		).
		WillReturnResult(sqlmock.NewResult(1, 1))

	archivist.compressAndStore(context.Background(), teamID, events)

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unmet DB expectations: %v", err)
	}
}
