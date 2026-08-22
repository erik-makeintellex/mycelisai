package server

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/mycelis/core/internal/conversations"
	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) logConfigDocumentThreadReceiptsTx(
	ctx context.Context,
	tx *sql.Tx,
	scope *protocol.ScopeValidation,
	results []plannedToolExecutionResult,
) error {
	if tx == nil {
		return fmt.Errorf("configuration receipt transaction is required")
	}
	if s.Conversations == nil {
		return fmt.Errorf("configuration receipt store is unavailable")
	}
	if scope == nil || strings.TrimSpace(scope.ConversationSessionID) == "" {
		return fmt.Errorf("configuration receipt session is required")
	}
	if _, err := tx.ExecContext(ctx, `
		SELECT pg_advisory_xact_lock(hashtextextended($1, 0))
	`, scope.ConversationSessionID); err != nil {
		return fmt.Errorf("lock configuration receipt session: %w", err)
	}
	var turnIndex int
	if err := tx.QueryRowContext(ctx, `
		SELECT COALESCE(MAX(turn_index), -1) + 1
		FROM conversation_turns
		WHERE session_id = $1::uuid
	`, scope.ConversationSessionID).Scan(&turnIndex); err != nil {
		return err
	}
	for _, result := range results {
		if !isConfigDocumentMutationTool(result.Name) {
			continue
		}
		if _, ok := configDocumentReceiptFromResult(result); !ok {
			return sql.ErrNoRows
		}
		if _, err := s.Conversations.LogTurnTx(ctx, tx, protocol.ConversationTurnData{
			SessionID: scope.ConversationSessionID, TenantID: "default", AgentID: "admin",
			TeamID: "admin-core", TurnIndex: turnIndex, Role: "assistant",
			Content: configDocumentResultSummary(result),
		}); err != nil {
			return err
		}
		turnIndex++
		if _, err := s.Conversations.LogTurnTx(ctx, tx, protocol.ConversationTurnData{
			SessionID: scope.ConversationSessionID, TenantID: "default", AgentID: "admin",
			TeamID: "admin-core", TurnIndex: turnIndex, Role: "tool_result",
			Content: result.Output, ToolName: result.Name,
		}); err != nil {
			return err
		}
		turnIndex++
	}
	return nil
}

func (s *AdminServer) latestSessionConfigDocumentReceipt(
	ctx context.Context,
	sessionID string,
) (configDocumentReceipt, bool) {
	if s == nil || s.Conversations == nil || strings.TrimSpace(sessionID) == "" {
		return configDocumentReceipt{}, false
	}
	turns, err := s.Conversations.GetSessionTurns(ctx, sessionID)
	if err != nil {
		return configDocumentReceipt{}, false
	}
	return latestConfigDocumentReceiptFromTurns(turns)
}

func latestConfigDocumentReceiptFromTurns(turns []conversations.ConversationTurn) (configDocumentReceipt, bool) {
	for i := len(turns) - 1; i >= 0; i-- {
		toolName := strings.TrimSpace(turns[i].ToolName)
		if toolName != "store_config_document" && toolName != "activate_config_document" {
			continue
		}
		if receipt, ok := configDocumentReceiptFromResult(plannedToolExecutionResult{
			Name: toolName, Output: turns[i].Content,
		}); ok {
			return receipt, true
		}
	}
	return configDocumentReceipt{}, false
}
