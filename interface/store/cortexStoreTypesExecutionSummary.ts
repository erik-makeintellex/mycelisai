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
    open_url?: string;
    path?: string;
    id?: string;
    artifact_id?: string;
    proof_artifact_id?: string;
    type?: string;
    status?: string;
    kind?: string;
    retained?: boolean;
    risk?: string;
    reason?: string;
    entrypoint?: string;
    folder?: string;
    files?: string[];
    validation?: string;
    proof?: OutputProofEnvelope;
}

export interface OutputProofEnvelope {
    proof_id?: string;
    output_ref_id?: string;
    artifact_id?: string;
    storage_ref?: string;
    source_run_id?: string;
    source_contract_id?: string;
    execution_status?: string;
    path_boundary_status?: string;
    checksum?: string;
    checksum_algorithm?: string;
    bytes?: number;
    content_type?: string;
    readback_status?: string;
    recovery_hint?: string;
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

export interface ExecutionDegradation {
    code?: string;
    what_failed?: string;
    trusted_state?: string;
    invalidated_proof?: string;
    safe_continuation?: string;
    requires_attention?: boolean;
}

export type UIResponseStateKind =
    | 'direct_answer'
    | 'proposal'
    | 'awaiting_approval'
    | 'running'
    | 'execution_result'
    | 'partial_completion'
    | 'degraded_execution'
    | 'blocker'
    | 'retry_required'
    | 'recovery_state';
export type UIResponseStateTone = 'neutral' | 'info' | 'success' | 'warning' | 'danger';

export interface UIResponseStateProjection {
    kind: UIResponseStateKind;
    label?: string;
    detail?: string;
    tone?: UIResponseStateTone;
}

export interface ExecutionSummaryData {
    ui_response_state?: UIResponseStateProjection;
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
        degradation?: ExecutionDegradation;
    });
    next_step?: string | ExecutionSummaryLink & { action?: string };
}
