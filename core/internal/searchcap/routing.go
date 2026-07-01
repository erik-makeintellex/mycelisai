package searchcap

import (
	"context"
	"errors"
	"net/http"
	"strings"
)

func (s *Service) routeSelectedSource(ctx context.Context, req Request, resp Response) (Response, bool, error) {
	sourceID := strings.TrimSpace(req.SourceID)
	if sourceID == "" {
		return resp, false, nil
	}
	source, ok := s.findSource(sourceID)
	if !ok {
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "search_source_not_found", Message: "The requested search source is not configured.", NextAction: "Open Resources > Capabilities and choose an available search source."}
		return resp, true, nil
	}
	resp.Metadata["selected_source_id"] = source.ID
	resp.Metadata["selected_source_name"] = source.Name
	resp.Metadata["selected_source_boundary"] = source.Boundary
	resp.Metadata["selected_source_type"] = source.SourceType
	if blocker := sourceSelectionBlocker(source, req); blocker != nil {
		resp.Status = "blocked"
		resp.Blocker = blocker
		return resp, true, nil
	}
	switch sourceProvider(source) {
	case ProviderLocalSources:
		req.SourceScope = "local_sources"
		routed, err := s.searchLocalSources(ctx, req, resp)
		return routed, true, err
	case ProviderLocalAPI:
		if source.Endpoint == "" {
			resp.Status = "blocked"
			resp.Blocker = &Blocker{Code: "search_source_endpoint_missing", Message: "The selected search source has no endpoint configured.", NextAction: "Update the search source endpoint in Resources > Capabilities."}
			return resp, true, nil
		}
		routed, err := s.searchLocalAPIEndpoint(ctx, req, resp, source.Endpoint)
		return routed, true, err
	case ProviderSearXNG:
		if source.Endpoint == "" {
			resp.Status = "blocked"
			resp.Blocker = &Blocker{Code: "search_source_endpoint_missing", Message: "The selected search source has no endpoint configured.", NextAction: "Update the search source endpoint in Resources > Capabilities."}
			return resp, true, nil
		}
		routed, err := s.searchSearXNGEndpoint(ctx, req, resp, source.Endpoint)
		return routed, true, err
	default:
		resp.Status = "blocked"
		resp.Blocker = &Blocker{Code: "search_source_adapter_unavailable", Message: "The selected search source is registered, but this runtime does not yet have a safe adapter for it.", NextAction: "Use a local-source, SearXNG, or local API source, or add an adapter before routing Soma through this source."}
		return resp, true, nil
	}
}

func (s *Service) findSource(id string) (Source, bool) {
	for _, source := range s.ListSources() {
		if source.ID == id {
			return source, true
		}
	}
	return Source{}, false
}

func sourceSelectionBlocker(source Source, req Request) *Blocker {
	status := normalizeSourceToken(source.Status)
	if status != "" && status != "available" && status != "ready" && status != "online" {
		return &Blocker{Code: "search_source_not_ready", Message: "The selected search source is not ready.", NextAction: firstString(source.Recovery, "Repair or choose another search source in Resources > Capabilities.")}
	}
	if source.AuthScheme != "none" && source.AuthScheme != "service_managed" {
		return &Blocker{Code: "search_source_auth_adapter_required", Message: "The selected source uses saved authentication, but search routing cannot safely apply that credential yet.", NextAction: "Use a service-managed source or add a governed auth adapter for this source."}
	}
	switch source.ScopeKind {
	case "", "all":
		return nil
	case "group":
		if strings.TrimSpace(req.TeamID) == source.ScopeRef {
			return nil
		}
		return &Blocker{Code: "search_source_out_of_scope", Message: "The selected search source is scoped to a different group.", NextAction: "Switch to the matching group or choose a source visible to this work."}
	case "host":
		if strings.TrimSpace(req.HostID) == source.ScopeRef {
			return nil
		}
		return &Blocker{Code: "search_source_out_of_scope", Message: "The selected search source is scoped to a different host.", NextAction: "Use the configured host or choose a source visible to everyone."}
	default:
		return &Blocker{Code: "search_source_scope_unknown", Message: "The selected search source has an unsupported scope.", NextAction: "Update the source scope in Resources > Capabilities."}
	}
}

func sourceProvider(source Source) string {
	provider := normalizeSourceToken(source.Provider)
	sourceType := normalizeSourceToken(source.SourceType)
	switch {
	case provider == ProviderLocalSources || sourceType == ProviderLocalSources || sourceType == "knowledge_collection":
		return ProviderLocalSources
	case provider == ProviderLocalAPI || sourceType == ProviderLocalAPI || sourceType == "private_api" || sourceType == "authenticated_api" || sourceType == "client_or_public_api":
		return ProviderLocalAPI
	case provider == ProviderSearXNG || source.ID == "searxng":
		return ProviderSearXNG
	default:
		return provider
	}
}

func SourceErrorStatus(err error) int {
	if errors.Is(err, errSourceNotFound) {
		return http.StatusNotFound
	}
	return http.StatusBadRequest
}
