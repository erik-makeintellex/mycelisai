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

// Phase 5.0: The Archivist Daemon.

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
	// Try CTSEnvelope first (the canonical V6.2 format; requires signal_type)
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
