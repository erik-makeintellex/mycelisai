package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/mycelis/core/internal/configdocuments"
	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) resolveThreadOutcomeTemplateActivationOrRespond(
	w http.ResponseWriter,
	r *http.Request,
	sessionID string,
	messages []chatRequestMessage,
	organizationID string,
	teamID string,
	actorID string,
	planned *[]protocol.PlannedToolCall,
) bool {
	resolved, err := s.resolveThreadOutcomeTemplateActivation(
		r.Context(), sessionID, messages, organizationID, teamID, actorID, *planned,
	)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusConflict)
		return false
	}
	*planned = resolved
	return true
}

func (s *AdminServer) applyThreadOutcomeTemplateOrRespond(
	w http.ResponseWriter,
	r *http.Request,
	sessionID string,
	messages []chatRequestMessage,
	latestRequest string,
	organizationID string,
	teamID string,
	actorID string,
	display *proposalDisplayContract,
) bool {
	_, err := s.applyThreadOutcomeTemplate(
		r.Context(), sessionID, messages, latestRequest, organizationID, teamID, actorID, display,
	)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusConflict)
		return false
	}
	return true
}

func (s *AdminServer) resolveThreadOutcomeTemplateActivation(
	ctx context.Context,
	sessionID string,
	messages []chatRequestMessage,
	organizationID string,
	teamID string,
	actorID string,
	planned []protocol.PlannedToolCall,
) ([]protocol.PlannedToolCall, error) {
	hasActivation := false
	for _, call := range planned {
		if strings.EqualFold(strings.TrimSpace(call.Name), "activate_config_document") {
			hasActivation = true
			break
		}
	}
	if !hasActivation {
		return planned, nil
	}
	store, err := s.runtimeConfigDocumentStore()
	if err != nil {
		return nil, err
	}
	document, ok := s.latestThreadOutcomeTemplate(ctx, sessionID, messages)
	if !ok {
		return nil, fmt.Errorf("Soma could not identify which Outcome Template revision to activate")
	}
	if scopeErr := validateOutcomeTemplateRequestScope(document, organizationID, teamID, actorID); scopeErr != nil {
		return nil, scopeErr
	}
	digest, err := protocol.CanonicalConfigDocumentDigest(document)
	if err != nil {
		return nil, err
	}

	resolvedRecordID := ""
	for i, call := range planned {
		if !strings.EqualFold(strings.TrimSpace(call.Name), "activate_config_document") {
			continue
		}
		recordID := firstNonEmptyString(call.Arguments["record_id"])
		if recordID != "" {
			record, loadErr := store.GetRevision(ctx, "default", recordID)
			if loadErr != nil {
				return nil, fmt.Errorf("load selected configuration revision: %w", loadErr)
			}
			if scopeErr := validateOutcomeTemplateRequestScope(record.Document, organizationID, teamID, actorID); scopeErr != nil {
				return nil, scopeErr
			}
			recordDigest, digestErr := protocol.CanonicalConfigDocumentDigest(record.Document)
			if digestErr != nil || record.Document.Metadata.ID != document.Metadata.ID ||
				record.Document.Metadata.Version != document.Metadata.Version || record.Digest != digest || recordDigest != digest {
				return nil, fmt.Errorf("the selected Outcome Template revision does not match this thread")
			}
			continue
		}
		if resolvedRecordID == "" {
			records, listErr := store.List(ctx, "default", configdocuments.ListFilter{
				Kind:       document.Kind,
				ScopeKind:  document.Metadata.Scope.Kind,
				ScopeRef:   document.Metadata.Scope.Ref,
				DocumentID: document.Metadata.ID,
				Limit:      20,
			})
			if listErr != nil {
				return nil, listErr
			}
			for _, record := range records {
				if record.Digest == digest {
					resolvedRecordID = record.RecordID
					break
				}
			}
			if resolvedRecordID == "" {
				return nil, fmt.Errorf("the Outcome Template revision from this thread has not been saved")
			}
		}
		if call.Arguments == nil {
			call.Arguments = map[string]any{}
		}
		call.Arguments["record_id"] = resolvedRecordID
		if firstNonEmptyString(call.Arguments["action"]) == "" {
			call.Arguments["action"] = string(configdocuments.ActivationActionActivate)
		}
		planned[i] = call
	}
	return planned, nil
}

func (s *AdminServer) applyThreadOutcomeTemplate(
	ctx context.Context,
	sessionID string,
	messages []chatRequestMessage,
	latestRequest string,
	organizationID string,
	teamID string,
	actorID string,
	display *proposalDisplayContract,
) (bool, error) {
	if display == nil || !outcomeTemplateWorkApplication(strings.ToLower(strings.TrimSpace(latestRequest))) {
		return false, nil
	}
	document, ok := s.latestThreadOutcomeTemplate(ctx, sessionID, messages)
	if !ok {
		return false, fmt.Errorf("Soma could not identify the Outcome Template referenced by this work")
	}
	if err := validateOutcomeTemplateRequestScope(document, organizationID, teamID, actorID); err != nil {
		return false, err
	}
	store, err := s.runtimeConfigDocumentStore()
	if err != nil {
		return false, err
	}
	record, err := store.GetActiveRevision(
		ctx, "default", document.Kind, document.Metadata.ID, document.Metadata.Scope,
	)
	if err != nil {
		return false, fmt.Errorf("load active Outcome Template: %w", err)
	}
	if err := validateOutcomeTemplateRequestScope(record.Document, organizationID, teamID, actorID); err != nil {
		return false, err
	}
	compiled, err := configdocuments.CompileOutcomeTemplateDocument(
		record.Document,
		protocol.MinimumSufficientBrief{TargetOutcome: strings.TrimSpace(latestRequest)},
		protocol.MinimumSufficientBrief{},
	)
	if err != nil {
		return false, err
	}
	if !compiled.Ready || compiled.WorkIntent == nil {
		return false, fmt.Errorf("active Outcome Template needs clarification before governed work can start")
	}
	display.WorkIntent = mergeTemplateWorkIntent(compiled.WorkIntent, display.WorkIntent)
	display.OperatorSummary = fmt.Sprintf(
		"Using %s v%s to shape this work.",
		record.Document.Metadata.Name, record.Document.Metadata.Version,
	)
	return true, nil
}

func validateOutcomeTemplateRequestScope(
	document protocol.ConfigDocument,
	organizationID string,
	teamID string,
	actorID string,
) error {
	scope := document.Metadata.Scope
	actual := strings.TrimSpace(scope.Ref)
	expected := ""
	switch scope.Kind {
	case protocol.ConfigDocumentScopeBuiltIn:
		return nil
	case protocol.ConfigDocumentScopeOrganization:
		expected = strings.TrimSpace(organizationID)
	case protocol.ConfigDocumentScopeWorkspace:
		expected = firstNonEmptyString(organizationID, teamID)
	case protocol.ConfigDocumentScopeOperator:
		expected = strings.TrimSpace(actorID)
	default:
		return fmt.Errorf("the Outcome Template has an unsupported scope")
	}
	if expected == "" || actual != expected {
		return fmt.Errorf("the Outcome Template is not available in the current workspace")
	}
	return nil
}

func (s *AdminServer) latestThreadOutcomeTemplate(
	ctx context.Context,
	sessionID string,
	messages []chatRequestMessage,
) (protocol.ConfigDocument, bool) {
	for i := len(messages) - 1; i >= 0; i-- {
		_, _, document, ok := parseInlineConfigDocument(messages[i].Content)
		if ok && document.Kind == protocol.ConfigDocumentKindOutcomeTemplate {
			return document, true
		}
	}
	if receipt, ok := s.latestSessionConfigDocumentReceipt(ctx, sessionID); ok {
		store, err := s.runtimeConfigDocumentStore()
		if err != nil {
			return protocol.ConfigDocument{}, false
		}
		record, err := store.GetRevision(ctx, "default", receipt.RecordID)
		if err == nil && record.Digest == receipt.Digest {
			return record.Document, true
		}
	}
	return protocol.ConfigDocument{}, false
}

func (s *AdminServer) runtimeConfigDocumentStore() (*configdocuments.Store, error) {
	if s == nil || s.getDB() == nil {
		return nil, fmt.Errorf("Configuration database unavailable")
	}
	return configdocuments.NewStore(s.getDB()), nil
}

func mergeTemplateWorkIntent(compiled, runtime *protocol.WorkIntent) *protocol.WorkIntent {
	if compiled == nil {
		return runtime
	}
	if runtime == nil {
		return compiled
	}
	if runtime.Kind != "" {
		compiled.Kind = runtime.Kind
	}
	if runtime.Objective != "" {
		compiled.Objective = runtime.Objective
	}
	if runtime.Cadence != "" {
		compiled.Cadence = runtime.Cadence
	}
	if runtime.ScheduleSummary != "" {
		compiled.ScheduleSummary = runtime.ScheduleSummary
	}
	if runtime.RuntimePosture != "" {
		compiled.RuntimePosture = runtime.RuntimePosture
	}
	if runtime.TargetTeamID != "" {
		compiled.TargetTeamID = runtime.TargetTeamID
	}
	if runtime.BusScope != "" {
		compiled.BusScope = runtime.BusScope
	}
	if len(runtime.NATSSubjects) > 0 {
		compiled.NATSSubjects = append([]string(nil), runtime.NATSSubjects...)
	}
	if len(runtime.ServiceRefs) > 0 {
		compiled.ServiceRefs = append([]string(nil), runtime.ServiceRefs...)
	}
	if runtime.ProjectRef != "" {
		compiled.ProjectRef = runtime.ProjectRef
	}
	if runtime.Lifecycle != nil {
		compiled.Lifecycle = runtime.Lifecycle
	}
	if runtime.SideEffect != nil {
		compiled.SideEffect = runtime.SideEffect
	}
	if runtime.OutputContract != nil {
		if compiled.OutputContract == nil {
			compiled.OutputContract = &protocol.WorkOutputContract{}
		}
		if runtime.OutputContract.Shape != "" {
			compiled.OutputContract.Shape = runtime.OutputContract.Shape
		}
		if runtime.OutputContract.PrimaryDeliverable != "" {
			compiled.OutputContract.PrimaryDeliverable = runtime.OutputContract.PrimaryDeliverable
		}
		if runtime.OutputContract.Retention != "" {
			compiled.OutputContract.Retention = runtime.OutputContract.Retention
		}
		if runtime.OutputContract.LaunchHint != "" {
			compiled.OutputContract.LaunchHint = runtime.OutputContract.LaunchHint
		}
		if len(runtime.OutputContract.Validation) > 0 {
			compiled.OutputContract.Validation = append([]string(nil), runtime.OutputContract.Validation...)
		}
		if runtime.OutputContract.OutputValidation != nil {
			compiled.OutputContract.OutputValidation = runtime.OutputContract.OutputValidation
		}
	}
	return protocol.NormalizeWorkIntent(compiled)
}
