package server

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode"

	"github.com/google/uuid"
	"github.com/mycelis/core/internal/bootstrap"
)

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
