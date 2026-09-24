package capabilities

import (
	"context"
	"testing"
)

func TestCountingOptInHasPinnedBindingAndNoLegacyTool(t *testing.T) {
	svc := NewService(Dependencies{CountingEndpoint: "http://127.0.0.1:9123"})
	snapshot, err := svc.Refresh(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	for _, m := range snapshot.Manifests {
		if m.ID != "counting.increment" {
			continue
		}
		if len(m.ToolRefs) != 0 || !m.ApprovalRequired || !m.AuditRequired {
			t.Fatalf("unsafe counting manifest: %+v", m)
		}
		binding := m.Metadata["invocation_binding"].(map[string]any)
		if binding["endpoint"] != "http://127.0.0.1:9123" || binding["unit_cost"] != 1 {
			t.Fatal(binding)
		}
		return
	}
	t.Fatal("counting manifest absent")
}

func TestCountingDefaultsOffAndMalformedEndpointDisabled(t *testing.T) {
	if _, ok := countingManifest(""); ok {
		t.Fatal("fixture enabled by default")
	}
	for _, endpoint := range []string{"file:///tmp/socket", "http://user:secret@host", "http://host?token=value", "http://host/#fragment"} {
		m, ok := countingManifest(endpoint)
		if !ok || m.Status != "disabled" {
			t.Fatalf("unsafe endpoint accepted: %q", endpoint)
		}
		if m.Metadata["invocation_binding"].(map[string]any)["endpoint"] != "" {
			t.Fatal("invalid endpoint retained")
		}
	}
}
