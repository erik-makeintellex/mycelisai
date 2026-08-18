package server

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"sort"
	"strings"
	"time"
)

type fixtureConfigActivation struct {
	HistoryID  string
	Kind       string
	DocumentID string
	ScopeKind  string
	ScopeRef   string
	FromRecord string
	ToRecord   string
	CreatedAt  time.Time
}

type fixtureConfigActivationRestore struct {
	Activation fixtureConfigActivation
	Current    string
	Baseline   string
}

func loadFixtureConfigActivations(
	ctx context.Context,
	tx *sql.Tx,
	tenantID string,
	historyIDs []string,
) ([]fixtureConfigActivation, error) {
	transitions := make([]fixtureConfigActivation, 0, len(historyIDs))
	for _, historyID := range historyIDs {
		var item fixtureConfigActivation
		var from sql.NullString
		err := tx.QueryRowContext(ctx, `
			SELECT id::text, kind, document_id, scope_kind, scope_ref,
			       from_record_id::text, to_record_id::text, created_at
			FROM config_document_activation_history
			WHERE tenant_id=$1 AND id=$2::uuid
			FOR UPDATE
		`, tenantID, historyID).Scan(
			&item.HistoryID, &item.Kind, &item.DocumentID, &item.ScopeKind,
			&item.ScopeRef, &from, &item.ToRecord, &item.CreatedAt,
		)
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("fixture config activation history %q is missing", historyID)
		}
		if err != nil {
			return nil, err
		}
		item.FromRecord = from.String
		transitions = append(transitions, item)
	}
	return transitions, nil
}

func planFixtureConfigActivationRestores(
	ctx context.Context,
	tx *sql.Tx,
	tenantID string,
	transitions []fixtureConfigActivation,
) ([]fixtureConfigActivationRestore, error) {
	groups := make(map[string][]fixtureConfigActivation)
	for _, item := range transitions {
		key := fixtureConfigActivationKey(item.Kind, item.DocumentID, item.ScopeKind, item.ScopeRef)
		groups[key] = append(groups[key], item)
	}
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	restores := make([]fixtureConfigActivationRestore, 0, len(keys))
	for _, key := range keys {
		group := groups[key]
		if err := ensureFixtureConfigActivationHistoryOwned(ctx, tx, tenantID, group); err != nil {
			return nil, err
		}
		current, err := currentFixtureConfigActivation(ctx, tx, tenantID, group[0])
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("fixture config activation chain %q has no current activation", key)
		}
		if err != nil {
			return nil, err
		}
		baseline, intact := fixtureConfigActivationBaseline(current, group)
		if !intact {
			return nil, fmt.Errorf("fixture config activation chain %q is not an owned suffix of current activation %q", key, current)
		}
		restores = append(restores, fixtureConfigActivationRestore{
			Activation: group[0], Current: current, Baseline: baseline,
		})
	}
	return restores, nil
}

func ensureFixtureConfigActivationHistoryOwned(
	ctx context.Context,
	tx *sql.Tx,
	tenantID string,
	transitions []fixtureConfigActivation,
) error {
	claimed := make(map[string]bool, len(transitions))
	earliest := transitions[0].CreatedAt
	for _, item := range transitions {
		claimed[item.HistoryID] = true
		if item.CreatedAt.Before(earliest) {
			earliest = item.CreatedAt
		}
	}
	item := transitions[0]
	rows, err := tx.QueryContext(ctx, `
		SELECT id::text
		FROM config_document_activation_history
		WHERE tenant_id=$1 AND kind=$2 AND document_id=$3 AND scope_kind=$4
		  AND scope_ref=$5 AND created_at >= $6
		FOR UPDATE
	`, tenantID, item.Kind, item.DocumentID, item.ScopeKind, item.ScopeRef, earliest)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var historyID string
		if err := rows.Scan(&historyID); err != nil {
			return err
		}
		if !claimed[historyID] {
			return fmt.Errorf("fixture config activation cleanup found unowned history %q after the fixture", historyID)
		}
	}
	return rows.Err()
}

func validateFixtureConfigRevisionOwnership(
	ctx context.Context,
	tx *sql.Tx,
	tenantID string,
	revisionIDs []string,
	transitions []fixtureConfigActivation,
	restores []fixtureConfigActivationRestore,
) error {
	claimedTransitions := make(map[string]bool, len(transitions))
	for _, transition := range transitions {
		claimedTransitions[transition.HistoryID] = true
	}
	claimedRevisions := make(map[string]bool, len(revisionIDs))
	for _, recordID := range revisionIDs {
		claimedRevisions[recordID] = true
	}
	restoreByKey := make(map[string]fixtureConfigActivationRestore, len(restores))
	for _, restore := range restores {
		if claimedRevisions[restore.Baseline] {
			return fmt.Errorf("fixture config activation cleanup would restore claimed revision %q", restore.Baseline)
		}
		item := restore.Activation
		restoreByKey[fixtureConfigActivationKey(item.Kind, item.DocumentID, item.ScopeKind, item.ScopeRef)] = restore
	}

	for _, recordID := range revisionIDs {
		rows, err := tx.QueryContext(ctx, `
			SELECT id::text
			FROM config_document_activation_history
			WHERE tenant_id=$1 AND (to_record_id=$2::uuid OR from_record_id=$2::uuid)
			FOR UPDATE
		`, tenantID, recordID)
		if err != nil {
			return err
		}
		for rows.Next() {
			var historyID string
			if err := rows.Scan(&historyID); err != nil {
				rows.Close()
				return err
			}
			if !claimedTransitions[historyID] {
				rows.Close()
				return fmt.Errorf("fixture config revision %q is referenced by unowned activation history %q", recordID, historyID)
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		if err := rows.Close(); err != nil {
			return err
		}

		rows, err = tx.QueryContext(ctx, `
			SELECT kind, document_id, scope_kind, scope_ref
			FROM config_document_activations
			WHERE tenant_id=$1 AND config_document_record_id=$2::uuid
			FOR UPDATE
		`, tenantID, recordID)
		if err != nil {
			return err
		}
		for rows.Next() {
			var kind, documentID, scopeKind, scopeRef string
			if err := rows.Scan(&kind, &documentID, &scopeKind, &scopeRef); err != nil {
				rows.Close()
				return err
			}
			key := fixtureConfigActivationKey(kind, documentID, scopeKind, scopeRef)
			restore, ok := restoreByKey[key]
			if !ok || restore.Current != recordID {
				rows.Close()
				return fmt.Errorf("fixture config revision %q is retained by an unowned current activation", recordID)
			}
		}
		if err := rows.Err(); err != nil {
			rows.Close()
			return err
		}
		if err := rows.Close(); err != nil {
			return err
		}
	}
	return nil
}

func restoreFixtureConfigActivations(
	ctx context.Context,
	tx *sql.Tx,
	tenantID string,
	restores []fixtureConfigActivationRestore,
	deleted map[string]int64,
) error {
	for _, restore := range restores {
		item := restore.Activation
		var result sql.Result
		var err error
		if restore.Baseline == "" {
			result, err = tx.ExecContext(ctx, `
				DELETE FROM config_document_activations
				WHERE tenant_id=$1 AND kind=$2 AND document_id=$3 AND scope_kind=$4
				  AND scope_ref=$5 AND config_document_record_id=$6::uuid
			`, tenantID, item.Kind, item.DocumentID, item.ScopeKind, item.ScopeRef, restore.Current)
		} else {
			result, err = tx.ExecContext(ctx, `
				UPDATE config_document_activations
				SET config_document_record_id=$6::uuid, activated_by='qa-fixture-cleanup', activated_at=NOW()
				WHERE tenant_id=$1 AND kind=$2 AND document_id=$3 AND scope_kind=$4
				  AND scope_ref=$5 AND config_document_record_id=$7::uuid
			`, tenantID, item.Kind, item.DocumentID, item.ScopeKind, item.ScopeRef, restore.Baseline, restore.Current)
		}
		if err != nil {
			return err
		}
		rows, err := result.RowsAffected()
		if err != nil {
			return err
		}
		if rows != 1 {
			return fmt.Errorf("fixture config activation cleanup lost ownership of current activation %q", restore.Current)
		}
		deleted["config_document_activations_restored"] += rows
	}
	return nil
}

func fixtureConfigActivationKey(kind, documentID, scopeKind, scopeRef string) string {
	return strings.Join([]string{kind, documentID, scopeKind, scopeRef}, "\x00")
}

func currentFixtureConfigActivation(
	ctx context.Context,
	tx *sql.Tx,
	tenantID string,
	item fixtureConfigActivation,
) (string, error) {
	var current string
	err := tx.QueryRowContext(ctx, `
		SELECT config_document_record_id::text
		FROM config_document_activations
		WHERE tenant_id=$1 AND kind=$2 AND document_id=$3 AND scope_kind=$4 AND scope_ref=$5
		FOR UPDATE
	`, tenantID, item.Kind, item.DocumentID, item.ScopeKind, item.ScopeRef).Scan(&current)
	return current, err
}

func fixtureConfigActivationBaseline(
	current string,
	transitions []fixtureConfigActivation,
) (string, bool) {
	if len(transitions) == 0 {
		return current, false
	}
	baseline := current
	used := make([]bool, len(transitions))
	var cutoff time.Time
	for step := 0; step < len(transitions); step++ {
		match := -1
		ambiguous := false
		for i, item := range transitions {
			if used[i] || item.ToRecord != baseline || (!cutoff.IsZero() && item.CreatedAt.After(cutoff)) {
				continue
			}
			if match < 0 || item.CreatedAt.After(transitions[match].CreatedAt) {
				match = i
				ambiguous = false
			} else if item.CreatedAt.Equal(transitions[match].CreatedAt) {
				ambiguous = true
			}
		}
		if match < 0 || ambiguous {
			return current, false
		}
		used[match] = true
		baseline = transitions[match].FromRecord
		cutoff = transitions[match].CreatedAt
	}
	return baseline, true
}
