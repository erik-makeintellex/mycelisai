package workers

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
)

type countedFrameworkSecret struct {
	calls atomic.Int32
	refs  []string
}

func (resolver *countedFrameworkSecret) ResolveSecret(_ context.Context, ref string) (string, error) {
	resolver.calls.Add(1)
	resolver.refs = append(resolver.refs, ref)
	return "fixture-scoped-token", nil
}

func TestFrameworkRunsProductionOriginAndFixtureHTTPBoundary(t *testing.T) {
	for _, candidate := range []string{
		"http://127.0.0.1:8091", "http://localhost:8091", "http://workers.example.test",
		"https://user@workers.example.test", "https://workers.example.test/prefix",
		"https://workers.example.test/?mode=run", "https://workers.example.test/#section",
		"https://0.0.0.0:8091", "https://[::]:8091", "https://workers.example.test:",
		"https://workers.example.test:0", "https://workers.example.test:65536", "https://workers.example.test:08091",
	} {
		_, err := NewFrameworkRunsBackend(WorkerConfig{
			BaseURL: candidate, APIKeySecretRef: "env:MYCELIS_WORKER_API_KEY",
		}, &countedFrameworkSecret{})
		if err == nil {
			t.Errorf("production accepted unsafe origin %q", candidate)
		}
	}
	for _, candidate := range []string{
		"http://localhost:8091", "http://0.0.0.0:8091", "http://[::]:8091", "http://workers.example.test",
	} {
		_, err := newFrameworkRunsBackend(WorkerConfig{
			BaseURL: candidate, APIKeySecretRef: "env:MYCELIS_WORKER_API_KEY",
		}, &countedFrameworkSecret{}, true)
		if err == nil {
			t.Errorf("fixture accepted nonliteral-loopback HTTP origin %q", candidate)
		}
	}
	for _, candidate := range []string{"http://127.0.0.1:8091", "http://[::1]:8091"} {
		if _, err := newFrameworkRunsBackend(WorkerConfig{
			BaseURL: candidate, APIKeySecretRef: "env:MYCELIS_WORKER_API_KEY",
		}, &countedFrameworkSecret{}, true); err != nil {
			t.Errorf("fixture rejected literal loopback %q: %v", candidate, err)
		}
	}
}

func TestFrameworkRunsPublicIDsCannotSelectAnotherRoute(t *testing.T) {
	resolver := &countedFrameworkSecret{}
	backend, err := NewFrameworkRunsBackend(WorkerConfig{
		BaseURL: "https://workers.example.test:8091", APIKeySecretRef: "env:MYCELIS_WORKER_API_KEY",
	}, resolver)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := backend.GetRun(t.Context(), "run-1/events"); err == nil {
		t.Fatal("GetRun accepted an SSE route as a run identifier")
	}
	if _, err := backend.StreamRunEventsAfter(t.Context(), "run-1/stop", 0); err == nil {
		t.Fatal("StreamRunEventsAfter accepted a control route as a run identifier")
	}
	if _, err := backend.CreateRun(t.Context(), correlatedTestRunRequest("run-1/events", "build")); err == nil {
		t.Fatal("CreateRun accepted a route-bearing run identifier")
	}
	if _, err := backend.StopRunCommand(t.Context(), "run-1/events", WorkerStopCommand{
		CommandID: "stop-1", ActorID: "actor-1", ExpectedVersion: 1,
	}); err == nil {
		t.Fatal("StopRunCommand accepted a route-bearing run identifier")
	}
	if _, err := backend.SubmitApprovalCommand(t.Context(), "run-1", WorkerApprovalDecision{
		ApprovalID: "approve/stop", Decision: DecisionApprove,
		CommandID: "approve-1", ActorID: "actor-1", ExpectedVersion: 1,
	}); err == nil {
		t.Fatal("SubmitApprovalCommand accepted a route-bearing approval identifier")
	}
	if resolver.calls.Load() != 0 {
		t.Fatalf("invalid public identifier resolved a credential %d times", resolver.calls.Load())
	}
}

func TestFrameworkRunsOriginAndPathsFailBeforeCredentialResolution(t *testing.T) {
	resolver := &countedFrameworkSecret{}
	for _, cfg := range []WorkerConfig{
		{BaseURL: "https://workers.example.test"},
		{BaseURL: "https://workers.example.test", APIKeySecretRef: "raw-secret"},
		{BaseURL: "https://workers.example.test", APIKeySecretRef: "env:MYCELIS_WORKER_API_KEY", CapabilitiesPath: "//evil.test/steal"},
		{BaseURL: "https://workers.example.test", APIKeySecretRef: "env:MYCELIS_WORKER_API_KEY", HealthPath: "/health?redirect=1"},
	} {
		if _, err := NewFrameworkRunsBackend(cfg, resolver); err == nil {
			t.Errorf("accepted unsafe config %#v", cfg)
		}
	}
	backend, err := NewFrameworkRunsBackend(WorkerConfig{
		BaseURL: "https://workers.example.test:8091/", APIKeySecretRef: "env:MYCELIS_WORKER_API_KEY",
	}, resolver)
	if err != nil {
		t.Fatal(err)
	}
	if resolver.calls.Load() != 0 {
		t.Fatal("constructor resolved a secret")
	}
	for _, path := range []string{
		"https://evil.test/v1/runs", "//evil.test/v1/runs", "/v1/runs/../health",
		"/v1/runs/%2e%2e/health", "/v1/runs/run-1?redirect=1", "/v1/runs/run-1#fragment",
		"/v1//runs", "/v1/runs/run\\other", "/other",
	} {
		if _, err := backend.newRequest(t.Context(), http.MethodGet, path, nil); err == nil {
			t.Errorf("accepted unsafe request path %q", path)
		}
	}
	if resolver.calls.Load() != 0 {
		t.Fatal("invalid path resolved a secret")
	}
	req, err := backend.newRequest(t.Context(), http.MethodGet, "/v1/runs/run:one", nil)
	if err != nil || req.URL.String() != "https://workers.example.test:8091/v1/runs/run:one" {
		t.Fatalf("canonical request = %v, %v", req, err)
	}
	if resolver.calls.Load() != 1 {
		t.Fatal("valid request did not resolve one scoped secret")
	}
}

func TestFrameworkRunsConfigSnapshotCannotRetarget(t *testing.T) {
	resolver := &countedFrameworkSecret{}
	cfg := WorkerConfig{
		BaseURL: "https://workers.example.test:8091", APIKeySecretRef: "env:ORIGINAL_WORKER_TOKEN",
		HealthPath: "/health", CapabilitiesPath: "/v1/capabilities",
	}
	backend, err := NewFrameworkRunsBackend(cfg, resolver)
	if err != nil {
		t.Fatal(err)
	}
	cfg.BaseURL = "https://other.example.test"
	cfg.APIKeySecretRef = "env:OTHER_WORKER_TOKEN"
	cfg.HealthPath = "//other.example.test/health"
	req, err := backend.newRequest(t.Context(), http.MethodGet, backend.config.HealthPath, nil)
	if err != nil || req.URL.Host != "workers.example.test:8091" ||
		req.Header.Get("Authorization") != "Bearer fixture-scoped-token" ||
		len(resolver.refs) != 1 || resolver.refs[0] != "env:ORIGINAL_WORKER_TOKEN" {
		t.Fatalf("mutable caller config retargeted request: req=%v err=%v refs=%v", req, err, resolver.refs)
	}
	transport := backend.client.Transport.(*http.Transport)
	if transport.Proxy != nil || transport.DialTLSContext != nil || transport.TLSClientConfig.InsecureSkipVerify ||
		transport.TLSClientConfig.MinVersion < tls.VersionTLS12 || backend.client.CheckRedirect == nil {
		t.Fatal("framework transport inherited unsafe proxy/TLS/redirect behavior")
	}
}

func newTrustedFrameworkTestClient(t *testing.T, server *httptest.Server) *FrameworkRunsBackend {
	t.Helper()
	backend, err := NewFrameworkRunsBackend(WorkerConfig{
		BaseURL: server.URL, APIKeySecretRef: "env:MYCELIS_WORKER_API_KEY",
	}, &countedFrameworkSecret{})
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(server.Certificate())
	backend.client.Transport.(*http.Transport).TLSClientConfig.RootCAs = roots
	return backend
}

func TestFrameworkRunsTLSRequiresTrustedCAAndHostname(t *testing.T) {
	var requests atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		requests.Add(1)
		_ = json.NewEncoder(w).Encode(map[string]any{"healthy": true})
	}))
	defer server.Close()
	backend, err := NewFrameworkRunsBackend(WorkerConfig{
		BaseURL: server.URL, APIKeySecretRef: "env:MYCELIS_WORKER_API_KEY",
	}, &countedFrameworkSecret{})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := backend.HealthCheck(t.Context()); err == nil || requests.Load() != 0 {
		t.Fatalf("untrusted CA reached handler: error=%v requests=%d", err, requests.Load())
	}
	trusted := newTrustedFrameworkTestClient(t, server)
	if health, err := trusted.HealthCheck(t.Context()); err != nil || !health.Healthy || requests.Load() != 1 {
		t.Fatalf("trusted TLS failed: health=%#v error=%v requests=%d", health, err, requests.Load())
	}
	address, _ := url.Parse(server.URL)
	wrongName, err := NewFrameworkRunsBackend(WorkerConfig{
		BaseURL: "https://wrong.example.test:" + address.Port(), APIKeySecretRef: "env:MYCELIS_WORKER_API_KEY",
	}, &countedFrameworkSecret{})
	if err != nil {
		t.Fatal(err)
	}
	wrongTransport := wrongName.client.Transport.(*http.Transport)
	wrongTransport.TLSClientConfig.RootCAs = trusted.client.Transport.(*http.Transport).TLSClientConfig.RootCAs
	wrongTransport.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, address.Host)
	}
	if _, err := wrongName.HealthCheck(t.Context()); err == nil || requests.Load() != 1 {
		t.Fatalf("wrong hostname reached handler: error=%v requests=%d", err, requests.Load())
	}
}

func TestFrameworkRunsRedirectsNeverContactDestination(t *testing.T) {
	for _, status := range []int{http.StatusMovedPermanently, http.StatusFound, http.StatusTemporaryRedirect, http.StatusPermanentRedirect} {
		for _, operation := range []string{"read", "create", "events", "control"} {
			t.Run(fmt.Sprintf("%d_%s", status, operation), func(t *testing.T) {
				var sinkContacts atomic.Int32
				sink := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					sinkContacts.Add(1)
					w.WriteHeader(http.StatusOK)
				}))
				defer sink.Close()
				var originContacts atomic.Int32
				origin := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
					originContacts.Add(1)
					w.Header().Set("Location", sink.URL+"/stolen")
					w.WriteHeader(status)
				}))
				defer origin.Close()
				backend := newTrustedFrameworkTestClient(t, origin)
				var err error
				switch operation {
				case "read":
					var out map[string]any
					err = backend.doJSON(t.Context(), http.MethodGet, "/health", nil, &out)
				case "create":
					var out map[string]any
					err = backend.doJSON(t.Context(), http.MethodPost, "/v1/runs", map[string]any{"run_id": "run-1"}, &out)
				case "events":
					_, err = backend.StreamRunEventsAfter(t.Context(), "run-1", 0)
				case "control":
					_, err = backend.postControl(t.Context(), "/v1/runs/run-1/stop", map[string]any{"command_id": "stop-1"}, "run-1", "stop-1")
				}
				if err == nil || originContacts.Load() != 1 || sinkContacts.Load() != 0 {
					t.Fatalf("redirect escaped: error=%v origin=%d sink=%d", err, originContacts.Load(), sinkContacts.Load())
				}
			})
		}
	}
}

func TestFrameworkRunsRejectsSameOriginRedirect(t *testing.T) {
	var redirected, target atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			redirected.Add(1)
			w.Header().Set("Location", "/v1/capabilities")
			w.WriteHeader(http.StatusTemporaryRedirect)
			return
		}
		target.Add(1)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()
	backend := newTrustedFrameworkTestClient(t, server)
	if _, err := backend.HealthCheck(t.Context()); err == nil || redirected.Load() != 1 || target.Load() != 0 {
		t.Fatalf("same-origin redirect followed: error=%v first=%d target=%d", err, redirected.Load(), target.Load())
	}
}

func TestFrameworkRunsIgnoresEnvironmentProxy(t *testing.T) {
	var proxyContacts, originContacts atomic.Int32
	proxy := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		proxyContacts.Add(1)
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer proxy.Close()
	t.Setenv("HTTPS_PROXY", proxy.URL)
	t.Setenv("https_proxy", proxy.URL)
	t.Setenv("NO_PROXY", "")
	t.Setenv("no_proxy", "")
	origin := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		originContacts.Add(1)
		_ = json.NewEncoder(w).Encode(map[string]any{"healthy": true})
	}))
	defer origin.Close()
	address, _ := url.Parse(origin.URL)
	backend, err := NewFrameworkRunsBackend(WorkerConfig{
		BaseURL:         "https://example.com:" + address.Port(),
		APIKeySecretRef: "env:MYCELIS_WORKER_API_KEY",
	}, &countedFrameworkSecret{})
	if err != nil {
		t.Fatal(err)
	}
	roots := x509.NewCertPool()
	roots.AddCert(origin.Certificate())
	transport := backend.client.Transport.(*http.Transport)
	transport.TLSClientConfig.RootCAs = roots
	transport.DialContext = func(ctx context.Context, network, _ string) (net.Conn, error) {
		return (&net.Dialer{}).DialContext(ctx, network, address.Host)
	}
	if health, err := backend.HealthCheck(t.Context()); err != nil || !health.Healthy || originContacts.Load() != 1 || proxyContacts.Load() != 0 {
		t.Fatalf("proxy intercepted request: health=%#v error=%v origin=%d proxy=%d", health, err, originContacts.Load(), proxyContacts.Load())
	}
}

func TestFrameworkRunsHTTPFixtureRejectsNonLoopbackEvenWithPort(t *testing.T) {
	for _, hostname := range []string{"localhost", "0.0.0.0", "[::]", "192.168.1.2"} {
		_, err := newFrameworkRunsBackend(WorkerConfig{
			BaseURL: "http://" + hostname + ":8091", APIKeySecretRef: "env:MYCELIS_WORKER_API_KEY",
		}, &countedFrameworkSecret{}, true)
		if err == nil {
			t.Errorf("HTTP fixture accepted %s: %v", hostname, err)
		}
	}
}
