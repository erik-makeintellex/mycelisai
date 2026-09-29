package swarm

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/internal/memory"
)

// recallStructuredMemories reads org-wide (visibility global) agent_memories
// rows only (MEM-LANES). The table has no owner column, so its private and
// team rows cannot be attributed to a user: a user's own facts are recalled
// from their owner-stamped vector rows instead (recallVectorMemories).
func recallStructuredMemories(ctx context.Context, db *sql.DB, query, category string, limit int, scope memoryScope) []memoryResult {
	if db == nil {
		return nil
	}
	clauses := []string{"tenant_id = $1", "content ILIKE '%' || $2 || '%'", "visibility = 'global'"}
	queryArgs := []any{scope.TenantID, query}
	nextArg := 3
	if category != "" {
		clauses = append(clauses, fmt.Sprintf("category = $%d", nextArg))
		queryArgs = append(queryArgs, category)
		nextArg++
	}

	rdbmsQuery := `
		SELECT category, content, COALESCE(context, ''), created_at
		FROM agent_memories
		WHERE ` + strings.Join(clauses, " AND ") + `
		ORDER BY created_at DESC LIMIT $` + fmt.Sprintf("%d", nextArg)
	queryArgs = append(queryArgs, limit)
	rows, err := db.QueryContext(ctx, rdbmsQuery, queryArgs...)
	if err != nil {
		return nil
	}
	defer rows.Close()

	var results []memoryResult
	for rows.Next() {
		var m memoryResult
		var createdAt time.Time
		if err := rows.Scan(&m.Category, &m.Content, &m.Context, &createdAt); err == nil {
			m.CreatedAt = createdAt.Format(time.RFC3339)
			m.Source = "rdbms"
			results = append(results, m)
		}
	}
	return results
}

// recallVectorMemories admits governed rows only through reader (SRU); a nil
// reader would be unscoped, so it recalls nothing.
func recallVectorMemories(ctx context.Context, brain *cognitive.Router, mem *memory.Service, query string, limit int, scope memoryScope, reader *memory.GovernedReader) []memoryResult {
	if mem == nil || reader == nil {
		return nil
	}
	// Semantic when an embedding engine works; PostgreSQL keyword ranking otherwise.
	vecResults, _, err := mem.RecallGoverned(ctx, brain, query, memory.SemanticSearchOptions{
		Limit:               limit,
		TenantID:            scope.TenantID,
		TeamID:              scope.TeamID,
		AgentID:             scope.AgentID,
		RunID:               scope.RunID,
		AllowGlobal:         true,
		AllowLegacyUnscoped: scope.TeamID == "" && scope.AgentID == "",
		Reader:              reader,
	})
	if err != nil {
		return nil
	}
	results := make([]memoryResult, 0, len(vecResults))
	for _, vr := range vecResults {
		source := "vector"
		if vr.RetrievalMode == memory.RecallModeKeyword {
			source = "keyword"
		}
		results = append(results, memoryResult{Content: vr.Content, Score: vr.Score, Source: source})
	}
	return results
}

func (r *InternalToolRegistry) summarizeConversation(ctx context.Context, messagesText string) (parsedConversationSummary, error) {
	req := cognitive.InferRequest{
		Profile: "chat",
		Messages: []cognitive.ChatMessage{
			{Role: "system", Content: "You are a conversation summarizer. Output only valid JSON."},
			{Role: "user", Content: `Summarize this conversation in 2-3 sentences. Then extract structured metadata.

Respond with ONLY this JSON (no markdown fences):
{
  "summary": "2-3 sentence summary of the conversation",
  "key_topics": ["topic1", "topic2"],
  "user_preferences": {"preference_key": "preference_value"},
  "personality_notes": "how the user wants to be addressed or treated",
  "data_references": [{"type": "file/url/artifact", "ref": "the reference"}]
}

Conversation:
` + messagesText},
		},
	}
	resp, err := r.brain.InferWithContract(ctx, req)
	if err != nil {
		return parsedConversationSummary{}, fmt.Errorf("LLM compression failed: %w", err)
	}
	return normalizeConversationSummary(resp.Text), nil
}

func normalizeConversationSummary(text string) parsedConversationSummary {
	text = strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(strings.TrimSpace(text), "```json"), "```"))
	var parsed parsedConversationSummary
	if err := json.Unmarshal([]byte(strings.TrimSpace(text)), &parsed); err != nil {
		log.Printf("summarizeAndStore: LLM output not valid JSON, using raw text: %v", err)
		parsed.Summary = text
	}
	if parsed.Summary == "" {
		parsed.Summary = text
	}
	if parsed.KeyTopics == nil {
		parsed.KeyTopics = []string{}
	}
	if parsed.UserPreferences == nil {
		parsed.UserPreferences = map[string]any{}
	}
	if parsed.DataReferences == nil {
		parsed.DataReferences = []any{}
	}
	return parsed
}

func (r *InternalToolRegistry) summarizeAndStore(ctx context.Context, scope memoryScope, owner, messagesText string, msgCount int) (string, error) {
	parsed, err := r.summarizeConversation(ctx, messagesText)
	if err != nil {
		return "", err
	}
	embedFunc := memory.EmbedFunc(r.brain.Embed)
	return r.mem.StoreConversationSummary(ctx, embedFunc, memory.ConversationSummaryInput{
		AgentID: scope.AgentID, TenantID: scope.TenantID, TeamID: scope.TeamID, RunID: scope.RunID, Visibility: scope.Visibility, OwnerUserID: owner,
		Summary: parsed.Summary, KeyTopics: parsed.KeyTopics, UserPreferences: parsed.UserPreferences, PersonalityNotes: parsed.PersonalityNotes, DataReferences: parsed.DataReferences, MessageCount: msgCount,
	})
}

func (r *InternalToolRegistry) summarizeAndCheckpoint(ctx context.Context, scope memoryScope, owner, messagesText string, msgCount int) (string, error) {
	parsed, err := r.summarizeConversation(ctx, messagesText)
	if err != nil {
		return "", err
	}
	content := strings.TrimSpace(parsed.Summary)
	if len(parsed.KeyTopics) > 0 {
		content += "\nTopics: " + strings.Join(parsed.KeyTopics, ", ")
	}
	if strings.TrimSpace(parsed.PersonalityNotes) != "" {
		content += "\nInteraction notes: " + strings.TrimSpace(parsed.PersonalityNotes)
	}
	metadata := map[string]any{"summary_type": "planning_continuity", "tenant_id": scope.TenantID, "team_id": scope.TeamID, "agent_id": scope.AgentID, "run_id": scope.RunID, "visibility": scope.Visibility, "message_count": msgCount, "key_topics": parsed.KeyTopics, "user_preferences": parsed.UserPreferences, "personality_notes": parsed.PersonalityNotes, "data_references": parsed.DataReferences, "promotion_boundary": "temporary_only"}
	channelKey := fmt.Sprintf("lead.%s", scope.AgentID)
	if strings.TrimSpace(scope.TeamID) != "" {
		channelKey = fmt.Sprintf("team.%s.planning", scope.TeamID)
	}
	return r.mem.PutTempMemory(ctx, scope.TenantID, channelKey, scope.AgentID, content, metadata, 240, owner)
}
