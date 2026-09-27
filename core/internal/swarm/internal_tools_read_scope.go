package swarm

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// workspaceWideReadTeams are the standing system teams whose agents read the
// whole governed workspace: Soma (admin-core) and her council. The ID alone
// grants nothing; the loaded team must also be Core-owned (IsCoreOwnedTeam).
var workspaceWideReadTeams = map[string]struct{}{"admin-core": {}, "council-core": {}}

// confineTeamRead limits a team agent's read_file to groups/<team id> plus the
// team's declared SharedPaths (S7b). absTarget has already passed
// validateToolPath. Calls without an agent team (Core-owned steps, operator-
// approved plans) and system teams stay workspace-wide. Handoff inputs are DB
// rows, not files, so they need no path grant.
func (r *InternalToolRegistry) confineTeamRead(ctx context.Context, rawPath, absTarget string) error {
	inv, ok := ToolInvocationContextFromContext(ctx)
	if !ok {
		return nil
	}
	teamID := strings.TrimSpace(inv.TeamID)
	if teamID == "" {
		if strings.TrimSpace(inv.AgentRole) != "" {
			return fmt.Errorf("read_file: %q is outside this agent's workspace folder", rawPath)
		}
		return nil
	}
	if _, wide := workspaceWideReadTeams[teamID]; wide && r != nil && r.somaRef.IsCoreOwnedTeam(teamID) {
		return nil
	}
	absWorkspace, err := toolWorkspaceRoot()
	if err != nil {
		return err
	}
	realTarget := resolvedPath(absTarget)
	for _, root := range r.teamReadRoots(teamID) {
		absRoot := filepath.Join(absWorkspace, filepath.FromSlash(root))
		if pathWithin(absTarget, absRoot) && pathWithin(realTarget, resolvedPath(absRoot)) {
			return nil
		}
	}
	return fmt.Errorf("read_file: %q is outside this team's workspace folder", rawPath)
}

// teamReadRoots returns groups/<team id> and the team's declared shared
// paths, each workspace-relative and free of traversal.
func (r *InternalToolRegistry) teamReadRoots(teamID string) []string {
	var roots []string
	if safeTeamIDPattern.MatchString(teamID) && !strings.Contains(teamID, "..") {
		roots = append(roots, "groups/"+teamID)
	}
	if r == nil || r.somaRef == nil {
		return roots
	}
	for _, manifest := range r.somaRef.ListTeams() {
		if manifest == nil || manifest.ID != teamID {
			continue
		}
		for _, shared := range manifest.SharedPaths {
			if clean := cleanSharedReadPath(shared); clean != "" {
				roots = append(roots, clean)
			}
		}
	}
	return roots
}

func cleanSharedReadPath(raw string) string {
	normalized := strings.ReplaceAll(strings.TrimSpace(raw), "\\", "/")
	if normalized == "" || strings.HasPrefix(normalized, "/") || filepath.IsAbs(raw) {
		return ""
	}
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(normalized)))
	clean = strings.TrimPrefix(clean, "workspace/")
	if clean == "." || clean == ".." || clean == "workspace" || strings.HasPrefix(clean, "../") {
		return ""
	}
	return clean
}

// resolvedPath follows symlinks when the path exists, else returns it cleaned.
func resolvedPath(path string) string {
	if real, err := filepath.EvalSymlinks(path); err == nil {
		return real
	}
	return filepath.Clean(path)
}

func pathWithin(target, root string) bool {
	rel, err := filepath.Rel(root, target)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

// toolWorkspaceRoot is the absolute governed workspace validateToolPath uses.
func toolWorkspaceRoot() (string, error) {
	workspace := os.Getenv("MYCELIS_WORKSPACE")
	if workspace == "" {
		workspace = "./workspace"
	}
	abs, err := filepath.Abs(workspace)
	if err != nil {
		return "", fmt.Errorf("invalid workspace path: %w", err)
	}
	return abs, nil
}
