export type TeamWorkItemState =
    | 'new'
    | 'queued'
    | 'running'
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
    | 'archive';

export interface TeamInteraction {
    action: TeamInteractionAction;
    label: string;
    href?: string;
    disabled?: boolean;
    disabledReason?: string;
    audited?: boolean;
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
    advanced?: {
        inputs?: string[];
        deliveries?: string[];
        modelIds?: string[];
        toolIds?: string[];
        promptCount?: number;
        policyRef?: string;
        capabilityIds?: string[];
    };
}
