package cognitive

import (
	"context"
	"strings"
	"time"
)

type Tier string

const (
	TierS Tier = "S" // Supreme (Reasoning)
	TierA Tier = "A" // Advanced (Code/Logic)
	TierB Tier = "B" // Basic (Fast Local)
	TierC Tier = "C" // Edge (IoT)
	TierU Tier = "U" // Unknown
)

type ValidationResult struct {
	ProviderID string
	Healthy    bool
	ModelID    string
	Tier       Tier
	Error      error
}

// ServiceDiscovery manages probing and grading of providers
type ServiceDiscovery struct {
	Providers map[string]LLMProvider
}

func NewServiceDiscovery(providers map[string]LLMProvider) *ServiceDiscovery {
	return &ServiceDiscovery{Providers: providers}
}

// DiscoverAll probes all providers and returns a health map
func (s *ServiceDiscovery) DiscoverAll(ctx context.Context) map[string]ValidationResult {
	results := make(map[string]ValidationResult)

	for id, provider := range s.Providers {
		// 10s timeout for probing (Kind->Host LAN can be slow)
		ctxProbe, cancel := context.WithTimeout(ctx, 10*time.Second)
		healthy, err := provider.Probe(ctxProbe)
		cancel()

		// Get Model ID? The provider struct has it, but interface doesn't expose it.
		// For now, we assume Config provided ModelID is the one we use.
		// In future: Probe() could return metadata.

		tier := TierU
		// Simple Mock Grading based on Provider ID/Config if readable?
		// Limitation: Provider interface doesn't expose Config.
		// Ideally, we pass the Config map into Discovery.

		results[id] = ValidationResult{
			ProviderID: id,
			Healthy:    healthy,
			Error:      err,
			Tier:       tier, // Filled by Grader later
		}
	}
	return results
}

// GradeModel assigns a Tier based on the Model ID string
func GradeModel(modelID string) Tier {
	m := strings.ToLower(modelID)

	// Tier S: Reasoning / Top Commercial
	if strings.Contains(m, "gpt-4") ||
		strings.Contains(m, "claude-3-opus") ||
		strings.Contains(m, "gemini-1.5-pro") ||
		strings.Contains(m, "o1") {
		return TierS
	}

	// Tier A: Smart Local / Mid Commercial
	if strings.Contains(m, "claude-3-5-sonnet") ||
		strings.Contains(m, "qwen2.5-32b") ||
		strings.Contains(m, "llama-3-70b") {
		return TierA
	}

	// Tier B: Standard Local
	if strings.Contains(m, "qwen") ||
		strings.Contains(m, "llama") ||
		strings.Contains(m, "mistral") ||
		strings.Contains(m, "gemma") {
		// Default bucket for 7B-14B models
		return TierB
	}

	// Tier C: Light
	if strings.Contains(m, "phi") ||
		strings.Contains(m, "tiny") {
		return TierC
	}

	return TierB // Default to B if unknown but valid
}
