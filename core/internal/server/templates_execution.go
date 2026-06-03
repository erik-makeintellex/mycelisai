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
	"github.com/mycelis/core/internal/trust"
	"github.com/mycelis/core/pkg/protocol"
)

type plannedToolExecutionResult struct {
	Name      string
	ToolRef   string
	Arguments map[string]any
	Output    string
	Artifacts []protocol.ChatArtifactRef
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

func (s *AdminServer) ensureExecutionContractTx(ctx context.Context, tx *sql.Tx, proofID, runID string) (string, error) {
	if tx == nil {
		return "", errDBUnavailable
	}
	if proofID == "" {
		return "", fmt.Errorf("proof_id is required")
	}
	proofUUID, err := uuid.Parse(proofID)
	if err != nil {
		return "", err
	}

	var templateID string
	var resolvedIntent string
	var auditEventID string
	if err := tx.QueryRowContext(ctx,
		`SELECT template_id, resolved_intent, COALESCE(audit_event_id::text, '') FROM intent_proofs WHERE id = $1`,
		proofUUID,
	).Scan(&templateID, &resolvedIntent, &auditEventID); err != nil {
		return "", err
	}

	return trust.UpsertContract(ctx, tx, trust.ContractInput{
		IntentProofID:  proofID,
		RunID:          runID,
		TemplateID:     protocol.TemplateID(templateID),
		ResolvedIntent: resolvedIntent,
		AuditEventID:   auditEventID,
	})
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

func (s *AdminServer) executePlannedToolCalls(ctx context.Context, scope *protocol.ScopeValidation, auditUser string) ([]plannedToolExecutionResult, error) {
	if scope == nil || len(scope.PlannedToolCalls) == 0 {
		return nil, fmt.Errorf("no approved execution plan was stored for this proposal")
	}

	registry := swarm.NewInternalToolRegistry(swarm.InternalToolDeps{
		NC:    s.NC,
		Brain: s.Cognitive,
		DB:    s.getDB(),
	})
	registry.SetSoma(s.Soma)
	mcpExec := s.plannedMCPToolExecutor()
	executor := swarm.NewCompositeToolExecutor(registry, mcpExec)
	toolCtx := swarm.WithToolInvocationContext(ctx, swarm.ToolInvocationContext{
		SourceKind:    protocol.SourceKindWebAPI,
		SourceChannel: "api.intent.confirm-action",
		PayloadKind:   protocol.PayloadKindCommand,
		Timestamp:     time.Now(),
		UserLabel:     auditUser,
		PlanningOnly:  false,
	})

	results := make([]plannedToolExecutionResult, 0, len(scope.PlannedToolCalls))
	lastGeneratedImageArtifactID := ""
	for _, planned := range scope.PlannedToolCalls {
		planned = normalizePlannedToolCall(planned)
		toolName := strings.TrimSpace(planned.Name)
		if toolName == "" {
			return results, fmt.Errorf("approved execution plan contained an empty tool name")
		}
		if strings.EqualFold(toolName, "save_cached_image") && firstNonEmptyString(planned.Arguments["artifact_id"]) == "" && lastGeneratedImageArtifactID != "" {
			if planned.Arguments == nil {
				planned.Arguments = map[string]any{}
			}
			planned.Arguments["artifact_id"] = lastGeneratedImageArtifactID
		}
		serverID, resolvedToolName, err := s.resolveApprovedToolCall(toolCtx, executor, mcpExec, planned)
		if err != nil {
			return results, err
		}
		output, err := executor.CallTool(toolCtx, serverID, resolvedToolName, planned.Arguments)
		if err != nil {
			return results, err
		}
		artifacts := []protocol.ChatArtifactRef(nil)
		if message, parsedArtifacts, ok := extractPlannedToolOutputArtifacts(output); ok {
			if strings.TrimSpace(message) != "" {
				output = message
			}
			artifacts = parsedArtifacts
		}
		if strings.EqualFold(resolvedToolName, "generate_image") {
			lastGeneratedImageArtifactID = firstGeneratedImageArtifactID(artifacts)
		}
		results = append(results, plannedToolExecutionResult{
			Name:      resolvedToolName,
			ToolRef:   strings.TrimSpace(planned.ToolRef),
			Arguments: planned.Arguments,
			Output:    output,
			Artifacts: artifacts,
		})
		s.auditExecutedPlannedTool(planned, resolvedToolName, auditUser)
	}

	return results, nil
}

func (s *AdminServer) auditExecutedPlannedTool(planned protocol.PlannedToolCall, resolvedToolName, auditUser string) {
	resource := firstNonEmptyString(planned.Arguments["path"], planned.Arguments["subject"], planned.Arguments["channel"])
	capabilityID := capabilityForPlannedTool(firstNonEmptyString(planned.ToolRef, resolvedToolName))
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
