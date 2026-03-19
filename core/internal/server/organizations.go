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

type OrganizationTemplateSummary struct {
	ID                       string `json:"id"`
	Name                     string `json:"name"`
	Description              string `json:"description"`
	OrganizationType         string `json:"organization_type"`
	TeamLeadLabel            string `json:"team_lead_label"`
	AdvisorCount             int    `json:"advisor_count"`
	DepartmentCount          int    `json:"department_count"`
	SpecialistCount          int    `json:"specialist_count"`
	AIEngineSettingsSummary  string `json:"ai_engine_settings_summary"`
	MemoryPersonalitySummary string `json:"memory_personality_summary"`
}

type OrganizationSummary struct {
	ID                       string                `json:"id"`
	Name                     string                `json:"name"`
	Purpose                  string                `json:"purpose"`
	StartMode                OrganizationStartMode `json:"start_mode"`
	TemplateID               string                `json:"template_id,omitempty"`
	TemplateName             string                `json:"template_name,omitempty"`
	TeamLeadLabel            string                `json:"team_lead_label"`
	AdvisorCount             int                   `json:"advisor_count"`
	DepartmentCount          int                   `json:"department_count"`
	SpecialistCount          int                   `json:"specialist_count"`
	AIEngineSettingsSummary  string                `json:"ai_engine_settings_summary"`
	MemoryPersonalitySummary string                `json:"memory_personality_summary"`
	Status                   string                `json:"status"`
}

type OrganizationHomePayload struct {
	OrganizationSummary
	Description string `json:"description,omitempty"`
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

type TeamLeadGuidanceResponse struct {
	Action             TeamLeadGuidedAction `json:"action"`
	RequestLabel       string               `json:"request_label"`
	Headline           string               `json:"headline"`
	Summary            string               `json:"summary"`
	PrioritySteps      []string             `json:"priority_steps"`
	SuggestedFollowUps []string             `json:"suggested_follow_ups"`
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
	for _, team := range bundle.Teams {
		specialistCount += len(team.Members)
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
		AIEngineSettingsSummary:  summarizeAIEngineSettings(bundle.ProviderPolicy),
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
	home := OrganizationHomePayload{
		OrganizationSummary: OrganizationSummary{
			ID:                       uuid.NewString(),
			Name:                     strings.TrimSpace(req.Name),
			Purpose:                  strings.TrimSpace(req.Purpose),
			StartMode:                req.StartMode,
			Status:                   "ready",
			TeamLeadLabel:            "Team Lead",
			AIEngineSettingsSummary:  "Set up later in Advanced mode",
			MemoryPersonalitySummary: "Set up later in Advanced mode",
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
		home.MemoryPersonalitySummary = template.MemoryPersonalitySummary
		home.Description = template.Description
	}

	return home
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

	respondAPIJSON(w, http.StatusOK, protocol.NewAPISuccess(home))
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
