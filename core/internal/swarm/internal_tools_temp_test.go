package swarm

import (
	"strings"
	"testing"
)

func TestIsLeadAgent(t *testing.T) {
	r := &InternalToolRegistry{}
	cases := []struct {
		agentID string
		teamID  string
		role    string
		want    bool
	}{
		{"admin", "admin-core", "orchestrator", true},
		{"council-architect", "council-core", "architect", true},
		{"builder-1", "team-x", "lead_engineer", true},
		{"worker-1", "team-y", "assistant", false},
	}
	for _, tc := range cases {
		got := r.isLeadAgent(tc.agentID, tc.teamID, tc.role)
		if got != tc.want {
			t.Fatalf("isLeadAgent(%q,%q,%q)=%v want %v", tc.agentID, tc.teamID, tc.role, got, tc.want)
		}
	}
}

func TestBuildContext_LeadIncludesTempMemoryContract(t *testing.T) {
	r := &InternalToolRegistry{}
	ctx := r.BuildContext("admin", "admin-core", "orchestrator", nil, nil, "hello")
	if !strings.Contains(ctx, "Persistent Temp Memory Channels") {
		t.Fatalf("expected lead context to include temp memory section")
	}
	if !strings.Contains(ctx, "Stability rules: preserve user interaction style") {
		t.Fatalf("expected lead context to include stability rules")
	}
}

