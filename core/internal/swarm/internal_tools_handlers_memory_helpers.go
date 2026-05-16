package swarm

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/exchange"
	"github.com/mycelis/core/internal/memory"
)

type memoryResult struct {
	Category  string  `json:"category"`
	Content   string  `json:"content"`
	Context   string  `json:"context,omitempty"`
	CreatedAt string  `json:"created_at"`
	Score     float64 `json:"score,omitempty"`
	Source    string  `json:"source"`
}

type parsedConversationSummary struct {
	Summary          string         `json:"summary"`
	KeyTopics        []string       `json:"key_topics"`
	UserPreferences  map[string]any `json:"user_preferences"`
	PersonalityNotes string         `json:"personality_notes"`
	DataReferences   []any          `json:"data_references"`
}

func validateArtifactContent(artType, content string) (string, error) {
	if artType == "project_package" {
		return "application/vnd.mycelis.project+json", nil
	}
	if artType != "chart" {
		return "text/plain", nil
	}
	var spec map[string]any
	if err := json.Unmarshal([]byte(content), &spec); err != nil {
		return "", fmt.Errorf("chart content must be valid JSON: %w", err)
	}
	if _, ok := spec["chart_type"]; !ok {
		return "", fmt.Errorf("chart spec missing required 'chart_type' field")
	}
	dataArr, ok := spec["data"].([]any)
	if !ok {
		return "", fmt.Errorf("chart spec missing required 'data' array")
	}
	if len(dataArr) > 2000 {
		return "", fmt.Errorf("chart data exceeds 2000-row limit (%d rows); summarize or aggregate first", len(dataArr))
	}
	return "application/vnd.mycelis.chart+json", nil
}

func artifactMetadataJSON(metadata any) string {
	if metadata == nil {
		return "{}"
	}
	if b, err := json.Marshal(metadata); err == nil {
		return string(b)
	}
	return "{}"
}

func (r *InternalToolRegistry) insertArtifact(ctx context.Context, artType, title, contentType, content, metaJSON string) (string, error) {
	var artifactID string
	err := r.db.QueryRowContext(ctx, `
		INSERT INTO artifacts (agent_id, artifact_type, title, content_type, content, metadata, status)
		VALUES ('internal', $1, $2, $3, $4, $5, 'pending')
		RETURNING id
	`, artType, title, contentType, content, metaJSON).Scan(&artifactID)
	return artifactID, err
}

func artifactResultPayload(artifactID, artType, title, contentType, content string, metadata any) map[string]any {
	if packageKindFromMetadata(metadata) == "project_package" && (artType == "file" || artType == "data") {
		artType = "project_package"
	}
	if artType == "project_package" && (strings.TrimSpace(contentType) == "" || contentType == "text/plain") {
		contentType = "application/vnd.mycelis.project+json"
	}
	artifactRef := map[string]any{"id": artifactID, "type": artType, "title": title, "content_type": contentType}
	if artType != "image" && artType != "audio" && len(content) < 500_000 {
		artifactRef["content"] = content
	}
	if artType == "project_package" {
		for key, value := range projectPackageMetadata(metadata) {
			artifactRef[key] = value
		}
	}
	return map[string]any{"message": fmt.Sprintf("Artifact '%s' stored (type: %s, id: %s).", title, artType, artifactID), "artifact": artifactRef}
}

func projectPackageMetadata(metadata any) map[string]any {
	source, _ := metadata.(map[string]any)
	if len(source) == 0 {
		return map[string]any{}
	}
	out := map[string]any{}
	if value := firstNonEmptyMetadataString(source, "entrypoint", "package_entrypoint", "path", "file_path", "target_path"); value != "" {
		out["entrypoint"] = value
	}
	if value := firstNonEmptyMetadataString(source, "folder", "package_folder", "project_folder"); value != "" {
		out["folder"] = value
	}
	if value := firstNonEmptyMetadataString(source, "validation", "validation_summary", "proof_summary"); value != "" {
		out["validation"] = value
	}
	if files := firstMetadataStringSlice(source, "files", "package_files"); len(files) > 0 {
		out["files"] = files
	}
	return out
}

func packageKindFromMetadata(metadata any) string {
	source, _ := metadata.(map[string]any)
	if len(source) == 0 {
		return ""
	}
	return firstNonEmptyMetadataString(source, "package_kind", "output_kind", "kind")
}

func firstNonEmptyMetadataString(source map[string]any, keys ...string) string {
	for _, key := range keys {
		if value := strings.TrimSpace(stringValue(source[key])); value != "" {
			return value
		}
	}
	return ""
}

func firstMetadataStringSlice(source map[string]any, keys ...string) []string {
	for _, key := range keys {
		if values := stringSlice(source[key]); len(values) > 0 {
			return values
		}
		if value := strings.TrimSpace(stringValue(source[key])); value != "" {
			var decoded []string
			if strings.HasPrefix(value, "[") && json.Unmarshal([]byte(value), &decoded) == nil {
				return normalizeStringSlice(decoded)
			}
			return normalizeStringSlice(strings.Split(value, ","))
		}
	}
	return nil
}

func normalizeStringSlice(items []string) []string {
	out := make([]string, 0, len(items))
	seen := map[string]struct{}{}
	for _, item := range items {
		trimmed := strings.TrimSpace(item)
		if trimmed == "" {
			continue
		}
		if _, ok := seen[trimmed]; ok {
			continue
		}
		seen[trimmed] = struct{}{}
		out = append(out, trimmed)
	}
	return out
}

func publishArtifactToExchange(ctx context.Context, svc *exchange.Service, artifactID, artType, title string) {
	parsedID, err := uuid.Parse(artifactID)
	if err != nil {
		return
	}
	runID, runClass, noRunReason := artifactExchangeRuntimeMetadata(ctx)
	_, _ = svc.PublishArtifact(ctx, exchange.ArtifactNormalizationInput{
		ArtifactID:     parsedID,
		ArtifactType:   artType,
		Title:          title,
		AgentID:        "internal",
		RunID:          runID,
		RunClass:       runClass,
		NoRunReason:    noRunReason,
		RetentionClass: "retained",
		Status:         "pending",
		TargetRole:     "soma",
		Tags:           []string{"artifact", artType},
	})
}

func artifactExchangeRuntimeMetadata(ctx context.Context) (runID, runClass, noRunReason string) {
	if inv, ok := ToolInvocationContextFromContext(ctx); ok {
		if runID = strings.TrimSpace(inv.RunID); runID != "" {
			return runID, "run_linked", ""
		}
	}
	return "", "no_run", "Artifact normalization did not receive an execution run id."
}

func storeMemoryVector(ctx context.Context, brain *cognitive.Router, mem *memory.Service, category, content, memContext string, scope memoryScope) {
	if brain == nil || mem == nil {
		return
	}
	embeddingText := fmt.Sprintf("[%s] %s", category, content)
	if memContext != "" {
		embeddingText += " (context: " + memContext + ")"
	}
	vec, err := brain.Embed(ctx, embeddingText, "")
	if err != nil {
		log.Printf("remember: embedding failed (non-fatal): %v", err)
		return
	}
	meta := map[string]any{"type": "agent_memory", "category": category, "source": "agent_memory", "tenant_id": scope.TenantID, "team_id": scope.TeamID, "agent_id": scope.AgentID, "run_id": scope.RunID, "visibility": scope.Visibility}
	if insertErr := mem.StoreVector(ctx, embeddingText, vec, meta); insertErr != nil {
		log.Printf("remember: vector insert failed: %v", insertErr)
	}
}
