package deploymentcontext

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
)

func (s *Service) Promote(ctx context.Context, req PromoteRequest) (*IngestResult, error) {
	if err := s.Ready(); err != nil {
		return nil, err
	}
	if s.Artifacts == nil || s.Artifacts.DB == nil {
		return nil, fmt.Errorf("deployment context store unavailable")
	}

	sourceArtifactID := strings.TrimSpace(req.SourceArtifactID)
	if sourceArtifactID == "" {
		return nil, fmt.Errorf("source_artifact_id is required")
	}

	var (
		sourceTitle   string
		sourceContent sql.NullString
		sourceMetaRaw []byte
	)
	err := s.Artifacts.DB.QueryRowContext(ctx, `
		SELECT title, content, metadata
		FROM artifacts
		WHERE id::text = $1
	`, sourceArtifactID).Scan(&sourceTitle, &sourceContent, &sourceMetaRaw)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("source deployment context entry not found")
	}
	if err != nil {
		return nil, fmt.Errorf("load source deployment context entry: %w", err)
	}

	sourceMeta := map[string]any{}
	if len(sourceMetaRaw) > 0 {
		_ = json.Unmarshal(sourceMetaRaw, &sourceMeta)
	}
	sourceKnowledgeClass := stringMeta(sourceMeta, "knowledge_class", KnowledgeClassCustomerContext)
	if sourceKnowledgeClass != KnowledgeClassCustomerContext {
		return nil, fmt.Errorf("only customer_context entries can be promoted into company_knowledge")
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		title = strings.TrimSpace(sourceTitle)
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		content = strings.TrimSpace(sourceContent.String)
	}
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("promotion content is required")
	}

	sourceLabel := strings.TrimSpace(req.SourceLabel)
	if sourceLabel == "" {
		sourceLabel = fmt.Sprintf("promoted from %s", strings.TrimSpace(sourceTitle))
	}
	sourceKind := strings.TrimSpace(req.SourceKind)
	if sourceKind == "" {
		sourceKind = stringMeta(sourceMeta, "source_kind", "user_document")
	}
	visibility := strings.TrimSpace(req.Visibility)
	if visibility == "" {
		visibility = stringMeta(sourceMeta, "visibility", "global")
	}
	sensitivityClass := strings.TrimSpace(req.SensitivityClass)
	if sensitivityClass == "" {
		sensitivityClass = stringMeta(sourceMeta, "sensitivity_class", "role_scoped")
	}
	trustClass := strings.TrimSpace(req.TrustClass)
	if trustClass == "" {
		trustClass = "trusted_internal"
	}
	tags := req.Tags
	if len(tags) == 0 {
		tags = append(tags, stringSliceMeta(sourceMeta, "tags")...)
	}
	tags = append(tags, "promoted-company-knowledge")

	return s.Ingest(ctx, IngestRequest{
		KnowledgeClass:   KnowledgeClassCompanyKnowledge,
		Title:            title,
		Content:          content,
		ContentType:      req.ContentType,
		SourceLabel:      sourceLabel,
		SourceKind:       sourceKind,
		Visibility:       visibility,
		SensitivityClass: sensitivityClass,
		TrustClass:       trustClass,
		Tags:             tags,
		AgentID:          req.AgentID,
		TeamID:           req.TeamID,
		UserLabel:        req.UserLabel,
		ExtraMetadata: map[string]any{
			"promotion_kind":                "customer_to_company",
			"promoted_from_artifact_id":     sourceArtifactID,
			"promoted_from_knowledge_class": sourceKnowledgeClass,
			"promotion_source_title":        strings.TrimSpace(sourceTitle),
			"promotion_source_label":        stringMeta(sourceMeta, "source_label", "operator provided"),
			"promoted_by":                   strings.TrimSpace(req.UserLabel),
		},
	})
}
