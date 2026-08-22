package swarm

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
)

func TestPreviewConfigDocumentReturnsBoundedReadyWorkIntent(t *testing.T) {
	registry := NewInternalToolRegistry(InternalToolDeps{})
	tool := registry.Get("preview_config_document")
	if tool == nil || tool.Manifest == nil {
		t.Fatal("preview_config_document must be registered with command metadata")
	}

	output, err := tool.Handler(context.Background(), map[string]any{
		"format": "yaml",
		"content": `apiVersion: mycelis.ai/v1
kind: OutcomeTemplate
metadata:
  id: delivery-brief
  name: Delivery brief
  version: "1"
  owner_id: soma
  scope: {kind: workspace, ref: primary}
  enabled: true
  source: {kind: soma, ref: conversation:test}
  governance: {risk_level: low, approval_posture: optional}
spec:
  defaults:
    target_outcome: Produce a usable deliverable
    delivery_form: Reviewable package
    acceptance_evidence: [Package opens successfully]
`,
	})
	if err != nil {
		t.Fatalf("preview_config_document error = %v", err)
	}
	var response map[string]any
	if err := json.Unmarshal([]byte(output), &response); err != nil {
		t.Fatalf("decode tool output: %v", err)
	}
	compiled, ok := response["compiled"].(map[string]any)
	if !ok || compiled["ready"] != true || compiled["work_intent"] == nil {
		t.Fatalf("expected ready WorkIntent preview, got %#v", response)
	}
}

func TestPreviewConfigDocumentRequiresOneSource(t *testing.T) {
	registry := NewInternalToolRegistry(InternalToolDeps{})
	_, err := registry.Get("preview_config_document").Handler(context.Background(), map[string]any{
		"content": "{}",
		"path":    "outcomes/example.yaml",
	})
	if err == nil || !strings.Contains(err.Error(), "exactly one") {
		t.Fatalf("error = %v, want one-source validation", err)
	}
}

func TestPreviewConfigDocumentReturnsWorkerProfileSnapshot(t *testing.T) {
	registry := NewInternalToolRegistry(InternalToolDeps{})
	output, err := registry.Get("preview_config_document").Handler(context.Background(), map[string]any{
		"format": "yaml",
		"content": `apiVersion: mycelis.ai/v1
kind: WorkerProfile
metadata:
  id: source-reviewer
  name: Source reviewer
  version: "1"
  owner_id: soma
  scope: {kind: workspace, ref: primary}
  enabled: true
  source: {kind: soma, ref: conversation:test}
  governance: {risk_level: low, approval_posture: optional}
spec:
  role: reviewer
  system_prompt: Review claims against retained sources.
  outputs: [review_report]
`,
	})
	if err != nil {
		t.Fatalf("preview_config_document error = %v", err)
	}
	var response map[string]any
	if err := json.Unmarshal([]byte(output), &response); err != nil {
		t.Fatalf("decode tool output: %v", err)
	}
	compiled, ok := response["compiled"].(map[string]any)
	if !ok || compiled["ready"] != true {
		t.Fatalf("compiled = %#v", response["compiled"])
	}
	snapshot, ok := compiled["profile_snapshot"].(map[string]any)
	if !ok || snapshot["id"] != "source-reviewer" {
		t.Fatalf("snapshot = %#v", compiled["profile_snapshot"])
	}
}
