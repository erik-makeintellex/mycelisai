package inputs

import "testing"

func TestNormalizeSourceInputDefaultsGuardedIngress(t *testing.T) {
	source, err := NormalizeSourceInput(SourceInput{
		Name:        "Factory Sensor Feed",
		SourceType:  "sensor",
		AdapterKind: "sensor",
		BufferMode:  BufferLatestState,
	})
	if err != nil {
		t.Fatalf("NormalizeSourceInput: %v", err)
	}
	if source.ID != "factory-sensor-feed" {
		t.Fatalf("id = %q", source.ID)
	}
	if source.AllowedIngressSubject != "swarm.global.input.factory-sensor-feed" {
		t.Fatalf("subject = %q", source.AllowedIngressSubject)
	}
	if source.ScopeKind != ScopeAll || source.BufferMode != BufferLatestState {
		t.Fatalf("source = %+v", source)
	}
}

func TestNormalizeSourceInputRejectsRawCredential(t *testing.T) {
	_, err := NormalizeSourceInput(SourceInput{
		ID:         "vendor-api",
		Name:       "Vendor API",
		SourceType: "api",
		AuthScheme: AuthBearerToken,
		SecretRef:  "sk-live-raw-secret-value",
	})
	if err == nil {
		t.Fatal("expected raw credential rejection")
	}
}

func TestNormalizeSourceInputRequiresScopedReference(t *testing.T) {
	_, err := NormalizeSourceInput(SourceInput{
		ID:        "group-feed",
		Name:      "Group Feed",
		ScopeKind: ScopeGroup,
	})
	if err == nil {
		t.Fatal("expected scope_ref requirement")
	}
}

func TestNormalizeSourceInputRejectsWildcardIngress(t *testing.T) {
	_, err := NormalizeSourceInput(SourceInput{
		ID:                    "bad-feed",
		Name:                  "Bad Feed",
		AllowedIngressSubject: "swarm.global.input.sensor.>",
	})
	if err == nil {
		t.Fatal("expected wildcard ingress rejection")
	}
}
