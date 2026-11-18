package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/runs"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

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

func (s *AdminServer) executePlannedToolCalls(ctx context.Context, tx *sql.Tx, scope *protocol.ScopeValidation, auditUser, runID, proofID string) (*packageOutputSummary, error) {
	if scope == nil || len(scope.PlannedToolCalls) == 0 {
		return nil, fmt.Errorf("no approved execution plan was stored for this proposal")
	}

	registry := swarm.NewInternalToolRegistry(swarm.InternalToolDeps{
		NC:    s.NC,
		Brain: s.Cognitive,
		DB:    s.getDB(),
	})
	registry.SetSoma(s.Soma)
	executor := swarm.NewCompositeToolExecutor(registry, nil)
	toolCtx := swarm.WithToolInvocationContext(ctx, swarm.ToolInvocationContext{
		RunID:         runID,
		SourceKind:    protocol.SourceKindWebAPI,
		SourceChannel: "api.intent.confirm-action",
		PayloadKind:   protocol.PayloadKindCommand,
		Timestamp:     time.Now(),
		UserLabel:     auditUser,
		PlanningOnly:  false,
	})

	prevRunID := runID
	teamDepth := 0
	for _, planned := range scope.PlannedToolCalls {
		toolName := strings.TrimSpace(planned.Name)
		if toolName == "" {
			return nil, fmt.Errorf("approved execution plan contained an empty tool name")
		}
		serverID, resolvedToolName, err := executor.FindToolByName(toolCtx, toolName)
		if err != nil {
			return nil, err
		}
		if _, err := executor.CallTool(toolCtx, serverID, resolvedToolName, planned.Arguments); err != nil {
			return nil, err
		}
		if resolvedToolName == "create_team" {
			teamDepth++
			prevRunID = recordPlannedTeamChildRunTx(tx, proofID, prevRunID, teamDepth, planned)
		}
		s.auditExecutedPlannedTool(planned, resolvedToolName, auditUser)
	}

	packageSummary, err := s.finalizeCodexTaskBundleTx(ctx, tx, scope.PlannedToolCalls, runID, proofID)
	if err != nil {
		return nil, err
	}

	return packageSummary, nil
}

func stringSliceFromAny(raw any) []string {
	switch v := raw.(type) {
	case []string:
		return v
	case []any:
		out := make([]string, 0, len(v))
		for _, item := range v {
			if s := firstNonEmptyString(item); s != "" {
				out = append(out, s)
			}
		}
		return out
	}
	return nil
}

func recordPlannedTeamChildRunTx(tx *sql.Tx, proofID, parentRunID string, depth int, planned protocol.PlannedToolCall) string {
	if tx == nil || strings.TrimSpace(planned.Name) != "create_team" {
		return parentRunID
	}
	teamID := firstNonEmptyString(planned.Arguments["team_id"], planned.Arguments["id"])
	teamName := firstNonEmptyString(planned.Arguments["name"], teamID)
	listensOn := strings.Join(stringSliceFromAny(planned.Arguments["inputs"]), ", ")
	speaksOn := strings.Join(stringSliceFromAny(planned.Arguments["deliveries"]), ", ")
	childRunID := uuid.New().String()
	now := time.Now()

	_, err := tx.Exec(
		"INSERT INTO mission_runs (id, mission_id, tenant_id, status, run_depth, parent_run_id, started_at, completed_at) VALUES ($1, $2, $3, $4, $5, NULLIF($6,'')::uuid, $7, $8)",
		childRunID, proofID, "default", runs.StatusCompleted, depth, parentRunID, now, now,
	)
	if err != nil {
		return parentRunID
	}

	payload, _ := json.Marshal(map[string]any{
		"team_id": teamID, "team_name": teamName,
		"listens_on": listensOn, "speaks_on": speaksOn,
		"action": "team_manifested_and_executed",
	})
	const insertEvSQL = `INSERT INTO mission_events (id, run_id, tenant_id, event_type, severity, source_agent, source_team, payload, emitted_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)`
	_, _ = tx.Exec(insertEvSQL, uuid.New().String(), childRunID, "default", string(protocol.EventToolCompleted), string(protocol.SeverityInfo), "soma", teamID, payload, now)
	_, _ = tx.Exec(insertEvSQL, uuid.New().String(), parentRunID, "default", "signal.published", string(protocol.SeverityInfo), "soma", teamID, payload, now)
	return childRunID
}

func (s *AdminServer) auditExecutedPlannedTool(planned protocol.PlannedToolCall, resolvedToolName, auditUser string) {
	resource := firstNonEmptyString(planned.Arguments["path"], planned.Arguments["team_id"], planned.Arguments["subject"], planned.Arguments["channel"])
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
