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
    profile_ref?: string;
    role: string;
    system_prompt?: string;
    model?: string;
    inputs?: string[];
    outputs?: string[];
    tools?: string[];
    context?: AgentContextBinding[];
    usage?: AgentUsagePolicy;
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
    profile_key?: string;
    name: string;
    description?: string;
    role: string;
    source?: 'built_in' | 'user';
    locked?: boolean;
    system_prompt?: string;
    model?: string;
    tools: string[];
    capability_refs?: string[];
    context_bindings?: AgentContextBinding[];
    usage_policy?: AgentUsagePolicy;
    inputs: string[];
    outputs: string[];
    verification_strategy?: string;
    verification_rubric: string[];
    validation_command?: string;
    created_at: string;
    updated_at: string;
}

export interface AgentContextBinding {
    kind: string;
    ref?: string;
    access?: 'read' | 'write' | 'read_write' | string;
}

export interface AgentUsagePolicy {
    selection?: 'soma_or_manual' | 'soma' | 'manual' | 'automatic' | string;
    scope?: 'workspace' | 'outcome' | 'team' | string;
}

export type ArtifactType = 'code' | 'document' | 'image' | 'audio' | 'data' | 'file' | 'chart' | 'project_package';
export type ArtifactStatus = 'pending' | 'approved' | 'rejected' | 'archived';

export interface Artifact {
    id: string;
    mission_id?: string;
    team_id?: string;
    agent_id: string;
    trace_id?: string;
    artifact_type: ArtifactType;
    output_class?: string;
    title: string;
    content_type: string;
    content?: string;
    file_path?: string;
    file_size_bytes?: number;
    metadata: Record<string, unknown>;
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
