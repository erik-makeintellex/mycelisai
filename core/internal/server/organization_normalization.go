package server

import (
	"fmt"
	"strings"
)

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
