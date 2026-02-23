// Package events provides the V7 persistent event store for mission audit trails.
// DB-first rule: Emit() persists to mission_events BEFORE publishing the CTS signal.
// If NATS is offline, events still persist (degraded mode — no data loss, no panic).
package events

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
	"github.com/nats-io/nats.go"
)

// Store persists mission events to the database and optionally publishes CTS signals.
// Implements protocol.EventEmitter.
type Store struct {
	db *sql.DB
	nc *nats.Conn // optional; nil = NATS offline (degraded mode — events still persist)
}

// NewStore creates a new events Store. nc may be nil (NATS is optional).
func NewStore(db *sql.DB, nc *nats.Conn) *Store {
	return &Store{db: db, nc: nc}
}

// Emit persists an event to mission_events and optionally publishes a CTS signal.
// This is the implementation of protocol.EventEmitter.
//
// Guarantee: DB insert happens first. CTS publish is best-effort in a goroutine.
// If NATS is offline, the event still persists with a degraded-mode log warning.
func (s *Store) Emit(ctx context.Context, runID string, eventType protocol.EventType, severity protocol.EventSeverity,
	sourceAgent, sourceTeam string, payload map[string]interface{}) (string, error) {

	if s.db == nil {
		return "", fmt.Errorf("events: database not available")
	}
	if runID == "" {
		return "", fmt.Errorf("events: run_id is required")
	}

	id := uuid.New().String()
	now := time.Now()

	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		payloadJSON = []byte("{}")
	}

	_, err = s.db.ExecContext(ctx, `
		INSERT INTO mission_events
			(id, run_id, tenant_id, event_type, severity, source_agent, source_team, payload, emitted_at)
		VALUES ($1, $2, $3, $4, $5, NULLIF($6,''), NULLIF($7,''), $8, $9)
	`, id, runID, "default", string(eventType), string(severity),
		sourceAgent, sourceTeam, payloadJSON, now)
	if err != nil {
		return "", fmt.Errorf("events: persist failed: %w", err)
	}

	// CTS publish: non-blocking, best-effort. Failures are logged, never fatal.
	if s.nc != nil {
		go s.publishCTS(runID, id, eventType, sourceAgent, payload)
	} else {
		log.Printf("[events] NATS offline: event %s (%s) persisted to DB only (degraded mode)", id, eventType)
	}

	return id, nil
}

// publishCTS publishes a lightweight CTS signal referencing the persisted event.
// The CTS payload does NOT duplicate all fields — it carries mission_event_id for linkage.
func (s *Store) publishCTS(runID, eventID string, eventType protocol.EventType, sourceAgent string, payload map[string]interface{}) {
	topic := fmt.Sprintf(protocol.TopicMissionEventsFmt, runID)
	msg := map[string]interface{}{
		"mission_event_id": eventID,
		"run_id":           runID,
		"event_type":       string(eventType),
		"source_agent":     sourceAgent,
		"payload_summary":  summarizePayload(payload),
	}
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("[events] CTS marshal error for event %s: %v", eventID, err)
		return
	}
	if err := s.nc.Publish(topic, data); err != nil {
		log.Printf("[events] CTS publish failed for event %s: %v", eventID, err)
	}
}

// summarizePayload extracts a short summary of payload keys for CTS (avoid large payloads on NATS).
func summarizePayload(payload map[string]interface{}) map[string]interface{} {
	if len(payload) == 0 {
		return nil
	}
	keys := make([]string, 0, len(payload))
	for k := range payload {
		keys = append(keys, k)
	}
	return map[string]interface{}{"keys": keys}
}

// GetRunTimeline returns all events for a run in chronological order.
// Used by GET /api/v1/runs/{id}/events.
func (s *Store) GetRunTimeline(ctx context.Context, runID string) ([]protocol.MissionEventEnvelope, error) {
	if s.db == nil {
		return nil, fmt.Errorf("events: database not available")
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, run_id, tenant_id, event_type, severity,
		       COALESCE(source_agent, ''), COALESCE(source_team, ''),
		       COALESCE(payload, '{}'), COALESCE(audit_event_id::text, ''), emitted_at
		FROM mission_events
		WHERE run_id = $1
		ORDER BY emitted_at ASC
	`, runID)
	if err != nil {
		return nil, fmt.Errorf("events: query failed: %w", err)
	}
	defer rows.Close()

	var events []protocol.MissionEventEnvelope
	for rows.Next() {
		var ev protocol.MissionEventEnvelope
		var payloadJSON []byte
		var evType, severity string
		if err := rows.Scan(
			&ev.ID, &ev.RunID, &ev.TenantID,
			&evType, &severity,
			&ev.SourceAgent, &ev.SourceTeam,
			&payloadJSON, &ev.AuditEventID, &ev.EmittedAt,
		); err != nil {
			log.Printf("[events] scan error: %v", err)
			continue
		}
		ev.EventType = protocol.EventType(evType)
		ev.Severity = protocol.EventSeverity(severity)
		if len(payloadJSON) > 2 { // skip "{}"
			json.Unmarshal(payloadJSON, &ev.Payload)
		}
		events = append(events, ev)
	}

	if events == nil {
		events = []protocol.MissionEventEnvelope{}
	}
	return events, nil
}
