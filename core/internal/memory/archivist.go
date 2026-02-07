package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/mycelis/core/internal/cognitive"
)

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

	log.Printf("✅ Archivist Generated SitRep for %s", teamID)
	return nil
}

// GenerateContextVector creates an embedding for a piece of content
func (a *Archivist) GenerateContextVector(ctx context.Context, content string, meta map[string]interface{}) error {
	// TODO: use cognitive.Embed() if available.
	// For now, placeholder or use "embedding" profile if implemented.
	return nil
}
