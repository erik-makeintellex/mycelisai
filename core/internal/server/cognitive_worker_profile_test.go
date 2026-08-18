package server

import (
	"slices"
	"testing"
)

func TestWorkerProfileIntentUsesSharedConfigurationTools(t *testing.T) {
	if got := inferReadOnlyConfigToolsFromText("Preview this worker profile"); !slices.Equal(got, []string{"preview_config_document"}) {
		t.Fatalf("read-only tools = %v", got)
	}
	tools, recognized := outcomeTemplateMutationTools("Save this worker profile")
	if !recognized || !slices.Equal(tools, []string{"store_config_document"}) {
		t.Fatalf("mutation tools = %v, recognized = %v", tools, recognized)
	}
	tools, recognized = outcomeTemplateMutationTools("Activate this worker profile")
	if !recognized || !slices.Equal(tools, []string{"activate_config_document"}) {
		t.Fatalf("activation tools = %v, recognized = %v", tools, recognized)
	}
}
