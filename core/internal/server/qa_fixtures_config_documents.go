package server

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/google/uuid"
)

const configDocumentActivationFixturePrefix = "activation:"

func configDocumentActivationFixtureRef(historyID string) string {
	return configDocumentActivationFixturePrefix + strings.TrimSpace(historyID)
}

func parseConfigDocumentFixtureRef(ref string) (string, string, bool) {
	ref = strings.TrimSpace(ref)
	kind := "revision"
	value := ref
	if strings.HasPrefix(ref, configDocumentActivationFixturePrefix) {
		kind = "activation"
		value = strings.TrimPrefix(ref, configDocumentActivationFixturePrefix)
	}
	if _, err := uuid.Parse(value); err != nil {
		return "", "", false
	}
	return kind, value, true
}

func deleteQAFixtureConfigDocuments(
	ctx context.Context,
	tx *sql.Tx,
	tenantID string,
	refs []string,
	deleted map[string]int64,
) error {
	var revisionIDs, activationIDs []string
	revisions := make(map[string]bool)
	activations := make(map[string]bool)
	for _, ref := range refs {
		kind, id, ok := parseConfigDocumentFixtureRef(ref)
		if !ok {
			continue
		}
		if kind == "activation" {
			if !activations[id] {
				activationIDs = append(activationIDs, id)
				activations[id] = true
			}
		} else if !revisions[id] {
			revisionIDs = append(revisionIDs, id)
			revisions[id] = true
		}
	}
	transitions, err := loadFixtureConfigActivations(ctx, tx, tenantID, activationIDs)
	if err != nil {
		return err
	}
	restores, err := planFixtureConfigActivationRestores(ctx, tx, tenantID, transitions)
	if err != nil {
		return err
	}
	if err := validateFixtureConfigRevisionOwnership(ctx, tx, tenantID, revisionIDs, transitions, restores); err != nil {
		return err
	}
	if err := restoreFixtureConfigActivations(ctx, tx, tenantID, restores, deleted); err != nil {
		return err
	}
	for _, transition := range transitions {
		result, err := tx.ExecContext(ctx, `
			DELETE FROM config_document_activation_history
			WHERE tenant_id=$1 AND id=$2::uuid
		`, tenantID, transition.HistoryID)
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows != 1 {
			return fmt.Errorf("fixture config activation cleanup lost ownership of history %q", transition.HistoryID)
		}
		deleted["config_document_activation_history"] += rows
	}
	for _, recordID := range revisionIDs {
		if err := deleteFixtureConfigRevision(ctx, tx, tenantID, recordID, deleted); err != nil {
			return err
		}
	}
	return nil
}

func deleteFixtureConfigRevision(
	ctx context.Context,
	tx *sql.Tx,
	tenantID string,
	recordID string,
	deleted map[string]int64,
) error {
	result, err := tx.ExecContext(ctx, `
		DELETE FROM config_documents
		WHERE tenant_id=$1 AND record_id=$2::uuid
	`, tenantID, recordID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	deleted["config_documents"] += rows
	return nil
}
