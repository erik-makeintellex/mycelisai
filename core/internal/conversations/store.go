// Package conversations provides the V7 persistent conversation turn store.
// DB-first rule: LogTurn() persists to conversation_turns before returning.
// Follows the same pattern as internal/events.Store (nil-safe, goroutine-callable).
package conversations

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

// ConversationTurn is the DB row representation returned by query methods.
type ConversationTurn struct {
	ID             string                 `json:"id"`
	RunID          string                 `json:"run_id,omitempty"`
	SessionID      string                 `json:"session_id"`
	TenantID       string                 `json:"tenant_id"`
	AgentID        string                 `json:"agent_id"`
	TeamID         string                 `json:"team_id,omitempty"`
	TurnIndex      int                    `json:"turn_index"`
	Role           string                 `json:"role"`
	Content        string                 `json:"content"`
	ProviderID     string                 `json:"provider_id,omitempty"`
	ModelUsed      string                 `json:"model_used,omitempty"`
	ToolName       string                 `json:"tool_name,omitempty"`
	ToolArgs       map[string]interface{} `json:"tool_args,omitempty"`
	ParentTurnID   string                 `json:"parent_turn_id,omitempty"`
	ConsultationOf string                 `json:"consultation_of,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
}

// Store persists conversation turns to the database.
// Implements protocol.ConversationLogger.
type Store struct {
	db *sql.DB
}

// NewStore creates a new conversations Store. db may be nil (degraded mode).
func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// LogTurn persists a single conversation turn to the database.
// This is the implementation of protocol.ConversationLogger.
func (s *Store) LogTurn(ctx context.Context, turn protocol.ConversationTurnData) (string, error) {
	if s.db == nil {
		return "", fmt.Errorf("conversations: database not available")
	}

	id := uuid.New().String()
	now := time.Now()

	if turn.TenantID == "" {
		turn.TenantID = "default"
	}

	var toolArgsJSON []byte
	if len(turn.ToolArgs) > 0 {
		var err error
		toolArgsJSON, err = json.Marshal(turn.ToolArgs)
		if err != nil {
			toolArgsJSON = nil
		}
	}

	_, err := s.db.ExecContext(ctx, `
		INSERT INTO conversation_turns
			(id, run_id, session_id, tenant_id, agent_id, team_id, turn_index, role, content,
			 provider_id, model_used, tool_name, tool_args, parent_turn_id, consultation_of, created_at)
		VALUES ($1, NULLIF($2,''), $3, $4, $5, NULLIF($6,''), $7, $8, $9,
		        NULLIF($10,''), NULLIF($11,''), NULLIF($12,''), $13, NULLIF($14,'')::uuid, NULLIF($15,''), $16)
	`, id, turn.RunID, turn.SessionID, turn.TenantID, turn.AgentID, turn.TeamID,
		turn.TurnIndex, turn.Role, turn.Content,
		turn.ProviderID, turn.ModelUsed, turn.ToolName, toolArgsJSON,
		turn.ParentTurnID, turn.ConsultationOf, now)
	if err != nil {
		return "", fmt.Errorf("conversations: persist failed: %w", err)
	}

	return id, nil
}

// GetRunConversation returns all turns for a run in chronological order.
// Optionally filtered by agent_id if agentFilter is non-empty.
func (s *Store) GetRunConversation(ctx context.Context, runID string, agentFilter string) ([]ConversationTurn, error) {
	if s.db == nil {
		return nil, fmt.Errorf("conversations: database not available")
	}

	var rows *sql.Rows
	var err error
	if agentFilter != "" {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, COALESCE(run_id::text, ''), session_id, tenant_id, agent_id,
			       COALESCE(team_id, ''), turn_index, role, content,
			       COALESCE(provider_id, ''), COALESCE(model_used, ''),
			       COALESCE(tool_name, ''), COALESCE(tool_args::text, ''),
			       COALESCE(parent_turn_id::text, ''), COALESCE(consultation_of, ''),
			       created_at
			FROM conversation_turns
			WHERE run_id = $1 AND agent_id = $2
			ORDER BY created_at ASC, turn_index ASC
		`, runID, agentFilter)
	} else {
		rows, err = s.db.QueryContext(ctx, `
			SELECT id, COALESCE(run_id::text, ''), session_id, tenant_id, agent_id,
			       COALESCE(team_id, ''), turn_index, role, content,
			       COALESCE(provider_id, ''), COALESCE(model_used, ''),
			       COALESCE(tool_name, ''), COALESCE(tool_args::text, ''),
			       COALESCE(parent_turn_id::text, ''), COALESCE(consultation_of, ''),
			       created_at
			FROM conversation_turns
			WHERE run_id = $1
			ORDER BY created_at ASC, turn_index ASC
		`, runID)
	}
	if err != nil {
		return nil, fmt.Errorf("conversations: query failed: %w", err)
	}
	defer rows.Close()

	return scanTurns(rows)
}

// GetSessionTurns returns all turns for a specific session.
func (s *Store) GetSessionTurns(ctx context.Context, sessionID string) ([]ConversationTurn, error) {
	if s.db == nil {
		return nil, fmt.Errorf("conversations: database not available")
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT id, COALESCE(run_id::text, ''), session_id, tenant_id, agent_id,
		       COALESCE(team_id, ''), turn_index, role, content,
		       COALESCE(provider_id, ''), COALESCE(model_used, ''),
		       COALESCE(tool_name, ''), COALESCE(tool_args::text, ''),
		       COALESCE(parent_turn_id::text, ''), COALESCE(consultation_of, ''),
		       created_at
		FROM conversation_turns
		WHERE session_id = $1
		ORDER BY turn_index ASC
	`, sessionID)
	if err != nil {
		return nil, fmt.Errorf("conversations: query failed: %w", err)
	}
	defer rows.Close()

	return scanTurns(rows)
}

// scanTurns scans rows into ConversationTurn slices.
func scanTurns(rows *sql.Rows) ([]ConversationTurn, error) {
	var turns []ConversationTurn
	for rows.Next() {
		var t ConversationTurn
		var toolArgsStr string
		if err := rows.Scan(
			&t.ID, &t.RunID, &t.SessionID, &t.TenantID, &t.AgentID,
			&t.TeamID, &t.TurnIndex, &t.Role, &t.Content,
			&t.ProviderID, &t.ModelUsed,
			&t.ToolName, &toolArgsStr,
			&t.ParentTurnID, &t.ConsultationOf,
			&t.CreatedAt,
		); err != nil {
			log.Printf("[conversations] scan error: %v", err)
			continue
		}
		if toolArgsStr != "" && toolArgsStr != "{}" {
			json.Unmarshal([]byte(toolArgsStr), &t.ToolArgs)
		}
		turns = append(turns, t)
	}

	if turns == nil {
		turns = []ConversationTurn{}
	}
	return turns, nil
}
