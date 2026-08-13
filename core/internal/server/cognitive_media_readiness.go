package server

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/mycelis/core/pkg/protocol"
)

func plannedCallsNeedMediaGeneration(calls []protocol.PlannedToolCall) bool {
	for _, call := range calls {
		if strings.EqualFold(strings.TrimSpace(call.Name), "generate_image") {
			return true
		}
	}
	return false
}

func (s *AdminServer) mediaGenerationPreflight(ctx context.Context, calls []protocol.PlannedToolCall) *cognitive.ExecutionAvailability {
	if !plannedCallsNeedMediaGeneration(calls) {
		return nil
	}
	if s.Cognitive == nil || s.Cognitive.Config == nil || s.Cognitive.Config.Media == nil {
		return mediaSetupBlocker("No image generator is configured.", "Connect a local or hosted image generator in Resources, then ask Soma to try the image again.", "")
	}
	provider := s.Cognitive.Config.Media.EffectiveProvider()
	if !provider.IsEnabled() || strings.TrimSpace(provider.Endpoint) == "" {
		return mediaSetupBlocker("Image generation is not enabled for this workspace.", "Enable an image generator in Resources, then ask Soma to try the image again.", provider.ProviderID)
	}
	if !strings.EqualFold(strings.TrimSpace(provider.Type), cognitive.MediaProviderTypeForge) {
		return nil
	}

	endpoint := strings.TrimRight(strings.TrimSpace(provider.Endpoint), "/")
	probeCtx, cancel := context.WithTimeout(ctx, 1500*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(probeCtx, http.MethodGet, endpoint+"/sdapi/v1/options", nil)
	if err != nil {
		return mediaSetupBlocker("The configured Forge address is invalid.", "Correct the Forge address in Resources, then ask Soma to try the image again.", provider.ProviderID)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return mediaSetupBlocker("Soma cannot reach the configured image generator.", fmt.Sprintf("Start Forge at %s, then ask Soma to try the image again.", endpoint), provider.ProviderID)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusNotFound {
		return mediaSetupBlocker("Forge is open, but image generation is not enabled.", "Enable API mode in the Forge/Pinokio launch settings, restart Forge, then ask Soma to try the image again.", provider.ProviderID)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return mediaSetupBlocker("Forge is running but not ready to generate images.", fmt.Sprintf("Check Forge at %s, resolve its startup error, then ask Soma to try the image again.", endpoint), provider.ProviderID)
	}
	return nil
}

func mediaSetupBlocker(summary, action, providerID string) *cognitive.ExecutionAvailability {
	return &cognitive.ExecutionAvailability{
		Available:         false,
		Code:              "media_provider_not_ready",
		Summary:           summary,
		RecommendedAction: action,
		ProviderID:        strings.TrimSpace(providerID),
		SetupRequired:     true,
		SetupPath:         "/resources?section=capabilities",
	}
}
