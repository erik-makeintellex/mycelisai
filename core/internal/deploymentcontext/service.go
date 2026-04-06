package deploymentcontext

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/memory"
)

const (
	defaultChunkSize               = 1200
	defaultChunkOverlap            = 160
	KnowledgeClassCustomerContext  = "customer_context"
	KnowledgeClassCompanyKnowledge = "company_knowledge"
	KnowledgeClassSomaOperating    = "soma_operating_context"
)

type IngestRequest struct {
	KnowledgeClass    string
	Title             string
	Content           string
	ContentType       string
	SourceLabel       string
	SourceKind        string
	Visibility        string
	SensitivityClass  string
	TrustClass        string
	Tags              []string
	AgentID           string
	TeamID            string
	UserLabel         string
	SomaContextKind   string
	OutputSpecificity string
	ExtraMetadata     map[string]any
}

type PromoteRequest struct {
	SourceArtifactID string
	Title            string
	Content          string
	ContentType      string
	SourceLabel      string
	SourceKind       string
	Visibility       string
	SensitivityClass string
	TrustClass       string
	Tags             []string
	AgentID          string
	TeamID           string
	UserLabel        string
}

type IngestResult struct {
	ArtifactID       string    `json:"artifact_id"`
	KnowledgeClass   string    `json:"knowledge_class"`
	Title            string    `json:"title"`
	SourceLabel      string    `json:"source_label"`
	SourceKind       string    `json:"source_kind"`
	Visibility       string    `json:"visibility"`
	SensitivityClass string    `json:"sensitivity_class"`
	TrustClass       string    `json:"trust_class"`
	ChunkCount       int       `json:"chunk_count"`
	VectorCount      int       `json:"vector_count"`
	ContentPreview   string    `json:"content_preview"`
	ContentLength    int       `json:"content_length"`
	CreatedAt        time.Time `json:"created_at"`
}

type Entry struct {
	ArtifactID       string    `json:"artifact_id"`
	KnowledgeClass   string    `json:"knowledge_class"`
	Title            string    `json:"title"`
	SourceLabel      string    `json:"source_label"`
	SourceKind       string    `json:"source_kind"`
	Visibility       string    `json:"visibility"`
	SensitivityClass string    `json:"sensitivity_class"`
	TrustClass       string    `json:"trust_class"`
	ChunkCount       int       `json:"chunk_count"`
	VectorCount      int       `json:"vector_count"`
	ContentPreview   string    `json:"content_preview"`
	ContentLength    int       `json:"content_length"`
	CreatedAt        time.Time `json:"created_at"`
}

type Service struct {
	Artifacts *artifacts.Service
	Memory    *memory.Service
	Cognitive *cognitive.Router
}

func NewService(artifactsSvc *artifacts.Service, memorySvc *memory.Service, cognitiveSvc *cognitive.Router) *Service {
	return &Service{
		Artifacts: artifactsSvc,
		Memory:    memorySvc,
		Cognitive: cognitiveSvc,
	}
}

func (s *Service) Ready() error {
	switch {
	case s == nil:
		return fmt.Errorf("deployment context service unavailable")
	case s.Artifacts == nil:
		return fmt.Errorf("artifacts service offline")
	case s.Memory == nil:
		return fmt.Errorf("memory service offline")
	case s.Cognitive == nil:
		return fmt.Errorf("cognitive engine offline")
	default:
		return nil
	}
}

func (s *Service) Ingest(ctx context.Context, req IngestRequest) (*IngestResult, error) {
	if err := s.Ready(); err != nil {
		return nil, err
	}

	title := strings.TrimSpace(req.Title)
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}
	content := strings.TrimSpace(req.Content)
	if content == "" {
		return nil, fmt.Errorf("content is required")
	}

	contentType := normalizeContentType(req.ContentType)
	knowledgeClass := normalizeKnowledgeClass(req.KnowledgeClass)
	sourceLabel := normalizeSourceLabel(req.SourceLabel)
	sourceKind := normalizeSourceKind(req.SourceKind)
	visibility := normalizeVisibility(req.Visibility)
	sensitivityClass := normalizeSensitivityClass(req.SensitivityClass)
	trustClass := normalizeTrustClass(req.TrustClass)
	agentID := strings.TrimSpace(req.AgentID)
	if agentID == "" {
		agentID = "soma"
	}
	teamID := strings.TrimSpace(req.TeamID)
	userLabel := strings.TrimSpace(req.UserLabel)
	if userLabel == "" {
		userLabel = "local-user"
	}
	tags := normalizeTags(req.Tags)
	somaContextKind := normalizeSomaContextKind(req.SomaContextKind)
	outputSpecificity := normalizeOutputSpecificity(req.OutputSpecificity)
	if knowledgeClass == KnowledgeClassSomaOperating {
		if sourceLabel == "operator provided" {
			sourceLabel = "admin guidance"
		}
		if sourceKind == "user_document" {
			sourceKind = "user_note"
		}
		if trustClass == "user_provided" {
			trustClass = "trusted_internal"
		}
		if sensitivityClass == "role_scoped" {
			sensitivityClass = "restricted"
		}
		if visibility == "private" {
			visibility = "global"
		}
		tags = append(tags, "soma-operating-context")
		if outputSpecificity != "" {
			tags = append(tags, "shared-output-specificity")
		}
		tags = normalizeTags(tags)
	}
	chunks := chunkText(content, defaultChunkSize, defaultChunkOverlap)
	if len(chunks) == 0 {
		return nil, fmt.Errorf("content is required")
	}

	// Governed knowledge stays in a dedicated pgvector lane that is separate
	// from Soma's ordinary remembered facts and conversation continuity.
	metadataMap := map[string]any{
		"context_kind":       "governed_knowledge",
		"knowledge_store":    "governed_context_store",
		"knowledge_class":    knowledgeClass,
		"deployment_context": true,
		"source_kind":        sourceKind,
		"source_label":       sourceLabel,
		"visibility":         visibility,
		"sensitivity_class":  sensitivityClass,
		"trust_class":        trustClass,
		"tags":               tags,
		"chunk_count":        len(chunks),
		"vector_count":       len(chunks),
		"content_length":     utf8.RuneCountInString(content),
		"loaded_by":          userLabel,
		"team_id":            teamID,
		"soma_context_kind":  somaContextKind,
		"output_specificity": outputSpecificity,
	}
	for key, value := range req.ExtraMetadata {
		if _, exists := metadataMap[key]; exists {
			continue
		}
		metadataMap[key] = value
	}
	metadataJSON, _ := json.Marshal(metadataMap)

	stored, err := s.Artifacts.Store(ctx, artifacts.Artifact{
		AgentID:      agentID,
		ArtifactType: artifacts.TypeDocument,
		Title:        title,
		ContentType:  contentType,
		Content:      content,
		Metadata:     metadataJSON,
		Status:       "approved",
	})
	if err != nil {
		return nil, fmt.Errorf("store deployment context artifact: %w", err)
	}

	for idx, chunk := range chunks {
		embeddingText := buildEmbeddingText(knowledgeClass, title, sourceLabel, sourceKind, chunk, idx+1, len(chunks))
		vec, err := s.Cognitive.Embed(ctx, embeddingText, "")
		if err != nil {
			return nil, fmt.Errorf("embed deployment context chunk %d: %w", idx+1, err)
		}

		// The vector type is the knowledge class itself so recall can keep
		// customer-provided context separate from approved company knowledge.
		vectorMeta := map[string]any{
			"type":               knowledgeClass,
			"source":             "governed_context_store",
			"knowledge_store":    "governed_context_store",
			"knowledge_class":    knowledgeClass,
			"artifact_id":        stored.ID.String(),
			"artifact_title":     title,
			"tenant_id":          "default",
			"team_id":            teamID,
			"agent_id":           agentID,
			"visibility":         visibility,
			"sensitivity_class":  sensitivityClass,
			"trust_class":        trustClass,
			"source_kind":        sourceKind,
			"source_label":       sourceLabel,
			"tags":               tags,
			"soma_context_kind":  somaContextKind,
			"output_specificity": outputSpecificity,
			"chunk_index":        idx,
			"chunk_count":        len(chunks),
			"loaded_by":          userLabel,
		}
		for key, value := range req.ExtraMetadata {
			if _, exists := vectorMeta[key]; exists {
				continue
			}
			vectorMeta[key] = value
		}

		if err := s.Memory.StoreVector(ctx, embeddingText, vec, vectorMeta); err != nil {
			return nil, fmt.Errorf("store deployment context vector %d: %w", idx+1, err)
		}
	}

	return &IngestResult{
		ArtifactID:       stored.ID.String(),
		KnowledgeClass:   knowledgeClass,
		Title:            title,
		SourceLabel:      sourceLabel,
		SourceKind:       sourceKind,
		Visibility:       visibility,
		SensitivityClass: sensitivityClass,
		TrustClass:       trustClass,
		ChunkCount:       len(chunks),
		VectorCount:      len(chunks),
		ContentPreview:   previewContent(content, 220),
		ContentLength:    utf8.RuneCountInString(content),
		CreatedAt:        stored.CreatedAt,
	}, nil
}

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

func (s *Service) List(ctx context.Context, limit int) ([]Entry, error) {
	if s == nil || s.Artifacts == nil || s.Artifacts.DB == nil {
		return nil, fmt.Errorf("deployment context store unavailable")
	}
	if limit <= 0 {
		limit = 20
	}

	rows, err := s.Artifacts.DB.QueryContext(ctx, `
		SELECT id::text,
		       title,
		       content,
		       metadata,
		       created_at
		FROM artifacts
		WHERE COALESCE(metadata->>'knowledge_store', '') = 'governed_context_store'
		ORDER BY created_at DESC
		LIMIT $1
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("list deployment context entries: %w", err)
	}
	defer rows.Close()

	var entries []Entry
	for rows.Next() {
		var (
			id        string
			title     string
			content   sql.NullString
			metaJSON  []byte
			createdAt time.Time
		)
		if err := rows.Scan(&id, &title, &content, &metaJSON, &createdAt); err != nil {
			return nil, fmt.Errorf("scan deployment context entry: %w", err)
		}

		entry := Entry{
			ArtifactID:       id,
			KnowledgeClass:   KnowledgeClassCustomerContext,
			Title:            title,
			ContentPreview:   previewContent(content.String, 220),
			ContentLength:    utf8.RuneCountInString(content.String),
			CreatedAt:        createdAt,
			Visibility:       "global",
			SensitivityClass: "role_scoped",
			TrustClass:       "user_provided",
			SourceKind:       "user_document",
			SourceLabel:      "operator provided",
		}

		if len(metaJSON) > 0 {
			var meta map[string]any
			if err := json.Unmarshal(metaJSON, &meta); err == nil {
				entry.SourceLabel = stringMeta(meta, "source_label", entry.SourceLabel)
				entry.KnowledgeClass = stringMeta(meta, "knowledge_class", entry.KnowledgeClass)
				entry.SourceKind = stringMeta(meta, "source_kind", entry.SourceKind)
				entry.Visibility = stringMeta(meta, "visibility", entry.Visibility)
				entry.SensitivityClass = stringMeta(meta, "sensitivity_class", entry.SensitivityClass)
				entry.TrustClass = stringMeta(meta, "trust_class", entry.TrustClass)
				entry.ChunkCount = intMeta(meta, "chunk_count", 0)
				entry.VectorCount = intMeta(meta, "vector_count", entry.ChunkCount)
				entry.ContentLength = intMeta(meta, "content_length", entry.ContentLength)
			}
		}

		entries = append(entries, entry)
	}

	if entries == nil {
		entries = []Entry{}
	}
	return entries, rows.Err()
}

func normalizeSourceKind(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "workspace_file":
		return "workspace_file"
	case "web_research":
		return "web_research"
	case "user_note":
		return "user_note"
	default:
		return "user_document"
	}
}

func normalizeKnowledgeClass(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case KnowledgeClassCompanyKnowledge:
		return KnowledgeClassCompanyKnowledge
	case KnowledgeClassSomaOperating:
		return KnowledgeClassSomaOperating
	default:
		return KnowledgeClassCustomerContext
	}
}

func normalizeSomaContextKind(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "output_specificity":
		return "output_specificity"
	case "operating_stance":
		return "operating_stance"
	case "identity":
		return "identity"
	case "policy":
		return "policy"
	default:
		return ""
	}
}

func normalizeOutputSpecificity(raw string) string {
	value := strings.TrimSpace(strings.ToLower(raw))
	switch value {
	case "concise", "balanced", "detailed", "executive":
		return value
	default:
		return ""
	}
}

func normalizeSourceLabel(raw string) string {
	label := strings.TrimSpace(raw)
	if label == "" {
		return "operator provided"
	}
	return label
}

func normalizeVisibility(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "private":
		return "private"
	case "team":
		return "team"
	default:
		return "global"
	}
}

func normalizeSensitivityClass(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "restricted":
		return "restricted"
	case "team_scoped":
		return "team_scoped"
	default:
		return "role_scoped"
	}
}

func normalizeTrustClass(raw string) string {
	switch strings.TrimSpace(strings.ToLower(raw)) {
	case "validated_external":
		return "validated_external"
	case "bounded_external":
		return "bounded_external"
	case "trusted_internal":
		return "trusted_internal"
	default:
		return "user_provided"
	}
}

func normalizeContentType(raw string) string {
	value := strings.TrimSpace(strings.ToLower(raw))
	if value == "" {
		return "text/markdown"
	}
	return value
}

func normalizeTags(tags []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(tags))
	for _, raw := range tags {
		tag := strings.TrimSpace(strings.ToLower(raw))
		if tag == "" {
			continue
		}
		if _, ok := seen[tag]; ok {
			continue
		}
		seen[tag] = struct{}{}
		out = append(out, tag)
	}
	return out
}

func buildEmbeddingText(knowledgeClass, title, sourceLabel, sourceKind, chunk string, chunkIndex, chunkCount int) string {
	return fmt.Sprintf(
		"[%s] title=%s source=%s source_kind=%s chunk=%d/%d\n%s",
		knowledgeClass,
		title,
		sourceLabel,
		sourceKind,
		chunkIndex,
		chunkCount,
		chunk,
	)
}

func chunkText(content string, maxRunes, overlapRunes int) []string {
	content = strings.TrimSpace(content)
	if content == "" {
		return nil
	}
	if maxRunes <= 0 {
		maxRunes = defaultChunkSize
	}
	if overlapRunes < 0 {
		overlapRunes = 0
	}

	runes := []rune(content)
	if len(runes) <= maxRunes {
		return []string{content}
	}

	var chunks []string
	for start := 0; start < len(runes); {
		end := start + maxRunes
		if end > len(runes) {
			end = len(runes)
		}

		if end < len(runes) {
			window := string(runes[start:end])
			if split := strings.LastIndex(window, "\n\n"); split >= maxRunes/2 {
				end = start + split
			} else if split := strings.LastIndex(window, "\n"); split >= maxRunes/2 {
				end = start + split
			} else if split := strings.LastIndex(window, ". "); split >= maxRunes/2 {
				end = start + split + 1
			}
		}

		chunk := strings.TrimSpace(string(runes[start:end]))
		if chunk != "" {
			chunks = append(chunks, chunk)
		}

		if end >= len(runes) {
			break
		}

		nextStart := end - overlapRunes
		if nextStart <= start {
			nextStart = end
		}
		start = nextStart
	}

	return chunks
}

func previewContent(content string, maxRunes int) string {
	content = strings.TrimSpace(content)
	if content == "" {
		return ""
	}
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\n", " ")
	content = strings.Join(strings.Fields(content), " ")
	runes := []rune(content)
	if len(runes) <= maxRunes {
		return content
	}
	return strings.TrimSpace(string(runes[:maxRunes])) + "..."
}

func stringMeta(meta map[string]any, key, fallback string) string {
	if meta == nil {
		return fallback
	}
	if value, ok := meta[key].(string); ok && strings.TrimSpace(value) != "" {
		return strings.TrimSpace(value)
	}
	return fallback
}

func intMeta(meta map[string]any, key string, fallback int) int {
	if meta == nil {
		return fallback
	}
	switch value := meta[key].(type) {
	case float64:
		return int(value)
	case int:
		return value
	default:
		return fallback
	}
}

func stringSliceMeta(meta map[string]any, key string) []string {
	if meta == nil {
		return nil
	}
	raw, ok := meta[key]
	if !ok {
		return nil
	}
	switch value := raw.(type) {
	case []any:
		out := make([]string, 0, len(value))
		for _, item := range value {
			if text, ok := item.(string); ok && strings.TrimSpace(text) != "" {
				out = append(out, strings.TrimSpace(text))
			}
		}
		return out
	case []string:
		return normalizeTags(value)
	default:
		return nil
	}
}
