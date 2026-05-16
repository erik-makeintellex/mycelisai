package swarm

import "testing"

func TestArtifactResultPayload_NormalizesProjectPackageMetadataAliases(t *testing.T) {
	payload := artifactResultPayload(
		"artifact-game",
		"file",
		"Coin Runner Game",
		"text/plain",
		"large manifest body",
		map[string]any{
			"package_kind":       "project_package",
			"package_entrypoint": "workspace/generated/coin-runner/index.html",
			"package_folder":     "workspace/generated/coin-runner",
			"package_files":      `["index.html","game.js","styles.css"]`,
			"validation_summary": "Browser opened and score increased after click.",
		},
	)

	artifact, ok := payload["artifact"].(map[string]any)
	if !ok {
		t.Fatalf("artifact payload = %#v", payload["artifact"])
	}
	if artifact["type"] != "project_package" {
		t.Fatalf("type = %#v", artifact["type"])
	}
	if artifact["entrypoint"] != "workspace/generated/coin-runner/index.html" {
		t.Fatalf("entrypoint = %#v", artifact["entrypoint"])
	}
	if artifact["folder"] != "workspace/generated/coin-runner" {
		t.Fatalf("folder = %#v", artifact["folder"])
	}
	files, ok := artifact["files"].([]string)
	if !ok {
		t.Fatalf("files = %#v", artifact["files"])
	}
	if len(files) != 3 || files[1] != "game.js" {
		t.Fatalf("files = %#v", files)
	}
	if artifact["validation"] != "Browser opened and score increased after click." {
		t.Fatalf("validation = %#v", artifact["validation"])
	}
}
