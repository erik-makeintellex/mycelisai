package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/bootstrap"
	"github.com/mycelis/core/internal/swarm"
	"github.com/mycelis/core/pkg/protocol"
)

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

type outputModelCatalogSeed struct {
	ModelID        string
	Label          string
	Summary        string
	OutputTypeIDs  []OrganizationOutputTypeID
	Popular        bool
	HostingFit     string
	PopularityNote string
}

var defaultOrganizationOutputModelBindings = []OrganizationOutputModelBinding{
	{OutputTypeID: string(OrganizationOutputTypeGeneralText), OutputTypeLabel: "General text", ModelID: "qwen3:8b"},
	{OutputTypeID: string(OrganizationOutputTypeResearchReasoning), OutputTypeLabel: "Research & reasoning", ModelID: "llama3.1:8b"},
	{OutputTypeID: string(OrganizationOutputTypeCodeGeneration), OutputTypeLabel: "Code generation", ModelID: "qwen2.5-coder:7b"},
	{OutputTypeID: string(OrganizationOutputTypeVisionAnalysis), OutputTypeLabel: "Vision analysis", ModelID: "llava:7b"},
}

var curatedOutputModelCatalog = []outputModelCatalogSeed{
	{
		ModelID:       "qwen3:14b",
		Label:         "Qwen3 14B",
		Summary:       "Higher-capacity local reasoning and planning model when latency and memory budget allow it.",
		OutputTypeIDs: []OrganizationOutputTypeID{OrganizationOutputTypeGeneralText, OrganizationOutputTypeResearchReasoning},
		HostingFit:    "Best fit when the host can spare roughly 9GB for a stronger local reasoning lane.",
	},
	{
		ModelID:        "qwen3:8b",
		Label:          "Qwen3 8B",
		Summary:        "Strong local-first default for general text, agent planning, and multi-step reasoning.",
		OutputTypeIDs:  []OrganizationOutputTypeID{OrganizationOutputTypeGeneralText, OrganizationOutputTypeResearchReasoning},
		Popular:        true,
		HostingFit:     "Fits well on the current self-hosted GPU class and is already a common local-first general model.",
		PopularityNote: "Official Ollama library surfaces Qwen3 as a heavily used local model family.",
	},
	{
		ModelID:        "llama3.1:8b",
		Label:          "Llama 3.1 8B",
		Summary:        "Popular local general model with long context and strong multilingual/research-oriented posture.",
		OutputTypeIDs:  []OrganizationOutputTypeID{OrganizationOutputTypeGeneralText, OrganizationOutputTypeResearchReasoning},
		Popular:        true,
		HostingFit:     "Fits well on the current self-hosted GPU class and gives a strong second general-purpose local option.",
		PopularityNote: "Official Ollama library shows Llama 3.1 as one of the most widely downloaded local families.",
	},
	{
		ModelID:       "qwen2.5-coder:14b",
		Label:         "Qwen2.5 Coder 14B",
		Summary:       "Stronger local code model for implementation, code repair, and website or application build tasks when the host can afford the larger model.",
		OutputTypeIDs: []OrganizationOutputTypeID{OrganizationOutputTypeCodeGeneration},
		HostingFit:    "Best fit for heavier code/test generation on the current 16GB-class local host when latency is acceptable.",
	},
	{
		ModelID:       "qwen2.5-coder:7b",
		Label:         "Qwen2.5 Coder 7B",
		Summary:       "Focused local model for code generation, code repair, and implementation-heavy team lanes.",
		OutputTypeIDs: []OrganizationOutputTypeID{OrganizationOutputTypeCodeGeneration},
		Popular:       false,
		HostingFit:    "Fits well on the current self-hosted GPU class and aligns with the current local coding default.",
	},
	{
		ModelID:       "deepseek-coder-v2:16b",
		Label:         "DeepSeek Coder V2 16B",
		Summary:       "Alternative local coding specialist for implementation-heavy and repair-heavy asks when a second code model is useful for comparison.",
		OutputTypeIDs: []OrganizationOutputTypeID{OrganizationOutputTypeCodeGeneration},
		HostingFit:    "Useful as a second code-generation candidate when the host can keep a larger specialist loaded.",
	},
	{
		ModelID:       "llava:7b",
		Label:         "LLaVA 7B",
		Summary:       "Local multimodal model for image understanding, OCR, and visual review work.",
		OutputTypeIDs: []OrganizationOutputTypeID{OrganizationOutputTypeVisionAnalysis},
		Popular:       false,
		HostingFit:    "Fits the current self-hosted GPU class for vision analysis without requiring a separate cloud path.",
	},
	{
		ModelID:       "gemma3:12b",
		Label:         "Gemma 3 12B",
		Summary:       "Local multimodal alternative for long-context text plus image review when the host can run a larger vision-capable model.",
		OutputTypeIDs: []OrganizationOutputTypeID{OrganizationOutputTypeVisionAnalysis, OrganizationOutputTypeResearchReasoning},
		HostingFit:    "Good future pull candidate for a stronger single-GPU multimodal review lane.",
	},
}

type TeamLeadExecutionMode string

const (
	TeamLeadExecutionModeGuidedReview             TeamLeadExecutionMode = "guided_review"
	TeamLeadExecutionModeNativeTeam               TeamLeadExecutionMode = "native_team"
	TeamLeadExecutionModeExternalWorkflowContract TeamLeadExecutionMode = "external_workflow_contract"
	TeamLeadExecutionModeContinuityResume         TeamLeadExecutionMode = "continuity_resume"
)

type TeamLeadExecutionContract struct {
	ExecutionMode              TeamLeadExecutionMode       `json:"execution_mode"`
	OwnerLabel                 string                      `json:"owner_label"`
	Summary                    string                      `json:"summary"`
	ContinuityLabel            string                      `json:"continuity_label,omitempty"`
	ContinuitySummary          string                      `json:"continuity_summary,omitempty"`
	ResumeCheckpoint           string                      `json:"resume_checkpoint,omitempty"`
	TeamName                   string                      `json:"team_name,omitempty"`
	ExternalTarget             string                      `json:"external_target,omitempty"`
	CoordinationModel          string                      `json:"coordination_model,omitempty"`
	RecommendedTeamShape       string                      `json:"recommended_team_shape,omitempty"`
	RecommendedTeamCount       int                         `json:"recommended_team_count,omitempty"`
	RecommendedTeamMemberLimit int                         `json:"recommended_team_member_limit,omitempty"`
	TargetOutputs              []string                    `json:"target_outputs"`
	WorkflowGroup              *TeamLeadWorkflowGroupDraft `json:"workflow_group,omitempty"`
}

type TeamLeadWorkflowGroupDraft struct {
	Name                   string   `json:"name"`
	GoalStatement          string   `json:"goal_statement"`
	WorkMode               string   `json:"work_mode"`
	CoordinatorProfile     string   `json:"coordinator_profile"`
	AllowedCapabilities    []string `json:"allowed_capabilities,omitempty"`
	RecommendedMemberLimit int      `json:"recommended_member_limit,omitempty"`
	ExpiryHours            int      `json:"expiry_hours,omitempty"`
	Summary                string   `json:"summary"`
}

var organizationAIEngineProfiles = []organizationAIEngineProfile{
	{
		ID:          OrganizationAIEngineProfileStarterDefaults,
		Summary:     "Starter Defaults",
		Description: "Keeps the guided starter profile that came with this AI Organization.",
		BestFor:     "Helpful when the organization is still settling into its first operating rhythm.",
	},
	{
		ID:          OrganizationAIEngineProfileBalanced,
		Summary:     "Balanced",
		Description: "Steady planning depth and response quality for everyday work.",
		BestFor:     "Best for day-to-day Team Lead guidance and general organization coordination.",
	},
	{
		ID:          OrganizationAIEngineProfileHighReasoning,
		Summary:     "High Reasoning",
		Description: "Adds more careful thinking when planning and tradeoffs need extra attention.",
		BestFor:     "Best for complex planning, review, and higher-stakes decisions.",
	},
	{
		ID:          OrganizationAIEngineProfileFastLightweight,
		Summary:     "Fast & Lightweight",
		Description: "Keeps responses quick and keeps planning lighter for rapid iteration.",
		BestFor:     "Best for fast-moving loops, check-ins, and quick coordination.",
	},
	{
		ID:          OrganizationAIEngineProfileDeepPlanning,
		Summary:     "Deep Planning",
		Description: "Leans into longer multi-step planning and more deliberate organization shaping.",
		BestFor:     "Best for designing larger workstreams and sequencing bigger efforts.",
	},
}

var responseContractProfiles = []responseContractProfile{
	{
		ID:        ResponseContractProfileClearBalanced,
		Summary:   "Clear & Balanced",
		ToneStyle: "Straightforward and steady without sounding cold.",
		Structure: "Uses clear sections and practical takeaways when helpful.",
		Verbosity: "Balanced detail with enough context to act confidently.",
		BestFor:   "Best for everyday Team Lead guidance, reviews, and general coordination.",
	},
	{
		ID:        ResponseContractProfileStructuredAnalytical,
		Summary:   "Structured & Analytical",
		ToneStyle: "Measured, methodical, and reasoning-forward.",
		Structure: "Organizes answers into clear steps, comparisons, or frameworks.",
		Verbosity: "Moderate-to-detailed when structure improves decision-making.",
		BestFor:   "Best for planning, tradeoffs, diagnosis, and deeper review work.",
	},
	{
		ID:        ResponseContractProfileConciseDirect,
		Summary:   "Concise & Direct",
		ToneStyle: "Focused, efficient, and low-friction.",
		Structure: "Keeps responses short and action-led unless more detail is needed.",
		Verbosity: "Intentionally brief with only the highest-signal details.",
		BestFor:   "Best for quick decisions, status checks, and fast-moving execution loops.",
	},
	{
		ID:        ResponseContractProfileWarmSupportive,
		Summary:   "Warm & Supportive",
		ToneStyle: "Encouraging, collaborative, and reassuring.",
		Structure: "Still organized, but written to feel more human and supportive.",
		Verbosity: "Balanced detail with a little more guidance and framing.",
		BestFor:   "Best for onboarding, operator guidance, and people-facing support work.",
	},
}

type OrganizationStore struct {
	mu    sync.RWMutex
	items map[string]OrganizationHomePayload
}

func NewOrganizationStore() *OrganizationStore {
	return &OrganizationStore{items: make(map[string]OrganizationHomePayload)}
}

func (s *OrganizationStore) List() []OrganizationSummary {
	s.mu.RLock()
	defer s.mu.RUnlock()

	summaries := make([]OrganizationSummary, 0, len(s.items))
	for _, item := range s.items {
		summaries = append(summaries, item.OrganizationSummary)
	}
	sort.Slice(summaries, func(i, j int) bool {
		return summaries[i].Name < summaries[j].Name
	})
	return summaries
}

func (s *OrganizationStore) Save(home OrganizationHomePayload) OrganizationHomePayload {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[home.ID] = home
	return home
}

func (s *OrganizationStore) Get(id string) (OrganizationHomePayload, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[id]
	return item, ok
}

func (s *OrganizationStore) Update(id string, update func(OrganizationHomePayload) OrganizationHomePayload) (OrganizationHomePayload, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[id]
	if !ok {
		return OrganizationHomePayload{}, false
	}

	item = update(item)
	s.items[id] = item
	return item, true
}

func (s *AdminServer) templateBundlesPath() string {
	if strings.TrimSpace(s.TemplateBundlesPath) != "" {
		return s.TemplateBundlesPath
	}
	return "config/templates"
}

func (s *AdminServer) organizationStore() *OrganizationStore {
	if s.Organizations == nil {
		s.Organizations = NewOrganizationStore()
	}
	return s.Organizations
}

func (s *AdminServer) listLocalOllamaModelIDs() []string {
	if s.Cognitive == nil || s.Cognitive.Config == nil {
		return nil
	}
	provider, ok := s.Cognitive.Config.Providers["ollama"]
	if !ok {
		return nil
	}
	baseURL := trimOllamaBaseURL(provider.Endpoint)
	if baseURL == "" {
		return nil
	}

	req, err := http.NewRequest(http.MethodGet, baseURL+"/api/tags", nil)
	if err != nil {
		return nil
	}

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var payload struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil
	}

	modelIDs := make([]string, 0, len(payload.Models))
	for _, model := range payload.Models {
		if name := strings.TrimSpace(model.Name); name != "" {
			modelIDs = append(modelIDs, name)
		}
	}
	sort.Strings(modelIDs)
	return modelIDs
}

func (s *AdminServer) loadOrganizationStarterTemplates() ([]OrganizationTemplateSummary, error) {
	loader := bootstrap.NewTemplateLoader(s.templateBundlesPath())
	bundles, err := loader.LoadBundles()
	if err != nil {
		return nil, err
	}

	templates := make([]OrganizationTemplateSummary, 0, len(bundles))
	for _, bundle := range bundles {
		templates = append(templates, summarizeStarterBundle(bundle))
	}
	sort.Slice(templates, func(i, j int) bool {
		return templates[i].Name < templates[j].Name
	})
	return templates, nil
}

func summarizeStarterBundle(bundle *bootstrap.TemplateBundle) OrganizationTemplateSummary {
	departmentCount := len(bundle.Teams)
	specialistCount := 0
	departments := make([]OrganizationDepartmentSummary, 0, len(bundle.Teams))
	responseContract := defaultResponseContractProfile()
	for _, team := range bundle.Teams {
		memberCount := len(team.Members)
		specialistCount += memberCount
		departments = append(departments, OrganizationDepartmentSummary{
			ID:                team.ID,
			Name:              strings.TrimSpace(team.Name),
			SpecialistCount:   memberCount,
			AgentTypeProfiles: summarizeAgentTypeProfiles(team),
		})
	}

	return OrganizationTemplateSummary{
		ID:                       bundle.ID,
		Name:                     bundle.Name,
		Description:              bundle.Description,
		OrganizationType:         "AI Organization starter",
		TeamLeadLabel:            "Team Lead",
		AdvisorCount:             countAdvisors(bundle.Council.Mode),
		DepartmentCount:          departmentCount,
		SpecialistCount:          specialistCount,
		Departments:              departments,
		AIEngineSettingsSummary:  summarizeAIEngineSettings(bundle.ProviderPolicy),
		ResponseContractSummary:  responseContract.Summary,
		MemoryPersonalitySummary: summarizeMemoryPersonality(bundle),
	}
}

func countAdvisors(councilMode string) int {
	if strings.TrimSpace(councilMode) == "" || strings.EqualFold(strings.TrimSpace(councilMode), "disabled") {
		return 0
	}
	return 1
}

func summarizeAIEngineSettings(policy swarm.ProviderPolicy) string {
	if strings.TrimSpace(policy.Provider) == "" {
		return "Starter defaults included"
	}
	return "Starter defaults included"
}

func defaultAIEngineProfileID(startMode OrganizationStartMode, template *OrganizationTemplateSummary) string {
	if template != nil {
		return string(OrganizationAIEngineProfileStarterDefaults)
	}
	if startMode == OrganizationStartModeTemplate {
		return string(OrganizationAIEngineProfileStarterDefaults)
	}
	return ""
}

func lookupOrganizationAIEngineProfile(id string) (organizationAIEngineProfile, bool) {
	for _, profile := range organizationAIEngineProfiles {
		if string(profile.ID) == strings.TrimSpace(id) {
			return profile, true
		}
	}
	return organizationAIEngineProfile{}, false
}

func organizationAIEngineSummaryForProfile(id string) string {
	profile, ok := lookupOrganizationAIEngineProfile(id)
	if !ok {
		return "Set up later in Advanced mode"
	}
	return profile.Summary
}

func defaultOrganizationOutputModelID() string {
	return "qwen2.5-coder:7b-instruct"
}

func normalizeOrganizationOutputModelRoutingMode(value string) OrganizationOutputModelRoutingMode {
	switch OrganizationOutputModelRoutingMode(strings.TrimSpace(value)) {
	case OrganizationOutputModelRoutingModeDetectedOutputTypes:
		return OrganizationOutputModelRoutingModeDetectedOutputTypes
	default:
		return OrganizationOutputModelRoutingModeSingleModel
	}
}

func outputTypeLabel(id string) string {
	switch OrganizationOutputTypeID(strings.TrimSpace(id)) {
	case OrganizationOutputTypeResearchReasoning:
		return "Research & reasoning"
	case OrganizationOutputTypeCodeGeneration:
		return "Code generation"
	case OrganizationOutputTypeVisionAnalysis:
		return "Vision analysis"
	default:
		return "General text"
	}
}

func canonicalOutputTypeIDs() []OrganizationOutputTypeID {
	return []OrganizationOutputTypeID{
		OrganizationOutputTypeGeneralText,
		OrganizationOutputTypeResearchReasoning,
		OrganizationOutputTypeCodeGeneration,
		OrganizationOutputTypeVisionAnalysis,
	}
}

func outputModelLabel(modelID string) string {
	normalized := strings.TrimSpace(modelID)
	if normalized == "" {
		return "Set up later in Advanced mode"
	}
	for _, entry := range curatedOutputModelCatalog {
		if matchesCatalogModel(normalized, entry.ModelID) {
			return entry.Label
		}
	}
	return normalized
}

func outputModelBindingsMap(bindings []OrganizationOutputModelBinding) map[string]OrganizationOutputModelBinding {
	result := make(map[string]OrganizationOutputModelBinding, len(bindings))
	for _, binding := range bindings {
		outputTypeID := strings.TrimSpace(binding.OutputTypeID)
		if outputTypeID == "" {
			continue
		}
		result[outputTypeID] = binding
	}
	return result
}

func outputModelAutomaticSelectionCriteria() []string {
	return []string{
		"Prefer an installed self-hosted model that declares fit for the detected output type before suggesting a pull or remote provider.",
		"Prefer higher-capacity local models for planning, research, code generation, and website-building asks when latency and memory budget are acceptable.",
		"Use vision-capable models for image understanding, OCR, and visual review, but do not claim Ollama text or vision models can generate images or voice without a configured media engine.",
		"Keep the operator in control: ask for owner approval before running a model-behavior review or changing the organization's saved routing policy.",
	}
}

func normalizedOrganizationOutputModelBindings(existing []OrganizationOutputModelBinding, defaultModelID string) []OrganizationOutputModelBinding {
	byType := outputModelBindingsMap(existing)
	normalized := make([]OrganizationOutputModelBinding, 0, len(defaultOrganizationOutputModelBindings))
	for _, canonical := range defaultOrganizationOutputModelBindings {
		outputTypeID := strings.TrimSpace(canonical.OutputTypeID)
		binding := canonical
		if existingBinding, ok := byType[outputTypeID]; ok {
			if modelID := strings.TrimSpace(existingBinding.ModelID); modelID != "" {
				binding.ModelID = modelID
			}
		}
		if strings.TrimSpace(binding.ModelID) == "" {
			binding.ModelID = defaultModelID
		}
		binding.OutputTypeLabel = outputTypeLabel(outputTypeID)
		binding.ModelSummary = outputModelLabel(binding.ModelID)
		normalized = append(normalized, binding)
	}
	return normalized
}

func inferAgentTypeOutputType(member protocol.AgentManifest) OrganizationOutputTypeID {
	switch strings.TrimSpace(strings.ToLower(member.Role)) {
	case "research", "researcher", "lead", "planner", "review", "reviewer", "qa", "quality":
		return OrganizationOutputTypeResearchReasoning
	case "builder", "implementer", "delivery", "coder", "developer":
		return OrganizationOutputTypeCodeGeneration
	case "vision", "image", "visualizer", "data_visualizer":
		return OrganizationOutputTypeVisionAnalysis
	default:
		return OrganizationOutputTypeGeneralText
	}
}

func inferAgentTypeOutputTypeFromProfile(profile OrganizationAgentTypeProfileSummary) OrganizationOutputTypeID {
	switch strings.TrimSpace(profile.ID) {
	case "planner", "reviewer", "research-specialist":
		return OrganizationOutputTypeResearchReasoning
	case "delivery-specialist":
		return OrganizationOutputTypeCodeGeneration
	default:
		return OrganizationOutputTypeGeneralText
	}
}

func matchesCatalogModel(candidate, catalogModelID string) bool {
	candidate = strings.TrimSpace(strings.ToLower(candidate))
	catalogModelID = strings.TrimSpace(strings.ToLower(catalogModelID))
	if candidate == "" || catalogModelID == "" {
		return false
	}
	if candidate == catalogModelID {
		return true
	}
	return strings.HasPrefix(candidate, catalogModelID) || strings.HasPrefix(catalogModelID, candidate)
}

func synthesizeOutputModelCatalog(installedModelIDs []string) []OrganizationOutputModelCatalogEntry {
	installedSet := make(map[string]struct{}, len(installedModelIDs))
	for _, id := range installedModelIDs {
		normalized := strings.TrimSpace(id)
		if normalized == "" {
			continue
		}
		installedSet[normalized] = struct{}{}
	}

	out := make([]OrganizationOutputModelCatalogEntry, 0, len(curatedOutputModelCatalog)+len(installedSet))
	for _, seed := range curatedOutputModelCatalog {
		entry := OrganizationOutputModelCatalogEntry{
			ModelID:        seed.ModelID,
			Label:          seed.Label,
			Summary:        seed.Summary,
			ProviderID:     "ollama",
			Popular:        seed.Popular,
			SelfHostable:   true,
			HostingFit:     seed.HostingFit,
			PopularityNote: seed.PopularityNote,
			Source:         "curated_ollama",
		}
		for _, outputTypeID := range seed.OutputTypeIDs {
			entry.OutputTypeIDs = append(entry.OutputTypeIDs, string(outputTypeID))
		}
		for installedID := range installedSet {
			if matchesCatalogModel(installedID, seed.ModelID) {
				entry.Installed = true
				break
			}
		}
		out = append(out, entry)
	}

	for installedID := range installedSet {
		matched := false
		for _, entry := range out {
			if matchesCatalogModel(installedID, entry.ModelID) {
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		out = append(out, OrganizationOutputModelCatalogEntry{
			ModelID:      installedID,
			Label:        installedID,
			Summary:      "Installed in the local Ollama inventory and available for self-hosted routing.",
			ProviderID:   "ollama",
			Installed:    true,
			SelfHostable: true,
			HostingFit:   "Already present in the current local Ollama inventory.",
			Source:       "installed_ollama",
		})
	}

	sort.Slice(out, func(i, j int) bool {
		if out[i].Popular != out[j].Popular {
			return out[i].Popular
		}
		if out[i].Installed != out[j].Installed {
			return out[i].Installed
		}
		return out[i].Label < out[j].Label
	})
	return out
}

func outputModelReviewCriteria(outputTypeID OrganizationOutputTypeID) []string {
	switch outputTypeID {
	case OrganizationOutputTypeResearchReasoning:
		return []string{
			"prioritize planning depth, synthesis quality, and long-context behavior",
			"prefer installed higher-capacity reasoning models when latency is acceptable",
		}
	case OrganizationOutputTypeCodeGeneration:
		return []string{
			"prioritize implementation accuracy, test repair, and structured code output",
			"prefer coding-specialized models for websites, application code, and developer workflow artifacts",
		}
	case OrganizationOutputTypeVisionAnalysis:
		return []string{
			"prioritize multimodal image understanding, OCR, and visual review reliability",
			"route actual image or voice generation to the configured media engine instead of pretending a vision model can create the binary artifact",
		}
	default:
		return []string{
			"prioritize readable direct answers, broad instruction following, and low-friction drafting",
			"prefer installed general-purpose local models before suggesting a pull or remote provider",
		}
	}
}

func outputModelPreferenceRank(outputTypeID OrganizationOutputTypeID, modelID string) int {
	preferences := map[OrganizationOutputTypeID][]string{
		OrganizationOutputTypeGeneralText:       {"qwen3:14b", "qwen3:8b", "llama3.1:8b"},
		OrganizationOutputTypeResearchReasoning: {"qwen3:14b", "qwen3:8b", "llama3.1:8b", "gemma3:12b"},
		OrganizationOutputTypeCodeGeneration:    {"qwen2.5-coder:14b", "deepseek-coder-v2:16b", "qwen2.5-coder:7b"},
		OrganizationOutputTypeVisionAnalysis:    {"gemma3:12b", "llava:7b"},
	}
	for index, preferred := range preferences[outputTypeID] {
		if matchesCatalogModel(modelID, preferred) {
			return index
		}
	}
	return len(preferences[outputTypeID]) + 100
}

func bestOutputModelCandidate(catalog []OrganizationOutputModelCatalogEntry, outputTypeID OrganizationOutputTypeID) (OrganizationOutputModelCatalogEntry, bool) {
	candidates := make([]OrganizationOutputModelCatalogEntry, 0, len(catalog))
	for _, entry := range catalog {
		for _, candidateTypeID := range entry.OutputTypeIDs {
			if candidateTypeID == string(outputTypeID) {
				candidates = append(candidates, entry)
				break
			}
		}
	}
	if len(candidates) == 0 {
		return OrganizationOutputModelCatalogEntry{}, false
	}
	sort.SliceStable(candidates, func(i, j int) bool {
		if candidates[i].Installed != candidates[j].Installed {
			return candidates[i].Installed
		}
		leftRank := outputModelPreferenceRank(outputTypeID, candidates[i].ModelID)
		rightRank := outputModelPreferenceRank(outputTypeID, candidates[j].ModelID)
		if leftRank != rightRank {
			return leftRank < rightRank
		}
		if candidates[i].Popular != candidates[j].Popular {
			return candidates[i].Popular
		}
		return candidates[i].Label < candidates[j].Label
	})
	return candidates[0], true
}

func outputModelReviewCandidates(catalog []OrganizationOutputModelCatalogEntry) []OrganizationOutputModelReviewCandidate {
	candidates := make([]OrganizationOutputModelReviewCandidate, 0, len(canonicalOutputTypeIDs()))
	for _, outputTypeID := range canonicalOutputTypeIDs() {
		entry, ok := bestOutputModelCandidate(catalog, outputTypeID)
		if !ok {
			continue
		}
		candidates = append(candidates, OrganizationOutputModelReviewCandidate{
			OutputTypeID:    string(outputTypeID),
			OutputTypeLabel: outputTypeLabel(string(outputTypeID)),
			ModelID:         entry.ModelID,
			ModelSummary:    entry.Label,
			Installed:       entry.Installed,
			ReviewCriteria:  outputModelReviewCriteria(outputTypeID),
		})
	}
	return candidates
}

func recommendedOutputModels(catalog []OrganizationOutputModelCatalogEntry) []OrganizationOutputModelCatalogEntry {
	recommended := make([]OrganizationOutputModelCatalogEntry, 0, len(catalog))
	for _, entry := range catalog {
		if entry.Popular {
			recommended = append(recommended, entry)
		}
	}
	return recommended
}

func trimOllamaBaseURL(endpoint string) string {
	base := strings.TrimSpace(endpoint)
	if base == "" {
		return ""
	}
	base = strings.TrimRight(base, "/")
	if strings.HasSuffix(base, "/v1") {
		return strings.TrimSuffix(base, "/v1")
	}
	return base
}

func lookupResponseContractProfile(id string) (responseContractProfile, bool) {
	for _, profile := range responseContractProfiles {
		if string(profile.ID) == strings.TrimSpace(id) {
			return profile, true
		}
	}
	return responseContractProfile{}, false
}

func defaultResponseContractProfile() responseContractProfile {
	return responseContractProfiles[0]
}

func summarizeAgentTypeProfiles(team *swarm.TeamManifest) []OrganizationAgentTypeProfileSummary {
	if team == nil || len(team.Members) == 0 {
		return nil
	}

	profiles := make([]OrganizationAgentTypeProfileSummary, 0, len(team.Members))
	seen := make(map[string]struct{}, len(team.Members))
	for index, member := range team.Members {
		profile := buildAgentTypeProfileSummary(member, index)
		if _, exists := seen[profile.ID]; exists {
			continue
		}
		seen[profile.ID] = struct{}{}
		profiles = append(profiles, profile)
	}
	return profiles
}

func buildAgentTypeProfileSummary(member protocol.AgentManifest, fallbackIndex int) OrganizationAgentTypeProfileSummary {
	id, name, helpsWith := agentTypeProfileIdentity(member, fallbackIndex)
	return OrganizationAgentTypeProfileSummary{
		ID:                               id,
		Name:                             name,
		HelpsWith:                        helpsWith,
		AIEngineBindingProfileID:         inferAgentTypeAIEngineBinding(member),
		ResponseContractBindingProfileID: inferAgentTypeResponseBinding(member),
		OutputTypeID:                     string(inferAgentTypeOutputType(member)),
		OutputTypeLabel:                  outputTypeLabel(string(inferAgentTypeOutputType(member))),
	}
}

func agentTypeProfileIdentity(member protocol.AgentManifest, fallbackIndex int) (string, string, string) {
	role := strings.TrimSpace(strings.ToLower(member.Role))
	switch role {
	case "lead", "planner":
		return "planner", "Planner", "Turns organization goals into practical next steps, delivery sequencing, and clear priorities."
	case "research", "researcher":
		return "research-specialist", "Research Specialist", "Builds the background, options, and supporting context the Team Lead needs before decisions move forward."
	case "review", "reviewer", "qa", "quality":
		return "reviewer", "Reviewer", "Checks work for quality, risk, and readiness before the Team Lead advances the next move."
	case "builder", "implementer", "delivery":
		return "delivery-specialist", "Delivery Specialist", "Carries the work from plan into execution and keeps the main delivery lane moving."
	case "operations", "operator", "coordinator":
		return "operations-specialist", "Operations Specialist", "Keeps follow-through organized, reduces friction, and supports steady execution across the Department."
	case "support", "assistant", "guide":
		return "support-specialist", "Support Specialist", "Helps the Team Lead keep operator requests clear, coordinated, and easy to act on."
	}

	baseName := strings.TrimSpace(member.Role)
	if baseName == "" {
		baseName = strings.TrimSpace(member.ID)
	}
	baseName = humanizeMode(baseName)
	if baseName == "Guided" {
		baseName = fmt.Sprintf("Specialist %d", fallbackIndex+1)
	}
	return slugifyDepartmentID(baseName, fallbackIndex), baseName, "Supports the Department with a focused specialist role when the Team Lead needs more targeted help."
}

func inferAgentTypeAIEngineBinding(member protocol.AgentManifest) string {
	if strings.TrimSpace(member.Model) == "" && strings.TrimSpace(member.Provider) == "" {
		return ""
	}

	switch strings.TrimSpace(strings.ToLower(member.Role)) {
	case "lead", "planner", "review", "reviewer", "qa", "quality":
		return string(OrganizationAIEngineProfileHighReasoning)
	case "research", "researcher":
		return string(OrganizationAIEngineProfileDeepPlanning)
	case "builder", "implementer", "delivery", "operations", "operator", "coordinator":
		return string(OrganizationAIEngineProfileFastLightweight)
	default:
		return string(OrganizationAIEngineProfileBalanced)
	}
}

func inferAgentTypeResponseBinding(member protocol.AgentManifest) string {
	if strings.TrimSpace(member.SystemPrompt) == "" {
		return ""
	}

	switch strings.TrimSpace(strings.ToLower(member.Role)) {
	case "lead", "planner", "review", "reviewer", "qa", "quality":
		return string(ResponseContractProfileStructuredAnalytical)
	case "builder", "implementer", "delivery", "operations", "operator", "coordinator":
		return string(ResponseContractProfileConciseDirect)
	case "support", "assistant", "guide":
		return string(ResponseContractProfileWarmSupportive)
	default:
		return ""
	}
}

func normalizeDepartmentName(name string, fallbackIndex int) string {
	name = strings.TrimSpace(name)
	if name != "" {
		return name
	}
	if fallbackIndex == 0 {
		return "Core Delivery Department"
	}
	return fmt.Sprintf("Department %d", fallbackIndex+1)
}

func normalizeOrganizationHome(home OrganizationHomePayload) OrganizationHomePayload {
	home.OutputModelRoutingMode = string(normalizeOrganizationOutputModelRoutingMode(home.OutputModelRoutingMode))
	if strings.TrimSpace(home.DefaultOutputModelID) == "" {
		home.DefaultOutputModelID = defaultOrganizationOutputModelID()
	}
	home.DefaultOutputModelSummary = outputModelLabel(home.DefaultOutputModelID)
	home.OutputModelBindings = normalizedOrganizationOutputModelBindings(home.OutputModelBindings, home.DefaultOutputModelID)

	if len(home.Departments) == 0 && home.DepartmentCount > 0 {
		home.Departments = generateFallbackDepartments(home.DepartmentCount, home.SpecialistCount)
	}

	if len(home.Departments) > 0 {
		home.DepartmentCount = len(home.Departments)
		totalSpecialists := 0
		for index, department := range home.Departments {
			if department.SpecialistCount <= 0 {
				department.SpecialistCount = spreadSpecialists(home.SpecialistCount, len(home.Departments), index)
			}
			department.Name = normalizeDepartmentName(department.Name, index)
			if department.ID == "" {
				department.ID = slugifyDepartmentID(department.Name, index)
			}

			overrideSummary := ""
			if strings.TrimSpace(department.AIEngineOverrideProfileID) != "" {
				if department.AIEngineOverrideProfileID == home.AIEngineProfileID {
					department.AIEngineOverrideProfileID = ""
				}
				overrideSummary = organizationAIEngineSummaryForProfile(department.AIEngineOverrideProfileID)
				if strings.TrimSpace(overrideSummary) == "Set up later in Advanced mode" {
					department.AIEngineOverrideProfileID = ""
				}
			}

			if strings.TrimSpace(department.AIEngineOverrideProfileID) != "" {
				department.InheritsOrganizationAIEngine = false
				department.AIEngineOverrideSummary = overrideSummary
				department.AIEngineEffectiveProfileID = department.AIEngineOverrideProfileID
				department.AIEngineEffectiveSummary = overrideSummary
			} else {
				department.InheritsOrganizationAIEngine = true
				department.AIEngineOverrideSummary = ""
				department.AIEngineEffectiveProfileID = home.AIEngineProfileID
				department.AIEngineEffectiveSummary = home.AIEngineSettingsSummary
			}

			if strings.TrimSpace(department.AIEngineEffectiveSummary) == "" {
				department.AIEngineEffectiveSummary = "Set up later in Advanced mode"
			}
			department.AgentTypeProfiles = normalizeAgentTypeProfiles(
				department.AgentTypeProfiles,
				department.AIEngineEffectiveProfileID,
				department.AIEngineEffectiveSummary,
				home.ResponseContractProfileID,
				home.ResponseContractSummary,
				home.OutputModelRoutingMode,
				home.DefaultOutputModelID,
				home.DefaultOutputModelSummary,
				home.OutputModelBindings,
			)

			home.Departments[index] = department
			totalSpecialists += department.SpecialistCount
		}
		home.SpecialistCount = totalSpecialists
	}

	return home
}

func normalizeAgentTypeProfiles(
	profiles []OrganizationAgentTypeProfileSummary,
	departmentAIEngineProfileID string,
	departmentAIEngineSummary string,
	responseContractProfileID string,
	responseContractSummary string,
	outputModelRoutingMode string,
	defaultOutputModelID string,
	defaultOutputModelSummary string,
	outputModelBindings []OrganizationOutputModelBinding,
) []OrganizationAgentTypeProfileSummary {
	if len(profiles) == 0 {
		return nil
	}

	normalized := make([]OrganizationAgentTypeProfileSummary, 0, len(profiles))
	bindingsByType := outputModelBindingsMap(outputModelBindings)
	for index, profile := range profiles {
		profile.Name = strings.TrimSpace(profile.Name)
		if profile.Name == "" {
			profile.Name = fmt.Sprintf("Agent Type %d", index+1)
		}
		if profile.ID == "" {
			profile.ID = slugifyDepartmentID(profile.Name, index)
		}
		profile.HelpsWith = strings.TrimSpace(profile.HelpsWith)
		if profile.HelpsWith == "" {
			profile.HelpsWith = "Supports the Department with focused specialist work when the Team Lead needs more targeted help."
		}

		if binding, ok := lookupOrganizationAIEngineProfile(profile.AIEngineBindingProfileID); ok {
			profile.InheritsDepartmentAIEngine = false
			profile.AIEngineBindingProfileID = string(binding.ID)
			profile.AIEngineEffectiveProfileID = string(binding.ID)
			profile.AIEngineEffectiveSummary = binding.Summary
		} else {
			profile.InheritsDepartmentAIEngine = true
			profile.AIEngineBindingProfileID = ""
			profile.AIEngineEffectiveProfileID = departmentAIEngineProfileID
			profile.AIEngineEffectiveSummary = departmentAIEngineSummary
		}
		if strings.TrimSpace(profile.AIEngineEffectiveSummary) == "" {
			profile.AIEngineEffectiveSummary = "Set up later in Advanced mode"
		}

		if binding, ok := lookupResponseContractProfile(profile.ResponseContractBindingProfileID); ok {
			profile.InheritsDefaultResponseContract = false
			profile.ResponseContractBindingProfileID = string(binding.ID)
			profile.ResponseContractEffectiveProfileID = string(binding.ID)
			profile.ResponseContractEffectiveSummary = binding.Summary
		} else {
			profile.InheritsDefaultResponseContract = true
			profile.ResponseContractBindingProfileID = ""
			profile.ResponseContractEffectiveProfileID = responseContractProfileID
			profile.ResponseContractEffectiveSummary = responseContractSummary
		}
		if strings.TrimSpace(profile.ResponseContractEffectiveSummary) == "" {
			profile.ResponseContractEffectiveProfileID = string(defaultResponseContractProfile().ID)
			profile.ResponseContractEffectiveSummary = defaultResponseContractProfile().Summary
		}

		outputTypeID := strings.TrimSpace(profile.OutputTypeID)
		if outputTypeID == "" {
			outputTypeID = string(inferAgentTypeOutputTypeFromProfile(profile))
		}
		profile.OutputTypeID = outputTypeID
		profile.OutputTypeLabel = outputTypeLabel(outputTypeID)

		profile.OutputModelEffectiveID = defaultOutputModelID
		profile.OutputModelEffectiveSummary = defaultOutputModelSummary
		profile.InheritsDefaultOutputModel = true
		if normalizeOrganizationOutputModelRoutingMode(outputModelRoutingMode) == OrganizationOutputModelRoutingModeDetectedOutputTypes {
			if binding, ok := bindingsByType[outputTypeID]; ok && strings.TrimSpace(binding.ModelID) != "" {
				profile.OutputModelEffectiveID = binding.ModelID
				profile.OutputModelEffectiveSummary = binding.ModelSummary
				profile.InheritsDefaultOutputModel = strings.EqualFold(strings.TrimSpace(binding.ModelID), strings.TrimSpace(defaultOutputModelID))
			}
		}
		if strings.TrimSpace(profile.OutputModelEffectiveSummary) == "" {
			profile.OutputModelEffectiveSummary = outputModelLabel(profile.OutputModelEffectiveID)
		}

		normalized = append(normalized, profile)
	}

	return normalized
}

func generateFallbackDepartments(departmentCount, specialistCount int) []OrganizationDepartmentSummary {
	if departmentCount <= 0 {
		return nil
	}

	departments := make([]OrganizationDepartmentSummary, 0, departmentCount)
	names := []string{"Core Delivery Department", "Planning Department", "Operations Department", "Support Department"}
	for index := 0; index < departmentCount; index++ {
		name := names[min(index, len(names)-1)]
		if index >= len(names) {
			name = fmt.Sprintf("Department %d", index+1)
		}
		departments = append(departments, OrganizationDepartmentSummary{
			ID:              slugifyDepartmentID(name, index),
			Name:            name,
			SpecialistCount: spreadSpecialists(specialistCount, departmentCount, index),
		})
	}
	return departments
}

func slugifyDepartmentID(name string, fallbackIndex int) string {
	name = strings.TrimSpace(strings.ToLower(name))
	if name == "" {
		return fmt.Sprintf("department-%d", fallbackIndex+1)
	}
	var b strings.Builder
	lastDash := false
	for _, r := range name {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			lastDash = false
			continue
		}
		if !lastDash {
			b.WriteRune('-')
			lastDash = true
		}
	}
	value := strings.Trim(b.String(), "-")
	if value == "" {
		return fmt.Sprintf("department-%d", fallbackIndex+1)
	}
	return value
}

func summarizeMemoryPersonality(bundle *bootstrap.TemplateBundle) string {
	if strings.TrimSpace(bundle.Kernel.Mode) == "" {
		return "Starter defaults included"
	}
	return fmt.Sprintf("Prepared for %s work", humanizeMode(bundle.Kernel.Mode))
}

func humanizeMode(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "guided"
	}

	var out []rune
	lastWasSpace := false
	for _, r := range value {
		if r == '-' || r == '_' || unicode.IsSpace(r) {
			if !lastWasSpace {
				out = append(out, ' ')
				lastWasSpace = true
			}
			continue
		}
		out = append(out, unicode.ToLower(r))
		lastWasSpace = false
	}

	normalized := strings.TrimSpace(string(out))
	if normalized == "" {
		return "guided"
	}

	words := strings.Fields(normalized)
	for i, word := range words {
		if word == "ai" {
			words[i] = "AI"
			continue
		}
		runes := []rune(word)
		runes[0] = unicode.ToUpper(runes[0])
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

func (s *AdminServer) resolveStarterTemplate(id string) (*OrganizationTemplateSummary, error) {
	templates, err := s.loadOrganizationStarterTemplates()
	if err != nil {
		return nil, err
	}
	for _, template := range templates {
		if template.ID == id {
			clone := template
			return &clone, nil
		}
	}
	return nil, os.ErrNotExist
}

func (s *AdminServer) buildOrganizationHome(req OrganizationCreateRequest, template *OrganizationTemplateSummary) OrganizationHomePayload {
	responseContract := defaultResponseContractProfile()
	home := OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                        uuid.NewString(),
			Name:                      strings.TrimSpace(req.Name),
			Purpose:                   strings.TrimSpace(req.Purpose),
			StartMode:                 req.StartMode,
			Status:                    "ready",
			TeamLeadLabel:             "Team Lead",
			AIEngineSettingsSummary:   "Set up later in Advanced mode",
			ResponseContractProfileID: string(responseContract.ID),
			ResponseContractSummary:   responseContract.Summary,
			MemoryPersonalitySummary:  "Set up later in Advanced mode",
			AIEngineProfileID:         defaultAIEngineProfileID(req.StartMode, template),
			OutputModelRoutingMode:    string(OrganizationOutputModelRoutingModeSingleModel),
			DefaultOutputModelID:      defaultOrganizationOutputModelID(),
		},
		OutputModelBindings: append([]OrganizationOutputModelBinding(nil), defaultOrganizationOutputModelBindings...),
	}

	if template != nil {
		home.TemplateID = template.ID
		home.TemplateName = template.Name
		home.TeamLeadLabel = template.TeamLeadLabel
		home.AdvisorCount = template.AdvisorCount
		home.DepartmentCount = template.DepartmentCount
		home.SpecialistCount = template.SpecialistCount
		home.AIEngineSettingsSummary = template.AIEngineSettingsSummary
		home.ResponseContractSummary = template.ResponseContractSummary
		home.MemoryPersonalitySummary = template.MemoryPersonalitySummary
		home.Description = template.Description
		home.Departments = append([]OrganizationDepartmentSummary(nil), template.Departments...)
	}

	return normalizeOrganizationHome(home)
}

func (s *AdminServer) handleListOrganizations(w http.ResponseWriter, r *http.Request) {
	summaries := s.organizationStore().List()
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(summaries))
}

func (s *AdminServer) emitReviewLoopEvent(orgID string, eventKind ReviewLoopEventKind) {
	if _, err := s.triggerReviewLoopsForEvent(orgID, eventKind); err != nil {
		log.Printf("[review-loop-event] organization=%s event=%s skipped error=%v", orgID, eventKind, err)
	}
}

func (s *AdminServer) handleCreateOrganization(w http.ResponseWriter, r *http.Request) {
	var req OrganizationCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "invalid organization create request", http.StatusBadRequest)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.Purpose = strings.TrimSpace(req.Purpose)
	req.TemplateID = strings.TrimSpace(req.TemplateID)

	if req.Name == "" {
		respondAPIError(w, "organization name is required", http.StatusBadRequest)
		return
	}
	if req.Purpose == "" {
		respondAPIError(w, "organization purpose is required", http.StatusBadRequest)
		return
	}
	if req.StartMode != OrganizationStartModeTemplate && req.StartMode != OrganizationStartModeEmpty {
		respondAPIError(w, "start_mode must be template or empty", http.StatusBadRequest)
		return
	}

	var template *OrganizationTemplateSummary
	if req.StartMode == OrganizationStartModeTemplate {
		if req.TemplateID == "" {
			respondAPIError(w, "template_id is required when start_mode is template", http.StatusBadRequest)
			return
		}
		resolved, err := s.resolveStarterTemplate(req.TemplateID)
		if err != nil {
			respondAPIError(w, "starter template not found", http.StatusNotFound)
			return
		}
		template = resolved
	}

	home := s.buildOrganizationHome(req, template)
	home = s.organizationStore().Save(home)
	s.loopProfileStore().EnsureDefaults(home)
	s.emitReviewLoopEvent(home.ID, ReviewLoopEventOrganizationCreated)
	respondAPIJSON(w, http.StatusCreated, protocol.NewAPISuccess(home))
}

func (s *AdminServer) handleGetOrganizationHome(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		respondAPIError(w, "organization id is required", http.StatusBadRequest)
		return
	}

	home, ok := s.organizationStore().Get(id)
	if !ok {
		respondAPIError(w, "organization not found", http.StatusNotFound)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(normalizeOrganizationHome(home)))
}

func (s *AdminServer) handleUpdateOrganizationAIEngine(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		respondAPIError(w, "organization id is required", http.StatusBadRequest)
		return
	}

	var req OrganizationAIEngineUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "invalid AI Engine update request", http.StatusBadRequest)
		return
	}

	profile, ok := lookupOrganizationAIEngineProfile(req.ProfileID)
	if !ok {
		respondAPIError(w, "profile_id must be one of the guided AI Engine options", http.StatusBadRequest)
		return
	}

	updated, ok := s.organizationStore().Update(id, func(home OrganizationHomePayload) OrganizationHomePayload {
		home.AIEngineProfileID = string(profile.ID)
		home.AIEngineSettingsSummary = profile.Summary
		return normalizeOrganizationHome(home)
	})
	if !ok {
		respondAPIError(w, "organization not found", http.StatusNotFound)
		return
	}

	s.emitReviewLoopEvent(updated.ID, ReviewLoopEventOrganizationAIEngineChanged)
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(updated))
}

func (s *AdminServer) handleUpdateResponseContract(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		respondAPIError(w, "organization id is required", http.StatusBadRequest)
		return
	}

	var req ResponseContractUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "invalid Response Contract update request", http.StatusBadRequest)
		return
	}

	profile, ok := lookupResponseContractProfile(req.ProfileID)
	if !ok {
		respondAPIError(w, "profile_id must be one of the guided Response Style options", http.StatusBadRequest)
		return
	}

	updated, ok := s.organizationStore().Update(id, func(home OrganizationHomePayload) OrganizationHomePayload {
		home.ResponseContractProfileID = string(profile.ID)
		home.ResponseContractSummary = profile.Summary
		return normalizeOrganizationHome(home)
	})
	if !ok {
		respondAPIError(w, "organization not found", http.StatusNotFound)
		return
	}

	s.emitReviewLoopEvent(updated.ID, ReviewLoopEventResponseContractChanged)
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(updated))
}

func (s *AdminServer) handleGetOrganizationOutputModelRouting(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		respondAPIError(w, "organization id is required", http.StatusBadRequest)
		return
	}

	home, ok := s.organizationStore().Get(id)
	if !ok {
		respondAPIError(w, "organization not found", http.StatusNotFound)
		return
	}
	home = normalizeOrganizationHome(home)

	catalog := synthesizeOutputModelCatalog(s.listLocalOllamaModelIDs())
	payload := OrganizationOutputModelRoutingPayload{
		RoutingMode:                home.OutputModelRoutingMode,
		DefaultModelID:             home.DefaultOutputModelID,
		DefaultModelSummary:        home.DefaultOutputModelSummary,
		Bindings:                   append([]OrganizationOutputModelBinding(nil), home.OutputModelBindings...),
		AvailableModels:            catalog,
		RecommendedModels:          recommendedOutputModels(catalog),
		ReviewCandidates:           outputModelReviewCandidates(catalog),
		HardwareSummary:            "Local-first self-hosted posture tuned for the current Ollama inventory and a 16GB-class GPU host.",
		ReviewPermissionPrompt:     "Ask the owner/admin before Soma reviews potential model behavior for a requested output or changes saved routing.",
		AutomaticSelectionCriteria: outputModelAutomaticSelectionCriteria(),
	}
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(payload))
}

func (s *AdminServer) handleUpdateOrganizationOutputModelRouting(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		respondAPIError(w, "organization id is required", http.StatusBadRequest)
		return
	}

	var req OrganizationOutputModelRoutingUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "invalid output model routing update request", http.StatusBadRequest)
		return
	}

	routingMode := normalizeOrganizationOutputModelRoutingMode(req.RoutingMode)
	defaultModelID := strings.TrimSpace(req.DefaultModelID)
	if defaultModelID == "" {
		respondAPIError(w, "default_model_id is required", http.StatusBadRequest)
		return
	}

	validOutputTypes := make(map[string]struct{}, len(canonicalOutputTypeIDs()))
	for _, outputTypeID := range canonicalOutputTypeIDs() {
		validOutputTypes[string(outputTypeID)] = struct{}{}
	}

	normalizedBindings := make([]OrganizationOutputModelBinding, 0, len(req.Bindings))
	for _, binding := range req.Bindings {
		outputTypeID := strings.TrimSpace(binding.OutputTypeID)
		if _, ok := validOutputTypes[outputTypeID]; !ok {
			respondAPIError(w, "bindings contain an unknown output_type_id", http.StatusBadRequest)
			return
		}
		modelID := strings.TrimSpace(binding.ModelID)
		if binding.UseOrganizationDefault {
			modelID = defaultModelID
		}
		if modelID == "" {
			respondAPIError(w, "bindings require a model_id unless they use the organization default", http.StatusBadRequest)
			return
		}
		normalizedBindings = append(normalizedBindings, OrganizationOutputModelBinding{
			OutputTypeID:    outputTypeID,
			OutputTypeLabel: outputTypeLabel(outputTypeID),
			ModelID:         modelID,
			ModelSummary:    outputModelLabel(modelID),
		})
	}

	updated, ok := s.organizationStore().Update(id, func(home OrganizationHomePayload) OrganizationHomePayload {
		home = normalizeOrganizationHome(home)
		home.OutputModelRoutingMode = string(routingMode)
		home.DefaultOutputModelID = defaultModelID
		home.DefaultOutputModelSummary = outputModelLabel(defaultModelID)
		home.OutputModelBindings = normalizedOrganizationOutputModelBindings(normalizedBindings, defaultModelID)
		return normalizeOrganizationHome(home)
	})
	if !ok {
		respondAPIError(w, "organization not found", http.StatusNotFound)
		return
	}

	s.emitReviewLoopEvent(updated.ID, ReviewLoopEventOrganizationAIEngineChanged)
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(updated))
}

func (s *AdminServer) handleUpdateDepartmentAIEngine(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	departmentID := strings.TrimSpace(r.PathValue("departmentId"))
	if id == "" || departmentID == "" {
		respondAPIError(w, "organization id and department id are required", http.StatusBadRequest)
		return
	}

	var req DepartmentAIEngineUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "invalid Department AI Engine update request", http.StatusBadRequest)
		return
	}

	profileID := strings.TrimSpace(req.ProfileID)
	if req.RevertToOrganizationDefault {
		profileID = ""
	} else {
		if profileID == "" {
			respondAPIError(w, "profile_id is required unless reverting to the organization default", http.StatusBadRequest)
			return
		}
		if _, ok := lookupOrganizationAIEngineProfile(profileID); !ok {
			respondAPIError(w, "profile_id must be one of the guided AI Engine options", http.StatusBadRequest)
			return
		}
	}

	departmentFound := false
	updated, ok := s.organizationStore().Update(id, func(home OrganizationHomePayload) OrganizationHomePayload {
		home = normalizeOrganizationHome(home)
		for index, department := range home.Departments {
			if department.ID != departmentID {
				continue
			}
			departmentFound = true
			if profileID == "" || profileID == home.AIEngineProfileID {
				department.AIEngineOverrideProfileID = ""
				department.AIEngineOverrideSummary = ""
			} else {
				department.AIEngineOverrideProfileID = profileID
				department.AIEngineOverrideSummary = organizationAIEngineSummaryForProfile(profileID)
			}
			home.Departments[index] = department
			break
		}
		return normalizeOrganizationHome(home)
	})
	if !ok {
		respondAPIError(w, "organization not found", http.StatusNotFound)
		return
	}
	if !departmentFound {
		respondAPIError(w, "department not found", http.StatusNotFound)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(updated))
}

func (s *AdminServer) handleUpdateAgentTypeAIEngine(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	departmentID := strings.TrimSpace(r.PathValue("departmentId"))
	agentTypeID := strings.TrimSpace(r.PathValue("agentTypeId"))
	if id == "" || departmentID == "" || agentTypeID == "" {
		respondAPIError(w, "organization id, department id, and agent type id are required", http.StatusBadRequest)
		return
	}

	var req AgentTypeAIEngineUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "invalid Agent Type AI Engine update request", http.StatusBadRequest)
		return
	}

	profileID := strings.TrimSpace(req.ProfileID)
	if req.UseTeamDefault {
		profileID = ""
	} else {
		if profileID == "" {
			respondAPIError(w, "profile_id is required unless returning to the Team default", http.StatusBadRequest)
			return
		}
		if _, ok := lookupOrganizationAIEngineProfile(profileID); !ok {
			respondAPIError(w, "profile_id must be one of the guided AI Engine options", http.StatusBadRequest)
			return
		}
	}

	departmentFound := false
	agentTypeFound := false
	updated, ok := s.organizationStore().Update(id, func(home OrganizationHomePayload) OrganizationHomePayload {
		home = normalizeOrganizationHome(home)
		for departmentIndex, department := range home.Departments {
			if department.ID != departmentID {
				continue
			}
			departmentFound = true
			for profileIndex, profile := range department.AgentTypeProfiles {
				if profile.ID != agentTypeID {
					continue
				}
				agentTypeFound = true
				if profileID == "" || profileID == department.AIEngineEffectiveProfileID {
					profile.AIEngineBindingProfileID = ""
				} else {
					profile.AIEngineBindingProfileID = profileID
				}
				department.AgentTypeProfiles[profileIndex] = profile
				break
			}
			home.Departments[departmentIndex] = department
			break
		}
		return normalizeOrganizationHome(home)
	})
	if !ok {
		respondAPIError(w, "organization not found", http.StatusNotFound)
		return
	}
	if !departmentFound {
		respondAPIError(w, "department not found", http.StatusNotFound)
		return
	}
	if !agentTypeFound {
		respondAPIError(w, "agent type profile not found", http.StatusNotFound)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(updated))
}

func (s *AdminServer) handleUpdateAgentTypeResponseContract(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	departmentID := strings.TrimSpace(r.PathValue("departmentId"))
	agentTypeID := strings.TrimSpace(r.PathValue("agentTypeId"))
	if id == "" || departmentID == "" || agentTypeID == "" {
		respondAPIError(w, "organization id, department id, and agent type id are required", http.StatusBadRequest)
		return
	}

	var req AgentTypeResponseContractUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "invalid Agent Type Response Style update request", http.StatusBadRequest)
		return
	}

	profileID := strings.TrimSpace(req.ProfileID)
	if req.UseOrganizationOrTeamDefault {
		profileID = ""
	} else {
		if profileID == "" {
			respondAPIError(w, "profile_id is required unless returning to the Organization / Team default", http.StatusBadRequest)
			return
		}
		if _, ok := lookupResponseContractProfile(profileID); !ok {
			respondAPIError(w, "profile_id must be one of the guided Response Style options", http.StatusBadRequest)
			return
		}
	}

	departmentFound := false
	agentTypeFound := false
	updated, ok := s.organizationStore().Update(id, func(home OrganizationHomePayload) OrganizationHomePayload {
		home = normalizeOrganizationHome(home)
		defaultProfileID := strings.TrimSpace(home.ResponseContractProfileID)
		if defaultProfileID == "" {
			defaultProfileID = string(defaultResponseContractProfile().ID)
		}
		for departmentIndex, department := range home.Departments {
			if department.ID != departmentID {
				continue
			}
			departmentFound = true
			for profileIndex, profile := range department.AgentTypeProfiles {
				if profile.ID != agentTypeID {
					continue
				}
				agentTypeFound = true
				if profileID == "" || profileID == defaultProfileID {
					profile.ResponseContractBindingProfileID = ""
				} else {
					profile.ResponseContractBindingProfileID = profileID
				}
				department.AgentTypeProfiles[profileIndex] = profile
				break
			}
			home.Departments[departmentIndex] = department
			break
		}
		return normalizeOrganizationHome(home)
	})
	if !ok {
		respondAPIError(w, "organization not found", http.StatusNotFound)
		return
	}
	if !departmentFound {
		respondAPIError(w, "department not found", http.StatusNotFound)
		return
	}
	if !agentTypeFound {
		respondAPIError(w, "agent type profile not found", http.StatusNotFound)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(updated))
}

func (s *AdminServer) handleTeamLeadGuidedAction(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		respondAPIError(w, "organization id is required", http.StatusBadRequest)
		return
	}

	home, ok := s.organizationStore().Get(id)
	if !ok {
		respondAPIError(w, "organization not found", http.StatusNotFound)
		return
	}

	var req TeamLeadGuidanceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondAPIError(w, "invalid Team Lead action request", http.StatusBadRequest)
		return
	}

	response, err := buildTeamLeadGuidance(home, req.Action, req.RequestContext)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}

	s.emitReviewLoopEvent(home.ID, ReviewLoopEventTeamLeadActionCompleted)
	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(response))
}

func buildTeamLeadGuidance(home OrganizationHomePayload, action TeamLeadGuidedAction, requestContext string) (TeamLeadGuidanceResponse, error) {
	organizationName := safeOrganizationName(home.Name)
	teamLeadLabel := safeTeamLeadLabel(home.TeamLeadLabel)
	purposeText := safePurposeText(home.Purpose)

	switch action {
	case TeamLeadGuidedActionPlanNextSteps:
		executionContract := buildTeamLeadExecutionContract(home, requestContext)
		steps := []string{
			fmt.Sprintf("Align the first outcome with this purpose: %s.", purposeText),
			firstDepartmentStep(home),
			firstSpecialistStep(home),
		}
		return TeamLeadGuidanceResponse{
			Action:        action,
			RequestLabel:  "Plan next steps for this organization",
			Headline:      fmt.Sprintf("Team Lead plan for %s", organizationName),
			Summary:       fmt.Sprintf("%s recommends moving %s from setup into a focused first delivery loop.", teamLeadLabel, organizationName),
			PrioritySteps: steps,
			SuggestedFollowUps: []string{
				"Review my organization setup",
				"What should I focus on first?",
				templateSpecificSuggestion(home),
			},
			ExecutionContract: executionContract,
		}, nil
	case TeamLeadGuidedActionFocusFirst:
		executionContract := buildTeamLeadExecutionContract(home, requestContext)
		return TeamLeadGuidanceResponse{
			Action:       action,
			RequestLabel: "What should I focus on first?",
			Headline:     fmt.Sprintf("First focus for %s", organizationName),
			Summary:      firstFocusSummary(home),
			PrioritySteps: []string{
				firstDepartmentStep(home),
				firstAdvisorStep(home),
				"Keep the Team Lead as the primary working counterpart while the organization takes shape.",
			},
			SuggestedFollowUps: []string{
				"Plan next steps for this organization",
				"Review my organization setup",
				"Review the Team Lead guidance before expanding into deeper structure.",
			},
			ExecutionContract: executionContract,
		}, nil
	case TeamLeadGuidedActionReviewSetup:
		executionContract := buildTeamLeadExecutionContract(home, requestContext)
		return TeamLeadGuidanceResponse{
			Action:       action,
			RequestLabel: "Review my organization setup",
			Headline:     fmt.Sprintf("Organization setup review for %s", organizationName),
			Summary:      fmt.Sprintf("%s is ready to review the current AI Organization shape before the next action begins.", teamLeadLabel),
			PrioritySteps: []string{
				fmt.Sprintf("Advisors: %s.", formatConfiguredCountForGuidance(home.AdvisorCount, "advisor")),
				fmt.Sprintf("Departments: %s.", formatConfiguredCountForGuidance(home.DepartmentCount, "department")),
				fmt.Sprintf("Specialists: %s.", formatConfiguredCountForGuidance(home.SpecialistCount, "specialist")),
			},
			SuggestedFollowUps: []string{
				"Plan next steps for this organization",
				"What should I focus on first?",
				fmt.Sprintf("Review the %s summary and confirm the Team Lead has what it needs.", home.startingPointLabel()),
			},
			ExecutionContract: executionContract,
		}, nil
	case TeamLeadGuidedActionResumeRetainedPackage:
		executionContract := buildRetainedPackageContinuityContract(home, requestContext)
		return TeamLeadGuidanceResponse{
			Action:       action,
			RequestLabel: "Resume retained package continuity",
			Headline:     fmt.Sprintf("Retained package continuity for %s", organizationName),
			Summary:      fmt.Sprintf("%s resumes the retained package for %s so completed work stays durable and the next step stays explicit after a reboot or reload.", teamLeadLabel, organizationName),
			PrioritySteps: []string{
				"Open the retained package and confirm the latest durable outputs.",
				"Record what is already complete, what remains, and who owns the next step.",
				"Continue from the retained package without rebuilding finished work.",
			},
			SuggestedFollowUps: []string{
				"Plan next steps for this organization",
				"Review my organization setup",
				"Use the retained package continuity contract as the starting point for UI or live test automation.",
			},
			ExecutionContract: executionContract,
		}, nil
	default:
		return TeamLeadGuidanceResponse{}, fmt.Errorf("action must be plan_next_steps, focus_first, review_setup, or resume_retained_package")
	}
}

func firstDepartmentStep(home OrganizationHomePayload) string {
	if home.DepartmentCount > 0 {
		return fmt.Sprintf("Use %d Department%s as the first routing layer for work.", home.DepartmentCount, pluralSuffix(home.DepartmentCount))
	}
	return "Define the first Department so the Team Lead has a clear execution lane."
}

func firstSpecialistStep(home OrganizationHomePayload) string {
	if home.SpecialistCount > 0 {
		return fmt.Sprintf("Bring %d Specialist%s in after the Team Lead confirms the plan.", home.SpecialistCount, pluralSuffix(home.SpecialistCount))
	}
	return "Add Specialists only after the Team Lead confirms the first Department-level plan."
}

func firstAdvisorStep(home OrganizationHomePayload) string {
	if home.AdvisorCount > 0 {
		return fmt.Sprintf("Use %d Advisor%s when the Team Lead needs review or decision support.", home.AdvisorCount, pluralSuffix(home.AdvisorCount))
	}
	return "Decide whether advisor guidance is needed before the next planning cycle."
}

func firstFocusSummary(home OrganizationHomePayload) string {
	if home.StartMode == OrganizationStartModeTemplate && strings.TrimSpace(home.TemplateName) != "" {
		return fmt.Sprintf("Start by using %s as the first working shape, then let the Team Lead confirm which part of the organization should lead.", home.TemplateName)
	}
	return fmt.Sprintf("Start by confirming the first outcome this AI Organization should deliver around %s, then let the Team Lead shape the initial structure around that goal.", safePurposeText(home.Purpose))
}

func buildTeamLeadExecutionContract(home OrganizationHomePayload, requestContext string) *TeamLeadExecutionContract {
	normalized := strings.TrimSpace(strings.ToLower(requestContext))
	if normalized == "" {
		return nil
	}

	if referencesExternalWorkflowContract(normalized) {
		return &TeamLeadExecutionContract{
			ExecutionMode:  TeamLeadExecutionModeExternalWorkflowContract,
			OwnerLabel:     "External workflow contract",
			ExternalTarget: externalWorkflowTarget(normalized),
			Summary:        "This request is best handled as an external workflow contract so Mycelis can keep invocation posture, governance, and normalized result return clear without pretending the external graph is a native team.",
			TargetOutputs: []string{
				"Normalized workflow result",
				"Linked artifact or execution note",
			},
		}
	}

	if isBroadCoordinationRequest(normalized) {
		teamName := "Program Orchestration Team"
		outputs := []string{
			"Program orchestration brief",
			"Per-team delivery plans",
			"Cross-team coordination summary",
		}
		return &TeamLeadExecutionContract{
			ExecutionMode:              TeamLeadExecutionModeNativeTeam,
			OwnerLabel:                 "Soma and Council orchestration",
			TeamName:                   teamName,
			CoordinationModel:          "multi_team_orchestration",
			RecommendedTeamShape:       "Several small teams coordinated by Soma and Council over NATS/exchange, with no single team exceeding the member cap.",
			RecommendedTeamCount:       3,
			RecommendedTeamMemberLimit: 5,
			Summary:                    fmt.Sprintf("This request is broad enough to split into several compact teams instead of one oversized group. Use Soma and Council to coordinate the lanes over NATS/exchange, keep each team small, and return one orchestration summary plus the team-level outputs for %s.", safeOrganizationName(home.Name)),
			TargetOutputs:              outputs,
			WorkflowGroup:              buildWorkflowGroupDraft(home, teamName, requestContext, "propose_only", outputs, []string{"team.coordinate", "artifact.review", "broadcast"}, 5),
		}
	}

	if referencesImageTeamOutput(normalized) {
		teamName := "Creative Delivery Team"
		outputs := []string{
			"Reviewable image artifact",
			"Short concept note",
		}
		return &TeamLeadExecutionContract{
			ExecutionMode:              TeamLeadExecutionModeNativeTeam,
			OwnerLabel:                 "Native Mycelis team",
			TeamName:                   teamName,
			CoordinationModel:          "compact_team",
			RecommendedTeamShape:       "One focused team with a small specialist roster.",
			RecommendedTeamCount:       1,
			RecommendedTeamMemberLimit: 6,
			Summary:                    fmt.Sprintf("Use a bounded creative team inside %s so Soma can shape the work, route it through the right specialists, and return the generated image as a managed artifact.", safeOrganizationName(home.Name)),
			TargetOutputs:              outputs,
			WorkflowGroup:              buildWorkflowGroupDraft(home, teamName, requestContext, "propose_only", outputs, []string{"content.plan", "artifact.review"}, 6),
		}
	}

	if teamName, outputs, ok := inferBusinessTeamExecution(normalized); ok {
		return &TeamLeadExecutionContract{
			ExecutionMode:              TeamLeadExecutionModeNativeTeam,
			OwnerLabel:                 "Native Mycelis team",
			TeamName:                   teamName,
			CoordinationModel:          "compact_team",
			RecommendedTeamShape:       "One focused team with a small specialist roster.",
			RecommendedTeamCount:       1,
			RecommendedTeamMemberLimit: 6,
			Summary:                    fmt.Sprintf("Use a bounded %s lane inside %s so Soma can stand up a focused delivery group, coordinate the right specialists, and keep the resulting outputs reviewable in one place.", strings.ToLower(teamName), safeOrganizationName(home.Name)),
			TargetOutputs:              outputs,
			WorkflowGroup:              buildWorkflowGroupDraft(home, teamName, requestContext, "propose_only", outputs, []string{"team.coordinate", "artifact.review"}, 6),
		}
	}

	return nil
}

func buildRetainedPackageContinuityContract(home OrganizationHomePayload, requestContext string) *TeamLeadExecutionContract {
	organizationName := safeOrganizationName(home.Name)
	resumeGoal := strings.TrimSpace(requestContext)
	if resumeGoal == "" {
		resumeGoal = fmt.Sprintf("Resume the retained package for %s after a reboot or reload.", organizationName)
	}

	targetOutputs := []string{
		"Retained package continuity summary",
		"Completed work snapshot",
		"Remaining work checklist",
	}
	return &TeamLeadExecutionContract{
		ExecutionMode:     TeamLeadExecutionModeContinuityResume,
		OwnerLabel:        "Team Lead continuity",
		Summary:           fmt.Sprintf("Resume the retained package for %s, confirm completed work, and keep the remaining steps reviewable after a reboot or reload.", organizationName),
		ContinuityLabel:   "Retained package continuity",
		ContinuitySummary: fmt.Sprintf("Continuity resumes from the last durable outputs for %s without rebuilding finished work.", organizationName),
		ResumeCheckpoint:  "Continue from the last retained package after reload or reboot.",
		TargetOutputs:     targetOutputs,
		WorkflowGroup:     buildWorkflowGroupDraft(home, "Retained Package Continuity", resumeGoal, "resume_continuity", targetOutputs, []string{"artifact.review", "team.coordinate"}, 4),
	}
}

func inferBusinessTeamExecution(normalized string) (string, []string, bool) {
	switch {
	case strings.Contains(normalized, "marketing"):
		return "Marketing Launch Team", []string{
			"Launch plan",
			"Messaging brief",
			"Campaign asset list",
		}, true
	case strings.Contains(normalized, "customer research") || strings.Contains(normalized, "user research"):
		return "Customer Research Team", []string{
			"Research brief",
			"Interview guide",
			"Findings summary",
		}, true
	case strings.Contains(normalized, "revops") || strings.Contains(normalized, "revenue operations") || strings.Contains(normalized, "lead routing"):
		return "Revenue Operations Team", []string{
			"Workflow recommendation",
			"Operational checklist",
			"Handoff summary",
		}, true
	case strings.Contains(normalized, "security"):
		return "Security Review Team", []string{
			"Risk review",
			"Mitigation checklist",
			"Approval notes",
		}, true
	case strings.Contains(normalized, "creative team"):
		return "Creative Delivery Team", []string{
			"Creative brief",
			"Artifact package",
			"Review notes",
		}, true
	default:
		return "", nil, false
	}
}

func isBroadCoordinationRequest(normalized string) bool {
	breadthSignals := 0
	for _, marker := range []string{
		"company-wide",
		"organization-wide",
		"cross-functional",
		"multi-team",
		"multiple teams",
		"several teams",
		"several workstreams",
		"whole organization",
		"enterprise-wide",
		"all teams",
		"across teams",
		"across functions",
	} {
		if strings.Contains(normalized, marker) {
			breadthSignals++
		}
	}
	if breadthSignals >= 1 {
		return true
	}
	return strings.Contains(normalized, "program") && (strings.Contains(normalized, "across") || strings.Contains(normalized, "multiple") || strings.Contains(normalized, "several") || strings.Contains(normalized, "cross"))
}

func buildWorkflowGroupDraft(home OrganizationHomePayload, teamName, requestContext, workMode string, targetOutputs, allowedCapabilities []string, recommendedMemberLimit int) *TeamLeadWorkflowGroupDraft {
	organizationName := safeOrganizationName(home.Name)
	goal := strings.TrimSpace(requestContext)
	if goal == "" {
		goal = fmt.Sprintf("Coordinate a focused %s workflow inside %s.", strings.ToLower(teamName), organizationName)
	}
	return &TeamLeadWorkflowGroupDraft{
		Name:                   fmt.Sprintf("%s temporary workflow", teamName),
		GoalStatement:          goal,
		WorkMode:               workMode,
		CoordinatorProfile:     fmt.Sprintf("%s lead", teamName),
		AllowedCapabilities:    normalizeExecutionCapabilityList(allowedCapabilities),
		RecommendedMemberLimit: recommendedMemberLimit,
		ExpiryHours:            72,
		Summary:                fmt.Sprintf("Launch a temporary workflow group for %s, keep coordination bounded to at most %d members, and retain outputs like %s after the lane is archived.", teamName, recommendedMemberLimit, humanJoin(targetOutputs)),
	}
}

func humanJoin(items []string) string {
	items = normalizeExecutionCapabilityList(items)
	switch len(items) {
	case 0:
		return "planned outputs"
	case 1:
		return items[0]
	case 2:
		return items[0] + " and " + items[1]
	default:
		return strings.Join(items[:len(items)-1], ", ") + ", and " + items[len(items)-1]
	}
}

func normalizeExecutionCapabilityList(items []string) []string {
	if len(items) == 0 {
		return []string{}
	}
	out := make([]string, 0, len(items))
	seen := make(map[string]struct{}, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}

func referencesImageTeamOutput(normalized string) bool {
	return strings.Contains(normalized, "image") ||
		strings.Contains(normalized, "hero art") ||
		strings.Contains(normalized, "hero image") ||
		strings.Contains(normalized, "visual") ||
		strings.Contains(normalized, "illustration") ||
		strings.Contains(normalized, "poster") ||
		strings.Contains(normalized, "moodboard")
}

func referencesExternalWorkflowContract(normalized string) bool {
	return strings.Contains(normalized, "n8n") ||
		strings.Contains(normalized, "workflow contract") ||
		strings.Contains(normalized, "external workflow") ||
		strings.Contains(normalized, "comfyui")
}

func externalWorkflowTarget(normalized string) string {
	switch {
	case strings.Contains(normalized, "n8n"):
		return "n8n workflow contract"
	case strings.Contains(normalized, "comfyui"):
		return "ComfyUI workflow contract"
	default:
		return "External workflow contract"
	}
}

func templateSpecificSuggestion(home OrganizationHomePayload) string {
	if home.StartMode == OrganizationStartModeTemplate && strings.TrimSpace(home.TemplateName) != "" {
		return fmt.Sprintf("Use the %s starter as the first operating guide.", home.TemplateName)
	}
	return "Review how the Team Lead should shape the first Department and Specialist setup."
}

func formatConfiguredCountForGuidance(count int, label string) string {
	if count == 0 {
		return "not configured yet"
	}
	return fmt.Sprintf("%d %s%s ready", count, label, pluralSuffix(count))
}

func spreadSpecialists(total, departmentCount, index int) int {
	if departmentCount <= 0 || total <= 0 {
		return 0
	}
	base := total / departmentCount
	remainder := total % departmentCount
	if index < remainder {
		return base + 1
	}
	return base
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func safeOrganizationName(name string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return "this AI Organization"
	}
	return name
}

func safeTeamLeadLabel(label string) string {
	label = strings.TrimSpace(label)
	if label == "" {
		return "Team Lead"
	}
	return label
}

func safePurposeText(purpose string) string {
	purpose = strings.TrimSpace(purpose)
	if purpose == "" {
		return "the current AI Organization priorities"
	}
	return purpose
}

func pluralSuffix(count int) string {
	if count == 1 {
		return ""
	}
	return "s"
}

func (h OrganizationHomePayload) startingPointLabel() string {
	if h.StartMode == OrganizationStartModeTemplate && strings.TrimSpace(h.TemplateName) != "" {
		return h.TemplateName
	}
	return "starting organization shape"
}
