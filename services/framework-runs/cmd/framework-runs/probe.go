package main

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/mycelis/framework-runs/internal/config"
)

func probe(settings config.Config) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	scheme := "https"
	transport := &http.Transport{
		Proxy:             nil,
		DialContext:       (&net.Dialer{Timeout: time.Second}).DialContext,
		DisableKeepAlives: true,
	}
	if settings.AllowInsecureLocalHTTP {
		scheme = "http"
	} else {
		caPEM, err := os.ReadFile(settings.TLSCAFile)
		if err != nil {
			return errors.New("probe trust bundle unavailable")
		}
		roots := x509.NewCertPool()
		if !roots.AppendCertsFromPEM(caPEM) {
			return errors.New("probe trust bundle invalid")
		}
		transport.TLSClientConfig = &tls.Config{
			MinVersion: tls.VersionTLS13,
			RootCAs:    roots,
			ServerName: settings.ProbeServerName,
		}
	}
	client := &http.Client{
		Transport: transport,
		Timeout:   2 * time.Second,
		CheckRedirect: func(*http.Request, []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, scheme+"://"+settings.ProbeAddress+"/health", nil)
	if err != nil {
		return errors.New("probe target invalid")
	}
	request.Header.Set("Authorization", "Bearer "+settings.CoreToken)
	response, err := client.Do(request)
	if err != nil {
		return fmt.Errorf("probe request failed: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("probe readiness rejected with HTTP %d", response.StatusCode)
	}
	var state struct {
		Healthy         bool   `json:"healthy"`
		ControllerReady bool   `json:"controller_ready"`
		ProductionReady *bool  `json:"production_ready"`
		Backend         string `json:"backend"`
		Protocol        string `json:"protocol"`
	}
	if err := json.NewDecoder(io.LimitReader(response.Body, 4096)).Decode(&state); err != nil {
		return errors.New("probe readiness body invalid")
	}
	if !state.Healthy || !state.ControllerReady || state.ProductionReady == nil || *state.ProductionReady || state.Backend != "framework_runs" || state.Protocol != "runs_api" {
		return errors.New("probe readiness contract not met")
	}
	return nil
}
