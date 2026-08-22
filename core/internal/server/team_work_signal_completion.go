package server

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func signalWorkItemID(payload map[string]any) string {
	if id := stringField(payload, "work_item_id"); id != "" {
		return id
	}
	if contextValue, ok := payload["context"].(map[string]any); ok {
		return stringField(contextValue, "work_item_id")
	}
	return ""
}

func (p *teamWorkSignalProjection) isFinalLinkedTeamResult(ctx context.Context, tx *sql.Tx, item protocol.TeamWorkItem, payloadKind protocol.SignalPayloadKind) (bool, error) {
	if payloadKind != protocol.PayloadKindResult || item.State != protocol.TeamWorkStateOutputReady || strings.TrimSpace(item.RunID) == "" {
		return false, nil
	}
	var pending int
	err := tx.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM team_work_items
		WHERE run_id = $1
		  AND id <> $2
		  AND execution_shape <> 'create_team'
		  AND state NOT IN ('output_ready', 'archived')`, item.RunID, item.WorkItemID).Scan(&pending)
	if err != nil {
		return false, fmt.Errorf("count pending linked team work: %w", err)
	}
	return pending == 0, nil
}
