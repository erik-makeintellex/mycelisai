package server

import (
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"
)

// AUTH-C1b follow-up: an unreadable or corrupt settings file makes the
// approval policy fail strict (every tool action needs approval, strictest
// profile values), and GET /me says so. A missing file is the defaults. The
// approver role still comes only from the identity.

func authc1bPolicy(who *RequestIdentity, tool string) map[string]any {
	raw, _ := json.Marshal(buildApprovalPolicy(userGovernanceProfileFromRequest(requestAs(who)), nil, []string{tool}))
	var out map[string]any
	_ = json.Unmarshal(raw, &out)
	return out
}

func TestAuthC1bCorruptSettingsFailStrictPolicy(t *testing.T) {
	who := standardUserIdentity()
	for _, tool := range []string{"read_file", "write_file"} {
		t.Setenv("MYCELIS_USER_SETTINGS_PATH", filepath.Join(t.TempDir(), "missing.json"))
		defaults := authc1bPolicy(who, tool)
		if tool == "read_file" && defaults["approval_required"] == true {
			t.Fatalf("precondition: defaults already require approval for %s: %v", tool, defaults)
		}
		for kind, seed := range map[string]string{"truncated": `{"review_strictness":"li`, "empty": ``, "array": `[]`} {
			authc1SettingsFile(t, seed)
			got := authc1bPolicy(who, tool)
			if got["approval_required"] != true || got["approval_reason"] != approvalReasonSettingsUnavailable {
				t.Fatalf("%s/%s: corrupt settings policy = %v, want approval required (%s)", tool, kind, got, approvalReasonSettingsUnavailable)
			}
			if got["required_approver_role"] != defaults["required_approver_role"] {
				t.Fatalf("approver role %v, want the identity's %v", got["required_approver_role"], defaults["required_approver_role"])
			}
			profile, _ := got["governance_profile"].(map[string]any)
			if profile["review_strictness"] != "strict" || profile["automation_tolerance"] != "cautious" ||
				profile["cost_sensitivity"] != "high" || profile["escalation_preference"] != "halt" {
				t.Fatalf("%s/%s: profile %v, want the strictest values", tool, kind, profile)
			}
		}
	}
}

func TestAuthC1bMissingSettingsFileIsDefaults(t *testing.T) {
	t.Setenv("MYCELIS_USER_SETTINGS_PATH", filepath.Join(t.TempDir(), "missing.json"))
	who := standardUserIdentity()
	for _, tool := range []string{"read_file", "write_file", "web_search"} {
		want, _ := json.Marshal(buildApprovalPolicy(defaultUserGovernanceProfile(who.Role), nil, []string{tool}))
		got, _ := json.Marshal(authc1bPolicy(who, tool))
		var wantMap map[string]any
		_ = json.Unmarshal(want, &wantMap)
		wantNorm, _ := json.Marshal(wantMap)
		if string(got) != string(wantNorm) {
			t.Fatalf("%s: missing-file policy %s, want defaults %s", tool, got, wantNorm)
		}
	}
}

func TestAuthC1bMeSurfacesFailStrictSettings(t *testing.T) {
	s := newTestServer()
	me := func() map[string]any {
		rr := doAuthenticatedRequestAs(t, http.HandlerFunc(s.HandleMe), http.MethodGet, "/api/v1/user/me", "", standardUserIdentity())
		var out map[string]any
		if err := json.Unmarshal(rr.Body.Bytes(), &out); err != nil {
			t.Fatalf("me: %v %s", err, rr.Body.String())
		}
		return out
	}
	t.Setenv("MYCELIS_USER_SETTINGS_PATH", filepath.Join(t.TempDir(), "missing.json"))
	if got := me(); got["settings_status"] != nil {
		t.Fatalf("missing file: settings_status = %v, want none", got["settings_status"])
	}
	authc1SettingsFile(t, `{"theme":`)
	status, _ := me()["settings_status"].(map[string]any)
	if status["code"] != codeSettingsStoreUnavailable || status["approval_policy"] != "fail_strict" {
		t.Fatalf("corrupt file: settings_status = %v", status)
	}
}
