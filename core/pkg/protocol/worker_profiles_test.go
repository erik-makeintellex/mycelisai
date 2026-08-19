package protocol

import (
	"encoding/json"
	"testing"
)

func validWorkerProfileDocument() ConfigDocument {
	document := validConfigDocument()
	document.Kind = ConfigDocumentKindWorkerProfile
	document.Metadata.ID = "research-worker"
	document.Metadata.Name = "Research worker"
	document.Metadata.Source.Ref = "profiles/research-worker.yaml"
	document.Spec = json.RawMessage(`{
		"description":"Researches approved sources and returns concise evidence.",
		"role":"researcher",
		"system_prompt":"Research the assigned question and cite the retained evidence.",
		"model":"local-default",
		"capability_refs":["web_search","artifact.review"],
		"context_bindings":[{"kind":"public_web","access":"search"}],
		"usage_policy":{"selection":"soma_or_manual","scope":"workspace"},
		"inputs":["research_question"],
		"outputs":["research_summary","source_refs"],
		"verification_strategy":"semantic",
		"verification_rubric":["Claims identify their evidence"]
	}`)
	return document
}

func TestValidateConfigDocumentAcceptsWorkerProfile(t *testing.T) {
	document := validWorkerProfileDocument()
	if issues := ValidateConfigDocument(document); len(issues) != 0 {
		t.Fatalf("ValidateConfigDocument() issues = %+v, want none", issues)
	}
	result := DryRunConfigDocument(document)
	if !result.Valid || result.Digest == "" || result.Effect.Action != ConfigDocumentEffectActivate {
		t.Fatalf("DryRunConfigDocument() = %+v, want valid activation preview", result)
	}
}

func TestValidateWorkerProfileSpecFailsClosed(t *testing.T) {
	tests := []struct {
		name string
		spec string
		code string
	}{
		{name: "unknown field", spec: `{"role":"researcher","system_prompt":"Research.","surprise":true}`, code: "worker_profile.invalid_spec"},
		{name: "missing role", spec: `{"system_prompt":"Research."}`, code: "worker_profile.missing_role"},
		{name: "missing instructions", spec: `{"role":"researcher"}`, code: "worker_profile.missing_system_prompt"},
		{name: "duplicate capability", spec: `{"role":"researcher","system_prompt":"Research.","capability_refs":["web_search","web_search"]}`, code: "worker_profile.duplicate_capability_refs"},
		{name: "unsupported selection", spec: `{"role":"researcher","system_prompt":"Research.","usage_policy":{"selection":"always"}}`, code: "worker_profile.unsupported_selection"},
		{name: "semantic rubric required", spec: `{"role":"researcher","system_prompt":"Research.","verification_strategy":"semantic"}`, code: "worker_profile.missing_verification_rubric"},
		{name: "empirical command required", spec: `{"role":"builder","system_prompt":"Build.","verification_strategy":"empirical"}`, code: "worker_profile.missing_validation_command"},
		{name: "command only empirical", spec: `{"role":"builder","system_prompt":"Build.","validation_command":"go test ./..."}`, code: "worker_profile.unexpected_validation_command"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			document := validWorkerProfileDocument()
			document.Spec = json.RawMessage(tt.spec)
			issues := ValidateConfigDocument(document)
			if !hasConfigDocumentIssue(issues, tt.code) {
				t.Fatalf("issues = %+v, want %q", issues, tt.code)
			}
		})
	}
}

func TestValidateWorkerProfileSpecAllowsTeamUsageScope(t *testing.T) {
	document := validWorkerProfileDocument()
	document.Spec = json.RawMessage(`{"role":"builder","system_prompt":"Build the approved work.","usage_policy":{"selection":"soma_or_manual","scope":"team"}}`)
	if issues := ValidateWorkerProfileSpec(document.Spec); len(issues) != 0 {
		t.Fatalf("ValidateWorkerProfileSpec() issues = %#v", issues)
	}
}
