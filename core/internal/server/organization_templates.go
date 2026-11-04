package server

import (
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/bootstrap"
)

// loadOrganizationStarterTemplates returns the summaries of every available
// catalog entry (top-level bundles plus starters/). Unavailable starters are
// left out so they cannot be used to create an organization.
func (s *AdminServer) loadOrganizationStarterTemplates() ([]OrganizationTemplateSummary, error) {
	catalog, err := s.loadOrganizationStarterCatalog()
	if err != nil {
		return nil, err
	}

	templates := make([]OrganizationTemplateSummary, 0, len(catalog))
	for _, item := range catalog {
		if item.Available {
			templates = append(templates, item.OrganizationTemplateSummary)
		}
	}
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

func (s *AdminServer) buildOrganizationHome(req OrganizationCreateRequest, template *OrganizationStarterTemplate) OrganizationHomePayload {
	responseContract := defaultResponseContractProfile()
	var templateSummary *OrganizationTemplateSummary
	if template != nil {
		templateSummary = &template.OrganizationTemplateSummary
	}
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
			AIEngineProfileID:         defaultAIEngineProfileID(req.StartMode, templateSummary),
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
		// Departments are copied as planned summaries only; no teams are
		// instantiated from the starter at organization creation (RC-004
		// owner decision 2).
		home.Departments = append([]OrganizationDepartmentSummary(nil), template.Departments...)
		home.DeliveryDefaults = template.DeliveryDefaults
		home.LoopProfiles = append([]bootstrap.TemplateLoopProfile(nil), template.LoopProfiles...)
		if template.Starter != nil {
			home.SomaIntro = template.Starter.SomaIntro
		}
	}

	return normalizeOrganizationHome(home)
}
