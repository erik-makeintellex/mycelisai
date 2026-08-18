package server

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mycelis/core/internal/configdocuments"
	"github.com/mycelis/core/pkg/protocol"
)

type configDocumentReceipt struct {
	RecordID string
	Name     string
	Version  string
	Digest   string
	Scope    protocol.ConfigDocumentScope
}

func configDocumentReceiptFromResult(result plannedToolExecutionResult) (configDocumentReceipt, bool) {
	var output struct {
		Revision   *configdocuments.RevisionRecord   `json:"revision"`
		Activation *configdocuments.ActivationResult `json:"activation"`
	}
	if err := json.Unmarshal([]byte(result.Output), &output); err != nil {
		return configDocumentReceipt{}, false
	}
	revision := output.Revision
	if output.Activation != nil {
		revision = &output.Activation.Revision
	}
	if revision == nil || strings.TrimSpace(revision.RecordID) == "" {
		return configDocumentReceipt{}, false
	}
	return configDocumentReceipt{
		RecordID: revision.RecordID,
		Name:     revision.Document.Metadata.Name,
		Version:  revision.Document.Metadata.Version,
		Digest:   revision.Digest,
		Scope:    revision.Document.Metadata.Scope,
	}, true
}

func configDocumentVisibleLabel(receipt configDocumentReceipt) string {
	name := firstNonEmptyString(receipt.Name, "Outcome Template")
	if version := strings.TrimSpace(receipt.Version); version != "" {
		return fmt.Sprintf("%s v%s", name, version)
	}
	return name
}

func configDocumentScopeLabel(scope protocol.ConfigDocumentScope) string {
	switch scope.Kind {
	case protocol.ConfigDocumentScopeWorkspace:
		return "this workspace"
	case protocol.ConfigDocumentScopeOrganization:
		return "this organization"
	case protocol.ConfigDocumentScopeOperator:
		return "this operator"
	default:
		return "its configured scope"
	}
}

func configDocumentResultSummary(result plannedToolExecutionResult) string {
	receipt, ok := configDocumentReceiptFromResult(result)
	if !ok {
		if result.Name == "store_config_document" {
			return "Outcome Template saved but not active."
		}
		return "Outcome Template active for its configured scope."
	}
	label := configDocumentVisibleLabel(receipt)
	if result.Name == "store_config_document" {
		return label + " saved but not active."
	}
	return fmt.Sprintf("%s is active for %s.", label, configDocumentScopeLabel(receipt.Scope))
}

func firstConfigDocumentResult(results []plannedToolExecutionResult, name string) (plannedToolExecutionResult, bool) {
	for _, result := range results {
		if strings.EqualFold(strings.TrimSpace(result.Name), name) {
			return result, true
		}
	}
	return plannedToolExecutionResult{}, false
}

func isSynchronousConfigAction(results []plannedToolExecutionResult) bool {
	if len(results) == 0 {
		return false
	}
	for _, result := range results {
		switch strings.ToLower(strings.TrimSpace(result.Name)) {
		case "store_config_document", "activate_config_document":
			continue
		default:
			return false
		}
	}
	return true
}
