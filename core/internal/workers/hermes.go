package workers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type SecretResolver interface {
	ResolveSecret(ctx context.Context, ref string) (string, error)
}

type HermesAPIBackend struct {
	Config  WorkerConfig
	Client  *http.Client
	Secrets SecretResolver
}

func NewHermesAPIBackend(cfg WorkerConfig, secrets SecretResolver) (*HermesAPIBackend, error) {
	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("hermes_api base_url is required")
	}
	if cfg.HealthPath == "" {
		cfg.HealthPath = "/health"
	}
	if cfg.CapabilitiesPath == "" {
		cfg.CapabilitiesPath = "/v1/capabilities"
	}
	if cfg.PreferredProtocol == "" {
		cfg.PreferredProtocol = ProtocolRunsAPI
	}
	return &HermesAPIBackend{
		Config:  cfg,
		Client:  &http.Client{Timeout: 15 * time.Second},
		Secrets: secrets,
	}, nil
}

func (b *HermesAPIBackend) CreateRun(ctx context.Context, req WorkerRunRequest) (WorkerRunHandle, error) {
	caps, err := b.GetCapabilities(ctx)
	if err != nil {
		return WorkerRunHandle{}, err
	}
	protocol := selectProtocol(b.Config.PreferredProtocol, caps)
	if protocol != ProtocolRunsAPI {
		return WorkerRunHandle{}, WorkerBackendError("unsupported_protocol", "Hermes backend does not expose a durable runs protocol.", true)
	}
	payload := map[string]any{
		"intent":       req.Intent,
		"instructions": req.Instructions,
		"input":        req.Input,
		"metadata":     req.Metadata,
	}
	var out map[string]any
	if err := b.doJSON(ctx, http.MethodPost, "/v1/runs", payload, &out); err != nil {
		return WorkerRunHandle{}, err
	}
	return runHandleFromMap(out, BackendHermesAPI, protocol), nil
}

func (b *HermesAPIBackend) StreamRunEvents(ctx context.Context, runID string) (<-chan WorkerEvent, error) {
	req, err := b.newRequest(ctx, http.MethodGet, "/v1/runs/"+url.PathEscape(runID)+"/events", nil)
	if err != nil {
		return nil, err
	}
	res, err := b.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("hermes events request failed: %w", err)
	}
	if res.StatusCode >= 300 {
		defer res.Body.Close()
		return nil, statusError("hermes events", res)
	}
	events := make(chan WorkerEvent)
	go func() {
		defer close(events)
		defer res.Body.Close()
		scanner := bufio.NewScanner(res.Body)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, ":") {
				continue
			}
			line = strings.TrimPrefix(line, "data:")
			var raw map[string]any
			if json.Unmarshal([]byte(strings.TrimSpace(line)), &raw) == nil {
				event := eventFromMap(raw, runID, BackendHermesAPI)
				select {
				case <-ctx.Done():
					return
				case events <- event:
				}
			}
		}
	}()
	return events, nil
}

func (b *HermesAPIBackend) GetRun(ctx context.Context, runID string) (WorkerRunHandle, error) {
	var out map[string]any
	if err := b.doJSON(ctx, http.MethodGet, "/v1/runs/"+url.PathEscape(runID), nil, &out); err != nil {
		return WorkerRunHandle{}, err
	}
	return runHandleFromMap(out, BackendHermesAPI, ProtocolRunsAPI), nil
}

func (b *HermesAPIBackend) StopRun(ctx context.Context, runID string) error {
	return b.doJSON(ctx, http.MethodPost, "/v1/runs/"+url.PathEscape(runID)+"/stop", map[string]any{}, nil)
}

func (b *HermesAPIBackend) SubmitApproval(ctx context.Context, runID string, approval WorkerApprovalDecision) error {
	return b.doJSON(ctx, http.MethodPost, "/v1/runs/"+url.PathEscape(runID)+"/approvals/"+url.PathEscape(approval.ApprovalID), approval, nil)
}

func (b *HermesAPIBackend) GetCapabilities(ctx context.Context) (WorkerCapabilities, error) {
	var raw map[string]any
	if err := b.doJSON(ctx, http.MethodGet, b.Config.CapabilitiesPath, nil, &raw); err != nil {
		return WorkerCapabilities{}, err
	}
	return capabilitiesFromMap(raw, BackendHermesAPI), nil
}

func (b *HermesAPIBackend) HealthCheck(ctx context.Context) (WorkerHealth, error) {
	var raw map[string]any
	if err := b.doJSON(ctx, http.MethodGet, b.Config.HealthPath, nil, &raw); err != nil {
		return WorkerHealth{}, err
	}
	return WorkerHealth{Backend: BackendHermesAPI, Healthy: truthy(raw["healthy"]) || truthy(raw["ok"]), Message: stringValue(raw["message"]), CheckedAt: time.Now().UTC(), Raw: raw}, nil
}

func (b *HermesAPIBackend) doJSON(ctx context.Context, method, path string, body any, out any) error {
	req, err := b.newRequest(ctx, method, path, body)
	if err != nil {
		return err
	}
	res, err := b.Client.Do(req)
	if err != nil {
		return fmt.Errorf("hermes request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return statusError("hermes request", res)
	}
	if out == nil {
		io.Copy(io.Discard, res.Body)
		return nil
	}
	return json.NewDecoder(res.Body).Decode(out)
}

func (b *HermesAPIBackend) newRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	base, err := url.Parse(b.Config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid hermes base_url: %w", err)
	}
	ref, err := url.Parse(path)
	if err != nil {
		return nil, err
	}
	var reader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reader = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, base.ResolveReference(ref).String(), reader)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if b.Config.APIKeySecretRef != "" && b.Secrets != nil {
		token, err := b.Secrets.ResolveSecret(ctx, b.Config.APIKeySecretRef)
		if err != nil {
			return nil, fmt.Errorf("resolve hermes api key secret ref: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req, nil
}

func selectProtocol(preferred Protocol, caps WorkerCapabilities) Protocol {
	if hasProtocol(caps.SupportedProtocols, preferred) {
		return preferred
	}
	for _, protocol := range []Protocol{ProtocolRunsAPI, ProtocolResponsesAPI, ProtocolChatCompletion} {
		if hasProtocol(caps.SupportedProtocols, protocol) {
			return protocol
		}
	}
	return ProtocolUnknown
}

func hasProtocol(protocols []Protocol, want Protocol) bool {
	for _, protocol := range protocols {
		if protocol == want {
			return true
		}
	}
	return false
}
