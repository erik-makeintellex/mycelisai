package exchange

import "testing"

func TestClassifyMCPOutput(t *testing.T) {
	channel, schema, _ := classifyMCPOutput("fetch", "browser_search")
	if channel != "browser.research.results" || schema != "ToolResult" {
		t.Fatalf("browser classification = %s/%s", channel, schema)
	}

	channel, schema, _ = classifyMCPOutput("media", "image_generate")
	if channel != "media.image.output" || schema != "MediaResult" {
		t.Fatalf("media classification = %s/%s", channel, schema)
	}

	channel, schema, _ = classifyMCPOutput("crm", "list_accounts")
	if channel != "api.data.output" || schema != "ToolResult" {
		t.Fatalf("default classification = %s/%s", channel, schema)
	}
}

func TestMCPNormalizationInput_CarriesPersistentIdentityFields(t *testing.T) {
	input := MCPNormalizationInput{
		ServerID:       "srv-123",
		ServerName:     "filesystem",
		ToolName:       "read_file",
		ResultPreview:  "Loaded workspace brief.",
		SourceTeam:     "alpha",
		AgentID:        "soma-admin",
		RunID:          "run-42",
		RunClass:       "run_linked",
		RetentionClass: "retained",
	}

	if input.ServerID != "srv-123" || input.SourceTeam != "alpha" || input.AgentID != "soma-admin" || input.RunID != "run-42" || input.RunClass != "run_linked" {
		t.Fatalf("normalization input lost persistent fields: %+v", input)
	}
}

func TestOutputProofMetadata_ClassifiesMissingRun(t *testing.T) {
	proof := outputProofMetadata("", "", "", "")

	if proof["run_class"] != "no_run" {
		t.Fatalf("run_class = %v, want no_run", proof["run_class"])
	}
	if proof["no_run_reason"] == "" {
		t.Fatalf("expected no_run_reason, got %+v", proof)
	}
	if proof["retention_class"] != "retained" {
		t.Fatalf("retention_class = %v, want retained", proof["retention_class"])
	}
}

func TestApplyOutputProofPayload_PassesRunAndRetentionFields(t *testing.T) {
	payload := map[string]any{}
	proof := outputProofMetadata("run-42", "", "", "non_retained")

	applyOutputProofPayload(payload, proof)

	if payload["run_id"] != "run-42" || payload["run_class"] != "run_linked" {
		t.Fatalf("run proof payload = %+v", payload)
	}
	if payload["retention_class"] != "non_retained" {
		t.Fatalf("retention_class = %v", payload["retention_class"])
	}
}
