package server

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
)

const groupWorkspaceRoot = "groups"

func assignGroupWorkspaceFolder(g *CollaborationGroup, requested string) error {
	if g == nil {
		return nil
	}
	folder, err := collaborationGroupWorkspaceFolder(g.ID, g.Name, g.TeamIDs, requested)
	if err != nil {
		return err
	}
	g.WorkspaceFolder = folder
	return nil
}

func collaborationGroupWorkspaceFolder(groupID, name string, teamIDs []string, requested string) (string, error) {
	if folder := strings.TrimSpace(requested); folder != "" {
		return normalizeRequestedGroupWorkspaceFolder(folder)
	}
	if len(teamIDs) > 0 {
		if folder := groupWorkspaceFolderForTeamID(teamIDs[0]); folder != "" {
			return folder, nil
		}
	}
	slug := slugID(name)
	if slug == "" {
		slug = "group"
	}
	if short := shortGroupID(groupID); short != "" {
		return groupWorkspaceRoot + "/" + slug + "-" + short, nil
	}
	return groupWorkspaceRoot + "/" + slug, nil
}

func normalizeRequestedGroupWorkspaceFolder(folder string) (string, error) {
	normalized := strings.ReplaceAll(strings.TrimSpace(folder), "\\", "/")
	normalized = strings.TrimPrefix(path.Clean(normalized), "./")
	normalized = strings.Trim(normalized, "/")
	if normalized == "" || normalized == "." {
		return "", fmt.Errorf("workspace_folder is required")
	}
	if strings.HasPrefix(normalized, "../") || strings.Contains(normalized, "/../") || filepath.IsAbs(folder) {
		return "", fmt.Errorf("workspace_folder must stay inside the governed workspace")
	}
	if normalized == groupWorkspaceRoot || strings.HasPrefix(normalized, groupWorkspaceRoot+"/") {
		return normalized, nil
	}
	slug := slugID(normalized)
	if slug == "" {
		return "", fmt.Errorf("workspace_folder must include at least one letter or number")
	}
	return groupWorkspaceRoot + "/" + slug, nil
}

func groupWorkspaceFolderForTeamID(teamID string) string {
	slug := slugID(teamID)
	if slug == "" {
		return ""
	}
	return groupWorkspaceRoot + "/" + slug
}

func shortGroupID(groupID string) string {
	id := strings.TrimSpace(groupID)
	if len(id) >= 8 {
		return id[:8]
	}
	return id
}

func ensureGroupWorkspaceFolder(folder string) error {
	if strings.TrimSpace(folder) == "" {
		return nil
	}
	target, _, err := resolveWorkspacePath(folder, false)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(target, 0o755); err != nil {
		return fmt.Errorf("create group workspace folder: %w", err)
	}
	return nil
}
