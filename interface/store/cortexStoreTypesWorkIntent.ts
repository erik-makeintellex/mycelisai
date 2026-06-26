export type TaskCadence = "run_once" | "scheduled" | "continuous" | "event_driven";
export type BusScope = "none" | "current_team" | "multi_team" | "global";

export type WorkExecutionMode =
    | "answer"
    | "propose"
    | "confirm_then_execute"
    | "auto_execute"
    | "schedule_handoff"
    | "team_async";

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
}
