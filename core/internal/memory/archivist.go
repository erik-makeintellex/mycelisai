package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

// ── Daemon Types ──────────────────────────────────────────────

// BufferedEvent is a normalized trace event for the Archivist's sliding window.
type BufferedEvent struct {
	Source    string    `json:"source"`
	Signal    string    `json:"signal"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}

// eventBuffer is a thread-safe sliding window accumulator keyed by team ID.
type eventBuffer struct {
	mu        sync.Mutex
	buckets   map[string][]BufferedEvent
	threshold int
}

func newEventBuffer(threshold int) *eventBuffer {
	return &eventBuffer{
		buckets:   make(map[string][]BufferedEvent),
		threshold: threshold,
	}
}

// push adds an event to the named bucket.
// Returns (teamID, events, true) when the bucket should be flushed.
func (b *eventBuffer) push(teamID string, event BufferedEvent) (string, []BufferedEvent, bool) {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.buckets[teamID] = append(b.buckets[teamID], event)

	// Flush conditions: threshold reached OR artifact/output signal received
	if len(b.buckets[teamID]) >= b.threshold ||
		event.Signal == "artifact" || event.Signal == "output" ||
		event.Signal == string(protocol.SignalTaskComplete) {
		events := b.buckets[teamID]
		delete(b.buckets, teamID)
		return teamID, events, true
	}

	return "", nil, false
}

// Drain returns all non-empty buckets and clears the buffer.
func (b *eventBuffer) Drain() map[string][]BufferedEvent {
	b.mu.Lock()
	defer b.mu.Unlock()

	out := b.buckets
	b.buckets = make(map[string][]BufferedEvent)
	return out
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
		"team_id":           teamID,
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

// ── Phase 5.0: The Archivist Daemon ──────────────────────────

const defaultBufferThreshold = 20

// StartDaemon subscribes to the NATS audit trace, buffers CTSEnvelopes per team,
// and flushes through the LLM compressor when the threshold (20 events) is reached
// or an artifact/output signal arrives. It blocks until ctx is cancelled.
func (a *Archivist) StartDaemon(ctx context.Context, nc *nats.Conn, defaultTeamID string) error {
	buffer := newEventBuffer(defaultBufferThreshold)

	sub, err := nc.Subscribe(protocol.TopicAuditTrace, func(msg *nats.Msg) {
		event := parseTraceEvent(msg.Data)
		if event == nil {
			return
		}

		teamID := defaultTeamID
		if teamID == "" {
			teamID = event.Source // Fallback: group by source
		}

		_, events, flush := buffer.push(teamID, *event)
		if flush {
			go a.compressAndStore(ctx, teamID, events)
		}
	})
	if err != nil {
		return fmt.Errorf("archivist daemon: subscribe failed: %w", err)
	}

	log.Printf("Archivist Daemon Active (subscribed to %s, threshold=%d)", protocol.TopicAuditTrace, defaultBufferThreshold)

	// Block until shutdown, then drain remaining buffers
	<-ctx.Done()
	sub.Unsubscribe()

	// Flush any remaining buffered events before exit
	remaining := buffer.Drain()
	for teamID, events := range remaining {
		if len(events) > 0 {
			log.Printf("Archivist Daemon: Flushing %d remaining events for %s", len(events), teamID)
			a.compressAndStore(context.Background(), teamID, events)
		}
	}

	log.Println("Archivist Daemon Stopped.")
	return nil
}

// parseTraceEvent normalizes various NATS message formats into a BufferedEvent.
// Supports both CTSEnvelope (signal_type) and generic trace JSON (signal).
func parseTraceEvent(data []byte) *BufferedEvent {
	// Try CTSEnvelope first (the canonical V6.2 format — requires signal_type)
	var cts protocol.CTSEnvelope
	if err := json.Unmarshal(data, &cts); err == nil && cts.Meta.SourceNode != "" && cts.SignalType != "" {
		content := extractPayloadContent(cts.Payload)
		return &BufferedEvent{
			Source:    cts.Meta.SourceNode,
			Signal:    string(cts.SignalType),
			Content:   content,
			Timestamp: cts.Meta.Timestamp,
		}
	}

	// Fallback: generic trace JSON (e.g. {"meta":{"source_node":"..."}, "signal":"thought", "payload":{...}})
	var generic struct {
		Meta struct {
			SourceNode string `json:"source_node"`
		} `json:"meta"`
		Signal  string          `json:"signal"`
		Payload json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(data, &generic); err == nil && generic.Meta.SourceNode != "" {
		content := extractPayloadContent(generic.Payload)
		return &BufferedEvent{
			Source:    generic.Meta.SourceNode,
			Signal:    generic.Signal,
			Content:   content,
			Timestamp: time.Now(),
		}
	}

	return nil
}

// extractPayloadContent pulls the "content" string from a JSON payload.
func extractPayloadContent(payload json.RawMessage) string {
	if len(payload) == 0 {
		return ""
	}
	var p map[string]interface{}
	if json.Unmarshal(payload, &p) == nil {
		if c, ok := p["content"].(string); ok {
			return c
		}
	}
	return string(payload)
}

// compressAndStore flushes the buffered events through the LLM and persists the SitRep.
func (a *Archivist) compressAndStore(ctx context.Context, teamID string, events []BufferedEvent) {
	if len(events) == 0 {
		return
	}

	windowStart := events[0].Timestamp
	windowEnd := events[len(events)-1].Timestamp
	if windowStart.IsZero() {
		windowStart = time.Now()
	}
	if windowEnd.IsZero() {
		windowEnd = time.Now()
	}

	// Build log digest for the LLM
	var digest strings.Builder
	for _, e := range events {
		ts := e.Timestamp.Format(time.RFC3339)
		if e.Timestamp.IsZero() {
			ts = "now"
		}
		digest.WriteString(fmt.Sprintf("[%s] %s (%s): %s\n", ts, e.Source, e.Signal, e.Content))
	}

	// Compress via Cognitive Engine
	prompt := fmt.Sprintf(`You are the Archivist. Compress these raw system logs into a strict, 3-sentence Situation Report (SitRep). Focus ONLY on actionable outcomes, current state, and blockers.

Source: %s
Event Count: %d
Window: %s to %s

LOGS:
%s
FORMAT (JSON):
{"summary": "3-sentence SitRep", "key_events": ["event1", "event2"], "strategies": "Next steps"}`,
		teamID,
		len(events),
		windowStart.Format(time.RFC3339),
		windowEnd.Format(time.RFC3339),
		digest.String(),
	)

	req := cognitive.InferRequest{
		Profile: "architect",
		Prompt:  prompt,
	}

	resp, err := a.Cog.InferWithContract(ctx, req)
	if err != nil {
		log.Printf("Archivist Daemon: Compression failed for %s: %v", teamID, err)
		return
	}

	// Parse the LLM response
	var sitrep struct {
		Summary    string   `json:"summary"`
		KeyEvents  []string `json:"key_events"`
		Strategies string   `json:"strategies"`
	}

	cleanText := extractJSON(resp.Text)
	if err := json.Unmarshal([]byte(cleanText), &sitrep); err != nil {
		log.Printf("Archivist Daemon: JSON parse failed: %v. Storing raw.", err)
		sitrep.Summary = resp.Text
	}

	// Persist to sitreps table
	eventsJSON, _ := json.Marshal(sitrep.KeyEvents)

	_, err = a.Mem.db.ExecContext(ctx, `
		INSERT INTO sitreps (team_id, time_window_start, time_window_end, summary, key_events, strategies)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, teamID, windowStart, windowEnd, sitrep.Summary, eventsJSON, sitrep.Strategies)

	if err != nil {
		log.Printf("Archivist Daemon: DB write failed for %s: %v", teamID, err)
		return
	}

	log.Printf("Archivist: SitRep compressed for %s (%d events -> %d chars)", teamID, len(events), len(sitrep.Summary))

	// Auto-embed the compressed SitRep for RAG
	a.embedSitRep(ctx, teamID, sitrep.Summary, windowStart, windowEnd)
}

// extractJSON strips markdown code fences from LLM output to extract raw JSON.
func extractJSON(text string) string {
	// Strip ```json ... ``` wrapping
	text = strings.TrimSpace(text)
	if strings.HasPrefix(text, "```") {
		// Find closing ```
		if idx := strings.Index(text[3:], "```"); idx >= 0 {
			text = text[3 : idx+3]
			// Remove language tag on first line (e.g. "json\n")
			if nl := strings.IndexByte(text, '\n'); nl >= 0 {
				text = text[nl+1:]
			}
		}
	}
	return strings.TrimSpace(text)
}
