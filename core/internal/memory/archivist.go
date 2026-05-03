package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/mycelis/core/internal/cognitive"
)

// ── Daemon Types ──────────────────────────────────────────────

// BufferedEvent is a normalized trace event for the Archivist's sliding window.
type BufferedEvent struct {
	Source    string    `json:"source"`
	Signal    string    `json:"signal"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// Archivist manages high-level context and reflection
type Archivist struct {
	Mem *Service
	Cog *cognitive.Router
}

func NewArchivist(mem *Service, cog *cognitive.Router) *Archivist {
	return &Archivist{
		Mem: mem,
		Cog: cog,
	}
}

// GenerateSitRep summarizes the last window of time for a given team
func (a *Archivist) GenerateSitRep(ctx context.Context, teamID string, duration time.Duration) error {
	end := time.Now()
	start := end.Add(-duration)

	// 1. Retrieve Logs
	logs, err := a.Mem.ListLogs(ctx, start, end, 500) // Cap at 500 logs for context window
	if err != nil {
		return fmt.Errorf("failed to fetch logs: %w", err)
	}

	if len(logs) == 0 {
		return fmt.Errorf("no logs found for window")
	}

	// 2. Compress Context
	logSummary := ""
	for _, l := range logs {
		logSummary += fmt.Sprintf("[%s] %s: %s (%s)\n", l.Timestamp.Format(time.RFC3339), l.Source, l.Message, l.Intent)
	}

	// 3. Prompt Cognitive Engine
	prompt := fmt.Sprintf(`
Analyze the following system logs for Team %s and generate a SITREP (Situation Report).
Time Window: %s to %s

LOGS:
%s

FORMAT (JSON):
{
	"summary": "High level narrative of what happened.",
	"key_events": ["list of critical check points or errors"],
	"strategies": "Recommendation for next steps."
}
	`, teamID, start.Format(time.RFC3339), end.Format(time.RFC3339), logSummary)

	req := cognitive.InferRequest{
		Profile: "architect", // Use the Architect profile (likely stronger model)
		Prompt:  prompt,
	}

	// Make Schema Enforcement Explicit?
	// For now we trust the Architect to follow instructions.

	resp, err := a.Cog.InferWithContract(ctx, req)
	if err != nil {
		// Log error but treat as "brain fog"
		log.Printf("Archivist: Brain Fog (Inference Failed): %v", err)
		return err
	}

	// 4. Parse Response
	var sitrep struct {
		Summary    string   `json:"summary"`
		KeyEvents  []string `json:"key_events"`
		Strategies string   `json:"strategies"`
	}

	// Try to unmarshal JSON from text (it might be wrapped in Markdown ```json ... ```)
	// Simple cleanup: find { and }
	cleanText := resp.Text // TODO: robust JSON extractor

	if err := json.Unmarshal([]byte(cleanText), &sitrep); err != nil {
		log.Printf("Archivist: Failed to parse SitRep JSON: %v. Raw: %s", err, resp.Text)
		// Fallback: Store raw text as summary
		sitrep.Summary = resp.Text
	}

	// 5. Save to DB
	eventsJSON, _ := json.Marshal(sitrep.KeyEvents)

	_, err = a.Mem.db.ExecContext(ctx, `
		INSERT INTO sitreps (team_id, time_window_start, time_window_end, summary, key_events, strategies)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, teamID, start, end, sitrep.Summary, eventsJSON, sitrep.Strategies)

	if err != nil {
		return fmt.Errorf("failed to save sitrep: %w", err)
	}

	log.Printf("Archivist Generated SitRep for %s", teamID)

	// Auto-embed the SitRep summary into context_vectors for RAG retrieval
	a.embedSitRep(ctx, teamID, sitrep.Summary, start, end)

	return nil
}

// embedSitRep generates a vector embedding of a SitRep summary and stores it in context_vectors.
// Failures are logged but never propagate — embedding is best-effort enrichment.
func (a *Archivist) embedSitRep(ctx context.Context, teamID string, summary string, windowStart, windowEnd time.Time) {
	if summary == "" || a.Cog == nil {
		return
	}

	vec, err := a.Cog.Embed(ctx, summary, "")
	if err != nil {
		log.Printf("Archivist: Embedding skipped (no embed provider): %v", err)
		return
	}

	metadata := map[string]any{
		"type":              "sitrep",
		"tenant_id":         "default",
		"team_id":           teamID,
		"visibility":        "team",
		"time_window_start": windowStart.Format(time.RFC3339),
		"time_window_end":   windowEnd.Format(time.RFC3339),
	}

	if err := a.Mem.StoreVector(ctx, summary, vec, metadata); err != nil {
		log.Printf("Archivist: Vector storage failed: %v", err)
		return
	}

	log.Printf("Archivist: SitRep embedded (%d dims) for %s", len(vec), teamID)
}

// StartLoop runs the Archivist's SitRep generation on a periodic interval.
// It blocks until the context is cancelled.
func (a *Archivist) StartLoop(ctx context.Context, interval time.Duration, teamID string) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	log.Printf("Archivist Loop Active (every %s for team %s)", interval, teamID)

	for {
		select {
		case <-ctx.Done():
			log.Println("Archivist Loop Stopped.")
			return
		case <-ticker.C:
			if err := a.GenerateSitRep(ctx, teamID, interval); err != nil {
				log.Printf("Archivist: SitRep generation skipped: %v", err)
			}
		}
	}
}
