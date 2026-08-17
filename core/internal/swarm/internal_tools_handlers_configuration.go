package swarm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mycelis/core/internal/configdocuments"
	"github.com/mycelis/core/pkg/protocol"
)

func (r *InternalToolRegistry) handlePreviewConfigDocument(_ context.Context, args map[string]any) (string, error) {
	document, err := configDocumentFromToolArgs(args)
	if err != nil {
		return "", err
	}
	dryRun := protocol.DryRunConfigDocument(document)
	result := map[string]any{"dry_run": dryRun}
	if dryRun.Valid && document.Kind == protocol.ConfigDocumentKindOutcomeTemplate {
		compiled, err := configdocuments.CompileOutcomeTemplateDocument(
			document,
			minimumBriefFromArgs(args["operator_values"]),
			minimumBriefFromArgs(args["policy_values"]),
		)
		if err != nil {
			return "", fmt.Errorf("preview config document: %w", err)
		}
		result["compiled"] = compiled
	}
	return mustJSON(result), nil
}

func (r *InternalToolRegistry) handleStoreConfigDocument(ctx context.Context, args map[string]any) (string, error) {
	if r.db == nil {
		return "", fmt.Errorf("config document store not available")
	}
	document, err := configDocumentFromToolArgs(args)
	if err != nil {
		return "", err
	}
	scope := resolveMemoryScope(ctx, args)
	record, err := configdocuments.NewStore(r.db).StoreRevision(ctx, "default", scope.AgentID, document)
	if err != nil {
		return "", fmt.Errorf("store config document: %w", err)
	}
	return mustJSON(map[string]any{
		"message":  "Configuration revision saved. It is not active until approved and activated.",
		"revision": record,
	}), nil
}

func (r *InternalToolRegistry) handleActivateConfigDocument(ctx context.Context, args map[string]any) (string, error) {
	if r.db == nil {
		return "", fmt.Errorf("config document store not available")
	}
	recordID := strings.TrimSpace(stringValue(args["record_id"]))
	if recordID == "" {
		return "", fmt.Errorf("activate_config_document requires 'record_id'")
	}
	action := configdocuments.ActivationAction(strings.TrimSpace(stringValue(args["action"])))
	if action == "" {
		action = configdocuments.ActivationActionActivate
	}
	scope := resolveMemoryScope(ctx, args)
	result, err := configdocuments.NewStore(r.db).ActivateRevision(ctx, "default", recordID, scope.AgentID, "", action)
	if err != nil {
		return "", fmt.Errorf("activate config document: %w", err)
	}
	return mustJSON(map[string]any{"message": "Configuration revision is active.", "activation": result}), nil
}

func configDocumentFromToolArgs(args map[string]any) (protocol.ConfigDocument, error) {
	content := stringValue(args["content"])
	path := stringValue(args["path"])
	if (strings.TrimSpace(content) == "") == (strings.TrimSpace(path) == "") {
		return protocol.ConfigDocument{}, fmt.Errorf("provide exactly one of 'content' or 'path'")
	}
	if path != "" {
		return configdocuments.LoadDocumentFile(configdocuments.ConfiguredRoot(), path)
	}
	return configdocuments.ParseDocument([]byte(content), stringValue(args["format"]))
}

func minimumBriefFromArgs(raw any) protocol.MinimumSufficientBrief {
	encoded, err := json.Marshal(raw)
	if err != nil || string(encoded) == "null" {
		return protocol.MinimumSufficientBrief{}
	}
	var brief protocol.MinimumSufficientBrief
	_ = json.Unmarshal(encoded, &brief)
	return brief
}
