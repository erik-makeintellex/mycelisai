package swarm

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/mycelis/core/internal/cognitive"
)

func TestHandleGenerateImage_UsesMediaProviderAuthKeyEnv(t *testing.T) {
	t.Setenv("HOSTED_MEDIA_API_KEY", "test-media-key")

	var authHeader string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/images/generations" {
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
		authHeader = r.Header.Get("Authorization")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"data":[{"b64_json":"aW1hZ2U="}]}`))
	}))
	defer server.Close()

	enabled := true
	registry := NewInternalToolRegistry(InternalToolDeps{
		Brain: &cognitive.Router{
			Config: &cognitive.BrainConfig{
				Media: &cognitive.MediaConfig{
					Provider: cognitive.MediaProviderConfig{
						ProviderID: "hosted-media",
						Type:       "openai_hosted",
						Endpoint:   server.URL + "/v1/images/generations",
						ModelID:    "sdxl",
						AuthKeyEnv: "HOSTED_MEDIA_API_KEY",
						Enabled:    &enabled,
					},
				},
			},
		},
	})

	result, err := registry.handleGenerateImage(context.Background(), map[string]any{"prompt": "launch graphic"})
	if err != nil {
		t.Fatalf("handleGenerateImage: %v", err)
	}
	if authHeader != "Bearer test-media-key" {
		t.Fatalf("Authorization header = %q, want bearer token", authHeader)
	}
	if !strings.Contains(result, "Image generated for") {
		t.Fatalf("expected generated image response, got %s", result)
	}
}

func TestWriteSavedImageFile_WritesAndVerifiesReadback(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)

	raw := []byte("png-binary-payload")
	rel, filename, err := writeSavedImageFile("Generated: Hero Concept", "11111111-2222-3333-4444-555555555555", "image/png", "saved-media", "", raw)
	if err != nil {
		t.Fatalf("writeSavedImageFile: %v", err)
	}
	if rel != "saved-media/generated-hero-concept.png" {
		t.Fatalf("rel = %q, want saved-media/generated-hero-concept.png", rel)
	}
	if filename != "generated-hero-concept.png" {
		t.Fatalf("filename = %q, want generated-hero-concept.png", filename)
	}
	diskBytes, err := os.ReadFile(filepath.Join(workspace, filepath.FromSlash(rel)))
	if err != nil {
		t.Fatalf("read saved file: %v", err)
	}
	if string(diskBytes) != string(raw) {
		t.Fatalf("disk bytes = %q, want %q", string(diskBytes), string(raw))
	}
}

func TestHandleSaveCachedImage_PersistsVerifiedDigestAndArtifactMetadata(t *testing.T) {
	workspace := t.TempDir()
	t.Setenv("MYCELIS_WORKSPACE", workspace)

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()

	artifactID := "aaaaaaaa-bbbb-cccc-dddd-eeeeeeeeeeee"
	imgBytes := []byte("verified-image-bytes")
	imgB64 := base64.StdEncoding.EncodeToString(imgBytes)
	wantSHA := fmt.Sprintf("%x", sha256.Sum256(imgBytes))

	mock.ExpectQuery(`SELECT id::text, title, content_type, content FROM artifacts WHERE id = \$1::uuid AND artifact_type = 'image'`).
		WithArgs(artifactID).
		WillReturnRows(sqlmock.NewRows([]string{"id", "title", "content_type", "content"}).
			AddRow(artifactID, "Generated: Launch Banner", "image/png", imgB64))

	mock.ExpectExec(`UPDATE artifacts SET file_path = \$1, file_size_bytes = \$2, metadata = COALESCE\(metadata, '\{\}'::jsonb\) \|\| jsonb_build_object\('saved', true, 'saved_path', \$1, 'saved_at', NOW\(\), 'bytes', \$2, 'sha256', \$3, 'content_type', \$4, 'verified', true\) WHERE id = \$5::uuid`).
		WithArgs("saved-media/launch-banner.png", int64(len(imgBytes)), wantSHA, "image/png", artifactID).
		WillReturnResult(sqlmock.NewResult(0, 1))

	registry := NewInternalToolRegistry(InternalToolDeps{DB: db})
	out, err := registry.handleSaveCachedImage(context.Background(), map[string]any{
		"artifact_id": artifactID,
		"folder":      "saved-media",
		"filename":    "launch-banner.png",
	})
	if err != nil {
		t.Fatalf("handleSaveCachedImage: %v", err)
	}

	var parsed struct {
		Message  string         `json:"message"`
		Artifact map[string]any `json:"artifact"`
	}
	if err := json.Unmarshal([]byte(out), &parsed); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if parsed.Artifact["id"] != artifactID || parsed.Artifact["type"] != "image" {
		t.Fatalf("artifact identity = %#v", parsed.Artifact)
	}
	if parsed.Artifact["saved_path"] != "saved-media/launch-banner.png" || parsed.Artifact["content_type"] != "image/png" {
		t.Fatalf("artifact path/content_type = %#v", parsed.Artifact)
	}
	if int(parsed.Artifact["bytes"].(float64)) != len(imgBytes) || parsed.Artifact["sha256"] != wantSHA || parsed.Artifact["verified"] != true {
		t.Fatalf("artifact proof fields = %#v", parsed.Artifact)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sqlmock expectations: %v", err)
	}
}
