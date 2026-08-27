package server

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

func (s *AdminServer) resolveOutputContinuationOwnership(ctx context.Context, input *chatContinuationContext) (*chatContinuationContext, error) {
	if input == nil || input.Intent == "inspect" || input.Intent == "follow_up" {
		return input, nil
	}
	if s.getDB() == nil {
		return nil, errors.New("output continuation ownership cannot be validated without durable work storage")
	}
	items, err := s.findContinuationWorkItems(ctx, input)
	if err != nil {
		return nil, err
	}
	if len(items) != 1 {
		return nil, fmt.Errorf("output continuation ownership matched %d retained work items", len(items))
	}
	item := items[0]
	output, ok := matchingContinuationOutput(item, input)
	if !ok {
		return nil, errors.New("output continuation does not match a retained output owned by the work item")
	}
	digest, err := teamWorkOutputDigest([]protocol.TeamOutputRef{output})
	if err != nil {
		return nil, fmt.Errorf("compute immutable continuation source digest: %w", err)
	}
	if input.SourceDigest != "" && input.SourceDigest != digest {
		return nil, errors.New("output continuation source digest is stale")
	}
	out := *input
	out.TeamID, out.RunID, out.WorkItemID = item.TeamID, item.RunID, item.WorkItemID
	out.OutputID, out.SourceDigest, out.SourceVersion = output.OutputID, digest, item.Version
	out.Reference = firstNonEmptyString(output.StorageRef, input.Reference)
	out.Proof = firstNonEmptyString(output.ProofRef, output.ProofID, input.Proof)
	out.RevisionTarget = continuationRevisionTarget(output.StorageRef)
	out.SourceWorkIntent = protocol.NormalizeWorkIntent(item.WorkIntent)
	out.OwnershipValidated = true
	return &out, nil
}

func (s *AdminServer) findContinuationWorkItems(ctx context.Context, input *chatContinuationContext) ([]protocol.TeamWorkItem, error) {
	if input.TeamID != "" && input.WorkItemID != "" {
		item, err := s.getTeamWorkItemDB(ctx, input.TeamID, input.WorkItemID)
		if err != nil {
			return nil, err
		}
		return []protocol.TeamWorkItem{item}, nil
	}
	rows, err := s.getDB().QueryContext(ctx, `
		SELECT id::text, team_id, COALESCE(run_id::text,''), COALESCE(intent_proof_id::text,''),
		       COALESCE(contract_id,''), COALESCE(proof_id,''), objective, scope, owner,
		       execution_shape, execution_mode, work_intent, expected_outputs, expected_proof, capability_requirements,
		       governance_posture, state, COALESCE(last_event, 'null'::jsonb), needs_operator,
		       degradation_state, recovery_options, output_refs, proof_refs, audit_refs,
		       created_at, updated_at, version
		FROM team_work_items
		WHERE tenant_id='default' AND EXISTS (
			SELECT 1 FROM jsonb_array_elements(output_refs) AS output
			WHERE output->>'storage_ref'=$1 OR output->>'proof_ref'=$2 OR output->>'proof_id'=$2
		)
		ORDER BY updated_at DESC LIMIT 2`, strings.TrimSpace(input.Reference), strings.TrimSpace(input.Proof))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	items := []protocol.TeamWorkItem{}
	for rows.Next() {
		item, scanErr := scanTeamWorkItem(rows)
		if scanErr != nil {
			return nil, scanErr
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

func matchingContinuationOutput(item protocol.TeamWorkItem, input *chatContinuationContext) (protocol.TeamOutputRef, bool) {
	if input.TeamID != "" && input.TeamID != item.TeamID {
		return protocol.TeamOutputRef{}, false
	}
	if input.RunID != "" && input.RunID != item.RunID {
		return protocol.TeamOutputRef{}, false
	}
	for _, output := range item.OutputRefs {
		if input.OutputID != "" && input.OutputID != output.OutputID {
			continue
		}
		referenceMatch := input.Reference == "" || input.Reference == output.StorageRef ||
			input.Reference == strings.TrimRight(output.StorageRef, "/")+"/"+strings.TrimLeft(output.Entrypoint, "/")
		proofMatch := input.Proof == "" || input.Proof == output.ProofRef || input.Proof == output.ProofID
		if referenceMatch && proofMatch {
			return output, true
		}
	}
	return protocol.TeamOutputRef{}, false
}

func continuationRevisionTarget(source string) string {
	return strings.TrimRight(strings.TrimSpace(source), "/") + "-v2"
}
