package server

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
)

func TestHandleDeploymentTrustUnknownsDoNotExposeSecrets(t *testing.T) {
	t.Setenv("MYCELIS_IMAGE_TAG", "")
	t.Setenv("MYCELIS_DEPLOYMENT_LANE", "")
	t.Setenv("MYCELIS_PROOF_LANE", "")
	t.Setenv("MYCELIS_ARTIFACT_ROOT", "")
	t.Setenv("MYCELIS_ARTIFACTS_ROOT", "")
	t.Setenv("DATA_DIR", "")
	t.Setenv("MYCELIS_BREAK_GLASS_API_KEY", "super-secret")

	s := newTestServer()
	mux := setupMux(t, "GET /api/v1/system/deployments/trust", s.HandleDeploymentTrust)
	rr := doRequest(t, mux, http.MethodGet, "/api/v1/system/deployments/trust", "")

	assertStatus(t, rr, http.StatusOK)
	body := rr.Body.String()
	if body == "" || json.Valid([]byte(body)) == false {
		t.Fatalf("expected JSON body, got %s", body)
	}
	if strings.Contains(body, "super-secret") {
		t.Fatalf("response exposed secret material: %s", body)
	}

	var payload struct {
		OK   bool                    `json:"ok"`
		Data DeploymentTrustSnapshot `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if !payload.OK {
		t.Fatalf("expected ok=true")
	}
	if payload.Data.ImageTag != unknownTrustValue {
		t.Fatalf("image_tag = %q, want unknown", payload.Data.ImageTag)
	}
	if payload.Data.ArtifactRoot != unknownTrustValue {
		t.Fatalf("artifact_root = %q, want unknown", payload.Data.ArtifactRoot)
	}
	if payload.Data.RuntimeHealth.Total == 0 {
		t.Fatalf("expected runtime health services")
	}
}

func TestHandleDeploymentTrustUsesDataDirArtifactFallback(t *testing.T) {
	t.Setenv("MYCELIS_ARTIFACT_ROOT", "")
	t.Setenv("MYCELIS_ARTIFACTS_ROOT", "")
	t.Setenv("DATA_DIR", "/data/artifacts")

	s := newTestServer()
	mux := setupMux(t, "GET /api/v1/system/deployments/trust", s.HandleDeploymentTrust)
	rr := doRequest(t, mux, http.MethodGet, "/api/v1/system/deployments/trust", "")

	assertStatus(t, rr, http.StatusOK)
	var payload struct {
		Data DeploymentTrustSnapshot `json:"data"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	if payload.Data.ArtifactRoot != "/data/artifacts" {
		t.Fatalf("artifact_root = %q, want /data/artifacts", payload.Data.ArtifactRoot)
	}
}

func TestSummarizeRuntimeHealth(t *testing.T) {
	summary := summarizeRuntimeHealth([]ServiceStatus{
		{Name: "nats", Status: "online"},
		{Name: "postgres", Status: "degraded"},
		{Name: "scheduler", Status: "offline"},
	})

	if summary.Status != "offline" {
		t.Fatalf("status = %q, want offline", summary.Status)
	}
	if summary.Online != 1 || summary.Degraded != 1 || summary.Offline != 1 || summary.Total != 3 {
		t.Fatalf("unexpected counts: %+v", summary)
	}
}
