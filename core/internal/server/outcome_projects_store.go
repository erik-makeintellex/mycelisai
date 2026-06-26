package server

import (
	"context"
	"encoding/json"
	"errors"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) insertOutcomeProjectDB(ctx context.Context, item *protocol.OutcomeProject) error {
	db := s.getDB()
	if db == nil {
		return errors.New("database not available")
	}
	return s.insertOutcomeProjectExec(ctx, db, item)
}

func (s *AdminServer) insertOutcomeProjectExec(ctx context.Context, exec teamWorkSQLExecutor, item *protocol.OutcomeProject) error {
	if exec == nil {
		return errors.New("database not available")
	}
	normalized := protocol.NormalizeOutcomeProject(*item)
	if strings.TrimSpace(normalized.ProjectID) == "" {
		normalized.ProjectID = uuid.NewString()
	}
	normalized.TargetRef = protocol.TargetRefForOutcomeProject(normalized)
	if err := protocol.ValidateOutcomeProject(normalized); err != nil {
		return err
	}
	if err := exec.QueryRowContext(ctx, `
		INSERT INTO outcome_projects (
			id, tenant_id, outcome_id, title, purpose, execution_mode, workspace_folder,
			status, run_id, intent_proof_id, contract_id, proof_id, work_item_refs,
			output_refs, proof_refs, recovery_refs, retention_policy, version
		) VALUES (
			$1, 'default', $2, $3, $4, $5, $6,
			$7, NULLIF($8,''), NULLIF($9,''), NULLIF($10,''), NULLIF($11,''), $12,
			$13, $14, $15, $16, $17
		)
		RETURNING created_at, updated_at`,
		normalized.ProjectID, normalized.OutcomeID, normalized.Title, normalized.Purpose,
		normalized.ExecutionMode, normalized.WorkspaceFolder, string(normalized.Status),
		normalized.RunID, normalized.IntentProofID, normalized.ContractID,
		normalized.ProofID, jsonArray(normalized.WorkItemRefs),
		jsonArray(normalized.OutputRefs), jsonArray(normalized.ProofRefs),
		jsonArray(normalized.RecoveryRefs), normalized.RetentionPolicy, normalized.Version,
	).Scan(&normalized.CreatedAt, &normalized.UpdatedAt); err != nil {
		return err
	}
	*item = normalized
	return nil
}

func (s *AdminServer) listOutcomeProjectsDB(ctx context.Context, limit int) ([]protocol.OutcomeProject, error) {
	db := s.getDB()
	if db == nil {
		return nil, errors.New("database not available")
	}
	rows, err := db.QueryContext(ctx, `
		SELECT id::text, outcome_id, title, purpose, execution_mode, workspace_folder,
		       status, COALESCE(run_id,''), COALESCE(intent_proof_id,''),
		       COALESCE(contract_id,''), COALESCE(proof_id,''), work_item_refs,
		       output_refs, proof_refs, recovery_refs, retention_policy,
		       created_at, updated_at, version
		FROM outcome_projects
		WHERE tenant_id='default'
		ORDER BY updated_at DESC
		LIMIT $1`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	projects := make([]protocol.OutcomeProject, 0)
	for rows.Next() {
		item, scanErr := scanOutcomeProject(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		entries, _ := s.listTeamRegistryEntriesDB(ctx, item.ProjectID, 50)
		for _, entry := range entries {
			item.TeamRegistryRefs = append(item.TeamRegistryRefs, entry.RegistryID)
		}
		projects = append(projects, item)
	}
	return projects, rows.Err()
}

func (s *AdminServer) getOutcomeProjectDB(ctx context.Context, projectID string) (protocol.OutcomeProject, error) {
	db := s.getDB()
	if db == nil {
		return protocol.OutcomeProject{}, errors.New("database not available")
	}
	item, err := scanOutcomeProject(db.QueryRowContext(ctx, `
		SELECT id::text, outcome_id, title, purpose, execution_mode, workspace_folder,
		       status, COALESCE(run_id,''), COALESCE(intent_proof_id,''),
		       COALESCE(contract_id,''), COALESCE(proof_id,''), work_item_refs,
		       output_refs, proof_refs, recovery_refs, retention_policy,
		       created_at, updated_at, version
		FROM outcome_projects
		WHERE tenant_id='default' AND id=$1`, projectID))
	if err != nil {
		return item, err
	}
	entries, _ := s.listTeamRegistryEntriesDB(ctx, item.ProjectID, 50)
	for _, entry := range entries {
		item.TeamRegistryRefs = append(item.TeamRegistryRefs, entry.RegistryID)
	}
	return item, nil
}

func (s *AdminServer) insertTeamRegistryEntryDB(ctx context.Context, item *protocol.TeamRegistryEntry) error {
	db := s.getDB()
	if db == nil {
		return errors.New("database not available")
	}
	return s.insertTeamRegistryEntryExec(ctx, db, item)
}

func (s *AdminServer) insertTeamRegistryEntryExec(ctx context.Context, exec teamWorkSQLExecutor, item *protocol.TeamRegistryEntry) error {
	if exec == nil {
		return errors.New("database not available")
	}
	normalized := protocol.NormalizeTeamRegistryEntry(*item)
	if strings.TrimSpace(normalized.RegistryID) == "" {
		normalized.RegistryID = uuid.NewString()
	}
	if err := protocol.ValidateTeamRegistryEntry(normalized); err != nil {
		return err
	}
	if err := exec.QueryRowContext(ctx, `
		INSERT INTO team_registry_entries (
			id, tenant_id, project_id, group_id, role, team_id, agent_id,
			assignment_reason, temporary, expires_at, status, version
		) VALUES (
			$1, 'default', $2, $3, $4, $5, $6,
			$7, $8, $9, $10, $11
		)
		RETURNING created_at, updated_at`,
		normalized.RegistryID, normalized.ProjectID, normalized.GroupID, normalized.Role,
		normalized.TeamID, normalized.AgentID, normalized.AssignmentReason,
		normalized.Temporary, normalized.ExpiresAt, normalized.Status, normalized.Version,
	).Scan(&normalized.CreatedAt, &normalized.UpdatedAt); err != nil {
		return err
	}
	*item = normalized
	return nil
}

func (s *AdminServer) listTeamRegistryEntriesDB(ctx context.Context, projectID string, limit int) ([]protocol.TeamRegistryEntry, error) {
	db := s.getDB()
	if db == nil {
		return nil, errors.New("database not available")
	}
	rows, err := db.QueryContext(ctx, `
		SELECT id::text, project_id::text, group_id, role, team_id, agent_id,
		       assignment_reason, temporary, expires_at, status, created_at, updated_at, version
		FROM team_registry_entries
		WHERE tenant_id='default' AND project_id=$1
		ORDER BY created_at ASC
		LIMIT $2`, projectID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	entries := make([]protocol.TeamRegistryEntry, 0)
	for rows.Next() {
		item, scanErr := scanTeamRegistryEntry(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		entries = append(entries, item)
	}
	return entries, rows.Err()
}

func scanOutcomeProject(scanner interface{ Scan(dest ...any) error }) (protocol.OutcomeProject, error) {
	var item protocol.OutcomeProject
	var status string
	var workRefs, outputRefs, proofRefs, recoveryRefs []byte
	if err := scanner.Scan(
		&item.ProjectID, &item.OutcomeID, &item.Title, &item.Purpose, &item.ExecutionMode,
		&item.WorkspaceFolder, &status, &item.RunID, &item.IntentProofID, &item.ContractID,
		&item.ProofID, &workRefs, &outputRefs, &proofRefs, &recoveryRefs, &item.RetentionPolicy,
		&item.CreatedAt, &item.UpdatedAt, &item.Version,
	); err != nil {
		return item, err
	}
	item.Status = protocol.OutcomeProjectStatus(status)
	item.WorkItemRefs = decodeStringList(workRefs)
	item.ProofRefs = decodeStringList(proofRefs)
	item.RecoveryRefs = decodeStringList(recoveryRefs)
	_ = json.Unmarshal(outputRefs, &item.OutputRefs)
	return protocol.NormalizeOutcomeProject(item), nil
}

func scanTeamRegistryEntry(scanner interface{ Scan(dest ...any) error }) (protocol.TeamRegistryEntry, error) {
	var item protocol.TeamRegistryEntry
	if err := scanner.Scan(
		&item.RegistryID, &item.ProjectID, &item.GroupID, &item.Role, &item.TeamID,
		&item.AgentID, &item.AssignmentReason, &item.Temporary, &item.ExpiresAt,
		&item.Status, &item.CreatedAt, &item.UpdatedAt, &item.Version,
	); err != nil {
		return item, err
	}
	return protocol.NormalizeTeamRegistryEntry(item), nil
}
