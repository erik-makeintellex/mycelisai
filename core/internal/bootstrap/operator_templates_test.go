package bootstrap

import (
	"strings"
	"testing"
)

func fixtureWithID(t *testing.T, id string) []byte {
	t.Helper()
	return []byte(strings.Replace(readFixture(t), "id: fixture-minimal-starter", "id: "+id, 1))
}

func TestValidateStarterDocumentAcceptsFixture(t *testing.T) {
	bundle, err := ValidateStarterDocument([]byte(readFixture(t)), fixturePolicy)
	if err != nil {
		t.Fatalf("ValidateStarterDocument() error = %v", err)
	}
	if bundle.ID != "fixture-minimal-starter" || bundle.Starter == nil {
		t.Fatalf("unexpected bundle: %+v", bundle)
	}
}

func TestValidateStarterDocumentRejectsStarterRuleFailures(t *testing.T) {
	cases := map[string]struct {
		body   string
		policy map[string]struct{}
		want   string
	}{
		"unknown policy": {readFixture(t), map[string]struct{}{"other": {}}, "is not a group in policy.yaml"},
		"no policy":      {readFixture(t), nil, "governance policy is not loaded"},
		"bad work mode":  {strings.Replace(readFixture(t), "work_mode: propose_only", "work_mode: wild", 1), fixturePolicy, "unknown work_mode"},
		"no starter": {strings.Replace(readFixture(t), "starter:\n  label: Fixture studio", "unused:\n  label: Fixture studio", 1),
			fixturePolicy, "must declare a starter block"},
		"not yaml": {"id: [", fixturePolicy, "parse"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, err := ValidateStarterDocument([]byte(tc.body), tc.policy)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("error = %v, want containing %q", err, tc.want)
			}
		})
	}
}

func TestValidateStarterDocumentAcceptsJSON(t *testing.T) {
	doc := `{"id":"json-starter","name":"JSON Starter","starter":{"label":"JSON"},
	"delivery_defaults":{"work_mode":"propose_only","approval_policy_ref":"fixture-approve-external",
	"approval_posture":"approve_external_sends","outputs":{"retain":true,"proof_required":true},"max_team_members":2},
	"teams":[{"id":"t1","name":"T1","inputs":["intent"],"deliveries":["answer"],
	"members":[{"id":"lead","role":"lead"}]}]}`
	if _, err := ValidateStarterDocument([]byte(doc), fixturePolicy); err != nil {
		t.Fatalf("JSON starter should validate: %v", err)
	}
}

func TestLoadOperatorStarterEntries(t *testing.T) {
	docs := []OperatorTemplateDocument{
		{ID: "custom-ok", Document: fixtureWithID(t, "custom-ok")},
		{ID: "custom-mismatch", Document: fixtureWithID(t, "something-else")},
		{ID: "starter-taken", Document: fixtureWithID(t, "starter-taken")},
		{ID: "custom-broken", Document: []byte("id: [")},
	}
	reserved := map[string]struct{}{"starter-taken": {}}
	entries := LoadOperatorStarterEntries(docs, reserved, fixturePolicy)
	if len(entries) != 4 {
		t.Fatalf("entries = %d, want 4", len(entries))
	}
	if ok := entryByID(t, entries, "custom-ok"); !ok.Available || ok.Bundle == nil {
		t.Fatalf("custom-ok should be available: %+v", ok)
	}
	for id, want := range map[string]string{
		"custom-mismatch": "does not match the stored id",
		"starter-taken":   "already used by a template file",
		"custom-broken":   "parse operator template custom-broken",
	} {
		entry := entryByID(t, entries, id)
		if entry.Available || !strings.Contains(entry.UnavailableReason, want) {
			t.Fatalf("%s: available=%v reason=%q, want %q", id, entry.Available, entry.UnavailableReason, want)
		}
	}
}
