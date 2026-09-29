package swarm

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/deploymentcontext"
	"github.com/mycelis/core/internal/memory"
)

// noUserSaveOwner is the owner id of a Soma-tool save with no user and no
// Core-set user label. It is no user's id, so the save is never readable
// through the legacy loaded_by label rule by an identity named "soma".
const noUserSaveOwner = "system:soma"

// toolSaveOwner is whom a Soma-tool deployment-context save belongs to
// (MEM-LANES): the turn's verified user (owner id plus their label), else the
// Core-set invocation label (the confirmed path), else nobody.
func toolSaveOwner(ctx context.Context) (string, map[string]any, error) {
	access := recallAccessFromContext(ctx)
	owner, err := access.ownerUserID()
	if err != nil {
		return "", nil, err
	}
	if owner != "" {
		label := strings.TrimSpace(access.Reader.Label)
		if label == "" {
			label = owner
		}
		return label, map[string]any{memory.OwnerUserIDKey: owner}, nil
	}
	if inv, ok := ToolInvocationContextFromContext(ctx); ok && strings.TrimSpace(inv.UserLabel) != "" {
		return strings.TrimSpace(inv.UserLabel), nil, nil
	}
	return "soma", map[string]any{memory.OwnerUserIDKey: noUserSaveOwner}, nil
}

func (r *InternalToolRegistry) handleStoreArtifact(ctx context.Context, args map[string]any) (string, error) {
	artType := stringValue(args["type"])
	title := stringValue(args["title"])
	content := stringValue(args["content"])
	if artType == "" || title == "" || content == "" {
		return "", fmt.Errorf("store_artifact requires 'type', 'title', and 'content'")
	}
	if r.db == nil {
		return "", fmt.Errorf("database not available — cannot store artifact")
	}

	metadata := args["metadata"]
	if packageKindFromMetadata(metadata) == "project_package" && (artType == "file" || artType == "data") {
		artType = "project_package"
	}
	contentType, err := validateArtifactContent(artType, content)
	if err != nil {
		return "", err
	}
	artifactID, err := r.insertArtifact(ctx, artType, title, contentType, content, artifactMetadataJSON(metadata))
	if err != nil {
		return "", fmt.Errorf("store_artifact failed to persist the artifact: %w", err)
	}
	if r.exchange != nil {
		publishArtifactToExchange(ctx, r.exchange, artifactID, artType, title)
	}
	return mustJSON(artifactResultPayload(artifactID, artType, title, contentType, content, metadata)), nil
}

func (r *InternalToolRegistry) handleRemember(ctx context.Context, args map[string]any) (string, error) {
	category := stringValue(args["category"])
	content := stringValue(args["content"])
	memContext := stringValue(args["context"])
	if category == "" || content == "" {
		return "", fmt.Errorf("remember requires 'category' and 'content'")
	}
	if r.db == nil {
		return "", fmt.Errorf("database not available — cannot persist memory")
	}
	// MEM-LANES: the fact belongs to the turn's verified user; an unverified
	// user saves nothing.
	access := recallAccessFromContext(ctx)
	owner, err := access.ownerUserID()
	if err != nil {
		return "", fmt.Errorf("remember refused: %w. Nothing was saved.", err)
	}

	scope := resolveMemoryScope(ctx, args)
	var note string // MEM-LANES-2: the model's visibility is a request only
	scope.Visibility, note = ownerLaneVisibility(access, scope)
	_, err = r.db.ExecContext(ctx, `
		INSERT INTO agent_memories (category, content, context, tenant_id, team_id, agent_id, run_id, visibility, created_at)
		VALUES ($1, $2, $3, $4, NULLIF($5,''), $6, NULLIF($7,''), $8, NOW())
	`, category, content, memContext, scope.TenantID, scope.TeamID, scope.AgentID, scope.RunID, scope.Visibility)
	if err != nil {
		return "", fmt.Errorf("remember failed to persist the memory: %w", err)
	}

	// The owner-attributed vector row is the only copy recall returns for a
	// user's own fact (the agent_memories row has no owner column).
	if err := storeMemoryVector(ctx, r.brain, r.mem, category, content, memContext, scope, owner); err != nil {
		if owner != "" {
			return "", fmt.Errorf("remember stored the fact but could not index it for your recall: %w", err)
		}
		log.Printf("remember: vector index failed (non-fatal for a no-user turn): %v", err)
	}
	return withVisibilityNote(fmt.Sprintf("Remembered [%s]: %s", category, content), note), nil
}

func (r *InternalToolRegistry) handleLoadDeploymentContext(ctx context.Context, args map[string]any) (string, error) {
	if err := requireConfirmedOrgWideWrite(ctx, stringValue(args["knowledge_class"])); err != nil {
		return "", err
	}
	var artifactService *artifacts.Service
	if r.db != nil {
		artifactService = &artifacts.Service{DB: r.db}
	}
	svc := deploymentcontext.NewService(artifactService, r.mem, r.brain)
	if err := svc.Ready(); err != nil {
		return "", err
	}

	title := stringValue(args["title"])
	content := stringValue(args["content"])
	if title == "" || content == "" {
		return "", fmt.Errorf("load_deployment_context requires 'title' and 'content'")
	}

	label, ownerMeta, err := toolSaveOwner(ctx)
	if err != nil {
		return "", fmt.Errorf("load_deployment_context refused: %w. Nothing was saved.", err)
	}
	scope := resolveMemoryScope(ctx, args)
	result, err := svc.Ingest(ctx, deploymentcontext.IngestRequest{
		KnowledgeClass:    stringValue(args["knowledge_class"]),
		Title:             title,
		Content:           content,
		ContentType:       stringValue(args["content_type"]),
		SourceLabel:       stringValue(args["source_label"]),
		SourceKind:        stringValue(args["source_kind"]),
		Visibility:        stringValue(args["visibility"]),
		SensitivityClass:  stringValue(args["sensitivity_class"]),
		TrustClass:        stringValue(args["trust_class"]),
		Tags:              stringSlice(args["tags"]),
		AgentID:           scope.AgentID,
		TeamID:            scope.TeamID,
		UserLabel:         label,
		SomaContextKind:   stringValue(args["soma_context_kind"]),
		OutputSpecificity: stringValue(args["output_specificity"]),
		ContentDomain:     stringValue(args["content_domain"]),
		TargetGoalSets:    stringSlice(args["target_goal_sets"]),
		ExtraMetadata:     ownerMeta,
	})
	if err != nil {
		return "", err
	}

	return mustJSON(map[string]any{
		"message":          fmt.Sprintf("Knowledge entry '%s' loaded into the governed context store.", result.Title),
		"artifact":         map[string]any{"id": result.ArtifactID, "type": "document", "title": result.Title, "content_type": "text/markdown"},
		"knowledge_class":  result.KnowledgeClass,
		"chunk_count":      result.ChunkCount,
		"vector_count":     result.VectorCount,
		"source_label":     result.SourceLabel,
		"source_kind":      result.SourceKind,
		"visibility":       result.Visibility,
		"trust_class":      result.TrustClass,
		"content_domain":   result.ContentDomain,
		"target_goal_sets": result.TargetGoalSets,
		"context_kind":     "governed_knowledge",
		"embedding_status": result.EmbeddingStatus,
		"retrieval_modes":  result.RetrievalModes,
		"description":      result.StatusMessage,
	}), nil
}

func (r *InternalToolRegistry) handlePromoteDeploymentContext(ctx context.Context, args map[string]any) (string, error) {
	if err := requireConfirmedOrgWideWrite(ctx, "company_knowledge"); err != nil {
		return "", err
	}
	var artifactService *artifacts.Service
	if r.db != nil {
		artifactService = &artifacts.Service{DB: r.db}
	}
	svc := deploymentcontext.NewService(artifactService, r.mem, r.brain)
	if err := svc.Ready(); err != nil {
		return "", err
	}

	sourceArtifactID := stringValue(args["source_artifact_id"])
	if sourceArtifactID == "" {
		return "", fmt.Errorf("promote_deployment_context requires 'source_artifact_id'")
	}
	if err := r.requirePromotableSource(ctx, sourceArtifactID); err != nil {
		return "", err
	}

	label, _, err := toolSaveOwner(ctx)
	if err != nil {
		return "", fmt.Errorf("promote_deployment_context refused: %w. Nothing was promoted.", err)
	}
	scope := resolveMemoryScope(ctx, args)
	result, err := svc.Promote(ctx, deploymentcontext.PromoteRequest{
		SourceArtifactID: sourceArtifactID,
		Title:            stringValue(args["title"]),
		Content:          stringValue(args["content"]),
		ContentType:      stringValue(args["content_type"]),
		SourceLabel:      stringValue(args["source_label"]),
		SourceKind:       stringValue(args["source_kind"]),
		Visibility:       stringValue(args["visibility"]),
		SensitivityClass: stringValue(args["sensitivity_class"]),
		TrustClass:       stringValue(args["trust_class"]),
		Tags:             stringSlice(args["tags"]),
		AgentID:          scope.AgentID,
		TeamID:           scope.TeamID,
		UserLabel:        label,
	})
	if err != nil {
		return "", err
	}

	return mustJSON(map[string]any{
		"message":            fmt.Sprintf("Knowledge entry '%s' promoted into approved company knowledge.", result.Title),
		"artifact":           map[string]any{"id": result.ArtifactID, "type": "document", "title": result.Title, "content_type": "text/markdown"},
		"knowledge_class":    result.KnowledgeClass,
		"chunk_count":        result.ChunkCount,
		"vector_count":       result.VectorCount,
		"source_artifact_id": sourceArtifactID,
		"source_kind":        result.SourceKind,
		"visibility":         result.Visibility,
		"trust_class":        result.TrustClass,
		"context_kind":       "governed_knowledge",
		"description":        "Stored as approved company knowledge with lineage back to the original customer context entry.",
	}), nil
}

func (r *InternalToolRegistry) handleRecall(ctx context.Context, args map[string]any) (string, error) {
	query := stringValue(args["query"])
	if query == "" {
		return "", fmt.Errorf("recall requires 'query'")
	}
	category := stringValue(args["category"])
	limit := 5
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}

	reader, err := recallAccessFromContext(ctx).governedReader()
	if err != nil {
		return "", fmt.Errorf("recall unavailable: %w", err)
	}
	scope := resolveMemoryScope(ctx, args)
	results := recallStructuredMemories(ctx, r.db, query, category, limit, scope)
	results = append(results, recallVectorMemories(ctx, r.brain, r.mem, query, limit, scope, reader)...)
	if results == nil {
		results = []memoryResult{}
	}
	return mustJSON(results), nil
}

func (r *InternalToolRegistry) handleTempMemoryWrite(ctx context.Context, args map[string]any) (string, error) {
	if r.mem == nil {
		return "", fmt.Errorf("memory service offline — temp channels unavailable")
	}
	ownerUserID, err := recallAccessFromContext(ctx).ownerUserID()
	if err != nil {
		return "", fmt.Errorf("temp_memory_write refused: %w. Nothing was saved.", err)
	}
	channel := stringValue(args["channel"])
	if ownerUserID == "" && memory.SharedTempChannel(channel) {
		return "", fmt.Errorf("temp_memory_write refused: %w. Nothing was saved.", errSharedChannelNoUser)
	}
	content := stringValue(args["content"])
	owner := stringValue(args["owner_agent_id"])
	metadata, _ := args["metadata"].(map[string]any)
	ttl := 0
	if ttlRaw, ok := args["ttl_minutes"].(float64); ok {
		ttl = int(ttlRaw)
	}
	id, err := r.mem.PutTempMemory(ctx, "default", channel, owner, content, metadata, ttl, ownerUserID)
	if err != nil {
		return "", err
	}
	return mustJSON(map[string]any{"message": fmt.Sprintf("temp memory stored in channel %q", channel), "id": id, "channel": channel}), nil
}

func (r *InternalToolRegistry) handleTempMemoryRead(ctx context.Context, args map[string]any) (string, error) {
	if r.mem == nil {
		return "", fmt.Errorf("memory service offline — temp channels unavailable")
	}
	reader, err := recallAccessFromContext(ctx).laneReader()
	if err != nil {
		return "", fmt.Errorf("temp_memory_read unavailable: %w", err)
	}
	channel := stringValue(args["channel"])
	limit := 10
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}
	entries, err := r.mem.GetTempMemory(ctx, "default", channel, limit, reader)
	if err != nil {
		return "", err
	}
	return mustJSON(entries), nil
}

func (r *InternalToolRegistry) handleTempMemoryClear(ctx context.Context, args map[string]any) (string, error) {
	if r.mem == nil {
		return "", fmt.Errorf("memory service offline — temp channels unavailable")
	}
	reader, err := recallAccessFromContext(ctx).laneReader()
	if err != nil {
		return "", fmt.Errorf("temp_memory_clear refused: %w. Nothing was cleared.", err)
	}
	channel := stringValue(args["channel"])
	deleted, err := r.mem.ClearTempMemory(ctx, "default", channel, reader)
	if err != nil {
		return "", err
	}
	return mustJSON(map[string]any{"message": fmt.Sprintf("temp memory channel %q cleared", channel), "deleted": deleted, "channel": channel}), nil
}

func (r *InternalToolRegistry) handleSummarizeConversation(ctx context.Context, args map[string]any) (string, error) {
	messagesText := stringValue(args["messages"])
	if messagesText == "" {
		return "", fmt.Errorf("summarize_conversation requires 'messages'")
	}
	if r.brain == nil {
		return "", fmt.Errorf("cognitive engine offline — cannot summarize")
	}
	if r.mem == nil {
		return "", fmt.Errorf("memory service offline — cannot store summary")
	}
	access := recallAccessFromContext(ctx)
	owner, err := access.ownerUserID()
	if err != nil {
		return "", fmt.Errorf("summarize_conversation refused: %w. Nothing was saved.", err)
	}
	scope := resolveMemoryScope(ctx, args)
	var note string // MEM-LANES-2: the model's visibility is a request only
	scope.Visibility, note = ownerLaneVisibility(access, scope)
	id, err := r.summarizeAndStore(ctx, scope, owner, messagesText, 0)
	if err != nil {
		return "", err
	}
	return withVisibilityNote(id, note), nil
}

// AutoSummarize compresses a chat history window into a temporary continuity
// checkpoint owned by the turn's user (access; MEM-LANES). A turn whose user
// could not be verified writes nothing.
func (r *InternalToolRegistry) AutoSummarize(ctx context.Context, access RecallAccess, agentID, teamID string, history []cognitive.ChatMessage) {
	if r.brain == nil || r.mem == nil {
		return
	}
	owner, err := access.ownerUserID()
	if err != nil {
		log.Printf("AutoSummarize [%s]: skipped: %v", agentID, err)
		return
	}
	var sb strings.Builder
	for _, m := range history {
		sb.WriteString(fmt.Sprintf("[%s]: %s\n", m.Role, m.Content))
	}
	scope := memoryScope{TenantID: "default", TeamID: strings.TrimSpace(teamID), AgentID: strings.TrimSpace(agentID), Visibility: "team"}
	if owner != "" {
		scope.Visibility = "private"
	}
	checkpointID, err := r.summarizeAndCheckpoint(ctx, scope, owner, sb.String(), len(history))
	if err != nil {
		log.Printf("AutoSummarize [%s]: failed: %v", agentID, err)
		return
	}
	log.Printf("AutoSummarize [%s]: stored continuity checkpoint %s (%d messages)", agentID, checkpointID, len(history))
}
