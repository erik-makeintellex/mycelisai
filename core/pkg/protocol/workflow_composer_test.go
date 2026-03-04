package protocol

import "testing"

func hasIssue(issues []WorkflowValidationIssue, code string) bool {
	for _, i := range issues {
		if i.Code == code {
			return true
		}
	}
	return false
}

func TestValidateWorkflowDefinition_ValidGraph(t *testing.T) {
	def := WorkflowDefinition{
		ID:      "wf-1",
		Name:    "valid",
		Version: 1,
		Nodes: []WorkflowNode{
			{ID: "n1", Type: WorkflowNodeTrigger, Label: "trigger"},
			{ID: "n2", Type: WorkflowNodeApproval, Label: "approval"},
			{ID: "n3", Type: WorkflowNodeManifestTeam, Label: "manifest"},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", SourceNodeID: "n1", TargetNodeID: "n2"},
			{ID: "e2", SourceNodeID: "n2", TargetNodeID: "n3"},
		},
	}

	issues := ValidateWorkflowDefinition(def)
	if len(issues) != 0 {
		t.Fatalf("expected no issues, got: %+v", issues)
	}
}

func TestValidateWorkflowDefinition_MissingApprovalForHighImpact(t *testing.T) {
	def := WorkflowDefinition{
		ID:      "wf-2",
		Name:    "missing-approval",
		Version: 1,
		Nodes: []WorkflowNode{
			{ID: "n1", Type: WorkflowNodeTrigger, Label: "trigger"},
			{ID: "n2", Type: WorkflowNodeManifestTeam, Label: "manifest"},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", SourceNodeID: "n1", TargetNodeID: "n2"},
		},
	}

	issues := ValidateWorkflowDefinition(def)
	if !hasIssue(issues, "workflow.missing_approval") {
		t.Fatalf("expected workflow.missing_approval issue, got: %+v", issues)
	}
}

func TestValidateWorkflowDefinition_CycleDetected(t *testing.T) {
	def := WorkflowDefinition{
		ID:      "wf-3",
		Name:    "cycle",
		Version: 1,
		Nodes: []WorkflowNode{
			{ID: "n1", Type: WorkflowNodeTrigger, Label: "trigger"},
			{ID: "n2", Type: WorkflowNodeApproval, Label: "approval"},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", SourceNodeID: "n1", TargetNodeID: "n2"},
			{ID: "e2", SourceNodeID: "n2", TargetNodeID: "n1"},
		},
	}

	issues := ValidateWorkflowDefinition(def)
	if !hasIssue(issues, "workflow.cycle_detected") {
		t.Fatalf("expected cycle issue, got: %+v", issues)
	}
}

func TestValidateWorkflowDefinition_MissingEdgeTarget(t *testing.T) {
	def := WorkflowDefinition{
		ID:      "wf-4",
		Name:    "missing-edge-target",
		Version: 1,
		Nodes: []WorkflowNode{
			{ID: "n1", Type: WorkflowNodeTrigger, Label: "trigger"},
		},
		Edges: []WorkflowEdge{
			{ID: "e1", SourceNodeID: "n1", TargetNodeID: "n2"},
		},
	}

	issues := ValidateWorkflowDefinition(def)
	if !hasIssue(issues, "edge.missing_target") {
		t.Fatalf("expected edge.missing_target issue, got: %+v", issues)
	}
}
