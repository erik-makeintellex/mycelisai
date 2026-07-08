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
