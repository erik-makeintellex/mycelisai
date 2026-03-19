export type OrganizationStartMode = "template" | "empty";
export type OrganizationAIEngineProfileId = "starter_defaults" | "balanced" | "high_reasoning" | "fast_lightweight" | "deep_planning";
export type ResponseContractProfileId = "clear_balanced" | "structured_analytical" | "concise_direct" | "warm_supportive";

export interface OrganizationAgentTypeProfileSummary {
    id: string;
    name: string;
    helps_with: string;
    ai_engine_binding_profile_id?: OrganizationAIEngineProfileId;
    ai_engine_effective_profile_id?: OrganizationAIEngineProfileId;
    ai_engine_effective_summary: string;
    inherits_department_ai_engine: boolean;
    response_contract_binding_profile_id?: ResponseContractProfileId;
    response_contract_effective_profile_id?: ResponseContractProfileId;
    response_contract_effective_summary: string;
    inherits_default_response_contract: boolean;
}

export interface OrganizationDepartmentSummary {
    id: string;
    name: string;
    specialist_count: number;
    ai_engine_override_profile_id?: OrganizationAIEngineProfileId;
    ai_engine_override_summary?: string;
    ai_engine_effective_profile_id?: OrganizationAIEngineProfileId;
    ai_engine_effective_summary: string;
    inherits_organization_ai_engine: boolean;
    agent_type_profiles?: OrganizationAgentTypeProfileSummary[];
}

export interface OrganizationTemplateSummary {
    id: string;
    name: string;
    description: string;
    organization_type: string;
    team_lead_label: string;
    advisor_count: number;
    department_count: number;
    specialist_count: number;
    ai_engine_profile_id?: OrganizationAIEngineProfileId;
    ai_engine_settings_summary: string;
    response_contract_profile_id?: ResponseContractProfileId;
    response_contract_summary: string;
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
    ai_engine_profile_id?: OrganizationAIEngineProfileId;
    ai_engine_settings_summary: string;
    response_contract_profile_id?: ResponseContractProfileId;
    response_contract_summary: string;
    memory_personality_summary: string;
    status: string;
}

export interface OrganizationHomePayload extends OrganizationSummary {
    description?: string;
    departments?: OrganizationDepartmentSummary[];
}

export interface OrganizationCreateRequest {
    name: string;
    purpose: string;
    start_mode: OrganizationStartMode;
    template_id?: string;
}

export interface OrganizationAIEngineUpdateRequest {
    profile_id: OrganizationAIEngineProfileId;
}

export interface DepartmentAIEngineUpdateRequest {
    profile_id?: OrganizationAIEngineProfileId;
    revert_to_organization_default?: boolean;
}

export interface ResponseContractUpdateRequest {
    profile_id: ResponseContractProfileId;
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
