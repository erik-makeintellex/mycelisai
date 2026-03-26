package server

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) handleListAuditLog(w http.ResponseWriter, r *http.Request) {
	db := s.getDB()
	if db == nil {
		respondAPIError(w, "database not available", http.StatusServiceUnavailable)
		return
	}

	limit := 20
	if raw := strings.TrimSpace(r.URL.Query().Get("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	rows, err := db.Query(
		`SELECT id, COALESCE(intent, ''), COALESCE(source, ''), COALESCE(message, ''), timestamp, COALESCE(context, '{}'::jsonb)
		 FROM log_entries
		 WHERE level = 'audit'
		 ORDER BY timestamp DESC
		 LIMIT $1`,
		limit,
	)
	if err != nil {
		respondAPIError(w, "failed to load audit log", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	records := []protocol.AuditRecord{}
	for rows.Next() {
		var (
			id         string
			templateID string
			source     string
			message    string
			ts         time.Time
			contextRaw []byte
		)
		if err := rows.Scan(&id, &templateID, &source, &message, &ts, &contextRaw); err != nil {
			continue
		}
		records = append(records, buildAuditRecord(id, templateID, source, message, ts, contextRaw))
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(records))
}

func buildAuditRecord(id, templateID, source, message string, ts time.Time, contextRaw []byte) protocol.AuditRecord {
	ctx := map[string]any{}
	if len(contextRaw) > 0 {
		_ = json.Unmarshal(contextRaw, &ctx)
	}

	record := protocol.AuditRecord{
		ID:             strings.TrimSpace(id),
		TemplateID:     strings.TrimSpace(templateID),
		Actor:          firstNonEmptyString(ctx["actor"], source, "Soma"),
		User:           firstNonEmptyString(ctx["user"], "local-user"),
		Action:         firstNonEmptyString(ctx["action"], message),
		Timestamp:      ts.Format(time.RFC3339),
		ResultStatus:   firstNonEmptyString(ctx["result_status"], "completed"),
		ApprovalStatus: firstNonEmptyString(ctx["approval_status"]),
		ApprovalReason: firstNonEmptyString(ctx["approval_reason"]),
		RunID:          firstNonEmptyString(ctx["run_id"]),
		IntentProofID:  firstNonEmptyString(ctx["intent_proof_id"], ctx["proof_id"]),
		Resource:       firstNonEmptyString(ctx["resource"]),
		CapabilityUsed: firstNonEmptyString(ctx["capability_used"]),
	}

	if details, ok := ctx["details"].(map[string]any); ok && len(details) > 0 {
		record.Details = details
	}
	return record
}

func firstNonEmptyString(values ...any) string {
	for _, value := range values {
		switch typed := value.(type) {
		case string:
			if trimmed := strings.TrimSpace(typed); trimmed != "" {
				return trimmed
			}
		}
	}
	return ""
}

func (s *AdminServer) HandleCancelAction(w http.ResponseWriter, r *http.Request) {
	var req struct {
		IntentProofID string `json:"intent_proof_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	if strings.TrimSpace(req.IntentProofID) == "" {
		respondAPIError(w, "Missing intent_proof_id", http.StatusBadRequest)
		return
	}

	db := s.getDB()
	if db == nil {
		respondAPIError(w, "database not available", http.StatusServiceUnavailable)
		return
	}

	proofUUID, err := uuid.Parse(req.IntentProofID)
	if err != nil {
		respondAPIError(w, "invalid intent_proof_id", http.StatusBadRequest)
		return
	}

	result, err := db.Exec(`UPDATE intent_proofs SET status = 'cancelled' WHERE id = $1`, proofUUID)
	if err != nil {
		respondAPIError(w, "failed to cancel proposal", http.StatusInternalServerError)
		return
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		respondAPIError(w, "intent proof not found", http.StatusNotFound)
		return
	}

	auditID, _ := s.createAuditEvent(
		protocol.TemplateChatToProposal, "cancel-action",
		"Chat proposal cancelled",
		map[string]any{
			"actor":           "Soma",
			"user":            auditUserLabelFromRequest(r),
			"action":          "proposal_cancelled",
			"result_status":   "cancelled",
			"intent_proof_id": req.IntentProofID,
			"approval_status": "cancelled",
		},
	)

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{
		"cancelled":       true,
		"intent_proof_id": req.IntentProofID,
		"audit_event_id":  auditID,
	}))
}

func rowsAffected(result sql.Result) int64 {
	if result == nil {
		return 0
	}
	count, _ := result.RowsAffected()
	return count
}
