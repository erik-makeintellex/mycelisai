package server

import (
	"context"
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) applyThreadWorkerProfileSelection(
	ctx context.Context,
	sessionID string,
	messages []chatRequestMessage,
	latestRequest string,
	organizationID string,
	teamID string,
	actorID string,
	planned []protocol.PlannedToolCall,
) ([]protocol.PlannedToolCall, error) {
	if !plannedContainsTool(planned, "create_team") || !requestReferencesThreadWorkerProfile(latestRequest) {
		return planned, nil
	}
	document, ok := s.latestThreadConfigDocument(ctx, sessionID, messages)
	if !ok || document.Kind != protocol.ConfigDocumentKindWorkerProfile {
		return nil, fmt.Errorf("Soma could not identify the Worker Profile referenced by this team request")
	}
	if err := validateOutcomeTemplateRequestScope(document, organizationID, teamID, actorID); err != nil {
		return nil, err
	}
	store, err := s.runtimeConfigDocumentStore()
	if err != nil {
		return nil, err
	}
	active, err := store.GetActiveRevision(ctx, "default", document.Kind, document.Metadata.ID, document.Metadata.Scope)
	if err != nil {
		return nil, fmt.Errorf("activate this Worker Profile before assigning it to a team: %w", err)
	}
	if err := validateOutcomeTemplateRequestScope(active.Document, organizationID, teamID, actorID); err != nil {
		return nil, err
	}

	for index, call := range planned {
		if !strings.EqualFold(strings.TrimSpace(call.Name), "create_team") {
			continue
		}
		if call.Arguments == nil {
			call.Arguments = map[string]any{}
		}
		delete(call.Arguments, "profile_snapshot")
		call.Arguments["profile_ref"] = document.Metadata.ID
		planned[index] = call
	}
	return planned, nil
}

func requestReferencesThreadWorkerProfile(request string) bool {
	lower := strings.ToLower(strings.TrimSpace(request))
	if !strings.Contains(lower, "worker profile") {
		return false
	}
	return strings.Contains(lower, "this worker profile") ||
		strings.Contains(lower, "the worker profile") ||
		strings.Contains(lower, "using worker profile") ||
		strings.Contains(lower, "use worker profile")
}
