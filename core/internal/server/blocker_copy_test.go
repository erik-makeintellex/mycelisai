package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/mycelis/core/internal/cognitive"
	"github.com/nats-io/nats.go"
)

// UX1: user copy never names an API path, env var, URL, permission string or
// internal term; codes and statuses are unchanged; admins get their variant.

var forbiddenUserCopy = []string{
	"/api/", "MYCELIS_", "://", "cognitive:", "approvals:", "governance:",
	"token", "purpose", "digest", "scope", "degraded", "swarm", "NATS", ".yaml", ".env",
}

func assertUserSafe(t *testing.T, label string, texts ...string) {
	t.Helper()
	for _, text := range texts {
		lower := strings.ToLower(text)
		for _, bad := range forbiddenUserCopy {
			if strings.Contains(lower, strings.ToLower(bad)) {
				t.Fatalf("%s: user copy %q contains %q", label, text, bad)
			}
		}
	}
}

type blockerEnvelope struct {
	OK    bool           `json:"ok"`
	Error string         `json:"error"`
	Data  map[string]any `json:"data"`
}

func decodeBlocker(t *testing.T, rr *httptest.ResponseRecorder) blockerEnvelope {
	t.Helper()
	var env blockerEnvelope
	if err := json.Unmarshal(rr.Body.Bytes(), &env); err != nil {
		t.Fatalf("decode %q: %v", rr.Body.String(), err)
	}
	return env
}

// userVisible is everything a non-admin reads from a blocker envelope.
func userVisible(env blockerEnvelope) []string {
	out := []string{env.Error}
	for _, key := range []string{"recommended_action", "detail", "required_scope"} {
		if v, ok := env.Data[key].(string); ok {
			out = append(out, v)
		}
	}
	return out
}

func adminNoApprovals() *RequestIdentity { return adminWithScopes("governance:read") }

func TestBlockerCopies_UserVariantsArePlain(t *testing.T) {
	for code, copy := range blockerCopies {
		if copy.User.Message == "" || copy.User.Action == "" {
			t.Fatalf("%s: user copy needs a message and an action", code)
		}
		assertUserSafe(t, code, copy.User.Message, copy.User.Action)
	}
	assertUserSafe(t, "missing team plan", missingTeamPlanApproval.User.Message, missingTeamPlanApproval.User.Action)
	for _, reason := range []string{"", "policy", "capability_risk", "cost"} {
		what, next := approverRequiredCopy(reason, false)
		assertUserSafe(t, "approver_required/"+reason, what, next)
	}
	assertUserSafe(t, "governance lock", governanceLockDetail(false))
}

func TestConfirmTokenErrors_RoleAwareCopyKeepsCodesAndStatuses(t *testing.T) {
	cases := []struct {
		err        error
		code       string
		status     int
		userPhrase string
	}{
		{errTokenAlreadyUsed, codeTokenAlreadyUsed, http.StatusConflict, "already approved"},
		{errTokenConsumed, codeTokenAlreadyUsed, http.StatusConflict, "already approved"},
		{errTokenPurposeUnknown, codeTokenPurposeUnknown, http.StatusConflict, "out of date"},
		{errTokenWrongPurpose, codeTokenWrongPurpose, http.StatusBadRequest, "can't be approved here"},
		{errBlueprintMismatch, codeBlueprintMismatch, http.StatusConflict, "team plan changed"},
		{errConfirmerNotProposer, codeConfirmerNotProposer, http.StatusForbidden, "person who asked"},
		{errTokenExpired, codeInvalidConfirmToken, http.StatusBadRequest, "no longer valid"},
	}
	for _, c := range cases {
		user := httptest.NewRecorder()
		respondConfirmTokenError(user, requestAs(standardUserIdentity()), c.err, http.StatusBadRequest)
		env := decodeBlocker(t, user)
		if user.Code != c.status || env.Data["code"] != c.code || !strings.Contains(env.Error, c.userPhrase) {
			t.Fatalf("%s user: %d %+v", c.code, user.Code, env)
		}
		if _, leaked := env.Data["detail"]; leaked {
			t.Fatalf("%s: non-admin got technical detail %v", c.code, env.Data["detail"])
		}
		assertUserSafe(t, c.code, userVisible(env)...)

		admin := httptest.NewRecorder()
		respondConfirmTokenError(admin, requestAs(adminNoApprovals()), c.err, http.StatusBadRequest)
		aenv := decodeBlocker(t, admin)
		detail, _ := aenv.Data["detail"].(string)
		if admin.Code != c.status || aenv.Data["code"] != c.code || !strings.Contains(detail, c.err.Error()) {
			t.Fatalf("%s admin: %d %+v", c.code, admin.Code, aenv)
		}
	}
	// The admin variant of confirmer_not_proposer names the missing permission.
	admin := httptest.NewRecorder()
	respondConfirmerNotProposer(admin, requestAs(adminNoApprovals()))
	if env := decodeBlocker(t, admin); !strings.Contains(env.Data["recommended_action"].(string), scopeApprovalsDecide) {
		t.Fatalf("admin confirmer_not_proposer = %+v", env)
	}
}

func TestApproverRequired_NamesTheTierReason(t *testing.T) {
	for reason, phrase := range map[string]string{
		"policy": "organization's rules", "capability_risk": "high-risk tool", "cost": "cost more than 5.00", "": "before Soma starts.",
	} {
		user := httptest.NewRecorder()
		respondApproverRequired(user, requestAs(standardUserIdentity()), reason)
		env := decodeBlocker(t, user)
		summary, _ := json.Marshal(env.Data["execution_summary"])
		if user.Code != http.StatusForbidden || env.Error != "Needs admin approval" || env.Data["code"] != "approver_required" ||
			!strings.Contains(string(summary), phrase) {
			t.Fatalf("%q user: %d %+v", reason, user.Code, env)
		}
		if got, _ := env.Data["approval_reason"].(string); got != reason {
			t.Fatalf("%q approval_reason = %q", reason, got)
		}
		action := env.Data["recommended_action"].(string)
		assertUserSafe(t, "approver_required/"+reason, env.Error, action)

		admin := httptest.NewRecorder()
		respondApproverRequired(admin, requestAs(adminNoApprovals()), reason)
		aenv := decodeBlocker(t, admin)
		if admin.Code != http.StatusForbidden || !strings.Contains(aenv.Data["recommended_action"].(string), scopeApprovalsDecide) {
			t.Fatalf("%q admin: %+v", reason, aenv)
		}
	}
}

func TestAdminRequired_GenericForbiddenOnAdminSurfaces(t *testing.T) {
	gate := func(w http.ResponseWriter, r *http.Request) {
		if _, ok := requireRootAdminScope(w, r, scopeGovernanceWrite); ok {
			w.WriteHeader(http.StatusNoContent)
		}
	}
	user := doAuthenticatedRequestAs(t, http.HandlerFunc(gate), "GET", "/", "", standardUserIdentity())
	env := decodeBlocker(t, user)
	if user.Code != http.StatusForbidden || env.OK || env.Data["code"] != codeAdminRequired || env.Data["required_scope"] != nil {
		t.Fatalf("user: %d %+v", user.Code, env)
	}
	assertUserSafe(t, "admin_required", userVisible(env)...)

	admin := doAuthenticatedRequestAs(t, http.HandlerFunc(gate), "GET", "/", "", adminWithScopes("governance:read"))
	aenv := decodeBlocker(t, admin)
	if admin.Code != http.StatusForbidden || aenv.Data["code"] != codeAdminRequired || aenv.Data["required_scope"] != scopeGovernanceWrite {
		t.Fatalf("admin: %d %+v", admin.Code, aenv)
	}
	if ok := doAuthenticatedRequestAs(t, http.HandlerFunc(gate), "GET", "/", "", adminWithScopes(scopeGovernanceWrite)); ok.Code != http.StatusNoContent {
		t.Fatalf("scoped admin must pass: %d", ok.Code)
	}
	if anon := doRequest(t, http.HandlerFunc(gate), "GET", "/", ""); anon.Code != http.StatusUnauthorized {
		t.Fatalf("anonymous must stay 401: %d", anon.Code)
	}
}

func TestIntentCommit_MissingApprovalIsPlainAndCoded(t *testing.T) {
	s := newTestServer()
	body := `{"intent":"launch","mission_blueprint":{"mission_id":"m","teams":[]}}`
	for _, c := range []struct {
		name     string
		identity *RequestIdentity
	}{{"user", standardUserIdentity()}, {"admin", localAdminIdentityForTest()}} {
		rr := doAuthenticatedRequestAs(t, http.HandlerFunc(s.handleIntentCommit), "POST", "/api/v1/intent/commit", body, c.identity)
		env := decodeBlocker(t, rr)
		if rr.Code != http.StatusForbidden || env.Data["code"] != codeInvalidConfirmToken || !strings.Contains(env.Error, "team plan") {
			t.Fatalf("%s: %d %+v", c.name, rr.Code, env)
		}
		if c.name == "user" {
			assertUserSafe(t, "commit/user", userVisible(env)...)
		} else if !strings.Contains(env.Data["recommended_action"].(string), "confirm_token") {
			t.Fatalf("admin commit copy = %+v", env)
		}
	}
}

func TestGovernancePolicyUnavailable_AdminFixUserPlain(t *testing.T) {
	user := httptest.NewRecorder()
	respondBlocker(user, requestAs(standardUserIdentity()), http.StatusServiceUnavailable, governancePolicyUnavailableCode, "", nil)
	assertUserSafe(t, "policy/user", userVisible(decodeBlocker(t, user))...)
	admin := httptest.NewRecorder()
	respondBlocker(admin, requestAs(localAdminIdentityForTest()), http.StatusServiceUnavailable, governancePolicyUnavailableCode, "", nil)
	if env := decodeBlocker(t, admin); !strings.Contains(env.Data["recommended_action"].(string), "policy.yaml") {
		t.Fatalf("admin policy copy = %+v", env)
	}
	if !strings.HasPrefix(governanceLockDetail(false), "policy_unavailable: ") || !strings.Contains(governanceLockDetail(true), "policy.yaml") {
		t.Fatal("lock detail must keep its code prefix; only admins read the fix")
	}
}

func TestTransportAndAvailability_RoleAwareActions(t *testing.T) {
	errs := []error{errors.New("context deadline exceeded"), errors.New("nats: outbound buffer limit exceeded"), nats.ErrNoResponders, errors.New("boom")}
	for _, err := range errs {
		_, blocker := buildTransportChatBlocker("Soma", err)
		user := availabilityForViewer(requestAs(standardUserIdentity()), blocker)
		assertUserSafe(t, blocker.Code, user.Summary, user.RecommendedAction)
		admin := availabilityForViewer(requestAs(localAdminIdentityForTest()), blocker)
		if !strings.Contains(admin.RecommendedAction, "NATS") || admin.AdminAction != "" {
			t.Fatalf("%s admin action = %+v", blocker.Code, admin)
		}
	}
	unset := cognitive.ExecutionAvailability{RecommendedAction: cognitive.UserEngineSetupAction}
	if got := availabilityForViewer(requestAs(localAdminIdentityForTest()), unset); got.RecommendedAction != cognitive.UserEngineSetupAction {
		t.Fatalf("admin without an admin action keeps the user action: %+v", got)
	}
	raw, _ := json.Marshal(cognitive.ExecutionAvailability{AdminAction: "DELETE /api/v1/x"})
	if strings.Contains(string(raw), "/api/") {
		t.Fatalf("AdminAction must never serialize: %s", raw)
	}
}

func TestNarrowCognitiveStatus_UsesUserCopy(t *testing.T) {
	text := cognitiveTextStatus{Status: "offline", Detail: "The chat profile's provider vllm did not answer its health probe.",
		RecommendedAction: "Start or repair provider vllm, or reset the chat profile to root.",
		userDetail:        "The AI engine Soma uses for chat isn't answering.", userAction: cognitive.UserEngineSetupAction}
	narrow := narrowCognitiveStatus(text, "offline", nil, "", false)["text"].(cognitiveTextSummary)
	assertUserSafe(t, "narrow status", narrow.Detail, narrow.RecommendedAction)
	if strings.Contains(narrow.Detail+narrow.RecommendedAction, "vllm") {
		t.Fatalf("narrow view leaked the provider id: %+v", narrow)
	}
}

// denialLeaksData reports whether a 401/403 body carries anything beyond the
// normalized blocker keys (UX1 admin_required): no resource data may leak.
func denialLeaksData(body string) bool {
	var env struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &env); err != nil {
		return true
	}
	if len(env.Data) == 0 {
		return false
	}
	var data map[string]string
	if err := json.Unmarshal(env.Data, &data); err != nil {
		return true
	}
	for key := range data {
		switch key {
		case "code", "recommended_action", "detail", "required_scope":
		default:
			return true
		}
	}
	return data["code"] != codeAdminRequired
}

func TestServiceOutages_RoleAwareCopyKeepsCodes(t *testing.T) {
	s := newTestServer() // no Soma, no NATS
	for _, c := range []struct {
		name, code, userPhrase, adminPhrase string
		call                                func(*RequestIdentity) *httptest.ResponseRecorder
	}{
		{"council members", codeTeamServiceOffline, "team service isn't running", "agent runtime", func(id *RequestIdentity) *httptest.ResponseRecorder {
			return doAuthenticatedRequestAs(t, http.HandlerFunc(s.HandleListCouncilMembers), "GET", "/api/v1/council/members", "", id)
		}},
		{"routing audit", codeServiceUnavailable, "activity log is unavailable", "audit store", func(id *RequestIdentity) *httptest.ResponseRecorder {
			rr := httptest.NewRecorder()
			respondBlocker(rr, requestAs(id), http.StatusServiceUnavailable, codeServiceUnavailable, "Audit unavailable: routing not changed", nil)
			return rr
		}},
	} {
		user := c.call(standardUserIdentity())
		env := decodeBlocker(t, user)
		if user.Code != http.StatusServiceUnavailable || env.Data["code"] != c.code || !strings.Contains(env.Error, c.userPhrase) || env.Data["detail"] != nil {
			t.Fatalf("%s user: %d %+v", c.name, user.Code, env)
		}
		assertUserSafe(t, c.name, userVisible(env)...)
		admin := c.call(localAdminIdentityForTest())
		aenv := decodeBlocker(t, admin)
		if admin.Code != http.StatusServiceUnavailable || !strings.Contains(aenv.Data["recommended_action"].(string), c.adminPhrase) || aenv.Data["detail"] == nil {
			t.Fatalf("%s admin: %d %+v", c.name, admin.Code, aenv)
		}
	}
}
