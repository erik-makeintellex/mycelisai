package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/dispatchoutbox"
	"github.com/mycelis/core/internal/outputvalidation"
	"github.com/mycelis/core/pkg/protocol"
)

const (
	teamWorkValidationDispatchKind = "team_output_runtime_validation"
	teamWorkValidationMaxAttempts  = 3
)

type teamWorkValidationDispatchPayload struct {
	Plan          protocol.OutputValidationPlan `json:"plan"`
	ContentDigest string                        `json:"content_digest"`
	LaunchURL     string                        `json:"launch_url"`
	EvidenceRef   string                        `json:"evidence_ref"`
}

func requiredTeamWorkValidationPlan(item protocol.TeamWorkItem) *protocol.OutputValidationPlan {
	if item.WorkIntent == nil || item.WorkIntent.OutputContract == nil {
		return nil
	}
	plan := protocol.NormalizeOutputValidationPlan(item.WorkIntent.OutputContract.OutputValidation)
	if plan == nil || !plan.Required {
		return nil
	}
	return plan
}

func prepareTeamWorkValidation(item protocol.TeamWorkItem, refs []protocol.TeamOutputRef) (*teamWorkValidationDispatchPayload, error) {
	plan := requiredTeamWorkValidationPlan(item)
	if plan == nil {
		return nil, nil
	}
	if err := plan.Validate(); err != nil {
		return nil, err
	}
	digest, err := teamWorkOutputDigest(refs)
	if err != nil {
		return nil, err
	}
	launchURL, err := teamWorkValidationLaunchURL(refs)
	if err != nil {
		return nil, err
	}
	shortDigest := strings.TrimPrefix(digest, "sha256:")
	if len(shortDigest) > 12 {
		shortDigest = shortDigest[:12]
	}
	evidenceRef := path.Join("groups", item.TeamID, "proof", "runtime-validation", item.WorkItemID, shortDigest)
	return &teamWorkValidationDispatchPayload{
		Plan: *plan, ContentDigest: digest, LaunchURL: launchURL, EvidenceRef: evidenceRef,
	}, nil
}

func (s *AdminServer) stageTeamWorkValidationTx(ctx context.Context, tx *sql.Tx, item protocol.TeamWorkItem, payload teamWorkValidationDispatchPayload) (string, error) {
	if s.DispatchOutbox == nil {
		return "", dispatchoutbox.ErrUnavailable
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	key := "team-output-validation:" + item.WorkItemID + ":" + payload.ContentDigest
	_, err = s.DispatchOutbox.EnqueueTx(ctx, tx, dispatchoutbox.Item{
		ID: uuid.NewString(), IdempotencyKey: key, DispatchKind: teamWorkValidationDispatchKind,
		RunID: item.RunID, IntentProofID: item.IntentProofID, ContractID: item.ContractID,
		TeamID: item.TeamID, WorkItemID: item.WorkItemID,
		SourceKind: string(protocol.SourceKindSystem), SourceChannel: "team-work.result-validation",
		PayloadKind: string(protocol.PayloadKindCommand), Payload: raw,
		Recovery: json.RawMessage(`{"action":"retry_output_validation","operator_required":false}`),
	})
	return key, err
}

func (s *AdminServer) dispatchClaimedTeamWorkValidation(ctx context.Context, outboxItem *dispatchoutbox.Item) error {
	var payload teamWorkValidationDispatchPayload
	if err := json.Unmarshal(outboxItem.Payload, &payload); err != nil {
		_ = s.DispatchOutbox.MarkFailed(ctx, outboxItem.ID, err)
		return err
	}
	if s.OutputValidator == nil {
		return s.retryOrFinalizeTeamWorkValidation(ctx, outboxItem, payload, errors.New("browser validation runtime is unavailable"))
	}
	evidencePath, _, err := resolveWorkspacePath(payload.EvidenceRef, false)
	if err != nil {
		return s.retryOrFinalizeTeamWorkValidation(ctx, outboxItem, payload, err)
	}
	if err := os.MkdirAll(evidencePath, 0o755); err != nil {
		return s.retryOrFinalizeTeamWorkValidation(ctx, outboxItem, payload, err)
	}
	report, err := s.OutputValidator.Validate(ctx, outputvalidation.Request{
		LaunchURL: payload.LaunchURL, ContentDigest: payload.ContentDigest,
		EvidencePath: evidencePath, Plan: payload.Plan,
	})
	if err != nil || report.Status == outputvalidation.StatusUnavailable || report.Status == outputvalidation.StatusError {
		if err == nil {
			err = errors.New(validationDiagnosticSummary(report))
		}
		return s.retryOrFinalizeTeamWorkValidation(ctx, outboxItem, payload, err)
	}
	if report.ContentDigest != payload.ContentDigest {
		return s.finalizeTeamWorkValidation(ctx, outboxItem, payload, report, false, "runtime_validation_stale")
	}
	passed := report.Status == outputvalidation.StatusPassed
	degradation := "runtime_validation_failed"
	if passed {
		degradation = ""
	}
	if err := s.finalizeTeamWorkValidation(ctx, outboxItem, payload, report, passed, degradation); err != nil {
		return err
	}
	return s.DispatchOutbox.MarkCompleted(ctx, outboxItem.ID)
}

func (s *AdminServer) retryOrFinalizeTeamWorkValidation(ctx context.Context, item *dispatchoutbox.Item, payload teamWorkValidationDispatchPayload, cause error) error {
	if item.AttemptCount < teamWorkValidationMaxAttempts {
		return s.DispatchOutbox.MarkRetry(ctx, item.ID, cause, time.Duration(item.AttemptCount)*time.Second)
	}
	report := outputvalidation.Report{
		Status: outputvalidation.StatusUnavailable, ContentDigest: payload.ContentDigest,
		LaunchURL: payload.LaunchURL, StartedAt: time.Now().UTC(), FinishedAt: time.Now().UTC(),
		Diagnostics: []outputvalidation.Diagnostic{{Code: "validator_unavailable", Message: cause.Error(), Severity: "error"}},
	}
	if err := s.finalizeTeamWorkValidation(ctx, item, payload, report, false, "runtime_validation_unavailable"); err != nil {
		return err
	}
	return s.DispatchOutbox.MarkFailed(ctx, item.ID, cause)
}

func validationDiagnosticSummary(report outputvalidation.Report) string {
	if len(report.Diagnostics) == 0 {
		if report.Status == outputvalidation.StatusPassed {
			return "The approved runtime workflow passed."
		}
		return "Runtime validation did not complete."
	}
	return strings.TrimSpace(report.Diagnostics[0].Message)
}

func validationEvidenceRefs(report outputvalidation.Report, validationRef string) []string {
	refs := []string{validationRef}
	root, err := workspaceRoot()
	if err != nil {
		return refs
	}
	for _, evidence := range report.EvidenceRefs {
		rel, relErr := filepath.Rel(root, evidence.Path)
		if relErr == nil && !pathEscapesWorkspace(rel) {
			refs = append(refs, filepath.ToSlash(rel))
		}
	}
	return normalizeStringSlice(refs)
}
