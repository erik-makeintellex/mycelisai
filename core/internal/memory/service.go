package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
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
