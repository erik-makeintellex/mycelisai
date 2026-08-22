package swarm

import "testing"

func TestProjectPackageArtifactFromSuccessfulWriteNormalizesMCPWrite(t *testing.T) {
	artifact, ok := projectPackageArtifactFromSuccessfulWrite(map[string]any{
		"path": "groups/app-team/generated/portal/play.html",
	}, "Deliver a browser application package with a direct entrypoint.")
	if !ok {
		t.Fatal("expected successful generated HTML write to become a package artifact")
	}
	if artifact.Type != "project_package" || artifact.Entrypoint != "groups/app-team/generated/portal/play.html" {
		t.Fatalf("artifact = %+v", artifact)
	}
	if artifact.Folder != "groups/app-team/generated/portal" || len(artifact.Files) != 1 || artifact.Files[0] != "play.html" {
		t.Fatalf("package shape = %+v", artifact)
	}
	if artifact.Validation != "" {
		t.Fatalf("validation = %q, want no claim before readback", artifact.Validation)
	}
}

func TestProjectPackageArtifactFromSuccessfulWriteIgnoresInternalSource(t *testing.T) {
	_, ok := projectPackageArtifactFromSuccessfulWrite(map[string]any{
		"path": "groups/app-team/source/review.html",
	}, "Deliver a browser application package.")
	if ok {
		t.Fatal("internal source write must not become a user deliverable")
	}
}

func TestNormalizeTeamOwnedWriteArgumentsRedirectsPackageOutput(t *testing.T) {
	call := &toolCallPayload{Name: "write_file", Arguments: map[string]any{
		"path": "output/portal/index.html", "content": "<!doctype html>",
	}}

	normalizeTeamOwnedWriteArguments(call, "application-delivery-team-a1b2c", "Deliver a browser application package.")
	autofillToolArguments(call, "Deliver a browser application package.")

	want := "groups/application-delivery-team-a1b2c/generated/portal/index.html"
	if call.Arguments["path"] != want || call.Arguments["package_entrypoint"] != want {
		t.Fatalf("normalized arguments = %#v", call.Arguments)
	}
}

func TestNormalizeTeamOwnedWriteArgumentsRedirectsLegacyTeamOutput(t *testing.T) {
	call := &toolCallPayload{Name: "write_file", Arguments: map[string]any{
		"path":               "groups/application-delivery-team-a1b2c/output/index.html",
		"content":            "<!doctype html>",
		"package_folder":     "output",
		"package_entrypoint": "index.html",
	}}

	normalizeTeamOwnedWriteArguments(call, "application-delivery-team-a1b2c", "Write the requested deliverable.")

	want := "groups/application-delivery-team-a1b2c/generated/index.html"
	if got := stringValue(call.Arguments["path"]); got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
	if _, ok := call.Arguments["package_folder"]; ok {
		t.Fatal("stale package_folder was not removed")
	}
	if _, ok := call.Arguments["package_entrypoint"]; ok {
		t.Fatal("stale package_entrypoint was not removed")
	}
}

func TestNormalizeTeamOwnedWriteArgumentsPreservesInternalTeamPath(t *testing.T) {
	call := &toolCallPayload{Name: "write_file", Arguments: map[string]any{
		"path": "groups/application-delivery-team-a1b2c/planning/review.md",
	}}

	normalizeTeamOwnedWriteArguments(call, "application-delivery-team-a1b2c", "Prepare the internal review.")

	if got := stringValue(call.Arguments["path"]); got != "groups/application-delivery-team-a1b2c/planning/review.md" {
		t.Fatalf("path = %q, want internal path unchanged", got)
	}
}

func TestNormalizeAgentToolCallArgumentsRepairsAliasBeforeTeamOwnership(t *testing.T) {
	call := &toolCallPayload{Name: "write_file", Arguments: map[string]any{
		"file_path": "groups/application-delivery-team-a1b2c/output/first-game/index.html",
		"body":      "<!doctype html>",
	}}

	normalizeAgentToolCallArguments(call, "application-delivery-team-a1b2c", "Deliver a playable browser application.")

	want := "groups/application-delivery-team-a1b2c/generated/first-game/index.html"
	if got := stringValue(call.Arguments["path"]); got != want {
		t.Fatalf("path = %q, want %q", got, want)
	}
	if got := stringValue(call.Arguments["package_entrypoint"]); got != want {
		t.Fatalf("package_entrypoint = %q, want %q", got, want)
	}
}
