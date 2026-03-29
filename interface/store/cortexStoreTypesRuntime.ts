export interface CTSEnvelope {
    id: string;
    source: string;
    signal: 'artifact' | 'governance_halt';
    timestamp: string;
    trust_score?: number;
    payload: {
        content: string;
        content_type: 'markdown' | 'json' | 'text' | 'image';
        title?: string;
    };
    proof?: {
        method: 'semantic' | 'empirical';
        logs: string;
        rubric_score: string;
        pass: boolean;
    };
}

export interface SensorNode {
    id: string;
    type: string;
    status: 'online' | 'offline' | 'degraded';
    last_seen: string;
    label: string;
}

export interface SignalDetail {
    type: string;
    source: string;
    level?: string;
    message: string;
    timestamp: string;
    topic?: string;
    payload?: any;
    source_kind?: string;
    source_channel?: string;
    payload_kind?: string;
    team_id?: string;
    agent_id?: string;
    run_id?: string;
    id?: string;
    trace_id?: string;
    intent?: string;
    context?: Record<string, unknown>;
    trust_score?: number;
}

export interface LogEntry {
    id: string;
    trace_id: string;
    timestamp: string;
    level: string;
    source: string;
    intent: string;
    message: string;
    context: Record<string, unknown>;
}

export interface MissionRun {
    id: string;
    mission_id: string;
    tenant_id: string;
    status: 'pending' | 'running' | 'completed' | 'failed';
    run_depth: number;
    parent_run_id?: string;
    started_at: string;
    completed_at?: string;
    metadata?: Record<string, unknown>;
}

export interface MissionEvent {
    id: string;
    run_id: string;
    tenant_id: string;
    event_type: string;
    severity: string;
    source_agent?: string;
    source_team?: string;
    payload?: Record<string, unknown>;
    audit_event_id?: string;
    emitted_at: string;
}

export interface TriggerRule {
    id: string;
    tenant_id: string;
    name: string;
    description?: string;
    event_pattern: string;
    condition: Record<string, unknown>;
    target_mission_id: string;
    mode: 'propose' | 'auto_execute';
    cooldown_seconds: number;
    max_depth: number;
    max_active_runs: number;
    is_active: boolean;
    last_fired_at?: string;
    created_at: string;
    updated_at: string;
}

export interface TriggerRuleCreate {
    name: string;
    description?: string;
    event_pattern: string;
    condition?: Record<string, unknown>;
    target_mission_id: string;
    mode?: 'propose' | 'auto_execute';
    cooldown_seconds?: number;
    max_depth?: number;
    max_active_runs?: number;
    is_active?: boolean;
}

export interface TriggerExecution {
    id: string;
    rule_id: string;
    event_id: string;
    run_id?: string;
    status: 'fired' | 'skipped' | 'proposed';
    skip_reason?: string;
    executed_at: string;
}

export interface PolicyRule {
    intent: string;
    condition: string;
    action: 'ALLOW' | 'DENY' | 'REQUIRE_APPROVAL';
}

export interface PolicyGroup {
    name: string;
    description: string;
    targets: string[];
    rules: PolicyRule[];
}

export interface PolicyConfig {
    groups: PolicyGroup[];
    defaults: { default_action: string };
}

export interface PendingApproval {
    id: string;
    reason: string;
    source_agent: string;
    team_id: string;
    intent: string;
    timestamp: string;
    expires_at: string;
}

export interface AuditLogEntry {
    id: string;
    template_id?: string;
    actor: string;
    user: string;
    action: string;
    timestamp: string;
    capability_used?: string;
    result_status: string;
    approval_status?: string;
    approval_reason?: string;
    run_id?: string;
    intent_proof_id?: string;
    resource?: string;
    details?: Record<string, unknown>;
}

export interface CognitiveEngineStatus {
    status: 'online' | 'offline';
    endpoint?: string;
    model?: string;
}

export interface CognitiveStatus {
    text: CognitiveEngineStatus;
    media: CognitiveEngineStatus;
}

export interface ServiceHealthStatus {
    name: string;
    status: 'online' | 'offline' | 'degraded';
    detail?: string;
    latency_ms?: number;
}
