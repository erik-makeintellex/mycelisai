package swarm

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEnsureRuntimeTeamWorkspaceCreatesIsolatedDeliveryFolders(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)

	relative, err := ensureRuntimeTeamWorkspace("delivery-team-abc12")
	if err != nil {
		t.Fatalf("ensure team workspace: %v", err)
	}
	if relative != "groups/delivery-team-abc12" {
		t.Fatalf("relative workspace = %q", relative)
	}
	for _, name := range []string{"planning", "source", "generated"} {
		info, statErr := os.Stat(filepath.Join(workspace, relative, name))
		if statErr != nil || !info.IsDir() {
			t.Fatalf("%s folder was not created: %v", name, statErr)
		}
	}
}

func TestEnsureRuntimeTeamWorkspaceRejectsNestedTeamID(t *testing.T) {
	t.Setenv("MYCELIS_WORKSPACE", t.TempDir())
	if _, err := ensureRuntimeTeamWorkspace("other-team/escape"); err == nil {
		t.Fatal("nested team ID must not define another workspace path")
	}
}
