export type OrganizationStartMode = "template" | "empty";
export type OrganizationAIEngineProfileId = "starter_defaults" | "balanced" | "high_reasoning" | "fast_lightweight" | "deep_planning";
export type ResponseContractProfileId = "clear_balanced" | "structured_analytical" | "concise_direct" | "warm_supportive";
export type OrganizationOutputModelRoutingMode = "single_model" | "detected_output_types";
export type OrganizationOutputTypeId = "general_text" | "research_reasoning" | "code_generation" | "vision_analysis";
export type LoopActivityStatus = "success" | "warning" | "failed";

export interface OrganizationLoopActivityItem {
    id: string;
    name: string;
    last_run_at: string;
    status: LoopActivityStatus;
    summary: string;
}

export type OrganizationAutomationTriggerType = "scheduled" | "event_driven";
export type OrganizationLearningInsightStrength = "emerging" | "consistent" | "strong";

export interface OrganizationAutomationOutcomeItem {
    summary: string;
    occurred_at: string;
}

export interface OrganizationAutomationItem {
    id: string;
    name: string;
    purpose: string;
    trigger_type: OrganizationAutomationTriggerType;
    owner_label: string;
    status: LoopActivityStatus;
    watches: string;
    trigger_summary: string;
    recent_outcomes?: OrganizationAutomationOutcomeItem[];
}

export interface OrganizationLearningInsightItem {
    id: string;
    summary: string;
    source: string;
    observed_at: string;
    strength: OrganizationLearningInsightStrength;
}

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
    output_type_id?: OrganizationOutputTypeId;
    output_type_label?: string;
    output_model_effective_id?: string;
    output_model_effective_summary?: string;
    inherits_default_output_model: boolean;
}

export interface OrganizationOutputModelBinding {
    output_type_id: OrganizationOutputTypeId;
    output_type_label: string;
    model_id?: string;
    model_summary: string;
}

export interface OrganizationOutputModelCatalogEntry {
    model_id: string;
    label: string;
    summary: string;
    output_type_ids?: OrganizationOutputTypeId[];
    provider_id?: string;
    installed: boolean;
    popular: boolean;
    self_hostable: boolean;
    hosting_fit?: string;
    popularity_note?: string;
    source?: string;
}

export interface OrganizationOutputModelReviewCandidate {
    output_type_id: OrganizationOutputTypeId;
    output_type_label: string;
    model_id: string;
    model_summary: string;
    installed: boolean;
    review_criteria: string[];
}

export interface OrganizationOutputModelRoutingPayload {
    routing_mode: OrganizationOutputModelRoutingMode;
    default_model_id?: string;
    default_model_summary: string;
    bindings?: OrganizationOutputModelBinding[];
    available_models?: OrganizationOutputModelCatalogEntry[];
    recommended_models?: OrganizationOutputModelCatalogEntry[];
    review_candidates?: OrganizationOutputModelReviewCandidate[];
    hardware_summary: string;
    review_permission_prompt: string;
    automatic_selection_criteria?: string[];
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
    output_model_routing_mode?: OrganizationOutputModelRoutingMode;
    default_output_model_id?: string;
    default_output_model_summary?: string;
    status: string;
}

export interface OrganizationHomePayload extends OrganizationSummary {
    description?: string;
    departments?: OrganizationDepartmentSummary[];
    output_model_bindings?: OrganizationOutputModelBinding[];
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

export interface AgentTypeAIEngineUpdateRequest {
    profile_id?: OrganizationAIEngineProfileId;
    use_team_default?: boolean;
}

export interface AgentTypeResponseContractUpdateRequest {
    profile_id?: ResponseContractProfileId;
    use_organization_or_team_default?: boolean;
}

export interface ResponseContractUpdateRequest {
    profile_id: ResponseContractProfileId;
}

export interface OrganizationOutputModelBindingUpdateRequest {
    output_type_id: OrganizationOutputTypeId;
    model_id?: string;
    use_organization_default?: boolean;
}

export interface OrganizationOutputModelRoutingUpdateRequest {
    routing_mode: OrganizationOutputModelRoutingMode;
    default_model_id: string;
    bindings: OrganizationOutputModelBindingUpdateRequest[];
}

export type TeamLeadGuidedAction = "plan_next_steps" | "focus_first" | "review_setup";
export type TeamLeadExecutionMode = "guided_review" | "native_team" | "external_workflow_contract";

export interface TeamLeadGuidanceRequest {
    action: TeamLeadGuidedAction;
    request_context?: string;
}

export interface TeamLeadExecutionContract {
    execution_mode: TeamLeadExecutionMode;
    owner_label: string;
    summary: string;
    team_name?: string;
    external_target?: string;
    target_outputs: string[];
    workflow_group?: TeamLeadWorkflowGroupDraft;
}

export interface TeamLeadWorkflowGroupDraft {
    name: string;
    goal_statement: string;
    work_mode: "read_only" | "propose_only" | "execute_with_approval" | "execute_bounded";
    coordinator_profile: string;
    allowed_capabilities?: string[];
    expiry_hours?: number;
    summary: string;
}

export interface TeamLeadGuidanceResponse {
    action: TeamLeadGuidedAction;
    request_label: string;
    headline: string;
    summary: string;
    priority_steps: string[];
    suggested_follow_ups: string[];
    execution_contract?: TeamLeadExecutionContract;
}
