package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

const validToken = "0123456789abcdef0123456789abcdef"

func localEnv(t *testing.T) {
	t.Helper()
	t.Setenv("FRAMEWORK_RUNS_LISTEN_ADDRESS", "127.0.0.1:8091")
	t.Setenv("FRAMEWORK_RUNS_PROBE_ADDRESS", "")
	t.Setenv("FRAMEWORK_RUNS_ALLOW_INSECURE_LOCAL_HTTP", "1")
	t.Setenv("FRAMEWORK_RUNS_DATABASE_URL", "postgres://service.invalid/runs")
	t.Setenv("FRAMEWORK_RUNS_DATABASE_URL_FILE", "")
	t.Setenv("FRAMEWORK_RUNS_CORE_TOKEN", validToken)
	t.Setenv("FRAMEWORK_RUNS_CORE_TOKEN_FILE", "")
	t.Setenv("FRAMEWORK_RUNS_TLS_CERT_FILE", "")
	t.Setenv("FRAMEWORK_RUNS_TLS_KEY_FILE", "")
	t.Setenv("FRAMEWORK_RUNS_TLS_CA_FILE", "")
	t.Setenv("FRAMEWORK_RUNS_PROBE_SERVER_NAME", "")
}

func TestFromEnvRequiresDatabaseAndCanonicalCredential(t *testing.T) {
	localEnv(t)
	t.Setenv("FRAMEWORK_RUNS_DATABASE_URL", "")
	t.Setenv("FRAMEWORK_RUNS_CORE_TOKEN", "")
	if _, err := FromEnv(); err == nil {
		t.Fatal("missing database and credential were accepted")
	}
	t.Setenv("FRAMEWORK_RUNS_DATABASE_URL", "postgres://service.invalid/runs")
	t.Setenv("FRAMEWORK_RUNS_CORE_TOKEN", " short credential with whitespace ")
	if _, err := FromEnv(); err == nil {
		t.Fatal("noncanonical credential was accepted")
	}
}

func TestFromEnvAppliesBoundedOperationalSettings(t *testing.T) {
	localEnv(t)
	t.Setenv("FRAMEWORK_RUNS_DATABASE_URL", "postgres://service.invalid/runs")
	t.Setenv("FRAMEWORK_RUNS_CORE_TOKEN", "0123456789abcdef0123456789abcdef")
	t.Setenv("FRAMEWORK_RUNS_MAX_RUNS", "64")
	t.Setenv("FRAMEWORK_RUNS_LEASE_SECONDS", "45")
	settings, err := FromEnv()
	if err != nil || settings.MaxRuns != 64 || settings.LeaseDuration.Seconds() != 45 {
		t.Fatalf("settings = %#v, %v", settings, err)
	}
	for key, value := range map[string]string{
		"FRAMEWORK_RUNS_MAX_RUNS": "0", "FRAMEWORK_RUNS_LEASE_SECONDS": "3601",
	} {
		t.Setenv("FRAMEWORK_RUNS_MAX_RUNS", "64")
		t.Setenv("FRAMEWORK_RUNS_LEASE_SECONDS", "45")
		t.Setenv(key, value)
		if _, err := FromEnv(); err == nil {
			t.Fatalf("invalid %s was accepted", key)
		}
	}
}

func TestTransportFailsClosedWithoutTLSOrLiteralLoopbackException(t *testing.T) {
	localEnv(t)
	t.Setenv("FRAMEWORK_RUNS_ALLOW_INSECURE_LOCAL_HTTP", "")
	if _, err := FromEnv(); err == nil {
		t.Fatal("default plain HTTP was accepted")
	}
	for _, address := range []string{"0.0.0.0:8091", "localhost:8091", "framework-runs-control:8091"} {
		t.Setenv("FRAMEWORK_RUNS_LISTEN_ADDRESS", address)
		t.Setenv("FRAMEWORK_RUNS_ALLOW_INSECURE_LOCAL_HTTP", "1")
		if _, err := FromEnv(); err == nil {
			t.Fatalf("nonloopback HTTP listener %s was accepted", address)
		}
	}
	t.Setenv("FRAMEWORK_RUNS_LISTEN_ADDRESS", "127.0.0.1:8091")
	t.Setenv("FRAMEWORK_RUNS_TLS_CERT_FILE", "/missing/cert.pem")
	if _, err := FromEnv(); err == nil {
		t.Fatal("partial TLS configuration fell back to HTTP")
	}
}

func TestTLSRequiresPairTrustAndScopedProbeTarget(t *testing.T) {
	localEnv(t)
	directory := t.TempDir()
	certificate := filepath.Join(directory, "cert.pem")
	key := filepath.Join(directory, "key.pem")
	ca := filepath.Join(directory, "ca.pem")
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	serial, err := rand.Int(rand.Reader, new(big.Int).Lsh(big.NewInt(1), 128))
	if err != nil {
		t.Fatal(err)
	}
	template := &x509.Certificate{
		SerialNumber:          serial,
		Subject:               pkix.Name{CommonName: "framework-runs-control"},
		DNSNames:              []string{"framework-runs-control"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IsCA:                  true,
	}
	certDER, err := x509.CreateCertificate(rand.Reader, template, template, &privateKey.PublicKey, privateKey)
	if err != nil {
		t.Fatal(err)
	}
	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certDER})
	keyDER, err := x509.MarshalPKCS8PrivateKey(privateKey)
	if err != nil {
		t.Fatal(err)
	}
	for name, data := range map[string][]byte{
		certificate: certPEM,
		key:         pem.EncodeToMemory(&pem.Block{Type: "PRIVATE KEY", Bytes: keyDER}),
		ca:          certPEM,
	} {
		if err := os.WriteFile(name, data, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	t.Setenv("FRAMEWORK_RUNS_ALLOW_INSECURE_LOCAL_HTTP", "")
	t.Setenv("FRAMEWORK_RUNS_LISTEN_ADDRESS", "framework-runs-control:8091")
	t.Setenv("FRAMEWORK_RUNS_PROBE_ADDRESS", "framework-runs-control:8091")
	t.Setenv("FRAMEWORK_RUNS_PROBE_SERVER_NAME", "framework-runs-control")
	t.Setenv("FRAMEWORK_RUNS_TLS_CERT_FILE", certificate)
	t.Setenv("FRAMEWORK_RUNS_TLS_KEY_FILE", key)
	t.Setenv("FRAMEWORK_RUNS_TLS_CA_FILE", ca)
	if _, err := FromEnv(); err != nil {
		t.Fatalf("TLS control alias rejected: %v", err)
	}
	t.Setenv("FRAMEWORK_RUNS_PROBE_ADDRESS", "unrelated-service:8091")
	if _, err := FromEnv(); err == nil {
		t.Fatal("unrelated probe target was accepted")
	}
	t.Setenv("FRAMEWORK_RUNS_PROBE_ADDRESS", "framework-runs-control:8091")
	if err := os.WriteFile(key, []byte("invalid key"), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := FromEnv(); err == nil {
		t.Fatal("invalid TLS key was accepted")
	}
}

func TestSecretFilesAreBoundedAndMutuallyExclusive(t *testing.T) {
	localEnv(t)
	directory := t.TempDir()
	tokenFile := filepath.Join(directory, "token")
	databaseFile := filepath.Join(directory, "database")
	if err := os.WriteFile(tokenFile, []byte(validToken+"\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(databaseFile, []byte("postgres://service.invalid/runs\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	t.Setenv("FRAMEWORK_RUNS_CORE_TOKEN", "")
	t.Setenv("FRAMEWORK_RUNS_DATABASE_URL", "")
	t.Setenv("FRAMEWORK_RUNS_CORE_TOKEN_FILE", tokenFile)
	t.Setenv("FRAMEWORK_RUNS_DATABASE_URL_FILE", databaseFile)
	settings, err := FromEnv()
	if err != nil || settings.CoreToken != validToken || settings.DatabaseURL != "postgres://service.invalid/runs" {
		t.Fatalf("secret file configuration rejected: %v", err)
	}
	t.Setenv("FRAMEWORK_RUNS_CORE_TOKEN", validToken)
	if _, err := FromEnv(); err == nil {
		t.Fatal("simultaneous token value and file were accepted")
	}
	t.Setenv("FRAMEWORK_RUNS_CORE_TOKEN", "")
	if err := os.WriteFile(tokenFile, make([]byte, (16<<10)+1), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := FromEnv(); err == nil {
		t.Fatal("oversize token file was accepted")
	}
}
