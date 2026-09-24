package main

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/mycelis/framework-runs/internal/config"
)

const probeToken = "0123456789abcdef0123456789abcdef"

func TestProbeRequiresAuthenticatedControllerReadiness(t *testing.T) {
	status := http.StatusOK
	readyBody := `{"healthy":true,"controller_ready":true,"production_ready":false,"backend":"framework_runs","protocol":"runs_api"}`
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer "+probeToken {
			response.WriteHeader(http.StatusUnauthorized)
			return
		}
		response.WriteHeader(status)
		_, _ = response.Write([]byte(readyBody))
	}))
	defer server.Close()
	settings := probeSettings(t, server)
	if err := probe(settings); err != nil {
		t.Fatalf("authenticated TLS readiness rejected: %v", err)
	}
	settings.CoreToken = "wrong"
	if err := probe(settings); err == nil {
		t.Fatal("wrong bearer passed readiness")
	}
	settings.CoreToken = probeToken
	status = http.StatusServiceUnavailable
	if err := probe(settings); err == nil {
		t.Fatal("database-unready response passed readiness")
	}
	status = http.StatusOK
	readyBody = `{"healthy":true,"controller_ready":true,"production_ready":true,"backend":"framework_runs","protocol":"runs_api"}`
	if err := probe(settings); err == nil {
		t.Fatal("production-ready claim passed protocol-only B2 probe")
	}
	readyBody = `{"healthy":true,"controller_ready":true,"backend":"framework_runs","protocol":"runs_api"}`
	if err := probe(settings); err == nil {
		t.Fatal("missing production readiness field passed probe")
	}
}

func TestProbeRejectsRedirectWrongTrustAndUnknownHost(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		http.Redirect(response, request, "/health", http.StatusFound)
	}))
	defer server.Close()
	settings := probeSettings(t, server)
	if err := probe(settings); err == nil {
		t.Fatal("redirect passed readiness")
	}
	settings.ProbeServerName = "wrong.example"
	if err := probe(settings); err == nil {
		t.Fatal("wrong certificate name passed readiness")
	}
	settings.ProbeServerName = "127.0.0.1"
	wrongCA := filepath.Join(t.TempDir(), "wrong-ca.pem")
	writeUnrelatedCA(t, wrongCA)
	settings.TLSCAFile = wrongCA
	if err := probe(settings); err == nil {
		t.Fatal("wrong trust anchor passed readiness")
	}
	settings.ProbeAddress = "no-such-framework-runs.invalid:8091"
	if err := probe(settings); err == nil {
		t.Fatal("unknown probe host passed readiness")
	}
}

func TestProbeTimesOutOnStalledAuthenticatedResponse(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(_ http.ResponseWriter, request *http.Request) {
		<-request.Context().Done()
	}))
	defer server.Close()
	settings := probeSettings(t, server)
	start := time.Now()
	if err := probe(settings); err == nil {
		t.Fatal("stalled response passed readiness")
	}
	if elapsed := time.Since(start); elapsed > 3*time.Second {
		t.Fatalf("stalled probe exceeded bounded timeout: %s", elapsed)
	}
}

func writeUnrelatedCA(t *testing.T, path string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		t.Fatal(err)
	}
	certificate := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "unrelated test CA"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, certificate, certificate, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0o600); err != nil {
		t.Fatal(err)
	}
}

func probeSettings(t *testing.T, server *httptest.Server) config.Config {
	t.Helper()
	ca := filepath.Join(t.TempDir(), "ca.pem")
	writeCert(t, ca, server)
	return config.Config{
		ProbeAddress:    strings.TrimPrefix(server.URL, "https://"),
		ProbeServerName: "127.0.0.1",
		TLSCAFile:       ca,
		CoreToken:       probeToken,
	}
}

func writeCert(t *testing.T, path string, server *httptest.Server) {
	t.Helper()
	data := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: server.Certificate().Raw})
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
}
