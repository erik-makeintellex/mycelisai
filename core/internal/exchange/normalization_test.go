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
