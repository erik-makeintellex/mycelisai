package protocol

import (
	"fmt"
	"time"
)

// WorkflowNodeType enumerates supported node classes for the workflow composer.
type WorkflowNodeType string

const (
	WorkflowNodeTrigger      WorkflowNodeType = "trigger"
	WorkflowNodeSchedule     WorkflowNodeType = "schedule"
	WorkflowNodeManifestTeam WorkflowNodeType = "manifest_team"
	WorkflowNodeDelegateTask WorkflowNodeType = "delegate_task"
	WorkflowNodeDecision     WorkflowNodeType = "decision_gate"
	WorkflowNodeApproval     WorkflowNodeType = "approval"
	WorkflowNodeArtifactOut  WorkflowNodeType = "artifact_output"
	WorkflowNodeMCPAction    WorkflowNodeType = "mcp_action"
)

var allowedWorkflowNodeTypes = map[WorkflowNodeType]struct{}{
	WorkflowNodeTrigger:      {},
	WorkflowNodeSchedule:     {},
	WorkflowNodeManifestTeam: {},
	WorkflowNodeDelegateTask: {},
	WorkflowNodeDecision:     {},
	WorkflowNodeApproval:     {},
	WorkflowNodeArtifactOut:  {},
	WorkflowNodeMCPAction:    {},
}

func (t WorkflowNodeType) IsValid() bool {
	_, ok := allowedWorkflowNodeTypes[t]
	return ok
}

type WorkflowNodePosition struct {
	X float64 `json:"x"`
	Y float64 `json:"y"`
}

type WorkflowNode struct {
	ID       string                 `json:"id"`
	Type     WorkflowNodeType       `json:"type"`
	Label    string                 `json:"label"`
	Config   map[string]interface{} `json:"config,omitempty"`
	Position WorkflowNodePosition   `json:"position,omitempty"`
}

type WorkflowEdge struct {
	ID           string `json:"id"`
	SourceNodeID string `json:"source_node_id"`
	TargetNodeID string `json:"target_node_id"`
	Condition    string `json:"condition,omitempty"`
}

type WorkflowDefinition struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Version   int            `json:"version"`
	Nodes     []WorkflowNode `json:"nodes"`
	Edges     []WorkflowEdge `json:"edges"`
	CreatedAt time.Time      `json:"created_at,omitempty"`
	UpdatedAt time.Time      `json:"updated_at,omitempty"`
}

type WorkflowValidationIssue struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	NodeID  string `json:"node_id,omitempty"`
	EdgeID  string `json:"edge_id,omitempty"`
}

func ValidateWorkflowDefinition(def WorkflowDefinition) []WorkflowValidationIssue {
	issues := make([]WorkflowValidationIssue, 0)
	if def.ID == "" {
		issues = append(issues, WorkflowValidationIssue{
			Code:    "workflow.missing_id",
			Message: "workflow id is required",
		})
	}
	if def.Name == "" {
		issues = append(issues, WorkflowValidationIssue{
			Code:    "workflow.missing_name",
			Message: "workflow name is required",
		})
	}
	if len(def.Nodes) == 0 {
		issues = append(issues, WorkflowValidationIssue{
			Code:    "workflow.no_nodes",
			Message: "workflow must contain at least one node",
		})
		return issues
	}

	nodeIndex := make(map[string]WorkflowNode, len(def.Nodes))
	hasApproval := false
	hasHighImpact := false

	for _, n := range def.Nodes {
		if n.ID == "" {
			issues = append(issues, WorkflowValidationIssue{
				Code:    "node.missing_id",
				Message: "node id is required",
			})
			continue
		}
		if _, exists := nodeIndex[n.ID]; exists {
			issues = append(issues, WorkflowValidationIssue{
				Code:    "node.duplicate_id",
				Message: fmt.Sprintf("duplicate node id '%s'", n.ID),
				NodeID:  n.ID,
			})
			continue
		}
		if !n.Type.IsValid() {
			issues = append(issues, WorkflowValidationIssue{
				Code:    "node.invalid_type",
				Message: fmt.Sprintf("node '%s' has unsupported type '%s'", n.ID, n.Type),
				NodeID:  n.ID,
			})
		}
		if n.Type == WorkflowNodeApproval {
			hasApproval = true
		}
		if n.Type == WorkflowNodeManifestTeam || n.Type == WorkflowNodeDelegateTask || n.Type == WorkflowNodeMCPAction {
			hasHighImpact = true
		}
		nodeIndex[n.ID] = n
	}

	if len(def.Edges) == 0 {
		issues = append(issues, WorkflowValidationIssue{
			Code:    "workflow.no_edges",
			Message: "workflow must contain at least one edge",
		})
	}

	graph := make(map[string][]string, len(def.Nodes))
	for _, e := range def.Edges {
		if e.SourceNodeID == "" || e.TargetNodeID == "" {
			issues = append(issues, WorkflowValidationIssue{
				Code:    "edge.missing_endpoint",
				Message: "edge source and target node ids are required",
				EdgeID:  e.ID,
			})
			continue
		}
		if _, ok := nodeIndex[e.SourceNodeID]; !ok {
			issues = append(issues, WorkflowValidationIssue{
				Code:    "edge.missing_source",
				Message: fmt.Sprintf("edge '%s' references missing source node '%s'", e.ID, e.SourceNodeID),
				EdgeID:  e.ID,
			})
			continue
		}
		if _, ok := nodeIndex[e.TargetNodeID]; !ok {
			issues = append(issues, WorkflowValidationIssue{
				Code:    "edge.missing_target",
				Message: fmt.Sprintf("edge '%s' references missing target node '%s'", e.ID, e.TargetNodeID),
				EdgeID:  e.ID,
			})
			continue
		}
		if e.SourceNodeID == e.TargetNodeID {
			issues = append(issues, WorkflowValidationIssue{
				Code:    "edge.self_loop",
				Message: fmt.Sprintf("edge '%s' cannot create self-loop on node '%s'", e.ID, e.SourceNodeID),
				EdgeID:  e.ID,
			})
		}
		graph[e.SourceNodeID] = append(graph[e.SourceNodeID], e.TargetNodeID)
	}

	if hasHighImpact && !hasApproval {
		issues = append(issues, WorkflowValidationIssue{
			Code:    "workflow.missing_approval",
			Message: "workflow contains high-impact nodes but no approval gate",
		})
	}

	if hasCycle(graph) {
		issues = append(issues, WorkflowValidationIssue{
			Code:    "workflow.cycle_detected",
			Message: "workflow must be acyclic",
		})
	}

	return issues
}

func hasCycle(graph map[string][]string) bool {
	visiting := map[string]bool{}
	visited := map[string]bool{}

	var dfs func(node string) bool
	dfs = func(node string) bool {
		if visiting[node] {
			return true
		}
		if visited[node] {
			return false
		}
		visiting[node] = true
		for _, next := range graph[node] {
			if dfs(next) {
				return true
			}
		}
		visiting[node] = false
		visited[node] = true
		return false
	}

	for node := range graph {
		if dfs(node) {
			return true
		}
	}
	return false
}
