package server

import (
	"encoding/json"
	"net/http"
	"sort"
	"strings"
	"time"
)

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
