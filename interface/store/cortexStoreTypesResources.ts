export interface MCPServer {
    id: string;
    name: string;
    transport: 'stdio' | 'sse';
    command?: string;
    args?: string[];
    env?: Record<string, string>;
    url?: string;
    headers?: Record<string, string>;
    status: string;
    error?: string;
    created_at: string;
    capability_ids?: string[];
}

export interface MCPTool {
    id: string;
    server_id: string;
    name: string;
    description?: string;
    input_schema: Record<string, any>;
    capability_id?: string;
}

export interface MCPServerWithTools extends MCPServer {
    tools: MCPTool[];
}

export interface MCPActivityEntry {
    id: string;
    server_id?: string;
    server_name: string;
    tool_name: string;
    state: string;
    summary: string;
    message: string;
    channel_name: string;
    run_id?: string;
    team_id?: string;
    agent_id?: string;
    timestamp: string;
}

export interface MCPLibraryPackage {
    registry_type: string;
    identifier: string;
    version?: string;
    transport: {
        type: string;
    };
}

export interface MCPLibraryEnvVar {
    name: string;
    description?: string;
    required?: boolean;
    secret?: boolean;
    default_value?: string;
}

export interface MCPLibraryEntry {
    name: string;
    title?: string;
    description: string;
    version?: string;
    transport: string;
    command: string;
    args: string[];
    env?: Record<string, string>;
    environment_variables?: MCPLibraryEnvVar[];
    url?: string;
    packages?: MCPLibraryPackage[];
    repository?: string;
    homepage?: string;
    tags: string[];
    tool_set?: string;
}

export interface MCPLibraryCategory {
    name: string;
    servers: MCPLibraryEntry[];
}

export interface MCPGovernanceDecision {
    decision: 'allow' | 'require_approval' | 'deny';
    approval_required: boolean;
    approval_mode?: string;
    approval_reason?: string;
    source_surface?: string;
    config_scope?: string;
    group_id?: string;
    reasons?: string[];
}

export interface MCPInstallResult {
    ok: boolean;
    message?: string;
    governance?: MCPGovernanceDecision;
}

export interface CapabilityManifest {
    id: string;
    name: string;
    description?: string;
    source: 'builtin' | 'mcp' | 'openapi' | 'local_script' | 'external_api' | 'python' | 'a2a' | 'plugin' | string;
    category: string;
    risk: 'low' | 'medium' | 'high' | string;
    approval: 'none' | 'optional' | 'required' | 'policy_resolved' | string;
    inputs?: string[];
    outputs?: string[];
    writes?: string[];
    allowed_roles?: string[];
    audit?: 'none' | 'optional' | 'required' | string;
    health_check?: boolean;
    availability_status?: 'available' | 'degraded' | 'unavailable' | 'unknown' | string;
    fallback_behavior?: string;
    retention_policy?: string;
    review_required?: boolean;
    server_or_package?: string;
    config_refs?: string[];
    secret_refs?: string[];
    provider?: string;
    bound_server_id?: string;
    bound_server_name?: string;
    bound_tool_id?: string;
    bound_tool_name?: string;
}
