package server

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/mycelis/core/pkg/protocol"
)

type clearGroupRequest struct {
	IncludeOutputs bool `json:"include_outputs"`
}

type clearGroupResult struct {
	Group               *CollaborationGroup `json:"group"`
	OutputsCleared      bool                `json:"outputs_cleared"`
	WorkspaceFolder     string              `json:"workspace_folder,omitempty"`
	WorkspaceRemoved    bool                `json:"workspace_removed"`
	ArtifactsArchived   int64               `json:"artifacts_archived"`
	OperatorDescription string              `json:"operator_description"`
	Warnings            []string            `json:"warnings,omitempty"`
}

// POST /api/v1/groups/{id}/clear
func (s *AdminServer) HandleClearGroup(w http.ResponseWriter, r *http.Request) {
	if _, ok := requireRootAdminScope(w, r, "groups:write"); !ok {
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		respondAPIError(w, "Missing group ID", http.StatusBadRequest)
		return
	}
	var req clearGroupRequest
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			respondAPIError(w, "Invalid JSON body", http.StatusBadRequest)
			return
		}
	}
	group, err := s.getGroupDB(r.Context(), id)
	if errors.Is(err, sql.ErrNoRows) || group == nil {
		respondAPIError(w, "Group not found", http.StatusNotFound)
		return
	}
	if err != nil {
		respondAPIError(w, "Failed to load group: "+err.Error(), http.StatusServiceUnavailable)
		return
	}
	result, clearErr := s.clearGroup(r.Context(), group, req.IncludeOutputs)
	if clearErr != nil {
		respondAPIError(w, clearErr.Error(), http.StatusServiceUnavailable)
		return
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(result))
}

func (s *AdminServer) clearGroup(ctx context.Context, group *CollaborationGroup, includeOutputs bool) (clearGroupResult, error) {
	auditEventID, _ := s.createAuditEvent(
		protocol.TemplateChatToAnswer,
		"groups.clear",
		"Cleared collaboration group",
		map[string]any{
			"group_id":        group.ID,
			"include_outputs": includeOutputs,
			"workspace_folder": strings.TrimSpace(
				group.WorkspaceFolder,
			),
		},
	)
	archivedGroup, err := s.updateGroupStatusDB(ctx, group.ID, groupStatusArchived, auditEventID)
	if err != nil {
		return clearGroupResult{}, err
	}
	result := clearGroupResult{
		Group: archivedGroup,
		WorkspaceFolder: strings.TrimSpace(
			group.WorkspaceFolder,
		),
		OperatorDescription: "Group cleared from active lanes. Message-bus handoff data is transient; retained output files were kept.",
	}
	if !includeOutputs {
		return result, nil
	}
	result.OutputsCleared = true
	result.OperatorDescription = "Group cleared from active lanes and retained output files were removed from the group workspace."
	if removed, removeErr := removeGroupWorkspaceFolder(group.WorkspaceFolder); removeErr != nil {
		result.Warnings = append(result.Warnings, removeErr.Error())
	} else {
		result.WorkspaceRemoved = removed
	}
	if archived, archiveErr := s.archiveGroupOutputArtifactsDB(ctx, group.TeamIDs); archiveErr != nil {
		result.Warnings = append(result.Warnings, archiveErr.Error())
	} else {
		result.ArtifactsArchived = archived
	}
	return result, nil
}

func removeGroupWorkspaceFolder(folder string) (bool, error) {
	normalized, err := normalizeRequestedGroupWorkspaceFolder(folder)
	if err != nil {
		return false, err
	}
	if normalized == groupWorkspaceRoot || !strings.HasPrefix(normalized, groupWorkspaceRoot+"/") {
		return false, errors.New("group output cleanup must target a group workspace folder")
	}
	target, rel, err := resolveWorkspacePath(normalized, false)
	if err != nil {
		return false, err
	}
	if rel == "." || rel == groupWorkspaceRoot {
		return false, errors.New("group output cleanup cannot remove the workspace root")
	}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return false, nil
	}
	if err := os.RemoveAll(target); err != nil {
		return false, err
	}
	return true, nil
}

func (s *AdminServer) archiveGroupOutputArtifactsDB(ctx context.Context, teamIDs []string) (int64, error) {
	db := s.getDB()
	if db == nil {
		return 0, errors.New("database not available")
	}
	var total int64
	for _, rawTeamID := range teamIDs {
		teamRef := strings.TrimSpace(rawTeamID)
		if teamRef == "" {
			continue
		}
		var res sql.Result
		var err error
		if teamID, parseErr := uuid.Parse(teamRef); parseErr == nil {
			res, err = db.ExecContext(ctx, `UPDATE artifacts SET status=$1 WHERE team_id=$2`, "archived", teamID)
		} else {
			res, err = db.ExecContext(ctx, `UPDATE artifacts SET status=$1 WHERE agent_id=$2`, "archived", teamRef)
		}
		if err != nil {
			return total, err
		}
		rows, _ := res.RowsAffected()
		total += rows
	}
	return total, nil
}
