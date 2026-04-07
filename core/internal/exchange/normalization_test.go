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
		ServerID:      "srv-123",
		ServerName:    "filesystem",
		ToolName:      "read_file",
		ResultPreview: "Loaded workspace brief.",
		SourceTeam:    "alpha",
		AgentID:       "soma-admin",
		RunID:         "run-42",
	}

	if input.ServerID != "srv-123" || input.SourceTeam != "alpha" || input.AgentID != "soma-admin" || input.RunID != "run-42" {
		t.Fatalf("normalization input lost persistent fields: %+v", input)
	}
}
