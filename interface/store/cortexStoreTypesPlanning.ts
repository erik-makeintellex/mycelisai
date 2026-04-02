export interface MissionProfile {
    id: string;
    name: string;
    description?: string;
    role_providers: Record<string, string>;
    subscriptions: { topic: string; condition?: string }[];
    context_strategy: 'fresh' | 'warm' | string;
    auto_start: boolean;
    is_active: boolean;
    tenant_id: string;
    created_at: string;
    updated_at: string;
}

export interface MissionProfileCreate {
    name: string;
    description?: string;
    role_providers: Record<string, string>;
    subscriptions: { topic: string; condition?: string }[];
    context_strategy: string;
    auto_start: boolean;
}

export interface ContextSnapshot {
    id: string;
    name: string;
    description?: string;
    source_profile?: string;
    tenant_id: string;
    created_at: string;
}

export interface AgentManifest {
    id: string;
    role: string;
    system_prompt?: string;
    model?: string;
    inputs?: string[];
    outputs?: string[];
    tools?: string[];
}

export interface BlueprintTeam {
    name: string;
    role: string;
    agents: AgentManifest[];
}

export interface Constraint {
    constraint_id?: string;
    description: string;
}

export interface MissionBlueprint {
    mission_id: string;
    intent: string;
    teams: BlueprintTeam[];
    constraints?: Constraint[];
}

export type MissionStatus = 'idle' | 'draft' | 'active';

export interface Mission {
    id: string;
    intent: string;
    status: 'active' | 'completed' | 'failed';
    teams: number;
    agents: number;
    created_at?: string;
}

export interface ProposedAgent {
    id: string;
    role: string;
    system_prompt?: string;
    model?: string;
}

export interface TeamProposal {
    id: string;
    name: string;
    role: string;
    agents: ProposedAgent[];
    reason: string;
    status: 'pending' | 'approved' | 'rejected';
    created_at: string;
}

export interface CatalogueAgent {
    id: string;
    name: string;
    role: string;
    system_prompt?: string;
    model?: string;
    tools: string[];
    inputs: string[];
    outputs: string[];
    verification_strategy?: string;
    verification_rubric: string[];
    validation_command?: string;
    created_at: string;
    updated_at: string;
}

export type ArtifactType = 'code' | 'document' | 'image' | 'audio' | 'data' | 'file' | 'chart';
export type ArtifactStatus = 'pending' | 'approved' | 'rejected' | 'archived';

export interface Artifact {
    id: string;
    mission_id?: string;
    team_id?: string;
    agent_id: string;
    trace_id?: string;
    artifact_type: ArtifactType;
    title: string;
    content_type: string;
    content?: string;
    file_path?: string;
    file_size_bytes?: number;
    metadata: Record<string, any>;
    trust_score?: number;
    status: ArtifactStatus;
    created_at: string;
}

export interface ArtifactFilters {
    mission_id?: string;
    team_id?: string;
    agent_id?: string;
    limit?: number;
}

export interface TeamAgent {
    id: string;
    name: string;
    team_id: string;
    status: number;
    last_heartbeat: string;
}

export interface TeamDetail {
    id: string;
    name: string;
    role: string;
    agents: TeamAgent[];
}

export interface TeamDetailAgentEntry {
    id: string;
    role: string;
    status: number;
    last_heartbeat: string;
    tools: string[];
    model: string;
    system_prompt?: string;
}

export interface TeamDetailEntry {
    id: string;
    name: string;
    role: string;
    type: 'standing' | 'mission';
    mission_id: string | null;
    mission_intent: string | null;
    inputs: string[];
    deliveries: string[];
    agents: TeamDetailAgentEntry[];
}

export type TeamsFilter = 'all' | 'standing' | 'mission';
export type StreamConnectionState = 'idle' | 'connecting' | 'online' | 'offline';

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
}

export interface MCPTool {
    id: string;
    server_id: string;
    name: string;
    description?: string;
    input_schema: Record<string, any>;
}

export interface MCPServerWithTools extends MCPServer {
    tools: MCPTool[];
}

export interface MCPLibraryEntry {
    name: string;
    description: string;
    transport: string;
    command: string;
    args: string[];
    env?: Record<string, string>;
    url?: string;
    tags: string[];
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
