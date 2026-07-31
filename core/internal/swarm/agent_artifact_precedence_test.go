package swarm

import (
	"reflect"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestDedupeAgentArtifactsPreservesValidProjectPackageOverLaterIncompleteArtifact(t *testing.T) {
	valid := completeProjectPackageArtifact("write-result")
	incomplete := protocol.ChatArtifactRef{
		ID:      "store-result",
		Type:    "project_package",
		Title:   valid.Title,
		Content: "output/",
	}

	got := dedupeAgentArtifacts([]protocol.ChatArtifactRef{valid, incomplete})

	if len(got) != 1 {
		t.Fatalf("deduped artifacts = %#v, want one logical project package", got)
	}
	if !reflect.DeepEqual(got[0], valid) {
		t.Fatalf("retained artifact = %#v, want complete package %#v", got[0], valid)
	}
}

func TestDedupeAgentArtifactsPromotesValidProjectPackageOverEarlierIncompleteArtifact(t *testing.T) {
	valid := completeProjectPackageArtifact("write-result")
	incomplete := protocol.ChatArtifactRef{
		ID:        "store-result",
		Type:      "project_package",
		Title:     valid.Title,
		Content:   "output/",
		SavedPath: valid.Folder,
	}

	got := dedupeAgentArtifacts([]protocol.ChatArtifactRef{incomplete, valid})

	if len(got) != 1 {
		t.Fatalf("deduped artifacts = %#v, want one logical project package", got)
	}
	if !reflect.DeepEqual(got[0], valid) {
		t.Fatalf("retained artifact = %#v, want complete package %#v", got[0], valid)
	}
}

func TestDedupeAgentArtifactsConvergesDuplicateValidProjectPackagesDeterministically(t *testing.T) {
	first := completeProjectPackageArtifact("write-result")
	duplicate := first
	duplicate.ID = "store-result"
	duplicate.Content = `{"kind":"project_package","folder":"groups/app-team/generated/portal","entrypoint":"groups/app-team/generated/portal/index.html"}`

	got := dedupeAgentArtifacts([]protocol.ChatArtifactRef{first, duplicate, first})

	if len(got) != 1 {
		t.Fatalf("deduped artifacts = %#v, want one logical project package", got)
	}
	if !reflect.DeepEqual(got[0], first) {
		t.Fatalf("retained artifact = %#v, want first complete package %#v", got[0], first)
	}
}

func completeProjectPackageArtifact(id string) protocol.ChatArtifactRef {
	return protocol.ChatArtifactRef{
		ID:          id,
		Type:        "project_package",
		Title:       "Portal",
		ContentType: "application/vnd.mycelis.project+json",
		Content:     `{"kind":"project_package"}`,
		SavedPath:   "groups/app-team/generated/portal",
		Entrypoint:  "groups/app-team/generated/portal/index.html",
		Folder:      "groups/app-team/generated/portal",
		Files:       []string{"index.html", "README.md", "PROOF.md", "project-package.json"},
		Validation:  "Entrypoint opened and primary interaction passed.",
	}
}
