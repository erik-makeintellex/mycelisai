package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

func projectedSignalOutputRefs(item protocol.TeamWorkItem, env protocol.SignalEnvelope, payload map[string]any) []protocol.TeamOutputRef {
	refs := outputRefsFromRaw(item, env, payload["output_refs"])
	refs = append(refs, outputRefsFromRaw(item, env, payload["outputs"])...)
	if ref, ok := outputRefFromMap(item, env, payload); ok {
		refs = append(refs, ref)
	}
	return refs
}

func outputRefsFromRaw(item protocol.TeamWorkItem, env protocol.SignalEnvelope, raw any) []protocol.TeamOutputRef {
	values, ok := raw.([]any)
	if !ok {
		return nil
	}
	refs := make([]protocol.TeamOutputRef, 0, len(values))
	for _, value := range values {
		data, ok := value.(map[string]any)
		if !ok {
			continue
		}
		ref, ok := outputRefFromMap(item, env, data)
		if ok {
			refs = append(refs, ref)
		}
	}
	return refs
}

func outputRefFromMap(item protocol.TeamWorkItem, env protocol.SignalEnvelope, data map[string]any) (protocol.TeamOutputRef, bool) {
	outputID := firstNonEmptyString(stringField(data, "output_id"), stringField(data, "id"), stringField(data, "artifact_id"))
	label := firstNonEmptyString(stringField(data, "label"), stringField(data, "title"), stringField(data, "summary"))
	kind := firstNonEmptyString(stringField(data, "kind"), stringField(data, "output_kind"), "output")
	storageRef := teamOutputStorageRefFromMap(kind, data)
	entrypoint := relativeTeamOutputEntrypoint(storageRef, stringField(data, "entrypoint"))
	if outputID == "" && storageRef == "" && entrypoint == "" && label == "" {
		return protocol.TeamOutputRef{}, false
	}
	if outputID == "" {
		outputID = derivedTeamOutputID(item, kind, label, storageRef, entrypoint)
	}
	if label == "" {
		label = firstNonEmptyString(outputID, "Team output")
	}
	ref := protocol.TeamOutputRef{
		OutputID:      outputID,
		TeamID:        firstNonEmptyString(stringField(data, "team_id"), env.Meta.TeamID, item.TeamID),
		WorkItemID:    firstNonEmptyString(stringField(data, "work_item_id"), item.WorkItemID),
		RunID:         firstNonEmptyString(stringField(data, "run_id"), env.Meta.RunID, item.RunID),
		Kind:          kind,
		Label:         label,
		StorageRef:    storageRef,
		Entrypoint:    entrypoint,
		ValidationRef: firstNonEmptyString(stringField(data, "validation_ref"), stringField(data, "validation")),
		ProofRef:      firstNonEmptyString(stringField(data, "proof_ref"), stringField(data, "proof_id"), stringField(data, "proof_artifact_id"), item.ProofID, firstTeamSignalString(item.ProofRefs)),
		ContractID:    firstNonEmptyString(stringField(data, "contract_id"), item.ContractID),
		ProofID:       firstNonEmptyString(stringField(data, "proof_id"), stringField(data, "proof_artifact_id"), item.ProofID),
		AuditRefs:     firstTeamSignalStringSlice(stringSliceField(data, "audit_refs"), item.AuditRefs),
		CreatedAt:     time.Now().UTC(),
	}
	return ref, true
}

func mergeTeamOutputRefs(existing, incoming []protocol.TeamOutputRef) []protocol.TeamOutputRef {
	if len(incoming) == 0 {
		return existing
	}
	merged := make([]protocol.TeamOutputRef, 0, len(existing)+len(incoming))
	seen := map[string]bool{}
	for _, ref := range existing {
		key := teamOutputRefKey(ref)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		merged = append(merged, ref)
	}
	for _, ref := range incoming {
		key := teamOutputRefKey(ref)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		merged = append(merged, ref)
	}
	return merged
}

func teamOutputRefKey(ref protocol.TeamOutputRef) string {
	if strings.TrimSpace(ref.OutputID) != "" {
		return "id:" + strings.TrimSpace(ref.OutputID)
	}
	key := strings.Join([]string{ref.Kind, ref.Label, ref.StorageRef, ref.Entrypoint}, "|")
	key = strings.Trim(key, "| ")
	if key == "" {
		return ""
	}
	return "shape:" + key
}

func derivedTeamOutputID(item protocol.TeamWorkItem, kind, label, storageRef, entrypoint string) string {
	if storageRef == "" && entrypoint == "" && label == "" {
		return uuid.NewString()
	}
	seed := strings.Join([]string{item.TeamID, item.WorkItemID, kind, label, storageRef, entrypoint}, "\x00")
	return fmt.Sprintf("team-output-%s", uuid.NewSHA1(uuid.NameSpaceOID, []byte(seed)).String())
}

func firstTeamSignalString(values []string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func firstTeamSignalStringSlice(candidates ...[]string) []string {
	for _, values := range candidates {
		if len(values) > 0 {
			return values
		}
	}
	return nil
}
