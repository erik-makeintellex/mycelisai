export type OrganizationStartMode = "template" | "empty";

export interface OrganizationTemplateSummary {
    id: string;
    name: string;
    description: string;
    organization_type: string;
    team_lead_label: string;
    advisor_count: number;
    department_count: number;
    specialist_count: number;
    ai_engine_settings_summary: string;
    memory_personality_summary: string;
}

export interface OrganizationSummary {
    id: string;
    name: string;
    purpose: string;
    start_mode: OrganizationStartMode;
    template_id?: string;
    template_name?: string;
    team_lead_label: string;
    advisor_count: number;
    department_count: number;
    specialist_count: number;
    ai_engine_settings_summary: string;
    memory_personality_summary: string;
    status: string;
}

export interface OrganizationHomePayload extends OrganizationSummary {
    description?: string;
}

export interface OrganizationCreateRequest {
    name: string;
    purpose: string;
    start_mode: OrganizationStartMode;
    template_id?: string;
}

export type TeamLeadGuidedAction = "plan_next_steps" | "focus_first" | "review_setup";

export interface TeamLeadGuidanceRequest {
    action: TeamLeadGuidedAction;
}

export interface TeamLeadGuidanceResponse {
    action: TeamLeadGuidedAction;
    request_label: string;
    headline: string;
    summary: string;
    priority_steps: string[];
    suggested_follow_ups: string[];
}
