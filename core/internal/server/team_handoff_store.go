package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/dispatchoutbox"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

type teamHandoffArtifact struct {
	ID          string
	AgentID     string
	Title       string
	ContentType string
	Content     string
	Metadata    map[string]any
}

// teamHandoffRecord is the HandoffNote payload stored on the
// organization.team.handoffs exchange item whose id is the handoff id.
type teamHandoffRecord struct {
	HandoffID      string                  `json:"handoff_id"`
	IdempotencyKey string                  `json:"idempotency_key"`
	RunID          string                  `json:"run_id"`
	SourceTeamID   string                  `json:"source_team"`
	SourceAgentID  string                  `json:"source_agent_id"`
	TargetTeamID   string                  `json:"target_team"`
	TargetRole     string                  `json:"target_role,omitempty"`
	Note           string                  `json:"note"`
	ExpectedAction string                  `json:"expected_action"`
	WorkItemID     string                  `json:"work_item_id"`
	Inputs         []swarm.HandoffInputRef `json:"artifact_refs"`
	CreatedAt      time.Time               `json:"-"`
}

type teamHandoffCommit struct {
	Record      teamHandoffRecord
	WorkItem    protocol.TeamWorkItem
	Event       protocol.TeamStatusEvent
	Interaction protocol.TeamInteraction
	Outbox      dispatchoutbox.Item
}

type teamHandoffStore interface {
	LoadArtifacts(ctx context.Context, ids []string) (map[string]teamHandoffArtifact, error)
	RunWorkItemForTeam(ctx context.Context, runID, teamID string) (*protocol.TeamWorkItem, error)
	GetHandoff(ctx context.Context, handoffID string) (*teamHandoffRecord, error)
	CommitHandoff(ctx context.Context, commit teamHandoffCommit) error
	RecordInteraction(ctx context.Context, interaction protocol.TeamInteraction) error
	ListHandoffs(ctx context.Context, limit int) ([]teamHandoffRecord, error)
	WorkItem(ctx context.Context, teamID, workItemID string) (protocol.TeamWorkItem, error)
	Interactions(ctx context.Context, teamID, workItemID string) ([]protocol.TeamInteraction, error)
}

type sqlTeamHandoffStore struct{ server *AdminServer }

func (st *sqlTeamHandoffStore) db() (*sql.DB, error) {
	if db := st.server.getDB(); db != nil {
		return db, nil
	}
	return nil, errors.New("database not available")
}

func (st *sqlTeamHandoffStore) LoadArtifacts(ctx context.Context, ids []string) (map[string]teamHandoffArtifact, error) {
	db, err := st.db()
	if err != nil {
		return nil, err
	}
	out := map[string]teamHandoffArtifact{}
	for _, id := range ids {
		if _, err := uuid.Parse(id); err != nil {
			continue
		}
		var artifact teamHandoffArtifact
		var meta []byte
		err := db.QueryRowContext(ctx, `
			SELECT id::text, agent_id, title, content_type, COALESCE(content, ''), COALESCE(metadata, '{}'::jsonb)
			FROM artifacts WHERE id = $1::uuid`, id).
			Scan(&artifact.ID, &artifact.AgentID, &artifact.Title, &artifact.ContentType, &artifact.Content, &meta)
		if errors.Is(err, sql.ErrNoRows) {
			continue
		}
		if err != nil {
			return nil, err
		}
		_ = json.Unmarshal(meta, &artifact.Metadata)
		out[id] = artifact
	}
	return out, nil
}

func (st *sqlTeamHandoffStore) RunWorkItemForTeam(ctx context.Context, runID, teamID string) (*protocol.TeamWorkItem, error) {
	if _, err := uuid.Parse(runID); err != nil {
		return nil, nil // work item runs are UUIDs; anything else is not an approved run
	}
	db, err := st.db()
	if err != nil {
		return nil, err
	}
	var workItemID string
	err = db.QueryRowContext(ctx, `
		SELECT id::text FROM team_work_items
		WHERE tenant_id='default' AND team_id=$1 AND run_id=$2::uuid AND intent_proof_id IS NOT NULL AND state <> 'archived'
		ORDER BY created_at ASC LIMIT 1`, teamID, runID).Scan(&workItemID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	item, err := st.server.getTeamWorkItemDB(ctx, teamID, workItemID)
	if err != nil {
		return nil, err
	}
	return &item, nil
}

const teamHandoffSelect = `
	SELECT i.id::text, i.payload, i.created_at
	FROM exchange_items i JOIN exchange_channels c ON c.id = i.channel_id
	WHERE c.name = $1`

func scanTeamHandoff(scan func(...any) error) (teamHandoffRecord, error) {
	var record teamHandoffRecord
	var id string
	var payload []byte
	if err := scan(&id, &payload, &record.CreatedAt); err != nil {
		return record, err
	}
	if err := json.Unmarshal(payload, &record); err != nil {
		return record, fmt.Errorf("decode handoff %s: %w", id, err)
	}
	record.HandoffID = id
	return record, nil
}

func (st *sqlTeamHandoffStore) GetHandoff(ctx context.Context, handoffID string) (*teamHandoffRecord, error) {
	db, err := st.db()
	if err != nil {
		return nil, err
	}
	record, err := scanTeamHandoff(db.QueryRowContext(ctx, teamHandoffSelect+` AND i.id = $2::uuid`, teamHandoffChannel, handoffID).Scan)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &record, nil
}

func (st *sqlTeamHandoffStore) ListHandoffs(ctx context.Context, limit int) ([]teamHandoffRecord, error) {
	db, err := st.db()
	if err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, teamHandoffSelect+` ORDER BY i.created_at DESC LIMIT $2`, teamHandoffChannel, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []teamHandoffRecord{}
	for rows.Next() {
		record, err := scanTeamHandoff(rows.Scan)
		if err != nil {
			return nil, err
		}
		out = append(out, record)
	}
	return out, rows.Err()
}

// CommitHandoff writes the HandoffNote, the receiving team's queued work item
// with its event and handoff_sent interaction, and the team_handoff outbox row
// in one transaction. A missing channel seed fails the whole handoff.
func (st *sqlTeamHandoffStore) CommitHandoff(ctx context.Context, commit teamHandoffCommit) error {
	db, err := st.db()
	if err != nil {
		return err
	}
	if st.server.DispatchOutbox == nil {
		return dispatchoutbox.ErrUnavailable
	}
	payload, err := teamHandoffPayload(commit.Record)
	if err != nil {
		return err
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	result, err := tx.ExecContext(ctx, `
		INSERT INTO exchange_items (id, channel_id, schema_id, payload, created_by, visibility, sensitivity_class, source_role, source_team, target_role, target_team, allowed_consumers, capability_id, trust_class, review_required, metadata, summary)
		SELECT $1::uuid, c.id, 'HandoffNote', $3::jsonb, $4, c.visibility, 'role_scoped', 'team_lead', $5, 'team_lead', $6, '["team_lead"]'::jsonb, 'team_orchestration', 'trusted_internal', FALSE, $7::jsonb, $8
		FROM exchange_channels c WHERE c.name = $2`,
		commit.Record.HandoffID, teamHandoffChannel, string(payload), commit.Record.SourceAgentID, commit.Record.SourceTeamID,
		commit.Record.TargetTeamID, `{"recorded_by":"core.team_handoff"}`, teamHandoffSummary(commit.Record))
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err != nil || affected != 1 {
		return fmt.Errorf("exchange channel %s is not registered", teamHandoffChannel)
	}
	item, event, interaction := commit.WorkItem, commit.Event, commit.Interaction
	if err := st.server.insertTeamWorkItemExec(ctx, tx, &item); err != nil {
		return err
	}
	if err := st.server.insertTeamStatusEventExec(ctx, tx, &event); err != nil {
		return err
	}
	if err := st.server.updateTeamWorkItemLastEventExec(ctx, tx, &item, event); err != nil {
		return err
	}
	if err := st.server.insertTeamInteractionExec(ctx, tx, &interaction); err != nil {
		return err
	}
	if _, err := st.server.DispatchOutbox.EnqueueTx(ctx, tx, commit.Outbox); err != nil {
		return err
	}
	return tx.Commit()
}

func (st *sqlTeamHandoffStore) RecordInteraction(ctx context.Context, interaction protocol.TeamInteraction) error {
	return st.server.insertTeamInteractionDB(ctx, &interaction)
}

func (st *sqlTeamHandoffStore) WorkItem(ctx context.Context, teamID, workItemID string) (protocol.TeamWorkItem, error) {
	return st.server.getTeamWorkItemDB(ctx, teamID, workItemID)
}

func (st *sqlTeamHandoffStore) Interactions(ctx context.Context, teamID, workItemID string) ([]protocol.TeamInteraction, error) {
	return st.server.listTeamInteractionsDB(ctx, teamID, workItemID, 50)
}

func teamHandoffSummary(record teamHandoffRecord) string {
	return fmt.Sprintf("Handoff from %s to %s: %s", record.SourceTeamID, record.TargetTeamID, record.Note)
}

func teamHandoffPayload(record teamHandoffRecord) ([]byte, error) {
	raw, err := json.Marshal(record)
	if err != nil {
		return nil, err
	}
	payload := map[string]any{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	payload["summary"] = teamHandoffSummary(record)
	payload["status"] = "recorded"
	payload["source_role"] = "team_lead"
	payload["created_at"] = time.Now().UTC().Format(time.RFC3339)
	if record.TargetRole == "" {
		payload["target_role"] = "team_lead"
	}
	return json.Marshal(payload)
}
