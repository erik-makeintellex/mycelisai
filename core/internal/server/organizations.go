package server

type OrganizationStartMode string

const (
	OrganizationStartModeTemplate OrganizationStartMode = "template"
	OrganizationStartModeEmpty    OrganizationStartMode = "empty"
)

type OrganizationAIEngineProfileID string
type ResponseContractProfileID string
type OrganizationOutputModelRoutingMode string
type OrganizationOutputTypeID string

const (
	OrganizationAIEngineProfileStarterDefaults OrganizationAIEngineProfileID = "starter_defaults"
	OrganizationAIEngineProfileBalanced        OrganizationAIEngineProfileID = "balanced"
	OrganizationAIEngineProfileHighReasoning   OrganizationAIEngineProfileID = "high_reasoning"
	OrganizationAIEngineProfileFastLightweight OrganizationAIEngineProfileID = "fast_lightweight"
	OrganizationAIEngineProfileDeepPlanning    OrganizationAIEngineProfileID = "deep_planning"
)

const (
	OrganizationOutputModelRoutingModeSingleModel         OrganizationOutputModelRoutingMode = "single_model"
	OrganizationOutputModelRoutingModeDetectedOutputTypes OrganizationOutputModelRoutingMode = "detected_output_types"
)

const (
	OrganizationOutputTypeGeneralText       OrganizationOutputTypeID = "general_text"
	OrganizationOutputTypeResearchReasoning OrganizationOutputTypeID = "research_reasoning"
	OrganizationOutputTypeCodeGeneration    OrganizationOutputTypeID = "code_generation"
	OrganizationOutputTypeVisionAnalysis    OrganizationOutputTypeID = "vision_analysis"
)

const (
	ResponseContractProfileClearBalanced        ResponseContractProfileID = "clear_balanced"
	ResponseContractProfileStructuredAnalytical ResponseContractProfileID = "structured_analytical"
	ResponseContractProfileConciseDirect        ResponseContractProfileID = "concise_direct"
	ResponseContractProfileWarmSupportive       ResponseContractProfileID = "warm_supportive"
)

type organizationAIEngineProfile struct {
	ID          OrganizationAIEngineProfileID
	Summary     string
	Description string
	BestFor     string
}

type responseContractProfile struct {
	ID        ResponseContractProfileID
	Summary   string
	ToneStyle string
	Structure string
	Verbosity string
	BestFor   string
}

type OrganizationAgentTypeProfileSummary struct {
	ID                                 string `json:"id"`
	Name                               string `json:"name"`
	HelpsWith                          string `json:"helps_with"`
	AIEngineBindingProfileID           string `json:"ai_engine_binding_profile_id,omitempty"`
	AIEngineEffectiveProfileID         string `json:"ai_engine_effective_profile_id,omitempty"`
	AIEngineEffectiveSummary           string `json:"ai_engine_effective_summary"`
	InheritsDepartmentAIEngine         bool   `json:"inherits_department_ai_engine"`
	ResponseContractBindingProfileID   string `json:"response_contract_binding_profile_id,omitempty"`
	ResponseContractEffectiveProfileID string `json:"response_contract_effective_profile_id,omitempty"`
	ResponseContractEffectiveSummary   string `json:"response_contract_effective_summary"`
	InheritsDefaultResponseContract    bool   `json:"inherits_default_response_contract"`
	OutputTypeID                       string `json:"output_type_id,omitempty"`
	OutputTypeLabel                    string `json:"output_type_label,omitempty"`
	OutputModelEffectiveID             string `json:"output_model_effective_id,omitempty"`
	OutputModelEffectiveSummary        string `json:"output_model_effective_summary,omitempty"`
	InheritsDefaultOutputModel         bool   `json:"inherits_default_output_model"`
}

type OrganizationOutputModelBinding struct {
	OutputTypeID    string `json:"output_type_id"`
	OutputTypeLabel string `json:"output_type_label"`
	ModelID         string `json:"model_id,omitempty"`
	ModelSummary    string `json:"model_summary"`
}

type OrganizationOutputModelCatalogEntry struct {
	ModelID        string   `json:"model_id"`
	Label          string   `json:"label"`
	Summary        string   `json:"summary"`
	OutputTypeIDs  []string `json:"output_type_ids,omitempty"`
	ProviderID     string   `json:"provider_id,omitempty"`
	Installed      bool     `json:"installed"`
	Popular        bool     `json:"popular"`
	SelfHostable   bool     `json:"self_hostable"`
	HostingFit     string   `json:"hosting_fit,omitempty"`
	PopularityNote string   `json:"popularity_note,omitempty"`
	Source         string   `json:"source,omitempty"`
}

type OrganizationOutputModelReviewCandidate struct {
	OutputTypeID    string   `json:"output_type_id"`
	OutputTypeLabel string   `json:"output_type_label"`
	ModelID         string   `json:"model_id"`
	ModelSummary    string   `json:"model_summary"`
	Installed       bool     `json:"installed"`
	ReviewCriteria  []string `json:"review_criteria"`
}

type OrganizationOutputModelRoutingPayload struct {
	RoutingMode                string                                   `json:"routing_mode"`
	DefaultModelID             string                                   `json:"default_model_id,omitempty"`
	DefaultModelSummary        string                                   `json:"default_model_summary"`
	Bindings                   []OrganizationOutputModelBinding         `json:"bindings,omitempty"`
	AvailableModels            []OrganizationOutputModelCatalogEntry    `json:"available_models,omitempty"`
	RecommendedModels          []OrganizationOutputModelCatalogEntry    `json:"recommended_models,omitempty"`
	ReviewCandidates           []OrganizationOutputModelReviewCandidate `json:"review_candidates,omitempty"`
	HardwareSummary            string                                   `json:"hardware_summary"`
	ReviewPermissionPrompt     string                                   `json:"review_permission_prompt"`
	AutomaticSelectionCriteria []string                                 `json:"automatic_selection_criteria,omitempty"`
}

type OrganizationDepartmentSummary struct {
	ID                           string                                `json:"id"`
	Name                         string                                `json:"name"`
	SpecialistCount              int                                   `json:"specialist_count"`
	AIEngineOverrideProfileID    string                                `json:"ai_engine_override_profile_id,omitempty"`
	AIEngineOverrideSummary      string                                `json:"ai_engine_override_summary,omitempty"`
	AIEngineEffectiveProfileID   string                                `json:"ai_engine_effective_profile_id,omitempty"`
	AIEngineEffectiveSummary     string                                `json:"ai_engine_effective_summary"`
	InheritsOrganizationAIEngine bool                                  `json:"inherits_organization_ai_engine"`
	AgentTypeProfiles            []OrganizationAgentTypeProfileSummary `json:"agent_type_profiles,omitempty"`
}

type OrganizationTemplateSummary struct {
	ID                       string                          `json:"id"`
	Name                     string                          `json:"name"`
	Description              string                          `json:"description"`
	OrganizationType         string                          `json:"organization_type"`
	TeamLeadLabel            string                          `json:"team_lead_label"`
	AdvisorCount             int                             `json:"advisor_count"`
	DepartmentCount          int                             `json:"department_count"`
	SpecialistCount          int                             `json:"specialist_count"`
	Departments              []OrganizationDepartmentSummary `json:"departments,omitempty"`
	AIEngineSettingsSummary  string                          `json:"ai_engine_settings_summary"`
	ResponseContractSummary  string                          `json:"response_contract_summary"`
	MemoryPersonalitySummary string                          `json:"memory_personality_summary"`
}

type OrganizationSummary struct {
	ID                        string                `json:"id"`
	Name                      string                `json:"name"`
	Purpose                   string                `json:"purpose"`
	StartMode                 OrganizationStartMode `json:"start_mode"`
	TemplateID                string                `json:"template_id,omitempty"`
	TemplateName              string                `json:"template_name,omitempty"`
	TeamLeadLabel             string                `json:"team_lead_label"`
	AdvisorCount              int                   `json:"advisor_count"`
	DepartmentCount           int                   `json:"department_count"`
	SpecialistCount           int                   `json:"specialist_count"`
	AIEngineProfileID         string                `json:"ai_engine_profile_id,omitempty"`
	AIEngineSettingsSummary   string                `json:"ai_engine_settings_summary"`
	ResponseContractProfileID string                `json:"response_contract_profile_id,omitempty"`
	ResponseContractSummary   string                `json:"response_contract_summary"`
	MemoryPersonalitySummary  string                `json:"memory_personality_summary"`
	OutputModelRoutingMode    string                `json:"output_model_routing_mode,omitempty"`
	DefaultOutputModelID      string                `json:"default_output_model_id,omitempty"`
	DefaultOutputModelSummary string                `json:"default_output_model_summary,omitempty"`
	Status                    string                `json:"status"`
}

type OrganizationHomePayload struct {
	OrganizationSummary
	Description         string                           `json:"description,omitempty"`
	Departments         []OrganizationDepartmentSummary  `json:"departments,omitempty"`
	OutputModelBindings []OrganizationOutputModelBinding `json:"output_model_bindings,omitempty"`
}

type TeamLeadGuidedAction string

const (
	TeamLeadGuidedActionPlanNextSteps         TeamLeadGuidedAction = "plan_next_steps"
	TeamLeadGuidedActionFocusFirst            TeamLeadGuidedAction = "focus_first"
	TeamLeadGuidedActionReviewSetup           TeamLeadGuidedAction = "review_setup"
	TeamLeadGuidedActionResumeRetainedPackage TeamLeadGuidedAction = "resume_retained_package"
)

type OrganizationCreateRequest struct {
	Name       string                `json:"name"`
	Purpose    string                `json:"purpose"`
	StartMode  OrganizationStartMode `json:"start_mode"`
	TemplateID string                `json:"template_id,omitempty"`
}

type TeamLeadGuidanceRequest struct {
	Action         TeamLeadGuidedAction `json:"action"`
	RequestContext string               `json:"request_context,omitempty"`
}

type OrganizationAIEngineUpdateRequest struct {
	ProfileID string `json:"profile_id"`
}

type DepartmentAIEngineUpdateRequest struct {
	ProfileID                   string `json:"profile_id,omitempty"`
	RevertToOrganizationDefault bool   `json:"revert_to_organization_default,omitempty"`
}

type AgentTypeAIEngineUpdateRequest struct {
	ProfileID      string `json:"profile_id,omitempty"`
	UseTeamDefault bool   `json:"use_team_default,omitempty"`
}

type AgentTypeResponseContractUpdateRequest struct {
	ProfileID                    string `json:"profile_id,omitempty"`
	UseOrganizationOrTeamDefault bool   `json:"use_organization_or_team_default,omitempty"`
}

type OrganizationOutputModelRoutingUpdateRequest struct {
	RoutingMode    string                                        `json:"routing_mode"`
	DefaultModelID string                                        `json:"default_model_id,omitempty"`
	Bindings       []OrganizationOutputModelBindingUpdateRequest `json:"bindings,omitempty"`
}

type OrganizationOutputModelBindingUpdateRequest struct {
	OutputTypeID           string `json:"output_type_id"`
	ModelID                string `json:"model_id,omitempty"`
	UseOrganizationDefault bool   `json:"use_organization_default,omitempty"`
}

type ResponseContractUpdateRequest struct {
	ProfileID string `json:"profile_id"`
}

type TeamLeadGuidanceResponse struct {
	Action             TeamLeadGuidedAction       `json:"action"`
	RequestLabel       string                     `json:"request_label"`
	Headline           string                     `json:"headline"`
	Summary            string                     `json:"summary"`
	PrioritySteps      []string                   `json:"priority_steps"`
	SuggestedFollowUps []string                   `json:"suggested_follow_ups"`
	ExecutionContract  *TeamLeadExecutionContract `json:"execution_contract,omitempty"`
}
