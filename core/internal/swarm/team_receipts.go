package swarm

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	"github.com/google/uuid"
)

// CommandReceiptStore durably accepts a correlated team command once.
type CommandReceiptStore interface {
	AcceptCommand(context.Context, teamCommandCorrelation, string) (bool, error)
}

type PostgresCommandReceiptStore struct {
	db *sql.DB
}

func NewPostgresCommandReceiptStore(db *sql.DB) *PostgresCommandReceiptStore {
	return &PostgresCommandReceiptStore{db: db}
}

func (s *PostgresCommandReceiptStore) AcceptCommand(ctx context.Context, correlation teamCommandCorrelation, sourceChannel string) (bool, error) {
	if s == nil || s.db == nil {
		return false, errors.New("team command receipt store unavailable")
	}
	teamID := strings.TrimSpace(correlation.TeamID)
	workItemID := strings.TrimSpace(correlation.WorkItemID)
	signalKey := strings.TrimSpace(correlation.commandKey())
	if teamID == "" || workItemID == "" || signalKey == "" {
		return false, errors.New("team command receipt requires team, work item, and signal key")
	}
	result, err := s.db.ExecContext(ctx, `
INSERT INTO team_signal_receipts (
    id, tenant_id, team_id, work_item_id, direction, signal_key, source_channel
) VALUES ($1, 'default', $2, $3, 'command', $4, $5)
ON CONFLICT (tenant_id, team_id, work_item_id, direction, signal_key) DO NOTHING`,
		uuid.NewString(), teamID, workItemID, signalKey, strings.TrimSpace(sourceChannel))
	if err != nil {
		return false, err
	}
	rows, err := result.RowsAffected()
	return rows == 1, err
}
