package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/mycelis/core/internal/configdocuments"
	"github.com/mycelis/core/pkg/protocol"
)

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
		return fmt.Errorf("the configuration has an unsupported scope")
	}
	if expected == "" || actual != expected {
		return fmt.Errorf("the configuration is not available in the current workspace")
	}
	return nil
}

func (s *AdminServer) latestThreadOutcomeTemplate(
	ctx context.Context,
	sessionID string,
	messages []chatRequestMessage,
) (protocol.ConfigDocument, bool) {
	document, ok := s.latestThreadConfigDocument(ctx, sessionID, messages)
	if ok && document.Kind == protocol.ConfigDocumentKindOutcomeTemplate {
		return document, true
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
