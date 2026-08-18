package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mycelis/core/internal/configdocuments"
	"github.com/mycelis/core/pkg/protocol"
)

func isConfigDocumentMutationTool(name string) bool {
	switch strings.ToLower(strings.TrimSpace(name)) {
	case "store_config_document", "activate_config_document":
		return true
	default:
		return false
	}
}

func (s *AdminServer) executeConfigDocumentMutationTx(
	ctx context.Context,
	tx *sql.Tx,
	toolName string,
	arguments map[string]any,
	scope *protocol.ScopeValidation,
	actorID string,
) (string, error) {
	store := configdocuments.NewStore(s.getDB())
	switch strings.ToLower(strings.TrimSpace(toolName)) {
	case "store_config_document":
		var boundary *protocol.ConfigDocumentRequestBoundary
		if scope != nil {
			boundary = scope.ConfigRequestBoundary
		}
		document, err := configDocumentFromConfirmedArguments(arguments, boundary)
		if err != nil {
			return "", err
		}
		record, err := store.StoreRevisionTx(ctx, tx, "default", actorID, document)
		if err != nil {
			return "", fmt.Errorf("store config document: %w", err)
		}
		return configDocumentMutationJSON(map[string]any{
			"message":  "Configuration revision saved. It is not active until approved and activated.",
			"revision": record,
		})
	case "activate_config_document":
		recordID := firstNonEmptyString(arguments["record_id"])
		if recordID == "" {
			return "", fmt.Errorf("activate_config_document requires 'record_id'")
		}
		action := configdocuments.ActivationAction(firstNonEmptyString(arguments["action"]))
		if action == "" {
			action = configdocuments.ActivationActionActivate
		}
		result, err := store.ActivateRevisionTx(ctx, tx, "default", recordID, actorID, "", action)
		if err != nil {
			return "", fmt.Errorf("activate config document: %w", err)
		}
		return configDocumentMutationJSON(map[string]any{
			"message":    "Configuration revision is active.",
			"activation": result,
		})
	default:
		return "", fmt.Errorf("unsupported configuration mutation %q", toolName)
	}
}

func configDocumentFromConfirmedArguments(arguments map[string]any, boundary *protocol.ConfigDocumentRequestBoundary) (protocol.ConfigDocument, error) {
	content, _ := arguments["content"].(string)
	path, _ := arguments["path"].(string)
	if (strings.TrimSpace(content) == "") == (strings.TrimSpace(path) == "") {
		return protocol.ConfigDocument{}, fmt.Errorf("provide exactly one of 'content' or 'path'")
	}
	var document protocol.ConfigDocument
	var err error
	if strings.TrimSpace(path) != "" {
		document, err = configdocuments.LoadDocumentFile(configdocuments.ConfiguredRoot(), path)
	} else {
		format, _ := arguments["format"].(string)
		document, err = configdocuments.ParseDocument([]byte(content), format)
	}
	if err != nil {
		return protocol.ConfigDocument{}, err
	}
	if issues := protocol.ValidateConfigDocument(document); len(issues) != 0 {
		return protocol.ConfigDocument{}, &configdocuments.ValidationError{Issues: issues}
	}
	if err := validateConfirmedConfigDocumentBoundary(document, boundary); err != nil {
		return protocol.ConfigDocument{}, err
	}
	return document, nil
}

func validateConfirmedConfigDocumentBoundary(document protocol.ConfigDocument, boundary *protocol.ConfigDocumentRequestBoundary) error {
	if boundary == nil {
		return fmt.Errorf("confirmed config document request boundary is missing")
	}
	scope := document.Metadata.Scope
	switch scope.Kind {
	case protocol.ConfigDocumentScopeBuiltIn:
		return fmt.Errorf("built-in config document scope cannot be stored by a confirmed action")
	case protocol.ConfigDocumentScopeOrganization:
		if matchesConfigDocumentBoundary(scope.Ref, boundary.OrganizationID) {
			return nil
		}
	case protocol.ConfigDocumentScopeWorkspace:
		if matchesConfigDocumentBoundary(scope.Ref, boundary.WorkspaceID, boundary.TeamID) {
			return nil
		}
	case protocol.ConfigDocumentScopeOperator:
		if matchesConfigDocumentBoundary(scope.Ref, boundary.OperatorID) {
			return nil
		}
	default:
		return fmt.Errorf("unsupported config document scope %q", scope.Kind)
	}
	return fmt.Errorf("config document %s scope %q is outside the approved request boundary", scope.Kind, scope.Ref)
}

func matchesConfigDocumentBoundary(actual string, approved ...string) bool {
	actual = strings.TrimSpace(actual)
	if actual == "" {
		return false
	}
	for _, candidate := range approved {
		if actual == strings.TrimSpace(candidate) && strings.TrimSpace(candidate) != "" {
			return true
		}
	}
	return false
}

func configDocumentMutationJSON(value any) (string, error) {
	encoded, err := json.Marshal(value)
	if err != nil {
		return "", fmt.Errorf("encode configuration result: %w", err)
	}
	return string(encoded), nil
}
