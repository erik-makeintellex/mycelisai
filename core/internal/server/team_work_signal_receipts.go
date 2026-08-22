package server

import (
	"context"
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

func claimProjectedTeamSignal(ctx context.Context, exec teamWorkSQLExecutor, item protocol.TeamWorkItem, subject string, payloadKind protocol.SignalPayloadKind, payload map[string]any, raw []byte) (string, bool, error) {
	signalKey := projectedTeamSignalKey(subject, payloadKind, payload, raw)
	receiptID := projectedTeamSignalReceiptID(item, signalKey)
	result, err := exec.ExecContext(ctx, `
INSERT INTO team_signal_receipts (
    id, tenant_id, team_id, work_item_id, direction, signal_key, source_channel
) VALUES ($1, 'default', $2, $3, 'result', $4, $5)
ON CONFLICT (tenant_id, team_id, work_item_id, direction, signal_key) DO NOTHING`,
		receiptID, item.TeamID, item.WorkItemID, signalKey, subject)
	if err != nil {
		return "", false, err
	}
	rows, err := result.RowsAffected()
	return receiptID, rows == 1, err
}

func projectedTeamSignalReceiptID(item protocol.TeamWorkItem, signalKey string) string {
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(strings.Join([]string{
		item.TeamID, item.WorkItemID, "projection", signalKey,
	}, ":"))).String()
}

func projectedTeamSignalKey(subject string, payloadKind protocol.SignalPayloadKind, payload map[string]any, raw []byte) string {
	if key := stringField(payload, "idempotency_key"); key != "" {
		return string(payloadKind) + ":" + key
	}
	hash := sha256.Sum256(raw)
	return fmt.Sprintf("%s:%s:%x", payloadKind, subject, hash)
}
