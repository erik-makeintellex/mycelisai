package swarm

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/deploymentcontext"
)

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

	contentType, err := validateArtifactContent(artType, content)
	if err != nil {
		return "", err
	}
	artifactID, err := r.insertArtifact(ctx, artType, title, contentType, content, artifactMetadataJSON(args["metadata"]))
	if err != nil {
		log.Printf("store_artifact: %v", err)
		return fmt.Sprintf("Failed to store artifact: %v", err), nil
	}
	if r.exchange != nil {
		publishArtifactToExchange(ctx, r.exchange, artifactID, artType, title)
	}
	return mustJSON(artifactResultPayload(artifactID, artType, title, contentType, content)), nil
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

	scope := resolveMemoryScope(ctx, args)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO agent_memories (category, content, context, tenant_id, team_id, agent_id, run_id, visibility, created_at)
		VALUES ($1, $2, $3, $4, NULLIF($5,''), $6, NULLIF($7,''), $8, NOW())
	`, category, content, memContext, scope.TenantID, scope.TeamID, scope.AgentID, scope.RunID, scope.Visibility)
	if err != nil {
		log.Printf("remember: RDBMS insert failed: %v", err)
		return fmt.Sprintf("Failed to store memory: %v", err), nil
	}

	storeMemoryVector(ctx, r.brain, r.mem, category, content, memContext, scope)
	return fmt.Sprintf("Remembered [%s]: %s", category, content), nil
}

func (r *InternalToolRegistry) handleLoadDeploymentContext(ctx context.Context, args map[string]any) (string, error) {
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

	scope := resolveMemoryScope(ctx, args)
	result, err := svc.Ingest(ctx, deploymentcontext.IngestRequest{
		KnowledgeClass:   stringValue(args["knowledge_class"]),
		Title:            title,
		Content:          content,
		ContentType:      stringValue(args["content_type"]),
		SourceLabel:      stringValue(args["source_label"]),
		SourceKind:       stringValue(args["source_kind"]),
		Visibility:       stringValue(args["visibility"]),
		SensitivityClass: stringValue(args["sensitivity_class"]),
		TrustClass:       stringValue(args["trust_class"]),
		Tags:             stringSlice(args["tags"]),
		AgentID:          scope.AgentID,
		TeamID:           scope.TeamID,
		UserLabel:        "soma",
	})
	if err != nil {
		return "", err
	}

	return mustJSON(map[string]any{
		"message":         fmt.Sprintf("Knowledge entry '%s' loaded into the governed context store.", result.Title),
		"artifact":        map[string]any{"id": result.ArtifactID, "type": "document", "title": result.Title, "content_type": "text/markdown"},
		"knowledge_class": result.KnowledgeClass,
		"chunk_count":     result.ChunkCount,
		"vector_count":    result.VectorCount,
		"source_label":    result.SourceLabel,
		"source_kind":     result.SourceKind,
		"visibility":      result.Visibility,
		"trust_class":     result.TrustClass,
		"context_kind":    "governed_knowledge",
		"description":     "Stored in the governed context store for later Soma recall, separate from Soma memory.",
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

	scope := resolveMemoryScope(ctx, args)
	results := recallStructuredMemories(ctx, r.db, query, category, limit, scope)
	results = append(results, recallVectorMemories(ctx, r.brain, r.mem, query, limit, scope)...)
	if results == nil {
		results = []memoryResult{}
	}
	return mustJSON(results), nil
}

func (r *InternalToolRegistry) handleTempMemoryWrite(ctx context.Context, args map[string]any) (string, error) {
	if r.mem == nil {
		return "", fmt.Errorf("memory service offline — temp channels unavailable")
	}
	channel := stringValue(args["channel"])
	content := stringValue(args["content"])
	owner := stringValue(args["owner_agent_id"])
	metadata, _ := args["metadata"].(map[string]any)
	ttl := 0
	if ttlRaw, ok := args["ttl_minutes"].(float64); ok {
		ttl = int(ttlRaw)
	}
	id, err := r.mem.PutTempMemory(ctx, "default", channel, owner, content, metadata, ttl)
	if err != nil {
		return "", err
	}
	return mustJSON(map[string]any{"message": fmt.Sprintf("temp memory stored in channel %q", channel), "id": id, "channel": channel}), nil
}

func (r *InternalToolRegistry) handleTempMemoryRead(ctx context.Context, args map[string]any) (string, error) {
	if r.mem == nil {
		return "", fmt.Errorf("memory service offline — temp channels unavailable")
	}
	channel := stringValue(args["channel"])
	limit := 10
	if l, ok := args["limit"].(float64); ok && l > 0 {
		limit = int(l)
	}
	entries, err := r.mem.GetTempMemory(ctx, "default", channel, limit)
	if err != nil {
		return "", err
	}
	return mustJSON(entries), nil
}

func (r *InternalToolRegistry) handleTempMemoryClear(ctx context.Context, args map[string]any) (string, error) {
	if r.mem == nil {
		return "", fmt.Errorf("memory service offline — temp channels unavailable")
	}
	channel := stringValue(args["channel"])
	deleted, err := r.mem.ClearTempMemory(ctx, "default", channel)
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
	return r.summarizeAndStore(ctx, resolveMemoryScope(ctx, args), messagesText, 0)
}

// AutoSummarize compresses a chat history window into a temporary continuity checkpoint.
func (r *InternalToolRegistry) AutoSummarize(ctx context.Context, agentID, teamID string, history []cognitive.ChatMessage) {
	if r.brain == nil || r.mem == nil {
		return
	}
	var sb strings.Builder
	for _, m := range history {
		sb.WriteString(fmt.Sprintf("[%s]: %s\n", m.Role, m.Content))
	}
	scope := memoryScope{TenantID: "default", TeamID: strings.TrimSpace(teamID), AgentID: strings.TrimSpace(agentID), Visibility: "team"}
	checkpointID, err := r.summarizeAndCheckpoint(ctx, scope, sb.String(), len(history))
	if err != nil {
		log.Printf("AutoSummarize [%s]: failed: %v", agentID, err)
		return
	}
	log.Printf("AutoSummarize [%s]: stored continuity checkpoint %s (%d messages)", agentID, checkpointID, len(history))
}
