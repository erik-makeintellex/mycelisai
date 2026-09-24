package config

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	ListenAddress          string
	ProbeAddress           string
	ProbeServerName        string
	TLSCertFile            string
	TLSKeyFile             string
	TLSCAFile              string
	AllowInsecureLocalHTTP bool
	DatabaseURL            string
	CoreToken              string
	MaxRuns                int
	LeaseDuration          time.Duration
}

func FromEnv() (Config, error) {
	databaseURL, err := secretFromEnv("FRAMEWORK_RUNS_DATABASE_URL", "FRAMEWORK_RUNS_DATABASE_URL_FILE")
	if err != nil {
		return Config{}, err
	}
	coreToken, err := secretFromEnv("FRAMEWORK_RUNS_CORE_TOKEN", "FRAMEWORK_RUNS_CORE_TOKEN_FILE")
	if err != nil {
		return Config{}, err
	}
	config := Config{
		ListenAddress:   envOr("FRAMEWORK_RUNS_LISTEN_ADDRESS", "127.0.0.1:8091"),
		ProbeServerName: strings.TrimSpace(os.Getenv("FRAMEWORK_RUNS_PROBE_SERVER_NAME")),
		TLSCertFile:     strings.TrimSpace(os.Getenv("FRAMEWORK_RUNS_TLS_CERT_FILE")),
		TLSKeyFile:      strings.TrimSpace(os.Getenv("FRAMEWORK_RUNS_TLS_KEY_FILE")),
		TLSCAFile:       strings.TrimSpace(os.Getenv("FRAMEWORK_RUNS_TLS_CA_FILE")),
		DatabaseURL:     strings.TrimSpace(databaseURL),
		CoreToken:       coreToken,
		MaxRuns:         10_000,
		LeaseDuration:   30 * time.Second,
	}
	config.ProbeAddress = envOr("FRAMEWORK_RUNS_PROBE_ADDRESS", config.ListenAddress)
	switch os.Getenv("FRAMEWORK_RUNS_ALLOW_INSECURE_LOCAL_HTTP") {
	case "":
	case "1":
		config.AllowInsecureLocalHTTP = true
	default:
		return Config{}, errors.New("FRAMEWORK_RUNS_ALLOW_INSECURE_LOCAL_HTTP must be 1 or unset")
	}
	if config.DatabaseURL == "" {
		return Config{}, errors.New("FRAMEWORK_RUNS_DATABASE_URL or FRAMEWORK_RUNS_DATABASE_URL_FILE is required")
	}
	if config.CoreToken != strings.TrimSpace(config.CoreToken) || len(config.CoreToken) < 32 {
		return Config{}, errors.New("FRAMEWORK_RUNS_CORE_TOKEN must be canonical and at least 32 bytes")
	}
	if err := config.validateTransport(); err != nil {
		return Config{}, err
	}
	if raw := strings.TrimSpace(os.Getenv("FRAMEWORK_RUNS_MAX_RUNS")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 {
			return Config{}, errors.New("FRAMEWORK_RUNS_MAX_RUNS must be positive")
		}
		config.MaxRuns = value
	}
	if raw := strings.TrimSpace(os.Getenv("FRAMEWORK_RUNS_LEASE_SECONDS")); raw != "" {
		value, err := strconv.Atoi(raw)
		if err != nil || value < 1 || value > 3600 {
			return Config{}, errors.New("FRAMEWORK_RUNS_LEASE_SECONDS must be from 1 through 3600")
		}
		config.LeaseDuration = time.Duration(value) * time.Second
	}
	return config, nil
}

func (config Config) validateTransport() error {
	listenHost, listenPort, err := net.SplitHostPort(config.ListenAddress)
	if err != nil || listenHost == "" || invalidPort(listenPort) {
		return errors.New("FRAMEWORK_RUNS_LISTEN_ADDRESS must be a host:port with a valid port")
	}
	probeHost, probePort, err := net.SplitHostPort(config.ProbeAddress)
	if err != nil || probeHost == "" || invalidPort(probePort) || probePort != listenPort {
		return errors.New("FRAMEWORK_RUNS_PROBE_ADDRESS must use the listener port")
	}
	if probeHost != listenHost && !(isWildcard(listenHost) && net.ParseIP(probeHost) != nil && net.ParseIP(probeHost).IsLoopback()) {
		return errors.New("FRAMEWORK_RUNS_PROBE_ADDRESS must target the listener or loopback on a wildcard listener")
	}
	if config.TLSCertFile == "" || config.TLSKeyFile == "" {
		if config.TLSCertFile != "" || config.TLSKeyFile != "" || config.TLSCAFile != "" || config.ProbeServerName != "" {
			return errors.New("TLS certificate and key must be configured together")
		}
		if !config.AllowInsecureLocalHTTP || net.ParseIP(listenHost) == nil || !net.ParseIP(listenHost).IsLoopback() {
			return errors.New("plain HTTP requires explicit isolated local proof on a literal loopback listener")
		}
		return nil
	}
	if config.AllowInsecureLocalHTTP {
		return errors.New("local HTTP exception cannot be combined with TLS")
	}
	if config.TLSCAFile == "" || config.ProbeServerName == "" {
		return errors.New("TLS requires probe CA file and server name")
	}
	if _, err := tls.LoadX509KeyPair(config.TLSCertFile, config.TLSKeyFile); err != nil {
		return fmt.Errorf("TLS certificate or key is invalid: %w", err)
	}
	caPEM, err := os.ReadFile(config.TLSCAFile)
	if err != nil {
		return fmt.Errorf("TLS probe CA is unreadable: %w", err)
	}
	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(caPEM) {
		return errors.New("TLS probe CA is invalid")
	}
	return nil
}

func invalidPort(value string) bool {
	port, err := strconv.Atoi(value)
	return err != nil || port < 1 || port > 65535
}

func isWildcard(host string) bool { return host == "0.0.0.0" || host == "::" }

func secretFromEnv(valueKey, fileKey string) (string, error) {
	value := os.Getenv(valueKey)
	file := os.Getenv(fileKey)
	if value != "" && file != "" {
		return "", fmt.Errorf("%s and %s are mutually exclusive", valueKey, fileKey)
	}
	if file == "" {
		return value, nil
	}
	if file != strings.TrimSpace(file) {
		return "", fmt.Errorf("%s path is not canonical", fileKey)
	}
	reader, err := os.Open(file)
	if err != nil {
		return "", fmt.Errorf("%s is unreadable: %w", fileKey, err)
	}
	defer reader.Close()
	const maxSecretBytes = 16 << 10
	content, err := io.ReadAll(io.LimitReader(reader, maxSecretBytes+1))
	if err != nil || len(content) > maxSecretBytes {
		return "", fmt.Errorf("%s is unreadable or too large", fileKey)
	}
	result := strings.TrimSuffix(string(content), "\n")
	result = strings.TrimSuffix(result, "\r")
	if result == "" || strings.ContainsAny(result, "\r\n") {
		return "", fmt.Errorf("%s must contain one nonempty line", fileKey)
	}
	return result, nil
}

func envOr(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
