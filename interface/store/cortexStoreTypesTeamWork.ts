import type { OutputProofEnvelope } from "./cortexStoreTypesExecutionSummary";

export type TeamWorkItemState =
    | 'new'
    | 'briefed'
    | 'queued'
    | 'running'
    | 'reviewing'
    | 'paused'
    | 'output_ready'
    | 'degraded'
    | 'needs_operator'
    | 'archived';

export type TeamInteractionAction =
    | 'inspect'
    | 'steer'
    | 'start_work'
    | 'pause'
    | 'resume'
    | 'recover'
    | 'archive';

export interface TeamInteraction {
    action: TeamInteractionAction;
    label: string;
    href?: string;
    disabled?: boolean;
    disabledReason?: string;
    audited?: boolean;
}

export interface TeamOutputRef {
    output_id: string;
    team_id: string;
    work_item_id: string;
    run_id?: string;
    kind: string;
    label: string;
    storage_ref?: string;
    entrypoint?: string;
    validation_ref?: string;
    proof_ref?: string;
    contract_id?: string;
    proof_id?: string;
    proof?: OutputProofEnvelope;
    audit_refs?: string[];
    created_at?: string;
}

export interface TargetRef {
    type: string;
    id: string;
    run_id?: string;
    team_id?: string;
    work_item_id?: string;
    project_id?: string;
    output_id?: string;
    label?: string;
}

export interface TeamWorkItem {
    id: string;
    title: string;
    description?: string;
    state: TeamWorkItemState;
    ownerLabel: string;
    scopeLabel: string;
    updatedAt?: string | null;
    outputCount?: number;
    teamIds: string[];
    interactions: TeamInteraction[];
    source?: 'durable' | 'projection';
    sourceLabel?: string;
    fallbackReason?: string;
    runId?: string;
    outputRefs?: TeamOutputRef[];
    proofRefs?: string[];
    auditRefs?: string[];
    needsOperator?: boolean;
    nextAction?: string;
    recoveryOptions?: string[];
    targetRef?: TargetRef;
    advanced?: {
        inputs?: string[];
        deliveries?: string[];
        modelIds?: string[];
        toolIds?: string[];
        promptCount?: number;
        policyRef?: string;
        capabilityIds?: string[];
        expectedOutputs?: string[];
        expectedProof?: string[];
        executionShape?: string[];
    };
}
