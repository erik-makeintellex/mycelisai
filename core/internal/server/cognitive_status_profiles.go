package server

import (
	"context"

	"github.com/mycelis/core/internal/cognitive"
)

// cognitiveTextStatus is the text engine view of GET /api/v1/cognitive/status.
type cognitiveTextStatus struct {
	Status            string `json:"status"`
	Endpoint          string `json:"endpoint,omitempty"`
	Model             string `json:"model,omitempty"`
	ProviderID        string `json:"provider_id,omitempty"`
	Detail            string `json:"detail,omitempty"`
	RecommendedAction string `json:"recommended_action,omitempty"`
	SetupRequired     bool   `json:"setup_required,omitempty"`
}

// cognitiveProfileRouteStatus resolves and probes every profile route and
// derives text.status from the chat profile's effective provider only, so a
// misrouted chat profile is never reported online because some other
// provider answered its probe.
func (s *AdminServer) cognitiveProfileRouteStatus(ctx context.Context) (cognitiveTextStatus, map[string]cognitive.ProfileRoute, string, bool) {
	routes, overlayError := s.Cognitive.ProfileRoutes()
	cognitive.ProbeProfileRoutes(ctx, routes, cognitive.ProfileRouteProbeTimeout)
	health := cognitive.ProfileRouteHealth(routes, overlayError)

	chat := routes[cognitive.DefaultExecutionProfileName]
	text := cognitiveTextStatus{Status: "offline", ProviderID: chat.ProviderID, Model: chat.ModelID}
	switch {
	case !chat.Available:
		availability := s.Cognitive.ExecutionAvailability(cognitive.DefaultExecutionProfileName, "")
		text.Detail = availability.Summary
		text.RecommendedAction = availability.RecommendedAction
		text.SetupRequired = availability.SetupRequired
	case chat.Reachable == nil:
		text.Status = "configured"
		text.Detail = "The chat profile routes to a provider that is not probed (hosted or gateway); live health is checked during inference."
	case *chat.Reachable:
		text.Status = "online"
		if provider, ok := s.Cognitive.ProviderSnapshot(chat.ProviderID); ok {
			text.Endpoint = provider.Endpoint
		}
	default:
		text.Detail = "The chat profile's provider " + chat.ProviderID + " did not answer its health probe."
		text.RecommendedAction = "Start or repair provider " + chat.ProviderID + ", or reset the chat profile to root."
		text.SetupRequired = true
	}
	return text, routes, health, overlayError
}

// Operational summary for callers without the full cognitive view (S6e):
// see cognitiveFullView. Endpoints, model ids, provider ids, config
// snapshots, db_row_present, and override-origin detail are omitted (the
// keys are absent, not null).
type cognitiveTextSummary struct {
	Status            string `json:"status"`
	Detail            string `json:"detail,omitempty"`
	RecommendedAction string `json:"recommended_action,omitempty"`
	SetupRequired     bool   `json:"setup_required,omitempty"`
}

type cognitiveProfileSummary struct {
	Available bool   `json:"available"`
	Code      string `json:"code"`
	Reachable *bool  `json:"reachable"`
}

// narrowCognitiveStatus builds the non-admin GET /api/v1/cognitive/status body.
func narrowCognitiveStatus(text cognitiveTextStatus, mediaStatus string, routes map[string]cognitive.ProfileRoute, health string, overlayError bool) map[string]any {
	profiles := make(map[string]cognitiveProfileSummary, len(routes))
	for name, route := range routes {
		profiles[name] = cognitiveProfileSummary{Available: route.Available, Code: route.Code, Reachable: route.Reachable}
	}
	return map[string]any{
		"text": cognitiveTextSummary{
			Status: text.Status, Detail: text.Detail,
			RecommendedAction: text.RecommendedAction, SetupRequired: text.SetupRequired,
		},
		"media":                map[string]string{"status": mediaStatus},
		"profiles":             profiles,
		"profile_route_health": health,
		"overlay_error":        overlayError,
	}
}
