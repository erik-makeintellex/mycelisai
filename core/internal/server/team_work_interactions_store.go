package server

import (
	"context"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) listTeamInteractionsDB(ctx context.Context, teamID, workItemID string, limit int) ([]protocol.TeamInteraction, error) {
	db := s.getDB()
	if db == nil {
		return nil, errors.New("database not available")
	}
	rows, err := db.QueryContext(ctx, `
		SELECT id::text, team_id, work_item_id::text, COALESCE(run_id::text,''),
		       COALESCE(intent_proof_id::text,''), COALESCE(contract_id,''), COALESCE(proof_id,''),
		       source_kind, source_channel, actor_ref, verb, summary, payload_kind,
		       payload_ref, COALESCE(payload, 'null'::jsonb), approval_ref, audit_refs, timestamp, version
		FROM team_interactions
		WHERE tenant_id='default' AND team_id=$1 AND work_item_id=$2
		ORDER BY timestamp ASC
		LIMIT $3`, strings.TrimSpace(teamID), workItemID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]protocol.TeamInteraction, 0)
	for rows.Next() {
		item, scanErr := scanTeamInteraction(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return items, nil
}

func (s *AdminServer) insertTeamInteractionDB(ctx context.Context, item *protocol.TeamInteraction) error {
	db := s.getDB()
	if db == nil {
		return errors.New("database not available")
	}
	return s.insertTeamInteractionExec(ctx, db, item)
}

func (s *AdminServer) insertTeamInteractionExec(ctx context.Context, exec teamWorkSQLExecutor, item *protocol.TeamInteraction) error {
	if exec == nil {
		return errors.New("database not available")
	}
	if strings.TrimSpace(item.InteractionID) == "" {
		item.InteractionID = uuid.NewString()
	}
	if err := validateTeamInteractionUUIDLinks(*item); err != nil {
		return err
	}
	return exec.QueryRowContext(ctx, `
		INSERT INTO team_interactions (
			id, tenant_id, team_id, work_item_id, run_id, intent_proof_id, contract_id, proof_id,
			source_kind, source_channel, actor_ref, verb, summary, payload_kind,
			payload_ref, payload, approval_ref, audit_refs, version
		) VALUES (
			$1, 'default', $2, $3, $4, $5, NULLIF($6,''), NULLIF($7,''),
			$8, $9, $10, $11, $12, $13,
			$14, $15, $16, $17, $18
		)
		RETURNING timestamp`,
		item.InteractionID, item.TeamID, item.WorkItemID, nullableUUID(item.RunID),
		nullableUUID(item.IntentProofID), item.ContractID, item.ProofID, item.SourceKind,
		item.SourceChannel, item.ActorRef, item.Verb, item.Summary, item.PayloadKind,
		item.PayloadRef, jsonObjectOrNil(item.Payload), item.ApprovalRef, jsonArray(item.AuditRefs), item.Version,
	).Scan(&item.Timestamp)
}
