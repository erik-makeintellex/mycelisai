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
	if err := validateWorkerCorrelation(req); err != nil {
		return WorkerRunHandle{}, err
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
		"run_id":             req.RunID,
		"org_id":             req.OrgID,
		"project_id":         req.ProjectID,
		"user_id":            req.UserID,
		"requested_by":       req.RequestedBy,
		"intent":             req.Intent,
		"instructions":       req.Instructions,
		"input":              req.Input,
		"required_protocols": req.RequiredProtocols,
		"required_features":  req.RequiredFeatures,
		"correlation":        req.Correlation,
		"metadata":           req.Metadata,
	}
	var out map[string]any
	if err := b.doJSON(ctx, http.MethodPost, "/v1/runs", payload, &out); err != nil {
		return WorkerRunHandle{}, err
	}
	handle, err := runHandleFromMap(out, BackendFrameworkRuns, protocol)
	if err != nil {
		return WorkerRunHandle{}, err
	}
	backendRunID := handle.RunID
	if strings.TrimSpace(backendRunID) == "" {
		return WorkerRunHandle{}, WorkerBackendError("invalid_backend_response", "External worker backend returned no run_id.", true)
	}
	if backendRunID != req.RunID {
		return WorkerRunHandle{}, WorkerBackendError("run_identity_mismatch", "External worker backend did not preserve the authoritative Mycelis run_id.", false)
	}
	handle.BackendRunID = backendRunID
	return handle, nil
}

func validateWorkerCorrelation(req WorkerRunRequest) error {
	if strings.TrimSpace(req.RunID) == "" || strings.TrimSpace(req.Correlation.RunID) == "" {
		return fmt.Errorf("worker run_id correlation is required")
	}
	if req.RunID != req.Correlation.RunID {
		return fmt.Errorf("worker run_id correlation mismatch")
	}
	if req.RunID != strings.TrimSpace(req.RunID) {
		return fmt.Errorf("worker run_id must be canonical")
	}
	for name, value := range map[string]string{
		"intent_proof_id":       req.Correlation.IntentProofID,
		"execution_contract_id": req.Correlation.ExecutionContractID,
		"work_item_id":          req.Correlation.WorkItemID,
		"idempotency_key":       req.Correlation.IdempotencyKey,
		"source_kind":           req.Correlation.SourceKind,
		"source_channel":        req.Correlation.SourceChannel,
		"payload_kind":          req.Correlation.PayloadKind,
		"graph_revision":        req.Correlation.GraphRevision,
	} {
		if strings.TrimSpace(value) == "" {
			return fmt.Errorf("worker %s correlation is required", name)
		}
	}
	for _, name := range []string{"run_id", "correlation_id", "correlation", "intent_proof_id", "execution_contract_id", "team_id", "outcome_id", "work_item_id", "idempotency_key", "source_kind", "source_channel", "payload_kind", "graph_revision"} {
		if _, duplicated := req.Metadata[name]; duplicated {
			return fmt.Errorf("worker metadata must not duplicate typed correlation field %q", name)
		}
	}
	return nil
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
		streamEventID := ""
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if strings.HasPrefix(line, "id:") {
				streamEventID = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
				continue
			}
			if line == "" || strings.HasPrefix(line, ":") || strings.HasPrefix(line, "event:") {
				continue
			}
			line = strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			var raw map[string]any
			if json.Unmarshal([]byte(line), &raw) != nil {
				emitFrameworkStreamFailure(ctx, events, runID, "External worker returned an invalid event.")
				return
			}
			if streamEventID != "" && stringValue(raw["event_id"]) == "" {
				raw["event_id"] = streamEventID
			}
			streamEventID = ""
			event, err := eventFromMap(raw, runID, BackendFrameworkRuns)
			if err != nil {
				emitFrameworkStreamFailure(ctx, events, runID, err.Error())
				return
			}
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
		EventID: "framework-stream-failure:" + runID, RunID: runID, BackendRunID: runID,
		Backend: BackendFrameworkRuns, Kind: EventFailed, Status: StatusFailed,
		Message: message, Error: WorkerBackendError("invalid_event_stream", message, true), Timestamp: time.Now().UTC(),
	}
	select {
	case <-ctx.Done():
	case events <- event:
	}
}

func (b *FrameworkRunsBackend) GetRun(ctx context.Context, runID string) (WorkerRunHandle, error) {
	canonicalRunID := strings.TrimSpace(runID)
	if canonicalRunID == "" {
		return WorkerRunHandle{}, fmt.Errorf("worker run_id is required")
	}
	var out map[string]any
	if err := b.doJSON(ctx, http.MethodGet, "/v1/runs/"+url.PathEscape(canonicalRunID), nil, &out); err != nil {
		return WorkerRunHandle{}, err
	}
	handle, err := runHandleFromMap(out, BackendFrameworkRuns, ProtocolRunsAPI)
	if err != nil {
		return WorkerRunHandle{}, err
	}
	if handle.RunID != canonicalRunID {
		return WorkerRunHandle{}, WorkerBackendError("run_identity_mismatch", "External worker backend did not preserve the authoritative Mycelis run_id.", false)
	}
	handle.BackendRunID = canonicalRunID
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
	healthy := boolValue(raw["healthy"])
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
	return ""
}

func hasProtocol(protocols []Protocol, want Protocol) bool {
	for _, protocol := range protocols {
		if protocol == want {
			return true
		}
	}
	return false
}
