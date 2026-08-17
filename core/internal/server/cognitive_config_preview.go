package server

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/mycelis/core/internal/configdocuments"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

type configPreviewToolResult struct {
	DryRun protocol.ConfigDocumentDryRunResult `json:"dry_run"`
}

func (s *AdminServer) executeRequestedConfigPreview(ctx context.Context, request string, fallback chatAgentResult) (chatAgentResult, bool) {
	if !containsToolName(inferReadOnlyConfigToolsFromText(request), "preview_config_document") {
		return chatAgentResult{}, false
	}

	content, format, document, ok := parseInlineConfigDocument(request)
	if !ok {
		content, format, document, ok = parseInlineConfigDocument(fallback.Text)
	}
	if !ok {
		return chatAgentResult{}, false
	}

	registry := swarm.NewInternalToolRegistry(swarm.InternalToolDeps{})
	tool := registry.Get("preview_config_document")
	if tool == nil {
		return configPreviewFailure(fallback, "Outcome Template preview is unavailable in this runtime."), true
	}
	raw, err := tool.Handler(ctx, map[string]any{"content": content, "format": format})
	if err != nil {
		return configPreviewFailure(fallback, "Soma could not validate this Outcome Template: "+err.Error()), true
	}

	var preview configPreviewToolResult
	if err := json.Unmarshal([]byte(raw), &preview); err != nil {
		return configPreviewFailure(fallback, "Soma validated the Outcome Template but could not read the preview result."), true
	}
	result := fallback
	result.ToolsUsed = []string{"preview_config_document"}
	result.PlannedToolCalls = nil
	result.Artifacts = nil
	result.Availability = nil
	result.Text = readableConfigPreview(document, preview.DryRun)
	return result, true
}

func parseInlineConfigDocument(text string) (string, string, protocol.ConfigDocument, bool) {
	trimmed := strings.TrimSpace(text)
	if trimmed == "" {
		return "", "", protocol.ConfigDocument{}, false
	}
	candidates := []struct {
		content string
		format  string
	}{}
	if index := strings.Index(trimmed, "apiVersion:"); index >= 0 {
		candidates = append(candidates, struct {
			content string
			format  string
		}{strings.TrimSpace(strings.TrimSuffix(trimmed[index:], "```")), "yaml"})
	}
	if index := strings.Index(trimmed, `{"apiVersion"`); index >= 0 {
		candidates = append(candidates, struct {
			content string
			format  string
		}{strings.TrimSpace(strings.TrimSuffix(trimmed[index:], "```")), "json"})
	}

	for _, candidate := range candidates {
		document, err := configdocuments.ParseDocument([]byte(candidate.content), candidate.format)
		if err == nil {
			return candidate.content, candidate.format, document, true
		}
	}
	return "", "", protocol.ConfigDocument{}, false
}

func readableConfigPreview(document protocol.ConfigDocument, preview protocol.ConfigDocumentDryRunResult) string {
	if preview.Valid {
		scope := string(preview.Effect.Scope.Kind)
		if preview.Effect.Scope.Ref != "" {
			scope += " " + preview.Effect.Scope.Ref
		}
		return fmt.Sprintf(
			"Outcome Template preview is valid. `%s` version `%s` would be available for %s after approval and activation. Nothing was saved or activated.",
			document.Metadata.Name, document.Metadata.Version, scope,
		)
	}
	if len(preview.Issues) == 0 {
		return "Outcome Template preview is invalid. Nothing was saved or activated."
	}
	issue := preview.Issues[0]
	return fmt.Sprintf("Outcome Template preview needs correction: %s: %s. Nothing was saved or activated.", issue.Field, issue.Message)
}

func configPreviewFailure(fallback chatAgentResult, message string) chatAgentResult {
	fallback.Text = message
	fallback.ToolsUsed = []string{"preview_config_document"}
	fallback.PlannedToolCalls = nil
	fallback.Artifacts = nil
	fallback.Availability = nil
	return fallback
}
