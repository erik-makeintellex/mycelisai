package server

import (
	"strings"

	"github.com/mycelis/core/pkg/protocol"
)

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
