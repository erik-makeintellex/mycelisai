export type TaskCadence = "run_once" | "scheduled" | "continuous" | "event_driven";
export type BusScope = "none" | "current_team" | "multi_team" | "global";

export type WorkExecutionMode =
    | "answer"
    | "propose"
    | "confirm_then_execute"
    | "auto_execute"
    | "schedule_handoff"
    | "team_async";

export interface WorkOutputContractData {
    shape?: string;
    primary_deliverable?: string;
    retention?: string;
    launch_hint?: string;
    validation?: string[];
}

export interface WorkLifecycleContractData {
    stop_action?: string;
    retry_action?: string;
    recovery_action?: string;
    control_summary?: string;
}

export interface WorkIntentData {
    kind?: string;
    objective?: string;
    cadence?: TaskCadence;
    schedule_summary?: string;
    runtime_posture?: string;
    target_team_id?: string;
    bus_scope?: BusScope;
    nats_subjects?: string[];
    service_refs?: string[];
    project_ref?: string;
    output_contract?: WorkOutputContractData;
    lifecycle?: WorkLifecycleContractData;
}
