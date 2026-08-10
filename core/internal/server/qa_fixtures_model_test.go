package server

import (
	"strings"
	"testing"
	"time"
)

func TestQAFixtureManagementRequiresExplicitOptIn(t *testing.T) {
	t.Setenv("MYCELIS_QA_FIXTURE_MANAGEMENT", "")
	if qaFixtureManagementEnabled() {
		t.Fatal("fixture management must be disabled by default")
	}
	t.Setenv("MYCELIS_QA_FIXTURE_MANAGEMENT", "true")
	if !qaFixtureManagementEnabled() {
		t.Fatal("fixture management should accept explicit true")
	}
}

func TestNormalizeQAFixtureResourceRejectsSharedOrImplicitTargets(t *testing.T) {
	for _, resource := range []qaFixtureResource{
		{Kind: "nats", Ref: "swarm.>"},
		{Kind: "workspace_path", Ref: "generated/test"},
		{Kind: "workspace_path", Ref: "groups"},
		{Kind: "run", Ref: "not-a-uuid"},
	} {
		if _, err := normalizeQAFixtureResource(resource); err == nil {
			t.Fatalf("expected resource to be rejected: %#v", resource)
		}
	}
}

func TestNormalizeQAFixtureResourceAcceptsGovernedWorkspace(t *testing.T) {
	resource, err := normalizeQAFixtureResource(qaFixtureResource{
		Kind: "workspace_path",
		Ref:  `groups\acceptance-123\generated`,
	})
	if err != nil {
		t.Fatalf("normalize resource: %v", err)
	}
	if resource.Ref != "groups/acceptance-123/generated" {
		t.Fatalf("unexpected normalized path %q", resource.Ref)
	}
}

func TestQAFixtureExpiryIsBounded(t *testing.T) {
	if _, err := qaFixtureExpiry(-1); err == nil {
		t.Fatal("negative fixture TTL must be rejected")
	}
	expires, err := qaFixtureExpiry(60)
	if err != nil {
		t.Fatalf("fixture expiry: %v", err)
	}
	if remaining := time.Until(expires); remaining < 55*time.Second || remaining > 65*time.Second {
		t.Fatalf("unexpected fixture TTL %s", remaining)
	}
	if _, err := qaFixtureExpiry(int((qaFixtureMaximumTTL + time.Second).Seconds())); err == nil ||
		!strings.Contains(err.Error(), "7 days") {
		t.Fatalf("expected bounded TTL error, got %v", err)
	}
}
