package main

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/nats-io/nats.go"
)

const defaultNATSServiceID = "mycelis-core"

var natsServiceIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{2,62}$`)

type natsRuntimeConfig struct {
	URL       string
	ServiceID string
}

func resolveNATSRuntimeConfig() (natsRuntimeConfig, error) {
	serviceID := strings.ToLower(strings.TrimSpace(os.Getenv("MYCELIS_NATS_SERVICE_ID")))
	if serviceID == "" {
		serviceID = defaultNATSServiceID
	}
	if !natsServiceIDPattern.MatchString(serviceID) {
		return natsRuntimeConfig{}, fmt.Errorf(
			"MYCELIS_NATS_SERVICE_ID must be 3-63 lowercase letters, numbers, dots, dashes, or underscores",
		)
	}

	natsURL := strings.TrimSpace(os.Getenv("NATS_URL"))
	if natsURL == "" {
		natsURL = nats.DefaultURL
	}
	return natsRuntimeConfig{URL: natsURL, ServiceID: serviceID}, nil
}

func (cfg natsRuntimeConfig) connectionName(lane string) string {
	return cfg.ServiceID + "." + strings.TrimSpace(lane)
}

func natsEndpointLabel(raw string) string {
	parts := strings.Split(raw, ",")
	labels := make([]string, 0, len(parts))
	for _, part := range parts {
		parsed, err := url.Parse(strings.TrimSpace(part))
		if err != nil || parsed.Host == "" {
			return "configured NATS host"
		}
		parsed.User = nil
		parsed.RawQuery = ""
		parsed.Fragment = ""
		labels = append(labels, parsed.String())
	}
	return strings.Join(labels, ",")
}
