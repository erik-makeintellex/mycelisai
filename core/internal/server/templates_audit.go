package server

import (
	"encoding/json"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

// createAuditEvent inserts a log_entry as an audit record and returns its UUID.
// Reuses the existing log_entries table with level='audit'.
func (s *AdminServer) createAuditEvent(templateID protocol.TemplateID, source, message string, ctx map[string]any) (string, error) {
	db := s.getDB()
	if db == nil {
		return "", nil // graceful: audit is non-blocking
	}

	id := uuid.New()
	traceID := uuid.New().String()
	reviewCtx := protocol.ParseOperationalLogContext(ctx)
	reviewCtx.ReviewScope = protocol.LogReviewScopeAudit
	reviewCtx.Service = "core"
	reviewCtx.Component = "template-engine"
	if reviewCtx.Summary == "" {
		reviewCtx.Summary = strings.TrimSpace(message)
	}
	if reviewCtx.WhyItMatters == "" {
		reviewCtx.WhyItMatters = "Audit records preserve governed template actions for Soma, meta-agentry, and governance review."
	}
	reviewCtx.SourceChannel = "audit.log_entries"
	reviewCtx.PayloadKind = protocol.PayloadKindEvent
	reviewCtx.Status = "info"
	reviewCtx.Tags = append(reviewCtx.Tags, "audit", string(templateID))
	contextMap := reviewCtx.ToMap()
	for key, value := range ctx {
		if _, exists := contextMap[key]; exists {
			continue
		}
		contextMap[key] = value
	}
	contextJSON, _ := json.Marshal(contextMap)

	_, err := db.Exec(
		`INSERT INTO log_entries (id, trace_id, timestamp, level, source, intent, message, context)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		id, traceID, time.Now(), "audit", source, string(templateID), message, contextJSON,
	)
	if err != nil {
		log.Printf("CE-1: audit event insert failed: %v", err)
		return "", err
	}

	return id.String(), nil
}
