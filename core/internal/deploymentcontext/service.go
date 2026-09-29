package deploymentcontext

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"unicode/utf8"

	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/memory"
)

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
	case s.Artifacts.DB == nil:
		return fmt.Errorf("deployment context store unavailable")
	case s.Memory == nil:
		return fmt.Errorf("memory service offline")
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
	rawSourceKind := strings.TrimSpace(req.SourceKind)
	sourceKind := normalizeSourceKind(req.SourceKind)
	visibility := ResolveVisibility(req.KnowledgeClass, req.Visibility)
	sensitivityClass := normalizeSensitivityClass(req.SensitivityClass)
	rawTrustClass := strings.TrimSpace(req.TrustClass)
	trustClass := normalizeTrustClass(req.TrustClass)
	agentID := strings.TrimSpace(req.AgentID)
	if agentID == "" {
		if knowledgeClass == KnowledgeClassUserPrivate || knowledgeClass == KnowledgeClassReflection {
			agentID = "admin"
		} else {
			agentID = "soma"
		}
	}
	teamID := strings.TrimSpace(req.TeamID)
	userLabel := strings.TrimSpace(req.UserLabel)
	if userLabel == "" {
		userLabel = "local-user"
	}
	tags := normalizeTags(req.Tags)
	somaContextKind := normalizeSomaContextKind(req.SomaContextKind)
	outputSpecificity := normalizeOutputSpecificity(req.OutputSpecificity)
	contentDomain := normalizeContentDomain(req.ContentDomain)
	targetGoalSets := normalizeTags(req.TargetGoalSets)
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
		tags = append(tags, "soma-operating-context")
		if outputSpecificity != "" {
			tags = append(tags, "shared-output-specificity")
		}
		tags = normalizeTags(tags)
	}
	if knowledgeClass == KnowledgeClassUserPrivate {
		if sourceLabel == "operator provided" {
			sourceLabel = "private user content"
		}
		if sourceKind == "user_document" {
			sourceKind = "user_record"
		}
		sensitivityClass = "restricted"
		if contentDomain != "" {
			tags = append(tags, contentDomain)
		}
		tags = append(tags, "user-private-context")
		tags = normalizeTags(tags)
	}
	if knowledgeClass == KnowledgeClassReflection {
		if sourceLabel == "operator provided" {
			sourceLabel = "reflection synthesis"
		}
		if rawSourceKind == "" || sourceKind == "user_document" {
			sourceKind = "synthesis_note"
		}
		sensitivityClass = "restricted"
		if rawTrustClass == "" {
			trustClass = "trusted_internal"
		}
		if contentDomain == "" {
			contentDomain = "reflection"
		}
		tags = append(tags, "reflection-synthesis-memory", sourceKind)
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
		"vector_count":       0,
		"embedding_status":   EmbeddingStatusPending,
		"lexical_indexed":    true,
		"content_length":     utf8.RuneCountInString(content),
		"loaded_by":          userLabel,
		"team_id":            teamID,
		"agent_id":           agentID,
		"soma_context_kind":  somaContextKind,
		"output_specificity": outputSpecificity,
		"content_domain":     contentDomain,
		"target_goal_sets":   targetGoalSets,
	}
	if knowledgeClass == KnowledgeClassReflection {
		metadataMap["reflection_kind"] = sourceKind
	}
	for key, value := range req.ExtraMetadata {
		if _, exists := metadataMap[key]; exists {
			continue
		}
		metadataMap[key] = value
	}
	metadataJSON, _ := json.Marshal(metadataMap)

	chunkMeta := make([]map[string]any, len(chunks))
	pending := make([]pendingChunk, len(chunks))
	for idx, chunk := range chunks {
		// The vector type is the knowledge class itself so recall can keep
		// customer-provided context separate from approved company knowledge.
		vectorMeta := map[string]any{
			"type":               knowledgeClass,
			"source":             "governed_context_store",
			"knowledge_store":    "governed_context_store",
			"knowledge_class":    knowledgeClass,
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
			"content_domain":     contentDomain,
			"target_goal_sets":   targetGoalSets,
			"chunk_index":        idx,
			"chunk_count":        len(chunks),
			"loaded_by":          userLabel,
			"embedding_status":   EmbeddingStatusPending,
		}
		if knowledgeClass == KnowledgeClassReflection {
			vectorMeta["reflection_kind"] = sourceKind
		}
		for key, value := range req.ExtraMetadata {
			if _, exists := vectorMeta[key]; !exists {
				vectorMeta[key] = value
			}
		}
		chunkMeta[idx] = vectorMeta
		pending[idx].text = buildEmbeddingText(knowledgeClass, title, sourceLabel, sourceKind, chunk, idx+1, len(chunks))
	}

	// One transaction: the artifact and its chunk rows exist together or not
	// at all. Chunks are keyword-searchable as soon as it commits.
	artifactID, createdAt, chunkIDs, err := s.insertAtomic(ctx, agentID, title, contentType, content, metadataJSON, chunks, chunkMeta)
	if err != nil {
		return nil, &SaveError{Code: SaveFailedCode, Err: err}
	}
	for idx := range pending {
		pending[idx].id = chunkIDs[idx]
	}
	embedded := s.embedChunks(ctx, pending)
	status, vectorCount := EmbeddingStatusPending, 0
	if statuses, counts, err := s.refreshArtifacts(ctx, []string{artifactID}); err != nil {
		log.Printf("deployment context: refresh embedding status for %s: %v", artifactID, err)
	} else if value, ok := statuses[artifactID]; ok {
		status, vectorCount = value, counts[artifactID]
	}
	if embedded > 0 {
		if _, err := s.BackfillEmbeddings(ctx, opportunisticBackfillLimit); err != nil {
			log.Printf("deployment context: opportunistic backfill: %v", err)
		}
	}
	modes, message := retrievalModes(status)

	return &IngestResult{
		ArtifactID:       artifactID,
		KnowledgeClass:   knowledgeClass,
		Title:            title,
		SourceLabel:      sourceLabel,
		SourceKind:       sourceKind,
		Visibility:       visibility,
		SensitivityClass: sensitivityClass,
		TrustClass:       trustClass,
		ChunkCount:       len(chunks),
		VectorCount:      vectorCount,
		ContentPreview:   previewContent(content, 220),
		ContentLength:    utf8.RuneCountInString(content),
		ContentDomain:    contentDomain,
		TargetGoalSets:   targetGoalSets,
		CreatedAt:        createdAt,
		EmbeddingStatus:  status,
		RetrievalModes:   modes,
		StatusMessage:    message,
	}, nil
}
