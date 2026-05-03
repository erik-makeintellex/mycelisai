package server

import (
	"strings"

	"github.com/mycelis/core/internal/swarm"
)

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
