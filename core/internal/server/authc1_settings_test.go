package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// AUTH-C1 A2: settings that feed the approval policy are organization policy.
// Changing them needs root admin + governance:write and audits first; personal
// preferences in the same body still save. The approval policy never takes the
// role from the settings file.

const authc1SettingsSeed = `{"theme":"aero-light","assistant_name":"Soma","role":"owner","cost_sensitivity":"balanced","review_strictness":"standard","automation_tolerance":"balanced","escalation_preference":"ask"}`

func authc1SettingsFile(t *testing.T, seed string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "user-settings.json")
	if err := os.WriteFile(path, []byte(seed), 0o644); err != nil {
		t.Fatalf("seed settings: %v", err)
	}
	t.Setenv("MYCELIS_USER_SETTINGS_PATH", path)
	return path
}

func authc1ReadFile(t *testing.T, path string) string {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read settings: %v", err)
	}
	return string(raw)
}

func authc1PutSettings(t *testing.T, s *AdminServer, body string, who *RequestIdentity) (int, string) {
	t.Helper()
	rr := doAuthenticatedRequestAs(t, http.HandlerFunc(s.HandleUserSettings), http.MethodPut, "/api/v1/user/settings", body, who)
	return rr.Code, rr.Body.String()
}

// authc1PolicyFor is the approval policy another request would get right now.
func authc1PolicyFor(who *RequestIdentity) any {
	return buildApprovalPolicy(userGovernanceProfileFromRequest(requestAs(who)), nil, []string{"write_file", "web_search"})
}

func TestAuthC1SettingsGovernanceKeysRefusedForNonAdmin(t *testing.T) {
	bodies := map[string]string{
		"review_strictness":     `{"review_strictness":"light"}`,
		"automation_tolerance":  `{"automation_tolerance":"aggressive"}`,
		"cost_sensitivity":      `{"cost_sensitivity":"low"}`,
		"escalation_preference": `{"escalation_preference":"notify"}`,
		"role":                  `{"role":"reviewer"}`, // "admin" normalizes to the seeded "owner": no change
	}
	callers := map[string]*RequestIdentity{
		"signed-in non-admin":       standardUserIdentity(),
		"non-admin wildcard scope":  authc1UserWithScopes("*"),
		"admin without write scope": adminWithScopes(scopeGovernanceRead, scopeApprovalsDecide),
	}
	for key, body := range bodies {
		for name, who := range callers {
			t.Run(key+"/"+name, func(t *testing.T) {
				dbOpt, mock := withDB(t)
				s := newTestServer(dbOpt)
				path := authc1SettingsFile(t, authc1SettingsSeed)
				bystander := adminWithScopes("*")
				before := authc1PolicyFor(bystander)

				status, resp := authc1PutSettings(t, s, body, who)
				if status != http.StatusForbidden {
					t.Fatalf("status = %d, want 403: %s", status, resp)
				}
				data := authc1BlockerData(t, resp)
				if data["code"] != codeSettingsPolicyForbidden || data["required_scope"] != scopeGovernanceWrite {
					t.Fatalf("unexpected blocker data %v", data)
				}
				if data["refused_keys"] != key || data["preferences_saved"] != "false" {
					t.Fatalf("refused_keys/preferences_saved = %v/%v", data["refused_keys"], data["preferences_saved"])
				}
				if got := authc1ReadFile(t, path); got != authc1SettingsSeed {
					t.Fatalf("settings file changed on refusal: %s", got)
				}
				if after := authc1PolicyFor(bystander); !reflect.DeepEqual(before, after) {
					t.Fatalf("another user's approval policy changed: %+v -> %+v", before, after)
				}
				if err := mock.ExpectationsWereMet(); err != nil {
					t.Fatalf("refusal touched the audit store: %v", err)
				}
			})
		}
	}
}

func TestAuthC1SettingsPersonalPreferencesStillSave(t *testing.T) {
	s := newTestServer()
	path := authc1SettingsFile(t, authc1SettingsSeed)
	status, resp := authc1PutSettings(t, s, `{"theme":"midnight-cortex","review_strictness":"light","automation_tolerance":"aggressive"}`, standardUserIdentity())
	if status != http.StatusForbidden {
		t.Fatalf("status = %d, want 403: %s", status, resp)
	}
	data := authc1BlockerData(t, resp)
	if data["refused_keys"] != "automation_tolerance,review_strictness" || data["preferences_saved"] != "true" {
		t.Fatalf("unexpected blocker data %v", data)
	}
	var saved map[string]any
	if err := json.Unmarshal([]byte(authc1ReadFile(t, path)), &saved); err != nil {
		t.Fatal(err)
	}
	if saved["theme"] != "midnight-cortex" || saved["review_strictness"] != "standard" || saved["automation_tolerance"] != "balanced" {
		t.Fatalf("persisted settings = %v, want theme saved and policy untouched", saved)
	}

	// Personal-only writes and unchanged policy echoes stay 200 for any user.
	for _, body := range []string{`{"assistant_name":"Nova"}`, `{"theme":"system","review_strictness":"standard","role":"owner"}`} {
		if status, resp := authc1PutSettings(t, s, body, standardUserIdentity()); status != http.StatusOK {
			t.Fatalf("personal PUT %s = %d: %s", body, status, resp)
		}
	}
}

func TestAuthC1SettingsPutNeedsIdentity(t *testing.T) {
	s := newTestServer()
	path := authc1SettingsFile(t, authc1SettingsSeed)
	rr := doRequest(t, http.HandlerFunc(s.HandleUserSettings), http.MethodPut, "/api/v1/user/settings", `{"theme":"system"}`)
	assertStatus(t, rr, http.StatusUnauthorized)
	if got := authc1ReadFile(t, path); got != authc1SettingsSeed {
		t.Fatalf("anonymous PUT changed settings: %s", got)
	}
}

func TestAuthC1SettingsAdminWithGovernanceWriteAuditsAndSaves(t *testing.T) {
	dbOpt, mock := withDB(t)
	s := newTestServer(dbOpt)
	path := authc1SettingsFile(t, authc1SettingsSeed)
	expectAudits(mock, 2)
	status, resp := authc1PutSettings(t, s, `{"review_strictness":"strict","theme":"system"}`, adminWithScopes(scopeGovernanceWrite))
	if status != http.StatusOK {
		t.Fatalf("status = %d: %s", status, resp)
	}
	saved := authc1ReadFile(t, path)
	if !strings.Contains(saved, `"review_strictness": "strict"`) || !strings.Contains(saved, `"theme": "system"`) {
		t.Fatalf("admin change not persisted: %s", saved)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("policy change not audited before and after: %v", err)
	}
	if got := userGovernanceProfileFromRequest(requestAs(standardUserIdentity())).ReviewStrictness; got != "strict" {
		t.Fatalf("organization policy not applied: review_strictness %q", got)
	}
}

func TestAuthC1SettingsAuditFailureChangesNothing(t *testing.T) {
	dbOpt, mock := withDB(t)
	mock.ExpectExec("INSERT INTO log_entries").WillReturnError(errors.New("audit down"))
	for name, s := range map[string]*AdminServer{"insert-error": newTestServer(dbOpt), "no-audit-db": newTestServer()} {
		path := authc1SettingsFile(t, authc1SettingsSeed)
		status, resp := authc1PutSettings(t, s, `{"review_strictness":"strict","theme":"system"}`, adminWithScopes(scopeGovernanceWrite))
		if status != http.StatusServiceUnavailable {
			t.Fatalf("%s: status = %d, want 503: %s", name, status, resp)
		}
		if data := authc1BlockerData(t, resp); data["code"] != codeServiceUnavailable {
			t.Fatalf("%s: code = %v", name, data["code"])
		}
		if got := authc1ReadFile(t, path); got != authc1SettingsSeed {
			t.Fatalf("%s: settings changed without audit: %s", name, got)
		}
	}
}

func TestAuthC1ForgedSettingsRoleNeverSetsApproverRole(t *testing.T) {
	for _, forged := range []string{"admin", "owner", "reviewer"} {
		for name, c := range map[string]struct {
			who  *RequestIdentity
			want string
		}{
			"operator identity": {standardUserIdentity(), "operator"},
			"admin identity":    {adminWithScopes("*"), "owner"},
		} {
			t.Run(forged+"/"+name, func(t *testing.T) {
				authc1SettingsFile(t, `{"role":"`+forged+`","review_strictness":"standard"}`)
				profile := userGovernanceProfileFromRequest(requestAs(c.who))
				policy := buildApprovalPolicy(profile, nil, []string{"write_file"})
				if policy == nil || policy.RequiredApproverRole != c.want || profile.snapshot().Role != c.want {
					t.Fatalf("forged settings role %q gave approver role %+v, want %q", forged, policy, c.want)
				}
			})
		}
	}
}
