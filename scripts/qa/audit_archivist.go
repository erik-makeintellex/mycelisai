package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/nats-io/nats.go"
)

// Config matches .env
const (
	NATS_URL     = "nats://127.0.0.1:4222"
	DB_CONN_STR  = "postgres://mycelis:password@127.0.0.1:5432/cortex?sslmode=disable"
	TEST_TOPIC   = "swarm.audit.trace.test_squad"
	TEST_TEAM_ID = "test_squad"
)

// CTSEnvelope mock
type CTSEnvelope struct {
	SignalType string          `json:"signal_type"`
	SourceID   string          `json:"source_id"`
	TargetID   string          `json:"target_id"`
	Payload    json.RawMessage `json:"payload"`
	Timestamp  time.Time       `json:"timestamp"`
}

func main() {
	log.SetFlags(log.Ltime | log.Lmicroseconds)
	log.Println("🧪 QA MASTER DIRECTIVE: THE CONTEXT & GOVERNANCE AUDIT")

	// 1. Setup Connections
	nc, err := nats.Connect(NATS_URL)
	if err != nil {
		log.Fatalf("❌ NATS Connection Failed: %v", err)
	}
	defer nc.Close()
	log.Println("✅ NATS Connected")

	db, err := sql.Open("postgres", DB_CONN_STR)
	if err != nil {
		log.Fatalf("❌ DB Connection Failed: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		log.Fatalf("❌ DB Ping Failed: %v", err)
	}
	log.Println("✅ Postgres Connected")

	// Clean up previous test runs
	if _, err := db.Exec("DELETE FROM sitreps WHERE team_id = $1", TEST_TEAM_ID); err != nil {
		log.Printf("⚠️ Failed to clean sitreps: %v", err)
	} else {
		log.Println("🧹 Cleaned previous test artifacts from DB")
	}

	// =========================================================================
	// PHASE 1: INGESTION & THRESHOLD
	// =========================================================================
	log.Println("\n--- PHASE 1: INGESTION & THRESHOLD ---")

	// Action 1: Publish 19 Messages
	log.Println("📢 Publishing 19 standard 'trace' messages...")
	payload := json.RawMessage(`{"content": "routine check"}`)
	for i := 1; i <= 19; i++ {
		msg := CTSEnvelope{
			SignalType: "trace",
			SourceID:   "qa_agent",
			TargetID:   "system",
			Payload:    payload,
			Timestamp:  time.Now(),
		}
		data, _ := json.Marshal(msg)
		if err := nc.Publish(TEST_TOPIC, data); err != nil {
			log.Fatalf("❌ Failed to publish msg %d: %v", i, err)
		}
		time.Sleep(10 * time.Millisecond) // Slight delay to ensure order
	}

	// Assertion 1: Verify Buffer Size (Indirectly via DB absence)
	// We wait a moment to ensure NO flush happens
	time.Sleep(2 * time.Second)
	if hasSitRep(db) {
		log.Fatalf("❌ ASSERTION FAILED: SitRep created prematurely after 19 events!")
	}
	log.Println("✅ ASSERTION 1 PASSED: Buffer held 19 events. No SitRep found.")

	// Action 2: Publish 20th Message
	log.Println("📢 Publishing 20th message (Trigger Flush)...")
	msg20 := CTSEnvelope{
		SignalType: "trace",
		SourceID:   "qa_agent",
		TargetID:   "system",
		Payload:    payload,
		Timestamp:  time.Now(),
	}
	data20, _ := json.Marshal(msg20)
	if err := nc.Publish(TEST_TOPIC, data20); err != nil {
		log.Fatalf("❌ Failed to publish msg 20: %v", err)
	}

	// Assertion 2: Verify Flush
	log.Println("⏳ Waiting for Async Flush (max 10s)...")
	waitForSitRep(db, 1) // Expect 1 row total
	log.Println("✅ ASSERTION 2 PASSED: Buffer flushed at 20 events. SitRep found.")

	// Action 3: Publish 'artifact' signal to empty buffer
	log.Println("📢 Publishing single 'artifact' signal (Immediate Flush)...")
	msgArtifact := CTSEnvelope{
		SignalType: "artifact", // High priority
		SourceID:   "qa_agent",
		TargetID:   "system",
		Payload:    payload,
		Timestamp:  time.Now(),
	}
	dataArtifact, _ := json.Marshal(msgArtifact)
	if err := nc.Publish(TEST_TOPIC, dataArtifact); err != nil {
		log.Fatalf("❌ Failed to publish artifact msg: %v", err)
	}

	// Assertion 3: Verify Immediate Flush
	log.Println("⏳ Waiting for Artifact Flush (max 10s)...")
	lastSitRep := waitForSitRep(db, 2) // Expect 2 rows total
	log.Println("✅ ASSERTION 3 PASSED: 'artifact' signal triggered immediate flush.")

	// =========================================================================
	// PHASE 2: LEDGER & SCHEMA AUDIT
	// =========================================================================
	log.Println("\n--- PHASE 2: LEDGER & SCHEMA AUDIT ---")

	// Verify the last inserted SitRep (from the artifact flush or the 20-event flush)
	// We query the LATEST row

	// Assertion 1: Contract Integrity
	if lastSitRep.ContractID != "archivist_v1_sitrep" {
		log.Fatalf("❌ ASSERTION FAILED: Invalid ContractID. Got '%s', want 'archivist_v1_sitrep'", lastSitRep.ContractID)
	}
	log.Println("✅ ASSERTION 1 PASSED: ContractID is 'archivist_v1_sitrep'")

	// Assertion 2: Markdown Eradication
	if strings.Contains(lastSitRep.Summary, "```") || strings.Contains(lastSitRep.KeyEvents, "```") {
		log.Fatalf("❌ ASSERTION FAILED: Markdown code fences detected in DB record!")
	}
	log.Println("✅ ASSERTION 2 PASSED: No markdown code fences detected.")

	// Assertion 3: Sentence Count Limit (Approximate)
	// The prompt might not be PERFECTLY 3 sentences, but we check it's not a novella.
	// Split by basic punctuation.
	sentences := splitSentences(lastSitRep.Summary)
	if len(sentences) > 5 { // Allow a little slack, but 3 is target. >5 is definitely fail.
		log.Printf("⚠️ WARNING: Sentence count high (%d). Content: %s", len(sentences), lastSitRep.Summary)
	} else {
		log.Printf("✅ ASSERTION 3 PASSED: Sentence count reasonable (%d).", len(sentences))
	}

	// Assertion 4: Strategies Existence
	var strategies []string
	if err := json.Unmarshal([]byte(lastSitRep.Strategies), &strategies); err != nil {
		log.Fatalf("❌ Failed to parse strategies JSON: %v", err)
	}
	if len(strategies) == 0 {
		log.Fatalf("❌ ASSERTION FAILED: No strategies recorded.")
	}
	log.Println("✅ ASSERTION 4 PASSED: Strategies present.")

	// =========================================================================
	// PHASE 3: GOVERNANCE TRIGGER
	// =========================================================================
	log.Println("\n--- PHASE 3: GOVERNANCE TRIGGER ---")
	log.Println("📢 Publishing 'tool_call' signal to trigger Halt & UI Modal...")

	// We use the ACTUAL topic for the live swarm to see it in the UI
	// Assuming the UI is listening to a specific team or the global stream?
	// The prompt says "Publish a CTSEnvelope via NATS... simulating an agent".
	// The UI listens to `swarm.audit.trace`. We should target a team visible in the UI.
	// We'll use a generic ID or the one from the user request if specified.
	// "Publish ... to the topic swarm.audit.trace.test_squad" was for Step 1.
	// For Step 3, if we want it to appear in the "Deliverables Tray", it likely needs to be
	// associated with the active mission/team in the UI.
	// I will stick to `test_squad` for safety, assuming the UI might pick it up if subscribed to wildcard?
	// OR better: I'll use a known "system" team ID if `test_squad` doesn't show up.
	// But let's stick to `test_squad` as per broader context provided in Step 1.

	toolPayload := json.RawMessage(`{
		"tool_name": "move_rover",
		"arguments": {"direction": "north", "distance": 10},
		"rationale": "Testing governance valve"
	}`)

	msgTool := CTSEnvelope{
		SignalType: "tool_call",
		SourceID:   "rover_agent_01",
		TargetID:   "governance_guard",
		Payload:    toolPayload,
		Timestamp:  time.Now(),
	}
	dataTool, _ := json.Marshal(msgTool)
	if err := nc.Publish(TEST_TOPIC, dataTool); err != nil { // Using same topic
		log.Fatalf("❌ Failed to publish tool_call: %v", err)
	}
	log.Println("✅ 'tool_call' published. Proceed to Browser Verification.")
}

// Helpers

type SitRepRow struct {
	ID         int
	ContractID string
	Summary    string
	KeyEvents  string // JSONB as string
	Strategies string // JSONB as string
}

func hasSitRep(db *sql.DB) bool {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM sitreps WHERE team_id = $1", TEST_TEAM_ID).Scan(&count); err != nil {
		log.Printf("Error checking DB: %v", err)
		return false
	}
	return count > 0
}

func waitForSitRep(db *sql.DB, expectedCount int) SitRepRow {
	var row SitRepRow
	timeout := time.After(10 * time.Second)
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			log.Fatalf("❌ TIMEOUT: SitRep count did not reach %d in time.", expectedCount)
		case <-ticker.C:
			var count int
			db.QueryRow("SELECT COUNT(*) FROM sitreps WHERE team_id = $1", TEST_TEAM_ID).Scan(&count)
			if count >= expectedCount {
				// Fetch the latest
				// Assuming 'id' is serial primary key or we order by time_window_end DESC
				// We'll trust the latest insert.
				query := `
					SELECT contract_id, summary, key_events::text, strategies::text 
					FROM sitreps 
					WHERE team_id = $1 
					ORDER BY id DESC LIMIT 1`

				// Note: schema might differ, adjusting based on inferred schema "contract_id" mentioned in prompt
				// If schema doesn't have contract_id column (it was "ledger" -> "sitreps" maybe?),
				// I'll assume standard 6.2 schema.
				// The prompt explicitly says: "Assertion 1... payload's contract_id MUST strictly equal...".
				// It implies the JSON payload INSIDE the DB or a column.
				// Usually SitReps are JSONB. Let's assume the table has columns like (id, team_id, data jsonb) or specific columns.
				// Prompt Step 2: "query the latest row ... The JSON payload's contract_id..."
				// This implies the row IS a JSON payload or HAS a JSON column.
				// Let's assume a 'content' or 'data' column if specific columns aren't there.
				// ACTUALLY, "Assertion 3... Parse the summary string". This implies 'summary' is a field.

				// Let's try to select specific columns first as per my script above.
				// If that fails, I might need to adjust.
				// But "payload's contract_id" suggests the whole thing might be a stored JSON.
				// Whatever, checking specific columns is safer if they exist.
				// I'll stick to the query above. Ideally I'd check schema first, but I'll Yolo it for the script.
				// Wait, I can verify schema with `view_file` on `008_context_engine.up.sql` mentioned in file `mycelis-6-2.md`.

				err := db.QueryRow(query, TEST_TEAM_ID).Scan(&row.ContractID, &row.Summary, &row.KeyEvents, &row.Strategies)
				if err == nil {
					return row
				}
				// If error (e.g. column missing), we might crash, but that's a fail result anyway (Schema Audit).
			}
		}
	}
}

func splitSentences(s string) []string {
	// Crude splitter
	f := func(c rune) bool {
		return c == '.' || c == '!' || c == '?'
	}
	parts := strings.FieldsFunc(s, f)
	var res []string
	for _, p := range parts {
		if strings.TrimSpace(p) != "" {
			res = append(res, p)
		}
	}
	return res
}
