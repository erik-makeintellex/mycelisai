package workers

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"regexp"
	"strings"
	"time"
)

var frameworkPathID = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._:-]{0,127}$`)

// The client is deliberately private to the backend. Address authorization and
// private-CA binding remain C2 prerequisites before external execution is enabled.
func newFrameworkHTTPClient(policy TimeoutPolicy) *http.Client {
	transport := &http.Transport{
		Proxy: nil,
		DialContext: (&net.Dialer{
			Timeout:   durationMS(policy.ConnectMS, defaultConnectTimeout),
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSClientConfig:       &tls.Config{MinVersion: tls.VersionTLS12},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: time.Second,
	}
	return &http.Client{
		Transport: transport,
		Timeout:   durationMS(policy.RunMS, defaultRunTimeout),
		CheckRedirect: func(_ *http.Request, _ []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

func validateFrameworkRequestPath(path string) error {
	if !strings.HasPrefix(path, "/") || strings.ContainsAny(path, "%?#\\") || strings.Contains(path, "//") {
		return fmt.Errorf("framework runs request path is not canonical")
	}
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	valid := false
	switch {
	case len(parts) == 1 && parts[0] == "health":
		valid = true
	case len(parts) == 2 && parts[0] == "v1" && parts[1] == "capabilities":
		valid = true
	case len(parts) == 2 && parts[0] == "v1" && parts[1] == "runs":
		valid = true
	case len(parts) >= 3 && parts[0] == "v1" && parts[1] == "runs" && frameworkPathID.MatchString(parts[2]):
		valid = len(parts) == 3 ||
			(len(parts) == 4 && (parts[3] == "events" || parts[3] == "stop")) ||
			(len(parts) == 5 && parts[3] == "approvals" && frameworkPathID.MatchString(parts[4]))
	}
	if !valid {
		return fmt.Errorf("framework runs request path is not canonical")
	}
	return nil
}

func validateFrameworkPathID(name, id string) error {
	if !frameworkPathID.MatchString(id) {
		return fmt.Errorf("framework runs %s is not a canonical path identifier", name)
	}
	return nil
}
