package configdocuments

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/mycelis/core/pkg/protocol"
)

func TestCompileWorkerProfileDocumentUsesEnvelopeLineage(t *testing.T) {
	document := validDocument()
	document.Kind = protocol.ConfigDocumentKindWorkerProfile
	document.Metadata.ID = "review-worker"
	document.Metadata.Name = "Review worker"
	document.Metadata.Version = "2.1.0"
	document.Spec = json.RawMessage(`{
		"role":"reviewer",
		"system_prompt":"Review the assigned output against its acceptance evidence.",
		"capability_refs":["artifact.review"],
		"outputs":["review_report"],
		"verification_strategy":"semantic",
		"verification_rubric":["Every finding identifies evidence"]
	}`)

	result, err := CompileWorkerProfileDocument(document)
	if err != nil {
		t.Fatalf("CompileWorkerProfileDocument() error = %v", err)
	}
	if !result.Ready || result.Snapshot.ID != "review-worker" || result.Snapshot.Version != "2.1.0" {
		t.Fatalf("compile result = %+v", result)
	}
	if !strings.HasPrefix(result.Snapshot.Digest, "sha256:") || result.Profile.Role != "reviewer" {
		t.Fatalf("compile result = %+v", result)
	}
}

func TestCompileDocumentDispatchesWorkerProfile(t *testing.T) {
	document := validDocument()
	document.Kind = protocol.ConfigDocumentKindWorkerProfile
	document.Metadata.ID = "builder-worker"
	document.Spec = json.RawMessage(`{"role":"builder","system_prompt":"Build the approved output.","outputs":["deliverable"]}`)

	compiled, err := CompileDocument(document, protocol.MinimumSufficientBrief{}, protocol.MinimumSufficientBrief{})
	if err != nil {
		t.Fatalf("CompileDocument() error = %v", err)
	}
	result, ok := compiled.(protocol.WorkerProfileCompileResult)
	if !ok || !result.Ready || result.Profile.Outputs[0] != "deliverable" {
		t.Fatalf("CompileDocument() = %#v", compiled)
	}
}
