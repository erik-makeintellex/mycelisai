package swarm

import (
	"reflect"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestExtractToolOutputArtifacts_SupportsSingleArtifact(t *testing.T) {
	toolResult := `{"message":"stored output","artifact":{"id":"a1","type":"document","title":"Plan"}}`

	message, artifacts, ok := extractToolOutputArtifacts(toolResult)
	if !ok {
		t.Fatal("expected structured tool output to be detected")
	}
	if message != "stored output" {
		t.Fatalf("message = %q, want %q", message, "stored output")
	}
	if len(artifacts) != 1 {
		t.Fatalf("artifacts len = %d, want 1", len(artifacts))
	}
	if artifacts[0].ID != "a1" || artifacts[0].Type != "document" {
		t.Fatalf("unexpected artifact: %+v", artifacts[0])
	}
}

func TestExtractToolOutputArtifacts_SupportsArtifactArrays(t *testing.T) {
	toolResult := `{"message":"creative samples ready","artifacts":[{"id":"img-1","type":"image","title":"Moodboard"},{"id":"doc-1","type":"document","title":"Concept Notes"}]}`

	message, artifacts, ok := extractToolOutputArtifacts(toolResult)
	if !ok {
		t.Fatal("expected artifact array to be detected")
	}
	if message != "creative samples ready" {
		t.Fatalf("message = %q, want %q", message, "creative samples ready")
	}
	if len(artifacts) != 2 {
		t.Fatalf("artifacts len = %d, want 2", len(artifacts))
	}
	if artifacts[0].Type != "image" || artifacts[1].Type != "document" {
		t.Fatalf("unexpected artifacts: %+v", artifacts)
	}
}

func TestExtractToolOutputArtifacts_IgnoresPlainText(t *testing.T) {
	message, artifacts, ok := extractToolOutputArtifacts("plain council response")
	if ok {
		t.Fatal("did not expect plain text to parse as structured tool output")
	}
	if message != "plain council response" {
		t.Fatalf("message = %q", message)
	}
	if artifacts != nil {
		t.Fatalf("artifacts = %+v, want nil", artifacts)
	}
}

func TestExtractToolOutputArtifacts_PreservesArtifactShape(t *testing.T) {
	toolResult := `{"message":"visual ready","artifacts":[{"id":"img-1","type":"image","title":"Key Art","content_type":"image/png","cached":true,"expires_at":"2026-03-22T12:00:00Z"}]}`

	_, artifacts, ok := extractToolOutputArtifacts(toolResult)
	if !ok {
		t.Fatal("expected structured image artifact")
	}

	want := protocol.ChatArtifactRef{
		ID:          "img-1",
		Type:        "image",
		Title:       "Key Art",
		ContentType: "image/png",
		Cached:      true,
		ExpiresAt:   "2026-03-22T12:00:00Z",
	}
	if !reflect.DeepEqual(artifacts[0], want) {
		t.Fatalf("artifact = %+v, want %+v", artifacts[0], want)
	}
}

func TestExtractToolOutputArtifacts_PreservesProjectPackageShape(t *testing.T) {
	toolResult := `{"message":"game ready","artifact":{"id":"pkg-1","type":"project_package","title":"Coin Runner","content_type":"application/vnd.mycelis.project+json","entrypoint":"workspace/generated/coin-runner/index.html","folder":"workspace/generated/coin-runner","files":["index.html","game.js","styles.css"],"validation":"Opened in browser and score increased after click."}}`

	message, artifacts, ok := extractToolOutputArtifacts(toolResult)
	if !ok {
		t.Fatal("expected structured project package artifact")
	}
	if message != "game ready" {
		t.Fatalf("message = %q, want %q", message, "game ready")
	}
	if len(artifacts) != 1 {
		t.Fatalf("artifacts len = %d, want 1", len(artifacts))
	}

	got := artifacts[0]
	if got.ID != "pkg-1" || got.Type != "project_package" || got.Title != "Coin Runner" {
		t.Fatalf("artifact identity = %+v", got)
	}
	if got.ContentType != "application/vnd.mycelis.project+json" {
		t.Fatalf("content_type = %q", got.ContentType)
	}
	if got.Entrypoint != "workspace/generated/coin-runner/index.html" {
		t.Fatalf("entrypoint = %q", got.Entrypoint)
	}
	if got.Folder != "workspace/generated/coin-runner" {
		t.Fatalf("folder = %q", got.Folder)
	}
	if len(got.Files) != 3 || got.Files[1] != "game.js" {
		t.Fatalf("files = %+v", got.Files)
	}
	if got.Validation != "Opened in browser and score increased after click." {
		t.Fatalf("validation = %q", got.Validation)
	}
}
