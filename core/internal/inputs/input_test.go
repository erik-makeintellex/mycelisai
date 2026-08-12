package inputs

import (
	"context"
	"errors"
	"testing"
)

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

func TestServiceRejectsDuplicateIngressSubject(t *testing.T) {
	service := NewService()
	ctx := context.Background()
	subject := "swarm.global.input.shared-service-events"

	_, err := service.Add(ctx, SourceInput{
		ID:                    "service-a",
		Name:                  "Service A",
		AllowedIngressSubject: subject,
	})
	if err != nil {
		t.Fatalf("add first source: %v", err)
	}
	_, err = service.Add(ctx, SourceInput{
		ID:                    "service-b",
		Name:                  "Service B",
		AllowedIngressSubject: subject,
	})
	if !errors.Is(err, ErrSubjectInUse) {
		t.Fatalf("duplicate error = %v, want ErrSubjectInUse", err)
	}
	if got := ErrorStatus(err); got != 409 {
		t.Fatalf("ErrorStatus = %d, want 409", got)
	}
}

func TestServiceUpdateKeepsItsOwnIngressSubject(t *testing.T) {
	service := NewService()
	ctx := context.Background()
	subject := "swarm.global.input.service-a"
	_, err := service.Add(ctx, SourceInput{
		ID:                    "service-a",
		Name:                  "Service A",
		AllowedIngressSubject: subject,
	})
	if err != nil {
		t.Fatalf("add source: %v", err)
	}
	_, err = service.Update(ctx, "service-a", SourceInput{
		Name:                  "Service A Updated",
		AllowedIngressSubject: subject,
	})
	if err != nil {
		t.Fatalf("update source: %v", err)
	}
}
