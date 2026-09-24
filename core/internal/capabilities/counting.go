package capabilities

import (
	"net/url"
	"strings"
)

// countingManifest projects operator configuration into the existing registry.
// It deliberately has no ToolRefs: legacy discovery must not execute it.
func countingManifest(endpoint string) (Manifest, bool) {
	endpoint = strings.TrimSpace(endpoint)
	if endpoint == "" {
		return Manifest{}, false
	}
	parsed, err := url.Parse(endpoint)
	valid := err == nil && (parsed.Scheme == "http" || parsed.Scheme == "https") && parsed.Hostname() != "" && parsed.User == nil && parsed.RawQuery == "" && parsed.Fragment == ""
	status, health := "available", "unknown"
	if !valid {
		status, health = "disabled", "unavailable"
		endpoint = ""
	}
	return Manifest{
		ID: "counting.increment", CapabilityID: "counting.increment",
		DisplayName: "Counting verification capability", Kind: "governed_effect",
		Source: "core", Status: status, Health: health, RiskClass: "low-risk",
		Description:   "Non-secret counting fixture; requires confirmed grant and durable invocation.",
		AuditRequired: true, ApprovalRequired: true, Owner: "core",
		SecretRefPolicy: "none", FailurePosture: "unknown_effect_no_retry",
		Metadata: map[string]any{"invocation_binding": map[string]any{
			"endpoint": endpoint, "adapter": "counting-http", "version": "1",
			"method": "POST", "schema": "counting.v1", "unit_cost": 1,
		}},
	}, true
}
