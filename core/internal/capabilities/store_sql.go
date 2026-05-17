package capabilities

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
)

func selectCapabilityManifestsSQL() string {
	return `
		SELECT id, capability_id, version, manifest_version, display_name, kind,
		       source, status, health, risk_class, description, purpose, tool_refs,
		       default_allowed_roles, allowed_roles, audit_required, approval_required,
		       approval_posture, input_schema_ref, output_schema_ref, last_probe_status,
		       last_probe_at, failure_posture, recovery_posture, audit_policy,
		       secret_ref_policy, owner, metadata, derived_at, updated_at
		FROM capability_manifests`
}

func upsertManifest(ctx context.Context, tx *sql.Tx, m Manifest) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO capability_manifests
			(id, capability_id, version, manifest_version, display_name, kind, source,
			 status, health, risk_class, description, purpose, tool_refs,
			 default_allowed_roles, allowed_roles, audit_required, approval_required,
			 approval_posture, input_schema_ref, output_schema_ref, last_probe_status,
			 last_probe_at, failure_posture, recovery_posture, audit_policy,
			 secret_ref_policy, owner, metadata, derived_at, updated_at)
		VALUES
			($1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			 $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
			 $21, $22, $23, $24, $25, $26, $27, $28, $29, $30)
		ON CONFLICT (id) DO UPDATE SET
			capability_id = EXCLUDED.capability_id,
			version = EXCLUDED.version,
			manifest_version = EXCLUDED.manifest_version,
			display_name = EXCLUDED.display_name,
			kind = EXCLUDED.kind,
			source = EXCLUDED.source,
			status = EXCLUDED.status,
			health = EXCLUDED.health,
			risk_class = EXCLUDED.risk_class,
			description = EXCLUDED.description,
			purpose = EXCLUDED.purpose,
			tool_refs = EXCLUDED.tool_refs,
			default_allowed_roles = EXCLUDED.default_allowed_roles,
			allowed_roles = EXCLUDED.allowed_roles,
			audit_required = EXCLUDED.audit_required,
			approval_required = EXCLUDED.approval_required,
			approval_posture = EXCLUDED.approval_posture,
			input_schema_ref = EXCLUDED.input_schema_ref,
			output_schema_ref = EXCLUDED.output_schema_ref,
			last_probe_status = EXCLUDED.last_probe_status,
			last_probe_at = EXCLUDED.last_probe_at,
			failure_posture = EXCLUDED.failure_posture,
			recovery_posture = EXCLUDED.recovery_posture,
			audit_policy = EXCLUDED.audit_policy,
			secret_ref_policy = EXCLUDED.secret_ref_policy,
			owner = EXCLUDED.owner,
			metadata = EXCLUDED.metadata,
			derived_at = EXCLUDED.derived_at,
			updated_at = EXCLUDED.updated_at`,
		m.ID, m.CapabilityID, m.Version, m.ManifestVersion, m.DisplayName, m.Kind,
		m.Source, m.Status, m.Health, m.RiskClass, m.Description, m.Purpose,
		marshalJSON(m.ToolRefs), marshalJSON(m.DefaultAllowedRoles), marshalJSON(m.AllowedRoles),
		m.AuditRequired, m.ApprovalRequired, m.ApprovalPosture, m.InputSchemaRef,
		m.OutputSchemaRef, m.LastProbeStatus, m.LastProbeAt, m.FailurePosture,
		m.RecoveryPosture, m.AuditPolicy, m.SecretRefPolicy, m.Owner,
		marshalJSON(m.Metadata), m.DerivedAt, m.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("capability manifests: upsert %s: %w", m.ID, err)
	}
	return nil
}

func deleteStaleManifests(ctx context.Context, tx *sql.Tx, ids []string) error {
	if len(ids) == 0 {
		_, err := tx.ExecContext(ctx, `DELETE FROM capability_manifests`)
		if err != nil {
			return fmt.Errorf("capability manifests: clear stale rows: %w", err)
		}
		return nil
	}
	placeholders := make([]string, len(ids))
	args := make([]any, len(ids))
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+1)
		args[i] = id
	}
	_, err := tx.ExecContext(ctx,
		`DELETE FROM capability_manifests WHERE id NOT IN (`+strings.Join(placeholders, ", ")+`)`,
		args...,
	)
	if err != nil {
		return fmt.Errorf("capability manifests: delete stale rows: %w", err)
	}
	return nil
}
