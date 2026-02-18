package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // Use pgx driver
)

// LogEntry matches the SQL schema log_entries table
type LogEntry struct {
	ID        string         `json:"id"`
	TraceId   string         `json:"trace_id"`
	Timestamp time.Time      `json:"timestamp"`
	Level     string         `json:"level"`
	Source    string         `json:"source"`
	Intent    string         `json:"intent"`
	Message   string         `json:"message"`
	Context   map[string]any `json:"context"` // Generic map for JSONB
}

// Service manages the projection of stream events to state.
type Service struct {
	db     *sql.DB
	events chan *LogEntry // Buffered channel to prevent blocking
}

func NewService(dbUrl string) (*Service, error) {
	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		return nil, err
	}

	// Validate connection with Retry (Stabilization)
	var pingErr error
	for i := 0; i < 30; i++ {
		pingErr = db.Ping()
		if pingErr == nil {
			log.Println("Memory: Connected to Cortex (Postgres)")
			break
		}
		log.Printf("Memory: Waiting for Cortex... (%d/30) - %v", i+1, pingErr)
		time.Sleep(1 * time.Second)
	}

	if pingErr != nil {
		return nil, pingErr // Fatal after 30s
	}

	return &Service{
		db:     db,
		events: make(chan *LogEntry, 1000), // Buffer 1000 logs
	}, nil
}

// Push adds an event to the processing queue non-blocking.
func (s *Service) Push(entry *LogEntry) {
	select {
	case s.events <- entry:
		// Queued
	default:
		// Buffer full: Log error to stderr or drop (Do not crash Core)
		log.Println("WARN: Memory Buffer Full: Dropping Event")
	}
}

// Start begins the projection loop.
func (s *Service) Start(ctx context.Context) {
	log.Println("Memory: Projection Loop Started")
	for {
		select {
		case <-ctx.Done():
			return
		case entry := <-s.events:
			s.persist(entry)
		}
	}
}

func (s *Service) persist(entry *LogEntry) {
	// 1. Insert Log
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	contextJSON, _ := json.Marshal(entry.Context)

	// Use time.Now() for consistency or entry.Timestamp if we trust it?
	// SQL uses DEFAULT NOW(), but we pass $2.
	// If entry.Timestamp is set, use it. Else Now().
	ts := entry.Timestamp
	if ts.IsZero() {
		ts = time.Now()
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT INTO log_entries (trace_id, timestamp, level, source, intent, message, context) 
         VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		entry.TraceId, ts, entry.Level, entry.Source, entry.Intent, entry.Message, contextJSON,
	)
	if err != nil {
		log.Printf("ERROR: Memory Save Error: %v", err)
	}

	// 2. Upsert Registry (Live State)
	// "On Conflict" logic updates the 'last_seen' timestamp automatically.
	_, err = s.db.ExecContext(ctx,
		`INSERT INTO agent_registry (agent_id, team_id, status, last_seen)
         VALUES ($1, $2, $3, NOW())
         ON CONFLICT (agent_id) DO UPDATE 
         SET last_seen = NOW(), status = EXCLUDED.status`,
		entry.Source, "default", "ACTIVE", // Default Team/Status for now
	)
	if err != nil {
		log.Printf("ERROR: Registry Update Error: %v", err)
	}
}

// ListRecent retrieves the last N logs.
func (s *Service) ListRecent(limit int) ([]*LogEntry, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := s.db.Query(`SELECT trace_id, timestamp, level, source, intent, message, context 
                             FROM log_entries ORDER BY timestamp DESC LIMIT $1`, limit)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*LogEntry
	for rows.Next() {
		var entry LogEntry
		var ctxJSON []byte

		if err := rows.Scan(&entry.TraceId, &entry.Timestamp, &entry.Level, &entry.Source, &entry.Intent, &entry.Message, &ctxJSON); err != nil {
			return nil, err
		}

		if len(ctxJSON) > 0 {
			if err := json.Unmarshal(ctxJSON, &entry.Context); err != nil {
				// Log warning?
			}
		}

		logs = append(logs, &entry)
	}
	return logs, nil
}

// ListLogs retrieves logs within a time range
func (s *Service) ListLogs(ctx context.Context, start, end time.Time, limit int) ([]*LogEntry, error) {
	stmt := `SELECT trace_id, timestamp, level, source, intent, message, context 
             FROM log_entries 
             WHERE timestamp BETWEEN $1 AND $2 
             ORDER BY timestamp ASC LIMIT $3`

	rows, err := s.db.QueryContext(ctx, stmt, start, end, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []*LogEntry
	for rows.Next() {
		var entry LogEntry
		var ctxJSON []byte
		if err := rows.Scan(&entry.TraceId, &entry.Timestamp, &entry.Level, &entry.Source, &entry.Intent, &entry.Message, &ctxJSON); err != nil {
			return nil, err
		}
		if len(ctxJSON) > 0 {
			_ = json.Unmarshal(ctxJSON, &entry.Context)
		}
		logs = append(logs, &entry)
	}
	return logs, nil
}

// ── Vector Memory (pgvector) ─────────────────────────────────

// VectorResult is a single semantic search hit from context_vectors.
type VectorResult struct {
	ID        string         `json:"id"`
	Content   string         `json:"content"`
	Metadata  map[string]any `json:"metadata"`
	Score     float64        `json:"score"` // cosine similarity (1.0 = identical)
	CreatedAt time.Time      `json:"created_at"`
}

// StoreVector persists an embedding into context_vectors for future RAG retrieval.
func (s *Service) StoreVector(ctx context.Context, content string, embedding []float64, metadata map[string]any) error {
	metaJSON, _ := json.Marshal(metadata)
	vecStr := formatVector(embedding)

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO context_vectors (content, embedding, metadata)
		VALUES ($1, $2::vector, $3)
	`, content, vecStr, metaJSON)

	if err != nil {
		return fmt.Errorf("store vector failed: %w", err)
	}
	return nil
}

// SemanticSearch finds the top-K nearest vectors by cosine similarity.
func (s *Service) SemanticSearch(ctx context.Context, queryVec []float64, limit int) ([]VectorResult, error) {
	if limit <= 0 {
		limit = 5
	}

	vecStr := formatVector(queryVec)

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, content, metadata, 1 - (embedding <=> $1::vector) AS score, created_at
		FROM context_vectors
		WHERE embedding IS NOT NULL
		ORDER BY embedding <=> $1::vector
		LIMIT $2
	`, vecStr, limit)
	if err != nil {
		return nil, fmt.Errorf("semantic search failed: %w", err)
	}
	defer rows.Close()

	var results []VectorResult
	for rows.Next() {
		var r VectorResult
		var metaJSON []byte
		if err := rows.Scan(&r.ID, &r.Content, &metaJSON, &r.Score, &r.CreatedAt); err != nil {
			return nil, err
		}
		if len(metaJSON) > 0 {
			_ = json.Unmarshal(metaJSON, &r.Metadata)
		}
		results = append(results, r)
	}
	return results, nil
}

// ListSitReps retrieves recent SitReps for a team.
func (s *Service) ListSitReps(ctx context.Context, teamID string, limit int) ([]map[string]any, error) {
	if limit <= 0 {
		limit = 10
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, team_id, timestamp, time_window_start, time_window_end, summary, key_events, strategies, status
		FROM sitreps
		WHERE team_id = $1
		ORDER BY timestamp DESC
		LIMIT $2
	`, teamID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []map[string]any
	for rows.Next() {
		var id, tID, summary, status string
		var strategies sql.NullString
		var ts, winStart, winEnd time.Time
		var keyEventsJSON []byte

		if err := rows.Scan(&id, &tID, &ts, &winStart, &winEnd, &summary, &keyEventsJSON, &strategies, &status); err != nil {
			return nil, err
		}

		entry := map[string]any{
			"id":                id,
			"team_id":           tID,
			"timestamp":         ts,
			"time_window_start": winStart,
			"time_window_end":   winEnd,
			"summary":           summary,
			"strategies":        strategies.String,
			"status":            status,
		}

		var keyEvents []string
		if json.Unmarshal(keyEventsJSON, &keyEvents) == nil {
			entry["key_events"] = keyEvents
		}

		results = append(results, entry)
	}
	return results, nil
}

// formatVector converts a float64 slice to pgvector string format: "[0.1,0.2,0.3]"
func formatVector(v []float64) string {
	parts := make([]string, len(v))
	for i, f := range v {
		parts[i] = fmt.Sprintf("%g", f)
	}
	return "[" + strings.Join(parts, ",") + "]"
}
