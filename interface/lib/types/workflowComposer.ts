export type WorkflowNodeType =
    | 'trigger'
    | 'schedule'
    | 'manifest_team'
    | 'delegate_task'
    | 'decision_gate'
    | 'approval'
    | 'artifact_output'
    | 'mcp_action';

export interface WorkflowNodePosition {
    x: number;
    y: number;
}

export interface WorkflowNode {
    id: string;
    type: WorkflowNodeType;
    label: string;
    config?: Record<string, unknown>;
    position?: WorkflowNodePosition;
}

export interface WorkflowEdge {
    id: string;
    source_node_id: string;
    target_node_id: string;
    condition?: string;
}

export interface WorkflowDefinition {
    id: string;
    name: string;
    version: number;
    nodes: WorkflowNode[];
    edges: WorkflowEdge[];
    created_at?: string;
    updated_at?: string;
}

export interface WorkflowValidationIssue {
    code: string;
    message: string;
    node_id?: string;
    edge_id?: string;
}
