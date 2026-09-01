package workers

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultConnectTimeout = 5 * time.Second
	defaultRunTimeout     = 15 * time.Second
)

type SecretResolver interface {
	ResolveSecret(ctx context.Context, ref string) (string, error)
}

// FrameworkRunsBackend adapts a durable, HTTP runs API to Mycelis' normalized
// worker lifecycle. It deliberately does not implement RunFinalizer: an
// external completion is evidence for the Mycelis projection path, not direct
// authority to finalize an Outcome.
type FrameworkRunsBackend struct {
	Config  WorkerConfig
	Client  *http.Client
	Secrets SecretResolver
}

func NewFrameworkRunsBackend(cfg WorkerConfig, secrets SecretResolver) (*FrameworkRunsBackend, error) {
	cfg.Backend = canonicalBackendKind(cfg.Backend)
	if cfg.Backend == "" {
		cfg.Backend = BackendFrameworkRuns
	}
	if cfg.Backend != BackendFrameworkRuns {
		return nil, fmt.Errorf("framework_runs backend kind is required")
	}
	if err := validateFrameworkRunsConfig(cfg, secrets); err != nil {
		return nil, err
	}
	cfg.BaseURL = strings.TrimRight(strings.TrimSpace(cfg.BaseURL), "/")
	if cfg.HealthPath == "" {
		cfg.HealthPath = "/health"
	}
	if cfg.CapabilitiesPath == "" {
		cfg.CapabilitiesPath = "/v1/capabilities"
	}
	if cfg.PreferredProtocol == "" {
		cfg.PreferredProtocol = ProtocolRunsAPI
	}
	connectTimeout := durationMS(cfg.TimeoutPolicy.ConnectMS, defaultConnectTimeout)
	runTimeout := durationMS(cfg.TimeoutPolicy.RunMS, defaultRunTimeout)
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.DialContext = (&net.Dialer{Timeout: connectTimeout, KeepAlive: 30 * time.Second}).DialContext
	return &FrameworkRunsBackend{
		Config:  cfg,
		Client:  &http.Client{Transport: transport, Timeout: runTimeout},
		Secrets: secrets,
	}, nil
}

func (b *FrameworkRunsBackend) CreateRun(ctx context.Context, req WorkerRunRequest) (WorkerRunHandle, error) {
	if strings.TrimSpace(req.Intent) == "" {
		return WorkerRunHandle{}, fmt.Errorf("worker run intent is required")
	}
	caps, err := b.GetCapabilities(ctx)
	if err != nil {
		return WorkerRunHandle{}, err
	}
	protocol := selectProtocol(b.Config.PreferredProtocol, caps)
	if protocol != ProtocolRunsAPI {
		return WorkerRunHandle{}, WorkerBackendError("unsupported_protocol", "External worker backend does not expose a durable runs protocol.", true)
	}
	payload := map[string]any{
		"org_id":             req.OrgID,
		"project_id":         req.ProjectID,
		"user_id":            req.UserID,
		"requested_by":       req.RequestedBy,
		"intent":             req.Intent,
		"instructions":       req.Instructions,
		"input":              req.Input,
		"required_protocols": req.RequiredProtocols,
		"required_features":  req.RequiredFeatures,
		"metadata":           req.Metadata,
	}
	var out map[string]any
	if err := b.doJSON(ctx, http.MethodPost, "/v1/runs", payload, &out); err != nil {
		return WorkerRunHandle{}, err
	}
	handle := runHandleFromMap(out, BackendFrameworkRuns, protocol)
	if strings.TrimSpace(handle.RunID) == "" {
		return WorkerRunHandle{}, WorkerBackendError("invalid_backend_response", "External worker backend returned no run_id.", true)
	}
	return handle, nil
}

func (b *FrameworkRunsBackend) StreamRunEvents(ctx context.Context, runID string) (<-chan WorkerEvent, error) {
	if strings.TrimSpace(runID) == "" {
		return nil, fmt.Errorf("worker run_id is required")
	}
	req, err := b.newRequest(ctx, http.MethodGet, "/v1/runs/"+url.PathEscape(runID)+"/events", nil)
	if err != nil {
		return nil, err
	}
	res, err := b.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("framework runs events request failed: %w", err)
	}
	if res.StatusCode >= 300 {
		defer res.Body.Close()
		return nil, statusError("framework runs events", res)
	}
	events := make(chan WorkerEvent)
	go func() {
		defer close(events)
		defer res.Body.Close()
		scanner := bufio.NewScanner(res.Body)
		scanner.Buffer(make([]byte, 64*1024), 1024*1024)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" || strings.HasPrefix(line, ":") || strings.HasPrefix(line, "event:") || strings.HasPrefix(line, "id:") {
				continue
			}
			line = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			var raw map[string]any
			if json.Unmarshal([]byte(line), &raw) != nil {
				emitFrameworkStreamFailure(ctx, events, runID, "External worker returned an invalid event.")
				return
			}
			event := eventFromMap(raw, runID, BackendFrameworkRuns)
			select {
			case <-ctx.Done():
				return
			case events <- event:
			}
		}
		if scanner.Err() != nil && ctx.Err() == nil {
			emitFrameworkStreamFailure(ctx, events, runID, "External worker event stream failed.")
		}
	}()
	return events, nil
}

func emitFrameworkStreamFailure(ctx context.Context, events chan<- WorkerEvent, runID, message string) {
	event := WorkerEvent{
		RunID: runID, Backend: BackendFrameworkRuns, Kind: EventFailed, Status: StatusFailed,
		Message: message, Error: WorkerBackendError("invalid_event_stream", message, true), Timestamp: time.Now().UTC(),
	}
	select {
	case <-ctx.Done():
	case events <- event:
	}
}

func (b *FrameworkRunsBackend) GetRun(ctx context.Context, runID string) (WorkerRunHandle, error) {
	if strings.TrimSpace(runID) == "" {
		return WorkerRunHandle{}, fmt.Errorf("worker run_id is required")
	}
	var out map[string]any
	if err := b.doJSON(ctx, http.MethodGet, "/v1/runs/"+url.PathEscape(runID), nil, &out); err != nil {
		return WorkerRunHandle{}, err
	}
	handle := runHandleFromMap(out, BackendFrameworkRuns, ProtocolRunsAPI)
	if handle.RunID == "" {
		handle.RunID = runID
	}
	return handle, nil
}

func (b *FrameworkRunsBackend) StopRun(ctx context.Context, runID string) error {
	if strings.TrimSpace(runID) == "" {
		return fmt.Errorf("worker run_id is required")
	}
	return b.doJSON(ctx, http.MethodPost, "/v1/runs/"+url.PathEscape(runID)+"/stop", map[string]any{}, nil)
}

func (b *FrameworkRunsBackend) SubmitApproval(ctx context.Context, runID string, approval WorkerApprovalDecision) error {
	if strings.TrimSpace(runID) == "" || strings.TrimSpace(approval.ApprovalID) == "" {
		return fmt.Errorf("worker run_id and approval_id are required")
	}
	if approval.Decision != DecisionApprove && approval.Decision != DecisionDeny {
		return fmt.Errorf("unsupported approval decision %q", approval.Decision)
	}
	return b.doJSON(ctx, http.MethodPost, "/v1/runs/"+url.PathEscape(runID)+"/approvals/"+url.PathEscape(approval.ApprovalID), approval, nil)
}

func (b *FrameworkRunsBackend) GetCapabilities(ctx context.Context) (WorkerCapabilities, error) {
	var raw map[string]any
	if err := b.doJSON(ctx, http.MethodGet, b.Config.CapabilitiesPath, nil, &raw); err != nil {
		return WorkerCapabilities{}, err
	}
	return capabilitiesFromMap(raw, BackendFrameworkRuns), nil
}

func (b *FrameworkRunsBackend) HealthCheck(ctx context.Context) (WorkerHealth, error) {
	var raw map[string]any
	if err := b.doJSON(ctx, http.MethodGet, b.Config.HealthPath, nil, &raw); err != nil {
		return WorkerHealth{}, err
	}
	healthy := truthy(raw["healthy"]) || truthy(raw["ok"]) || truthy(raw["status"])
	return WorkerHealth{Backend: BackendFrameworkRuns, Healthy: healthy, Message: stringValue(raw["message"]), CheckedAt: time.Now().UTC(), Raw: raw}, nil
}

func (b *FrameworkRunsBackend) doJSON(ctx context.Context, method, path string, body any, out any) error {
	req, err := b.newRequest(ctx, method, path, body)
	if err != nil {
		return err
	}
	res, err := b.Client.Do(req)
	if err != nil {
		return fmt.Errorf("framework runs request failed: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode >= 300 {
		return statusError("framework runs request", res)
	}
	if out == nil {
		_, _ = io.Copy(io.Discard, res.Body)
		return nil
	}
	if err := json.NewDecoder(res.Body).Decode(out); err != nil {
		return WorkerBackendError("invalid_backend_response", "External worker backend returned invalid JSON.", true)
	}
	return nil
}

func (b *FrameworkRunsBackend) newRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	base, err := url.Parse(b.Config.BaseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid framework_runs base_url: %w", err)
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
	if b.Config.APIKeySecretRef != "" {
		token, err := b.Secrets.ResolveSecret(ctx, b.Config.APIKeySecretRef)
		if err != nil {
			return nil, fmt.Errorf("resolve framework worker API credential: %w", err)
		}
		if strings.TrimSpace(token) == "" {
			return nil, fmt.Errorf("resolved framework worker API credential is empty")
		}
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req, nil
}

func durationMS(value int, fallback time.Duration) time.Duration {
	if value <= 0 {
		return fallback
	}
	return time.Duration(value) * time.Millisecond
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
