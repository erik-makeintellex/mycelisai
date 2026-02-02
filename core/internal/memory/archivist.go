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

// Archivist manages the projection of stream events to state.
type Archivist struct {
	db     *sql.DB
	events chan *LogEntry // Buffered channel to prevent blocking
}

func NewArchivist(dbUrl string) (*Archivist, error) {
	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		return nil, err
	}

	// Validate connection with Retry (Stabilization)
	var pingErr error
	for i := 0; i < 30; i++ {
		pingErr = db.Ping()
		if pingErr == nil {
			log.Println("✅ Archivist: Connected to Hippocampus (Postgres)")
			break
		}
		log.Printf("⏳ Archivist: Waiting for Hippocampus... (%d/30) - %v", i+1, pingErr)
		time.Sleep(1 * time.Second)
	}

	if pingErr != nil {
		return nil, pingErr // Fatal after 30s
	}

	return &Archivist{
		db:     db,
		events: make(chan *LogEntry, 1000), // Buffer 1000 logs
	}, nil
}

// Push adds an event to the processing queue non-blocking.
func (a *Archivist) Push(entry *LogEntry) {
	select {
	case a.events <- entry:
		// Queued
	default:
		// Buffer full: Log error to stderr or drop (Do not crash Core)
		log.Println("⚠️ Archivist Buffer Full: Dropping Event")
	}
}

// Start begins the projection loop.
func (a *Archivist) Start(ctx context.Context) {
	log.Println("🧠 Archivist: Projection Loop Started")
	for {
		select {
		case <-ctx.Done():
			return
		case entry := <-a.events:
			a.persist(entry)
		}
	}
}

func (a *Archivist) persist(entry *LogEntry) {
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

	_, err := a.db.ExecContext(ctx,
		`INSERT INTO log_entries (trace_id, timestamp, level, source, intent, message, context) 
         VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		entry.TraceId, ts, entry.Level, entry.Source, entry.Intent, entry.Message, contextJSON,
	)
	if err != nil {
		log.Printf("❌ Archivist Save Error: %v", err)
	}

	// 2. Upsert Registry (Live State)
	// "On Conflict" logic updates the 'last_seen' timestamp automatically.
	_, err = a.db.ExecContext(ctx,
		`INSERT INTO agent_registry (agent_id, team_id, status, last_seen)
         VALUES ($1, $2, $3, NOW())
         ON CONFLICT (agent_id) DO UPDATE 
         SET last_seen = NOW(), status = EXCLUDED.status`,
		entry.Source, "default", "ACTIVE", // Default Team/Status for now
	)
	if err != nil {
		log.Printf("❌ Registry Update Error: %v", err)
	}
}

// ListRecent retrieves the last N logs.
func (a *Archivist) ListRecent(limit int) ([]*LogEntry, error) {
	if limit <= 0 {
		limit = 50
	}

	rows, err := a.db.Query(`SELECT trace_id, timestamp, level, source, intent, message, context 
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
