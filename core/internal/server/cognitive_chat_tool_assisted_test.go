package server

import (
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestChatResponseTools_PreservesReadOnlyTools(t *testing.T) {
	got := chatResponseTools(false, []string{"web_search", "web_search"}, nil)
	if len(got) != 1 || got[0] != "web_search" {
		t.Fatalf("tools = %v, want [web_search]", got)
	}
}

func TestBuildDirectChatExecutionSummary_MarksToolsAndArtifactsAssisted(t *testing.T) {
	summary := buildDirectChatExecutionSummary("Summarize status", "Done.", "audit-1", chatAgentResult{
		ToolsUsed: []string{"web_search"},
		Artifacts: []protocol.ChatArtifactRef{{
			ID:    "artifact-1",
			Type:  "document",
			Title: "Status brief",
		}},
	})

	if summary.Execution.Shape != protocol.ExecutionShapeToolAssistedWork {
		t.Fatalf("execution.shape = %q", summary.Execution.Shape)
	}
	if summary.Execution.Status != protocol.ExecutionStatusCompleted {
		t.Fatalf("execution.status = %q", summary.Execution.Status)
	}
	if len(summary.CapabilityUse) != 1 || summary.CapabilityUse[0].ID != "web_search" {
		t.Fatalf("capability_use = %+v", summary.CapabilityUse)
	}
	if summary.Proof.RunClass != protocol.ExecutionRunClassNoRun || summary.Proof.NoRunReason == "" {
		t.Fatalf("proof no-run classification = %+v", summary.Proof)
	}
	if len(summary.Outputs) == 0 || summary.Outputs[0].RetentionClass != protocol.ExecutionRetentionClassNonRetained {
		t.Fatalf("answer output retention = %+v", summary.Outputs)
	}
}

func TestBuildDirectChatExecutionSummary_RetainsProjectPackageArtifact(t *testing.T) {
	summary := buildDirectChatExecutionSummary("Build a playable game", "Game ready.", "audit-game", chatAgentResult{
		ToolsUsed: []string{"store_artifact"},
		Artifacts: []protocol.ChatArtifactRef{{
			ID:          "pkg-1",
			Type:        "project_package",
			Title:       "Coin Runner",
			ContentType: "application/vnd.mycelis.project+json",
			Entrypoint:  "workspace/generated/coin-runner/index.html",
			Folder:      "workspace/generated/coin-runner",
			Files:       []string{"index.html", "game.js", "styles.css"},
			Validation:  "Opened in browser and score increased after click.",
		}},
	})

	var pkg *protocol.ExecutionOutput
	for i := range summary.Outputs {
		if summary.Outputs[i].Kind == "project_package" {
			pkg = &summary.Outputs[i]
			break
		}
	}
	if pkg == nil {
		t.Fatalf("outputs = %+v, want project_package output", summary.Outputs)
	}
	if pkg.Href != "/api/v1/workspace/files/view?path=workspace%2Fgenerated%2Fcoin-runner%2Findex.html" {
		t.Fatalf("href = %q", pkg.Href)
	}
	if pkg.Retained == nil || !*pkg.Retained || pkg.RetentionClass != protocol.ExecutionRetentionClassRetained {
		t.Fatalf("retention = retained:%v class:%q", pkg.Retained, pkg.RetentionClass)
	}
	if pkg.Entrypoint != "workspace/generated/coin-runner/index.html" || pkg.Folder != "workspace/generated/coin-runner" {
		t.Fatalf("package paths = %+v", pkg)
	}
	if len(pkg.Files) != 3 || pkg.Files[2] != "styles.css" {
		t.Fatalf("files = %+v", pkg.Files)
	}
	if pkg.Validation != "Opened in browser and score increased after click." {
		t.Fatalf("validation = %q", pkg.Validation)
	}
}

func TestBuildConfirmActionFailureExecutionSummary_DescribesDegradation(t *testing.T) {
	summary := buildConfirmActionFailureExecutionSummary("proof-1", "contract-1", "artifact-1", "run-1", "audit-1", errors.New("tool unavailable"))

	if summary.Execution.Status != protocol.ExecutionStatusFailed {
		t.Fatalf("execution.status = %q", summary.Execution.Status)
	}
	if summary.Proof.RunID != "run-1" || summary.Proof.Verified == nil || *summary.Proof.Verified {
		t.Fatalf("proof = %+v", summary.Proof)
	}
	degradation := summary.AuditRecovery.Degradation
	if degradation == nil {
		t.Fatal("expected degradation metadata")
	}
	if degradation.Code != "approved_execution_failed" || !degradation.RequiresAttention {
		t.Fatalf("degradation = %+v", degradation)
	}
	if degradation.TrustedState == "" || degradation.InvalidatedProof == "" || degradation.SafeContinuation == "" {
		t.Fatalf("degradation trust boundaries incomplete: %+v", degradation)
	}
	if summary.NextStep == nil || summary.NextStep.Href != "/api/v1/runs/run-1" {
		t.Fatalf("next step = %+v", summary.NextStep)
	}
}

func TestBuildConfirmActionFailureExecutionSummary_SkipsRunLinkWithoutRunID(t *testing.T) {
	summary := buildConfirmActionFailureExecutionSummary("proof-1", "contract-1", "artifact-1", "", "audit-1", errors.New("tool unavailable"))

	if summary.Proof.RunClass != protocol.ExecutionRunClassNoRun || summary.Proof.NoRunReason == "" {
		t.Fatalf("proof = %+v", summary.Proof)
	}
	if summary.NextStep != nil {
		t.Fatalf("next step = %+v, want nil without run id", summary.NextStep)
	}
}

func decodeChatPayloadFromAPIResponse(t *testing.T, rr *httptest.ResponseRecorder) protocol.ChatResponsePayload {
	t.Helper()
	var resp protocol.APIResponse
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	data, err := json.Marshal(resp.Data)
	if err != nil {
		t.Fatalf("marshal response data: %v", err)
	}
	var envelope protocol.CTSEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("decode envelope: %v", err)
	}
	var payload protocol.ChatResponsePayload
	if err := json.Unmarshal(envelope.Payload, &payload); err != nil {
		t.Fatalf("decode payload: %v", err)
	}
	return payload
}
