export interface ChatConsultation {
    member: string;
    summary: string;
}

export interface ChatArtifactRef {
    id?: string;
    type: string;
    title: string;
    content_type?: string;
    content?: string;
    url?: string;
    cached?: boolean;
    expires_at?: string;
    saved_path?: string;
}

export interface ExecutionSummaryLink {
    label?: string;
    title?: string;
    url?: string;
    href?: string;
    path?: string;
    id?: string;
    run_id?: string;
    run_class?: string;
    proof_class?: string;
    no_run_reason?: string;
    audit_event_id?: string;
    intent_proof_id?: string;
    exchange_item_id?: string;
    verified?: boolean;
}

export interface ExecutionSummaryItem {
    label?: string;
    title?: string;
    name?: string;
    summary?: string;
    value?: string;
    url?: string;
    href?: string;
    path?: string;
    id?: string;
    type?: string;
    status?: string;
    kind?: string;
    retained?: boolean;
    risk?: string;
    reason?: string;
}

export interface ExecutionSummaryIntent {
    original?: string;
    resolved?: string;
}

export interface ExecutionSummaryUnderstanding {
    summary?: string;
    assumptions?: string[];
}

export interface ExecutionSummaryExecution {
    shape?: string;
    status?: string;
    summary?: string;
}

export interface ExecutionSummaryCapabilityUse {
    capabilities?: Array<string | ExecutionSummaryItem>;
    teams?: Array<string | ExecutionSummaryItem>;
    agents?: Array<string | ExecutionSummaryItem>;
    tools?: Array<string | ExecutionSummaryItem>;
    used?: Array<string | ExecutionSummaryItem>;
}

export interface ExecutionSummaryData {
    intent?: string | ExecutionSummaryIntent;
    understanding?: string | ExecutionSummaryUnderstanding;
    execution?: ExecutionSummaryExecution;
    execution_shape?: string;
    execution_status?: string;
    execution_summary?: string;
    capability_use?: ExecutionSummaryCapabilityUse | Array<string | ExecutionSummaryItem>;
    outputs?: Array<string | ExecutionSummaryItem> | string;
    proof?: Array<string | ExecutionSummaryLink> | ExecutionSummaryLink | string;
    audit_recovery?: string | (ExecutionSummaryItem & {
        approval_status?: string;
        recovery_state?: string;
        blocker?: string;
        retryable?: boolean;
    });
    next_step?: string | ExecutionSummaryLink & { action?: string };
}

export type AskClass =
    | 'direct_answer'
    | 'governed_mutation'
    | 'governed_artifact'
    | 'specialist_consultation'
    | 'execution_blocker';

export type TemplateID = 'chat-to-answer' | 'chat-to-proposal';
export type ExecutionMode = 'answer' | 'proposal' | 'execution_result' | 'blocker';
export type ProposalLifecycleStatus =
    | 'active'
    | 'cancelled'
    | 'confirmed_pending_execution'
    | 'executed'
    | 'failed';

export interface AnswerProvenance {
    resolved_intent: string;
    permission_check: string;
    policy_decision: string;
    audit_event_id: string;
    consult_chain?: string[];
}

export interface ModuleBindingData {
    binding_id?: string;
    module_id: string;
    adapter_kind?: string;
    operation?: string;
}

export interface TeamExpressionData {
    expression_id?: string;
    team_id?: string;
    objective: string;
    role_plan?: string[];
    module_bindings?: ModuleBindingData[];
}

export type TaskCadence = 'run_once' | 'scheduled' | 'continuous' | 'event_driven';
export type BusScope = 'none' | 'current_team' | 'multi_team' | 'global';

export interface ProposalData {
    intent: string;
    operator_summary?: string;
    expected_result?: string;
    affected_resources?: string[];
    teams: number;
    agents: number;
    tools: string[];
    risk_level: string;
    confirm_token: string;
    intent_proof_id: string;
    approval_required?: boolean;
    approval_reason?: string;
    approval_mode?: string;
    capability_risk?: string;
    capability_ids?: string[];
    external_data_use?: boolean;
    estimated_cost?: number;
    team_expressions?: TeamExpressionData[];
    task_cadence?: TaskCadence;
    schedule_summary?: string;
    runtime_posture?: string;
    bus_scope?: BusScope;
    nats_subjects?: string[];
}

export interface ConfirmProposalResult {
    ok: boolean;
    runId: string | null;
    error?: string;
}

export interface BrainProvenance {
    provider_id: string;
    provider_name?: string;
    model_id: string;
    location: 'local' | 'remote';
    data_boundary: 'local_only' | 'leaves_org';
    tokens_used?: number;
}

export interface ChatMessage {
    role: 'user' | 'architect' | 'admin' | 'council' | 'system';
    content: string;
    consultations?: ChatConsultation[];
    tools_used?: string[];
    source_node?: string;
    trust_score?: number;
    timestamp?: string;
    artifacts?: ChatArtifactRef[];
    ask_class?: AskClass;
    template_id?: TemplateID;
    mode?: ExecutionMode;
    provenance?: AnswerProvenance;
    proposal?: ProposalData;
    proposal_status?: ProposalLifecycleStatus;
    execution_summary?: ExecutionSummaryData;
    brain?: BrainProvenance;
    run_id?: string;
}

export interface CouncilMember {
    id: string;
    role: string;
    team: string;
}

export interface APIResponse<T = unknown> {
    ok: boolean;
    data?: T;
    error?: string;
}

export interface CTSChatEnvelope {
    meta: { source_node: string; timestamp: string; trace_id?: string };
    signal_type: string;
    trust_score: number;
    template_id?: TemplateID;
    mode?: ExecutionMode;
    payload: {
        text: string;
        ask_class?: AskClass;
        consultations?: ChatConsultation[];
        tools_used?: string[];
        artifacts?: ChatArtifactRef[];
        provenance?: AnswerProvenance;
        brain?: BrainProvenance;
        execution_summary?: ExecutionSummaryData;
        proposal?: {
            intent: string;
            tools: string[];
            risk_level: string;
            confirm_token: string;
            intent_proof_id: string;
            team_expressions?: {
                expression_id?: string;
                team_id?: string;
                objective: string;
                role_plan?: string[];
                module_bindings?: {
                    binding_id?: string;
                    module_id: string;
                    adapter_kind?: string;
                    operation?: string;
                }[];
            }[];
            task_cadence?: TaskCadence;
            schedule_summary?: string;
            runtime_posture?: string;
            bus_scope?: BusScope;
            nats_subjects?: string[];
        };
    };
}

export interface StreamSignal {
    type: string;
    source?: string;
    level?: string;
    message?: string;
    timestamp?: string;
    trust_score?: number;
    payload?: any;
    topic?: string;
    source_kind?: string;
    source_channel?: string;
    payload_kind?: string;
    team_id?: string;
    agent_id?: string;
    run_id?: string;
}
