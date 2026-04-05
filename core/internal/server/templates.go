package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

// ── CE-1: Template Engine — Confirm Tokens, Intent Proofs, Audit ────

const confirmTokenTTL = 15 * time.Minute

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

// createIntentProof builds and persists a full intent proof bundle (status=pending).
func (s *AdminServer) createIntentProof(templateID protocol.TemplateID, intent string, scope *protocol.ScopeValidation, auditEventID string) (*protocol.IntentProof, error) {
	db := s.getDB()
	if db == nil {
		return nil, nil // graceful: proof is non-blocking if DB unavailable
	}

	id := uuid.New()
	expiresAt := time.Now().Add(confirmTokenTTL)

	var scopeJSON []byte
	if scope != nil {
		scopeJSON, _ = json.Marshal(scope)
	}

	var auditUUID *uuid.UUID
	if auditEventID != "" {
		parsed, err := uuid.Parse(auditEventID)
		if err == nil {
			auditUUID = &parsed
		}
	}

	policyDecision := "allow"
	if scope != nil && scope.Approval != nil && scope.Approval.ApprovalRequired {
		policyDecision = "require_approval"
	}

	_, err := db.Exec(
		`INSERT INTO intent_proofs (id, template_id, resolved_intent, permission_check, policy_decision, scope_validation, audit_event_id, status, expires_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		id, string(templateID), intent, "pass", policyDecision, scopeJSON, auditUUID, "pending", expiresAt,
	)
	if err != nil {
		log.Printf("CE-1: intent proof insert failed: %v", err)
		return nil, err
	}

	now := time.Now()
	return &protocol.IntentProof{
		ID:              id.String(),
		TemplateID:      templateID,
		ResolvedIntent:  intent,
		PermissionCheck: "pass",
		PolicyDecision:  policyDecision,
		ScopeValidation: scope,
		AuditEventID:    auditEventID,
		Status:          "pending",
		CreatedAt:       now,
	}, nil
}

// generateConfirmToken creates a single-use token bound to an intent proof.
// Token expires after confirmTokenTTL (15 minutes).
func (s *AdminServer) generateConfirmToken(proofID string, templateID protocol.TemplateID) (*protocol.ConfirmToken, error) {
	db := s.getDB()
	if db == nil {
		return nil, nil
	}

	token := uuid.New()
	now := time.Now()
	expiresAt := now.Add(confirmTokenTTL)

	_, err := db.Exec(
		`INSERT INTO confirm_tokens (token, intent_proof_id, template_id, expires_at)
		 VALUES ($1, $2, $3, $4)`,
		token, proofID, string(templateID), expiresAt,
	)
	if err != nil {
		log.Printf("CE-1: confirm token insert failed: %v", err)
		return nil, err
	}

	return &protocol.ConfirmToken{
		Token:         token.String(),
		IntentProofID: proofID,
		TemplateID:    templateID,
		CreatedAt:     now,
		ExpiresAt:     expiresAt,
	}, nil
}

// validateConfirmToken checks that a token exists, is not consumed, and not expired.
// Returns the intent proof ID on success. Marks the token as consumed atomically.
func (s *AdminServer) validateConfirmToken(token string) (string, error) {
	db := s.getDB()
	if db == nil {
		return "", errDBUnavailable
	}

	tokenUUID, err := uuid.Parse(token)
	if err != nil {
		return "", errInvalidToken
	}

	var proofID string
	var consumed bool
	var expiresAt time.Time

	err = db.QueryRow(
		`SELECT intent_proof_id, consumed, expires_at FROM confirm_tokens WHERE token = $1`,
		tokenUUID,
	).Scan(&proofID, &consumed, &expiresAt)
	if err != nil {
		return "", errTokenNotFound
	}

	if consumed {
		return "", errTokenConsumed
	}

	if time.Now().After(expiresAt) {
		return "", errTokenExpired
	}

	// Mark consumed atomically
	_, err = db.Exec(
		`UPDATE confirm_tokens SET consumed = TRUE, consumed_at = $1 WHERE token = $2 AND consumed = FALSE`,
		time.Now(), tokenUUID,
	)
	if err != nil {
		log.Printf("CE-1: confirm token consume failed: %v", err)
		return "", err
	}

	return proofID, nil
}

// confirmIntentProof updates a proof's status to confirmed after successful commit.
func (s *AdminServer) confirmIntentProof(proofID, missionID string) {
	db := s.getDB()
	if db == nil {
		return
	}

	proofUUID, err := uuid.Parse(proofID)
	if err != nil {
		return
	}
	missionUUID, err := uuid.Parse(missionID)
	if err != nil {
		return
	}

	_, err = db.Exec(
		`UPDATE intent_proofs SET status = 'confirmed', mission_id = $1, confirmed_at = $2 WHERE id = $3`,
		missionUUID, time.Now(), proofUUID,
	)
	if err != nil {
		log.Printf("CE-1: confirm intent proof update failed: %v", err)
	}
}

// ── Token validation errors ─────────────────────────────────────────

type tokenError string

func (e tokenError) Error() string { return string(e) }

const (
	errDBUnavailable tokenError = "database not available"
	errInvalidToken  tokenError = "invalid token format"
	errTokenNotFound tokenError = "token not found"
	errTokenConsumed tokenError = "token already consumed"
	errTokenExpired  tokenError = "token expired"
)

// ── API Handlers ────────────────────────────────────────────────────

// GET /api/v1/templates — list registered templates.
func (s *AdminServer) handleListTemplatesAPI(w http.ResponseWriter, r *http.Request) {
	view := strings.TrimSpace(r.URL.Query().Get("view"))
	if view == "starter" || view == "organization-starters" {
		templates, err := s.loadOrganizationStarterTemplates()
		if err != nil {
			respondAPIError(w, "failed to load organization starter templates", http.StatusInternalServerError)
			return
		}
		respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(templates))
		return
	}

	templates := make([]protocol.TemplateSpec, 0, len(protocol.TemplateRegistry))
	for _, t := range protocol.TemplateRegistry {
		templates = append(templates, t)
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(templates))
}

// GET /api/v1/intent/proof/{id} — retrieve a specific intent proof.
func (s *AdminServer) handleGetIntentProof(w http.ResponseWriter, r *http.Request) {
	proofID := r.PathValue("id")
	db := s.getDB()
	if db == nil {
		respondAPIError(w, "database not available", http.StatusServiceUnavailable)
		return
	}

	proofUUID, err := uuid.Parse(proofID)
	if err != nil {
		respondAPIError(w, "invalid proof ID", http.StatusBadRequest)
		return
	}

	var proof protocol.IntentProof
	var scopeJSON []byte
	var auditEventID *string
	var missionID *string
	var confirmedAt *time.Time

	err = db.QueryRow(
		`SELECT id, template_id, resolved_intent, user_confirmation_token, permission_check,
		        policy_decision, scope_validation, audit_event_id, mission_id, status, created_at, confirmed_at
		 FROM intent_proofs WHERE id = $1`,
		proofUUID,
	).Scan(
		&proof.ID, &proof.TemplateID, &proof.ResolvedIntent, &proof.UserConfirmToken,
		&proof.PermissionCheck, &proof.PolicyDecision, &scopeJSON, &auditEventID,
		&missionID, &proof.Status, &proof.CreatedAt, &confirmedAt,
	)
	if err != nil {
		respondAPIError(w, "intent proof not found", http.StatusNotFound)
		return
	}

	if len(scopeJSON) > 0 {
		var scope protocol.ScopeValidation
		if json.Unmarshal(scopeJSON, &scope) == nil {
			proof.ScopeValidation = &scope
		}
	}
	if auditEventID != nil {
		proof.AuditEventID = *auditEventID
	}
	if missionID != nil {
		proof.MissionID = *missionID
	}
	if confirmedAt != nil {
		proof.ConfirmedAt = confirmedAt
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(proof))
}

// POST /api/v1/intent/confirm-action — confirm a chat-based proposal action.
// Validates the confirm token, executes the approved mutation plan, persists a
// durable run record, marks the proof confirmed, and creates audit trail. This
// is the lightweight counterpart to intent/commit (which requires a full blueprint).
func (s *AdminServer) HandleConfirmAction(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ConfirmToken string `json:"confirm_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "Bad JSON", http.StatusBadRequest)
		return
	}
	if req.ConfirmToken == "" {
		respondAPIError(w, "Missing confirm_token", http.StatusBadRequest)
		return
	}

	db := s.getDB()
	if db == nil {
		respondAPIError(w, "database not available", http.StatusServiceUnavailable)
		return
	}

	tx, err := db.BeginTx(r.Context(), nil)
	if err != nil {
		log.Printf("CE-1: confirm-action tx begin failed: %v", err)
		respondAPIError(w, "database transaction failed", http.StatusInternalServerError)
		return
	}
	defer tx.Rollback()

	proofID, err := s.consumeConfirmTokenTx(tx, req.ConfirmToken)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}

	scope, err := s.loadIntentProofScopeTx(tx, proofID)
	if err != nil {
		log.Printf("CE-1: confirm-action scope load failed: %v", err)
		respondAPIError(w, "failed to load approved execution plan", http.StatusInternalServerError)
		return
	}

	runID, err := s.createExecutionRunTx(tx, proofID)
	if err != nil {
		log.Printf("CE-1: confirm-action run creation failed: %v", err)
		respondAPIError(w, "failed to create execution record", http.StatusInternalServerError)
		return
	}

	auditUser := auditUserLabelFromRequest(r)
	if err := s.executePlannedToolCalls(r.Context(), scope, auditUser); err != nil {
		if failErr := s.failChatProofTx(tx, proofID); failErr != nil {
			log.Printf("CE-1: confirm-action proof failure update failed: %v", failErr)
		}
		if runErr := s.markRunFailedTx(tx, runID, proofID, err.Error()); runErr != nil {
			log.Printf("CE-1: confirm-action failed run update failed: %v", runErr)
		}
		if commitErr := tx.Commit(); commitErr != nil {
			log.Printf("CE-1: confirm-action failure tx commit failed: %v", commitErr)
		}
		_, _ = s.createAuditEvent(
			protocol.TemplateChatToProposal, "confirm-action",
			"Chat proposal confirmation failed",
			map[string]any{
				"actor":           "Soma",
				"user":            auditUser,
				"action":          "proposal_confirmed",
				"result_status":   "failed",
				"run_id":          runID,
				"intent_proof_id": proofID,
				"approval_status": "failed",
				"approval_reason": err.Error(),
			},
		)
		respondAPIError(w, fmt.Sprintf("approved execution failed: %v", err), http.StatusInternalServerError)
		return
	}

	if err := s.markRunCompletedTx(tx, runID, proofID); err != nil {
		log.Printf("CE-1: confirm-action run completion failed: %v", err)
		respondAPIError(w, "failed to finalize execution record", http.StatusInternalServerError)
		return
	}

	if err := s.confirmChatProofTx(tx, proofID); err != nil {
		log.Printf("CE-1: confirm-action proof update failed: %v", err)
		respondAPIError(w, "failed to confirm intent proof", http.StatusInternalServerError)
		return
	}

	if err := tx.Commit(); err != nil {
		log.Printf("CE-1: confirm-action tx commit failed: %v", err)
		respondAPIError(w, "transaction commit failed", http.StatusInternalServerError)
		return
	}

	// Best-effort audit trail after the durable execution record is committed.
	auditID, _ := s.createAuditEvent(
		protocol.TemplateChatToProposal, "confirm-action",
		"Chat proposal confirmed and verified execution record created",
		map[string]any{
			"proof_id":        proofID,
			"run_id":          runID,
			"execution_state": "verified",
			"actor":           "Soma",
			"user":            auditUser,
			"action":          "proposal_confirmed",
			"result_status":   "confirmed",
			"approval_status": "confirmed",
			"intent_proof_id": proofID,
			"capability_used": strings.Join(scope.CapabilityIDs, ","),
		},
	)
	_, _ = s.createAuditEvent(
		protocol.TemplateChatToProposal, "confirm-action",
		"Execution run completed for confirmed chat proposal",
		map[string]any{
			"actor":           "Soma",
			"user":            auditUser,
			"action":          "execution_run",
			"result_status":   "completed",
			"run_id":          runID,
			"intent_proof_id": proofID,
			"capability_used": strings.Join(scope.CapabilityIDs, ","),
		},
	)

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(map[string]any{
		"confirmed":       true,
		"verified":        true,
		"execution_state": "verified",
		"proof_id":        proofID,
		"audit_event_id":  auditID,
		"run_id":          runID,
		"run_status":      runs.StatusCompleted,
	}))
}

// consumeConfirmTokenTx validates and consumes a confirm token inside a transaction.
// It keeps the token unconsumed if the transaction rolls back.
func (s *AdminServer) consumeConfirmTokenTx(tx *sql.Tx, token string) (string, error) {
	if tx == nil {
		return "", errDBUnavailable
	}
	tokenUUID, err := uuid.Parse(token)
	if err != nil {
		return "", errInvalidToken
	}

	var proofID string
	var consumed bool
	var expiresAt time.Time

	err = tx.QueryRow(
		`SELECT intent_proof_id, consumed, expires_at FROM confirm_tokens WHERE token = $1`,
		tokenUUID,
	).Scan(&proofID, &consumed, &expiresAt)
	if err == sql.ErrNoRows {
		return "", errTokenNotFound
	} else if err != nil {
		return "", err
	}
	if consumed {
		return "", errTokenConsumed
	}
	if time.Now().After(expiresAt) {
		return "", errTokenExpired
	}

	_, err = tx.Exec(
		`UPDATE confirm_tokens SET consumed = TRUE, consumed_at = $1 WHERE token = $2 AND consumed = FALSE`,
		time.Now(), tokenUUID,
	)
	if err != nil {
		return "", err
	}

	return proofID, nil
}

func (s *AdminServer) loadIntentProofScopeTx(tx *sql.Tx, proofID string) (*protocol.ScopeValidation, error) {
	if tx == nil {
		return nil, errDBUnavailable
	}
	if proofID == "" {
		return nil, fmt.Errorf("proof_id is required")
	}

	proofUUID, err := uuid.Parse(proofID)
	if err != nil {
		return nil, err
	}

	var scopeJSON []byte
	err = tx.QueryRow(`SELECT scope_validation FROM intent_proofs WHERE id = $1`, proofUUID).Scan(&scopeJSON)
	if err != nil {
		return nil, err
	}

	scope := &protocol.ScopeValidation{}
	if len(scopeJSON) == 0 {
		return scope, nil
	}
	if err := json.Unmarshal(scopeJSON, scope); err != nil {
		return nil, err
	}
	return scope, nil
}

// createExecutionRunTx persists a durable execution record before the approved
// action is executed so later status updates have a stable identity to target.
func (s *AdminServer) createExecutionRunTx(tx *sql.Tx, proofID string) (string, error) {
	if tx == nil {
		return "", errDBUnavailable
	}
	if proofID == "" {
		return "", fmt.Errorf("proof_id is required")
	}

	runID := uuid.New().String()
	now := time.Now()
	_, err := tx.Exec(
		`INSERT INTO mission_runs (id, mission_id, tenant_id, status, run_depth, started_at)
		 VALUES ($1, $2, $3, $4, $5, $6)`,
		runID, proofID, "default", runs.StatusRunning, 0, now,
	)
	if err != nil {
		return "", err
	}

	return runID, nil
}

func (s *AdminServer) markRunCompletedTx(tx *sql.Tx, runID, proofID string) error {
	if tx == nil {
		return errDBUnavailable
	}
	if strings.TrimSpace(runID) == "" {
		return fmt.Errorf("run_id is required")
	}
	now := time.Now()

	_, err := tx.Exec(
		`UPDATE mission_runs SET status = $1, completed_at = NOW() WHERE id = $2`,
		runs.StatusCompleted, runID,
	)
	if err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]any{
		"proof_id":        proofID,
		"execution_state": "verified",
		"run_status":      runs.StatusCompleted,
	})
	_, err = tx.Exec(
		`INSERT INTO mission_events
			(id, run_id, tenant_id, event_type, severity, source_agent, source_team, payload, emitted_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		uuid.New().String(), runID, "default", string(protocol.EventMissionCompleted), string(protocol.SeverityInfo),
		"admin", "governance", payload, now,
	)
	if err != nil {
		return err
	}

	return nil
}

func (s *AdminServer) markRunFailedTx(tx *sql.Tx, runID, proofID, reason string) error {
	if tx == nil {
		return errDBUnavailable
	}
	if strings.TrimSpace(runID) == "" {
		return fmt.Errorf("run_id is required")
	}

	_, err := tx.Exec(
		`UPDATE mission_runs SET status = $1, completed_at = NOW() WHERE id = $2`,
		runs.StatusFailed, runID,
	)
	if err != nil {
		return err
	}

	payload, _ := json.Marshal(map[string]any{
		"proof_id":        proofID,
		"execution_state": "failed",
		"run_status":      runs.StatusFailed,
		"reason":          strings.TrimSpace(reason),
	})
	_, err = tx.Exec(
		`INSERT INTO mission_events
			(id, run_id, tenant_id, event_type, severity, source_agent, source_team, payload, emitted_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`,
		uuid.New().String(), runID, "default", string(protocol.EventMissionFailed), string(protocol.SeverityError),
		"admin", "governance", payload, time.Now(),
	)
	return err
}

// confirmChatProofTx updates a proof's status to confirmed for chat-based proposals.
// Unlike confirmIntentProof, this doesn't require a mission ID.
func (s *AdminServer) confirmChatProofTx(tx *sql.Tx, proofID string) error {
	if tx == nil {
		return errDBUnavailable
	}
	proofUUID, err := uuid.Parse(proofID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(
		`UPDATE intent_proofs SET status = 'confirmed', confirmed_at = $1 WHERE id = $2`,
		time.Now(), proofUUID,
	)
	if err != nil {
		return err
	}
	return nil
}

func (s *AdminServer) failChatProofTx(tx *sql.Tx, proofID string) error {
	if tx == nil {
		return errDBUnavailable
	}
	proofUUID, err := uuid.Parse(proofID)
	if err != nil {
		return err
	}

	_, err = tx.Exec(`UPDATE intent_proofs SET status = 'failed' WHERE id = $1`, proofUUID)
	return err
}

func (s *AdminServer) executePlannedToolCalls(ctx context.Context, scope *protocol.ScopeValidation, auditUser string) error {
	if scope == nil || len(scope.PlannedToolCalls) == 0 {
		return fmt.Errorf("no approved execution plan was stored for this proposal")
	}

	// Confirmation executes only the stored approved plan. The model does not
	// get a fresh chance to improvise a different action during confirm-action.
	registry := swarm.NewInternalToolRegistry(swarm.InternalToolDeps{
		NC:    s.NC,
		Brain: s.Cognitive,
		DB:    s.getDB(),
	})
	executor := swarm.NewCompositeToolExecutor(registry, nil)
	toolCtx := swarm.WithToolInvocationContext(ctx, swarm.ToolInvocationContext{
		SourceKind:    protocol.SourceKindWebAPI,
		SourceChannel: "api.intent.confirm-action",
		PayloadKind:   protocol.PayloadKindCommand,
		Timestamp:     time.Now(),
		UserLabel:     auditUser,
		PlanningOnly:  false,
	})

	for _, planned := range scope.PlannedToolCalls {
		toolName := strings.TrimSpace(planned.Name)
		if toolName == "" {
			return fmt.Errorf("approved execution plan contained an empty tool name")
		}
		serverID, resolvedToolName, err := executor.FindToolByName(toolCtx, toolName)
		if err != nil {
			return err
		}
		if _, err := executor.CallTool(toolCtx, serverID, resolvedToolName, planned.Arguments); err != nil {
			return err
		}
		resource := firstNonEmptyString(planned.Arguments["path"], planned.Arguments["subject"], planned.Arguments["channel"])
		capabilityID := capabilityForPlannedTool(resolvedToolName)
		details := buildExecutionAuditDetailsForTool(planned, resolvedToolName)
		_, _ = s.createAuditEvent(
			protocol.TemplateChatToProposal, "confirm-action",
			"Approved capability usage executed",
			map[string]any{
				"actor":           "Soma",
				"user":            auditUser,
				"action":          "capability_usage",
				"result_status":   "completed",
				"capability_used": capabilityID,
				"resource":        resource,
				"details":         details,
			},
		)
		if resolvedToolName == "publish_signal" || resolvedToolName == "broadcast" {
			_, _ = s.createAuditEvent(
				protocol.TemplateChatToProposal, "confirm-action",
				"Governed channel write executed",
				map[string]any{
					"actor":           "Soma",
					"user":            auditUser,
					"action":          "channel_written",
					"result_status":   "completed",
					"capability_used": capabilityID,
					"resource":        resource,
				},
			)
		}
		if resolvedToolName == "write_file" || resolvedToolName == "promote_deployment_context" {
			_, _ = s.createAuditEvent(
				protocol.TemplateChatToProposal, "confirm-action",
				"Governed artifact created",
				map[string]any{
					"actor":           "Soma",
					"user":            auditUser,
					"action":          "artifact_created",
					"result_status":   "completed",
					"capability_used": capabilityID,
					"resource":        resource,
				},
			)
		}
	}

	return nil
}

// Delegation is the first governed tool where operators benefit from seeing the
// underlying structured ask without exposing raw command payloads by default.
func buildExecutionAuditDetailsForTool(planned protocol.PlannedToolCall, resolvedToolName string) map[string]any {
	details := map[string]any{"tool": resolvedToolName}
	if resolvedToolName != "delegate_task" {
		return details
	}

	teamID, ask := extractDelegationAuditFields(planned.Arguments)
	if teamID != "" {
		details["team_id"] = teamID
	}
	if ask.IsZero() {
		return details
	}

	details["ask_kind"] = string(ask.AskKind)
	details["lane_role"] = string(ask.LaneRole)
	if goal := strings.TrimSpace(ask.Goal); goal != "" {
		details["goal"] = goal
	}
	if summary := protocol.SummarizeTeamAsk(ask); summary != "" {
		details["operator_summary"] = summary
	}
	return details
}

func extractDelegationAuditFields(args map[string]any) (string, protocol.TeamAsk) {
	if len(args) == 0 {
		return "", protocol.TeamAsk{}
	}

	teamID := firstNonEmptyString(args["team_id"], args["teamId"], args["target_team"])
	if teamID == "" {
		switch team := args["team"].(type) {
		case map[string]any:
			teamID = firstNonEmptyString(team["id"], team["team_id"], team["name"])
		case string:
			teamID = strings.TrimSpace(team)
		}
	}

	if rawAsk, ok := args["ask"].(map[string]any); ok {
		return teamID, protocol.TeamAskFromMap(rawAsk)
	}
	switch task := args["task"].(type) {
	case map[string]any:
		return teamID, protocol.TeamAskFromMap(task)
	case string:
		return teamID, protocol.NormalizeTeamAsk(protocol.TeamAsk{Goal: strings.TrimSpace(task)})
	}

	return teamID, protocol.NormalizeTeamAsk(protocol.TeamAsk{
		AskKind:  protocol.TeamAskKind(firstNonEmptyString(args["ask_kind"])),
		LaneRole: protocol.TeamLaneRole(firstNonEmptyString(args["lane_role"])),
		Goal: firstNonEmptyString(
			args["goal"],
			args["intent"],
			args["message"],
			args["operation"],
		),
	})
}

// ── Scope Validation Helpers ────────────────────────────────────────

// buildScopeFromBlueprint extracts scope validation metadata from a blueprint.
func buildScopeFromBlueprint(bp *protocol.MissionBlueprint) *protocol.ScopeValidation {
	toolSet := make(map[string]bool)
	for _, team := range bp.Teams {
		for _, agent := range team.Agents {
			for _, tool := range agent.Tools {
				toolSet[tool] = true
			}
		}
	}

	tools := make([]string, 0, len(toolSet))
	for t := range toolSet {
		tools = append(tools, t)
	}

	totalAgents := 0
	for _, team := range bp.Teams {
		totalAgents += len(team.Agents)
	}

	risk := "low"
	if totalAgents > 5 || len(bp.Teams) > 2 {
		risk = "medium"
	}
	if totalAgents > 10 || len(bp.Teams) > 4 {
		risk = "high"
	}

	return &protocol.ScopeValidation{
		Tools:             tools,
		AffectedResources: []string{"missions", "teams", "service_manifests"},
		RiskLevel:         risk,
	}
}
