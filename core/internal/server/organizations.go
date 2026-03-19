package server

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"sync"
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

const (
	OrganizationAIEngineProfileStarterDefaults OrganizationAIEngineProfileID = "starter_defaults"
	OrganizationAIEngineProfileBalanced        OrganizationAIEngineProfileID = "balanced"
	OrganizationAIEngineProfileHighReasoning   OrganizationAIEngineProfileID = "high_reasoning"
	OrganizationAIEngineProfileFastLightweight OrganizationAIEngineProfileID = "fast_lightweight"
	OrganizationAIEngineProfileDeepPlanning    OrganizationAIEngineProfileID = "deep_planning"
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
	Status                    string                `json:"status"`
}

type OrganizationHomePayload struct {
	OrganizationSummary
	Description string                          `json:"description,omitempty"`
	Departments []OrganizationDepartmentSummary `json:"departments,omitempty"`
}

type TeamLeadGuidedAction string

const (
	TeamLeadGuidedActionPlanNextSteps TeamLeadGuidedAction = "plan_next_steps"
	TeamLeadGuidedActionFocusFirst    TeamLeadGuidedAction = "focus_first"
	TeamLeadGuidedActionReviewSetup   TeamLeadGuidedAction = "review_setup"
)

type OrganizationCreateRequest struct {
	Name       string                `json:"name"`
	Purpose    string                `json:"purpose"`
	StartMode  OrganizationStartMode `json:"start_mode"`
	TemplateID string                `json:"template_id,omitempty"`
}

type TeamLeadGuidanceRequest struct {
	Action TeamLeadGuidedAction `json:"action"`
}

type OrganizationAIEngineUpdateRequest struct {
	ProfileID string `json:"profile_id"`
}

type DepartmentAIEngineUpdateRequest struct {
	ProfileID                   string `json:"profile_id,omitempty"`
	RevertToOrganizationDefault bool   `json:"revert_to_organization_default,omitempty"`
}

type ResponseContractUpdateRequest struct {
	ProfileID string `json:"profile_id"`
}

type TeamLeadGuidanceResponse struct {
	Action             TeamLeadGuidedAction `json:"action"`
	RequestLabel       string               `json:"request_label"`
	Headline           string               `json:"headline"`
	Summary            string               `json:"summary"`
	PrioritySteps      []string             `json:"priority_steps"`
	SuggestedFollowUps []string             `json:"suggested_follow_ups"`
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
) []OrganizationAgentTypeProfileSummary {
	if len(profiles) == 0 {
		return nil
	}

	normalized := make([]OrganizationAgentTypeProfileSummary, 0, len(profiles))
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
		},
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

	response, err := buildTeamLeadGuidance(home, req.Action)
	if err != nil {
		respondAPIError(w, err.Error(), http.StatusBadRequest)
		return
	}

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(response))
}

func buildTeamLeadGuidance(home OrganizationHomePayload, action TeamLeadGuidedAction) (TeamLeadGuidanceResponse, error) {
	organizationName := safeOrganizationName(home.Name)
	teamLeadLabel := safeTeamLeadLabel(home.TeamLeadLabel)
	purposeText := safePurposeText(home.Purpose)

	switch action {
	case TeamLeadGuidedActionPlanNextSteps:
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
		}, nil
	case TeamLeadGuidedActionFocusFirst:
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
		}, nil
	case TeamLeadGuidedActionReviewSetup:
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
		}, nil
	default:
		return TeamLeadGuidanceResponse{}, fmt.Errorf("action must be plan_next_steps, focus_first, or review_setup")
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
