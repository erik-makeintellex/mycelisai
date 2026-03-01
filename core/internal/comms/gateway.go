package comms

import (
	"context"
	"fmt"
	"sort"
	"strings"
)

// SendRequest is a normalized outbound communication payload.
type SendRequest struct {
	Provider  string         `json:"provider"`
	Recipient string         `json:"recipient,omitempty"`
	Message   string         `json:"message"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// SendResult is the provider response normalized for API/tool callers.
type SendResult struct {
	Provider          string         `json:"provider"`
	ProviderMessageID string         `json:"provider_message_id,omitempty"`
	Status            string         `json:"status"`
	Metadata          map[string]any `json:"metadata,omitempty"`
}

// ProviderInfo describes one provider and whether it is currently configured.
type ProviderInfo struct {
	Name        string `json:"name"`
	Channel     string `json:"channel"`
	Description string `json:"description"`
	Configured  bool   `json:"configured"`
}

// Provider is implemented by concrete communication channel adapters.
type Provider interface {
	Info() ProviderInfo
	Send(ctx context.Context, req SendRequest) (SendResult, error)
}

// Gateway routes outbound communication requests to configured providers.
type Gateway struct {
	providers map[string]Provider
}

func NewGateway() *Gateway {
	return &Gateway{providers: map[string]Provider{}}
}

func (g *Gateway) Register(p Provider) {
	if g == nil || p == nil {
		return
	}
	info := p.Info()
	name := strings.ToLower(strings.TrimSpace(info.Name))
	if name == "" {
		return
	}
	g.providers[name] = p
}

func (g *Gateway) ListProviders() []ProviderInfo {
	if g == nil {
		return []ProviderInfo{}
	}
	out := make([]ProviderInfo, 0, len(g.providers))
	for _, p := range g.providers {
		out = append(out, p.Info())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func (g *Gateway) Send(ctx context.Context, req SendRequest) (SendResult, error) {
	if g == nil {
		return SendResult{}, fmt.Errorf("communications gateway unavailable")
	}
	provider := strings.ToLower(strings.TrimSpace(req.Provider))
	if provider == "" {
		return SendResult{}, fmt.Errorf("provider is required")
	}
	if strings.TrimSpace(req.Message) == "" {
		return SendResult{}, fmt.Errorf("message is required")
	}
	if req.Metadata == nil {
		req.Metadata = map[string]any{}
	}

	p, ok := g.providers[provider]
	if !ok {
		return SendResult{}, fmt.Errorf("provider %q not registered", provider)
	}
	req.Provider = provider
	return p.Send(ctx, req)
}
