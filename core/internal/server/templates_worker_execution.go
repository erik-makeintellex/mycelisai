package server

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mycelis/core/internal/workers"
	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) confirmedActionWorkerBackend() workers.WorkerBackend {
	if s.WorkerBackend != nil {
		return s.WorkerBackend
	}
	s.WorkerBackend = workers.NewCentralBackend()
	return s.WorkerBackend
}

func buildConfirmedActionWorkerRunRequest(proofID string, scope *protocol.ScopeValidation, auditUser string) workers.WorkerRunRequest {
	metadata := map[string]any{
		"intent_proof_id":    strings.TrimSpace(proofID),
		"source_kind":        string(protocol.SourceKindWebAPI),
		"source_channel":     "api.intent.confirm-action",
		"payload_kind":       string(protocol.PayloadKindCommand),
		"planned_tool_count": 0,
	}
	intent := "Confirmed Soma work"
	if scope != nil {
		intent = firstNonEmptyString(scopeWorkObjective(scope), scope.ExecutionMode, intent)
		metadata["risk_level"] = strings.TrimSpace(scope.RiskLevel)
		metadata["affected_resources"] = normalizeStringSlice(scope.AffectedResources)
		metadata["capability_ids"] = normalizeStringSlice(scope.CapabilityIDs)
		metadata["execution_mode"] = scopeExecutionMode(scope)
		metadata["planned_tool_count"] = len(scope.PlannedToolCalls)
		if scope.WorkIntent != nil {
			metadata["work_intent"] = scope.WorkIntent
		}
	}
	return workers.WorkerRunRequest{
		UserID:            strings.TrimSpace(auditUser),
		RequestedBy:       firstNonEmptyString(auditUser, "Soma"),
		Intent:            intent,
		Instructions:      "Execute the operator-approved Soma proposal through governed planned tools.",
		RequiredFeatures:  normalizeStringSlice(scopeCapabilityIDs(scope)),
		RequiredProtocols: []workers.Protocol{workers.ProtocolRunsAPI},
		Metadata:          metadata,
	}
}

func scopeWorkObjective(scope *protocol.ScopeValidation) string {
	if scope == nil || scope.WorkIntent == nil {
		return ""
	}
	return strings.TrimSpace(scope.WorkIntent.Objective)
}

func scopeCapabilityIDs(scope *protocol.ScopeValidation) []string {
	if scope == nil {
		return nil
	}
	return scope.CapabilityIDs
}

func (s *AdminServer) completeConfirmedActionWorkerRun(ctx context.Context, runID string, results []plannedToolExecutionResult) {
	finalizer, ok := s.confirmedActionWorkerBackend().(workers.RunFinalizer)
	if !ok {
		return
	}
	if err := finalizer.CompleteRun(ctx, runID, workerResultFromPlannedResults(results)); err != nil {
		log.Printf("CE-1: worker run completion failed: %v", err)
	}
}

func (s *AdminServer) failConfirmedActionWorkerRun(ctx context.Context, runID string, err error) {
	finalizer, ok := s.confirmedActionWorkerBackend().(workers.RunFinalizer)
	if !ok {
		return
	}
	workerErr := &workers.WorkerError{
		Code:        "confirmed_action_failed",
		Message:     strings.TrimSpace(errorString(err)),
		Recoverable: true,
	}
	if workerErr.Message == "" {
		workerErr.Message = "Confirmed action failed."
	}
	if failErr := finalizer.FailRun(ctx, runID, workerErr); failErr != nil {
		log.Printf("CE-1: worker run failure update failed: %v", failErr)
	}
}

func workerResultFromPlannedResults(results []plannedToolExecutionResult) workers.WorkerResult {
	outputs := executionOutputsFromToolResults(results)
	workerOutputs := make([]workers.WorkerOutput, 0, len(outputs))
	for _, output := range outputs {
		workerOutputs = append(workerOutputs, workers.WorkerOutput{
			ID:   firstNonEmptyString(output.ID, output.Title),
			Kind: firstNonEmptyString(output.Kind, "output"),
			Name: output.Title,
			URI:  firstNonEmptyString(output.Folder, output.Entrypoint, output.OpenURL, output.Href),
			Metadata: map[string]any{
				"validation_ref": output.Validation,
				"artifact_id":    output.ArtifactID,
				"proof_artifact": output.ProofArtifactID,
			},
		})
	}
	return workers.WorkerResult{
		Summary:    workerResultSummary(results),
		Outputs:    workerOutputs,
		FinishedAt: time.Now().UTC(),
	}
}

func workerResultSummary(results []plannedToolExecutionResult) string {
	if len(results) == 0 {
		return "Confirmed action completed."
	}
	if len(results) == 1 {
		return firstNonEmptyString(results[0].Output, "Confirmed action completed.")
	}
	return fmt.Sprintf("Confirmed action completed with %d governed tool results.", len(results))
}

func errorString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
