# Implementation Directive: Memory Service (Phase 2B)
**Target:** `mycelis-core` (Go)
**Goal:** Transform the Core from a stateless router into a "State Engine" by projecting NATS events into PostgreSQL.
**Security Level:** High (No raw SQL strings, usage of Env Vars).

---

## 1. The Cortex Schema (Migration)
Create the database structure that supports both the immutable event log and the live state registry.

**File:** `core/migrations/001_init_memory.sql`

```sql
-- Enable Vector Search for future Intelligence phases
CREATE EXTENSION IF NOT EXISTS vector;

-- 1. The "Log" (Immutable Event History)
-- Optimized for high-volume time-series writes.
CREATE TABLE IF NOT EXISTS log_entries (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    trace_id TEXT NOT NULL,
    timestamp TIMESTAMPTZ NOT NULL,
    level TEXT NOT NULL,       -- INFO, WARN, ERROR
    source TEXT NOT NULL,      -- agent:finance
    intent TEXT NOT NULL,      -- spend_money
    message TEXT,
    context JSONB,             -- The full payload snapshot
    created_at TIMESTAMPTZ DEFAULT NOW()
);

-- Optimize for Cortex UI Stream retrieval
CREATE INDEX IF NOT EXISTS idx_logs_time ON log_entries(timestamp DESC);
CREATE INDEX IF NOT EXISTS idx_logs_source ON log_entries(source);

-- 2. The "State" (Current Projected Reality)
-- A unified registry of who is alive and what they are doing.
CREATE TABLE IF NOT EXISTS agent_registry (
    agent_id TEXT PRIMARY KEY,
    team_id TEXT NOT NULL,
    status TEXT NOT NULL,      -- IDLE, BUSY, OFFLINE, ERROR
    last_seen TIMESTAMPTZ NOT NULL,
    meta JSONB,                -- Battery, Location, Current Task
    last_error_summary TEXT    -- For self-healing logic
);
```

## 2. The Memory Service Package (Write Logic)
Implement the worker that listens to the stream and writes to the DB without blocking the reflex loop.

**File:** `core/internal/memory/service.go`

```go
package memory

import (
    "context"
    "database/sql"
    "encoding/json"
    "time"

    _ "github.com/jackc/pgx/v5/stdlib" // Use pgx driver
    pb "github.com/mycelis/proto/swarm/v1"
)

// Service manages the projection of stream events to state.
type Service struct {
    db     *sql.DB
    events chan *pb.LogEntry // Buffered channel to prevent blocking
}

func NewService(dbUrl string) (*Service, error) {
    db, err := sql.Open("pgx", dbUrl)
    if err != nil {
        return nil, err
    }
    
    // Validate connection
    if err := db.Ping(); err != nil {
        return nil, err
    }

    return &Service{
        db:     db,
        events: make(chan *pb.LogEntry, 1000), // Buffer 1000 logs
    }, nil
}

// Push adds an event to the processing queue non-blocking.
func (s *Service) Push(entry *pb.LogEntry) {
    select {
    case s.events <- entry:
        // Queued
    default:
        // Buffer full: Log error to stderr or drop (Do not crash Core)
        // In Prod, we might want a secondary fallback logger.
    }
}

// Start begins the projection loop.
func (s *Service) Start(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        case entry := <-s.events:
            s.persist(entry)
        }
    }
}

func (s *Service) persist(entry *pb.LogEntry) {
    // 1. Insert Log
    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    contextJSON, _ := json.Marshal(entry.Context)

    _, err := s.db.ExecContext(ctx, 
        `INSERT INTO log_entries (trace_id, timestamp, level, source, intent, message, context) 
         VALUES ($1, $2, $3, $4, $5, $6, $7)`,
        entry.TraceId, time.Now(), entry.Level, entry.Source, entry.Intent, entry.Message, contextJSON,
    )
    if err != nil {
        // Handle DB error
    }

    // 2. Upsert Registry (Live State)
    // "On Conflict" logic updates the 'last_seen' timestamp automatically.
    _, err = s.db.ExecContext(ctx,
        `INSERT INTO agent_registry (agent_id, team_id, status, last_seen)
         VALUES ($1, $2, $3, NOW())
         ON CONFLICT (agent_id) DO UPDATE 
         SET last_seen = NOW(), status = EXCLUDED.status`,
         entry.Source, "default", "ACTIVE", // You might parse TeamID from Source string
    )
}
```

## 3. The Retrieval API (Read Logic)
Expose the memory to the Cortex UI.

**File:** `core/internal/server/memory.go` (Update existing handlers)

```go
// GetMemoryStream handles GET /api/v1/memory/stream
func (s *Server) GetMemoryStream(w http.ResponseWriter, r *http.Request) {
    limit := 50
    rows, err := s.mem.db.Query("SELECT * FROM log_entries ORDER BY timestamp DESC LIMIT $1", limit)
    if err != nil {
        http.Error(w, "Memory access failure", 500)
        return
    }
    defer rows.Close()
    
    // Marshal rows to JSON and return...
}
```

## 4. Wiring (The Cortex Connection)
Connect the Database and start the Service in the main loop.

**File:** `core/cmd/server/main.go`

```go
func main() {
    // ... config loading ...

    // 1. Connect to Cortex Memory (Postgres)
    dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_HOST"), // Value: mycelis-pg
        os.Getenv("DB_PORT"),
        os.Getenv("DB_NAME"),
    )
    
    memService, err := memory.NewService(dbURL)
    if err != nil {
        log.Fatalf("❌ Memory Corruption: %v", err)
    }

    // 2. Start Projection Routine (Background)
    go memService.Start(context.Background())

    // 3. Start NATS Subscriber
    nc.Subscribe("swarm.>", func(msg *nats.Msg) {
        // ... unmarshal logic ...
        
        // A. The Reflex (Router)
        router.Dispatch(entry)

        // B. The Memory (Service)
        memService.Push(entry)
    })
    
    // ... start HTTP server ...
}
```

## 5. Verification Plan
1.  **Deploy:** `inv k8s.deploy`
2.  **Generate Traffic:** Run the bridge and send a test message.
3.  **Verify State:**
    ```bash
    # Shell into the Database
    kubectl exec -it mycelis-pg-0 -- psql -U postgres -c "SELECT * FROM agent_registry;"
    ```
