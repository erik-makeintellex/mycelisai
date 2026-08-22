package server

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestHandleConfigDocumentDryRunCompilesDirectYAML(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/config-documents/dry-run", s.HandleConfigDocumentDryRun)
	body, err := json.Marshal(map[string]any{
		"format": "yaml",
		"content": `apiVersion: mycelis.ai/v1
kind: OutcomeTemplate
metadata:
  id: product-brief
  name: Product brief
  version: "1"
  owner_id: workspace-owner
  scope:
    kind: workspace
    ref: primary
  enabled: true
  source:
    kind: file
    ref: config/outcomes/product-brief.yaml
  governance:
    risk_level: low
    approval_posture: optional
spec:
  defaults:
    target_outcome: Produce a reviewable product brief
    delivery_form: Markdown document
    acceptance_evidence:
      - Required sections are present
`,
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	rr := doAuthenticatedRequest(t, mux, http.MethodPost, "/api/v1/config-documents/dry-run", string(body))
	assertStatus(t, rr, http.StatusOK)

	var response protocol.APIResponse
	assertJSON(t, rr, &response)
	data, ok := response.Data.(map[string]any)
	if !ok {
		t.Fatalf("expected object payload, got %T", response.Data)
	}
	dryRun, ok := data["dry_run"].(map[string]any)
	if !ok || dryRun["valid"] != true {
		t.Fatalf("expected valid dry run, got %#v", data["dry_run"])
	}
	compiled, ok := data["compiled"].(map[string]any)
	if !ok || compiled["ready"] != true {
		t.Fatalf("expected ready compiled template, got %#v", data["compiled"])
	}
	snapshot, ok := compiled["template_snapshot"].(map[string]any)
	if !ok || snapshot["id"] != "product-brief" {
		t.Fatalf("expected stable template snapshot, got %#v", compiled["template_snapshot"])
	}
	if digest, _ := snapshot["digest"].(string); len(digest) < len("sha256:") || digest[:len("sha256:")] != "sha256:" {
		t.Fatalf("expected canonical digest, got %#v", snapshot["digest"])
	}
}

func TestHandleConfigDocumentDryRunCompilesWorkerProfileYAML(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/config-documents/dry-run", s.HandleConfigDocumentDryRun)
	body, err := json.Marshal(map[string]any{
		"format": "yaml",
		"content": `apiVersion: mycelis.ai/v1
kind: WorkerProfile
metadata:
  id: evidence-reviewer
  name: Evidence reviewer
  version: "1"
  owner_id: workspace-owner
  scope:
    kind: workspace
    ref: primary
  enabled: true
  source:
    kind: file
    ref: config/profiles/evidence-reviewer.yaml
  governance:
    risk_level: low
    approval_posture: optional
spec:
  role: reviewer
  system_prompt: Review the assigned output against its acceptance evidence.
  capability_refs:
    - artifact.review
  outputs:
    - review_report
  verification_strategy: semantic
  verification_rubric:
    - Every finding identifies evidence
`,
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	rr := doAuthenticatedRequest(t, mux, http.MethodPost, "/api/v1/config-documents/dry-run", string(body))
	assertStatus(t, rr, http.StatusOK)
	var response protocol.APIResponse
	assertJSON(t, rr, &response)
	data := response.Data.(map[string]any)
	compiled := data["compiled"].(map[string]any)
	if compiled["ready"] != true {
		t.Fatalf("compiled = %#v, want ready", compiled)
	}
	snapshot := compiled["profile_snapshot"].(map[string]any)
	if snapshot["id"] != "evidence-reviewer" {
		t.Fatalf("snapshot = %#v", snapshot)
	}
	profile := compiled["profile"].(map[string]any)
	if profile["role"] != "reviewer" {
		t.Fatalf("profile = %#v", profile)
	}
}

func TestHandleConfigDocumentDryRunRejectsAmbiguousSource(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/config-documents/dry-run", s.HandleConfigDocumentDryRun)
	rr := doAuthenticatedRequest(t, mux, http.MethodPost, "/api/v1/config-documents/dry-run", `{
		"content":"apiVersion: mycelis.ai/v1",
		"path":"outcomes/product-brief.yaml"
	}`)
	assertStatus(t, rr, http.StatusBadRequest)
}

func TestHandleConfigDocumentDryRunRequiresRootScope(t *testing.T) {
	s := newTestServer()
	mux := setupMux(t, "POST /api/v1/config-documents/dry-run", s.HandleConfigDocumentDryRun)
	rr := doRequest(t, mux, http.MethodPost, "/api/v1/config-documents/dry-run", `{}`)
	assertStatus(t, rr, http.StatusUnauthorized)
}
