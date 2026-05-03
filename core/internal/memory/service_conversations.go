package memory

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"
)

// ConversationSummary represents a stored conversation summary with extracted metadata.
type ConversationSummary struct {
	ID               string         `json:"id"`
	AgentID          string         `json:"agent_id"`
	Summary          string         `json:"summary"`
	KeyTopics        []string       `json:"key_topics"`
	UserPreferences  map[string]any `json:"user_preferences"`
	PersonalityNotes string         `json:"personality_notes"`
	DataReferences   []any          `json:"data_references"`
	MessageCount     int            `json:"message_count"`
	CreatedAt        time.Time      `json:"created_at"`
	Score            float64        `json:"score,omitempty"` // cosine similarity (only on recall)
}

// ConversationSummaryInput captures the durable summary payload plus the
// recall-scope metadata that will be stored alongside its vector embedding.
type ConversationSummaryInput struct {
	AgentID          string
	TenantID         string
	TeamID           string
	RunID            string
	Visibility       string
	Summary          string
	KeyTopics        []string
	UserPreferences  map[string]any
	PersonalityNotes string
	DataReferences   []any
	MessageCount     int
}

// EmbedFunc is a callback for embedding text (decouples memory from cognitive package).
type EmbedFunc func(ctx context.Context, text string, model string) ([]float64, error)

// StoreConversationSummary persists a conversation summary and embeds it into pgvector.
// The embedding is stored in context_vectors with metadata linking back to this summary.
// cog may be nil; in that case, only the RDBMS record is created (no vector).
func (s *Service) StoreConversationSummary(ctx context.Context, cog EmbedFunc, input ConversationSummaryInput) (string, error) {
	if strings.TrimSpace(input.AgentID) == "" {
		input.AgentID = "admin"
	}
	if strings.TrimSpace(input.TenantID) == "" {
		input.TenantID = "default"
	}
	if strings.TrimSpace(input.Visibility) == "" {
		input.Visibility = "private"
	}
	if input.UserPreferences == nil {
		input.UserPreferences = map[string]any{}
	}
	if input.DataReferences == nil {
		input.DataReferences = []any{}
	}
	prefsJSON, _ := json.Marshal(input.UserPreferences)
	refsJSON, _ := json.Marshal(input.DataReferences)

	var summaryID string
	err := s.db.QueryRowContext(ctx, `
		INSERT INTO conversation_summaries (agent_id, summary, key_topics, user_preferences, personality_notes, data_references, message_count)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`, input.AgentID, input.Summary, pqTextArray(input.KeyTopics), prefsJSON, input.PersonalityNotes, refsJSON, input.MessageCount).Scan(&summaryID)
	if err != nil {
		return "", fmt.Errorf("store conversation summary: %w", err)
	}

	// Embed and store vector (best-effort)
	if cog != nil {
		embText := fmt.Sprintf("[conversation] %s", input.Summary)
		vec, err := cog(ctx, embText, "")
		if err == nil {
			meta := map[string]any{
				"type":       "conversation",
				"agent_id":   input.AgentID,
				"summary_id": summaryID,
				"tenant_id":  input.TenantID,
				"team_id":    strings.TrimSpace(input.TeamID),
				"run_id":     strings.TrimSpace(input.RunID),
				"visibility": strings.ToLower(strings.TrimSpace(input.Visibility)),
				"source":     "conversation_summary",
			}
			if storeErr := s.StoreVector(ctx, embText, vec, meta); storeErr != nil {
				log.Printf("conversation summary: vector store failed (non-fatal): %v", storeErr)
			}
		} else {
			log.Printf("conversation summary: embedding failed (non-fatal): %v", err)
		}
	}

	return summaryID, nil
}

// RecallConversations finds the most relevant past conversation summaries for a query.
// Uses semantic search on context_vectors filtered by type="conversation", then JOINs
// with conversation_summaries for structured data.
func (s *Service) RecallConversations(ctx context.Context, queryVec []float64, agentID string, limit int) ([]ConversationSummary, error) {
	if limit <= 0 {
		limit = 3
	}

	vecStr := formatVector(queryVec)

	rows, err := s.db.QueryContext(ctx, `
		SELECT cs.id, cs.agent_id, cs.summary, cs.key_topics, cs.user_preferences,
		       cs.personality_notes, cs.data_references, cs.message_count, cs.created_at,
		       1 - (cv.embedding <=> $1::vector) AS score
		FROM context_vectors cv
		JOIN conversation_summaries cs ON cs.id::text = cv.metadata->>'summary_id'
		WHERE cv.metadata->>'type' = 'conversation'
		  AND cv.embedding IS NOT NULL
		  AND ($2 = '' OR cv.metadata->>'agent_id' = $2)
		ORDER BY cv.embedding <=> $1::vector
		LIMIT $3
	`, vecStr, agentID, limit)
	if err != nil {
		return nil, fmt.Errorf("recall conversations: %w", err)
	}
	defer rows.Close()

	var results []ConversationSummary
	for rows.Next() {
		var cs ConversationSummary
		var topicsArr []string
		var prefsJSON, refsJSON []byte

		if err := rows.Scan(&cs.ID, &cs.AgentID, &cs.Summary, pqScanArray(&topicsArr),
			&prefsJSON, &cs.PersonalityNotes, &refsJSON, &cs.MessageCount, &cs.CreatedAt, &cs.Score); err != nil {
			log.Printf("recall conversations: scan error: %v", err)
			continue
		}
		cs.KeyTopics = topicsArr
		if len(prefsJSON) > 0 {
			_ = json.Unmarshal(prefsJSON, &cs.UserPreferences)
		}
		if len(refsJSON) > 0 {
			_ = json.Unmarshal(refsJSON, &cs.DataReferences)
		}
		results = append(results, cs)
	}
	return results, nil
}
