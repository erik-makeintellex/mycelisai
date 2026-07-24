package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/artifacts"
	"github.com/mycelis/core/pkg/protocol"
)

func parseLimit(raw string, fallback int) int {
	if fallback <= 0 {
		fallback = 20
	}
	if strings.TrimSpace(raw) == "" {
		return fallback
	}
	var limit int
	if _, err := fmt.Sscanf(strings.TrimSpace(raw), "%d", &limit); err != nil || limit <= 0 {
		return fallback
	}
	return limit
}

func (s *AdminServer) listGroupOutputs(ctx context.Context, group *CollaborationGroup, limit int) ([]artifacts.Artifact, error) {
	return s.listGroupOutputsWithOptions(ctx, group, limit, false)
}

func (s *AdminServer) listGroupOutputsWithOptions(ctx context.Context, group *CollaborationGroup, limit int, includeInternal bool) ([]artifacts.Artifact, error) {
	if s.Artifacts == nil {
		return nil, errors.New("artifacts not initialized")
	}
	if group == nil {
		return []artifacts.Artifact{}, nil
	}

	merged := make(map[uuid.UUID]artifacts.Artifact)
	seen := make(map[string]struct{})
	for _, rawTeamID := range group.TeamIDs {
		teamRef := strings.TrimSpace(rawTeamID)
		if teamRef == "" {
			continue
		}
		var items []artifacts.Artifact
		var err error
		if teamID, parseErr := uuid.Parse(teamRef); parseErr == nil {
			items, err = s.Artifacts.ListByTeam(ctx, teamID, limit)
		} else {
			items, err = s.Artifacts.ListByAgent(ctx, teamRef, limit)
		}
		if err != nil {
			return nil, err
		}
		for _, item := range items {
			if strings.EqualFold(strings.TrimSpace(item.Status), "archived") {
				continue
			}
			if !includeInternal && !isUserFacingGroupOutput(item) {
				continue
			}
			merged[item.ID] = item
			for _, key := range groupArtifactDedupeKeys(item) {
				seen[key] = struct{}{}
			}
		}
		refItems, err := s.listGroupOutputRefs(ctx, teamRef, limit, includeInternal)
		if err != nil {
			return nil, err
		}
		for _, item := range refItems {
			if hasAnyGroupOutputKey(seen, groupArtifactDedupeKeys(item)) {
				continue
			}
			merged[item.ID] = item
			for _, key := range groupArtifactDedupeKeys(item) {
				seen[key] = struct{}{}
			}
		}
	}

	outputs := make([]artifacts.Artifact, 0, len(merged))
	for _, item := range merged {
		outputs = append(outputs, item)
	}
	sort.Slice(outputs, func(i, j int) bool {
		return outputs[i].CreatedAt.After(outputs[j].CreatedAt)
	})
	if limit > 0 && len(outputs) > limit {
		outputs = outputs[:limit]
	}
	return outputs, nil
}

func (s *AdminServer) listGroupOutputRefs(ctx context.Context, teamRef string, limit int, includeInternal bool) ([]artifacts.Artifact, error) {
	items, err := s.listTeamWorkItemsDB(ctx, teamRef, limit, true)
	if err != nil {
		return nil, err
	}
	outputs := make([]artifacts.Artifact, 0)
	for _, item := range items {
		for _, ref := range item.OutputRefs {
			if !includeInternal && !isUserDeliverableTeamOutputRef(ref) {
				continue
			}
			outputs = append(outputs, artifactFromTeamOutputRef(item, ref))
		}
	}
	return outputs, nil
}

func artifactFromTeamOutputRef(item protocol.TeamWorkItem, ref protocol.TeamOutputRef) artifacts.Artifact {
	createdAt := ref.CreatedAt
	if createdAt.IsZero() {
		createdAt = item.UpdatedAt
	}
	if createdAt.IsZero() {
		createdAt = item.CreatedAt
	}
	metadata, _ := json.Marshal(map[string]any{
		"source":         "team_output_ref",
		"output_id":      ref.OutputID,
		"work_item_id":   firstNonEmptyString(ref.WorkItemID, item.WorkItemID),
		"run_id":         firstNonEmptyString(ref.RunID, item.RunID),
		"output_class":   string(outputClassForTeamRef(ref)),
		"entrypoint":     ref.Entrypoint,
		"folder":         projectPackageFolder(ref),
		"validation_ref": ref.ValidationRef,
		"proof_ref":      ref.ProofRef,
		"contract_id":    firstNonEmptyString(ref.ContractID, item.ContractID),
		"proof_id":       firstNonEmptyString(ref.ProofID, item.ProofID),
		"audit_refs":     ref.AuditRefs,
	})
	return artifacts.Artifact{
		ID:           stableGroupOutputRefID(item, ref),
		AgentID:      firstNonEmptyString(ref.TeamID, item.TeamID),
		ArtifactType: artifactTypeForTeamOutputRef(ref),
		Title:        firstNonEmptyString(ref.Label, ref.StorageRef, ref.OutputID, item.Objective),
		ContentType:  contentTypeForTeamOutputRef(ref),
		FilePath:     firstNonEmptyString(ref.StorageRef, ref.Entrypoint),
		Metadata:     metadata,
		Status:       "approved",
		CreatedAt:    createdAt,
	}
}

func projectPackageFolder(ref protocol.TeamOutputRef) string {
	if artifactTypeForTeamOutputRef(ref) != artifacts.TypeProjectPackage {
		return ""
	}
	return strings.TrimSpace(ref.StorageRef)
}

func stableGroupOutputRefID(item protocol.TeamWorkItem, ref protocol.TeamOutputRef) uuid.UUID {
	key := strings.Join([]string{
		"group-output-ref",
		item.TeamID,
		item.WorkItemID,
		ref.OutputID,
		ref.StorageRef,
		ref.Label,
	}, ":")
	return uuid.NewSHA1(uuid.NameSpaceOID, []byte(key))
}

func artifactTypeForTeamOutputRef(ref protocol.TeamOutputRef) artifacts.ArtifactType {
	switch strings.ToLower(strings.TrimSpace(ref.Kind)) {
	case "project_package", "package", "app":
		return artifacts.TypeProjectPackage
	case "image", "media_image":
		return artifacts.TypeImage
	case "audio":
		return artifacts.TypeAudio
	case "data", "dataset", "json", "csv":
		return artifacts.TypeData
	case "code", "script":
		return artifacts.TypeCode
	case "document", "markdown", "text", "text_reply":
		return artifacts.TypeDocument
	default:
		return artifacts.TypeFile
	}
}

func contentTypeForTeamOutputRef(ref protocol.TeamOutputRef) string {
	switch artifactTypeForTeamOutputRef(ref) {
	case artifacts.TypeProjectPackage:
		return "application/vnd.mycelis.project-package+json"
	case artifacts.TypeImage:
		return "image/*"
	case artifacts.TypeAudio:
		return "audio/*"
	case artifacts.TypeData:
		return "application/json"
	case artifacts.TypeCode:
		return "text/plain"
	default:
		return "text/markdown"
	}
}

func groupArtifactDedupeKeys(item artifacts.Artifact) []string {
	keys := make([]string, 0, 3)
	for _, value := range []string{item.FilePath, item.Title, artifactMetadataString(item.Metadata, "output_id")} {
		cleaned := strings.ToLower(strings.TrimSpace(value))
		if cleaned != "" {
			keys = append(keys, cleaned)
		}
	}
	return keys
}

func hasAnyGroupOutputKey(seen map[string]struct{}, keys []string) bool {
	for _, key := range keys {
		if _, ok := seen[key]; ok {
			return true
		}
	}
	return false
}

func isUserFacingGroupOutput(item artifacts.Artifact) bool {
	return groupOutputClass(item) == string(protocol.OutputClassUserDeliverable)
}

func groupOutputClass(item artifacts.Artifact) string {
	if class := protocol.NormalizeOutputClass(artifactMetadataString(item.Metadata, "output_class")); class != "" {
		return string(class)
	}
	return string(protocol.InferOutputClass(string(item.ArtifactType), item.FilePath, item.Title))
}

func artifactMetadataString(raw json.RawMessage, key string) string {
	if len(raw) == 0 {
		return ""
	}
	var metadata map[string]any
	if err := json.Unmarshal(raw, &metadata); err != nil {
		return ""
	}
	value, ok := metadata[key]
	if !ok {
		return ""
	}
	text, ok := value.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(text)
}
