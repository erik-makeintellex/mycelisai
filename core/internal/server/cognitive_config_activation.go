package server

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/mycelis/core/internal/configdocuments"
	"github.com/mycelis/core/pkg/protocol"
)

var requestedConfigVersionPattern = regexp.MustCompile(`(?i)\bversion\s+["']?([a-z0-9](?:[a-z0-9._-]*[a-z0-9])?)`)
var requestedConfigRollbackPattern = regexp.MustCompile(`(?i)\b(?:roll\s*back|revert|restore)\b`)

func (s *AdminServer) resolveThreadConfigurationMutationsOrRespond(
	w http.ResponseWriter,
	r *http.Request,
	sessionID string,
	messages []chatRequestMessage,
	latestRequest string,
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
	resolved, err = s.applyThreadWorkerProfileSelection(
		r.Context(), sessionID, messages, latestRequest, organizationID, teamID, actorID, resolved,
	)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusConflict)
		return false
	}
	*planned = resolved
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
	if !plannedContainsTool(planned, "activate_config_document") {
		return planned, nil
	}
	store, err := s.runtimeConfigDocumentStore()
	if err != nil {
		return nil, err
	}
	document, ok := s.latestThreadConfigDocument(ctx, sessionID, messages)
	if !ok {
		return nil, fmt.Errorf("Soma could not identify which saved configuration revision to activate")
	}
	if err := validateOutcomeTemplateRequestScope(document, organizationID, teamID, actorID); err != nil {
		return nil, err
	}
	digest, err := protocol.CanonicalConfigDocumentDigest(document)
	if err != nil {
		return nil, err
	}
	targetVersion := requestedConfigDocumentVersion(messages)
	requestedAction := requestedConfigDocumentActivationAction(messages)

	for i, call := range planned {
		if !strings.EqualFold(strings.TrimSpace(call.Name), "activate_config_document") {
			continue
		}
		action := configdocuments.ActivationAction(firstNonEmptyString(call.Arguments["action"]))
		if requestedAction == configdocuments.ActivationActionRollback {
			action = requestedAction
		} else if action == "" {
			action = configdocuments.ActivationActionActivate
		}
		recordID := firstNonEmptyString(call.Arguments["record_id"])
		if recordID == "" {
			recordID, err = resolveThreadConfigRecordID(ctx, store, document, digest, action, targetVersion)
			if err != nil {
				return nil, err
			}
		} else if err := validateSelectedConfigRecord(
			ctx, store, recordID, document, digest, action, targetVersion,
			organizationID, teamID, actorID,
		); err != nil {
			return nil, err
		}
		if call.Arguments == nil {
			call.Arguments = map[string]any{}
		}
		call.Arguments["record_id"] = recordID
		call.Arguments["action"] = string(action)
		planned[i] = call
	}
	return planned, nil
}

func resolveThreadConfigRecordID(
	ctx context.Context,
	store *configdocuments.Store,
	document protocol.ConfigDocument,
	digest string,
	action configdocuments.ActivationAction,
	targetVersion string,
) (string, error) {
	if action == configdocuments.ActivationActionRollback && targetVersion == "" {
		return "", fmt.Errorf("Soma needs the target configuration version before rollback")
	}
	records, err := store.List(ctx, "default", configdocuments.ListFilter{
		Kind: document.Kind, ScopeKind: document.Metadata.Scope.Kind,
		ScopeRef: document.Metadata.Scope.Ref, DocumentID: document.Metadata.ID, Limit: 20,
	})
	if err != nil {
		return "", err
	}
	for _, record := range records {
		if action == configdocuments.ActivationActionRollback {
			if strings.EqualFold(record.Document.Metadata.Version, targetVersion) && configRecordDigestValid(&record) {
				return record.RecordID, nil
			}
		} else if record.Digest == digest {
			return record.RecordID, nil
		}
	}
	if action == configdocuments.ActivationActionRollback {
		return "", fmt.Errorf("configuration version %q is not saved in this scope", targetVersion)
	}
	return "", fmt.Errorf("the configuration revision from this thread has not been saved")
}

func validateSelectedConfigRecord(
	ctx context.Context,
	store *configdocuments.Store,
	recordID string,
	document protocol.ConfigDocument,
	digest string,
	action configdocuments.ActivationAction,
	targetVersion string,
	organizationID string,
	teamID string,
	actorID string,
) error {
	record, err := store.GetRevision(ctx, "default", recordID)
	if err != nil {
		return fmt.Errorf("load selected configuration revision: %w", err)
	}
	if err := validateOutcomeTemplateRequestScope(record.Document, organizationID, teamID, actorID); err != nil {
		return err
	}
	if record.Document.Kind != document.Kind || record.Document.Metadata.ID != document.Metadata.ID {
		return fmt.Errorf("the selected configuration revision does not match this thread")
	}
	if !configRecordDigestValid(record) {
		return fmt.Errorf("the selected configuration revision failed integrity validation")
	}
	if action == configdocuments.ActivationActionRollback {
		if targetVersion != "" && !strings.EqualFold(record.Document.Metadata.Version, targetVersion) {
			return fmt.Errorf("the selected configuration revision is not version %q", targetVersion)
		}
		return nil
	}
	recordDigest, digestErr := protocol.CanonicalConfigDocumentDigest(record.Document)
	if digestErr != nil || record.Document.Metadata.Version != document.Metadata.Version ||
		record.Digest != digest || recordDigest != digest {
		return fmt.Errorf("the selected configuration revision does not match this thread")
	}
	return nil
}

func configRecordDigestValid(record *configdocuments.RevisionRecord) bool {
	digest, err := protocol.CanonicalConfigDocumentDigest(record.Document)
	return err == nil && digest == record.Digest
}

func (s *AdminServer) latestThreadConfigDocument(
	ctx context.Context,
	sessionID string,
	messages []chatRequestMessage,
) (protocol.ConfigDocument, bool) {
	for i := len(messages) - 1; i >= 0; i-- {
		_, _, document, ok := parseInlineConfigDocument(messages[i].Content)
		if ok && (document.Kind == protocol.ConfigDocumentKindOutcomeTemplate ||
			document.Kind == protocol.ConfigDocumentKindWorkerProfile) {
			return document, true
		}
	}
	if receipt, ok := s.latestSessionConfigDocumentReceipt(ctx, sessionID); ok {
		store, err := s.runtimeConfigDocumentStore()
		if err == nil {
			record, loadErr := store.GetRevision(ctx, "default", receipt.RecordID)
			if loadErr == nil && record.Digest == receipt.Digest {
				return record.Document, true
			}
		}
	}
	return protocol.ConfigDocument{}, false
}

func requestedConfigDocumentVersion(messages []chatRequestMessage) string {
	for i := len(messages) - 1; i >= 0; i-- {
		if !strings.EqualFold(strings.TrimSpace(messages[i].Role), "user") {
			continue
		}
		match := requestedConfigVersionPattern.FindStringSubmatch(messages[i].Content)
		if len(match) == 2 {
			return strings.TrimSpace(match[1])
		}
		return ""
	}
	return ""
}

func requestedConfigDocumentActivationAction(messages []chatRequestMessage) configdocuments.ActivationAction {
	for i := len(messages) - 1; i >= 0; i-- {
		if !strings.EqualFold(strings.TrimSpace(messages[i].Role), "user") {
			continue
		}
		if requestedConfigRollbackPattern.MatchString(messages[i].Content) {
			return configdocuments.ActivationActionRollback
		}
		return configdocuments.ActivationActionActivate
	}
	return configdocuments.ActivationActionActivate
}

func plannedContainsTool(planned []protocol.PlannedToolCall, name string) bool {
	for _, call := range planned {
		if strings.EqualFold(strings.TrimSpace(call.Name), name) {
			return true
		}
	}
	return false
}
