package workers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

var (
	ErrExternalExecutionUnavailable = errors.New("external worker execution is unavailable until correlated event projection is production-certified")
	envSecretNamePattern            = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)
)

// LoadEngineConfig reads only the worker execution section from engine.yaml.
// Other engine sections remain owned by their respective runtime packages.
func LoadEngineConfig(path string) (WorkerConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return WorkerConfig{}, err
	}
	return ParseEngineConfig(data)
}

func ParseEngineConfig(data []byte) (WorkerConfig, error) {
	var root yaml.Node
	if err := yaml.Unmarshal(data, &root); err != nil {
		return WorkerConfig{}, fmt.Errorf("parse engine worker configuration: %w", err)
	}
	cfg := WorkerConfig{Backend: BackendCentral}
	if len(root.Content) == 0 || len(root.Content[0].Content) == 0 {
		return cfg, nil
	}
	mapping := root.Content[0]
	if mapping.Kind != yaml.MappingNode {
		return WorkerConfig{}, fmt.Errorf("engine configuration must be a mapping")
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value != "worker_runtime" {
			continue
		}
		encoded, err := yaml.Marshal(mapping.Content[i+1])
		if err != nil {
			return WorkerConfig{}, fmt.Errorf("encode engine worker_runtime configuration: %w", err)
		}
		decoder := yaml.NewDecoder(bytes.NewReader(encoded))
		decoder.KnownFields(true)
		if err := decoder.Decode(&cfg); err != nil {
			return WorkerConfig{}, fmt.Errorf("parse engine worker_runtime configuration: %w", err)
		}
		break
	}
	if cfg.Backend == "" {
		cfg.Backend = BackendCentral
	}
	return cfg, nil
}

// NewBackend constructs a normalized backend client. It does not authorize
// that backend to execute Mycelis Outcomes.
func NewBackend(cfg WorkerConfig, secrets SecretResolver) (WorkerBackend, error) {
	switch canonicalBackendKind(cfg.Backend) {
	case "", BackendCentral:
		return NewCentralBackend(), nil
	case BackendFrameworkRuns:
		return NewFrameworkRunsBackend(cfg, secrets)
	default:
		return nil, fmt.Errorf("unsupported worker backend %q", cfg.Backend)
	}
}

// NewExecutionBackend returns only production-certified execution backends.
// Framework Runs remains fail-closed until its persistence, auth, isolation,
// restart, and live projection gates are all proven.
func NewExecutionBackend(cfg WorkerConfig, secrets SecretResolver) (WorkerBackend, error) {
	backend, err := NewBackend(cfg, secrets)
	if err != nil {
		return nil, err
	}
	if canonicalBackendKind(cfg.Backend) == BackendFrameworkRuns {
		return nil, ErrExternalExecutionUnavailable
	}
	return backend, nil
}

func canonicalBackendKind(kind BackendKind) BackendKind {
	return BackendKind(strings.ToLower(strings.TrimSpace(string(kind))))
}

func validateFrameworkRunsConfig(cfg WorkerConfig, secrets SecretResolver) error {
	baseURL, err := url.Parse(strings.TrimSpace(cfg.BaseURL))
	if err != nil || baseURL.Host == "" || (baseURL.Scheme != "http" && baseURL.Scheme != "https") {
		return fmt.Errorf("framework_runs base_url must be an absolute HTTP(S) URL")
	}
	if baseURL.User != nil || baseURL.RawQuery != "" || baseURL.Fragment != "" {
		return fmt.Errorf("framework_runs base_url must not contain credentials, query parameters, or fragments")
	}
	if cfg.APIKeySecretRef != "" {
		if !isManagedSecretRef(cfg.APIKeySecretRef) {
			return fmt.Errorf("framework_runs api_key_secret_ref must be an env: or secret:// reference")
		}
		if secrets == nil {
			return fmt.Errorf("framework_runs api_key_secret_ref requires a secret resolver")
		}
	}
	for name, path := range map[string]string{"capabilities_endpoint": cfg.CapabilitiesPath, "health_endpoint": cfg.HealthPath} {
		if path != "" && (!strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//")) {
			return fmt.Errorf("framework_runs %s must be an absolute API path", name)
		}
	}
	if cfg.PreferredProtocol != "" && cfg.PreferredProtocol != ProtocolRunsAPI {
		return fmt.Errorf("framework_runs preferred_protocol must be runs_api")
	}
	if cfg.EventStreamMode != "" && cfg.EventStreamMode != "sse" {
		return fmt.Errorf("framework_runs event_stream_mode must be sse")
	}
	if cfg.ApprovalMode != "" && cfg.ApprovalMode != "mycelis_control_plane" {
		return fmt.Errorf("framework_runs approval_mode must be mycelis_control_plane")
	}
	if cfg.TimeoutPolicy.ConnectMS < 0 || cfg.TimeoutPolicy.RunMS < 0 || cfg.TimeoutPolicy.StreamMS < 0 {
		return fmt.Errorf("framework_runs timeout values cannot be negative")
	}
	return nil
}

func isManagedSecretRef(ref string) bool {
	ref = strings.TrimSpace(ref)
	if strings.HasPrefix(ref, "secret://") {
		return len(strings.TrimPrefix(ref, "secret://")) > 0
	}
	if strings.HasPrefix(ref, "env:") {
		return envSecretNamePattern.MatchString(strings.TrimPrefix(ref, "env:"))
	}
	return false
}

// EnvSecretResolver resolves deployment-owned environment references without
// accepting raw credentials from configuration.
type EnvSecretResolver struct{}

func (EnvSecretResolver) ResolveSecret(_ context.Context, ref string) (string, error) {
	if !strings.HasPrefix(ref, "env:") {
		return "", fmt.Errorf("secret reference provider is unavailable")
	}
	name := strings.TrimPrefix(ref, "env:")
	if !envSecretNamePattern.MatchString(name) {
		return "", fmt.Errorf("invalid environment secret reference")
	}
	value, ok := os.LookupEnv(name)
	if !ok || strings.TrimSpace(value) == "" {
		return "", fmt.Errorf("referenced environment secret is unavailable")
	}
	return value, nil
}
